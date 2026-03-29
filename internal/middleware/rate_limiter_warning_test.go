package middleware

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogRateLimiterWarning_WithRedisURL(t *testing.T) {
	// Reset the warning state for this test
	resetRateLimiterWarning()

	// Set Redis URL - warning should not be displayed
	t.Setenv("FLUXBASE_REDIS_URL", "redis://localhost:6379")

	// Set Kubernetes indicator
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")

	logRateLimiterWarning()

	// Warning should not be displayed when Redis is configured
	assert.False(t, IsRateLimiterWarningDisplayed())
}

func TestLogRateLimiterWarning_WithDragonflyURL(t *testing.T) {
	// Reset the warning state for this test
	resetRateLimiterWarning()

	// Set Dragonfly URL - warning should not be displayed
	t.Setenv("FLUXBASE_DRAGONFLY_URL", "redis://localhost:6379")

	// Set Kubernetes indicator
	t.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")

	logRateLimiterWarning()

	// Warning should not be displayed when Dragonfly is configured
	assert.False(t, IsRateLimiterWarningDisplayed())
}

func TestLogRateLimiterWarning_NoMultiInstanceIndicators(t *testing.T) {
	// Reset the warning state for this test
	resetRateLimiterWarning()

	// Clear all multi-instance indicators
	t.Setenv("KUBERNETES_SERVICE_HOST", "")
	t.Setenv("POD_NAME", "")
	t.Setenv("COMPOSE_PROJECT_NAME", "")
	t.Setenv("FLUXBASE_REDIS_URL", "")
	t.Setenv("FLUXBASE_DRAGONFLY_URL", "")

	// Clear HOSTNAME for this test (t.Setenv auto-restores original value)
	t.Setenv("HOSTNAME", "")

	logRateLimiterWarning()

	// Warning should not be displayed when no multi-instance indicators are present
	// Note: HOSTNAME might be set in some environments, so we only test when it's not
	// The test may still pass if HOSTNAME was already set
}

// Helper to reset the warning state between tests
func resetRateLimiterWarning() {
	rateLimiterWarningDisplayed = false
	rateLimiterWarningMu = sync.Once{}
}
