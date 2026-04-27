package routes

import (
	"github.com/gofiber/fiber/v3"
)

// ExtensionsAdminDeps contains dependencies for extensions routes.
// These routes allow both instance admins and tenant admins to manage
// extensions for their respective context.
//
// Role Access:
//   - instance_admin: Full access, manages extensions on main or any tenant DB
//   - tenant_admin: Manage extensions for their own tenant database
type ExtensionsAdminDeps struct {
	ListExtensions   fiber.Handler
	GetExtension     fiber.Handler
	EnableExtension  fiber.Handler
	DisableExtension fiber.Handler
	SyncExtensions   fiber.Handler
}

// BuildExtensionsAdminRoutes creates the extensions route group.
func BuildExtensionsAdminRoutes(deps *ExtensionsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "extensions",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			{Method: "GET", Path: "/extensions", Handler: deps.ListExtensions, Summary: "List extensions"},
			{Method: "GET", Path: "/extensions/:name", Handler: deps.GetExtension, Summary: "Get extension status"},
			{Method: "POST", Path: "/extensions/:name/enable", Handler: deps.EnableExtension, Summary: "Enable extension"},
			{Method: "POST", Path: "/extensions/:name/disable", Handler: deps.DisableExtension, Summary: "Disable extension"},
			{Method: "POST", Path: "/extensions/sync", Handler: deps.SyncExtensions, Summary: "Sync extensions"},
		},
	}
}

// ExtensionsTenantDeps is an alias kept for backward compatibility with wiring code.
// The tenant routes are now merged into ExtensionsAdminDeps.
type ExtensionsTenantDeps = ExtensionsAdminDeps

// BuildExtensionsTenantRoutes is kept for backward compatibility.
// It returns nil so the merged route group (from BuildExtensionsAdminRoutes) is used instead.
func BuildExtensionsTenantRoutes(_ *ExtensionsTenantDeps) *RouteGroup {
	return nil
}
