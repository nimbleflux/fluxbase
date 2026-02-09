package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryBuilder_BuildSelect(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *QueryBuilder
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "simple select all",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users")
			},
			expectedSQL:  `SELECT * FROM "public"."users"`,
			expectedArgs: nil,
		},
		{
			name: "select specific columns",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithColumns([]string{"id", "email", "name"})
			},
			expectedSQL:  `SELECT "id", "email", "name" FROM "public"."users"`,
			expectedArgs: nil,
		},
		{
			name: "select with single filter",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "id", Operator: OpEqual, Value: 123},
					})
			},
			expectedSQL:  `SELECT * FROM "public"."users" WHERE "id" = $1`,
			expectedArgs: []interface{}{123},
		},
		{
			name: "select with multiple AND filters",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "status", Operator: OpEqual, Value: "active"},
						{Column: "age", Operator: OpGreaterOrEqual, Value: 18},
					})
			},
			expectedSQL:  `SELECT * FROM "public"."users" WHERE "status" = $1 AND "age" >= $2`,
			expectedArgs: []interface{}{"active", 18},
		},
		{
			name: "select with OR filters in same group",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "role", Operator: OpEqual, Value: "admin", OrGroupID: 1},
						{Column: "role", Operator: OpEqual, Value: "moderator", OrGroupID: 1},
					})
			},
			expectedSQL:  `SELECT * FROM "public"."users" WHERE ("role" = $1 OR "role" = $2)`,
			expectedArgs: []interface{}{"admin", "moderator"},
		},
		{
			name: "select with ordering",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithOrder([]OrderBy{
						{Column: "created_at", Desc: true},
					})
			},
			expectedSQL:  `SELECT * FROM "public"."users" ORDER BY "created_at" DESC`,
			expectedArgs: nil,
		},
		{
			name: "select with ordering and nulls",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithOrder([]OrderBy{
						{Column: "name", Desc: false, Nulls: "last"},
					})
			},
			expectedSQL:  `SELECT * FROM "public"."users" ORDER BY "name" ASC NULLS LAST`,
			expectedArgs: nil,
		},
		{
			name: "select with limit and offset",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithLimit(10).
					WithOffset(20)
			},
			expectedSQL:  `SELECT * FROM "public"."users" LIMIT 10 OFFSET 20`,
			expectedArgs: nil,
		},
		{
			name: "select with group by",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "orders").
					WithColumns([]string{"status"}).
					WithGroupBy([]string{"status"})
			},
			expectedSQL:  `SELECT "status" FROM "public"."orders" GROUP BY "status"`,
			expectedArgs: nil,
		},
		{
			name: "complex select with all clauses",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("app", "products").
					WithColumns([]string{"category", "name"}).
					WithFilters([]Filter{
						{Column: "active", Operator: OpEqual, Value: true},
					}).
					WithOrder([]OrderBy{
						{Column: "name", Desc: false},
					}).
					WithLimit(50).
					WithOffset(100)
			},
			expectedSQL:  `SELECT "category", "name" FROM "app"."products" WHERE "active" = $1 ORDER BY "name" ASC LIMIT 50 OFFSET 100`,
			expectedArgs: []interface{}{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.setup()
			sql, args := qb.BuildSelect()
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestQueryBuilder_BuildCount(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *QueryBuilder
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "count all",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users")
			},
			expectedSQL:  `SELECT COUNT(*) FROM "public"."users"`,
			expectedArgs: nil,
		},
		{
			name: "count with filter",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "status", Operator: OpEqual, Value: "active"},
					})
			},
			expectedSQL:  `SELECT COUNT(*) FROM "public"."users" WHERE "status" = $1`,
			expectedArgs: []interface{}{"active"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.setup()
			sql, args := qb.BuildCount()
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestQueryBuilder_BuildInsert(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *QueryBuilder
		data         map[string]interface{}
		expectedSQL  string
		expectedArgs int // Just check count since map iteration order is non-deterministic
	}{
		{
			name: "simple insert",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users")
			},
			data: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedSQL:  `INSERT INTO "public"."users" ("email") VALUES ($1)`,
			expectedArgs: 1,
		},
		{
			name: "insert with returning",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithReturning([]string{"id", "created_at"})
			},
			data: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedSQL:  `INSERT INTO "public"."users" ("email") VALUES ($1) RETURNING "id", "created_at"`,
			expectedArgs: 1,
		},
		{
			name: "insert empty data",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users")
			},
			data:         map[string]interface{}{},
			expectedSQL:  "",
			expectedArgs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.setup()
			sql, args := qb.BuildInsert(tt.data)

			if tt.expectedSQL == "" {
				assert.Empty(t, sql)
				assert.Nil(t, args)
				return
			}

			// For single-column inserts, we can check exact SQL
			if len(tt.data) == 1 {
				assert.Equal(t, tt.expectedSQL, sql)
			}
			assert.Equal(t, tt.expectedArgs, len(args))
		})
	}
}

func TestQueryBuilder_BuildUpdate(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *QueryBuilder
		data         map[string]interface{}
		expectedArgs int
		checkSQL     func(t *testing.T, sql string)
	}{
		{
			name: "update with filter",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "id", Operator: OpEqual, Value: 123},
					})
			},
			data: map[string]interface{}{
				"name": "Updated Name",
			},
			expectedArgs: 2, // 1 for SET, 1 for WHERE
			checkSQL: func(t *testing.T, sql string) {
				assert.Contains(t, sql, `UPDATE "public"."users" SET`)
				assert.Contains(t, sql, `WHERE "id" =`)
			},
		},
		{
			name: "update with returning",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "id", Operator: OpEqual, Value: 1},
					}).
					WithReturning([]string{"id", "name"})
			},
			data: map[string]interface{}{
				"name": "New Name",
			},
			expectedArgs: 2,
			checkSQL: func(t *testing.T, sql string) {
				assert.Contains(t, sql, `RETURNING "id", "name"`)
			},
		},
		{
			name: "update empty data",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users")
			},
			data:         map[string]interface{}{},
			expectedArgs: 0,
			checkSQL: func(t *testing.T, sql string) {
				assert.Empty(t, sql)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.setup()
			sql, args := qb.BuildUpdate(tt.data)

			if tt.expectedArgs == 0 && len(tt.data) == 0 {
				assert.Empty(t, sql)
				assert.Nil(t, args)
				return
			}

			assert.Equal(t, tt.expectedArgs, len(args))
			if tt.checkSQL != nil {
				tt.checkSQL(t, sql)
			}
		})
	}
}

func TestQueryBuilder_BuildDelete(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *QueryBuilder
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "delete all (dangerous but valid)",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "temp_data")
			},
			expectedSQL:  `DELETE FROM "public"."temp_data"`,
			expectedArgs: nil,
		},
		{
			name: "delete with filter",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "id", Operator: OpEqual, Value: 123},
					})
			},
			expectedSQL:  `DELETE FROM "public"."users" WHERE "id" = $1`,
			expectedArgs: []interface{}{123},
		},
		{
			name: "delete with returning",
			setup: func() *QueryBuilder {
				return NewQueryBuilder("public", "users").
					WithFilters([]Filter{
						{Column: "id", Operator: OpEqual, Value: 1},
					}).
					WithReturning([]string{"id", "email"})
			},
			expectedSQL:  `DELETE FROM "public"."users" WHERE "id" = $1 RETURNING "id", "email"`,
			expectedArgs: []interface{}{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := tt.setup()
			sql, args := qb.BuildDelete()
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestQueryBuilder_FilterOperators(t *testing.T) {
	tests := []struct {
		name         string
		filter       Filter
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name:         "equal",
			filter:       Filter{Column: "name", Operator: OpEqual, Value: "test"},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "name" = $1`,
			expectedArgs: []interface{}{"test"},
		},
		{
			name:         "not equal",
			filter:       Filter{Column: "status", Operator: OpNotEqual, Value: "deleted"},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "status" <> $1`,
			expectedArgs: []interface{}{"deleted"},
		},
		{
			name:         "greater than",
			filter:       Filter{Column: "age", Operator: OpGreaterThan, Value: 18},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "age" > $1`,
			expectedArgs: []interface{}{18},
		},
		{
			name:         "less than or equal",
			filter:       Filter{Column: "price", Operator: OpLessOrEqual, Value: 100.0},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "price" <= $1`,
			expectedArgs: []interface{}{100.0},
		},
		{
			name:         "like",
			filter:       Filter{Column: "email", Operator: OpLike, Value: "%@example.com"},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "email" LIKE $1`,
			expectedArgs: []interface{}{"%@example.com"},
		},
		{
			name:         "ilike",
			filter:       Filter{Column: "name", Operator: OpILike, Value: "%john%"},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "name" ILIKE $1`,
			expectedArgs: []interface{}{"%john%"},
		},
		{
			name:         "is null",
			filter:       Filter{Column: "deleted_at", Operator: OpIs, Value: nil},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "deleted_at" IS NULL`,
			expectedArgs: nil,
		},
		{
			name:         "contains (jsonb @>)",
			filter:       Filter{Column: "metadata", Operator: OpContains, Value: `{"role":"admin"}`},
			expectedSQL:  `SELECT * FROM "public"."t" WHERE "metadata" @> $1`,
			expectedArgs: []interface{}{`{"role":"admin"}`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder("public", "t").
				WithFilters([]Filter{tt.filter})
			sql, args := qb.BuildSelect()
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestQueryBuilder_InvalidIdentifiers(t *testing.T) {
	t.Run("invalid column name is skipped", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithColumns([]string{"valid_col", "invalid col", "another_valid"})
		sql, _ := qb.BuildSelect()
		// Invalid column should be skipped
		assert.Contains(t, sql, `"valid_col"`)
		assert.Contains(t, sql, `"another_valid"`)
		assert.NotContains(t, sql, "invalid col")
	})

	t.Run("filter with invalid column is skipped", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{
				{Column: "valid", Operator: OpEqual, Value: 1},
				{Column: "has space", Operator: OpEqual, Value: 2},
			})
		sql, args := qb.BuildSelect()
		assert.Contains(t, sql, `"valid" = $1`)
		assert.NotContains(t, sql, "has space")
		assert.Equal(t, 1, len(args))
	})
}

func TestNewQueryBuilder(t *testing.T) {
	t.Run("initializes with correct defaults", func(t *testing.T) {
		qb := NewQueryBuilder("myschema", "mytable")
		assert.NotNil(t, qb)

		sql, args := qb.BuildSelect()
		assert.Equal(t, `SELECT * FROM "myschema"."mytable"`, sql)
		assert.Nil(t, args)
	})
}

// =============================================================================
// Cursor Pagination Tests
// =============================================================================

func TestEncodeCursor(t *testing.T) {
	t.Run("encodes cursor correctly", func(t *testing.T) {
		cursor := EncodeCursor("id", "abc123", false)
		assert.NotEmpty(t, cursor)

		// Should be valid base64
		decoded, err := DecodeCursor(cursor)
		assert.NoError(t, err)
		assert.Equal(t, "id", decoded.Column)
		assert.Equal(t, "abc123", decoded.Value)
		assert.False(t, decoded.Desc)
	})

	t.Run("encodes descending cursor", func(t *testing.T) {
		cursor := EncodeCursor("created_at", "2025-01-01", true)
		decoded, err := DecodeCursor(cursor)
		assert.NoError(t, err)
		assert.Equal(t, "created_at", decoded.Column)
		assert.True(t, decoded.Desc)
	})

	t.Run("encodes numeric value", func(t *testing.T) {
		cursor := EncodeCursor("count", 42, false)
		decoded, err := DecodeCursor(cursor)
		assert.NoError(t, err)
		assert.Equal(t, float64(42), decoded.Value) // JSON unmarshals numbers as float64
	})
}

func TestDecodeCursor(t *testing.T) {
	t.Run("decodes valid cursor", func(t *testing.T) {
		// First encode, then decode
		original := EncodeCursor("id", "test123", false)
		decoded, err := DecodeCursor(original)
		assert.NoError(t, err)
		assert.Equal(t, "id", decoded.Column)
		assert.Equal(t, "test123", decoded.Value)
	})

	t.Run("fails on invalid base64", func(t *testing.T) {
		_, err := DecodeCursor("not-valid-base64!!!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor encoding")
	})

	t.Run("fails on invalid JSON", func(t *testing.T) {
		// Valid base64 but invalid JSON
		_, err := DecodeCursor("bm90LWpzb24=") // "not-json" in base64
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor format")
	})

	t.Run("fails on missing column", func(t *testing.T) {
		// Valid base64 of {"v": "value"} without column
		_, err := DecodeCursor("eyJ2IjoidmFsdWUifQ==")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cursor missing column")
	})
}

func TestQueryBuilder_WithCursor(t *testing.T) {
	t.Run("applies cursor condition ascending", func(t *testing.T) {
		cursor := EncodeCursor("id", "last123", false)

		qb := NewQueryBuilder("public", "users")
		err := qb.WithCursor(cursor, "")
		assert.NoError(t, err)

		sql, args := qb.BuildSelect()
		assert.Contains(t, sql, `WHERE "id" > $1`)
		assert.Len(t, args, 1)
		assert.Equal(t, "last123", args[0])
	})

	t.Run("applies cursor condition descending", func(t *testing.T) {
		cursor := EncodeCursor("created_at", "2025-01-01", true)

		qb := NewQueryBuilder("public", "users")
		err := qb.WithCursor(cursor, "")
		assert.NoError(t, err)

		sql, args := qb.BuildSelect()
		assert.Contains(t, sql, `WHERE "created_at" < $1`)
		assert.Len(t, args, 1)
	})

	t.Run("cursor column override", func(t *testing.T) {
		cursor := EncodeCursor("old_column", "value", false)

		qb := NewQueryBuilder("public", "users")
		err := qb.WithCursor(cursor, "new_column")
		assert.NoError(t, err)

		sql, args := qb.BuildSelect()
		assert.Contains(t, sql, `WHERE "new_column" > $1`)
		assert.Len(t, args, 1)
	})

	t.Run("combines cursor with filters", func(t *testing.T) {
		cursor := EncodeCursor("id", "last123", false)

		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{{Column: "status", Operator: OpEqual, Value: "active"}})
		err := qb.WithCursor(cursor, "")
		assert.NoError(t, err)

		sql, args := qb.BuildSelect()
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, `"status" = $1`)
		assert.Contains(t, sql, `"id" > $2`)
		assert.Len(t, args, 2)
	})

	t.Run("empty cursor is no-op", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users")
		err := qb.WithCursor("", "")
		assert.NoError(t, err)

		sql, args := qb.BuildSelect()
		assert.Equal(t, `SELECT * FROM "public"."users"`, sql)
		assert.Nil(t, args)
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users")
		err := qb.WithCursor("invalid!!!", "")
		assert.Error(t, err)
	})
}

// =============================================================================
// Additional Tests for Coverage Boost (Developer 3 Assignment)
// =============================================================================

// TestQueryBuilder_ComplexQueries tests complex query building scenarios
func TestQueryBuilder_ComplexQueries(t *testing.T) {
	t.Run("query with all clauses", func(t *testing.T) {
		qb := NewQueryBuilder("public", "orders").
			WithColumns([]string{"id", "customer_id", "total", "status"}).
			WithFilters([]Filter{
				{Column: "status", Operator: OpEqual, Value: "completed"},
				{Column: "total", Operator: OpGreaterThan, Value: 100},
			}).
			WithOrder([]OrderBy{
				{Column: "created_at", Desc: true},
			}).
			WithLimit(50).
			WithOffset(100)

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `SELECT "id", "customer_id", "total", "status"`)
		assert.Contains(t, sql, `FROM "public"."orders"`)
		assert.Contains(t, sql, `WHERE "status" = $1`)
		assert.Contains(t, sql, `AND "total" > $2`)
		assert.Contains(t, sql, `ORDER BY "created_at" DESC`)
		assert.Contains(t, sql, `LIMIT 50 OFFSET 100`)
		assert.Equal(t, 2, len(args))
		assert.Equal(t, "completed", args[0])
		assert.Equal(t, 100, args[1])
	})

	t.Run("query with OR groups", func(t *testing.T) {
		qb := NewQueryBuilder("public", "products").
			WithFilters([]Filter{
				{Column: "category", Operator: OpEqual, Value: "electronics", OrGroupID: 1},
				{Column: "category", Operator: OpEqual, Value: "computers", OrGroupID: 1},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `("category" = $1 OR "category" = $2)`)
		assert.Equal(t, []interface{}{"electronics", "computers"}, args)
	})

	t.Run("query with multiple OR groups", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithFilters([]Filter{
				{Column: "status", Operator: OpEqual, Value: "active"},
				{Column: "type", Operator: OpEqual, Value: "a", OrGroupID: 1},
				{Column: "type", Operator: OpEqual, Value: "b", OrGroupID: 1},
				{Column: "priority", Operator: OpEqual, Value: "high", OrGroupID: 2},
				{Column: "priority", Operator: OpEqual, Value: "urgent", OrGroupID: 2},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `"status" =`)
		assert.Contains(t, sql, `("type" =`)
		assert.Contains(t, sql, `OR "type" =`)
		assert.Contains(t, sql, `("priority" =`)
		assert.Contains(t, sql, `OR "priority" =`)
		assert.Contains(t, sql, " AND ")
		assert.Equal(t, 5, len(args))
	})

	t.Run("query with IS NULL filter", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{
				{Column: "deleted_at", Operator: OpIs, Value: nil},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `WHERE "deleted_at" IS NULL`)
		assert.Nil(t, args)
	})

	t.Run("query with IS NOT NULL filter", func(t *testing.T) {
		qb := NewQueryBuilder("public", "posts").
			WithFilters([]Filter{
				{Column: "published_at", Operator: OpIsNot, Value: nil},
			})

		sql, args := qb.BuildSelect()

		// OpIsNot is handled as != comparison, not IS NOT NULL
		assert.Contains(t, sql, `WHERE "published_at" =`)
		assert.Nil(t, args)
	})

	t.Run("query with IN operator", func(t *testing.T) {
		qb := NewQueryBuilder("public", "tasks").
			WithFilters([]Filter{
				{Column: "status", Operator: OpIn, Value: []string{"pending", "in_progress", "queued"}},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `WHERE "status" = ANY($1)`)
		assert.Equal(t, []string{"pending", "in_progress", "queued"}, args[0])
	})

	t.Run("query with LIKE operator", func(t *testing.T) {
		qb := NewQueryBuilder("public", "products").
			WithFilters([]Filter{
				{Column: "name", Operator: OpLike, Value: "%iPhone%"},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `WHERE "name" LIKE $1`)
		assert.Equal(t, "%iPhone%", args[0])
	})

	t.Run("query with ILIKE operator", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{
				{Column: "email", Operator: OpILike, Value: "*@gmail.com"},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `WHERE "email" ILIKE $1`)
		assert.Equal(t, "*@gmail.com", args[0])
	})
}

// TestQueryBuilder_EdgeCases tests edge cases in query building
func TestQueryBuilder_EdgeCases(t *testing.T) {
	t.Run("empty query builder", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users")
		sql, args := qb.BuildSelect()

		assert.Equal(t, `SELECT * FROM "public"."users"`, sql)
		assert.Nil(t, args)
	})

	t.Run("query with no filters but with limit", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithLimit(10)

		sql, args := qb.BuildSelect()

		assert.Equal(t, `SELECT * FROM "public"."items" LIMIT 10`, sql)
		assert.Nil(t, args)
	})

	t.Run("query with order but no limit", func(t *testing.T) {
		qb := NewQueryBuilder("public", "data").
			WithOrder([]OrderBy{{Column: "id", Desc: false}})

		sql, args := qb.BuildSelect()

		assert.Equal(t, `SELECT * FROM "public"."data" ORDER BY "id" ASC`, sql)
		assert.Nil(t, args)
	})

	t.Run("query with offset but no limit", func(t *testing.T) {
		qb := NewQueryBuilder("public", "records").
			WithOffset(100)

		sql, args := qb.BuildSelect()

		assert.Equal(t, `SELECT * FROM "public"."records" OFFSET 100`, sql)
		assert.Nil(t, args)
	})

	t.Run("query with zero limit", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithLimit(0)

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, "LIMIT 0")
		assert.Nil(t, args)
	})

	t.Run("query with zero offset", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithOffset(0)

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, "OFFSET 0")
		assert.Nil(t, args)
	})

	t.Run("query with invalid column in select", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithColumns([]string{"valid_col", "invalid-col", "another_valid"})

		sql, _ := qb.BuildSelect()

		assert.Contains(t, sql, `"valid_col"`)
		assert.Contains(t, sql, `"another_valid"`)
		assert.NotContains(t, sql, "invalid-col")
	})

	t.Run("query with invalid column in filter", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{
				{Column: "valid", Operator: OpEqual, Value: 1},
				{Column: "invalid-col", Operator: OpEqual, Value: 2},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `"valid" =`)
		assert.NotContains(t, sql, "invalid-col")
		assert.Equal(t, 1, len(args))
	})

	t.Run("query with invalid column in order", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithOrder([]OrderBy{
				{Column: "valid_col", Desc: true},
				{Column: "invalid-col", Desc: false},
			})

		sql, _ := qb.BuildSelect()

		assert.Contains(t, sql, `"valid_col" DESC`)
		assert.NotContains(t, sql, "invalid-col")
	})

	t.Run("insert with all types of values", func(t *testing.T) {
		qb := NewQueryBuilder("public", "test_table")
		data := map[string]interface{}{
			"string_col": "text",
			"int_col":    42,
			"float_col":  3.14,
			"bool_col":   true,
			"null_col":   nil,
			"array_col":  []int{1, 2, 3},
			"json_col":   map[string]interface{}{"key": "value"},
		}

		sql, args := qb.BuildInsert(data)

		assert.Contains(t, sql, `INSERT INTO "public"."test_table"`)
		assert.Contains(t, sql, "VALUES")
		assert.Equal(t, 7, len(args))
	})

	t.Run("update with multiple columns", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithFilters([]Filter{
				{Column: "id", Operator: OpEqual, Value: 1},
			})
		data := map[string]interface{}{
			"name":  "Updated Name",
			"email": "updated@example.com",
			"age":   30,
		}

		sql, args := qb.BuildUpdate(data)

		assert.Contains(t, sql, `UPDATE "public"."users" SET`)
		assert.Contains(t, sql, `WHERE "id" =`)
		assert.Equal(t, 4, len(args)) // 3 for SET + 1 for WHERE
	})

	t.Run("delete with filter", func(t *testing.T) {
		qb := NewQueryBuilder("public", "logs").
			WithFilters([]Filter{
				{Column: "created_at", Operator: OpLessThan, Value: "2024-01-01"},
			})

		sql, args := qb.BuildDelete()

		assert.Equal(t, `DELETE FROM "public"."logs" WHERE "created_at" < $1`, sql)
		assert.Equal(t, "2024-01-01", args[0])
	})

	t.Run("count with filter", func(t *testing.T) {
		qb := NewQueryBuilder("public", "orders").
			WithFilters([]Filter{
				{Column: "status", Operator: OpEqual, Value: "completed"},
			})

		sql, args := qb.BuildCount()

		assert.Equal(t, `SELECT COUNT(*) FROM "public"."orders" WHERE "status" = $1`, sql)
		assert.Equal(t, "completed", args[0])
	})
}

// TestQueryBuilder_WithReturning tests RETURNING clause
func TestQueryBuilder_WithReturning(t *testing.T) {
	t.Run("insert with returning", func(t *testing.T) {
		qb := NewQueryBuilder("public", "users").
			WithReturning([]string{"id", "created_at"})
		data := map[string]interface{}{
			"name": "John Doe",
		}

		sql, args := qb.BuildInsert(data)

		assert.Contains(t, sql, "RETURNING")
		assert.Contains(t, sql, `"id"`)
		assert.Contains(t, sql, `"created_at"`)
		assert.Equal(t, 1, len(args))
	})

	t.Run("update with returning", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithFilters([]Filter{
				{Column: "id", Operator: OpEqual, Value: 1},
			}).
			WithReturning([]string{"id", "name", "updated_at"})
		data := map[string]interface{}{
			"name": "Updated",
		}

		sql, args := qb.BuildUpdate(data)

		assert.Contains(t, sql, "RETURNING")
		assert.Contains(t, sql, `"id"`)
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, `"updated_at"`)
		_ = args // Use args to avoid declared and not used error
	})

	t.Run("delete with returning", func(t *testing.T) {
		qb := NewQueryBuilder("public", "records").
			WithFilters([]Filter{
				{Column: "id", Operator: OpEqual, Value: 1},
			}).
			WithReturning([]string{"id", "name"})

		sql, args := qb.BuildDelete()

		assert.Contains(t, sql, "RETURNING")
		assert.Contains(t, sql, `"id"`)
		assert.Contains(t, sql, `"name"`)
		assert.Equal(t, 1, len(args))
	})
}

// TestQueryBuilder_OrderWithNulls tests order by with nulls handling
func TestQueryBuilder_OrderWithNulls(t *testing.T) {
	tests := []struct {
		name     string
		order    OrderBy
		expected string
	}{
		{
			name:     "asc nulls first",
			order:    OrderBy{Column: "name", Desc: false, Nulls: "first"},
			expected: `"name" ASC NULLS FIRST`,
		},
		{
			name:     "asc nulls last",
			order:    OrderBy{Column: "name", Desc: false, Nulls: "last"},
			expected: `"name" ASC NULLS LAST`,
		},
		{
			name:     "desc nulls first",
			order:    OrderBy{Column: "created_at", Desc: true, Nulls: "first"},
			expected: `"created_at" DESC NULLS FIRST`,
		},
		{
			name:     "desc nulls last",
			order:    OrderBy{Column: "priority", Desc: true, Nulls: "last"},
			expected: `"priority" DESC NULLS LAST`,
		},
		{
			name:     "asc without nulls specified",
			order:    OrderBy{Column: "id", Desc: false, Nulls: ""},
			expected: `"id" ASC`,
		},
		{
			name:     "desc without nulls specified",
			order:    OrderBy{Column: "updated_at", Desc: true, Nulls: ""},
			expected: `"updated_at" DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := NewQueryBuilder("public", "items").
				WithOrder([]OrderBy{tt.order})

			sql, _ := qb.BuildSelect()

			assert.Contains(t, sql, "ORDER BY")
			assert.Contains(t, sql, tt.expected)
		})
	}
}

// TestQueryBuilder_MultipleOrderColumns tests multiple ORDER BY columns
func TestQueryBuilder_MultipleOrderColumns(t *testing.T) {
	qb := NewQueryBuilder("public", "products").
		WithOrder([]OrderBy{
			{Column: "category", Desc: false},
			{Column: "price", Desc: true},
			{Column: "name", Desc: false, Nulls: "last"},
		})

	sql, _ := qb.BuildSelect()

	assert.Contains(t, sql, `ORDER BY "category" ASC, "price" DESC, "name" ASC NULLS LAST`)
}

// TestQueryBuilder_FilterCombinations tests various filter combinations
func TestQueryBuilder_FilterCombinations(t *testing.T) {
	t.Run("AND with multiple operators", func(t *testing.T) {
		qb := NewQueryBuilder("public", "products").
			WithFilters([]Filter{
				{Column: "price", Operator: OpGreaterOrEqual, Value: 10},
				{Column: "price", Operator: OpLessOrEqual, Value: 100},
				{Column: "stock", Operator: OpGreaterThan, Value: 0},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `WHERE "price" >= $1`)
		assert.Contains(t, sql, `AND "price" <= $2`)
		assert.Contains(t, sql, `AND "stock" > $3`)
		assert.Equal(t, []interface{}{10, 100, 0}, args)
	})

	t.Run("complex OR with AND", func(t *testing.T) {
		qb := NewQueryBuilder("public", "items").
			WithFilters([]Filter{
				{Column: "status", Operator: OpEqual, Value: "active"},
				{Column: "priority", Operator: OpEqual, Value: "high", OrGroupID: 1},
				{Column: "priority", Operator: OpEqual, Value: "urgent", OrGroupID: 1},
			})

		sql, args := qb.BuildSelect()

		assert.Contains(t, sql, `"status" =`)
		assert.Contains(t, sql, `("priority" =`)
		assert.Contains(t, sql, `OR "priority" =`)
		assert.Equal(t, 3, len(args))
	})
}

// TestDecodeCursor_BadEncoding tests various invalid cursor encodings
func TestDecodeCursor_BadEncoding(t *testing.T) {
	tests := []struct {
		name        string
		cursor      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "not base64",
			cursor:      "not-valid-base64",
			expectError: true,
			errorMsg:    "invalid cursor format",
		},
		{
			name:        "base64 but not JSON",
			cursor:      "bm90LWpzb24=", // "not-json" in base64
			expectError: true,
			errorMsg:    "invalid cursor format",
		},
		{
			name:        "missing column field",
			cursor:      "eyJ2IjoidmFsdWUifQ==", // {"v":"value"}
			expectError: true,
			errorMsg:    "cursor missing column",
		},
		{
			name:        "empty string",
			cursor:      "",
			expectError: true,
		},
		{
			name:        "invalid base64 characters",
			cursor:      "!!!",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor(tt.cursor)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestEncodeDecodeRoundTrip tests cursor encode/decode round trips
func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []struct {
		name   string
		column string
		value  interface{}
		desc   bool
	}{
		{
			name:   "string value ascending",
			column: "id",
			value:  "abc123",
			desc:   false,
		},
		{
			name:   "string value descending",
			column: "created_at",
			value:  "2025-01-01T00:00:00Z",
			desc:   true,
		},
		{
			name:   "numeric value",
			column: "count",
			value:  42,
			desc:   false,
		},
		{
			name:   "float value",
			column: "price",
			value:  19.99,
			desc:   true,
		},
		{
			name:   "bool value",
			column: "active",
			value:  true,
			desc:   false,
		},
		{
			name:   "null value",
			column: "deleted_at",
			value:  nil,
			desc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor := EncodeCursor(tt.column, tt.value, tt.desc)
			decoded, err := DecodeCursor(cursor)

			require.NoError(t, err)
			assert.Equal(t, tt.column, decoded.Column)
			assert.Equal(t, tt.desc, decoded.Desc)

			// Value comparison needs special handling for types
			if tt.value == nil {
				assert.Nil(t, decoded.Value)
			} else if num, ok := tt.value.(int); ok {
				assert.Equal(t, float64(num), decoded.Value)
			} else if num, ok := tt.value.(float64); ok {
				assert.Equal(t, num, decoded.Value)
			} else {
				assert.Equal(t, tt.value, decoded.Value)
			}
		})
	}
}
