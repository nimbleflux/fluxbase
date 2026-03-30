package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretsList_Success(t *testing.T) {
	resetSecretsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/secrets")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "db-password", "scope": "system", "created_at": "2024-01-01T00:00:00Z"},
			{"name": "api-key", "scope": "functions", "created_at": "2024-01-02T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runSecretsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "db-password", result[0]["name"])
}

func TestSecretsList_Empty(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runSecretsList(nil, []string{})
	require.NoError(t, err)
}

func TestSecretsList_APIError(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runSecretsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestSecretsSet_Success(t *testing.T) {
	resetSecretsFlags()

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/api/v1/secrets")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "my-secret-value", body["value"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "db-password", "scope": "system",
		})
	})
	defer cleanup()

	err := runSecretsSet(nil, []string{"db-password", "my-secret-value"})
	require.NoError(t, err)
}

func TestSecretsSet_APIError(t *testing.T) {
	resetSecretsFlags()

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		// PUT fails with not found, POST also fails
		if r.Method == http.MethodPut {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusBadRequest, "invalid secret name")
	})
	defer cleanup()

	err := runSecretsSet(nil, []string{"BAD-NAME", "value"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid secret name")
}

func TestSecretsGet_Success(t *testing.T) {
	resetSecretsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/secrets/by-name/db-password")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "db-password", "scope": "system", "value": "s3cret",
		})
	})
	defer cleanup()

	err := runSecretsGet(nil, []string{"db-password"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "db-password", result["name"])
}

func TestSecretsGet_APIError(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "secret not found")
	})
	defer cleanup()

	err := runSecretsGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
}

func TestSecretsDelete_Success(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/secrets/by-name/db-password")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runSecretsDelete(nil, []string{"db-password"})
	require.NoError(t, err)
}

func TestSecretsDelete_APIError(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "secret not found")
	})
	defer cleanup()

	err := runSecretsDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
}

func TestSecretsHistory_Success(t *testing.T) {
	resetSecretsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/versions")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"version": float64(1), "created_at": "2024-01-01T00:00:00Z"},
			{"version": float64(2), "created_at": "2024-01-02T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runSecretsHistory(nil, []string{"db-password"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
}

func TestSecretsHistory_Empty(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runSecretsHistory(nil, []string{"db-password"})
	require.NoError(t, err)
}

func TestSecretsHistory_APIError(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "secret not found")
	})
	defer cleanup()

	err := runSecretsHistory(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "secret not found")
}

func TestSecretsRollback_Success(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/rollback/")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "db-password", "version": float64(1),
		})
	})
	defer cleanup()

	err := runSecretsRollback(nil, []string{"db-password", "1"})
	require.NoError(t, err)
}

func TestSecretsRollback_APIError(t *testing.T) {
	resetSecretsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "version not found")
	})
	defer cleanup()

	err := runSecretsRollback(nil, []string{"db-password", "99"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "version not found")
}
