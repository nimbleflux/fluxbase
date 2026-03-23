package routes

import (
	"github.com/gofiber/fiber/v3"
)

type AdminDeps struct {
	UnifiedAuth fiber.Handler
	RequireRole func(...string) fiber.Handler

	GetTables              fiber.Handler
	GetTableSchema         fiber.Handler
	GetSchemas             fiber.Handler
	ExecuteQuery           fiber.Handler
	ListSchemasDDL         fiber.Handler
	CreateSchemaDDL        fiber.Handler
	ListTablesDDL          fiber.Handler
	CreateTableDDL         fiber.Handler
	DeleteTableDDL         fiber.Handler
	RenameTableDDL         fiber.Handler
	AddColumnDDL           fiber.Handler
	DropColumnDDL          fiber.Handler
	EnableRealtime         fiber.Handler
	ListRealtimeTables     fiber.Handler
	GetRealtimeStatus      fiber.Handler
	UpdateRealtimeConfig   fiber.Handler
	DisableRealtime        fiber.Handler
	ListOAuthProviders     fiber.Handler
	GetOAuthProvider       fiber.Handler
	CreateOAuthProvider    fiber.Handler
	UpdateOAuthProvider    fiber.Handler
	DeleteOAuthProvider    fiber.Handler
	ListSAMLProviders      fiber.Handler
	GetSAMLProvider        fiber.Handler
	CreateSAMLProvider     fiber.Handler
	UpdateSAMLProvider     fiber.Handler
	DeleteSAMLProvider     fiber.Handler
	ValidateSAML           fiber.Handler
	UploadSAMLMetadata     fiber.Handler
	GetAuthSettings        fiber.Handler
	UpdateAuthSettings     fiber.Handler
	ListSessions           fiber.Handler
	RevokeSession          fiber.Handler
	RevokeUserSessions     fiber.Handler
	ListSystemSettings     fiber.Handler
	GetSystemSetting       fiber.Handler
	UpdateSystemSetting    fiber.Handler
	DeleteSystemSetting    fiber.Handler
	CreateCustomSetting    fiber.Handler
	ListCustomSettings     fiber.Handler
	CreateSecretSetting    fiber.Handler
	ListSecretSettings     fiber.Handler
	GetSecretSetting       fiber.Handler
	UpdateSecretSetting    fiber.Handler
	DeleteSecretSetting    fiber.Handler
	GetUserSecretValue     fiber.Handler
	GetCustomSetting       fiber.Handler
	UpdateCustomSetting    fiber.Handler
	DeleteCustomSetting    fiber.Handler
	GetAppSettings         fiber.Handler
	UpdateAppSettings      fiber.Handler
	ListEmailSettings      fiber.Handler
	GetEmailSetting        fiber.Handler
	UpdateEmailSetting     fiber.Handler
	TestEmailSettings      fiber.Handler
	ListEmailTemplates     fiber.Handler
	GetEmailTemplate       fiber.Handler
	UpdateEmailTemplate    fiber.Handler
	TestEmailTemplate      fiber.Handler
	GetCaptchaSettings     fiber.Handler
	UpdateCaptchaSettings  fiber.Handler
	ListUsers              fiber.Handler
	InviteUser             fiber.Handler
	DeleteUser             fiber.Handler
	UpdateUser             fiber.Handler
	UpdateUserRole         fiber.Handler
	ResetUserPassword      fiber.Handler
	ListUsersWithQuotas    fiber.Handler
	GetUserQuota           fiber.Handler
	SetUserQuota           fiber.Handler
	CreateInvitation       fiber.Handler
	ListInvitations        fiber.Handler
	RevokeInvitation       fiber.Handler
	ListServiceKeys        fiber.Handler
	GetServiceKey          fiber.Handler
	CreateServiceKey       fiber.Handler
	UpdateServiceKey       fiber.Handler
	DeleteServiceKey       fiber.Handler
	DisableServiceKey      fiber.Handler
	EnableServiceKey       fiber.Handler
	RevokeServiceKey       fiber.Handler
	DeprecateServiceKey    fiber.Handler
	RotateServiceKey       fiber.Handler
	GetRevocationHistory   fiber.Handler
	ListMyTenants          fiber.Handler
	ListTenants            fiber.Handler
	CreateTenant           fiber.Handler
	GetTenant              fiber.Handler
	UpdateTenant           fiber.Handler
	DeleteTenant           fiber.Handler
	MigrateTenant          fiber.Handler
	ListAdmins             fiber.Handler
	AssignAdmin            fiber.Handler
	RemoveAdmin            fiber.Handler
	ExecuteSQL             fiber.Handler
	ExportTypeScript       fiber.Handler
	ReloadFunctions        fiber.Handler
	ListFunctionNamespaces fiber.Handler
	ListAllExecutions      fiber.Handler
	GetExecutionLogs       fiber.Handler
	ListJobNamespaces      fiber.Handler
	ListJobFunctions       fiber.Handler
	GetJobFunction         fiber.Handler
	DeleteJobFunction      fiber.Handler
	GetJobStats            fiber.Handler
	ListWorkers            fiber.Handler
	ListAllJobs            fiber.Handler
	GetJobAdmin            fiber.Handler
	TerminateJob           fiber.Handler
	CancelJobAdmin         fiber.Handler
	RetryJobAdmin          fiber.Handler
	ResubmitJobAdmin       fiber.Handler
	ListChatbots           fiber.Handler
	GetChatbot             fiber.Handler
	ToggleChatbot          fiber.Handler
	UpdateChatbot          fiber.Handler
	DeleteChatbot          fiber.Handler
	GetAIMetrics           fiber.Handler
	SyncFunctions          fiber.Handler
	SyncJobs               fiber.Handler
	SyncChatbots           fiber.Handler
	SyncProcedures         fiber.Handler
	RefreshSchema          fiber.Handler
	ListLogs               fiber.Handler
	GetLogStats            fiber.Handler
	GetExecutionLogsAdmin  fiber.Handler
	FlushLogs              fiber.Handler
	GenerateTestLogs       fiber.Handler

	// Instance Settings
	GetInstanceSettings       fiber.Handler
	UpdateInstanceSettings    fiber.Handler
	GetOverridableSettings    fiber.Handler
	UpdateOverridableSettings fiber.Handler

	// Tenant Settings
	GetTenantSettings    fiber.Handler
	UpdateTenantSettings fiber.Handler
	DeleteTenantSetting  fiber.Handler
	GetTenantSetting     fiber.Handler

	// Extensions (instance-level)
	ListExtensions   fiber.Handler
	GetExtension     fiber.Handler
	EnableExtension  fiber.Handler
	DisableExtension fiber.Handler
	SyncExtensions   fiber.Handler

	// AI Admin
	ListAIProviders            fiber.Handler
	ListAIConversations        fiber.Handler
	GetAIConversationMessages  fiber.Handler
	GetAIAuditLog              fiber.Handler
	ListExportableTables       fiber.Handler
	GetExportableTableDetails  fiber.Handler
	ExportTableToKnowledgeBase fiber.Handler

	// RPC Admin
	ListRPCNamespaces   fiber.Handler
	ListProcedures      fiber.Handler
	GetProcedure        fiber.Handler
	UpdateProcedure     fiber.Handler
	DeleteProcedure     fiber.Handler
	ListRPCExecutions   fiber.Handler
	GetRPCExecution     fiber.Handler
	GetRPCExecutionLogs fiber.Handler
	CancelRPCExecution  fiber.Handler

	// Schema/RLS/Policy
	GetSchemaGraph        fiber.Handler
	GetTableRelationships fiber.Handler
	GetTablesWithRLS      fiber.Handler
	GetTableRLSStatus     fiber.Handler
	ToggleTableRLS        fiber.Handler
	ListPolicies          fiber.Handler
	CreatePolicy          fiber.Handler
	UpdatePolicy          fiber.Handler
	DeletePolicy          fiber.Handler
	GetPolicyTemplates    fiber.Handler
	GetSecurityWarnings   fiber.Handler
}

func BuildAdminRoutes(deps *AdminDeps) *RouteGroup {
	routes := []Route{
		{Method: "GET", Path: "/tables", Handler: deps.GetTables, Summary: "List all tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/tables/:schema/:table", Handler: deps.GetTableSchema, Summary: "Get table schema", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/schemas", Handler: deps.GetSchemas, Summary: "List schemas", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/query", Handler: deps.ExecuteQuery, Summary: "Execute SQL query", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "GET", Path: "/ddl/schemas", Handler: deps.ListSchemasDDL, Summary: "List schemas for DDL", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/ddl/schemas", Handler: deps.CreateSchemaDDL, Summary: "Create schema", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/ddl/tables", Handler: deps.ListTablesDDL, Summary: "List tables for DDL", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/ddl/tables", Handler: deps.CreateTableDDL, Summary: "Create table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/ddl/tables/:schema/:table", Handler: deps.DeleteTableDDL, Summary: "Delete table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/schemas", Handler: deps.CreateSchemaDDL, Summary: "Create schema (legacy)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/tables", Handler: deps.CreateTableDDL, Summary: "Create table (legacy)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tables/:schema/:table", Handler: deps.DeleteTableDDL, Summary: "Delete table (legacy)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PATCH", Path: "/tables/:schema/:table", Handler: deps.RenameTableDDL, Summary: "Rename table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/tables/:schema/:table/columns", Handler: deps.AddColumnDDL, Summary: "Add column", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tables/:schema/:table/columns/:column", Handler: deps.DropColumnDDL, Summary: "Drop column", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "POST", Path: "/realtime/tables", Handler: deps.EnableRealtime, Summary: "Enable realtime for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/realtime/tables", Handler: deps.ListRealtimeTables, Summary: "List realtime tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/realtime/tables/:schema/:table", Handler: deps.GetRealtimeStatus, Summary: "Get realtime status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PATCH", Path: "/realtime/tables/:schema/:table", Handler: deps.UpdateRealtimeConfig, Summary: "Update realtime config", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/realtime/tables/:schema/:table", Handler: deps.DisableRealtime, Summary: "Disable realtime for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "GET", Path: "/oauth/providers", Handler: deps.ListOAuthProviders, Summary: "List OAuth providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/oauth/providers/:id", Handler: deps.GetOAuthProvider, Summary: "Get OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/oauth/providers", Handler: deps.CreateOAuthProvider, Summary: "Create OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/oauth/providers/:id", Handler: deps.UpdateOAuthProvider, Summary: "Update OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/oauth/providers/:id", Handler: deps.DeleteOAuthProvider, Summary: "Delete OAuth provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "GET", Path: "/saml/providers", Handler: deps.ListSAMLProviders, Summary: "List SAML providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/saml/providers/:id", Handler: deps.GetSAMLProvider, Summary: "Get SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/saml/providers", Handler: deps.CreateSAMLProvider, Summary: "Create SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/saml/providers/:id", Handler: deps.UpdateSAMLProvider, Summary: "Update SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/saml/providers/:id", Handler: deps.DeleteSAMLProvider, Summary: "Delete SAML provider", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/saml/validate-metadata", Handler: deps.ValidateSAML, Summary: "Validate SAML metadata", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/saml/upload-metadata", Handler: deps.UploadSAMLMetadata, Summary: "Upload SAML metadata", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "GET", Path: "/auth/settings", Handler: deps.GetAuthSettings, Summary: "Get auth settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/auth/settings", Handler: deps.UpdateAuthSettings, Summary: "Update auth settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/auth/sessions", Handler: deps.ListSessions, Summary: "List sessions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/auth/sessions/:id", Handler: deps.RevokeSession, Summary: "Revoke session", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/auth/sessions/user/:user_id", Handler: deps.RevokeUserSessions, Summary: "Revoke user sessions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "GET", Path: "/system/settings", Handler: deps.ListSystemSettings, Summary: "List system settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/system/settings/*", Handler: deps.GetSystemSetting, Summary: "Get system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/system/settings/*", Handler: deps.UpdateSystemSetting, Summary: "Update system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/system/settings/*", Handler: deps.DeleteSystemSetting, Summary: "Delete system setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "POST", Path: "/settings/custom", Handler: deps.CreateCustomSetting, Summary: "Create custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "GET", Path: "/settings/custom", Handler: deps.ListCustomSettings, Summary: "List custom settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "POST", Path: "/settings/custom/secret", Handler: deps.CreateSecretSetting, Summary: "Create secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "GET", Path: "/settings/custom/secrets", Handler: deps.ListSecretSettings, Summary: "List secret settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "GET", Path: "/settings/custom/secret/*", Handler: deps.GetSecretSetting, Summary: "Get secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "PUT", Path: "/settings/custom/secret/*", Handler: deps.UpdateSecretSetting, Summary: "Update secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "DELETE", Path: "/settings/custom/secret/*", Handler: deps.DeleteSecretSetting, Summary: "Delete secret setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "GET", Path: "/settings/user/:user_id/secret/:key/decrypt", Handler: deps.GetUserSecretValue, Summary: "Decrypt user secret (service_role only)", Auth: AuthRequired, Roles: []string{"service_role"}},
		{Method: "GET", Path: "/settings/custom/*", Handler: deps.GetCustomSetting, Summary: "Get custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "PUT", Path: "/settings/custom/*", Handler: deps.UpdateCustomSetting, Summary: "Update custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},
		{Method: "DELETE", Path: "/settings/custom/*", Handler: deps.DeleteCustomSetting, Summary: "Delete custom setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin", "service_role"}},

		{Method: "GET", Path: "/app/settings", Handler: deps.GetAppSettings, Summary: "Get app settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PUT", Path: "/app/settings", Handler: deps.UpdateAppSettings, Summary: "Update app settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "GET", Path: "/email/settings", Handler: deps.ListEmailSettings, Summary: "List email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/email/settings/:provider", Handler: deps.GetEmailSetting, Summary: "Get email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/email/settings/:provider", Handler: deps.UpdateEmailSetting, Summary: "Update email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/email/settings/:provider/test", Handler: deps.TestEmailSettings, Summary: "Test email settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/email/templates", Handler: deps.ListEmailTemplates, Summary: "List email templates", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/email/templates/:name", Handler: deps.GetEmailTemplate, Summary: "Get email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/email/templates/:name", Handler: deps.UpdateEmailTemplate, Summary: "Update email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/email/templates/:name/test", Handler: deps.TestEmailTemplate, Summary: "Test email template", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "GET", Path: "/settings/captcha", Handler: deps.GetCaptchaSettings, Summary: "Get captcha settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/settings/captcha", Handler: deps.UpdateCaptchaSettings, Summary: "Update captcha settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "GET", Path: "/users", Handler: deps.ListUsers, Summary: "List users", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/users/invite", Handler: deps.InviteUser, Summary: "Invite user", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/users/:id", Handler: deps.ListUsers, Summary: "Get user by ID", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PATCH", Path: "/users/:id", Handler: deps.UpdateUser, Summary: "Update user", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/users/:id", Handler: deps.DeleteUser, Summary: "Delete user", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PATCH", Path: "/users/:id/role", Handler: deps.UpdateUserRole, Summary: "Update user role", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/users/:id/reset-password", Handler: deps.ResetUserPassword, Summary: "Reset user password", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		// TODO: Implement quota handlers
		// {Method: "GET", Path: "/users/quotas", Handler: deps.ListUsersWithQuotas, Summary: "List users with quotas", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		// {Method: "GET", Path: "/users/:id/quota", Handler: deps.GetUserQuota, Summary: "Get user quota", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		// {Method: "PUT", Path: "/users/:id/quota", Handler: deps.SetUserQuota, Summary: "Set user quota", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "POST", Path: "/invitations", Handler: deps.CreateInvitation, Summary: "Create invitation", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/invitations", Handler: deps.ListInvitations, Summary: "List invitations", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/invitations/:id", Handler: deps.RevokeInvitation, Summary: "Revoke invitation", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "GET", Path: "/service-keys", Handler: deps.ListServiceKeys, Summary: "List service keys", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys", Handler: deps.CreateServiceKey, Summary: "Create service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/service-keys/:id", Handler: deps.GetServiceKey, Summary: "Get service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/service-keys/:id", Handler: deps.UpdateServiceKey, Summary: "Update service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/service-keys/:id", Handler: deps.DeleteServiceKey, Summary: "Delete service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys/:id/disable", Handler: deps.DisableServiceKey, Summary: "Disable service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys/:id/enable", Handler: deps.EnableServiceKey, Summary: "Enable service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys/:id/revoke", Handler: deps.RevokeServiceKey, Summary: "Revoke service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys/:id/deprecate", Handler: deps.DeprecateServiceKey, Summary: "Deprecate service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/service-keys/:id/rotate", Handler: deps.RotateServiceKey, Summary: "Rotate service key", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/service-keys/:id/revocations", Handler: deps.GetRevocationHistory, Summary: "Get revocation history", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		{Method: "GET", Path: "/tenants/mine", Handler: deps.ListMyTenants, Summary: "List my tenants", Auth: AuthRequired},
		{Method: "GET", Path: "/tenants", Handler: deps.ListTenants, Summary: "List all tenants", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/tenants", Handler: deps.CreateTenant, Summary: "Create tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/tenants/:id", Handler: deps.GetTenant, Summary: "Get tenant", Auth: AuthRequired},
		{Method: "PATCH", Path: "/tenants/:id", Handler: deps.UpdateTenant, Summary: "Update tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tenants/:id", Handler: deps.DeleteTenant, Summary: "Delete tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/tenants/:id/migrate", Handler: deps.MigrateTenant, Summary: "Migrate tenant", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/tenants/:id/admins", Handler: deps.ListAdmins, Summary: "List tenant admins", Auth: AuthRequired},
		{Method: "POST", Path: "/tenants/:id/admins", Handler: deps.AssignAdmin, Summary: "Assign tenant admin", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tenants/:id/admins/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant admin", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "POST", Path: "/sql", Handler: deps.ExecuteSQL, Summary: "Execute SQL", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/schema/export/typescript", Handler: deps.ExportTypeScript, Summary: "Export TypeScript types", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/schema/refresh", Handler: deps.RefreshSchema, Summary: "Refresh schema cache", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		{Method: "POST", Path: "/functions/reload", Handler: deps.ReloadFunctions, Summary: "Reload functions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/functions/namespaces", Handler: deps.ListFunctionNamespaces, Summary: "List function namespaces", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/functions/executions", Handler: deps.ListAllExecutions, Summary: "List all function executions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/functions/executions/:id/logs", Handler: deps.GetExecutionLogs, Summary: "Get function execution logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		// Job admin handlers
		{Method: "GET", Path: "/jobs/namespaces", Handler: deps.ListJobNamespaces, Summary: "List job namespaces", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/jobs/functions", Handler: deps.ListJobFunctions, Summary: "List job functions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/jobs/functions/:namespace/:name", Handler: deps.GetJobFunction, Summary: "Get job function", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/jobs/functions/:namespace/:name", Handler: deps.DeleteJobFunction, Summary: "Delete job function", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/jobs/stats", Handler: deps.GetJobStats, Summary: "Get job stats", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/jobs/workers", Handler: deps.ListWorkers, Summary: "List workers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/jobs", Handler: deps.ListAllJobs, Summary: "List all jobs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/jobs/:id", Handler: deps.GetJobAdmin, Summary: "Get job (admin)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/:id/terminate", Handler: deps.TerminateJob, Summary: "Terminate job", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/:id/cancel", Handler: deps.CancelJobAdmin, Summary: "Cancel job (admin)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/:id/retry", Handler: deps.RetryJobAdmin, Summary: "Retry job (admin)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/:id/resubmit", Handler: deps.ResubmitJobAdmin, Summary: "Resubmit job", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/sync", Handler: deps.SyncJobs, Summary: "Sync jobs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Job queue route aliases (frontend compatibility)
		{Method: "GET", Path: "/jobs/queue", Handler: deps.ListAllJobs, Summary: "List all jobs (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/jobs/queue/:id", Handler: deps.GetJobAdmin, Summary: "Get job (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/queue/:id/cancel", Handler: deps.CancelJobAdmin, Summary: "Cancel job (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/queue/:id/terminate", Handler: deps.TerminateJob, Summary: "Terminate job (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/queue/:id/retry", Handler: deps.RetryJobAdmin, Summary: "Retry job (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/jobs/queue/:id/resubmit", Handler: deps.ResubmitJobAdmin, Summary: "Resubmit job (queue alias)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Chatbot admin handlers
		{Method: "GET", Path: "/ai/chatbots", Handler: deps.ListChatbots, Summary: "List chatbots", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/ai/chatbots/:id", Handler: deps.GetChatbot, Summary: "Get chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/ai/chatbots/:id/toggle", Handler: deps.ToggleChatbot, Summary: "Toggle chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PUT", Path: "/ai/chatbots/:id", Handler: deps.UpdateChatbot, Summary: "Update chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/ai/chatbots/:id", Handler: deps.DeleteChatbot, Summary: "Delete chatbot", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/ai/chatbots/sync", Handler: deps.SyncChatbots, Summary: "Sync chatbots", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/ai/metrics", Handler: deps.GetAIMetrics, Summary: "Get AI metrics", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Log admin handlers
		{Method: "GET", Path: "/logs", Handler: deps.ListLogs, Summary: "List logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/logs/stats", Handler: deps.GetLogStats, Summary: "Get log stats", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/logs/executions/:id", Handler: deps.GetExecutionLogsAdmin, Summary: "Get execution logs (admin)", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/logs/flush", Handler: deps.FlushLogs, Summary: "Flush logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/logs/test", Handler: deps.GenerateTestLogs, Summary: "Generate test logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Instance Settings
		{Method: "GET", Path: "/instance/settings", Handler: deps.GetInstanceSettings, Summary: "Get instance settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PATCH", Path: "/instance/settings", Handler: deps.UpdateInstanceSettings, Summary: "Update instance settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/instance/settings/overridable", Handler: deps.GetOverridableSettings, Summary: "Get overridable settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/instance/settings/overridable", Handler: deps.UpdateOverridableSettings, Summary: "Update overridable settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Extensions (instance-level)
		{Method: "GET", Path: "/extensions", Handler: deps.ListExtensions, Summary: "List extensions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/extensions/:name", Handler: deps.GetExtension, Summary: "Get extension status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/extensions/:name/enable", Handler: deps.EnableExtension, Summary: "Enable extension", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/extensions/:name/disable", Handler: deps.DisableExtension, Summary: "Disable extension", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/extensions/sync", Handler: deps.SyncExtensions, Summary: "Sync extensions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// AI Providers (instance-level)
		{Method: "GET", Path: "/ai/providers", Handler: deps.ListAIProviders, Summary: "List AI providers", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// AI Admin - Conversations, Audit, Tables
		{Method: "GET", Path: "/ai/conversations", Handler: deps.ListAIConversations, Summary: "List AI conversations", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/ai/conversations/:id/messages", Handler: deps.GetAIConversationMessages, Summary: "Get AI conversation messages", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/ai/audit", Handler: deps.GetAIAuditLog, Summary: "Get AI audit log", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/ai/tables", Handler: deps.ListExportableTables, Summary: "List exportable AI tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/ai/tables/:schema/:table", Handler: deps.GetExportableTableDetails, Summary: "Get exportable table details", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/ai/tables/:schema/:table/export", Handler: deps.ExportTableToKnowledgeBase, Summary: "Export table to knowledge base", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		// Tenant Settings
		{Method: "GET", Path: "/tenants/:id/settings", Handler: deps.GetTenantSettings, Summary: "Get tenant settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PATCH", Path: "/tenants/:id/settings", Handler: deps.UpdateTenantSettings, Summary: "Update tenant settings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tenants/:id/settings/*", Handler: deps.DeleteTenantSetting, Summary: "Delete tenant setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/tenants/:id/settings/*", Handler: deps.GetTenantSetting, Summary: "Get tenant setting", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		// Tenant members (aliases for frontend compatibility - backend uses /admins)
		{Method: "GET", Path: "/tenants/:id/members", Handler: deps.ListAdmins, Summary: "List tenant members", Auth: AuthRequired},
		{Method: "POST", Path: "/tenants/:id/members", Handler: deps.AssignAdmin, Summary: "Add tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PATCH", Path: "/tenants/:id/members/:user_id", Handler: deps.AssignAdmin, Summary: "Update tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "DELETE", Path: "/tenants/:id/members/:user_id", Handler: deps.RemoveAdmin, Summary: "Remove tenant member", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

		// RPC Admin routes
		{Method: "GET", Path: "/rpc/namespaces", Handler: deps.ListRPCNamespaces, Summary: "List RPC namespaces", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/rpc/procedures", Handler: deps.ListProcedures, Summary: "List RPC procedures", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/rpc/procedures/:namespace/:name", Handler: deps.GetProcedure, Summary: "Get RPC procedure", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "PUT", Path: "/rpc/procedures/:namespace/:name", Handler: deps.UpdateProcedure, Summary: "Update RPC procedure", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/rpc/procedures/:namespace/:name", Handler: deps.DeleteProcedure, Summary: "Delete RPC procedure", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/rpc/sync", Handler: deps.SyncProcedures, Summary: "Sync RPC procedures", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/rpc/executions", Handler: deps.ListRPCExecutions, Summary: "List RPC executions", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/rpc/executions/:id", Handler: deps.GetRPCExecution, Summary: "Get RPC execution", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/rpc/executions/:id/logs", Handler: deps.GetRPCExecutionLogs, Summary: "Get RPC execution logs", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "POST", Path: "/rpc/executions/:id/cancel", Handler: deps.CancelRPCExecution, Summary: "Cancel RPC execution", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

		// Schema/RLS/Policy routes
		{Method: "GET", Path: "/schema/graph", Handler: deps.GetSchemaGraph, Summary: "Get schema graph", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/tables/:schema/:table/relationships", Handler: deps.GetTableRelationships, Summary: "Get table relationships", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/tables/rls", Handler: deps.GetTablesWithRLS, Summary: "Get tables with RLS status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/tables/:schema/:table/rls", Handler: deps.GetTableRLSStatus, Summary: "Get table RLS status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/tables/:schema/:table/rls/toggle", Handler: deps.ToggleTableRLS, Summary: "Toggle table RLS", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/policies", Handler: deps.ListPolicies, Summary: "List RLS policies", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "POST", Path: "/policies", Handler: deps.CreatePolicy, Summary: "Create RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/policies/:schema/:table/:policy", Handler: deps.GetTableRLSStatus, Summary: "Get policies for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "PUT", Path: "/policies/:schema/:table/:policy", Handler: deps.UpdatePolicy, Summary: "Update RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "DELETE", Path: "/policies/:schema/:table/:policy", Handler: deps.DeletePolicy, Summary: "Delete RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		{Method: "GET", Path: "/policies/templates", Handler: deps.GetPolicyTemplates, Summary: "Get policy templates", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
		{Method: "GET", Path: "/security/warnings", Handler: deps.GetSecurityWarnings, Summary: "Get security warnings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
	}

	return &RouteGroup{
		Name:   "admin",
		Prefix: "/api/v1/admin",
		Routes: routes,
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.UnifiedAuth,
			Unified:  deps.UnifiedAuth,
		},
		RequireRole: deps.RequireRole,
	}
}
