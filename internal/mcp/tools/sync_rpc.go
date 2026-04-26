package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/nimbleflux/fluxbase/internal/rpc"
)

// SyncRPCTool implements the sync_rpc MCP tool for deploying RPC procedures
type SyncRPCTool struct {
	storage *rpc.Storage
}

// NewSyncRPCTool creates a new sync_rpc tool
func NewSyncRPCTool(storage *rpc.Storage) *SyncRPCTool {
	return &SyncRPCTool{
		storage: storage,
	}
}

func (t *SyncRPCTool) Name() string {
	return "sync_rpc"
}

func (t *SyncRPCTool) Description() string {
	return `Deploy or update an RPC procedure (stored SQL). Parses @fluxbase annotations from SQL comments.

Supported annotations:
  @fluxbase:description <text> - Procedure description
  @fluxbase:public - Make procedure publicly discoverable
  @fluxbase:timeout <seconds> - Max execution time (default: 30)
  @fluxbase:require-role <role> - Required role: admin, authenticated, anon
  @fluxbase:allowed-tables <tables> - Comma-separated list of allowed tables
  @fluxbase:allowed-schemas <schemas> - Comma-separated list of allowed schemas
  @fluxbase:schedule "<cron>" - Optional cron schedule for periodic execution
  @fluxbase:disable-logs - Disable execution logging

Example:
-- @fluxbase:description Get user profile with stats
-- @fluxbase:public
-- @fluxbase:allowed-tables users,user_stats
-- @fluxbase:timeout 10
SELECT u.*, s.total_posts, s.followers
FROM users u
LEFT JOIN user_stats s ON u.id = s.user_id
WHERE u.id = $1;`
}

func (t *SyncRPCTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Procedure name (alphanumeric, hyphens, underscores)",
			},
			"sql_code": map[string]any{
				"type":        "string",
				"description": "SQL code with optional @fluxbase annotations in comments",
			},
			"namespace": map[string]any{
				"type":        "string",
				"description": "Namespace for isolating procedures (default: 'default')",
				"default":     "default",
			},
		},
		"required": []string{"name", "sql_code"},
	}
}

func (t *SyncRPCTool) RequiredScopes() []string {
	return []string{mcp.ScopeSyncRPC}
}

func (t *SyncRPCTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	// Parse arguments
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("procedure name is required")
	}

	sqlCode, ok := args["sql_code"].(string)
	if !ok || sqlCode == "" {
		return nil, fmt.Errorf("sql_code is required")
	}

	namespace := "default"
	if ns, ok := args["namespace"].(string); ok && ns != "" {
		namespace = ns
	}

	// Validate name format
	if !isValidFunctionName(name) {
		return nil, fmt.Errorf("invalid procedure name: must be alphanumeric with hyphens/underscores, 1-63 characters")
	}

	// Check namespace access
	if !authCtx.HasNamespaceAccess(namespace) {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Access denied to namespace: %s", namespace))},
			IsError: true,
		}, nil
	}

	config := rpc.ParseRPCAnnotations(sqlCode)

	log.Debug().
		Str("name", name).
		Str("namespace", namespace).
		Interface("config", config).
		Msg("MCP: sync_rpc - parsed annotations")

	// Check if procedure already exists
	existing, err := t.storage.GetProcedureByName(ctx, namespace, name)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to check existing procedure: %v", err))},
			IsError: true,
		}, nil
	}

	var result map[string]any

	if existing == nil {
		// Create new procedure
		proc := &rpc.Procedure{
			Name:                    name,
			Namespace:               namespace,
			Description:             config.Description,
			SQLQuery:                sqlCode,
			OriginalCode:            sqlCode,
			AllowedTables:           config.AllowedTables,
			AllowedSchemas:          config.AllowedSchemas,
			MaxExecutionTimeSeconds: config.Timeout,
			RequireRoles:            config.RequireRoles,
			IsPublic:                config.IsPublic,
			DisableExecutionLogs:    config.DisableLogs,
			Schedule:                config.Schedule,
			Enabled:                 true,
			Version:                 1,
			Source:                  "mcp",
		}

		if err := t.storage.CreateProcedure(ctx, proc); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to create procedure: %v", err))},
				IsError: true,
			}, nil
		}

		result = map[string]any{
			"action":    "created",
			"id":        proc.ID,
			"name":      proc.Name,
			"namespace": proc.Namespace,
			"version":   proc.Version,
		}

		log.Info().
			Str("name", name).
			Str("namespace", namespace).
			Str("id", proc.ID).
			Msg("MCP: sync_rpc - created new procedure")

	} else {
		// Update existing procedure
		existing.Description = config.Description
		existing.SQLQuery = sqlCode
		existing.OriginalCode = sqlCode
		existing.AllowedTables = config.AllowedTables
		existing.AllowedSchemas = config.AllowedSchemas
		existing.MaxExecutionTimeSeconds = config.Timeout
		existing.RequireRoles = config.RequireRoles
		existing.IsPublic = config.IsPublic
		existing.DisableExecutionLogs = config.DisableLogs
		existing.Schedule = config.Schedule
		existing.Source = "mcp"

		if err := t.storage.UpdateProcedure(ctx, existing); err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to update procedure: %v", err))},
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
			Msg("MCP: sync_rpc - updated existing procedure")
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
