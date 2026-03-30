package routes

import (
	"github.com/gofiber/fiber/v3"
)

type MonitoringDeps struct {
	RequireAuth  fiber.Handler
	RequireScope func(...string) fiber.Handler
	GetMetrics   fiber.Handler
	GetHealth    fiber.Handler
	GetLogs      fiber.Handler
}

func BuildMonitoringRoutes(deps *MonitoringDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "monitoring",
		Prefix: "/api/v1/monitoring",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/metrics",
				Handler: deps.GetMetrics,
				Summary: "Get system metrics",
				Auth:    AuthRequired,
				Scopes:  []string{"monitoring:read"},
			},
			{
				Method:  "GET",
				Path:    "/health",
				Handler: deps.GetHealth,
				Summary: "Get system health status",
				Auth:    AuthRequired,
				Scopes:  []string{"monitoring:read"},
			},
			{
				Method:  "GET",
				Path:    "/logs",
				Handler: deps.GetLogs,
				Summary: "Get system logs",
				Auth:    AuthRequired,
				Scopes:  []string{"monitoring:read"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
