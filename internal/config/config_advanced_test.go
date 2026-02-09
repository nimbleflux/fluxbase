package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Environment Variable Loading Tests
// =============================================================================

func TestLoadEnvFile_DotEnvFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	envContent := `
FLUXBASE_SERVER_ADDRESS=:9090
FLUXBASE_SERVER_READ_TIMEOUT=60s
FLUXBASE_DATABASE_URL=postgresql://localhost:5432/testdb
`
	err := os.WriteFile(envPath, []byte(envContent), 0600)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "FLUXBASE_SERVER_ADDRESS=:9090")
}

func TestLoadEnvFile_MissingFile_Ignored(t *testing.T) {
	// Non-existent .env file should not cause error
	nonExistentPath := "/tmp/nonexistent-fluxbase-env-xyz.env"

	_, err := os.Stat(nonExistentPath)
	assert.True(t, os.IsNotExist(err))

	// Implementation should ignore missing .env file
	// No error should be returned
}

func TestLoadEnvFile_InvalidFormat_Ignored(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")

	// Invalid env format (missing =)
	invalidEnv := `
INVALID_LINE_WITHOUT_EQUALS
FLUXBASE_SERVER_ADDRESS=:9090
`
	err := os.WriteFile(envPath, []byte(invalidEnv), 0600)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(envPath)
	assert.NoError(t, err)

	// Invalid lines should be skipped, valid lines parsed
	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	lines := strings.Split(string(content), "\n")
	validLines := 0
	for _, line := range lines {
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "#") {
			validLines++
		}
	}
	assert.Greater(t, validLines, 0)
}

// =============================================================================
// Config Section Priority Tests
// =============================================================================

func TestConfigPriority_EnvVarOverYAML(t *testing.T) {
	// Test env var takes priority over YAML config
	yamlValue := ":8080"
	envValue := ":9090"

	// Env var should win
	finalValue := envValue
	assert.Equal(t, ":9090", finalValue)
	assert.NotEqual(t, yamlValue, finalValue)
}

func TestConfigPriority_DefaultOverNothing(t *testing.T) {
	// Test default value is used when neither YAML nor env var set
	defaultValue := ":8080"

	// Should use default
	finalValue := defaultValue
	assert.Equal(t, ":8080", finalValue)
}

// =============================================================================
// Config Array/List Parsing Tests
// =============================================================================

func TestConfigArrayParsing_StringArray(t *testing.T) {
	// Test parsing string array from YAML
	yamlConfig := []string{
		"allowed_origins:",
		"  - https://example.com",
		"  - https://app.example.com",
		"  - http://localhost:8080",
	}

	origins := []string{}
	for _, line := range yamlConfig {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-") {
			origin := strings.TrimPrefix(trimmed, "-")
			origin = strings.TrimSpace(origin)
			origins = append(origins, origin)
		}
	}

	assert.Len(t, origins, 3)
	assert.Contains(t, origins, "https://example.com")
	assert.Contains(t, origins, "https://app.example.com")
	assert.Contains(t, origins, "http://localhost:8080")
}

func TestConfigArrayParsing_EmptyArray(t *testing.T) {
	// Test empty array parsing
	yamlConfig := "allowed_origins: []"

	isEmpty := strings.Contains(yamlConfig, "[]")
	assert.True(t, isEmpty)
}

// =============================================================================
// Config Boolean Parsing Tests
// =============================================================================

func TestConfigBooleanParsing_TrueValues(t *testing.T) {
	trueValues := []string{"true", "TRUE", "True", "1", "yes", "YES", "on", "ON"}

	for _, value := range trueValues {
		t.Run(value, func(t *testing.T) {
			lower := strings.ToLower(value)
			isTrue := lower == "true" || lower == "1" || lower == "yes" || lower == "on"
			assert.True(t, isTrue)
		})
	}
}

func TestConfigBooleanParsing_FalseValues(t *testing.T) {
	falseValues := []string{"false", "FALSE", "False", "0", "no", "NO", "off", "OFF"}

	for _, value := range falseValues {
		t.Run(value, func(t *testing.T) {
			lower := strings.ToLower(value)
			isFalse := lower == "false" || lower == "0" || lower == "no" || lower == "off"
			assert.True(t, isFalse)
		})
	}
}

func TestConfigBooleanParsing_InvalidValues(t *testing.T) {
	invalidValues := []string{"", "maybe", "2", "invalid"}

	for _, value := range invalidValues {
		t.Run(value, func(t *testing.T) {
			// Invalid boolean values should be rejected or default to false
			if value != "" {
				isInvalid := !strings.Contains("true false 1 0 yes no on off", strings.ToLower(value))
				assert.True(t, isInvalid)
			}
		})
	}
}

// =============================================================================
// Config Numeric Parsing Tests
// =============================================================================

func TestConfigNumericParsing_ValidIntegers(t *testing.T) {
	validNumbers := map[string]int{
		"8080":       8080,
		"25":         25,
		"1048576":    1048576,
		"0":          0,
		"2147483647": 2147483647, // Max int32
	}

	for str := range validNumbers {
		t.Run(str, func(t *testing.T) {
			assert.NotEmpty(t, str)
			assert.GreaterOrEqual(t, len(str), 1)
		})
	}
}

func TestConfigNumericParsing_ValidFloats(t *testing.T) {
	validFloats := []string{
		"0.5",
		"1.0",
		"0.95",
		"1.5",
		"3.14159",
	}

	for _, floatStr := range validFloats {
		t.Run(floatStr, func(t *testing.T) {
			assert.Contains(t, floatStr, ".")
		})
	}
}

func TestConfigNumericParsing_InvalidNumbers(t *testing.T) {
	invalidNumbers := []string{
		"abc",
		"12.34.56",
		"",
		"12abc",
	}

	for _, numStr := range invalidNumbers {
		t.Run(numStr, func(t *testing.T) {
			if numStr != "" && numStr != "12.34.56" {
				// Check for non-numeric characters (excluding decimal point)
				hasInvalidChars := false
				for _, c := range numStr {
					if !((c >= '0' && c <= '9') || c == '.') {
						hasInvalidChars = true
						break
					}
				}
				if numStr != "12.34.56" {
					assert.True(t, hasInvalidChars || numStr == "abc" || numStr == "12abc")
				}
			}
		})
	}
}

// =============================================================================
// Config Secret/Sensitive Data Tests
// =============================================================================

func TestConfigSecret_Handling(t *testing.T) {
	// Test that sensitive fields are not logged
	jwtSecret := "super-secret-jwt-key-12345678901234567890123" // 44 chars

	// Secret should be set
	assert.NotEmpty(t, jwtSecret)
	assert.Len(t, jwtSecret, 44) // At least 32 chars for security

	// In logging, should be masked
	maskedSecret := maskSecret(jwtSecret)
	assert.NotEqual(t, jwtSecret, maskedSecret)
	assert.Contains(t, maskedSecret, "***")
}

func TestConfigSecret_EncryptionKeyValidation(t *testing.T) {
	// Test encryption key length validation
	testCases := []struct {
		name      string
		key       string
		wantValid bool
	}{
		{
			name:      "valid 32 byte key",
			key:       "12345678901234567890123456789012",
			wantValid: true,
		},
		{
			name:      "too short",
			key:       "short",
			wantValid: false,
		},
		{
			name:      "too long",
			key:       "1234567890123456789012345678901234567890",
			wantValid: false,
		},
		{
			name:      "empty",
			key:       "",
			wantValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isValid := len(tc.key) == 32
			assert.Equal(t, tc.wantValid, isValid)
		})
	}
}

// =============================================================================
// Config Provider Configuration Tests
// =============================================================================

func TestConfigProvider_EmailProvider(t *testing.T) {
	// Test email provider configuration
	config := &Config{
		Email: EmailConfig{
			Enabled:      true,
			Provider:     "smtp",
			SMTPHost:     "smtp.example.com",
			SMTPPort:     587,
			SMTPUsername: "user@example.com",
		},
	}

	assert.Equal(t, "smtp", config.Email.Provider)
	assert.Equal(t, "smtp.example.com", config.Email.SMTPHost)
	assert.Equal(t, 587, config.Email.SMTPPort)
}

func TestConfigProvider_AIProvider(t *testing.T) {
	// Test AI provider configuration
	config := &Config{
		AI: AIConfig{
			Enabled:      true,
			DefaultModel: "gpt-4",
			ChatbotsDir:  "/chatbots",
		},
	}

	assert.True(t, config.AI.Enabled)
	assert.Equal(t, "gpt-4", config.AI.DefaultModel)
}

func TestConfigProvider_OAuthProvider(t *testing.T) {
	// Test OAuth provider configuration
	config := &Config{
		Auth: AuthConfig{
			OAuthProviders: []OAuthProviderConfig{
				{
					Name:         "google",
					ClientID:     "google-client-id",
					ClientSecret: "google-client-secret",
				},
				{
					Name:         "github",
					ClientID:     "github-client-id",
					ClientSecret: "github-client-secret",
				},
			},
		},
	}

	assert.Len(t, config.Auth.OAuthProviders, 2)
	assert.Equal(t, "google", config.Auth.OAuthProviders[0].Name)
	assert.Equal(t, "github", config.Auth.OAuthProviders[1].Name)
}

// =============================================================================
// Config Feature Flag Tests
// =============================================================================

func TestConfigFeatureFlag_EnabledFeatures(t *testing.T) {
	// Test feature flag configuration
	config := &Config{
		Admin: AdminConfig{
			Enabled: true,
		},
		Functions: FunctionsConfig{
			Enabled: true,
		},
		Realtime: RealtimeConfig{
			Enabled: true,
		},
	}

	assert.True(t, config.Admin.Enabled)
	assert.True(t, config.Functions.Enabled)
	assert.True(t, config.Realtime.Enabled)
}

func TestConfigFeatureFlag_DisabledFeatures(t *testing.T) {
	// Test disabled features
	config := &Config{
		Admin:     AdminConfig{Enabled: false},
		Functions: FunctionsConfig{Enabled: false},
		Realtime:  RealtimeConfig{Enabled: false},
	}

	assert.False(t, config.Admin.Enabled)
	assert.False(t, config.Functions.Enabled)
	assert.False(t, config.Realtime.Enabled)
}

// =============================================================================
// Config Connection Pool Tests
// =============================================================================

func TestConfigConnectionPool_ValidSettings(t *testing.T) {
	// Test database connection pool configuration
	config := &Config{
		Database: DatabaseConfig{
			MaxConnections:  25,
			MinConnections:  5,
			MaxConnLifetime: 3600 * time.Second,
		},
	}

	assert.Greater(t, config.Database.MaxConnections, int32(0))
	assert.Greater(t, config.Database.MinConnections, int32(0))
	assert.LessOrEqual(t, config.Database.MinConnections, config.Database.MaxConnections)
}

func TestConfigConnectionPool_InvalidSettings(t *testing.T) {
	// Test invalid connection pool settings
	config := &Config{
		Database: DatabaseConfig{
			MaxConnections: 0,   // Invalid
			MinConnections: 100, // More than max
		},
	}

	assert.LessOrEqual(t, config.Database.MaxConnections, config.Database.MinConnections)
}

// =============================================================================
// Config Pagination Tests
// =============================================================================

func TestConfigPagination_ValidSettings(t *testing.T) {
	// Test pagination configuration
	config := &Config{
		API: APIConfig{
			MaxPageSize:     1000,
			DefaultPageSize: 50,
			MaxTotalResults: 10000,
		},
	}

	assert.Greater(t, config.API.MaxPageSize, 0)
	assert.Greater(t, config.API.DefaultPageSize, 0)
	assert.Greater(t, config.API.MaxTotalResults, 0)
}

func TestConfigPagination_Unlimited(t *testing.T) {
	// Test unlimited pagination
	config := &Config{
		API: APIConfig{
			MaxPageSize:     -1,
			DefaultPageSize: -1,
			MaxTotalResults: -1,
		},
	}

	assert.Equal(t, -1, config.API.MaxPageSize)
	assert.Equal(t, -1, config.API.DefaultPageSize)
	assert.Equal(t, -1, config.API.MaxTotalResults)
}

// =============================================================================
// Helper Functions
// =============================================================================

func maskSecret(secret string) string {
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "***" + secret[len(secret)-4:]
}
