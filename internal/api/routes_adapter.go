package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/api/routes"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// buildBranchContextMiddleware returns the branch context middleware for the
// branching router, or nil when branching is disabled. Access checks are
// enforced for explicit per-request branch selections (X-Fluxbase-Branch
// header / ?branch= query param) made by non-admin principals.
func (s *Server) buildBranchContextMiddleware() fiber.Handler {
	if s.Branching == nil || s.Branching.Router == nil {
		return nil
	}
	return middleware.BranchContext(middleware.BranchContextConfig{
		Router:         s.Branching.Router,
		RequireAccess:  true,
		AllowAnonymous: false,
	})
}

func (s *Server) registerRoutesViaRegistry() error {
	deps := &routes.AllDeps{
		Health:             s.buildHealthRouteDeps(),
		Realtime:           s.buildRealtimeRouteDeps(),
		Storage:            s.buildStorageRouteDeps(),
		REST:               s.buildRESTRouteDeps(),
		GraphQL:            s.buildGraphQLRouteDeps(),
		Vector:             s.buildVectorRouteDeps(),
		RPC:                s.buildRPCRouteDeps(),
		AI:                 s.buildAIRouteDeps(),
		Settings:           s.buildSettingsRouteDeps(),
		UserSettings:       s.buildUserSettingsRouteDeps(),
		Dashboard:          s.buildDashboardAuthRouteDeps(),
		OpenAPI:            s.buildOpenAPIRouteDeps(),
		Auth:               s.buildAuthRouteDeps(),
		InternalAI:         s.buildInternalAIRouteDeps(),
		GitHubWebhook:      s.buildGitHubWebhookRouteDeps(),
		Invitation:         s.buildInvitationRouteDeps(),
		Webhook:            s.buildWebhookRouteDeps(),
		Monitoring:         s.buildMonitoringRouteDeps(),
		Functions:          s.buildFunctionsRouteDeps(),
		Jobs:               s.buildJobsRouteDeps(),
		ClientKeys:         s.buildClientKeysRouteDeps(),
		Secrets:            s.buildSecretsRouteDeps(),
		Sync:               s.buildSyncRouteDeps(),
		Admin:              s.buildAdminRouteDeps(),
		DashboardUserAuth:  s.buildDashboardUserAuthRouteDeps(),
		CustomMCP:          s.buildCustomMCPRouteDeps(),
		MCP:                s.buildMCPRouteDeps(),
		MCPOAuth:           s.buildMCPOAuthRouteDeps(),
		Migrations:         s.buildMigrationsRouteDeps(),
		KnowledgeBase:      s.buildKnowledgeBaseRouteDeps(),
		Root:               s.handleHealth,
		EnsureTenantAccess: s.Middleware.EnsureTenantAccess,
		BranchContext:      s.buildBranchContextMiddleware(),
	}

	return routes.RegisterAllRoutes(s.app, deps)
}

func (s *Server) auditRegisteredRoutes() []routes.RouteAuditEntry {
	deps := &routes.AllDeps{
		Health:            s.buildHealthRouteDeps(),
		Realtime:          s.buildRealtimeRouteDeps(),
		Storage:           s.buildStorageRouteDeps(),
		REST:              s.buildRESTRouteDeps(),
		GraphQL:           s.buildGraphQLRouteDeps(),
		Vector:            s.buildVectorRouteDeps(),
		RPC:               s.buildRPCRouteDeps(),
		AI:                s.buildAIRouteDeps(),
		Settings:          s.buildSettingsRouteDeps(),
		UserSettings:      s.buildUserSettingsRouteDeps(),
		Dashboard:         s.buildDashboardAuthRouteDeps(),
		OpenAPI:           s.buildOpenAPIRouteDeps(),
		Auth:              s.buildAuthRouteDeps(),
		InternalAI:        s.buildInternalAIRouteDeps(),
		GitHubWebhook:     s.buildGitHubWebhookRouteDeps(),
		Invitation:        s.buildInvitationRouteDeps(),
		Webhook:           s.buildWebhookRouteDeps(),
		Monitoring:        s.buildMonitoringRouteDeps(),
		Functions:         s.buildFunctionsRouteDeps(),
		Jobs:              s.buildJobsRouteDeps(),
		ClientKeys:        s.buildClientKeysRouteDeps(),
		Secrets:           s.buildSecretsRouteDeps(),
		Sync:              s.buildSyncRouteDeps(),
		Admin:             s.buildAdminRouteDeps(),
		DashboardUserAuth: s.buildDashboardUserAuthRouteDeps(),
		CustomMCP:         s.buildCustomMCPRouteDeps(),
		MCP:               s.buildMCPRouteDeps(),
		MCPOAuth:          s.buildMCPOAuthRouteDeps(),
		Migrations:        s.buildMigrationsRouteDeps(),
		KnowledgeBase:     s.buildKnowledgeBaseRouteDeps(),
		Root:              s.handleHealth,
	}

	return routes.AuditRoutes(deps)
}
