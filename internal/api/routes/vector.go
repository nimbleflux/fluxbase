package routes

import (
	"github.com/gofiber/fiber/v3"
)

type VectorDeps struct {
	RequireAuth        fiber.Handler
	HandleCapabilities fiber.Handler
	HandleEmbed        fiber.Handler
	HandleSearch       fiber.Handler
}

func BuildVectorRoutes(deps *VectorDeps) *RouteGroup {
	return &RouteGroup{
		Name: "vector",
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
				Middlewares: []Middleware{
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Generate embeddings for text",
				Auth:    AuthRequired,
			},
			{
				Method:  "POST",
				Path:    "/api/v1/vector/search",
				Handler: deps.HandleSearch,
				Middlewares: []Middleware{
					{Name: "RequireAuth", Handler: deps.RequireAuth},
				},
				Summary: "Perform vector similarity search",
				Auth:    AuthRequired,
			},
		},
	}
}
