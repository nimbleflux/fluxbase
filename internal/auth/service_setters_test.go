package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Service Setter Methods Tests
// =============================================================================

func TestService_SetEncryptionKey(t *testing.T) {
	service := &Service{}

	t.Run("sets encryption key", func(t *testing.T) {
		key := "test-encryption-key-123"
		service.SetEncryptionKey(key)

		assert.Equal(t, key, service.encryptionKey)
	})

	t.Run("can be updated multiple times", func(t *testing.T) {
		service.SetEncryptionKey("key1")
		assert.Equal(t, "key1", service.encryptionKey)

		service.SetEncryptionKey("key2")
		assert.Equal(t, "key2", service.encryptionKey)
	})

	t.Run("empty string is allowed", func(t *testing.T) {
		service.SetEncryptionKey("")
		assert.Equal(t, "", service.encryptionKey)
	})
}

func TestService_SetTOTPRateLimiter(t *testing.T) {
	service := &Service{}

	t.Run("sets TOTP rate limiter", func(t *testing.T) {
		limiter := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter)

		assert.Same(t, limiter, service.totpRateLimiter)
	})

	t.Run("can be updated", func(t *testing.T) {
		limiter1 := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter1)
		assert.Same(t, limiter1, service.totpRateLimiter)

		limiter2 := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter2)
		assert.Same(t, limiter2, service.totpRateLimiter)
	})

	t.Run("nil is allowed", func(t *testing.T) {
		service.SetTOTPRateLimiter(nil)
		assert.Nil(t, service.totpRateLimiter)
	})
}

func TestService_SetMetrics(t *testing.T) {
	service := &Service{}

	t.Run("sets metrics instance", func(t *testing.T) {
		// We can't easily create a real Metrics instance, so we'll test with nil
		service.SetMetrics(nil)

		assert.Nil(t, service.metrics)
	})
}

func TestService_RecordAuthAttempt(t *testing.T) {
	service := &Service{}

	t.Run("does not panic with nil metrics", func(t *testing.T) {
		// Should not panic even though metrics is nil
		assert.NotPanics(t, func() {
			service.recordAuthAttempt("password", true, "")
		})
	})

	t.Run("does not panic with nil metrics and failure", func(t *testing.T) {
		assert.NotPanics(t, func() {
			service.recordAuthAttempt("otp", false, "invalid_code")
		})
	})
}

func TestService_RecordAuthToken(t *testing.T) {
	service := &Service{}

	t.Run("does not panic with nil metrics", func(t *testing.T) {
		// Should not panic even though metrics is nil
		assert.NotPanics(t, func() {
			service.recordAuthToken("access")
		})
	})

	t.Run("does not panic with different token types", func(t *testing.T) {
		assert.NotPanics(t, func() {
			service.recordAuthToken("refresh")
		})

		assert.NotPanics(t, func() {
			service.recordAuthToken("service_role")
		})
	})
}
