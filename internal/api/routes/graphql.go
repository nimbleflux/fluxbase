package routes

import (
	"github.com/gofiber/fiber/v3"
)

type GraphQLDeps struct {
	OptionalAuth     fiber.Handler
	HandleGraphQL    fiber.Handler
	HandleIntrospect fiber.Handler

	// Middleware for tenant context
	TenantMiddleware   fiber.Handler
	TenantDBMiddleware fiber.Handler
}

func BuildGraphQLRoutes(deps *GraphQLDeps) *RouteGroup {
	// Build middlewares for tenant context
	var middlewares []Middleware
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantContext", Handler: deps.TenantMiddleware,
		})
	}
	if deps.TenantDBMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name: "TenantDBContext", Handler: deps.TenantDBMiddleware,
		})
	}

	return &RouteGroup{
		Name:        "graphql",
		Prefix:      "/api/v1/graphql",
		Middlewares: middlewares,
		Routes: []Route{
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.HandleGraphQL,
				Summary: "Execute GraphQL query",
				Auth:    AuthOptional,
			},
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.HandleIntrospect,
				Summary: "GraphQL schema introspection",
				Auth:    AuthOptional,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
		},
	}
}
