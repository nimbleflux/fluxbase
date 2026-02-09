package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// ParseFunctionConfig Tests
// =============================================================================

func TestParseFunctionConfig_Defaults(t *testing.T) {
	code := `
		// No special directives
		export default async function(req) {
			return new Response("Hello");
		}
	`

	config := ParseFunctionConfig(code)

	assert.False(t, config.AllowUnauthenticated, "Should require auth by default")
	assert.True(t, config.IsPublic, "Should be public by default")
	assert.False(t, config.DisableExecutionLogs, "Should have logs enabled by default")
	assert.Nil(t, config.CorsOrigins, "CORS origins should be nil (use global defaults)")
	assert.Nil(t, config.CorsMethods, "CORS methods should be nil")
	assert.Nil(t, config.CorsHeaders, "CORS headers should be nil")
	assert.Nil(t, config.CorsCredentials, "CORS credentials should be nil")
	assert.Nil(t, config.CorsMaxAge, "CORS max-age should be nil")
	assert.Nil(t, config.RateLimitPerMinute, "Rate limit should be nil (unlimited)")
	assert.Nil(t, config.RateLimitPerHour, "Rate limit should be nil")
	assert.Nil(t, config.RateLimitPerDay, "Rate limit should be nil")
}

func TestParseFunctionConfig_AllowUnauthenticated(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "single line comment",
			code: `
				// @fluxbase:allow-unauthenticated
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "multi-line comment",
			code: `
				/*
				 * @fluxbase:allow-unauthenticated
				 */
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "star comment",
			code: `
				* @fluxbase:allow-unauthenticated
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "with whitespace",
			code: `
			    	// @fluxbase:allow-unauthenticated
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			assert.Equal(t, tt.expected, config.AllowUnauthenticated)
		})
	}
}

func TestParseFunctionConfig_Public(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "explicitly public",
			code: `
				// @fluxbase:public true
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "explicitly private",
			code: `
				// @fluxbase:public false
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: false,
		},
		{
			name: "implicit public (directive only)",
			code: `
				// @fluxbase:public
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "no directive (default public)",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			assert.Equal(t, tt.expected, config.IsPublic)
		})
	}
}

func TestParseFunctionConfig_DisableExecutionLogs(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected bool
	}{
		{
			name: "explicitly disabled with true",
			code: `
				// @fluxbase:disable-execution-logs true
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "explicitly disabled (no value)",
			code: `
				// @fluxbase:disable-execution-logs
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: true,
		},
		{
			name: "explicitly enabled with false",
			code: `
				// @fluxbase:disable-execution-logs false
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: false,
		},
		{
			name: "no directive (default enabled)",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			assert.Equal(t, tt.expected, config.DisableExecutionLogs)
		})
	}
}

func TestParseFunctionConfig_CorsOrigins(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *string
	}{
		{
			name: "single origin",
			code: `
				// @fluxbase:cors-origins https://example.com
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("https://example.com"),
		},
		{
			name: "multiple origins",
			code: `
				// @fluxbase:cors-origins https://example.com,https://api.example.com
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("https://example.com,https://api.example.com"),
		},
		{
			name: "wildcard",
			code: `
				// @fluxbase:cors-origins *
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("*"),
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			if tt.expected == nil {
				assert.Nil(t, config.CorsOrigins)
			} else {
				require.NotNil(t, config.CorsOrigins)
				assert.Equal(t, *tt.expected, *config.CorsOrigins)
			}
		})
	}
}

func TestParseFunctionConfig_CorsMethods(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *string
	}{
		{
			name: "single method",
			code: `
				// @fluxbase:cors-methods POST
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("POST"),
		},
		{
			name: "multiple methods",
			code: `
				// @fluxbase:cors-methods GET,POST,PUT,DELETE
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("GET,POST,PUT,DELETE"),
		},
		{
			name: "all methods",
			code: `
				// @fluxbase:cors-methods *
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("*"),
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			if tt.expected == nil {
				assert.Nil(t, config.CorsMethods)
			} else {
				require.NotNil(t, config.CorsMethods)
				assert.Equal(t, *tt.expected, *config.CorsMethods)
			}
		})
	}
}

func TestParseFunctionConfig_CorsHeaders(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *string
	}{
		{
			name: "single header",
			code: `
				// @fluxbase:cors-headers Content-Type
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("Content-Type"),
		},
		{
			name: "multiple headers",
			code: `
				// @fluxbase:cors-headers Content-Type,Authorization,X-Custom-Header
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: strPtr("Content-Type,Authorization,X-Custom-Header"),
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			if tt.expected == nil {
				assert.Nil(t, config.CorsHeaders)
			} else {
				require.NotNil(t, config.CorsHeaders)
				assert.Equal(t, *tt.expected, *config.CorsHeaders)
			}
		})
	}
}

func TestParseFunctionConfig_CorsCredentials(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *bool
	}{
		{
			name: "credentials allowed",
			code: `
				// @fluxbase:cors-credentials true
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: boolPtr(true),
		},
		{
			name: "credentials not allowed",
			code: `
				// @fluxbase:cors-credentials false
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: boolPtr(false),
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			if tt.expected == nil {
				assert.Nil(t, config.CorsCredentials)
			} else {
				require.NotNil(t, config.CorsCredentials)
				assert.Equal(t, *tt.expected, *config.CorsCredentials)
			}
		})
	}
}

func TestParseFunctionConfig_CorsMaxAge(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected *int
	}{
		{
			name: "max age set",
			code: `
				// @fluxbase:cors-max-age 3600
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: intPtr(3600),
		},
		{
			name: "max age zero",
			code: `
				// @fluxbase:cors-max-age 0
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: intPtr(0),
		},
		{
			name: "no directive",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			if tt.expected == nil {
				assert.Nil(t, config.CorsMaxAge)
			} else {
				require.NotNil(t, config.CorsMaxAge)
				assert.Equal(t, *tt.expected, *config.CorsMaxAge)
			}
		})
	}
}

func TestParseFunctionConfig_RateLimit(t *testing.T) {
	tests := []struct {
		name              string
		code              string
		expectedPerMinute *int
		expectedPerHour   *int
		expectedPerDay    *int
	}{
		{
			name: "rate limit per minute",
			code: `
				// @fluxbase:rate-limit 100/min
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expectedPerMinute: intPtr(100),
			expectedPerHour:   nil,
			expectedPerDay:    nil,
		},
		{
			name: "rate limit per hour",
			code: `
				// @fluxbase:rate-limit 1000/hour
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expectedPerMinute: nil,
			expectedPerHour:   intPtr(1000),
			expectedPerDay:    nil,
		},
		{
			name: "rate limit per day",
			code: `
				// @fluxbase:rate-limit 10000/day
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expectedPerMinute: nil,
			expectedPerHour:   nil,
			expectedPerDay:    intPtr(10000),
		},
		{
			name: "multiple rate limits",
			code: `
				// @fluxbase:rate-limit 100/min
				// @fluxbase:rate-limit 1000/hour
				// @fluxbase:rate-limit 10000/day
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expectedPerMinute: intPtr(100),
			expectedPerHour:   intPtr(1000),
			expectedPerDay:    intPtr(10000),
		},
		{
			name: "no rate limit",
			code: `
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			expectedPerMinute: nil,
			expectedPerHour:   nil,
			expectedPerDay:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)

			if tt.expectedPerMinute == nil {
				assert.Nil(t, config.RateLimitPerMinute)
			} else {
				require.NotNil(t, config.RateLimitPerMinute)
				assert.Equal(t, *tt.expectedPerMinute, *config.RateLimitPerMinute)
			}

			if tt.expectedPerHour == nil {
				assert.Nil(t, config.RateLimitPerHour)
			} else {
				require.NotNil(t, config.RateLimitPerHour)
				assert.Equal(t, *tt.expectedPerHour, *config.RateLimitPerHour)
			}

			if tt.expectedPerDay == nil {
				assert.Nil(t, config.RateLimitPerDay)
			} else {
				require.NotNil(t, config.RateLimitPerDay)
				assert.Equal(t, *tt.expectedPerDay, *config.RateLimitPerDay)
			}
		})
	}
}

func TestParseFunctionConfig_Complex(t *testing.T) {
	code := `
		// @fluxbase:allow-unauthenticated
		// @fluxbase:public false
		// @fluxbase:cors-origins https://example.com
		// @fluxbase:cors-methods GET,POST
		// @fluxbase:cors-headers Content-Type,Authorization
		// @fluxbase:cors-credentials true
		// @fluxbase:cors-max-age 3600
		// @fluxbase:rate-limit 100/min
		// @fluxbase:rate-limit 1000/hour
		// @fluxbase:disable-execution-logs true

		export default async function(req) {
			return new Response("Hello");
		}
	`

	config := ParseFunctionConfig(code)

	assert.True(t, config.AllowUnauthenticated)
	assert.False(t, config.IsPublic)
	assert.True(t, config.DisableExecutionLogs)
	require.NotNil(t, config.CorsOrigins)
	assert.Equal(t, "https://example.com", *config.CorsOrigins)
	require.NotNil(t, config.CorsMethods)
	assert.Equal(t, "GET,POST", *config.CorsMethods)
	require.NotNil(t, config.CorsHeaders)
	assert.Equal(t, "Content-Type,Authorization", *config.CorsHeaders)
	require.NotNil(t, config.CorsCredentials)
	assert.Equal(t, true, *config.CorsCredentials)
	require.NotNil(t, config.CorsMaxAge)
	assert.Equal(t, 3600, *config.CorsMaxAge)
	require.NotNil(t, config.RateLimitPerMinute)
	assert.Equal(t, 100, *config.RateLimitPerMinute)
	require.NotNil(t, config.RateLimitPerHour)
	assert.Equal(t, 1000, *config.RateLimitPerHour)
}

func TestParseFunctionConfig_WhitespaceHandling(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		testFunc func(t *testing.T, config FunctionConfig)
	}{
		{
			name: "extra spaces around values",
			code: `
				// @fluxbase:cors-origins    https://example.com
				// @fluxbase:cors-methods  GET , POST
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			testFunc: func(t *testing.T, config FunctionConfig) {
				require.NotNil(t, config.CorsOrigins)
				assert.Equal(t, "https://example.com", *config.CorsOrigins)
				require.NotNil(t, config.CorsMethods)
				assert.Equal(t, "GET , POST", *config.CorsMethods) // Note: doesn't trim internal spaces
			},
		},
		{
			name: "tabs instead of spaces",
			code: `
				//	@fluxbase:allow-unauthenticated
				export default async function(req) {
					return new Response("Hello");
				}
			`,
			testFunc: func(t *testing.T, config FunctionConfig) {
				assert.True(t, config.AllowUnauthenticated)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ParseFunctionConfig(tt.code)
			tt.testFunc(t, config)
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
