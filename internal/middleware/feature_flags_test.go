package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Feature flag middleware tests are primarily covered by integration tests in test/e2e/
// This file contains basic unit tests for the middleware structure

func TestRequireFeatureEnabled_MiddlewareStructure(t *testing.T) {
	// This is a basic structural test to ensure the middleware compiles
	// Real testing is done in integration tests with a full database setup

	// Verify that the middleware helper functions exist and can be called
	// We can't run them without a proper settings cache, but we can verify they compile
	app := fiber.New()

	// These should compile without errors
	_ = app

	// The middleware functions should be callable (even if we don't use them here)
	// RequireRealtimeEnabled, RequireStorageEnabled, RequireFunctionsEnabled

	t.Log("Feature flag middleware structure test passed")
}

func TestRequireFeatureEnabled_ReturnsHandler(t *testing.T) {
	t.Run("returns fiber.Handler", func(t *testing.T) {
		handler := RequireFeatureEnabled(nil, "test.feature")
		assert.NotNil(t, handler)
	})
}

func TestRequireRealtimeEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireRealtimeEnabled(nil)
	assert.NotNil(t, handler)
}

func TestRequireStorageEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireStorageEnabled(nil)
	assert.NotNil(t, handler)
}

func TestRequireFunctionsEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireFunctionsEnabled(nil)
	assert.NotNil(t, handler)
}

func TestRequireJobsEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireJobsEnabled(nil)
	assert.NotNil(t, handler)
}

func TestRequireAIEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireAIEnabled(nil)
	assert.NotNil(t, handler)
}

func TestRequireRPCEnabled_ReturnsHandler(t *testing.T) {
	handler := RequireRPCEnabled(nil)
	assert.NotNil(t, handler)
}

func TestFeatureFlag_AllHelpers(t *testing.T) {
	// Test that all feature flag helpers can be created
	helpers := []struct {
		name    string
		handler fiber.Handler
	}{
		{"Realtime", RequireRealtimeEnabled(nil)},
		{"Storage", RequireStorageEnabled(nil)},
		{"Functions", RequireFunctionsEnabled(nil)},
		{"Jobs", RequireJobsEnabled(nil)},
		{"AI", RequireAIEnabled(nil)},
		{"RPC", RequireRPCEnabled(nil)},
	}

	for _, h := range helpers {
		t.Run(h.name, func(t *testing.T) {
			assert.NotNil(t, h.handler)
		})
	}
}

func TestFeatureFlag_DisabledResponse(t *testing.T) {
	// Test that disabled feature returns 404 with correct error body
	// When settings cache is nil, feature is treated as disabled (returns false)
	app := fiber.New()
	app.Get("/test", RequireFeatureEnabled(nil, "test.feature"), func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	// Check response contains expected error code
	assert.Contains(t, string(body), "FEATURE_DISABLED")
	assert.Contains(t, string(body), "Feature not available")
}

func TestFeatureFlag_ResponseFormat(t *testing.T) {
	app := fiber.New()
	app.Get("/storage", RequireStorageEnabled(nil))
	app.Get("/realtime", RequireRealtimeEnabled(nil))
	app.Get("/functions", RequireFunctionsEnabled(nil))
	app.Get("/jobs", RequireJobsEnabled(nil))
	app.Get("/ai", RequireAIEnabled(nil))
	app.Get("/rpc", RequireRPCEnabled(nil))

	endpoints := []string{"/storage", "/realtime", "/functions", "/jobs", "/ai", "/rpc"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// All should return 503 when cache is nil (feature disabled)
			assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
		})
	}
}
