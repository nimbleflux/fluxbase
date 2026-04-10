package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJobsList_Success(t *testing.T) {
	resetJobFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/jobs/functions")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "cleanup", "namespace": "default", "enabled": true, "timeout_seconds": float64(300), "schedule": "0 * * * *"},
			{"name": "process", "namespace": "default", "enabled": true, "timeout_seconds": float64(60), "schedule": ""},
		})
	})
	defer cleanup()

	err := runJobsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "cleanup", result[0]["name"])
}

func TestJobsList_Empty(t *testing.T) {
	resetJobFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runJobsList(nil, []string{})
	require.NoError(t, err)
}

func TestJobsSubmit_Success(t *testing.T) {
	resetJobFlags()
	jobPayload = `{"key":"value"}`

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/jobs/submit")

		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, "my-job", body["job_name"])

		payload, ok := body["payload"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "value", payload["key"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id":     "job-abc123",
			"status": "pending",
		})
	})
	defer cleanup()

	err := runJobsSubmit(nil, []string{"my-job"})
	require.NoError(t, err)
}

func TestJobsSubmit_InvalidPayload(t *testing.T) {
	resetJobFlags()
	jobPayload = `{invalid json}`

	err := runJobsSubmit(nil, []string{"my-job"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON payload")
}

func TestJobsSubmit_WithPriority(t *testing.T) {
	resetJobFlags()
	jobPayload = `{}`
	jobPriority = 10

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		readRequestBody(t, r, &body)
		assert.Equal(t, float64(10), body["priority"])

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"id": "job-xyz",
		})
	})
	defer cleanup()

	err := runJobsSubmit(nil, []string{"priority-job"})
	require.NoError(t, err)
}

func TestJobsStatus_Success(t *testing.T) {
	resetJobFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/jobs/job-abc123")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":     "job-abc123",
			"status": "completed",
			"result": map[string]interface{}{"processed": float64(42)},
		})
	})
	defer cleanup()

	err := runJobsStatus(nil, []string{"job-abc123"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, "completed", result["status"])
}

func TestJobsStatus_NotFound(t *testing.T) {
	resetJobFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "job not found")
	})
	defer cleanup()

	err := runJobsStatus(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "job not found")
}

func TestJobsCancel_Success(t *testing.T) {
	resetJobFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/jobs/job-abc123/cancel")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runJobsCancel(nil, []string{"job-abc123"})
	require.NoError(t, err)
}

func TestJobsRetry_Success(t *testing.T) {
	resetJobFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/jobs/job-abc123/retry")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"id":     "job-def456",
			"status": "pending",
		})
	})
	defer cleanup()

	err := runJobsRetry(nil, []string{"job-abc123"})
	require.NoError(t, err)
}

func TestJobsStats_Success(t *testing.T) {
	resetJobFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/jobs/stats")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"pending":   float64(5),
			"running":   float64(2),
			"completed": float64(100),
			"failed":    float64(3),
		})
	})
	defer cleanup()

	err := runJobsStats(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, float64(5), result["pending"])
	assert.Equal(t, float64(100), result["completed"])
}

func TestJobsSubmit_APIError(t *testing.T) {
	resetJobFlags()
	jobPayload = `{}`

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusBadRequest, "invalid job name")
	})
	defer cleanup()

	err := runJobsSubmit(nil, []string{"bad-job"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid job name")
}
