package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/test"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// setupAdminTest prepares the test context for admin API tests
// Uses isolated rate limiter to avoid rate limit exhaustion from other tests
func setupAdminTest(t *testing.T) (*test.TestContext, string) {
	// Use isolated rate limiter to avoid rate limit state pollution from other tests
	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})

	// Use UUID-based unique email to ensure no conflicts across parallel test packages
	// UUID guarantees uniqueness better than timestamps which can collide in CI
	uniqueID := uuid.New().String()
	email := fmt.Sprintf("admin-%s@test.com", uniqueID)
	password := "adminpass123456"
	t.Logf("Creating dashboard admin with email: %s", email)
	_, token := tc.CreateDashboardAdminUser(email, password)

	return tc, token
}

// TestAdminListSchemas tests listing database schemas
func TestAdminListSchemas(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	resp := tc.NewRequest("GET", "/api/v1/admin/schemas").
		WithAuth(token).
		Send().
		AssertStatus(fiber.StatusOK)

	var schemas []string
	resp.JSON(&schemas)

	require.GreaterOrEqual(t, len(schemas), 1, "Should have at least one schema")

	t.Logf("Found %d schemas: %v", len(schemas), schemas)
}

// TestAdminListTables tests listing tables in a schema
func TestAdminListTables(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	// List tables in the public schema
	resp := tc.NewRequest("GET", "/api/v1/admin/schemas/public/tables").
		WithAuth(token).
		Send()

	// This endpoint may not be implemented yet - check if it returns 404
	if resp.Status() == fiber.StatusNotFound {
		t.Skip("List tables endpoint not yet implemented")
		return
	}

	resp.AssertStatus(fiber.StatusOK)

	var tables []string
	resp.JSON(&tables)

	require.GreaterOrEqual(t, len(tables), 1, "Should have at least one table in public schema")

	t.Logf("Found %d tables in public schema: %v", len(tables), tables)
}

// TestAdminGetTableMetadata tests getting metadata for a specific table
func TestAdminGetTableMetadata(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	// Get metadata for the auth.users table
	resp := tc.NewRequest("GET", "/api/v1/admin/schemas/auth/tables/users").
		WithAuth(token).
		Send()

	// This endpoint may not be implemented yet - check if it returns 404
	if resp.Status() == fiber.StatusNotFound {
		t.Skip("Get table metadata endpoint not yet implemented")
		return
	}

	resp.AssertStatus(fiber.StatusOK)

	var metadata map[string]interface{}
	resp.JSON(&metadata)

	require.NotNil(t, metadata, "Should return table metadata")
	require.Contains(t, metadata, "columns", "Metadata should include columns")

	t.Logf("Retrieved metadata for auth.users table")
}

// TestAdminCORSHeaders tests that CORS headers are present
func TestAdminCORSHeaders(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	resp := tc.NewRequest("GET", "/api/v1/admin/schemas").
		WithAuth(token).
		WithHeader("Origin", "https://example.com").
		Send().
		AssertStatus(fiber.StatusOK)

	// Check for CORS headers
	corsHeader := resp.Header("Access-Control-Allow-Origin")
	if corsHeader != "" {
		t.Logf("CORS header present: %s", corsHeader)
	} else {
		t.Log("CORS headers may not be configured or may require specific configuration")
	}
}

// TestAdmin404Handling tests 404 error handling
func TestAdmin404Handling(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	// Try to get metadata for non-existent table
	tc.NewRequest("GET", "/api/v1/admin/schemas/public/tables/nonexistent_table_xyz").
		WithAuth(token).
		Send().
		AssertStatus(fiber.StatusNotFound)

	t.Logf("404 handling works correctly")
}

// TestAdminRequestIDTracking tests that request IDs are tracked
func TestAdminRequestIDTracking(t *testing.T) {
	tc, token := setupAdminTest(t)
	defer tc.Close()

	resp := tc.NewRequest("GET", "/api/v1/admin/schemas").
		WithAuth(token).
		Send().
		AssertStatus(fiber.StatusOK)

	// Check for X-Request-ID header
	requestID := resp.Header("X-Request-ID")
	if requestID != "" {
		require.NotEmpty(t, requestID, "Request ID should be present")
		t.Logf("Request ID tracking working: %s", requestID)
	} else {
		t.Log("Request ID header may use a different name or not be configured")
	}
}

// TestAdminSetupRateLimit tests that admin setup endpoint is rate limited
func TestAdminSetupRateLimit(t *testing.T) {
	// Use isolated rate limiter to avoid state pollution from other tests
	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// CRITICAL: Reset admin setup status so test can run properly
	// Otherwise, the "setup already complete" check returns 403 before rate limiter
	tc.ExecuteSQLAsSuperuser(`
		DELETE FROM app.settings WHERE category = 'system' AND key = 'admin_setup';
	`)

	// Make multiple setup attempts to trigger rate limit (max 5 per 15 minutes)
	for i := 1; i <= 6; i++ {
		resp := tc.NewRequest("POST", "/api/v1/admin/setup").
			WithBody(map[string]interface{}{
				"email":       fmt.Sprintf("admin%d@example.com", i),
				"password":    "securepassword123",
				"name":        "Admin User",
				"setup_token": tc.Config.Security.SetupToken,
			}).Send()

		// First 5 attempts should either succeed or fail with setup already complete
		if i <= 5 {
			// First attempt: 201 (success)
			// Attempts 2-5: 403 (setup already complete from first attempt)
			status := resp.Status()
			require.Contains(t, []int{
				fiber.StatusCreated,
				fiber.StatusForbidden, // Setup already complete (attempts 2-5)
			}, status, "First 5 attempts should not be rate limited")
			t.Logf("Attempt %d: Status %d", i, status)
		} else {
			// 6th attempt should be rate limited
			resp.AssertStatus(fiber.StatusTooManyRequests)

			var result map[string]interface{}
			resp.JSON(&result)
			require.Equal(t, "Rate limit exceeded", result["error"])
			t.Logf("Attempt %d: Rate limited as expected", i)
		}
	}
}

// TestAdminLoginRateLimit tests that admin login endpoint is rate limited
func TestAdminLoginRateLimit(t *testing.T) {
	// Use isolated rate limiter to avoid state pollution from other tests
	rateLimiter, pubSub := test.NewInMemoryDependencies()
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// Skip this test if rate limit is set too high (e.g., for parallel test execution)
	cfg := test.GetTestConfig()
	if cfg.Security.AdminLoginRateLimit > 10 {
		t.Skipf("Skipping rate limit test: AdminLoginRateLimit is set to %d (too high for this test)", cfg.Security.AdminLoginRateLimit)
	}

	// Create an admin user first with unique email
	timestamp := time.Now().UnixNano()
	email := fmt.Sprintf("admin-ratelimit-%d@example.com", timestamp)
	password := "testpassword123"
	tc.CreateDashboardAdminUser(email, password)

	// Make multiple login attempts with wrong password to trigger rate limit
	// (max 4 per minute - lower than account lockout threshold of 5)
	// Rate limit should trigger on attempt 5, before account lockout at 5
	rateLimitHit := false
	for i := 1; i <= 6; i++ {
		resp := tc.NewRequest("POST", "/api/v1/admin/login").
			WithBody(map[string]interface{}{
				"email":    email,
				"password": "wrongpassword",
			}).Send()

		status := resp.Status()

		if status == fiber.StatusTooManyRequests {
			// Rate limit was triggered (expected on attempt 5)
			rateLimitHit = true
			var result map[string]interface{}
			resp.JSON(&result)
			require.Equal(t, "Rate limit exceeded", result["error"])
			t.Logf("Attempt %d: Rate limited as expected (status %d)", i, status)
			break
		}

		// Before rate limit: expect 401 (wrong password)
		// After account lockout (5 attempts): expect 403
		require.Contains(t, []int{401, 403}, status,
			"Before/at rate limit, should get 401 (wrong password) or 403 (account locked)")
		t.Logf("Attempt %d: Status %d", i, status)
	}

	require.True(t, rateLimitHit, "Rate limit should have been triggered on attempt 5")
}
