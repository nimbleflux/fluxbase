package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Realtime Stats ---

func TestRealtimeStats_Success(t *testing.T) {
	resetRealtimeFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/realtime/stats")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"total_connections":   float64(42),
			"total_subscriptions": float64(128),
			"channels": map[string]interface{}{
				"notifications": map[string]interface{}{
					"subscribers": float64(15),
				},
			},
		})
	})
	defer cleanup()

	err := runRealtimeStats(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, float64(42), result["total_connections"])
	assert.Equal(t, float64(128), result["total_subscriptions"])
}

func TestRealtimeStats_APIError(t *testing.T) {
	resetRealtimeFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "realtime service unavailable")
	})
	defer cleanup()

	err := runRealtimeStats(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "realtime service unavailable")
}

// --- Realtime Broadcast ---

func TestRealtimeBroadcast_Success(t *testing.T) {
	resetRealtimeFlags()
	rtMessage = `{"type": "notification", "text": "Hello!"}`
	rtEvent = "broadcast"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/realtime/broadcast")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "my-channel", body["channel"])
		assert.Equal(t, "broadcast", body["event"])
		assert.Equal(t, `{"type": "notification", "text": "Hello!"}`, body["payload"])

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runRealtimeBroadcast(nil, []string{"my-channel"})
	require.NoError(t, err)
}

func TestRealtimeBroadcast_CustomEvent(t *testing.T) {
	resetRealtimeFlags()
	rtMessage = `{"data": "value"}`
	rtEvent = "custom_event"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "updates", body["channel"])
		assert.Equal(t, "custom_event", body["event"])
		assert.Equal(t, `{"data": "value"}`, body["payload"])

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runRealtimeBroadcast(nil, []string{"updates"})
	require.NoError(t, err)
}

func TestRealtimeBroadcast_APIError(t *testing.T) {
	resetRealtimeFlags()
	rtMessage = `{"data": "test"}`
	rtEvent = "broadcast"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusBadRequest, "invalid channel name")
	})
	defer cleanup()

	err := runRealtimeBroadcast(nil, []string{"invalid channel!"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid channel name")
}
