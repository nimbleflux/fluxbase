package cmd

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBucketsList_Success(t *testing.T) {
	resetStorageFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/buckets")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"buckets": []map[string]interface{}{
				{"name": "uploads", "public": true, "created_at": "2024-01-01T00:00:00Z"},
				{"name": "private", "public": false, "created_at": "2024-01-02T00:00:00Z"},
			},
		})
	})
	defer cleanup()

	err := runBucketsList(nil, []string{})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
	assert.Equal(t, "uploads", result[0]["name"])
}

func TestBucketsList_Empty(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"buckets": []map[string]interface{}{},
		})
	})
	defer cleanup()

	err := runBucketsList(nil, []string{})
	require.NoError(t, err)
}

func TestBucketsList_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusInternalServerError, "storage error")
	})
	defer cleanup()

	err := runBucketsList(nil, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "storage error")
}

func TestBucketsCreate_Success(t *testing.T) {
	resetStorageFlags()
	bucketPublic = true

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/buckets/test-bucket")

		respondJSON(w, http.StatusCreated, map[string]interface{}{
			"name": "test-bucket", "public": true,
		})
	})
	defer cleanup()

	err := runBucketsCreate(nil, []string{"test-bucket"})
	require.NoError(t, err)
}

func TestBucketsCreate_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusConflict, "bucket already exists")
	})
	defer cleanup()

	err := runBucketsCreate(nil, []string{"test-bucket"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bucket already exists")
}

func TestBucketsDelete_Success(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/buckets/test-bucket")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runBucketsDelete(nil, []string{"test-bucket"})
	require.NoError(t, err)
}

func TestBucketsDelete_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "bucket not found")
	})
	defer cleanup()

	err := runBucketsDelete(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bucket not found")
}

func TestObjectsList_Success(t *testing.T) {
	resetStorageFlags()
	_, buf, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/test-bucket")

		respondJSON(w, http.StatusOK, []map[string]interface{}{
			{"name": "file1.txt", "size": float64(1024)},
			{"name": "file2.txt", "size": float64(2048)},
		})
	})
	defer cleanup()

	err := runObjectsList(nil, []string{"test-bucket"})
	require.NoError(t, err)

	var result []map[string]interface{}
	require.NoError(t, json.Unmarshal(buf.Bytes(), &result))
	require.Len(t, result, 2)
}

func TestObjectsList_Empty(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondJSON(w, http.StatusOK, []map[string]interface{}{})
	})
	defer cleanup()

	err := runObjectsList(nil, []string{"test-bucket"})
	require.NoError(t, err)
}

func TestObjectsList_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "bucket not found")
	})
	defer cleanup()

	err := runObjectsList(nil, []string{"nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bucket not found")
}

func TestObjectsUpload_FileNotFound(t *testing.T) {
	resetStorageFlags()
	err := runObjectsUpload(nil, []string{"test-bucket", "test.txt", "/nonexistent/file.txt"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestObjectsDownload_Success(t *testing.T) {
	resetStorageFlags()
	tmpFile := t.TempDir() + "/downloaded.txt"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("downloaded content"))
	})
	defer cleanup()

	err := runObjectsDownload(nil, []string{"test-bucket", "test.txt", tmpFile})
	require.NoError(t, err)
}

func TestObjectsDownload_APIError(t *testing.T) {
	resetStorageFlags()
	tmpFile := t.TempDir() + "/downloaded.txt"

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "object not found")
	})
	defer cleanup()

	err := runObjectsDownload(nil, []string{"test-bucket", "nonexistent.txt", tmpFile})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download failed")
}

func TestObjectsDelete_Success(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/storage/")

		respondJSON(w, http.StatusOK, map[string]interface{}{})
	})
	defer cleanup()

	err := runObjectsDelete(nil, []string{"test-bucket", "test.txt"})
	require.NoError(t, err)
}

func TestObjectsDelete_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "object not found")
	})
	defer cleanup()

	err := runObjectsDelete(nil, []string{"test-bucket", "nonexistent.txt"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "object not found")
}

// TestObjectsURL_Success verifies the signed URL API call succeeds.
// Note: runObjectsURL uses fmt.Println for output, not formatter.Writer.
func TestObjectsURL_Success(t *testing.T) {
	resetStorageFlags()
	urlExpires = 3600

	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Contains(t, r.URL.Path, "/sign/")

		respondJSON(w, http.StatusOK, map[string]interface{}{
			"url": "https://storage.example.com/signed?token=abc",
		})
	})
	defer cleanup()

	err := runObjectsURL(nil, []string{"test-bucket", "test.txt"})
	require.NoError(t, err)
	// runObjectsURL uses fmt.Println, not formatter.Writer, so output goes to stdout
}

func TestObjectsURL_APIError(t *testing.T) {
	resetStorageFlags()
	_, _, cleanup := setupTestEnvWithHandler(t, func(w http.ResponseWriter, r *http.Request) {
		respondError(w, http.StatusNotFound, "bucket not found")
	})
	defer cleanup()

	err := runObjectsURL(nil, []string{"nonexistent", "test.txt"})
	require.Error(t, err)
}
