package routes

import (
	"github.com/gofiber/fiber/v3"
)

// SchemaAdminDeps contains dependencies for schema admin routes.
// Auth middleware is inherited from the parent admin route group.
//
// Role Access:
//   - instance_admin: Full access to all schema operations (bypasses RLS)
//   - tenant_admin: Access to own tenant's schema operations (RLS enforced)
type SchemaAdminDeps struct {
	GetTables               fiber.Handler
	GetTableSchema          fiber.Handler
	GetSchemas              fiber.Handler
	ExecuteQuery            fiber.Handler
	ListSchemasDDL          fiber.Handler
	CreateSchemaDDL         fiber.Handler
	ListTablesDDL           fiber.Handler
	CreateTableDDL          fiber.Handler
	DeleteTableDDL          fiber.Handler
	RenameTableDDL          fiber.Handler
	AddColumnDDL            fiber.Handler
	DropColumnDDL           fiber.Handler
	EnableRealtime          fiber.Handler
	ListRealtimeTables      fiber.Handler
	GetRealtimeStatus       fiber.Handler
	UpdateRealtimeConfig    fiber.Handler
	DisableRealtime         fiber.Handler
	ExecuteSQL              fiber.Handler
	ExportTypeScript        fiber.Handler
	RefreshSchema           fiber.Handler
	GetSchemaGraph          fiber.Handler
	GetTableRelationships   fiber.Handler
	GetTablesWithRLS        fiber.Handler
	GetTableRLSStatus       fiber.Handler
	ToggleTableRLS          fiber.Handler
	ListPolicies            fiber.Handler
	CreatePolicy            fiber.Handler
	UpdatePolicy            fiber.Handler
	DeletePolicy            fiber.Handler
	GetPolicyTemplates      fiber.Handler
	GetSecurityWarnings     fiber.Handler
	DumpInternalSchema      fiber.Handler
	PlanInternalSchema      fiber.Handler
	ApplyInternalSchema     fiber.Handler
	ValidateInternalSchema  fiber.Handler
	GetInternalSchemaStatus fiber.Handler
	MigrateInternalSchema   fiber.Handler
}

// BuildSchemaAdminRoutes creates the schema admin route group.
func BuildSchemaAdminRoutes(deps *SchemaAdminDeps) *RouteGroup {
	if deps == nil {
		return nil
	}

	return &RouteGroup{
		Name: "schema_admin",
		Routes: []Route{
			// Tables
			{Method: "GET", Path: "/tables", Handler: deps.GetTables, Summary: "List all tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/tables/:schema/:table", Handler: deps.GetTableSchema, Summary: "Get table schema", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/schemas", Handler: deps.GetSchemas, Summary: "List schemas", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/query", Handler: deps.ExecuteQuery, Summary: "Execute SQL query", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// DDL
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

			// Realtime
			{Method: "POST", Path: "/realtime/tables", Handler: deps.EnableRealtime, Summary: "Enable realtime for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/realtime/tables", Handler: deps.ListRealtimeTables, Summary: "List realtime tables", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/realtime/tables/:schema/:table", Handler: deps.GetRealtimeStatus, Summary: "Get realtime status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PATCH", Path: "/realtime/tables/:schema/:table", Handler: deps.UpdateRealtimeConfig, Summary: "Update realtime config", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "DELETE", Path: "/realtime/tables/:schema/:table", Handler: deps.DisableRealtime, Summary: "Disable realtime for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// SQL & Schema Export
			{Method: "POST", Path: "/sql", Handler: deps.ExecuteSQL, Summary: "Execute SQL", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/schema/export/typescript", Handler: deps.ExportTypeScript, Summary: "Export TypeScript types", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/schema/refresh", Handler: deps.RefreshSchema, Summary: "Refresh schema cache", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Schema Graph & Relationships
			{Method: "GET", Path: "/schema/graph", Handler: deps.GetSchemaGraph, Summary: "Get schema graph", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/tables/:schema/:table/relationships", Handler: deps.GetTableRelationships, Summary: "Get table relationships", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// RLS
			{Method: "GET", Path: "/tables/rls", Handler: deps.GetTablesWithRLS, Summary: "Get tables with RLS status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/tables/:schema/:table/rls", Handler: deps.GetTableRLSStatus, Summary: "Get table RLS status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/tables/:schema/:table/rls/toggle", Handler: deps.ToggleTableRLS, Summary: "Toggle table RLS", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},

			// Policies
			{Method: "GET", Path: "/policies", Handler: deps.ListPolicies, Summary: "List RLS policies", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "POST", Path: "/policies", Handler: deps.CreatePolicy, Summary: "Create RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/policies/:schema/:table/:policy", Handler: deps.GetTableRLSStatus, Summary: "Get policies for table", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "PUT", Path: "/policies/:schema/:table/:policy", Handler: deps.UpdatePolicy, Summary: "Update RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "DELETE", Path: "/policies/:schema/:table/:policy", Handler: deps.DeletePolicy, Summary: "Delete RLS policy", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/policies/templates", Handler: deps.GetPolicyTemplates, Summary: "Get policy templates", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},
			{Method: "GET", Path: "/security/warnings", Handler: deps.GetSecurityWarnings, Summary: "Get security warnings", Auth: AuthRequired, Roles: []string{"admin", "instance_admin", "tenant_admin"}},

			// Internal Schema (Declarative) - instance admin only
			{Method: "POST", Path: "/internal-schema/dump", Handler: deps.DumpInternalSchema, Summary: "Dump internal schema to SQL", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/internal-schema/plan", Handler: deps.PlanInternalSchema, Summary: "Plan schema changes", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/internal-schema/apply", Handler: deps.ApplyInternalSchema, Summary: "Apply schema changes", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/internal-schema/validate", Handler: deps.ValidateInternalSchema, Summary: "Validate schema for drift", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "GET", Path: "/internal-schema/status", Handler: deps.GetInternalSchemaStatus, Summary: "Get schema status", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
			{Method: "POST", Path: "/internal-schema/migrate", Handler: deps.MigrateInternalSchema, Summary: "Migrate from imperative to declarative", Auth: AuthRequired, Roles: []string{"admin", "instance_admin"}},
		},
	}
}
