package api

import (
	"net/url"
	"strings"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig creates a test config with default API settings for testing
func testConfig() *config.Config {
	return &config.Config{
		API: config.APIConfig{
			MaxPageSize:     -1, // Unlimited for most tests
			MaxTotalResults: -1, // Unlimited for most tests
			DefaultPageSize: -1, // No default for most tests
		},
	}
}

func TestQueryParser_ParseSelect(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name     string
		query    string
		expected []string
	}{
		{
			name:     "simple select",
			query:    "select=id,name,email",
			expected: []string{"id", "name", "email"},
		},
		{
			name:     "select with spaces",
			query:    "select=id, name, email",
			expected: []string{"id", "name", "email"},
		},
		{
			name:     "select with relation",
			query:    "select=id,name,posts(id,title)",
			expected: []string{"id", "name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, params.Select)
		})
	}
}

func TestQueryParser_ParseFilters(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedOp     FilterOperator
		expectedValue  interface{}
	}{
		{
			name:           "equal filter",
			query:          "name.eq=John",
			expectedColumn: "name",
			expectedOp:     OpEqual,
			expectedValue:  "John",
		},
		{
			name:           "greater than filter",
			query:          "age.gt=18",
			expectedColumn: "age",
			expectedOp:     OpGreaterThan,
			expectedValue:  "18",
		},
		{
			name:           "like filter",
			query:          "email.like=*@example.com",
			expectedColumn: "email",
			expectedOp:     OpLike,
			expectedValue:  "*@example.com",
		},
		{
			name:           "is null filter",
			query:          "deleted_at.is=null",
			expectedColumn: "deleted_at",
			expectedOp:     OpIs,
			expectedValue:  nil,
		},
		{
			name:           "in filter with array",
			query:          "status.in=queued,running",
			expectedColumn: "status",
			expectedOp:     OpIn,
			expectedValue:  []string{"queued", "running"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			assert.NoError(t, err)
			assert.Len(t, params.Filters, 1)
			assert.Equal(t, tt.expectedColumn, params.Filters[0].Column)
			assert.Equal(t, tt.expectedOp, params.Filters[0].Operator)
			assert.Equal(t, tt.expectedValue, params.Filters[0].Value)
		})
	}
}

func TestQueryParser_MultipleFiltersOnSameColumn(t *testing.T) {
	parser := NewQueryParser(testConfig())

	// Test range query: recorded_at=gte.2025-01-01&recorded_at=lte.2025-12-31
	// This should create TWO filters, not just one
	values := url.Values{}
	values.Add("recorded_at", "gte.2025-01-01")
	values.Add("recorded_at", "lte.2025-12-31")

	params, err := parser.Parse(values)
	require.NoError(t, err)
	require.Len(t, params.Filters, 2, "Expected 2 filters for range query")

	// Find gte and lte filters (order may vary due to map iteration)
	var gteFilter, lteFilter *Filter
	for i := range params.Filters {
		if params.Filters[i].Operator == OpGreaterOrEqual {
			gteFilter = &params.Filters[i]
		}
		if params.Filters[i].Operator == OpLessOrEqual {
			lteFilter = &params.Filters[i]
		}
	}

	require.NotNil(t, gteFilter, "Expected gte filter")
	assert.Equal(t, "recorded_at", gteFilter.Column)
	assert.Equal(t, "2025-01-01", gteFilter.Value)

	require.NotNil(t, lteFilter, "Expected lte filter")
	assert.Equal(t, "recorded_at", lteFilter.Column)
	assert.Equal(t, "2025-12-31", lteFilter.Value)
}

func TestQueryParser_ParseOrder(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedDesc   bool
		expectedNulls  string
	}{
		{
			name:           "ascending order",
			query:          "order=name.asc",
			expectedColumn: "name",
			expectedDesc:   false,
			expectedNulls:  "",
		},
		{
			name:           "descending order",
			query:          "order=created_at.desc",
			expectedColumn: "created_at",
			expectedDesc:   true,
			expectedNulls:  "",
		},
		{
			name:           "order with nulls last",
			query:          "order=updated_at.desc.nullslast",
			expectedColumn: "updated_at",
			expectedDesc:   true,
			expectedNulls:  "last",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			assert.NoError(t, err)
			assert.Len(t, params.Order, 1)
			assert.Equal(t, tt.expectedColumn, params.Order[0].Column)
			assert.Equal(t, tt.expectedDesc, params.Order[0].Desc)
			assert.Equal(t, tt.expectedNulls, params.Order[0].Nulls)
		})
	}
}

func TestQueryParser_ParsePagination(t *testing.T) {
	parser := NewQueryParser(testConfig())

	values, _ := url.ParseQuery("limit=10&offset=20")
	params, err := parser.Parse(values)

	assert.NoError(t, err)
	assert.NotNil(t, params.Limit)
	assert.Equal(t, 10, *params.Limit)
	assert.NotNil(t, params.Offset)
	assert.Equal(t, 20, *params.Offset)
}

func TestQueryParser_ParseCursor(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("parses cursor parameter", func(t *testing.T) {
		cursor := EncodeCursor("id", "abc123", false)
		values, _ := url.ParseQuery("cursor=" + cursor)
		params, err := parser.Parse(values)

		assert.NoError(t, err)
		assert.NotNil(t, params.Cursor)
		assert.Equal(t, cursor, *params.Cursor)
	})

	t.Run("parses cursor_column parameter", func(t *testing.T) {
		values, _ := url.ParseQuery("cursor_column=created_at")
		params, err := parser.Parse(values)

		assert.NoError(t, err)
		assert.NotNil(t, params.CursorColumn)
		assert.Equal(t, "created_at", *params.CursorColumn)
	})

	t.Run("parses both cursor and cursor_column", func(t *testing.T) {
		cursor := EncodeCursor("id", "abc123", false)
		values, _ := url.ParseQuery("cursor=" + cursor + "&cursor_column=updated_at")
		params, err := parser.Parse(values)

		assert.NoError(t, err)
		assert.NotNil(t, params.Cursor)
		assert.NotNil(t, params.CursorColumn)
		assert.Equal(t, "updated_at", *params.CursorColumn)
	})

	t.Run("rejects invalid cursor_column", func(t *testing.T) {
		values, _ := url.ParseQuery("cursor_column=invalid-column-name")
		_, err := parser.Parse(values)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid cursor_column")
	})

	t.Run("empty cursor is ignored", func(t *testing.T) {
		values, _ := url.ParseQuery("cursor=")
		params, err := parser.Parse(values)

		assert.NoError(t, err)
		assert.Nil(t, params.Cursor)
	})
}

func TestQueryParams_ToSQL(t *testing.T) {
	tests := []struct {
		name         string
		params       QueryParams
		expectedSQL  string
		expectedArgs []interface{}
	}{
		{
			name: "simple where clause",
			params: QueryParams{
				Filters: []Filter{
					{Column: "name", Operator: OpEqual, Value: "John"},
				},
			},
			expectedSQL:  `WHERE "name" = $1`,
			expectedArgs: []interface{}{"John"},
		},
		{
			name: "multiple filters",
			params: QueryParams{
				Filters: []Filter{
					{Column: "name", Operator: OpEqual, Value: "John"},
					{Column: "age", Operator: OpGreaterThan, Value: "18"},
				},
			},
			expectedSQL:  `WHERE "name" = $1 AND "age" > $2`,
			expectedArgs: []interface{}{"John", "18"},
		},
		{
			name: "in filter with string array",
			params: QueryParams{
				Filters: []Filter{
					{Column: "status", Operator: OpIn, Value: []string{"queued", "running"}},
				},
			},
			expectedSQL:  `WHERE "status" = ANY($1)`,
			expectedArgs: []interface{}{[]string{"queued", "running"}},
		},
		{
			name: "in filter with single element",
			params: QueryParams{
				Filters: []Filter{
					{Column: "status", Operator: OpIn, Value: []string{"active"}},
				},
			},
			expectedSQL:  `WHERE "status" = ANY($1)`,
			expectedArgs: []interface{}{[]string{"active"}},
		},
		{
			name: "in filter with multiple filters",
			params: QueryParams{
				Filters: []Filter{
					{Column: "user_id", Operator: OpEqual, Value: "123"},
					{Column: "status", Operator: OpIn, Value: []string{"queued", "running"}},
				},
			},
			expectedSQL:  `WHERE "user_id" = $1 AND "status" = ANY($2)`,
			expectedArgs: []interface{}{"123", []string{"queued", "running"}},
		},
		{
			name: "with order and limit",
			params: QueryParams{
				Order: []OrderBy{
					{Column: "created_at", Desc: true},
				},
				Limit: intPtr(10),
			},
			expectedSQL:  `ORDER BY "created_at" DESC LIMIT $1`,
			expectedArgs: []interface{}{10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := tt.params.ToSQL("users")
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestQueryParser_ParseAggregations(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name                 string
		query                string
		expectedSelect       []string
		expectedAggregations []Aggregation
	}{
		{
			name:           "count(*)",
			query:          "select=count(*)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggCountAll, Column: "", Alias: ""},
			},
		},
		{
			name:           "count(column)",
			query:          "select=count(id)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggCount, Column: "id", Alias: ""},
			},
		},
		{
			name:           "sum",
			query:          "select=sum(price)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggSum, Column: "price", Alias: ""},
			},
		},
		{
			name:           "avg",
			query:          "select=avg(rating)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggAvg, Column: "rating", Alias: ""},
			},
		},
		{
			name:           "min",
			query:          "select=min(created_at)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggMin, Column: "created_at", Alias: ""},
			},
		},
		{
			name:           "max",
			query:          "select=max(updated_at)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggMax, Column: "updated_at", Alias: ""},
			},
		},
		{
			name:           "multiple aggregations",
			query:          "select=count(*),sum(price),avg(rating)",
			expectedSelect: []string{},
			expectedAggregations: []Aggregation{
				{Function: AggCountAll, Column: "", Alias: ""},
				{Function: AggSum, Column: "price", Alias: ""},
				{Function: AggAvg, Column: "rating", Alias: ""},
			},
		},
		{
			name:           "aggregation with regular fields",
			query:          "select=category,count(*),sum(price)",
			expectedSelect: []string{"category"},
			expectedAggregations: []Aggregation{
				{Function: AggCountAll, Column: "", Alias: ""},
				{Function: AggSum, Column: "price", Alias: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedSelect, params.Select)
			assert.Equal(t, len(tt.expectedAggregations), len(params.Aggregations))

			for i, expectedAgg := range tt.expectedAggregations {
				assert.Equal(t, expectedAgg.Function, params.Aggregations[i].Function)
				assert.Equal(t, expectedAgg.Column, params.Aggregations[i].Column)
			}
		})
	}
}

func TestQueryParser_ParseGroupBy(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name            string
		query           string
		expectedGroupBy []string
	}{
		{
			name:            "single group by",
			query:           "group_by=category",
			expectedGroupBy: []string{"category"},
		},
		{
			name:            "multiple group by",
			query:           "group_by=category,status",
			expectedGroupBy: []string{"category", "status"},
		},
		{
			name:            "group by with spaces",
			query:           "group_by=category, status, region",
			expectedGroupBy: []string{"category", "status", "region"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedGroupBy, params.GroupBy)
		})
	}
}

func TestQueryParams_BuildSelectClause(t *testing.T) {
	tests := []struct {
		name        string
		params      QueryParams
		expectedSQL string
	}{
		{
			name: "aggregation only - count(*)",
			params: QueryParams{
				Aggregations: []Aggregation{
					{Function: AggCountAll, Column: "", Alias: ""},
				},
			},
			expectedSQL: `COUNT(*) AS "count"`,
		},
		{
			name: "aggregation only - sum",
			params: QueryParams{
				Aggregations: []Aggregation{
					{Function: AggSum, Column: "price", Alias: ""},
				},
			},
			expectedSQL: `SUM("price") AS "sum_price"`,
		},
		{
			name: "multiple aggregations",
			params: QueryParams{
				Aggregations: []Aggregation{
					{Function: AggCount, Column: "id", Alias: ""},
					{Function: AggSum, Column: "price", Alias: ""},
					{Function: AggAvg, Column: "rating", Alias: ""},
				},
			},
			expectedSQL: `COUNT("id") AS "count_id", SUM("price") AS "sum_price", AVG("rating") AS "avg_rating"`,
		},
		{
			name: "fields with aggregations",
			params: QueryParams{
				Select: []string{"category"},
				Aggregations: []Aggregation{
					{Function: AggCountAll, Column: "", Alias: ""},
					{Function: AggSum, Column: "price", Alias: "total"},
				},
			},
			expectedSQL: `"category", COUNT(*) AS "count", SUM("price") AS "total"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := tt.params.BuildSelectClause("products")
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestQueryParams_BuildGroupByClause(t *testing.T) {
	tests := []struct {
		name        string
		params      QueryParams
		expectedSQL string
	}{
		{
			name: "no group by",
			params: QueryParams{
				GroupBy: []string{},
			},
			expectedSQL: "",
		},
		{
			name: "single group by",
			params: QueryParams{
				GroupBy: []string{"category"},
			},
			expectedSQL: ` GROUP BY "category"`,
		},
		{
			name: "multiple group by",
			params: QueryParams{
				GroupBy: []string{"category", "status", "region"},
			},
			expectedSQL: ` GROUP BY "category", "status", "region"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := tt.params.BuildGroupByClause()
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestAggregation_ToSQL(t *testing.T) {
	tests := []struct {
		name        string
		agg         Aggregation
		expectedSQL string
	}{
		{
			name:        "COUNT(*)",
			agg:         Aggregation{Function: AggCountAll, Column: "", Alias: ""},
			expectedSQL: `COUNT(*) AS "count"`,
		},
		{
			name:        "COUNT(column)",
			agg:         Aggregation{Function: AggCount, Column: "id", Alias: ""},
			expectedSQL: `COUNT("id") AS "count_id"`,
		},
		{
			name:        "SUM",
			agg:         Aggregation{Function: AggSum, Column: "price", Alias: ""},
			expectedSQL: `SUM("price") AS "sum_price"`,
		},
		{
			name:        "AVG",
			agg:         Aggregation{Function: AggAvg, Column: "rating", Alias: ""},
			expectedSQL: `AVG("rating") AS "avg_rating"`,
		},
		{
			name:        "MIN",
			agg:         Aggregation{Function: AggMin, Column: "price", Alias: ""},
			expectedSQL: `MIN("price") AS "min_price"`,
		},
		{
			name:        "MAX",
			agg:         Aggregation{Function: AggMax, Column: "price", Alias: ""},
			expectedSQL: `MAX("price") AS "max_price"`,
		},
		{
			name:        "custom alias",
			agg:         Aggregation{Function: AggSum, Column: "price", Alias: "total"},
			expectedSQL: `SUM("price") AS "total"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := tt.agg.ToSQL()
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestPaginationLimitEnforcement(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		queryString    string
		expectedLimit  *int
		expectedOffset *int
		description    string
	}{
		{
			name: "Enforce max_page_size - cap requested limit",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     1000,
					MaxTotalResults: -1,
					DefaultPageSize: -1,
				},
			},
			queryString:   "limit=5000",
			expectedLimit: intPtr(1000),
			description:   "Requested limit of 5000 should be capped to max_page_size of 1000",
		},
		{
			name: "Apply default_page_size when no limit specified",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     10000,
					MaxTotalResults: -1,
					DefaultPageSize: 1000,
				},
			},
			queryString:   "",
			expectedLimit: intPtr(1000),
			description:   "No limit specified should apply default_page_size of 1000",
		},
		{
			name: "No default applied when default_page_size is -1",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     10000,
					MaxTotalResults: -1,
					DefaultPageSize: -1,
				},
			},
			queryString:   "",
			expectedLimit: nil,
			description:   "When default_page_size is -1, no default limit should be applied",
		},
		{
			name: "Enforce max_total_results - cap limit based on offset",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     5000,
					MaxTotalResults: 10000,
					DefaultPageSize: -1,
				},
			},
			queryString:    "offset=9500&limit=1000",
			expectedLimit:  intPtr(500),
			expectedOffset: intPtr(9500),
			description:    "Offset 9500 + limit 1000 exceeds max_total_results 10000, should cap limit to 500",
		},
		{
			name: "Enforce max_total_results - zero limit when offset exceeds max",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     5000,
					MaxTotalResults: 10000,
					DefaultPageSize: -1,
				},
			},
			queryString:    "offset=10500&limit=1000",
			expectedLimit:  intPtr(0),
			expectedOffset: intPtr(10500),
			description:    "Offset 10500 exceeds max_total_results 10000, should cap limit to 0",
		},
		{
			name: "Allow unlimited when max_page_size is -1",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     -1,
					MaxTotalResults: -1,
					DefaultPageSize: -1,
				},
			},
			queryString:   "limit=100000",
			expectedLimit: intPtr(100000),
			description:   "When max_page_size is -1, allow any limit",
		},
		{
			name: "Combine max_page_size and max_total_results enforcement",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     1000,
					MaxTotalResults: 5000,
					DefaultPageSize: 500,
				},
			},
			queryString:    "offset=4500&limit=2000",
			expectedLimit:  intPtr(500),
			expectedOffset: intPtr(4500),
			description:    "Limit 2000 capped to max_page_size 1000, then further capped to 500 due to max_total_results",
		},
		{
			name: "Default limit respects max_total_results",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize:     5000,
					MaxTotalResults: 10000,
					DefaultPageSize: 1000,
				},
			},
			queryString:    "offset=9800",
			expectedLimit:  intPtr(200),
			expectedOffset: intPtr(9800),
			description:    "Default limit 1000 applied, then capped to 200 due to max_total_results",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(tt.config)
			values, err := url.ParseQuery(tt.queryString)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			require.NoError(t, err, tt.description)

			if tt.expectedLimit != nil {
				require.NotNil(t, params.Limit, "Expected limit to be set but got nil")
				assert.Equal(t, *tt.expectedLimit, *params.Limit, tt.description)
			} else {
				assert.Nil(t, params.Limit, "Expected limit to be nil but got a value")
			}

			if tt.expectedOffset != nil {
				require.NotNil(t, params.Offset, "Expected offset to be set but got nil")
				assert.Equal(t, *tt.expectedOffset, *params.Offset)
			}
		})
	}
}

func TestParseJSONBPath(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		expected string
	}{
		{
			name:     "simple column",
			column:   "name",
			expected: `"name"`,
		},
		{
			name:     "json access single key",
			column:   "data->key",
			expected: `"data"->'key'`,
		},
		{
			name:     "text access single key",
			column:   "data->>key",
			expected: `"data"->>'key'`,
		},
		{
			name:     "chained json access",
			column:   "data->nested->value",
			expected: `"data"->'nested'->'value'`,
		},
		{
			name:     "mixed json and text access",
			column:   "data->nested->>value",
			expected: `"data"->'nested'->>'value'`,
		},
		{
			name:     "deep nesting",
			column:   "a->b->c->d->>e",
			expected: `"a"->'b'->'c'->'d'->>'e'`,
		},
		{
			name:     "array index",
			column:   "data->0",
			expected: `"data"->0`,
		},
		{
			name:     "array index with nested key",
			column:   "data->0->name",
			expected: `"data"->0->'name'`,
		},
		{
			name:     "array index with text extraction",
			column:   "data->0->>name",
			expected: `"data"->0->>'name'`,
		},
		{
			name:     "realistic geocode example",
			column:   "geocode->properties->>country",
			expected: `"geocode"->'properties'->>'country'`,
		},
		{
			name:     "metadata stats count",
			column:   "metadata->stats->>count",
			expected: `"metadata"->'stats'->>'count'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseJSONBPath(tt.column)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterToSQLWithJSONBPath(t *testing.T) {
	tests := []struct {
		name        string
		filter      Filter
		expectedSQL string
		expectValue bool
	}{
		{
			name: "simple column equality",
			filter: Filter{
				Column:   "name",
				Operator: OpEqual,
				Value:    "John",
			},
			expectedSQL: `"name" = $1`,
			expectValue: true,
		},
		{
			name: "jsonb path equality",
			filter: Filter{
				Column:   "data->key",
				Operator: OpEqual,
				Value:    "value",
			},
			expectedSQL: `"data"->'key' = $1`,
			expectValue: true,
		},
		{
			name: "jsonb text extraction equality",
			filter: Filter{
				Column:   "data->>key",
				Operator: OpEqual,
				Value:    "value",
			},
			expectedSQL: `"data"->>'key' = $1`,
			expectValue: true,
		},
		{
			name: "nested jsonb path IS NULL",
			filter: Filter{
				Column:   "geocode->properties->>country",
				Operator: OpIs,
				Value:    nil,
			},
			expectedSQL: `"geocode"->'properties'->>'country' IS NULL`,
			expectValue: false,
		},
		{
			name: "jsonb text extraction greater than with numeric",
			filter: Filter{
				Column:   "metadata->stats->>count",
				Operator: OpGreaterThan,
				Value:    10,
			},
			expectedSQL: `("metadata"->'stats'->>'count')::numeric > $1`,
			expectValue: true,
		},
		{
			name: "jsonb text extraction less than with string number",
			filter: Filter{
				Column:   "data->>amount",
				Operator: OpLessThan,
				Value:    "100",
			},
			expectedSQL: `("data"->>'amount')::numeric < $1`,
			expectValue: true,
		},
		{
			name: "jsonb json access greater than (no cast)",
			filter: Filter{
				Column:   "data->count",
				Operator: OpGreaterThan,
				Value:    10,
			},
			expectedSQL: `"data"->'count' > $1`,
			expectValue: true,
		},
		{
			name: "jsonb IN operator",
			filter: Filter{
				Column:   "data->>status",
				Operator: OpIn,
				Value:    []string{"active", "pending"},
			},
			expectedSQL: `"data"->>'status' = ANY($1)`,
			expectValue: true,
		},
		{
			name: "jsonb LIKE operator",
			filter: Filter{
				Column:   "data->>email",
				Operator: OpLike,
				Value:    "%@example.com",
			},
			expectedSQL: `"data"->>'email' LIKE $1`,
			expectValue: true,
		},
		{
			name: "array index access",
			filter: Filter{
				Column:   "items->0->>name",
				Operator: OpEqual,
				Value:    "first",
			},
			expectedSQL: `"items"->0->>'name' = $1`,
			expectValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argCounter := 1
			sql, value := filterToSQL(tt.filter, &argCounter)

			assert.Equal(t, tt.expectedSQL, sql)

			if tt.expectValue {
				assert.Equal(t, tt.filter.Value, value)
			} else {
				assert.Nil(t, value)
			}
		})
	}
}

func TestNeedsNumericCast(t *testing.T) {
	tests := []struct {
		name     string
		column   string
		value    interface{}
		expected bool
	}{
		{
			name:     "text extraction with int",
			column:   "data->>count",
			value:    10,
			expected: true,
		},
		{
			name:     "text extraction with float",
			column:   "data->>price",
			value:    19.99,
			expected: true,
		},
		{
			name:     "text extraction with string number",
			column:   "data->>count",
			value:    "10",
			expected: true,
		},
		{
			name:     "text extraction with non-numeric string",
			column:   "data->>name",
			value:    "John",
			expected: false,
		},
		{
			name:     "json access with int (no cast needed)",
			column:   "data->count",
			value:    10,
			expected: false,
		},
		{
			name:     "simple column with int (no cast)",
			column:   "count",
			value:    10,
			expected: false,
		},
		{
			name:     "nested text extraction with int",
			column:   "metadata->stats->>total",
			value:    100,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := needsNumericCast(tt.column, tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParser_NestedLogicalFilters(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedCount  int
		expectOrGroups bool
	}{
		{
			name:           "simple or filter",
			query:          "or=(name.eq.John,name.eq.Jane)",
			expectedCount:  2,
			expectOrGroups: false,
		},
		{
			name:           "nested or in and filter",
			query:          "and=(or(col.lt.10,col.gt.20),or(col.lt.30,col.gt.40))",
			expectedCount:  4,
			expectOrGroups: true,
		},
		{
			name:           "complex nested expression",
			query:          "and=(or(date.lt.2024-01-01,date.gt.2024-01-10),or(date.lt.2024-02-01,date.gt.2024-02-10),or(date.lt.2024-03-01,date.gt.2024-03-10))",
			expectedCount:  6,
			expectOrGroups: true,
		},
		{
			name:           "or filter with is.null",
			query:          "or=(name.is.null,name.eq.)",
			expectedCount:  2,
			expectOrGroups: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedCount, len(params.Filters))

			if tt.expectOrGroups {
				// Check that filters have OrGroupID set
				groupIDs := make(map[int]bool)
				for _, f := range params.Filters {
					if f.OrGroupID > 0 {
						groupIDs[f.OrGroupID] = true
					}
				}
				assert.Greater(t, len(groupIDs), 0, "expected OR groups to be assigned")
			}
		})
	}
}

func TestQueryParser_OrFilterIsNullValueParsing(t *testing.T) {
	parser := NewQueryParser(testConfig())

	// Test that is.null in OR filters gets properly parsed to nil, not string "null"
	values, _ := url.ParseQuery("or=(name.is.null,status.is.true,active.is.false)")
	params, err := parser.Parse(values)

	require.NoError(t, err)
	require.Equal(t, 3, len(params.Filters))

	// Find each filter and verify its value
	for _, f := range params.Filters {
		switch f.Column {
		case "name":
			assert.Equal(t, OpIs, f.Operator)
			assert.Nil(t, f.Value, "is.null should parse to nil, not string 'null'")
		case "status":
			assert.Equal(t, OpIs, f.Operator)
			assert.Equal(t, true, f.Value, "is.true should parse to bool true")
		case "active":
			assert.Equal(t, OpIs, f.Operator)
			assert.Equal(t, false, f.Value, "is.false should parse to bool false")
		}
	}

	// Verify SQL generation produces IS NULL, not IS $1
	argCounter := 1
	whereClause, args := params.buildWhereClause(&argCounter)
	assert.Contains(t, whereClause, "IS NULL")
	assert.Contains(t, whereClause, "IS $1") // for true
	assert.Contains(t, whereClause, "IS $2") // for false
	assert.Equal(t, 2, len(args), "should have 2 args (true, false), null should not be parameterized")
}

func TestQueryParser_ParseNestedFilters(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "simple comma separated",
			value:    "a.eq.1,b.eq.2",
			expected: []string{"a.eq.1", "b.eq.2"},
		},
		{
			name:     "nested parentheses",
			value:    "or(a.eq.1,b.eq.2),or(c.eq.3,d.eq.4)",
			expected: []string{"or(a.eq.1,b.eq.2)", "or(c.eq.3,d.eq.4)"},
		},
		{
			name:     "single nested expression",
			value:    "or(col.lt.10,col.gt.20)",
			expected: []string{"or(col.lt.10,col.gt.20)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.parseNestedFilters(tt.value)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParams_BuildWhereClause_OrGroups(t *testing.T) {
	tests := []struct {
		name           string
		filters        []Filter
		expectedParts  []string
		unexpectedPart string
	}{
		{
			name: "separate OR groups",
			filters: []Filter{
				{Column: "col", Operator: OpLessThan, Value: "10", IsOr: true, OrGroupID: 1},
				{Column: "col", Operator: OpGreaterThan, Value: "20", IsOr: true, OrGroupID: 1},
				{Column: "col", Operator: OpLessThan, Value: "30", IsOr: true, OrGroupID: 2},
				{Column: "col", Operator: OpGreaterThan, Value: "40", IsOr: true, OrGroupID: 2},
			},
			expectedParts: []string{
				`("col" < $1 OR "col" > $2)`,
				`("col" < $3 OR "col" > $4)`,
				" AND ",
			},
			unexpectedPart: "col" + ` < $1 OR "col" > $2 OR "col" < $3`, // Should NOT group all together
		},
		{
			name: "mixed AND and OR groups",
			filters: []Filter{
				{Column: "status", Operator: OpEqual, Value: "active", IsOr: false},
				{Column: "col", Operator: OpLessThan, Value: "10", IsOr: true, OrGroupID: 1},
				{Column: "col", Operator: OpGreaterThan, Value: "20", IsOr: true, OrGroupID: 1},
			},
			expectedParts: []string{
				`"status" = $1`,
				`("col" < $2 OR "col" > $3)`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &QueryParams{Filters: tt.filters}
			argCounter := 1
			whereClause, _ := params.buildWhereClause(&argCounter)

			for _, expected := range tt.expectedParts {
				assert.Contains(t, whereClause, expected)
			}

			if tt.unexpectedPart != "" {
				assert.NotContains(t, whereClause, tt.unexpectedPart)
			}
		})
	}
}

func TestFormatVectorValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string with brackets",
			input:    "[0.1,0.2,0.3]",
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "string without brackets",
			input:    "0.1,0.2,0.3",
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "float64 slice",
			input:    []float64{0.1, 0.2, 0.3},
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "float32 slice",
			input:    []float32{0.1, 0.2, 0.3},
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "interface slice with floats",
			input:    []interface{}{0.1, 0.2, 0.3},
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "interface slice with ints",
			input:    []interface{}{1, 2, 3},
			expected: "[1,2,3]",
		},
		{
			name:     "empty slice",
			input:    []float64{},
			expected: "[]",
		},
		{
			name:     "string with leading bracket only",
			input:    "[0.1,0.2",
			expected: "[0.1,0.2]",
		},
		{
			name:     "string with trailing bracket only",
			input:    "0.1,0.2]",
			expected: "[0.1,0.2]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVectorValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseSTDWithinValue(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedDistance float64
		expectedGeometry string
		expectError      bool
		errorContains    string
	}{
		{
			name:             "valid point with integer distance",
			input:            `1000,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectedDistance: 1000,
			expectedGeometry: `{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:      false,
		},
		{
			name:             "valid point with float distance",
			input:            `1500.5,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectedDistance: 1500.5,
			expectedGeometry: `{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:      false,
		},
		{
			name:             "valid polygon with distance",
			input:            `500,{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.5,37.85],[-122.35,37.85],[-122.35,37.7],[-122.5,37.7]]]}`,
			expectedDistance: 500,
			expectedGeometry: `{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.5,37.85],[-122.35,37.85],[-122.35,37.7],[-122.5,37.7]]]}`,
			expectError:      false,
		},
		{
			name:             "zero distance",
			input:            `0,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectedDistance: 0,
			expectedGeometry: `{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:      false,
		},
		{
			name:          "negative distance",
			input:         `-100,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:   true,
			errorContains: "distance cannot be negative",
		},
		{
			name:          "missing distance",
			input:         `{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:   true,
			errorContains: "st_dwithin value must be in format",
		},
		{
			name:          "invalid distance - not a number",
			input:         `abc,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:   true,
			errorContains: "invalid distance value",
		},
		{
			name:          "missing geometry",
			input:         `1000,`,
			expectError:   true,
			errorContains: "geometry must be a valid GeoJSON object",
		},
		{
			name:          "invalid geometry - not JSON",
			input:         `1000,not-json`,
			expectError:   true,
			errorContains: "geometry must be a valid GeoJSON object",
		},
		{
			name:          "empty input",
			input:         ``,
			expectError:   true,
			errorContains: "st_dwithin value must be in format",
		},
		{
			name:          "only comma",
			input:         `,`,
			expectError:   true,
			errorContains: "st_dwithin value must be in format",
		},
		{
			name:             "distance with spaces",
			input:            ` 1000 , {"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectedDistance: 1000,
			expectedGeometry: `{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance, geometry, err := parseSTDWithinValue(tt.input)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedDistance, distance)
				assert.Equal(t, tt.expectedGeometry, geometry)
			}
		})
	}
}

func TestSTDWithinFilter(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectSQL   string
		expectArgs  []interface{}
		expectError bool
	}{
		{
			name:       "st_dwithin with point",
			query:      `location.st_dwithin=1000,{"type":"Point","coordinates":[-122.4783,37.8199]}`,
			expectSQL:  `ST_DWithin("location", ST_GeomFromGeoJSON($1), $2)`,
			expectArgs: []interface{}{`{"type":"Point","coordinates":[-122.4783,37.8199]}`, float64(1000)},
		},
		{
			name:       "st_dwithin with polygon",
			query:      `geom.st_dwithin=500.5,{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.5,37.85],[-122.35,37.85],[-122.5,37.7]]]}`,
			expectSQL:  `ST_DWithin("geom", ST_GeomFromGeoJSON($1), $2)`,
			expectArgs: []interface{}{`{"type":"Polygon","coordinates":[[[-122.5,37.7],[-122.5,37.85],[-122.35,37.85],[-122.5,37.7]]]}`, float64(500.5)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			argCounter := 1
			sql, args := params.buildWhereClause(&argCounter)

			assert.Equal(t, tt.expectSQL, sql)
			assert.Equal(t, tt.expectArgs, args)
		})
	}
}

func TestParseVectorOrder(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name   string
		order  string
		want   OrderBy
		wantOK bool
	}{
		{
			name:  "vector L2 order ascending",
			order: "embedding.vec_l2.[0.1,0.2,0.3].asc",
			want: OrderBy{
				Column:      "embedding",
				Desc:        false,
				VectorOp:    OpVectorL2,
				VectorValue: "[0.1,0.2,0.3]",
			},
			wantOK: true,
		},
		{
			name:  "vector cosine order ascending",
			order: "embedding.vec_cos.[0.1,0.2,0.3].asc",
			want: OrderBy{
				Column:      "embedding",
				Desc:        false,
				VectorOp:    OpVectorCosine,
				VectorValue: "[0.1,0.2,0.3]",
			},
			wantOK: true,
		},
		{
			name:  "vector inner product order descending",
			order: "embedding.vec_ip.[0.1,0.2,0.3].desc",
			want: OrderBy{
				Column:      "embedding",
				Desc:        true,
				VectorOp:    OpVectorIP,
				VectorValue: "[0.1,0.2,0.3]",
			},
			wantOK: true,
		},
		{
			name:  "vector order with default direction (ascending)",
			order: "embedding.vec_l2.[1,2,3]",
			want: OrderBy{
				Column:      "embedding",
				Desc:        false,
				VectorOp:    OpVectorL2,
				VectorValue: "[1,2,3]",
			},
			wantOK: true,
		},
		{
			name:  "vector order with many dimensions",
			order: "features.vec_cos.[0.1,0.2,0.3,0.4,0.5,0.6,0.7,0.8].desc",
			want: OrderBy{
				Column:      "features",
				Desc:        true,
				VectorOp:    OpVectorCosine,
				VectorValue: "[0.1,0.2,0.3,0.4,0.5,0.6,0.7,0.8]",
			},
			wantOK: true,
		},
		{
			name:   "not a vector order - regular column",
			order:  "created_at.desc",
			wantOK: false,
		},
		{
			name:   "not a vector order - regular ascending",
			order:  "name.asc",
			wantOK: false,
		},
		{
			name:   "invalid - no operator",
			order:  "embedding.[0.1,0.2]",
			wantOK: false,
		},
		{
			name:   "invalid - missing brackets",
			order:  "embedding.vec_l2.0.1,0.2,0.3.asc",
			wantOK: false,
		},
		{
			name:   "invalid - empty column",
			order:  ".vec_l2.[0.1,0.2].asc",
			wantOK: false,
		},
		{
			name:   "invalid - operator at start",
			order:  "vec_l2.[0.1,0.2].asc",
			wantOK: false,
		},
		{
			name:   "invalid column name - contains special chars",
			order:  "embed-ding.vec_l2.[0.1,0.2].asc",
			wantOK: false,
		},
		{
			name:   "invalid - missing closing bracket",
			order:  "embedding.vec_l2.[0.1,0.2.asc",
			wantOK: false,
		},
		{
			name:   "invalid - missing opening bracket",
			order:  "embedding.vec_l2.0.1,0.2].asc",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parser.parseVectorOrder(tt.order)
			assert.Equal(t, tt.wantOK, ok, "unexpected ok value")
			if tt.wantOK {
				assert.Equal(t, tt.want.Column, got.Column)
				assert.Equal(t, tt.want.Desc, got.Desc)
				assert.Equal(t, tt.want.VectorOp, got.VectorOp)
				assert.Equal(t, tt.want.VectorValue, got.VectorValue)
			}
		})
	}
}

func TestParseVectorOrderIntegration(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name         string
		query        string
		expectColumn string
		expectVecOp  FilterOperator
		expectVecVal string
		expectDesc   bool
	}{
		{
			name:         "vector L2 order via query parameter",
			query:        "order=embedding.vec_l2.[0.5,0.5,0.5].asc",
			expectColumn: "embedding",
			expectVecOp:  OpVectorL2,
			expectVecVal: "[0.5,0.5,0.5]",
			expectDesc:   false,
		},
		{
			name:         "vector cosine order descending",
			query:        "order=features.vec_cos.[1,2,3].desc",
			expectColumn: "features",
			expectVecOp:  OpVectorCosine,
			expectVecVal: "[1,2,3]",
			expectDesc:   true,
		},
		{
			name:         "vector inner product order",
			query:        "order=document_embedding.vec_ip.[0.1,0.2,0.3,0.4]",
			expectColumn: "document_embedding",
			expectVecOp:  OpVectorIP,
			expectVecVal: "[0.1,0.2,0.3,0.4]",
			expectDesc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			require.NoError(t, err)
			require.Len(t, params.Order, 1)

			order := params.Order[0]
			assert.Equal(t, tt.expectColumn, order.Column)
			assert.Equal(t, tt.expectVecOp, order.VectorOp)
			assert.Equal(t, tt.expectVecVal, order.VectorValue)
			assert.Equal(t, tt.expectDesc, order.Desc)
		})
	}
}

// =============================================================================
// Additional Tests for Coverage Boost (Priority 1.1)
// =============================================================================

// TestFilterToSQL_EdgeCases tests additional filter scenarios
func TestFilterToSQL_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		filter      Filter
		expectedSQL string
		expectValue interface{}
		expectError bool
	}{
		{
			name: "OpLike with escape characters",
			filter: Filter{
				Column:   "name",
				Operator: OpLike,
				Value:    "test\\_value",
			},
			expectedSQL: `"name" LIKE $1`,
			expectValue: "test\\_value",
			expectError: false,
		},
		{
			name: "OpILike case insensitive",
			filter: Filter{
				Column:   "email",
				Operator: OpILike,
				Value:    "*@GMAIL.COM",
			},
			expectedSQL: `"email" ILIKE $1`,
			expectValue: "*@GMAIL.COM",
			expectError: false,
		},
		{
			name: "OpLike with empty pattern",
			filter: Filter{
				Column:   "name",
				Operator: OpLike,
				Value:    "",
			},
			expectedSQL: `"name" LIKE $1`,
			expectValue: "",
			expectError: false,
		},
		{
			name: "OpIn with single value",
			filter: Filter{
				Column:   "status",
				Operator: OpIn,
				Value:    []string{"active"},
			},
			expectedSQL: `"status" = ANY($1)`,
			expectValue: []string{"active"},
			expectError: false,
		},
		{
			name: "OpIn with multiple values",
			filter: Filter{
				Column:   "id",
				Operator: OpIn,
				Value:    []string{"1", "2", "3"},
			},
			expectedSQL: `"id" = ANY($1)`,
			expectValue: []string{"1", "2", "3"},
			expectError: false,
		},
		{
			name: "OpNotIn with values - handled as equality",
			filter: Filter{
				Column:   "status",
				Operator: OpNotIn,
				Value:    []string{"deleted", "archived"},
			},
			expectedSQL: `"status" = $1`,
			expectValue: []string{"deleted", "archived"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			argCounter := 1
			sql, value := filterToSQL(tt.filter, &argCounter)

			if tt.expectError {
				assert.Empty(t, sql)
			} else {
				assert.NotEmpty(t, sql)
			}

			if tt.expectedSQL != "" {
				assert.Equal(t, tt.expectedSQL, sql)
			}
			if tt.expectValue != nil {
				assert.Equal(t, tt.expectValue, value)
			}
		})
	}
}

// TestParseFilter_EdgeCases tests filter parsing edge cases
func TestParseFilter_EdgeCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "malformed operator - missing value",
			query:       "name.eq=",
			expectError: false, // Empty value is valid
		},
		{
			name:        "multiple filters on same column",
			query:       "age.gt=18&age.lt=65",
			expectError: false,
		},
		{
			name:        "special characters in value",
			query:       "description.like=*test*",
			expectError: false,
		},
		{
			name:        "unicode characters in value",
			query:       "name.eq=测试",
			expectError: false,
		},
		{
			name:        "filter with dots in column name",
			query:       "user_profile.age.gt=18",
			expectError: false,
		},
		{
			name:        "multiple AND filters",
			query:       "status.eq=active&deleted_at.is=null&verified.eq=true",
			expectError: false,
		},
		{
			name:        "OR group with parentheses",
			query:       "or=(status.eq.active,status.eq.pending)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params.Filters)
			}
		})
	}
}

// TestParseOrder_EdgeCases tests order parsing edge cases
func TestParseOrder_EdgeCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "order with multiple columns",
			query:       "order=name.asc,age.desc",
			expectError: false,
		},
		{
			name:        "order with nulls first",
			query:       "order=priority.desc.nullsfirst",
			expectError: false,
		},
		{
			name:        "order with nulls last",
			query:       "order=name.asc.nullslast",
			expectError: false,
		},
		{
			name:        "order with default nulls handling",
			query:       "order=created_at.desc",
			expectError: false,
		},
		{
			name:        "order with qualified column",
			query:       "order=user.profile.name.asc",
			expectError: false,
		},
		{
			name:        "empty order value",
			query:       "order=",
			expectError: false, // Parser is lenient - empty values are ignored
		},
		{
			name:        "invalid order direction",
			query:       "order=name.invalid",
			expectError: false, // Parser is lenient - doesn't validate direction
		},
		{
			name:        "missing column",
			query:       "order=.asc",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params.Order)
			}
		})
	}
}

// TestParseAggregation_EdgeCases tests aggregation parsing edge cases
func TestParseAggregation_EdgeCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "single aggregation",
			query:       "select=count(*)",
			expectError: false,
		},
		{
			name:        "multiple aggregations",
			query:       "select=count(*),sum(price),avg(rating)",
			expectError: false,
		},
		{
			name:        "aggregation with column",
			query:       "select=sum(price)",
			expectError: false,
		},
		{
			name:        "aggregation with qualified column",
			query:       "select=avg(order_items.total)",
			expectError: false,
		},
		{
			name:        "empty select",
			query:       "select=",
			expectError: false, // Empty select is valid (means select all)
		},
		{
			name:        "invalid aggregation function - ignored",
			query:       "select=invalid_func(column)",
			expectError: false, // Parser is lenient - unknown functions are ignored
		},
		{
			name:        "count with asterisk",
			query:       "select=count(*)",
			expectError: false,
		},
		{
			name:        "max and min aggregations",
			query:       "select=max(price),min(price)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Aggregations are parsed from select, check if params has any aggregations or select fields
				// Note: invalid functions may be ignored (empty Select and Aggregations)
				if tt.query != "select=" && !strings.Contains(tt.query, "invalid") {
					// For non-empty valid selects, either Select or Aggregations should be populated
					hasSelections := len(params.Select) > 0 || len(params.Aggregations) > 0
					assert.True(t, hasSelections, "Expected either Select or Aggregations to be populated")
				}
			}
		})
	}
}

// TestParseSelect_EdgeCases tests select parsing edge cases
func TestParseSelect_EdgeCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "single column select",
			query:       "select=name",
			expectError: false,
		},
		{
			name:        "multiple column select",
			query:       "select=id,name,email",
			expectError: false,
		},
		{
			name:        "select with qualified column",
			query:       "select=user.profile.name",
			expectError: false,
		},
		{
			name:        "select with wildcard",
			query:       "select=*",
			expectError: false,
		},
		{
			name:        "empty select",
			query:       "select=",
			expectError: false, // Empty select is valid (means select all)
		},
		{
			name:        "select with special characters",
			query:       "select=user_name,first_name,last_name",
			expectError: false,
		},
		{
			name:        "select with column aliases",
			query:       "select=name,description",
			expectError: false,
		},
		{
			name:        "duplicate columns",
			query:       "select=id,name,id",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params.Select)
			}
		})
	}
}

// TestParseLogicalFilter_EdgeCases tests logical filter combinations
func TestParseLogicalFilter_EdgeCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectError bool
	}{
		{
			name:        "simple OR group",
			query:       "or=(status.eq.active,status.eq.pending)",
			expectError: false,
		},
		{
			name:        "nested OR groups",
			query:       "or=(status.eq.active,status.eq.pending,type.eq.vip)",
			expectError: false,
		},
		{
			name:        "multiple OR groups",
			query:       "or=(status.eq.active,status.eq.pending)&or=(type.eq.premium,type.eq.vip)",
			expectError: false,
		},
		{
			name:        "empty OR group",
			query:       "or=()",
			expectError: false, // Empty OR group is treated as no filter
		},
		{
			name:        "malformed logical group",
			query:       "or=(invalid",
			expectError: true,
		},
		{
			name:        "OR with single value",
			query:       "or=(status.eq.active)",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, params.Filters)
			}
		})
	}
}

// =============================================================================
// Additional Tests for Coverage Boost (Developer 3 Assignment)
// =============================================================================

// TestParseFilterOperators tests all filter operators
func TestParseFilterOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name          string
		query         string
		expectedOp    FilterOperator
		expectedValue interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:          "eq operator",
			query:         "status.eq=active",
			expectedOp:    OpEqual,
			expectedValue: "active",
			expectError:   false,
		},
		{
			name:          "neq operator",
			query:         "status.neq=deleted",
			expectedOp:    OpNotEqual,
			expectedValue: "deleted",
			expectError:   false,
		},
		{
			name:          "gt operator",
			query:         "age.gt=18",
			expectedOp:    OpGreaterThan,
			expectedValue: "18",
			expectError:   false,
		},
		{
			name:          "gte operator",
			query:         "rating.gte=4.5",
			expectedOp:    OpGreaterOrEqual,
			expectedValue: "4.5",
			expectError:   false,
		},
		{
			name:          "lt operator",
			query:         "price.lt=100",
			expectedOp:    OpLessThan,
			expectedValue: "100",
			expectError:   false,
		},
		{
			name:          "lte operator",
			query:         "quantity.lte=50",
			expectedOp:    OpLessOrEqual,
			expectedValue: "50",
			expectError:   false,
		},
		{
			name:          "like operator",
			query:         "name.like=*John*",
			expectedOp:    OpLike,
			expectedValue: "*John*",
			expectError:   false,
		},
		{
			name:          "ilike operator",
			query:         "email.ilike=*@gmail.com",
			expectedOp:    OpILike,
			expectedValue: "*@gmail.com",
			expectError:   false,
		},
		{
			name:          "is operator with null",
			query:         "deleted_at.is=null",
			expectedOp:    OpIs,
			expectedValue: nil,
			expectError:   false,
		},
		{
			name:          "is operator with true",
			query:         "verified.is=true",
			expectedOp:    OpIs,
			expectedValue: true,
			expectError:   false,
		},
		{
			name:          "is operator with false",
			query:         "active.is=false",
			expectedOp:    OpIs,
			expectedValue: false,
			expectError:   false,
		},
		{
			name:          "in operator single value",
			query:         "status.in=active",
			expectedOp:    OpIn,
			expectedValue: []string{"active"},
			expectError:   false,
		},
		{
			name:          "in operator multiple values",
			query:         "status.in=active,pending,completed",
			expectedOp:    OpIn,
			expectedValue: []string{"active", "pending", "completed"},
			expectError:   false,
		},
		{
			name:          "contains operator jsonb",
			query:         "metadata.cs={\"role\":\"admin\"}",
			expectedOp:    OpContains,
			expectedValue: `{"role":"admin"}`,
			expectError:   false,
		},
		{
			name:          "cd operator (contained)",
			query:         "tags.cd=red",
			expectedOp:    OpContained,
			expectedValue: "red",
			expectError:   false,
		},
		{
			name:          "sl operator (strictly left)",
			query:         "range.sl=10",
			expectedOp:    OpStrictlyLeft,
			expectedValue: "10",
			expectError:   false,
		},
		{
			name:          "sr operator (strictly right)",
			query:         "range.sr=20",
			expectedOp:    OpStrictlyRight,
			expectedValue: "20",
			expectError:   false,
		},
		{
			name:          "nxr operator (not extend right)",
			query:         "period.nxr=30",
			expectedOp:    OpNotExtendRight,
			expectedValue: "30",
			expectError:   false,
		},
		{
			name:          "nxl operator (not extend left)",
			query:         "period.nxl=40",
			expectedOp:    OpNotExtendLeft,
			expectedValue: "40",
			expectError:   false,
		},
		{
			name:          "adj operator (adjacent)",
			query:         "value.adj=5",
			expectedOp:    OpAdjacent,
			expectedValue: "5",
			expectError:   false,
		},
		// Note: text search operators use short form (ts, phs, wsb) not long form
		{
			name:          "text search operator (short form)",
			query:         "content.ts=search+term",
			expectedOp:    "ts",
			expectedValue: "search term",
			expectError:   false,
		},
		{
			name:          "phrase search operator (short form)",
			query:         "description.phs=exact phrase",
			expectedOp:    "phs",
			expectedValue: "exact phrase",
			expectError:   false,
		},
		{
			name:          "web search operator (short form)",
			query:         "text.wsb=search query",
			expectedOp:    "wsb",
			expectedValue: "search query",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			assert.Equal(t, tt.expectedOp, params.Filters[0].Operator)
			assert.Equal(t, tt.expectedValue, params.Filters[0].Value)
		})
	}
}

// TestParseOrder_MoreCases tests additional order parsing scenarios
func TestParseOrder_MoreCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name          string
		query         string
		expectColumns []string
		expectDesc    []bool
		expectNulls   []string
		expectError   bool
	}{
		{
			name:          "multiple order columns",
			query:         "order=name.asc,age.desc",
			expectColumns: []string{"name", "age"},
			expectDesc:    []bool{false, true},
			expectNulls:   []string{"", ""},
			expectError:   false,
		},
		{
			name:          "order with nulls first",
			query:         "order=priority.desc.nullsfirst",
			expectColumns: []string{"priority"},
			expectDesc:    []bool{true},
			expectNulls:   []string{"first"},
			expectError:   false,
		},
		{
			name:          "order with nulls last",
			query:         "order=created_at.asc.nullslast",
			expectColumns: []string{"created_at"},
			expectDesc:    []bool{false},
			expectNulls:   []string{"last"},
			expectError:   false,
		},
		{
			name:          "multiple order with nulls",
			query:         "order=name.asc.nullslast,age.desc.nullsfirst",
			expectColumns: []string{"name", "age"},
			expectDesc:    []bool{false, true},
			expectNulls:   []string{"last", "first"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, params.Order, len(tt.expectColumns))

				for i, col := range tt.expectColumns {
					assert.Equal(t, col, params.Order[i].Column)
					if tt.expectDesc != nil {
						assert.Equal(t, tt.expectDesc[i], params.Order[i].Desc)
					}
					if tt.expectNulls != nil {
						assert.Equal(t, tt.expectNulls[i], params.Order[i].Nulls)
					}
				}
			}
		})
	}
}

// TestParsePagination_MoreCases tests additional pagination scenarios
func TestParsePagination_MoreCases(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		query          string
		expectedLimit  *int
		expectedOffset *int
		expectError    bool
	}{
		{
			name:           "both limit and offset",
			config:         testConfig(),
			query:          "limit=100&offset=50",
			expectedLimit:  intPtr(100),
			expectedOffset: intPtr(50),
			expectError:    false,
		},
		{
			name:           "only limit",
			config:         testConfig(),
			query:          "limit=50",
			expectedLimit:  intPtr(50),
			expectedOffset: nil,
			expectError:    false,
		},
		{
			name:           "only offset",
			config:         testConfig(),
			query:          "offset=100",
			expectedLimit:  nil,
			expectedOffset: intPtr(100),
			expectError:    false,
		},
		{
			name:           "zero limit",
			config:         testConfig(),
			query:          "limit=0",
			expectedLimit:  intPtr(0),
			expectedOffset: nil,
			expectError:    false,
		},
		{
			name:           "zero offset",
			config:         testConfig(),
			query:          "offset=0",
			expectedLimit:  nil,
			expectedOffset: intPtr(0),
			expectError:    false,
		},
		{
			name:           "negative limit",
			config:         testConfig(),
			query:          "limit=-10",
			expectedLimit:  intPtr(-10),
			expectedOffset: nil,
			expectError:    false,
		},
		{
			name:           "negative offset",
			config:         testConfig(),
			query:          "offset=-5",
			expectedLimit:  nil,
			expectedOffset: intPtr(-5),
			expectError:    false,
		},
		{
			name:           "invalid limit - not a number",
			config:         testConfig(),
			query:          "limit=abc",
			expectedLimit:  nil,
			expectedOffset: nil,
			expectError:    true,
		},
		{
			name:           "invalid offset - not a number",
			config:         testConfig(),
			query:          "offset=xyz",
			expectedLimit:  nil,
			expectedOffset: nil,
			expectError:    true,
		},
		{
			name: "large limit with max_page_size",
			config: &config.Config{
				API: config.APIConfig{
					MaxPageSize: 1000,
				},
			},
			query:          "limit=10000",
			expectedLimit:  intPtr(1000), // Capped to MaxPageSize
			expectedOffset: nil,
			expectError:    false,
		},
		{
			name:           "empty query params",
			config:         testConfig(),
			query:          "",
			expectedLimit:  nil,
			expectedOffset: nil,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(tt.config)
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedLimit, params.Limit)
				assert.Equal(t, tt.expectedOffset, params.Offset)
			}
		})
	}
}

// TestParseSelect_MoreCases tests additional select scenarios
func TestParseSelect_MoreCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name         string
		query        string
		expectedCols []string
		expectedAggs int
		expectError  bool
	}{
		{
			name:         "single column",
			query:        "select=id",
			expectedCols: []string{"id"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "multiple columns with spaces",
			query:        "select=id, name, email",
			expectedCols: []string{"id", "name", "email"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "qualified columns",
			query:        "select=user.id,user.profile.name",
			expectedCols: []string{"user.id", "user.profile.name"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "columns with underscores",
			query:        "select=user_id,first_name,last_name",
			expectedCols: []string{"user_id", "first_name", "last_name"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "mixed columns and aggregations",
			query:        "select=category,count(*),sum(price)",
			expectedCols: []string{"category"},
			expectedAggs: 2,
			expectError:  false,
		},
		{
			name:         "only aggregations",
			query:        "select=count(*),avg(rating),min(price),max(price)",
			expectedCols: []string{},
			expectedAggs: 4,
			expectError:  false,
		},
		{
			name:         "duplicate columns",
			query:        "select=id,name,id",
			expectedCols: []string{"id", "name", "id"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "column with numbers",
			query:        "select=col1,col2,col3",
			expectedCols: []string{"col1", "col2", "col3"},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "empty select",
			query:        "select=",
			expectedCols: []string{},
			expectedAggs: 0,
			expectError:  false,
		},
		{
			name:         "wildcard",
			query:        "select=*",
			expectedCols: []string{"*"},
			expectedAggs: 0,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCols, params.Select)
				assert.Equal(t, tt.expectedAggs, len(params.Aggregations))
			}
		})
	}
}

// TestParseGroupBy_MoreCases tests additional group by scenarios
func TestParseGroupBy_MoreCases(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name            string
		query           string
		expectedColumns []string
		expectError     bool
	}{
		{
			name:            "single column",
			query:           "group_by=category",
			expectedColumns: []string{"category"},
			expectError:     false,
		},
		{
			name:            "multiple columns",
			query:           "group_by=category,status,region",
			expectedColumns: []string{"category", "status", "region"},
			expectError:     false,
		},
		{
			name:            "columns with spaces",
			query:           "group_by=category, status, region",
			expectedColumns: []string{"category", "status", "region"},
			expectError:     false,
		},
		{
			name:            "columns with underscores",
			query:           "group_by=user_id,created_date",
			expectedColumns: []string{"user_id", "created_date"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedColumns, params.GroupBy)
			}
		})
	}
}

// TestParseBuiltinRangeOperators tests range operators for built-in types
func TestParseBuiltinRangeOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name        string
		query       string
		expectedOp  FilterOperator
		expectedVal interface{}
	}{
		{
			name:        "strictly left",
			query:       "range.sl=[1,10]",
			expectedOp:  OpStrictlyLeft,
			expectedVal: "[1,10]",
		},
		{
			name:        "strictly right",
			query:       "range.sr=[20,30]",
			expectedOp:  OpStrictlyRight,
			expectedVal: "[20,30]",
		},
		{
			name:        "not extend right",
			query:       "period.nxr=[1,10]",
			expectedOp:  OpNotExtendRight,
			expectedVal: "[1,10]",
		},
		{
			name:        "not extend left",
			query:       "period.nxl=[20,30]",
			expectedOp:  OpNotExtendLeft,
			expectedVal: "[20,30]",
		},
		{
			name:        "adjacent",
			query:       "value.adj=10",
			expectedOp:  OpAdjacent,
			expectedVal: "10",
		},
		{
			name:        "overlaps",
			query:       "range.ov=[1,10]",
			expectedOp:  OpOverlaps,
			expectedVal: "[1,10]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			params, err := parser.Parse(values)
			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			assert.Equal(t, tt.expectedOp, params.Filters[0].Operator)
			assert.Equal(t, tt.expectedVal, params.Filters[0].Value)
		})
	}
}

// TestParseWithBypassMaxTotalResults tests admin bypass behavior
func TestParseWithBypassMaxTotalResults(t *testing.T) {
	config := &config.Config{
		API: config.APIConfig{
			MaxPageSize:     100,
			MaxTotalResults: 1000,
			DefaultPageSize: 50,
		},
	}

	tests := []struct {
		name          string
		query         string
		bypass        bool
		expectedLimit *int
		description   string
	}{
		{
			name:          "normal request respects max_total_results",
			query:         "offset=950&limit=100",
			bypass:        false,
			expectedLimit: intPtr(50), // Capped due to max_total_results
			description:   "Offset 950 + limit 100 exceeds max_total_results 1000",
		},
		{
			name:          "bypass ignores max_total_results",
			query:         "offset=950&limit=100",
			bypass:        true,
			expectedLimit: intPtr(100), // Not capped
			description:   "Admin bypass allows exceeding max_total_results",
		},
		{
			name:          "bypass still respects max_page_size",
			query:         "limit=10000",
			bypass:        true,
			expectedLimit: intPtr(100), // Still capped to max_page_size
			description:   "max_page_size is always enforced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewQueryParser(config)
			values, err := url.ParseQuery(tt.query)
			require.NoError(t, err)

			opts := ParseOptions{BypassMaxTotalResults: tt.bypass}
			params, err := parser.ParseWithOptions(values, opts)
			require.NoError(t, err, tt.description)

			require.NotNil(t, params.Limit)
			assert.Equal(t, *tt.expectedLimit, *params.Limit, tt.description)
		})
	}
}
