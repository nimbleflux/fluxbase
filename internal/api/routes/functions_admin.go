package routes

import (
	"github.com/gofiber/fiber/v3"
)

// FunctionsAdminDeps contains dependencies for functions admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to function management and reload
//   - tenant_admin: Can view function executions and logs for their tenant
type FunctionsAdminDeps struct {
	ReloadFunctions        fiber.Handler
	ListFunctionNamespaces fiber.Handler
	ListAllExecutions      fiber.Handler
	GetExecutionLogs       fiber.Handler
	SyncFunctions          fiber.Handler
}

// BuildFunctionsAdminRoutes creates the functions admin route group.
func BuildFunctionsAdminRoutes(deps *FunctionsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "functions_admin",
		Routes: []Route{
			{Method: "POST", Path: "/functions/reload", Handler: deps.ReloadFunctions, Summary: "Reload functions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/functions/namespaces", Handler: deps.ListFunctionNamespaces, Summary: "List function namespaces", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/functions/executions", Handler: deps.ListAllExecutions, Summary: "List all function executions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/functions/executions/:id/logs", Handler: deps.GetExecutionLogs, Summary: "Get function execution logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
