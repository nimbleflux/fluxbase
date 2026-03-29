package tenantdb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoolConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, int32(100), cfg.Pool.MaxTotalConnections)
	assert.Equal(t, 30*time.Minute, cfg.Pool.EvictionAge)
}

func TestMigrationsConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 5*time.Minute, cfg.Migrations.CheckInterval)
	assert.True(t, cfg.Migrations.OnCreate) // Run system migrations after bootstrap on tenant creation
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
