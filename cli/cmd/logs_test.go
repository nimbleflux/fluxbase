package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogsList_Success(t *testing.T) {
	resetLogsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/logs")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": []map[string]interface{}{
				{
					"timestamp": "2024-06-15T10:00:00Z",
					"level":     "info",
					"category":  "system",
					"component": "auth",
					"message":   "User logged in",
				},
				{
					"timestamp": "2024-06-15T10:01:00Z",
					"level":     "error",
					"category":  "http",
					"component": "api",
					"message":   "Request failed",
				},
			},
			"total_count": float64(2),
			"has_more":    false,
		})
	})
	defer cleanup()

	err := runLogsList(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	entries, ok := result["entries"].([]interface{})
	require.True(t, ok)
	require.Len(t, entries, 2)
}

func TestLogsList_Empty(t *testing.T) {
	resetLogsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries":     []map[string]interface{}{},
			"total_count": float64(0),
			"has_more":    false,
		})
	})
	defer cleanup()

	err := runLogsList(nil, []string{})
	require.NoError(t, err)
}

func TestLogsList_WithFilters(t *testing.T) {
	resetLogsFlags()
	logsCategory = "system"
	logsLevel = "error"
	logsComponent = "auth"
	logsSearch = "database"
	logsLimit = 50
	logsSortAsc = true

	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "system", q.Get("category"))
		assert.Equal(t, "error", q.Get("level"))
		assert.Equal(t, "auth", q.Get("component"))
		assert.Equal(t, "database", q.Get("search"))
		assert.Equal(t, "50", q.Get("limit"))
		assert.Equal(t, "true", q.Get("sort_asc"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": []map[string]interface{}{
				{
					"timestamp": "2024-06-15T10:00:00Z",
					"level":     "error",
					"category":  "system",
					"component": "auth",
					"message":   "database connection failed",
				},
			},
			"total_count": float64(1),
			"has_more":    false,
		})
	})
	defer cleanup()

	err := runLogsList(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["entries"])
}

func TestLogsList_WithSinceFilter(t *testing.T) {
	resetLogsFlags()
	logsSince = "1h"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.NotEmpty(t, q.Get("start_time"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries":     []map[string]interface{}{},
			"total_count": float64(0),
			"has_more":    false,
		})
	})
	defer cleanup()

	err := runLogsList(nil, []string{})
	require.NoError(t, err)
}

func TestLogsList_APIError(t *testing.T) {
	resetLogsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "database error")
	})
	defer cleanup()

	err := runLogsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestLogsStats_Success(t *testing.T) {
	resetLogsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/logs/stats")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"total_entries": float64(1500),
			"entries_by_category": map[string]interface{}{
				"system":   float64(500),
				"http":     float64(800),
				"security": float64(200),
			},
			"entries_by_level": map[string]interface{}{
				"info":  float64(1000),
				"warn":  float64(300),
				"error": float64(200),
			},
			"oldest_entry": "2024-01-01T00:00:00Z",
			"newest_entry": "2024-12-31T23:59:59Z",
		})
	})
	defer cleanup()

	err := runLogsStats(nil, []string{})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.Equal(t, float64(1500), result["total_entries"])
}

func TestLogsStats_APIError(t *testing.T) {
	resetLogsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "internal error")
	})
	defer cleanup()

	err := runLogsStats(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internal error")
}

func TestLogsExecution_Success(t *testing.T) {
	resetLogsFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/admin/logs/executions/exec123")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": []map[string]interface{}{
				{
					"line_number": float64(1),
					"level":       "info",
					"message":     "Starting execution",
					"timestamp":   "2024-06-15T10:00:00Z",
				},
				{
					"line_number": float64(2),
					"level":       "info",
					"message":     "Execution completed",
					"timestamp":   "2024-06-15T10:00:01Z",
				},
			},
			"count": 2,
		})
	})
	defer cleanup()

	err := runLogsExecution(nil, []string{"exec123"})
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	assert.NotNil(t, result["entries"])
}

func TestLogsExecution_Empty(t *testing.T) {
	resetLogsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": []map[string]interface{}{},
			"count":   0,
		})
	})
	defer cleanup()

	err := runLogsExecution(nil, []string{"exec-nonexistent"})
	require.NoError(t, err)
}

func TestLogsExecution_WithTail(t *testing.T) {
	resetLogsFlags()
	logsTail = 10

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		assert.Equal(t, "-10", q.Get("after_line"))

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"entries": []map[string]interface{}{
				{"line_number": float64(1), "level": "info", "message": "test", "timestamp": "2024-06-15T10:00:00Z"},
			},
			"count": 1,
		})
	})
	defer cleanup()

	err := runLogsExecution(nil, []string{"exec123"})
	require.NoError(t, err)
}

func TestLogsExecution_APIError(t *testing.T) {
	resetLogsFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "execution not found")
	})
	defer cleanup()

	err := runLogsExecution(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "execution not found")
}
