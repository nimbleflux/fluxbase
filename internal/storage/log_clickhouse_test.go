package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ClickHouseLogStorage Name Tests
// =============================================================================

func TestClickHouseLogStorage_Name(t *testing.T) {
	t.Run("returns clickhouse", func(t *testing.T) {
		storage := &ClickHouseLogStorage{}

		assert.Equal(t, "clickhouse", storage.Name())
	})
}

// =============================================================================
// ClickHouseLogStorage buildQuery Tests
// =============================================================================

func TestClickHouseLogStorage_buildQuery(t *testing.T) {
	storage := &ClickHouseLogStorage{}

	t.Run("returns empty for no filters", func(t *testing.T) {
		opts := LogQueryOptions{}

		query := storage.buildQuery(opts)

		assert.Empty(t, query.where)
		assert.Empty(t, query.args)
	})

	t.Run("filters by category", func(t *testing.T) {
		opts := LogQueryOptions{Category: LogCategoryHTTP}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "category = ?")
		assert.Len(t, query.args, 1)
		assert.Equal(t, "http", query.args[0])
	})

	t.Run("filters by custom category", func(t *testing.T) {
		opts := LogQueryOptions{CustomCategory: "my_category"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "custom_category = ?")
		assert.Len(t, query.args, 1)
		assert.Equal(t, "my_category", query.args[0])
	})

	t.Run("filters by multiple levels", func(t *testing.T) {
		opts := LogQueryOptions{Levels: []LogLevel{LogLevelError, LogLevelWarn}}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "level IN (")
		assert.Len(t, query.args, 2)
		assert.Equal(t, "error", query.args[0])
		assert.Equal(t, "warn", query.args[1])
	})

	t.Run("filters by component", func(t *testing.T) {
		opts := LogQueryOptions{Component: "auth"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "component = ?")
		assert.Equal(t, "auth", query.args[0])
	})

	t.Run("filters by request_id", func(t *testing.T) {
		opts := LogQueryOptions{RequestID: "req-123"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "request_id = ?")
		assert.Equal(t, "req-123", query.args[0])
	})

	t.Run("filters by trace_id", func(t *testing.T) {
		opts := LogQueryOptions{TraceID: "trace-456"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "trace_id = ?")
		assert.Equal(t, "trace-456", query.args[0])
	})

	t.Run("filters by user_id with valid UUID", func(t *testing.T) {
		userID := uuid.New().String()
		opts := LogQueryOptions{UserID: userID}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "user_id = ?")
		assert.Len(t, query.args, 1)
	})

	t.Run("ignores invalid user_id UUID", func(t *testing.T) {
		opts := LogQueryOptions{UserID: "not-a-uuid"}

		query := storage.buildQuery(opts)

		assert.NotContains(t, query.where, "user_id")
		assert.Empty(t, query.args)
	})

	t.Run("filters by execution_id with valid UUID", func(t *testing.T) {
		execID := uuid.New().String()
		opts := LogQueryOptions{ExecutionID: execID}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "execution_id = ?")
		assert.Len(t, query.args, 1)
	})

	t.Run("ignores invalid execution_id UUID", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionID: "not-a-uuid"}

		query := storage.buildQuery(opts)

		assert.NotContains(t, query.where, "execution_id")
		assert.Empty(t, query.args)
	})

	t.Run("filters by execution_type in execution_data map", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionType: "function"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "execution_data['execution_type'] = ?")
		assert.Equal(t, "function", query.args[0])
	})

	t.Run("filters by start_time", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		opts := LogQueryOptions{StartTime: startTime}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "timestamp >= ?")
		assert.Len(t, query.args, 1)
	})

	t.Run("filters by end_time", func(t *testing.T) {
		endTime := time.Now()
		opts := LogQueryOptions{EndTime: endTime}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "timestamp <= ?")
		assert.Len(t, query.args, 1)
	})

	t.Run("filters by search text using positionUTF8", func(t *testing.T) {
		opts := LogQueryOptions{Search: "error message"}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "positionUTF8CaseInsensitive(message, ?) > 0")
		assert.Equal(t, "error message", query.args[0])
	})

	t.Run("filters by after_line", func(t *testing.T) {
		opts := LogQueryOptions{AfterLine: 10}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, "line_number > ?")
		assert.Equal(t, 10, query.args[0])
	})

	t.Run("filters static assets when HideStaticAssets is enabled", func(t *testing.T) {
		opts := LogQueryOptions{
			Category:         LogCategoryHTTP,
			HideStaticAssets: true,
		}

		query := storage.buildQuery(opts)

		// Should have exclusion patterns for static assets
		assert.Contains(t, query.where, "http_data['path'] NOT LIKE ?")
		// Args should contain static extensions
		assert.True(t, len(query.args) > 0)
	})

	t.Run("combines multiple filters with AND", func(t *testing.T) {
		opts := LogQueryOptions{
			Category:  LogCategorySecurity,
			Component: "middleware",
			Levels:    []LogLevel{LogLevelInfo},
		}

		query := storage.buildQuery(opts)

		assert.Contains(t, query.where, " AND ")
		assert.Len(t, query.args, 3)
	})
}

// =============================================================================
// ClickHouseLogStorage toRow Tests
// =============================================================================

func TestClickHouseLogStorage_toRow(t *testing.T) {
	storage := &ClickHouseLogStorage{}

	t.Run("converts HTTP category entry", func(t *testing.T) {
		entryID := uuid.New()
		entry := &LogEntry{
			ID:        entryID,
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "GET /api/users",
			Component: "api",
			UserID:    uuid.New().String(),
			RequestID: "req-123",
			TraceID:   "trace-456",
			Fields: map[string]any{
				"method":     "GET",
				"path":       "/api/users",
				"status":     "200",
				"duration":   "45ms",
				"user_agent": "test-agent",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, entryID, row.id)
		assert.Equal(t, "http", row.category)
		assert.Equal(t, "info", row.level)
		assert.Equal(t, "GET /api/users", row.message)
		assert.NotNil(t, row.component)
		assert.Equal(t, "api", *row.component)
		assert.NotNil(t, row.userID)
		assert.NotNil(t, row.requestID)
		assert.Equal(t, "req-123", *row.requestID)
		assert.NotNil(t, row.traceID)
		assert.Equal(t, "trace-456", *row.traceID)
		assert.Len(t, row.httpData, 5)
		assert.Equal(t, "GET", row.httpData["method"])
		assert.Equal(t, "/api/users", row.httpData["path"])
	})

	t.Run("converts Security category entry", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategorySecurity,
			Level:    LogLevelWarn,
			Message:  "Failed login attempt",
			Fields: map[string]any{
				"event":  "login_failed",
				"ip":     "192.168.1.1",
				"user":   "attacker",
				"reason": "invalid_password",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "security", row.category)
		assert.Equal(t, "warn", row.level)
		assert.Len(t, row.securityData, 4)
		assert.Equal(t, "login_failed", row.securityData["event"])
		assert.Equal(t, "192.168.1.1", row.securityData["ip"])
	})

	t.Run("converts Execution category entry", func(t *testing.T) {
		execID := uuid.New()
		entry := &LogEntry{
			Category:    LogCategoryExecution,
			Level:       LogLevelInfo,
			Message:     "Function execution started",
			ExecutionID: execID.String(),
			LineNumber:  1,
			Fields: map[string]any{
				"execution_type": "function",
				"function_name":  "processData",
				"status":         "running",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "execution", row.category)
		assert.NotNil(t, row.executionID)
		assert.Equal(t, execID, *row.executionID)
		assert.NotNil(t, row.lineNumber)
		assert.Equal(t, uint32(1), *row.lineNumber)
		assert.Len(t, row.executionData, 3)
		assert.Equal(t, "function", row.executionData["execution_type"])
	})

	t.Run("converts AI category entry", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryAI,
			Level:    LogLevelInfo,
			Message:  "Vector search completed",
			Fields: map[string]any{
				"vector_dimension": 1536,
				"result_count":     "10",
				"search_time":      "23ms",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "ai", row.category)
		assert.Len(t, row.aiData, 3)
		assert.Equal(t, "1536", row.aiData["vector_dimension"])
		assert.Equal(t, "10", row.aiData["result_count"])
	})

	t.Run("converts Custom category entry", func(t *testing.T) {
		entry := &LogEntry{
			Category:       LogCategoryCustom,
			CustomCategory: "my_custom_category",
			Level:          LogLevelInfo,
			Message:        "Custom log entry",
			Fields: map[string]any{
				"custom_field1": "value1",
				"custom_field2": "value2",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "custom", row.category)
		assert.Equal(t, "my_custom_category", row.customCategory)
		assert.Len(t, row.customData, 2)
		assert.Equal(t, "value1", row.customData["custom_field1"])
	})

	t.Run("handles unknown category with custom_data", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategorySystem,
			Level:    LogLevelInfo,
			Message:  "System log",
			Fields: map[string]any{
				"system_field": "value",
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "system", row.category)
		assert.Len(t, row.customData, 1)
		assert.Equal(t, "value", row.customData["system_field"])
	})

	t.Run("handles nullable fields correctly", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "Minimal entry",
		}

		row := storage.toRow(entry)

		assert.Nil(t, row.component)
		assert.Nil(t, row.userID)
		assert.Nil(t, row.requestID)
		assert.Nil(t, row.traceID)
		assert.Nil(t, row.executionID)
		assert.Nil(t, row.lineNumber)
	})

	t.Run("handles invalid UUIDs gracefully", func(t *testing.T) {
		entry := &LogEntry{
			Category:    LogCategoryHTTP,
			Level:       LogLevelInfo,
			Message:     "Invalid UUID test",
			UserID:      "not-a-uuid",
			ExecutionID: "also-not-a-uuid",
		}

		row := storage.toRow(entry)

		assert.Nil(t, row.userID)
		assert.Nil(t, row.executionID)
	})
}

// =============================================================================
// Write Empty Batch Tests
// =============================================================================

func TestClickHouseLogStorage_Write_EmptyBatch(t *testing.T) {
	t.Run("returns nil for empty entries", func(t *testing.T) {
		storage := &ClickHouseLogStorage{}
		ctx := context.Background()
		entries := []*LogEntry{}

		err := storage.Write(ctx, entries)
		assert.NoError(t, err)
	})
}

// =============================================================================
// Entry ID and Timestamp Generation Tests
// =============================================================================

func TestClickHouseLogStorage_IDAndTimestampGeneration(t *testing.T) {
	t.Run("entry with nil ID should get UUID assigned", func(t *testing.T) {
		entry := &LogEntry{
			ID:       uuid.Nil,
			Category: LogCategoryHTTP,
			Level:    LogLevelInfo,
			Message:  "test",
		}

		// Verify the entry starts with nil UUID
		assert.Equal(t, uuid.Nil, entry.ID)
	})

	t.Run("entry with zero timestamp should get time assigned", func(t *testing.T) {
		entry := &LogEntry{
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "test",
			Timestamp: time.Time{},
		}

		// Verify the timestamp is zero
		assert.True(t, entry.Timestamp.IsZero())
	})
}

// =============================================================================
// Query Options Defaults Tests
// =============================================================================

func TestClickHouseLogStorage_QueryOptions(t *testing.T) {
	t.Run("default limit behavior", func(t *testing.T) {
		opts := LogQueryOptions{Limit: 0}
		assert.Equal(t, 0, opts.Limit)

		opts = LogQueryOptions{Limit: -5}
		assert.Equal(t, -5, opts.Limit)
	})

	t.Run("default offset behavior", func(t *testing.T) {
		opts := LogQueryOptions{Offset: -10}
		assert.Equal(t, -10, opts.Offset)
	})
}

// =============================================================================
// Delete Validation Tests
// =============================================================================

func TestClickHouseLogStorage_Delete_RequiresFilter(t *testing.T) {
	t.Run("returns error when no filters provided", func(t *testing.T) {
		storage := &ClickHouseLogStorage{}
		ctx := context.Background()

		count, err := storage.Delete(ctx, LogQueryOptions{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete requires at least one filter")
		assert.Equal(t, int64(0), count)
	})
}

// =============================================================================
// Join Conditions Tests
// =============================================================================

func TestJoinConditions(t *testing.T) {
	t.Run("joins single condition", func(t *testing.T) {
		conditions := []string{"category = ?"}

		result := joinConditions(conditions)

		assert.Equal(t, "category = ?", result)
	})

	t.Run("joins multiple conditions with AND", func(t *testing.T) {
		conditions := []string{"category = ?", "level = ?", "timestamp >= ?"}

		result := joinConditions(conditions)

		assert.Equal(t, "category = ? AND level = ? AND timestamp >= ?", result)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		conditions := []string{}

		result := joinConditions(conditions)

		assert.Empty(t, result)
	})
}

// =============================================================================
// Join Placeholders Tests
// =============================================================================

func TestJoinPlaceholders(t *testing.T) {
	t.Run("joins single placeholder", func(t *testing.T) {
		placeholders := []string{"?"}

		result := joinPlaceholders(placeholders)

		assert.Equal(t, "?", result)
	})

	t.Run("joins multiple placeholders with comma", func(t *testing.T) {
		placeholders := []string{"?", "?", "?"}

		result := joinPlaceholders(placeholders)

		assert.Equal(t, "?, ?, ?", result)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		placeholders := []string{}

		result := joinPlaceholders(placeholders)

		assert.Empty(t, result)
	})
}

// =============================================================================
// All Log Categories Tests
// =============================================================================

func TestClickHouseLogStorage_AllLogCategories(t *testing.T) {
	storage := &ClickHouseLogStorage{}

	t.Run("supports HTTP category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryHTTP,
			Message:  "HTTP request",
			Fields:   map[string]any{"method": "GET"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "http", row.category)
		assert.NotNil(t, row.httpData)
	})

	t.Run("supports Security category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategorySecurity,
			Message:  "Security event",
			Fields:   map[string]any{"event": "login"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "security", row.category)
		assert.NotNil(t, row.securityData)
	})

	t.Run("supports Execution category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryExecution,
			Message:  "Execution log",
			Fields:   map[string]any{"status": "running"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "execution", row.category)
		assert.NotNil(t, row.executionData)
	})

	t.Run("supports AI category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryAI,
			Message:  "AI operation",
			Fields:   map[string]any{"operation": "search"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "ai", row.category)
		assert.NotNil(t, row.aiData)
	})

	t.Run("supports Custom category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryCustom,
			Message:  "Custom log",
			Fields:   map[string]any{"custom": "value"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "custom", row.category)
		assert.NotNil(t, row.customData)
	})

	t.Run("supports System category", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategorySystem,
			Message:  "System log",
			Fields:   map[string]any{"system": "info"},
		}

		row := storage.toRow(entry)
		assert.Equal(t, "system", row.category)
		// System goes to custom_data
		assert.NotNil(t, row.customData)
	})
}

// =============================================================================
// Map Conversion Tests
// =============================================================================

func TestClickHouseLogStorage_MapConversion(t *testing.T) {
	storage := &ClickHouseLogStorage{}

	t.Run("converts various field types to strings", func(t *testing.T) {
		entry := &LogEntry{
			Category: LogCategoryHTTP,
			Fields: map[string]any{
				"string_val": "hello",
				"int_val":    42,
				"float_val":  3.14,
				"bool_val":   true,
				"null_val":   nil,
			},
		}

		row := storage.toRow(entry)

		assert.Equal(t, "hello", row.httpData["string_val"])
		assert.Equal(t, "42", row.httpData["int_val"])
		assert.Equal(t, "3.14", row.httpData["float_val"])
		assert.Equal(t, "true", row.httpData["bool_val"])
		assert.Equal(t, "", row.httpData["null_val"]) // nil becomes empty string
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkClickHouseLogStorage_buildQuery_Simple(b *testing.B) {
	storage := &ClickHouseLogStorage{}
	opts := LogQueryOptions{
		Category: LogCategoryHTTP,
		Levels:   []LogLevel{LogLevelInfo},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildQuery(opts)
	}
}

func BenchmarkClickHouseLogStorage_buildQuery_Complex(b *testing.B) {
	storage := &ClickHouseLogStorage{}
	userID := uuid.New().String()
	opts := LogQueryOptions{
		Category:         LogCategoryHTTP,
		Levels:           []LogLevel{LogLevelInfo, LogLevelWarn, LogLevelError},
		Component:        "auth",
		UserID:           userID,
		StartTime:        time.Now().Add(-24 * time.Hour),
		EndTime:          time.Now(),
		Search:           "failed login",
		HideStaticAssets: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildQuery(opts)
	}
}

func BenchmarkClickHouseLogStorage_toRow(b *testing.B) {
	storage := &ClickHouseLogStorage{}
	entry := &LogEntry{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Category:  LogCategoryHTTP,
		Level:     LogLevelInfo,
		Message:   "GET /api/users",
		Component: "api",
		UserID:    uuid.New().String(),
		RequestID: "req-123",
		TraceID:   "trace-456",
		Fields: map[string]any{
			"method":     "GET",
			"path":       "/api/users",
			"status":     "200",
			"duration":   "45ms",
			"user_agent": "test-agent",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.toRow(entry)
	}
}

func BenchmarkJoinConditions(b *testing.B) {
	conditions := []string{
		"category = ?",
		"level = ?",
		"timestamp >= ?",
		"component = ?",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = joinConditions(conditions)
	}
}

func BenchmarkJoinPlaceholders(b *testing.B) {
	placeholders := []string{"?", "?", "?", "?", "?"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = joinPlaceholders(placeholders)
	}
}
