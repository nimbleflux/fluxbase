package tools

import (
	"context"
	"sync"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Tool Handler
// =============================================================================

type mockToolHandler struct {
	name           string
	description    string
	inputSchema    map[string]any
	requiredScopes []string
	executeResult  *mcp.ToolResult
	executeErr     error
}

func (m *mockToolHandler) Name() string                { return m.name }
func (m *mockToolHandler) Description() string         { return m.description }
func (m *mockToolHandler) InputSchema() map[string]any { return m.inputSchema }
func (m *mockToolHandler) RequiredScopes() []string    { return m.requiredScopes }
func (m *mockToolHandler) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	return m.executeResult, m.executeErr
}

// =============================================================================
// NewRegistry Tests
// =============================================================================

func TestNewRegistry(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		registry := NewRegistry()

		require.NotNil(t, registry)
		assert.Empty(t, registry.tools)
	})

	t.Run("tools map is initialized", func(t *testing.T) {
		registry := NewRegistry()

		assert.NotNil(t, registry.tools)
		assert.Equal(t, 0, len(registry.tools))
	})
}

// =============================================================================
// Register Tests
// =============================================================================

func TestRegistry_Register(t *testing.T) {
	t.Run("registers single tool", func(t *testing.T) {
		registry := NewRegistry()
		tool := &mockToolHandler{name: "test_tool"}

		registry.Register(tool)

		assert.Len(t, registry.tools, 1)
		assert.Equal(t, tool, registry.tools["test_tool"])
	})

	t.Run("registers multiple tools", func(t *testing.T) {
		registry := NewRegistry()
		tool1 := &mockToolHandler{name: "tool1"}
		tool2 := &mockToolHandler{name: "tool2"}
		tool3 := &mockToolHandler{name: "tool3"}

		registry.Register(tool1)
		registry.Register(tool2)
		registry.Register(tool3)

		assert.Len(t, registry.tools, 3)
	})

	t.Run("overwrites tool with same name", func(t *testing.T) {
		registry := NewRegistry()
		tool1 := &mockToolHandler{name: "duplicate", description: "first"}
		tool2 := &mockToolHandler{name: "duplicate", description: "second"}

		registry.Register(tool1)
		registry.Register(tool2)

		assert.Len(t, registry.tools, 1)
		assert.Equal(t, "second", registry.tools["duplicate"].Description())
	})
}

// =============================================================================
// Get Tests
// =============================================================================

func TestRegistry_Get(t *testing.T) {
	t.Run("returns nil for empty registry", func(t *testing.T) {
		registry := NewRegistry()

		tool := registry.Get("nonexistent")

		assert.Nil(t, tool)
	})

	t.Run("returns registered tool", func(t *testing.T) {
		registry := NewRegistry()
		expected := &mockToolHandler{name: "my_tool", description: "My tool"}
		registry.Register(expected)

		tool := registry.Get("my_tool")

		assert.Equal(t, expected, tool)
	})

	t.Run("returns nil for non-existent tool", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{name: "exists"})

		tool := registry.Get("does_not_exist")

		assert.Nil(t, tool)
	})

	t.Run("finds correct tool among multiple", func(t *testing.T) {
		registry := NewRegistry()
		tool1 := &mockToolHandler{name: "tool1", description: "First"}
		tool2 := &mockToolHandler{name: "tool2", description: "Second"}
		tool3 := &mockToolHandler{name: "tool3", description: "Third"}
		registry.Register(tool1)
		registry.Register(tool2)
		registry.Register(tool3)

		found := registry.Get("tool2")

		require.NotNil(t, found)
		assert.Equal(t, "Second", found.Description())
	})
}

// =============================================================================
// List Tests
// =============================================================================

func TestRegistry_List(t *testing.T) {
	t.Run("returns empty list for empty registry", func(t *testing.T) {
		registry := NewRegistry()
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		tools := registry.List(authCtx)

		assert.Empty(t, tools)
	})

	t.Run("returns all tools when user has required scopes", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{
			name:           "tool1",
			description:    "Tool 1",
			inputSchema:    map[string]any{"type": "object"},
			requiredScopes: []string{"read"},
		})
		registry.Register(&mockToolHandler{
			name:           "tool2",
			description:    "Tool 2",
			inputSchema:    map[string]any{"type": "object"},
			requiredScopes: []string{"read"},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		tools := registry.List(authCtx)

		assert.Len(t, tools, 2)
	})

	t.Run("filters tools by required scopes", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{
			name:           "public_tool",
			requiredScopes: []string{"read"},
		})
		registry.Register(&mockToolHandler{
			name:           "admin_tool",
			requiredScopes: []string{"admin"},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		tools := registry.List(authCtx)

		assert.Len(t, tools, 1)
		assert.Equal(t, "public_tool", tools[0].Name)
	})

	t.Run("returns tool with all properties", func(t *testing.T) {
		registry := NewRegistry()
		inputSchema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}
		registry.Register(&mockToolHandler{
			name:           "detailed_tool",
			description:    "A detailed tool",
			inputSchema:    inputSchema,
			requiredScopes: []string{},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		tools := registry.List(authCtx)

		require.Len(t, tools, 1)
		assert.Equal(t, "detailed_tool", tools[0].Name)
		assert.Equal(t, "A detailed tool", tools[0].Description)
		assert.NotNil(t, tools[0].InputSchema)
	})

	t.Run("tool with no required scopes is accessible", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{
			name:           "open_tool",
			requiredScopes: []string{},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		tools := registry.List(authCtx)

		assert.Len(t, tools, 1)
	})

	t.Run("user with multiple scopes can access restricted tools", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{
			name:           "multi_scope_tool",
			requiredScopes: []string{"read", "write"},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{"read", "write", "admin"}}

		tools := registry.List(authCtx)

		assert.Len(t, tools, 1)
	})

	t.Run("user missing one scope cannot access tool", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{
			name:           "multi_scope_tool",
			requiredScopes: []string{"read", "write"},
		})
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		tools := registry.List(authCtx)

		assert.Empty(t, tools)
	})
}

// =============================================================================
// Names Tests
// =============================================================================

func TestRegistry_Names(t *testing.T) {
	t.Run("returns empty list for empty registry", func(t *testing.T) {
		registry := NewRegistry()

		names := registry.Names()

		assert.Empty(t, names)
	})

	t.Run("returns all tool names", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{name: "alpha"})
		registry.Register(&mockToolHandler{name: "beta"})
		registry.Register(&mockToolHandler{name: "gamma"})

		names := registry.Names()

		assert.Len(t, names, 3)
		assert.Contains(t, names, "alpha")
		assert.Contains(t, names, "beta")
		assert.Contains(t, names, "gamma")
	})

	t.Run("returns names without duplicates after overwrite", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{name: "tool"})
		registry.Register(&mockToolHandler{name: "tool"}) // Overwrite

		names := registry.Names()

		assert.Len(t, names, 1)
		assert.Contains(t, names, "tool")
	})
}

// =============================================================================
// Registry Struct Tests
// =============================================================================

func TestRegistry_Struct(t *testing.T) {
	t.Run("has tools map and mutex", func(t *testing.T) {
		registry := &Registry{
			tools: make(map[string]ToolHandler),
		}

		assert.NotNil(t, registry.tools)
	})
}

// =============================================================================
// ToolHandler Interface Tests
// =============================================================================

func TestToolHandler_Interface(t *testing.T) {
	t.Run("mock implements interface", func(t *testing.T) {
		var _ ToolHandler = &mockToolHandler{}
	})

	t.Run("returns expected values", func(t *testing.T) {
		handler := &mockToolHandler{
			name:           "test",
			description:    "Test tool",
			inputSchema:    map[string]any{"type": "object"},
			requiredScopes: []string{"read"},
		}

		assert.Equal(t, "test", handler.Name())
		assert.Equal(t, "Test tool", handler.Description())
		assert.NotNil(t, handler.InputSchema())
		assert.Equal(t, []string{"read"}, handler.RequiredScopes())
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent registrations", func(t *testing.T) {
		registry := NewRegistry()
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				tool := &mockToolHandler{name: "tool" + string(rune('a'+n%26))}
				registry.Register(tool)
			}(i)
		}

		wg.Wait()
		// Should have at most 26 unique tools (a-z)
		assert.LessOrEqual(t, len(registry.tools), 26)
	})

	t.Run("handles concurrent reads", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockToolHandler{name: "tool", requiredScopes: []string{}})
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = registry.Get("tool")
				_ = registry.List(authCtx)
				_ = registry.Names()
			}()
		}

		wg.Wait()
	})

	t.Run("handles concurrent reads and writes", func(t *testing.T) {
		registry := NewRegistry()
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		var wg sync.WaitGroup

		// Writers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func(n int) {
				defer wg.Done()
				tool := &mockToolHandler{name: "tool" + string(rune('0'+n%10)), requiredScopes: []string{}}
				registry.Register(tool)
			}(i)
		}

		// Readers
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = registry.Get("tool0")
				_ = registry.List(authCtx)
				_ = registry.Names()
			}()
		}

		wg.Wait()
	})
}

// =============================================================================
// Tool Struct Tests
// =============================================================================

func TestTool_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		tool := mcp.Tool{
			Name:        "my_tool",
			Description: "My awesome tool",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		}

		assert.Equal(t, "my_tool", tool.Name)
		assert.Equal(t, "My awesome tool", tool.Description)
		assert.NotNil(t, tool.InputSchema)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRegistry_Register(b *testing.B) {
	registry := NewRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool := &mockToolHandler{name: "tool"}
		registry.Register(tool)
	}
}

func BenchmarkRegistry_Get(b *testing.B) {
	registry := NewRegistry()
	registry.Register(&mockToolHandler{name: "target"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Get("target")
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	registry := NewRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(&mockToolHandler{
			name:           "tool" + string(rune(i)),
			requiredScopes: []string{},
		})
	}
	authCtx := &mcp.AuthContext{Scopes: []string{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.List(authCtx)
	}
}

func BenchmarkRegistry_Names(b *testing.B) {
	registry := NewRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(&mockToolHandler{name: "tool" + string(rune(i))})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.Names()
	}
}
