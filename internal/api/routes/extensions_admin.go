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
