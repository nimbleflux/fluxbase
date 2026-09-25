package mcp

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

// =============================================================================
// Service key scope enforcement (finding: raw service keys short-circuited
// HasScope, making key scopes decorative)
// =============================================================================

func TestExtractAuthContext_ServiceKeyScopes(t *testing.T) {
	newCtx := func(locals map[string]any) *AuthContext {
		app := fiber.New()
		c := app.AcquireCtx(&fasthttp.RequestCtx{})
		defer app.ReleaseCtx(c)
		for k, v := range locals {
			c.Locals(k, v)
		}
		return ExtractAuthContext(c)
	}

	t.Run("service key with empty scopes keeps service role (allow-all)", func(t *testing.T) {
		authCtx := newCtx(map[string]any{
			"auth_type":          "service_key",
			"service_key_scopes": []string{},
		})
		assert.True(t, authCtx.IsServiceRole)
		assert.True(t, authCtx.HasScope("tables:read"))
	})

	t.Run("service key with nil scopes keeps service role (allow-all)", func(t *testing.T) {
		authCtx := newCtx(map[string]any{
			"auth_type": "service_key",
		})
		assert.True(t, authCtx.IsServiceRole)
	})

	t.Run("service key with explicit scopes has them enforced", func(t *testing.T) {
		authCtx := newCtx(map[string]any{
			"auth_type":          "service_key",
			"service_key_scopes": []string{"tables:read"},
		})
		assert.False(t, authCtx.IsServiceRole)
		assert.True(t, authCtx.HasScope("tables:read"))
		assert.False(t, authCtx.HasScope("admin:ddl"))
	})

	t.Run("service_role role still short-circuits", func(t *testing.T) {
		authCtx := newCtx(map[string]any{
			"auth_type": "service_key",
			"user_role": "service_role",
		})
		assert.True(t, authCtx.IsServiceRole)
	})
}

func TestExtractAuthContext_ClientKeyRoleMirror(t *testing.T) {
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)

	c.Locals("auth_type", "clientkey")
	c.Locals("client_key_id", "key-123")
	c.Locals("user_id", "user-1")
	c.Locals("client_key_scopes", []string{"tables:read"})

	authCtx := ExtractAuthContext(c)
	assert.Equal(t, "authenticated", authCtx.UserRole, "user-bound client keys act as authenticated role for RLS")

	// Key without a bound user stays anonymous
	app2 := fiber.New()
	c2 := app2.AcquireCtx(&fasthttp.RequestCtx{})
	defer app2.ReleaseCtx(c2)
	c2.Locals("auth_type", "clientkey")
	c2.Locals("client_key_id", "key-123")

	authCtx2 := ExtractAuthContext(c2)
	assert.Empty(t, authCtx2.UserRole)
}

func TestExtractAuthContext_TenantID(t *testing.T) {
	app := fiber.New()
	c := app.AcquireCtx(&fasthttp.RequestCtx{})
	defer app.ReleaseCtx(c)
	c.Locals("tenant_id", "550e8400-e29b-41d4-a716-446655440000")

	authCtx := ExtractAuthContext(c)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", authCtx.TenantID)
}

// =============================================================================
// Tenant-scoped registry filtering
// =============================================================================

type tenantScopedTool struct {
	mockToolHandler
	tenantID string
}

func (t *tenantScopedTool) OwnerTenantID() string { return t.tenantID }

type tenantScopedResource struct {
	mockResourceProvider
	tenantID string
}

func (r *tenantScopedResource) OwnerTenantID() string { return r.tenantID }

func TestRegistry_TenantFiltering(t *testing.T) {
	newRegistry := func() (*ToolRegistry, *ResourceRegistry) {
		tools := NewToolRegistry()
		tools.Register(&mockToolHandler{name: "global_tool", requiredScopes: []string{}})
		tools.Register(&tenantScopedTool{
			mockToolHandler: mockToolHandler{name: "tenant_tool", requiredScopes: []string{}},
			tenantID:        "550e8400-e29b-41d4-a716-446655440000",
		})

		resources := NewResourceRegistry()
		resources.Register(&mockResourceProvider{uri: "fluxbase://global", requiredScopes: []string{}})
		resources.Register(&tenantScopedResource{
			mockResourceProvider: mockResourceProvider{uri: "fluxbase://tenant", requiredScopes: []string{}},
			tenantID:             "550e8400-e29b-41d4-a716-446655440000",
		})
		return tools, resources
	}

	tenant := "550e8400-e29b-41d4-a716-446655440000"
	other := "11111111-2222-3333-4444-555555555555"

	t.Run("caller sees global tools and own tenant tools", func(t *testing.T) {
		tools, resources := newRegistry()
		authCtx := &AuthContext{TenantID: tenant}

		names := map[string]bool{}
		for _, tool := range tools.ListTools(authCtx) {
			names[tool.Name] = true
		}
		assert.True(t, names["global_tool"])
		assert.True(t, names["tenant_tool"])

		uris := map[string]bool{}
		for _, res := range resources.ListResources(authCtx) {
			uris[res.URI] = true
		}
		assert.True(t, uris["fluxbase://global"])
		assert.True(t, uris["fluxbase://tenant"])
	})

	t.Run("caller does not see other tenants' tools or resources", func(t *testing.T) {
		tools, resources := newRegistry()
		authCtx := &AuthContext{TenantID: other}

		for _, tool := range tools.ListTools(authCtx) {
			assert.NotEqual(t, "tenant_tool", tool.Name)
		}
		for _, res := range resources.ListResources(authCtx) {
			assert.NotEqual(t, "fluxbase://tenant", res.URI)
		}
	})

	t.Run("instance admins see everything", func(t *testing.T) {
		tools, resources := newRegistry()
		authCtx := &AuthContext{TenantID: other, UserRole: "instance_admin"}

		names := map[string]bool{}
		for _, tool := range tools.ListTools(authCtx) {
			names[tool.Name] = true
		}
		assert.True(t, names["tenant_tool"])
		assert.True(t, names["global_tool"])

		assert.NotEmpty(t, resources.ListResources(authCtx))
	})

	t.Run("tenantAllowsAccess guards execution time lookups", func(t *testing.T) {
		authCtx := &AuthContext{TenantID: other}
		scoped := &tenantScopedTool{tenantID: tenant}
		assert.False(t, tenantAllowsAccess(authCtx, scoped))
		assert.True(t, tenantAllowsAccess(&AuthContext{TenantID: tenant}, scoped))
		assert.True(t, tenantAllowsAccess(&AuthContext{TenantID: other, UserRole: "admin"}, scoped))
		// Non-tenant-scoped handlers are always accessible
		assert.True(t, tenantAllowsAccess(authCtx, &mockToolHandler{}))
	})
}
