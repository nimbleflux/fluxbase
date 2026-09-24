package api

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/storage"
)

// newStorageHandlerForTest builds a StorageHandler backed by a local storage
// manager without a database, so request-validation paths that run before any
// DB access can be exercised in unit tests.
func newStorageHandlerForTest(t *testing.T) *StorageHandler {
	t.Helper()

	cfg := &config.StorageConfig{
		Provider:      "local",
		LocalPath:     t.TempDir(),
		MaxUploadSize: 10 * 1024 * 1024,
	}
	mgr, err := storage.NewManager(cfg, "http://localhost:8080", "test-jwt-secret", nil)
	require.NoError(t, err)

	return NewStorageHandlerWithCache(mgr, nil, &config.Config{Storage: *cfg}, nil, nil)
}

// F1: expires_in above the signed-URL cap is rejected before any signing.
func TestGenerateSignedURL_ExpiryCap(t *testing.T) {
	h := newStorageHandlerForTest(t)
	app := newTestApp(t)
	app.Post("/api/v1/storage/:bucket/sign/*", h.GenerateSignedURL)

	cases := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"over 24h cap", `{"expires_in": 90001}`, fiber.StatusBadRequest},
		{"exactly 24h is allowed (rejected later by probe, not validation)", `{"expires_in": 86400}`, fiber.StatusInternalServerError},
		{"negative", `{"expires_in": -5}`, fiber.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, "/api/v1/storage/bkt/some/key/sign", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			assert.Equal(t, tc.wantCode, resp.StatusCode)
		})
	}
}

// F1: unsupported methods are rejected before any signing.
func TestGenerateSignedURL_MethodWhitelist(t *testing.T) {
	h := newStorageHandlerForTest(t)
	app := newTestApp(t)
	app.Post("/api/v1/storage/:bucket/sign/*", h.GenerateSignedURL)

	req := httptest.NewRequest(fiber.MethodPost, "/api/v1/storage/bkt/some/key/sign", bytes.NewBufferString(`{"method": "POST"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// F9: the copy/move routes must be reachable instead of being swallowed by
// the /:bucket/* upload wildcard. A JSON body to the copy route must produce
// a validation error from the copy handler (400 for a missing from_path), not
// the upload handler's response.
func TestStorageRoutes_CopyMoveNotShadowedByWildcard(t *testing.T) {
	h := newStorageHandlerForTest(t)

	app := newTestApp(t)
	// Register in the same order as routes/storage.go: copy/move BEFORE the
	// upload wildcard.
	app.Post("/api/v1/storage/:bucket/copy", h.CopyObjectHandler)
	app.Post("/api/v1/storage/:bucket/move", h.MoveObjectHandler)
	app.Post("/api/v1/storage/:bucket/*", h.UploadFile)

	cases := []struct {
		name string
		path string
		body string
	}{
		{"copy without body", "/api/v1/storage/bkt/copy", ``},
		{"copy missing from_path", "/api/v1/storage/bkt/copy", `{"to_path": "b.txt"}`},
		{"move without body", "/api/v1/storage/bkt/move", ``},
		{"move missing to_path", "/api/v1/storage/bkt/move", `{"from_path": "a.txt"}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(fiber.MethodPost, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()
			// The copy/move handlers validate the JSON body BEFORE touching
			// storage or the DB; the wildcard upload handler would instead
			// fail on form parsing / service lookup.
			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
		})
	}
}

// F8: the signed-URL IP rate limiter must drop expired windows on sweep.
func TestIPRateLimiter_SweepRemovesExpiredEntries(t *testing.T) {
	limiter := &ipRateLimiter{
		requests: map[string]*rateLimitEntry{
			"10.0.0.1": {count: 3, windowEnd: time.Now().Add(-time.Minute)}, // expired
			"10.0.0.2": {count: 1, windowEnd: time.Now().Add(time.Minute)},  // active
		},
		limit:  10,
		window: time.Minute,
	}

	limiter.sweep()

	assert.NotContains(t, limiter.requests, "10.0.0.1")
	assert.Contains(t, limiter.requests, "10.0.0.2")
}

// F8: transform rate limiters idle beyond the TTL are evicted by the sweep;
// active ones are kept.
func TestTransformLimiter_SweepEvictsIdleEntries(t *testing.T) {
	h := NewStorageHandlerWithCache(nil, nil, nil, nil, nil)
	h.transformLimiters = map[string]*limiterEntry{
		"1.1.1.1:user-a": {limiter: rate.NewLimiter(1, 1), lastSeen: time.Now().Add(-2 * time.Hour)},
		"2.2.2.2:user-b": {limiter: rate.NewLimiter(1, 1), lastSeen: time.Now()},
	}

	h.sweepRateLimiters()

	assert.NotContains(t, h.transformLimiters, "1.1.1.1:user-a")
	assert.Contains(t, h.transformLimiters, "2.2.2.2:user-b")
}

// F8: using a limiter refreshes its lastSeen so active users are not evicted.
func TestTransformLimiter_LastSeenRefreshed(t *testing.T) {
	h := NewStorageHandlerWithCache(nil, nil, nil, nil, nil)
	h.transformRateLimit = rate.Limit(1000)
	h.transformBurst = 1

	lim := h.getTransformLimiter("3.3.3.3:user-c")
	assert.NotNil(t, lim)
	entry := h.transformLimiters["3.3.3.3:user-c"]
	require.NotNil(t, entry)
	assert.WithinDuration(t, time.Now(), entry.lastSeen, time.Minute)
}

// F3: per-chunk caps — non-final chunks are capped at ChunkSize, the final
// chunk at the declared remainder.
func TestMaxChunkSizeFor(t *testing.T) {
	session := &storage.ChunkedUploadSession{
		TotalSize:   10 * 1024 * 1024, // 10MiB
		ChunkSize:   5 * 1024 * 1024,  // 5MiB
		TotalChunks: 2,
	}

	assert.Equal(t, int64(5*1024*1024), maxChunkSizeFor(session, 0))
	assert.Equal(t, int64(5*1024*1024), maxChunkSizeFor(session, 1), "final chunk capped at remainder")
	assert.Equal(t, int64(0), maxChunkSizeFor(session, 2), "out of range")
	assert.Equal(t, int64(0), maxChunkSizeFor(nil, 0))

	// Remainder smaller than ChunkSize
	session = &storage.ChunkedUploadSession{
		TotalSize:   12 * 1024 * 1024,
		ChunkSize:   5 * 1024 * 1024,
		TotalChunks: 3,
	}
	assert.Equal(t, int64(5*1024*1024), maxChunkSizeFor(session, 0))
	assert.Equal(t, int64(5*1024*1024), maxChunkSizeFor(session, 1))
	assert.Equal(t, int64(2*1024*1024), maxChunkSizeFor(session, 2))
}
