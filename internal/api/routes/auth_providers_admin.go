package routes

import (
	"github.com/gofiber/fiber/v3"
)

// AuthProvidersAdminDeps contains dependencies for auth providers admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all auth provider configurations
//   - tenant_admin: Access to sessions within their tenant (RLS enforced)
type AuthProvidersAdminDeps struct {
	ListOAuthProviders  fiber.Handler
	GetOAuthProvider    fiber.Handler
	CreateOAuthProvider fiber.Handler
	UpdateOAuthProvider fiber.Handler
	DeleteOAuthProvider fiber.Handler
	ListSAMLProviders   fiber.Handler
	GetSAMLProvider     fiber.Handler
	CreateSAMLProvider  fiber.Handler
	UpdateSAMLProvider  fiber.Handler
	DeleteSAMLProvider  fiber.Handler
	ValidateSAML        fiber.Handler
	UploadSAMLMetadata  fiber.Handler
	GetAuthSettings     fiber.Handler
	UpdateAuthSettings  fiber.Handler
	ListSessions        fiber.Handler
	RevokeSession       fiber.Handler
	RevokeUserSessions  fiber.Handler
}

// BuildAuthProvidersAdminRoutes creates the auth providers admin route group.
func BuildAuthProvidersAdminRoutes(deps *AuthProvidersAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "auth_providers_admin",
		Routes: []Route{
			// OAuth Providers - instance admin only
			{Method: "GET", Path: "/oauth/providers", Handler: deps.ListOAuthProviders, Summary: "List OAuth providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/oauth/providers/:id", Handler: deps.GetOAuthProvider, Summary: "Get OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/oauth/providers", Handler: deps.CreateOAuthProvider, Summary: "Create OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/oauth/providers/:id", Handler: deps.UpdateOAuthProvider, Summary: "Update OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "DELETE", Path: "/oauth/providers/:id", Handler: deps.DeleteOAuthProvider, Summary: "Delete OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// SAML Providers - instance admin only
			{Method: "GET", Path: "/saml/providers", Handler: deps.ListSAMLProviders, Summary: "List SAML providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/saml/providers/:id", Handler: deps.GetSAMLProvider, Summary: "Get SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/saml/providers", Handler: deps.CreateSAMLProvider, Summary: "Create SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/saml/providers/:id", Handler: deps.UpdateSAMLProvider, Summary: "Update SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "DELETE", Path: "/saml/providers/:id", Handler: deps.DeleteSAMLProvider, Summary: "Delete SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/saml/validate-metadata", Handler: deps.ValidateSAML, Summary: "Validate SAML metadata", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/saml/upload-metadata", Handler: deps.UploadSAMLMetadata, Summary: "Upload SAML metadata", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Auth Settings - instance admin only
			{Method: "GET", Path: "/auth/settings", Handler: deps.GetAuthSettings, Summary: "Get auth settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "PUT", Path: "/auth/settings", Handler: deps.UpdateAuthSettings, Summary: "Update auth settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Sessions - tenant admin can view/revoke sessions in their tenant
			{Method: "GET", Path: "/auth/sessions", Handler: deps.ListSessions, Summary: "List sessions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/auth/sessions/:id", Handler: deps.RevokeSession, Summary: "Revoke session", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/auth/sessions/user/:user_id", Handler: deps.RevokeUserSessions, Summary: "Revoke user sessions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		},
	}
}
