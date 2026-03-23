package routes

import "github.com/gofiber/fiber/v3"

type MCPOAuthDeps struct {
	BasePath                          string
	HandleAuthorizationServerMetadata fiber.Handler
	HandleProtectedResourceMetadata   fiber.Handler
	HandleClientRegistration          fiber.Handler
	HandleAuthorize                   fiber.Handler
	HandleAuthorizeConsent            fiber.Handler
	HandleToken                       fiber.Handler
	HandleRevoke                      fiber.Handler
}

func BuildMCPOAuthRoutes(deps *MCPOAuthDeps) *RouteGroup {
	return &RouteGroup{
		Name:   "mcp-oauth",
		Prefix: deps.BasePath,
		Routes: []Route{
			{Method: "GET", Path: "/.well-known/oauth-authorization-server", Handler: deps.HandleAuthorizationServerMetadata, Summary: "OAuth authorization server metadata", Auth: AuthNone, Public: true},
			{Method: "GET", Path: "/.well-known/oauth-protected-resource", Handler: deps.HandleProtectedResourceMetadata, Summary: "OAuth protected resource metadata", Auth: AuthNone, Public: true},
			{Method: "GET", Path: "/.well-known/oauth-protected-resource/mcp", Handler: deps.HandleProtectedResourceMetadata, Summary: "OAuth protected resource metadata for MCP", Auth: AuthNone, Public: true},
		},
		SubGroups: []*RouteGroup{
			{
				Name:   "mcp-oauth-endpoints",
				Prefix: deps.BasePath + "/oauth",
				Routes: []Route{
					{Method: "POST", Path: "/register", Handler: deps.HandleClientRegistration, Summary: "Dynamic client registration", Auth: AuthNone, Public: true},
					{Method: "GET", Path: "/authorize", Handler: deps.HandleAuthorize, Summary: "OAuth authorization", Auth: AuthNone, Public: true},
					{Method: "POST", Path: "/authorize", Handler: deps.HandleAuthorizeConsent, Summary: "OAuth authorization consent", Auth: AuthNone, Public: true},
					{Method: "POST", Path: "/token", Handler: deps.HandleToken, Summary: "OAuth token exchange", Auth: AuthNone, Public: true},
					{Method: "POST", Path: "/revoke", Handler: deps.HandleRevoke, Summary: "OAuth token revocation", Auth: AuthNone, Public: true},
				},
			},
		},
	}
}
