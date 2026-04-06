package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientKeysList_Success(t *testing.T) {
	resetClientKeyFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/client-keys")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"id": "ck-1", "name": "my-key", "scopes": []string{"read"}, "created_at": "2024-01-01T00:00:00Z"},
			{"id": "ck-2", "name": "other-key", "scopes": []string{"write"}, "created_at": "2024-01-02T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runClientKeysList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "my-key", result[0]["name"])
}

func TestClientKeysList_APIError(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runClientKeysList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestClientKeysList_Empty(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runClientKeysList(nil, []string{})
	require.NoError(t, err)
}

func TestClientKeysCreate_Success(t *testing.T) {
	resetClientKeyFlags()
	ckName = "new-key"
	ckScopes = "read,write"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/client-keys")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "new-key", body["name"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id": "ck-new", "name": "new-key", "key": "fb_ck_abc123",
		})
	})
	defer cleanup()

	err := runClientKeysCreate(nil, []string{})
	require.NoError(t, err)
	// runClientKeysCreate uses fmt.Printf, not formatter.Writer, so output goes to stdout
}

func TestClientKeysCreate_APIError(t *testing.T) {
	resetClientKeyFlags()
	ckName = "new-key"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusBadRequest, "invalid scopes")
	})
	defer cleanup()

	err := runClientKeysCreate(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid scopes")
}

func TestClientKeysGet_Success(t *testing.T) {
	resetClientKeyFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/client-keys/ck-1")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id": "ck-1", "name": "my-key", "scopes": []string{"read"},
		})
	})
	defer cleanup()

	err := runClientKeysGet(nil, []string{"ck-1"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "ck-1", result["id"])
}

func TestClientKeysGet_NotFound(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "client key not found")
	})
	defer cleanup()

	err := runClientKeysGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client key not found")
}

func TestClientKeysRevoke_APIError(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "client key not found")
	})
	defer cleanup()

	err := runClientKeysRevoke(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client key not found")
}

func TestClientKeysRevoke_Success(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/revoke")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runClientKeysRevoke(nil, []string{"ck-1"})
	require.NoError(t, err)
}

func TestClientKeysDelete_APIError(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "client key not found")
	})
	defer cleanup()

	err := runClientKeysDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "client key not found")
}

func TestClientKeysDelete_Success(t *testing.T) {
	resetClientKeyFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/client-keys/ck-1")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runClientKeysDelete(nil, []string{"ck-1"})
	require.NoError(t, err)
}
