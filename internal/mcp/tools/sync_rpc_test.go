package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/loader"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

func TestNewSyncRPCTool(t *testing.T) {
	t.Run("creates tool with nil storage", func(t *testing.T) {
		tool := NewSyncRPCTool(nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.storage)
	})
}

func TestSyncRPCTool_Metadata(t *testing.T) {
	tool := NewSyncRPCTool(nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "sync_rpc", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Deploy or update an RPC procedure")
		assert.Contains(t, desc, "@fluxbase:description")
		assert.Contains(t, desc, "@fluxbase:public")
		assert.Contains(t, desc, "@fluxbase:timeout")
		assert.Contains(t, desc, "@fluxbase:allowed-tables")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "name")
		assert.Contains(t, props, "sql_code")
		assert.Contains(t, props, "namespace")

		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "sql_code")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeSyncRPC)
	})
}

func TestSyncRPCTool_Execute_Validation(t *testing.T) {
	tool := NewSyncRPCTool(nil)

	t.Run("missing name", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"sql_code": "SELECT * FROM users;",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("missing sql_code", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "get_users",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sql_code is required")
	})

	t.Run("invalid name format", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name":     "invalid procedure name!",
			"sql_code": "SELECT 1;",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid procedure name")
	})
}

func TestRPCConfig_Struct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var config RPCConfig
		assert.Empty(t, config.Description)
		assert.False(t, config.IsPublic)
		assert.Equal(t, 0, config.Timeout)
		assert.Nil(t, config.RequireRoles)
		assert.Nil(t, config.AllowedTables)
		assert.Nil(t, config.AllowedSchemas)
		assert.Nil(t, config.Schedule)
	})

	t.Run("all fields", func(t *testing.T) {
		schedule := "0 * * * *"
		config := RPCConfig{
			Description:    "Get user profile",
			IsPublic:       true,
			Timeout:        60,
			RequireRoles:   []string{"authenticated"},
			AllowedTables:  []string{"users", "profiles"},
			AllowedSchemas: []string{"public", "analytics"},
			Schedule:       &schedule,
			DisableLogs:    true,
		}

		assert.Equal(t, "Get user profile", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 60, config.Timeout)
		assert.Equal(t, "0 * * * *", *config.Schedule)
	})
}

func TestParseRPCAnnotations(t *testing.T) {
	t.Run("empty code returns defaults", func(t *testing.T) {
		config := parseRPCAnnotations("")
		assert.Equal(t, 30, config.Timeout)
		assert.Equal(t, []string{"public"}, config.AllowedSchemas)
		assert.Empty(t, config.AllowedTables)
		assert.False(t, config.IsPublic)
	})

	t.Run("description annotation", func(t *testing.T) {
		code := `-- @fluxbase:description Get user profile with stats
SELECT * FROM users WHERE id = $1;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, "Get user profile with stats", config.Description)
	})

	t.Run("public annotation", func(t *testing.T) {
		code := `-- @fluxbase:public
SELECT * FROM public_data;`
		config := parseRPCAnnotations(code)
		assert.True(t, config.IsPublic)
	})

	t.Run("timeout annotation", func(t *testing.T) {
		code := `-- @fluxbase:timeout 60
SELECT * FROM large_table;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, 60, config.Timeout)
	})

	t.Run("require-role annotation", func(t *testing.T) {
		code := `-- @fluxbase:require-role admin, authenticated
SELECT * FROM sensitive_data;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, []string{"admin", "authenticated"}, config.RequireRoles)
	})

	t.Run("allowed-tables annotation", func(t *testing.T) {
		code := `-- @fluxbase:allowed-tables users, orders, products
SELECT * FROM users;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, []string{"users", "orders", "products"}, config.AllowedTables)
	})

	t.Run("allowed-schemas annotation", func(t *testing.T) {
		code := `-- @fluxbase:allowed-schemas public, analytics
SELECT * FROM analytics.metrics;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, []string{"public", "analytics"}, config.AllowedSchemas)
	})

	t.Run("schedule annotation with quotes", func(t *testing.T) {
		code := `-- @fluxbase:schedule "0 * * * *"
SELECT cleanup_old_records();`
		config := parseRPCAnnotations(code)
		assert.NotNil(t, config.Schedule)
		assert.Equal(t, "0 * * * *", *config.Schedule)
	})

	t.Run("schedule annotation with single quotes", func(t *testing.T) {
		code := `-- @fluxbase:schedule '0 0 * * *'
SELECT daily_summary();`
		config := parseRPCAnnotations(code)
		assert.NotNil(t, config.Schedule)
		assert.Equal(t, "0 0 * * *", *config.Schedule)
	})

	t.Run("disable-logs annotation", func(t *testing.T) {
		code := `-- @fluxbase:disable-logs
SELECT sensitive_data();`
		config := parseRPCAnnotations(code)
		assert.True(t, config.DisableLogs)
	})

	t.Run("multiple annotations", func(t *testing.T) {
		code := `-- @fluxbase:description Get user profile with all details
-- @fluxbase:public
-- @fluxbase:timeout 10
-- @fluxbase:allowed-tables users, user_stats, profiles
-- @fluxbase:allowed-schemas public
SELECT u.*, s.total_posts, p.bio
FROM users u
LEFT JOIN user_stats s ON u.id = s.user_id
LEFT JOIN profiles p ON u.id = p.user_id
WHERE u.id = $1;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, "Get user profile with all details", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 10, config.Timeout)
		assert.Equal(t, []string{"users", "user_stats", "profiles"}, config.AllowedTables)
		assert.Equal(t, []string{"public"}, config.AllowedSchemas)
	})

	t.Run("block comment style", func(t *testing.T) {
		code := `/*
 * @fluxbase:description Block comment procedure
 * @fluxbase:public
 * @fluxbase:timeout 45
 */
SELECT * FROM data;`
		config := parseRPCAnnotations(code)
		assert.Equal(t, "Block comment procedure", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 45, config.Timeout)
	})

	t.Run("case insensitivity for annotations", func(t *testing.T) {
		code := `-- @fluxbase:DESCRIPTION Uppercase test
-- @fluxbase:PUBLIC
-- @fluxbase:Timeout 20`
		config := parseRPCAnnotations(code)
		assert.Equal(t, "Uppercase test", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 20, config.Timeout)
	})
}

func TestParseCommaSeparatedList(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		result := loader.ParseCommaList("")
		assert.Nil(t, result)
	})

	t.Run("single item", func(t *testing.T) {
		result := loader.ParseCommaList("users")
		assert.Equal(t, []string{"users"}, result)
	})

	t.Run("multiple items", func(t *testing.T) {
		result := loader.ParseCommaList("users, orders, products")
		assert.Equal(t, []string{"users", "orders", "products"}, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		result := loader.ParseCommaList("  users  ,  orders  ,  products  ")
		assert.Equal(t, []string{"users", "orders", "products"}, result)
	})

	t.Run("filters empty items", func(t *testing.T) {
		result := loader.ParseCommaList("users,,orders,  ,products")
		assert.Equal(t, []string{"users", "orders", "products"}, result)
	})

	t.Run("no commas", func(t *testing.T) {
		result := loader.ParseCommaList("singlevalue")
		assert.Equal(t, []string{"singlevalue"}, result)
	})
}
