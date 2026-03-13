package tools

import (
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SubmitJobTool Tests
// =============================================================================

func TestSubmitJobTool_Name(t *testing.T) {
	tool := &SubmitJobTool{}
	assert.Equal(t, "submit_job", tool.Name())
}

func TestSubmitJobTool_Description(t *testing.T) {
	tool := &SubmitJobTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "Submit")
	assert.Contains(t, desc, "background job")
}

func TestSubmitJobTool_RequiredScopes(t *testing.T) {
	tool := &SubmitJobTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteJobs, scopes[0])
}

func TestSubmitJobTool_InputSchema(t *testing.T) {
	tool := &SubmitJobTool{}
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check job_name property
	jobNameProp, ok := props["job_name"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", jobNameProp["type"])

	// Check namespace property
	namespaceProp, ok := props["namespace"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", namespaceProp["type"])

	// Check payload property
	payloadProp, ok := props["payload"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "object", payloadProp["type"])

	// Check priority property
	priorityProp, ok := props["priority"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "integer", priorityProp["type"])
	assert.Equal(t, 0, priorityProp["default"])

	// Check scheduled_at property
	scheduledAtProp, ok := props["scheduled_at"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", scheduledAtProp["type"])
	assert.Contains(t, scheduledAtProp["description"].(string), "ISO 8601")

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "job_name")
	assert.Len(t, required, 1)
}

func TestSubmitJobTool_InputSchema_PropertyDescriptions(t *testing.T) {
	tool := &SubmitJobTool{}
	schema := tool.InputSchema()
	props := schema["properties"].(map[string]any)

	t.Run("job_name has description", func(t *testing.T) {
		prop := props["job_name"].(map[string]any)
		desc, ok := prop["description"].(string)
		require.True(t, ok)
		assert.NotEmpty(t, desc)
	})

	t.Run("payload has description", func(t *testing.T) {
		prop := props["payload"].(map[string]any)
		desc, ok := prop["description"].(string)
		require.True(t, ok)
		assert.Contains(t, desc, "Data")
	})

	t.Run("priority has description", func(t *testing.T) {
		prop := props["priority"].(map[string]any)
		desc, ok := prop["description"].(string)
		require.True(t, ok)
		assert.Contains(t, desc, "priority")
	})
}

// =============================================================================
// GetJobStatusTool Tests
// =============================================================================

func TestGetJobStatusTool_Name(t *testing.T) {
	tool := &GetJobStatusTool{}
	assert.Equal(t, "get_job_status", tool.Name())
}

func TestGetJobStatusTool_Description(t *testing.T) {
	tool := &GetJobStatusTool{}
	desc := tool.Description()
	assert.Contains(t, desc, "status")
	assert.Contains(t, desc, "job")
}

func TestGetJobStatusTool_RequiredScopes(t *testing.T) {
	tool := &GetJobStatusTool{}
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteJobs, scopes[0])
}

func TestGetJobStatusTool_InputSchema(t *testing.T) {
	tool := &GetJobStatusTool{}
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check job_id property
	jobIDProp, ok := props["job_id"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", jobIDProp["type"])
	assert.Contains(t, jobIDProp["description"].(string), "job ID")

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "job_id")
	assert.Len(t, required, 1)
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewSubmitJobTool(t *testing.T) {
	// Test with nil storage (acceptable for metadata-only tests)
	tool := NewSubmitJobTool(nil)
	require.NotNil(t, tool)
	assert.Nil(t, tool.storage)
}

func TestNewGetJobStatusTool(t *testing.T) {
	// Test with nil storage (acceptable for metadata-only tests)
	tool := NewGetJobStatusTool(nil)
	require.NotNil(t, tool)
	assert.Nil(t, tool.storage)
}
