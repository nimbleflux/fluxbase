package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/database"
)

func TestIsGeoJSON_GeoTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name: "valid Point",
			input: map[string]interface{}{
				"type":        "Point",
				"coordinates": []interface{}{-122.4, 37.8},
			},
			expected: true,
		},
		{
			name: "valid LineString",
			input: map[string]interface{}{
				"type":        "LineString",
				"coordinates": []interface{}{[]interface{}{0, 0}, []interface{}{1, 1}},
			},
			expected: true,
		},
		{
			name: "valid Polygon",
			input: map[string]interface{}{
				"type":        "Polygon",
				"coordinates": []interface{}{[]interface{}{[]interface{}{0, 0}, []interface{}{1, 0}, []interface{}{1, 1}, []interface{}{0, 0}}},
			},
			expected: true,
		},
		{
			name: "valid MultiPoint",
			input: map[string]interface{}{
				"type":        "MultiPoint",
				"coordinates": []interface{}{[]interface{}{0, 0}, []interface{}{1, 1}},
			},
			expected: true,
		},
		{
			name: "valid MultiLineString",
			input: map[string]interface{}{
				"type":        "MultiLineString",
				"coordinates": []interface{}{},
			},
			expected: true,
		},
		{
			name: "valid MultiPolygon",
			input: map[string]interface{}{
				"type":        "MultiPolygon",
				"coordinates": []interface{}{},
			},
			expected: true,
		},
		{
			name: "valid GeometryCollection",
			input: map[string]interface{}{
				"type":        "GeometryCollection",
				"coordinates": []interface{}{},
			},
			expected: true,
		},
		{
			name:     "not a map",
			input:    "not a map",
			expected: false,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name: "missing type",
			input: map[string]interface{}{
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
		{
			name: "missing coordinates",
			input: map[string]interface{}{
				"type": "Point",
			},
			expected: false,
		},
		{
			name: "invalid type",
			input: map[string]interface{}{
				"type":        "InvalidType",
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
		{
			name: "type is not a string",
			input: map[string]interface{}{
				"type":        123,
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: false,
		},
		{
			name: "Feature type (not a geometry)",
			input: map[string]interface{}{
				"type":        "Feature",
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGeoJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPartialGeoJSON_Validation(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name: "has type but no coordinates - partial",
			input: map[string]interface{}{
				"type": "Point",
			},
			expected: true,
		},
		{
			name: "has type and coordinates - complete",
			input: map[string]interface{}{
				"type":        "Point",
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
		{
			name: "has coordinates but no type",
			input: map[string]interface{}{
				"coordinates": []interface{}{0, 0},
			},
			expected: false,
		},
		{
			name:     "not a map",
			input:    "string",
			expected: false,
		},
		{
			name:     "nil input",
			input:    nil,
			expected: false,
		},
		{
			name:     "empty map",
			input:    map[string]interface{}{},
			expected: false,
		},
		{
			name: "has type and other fields but no coordinates",
			input: map[string]interface{}{
				"type":       "Point",
				"properties": map[string]interface{}{"name": "test"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPartialGeoJSON(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsGeometryColumn_DataTypes(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected bool
	}{
		{"geometry lowercase", "geometry", true},
		{"geometry uppercase", "GEOMETRY", true},
		{"geometry mixed case", "Geometry", true},
		{"geometry with srid", "geometry(Point,4326)", true},
		{"geography", "geography", true},
		{"geography with srid", "geography(Point,4326)", true},
		{"text column", "text", false},
		{"varchar column", "varchar", false},
		{"integer column", "integer", false},
		{"json column", "json", false},
		{"jsonb column", "jsonb", false},
		{"empty string", "", false},
		{"similar but not geometry", "geometrical", false}, // does not actually contain "geometry"
		{"point without geometry prefix", "Point", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGeometryColumn(tt.dataType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsTextColumn(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		expected bool
	}{
		{"text", "text", true},
		{"TEXT uppercase", "TEXT", true},
		{"varchar", "varchar", true},
		{"varchar with length", "varchar(255)", true},
		{"character varying", "character varying", true},
		{"character varying with length", "character varying(100)", true},
		{"char", "char", true},
		{"char with length", "char(10)", true},
		{"character", "character", true},
		{"character with length", "character(50)", true},
		{"integer", "integer", false},
		{"json", "json", false},
		{"jsonb", "jsonb", false},
		{"bytea", "bytea", false},
		{"boolean", "boolean", false},
		{"timestamp", "timestamp", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTextColumn(tt.dataType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildSelectColumns_GeoJSON(t *testing.T) {
	t.Run("basic columns", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "name", DataType: "text"},
				{Name: "email", DataType: "varchar(255)"},
			},
		}
		result := buildSelectColumns(table)
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `"name"`)
		assert.Contains(t, result, `"email"`)
	})

	t.Run("with geometry column", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "locations",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "location", DataType: "geometry(Point,4326)"},
			},
		}
		result := buildSelectColumns(table)
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `ST_AsGeoJSON("location")::jsonb AS "location"`)
	})

	t.Run("with geography column", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "areas",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "area", DataType: "geography"},
			},
		}
		result := buildSelectColumns(table)
		assert.Contains(t, result, `ST_AsGeoJSON("area")::jsonb AS "area"`)
	})

	t.Run("empty columns", func(t *testing.T) {
		table := database.TableInfo{
			Schema:  "public",
			Name:    "empty",
			Columns: []database.ColumnInfo{},
		}
		result := buildSelectColumns(table)
		assert.Equal(t, "", result)
	})
}

func TestBuildSelectColumnsWithTruncation(t *testing.T) {
	t.Run("no truncation when nil", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "posts",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "content", DataType: "text"},
			},
		}
		result := buildSelectColumnsWithTruncation(table, nil)
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `"content"`)
		assert.NotContains(t, result, "LEFT")
	})

	t.Run("no truncation when zero", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "posts",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "content", DataType: "text"},
			},
		}
		truncLen := 0
		result := buildSelectColumnsWithTruncation(table, &truncLen)
		assert.NotContains(t, result, "LEFT")
	})

	t.Run("truncates text columns", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "posts",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "content", DataType: "text"},
				{Name: "title", DataType: "varchar(255)"},
			},
		}
		truncLen := 100
		result := buildSelectColumnsWithTruncation(table, &truncLen)
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, "LEFT")
		assert.Contains(t, result, "100")
	})

	t.Run("does not truncate non-text columns", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "data",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "data", DataType: "jsonb"},
			},
		}
		truncLen := 100
		result := buildSelectColumnsWithTruncation(table, &truncLen)
		// Should only have simple column references for non-text
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `"data"`)
	})

	t.Run("handles geometry and truncation together", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "places",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "name", DataType: "text"},
				{Name: "location", DataType: "geometry(Point,4326)"},
			},
		}
		truncLen := 50
		result := buildSelectColumnsWithTruncation(table, &truncLen)
		assert.Contains(t, result, `ST_AsGeoJSON("location")::jsonb AS "location"`)
		assert.Contains(t, result, "LEFT")
		assert.Contains(t, result, "50")
	})
}

func TestBuildReturningClause_GeoJSON(t *testing.T) {
	t.Run("basic table", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "users",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "name", DataType: "text"},
			},
		}
		result := buildReturningClause(table)
		assert.True(t, len(result) > 0)
		assert.Contains(t, result, "RETURNING")
		assert.Contains(t, result, `"id"`)
		assert.Contains(t, result, `"name"`)
	})

	t.Run("with geometry converts to GeoJSON", func(t *testing.T) {
		table := database.TableInfo{
			Schema: "public",
			Name:   "locations",
			Columns: []database.ColumnInfo{
				{Name: "id", DataType: "integer"},
				{Name: "geom", DataType: "geometry"},
			},
		}
		result := buildReturningClause(table)
		assert.Contains(t, result, "RETURNING")
		assert.Contains(t, result, "ST_AsGeoJSON")
	})
}
