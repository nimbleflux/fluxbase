package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
)

type RealtimeDeps struct {
	RequireRealtimeEnabled fiber.Handler
	OptionalAuth           fiber.Handler
	RequireAuth            fiber.Handler
	RequireScope           func(...string) fiber.Handler
	HandleWebSocket        fiber.Handler
	HandleStats            fiber.Handler
	HandleBroadcast        fiber.Handler
}

func BuildRealtimeRoutes(deps *RealtimeDeps) *RouteGroup {
	return &RouteGroup{
		Name: "realtime",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/realtime",
				Handler: deps.HandleWebSocket,
				Middlewares: []Middleware{
					{Name: "RequireRealtimeEnabled", Handler: deps.RequireRealtimeEnabled},
					{Name: "OptionalAuth", Handler: deps.OptionalAuth},
					{Name: "RequireScope(realtime:connect)", Handler: deps.RequireScope(auth.ScopeRealtimeConnect)},
				},
				Summary: "WebSocket endpoint for realtime subscriptions",
				Auth:    AuthOptional,
				Scopes:  []string{auth.ScopeRealtimeConnect},
				Public:  false,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/realtime/stats",
				Handler: deps.HandleStats,
				Middlewares: []Middleware{
					{Name: "RequireRealtimeEnabled", Handler: deps.RequireRealtimeEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(realtime:connect)", Handler: deps.RequireScope(auth.ScopeRealtimeConnect)},
				},
				Summary: "Get realtime connection statistics",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeRealtimeConnect},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/realtime/broadcast",
				Handler: deps.HandleBroadcast,
				Middlewares: []Middleware{
					{Name: "RequireRealtimeEnabled", Handler: deps.RequireRealtimeEnabled},
					{Name: "RequireAuth", Handler: deps.RequireAuth},
					{Name: "RequireScope(realtime:broadcast)", Handler: deps.RequireScope(auth.ScopeRealtimeBroadcast)},
				},
				Summary: "Broadcast message to all connected clients",
				Auth:    AuthRequired,
				Scopes:  []string{auth.ScopeRealtimeBroadcast},
			},
		},
	}
}
