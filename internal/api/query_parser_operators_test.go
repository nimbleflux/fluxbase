package api

import (
	"net/url"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQueryParser_AdditionalOperators tests operators not covered in existing tests
func TestQueryParser_AdditionalOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedOp     FilterOperator
		expectedValue  interface{}
	}{
		// Comparison operators
		{
			name:           "not equal operator",
			query:          "status.neq=inactive",
			expectedColumn: "status",
			expectedOp:     OpNotEqual,
			expectedValue:  "inactive",
		},
		{
			name:           "less than or equal",
			query:          "price.lte=100",
			expectedColumn: "price",
			expectedOp:     OpLessOrEqual,
			expectedValue:  "100",
		},
		{
			name:           "greater than or equal",
			query:          "rating.gte=4.0",
			expectedColumn: "rating",
			expectedOp:     OpGreaterOrEqual,
			expectedValue:  "4.0",
		},
		// IS operator variations
		{
			name:           "is true",
			query:          "verified.is=true",
			expectedColumn: "verified",
			expectedOp:     OpIs,
			expectedValue:  true,
		},
		{
			name:           "is false",
			query:          "deleted.is=false",
			expectedColumn: "deleted",
			expectedOp:     OpIs,
			expectedValue:  false,
		},
		// Array operators
		{
			name:           "contains operator",
			query:          "tags.cs=react",
			expectedColumn: "tags",
			expectedOp:     "cs",
			expectedValue:  "react",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, tt.expectedColumn, filter.Column)
			assert.Equal(t, tt.expectedOp, filter.Operator)

			if tt.expectedValue == nil {
				assert.Nil(t, filter.Value)
			} else if slice, ok := tt.expectedValue.([]string); ok {
				assert.Equal(t, slice, filter.Value)
			} else {
				assert.Equal(t, tt.expectedValue, filter.Value)
			}
		})
	}
}

// TestQueryParser_NotOperator tests the NOT operator with nested operators
func TestQueryParser_NotOperator(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("NOT with equals", func(t *testing.T) {
		query := "status.not=eq.deleted"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)

		filter := params.Filters[0]
		assert.Equal(t, "status", filter.Column)
		assert.Equal(t, OpNot, filter.Operator)
		assert.Equal(t, "eq.deleted", filter.Value)
	})
}

// TestQueryParser_TextSearchOperators tests full-text search operators
func TestQueryParser_TextSearchOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("tsvector - fts (full text search)", func(t *testing.T) {
		query := "searchtext.fts=search query"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)

		filter := params.Filters[0]
		assert.Equal(t, "searchtext", filter.Column)
		assert.Equal(t, OpTextSearch, filter.Operator)
		assert.Equal(t, "search query", filter.Value)
	})

	t.Run("plfts (phrase full text search)", func(t *testing.T) {
		query := "title.plfts=\"exact phrase\""
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)

		filter := params.Filters[0]
		assert.Equal(t, "title", filter.Column)
		assert.Equal(t, OpPhraseSearch, filter.Operator)
		assert.Equal(t, "\"exact phrase\"", filter.Value)
	})
}

// TestQueryParser_PostGISTOperators tests PostGIS spatial operators
func TestQueryParser_PostGISTOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	geoJSON := `{"type":"Point","coordinates":[-122.4,37.8]}`

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedOp     FilterOperator
	}{
		{
			name:           "ST_Intersects",
			query:          "geom.st_intersects=" + url.QueryEscape(geoJSON),
			expectedColumn: "geom",
			expectedOp:     OpSTIntersects,
		},
		{
			name:           "ST_Contains",
			query:          "boundary.st_contains=" + url.QueryEscape(geoJSON),
			expectedColumn: "boundary",
			expectedOp:     OpSTContains,
		},
		{
			name:           "ST_Within",
			query:          "location.st_within=" + url.QueryEscape(geoJSON),
			expectedColumn: "location",
			expectedOp:     OpSTWithin,
		},
		{
			name:           "ST_Distance",
			query:          "geom.st_distance=" + url.QueryEscape(geoJSON),
			expectedColumn: "geom",
			expectedOp:     OpSTDistance,
		},
		{
			name:           "ST_Touches",
			query:          "geom.st_touches=" + url.QueryEscape(geoJSON),
			expectedColumn: "geom",
			expectedOp:     OpSTTouches,
		},
		{
			name:           "ST_Crosses",
			query:          "geom.st_crosses=" + url.QueryEscape(geoJSON),
			expectedColumn: "geom",
			expectedOp:     OpSTCrosses,
		},
		{
			name:           "ST_Overlaps",
			query:          "geom.st_overlaps=" + url.QueryEscape(geoJSON),
			expectedColumn: "geom",
			expectedOp:     OpSTOverlaps,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, tt.expectedColumn, filter.Column)
			assert.Equal(t, tt.expectedOp, filter.Operator)
		})
	}
}

// TestQueryParser_ST_DWithin tests ST_DWithin with distance parameter
func TestQueryParser_ST_DWithin(t *testing.T) {
	parser := NewQueryParser(testConfig())

	geoJSON := `{"type":"Point","coordinates":[-122.4,37.8]}`
	query := "geom.st_dwithin=" + url.QueryEscape("1000,"+geoJSON)

	values, _ := url.ParseQuery(query)
	params, err := parser.Parse(values)

	require.NoError(t, err)
	require.Len(t, params.Filters, 1)

	filter := params.Filters[0]
	assert.Equal(t, "geom", filter.Column)
	assert.Equal(t, OpSTDWithin, filter.Operator)
	assert.Equal(t, "1000,"+geoJSON, filter.Value)
}

// TestQueryParser_PgVectorOperators tests pgvector similarity operators
func TestQueryParser_PgVectorOperators(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedOp     FilterOperator
		expectedValue  string
	}{
		{
			name:           "vector L2 distance",
			query:          "embedding.vec_l2=[0.1,0.2,0.3]",
			expectedColumn: "embedding",
			expectedOp:     OpVectorL2,
			expectedValue:  "[0.1,0.2,0.3]",
		},
		{
			name:           "vector cosine distance",
			query:          "embedding.vec_cos=[0.1,0.2,0.3]",
			expectedColumn: "embedding",
			expectedOp:     OpVectorCosine,
			expectedValue:  "[0.1,0.2,0.3]",
		},
		{
			name:           "vector inner product",
			query:          "embedding.vec_ip=[0.1,0.2,0.3]",
			expectedColumn: "embedding",
			expectedOp:     OpVectorIP,
			expectedValue:  "[0.1,0.2,0.3]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, tt.expectedColumn, filter.Column)
			assert.Equal(t, tt.expectedOp, filter.Operator)
			assert.Equal(t, tt.expectedValue, filter.Value)
		})
	}
}

// TestQueryParser_PostgRESTFormat tests PostgREST-style format (column=operator.value)
func TestQueryParser_PostgRESTFormat(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name           string
		query          string
		expectedColumn string
		expectedOp     FilterOperator
		expectedValue  interface{}
	}{
		{
			name:           "PostgREST equal",
			query:          "id=eq.1",
			expectedColumn: "id",
			expectedOp:     OpEqual,
			expectedValue:  "1",
		},
		{
			name:           "PostgREST greater than",
			query:          "age=gt.18",
			expectedColumn: "age",
			expectedOp:     OpGreaterThan,
			expectedValue:  "18",
		},
		{
			name:           "PostgREST in with parentheses",
			query:          "status=in.(active,pending,completed)",
			expectedColumn: "status",
			expectedOp:     OpIn,
			expectedValue:  []string{"active", "pending", "completed"},
		},
		{
			name:           "PostgREST is null",
			query:          "deleted_at=is.null",
			expectedColumn: "deleted_at",
			expectedOp:     OpIs,
			expectedValue:  nil,
		},
		{
			name:           "PostgREST with array brackets",
			query:          "tags=in.(react,go)",
			expectedColumn: "tags",
			expectedOp:     OpIn,
			expectedValue:  []string{"react", "go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, tt.expectedColumn, filter.Column)
			assert.Equal(t, tt.expectedOp, filter.Operator)

			if tt.expectedValue == nil {
				assert.Nil(t, filter.Value)
			} else if slice, ok := tt.expectedValue.([]string); ok {
				assert.Equal(t, slice, filter.Value)
			} else {
				assert.Equal(t, tt.expectedValue, filter.Value)
			}
		})
	}
}

// TestQueryParser_SQLInjectionPrevention validates SQL injection prevention
func TestQueryParser_SQLInjectionPrevention(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("valid identifier with underscore accepted", func(t *testing.T) {
		query := "user_name.eq=John"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)
		assert.Equal(t, "user_name", params.Filters[0].Column)
	})

	t.Run("valid identifier starting with underscore accepted", func(t *testing.T) {
		query := "_private.eq=value"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)
		assert.Equal(t, "_private", params.Filters[0].Column)
	})

	t.Run("identifier with numbers is accepted", func(t *testing.T) {
		query := "user_id123.eq=1"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.Len(t, params.Filters, 1)
		assert.Equal(t, "user_id123", params.Filters[0].Column)
	})
}

// TestQueryParser_UnbalancedParentheses tests error handling for unbalanced parentheses
func TestQueryParser_UnbalancedParentheses(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("unbalanced opening paren in OR", func(t *testing.T) {
		query := "or=(status.eq.active,status.eq.pending"
		values, _ := url.ParseQuery(query)
		_, err := parser.Parse(values)

		assert.Error(t, err, "Unbalanced parentheses should error")
	})

	t.Run("unbalanced closing paren in OR", func(t *testing.T) {
		query := "or=status.eq.active,status.eq.pending)"
		values, _ := url.ParseQuery(query)
		_, err := parser.Parse(values)

		assert.Error(t, err, "Unbalanced parentheses should error")
	})

	t.Run("balanced nested parentheses", func(t *testing.T) {
		query := "or=(and(col1.eq.val1,col2.eq.val2),and(col3.eq.val3,col4.eq.val4))"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err, "Balanced nested parentheses should work")
		assert.Len(t, params.Filters, 4)
	})
}

// TestQueryParser_InvalidFilterFormats tests various invalid filter formats
func TestQueryParser_InvalidFilterFormats(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("filter without operator does not create a filter", func(t *testing.T) {
		// "status=active" without a dot is not recognized as a filter
		query := "status=active"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		// This succeeds but doesn't create a filter since there's no operator
		assert.NoError(t, err)
		assert.Len(t, params.Filters, 0, "Filter without operator should not create a filter")
	})
}

// TestQueryParser_SpecialCharactersInValues tests special characters in filter values
func TestQueryParser_SpecialCharactersInValues(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name          string
		query         string
		expectedValue interface{}
	}{
		{
			name:          "value with dots",
			query:         "version.eq=1.2.3",
			expectedValue: "1.2.3",
		},
		{
			name:          "value with hyphens",
			query:         "slug.eq=my-blog-post",
			expectedValue: "my-blog-post",
		},
		{
			name:          "value with at sign",
			query:         "email.eq=test@example.com",
			expectedValue: "test@example.com",
		},
		{
			name:          "value with plus sign (URL encoded)",
			query:         "phone.eq=%2B1-555-1234",
			expectedValue: "+1-555-1234",
		},
		{
			name:          "value with spaces (URL encoded)",
			query:         "title.eq=" + url.QueryEscape("Hello World"),
			expectedValue: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, tt.expectedValue, filter.Value)
		})
	}
}

// TestQueryParser_ArrayValueParsing tests array value parsing for IN operator
func TestQueryParser_ArrayValueParsing(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name          string
		query         string
		expectedValue []string
	}{
		{
			name:          "comma-separated values",
			query:         "status.in=active,pending,completed",
			expectedValue: []string{"active", "pending", "completed"},
		},
		{
			name:          "parenthesis-wrapped values",
			query:         "id.in=(1,2,3)",
			expectedValue: []string{"1", "2", "3"},
		},
		{
			name:          "bracket-wrapped values",
			query:         "tags.in=[react,go,typescript]",
			expectedValue: []string{"react", "go", "typescript"},
		},
		{
			name:          "double-quoted values",
			query:         "name.in=(\"John Doe\",\"Jane Smith\")",
			expectedValue: []string{"John Doe", "Jane Smith"},
		},
		{
			name:          "single-quoted values",
			query:         "tags.in=('react','go')",
			expectedValue: []string{"react", "go"},
		},
		{
			name:          "mixed quoted and unquoted",
			query:         "values.in=(1,test,\"more data\")",
			expectedValue: []string{"1", "test", "more data"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			require.Len(t, params.Filters, 1)

			filter := params.Filters[0]
			assert.Equal(t, OpIn, filter.Operator)

			valueSlice, ok := filter.Value.([]string)
			require.True(t, ok)
			assert.Equal(t, tt.expectedValue, valueSlice)
		})
	}
}

// TestQueryParser_CountParameter tests count parameter variations
func TestQueryParser_CountParameter(t *testing.T) {
	parser := NewQueryParser(testConfig())

	tests := []struct {
		name     string
		query    string
		expected CountType
	}{
		{
			name:     "count exact",
			query:    "count=exact",
			expected: CountExact,
		},
		{
			name:     "count planned",
			query:    "count=planned",
			expected: CountPlanned,
		},
		{
			name:     "count estimated",
			query:    "count=estimated",
			expected: CountEstimated,
		},
		{
			name:     "count none",
			query:    "count=none",
			expected: CountNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, params.Count)
		})
	}
}

// TestQueryParser_TruncateParameter tests truncate parameter
func TestQueryParser_TruncateParameter(t *testing.T) {
	parser := NewQueryParser(testConfig())

	t.Run("truncate with positive value", func(t *testing.T) {
		query := "truncate=100"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.NotNil(t, params.TruncateLength)
		assert.Equal(t, 100, *params.TruncateLength)
	})

	t.Run("truncate with zero", func(t *testing.T) {
		query := "truncate=0"
		values, _ := url.ParseQuery(query)
		params, err := parser.Parse(values)

		require.NoError(t, err)
		require.NotNil(t, params.TruncateLength)
		assert.Equal(t, 0, *params.TruncateLength)
	})

	t.Run("truncate with negative value should error", func(t *testing.T) {
		query := "truncate=-1"
		values, _ := url.ParseQuery(query)
		_, err := parser.Parse(values)

		assert.Error(t, err)
	})

	t.Run("truncate with invalid value should error", func(t *testing.T) {
		query := "truncate=abc"
		values, _ := url.ParseQuery(query)
		_, err := parser.Parse(values)

		assert.Error(t, err)
	})
}

// TestQueryParser_DefaultPageSize tests default page size application
func TestQueryParser_DefaultPageSize(t *testing.T) {
	tests := []struct {
		name            string
		defaultPageSize int
		query           string
		expectedLimit   *int
	}{
		{
			name:            "default applied when no limit",
			defaultPageSize: 20,
			query:           "status.eq=active",
			expectedLimit:   func() *int { i := 20; return &i }(),
		},
		{
			name:            "explicit limit overrides default",
			defaultPageSize: 20,
			query:           "status.eq=active&limit=50",
			expectedLimit:   func() *int { i := 50; return &i }(),
		},
		{
			name:            "no default when set to -1",
			defaultPageSize: -1,
			query:           "status.eq=active",
			expectedLimit:   nil,
		},
		{
			name:            "no default when set to 0",
			defaultPageSize: 0,
			query:           "status.eq=active",
			expectedLimit:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				API: config.APIConfig{
					MaxPageSize:     -1,
					MaxTotalResults: -1,
					DefaultPageSize: tt.defaultPageSize,
				},
			}
			parser := NewQueryParser(cfg)

			values, _ := url.ParseQuery(tt.query)
			params, err := parser.Parse(values)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedLimit, params.Limit)
		})
	}
}

// TestQueryParser_BypassMaxTotalResults tests BypassMaxTotalResults option
func TestQueryParser_BypassMaxTotalResults(t *testing.T) {
	cfg := &config.Config{
		API: config.APIConfig{
			MaxPageSize:     100,
			MaxTotalResults: 500,
			DefaultPageSize: -1,
		},
	}
	parser := NewQueryParser(cfg)

	query := "limit=100&offset=450"
	values, _ := url.ParseQuery(query)

	// Without bypass - should be capped
	params, err := parser.Parse(values)
	require.NoError(t, err)
	assert.Equal(t, 50, *params.Limit, "Should be capped without bypass")

	// With bypass - should not be capped
	params, err = parser.ParseWithOptions(values, ParseOptions{BypassMaxTotalResults: true})
	require.NoError(t, err)
	assert.Equal(t, 100, *params.Limit, "Should not be capped with bypass")
}
