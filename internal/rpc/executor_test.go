package rpc

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/auth"
	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Executor Construction Tests
// =============================================================================

func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with nil dependencies", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		require.NotNil(t, executor)
		assert.Nil(t, executor.db)
		assert.Nil(t, executor.storage)
		assert.Nil(t, executor.metrics)
		assert.Nil(t, executor.config)
		assert.NotNil(t, executor.validator) // Validator should always be created
	})

	t.Run("creates executor with config", func(t *testing.T) {
		cfg := &config.RPCConfig{
			DefaultMaxRows: 500,
		}

		executor := NewExecutor(nil, nil, nil, cfg)

		require.NotNil(t, executor)
		assert.Equal(t, cfg, executor.config)
	})
}

// =============================================================================
// formatValue Tests - Critical for SQL injection prevention
// =============================================================================

func TestExecutor_formatValue(t *testing.T) {
	executor := NewExecutor(nil, nil, nil, nil)

	t.Run("formats nil as NULL", func(t *testing.T) {
		result := executor.formatValue(nil)
		assert.Equal(t, "NULL", result)
	})

	t.Run("formats string with escaping", func(t *testing.T) {
		result := executor.formatValue("hello")
		assert.Equal(t, "'hello'", result)
	})

	t.Run("escapes single quotes in strings - SQL injection prevention", func(t *testing.T) {
		result := executor.formatValue("O'Brien")
		assert.Equal(t, "'O''Brien'", result)
	})

	t.Run("escapes multiple single quotes", func(t *testing.T) {
		result := executor.formatValue("it's a 'test' string")
		assert.Equal(t, "'it''s a ''test'' string'", result)
	})

	t.Run("formats integer types", func(t *testing.T) {
		assert.Equal(t, "42", executor.formatValue(42))
		assert.Equal(t, "42", executor.formatValue(int32(42)))
		assert.Equal(t, "42", executor.formatValue(int64(42)))
	})

	t.Run("formats float types", func(t *testing.T) {
		result := executor.formatValue(3.14)
		assert.Equal(t, "3.14", result)

		result = executor.formatValue(float32(2.5))
		assert.Equal(t, "2.5", result)
	})

	t.Run("formats boolean true", func(t *testing.T) {
		result := executor.formatValue(true)
		assert.Equal(t, "TRUE", result)
	})

	t.Run("formats boolean false", func(t *testing.T) {
		result := executor.formatValue(false)
		assert.Equal(t, "FALSE", result)
	})

	t.Run("formats json.Number", func(t *testing.T) {
		num := json.Number("123.456")
		result := executor.formatValue(num)
		assert.Equal(t, "123.456", result)
	})

	t.Run("formats []float32 as vector literal", func(t *testing.T) {
		vec := []float32{0.1, 0.2, 0.3}
		result := executor.formatValue(vec)
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})

	t.Run("formats []float64 as vector literal", func(t *testing.T) {
		vec := []float64{0.1, 0.2, 0.3}
		result := executor.formatValue(vec)
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})

	t.Run("formats numeric []interface{} as vector literal", func(t *testing.T) {
		vec := []interface{}{0.1, 0.2, 0.3}
		result := executor.formatValue(vec)
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})

	t.Run("formats non-numeric []interface{} as ARRAY", func(t *testing.T) {
		arr := []interface{}{"a", "b", "c"}
		result := executor.formatValue(arr)
		assert.Contains(t, result, "ARRAY[")
		assert.Contains(t, result, "'a'")
		assert.Contains(t, result, "'b'")
		assert.Contains(t, result, "'c'")
	})

	t.Run("formats map as JSONB", func(t *testing.T) {
		m := map[string]interface{}{"key": "value", "num": 42}
		result := executor.formatValue(m)
		assert.Contains(t, result, "::jsonb")
		assert.Contains(t, result, "key")
		assert.Contains(t, result, "value")
	})

	t.Run("escapes single quotes in JSONB", func(t *testing.T) {
		m := map[string]interface{}{"name": "O'Brien"}
		result := executor.formatValue(m)
		assert.Contains(t, result, "::jsonb")
		// The JSON itself will have the single quote, but it's escaped in SQL
		assert.Contains(t, result, "O''Brien")
	})

	t.Run("formats unknown types as escaped strings", func(t *testing.T) {
		type customType struct{ value int }
		result := executor.formatValue(customType{value: 42})
		assert.Contains(t, result, "'")
	})
}

// =============================================================================
// buildSQL Tests
// =============================================================================

func TestExecutor_buildSQL(t *testing.T) {
	executor := NewExecutor(nil, nil, nil, nil)

	t.Run("substitutes simple parameter", func(t *testing.T) {
		execCtx := &ExecuteContext{
			UserID:   "user-123",
			UserRole: "admin",
			Params:   map[string]interface{}{"name": "test"},
		}

		sql, err := executor.buildSQL("SELECT * FROM users WHERE name = $name", execCtx.Params, execCtx)

		require.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE name = 'test'", sql)
	})

	t.Run("substitutes multiple parameters", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"name":   "John",
				"age":    30,
				"active": true,
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE name = $name AND age > $age AND active = $active",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		assert.Contains(t, sql, "'John'")
		assert.Contains(t, sql, "30")
		assert.Contains(t, sql, "TRUE")
	})

	t.Run("adds caller context parameters", func(t *testing.T) {
		execCtx := &ExecuteContext{
			UserID:    "user-123",
			UserRole:  "admin",
			UserEmail: "admin@example.com",
			Params:    map[string]interface{}{},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE created_by = $caller_id",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		assert.Contains(t, sql, "'user-123'")
	})

	t.Run("user params override caller params", func(t *testing.T) {
		execCtx := &ExecuteContext{
			UserID: "default-user",
			Params: map[string]interface{}{
				"caller_id": "override-user",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE id = $caller_id",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		assert.Contains(t, sql, "'override-user'")
		assert.NotContains(t, sql, "default-user")
	})

	t.Run("returns error for missing parameters", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{},
		}

		_, err := executor.buildSQL(
			"SELECT * FROM users WHERE name = $missing_param",
			execCtx.Params,
			execCtx,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required parameters")
		assert.Contains(t, err.Error(), "missing_param")
	})

	t.Run("returns error for multiple missing parameters", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{},
		}

		_, err := executor.buildSQL(
			"SELECT * FROM users WHERE name = $param1 AND age = $param2",
			execCtx.Params,
			execCtx,
		)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "param1")
		assert.Contains(t, err.Error(), "param2")
	})

	t.Run("handles parameter with underscore", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"user_name": "test",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE user_name = $user_name",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		assert.Contains(t, sql, "'test'")
	})

	t.Run("handles parameter with numbers", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"param1": "value1",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE col = $param1",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		assert.Contains(t, sql, "'value1'")
	})
}

// =============================================================================
// Vector Literal Formatting Tests
// =============================================================================

func TestFormatVectorLiteral32(t *testing.T) {
	t.Run("formats empty vector", func(t *testing.T) {
		result := formatVectorLiteral32([]float32{})
		assert.Equal(t, "'[]'::vector", result)
	})

	t.Run("formats single element", func(t *testing.T) {
		result := formatVectorLiteral32([]float32{0.5})
		assert.Equal(t, "'[0.5]'::vector", result)
	})

	t.Run("formats multiple elements", func(t *testing.T) {
		result := formatVectorLiteral32([]float32{0.1, 0.2, 0.3})
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})

	t.Run("handles scientific notation", func(t *testing.T) {
		result := formatVectorLiteral32([]float32{1e-10, 1e10})
		assert.Contains(t, result, "'[")
		assert.Contains(t, result, "]'::vector")
	})
}

func TestFormatVectorLiteral64(t *testing.T) {
	t.Run("formats empty vector", func(t *testing.T) {
		result := formatVectorLiteral64([]float64{})
		assert.Equal(t, "'[]'::vector", result)
	})

	t.Run("formats single element", func(t *testing.T) {
		result := formatVectorLiteral64([]float64{0.5})
		assert.Equal(t, "'[0.5]'::vector", result)
	})

	t.Run("formats multiple elements", func(t *testing.T) {
		result := formatVectorLiteral64([]float64{0.1, 0.2, 0.3})
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})
}

func TestFormatVectorLiteralInterface(t *testing.T) {
	t.Run("formats float64 elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{0.1, 0.2, 0.3})
		assert.Equal(t, "'[0.1,0.2,0.3]'::vector", result)
	})

	t.Run("formats int elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{1, 2, 3})
		assert.Equal(t, "'[1,2,3]'::vector", result)
	})

	t.Run("formats int64 elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{int64(1), int64(2)})
		assert.Equal(t, "'[1,2]'::vector", result)
	})

	t.Run("formats json.Number elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{json.Number("1.5"), json.Number("2.5")})
		assert.Equal(t, "'[1.5,2.5]'::vector", result)
	})

	t.Run("formats mixed numeric types", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{1, 2.5, int64(3)})
		assert.Contains(t, result, "'[")
		assert.Contains(t, result, "]'::vector")
	})

	t.Run("formats float32 elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{float32(0.1), float32(0.2)})
		assert.Equal(t, "'[0.1,0.2]'::vector", result)
	})

	t.Run("formats int32 elements", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{int32(1), int32(2)})
		assert.Equal(t, "'[1,2]'::vector", result)
	})

	t.Run("handles unknown types with default formatting", func(t *testing.T) {
		// This tests the default case in the switch statement
		// Using uint which would fall into the default case
		result := formatVectorLiteralInterface([]interface{}{uint(1), uint(2)})
		assert.Contains(t, result, "'[")
		assert.Contains(t, result, "]'::vector")
		assert.Contains(t, result, "1")
		assert.Contains(t, result, "2")
	})

	t.Run("handles mixed types with float32", func(t *testing.T) {
		result := formatVectorLiteralInterface([]interface{}{float32(1.5), int64(2), float64(3.5)})
		assert.Contains(t, result, "'[")
		assert.Contains(t, result, "]'::vector")
	})
}

// =============================================================================
// isNumericArray Tests
// =============================================================================

func TestIsNumericArray(t *testing.T) {
	t.Run("returns false for empty array", func(t *testing.T) {
		result := isNumericArray([]interface{}{})
		assert.False(t, result)
	})

	t.Run("returns true for float64 array", func(t *testing.T) {
		result := isNumericArray([]interface{}{1.0, 2.0, 3.0})
		assert.True(t, result)
	})

	t.Run("returns true for int array", func(t *testing.T) {
		result := isNumericArray([]interface{}{1, 2, 3})
		assert.True(t, result)
	})

	t.Run("returns true for int64 array", func(t *testing.T) {
		result := isNumericArray([]interface{}{int64(1), int64(2)})
		assert.True(t, result)
	})

	t.Run("returns true for json.Number array", func(t *testing.T) {
		result := isNumericArray([]interface{}{json.Number("1"), json.Number("2")})
		assert.True(t, result)
	})

	t.Run("returns true for mixed numeric types", func(t *testing.T) {
		result := isNumericArray([]interface{}{1, 2.0, int64(3), json.Number("4")})
		assert.True(t, result)
	})

	t.Run("returns false for string array", func(t *testing.T) {
		result := isNumericArray([]interface{}{"a", "b", "c"})
		assert.False(t, result)
	})

	t.Run("returns false for mixed array with strings", func(t *testing.T) {
		result := isNumericArray([]interface{}{1, "two", 3})
		assert.False(t, result)
	})

	t.Run("returns false for boolean array", func(t *testing.T) {
		result := isNumericArray([]interface{}{true, false})
		assert.False(t, result)
	})

	t.Run("returns false for nil elements", func(t *testing.T) {
		result := isNumericArray([]interface{}{1, nil, 3})
		assert.False(t, result)
	})
}

// =============================================================================
// convertValue Tests
// =============================================================================

func TestConvertValue(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		result := convertValue(nil)
		assert.Nil(t, result)
	})

	t.Run("converts byte slice to string", func(t *testing.T) {
		result := convertValue([]byte("hello"))
		assert.Equal(t, "hello", result)
	})

	t.Run("parses JSON byte slice", func(t *testing.T) {
		result := convertValue([]byte(`{"key": "value"}`))
		assert.NotNil(t, result)
		m, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "value", m["key"])
	})

	t.Run("parses JSON array byte slice", func(t *testing.T) {
		result := convertValue([]byte(`[1, 2, 3]`))
		assert.NotNil(t, result)
		arr, ok := result.([]interface{})
		assert.True(t, ok)
		assert.Len(t, arr, 3)
	})

	t.Run("converts time.Time to RFC3339", func(t *testing.T) {
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := convertValue(now)
		assert.Equal(t, "2024-01-15T10:30:00Z", result)
	})

	t.Run("passes through other types", func(t *testing.T) {
		assert.Equal(t, 42, convertValue(42))
		assert.Equal(t, "string", convertValue("string"))
		assert.Equal(t, true, convertValue(true))
		assert.Equal(t, 3.14, convertValue(3.14))
	})
}

// =============================================================================
// ExecuteContext Tests
// =============================================================================

func TestExecuteContext(t *testing.T) {
	t.Run("creates context with all fields", func(t *testing.T) {
		claims := &auth.TokenClaims{
			UserID: "user-123",
			Role:   "admin",
		}

		ctx := &ExecuteContext{
			Procedure: &Procedure{
				ID:        "proc-1",
				Name:      "test_proc",
				Namespace: "public",
			},
			Params:               map[string]interface{}{"key": "value"},
			UserID:               "user-123",
			UserRole:             "admin",
			UserEmail:            "admin@example.com",
			Claims:               claims,
			IsAsync:              true,
			ExecutionID:          "exec-456",
			DisableExecutionLogs: false,
		}

		assert.Equal(t, "test_proc", ctx.Procedure.Name)
		assert.Equal(t, "user-123", ctx.UserID)
		assert.True(t, ctx.IsAsync)
	})
}

// =============================================================================
// ExecuteResult Tests
// =============================================================================

func TestExecuteResult(t *testing.T) {
	t.Run("creates result with all fields", func(t *testing.T) {
		rowCount := 10
		duration := 50
		result := &ExecuteResult{
			ExecutionID:  "exec-123",
			Status:       StatusCompleted,
			Result:       json.RawMessage(`[{"id": 1}]`),
			RowsReturned: &rowCount,
			DurationMs:   &duration,
		}

		assert.Equal(t, "exec-123", result.ExecutionID)
		assert.Equal(t, StatusCompleted, result.Status)
		assert.Equal(t, 10, *result.RowsReturned)
		assert.Equal(t, 50, *result.DurationMs)
	})

	t.Run("creates error result", func(t *testing.T) {
		errorMsg := "query failed"
		duration := 10
		result := &ExecuteResult{
			ExecutionID: "exec-456",
			Status:      StatusFailed,
			DurationMs:  &duration,
			Error:       &errorMsg,
		}

		assert.Equal(t, StatusFailed, result.Status)
		assert.Equal(t, "query failed", *result.Error)
		assert.Nil(t, result.Result)
	})
}

// =============================================================================
// SQL Injection Prevention Tests
// =============================================================================

func TestExecutor_SQLInjectionPrevention(t *testing.T) {
	executor := NewExecutor(nil, nil, nil, nil)

	t.Run("escapes classic SQL injection", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"name": "'; DROP TABLE users; --",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM users WHERE name = $name",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		// Single quotes should be escaped
		assert.Contains(t, sql, "''")
		// The entire malicious string is quoted and escaped
		assert.Contains(t, sql, "'''; DROP TABLE users; --'")
	})

	t.Run("escapes union-based injection", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"id": "1' UNION SELECT password FROM admin --",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM products WHERE id = $id",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		// The entire input is treated as a string
		assert.Contains(t, sql, "'1'' UNION SELECT password FROM admin --'")
	})

	t.Run("escapes nested quotes", func(t *testing.T) {
		execCtx := &ExecuteContext{
			Params: map[string]interface{}{
				"input": "test''injection",
			},
		}

		sql, err := executor.buildSQL(
			"SELECT * FROM data WHERE val = $input",
			execCtx.Params,
			execCtx,
		)

		require.NoError(t, err)
		// Already escaped quotes should be double-escaped
		assert.Contains(t, sql, "''''")
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkExecutor_formatValue_String(b *testing.B) {
	executor := NewExecutor(nil, nil, nil, nil)
	value := "test string with 'quotes'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.formatValue(value)
	}
}

func BenchmarkExecutor_formatValue_Vector(b *testing.B) {
	executor := NewExecutor(nil, nil, nil, nil)
	value := []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.formatValue(value)
	}
}

func BenchmarkExecutor_formatValue_Map(b *testing.B) {
	executor := NewExecutor(nil, nil, nil, nil)
	value := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.formatValue(value)
	}
}

func BenchmarkExecutor_buildSQL_Simple(b *testing.B) {
	executor := NewExecutor(nil, nil, nil, nil)
	execCtx := &ExecuteContext{
		UserID:   "user-123",
		UserRole: "admin",
		Params:   map[string]interface{}{"name": "test"},
	}
	template := "SELECT * FROM users WHERE name = $name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.buildSQL(template, execCtx.Params, execCtx)
	}
}

func BenchmarkExecutor_buildSQL_Complex(b *testing.B) {
	executor := NewExecutor(nil, nil, nil, nil)
	execCtx := &ExecuteContext{
		UserID:    "user-123",
		UserRole:  "admin",
		UserEmail: "admin@example.com",
		Params: map[string]interface{}{
			"name":   "test",
			"age":    30,
			"active": true,
			"tags":   []interface{}{"a", "b", "c"},
			"meta":   map[string]interface{}{"key": "value"},
		},
	}
	template := "SELECT * FROM users WHERE name = $name AND age > $age AND active = $active AND created_by = $caller_id"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = executor.buildSQL(template, execCtx.Params, execCtx)
	}
}

func BenchmarkIsNumericArray(b *testing.B) {
	arr := []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = isNumericArray(arr)
	}
}

func BenchmarkConvertValue_JSON(b *testing.B) {
	data := []byte(`{"key": "value", "number": 42, "nested": {"a": 1}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertValue(data)
	}
}

// =============================================================================
// failExecutionWithContext Tests
// =============================================================================

func TestExecutor_failExecutionWithContext(t *testing.T) {
	t.Run("marks execution as failed", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		exec := &Execution{
			ID:            "exec-123",
			ProcedureName: "test_proc",
			Status:        StatusRunning,
		}

		execCtx := &ExecuteContext{
			Procedure:            &Procedure{},
			DisableExecutionLogs: true, // Skip storage updates
			IsAsync:              false,
		}

		start := time.Now()
		result, err := executor.failExecutionWithContext(context.Background(), exec, execCtx, start, "test error")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "exec-123", result.ExecutionID)
		assert.Equal(t, StatusFailed, result.Status)
		assert.NotNil(t, result.Error)
		assert.Equal(t, "test error", *result.Error)
	})

	t.Run("skips storage updates when logs are disabled", func(t *testing.T) {
		storage := &Storage{}
		executor := NewExecutor(nil, storage, nil, nil)

		exec := &Execution{
			ID:            "exec-123",
			ProcedureName: "test_proc",
			Status:        StatusRunning,
		}

		execCtx := &ExecuteContext{
			Procedure:            &Procedure{},
			DisableExecutionLogs: true,
			IsAsync:              false,
		}

		start := time.Now()
		_, err := executor.failExecutionWithContext(context.Background(), exec, execCtx, start, "test error")

		assert.NoError(t, err)
		assert.Equal(t, StatusFailed, exec.Status)
	})

	t.Run("sets duration and completion time", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		exec := &Execution{
			ID:            "exec-123",
			ProcedureName: "test_proc",
			Status:        StatusRunning,
		}

		execCtx := &ExecuteContext{
			Procedure:            &Procedure{},
			DisableExecutionLogs: true,
			IsAsync:              false,
		}

		start := time.Now()
		time.Sleep(10 * time.Millisecond) // Ensure some duration
		result, err := executor.failExecutionWithContext(context.Background(), exec, execCtx, start, "test error")

		assert.NoError(t, err)
		assert.NotNil(t, result.DurationMs)
		assert.Greater(t, *result.DurationMs, 0)
		assert.NotNil(t, exec.CompletedAt)
	})
}

// =============================================================================
// appendLog Tests
// =============================================================================

func TestExecutor_appendLog(t *testing.T) {
	t.Run("logs execution info", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		// Should not panic
		executor.appendLog(context.Background(), "exec-123", 1, "info", "test message")
	})

	t.Run("logs execution errors", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		// Should not panic
		executor.appendLog(context.Background(), "exec-123", 99, "error", "test error")
	})
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

func TestExecutor_EdgeCases(t *testing.T) {
	t.Run("handles nil config", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		assert.Nil(t, executor.config)
	})

	t.Run("handles nil storage", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		assert.Nil(t, executor.storage)
	})

	t.Run("handles nil metrics", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		assert.Nil(t, executor.metrics)
	})

	t.Run("handles empty procedure name", func(t *testing.T) {
		executor := NewExecutor(nil, nil, nil, nil)

		proc := &Procedure{
			ID:        uuid.New().String(),
			Name:      "",
			Namespace: "public",
			SQLQuery:  "SELECT 1",
		}

		execCtx := &ExecuteContext{
			Procedure: proc,
			Params:    map[string]interface{}{},
		}

		_, err := executor.buildSQL("SELECT 1", execCtx.Params, execCtx)

		assert.NoError(t, err)
	})

	t.Run("handles procedure with no allowed tables", func(t *testing.T) {
		validator := NewValidator()
		result := validator.ValidateSQL("SELECT 1", []string{}, []string{})

		// SELECT 1 should be valid (no tables accessed)
		assert.True(t, result.Valid)
	})
}

// =============================================================================
// ExecuteContext Field Tests
// =============================================================================

func TestExecuteContext_Fields(t *testing.T) {
	t.Run("creates empty context", func(t *testing.T) {
		execCtx := &ExecuteContext{}

		assert.Nil(t, execCtx.Procedure)
		assert.Nil(t, execCtx.Params)
		assert.Empty(t, execCtx.UserID)
		assert.Empty(t, execCtx.UserRole)
		assert.Empty(t, execCtx.UserEmail)
		assert.Nil(t, execCtx.Claims)
		assert.False(t, execCtx.IsAsync)
		assert.Empty(t, execCtx.ExecutionID)
		assert.False(t, execCtx.DisableExecutionLogs)
	})

	t.Run("creates context with all fields", func(t *testing.T) {
		claims := &auth.TokenClaims{
			UserID: "user-123",
			Role:   "authenticated",
		}

		proc := &Procedure{
			ID:        uuid.New().String(),
			Name:      "test_proc",
			Namespace: "public",
		}

		execCtx := &ExecuteContext{
			Procedure:            proc,
			Params:               map[string]interface{}{"key": "value"},
			UserID:               "user-123",
			UserRole:             "authenticated",
			UserEmail:            "test@example.com",
			Claims:               claims,
			IsAsync:              true,
			ExecutionID:          "exec-456",
			DisableExecutionLogs: true,
		}

		assert.Equal(t, proc, execCtx.Procedure)
		assert.NotEmpty(t, execCtx.Params)
		assert.Equal(t, "user-123", execCtx.UserID)
		assert.Equal(t, "authenticated", execCtx.UserRole)
		assert.Equal(t, "test@example.com", execCtx.UserEmail)
		assert.Equal(t, claims, execCtx.Claims)
		assert.True(t, execCtx.IsAsync)
		assert.Equal(t, "exec-456", execCtx.ExecutionID)
		assert.True(t, execCtx.DisableExecutionLogs)
	})
}

// =============================================================================
// ExecuteResult Field Tests
// =============================================================================

func TestExecuteResult_Fields(t *testing.T) {
	t.Run("creates result with minimal fields", func(t *testing.T) {
		result := &ExecuteResult{
			ExecutionID: "exec-123",
			Status:      StatusCompleted,
		}

		assert.Equal(t, "exec-123", result.ExecutionID)
		assert.Equal(t, StatusCompleted, result.Status)
		assert.Nil(t, result.Result)
		assert.Nil(t, result.RowsReturned)
		assert.Nil(t, result.DurationMs)
		assert.Nil(t, result.Error)
	})

	t.Run("creates result with all fields", func(t *testing.T) {
		rowCount := 10
		duration := 50
		errorMsg := "test error"

		result := &ExecuteResult{
			ExecutionID:  "exec-123",
			Status:       StatusFailed,
			Result:       json.RawMessage(`[{"id": 1}]`),
			RowsReturned: &rowCount,
			DurationMs:   &duration,
			Error:        &errorMsg,
		}

		assert.Equal(t, "exec-123", result.ExecutionID)
		assert.Equal(t, StatusFailed, result.Status)
		assert.NotNil(t, result.Result)
		assert.Equal(t, 10, *result.RowsReturned)
		assert.Equal(t, 50, *result.DurationMs)
		assert.Equal(t, "test error", *result.Error)
	})
}
