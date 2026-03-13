package tools

import (
	"testing"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// quoteIdentifier Tests
// =============================================================================

func TestQuoteIdentifier(t *testing.T) {
	t.Run("quotes simple identifier", func(t *testing.T) {
		result := quoteIdentifier("users")
		assert.Equal(t, `"users"`, result)
	})

	t.Run("quotes identifier with underscore", func(t *testing.T) {
		result := quoteIdentifier("user_accounts")
		assert.Equal(t, `"user_accounts"`, result)
	})

	t.Run("escapes double quotes", func(t *testing.T) {
		result := quoteIdentifier(`user"name`)
		assert.Equal(t, `"user""name"`, result)
	})

	t.Run("escapes multiple double quotes", func(t *testing.T) {
		result := quoteIdentifier(`a"b"c`)
		assert.Equal(t, `"a""b""c"`, result)
	})

	t.Run("handles empty string", func(t *testing.T) {
		result := quoteIdentifier("")
		assert.Equal(t, `""`, result)
	})

	t.Run("handles numeric string", func(t *testing.T) {
		result := quoteIdentifier("123")
		assert.Equal(t, `"123"`, result)
	})

	t.Run("handles reserved keywords", func(t *testing.T) {
		result := quoteIdentifier("select")
		assert.Equal(t, `"select"`, result)
	})

	t.Run("handles mixed case", func(t *testing.T) {
		result := quoteIdentifier("UserName")
		assert.Equal(t, `"UserName"`, result)
	})
}

// =============================================================================
// columnExists Tests
// =============================================================================

func TestColumnExists(t *testing.T) {
	tableInfo := &database.TableInfo{
		Columns: []database.ColumnInfo{
			{Name: "id"},
			{Name: "name"},
			{Name: "email"},
			{Name: "created_at"},
		},
	}

	t.Run("returns true for existing column", func(t *testing.T) {
		assert.True(t, columnExists(tableInfo, "id"))
		assert.True(t, columnExists(tableInfo, "name"))
		assert.True(t, columnExists(tableInfo, "email"))
		assert.True(t, columnExists(tableInfo, "created_at"))
	})

	t.Run("returns false for non-existent column", func(t *testing.T) {
		assert.False(t, columnExists(tableInfo, "nonexistent"))
		assert.False(t, columnExists(tableInfo, "ID")) // case sensitive
		assert.False(t, columnExists(tableInfo, "Name"))
	})

	t.Run("returns false for empty column name", func(t *testing.T) {
		assert.False(t, columnExists(tableInfo, ""))
	})

	t.Run("handles empty columns list", func(t *testing.T) {
		emptyTable := &database.TableInfo{Columns: []database.ColumnInfo{}}
		assert.False(t, columnExists(emptyTable, "id"))
	})

	t.Run("handles nil columns", func(t *testing.T) {
		nilTable := &database.TableInfo{}
		assert.False(t, columnExists(nilTable, "id"))
	})
}

// =============================================================================
// validateAndQuoteReturning Tests
// =============================================================================

func TestValidateAndQuoteReturning(t *testing.T) {
	tableInfo := &database.TableInfo{
		Columns: []database.ColumnInfo{
			{Name: "id"},
			{Name: "name"},
			{Name: "email"},
		},
	}

	t.Run("returns star for empty string", func(t *testing.T) {
		result, err := validateAndQuoteReturning("", nil)
		require.NoError(t, err)
		assert.Equal(t, "*", result)
	})

	t.Run("returns star for star input", func(t *testing.T) {
		result, err := validateAndQuoteReturning("*", nil)
		require.NoError(t, err)
		assert.Equal(t, "*", result)
	})

	t.Run("returns star for whitespace-only input", func(t *testing.T) {
		result, err := validateAndQuoteReturning("   ", nil)
		require.NoError(t, err)
		assert.Equal(t, "*", result)
	})

	t.Run("quotes single column", func(t *testing.T) {
		result, err := validateAndQuoteReturning("id", nil)
		require.NoError(t, err)
		assert.Equal(t, `"id"`, result)
	})

	t.Run("quotes multiple columns", func(t *testing.T) {
		result, err := validateAndQuoteReturning("id,name,email", nil)
		require.NoError(t, err)
		assert.Equal(t, `"id", "name", "email"`, result)
	})

	t.Run("trims whitespace around columns", func(t *testing.T) {
		result, err := validateAndQuoteReturning("  id  ,  name  ", nil)
		require.NoError(t, err)
		assert.Equal(t, `"id", "name"`, result)
	})

	t.Run("skips empty columns", func(t *testing.T) {
		result, err := validateAndQuoteReturning("id,,name", nil)
		require.NoError(t, err)
		assert.Equal(t, `"id", "name"`, result)
	})

	t.Run("returns star for only empty columns", func(t *testing.T) {
		result, err := validateAndQuoteReturning(",,,", nil)
		require.NoError(t, err)
		assert.Equal(t, "*", result)
	})

	t.Run("rejects invalid identifier format", func(t *testing.T) {
		_, err := validateAndQuoteReturning("1invalid", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid column name")
	})

	t.Run("rejects SQL injection attempt", func(t *testing.T) {
		_, err := validateAndQuoteReturning("id; DROP TABLE users", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid column name")
	})

	t.Run("rejects column with special characters", func(t *testing.T) {
		_, err := validateAndQuoteReturning("user-name", nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid column name")
	})

	t.Run("validates column exists in table", func(t *testing.T) {
		result, err := validateAndQuoteReturning("id,name", tableInfo)
		require.NoError(t, err)
		assert.Equal(t, `"id", "name"`, result)
	})

	t.Run("rejects unknown column when table info provided", func(t *testing.T) {
		_, err := validateAndQuoteReturning("unknown_column", tableInfo)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown column")
	})

	t.Run("skips column validation when table info nil", func(t *testing.T) {
		result, err := validateAndQuoteReturning("any_column", nil)
		require.NoError(t, err)
		assert.Equal(t, `"any_column"`, result)
	})
}

// =============================================================================
// parseFilterToSQL Tests
// =============================================================================

func TestParseFilterToSQL(t *testing.T) {
	t.Run("parses eq operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("id", "eq.123", 1)
		require.NoError(t, err)
		assert.Equal(t, `"id" = $1`, clause)
		assert.Equal(t, "123", value)
	})

	t.Run("parses neq operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("status", "neq.active", 2)
		require.NoError(t, err)
		assert.Equal(t, `"status" != $2`, clause)
		assert.Equal(t, "active", value)
	})

	t.Run("parses gt operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("age", "gt.18", 3)
		require.NoError(t, err)
		assert.Equal(t, `"age" > $3`, clause)
		assert.Equal(t, "18", value)
	})

	t.Run("parses gte operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("score", "gte.100", 4)
		require.NoError(t, err)
		assert.Equal(t, `"score" >= $4`, clause)
		assert.Equal(t, "100", value)
	})

	t.Run("parses lt operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("price", "lt.50", 5)
		require.NoError(t, err)
		assert.Equal(t, `"price" < $5`, clause)
		assert.Equal(t, "50", value)
	})

	t.Run("parses lte operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("count", "lte.1000", 6)
		require.NoError(t, err)
		assert.Equal(t, `"count" <= $6`, clause)
		assert.Equal(t, "1000", value)
	})

	t.Run("parses like operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("name", "like.%john%", 7)
		require.NoError(t, err)
		assert.Equal(t, `"name" LIKE $7`, clause)
		assert.Equal(t, "%john%", value)
	})

	t.Run("parses ilike operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("email", "ilike.%@example.com", 8)
		require.NoError(t, err)
		assert.Equal(t, `"email" ILIKE $8`, clause)
		assert.Equal(t, "%@example.com", value)
	})

	t.Run("parses is null", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("deleted_at", "is.null", 9)
		require.NoError(t, err)
		assert.Equal(t, `"deleted_at" IS NULL`, clause)
		assert.Nil(t, value)
	})

	t.Run("parses is true", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("active", "is.true", 10)
		require.NoError(t, err)
		assert.Equal(t, `"active" IS TRUE`, clause)
		assert.Nil(t, value)
	})

	t.Run("parses is false", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("verified", "is.false", 11)
		require.NoError(t, err)
		assert.Equal(t, `"verified" IS FALSE`, clause)
		assert.Nil(t, value)
	})

	t.Run("is operator case insensitive", func(t *testing.T) {
		clause, _, err := parseFilterToSQL("col", "is.NULL", 1)
		require.NoError(t, err)
		assert.Equal(t, `"col" IS NULL`, clause)

		clause, _, err = parseFilterToSQL("col", "is.TRUE", 1)
		require.NoError(t, err)
		assert.Equal(t, `"col" IS TRUE`, clause)

		clause, _, err = parseFilterToSQL("col", "is.FALSE", 1)
		require.NoError(t, err)
		assert.Equal(t, `"col" IS FALSE`, clause)
	})

	t.Run("parses in operator", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("status", "in.(active,pending,completed)", 12)
		require.NoError(t, err)
		assert.Equal(t, `"status" = ANY($12)`, clause)
		assert.Equal(t, []string{"active", "pending", "completed"}, value)
	})

	t.Run("parses in operator without parentheses", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("id", "in.1,2,3", 13)
		require.NoError(t, err)
		assert.Equal(t, `"id" = ANY($13)`, clause)
		assert.Equal(t, []string{"1", "2", "3"}, value)
	})

	t.Run("returns error for invalid format", func(t *testing.T) {
		_, _, err := parseFilterToSQL("id", "nooperator", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "expected format")
	})

	t.Run("returns error for unsupported operator", func(t *testing.T) {
		_, _, err := parseFilterToSQL("id", "contains.value", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported operator")
	})

	t.Run("returns error for invalid is value", func(t *testing.T) {
		_, _, err := parseFilterToSQL("col", "is.invalid", 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid 'is' value")
	})

	t.Run("quotes column names with special chars", func(t *testing.T) {
		clause, _, err := parseFilterToSQL("user_name", "eq.test", 1)
		require.NoError(t, err)
		assert.Equal(t, `"user_name" = $1`, clause)
	})

	t.Run("handles empty value", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("name", "eq.", 1)
		require.NoError(t, err)
		assert.Equal(t, `"name" = $1`, clause)
		assert.Equal(t, "", value)
	})

	t.Run("handles value with dots", func(t *testing.T) {
		clause, value, err := parseFilterToSQL("email", "eq.user@example.com", 1)
		require.NoError(t, err)
		assert.Equal(t, `"email" = $1`, clause)
		assert.Equal(t, "user@example.com", value)
	})
}

// =============================================================================
// validIdentifierRegex Tests
// =============================================================================

func TestValidIdentifierRegex(t *testing.T) {
	t.Run("matches valid identifiers", func(t *testing.T) {
		valid := []string{
			"users",
			"_private",
			"Table1",
			"user_accounts",
			"a",
			"_",
			"A1B2C3",
		}
		for _, name := range valid {
			assert.True(t, validIdentifierRegex.MatchString(name), "should match: %s", name)
		}
	})

	t.Run("rejects invalid identifiers", func(t *testing.T) {
		invalid := []string{
			"1users",     // starts with number
			"user-name",  // hyphen
			"user.name",  // dot
			"user name",  // space
			"user@email", // @
			"",           // empty
		}
		for _, name := range invalid {
			assert.False(t, validIdentifierRegex.MatchString(name), "should not match: %s", name)
		}
	})
}

// =============================================================================
// InsertRecordTool Tests
// =============================================================================

func TestInsertRecordTool(t *testing.T) {
	t.Run("Name returns correct name", func(t *testing.T) {
		tool := NewInsertRecordTool(nil, nil)
		assert.Equal(t, "insert_record", tool.Name())
	})

	t.Run("Description returns non-empty string", func(t *testing.T) {
		tool := NewInsertRecordTool(nil, nil)
		assert.NotEmpty(t, tool.Description())
		assert.Contains(t, tool.Description(), "Insert")
	})

	t.Run("InputSchema has required properties", func(t *testing.T) {
		tool := NewInsertRecordTool(nil, nil)
		schema := tool.InputSchema()

		assert.Equal(t, "object", schema["type"])

		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, properties, "table")
		assert.Contains(t, properties, "data")
		assert.Contains(t, properties, "returning")

		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "data")
	})

	t.Run("RequiredScopes returns write scope", func(t *testing.T) {
		tool := NewInsertRecordTool(nil, nil)
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeWriteTables)
	})
}

// =============================================================================
// UpdateRecordTool Tests
// =============================================================================

func TestUpdateRecordTool(t *testing.T) {
	t.Run("Name returns correct name", func(t *testing.T) {
		tool := NewUpdateRecordTool(nil, nil)
		assert.Equal(t, "update_record", tool.Name())
	})

	t.Run("Description returns non-empty string", func(t *testing.T) {
		tool := NewUpdateRecordTool(nil, nil)
		assert.NotEmpty(t, tool.Description())
		assert.Contains(t, tool.Description(), "Update")
	})

	t.Run("InputSchema has required properties", func(t *testing.T) {
		tool := NewUpdateRecordTool(nil, nil)
		schema := tool.InputSchema()

		assert.Equal(t, "object", schema["type"])

		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, properties, "table")
		assert.Contains(t, properties, "data")
		assert.Contains(t, properties, "filter")
		assert.Contains(t, properties, "returning")

		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "data")
		assert.Contains(t, required, "filter")
	})

	t.Run("RequiredScopes returns write scope", func(t *testing.T) {
		tool := NewUpdateRecordTool(nil, nil)
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeWriteTables)
	})
}

// =============================================================================
// DeleteRecordTool Tests
// =============================================================================

func TestDeleteRecordTool(t *testing.T) {
	t.Run("Name returns correct name", func(t *testing.T) {
		tool := NewDeleteRecordTool(nil, nil)
		assert.Equal(t, "delete_record", tool.Name())
	})

	t.Run("Description returns non-empty string", func(t *testing.T) {
		tool := NewDeleteRecordTool(nil, nil)
		assert.NotEmpty(t, tool.Description())
		assert.Contains(t, tool.Description(), "Delete")
	})

	t.Run("InputSchema has required properties", func(t *testing.T) {
		tool := NewDeleteRecordTool(nil, nil)
		schema := tool.InputSchema()

		assert.Equal(t, "object", schema["type"])

		properties, ok := schema["properties"].(map[string]any)
		require.True(t, ok)
		assert.Contains(t, properties, "table")
		assert.Contains(t, properties, "filter")
		assert.Contains(t, properties, "returning")

		required, ok := schema["required"].([]string)
		require.True(t, ok)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "filter")
	})

	t.Run("RequiredScopes returns write scope", func(t *testing.T) {
		tool := NewDeleteRecordTool(nil, nil)
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeWriteTables)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkQuoteIdentifier(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quoteIdentifier("user_accounts")
	}
}

func BenchmarkQuoteIdentifier_WithEscaping(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		quoteIdentifier(`user"name`)
	}
}

func BenchmarkColumnExists(b *testing.B) {
	tableInfo := &database.TableInfo{
		Columns: []database.ColumnInfo{
			{Name: "id"},
			{Name: "name"},
			{Name: "email"},
			{Name: "created_at"},
			{Name: "updated_at"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		columnExists(tableInfo, "created_at")
	}
}

func BenchmarkValidateAndQuoteReturning(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validateAndQuoteReturning("id,name,email,created_at", nil)
	}
}

func BenchmarkParseFilterToSQL(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = parseFilterToSQL("email", "eq.user@example.com", 1)
	}
}

func BenchmarkValidIdentifierRegex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validIdentifierRegex.MatchString("user_accounts_table")
	}
}
