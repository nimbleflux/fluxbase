package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Settings (system settings) ---

func TestSettingsList_Success(t *testing.T) {
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/system/settings")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"site_name":       "My App",
			"max_upload_size": float64(10485760),
		})
	})
	defer cleanup()

	err := runSettingsList(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Contains(t, result, "site_name")
}

func TestSettingsGet_Success(t *testing.T) {
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/system/settings")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"auth": map[string]interface{}{
				"signup_enabled": true,
			},
		})
	})
	defer cleanup()

	err := runSettingsGet(nil, []string{"auth.signup_enabled"})
	require.NoError(t, err)
}

func TestSettingsSet_Success(t *testing.T) {
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v1/admin/system/settings")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "site_name", body["key"])
		assert.Equal(t, "New Name", body["value"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"key": "site_name", "value": "New Name",
		})
	})
	defer cleanup()

	err := runSettingsSet(nil, []string{"site_name", "New Name"})
	require.NoError(t, err)
}

// --- Settings Secrets ---

func TestSettingsSecretsList_Success(t *testing.T) {
	resetSettingsSecretsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"key": "smtp_password", "description": "SMTP password", "created_at": "2024-01-01T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runSettingsSecretsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
}

func TestSettingsSecretsSet_Success(t *testing.T) {
	resetSettingsSecretsFlags()
	settingsSecretDescription = "SMTP password"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v1/admin/settings/custom/secret")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "smtp_password", body["key"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"key": "smtp_password",
		})
	})
	defer cleanup()

	err := runSettingsSecretsSet(nil, []string{"smtp_password", "s3cret123"})
	require.NoError(t, err)
}

func TestSettingsSecretsGet_Success(t *testing.T) {
	resetSettingsSecretsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/settings/custom/secret/smtp_password")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"key": "smtp_password", "description": "SMTP password",
		})
	})
	defer cleanup()

	err := runSettingsSecretsGet(nil, []string{"smtp_password"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "smtp_password", result["key"])
}

func TestSettingsSecretsDelete_Success(t *testing.T) {
	resetSettingsSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/settings/custom/secret/smtp_password")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runSettingsSecretsDelete(nil, []string{"smtp_password"})
	require.NoError(t, err)
}
