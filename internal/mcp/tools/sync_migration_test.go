package tools

import (
	"context"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
)

func TestNewSyncMigrationTool(t *testing.T) {
	t.Run("creates tool with nil dependencies", func(t *testing.T) {
		tool := NewSyncMigrationTool(nil, nil)
		assert.NotNil(t, tool)
		assert.Nil(t, tool.storage)
		assert.Nil(t, tool.executor)
	})
}

func TestSyncMigrationTool_Metadata(t *testing.T) {
	tool := NewSyncMigrationTool(nil, nil)

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "sync_migration", tool.Name())
	})

	t.Run("description", func(t *testing.T) {
		desc := tool.Description()
		assert.Contains(t, desc, "Create a database migration")
		assert.Contains(t, desc, "CAUTION")
		assert.Contains(t, desc, "up_sql")
		assert.Contains(t, desc, "down_sql")
		assert.Contains(t, desc, "dry_run")
	})

	t.Run("input schema", func(t *testing.T) {
		schema := tool.InputSchema()
		assert.Equal(t, "object", schema["type"])

		props := schema["properties"].(map[string]any)
		assert.Contains(t, props, "name")
		assert.Contains(t, props, "up_sql")
		assert.Contains(t, props, "down_sql")
		assert.Contains(t, props, "namespace")
		assert.Contains(t, props, "description")
		assert.Contains(t, props, "auto_apply")
		assert.Contains(t, props, "dry_run")

		required := schema["required"].([]string)
		assert.Contains(t, required, "name")
		assert.Contains(t, required, "up_sql")
	})

	t.Run("required scopes", func(t *testing.T) {
		scopes := tool.RequiredScopes()
		assert.Contains(t, scopes, mcp.ScopeSyncMigrations)
	})
}

func TestSyncMigrationTool_Execute_Validation(t *testing.T) {
	tool := NewSyncMigrationTool(nil, nil)

	t.Run("missing name", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"up_sql": "CREATE TABLE test (id INT);",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("missing up_sql", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name": "add_test_table",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "up_sql is required")
	})

	t.Run("invalid name format", func(t *testing.T) {
		_, err := tool.Execute(context.Background(), map[string]any{
			"name":   "invalid migration name!",
			"up_sql": "CREATE TABLE test (id INT);",
		}, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid migration name")
	})
}

func TestIsValidMigrationName(t *testing.T) {
	t.Run("valid names", func(t *testing.T) {
		validNames := []string{
			"add_users_table",
			"create-index",
			"update_schema",
			"v1_migration",
			"_private_migration",
			"A",
		}

		for _, name := range validNames {
			assert.True(t, isValidMigrationName(name), "Expected %q to be valid", name)
		}
	})

	t.Run("invalid names", func(t *testing.T) {
		invalidNames := []string{
			"",             // empty
			"1_migration",  // starts with number
			"-migration",   // starts with hyphen
			"my migration", // contains space
			"migration.up", // contains dot
			"migration@v1", // contains special char
		}

		for _, name := range invalidNames {
			assert.False(t, isValidMigrationName(name), "Expected %q to be invalid", name)
		}
	})

	t.Run("boundary - 100 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		assert.Equal(t, 100, len(name))
		assert.True(t, isValidMigrationName(name))
	})

	t.Run("boundary - 101 characters", func(t *testing.T) {
		name := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		assert.Equal(t, 101, len(name))
		assert.False(t, isValidMigrationName(name))
	})
}

func TestValidateMigrationSQL(t *testing.T) {
	t.Run("valid SQL", func(t *testing.T) {
		validSQL := []string{
			"CREATE TABLE users (id UUID PRIMARY KEY);",
			"ALTER TABLE users ADD COLUMN email VARCHAR(255);",
			"CREATE INDEX idx_users_email ON users(email);",
			"DROP TABLE IF EXISTS old_table;",
			"INSERT INTO settings (key, value) VALUES ('version', '1.0');",
		}

		for _, sql := range validSQL {
			err := validateMigrationSQL(sql)
			assert.NoError(t, err, "Expected %q to be valid", sql)
		}
	})

	t.Run("empty SQL", func(t *testing.T) {
		err := validateMigrationSQL("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("whitespace-only SQL", func(t *testing.T) {
		err := validateMigrationSQL("   \n\t  ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("drop system schema - auth", func(t *testing.T) {
		err := validateMigrationSQL("DROP SCHEMA auth CASCADE;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot drop system schemas")
	})

	t.Run("drop system schema - storage", func(t *testing.T) {
		err := validateMigrationSQL("DROP SCHEMA storage;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot drop system schemas")
	})

	t.Run("alter system schema", func(t *testing.T) {
		err := validateMigrationSQL("ALTER SCHEMA jobs RENAME TO old_jobs;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot alter system schemas")
	})

	t.Run("drop database", func(t *testing.T) {
		err := validateMigrationSQL("DROP DATABASE mydb;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot drop database")
	})

	t.Run("create database", func(t *testing.T) {
		err := validateMigrationSQL("CREATE DATABASE newdb;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot create database")
	})

	t.Run("truncate system table", func(t *testing.T) {
		err := validateMigrationSQL("TRUNCATE auth.users;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot truncate system tables")
	})

	t.Run("drop system table", func(t *testing.T) {
		err := validateMigrationSQL("DROP TABLE storage.objects;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot drop system tables")
	})

	t.Run("case insensitivity", func(t *testing.T) {
		err := validateMigrationSQL("DROP SCHEMA Auth CASCADE;")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot drop system schemas")
	})

	t.Run("allows user schema operations", func(t *testing.T) {
		validSQL := []string{
			"DROP SCHEMA my_app CASCADE;",
			"ALTER SCHEMA public RENAME TO old_public;",
			"TRUNCATE public.users;",
			"DROP TABLE my_table;",
		}

		for _, sql := range validSQL {
			err := validateMigrationSQL(sql)
			assert.NoError(t, err, "Expected %q to be valid", sql)
		}
	})
}
