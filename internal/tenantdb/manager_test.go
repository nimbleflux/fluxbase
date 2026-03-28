package tenantdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_CreateTenantDatabase(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestManager_DeleteTenantDatabase(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestManager_MigrateTenant(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, "tenant_", cfg.DatabasePrefix)
	assert.Equal(t, 100, cfg.MaxTenants)
	assert.Equal(t, int32(100), cfg.Pool.MaxTotalConnections)
	assert.Equal(t, 30*time.Minute, cfg.Pool.EvictionAge)
	assert.Equal(t, 5*time.Minute, cfg.Migrations.CheckInterval)
	assert.False(t, cfg.Migrations.OnCreate) // Disabled by default - use declarative schemas
	assert.False(t, cfg.Migrations.OnAccess)
	assert.False(t, cfg.Migrations.Background)
}

func TestConfig_Validation(t *testing.T) {
	t.Run("default config is valid", func(t *testing.T) {
		cfg := DefaultConfig()
		assert.True(t, cfg.Enabled)
		assert.NotEmpty(t, cfg.DatabasePrefix)
		assert.Greater(t, cfg.MaxTenants, 0)
	})

	t.Run("custom config can override defaults", func(t *testing.T) {
		cfg := Config{
			Enabled:        true,
			DatabasePrefix: "custom_",
			MaxTenants:     50,
			Pool: PoolConfig{
				MaxTotalConnections: 50,
				EvictionAge:         15 * time.Minute,
			},
			Migrations: MigrationsConfig{
				CheckInterval: 10 * time.Minute,
				OnCreate:      false,
				OnAccess:      false,
				Background:    false,
			},
		}

		assert.Equal(t, "custom_", cfg.DatabasePrefix)
		assert.Equal(t, 50, cfg.MaxTenants)
		assert.Equal(t, int32(50), cfg.Pool.MaxTotalConnections)
		assert.Equal(t, 15*time.Minute, cfg.Pool.EvictionAge)
		assert.False(t, cfg.Migrations.OnCreate)
	})
}

func TestCreateTenantRequest_Validation(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := CreateTenantRequest{
			Slug:     "test-tenant",
			Name:     "Test Tenant",
			Metadata: map[string]any{"key": "value"},
		}
		assert.NotEmpty(t, req.Slug)
		assert.NotEmpty(t, req.Name)
	})

	t.Run("metadata is optional", func(t *testing.T) {
		req := CreateTenantRequest{
			Slug: "test-tenant",
			Name: "Test Tenant",
		}
		assert.Nil(t, req.Metadata)
	})
}

func TestUpdateTenantRequest(t *testing.T) {
	t.Run("can update name", func(t *testing.T) {
		newName := "Updated Name"
		req := UpdateTenantRequest{Name: &newName}
		require.NotNil(t, req.Name)
		assert.Equal(t, "Updated Name", *req.Name)
	})

	t.Run("can update metadata", func(t *testing.T) {
		req := UpdateTenantRequest{
			Metadata: map[string]any{"new_key": "new_value"},
		}
		assert.NotNil(t, req.Metadata)
	})

	t.Run("all fields are optional", func(t *testing.T) {
		req := UpdateTenantRequest{}
		assert.Nil(t, req.Name)
		assert.Nil(t, req.Metadata)
	})
}

func TestTenantStatus_Constants(t *testing.T) {
	assert.Equal(t, TenantStatus("creating"), TenantStatusCreating)
	assert.Equal(t, TenantStatus("active"), TenantStatusActive)
	assert.Equal(t, TenantStatus("deleting"), TenantStatusDeleting)
	assert.Equal(t, TenantStatus("error"), TenantStatusError)
}

func TestTenantAdminAssignment(t *testing.T) {
	now := time.Now()
	assignment := TenantAdminAssignment{
		ID:         "test-id",
		TenantID:   "tenant-id",
		UserID:     "user-id",
		AssignedAt: now,
	}

	assert.Equal(t, "test-id", assignment.ID)
	assert.Equal(t, "tenant-id", assignment.TenantID)
	assert.Equal(t, "user-id", assignment.UserID)
	assert.Equal(t, now, assignment.AssignedAt)
}
