package routes

import (
	"github.com/gofiber/fiber/v3"
)

type RPCDeps struct {
	RequireRPCEnabled fiber.Handler
	OptionalAuth      fiber.Handler
	RequireScope      func(...string) fiber.Handler
	ListProcedures    fiber.Handler
	Invoke            fiber.Handler
	GetExecution      fiber.Handler
	GetExecutionLogs  fiber.Handler
}

func BuildRPCRoutes(deps *RPCDeps) *RouteGroup {
	return &RouteGroup{
		Name: "rpc",
		Middlewares: []Middleware{
			{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/procedures",
				Handler: deps.ListProcedures,
				Summary: "List available RPC procedures",
				Auth:    AuthOptional,
				Scopes:  []string{"rpc:read"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/rpc/:namespace/:name",
				Handler: deps.Invoke,
				Summary: "Invoke RPC procedure",
				Auth:    AuthOptional,
				Scopes:  []string{"rpc:execute"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/executions/:id",
				Handler: deps.GetExecution,
				Summary: "Get RPC execution status",
				Auth:    AuthOptional,
				Scopes:  []string{"rpc:read"},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/executions/:id/logs",
				Handler: deps.GetExecutionLogs,
				Summary: "Get RPC execution logs",
				Auth:    AuthOptional,
				Scopes:  []string{"rpc:read"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
