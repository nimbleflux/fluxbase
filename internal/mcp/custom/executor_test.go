package custom

import (
	"strings"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/nimbleflux/fluxbase/internal/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewExecutor Tests
// =============================================================================

func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with parameters", func(t *testing.T) {
		executor := NewExecutor("jwt-secret", "https://example.com", nil)

		require.NotNil(t, executor)
		assert.NotNil(t, executor.runtime)
		assert.Equal(t, "https://example.com", executor.publicURL)
		assert.Equal(t, "jwt-secret", executor.jwtSecret)
		assert.Nil(t, executor.secretsService)
	})

	t.Run("creates executor with empty parameters", func(t *testing.T) {
		executor := NewExecutor("", "", nil)

		require.NotNil(t, executor)
		assert.NotNil(t, executor.runtime)
		assert.Equal(t, "", executor.publicURL)
		assert.Equal(t, "", executor.jwtSecret)
	})
}

// =============================================================================
// Executor Struct Tests
// =============================================================================

func TestExecutor_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		executor := &Executor{
			publicURL: "https://example.com",
			jwtSecret: "secret123",
		}

		assert.Equal(t, "https://example.com", executor.publicURL)
		assert.Equal(t, "secret123", executor.jwtSecret)
	})
}

// =============================================================================
// ValidateToolCode Tests
// =============================================================================

func TestValidateToolCode(t *testing.T) {
	t.Run("accepts code with handler function", func(t *testing.T) {
		code := `
function handler(args, fluxbase, fluxbaseService, utils) {
	return { message: "Hello" };
}
`
		err := ValidateToolCode(code)
		assert.NoError(t, err)
	})

	t.Run("accepts code with export default", func(t *testing.T) {
		code := `
export default async function(args, fluxbase) {
	return "result";
}
`
		err := ValidateToolCode(code)
		assert.NoError(t, err)
	})

	t.Run("accepts code with export async function", func(t *testing.T) {
		code := `
export async function handler(args) {
	return args.name;
}
`
		err := ValidateToolCode(code)
		assert.NoError(t, err)
	})

	t.Run("rejects empty code", func(t *testing.T) {
		err := ValidateToolCode("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects whitespace only code", func(t *testing.T) {
		err := ValidateToolCode("   \n\t  ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects code without handler export", func(t *testing.T) {
		code := `
function processData(data) {
	return data * 2;
}
const result = processData(42);
`
		err := ValidateToolCode(code)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must export")
	})

	t.Run("accepts code with async handler", func(t *testing.T) {
		code := `
async function handler(args) {
	const result = await fetch('https://api.example.com');
	return result.json();
}
`
		err := ValidateToolCode(code)
		assert.NoError(t, err)
	})

	t.Run("accepts code with handler as arrow function", func(t *testing.T) {
		code := `const handler = async (args) => args.name;`
		err := ValidateToolCode(code)
		assert.NoError(t, err)
	})
}

// =============================================================================
// ValidateResourceCode Tests
// =============================================================================

func TestValidateResourceCode(t *testing.T) {
	t.Run("accepts code with handler function", func(t *testing.T) {
		code := `
function handler(params, fluxbase) {
	return { data: params.id };
}
`
		err := ValidateResourceCode(code)
		assert.NoError(t, err)
	})

	t.Run("rejects empty code", func(t *testing.T) {
		err := ValidateResourceCode("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects code without handler", func(t *testing.T) {
		code := `const x = 42;`
		err := ValidateResourceCode(code)
		assert.Error(t, err)
	})
}

// =============================================================================
// parseToolResult Tests
// =============================================================================

func TestExecutor_parseToolResult(t *testing.T) {
	executor := &Executor{}

	t.Run("parses successful result with content", func(t *testing.T) {
		body := `{"content":[{"type":"text","text":"Hello World"}],"isError":false}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		require.NotNil(t, toolResult)
		assert.False(t, toolResult.IsError)
		require.Len(t, toolResult.Content, 1)
		assert.Equal(t, mcp.ContentTypeText, toolResult.Content[0].Type)
		assert.Equal(t, "Hello World", toolResult.Content[0].Text)
	})

	t.Run("parses error result", func(t *testing.T) {
		body := `{"content":[{"type":"text","text":"Something went wrong"}],"isError":true}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		require.NotNil(t, toolResult)
		assert.True(t, toolResult.IsError)
		require.Len(t, toolResult.Content, 1)
		assert.Contains(t, toolResult.Content[0].Text, "Something went wrong")
	})

	t.Run("handles execution failure", func(t *testing.T) {
		result := &runtime.ExecutionResult{
			Success: false,
			Error:   "Runtime error: undefined variable",
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		require.NotNil(t, toolResult)
		assert.True(t, toolResult.IsError)
		require.Len(t, toolResult.Content, 1)
		assert.Contains(t, toolResult.Content[0].Text, "Runtime error")
	})

	t.Run("handles invalid JSON body", func(t *testing.T) {
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    "plain text result",
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		require.NotNil(t, toolResult)
		assert.False(t, toolResult.IsError)
		require.Len(t, toolResult.Content, 1)
		assert.Equal(t, "plain text result", toolResult.Content[0].Text)
	})

	t.Run("handles multiple content items", func(t *testing.T) {
		body := `{"content":[{"type":"text","text":"Part 1"},{"type":"text","text":"Part 2"}],"isError":false}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		require.Len(t, toolResult.Content, 2)
		assert.Equal(t, "Part 1", toolResult.Content[0].Text)
		assert.Equal(t, "Part 2", toolResult.Content[1].Text)
	})

	t.Run("handles empty content array", func(t *testing.T) {
		body := `{"content":[],"isError":false}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		toolResult, err := executor.parseToolResult(result)

		require.NoError(t, err)
		assert.Empty(t, toolResult.Content)
	})
}

// =============================================================================
// parseResourceResult Tests
// =============================================================================

func TestExecutor_parseResourceResult(t *testing.T) {
	executor := &Executor{}

	t.Run("parses successful result with contents", func(t *testing.T) {
		body := `{"contents":[{"type":"text","text":"Resource data"}]}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		contents, err := executor.parseResourceResult(result)

		require.NoError(t, err)
		require.Len(t, contents, 1)
		assert.Equal(t, mcp.ContentTypeText, contents[0].Type)
		assert.Equal(t, "Resource data", contents[0].Text)
	})

	t.Run("handles execution failure", func(t *testing.T) {
		result := &runtime.ExecutionResult{
			Success: false,
			Error:   "Failed to fetch resource",
		}

		contents, err := executor.parseResourceResult(result)

		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.Contains(t, err.Error(), "Failed to fetch resource")
	})

	t.Run("handles error in response", func(t *testing.T) {
		body := `{"error":"Resource not found"}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		contents, err := executor.parseResourceResult(result)

		assert.Error(t, err)
		assert.Nil(t, contents)
		assert.Contains(t, err.Error(), "Resource not found")
	})

	t.Run("handles invalid JSON body", func(t *testing.T) {
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    "plain text content",
		}

		contents, err := executor.parseResourceResult(result)

		require.NoError(t, err)
		require.Len(t, contents, 1)
		assert.Equal(t, "plain text content", contents[0].Text)
	})

	t.Run("handles multiple content items", func(t *testing.T) {
		body := `{"contents":[{"type":"text","text":"Line 1"},{"type":"text","text":"Line 2"}]}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		contents, err := executor.parseResourceResult(result)

		require.NoError(t, err)
		require.Len(t, contents, 2)
		assert.Equal(t, "Line 1", contents[0].Text)
		assert.Equal(t, "Line 2", contents[1].Text)
	})

	t.Run("handles empty contents array", func(t *testing.T) {
		body := `{"contents":[]}`
		result := &runtime.ExecutionResult{
			Success: true,
			Body:    body,
		}

		contents, err := executor.parseResourceResult(result)

		require.NoError(t, err)
		assert.Empty(t, contents)
	})
}

// =============================================================================
// wrapToolCode Tests
// =============================================================================

func TestExecutor_wrapToolCode(t *testing.T) {
	executor := &Executor{}

	t.Run("wraps code with arguments and context", func(t *testing.T) {
		code := `function handler(args) { return args.name; }`
		args := map[string]any{"name": "test", "count": 42}
		ctx := map[string]any{"tool_name": "my_tool", "user_id": "user123"}

		wrapped := executor.wrapToolCode(code, args, ctx)

		// Verify the wrapped code contains key elements
		assert.Contains(t, wrapped, "__MCP_ARGS__")
		assert.Contains(t, wrapped, "__MCP_CONTEXT__")
		assert.Contains(t, wrapped, "FluxbaseClient")
		assert.Contains(t, wrapped, "fluxbase")
		assert.Contains(t, wrapped, "fluxbaseService")
		assert.Contains(t, wrapped, "toolUtils")
		assert.Contains(t, wrapped, code)

		// Verify JSON serialized arguments
		assert.Contains(t, wrapped, `"name":"test"`)
		assert.Contains(t, wrapped, `"count":42`)
		assert.Contains(t, wrapped, `"tool_name":"my_tool"`)
	})

	t.Run("includes SDK code", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapToolCode(code, map[string]any{}, map[string]any{})

		// Verify SDK classes are present
		assert.Contains(t, wrapped, "class QueryBuilder")
		assert.Contains(t, wrapped, "class InsertBuilder")
		assert.Contains(t, wrapped, "class UpdateBuilder")
		assert.Contains(t, wrapped, "class DeleteBuilder")
		assert.Contains(t, wrapped, "class FluxbaseClient")
	})

	t.Run("includes execution wrapper", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapToolCode(code, map[string]any{}, map[string]any{})

		// Verify execution logic
		assert.Contains(t, wrapped, "Execute the handler")
		assert.Contains(t, wrapped, "__RESULT__")
		assert.Contains(t, wrapped, "catch")
	})

	t.Run("handles empty arguments", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapToolCode(code, map[string]any{}, map[string]any{})

		// Should contain empty JSON objects
		assert.Contains(t, wrapped, "{}")
	})

	t.Run("includes AI utilities", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapToolCode(code, map[string]any{}, map[string]any{})

		assert.Contains(t, wrapped, "ai:")
		assert.Contains(t, wrapped, "async chat(options)")
		assert.Contains(t, wrapped, "async embed(options)")
		assert.Contains(t, wrapped, "async listProviders()")
	})

	t.Run("includes secrets accessor", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapToolCode(code, map[string]any{}, map[string]any{})

		assert.Contains(t, wrapped, "secrets:")
		assert.Contains(t, wrapped, "FLUXBASE_SECRET_")
	})
}

// =============================================================================
// wrapResourceCode Tests
// =============================================================================

func TestExecutor_wrapResourceCode(t *testing.T) {
	executor := &Executor{}

	t.Run("wraps code with params and context", func(t *testing.T) {
		code := `function handler(params) { return params.id; }`
		params := map[string]string{"id": "123", "name": "test"}
		ctx := map[string]any{"resource_uri": "test://resource", "user_id": "user456"}

		wrapped := executor.wrapResourceCode(code, params, ctx)

		// Verify the wrapped code contains key elements
		assert.Contains(t, wrapped, "__MCP_PARAMS__")
		assert.Contains(t, wrapped, "__MCP_CONTEXT__")
		assert.Contains(t, wrapped, "FluxbaseClient")
		assert.Contains(t, wrapped, "resourceUtils")
		assert.Contains(t, wrapped, code)

		// Verify JSON serialized params
		assert.Contains(t, wrapped, `"id":"123"`)
		assert.Contains(t, wrapped, `"name":"test"`)
	})

	t.Run("includes SDK code", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapResourceCode(code, map[string]string{}, map[string]any{})

		assert.Contains(t, wrapped, "class FluxbaseClient")
	})

	t.Run("includes resource execution wrapper", func(t *testing.T) {
		code := `function handler() {}`
		wrapped := executor.wrapResourceCode(code, map[string]string{}, map[string]any{})

		assert.Contains(t, wrapped, "Execute the handler")
		assert.Contains(t, wrapped, "__RESULT__")
		assert.Contains(t, wrapped, "contents")
	})
}

// =============================================================================
// fluxbaseSDKCode Tests
// =============================================================================

func TestFluxbaseSDKCode(t *testing.T) {
	t.Run("returns non-empty JavaScript code", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.NotEmpty(t, code)
		assert.True(t, len(code) > 1000, "SDK code should be substantial")
	})

	t.Run("contains QueryBuilder class", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "class QueryBuilder")
		assert.Contains(t, code, "select(columns)")
		assert.Contains(t, code, "eq(column, value)")
		assert.Contains(t, code, "neq(column, value)")
		assert.Contains(t, code, "gt(column, value)")
		assert.Contains(t, code, "gte(column, value)")
		assert.Contains(t, code, "lt(column, value)")
		assert.Contains(t, code, "lte(column, value)")
		assert.Contains(t, code, "like(column, pattern)")
		assert.Contains(t, code, "ilike(column, pattern)")
		assert.Contains(t, code, "order(column, options")
		assert.Contains(t, code, "limit(count)")
		assert.Contains(t, code, "offset(count)")
		assert.Contains(t, code, "single()")
		assert.Contains(t, code, "maybeSingle()")
	})

	t.Run("contains InsertBuilder class", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "class InsertBuilder")
		assert.Contains(t, code, "onConflict(columns)")
	})

	t.Run("contains UpdateBuilder class", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "class UpdateBuilder")
	})

	t.Run("contains DeleteBuilder class", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "class DeleteBuilder")
	})

	t.Run("contains FluxbaseClient class", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "class FluxbaseClient")
		assert.Contains(t, code, "_fetch(path, options")
		assert.Contains(t, code, "from(table)")
		assert.Contains(t, code, "insert(table, data)")
		assert.Contains(t, code, "update(table, data)")
		assert.Contains(t, code, "delete(table)")
		assert.Contains(t, code, "rpc(functionName, params")
	})

	t.Run("contains storage operations", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "storage =")
		assert.Contains(t, code, "async list(bucket, options")
		assert.Contains(t, code, "async download(bucket, path)")
		assert.Contains(t, code, "async upload(bucket, path, file")
		assert.Contains(t, code, "async remove(bucket, paths)")
		assert.Contains(t, code, "getPublicUrl(bucket, path)")
	})

	t.Run("contains functions operations", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, "functions =")
		assert.Contains(t, code, "async invoke(functionName, options")
	})

	t.Run("includes proper authorization headers", func(t *testing.T) {
		code := fluxbaseSDKCode()

		assert.Contains(t, code, `"Authorization": "Bearer "`)
	})
}

// =============================================================================
// CustomTool Struct Tests
// =============================================================================

func TestCustomTool_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		inputSchema := map[string]any{"type": "object"}
		tool := &CustomTool{
			Name:           "my_tool",
			Description:    "A test tool",
			Namespace:      "default",
			Code:           "function handler() {}",
			InputSchema:    inputSchema,
			RequiredScopes: []string{"read", "write"},
			AllowNet:       true,
			AllowEnv:       true,
			AllowRead:      false,
			AllowWrite:     false,
			MemoryLimitMB:  64,
			TimeoutSeconds: 30,
		}

		assert.Equal(t, "my_tool", tool.Name)
		assert.Equal(t, "A test tool", tool.Description)
		assert.Equal(t, "default", tool.Namespace)
		assert.Equal(t, "function handler() {}", tool.Code)
		assert.Equal(t, inputSchema, tool.InputSchema)
		assert.Equal(t, []string{"read", "write"}, tool.RequiredScopes)
		assert.True(t, tool.AllowNet)
		assert.True(t, tool.AllowEnv)
		assert.False(t, tool.AllowRead)
		assert.False(t, tool.AllowWrite)
		assert.Equal(t, 64, tool.MemoryLimitMB)
		assert.Equal(t, 30, tool.TimeoutSeconds)
	})
}

// =============================================================================
// CustomResource Struct Tests
// =============================================================================

func TestCustomResourceExecutor_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		resource := &CustomResource{
			URI:            "test://resource/{id}",
			Name:           "Test Resource",
			Description:    "A test resource",
			Namespace:      "default",
			MimeType:       "application/json",
			Code:           "function handler(params) {}",
			RequiredScopes: []string{"read"},
			IsTemplate:     true,
			TimeoutSeconds: 15,
		}

		assert.Equal(t, "test://resource/{id}", resource.URI)
		assert.Equal(t, "Test Resource", resource.Name)
		assert.Equal(t, "A test resource", resource.Description)
		assert.Equal(t, "default", resource.Namespace)
		assert.Equal(t, "application/json", resource.MimeType)
		assert.Equal(t, "function handler(params) {}", resource.Code)
		assert.Equal(t, []string{"read"}, resource.RequiredScopes)
		assert.True(t, resource.IsTemplate)
		assert.Equal(t, 15, resource.TimeoutSeconds)
	})
}

// =============================================================================
// MCP Content Helper Tests
// =============================================================================

func TestMCPContent_Helpers(t *testing.T) {
	t.Run("TextContent creates text content", func(t *testing.T) {
		content := mcp.TextContent("Hello World")

		assert.Equal(t, mcp.ContentTypeText, content.Type)
		assert.Equal(t, "Hello World", content.Text)
	})

	t.Run("ErrorContent creates error content", func(t *testing.T) {
		content := mcp.ErrorContent("Something failed")

		assert.Equal(t, mcp.ContentTypeText, content.Type)
		assert.Contains(t, content.Text, "Something failed")
	})
}

// =============================================================================
// Query Builder Methods in SDK Tests
// =============================================================================

func TestSDKQueryBuilderMethods(t *testing.T) {
	code := fluxbaseSDKCode()

	t.Run("contains array filter methods", func(t *testing.T) {
		assert.Contains(t, code, "is(column, value)")
		assert.Contains(t, code, "in(column, values)")
		assert.Contains(t, code, "contains(column, value)")
		assert.Contains(t, code, "containedBy(column, value)")
	})

	t.Run("contains proper URL encoding", func(t *testing.T) {
		assert.Contains(t, code, "encodeURIComponent")
	})

	t.Run("contains order options", func(t *testing.T) {
		assert.Contains(t, code, "ascending")
		assert.Contains(t, code, "nullsFirst")
		assert.Contains(t, code, "nullsfirst")
	})
}

// =============================================================================
// JSON Marshaling Tests
// =============================================================================

func TestWrapCodeJSONMarshaling(t *testing.T) {
	executor := &Executor{}

	t.Run("handles special characters in args", func(t *testing.T) {
		code := `function handler() {}`
		args := map[string]any{
			"text":    "Hello \"World\"",
			"newline": "Line1\nLine2",
			"unicode": "日本語",
		}

		wrapped := executor.wrapToolCode(code, args, map[string]any{})

		// Verify special characters are properly escaped
		assert.Contains(t, wrapped, `"text":"Hello \"World\""`)
		assert.True(t, strings.Contains(wrapped, "日本語"))
	})

	t.Run("handles nested objects in args", func(t *testing.T) {
		code := `function handler() {}`
		args := map[string]any{
			"user": map[string]any{
				"name": "John",
				"age":  30,
			},
		}

		wrapped := executor.wrapToolCode(code, args, map[string]any{})

		assert.Contains(t, wrapped, `"user":{`)
		assert.Contains(t, wrapped, `"name":"John"`)
	})

	t.Run("handles arrays in args", func(t *testing.T) {
		code := `function handler() {}`
		args := map[string]any{
			"tags": []string{"a", "b", "c"},
		}

		wrapped := executor.wrapToolCode(code, args, map[string]any{})

		assert.Contains(t, wrapped, `"tags":["a","b","c"]`)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkValidateToolCode_Valid(b *testing.B) {
	code := `
async function handler(args, fluxbase, fluxbaseService, utils) {
	const result = await fluxbase.from('users').select('*').execute();
	return result;
}
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateToolCode(code)
	}
}

func BenchmarkValidateToolCode_Invalid(b *testing.B) {
	code := `const x = 42;`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateToolCode(code)
	}
}

func BenchmarkWrapToolCode(b *testing.B) {
	executor := &Executor{}
	code := `function handler(args) { return args; }`
	args := map[string]any{"name": "test", "count": 42}
	ctx := map[string]any{"tool_name": "benchmark_tool"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.wrapToolCode(code, args, ctx)
	}
}

func BenchmarkFluxbaseSDKCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fluxbaseSDKCode()
	}
}
