package routes

import (
	"github.com/gofiber/fiber/v3"
)

type GraphQLDeps struct {
	OptionalAuth     fiber.Handler
	HandleGraphQL    fiber.Handler
	HandleIntrospect fiber.Handler
}

func BuildGraphQLRoutes(deps *GraphQLDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "graphql",
		Prefix: "/api/v1/graphql",
		Middlewares: []Middleware{
			{Name: "OptionalAuth", Handler: deps.OptionalAuth},
		},
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
	}
}
