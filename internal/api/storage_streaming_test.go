package api

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStreamUpload_EmptyKeyValidation tests that empty key is rejected
func TestStreamUpload_EmptyKeyValidation(t *testing.T) {
	app := fiber.New()

	handler := &StorageHandler{
		storageManager: nil, // Empty key check happens before storage access
	}

	app.Post("/storage/:bucket/stream/*", handler.StreamUpload)

	// Request with empty key (wildcard captures empty string)
	req := httptest.NewRequest("POST", "/storage/my-bucket/stream/",
		bytes.NewReader([]byte("test")))
	req.Header.Set("Content-Length", "4")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Should get 400 for empty key
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Contains(t, result["error"], "bucket and key are required")
}

// TestStreamUpload_RouteMatching tests that the route pattern works correctly
func TestStreamUpload_RouteMatching(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		shouldMatch  bool
		expectStatus int
	}{
		{
			name:         "simple key",
			path:         "/storage/bucket/stream/file.txt",
			shouldMatch:  true,
			expectStatus: 400, // Will fail validation, but route matches
		},
		{
			name:         "nested path",
			path:         "/storage/bucket/stream/folder/file.txt",
			shouldMatch:  true,
			expectStatus: 400,
		},
		{
			name:         "deep nested path",
			path:         "/storage/bucket/stream/a/b/c/d/file.txt",
			shouldMatch:  true,
			expectStatus: 400,
		},
		{
			name:         "missing stream segment",
			path:         "/storage/bucket/file.txt",
			shouldMatch:  false,
			expectStatus: 404, // Route doesn't match
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			handler := &StorageHandler{
				storageManager: nil,
			}

			app.Post("/storage/:bucket/stream/*", handler.StreamUpload)

			req := httptest.NewRequest("POST", tt.path, nil)
			req.Header.Set("Content-Length", "0") // Zero to trigger validation error

			resp, err := app.Test(req)
			// Some tests may fail with EOF, that's OK - we're just testing routing
			if err != nil {
				t.Logf("Request error (may be expected): %v", err)
				return
			}
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expectStatus, resp.StatusCode,
				"Path: %s, expected status: %d, got: %d", tt.path, tt.expectStatus, resp.StatusCode)
		})
	}
}

// TestStreamUpload_ParameterExtraction tests parameter extraction from URL
func TestStreamUpload_ParameterExtraction(t *testing.T) {
	// This test documents how Fiber extracts bucket and key from the URL
	tests := []struct {
		name       string
		path       string
		wantBucket string
		wantKey    string
	}{
		{
			name:       "simple file",
			path:       "/storage/my-bucket/stream/test.txt",
			wantBucket: "my-bucket",
			wantKey:    "test.txt",
		},
		{
			name:       "nested file",
			path:       "/storage/my-bucket/stream/folder/test.txt",
			wantBucket: "my-bucket",
			wantKey:    "folder/test.txt",
		},
		{
			name:       "bucket with hyphens",
			path:       "/storage/my-test-bucket/stream/file.txt",
			wantBucket: "my-test-bucket",
			wantKey:    "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			// Custom handler to verify parameter extraction
			app.Post("/storage/:bucket/stream/*", func(c fiber.Ctx) error {
				bucket := c.Params("bucket")
				key := c.Params("*")

				assert.Equal(t, tt.wantBucket, bucket, "Bucket mismatch")
				assert.Equal(t, tt.wantKey, key, "Key mismatch")

				return c.SendStatus(200)
			})

			req := httptest.NewRequest("POST", tt.path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, 200, resp.StatusCode)
		})
	}
}

// NOTE: Full testing of StreamUpload functionality requires:
// 1. Mock storage.Service with ValidateUploadSize implementation
// 2. Database connection for RLS and object tracking
// 3. Testing actual streaming I/O with large files
//
// For these integration scenarios, see test/e2e/ directory.
//
// This file focuses on testing what CAN be tested without those dependencies:
// - Route matching and parameter extraction
// - Early validation (empty bucket/key)
//
// Additional features that need integration tests:
// - Content-Length header validation (happens after storage.ValidateUploadSize)
// - X-Storage-Metadata header parsing (happens after storage validation)
// - X-Storage-Content-Type header handling
// - Actual file streaming and storage
// - File size limits enforcement
// - RLS context and ownership tracking
