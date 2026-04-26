package runtime

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/nimbleflux/fluxbase/internal/util"
)

func TestBuildEnvForFunction(t *testing.T) {
	// Set up test environment variables
	testVars := map[string]string{
		// Should be included
		"FLUXBASE_BASE_URL": "http://localhost:8080",
		"FLUXBASE_DEBUG":    "true",
		// Should be blocked (secrets)
		"FLUXBASE_AUTH_JWT_SECRET":             "super-secret",
		"FLUXBASE_DATABASE_PASSWORD":           "db-password",
		"FLUXBASE_STORAGE_S3_SECRET_KEY":       "s3-secret",
		"FLUXBASE_EMAIL_SMTP_PASSWORD":         "smtp-password",
		"FLUXBASE_SECURITY_SETUP_TOKEN":        "setup-token",
		"FLUXBASE_DATABASE_ADMIN_PASSWORD":     "admin-password",
		"FLUXBASE_STORAGE_S3_ACCESS_KEY":       "s3-access-key",
		"FLUXBASE_SERVICE_ROLE_KEY":            "test-service-key",
		"FLUXBASE_ANON_KEY":                    "test-anon-key",
		"FLUXBASE_DATABASE_URL":                "postgres://user:pass@host:5432/db",
		"FLUXBASE_DATABASE_ADMIN_URL":          "postgres://admin:pass@host:5432/db",
		"FLUXBASE_EMAIL_SENDGRID_API_KEY":      "sg-test-key",
		"FLUXBASE_EMAIL_MAILGUN_API_KEY":       "mg-test-key",
		"FLUXBASE_EMAIL_SES_SECRET_ACCESS_KEY": "ses-test-key",
	}

	// Set environment variables
	for key, value := range testVars {
		t.Setenv(key, value)
	}

	// Test with RuntimeTypeFunction
	req := ExecutionRequest{
		ID:        uuid.New(),
		Name:      "test-function",
		Namespace: "default",
	}
	env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, nil)

	// Convert to map for easier testing
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Test that allowed variables are included
	allowedVars := []string{
		"FLUXBASE_BASE_URL",
		"FLUXBASE_DEBUG",
	}

	for _, key := range allowedVars {
		if value, ok := envMap[key]; !ok {
			t.Errorf("Expected environment variable %s to be included, but it was not", key)
		} else if value != testVars[key] {
			t.Errorf("Expected %s=%s, got %s=%s", key, testVars[key], key, value)
		}
	}

	// Test that blocked variables are excluded
	blockedVarsToCheck := []string{
		"FLUXBASE_AUTH_JWT_SECRET",
		"FLUXBASE_DATABASE_PASSWORD",
		"FLUXBASE_STORAGE_S3_SECRET_KEY",
		"FLUXBASE_EMAIL_SMTP_PASSWORD",
		"FLUXBASE_SECURITY_SETUP_TOKEN",
		"FLUXBASE_DATABASE_ADMIN_PASSWORD",
		"FLUXBASE_STORAGE_S3_ACCESS_KEY",
		"FLUXBASE_SERVICE_ROLE_KEY",
		"FLUXBASE_ANON_KEY",
		"FLUXBASE_DATABASE_URL",
		"FLUXBASE_DATABASE_ADMIN_URL",
		"FLUXBASE_EMAIL_SENDGRID_API_KEY",
		"FLUXBASE_EMAIL_MAILGUN_API_KEY",
		"FLUXBASE_EMAIL_SES_SECRET_ACCESS_KEY",
	}

	for _, key := range blockedVarsToCheck {
		if _, ok := envMap[key]; ok {
			t.Errorf("Expected environment variable %s to be blocked, but it was included", key)
		}
	}

	// Test system variables behavior
	t.Setenv("PATH", "/usr/bin")
	t.Setenv("HOME", "/home/user")
	t.Setenv("RANDOM_VAR", "should-be-excluded")

	env = buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, nil)
	envMap = make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// PATH is intentionally included for subprocess operation (finding executables)
	if envMap["PATH"] != "/usr/bin" {
		t.Errorf("Expected PATH=/usr/bin (for subprocess operation), got PATH=%s", envMap["PATH"])
	}
	// HOME is intentionally set to /tmp for Deno runtime requirements (overrides any existing value)
	if envMap["HOME"] != "/tmp" {
		t.Errorf("Expected HOME=/tmp (for Deno), got HOME=%s", envMap["HOME"])
	}
	// Random non-system, non-FLUXBASE variables should be excluded
	if _, ok := envMap["RANDOM_VAR"]; ok {
		t.Error("Expected RANDOM_VAR to be excluded, but it was included")
	}

	// Test that function-specific variables are included
	if _, ok := envMap["FLUXBASE_EXECUTION_ID"]; !ok {
		t.Error("Expected FLUXBASE_EXECUTION_ID to be included")
	}
	if envMap["FLUXBASE_FUNCTION_NAME"] != "test-function" {
		t.Errorf("Expected FLUXBASE_FUNCTION_NAME=test-function, got %s", envMap["FLUXBASE_FUNCTION_NAME"])
	}
	if envMap["FLUXBASE_USER_TOKEN"] != "user-token" {
		t.Errorf("Expected FLUXBASE_USER_TOKEN=user-token, got %s", envMap["FLUXBASE_USER_TOKEN"])
	}
	if envMap["FLUXBASE_SERVICE_TOKEN"] != "service-token" {
		t.Errorf("Expected FLUXBASE_SERVICE_TOKEN=service-token, got %s", envMap["FLUXBASE_SERVICE_TOKEN"])
	}
}

func TestBuildEnvForJob(t *testing.T) {
	// Test with RuntimeTypeJob
	req := ExecutionRequest{
		ID:        uuid.New(),
		Name:      "test-job",
		Namespace: "default",
	}
	env := buildEnv(req, RuntimeTypeJob, "http://localhost:8080", "job-token", "service-token", nil, nil)

	// Convert to map for easier testing
	envMap := make(map[string]string)
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	// Test that job-specific variables are included
	if _, ok := envMap["FLUXBASE_JOB_ID"]; !ok {
		t.Error("Expected FLUXBASE_JOB_ID to be included")
	}
	if envMap["FLUXBASE_JOB_NAME"] != "test-job" {
		t.Errorf("Expected FLUXBASE_JOB_NAME=test-job, got %s", envMap["FLUXBASE_JOB_NAME"])
	}
	if envMap["FLUXBASE_JOB_TOKEN"] != "job-token" {
		t.Errorf("Expected FLUXBASE_JOB_TOKEN=job-token, got %s", envMap["FLUXBASE_JOB_TOKEN"])
	}
	if envMap["FLUXBASE_SERVICE_TOKEN"] != "service-token" {
		t.Errorf("Expected FLUXBASE_SERVICE_TOKEN=service-token, got %s", envMap["FLUXBASE_SERVICE_TOKEN"])
	}
}

func TestRuntimeType(t *testing.T) {
	if RuntimeTypeFunction.String() != "function" {
		t.Errorf("Expected RuntimeTypeFunction.String() = 'function', got '%s'", RuntimeTypeFunction.String())
	}
	if RuntimeTypeJob.String() != "job" {
		t.Errorf("Expected RuntimeTypeJob.String() = 'job', got '%s'", RuntimeTypeJob.String())
	}
}

func TestCancelSignal(t *testing.T) {
	signal := NewCancelSignal()

	if signal.IsCancelled() {
		t.Error("Expected new signal to not be cancelled")
	}

	signal.Cancel()

	if !signal.IsCancelled() {
		t.Error("Expected signal to be cancelled after Cancel()")
	}

	// Verify context is done
	select {
	case <-signal.Context().Done():
		// Good, context was cancelled
	default:
		t.Error("Expected context to be done after Cancel()")
	}
}

func TestBuildEnvWithSecrets(t *testing.T) {
	req := ExecutionRequest{
		ID:        uuid.New(),
		Name:      "test-function",
		Namespace: "default",
	}

	t.Run("secrets are injected as FLUXBASE_SECRET_NAME", func(t *testing.T) {
		secrets := map[string]string{
			"API_KEY":     "sk-1234567890",
			"DB_PASSWORD": "supersecret",
			"OAUTH_TOKEN": "oauth-xyz",
		}

		env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, secrets)

		envMap := make(map[string]string)
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		// Check that secrets are injected with correct prefix
		if envMap["FLUXBASE_SECRET_API_KEY"] != "sk-1234567890" {
			t.Errorf("Expected FLUXBASE_SECRET_API_KEY=sk-1234567890, got %s", envMap["FLUXBASE_SECRET_API_KEY"])
		}
		if envMap["FLUXBASE_SECRET_DB_PASSWORD"] != "supersecret" {
			t.Errorf("Expected FLUXBASE_SECRET_DB_PASSWORD=supersecret, got %s", envMap["FLUXBASE_SECRET_DB_PASSWORD"])
		}
		if envMap["FLUXBASE_SECRET_OAUTH_TOKEN"] != "oauth-xyz" {
			t.Errorf("Expected FLUXBASE_SECRET_OAUTH_TOKEN=oauth-xyz, got %s", envMap["FLUXBASE_SECRET_OAUTH_TOKEN"])
		}
	})

	t.Run("secret names are uppercased", func(t *testing.T) {
		secrets := map[string]string{
			"lowercase_key": "value1",
			"MixedCase_Key": "value2",
		}

		env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, secrets)

		envMap := make(map[string]string)
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		// Names should be uppercased
		if envMap["FLUXBASE_SECRET_LOWERCASE_KEY"] != "value1" {
			t.Errorf("Expected FLUXBASE_SECRET_LOWERCASE_KEY=value1, got %s", envMap["FLUXBASE_SECRET_LOWERCASE_KEY"])
		}
		if envMap["FLUXBASE_SECRET_MIXEDCASE_KEY"] != "value2" {
			t.Errorf("Expected FLUXBASE_SECRET_MIXEDCASE_KEY=value2, got %s", envMap["FLUXBASE_SECRET_MIXEDCASE_KEY"])
		}
	})

	t.Run("empty secrets map does not add any secret vars", func(t *testing.T) {
		secrets := map[string]string{}

		env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, secrets)

		for _, e := range env {
			if strings.HasPrefix(e, "FLUXBASE_SECRET_") {
				t.Errorf("Expected no FLUXBASE_SECRET_ vars with empty secrets map, but found: %s", e)
			}
		}
	})

	t.Run("nil secrets map does not add any secret vars", func(t *testing.T) {
		env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, nil)

		for _, e := range env {
			if strings.HasPrefix(e, "FLUXBASE_SECRET_") {
				t.Errorf("Expected no FLUXBASE_SECRET_ vars with nil secrets, but found: %s", e)
			}
		}
	})

	t.Run("secrets work for job runtime type", func(t *testing.T) {
		jobReq := ExecutionRequest{
			ID:        uuid.New(),
			Name:      "test-job",
			Namespace: "default",
		}
		secrets := map[string]string{
			"JOB_SECRET": "job-secret-value",
		}

		env := buildEnv(jobReq, RuntimeTypeJob, "http://localhost:8080", "job-token", "service-token", nil, secrets)

		envMap := make(map[string]string)
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		if envMap["FLUXBASE_SECRET_JOB_SECRET"] != "job-secret-value" {
			t.Errorf("Expected FLUXBASE_SECRET_JOB_SECRET=job-secret-value, got %s", envMap["FLUXBASE_SECRET_JOB_SECRET"])
		}
	})

	t.Run("secrets with special characters in values", func(t *testing.T) {
		secrets := map[string]string{
			"SPECIAL":     "p@$$w0rd!#$%^&*()",
			"JSON_SECRET": `{"key": "value"}`,
			"MULTILINE":   "line1\nline2",
			"EQUALS_SIGN": "key=value=more",
			"UNICODE":     "日本語🔐",
		}

		env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, secrets)

		envMap := make(map[string]string)
		for _, e := range env {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				envMap[parts[0]] = parts[1]
			}
		}

		if envMap["FLUXBASE_SECRET_SPECIAL"] != "p@$$w0rd!#$%^&*()" {
			t.Errorf("Special characters not preserved: got %s", envMap["FLUXBASE_SECRET_SPECIAL"])
		}
		if envMap["FLUXBASE_SECRET_JSON_SECRET"] != `{"key": "value"}` {
			t.Errorf("JSON not preserved: got %s", envMap["FLUXBASE_SECRET_JSON_SECRET"])
		}
		if envMap["FLUXBASE_SECRET_EQUALS_SIGN"] != "key=value=more" {
			t.Errorf("Equals signs not preserved: got %s", envMap["FLUXBASE_SECRET_EQUALS_SIGN"])
		}
		if envMap["FLUXBASE_SECRET_UNICODE"] != "日本語🔐" {
			t.Errorf("Unicode not preserved: got %s", envMap["FLUXBASE_SECRET_UNICODE"])
		}
	})
}

func TestAllowedEnvVarsWithSecrets(t *testing.T) {
	t.Run("function runtime includes secret names", func(t *testing.T) {
		secretNames := []string{"API_KEY", "DB_PASSWORD"}
		allowed := allowedEnvVars(RuntimeTypeFunction, secretNames)

		if !strings.Contains(allowed, "FLUXBASE_SECRET_API_KEY") {
			t.Error("Expected FLUXBASE_SECRET_API_KEY in allowed vars")
		}
		if !strings.Contains(allowed, "FLUXBASE_SECRET_DB_PASSWORD") {
			t.Error("Expected FLUXBASE_SECRET_DB_PASSWORD in allowed vars")
		}
		// Base function vars should still be present
		if !strings.Contains(allowed, "FLUXBASE_URL") {
			t.Error("Expected FLUXBASE_URL in allowed vars")
		}
		if !strings.Contains(allowed, "FLUXBASE_FUNCTION_NAME") {
			t.Error("Expected FLUXBASE_FUNCTION_NAME in allowed vars")
		}
	})

	t.Run("job runtime includes secret names", func(t *testing.T) {
		secretNames := []string{"JOB_SECRET"}
		allowed := allowedEnvVars(RuntimeTypeJob, secretNames)

		if !strings.Contains(allowed, "FLUXBASE_SECRET_JOB_SECRET") {
			t.Error("Expected FLUXBASE_SECRET_JOB_SECRET in allowed vars")
		}
		// Base job vars should still be present
		if !strings.Contains(allowed, "FLUXBASE_URL") {
			t.Error("Expected FLUXBASE_URL in allowed vars")
		}
		if !strings.Contains(allowed, "FLUXBASE_JOB_NAME") {
			t.Error("Expected FLUXBASE_JOB_NAME in allowed vars")
		}
	})

	t.Run("empty secret names does not add secret vars", func(t *testing.T) {
		allowed := allowedEnvVars(RuntimeTypeFunction, []string{})

		if strings.Contains(allowed, "FLUXBASE_SECRET_") {
			t.Error("Expected no FLUXBASE_SECRET_ in allowed vars with empty slice")
		}
	})

	t.Run("nil secret names does not add secret vars", func(t *testing.T) {
		allowed := allowedEnvVars(RuntimeTypeFunction, nil)

		if strings.Contains(allowed, "FLUXBASE_SECRET_") {
			t.Error("Expected no FLUXBASE_SECRET_ in allowed vars with nil slice")
		}
	})

	t.Run("secret names are uppercased in allowed list", func(t *testing.T) {
		secretNames := []string{"lowercase", "MixedCase"}
		allowed := allowedEnvVars(RuntimeTypeFunction, secretNames)

		if !strings.Contains(allowed, "FLUXBASE_SECRET_LOWERCASE") {
			t.Error("Expected FLUXBASE_SECRET_LOWERCASE (uppercased) in allowed vars")
		}
		if !strings.Contains(allowed, "FLUXBASE_SECRET_MIXEDCASE") {
			t.Error("Expected FLUXBASE_SECRET_MIXEDCASE (uppercased) in allowed vars")
		}
	})
}

func TestEncryptionKeyBlocked(t *testing.T) {
	// Verify that FLUXBASE_ENCRYPTION_KEY is in the blocked list
	if !blockedVars["FLUXBASE_ENCRYPTION_KEY"] {
		t.Error("FLUXBASE_ENCRYPTION_KEY should be in blockedVars")
	}

	// Set the encryption key env var
	t.Setenv("FLUXBASE_ENCRYPTION_KEY", "my-secret-encryption-key")

	req := ExecutionRequest{
		ID:        uuid.New(),
		Name:      "test-function",
		Namespace: "default",
	}

	env := buildEnv(req, RuntimeTypeFunction, "http://localhost:8080", "user-token", "service-token", nil, nil)

	// Check that FLUXBASE_ENCRYPTION_KEY is not in the env
	for _, e := range env {
		if strings.HasPrefix(e, "FLUXBASE_ENCRYPTION_KEY=") {
			t.Error("FLUXBASE_ENCRYPTION_KEY should be blocked from env")
		}
	}
}

// =============================================================================
// Output Size Limit Tests
// =============================================================================

func TestWithMaxOutputSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int
		expected int
	}{
		{"zero (unlimited)", 0, 0},
		{"10MB", 10 * 1024 * 1024, 10 * 1024 * 1024},
		{"1KB", 1024, 1024},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRuntime(RuntimeTypeFunction, "secret", "http://localhost")
			WithMaxOutputSize(tt.bytes)(r)
			if r.maxOutputSize != tt.expected {
				t.Errorf("expected maxOutputSize=%d, got %d", tt.expected, r.maxOutputSize)
			}
		})
	}
}

func TestNewRuntime_DefaultMaxOutputSize(t *testing.T) {
	t.Run("function runtime has 10MB default", func(t *testing.T) {
		r := NewRuntime(RuntimeTypeFunction, "secret", "http://localhost")
		expected := 10 * 1024 * 1024
		if r.maxOutputSize != expected {
			t.Errorf("expected maxOutputSize=%d for functions, got %d", expected, r.maxOutputSize)
		}
	})

	t.Run("job runtime has 50MB default", func(t *testing.T) {
		r := NewRuntime(RuntimeTypeJob, "secret", "http://localhost")
		expected := 50 * 1024 * 1024
		if r.maxOutputSize != expected {
			t.Errorf("expected maxOutputSize=%d for jobs, got %d", expected, r.maxOutputSize)
		}
	})

	t.Run("custom option overrides default", func(t *testing.T) {
		r := NewRuntime(RuntimeTypeFunction, "secret", "http://localhost", WithMaxOutputSize(5*1024*1024))
		expected := 5 * 1024 * 1024
		if r.maxOutputSize != expected {
			t.Errorf("expected maxOutputSize=%d with custom option, got %d", expected, r.maxOutputSize)
		}
	})
}

func TestWithMemoryLimit(t *testing.T) {
	r := NewRuntime(RuntimeTypeFunction, "secret", "http://localhost")
	WithMemoryLimit(256)(r)
	if r.memoryLimitMB != 256 {
		t.Errorf("expected memoryLimitMB=256, got %d", r.memoryLimitMB)
	}
}

func TestWithTimeout(t *testing.T) {
	r := NewRuntime(RuntimeTypeFunction, "secret", "http://localhost")
	timeout := 60 * 1000 * 1000 * 1000 // 60 seconds in nanoseconds
	WithTimeout(60 * 1000000000)(r)
	if r.defaultTimeout != 60*1000000000 {
		t.Errorf("expected defaultTimeout=%d, got %d", timeout, r.defaultTimeout)
	}
}

// =============================================================================
// truncateString Tests
// =============================================================================

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than maxLen",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string equal to maxLen",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than maxLen gets truncated",
			input:    "hello world",
			maxLen:   8,
			expected: "hello wo...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   10,
			expected: "",
		},
		{
			name:     "maxLen of 3",
			input:    "hello",
			maxLen:   3,
			expected: "hel...",
		},
		{
			name:     "unicode string truncation",
			input:    "hello 世界 world",
			maxLen:   12,
			expected: "hello 世界...",
		},
		{
			name:     "single character with small maxLen",
			input:    "abcdef",
			maxLen:   4,
			expected: "abcd...",
		},
		{
			name:     "long string truncation",
			input:    "This is a very long string that should be truncated",
			maxLen:   20,
			expected: "This is a very long ...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := util.TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("util.TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// classifyStderrLine Tests
// =============================================================================

func TestClassifyStderrLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		// Info patterns
		{
			name:     "Download message",
			line:     "Download https://deno.land/std/testing/asserts.ts",
			expected: "info",
		},
		{
			name:     "Downloading message",
			line:     "Downloading https://example.com/module.ts",
			expected: "info",
		},
		{
			name:     "Check message",
			line:     "Check file:///tmp/function.ts",
			expected: "info",
		},
		{
			name:     "Checking message",
			line:     "Checking file:///tmp/script.ts",
			expected: "info",
		},
		{
			name:     "Compile message",
			line:     "Compile file:///tmp/main.ts",
			expected: "info",
		},
		{
			name:     "Compiling message",
			line:     "Compiling file:///tmp/app.ts",
			expected: "info",
		},
		// Warning patterns
		{
			name:     "Warning keyword",
			line:     "Warning: this feature is deprecated",
			expected: "warn",
		},
		{
			name:     "warning colon lowercase",
			line:     "warning: unused variable 'x'",
			expected: "warn",
		},
		// Error patterns (default)
		{
			name:     "error message",
			line:     "Error: Cannot find module",
			expected: "error",
		},
		{
			name:     "syntax error",
			line:     "SyntaxError: Unexpected token",
			expected: "error",
		},
		{
			name:     "generic stderr",
			line:     "Something went wrong",
			expected: "error",
		},
		{
			name:     "empty line",
			line:     "",
			expected: "error",
		},
		// ANSI codes should be stripped
		{
			name:     "Download with ANSI colors",
			line:     "\x1b[32mDownload\x1b[0m https://example.com/mod.ts",
			expected: "info",
		},
		{
			name:     "Warning with ANSI colors",
			line:     "\x1b[33mWarning:\x1b[0m deprecated API",
			expected: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyStderrLine(tt.line)
			if result != tt.expected {
				t.Errorf("classifyStderrLine(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// stripAnsiCodes Tests
// =============================================================================

func TestStripAnsiCodes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no ANSI codes",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single color code",
			input:    "\x1b[32mgreen text\x1b[0m",
			expected: "green text",
		},
		{
			name:     "multiple color codes",
			input:    "\x1b[31mred\x1b[0m and \x1b[34mblue\x1b[0m",
			expected: "red and blue",
		},
		{
			name:     "bold and colors",
			input:    "\x1b[1m\x1b[33mbold yellow\x1b[0m",
			expected: "bold yellow",
		},
		{
			name:     "cursor movement codes",
			input:    "\x1b[2Jcleared screen",
			expected: "cleared screen",
		},
		{
			name:     "256 color code",
			input:    "\x1b[38;5;196mdeep red\x1b[0m",
			expected: "deep red",
		},
		{
			name:     "RGB color code",
			input:    "\x1b[38;2;255;100;50mRGB color\x1b[0m",
			expected: "RGB color",
		},
		{
			name:     "mixed content",
			input:    "before \x1b[31mcolored\x1b[0m after",
			expected: "before colored after",
		},
		{
			name:     "consecutive codes",
			input:    "\x1b[1m\x1b[4m\x1b[32mstacked\x1b[0m",
			expected: "stacked",
		},
		{
			name:     "underline code",
			input:    "\x1b[4munderlined\x1b[24m",
			expected: "underlined",
		},
		{
			name:     "background color",
			input:    "\x1b[44mblue background\x1b[49m",
			expected: "blue background",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripAnsiCodes(tt.input)
			if result != tt.expected {
				t.Errorf("stripAnsiCodes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// Utility Function Benchmarks
// =============================================================================

func BenchmarkTruncateString_Short(b *testing.B) {
	input := "short"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.TruncateString(input, 100)
	}
}

func BenchmarkTruncateString_NeedsTruncation(b *testing.B) {
	input := "This is a very long string that definitely needs to be truncated"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		util.TruncateString(input, 20)
	}
}

func BenchmarkClassifyStderrLine_Info(b *testing.B) {
	line := "Download https://deno.land/std/module.ts"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifyStderrLine(line)
	}
}

func BenchmarkClassifyStderrLine_Error(b *testing.B) {
	line := "Error: Something went wrong"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		classifyStderrLine(line)
	}
}

func BenchmarkStripAnsiCodes_Plain(b *testing.B) {
	input := "plain text without any codes"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stripAnsiCodes(input)
	}
}

func BenchmarkStripAnsiCodes_WithCodes(b *testing.B) {
	input := "\x1b[32mgreen\x1b[0m \x1b[31mred\x1b[0m \x1b[34mblue\x1b[0m"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stripAnsiCodes(input)
	}
}
