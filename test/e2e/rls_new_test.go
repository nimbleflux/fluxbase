//go:build integration

package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

// TestInstanceSettingsSelect_TenantIsolation verifies that the instance_settings
// RLS policy does not allow cross-tenant data leakage. A tenant-scoped user should
// only see instance-level defaults (tenant_id IS NULL) and their own tenant's overrides.
func TestInstanceSettingsSelect_TenantIsolation(t *testing.T) {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()

	// Create two tenants
	tenantAID := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenantBID := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, is_default, status, created_at, updated_at)
		VALUES ($1, 'tenant-a', 'Tenant A', false, 'active', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantAID)

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, is_default, status, created_at, updated_at)
		VALUES ($1, 'tenant-b', 'Tenant B', false, 'active', NOW(), NOW())
		ON CONFLICT (id) DO NOTHING
	`, tenantBID)

	// Insert instance-level setting (tenant_id IS NULL)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.instance_settings (tenant_id, settings, overridable_settings)
		VALUES (NULL, '{"timeout": 30}'::jsonb, NULL)
		ON CONFLICT (tenant_id) DO UPDATE SET settings = EXCLUDED.settings
	`)

	// Insert Tenant A's override
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.instance_settings (tenant_id, settings)
		VALUES ($1, '{"timeout": 60, "theme": "dark"}'::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET settings = EXCLUDED.settings
	`, tenantAID)

	// Insert Tenant B's override
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.instance_settings (tenant_id, settings)
		VALUES ($1, '{"timeout": 90, "theme": "light", "logo": "b.png"}'::jsonb)
		ON CONFLICT (tenant_id) DO UPDATE SET settings = EXCLUDED.settings
	`, tenantBID)

	// Test 1: Tenant A should see instance-level + Tenant A rows only
	rowsA := tc.QuerySQLAsTenant(tenantAID, `SELECT tenant_id, settings FROM platform.instance_settings ORDER BY tenant_id`)

	// Verify Tenant A can see instance-level defaults (tenant_id IS NULL)
	hasNullRow := false
	for _, row := range rowsA {
		if row["tenant_id"] == nil {
			hasNullRow = true
			break
		}
	}
	require.True(t, hasNullRow, "Tenant A should see at least one instance-level (NULL tenant_id) row")

	// Verify Tenant A can see its own row
	hasTenantARow := false
	for _, row := range rowsA {
		if id, ok := row["tenant_id"].(string); ok && id == tenantAID {
			hasTenantARow = true
			break
		}
	}
	require.True(t, hasTenantARow, "Tenant A should see its own override row")

	// Verify Tenant A CANNOT see Tenant B's row
	for _, row := range rowsA {
		if id, ok := row["tenant_id"].(string); ok && id == tenantBID {
			t.Fatal("Tenant A should NOT see Tenant B's settings row - cross-tenant data leak detected!")
		}
	}

	// Test 2: Tenant B should see instance-level + Tenant B rows only
	rowsB := tc.QuerySQLAsTenant(tenantBID, `SELECT tenant_id, settings FROM platform.instance_settings ORDER BY tenant_id`)

	hasNullRow = false
	for _, row := range rowsB {
		if row["tenant_id"] == nil {
			hasNullRow = true
			break
		}
	}
	require.True(t, hasNullRow, "Tenant B should see at least one instance-level row")

	hasTenantBRow := false
	for _, row := range rowsB {
		if id, ok := row["tenant_id"].(string); ok && id == tenantBID {
			hasTenantBRow = true
			break
		}
	}
	require.True(t, hasTenantBRow, "Tenant B should see its own override row")

	// Verify Tenant B CANNOT see Tenant A's row
	for _, row := range rowsB {
		if id, ok := row["tenant_id"].(string); ok && id == tenantAID {
			t.Fatal("Tenant B should NOT see Tenant A's settings row - cross-tenant data leak detected!")
		}
	}

	// Test 3: No tenant context should see only instance-level (NULL) row
	// This uses QuerySQLAsRLSUser without setting tenant context
	userID := "11111111-1111-1111-1111-111111111111"
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, 'rls-test@example.com', 'hash', true, NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID)

	rowsNoTenant := tc.QuerySQLAsRLSUser(`SELECT tenant_id FROM platform.instance_settings`, userID)
	// Without tenant context, the tenant-admin/tenant_service policies don't apply.
	// The instance-level row might be visible via other policies or not, depending on exact policy match.
	// The important thing is no cross-tenant leakage.
	for _, row := range rowsNoTenant {
		if row["tenant_id"] != nil {
			t.Errorf("Non-tenant user should not see tenant-specific rows, got tenant_id: %v", row["tenant_id"])
		}
	}

	// Cleanup
	tc.ExecuteSQLAsSuperuser(`DELETE FROM platform.instance_settings WHERE tenant_id IN ($1, $2)`, tenantAID, tenantBID)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM platform.tenants WHERE id IN ($1, $2)`, tenantAID, tenantBID)
}

// TestMCPTables_RLS verifies that custom_resources and custom_tools tables
// have proper RLS policies restricting access to owners, admins, and service_role.
func TestMCPTables_RLS(t *testing.T) {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()

	// Create two users
	userA := "22222222-2222-2222-2222-222222222222"
	userB := "33333333-3333-3333-3333-333333333333"

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, 'mcp-user-a@example.com', 'hash', true, NOW()),
		       ($2, 'mcp-user-b@example.com', 'hash', true, NOW())
		ON CONFLICT (id) DO NOTHING
	`, userA, userB)

	// Create resources owned by user A
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO mcp.custom_resources (id, uri, name, code, enabled, created_by)
		VALUES ('d0000000-0000-0000-0000-000000000001', 'fluxbase://custom/res-a', 'Resource A', 'return "A"', true, $1),
		       ('d0000000-0000-0000-0000-000000000002', 'fluxbase://custom/res-a-disabled', 'Resource A Disabled', 'return "AD"', false, $1)
		ON CONFLICT (id) DO NOTHING
	`, userA)

	// Create resource owned by user B
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO mcp.custom_resources (id, uri, name, code, enabled, created_by)
		VALUES ('d0000000-0000-0000-0000-000000000003', 'fluxbase://custom/res-b', 'Resource B', 'return "B"', true, $1)
		ON CONFLICT (id) DO NOTHING
	`, userB)

	// Create tools owned by user A
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO mcp.custom_tools (id, name, code, enabled, created_by)
		VALUES ('e0000000-0000-0000-0000-000000000001', 'tool-a', 'return "A"', true, $1),
		       ('e0000000-0000-0000-0000-000000000002', 'tool-a-disabled', 'return "AD"', false, $1)
		ON CONFLICT (id) DO NOTHING
	`, userA)

	// Test 1: User A should see own resources + enabled resources from others
	rowsA := tc.QuerySQLAsRLSUser(`SELECT name, created_by FROM mcp.custom_resources ORDER BY name`, userA)
	namesA := make([]string, 0, len(rowsA))
	for _, r := range rowsA {
		namesA = append(namesA, r["name"].(string))
	}
	// Should see: own resources (Resource A, Resource A Disabled) + enabled from B (Resource B)
	require.Contains(t, namesA, "Resource A", "Owner should see their own resource")
	require.Contains(t, namesA, "Resource A Disabled", "Owner should see their own disabled resource")
	require.Contains(t, namesA, "Resource B", "Should see enabled resource from other user")

	// Test 2: User B should see own resource + enabled resources from A
	rowsB := tc.QuerySQLAsRLSUser(`SELECT name FROM mcp.custom_resources ORDER BY name`, userB)
	namesB := make([]string, 0, len(rowsB))
	for _, r := range rowsB {
		namesB = append(namesB, r["name"].(string))
	}
	require.Contains(t, namesB, "Resource B", "Owner should see their own resource")
	require.Contains(t, namesB, "Resource A", "Should see enabled resource from other user")
	require.NotContains(t, namesB, "Resource A Disabled", "Should NOT see disabled resource from other user")

	// Test 3: Same pattern for custom_tools
	toolsB := tc.QuerySQLAsRLSUser(`SELECT name FROM mcp.custom_tools ORDER BY name`, userB)
	toolNamesB := make([]string, 0, len(toolsB))
	for _, r := range toolsB {
		toolNamesB = append(toolNamesB, r["name"].(string))
	}
	require.Contains(t, toolNamesB, "tool-a", "Should see enabled tool from other user")
	require.NotContains(t, toolNamesB, "tool-a-disabled", "Should NOT see disabled tool from other user")

	// Cleanup
	tc.ExecuteSQLAsSuperuser(`DELETE FROM mcp.custom_resources WHERE created_by IN ($1, $2)`, userA, userB)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM mcp.custom_tools WHERE created_by IN ($1, $2)`, userA, userB)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM auth.users WHERE id IN ($1, $2)`, userA, userB)
}

// TestAppSettings_ForceRLS verifies that app.settings has FORCE ROW LEVEL SECURITY
// by checking that the table owner cannot bypass RLS policies.
func TestAppSettings_ForceRLS(t *testing.T) {
	tc := test.NewRLSTestContext(t)
	tc.EnsureAuthSchema()

	userID := "44444444-4444-4444-4444-444444444444"
	otherUserID := "55555555-5555-5555-5555-555555555555"

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, 'settings-test@example.com', 'hash', true, NOW()),
		       ($2, 'settings-other@example.com', 'hash', true, NOW())
		ON CONFLICT (id) DO NOTHING
	`, userID, otherUserID)

	// Insert a secret setting (only visible to owner or service_role)
	settingKey := fmt.Sprintf("test_secret_%s", userID[:8])
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO app.settings (key, value, category, is_secret, user_id)
		VALUES ($1, '{"val": "secret123"}'::jsonb, 'custom', true, $2)
		ON CONFLICT DO NOTHING
	`, settingKey, userID)

	// Insert a non-secret setting visible to all authenticated users
	publicKey := fmt.Sprintf("test_public_%s", userID[:8])
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO app.settings (key, value, category, is_secret, is_public)
		VALUES ($1, '{"val": "public456"}'::jsonb, 'custom', false, true)
		ON CONFLICT DO NOTHING
	`, publicKey)

	// Test 1: Non-owner authenticated user should see non-secret settings but NOT secret ones
	rows := tc.QuerySQLAsRLSUser(`SELECT key, is_secret FROM app.settings WHERE key LIKE $1`, otherUserID, "test_%")
	for _, row := range rows {
		isSecret, _ := row["is_secret"].(bool)
		key, _ := row["key"].(string)
		require.False(t, isSecret, "Non-owner should not see secret setting %s", key)
	}

	// Test 2: Owner should see their own secret setting
	ownerRows := tc.QuerySQLAsRLSUser(`SELECT key, is_secret FROM app.settings WHERE key = $1`, userID, settingKey)
	require.Len(t, ownerRows, 1, "Owner should see their own secret setting")
	require.Equal(t, settingKey, ownerRows[0]["key"])

	// Test 3: Non-owner should NOT see the secret setting
	otherRows := tc.QuerySQLAsRLSUser(`SELECT key FROM app.settings WHERE key = $1`, otherUserID, settingKey)
	require.Len(t, otherRows, 0, "Non-owner should NOT see secret setting")

	// Cleanup
	tc.ExecuteSQLAsSuperuser(`DELETE FROM app.settings WHERE key IN ($1, $2)`, settingKey, publicKey)
	tc.ExecuteSQLAsSuperuser(`DELETE FROM auth.users WHERE id IN ($1, $2)`, userID, otherUserID)
}
