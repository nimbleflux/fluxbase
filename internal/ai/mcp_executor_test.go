package ai

import (
	"encoding/json"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// isDataTool Tests
// =============================================================================

func TestIsDataTool(t *testing.T) {
	dataTools := []string{
		"query_table",
		"insert_record",
		"update_record",
		"delete_record",
		"search_vectors",
	}

	for _, tool := range dataTools {
		t.Run(tool+" is a data tool", func(t *testing.T) {
			assert.True(t, isDataTool(tool))
		})
	}

	nonDataTools := []string{
		"execute_sql",
		"invoke_function",
		"invoke_rpc",
		"submit_job",
		"get_job_status",
		"http_request",
		"list_objects",
		"download_object",
		"upload_object",
		"delete_object",
		"create_table",
		"alter_table",
		"drop_table",
		"list_branches",
		"create_branch",
		"",
		"unknown_tool",
	}

	for _, tool := range nonDataTools {
		t.Run(tool+" is not a data tool", func(t *testing.T) {
			assert.False(t, isDataTool(tool))
		})
	}
}

// =============================================================================
// extractResultContent Tests
// =============================================================================

func TestExtractResultContent(t *testing.T) {
	t.Run("nil result returns empty string", func(t *testing.T) {
		result := extractResultContent(nil)
		assert.Equal(t, "", result)
	})

	t.Run("empty content returns empty string", func(t *testing.T) {
		result := extractResultContent(&mcp.ToolResult{
			Content: []mcp.Content{},
		})
		assert.Equal(t, "", result)
	})

	t.Run("single text content", func(t *testing.T) {
		result := extractResultContent(&mcp.ToolResult{
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: "Hello world"},
			},
		})
		assert.Equal(t, "Hello world", result)
	})

	t.Run("multiple text contents joined with newlines", func(t *testing.T) {
		result := extractResultContent(&mcp.ToolResult{
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: "Line 1"},
				{Type: mcp.ContentTypeText, Text: "Line 2"},
				{Type: mcp.ContentTypeText, Text: "Line 3"},
			},
		})
		assert.Equal(t, "Line 1\nLine 2\nLine 3", result)
	})

	t.Run("skips non-text content", func(t *testing.T) {
		result := extractResultContent(&mcp.ToolResult{
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: "Text content"},
				{Type: "image", Text: ""},
				{Type: mcp.ContentTypeText, Text: "More text"},
			},
		})
		assert.Equal(t, "Text content\nMore text", result)
	})

	t.Run("skips empty text", func(t *testing.T) {
		result := extractResultContent(&mcp.ToolResult{
			Content: []mcp.Content{
				{Type: mcp.ContentTypeText, Text: ""},
				{Type: mcp.ContentTypeText, Text: "Non-empty"},
			},
		})
		assert.Equal(t, "Non-empty", result)
	})
}

// =============================================================================
// ToAnthropicFormat Tests
// =============================================================================

func TestToAnthropicFormat(t *testing.T) {
	t.Run("empty tools list", func(t *testing.T) {
		result := ToAnthropicFormat([]ToolDefinition{})
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("nil tools list", func(t *testing.T) {
		result := ToAnthropicFormat(nil)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("single tool", func(t *testing.T) {
		tools := []ToolDefinition{
			{
				Name:        "query_table",
				Description: "Query a database table",
				Parameters: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"table": map[string]any{
							"type": "string",
						},
					},
				},
			},
		}

		result := ToAnthropicFormat(tools)

		require.Len(t, result, 1)
		assert.Equal(t, "query_table", result[0]["name"])
		assert.Equal(t, "Query a database table", result[0]["description"])
		assert.NotNil(t, result[0]["input_schema"])
	})

	t.Run("multiple tools", func(t *testing.T) {
		tools := []ToolDefinition{
			{Name: "tool1", Description: "First tool", Parameters: map[string]any{}},
			{Name: "tool2", Description: "Second tool", Parameters: map[string]any{}},
		}

		result := ToAnthropicFormat(tools)

		require.Len(t, result, 2)
		assert.Equal(t, "tool1", result[0]["name"])
		assert.Equal(t, "tool2", result[1]["name"])
	})
}

// =============================================================================
// ToOpenAIFormat Tests
// =============================================================================

func TestToOpenAIFormat(t *testing.T) {
	t.Run("empty tools list", func(t *testing.T) {
		result := ToOpenAIFormat([]ToolDefinition{})
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("nil tools list", func(t *testing.T) {
		result := ToOpenAIFormat(nil)
		assert.NotNil(t, result)
		assert.Len(t, result, 0)
	})

	t.Run("single tool has function type", func(t *testing.T) {
		tools := []ToolDefinition{
			{
				Name:        "query_table",
				Description: "Query a database table",
				Parameters: map[string]any{
					"type": "object",
				},
			},
		}

		result := ToOpenAIFormat(tools)

		require.Len(t, result, 1)
		assert.Equal(t, "function", result[0]["type"])

		function := result[0]["function"].(map[string]any)
		assert.Equal(t, "query_table", function["name"])
		assert.Equal(t, "Query a database table", function["description"])
		assert.NotNil(t, function["parameters"])
	})

	t.Run("multiple tools all have function type", func(t *testing.T) {
		tools := []ToolDefinition{
			{Name: "tool1", Description: "First tool", Parameters: map[string]any{}},
			{Name: "tool2", Description: "Second tool", Parameters: map[string]any{}},
		}

		result := ToOpenAIFormat(tools)

		require.Len(t, result, 2)
		for _, item := range result {
			assert.Equal(t, "function", item["type"])
		}
	})
}

// =============================================================================
// ParseToolCall Tests
// =============================================================================

func TestParseToolCall(t *testing.T) {
	t.Run("empty args", func(t *testing.T) {
		name, args, err := ParseToolCall("query_table", "")

		require.NoError(t, err)
		assert.Equal(t, "query_table", name)
		assert.Nil(t, args)
	})

	t.Run("valid JSON args", func(t *testing.T) {
		name, args, err := ParseToolCall("query_table", `{"table": "users", "limit": 10}`)

		require.NoError(t, err)
		assert.Equal(t, "query_table", name)
		require.NotNil(t, args)
		assert.Equal(t, "users", args["table"])
		assert.Equal(t, float64(10), args["limit"])
	})

	t.Run("complex nested args", func(t *testing.T) {
		name, args, err := ParseToolCall("insert_record", `{
			"table": "users",
			"values": {
				"name": "John",
				"email": "john@example.com"
			}
		}`)

		require.NoError(t, err)
		assert.Equal(t, "insert_record", name)
		require.NotNil(t, args)
		assert.Equal(t, "users", args["table"])

		values := args["values"].(map[string]any)
		assert.Equal(t, "John", values["name"])
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, args, err := ParseToolCall("query_table", `{invalid json}`)

		require.Error(t, err)
		assert.Nil(t, args)
		assert.Contains(t, err.Error(), "failed to parse tool arguments")
	})

	t.Run("array args", func(t *testing.T) {
		name, args, err := ParseToolCall("query_table", `{"columns": ["id", "name", "email"]}`)

		require.NoError(t, err)
		assert.Equal(t, "query_table", name)
		require.NotNil(t, args)

		columns := args["columns"].([]any)
		assert.Len(t, columns, 3)
	})
}

// =============================================================================
// ToolDefinition Struct Tests
// =============================================================================

func TestToolDefinition_Struct(t *testing.T) {
	t.Run("zero value", func(t *testing.T) {
		var def ToolDefinition
		assert.Equal(t, "", def.Name)
		assert.Equal(t, "", def.Description)
		assert.Nil(t, def.Parameters)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		def := ToolDefinition{
			Name:        "query_table",
			Description: "Query a database table",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"table": map[string]any{
						"type": "string",
					},
				},
			},
		}

		data, err := json.Marshal(def)
		require.NoError(t, err)

		var parsed ToolDefinition
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, def.Name, parsed.Name)
		assert.Equal(t, def.Description, parsed.Description)
		assert.NotNil(t, parsed.Parameters)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonStr := `{
			"name": "test_tool",
			"description": "A test tool",
			"parameters": {"type": "object"}
		}`

		var def ToolDefinition
		err := json.Unmarshal([]byte(jsonStr), &def)
		require.NoError(t, err)

		assert.Equal(t, "test_tool", def.Name)
		assert.Equal(t, "A test tool", def.Description)
		assert.NotNil(t, def.Parameters)
	})
}

// =============================================================================
// ExecuteToolResult Struct Tests
// =============================================================================

func TestExecuteToolResult_Struct(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &ExecuteToolResult{
			Content: "Query executed successfully",
			IsError: false,
		}
		assert.Equal(t, "Query executed successfully", result.Content)
		assert.False(t, result.IsError)
	})

	t.Run("error result", func(t *testing.T) {
		result := &ExecuteToolResult{
			Content: "Table not found",
			IsError: true,
		}
		assert.Equal(t, "Table not found", result.Content)
		assert.True(t, result.IsError)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		result := &ExecuteToolResult{
			Content: "Test content",
			IsError: false,
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var parsed ExecuteToolResult
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, result.Content, parsed.Content)
		assert.Equal(t, result.IsError, parsed.IsError)
	})
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewMCPToolExecutor(t *testing.T) {
	t.Run("creates with nil registry", func(t *testing.T) {
		executor := NewMCPToolExecutor(nil)
		require.NotNil(t, executor)
		assert.Nil(t, executor.toolRegistry)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkIsDataTool_Match(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isDataTool("query_table")
	}
}

func BenchmarkIsDataTool_NoMatch(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isDataTool("execute_sql")
	}
}

func BenchmarkToAnthropicFormat(b *testing.B) {
	tools := []ToolDefinition{
		{Name: "tool1", Description: "Desc 1", Parameters: map[string]any{"type": "object"}},
		{Name: "tool2", Description: "Desc 2", Parameters: map[string]any{"type": "object"}},
		{Name: "tool3", Description: "Desc 3", Parameters: map[string]any{"type": "object"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToAnthropicFormat(tools)
	}
}

func BenchmarkToOpenAIFormat(b *testing.B) {
	tools := []ToolDefinition{
		{Name: "tool1", Description: "Desc 1", Parameters: map[string]any{"type": "object"}},
		{Name: "tool2", Description: "Desc 2", Parameters: map[string]any{"type": "object"}},
		{Name: "tool3", Description: "Desc 3", Parameters: map[string]any{"type": "object"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToOpenAIFormat(tools)
	}
}

func BenchmarkParseToolCall_Simple(b *testing.B) {
	argsJSON := `{"table": "users", "limit": 10}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseToolCall("query_table", argsJSON)
	}
}

func BenchmarkParseToolCall_Complex(b *testing.B) {
	argsJSON := `{"table": "users", "values": {"name": "John", "email": "john@example.com", "age": 30}}`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseToolCall("insert_record", argsJSON)
	}
}
