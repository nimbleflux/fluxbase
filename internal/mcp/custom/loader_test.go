package custom

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewLoader Tests
// =============================================================================

func TestNewLoader(t *testing.T) {
	t.Run("creates loader with directory", func(t *testing.T) {
		loader := NewLoader("/path/to/tools")

		require.NotNil(t, loader)
		assert.Equal(t, "/path/to/tools", loader.toolsDir)
	})

	t.Run("creates loader with empty directory", func(t *testing.T) {
		loader := NewLoader("")

		require.NotNil(t, loader)
		assert.Equal(t, "", loader.toolsDir)
	})
}

// =============================================================================
// Loader Struct Tests
// =============================================================================

func TestLoader_Struct(t *testing.T) {
	t.Run("stores tools directory", func(t *testing.T) {
		loader := &Loader{
			toolsDir: "/custom/tools",
		}

		assert.Equal(t, "/custom/tools", loader.toolsDir)
	})
}

// =============================================================================
// LoadedTool Struct Tests
// =============================================================================

func TestLoadedTool_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		tool := &LoadedTool{
			Name:           "my_tool",
			Namespace:      "admin",
			Description:    "A test tool",
			Code:           "console.log('hello');",
			TimeoutSeconds: 60,
			MemoryLimitMB:  256,
			AllowNet:       true,
			AllowEnv:       true,
			AllowRead:      true,
			AllowWrite:     true,
			RequiredScopes: []string{"admin", "write"},
		}

		assert.Equal(t, "my_tool", tool.Name)
		assert.Equal(t, "admin", tool.Namespace)
		assert.Equal(t, "A test tool", tool.Description)
		assert.Contains(t, tool.Code, "console.log")
		assert.Equal(t, 60, tool.TimeoutSeconds)
		assert.Equal(t, 256, tool.MemoryLimitMB)
		assert.True(t, tool.AllowNet)
		assert.True(t, tool.AllowEnv)
		assert.True(t, tool.AllowRead)
		assert.True(t, tool.AllowWrite)
		assert.Equal(t, []string{"admin", "write"}, tool.RequiredScopes)
	})

	t.Run("default values are zero/false/nil", func(t *testing.T) {
		tool := &LoadedTool{}

		assert.Empty(t, tool.Name)
		assert.Empty(t, tool.Namespace)
		assert.Empty(t, tool.Description)
		assert.Empty(t, tool.Code)
		assert.Equal(t, 0, tool.TimeoutSeconds)
		assert.Equal(t, 0, tool.MemoryLimitMB)
		assert.False(t, tool.AllowNet)
		assert.False(t, tool.AllowEnv)
		assert.False(t, tool.AllowRead)
		assert.False(t, tool.AllowWrite)
		assert.Nil(t, tool.RequiredScopes)
	})
}

// =============================================================================
// LoadAll Tests
// =============================================================================

func TestLoader_LoadAll(t *testing.T) {
	t.Run("returns nil for non-existent directory", func(t *testing.T) {
		loader := NewLoader("/non/existent/directory")

		tools, err := loader.LoadAll()

		assert.NoError(t, err)
		assert.Nil(t, tools)
	})

	t.Run("loads .ts files from directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a .ts file
		content := `// @fluxbase:name test_tool
// @fluxbase:description A test tool
console.log('test');`
		err := os.WriteFile(filepath.Join(tmpDir, "test.ts"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tools, err := loader.LoadAll()

		require.NoError(t, err)
		require.Len(t, tools, 1)
		assert.Equal(t, "test_tool", tools[0].Name)
		assert.Equal(t, "A test tool", tools[0].Description)
	})

	t.Run("loads .js files from directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		content := `// @fluxbase:name js_tool
console.log('js');`
		err := os.WriteFile(filepath.Join(tmpDir, "tool.js"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tools, err := loader.LoadAll()

		require.NoError(t, err)
		require.Len(t, tools, 1)
		assert.Equal(t, "js_tool", tools[0].Name)
	})

	t.Run("ignores non-.ts/.js files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create various file types
		err := os.WriteFile(filepath.Join(tmpDir, "tool.ts"), []byte("// TS file"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("readme"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "config.json"), []byte("{}"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "script.sh"), []byte("#!/bin/bash"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tools, err := loader.LoadAll()

		require.NoError(t, err)
		assert.Len(t, tools, 1)
	})

	t.Run("ignores directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a subdirectory with .ts name (weird but should be ignored)
		err := os.Mkdir(filepath.Join(tmpDir, "subdir"), 0o755)
		require.NoError(t, err)

		// Create a valid tool
		err = os.WriteFile(filepath.Join(tmpDir, "tool.ts"), []byte("// tool"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tools, err := loader.LoadAll()

		require.NoError(t, err)
		assert.Len(t, tools, 1)
	})

	t.Run("loads multiple tools", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create multiple tool files
		err := os.WriteFile(filepath.Join(tmpDir, "tool1.ts"), []byte("// @fluxbase:name tool1"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "tool2.ts"), []byte("// @fluxbase:name tool2"), 0o644)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "tool3.js"), []byte("// @fluxbase:name tool3"), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tools, err := loader.LoadAll()

		require.NoError(t, err)
		assert.Len(t, tools, 3)
	})
}

// =============================================================================
// parseAnnotations Tests
// =============================================================================

func TestParseAnnotations(t *testing.T) {
	t.Run("parses name from filename", func(t *testing.T) {
		code := "console.log('test');"
		name, _ := parseAnnotations(code, "my-tool.ts")

		assert.Equal(t, "my_tool", name)
	})

	t.Run("removes .ts extension", func(t *testing.T) {
		code := ""
		name, _ := parseAnnotations(code, "some-tool.ts")

		assert.Equal(t, "some_tool", name)
	})

	t.Run("removes .js extension", func(t *testing.T) {
		code := ""
		name, _ := parseAnnotations(code, "other-tool.js")

		assert.Equal(t, "other_tool", name)
	})

	t.Run("replaces dashes with underscores in name", func(t *testing.T) {
		code := ""
		name, _ := parseAnnotations(code, "my-cool-tool.ts")

		assert.Equal(t, "my_cool_tool", name)
	})

	t.Run("parses @fluxbase:name annotation", func(t *testing.T) {
		code := `// @fluxbase:name custom_name
console.log('test');`
		name, _ := parseAnnotations(code, "original.ts")

		assert.Equal(t, "custom_name", name)
	})

	t.Run("parses @fluxbase:description annotation", func(t *testing.T) {
		code := `// @fluxbase:description This is a test tool
console.log('test');`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, "This is a test tool", annotations["description"])
	})

	t.Run("parses @fluxbase:namespace annotation", func(t *testing.T) {
		code := `// @fluxbase:namespace admin`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, "admin", annotations["namespace"])
	})

	t.Run("parses @fluxbase:timeout annotation as int", func(t *testing.T) {
		code := `// @fluxbase:timeout 120`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, 120, annotations["timeout"])
	})

	t.Run("parses @fluxbase:memory annotation as int", func(t *testing.T) {
		code := `// @fluxbase:memory 512`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, 512, annotations["memory"])
	})

	t.Run("parses @fluxbase:allow-net annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-net`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, true, annotations["allow-net"])
	})

	t.Run("parses @fluxbase:allow-env annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-env`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, true, annotations["allow-env"])
	})

	t.Run("parses @fluxbase:allow-read annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-read`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, true, annotations["allow-read"])
	})

	t.Run("parses @fluxbase:allow-write annotation", func(t *testing.T) {
		code := `// @fluxbase:allow-write`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, true, annotations["allow-write"])
	})

	t.Run("parses @fluxbase:scopes annotation", func(t *testing.T) {
		code := `// @fluxbase:scopes admin, write, read`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, "admin, write, read", annotations["scopes"])
	})

	t.Run("parses multiple annotations", func(t *testing.T) {
		code := `// @fluxbase:name my_tool
// @fluxbase:description A cool tool
// @fluxbase:namespace admin
// @fluxbase:timeout 60
// @fluxbase:allow-net
// @fluxbase:scopes admin, write

function doStuff() {
  console.log('stuff');
}`
		name, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, "my_tool", name)
		assert.Equal(t, "A cool tool", annotations["description"])
		assert.Equal(t, "admin", annotations["namespace"])
		assert.Equal(t, 60, annotations["timeout"])
		assert.Equal(t, true, annotations["allow-net"])
		assert.Equal(t, "admin, write", annotations["scopes"])
	})

	t.Run("ignores non-fluxbase annotations", func(t *testing.T) {
		code := `// @param name - The name
// @returns The result
// Regular comment`
		_, annotations := parseAnnotations(code, "tool.ts")

		assert.Empty(t, annotations)
	})

	t.Run("ignores non-comment lines", func(t *testing.T) {
		code := `const x = "@fluxbase:name fake";`
		name, annotations := parseAnnotations(code, "tool.ts")

		assert.Equal(t, "tool", name) // Default from filename
		assert.Empty(t, annotations)
	})
}

// =============================================================================
// loadTool Tests
// =============================================================================

func TestLoader_loadTool(t *testing.T) {
	t.Run("loads tool with defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `console.log('hello');`
		err := os.WriteFile(filepath.Join(tmpDir, "basic.ts"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tool, err := loader.loadTool("basic.ts")

		require.NoError(t, err)
		assert.Equal(t, "basic", tool.Name)
		assert.Equal(t, "default", tool.Namespace)
		assert.Equal(t, 30, tool.TimeoutSeconds)
		assert.Equal(t, 128, tool.MemoryLimitMB)
		assert.True(t, tool.AllowNet) // Default is true
		assert.False(t, tool.AllowEnv)
		assert.False(t, tool.AllowRead)
		assert.False(t, tool.AllowWrite)
	})

	t.Run("loads tool with annotations", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `// @fluxbase:name custom_tool
// @fluxbase:namespace admin
// @fluxbase:description Test description
// @fluxbase:timeout 120
// @fluxbase:memory 256
// @fluxbase:allow-net
// @fluxbase:allow-env
// @fluxbase:allow-read
// @fluxbase:allow-write
// @fluxbase:scopes admin, write

console.log('configured');`
		err := os.WriteFile(filepath.Join(tmpDir, "configured.ts"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tool, err := loader.loadTool("configured.ts")

		require.NoError(t, err)
		assert.Equal(t, "custom_tool", tool.Name)
		assert.Equal(t, "admin", tool.Namespace)
		assert.Equal(t, "Test description", tool.Description)
		assert.Equal(t, 120, tool.TimeoutSeconds)
		assert.Equal(t, 256, tool.MemoryLimitMB)
		assert.True(t, tool.AllowNet)
		assert.True(t, tool.AllowEnv)
		assert.True(t, tool.AllowRead)
		assert.True(t, tool.AllowWrite)
		assert.Equal(t, []string{"admin", "write"}, tool.RequiredScopes)
	})

	t.Run("parses scopes with spaces", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `// @fluxbase:scopes admin,  write ,  read`
		err := os.WriteFile(filepath.Join(tmpDir, "scopes.ts"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tool, err := loader.loadTool("scopes.ts")

		require.NoError(t, err)
		// Scopes are split and trimmed
		assert.Equal(t, []string{"admin", "write", "read"}, tool.RequiredScopes)
	})

	t.Run("stores code content", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `// @fluxbase:name mytool
function test() {
  return "hello";
}`
		err := os.WriteFile(filepath.Join(tmpDir, "code.ts"), []byte(content), 0o644)
		require.NoError(t, err)

		loader := NewLoader(tmpDir)

		tool, err := loader.loadTool("code.ts")

		require.NoError(t, err)
		assert.Contains(t, tool.Code, "function test()")
		assert.Contains(t, tool.Code, "@fluxbase:name")
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		loader := NewLoader(tmpDir)

		_, err := loader.loadTool("missing.ts")

		assert.Error(t, err)
	})
}

// =============================================================================
// File Extension Handling Tests
// =============================================================================

func TestFileExtensionHandling(t *testing.T) {
	t.Run(".ts extension is valid", func(t *testing.T) {
		filename := "tool.ts"
		isValid := strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".js")

		assert.True(t, isValid)
	})

	t.Run(".js extension is valid", func(t *testing.T) {
		filename := "tool.js"
		isValid := strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".js")

		assert.True(t, isValid)
	})

	t.Run(".tsx extension is not valid", func(t *testing.T) {
		filename := "tool.tsx"
		isValid := strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".js")

		assert.False(t, isValid)
	})

	t.Run(".py extension is not valid", func(t *testing.T) {
		filename := "tool.py"
		isValid := strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".js")

		assert.False(t, isValid)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkParseAnnotations(b *testing.B) {
	code := `// @fluxbase:name my_tool
// @fluxbase:description A comprehensive tool for testing
// @fluxbase:namespace admin
// @fluxbase:timeout 60
// @fluxbase:memory 256
// @fluxbase:allow-net
// @fluxbase:allow-env
// @fluxbase:scopes admin, write, read

function execute(args) {
  console.log('Executing with', args);
  return { success: true };
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseAnnotations(code, "benchmark.ts")
	}
}

func BenchmarkNewLoader(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewLoader("/path/to/tools")
	}
}
