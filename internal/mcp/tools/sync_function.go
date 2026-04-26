package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/functions"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// SyncFunctionTool implements the sync_function MCP tool for deploying edge functions
type SyncFunctionTool struct {
	storage *functions.Storage
}

// NewSyncFunctionTool creates a new sync_function tool
func NewSyncFunctionTool(storage *functions.Storage) *SyncFunctionTool {
	return &SyncFunctionTool{
		storage: storage,
	}
}

func (t *SyncFunctionTool) Name() string {
	return "sync_function"
}

func (t *SyncFunctionTool) Description() string {
	return `Deploy or update an edge function. Parses @fluxbase annotations from code comments for configuration.

Supported annotations:
  @fluxbase:public - Make function publicly listed
  @fluxbase:allow-unauthenticated - Allow invocation without authentication
  @fluxbase:timeout <seconds> - Set execution timeout (default: 30)
  @fluxbase:memory <mb> - Set memory limit in MB (default: 128)
  @fluxbase:rate-limit <N>/<period> - Rate limiting (e.g., "100/min", "1000/hour", "10000/day")
  @fluxbase:cors-origins <origins> - CORS allowed origins
  @fluxbase:description <text> - Function description
  @fluxbase:allow-net - Allow network access (default: true)
  @fluxbase:allow-env - Allow environment variable access (default: true)
  @fluxbase:disable-logs - Disable execution logging

Example:
// @fluxbase:description Handle user registration
// @fluxbase:public
// @fluxbase:rate-limit 10/min
export default async function handler(req: Request) {
  return new Response("Hello!");
}`
}

func (t *SyncFunctionTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Function name (alphanumeric, hyphens, underscores)",
			},
			"code": map[string]any{
				"type":        "string",
				"description": "TypeScript/JavaScript code with optional @fluxbase annotations in comments",
			},
			"namespace": map[string]any{
				"type":        "string",
				"description": "Namespace for isolating functions (default: 'default')",
				"default":     "default",
			},
		},
		"required": []string{"name", "code"},
	}
}

func (t *SyncFunctionTool) RequiredScopes() []string {
	return []string{mcp.ScopeSyncFunctions}
}

func (t *SyncFunctionTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	// Parse arguments
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("function name is required")
	}

	code, ok := args["code"].(string)
	if !ok || code == "" {
		return nil, fmt.Errorf("function code is required")
	}

	namespace := "default"
	if ns, ok := args["namespace"].(string); ok && ns != "" {
		namespace = ns
	}

	// Validate name format
	if !isValidFunctionName(name) {
		return nil, fmt.Errorf("invalid function name: must be alphanumeric with hyphens/underscores, 1-63 characters")
	}

	// Check namespace access
	if !authCtx.HasNamespaceAccess(namespace) {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Access denied to namespace: %s", namespace))},
			IsError: true,
		}, nil
	}

	config := functions.ParseFunctionConfig(code)

	log.Debug().
		Str("name", name).
		Str("namespace", namespace).
		Interface("config", config).
		Msg("MCP: sync_function - parsed annotations")

	// Check if function already exists
	existing, err := t.storage.GetFunctionByNamespace(ctx, name, namespace)
	isNew := false
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			isNew = true
		} else {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to check existing function: %v", err))},
				IsError: true,
			}, nil
		}
	}

	var result map[string]any

	if isNew {
		fn := &functions.EdgeFunction{
			Name:                 name,
			Namespace:            namespace,
			Code:                 code,
			OriginalCode:         &code,
			IsBundled:            false,
			Enabled:              true,
			Source:               "mcp",
			TimeoutSeconds:       config.Timeout,
			MemoryLimitMB:        config.Memory,
			AllowNet:             config.AllowNet,
			AllowEnv:             config.AllowEnv,
			AllowRead:            true,
			AllowWrite:           false,
			AllowUnauthenticated: config.AllowUnauthenticated,
			IsPublic:             config.IsPublic,
			DisableExecutionLogs: config.DisableExecutionLogs,
		}

		if config.Description != "" {
			fn.Description = &config.Description
		}
		if config.CorsOrigins != nil {
			fn.CorsOrigins = config.CorsOrigins
		}
		if config.RateLimitPerMinute != nil {
			fn.RateLimitPerMinute = config.RateLimitPerMinute
		}
		if config.RateLimitPerHour != nil {
			fn.RateLimitPerHour = config.RateLimitPerHour
		}
		if config.RateLimitPerDay != nil {
			fn.RateLimitPerDay = config.RateLimitPerDay
		}

		// Note: fn.CreatedBy requires uuid.UUID but authCtx.UserID is *string
		// The EdgeFunction struct uses *uuid.UUID for CreatedBy, which would need parsing

		if err := t.storage.CreateFunction(ctx, fn); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to create function: %v", err))},
				IsError: true,
			}, nil
		}

		result = map[string]any{
			"action":    "created",
			"id":        fn.ID.String(),
			"name":      fn.Name,
			"namespace": fn.Namespace,
			"version":   fn.Version,
		}

		log.Info().
			Str("name", name).
			Str("namespace", namespace).
			Str("id", fn.ID.String()).
			Msg("MCP: sync_function - created new function")

	} else {
		updates := map[string]interface{}{
			"code":                   code,
			"original_code":          code,
			"is_bundled":             false,
			"timeout_seconds":        config.Timeout,
			"memory_limit_mb":        config.Memory,
			"allow_net":              config.AllowNet,
			"allow_env":              config.AllowEnv,
			"allow_unauthenticated":  config.AllowUnauthenticated,
			"is_public":              config.IsPublic,
			"disable_execution_logs": config.DisableExecutionLogs,
			"source":                 "mcp",
		}

		if config.Description != "" {
			updates["description"] = config.Description
		}
		if config.CorsOrigins != nil {
			updates["cors_origins"] = *config.CorsOrigins
		}
		if config.RateLimitPerMinute != nil {
			updates["rate_limit_per_minute"] = *config.RateLimitPerMinute
		}
		if config.RateLimitPerHour != nil {
			updates["rate_limit_per_hour"] = *config.RateLimitPerHour
		}
		if config.RateLimitPerDay != nil {
			updates["rate_limit_per_day"] = *config.RateLimitPerDay
		}

		if err := t.storage.UpdateFunctionByNamespace(ctx, name, namespace, updates); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to update function: %v", err))},
				IsError: true,
			}, nil
		}

		result = map[string]any{
			"action":           "updated",
			"id":               existing.ID.String(),
			"name":             name,
			"namespace":        namespace,
			"previous_version": existing.Version,
		}

		log.Info().
			Str("name", name).
			Str("namespace", namespace).
			Str("id", existing.ID.String()).
			Int("previous_version", existing.Version).
			Msg("MCP: sync_function - updated existing function")
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

func isValidFunctionName(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_-]*$`, name)
	return match
}
