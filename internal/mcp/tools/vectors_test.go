package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// =============================================================================
// SearchVectorsTool Tests
// =============================================================================

func TestSearchVectorsTool_Name(t *testing.T) {
	tool := &SearchVectorsTool{}
	assert.Equal(t, "search_vectors", tool.Name())
}

func TestSearchVectorsTool_Description(t *testing.T) {
	tool := &SearchVectorsTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Search")
	assert.Contains(t, desc, "vector")
	assert.Contains(t, desc, "semantic")
}

func TestSearchVectorsTool_RequiredScopes(t *testing.T) {
	tool := &SearchVectorsTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadVectors, scopes[0])
}

func TestSearchVectorsTool_InputSchema(t *testing.T) {
	tool := &SearchVectorsTool{}
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check query property
	queryProp, ok := props["query"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", queryProp["type"])
	assert.NotEmpty(t, queryProp["description"])

	// Check chatbot_id property
	chatbotIDProp, ok := props["chatbot_id"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", chatbotIDProp["type"])

	// Check knowledge_bases property
	kbProp, ok := props["knowledge_bases"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", kbProp["type"])
	items, ok := kbProp["items"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", items["type"])

	// Check limit property
	limitProp, ok := props["limit"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "integer", limitProp["type"])
	assert.Equal(t, 5, limitProp["default"])
	assert.Equal(t, 20, limitProp["maximum"])

	// Check threshold property
	thresholdProp, ok := props["threshold"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "number", thresholdProp["type"])
	assert.Equal(t, 0.7, thresholdProp["default"])

	// Check tags property
	tagsProp, ok := props["tags"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "array", tagsProp["type"])

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "query")
	assert.Len(t, required, 1) // Only query is required, chatbot_id is optional
}

func TestSearchVectorsTool_InputSchema_PropertyDescriptions(t *testing.T) {
	tool := &SearchVectorsTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	testCases := []struct {
		name     string
		propName string
		contains string
	}{
		{"query description", "query", "search"},
		{"chatbot_id description", "chatbot_id", "chatbot"},
		{"knowledge_bases description", "knowledge_bases", "knowledge base"},
		{"limit description", "limit", "Maximum"},
		{"threshold description", "threshold", "similarity"},
		{"tags description", "tags", "tags"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prop := props[tc.propName].(map[string]any)
			desc, ok := prop["description"].(string)
			require.True(t, ok, "property %s should have description", tc.propName)
			assert.Contains(t, desc, tc.contains, "property %s description should contain %q", tc.propName, tc.contains)
		})
	}
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewSearchVectorsTool(t *testing.T) {
	// Test with nil RAG service (acceptable for metadata-only tests)
	tool := NewSearchVectorsTool(nil)
	require.NotNil(t, tool)
	assert.Nil(t, tool.ragService)
}

// =============================================================================
// Default Values Tests
// =============================================================================

func TestSearchVectorsTool_DefaultValues(t *testing.T) {
	tool := &SearchVectorsTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	t.Run("limit default is 5", func(t *testing.T) {
		limitProp := props["limit"].(map[string]any)
		assert.Equal(t, 5, limitProp["default"])
	})

	t.Run("threshold default is 0.7", func(t *testing.T) {
		thresholdProp := props["threshold"].(map[string]any)
		assert.Equal(t, 0.7, thresholdProp["default"])
	})

	t.Run("limit maximum is 20", func(t *testing.T) {
		limitProp := props["limit"].(map[string]any)
		assert.Equal(t, 20, limitProp["maximum"])
	})
}
