package functions

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ValidateFunctionName Tests
// =============================================================================

func TestValidateFunctionName_ValidNames(t *testing.T) {
	validNames := []string{
		"myFunction",
		"my-function",
		"my_function",
		"MyFunction123",
		"abc",
		"test123",
		"a-b-c",
		"x_y_z",
		"ABC123",
	}

	for _, name := range validNames {
		t.Run(name, func(t *testing.T) {
			err := ValidateFunctionName(name)
			assert.NoError(t, err, "Name should be valid: %s", name)
		})
	}
}

func TestValidateFunctionName_InvalidNames(t *testing.T) {
	tests := []struct {
		name         string
		functionName string
		expectedErr  string
	}{
		{
			name:         "empty name",
			functionName: "",
			expectedErr:  "function name cannot be empty",
		},
		{
			name:         "too long",
			functionName: "a", // 65 characters
			expectedErr:  "function name too long",
		},
		{
			name:         "reserved dot",
			functionName: ".",
			expectedErr:  "function name '.' is reserved",
		},
		{
			name:         "reserved dotdot",
			functionName: "..",
			expectedErr:  "function name '..' is reserved",
		},
		{
			name:         "reserved index",
			functionName: "index",
			expectedErr:  "function name 'index' is reserved",
		},
		{
			name:         "reserved main",
			functionName: "main",
			expectedErr:  "function name 'main' is reserved",
		},
		{
			name:         "path separator slash",
			functionName: "my/function",
			expectedErr:  "must contain only letters",
		},
		{
			name:         "path separator backslash",
			functionName: "my\\function",
			expectedErr:  "must contain only letters",
		},
		{
			name:         "special characters",
			functionName: "my.function",
			expectedErr:  "must contain only letters",
		},
		{
			name:         "space",
			functionName: "my function",
			expectedErr:  "must contain only letters",
		},
		{
			name:         "null byte attempt",
			functionName: "my\x00function",
			expectedErr:  "must contain only letters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create too long name if needed
			name := tt.functionName
			if name == "a" {
				name = string(make([]byte, 65))
				for i := range name {
					name = name[:i] + "a" + name[i+1:]
				}
			}

			err := ValidateFunctionName(name)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

// =============================================================================
// ValidateFunctionPath Tests
// =============================================================================

func TestValidateFunctionPath_ValidPaths(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		functionName string
	}{
		{
			name:         "simple name",
			functionName: "myfunction",
		},
		{
			name:         "with hyphens",
			functionName: "my-function",
		},
		{
			name:         "with underscores",
			functionName: "my_function",
		},
		{
			name:         "mixed case",
			functionName: "MyFunction",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ValidateFunctionPath(tmpDir, tt.functionName)
			require.NoError(t, err)
			assert.Contains(t, path, tt.functionName+".ts")
			assert.True(t, filepath.IsAbs(path))
		})
	}
}

func TestValidateFunctionPath_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name         string
		functionName string
		expectedErr  string
	}{
		{
			name:         "dotdot slash",
			functionName: "../escape",
			expectedErr:  "must contain only letters",
		},
		{
			name:         "slash in name",
			functionName: "subdir/function",
			expectedErr:  "must contain only letters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateFunctionPath(tmpDir, tt.functionName)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestValidateFunctionPath_DirectoryEscape(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to escape using path traversal
	_, err := ValidateFunctionPath(tmpDir, "../../etc/passwd")
	require.Error(t, err)
}

// =============================================================================
// ValidateFunctionCode Tests
// =============================================================================

func TestValidateFunctionCode_ValidCode(t *testing.T) {
	validCode := `
		export default async function(req) {
			return new Response("Hello World");
		}
	`

	err := ValidateFunctionCode(validCode)
	assert.NoError(t, err)
}

func TestValidateFunctionCode_EmptyCode(t *testing.T) {
	err := ValidateFunctionCode("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

func TestValidateFunctionCode_CodeTooLarge(t *testing.T) {
	// Create code larger than 1MB
	largeCode := string(make([]byte, 1024*1024+1))

	err := ValidateFunctionCode(largeCode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}

// =============================================================================
// LoadFunctionCode Tests
// =============================================================================

func TestLoadFunctionCode_FlatFile(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`
	filePath := filepath.Join(tmpDir, "myfunction.ts")

	err := os.WriteFile(filePath, []byte(code), 0644)
	require.NoError(t, err)

	loadedCode, err := LoadFunctionCode(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, code, loadedCode)
}

func TestLoadFunctionCode_DirectoryBased(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`
	funcDir := filepath.Join(tmpDir, "myfunction")

	err := os.MkdirAll(funcDir, 0755)
	require.NoError(t, err)

	indexPath := filepath.Join(funcDir, "index.ts")
	err = os.WriteFile(indexPath, []byte(code), 0644)
	require.NoError(t, err)

	loadedCode, err := LoadFunctionCode(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, code, loadedCode)
}

func TestLoadFunctionCode_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := LoadFunctionCode(tmpDir, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not resolve function path")
}

func TestLoadFunctionCode_Priority(t *testing.T) {
	tmpDir := t.TempDir()

	// Create both flat file and directory
	flatCode := `// Flat file`
	directoryCode := `// Directory index`

	err := os.WriteFile(filepath.Join(tmpDir, "myfunction.ts"), []byte(flatCode), 0644)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "myfunction"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "myfunction", "index.ts"), []byte(directoryCode), 0644)
	require.NoError(t, err)

	// Flat file should have priority
	loadedCode, err := LoadFunctionCode(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, flatCode, loadedCode)
}

// =============================================================================
// LoadFunctionCodeWithFiles Tests
// =============================================================================

func TestLoadFunctionCodeWithFiles_FlatFileNoSupporting(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	err := os.WriteFile(filepath.Join(tmpDir, "myfunction.ts"), []byte(code), 0644)
	require.NoError(t, err)

	mainCode, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, code, mainCode)
	assert.Empty(t, supportingFiles)
}

func TestLoadFunctionCodeWithFiles_DirectoryWithSupporting(t *testing.T) {
	tmpDir := t.TempDir()
	mainCode := `export default async function(req) { return new Response("Hello"); }`
	utilsCode := `export function helper() { return "helper"; }`

	funcDir := filepath.Join(tmpDir, "myfunction")
	err := os.MkdirAll(funcDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(mainCode), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "utils.ts"), []byte(utilsCode), 0644)
	require.NoError(t, err)

	loadedMain, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, mainCode, loadedMain)
	assert.Len(t, supportingFiles, 1)
	assert.Equal(t, utilsCode, supportingFiles["utils.ts"])
}

func TestLoadFunctionCodeWithFiles_NestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	mainCode := `export default async function(req) { return new Response("Hello"); }`
	utilsCode := `export function helper() { return "helper"; }`

	funcDir := filepath.Join(tmpDir, "myfunction")
	err := os.MkdirAll(filepath.Join(funcDir, "helpers"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(mainCode), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "helpers", "utils.ts"), []byte(utilsCode), 0644)
	require.NoError(t, err)

	loadedMain, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, mainCode, loadedMain)
	assert.Len(t, supportingFiles, 1)
	assert.Equal(t, utilsCode, supportingFiles["helpers/utils.ts"])
}

func TestLoadFunctionCodeWithFiles_DenoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	mainCode := `export default async function(req) { return new Response("Hello"); }`
	denoConfig := `{
		"imports": {
			"hooks": "npm:@deno/hooks/x"
		}
	}`

	funcDir := filepath.Join(tmpDir, "myfunction")
	err := os.MkdirAll(funcDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(mainCode), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "deno.json"), []byte(denoConfig), 0644)
	require.NoError(t, err)

	loadedMain, supportingFiles, err := LoadFunctionCodeWithFiles(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.Equal(t, mainCode, loadedMain)
	assert.Len(t, supportingFiles, 1)
	assert.Equal(t, denoConfig, supportingFiles["deno.json"])
}

// =============================================================================
// SaveFunctionCode Tests
// =============================================================================

func TestSaveFunctionCode_Success(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	err := SaveFunctionCode(tmpDir, "myfunction", code)
	require.NoError(t, err)

	// Verify file was created
	filePath := filepath.Join(tmpDir, "myfunction.ts")
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Verify content
	loadedCode, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Equal(t, code, string(loadedCode))
}

func TestSaveFunctionCode_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	err := SaveFunctionCode(tmpDir, "../escape", code)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid function name")
}

func TestSaveFunctionCode_EmptyCode(t *testing.T) {
	tmpDir := t.TempDir()

	err := SaveFunctionCode(tmpDir, "myfunction", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid function code")
}

// =============================================================================
// DeleteFunctionCode Tests
// =============================================================================

func TestDeleteFunctionCode_FlatFile(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	filePath := filepath.Join(tmpDir, "myfunction.ts")
	err := os.WriteFile(filePath, []byte(code), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Delete
	err = DeleteFunctionCode(tmpDir, "myfunction")
	require.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(filePath)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFunctionCode_DirectoryBased(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	funcDir := filepath.Join(tmpDir, "myfunction")
	err := os.MkdirAll(funcDir, 0755)
	require.NoError(t, err)

	indexPath := filepath.Join(funcDir, "index.ts")
	err = os.WriteFile(indexPath, []byte(code), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(indexPath)
	assert.NoError(t, err)

	// Delete
	err = DeleteFunctionCode(tmpDir, "myfunction")
	require.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(indexPath)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFunctionCode_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	err := DeleteFunctionCode(tmpDir, "nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not resolve function path")
}

// =============================================================================
// FunctionExists Tests
// =============================================================================

func TestFunctionExists_FlatFile(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	err := os.WriteFile(filepath.Join(tmpDir, "myfunction.ts"), []byte(code), 0644)
	require.NoError(t, err)

	exists, err := FunctionExists(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFunctionExists_DirectoryBased(t *testing.T) {
	tmpDir := t.TempDir()
	code := `export default async function(req) { return new Response("Hello"); }`

	funcDir := filepath.Join(tmpDir, "myfunction")
	err := os.MkdirAll(funcDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(code), 0644)
	require.NoError(t, err)

	exists, err := FunctionExists(tmpDir, "myfunction")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestFunctionExists_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	exists, err := FunctionExists(tmpDir, "nonexistent")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFunctionExists_InvalidName(t *testing.T) {
	tmpDir := t.TempDir()

	exists, err := FunctionExists(tmpDir, "../escape")
	require.Error(t, err)
	assert.False(t, exists)
}

// =============================================================================
// ListFunctionFiles Tests
// =============================================================================

func TestListFunctionFiles_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	files, err := ListFunctionFiles(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, files)
}

func TestListFunctionFiles_FlatFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple function files
	code := `export default async function(req) { return new Response("Hello"); }`
	for _, name := range []string{"func1", "func2", "func3"} {
		err := os.WriteFile(filepath.Join(tmpDir, name+".ts"), []byte(code), 0644)
		require.NoError(t, err)
	}

	files, err := ListFunctionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 3)

	// Check names
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name
	}
	assert.Contains(t, names, "func1")
	assert.Contains(t, names, "func2")
	assert.Contains(t, names, "func3")
}

func TestListFunctionFiles_DirectoryBased(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory-based functions
	code := `export default async function(req) { return new Response("Hello"); }`
	for _, name := range []string{"func1", "func2"} {
		funcDir := filepath.Join(tmpDir, name)
		err := os.MkdirAll(funcDir, 0755)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(funcDir, "index.ts"), []byte(code), 0644)
		require.NoError(t, err)
	}

	files, err := ListFunctionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestListFunctionFiles_FlatFilePriority(t *testing.T) {
	tmpDir := t.TempDir()

	code := `export default async function(req) { return new Response("Hello"); }`

	// Create both flat file and directory for same function
	err := os.WriteFile(filepath.Join(tmpDir, "myfunc.ts"), []byte(code), 0644)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tmpDir, "myfunc"), 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "myfunc", "index.ts"), []byte(code), 0644)
	require.NoError(t, err)

	files, err := ListFunctionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Should be flat file path
	assert.Contains(t, files[0].Path, "myfunc.ts")
	assert.NotContains(t, files[0].Path, "index.ts")
}

func TestListFunctionFiles_SkipsInvalid(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid function
	code := `export default async function(req) { return new Response("Hello"); }`
	err := os.WriteFile(filepath.Join(tmpDir, "valid-func.ts"), []byte(code), 0644)
	require.NoError(t, err)

	// Create files that should be skipped
	err = os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# readme"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, ".gitignore"), []byte("node_modules"), 0644)
	require.NoError(t, err)

	files, err := ListFunctionFiles(tmpDir)
	require.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, "valid-func", files[0].Name)
}

func TestListFunctionFiles_NonExistentDirectory(t *testing.T) {
	files, err := ListFunctionFiles("/nonexistent/path/xyz")
	require.NoError(t, err)
	assert.Empty(t, files)
}

// =============================================================================
// LoadSharedModulesFromFilesystem Tests
// =============================================================================

func TestLoadSharedModulesFromFilesystem_NoSharedDir(t *testing.T) {
	tmpDir := t.TempDir()

	modules, err := LoadSharedModulesFromFilesystem(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, modules)
}

func TestLoadSharedModulesFromFilesystem_WithModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create _shared directory
	sharedDir := filepath.Join(tmpDir, "_shared")
	err := os.MkdirAll(sharedDir, 0755)
	require.NoError(t, err)

	// Create shared modules
	corsCode := `export const corsHeaders = { "Access-Control-Allow-Origin": "*" };`
	utilsCode := `export function helper() { return "helper"; }`

	err = os.WriteFile(filepath.Join(sharedDir, "cors.ts"), []byte(corsCode), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(sharedDir, "utils.ts"), []byte(utilsCode), 0644)
	require.NoError(t, err)

	modules, err := LoadSharedModulesFromFilesystem(tmpDir)
	require.NoError(t, err)
	assert.Len(t, modules, 2)
	assert.Equal(t, corsCode, modules["_shared/cors.ts"])
	assert.Equal(t, utilsCode, modules["_shared/utils.ts"])
}

func TestLoadSharedModulesFromFilesystem_NestedModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create _shared directory with nested structure
	sharedDir := filepath.Join(tmpDir, "_shared")
	err := os.MkdirAll(filepath.Join(sharedDir, "helpers"), 0755)
	require.NoError(t, err)

	utilsCode := `export function helper() { return "helper"; }`
	err = os.WriteFile(filepath.Join(sharedDir, "helpers", "utils.ts"), []byte(utilsCode), 0644)
	require.NoError(t, err)

	modules, err := LoadSharedModulesFromFilesystem(tmpDir)
	require.NoError(t, err)
	assert.Len(t, modules, 1)
	assert.Equal(t, utilsCode, modules["_shared/helpers/utils.ts"])
}
