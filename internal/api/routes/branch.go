package routes

import (
	"github.com/gofiber/fiber/v3"
)

type BranchDeps struct {
	RequireAuth        fiber.Handler
	RequireRole        fiber.Handler
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

func BuildBranchRoutes(deps *BranchDeps) *RouteGroup {
	if deps == nil {
		return nil
	}
	requireRole := []Middleware{{Name: "RequireRole", Handler: deps.RequireRole}}

	return &RouteGroup{
		Name:        "branch",
		Middlewares: requireRole,
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/active",
				Handler: deps.GetActiveBranch,
				Summary: "Get active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/admin/branches/active",
				Handler: deps.SetActiveBranch,
				Summary: "Set active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/admin/branches/active",
				Handler: deps.ResetActiveBranch,
				Summary: "Reset active branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/stats/pools",
				Handler: deps.GetPoolStats,
				Summary: "Get branch pool stats",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/admin/branches",
				Handler: deps.CreateBranch,
				Summary: "Create a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches",
				Handler: deps.ListBranches,
				Summary: "List branches",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/:id",
				Handler: deps.GetBranch,
				Summary: "Get a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/admin/branches/:id",
				Handler: deps.DeleteBranch,
				Summary: "Delete a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/admin/branches/:id/reset",
				Handler: deps.ResetBranch,
				Summary: "Reset a branch",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/:id/activity",
				Handler: deps.GetBranchActivity,
				Summary: "Get branch activity",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/:id/access",
				Handler: deps.ListBranchAccess,
				Summary: "List branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/admin/branches/:id/access",
				Handler: deps.GrantBranchAccess,
				Summary: "Grant branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/admin/branches/:id/access/:user_id",
				Handler: deps.RevokeBranchAccess,
				Summary: "Revoke branch access",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/admin/branches/github/configs",
				Handler: deps.ListGitHubConfigs,
				Summary: "List GitHub configs",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/admin/branches/github/configs",
				Handler: deps.UpsertGitHubConfig,
				Summary: "Upsert GitHub config",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
			{
				Method:  "DELETE",
				Path:    "/api/v1/admin/branches/github/configs/:repository",
				Handler: deps.DeleteGitHubConfig,
				Summary: "Delete GitHub config",
				Auth:    AuthRequired,
				Roles:   []string{"admin", "instance_admin", "service_role"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
