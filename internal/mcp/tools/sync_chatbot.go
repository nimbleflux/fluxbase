package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/loader"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// SyncChatbotTool implements the sync_chatbot MCP tool for deploying AI chatbots
type SyncChatbotTool struct {
	storage *ai.Storage
}

// NewSyncChatbotTool creates a new sync_chatbot tool
func NewSyncChatbotTool(storage *ai.Storage) *SyncChatbotTool {
	return &SyncChatbotTool{
		storage: storage,
	}
}

func (t *SyncChatbotTool) Name() string {
	return "sync_chatbot"
}

func (t *SyncChatbotTool) Description() string {
	return `Deploy or update an AI chatbot. Parses @fluxbase annotations from code comments for configuration.

Configuration annotations:
  @fluxbase:description <text> - Chatbot description
  @fluxbase:allowed-tables <tables> - Comma-separated list of tables the chatbot can query (supports schema.table format)
  @fluxbase:allowed-operations <ops> - Allowed SQL operations: SELECT, INSERT, UPDATE, DELETE
  @fluxbase:allowed-schemas <schemas> - Allowed schemas (default: public)
  @fluxbase:mcp-tools <tools> - Comma-separated MCP tools to enable (e.g., query_table,insert_record,invoke_function)
  @fluxbase:use-mcp-schema - Fetch schema from MCP resources instead of direct DB introspection
  @fluxbase:public - Make chatbot publicly discoverable
  @fluxbase:allow-unauthenticated - Allow anonymous access
  @fluxbase:model <name> - AI model to use
  @fluxbase:max-tokens <n> - Max tokens per response (default: 4096)
  @fluxbase:temperature <n> - Temperature 0-2 (default: 0.7)
  @fluxbase:rate-limit <n>/min - Rate limit per minute (default: 20)
  @fluxbase:daily-limit <n> - Daily request limit (default: 500)
  @fluxbase:persist-conversations - Enable conversation persistence
  @fluxbase:conversation-ttl <hours> - Conversation TTL in hours (default: 24)
  @fluxbase:response-language <lang> - Response language (default: "auto")
  @fluxbase:disable-logs - Disable execution logging

MCP Tools (use with @fluxbase:mcp-tools):
  query_table - Query tables with filters, ordering, pagination, and optional vector search
  insert_record - Insert a new record into a table
  update_record - Update records matching filters
  delete_record - Delete records matching filters
  invoke_function - Call an edge function
  invoke_rpc - Execute an RPC procedure
  submit_job / get_job_status - Background job management
  list_objects / upload_object / download_object / delete_object - Storage operations
  search_vectors - Semantic search over vector embeddings

Example:
// @fluxbase:description Customer support assistant
// @fluxbase:allowed-tables users,orders,products,analytics.metrics
// @fluxbase:mcp-tools query_table,insert_record,invoke_function
// @fluxbase:use-mcp-schema
// @fluxbase:public
// @fluxbase:persist-conversations
// @fluxbase:rate-limit 30/min

You are a helpful customer support agent...`
}

func (t *SyncChatbotTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Chatbot name (alphanumeric, hyphens, underscores)",
			},
			"code": map[string]any{
				"type":        "string",
				"description": "Chatbot system prompt/code with @fluxbase annotations",
			},
			"namespace": map[string]any{
				"type":        "string",
				"description": "Namespace for isolating chatbots (default: 'default')",
				"default":     "default",
			},
		},
		"required": []string{"name", "code"},
	}
}

func (t *SyncChatbotTool) RequiredScopes() []string {
	return []string{mcp.ScopeSyncChatbots}
}

func (t *SyncChatbotTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	// Parse arguments
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("chatbot name is required")
	}

	code, ok := args["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("chatbot code is required")
	}

	namespace := "default"
	if ns, ok := args["namespace"].(string); ok && ns != "" {
		namespace = ns
	}

	// Validate name format
	if !isValidFunctionName(name) {
		return nil, fmt.Errorf("invalid chatbot name: must be alphanumeric with hyphens/underscores, 1-63 characters")
	}

	// Check namespace access
	if !authCtx.HasNamespaceAccess(namespace) {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Access denied to namespace: %s", namespace))},
			IsError: true,
		}, nil
	}

	// Parse annotations from code
	config := parseChatbotAnnotations(code)

	log.Debug().
		Str("name", name).
		Str("namespace", namespace).
		Interface("config", config).
		Msg("MCP: sync_chatbot - parsed annotations")

	// Check if chatbot already exists
	existing, err := t.storage.GetChatbotByName(ctx, namespace, name)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to check existing chatbot: %v", err))},
			IsError: true,
		}, nil
	}

	var result map[string]any

	if existing == nil {
		// Create new chatbot
		chatbot := &ai.Chatbot{
			ID:                   uuid.New().String(),
			Name:                 name,
			Namespace:            namespace,
			Description:          config.Description,
			Code:                 code,
			OriginalCode:         code,
			IsBundled:            false,
			AllowedTables:        config.AllowedTables,
			AllowedOperations:    config.AllowedOperations,
			AllowedSchemas:       config.AllowedSchemas,
			HTTPAllowedDomains:   config.HTTPAllowedDomains,
			MCPTools:             config.MCPTools,
			UseMCPSchema:         config.UseMCPSchema,
			Enabled:              true,
			MaxTokens:            config.MaxTokens,
			Temperature:          config.Temperature,
			Model:                config.Model,
			PersistConversations: config.PersistConversations,
			ConversationTTLHours: config.ConversationTTLHours,
			MaxConversationTurns: config.MaxTurns,
			RateLimitPerMinute:   config.RateLimitPerMinute,
			DailyRequestLimit:    config.DailyRequestLimit,
			DailyTokenBudget:     config.DailyTokenBudget,
			AllowUnauthenticated: config.AllowUnauthenticated,
			IsPublic:             config.IsPublic,
			RequireRoles:         config.RequireRoles,
			ResponseLanguage:     config.ResponseLanguage,
			DisableExecutionLogs: config.DisableLogs,
			Version:              1,
			Source:               "mcp",
		}

		if authCtx.UserID != nil {
			chatbot.CreatedBy = authCtx.UserID
		}

		if err := t.storage.CreateChatbot(ctx, chatbot); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to create chatbot: %v", err))},
				IsError: true,
			}, nil
		}

		result = map[string]any{
			"action":    "created",
			"id":        chatbot.ID,
			"name":      chatbot.Name,
			"namespace": chatbot.Namespace,
			"version":   chatbot.Version,
		}

		log.Info().
			Str("name", name).
			Str("namespace", namespace).
			Str("id", chatbot.ID).
			Msg("MCP: sync_chatbot - created new chatbot")

	} else {
		// Update existing chatbot
		existing.Description = config.Description
		existing.Code = code
		existing.OriginalCode = code
		existing.IsBundled = false
		existing.BundleError = ""
		existing.AllowedTables = config.AllowedTables
		existing.AllowedOperations = config.AllowedOperations
		existing.AllowedSchemas = config.AllowedSchemas
		existing.HTTPAllowedDomains = config.HTTPAllowedDomains
		existing.MCPTools = config.MCPTools
		existing.UseMCPSchema = config.UseMCPSchema
		existing.MaxTokens = config.MaxTokens
		existing.Temperature = config.Temperature
		existing.Model = config.Model
		existing.PersistConversations = config.PersistConversations
		existing.ConversationTTLHours = config.ConversationTTLHours
		existing.MaxConversationTurns = config.MaxTurns
		existing.RateLimitPerMinute = config.RateLimitPerMinute
		existing.DailyRequestLimit = config.DailyRequestLimit
		existing.DailyTokenBudget = config.DailyTokenBudget
		existing.AllowUnauthenticated = config.AllowUnauthenticated
		existing.IsPublic = config.IsPublic
		existing.RequireRoles = config.RequireRoles
		existing.ResponseLanguage = config.ResponseLanguage
		existing.DisableExecutionLogs = config.DisableLogs
		existing.Source = "mcp"

		if err := t.storage.UpdateChatbot(ctx, existing); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to update chatbot: %v", err))},
				IsError: true,
			}, nil
		}

		result = map[string]any{
			"action":           "updated",
			"id":               existing.ID,
			"name":             name,
			"namespace":        namespace,
			"previous_version": existing.Version,
		}

		log.Info().
			Str("name", name).
			Str("namespace", namespace).
			Str("id", existing.ID).
			Int("previous_version", existing.Version).
			Msg("MCP: sync_chatbot - updated existing chatbot")
	}

	// Serialize result
	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to serialize result: %v", err))},
			IsError: true,
		}, nil
	}

	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// ChatbotToolConfig holds parsed @fluxbase annotations for chatbots
type ChatbotToolConfig struct {
	Description          string
	AllowedTables        []string
	AllowedOperations    []string
	AllowedSchemas       []string
	HTTPAllowedDomains   []string
	MCPTools             []string // MCP tools this chatbot can use
	UseMCPSchema         bool     // If true, fetch schema from MCP resources
	IsPublic             bool
	AllowUnauthenticated bool
	RequireRoles         []string // Required roles to access (OR semantics)
	Model                string
	MaxTokens            int
	Temperature          float64
	RateLimitPerMinute   int
	DailyRequestLimit    int
	DailyTokenBudget     int
	PersistConversations bool
	ConversationTTLHours int
	MaxTurns             int
	ResponseLanguage     string
	DisableLogs          bool
}

func parseChatbotAnnotations(code string) ChatbotToolConfig {
	annotations := loader.ParseAnnotations(code, []string{"//"})
	config := ChatbotToolConfig{
		AllowedTables:        []string{},
		AllowedOperations:    []string{"SELECT"},
		AllowedSchemas:       []string{"public"},
		HTTPAllowedDomains:   []string{},
		MCPTools:             []string{},
		UseMCPSchema:         false,
		MaxTokens:            4096,
		Temperature:          0.7,
		RateLimitPerMinute:   20,
		DailyRequestLimit:    500,
		DailyTokenBudget:     100000,
		ConversationTTLHours: 24,
		MaxTurns:             50,
		ResponseLanguage:     "auto",
	}

	if v, ok := annotations["description"]; ok {
		config.Description = v
	}
	if _, ok := annotations["public"]; ok {
		config.IsPublic = true
	}
	if _, ok := annotations["allow-unauthenticated"]; ok {
		config.AllowUnauthenticated = true
	}
	if v, ok := annotations["require-role"]; ok {
		roles := loader.ParseRoleList(v)
		if len(roles) > 0 {
			config.RequireRoles = roles
		}
	}
	if v, ok := annotations["allowed-tables"]; ok {
		tables := loader.ParseCommaList(v)
		if len(tables) > 0 {
			config.AllowedTables = tables
		}
	}
	if v, ok := annotations["allowed-operations"]; ok {
		ops := loader.ParseCommaList(v)
		if len(ops) > 0 {
			for i, op := range ops {
				ops[i] = strings.ToUpper(op)
			}
			config.AllowedOperations = ops
		}
	}
	if v, ok := annotations["allowed-schemas"]; ok {
		schemas := loader.ParseCommaList(v)
		if len(schemas) > 0 {
			config.AllowedSchemas = schemas
		}
	}
	if v, ok := annotations["http-allowed-domains"]; ok {
		domains := loader.ParseCommaList(v)
		if len(domains) > 0 {
			config.HTTPAllowedDomains = domains
		}
	}
	if v, ok := annotations["model"]; ok {
		config.Model = v
	}
	if v, ok := annotations["max-tokens"]; ok {
		if t, err := strconv.Atoi(v); err == nil && t > 0 {
			config.MaxTokens = t
		}
	}
	if v, ok := annotations["temperature"]; ok {
		if t, err := strconv.ParseFloat(v, 64); err == nil && t >= 0 && t <= 2 {
			config.Temperature = t
		}
	}
	if v, ok := annotations["rate-limit"]; ok {
		if idx := strings.Index(v, "/"); idx > 0 {
			if n, err := strconv.Atoi(strings.TrimSpace(v[:idx])); err == nil && n > 0 {
				config.RateLimitPerMinute = n
			}
		}
	}
	if v, ok := annotations["daily-limit"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.DailyRequestLimit = n
		}
	}
	if v, ok := annotations["daily-token-budget"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.DailyTokenBudget = n
		}
	}
	if _, ok := annotations["persist-conversations"]; ok {
		config.PersistConversations = true
	}
	if v, ok := annotations["conversation-ttl"]; ok {
		if h, err := strconv.Atoi(v); err == nil && h > 0 {
			config.ConversationTTLHours = h
		}
	}
	if v, ok := annotations["max-turns"]; ok {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			config.MaxTurns = n
		}
	}
	if v, ok := annotations["response-language"]; ok {
		config.ResponseLanguage = v
	}
	if _, ok := annotations["disable-logs"]; ok {
		config.DisableLogs = true
	}
	if v, ok := annotations["mcp-tools"]; ok {
		tools := loader.ParseCommaList(v)
		if len(tools) > 0 {
			config.MCPTools = tools
		}
	}
	if v, ok := annotations["use-mcp-schema"]; ok {
		if v == "" || strings.ToLower(v) == "true" {
			config.UseMCPSchema = true
		}
	}

	return config
}
