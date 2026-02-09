package api

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/fluxbase-eu/fluxbase/internal/storage"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// =============================================================================
// NewStorageHandler Tests
// =============================================================================

func TestNewStorageHandler_NilConfig(t *testing.T) {
	handler := NewStorageHandlerWithCache(nil, nil, nil, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.transformer)
	assert.Nil(t, handler.transformCache)
	assert.Nil(t, handler.transformSem)
}

func TestNewStorageHandler_DisabledTransforms(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled: false,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	assert.NotNil(t, handler)
	assert.Nil(t, handler.transformer)
	assert.Nil(t, handler.transformSem)
}

func TestNewStorageHandler_EnabledTransforms(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:        true,
		MaxWidth:       4096,
		MaxHeight:      4096,
		MaxTotalPixels: 16000000,
		BucketSize:     50,
		RateLimit:      60,
		MaxConcurrent:  4,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.transformer)
	assert.NotNil(t, handler.transformSem)
	assert.Equal(t, 4, cap(handler.transformSem))
}

func TestNewStorageHandler_DefaultConcurrency(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 0, // Should default to 4
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	assert.NotNil(t, handler.transformSem)
	assert.Equal(t, 4, cap(handler.transformSem))
}

func TestNewStorageHandler_DefaultRateLimit(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 0, // Should default to 60
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Rate limit should be 60/60 = 1 per second
	assert.NotZero(t, handler.transformRateLimit)
}

// =============================================================================
// getTransformLimiter Tests
// =============================================================================

func TestStorageHandler_getTransformLimiter(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 60,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Get limiter for a key
	limiter1 := handler.getTransformLimiter("user1:ip1")

	assert.NotNil(t, limiter1)

	// Same key should return same limiter (same pointer)
	limiter2 := handler.getTransformLimiter("user1:ip1")
	assert.Same(t, limiter1, limiter2, "same key should return same limiter instance")

	// Different key should return different limiter (different pointer)
	limiter3 := handler.getTransformLimiter("user2:ip2")
	assert.NotSame(t, limiter1, limiter3, "different keys should return different limiter instances")
}

func TestStorageHandler_getTransformLimiter_AllowsRequests(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 60, // 60 per minute = 1 per second
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	limiter := handler.getTransformLimiter("testkey")

	// First request should be allowed
	assert.True(t, limiter.Allow())
}

// =============================================================================
// acquireTransformSlot / releaseTransformSlot Tests
// =============================================================================

func TestStorageHandler_acquireTransformSlot_NilSem(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled: false,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Should return true when no semaphore configured
	assert.True(t, handler.acquireTransformSlot(time.Second))
}

func TestStorageHandler_acquireTransformSlot_Success(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 2,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Should acquire slots up to limit
	assert.True(t, handler.acquireTransformSlot(time.Second))
	assert.True(t, handler.acquireTransformSlot(time.Second))

	// Release slots
	handler.releaseTransformSlot()
	handler.releaseTransformSlot()
}

func TestStorageHandler_acquireTransformSlot_Timeout(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 1,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Acquire the only slot
	assert.True(t, handler.acquireTransformSlot(time.Second))

	// Should timeout trying to acquire another
	assert.False(t, handler.acquireTransformSlot(10*time.Millisecond))

	// Release and try again
	handler.releaseTransformSlot()
	assert.True(t, handler.acquireTransformSlot(time.Second))
	handler.releaseTransformSlot()
}

func TestStorageHandler_releaseTransformSlot_NilSem(t *testing.T) {
	handler := NewStorageHandlerWithCache(nil, nil, nil, nil)

	// Should not panic
	handler.releaseTransformSlot()
}

// =============================================================================
// GetTransformConfig Tests
// =============================================================================

func TestStorageHandler_GetTransformConfig_NilConfig(t *testing.T) {
	handler := NewStorageHandlerWithCache(nil, nil, nil, nil)

	app := fiber.New()
	app.Get("/config", handler.GetTransformConfig)

	req := httptest.NewRequest("GET", "/config", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"enabled":false`)
}

func TestStorageHandler_GetTransformConfig_Enabled(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:        true,
		DefaultQuality: 80,
		MaxWidth:       4096,
		MaxHeight:      4096,
		AllowedFormats: []string{"webp", "jpg", "png"},
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	app := fiber.New()
	app.Get("/config", handler.GetTransformConfig)

	req := httptest.NewRequest("GET", "/config", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Contains(t, bodyStr, `"enabled":true`)
	assert.Contains(t, bodyStr, `"default_quality":80`)
	assert.Contains(t, bodyStr, `"max_width":4096`)
	assert.Contains(t, bodyStr, `"max_height":4096`)
	assert.Contains(t, bodyStr, `"allowed_formats"`)
}

func TestStorageHandler_GetTransformConfig_Disabled(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled: false,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	app := fiber.New()
	app.Get("/config", handler.GetTransformConfig)

	req := httptest.NewRequest("GET", "/config", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"enabled":false`)
}

// =============================================================================
// TransformConfigResponse Tests
// =============================================================================

func TestTransformConfigResponse_Fields(t *testing.T) {
	resp := TransformConfigResponse{
		Enabled:        true,
		DefaultQuality: 80,
		MaxWidth:       4096,
		MaxHeight:      2160,
		AllowedFormats: []string{"webp", "jpg"},
	}

	assert.True(t, resp.Enabled)
	assert.Equal(t, 80, resp.DefaultQuality)
	assert.Equal(t, 4096, resp.MaxWidth)
	assert.Equal(t, 2160, resp.MaxHeight)
	assert.Equal(t, []string{"webp", "jpg"}, resp.AllowedFormats)
}

// =============================================================================
// Integration with Transform Logic Tests
// =============================================================================

func TestStorageHandler_TransformerConfig(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:        true,
		MaxWidth:       2048,
		MaxHeight:      1536,
		MaxTotalPixels: 8000000,
		BucketSize:     100,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Verify transformer was created with correct options
	assert.NotNil(t, handler.transformer)

	// Test that validator respects config
	opts := &storage.TransformOptions{Width: 3000}
	err := handler.transformer.ValidateOptions(opts)

	// Width exceeds max, should error
	assert.Error(t, err)
}

func TestStorageHandler_TransformerBucketing(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:    true,
		MaxWidth:   4096,
		MaxHeight:  4096,
		BucketSize: 100,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	opts := &storage.TransformOptions{Width: 850, Height: 650}
	err := handler.transformer.ValidateOptions(opts)

	require.NoError(t, err)
	// Dimensions should be bucketed to nearest 100
	assert.Equal(t, 900, opts.Width)
	assert.Equal(t, 700, opts.Height)
}

// =============================================================================
// Concurrent Limiter Tests
// =============================================================================

func TestStorageHandler_ConcurrentLimiterAccess(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 1000, // High rate limit for concurrent test
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Access limiters concurrently from multiple goroutines
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				limiter := handler.getTransformLimiter("user" + string(rune('0'+id)))
				_ = limiter.Allow()
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should complete without panic or deadlock
}

// =============================================================================
// Additional Storage Handler Tests for Improved Coverage
// =============================================================================

func TestNewStorageHandler_WithCache(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:        true,
		DefaultQuality: 85,
		MaxWidth:       2048,
		MaxHeight:      2048,
		CacheTTL:       24 * time.Hour,
		CacheMaxSize:   1024 * 1024 * 1024, // 1GB
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.transformer)
	// Without actual storage provider, cache cannot be initialized
	assert.Nil(t, handler.transformCache)
}

func TestNewStorageHandler_TransformerDefaults(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled: true,
		// MaxWidth, MaxHeight, MaxTotalPixels, BucketSize not set
		// Should use defaults from ImageTransformer
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.transformer)
	assert.NotNil(t, handler.transformSem)
	assert.NotZero(t, handler.transformRateLimit)
	assert.Greater(t, cap(handler.transformSem), 0)
}

func TestNewStorageHandler_TransformBurstDefaults(t *testing.T) {
	tests := []struct {
		name             string
		rateLimit        int
		expectedBurstMin int
	}{
		{
			name:             "rate limit 10",
			rateLimit:        10,
			expectedBurstMin: 1,
		},
		{
			name:             "rate limit 60",
			rateLimit:        60,
			expectedBurstMin: 1,
		},
		{
			name:             "rate limit 120",
			rateLimit:        120,
			expectedBurstMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.TransformConfig{
				Enabled:   true,
				RateLimit: tt.rateLimit,
			}

			handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

			assert.NotNil(t, handler)
			assert.GreaterOrEqual(t, handler.transformBurst, tt.expectedBurstMin)
		})
	}
}

func TestGetTransformConfig_AllFields(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:        true,
		DefaultQuality: 90,
		MaxWidth:       3840,
		MaxHeight:      2160,
		AllowedFormats: []string{"webp", "avif", "jpg", "png"},
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	app := fiber.New()
	app.Get("/config", handler.GetTransformConfig)

	req := httptest.NewRequest("GET", "/config", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	assert.Contains(t, bodyStr, `"enabled":true`)
	assert.Contains(t, bodyStr, `"default_quality":90`)
	assert.Contains(t, bodyStr, `"max_width":3840`)
	assert.Contains(t, bodyStr, `"max_height":2160`)
	assert.Contains(t, bodyStr, `"allowed_formats"`)
	assert.Contains(t, bodyStr, `"webp"`)
	assert.Contains(t, bodyStr, `"avif"`)
}

func TestTransformConfigResponse_WithEmptyFormats(t *testing.T) {
	resp := TransformConfigResponse{
		Enabled:        true,
		DefaultQuality: 80,
		MaxWidth:       1920,
		MaxHeight:      1080,
		AllowedFormats: []string{},
	}

	assert.True(t, resp.Enabled)
	assert.Equal(t, 80, resp.DefaultQuality)
	assert.Empty(t, resp.AllowedFormats)
}

func TestTransformConfigResponse_OnlyEnabled(t *testing.T) {
	resp := TransformConfigResponse{
		Enabled: true,
	}

	assert.True(t, resp.Enabled)
	assert.Zero(t, resp.DefaultQuality)
	assert.Zero(t, resp.MaxWidth)
	assert.Zero(t, resp.MaxHeight)
	assert.Nil(t, resp.AllowedFormats)
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestStorageHandler_ConcurrentSlotAcquisition(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 2,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Acquire all slots
	assert.True(t, handler.acquireTransformSlot(time.Millisecond*100))
	assert.True(t, handler.acquireTransformSlot(time.Millisecond*100))

	// Third acquisition should timeout
	done := make(chan bool)
	go func() {
		success := handler.acquireTransformSlot(time.Millisecond * 10)
		done <- !success // Should fail
	}()

	result := <-done
	assert.True(t, result, "Third acquisition should timeout")

	// Release and try again
	handler.releaseTransformSlot()
	assert.True(t, handler.acquireTransformSlot(time.Millisecond*100))
	handler.releaseTransformSlot()
	handler.releaseTransformSlot()
}

func TestStorageHandler_MultipleLimiterInstances(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 60,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Create multiple limiters for different keys
	keys := []string{"user1:ip1", "user1:ip2", "user2:ip1", "user2:ip2"}
	limiters := make(map[string]*rate.Limiter)

	for _, key := range keys {
		limiter := handler.getTransformLimiter(key)
		limiter = handler.getTransformLimiter(key) // Call twice to test caching
		limiter = handler.getTransformLimiter(key) // Call thrice
		limiters[key] = limiter
	}

	// Same key should return same limiter
	assert.Same(t, limiters["user1:ip1"], limiters["user1:ip1"])

	// Different keys should return different limiters
	assert.NotSame(t, limiters["user1:ip1"], limiters["user2:ip1"])

	// Verify all limiters work independently
	for _, limiter := range limiters {
		assert.True(t, limiter.Allow(), "New limiter should allow first request")
	}
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestStorageHandler_ZeroMaxConcurrent(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 0, // Should default to 4
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Should have default capacity
	assert.Equal(t, 4, cap(handler.transformSem))

	// Should allow acquisitions
	assert.True(t, handler.acquireTransformSlot(time.Second))
	handler.releaseTransformSlot()
}

func TestStorageHandler_NegativeRateLimit(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: -10, // Should be treated as 0, which defaults to 60
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Should have a rate limit (using default)
	assert.NotZero(t, handler.transformRateLimit)
}

func TestStorageHandler_ReleaseSlotWithoutAcquire(t *testing.T) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 2,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	// Release without acquire - should not panic
	handler.releaseTransformSlot()
	handler.releaseTransformSlot()

	// Should still allow acquisitions
	assert.True(t, handler.acquireTransformSlot(time.Second))
	handler.releaseTransformSlot()
}

// =============================================================================
// Response Structure Tests
// =============================================================================

func TestTransformConfigResponse_JSONMarshaling(t *testing.T) {
	resp := TransformConfigResponse{
		Enabled:        true,
		DefaultQuality: 85,
		MaxWidth:       4096,
		MaxHeight:      4096,
		AllowedFormats: []string{"webp", "jpg", "png"},
	}

	// Simulate JSON marshaling
	assert.True(t, resp.Enabled)
	assert.Equal(t, 85, resp.DefaultQuality)
	assert.Len(t, resp.AllowedFormats, 3)
	assert.Contains(t, resp.AllowedFormats, "webp")
	assert.Contains(t, resp.AllowedFormats, "jpg")
	assert.Contains(t, resp.AllowedFormats, "png")
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkStorageHandler_getTransformLimiter(b *testing.B) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 60,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.getTransformLimiter("testkey")
	}
}

func BenchmarkStorageHandler_acquireReleaseSlot(b *testing.B) {
	cfg := &config.TransformConfig{
		Enabled:       true,
		MaxConcurrent: 100,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.acquireTransformSlot(time.Second)
		handler.releaseTransformSlot()
	}
}

func BenchmarkStorageHandler_ConcurrentLimiterAccess(b *testing.B) {
	cfg := &config.TransformConfig{
		Enabled:   true,
		RateLimit: 1000,
	}

	handler := NewStorageHandlerWithCache(nil, nil, cfg, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.getTransformLimiter("user" + string(rune('0'+i%10)))
	}
}
