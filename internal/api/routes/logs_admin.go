package routes

import (
	"github.com/gofiber/fiber/v3"
)

// LogsAdminDeps contains dependencies for logs admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all log management operations
//   - tenant_admin: View tenant-scoped logs (RLS enforced), no flush/test
type LogsAdminDeps struct {
	ListLogs              fiber.Handler
	GetLogStats           fiber.Handler
	GetExecutionLogsAdmin fiber.Handler
	FlushLogs             fiber.Handler
	GenerateTestLogs      fiber.Handler
}

// BuildLogsAdminRoutes creates the logs admin route group.
func BuildLogsAdminRoutes(deps *LogsAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name:         "logs_admin",
		DefaultAuth:  AuthRequired,
		DefaultRoles: []string{"admin", "instance_admin", "tenant_admin"},
		Routes: []Route{
			{Method: "GET", Path: "/logs", Handler: deps.ListLogs, Summary: "List logs"},
			{Method: "GET", Path: "/logs/stats", Handler: deps.GetLogStats, Summary: "Get log stats"},
			{Method: "GET", Path: "/logs/executions/:id", Handler: deps.GetExecutionLogsAdmin, Summary: "Get execution logs (admin)"},
			// Flush and test are instance_admin only (override default roles)
			{Method: "POST", Path: "/logs/flush", Handler: deps.FlushLogs, Summary: "Flush logs", Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/logs/test", Handler: deps.GenerateTestLogs, Summary: "Generate test logs", Roles: []string{"admin", "instance_admin"}},
		},
	}
}
