package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceKeysList_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": "sk1", "name": "Production", "key_prefix": "fb_prod_", "scopes": []interface{}{"*"}, "enabled": true, "created_at": "2024-01-01T00:00:00Z"},
			{"id": "sk2", "name": "Staging", "key_prefix": "fb_stag_", "scopes": []interface{}{"read"}, "enabled": false, "created_at": "2024-02-01T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runServiceKeysList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "sk1", result[0]["id"])
	assert.Equal(t, "sk2", result[1]["id"])
}

func TestServiceKeysList_Empty(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runServiceKeysList(nil, []string{})
	require.NoError(t, err)
}

func TestServiceKeysList_APIError(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runServiceKeysList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestServiceKeysCreate_Success(t *testing.T) {
	resetServiceKeyFlags()
	skName = "Migrations Key"
	skDescription = "Key for running migrations"
	skScopes = "migrations:*,tables:read"
	skRateLimitPerMinute = 100
	skRateLimitPerHour = 5000

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "Migrations Key", body["name"])
		assert.Equal(t, "Key for running migrations", body["description"])

		scopes, ok := body["scopes"].([]interface{})
		require.True(t, ok)
		assert.Contains(t, scopes, "migrations:*")
		assert.Contains(t, scopes, "tables:read")

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":         "sk-new",
			"name":       "Migrations Key",
			"key":        "fb_prod_abc123xyz",
			"key_prefix": "fb_prod_",
		})
	})
	defer cleanup()

	err := runServiceKeysCreate(nil, []string{})
	require.NoError(t, err)
}

func TestServiceKeysCreate_Minimal(t *testing.T) {
	resetServiceKeyFlags()
	skName = "Simple Key"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "Simple Key", body["name"])
		_, hasDesc := body["description"]
		assert.False(t, hasDesc)
		_, hasScopes := body["scopes"]
		assert.False(t, hasScopes)

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":         "sk-new",
			"name":       "Simple Key",
			"key":        "fb_prod_def456",
			"key_prefix": "fb_prod_",
		})
	})
	defer cleanup()

	err := runServiceKeysCreate(nil, []string{})
	require.NoError(t, err)
}

func TestServiceKeysCreate_APIError(t *testing.T) {
	resetServiceKeyFlags()
	skName = "Test Key"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusBadRequest, "name is required")
	})
	defer cleanup()

	err := runServiceKeysCreate(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestServiceKeysGet_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":         "sk1",
			"name":       "Production",
			"key_prefix": "fb_prod_",
			"scopes":     []interface{}{"*"},
			"enabled":    true,
		})
	})
	defer cleanup()

	err := runServiceKeysGet(nil, []string{"sk1"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "sk1", result["id"])
	assert.Equal(t, "Production", result["name"])
}

func TestServiceKeysGet_NotFound(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysUpdate_Success(t *testing.T) {
	resetServiceKeyFlags()
	skName = "New Name"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	// runServiceKeysUpdate uses cmd.Flags().Changed(), requires a real cobra command
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&skName, "name", "", "")
	cmd.Flags().StringVar(&skDescription, "description", "", "")
	cmd.Flags().StringVar(&skScopes, "scopes", "", "")
	cmd.Flags().IntVar(&skRateLimitPerMinute, "rate-limit-per-minute", 0, "")
	cmd.Flags().IntVar(&skRateLimitPerHour, "rate-limit-per-hour", 0, "")
	cmd.Flags().BoolVar(&skEnabled, "enabled", true, "")
	_ = cmd.Flags().Set("name", "New Name")

	err := runServiceKeysUpdate(cmd, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysUpdate_NoUpdates(t *testing.T) {
	resetServiceKeyFlags()

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&skName, "name", "", "")
	cmd.Flags().StringVar(&skDescription, "description", "", "")
	cmd.Flags().StringVar(&skScopes, "scopes", "", "")
	cmd.Flags().IntVar(&skRateLimitPerMinute, "rate-limit-per-minute", 0, "")
	cmd.Flags().IntVar(&skRateLimitPerHour, "rate-limit-per-hour", 0, "")
	cmd.Flags().BoolVar(&skEnabled, "enabled", true, "")

	err := runServiceKeysUpdate(cmd, []string{"sk1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no fields to update")
}

func TestServiceKeysUpdate_APIError(t *testing.T) {
	resetServiceKeyFlags()
	skName = "Updated"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&skName, "name", "", "")
	cmd.Flags().StringVar(&skDescription, "description", "", "")
	cmd.Flags().StringVar(&skScopes, "scopes", "", "")
	cmd.Flags().IntVar(&skRateLimitPerMinute, "rate-limit-per-minute", 0, "")
	cmd.Flags().IntVar(&skRateLimitPerHour, "rate-limit-per-hour", 0, "")
	cmd.Flags().BoolVar(&skEnabled, "enabled", true, "")
	_ = cmd.Flags().Set("name", "Updated")

	err := runServiceKeysUpdate(cmd, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysDisable_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/disable")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runServiceKeysDisable(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysDisable_APIError(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysDisable(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysEnable_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/enable")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runServiceKeysEnable(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysEnable_APIError(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysEnable(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysDelete_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runServiceKeysDelete(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysDelete_APIError(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysRevoke_Success(t *testing.T) {
	resetServiceKeyFlags()
	skRevokeReason = "Key compromised"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/revoke")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "Key compromised", body["reason"])

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runServiceKeysRevoke(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysRevoke_APIError(t *testing.T) {
	resetServiceKeyFlags()
	skRevokeReason = "test"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysRevoke(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysDeprecate_Success(t *testing.T) {
	resetServiceKeyFlags()
	skGracePeriod = "24h"
	skRevokeReason = "Scheduled rotation"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/deprecate")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "24h", body["grace_period"])
		assert.Equal(t, "Scheduled rotation", body["reason"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"grace_period_ends_at": "2024-12-31T23:59:59Z",
		})
	})
	defer cleanup()

	err := runServiceKeysDeprecate(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysDeprecate_APIError(t *testing.T) {
	resetServiceKeyFlags()
	skGracePeriod = "24h"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysDeprecate(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysRotate_Success(t *testing.T) {
	resetServiceKeyFlags()
	skGracePeriod = "7d"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/rotate")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "7d", body["grace_period"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":                   "sk-new",
			"key":                  "fb_prod_rotated_key",
			"key_prefix":           "fb_prod_",
			"grace_period_ends_at": "2024-12-31T23:59:59Z",
		})
	})
	defer cleanup()

	err := runServiceKeysRotate(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysRotate_APIError(t *testing.T) {
	resetServiceKeyFlags()
	skGracePeriod = "24h"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysRotate(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}

func TestServiceKeysRevocations_Success(t *testing.T) {
	resetServiceKeyFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/service-keys/sk1/revocations")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{
				"revocation_type": "emergency",
				"reason":          "Key compromised",
				"revoked_by":      "admin",
				"created_at":      "2024-06-01T10:00:00Z",
			},
		})
	})
	defer cleanup()

	err := runServiceKeysRevocations(nil, []string{"sk1"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "emergency", result[0]["revocation_type"])
}

func TestServiceKeysRevocations_Empty(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runServiceKeysRevocations(nil, []string{"sk1"})
	require.NoError(t, err)
}

func TestServiceKeysRevocations_APIError(t *testing.T) {
	resetServiceKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "service key not found")
	})
	defer cleanup()

	err := runServiceKeysRevocations(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "service key not found")
}
