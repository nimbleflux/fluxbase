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
	ListMyTenants             fiber.Handler
	ListTenants               fiber.Handler
	CreateTenant              fiber.Handler
	GetTenant                 fiber.Handler
	UpdateTenant              fiber.Handler
	DeleteTenant              fiber.Handler
	MigrateTenant             fiber.Handler
	RepairTenant              fiber.Handler
	ListAdmins                fiber.Handler
	AssignAdmin               fiber.Handler
	RemoveAdmin               fiber.Handler
	GetTenantSettings         fiber.Handler
	UpdateTenantSettings      fiber.Handler
	DeleteTenantSetting       fiber.Handler
	GetTenantSetting          fiber.Handler
	GetTenantSchemaStatus     fiber.Handler
	ApplyTenantSchema         fiber.Handler
	GetStoredSchema           fiber.Handler
	UploadTenantSchema        fiber.Handler
	ApplyUploadedTenantSchema fiber.Handler
	DeleteStoredSchema        fiber.Handler
}

// BuildTenantsAdminRoutes creates the tenants admin route group.
func BuildTenantsAdminRoutes(deps *TenantsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "tenants_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			// Tenant listing
			{Method: "GET", Path: "/tenants/mine", Handler: deps.ListMyTenants, Summary: "List my tenants", Roles: nil}, // No role restriction
			{Method: "GET", Path: "/tenants", Handler: deps.ListTenants, Summary: "List all tenants", Roles: []string{"admin", "instance_admin"}},

			// Tenant CRUD - instance admin only for create/delete
			{Method: "POST", Path: "/tenants", Handler: deps.CreateTenant, Summary: "Create tenant", Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/tenants/:id", Handler: deps.GetTenant, Summary: "Get tenant", Roles: nil}, // No role restriction
			{Method: "PATCH", Path: "/tenants/:id", Handler: deps.UpdateTenant, Summary: "Update tenant"},     // Uses default roles
			{Method: "DELETE", Path: "/tenants/:id", Handler: deps.DeleteTenant, Summary: "Delete tenant", Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/tenants/:id/migrate", Handler: deps.MigrateTenant, Summary: "Migrate tenant", Roles: []string{"admin", "instance_admin"}},

			// Tenant Admins/Members (uses default roles)
			{Method: "GET", Path: "/tenants/:id/admins", Handler: deps.ListAdmins, Summary: "List tenant admins", Roles: nil}, // No role restriction
			{Method: "POST", Path: "/tenants/:id/admins", Handler: deps.AssignAdmin, Summary: "Assign tenant admin"},
			{Method: "DELETE", Path: "/tenants/:id/admins/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant admin"},

			// Tenant members (aliases for frontend compatibility - backend uses /admins)
			{Method: "GET", Path: "/tenants/:id/members", Handler: deps.ListAdmins, Summary: "List tenant members", Roles: nil}, // No role restriction
			{Method: "POST", Path: "/tenants/:id/members", Handler: deps.AssignAdmin, Summary: "Add tenant member"},
			{Method: "PATCH", Path: "/tenants/:id/members/:user_id", Handler: deps.AssignAdmin, Summary: "Update tenant member"},
			{Method: "DELETE", Path: "/tenants/:id/members/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant member"},

			// Tenant Settings (uses default roles)
			{Method: "GET", Path: "/tenants/:id/settings", Handler: deps.GetTenantSettings, Summary: "Get tenant settings"},
			{Method: "PATCH", Path: "/tenants/:id/settings", Handler: deps.UpdateTenantSettings, Summary: "Update tenant settings"},
			{Method: "DELETE", Path: "/tenants/:id/settings/*", Handler: deps.DeleteTenantSetting, Summary: "Delete tenant setting"},
			{Method: "GET", Path: "/tenants/:id/settings/*", Handler: deps.GetTenantSetting, Summary: "Get tenant setting"},

			// Tenant Declarative Schema
			{Method: "GET", Path: "/tenants/:id/schema", Handler: deps.GetTenantSchemaStatus, Summary: "Get tenant schema status"},
			{Method: "POST", Path: "/tenants/:id/schema/apply", Handler: deps.ApplyTenantSchema, Summary: "Apply tenant schema", Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Tenant Schema Content Management (API-driven)
			{Method: "GET", Path: "/tenants/:id/schema/content", Handler: deps.GetStoredSchema, Summary: "Get stored schema content"},
			{Method: "POST", Path: "/tenants/:id/schema/content", Handler: deps.UploadTenantSchema, Summary: "Upload schema content"},
			{Method: "POST", Path: "/tenants/:id/schema/content/apply", Handler: deps.ApplyUploadedTenantSchema, Summary: "Apply uploaded schema"},
			{Method: "DELETE", Path: "/tenants/:id/schema/content", Handler: deps.DeleteStoredSchema, Summary: "Delete stored schema"},
		},
	}
}
