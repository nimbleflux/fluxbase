//go:build integration

package e2e

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

// TestRealtimeStats tests GET /api/v1/realtime/stats returns 200 with connection count.
//
// Note: WebSocket testing (/realtime) requires special tooling (e.g. gorilla/websocket
// or nhooyr.io/websocket) and is not covered here. Only REST endpoints are tested.
func TestRealtimeStats(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// Create a test user to get an auth token
	email := test.E2ETestEmail()
	password := "testpassword123"
	_, token := tc.CreateTestUser(email, password)
	require.NotEmpty(t, token, "Should receive token from signup")

	// Realtime is disabled by default in test config; skip if not enabled
	if !tc.Config.Realtime.Enabled {
		t.Skip("Realtime is not enabled in test config")
	}

	// Request realtime stats with authentication
	resp := tc.NewRequest("GET", "/api/v1/realtime/stats").
		WithAuth(token).
		Send()

	// The endpoint should return 200 OK (or 503 if realtime subsystem not fully initialized)
	status := resp.Status()
	require.True(t, status == fiber.StatusOK || status == fiber.StatusServiceUnavailable,
		"Expected 200 or 503, got %d. Body: %s", status, string(resp.Body()))

	// If 200, verify response structure contains connection info
	if status == fiber.StatusOK {
		var result map[string]interface{}
		resp.JSON(&result)
		_, hasConnections := result["connections"]
		require.True(t, hasConnections, "Response should contain 'connections' field")
	}
}

// TestRealtimeBroadcast tests POST /api/v1/realtime/broadcast with valid message returns 200.
func TestRealtimeBroadcast(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	// Create a test user to get an auth token
	email := test.E2ETestEmail()
	password := "testpassword123"
	_, token := tc.CreateTestUser(email, password)
	require.NotEmpty(t, token, "Should receive token from signup")

	if !tc.Config.Realtime.Enabled {
		t.Skip("Realtime is not enabled in test config")
	}

	// Broadcast a message with authentication
	resp := tc.NewRequest("POST", "/api/v1/realtime/broadcast").
		WithAuth(token).
		WithBody(map[string]interface{}{
			"channel": "test-channel",
			"message": map[string]interface{}{
				"text": "Hello from E2E test",
			},
		}).
		Send()

	// The endpoint should succeed or fail gracefully (not 500)
	status := resp.Status()
	require.True(t, status == fiber.StatusOK || status == fiber.StatusServiceUnavailable,
		"Expected 200 or 503 (if realtime not fully initialized), got %d. Body: %s",
		status, string(resp.Body()))
}

// TestRealtimeBroadcastUnauthorized tests POST without auth returns 401.
func TestRealtimeBroadcastUnauthorized(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	if !tc.Config.Realtime.Enabled {
		t.Skip("Realtime is not enabled in test config")
	}

	// Attempt to broadcast without authentication
	tc.NewRequest("POST", "/api/v1/realtime/broadcast").
		WithBody(map[string]interface{}{
			"channel": "test-channel",
			"message": map[string]interface{}{
				"text": "Unauthorized broadcast",
			},
		}).
		Send().
		AssertStatus(fiber.StatusUnauthorized)
}

// TestRealtimeStatsUnauthorized tests GET stats without auth returns 401.
func TestRealtimeStatsUnauthorized(t *testing.T) {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	defer tc.Close()

	if !tc.Config.Realtime.Enabled {
		t.Skip("Realtime is not enabled in test config")
	}

	// Attempt to get stats without authentication
	tc.NewRequest("GET", "/api/v1/realtime/stats").
		Send().
		AssertStatus(fiber.StatusUnauthorized)
}
