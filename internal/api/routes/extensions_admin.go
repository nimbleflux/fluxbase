package routes

import (
	"github.com/gofiber/fiber/v3"
)

// ExtensionsAdminDeps contains dependencies for extensions admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all extension management operations
type ExtensionsAdminDeps struct {
	ListExtensions   fiber.Handler
	GetExtension     fiber.Handler
	EnableExtension  fiber.Handler
	DisableExtension fiber.Handler
	SyncExtensions   fiber.Handler
}

// BuildExtensionsAdminRoutes creates the extensions admin route group.
func BuildExtensionsAdminRoutes(deps *ExtensionsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "extensions_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin"},
		Routes: []Route{
			{Method: "GET", Path: "/extensions", Handler: deps.ListExtensions, Summary: "List extensions"},
			{Method: "GET", Path: "/extensions/:name", Handler: deps.GetExtension, Summary: "Get extension status"},
			{Method: "POST", Path: "/extensions/:name/enable", Handler: deps.EnableExtension, Summary: "Enable extension"},
			{Method: "POST", Path: "/extensions/:name/disable", Handler: deps.DisableExtension, Summary: "Disable extension"},
			{Method: "POST", Path: "/extensions/sync", Handler: deps.SyncExtensions, Summary: "Sync extensions"},
		},
	}
}

// ExtensionsTenantDeps contains dependencies for tenant-scoped extension routes.
// These routes allow tenant_admin users to manage extensions for their own tenant database.
//
// Role Access:
//   - tenant_admin: Manage extensions for their own tenant
//   - instance_admin: Manage extensions for any tenant
type ExtensionsTenantDeps struct {
	ListExtensions   fiber.Handler
	GetExtension     fiber.Handler
	EnableExtension  fiber.Handler
	DisableExtension fiber.Handler
}

// BuildExtensionsTenantRoutes creates the tenant-scoped extensions route group.
func BuildExtensionsTenantRoutes(deps *ExtensionsTenantDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "extensions_tenant",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			{Method: "GET", Path: "/extensions", Handler: deps.ListExtensions, Summary: "List extensions for tenant"},
			{Method: "GET", Path: "/extensions/:name", Handler: deps.GetExtension, Summary: "Get extension status for tenant"},
			{Method: "POST", Path: "/extensions/:name/enable", Handler: deps.EnableExtension, Summary: "Enable extension for tenant"},
			{Method: "POST", Path: "/extensions/:name/disable", Handler: deps.DisableExtension, Summary: "Disable extension for tenant"},
		},
	}
}
