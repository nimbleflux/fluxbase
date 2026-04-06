package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunctionsList_Success(t *testing.T) {
	resetFunctionFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/functions/")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "hello", "namespace": "default", "enabled": true, "timeout_seconds": float64(30), "memory_limit_mb": float64(128)},
			{"name": "process", "namespace": "production", "enabled": true, "timeout_seconds": float64(60), "memory_limit_mb": float64(256)},
		})
	})
	defer cleanup()

	err := runFunctionsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "hello", result[0]["name"])
	assert.Equal(t, "process", result[1]["name"])
}

func TestFunctionsList_Empty(t *testing.T) {
	resetFunctionFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runFunctionsList(nil, []string{})
	require.NoError(t, err)
}

func TestFunctionsList_WithNamespace(t *testing.T) {
	resetFunctionFlags()
	fnNamespace = "production"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "production", r.URL.Query().Get("namespace"))
		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "process", "namespace": "production", "enabled": true, "timeout_seconds": float64(60), "memory_limit_mb": float64(256)},
		})
	})
	defer cleanup()

	err := runFunctionsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
}

func TestFunctionsGet_Success(t *testing.T) {
	resetFunctionFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/functions/hello")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name":            "hello",
			"namespace":       "default",
			"enabled":         true,
			"timeout_seconds": float64(30),
			"memory_limit_mb": float64(128),
			"code":            "export default function() { return 'hello'; }",
		})
	})
	defer cleanup()

	err := runFunctionsGet(nil, []string{"hello"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "hello", result["name"])
}

func TestFunctionsGet_NotFound(t *testing.T) {
	resetFunctionFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "function not found")
	})
	defer cleanup()

	err := runFunctionsGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "function not found")
}

func TestFunctionsDelete_Success(t *testing.T) {
	resetFunctionFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/functions/my-func")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runFunctionsDelete(nil, []string{"my-func"})
	require.NoError(t, err)
}

func TestFunctionsDelete_NotFound(t *testing.T) {
	resetFunctionFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "function not found")
	})
	defer cleanup()

	err := runFunctionsDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
}

func TestFunctionsUpdate_NoUpdates(t *testing.T) {
	resetFunctionFlags()
	// All update flags at default/zero values
	err := runFunctionsUpdate(nil, []string{"my-func"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no updates specified")
}

func TestFunctionsUpdate_Success(t *testing.T) {
	resetFunctionFlags()
	fnDescription = "updated description"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/functions/my-func")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "updated description", body["description"])

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	})
	defer cleanup()

	err := runFunctionsUpdate(nil, []string{"my-func"})
	require.NoError(t, err)
}

func TestFunctionsCreate_NoCodeFile(t *testing.T) {
	resetFunctionFlags()
	fnCodeFile = ""

	err := runFunctionsCreate(nil, []string{"my-func"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read code file")
}

func TestFunctionsList_APIError(t *testing.T) {
	resetFunctionFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "server error")
	})
	defer cleanup()

	err := runFunctionsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server error")
}
