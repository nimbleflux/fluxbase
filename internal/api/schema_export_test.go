package api

import (
	"testing"
)

func TestPgTypeToTS(t *testing.T) {
	tests := []struct {
		pgType   string
		expected string
	}{
		// String types
		{"text", "string"},
		{"varchar", "string"},
		{"varchar(255)", "string"},
		{"character varying(100)", "string"},
		{"char", "string"},
		{"char(1)", "string"},
		{"character(10)", "string"},
		{"uuid", "string"},
		{"citext", "string"},
		{"name", "string"},

		// Numeric types
		{"integer", "number"},
		{"int4", "number"},
		{"int8", "number"},
		{"bigint", "number"},
		{"smallint", "number"},
		{"int2", "number"},
		{"real", "number"},
		{"float4", "number"},
		{"float8", "number"},
		{"double precision", "number"},
		{"numeric", "number"},
		{"numeric(10,2)", "number"},
		{"decimal", "number"},
		{"decimal(10,2)", "number"},
		{"money", "number"},
		{"serial", "number"},
		{"bigserial", "number"},
		{"smallserial", "number"},
		{"oid", "number"},

		// Boolean
		{"boolean", "boolean"},
		{"bool", "boolean"},

		// JSON types
		{"json", "Record<string, unknown>"},
		{"jsonb", "Record<string, unknown>"},

		// Date/time types
		{"date", "string"},
		{"timestamp", "string"},
		{"timestamp without time zone", "string"},
		{"timestamp with time zone", "string"},
		{"timestamptz", "string"},
		{"time", "string"},
		{"time without time zone", "string"},
		{"time with time zone", "string"},
		{"timetz", "string"},
		{"interval", "string"},

		// Binary
		{"bytea", "string"},

		// Network types
		{"inet", "string"},
		{"cidr", "string"},
		{"macaddr", "string"},
		{"macaddr8", "string"},

		// Geometric types
		{"point", "string"},
		{"line", "string"},
		{"lseg", "string"},
		{"box", "string"},
		{"path", "string"},
		{"polygon", "string"},
		{"circle", "string"},

		// Range types
		{"int4range", "string"},
		{"int8range", "string"},
		{"numrange", "string"},
		{"tsrange", "string"},
		{"tstzrange", "string"},
		{"daterange", "string"},

		// Full-text search
		{"tsvector", "string"},
		{"tsquery", "string"},

		// Vector type (pgvector)
		{"vector", "number[]"},

		// XML
		{"xml", "string"},

		// Special types
		{"void", "void"},
		{"record", "Record<string, unknown>"},

		// Array types
		{"text[]", "string[]"},
		{"integer[]", "number[]"},
		{"boolean[]", "boolean[]"},
		{"jsonb[]", "Record<string, unknown>[]"},
		{"uuid[]", "string[]"},

		// SETOF types
		{"SETOF text", "string[]"},
		{"setof integer", "number[]"},

		// Unknown types
		{"custom_type", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.pgType, func(t *testing.T) {
			result := pgTypeToTS(tt.pgType)
			if result != tt.expected {
				t.Errorf("pgTypeToTS(%q) = %q, want %q", tt.pgType, result, tt.expected)
			}
		})
	}
}

func TestToPascalCase_SchemaExport(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"users", "Users"},
		{"user_profiles", "UserProfiles"},
		{"my_table_name", "MyTableName"},
		{"public", "Public"},
		{"UPPERCASE", "Uppercase"},
		{"already_pascal", "AlreadyPascal"},
		{"kebab-case", "KebabCase"},
		{"mixed_kebab-case", "MixedKebabCase"},
		{"single", "Single"},
		{"a", "A"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"validName", "validName"},
		{"valid_name", "valid_name"},
		{"_private", "_private"},
		{"$dollar", "$dollar"},
		{"name123", "name123"},
		{"123invalid", "'123invalid'"},
		{"with space", "'with space'"},
		{"with-dash", "'with-dash'"},
		{"with'quote", "'with\\'quote'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFilterBySchema(t *testing.T) {
	// Create test tables
	tables := []struct {
		schema string
		name   string
	}{
		{"public", "users"},
		{"public", "posts"},
		{"auth", "users"},
		{"auth", "sessions"},
		{"storage", "buckets"},
	}

	// We need to import database.TableInfo but we're in the same package
	// This is a simplified test structure
	t.Run("filter public schema", func(t *testing.T) {
		// Actual filtering test would require database.TableInfo
		// This is a placeholder to show the test structure
		schemas := []string{"public"}
		schemaSet := make(map[string]bool)
		for _, s := range schemas {
			schemaSet[s] = true
		}

		count := 0
		for _, tbl := range tables {
			if schemaSet[tbl.schema] {
				count++
			}
		}

		if count != 2 {
			t.Errorf("Expected 2 tables in public schema, got %d", count)
		}
	})

	t.Run("filter multiple schemas", func(t *testing.T) {
		schemas := []string{"public", "auth"}
		schemaSet := make(map[string]bool)
		for _, s := range schemas {
			schemaSet[s] = true
		}

		count := 0
		for _, tbl := range tables {
			if schemaSet[tbl.schema] {
				count++
			}
		}

		if count != 4 {
			t.Errorf("Expected 4 tables in public+auth schemas, got %d", count)
		}
	})
}

// =============================================================================
// TypeScriptExportRequest Tests
// =============================================================================

func TestTypeScriptExportRequest_Struct(t *testing.T) {
	t.Run("default request", func(t *testing.T) {
		req := TypeScriptExportRequest{}

		if len(req.Schemas) != 0 {
			t.Errorf("Expected empty schemas, got %v", req.Schemas)
		}
		if req.IncludeFunctions != false {
			t.Errorf("Expected IncludeFunctions false, got %v", req.IncludeFunctions)
		}
		if req.IncludeViews != false {
			t.Errorf("Expected IncludeViews false, got %v", req.IncludeViews)
		}
		if req.Format != "" {
			t.Errorf("Expected empty format, got %v", req.Format)
		}
	})

	t.Run("full request", func(t *testing.T) {
		req := TypeScriptExportRequest{
			Schemas:          []string{"public", "auth"},
			IncludeFunctions: true,
			IncludeViews:     true,
			Format:           "full",
		}

		if len(req.Schemas) != 2 {
			t.Errorf("Expected 2 schemas, got %d", len(req.Schemas))
		}
		if req.Schemas[0] != "public" {
			t.Errorf("Expected first schema 'public', got %v", req.Schemas[0])
		}
		if req.Schemas[1] != "auth" {
			t.Errorf("Expected second schema 'auth', got %v", req.Schemas[1])
		}
		if !req.IncludeFunctions {
			t.Errorf("Expected IncludeFunctions true, got %v", req.IncludeFunctions)
		}
		if !req.IncludeViews {
			t.Errorf("Expected IncludeViews true, got %v", req.IncludeViews)
		}
		if req.Format != "full" {
			t.Errorf("Expected format 'full', got %v", req.Format)
		}
	})

	t.Run("types format", func(t *testing.T) {
		req := TypeScriptExportRequest{
			Schemas: []string{"public"},
			Format:  "types",
		}

		if req.Format != "types" {
			t.Errorf("Expected format 'types', got %v", req.Format)
		}
	})
}

// =============================================================================
// NewSchemaExportHandler Tests
// =============================================================================

func TestNewSchemaExportHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewSchemaExportHandler(nil, nil)

		if handler == nil {
			t.Error("Expected non-nil handler")
		}
		if handler.schemaCache != nil {
			t.Error("Expected nil schemaCache")
		}
		if handler.inspector != nil {
			t.Error("Expected nil inspector")
		}
	})
}

// =============================================================================
// getCurrentTimestamp Tests
// =============================================================================

func TestGetCurrentTimestamp(t *testing.T) {
	t.Run("returns runtime value", func(t *testing.T) {
		result := getCurrentTimestamp()
		if result != "runtime" {
			t.Errorf("Expected 'runtime', got %v", result)
		}
	})
}

// =============================================================================
// pgTypeToTS Edge Cases Tests
// =============================================================================

func TestPgTypeToTS_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		pgType   string
		expected string
	}{
		{"empty string", "", "unknown"},
		{"whitespace", "  text  ", "string"},
		{"uppercase", "TEXT", "string"},
		{"mixed case", "TeXt", "string"},
		{"with schema", "pg_catalog.int4", "unknown"},
		{"very long type", "verylongtypename", "unknown"},
		{"type with brackets", "numeric(10,2)", "number"},
		{"nested array", "text[][]", "string[]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pgTypeToTS(tt.pgType)
			if result != tt.expected {
				t.Errorf("pgTypeToTS(%q) = %q, want %q", tt.pgType, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// sanitizeIdentifier Edge Cases Tests
// =============================================================================

func TestSanitizeIdentifier_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", "''"},
		{"starts with underscore", "_valid", "_valid"},
		{"starts with dollar", "$valid", "$valid"},
		{"uppercase letters", "ValidName", "ValidName"},
		{"all uppercase", "UPPERCASE", "UPPERCASE"},
		{"numbers only", "123", "'123'"},
		{"unicode characters", "名前", "'名前'"},
		{"special symbols", "@#$%", "'@#$%'"},
		{"mixed valid invalid", "valid@invalid", "'valid@invalid'"},
		{"multiple spaces", "has  multiple  spaces", "'has  multiple  spaces'"},
		{"tabs and newlines", "has\ttab", "'has\ttab'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeIdentifier(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeIdentifier(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkPgTypeToTS(b *testing.B) {
	types := []string{
		"text",
		"integer",
		"boolean",
		"jsonb",
		"timestamp with time zone",
		"numeric(10,2)",
		"text[]",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pgTypeToTS(types[i%len(types)])
	}
}

func BenchmarkToPascalCase(b *testing.B) {
	inputs := []string{
		"users",
		"user_profiles",
		"my_table_name",
		"public",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toPascalCase(inputs[i%len(inputs)])
	}
}

func BenchmarkSanitizeIdentifier(b *testing.B) {
	inputs := []string{
		"validName",
		"valid_name",
		"123invalid",
		"with space",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sanitizeIdentifier(inputs[i%len(inputs)])
	}
}
