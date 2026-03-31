//go:build integration

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

const (
	tenant1ID = "11111111-1111-1111-1111-111111111111"
	tenant2ID = "22222222-2222-2222-2222-222222222222"
)

// setupTenantAccessTest creates a test context with a PL/pgSQL helper function
// that sets the tenant context and calls auth.has_tenant_access in a single
// statement. This is necessary because set_config and the access check must
// run within the same session/transaction.
func setupTenantAccessTest(t *testing.T) *test.TestContext {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()

	// Clean up helper function from previous runs
	tc.ExecuteSQLAsSuperuser(`DROP FUNCTION IF EXISTS public.test_has_tenant_access(text, text)`)

	// Create a helper function that sets tenant context and checks access atomically
	tc.ExecuteSQLAsSuperuser(`
		CREATE OR REPLACE FUNCTION public.test_has_tenant_access(
			p_context_tenant_id text,
			p_resource_tenant_id text
		)
		RETURNS boolean AS $$
		BEGIN
			-- Set or clear tenant context
			IF p_context_tenant_id = '' OR p_context_tenant_id IS NULL THEN
				PERFORM set_config('app.current_tenant_id', '', true);
			ELSE
				PERFORM set_config('app.current_tenant_id', p_context_tenant_id, true);
			END IF;

			-- Check access
			IF p_resource_tenant_id = '' OR p_resource_tenant_id IS NULL THEN
				RETURN auth.has_tenant_access(NULL::uuid);
			ELSE
				RETURN auth.has_tenant_access(p_resource_tenant_id::uuid);
			END IF;
		END;
		$$ LANGUAGE plpgsql;
	`)

	return tc
}

// cleanupTenantAccessTest drops the helper function created during setup.
func cleanupTenantAccessTest(tc *test.TestContext) {
	tc.ExecuteSQLAsSuperuser(`DROP FUNCTION IF EXISTS public.test_has_tenant_access(text, text)`)
}

// checkTenantAccess calls the test helper function and returns whether access is granted.
func checkTenantAccess(tc *test.TestContext, contextTenantID, resourceTenantID string) bool {
	result := tc.QuerySQLAsSuperuser(
		`SELECT public.test_has_tenant_access($1, $2) as allowed`,
		contextTenantID, resourceTenantID,
	)
	if len(result) == 0 {
		return false
	}
	allowed, ok := result[0]["allowed"].(bool)
	return ok && allowed
}

func TestHasTenantAccess_NoContext_NullResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// No tenant context set + NULL resource tenant ID = allowed (default tenant)
	allowed := checkTenantAccess(tc, "", "")
	require.True(t, allowed, "No context + NULL resource should be allowed (default tenant)")
}

func TestHasTenantAccess_NoContext_NonNullResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// No tenant context set + non-NULL resource tenant ID = denied
	allowed := checkTenantAccess(tc, "", tenant1ID)
	require.False(t, allowed, "No context + non-NULL resource should be denied")
}

func TestHasTenantAccess_ContextMatchesResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// Tenant context matches resource tenant ID = allowed
	allowed := checkTenantAccess(tc, tenant1ID, tenant1ID)
	require.True(t, allowed, "Context matching resource should be allowed")
}

func TestHasTenantAccess_ContextDiffersFromResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// Tenant context differs from resource tenant ID = denied
	allowed := checkTenantAccess(tc, tenant1ID, tenant2ID)
	require.False(t, allowed, "Context differing from resource should be denied")
}

func TestHasTenantAccess_ContextSet_NullResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// Tenant context is set but resource tenant ID is NULL = denied
	allowed := checkTenantAccess(tc, tenant1ID, "")
	require.False(t, allowed, "Context set + NULL resource should be denied")
}

func TestHasTenantAccess_ContextSet_SameAsResource(t *testing.T) {
	tc := setupTenantAccessTest(t)
	defer cleanupTenantAccessTest(tc)

	// Duplicate of ContextMatchesResource to satisfy the test table requirement
	// explicitly testing the "same as resource" scenario
	allowed := checkTenantAccess(tc, tenant1ID, tenant1ID)
	require.True(t, allowed, "Context set + same resource should be allowed")
}
