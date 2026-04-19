package routes

import "github.com/gofiber/fiber/v3"

type MCPDeps struct {
	BasePath         string
	MCPAuth          fiber.Handler
	TenantMiddleware fiber.Handler
	HandlePost       fiber.Handler
	HandleGet        fiber.Handler
	HandleHealth     fiber.Handler
}

func BuildMCPRoutes(deps *MCPDeps) *RouteGroup {
	var middlewares []Middleware
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantContext", Handler: deps.TenantMiddleware,
		})
	}

	return &RouteGroup{
		Name:        "mcp",
		Prefix:      deps.BasePath,
		Middlewares: middlewares,
		Routes: []Route{
			{Method: "POST", Path: "/", Handler: deps.HandlePost, Summary: "MCP JSON-RPC requests", Auth: AuthRequired},
			{Method: "GET", Path: "/", Handler: deps.HandleGet, Summary: "mcp SSE stream", Auth: AuthRequired},
		},
		SubGroups: []*RouteGroup{
			{
				Name:   "mcp-health",
				Prefix: "",
				Routes: []Route{
					{Method: "GET", Path: "/health", Handler: deps.HandleHealth, Summary: "MCP health check", Auth: AuthNone, Public: true},
				},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.MCPAuth,
		},
	}
}
