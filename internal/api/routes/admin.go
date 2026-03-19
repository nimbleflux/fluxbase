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
	ListMembers            fiber.Handler
	AddMember              fiber.Handler
	UpdateMemberRole       fiber.Handler
	RemoveMember           fiber.Handler
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
}

func BuildAdminRoutes(deps *AdminDeps) *RouteGroup {
	auth := []Middleware{{Name: "UnifiedAuth", Handler: deps.UnifiedAuth}}

	routes := []Route{
		{Method: "GET", Path: "/tables", Handler: deps.GetTables, Summary: "List all tables", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "GET", Path: "/tables/:schema/:table", Handler: deps.GetTableSchema, Summary: "Get table schema", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "GET", Path: "/schemas", Handler: deps.GetSchemas, Summary: "List schemas", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/query", Handler: deps.ExecuteQuery, Summary: "Execute SQL query", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},

		{Method: "GET", Path: "/ddl/schemas", Handler: deps.ListSchemasDDL, Summary: "List schemas for DDL", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/ddl/schemas", Handler: deps.CreateSchemaDDL, Summary: "Create schema", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "GET", Path: "/ddl/tables", Handler: deps.ListTablesDDL, Summary: "List tables for DDL", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/ddl/tables", Handler: deps.CreateTableDDL, Summary: "Create table", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/ddl/tables/:schema/:table", Handler: deps.DeleteTableDDL, Summary: "Delete table", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/schemas", Handler: deps.CreateSchemaDDL, Summary: "Create schema (legacy)", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/tables", Handler: deps.CreateTableDDL, Summary: "Create table (legacy)", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/tables/:schema/:table", Handler: deps.DeleteTableDDL, Summary: "Delete table (legacy)", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "PATCH", Path: "/tables/:schema/:table", Handler: deps.RenameTableDDL, Summary: "Rename table", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "POST", Path: "/tables/:schema/:table/columns", Handler: deps.AddColumnDDL, Summary: "Add column", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/tables/:schema/:table/columns/:column", Handler: deps.DropColumnDDL, Summary: "Drop column", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},

		{Method: "POST", Path: "/realtime/tables", Handler: deps.EnableRealtime, Summary: "Enable realtime for table", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "GET", Path: "/realtime/tables", Handler: deps.ListRealtimeTables, Summary: "List realtime tables", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "GET", Path: "/realtime/tables/:schema/:table", Handler: deps.GetRealtimeStatus, Summary: "Get realtime status", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "PATCH", Path: "/realtime/tables/:schema/:table", Handler: deps.UpdateRealtimeConfig, Summary: "Update realtime config", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/realtime/tables/:schema/:table", Handler: deps.DisableRealtime, Summary: "Disable realtime for table", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},

		{Method: "GET", Path: "/oauth/providers", Handler: deps.ListOAuthProviders, Summary: "List OAuth providers", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "GET", Path: "/oauth/providers/:id", Handler: deps.GetOAuthProvider, Summary: "Get OAuth provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "POST", Path: "/oauth/providers", Handler: deps.CreateOAuthProvider, Summary: "Create OAuth provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "PUT", Path: "/oauth/providers/:id", Handler: deps.UpdateOAuthProvider, Summary: "Update OAuth provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "DELETE", Path: "/oauth/providers/:id", Handler: deps.DeleteOAuthProvider, Summary: "Delete OAuth provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},

		{Method: "GET", Path: "/saml/providers", Handler: deps.ListSAMLProviders, Summary: "List SAML providers", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "GET", Path: "/saml/providers/:id", Handler: deps.GetSAMLProvider, Summary: "Get SAML provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "POST", Path: "/saml/providers", Handler: deps.CreateSAMLProvider, Summary: "Create SAML provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "PUT", Path: "/saml/providers/:id", Handler: deps.UpdateSAMLProvider, Summary: "Update SAML provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "DELETE", Path: "/saml/providers/:id", Handler: deps.DeleteSAMLProvider, Summary: "Delete SAML provider", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "POST", Path: "/saml/validate-metadata", Handler: deps.ValidateSAML, Summary: "Validate SAML metadata", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "POST", Path: "/saml/upload-metadata", Handler: deps.UploadSAMLMetadata, Summary: "Upload SAML metadata", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},

		{Method: "GET", Path: "/auth/settings", Handler: deps.GetAuthSettings, Summary: "Get auth settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "PUT", Path: "/auth/settings", Handler: deps.UpdateAuthSettings, Summary: "Update auth settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "GET", Path: "/auth/sessions", Handler: deps.ListSessions, Summary: "List sessions", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/auth/sessions/:id", Handler: deps.RevokeSession, Summary: "Revoke session", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "DELETE", Path: "/auth/sessions/user/:user_id", Handler: deps.RevokeUserSessions, Summary: "Revoke user sessions", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},

		{Method: "GET", Path: "/system/settings", Handler: deps.ListSystemSettings, Summary: "List system settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "GET", Path: "/system/settings/*", Handler: deps.GetSystemSetting, Summary: "Get system setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "PUT", Path: "/system/settings/*", Handler: deps.UpdateSystemSetting, Summary: "Update system setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},
		{Method: "DELETE", Path: "/system/settings/*", Handler: deps.DeleteSystemSetting, Summary: "Delete system setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin")})},

		{Method: "POST", Path: "/settings/custom", Handler: deps.CreateCustomSetting, Summary: "Create custom setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "GET", Path: "/settings/custom", Handler: deps.ListCustomSettings, Summary: "List custom settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "POST", Path: "/settings/custom/secret", Handler: deps.CreateSecretSetting, Summary: "Create secret setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "GET", Path: "/settings/custom/secrets", Handler: deps.ListSecretSettings, Summary: "List secret settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "GET", Path: "/settings/custom/secret/*", Handler: deps.GetSecretSetting, Summary: "Get secret setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "PUT", Path: "/settings/custom/secret/*", Handler: deps.UpdateSecretSetting, Summary: "Update secret setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "DELETE", Path: "/settings/custom/secret/*", Handler: deps.DeleteSecretSetting, Summary: "Delete secret setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "GET", Path: "/settings/user/:user_id/secret/:key/decrypt", Handler: deps.GetUserSecretValue, Summary: "Decrypt user secret (service_role only)", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("service_role")})},
		{Method: "GET", Path: "/settings/custom/*", Handler: deps.GetCustomSetting, Summary: "Get custom setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "PUT", Path: "/settings/custom/*", Handler: deps.UpdateCustomSetting, Summary: "Update custom setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},
		{Method: "DELETE", Path: "/settings/custom/*", Handler: deps.DeleteCustomSetting, Summary: "Delete custom setting", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin", "service_role")})},

		{Method: "GET", Path: "/app/settings", Handler: deps.GetAppSettings, Summary: "Get app settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
		{Method: "PUT", Path: "/app/settings", Handler: deps.UpdateAppSettings, Summary: "Update app settings", Auth: AuthRequired, Middlewares: append(auth, Middleware{Name: "RequireRole", Handler: deps.RequireRole("admin", "instance_admin", "tenant_admin")})},
	}

	return &RouteGroup{Name: "admin", Prefix: "/api/v1/admin", Routes: routes}
}
