package migrations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMigrationState tests the MigrationState struct
func TestMigrationState(t *testing.T) {
	state := &MigrationState{
		HasImperativeMigrations: true,
		HasDeclarativeState:     false,
		LastAppliedVersion:      112,
		SchemaFingerprint:       "abc123def456",
	}

	assert.True(t, state.HasImperativeMigrations)
	assert.False(t, state.HasDeclarativeState)
	assert.Equal(t, int64(112), state.LastAppliedVersion)
	assert.Equal(t, "abc123def456", state.SchemaFingerprint)
}

// TestDeclarativeState tests the DeclarativeState struct
func TestDeclarativeState(t *testing.T) {
	state := DeclarativeState{
		ID:                1,
		SchemaFingerprint: "fingerprint123",
		Source:            "transitioned",
	}

	assert.Equal(t, 1, state.ID)
	assert.Equal(t, "fingerprint123", state.SchemaFingerprint)
	assert.Equal(t, "transitioned", state.Source)
}

// TestPlan tests the Plan struct
func TestPlan(t *testing.T) {
	plan := &Plan{
		Changes: []Change{
			{Type: ChangeCreate, ObjectType: "TABLE", Schema: "auth", Name: "users", SQL: "CREATE TABLE auth.users (...)", Destructive: false},
			{Type: ChangeAlter, ObjectType: "INDEX", Schema: "auth", Name: "idx_users_email", SQL: "CREATE INDEX ...", Destructive: false},
			{Type: ChangeDrop, ObjectType: "TABLE", Schema: "test", Name: "old_table", SQL: "DROP TABLE test.old_table", Destructive: true},
		},
		DDL:         "BEGIN; ... COMMIT;",
		Transaction: true,
	}

	assert.Len(t, plan.Changes, 3)
	assert.Equal(t, 1, countByTypeLocal(plan.Changes, ChangeCreate))
	assert.Equal(t, 1, countByTypeLocal(plan.Changes, ChangeAlter))
	assert.Equal(t, 1, countByTypeLocal(plan.Changes, ChangeDrop))
	assert.Equal(t, 1, countDestructiveLocal(plan.Changes))
}

// TestChange tests the Change struct
func TestChange(t *testing.T) {
	change := Change{
		Type:        ChangeCreate,
		ObjectType:  "TABLE",
		Schema:      "auth",
		Name:        "users",
		SQL:         "CREATE TABLE auth.users (...)",
		Destructive: false,
		DependsOn:   []string{"auth.schema"},
	}

	assert.Equal(t, ChangeCreate, change.Type)
	assert.Equal(t, "TABLE", change.ObjectType)
	assert.Equal(t, "auth", change.Schema)
	assert.Equal(t, "users", change.Name)
	assert.False(t, change.Destructive)
	assert.Len(t, change.DependsOn, 1)
}

// TestValidationResult tests the ValidationResult struct
func TestValidationResult(t *testing.T) {
	result := &ValidationResult{
		Valid:  true,
		Drifts: []Drift{},
		Error:  nil,
	}

	assert.True(t, result.Valid)
	assert.Len(t, result.Drifts, 0)
	assert.Nil(t, result.Error)
}

// TestDrift tests the Drift struct
func TestDrift(t *testing.T) {
	drift := Drift{
		Type:        "CREATE",
		ObjectType:  "TABLE",
		Schema:      "auth",
		Name:        "new_table",
		SQL:         "CREATE TABLE auth.new_table (...)",
		Destructive: false,
	}

	assert.Equal(t, "CREATE", drift.Type)
	assert.Equal(t, "TABLE", drift.ObjectType)
	assert.Equal(t, "auth", drift.Schema)
	assert.Equal(t, "new_table", drift.Name)
	assert.False(t, drift.Destructive)
}

// TestApplyResult tests the ApplyResult struct
func TestApplyResult(t *testing.T) {
	result := &ApplyResult{
		Applied: []Change{
			{Type: ChangeCreate, ObjectType: "TABLE", Schema: "auth", Name: "users"},
		},
		Duration: 1000000000, // 1 second in nanoseconds
		Error:    nil,
	}

	assert.Len(t, result.Applied, 1)
	assert.Equal(t, int64(1000000000), int64(result.Duration))
	assert.Nil(t, result.Error)
}

// TestTransitionOptions tests the TransitionOptions struct
func TestTransitionOptions(t *testing.T) {
	opts := TransitionOptions{
		SchemaDir:         "internal/database/schema/schemas",
		KeepOldMigrations: true, // Keep old migrations for user transition
		MigrationsDir:     "",   // No longer used - internal migrations removed
	}

	assert.Equal(t, "internal/database/schema/schemas", opts.SchemaDir)
	assert.True(t, opts.KeepOldMigrations)
	assert.Equal(t, "", opts.MigrationsDir)
}

// TestTransitionResult tests the TransitionResult struct
func TestTransitionResult(t *testing.T) {
	result := &TransitionResult{
		Success:              true,
		SchemaFile:           "internal/database/schema/schemas",
		SchemaFingerprint:    "abc123def456",
		LastMigrationVersion: 112,
	}

	assert.True(t, result.Success)
	assert.Equal(t, "internal/database/schema/schemas", result.SchemaFile)
	assert.Equal(t, "abc123def456", result.SchemaFingerprint)
	assert.Equal(t, int64(112), result.LastMigrationVersion)
}

// TestTransitionStatus tests the TransitionStatus struct
func TestTransitionStatus(t *testing.T) {
	now := time.Now()
	status := &TransitionStatus{
		HasImperativeMigrations: true,
		HasDeclarativeState:     true,
		LastMigrationVersion:    112,
		TransitionedAt:          &now,
		Source:                  "transitioned",
	}

	assert.True(t, status.HasImperativeMigrations)
	assert.True(t, status.HasDeclarativeState)
	assert.Equal(t, int64(112), status.LastMigrationVersion)
	assert.NotNil(t, status.TransitionedAt)
	assert.Equal(t, "transitioned", status.Source)
}

// Helper functions for tests
func countByTypeLocal(changes []Change, changeType ChangeType) int {
	count := 0
	for _, c := range changes {
		if c.Type == changeType {
			count++
		}
	}
	return count
}

func countDestructiveLocal(changes []Change) int {
	count := 0
	for _, c := range changes {
		if c.Destructive {
			count++
		}
	}
	return count
}
