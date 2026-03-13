package tools

import (
	"net"
	"testing"
	"time"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Constants Tests
// =============================================================================

func TestHttpRequestConstants(t *testing.T) {
	t.Run("httpRequestTimeout is 10 seconds", func(t *testing.T) {
		assert.Equal(t, 10*time.Second, httpRequestTimeout)
	})

	t.Run("maxResponseSize is 1MB", func(t *testing.T) {
		assert.Equal(t, 1024*1024, maxResponseSize)
	})

	t.Run("httpUserAgent is set", func(t *testing.T) {
		assert.Equal(t, "Fluxbase-MCP/1.0", httpUserAgent)
	})
}

// =============================================================================
// HttpRequestTool Metadata Tests
// =============================================================================

func TestHttpRequestTool_Name(t *testing.T) {
	tool := NewHttpRequestTool()
	assert.Equal(t, "http_request", tool.Name())
}

func TestHttpRequestTool_Description(t *testing.T) {
	tool := NewHttpRequestTool()
	desc := tool.Description()
	assert.Contains(t, desc, "HTTP")
	assert.Contains(t, desc, "GET")
}

func TestHttpRequestTool_RequiredScopes(t *testing.T) {
	tool := NewHttpRequestTool()
	scopes := tool.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteHTTP, scopes[0])
}

func TestHttpRequestTool_InputSchema(t *testing.T) {
	tool := NewHttpRequestTool()
	schema := tool.InputSchema()

	// Check type
	assert.Equal(t, "object", schema["type"])

	// Check properties exist
	props, ok := schema["properties"].(map[string]any)
	require.True(t, ok)

	// Check url property
	urlProp, ok := props["url"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", urlProp["type"])

	// Check method property
	methodProp, ok := props["method"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "string", methodProp["type"])
	assert.Equal(t, "GET", methodProp["default"])

	// Check required fields
	required, ok := schema["required"].([]string)
	require.True(t, ok)
	assert.Contains(t, required, "url")
}

func TestNewHttpRequestTool(t *testing.T) {
	tool := NewHttpRequestTool()
	require.NotNil(t, tool)
	assert.NotNil(t, tool.client)
}

// =============================================================================
// isDomainAllowed Tests
// =============================================================================

func TestIsDomainAllowed(t *testing.T) {
	tests := []struct {
		name           string
		hostname       string
		allowedDomains []string
		expected       bool
	}{
		// Exact match cases
		{
			name:           "exact match",
			hostname:       "api.example.com",
			allowedDomains: []string{"api.example.com"},
			expected:       true,
		},
		{
			name:           "exact match case insensitive",
			hostname:       "API.EXAMPLE.COM",
			allowedDomains: []string{"api.example.com"},
			expected:       true,
		},
		{
			name:           "no match",
			hostname:       "other.com",
			allowedDomains: []string{"api.example.com"},
			expected:       false,
		},

		// Wildcard cases
		{
			name:           "wildcard matches subdomain",
			hostname:       "sub.example.com",
			allowedDomains: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard matches base domain",
			hostname:       "example.com",
			allowedDomains: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard matches deep subdomain",
			hostname:       "deep.sub.example.com",
			allowedDomains: []string{"*.example.com"},
			expected:       true,
		},
		{
			name:           "wildcard does not match different domain",
			hostname:       "example.org",
			allowedDomains: []string{"*.example.com"},
			expected:       false,
		},

		// Multiple allowed domains
		{
			name:           "matches one of multiple domains",
			hostname:       "api.github.com",
			allowedDomains: []string{"api.example.com", "*.github.com"},
			expected:       true,
		},

		// Edge cases
		{
			name:           "empty allowed domains",
			hostname:       "api.example.com",
			allowedDomains: []string{},
			expected:       false,
		},
		{
			name:           "empty string in allowed domains ignored",
			hostname:       "api.example.com",
			allowedDomains: []string{"", "api.example.com"},
			expected:       true,
		},
		{
			name:           "whitespace trimmed",
			hostname:       "api.example.com",
			allowedDomains: []string{"  api.example.com  "},
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDomainAllowed(tt.hostname, tt.allowedDomains)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// isPrivateIPAddress Tests
// =============================================================================

func TestIsPrivateIPAddress(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Loopback
		{"IPv4 loopback", "127.0.0.1", true},
		{"IPv4 loopback range", "127.0.0.5", true},
		{"IPv6 loopback", "::1", true},

		// Private ranges (RFC 1918)
		{"10.x.x.x private", "10.0.0.1", true},
		{"10.x.x.x private range", "10.255.255.255", true},
		{"172.16.x.x private", "172.16.0.1", true},
		{"172.31.x.x private", "172.31.255.255", true},
		{"192.168.x.x private", "192.168.1.1", true},

		// Link-local (AWS metadata range)
		{"169.254.x.x link-local", "169.254.169.254", true},
		{"169.254.1.1 link-local", "169.254.1.1", true},

		// Carrier-grade NAT
		{"100.64.x.x CGNAT", "100.64.0.1", true},
		{"100.127.x.x CGNAT", "100.127.255.255", true},

		// Test networks
		{"192.0.2.x TEST-NET-1", "192.0.2.1", true},
		{"198.51.100.x TEST-NET-2", "198.51.100.1", true},
		{"203.0.113.x TEST-NET-3", "203.0.113.1", true},

		// Multicast
		{"224.x.x.x multicast", "224.0.0.1", true},
		{"239.x.x.x multicast", "239.255.255.255", true},

		// Reserved
		{"240.x.x.x reserved", "240.0.0.1", true},

		// Public IPs - should NOT be private
		{"8.8.8.8 Google DNS public", "8.8.8.8", false},
		{"1.1.1.1 Cloudflare public", "1.1.1.1", false},
		{"93.184.216.34 example.com public", "93.184.216.34", false},
		{"104.16.0.1 public", "104.16.0.1", false},

		// IPv6 private ranges
		{"fc00:: unique local", "fc00::1", true},
		{"fd00:: unique local", "fd00::1", true},
		{"fe80:: link local", "fe80::1", true},

		// IPv6 public
		{"2001:4860:4860::8888 Google DNS public", "2001:4860:4860::8888", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			require.NotNil(t, ip, "Failed to parse IP: %s", tt.ip)
			result := isPrivateIPAddress(ip)
			assert.Equal(t, tt.expected, result, "IP: %s", tt.ip)
		})
	}

	t.Run("nil IP returns false", func(t *testing.T) {
		result := isPrivateIPAddress(nil)
		assert.False(t, result)
	})
}

// =============================================================================
// httpRequestResult Tests
// =============================================================================

func TestHttpRequestResult_Struct(t *testing.T) {
	t.Run("success result", func(t *testing.T) {
		result := &httpRequestResult{
			Success: true,
			Data:    map[string]any{"key": "value"},
			Status:  200,
		}
		assert.True(t, result.Success)
		assert.Equal(t, 200, result.Status)
		assert.NotNil(t, result.Data)
	})

	t.Run("error result", func(t *testing.T) {
		result := &httpRequestResult{
			Success:        false,
			Error:          "domain not allowed",
			AllowedDomains: []string{"api.example.com"},
		}
		assert.False(t, result.Success)
		assert.Equal(t, "domain not allowed", result.Error)
		assert.Len(t, result.AllowedDomains, 1)
	})

	t.Run("empty result", func(t *testing.T) {
		result := &httpRequestResult{}
		assert.False(t, result.Success)
		assert.Equal(t, 0, result.Status)
		assert.Nil(t, result.Data)
		assert.Empty(t, result.Error)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkIsDomainAllowed_ExactMatch(b *testing.B) {
	domains := []string{"api.example.com", "api.github.com", "api.google.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isDomainAllowed("api.github.com", domains)
	}
}

func BenchmarkIsDomainAllowed_Wildcard(b *testing.B) {
	domains := []string{"*.example.com", "*.github.com", "*.google.com"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isDomainAllowed("sub.github.com", domains)
	}
}

func BenchmarkIsPrivateIPAddress_Public(b *testing.B) {
	ip := net.ParseIP("8.8.8.8")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isPrivateIPAddress(ip)
	}
}

func BenchmarkIsPrivateIPAddress_Private(b *testing.B) {
	ip := net.ParseIP("192.168.1.1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isPrivateIPAddress(ip)
	}
}
