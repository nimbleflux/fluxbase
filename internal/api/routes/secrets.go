package routes

import (
	"github.com/gofiber/fiber/v3"
)

type SecretsDeps struct {
	RequireAuth        fiber.Handler
	RequireScope       func(...string) fiber.Handler
	ListSecrets        fiber.Handler
	GetStats           fiber.Handler
	GetSecretByName    fiber.Handler
	GetVersionsByName  fiber.Handler
	UpdateSecretByName fiber.Handler
	DeleteSecretByName fiber.Handler
	RollbackByName     fiber.Handler
	GetSecret          fiber.Handler
	GetVersions        fiber.Handler
	CreateSecret       fiber.Handler
	UpdateSecret       fiber.Handler
	DeleteSecret       fiber.Handler
	RollbackToVersion  fiber.Handler

	// Middleware for tenant context
	TenantMiddleware   fiber.Handler
	TenantDBMiddleware fiber.Handler
}

func BuildSecretsRoutes(deps *SecretsDeps) *RouteGroup {
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
		Name:        "secrets",
		Prefix:      "/api/v1/secrets",
		Middlewares: middlewares,
		Routes: []Route{
			{
				Method:  "GET",
				Path:    "/",
				Handler: deps.ListSecrets,
				Summary: "List secrets",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "GET",
				Path:    "/stats",
				Handler: deps.GetStats,
				Summary: "Get secrets stats",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "GET",
				Path:    "/by-name/:name",
				Handler: deps.GetSecretByName,
				Summary: "Get secret by name",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "GET",
				Path:    "/by-name/:name/versions",
				Handler: deps.GetVersionsByName,
				Summary: "Get secret versions by name",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "PUT",
				Path:    "/by-name/:name",
				Handler: deps.UpdateSecretByName,
				Summary: "Update secret by name",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/by-name/:name",
				Handler: deps.DeleteSecretByName,
				Summary: "Delete secret by name",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "POST",
				Path:    "/by-name/:name/rollback/:version",
				Handler: deps.RollbackByName,
				Summary: "Rollback secret by name",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "GET",
				Path:    "/:id",
				Handler: deps.GetSecret,
				Summary: "Get secret by ID",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "GET",
				Path:    "/:id/versions",
				Handler: deps.GetVersions,
				Summary: "Get secret versions by ID",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:read"},
			},
			{
				Method:  "POST",
				Path:    "/",
				Handler: deps.CreateSecret,
				Summary: "Create a secret",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "PUT",
				Path:    "/:id",
				Handler: deps.UpdateSecret,
				Summary: "Update a secret",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "DELETE",
				Path:    "/:id",
				Handler: deps.DeleteSecret,
				Summary: "Delete a secret",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
			{
				Method:  "POST",
				Path:    "/:id/rollback/:version",
				Handler: deps.RollbackToVersion,
				Summary: "Rollback secret to version",
				Auth:    AuthRequired,
				Scopes:  []string{"secrets:write"},
			},
		},
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.RequireAuth,
		},
		RequireScope: deps.RequireScope,
	}
}
