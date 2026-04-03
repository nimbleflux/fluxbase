package routes

import (
	"github.com/gofiber/fiber/v3"
)

// AdminDeps contains dependencies for admin routes.
// Auth middleware is inherited by all subgroups.
//
// Role Access:
//   - instance_admin: Full access to all admin operations (bypasses RLS)
//   - tenant_admin: Access to tenant-scoped operations (RLS enforced)
//   - service_role: Full programmatic access
//
// Subgroups:
//   - branch: Database branching operations
//   - schema: Tables, DDL, realtime, RLS, SQL execution
//   - auth_providers: OAuth/SAML providers, auth settings, sessions
//   - users: User management, invitations
//   - tenants: Tenant CRUD, settings, members
//   - service_keys: API key management
//   - functions: Edge functions management
//   - jobs: Background jobs management
//   - ai: AI chatbots, knowledge bases, providers
//   - rpc: RPC procedures
//   - logs: Log viewing and management
//   - settings: System, custom, email, instance settings
//   - extensions: PostgreSQL extensions
type AdminDeps struct {
	UnifiedAuth fiber.Handler
	RequireRole func(...string) fiber.Handler

	// Middleware for tenant context (inherited by all subgroups)
	TenantMiddleware   fiber.Handler
	TenantDBMiddleware fiber.Handler

	// RequireExplicitTenant rejects default-tenant fallback on tenant-scoped subgroups
	RequireExplicitTenant fiber.Handler

	// Subgroup dependencies
	Branch           *BranchDeps
	Schema           *SchemaAdminDeps
	AuthProviders    *AuthProvidersAdminDeps
	Users            *UsersAdminDeps
	Tenants          *TenantsAdminDeps
	ServiceKeys      *ServiceKeysAdminDeps
	Functions        *FunctionsAdminDeps
	Jobs             *JobsAdminDeps
	AI               *AIAdminDeps
	RPC              *RPCAdminDeps
	Logs             *LogsAdminDeps
	Settings         *SettingsAdminDeps
	Extensions       *ExtensionsAdminDeps
	ExtensionsTenant *ExtensionsTenantDeps
}

// BuildAdminRoutes creates the admin route group with proper role-based access control.
//
// Role Hierarchy:
//   - instance_admin: Global admin with access to ALL tenants and instance-level settings.
//     Bypasses RLS policies for cross-tenant operations.
//   - tenant_admin:   Admin for specific assigned tenant(s). RLS policies enforce
//     data isolation to their tenant(s) only.
//
// Instance-Level Operations (instance_admin only):
//   - /instance/settings/*     - Global instance configuration
//   - /extensions/*            - PostgreSQL extension management
//   - POST/DELETE /tenants     - Create/delete tenant databases
//   - /tenants/:id/admins      - Assign/remove tenant admins (instance_admin can manage all)
//   - /oauth/providers/*       - OAuth provider configuration
//   - /saml/providers/*        - SAML provider configuration
//   - /auth/settings           - Authentication settings
//   - /system/settings/*       - System-wide settings
//   - /email/settings/*        - Email provider configuration
//   - /settings/captcha        - CAPTCHA configuration
//   - /ai/providers            - AI provider configuration
//   - /ai/conversations/*      - All AI conversations (cross-tenant)
//   - /ai/audit                - AI audit logs
//   - /rpc/*                   - RPC procedure management
//   - /logs/*                  - Centralized logging
//
// Tenant-Scoped Operations (tenant_admin + instance_admin):
//   - /tables/*                - Schema introspection and DDL
//   - /realtime/*              - Realtime configuration
//   - /sql                     - SQL execution (RLS filtered)
//   - /schema/*                - Schema export and graph
//   - /auth/sessions/*         - Session management (own tenant)
//   - /users/*                 - User management (own tenant)
//   - /invitations/*           - Invitation management (own tenant)
//   - /tenants/:id             - View/update own tenant
//   - /tenants/:id/settings/*  - Own tenant settings
//   - /tenants/:id/members/*   - Own tenant member management
//   - /functions/executions/*  - Function execution logs
//   - /jobs                    - View jobs (own tenant)
//   - /ai/chatbots/*           - Chatbot management (own tenant)
//   - /ai/tables/*             - AI table export
//   - /settings/custom/*       - Custom settings (own tenant)
//   - /app/settings            - App settings (own tenant)
//   - /branches/*              - Branch operations (own tenant)
//
// Auth Middleware Location: internal/middleware/tenant.go
// RLS Context Setup: internal/middleware/rls.go
// RequireRole Middleware: internal/api/auth_middleware.go
func BuildAdminRoutes(deps *AdminDeps) *RouteGroup {
	var subgroups []*RouteGroup

	// Tenant-scoped subgroups: require explicit tenant selection (no default fallback)
	tenantScoped := map[string]bool{
		"branch":               true,
		"schema_admin":         true,
		"auth_providers_admin": true,
		"users_admin":          true,
		"service_keys_admin":   true,
		"functions_admin":      true,
		"jobs_admin":           true,
		"ai_admin":             true,
		"rpc_admin":            true,
		"logs_admin":           true,
		"extensions_tenant":    true,
	}

	// Helper to conditionally add RequireExplicitTenant to tenant-scoped subgroups
	applyExplicitTenant := func(sg *RouteGroup) *RouteGroup {
		if sg == nil || deps.RequireExplicitTenant == nil || !tenantScoped[sg.Name] {
			return sg
		}
		sg.Middlewares = append(sg.Middlewares, Middleware{
			Name:    "RequireExplicitTenant",
			Handler: deps.RequireExplicitTenant,
		})
		return sg
	}

	// Register subgroups
	if branch := BuildBranchRoutes(deps.Branch); branch != nil {
		subgroups = append(subgroups, applyExplicitTenant(branch))
	}
	if schema := BuildSchemaAdminRoutes(deps.Schema); schema != nil {
		subgroups = append(subgroups, applyExplicitTenant(schema))
	}
	if authProviders := BuildAuthProvidersAdminRoutes(deps.AuthProviders); authProviders != nil {
		subgroups = append(subgroups, applyExplicitTenant(authProviders))
	}
	if users := BuildUsersAdminRoutes(deps.Users); users != nil {
		subgroups = append(subgroups, applyExplicitTenant(users))
	}
	if tenants := BuildTenantsAdminRoutes(deps.Tenants); tenants != nil {
		subgroups = append(subgroups, tenants) // instance-level: no explicit tenant
	}
	if serviceKeys := BuildServiceKeysAdminRoutes(deps.ServiceKeys); serviceKeys != nil {
		subgroups = append(subgroups, applyExplicitTenant(serviceKeys))
	}
	if functions := BuildFunctionsAdminRoutes(deps.Functions); functions != nil {
		subgroups = append(subgroups, applyExplicitTenant(functions))
	}
	if jobs := BuildJobsAdminRoutes(deps.Jobs); jobs != nil {
		subgroups = append(subgroups, applyExplicitTenant(jobs))
	}
	if ai := BuildAIAdminRoutes(deps.AI); ai != nil {
		subgroups = append(subgroups, applyExplicitTenant(ai))
	}
	if rpc := BuildRPCAdminRoutes(deps.RPC); rpc != nil {
		subgroups = append(subgroups, applyExplicitTenant(rpc))
	}
	if logs := BuildLogsAdminRoutes(deps.Logs); logs != nil {
		subgroups = append(subgroups, applyExplicitTenant(logs))
	}
	if settings := BuildSettingsAdminRoutes(deps.Settings); settings != nil {
		subgroups = append(subgroups, settings) // mixed: instance + tenant routes, no explicit tenant
	}
	if extensions := BuildExtensionsAdminRoutes(deps.Extensions); extensions != nil {
		subgroups = append(subgroups, extensions) // instance-level: no explicit tenant
	}
	if extensionsTenant := BuildExtensionsTenantRoutes(deps.ExtensionsTenant); extensionsTenant != nil {
		subgroups = append(subgroups, applyExplicitTenant(extensionsTenant))
	}

	// Build middlewares for tenant context (inherited by all subgroups)
	var middlewares []Middleware
	if deps.TenantMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name:    "TenantContext",
			Handler: deps.TenantMiddleware,
		})
	}
	if deps.TenantDBMiddleware != nil {
		middlewares = append(middlewares, Middleware{
			Name:    "TenantDBContext",
			Handler: deps.TenantDBMiddleware,
		})
	}

	return &RouteGroup{
		Name:      "admin",
		Prefix:    "/api/v1/admin",
		Routes:    []Route{}, // All routes are now in subgroups
		SubGroups: subgroups,
		AuthMiddlewares: &AuthMiddlewares{
			Required: deps.UnifiedAuth,
			Unified:  deps.UnifiedAuth,
		},
		Middlewares: middlewares,
		RequireRole: deps.RequireRole,
	}
}
