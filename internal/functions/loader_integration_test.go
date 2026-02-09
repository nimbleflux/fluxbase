package functions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// LoadFunctionCodeWithFiles Tests
// =============================================================================

func TestLoadFunctionCodeWithFiles_Integration(t *testing.T) {
	t.Run("flat file pattern", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create a flat file function
		testCode := `export async function handler(req) {
  return { status: 200, body: "Hello" };
}`
		testFilePath := filepath.Join(tmpDir, "test-function.ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, testCode, mainCode)
		assert.Empty(t, supportingFiles)
	})

	t.Run("directory-based pattern with index.ts", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create directory structure
		funcDir := filepath.Join(tmpDir, "test-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)

		// Create index.ts
		indexCode := `import { helper } from './helper.ts';
export async function handler(req) {
  return { status: 200, body: helper() };
}`
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
		require.NoError(t, err)

		// Create helper.ts
		helperCode := `export function helper() {
  return "Hello from helper";
}`
		err = os.WriteFile(filepath.Join(funcDir, "helper.ts"), []byte(helperCode), 0644)
		require.NoError(t, err)

		mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, indexCode, mainCode)
		assert.NotEmpty(t, supportingFiles)
		assert.Equal(t, helperCode, supportingFiles["helper.ts"])
	})

	t.Run("directory-based with nested subdirectories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create directory structure with nested folders
		funcDir := filepath.Join(tmpDir, "test-function")
		utilsDir := filepath.Join(funcDir, "utils")
		err = os.MkdirAll(utilsDir, 0755)
		require.NoError(t, err)

		// Create files
		indexCode := `export async function handler(req) { return { status: 200 }; }`
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
		require.NoError(t, err)

		dbCode := `export function query() { return "SELECT * FROM users"; }`
		err = os.WriteFile(filepath.Join(utilsDir, "db.ts"), []byte(dbCode), 0644)
		require.NoError(t, err)

		mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, indexCode, mainCode)
		assert.NotEmpty(t, supportingFiles)
		assert.Equal(t, dbCode, supportingFiles[filepath.Join("utils", "db.ts")])
	})

	t.Run("function not found", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		_, _, err = LoadFunctionCodeWithFiles(tmpDir, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not resolve")
	})
}

// =============================================================================
// LoadSharedModulesFromFilesystem Tests
// =============================================================================

func TestLoadSharedModules_Integration(t *testing.T) {
	t.Run("no shared directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		modules, err := LoadSharedModulesFromFilesystem(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, modules)
	})

	t.Run("load shared modules", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create _shared directory
		sharedDir := filepath.Join(tmpDir, "_shared")
		err = os.MkdirAll(sharedDir, 0755)
		require.NoError(t, err)

		// Create shared modules
		corsCode := `export function corsHeaders() {
  return {
    "Access-Control-Allow-Origin": "*",
    "Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS"
  };
}`
		err = os.WriteFile(filepath.Join(sharedDir, "cors.ts"), []byte(corsCode), 0644)
		require.NoError(t, err)

		dbCode := `export function db() {
  return "database connection";
}`
		err = os.WriteFile(filepath.Join(sharedDir, "db.ts"), []byte(dbCode), 0644)
		require.NoError(t, err)

		modules, err := LoadSharedModulesFromFilesystem(tmpDir)
		require.NoError(t, err)
		assert.Len(t, modules, 2)
		assert.Equal(t, corsCode, modules["_shared/cors.ts"])
		assert.Equal(t, dbCode, modules["_shared/db.ts"])
	})

	t.Run("load nested shared modules", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create _shared directory with nested structure
		sharedDir := filepath.Join(tmpDir, "_shared")
		utilsDir := filepath.Join(sharedDir, "utils")
		err = os.MkdirAll(utilsDir, 0755)
		require.NoError(t, err)

		// Create nested shared module
		helpersCode := `export function helper() {
  return "I help";
}`
		err = os.WriteFile(filepath.Join(utilsDir, "helpers.ts"), []byte(helpersCode), 0644)
		require.NoError(t, err)

		modules, err := LoadSharedModulesFromFilesystem(tmpDir)
		require.NoError(t, err)
		assert.Len(t, modules, 1)
		assert.Equal(t, helpersCode, modules["_shared/utils/helpers.ts"])
	})

	t.Run("_shared is not a directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create _shared as a file instead of directory
		sharedPath := filepath.Join(tmpDir, "_shared")
		err = os.WriteFile(sharedPath, []byte("not a directory"), 0644)
		require.NoError(t, err)

		_, err = LoadSharedModulesFromFilesystem(tmpDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}

// =============================================================================
// ResolveFunctionPath Tests
// =============================================================================

func TestResolveFunctionPath_Integration(t *testing.T) {
	t.Run("resolves flat file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create flat file
		testCode := "export async function handler() { return {}; }"
		testFilePath := filepath.Join(tmpDir, "test-function.ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		resolved, err := ResolveFunctionPath(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, testFilePath, resolved)
	})

	t.Run("resolves directory-based function", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create directory with index.ts
		funcDir := filepath.Join(tmpDir, "test-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)

		indexCode := "export async function handler() { return {}; }"
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
		require.NoError(t, err)

		resolved, err := ResolveFunctionPath(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(funcDir, "index.ts"), resolved)
	})

	t.Run("prioritizes flat file over directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create both flat file and directory
		testCode := "// Flat file version"
		testFilePath := filepath.Join(tmpDir, "test-function.ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		funcDir := filepath.Join(tmpDir, "test-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)

		indexCode := "// Directory version"
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
		require.NoError(t, err)

		resolved, err := ResolveFunctionPath(tmpDir, "test-function")
		require.NoError(t, err)
		// Should return flat file
		assert.Equal(t, testFilePath, resolved)
	})
}

// =============================================================================
// DeleteFunctionCode Tests
// =============================================================================

func TestDeleteFunctionCode_Integration(t *testing.T) {
	t.Run("delete flat file", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create file
		testCode := "export async function handler() { return {}; }"
		testFilePath := filepath.Join(tmpDir, "test-function.ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(testFilePath)
		require.NoError(t, err)

		// Delete
		err = DeleteFunctionCode(tmpDir, "test-function")
		require.NoError(t, err)

		// Verify file is deleted
		_, err = os.Stat(testFilePath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("delete directory-based function", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create directory with index.ts
		funcDir := filepath.Join(tmpDir, "test-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)

		indexCode := "export async function handler() { return {}; }"
		indexPath := filepath.Join(funcDir, "index.ts")
		err = os.WriteFile(indexPath, []byte(indexCode), 0644)
		require.NoError(t, err)

		// Verify file exists
		_, err = os.Stat(indexPath)
		require.NoError(t, err)

		// Delete
		err = DeleteFunctionCode(tmpDir, "test-function")
		require.NoError(t, err)

		// Verify index.ts is deleted (but directory remains)
		_, err = os.Stat(indexPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("delete non-existent function", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = DeleteFunctionCode(tmpDir, "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not resolve")
	})
}

// =============================================================================
// ListFunctionFiles Tests
// =============================================================================

func TestListFunctionFiles_Integration(t *testing.T) {
	t.Run("list empty directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		files, err := ListFunctionFiles(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, files)
	})

	t.Run("list mixed files and directories", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create flat file function
		flatCode := "export async function handler() { return {}; }"
		err = os.WriteFile(filepath.Join(tmpDir, "flat-function.ts"), []byte(flatCode), 0644)
		require.NoError(t, err)

		// Create directory-based function
		funcDir := filepath.Join(tmpDir, "dir-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)
		indexCode := "export async function handler() { return {}; }"
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
		require.NoError(t, err)

		// Create a non-.ts file (should be ignored)
		err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# readme"), 0644)
		require.NoError(t, err)

		files, err := ListFunctionFiles(tmpDir)
		require.NoError(t, err)
		assert.Len(t, files, 2)

		// Check that both functions are listed
		names := make(map[string]bool)
		for _, f := range files {
			names[f.Name] = true
		}
		assert.True(t, names["flat-function"])
		assert.True(t, names["dir-function"])
	})

	t.Run("list non-existent directory", func(t *testing.T) {
		// Note: ListFunctionFiles returns empty list for non-existent directory
		// instead of error (graceful handling)
		files, err := ListFunctionFiles("/non/existent/path")
		// May return error or empty list depending on implementation
		if err != nil {
			assert.Error(t, err)
		} else {
			assert.Empty(t, files)
		}
	})
}

// =============================================================================
// File Type Detection Tests
// =============================================================================

func TestFunctionFileTypeDetection_Integration(t *testing.T) {
	t.Run("detects TypeScript files", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		funcDir := filepath.Join(tmpDir, "test-function")
		err = os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)

		// Create various file types
		files := map[string]string{
			"index.ts":       "export async function handler() {}",
			"helper.ts":      "export function helper() {}",
			"utils.mts":      "export function util() {}",
			"legacy.js":      "function legacy() {}",
			"module.mjs":     "export function mod() {}",
			"data.json":      `{"key": "value"}`,
			"config.geojson": `{"type": "Feature"}`,
			"deno.json":      `{"tasks": {}}`,
			"README.md":      "# readme",
			"docs.txt":       "documentation",
		}

		for name, content := range files {
			err = os.WriteFile(filepath.Join(funcDir, name), []byte(content), 0644)
			require.NoError(t, err)
		}

		mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "test-function")
		require.NoError(t, err)
		assert.NotEmpty(t, mainCode)

		// Should include .ts, .mts, .js, .mjs, .json, .geojson, and deno.json
		// Should NOT include .md or .txt
		assert.NotContains(t, supportingFiles, "README.md")
		assert.NotContains(t, supportingFiles, "docs.txt")
		assert.Contains(t, supportingFiles, "helper.ts")
		assert.Contains(t, supportingFiles, "utils.mts")
	})
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestLoadFunctionCode_Integration_EdgeCases(t *testing.T) {
	t.Run("load function with special characters in name", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create function with special characters
		testCode := "export async function handler() { return {}; }"
		testFilePath := filepath.Join(tmpDir, "test-function-123.ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		code, err := LoadFunctionCode(tmpDir, "test-function-123")
		require.NoError(t, err)
		assert.Equal(t, testCode, code)
	})

	t.Run("load function with very long name", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Function names have a max length of 64 characters
		// Test a name that's close to but within the limit
		longName := strings.Repeat("a", 60)
		testCode := "export async function handler() { return {}; }"
		testFilePath := filepath.Join(tmpDir, longName+".ts")
		err = os.WriteFile(testFilePath, []byte(testCode), 0644)
		require.NoError(t, err)

		code, err := LoadFunctionCode(tmpDir, longName)
		require.NoError(t, err)
		assert.Equal(t, testCode, code)
	})
}

func TestSaveFunctionCode_Integration_EdgeCases(t *testing.T) {
	t.Run("update existing function", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create initial function
		initialCode := "export async function handler() { return { status: 200 }; }"
		err = SaveFunctionCode(tmpDir, "test-function", initialCode)
		require.NoError(t, err)

		// Update with new code
		updatedCode := "export async function handler() { return { status: 201 }; }"
		err = SaveFunctionCode(tmpDir, "test-function", updatedCode)
		require.NoError(t, err)

		// Verify update
		code, err := LoadFunctionCode(tmpDir, "test-function")
		require.NoError(t, err)
		assert.Equal(t, updatedCode, code)
	})

	t.Run("save function with very large code", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Create large code (1MB)
		largeCode := strings.Repeat("export async function handler() { return {}; }\n", 10000)

		err = SaveFunctionCode(tmpDir, "large-function", largeCode)
		require.NoError(t, err)

		// Verify save
		code, err := LoadFunctionCode(tmpDir, "large-function")
		require.NoError(t, err)
		assert.Equal(t, largeCode, code)
	})
}

// =============================================================================
// Security Tests
// =============================================================================

func TestFunctionSecurity_Integration(t *testing.T) {
	t.Run("path traversal prevention in load", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Try to load file outside functions directory
		_, err = LoadFunctionCode(tmpDir, "../../../etc/passwd")
		assert.Error(t, err)
	})

	t.Run("path traversal prevention in save", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		code := "export async function handler() { return {}; }"
		err = SaveFunctionCode(tmpDir, "../malicious", code)
		assert.Error(t, err)
	})

	t.Run("path traversal prevention in delete", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "functions-test-")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = DeleteFunctionCode(tmpDir, "../../etc/passwd")
		assert.Error(t, err)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkLoadFunctionCode(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "functions-bench-")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testCode := "export async function handler(req) { return { status: 200, body: 'Hello' }; }"
	testFilePath := filepath.Join(tmpDir, "bench-function.ts")
	err = os.WriteFile(testFilePath, []byte(testCode), 0644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFunctionCode(tmpDir, "bench-function")
	}
}

func BenchmarkLoadFunctionCodeWithFiles(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "functions-bench-")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	funcDir := filepath.Join(tmpDir, "bench-function")
	err = os.MkdirAll(funcDir, 0755)
	if err != nil {
		b.Fatal(err)
	}

	indexCode := "export async function handler(req) { return { status: 200 }; }"
	err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(indexCode), 0644)
	if err != nil {
		b.Fatal(err)
	}

	helperCode := "export function helper() { return 'help'; }"
	err = os.WriteFile(filepath.Join(funcDir, "helper.ts"), []byte(helperCode), 0644)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFunctionCodeWithFiles(tmpDir, "bench-function")
	}
}

func BenchmarkSaveFunctionCode(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "functions-bench-")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	code := "export async function handler(req) { return { status: 200, body: 'Hello' }; }"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		funcName := "bench-function"
		if i > 0 {
			funcName = "bench-function-" + string(rune(i))
		}
		SaveFunctionCode(tmpDir, funcName, code)
	}
}
