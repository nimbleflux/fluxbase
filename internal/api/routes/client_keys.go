package routes

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/auth"
)

type ClientKeysDeps struct {
	RequireAuth                      fiber.Handler
	RequireAdminIfClientKeysDisabled fiber.Handler
	RequireScope                     func(...string) fiber.Handler
	ListClientKeys                   fiber.Handler
	GetClientKey                     fiber.Handler
	CreateClientKey                  fiber.Handler
	UpdateClientKey                  fiber.Handler
	DeleteClientKey                  fiber.Handler
	RevokeClientKey                  fiber.Handler
}

func BuildClientKeysRoutes(deps *ClientKeysDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "client_keys",
		Prefix: "/api/v1/client-keys",
		Middlewares: []Middleware{
			{Name: "RequireAuth", Handler: deps.RequireAuth},
			{Name: "RequireAdminIfClientKeysDisabled", Handler: deps.RequireAdminIfClientKeysDisabled},
		},
		Routes: []Route{
			{
				Method:      "GET",
				Path:        "/",
				Handler:     deps.ListClientKeys,
				Summary:     "List client keys",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysRead},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:read)", Handler: deps.RequireScope(auth.ScopeClientKeysRead)}},
			},
			{
				Method:      "GET",
				Path:        "/:id",
				Handler:     deps.GetClientKey,
				Summary:     "Get a client key",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysRead},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:read)", Handler: deps.RequireScope(auth.ScopeClientKeysRead)}},
			},
			{
				Method:      "POST",
				Path:        "/",
				Handler:     deps.CreateClientKey,
				Summary:     "Create a client key",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysWrite},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:write)", Handler: deps.RequireScope(auth.ScopeClientKeysWrite)}},
			},
			{
				Method:      "PATCH",
				Path:        "/:id",
				Handler:     deps.UpdateClientKey,
				Summary:     "Update a client key",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysWrite},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:write)", Handler: deps.RequireScope(auth.ScopeClientKeysWrite)}},
			},
			{
				Method:      "DELETE",
				Path:        "/:id",
				Handler:     deps.DeleteClientKey,
				Summary:     "Delete a client key",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysWrite},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:write)", Handler: deps.RequireScope(auth.ScopeClientKeysWrite)}},
			},
			{
				Method:      "POST",
				Path:        "/:id/revoke",
				Handler:     deps.RevokeClientKey,
				Summary:     "Revoke a client key",
				Auth:        AuthRequired,
				Scopes:      []string{auth.ScopeClientKeysWrite},
				Middlewares: []Middleware{{Name: "RequireScope(clientkeys:write)", Handler: deps.RequireScope(auth.ScopeClientKeysWrite)}},
			},
		},
	}
}
