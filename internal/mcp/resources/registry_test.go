package resources

import (
	"context"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Resource Providers
// =============================================================================

// mockProvider is a simple resource provider for testing
type mockProvider struct {
	uri            string
	name           string
	description    string
	mimeType       string
	requiredScopes []string
	content        []mcp.Content
	readErr        error
}

func (m *mockProvider) URI() string              { return m.uri }
func (m *mockProvider) Name() string             { return m.name }
func (m *mockProvider) Description() string      { return m.description }
func (m *mockProvider) MimeType() string         { return m.mimeType }
func (m *mockProvider) RequiredScopes() []string { return m.requiredScopes }
func (m *mockProvider) Read(ctx context.Context, authCtx *mcp.AuthContext) ([]mcp.Content, error) {
	return m.content, m.readErr
}

// mockTemplateProvider is a template resource provider for testing
type mockTemplateProvider struct {
	mockProvider
	isTemplate     bool
	matchParams    map[string]string
	matchResult    bool
	readWithParams []mcp.Content
	readParamsErr  error
}

func (m *mockTemplateProvider) IsTemplate() bool {
	return m.isTemplate
}

func (m *mockTemplateProvider) MatchURI(uri string) (map[string]string, bool) {
	// Use the MatchTemplate helper for actual template matching
	if m.matchParams != nil || m.matchResult {
		return m.matchParams, m.matchResult
	}
	return MatchTemplate(m.uri, uri)
}

func (m *mockTemplateProvider) ReadWithParams(ctx context.Context, authCtx *mcp.AuthContext, params map[string]string) ([]mcp.Content, error) {
	return m.readWithParams, m.readParamsErr
}

// =============================================================================
// NewRegistry Tests
// =============================================================================

func TestNewRegistry(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		registry := NewRegistry()

		require.NotNil(t, registry)
		assert.Empty(t, registry.providers)
	})

	t.Run("providers slice is initialized", func(t *testing.T) {
		registry := NewRegistry()

		assert.NotNil(t, registry.providers)
		assert.Equal(t, 0, len(registry.providers))
	})
}

// =============================================================================
// Register Tests
// =============================================================================

func TestRegistry_Register(t *testing.T) {
	t.Run("registers single provider", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{uri: "test://resource"}

		registry.Register(provider)

		assert.Len(t, registry.providers, 1)
		assert.Equal(t, provider, registry.providers[0])
	})

	t.Run("registers multiple providers", func(t *testing.T) {
		registry := NewRegistry()
		provider1 := &mockProvider{uri: "test://resource1"}
		provider2 := &mockProvider{uri: "test://resource2"}
		provider3 := &mockProvider{uri: "test://resource3"}

		registry.Register(provider1)
		registry.Register(provider2)
		registry.Register(provider3)

		assert.Len(t, registry.providers, 3)
	})

	t.Run("preserves registration order", func(t *testing.T) {
		registry := NewRegistry()
		provider1 := &mockProvider{uri: "test://first"}
		provider2 := &mockProvider{uri: "test://second"}

		registry.Register(provider1)
		registry.Register(provider2)

		assert.Equal(t, "test://first", registry.providers[0].URI())
		assert.Equal(t, "test://second", registry.providers[1].URI())
	})
}

// =============================================================================
// GetProvider Tests
// =============================================================================

func TestRegistry_GetProvider(t *testing.T) {
	t.Run("returns nil for empty registry", func(t *testing.T) {
		registry := NewRegistry()

		provider := registry.GetProvider("test://resource")

		assert.Nil(t, provider)
	})

	t.Run("finds exact match", func(t *testing.T) {
		registry := NewRegistry()
		expected := &mockProvider{uri: "test://resource"}
		registry.Register(expected)

		provider := registry.GetProvider("test://resource")

		assert.Equal(t, expected, provider)
	})

	t.Run("returns nil for no match", func(t *testing.T) {
		registry := NewRegistry()
		registry.Register(&mockProvider{uri: "test://other"})

		provider := registry.GetProvider("test://resource")

		assert.Nil(t, provider)
	})

	t.Run("finds template match", func(t *testing.T) {
		registry := NewRegistry()
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{uri: "test://tables/{table}"},
			isTemplate:   true,
		}
		registry.Register(templateProvider)

		provider := registry.GetProvider("test://tables/users")

		assert.Equal(t, templateProvider, provider)
	})

	t.Run("prefers exact match over template", func(t *testing.T) {
		registry := NewRegistry()
		exactProvider := &mockProvider{uri: "test://tables/users"}
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{uri: "test://tables/{table}"},
			isTemplate:   true,
		}
		registry.Register(exactProvider)
		registry.Register(templateProvider)

		provider := registry.GetProvider("test://tables/users")

		assert.Equal(t, exactProvider, provider)
	})

	t.Run("template provider with IsTemplate false is not matched as template", func(t *testing.T) {
		registry := NewRegistry()
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{uri: "test://tables/{table}"},
			isTemplate:   false, // Not a template
		}
		registry.Register(templateProvider)

		provider := registry.GetProvider("test://tables/users")

		assert.Nil(t, provider)
	})
}

// =============================================================================
// ReadResource Tests
// =============================================================================

func TestRegistry_ReadResource(t *testing.T) {
	t.Run("returns nil for no matching provider", func(t *testing.T) {
		registry := NewRegistry()
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		content, err := registry.ReadResource(context.Background(), "test://missing", authCtx)

		assert.NoError(t, err)
		assert.Nil(t, content)
	})

	t.Run("reads from exact match provider", func(t *testing.T) {
		registry := NewRegistry()
		expectedContent := []mcp.Content{{Type: "text", Text: "hello"}}
		provider := &mockProvider{
			uri:            "test://resource",
			requiredScopes: []string{"read"},
			content:        expectedContent,
		}
		registry.Register(provider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		content, err := registry.ReadResource(context.Background(), "test://resource", authCtx)

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("skips provider when missing required scope", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{
			uri:            "test://resource",
			requiredScopes: []string{"admin"},
			content:        []mcp.Content{{Type: "text", Text: "secret"}},
		}
		registry.Register(provider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		content, err := registry.ReadResource(context.Background(), "test://resource", authCtx)

		assert.NoError(t, err)
		assert.Nil(t, content)
	})

	t.Run("reads from template provider with params", func(t *testing.T) {
		registry := NewRegistry()
		expectedContent := []mcp.Content{{Type: "text", Text: "table data"}}
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				requiredScopes: []string{"read"},
			},
			isTemplate:     true,
			readWithParams: expectedContent,
		}
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		content, err := registry.ReadResource(context.Background(), "test://tables/users", authCtx)

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})

	t.Run("skips template provider when missing scope", func(t *testing.T) {
		registry := NewRegistry()
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				requiredScopes: []string{"admin"},
			},
			isTemplate:     true,
			readWithParams: []mcp.Content{{Type: "text", Text: "secret"}},
		}
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		content, err := registry.ReadResource(context.Background(), "test://tables/users", authCtx)

		assert.NoError(t, err)
		assert.Nil(t, content)
	})

	t.Run("provider with no required scopes is accessible", func(t *testing.T) {
		registry := NewRegistry()
		expectedContent := []mcp.Content{{Type: "text", Text: "public"}}
		provider := &mockProvider{
			uri:            "test://public",
			requiredScopes: []string{}, // No scopes required
			content:        expectedContent,
		}
		registry.Register(provider)
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		content, err := registry.ReadResource(context.Background(), "test://public", authCtx)

		assert.NoError(t, err)
		assert.Equal(t, expectedContent, content)
	})
}

// =============================================================================
// ListResources Tests
// =============================================================================

func TestRegistry_ListResources(t *testing.T) {
	t.Run("returns empty list for empty registry", func(t *testing.T) {
		registry := NewRegistry()
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		resources := registry.ListResources(authCtx)

		assert.Empty(t, resources)
	})

	t.Run("lists static resources", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{
			uri:            "test://resource",
			name:           "Test Resource",
			description:    "A test resource",
			mimeType:       "application/json",
			requiredScopes: []string{"read"},
		}
		registry.Register(provider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		resources := registry.ListResources(authCtx)

		require.Len(t, resources, 1)
		assert.Equal(t, "test://resource", resources[0].URI)
		assert.Equal(t, "Test Resource", resources[0].Name)
		assert.Equal(t, "A test resource", resources[0].Description)
		assert.Equal(t, "application/json", resources[0].MimeType)
	})

	t.Run("excludes templates from list", func(t *testing.T) {
		registry := NewRegistry()
		staticProvider := &mockProvider{
			uri:            "test://static",
			requiredScopes: []string{},
		}
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				requiredScopes: []string{},
			},
			isTemplate: true,
		}
		registry.Register(staticProvider)
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		resources := registry.ListResources(authCtx)

		require.Len(t, resources, 1)
		assert.Equal(t, "test://static", resources[0].URI)
	})

	t.Run("filters by required scopes", func(t *testing.T) {
		registry := NewRegistry()
		publicProvider := &mockProvider{
			uri:            "test://public",
			requiredScopes: []string{},
		}
		privateProvider := &mockProvider{
			uri:            "test://private",
			requiredScopes: []string{"admin"},
		}
		registry.Register(publicProvider)
		registry.Register(privateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		resources := registry.ListResources(authCtx)

		require.Len(t, resources, 1)
		assert.Equal(t, "test://public", resources[0].URI)
	})

	t.Run("returns all accessible resources with matching scopes", func(t *testing.T) {
		registry := NewRegistry()
		provider1 := &mockProvider{
			uri:            "test://resource1",
			requiredScopes: []string{"read"},
		}
		provider2 := &mockProvider{
			uri:            "test://resource2",
			requiredScopes: []string{"write"},
		}
		registry.Register(provider1)
		registry.Register(provider2)
		authCtx := &mcp.AuthContext{Scopes: []string{"read", "write"}}

		resources := registry.ListResources(authCtx)

		assert.Len(t, resources, 2)
	})
}

// =============================================================================
// ListTemplates Tests
// =============================================================================

func TestRegistry_ListTemplates(t *testing.T) {
	t.Run("returns empty list for empty registry", func(t *testing.T) {
		registry := NewRegistry()
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		templates := registry.ListTemplates(authCtx)

		assert.Empty(t, templates)
	})

	t.Run("lists template resources", func(t *testing.T) {
		registry := NewRegistry()
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				name:           "Table Resource",
				description:    "Access table data",
				mimeType:       "application/json",
				requiredScopes: []string{"read"},
			},
			isTemplate: true,
		}
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		templates := registry.ListTemplates(authCtx)

		require.Len(t, templates, 1)
		assert.Equal(t, "test://tables/{table}", templates[0].URITemplate)
		assert.Equal(t, "Table Resource", templates[0].Name)
		assert.Equal(t, "Access table data", templates[0].Description)
		assert.Equal(t, "application/json", templates[0].MimeType)
	})

	t.Run("excludes static resources", func(t *testing.T) {
		registry := NewRegistry()
		staticProvider := &mockProvider{
			uri:            "test://static",
			requiredScopes: []string{},
		}
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				requiredScopes: []string{},
			},
			isTemplate: true,
		}
		registry.Register(staticProvider)
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		templates := registry.ListTemplates(authCtx)

		require.Len(t, templates, 1)
		assert.Equal(t, "test://tables/{table}", templates[0].URITemplate)
	})

	t.Run("filters by required scopes", func(t *testing.T) {
		registry := NewRegistry()
		publicTemplate := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://public/{id}",
				requiredScopes: []string{},
			},
			isTemplate: true,
		}
		privateTemplate := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://private/{id}",
				requiredScopes: []string{"admin"},
			},
			isTemplate: true,
		}
		registry.Register(publicTemplate)
		registry.Register(privateTemplate)
		authCtx := &mcp.AuthContext{Scopes: []string{"read"}}

		templates := registry.ListTemplates(authCtx)

		require.Len(t, templates, 1)
		assert.Equal(t, "test://public/{id}", templates[0].URITemplate)
	})

	t.Run("excludes template provider with IsTemplate false", func(t *testing.T) {
		registry := NewRegistry()
		templateProvider := &mockTemplateProvider{
			mockProvider: mockProvider{
				uri:            "test://tables/{table}",
				requiredScopes: []string{},
			},
			isTemplate: false, // Not a template
		}
		registry.Register(templateProvider)
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		templates := registry.ListTemplates(authCtx)

		assert.Empty(t, templates)
	})
}

// =============================================================================
// MatchTemplate Tests
// =============================================================================

func TestMatchTemplate(t *testing.T) {
	t.Run("matches simple template with one parameter", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}", "test://tables/users")

		assert.True(t, matched)
		assert.Equal(t, "users", params["table"])
	})

	t.Run("matches template with multiple parameters", func(t *testing.T) {
		params, matched := MatchTemplate("test://{schema}/{table}", "test://public/users")

		assert.True(t, matched)
		assert.Equal(t, "public", params["schema"])
		assert.Equal(t, "users", params["table"])
	})

	t.Run("matches template with literal and parameter parts", func(t *testing.T) {
		params, matched := MatchTemplate("test://schema/tables/{table}/columns/{column}", "test://schema/tables/users/columns/id")

		assert.True(t, matched)
		assert.Equal(t, "users", params["table"])
		assert.Equal(t, "id", params["column"])
	})

	t.Run("returns false for mismatched literal parts", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}", "test://schemas/users")

		assert.False(t, matched)
		assert.Nil(t, params)
	})

	t.Run("returns false for different part count", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}", "test://tables/users/columns")

		assert.False(t, matched)
		assert.Nil(t, params)
	})

	t.Run("returns false for shorter URI", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}/columns", "test://tables/users")

		assert.False(t, matched)
		assert.Nil(t, params)
	})

	t.Run("matches exact URI without parameters", func(t *testing.T) {
		params, matched := MatchTemplate("test://static/resource", "test://static/resource")

		assert.True(t, matched)
		assert.Empty(t, params)
	})

	t.Run("returns false for different exact URIs", func(t *testing.T) {
		params, matched := MatchTemplate("test://static/resource1", "test://static/resource2")

		assert.False(t, matched)
		assert.Nil(t, params)
	})

	t.Run("extracts empty parameter value", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}", "test://tables/")

		assert.True(t, matched)
		assert.Equal(t, "", params["table"])
	})

	t.Run("handles parameter with special characters in value", func(t *testing.T) {
		params, matched := MatchTemplate("test://tables/{table}", "test://tables/my-table_v2")

		assert.True(t, matched)
		assert.Equal(t, "my-table_v2", params["table"])
	})

	t.Run("returns empty params map for no parameters", func(t *testing.T) {
		params, matched := MatchTemplate("test://resource", "test://resource")

		assert.True(t, matched)
		require.NotNil(t, params)
		assert.Len(t, params, 0)
	})
}

// =============================================================================
// Resource and ResourceTemplate Struct Tests
// =============================================================================

func TestResource_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		resource := mcp.Resource{
			URI:         "test://resource",
			Name:        "Test Resource",
			Description: "A test resource",
			MimeType:    "application/json",
		}

		assert.Equal(t, "test://resource", resource.URI)
		assert.Equal(t, "Test Resource", resource.Name)
		assert.Equal(t, "A test resource", resource.Description)
		assert.Equal(t, "application/json", resource.MimeType)
	})
}

func TestResourceTemplate_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		template := mcp.ResourceTemplate{
			URITemplate: "test://tables/{table}",
			Name:        "Table Template",
			Description: "Access table data",
			MimeType:    "application/json",
		}

		assert.Equal(t, "test://tables/{table}", template.URITemplate)
		assert.Equal(t, "Table Template", template.Name)
		assert.Equal(t, "Access table data", template.Description)
		assert.Equal(t, "application/json", template.MimeType)
	})
}

// =============================================================================
// Content Struct Tests
// =============================================================================

func TestContent_Struct(t *testing.T) {
	t.Run("stores text content", func(t *testing.T) {
		content := mcp.Content{
			Type: mcp.ContentTypeText,
			Text: "Hello, World!",
		}

		assert.Equal(t, mcp.ContentTypeText, content.Type)
		assert.Equal(t, "Hello, World!", content.Text)
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestRegistry_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent registrations", func(t *testing.T) {
		registry := NewRegistry()
		done := make(chan bool)

		// Spawn multiple goroutines registering providers
		for i := 0; i < 10; i++ {
			go func(id int) {
				provider := &mockProvider{
					uri: "test://resource" + string(rune('0'+id)),
				}
				registry.Register(provider)
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		assert.Len(t, registry.providers, 10)
	})

	t.Run("handles concurrent reads", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{
			uri:            "test://resource",
			requiredScopes: []string{},
			content:        []mcp.Content{{Type: "text", Text: "data"}},
		}
		registry.Register(provider)
		authCtx := &mcp.AuthContext{Scopes: []string{}}

		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_, _ = registry.ReadResource(context.Background(), "test://resource", authCtx)
				_ = registry.GetProvider("test://resource")
				_ = registry.ListResources(authCtx)
				_ = registry.ListTemplates(authCtx)
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkMatchTemplate_Simple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MatchTemplate("test://tables/{table}", "test://tables/users")
	}
}

func BenchmarkMatchTemplate_Multiple(b *testing.B) {
	for i := 0; i < b.N; i++ {
		MatchTemplate("test://{schema}/{table}/{column}", "test://public/users/id")
	}
}

func BenchmarkRegistry_GetProvider(b *testing.B) {
	registry := NewRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(&mockProvider{
			uri: "test://resource" + string(rune(i)),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.GetProvider("test://resource50")
	}
}

func BenchmarkRegistry_ListResources(b *testing.B) {
	registry := NewRegistry()
	for i := 0; i < 100; i++ {
		registry.Register(&mockProvider{
			uri:            "test://resource" + string(rune(i)),
			requiredScopes: []string{},
		})
	}
	authCtx := &mcp.AuthContext{Scopes: []string{}}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry.ListResources(authCtx)
	}
}
