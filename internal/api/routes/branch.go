package routes

import (
	"github.com/gofiber/fiber/v3"
)

// BranchDeps contains dependencies for branch routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all branches (bypasses RLS)
//   - tenant_admin: Access to own tenant's branches (RLS enforced)
//   - service_role: Full access to all branches
type BranchDeps struct {
	GetActiveBranch    fiber.Handler
	SetActiveBranch    fiber.Handler
	ResetActiveBranch  fiber.Handler
	GetPoolStats       fiber.Handler
	CreateBranch       fiber.Handler
	ListBranches       fiber.Handler
	GetBranch          fiber.Handler
	DeleteBranch       fiber.Handler
	ResetBranch        fiber.Handler
	GetBranchActivity  fiber.Handler
	ListBranchAccess   fiber.Handler
	GrantBranchAccess  fiber.Handler
	RevokeBranchAccess fiber.Handler
	ListGitHubConfigs  fiber.Handler
	UpsertGitHubConfig fiber.Handler
	DeleteGitHubConfig fiber.Handler
}

// BuildBranchRoutes creates the branch route group.
// These routes are designed to be registered as a subgroup of admin routes,
// inheriting authentication and role middleware from the parent.
func BuildBranchRoutes(deps *BranchDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "branch",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/branches/active",
				Handler: deps.GetActiveBranch,
				Summary: "Get active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/branches/active",
				Handler: deps.SetActiveBranch,
				Summary: "Set active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/branches/active",
				Handler: deps.ResetActiveBranch,
				Summary: "Reset active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches/stats/pools",
				Handler: deps.GetPoolStats,
				Summary: "Get branch pool stats",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/branches",
				Handler: deps.CreateBranch,
				Summary: "Create a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches",
				Handler: deps.ListBranches,
				Summary: "List branches",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches/:id",
				Handler: deps.GetBranch,
				Summary: "Get a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/branches/:id",
				Handler: deps.DeleteBranch,
				Summary: "Delete a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/branches/:id/reset",
				Handler: deps.ResetBranch,
				Summary: "Reset a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches/:id/activity",
				Handler: deps.GetBranchActivity,
				Summary: "Get branch activity",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches/:id/access",
				Handler: deps.ListBranchAccess,
				Summary: "List branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/branches/:id/access",
				Handler: deps.GrantBranchAccess,
				Summary: "Grant branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/branches/:id/access/:user_id",
				Handler: deps.RevokeBranchAccess,
				Summary: "Revoke branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/branches/github/configs",
				Handler: deps.ListGitHubConfigs,
				Summary: "List GitHub configs",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/branches/github/configs",
				Handler: deps.UpsertGitHubConfig,
				Summary: "Upsert GitHub config",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/branches/github/configs/:repository",
				Handler: deps.DeleteGitHubConfig,
				Summary: "Delete GitHub config",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "tenant_admin", "service_role"},
			},
		},
	}
}
