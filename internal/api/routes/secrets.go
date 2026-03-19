package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
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
}

func BuildSecretsRoutes(deps *SecretsDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "secrets",
		Prefix: "/api/v1/secrets",
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
		},
		Routes: []Route{
			{
				Method:      "GET",
				Path:        "/",
				Handler:     deps.ListSecrets,
				Summary:     "List secrets",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "GET",
				Path:        "/stats",
				Handler:     deps.GetStats,
				Summary:     "Get secrets stats",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "GET",
				Path:        "/by-name/:name",
				Handler:     deps.GetSecretByName,
				Summary:     "Get secret by name",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "GET",
				Path:        "/by-name/:name/versions",
				Handler:     deps.GetVersionsByName,
				Summary:     "Get secret versions by name",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "PUT",
				Path:        "/by-name/:name",
				Handler:     deps.UpdateSecretByName,
				Summary:     "Update secret by name",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "DELETE",
				Path:        "/by-name/:name",
				Handler:     deps.DeleteSecretByName,
				Summary:     "Delete secret by name",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "POST",
				Path:        "/by-name/:name/rollback/:version",
				Handler:     deps.RollbackByName,
				Summary:     "Rollback secret by name",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "GET",
				Path:        "/:id",
				Handler:     deps.GetSecret,
				Summary:     "Get secret by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "GET",
				Path:        "/:id/versions",
				Handler:     deps.GetVersions,
				Summary:     "Get secret versions by ID",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsRead},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:read)", Handler: deps.RequireScope(auth.ScopeSecretsRead)}},
			},
			{
				Method:      "POST",
				Path:        "/",
				Handler:     deps.CreateSecret,
				Summary:     "Create a secret",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "PUT",
				Path:        "/:id",
				Handler:     deps.UpdateSecret,
				Summary:     "Update a secret",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "DELETE",
				Path:        "/:id",
				Handler:     deps.DeleteSecret,
				Summary:     "Delete a secret",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
			{
				Method:      "POST",
				Path:        "/:id/rollback/:version",
				Handler:     deps.RollbackToVersion,
				Summary:     "Rollback secret to version",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeSecretsWrite},
				Middlewares: []Middleware{{Name: "RequireScope(secrets:write)", Handler: deps.RequireScope(auth.ScopeSecretsWrite)}},
			},
		},
	}
}
