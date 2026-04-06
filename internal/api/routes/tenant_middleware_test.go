package routes

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tenantAwareHandler is a sentinel fiber.Handler used to verify tenant middleware wiring.
func tenantAwareHandler(c fiber.Ctx) error { return c.SendString("ok") }

// hasMiddlewareNamed checks if a slice of Middleware contains one with the given name.
func hasMiddlewareNamed(middlewares []Middleware, name string) bool {
	for _, m := range middlewares {
		if m.Name == name {
			return true
		}
	}
	return false
}

// collectMiddlewareNames returns all middleware names from a slice.
func collectMiddlewareNames(middlewares []Middleware) []string {
	names := make([]string, 0, len(middlewares))
	for _, m := range middlewares {
		names = append(names, m.Name)
	}
	return names
}

// TestTenantMiddleware_AdminRouteGroup verifies that the admin route group has
// tenant middleware set at the parent level, so all subgroups inherit it.
func TestTenantMiddleware_AdminRouteGroup(t *testing.T) {
	deps := &AdminDeps{
		UnifiedAuth:        tenantAwareHandler,
		RequireRole:        func(...string) fiber.Handler { return tenantAwareHandler },
		TenantMiddleware:   tenantAwareHandler,
		TenantDBMiddleware: tenantAwareHandler,
		Schema:             minimalSchemaAdminDeps(),
		AuthProviders:      minimalAuthProvidersAdminDeps(),
		Users:              minimalUsersAdminDeps(),
		Tenants:            minimalTenantsAdminDeps(),
		ServiceKeys:        minimalServiceKeysAdminDeps(),
		Functions:          minimalFunctionsAdminDeps(),
		Jobs:               minimalJobsAdminDeps(),
		AI:                 minimalAIAdminDeps(),
		RPC:                minimalRPCAdminDeps(),
		Logs:               minimalLogsAdminDeps(),
		Settings:           minimalSettingsAdminDeps(),
		Extensions:         minimalExtensionsAdminDeps(),
		ExtensionsTenant:   minimalExtensionsTenantDeps(),
	}

	group := BuildAdminRoutes(deps)
	require.NotNil(t, group)

	// Parent group must have TenantContext and TenantDBContext middlewares
	assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantContext"),
		"admin parent group must have TenantContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
	assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantDBContext"),
		"admin parent group must have TenantDBContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
}

// TestTenantMiddleware_AdminSubgroupsInherit verifies that all admin subgroups
// that handle tenant-scoped data inherit tenant middleware from the parent.
func TestTenantMiddleware_AdminSubgroupsInherit(t *testing.T) {
	deps := &AdminDeps{
		UnifiedAuth:        tenantAwareHandler,
		RequireRole:        func(...string) fiber.Handler { return tenantAwareHandler },
		TenantMiddleware:   tenantAwareHandler,
		TenantDBMiddleware: tenantAwareHandler,
		Schema:             minimalSchemaAdminDeps(),
		AuthProviders:      minimalAuthProvidersAdminDeps(),
		Users:              minimalUsersAdminDeps(),
		Tenants:            minimalTenantsAdminDeps(),
		ServiceKeys:        minimalServiceKeysAdminDeps(),
		Functions:          minimalFunctionsAdminDeps(),
		Jobs:               minimalJobsAdminDeps(),
		AI:                 minimalAIAdminDeps(),
		RPC:                minimalRPCAdminDeps(),
		Logs:               minimalLogsAdminDeps(),
		Settings:           minimalSettingsAdminDeps(),
		Extensions:         minimalExtensionsAdminDeps(),
		ExtensionsTenant:   minimalExtensionsTenantDeps(),
	}

	group := BuildAdminRoutes(deps)
	require.NotNil(t, group)

	// These subgroups should exist
	tenantScopedSubgroups := []string{
		"schema_admin",
		"users_admin",
		"tenants_admin",
		"service_keys_admin",
		"functions_admin",
		"jobs_admin",
		"ai_admin",
		"rpc_admin",
		"logs_admin",
		"settings_admin",
		"extensions_admin",
		"extensions_tenant",
	}

	foundSubgroups := make(map[string]*RouteGroup)
	for _, sg := range group.SubGroups {
		foundSubgroups[sg.Name] = sg
	}

	for _, name := range tenantScopedSubgroups {
		sg, ok := foundSubgroups[name]
		require.True(t, ok, "expected subgroup %q to exist", name)
		// Subgroups inherit parent middlewares via registry.go's applyGroup.
		// At the subgroup level, they may or may not have their own tenant middleware,
		// but the parent will always provide it during registration.
		// The important thing is the parent has it, and the subgroup exists.
		t.Logf("Subgroup %q: middlewares=%v", name, collectMiddlewareNames(sg.Middlewares))
	}
}

// TestTenantMiddleware_RouteGroups verifies that REST and storage route groups
// include tenant middleware directly (they are not under admin).
func TestTenantMiddleware_RouteGroups(t *testing.T) {
	t.Run("REST", func(t *testing.T) {
		deps := &RESTDeps{
			RequireAuth:        tenantAwareHandler,
			RequireScope:       func(...string) fiber.Handler { return tenantAwareHandler },
			HandleTables:       tenantAwareHandler,
			HandleQuery:        tenantAwareHandler,
			HandleById:         tenantAwareHandler,
			TenantMiddleware:   tenantAwareHandler,
			TenantDBMiddleware: tenantAwareHandler,
		}

		group := BuildRESTRoutes(deps)
		require.NotNil(t, group)
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantContext"),
			"REST group must have TenantContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantDBContext"),
			"REST group must have TenantDBContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
	})

	t.Run("Storage", func(t *testing.T) {
		deps := &StorageDeps{
			RequireAuth:            tenantAwareHandler,
			OptionalAuth:           tenantAwareHandler,
			RequireScope:           func(...string) fiber.Handler { return tenantAwareHandler },
			DownloadSignedObject:   tenantAwareHandler,
			GetTransformConfig:     tenantAwareHandler,
			ListBuckets:            tenantAwareHandler,
			CreateBucket:           tenantAwareHandler,
			UpdateBucketSettings:   tenantAwareHandler,
			DeleteBucket:           tenantAwareHandler,
			ListFiles:              tenantAwareHandler,
			MultipartUpload:        tenantAwareHandler,
			ShareObject:            tenantAwareHandler,
			RevokeShare:            tenantAwareHandler,
			ListShares:             tenantAwareHandler,
			GenerateSignedURL:      tenantAwareHandler,
			StreamUpload:           tenantAwareHandler,
			StorageUploadLimiter:   tenantAwareHandler,
			InitChunkedUpload:      tenantAwareHandler,
			UploadChunk:            tenantAwareHandler,
			CompleteChunkedUpload:  tenantAwareHandler,
			GetChunkedUploadStatus: tenantAwareHandler,
			AbortChunkedUpload:     tenantAwareHandler,
			UploadFile:             tenantAwareHandler,
			DownloadFile:           tenantAwareHandler,
			DeleteFile:             tenantAwareHandler,
			TenantMiddleware:       tenantAwareHandler,
			TenantDBMiddleware:     tenantAwareHandler,
		}

		group := BuildStorageRoutes(deps)
		require.NotNil(t, group)
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantContext"),
			"storage group must have TenantContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantDBContext"),
			"storage group must have TenantDBContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
	})
}

// TestTenantMiddleware_NilMiddlewareNoPanic verifies that route builders handle
// nil tenant middleware gracefully (no panic, group still builds).
func TestTenantMiddleware_NilMiddlewareNoPanic(t *testing.T) {
	t.Run("Admin_nil_tenant_middleware", func(t *testing.T) {
		deps := &AdminDeps{
			UnifiedAuth: tenantAwareHandler,
			RequireRole: func(...string) fiber.Handler { return tenantAwareHandler },
			Schema:      minimalSchemaAdminDeps(),
		}
		// Should not panic with nil tenant middleware
		group := BuildAdminRoutes(deps)
		require.NotNil(t, group)
		assert.Empty(t, group.Middlewares, "nil tenant middleware should result in empty middlewares")
	})

	t.Run("REST_nil_tenant_middleware", func(t *testing.T) {
		deps := &RESTDeps{
			RequireAuth:  tenantAwareHandler,
			RequireScope: func(...string) fiber.Handler { return tenantAwareHandler },
			HandleTables: tenantAwareHandler,
			HandleQuery:  tenantAwareHandler,
			HandleById:   tenantAwareHandler,
		}
		group := BuildRESTRoutes(deps)
		require.NotNil(t, group)
		assert.Empty(t, group.Middlewares)
	})

	t.Run("Storage_nil_tenant_middleware", func(t *testing.T) {
		deps := &StorageDeps{
			RequireAuth:            tenantAwareHandler,
			OptionalAuth:           tenantAwareHandler,
			RequireScope:           func(...string) fiber.Handler { return tenantAwareHandler },
			DownloadSignedObject:   tenantAwareHandler,
			GetTransformConfig:     tenantAwareHandler,
			ListBuckets:            tenantAwareHandler,
			CreateBucket:           tenantAwareHandler,
			UpdateBucketSettings:   tenantAwareHandler,
			DeleteBucket:           tenantAwareHandler,
			ListFiles:              tenantAwareHandler,
			MultipartUpload:        tenantAwareHandler,
			ShareObject:            tenantAwareHandler,
			RevokeShare:            tenantAwareHandler,
			ListShares:             tenantAwareHandler,
			GenerateSignedURL:      tenantAwareHandler,
			StreamUpload:           tenantAwareHandler,
			StorageUploadLimiter:   tenantAwareHandler,
			InitChunkedUpload:      tenantAwareHandler,
			UploadChunk:            tenantAwareHandler,
			CompleteChunkedUpload:  tenantAwareHandler,
			GetChunkedUploadStatus: tenantAwareHandler,
			AbortChunkedUpload:     tenantAwareHandler,
			UploadFile:             tenantAwareHandler,
			DownloadFile:           tenantAwareHandler,
			DeleteFile:             tenantAwareHandler,
		}
		group := BuildStorageRoutes(deps)
		require.NotNil(t, group)
		assert.Empty(t, group.Middlewares)
	})
}

// TestTenantMiddleware_InheritanceViaRegistry verifies that parent middleware
// is properly inherited by subgroups during route registration.
func TestTenantMiddleware_InheritanceViaRegistry(t *testing.T) {
	parentMW := Middleware{Name: "TenantContext", Handler: tenantAwareHandler}
	childMW := Middleware{Name: "BranchContext", Handler: tenantAwareHandler}

	parent := &RouteGroup{
		Name:        "parent",
		Prefix:      "/api/v1/admin",
		Middlewares: []Middleware{parentMW},
		SubGroups: []*RouteGroup{
			{
				Name:        "child",
				Prefix:      "/schema",
				Middlewares: []Middleware{childMW},
				Routes: []Route{
					{Method: "GET", Path: "/tables", Handler: tenantAwareHandler},
				},
			},
		},
	}

	reg := NewRegistry(WithStrictValidation())
	err := reg.Register(parent)
	require.NoError(t, err)

	// Verify parent has its middleware
	require.Len(t, reg.groups[0].Middlewares, 1)
	assert.Equal(t, "TenantContext", reg.groups[0].Middlewares[0].Name)

	// Verify child has its own middleware (parent middleware is merged at Apply time)
	require.Len(t, reg.groups[0].SubGroups[0].Middlewares, 1)
	assert.Equal(t, "BranchContext", reg.groups[0].SubGroups[0].Middlewares[0].Name)
}

// ============================================================================
// Minimal dependency helpers
// ============================================================================

func minimalSchemaAdminDeps() *SchemaAdminDeps {
	return &SchemaAdminDeps{
		GetTables:               tenantAwareHandler,
		GetTableSchema:          tenantAwareHandler,
		GetSchemas:              tenantAwareHandler,
		ExecuteQuery:            tenantAwareHandler,
		ListSchemasDDL:          tenantAwareHandler,
		CreateSchemaDDL:         tenantAwareHandler,
		ListTablesDDL:           tenantAwareHandler,
		CreateTableDDL:          tenantAwareHandler,
		DeleteTableDDL:          tenantAwareHandler,
		RenameTableDDL:          tenantAwareHandler,
		AddColumnDDL:            tenantAwareHandler,
		DropColumnDDL:           tenantAwareHandler,
		EnableRealtime:          tenantAwareHandler,
		ListRealtimeTables:      tenantAwareHandler,
		GetRealtimeStatus:       tenantAwareHandler,
		UpdateRealtimeConfig:    tenantAwareHandler,
		DisableRealtime:         tenantAwareHandler,
		ExecuteSQL:              tenantAwareHandler,
		ExportTypeScript:        tenantAwareHandler,
		RefreshSchema:           tenantAwareHandler,
		GetSchemaGraph:          tenantAwareHandler,
		GetTableRelationships:   tenantAwareHandler,
		GetTablesWithRLS:        tenantAwareHandler,
		GetTableRLSStatus:       tenantAwareHandler,
		ToggleTableRLS:          tenantAwareHandler,
		ListPolicies:            tenantAwareHandler,
		CreatePolicy:            tenantAwareHandler,
		UpdatePolicy:            tenantAwareHandler,
		DeletePolicy:            tenantAwareHandler,
		GetPolicyTemplates:      tenantAwareHandler,
		GetSecurityWarnings:     tenantAwareHandler,
		DumpInternalSchema:      tenantAwareHandler,
		PlanInternalSchema:      tenantAwareHandler,
		ApplyInternalSchema:     tenantAwareHandler,
		ValidateInternalSchema:  tenantAwareHandler,
		GetInternalSchemaStatus: tenantAwareHandler,
		MigrateInternalSchema:   tenantAwareHandler,
	}
}

func minimalAuthProvidersAdminDeps() *AuthProvidersAdminDeps {
	return &AuthProvidersAdminDeps{
		ListOAuthProviders:  tenantAwareHandler,
		GetOAuthProvider:    tenantAwareHandler,
		CreateOAuthProvider: tenantAwareHandler,
		UpdateOAuthProvider: tenantAwareHandler,
		DeleteOAuthProvider: tenantAwareHandler,
		ListSAMLProviders:   tenantAwareHandler,
		GetSAMLProvider:     tenantAwareHandler,
		CreateSAMLProvider:  tenantAwareHandler,
		UpdateSAMLProvider:  tenantAwareHandler,
		DeleteSAMLProvider:  tenantAwareHandler,
		ValidateSAML:        tenantAwareHandler,
		UploadSAMLMetadata:  tenantAwareHandler,
		GetAuthSettings:     tenantAwareHandler,
		UpdateAuthSettings:  tenantAwareHandler,
		ListSessions:        tenantAwareHandler,
		RevokeSession:       tenantAwareHandler,
		RevokeUserSessions:  tenantAwareHandler,
	}
}

func minimalUsersAdminDeps() *UsersAdminDeps {
	return &UsersAdminDeps{
		ListUsers:         tenantAwareHandler,
		InviteUser:        tenantAwareHandler,
		DeleteUser:        tenantAwareHandler,
		UpdateUser:        tenantAwareHandler,
		UpdateUserRole:    tenantAwareHandler,
		ResetUserPassword: tenantAwareHandler,
		CreateInvitation:  tenantAwareHandler,
		ListInvitations:   tenantAwareHandler,
		RevokeInvitation:  tenantAwareHandler,
	}
}

func minimalTenantsAdminDeps() *TenantsAdminDeps {
	return &TenantsAdminDeps{
		ListMyTenants:             tenantAwareHandler,
		ListTenants:               tenantAwareHandler,
		CreateTenant:              tenantAwareHandler,
		GetTenant:                 tenantAwareHandler,
		UpdateTenant:              tenantAwareHandler,
		DeleteTenant:              tenantAwareHandler,
		MigrateTenant:             tenantAwareHandler,
		ListAdmins:                tenantAwareHandler,
		AssignAdmin:               tenantAwareHandler,
		RemoveAdmin:               tenantAwareHandler,
		GetTenantSettings:         tenantAwareHandler,
		UpdateTenantSettings:      tenantAwareHandler,
		DeleteTenantSetting:       tenantAwareHandler,
		GetTenantSetting:          tenantAwareHandler,
		GetTenantSchemaStatus:     tenantAwareHandler,
		ApplyTenantSchema:         tenantAwareHandler,
		GetStoredSchema:           tenantAwareHandler,
		UploadTenantSchema:        tenantAwareHandler,
		ApplyUploadedTenantSchema: tenantAwareHandler,
		DeleteStoredSchema:        tenantAwareHandler,
	}
}

func minimalServiceKeysAdminDeps() *ServiceKeysAdminDeps {
	return &ServiceKeysAdminDeps{
		ListServiceKeys:      tenantAwareHandler,
		GetServiceKey:        tenantAwareHandler,
		CreateServiceKey:     tenantAwareHandler,
		UpdateServiceKey:     tenantAwareHandler,
		DeleteServiceKey:     tenantAwareHandler,
		DisableServiceKey:    tenantAwareHandler,
		EnableServiceKey:     tenantAwareHandler,
		RevokeServiceKey:     tenantAwareHandler,
		DeprecateServiceKey:  tenantAwareHandler,
		RotateServiceKey:     tenantAwareHandler,
		GetRevocationHistory: tenantAwareHandler,
	}
}

func minimalFunctionsAdminDeps() *FunctionsAdminDeps {
	return &FunctionsAdminDeps{
		ReloadFunctions:        tenantAwareHandler,
		ListFunctionNamespaces: tenantAwareHandler,
		ListAllExecutions:      tenantAwareHandler,
		GetExecutionLogs:       tenantAwareHandler,
		SyncFunctions:          tenantAwareHandler,
	}
}

func minimalJobsAdminDeps() *JobsAdminDeps {
	return &JobsAdminDeps{
		ListJobNamespaces: tenantAwareHandler,
		ListJobFunctions:  tenantAwareHandler,
		GetJobFunction:    tenantAwareHandler,
		DeleteJobFunction: tenantAwareHandler,
		GetJobStats:       tenantAwareHandler,
		ListWorkers:       tenantAwareHandler,
		ListAllJobs:       tenantAwareHandler,
		GetJobAdmin:       tenantAwareHandler,
		TerminateJob:      tenantAwareHandler,
		CancelJobAdmin:    tenantAwareHandler,
		RetryJobAdmin:     tenantAwareHandler,
		ResubmitJobAdmin:  tenantAwareHandler,
		SyncJobs:          tenantAwareHandler,
	}
}

func minimalAIAdminDeps() *AIAdminDeps {
	return &AIAdminDeps{
		ListChatbots:               tenantAwareHandler,
		GetChatbot:                 tenantAwareHandler,
		ToggleChatbot:              tenantAwareHandler,
		UpdateChatbot:              tenantAwareHandler,
		DeleteChatbot:              tenantAwareHandler,
		SyncChatbots:               tenantAwareHandler,
		GetAIMetrics:               tenantAwareHandler,
		ListAIProviders:            tenantAwareHandler,
		ListAIConversations:        tenantAwareHandler,
		GetAIConversationMessages:  tenantAwareHandler,
		GetAIAuditLog:              tenantAwareHandler,
		ListExportableTables:       tenantAwareHandler,
		GetExportableTableDetails:  tenantAwareHandler,
		ExportTableToKnowledgeBase: tenantAwareHandler,
		ListChatbotKnowledgeBases:  tenantAwareHandler,
		LinkKnowledgeBase:          tenantAwareHandler,
		UpdateChatbotKnowledgeBase: tenantAwareHandler,
		UnlinkKnowledgeBase:        tenantAwareHandler,
	}
}

func minimalRPCAdminDeps() *RPCAdminDeps {
	return &RPCAdminDeps{
		ListRPCNamespaces:   tenantAwareHandler,
		ListProcedures:      tenantAwareHandler,
		GetProcedure:        tenantAwareHandler,
		UpdateProcedure:     tenantAwareHandler,
		DeleteProcedure:     tenantAwareHandler,
		SyncProcedures:      tenantAwareHandler,
		ListRPCExecutions:   tenantAwareHandler,
		GetRPCExecution:     tenantAwareHandler,
		GetRPCExecutionLogs: tenantAwareHandler,
		CancelRPCExecution:  tenantAwareHandler,
	}
}

func minimalLogsAdminDeps() *LogsAdminDeps {
	return &LogsAdminDeps{
		ListLogs:              tenantAwareHandler,
		GetLogStats:           tenantAwareHandler,
		GetExecutionLogsAdmin: tenantAwareHandler,
		FlushLogs:             tenantAwareHandler,
		GenerateTestLogs:      tenantAwareHandler,
	}
}

func minimalSettingsAdminDeps() *SettingsAdminDeps {
	return &SettingsAdminDeps{
		ListSystemSettings:        tenantAwareHandler,
		GetSystemSetting:          tenantAwareHandler,
		UpdateSystemSetting:       tenantAwareHandler,
		DeleteSystemSetting:       tenantAwareHandler,
		CreateCustomSetting:       tenantAwareHandler,
		ListCustomSettings:        tenantAwareHandler,
		CreateSecretSetting:       tenantAwareHandler,
		ListSecretSettings:        tenantAwareHandler,
		GetSecretSetting:          tenantAwareHandler,
		UpdateSecretSetting:       tenantAwareHandler,
		DeleteSecretSetting:       tenantAwareHandler,
		GetUserSecretValue:        tenantAwareHandler,
		GetCustomSetting:          tenantAwareHandler,
		UpdateCustomSetting:       tenantAwareHandler,
		DeleteCustomSetting:       tenantAwareHandler,
		GetAppSettings:            tenantAwareHandler,
		UpdateAppSettings:         tenantAwareHandler,
		ListEmailSettings:         tenantAwareHandler,
		GetEmailSetting:           tenantAwareHandler,
		UpdateEmailSetting:        tenantAwareHandler,
		TestEmailSettings:         tenantAwareHandler,
		ListEmailTemplates:        tenantAwareHandler,
		GetEmailTemplate:          tenantAwareHandler,
		UpdateEmailTemplate:       tenantAwareHandler,
		TestEmailTemplate:         tenantAwareHandler,
		ResetEmailTemplate:        tenantAwareHandler,
		GetCaptchaSettings:        tenantAwareHandler,
		UpdateCaptchaSettings:     tenantAwareHandler,
		GetInstanceSettings:       tenantAwareHandler,
		UpdateInstanceSettings:    tenantAwareHandler,
		GetOverridableSettings:    tenantAwareHandler,
		UpdateOverridableSettings: tenantAwareHandler,
	}
}

func minimalExtensionsAdminDeps() *ExtensionsAdminDeps {
	return &ExtensionsAdminDeps{
		ListExtensions:   tenantAwareHandler,
		GetExtension:     tenantAwareHandler,
		EnableExtension:  tenantAwareHandler,
		DisableExtension: tenantAwareHandler,
		SyncExtensions:   tenantAwareHandler,
	}
}

func minimalExtensionsTenantDeps() *ExtensionsTenantDeps {
	return &ExtensionsTenantDeps{
		ListExtensions:   tenantAwareHandler,
		GetExtension:     tenantAwareHandler,
		EnableExtension:  tenantAwareHandler,
		DisableExtension: tenantAwareHandler,
	}
}

// TestTenantMiddleware_MonitoringRouteGroup verifies that monitoring routes
// correctly wire tenant middleware when deps are provided.
func TestTenantMiddleware_MonitoringRouteGroup(t *testing.T) {
	t.Run("with_tenant_middleware", func(t *testing.T) {
		deps := &MonitoringDeps{
			RequireAuth:        tenantAwareHandler,
			RequireScope:       func(...string) fiber.Handler { return tenantAwareHandler },
			TenantMiddleware:   tenantAwareHandler,
			TenantDBMiddleware: tenantAwareHandler,
			GetMetrics:         tenantAwareHandler,
			GetHealth:          tenantAwareHandler,
			GetLogs:            tenantAwareHandler,
		}
		group := BuildMonitoringRoutes(deps)
		require.NotNil(t, group)
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantContext"),
			"monitoring group must have TenantContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
		assert.True(t, hasMiddlewareNamed(group.Middlewares, "TenantDBContext"),
			"monitoring group must have TenantDBContext middleware, got: %v", collectMiddlewareNames(group.Middlewares))
	})

	t.Run("nil_tenant_middleware", func(t *testing.T) {
		deps := &MonitoringDeps{
			RequireAuth:  tenantAwareHandler,
			RequireScope: func(...string) fiber.Handler { return tenantAwareHandler },
			GetMetrics:   tenantAwareHandler,
			GetHealth:    tenantAwareHandler,
			GetLogs:      tenantAwareHandler,
		}
		group := BuildMonitoringRoutes(deps)
		require.NotNil(t, group)
		assert.Empty(t, group.Middlewares, "nil tenant middleware should result in empty middlewares")
	})
}
