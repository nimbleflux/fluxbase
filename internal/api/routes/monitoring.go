package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
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
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
		},
		Routes: []Route{
			{
				Method:      "GET",
				Path:        "/metrics",
				Handler:     deps.GetMetrics,
				Summary:     "Get system metrics",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeMonitoringRead},
				Middlewares: []Middleware{{Name: "RequireScope(monitoring:read)", Handler: deps.RequireScope(auth.ScopeMonitoringRead)}},
			},
			{
				Method:      "GET",
				Path:        "/health",
				Handler:     deps.GetHealth,
				Summary:     "Get system health status",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeMonitoringRead},
				Middlewares: []Middleware{{Name: "RequireScope(monitoring:read)", Handler: deps.RequireScope(auth.ScopeMonitoringRead)}},
			},
			{
				Method:      "GET",
				Path:        "/logs",
				Handler:     deps.GetLogs,
				Summary:     "Get system logs",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeMonitoringRead},
				Middlewares: []Middleware{{Name: "RequireScope(monitoring:read)", Handler: deps.RequireScope(auth.ScopeMonitoringRead)}},
			},
		},
	}
}
