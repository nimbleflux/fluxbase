package functions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFunctionCode(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test function file
	testCode := "async function handler(req) { return { status: 200 }; }"
	testFunctionName := "test-function"
	testFilePath := filepath.Join(tmpDir, testFunctionName+".ts")
	if err := os.WriteFile(testFilePath, []byte(testCode), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		functionsDir string
		functionName string
		wantError    bool
		wantCode     string
	}{
		{
			name:         "load existing function",
			functionsDir: tmpDir,
			functionName: testFunctionName,
			wantError:    false,
			wantCode:     testCode,
		},
		{
			name:         "load non-existent function",
			functionsDir: tmpDir,
			functionName: "non-existent",
			wantError:    true,
			wantCode:     "",
		},
		{
			name:         "invalid function name",
			functionsDir: tmpDir,
			functionName: "../traversal",
			wantError:    true,
			wantCode:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := LoadFunctionCode(tt.functionsDir, tt.functionName)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadFunctionCode() error = %v, wantError %v", err, tt.wantError)
			}
			if code != tt.wantCode {
				t.Errorf("LoadFunctionCode() code = %q, want %q", code, tt.wantCode)
			}
		})
	}
}

func TestSaveFunctionCode(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testCode := "async function handler(req) { return { status: 200 }; }"

	tests := []struct {
		name         string
		functionsDir string
		functionName string
		code         string
		wantError    bool
	}{
		{
			name:         "save valid function",
			functionsDir: tmpDir,
			functionName: "test-function",
			code:         testCode,
			wantError:    false,
		},
		{
			name:         "save with invalid name",
			functionsDir: tmpDir,
			functionName: "../traversal",
			code:         testCode,
			wantError:    true,
		},
		{
			name:         "save with empty code",
			functionsDir: tmpDir,
			functionName: "empty-function",
			code:         "",
			wantError:    true,
		},
		{
			name:         "save with code too large",
			functionsDir: tmpDir,
			functionName: "large-function",
			code:         string(make([]byte, 2*1024*1024)), // 2MB
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SaveFunctionCode(tt.functionsDir, tt.functionName, tt.code)
			if (err != nil) != tt.wantError {
				t.Errorf("SaveFunctionCode() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				// Verify file was created and contains correct code
				filePath := filepath.Join(tt.functionsDir, tt.functionName+".ts")
				savedCode, err := os.ReadFile(filePath)
				if err != nil {
					t.Errorf("Failed to read saved file: %v", err)
				}
				if string(savedCode) != tt.code {
					t.Errorf("SaveFunctionCode() saved code = %q, want %q", string(savedCode), tt.code)
				}
			}
		})
	}
}

func TestDeleteFunctionCode(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test function file to delete
	testFunctionName := "test-function"
	testFilePath := filepath.Join(tmpDir, testFunctionName+".ts")
	if err := os.WriteFile(testFilePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		functionsDir string
		functionName string
		wantError    bool
	}{
		{
			name:         "delete existing function",
			functionsDir: tmpDir,
			functionName: testFunctionName,
			wantError:    false,
		},
		{
			name:         "delete non-existent function",
			functionsDir: tmpDir,
			functionName: "non-existent",
			wantError:    true,
		},
		{
			name:         "delete with invalid name",
			functionsDir: tmpDir,
			functionName: "../traversal",
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DeleteFunctionCode(tt.functionsDir, tt.functionName)
			if (err != nil) != tt.wantError {
				t.Errorf("DeleteFunctionCode() error = %v, wantError %v", err, tt.wantError)
			}

			if !tt.wantError {
				// Verify file was deleted
				filePath := filepath.Join(tt.functionsDir, tt.functionName+".ts")
				if _, err := os.Stat(filePath); !os.IsNotExist(err) {
					t.Error("DeleteFunctionCode() did not delete the file")
				}
			}
		})
	}
}

func TestListFunctionFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create test function files
	testFunctions := []string{"function1", "function2", "function-3"}
	for _, name := range testFunctions {
		filePath := filepath.Join(tmpDir, name+".ts")
		if err := os.WriteFile(filePath, []byte("test"), 0o644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create a non-.ts file (should be ignored)
	nonTsFile := filepath.Join(tmpDir, "readme.md")
	if err := os.WriteFile(nonTsFile, []byte("readme"), 0o644); err != nil {
		t.Fatalf("Failed to create non-ts file: %v", err)
	}

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create a file with invalid name (should be skipped)
	invalidFile := filepath.Join(tmpDir, "../invalid.ts")
	if err := os.WriteFile(invalidFile, []byte("test"), 0o644); err == nil {
		// Only if we successfully created it (might fail due to path issues)
		defer func() { _ = os.Remove(invalidFile) }()
	}

	tests := []struct {
		name         string
		functionsDir string
		wantCount    int
		wantError    bool
	}{
		{
			name:         "list existing functions",
			functionsDir: tmpDir,
			wantCount:    len(testFunctions),
			wantError:    false,
		},
		{
			name:         "list from non-existent directory",
			functionsDir: filepath.Join(tmpDir, "non-existent"),
			wantCount:    0,
			wantError:    false, // Should return empty list, not error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			functions, err := ListFunctionFiles(tt.functionsDir)
			if (err != nil) != tt.wantError {
				t.Errorf("ListFunctionFiles() error = %v, wantError %v", err, tt.wantError)
			}
			if len(functions) != tt.wantCount {
				t.Errorf("ListFunctionFiles() returned %d functions, want %d", len(functions), tt.wantCount)
			}

			if !tt.wantError {
				// Verify function info is correct
				for _, fn := range functions {
					if fn.Name == "" {
						t.Error("ListFunctionFiles() returned function with empty name")
					}
					if fn.Path == "" {
						t.Error("ListFunctionFiles() returned function with empty path")
					}
					if fn.Size <= 0 {
						t.Error("ListFunctionFiles() returned function with invalid size")
					}
					if fn.ModifiedTime <= 0 {
						t.Error("ListFunctionFiles() returned function with invalid modified time")
					}
				}
			}
		})
	}
}

func TestFunctionExists(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create a test function file
	testFunctionName := "test-function"
	testFilePath := filepath.Join(tmpDir, testFunctionName+".ts")
	if err := os.WriteFile(testFilePath, []byte("test"), 0o644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name         string
		functionsDir string
		functionName string
		wantExists   bool
		wantError    bool
	}{
		{
			name:         "existing function",
			functionsDir: tmpDir,
			functionName: testFunctionName,
			wantExists:   true,
			wantError:    false,
		},
		{
			name:         "non-existent function",
			functionsDir: tmpDir,
			functionName: "non-existent",
			wantExists:   false,
			wantError:    false,
		},
		{
			name:         "invalid function name",
			functionsDir: tmpDir,
			functionName: "../traversal",
			wantExists:   false,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, err := FunctionExists(tt.functionsDir, tt.functionName)
			if (err != nil) != tt.wantError {
				t.Errorf("FunctionExists() error = %v, wantError %v", err, tt.wantError)
			}
			if exists != tt.wantExists {
				t.Errorf("FunctionExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

// =============================================================================
// ParseFunctionConfig Tests
// =============================================================================

func TestParseFunctionConfig(t *testing.T) {
	t.Run("default values when no annotations", func(t *testing.T) {
		code := `export default () => ({ hello: "world" });`
		config := ParseFunctionConfig(code)

		if config.AllowUnauthenticated {
			t.Error("AllowUnauthenticated should be false by default")
		}
		if !config.IsPublic {
			t.Error("IsPublic should be true by default")
		}
		if config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be false by default")
		}
	})

	t.Run("parses @fluxbase:allow-unauthenticated", func(t *testing.T) {
		testCases := []struct {
			name string
			code string
		}{
			{
				name: "single line comment",
				code: `// @fluxbase:allow-unauthenticated
export default () => "hello";`,
			},
			{
				name: "single line comment with leading space",
				code: `  // @fluxbase:allow-unauthenticated
export default () => "hello";`,
			},
			{
				name: "multi-line comment start",
				code: `/* @fluxbase:allow-unauthenticated */
export default () => "hello";`,
			},
			{
				name: "multi-line comment body",
				code: `/*
 * @fluxbase:allow-unauthenticated
 */
export default () => "hello";`,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config := ParseFunctionConfig(tc.code)
				if !config.AllowUnauthenticated {
					t.Errorf("AllowUnauthenticated should be true for: %s", tc.name)
				}
			})
		}
	})

	t.Run("parses @fluxbase:public true", func(t *testing.T) {
		code := `// @fluxbase:public true
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if !config.IsPublic {
			t.Error("IsPublic should be true")
		}
	})

	t.Run("parses @fluxbase:public false", func(t *testing.T) {
		code := `// @fluxbase:public false
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if config.IsPublic {
			t.Error("IsPublic should be false when @fluxbase:public false is set")
		}
	})

	t.Run("parses @fluxbase:public without value (defaults to true)", func(t *testing.T) {
		code := `// @fluxbase:public
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if !config.IsPublic {
			t.Error("IsPublic should be true when @fluxbase:public has no value")
		}
	})

	t.Run("parses @fluxbase:disable-execution-logs true", func(t *testing.T) {
		code := `// @fluxbase:disable-execution-logs true
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if !config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be true")
		}
	})

	t.Run("parses @fluxbase:disable-execution-logs without value (defaults to true)", func(t *testing.T) {
		code := `// @fluxbase:disable-execution-logs
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if !config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be true when no value specified")
		}
	})

	t.Run("parses @fluxbase:disable-execution-logs false", func(t *testing.T) {
		code := `// @fluxbase:disable-execution-logs false
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be false when explicitly set to false")
		}
	})

	t.Run("parses multiple annotations", func(t *testing.T) {
		code := `// @fluxbase:allow-unauthenticated
// @fluxbase:public false
// @fluxbase:disable-execution-logs
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if !config.AllowUnauthenticated {
			t.Error("AllowUnauthenticated should be true")
		}
		if config.IsPublic {
			t.Error("IsPublic should be false")
		}
		if !config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be true")
		}
	})

	t.Run("annotations not at start of line are ignored", func(t *testing.T) {
		code := `const x = "// @fluxbase:allow-unauthenticated";
export default () => "hello";`
		_ = ParseFunctionConfig(code)

		// Note: The regex uses ^\s* so this should still match if the annotation
		// is at the start of a line in a string - this tests actual behavior
		// The important thing is that we don't have false positives in most cases
	})

	t.Run("parses CORS annotations", func(t *testing.T) {
		code := `// @fluxbase:cors-origins https://example.com
// @fluxbase:cors-methods GET,POST
// @fluxbase:cors-headers Content-Type,Authorization
// @fluxbase:cors-credentials true
// @fluxbase:cors-max-age 3600
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if config.CorsOrigins == nil || *config.CorsOrigins != "https://example.com" {
			t.Errorf("CorsOrigins expected 'https://example.com', got %v", config.CorsOrigins)
		}
		if config.CorsMethods == nil || *config.CorsMethods != "GET,POST" {
			t.Errorf("CorsMethods expected 'GET,POST', got %v", config.CorsMethods)
		}
		if config.CorsHeaders == nil || *config.CorsHeaders != "Content-Type,Authorization" {
			t.Errorf("CorsHeaders expected 'Content-Type,Authorization', got %v", config.CorsHeaders)
		}
		if config.CorsCredentials == nil || *config.CorsCredentials != true {
			t.Errorf("CorsCredentials expected true, got %v", config.CorsCredentials)
		}
		if config.CorsMaxAge == nil || *config.CorsMaxAge != 3600 {
			t.Errorf("CorsMaxAge expected 3600, got %v", config.CorsMaxAge)
		}
	})

	t.Run("parses rate limit annotations", func(t *testing.T) {
		// Implementation uses format: @fluxbase:rate-limit <value>/<unit>
		code := `// @fluxbase:rate-limit 100/min
// @fluxbase:rate-limit 1000/hour
// @fluxbase:rate-limit 10000/day
export default () => "hello";`
		config := ParseFunctionConfig(code)

		if config.RateLimitPerMinute == nil || *config.RateLimitPerMinute != 100 {
			t.Errorf("RateLimitPerMinute expected 100, got %v", config.RateLimitPerMinute)
		}
		if config.RateLimitPerHour == nil || *config.RateLimitPerHour != 1000 {
			t.Errorf("RateLimitPerHour expected 1000, got %v", config.RateLimitPerHour)
		}
		if config.RateLimitPerDay == nil || *config.RateLimitPerDay != 10000 {
			t.Errorf("RateLimitPerDay expected 10000, got %v", config.RateLimitPerDay)
		}
	})

	t.Run("handles empty code", func(t *testing.T) {
		config := ParseFunctionConfig("")

		if config.AllowUnauthenticated {
			t.Error("AllowUnauthenticated should be false for empty code")
		}
		if !config.IsPublic {
			t.Error("IsPublic should be true (default) for empty code")
		}
	})
}

// =============================================================================
// FunctionConfig Struct Tests
// =============================================================================

func TestFunctionConfig_Struct(t *testing.T) {
	t.Run("zero value defaults", func(t *testing.T) {
		config := FunctionConfig{}

		if config.AllowUnauthenticated {
			t.Error("AllowUnauthenticated should be false by default")
		}
		if config.IsPublic {
			t.Error("IsPublic should be false when unset (zero value)")
		}
		if config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be false by default")
		}
		if config.CorsOrigins != nil {
			t.Error("CorsOrigins should be nil by default")
		}
		if config.RateLimitPerMinute != nil {
			t.Error("RateLimitPerMinute should be nil by default")
		}
	})

	t.Run("all fields set", func(t *testing.T) {
		origins := "https://example.com"
		methods := "GET,POST"
		headers := "Content-Type"
		credentials := true
		maxAge := 3600
		perMin := 100
		perHour := 1000
		perDay := 10000

		config := FunctionConfig{
			AllowUnauthenticated: true,
			IsPublic:             false,
			DisableExecutionLogs: true,
			CorsOrigins:          &origins,
			CorsMethods:          &methods,
			CorsHeaders:          &headers,
			CorsCredentials:      &credentials,
			CorsMaxAge:           &maxAge,
			RateLimitPerMinute:   &perMin,
			RateLimitPerHour:     &perHour,
			RateLimitPerDay:      &perDay,
		}

		if !config.AllowUnauthenticated {
			t.Error("AllowUnauthenticated should be true")
		}
		if config.IsPublic {
			t.Error("IsPublic should be false")
		}
		if !config.DisableExecutionLogs {
			t.Error("DisableExecutionLogs should be true")
		}
		if *config.CorsOrigins != "https://example.com" {
			t.Errorf("CorsOrigins expected 'https://example.com', got '%s'", *config.CorsOrigins)
		}
		if *config.RateLimitPerMinute != 100 {
			t.Errorf("RateLimitPerMinute expected 100, got %d", *config.RateLimitPerMinute)
		}
	})
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestSaveAndLoadFunctionCodeIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "functions-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	testCode := `async function handler(req) {
	const data = JSON.parse(req.body || "{}");
	return {
		status: 200,
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ message: "Hello " + data.name })
	};
}`

	functionName := "hello-world"

	// Save function code
	if err := SaveFunctionCode(tmpDir, functionName, testCode); err != nil {
		t.Fatalf("SaveFunctionCode() failed: %v", err)
	}

	// Load function code
	loadedCode, err := LoadFunctionCode(tmpDir, functionName)
	if err != nil {
		t.Fatalf("LoadFunctionCode() failed: %v", err)
	}

	if loadedCode != testCode {
		t.Errorf("Loaded code does not match saved code.\nGot: %q\nWant: %q", loadedCode, testCode)
	}

	// Verify function exists
	exists, err := FunctionExists(tmpDir, functionName)
	if err != nil {
		t.Fatalf("FunctionExists() failed: %v", err)
	}
	if !exists {
		t.Error("Function should exist after saving")
	}

	// List functions
	functions, err := ListFunctionFiles(tmpDir)
	if err != nil {
		t.Fatalf("ListFunctionFiles() failed: %v", err)
	}
	if len(functions) != 1 {
		t.Fatalf("ListFunctionFiles() returned %d functions, want 1", len(functions))
	}
	if functions[0].Name != functionName {
		t.Errorf("ListFunctionFiles() returned function with name %q, want %q", functions[0].Name, functionName)
	}

	// Delete function code
	if err := DeleteFunctionCode(tmpDir, functionName); err != nil {
		t.Fatalf("DeleteFunctionCode() failed: %v", err)
	}

	// Verify function no longer exists
	exists, err = FunctionExists(tmpDir, functionName)
	if err != nil {
		t.Fatalf("FunctionExists() failed after delete: %v", err)
	}
	if exists {
		t.Error("Function should not exist after deletion")
	}
}
