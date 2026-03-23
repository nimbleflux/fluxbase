package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// =============================================================================
// InvokeFunctionTool Tests
// =============================================================================

func TestInvokeFunctionTool_Name(t *testing.T) {
	tool := &InvokeFunctionTool{}
	assert.Equal(t, "invoke_function", tool.Name())
}

func TestInvokeFunctionTool_Description(t *testing.T) {
	tool := &InvokeFunctionTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Invoke")
	assert.Contains(t, desc, "edge function")
}

func TestInvokeFunctionTool_RequiredScopes(t *testing.T) {
	tool := &InvokeFunctionTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteFunctions, scopes[0])
}

func TestInvokeFunctionTool_InputSchema(t *testing.T) {
	tool := &InvokeFunctionTool{}
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check name property
	nameProp, ok := props["name"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", nameProp["type"])
	assert.NotEmpty(t, nameProp["description"])

	// Check namespace property
	namespaceProp, ok := props["namespace"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", namespaceProp["type"])

	// Check body property
	bodyProp, ok := props["body"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", bodyProp["type"])

	// Check method property
	methodProp, ok := props["method"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", methodProp["type"])
	assert.Equal(t, "POST", methodProp["default"])

	// Check enum values for method
	enumVals, ok := methodProp["enum"].([]string)
	require.True(t, ok)
	assert.Contains(t, enumVals, "GET")
	assert.Contains(t, enumVals, "POST")
	assert.Contains(t, enumVals, "PUT")
	assert.Contains(t, enumVals, "PATCH")
	assert.Contains(t, enumVals, "DELETE")

	// Check headers property
	headersProp, ok := props["headers"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", headersProp["type"])

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "name")
	assert.Len(t, required, 1)
}

func TestInvokeFunctionTool_InputSchema_PropertyDescriptions(t *testing.T) {
	tool := &InvokeFunctionTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	testCases := []struct {
		name     string
		propName string
		contains string
	}{
		{"name description", "name", "edge function"},
		{"namespace description", "namespace", "namespace"},
		{"body description", "body", "Request body"},
		{"method description", "method", "HTTP method"},
		{"headers description", "headers", "headers"},
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
// InvokeRPCTool Tests
// =============================================================================

func TestInvokeRPCTool_Name(t *testing.T) {
	tool := &InvokeRPCTool{}
	assert.Equal(t, "invoke_rpc", tool.Name())
}

func TestInvokeRPCTool_Description(t *testing.T) {
	tool := &InvokeRPCTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Invoke")
	assert.Contains(t, desc, "RPC")
	assert.Contains(t, desc, "procedure")
}

func TestInvokeRPCTool_RequiredScopes(t *testing.T) {
	tool := &InvokeRPCTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteRPC, scopes[0])
}

func TestInvokeRPCTool_InputSchema(t *testing.T) {
	tool := &InvokeRPCTool{}
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check name property
	nameProp, ok := props["name"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", nameProp["type"])

	// Check namespace property
	namespaceProp, ok := props["namespace"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", namespaceProp["type"])

	// Check params property
	paramsProp, ok := props["params"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", paramsProp["type"])

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "name")
	assert.Len(t, required, 1)
}

func TestInvokeRPCTool_InputSchema_PropertyDescriptions(t *testing.T) {
	tool := &InvokeRPCTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	testCases := []struct {
		name     string
		propName string
		contains string
	}{
		{"name description", "name", "RPC procedure"},
		{"namespace description", "namespace", "namespace"},
		{"params description", "params", "Parameters"},
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

func TestNewInvokeFunctionTool(t *testing.T) {
	// Test with nil dependencies (acceptable for metadata-only tests)
	tool := NewInvokeFunctionTool(nil, nil, "http://localhost:8080", "/functions")
	require.NotNil(t, tool)
	assert.Equal(t, "http://localhost:8080", tool.publicURL)
	assert.Equal(t, "/functions", tool.functionsDir)
}

func TestNewInvokeRPCTool(t *testing.T) {
	// Test with nil dependencies (acceptable for metadata-only tests)
	tool := NewInvokeRPCTool(nil, nil)
	require.NotNil(t, tool)
	assert.Nil(t, tool.executor)
	assert.Nil(t, tool.storage)
}

// =============================================================================
// HTTP Method Enum Tests
// =============================================================================

func TestInvokeFunctionTool_HTTPMethodEnum(t *testing.T) {
	tool := &InvokeFunctionTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)
	methodProp := props["method"].(map[string]any)
	enumVals := methodProp["enum"].([]string)

	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	for _, method := range expectedMethods {
		t.Run(method+" is supported", func(t *testing.T) {
			assert.Contains(t, enumVals, method)
		})
	}

	t.Run("exactly 5 methods supported", func(t *testing.T) {
		assert.Len(t, enumVals, 5)
	})
}
