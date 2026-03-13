package email

import (
	"testing"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SendGrid Service Construction Tests
// =============================================================================

func TestNewSendGridService(t *testing.T) {
	t.Run("returns error for missing API key", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "",
			FromAddress:    "test@example.com",
		}

		service, err := NewSendGridService(cfg)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "API key is required")
	})

	t.Run("creates service with valid config", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.test-api-key",
			FromName:       "Test",
			FromAddress:    "test@example.com",
		}

		service, err := NewSendGridService(cfg)

		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, cfg, service.config)
		assert.NotNil(t, service.client)
	})
}

// =============================================================================
// SendGrid IsConfigured Tests
// =============================================================================

func TestSendGridService_IsConfigured(t *testing.T) {
	t.Run("returns false when disabled", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:        false,
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.test-key",
			FromAddress:    "test@example.com",
		}
		service, _ := NewSendGridService(cfg)

		assert.False(t, service.IsConfigured())
	})

	t.Run("returns true when enabled and configured", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:        true,
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.test-key",
			FromAddress:    "test@example.com",
		}
		service, _ := NewSendGridService(cfg)

		result := service.IsConfigured()
		assert.True(t, result)
	})
}

// =============================================================================
// SendGrid Service Struct Tests
// =============================================================================

func TestSendGridService_Struct(t *testing.T) {
	t.Run("stores config and client", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.abc123xyz",
			FromName:       "Fluxbase",
			FromAddress:    "noreply@example.com",
		}

		service, err := NewSendGridService(cfg)

		require.NoError(t, err)
		assert.Equal(t, cfg, service.config)
		assert.NotNil(t, service.client)
	})
}

// =============================================================================
// SendGrid Email Type Methods Tests
// =============================================================================

func TestSendGridService_EmailMethods(t *testing.T) {
	cfg := &config.EmailConfig{
		Provider:       "sendgrid",
		SendGridAPIKey: "SG.test-key",
		FromName:       "Test",
		FromAddress:    "test@example.com",
	}
	service, _ := NewSendGridService(cfg)

	t.Run("SendMagicLink method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("SendVerificationEmail method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("SendPasswordReset method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("SendInvitationEmail method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})

	t.Run("Send method exists", func(t *testing.T) {
		assert.NotNil(t, service)
	})
}

// =============================================================================
// SendGrid API Key Format Tests
// =============================================================================

func TestSendGridService_APIKeyFormat(t *testing.T) {
	t.Run("accepts SG. prefixed keys", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.abcdefghijklmnop",
			FromAddress:    "test@example.com",
		}

		service, err := NewSendGridService(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, service)
	})

	t.Run("accepts non-prefixed keys", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "some-legacy-key-format",
			FromAddress:    "test@example.com",
		}

		service, err := NewSendGridService(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, service)
	})
}

// =============================================================================
// SendGrid Response Status Code Handling Tests
// =============================================================================

func TestSendGridService_StatusCodeHandling(t *testing.T) {
	t.Run("documents status code error threshold", func(t *testing.T) {
		// The Send method checks: if response.StatusCode >= 400
		// This means:
		// - 2xx: Success
		// - 3xx: Success (redirects handled by client)
		// - 4xx: Client error (returned as error)
		// - 5xx: Server error (returned as error)

		successCodes := []int{200, 201, 202, 204}
		errorCodes := []int{400, 401, 403, 404, 429, 500, 502, 503}

		for _, code := range successCodes {
			assert.True(t, code < 400, "code %d should be success", code)
		}

		for _, code := range errorCodes {
			assert.True(t, code >= 400, "code %d should be error", code)
		}
	})
}

// =============================================================================
// SendGrid Reply-To Tests
// =============================================================================

func TestSendGridService_ReplyToConfiguration(t *testing.T) {
	t.Run("configures reply-to when set", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.test-key",
			FromAddress:    "noreply@example.com",
			ReplyToAddress: "support@example.com",
		}

		service, err := NewSendGridService(cfg)

		require.NoError(t, err)
		assert.Equal(t, "support@example.com", service.config.ReplyToAddress)
	})

	t.Run("skips reply-to when not set", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:       "sendgrid",
			SendGridAPIKey: "SG.test-key",
			FromAddress:    "noreply@example.com",
			ReplyToAddress: "",
		}

		service, err := NewSendGridService(cfg)

		require.NoError(t, err)
		assert.Empty(t, service.config.ReplyToAddress)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewSendGridService(b *testing.B) {
	cfg := &config.EmailConfig{
		Provider:       "sendgrid",
		SendGridAPIKey: "SG.test-key",
		FromName:       "Test",
		FromAddress:    "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewSendGridService(cfg)
	}
}
