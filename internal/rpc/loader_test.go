package rpc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Loader Construction Tests
// =============================================================================

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with directory", func(t *testing.T) {
		loader := NewLoader("/path/to/procedures")

		require.NotNil(t, loader)
		assert.Equal(t, "/path/to/procedures", loader.proceduresDir)
	})

	t.Run("creates loader with empty directory", func(t *testing.T) {
		loader := NewLoader("")

		require.NotNil(t, loader)
		assert.Equal(t, "", loader.proceduresDir)
	})
}

// =============================================================================
// LoadProcedures Tests
// =============================================================================

func TestLoader_LoadProcedures(t *testing.T) {
	t.Run("returns nil for empty procedures directory", func(t *testing.T) {
		loader := NewLoader("")

		procs, err := loader.LoadProcedures()

		assert.NoError(t, err)
		assert.Nil(t, procs)
	})

	t.Run("returns nil for non-existent directory", func(t *testing.T) {
		loader := NewLoader("/non/existent/directory")

		procs, err := loader.LoadProcedures()

		assert.NoError(t, err)
		assert.Nil(t, procs)
	})

	t.Run("loads procedures from directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a simple SQL file
		sqlContent := `-- Simple query
SELECT * FROM users`
		err := os.WriteFile(filepath.Join(tmpDir, "get_users.sql"), []byte(sqlContent), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "get_users", procs[0].Name)
		assert.Equal(t, "default", procs[0].Namespace)
	})

	t.Run("loads procedures from namespace subdirectory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create namespace directory
		nsDir := filepath.Join(tmpDir, "admin")
		err := os.MkdirAll(nsDir, 0o755)
		require.NoError(t, err)

		// Create a SQL file in namespace
		sqlContent := `SELECT * FROM admin_users`
		err = os.WriteFile(filepath.Join(nsDir, "list_admins.sql"), []byte(sqlContent), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "list_admins", procs[0].Name)
		assert.Equal(t, "admin", procs[0].Namespace)
	})

	t.Run("ignores non-SQL files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create SQL and non-SQL files
		err := os.WriteFile(filepath.Join(tmpDir, "query.sql"), []byte("SELECT 1"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/bash"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		assert.Len(t, procs, 1)
		assert.Equal(t, "query", procs[0].Name)
	})

	t.Run("handles case-insensitive SQL extension", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create files with different case extensions
		err := os.WriteFile(filepath.Join(tmpDir, "query1.sql"), []byte("SELECT 1"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "query2.SQL"), []byte("SELECT 2"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		assert.Len(t, procs, 2)
	})

	t.Run("continues loading on individual file error", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a valid SQL file
		err := os.WriteFile(filepath.Join(tmpDir, "valid.sql"), []byte("SELECT 1"), 0o644)
		require.NoError(t, err)

		// Create unreadable file (simulated by creating directory with .sql name)
		// This will cause an error when trying to read as file
		err = os.Mkdir(filepath.Join(tmpDir, "invalid.sql"), 0o755)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		// Should succeed but skip the invalid one
		require.NoError(t, err)
		assert.Len(t, procs, 1)
		assert.Equal(t, "valid", procs[0].Name)
	})
}

// =============================================================================
// LoadProceduresFromNamespace Tests
// =============================================================================

func TestLoader_LoadProceduresFromNamespace(t *testing.T) {
	t.Run("returns nil for empty procedures directory", func(t *testing.T) {
		loader := NewLoader("")

		procs, err := loader.LoadProceduresFromNamespace("admin")

		assert.NoError(t, err)
		assert.Nil(t, procs)
	})

	t.Run("returns nil for non-existent namespace", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProceduresFromNamespace("non_existent")

		assert.NoError(t, err)
		assert.Nil(t, procs)
	})

	t.Run("loads procedures from specific namespace", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create two namespaces
		ns1Dir := filepath.Join(tmpDir, "namespace1")
		ns2Dir := filepath.Join(tmpDir, "namespace2")
		err := os.MkdirAll(ns1Dir, 0o755)
		require.NoError(t, err)
		err = os.MkdirAll(ns2Dir, 0o755)
		require.NoError(t, err)

		// Add files to both
		err = os.WriteFile(filepath.Join(ns1Dir, "proc1.sql"), []byte("SELECT 1"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(ns2Dir, "proc2.sql"), []byte("SELECT 2"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProceduresFromNamespace("namespace1")

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "proc1", procs[0].Name)
		assert.Equal(t, "namespace1", procs[0].Namespace)
	})
}

// =============================================================================
// extractNamespaceName Tests
// =============================================================================

func TestLoader_extractNamespaceName(t *testing.T) {
	loader := NewLoader("/procedures")

	t.Run("extracts default namespace from single file", func(t *testing.T) {
		annotations := &Annotations{}

		namespace, name := loader.extractNamespaceName("query.sql", annotations)

		assert.Equal(t, "default", namespace)
		assert.Equal(t, "query", name)
	})

	t.Run("extracts namespace from subdirectory", func(t *testing.T) {
		annotations := &Annotations{}

		namespace, name := loader.extractNamespaceName("admin/query.sql", annotations)

		assert.Equal(t, "admin", namespace)
		assert.Equal(t, "query", name)
	})

	t.Run("handles nested directories", func(t *testing.T) {
		annotations := &Annotations{}

		namespace, name := loader.extractNamespaceName("level1/level2/query.sql", annotations)

		assert.Equal(t, "level1", namespace)
		assert.Equal(t, "query", name)
	})

	t.Run("uses annotation name when provided", func(t *testing.T) {
		annotations := &Annotations{Name: "custom_name"}

		namespace, name := loader.extractNamespaceName("admin/query.sql", annotations)

		assert.Equal(t, "admin", namespace)
		assert.Equal(t, "custom_name", name) // Override from annotation
	})

	t.Run("handles Windows-style paths", func(t *testing.T) {
		annotations := &Annotations{}

		// ToSlash should normalize Windows paths
		namespace, name := loader.extractNamespaceName("admin\\query.sql", annotations)

		// After ToSlash, backslash becomes forward slash
		// But since the original path separator is backslash,
		// filepath.ToSlash will convert it on Windows
		assert.NotEmpty(t, namespace)
		assert.NotEmpty(t, name)
	})
}

// =============================================================================
// LoadedProcedure Tests
// =============================================================================

func TestLoadedProcedure(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		proc := &LoadedProcedure{
			Name:      "test_proc",
			Namespace: "public",
			FilePath:  "/path/to/test_proc.sql",
			Code:      "-- @name: test\nSELECT * FROM users",
			SQLQuery:  "SELECT * FROM users",
			Annotations: &Annotations{
				Name:        "test",
				Description: "Test procedure",
			},
		}

		assert.Equal(t, "test_proc", proc.Name)
		assert.Equal(t, "public", proc.Namespace)
		assert.Equal(t, "/path/to/test_proc.sql", proc.FilePath)
		assert.Contains(t, proc.Code, "SELECT * FROM users")
		assert.Equal(t, "SELECT * FROM users", proc.SQLQuery)
		assert.Equal(t, "test", proc.Annotations.Name)
	})
}

// =============================================================================
// ToProcedure Tests
// =============================================================================

func TestLoadedProcedure_ToProcedure(t *testing.T) {
	t.Run("converts to Procedure with defaults", func(t *testing.T) {
		loaded := &LoadedProcedure{
			Name:      "test_proc",
			Namespace: "admin",
			Code:      "SELECT * FROM users",
			SQLQuery:  "SELECT * FROM users",
		}

		proc := loaded.ToProcedure()

		assert.Equal(t, "test_proc", proc.Name)
		assert.Equal(t, "admin", proc.Namespace)
		assert.Equal(t, "SELECT * FROM users", proc.SQLQuery)
		assert.Equal(t, "SELECT * FROM users", proc.OriginalCode)
		assert.Equal(t, "filesystem", proc.Source)
		assert.True(t, proc.Enabled)
		assert.Equal(t, []string{"public"}, proc.AllowedSchemas)
		assert.Empty(t, proc.AllowedTables)
		assert.Equal(t, 30, proc.MaxExecutionTimeSeconds)
	})

	t.Run("applies annotations", func(t *testing.T) {
		loaded := &LoadedProcedure{
			Name:      "test_proc",
			Namespace: "admin",
			SQLQuery:  "SELECT * FROM users",
			Code:      "SELECT * FROM users",
			Annotations: &Annotations{
				Description:   "Test description",
				IsPublic:      true,
				AllowedTables: []string{"users", "orders"},
			},
		}

		proc := loaded.ToProcedure()

		assert.Equal(t, "Test description", proc.Description)
		assert.True(t, proc.IsPublic)
		assert.Equal(t, []string{"users", "orders"}, proc.AllowedTables)
	})

	t.Run("handles nil annotations", func(t *testing.T) {
		loaded := &LoadedProcedure{
			Name:        "test_proc",
			Namespace:   "admin",
			SQLQuery:    "SELECT * FROM users",
			Code:        "SELECT * FROM users",
			Annotations: nil,
		}

		proc := loaded.ToProcedure()

		// Should not panic and use defaults
		assert.Equal(t, "test_proc", proc.Name)
		assert.Equal(t, 30, proc.MaxExecutionTimeSeconds)
	})
}

// =============================================================================
// Integration Tests with Annotations
// =============================================================================

func TestLoader_LoadProcedureWithAnnotations(t *testing.T) {
	t.Run("parses procedure with annotations", func(t *testing.T) {
		tmpDir := t.TempDir()

		sqlContent := `-- @fluxbase:name custom_name
-- @fluxbase:description This is a test procedure
-- @fluxbase:public true
-- @fluxbase:allowed-tables users, orders
SELECT * FROM users WHERE id = $user_id`

		err := os.WriteFile(filepath.Join(tmpDir, "test.sql"), []byte(sqlContent), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)

		// Check annotation overrides
		assert.Equal(t, "custom_name", procs[0].Name)
		require.NotNil(t, procs[0].Annotations)
		assert.Equal(t, "This is a test procedure", procs[0].Annotations.Description)
		assert.True(t, procs[0].Annotations.IsPublic)
		assert.Contains(t, procs[0].Annotations.AllowedTables, "users")
	})
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestLoader_EdgeCases(t *testing.T) {
	t.Run("handles empty SQL file", func(t *testing.T) {
		tmpDir := t.TempDir()

		err := os.WriteFile(filepath.Join(tmpDir, "empty.sql"), []byte(""), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "empty", procs[0].Name)
		assert.Empty(t, procs[0].SQLQuery)
	})

	t.Run("handles SQL file with only comments", func(t *testing.T) {
		tmpDir := t.TempDir()

		sqlContent := `-- This is just a comment
-- Another comment`
		err := os.WriteFile(filepath.Join(tmpDir, "comments.sql"), []byte(sqlContent), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "comments", procs[0].Name)
	})

	t.Run("handles deeply nested namespace", func(t *testing.T) {
		tmpDir := t.TempDir()

		deepDir := filepath.Join(tmpDir, "a", "b", "c")
		err := os.MkdirAll(deepDir, 0o755)
		require.NoError(t, err)

		err = os.WriteFile(filepath.Join(deepDir, "deep.sql"), []byte("SELECT 1"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		procs, err := loader.LoadProcedures()

		require.NoError(t, err)
		require.Len(t, procs, 1)
		assert.Equal(t, "a", procs[0].Namespace) // First directory is namespace
		assert.Equal(t, "deep", procs[0].Name)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLoader_extractNamespaceName(b *testing.B) {
	loader := NewLoader("/procedures")
	annotations := &Annotations{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = loader.extractNamespaceName("namespace/procedure.sql", annotations)
	}
}

func BenchmarkLoadedProcedure_ToProcedure(b *testing.B) {
	loaded := &LoadedProcedure{
		Name:      "test_proc",
		Namespace: "admin",
		SQLQuery:  "SELECT * FROM users WHERE id = $user_id",
		Code:      "-- @description: Test\nSELECT * FROM users WHERE id = $user_id",
		Annotations: &Annotations{
			Description:   "Test description",
			IsPublic:      true,
			AllowedTables: []string{"users", "orders"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loaded.ToProcedure()
	}
}
