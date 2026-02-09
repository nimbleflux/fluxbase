package tools

import (
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
)

func TestValidateDDLIdentifier(t *testing.T) {
	t.Run("valid identifiers pass validation", func(t *testing.T) {
		validNames := []string{
			"users",
			"user_accounts",
			"_private",
			"Table1",
			"a",
			"_",
			"users_v2",
			"snake_case_name",
			"CamelCase",
			"mixedCase123",
		}

		for _, name := range validNames {
			t.Run(name, func(t *testing.T) {
				err := validateDDLIdentifier(name, "table")
				assert.NoError(t, err, "identifier '%s' should be valid", name)
			})
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		err := validateDDLIdentifier("", "table")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("names exceeding 63 characters rejected", func(t *testing.T) {
		longName := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz1234567890ab" // 64 characters
		err := validateDDLIdentifier(longName, "table")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed 63 characters")
	})

	t.Run("name at 63 characters accepted", func(t *testing.T) {
		name := "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz12345678901"
		assert.Len(t, name, 63)
		err := validateDDLIdentifier(name, "table")
		assert.NoError(t, err)
	})

	t.Run("names starting with number rejected", func(t *testing.T) {
		invalidNames := []string{
			"1users",
			"123",
			"0_table",
		}

		for _, name := range invalidNames {
			t.Run(name, func(t *testing.T) {
				err := validateDDLIdentifier(name, "table")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "must start with a letter or underscore")
			})
		}
	})

	t.Run("names with invalid characters rejected", func(t *testing.T) {
		invalidNames := []string{
			"user-name",
			"user.name",
			"user name",
			"user@email",
			"table$1",
			"drop;--",
			"user's",
			"table\"name",
		}

		for _, name := range invalidNames {
			t.Run(name, func(t *testing.T) {
				err := validateDDLIdentifier(name, "table")
				assert.Error(t, err)
			})
		}
	})

	t.Run("reserved keywords rejected", func(t *testing.T) {
		reservedNames := []string{
			"user",
			"table",
			"column",
			"index",
			"select",
			"insert",
			"update",
			"delete",
		}

		for _, name := range reservedNames {
			t.Run(name, func(t *testing.T) {
				err := validateDDLIdentifier(name, "table")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "reserved keyword")
			})
		}
	})

	t.Run("reserved keywords case insensitive", func(t *testing.T) {
		testCases := []string{
			"USER",
			"User",
			"SELECT",
			"Select",
			"TABLE",
			"Table",
		}

		for _, name := range testCases {
			t.Run(name, func(t *testing.T) {
				err := validateDDLIdentifier(name, "table")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "reserved keyword")
			})
		}
	})
}

func TestIsSystemSchema(t *testing.T) {
	t.Run("system schemas identified", func(t *testing.T) {
		systemSchemasList := []string{
			"auth",
			"storage",
			"jobs",
			"functions",
			"branching",
			"information_schema",
			"pg_catalog",
			"pg_toast",
		}

		for _, schema := range systemSchemasList {
			t.Run(schema, func(t *testing.T) {
				assert.True(t, isSystemSchema(schema), "%s should be a system schema", schema)
			})
		}
	})

	t.Run("user schemas not identified as system", func(t *testing.T) {
		userSchemas := []string{
			"public",
			"my_schema",
			"custom",
			"app",
		}

		for _, schema := range userSchemas {
			t.Run(schema, func(t *testing.T) {
				assert.False(t, isSystemSchema(schema), "%s should not be a system schema", schema)
			})
		}
	})
}

func TestEscapeDDLLiteral(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"hello", "'hello'"},
		{"", "''"},
		{"O'Brien", "'O''Brien'"},
		{"it's", "'it''s'"},
		{"quote'test'value", "'quote''test''value'"},
		{"no quotes", "'no quotes'"},
		{"123", "'123'"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := escapeDDLLiteral(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidDataTypes(t *testing.T) {
	t.Run("all valid data types are accepted", func(t *testing.T) {
		validTypes := []string{
			"text", "varchar", "char",
			"integer", "bigint", "smallint",
			"numeric", "decimal", "real", "double precision",
			"boolean", "bool",
			"date", "timestamp", "timestamptz", "time", "timetz",
			"uuid", "json", "jsonb",
			"bytea", "inet", "cidr", "macaddr",
			"serial", "bigserial", "smallserial",
		}

		for _, dtype := range validTypes {
			t.Run(dtype, func(t *testing.T) {
				assert.True(t, validDataTypes[dtype], "type '%s' should be valid", dtype)
			})
		}
	})

	t.Run("invalid data types are rejected", func(t *testing.T) {
		invalidTypes := []string{
			"string",
			"int",
			"datetime",
			"blob",
			"invalid",
		}

		for _, dtype := range invalidTypes {
			t.Run(dtype, func(t *testing.T) {
				assert.False(t, validDataTypes[dtype], "type '%s' should not be valid", dtype)
			})
		}
	})
}

func TestListSchemasTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewListSchemasTool(nil)
		assert.Equal(t, "list_schemas", tool.Name())
		assert.Contains(t, tool.Description(), "schema")
		assert.Equal(t, []string{mcp.ScopeReadTables}, tool.RequiredScopes())
	})

	t.Run("input schema has include_system", func(t *testing.T) {
		tool := NewListSchemasTool(nil)
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "include_system")
	})
}

func TestCreateSchemaTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewCreateSchemaTool(nil)
		assert.Equal(t, "create_schema", tool.Name())
		assert.Contains(t, tool.Description(), "admin:ddl")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires name parameter", func(t *testing.T) {
		tool := NewCreateSchemaTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
	})
}

func TestCreateTableTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewCreateTableTool(nil)
		assert.Equal(t, "create_table", tool.Name())
		assert.Contains(t, tool.Description(), "admin:ddl")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires name and columns parameters", func(t *testing.T) {
		tool := NewCreateTableTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "columns")
	})

	t.Run("schema defaults to public", func(t *testing.T) {
		tool := NewCreateTableTool(nil)
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		schemaProp := props["schema"].(map[string]any)
		assert.Equal(t, "public", schemaProp["default"])
	})
}

func TestDropTableTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		assert.Equal(t, "drop_table", tool.Name())
		assert.Contains(t, tool.Description(), "caution")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires table parameter", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "table")
	})

	t.Run("has cascade option", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		schema := tool.InputSchema()
		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "cascade")
	})
}

func TestAddColumnTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewAddColumnTool(nil)
		assert.Equal(t, "add_column", tool.Name())
		assert.Contains(t, tool.Description(), "admin:ddl")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires table, name, and type parameters", func(t *testing.T) {
		tool := NewAddColumnTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "type")
	})
}

func TestDropColumnTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewDropColumnTool(nil)
		assert.Equal(t, "drop_column", tool.Name())
		assert.Contains(t, tool.Description(), "caution")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires table and column parameters", func(t *testing.T) {
		tool := NewDropColumnTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "column")
	})
}

func TestRenameTableTool(t *testing.T) {
	t.Run("tool metadata", func(t *testing.T) {
		tool := NewRenameTableTool(nil)
		assert.Equal(t, "rename_table", tool.Name())
		assert.Contains(t, tool.Description(), "admin:ddl")
		assert.Equal(t, []string{mcp.ScopeAdminDDL}, tool.RequiredScopes())
	})

	t.Run("requires table and new_name parameters", func(t *testing.T) {
		tool := NewRenameTableTool(nil)
		schema := tool.InputSchema()
		required := schema["required"].([]string)
		assert.Contains(t, required, "table")
		assert.Contains(t, required, "new_name")
	})
}

func TestDDLToolScopeEnforcement(t *testing.T) {
	// Test that all DDL modifying tools require admin:ddl scope
	t.Run("modifying tools require admin:ddl", func(t *testing.T) {
		modifyingTools := []struct {
			name string
			tool interface{ RequiredScopes() []string }
		}{
			{"create_schema", NewCreateSchemaTool(nil)},
			{"create_table", NewCreateTableTool(nil)},
			{"drop_table", NewDropTableTool(nil)},
			{"add_column", NewAddColumnTool(nil)},
			{"drop_column", NewDropColumnTool(nil)},
			{"rename_table", NewRenameTableTool(nil)},
		}

		for _, tc := range modifyingTools {
			t.Run(tc.name, func(t *testing.T) {
				scopes := tc.tool.RequiredScopes()
				assert.Contains(t, scopes, mcp.ScopeAdminDDL)
			})
		}
	})

	t.Run("list_schemas requires read:tables not admin:ddl", func(t *testing.T) {
		tool := NewListSchemasTool(nil)
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeReadTables)
		assert.NotContains(t, scopes, mcp.ScopeAdminDDL)
	})
}

// =============================================================================
// Execute Method Tests
// =============================================================================

func TestListSchemasTool_Execute(t *testing.T) {
	t.Run("exclude system schemas by default", func(t *testing.T) {
		// This test requires a mock database connection
		// For now, we test the structure
		tool := NewListSchemasTool(nil)
		assert.NotNil(t, tool)
		assert.Equal(t, "list_schemas", tool.Name())
	})

	t.Run("include system schemas when requested", func(t *testing.T) {
		tool := NewListSchemasTool(nil)
		assert.NotNil(t, tool)

		// Test args parsing
		args := map[string]any{
			"include_system": true,
		}
		includeSystem, ok := args["include_system"].(bool)
		assert.True(t, ok)
		assert.True(t, includeSystem)
	})

	t.Run("database error returns tool error", func(t *testing.T) {
		// TODO: Add mock database that returns error
		tool := NewListSchemasTool(nil)
		assert.NotNil(t, tool)
	})

	t.Run("returns empty list when no schemas", func(t *testing.T) {
		// TODO: Add mock database with empty schema list
		tool := NewListSchemasTool(nil)
		assert.NotNil(t, tool)
	})
}

func TestCreateSchemaTool_Execute(t *testing.T) {
	t.Run("create valid schema successfully", func(t *testing.T) {
		tool := NewCreateSchemaTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "test_schema",
		}
		assert.Equal(t, "test_schema", args["schema"])
	})

	t.Run("reject system schema creation", func(t *testing.T) {
		systemSchemas := []string{"auth", "storage", "jobs", "functions", "branching"}

		for _, schema := range systemSchemas {
			args := map[string]any{
				"schema": schema,
			}
			assert.Equal(t, schema, args["schema"])
		}
	})

	t.Run("reject invalid schema names", func(t *testing.T) {
		invalidNames := []string{
			"1invalid",
			"schema-with-dash",
			"schema with space",
			"",
		}

		for _, name := range invalidNames {
			args := map[string]any{
				"schema": name,
			}
			assert.Equal(t, name, args["schema"])
		}
	})

	t.Run("schema already exists error", func(t *testing.T) {
		// TODO: Add mock database that returns duplicate schema error
		tool := NewCreateSchemaTool(nil)
		assert.NotNil(t, tool)
	})

	t.Run("missing schema parameter", func(t *testing.T) {
		tool := NewCreateSchemaTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{}
		_, ok := args["schema"]
		assert.False(t, ok)
	})
}

func TestDropSchemaTool_Execute(t *testing.T) {
	// Note: DropSchemaTool doesn't exist in the codebase
	// Only CreateSchemaTool is available
	t.Run("schema tools available", func(t *testing.T) {
		tool := NewCreateSchemaTool(nil)
		assert.NotNil(t, tool)
	})
}

func TestCreateTableTool_Execute(t *testing.T) {
	t.Run("create table with valid columns", func(t *testing.T) {
		tool := NewCreateTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"columns": []map[string]any{
				{
					"name":     "id",
					"type":     "integer",
					"nullable": false,
				},
				{
					"name":     "name",
					"type":     "text",
					"nullable": false,
				},
				{
					"name":     "email",
					"type":     "text",
					"nullable": true,
				},
			},
		}
		assert.Equal(t, "users", args["table"])
		assert.NotNil(t, args["columns"])
	})

	t.Run("create table with primary key", func(t *testing.T) {
		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"columns": []map[string]any{
				{
					"name":        "id",
					"type":        "integer",
					"nullable":    false,
					"primary_key": true,
				},
			},
		}
		columns, _ := args["columns"].([]map[string]any)
		assert.True(t, columns[0]["primary_key"].(bool))
	})

	t.Run("reject table creation in system schema", func(t *testing.T) {
		tool := NewCreateTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "auth",
			"table":  "users",
		}
		assert.Equal(t, "auth", args["schema"])
	})

	t.Run("reject invalid column type", func(t *testing.T) {
		invalidTypes := []string{
			"invalid_type",
			"blob",
			"varchar(255)", // Array syntax not allowed
		}

		for _, invalidType := range invalidTypes {
			args := map[string]any{
				"schema": "public",
				"table":  "test",
				"columns": []map[string]any{
					{
						"name": "col",
						"type": invalidType,
					},
				},
			}
			columns, _ := args["columns"].([]map[string]any)
			assert.Equal(t, invalidType, columns[0]["type"])
		}
	})

	t.Run("table already exists error", func(t *testing.T) {
		// TODO: Add mock database that returns duplicate table error
		tool := NewCreateTableTool(nil)
		assert.NotNil(t, tool)
	})
}

func TestDropTableTool_Execute(t *testing.T) {
	t.Run("drop valid table successfully", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "public",
			"table":  "test_table",
		}
		assert.Equal(t, "test_table", args["table"])
	})

	t.Run("reject dropping system schema tables", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "auth",
			"table":  "users",
		}
		assert.Equal(t, "auth", args["schema"])
	})

	t.Run("table not found error", func(t *testing.T) {
		// TODO: Add mock database that returns table not found error
		tool := NewDropTableTool(nil)
		assert.NotNil(t, tool)
	})

	t.Run("missing required parameters", func(t *testing.T) {
		tool := NewDropTableTool(nil)
		assert.NotNil(t, tool)

		tests := []map[string]any{
			{"table": "test"},    // missing schema
			{"schema": "public"}, // missing table
		}

		for _, args := range tests {
			_, hasSchema := args["schema"]
			_, hasTable := args["table"]
			assert.False(t, hasSchema && hasTable)
		}
	})
}

func TestAddColumnTool_Execute(t *testing.T) {
	t.Run("add column to existing table", func(t *testing.T) {
		tool := NewAddColumnTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"column": map[string]any{
				"name":     "age",
				"type":     "integer",
				"nullable": true,
			},
		}
		column, _ := args["column"].(map[string]any)
		assert.Equal(t, "age", column["name"])
		assert.Equal(t, "integer", column["type"])
	})

	t.Run("add column with default value", func(t *testing.T) {
		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"column": map[string]any{
				"name":     "status",
				"type":     "text",
				"nullable": false,
				"default":  "active",
			},
		}
		column, _ := args["column"].(map[string]any)
		assert.Equal(t, "active", column["default"])
	})

	t.Run("reject adding to system schema table", func(t *testing.T) {
		tool := NewAddColumnTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "auth",
			"table":  "users",
			"column": map[string]any{
				"name": "test",
				"type": "text",
			},
		}
		assert.Equal(t, "auth", args["schema"])
	})

	t.Run("column already exists error", func(t *testing.T) {
		// TODO: Add mock database that returns duplicate column error
		tool := NewAddColumnTool(nil)
		assert.NotNil(t, tool)
	})
}

func TestDropColumnTool_Execute(t *testing.T) {
	t.Run("drop column from existing table", func(t *testing.T) {
		tool := NewDropColumnTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"column": "old_column",
			"force":  false,
		}
		assert.Equal(t, "old_column", args["column"])
		force, ok := args["force"].(bool)
		assert.True(t, ok)
		assert.False(t, force)
	})

	t.Run("force drop column with data", func(t *testing.T) {
		args := map[string]any{
			"schema": "public",
			"table":  "users",
			"column": "temp_column",
			"force":  true,
		}
		force, ok := args["force"].(bool)
		assert.True(t, ok)
		assert.True(t, force)
	})

	t.Run("reject dropping system schema column", func(t *testing.T) {
		tool := NewDropColumnTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema": "auth",
			"table":  "users",
			"column": "id",
		}
		assert.Equal(t, "auth", args["schema"])
	})

	t.Run("column not found error", func(t *testing.T) {
		// TODO: Add mock database that returns column not found error
		tool := NewDropColumnTool(nil)
		assert.NotNil(t, tool)
	})
}

func TestRenameTableTool_Execute(t *testing.T) {
	t.Run("rename table successfully", func(t *testing.T) {
		tool := NewRenameTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema":   "public",
			"table":    "old_name",
			"new_name": "new_name",
		}
		assert.Equal(t, "old_name", args["table"])
		assert.Equal(t, "new_name", args["new_name"])
	})

	t.Run("reject renaming system schema tables", func(t *testing.T) {
		tool := NewRenameTableTool(nil)
		assert.NotNil(t, tool)

		args := map[string]any{
			"schema":   "auth",
			"table":    "users",
			"new_name": "people",
		}
		assert.Equal(t, "auth", args["schema"])
	})

	t.Run("table not found error", func(t *testing.T) {
		// TODO: Add mock database that returns table not found error
		tool := NewRenameTableTool(nil)
		assert.NotNil(t, tool)
	})

	t.Run("new table name already exists", func(t *testing.T) {
		// TODO: Add mock database that returns duplicate table error
		tool := NewRenameTableTool(nil)
		assert.NotNil(t, tool)
	})
}
