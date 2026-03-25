package routes

import (
	"github.com/gofiber/fiber/v3"
)

// ServiceKeysAdminDeps contains dependencies for service keys admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all service keys
type ServiceKeysAdminDeps struct {
	ListServiceKeys      fiber.Handler
	GetServiceKey        fiber.Handler
	CreateServiceKey     fiber.Handler
	UpdateServiceKey     fiber.Handler
	DeleteServiceKey     fiber.Handler
	DisableServiceKey    fiber.Handler
	EnableServiceKey     fiber.Handler
	RevokeServiceKey     fiber.Handler
	DeprecateServiceKey  fiber.Handler
	RotateServiceKey     fiber.Handler
	GetRevocationHistory fiber.Handler
}

// BuildServiceKeysAdminRoutes creates the service keys admin route group.
func BuildServiceKeysAdminRoutes(deps *ServiceKeysAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "service_keys_admin",
		Routes: []Route{
			{Method: "GET", Path: "/service-keys", Handler: deps.ListServiceKeys, Summary: "List service keys", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys", Handler: deps.CreateServiceKey, Summary: "Create service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/service-keys/:id", Handler: deps.GetServiceKey, Summary: "Get service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/service-keys/:id", Handler: deps.UpdateServiceKey, Summary: "Update service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "DELETE", Path: "/service-keys/:id", Handler: deps.DeleteServiceKey, Summary: "Delete service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys/:id/disable", Handler: deps.DisableServiceKey, Summary: "Disable service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys/:id/enable", Handler: deps.EnableServiceKey, Summary: "Enable service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys/:id/revoke", Handler: deps.RevokeServiceKey, Summary: "Revoke service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys/:id/deprecate", Handler: deps.DeprecateServiceKey, Summary: "Deprecate service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/service-keys/:id/rotate", Handler: deps.RotateServiceKey, Summary: "Rotate service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/service-keys/:id/revocations", Handler: deps.GetRevocationHistory, Summary: "Get revocation history", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		},
	}
}
