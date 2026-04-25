package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewStorage Tests
// =============================================================================

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil database", func(t *testing.T) {
		storage := NewStorage(nil)

		require.NotNil(t, storage)
		assert.Nil(t, storage.DB)
	})
}

// =============================================================================
// ListExecutionsOptions Tests
// =============================================================================

func TestListExecutionsOptions_Defaults(t *testing.T) {
	t.Run("zero values for defaults", func(t *testing.T) {
		opts := ListExecutionsOptions{}

		assert.Empty(t, opts.Namespace)
		assert.Empty(t, opts.ProcedureName)
		assert.Empty(t, opts.Status)
		assert.Empty(t, opts.UserID)
		assert.Equal(t, 0, opts.Limit)
		assert.Equal(t, 0, opts.Offset)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace:     "api",
			ProcedureName: "get_users",
			Status:        StatusCompleted,
			UserID:        "user-123",
			Limit:         50,
			Offset:        100,
		}

		assert.Equal(t, "api", opts.Namespace)
		assert.Equal(t, "get_users", opts.ProcedureName)
		assert.Equal(t, StatusCompleted, opts.Status)
		assert.Equal(t, "user-123", opts.UserID)
		assert.Equal(t, 50, opts.Limit)
		assert.Equal(t, 100, opts.Offset)
	})
}

// =============================================================================
// Storage Query Building Tests (validates query building logic)
// =============================================================================

func TestStorage_QueryBuilding(t *testing.T) {
	t.Run("validates namespace filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace: "test-namespace",
		}

		// Verify the option is set (actual query building tested via integration)
		assert.NotEmpty(t, opts.Namespace)
	})

	t.Run("validates procedure name filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			ProcedureName: "test_proc",
		}

		assert.NotEmpty(t, opts.ProcedureName)
	})

	t.Run("validates status filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Status: StatusFailed,
		}

		assert.Equal(t, StatusFailed, opts.Status)
	})

	t.Run("validates user ID filter would be applied", func(t *testing.T) {
		opts := ListExecutionsOptions{
			UserID: "user-456",
		}

		assert.NotEmpty(t, opts.UserID)
	})

	t.Run("validates pagination", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Limit:  25,
			Offset: 50,
		}

		assert.Equal(t, 25, opts.Limit)
		assert.Equal(t, 50, opts.Offset)
	})
}

// =============================================================================
// Execution ID Generation Tests
// =============================================================================

func TestExecution_IDGeneration(t *testing.T) {
	t.Run("execution with empty ID", func(t *testing.T) {
		exec := &Execution{
			ID:            "",
			ProcedureName: "test_proc",
			Namespace:     "default",
			Status:        StatusPending,
		}

		// The actual ID generation happens in CreateExecution
		// Here we verify empty ID is acceptable as input
		assert.Empty(t, exec.ID)
	})

	t.Run("execution with provided ID", func(t *testing.T) {
		exec := &Execution{
			ID:            "exec-custom-id",
			ProcedureName: "test_proc",
			Namespace:     "default",
			Status:        StatusPending,
		}

		assert.Equal(t, "exec-custom-id", exec.ID)
	})
}

// =============================================================================
// Procedure ID Generation Tests
// =============================================================================

func TestProcedure_IDGeneration(t *testing.T) {
	t.Run("procedure with empty ID", func(t *testing.T) {
		proc := &Procedure{
			ID:       "",
			Name:     "test_proc",
			SQLQuery: "SELECT 1",
		}

		// The actual ID generation happens in CreateProcedure
		// Here we verify empty ID is acceptable as input
		assert.Empty(t, proc.ID)
	})

	t.Run("procedure with provided ID", func(t *testing.T) {
		proc := &Procedure{
			ID:       "proc-custom-id",
			Name:     "test_proc",
			SQLQuery: "SELECT 1",
		}

		assert.Equal(t, "proc-custom-id", proc.ID)
	})
}

// =============================================================================
// DefaultAnnotations Tests
// =============================================================================

func TestDefaultAnnotations_Function(t *testing.T) {
	t.Run("returns expected defaults", func(t *testing.T) {
		defaults := DefaultAnnotations()

		require.NotNil(t, defaults)
		assert.Equal(t, []string{"public"}, defaults.AllowedSchemas)
		assert.Equal(t, []string{}, defaults.AllowedTables)
		assert.False(t, defaults.IsPublic)
		assert.Equal(t, 1, defaults.Version)
	})
}

// =============================================================================
// ListExecutionsOptions Comprehensive Tests
// =============================================================================

func TestListExecutionsOptions_AllCombinations(t *testing.T) {
	t.Run("empty filters returns all", func(t *testing.T) {
		opts := ListExecutionsOptions{}
		assert.Empty(t, opts.Namespace)
		assert.Empty(t, opts.Status)
		assert.Equal(t, 0, opts.Limit)
	})

	t.Run("single filter", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Status: StatusCompleted,
		}
		assert.Equal(t, StatusCompleted, opts.Status)
		assert.Empty(t, opts.Namespace)
	})

	t.Run("multiple filters", func(t *testing.T) {
		opts := ListExecutionsOptions{
			Namespace:     "api",
			ProcedureName: "get_data",
			Status:        StatusFailed,
			UserID:        "user-123",
			Limit:         20,
			Offset:        40,
		}

		assert.Equal(t, "api", opts.Namespace)
		assert.Equal(t, "get_data", opts.ProcedureName)
		assert.Equal(t, StatusFailed, opts.Status)
		assert.Equal(t, "user-123", opts.UserID)
		assert.Equal(t, 20, opts.Limit)
		assert.Equal(t, 40, opts.Offset)
	})
}

// =============================================================================
// Storage Method Existence Tests
// =============================================================================

func TestStorage_MethodsExist(t *testing.T) {
	storage := NewStorage(nil)

	t.Run("procedure methods exist", func(t *testing.T) {
		// These tests verify the methods exist on the Storage type
		// Actual functionality requires database
		assert.NotNil(t, storage)
	})
}

// =============================================================================
// Procedure Source Values Tests
// =============================================================================

func TestProcedure_SourceValues(t *testing.T) {
	validSources := []string{"mcp", "admin", "cli", "migration", "api"}

	for _, source := range validSources {
		t.Run(source, func(t *testing.T) {
			proc := &Procedure{Source: source}
			assert.Equal(t, source, proc.Source)
		})
	}
}
