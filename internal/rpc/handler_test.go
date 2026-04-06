package rpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewHandler Tests
// =============================================================================

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)

		require.NotNil(t, handler)
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.loader)
		assert.NotNil(t, handler.validator)
		assert.NotNil(t, handler.executor)
	})

	t.Run("creates handler with storage", func(t *testing.T) {
		storage := NewStorage(nil)
		handler := NewHandler(nil, storage, nil, nil, nil, nil, nil, nil)

		require.NotNil(t, handler)
		assert.Equal(t, storage, handler.storage)
	})
}

// =============================================================================
// Handler SetScheduler Tests
// =============================================================================

func TestHandler_SetScheduler(t *testing.T) {
	t.Run("sets scheduler correctly", func(t *testing.T) {
		handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)

		assert.Nil(t, handler.scheduler)

		scheduler := NewScheduler(nil, nil)
		handler.SetScheduler(scheduler)

		assert.Equal(t, scheduler, handler.scheduler)
	})
}

// =============================================================================
// Handler GetExecutor Tests
// =============================================================================

func TestHandler_GetExecutor(t *testing.T) {
	t.Run("returns executor", func(t *testing.T) {
		handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)

		executor := handler.GetExecutor()

		assert.NotNil(t, executor)
		assert.Equal(t, handler.executor, executor)
	})
}

// =============================================================================
// stringSlicesEqual Tests
// =============================================================================

func TestStringSlicesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        []string
		b        []string
		expected bool
	}{
		{
			name:     "both empty",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "nil vs empty",
			a:        nil,
			b:        []string{},
			expected: true,
		},
		{
			name:     "equal single element",
			a:        []string{"hello"},
			b:        []string{"hello"},
			expected: true,
		},
		{
			name:     "equal multiple elements",
			a:        []string{"a", "b", "c"},
			b:        []string{"a", "b", "c"},
			expected: true,
		},
		{
			name:     "different length",
			a:        []string{"a", "b"},
			b:        []string{"a"},
			expected: false,
		},
		{
			name:     "same length different values",
			a:        []string{"a", "b"},
			b:        []string{"a", "c"},
			expected: false,
		},
		{
			name:     "same values different order",
			a:        []string{"a", "b"},
			b:        []string{"b", "a"},
			expected: false,
		},
		{
			name:     "one empty one has values",
			a:        []string{},
			b:        []string{"a"},
			expected: false,
		},
		{
			name:     "case sensitive",
			a:        []string{"Hello"},
			b:        []string{"hello"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringSlicesEqual(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// needsUpdate Tests (using Handler instance)
// =============================================================================

func TestHandler_NeedsUpdate(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	now := time.Now()

	t.Run("identical procedures don't need update", func(t *testing.T) {
		schedule := "0 * * * *"
		existing := &Procedure{
			ID:                      "proc-1",
			Name:                    "test",
			SQLQuery:                "SELECT 1",
			OriginalCode:            "-- test\nSELECT 1",
			Description:             "Test procedure",
			MaxExecutionTimeSeconds: 30,
			IsPublic:                true,
			DisableExecutionLogs:    false,
			RequireRoles:            []string{"admin"},
			Schedule:                &schedule,
			AllowedTables:           []string{"users"},
			AllowedSchemas:          []string{"public"},
			CreatedAt:               now,
			UpdatedAt:               now,
		}

		new := &Procedure{
			ID:                      "proc-1",
			Name:                    "test",
			SQLQuery:                "SELECT 1",
			OriginalCode:            "-- test\nSELECT 1",
			Description:             "Test procedure",
			MaxExecutionTimeSeconds: 30,
			IsPublic:                true,
			DisableExecutionLogs:    false,
			RequireRoles:            []string{"admin"},
			Schedule:                &schedule,
			AllowedTables:           []string{"users"},
			AllowedSchemas:          []string{"public"},
			CreatedAt:               now,
			UpdatedAt:               now,
		}

		assert.False(t, handler.needsUpdate(existing, new))
	})

	t.Run("different SQLQuery needs update", func(t *testing.T) {
		existing := &Procedure{SQLQuery: "SELECT 1"}
		new := &Procedure{SQLQuery: "SELECT 2"}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different OriginalCode needs update", func(t *testing.T) {
		existing := &Procedure{OriginalCode: "-- v1\nSELECT 1"}
		new := &Procedure{OriginalCode: "-- v2\nSELECT 1"}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different Description needs update", func(t *testing.T) {
		existing := &Procedure{Description: "Old description"}
		new := &Procedure{Description: "New description"}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different MaxExecutionTimeSeconds needs update", func(t *testing.T) {
		existing := &Procedure{MaxExecutionTimeSeconds: 30}
		new := &Procedure{MaxExecutionTimeSeconds: 60}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different IsPublic needs update", func(t *testing.T) {
		existing := &Procedure{IsPublic: false}
		new := &Procedure{IsPublic: true}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different DisableExecutionLogs needs update", func(t *testing.T) {
		existing := &Procedure{DisableExecutionLogs: false}
		new := &Procedure{DisableExecutionLogs: true}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different RequireRoles length needs update", func(t *testing.T) {
		existing := &Procedure{RequireRoles: []string{"admin"}}
		new := &Procedure{RequireRoles: []string{"admin", "editor"}}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different RequireRoles values needs update", func(t *testing.T) {
		existing := &Procedure{RequireRoles: []string{"admin"}}
		new := &Procedure{RequireRoles: []string{"editor"}}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("schedule nil vs non-nil needs update", func(t *testing.T) {
		schedule := "0 * * * *"
		existing := &Procedure{Schedule: nil}
		new := &Procedure{Schedule: &schedule}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("schedule non-nil vs nil needs update", func(t *testing.T) {
		schedule := "0 * * * *"
		existing := &Procedure{Schedule: &schedule}
		new := &Procedure{Schedule: nil}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different schedule values needs update", func(t *testing.T) {
		schedule1 := "0 * * * *"
		schedule2 := "*/5 * * * *"
		existing := &Procedure{Schedule: &schedule1}
		new := &Procedure{Schedule: &schedule2}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different AllowedTables needs update", func(t *testing.T) {
		existing := &Procedure{AllowedTables: []string{"users"}}
		new := &Procedure{AllowedTables: []string{"users", "orders"}}

		assert.True(t, handler.needsUpdate(existing, new))
	})

	t.Run("different AllowedSchemas needs update", func(t *testing.T) {
		existing := &Procedure{AllowedSchemas: []string{"public"}}
		new := &Procedure{AllowedSchemas: []string{"public", "auth"}}

		assert.True(t, handler.needsUpdate(existing, new))
	})
}

// =============================================================================
// UpdateProcedureRequest Tests
// =============================================================================

func TestUpdateProcedureRequest_Struct(t *testing.T) {
	t.Run("all fields are optional", func(t *testing.T) {
		req := UpdateProcedureRequest{}

		assert.Nil(t, req.Description)
		assert.Nil(t, req.Enabled)
		assert.Nil(t, req.IsPublic)
		assert.Nil(t, req.RequireRoles)
		assert.Nil(t, req.MaxExecutionTimeSeconds)
		assert.Nil(t, req.AllowedTables)
		assert.Nil(t, req.AllowedSchemas)
		assert.Nil(t, req.Schedule)
	})

	t.Run("fields can be set", func(t *testing.T) {
		desc := "Updated description"
		enabled := true
		isPublic := false
		maxTime := 60
		schedule := "0 * * * *"

		req := UpdateProcedureRequest{
			Description:             &desc,
			Enabled:                 &enabled,
			IsPublic:                &isPublic,
			RequireRoles:            []string{"admin"},
			MaxExecutionTimeSeconds: &maxTime,
			AllowedTables:           []string{"users"},
			AllowedSchemas:          []string{"public"},
			Schedule:                &schedule,
		}

		assert.Equal(t, "Updated description", *req.Description)
		assert.True(t, *req.Enabled)
		assert.False(t, *req.IsPublic)
		assert.Equal(t, []string{"admin"}, req.RequireRoles)
		assert.Equal(t, 60, *req.MaxExecutionTimeSeconds)
		assert.Equal(t, "0 * * * *", *req.Schedule)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkStringSlicesEqual_Equal(b *testing.B) {
	a := []string{"admin", "editor", "viewer", "guest"}
	b2 := []string{"admin", "editor", "viewer", "guest"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stringSlicesEqual(a, b2)
	}
}

func BenchmarkStringSlicesEqual_NotEqual(b *testing.B) {
	a := []string{"admin", "editor", "viewer"}
	b2 := []string{"admin", "editor", "guest"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stringSlicesEqual(a, b2)
	}
}

func BenchmarkHandler_NeedsUpdate_NoUpdate(b *testing.B) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	schedule := "0 * * * *"

	existing := &Procedure{
		SQLQuery:                "SELECT 1",
		Description:             "Test",
		MaxExecutionTimeSeconds: 30,
		IsPublic:                true,
		RequireRoles:            []string{"admin"},
		Schedule:                &schedule,
		AllowedTables:           []string{"users"},
	}
	new := &Procedure{
		SQLQuery:                "SELECT 1",
		Description:             "Test",
		MaxExecutionTimeSeconds: 30,
		IsPublic:                true,
		RequireRoles:            []string{"admin"},
		Schedule:                &schedule,
		AllowedTables:           []string{"users"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.needsUpdate(existing, new)
	}
}

func BenchmarkHandler_NeedsUpdate_NeedsUpdate(b *testing.B) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)

	existing := &Procedure{SQLQuery: "SELECT 1"}
	new := &Procedure{SQLQuery: "SELECT 2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.needsUpdate(existing, new)
	}
}
