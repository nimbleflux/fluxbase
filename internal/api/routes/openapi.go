package routes

import (
	"github.com/gofiber/fiber/v3"
)

type OpenAPIDeps struct {
	OptionalAuth    fiber.Handler
	TenantContext   fiber.Handler
	TenantDBContext fiber.Handler
	GetOpenAPISpec  fiber.Handler
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
		Middlewares: []Middleware{
			{Name: "TenantContext", Handler: deps.TenantContext},
			{Name: "TenantDBContext", Handler: deps.TenantDBContext},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Optional: deps.OptionalAuth,
		},
	}
}
