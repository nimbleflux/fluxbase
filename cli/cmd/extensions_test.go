package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensionsList_Success(t *testing.T) {
	resetExtensionsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/extensions")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "pgcrypto", "installed": true, "default_version": "1.3"},
			{"name": "pgvector", "installed": false, "default_version": "0.5"},
		})
	})
	defer cleanup()

	err := runExtensionsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "pgcrypto", result[0]["name"])
}

func TestExtensionsList_Empty(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runExtensionsList(nil, []string{})
	require.NoError(t, err)
}

func TestExtensionsList_APIError(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runExtensionsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestExtensionsStatus_Success(t *testing.T) {
	resetExtensionsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/status")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "pgcrypto", "installed": true, "version": "1.3",
		})
	})
	defer cleanup()

	err := runExtensionsStatus(nil, []string{"pgcrypto"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "pgcrypto", result["name"])
}

func TestExtensionsStatus_APIError(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "extension not found")
	})
	defer cleanup()

	err := runExtensionsStatus(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "extension not found")
}

func TestExtensionsEnable_Success(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/enable")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "pgvector", "installed": true,
		})
	})
	defer cleanup()

	err := runExtensionsEnable(nil, []string{"pgvector"})
	require.NoError(t, err)
}

func TestExtensionsEnable_APIError(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "enable failed")
	})
	defer cleanup()

	err := runExtensionsEnable(nil, []string{"broken"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "enable failed")
}

func TestExtensionsDisable_Success(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/disable")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "pgvector", "installed": false,
		})
	})
	defer cleanup()

	err := runExtensionsDisable(nil, []string{"pgvector"})
	require.NoError(t, err)
}

func TestExtensionsDisable_APIError(t *testing.T) {
	resetExtensionsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "disable failed")
	})
	defer cleanup()

	err := runExtensionsDisable(nil, []string{"broken"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disable failed")
}
