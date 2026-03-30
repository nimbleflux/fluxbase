package routes

import (
	"github.com/gofiber/fiber/v3"
)

func BuildHealthRoutes(healthHandler fiber.Handler) *RouteGroup {
	return &RouteGroup{
		Name: "health",
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/",
				Handler: healthHandler,
				Summary: "Root health check",
				Auth:    AuthNone,
				Public:  true,
			},
			{
				Method:  "GET",
				Path:    "/health",
				Handler: healthHandler,
				Summary: "Detailed health check with database status",
				Auth:    AuthNone,
				Public:  true,
			},
		},
	}
}
