package routes

import (
	"github.com/gofiber/fiber/v3"
)

// LogsAdminDeps contains dependencies for logs admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all log management operations
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
		Name: "logs_admin",
		Routes: []Route{
			{Method: "GET", Path: "/logs", Handler: deps.ListLogs, Summary: "List logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/logs/stats", Handler: deps.GetLogStats, Summary: "Get log stats", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/logs/executions/:id", Handler: deps.GetExecutionLogsAdmin, Summary: "Get execution logs (admin)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/logs/flush", Handler: deps.FlushLogs, Summary: "Flush logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/logs/test", Handler: deps.GenerateTestLogs, Summary: "Generate test logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		},
	}
}
