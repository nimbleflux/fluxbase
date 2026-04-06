//go:build integration

package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

// tenantTestID1 and tenantTestID2 are stable UUIDs used for tenant isolation tests.
// These are distinct from the ones in has_tenant_access_test.go to avoid conflicts.
const (
	tenantTestID1 = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantTestID2 = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
)

// setupTenantIsolationTest creates a test context with clean tenant data for isolation tests.
func setupTenantIsolationTest(t *testing.T) *test.TestContext {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()

	// Clean up test data from previous runs.
	// Each DELETE is a separate call because pgx doesn't support
	// multiple statements in a single parameterized Exec.
	tc.ExecuteSQLAsSuperuser(`DELETE FROM auth.users WHERE email LIKE 'tenant-test-%'`)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM auth.sessions WHERE user_id NOT IN (SELECT id FROM auth.users)`)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM auth.service_keys WHERE name LIKE 'tenant-test-%'`)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM logging.entries WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM branching.branches WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM functions.edge_functions WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM ai.knowledge_bases WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM ai.documents WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM jobs.queue WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM rpc.procedures WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM realtime.schema_registry WHERE tenant_id IN ($1::uuid, $2::uuid)`, tenantTestID1, tenantTestID2)

	// Ensure tenant records exist in platform.tenants (needed for FK constraints).
	// Uses ON CONFLICT DO NOTHING to be idempotent.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, status)
		VALUES ($1, 'test-tenant-1', 'Test Tenant 1', 'active')
		ON CONFLICT (id) DO NOTHING
	`, tenantTestID1)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, status)
		VALUES ($1, 'test-tenant-2', 'Test Tenant 2', 'active')
		ON CONFLICT (id) DO NOTHING
	`, tenantTestID2)

	return tc
}

// ============================================================================
// AUTH SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_AuthUsers verifies that auth.users are isolated by tenant_id.
func TestTenantIsolation_AuthUsers(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert users for two different tenants as superuser
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, tenant_id, created_at)
		VALUES (gen_random_uuid(), 'tenant-test-user1@example.com', 'hash1', true, $1, NOW())
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, tenant_id, created_at)
		VALUES (gen_random_uuid(), 'tenant-test-user2@example.com', 'hash2', true, $1, NOW())
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's user
	tenant1Users := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT email FROM auth.users WHERE email LIKE 'tenant-test-%'`)

	require.Len(t, tenant1Users, 1, "Tenant1 should only see their own users")
	require.Equal(t, "tenant-test-user1@example.com", tenant1Users[0]["email"])

	// Verify as tenant2: should only see tenant2's user
	tenant2Users := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT email FROM auth.users WHERE email LIKE 'tenant-test-%'`)

	require.Len(t, tenant2Users, 1, "Tenant2 should only see their own users")
	require.Equal(t, "tenant-test-user2@example.com", tenant2Users[0]["email"])

	// Verify superuser sees both
	allUsers := tc.QuerySQLAsSuperuser(
		`SELECT email FROM auth.users WHERE email LIKE 'tenant-test-%' ORDER BY email`)
	require.Len(t, allUsers, 2, "Superuser should see all users")
}

// TestTenantIsolation_AuthServiceKeys verifies that service_keys are isolated by tenant_id.
func TestTenantIsolation_AuthServiceKeys(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert service keys for two tenants
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.service_keys (name, key_hash, key_prefix, tenant_id, key_type, enabled)
		VALUES ('tenant-test-key1', 'hash1', 'prefix1_', $1, 'service', true)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.service_keys (name, key_hash, key_prefix, tenant_id, key_type, enabled)
		VALUES ('tenant-test-key2', 'hash2', 'prefix2_', $1, 'service', true)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's key
	tenant1Keys := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM auth.service_keys WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant1Keys, 1, "Tenant1 should only see their own service keys")
	require.Equal(t, "tenant-test-key1", tenant1Keys[0]["name"])

	// Verify as tenant2: should only see tenant2's key
	tenant2Keys := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT name FROM auth.service_keys WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant2Keys, 1, "Tenant2 should only see their own service keys")
	require.Equal(t, "tenant-test-key2", tenant2Keys[0]["name"])
}

// ============================================================================
// LOGGING SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_LoggingEntries verifies that logging.entries are isolated by tenant_id.
func TestTenantIsolation_LoggingEntries(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert log entries for two tenants.
	// logging.entries requires category (partition key) and level, message (NOT NULL).
	// Column is "timestamp" not "created_at" (has DEFAULT now()).
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant1 log message', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant2 log message', $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's log
	tenant1Logs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT message FROM logging.entries WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	require.Len(t, tenant1Logs, 1, "Tenant1 should only see their own logs")
	require.Equal(t, "tenant1 log message", tenant1Logs[0]["message"])

	// Verify as tenant2: should only see tenant2's log
	tenant2Logs := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT message FROM logging.entries WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	require.Len(t, tenant2Logs, 1, "Tenant2 should only see their own logs")
	require.Equal(t, "tenant2 log message", tenant2Logs[0]["message"])
}

// ============================================================================
// BRANCHING SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_BranchingBranches verifies that branching.branches are isolated by tenant_id.
func TestTenantIsolation_BranchingBranches(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert branches for two tenants.
	// branching.branches requires name, slug, database_name (all NOT NULL).
	// Valid statuses: creating, ready, migrating, error, deleting, deleted.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO branching.branches (name, slug, database_name, status, tenant_id)
		VALUES ('tenant-test-branch1', 'tenant-test-branch1', 'branch_tenant1', 'ready', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO branching.branches (name, slug, database_name, status, tenant_id)
		VALUES ('tenant-test-branch2', 'tenant-test-branch2', 'branch_tenant2', 'ready', $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's branch
	tenant1Branches := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM branching.branches WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant1Branches, 1, "Tenant1 should only see their own branches")
	require.Equal(t, "tenant-test-branch1", tenant1Branches[0]["name"])

	// Verify as tenant2: should only see tenant2's branch
	tenant2Branches := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT name FROM branching.branches WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant2Branches, 1, "Tenant2 should only see their own branches")
	require.Equal(t, "tenant-test-branch2", tenant2Branches[0]["name"])
}

// ============================================================================
// FUNCTIONS SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_FunctionsEdgeFunctions verifies that functions.edge_functions are isolated by tenant_id.
func TestTenantIsolation_FunctionsEdgeFunctions(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert edge functions for two tenants.
	// functions.edge_functions requires name and code (NOT NULL).
	// Column is "code" not "entrypoint". Set is_public=false to avoid the
	// public_read policy which grants access regardless of tenant.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO functions.edge_functions (name, code, is_public, tenant_id)
		VALUES ('tenant-test-func1', 'export default function() { return "hello" }', false, $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO functions.edge_functions (name, code, is_public, tenant_id)
		VALUES ('tenant-test-func2', 'export default function() { return "hello" }', false, $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's function
	tenant1Funcs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM functions.edge_functions WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant1Funcs, 1, "Tenant1 should only see their own functions")
	require.Equal(t, "tenant-test-func1", tenant1Funcs[0]["name"])

	// Verify as tenant2: should only see tenant2's function
	tenant2Funcs := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT name FROM functions.edge_functions WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant2Funcs, 1, "Tenant2 should only see their own functions")
	require.Equal(t, "tenant-test-func2", tenant2Funcs[0]["name"])
}

// ============================================================================
// AI SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_AIKnowledgeBases verifies that ai.knowledge_bases are isolated by tenant_id.
func TestTenantIsolation_AIKnowledgeBases(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Create users for each tenant to own knowledge bases
	user1ID := "cccccccc-cccc-cccc-cccc-cccccccccccc"
	user2ID := "dddddddd-dddd-dddd-dddd-dddddddddddd"

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, tenant_id, created_at)
		VALUES ($1, 'tenant-test-owner1@example.com', 'hash1', true, $2, NOW())
	`, user1ID, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, tenant_id, created_at)
		VALUES ($1, 'tenant-test-owner2@example.com', 'hash2', true, $2, NOW())
	`, user2ID, tenantTestID2)

	// Insert knowledge bases for two tenants
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO ai.knowledge_bases (name, owner_id, visibility, tenant_id, created_at, updated_at)
		VALUES ('tenant-test-kb1', $1, 'private', $2, NOW(), NOW())
	`, user1ID, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO ai.knowledge_bases (name, owner_id, visibility, tenant_id, created_at, updated_at)
		VALUES ('tenant-test-kb2', $1, 'private', $2, NOW(), NOW())
	`, user2ID, tenantTestID2)

	// Verify as tenant1: should only see tenant1's knowledge base
	tenant1KBs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM ai.knowledge_bases WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant1KBs, 1, "Tenant1 should only see their own knowledge bases")
	require.Equal(t, "tenant-test-kb1", tenant1KBs[0]["name"])

	// Verify as tenant2: should only see tenant2's knowledge base
	tenant2KBs := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT name FROM ai.knowledge_bases WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant2KBs, 1, "Tenant2 should only see their own knowledge bases")
	require.Equal(t, "tenant-test-kb2", tenant2KBs[0]["name"])
}

// ============================================================================
// JOBS SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_JobsQueue verifies that jobs.queue are isolated by tenant_id.
func TestTenantIsolation_JobsQueue(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert job queue entries for two tenants.
	// jobs.queue requires namespace and job_name (NOT NULL).
	// Column is "job_name" not "type", and "namespace" has no default.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO jobs.queue (namespace, job_name, payload, status, tenant_id)
		VALUES ('default', 'test_job', '{"msg": "tenant1 job"}'::jsonb, 'pending', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO jobs.queue (namespace, job_name, payload, status, tenant_id)
		VALUES ('default', 'test_job', '{"msg": "tenant2 job"}'::jsonb, 'pending', $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's job
	tenant1Jobs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT payload->>'msg' as msg FROM jobs.queue WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	require.Len(t, tenant1Jobs, 1, "Tenant1 should only see their own jobs")
	require.Equal(t, "tenant1 job", tenant1Jobs[0]["msg"])

	// Verify as tenant2: should only see tenant2's job
	tenant2Jobs := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT payload->>'msg' as msg FROM jobs.queue WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	require.Len(t, tenant2Jobs, 1, "Tenant2 should only see their own jobs")
	require.Equal(t, "tenant2 job", tenant2Jobs[0]["msg"])
}

// ============================================================================
// RPC SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_RPCProcedures verifies that rpc.procedures are isolated by tenant_id.
func TestTenantIsolation_RPCProcedures(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert RPC procedures for two tenants.
	// rpc.procedures requires name and sql_query (NOT NULL).
	// Column is "sql_query" not "definition".
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO rpc.procedures (name, sql_query, tenant_id)
		VALUES ('tenant-test-proc1', 'SELECT 1', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO rpc.procedures (name, sql_query, tenant_id)
		VALUES ('tenant-test-proc2', 'SELECT 2', $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's procedure
	tenant1Procs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM rpc.procedures WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant1Procs, 1, "Tenant1 should only see their own procedures")
	require.Equal(t, "tenant-test-proc1", tenant1Procs[0]["name"])

	// Verify as tenant2: should only see tenant2's procedure
	tenant2Procs := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT name FROM rpc.procedures WHERE name LIKE 'tenant-test-%'`)

	require.Len(t, tenant2Procs, 1, "Tenant2 should only see their own procedures")
	require.Equal(t, "tenant-test-proc2", tenant2Procs[0]["name"])
}

// ============================================================================
// REALTIME SCHEMA TENANT ISOLATION
// ============================================================================

// TestTenantIsolation_RealtimeSchemaRegistry verifies that realtime.schema_registry is isolated by tenant_id.
func TestTenantIsolation_RealtimeSchemaRegistry(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert schema_registry entries for two tenants.
	// realtime.schema_registry requires schema_name and table_name (NOT NULL).
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO realtime.schema_registry (schema_name, table_name, tenant_id)
		VALUES ('public', 'tenant_test_table1', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO realtime.schema_registry (schema_name, table_name, tenant_id)
		VALUES ('public', 'tenant_test_table2', $1)
	`, tenantTestID2)

	// Verify as tenant1: should only see tenant1's entry
	tenant1Entries := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT table_name FROM realtime.schema_registry WHERE table_name LIKE 'tenant_test_%'`)

	require.Len(t, tenant1Entries, 1, "Tenant1 should only see their own schema_registry entries")
	require.Equal(t, "tenant_test_table1", tenant1Entries[0]["table_name"])

	// Verify as tenant2: should only see tenant2's entry
	tenant2Entries := tc.QuerySQLAsTenant(tenantTestID2,
		`SELECT table_name FROM realtime.schema_registry WHERE table_name LIKE 'tenant_test_%'`)

	require.Len(t, tenant2Entries, 1, "Tenant2 should only see their own schema_registry entries")
	require.Equal(t, "tenant_test_table2", tenant2Entries[0]["table_name"])
}

// ============================================================================
// CROSS-TENANT DATA LEAKAGE PREVENTION
// ============================================================================

// TestTenantIsolation_NoLeakageAcrossSchemas verifies that setting a tenant context
// in one schema does not leak data from other schemas.
func TestTenantIsolation_NoLeakageAcrossSchemas(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert data across multiple schemas for both tenants.
	// Each INSERT is a separate call because pgx doesn't support
	// multiple statements in a single parameterized Exec.
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant1-log', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant2-log', $1)
	`, tenantTestID2)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO branching.branches (name, slug, database_name, status, tenant_id)
		VALUES ('tenant1-branch', 'tenant1-branch', 'db1', 'ready', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO branching.branches (name, slug, database_name, status, tenant_id)
		VALUES ('tenant2-branch', 'tenant2-branch', 'db2', 'ready', $1)
	`, tenantTestID2)

	// Query as tenant1 across all schemas
	logs := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT message FROM logging.entries WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	branches := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT name FROM branching.branches WHERE tenant_id IN ($1::uuid, $2::uuid)`,
		tenantTestID1, tenantTestID2)

	// Verify no cross-tenant leakage
	require.Len(t, logs, 1, "Tenant1 should see exactly 1 log entry")
	require.Equal(t, "tenant1-log", logs[0]["message"])

	require.Len(t, branches, 1, "Tenant1 should see exactly 1 branch")
	require.Equal(t, "tenant1-branch", branches[0]["name"])
}

// TestTenantIsolation_NullTenantAccess verifies that records with NULL tenant_id
// are accessible when no tenant context is set (default/legacy behavior).
func TestTenantIsolation_NullTenantAccess(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert a record with NULL tenant_id (default tenant)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'null-tenant-log', NULL)
	`)

	// Query with no tenant context (empty string)
	// Using superuser to call has_tenant_access directly since rlsPool can't
	// easily query without setting a tenant context
	allowed := tc.QuerySQLAsSuperuser(`
		SELECT auth.has_tenant_access(NULL::uuid) as allowed
	`)
	require.Len(t, allowed, 1)
	require.Equal(t, true, allowed[0]["allowed"], "NULL tenant_id should be accessible with no context")

	// Verify the record with NULL tenant_id is NOT visible when tenant context IS set
	tenant1Entries := tc.QuerySQLAsTenant(tenantTestID1,
		`SELECT message FROM logging.entries WHERE message = 'null-tenant-log'`)
	require.Len(t, tenant1Entries, 0, "NULL tenant records should NOT be visible when tenant context is set")
}

// TestTenantIsolation_ServiceRoleBypassesTenantRLS verifies that service_role
// can see all tenant data (bypasses RLS).
func TestTenantIsolation_ServiceRoleBypassesTenantRLS(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert data for both tenants (separate calls - pgx doesn't support
	// multiple statements in a single parameterized Exec).
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant1-log', $1)
	`, tenantTestID1)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO logging.entries (category, level, message, tenant_id)
		VALUES ('system', 'info', 'tenant2-log', $1)
	`, tenantTestID2)

	// Superuser should see ALL entries across tenants
	allLogs := tc.QuerySQLAsSuperuser(`
		SELECT message FROM logging.entries
		WHERE tenant_id IN ($1::uuid, $2::uuid)
		ORDER BY message
	`, tenantTestID1, tenantTestID2)

	require.Len(t, allLogs, 2, "Service role should see ALL tenant data")
	require.Equal(t, "tenant1-log", allLogs[0]["message"])
	require.Equal(t, "tenant2-log", allLogs[1]["message"])
}

// TestTenantIsolation_AutoPopulateTenantID verifies that the set_tenant_id_from_context()
// trigger auto-populates tenant_id on INSERT when app.current_tenant_id is set.
// Uses fmt.Sprintf (no params) so pgx uses the simple query protocol which
// supports multiple statements — all three run on the same connection.
func TestTenantIsolation_AutoPopulateTenantID(t *testing.T) {
	tc := setupTenantIsolationTest(t)

	// Insert a log entry WITHOUT explicitly setting tenant_id, but with tenant context set.
	// This must be a single multi-statement call (no params) to ensure SET and INSERT
	// run on the same database connection.
	tc.ExecuteSQLAsSuperuser(fmt.Sprintf(`
		SET app.current_tenant_id = '%s';
		INSERT INTO logging.entries (category, level, message)
		VALUES ('system', 'info', 'auto-tenant-log');
		RESET app.current_tenant_id;
	`, tenantTestID1))

	// Verify the entry was auto-populated with the correct tenant_id
	entries := tc.QuerySQLAsSuperuser(`
		SELECT tenant_id, message FROM logging.entries WHERE message = 'auto-tenant-log'
	`)
	require.Len(t, entries, 1, "Auto-tenant insert should create exactly one entry")

	// The tenant_id should have been auto-populated
	tenantID := entries[0]["tenant_id"]
	require.NotNil(t, tenantID, "tenant_id should be auto-populated by trigger")

	// Convert to string for comparison (UUID comes as formatted string from convertPgTypeToGoType)
	require.Contains(t, fmt.Sprintf("%v", tenantID), tenantTestID1[:8],
		"Auto-populated tenant_id should match context tenant")
}
