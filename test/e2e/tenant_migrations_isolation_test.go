//go:build integration

package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/test"
)

// TestTenantMigrations_Isolation verifies that a tenant service key can only
// sync migrations within its own tenant scope, not globally or for other tenants.
func TestTenantMigrations_Isolation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	tc := test.NewTestContext(t)

	// Create two tenants
	tenant1ID := "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	tenant2ID := "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, status)
		VALUES ($1, 'mig-test-tenant-1', 'Migration Test Tenant 1', 'active')
		ON CONFLICT (id) DO NOTHING
	`, tenant1ID)
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, status)
		VALUES ($1, 'mig-test-tenant-2', 'Migration Test Tenant 2', 'active')
		ON CONFLICT (id) DO NOTHING
	`, tenant2ID)

	// Create a global service key
	globalKey := tc.CreateServiceKey("mig-test-global-key")

	// Create a tenant-scoped service key (CreateServiceKey creates global keys,
	// then we update it to be tenant-scoped)
	tenant1Key := tc.CreateServiceKey("mig-test-tenant1-key")
	tc.ExecuteSQLAsSuperuser(`
		UPDATE auth.service_keys
		SET tenant_id = $1, key_type = 'tenant_service'
		WHERE name = 'mig-test-tenant1-key'
	`, tenant1ID)

	syncBody := map[string]interface{}{
		"namespace": "isolation-test",
		"migrations": []map[string]interface{}{
			{
				"name":   "001_isolation_test",
				"up_sql": "CREATE TABLE IF NOT EXISTS isolation_test (id serial PRIMARY KEY)",
			},
		},
		"options": map[string]interface{}{
			"auto_apply": true,
		},
	}

	t.Run("global service key can sync without tenant header", func(t *testing.T) {
		resp := tc.NewRequest("POST", "/api/v1/admin/migrations/sync").
			WithServiceKey(globalKey).
			WithBody(syncBody).
			Send()

		assert.Equal(t, http.StatusOK, resp.Status())
	})

	t.Run("tenant service key can sync with own tenant header", func(t *testing.T) {
		resp := tc.NewRequest("POST", "/api/v1/admin/migrations/sync").
			WithServiceKey(tenant1Key).
			WithHeader("X-FB-Tenant", tenant1ID).
			WithBody(map[string]interface{}{
				"namespace": "tenant1-test",
				"migrations": []map[string]interface{}{
					{
						"name":   "001_tenant1_table",
						"up_sql": "CREATE TABLE IF NOT EXISTS tenant1_table (id serial PRIMARY KEY)",
					},
				},
				"options": map[string]interface{}{
					"auto_apply": true,
				},
			}).
			Send()

		assert.Equal(t, http.StatusOK, resp.Status())
	})

	t.Run("tenant service key without tenant header is rejected", func(t *testing.T) {
		resp := tc.NewRequest("POST", "/api/v1/admin/migrations/sync").
			WithServiceKey(tenant1Key).
			WithBody(syncBody).
			Send()

		// Tenant key not found in global DB (it's tenant-scoped), so auth fails
		assert.Equal(t, http.StatusUnauthorized, resp.Status())
	})
}

// TestTenantMigrations_RequireRole verifies that the RequireRole middleware
// correctly handles different key types for migration routes.
func TestTenantMigrations_RequireRole(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test")
	}

	tc := test.NewTestContext(t)

	// Create a tenant
	tenantID := "cccc3333-cccc-cccc-cccc-cccccccccccc"
	tc.ExecuteSQLAsSuperuser(`
		INSERT INTO platform.tenants (id, slug, name, status)
		VALUES ($1, 'role-test-tenant', 'Role Test Tenant', 'active')
		ON CONFLICT (id) DO NOTHING
	`, tenantID)

	// Create keys of different types
	globalKey := tc.CreateServiceKey("role-test-global")

	tenantKey := tc.CreateServiceKey("role-test-tenant-key")
	tc.ExecuteSQLAsSuperuser(`
		UPDATE auth.service_keys
		SET tenant_id = $1, key_type = 'tenant_service'
		WHERE name = 'role-test-tenant-key'
	`, tenantID)

	// Migrations routes require admin/instance_admin/tenant_admin roles
	t.Run("global service key can list migrations", func(t *testing.T) {
		resp := tc.NewRequest("GET", "/api/v1/admin/migrations").
			WithServiceKey(globalKey).
			Send()

		assert.Equal(t, http.StatusOK, resp.Status(), fmt.Sprintf("Body: %s", string(resp.Body())))
	})

	t.Run("tenant service key with tenant header can list migrations", func(t *testing.T) {
		resp := tc.NewRequest("GET", "/api/v1/admin/migrations").
			WithServiceKey(tenantKey).
			WithHeader("X-FB-Tenant", tenantID).
			Send()

		assert.Equal(t, http.StatusOK, resp.Status(), fmt.Sprintf("Body: %s", string(resp.Body())))
	})

	t.Run("tenant service key without tenant header rejected", func(t *testing.T) {
		resp := tc.NewRequest("GET", "/api/v1/admin/migrations").
			WithServiceKey(tenantKey).
			Send()

		assert.Equal(t, http.StatusUnauthorized, resp.Status())
	})
}
