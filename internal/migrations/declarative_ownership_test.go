package migrations

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbschema "github.com/nimbleflux/fluxbase/internal/database/schema"
)

// Tests for the declarative-plan ownership filter (filterUnownedDrops) and the
// destructive re-check (partitionDestructiveChanges).
//
// Regression context (v2026.9.3/.4 fresh-boot crash loop): pgschema's per-schema
// diff treats every live object in a schema as diffable desired state. Objects
// that Fluxbase manages OUTSIDE the declarative schema files — the ~45 RLS
// policies and set_tenant_id triggers applied by applyPostSchemaPolicies from
// post-schema.sql, and platform.app_schemas/app_schema_state created by the app
// declarative service — are therefore planned as DROPs on every boot after the
// first. With AllowDestructive=false (the startup default for platform) the
// destructive re-check added in #359 refuses to execute that plan and startup
// exits, crash-looping every fresh deployment. The ownership filter scopes drop
// planning to declared objects; these tests pin that behavior using the exact
// drop statements observed in the failing v2026.9.4 plan.
//
// Convention: testify, package migrations (white-box), DB-free.

// newOwnershipTestService returns a DeclarativeService whose SchemaDir points
// at the real embedded schema files, so tests exercise the production SQL.
func newOwnershipTestService(t *testing.T) *DeclarativeService {
	t.Helper()
	schemaDir, err := dbschema.ExtractSchemas()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(schemaDir) })
	return NewDeclarativeService("pgschema", "h", 1, "u", "p", "db", DeclarativeConfig{
		SchemaDir: schemaDir,
		Schemas:   DefaultFluxbaseSchemas,
	})
}

// The 47 destructive changes observed in the v2026.9.4 fresh-boot failure:
// 7 post-schema set_tenant_id triggers + 2 app-declarative tables + 38
// post-schema RLS policies. None of them are declared in platform.sql.
var freshBootUndeclaredDrops = []string{
	// DROP TRIGGER (path: schema.table.triggername)
	"DROP TRIGGER IF EXISTS platform_enabled_extensions_set_tenant_id ON enabled_extensions;",
	"DROP TRIGGER IF EXISTS platform_instance_settings_set_tenant_id ON instance_settings;",
	"DROP TRIGGER IF EXISTS platform_invitation_tokens_set_tenant_id ON invitation_tokens;",
	"DROP TRIGGER IF EXISTS platform_oauth_providers_set_tenant_id ON oauth_providers;",
	"DROP TRIGGER IF EXISTS platform_service_keys_set_tenant_id ON service_keys;",
	"DROP TRIGGER IF EXISTS platform_tenant_admin_assignments_set_tenant_id ON tenant_admin_assignments;",
	"DROP TRIGGER IF EXISTS platform_tenant_memberships_set_tenant_id ON tenant_memberships;",
	// DROP TABLE (path: schema.tablename — only 2 path parts)
	"DROP TABLE IF EXISTS app_schemas CASCADE;",
	"DROP TABLE IF EXISTS app_schema_state CASCADE;",
	// DROP POLICY (path: schema.table.policyname)
	"DROP POLICY IF EXISTS platform_activity_log_admin ON activity_log;",
	"DROP POLICY IF EXISTS platform_activity_log_service ON activity_log;",
	"DROP POLICY IF EXISTS platform_available_extensions_admin ON available_extensions;",
	"DROP POLICY IF EXISTS platform_available_extensions_service ON available_extensions;",
	"DROP POLICY IF EXISTS platform_email_verification_tokens_admin ON email_verification_tokens;",
	"DROP POLICY IF EXISTS platform_email_verification_tokens_service ON email_verification_tokens;",
	"DROP POLICY IF EXISTS platform_enabled_extensions_admin ON enabled_extensions;",
	"DROP POLICY IF EXISTS platform_enabled_extensions_service ON enabled_extensions;",
	"DROP POLICY IF EXISTS platform_enabled_extensions_tenant ON enabled_extensions;",
	"DROP POLICY IF EXISTS instance_settings_delete_tenant ON instance_settings;",
	"DROP POLICY IF EXISTS instance_settings_insert ON instance_settings;",
	"DROP POLICY IF EXISTS instance_settings_insert_tenant ON instance_settings;",
	"DROP POLICY IF EXISTS instance_settings_select ON instance_settings;",
	"DROP POLICY IF EXISTS instance_settings_update ON instance_settings;",
	"DROP POLICY IF EXISTS platform_invitation_tokens_admin ON invitation_tokens;",
	"DROP POLICY IF EXISTS platform_invitation_tokens_service ON invitation_tokens;",
	"DROP POLICY IF EXISTS platform_invitation_tokens_tenant ON invitation_tokens;",
	"DROP POLICY IF EXISTS platform_key_usage_service ON key_usage;",
	"DROP POLICY IF EXISTS platform_oauth_providers_tenant ON oauth_providers;",
	"DROP POLICY IF EXISTS platform_password_reset_tokens_admin ON password_reset_tokens;",
	"DROP POLICY IF EXISTS platform_password_reset_tokens_service ON password_reset_tokens;",
	"DROP POLICY IF EXISTS platform_schema_migrations_service ON schema_migrations;",
	"DROP POLICY IF EXISTS platform_service_keys_admin ON service_keys;",
	"DROP POLICY IF EXISTS platform_service_keys_service ON service_keys;",
	"DROP POLICY IF EXISTS platform_service_keys_tenant ON service_keys;",
	"DROP POLICY IF EXISTS platform_sessions_admin ON sessions;",
	"DROP POLICY IF EXISTS sso_identities_admin ON sso_identities;",
	"DROP POLICY IF EXISTS platform_tenant_admin_assignments_all ON tenant_admin_assignments;",
	"DROP POLICY IF EXISTS platform_tenant_admin_assignments_self ON tenant_admin_assignments;",
	"DROP POLICY IF EXISTS platform_tenant_admin_assignments_tenant ON tenant_admin_assignments;",
	"DROP POLICY IF EXISTS platform_tenant_memberships_admin ON tenant_memberships;",
	"DROP POLICY IF EXISTS platform_tenant_memberships_self ON tenant_memberships;",
	"DROP POLICY IF EXISTS platform_tenant_memberships_service ON tenant_memberships;",
	"DROP POLICY IF EXISTS platform_tenant_memberships_tenant ON tenant_memberships;",
	"DROP POLICY IF EXISTS platform_tenants_assigned ON tenants;",
	"DROP POLICY IF EXISTS platform_tenants_instance_admin ON tenants;",
	"DROP POLICY IF EXISTS platform_users_all ON users;",
	"DROP POLICY IF EXISTS platform_users_self ON users;",
}

// freshBootPlanFixture builds a plan shaped like pgschema's second-boot plan for
// schema "platform" (the one that failed startup on v2026.9.4): the 47
// undeclared-object drops, the 4 cross-schema FK constraint drops managed by
// post-schema-fks.sql, plus ordinary create steps and the default-privilege
// normalization steps pgschema emits on every plan.
func freshBootPlanFixture(t *testing.T) *Plan {
	t.Helper()
	steps := []PlanStep{
		{SQL: "CREATE TABLE IF NOT EXISTS available_extensions (\n    id uuid DEFAULT gen_random_uuid()\n);", Operation: "create", Path: "platform.table.available_extensions"},
		{SQL: "ALTER TABLE available_extensions ENABLE ROW LEVEL SECURITY;", Operation: "alter", Path: "platform.table.available_extensions"},
	}
	for _, sql := range freshBootUndeclaredDrops {
		steps = append(steps, PlanStep{SQL: sql, Operation: "drop", Path: "platform.declarative.object"})
	}
	// Cross-schema FK drops, managed by post-schema-fks.sql (filtered by
	// filterManagedFKDrops in the apply pipeline).
	steps = append(steps,
		PlanStep{SQL: "ALTER TABLE instance_settings DROP CONSTRAINT instance_settings_tenant_id_fkey;", Operation: "drop", Path: "platform.instance_settings.instance_settings_tenant_id_fkey"},
		PlanStep{SQL: "ALTER TABLE migrations DROP CONSTRAINT app_applied_by_fkey;", Operation: "drop", Path: "platform.migrations.app_applied_by_fkey"},
	)
	// Default-privilege normalization pgschema flags destructive on every plan.
	steps = append(steps, PlanStep{
		SQL:       "ALTER DEFAULT PRIVILEGES FOR ROLE fluxbase IN SCHEMA platform REVOKE SELECT, UPDATE, USAGE ON SEQUENCES FROM service_role;",
		Operation: "drop",
		Path:      "default_privileges.fluxbase.SEQUENCES.service_role",
	})
	return &Plan{Groups: []PlanGroup{{Steps: steps}}}
}

// TestFilterUnownedDrops_FreshBootPlanHasNoDestructiveChanges is the regression
// pin: the full apply pipeline (ownership filter -> managed-FK filter ->
// destructive re-check with allow_destructive=false) must yield zero blocked
// changes for the plan that crash-looped every fresh v2026.9.4 deployment.
func TestFilterUnownedDrops_FreshBootPlanHasNoDestructiveChanges(t *testing.T) {
	t.Parallel()
	svc := newOwnershipTestService(t)

	plan := freshBootPlanFixture(t)
	svc.filterUnownedDrops("platform", plan)
	plan.Changes = filterManagedFKDropsForTest(svc, extractChangesFromGroups(plan))
	_, blocked := partitionDestructiveChanges(plan.Changes, false)

	assert.Empty(t, blocked,
		"fresh-boot platform plan must not contain blocked destructive changes; got %d: %v",
		len(blocked), blocked)
	// The ordinary create steps must survive the filtering.
	assert.Len(t, plan.Changes, 3, "creates, alters and privilege normalization must survive the ownership filter")
}

// TestFilterUnownedDrops_RemovesExactlyTheUndeclaredDrops verifies the filter
// removes precisely the 47 undeclared drops (and nothing else) and keeps
// plan.Groups and plan.Changes consistent.
func TestFilterUnownedDrops_RemovesExactlyTheUndeclaredDrops(t *testing.T) {
	t.Parallel()
	svc := newOwnershipTestService(t)

	plan := freshBootPlanFixture(t)
	totalDrops := 0
	for _, g := range plan.Groups {
		for _, s := range g.Steps {
			if s.Operation == "drop" {
				totalDrops++
			}
		}
	}
	require.Equal(t, len(freshBootUndeclaredDrops)+3, totalDrops, "fixture drop count changed")

	svc.filterUnownedDrops("platform", plan)

	var remainingDrops []string
	for _, g := range plan.Groups {
		for _, s := range g.Steps {
			if s.Operation == "drop" {
				remainingDrops = append(remainingDrops, s.SQL)
			}
		}
	}
	// Only the default-privilege normalization drop and the 2 FK constraint
	// drops on declared tables remain (the FKs are removed later by
	// filterManagedFKDrops); the 47 undeclared-object drops must be gone.
	require.Len(t, remainingDrops, 3)
	for _, sql := range remainingDrops {
		assert.True(t, isPrivilegeNormalization(sql) || strings.Contains(strings.ToUpper(sql), "DROP CONSTRAINT"),
			"unexpected surviving drop: %s", sql)
	}
	assert.Len(t, plan.Changes, 5, "Changes must be rebuilt from the filtered groups")
}

// TestPlatformSchemaDoesNotDeclareExternallyOwnedObjects guards the declared-
// object parse of the production platform.sql: none of the objects owned by
// bootstrap, post-schema or the app declarative service may appear declared,
// and the core platform tables must be declared.
func TestPlatformSchemaDoesNotDeclareExternallyOwnedObjects(t *testing.T) {
	t.Parallel()
	svc := newOwnershipTestService(t)
	objs, err := declaredSchemaObjects(filepath.Join(svc.config.SchemaDir, "platform.sql"))
	require.NoError(t, err)

	for _, name := range freshBootUndeclaredDrops {
		obj := firstIdentifier(name)
		assert.False(t, objs.tables[obj] || objs.policies[obj] || objs.triggers[obj],
			"platform.sql must not declare externally-owned object %q", obj)
	}
	// Core platform tables (including bootstrap-owned bootstrap_state/
	// declarative_state, which platform.sql DOES declare) must parse.
	for _, table := range []string{"users", "tenants", "instance_settings", "bootstrap_state", "declarative_state"} {
		assert.Truef(t, objs.tables[table], "platform.sql should declare table %q", table)
	}
	// Objects platform.sql really declares must be recognized so their drops
	// (if ever planned) stay governed by allow_destructive.
	for _, policy := range []string{"platform_email_templates_read", "instance_settings_delete", "platform_oauth_providers_read"} {
		assert.Truef(t, objs.policies[policy], "platform.sql should declare policy %q", policy)
	}
	for _, trigger := range []string{"instance_settings_updated_at", "platform_tenants_updated_at", "update_platform_users_updated_at"} {
		assert.Truef(t, objs.triggers[trigger], "platform.sql should declare trigger %q", trigger)
	}
}

// TestDeclaredSchemaObjects_AllEmbeddedFilesDeclareTables guards the parse
// regexes against rot: every embedded per-schema file must yield at least one
// declared table. (DefaultFluxbaseSchemas also lists schemas without files,
// e.g. "system"; those are skipped here.)
func TestDeclaredSchemaObjects_AllEmbeddedFilesDeclareTables(t *testing.T) {
	t.Parallel()
	svc := newOwnershipTestService(t)
	checked := 0
	for _, schema := range DefaultFluxbaseSchemas {
		path := filepath.Join(svc.config.SchemaDir, schema+".sql")
		if _, err := os.Stat(path); err != nil {
			continue
		}
		objs, err := declaredSchemaObjects(path)
		require.NoErrorf(t, err, "schema %s", schema)
		assert.NotEmptyf(t, objs.tables, "schema %s must declare at least one table", schema)
		checked++
	}
	assert.GreaterOrEqual(t, checked, 10, "expected most DefaultFluxbaseSchemas entries to have schema files")
}

// TestFilterUnownedDrops_KeepsDeclaredObjectDrops verifies that drops of
// objects the schema file DOES declare survive the filter (their execution
// remains governed by allow_destructive, as before).
func TestFilterUnownedDrops_KeepsDeclaredObjectDrops(t *testing.T) {
	t.Parallel()
	svc := newOwnershipTestService(t)
	plan := &Plan{Groups: []PlanGroup{{Steps: []PlanStep{
		{SQL: "DROP TABLE IF EXISTS email_templates CASCADE;", Operation: "drop", Path: "platform.table.email_templates"},
		{SQL: "DROP POLICY IF EXISTS platform_email_templates_read ON email_templates;", Operation: "drop", Path: "platform.email_templates.platform_email_templates_read"},
		{SQL: "DROP TRIGGER IF EXISTS instance_settings_updated_at ON instance_settings;", Operation: "drop", Path: "platform.instance_settings.instance_settings_updated_at"},
		{SQL: "DROP INDEX IF EXISTS idx_instance_settings_settings;", Operation: "drop", Path: "platform.instance_settings.idx_instance_settings_settings"},
		{SQL: "ALTER TABLE instance_settings DROP COLUMN overridable_settings;", Operation: "drop", Path: "platform.instance_settings.instance_settings"},
	}}}}

	svc.filterUnownedDrops("platform", plan)

	var kept []string
	for _, g := range plan.Groups {
		for _, s := range g.Steps {
			kept = append(kept, s.SQL)
		}
	}
	assert.Len(t, kept, 5, "drops of declared objects must survive the ownership filter")
	assert.Len(t, plan.Changes, 5)
}

// TestFilterUnownedDrops_FailOpenWhenSchemaFileMissing verifies the filter
// never widens the plan: if the schema file cannot be read, all drops are kept.
func TestFilterUnownedDrops_FailOpenWhenSchemaFileMissing(t *testing.T) {
	t.Parallel()
	svc := NewDeclarativeService("pgschema", "h", 1, "u", "p", "db", DeclarativeConfig{
		SchemaDir: t.TempDir(),
	})
	plan := &Plan{Groups: []PlanGroup{{Steps: []PlanStep{
		{SQL: "DROP TABLE IF EXISTS app_schemas CASCADE;", Operation: "drop", Path: "platform.table.app_schemas"},
	}}}}
	plan.Changes = extractChangesFromGroups(plan)

	svc.filterUnownedDrops("platform", plan)

	require.Len(t, plan.Changes, 1, "drops must be kept when the schema file is unreadable")
}

// TestPartitionDestructiveChanges_ConstraintReplacementAllowed verifies that a
// same-plan DROP CONSTRAINT + ADD CONSTRAINT pair (e.g. widening the
// valid_module_path CHECK) executes even with allow_destructive=false, while a
// constraint drop the plan does not re-add stays blocked.
func TestPartitionDestructiveChanges_ConstraintReplacementAllowed(t *testing.T) {
	t.Parallel()
	drop := Change{
		Type: ChangeDrop, Destructive: true,
		SQL: "ALTER TABLE shared_modules DROP CONSTRAINT valid_module_path;",
	}
	add := Change{
		Type: ChangeCreate,
		SQL:  "ALTER TABLE functions.shared_modules\nADD CONSTRAINT valid_module_path CHECK (module_path ~ '^_shared/[a-zA-Z0-9_/.-]+\\.(ts|js|mts|mjs)$'::text AND module_path !~~ '%/../%'::text);",
	}
	loneDrop := Change{
		Type: ChangeDrop, Destructive: true,
		SQL: "ALTER TABLE shared_modules DROP CONSTRAINT some_removed_constraint;",
	}

	executable, blocked := partitionDestructiveChanges([]Change{drop, add}, false)
	assert.Empty(t, blocked, "constraint replacement must execute with allow_destructive=false")
	require.Len(t, executable, 2)
	assert.Equal(t, drop.SQL, executable[0].SQL)
	assert.Equal(t, add.SQL, executable[1].SQL)

	// A constraint drop the plan does not re-add stays blocked.
	_, blocked = partitionDestructiveChanges([]Change{loneDrop}, false)
	assert.Len(t, blocked, 1, "lone constraint drop must stay blocked")

	// Table name normalization: schema-qualified ADD, unqualified DROP.
	dropQ := Change{Type: ChangeDrop, Destructive: true, SQL: "ALTER TABLE functions.edge_files DROP CONSTRAINT valid_file_path;"}
	addQ := Change{Type: ChangeCreate, SQL: "ALTER TABLE functions.edge_files\nADD CONSTRAINT valid_file_path CHECK (file_path ~ '^x$');"}
	executable, blocked = partitionDestructiveChanges([]Change{dropQ, addQ}, false)
	assert.Empty(t, blocked)
	assert.Len(t, executable, 2)

	// With allow_destructive=true nothing is blocked anyway.
	_, blocked = partitionDestructiveChanges([]Change{drop, add, loneDrop}, true)
	assert.Empty(t, blocked)
}

// filterManagedFKDropsForTest is a thin wrapper so tests can call the
// unexported filter without re-implementing it.
func filterManagedFKDropsForTest(svc *DeclarativeService, changes []Change) []Change {
	return svc.filterManagedFKDrops(changes)
}

// firstIdentifier extracts the object name from a DROP statement for test
// assertions (third identifier in DROP <KIND> [IF EXISTS] <name> ...).
func firstIdentifier(dropSQL string) string {
	for _, re := range []*regexp.Regexp{dropPolicyRe, dropTriggerRe, dropTableRe} {
		if name, ok := firstMatch(re, dropSQL); ok {
			return name
		}
	}
	return ""
}

// The schema "ai" ships explicit REVOKE statements whose privilege state is
// re-created by bootstrap on every boot. Partition must classify them as
// privilege normalization (executable under allow_destructive=false): the
// observed v2026.9.4 failure was exactly these three REVOKEs re-planned on
// boot #2 as a destructive-ONLY plan, which the re-check turned into a
// startup crash. Also pins the mixed-plan case: executable changes proceed,
// genuinely destructive drops stay blocked.
func TestPartitionDestructiveChanges_PrivilegeRevokesExecutable(t *testing.T) {
	changes := []Change{
		{SQL: "REVOKE SELECT ON TABLE tool_audit_log FROM authenticated", Destructive: true},
		{SQL: "REVOKE DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE tool_integrations FROM service_role", Destructive: true},
		{SQL: "REVOKE ALL ON FUNCTION get_embedding_stats(uuid) FROM authenticated", Destructive: true},
		{SQL: "DROP TABLE IF EXISTS legacy_table CASCADE", Destructive: true},
	}

	executable, blocked := partitionDestructiveChanges(changes, false)

	assert.Len(t, executable, 3, "REVOKEs are privilege normalization and must execute")
	assert.Len(t, blocked, 1, "genuine destructive drops stay blocked")
	assert.Equal(t, "DROP TABLE IF EXISTS legacy_table CASCADE", blocked[0].SQL)
}

func TestIsPrivilegeNormalization(t *testing.T) {
	assert.True(t, isPrivilegeNormalization("ALTER DEFAULT PRIVILEGES FOR ROLE fluxbase IN SCHEMA auth REVOKE SELECT ON SEQUENCES FROM service_role"))
	assert.True(t, isPrivilegeNormalization("  revoke select on table tool_audit_log from authenticated"))
	assert.True(t, isPrivilegeNormalization("REVOKE ALL ON FUNCTION get_embedding_stats(uuid) FROM authenticated"))
	assert.False(t, isPrivilegeNormalization("DROP TABLE IF EXISTS app_schemas CASCADE"))
	assert.False(t, isPrivilegeNormalization("DELETE FROM users"))
	assert.False(t, isPrivilegeNormalization("TRUNCATE sessions"))
}
