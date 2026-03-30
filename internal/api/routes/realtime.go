package routes

import (
	"github.com/gofiber/fiber/v3"
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
		Middlewares: []Middleware{
			{Name: "RequireRealtimeEnabled", Handler: deps.RequireRealtimeEnabled},
		},
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/realtime",
				Handler: deps.HandleWebSocket,
				Summary: "WebSocket endpoint for realtime subscriptions",
				Auth:    AuthOptional,
				Scopes:  []string{"realtime:connect"},
				Public:  false,
			},
			{
				Method:  "GET",
				Path:    "/api/v1/realtime/stats",
				Handler: deps.HandleStats,
				Summary: "Get realtime connection statistics",
				Auth:    AuthRequired,
				Scopes:  []string{"realtime:connect"},
			},
			{
				Method:  "POST",
				Path:    "/api/v1/realtime/broadcast",
				Handler: deps.HandleBroadcast,
				Summary: "Broadcast message to all connected clients",
				Auth:    AuthRequired,
				Scopes:  []string{"realtime:broadcast"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
