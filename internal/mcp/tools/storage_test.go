package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// =============================================================================
// isTextContentType Tests
// =============================================================================

func TestIsTextContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		// Text types
		{"text/plain", "text/plain", true},
		{"text/html", "text/html", true},
		{"text/css", "text/css", true},
		{"text/csv", "text/csv", true},
		{"text/javascript", "text/javascript", true},

		// Application JSON types
		{"application/json", "application/json", true},
		{"application/json with charset", "application/json; charset=utf-8", true},

		// Application XML types
		{"application/xml", "application/xml", true},
		{"application/xml with charset", "application/xml; charset=utf-8", true},

		// JavaScript/TypeScript
		{"application/javascript", "application/javascript", true},
		{"application/typescript", "application/typescript", true},

		// YAML types
		{"application/x-yaml", "application/x-yaml", true},
		{"application/yaml", "application/yaml", true},

		// Case insensitivity
		{"TEXT/PLAIN uppercase", "TEXT/PLAIN", true},
		{"Application/Json mixed case", "Application/Json", true},

		// Binary types - should NOT be text
		{"image/png", "image/png", false},
		{"image/jpeg", "image/jpeg", false},
		{"image/gif", "image/gif", false},
		{"application/pdf", "application/pdf", false},
		{"application/octet-stream", "application/octet-stream", false},
		{"application/zip", "application/zip", false},
		{"video/mp4", "video/mp4", false},
		{"audio/mpeg", "audio/mpeg", false},

		// Edge cases
		{"empty string", "", false},
		{"unknown type", "application/unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTextContentType(tt.contentType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// ListObjectsTool Tests
// =============================================================================

func TestListObjectsTool_Name(t *testing.T) {
	tool := &ListObjectsTool{}
	assert.Equal(t, "list_objects", tool.Name())
}

func TestListObjectsTool_Description(t *testing.T) {
	tool := &ListObjectsTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "List")
	assert.Contains(t, desc, "bucket")
}

func TestListObjectsTool_RequiredScopes(t *testing.T) {
	tool := &ListObjectsTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadStorage, scopes[0])
}

func TestListObjectsTool_InputSchema(t *testing.T) {
	tool := &ListObjectsTool{}
	schema := tool.InputSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check bucket property
	bucketProp, ok := props["bucket"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", bucketProp["type"])

	// Check prefix property
	prefixProp, ok := props["prefix"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", prefixProp["type"])

	// Check limit property
	limitProp, ok := props["limit"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "integer", limitProp["type"])
	assert.Equal(t, 100, limitProp["default"])
	assert.Equal(t, 1000, limitProp["maximum"])

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "bucket")
}

// =============================================================================
// DownloadObjectTool Tests
// =============================================================================

func TestDownloadObjectTool_Name(t *testing.T) {
	tool := &DownloadObjectTool{}
	assert.Equal(t, "download_object", tool.Name())
}

func TestDownloadObjectTool_Description(t *testing.T) {
	tool := &DownloadObjectTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Download")
	assert.Contains(t, desc, "base64")
}

func TestDownloadObjectTool_RequiredScopes(t *testing.T) {
	tool := &DownloadObjectTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadStorage, scopes[0])
}

func TestDownloadObjectTool_InputSchema(t *testing.T) {
	tool := &DownloadObjectTool{}
	schema := tool.InputSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check bucket property
	_, ok = props["bucket"].(map[string]any)
	require.True(t, ok)

	// Check key property
	_, ok = props["key"].(map[string]any)
	require.True(t, ok)

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "bucket")
	assert.Contains(t, required, "key")
}

// =============================================================================
// UploadObjectTool Tests
// =============================================================================

func TestUploadObjectTool_Name(t *testing.T) {
	tool := &UploadObjectTool{}
	assert.Equal(t, "upload_object", tool.Name())
}

func TestUploadObjectTool_Description(t *testing.T) {
	tool := &UploadObjectTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Upload")
}

func TestUploadObjectTool_RequiredScopes(t *testing.T) {
	tool := &UploadObjectTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeWriteStorage, scopes[0])
}

func TestUploadObjectTool_InputSchema(t *testing.T) {
	tool := &UploadObjectTool{}
	schema := tool.InputSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check content property
	contentProp, ok := props["content"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", contentProp["type"])

	// Check encoding property
	encodingProp, ok := props["encoding"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", encodingProp["type"])
	assert.Equal(t, "text", encodingProp["default"])
	enumVals, ok := encodingProp["enum"].([]string)
	require.True(t, ok)
	assert.Contains(t, enumVals, "text")
	assert.Contains(t, enumVals, "base64")

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "bucket")
	assert.Contains(t, required, "key")
	assert.Contains(t, required, "content")
}

// =============================================================================
// DeleteObjectTool Tests
// =============================================================================

func TestDeleteObjectTool_Name(t *testing.T) {
	tool := &DeleteObjectTool{}
	assert.Equal(t, "delete_object", tool.Name())
}

func TestDeleteObjectTool_Description(t *testing.T) {
	tool := &DeleteObjectTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Delete")
}

func TestDeleteObjectTool_RequiredScopes(t *testing.T) {
	tool := &DeleteObjectTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeWriteStorage, scopes[0])
}

func TestDeleteObjectTool_InputSchema(t *testing.T) {
	tool := &DeleteObjectTool{}
	schema := tool.InputSchema()

	assert.Equal(t, "object", schema["type"])

	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check bucket property
	_, ok = props["bucket"].(map[string]any)
	require.True(t, ok)

	// Check key property
	_, ok = props["key"].(map[string]any)
	require.True(t, ok)

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "bucket")
	assert.Contains(t, required, "key")
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkIsTextContentType_Text(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isTextContentType("application/json; charset=utf-8")
	}
}

func BenchmarkIsTextContentType_Binary(b *testing.B) {
	for i := 0; i < b.N; i++ {
		isTextContentType("application/octet-stream")
	}
}
