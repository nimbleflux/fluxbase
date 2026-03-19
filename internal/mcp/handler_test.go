package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// rateLimiter Tests
// =============================================================================

func TestNewRateLimiter(t *testing.T) {
	t.Run("creates rate limiter with limit", func(t *testing.T) {
		rl := newRateLimiter(100)

		require.NotNil(t, rl)
		assert.Equal(t, 100, rl.limit)
		assert.NotNil(t, rl.requests)
	})

	t.Run("creates rate limiter with zero limit", func(t *testing.T) {
		rl := newRateLimiter(0)

		require.NotNil(t, rl)
		assert.Equal(t, 0, rl.limit)
	})
}

func TestRateLimiter_Allow(t *testing.T) {
	t.Run("allows requests when under limit", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    10,
		}

		for i := 0; i < 10; i++ {
			allowed := rl.allow("client-1")
			assert.True(t, allowed, "Request %d should be allowed", i+1)
		}
	})

	t.Run("blocks requests when at limit", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    5,
		}

		// Fill up the limit
		for i := 0; i < 5; i++ {
			rl.allow("client-1")
		}

		// Next request should be blocked
		allowed := rl.allow("client-1")
		assert.False(t, allowed)
	})

	t.Run("allows all requests when limit is zero", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    0, // Disabled
		}

		for i := 0; i < 100; i++ {
			allowed := rl.allow("client-1")
			assert.True(t, allowed)
		}
	})

	t.Run("allows all requests when limit is negative", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    -1, // Disabled
		}

		allowed := rl.allow("client-1")
		assert.True(t, allowed)
	})

	t.Run("different clients have separate limits", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2,
		}

		// Client 1 uses up their limit
		rl.allow("client-1")
		rl.allow("client-1")
		assert.False(t, rl.allow("client-1"))

		// Client 2 should still be allowed
		assert.True(t, rl.allow("client-2"))
		assert.True(t, rl.allow("client-2"))
		assert.False(t, rl.allow("client-2"))
	})

	t.Run("old requests outside window are not counted", func(t *testing.T) {
		rl := &rateLimiter{
			requests: make(map[string][]time.Time),
			limit:    2,
		}

		// Add old requests (more than 1 minute ago)
		oldTime := time.Now().Add(-2 * time.Minute)
		rl.requests["client-1"] = []time.Time{oldTime, oldTime}

		// Should allow new requests because old ones are expired
		allowed := rl.allow("client-1")
		assert.True(t, allowed)
	})
}

// =============================================================================
// Handler Struct Tests
// =============================================================================

func TestHandler_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled:         true,
			RateLimitPerMin: 60,
		}

		handler := &Handler{
			server:      nil,
			config:      cfg,
			db:          nil,
			rateLimiter: nil,
		}

		assert.Nil(t, handler.server)
		assert.Equal(t, cfg, handler.config)
		assert.Nil(t, handler.db)
		assert.Nil(t, handler.rateLimiter)
	})
}

// =============================================================================
// NewHandler Tests
// =============================================================================

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with config", func(t *testing.T) {
		cfg := &config.MCPConfig{
			Enabled:         true,
			RateLimitPerMin: 100,
		}

		handler := NewHandler(cfg, nil)

		require.NotNil(t, handler)
		assert.NotNil(t, handler.server)
		assert.Equal(t, cfg, handler.config)
		assert.Nil(t, handler.db)
		assert.NotNil(t, handler.rateLimiter)
	})

	t.Run("creates handler with rate limiter based on config", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 50,
		}

		handler := NewHandler(cfg, nil)

		assert.NotNil(t, handler.rateLimiter)
		assert.Equal(t, 50, handler.rateLimiter.limit)
	})
}

// =============================================================================
// Server() Method Tests
// =============================================================================

func TestHandler_Server(t *testing.T) {
	t.Run("returns underlying server", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		handler := NewHandler(cfg, nil)

		server := handler.Server()

		assert.NotNil(t, server)
		assert.Equal(t, handler.server, server)
	})
}

// =============================================================================
// handleHealth Tests
// =============================================================================

func TestHandler_handleHealth(t *testing.T) {
	t.Run("returns healthy status", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Get("/health", handler.HandleHealth)

		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

// =============================================================================
// handlePost Tests
// =============================================================================

func TestHandler_handlePost(t *testing.T) {
	t.Run("rejects non-JSON content type", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 0, // Disable rate limiting for tests
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "text/plain")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("accepts application/json content type", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 0,
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should not return UnsupportedMediaType
		assert.NotEqual(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("accepts application/json with charset", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 0,
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	})

	t.Run("rejects request exceeding max size", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 0,
			MaxMessageSize:  10, // Very small limit
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		// Body larger than MaxMessageSize
		largeBody := strings.Repeat("a", 100)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(largeBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("allows request within max size", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 0,
			MaxMessageSize:  1000,
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("enforces rate limiting", func(t *testing.T) {
		cfg := &config.MCPConfig{
			RateLimitPerMin: 2, // Very low limit for testing
		}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Post("/", handler.HandlePost)

		// Make requests up to the limit
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")
			resp, err := app.Test(req)
			require.NoError(t, err)
			_ = resp.Body.Close()
			assert.NotEqual(t, http.StatusTooManyRequests, resp.StatusCode, "Request %d should be allowed", i+1)
		}

		// Next request should be rate limited
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})
}

// =============================================================================
// handleGet Tests
// =============================================================================

func TestHandler_handleGet(t *testing.T) {
	t.Run("rejects without text/event-stream Accept header", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Get("/", handler.HandleGet)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotAcceptable, resp.StatusCode)
	})

	t.Run("returns not implemented for SSE stream", func(t *testing.T) {
		cfg := &config.MCPConfig{}
		handler := NewHandler(cfg, nil)

		app := fiber.New()
		app.Get("/", handler.HandleGet)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept", "text/event-stream")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
	})
}

// =============================================================================
// Rate Limit Key Detection Tests
// =============================================================================

func TestRateLimitKeyDetection(t *testing.T) {
	t.Run("uses client key ID when available", func(t *testing.T) {
		authCtx := &AuthContext{
			ClientKeyID: "key-123",
		}

		rateLimitKey := authCtx.ClientKeyID
		if rateLimitKey == "" && authCtx.UserID != nil && *authCtx.UserID != "" {
			rateLimitKey = *authCtx.UserID
		}

		assert.Equal(t, "key-123", rateLimitKey)
	})

	t.Run("uses user ID when client key is empty", func(t *testing.T) {
		userID := "user-456"
		authCtx := &AuthContext{
			ClientKeyID: "",
			UserID:      &userID,
		}

		rateLimitKey := authCtx.ClientKeyID
		if rateLimitKey == "" && authCtx.UserID != nil && *authCtx.UserID != "" {
			rateLimitKey = *authCtx.UserID
		}

		assert.Equal(t, "user-456", rateLimitKey)
	})

	t.Run("empty when both are empty", func(t *testing.T) {
		emptyUserID := ""
		authCtx := &AuthContext{
			ClientKeyID: "",
			UserID:      &emptyUserID,
		}

		rateLimitKey := authCtx.ClientKeyID
		if rateLimitKey == "" && authCtx.UserID != nil && *authCtx.UserID != "" {
			rateLimitKey = *authCtx.UserID
		}

		assert.Empty(t, rateLimitKey)
	})
}

// =============================================================================
// Content-Type Validation Tests
// =============================================================================

func TestContentTypeValidation(t *testing.T) {
	t.Run("application/json is valid", func(t *testing.T) {
		contentType := "application/json"
		isValid := strings.HasPrefix(contentType, "application/json")

		assert.True(t, isValid)
	})

	t.Run("application/json with charset is valid", func(t *testing.T) {
		contentType := "application/json; charset=utf-8"
		isValid := strings.HasPrefix(contentType, "application/json")

		assert.True(t, isValid)
	})

	t.Run("text/plain is invalid", func(t *testing.T) {
		contentType := "text/plain"
		isValid := strings.HasPrefix(contentType, "application/json")

		assert.False(t, isValid)
	})

	t.Run("text/event-stream is invalid", func(t *testing.T) {
		contentType := "text/event-stream"
		isValid := strings.HasPrefix(contentType, "application/json")

		assert.False(t, isValid)
	})
}

// =============================================================================
// Accept Header Validation Tests
// =============================================================================

func TestAcceptHeaderValidation(t *testing.T) {
	t.Run("text/event-stream is valid for SSE", func(t *testing.T) {
		accept := "text/event-stream"
		isSSE := strings.Contains(accept, "text/event-stream")

		assert.True(t, isSSE)
	})

	t.Run("application/json is not valid for SSE", func(t *testing.T) {
		accept := "application/json"
		isSSE := strings.Contains(accept, "text/event-stream")

		assert.False(t, isSSE)
	})

	t.Run("multiple accepts including event-stream is valid", func(t *testing.T) {
		accept := "application/json, text/event-stream"
		isSSE := strings.Contains(accept, "text/event-stream")

		assert.True(t, isSSE)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkRateLimiterAllow(b *testing.B) {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    1000,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.allow("benchmark-client")
	}
}

func BenchmarkContentTypeCheck(b *testing.B) {
	contentType := "application/json; charset=utf-8"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.HasPrefix(contentType, "application/json")
	}
}

func BenchmarkAcceptHeaderCheck(b *testing.B) {
	accept := "application/json, text/event-stream, */*"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Contains(accept, "text/event-stream")
	}
}
