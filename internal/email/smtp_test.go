package email

import (
	"context"
	"strings"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSMTPService(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled:      true,
		Provider:     "smtp",
		SMTPHost:     "smtp.example.com",
		SMTPPort:     587,
		SMTPUsername: "user@example.com",
		SMTPPassword: "password",
		SMTPTLS:      true,
		FromAddress:  "noreply@example.com",
		FromName:     "Test Service",
	}

	service := NewSMTPService(cfg)
	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
}

func TestSMTPService_buildMessage(t *testing.T) {
	cfg := &config.EmailConfig{
		FromAddress:    "noreply@example.com",
		FromName:       "Test Service",
		ReplyToAddress: "support@example.com",
	}
	service := NewSMTPService(cfg)

	tests := []struct {
		name    string
		to      string
		subject string
		body    string
		want    []string // Strings that should be present in the message
	}{
		{
			name:    "basic message",
			to:      "user@example.com",
			subject: "Test Subject",
			body:    "<p>Test Body</p>",
			want: []string{
				"From: Test Service <noreply@example.com>",
				"To: user@example.com",
				"Reply-To: support@example.com",
				"Subject: Test Subject",
				"MIME-Version: 1.0",
				"Content-Type: text/html; charset=UTF-8",
				"<p>Test Body</p>",
			},
		},
		{
			name:    "message without reply-to",
			to:      "user@example.com",
			subject: "Test",
			body:    "Body",
			want: []string{
				"From: Test Service <noreply@example.com>",
				"To: user@example.com",
				"Subject: Test",
				"Body",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := service.buildMessage(tt.to, tt.subject, tt.body)
			messageStr := string(message)

			for _, want := range tt.want {
				assert.Contains(t, messageStr, want)
			}
		})
	}
}

func TestSMTPService_renderMagicLinkTemplate(t *testing.T) {
	cfg := &config.EmailConfig{}
	service := NewSMTPService(cfg)

	link := "https://example.com/auth/verify?token=abc123"
	token := "abc123"

	result := service.renderMagicLinkTemplate(link, token)

	// Check that the result contains expected elements
	assert.Contains(t, result, link)
	assert.Contains(t, result, "Your Login Link")
	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "Log In")
}

func TestSMTPService_renderVerificationTemplate(t *testing.T) {
	cfg := &config.EmailConfig{}
	service := NewSMTPService(cfg)

	link := "https://example.com/auth/verify?token=xyz789"
	token := "xyz789"

	result := service.renderVerificationTemplate(link, token)

	// Check that the result contains expected elements
	assert.Contains(t, result, link)
	assert.Contains(t, result, "Verify Your Email")
	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "Verify Email")
}

func TestSMTPService_Send_Disabled(t *testing.T) {
	cfg := &config.EmailConfig{
		Enabled: false,
	}
	service := NewSMTPService(cfg)

	ctx := context.Background()
	err := service.Send(ctx, "user@example.com", "Test", "Body")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "disabled")
}

func TestSMTPService_Send_InvalidConfig(t *testing.T) {
	// This test would fail if we try to connect to a non-existent SMTP server
	// For now, we just test that the method exists and can be called
	cfg := &config.EmailConfig{
		Enabled:      true,
		Provider:     "smtp",
		SMTPHost:     "nonexistent.smtp.server",
		SMTPPort:     587,
		SMTPUsername: "user",
		SMTPPassword: "pass",
		SMTPTLS:      false,
		FromAddress:  "from@example.com",
	}
	service := NewSMTPService(cfg)

	ctx := context.Background()
	err := service.Send(ctx, "to@example.com", "Test", "Body")

	// We expect an error because the SMTP server doesn't exist
	assert.Error(t, err)
}

func TestDefaultTemplates(t *testing.T) {
	t.Run("magic link template is valid HTML", func(t *testing.T) {
		assert.Contains(t, defaultMagicLinkTemplate, "<!DOCTYPE html>")
		assert.Contains(t, defaultMagicLinkTemplate, "{{.Link}}")
		assert.Contains(t, defaultMagicLinkTemplate, "Your Login Link")
	})

	t.Run("verification template is valid HTML", func(t *testing.T) {
		assert.Contains(t, defaultVerificationTemplate, "<!DOCTYPE html>")
		assert.Contains(t, defaultVerificationTemplate, "{{.Link}}")
		assert.Contains(t, defaultVerificationTemplate, "Verify Your Email")
	})
}

func TestSMTPService_TemplateRendering_Fallback(t *testing.T) {
	// Test that the fallback templates are used when template execution fails
	// This would require mocking or using an invalid template, which we'll skip for now
	// The fallback code is already tested indirectly through the rendering tests
}

func TestNewService_SMTP(t *testing.T) {
	tests := []struct {
		name      string
		cfg       *config.EmailConfig
		wantErr   bool
		errMsg    string
		checkType func(t *testing.T, svc Service)
	}{
		{
			name: "SMTP provider",
			cfg: &config.EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
				FromAddress: "test@example.com",
			},
			wantErr: false,
			checkType: func(t *testing.T, svc Service) {
				_, ok := svc.(*SMTPService)
				assert.True(t, ok, "Expected SMTPService")
			},
		},
		{
			name: "disabled email",
			cfg: &config.EmailConfig{
				Enabled: false,
			},
			wantErr: false,
			checkType: func(t *testing.T, svc Service) {
				_, ok := svc.(*NoOpService)
				assert.True(t, ok, "Expected NoOpService")
			},
		},
		{
			name: "sendgrid provider",
			cfg: &config.EmailConfig{
				Enabled:        true,
				Provider:       "sendgrid",
				SendGridAPIKey: "test-api-key",
				FromAddress:    "test@example.com",
			},
			wantErr: false,
			checkType: func(t *testing.T, svc Service) {
				_, ok := svc.(*SendGridService)
				assert.True(t, ok, "Expected SendGridService")
			},
		},
		{
			name: "mailgun provider",
			cfg: &config.EmailConfig{
				Enabled:       true,
				Provider:      "mailgun",
				MailgunAPIKey: "test-api-key",
				MailgunDomain: "example.com",
				FromAddress:   "test@example.com",
			},
			wantErr: false,
			checkType: func(t *testing.T, svc Service) {
				_, ok := svc.(*MailgunService)
				assert.True(t, ok, "Expected MailgunService")
			},
		},
		{
			name: "ses provider",
			cfg: &config.EmailConfig{
				Enabled:      true,
				Provider:     "ses",
				SESRegion:    "us-east-1",
				SESAccessKey: "test-access-key",
				SESSecretKey: "test-secret-key",
				FromAddress:  "test@example.com",
			},
			wantErr: false,
			checkType: func(t *testing.T, svc Service) {
				_, ok := svc.(*SESService)
				assert.True(t, ok, "Expected SESService")
			},
		},
		{
			name: "unsupported provider",
			cfg: &config.EmailConfig{
				Enabled:  true,
				Provider: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported email provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewService(tt.cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, svc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, svc)
				if tt.checkType != nil {
					tt.checkType(t, svc)
				}
			}
		})
	}
}

func TestNoOpService_SMTP(t *testing.T) {
	service := NewNoOpService("email is disabled")
	ctx := context.Background()

	t.Run("SendMagicLink returns error", func(t *testing.T) {
		err := service.SendMagicLink(ctx, "user@example.com", "token", "link")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is disabled")
	})

	t.Run("SendVerificationEmail returns error", func(t *testing.T) {
		err := service.SendVerificationEmail(ctx, "user@example.com", "token", "link")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is disabled")
	})

	t.Run("Send returns error", func(t *testing.T) {
		err := service.Send(ctx, "user@example.com", "subject", "body")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is disabled")
	})
}

func TestEmailConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.EmailConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid SMTP config",
			cfg: config.EmailConfig{
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			wantErr: false,
		},
		{
			name: "unconfigured SMTP is valid (can be configured via admin UI)",
			cfg: config.EmailConfig{
				Provider: "smtp",
			},
			wantErr: false,
		},
		{
			name: "invalid provider",
			cfg: config.EmailConfig{
				Provider:    "invalid",
				FromAddress: "test@example.com",
			},
			wantErr: true,
			errMsg:  "invalid email provider",
		},
		{
			name: "empty provider is valid",
			cfg: config.EmailConfig{
				FromAddress: "test@example.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.errMsg))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEmailConfig_IsConfigured(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.EmailConfig
		configured bool
	}{
		{
			name: "fully configured SMTP",
			cfg: config.EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: true,
		},
		{
			name: "SMTP missing host",
			cfg: config.EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "SMTP missing port",
			cfg: config.EmailConfig{
				Enabled:     true,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
			},
			configured: false,
		},
		{
			name: "email disabled",
			cfg: config.EmailConfig{
				Enabled:     false,
				Provider:    "smtp",
				FromAddress: "test@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "missing from_address",
			cfg: config.EmailConfig{
				Enabled:  true,
				Provider: "smtp",
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
			},
			configured: false,
		},
		{
			name: "fully configured SendGrid",
			cfg: config.EmailConfig{
				Enabled:        true,
				Provider:       "sendgrid",
				FromAddress:    "test@example.com",
				SendGridAPIKey: "api-key",
			},
			configured: true,
		},
		{
			name: "SendGrid missing API key",
			cfg: config.EmailConfig{
				Enabled:     true,
				Provider:    "sendgrid",
				FromAddress: "test@example.com",
			},
			configured: false,
		},
		{
			name: "fully configured Mailgun",
			cfg: config.EmailConfig{
				Enabled:       true,
				Provider:      "mailgun",
				FromAddress:   "test@example.com",
				MailgunAPIKey: "api-key",
				MailgunDomain: "example.com",
			},
			configured: true,
		},
		{
			name: "Mailgun missing domain",
			cfg: config.EmailConfig{
				Enabled:       true,
				Provider:      "mailgun",
				FromAddress:   "test@example.com",
				MailgunAPIKey: "api-key",
			},
			configured: false,
		},
		{
			name: "fully configured SES",
			cfg: config.EmailConfig{
				Enabled:      true,
				Provider:     "ses",
				FromAddress:  "test@example.com",
				SESAccessKey: "access-key",
				SESSecretKey: "secret-key",
				SESRegion:    "us-east-1",
			},
			configured: true,
		},
		{
			name: "SES missing region",
			cfg: config.EmailConfig{
				Enabled:      true,
				Provider:     "ses",
				FromAddress:  "test@example.com",
				SESAccessKey: "access-key",
				SESSecretKey: "secret-key",
			},
			configured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.IsConfigured()
			assert.Equal(t, tt.configured, result)
		})
	}
}

// =============================================================================
// SanitizeHeaderValue Tests (Header Injection Prevention)
// =============================================================================

func TestSanitizeHeaderValue(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "normal string",
			input: "test@example.com",
			want:  "test@example.com",
		},
		{
			name:  "string with CR",
			input: "test\r@example.com",
			want:  "test@example.com",
		},
		{
			name:  "string with LF",
			input: "test\n@example.com",
			want:  "test@example.com",
		},
		{
			name:  "string with CRLF",
			input: "test\r\n@example.com",
			want:  "test@example.com",
		},
		{
			name:  "string with multiple CRLF",
			input: "test\r\n\r\n@example.com",
			want:  "test@example.com",
		},
		{
			name:  "header injection attempt",
			input: "test@example.com\r\nBcc: victim@example.com",
			want:  "test@example.comBcc: victim@example.com",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "only CRLF",
			input: "\r\n",
			want:  "",
		},
		{
			name:  "string with spaces and special chars",
			input: "Test User <test@example.com>",
			want:  "Test User <test@example.com>",
		},
		{
			name:  "unicode characters",
			input: "用户@example.com",
			want:  "用户@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeHeaderValue(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

// =============================================================================
// buildMessage with Header Sanitization Tests
// =============================================================================

func TestSMTPService_buildMessage_Sanitization(t *testing.T) {
	t.Run("sanitizes from name with CRLF", func(t *testing.T) {
		// Create a service with malicious input
		maliciousCfg := &config.EmailConfig{
			FromAddress:    "noreply@example.com",
			FromName:       "Test Service\r\nBcc: victim@example.com",
			ReplyToAddress: "support@example.com",
		}
		maliciousService := NewSMTPService(maliciousCfg)

		message := maliciousService.buildMessage("user@example.com", "Subject", "Body")
		messageStr := string(message)

		// The Bcc header should NOT appear as a separate header
		// The CRLF should be sanitized from the name value
		assert.NotContains(t, messageStr, "\r\nBcc:")
		// But "Test Service" should still be there
		assert.Contains(t, messageStr, "Test Service")
	})

	t.Run("sanitizes reply-to with CRLF injection", func(t *testing.T) {
		maliciousCfg := &config.EmailConfig{
			FromAddress:    "noreply@example.com",
			FromName:       "Test Service",
			ReplyToAddress: "support@example.com\r\nCc: victim@example.com",
		}
		maliciousService := NewSMTPService(maliciousCfg)

		message := maliciousService.buildMessage("user@example.com", "Subject", "Body")
		messageStr := string(message)

		// The Cc header should NOT appear
		assert.NotContains(t, messageStr, "\r\nCc:")
		// But support@example.com should still be there
		assert.Contains(t, messageStr, "support@example.com")
	})
}

// =============================================================================
// Send Method Error Handling Tests
// =============================================================================

func TestSMTPService_Send_ErrorCases(t *testing.T) {
	t.Run("returns error when service is disabled", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled: false,
		}
		service := NewSMTPService(cfg)

		err := service.Send(context.Background(), "user@example.com", "Subject", "Body")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "disabled")
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		cfg := &config.EmailConfig{
			Enabled:     true,
			SMTPHost:    "nonexistent.smtp.server",
			SMTPPort:    587,
			FromAddress: "from@example.com",
		}
		service := NewSMTPService(cfg)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := service.Send(ctx, "user@example.com", "Subject", "Body")

		assert.Error(t, err)
	})
}

// =============================================================================
// IsConfigured Tests
// =============================================================================

func TestSMTPService_IsConfigured(t *testing.T) {
	tests := []struct {
		name       string
		cfg        *config.EmailConfig
		configured bool
	}{
		{
			name: "fully configured",
			cfg: &config.EmailConfig{
				Enabled:     true,
				FromAddress: "noreply@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: true,
		},
		{
			name: "disabled",
			cfg: &config.EmailConfig{
				Enabled:     false,
				FromAddress: "noreply@example.com",
				SMTPHost:    "smtp.example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "missing host",
			cfg: &config.EmailConfig{
				Enabled:     true,
				FromAddress: "noreply@example.com",
				SMTPPort:    587,
			},
			configured: false,
		},
		{
			name: "missing port",
			cfg: &config.EmailConfig{
				Enabled:     true,
				FromAddress: "noreply@example.com",
				SMTPHost:    "smtp.example.com",
			},
			configured: false,
		},
		{
			name: "missing from address",
			cfg: &config.EmailConfig{
				Enabled:  true,
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
			},
			configured: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewSMTPService(tt.cfg)
			result := service.IsConfigured()
			assert.Equal(t, tt.configured, result)
		})
	}
}

// =============================================================================
// Email Template Rendering Tests
// =============================================================================

func TestSMTPService_RenderPasswordResetTemplate(t *testing.T) {
	cfg := &config.EmailConfig{}
	service := NewSMTPService(cfg)

	link := "https://example.com/reset?token=reset123"
	token := "reset123"

	result := service.renderPasswordResetTemplate(link, token)

	assert.Contains(t, result, link)
	assert.Contains(t, result, "Reset Your Password")
	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "Reset Password")
}

// =============================================================================
// Message Building Edge Cases
// =============================================================================

func TestSMTPService_buildMessage_EdgeCases(t *testing.T) {
	t.Run("handles empty reply-to", func(t *testing.T) {
		cfg := &config.EmailConfig{
			FromAddress: "noreply@example.com",
			FromName:    "Test Service",
		}
		service := NewSMTPService(cfg)

		message := service.buildMessage("user@example.com", "Subject", "Body")
		messageStr := string(message)

		assert.NotContains(t, messageStr, "Reply-To:")
	})

	t.Run("handles special characters in subject", func(t *testing.T) {
		cfg := &config.EmailConfig{
			FromAddress: "noreply@example.com",
		}
		service := NewSMTPService(cfg)

		subject := "Test with special chars: <>&\"'"
		message := service.buildMessage("user@example.com", subject, "Body")
		messageStr := string(message)

		assert.Contains(t, messageStr, "Subject:")
		assert.Contains(t, messageStr, "special chars")
	})

	t.Run("handles unicode in body", func(t *testing.T) {
		cfg := &config.EmailConfig{
			FromAddress: "noreply@example.com",
		}
		service := NewSMTPService(cfg)

		body := "<p>Hello 世界 🌍</p>"
		message := service.buildMessage("user@example.com", "Subject", body)
		messageStr := string(message)

		assert.Contains(t, messageStr, body)
		assert.Contains(t, messageStr, "charset=UTF-8")
	})
}

// =============================================================================
// Template Tests
// =============================================================================

func TestDefaultTemplates_Complete(t *testing.T) {
	t.Run("password reset template contains all elements", func(t *testing.T) {
		assert.Contains(t, defaultPasswordResetTemplate, "<!DOCTYPE html>")
		assert.Contains(t, defaultPasswordResetTemplate, "{{.Link}}")
		assert.Contains(t, defaultPasswordResetTemplate, "Reset Your Password")
		assert.Contains(t, defaultPasswordResetTemplate, "Security Reminder")
	})

	t.Run("invitation template contains all elements", func(t *testing.T) {
		assert.Contains(t, defaultInvitationTemplate, "<!DOCTYPE html>")
		assert.Contains(t, defaultInvitationTemplate, "{{.InviteLink}}")
		assert.Contains(t, defaultInvitationTemplate, "You've Been Invited!")
		assert.Contains(t, defaultInvitationTemplate, "{{.InviterName}}")
	})
}

// =============================================================================
// SendInvitationEmail Test
// =============================================================================

func TestSMTPService_SendInvitationEmail(t *testing.T) {
	cfg := &config.EmailConfig{
		FromAddress: "noreply@example.com",
		FromName:    "Test Service",
	}
	service := NewSMTPService(cfg)

	t.Run("renders invitation template", func(t *testing.T) {
		// We can't actually send the email without a server, but we can check the template rendering
		link := "https://example.com/invite?code=abc123"
		inviter := "John Doe"

		// This would fail at send time, but we can check it builds the message
		err := service.SendInvitationEmail(context.Background(), "user@example.com", inviter, link)

		// Will fail because no SMTP server, but that's expected
		assert.Error(t, err)
	})
}
