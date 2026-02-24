package tools

import (
	"context"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
)

func TestThinkTool_Name(t *testing.T) {
	tool := NewThinkTool()
	assert.Equal(t, "think", tool.Name())
}

func TestThinkTool_Description(t *testing.T) {
	tool := NewThinkTool()
	assert.NotEmpty(t, tool.Description())
	assert.Contains(t, tool.Description(), "plan")
}

func TestThinkTool_RequiredScopes(t *testing.T) {
	tool := NewThinkTool()
	scopes := tool.RequiredScopes()
	assert.Empty(t, scopes, "think tool should not require any scopes")
}

func TestThinkTool_InputSchema(t *testing.T) {
	tool := NewThinkTool()
	schema := tool.InputSchema()

	assert.NotNil(t, schema)
	assert.Equal(t, "object", schema["type"])

	properties, ok := schema["properties"].(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, properties, "analysis")
	assert.Contains(t, properties, "plan")
	assert.Contains(t, properties, "tool_choice")

	required, ok := schema["required"].([]string)
	assert.True(t, ok)
	assert.Contains(t, required, "analysis")
	assert.Contains(t, required, "plan")
	assert.Contains(t, required, "tool_choice")
}

func TestThinkTool_Execute(t *testing.T) {
	tool := NewThinkTool()
	ctx := context.Background()
	authCtx := &mcp.AuthContext{}

	t.Run("valid input", func(t *testing.T) {
		args := map[string]any{
			"analysis":    "User wants to find restaurants visited last week",
			"plan":        []any{"1. Query user_visits table", "2. Filter by date range"},
			"tool_choice": "query_table",
			"reasoning":   "This is a structured data query with time filter",
		}

		result, err := tool.Execute(ctx, args, authCtx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
		assert.Len(t, result.Content, 1)

		assert.Equal(t, mcp.ContentTypeText, result.Content[0].Type)
		assert.Contains(t, result.Content[0].Text, "plan_confirmed")
	})

	t.Run("hybrid tool choice", func(t *testing.T) {
		args := map[string]any{
			"analysis":    "User wants Italian restaurants visited recently",
			"plan":        []any{"1. Search KB for Italian cuisine", "2. Query visits with date filter"},
			"tool_choice": "hybrid",
		}

		result, err := tool.Execute(ctx, args, authCtx)
		assert.NoError(t, err)
		assert.False(t, result.IsError)

		assert.Equal(t, mcp.ContentTypeText, result.Content[0].Type)
		assert.Contains(t, result.Content[0].Text, "both")
	})

	t.Run("missing optional fields", func(t *testing.T) {
		args := map[string]any{
			"analysis":    "Simple query",
			"plan":        []any{"Query the table"},
			"tool_choice": "query_table",
		}

		result, err := tool.Execute(ctx, args, authCtx)
		assert.NoError(t, err)
		assert.False(t, result.IsError)
	})
}
