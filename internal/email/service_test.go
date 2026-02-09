package email

import (
	"context"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NewService Tests
// =============================================================================

func TestNewService(t *testing.T) {
	t.Run("disabled email returns NoOpService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled: false,
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())

		// Verify it's a NoOpService by checking error message
		err = service.Send(context.Background(), "test@example.com", "Subject", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "disabled")
	})

	t.Run("unsupported provider returns error", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:  true,
			Provider: "unsupported_provider",
		}

		service, err := NewService(cfg)

		require.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "unsupported email provider")
	})

	t.Run("smtp provider not configured returns NoOpService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "smtp",
			FromAddress: "test@example.com",
			// Missing SMTPHost and SMTPPort
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())
	})

	t.Run("smtp provider fully configured returns SMTPService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "smtp",
			FromAddress: "test@example.com",
			SMTPHost:    "smtp.example.com",
			SMTPPort:    587,
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.True(t, service.IsConfigured())
	})

	t.Run("empty provider defaults to smtp", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "", // Empty
			FromAddress: "test@example.com",
			SMTPHost:    "smtp.example.com",
			SMTPPort:    587,
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.True(t, service.IsConfigured())
	})

	t.Run("sendgrid not configured returns NoOpService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "sendgrid",
			FromAddress: "test@example.com",
			// Missing SendGridAPIKey
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())
	})

	t.Run("sendgrid fully configured returns SendGridService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:        true,
			Provider:       "sendgrid",
			FromAddress:    "test@example.com",
			SendGridAPIKey: "SG.test-api-key",
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.True(t, service.IsConfigured())
	})

	t.Run("mailgun not configured returns NoOpService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "mailgun",
			FromAddress: "test@example.com",
			// Missing MailgunAPIKey and MailgunDomain
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())
	})

	t.Run("mailgun fully configured returns MailgunService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:       true,
			Provider:      "mailgun",
			FromAddress:   "test@example.com",
			MailgunAPIKey: "test-api-key",
			MailgunDomain: "mg.example.com",
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.True(t, service.IsConfigured())
	})

	t.Run("ses not configured returns NoOpService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			Provider:    "ses",
			FromAddress: "test@example.com",
			// Missing SESAccessKey, SESSecretKey, SESRegion
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())
	})

	t.Run("ses fully configured returns SESService", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:      true,
			Provider:     "ses",
			FromAddress:  "test@example.com",
			SESAccessKey: "AKIAIOSFODNN7EXAMPLE",
			SESSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			SESRegion:    "us-east-1",
		}

		service, err := NewService(cfg)

		require.NoError(t, err)
		assert.True(t, service.IsConfigured())
	})
}

// =============================================================================
// NoOpService Tests
// =============================================================================

func TestNoOpService(t *testing.T) {
	t.Run("NewNoOpService creates service with reason", func(t *testing.T) {
		service := NewNoOpService("test reason")

		assert.NotNil(t, service)
		assert.Equal(t, "test reason", service.reason)
	})

	t.Run("IsConfigured returns false", func(t *testing.T) {
		service := NewNoOpService("not configured")

		assert.False(t, service.IsConfigured())
	})

	t.Run("SendMagicLink returns error with reason", func(t *testing.T) {
		service := NewNoOpService("email disabled")

		err := service.SendMagicLink(context.Background(), "user@example.com", "token", "https://example.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "email disabled")
		assert.Contains(t, err.Error(), "cannot send email")
	})

	t.Run("SendVerificationEmail returns error with reason", func(t *testing.T) {
		service := NewNoOpService("not configured")

		err := service.SendVerificationEmail(context.Background(), "user@example.com", "token", "https://example.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("SendPasswordReset returns error with reason", func(t *testing.T) {
		service := NewNoOpService("provider not set")

		err := service.SendPasswordReset(context.Background(), "user@example.com", "token", "https://example.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider not set")
	})

	t.Run("SendInvitationEmail returns error with reason", func(t *testing.T) {
		service := NewNoOpService("missing API key")

		err := service.SendInvitationEmail(context.Background(), "user@example.com", "Inviter", "https://example.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing API key")
	})

	t.Run("Send returns error with reason", func(t *testing.T) {
		service := NewNoOpService("smtp server not reachable")

		err := service.Send(context.Background(), "user@example.com", "Subject", "Body")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "smtp server not reachable")
	})

	t.Run("empty reason still works", func(t *testing.T) {
		service := NewNoOpService("")

		err := service.Send(context.Background(), "user@example.com", "Subject", "Body")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot send email")
	})
}

// =============================================================================
// Service Interface Tests
// =============================================================================

func TestServiceInterface(t *testing.T) {
	t.Run("NoOpService implements Service interface", func(t *testing.T) {
		var _ Service = (*NoOpService)(nil)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewService_SMTP(b *testing.B) {
	cfg := &config.EmailConfig{
		Enabled:     true,
		Provider:    "smtp",
		FromAddress: "test@example.com",
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
	}

	for i := 0; i < b.N; i++ {
		_, _ = NewService(cfg)
	}
}

func BenchmarkNewService_Disabled(b *testing.B) {
	cfg := &config.EmailConfig{
		Enabled: false,
	}

	for i := 0; i < b.N; i++ {
		_, _ = NewService(cfg)
	}
}

func BenchmarkNoOpService_Send(b *testing.B) {
	service := NewNoOpService("benchmark")
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		_ = service.Send(ctx, "test@example.com", "Subject", "Body")
	}
}

// =============================================================================
// TestEmailService Tests
// =============================================================================

func TestTestEmailService(t *testing.T) {
	t.Run("NewTestEmailService creates service", func(t *testing.T) {
		service := NewTestEmailService()

		assert.NotNil(t, service)
		assert.True(t, service.IsConfigured())
	})

	t.Run("SendMagicLink succeeds", func(t *testing.T) {
		service := NewTestEmailService()

		err := service.SendMagicLink(context.Background(), "user@example.com", "token", "https://example.com")

		assert.NoError(t, err)
	})

	t.Run("SendVerificationEmail succeeds", func(t *testing.T) {
		service := NewTestEmailService()

		err := service.SendVerificationEmail(context.Background(), "user@example.com", "token", "https://example.com")

		assert.NoError(t, err)
	})

	t.Run("SendPasswordReset succeeds", func(t *testing.T) {
		service := NewTestEmailService()

		err := service.SendPasswordReset(context.Background(), "user@example.com", "token", "https://example.com")

		assert.NoError(t, err)
	})

	t.Run("SendInvitationEmail succeeds", func(t *testing.T) {
		service := NewTestEmailService()

		err := service.SendInvitationEmail(context.Background(), "user@example.com", "Inviter Name", "https://example.com")

		assert.NoError(t, err)
	})

	t.Run("Send succeeds", func(t *testing.T) {
		service := NewTestEmailService()

		err := service.Send(context.Background(), "user@example.com", "Subject", "Body")

		assert.NoError(t, err)
	})

	t.Run("IsConfigured returns true", func(t *testing.T) {
		service := NewTestEmailService()

		assert.True(t, service.IsConfigured())
	})
}

// =============================================================================
// Additional NewService Edge Cases
// =============================================================================

func TestNewService_EdgeCases(t *testing.T) {
	t.Run("nil config returns NoOpService", func(t *testing.T) {
		service, err := NewService(nil)

		require.NoError(t, err)
		assert.False(t, service.IsConfigured())

		// Should be NoOpService
		err = service.Send(context.Background(), "test@example.com", "Subject", "Body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "disabled")
	})
}

// =============================================================================
// Service Interface Compliance Tests
// =============================================================================

func TestAllServicesImplementInterface(t *testing.T) {
	t.Run("NoOpService implements Service", func(t *testing.T) {
		var _ Service = &NoOpService{}
	})

	t.Run("TestEmailService implements Service", func(t *testing.T) {
		var _ Service = &TestEmailService{}
	})
}

// =============================================================================
// Context Handling Tests
// =============================================================================

func TestNoOpService_ContextHandling(t *testing.T) {
	service := NewNoOpService("test")

	t.Run("handles nil context", func(t *testing.T) {
		// Should not panic
		err := service.SendMagicLink(nil, "user@example.com", "token", "link")
		assert.Error(t, err)
	})

	t.Run("handles cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := service.Send(ctx, "user@example.com", "Subject", "Body")
		assert.Error(t, err)
	})
}

func TestTestEmailService_ContextHandling(t *testing.T) {
	service := NewTestEmailService()

	t.Run("handles nil context gracefully", func(t *testing.T) {
		// Should not panic
		err := service.Send(nil, "user@example.com", "Subject", "Body")
		assert.NoError(t, err)
	})

	t.Run("handles cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := service.Send(ctx, "user@example.com", "Subject", "Body")
		assert.NoError(t, err)
	})
}

// =============================================================================
// Error Message Tests
// =============================================================================

func TestNoOpService_ErrorMessages(t *testing.T) {
	service := NewNoOpService("email service not available")

	t.Run("SendMagicLink error includes reason", func(t *testing.T) {
		err := service.SendMagicLink(context.Background(), "user@example.com", "token", "link")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email service not available")
	})

	t.Run("SendVerificationEmail error includes reason", func(t *testing.T) {
		err := service.SendVerificationEmail(context.Background(), "user@example.com", "token", "link")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email service not available")
	})

	t.Run("SendPasswordReset error includes reason", func(t *testing.T) {
		err := service.SendPasswordReset(context.Background(), "user@example.com", "token", "link")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email service not available")
	})

	t.Run("SendInvitationEmail error includes reason", func(t *testing.T) {
		err := service.SendInvitationEmail(context.Background(), "user@example.com", "Inviter", "link")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email service not available")
	})
}

// =============================================================================
// Empty/Whitespace Input Tests
// =============================================================================

func TestNoOpService_EmptyInputs(t *testing.T) {
	service := NewNoOpService("test")

	t.Run("handles empty email address", func(t *testing.T) {
		err := service.Send(context.Background(), "", "Subject", "Body")
		assert.Error(t, err)
	})

	t.Run("handles empty subject", func(t *testing.T) {
		err := service.Send(context.Background(), "user@example.com", "", "Body")
		assert.Error(t, err)
	})

	t.Run("handles empty body", func(t *testing.T) {
		err := service.Send(context.Background(), "user@example.com", "Subject", "")
		assert.Error(t, err)
	})
}

func TestTestEmailService_EmptyInputs(t *testing.T) {
	service := NewTestEmailService()

	t.Run("handles empty email address", func(t *testing.T) {
		err := service.Send(context.Background(), "", "Subject", "Body")
		assert.NoError(t, err)
	})

	t.Run("handles empty subject", func(t *testing.T) {
		err := service.Send(context.Background(), "user@example.com", "", "Body")
		assert.NoError(t, err)
	})

	t.Run("handles empty body", func(t *testing.T) {
		err := service.Send(context.Background(), "user@example.com", "Subject", "")
		assert.NoError(t, err)
	})
}
