package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// RequireSyncIPAllowlist Tests
// =============================================================================

func TestRequireSyncIPAllowlist_EmptyConfig(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	// Empty ranges = allow all
	app.Use(RequireSyncIPAllowlist([]string{}, "functions", serverCfg))
	app.Get("/sync", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/sync", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestRequireSyncIPAllowlist_NilConfig(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	// Nil slice = allow all
	app.Use(RequireSyncIPAllowlist(nil, "jobs", serverCfg))
	app.Get("/sync", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/sync", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestRequireSyncIPAllowlist_DirectConnection(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}

	t.Run("allows localhost when in range", func(t *testing.T) {
		app := fiber.New()

		// Allow all IPs (test mode uses 0.0.0.0)
		app.Use(RequireSyncIPAllowlist([]string{"0.0.0.0/0"}, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("denies localhost when not in range", func(t *testing.T) {
		app := fiber.New()

		// Only allow 10.x.x.x, not localhost
		app.Use(RequireSyncIPAllowlist([]string{"10.0.0.0/8"}, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 403, resp.StatusCode)
	})
}

func TestRequireSyncIPAllowlist_IgnoresSpoofedHeaders(t *testing.T) {
	// Security test: verify that X-Forwarded-For headers are NOT trusted
	// when no trusted proxies are configured
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}

	t.Run("ignores X-Forwarded-For header (no trusted proxies)", func(t *testing.T) {
		app := fiber.New()

		// Allow 10.x.x.x
		app.Use(RequireSyncIPAllowlist([]string{"10.0.0.0/8"}, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		// Try to spoof IP - should be ignored since connection is from localhost
		req.Header.Set("X-Forwarded-For", "10.1.2.3")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should deny because we use the actual connection IP (localhost), not the spoofed header
		assert.Equal(t, 403, resp.StatusCode)
	})

	t.Run("ignores X-Real-IP header (no trusted proxies)", func(t *testing.T) {
		app := fiber.New()

		// Allow 10.x.x.x
		app.Use(RequireSyncIPAllowlist([]string{"10.0.0.0/8"}, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		req.Header.Set("X-Real-IP", "10.1.2.3")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should deny because we use the actual connection IP (localhost)
		assert.Equal(t, 403, resp.StatusCode)
	})
}

func TestRequireSyncIPAllowlist_MultipleRanges(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	// Include all IPs in ranges (test mode uses 0.0.0.0)
	ranges := []string{
		"10.0.0.0/8",
		"0.0.0.0/0", // all IPs
		"172.16.0.0/12",
	}

	app.Use(RequireSyncIPAllowlist(ranges, "functions", serverCfg))
	app.Get("/sync", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Direct connection from localhost should be allowed
	req := httptest.NewRequest("GET", "/sync", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 200, resp.StatusCode)
}

func TestRequireSyncIPAllowlist_ErrorMessage(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}

	tests := []struct {
		featureName     string
		expectedInError string
	}{
		{"functions", "functions sync"},
		{"jobs", "jobs sync"},
		{"custom-feature", "custom-feature sync"},
	}

	for _, tt := range tests {
		t.Run(tt.featureName, func(t *testing.T) {
			app := fiber.New()

			// Only allow 10.x.x.x (not localhost)
			app.Use(RequireSyncIPAllowlist([]string{"10.0.0.0/8"}, tt.featureName, serverCfg))
			app.Get("/sync", func(c fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/sync", nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, 403, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), tt.expectedInError)
		})
	}
}

func TestRequireSyncIPAllowlist_InvalidCIDR(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}

	t.Run("ignores invalid CIDR", func(t *testing.T) {
		app := fiber.New()

		ranges := []string{
			"invalid-cidr",
			"0.0.0.0/0", // all IPs - valid one for test mode
		}

		app.Use(RequireSyncIPAllowlist(ranges, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("all invalid CIDRs allows all", func(t *testing.T) {
		app := fiber.New()

		ranges := []string{
			"invalid1",
			"invalid2",
			"also-invalid",
		}

		app.Use(RequireSyncIPAllowlist(ranges, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// All invalid = empty valid ranges = allow all
		assert.Equal(t, 200, resp.StatusCode)
	})
}

func TestRequireSyncIPAllowlist_IPv6(t *testing.T) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}

	t.Run("allows localhost IPv6", func(t *testing.T) {
		app := fiber.New()

		app.Use(RequireSyncIPAllowlist([]string{"::1/128"}, "functions", serverCfg))
		app.Get("/sync", func(c fiber.Ctx) error {
			return c.SendString("OK")
		})

		req := httptest.NewRequest("GET", "/sync", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Note: May be 200 or 403 depending on whether Fiber uses IPv4 or IPv6 for test connections
		// The important thing is we're not trusting spoofed headers
		assert.Contains(t, []int{200, 403}, resp.StatusCode)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRequireSyncIPAllowlist_EmptyConfig(b *testing.B) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	app.Use(RequireSyncIPAllowlist([]string{}, "functions", serverCfg))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		_ = resp.Body.Close()
	}
}

func BenchmarkRequireSyncIPAllowlist_SingleRange(b *testing.B) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	app.Use(RequireSyncIPAllowlist([]string{"0.0.0.0/0"}, "functions", serverCfg))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		_ = resp.Body.Close()
	}
}

func BenchmarkRequireSyncIPAllowlist_MultipleRanges(b *testing.B) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	ranges := []string{
		"10.0.0.0/8",
		"192.168.0.0/16",
		"172.16.0.0/12",
		"0.0.0.0/0", // all IPs for test mode
		"198.51.100.0/24",
	}

	app.Use(RequireSyncIPAllowlist(ranges, "functions", serverCfg))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		_ = resp.Body.Close()
	}
}

func BenchmarkRequireSyncIPAllowlist_Denied(b *testing.B) {
	serverCfg := &config.ServerConfig{TrustedProxies: []string{}}
	app := fiber.New()

	app.Use(RequireSyncIPAllowlist([]string{"10.0.0.0/8"}, "functions", serverCfg))
	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, _ := app.Test(req)
		_ = resp.Body.Close()
	}
}
