package routes

import (
	"github.com/gofiber/fiber/v3"
)

// TenantsAdminDeps contains dependencies for tenants admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to create/delete tenants, manage all tenant admins
//   - tenant_admin: Access to view/update own tenant, manage own tenant settings and members
type TenantsAdminDeps struct {
	ListMyTenants        fiber.Handler
	ListTenants          fiber.Handler
	CreateTenant         fiber.Handler
	GetTenant            fiber.Handler
	UpdateTenant         fiber.Handler
	DeleteTenant         fiber.Handler
	MigrateTenant        fiber.Handler
	ListAdmins           fiber.Handler
	AssignAdmin          fiber.Handler
	RemoveAdmin          fiber.Handler
	GetTenantSettings    fiber.Handler
	UpdateTenantSettings fiber.Handler
	DeleteTenantSetting  fiber.Handler
	GetTenantSetting     fiber.Handler
}

// BuildTenantsAdminRoutes creates the tenants admin route group.
func BuildTenantsAdminRoutes(deps *TenantsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "tenants_admin",
		Routes: []Route{
			// Tenant listing
			{Method: "GET", Path: "/tenants/mine", Handler: deps.ListMyTenants, Summary: "List my tenants", Auth: AuthRequired},
			{Method: "GET", Path: "/tenants", Handler: deps.ListTenants, Summary: "List all tenants", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Tenant CRUD - instance admin only for create/delete
			{Method: "POST", Path: "/tenants", Handler: deps.CreateTenant, Summary: "Create tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/tenants/:id", Handler: deps.GetTenant, Summary: "Get tenant", Auth: AuthRequired},
			{Method: "PATCH", Path: "/tenants/:id", Handler: deps.UpdateTenant, Summary: "Update tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/tenants/:id", Handler: deps.DeleteTenant, Summary: "Delete tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/tenants/:id/migrate", Handler: deps.MigrateTenant, Summary: "Migrate tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Tenant Admins/Members
			{Method: "GET", Path: "/tenants/:id/admins", Handler: deps.ListAdmins, Summary: "List tenant admins", Auth: AuthRequired},
			{Method: "POST", Path: "/tenants/:id/admins", Handler: deps.AssignAdmin, Summary: "Assign tenant admin", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/tenants/:id/admins/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant admin", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Tenant members (aliases for frontend compatibility - backend uses /admins)
			{Method: "GET", Path: "/tenants/:id/members", Handler: deps.ListAdmins, Summary: "List tenant members", Auth: AuthRequired},
			{Method: "POST", Path: "/tenants/:id/members", Handler: deps.AssignAdmin, Summary: "Add tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PATCH", Path: "/tenants/:id/members/:user_id", Handler: deps.AssignAdmin, Summary: "Update tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/tenants/:id/members/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Tenant Settings
			{Method: "GET", Path: "/tenants/:id/settings", Handler: deps.GetTenantSettings, Summary: "Get tenant settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PATCH", Path: "/tenants/:id/settings", Handler: deps.UpdateTenantSettings, Summary: "Update tenant settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/tenants/:id/settings/*", Handler: deps.DeleteTenantSetting, Summary: "Delete tenant setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/tenants/:id/settings/*", Handler: deps.GetTenantSetting, Summary: "Get tenant setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
