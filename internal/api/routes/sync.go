package routes

import (
	"github.com/gofiber/fiber/v3"
)

type SyncDeps struct {
	RequireSyncAuth fiber.Handler
	RequireRole     fiber.Handler
	TenantContext   fiber.Handler

	// Function sync
	RequireFunctionsSyncIPAllowlist fiber.Handler
	SyncFunctions                   fiber.Handler

	// Jobs sync
	RequireJobsSyncIPAllowlist fiber.Handler
	SyncJobs                   fiber.Handler

	// AI sync
	RequireAIEnabled         fiber.Handler
	RequireAISyncIPAllowlist fiber.Handler
	SyncChatbots             fiber.Handler

	// RPC sync
	RequireRPCEnabled         fiber.Handler
	RequireRPCSyncIPAllowlist fiber.Handler
	SyncProcedures            fiber.Handler
}

func BuildSyncRoutes(deps *SyncDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	routes := []Route{}

	// Functions sync
	if deps.SyncFunctions != nil {
		routes = append(routes, Route{
			Method:  "POST",
			Path:    "/api/v1/admin/functions/sync",
			Handler: deps.SyncFunctions,
			Middlewares: []Middleware{
				{Name: "RequireFunctionsSyncIPAllowlist", Handler: deps.RequireFunctionsSyncIPAllowlist},
				{Name: "RequireRole", Handler: deps.RequireRole},
			},
			Summary: "Sync functions from filesystem",
			Auth:    AuthRequired,
			Roles:   []string{"admin", "instance_admin", "service_role"},
		})
	}

	// Jobs sync
	if deps.SyncJobs != nil {
		routes = append(routes, Route{
			Method:  "POST",
			Path:    "/api/v1/admin/jobs/sync",
			Handler: deps.SyncJobs,
			Middlewares: []Middleware{
				{Name: "RequireJobsSyncIPAllowlist", Handler: deps.RequireJobsSyncIPAllowlist},
				{Name: "RequireRole", Handler: deps.RequireRole},
			},
			Summary: "Sync jobs from filesystem",
			Auth:    AuthRequired,
			Roles:   []string{"admin", "instance_admin", "service_role"},
		})
	}

	// AI chatbots sync
	if deps.SyncChatbots != nil {
		routes = append(routes, Route{
			Method:  "POST",
			Path:    "/api/v1/admin/ai/chatbots/sync",
			Handler: deps.SyncChatbots,
			Middlewares: []Middleware{
				{Name: "RequireAIEnabled", Handler: deps.RequireAIEnabled},
				{Name: "RequireAISyncIPAllowlist", Handler: deps.RequireAISyncIPAllowlist},
				{Name: "RequireRole", Handler: deps.RequireRole},
			},
			Summary: "Sync AI chatbots from filesystem",
			Auth:    AuthRequired,
			Roles:   []string{"admin", "instance_admin", "service_role"},
		})
	}

	// RPC sync
	if deps.SyncProcedures != nil {
		routes = append(routes, Route{
			Method:  "POST",
			Path:    "/api/v1/admin/rpc/sync",
			Handler: deps.SyncProcedures,
			Middlewares: []Middleware{
				{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
				{Name: "RequireRPCSyncIPAllowlist", Handler: deps.RequireRPCSyncIPAllowlist},
				{Name: "RequireRole", Handler: deps.RequireRole},
			},
			Summary: "Sync RPC procedures from database",
			Auth:    AuthRequired,
			Roles:   []string{"admin", "instance_admin", "service_role"},
		})
	}

	return &RouteGroup{
		Name:   "sync",
		Routes: routes,
		Middlewares: []Middleware{
			{Name: "TenantContext", Handler: deps.TenantContext},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireSyncAuth,
		},
	}
}
