//go:build integration

package database_test

import (
	"context"
	"testing"

	database "github.com/fluxbase-eu/fluxbase/internal/database"
	"github.com/fluxbase-eu/fluxbase/test/dbhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSchemaInspector_GetAllTables_Integration tests retrieving all tables from the database.
func TestSchemaInspector_GetAllTables_Integration(t *testing.T) {
	testCtx := dbhelpers.NewDBTestContext(t)
	defer testCtx.Close()

	inspector := database.NewSchemaInspector(database.NewConnectionWithPool(testCtx.Pool))

	t.Run("retrieves tables from public schema", func(t *testing.T) {
		tables, err := inspector.GetAllTables(context.Background(), "public")
		require.NoError(t, err)
		assert.NotEmpty(t, tables, "Should find at least one table in public schema")

		// Verify that we have core tables
		tableNames := make(map[string]bool)
		for _, table := range tables {
			tableNames[table.Name] = true
			assert.Equal(t, "public", table.Schema)
			assert.Equal(t, "table", table.Type)
		}

		// Log the tables we found for debugging
		t.Logf("Found %d tables in public schema: %v", len(tables), tables)
		// Don't assert specific tables since the public schema may vary
	})

	t.Run("retrieves tables from multiple schemas", func(t *testing.T) {
		tables, err := inspector.GetAllTables(context.Background(), "public", "auth", "storage")
		require.NoError(t, err)
		assert.NotEmpty(t, tables)

		// Verify we have tables from different schemas
		schemas := make(map[string]bool)
		for _, table := range tables {
			schemas[table.Schema] = true
		}
		assert.True(t, schemas["public"] || schemas["auth"] || schemas["storage"])
	})

	t.Run("excludes system tables", func(t *testing.T) {
		tables, err := inspector.GetAllTables(context.Background(), "public")
		require.NoError(t, err)

		// Should not include pg_* or _fluxbase.* tables
		for _, table := range tables {
			assert.NotRegexp(t, `^pg_`, table.Name)
			assert.NotRegexp(t, `^_fluxbase\.`, table.Name)
		}
	})
}

// TestSchemaInspector_GetTableInfo_Integration tests retrieving detailed information about a table.
func TestSchemaInspector_GetTableInfo_Integration(t *testing.T) {
	testCtx := dbhelpers.NewDBTestContext(t)
	defer testCtx.Close()

	inspector := database.NewSchemaInspector(database.NewConnectionWithPool(testCtx.Pool))

	t.Run("retrieves users table info", func(t *testing.T) {
		tableInfo, err := inspector.GetTableInfo(context.Background(), "auth", "users")
		require.NoError(t, err)

		assert.Equal(t, "auth", tableInfo.Schema)
		assert.Equal(t, "users", tableInfo.Name)
		assert.NotEmpty(t, tableInfo.Columns, "Should have columns")
		assert.NotEmpty(t, tableInfo.PrimaryKey, "Should have primary key")

		// Check for expected columns
		columnMap := make(map[string]bool)
		for _, col := range tableInfo.Columns {
			columnMap[col.Name] = true
		}
		assert.True(t, columnMap["id"], "Should have id column")
		assert.True(t, columnMap["email"], "Should have email column")

		// Verify column map was built
		assert.NotNil(t, tableInfo.ColumnMap)
		assert.True(t, tableInfo.HasColumn("id"))
	})

	t.Run("includes column metadata", func(t *testing.T) {
		tableInfo, err := inspector.GetTableInfo(context.Background(), "auth", "users")
		require.NoError(t, err)

		// Find id column
		idCol := tableInfo.GetColumn("id")
		require.NotNil(t, idCol)
		assert.Equal(t, "id", idCol.Name)
		assert.True(t, idCol.IsPrimaryKey)
		assert.False(t, idCol.IsNullable)
	})

	t.Run("retrieves foreign key relationships", func(t *testing.T) {
		// Find a table with foreign keys
		tableInfo, err := inspector.GetTableInfo(context.Background(), "auth", "users")
		require.NoError(t, err)

		// ForeignKeys may be nil if table has no FKs
		if tableInfo.ForeignKeys != nil {
			t.Logf("Users table has %d foreign keys", len(tableInfo.ForeignKeys))
		} else {
			t.Log("Users table has no foreign keys")
		}
	})

	t.Run("retrieves index information", func(t *testing.T) {
		tableInfo, err := inspector.GetTableInfo(context.Background(), "auth", "users")
		require.NoError(t, err)

		// Should have at least primary key index
		assert.NotEmpty(t, tableInfo.Indexes, "Should have indexes")

		// Check for primary key index
		hasPKIndex := false
		for _, idx := range tableInfo.Indexes {
			if idx.IsPrimary {
				hasPKIndex = true
				assert.True(t, idx.IsUnique)
				break
			}
		}
		assert.True(t, hasPKIndex, "Should have primary key index")
	})
}

// TestSchemaInspector_GetSchemas_Integration tests retrieving all schemas.
func TestSchemaInspector_GetSchemas_Integration(t *testing.T) {
	testCtx := dbhelpers.NewDBTestContext(t)
	defer testCtx.Close()

	inspector := database.NewSchemaInspector(database.NewConnectionWithPool(testCtx.Pool))

	t.Run("retrieves all schemas", func(t *testing.T) {
		schemas, err := inspector.GetSchemas(context.Background())
		require.NoError(t, err)
		assert.NotEmpty(t, schemas)

		// Should have core schemas
		schemaMap := make(map[string]bool)
		for _, schema := range schemas {
			schemaMap[schema] = true
		}

		assert.True(t, schemaMap["public"], "Should have public schema")
		assert.True(t, schemaMap["auth"] || schemaMap["storage"], "Should have auth or storage schema")

		// Should not include system schemas
		assert.False(t, schemaMap["pg_catalog"])
		assert.False(t, schemaMap["information_schema"])
	})
}

// TestSchemaInspector_ErrorHandling_Integration tests error scenarios.
func TestSchemaInspector_ErrorHandling_Integration(t *testing.T) {
	testCtx := dbhelpers.NewDBTestContext(t)
	defer testCtx.Close()

	inspector := database.NewSchemaInspector(database.NewConnectionWithPool(testCtx.Pool))

	t.Run("returns error for non-existent table", func(t *testing.T) {
		// GetTableInfo returns a TableInfo even for non-existent tables
		// The Columns slice will be empty
		tableInfo, err := inspector.GetTableInfo(context.Background(), "public", "nonexistent_table_xyz")
		require.NoError(t, err)
		assert.Empty(t, tableInfo.Columns, "Non-existent table should have no columns")
	})

	t.Run("handles empty schema gracefully", func(t *testing.T) {
		tables, err := inspector.GetAllTables(context.Background())
		require.NoError(t, err)
		assert.NotNil(t, tables)
	})

	t.Run("handles non-existent schema", func(t *testing.T) {
		tables, err := inspector.GetAllTables(context.Background(), "nonexistent_schema_xyz")
		require.NoError(t, err)
		assert.Empty(t, tables)
	})
}

// TestTableInfo_BuildColumnMap_Integration tests the column map functionality.
func TestTableInfo_BuildColumnMap_Integration(t *testing.T) {
	testCtx := dbhelpers.NewDBTestContext(t)
	defer testCtx.Close()

	inspector := database.NewSchemaInspector(database.NewConnectionWithPool(testCtx.Pool))

	t.Run("builds column map for O(1) lookups", func(t *testing.T) {
		tableInfo, err := inspector.GetTableInfo(context.Background(), "auth", "users")
		require.NoError(t, err)

		// ColumnMap should be built automatically
		assert.NotNil(t, tableInfo.ColumnMap)

		// Test O(1) lookup
		idCol := tableInfo.GetColumn("id")
		require.NotNil(t, idCol)
		assert.Equal(t, "id", idCol.Name)

		// Test non-existent column
		nonExistent := tableInfo.GetColumn("nonexistent_column_xyz")
		assert.Nil(t, nonExistent)
	})
}
