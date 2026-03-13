package api

import (
	"testing"
	"time"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestMCPOAuthHandler_matchRedirectURI(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		uri      string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "https://claude.ai/api/mcp/auth_callback",
			uri:      "https://claude.ai/api/mcp/auth_callback",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "https://claude.ai/api/mcp/auth_callback",
			uri:      "https://evil.com/callback",
			expected: false,
		},
		{
			name:     "localhost port wildcard - match",
			pattern:  "http://localhost:*",
			uri:      "http://localhost:3000",
			expected: true,
		},
		{
			name:     "localhost port wildcard - match different port",
			pattern:  "http://localhost:*",
			uri:      "http://localhost:8080",
			expected: true,
		},
		{
			name:     "127.0.0.1 port wildcard - match",
			pattern:  "http://127.0.0.1:*",
			uri:      "http://127.0.0.1:33418",
			expected: true,
		},
		{
			name:     "cursor scheme wildcard - match",
			pattern:  "cursor://anysphere.cursor-mcp/oauth/*/callback",
			uri:      "cursor://anysphere.cursor-mcp/oauth/user-server/callback",
			expected: true,
		},
		{
			name:     "cursor scheme wildcard - different server name",
			pattern:  "cursor://anysphere.cursor-mcp/oauth/*/callback",
			uri:      "cursor://anysphere.cursor-mcp/oauth/my-fluxbase/callback",
			expected: true,
		},
		{
			name:     "cursor scheme wildcard - no match wrong ending",
			pattern:  "cursor://anysphere.cursor-mcp/oauth/*/callback",
			uri:      "cursor://anysphere.cursor-mcp/oauth/server/wrong",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchRedirectURI(tt.pattern, tt.uri)
			if result != tt.expected {
				t.Errorf("matchRedirectURI(%q, %q) = %v, want %v", tt.pattern, tt.uri, result, tt.expected)
			}
		})
	}
}

func TestMCPOAuthHandler_verifyPKCE(t *testing.T) {
	tests := []struct {
		name          string
		codeVerifier  string
		codeChallenge string
		method        string
		expected      bool
	}{
		{
			name:         "valid PKCE S256",
			codeVerifier: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			// SHA256("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk") base64url encoded
			codeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			method:        "S256",
			expected:      true,
		},
		{
			name:          "invalid code verifier",
			codeVerifier:  "wrong-verifier",
			codeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM",
			method:        "S256",
			expected:      false,
		},
		{
			name:          "unsupported method",
			codeVerifier:  "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			codeChallenge: "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk",
			method:        "plain",
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := verifyPKCE(tt.codeVerifier, tt.codeChallenge, tt.method)
			if result != tt.expected {
				t.Errorf("verifyPKCE() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMCPOAuthHandler_hashToken(t *testing.T) {
	token1 := "test-token-123"
	token2 := "test-token-456"

	hash1 := hashToken(token1)
	hash2 := hashToken(token2)
	hash1Again := hashToken(token1)

	// Same token should produce same hash
	if hash1 != hash1Again {
		t.Errorf("hashToken() produced different hashes for same input")
	}

	// Different tokens should produce different hashes
	if hash1 == hash2 {
		t.Errorf("hashToken() produced same hash for different inputs")
	}

	// Hash should be hex-encoded SHA256 (64 characters)
	if len(hash1) != 64 {
		t.Errorf("hashToken() hash length = %d, want 64", len(hash1))
	}
}

func TestMCPOAuthHandler_generateRandomString(t *testing.T) {
	lengths := []int{16, 24, 32, 48}

	for _, length := range lengths {
		t.Run("length_"+string(rune('0'+length/10))+string(rune('0'+length%10)), func(t *testing.T) {
			str1 := generateRandomString(length)
			str2 := generateRandomString(length)

			// Check length
			if len(str1) != length {
				t.Errorf("generateRandomString(%d) length = %d, want %d", length, len(str1), length)
			}

			// Should be unique
			if str1 == str2 {
				t.Errorf("generateRandomString() produced duplicate values")
			}
		})
	}
}

func TestMCPOAuthConfig_SetOAuthDefaults(t *testing.T) {
	cfg := &config.MCPConfig{
		Enabled:  true,
		BasePath: "/mcp",
		OAuth: config.MCPOAuthConfig{
			Enabled:    true,
			DCREnabled: true,
		},
	}

	cfg.SetOAuthDefaults()

	// Check defaults were set
	if cfg.OAuth.TokenExpiry != time.Hour {
		t.Errorf("TokenExpiry = %v, want %v", cfg.OAuth.TokenExpiry, time.Hour)
	}

	if cfg.OAuth.RefreshTokenExpiry != 168*time.Hour {
		t.Errorf("RefreshTokenExpiry = %v, want %v", cfg.OAuth.RefreshTokenExpiry, 168*time.Hour)
	}

	if len(cfg.OAuth.AllowedRedirectURIs) == 0 {
		t.Error("AllowedRedirectURIs should have default values")
	}

	// Check some expected default URIs are present
	foundClaude := false
	foundVSCode := false
	for _, uri := range cfg.OAuth.AllowedRedirectURIs {
		if uri == "https://claude.ai/api/mcp/auth_callback" {
			foundClaude = true
		}
		if uri == "http://127.0.0.1:33418" {
			foundVSCode = true
		}
	}

	if !foundClaude {
		t.Error("Default URIs should include Claude callback")
	}
	if !foundVSCode {
		t.Error("Default URIs should include VS Code callback")
	}
}

func TestMCPOAuthConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *config.MCPConfig
		wantErr bool
	}{
		{
			name: "valid config with OAuth",
			config: &config.MCPConfig{
				Enabled:  true,
				BasePath: "/mcp",
				OAuth: config.MCPOAuthConfig{
					Enabled:            true,
					TokenExpiry:        time.Hour,
					RefreshTokenExpiry: 168 * time.Hour,
				},
			},
			wantErr: false,
		},
		{
			name: "disabled MCP - no validation",
			config: &config.MCPConfig{
				Enabled: false,
			},
			wantErr: false,
		},
		{
			name: "negative token expiry",
			config: &config.MCPConfig{
				Enabled:  true,
				BasePath: "/mcp",
				OAuth: config.MCPOAuthConfig{
					Enabled:     true,
					TokenExpiry: -1 * time.Hour,
				},
			},
			wantErr: true,
		},
		{
			name: "negative refresh token expiry",
			config: &config.MCPConfig{
				Enabled:  true,
				BasePath: "/mcp",
				OAuth: config.MCPOAuthConfig{
					Enabled:            true,
					RefreshTokenExpiry: -1 * time.Hour,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultMCPOAuthRedirectURIs(t *testing.T) {
	uris := config.DefaultMCPOAuthRedirectURIs()

	if len(uris) == 0 {
		t.Error("DefaultMCPOAuthRedirectURIs() should return non-empty list")
	}

	// Check essential URIs are present
	essentialURIs := []string{
		"https://claude.ai/api/mcp/auth_callback",
		"https://claude.com/api/mcp/auth_callback",
		"http://127.0.0.1:33418",
		"http://localhost:*",
	}

	for _, essential := range essentialURIs {
		found := false
		for _, uri := range uris {
			if uri == essential {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultMCPOAuthRedirectURIs() missing essential URI: %s", essential)
		}
	}
}

func TestNullIfEmpty(t *testing.T) {
	str := func(s string) *string { return &s }

	tests := []struct {
		input    string
		expected *string
	}{
		{"", nil},
		{"value", str("value")},
		{"  ", str("  ")}, // whitespace is not empty
	}

	for _, tt := range tests {
		result := nullIfEmpty(tt.input)
		if tt.expected == nil {
			if result != nil {
				t.Errorf("nullIfEmpty(%q) = %v, want nil", tt.input, *result)
			}
		} else {
			if result == nil || *result != *tt.expected {
				t.Errorf("nullIfEmpty(%q) = %v, want %v", tt.input, result, *tt.expected)
			}
		}
	}
}
