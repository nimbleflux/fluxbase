package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhooksList_Success(t *testing.T) {
	resetWebhookFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": "wh1", "url": "https://example.com/hook", "events": "INSERT,UPDATE", "enabled": true},
			{"id": "wh2", "url": "https://example.com/hook2", "events": "DELETE", "enabled": false},
		})
	})
	defer cleanup()

	err := runWebhooksList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "wh1", result[0]["id"])
	assert.Equal(t, "wh2", result[1]["id"])
}

func TestWebhooksList_Empty(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runWebhooksList(nil, []string{})
	require.NoError(t, err)
}

func TestWebhooksList_APIError(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runWebhooksList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestWebhooksGet_Success(t *testing.T) {
	resetWebhookFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks/wh123")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":      "wh123",
			"url":     "https://example.com/hook",
			"events":  "INSERT,UPDATE",
			"enabled": true,
		})
	})
	defer cleanup()

	err := runWebhooksGet(nil, []string{"wh123"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "wh123", result["id"])
	assert.Equal(t, "https://example.com/hook", result["url"])
}

func TestWebhooksGet_APIError(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "webhook not found")
	})
	defer cleanup()

	err := runWebhooksGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
}

func TestWebhooksCreate_Success(t *testing.T) {
	resetWebhookFlags()
	whURL = "https://example.com/hook"
	whEvents = "INSERT,UPDATE"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "https://example.com/hook", body["url"])
		assert.Equal(t, "INSERT,UPDATE", body["events"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":      "wh-new",
			"url":     "https://example.com/hook",
			"enabled": true,
		})
	})
	defer cleanup()

	err := runWebhooksCreate(nil, []string{})
	require.NoError(t, err)
}

func TestWebhooksCreate_APIError(t *testing.T) {
	resetWebhookFlags()
	whURL = "https://example.com/hook"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusBadRequest, "invalid URL")
	})
	defer cleanup()

	err := runWebhooksCreate(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid URL")
}

func TestWebhooksUpdate_Success(t *testing.T) {
	resetWebhookFlags()
	whURL = "https://new-url.com/hook"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks/wh123")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "https://new-url.com/hook", body["url"])

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&whEnabled, "enabled", true, "Enable/disable webhook")
	err := runWebhooksUpdate(cmd, []string{"wh123"})
	require.NoError(t, err)
}

func TestWebhooksUpdate_APIError(t *testing.T) {
	resetWebhookFlags()
	whURL = "https://new-url.com/hook"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "webhook not found")
	})
	defer cleanup()

	cmd := &cobra.Command{}
	cmd.Flags().BoolVar(&whEnabled, "enabled", true, "Enable/disable webhook")
	err := runWebhooksUpdate(cmd, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
}

func TestWebhooksDelete_Success(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks/wh123")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runWebhooksDelete(nil, []string{"wh123"})
	require.NoError(t, err)
}

func TestWebhooksDelete_APIError(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "webhook not found")
	})
	defer cleanup()

	err := runWebhooksDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
}

func TestWebhooksTest_Success(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks/wh123/test")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
		})
	})
	defer cleanup()

	err := runWebhooksTest(nil, []string{"wh123"})
	require.NoError(t, err)
}

func TestWebhooksTest_APIError(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "webhook not found")
	})
	defer cleanup()

	err := runWebhooksTest(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
}

func TestWebhooksDeliveries_Success(t *testing.T) {
	resetWebhookFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/webhooks/wh123/deliveries")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": "d1", "response_status": float64(200), "response_body": "OK", "created_at": "2024-01-01T00:00:00Z"},
			{"id": "d2", "response_status": float64(500), "response_body": "Internal Server Error", "created_at": "2024-01-02T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runWebhooksDeliveries(nil, []string{"wh123"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "d1", result[0]["id"])
	assert.Equal(t, "d2", result[1]["id"])
}

func TestWebhooksDeliveries_Empty(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runWebhooksDeliveries(nil, []string{"wh123"})
	require.NoError(t, err)
}

func TestWebhooksDeliveries_APIError(t *testing.T) {
	resetWebhookFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "webhook not found")
	})
	defer cleanup()

	err := runWebhooksDeliveries(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook not found")
}
