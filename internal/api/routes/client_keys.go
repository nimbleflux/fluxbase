package routes

import (
	"github.com/gofiber/fiber/v3"
)

type ClientKeysDeps struct {
	RequireAuth                      fiber.Handler
	RequireAdminIfClientKeysDisabled fiber.Handler
	RequireScope                     func(...string) fiber.Handler
	TenantMiddleware                 fiber.Handler
	ListClientKeys                   fiber.Handler
	GetClientKey                     fiber.Handler
	CreateClientKey                  fiber.Handler
	UpdateClientKey                  fiber.Handler
	DeleteClientKey                  fiber.Handler
	RevokeClientKey                  fiber.Handler
}

func BuildClientKeysRoutes(deps *ClientKeysDeps) *RouteGroup {
	var middlewares []Middleware
	middlewares = append(middlewares, Middleware{
		Name:    "RequireAdminIfClientKeysDisabled",
		Handler: deps.RequireAdminIfClientKeysDisabled,
	})
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name:    "TenantContext",
			Handler: deps.TenantMiddleware,
		})
	}

	return &RouteGroup{
		Name:        "client_keys",
		Prefix:      "/api/v1/client-keys",
		Middlewares: middlewares,
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.ListClientKeys,
				Summary: "List client keys",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:read"},
			},
			{
				Method:  "GET",
				Path:    "/:id",
				Handler: deps.GetClientKey,
				Summary: "Get a client key",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:read"},
			},
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.CreateClientKey,
				Summary: "Create a client key",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:write"},
			},
			{
				Method:  "PATCH",
				Path:    "/:id",
				Handler: deps.UpdateClientKey,
				Summary: "Update a client key",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/:id",
				Handler: deps.DeleteClientKey,
				Summary: "Delete a client key",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:write"},
			},
			{
				Method:  "POST",
				Path:    "/:id/revoke",
				Handler: deps.RevokeClientKey,
				Summary: "Revoke a client key",
				Auth:    AuthRequired,
				Scopes:  []string{"clientkeys:write"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
