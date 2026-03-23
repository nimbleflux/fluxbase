package api

import (
	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/api/routes"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

func (s *Server) buildHealthRouteDeps() *routes.HealthDeps {
	return &routes.HealthDeps{
		Handler: s.handleHealth,
	}
}

func (s *Server) buildRealtimeRouteDeps() *routes.RealtimeDeps {
	return &routes.RealtimeDeps{
		RequireRealtimeEnabled: middleware.RequireRealtimeEnabled(s.authHandler.authService.GetSettingsCache()),
		OptionalAuth:           middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireAuth:            middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:           middleware.RequireScope,
		HandleWebSocket:        s.realtimeHandler.HandleWebSocket,
		HandleStats:            s.handleRealtimeStats,
		HandleBroadcast:        s.handleRealtimeBroadcast,
	}
}

func (s *Server) buildStorageRouteDeps() *routes.StorageDeps {
	return &routes.StorageDeps{
		RequireAuth:            middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		OptionalAuth:           middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:           middleware.RequireScope,
		DownloadSignedObject:   s.storageHandler.DownloadSignedObject,
		GetTransformConfig:     s.storageHandler.GetTransformConfig,
		ListBuckets:            s.storageHandler.ListBuckets,
		CreateBucket:           s.storageHandler.CreateBucket,
		UpdateBucketSettings:   s.storageHandler.UpdateBucketSettings,
		DeleteBucket:           s.storageHandler.DeleteBucket,
		ListFiles:              s.storageHandler.ListFiles,
		MultipartUpload:        s.storageHandler.MultipartUpload,
		ShareObject:            s.storageHandler.ShareObject,
		RevokeShare:            s.storageHandler.RevokeShare,
		ListShares:             s.storageHandler.ListShares,
		GenerateSignedURL:      s.storageHandler.GenerateSignedURL,
		StreamUpload:           s.storageHandler.StreamUpload,
		StorageUploadLimiter:   middleware.StorageUploadLimiter(s.sharedMiddlewareStorage),
		InitChunkedUpload:      s.storageHandler.InitChunkedUpload,
		UploadChunk:            s.storageHandler.UploadChunk,
		CompleteChunkedUpload:  s.storageHandler.CompleteChunkedUpload,
		GetChunkedUploadStatus: s.storageHandler.GetChunkedUploadStatus,
		AbortChunkedUpload:     s.storageHandler.AbortChunkedUpload,
		UploadFile:             s.storageHandler.UploadFile,
		DownloadFile:           s.storageHandler.DownloadFile,
		DeleteFile:             s.storageHandler.DeleteFile,
	}
}

func (s *Server) buildRESTRouteDeps() *routes.RESTDeps {
	return &routes.RESTDeps{
		RequireAuth:  middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.DB(), s.dashboardAuthHandler.jwtManager),
		RequireScope: middleware.RequireScope,
		HandleTables: s.rest.HandleDynamicTable,
		HandleQuery:  s.rest.HandleDynamicQuery,
		HandleById:   s.rest.HandleDynamicTableById,
	}
}

func (s *Server) buildGraphQLRouteDeps() *routes.GraphQLDeps {
	if s.graphqlHandler == nil {
		return nil
	}
	return &routes.GraphQLDeps{
		OptionalAuth:     middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.DB(), s.dashboardAuthHandler.jwtManager),
		HandleGraphQL:    s.graphqlHandler.HandleGraphQL,
		HandleIntrospect: s.graphqlHandler.HandleIntrospection,
	}
}

func (s *Server) buildVectorRouteDeps() *routes.VectorDeps {
	if s.vectorHandler == nil {
		return nil
	}
	return &routes.VectorDeps{
		RequireAuth:        middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		HandleCapabilities: s.vectorHandler.HandleGetCapabilities,
		HandleEmbed:        s.vectorHandler.HandleEmbed,
		HandleSearch:       s.vectorHandler.HandleSearch,
	}
}

func (s *Server) buildRPCRouteDeps() *routes.RPCDeps {
	if s.rpcHandler == nil {
		return nil
	}
	return &routes.RPCDeps{
		RequireRPCEnabled: middleware.RequireRPCEnabled(s.authHandler.authService.GetSettingsCache()),
		OptionalAuth:      middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:      middleware.RequireScope,
		ListProcedures:    s.rpcHandler.ListPublicProcedures,
		Invoke:            s.rpcHandler.Invoke,
		GetExecution:      s.rpcHandler.GetPublicExecution,
		GetExecutionLogs:  s.rpcHandler.GetPublicExecutionLogs,
	}
}

func (s *Server) buildAIRouteDeps() *routes.AIDeps {
	if s.aiChatHandler == nil || s.aiHandler == nil {
		return nil
	}
	return &routes.AIDeps{
		RequireAIEnabled:       middleware.RequireAIEnabled(s.authHandler.authService.GetSettingsCache()),
		OptionalAuth:           middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireAuth:            middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		HandleWebSocket:        s.aiChatHandler.HandleWebSocket,
		ListPublicChatbots:     s.aiHandler.ListPublicChatbots,
		LookupChatbotByName:    s.aiHandler.LookupChatbotByName,
		GetPublicChatbot:       s.aiHandler.GetPublicChatbot,
		ListUserConversations:  s.aiHandler.ListUserConversations,
		GetUserConversation:    s.aiHandler.GetUserConversation,
		DeleteUserConversation: s.aiHandler.DeleteUserConversation,
		UpdateUserConversation: s.aiHandler.UpdateUserConversation,
	}
}

func (s *Server) buildSettingsRouteDeps() *routes.SettingsDeps {
	return &routes.SettingsDeps{
		OptionalAuth: middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool()),
		GetSetting:   s.settingsHandler.GetSetting,
		GetSettings:  s.settingsHandler.GetSettings,
	}
}

func (s *Server) buildUserSettingsRouteDeps() *routes.UserSettingsDeps {
	return &routes.UserSettingsDeps{
		RequireAuth:       middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		ListSettings:      s.userSettingsHandler.ListSettings,
		GetUserOwnSetting: s.userSettingsHandler.GetUserOwnSetting,
		GetSystemSetting:  s.userSettingsHandler.GetSystemSettingPublic,
		GetSetting:        s.userSettingsHandler.GetSetting,
		SetSetting:        s.userSettingsHandler.SetSetting,
		DeleteSetting:     s.userSettingsHandler.DeleteSetting,
		CreateSecret:      s.userSettingsHandler.CreateSecret,
		ListSecrets:       s.userSettingsHandler.ListSecrets,
		GetSecret:         s.userSettingsHandler.GetSecret,
		UpdateSecret:      s.userSettingsHandler.UpdateSecret,
		DeleteSecret:      s.userSettingsHandler.DeleteSecret,
	}
}

func (s *Server) buildDashboardAuthRouteDeps() *routes.DashboardAuthDeps {
	return &routes.DashboardAuthDeps{
		SetupLimiter:    middleware.AdminSetupLimiterWithConfig(s.config.Security.AdminSetupRateLimit, s.config.Security.AdminSetupRateWindow, s.sharedMiddlewareStorage),
		LoginLimiter:    middleware.AdminLoginLimiterWithConfig(s.config.Security.AdminLoginRateLimit, s.config.Security.AdminLoginRateWindow, s.sharedMiddlewareStorage),
		GetSetupStatus:  s.adminAuthHandler.GetSetupStatus,
		InitialSetup:    s.adminAuthHandler.InitialSetup,
		AdminLogin:      s.adminAuthHandler.AdminLogin,
		RefreshToken:    s.adminAuthHandler.AdminRefreshToken,
		UnifiedAuth:     UnifiedAuthMiddleware(s.authHandler.authService, s.dashboardAuthHandler.jwtManager, s.db.Pool()),
		AdminLogout:     s.adminAuthHandler.AdminLogout,
		GetCurrentAdmin: s.adminAuthHandler.GetCurrentAdmin,
	}
}

func (s *Server) buildOpenAPIRouteDeps() *routes.OpenAPIDeps {
	return &routes.OpenAPIDeps{
		OptionalAuth:   middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		GetOpenAPISpec: NewOpenAPIHandler(s.db).GetOpenAPISpec,
	}
}

func (s *Server) buildAuthRouteDeps() *routes.AuthDeps {
	rateLimiters := map[string]fiber.Handler{
		"signup":         middleware.AuthSignupLimiterWithConfig(s.config.Security.AuthSignupRateLimit, s.config.Security.AuthSignupRateWindow, s.sharedMiddlewareStorage),
		"login":          middleware.AuthLoginLimiterWithConfig(s.config.Security.AuthLoginRateLimit, s.config.Security.AuthLoginRateWindow, s.sharedMiddlewareStorage),
		"refresh":        middleware.AuthRefreshLimiterWithConfig(s.config.Security.AuthRefreshRateLimit, s.config.Security.AuthRefreshRateWindow, s.sharedMiddlewareStorage),
		"magiclink":      middleware.AuthMagicLinkLimiterWithConfig(s.config.Security.AuthMagicLinkRateLimit, s.config.Security.AuthMagicLinkRateWindow, s.sharedMiddlewareStorage),
		"password_reset": middleware.AuthPasswordResetLimiterWithConfig(s.config.Security.AuthPasswordResetRateLimit, s.config.Security.AuthPasswordResetRateWindow, s.sharedMiddlewareStorage),
		"otp":            middleware.AuthMagicLinkLimiterWithConfig(s.config.Security.AuthMagicLinkRateLimit, s.config.Security.AuthMagicLinkRateWindow, s.sharedMiddlewareStorage),
		"2fa":            middleware.Auth2FALimiterWithConfig(s.config.Security.Auth2FARateLimit, s.config.Security.Auth2FARateWindow, s.sharedMiddlewareStorage),
	}

	return &routes.AuthDeps{
		AuthMiddleware:            AuthMiddleware(s.authHandler.authService),
		RequireScope:              middleware.RequireScope,
		RateLimiters:              rateLimiters,
		GetCSRFToken:              s.authHandler.GetCSRFToken,
		GetCaptchaConfig:          s.authHandler.GetCaptchaConfig,
		CheckCaptcha:              s.authHandler.CheckCaptcha,
		GetAuthConfig:             s.authHandler.GetAuthConfig,
		SignUp:                    s.authHandler.SignUp,
		SignIn:                    s.authHandler.SignIn,
		RefreshToken:              s.authHandler.RefreshToken,
		SendMagicLink:             s.authHandler.SendMagicLink,
		VerifyMagicLink:           s.authHandler.VerifyMagicLink,
		RequestPasswordReset:      s.authHandler.RequestPasswordReset,
		ResetPassword:             s.authHandler.ResetPassword,
		VerifyPasswordReset:       s.authHandler.VerifyPasswordResetToken,
		VerifyEmail:               s.authHandler.VerifyEmail,
		ResendVerification:        s.authHandler.ResendVerificationEmail,
		VerifyTOTP:                s.authHandler.VerifyTOTP,
		SendOTP:                   s.authHandler.SendOTP,
		VerifyOTP:                 s.authHandler.VerifyOTP,
		ResendOTP:                 s.authHandler.ResendOTP,
		SignInWithIDToken:         s.authHandler.SignInWithIDToken,
		SignOut:                   s.authHandler.SignOut,
		GetUser:                   s.authHandler.GetUser,
		UpdateUser:                s.authHandler.UpdateUser,
		StartImpersonation:        s.authHandler.StartImpersonation,
		StartAnonImpersonation:    s.authHandler.StartAnonImpersonation,
		StopImpersonation:         s.authHandler.StopImpersonation,
		GetActiveImpersonation:    s.authHandler.GetActiveImpersonation,
		ListImpersonationSessions: s.authHandler.ListImpersonationSessions,
		SetupTOTP:                 s.authHandler.SetupTOTP,
		EnableTOTP:                s.authHandler.EnableTOTP,
		DisableTOTP:               s.authHandler.DisableTOTP,
		GetTOTPStatus:             s.authHandler.GetTOTPStatus,
		GetUserIdentities:         s.authHandler.GetUserIdentities,
		LinkIdentity:              s.authHandler.LinkIdentity,
		UnlinkIdentity:            s.authHandler.UnlinkIdentity,
		Reauthenticate:            s.authHandler.Reauthenticate,
		ListOAuthProviders:        s.oauthHandler.ListEnabledProviders,
		OAuthAuthorize:            s.oauthHandler.Authorize,
		OAuthCallback:             s.oauthHandler.Callback,
	}
}

func (s *Server) buildInternalAIRouteDeps() *routes.InternalAIDeps {
	if s.internalAIHandler == nil {
		return nil
	}
	return &routes.InternalAIDeps{
		RequireInternal:     middleware.RequireInternal(),
		RequireAuth:         middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.DB(), s.dashboardAuthHandler.jwtManager),
		HandleChat:          s.internalAIHandler.HandleChat,
		HandleEmbed:         s.internalAIHandler.HandleEmbed,
		HandleListProviders: s.internalAIHandler.HandleListProviders,
	}
}

func (s *Server) buildGitHubWebhookRouteDeps() *routes.GitHubWebhookDeps {
	if s.githubWebhook == nil {
		return nil
	}
	return &routes.GitHubWebhookDeps{
		GitHubWebhookLimiter: middleware.GitHubWebhookLimiter(s.sharedMiddlewareStorage),
		HandleWebhook:        s.githubWebhook.HandleWebhook,
	}
}

func (s *Server) buildInvitationRouteDeps() *routes.InvitationDeps {
	return &routes.InvitationDeps{
		ValidateInvitation: s.invitationHandler.ValidateInvitation,
		AcceptInvitation:   s.invitationHandler.AcceptInvitation,
	}
}

func (s *Server) buildWebhookRouteDeps() *routes.WebhookDeps {
	return &routes.WebhookDeps{
		RequireAuth:    middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:   middleware.RequireScope,
		ListWebhooks:   s.webhookHandler.ListWebhooks,
		GetWebhook:     s.webhookHandler.GetWebhook,
		ListDeliveries: s.webhookHandler.ListDeliveries,
		CreateWebhook:  s.webhookHandler.CreateWebhook,
		UpdateWebhook:  s.webhookHandler.UpdateWebhook,
		DeleteWebhook:  s.webhookHandler.DeleteWebhook,
		TestWebhook:    s.webhookHandler.TestWebhook,
	}
}

func (s *Server) buildMonitoringRouteDeps() *routes.MonitoringDeps {
	return &routes.MonitoringDeps{
		RequireAuth:  middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope: middleware.RequireScope,
		GetMetrics:   s.monitoringHandler.GetMetrics,
		GetHealth:    s.monitoringHandler.GetHealth,
		GetLogs:      s.monitoringHandler.GetLogs,
	}
}

func (s *Server) buildFunctionsRouteDeps() *routes.FunctionsDeps {
	if s.functionsHandler == nil {
		return nil
	}
	return &routes.FunctionsDeps{
		RequireFunctionsEnabled: middleware.RequireFunctionsEnabled(s.authHandler.authService.GetSettingsCache()),
		RequireAuth:             middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		OptionalAuth:            middleware.OptionalAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:            middleware.RequireScope,
		ListFunctions:           s.functionsHandler.ListFunctions,
		GetFunction:             s.functionsHandler.GetFunction,
		CreateFunction:          s.functionsHandler.CreateFunction,
		UpdateFunction:          s.functionsHandler.UpdateFunction,
		DeleteFunction:          s.functionsHandler.DeleteFunction,
		InvokeFunction:          s.functionsHandler.InvokeFunction,
		GetExecutions:           s.functionsHandler.GetExecutions,
		ListSharedModules:       s.functionsHandler.ListSharedModules,
		GetSharedModule:         s.functionsHandler.GetSharedModule,
		CreateSharedModule:      s.functionsHandler.CreateSharedModule,
		UpdateSharedModule:      s.functionsHandler.UpdateSharedModule,
		DeleteSharedModule:      s.functionsHandler.DeleteSharedModule,
	}
}

func (s *Server) buildJobsRouteDeps() *routes.JobsDeps {
	if s.jobsHandler == nil {
		return nil
	}
	return &routes.JobsDeps{
		RequireJobsEnabled: middleware.RequireJobsEnabled(s.authHandler.authService.GetSettingsCache()),
		RequireAuth:        middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		SubmitJob:          s.jobsHandler.SubmitJob,
		GetJob:             s.jobsHandler.GetJob,
		ListJobs:           s.jobsHandler.ListJobs,
		CancelJob:          s.jobsHandler.CancelJob,
		RetryJob:           s.jobsHandler.RetryJob,
		GetJobLogsUser:     s.jobsHandler.GetJobLogsUser,
	}
}

func (s *Server) buildBranchRouteDeps() *routes.BranchDeps {
	if s.branchHandler == nil || !s.config.Branching.Enabled {
		return nil
	}
	return &routes.BranchDeps{
		RequireAuth:        middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireRole:        RequireRole("admin", "instance_admin", "service_role"),
		GetActiveBranch:    s.branchHandler.GetActiveBranch,
		SetActiveBranch:    s.branchHandler.SetActiveBranch,
		ResetActiveBranch:  s.branchHandler.ResetActiveBranch,
		GetPoolStats:       s.branchHandler.GetPoolStats,
		CreateBranch:       s.branchHandler.CreateBranch,
		ListBranches:       s.branchHandler.ListBranches,
		GetBranch:          s.branchHandler.GetBranch,
		DeleteBranch:       s.branchHandler.DeleteBranch,
		ResetBranch:        s.branchHandler.ResetBranch,
		GetBranchActivity:  s.branchHandler.GetBranchActivity,
		ListBranchAccess:   s.branchHandler.ListBranchAccess,
		GrantBranchAccess:  s.branchHandler.GrantBranchAccess,
		RevokeBranchAccess: s.branchHandler.RevokeBranchAccess,
		ListGitHubConfigs:  s.branchHandler.ListGitHubConfigs,
		UpsertGitHubConfig: s.branchHandler.UpsertGitHubConfig,
		DeleteGitHubConfig: s.branchHandler.DeleteGitHubConfig,
	}
}

func (s *Server) buildClientKeysRouteDeps() *routes.ClientKeysDeps {
	return &routes.ClientKeysDeps{
		RequireAuth:                      middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireAdminIfClientKeysDisabled: middleware.RequireAdminIfClientKeysDisabled(s.authHandler.authService.GetSettingsCache()),
		RequireScope:                     middleware.RequireScope,
		ListClientKeys:                   s.clientKeyHandler.ListClientKeys,
		GetClientKey:                     s.clientKeyHandler.GetClientKey,
		CreateClientKey:                  s.clientKeyHandler.CreateClientKey,
		UpdateClientKey:                  s.clientKeyHandler.UpdateClientKey,
		DeleteClientKey:                  s.clientKeyHandler.DeleteClientKey,
		RevokeClientKey:                  s.clientKeyHandler.RevokeClientKey,
	}
}

func (s *Server) buildSecretsRouteDeps() *routes.SecretsDeps {
	if s.secretsHandler == nil {
		return nil
	}
	return &routes.SecretsDeps{
		RequireAuth:        middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireScope:       middleware.RequireScope,
		ListSecrets:        s.secretsHandler.ListSecrets,
		GetStats:           s.secretsHandler.GetStats,
		GetSecretByName:    s.secretsHandler.GetSecretByName,
		GetVersionsByName:  s.secretsHandler.GetVersionsByName,
		UpdateSecretByName: s.secretsHandler.UpdateSecretByName,
		DeleteSecretByName: s.secretsHandler.DeleteSecretByName,
		RollbackByName:     s.secretsHandler.RollbackByName,
		GetSecret:          s.secretsHandler.GetSecret,
		GetVersions:        s.secretsHandler.GetVersions,
		CreateSecret:       s.secretsHandler.CreateSecret,
		UpdateSecret:       s.secretsHandler.UpdateSecret,
		DeleteSecret:       s.secretsHandler.DeleteSecret,
		RollbackToVersion:  s.secretsHandler.RollbackToVersion,
	}
}

func (s *Server) buildSyncRouteDeps() *routes.SyncDeps {
	deps := &routes.SyncDeps{
		RequireSyncAuth: UnifiedAuthMiddleware(s.authHandler.authService, s.dashboardAuthHandler.jwtManager, s.db.Pool()),
		RequireRole:     RequireRole("admin", "instance_admin", "service_role"),
	}

	// Functions sync
	if s.functionsHandler != nil {
		deps.RequireFunctionsSyncIPAllowlist = middleware.RequireSyncIPAllowlist(s.config.Functions.SyncAllowedIPRanges, "functions", &s.config.Server)
		deps.SyncFunctions = s.functionsHandler.SyncFunctions
	}

	// Jobs sync
	if s.jobsHandler != nil {
		deps.RequireJobsSyncIPAllowlist = middleware.RequireSyncIPAllowlist(s.config.Jobs.SyncAllowedIPRanges, "jobs", &s.config.Server)
		deps.SyncJobs = s.jobsHandler.SyncJobs
	}

	// AI sync
	if s.aiHandler != nil {
		deps.RequireAIEnabled = middleware.RequireAIEnabled(s.authHandler.authService.GetSettingsCache())
		deps.RequireAISyncIPAllowlist = middleware.RequireSyncIPAllowlist(s.config.AI.SyncAllowedIPRanges, "ai", &s.config.Server)
		deps.SyncChatbots = s.aiHandler.SyncChatbots
	}

	// RPC sync
	if s.rpcHandler != nil {
		deps.RequireRPCEnabled = middleware.RequireRPCEnabled(s.authHandler.authService.GetSettingsCache())
		deps.RequireRPCSyncIPAllowlist = middleware.RequireSyncIPAllowlist(s.config.RPC.SyncAllowedIPRanges, "rpc", &s.config.Server)
		deps.SyncProcedures = s.rpcHandler.SyncProcedures
	}

	return deps
}

func (s *Server) buildDashboardUserAuthRouteDeps() *routes.DashboardUserAuthDeps {
	return &routes.DashboardUserAuthDeps{
		RequireDashboardAuth:     s.dashboardAuthHandler.RequireDashboardAuth,
		Signup:                   s.dashboardAuthHandler.Signup,
		Login:                    s.dashboardAuthHandler.Login,
		RefreshToken:             s.dashboardAuthHandler.RefreshToken,
		VerifyTOTP:               s.dashboardAuthHandler.VerifyTOTP,
		RequestPasswordReset:     s.dashboardAuthHandler.RequestPasswordReset,
		VerifyPasswordResetToken: s.dashboardAuthHandler.VerifyPasswordResetToken,
		ConfirmPasswordReset:     s.dashboardAuthHandler.ConfirmPasswordReset,
		GetSSOProviders:          s.dashboardAuthHandler.GetSSOProviders,
		InitiateOAuthLogin:       s.dashboardAuthHandler.InitiateOAuthLogin,
		OAuthCallback:            s.dashboardAuthHandler.OAuthCallback,
		InitiateSAMLLogin:        s.dashboardAuthHandler.InitiateSAMLLogin,
		SAMLACSCallback:          s.dashboardAuthHandler.SAMLACSCallback,
		GetCurrentUser:           s.dashboardAuthHandler.GetCurrentUser,
		UpdateProfile:            s.dashboardAuthHandler.UpdateProfile,
		ChangePassword:           s.dashboardAuthHandler.ChangePassword,
		DeleteAccount:            s.dashboardAuthHandler.DeleteAccount,
		SetupTOTP:                s.dashboardAuthHandler.SetupTOTP,
		EnableTOTP:               s.dashboardAuthHandler.EnableTOTP,
		DisableTOTP:              s.dashboardAuthHandler.DisableTOTP,
	}
}

func (s *Server) buildCustomMCPRouteDeps() *routes.CustomMCPDeps {
	if s.customMCPHandler == nil {
		return nil
	}
	return &routes.CustomMCPDeps{
		RequireAuth:    middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		RequireAdmin:   middleware.RequireAdmin(),
		GetConfig:      s.customMCPHandler.GetConfig,
		ListTools:      s.customMCPHandler.ListTools,
		CreateTool:     s.customMCPHandler.CreateTool,
		SyncTool:       s.customMCPHandler.SyncTool,
		GetTool:        s.customMCPHandler.GetTool,
		UpdateTool:     s.customMCPHandler.UpdateTool,
		DeleteTool:     s.customMCPHandler.DeleteTool,
		TestTool:       s.customMCPHandler.TestTool,
		ListResources:  s.customMCPHandler.ListResources,
		CreateResource: s.customMCPHandler.CreateResource,
		SyncResource:   s.customMCPHandler.SyncResource,
		GetResource:    s.customMCPHandler.GetResource,
		UpdateResource: s.customMCPHandler.UpdateResource,
		DeleteResource: s.customMCPHandler.DeleteResource,
		TestResource:   s.customMCPHandler.TestResource,
	}
}

func (s *Server) buildMCPRouteDeps() *routes.MCPDeps {
	if s.mcpHandler == nil {
		return nil
	}
	return &routes.MCPDeps{
		BasePath:     s.config.MCP.BasePath,
		MCPAuth:      s.createMCPAuthMiddleware(),
		HandlePost:   s.mcpHandler.HandlePost,
		HandleGet:    s.mcpHandler.HandleGet,
		HandleHealth: s.mcpHandler.HandleHealth,
	}
}

func (s *Server) buildMCPOAuthRouteDeps() *routes.MCPOAuthDeps {
	if s.mcpOAuthHandler == nil {
		return nil
	}
	return &routes.MCPOAuthDeps{
		BasePath:                          s.config.MCP.BasePath,
		HandleAuthorizationServerMetadata: s.mcpOAuthHandler.HandleAuthorizationServerMetadata,
		HandleProtectedResourceMetadata:   s.mcpOAuthHandler.HandleProtectedResourceMetadata,
		HandleClientRegistration:          s.mcpOAuthHandler.HandleClientRegistration,
		HandleAuthorize:                   s.mcpOAuthHandler.HandleAuthorize,
		HandleAuthorizeConsent:            s.mcpOAuthHandler.HandleAuthorizeConsent,
		HandleToken:                       s.mcpOAuthHandler.HandleToken,
		HandleRevoke:                      s.mcpOAuthHandler.HandleRevoke,
	}
}

func (s *Server) buildMigrationsRouteDeps() *routes.MigrationsDeps {
	if s.migrationsHandler == nil || !s.config.Migrations.Enabled {
		return nil
	}

	var tenantPoolProvider middleware.MigrationsTenantPoolProvider
	if s.tenantManager != nil && s.tenantManager.GetRouter() != nil {
		tenantPoolProvider = s.tenantManager.GetRouter()
	}

	return &routes.MigrationsDeps{
		SecurityMiddleware: middleware.RequireMigrationsFullSecurityWithTenantProvider(
			&s.config.Migrations,
			&s.config.Server,
			s.db.Pool(),
			s.authHandler.authService,
			s.config.Security.ServiceRoleRateLimit,
			s.config.Security.ServiceRoleRateWindow,
			s.sharedMiddlewareStorage,
			tenantPoolProvider,
		),
		RequireRole:       RequireRole,
		CreateMigration:   s.migrationsHandler.CreateMigration,
		ListMigrations:    s.migrationsHandler.ListMigrations,
		GetMigration:      s.migrationsHandler.GetMigration,
		UpdateMigration:   s.migrationsHandler.UpdateMigration,
		DeleteMigration:   s.migrationsHandler.DeleteMigration,
		ApplyMigration:    s.migrationsHandler.ApplyMigration,
		RollbackMigration: s.migrationsHandler.RollbackMigration,
		ApplyPending:      s.migrationsHandler.ApplyPending,
		SyncMigrations:    s.migrationsHandler.SyncMigrations,
		GetExecutions:     s.migrationsHandler.GetExecutions,
	}
}

func (s *Server) buildKnowledgeBaseRouteDeps() *routes.KnowledgeBaseDeps {
	if s.kbStorage == nil {
		return nil
	}

	handler := ai.NewUserKnowledgeBaseHandler(s.kbStorage)
	if s.docProcessor != nil {
		handler = ai.NewUserKnowledgeBaseHandlerWithProcessor(s.kbStorage, s.docProcessor)
	}

	deps := &routes.KnowledgeBaseDeps{
		RequireAIEnabled: middleware.RequireAIEnabled(s.authHandler.authService.GetSettingsCache()),
		RequireAuth:      middleware.RequireAuthOrServiceKey(s.authHandler.authService, s.clientKeyService, s.db.Pool(), s.dashboardAuthHandler.jwtManager),
		ListKBs:          handler.ListMyKnowledgeBases,
		CreateKB:         handler.CreateMyKnowledgeBase,
		GetKB:            handler.GetMyKnowledgeBase,
		ShareKB:          handler.ShareKnowledgeBase,
		ListPermissions:  handler.ListPermissions,
		RevokePermission: handler.RevokePermission,
	}

	if s.docProcessor != nil {
		deps.ListDocuments = handler.ListMyDocuments
		deps.GetDocument = handler.GetMyDocument
		deps.AddDocument = handler.AddMyDocument
		deps.UploadDocument = handler.UploadMyDocument
		deps.DeleteDocument = handler.DeleteMyDocument
		deps.SearchKB = handler.SearchMyKB
	}

	return deps
}

func (s *Server) buildAdminRouteDeps() *routes.AdminDeps {
	unifiedAuth := UnifiedAuthMiddleware(s.authHandler.authService, s.dashboardAuthHandler.jwtManager, s.db.Pool())
	return &routes.AdminDeps{
		UnifiedAuth:            unifiedAuth,
		RequireRole:            RequireRole,
		GetTables:              s.handleGetTables,
		GetTableSchema:         s.handleGetTableSchema,
		GetSchemas:             s.handleGetSchemas,
		ExecuteQuery:           s.handleExecuteQuery,
		ListSchemasDDL:         s.ddlHandler.ListSchemas,
		CreateSchemaDDL:        s.ddlHandler.CreateSchema,
		ListTablesDDL:          s.ddlHandler.ListTables,
		CreateTableDDL:         s.ddlHandler.CreateTable,
		DeleteTableDDL:         s.ddlHandler.DeleteTable,
		RenameTableDDL:         s.ddlHandler.RenameTable,
		AddColumnDDL:           s.ddlHandler.AddColumn,
		DropColumnDDL:          s.ddlHandler.DropColumn,
		EnableRealtime:         s.realtimeAdminHandler.HandleEnableRealtime,
		ListRealtimeTables:     s.realtimeAdminHandler.HandleListRealtimeTables,
		GetRealtimeStatus:      s.realtimeAdminHandler.HandleGetRealtimeStatus,
		UpdateRealtimeConfig:   s.realtimeAdminHandler.HandleUpdateRealtimeConfig,
		DisableRealtime:        s.realtimeAdminHandler.HandleDisableRealtime,
		ListOAuthProviders:     s.oauthProviderHandler.ListOAuthProviders,
		GetOAuthProvider:       s.oauthProviderHandler.GetOAuthProvider,
		CreateOAuthProvider:    s.oauthProviderHandler.CreateOAuthProvider,
		UpdateOAuthProvider:    s.oauthProviderHandler.UpdateOAuthProvider,
		DeleteOAuthProvider:    s.oauthProviderHandler.DeleteOAuthProvider,
		ListSAMLProviders:      s.samlProviderHandler.ListSAMLProviders,
		GetSAMLProvider:        s.samlProviderHandler.GetSAMLProvider,
		CreateSAMLProvider:     s.samlProviderHandler.CreateSAMLProvider,
		UpdateSAMLProvider:     s.samlProviderHandler.UpdateSAMLProvider,
		DeleteSAMLProvider:     s.samlProviderHandler.DeleteSAMLProvider,
		ValidateSAML:           s.samlProviderHandler.ValidateMetadata,
		UploadSAMLMetadata:     s.samlProviderHandler.UploadMetadata,
		GetAuthSettings:        s.oauthProviderHandler.GetAuthSettings,
		UpdateAuthSettings:     s.oauthProviderHandler.UpdateAuthSettings,
		ListSessions:           s.adminSessionHandler.ListSessions,
		RevokeSession:          s.adminSessionHandler.RevokeSession,
		RevokeUserSessions:     s.adminSessionHandler.RevokeUserSessions,
		ListSystemSettings:     s.systemSettingsHandler.ListSettings,
		GetSystemSetting:       s.systemSettingsHandler.GetSetting,
		UpdateSystemSetting:    s.systemSettingsHandler.UpdateSetting,
		DeleteSystemSetting:    s.systemSettingsHandler.DeleteSetting,
		CreateCustomSetting:    s.customSettingsHandler.CreateSetting,
		ListCustomSettings:     s.customSettingsHandler.ListSettings,
		CreateSecretSetting:    s.customSettingsHandler.CreateSecretSetting,
		ListSecretSettings:     s.customSettingsHandler.ListSecretSettings,
		GetSecretSetting:       s.customSettingsHandler.GetSecretSetting,
		UpdateSecretSetting:    s.customSettingsHandler.UpdateSecretSetting,
		DeleteSecretSetting:    s.customSettingsHandler.DeleteSecretSetting,
		GetUserSecretValue:     s.userSettingsHandler.GetUserSecretValue,
		GetCustomSetting:       s.customSettingsHandler.GetSetting,
		UpdateCustomSetting:    s.customSettingsHandler.UpdateSetting,
		DeleteCustomSetting:    s.customSettingsHandler.DeleteSetting,
		GetAppSettings:         s.appSettingsHandler.GetAppSettings,
		UpdateAppSettings:      s.appSettingsHandler.UpdateAppSettings,
		ListEmailSettings:      s.emailSettingsHandler.GetSettings,
		GetEmailSetting:        s.emailSettingsHandler.GetSettings,
		UpdateEmailSetting:     s.emailSettingsHandler.UpdateSettings,
		TestEmailSettings:      s.emailSettingsHandler.TestSettings,
		ListEmailTemplates:     s.emailTemplateHandler.ListTemplates,
		GetEmailTemplate:       s.emailTemplateHandler.GetTemplate,
		UpdateEmailTemplate:    s.emailTemplateHandler.UpdateTemplate,
		TestEmailTemplate:      s.emailTemplateHandler.TestTemplate,
		GetCaptchaSettings:     s.captchaSettingsHandler.GetSettings,
		UpdateCaptchaSettings:  s.captchaSettingsHandler.UpdateSettings,
		ListUsers:              s.userManagementHandler.ListUsers,
		InviteUser:             s.userManagementHandler.InviteUser,
		DeleteUser:             s.userManagementHandler.DeleteUser,
		UpdateUser:             s.userManagementHandler.UpdateUser,
		UpdateUserRole:         s.userManagementHandler.UpdateUserRole,
		ResetUserPassword:      s.userManagementHandler.ResetUserPassword,
		ListUsersWithQuotas:    nil,
		GetUserQuota:           nil,
		SetUserQuota:           nil,
		CreateInvitation:       s.invitationHandler.CreateInvitation,
		ListInvitations:        s.invitationHandler.ListInvitations,
		RevokeInvitation:       s.invitationHandler.RevokeInvitation,
		ListServiceKeys:        s.serviceKeyHandler.ListServiceKeys,
		GetServiceKey:          s.serviceKeyHandler.GetServiceKey,
		CreateServiceKey:       s.serviceKeyHandler.CreateServiceKey,
		UpdateServiceKey:       s.serviceKeyHandler.UpdateServiceKey,
		DeleteServiceKey:       s.serviceKeyHandler.DeleteServiceKey,
		DisableServiceKey:      s.serviceKeyHandler.DisableServiceKey,
		EnableServiceKey:       s.serviceKeyHandler.EnableServiceKey,
		RevokeServiceKey:       s.serviceKeyHandler.RevokeServiceKey,
		DeprecateServiceKey:    s.serviceKeyHandler.DeprecateServiceKey,
		RotateServiceKey:       s.serviceKeyHandler.RotateServiceKey,
		GetRevocationHistory:   s.serviceKeyHandler.GetRevocationHistory,
		ListMyTenants:          s.tenantHandler.ListMyTenants,
		ListTenants:            s.tenantHandler.ListTenants,
		CreateTenant:           s.tenantHandler.CreateTenant,
		GetTenant:              s.tenantHandler.GetTenant,
		UpdateTenant:           s.tenantHandler.UpdateTenant,
		DeleteTenant:           s.tenantHandler.DeleteTenant,
		MigrateTenant:          s.tenantHandler.MigrateTenant,
		ListAdmins:             s.tenantHandler.ListAdmins,
		AssignAdmin:            s.tenantHandler.AssignAdmin,
		RemoveAdmin:            s.tenantHandler.RemoveAdmin,
		ExecuteSQL:             s.sqlHandler.ExecuteSQL,
		ExportTypeScript:       s.schemaExportHandler.HandleExportTypeScript,
		ReloadFunctions:        s.functionsHandler.ReloadFunctions,
		ListFunctionNamespaces: s.functionsHandler.ListNamespaces,
		ListAllExecutions:      s.functionsHandler.ListAllExecutions,
		GetExecutionLogs:       s.functionsHandler.GetExecutionLogs,
		ListJobNamespaces:      s.jobsHandler.ListNamespaces,
		ListJobFunctions:       s.jobsHandler.ListJobFunctions,
		GetJobFunction:         s.jobsHandler.GetJobFunction,
		DeleteJobFunction:      s.jobsHandler.DeleteJobFunction,
		GetJobStats:            s.jobsHandler.GetJobStats,
		ListWorkers:            s.jobsHandler.ListWorkers,
		ListAllJobs:            s.jobsHandler.ListAllJobs,
		GetJobAdmin:            s.jobsHandler.GetJobAdmin,
		TerminateJob:           s.jobsHandler.TerminateJob,
		CancelJobAdmin:         s.jobsHandler.CancelJobAdmin,
		RetryJobAdmin:          s.jobsHandler.RetryJobAdmin,
		ResubmitJobAdmin:       s.jobsHandler.ResubmitJobAdmin,
		ListChatbots:           s.aiHandler.ListChatbots,
		GetChatbot:             s.aiHandler.GetChatbot,
		ToggleChatbot:          s.aiHandler.ToggleChatbot,
		UpdateChatbot:          s.aiHandler.UpdateChatbot,
		DeleteChatbot:          s.aiHandler.DeleteChatbot,
		GetAIMetrics:           s.aiHandler.GetAIMetrics,
		SyncFunctions:          s.functionsHandler.SyncFunctions,
		SyncJobs:               s.jobsHandler.SyncJobs,
		SyncChatbots:           s.aiHandler.SyncChatbots,
		SyncProcedures:         s.rpcHandler.SyncProcedures,
		RefreshSchema:          s.handleRefreshSchema,
		ListLogs:               s.loggingHandler.QueryLogs,
		GetLogStats:            s.loggingHandler.GetLogStats,
		GetExecutionLogsAdmin:  s.loggingHandler.GetExecutionLogs,
		FlushLogs:              s.loggingHandler.FlushLogs,
		GenerateTestLogs:       s.loggingHandler.GenerateTestLogs,

		// Instance Settings
		GetInstanceSettings:       s.instanceSettingsHandler.GetInstanceSettings,
		UpdateInstanceSettings:    s.instanceSettingsHandler.UpdateInstanceSettings,
		GetOverridableSettings:    s.instanceSettingsHandler.GetOverridableSettings,
		UpdateOverridableSettings: s.instanceSettingsHandler.UpdateOverridableSettings,

		// Extensions (instance-level)
		ListExtensions:   s.extensionsHandler.ListExtensions,
		GetExtension:     s.extensionsHandler.GetExtensionStatus,
		EnableExtension:  s.extensionsHandler.EnableExtension,
		DisableExtension: s.extensionsHandler.DisableExtension,
		SyncExtensions:   s.extensionsHandler.SyncExtensions,

		// AI Admin
		ListAIProviders:            s.aiHandler.ListProviders,
		ListAIConversations:        s.aiHandler.GetConversations,
		GetAIConversationMessages:  s.aiHandler.GetConversationMessages,
		GetAIAuditLog:              s.aiHandler.GetAuditLog,
		ListExportableTables:       s.knowledgeBaseHandler.ListExportableTables,
		GetExportableTableDetails:  s.knowledgeBaseHandler.GetTableDetails,
		ExportTableToKnowledgeBase: s.knowledgeBaseHandler.ExportTableToKnowledgeBase,

		// RPC Admin
		ListRPCNamespaces:   s.rpcHandler.ListNamespaces,
		ListProcedures:      s.rpcHandler.ListProcedures,
		GetProcedure:        s.rpcHandler.GetProcedure,
		UpdateProcedure:     s.rpcHandler.UpdateProcedure,
		DeleteProcedure:     s.rpcHandler.DeleteProcedure,
		ListRPCExecutions:   s.rpcHandler.ListExecutions,
		GetRPCExecution:     s.rpcHandler.GetExecution,
		GetRPCExecutionLogs: s.rpcHandler.GetExecutionLogs,
		CancelRPCExecution:  s.rpcHandler.CancelExecution,

		// Schema/RLS/Policy
		GetSchemaGraph:        s.GetSchemaGraph,
		GetTableRelationships: s.GetTableRelationships,
		GetTablesWithRLS:      s.GetTablesWithRLS,
		GetTableRLSStatus:     s.GetTableRLSStatus,
		ToggleTableRLS:        s.ToggleTableRLS,
		ListPolicies:          s.ListPolicies,
		CreatePolicy:          s.CreatePolicy,
		UpdatePolicy:          s.UpdatePolicy,
		DeletePolicy:          s.DeletePolicy,
		GetPolicyTemplates:    s.GetPolicyTemplates,
		GetSecurityWarnings:   s.GetSecurityWarnings,

		// Tenant Settings
		GetTenantSettings:    s.tenantSettingsHandler.GetTenantSettings,
		UpdateTenantSettings: s.tenantSettingsHandler.UpdateTenantSettings,
		DeleteTenantSetting:  s.tenantSettingsHandler.DeleteTenantSetting,
		GetTenantSetting:     s.tenantSettingsHandler.GetTenantSetting,
	}
}

func (s *Server) registerRoutesViaRegistry() error {
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
		Branch:            s.buildBranchRouteDeps(),
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
		Branch:            s.buildBranchRouteDeps(),
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
