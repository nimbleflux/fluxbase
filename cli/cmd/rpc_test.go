package cmd

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

// TestRPCSync_UsesNameAnnotation is the regression test for issue #363:
// `fluxbase rpc sync` registered procedures under the filename stem and
// silently ignored the @fluxbase:name annotation. The sync payload must
// carry the annotated name.
func TestRPCSync_UsesNameAnnotation(t *testing.T) {
	resetRPCFlags()
	rpcSyncDir = t.TempDir()
	rpcNamespace = "wayli"

	require.NoError(t, os.WriteFile(filepath.Join(rpcSyncDir, "ensure-user-profile.sql"), []byte(`-- @fluxbase:name ensure_user_profile
-- @fluxbase:description Ensures a profile row exists
CREATE OR REPLACE FUNCTION ensure_user_profile() RETURNS void AS $$ BEGIN END; $$ LANGUAGE plpgsql;`), 0o644))
	// File without annotation keeps its filename-derived name
	require.NoError(t, os.WriteFile(filepath.Join(rpcSyncDir, "plain.sql"), []byte("SELECT 1;"), 0o644))

	var capturedBody map[string]interface{}
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/rpc/sync")
		readRequestBody(t, r, &capturedBody)
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"summary": map[string]interface{}{
				"created": 2, "updated": 0, "deleted": 0, "unchanged": 0, "errors": 0,
			},
		})
	})
	defer cleanup()

	require.NoError(t, runRPCSync(nil, []string{}))

	require.Equal(t, "wayli", capturedBody["namespace"])
	procs, ok := capturedBody["procedures"].([]interface{})
	require.True(t, ok)
	require.Len(t, procs, 2)

	names := map[string]bool{}
	for _, p := range procs {
		m := p.(map[string]interface{})
		names[m["name"].(string)] = true
	}
	assert.True(t, names["ensure_user_profile"], "annotated snake_case name should be used")
	assert.True(t, names["plain"], "filename stem should be used when no annotation")
	assert.False(t, names["ensure-user-profile"], "filename stem must not override the annotation")
}
