package routes

import (
	"github.com/gofiber/fiber/v3"
)

// UsersAdminDeps contains dependencies for users admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all users, can delete users
//   - tenant_admin: Access to users within their tenant (RLS enforced), cannot delete users
type UsersAdminDeps struct {
	ListUsers           fiber.Handler
	InviteUser          fiber.Handler
	DeleteUser          fiber.Handler
	UpdateUser          fiber.Handler
	UpdateUserRole      fiber.Handler
	ResetUserPassword   fiber.Handler
	ListUsersWithQuotas fiber.Handler
	GetUserQuota        fiber.Handler
	SetUserQuota        fiber.Handler
	CreateInvitation    fiber.Handler
	ListInvitations     fiber.Handler
	RevokeInvitation    fiber.Handler
}

// BuildUsersAdminRoutes creates the users admin route group.
func BuildUsersAdminRoutes(deps *UsersAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "users_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			// Users (uses default roles)
			{Method: "GET", Path: "/users", Handler: deps.ListUsers, Summary: "List users"},
			{Method: "POST", Path: "/users/invite", Handler: deps.InviteUser, Summary: "Invite user"},
			{Method: "GET", Path: "/users/:id", Handler: deps.ListUsers, Summary: "Get user by ID"},
			{Method: "PATCH", Path: "/users/:id", Handler: deps.UpdateUser, Summary: "Update user"},
			{Method: "POST", Path: "/users/:id/reset-password", Handler: deps.ResetUserPassword, Summary: "Reset user password"},
			// Delete and role update are instance_admin only (override roles)
			{Method: "DELETE", Path: "/users/:id", Handler: deps.DeleteUser, Summary: "Delete user", Roles: []string{"admin", "instance_admin"}},
			{Method: "PATCH", Path: "/users/:id/role", Handler: deps.UpdateUserRole, Summary: "Update user role", Roles: []string{"admin", "instance_admin"}},
			// TODO: Implement quota handlers
			// {Method: "GET", Path: "/users/quotas", Handler: deps.ListUsersWithQuotas, Summary: "List users with quotas"},
			// {Method: "GET", Path: "/users/:id/quota", Handler: deps.GetUserQuota, Summary: "Get user quota"},
			// {Method: "PUT", Path: "/users/:id/quota", Handler: deps.SetUserQuota, Summary: "Set user quota"},

			// Invitations (uses default roles)
			{Method: "POST", Path: "/invitations", Handler: deps.CreateInvitation, Summary: "Create invitation"},
			{Method: "GET", Path: "/invitations", Handler: deps.ListInvitations, Summary: "List invitations"},
			{Method: "DELETE", Path: "/invitations/:id", Handler: deps.RevokeInvitation, Summary: "Revoke invitation"},
		},
	}
}
