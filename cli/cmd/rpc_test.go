package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRPCList_Success(t *testing.T) {
	resetRPCFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/rpc/procedures")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"procedures": []map[string]interface{}{
				{"name": "calculate_total", "namespace": "default", "enabled": true, "is_public": false, "schedule": ""},
				{"name": "process_order", "namespace": "default", "enabled": true, "is_public": true, "schedule": "*/5 * * * *"},
			},
			"count": float64(2),
		})
	})
	defer cleanup()

	err := runRPCList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "calculate_total", result[0]["name"])
}

func TestRPCGet_Success(t *testing.T) {
	resetRPCFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/rpc/procedures/")
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "calculate_total", "namespace": "default", "type": "function",
		})
	})
	defer cleanup()

	err := runRPCGet(nil, []string{"default/calculate_total"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "calculate_total", result["name"])
}

func TestRPCInvoke_Success(t *testing.T) {
	resetRPCFlags()
	rpcParams = `{"x": 1, "y": 2}`

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/rpc/")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		params, ok := body["params"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(1), params["x"])

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"result": float64(3),
		})
	})
	defer cleanup()

	err := runRPCInvoke(nil, []string{"default/calculate_total"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, float64(3), result["result"])
}

func TestRPCSync_DryRun(t *testing.T) {
	resetRPCFlags()
	rpcSyncDir = t.TempDir()
	rpcDryRun = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called in dry-run mode")
	})
	defer cleanup()

	err := runRPCSync(nil, []string{})
	_ = err
}
