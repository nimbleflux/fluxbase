package tools

import (
	"testing"
	"time"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Constants Tests
// =============================================================================

func TestExecuteSQLConstants(t *testing.T) {
	t.Run("defaultMaxRows is 100", func(t *testing.T) {
		assert.Equal(t, 100, defaultMaxRows)
	})

	t.Run("defaultQueryTimeout is 30 seconds", func(t *testing.T) {
		assert.Equal(t, 30*time.Second, defaultQueryTimeout)
	})
}

// =============================================================================
// NewExecuteSQLTool Tests
// =============================================================================

func TestNewExecuteSQLTool(t *testing.T) {
	t.Run("creates tool with defaults", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)

		require.NotNil(t, tool)
		assert.Equal(t, defaultMaxRows, tool.maxRows)
		assert.Equal(t, defaultQueryTimeout, tool.timeout)
	})

	t.Run("stores db connection", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)

		assert.Nil(t, tool.db)
	})
}

// =============================================================================
// NewExecuteSQLToolWithOptions Tests
// =============================================================================

func TestNewExecuteSQLToolWithOptions(t *testing.T) {
	t.Run("creates tool with custom options", func(t *testing.T) {
		customMaxRows := 500
		customTimeout := 60 * time.Second

		tool := NewExecuteSQLToolWithOptions(nil, customMaxRows, customTimeout)

		require.NotNil(t, tool)
		assert.Equal(t, customMaxRows, tool.maxRows)
		assert.Equal(t, customTimeout, tool.timeout)
	})

	t.Run("allows zero maxRows", func(t *testing.T) {
		tool := NewExecuteSQLToolWithOptions(nil, 0, time.Minute)

		assert.Equal(t, 0, tool.maxRows)
	})

	t.Run("allows zero timeout", func(t *testing.T) {
		tool := NewExecuteSQLToolWithOptions(nil, 100, 0)

		assert.Equal(t, time.Duration(0), tool.timeout)
	})

	t.Run("allows very large maxRows", func(t *testing.T) {
		tool := NewExecuteSQLToolWithOptions(nil, 1000000, time.Minute)

		assert.Equal(t, 1000000, tool.maxRows)
	})

	t.Run("allows very long timeout", func(t *testing.T) {
		longTimeout := 10 * time.Minute
		tool := NewExecuteSQLToolWithOptions(nil, 100, longTimeout)

		assert.Equal(t, longTimeout, tool.timeout)
	})
}

// =============================================================================
// ExecuteSQLTool Method Tests
// =============================================================================

func TestExecuteSQLTool_Name(t *testing.T) {
	t.Run("returns correct name", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		assert.Equal(t, "execute_sql", tool.Name())
	})
}

func TestExecuteSQLTool_Description(t *testing.T) {
	t.Run("returns non-empty description", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		desc := tool.Description()

		assert.NotEmpty(t, desc)
		assert.Contains(t, desc, "SQL")
	})

	t.Run("mentions read-only", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		desc := tool.Description()

		assert.Contains(t, desc, "read-only")
	})

	t.Run("mentions SELECT", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		desc := tool.Description()

		assert.Contains(t, desc, "SELECT")
	})
}

func TestExecuteSQLTool_InputSchema(t *testing.T) {
	tool := NewExecuteSQLTool(nil)
	schema := tool.InputSchema()

	t.Run("has object type", func(t *testing.T) {
		assert.Equal(t, "object", schema["type"])
	})

	t.Run("has sql property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		sqlProp, ok := properties["sql"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", sqlProp["type"])
		assert.NotEmpty(t, sqlProp["description"])
	})

	t.Run("has description property", func(t *testing.T) {
		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)

		descProp, ok := properties["description"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "string", descProp["type"])
		assert.NotEmpty(t, descProp["description"])
	})

	t.Run("requires sql and description", func(t *testing.T) {
		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "sql")
		assert.Contains(t, required, "description")
	})
}

func TestExecuteSQLTool_RequiredScopes(t *testing.T) {
	t.Run("requires execute_sql scope", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		scopes := tool.RequiredScopes()

		assert.Contains(t, scopes, mcp.ScopeExecuteSQL)
	})

	t.Run("returns single scope", func(t *testing.T) {
		tool := NewExecuteSQLTool(nil)
		scopes := tool.RequiredScopes()

		assert.Len(t, scopes, 1)
	})
}

// =============================================================================
// convertSQLValue Tests
// =============================================================================

func TestConvertSQLValue(t *testing.T) {
	t.Run("returns nil for nil input", func(t *testing.T) {
		result := convertSQLValue(nil)
		assert.Nil(t, result)
	})

	t.Run("converts time.Time to RFC3339 string", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := convertSQLValue(testTime)

		assert.Equal(t, "2024-01-15T10:30:00Z", result)
	})

	t.Run("converts time.Time with timezone", func(t *testing.T) {
		loc, _ := time.LoadLocation("America/New_York")
		testTime := time.Date(2024, 6, 15, 14, 30, 0, 0, loc)
		result := convertSQLValue(testTime)

		assert.IsType(t, "", result)
		assert.Contains(t, result.(string), "2024-06-15")
	})

	t.Run("converts byte slice to string", func(t *testing.T) {
		input := []byte("hello world")
		result := convertSQLValue(input)

		assert.Equal(t, "hello world", result)
	})

	t.Run("converts empty byte slice to empty string", func(t *testing.T) {
		input := []byte{}
		result := convertSQLValue(input)

		assert.Equal(t, "", result)
	})

	t.Run("passes through string unchanged", func(t *testing.T) {
		input := "test string"
		result := convertSQLValue(input)

		assert.Equal(t, "test string", result)
	})

	t.Run("passes through int unchanged", func(t *testing.T) {
		input := 42
		result := convertSQLValue(input)

		assert.Equal(t, 42, result)
	})

	t.Run("passes through int64 unchanged", func(t *testing.T) {
		input := int64(9999999999)
		result := convertSQLValue(input)

		assert.Equal(t, int64(9999999999), result)
	})

	t.Run("passes through float64 unchanged", func(t *testing.T) {
		input := 3.14159
		result := convertSQLValue(input)

		assert.Equal(t, 3.14159, result)
	})

	t.Run("passes through bool unchanged", func(t *testing.T) {
		assert.Equal(t, true, convertSQLValue(true))
		assert.Equal(t, false, convertSQLValue(false))
	})

	t.Run("passes through slice unchanged", func(t *testing.T) {
		input := []int{1, 2, 3}
		result := convertSQLValue(input)

		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("passes through map unchanged", func(t *testing.T) {
		input := map[string]int{"a": 1, "b": 2}
		result := convertSQLValue(input)

		assert.Equal(t, map[string]int{"a": 1, "b": 2}, result)
	})
}

// =============================================================================
// ExecuteSQLTool Struct Tests
// =============================================================================

func TestExecuteSQLTool_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		tool := &ExecuteSQLTool{
			db:      nil,
			maxRows: 200,
			timeout: time.Minute,
		}

		assert.Nil(t, tool.db)
		assert.Equal(t, 200, tool.maxRows)
		assert.Equal(t, time.Minute, tool.timeout)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		tool := &ExecuteSQLTool{}

		assert.Nil(t, tool.db)
		assert.Zero(t, tool.maxRows)
		assert.Zero(t, tool.timeout)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkConvertSQLValue_Nil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertSQLValue(nil)
	}
}

func BenchmarkConvertSQLValue_Time(b *testing.B) {
	testTime := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertSQLValue(testTime)
	}
}

func BenchmarkConvertSQLValue_Bytes(b *testing.B) {
	input := []byte("test string value")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertSQLValue(input)
	}
}

func BenchmarkConvertSQLValue_Passthrough(b *testing.B) {
	input := "test string"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convertSQLValue(input)
	}
}

func BenchmarkNewExecuteSQLTool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewExecuteSQLTool(nil)
	}
}

func BenchmarkExecuteSQLTool_InputSchema(b *testing.B) {
	tool := NewExecuteSQLTool(nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tool.InputSchema()
	}
}
