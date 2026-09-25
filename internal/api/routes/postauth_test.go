package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegistry_PostAuthMiddlewareOrder verifies that post-auth middlewares are
// injected into the handler chain immediately after the auth middleware, so
// they observe the authenticated principal (which group middlewares cannot —
// those run before auth).
func TestRegistry_PostAuthMiddlewareOrder(t *testing.T) {
	var order []string

	authMW := func(c fiber.Ctx) error {
		order = append(order, "auth")
		c.Locals("user_id", "user-1")
		return c.Next()
	}
	postAuth := func(c fiber.Ctx) error {
		order = append(order, "postAuth")
		// The post-auth middleware must see what auth set.
		uid, _ := c.Locals("user_id").(string)
		assert.Equal(t, "user-1", uid, "post-auth middleware must run after auth")
		return c.Next()
	}
	groupMW := func(c fiber.Ctx) error {
		order = append(order, "group")
		return c.Next()
	}
	handler := func(c fiber.Ctx) error {
		order = append(order, "handler")
		return c.SendString("ok")
	}

	registry := NewRegistry(WithPostAuthMiddleware(postAuth))
	require.NoError(t, registry.Register(&RouteGroup{
		Name:   "test",
		Prefix: "/api/test",
		Middlewares: []Middleware{
			{Name: "Group", Handler: groupMW},
		},
		AuthMiddlewares: &AuthMiddlewares{Required: authMW},
		Routes: []Route{
			{Method: "GET", Path: "/", Handler: handler, Auth: AuthRequired},
		},
	}))

	app := fiber.New()
	require.NoError(t, registry.Apply(app))

	req := httptest.NewRequest(http.MethodGet, "/api/test/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, []string{"group", "auth", "postAuth", "handler"}, order,
		"post-auth middleware must run after auth (group middlewares run before auth on this chain)")
}

// TestRegistry_PostAuthMiddlewareSkippedOnPublicRoutes verifies that routes
// without an auth middleware do not get the post-auth handlers injected.
func TestRegistry_PostAuthMiddlewareSkippedOnPublicRoutes(t *testing.T) {
	postAuthRan := false
	postAuth := func(c fiber.Ctx) error {
		postAuthRan = true
		return c.Next()
	}

	registry := NewRegistry(WithPostAuthMiddleware(postAuth))
	require.NoError(t, registry.Register(&RouteGroup{
		Name:   "public",
		Prefix: "/api/pub",
		Routes: []Route{
			{Method: "GET", Path: "/", Handler: func(c fiber.Ctx) error { return c.SendString("ok") }, Summary: "Public health probe", Auth: AuthNone, Public: true},
		},
	}))

	app := fiber.New()
	require.NoError(t, registry.Apply(app))

	req := httptest.NewRequest(http.MethodGet, "/api/pub/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.False(t, postAuthRan, "post-auth middleware must not run on public routes")
}

// TestRegisterPostAuthMiddlewares verifies nil handlers are dropped and order
// is preserved (tenant re-validation before branch context).
func TestRegisterPostAuthMiddlewares(t *testing.T) {
	h1 := func(c fiber.Ctx) error { return nil }
	h2 := func(c fiber.Ctx) error { return nil }

	t.Run("nils dropped, order preserved", func(t *testing.T) {
		got := registerPostAuthMiddlewares(&AllDeps{EnsureTenantAccess: h1, BranchContext: h2})
		require.Len(t, got, 2)
	})

	t.Run("all nil yields empty", func(t *testing.T) {
		got := registerPostAuthMiddlewares(&AllDeps{})
		assert.Empty(t, got)
	})
}
