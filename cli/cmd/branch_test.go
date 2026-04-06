package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBranchListCmd creates a minimal cobra.Command with the "mine" flag for tests.
func newBranchListCmd() *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().BoolP("mine", "m", false, "Show only branches created by me")
	return cmd
}

// --- Branch List ---

func TestBranchList_Success(t *testing.T) {
	resetBranchFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches")
		assert.Equal(t, "100", r.URL.Query().Get("limit"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"branches": []map[string]interface{}{
				{"id": "br-1", "name": "main", "slug": "main", "status": "ready", "type": "main", "database_name": "fluxbase"},
				{"id": "br-2", "name": "my-feature", "slug": "my-feature", "status": "ready", "type": "preview", "database_name": "branch_my_feature"},
			},
			"total": 2,
		})
	})
	defer cleanup()

	cmd := newBranchListCmd()
	err := runBranchList(cmd, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Branches) outputs a JSON array
	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Len(t, result, 2)
	assert.Equal(t, "main", result[0]["name"])
	assert.Equal(t, "my-feature", result[1]["name"])
}

func TestBranchList_Empty(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"branches": []map[string]interface{}{},
			"total":    0,
		})
	})
	defer cleanup()

	cmd := newBranchListCmd()
	err := runBranchList(cmd, []string{})
	require.NoError(t, err)
}

func TestBranchList_APIError(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	cmd := newBranchListCmd()
	err := runBranchList(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

// --- Branch Get ---

func TestBranchGet_Success(t *testing.T) {
	resetBranchFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches/my-feature")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":              "br-2",
			"name":            "my-feature",
			"slug":            "my-feature",
			"status":          "ready",
			"type":            "preview",
			"database_name":   "branch_my_feature",
			"data_clone_mode": "schema_only",
		})
	})
	defer cleanup()

	err := runBranchGet(nil, []string{"my-feature"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "my-feature", result["name"])
	assert.Equal(t, "ready", result["status"])
}

func TestBranchGet_NotFound(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "branch not found")
	})
	defer cleanup()

	err := runBranchGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found")
}

// --- Branch Create ---

func TestBranchCreate_Success(t *testing.T) {
	resetBranchFlags()
	// Set defaults that cobra normally provides via flag defaults
	branchDataCloneMode = "schema_only"
	branchCreateType = "preview"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "my-feature", body["name"])
		assert.Equal(t, "schema_only", body["data_clone_mode"])
		assert.Equal(t, "preview", body["type"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":              "br-3",
			"name":            "my-feature",
			"slug":            "my-feature",
			"status":          "creating",
			"type":            "preview",
			"database_name":   "branch_my_feature",
			"data_clone_mode": "schema_only",
		})
	})
	defer cleanup()

	err := runBranchCreate(nil, []string{"my-feature"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "my-feature", result["name"])
}

func TestBranchCreate_WithGitHub(t *testing.T) {
	resetBranchFlags()
	branchGitHubPR = 42
	branchGitHubRepo = "owner/repo"

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, float64(42), body["github_pr_number"])
		assert.Equal(t, "owner/repo", body["github_repo"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":     "br-4",
			"name":   "pr-42",
			"slug":   "pr-42",
			"status": "creating",
		})
	})
	defer cleanup()

	err := runBranchCreate(nil, []string{"pr-42"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
}

func TestBranchCreate_APIError(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusConflict, "branch already exists")
	})
	defer cleanup()

	err := runBranchCreate(nil, []string{"existing-branch"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch already exists")
}

// --- Branch Delete ---

func TestBranchDelete_Success(t *testing.T) {
	resetBranchFlags()
	branchForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches/my-feature")

		w.WriteHeader(http.StatusNoContent)
	})
	defer cleanup()

	err := runBranchDelete(nil, []string{"my-feature"})
	require.NoError(t, err)
}

func TestBranchDelete_APIError(t *testing.T) {
	resetBranchFlags()
	branchForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "branch not found")
	})
	defer cleanup()

	err := runBranchDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found")
}

// --- Branch Reset ---

func TestBranchReset_Success(t *testing.T) {
	resetBranchFlags()
	branchForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches/my-feature/reset")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":     "br-2",
			"name":   "my-feature",
			"slug":   "my-feature",
			"status": "creating",
		})
	})
	defer cleanup()

	// runBranchReset uses fmt.Printf for output, not the formatter
	err := runBranchReset(nil, []string{"my-feature"})
	require.NoError(t, err)
}

func TestBranchReset_APIError(t *testing.T) {
	resetBranchFlags()
	branchForce = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "branch not found")
	})
	defer cleanup()

	err := runBranchReset(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found")
}

// --- Branch Status ---

func TestBranchStatus_Success(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches/my-feature")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":     "br-2",
			"name":   "my-feature",
			"slug":   "my-feature",
			"status": "ready",
		})
	})
	defer cleanup()

	err := runBranchStatus(nil, []string{"my-feature"})
	require.NoError(t, err)
}

func TestBranchStatus_NotFound(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "branch not found")
	})
	defer cleanup()

	err := runBranchStatus(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "branch not found")
}

// --- Branch Stats ---

func TestBranchStats_Success(t *testing.T) {
	resetBranchFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/branches/stats/pools")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"pools": map[string]interface{}{
				"main": map[string]interface{}{
					"total_conns":    float64(10),
					"idle_conns":     float64(5),
					"acquired_conns": float64(5),
					"max_conns":      float64(20),
					"acquire_count":  float64(100),
				},
			},
		})
	})
	defer cleanup()

	err := runBranchStats(nil, []string{})
	require.NoError(t, err)

	// formatter.Print(result.Pools) outputs the pools map without a "pools" wrapper
	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Contains(t, result, "main")
}

func TestBranchStats_APIError(t *testing.T) {
	resetBranchFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "pool stats unavailable")
	})
	defer cleanup()

	err := runBranchStats(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pool stats unavailable")
}
