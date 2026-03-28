package tenantdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRouter_GetPool(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestRouter_RemovePool(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestRouter_CloseAllPools(t *testing.T) {
	t.Skip("Requires database setup - run with test database")
}

func TestPoolConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, int32(100), cfg.Pool.MaxTotalConnections)
	assert.Equal(t, 30*time.Minute, cfg.Pool.EvictionAge)
}

func TestMigrationsConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 5*time.Minute, cfg.Migrations.CheckInterval)
	assert.False(t, cfg.Migrations.OnCreate) // Disabled by default - use declarative schemas
	assert.False(t, cfg.Migrations.OnAccess)
	assert.False(t, cfg.Migrations.Background)
}

func TestConfig_ZeroMaxTenants(t *testing.T) {
	cfg := Config{
		Enabled:        true,
		DatabasePrefix: "tenant_",
		MaxTenants:     0,
	}

	assert.Equal(t, 0, cfg.MaxTenants)
}
