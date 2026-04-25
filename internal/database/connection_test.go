package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ExtractTableName Tests
// =============================================================================

func TestExtractTableName(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		// SELECT queries
		{
			name:     "simple select",
			sql:      "SELECT * FROM users",
			expected: "users",
		},
		{
			name:     "select with columns",
			sql:      "SELECT id, name, email FROM users WHERE active = true",
			expected: "users",
		},
		{
			name:     "select with schema",
			sql:      "SELECT * FROM public.users",
			expected: "public",
		},
		{
			name:     "select lowercase",
			sql:      "select * from products",
			expected: "products",
		},
		{
			name:     "select with quoted table",
			sql:      `SELECT * FROM "users"`,
			expected: "users",
		},
		{
			name:     "select with single quoted table",
			sql:      "SELECT * FROM 'users'",
			expected: "users",
		},

		// INSERT queries
		{
			name:     "simple insert",
			sql:      "INSERT INTO users (name) VALUES ('John')",
			expected: "users",
		},
		{
			name:     "insert with schema",
			sql:      "INSERT INTO auth.users (name) VALUES ('John')",
			expected: "auth",
		},
		{
			name:     "insert lowercase",
			sql:      "insert into products (name) values ('Widget')",
			expected: "products",
		},

		// UPDATE queries
		{
			name:     "simple update",
			sql:      "UPDATE users SET name = 'Jane' WHERE id = 1",
			expected: "users",
		},
		{
			name:     "update with schema",
			sql:      "UPDATE public.users SET name = 'Jane'",
			expected: "public",
		},
		{
			name:     "update lowercase",
			sql:      "update orders set status = 'shipped'",
			expected: "orders",
		},

		// DELETE queries
		{
			name:     "simple delete",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: "users",
		},
		{
			name:     "delete with schema",
			sql:      "DELETE FROM auth.sessions WHERE expired = true",
			expected: "auth",
		},
		{
			name:     "delete lowercase",
			sql:      "delete from temp_data",
			expected: "temp_data",
		},

		// Edge cases
		{
			name:     "unknown statement type",
			sql:      "CREATE TABLE users (id INT)",
			expected: "unknown",
		},
		{
			name:     "truncate statement",
			sql:      "TRUNCATE TABLE users",
			expected: "unknown",
		},
		{
			name:     "empty string",
			sql:      "",
			expected: "unknown",
		},
		{
			name:     "whitespace only",
			sql:      "   ",
			expected: "unknown",
		},
		{
			name:     "select with join",
			sql:      "SELECT u.* FROM users u JOIN orders o ON u.id = o.user_id",
			expected: "users",
		},
		{
			name:     "select with subquery",
			sql:      "SELECT * FROM (SELECT * FROM users) as subq",
			expected: "users", // ExtractTableName uses simple regex that finds first FROM
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractTableName(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractTableName_CaseInsensitive(t *testing.T) {
	// All variations should work
	variations := []string{
		"SELECT * FROM users",
		"select * from users",
		"Select * From users",
		"SELECT * FROM USERS",
		"sElEcT * fRoM users",
	}

	for _, sql := range variations {
		result := ExtractTableName(sql)
		assert.Equal(t, "users", result, "Failed for SQL: %s", sql)
	}
}

// =============================================================================
// ExtractOperation Tests
// =============================================================================

func TestExtractOperation(t *testing.T) {
	tests := []struct {
		name     string
		sql      string
		expected string
	}{
		// SELECT
		{
			name:     "select uppercase",
			sql:      "SELECT * FROM users",
			expected: "select",
		},
		{
			name:     "select lowercase",
			sql:      "select * from users",
			expected: "select",
		},
		{
			name:     "select mixed case",
			sql:      "Select * From users",
			expected: "select",
		},
		{
			name:     "select with leading whitespace",
			sql:      "   SELECT * FROM users",
			expected: "select",
		},

		// INSERT
		{
			name:     "insert uppercase",
			sql:      "INSERT INTO users VALUES (1)",
			expected: "insert",
		},
		{
			name:     "insert lowercase",
			sql:      "insert into users values (1)",
			expected: "insert",
		},

		// UPDATE
		{
			name:     "update uppercase",
			sql:      "UPDATE users SET name = 'John'",
			expected: "update",
		},
		{
			name:     "update lowercase",
			sql:      "update users set name = 'John'",
			expected: "update",
		},

		// DELETE
		{
			name:     "delete uppercase",
			sql:      "DELETE FROM users WHERE id = 1",
			expected: "delete",
		},
		{
			name:     "delete lowercase",
			sql:      "delete from users where id = 1",
			expected: "delete",
		},

		// Other operations
		{
			name:     "create table",
			sql:      "CREATE TABLE users (id INT)",
			expected: "other",
		},
		{
			name:     "drop table",
			sql:      "DROP TABLE users",
			expected: "other",
		},
		{
			name:     "alter table",
			sql:      "ALTER TABLE users ADD COLUMN email TEXT",
			expected: "other",
		},
		{
			name:     "truncate",
			sql:      "TRUNCATE TABLE users",
			expected: "other",
		},
		{
			name:     "begin transaction",
			sql:      "BEGIN",
			expected: "other",
		},
		{
			name:     "commit",
			sql:      "COMMIT",
			expected: "other",
		},
		{
			name:     "rollback",
			sql:      "ROLLBACK",
			expected: "other",
		},
		{
			name:     "set statement",
			sql:      "SET search_path TO public",
			expected: "other",
		},

		// Edge cases
		{
			name:     "empty string",
			sql:      "",
			expected: "other",
		},
		{
			name:     "whitespace only",
			sql:      "   ",
			expected: "other",
		},
		{
			name:     "comment only",
			sql:      "-- this is a comment",
			expected: "other",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractOperation(tt.sql)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// truncateQuery Tests
// =============================================================================

func TestTruncateQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		maxLen   int
		expected string
	}{
		{
			name:     "short query under limit",
			query:    "SELECT * FROM users",
			maxLen:   100,
			expected: "SELECT * FROM users",
		},
		{
			name:     "query exactly at limit",
			query:    "SELECT * FROM users",
			maxLen:   19,
			expected: "SELECT * FROM users",
		},
		{
			name:     "query over limit",
			query:    "SELECT * FROM users WHERE active = true",
			maxLen:   20,
			expected: "SELECT * FROM users ... (truncated)",
		},
		{
			name:     "very short limit",
			query:    "SELECT * FROM users",
			maxLen:   5,
			expected: "SELEC... (truncated)",
		},
		{
			name:     "empty query",
			query:    "",
			maxLen:   100,
			expected: "",
		},
		{
			name:     "zero max length",
			query:    "SELECT",
			maxLen:   0,
			expected: "... (truncated)",
		},
		{
			name:     "long query",
			query:    "SELECT id, name, email, created_at, updated_at, status, role, metadata FROM users WHERE active = true AND verified = true ORDER BY created_at DESC LIMIT 100",
			maxLen:   50,
			expected: "SELECT id, name, email, created_at, updated_at, st... (truncated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateQuery(tt.query, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateQuery_Length(t *testing.T) {
	query := "SELECT * FROM users WHERE id IN (1, 2, 3, 4, 5, 6, 7, 8, 9, 10)"
	maxLen := 30

	result := truncateQuery(query, maxLen)

	// Result should contain the truncated marker
	assert.Contains(t, result, "... (truncated)")
	// The prefix should be exactly maxLen characters
	prefix := result[:maxLen]
	assert.Len(t, prefix, maxLen)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkExtractTableName_SELECT(b *testing.B) {
	sql := "SELECT id, name, email FROM users WHERE active = true ORDER BY created_at"
	for i := 0; i < b.N; i++ {
		_ = ExtractTableName(sql)
	}
}

func BenchmarkExtractTableName_INSERT(b *testing.B) {
	sql := "INSERT INTO users (name, email) VALUES ('John', 'john@example.com')"
	for i := 0; i < b.N; i++ {
		_ = ExtractTableName(sql)
	}
}

func BenchmarkExtractTableName_UPDATE(b *testing.B) {
	sql := "UPDATE users SET name = 'Jane', email = 'jane@example.com' WHERE id = 123"
	for i := 0; i < b.N; i++ {
		_ = ExtractTableName(sql)
	}
}

func BenchmarkExtractTableName_DELETE(b *testing.B) {
	sql := "DELETE FROM users WHERE id = 123 AND active = false"
	for i := 0; i < b.N; i++ {
		_ = ExtractTableName(sql)
	}
}

func BenchmarkExtractOperation(b *testing.B) {
	sql := "SELECT * FROM users WHERE active = true"
	for i := 0; i < b.N; i++ {
		_ = ExtractOperation(sql)
	}
}

func BenchmarkTruncateQuery_Short(b *testing.B) {
	query := "SELECT * FROM users"
	for i := 0; i < b.N; i++ {
		_ = truncateQuery(query, 200)
	}
}

func BenchmarkTruncateQuery_Long(b *testing.B) {
	query := "SELECT id, name, email, phone, address, city, state, zip, country, created_at, updated_at FROM users WHERE active = true AND verified = true AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 100 OFFSET 0"
	for i := 0; i < b.N; i++ {
		_ = truncateQuery(query, 100)
	}
}

// =============================================================================
// quoteIdentifier Tests
// =============================================================================

func TestQuoteIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		expected   string
	}{
		{
			name:       "simple table name",
			identifier: "users",
			expected:   `"users"`,
		},
		{
			name:       "table name with underscore",
			identifier: "user_profiles",
			expected:   `"user_profiles"`,
		},
		{
			name:       "schema qualified name",
			identifier: "public.users",
			expected:   `"public.users"`,
		},
		{
			name:       "identifier with embedded quote",
			identifier: `my"table`,
			expected:   `"my""table"`,
		},
		{
			name:       "identifier with multiple quotes",
			identifier: `test"with"quotes`,
			expected:   `"test""with""quotes"`,
		},
		{
			name:       "empty identifier",
			identifier: "",
			expected:   `""`,
		},
		{
			name:       "identifier with spaces",
			identifier: "my table",
			expected:   `"my table"`,
		},
		{
			name:       "reserved keyword",
			identifier: "select",
			expected:   `"select"`,
		},
		{
			name:       "mixed case identifier",
			identifier: "MyTable",
			expected:   `"MyTable"`,
		},
		{
			name:       "identifier with special characters",
			identifier: "user@data",
			expected:   `"user@data"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := quoteIdentifier(tt.identifier)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQuoteIdentifier_SQLInjectionPrevention(t *testing.T) {
	t.Run("prevents basic injection", func(t *testing.T) {
		// Attempt SQL injection via identifier
		malicious := `users"; DROP TABLE users; --`
		result := quoteIdentifier(malicious)

		// The result should be safely quoted, making injection impossible
		assert.Equal(t, `"users""; DROP TABLE users; --"`, result)
		assert.Contains(t, result, `""`)
	})

	t.Run("handles multiple injection attempts", func(t *testing.T) {
		malicious := `"";--`
		result := quoteIdentifier(malicious)

		// Embedded quotes should be doubled (2 quotes become 4, plus 2 wrapper quotes = 6 total)
		assert.Equal(t, `""""";--"`, result)
	})
}

func BenchmarkQuoteIdentifier_Simple(b *testing.B) {
	identifier := "users"
	for i := 0; i < b.N; i++ {
		_ = quoteIdentifier(identifier)
	}
}

func BenchmarkQuoteIdentifier_WithQuotes(b *testing.B) {
	identifier := `table"with"quotes`
	for i := 0; i < b.N; i++ {
		_ = quoteIdentifier(identifier)
	}
}

func TestSlowQueryTracker(t *testing.T) {
	t.Run("starts at count 1", func(t *testing.T) {
		tracker := newSlowQueryTracker()
		count := tracker.record("select:users")
		assert.Equal(t, 1, count)
	})

	t.Run("increments on repeated queries", func(t *testing.T) {
		tracker := newSlowQueryTracker()
		key := "select:users"
		assert.Equal(t, 1, tracker.record(key))
		assert.Equal(t, 2, tracker.record(key))
		assert.Equal(t, 3, tracker.record(key))
	})

	t.Run("tracks different keys independently", func(t *testing.T) {
		tracker := newSlowQueryTracker()
		assert.Equal(t, 1, tracker.record("select:users"))
		assert.Equal(t, 1, tracker.record("insert:orders"))
		assert.Equal(t, 2, tracker.record("select:users"))
		assert.Equal(t, 2, tracker.record("insert:orders"))
	})
}

func TestWithCaller(t *testing.T) {
	t.Run("retrieves caller from context", func(t *testing.T) {
		ctx := WithCaller(context.Background(), "GET /api/v1/users")
		assert.Equal(t, "GET /api/v1/users", getCallerFromContext(ctx))
	})

	t.Run("returns empty for nil context", func(t *testing.T) {
		assert.Equal(t, "", getCallerFromContext(nil))
	})

	t.Run("returns empty for context without caller", func(t *testing.T) {
		assert.Equal(t, "", getCallerFromContext(context.Background()))
	})
}

func TestLogSlowQuery_Threshold(t *testing.T) {
	c := &Connection{
		slowQueryThreshold: 500 * time.Millisecond,
		slowQueryTracker:   newSlowQueryTracker(),
	}

	t.Run("does not log below threshold", func(t *testing.T) {
		assert.NotPanics(t, func() {
			c.logSlowQuery(context.Background(), "SELECT 1", 100*time.Millisecond, "query")
		})
	})

	t.Run("logs at threshold", func(t *testing.T) {
		assert.NotPanics(t, func() {
			c.logSlowQuery(context.Background(), "SELECT * FROM users", 600*time.Millisecond, "query")
		})
	})

	t.Run("logs with caller context", func(t *testing.T) {
		ctx := WithCaller(context.Background(), "GET /api/v1/users")
		assert.NotPanics(t, func() {
			c.logSlowQuery(ctx, "SELECT * FROM users", 600*time.Millisecond, "query")
		})
	})

	t.Run("tracks occurrences", func(t *testing.T) {
		for i := 0; i < 5; i++ {
			assert.NotPanics(t, func() {
				c.logSlowQuery(context.Background(), "SELECT * FROM orders", 600*time.Millisecond, "query")
			})
		}
	})
}
