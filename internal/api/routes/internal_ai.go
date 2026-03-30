package routes

import (
	"github.com/gofiber/fiber/v3"
)

type InternalAIDeps struct {
	RequireInternal     fiber.Handler
	RequireAuth         fiber.Handler
	HandleChat          fiber.Handler
	HandleEmbed         fiber.Handler
	HandleListProviders fiber.Handler
}

func BuildInternalAIRoutes(deps *InternalAIDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "internal-ai",
		Prefix: "/api/v1/internal/ai",
		Middlewares: []Middleware{
			{Name: "RequireInternal", Handler: deps.RequireInternal},
		},
		Routes: []Route{
			{Method: "POST", Path: "/chat", Handler: deps.HandleChat, Summary: "Internal AI chat for MCP tools/functions/jobs", Auth: AuthInternal, Internal: true},
			{Method: "POST", Path: "/embed", Handler: deps.HandleEmbed, Summary: "Internal AI embed for MCP tools/functions/jobs", Auth: AuthInternal, Internal: true},
			{Method: "GET", Path: "/providers", Handler: deps.HandleListProviders, Summary: "List internal AI providers", Auth: AuthInternal, Internal: true},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Internal: deps.RequireInternal,
		},
	}
}
