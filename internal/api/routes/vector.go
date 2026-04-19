package routes

import (
	"github.com/gofiber/fiber/v3"
)

type VectorDeps struct {
	RequireAuth        fiber.Handler
	TenantMiddleware   fiber.Handler
	HandleCapabilities fiber.Handler
	HandleEmbed        fiber.Handler
	HandleSearch       fiber.Handler
}

func BuildVectorRoutes(deps *VectorDeps) *RouteGroup {
	var middlewares []Middleware
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantContext", Handler: deps.TenantMiddleware,
		})
	}

	return &RouteGroup{
		Name:        "vector",
		Middlewares: middlewares,
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/api/v1/capabilities/vector",
				Handler: deps.HandleCapabilities,
				Summary: "Get vector search capabilities (pgvector status)",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/vector/embed",
				Handler: deps.HandleEmbed,
				Summary: "Generate embeddings for text",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/vector/search",
				Handler: deps.HandleSearch,
				Summary: "Perform vector similarity search",
				Auth:    AuthRequired,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
	}
}
