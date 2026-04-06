package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// Constants Tests
// =============================================================================

func TestMCPVersion(t *testing.T) {
	t.Run("version is set", func(t *testing.T) {
		assert.NotEmpty(t, MCPVersion)
		assert.Equal(t, "2024-11-05", MCPVersion)
	})
}

func TestFluxbaseVersion(t *testing.T) {
	t.Run("default value is unknown", func(t *testing.T) {
		// FluxbaseVersion is set at build time, so in tests it's "unknown"
		assert.NotEmpty(t, FluxbaseVersion)
	})
}

// =============================================================================
// Server Struct Tests
// =============================================================================

func TestServer_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled: true,
		}

		server := &Server{
			config:    cfg,
			transport: nil,
			tools:     nil,
			resources: nil,
		}

		assert.Equal(t, cfg, server.config)
		assert.Nil(t, server.transport)
		assert.Nil(t, server.tools)
		assert.Nil(t, server.resources)
	})
}

// =============================================================================
// NewServer Tests
// =============================================================================

func TestNewServer(t *testing.T) {
	t.Run("creates server with config", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled:      true,
			AllowedTools: []string{"tool1", "tool2"},
		}

		server := NewServer(cfg)

		require.NotNil(t, server)
		assert.Equal(t, cfg, server.config)
		assert.NotNil(t, server.transport)
		assert.NotNil(t, server.tools)
		assert.NotNil(t, server.resources)
	})

	t.Run("creates server with nil config", func(t *testing.T) {
		server := NewServer(nil)

		require.NotNil(t, server)
		assert.Nil(t, server.config)
	})
}

// =============================================================================
// ToolRegistry Tests
// =============================================================================

func TestServer_ToolRegistry(t *testing.T) {
	t.Run("returns tool registry", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		registry := server.ToolRegistry()

		require.NotNil(t, registry)
		assert.Equal(t, server.tools, registry)
	})
}

// =============================================================================
// ResourceRegistry Tests
// =============================================================================

func TestServer_ResourceRegistry(t *testing.T) {
	t.Run("returns resource registry", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		registry := server.ResourceRegistry()

		require.NotNil(t, registry)
		assert.Equal(t, server.resources, registry)
	})
}

// =============================================================================
// HandleRequest Tests
// =============================================================================

func TestServer_HandleRequest(t *testing.T) {
	t.Run("returns parse error for invalid JSON", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		response := server.HandleRequest(context.Background(), []byte("invalid json"), authCtx)

		require.NotNil(t, response)
		assert.NotNil(t, response.Error)
		assert.Equal(t, ErrorCodeParseError, response.Error.Code)
	})

	t.Run("returns method not found for unknown method", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{"jsonrpc":"2.0","id":1,"method":"unknown/method"}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.NotNil(t, response.Error)
		assert.Equal(t, ErrorCodeMethodNotFound, response.Error.Code)
	})

	t.Run("handles ping request", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{"jsonrpc":"2.0","id":"ping-1","method":"ping"}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("handles initialize request", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "initialize",
			"params": {
				"protocolVersion": "2024-11-05",
				"clientInfo": {
					"name": "TestClient",
					"version": "1.0.0"
				}
			}
		}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)

		// Check result contains expected fields
		result, err := json.Marshal(response.Result)
		require.NoError(t, err)
		assert.Contains(t, string(result), "protocolVersion")
		assert.Contains(t, string(result), "serverInfo")
	})

	t.Run("handles tools/list request", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("handles resources/list request", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{"jsonrpc":"2.0","id":1,"method":"resources/list"}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})

	t.Run("handles resources/templates request", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		reqData := `{"jsonrpc":"2.0","id":1,"method":"resources/templates"}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.Nil(t, response.Error)
		assert.NotNil(t, response.Result)
	})
}

// =============================================================================
// SerializeResponse Tests
// =============================================================================

func TestServer_SerializeResponse(t *testing.T) {
	t.Run("serializes successful response", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		response := &Response{
			JSONRPC: "2.0",
			ID:      "test-id",
			Result:  map[string]string{"status": "ok"},
		}

		data, err := server.SerializeResponse(response)

		require.NoError(t, err)
		assert.Contains(t, string(data), "2.0")
		assert.Contains(t, string(data), "test-id")
		assert.Contains(t, string(data), "status")
	})

	t.Run("serializes error response", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		response := &Response{
			JSONRPC: "2.0",
			ID:      "error-id",
			Error: &Error{
				Code:    -32600,
				Message: "Invalid Request",
			},
		}

		data, err := server.SerializeResponse(response)

		require.NoError(t, err)
		assert.Contains(t, string(data), "error")
		assert.Contains(t, string(data), "-32600")
		assert.Contains(t, string(data), "Invalid Request")
	})
}

// =============================================================================
// Method Constant Tests
// =============================================================================

func TestMethodConstants(t *testing.T) {
	t.Run("initialize method", func(t *testing.T) {
		assert.Equal(t, "initialize", MethodInitialize)
	})

	t.Run("ping method", func(t *testing.T) {
		assert.Equal(t, "ping", MethodPing)
	})

	t.Run("tools/list method", func(t *testing.T) {
		assert.Equal(t, "tools/list", MethodToolsList)
	})

	t.Run("tools/call method", func(t *testing.T) {
		assert.Equal(t, "tools/call", MethodToolsCall)
	})

	t.Run("resources/list method", func(t *testing.T) {
		assert.Equal(t, "resources/list", MethodResourcesList)
	})

	t.Run("resources/read method", func(t *testing.T) {
		assert.Equal(t, "resources/read", MethodResourcesRead)
	})

	t.Run("resources/templates method", func(t *testing.T) {
		assert.Equal(t, "resources/templates", MethodResourcesTemplates)
	})
}

// =============================================================================
// AllowedTools Filtering Tests
// =============================================================================

func TestServer_AllowedToolsFiltering(t *testing.T) {
	t.Run("filters tools by allowed list", func(t *testing.T) {
		cfg := &config.MCPConfig{
			AllowedTools: []string{"allowed_tool"},
		}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		// Request tools/call for a non-allowed tool
		reqData := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "tools/call",
			"params": {
				"name": "blocked_tool",
				"arguments": {}
			}
		}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.NotNil(t, response.Error)
		// Should return tool not found
		assert.Contains(t, response.Error.Message, "blocked_tool")
	})
}

// =============================================================================
// AllowedResources Filtering Tests
// =============================================================================

func TestServer_AllowedResourcesFiltering(t *testing.T) {
	t.Run("filters resources by allowed list", func(t *testing.T) {
		cfg := &config.MCPConfig{
			AllowedResources: []string{"allowed://resource"},
		}
		server := NewServer(cfg)
		authCtx := &AuthContext{}

		// Request resources/read for a non-allowed resource
		reqData := `{
			"jsonrpc": "2.0",
			"id": 1,
			"method": "resources/read",
			"params": {
				"uri": "blocked://resource"
			}
		}`
		response := server.HandleRequest(context.Background(), []byte(reqData), authCtx)

		require.NotNil(t, response)
		assert.NotNil(t, response.Error)
		// Should return resource not found
		assert.Contains(t, response.Error.Message, "blocked://resource")
	})
}

// =============================================================================
// dispatch Method Tests
// =============================================================================

func TestServer_dispatch(t *testing.T) {
	t.Run("routes initialize method", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		req := &Request{
			JSONRPC: "2.0",
			ID:      1,
			Method:  MethodInitialize,
			Params: json.RawMessage(`{
				"protocolVersion": "2024-11-05",
				"clientInfo": {"name": "Test", "version": "1.0"}
			}`),
		}

		response := server.dispatch(context.Background(), req, &AuthContext{})

		assert.Nil(t, response.Error)
	})

	t.Run("routes ping method", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		req := &Request{
			JSONRPC: "2.0",
			ID:      1,
			Method:  MethodPing,
		}

		response := server.dispatch(context.Background(), req, &AuthContext{})

		assert.Nil(t, response.Error)
	})

	t.Run("returns method not found for unknown", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		server := NewServer(cfg)

		req := &Request{
			JSONRPC: "2.0",
			ID:      1,
			Method:  "unknown/method",
		}

		response := server.dispatch(context.Background(), req, &AuthContext{})

		require.NotNil(t, response.Error)
		assert.Equal(t, ErrorCodeMethodNotFound, response.Error.Code)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkServer_HandleRequest_Ping(b *testing.B) {
	cfg := &config.MCPConfig{}
	server := NewServer(cfg)
	authCtx := &AuthContext{}
	reqData := []byte(`{"jsonrpc":"2.0","id":1,"method":"ping"}`)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.HandleRequest(ctx, reqData, authCtx)
	}
}

func BenchmarkServer_HandleRequest_Initialize(b *testing.B) {
	cfg := &config.MCPConfig{}
	server := NewServer(cfg)
	authCtx := &AuthContext{}
	reqData := []byte(`{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "initialize",
		"params": {
			"protocolVersion": "2024-11-05",
			"clientInfo": {"name": "Test", "version": "1.0"}
		}
	}`)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.HandleRequest(ctx, reqData, authCtx)
	}
}

func BenchmarkServer_SerializeResponse(b *testing.B) {
	cfg := &config.MCPConfig{}
	server := NewServer(cfg)
	response := &Response{
		JSONRPC: "2.0",
		ID:      "test-id",
		Result:  map[string]string{"status": "ok"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = server.SerializeResponse(response)
	}
}
