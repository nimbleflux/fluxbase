package tools

import (
	"context"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/nimbleflux/fluxbase/internal/query"
	"github.com/stretchr/testify/assert"
)

func TestNewQueryTableTool(t *testing.T) {
	t.Run("creates tool with nil dependencies", func(t *testing.T) {
		tool := NewQueryTableTool(nil, nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.db)
		assert.Nil(t, tool.schemaCache)
		assert.Nil(t, tool.embeddingGenerator)
	})
}

func TestQueryTableTool_Metadata(t *testing.T) {
	tool := NewQueryTableTool(nil, nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "query_table", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Query a database table")
		assert.Contains(t, desc, "RLS")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "table")
		assert.Contains(t, props, "select")
		assert.Contains(t, props, "filter")
		assert.Contains(t, props, "order")
		assert.Contains(t, props, "limit")
		assert.Contains(t, props, "offset")
		assert.Contains(t, props, "vector_search")

		required := schema["required"].([]string)
		assert.Contains(t, required, "table")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeReadTables)
	})
}

func TestQueryTableTool_SetEmbeddingGenerator(t *testing.T) {
	tool := NewQueryTableTool(nil, nil)
	assert.Nil(t, tool.embeddingGenerator)

	// Mock embedding generator
	mockGen := &mockEmbeddingGenerator{}
	tool.SetEmbeddingGenerator(mockGen)
	assert.NotNil(t, tool.embeddingGenerator)
}

type mockEmbeddingGenerator struct{}

func (m *mockEmbeddingGenerator) GenerateEmbedding(_ context.Context, _ string) ([]float32, error) {
	return []float32{0.1, 0.2, 0.3}, nil
}

func TestParseOrder(t *testing.T) {
	t.Run("dot-separated desc", func(t *testing.T) {
		result, err := parseOrder("created_at.desc")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "created_at", result[0].Column)
		assert.True(t, result[0].Desc)
	})

	t.Run("dot-separated asc", func(t *testing.T) {
		result, err := parseOrder("name.asc")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "name", result[0].Column)
		assert.False(t, result[0].Desc)
	})

	t.Run("space-separated desc lowercase", func(t *testing.T) {
		result, err := parseOrder("visit_count desc")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "visit_count", result[0].Column)
		assert.True(t, result[0].Desc)
	})

	t.Run("space-separated DESC uppercase", func(t *testing.T) {
		result, err := parseOrder("visit_count DESC")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "visit_count", result[0].Column)
		assert.True(t, result[0].Desc)
	})

	t.Run("space-separated asc", func(t *testing.T) {
		result, err := parseOrder("name asc")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "name", result[0].Column)
		assert.False(t, result[0].Desc)
	})

	t.Run("column only defaults to asc", func(t *testing.T) {
		result, err := parseOrder("created_at")
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "created_at", result[0].Column)
		assert.False(t, result[0].Desc)
	})

	t.Run("multiple columns dot-separated", func(t *testing.T) {
		result, err := parseOrder("created_at.desc,name.asc")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "created_at", result[0].Column)
		assert.True(t, result[0].Desc)
		assert.Equal(t, "name", result[1].Column)
		assert.False(t, result[1].Desc)
	})

	t.Run("multiple columns space-separated", func(t *testing.T) {
		result, err := parseOrder("created_at desc, name asc")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "created_at", result[0].Column)
		assert.True(t, result[0].Desc)
		assert.Equal(t, "name", result[1].Column)
		assert.False(t, result[1].Desc)
	})

	t.Run("mixed formats", func(t *testing.T) {
		result, err := parseOrder("created_at.desc, name asc")
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "created_at", result[0].Column)
		assert.True(t, result[0].Desc)
		assert.Equal(t, "name", result[1].Column)
		assert.False(t, result[1].Desc)
	})

	t.Run("empty string", func(t *testing.T) {
		result, err := parseOrder("")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("whitespace only", func(t *testing.T) {
		result, err := parseOrder("   ")
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestIsSQLExpression(t *testing.T) {
	t.Run("simple column names are not expressions", func(t *testing.T) {
		simpleNames := []string{
			"id",
			"user_id",
			"created_at",
			"firstName",
			"column123",
		}
		for _, name := range simpleNames {
			assert.False(t, isSQLExpression(name), "expected %s to not be an expression", name)
		}
	})

	t.Run("SQL functions are expressions", func(t *testing.T) {
		expressions := []string{
			"sum(visit_count)",
			"COUNT(*)",
			"avg(price)",
			"MIN(created_at)",
			"max(id)",
			"COALESCE(name, 'unknown')",
		}
		for _, expr := range expressions {
			assert.True(t, isSQLExpression(expr), "expected %s to be an expression", expr)
		}
	})

	t.Run("aliases are expressions", func(t *testing.T) {
		expressions := []string{
			"sum(x) as total",
			"name AS display_name",
			"id as identifier",
		}
		for _, expr := range expressions {
			assert.True(t, isSQLExpression(expr), "expected %s to be an expression", expr)
		}
	})

	t.Run("arithmetic expressions", func(t *testing.T) {
		expressions := []string{
			"price * quantity",
			"total - discount",
			"a + b",
			"count / 100",
		}
		for _, expr := range expressions {
			assert.True(t, isSQLExpression(expr), "expected %s to be an expression", expr)
		}
	})

	t.Run("type casting is expression", func(t *testing.T) {
		expressions := []string{
			"id::text",
			"created_at::date",
		}
		for _, expr := range expressions {
			assert.True(t, isSQLExpression(expr), "expected %s to be an expression", expr)
		}
	})

	t.Run("standalone * is not expression", func(t *testing.T) {
		assert.False(t, isSQLExpression("*"))
	})
}

func TestQuoteColumnOrExpression(t *testing.T) {
	t.Run("quotes simple column names", func(t *testing.T) {
		assert.Equal(t, `"id"`, quoteColumnOrExpression("id"))
		assert.Equal(t, `"user_id"`, quoteColumnOrExpression("user_id"))
		assert.Equal(t, `"created_at"`, quoteColumnOrExpression("created_at"))
	})

	t.Run("passes through SQL expressions unchanged", func(t *testing.T) {
		assert.Equal(t, "sum(visit_count)", quoteColumnOrExpression("sum(visit_count)"))
		assert.Equal(t, "COUNT(*)", quoteColumnOrExpression("COUNT(*)"))
		assert.Equal(t, "sum(visit_count) as total_visits", quoteColumnOrExpression("sum(visit_count) as total_visits"))
		assert.Equal(t, "price * quantity", quoteColumnOrExpression("price * quantity"))
	})
}

func TestParseFilterValue(t *testing.T) {
	t.Run("equal operator", func(t *testing.T) {
		filter, err := parseFilterValue("status", "eq.active")
		assert.NoError(t, err)
		assert.Equal(t, "status", filter.Column)
		assert.Equal(t, query.OpEqual, filter.Operator)
		assert.Equal(t, "active", filter.Value)
	})

	t.Run("not equal operator", func(t *testing.T) {
		filter, err := parseFilterValue("status", "neq.deleted")
		assert.NoError(t, err)
		assert.Equal(t, query.OpNotEqual, filter.Operator)
	})

	t.Run("greater than operator", func(t *testing.T) {
		filter, err := parseFilterValue("age", "gt.18")
		assert.NoError(t, err)
		assert.Equal(t, query.OpGreaterThan, filter.Operator)
		assert.Equal(t, "18", filter.Value)
	})

	t.Run("greater or equal operator", func(t *testing.T) {
		filter, err := parseFilterValue("score", "gte.90")
		assert.NoError(t, err)
		assert.Equal(t, query.OpGreaterOrEqual, filter.Operator)
	})

	t.Run("less than operator", func(t *testing.T) {
		filter, err := parseFilterValue("price", "lt.100")
		assert.NoError(t, err)
		assert.Equal(t, query.OpLessThan, filter.Operator)
	})

	t.Run("less or equal operator", func(t *testing.T) {
		filter, err := parseFilterValue("quantity", "lte.10")
		assert.NoError(t, err)
		assert.Equal(t, query.OpLessOrEqual, filter.Operator)
	})

	t.Run("like operator", func(t *testing.T) {
		filter, err := parseFilterValue("name", "like.%john%")
		assert.NoError(t, err)
		assert.Equal(t, query.OpLike, filter.Operator)
		assert.Equal(t, "%john%", filter.Value)
	})

	t.Run("ilike operator", func(t *testing.T) {
		filter, err := parseFilterValue("email", "ilike.%@example.com")
		assert.NoError(t, err)
		assert.Equal(t, query.OpILike, filter.Operator)
	})

	t.Run("is null", func(t *testing.T) {
		filter, err := parseFilterValue("deleted_at", "is.null")
		assert.NoError(t, err)
		assert.Equal(t, query.OpIs, filter.Operator)
		assert.Nil(t, filter.Value)
	})

	t.Run("is true", func(t *testing.T) {
		filter, err := parseFilterValue("active", "is.true")
		assert.NoError(t, err)
		assert.Equal(t, query.OpIs, filter.Operator)
		assert.Equal(t, true, filter.Value)
	})

	t.Run("is false", func(t *testing.T) {
		filter, err := parseFilterValue("verified", "is.false")
		assert.NoError(t, err)
		assert.Equal(t, query.OpIs, filter.Operator)
		assert.Equal(t, false, filter.Value)
	})

	t.Run("is not null", func(t *testing.T) {
		filter, err := parseFilterValue("email", "isnot.null")
		assert.NoError(t, err)
		assert.Equal(t, query.OpIsNot, filter.Operator)
		assert.Nil(t, filter.Value)
	})

	t.Run("in operator", func(t *testing.T) {
		filter, err := parseFilterValue("status", "in.(active,pending,review)")
		assert.NoError(t, err)
		assert.Equal(t, query.OpIn, filter.Operator)
		assert.Equal(t, []string{"active", "pending", "review"}, filter.Value)
	})

	t.Run("not in operator", func(t *testing.T) {
		filter, err := parseFilterValue("role", "nin.(guest,banned)")
		assert.NoError(t, err)
		assert.Equal(t, query.OpNotIn, filter.Operator)
	})

	t.Run("invalid format - no dot", func(t *testing.T) {
		_, err := parseFilterValue("column", "value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected format")
	})

	t.Run("invalid operator", func(t *testing.T) {
		_, err := parseFilterValue("column", "invalid.value")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid operator")
	})
}

func TestFilterToSQL(t *testing.T) {
	t.Run("equal operator", func(t *testing.T) {
		filter := query.Filter{Column: "status", Operator: query.OpEqual, Value: "active"}
		sql, arg := filterToSQL(filter, `"status"`, 1)
		assert.Equal(t, `"status" = $1`, sql)
		assert.Equal(t, "active", arg)
	})

	t.Run("not equal operator", func(t *testing.T) {
		filter := query.Filter{Column: "status", Operator: query.OpNotEqual, Value: "deleted"}
		sql, arg := filterToSQL(filter, `"status"`, 1)
		assert.Equal(t, `"status" <> $1`, sql)
		assert.Equal(t, "deleted", arg)
	})

	t.Run("greater than operator", func(t *testing.T) {
		filter := query.Filter{Column: "age", Operator: query.OpGreaterThan, Value: 18}
		sql, arg := filterToSQL(filter, `"age"`, 2)
		assert.Equal(t, `"age" > $2`, sql)
		assert.Equal(t, 18, arg)
	})

	t.Run("is null", func(t *testing.T) {
		filter := query.Filter{Column: "deleted_at", Operator: query.OpIs, Value: nil}
		sql, arg := filterToSQL(filter, `"deleted_at"`, 1)
		assert.Equal(t, `"deleted_at" IS NULL`, sql)
		assert.Nil(t, arg)
	})

	t.Run("is not null", func(t *testing.T) {
		filter := query.Filter{Column: "email", Operator: query.OpIsNot, Value: nil}
		sql, arg := filterToSQL(filter, `"email"`, 1)
		assert.Equal(t, `"email" IS NOT NULL`, sql)
		assert.Nil(t, arg)
	})

	t.Run("like operator", func(t *testing.T) {
		filter := query.Filter{Column: "name", Operator: query.OpLike, Value: "%john%"}
		sql, arg := filterToSQL(filter, `"name"`, 1)
		assert.Equal(t, `"name" LIKE $1`, sql)
		assert.Equal(t, "%john%", arg)
	})

	t.Run("in operator", func(t *testing.T) {
		filter := query.Filter{Column: "status", Operator: query.OpIn, Value: []string{"a", "b"}}
		sql, arg := filterToSQL(filter, `"status"`, 1)
		assert.Equal(t, `"status" = ANY($1)`, sql)
		assert.NotNil(t, arg)
	})

	t.Run("contains operator (jsonb)", func(t *testing.T) {
		filter := query.Filter{Column: "tags", Operator: query.OpContains, Value: `["tag1"]`}
		sql, arg := filterToSQL(filter, `"tags"`, 1)
		assert.Equal(t, `"tags" @> $1`, sql)
		assert.NotNil(t, arg)
	})
}

func TestBuildSelectQuery(t *testing.T) {
	t.Run("simple select all", func(t *testing.T) {
		sql, args := buildSelectQuery("public", "users", nil, nil, nil, 100, 0)
		assert.Contains(t, sql, `SELECT * FROM "public"."users"`)
		assert.Contains(t, sql, "LIMIT 100")
		assert.Empty(t, args)
	})

	t.Run("select specific columns", func(t *testing.T) {
		sql, args := buildSelectQuery("public", "users", []string{"id", "name", "email"}, nil, nil, 50, 0)
		assert.Contains(t, sql, `"id"`)
		assert.Contains(t, sql, `"name"`)
		assert.Contains(t, sql, `"email"`)
		assert.Contains(t, sql, "LIMIT 50")
		assert.Empty(t, args)
	})

	t.Run("with filters", func(t *testing.T) {
		filters := []query.Filter{
			{Column: "active", Operator: query.OpEqual, Value: true},
		}
		sql, args := buildSelectQuery("public", "users", nil, filters, nil, 100, 0)
		assert.Contains(t, sql, "WHERE")
		assert.Contains(t, sql, `"active" = $1`)
		assert.Len(t, args, 1)
	})

	t.Run("with order by", func(t *testing.T) {
		orderBy := []query.OrderBy{
			{Column: "created_at", Desc: true},
			{Column: "name", Desc: false},
		}
		sql, _ := buildSelectQuery("public", "users", nil, nil, orderBy, 100, 0)
		assert.Contains(t, sql, "ORDER BY")
		assert.Contains(t, sql, `"created_at" DESC`)
		assert.Contains(t, sql, `"name" ASC`)
	})

	t.Run("with offset", func(t *testing.T) {
		sql, _ := buildSelectQuery("public", "users", nil, nil, nil, 100, 50)
		assert.Contains(t, sql, "OFFSET 50")
	})

	t.Run("sql expressions in columns", func(t *testing.T) {
		sql, _ := buildSelectQuery("public", "orders", []string{"COUNT(*)", "sum(total) as revenue"}, nil, nil, 100, 0)
		assert.Contains(t, sql, "COUNT(*)")
		assert.Contains(t, sql, "sum(total) as revenue")
	})
}

func TestEmbeddingToString(t *testing.T) {
	t.Run("converts embedding to vector string", func(t *testing.T) {
		embedding := []float32{0.1, 0.2, 0.3}
		result := embeddingToString(embedding)
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
		assert.Contains(t, result, "0.1")
		assert.Contains(t, result, "0.2")
		assert.Contains(t, result, "0.3")
	})

	t.Run("empty embedding", func(t *testing.T) {
		embedding := []float32{}
		result := embeddingToString(embedding)
		assert.Equal(t, "[]", result)
	})

	t.Run("single element", func(t *testing.T) {
		embedding := []float32{0.5}
		result := embeddingToString(embedding)
		assert.Contains(t, result, "0.5")
	})
}
