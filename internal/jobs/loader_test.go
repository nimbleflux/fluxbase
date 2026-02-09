package jobs

import (
	"os"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
)

func TestParseAnnotations_ProgressTimeout(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "default when no annotation",
			code:     "export function handler() {}",
			expected: 300, // Default after fix
		},
		{
			name:     "explicit 300",
			code:     "// @fluxbase:progress-timeout 300\nexport function handler() {}",
			expected: 300,
		},
		{
			name:     "explicit 600",
			code:     "// @fluxbase:progress-timeout 600\nexport function handler() {}",
			expected: 600,
		},
		{
			name:     "explicit 60",
			code:     "// @fluxbase:progress-timeout 60\nexport function handler() {}",
			expected: 60,
		},
		{
			name:     "with tab instead of space",
			code:     "// @fluxbase:progress-timeout\t120\nexport function handler() {}",
			expected: 120,
		},
		{
			name:     "with multiple spaces",
			code:     "// @fluxbase:progress-timeout   180\nexport function handler() {}",
			expected: 180,
		},
		{
			name:     "in multiline comment",
			code:     "/* @fluxbase:progress-timeout 240 */\nexport function handler() {}",
			expected: 240,
		},
		{
			name:     "annotation in middle of file",
			code:     "// Some comment\n// @fluxbase:progress-timeout 500\nexport function handler() {}",
			expected: 500,
		},
		{
			name:     "wrong format - no space",
			code:     "// @fluxbase:progress-timeout300\nexport function handler() {}",
			expected: 300, // Should fall back to default
		},
		{
			name:     "wrong format - with colon before number",
			code:     "// @fluxbase:progress-timeout:300\nexport function handler() {}",
			expected: 300, // Should fall back to default
		},
		{
			name:     "wrong format - with s suffix",
			code:     "// @fluxbase:progress-timeout 300s\nexport function handler() {}",
			expected: 300, // Should parse 300 (regex captures digits before 's')
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.ProgressTimeoutSeconds != tt.expected {
				t.Errorf("ProgressTimeoutSeconds = %d, want %d", annotations.ProgressTimeoutSeconds, tt.expected)
			}
		})
	}
}

func TestParseAnnotations_Timeout(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "default when no annotation",
			code:     "export function handler() {}",
			expected: 300,
		},
		{
			name:     "explicit 600",
			code:     "// @fluxbase:timeout 600\nexport function handler() {}",
			expected: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.TimeoutSeconds != tt.expected {
				t.Errorf("TimeoutSeconds = %d, want %d", annotations.TimeoutSeconds, tt.expected)
			}
		})
	}
}

func TestParseAnnotations_MaxRetries(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "default when no annotation",
			code:     "export function handler() {}",
			expected: 0,
		},
		{
			name:     "explicit 3",
			code:     "// @fluxbase:max-retries 3\nexport function handler() {}",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.MaxRetries != tt.expected {
				t.Errorf("MaxRetries = %d, want %d", annotations.MaxRetries, tt.expected)
			}
		})
	}
}

func TestParseAnnotations_Permissions(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		allowNet   bool
		allowEnv   bool
		allowRead  bool
		allowWrite bool
	}{
		{
			name:       "defaults",
			code:       "export function handler() {}",
			allowNet:   true,
			allowEnv:   true,
			allowRead:  false,
			allowWrite: false,
		},
		{
			name:       "allow-read true",
			code:       "// @fluxbase:allow-read true\nexport function handler() {}",
			allowNet:   true,
			allowEnv:   true,
			allowRead:  true,
			allowWrite: false,
		},
		{
			name:       "allow-net false",
			code:       "// @fluxbase:allow-net false\nexport function handler() {}",
			allowNet:   false,
			allowEnv:   true,
			allowRead:  false,
			allowWrite: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.AllowNet != tt.allowNet {
				t.Errorf("AllowNet = %v, want %v", annotations.AllowNet, tt.allowNet)
			}
			if annotations.AllowEnv != tt.allowEnv {
				t.Errorf("AllowEnv = %v, want %v", annotations.AllowEnv, tt.allowEnv)
			}
			if annotations.AllowRead != tt.allowRead {
				t.Errorf("AllowRead = %v, want %v", annotations.AllowRead, tt.allowRead)
			}
			if annotations.AllowWrite != tt.allowWrite {
				t.Errorf("AllowWrite = %v, want %v", annotations.AllowWrite, tt.allowWrite)
			}
		})
	}
}

func TestParseAnnotations_MultipleAnnotations(t *testing.T) {
	code := `// @fluxbase:timeout 600
// @fluxbase:progress-timeout 120
// @fluxbase:max-retries 3
// @fluxbase:memory 512
// @fluxbase:allow-read true
// @fluxbase:allow-net false

export async function handler(request: Request) {
  // job code
}`

	annotations := parseAnnotations(code)

	if annotations.TimeoutSeconds != 600 {
		t.Errorf("TimeoutSeconds = %d, want 600", annotations.TimeoutSeconds)
	}
	if annotations.ProgressTimeoutSeconds != 120 {
		t.Errorf("ProgressTimeoutSeconds = %d, want 120", annotations.ProgressTimeoutSeconds)
	}
	if annotations.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", annotations.MaxRetries)
	}
	if annotations.MemoryLimitMB != 512 {
		t.Errorf("MemoryLimitMB = %d, want 512", annotations.MemoryLimitMB)
	}
	if !annotations.AllowRead {
		t.Error("AllowRead should be true")
	}
	if annotations.AllowNet {
		t.Error("AllowNet should be false")
	}
}

func TestParseAnnotations_Schedule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectSchedule *string
	}{
		{
			name:           "no schedule",
			code:           "export function handler() {}",
			expectSchedule: nil,
		},
		{
			name:           "every 5 minutes",
			code:           "// @fluxbase:schedule */5 * * * *\nexport function handler() {}",
			expectSchedule: strPtr("*/5 * * * *"),
		},
		{
			name:           "daily at midnight",
			code:           "// @fluxbase:schedule 0 0 * * *\nexport function handler() {}",
			expectSchedule: strPtr("0 0 * * *"),
		},
		{
			name:           "every hour",
			code:           "// @fluxbase:schedule 0 * * * *\nexport function handler() {}",
			expectSchedule: strPtr("0 * * * *"),
		},
		{
			name:           "weekly on sunday",
			code:           "// @fluxbase:schedule 0 0 * * 0\nexport function handler() {}",
			expectSchedule: strPtr("0 0 * * 0"),
		},
		{
			name:           "every minute",
			code:           "// @fluxbase:schedule * * * * *\nexport function handler() {}",
			expectSchedule: strPtr("* * * * *"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if tt.expectSchedule == nil {
				if annotations.Schedule != nil {
					t.Errorf("Schedule = %v, want nil", *annotations.Schedule)
				}
			} else {
				if annotations.Schedule == nil {
					t.Errorf("Schedule = nil, want %v", *tt.expectSchedule)
				} else if *annotations.Schedule != *tt.expectSchedule {
					t.Errorf("Schedule = %v, want %v", *annotations.Schedule, *tt.expectSchedule)
				}
			}
		})
	}
}

func TestParseAnnotations_RequireRole(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectRoles []string
	}{
		{
			name:        "no require-role",
			code:        "export function handler() {}",
			expectRoles: nil,
		},
		{
			name:        "require admin",
			code:        "// @fluxbase:require-role admin\nexport function handler() {}",
			expectRoles: []string{"admin"},
		},
		{
			name:        "require authenticated",
			code:        "// @fluxbase:require-role authenticated\nexport function handler() {}",
			expectRoles: []string{"authenticated"},
		},
		{
			name:        "require anon",
			code:        "// @fluxbase:require-role anon\nexport function handler() {}",
			expectRoles: []string{"anon"},
		},
		{
			name:        "require multiple roles",
			code:        "// @fluxbase:require-role admin, editor, moderator\nexport function handler() {}",
			expectRoles: []string{"admin", "editor", "moderator"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if tt.expectRoles == nil {
				if len(annotations.RequireRoles) != 0 {
					t.Errorf("RequireRoles = %v, want nil/empty", annotations.RequireRoles)
				}
			} else {
				if len(annotations.RequireRoles) != len(tt.expectRoles) {
					t.Errorf("RequireRoles = %v, want %v", annotations.RequireRoles, tt.expectRoles)
				} else {
					for i, role := range annotations.RequireRoles {
						if role != tt.expectRoles[i] {
							t.Errorf("RequireRoles[%d] = %v, want %v", i, role, tt.expectRoles[i])
						}
					}
				}
			}
		})
	}
}

func TestParseAnnotations_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		enabled bool
	}{
		{
			name:    "default enabled",
			code:    "export function handler() {}",
			enabled: true,
		},
		{
			name:    "explicitly disabled",
			code:    "// @fluxbase:enabled false\nexport function handler() {}",
			enabled: false,
		},
		{
			name:    "explicitly enabled (redundant but valid)",
			code:    "// @fluxbase:enabled true\nexport function handler() {}",
			enabled: true, // Should remain true (default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.Enabled != tt.enabled {
				t.Errorf("Enabled = %v, want %v", annotations.Enabled, tt.enabled)
			}
		})
	}
}

func TestParseAnnotations_Memory(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{
			name:     "default memory",
			code:     "export function handler() {}",
			expected: 256, // Default
		},
		{
			name:     "explicit 128MB",
			code:     "// @fluxbase:memory 128\nexport function handler() {}",
			expected: 128,
		},
		{
			name:     "explicit 512MB",
			code:     "// @fluxbase:memory 512\nexport function handler() {}",
			expected: 512,
		},
		{
			name:     "explicit 1024MB (1GB)",
			code:     "// @fluxbase:memory 1024\nexport function handler() {}",
			expected: 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			annotations := parseAnnotations(tt.code)
			if annotations.MemoryLimitMB != tt.expected {
				t.Errorf("MemoryLimitMB = %d, want %d", annotations.MemoryLimitMB, tt.expected)
			}
		})
	}
}

func TestParseAnnotations_ScheduleWithParams(t *testing.T) {
	code := `// @fluxbase:schedule 0 2 * * *
// @fluxbase:schedule-params {"type": "daily", "notify": true}
export function handler() {}`

	annotations := parseAnnotations(code)

	if annotations.Schedule == nil {
		t.Fatal("Schedule should not be nil")
	}

	// Schedule should contain the combined format: cron|json
	// Note: JSON marshal may reorder keys, so we check for presence of both parts
	if annotations.Schedule == nil || !contains(*annotations.Schedule, "0 2 * * *") || !contains(*annotations.Schedule, "|") {
		t.Errorf("Schedule = %v, expected to contain cron and pipe separator", *annotations.Schedule)
	}
}

// Helper functions for tests
func strPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestNewLoader tests the Loader constructor
func TestNewLoader(t *testing.T) {
	tests := []struct {
		name         string
		storage      *Storage
		config       *mockJobsConfig
		npmRegistry  string
		jsrRegistry  string
		wantErr      bool
		validateFunc func(*testing.T, *Loader)
	}{
		{
			name:    "creates loader with defaults",
			storage: &Storage{},
			config: &mockJobsConfig{
				jobsDir: "/tmp/jobs",
			},
			wantErr: false,
			validateFunc: func(t *testing.T, l *Loader) {
				if l == nil {
					t.Fatal("Loader should not be nil")
				}
				if l.storage == nil {
					t.Error("storage should not be nil")
				}
				if l.config == nil {
					t.Error("config should not be nil")
				}
				if l.bundler == nil {
					t.Error("bundler should not be nil")
				}
			},
		},
		{
			name:    "creates loader with custom npm registry",
			storage: &Storage{},
			config: &mockJobsConfig{
				jobsDir: "/tmp/jobs",
			},
			npmRegistry: "https://npm.example.com",
			wantErr:     false,
		},
		{
			name:    "creates loader with custom jsr registry",
			storage: &Storage{},
			config: &mockJobsConfig{
				jobsDir: "/tmp/jobs",
			},
			jsrRegistry: "https://jsr.example.com",
			wantErr:     false,
		},
		{
			name:    "creates loader with both registries",
			storage: &Storage{},
			config: &mockJobsConfig{
				jobsDir: "/tmp/jobs",
			},
			npmRegistry: "https://npm.example.com",
			jsrRegistry: "https://jsr.example.com",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create real config from mock
			cfg := &config.JobsConfig{
				JobsDir: tt.config.jobsDir,
			}

			loader, err := NewLoader(tt.storage, cfg, tt.npmRegistry, tt.jsrRegistry)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewLoader() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("NewLoader() unexpected error = %v", err)
			}

			if tt.validateFunc != nil {
				tt.validateFunc(t, loader)
			}
		})
	}
}

// TestLoadNestedFiles tests recursive file loading
func TestLoadNestedFiles(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (rootPath string, cleanup func())
		relativePath string
		want         map[string]string
		wantErr      bool
	}{
		{
			name: "loads flat directory",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "file1.ts", "export function f1() {}")
				writeFile(t, dir, "file2.js", "export function f2() {}")
				return dir, nil
			},
			relativePath: ".",
			want: map[string]string{
				"file1.ts": "export function f1() {}",
				"file2.js": "export function f2() {}",
			},
		},
		{
			name: "loads nested directories",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				nestedDir := createDir(t, dir, "nested")
				writeFile(t, nestedDir, "file1.ts", "export function f1() {}")
				writeFile(t, dir, "file2.ts", "export function f2() {}")
				return dir, nil
			},
			relativePath: ".",
			want: map[string]string{
				"nested/file1.ts": "export function f1() {}",
				"file2.ts":        "export function f2() {}",
			},
		},
		{
			name: "loads deeply nested directories",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				level1 := createDir(t, dir, "level1")
				level2 := createDir(t, level1, "level2")
				writeFile(t, level2, "deep.ts", "export function deep() {}")
				writeFile(t, level1, "mid.ts", "export function mid() {}")
				writeFile(t, dir, "top.ts", "export function top() {}")
				return dir, nil
			},
			relativePath: ".",
			want: map[string]string{
				"level1/level2/deep.ts": "export function deep() {}",
				"level1/mid.ts":         "export function mid() {}",
				"top.ts":                "export function top() {}",
			},
		},
		{
			name: "includes JSON and GeoJSON files",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "data.json", `{"key": "value"}`)
				writeFile(t, dir, "map.geojson", `{"type": "Feature"}`)
				writeFile(t, dir, "code.ts", "export function f() {}")
				return dir, nil
			},
			relativePath: ".",
			want: map[string]string{
				"data.json":   `{"key": "value"}`,
				"map.geojson": `{"type": "Feature"}`,
				"code.ts":     "export function f() {}",
			},
		},
		{
			name: "excludes non-matching files",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "code.ts", "export function f() {}")
				writeFile(t, dir, "data.txt", "text file")
				writeFile(t, dir, "image.png", "binary data")
				writeFile(t, dir, "README.md", "readme")
				return dir, nil
			},
			relativePath: ".",
			want: map[string]string{
				"code.ts": "export function f() {}",
			},
		},
		{
			name: "handles empty directory",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				return dir, nil
			},
			relativePath: ".",
			want:         map[string]string{},
		},
		{
			name: "handles only non-matching files",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "README.md", "readme")
				writeFile(t, dir, "data.txt", "text")
				return dir, nil
			},
			relativePath: ".",
			want:         map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath, cleanup := tt.setup(t)
			if cleanup != nil {
				defer cleanup()
			}

			loader := &Loader{}
			got, err := loader.loadNestedFiles(rootPath, tt.relativePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadNestedFiles() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadNestedFiles() unexpected error = %v", err)
			}

			if len(got) != len(tt.want) {
				t.Errorf("loadNestedFiles() returned %d files, want %d", len(got), len(tt.want))
			}

			for path, content := range tt.want {
				if gotContent, exists := got[path]; !exists {
					t.Errorf("loadNestedFiles() missing file %q", path)
				} else if gotContent != content {
					t.Errorf("loadNestedFiles() file %q content mismatch\n got: %s\n want: %s", path, gotContent, content)
				}
			}
		})
	}
}

// TestLoadSharedModules tests loading of shared modules from _shared directory
func TestLoadSharedModules(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (jobsDir string)
		want    int // number of shared modules
		wantErr bool
	}{
		{
			name: "loads modules from _shared directory",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				sharedDir := createDir(t, jobsDir, "_shared")
				writeFile(t, sharedDir, "utils.ts", "export function util() {}")
				writeFile(t, sharedDir, "constants.ts", "export const CONST = 1")
				return jobsDir
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "loads nested shared modules",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				sharedDir := createDir(t, jobsDir, "_shared")
				nestedDir := createDir(t, sharedDir, "helpers")
				writeFile(t, nestedDir, "helper.ts", "export function help() {}")
				writeFile(t, sharedDir, "main.ts", "export function main() {}")
				return jobsDir
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "returns empty map when _shared does not exist",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				return jobsDir
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "includes JSON files in shared modules",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				sharedDir := createDir(t, jobsDir, "_shared")
				writeFile(t, sharedDir, "data.json", `{"key": "value"}`)
				writeFile(t, sharedDir, "code.ts", "export function f() {}")
				return jobsDir
			},
			want:    2,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobsDir := tt.setup(t)

			loader := &Loader{}
			got, err := loader.loadSharedModules(jobsDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadSharedModules() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadSharedModules() unexpected error = %v", err)
			}

			if len(got) != tt.want {
				t.Errorf("loadSharedModules() returned %d modules, want %d", len(got), tt.want)
			}

			// Verify all files have _shared prefix
			for path := range got {
				if path[:7] != "_shared" {
					t.Errorf("loadSharedModules() file path %q should start with '_shared'", path)
				}
			}
		})
	}
}

// TestLoadJobCode tests loading job code from files and directories
func TestLoadJobCode(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) (jobsDir string, entryName string)
		wantCode  string
		wantFiles map[string]string
		wantErr   bool
	}{
		{
			name: "loads flat TypeScript file",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				code := "export function handler() {}"
				writeFile(t, jobsDir, "my-job.ts", code)
				return jobsDir, "my-job.ts"
			},
			wantCode:  "export function handler() {}",
			wantFiles: map[string]string{},
		},
		{
			name: "loads flat JavaScript file",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				code := "export function handler() {}"
				writeFile(t, jobsDir, "my-job.js", code)
				return jobsDir, "my-job.js"
			},
			wantCode:  "export function handler() {}",
			wantFiles: map[string]string{},
		},
		{
			name: "loads directory-based job with index.ts",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				code := "export function handler() {}"
				writeFile(t, jobDir, "index.ts", code)
				writeFile(t, jobDir, "utils.ts", "export function util() {}")
				writeFile(t, jobDir, "deno.json", `{"compilerOptions": {}}`)
				return jobsDir, "my-job"
			},
			wantCode: "export function handler() {}",
			wantFiles: map[string]string{
				"utils.ts":  "export function util() {}",
				"deno.json": `{"compilerOptions": {}}`,
			},
		},
		{
			name: "loads directory-based job with supporting files",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				code := "export function handler() {}"
				writeFile(t, jobDir, "index.ts", code)
				writeFile(t, jobDir, "types.ts", "export type MyType = string;")
				writeFile(t, jobDir, "deno.json", `{"compilerOptions": {}}`)
				return jobsDir, "my-job"
			},
			wantCode: "export function handler() {}",
			wantFiles: map[string]string{
				"types.ts":  "export type MyType = string;",
				"deno.json": `{"compilerOptions": {}}`,
			},
		},
		{
			name: "loads nested supporting files",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				nestedDir := createDir(t, jobDir, "lib")
				code := "export function handler() {}"
				writeFile(t, jobDir, "index.ts", code)
				writeFile(t, nestedDir, "helper.ts", "export function help() {}")
				return jobsDir, "my-job"
			},
			wantCode: "export function handler() {}",
			wantFiles: map[string]string{
				"lib/helper.ts": "export function help() {}",
			},
		},
		{
			name: "returns error when directory job missing index.ts",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				writeFile(t, jobDir, "handler.ts", "export function handler() {}")
				return jobsDir, "my-job"
			},
			wantErr: true,
		},
		{
			name: "returns error for non-existent file",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				return jobsDir, "does-not-exist.ts"
			},
			wantErr: true,
		},
		{
			name: "excludes index.ts from supporting files",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				code := "export function handler() {}"
				writeFile(t, jobDir, "index.ts", code)
				return jobsDir, "my-job"
			},
			wantCode:  "export function handler() {}",
			wantFiles: map[string]string{},
		},
		{
			name: "handles directory with only index.ts",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				code := "export function handler() {}"
				writeFile(t, jobDir, "index.ts", code)
				return jobsDir, "my-job"
			},
			wantCode:  "export function handler() {}",
			wantFiles: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobsDir, entryName := tt.setup(t)

			loader := &Loader{}
			gotCode, gotFiles, err := loader.loadJobCode(jobsDir, entryName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadJobCode() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadJobCode() unexpected error = %v", err)
			}

			if gotCode != tt.wantCode {
				t.Errorf("loadJobCode() code mismatch\n got: %s\n want: %s", gotCode, tt.wantCode)
			}

			if len(gotFiles) != len(tt.wantFiles) {
				t.Errorf("loadJobCode() files count = %d, want %d", len(gotFiles), len(tt.wantFiles))
			}

			for path, content := range tt.wantFiles {
				if gotContent, exists := gotFiles[path]; !exists {
					t.Errorf("loadJobCode() missing file %q", path)
				} else if gotContent != content {
					t.Errorf("loadJobCode() file %q content mismatch", path)
				}
			}
		})
	}
}

// TestLoader_ParseAnnotations tests the public ParseAnnotations wrapper
func TestLoader_ParseAnnotations(t *testing.T) {
	tests := []struct {
		name string
		code string
		want JobAnnotations
	}{
		{
			name: "parses all annotations",
			code: `// @fluxbase:timeout 600
// @fluxbase:memory 512
// @fluxbase:max-retries 3
// @fluxbase:enabled false
// @fluxbase:allow-read true
// @fluxbase:require-role admin
export function handler() {}`,
			want: JobAnnotations{
				TimeoutSeconds: 600,
				MemoryLimitMB:  512,
				MaxRetries:     3,
				Enabled:        false,
				AllowRead:      true,
				AllowNet:       true,  // default
				AllowEnv:       true,  // default
				AllowWrite:     false, // default
				RequireRoles:   []string{"admin"},
				ScheduleParams: make(map[string]interface{}),
			},
		},
		{
			name: "uses defaults for missing annotations",
			code: "export function handler() {}",
			want: JobAnnotations{
				TimeoutSeconds:         300,
				MemoryLimitMB:          256,
				MaxRetries:             0,
				ProgressTimeoutSeconds: 300,
				Enabled:                true,
				AllowNet:               true,
				AllowEnv:               true,
				AllowRead:              false,
				AllowWrite:             false,
				ScheduleParams:         make(map[string]interface{}),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &Loader{}
			got := loader.ParseAnnotations(tt.code)

			if got.TimeoutSeconds != tt.want.TimeoutSeconds {
				t.Errorf("ParseAnnotations() TimeoutSeconds = %v, want %v", got.TimeoutSeconds, tt.want.TimeoutSeconds)
			}
			if got.MemoryLimitMB != tt.want.MemoryLimitMB {
				t.Errorf("ParseAnnotations() MemoryLimitMB = %v, want %v", got.MemoryLimitMB, tt.want.MemoryLimitMB)
			}
			if got.MaxRetries != tt.want.MaxRetries {
				t.Errorf("ParseAnnotations() MaxRetries = %v, want %v", got.MaxRetries, tt.want.MaxRetries)
			}
			if got.Enabled != tt.want.Enabled {
				t.Errorf("ParseAnnotations() Enabled = %v, want %v", got.Enabled, tt.want.Enabled)
			}
			if got.AllowRead != tt.want.AllowRead {
				t.Errorf("ParseAnnotations() AllowRead = %v, want %v", got.AllowRead, tt.want.AllowRead)
			}
			if got.AllowNet != tt.want.AllowNet {
				t.Errorf("ParseAnnotations() AllowNet = %v, want %v", got.AllowNet, tt.want.AllowNet)
			}
		})
	}
}

// TestLoadJobCode_EdgeCases tests edge cases for loadJobCode
func TestLoadJobCode_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (jobsDir string, entryName string)
		wantErr  bool
		validate func(t *testing.T, code string, files map[string]string)
	}{
		{
			name: "includes mjs and mts files",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				writeFile(t, jobDir, "index.ts", "export function handler() {}")
				writeFile(t, jobDir, "module.mjs", "export {}")
				writeFile(t, jobDir, "types.mts", "export type T = string")
				return jobsDir, "my-job"
			},
			wantErr: false,
			validate: func(t *testing.T, code string, files map[string]string) {
				if _, exists := files["module.mjs"]; !exists {
					t.Error("should include .mjs files")
				}
				if _, exists := files["types.mts"]; !exists {
					t.Error("should include .mts files")
				}
			},
		},
		{
			name: "includes both deno.json and deno.jsonc",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				writeFile(t, jobDir, "index.ts", "export function handler() {}")
				writeFile(t, jobDir, "deno.json", `{"config": "json"}`)
				writeFile(t, jobDir, "deno.jsonc", `{"config": "jsonc"}`)
				return jobsDir, "my-job"
			},
			wantErr: false,
			validate: func(t *testing.T, code string, files map[string]string) {
				if _, exists := files["deno.json"]; !exists {
					t.Error("should include deno.json")
				}
				if _, exists := files["deno.jsonc"]; !exists {
					t.Error("should include deno.jsonc")
				}
			},
		},
		{
			name: "handles directory that is actually a file",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				// Create a file with same name as would-be directory
				writeFile(t, jobsDir, "not-a-dir.ts", "export function handler() {}")
				return jobsDir, "not-a-dir.ts"
			},
			wantErr: false, // Should load as flat file
		},
		{
			name: "handles deeply nested supporting files",
			setup: func(t *testing.T) (string, string) {
				jobsDir := t.TempDir()
				jobDir := createDir(t, jobsDir, "my-job")
				level1 := createDir(t, jobDir, "lib")
				level2 := createDir(t, level1, "utils")
				writeFile(t, jobDir, "index.ts", "export function handler() {}")
				writeFile(t, level2, "deep.ts", "export function deep() {}")
				return jobsDir, "my-job"
			},
			wantErr: false,
			validate: func(t *testing.T, code string, files map[string]string) {
				if _, exists := files["lib/utils/deep.ts"]; !exists {
					t.Error("should include deeply nested files")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobsDir, entryName := tt.setup(t)

			loader := &Loader{}
			gotCode, gotFiles, err := loader.loadJobCode(jobsDir, entryName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadJobCode() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadJobCode() unexpected error = %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, gotCode, gotFiles)
			}
		})
	}
}

// TestLoadNestedFiles_SymlinkAndSpecialCases tests special file system cases
func TestLoadNestedFiles_SpecialCases(t *testing.T) {
	tests := []struct {
		name         string
		setup        func(t *testing.T) (rootPath string, cleanup func())
		relativePath string
		want         int // number of files loaded
		wantErr      bool
	}{
		{
			name: "handles directory with only excluded files",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "README.md", "readme")
				writeFile(t, dir, "LICENSE.txt", "license")
				writeFile(t, dir, "data.bin", "binary")
				return dir, nil
			},
			relativePath: ".",
			want:         0,
		},
		{
			name: "handles all supported file types",
			setup: func(t *testing.T) (string, func()) {
				dir := t.TempDir()
				writeFile(t, dir, "code.ts", "ts code")
				writeFile(t, dir, "script.js", "js code")
				writeFile(t, dir, "data.json", `{"key":"value"}`)
				writeFile(t, dir, "map.geojson", `{"type":"Feature"}`)
				return dir, nil
			},
			relativePath: ".",
			want:         4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootPath, cleanup := tt.setup(t)
			if cleanup != nil {
				defer cleanup()
			}

			loader := &Loader{}
			got, err := loader.loadNestedFiles(rootPath, tt.relativePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadNestedFiles() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadNestedFiles() unexpected error = %v", err)
			}

			if len(got) != tt.want {
				t.Errorf("loadNestedFiles() returned %d files, want %d", len(got), tt.want)
			}
		})
	}
}

// TestLoadSharedModules_SpecialCases tests shared modules edge cases
func TestLoadSharedModules_SpecialCases(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (jobsDir string)
		want    int
		wantErr bool
	}{
		{
			name: "handles empty _shared directory",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				_ = createDir(t, jobsDir, "_shared")
				return jobsDir
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "handles _shared with only excluded files",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				sharedDir := createDir(t, jobsDir, "_shared")
				writeFile(t, sharedDir, "README.md", "readme")
				writeFile(t, sharedDir, "data.txt", "text")
				return jobsDir
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "handles _shared with nested empty directories",
			setup: func(t *testing.T) string {
				jobsDir := t.TempDir()
				sharedDir := createDir(t, jobsDir, "_shared")
				_ = createDir(t, sharedDir, "empty")
				return jobsDir
			},
			want:    0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobsDir := tt.setup(t)

			loader := &Loader{}
			got, err := loader.loadSharedModules(jobsDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("loadSharedModules() should return error")
				}
				return
			}

			if err != nil {
				t.Fatalf("loadSharedModules() unexpected error = %v", err)
			}

			if len(got) != tt.want {
				t.Errorf("loadSharedModules() returned %d modules, want %d", len(got), tt.want)
			}
		})
	}
}

// Helper types and functions for testing

type mockJobsConfig struct {
	jobsDir string
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := dir + "/" + name
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func createDir(t *testing.T, parent, name string) string {
	t.Helper()
	path := parent + "/" + name
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", path, err)
	}
	return path
}
