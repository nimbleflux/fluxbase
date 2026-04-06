package routes

import "github.com/gofiber/fiber/v3"

type CustomMCPDeps struct {
	RequireAuth  fiber.Handler
	RequireAdmin fiber.Handler

	GetConfig      fiber.Handler
	ListTools      fiber.Handler
	CreateTool     fiber.Handler
	SyncTool       fiber.Handler
	GetTool        fiber.Handler
	UpdateTool     fiber.Handler
	DeleteTool     fiber.Handler
	TestTool       fiber.Handler
	ListResources  fiber.Handler
	CreateResource fiber.Handler
	SyncResource   fiber.Handler
	GetResource    fiber.Handler
	UpdateResource fiber.Handler
	DeleteResource fiber.Handler
	TestResource   fiber.Handler
}

func BuildCustomMCPRoutes(deps *CustomMCPDeps) *RouteGroup {
	requireAdmin := []Middleware{{Name: "RequireAdmin", Handler: deps.RequireAdmin}}

	routes := []Route{
		{Method: "GET", Path: "/config", Handler: deps.GetConfig, Summary: "Get MCP configuration", Auth: AuthRequired, Middlewares: requireAdmin},

		{Method: "GET", Path: "/tools", Handler: deps.ListTools, Summary: "List custom MCP tools", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/tools", Handler: deps.CreateTool, Summary: "Create custom MCP tool", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/tools/sync", Handler: deps.SyncTool, Summary: "Sync custom MCP tool (upsert)", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "GET", Path: "/tools/:id", Handler: deps.GetTool, Summary: "Get custom MCP tool", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "PUT", Path: "/tools/:id", Handler: deps.UpdateTool, Summary: "Update custom MCP tool", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "DELETE", Path: "/tools/:id", Handler: deps.DeleteTool, Summary: "Delete custom MCP tool", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/tools/:id/test", Handler: deps.TestTool, Summary: "Test custom MCP tool", Auth: AuthRequired, Middlewares: requireAdmin},

		{Method: "GET", Path: "/resources", Handler: deps.ListResources, Summary: "List custom MCP resources", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/resources", Handler: deps.CreateResource, Summary: "Create custom MCP resource", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/resources/sync", Handler: deps.SyncResource, Summary: "Sync custom MCP resource (upsert)", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "GET", Path: "/resources/:id", Handler: deps.GetResource, Summary: "Get custom MCP resource", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "PUT", Path: "/resources/:id", Handler: deps.UpdateResource, Summary: "Update custom MCP resource", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "DELETE", Path: "/resources/:id", Handler: deps.DeleteResource, Summary: "Delete custom MCP resource", Auth: AuthRequired, Middlewares: requireAdmin},
		{Method: "POST", Path: "/resources/:id/test", Handler: deps.TestResource, Summary: "Test custom MCP resource", Auth: AuthRequired, Middlewares: requireAdmin},
	}

	return &RouteGroup{
		Name:   "custom-mcp",
		Prefix: "/api/v1/mcp",
		Routes: routes,
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
