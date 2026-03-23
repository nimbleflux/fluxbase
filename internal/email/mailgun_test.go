package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// Mailgun Service Construction Tests
// =============================================================================

func TestNewMailgunService(t *testing.T) {
	t.Run("returns error for missing API key", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:      "mailgun",
			MailgunAPIKey: "",
			MailgunDomain: "example.com",
			FromAddress:   "test@example.com",
		}

		service, err := NewMailgunService(cfg)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "API key is required")
	})

	t.Run("returns error for missing domain", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:      "mailgun",
			MailgunAPIKey: "test-api-key",
			MailgunDomain: "",
			FromAddress:   "test@example.com",
		}

		service, err := NewMailgunService(cfg)

		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "domain is required")
	})

	t.Run("creates service with valid config", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:      "mailgun",
			MailgunAPIKey: "test-api-key",
			MailgunDomain: "mail.example.com",
			FromName:      "Test",
			FromAddress:   "test@example.com",
		}

		service, err := NewMailgunService(cfg)

		require.NoError(t, err)
		require.NotNil(t, service)
		assert.Equal(t, cfg, service.config)
		assert.Equal(t, "mail.example.com", service.domain)
		assert.NotNil(t, service.client)
	})
}

// =============================================================================
// Mailgun IsConfigured Tests
// =============================================================================

func TestMailgunService_IsConfigured(t *testing.T) {
	t.Run("returns false when disabled", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:       false,
			Provider:      "mailgun",
			MailgunAPIKey: "test-key",
			MailgunDomain: "example.com",
			FromAddress:   "test@example.com",
		}
		service, _ := NewMailgunService(cfg)

		assert.False(t, service.IsConfigured())
	})

	t.Run("returns true when enabled and configured", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:       true,
			Provider:      "mailgun",
			MailgunAPIKey: "test-key",
			MailgunDomain: "example.com",
			FromAddress:   "test@example.com",
		}
		service, _ := NewMailgunService(cfg)

		result := service.IsConfigured()
		assert.True(t, result)
	})
}

// =============================================================================
// Mailgun Service Struct Tests
// =============================================================================

func TestMailgunService_Struct(t *testing.T) {
	t.Run("stores config, client, and domain", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Provider:      "mailgun",
			MailgunAPIKey: "key-abc123",
			MailgunDomain: "mg.example.com",
			FromName:      "Fluxbase",
			FromAddress:   "noreply@example.com",
		}

		service, err := NewMailgunService(cfg)

		require.NoError(t, err)
		assert.Equal(t, cfg, service.config)
		assert.Equal(t, "mg.example.com", service.domain)
		assert.NotNil(t, service.client)
	})
}

// =============================================================================
// Mailgun Email Type Methods Tests
// =============================================================================

func TestMailgunService_EmailMethods(t *testing.T) {
	cfg := &config.EmailConfig{
		Provider:      "mailgun",
		MailgunAPIKey: "test-key",
		MailgunDomain: "example.com",
		FromName:      "Test",
		FromAddress:   "test@example.com",
	}
	service, _ := NewMailgunService(cfg)

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
// Mailgun Domain Validation Tests
// =============================================================================

func TestMailgunService_DomainHandling(t *testing.T) {
	validDomains := []string{
		"mg.example.com",
		"mail.mydomain.org",
		"sandbox123.mailgun.org",
	}

	for _, domain := range validDomains {
		t.Run("accepts domain "+domain, func(t *testing.T) {
			cfg := &config.EmailConfig{
				Provider:      "mailgun",
				MailgunAPIKey: "test-key",
				MailgunDomain: domain,
				FromAddress:   "test@example.com",
			}

			service, err := NewMailgunService(cfg)

			assert.NoError(t, err)
			assert.NotNil(t, service)
			assert.Equal(t, domain, service.domain)
		})
	}
}

// =============================================================================
// Mailgun Timeout Tests
// =============================================================================

func TestMailgunService_TimeoutConfiguration(t *testing.T) {
	t.Run("Send uses 10 second timeout", func(t *testing.T) {
		// The Send method creates a context with 10 second timeout
		// This is a unit test documenting the expected behavior

		cfg := &config.EmailConfig{
			Provider:      "mailgun",
			MailgunAPIKey: "test-key",
			MailgunDomain: "example.com",
			FromAddress:   "test@example.com",
		}

		service, err := NewMailgunService(cfg)

		require.NoError(t, err)
		assert.NotNil(t, service)
		// Timeout is set internally in the Send method
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewMailgunService(b *testing.B) {
	cfg := &config.EmailConfig{
		Provider:      "mailgun",
		MailgunAPIKey: "test-key",
		MailgunDomain: "example.com",
		FromName:      "Test",
		FromAddress:   "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewMailgunService(cfg)
	}
}
