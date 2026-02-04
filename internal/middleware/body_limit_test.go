package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatternBodyLimiter_GetLimit(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 1024,
		Patterns: []BodyLimitPattern{
			{Pattern: "/api/v1/storage/**", Limit: 100 * 1024 * 1024, Description: "storage"},
			{Pattern: "/api/v1/auth/**", Limit: 64 * 1024, Description: "auth"},
			{Pattern: "/api/v1/rest/*", Limit: 1024 * 1024, Description: "REST"},
			{Pattern: "/api/v1/rest/*/bulk", Limit: 10 * 1024 * 1024, Description: "bulk"},
		},
	}

	limiter := NewPatternBodyLimiter(config)

	tests := []struct {
		name      string
		path      string
		wantLimit int64
		wantDesc  string
	}{
		{
			name:      "storage upload",
			path:      "/api/v1/storage/bucket/file.txt",
			wantLimit: 100 * 1024 * 1024,
			wantDesc:  "storage",
		},
		{
			name:      "storage nested path",
			path:      "/api/v1/storage/bucket/folder/subfolder/file.txt",
			wantLimit: 100 * 1024 * 1024,
			wantDesc:  "storage",
		},
		{
			name:      "auth endpoint",
			path:      "/api/v1/auth/login",
			wantLimit: 64 * 1024,
			wantDesc:  "auth",
		},
		{
			name:      "auth nested",
			path:      "/api/v1/auth/2fa/verify",
			wantLimit: 64 * 1024,
			wantDesc:  "auth",
		},
		{
			name:      "REST endpoint single segment",
			path:      "/api/v1/rest/users",
			wantLimit: 1024 * 1024,
			wantDesc:  "REST",
		},
		{
			name:      "bulk operation - more specific match",
			path:      "/api/v1/rest/users/bulk",
			wantLimit: 10 * 1024 * 1024,
			wantDesc:  "bulk",
		},
		{
			name:      "unmatched path uses default",
			path:      "/health",
			wantLimit: 1024,
			wantDesc:  "default",
		},
		{
			name:      "unmatched nested path",
			path:      "/other/endpoint/deep/path",
			wantLimit: 1024,
			wantDesc:  "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			limit, desc := limiter.GetLimit(tt.path)
			assert.Equal(t, tt.wantLimit, limit, "limit mismatch for path %s", tt.path)
			assert.Equal(t, tt.wantDesc, desc, "description mismatch for path %s", tt.path)
		})
	}
}

func TestPatternBodyLimiter_Middleware_AcceptsUnderLimit(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 1024, // 1KB default
		Patterns: []BodyLimitPattern{
			{Pattern: "/api/**", Limit: 1024, Description: "API"},
		},
	}

	app := fiber.New()
	limiter := NewPatternBodyLimiter(config)
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Request with body under limit
	body := bytes.Repeat([]byte("a"), 500) // 500 bytes
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPatternBodyLimiter_Middleware_RejectsOverLimit(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 1024, // 1KB default
		Patterns: []BodyLimitPattern{
			{Pattern: "/api/**", Limit: 1024, Description: "API"},
		},
	}

	app := fiber.New()
	limiter := NewPatternBodyLimiter(config)
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Request with body over limit
	body := bytes.Repeat([]byte("a"), 2048) // 2KB
	req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestPatternBodyLimiter_Middleware_SkipsGET(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 100, // Very small limit
		Patterns:     []BodyLimitPattern{},
	}

	app := fiber.New()
	limiter := NewPatternBodyLimiter(config)
	app.Use(limiter.Middleware())
	app.Get("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPatternBodyLimiter_DifferentEndpointsDifferentLimits(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 1024,
		Patterns: []BodyLimitPattern{
			{Pattern: "/api/v1/storage/**", Limit: 10 * 1024, Description: "storage"},
			{Pattern: "/api/v1/auth/**", Limit: 512, Description: "auth"},
		},
	}

	app := fiber.New()
	limiter := NewPatternBodyLimiter(config)
	app.Use(limiter.Middleware())
	app.Post("/api/v1/storage/upload", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Post("/api/v1/auth/login", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Storage endpoint should accept 5KB
	storageBody := bytes.Repeat([]byte("a"), 5*1024)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/storage/upload", bytes.NewReader(storageBody))
	req.ContentLength = int64(len(storageBody))
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "storage should accept 5KB")

	// Auth endpoint should reject 1KB
	authBody := bytes.Repeat([]byte("a"), 1024)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(authBody))
	req.ContentLength = int64(len(authBody))
	resp, err = app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode, "auth should reject 1KB")
}

func TestJSONDepthLimiter_AcceptsShallowJSON(t *testing.T) {
	app := fiber.New()
	limiter := NewJSONDepthLimiter(5)
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Shallow JSON (depth 2)
	body := `{"user": {"name": "test"}}`
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestJSONDepthLimiter_RejectsDeepJSON(t *testing.T) {
	app := fiber.New()
	limiter := NewJSONDepthLimiter(3)
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Deep JSON (depth 5)
	body := `{"a": {"b": {"c": {"d": {"e": "value"}}}}}`
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestJSONDepthLimiter_SkipsNonJSON(t *testing.T) {
	app := fiber.New()
	limiter := NewJSONDepthLimiter(1) // Very strict
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Non-JSON content
	body := "this is plain text"
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestJSONDepthLimiter_SkipsGETRequests(t *testing.T) {
	app := fiber.New()
	limiter := NewJSONDepthLimiter(1)
	app.Use(limiter.Middleware())
	app.Get("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestJSONDepthLimiter_HandlesArrays(t *testing.T) {
	app := fiber.New()
	limiter := NewJSONDepthLimiter(3)
	app.Use(limiter.Middleware())
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Nested arrays (depth 4)
	body := `[[[["deep"]]]]`
	req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCheckJSONDepth(t *testing.T) {
	tests := []struct {
		name     string
		json     string
		maxDepth int
		wantErr  bool
		wantMax  int
	}{
		{
			name:     "empty object",
			json:     `{}`,
			maxDepth: 10,
			wantErr:  false,
			wantMax:  1,
		},
		{
			name:     "nested object within limit",
			json:     `{"a": {"b": "c"}}`,
			maxDepth: 3,
			wantErr:  false,
			wantMax:  2,
		},
		{
			name:     "nested object exceeds limit",
			json:     `{"a": {"b": {"c": "d"}}}`,
			maxDepth: 2,
			wantErr:  true,
			wantMax:  3,
		},
		{
			name:     "array within limit",
			json:     `[[1, 2, 3]]`,
			maxDepth: 3,
			wantErr:  false,
			wantMax:  2,
		},
		{
			name:     "mixed nesting",
			json:     `{"arr": [{"nested": true}]}`,
			maxDepth: 5,
			wantErr:  false,
			wantMax:  3,
		},
		{
			name:     "deeply nested array",
			json:     `[[[[[[[[[[1]]]]]]]]]]`,
			maxDepth: 5,
			wantErr:  true,
			wantMax:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			depth, err := checkJSONDepth([]byte(tt.json), tt.maxDepth)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantMax, depth)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{10485760, "10.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatBytes(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultBodyLimitConfig(t *testing.T) {
	config := DefaultBodyLimitConfig()

	assert.Equal(t, DefaultBodyLimit, config.DefaultLimit)
	assert.Equal(t, DefaultMaxJSONDepth, config.MaxJSONDepth)
	assert.NotEmpty(t, config.Patterns)

	// Verify some key patterns exist
	hasStoragePattern := false
	hasAuthPattern := false
	hasRESTPattern := false

	for _, p := range config.Patterns {
		if strings.Contains(p.Pattern, "storage") {
			hasStoragePattern = true
		}
		if strings.Contains(p.Pattern, "auth") {
			hasAuthPattern = true
		}
		if strings.Contains(p.Pattern, "rest") {
			hasRESTPattern = true
		}
	}

	assert.True(t, hasStoragePattern, "should have storage pattern")
	assert.True(t, hasAuthPattern, "should have auth pattern")
	assert.True(t, hasRESTPattern, "should have REST pattern")
}

func TestBodyLimitMiddleware_Combined(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 10 * 1024, // 10KB
		Patterns:     []BodyLimitPattern{},
		MaxJSONDepth: 3,
	}

	app := fiber.New()
	app.Use(BodyLimitMiddleware(config))
	app.Post("/api/test", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	t.Run("accepts valid request", func(t *testing.T) {
		body := `{"name": "test"}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("rejects oversized body", func(t *testing.T) {
		body := bytes.Repeat([]byte("a"), 20*1024) // 20KB
		req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.ContentLength = int64(len(body))

		resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		require.NoError(t, err)
		assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
	})

	t.Run("rejects deep JSON", func(t *testing.T) {
		body := `{"a": {"b": {"c": {"d": "value"}}}}`
		req := httptest.NewRequest(http.MethodPost, "/api/test", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, fiber.TestConfig{Timeout: 0, FailOnTimeout: false})
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestPatternMatching_EdgeCases(t *testing.T) {
	config := BodyLimitConfig{
		DefaultLimit: 100,
		Patterns: []BodyLimitPattern{
			{Pattern: "/api/v1/exact", Limit: 200, Description: "exact"},
			{Pattern: "/api/v1/wild/*", Limit: 300, Description: "single wild"},
			{Pattern: "/api/v1/double/**", Limit: 400, Description: "double wild"},
			{Pattern: "/api/v1/mixed/*/end", Limit: 500, Description: "mixed"},
			{Pattern: "/api/v1/complex/**/final", Limit: 600, Description: "complex"},
		},
	}

	limiter := NewPatternBodyLimiter(config)

	tests := []struct {
		path      string
		wantLimit int64
	}{
		// Exact match
		{"/api/v1/exact", 200},
		{"/api/v1/exact/extra", 100}, // No match - has extra segment

		// Single wildcard
		{"/api/v1/wild/anything", 300},
		{"/api/v1/wild/anything/more", 100}, // No match - too many segments

		// Double wildcard
		{"/api/v1/double/one", 400},
		{"/api/v1/double/one/two", 400},
		{"/api/v1/double/one/two/three/four", 400},

		// Mixed pattern
		{"/api/v1/mixed/anything/end", 500},
		{"/api/v1/mixed/something/end", 500},
		{"/api/v1/mixed/x/y/end", 100}, // No match - * only matches one segment

		// Complex pattern with ** in middle
		{"/api/v1/complex/final", 600},
		{"/api/v1/complex/a/final", 600},
		{"/api/v1/complex/a/b/c/final", 600},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			limit, _ := limiter.GetLimit(tt.path)
			assert.Equal(t, tt.wantLimit, limit, "limit mismatch for %s", tt.path)
		})
	}
}

func TestDefaultBodyLimitPatterns(t *testing.T) {
	patterns := DefaultBodyLimitPatterns()

	// Verify we have patterns
	assert.NotEmpty(t, patterns)

	// Build a map for easier verification
	patternMap := make(map[string]BodyLimitPattern)
	for _, p := range patterns {
		patternMap[p.Pattern] = p
	}

	// Verify specific patterns and their limits
	tests := []struct {
		pattern     string
		wantLimit   int64
		description string
	}{
		// Storage patterns - should use StorageUploadLimit
		{"/api/v1/storage/*/multipart", MultipartUploadLimit, "multipart upload"},
		{"/api/v1/storage/*/stream/**", StorageUploadLimit, "stream upload"},
		{"/api/v1/storage/*/chunked/**", StorageUploadLimit, "chunked upload"},
		{"/api/v1/storage/**", StorageUploadLimit, "storage"},

		// Admin sync patterns - should use StorageUploadLimit
		{"/api/v1/admin/functions/sync", StorageUploadLimit, "functions sync"},
		{"/api/v1/admin/jobs/sync", StorageUploadLimit, "jobs sync"},
		{"/api/v1/admin/ai/chatbots/sync", StorageUploadLimit, "chatbots sync"},
		{"/api/v1/admin/rpc/sync", StorageUploadLimit, "RPC sync"},
		{"/api/v1/admin/migrations/sync", StorageUploadLimit, "migrations sync"},

		// Admin general - should use AdminLimit
		{"/api/v1/admin/**", AdminLimit, "admin"},
		{"/api/v1/ai/**", AdminLimit, "AI/vectors"},

		// Auth - should use AuthBodyLimit
		{"/api/v1/auth/**", AuthBodyLimit, "auth"},

		// Webhooks - should use WebhookLimit
		{"/api/v1/webhooks/**", WebhookLimit, "webhooks"},
		{"/api/v1/functions/webhooks/**", WebhookLimit, "function webhooks"},

		// Bulk operations - should use LargePayloadLimit
		{"/api/v1/rest/*/bulk", LargePayloadLimit, "bulk operations"},
		{"/api/v1/rpc/**", LargePayloadLimit, "RPC"},

		// REST - should use RESTBodyLimit
		{"/api/v1/rest/**", RESTBodyLimit, "REST"},

		// GraphQL - should use LargePayloadLimit
		{"/graphql", LargePayloadLimit, "GraphQL"},

		// MCP - should use LargePayloadLimit
		{"/mcp/**", LargePayloadLimit, "MCP"},

		// Realtime - should use AuthBodyLimit
		{"/api/v1/realtime/**", AuthBodyLimit, "realtime"},

		// Default API
		{"/api/**", RESTBodyLimit, "API"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			p, exists := patternMap[tt.pattern]
			assert.True(t, exists, "pattern %s should exist", tt.pattern)
			if exists {
				assert.Equal(t, tt.wantLimit, p.Limit, "limit mismatch for %s", tt.pattern)
				assert.Equal(t, tt.description, p.Description, "description mismatch for %s", tt.pattern)
			}
		})
	}
}

func TestBodyLimitsFromConfig_AllDefaults(t *testing.T) {
	// All zero/negative values should use defaults
	config := BodyLimitsFromConfig(0, 0, 0, 0, 0, 0, 0)

	assert.Equal(t, DefaultBodyLimit, config.DefaultLimit)
	assert.Equal(t, DefaultMaxJSONDepth, config.MaxJSONDepth)
	assert.NotEmpty(t, config.Patterns)

	// Verify patterns use default limits
	limiter := NewPatternBodyLimiter(config)

	// Storage should use default StorageUploadLimit
	limit, desc := limiter.GetLimit("/api/v1/storage/bucket/file.txt")
	assert.Equal(t, StorageUploadLimit, limit)
	assert.Equal(t, "storage", desc)

	// Auth should use default AuthBodyLimit
	limit, desc = limiter.GetLimit("/api/v1/auth/login")
	assert.Equal(t, AuthBodyLimit, limit)
	assert.Equal(t, "auth", desc)

	// REST should use default RESTBodyLimit
	limit, desc = limiter.GetLimit("/api/v1/rest/users")
	assert.Equal(t, RESTBodyLimit, limit)
	assert.Equal(t, "REST", desc)

	// Admin should use default AdminLimit
	limit, desc = limiter.GetLimit("/api/v1/admin/settings")
	assert.Equal(t, AdminLimit, limit)
	assert.Equal(t, "admin", desc)

	// Bulk should use default LargePayloadLimit
	limit, desc = limiter.GetLimit("/api/v1/rest/users/bulk")
	assert.Equal(t, LargePayloadLimit, limit)
	assert.Equal(t, "bulk operations", desc)
}

func TestBodyLimitsFromConfig_NegativeValues(t *testing.T) {
	// Negative values should also use defaults
	config := BodyLimitsFromConfig(-1, -100, -50, -1000, -500, -200, -10)

	assert.Equal(t, DefaultBodyLimit, config.DefaultLimit)
	assert.Equal(t, DefaultMaxJSONDepth, config.MaxJSONDepth)
}

func TestBodyLimitsFromConfig_CustomValues(t *testing.T) {
	customDefault := int64(2 * 1024 * 1024)    // 2MB
	customREST := int64(5 * 1024 * 1024)       // 5MB
	customAuth := int64(128 * 1024)            // 128KB
	customStorage := int64(1024 * 1024 * 1024) // 1GB
	customBulk := int64(50 * 1024 * 1024)      // 50MB
	customAdmin := int64(20 * 1024 * 1024)     // 20MB
	customJSONDepth := 128

	config := BodyLimitsFromConfig(
		customDefault,
		customREST,
		customAuth,
		customStorage,
		customBulk,
		customAdmin,
		customJSONDepth,
	)

	assert.Equal(t, customDefault, config.DefaultLimit)
	assert.Equal(t, customJSONDepth, config.MaxJSONDepth)

	// Create limiter to verify patterns use custom values
	limiter := NewPatternBodyLimiter(config)

	// Storage should use custom storage limit
	limit, _ := limiter.GetLimit("/api/v1/storage/bucket/file.txt")
	assert.Equal(t, customStorage, limit)

	// Auth should use custom auth limit
	limit, _ = limiter.GetLimit("/api/v1/auth/login")
	assert.Equal(t, customAuth, limit)

	// REST should use custom REST limit
	limit, _ = limiter.GetLimit("/api/v1/rest/users")
	assert.Equal(t, customREST, limit)

	// Admin should use custom admin limit
	limit, _ = limiter.GetLimit("/api/v1/admin/settings")
	assert.Equal(t, customAdmin, limit)

	// Bulk should use custom bulk limit
	limit, _ = limiter.GetLimit("/api/v1/rest/users/bulk")
	assert.Equal(t, customBulk, limit)

	// RPC should use custom bulk limit
	limit, _ = limiter.GetLimit("/api/v1/rpc/my-function")
	assert.Equal(t, customBulk, limit)

	// GraphQL should use custom bulk limit
	limit, _ = limiter.GetLimit("/graphql")
	assert.Equal(t, customBulk, limit)

	// MCP should use custom bulk limit
	limit, _ = limiter.GetLimit("/mcp/tools")
	assert.Equal(t, customBulk, limit)

	// Realtime should use custom auth limit
	limit, _ = limiter.GetLimit("/api/v1/realtime/subscribe")
	assert.Equal(t, customAuth, limit)

	// Webhooks should use custom REST limit
	limit, _ = limiter.GetLimit("/api/v1/webhooks/github")
	assert.Equal(t, customREST, limit)

	// AI should use custom admin limit
	limit, _ = limiter.GetLimit("/api/v1/ai/vectors/search")
	assert.Equal(t, customAdmin, limit)

	// Admin sync endpoints should use custom storage limit
	limit, _ = limiter.GetLimit("/api/v1/admin/functions/sync")
	assert.Equal(t, customStorage, limit)
}

func TestBodyLimitsFromConfig_MixedDefaultsAndCustom(t *testing.T) {
	// Mix of custom and default values
	customREST := int64(10 * 1024 * 1024) // 10MB
	customAuth := int64(32 * 1024)        // 32KB

	config := BodyLimitsFromConfig(
		0,          // Use default
		customREST, // Custom
		customAuth, // Custom
		0,          // Use default
		0,          // Use default
		0,          // Use default
		0,          // Use default
	)

	assert.Equal(t, DefaultBodyLimit, config.DefaultLimit, "default limit should use default")
	assert.Equal(t, DefaultMaxJSONDepth, config.MaxJSONDepth, "JSON depth should use default")

	limiter := NewPatternBodyLimiter(config)

	// REST should use custom value
	limit, _ := limiter.GetLimit("/api/v1/rest/users")
	assert.Equal(t, customREST, limit)

	// Auth should use custom value
	limit, _ = limiter.GetLimit("/api/v1/auth/login")
	assert.Equal(t, customAuth, limit)

	// Storage should use default
	limit, _ = limiter.GetLimit("/api/v1/storage/bucket/file.txt")
	assert.Equal(t, StorageUploadLimit, limit)

	// Bulk should use default
	limit, _ = limiter.GetLimit("/api/v1/rest/users/bulk")
	assert.Equal(t, LargePayloadLimit, limit)

	// Admin should use default
	limit, _ = limiter.GetLimit("/api/v1/admin/settings")
	assert.Equal(t, AdminLimit, limit)
}

func TestBodyLimitsFromConfig_PatternCount(t *testing.T) {
	config := BodyLimitsFromConfig(0, 0, 0, 0, 0, 0, 0)

	// Should have the same number of patterns as DefaultBodyLimitPatterns
	defaultPatterns := DefaultBodyLimitPatterns()
	assert.Equal(t, len(defaultPatterns), len(config.Patterns), "pattern count should match default patterns")
}

func TestNewJSONDepthLimiter_DefaultDepth(t *testing.T) {
	// Zero or negative depth should use default
	limiter := NewJSONDepthLimiter(0)
	assert.Equal(t, DefaultMaxJSONDepth, limiter.maxDepth)

	limiter = NewJSONDepthLimiter(-5)
	assert.Equal(t, DefaultMaxJSONDepth, limiter.maxDepth)
}

func TestNewJSONDepthLimiter_CustomDepth(t *testing.T) {
	limiter := NewJSONDepthLimiter(10)
	assert.Equal(t, 10, limiter.maxDepth)

	limiter = NewJSONDepthLimiter(100)
	assert.Equal(t, 100, limiter.maxDepth)
}
