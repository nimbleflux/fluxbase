package ai

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseChatbotConfig_IntentRules(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:allowed-tables my_trips,my_place_visits\n" +
		" * @fluxbase:intent-rules [{\"keywords\":[\"restaurant\",\"cafe\"],\"requiredTable\":\"my_place_visits\",\"forbiddenTable\":\"my_trips\"},{\"keywords\":[\"trip\",\"travel\"],\"requiredTable\":\"my_trips\"}]\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	assert.Len(t, config.IntentRules, 2)

	// First rule
	assert.Equal(t, []string{"restaurant", "cafe"}, config.IntentRules[0].Keywords)
	assert.Equal(t, "my_place_visits", config.IntentRules[0].RequiredTable)
	assert.Equal(t, "my_trips", config.IntentRules[0].ForbiddenTable)

	// Second rule
	assert.Equal(t, []string{"trip", "travel"}, config.IntentRules[1].Keywords)
	assert.Equal(t, "my_trips", config.IntentRules[1].RequiredTable)
	assert.Equal(t, "", config.IntentRules[1].ForbiddenTable)
}

func TestParseChatbotConfig_RequiredColumns(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:required-columns my_trips=id,title,image_url my_place_visits=poi_name,city\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	assert.Len(t, config.RequiredColumns, 2)
	assert.Equal(t, []string{"id", "title", "image_url"}, config.RequiredColumns["my_trips"])
	assert.Equal(t, []string{"poi_name", "city"}, config.RequiredColumns["my_place_visits"])
}

func TestParseChatbotConfig_DefaultTable(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:default-table my_place_visits\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	assert.Equal(t, "my_place_visits", config.DefaultTable)
}

func TestParseChatbotConfig_AllIntentAnnotations(t *testing.T) {
	code := "/**\n" +
		" * Location Assistant\n" +
		" *\n" +
		" * @fluxbase:allowed-tables my_trips,my_place_visits,my_poi_summary\n" +
		" * @fluxbase:allowed-operations SELECT\n" +
		" * @fluxbase:default-table my_place_visits\n" +
		" * @fluxbase:required-columns my_trips=id,title,image_url\n" +
		" * @fluxbase:intent-rules [{\"keywords\":[\"restaurant\",\"cafe\",\"food\"],\"requiredTable\":\"my_place_visits\",\"forbiddenTable\":\"my_trips\"}]\n" +
		" */\n" +
		"\n" +
		"export default `You are a location assistant.`;\n"

	config := ParseChatbotConfig(code)

	// Check all intent-related fields
	assert.Equal(t, "my_place_visits", config.DefaultTable)
	assert.Len(t, config.IntentRules, 1)
	assert.Equal(t, []string{"restaurant", "cafe", "food"}, config.IntentRules[0].Keywords)
	assert.Len(t, config.RequiredColumns, 1)
	assert.Equal(t, []string{"id", "title", "image_url"}, config.RequiredColumns["my_trips"])
}

func TestParseChatbotConfig_NoIntentAnnotations(t *testing.T) {
	code := "/**\n" +
		" * Simple chatbot\n" +
		" *\n" +
		" * @fluxbase:allowed-tables users\n" +
		" * @fluxbase:allowed-operations SELECT\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	// Should be nil/empty when no intent annotations
	assert.Nil(t, config.IntentRules)
	assert.Nil(t, config.RequiredColumns)
	assert.Equal(t, "", config.DefaultTable)
}

func TestParseChatbotConfig_InvalidIntentRulesJSON(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:intent-rules not-valid-json\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	// Should be nil when JSON is invalid
	assert.Nil(t, config.IntentRules)
}

func TestParseRequiredColumns(t *testing.T) {
	// Single table
	result := parseRequiredColumns("my_trips=id,title,image_url")
	assert.Len(t, result, 1)
	assert.Equal(t, []string{"id", "title", "image_url"}, result["my_trips"])

	// Multiple tables
	result = parseRequiredColumns("my_trips=id,title my_places=name,city")
	assert.Len(t, result, 2)
	assert.Equal(t, []string{"id", "title"}, result["my_trips"])
	assert.Equal(t, []string{"name", "city"}, result["my_places"])

	// Empty input
	result = parseRequiredColumns("")
	assert.Len(t, result, 0)

	// Invalid format (no equals sign)
	result = parseRequiredColumns("invalid-format")
	assert.Len(t, result, 0)
}

func TestApplyConfig_IntentFields(t *testing.T) {
	config := ChatbotConfig{
		IntentRules: []IntentRule{
			{Keywords: []string{"test"}, RequiredTable: "test_table"},
		},
		RequiredColumns: RequiredColumnsMap{
			"table1": {"col1", "col2"},
		},
		DefaultTable: "default_table",
	}

	chatbot := &Chatbot{}
	chatbot.ApplyConfig(config)

	assert.Len(t, chatbot.IntentRules, 1)
	assert.Equal(t, "test_table", chatbot.IntentRules[0].RequiredTable)
	assert.Equal(t, []string{"col1", "col2"}, chatbot.RequiredColumns["table1"])
	assert.Equal(t, "default_table", chatbot.DefaultTable)
}

func TestParseChatbotConfig_MultipleRequiredColumnsAnnotations(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:required-columns my_trips=id,title,image_url\n" +
		" * @fluxbase:required-columns my_place_visits=poi_name,city\n" +
		" * @fluxbase:required-columns my_poi_summary=category,count\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	assert.Len(t, config.RequiredColumns, 3)
	assert.Equal(t, []string{"id", "title", "image_url"}, config.RequiredColumns["my_trips"])
	assert.Equal(t, []string{"poi_name", "city"}, config.RequiredColumns["my_place_visits"])
	assert.Equal(t, []string{"category", "count"}, config.RequiredColumns["my_poi_summary"])
}

func TestParseChatbotConfig_MultipleIntentRulesAnnotations(t *testing.T) {
	code := "/**\n" +
		" * Test chatbot\n" +
		" *\n" +
		" * @fluxbase:intent-rules [{\"keywords\":[\"restaurant\",\"cafe\"],\"requiredTable\":\"my_place_visits\"}]\n" +
		" * @fluxbase:intent-rules [{\"keywords\":[\"trip\",\"travel\"],\"requiredTable\":\"my_trips\"}]\n" +
		" */\n" +
		"\n" +
		"export default `You are a helpful assistant.`;\n"

	config := ParseChatbotConfig(code)

	assert.Len(t, config.IntentRules, 2)

	// First annotation
	assert.Equal(t, []string{"restaurant", "cafe"}, config.IntentRules[0].Keywords)
	assert.Equal(t, "my_place_visits", config.IntentRules[0].RequiredTable)

	// Second annotation
	assert.Equal(t, []string{"trip", "travel"}, config.IntentRules[1].Keywords)
	assert.Equal(t, "my_trips", config.IntentRules[1].RequiredTable)
}

func TestParseChatbotConfig_ResponseLanguage(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "auto (default)",
			code: "/**\n" +
				" * Test chatbot\n" +
				" *\n" +
				" * @fluxbase:allowed-tables users\n" +
				" */\n" +
				"\n" +
				"export default `You are a helpful assistant.`;\n",
			expected: "auto",
		},
		{
			name: "explicit auto",
			code: "/**\n" +
				" * Test chatbot\n" +
				" *\n" +
				" * @fluxbase:response-language auto\n" +
				" */\n" +
				"\n" +
				"export default `You are a helpful assistant.`;\n",
			expected: "auto",
		},
		{
			name: "ISO code",
			code: "/**\n" +
				" * Test chatbot\n" +
				" *\n" +
				" * @fluxbase:response-language de\n" +
				" */\n" +
				"\n" +
				"export default `You are a helpful assistant.`;\n",
			expected: "de",
		},
		{
			name: "language name English",
			code: "/**\n" +
				" * Test chatbot\n" +
				" *\n" +
				" * @fluxbase:response-language German\n" +
				" */\n" +
				"\n" +
				"export default `You are a helpful assistant.`;\n",
			expected: "German",
		},
		{
			name: "language name native",
			code: "/**\n" +
				" * Test chatbot\n" +
				" *\n" +
				" * @fluxbase:response-language Deutsch\n" +
				" */\n" +
				"\n" +
				"export default `You are a helpful assistant.`;\n",
			expected: "Deutsch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseChatbotConfig(tt.code)
			assert.Equal(t, tt.expected, config.ResponseLanguage)
		})
	}
}

func TestApplyConfig_ResponseLanguage(t *testing.T) {
	config := ChatbotConfig{
		ResponseLanguage: "German",
	}

	chatbot := &Chatbot{}
	chatbot.ApplyConfig(config)

	assert.Equal(t, "German", chatbot.ResponseLanguage)
}

func TestParseChatbotConfig_MCPTools(t *testing.T) {
	t.Run("parses mcp-tools annotation", func(t *testing.T) {
		code := "/**\n" +
			" * Test chatbot\n" +
			" *\n" +
			" * @fluxbase:mcp-tools query_table,insert_record,invoke_function\n" +
			" */\n" +
			"\n" +
			"export default `You are a helpful assistant.`;\n"

		config := ParseChatbotConfig(code)

		assert.Equal(t, []string{"query_table", "insert_record", "invoke_function"}, config.MCPTools)
	})

	t.Run("parses use-mcp-schema annotation", func(t *testing.T) {
		code := "/**\n" +
			" * Test chatbot\n" +
			" *\n" +
			" * @fluxbase:use-mcp-schema\n" +
			" */\n" +
			"\n" +
			"export default `You are a helpful assistant.`;\n"

		config := ParseChatbotConfig(code)

		assert.True(t, config.UseMCPSchema)
	})

	t.Run("use-mcp-schema with true value", func(t *testing.T) {
		code := "/**\n" +
			" * Test chatbot\n" +
			" *\n" +
			" * @fluxbase:use-mcp-schema true\n" +
			" */\n" +
			"\n" +
			"export default `You are a helpful assistant.`;\n"

		config := ParseChatbotConfig(code)

		assert.True(t, config.UseMCPSchema)
	})

	t.Run("defaults to empty/false for MCP fields", func(t *testing.T) {
		code := "/**\n" +
			" * Test chatbot\n" +
			" *\n" +
			" * @fluxbase:allowed-tables users\n" +
			" */\n" +
			"\n" +
			"export default `You are a helpful assistant.`;\n"

		config := ParseChatbotConfig(code)

		assert.Empty(t, config.MCPTools)
		assert.False(t, config.UseMCPSchema)
	})
}

func TestApplyConfig_MCPFields(t *testing.T) {
	config := ChatbotConfig{
		MCPTools:     []string{"query_table", "insert_record"},
		UseMCPSchema: true,
	}

	chatbot := &Chatbot{}
	chatbot.ApplyConfig(config)

	assert.Equal(t, []string{"query_table", "insert_record"}, chatbot.MCPTools)
	assert.True(t, chatbot.UseMCPSchema)
}

func TestChatbot_HasMCPTools(t *testing.T) {
	t.Run("returns true when MCP tools configured", func(t *testing.T) {
		chatbot := &Chatbot{
			MCPTools: []string{"query_table", "insert_record"},
		}
		assert.True(t, chatbot.HasMCPTools())
	})

	t.Run("returns false when MCP tools empty", func(t *testing.T) {
		chatbot := &Chatbot{
			MCPTools: []string{},
		}
		assert.False(t, chatbot.HasMCPTools())
	})

	t.Run("returns false when MCP tools nil", func(t *testing.T) {
		chatbot := &Chatbot{
			MCPTools: nil,
		}
		assert.False(t, chatbot.HasMCPTools())
	})
}

func TestParseQualifiedTables(t *testing.T) {
	t.Run("simple table names use default schema", func(t *testing.T) {
		result := ParseQualifiedTables([]string{"users", "orders"}, "public")
		assert.Len(t, result, 2)
		assert.Equal(t, "public", result[0].Schema)
		assert.Equal(t, "users", result[0].Table)
		assert.Equal(t, "public", result[1].Schema)
		assert.Equal(t, "orders", result[1].Table)
	})

	t.Run("qualified names extract schema", func(t *testing.T) {
		result := ParseQualifiedTables([]string{"analytics.metrics", "public.users"}, "public")
		assert.Len(t, result, 2)
		assert.Equal(t, "analytics", result[0].Schema)
		assert.Equal(t, "metrics", result[0].Table)
		assert.Equal(t, "public", result[1].Schema)
		assert.Equal(t, "users", result[1].Table)
	})

	t.Run("mixed qualified and simple names", func(t *testing.T) {
		result := ParseQualifiedTables([]string{"users", "analytics.metrics"}, "public")
		assert.Len(t, result, 2)
		assert.Equal(t, "public", result[0].Schema)
		assert.Equal(t, "users", result[0].Table)
		assert.Equal(t, "analytics", result[1].Schema)
		assert.Equal(t, "metrics", result[1].Table)
	})

	t.Run("empty list returns empty", func(t *testing.T) {
		result := ParseQualifiedTables([]string{}, "public")
		assert.Empty(t, result)
	})
}

func TestGroupTablesBySchema(t *testing.T) {
	tables := []QualifiedTable{
		{Schema: "public", Table: "users"},
		{Schema: "public", Table: "orders"},
		{Schema: "analytics", Table: "metrics"},
	}

	result := GroupTablesBySchema(tables)

	assert.Len(t, result["public"], 2)
	assert.Contains(t, result["public"], "users")
	assert.Contains(t, result["public"], "orders")
	assert.Len(t, result["analytics"], 1)
	assert.Contains(t, result["analytics"], "metrics")
}

// =============================================================================
// Additional ParseChatbotConfig Tests
// =============================================================================

func TestParseChatbotConfig_AllAnnotations(t *testing.T) {
	code := "/**\n" +
		" * Comprehensive Test Chatbot\n" +
		" *\n" +
		" * @fluxbase:allowed-tables users,products,orders\n" +
		" * @fluxbase:allowed-operations SELECT,INSERT,UPDATE,DELETE\n" +
		" * @fluxbase:allowed-schemas public,app\n" +
		" * @fluxbase:http-allowed-domains api.example.com,data.example.org\n" +
		" * @fluxbase:max-tokens 8192\n" +
		" * @fluxbase:temperature 0.5\n" +
		" * @fluxbase:model gpt-4-turbo\n" +
		" * @fluxbase:persist-conversations true\n" +
		" * @fluxbase:conversation-ttl 48h\n" +
		" * @fluxbase:max-turns 100\n" +
		" * @fluxbase:rate-limit 30/min\n" +
		" * @fluxbase:daily-limit 1000\n" +
		" * @fluxbase:token-budget 500000/day\n" +
		" * @fluxbase:allow-unauthenticated true\n" +
		" * @fluxbase:public false\n" +
		" * @fluxbase:version 3\n" +
		" * @fluxbase:default-table users\n" +
		" * @fluxbase:required-columns users=id,email,name orders=id,total,status\n" +
		" * @fluxbase:intent-rules [{\"keywords\":[\"user\",\"account\"],\"requiredTable\":\"users\"},{\"keywords\":[\"order\",\"purchase\"],\"requiredTable\":\"orders\"}]\n" +
		" * @fluxbase:knowledge-base docs-base,faq-base\n" +
		" * @fluxbase:rag-max-chunks 10\n" +
		" * @fluxbase:rag-similarity-threshold 0.8\n" +
		" * @fluxbase:rag-table documents\n" +
		" * @fluxbase:rag-column embedding\n" +
		" * @fluxbase:rag-content-column content\n" +
		" * @fluxbase:response-language Spanish\n" +
		" * @fluxbase:disable-execution-logs true\n" +
		" * @fluxbase:required-settings openai.api_key,stripe.endpoint\n" +
		" * @fluxbase:mcp-tools query_table,insert_record\n" +
		" * @fluxbase:use-mcp-schema true\n" +
		" */\n" +
		"\n" +
		"export default `You are a comprehensive test assistant.`;\n"

	config := ParseChatbotConfig(code)

	// Basic access control
	assert.Equal(t, []string{"users", "products", "orders"}, config.AllowedTables)
	assert.Equal(t, []string{"SELECT", "INSERT", "UPDATE", "DELETE"}, config.AllowedOperations)
	assert.Equal(t, []string{"public", "app"}, config.AllowedSchemas)
	assert.Equal(t, []string{"api.example.com", "data.example.org"}, config.HTTPAllowedDomains)

	// Model settings
	assert.Equal(t, 8192, config.MaxTokens)
	assert.Equal(t, 0.5, config.Temperature)
	assert.Equal(t, "gpt-4-turbo", config.Model)

	// Conversation settings
	assert.True(t, config.PersistConversations)
	assert.Equal(t, 48*time.Hour, config.ConversationTTL)
	assert.Equal(t, 100, config.MaxTurns)

	// Rate limiting
	assert.Equal(t, 30, config.RateLimitPerMinute)
	assert.Equal(t, 1000, config.DailyRequestLimit)
	assert.Equal(t, 500000, config.DailyTokenBudget)

	// Access control
	assert.True(t, config.AllowUnauthenticated)
	assert.False(t, config.IsPublic)
	assert.Equal(t, 3, config.Version)

	// Intent validation
	assert.Equal(t, "users", config.DefaultTable)
	assert.Len(t, config.IntentRules, 2)
	assert.Equal(t, []string{"id", "email", "name"}, config.RequiredColumns["users"])
	assert.Equal(t, []string{"id", "total", "status"}, config.RequiredColumns["orders"])

	// RAG/Knowledge Base
	assert.Equal(t, []string{"docs-base", "faq-base"}, config.KnowledgeBases)
	assert.Equal(t, 10, config.RAGMaxChunks)
	assert.Equal(t, 0.8, config.RAGSimilarityThreshold)
	assert.Equal(t, "documents", config.RAGTable)
	assert.Equal(t, "embedding", config.RAGColumn)
	assert.Equal(t, "content", config.RAGContentColumn)

	// Response language
	assert.Equal(t, "Spanish", config.ResponseLanguage)

	// Logging and settings
	assert.True(t, config.DisableExecutionLogs)
	assert.Equal(t, []string{"openai.api_key", "stripe.endpoint"}, config.RequiredSettings)

	// MCP integration
	assert.Equal(t, []string{"query_table", "insert_record"}, config.MCPTools)
	assert.True(t, config.UseMCPSchema)
}

func TestParseChatbotConfig_EdgeCases(t *testing.T) {
	t.Run("handles empty tables list", func(t *testing.T) {
		code := "/**\n" +
			" * Test\n" +
			" * @fluxbase:allowed-tables\n" +
			" */\n" +
			"export default `test`;\n"
		config := ParseChatbotConfig(code)
		assert.Empty(t, config.AllowedTables)
	})

	t.Run("handles whitespace in annotations", func(t *testing.T) {
		code := "/**\n" +
			" * Test\n" +
			" * @fluxbase:allowed-tables   users  ,  products  ,  orders\n" +
			" */\n" +
			"export default `test`;\n"
		config := ParseChatbotConfig(code)
		assert.Equal(t, []string{"users", "products", "orders"}, config.AllowedTables)
	})

	t.Run("handles zero values", func(t *testing.T) {
		code := "/**\n" +
			" * Test\n" +
			" * @fluxbase:max-tokens 0\n" +
			" * @fluxbase:temperature 0\n" +
			" * @fluxbase:daily-limit 0\n" +
			" */\n" +
			"export default `test`;\n"
		config := ParseChatbotConfig(code)
		assert.Equal(t, 0, config.MaxTokens)
		assert.Equal(t, 0.0, config.Temperature)
		assert.Equal(t, 0, config.DailyRequestLimit)
	})

	t.Run("handles temperature above 2", func(t *testing.T) {
		code := "/**\n" +
			" * Test\n" +
			" * @fluxbase:temperature 2.5\n" +
			" */\n" +
			"export default `test`;\n"
		config := ParseChatbotConfig(code)
		assert.Equal(t, 2.5, config.Temperature)
	})
}

func TestParseSystemPrompt(t *testing.T) {
	t.Run("extracts system prompt from export default", func(t *testing.T) {
		code := "export default `You are a helpful assistant.`;"
		prompt := ParseSystemPrompt(code)
		assert.Equal(t, "You are a helpful assistant.", prompt)
	})

	t.Run("handles multiline system prompt", func(t *testing.T) {
		code := "export default `You are a helpful assistant.\nYou provide concise answers.\nYou are friendly.`;"
		prompt := ParseSystemPrompt(code)
		assert.Contains(t, prompt, "You are a helpful assistant.")
		assert.Contains(t, prompt, "You provide concise answers.")
		assert.Contains(t, prompt, "You are friendly.")
	})

	t.Run("handles empty code", func(t *testing.T) {
		prompt := ParseSystemPrompt("")
		assert.Equal(t, "", prompt)
	})

	t.Run("handles code without export default", func(t *testing.T) {
		code := "const x = 42;"
		prompt := ParseSystemPrompt(code)
		assert.Equal(t, "", prompt)
	})

	t.Run("handles system prompt with special characters", func(t *testing.T) {
		code := "export default `System prompt with special characters`;"
		prompt := ParseSystemPrompt(code)
		assert.Contains(t, prompt, "special characters")
	})
}

func TestParseDescription(t *testing.T) {
	t.Run("extracts description from JSDoc", func(t *testing.T) {
		code := "/**\n" +
			" * This is a test chatbot description\n" +
			" *\n" +
			" * @fluxbase:allowed-tables users\n" +
			" */\n" +
			"export default `test`;\n"
		desc := ParseDescription(code)
		assert.Equal(t, "This is a test chatbot description", desc)
	})

	t.Run("handles empty description", func(t *testing.T) {
		code := "/**\n" +
			" * @fluxbase:allowed-tables users\n" +
			" */\n" +
			"export default `test`;\n"
		desc := ParseDescription(code)
		assert.Equal(t, "", desc)
	})

	t.Run("handles multiline description", func(t *testing.T) {
		code := "/**\n" +
			" * Line 1\n" +
			" * Line 2\n" +
			" * Line 3\n" +
			" */\n" +
			"export default `test`;\n"
		desc := ParseDescription(code)
		assert.Equal(t, "Line 1", desc)
	})

	t.Run("handles empty code", func(t *testing.T) {
		desc := ParseDescription("")
		assert.Equal(t, "", desc)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		code := "/**\n" +
			" *   Description with extra spaces\n" +
			" *\n" +
			" * @fluxbase:allowed-tables users\n" +
			" */\n" +
			"export default `test`;\n"
		desc := ParseDescription(code)
		assert.Equal(t, "Description with extra spaces", desc)
	})
}

func TestExtractBalancedJSON(t *testing.T) {
	t.Run("extracts simple JSON array", func(t *testing.T) {
		input := `[{"key":"value"}]`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, `[{"key":"value"}]`, result)
	})

	t.Run("extracts nested JSON", func(t *testing.T) {
		input := `[{"outer":{"inner":"value"}}]`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, `[{"outer":{"inner":"value"}}]`, result)
	})

	t.Run("extracts JSON with strings containing brackets", func(t *testing.T) {
		input := `[{"text":"hello [world]"}]`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, `[{"text":"hello [world]"}]`, result)
	})

	t.Run("extracts JSON with escaped quotes", func(t *testing.T) {
		input := `[{"text":"hello \"world\""}]`
		result := extractBalancedJSON(input, 0)
		// The function correctly handles escaped quotes and returns the full array
		assert.Equal(t, `[{"text":"hello \"world\""}]`, result)
	})

	t.Run("returns empty for unbalanced JSON", func(t *testing.T) {
		input := `[{"key":"value"}`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty for invalid start position", func(t *testing.T) {
		input := `not an array`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, "", result)
	})

	t.Run("handles empty array", func(t *testing.T) {
		input := `[]`
		result := extractBalancedJSON(input, 0)
		assert.Equal(t, `[]`, result)
	})

	t.Run("extracts from middle of string", func(t *testing.T) {
		input := `prefix [{"key":"value"}] suffix`
		result := extractBalancedJSON(input, 7)
		assert.Equal(t, `[{"key":"value"}]`, result)
	})
}

func TestParseCSV(t *testing.T) {
	t.Run("parses simple CSV", func(t *testing.T) {
		result := parseCSV("a,b,c")
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("trims whitespace", func(t *testing.T) {
		result := parseCSV(" a , b , c ")
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := parseCSV("")
		assert.Empty(t, result)
	})

	t.Run("handles only whitespace", func(t *testing.T) {
		result := parseCSV("   ,  ,  ")
		assert.Empty(t, result)
	})

	t.Run("handles single value", func(t *testing.T) {
		result := parseCSV("single")
		assert.Equal(t, []string{"single"}, result)
	})

	t.Run("handles trailing comma", func(t *testing.T) {
		result := parseCSV("a,b,")
		assert.Equal(t, []string{"a", "b"}, result)
	})

	t.Run("handles leading comma", func(t *testing.T) {
		result := parseCSV(",a,b")
		assert.Equal(t, []string{"a", "b"}, result)
	})
}

func TestChatbot_ToSummary(t *testing.T) {
	now := time.Now()

	chatbot := &Chatbot{
		ID:          "test-id",
		Name:        "Test Chatbot",
		Namespace:   "test-ns",
		Description: "Test Description",
		Model:       "gpt-4",
		Enabled:     true,
		IsPublic:    false,
		Source:      "filesystem",
		UpdatedAt:   now,
	}

	summary := chatbot.ToSummary()

	assert.Equal(t, "test-id", summary.ID)
	assert.Equal(t, "Test Chatbot", summary.Name)
	assert.Equal(t, "test-ns", summary.Namespace)
	assert.Equal(t, "Test Description", summary.Description)
	assert.Equal(t, "gpt-4", summary.Model)
	assert.True(t, summary.Enabled)
	assert.False(t, summary.IsPublic)
	assert.Equal(t, "filesystem", summary.Source)
	assert.Equal(t, now.Format(time.RFC3339), summary.UpdatedAt)
}

func TestChatbot_PopulateDerivedFields(t *testing.T) {
	t.Run("populates model from code when not set", func(t *testing.T) {
		code := "/**\n" +
			" * @fluxbase:model gpt-4-turbo\n" +
			" */\n" +
			"export default `test`;\n"

		chatbot := &Chatbot{
			Code:   code,
			Model:  "",
			Name:   "test",
			ID:     "test-id",
			Source: "api",
		}
		chatbot.PopulateDerivedFields()

		assert.Equal(t, "gpt-4-turbo", chatbot.Model)
	})

	t.Run("does not override existing model", func(t *testing.T) {
		code := "/**\n" +
			" * @fluxbase:model gpt-4-turbo\n" +
			" */\n" +
			"export default `test`;\n"

		chatbot := &Chatbot{
			Code:   code,
			Model:  "gpt-3.5-turbo",
			Name:   "test",
			ID:     "test-id",
			Source: "api",
		}
		chatbot.PopulateDerivedFields()

		assert.Equal(t, "gpt-3.5-turbo", chatbot.Model)
	})

	t.Run("handles empty code", func(t *testing.T) {
		chatbot := &Chatbot{
			Code:   "",
			Model:  "",
			Name:   "test",
			ID:     "test-id",
			Source: "api",
		}
		chatbot.PopulateDerivedFields()

		assert.Equal(t, "", chatbot.Model)
	})

	t.Run("handles code without model annotation", func(t *testing.T) {
		code := "/**\n" +
			" * @fluxbase:allowed-tables users\n" +
			" */\n" +
			"export default `test`;\n"

		chatbot := &Chatbot{
			Code:   code,
			Model:  "",
			Name:   "test",
			ID:     "test-id",
			Source: "api",
		}
		chatbot.PopulateDerivedFields()

		assert.Equal(t, "", chatbot.Model)
	})
}

func TestParseQualifiedTables_EdgeCases(t *testing.T) {
	t.Run("handles tables with multiple dots", func(t *testing.T) {
		// Only first dot separates schema from table
		result := ParseQualifiedTables([]string{"schema.table.part"}, "public")
		assert.Len(t, result, 1)
		assert.Equal(t, "schema", result[0].Schema)
		assert.Equal(t, "table.part", result[0].Table)
	})

	t.Run("handles empty default schema", func(t *testing.T) {
		result := ParseQualifiedTables([]string{"users"}, "")
		assert.Equal(t, "public", result[0].Schema) // Should default to public
		assert.Equal(t, "users", result[0].Table)
	})

	t.Run("handles table starting with dot", func(t *testing.T) {
		result := ParseQualifiedTables([]string{".hidden"}, "public")
		assert.Equal(t, "", result[0].Schema)
		assert.Equal(t, "hidden", result[0].Table)
	})

	t.Run("handles table ending with dot", func(t *testing.T) {
		result := ParseQualifiedTables([]string{"public."}, "default")
		assert.Equal(t, "public", result[0].Schema)
		assert.Equal(t, "", result[0].Table)
	})
}

func TestApplyConfig_FullConfig(t *testing.T) {
	config := ChatbotConfig{
		AllowedTables:      []string{"users", "products"},
		AllowedOperations:  []string{"SELECT", "INSERT"},
		AllowedSchemas:     []string{"public", "app"},
		HTTPAllowedDomains: []string{"api.example.com"},
		IntentRules: []IntentRule{
			{Keywords: []string{"test"}, RequiredTable: "test_table"},
		},
		RequiredColumns: RequiredColumnsMap{
			"users": {"id", "email"},
		},
		DefaultTable:         "users",
		MaxTokens:            4096,
		Temperature:          0.7,
		Model:                "gpt-4",
		PersistConversations: true,
		ConversationTTL:      24 * time.Hour,
		MaxTurns:             50,
		RateLimitPerMinute:   20,
		DailyRequestLimit:    500,
		DailyTokenBudget:     100000,
		AllowUnauthenticated: true,
		IsPublic:             false,
		ResponseLanguage:     "German",
		DisableExecutionLogs: true,
		RequiredSettings:     []string{"api_key"},
		MCPTools:             []string{"query_table"},
		UseMCPSchema:         true,
		Version:              2,
	}

	chatbot := &Chatbot{}
	chatbot.ApplyConfig(config)

	assert.Equal(t, []string{"users", "products"}, chatbot.AllowedTables)
	assert.Equal(t, []string{"SELECT", "INSERT"}, chatbot.AllowedOperations)
	assert.Equal(t, []string{"public", "app"}, chatbot.AllowedSchemas)
	assert.Equal(t, []string{"api.example.com"}, chatbot.HTTPAllowedDomains)
	assert.Len(t, chatbot.IntentRules, 1)
	assert.Equal(t, "test_table", chatbot.IntentRules[0].RequiredTable)
	assert.Equal(t, []string{"id", "email"}, chatbot.RequiredColumns["users"])
	assert.Equal(t, "users", chatbot.DefaultTable)
	assert.Equal(t, 4096, chatbot.MaxTokens)
	assert.Equal(t, 0.7, chatbot.Temperature)
	assert.Equal(t, "gpt-4", chatbot.Model)
	assert.True(t, chatbot.PersistConversations)
	assert.Equal(t, 24, chatbot.ConversationTTLHours)
	assert.Equal(t, 50, chatbot.MaxConversationTurns)
	assert.Equal(t, 20, chatbot.RateLimitPerMinute)
	assert.Equal(t, 500, chatbot.DailyRequestLimit)
	assert.Equal(t, 100000, chatbot.DailyTokenBudget)
	assert.True(t, chatbot.AllowUnauthenticated)
	assert.False(t, chatbot.IsPublic)
	assert.Equal(t, "German", chatbot.ResponseLanguage)
	assert.True(t, chatbot.DisableExecutionLogs)
	assert.Equal(t, []string{"api_key"}, chatbot.RequiredSettings)
	assert.Equal(t, []string{"query_table"}, chatbot.MCPTools)
	assert.True(t, chatbot.UseMCPSchema)
	assert.Equal(t, 2, chatbot.Version)
}
