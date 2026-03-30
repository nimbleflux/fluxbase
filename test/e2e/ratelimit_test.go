//go:build integration

package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

// TestRateLimitHeaders verifies that rate limit response headers are present after
// exceeding the rate limit on a low-limit endpoint.
//
// The admin_setup endpoint has a rate limit of 5 requests per 15 minutes (per IP).
// We use an isolated rate limiter so this test doesn't interfere with others.
func TestRateLimitHeaders(t *testing.T) {
	// Skip in CI - requires isolated server instance
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI: test requires isolated server instance")
	}

	// Use isolated rate limiter to avoid state pollution from other tests
	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// Reset admin setup status so we can trigger the endpoint
	tc.ExecuteSQLAsSuperuser(`DELETE FROM app.settings WHERE category = 'system' AND key = 'admin_setup'`)

	// Exhaust the rate limit (5 requests max)
	for i := 0; i < 5; i++ {
		tc.NewRequest("POST", "/api/v1/admin/setup").
			WithBody(map[string]interface{}{
				"email":       fmt.Sprintf("rl-test-%d@example.com", i),
				"password":    "securepassword123",
				"name":        "Rate Limit Test",
				"setup_token": tc.Config.Security.SetupToken,
			}).Send()
	}

	// The 6th request should be rate limited with Retry-After header
	resp := tc.NewRequest("POST", "/api/v1/admin/setup").
		WithBody(map[string]interface{}{
			"email":       "rl-test-overflow@example.com",
			"password":    "securepassword123",
			"name":        "Rate Limit Test",
			"setup_token": tc.Config.Security.SetupToken,
		}).Send()

	resp.AssertStatus(fiber.StatusTooManyRequests)

	// Verify Retry-After header is present
	retryAfter := resp.Header("Retry-After")
	assert.NotEmpty(t, retryAfter, "Retry-After header should be present")

	var result map[string]interface{}
	resp.JSON(&result)
	assert.Equal(t, "Rate limit exceeded", result["error"])
}

// TestRateLimitExceeded verifies that a rate-limited endpoint returns 429
// and the error body contains the expected fields.
func TestRateLimitExceeded(t *testing.T) {
	// Skip in CI - requires isolated server instance
	if os.Getenv("CI") != "" {
		t.Skip("Skipping in CI: test requires isolated server instance")
	}

	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// Reset admin setup status
	tc.ExecuteSQLAsSuperuser(`DELETE FROM app.settings WHERE category = 'system' AND key = 'admin_setup'`)

	// Exhaust the rate limit
	for i := 0; i < 5; i++ {
		tc.NewRequest("POST", "/api/v1/admin/setup").
			WithBody(map[string]interface{}{
				"email":       fmt.Sprintf("rl-exceeded-%d@example.com", i),
				"password":    "securepassword123",
				"name":        "Rate Limit Exceeded Test",
				"setup_token": tc.Config.Security.SetupToken,
			}).Send()
	}

	// Next request should return 429
	resp := tc.NewRequest("POST", "/api/v1/admin/setup").
		WithBody(map[string]interface{}{
			"email":       "rl-exceeded-overflow@example.com",
			"password":    "securepassword123",
			"name":        "Rate Limit Exceeded Test",
			"setup_token": tc.Config.Security.SetupToken,
		}).Send()

	require.Equal(t, fiber.StatusTooManyRequests, resp.Status())

	var result map[string]interface{}
	resp.JSON(&result)
	assert.Equal(t, "RATE_LIMIT_EXCEEDED", result["code"])
	assert.Equal(t, "Rate limit exceeded", result["error"])
}

// TestRateLimitOnPublicEndpoint verifies that the admin login endpoint
// respects rate limiting with a low default limit (4/minute per IP).
//
// This test uses the default test config which sets AdminLoginRateLimit to 10000,
// so it skips when the limit is too high for a practical test.
func TestRateLimitOnPublicEndpoint(t *testing.T) {
	cfg := test.GetTestConfig()
	if cfg.Security.AdminLoginRateLimit > 10 {
		t.Skipf("Skipping: AdminLoginRateLimit is %d (too high for rate limit test)", cfg.Security.AdminLoginRateLimit)
	}

	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// Create an admin user
	email := test.E2ETestEmail()
	password := "testpassword123"
	tc.CreateDashboardAdminUser(email, password)

	// Send login attempts with wrong password until rate limited
	rateLimitHit := false
	for i := 0; i < 10; i++ {
		resp := tc.NewRequest("POST", "/api/v1/admin/login").
			WithBody(map[string]interface{}{
				"email":    email,
				"password": "wrongpassword",
			}).Send()

		if resp.Status() == fiber.StatusTooManyRequests {
			rateLimitHit = true
			retryAfter := resp.Header("Retry-After")
			assert.NotEmpty(t, retryAfter, "Retry-After header should be present on 429")

			var result map[string]interface{}
			resp.JSON(&result)
			assert.Equal(t, "Rate limit exceeded", result["error"])
			break
		}
	}

	assert.True(t, rateLimitHit, "Should have hit rate limit within 10 attempts")
}
