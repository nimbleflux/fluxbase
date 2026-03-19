package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
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
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/procedures",
				Handler: deps.ListProcedures,
				Middlewares: []Middleware{
					{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(rpc:read)", Handler: deps.RequireScope(auth.ScopeRPCRead)},
				},
				Summary: "List available RPC procedures",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeRPCRead},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/rpc/:namespace/:name",
				Handler: deps.Invoke,
				Middlewares: []Middleware{
					{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(rpc:execute)", Handler: deps.RequireScope(auth.ScopeRPCExecute)},
				},
				Summary: "Invoke RPC procedure",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeRPCExecute},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/executions/:id",
				Handler: deps.GetExecution,
				Middlewares: []Middleware{
					{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(rpc:read)", Handler: deps.RequireScope(auth.ScopeRPCRead)},
				},
				Summary: "Get RPC execution status",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeRPCRead},
			},
			{
				Method:  "GET",
				Path:    "/api/v1/rpc/executions/:id/logs",
				Handler: deps.GetExecutionLogs,
				Middlewares: []Middleware{
					{Name: "RequireRPCEnabled", Handler: deps.RequireRPCEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(rpc:read)", Handler: deps.RequireScope(auth.ScopeRPCRead)},
				},
				Summary: "Get RPC execution logs",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeRPCRead},
			},
		},
	}
}
