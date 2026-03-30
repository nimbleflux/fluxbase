package bootstrap

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

// TestNeedsBootstrap tests the NeedsBootstrap function
func TestNeedsBootstrap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test requires a database connection
	// In CI, it would be run against a test database
	t.Run("returns true for empty database", func(t *testing.T) {
		// This would need a test database pool
		t.Skip("Requires database connection")
	})
}

// TestBootstrapSQL tests that the embedded SQL is valid
func TestBootstrapSQL(t *testing.T) {
	assert.NotEmpty(t, bootstrapSQL, "bootstrap SQL should not be empty")
	assert.Contains(t, bootstrapSQL, "CREATE SCHEMA IF NOT EXISTS auth", "should create auth schema")
	assert.Contains(t, bootstrapSQL, "CREATE SCHEMA IF NOT EXISTS storage", "should create storage schema")
	assert.Contains(t, bootstrapSQL, "CREATE ROLE anon", "should create anon role")
	assert.Contains(t, bootstrapSQL, "CREATE ROLE authenticated", "should create authenticated role")
	assert.Contains(t, bootstrapSQL, "CREATE ROLE service_role", "should create service_role")
}

// TestNewService tests service creation
func TestNewService(t *testing.T) {
	pool := &pgxpool.Pool{}
	svc := NewService(pool)
	assert.NotNil(t, svc, "service should not be nil")
}

// TestState tests state struct
func TestState(t *testing.T) {
	state := &State{
		Bootstrapped:   true,
		Version:        "1.0.0",
		Checksum:       "abc123",
		BootstrappedAt: "2024-01-01T00:00:00Z",
	}

	assert.True(t, state.Bootstrapped)
	assert.Equal(t, "1.0.0", state.Version)
	assert.Equal(t, "abc123", state.Checksum)
}

// Integration tests would go in internal/database/bootstrap/integration/
// They require a running database and would test:
// - NeedsBootstrap returns true for empty database
// - NeedsBootstrap returns false after RunBootstrap
// - IsBootstrapped returns correct state
// - RunBootstrap is idempotent
