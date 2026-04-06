package routes

import (
	"github.com/gofiber/fiber/v3"
)

type OpenAPIDeps struct {
	OptionalAuth   fiber.Handler
	GetOpenAPISpec fiber.Handler
}

func BuildOpenAPIRoutes(deps *OpenAPIDeps) *RouteGroup {
	return &RouteGroup{
		Name: "openapi",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/openapi.json",
				Handler: deps.GetOpenAPISpec,
				Summary: "OpenAPI specification (full spec for admins, minimal for others)",
				Auth:    AuthOptional,
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
		},
	}
}
