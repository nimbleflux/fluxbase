package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationsList_Success(t *testing.T) {
	resetMigrationsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/migrations")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "001_create_users", "status": "applied", "applied_at": "2024-01-01T00:00:00Z"},
		})
	})
	defer cleanup()

	err := runMigrationsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 1)
	assert.Equal(t, "001_create_users", result[0]["name"])
}

func TestMigrationsGet_Success(t *testing.T) {
	resetMigrationsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/migrations/001_create_users")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "001_create_users", "status": "applied",
		})
	})
	defer cleanup()

	err := runMigrationsGet(nil, []string{"001_create_users"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "001_create_users", result["name"])
}

func TestMigrationsCreate_Success(t *testing.T) {
	resetMigrationsFlags()
	migUpSQL = "CREATE TABLE test (id serial primary key);"
	migDownSQL = "DROP TABLE test;"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/migrations")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Contains(t, body["up_sql"], "CREATE TABLE")

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"name": "002_create_test", "status": "pending",
		})
	})
	defer cleanup()

	err := runMigrationsCreate(nil, []string{"002_create_test"})
	require.NoError(t, err)
}

func TestMigrationsApply_Success(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/apply")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "001_create_users", "status": "applied",
		})
	})
	defer cleanup()

	err := runMigrationsApply(nil, []string{"001_create_users"})
	require.NoError(t, err)
}

func TestMigrationsRollback_Success(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/rollback")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"name": "001_create_users", "status": "rolled_back",
		})
	})
	defer cleanup()

	err := runMigrationsRollback(nil, []string{"001_create_users"})
	require.NoError(t, err)
}

func TestMigrationsApplyPending_Success(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/migrations/apply-pending")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"applied": float64(3),
		})
	})
	defer cleanup()

	err := runMigrationsApplyPending(nil, []string{})
	require.NoError(t, err)
}

func TestMigrationsList_Empty(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runMigrationsList(nil, []string{})
	require.NoError(t, err)
}

func TestMigrationsList_APIError(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runMigrationsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestMigrationsGet_APIError(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "migration not found")
	})
	defer cleanup()

	err := runMigrationsGet(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migration not found")
}

func TestMigrationsCreate_APIError(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusConflict, "migration already exists")
	})
	defer cleanup()

	err := runMigrationsCreate(nil, []string{"duplicate"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "migration already exists")
}

func TestMigrationsApply_APIError(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "apply failed")
	})
	defer cleanup()

	err := runMigrationsApply(nil, []string{"broken"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "apply failed")
}

func TestMigrationsRollback_APIError(t *testing.T) {
	resetMigrationsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "rollback failed")
	})
	defer cleanup()

	err := runMigrationsRollback(nil, []string{"broken"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rollback failed")
}

func TestMigrationsSync_DryRun(t *testing.T) {
	resetMigrationsFlags()
	migSyncDir = t.TempDir()
	migDryRun = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("API should not be called in dry-run mode")
	})
	defer cleanup()

	err := runMigrationsSync(nil, []string{})
	_ = err
}
