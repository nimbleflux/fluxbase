package routes

import (
	"github.com/gofiber/fiber/v3"
)

// RPCAdminDeps contains dependencies for RPC admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all RPC management operations
//   - tenant_admin: View/cancel tenant-scoped executions (RLS enforced)
type RPCAdminDeps struct {
	ListRPCNamespaces   fiber.Handler
	ListProcedures      fiber.Handler
	GetProcedure        fiber.Handler
	UpdateProcedure     fiber.Handler
	DeleteProcedure     fiber.Handler
	SyncProcedures      fiber.Handler
	ListRPCExecutions   fiber.Handler
	GetRPCExecution     fiber.Handler
	GetRPCExecutionLogs fiber.Handler
	CancelRPCExecution  fiber.Handler
}

// BuildRPCAdminRoutes creates the RPC admin route group.
func BuildRPCAdminRoutes(deps *RPCAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "rpc_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin"},
		Routes: []Route{
			// Procedure management (instance_admin only - uses default roles)
			{Method: "GET", Path: "/rpc/namespaces", Handler: deps.ListRPCNamespaces, Summary: "List RPC namespaces"},
			{Method: "GET", Path: "/rpc/procedures", Handler: deps.ListProcedures, Summary: "List RPC procedures"},
			{Method: "GET", Path: "/rpc/procedures/:namespace/:name", Handler: deps.GetProcedure, Summary: "Get RPC procedure"},
			{Method: "PUT", Path: "/rpc/procedures/:namespace/:name", Handler: deps.UpdateProcedure, Summary: "Update RPC procedure"},
			{Method: "DELETE", Path: "/rpc/procedures/:namespace/:name", Handler: deps.DeleteProcedure, Summary: "Delete RPC procedure"},
			{Method: "POST", Path: "/rpc/sync", Handler: deps.SyncProcedures, Summary: "Sync RPC procedures"},

			// Executions (tenant_admin can access - override roles)
			{Method: "GET", Path: "/rpc/executions", Handler: deps.ListRPCExecutions, Summary: "List RPC executions", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/rpc/executions/:id", Handler: deps.GetRPCExecution, Summary: "Get RPC execution", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/rpc/executions/:id/logs", Handler: deps.GetRPCExecutionLogs, Summary: "Get RPC execution logs", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/rpc/executions/:id/cancel", Handler: deps.CancelRPCExecution, Summary: "Cancel RPC execution", Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
