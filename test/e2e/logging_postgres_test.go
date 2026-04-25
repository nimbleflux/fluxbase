//go:build integration && !no_e2e
// +build integration,!no_e2e

package e2e

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/storage"
	test "github.com/nimbleflux/fluxbase/test"
)

// =============================================================================
// Integration Tests - PostgreSQL Logging Backend
// =============================================================================

// cleanupLoggingEntries truncates the logging.entries table before each test
// to ensure test isolation.
func cleanupLoggingEntries(t *testing.T) {
	tc := test.NewTestContext(t)
	defer tc.Close()

	_, err := tc.DB.Exec(context.Background(), "TRUNCATE TABLE logging.entries CASCADE")

	if err != nil && !strings.Contains(err.Error(), "schema") {
		require.NoError(t, err)
	}
}

func TestLogging_Postgres_Write_Single(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	entry := &storage.LogEntry{
		ID:        uuid.New(),
		Timestamp: time.Now().Truncate(time.Microsecond),
		Category:  storage.LogCategorySystem,
		Level:     storage.LogLevelInfo,
		Message:   "Test log message",
		Component: "test-component",
		Fields:    map[string]any{"key": "value"},
	}

	err := logStorage.Write(ctx, []*storage.LogEntry{entry})
	require.NoError(t, err)

	// Verify entry was written
	result, err := logStorage.Query(ctx, storage.LogQueryOptions{
		Category: storage.LogCategorySystem,
		Limit:    1,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), result.TotalCount)
	if len(result.Entries) > 0 {
		assert.Equal(t, "Test log message", result.Entries[0].Message)
	}
}

func TestLogging_Postgres_Write_Batch(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	entries := make([]*storage.LogEntry, 10)
	for i := 0; i < 10; i++ {
		entries[i] = &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelInfo,
			Message:   fmt.Sprintf("HTTP log %d", i),
			Fields:    map[string]any{"index": i},
		}
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	// Verify all entries were written
	result, err := logStorage.Query(ctx, storage.LogQueryOptions{
		Category: storage.LogCategoryHTTP,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(10), result.TotalCount)
}

func TestLogging_Postgres_AllCategories(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)
	testUserID := uuid.New()

	t.Run("system logs", func(t *testing.T) {
		entry := &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelError,
			Message:   "System error occurred",
			Component: "auth",
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategorySystem,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, storage.LogCategorySystem, result.Entries[0].Category)
		assert.Equal(t, storage.LogLevelError, result.Entries[0].Level)
		assert.Equal(t, "auth", result.Entries[0].Component)
	})

	t.Run("HTTP logs", func(t *testing.T) {
		entry := &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelInfo,
			Message:   "GET /api/users -> 200 (45ms)",
			Fields: map[string]any{
				"method":      "GET",
				"path":        "/api/users",
				"status_code": 200,
				"duration_ms": 45,
			},
			RequestID: "req-123",
			TraceID:   "trace-456",
			UserID:    testUserID.String(),
			IPAddress: "192.168.1.1",
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategoryHTTP,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "req-123", result.Entries[0].RequestID)
	})

	t.Run("security logs", func(t *testing.T) {
		entry := &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategorySecurity,
			Level:     storage.LogLevelWarn,
			Message:   "login_failed",
			Fields: map[string]any{
				"event_type": "login_failed",
				"success":    false,
				"email":      "test@example.com",
			},
			UserID:    testUserID.String(),
			IPAddress: "10.0.0.1",
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategorySecurity,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
	})

	t.Run("execution logs", func(t *testing.T) {
		execID := uuid.New()
		entry := &storage.LogEntry{
			ID:          uuid.New(),
			Timestamp:   time.Now().Truncate(time.Microsecond),
			Category:    storage.LogCategoryExecution,
			Level:       storage.LogLevelDebug,
			Message:     "Starting function execution",
			ExecutionID: execID.String(),
			LineNumber:  1,
			Fields: map[string]any{
				"execution_type": "function",
				"function_name":  "my-function",
				"namespace":      "default",
			},
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategoryExecution,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, 1, result.Entries[0].LineNumber)
	})

	t.Run("AI logs", func(t *testing.T) {
		entry := &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryAI,
			Level:     storage.LogLevelInfo,
			Message:   "AI query",
			Fields: map[string]any{
				"query":           "What is the weather?",
				"model":           "gpt-4",
				"response_tokens": 150,
				"prompt_tokens":   20,
			},
			UserID: testUserID.String(),
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategoryAI,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
	})

	t.Run("custom category logs", func(t *testing.T) {
		entry := &storage.LogEntry{
			ID:             uuid.New(),
			Timestamp:      time.Now().Truncate(time.Microsecond),
			Category:       storage.LogCategoryCustom,
			CustomCategory: "billing",
			Level:          storage.LogLevelInfo,
			Message:        "Payment processed",
			Fields:         map[string]any{"amount": 100},
		}

		err := logStorage.Write(ctx, []*storage.LogEntry{entry})
		require.NoError(t, err)

		// Query by category and custom category name
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category:       storage.LogCategoryCustom,
			CustomCategory: "billing",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "billing", result.Entries[0].CustomCategory)
	})
}

func TestLogging_Postgres_Filters(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	// Insert test data for filtering
	entries := []*storage.LogEntry{
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-2 * time.Hour).Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelInfo,
			Message:   "Info request",
			Component: "api",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-1 * time.Hour).Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelWarn,
			Message:   "Warning request",
			Component: "api",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelError,
			Message:   "Error request",
			Component: "api",
		},
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	t.Run("filters by multiple levels", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategoryHTTP,
			Levels:   []storage.LogLevel{storage.LogLevelWarn, storage.LogLevelError},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("filters by component", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Component: "api",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(3), result.TotalCount)
	})

	t.Run("filters by time range", func(t *testing.T) {
		startTime := time.Now().Add(-90 * time.Minute)
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category:  storage.LogCategoryHTTP,
			StartTime: startTime,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, result.TotalCount, int64(2))
	})

	t.Run("searches by text", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Category: storage.LogCategoryHTTP,
			Search:   "Warning",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Contains(t, result.Entries[0].Message, "Warning")
	})
}

func TestLogging_Postgres_Pagination(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	// Insert 25 entries
	entries := make([]*storage.LogEntry, 25)
	for i := 0; i < 25; i++ {
		entries[i] = &storage.LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(time.Duration(i) * time.Second).Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelInfo,
			Message:   fmt.Sprintf("Log entry %d", i),
		}
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	t.Run("respects limit parameter", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Limit: 10,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(25), result.TotalCount)
		assert.Len(t, result.Entries, 10)
		assert.True(t, result.HasMore)
	})

	t.Run("respects offset parameter", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Limit:  10,
			Offset: 10,
		})
		require.NoError(t, err)
		assert.Len(t, result.Entries, 10)
		assert.True(t, result.HasMore)
	})

	t.Run("returns no more at end", func(t *testing.T) {
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{
			Limit:  10,
			Offset: 20,
		})
		require.NoError(t, err)
		assert.Len(t, result.Entries, 5)
		assert.False(t, result.HasMore)
	})
}

func TestLogging_Postgres_ExecutionLogs(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)
	execID := uuid.New()

	// Insert execution logs with line numbers
	entries := make([]*storage.LogEntry, 5)
	for i := 1; i <= 5; i++ {
		entries[i-1] = &storage.LogEntry{
			ID:          uuid.New(),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Millisecond).Truncate(time.Microsecond),
			Category:    storage.LogCategoryExecution,
			Level:       storage.LogLevelDebug,
			Message:     fmt.Sprintf("Line %d", i),
			ExecutionID: execID.String(),
			LineNumber:  i,
			Fields:      map[string]any{"execution_type": "function"},
		}
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	t.Run("gets all execution logs", func(t *testing.T) {
		logs, err := logStorage.GetExecutionLogs(ctx, execID.String(), 0)
		require.NoError(t, err)
		assert.Len(t, logs, 5)
		assert.Equal(t, 1, logs[0].LineNumber)
		assert.Equal(t, 5, logs[4].LineNumber)
	})

	t.Run("gets logs after line number", func(t *testing.T) {
		logs, err := logStorage.GetExecutionLogs(ctx, execID.String(), 2)
		require.NoError(t, err)
		assert.Len(t, logs, 3)
		assert.Equal(t, 3, logs[0].LineNumber)
	})

	t.Run("returns empty for non-existent execution", func(t *testing.T) {
		logs, err := logStorage.GetExecutionLogs(ctx, uuid.New().String(), 0)
		require.NoError(t, err)
		assert.Len(t, logs, 0)
	})
}

func TestLogging_Postgres_Delete(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	// Insert test data
	entries := []*storage.LogEntry{
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-48 * time.Hour).Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelInfo,
			Message:   "Old log 1",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelInfo,
			Message:   "Recent log",
		},
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	t.Run("deletes logs older than cutoff", func(t *testing.T) {
		cutoff := time.Now().Add(-24 * time.Hour)
		deleted, err := logStorage.Delete(ctx, storage.LogQueryOptions{
			EndTime: cutoff,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), deleted)

		// Verify remaining entries
		result, err := logStorage.Query(ctx, storage.LogQueryOptions{})
		require.NoError(t, err)
		assert.Equal(t, int64(1), result.TotalCount)
		assert.Equal(t, "Recent log", result.Entries[0].Message)
	})

	t.Run("requires at least one filter", func(t *testing.T) {
		_, err := logStorage.Delete(ctx, storage.LogQueryOptions{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires at least one filter")
	})
}

func TestLogging_Postgres_Stats(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	// Insert test data with different categories and levels
	entries := []*storage.LogEntry{
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-2 * time.Hour).Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelInfo,
			Message:   "System info",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Add(-1 * time.Hour).Truncate(time.Microsecond),
			Category:  storage.LogCategorySystem,
			Level:     storage.LogLevelError,
			Message:   "System error",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelInfo,
			Message:   "HTTP request",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategoryHTTP,
			Level:     storage.LogLevelWarn,
			Message:   "HTTP warning",
		},
		{
			ID:        uuid.New(),
			Timestamp: time.Now().Truncate(time.Microsecond),
			Category:  storage.LogCategorySecurity,
			Level:     storage.LogLevelInfo,
			Message:   "Security event",
		},
	}

	err := logStorage.Write(ctx, entries)
	require.NoError(t, err)

	t.Run("returns overall statistics", func(t *testing.T) {
		stats, err := logStorage.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(5), stats.TotalEntries)
		assert.NotNil(t, stats.OldestEntry)
		assert.False(t, stats.OldestEntry.IsZero())
		assert.NotNil(t, stats.NewestEntry)
		assert.False(t, stats.NewestEntry.IsZero())
		assert.True(t, stats.NewestEntry.After(*stats.OldestEntry))
	})

	t.Run("counts entries by category", func(t *testing.T) {
		stats, err := logStorage.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(2), stats.EntriesByCategory[storage.LogCategorySystem])
		assert.Equal(t, int64(2), stats.EntriesByCategory[storage.LogCategoryHTTP])
		assert.Equal(t, int64(1), stats.EntriesByCategory[storage.LogCategorySecurity])
	})

	t.Run("counts entries by level", func(t *testing.T) {
		stats, err := logStorage.Stats(ctx)
		require.NoError(t, err)
		assert.Equal(t, int64(3), stats.EntriesByLevel[storage.LogLevelInfo])
		assert.Equal(t, int64(1), stats.EntriesByLevel[storage.LogLevelError])
		assert.Equal(t, int64(1), stats.EntriesByLevel[storage.LogLevelWarn])
	})
}

func TestLogging_Postgres_Health(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	err := logStorage.Health(ctx)
	assert.NoError(t, err)
}

func TestLogging_Postgres_ErrorHandling(t *testing.T) {
	cleanupLoggingEntries(t)

	tc := test.NewTestContext(t)
	defer tc.Close()
	ctx := context.Background()
	db := database.NewConnectionWithPool(tc.DB.Pool())
	logStorage := storage.NewPostgresLogStorage(db)

	t.Run("handles invalid execution ID", func(t *testing.T) {
		_, err := logStorage.GetExecutionLogs(ctx, "not-a-uuid", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid execution ID")
	})
}
