//go:build integration

package e2e

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

// TestMCPHealthCheck verifies the MCP health endpoint returns 200 without auth.
func TestMCPHealthCheck(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// The MCP health endpoint is public and unauthenticated
	resp := tc.NewRequest("GET", "/mcp/health").
		Send()
	require.Equal(t, fiber.StatusOK, resp.Status(), "MCP health endpoint should return 200")

	// Verify response contains expected fields
	var result map[string]interface{}
	resp.JSON(&result)

	assert.Equal(t, "healthy", result["status"], "MCP server should be healthy")
	assert.Equal(t, "2024-11-05", result["protocolVersion"], "MCP protocol version should match the constant")
	assert.Contains(t, result, "serverVersion", "Server version should be present")
}

// TestMCPInitialize tests the MCP JSON-RPC initialize method.
func TestMCPInitialize(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// Create a test user for authentication
	email := test.E2ETestEmail()
	password := "testpassword123"
	_, token := tc.CreateTestUser(email, password)
	require.NotEmpty(t, token)

	// Send JSON-RPC 2.0 initialize request
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialize",
		"id":      1,
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "e2e-test-client",
				"version": "1.0.0",
			},
		},
	}

	resp := tc.NewRequest("POST", "/mcp").
		WithAuth(token).
		WithBody(body).
		Send().
		AssertStatus(fiber.StatusOK)

	// Parse the JSON-RPC response
	var result map[string]interface{}
	resp.JSON(&result)

	assert.Equal(t, "2.0", result["jsonrpc"])
	assert.Contains(t, result, "result", "Should have a result object")

	resultInfo, ok := result["result"].(map[string]interface{})
	require.True(t, ok, "result should be a map")
	assert.Equal(t, "2024-11-05", resultInfo["protocolVersion"])

	serverInfo, ok := resultInfo["serverInfo"].(map[string]interface{})
	require.True(t, ok, "serverInfo should be a map")
	assert.Equal(t, "Fluxbase MCP Server", serverInfo["name"])
}

// TestMCPListTools tests the MCP tools/list JSON-RPC method.
func TestMCPListTools(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// Create a test user for authentication
	email := test.E2ETestEmail()
	password := "testpassword123"
	_, token := tc.CreateTestUser(email, password)
	require.NotEmpty(t, token)

	// Send JSON-RPC 2.0 tools/list request
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"id":      2,
	}

	resp := tc.NewRequest("POST", "/mcp").
		WithAuth(token).
		WithBody(body).
		Send().
		AssertStatus(fiber.StatusOK)

	// Parse the JSON-RPC response
	var result map[string]interface{}
	resp.JSON(&result)

	assert.Equal(t, "2.0", result["jsonrpc"])
	assert.Contains(t, result, "result", "Should have tools list")

	resultMap, ok := result["result"].(map[string]interface{})
	require.True(t, ok, "result should be a map")
	toolsList, ok := resultMap["tools"].([]interface{})
	require.True(t, ok, "tools should be an array")
	assert.NotNil(t, toolsList)
}

// TestMCPUnauthorized verifies MCP POST endpoint requires authentication.
func TestMCPUnauthorized(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// Send JSON-RPC request without authentication
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "tools/list",
		"id":      1,
	}

	resp := tc.NewRequest("POST", "/mcp").
		WithBody(body).
		Send()

	assert.Equal(t, fiber.StatusUnauthorized, resp.Status())
}
