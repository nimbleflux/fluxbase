package tools

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

func TestNewSyncChatbotTool(t *testing.T) {
	t.Run("creates tool with nil storage", func(t *testing.T) {
		tool := NewSyncChatbotTool(nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.storage)
	})
}

func TestSyncChatbotTool_Metadata(t *testing.T) {
	tool := NewSyncChatbotTool(nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "sync_chatbot", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Deploy or update an AI chatbot")
		assert.Contains(t, desc, "@fluxbase:description")
		assert.Contains(t, desc, "@fluxbase:allowed-tables")
		assert.Contains(t, desc, "@fluxbase:mcp-tools")
		assert.Contains(t, desc, "@fluxbase:rate-limit")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "name")
		assert.Contains(t, props, "code")
		assert.Contains(t, props, "namespace")

		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "code")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeSyncChatbots)
	})
}

func TestSyncChatbotTool_Execute_Validation(t *testing.T) {
	tool := NewSyncChatbotTool(nil)

	t.Run("missing name", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"code": "system prompt",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("missing code", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "test-bot",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "code is required")
	})

	t.Run("invalid name format", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "invalid name with spaces!",
			"code": "system prompt",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid chatbot name")
	})
}

func TestChatbotToolConfig_Struct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var config ChatbotToolConfig
		assert.Empty(t, config.Description)
		assert.Nil(t, config.AllowedTables)
		assert.Nil(t, config.AllowedOperations)
		assert.False(t, config.IsPublic)
		assert.False(t, config.AllowUnauthenticated)
		assert.Equal(t, 0, config.MaxTokens)
		assert.Equal(t, 0.0, config.Temperature)
	})

	t.Run("all fields", func(t *testing.T) {
		config := ChatbotToolConfig{
			Description:          "Test bot",
			AllowedTables:        []string{"users", "orders"},
			AllowedOperations:    []string{"SELECT", "INSERT"},
			AllowedSchemas:       []string{"public", "analytics"},
			HTTPAllowedDomains:   []string{"api.example.com"},
			MCPTools:             []string{"query_table", "insert_record"},
			UseMCPSchema:         true,
			IsPublic:             true,
			AllowUnauthenticated: true,
			RequireRoles:         []string{"admin", "user"},
			Model:                "gpt-4",
			MaxTokens:            4096,
			Temperature:          0.7,
			RateLimitPerMinute:   30,
			DailyRequestLimit:    1000,
			DailyTokenBudget:     50000,
			PersistConversations: true,
			ConversationTTLHours: 48,
			MaxTurns:             100,
			ResponseLanguage:     "en",
			DisableLogs:          true,
		}

		assert.Equal(t, "Test bot", config.Description)
		assert.Len(t, config.AllowedTables, 2)
		assert.Len(t, config.MCPTools, 2)
		assert.True(t, config.UseMCPSchema)
		assert.Equal(t, 4096, config.MaxTokens)
	})
}

func TestParseChatbotAnnotations(t *testing.T) {
	t.Run("empty code", func(t *testing.T) {
		config := parseChatbotAnnotations("")
		assert.Empty(t, config.Description)
		assert.Equal(t, []string{"SELECT"}, config.AllowedOperations)
		assert.Equal(t, []string{"public"}, config.AllowedSchemas)
		assert.Equal(t, 4096, config.MaxTokens)
		assert.Equal(t, 0.7, config.Temperature)
		assert.Equal(t, 20, config.RateLimitPerMinute)
		assert.Equal(t, 500, config.DailyRequestLimit)
	})

	t.Run("description annotation", func(t *testing.T) {
		code := `// @fluxbase:description Customer support assistant
You are a helpful assistant.`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Customer support assistant", config.Description)
	})

	t.Run("allowed tables", func(t *testing.T) {
		code := `// @fluxbase:allowed-tables users, orders, products`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"users", "orders", "products"}, config.AllowedTables)
	})

	t.Run("allowed tables with schema prefix", func(t *testing.T) {
		code := `// @fluxbase:allowed-tables public.users, analytics.metrics`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"public.users", "analytics.metrics"}, config.AllowedTables)
	})

	t.Run("allowed operations", func(t *testing.T) {
		code := `// @fluxbase:allowed-operations select, insert, update`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"SELECT", "INSERT", "UPDATE"}, config.AllowedOperations)
	})

	t.Run("allowed schemas", func(t *testing.T) {
		code := `// @fluxbase:allowed-schemas public, analytics, reporting`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"public", "analytics", "reporting"}, config.AllowedSchemas)
	})

	t.Run("mcp tools", func(t *testing.T) {
		code := `// @fluxbase:mcp-tools query_table, insert_record, invoke_function`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"query_table", "insert_record", "invoke_function"}, config.MCPTools)
	})

	t.Run("use mcp schema - no value", func(t *testing.T) {
		code := `// @fluxbase:use-mcp-schema`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.UseMCPSchema)
	})

	t.Run("use mcp schema - explicit true", func(t *testing.T) {
		code := `// @fluxbase:use-mcp-schema true`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.UseMCPSchema)
	})

	t.Run("public annotation", func(t *testing.T) {
		code := `// @fluxbase:public`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.IsPublic)
	})

	t.Run("allow unauthenticated", func(t *testing.T) {
		code := `// @fluxbase:allow-unauthenticated`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.AllowUnauthenticated)
	})

	t.Run("require role", func(t *testing.T) {
		code := `// @fluxbase:require-role admin, support`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"admin", "support"}, config.RequireRoles)
	})

	t.Run("model annotation", func(t *testing.T) {
		code := `// @fluxbase:model gpt-4-turbo`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "gpt-4-turbo", config.Model)
	})

	t.Run("max tokens", func(t *testing.T) {
		code := `// @fluxbase:max-tokens 8192`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 8192, config.MaxTokens)
	})

	t.Run("max tokens - invalid value", func(t *testing.T) {
		code := `// @fluxbase:max-tokens invalid`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 4096, config.MaxTokens) // default
	})

	t.Run("temperature", func(t *testing.T) {
		code := `// @fluxbase:temperature 0.5`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 0.5, config.Temperature)
	})

	t.Run("temperature - out of range", func(t *testing.T) {
		code := `// @fluxbase:temperature 3.0`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 0.7, config.Temperature) // default, 3.0 > 2
	})

	t.Run("rate limit", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 50/min`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 50, config.RateLimitPerMinute)
	})

	t.Run("rate limit - no unit", func(t *testing.T) {
		code := `// @fluxbase:rate-limit 50`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 20, config.RateLimitPerMinute) // default, missing /min
	})

	t.Run("daily limit", func(t *testing.T) {
		code := `// @fluxbase:daily-limit 1000`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 1000, config.DailyRequestLimit)
	})

	t.Run("daily token budget", func(t *testing.T) {
		code := `// @fluxbase:daily-token-budget 200000`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 200000, config.DailyTokenBudget)
	})

	t.Run("persist conversations", func(t *testing.T) {
		code := `// @fluxbase:persist-conversations`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.PersistConversations)
	})

	t.Run("conversation ttl", func(t *testing.T) {
		code := `// @fluxbase:conversation-ttl 72`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 72, config.ConversationTTLHours)
	})

	t.Run("max turns", func(t *testing.T) {
		code := `// @fluxbase:max-turns 100`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, 100, config.MaxTurns)
	})

	t.Run("response language", func(t *testing.T) {
		code := `// @fluxbase:response-language es`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "es", config.ResponseLanguage)
	})

	t.Run("disable logs", func(t *testing.T) {
		code := `// @fluxbase:disable-logs`
		config := parseChatbotAnnotations(code)
		assert.True(t, config.DisableLogs)
	})

	t.Run("http allowed domains", func(t *testing.T) {
		code := `// @fluxbase:http-allowed-domains api.example.com, webhook.example.com`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, []string{"api.example.com", "webhook.example.com"}, config.HTTPAllowedDomains)
	})

	t.Run("multiple annotations", func(t *testing.T) {
		code := `// @fluxbase:description Sales assistant
// @fluxbase:allowed-tables orders, customers
// @fluxbase:allowed-operations SELECT, INSERT
// @fluxbase:mcp-tools query_table, insert_record
// @fluxbase:use-mcp-schema
// @fluxbase:public
// @fluxbase:persist-conversations
// @fluxbase:rate-limit 30/min
// @fluxbase:model gpt-4

You are a helpful sales assistant.`

		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Sales assistant", config.Description)
		assert.Equal(t, []string{"orders", "customers"}, config.AllowedTables)
		assert.Equal(t, []string{"SELECT", "INSERT"}, config.AllowedOperations)
		assert.Equal(t, []string{"query_table", "insert_record"}, config.MCPTools)
		assert.True(t, config.UseMCPSchema)
		assert.True(t, config.IsPublic)
		assert.True(t, config.PersistConversations)
		assert.Equal(t, 30, config.RateLimitPerMinute)
		assert.Equal(t, "gpt-4", config.Model)
	})

	t.Run("block comment style", func(t *testing.T) {
		code := `/*
 * @fluxbase:description Block comment bot
 * @fluxbase:public
 * @fluxbase:max-tokens 2048
 */`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Block comment bot", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 2048, config.MaxTokens)
	})

	t.Run("case insensitivity for annotations", func(t *testing.T) {
		code := `// @fluxbase:DESCRIPTION Uppercase test
// @fluxbase:PUBLIC
// @fluxbase:Max-Tokens 1024`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Uppercase test", config.Description)
		assert.True(t, config.IsPublic)
		assert.Equal(t, 1024, config.MaxTokens)
	})

	t.Run("ignores non-annotation comments", func(t *testing.T) {
		code := `// This is a regular comment
// @fluxbase:description Real annotation
// Another regular comment
func main() {}`
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Real annotation", config.Description)
	})

	t.Run("handles extra whitespace", func(t *testing.T) {
		code := `//   @fluxbase:description    Whitespace test
//  @fluxbase:allowed-tables   users  ,  orders  ,  products  `
		config := parseChatbotAnnotations(code)
		assert.Equal(t, "Whitespace test", config.Description)
		assert.Equal(t, []string{"users", "orders", "products"}, config.AllowedTables)
	})
}
