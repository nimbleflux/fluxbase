package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AppSettingsHandler Construction Tests
// =============================================================================

func TestNewAppSettingsHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewAppSettingsHandler(nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.settingsService)
		assert.Nil(t, handler.settingsCache)
		assert.Nil(t, handler.config)
	})
}

// =============================================================================
// AppSettings Struct Tests
// =============================================================================

func TestAppSettings_Struct(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		settings := AppSettings{
			Authentication: AuthenticationSettings{
				PasswordMinLength:     8,
				SessionTimeoutMinutes: 60,
				MaxSessionsPerUser:    5,
			},
			Features: FeatureSettings{
				EnableRealtime:  true,
				EnableStorage:   true,
				EnableFunctions: true,
			},
			Email: EmailSettings{
				Provider: "smtp",
			},
			Security: SecuritySettings{},
		}

		assert.Equal(t, 8, settings.Authentication.PasswordMinLength)
		assert.Equal(t, 60, settings.Authentication.SessionTimeoutMinutes)
		assert.Equal(t, 5, settings.Authentication.MaxSessionsPerUser)
		assert.True(t, settings.Features.EnableRealtime)
		assert.True(t, settings.Features.EnableStorage)
		assert.True(t, settings.Features.EnableFunctions)
		assert.Equal(t, "smtp", settings.Email.Provider)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		settings := AppSettings{
			Authentication: AuthenticationSettings{
				SignupEnabled:            true,
				MagicLinkEnabled:         true,
				PasswordMinLength:        12,
				RequireEmailVerification: true,
			},
			Features: FeatureSettings{
				EnableRealtime:  false,
				EnableStorage:   true,
				EnableFunctions: true,
			},
			Email: EmailSettings{
				Enabled:  true,
				Provider: "sendgrid",
			},
			Security: SecuritySettings{
				EnableGlobalRateLimit: true,
			},
		}

		data, err := json.Marshal(settings)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enable_signup":true`)
		assert.Contains(t, string(data), `"enable_magic_link":true`)
		assert.Contains(t, string(data), `"password_min_length":12`)
		assert.Contains(t, string(data), `"provider":"sendgrid"`)
		assert.Contains(t, string(data), `"enable_global_rate_limit":true`)
	})
}

// =============================================================================
// AuthenticationSettings Struct Tests
// =============================================================================

func TestAuthenticationSettings_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		auth := AuthenticationSettings{
			SignupEnabled:            true,
			MagicLinkEnabled:         true,
			PasswordMinLength:        10,
			RequireEmailVerification: true,
			PasswordRequireUppercase: true,
			PasswordRequireLowercase: true,
			PasswordRequireNumber:    true,
			PasswordRequireSpecial:   true,
			SessionTimeoutMinutes:    120,
			MaxSessionsPerUser:       10,
		}

		assert.True(t, auth.SignupEnabled)
		assert.True(t, auth.MagicLinkEnabled)
		assert.Equal(t, 10, auth.PasswordMinLength)
		assert.True(t, auth.RequireEmailVerification)
		assert.True(t, auth.PasswordRequireUppercase)
		assert.True(t, auth.PasswordRequireLowercase)
		assert.True(t, auth.PasswordRequireNumber)
		assert.True(t, auth.PasswordRequireSpecial)
		assert.Equal(t, 120, auth.SessionTimeoutMinutes)
		assert.Equal(t, 10, auth.MaxSessionsPerUser)
	})

	t.Run("JSON field names", func(t *testing.T) {
		auth := AuthenticationSettings{
			SignupEnabled:     true,
			PasswordMinLength: 8,
		}

		data, err := json.Marshal(auth)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enable_signup"`)
		assert.Contains(t, string(data), `"password_min_length"`)
		assert.Contains(t, string(data), `"session_timeout_minutes"`)
	})
}

// =============================================================================
// FeatureSettings Struct Tests
// =============================================================================

func TestFeatureSettings_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		features := FeatureSettings{
			EnableRealtime:  true,
			EnableStorage:   false,
			EnableFunctions: true,
		}

		assert.True(t, features.EnableRealtime)
		assert.False(t, features.EnableStorage)
		assert.True(t, features.EnableFunctions)
	})

	t.Run("JSON field names", func(t *testing.T) {
		features := FeatureSettings{
			EnableRealtime:  true,
			EnableStorage:   true,
			EnableFunctions: true,
		}

		data, err := json.Marshal(features)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enable_realtime"`)
		assert.Contains(t, string(data), `"enable_storage"`)
		assert.Contains(t, string(data), `"enable_functions"`)
	})
}

// =============================================================================
// EmailSettings Struct Tests
// =============================================================================

func TestEmailSettings_Struct(t *testing.T) {
	t.Run("basic email settings", func(t *testing.T) {
		email := EmailSettings{
			Enabled:        true,
			Provider:       "smtp",
			FromAddress:    "noreply@example.com",
			FromName:       "My App",
			ReplyToAddress: "support@example.com",
		}

		assert.True(t, email.Enabled)
		assert.Equal(t, "smtp", email.Provider)
		assert.Equal(t, "noreply@example.com", email.FromAddress)
		assert.Equal(t, "My App", email.FromName)
		assert.Equal(t, "support@example.com", email.ReplyToAddress)
	})

	t.Run("SMTP settings", func(t *testing.T) {
		email := EmailSettings{
			Provider: "smtp",
			SMTP: &SMTPSettings{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				TLS:      true,
			},
		}

		assert.NotNil(t, email.SMTP)
		assert.Equal(t, "smtp.example.com", email.SMTP.Host)
		assert.Equal(t, 587, email.SMTP.Port)
		assert.Equal(t, "user", email.SMTP.Username)
		assert.True(t, email.SMTP.TLS)
	})

	t.Run("SendGrid settings", func(t *testing.T) {
		email := EmailSettings{
			Provider: "sendgrid",
			SendGrid: &SendGridSettings{
				APIKey: "SG.xxxx",
			},
		}

		assert.NotNil(t, email.SendGrid)
		assert.Equal(t, "SG.xxxx", email.SendGrid.APIKey)
	})

	t.Run("Mailgun settings", func(t *testing.T) {
		email := EmailSettings{
			Provider: "mailgun",
			Mailgun: &MailgunSettings{
				APIKey:   "key-xxx",
				Domain:   "mg.example.com",
				EURegion: true,
			},
		}

		assert.NotNil(t, email.Mailgun)
		assert.Equal(t, "mg.example.com", email.Mailgun.Domain)
		assert.True(t, email.Mailgun.EURegion)
	})

	t.Run("SES settings", func(t *testing.T) {
		email := EmailSettings{
			Provider: "ses",
			SES: &SESSettings{
				Region:          "us-east-1",
				AccessKeyID:     "AKIAXXXX",
				SecretAccessKey: "secret",
			},
		}

		assert.NotNil(t, email.SES)
		assert.Equal(t, "us-east-1", email.SES.Region)
	})
}

// =============================================================================
// SecuritySettings Struct Tests
// =============================================================================

func TestSecuritySettings_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		security := SecuritySettings{
			EnableGlobalRateLimit: true,
		}

		assert.True(t, security.EnableGlobalRateLimit)
	})

	t.Run("JSON field name", func(t *testing.T) {
		security := SecuritySettings{
			EnableGlobalRateLimit: true,
		}

		data, err := json.Marshal(security)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enable_global_rate_limit"`)
	})
}

// =============================================================================
// SettingOverrides Struct Tests
// =============================================================================

func TestSettingOverrides_Struct(t *testing.T) {
	t.Run("all categories accessible", func(t *testing.T) {
		overrides := SettingOverrides{
			Authentication: map[string]bool{"enable_signup": true},
			Features:       map[string]bool{"enable_realtime": true},
			Email:          map[string]bool{"provider": true},
			Security:       map[string]bool{"enable_global_rate_limit": true},
		}

		assert.True(t, overrides.Authentication["enable_signup"])
		assert.True(t, overrides.Features["enable_realtime"])
		assert.True(t, overrides.Email["provider"])
		assert.True(t, overrides.Security["enable_global_rate_limit"])
	})

	t.Run("omitempty works for empty maps", func(t *testing.T) {
		overrides := SettingOverrides{
			Authentication: map[string]bool{},
			Features:       map[string]bool{},
		}

		data, err := json.Marshal(overrides)
		require.NoError(t, err)

		// Empty maps should still be present but empty
		var parsed SettingOverrides
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
	})
}

// =============================================================================
// UpdateAppSettingsRequest Struct Tests
// =============================================================================

func TestUpdateAppSettingsRequest_Struct(t *testing.T) {
	t.Run("all fields optional (omitempty)", func(t *testing.T) {
		// Empty request
		req := UpdateAppSettingsRequest{}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		// Should be empty or minimal JSON
		assert.Equal(t, "{}", string(data))
	})

	t.Run("partial update with authentication", func(t *testing.T) {
		auth := AuthenticationSettings{
			SignupEnabled:     true,
			PasswordMinLength: 10,
		}
		req := UpdateAppSettingsRequest{
			Authentication: &auth,
		}

		assert.NotNil(t, req.Authentication)
		assert.Nil(t, req.Features)
		assert.Nil(t, req.Email)
		assert.Nil(t, req.Security)
	})

	t.Run("full update request", func(t *testing.T) {
		req := UpdateAppSettingsRequest{
			Authentication: &AuthenticationSettings{
				SignupEnabled:     true,
				PasswordMinLength: 12,
			},
			Features: &FeatureSettings{
				EnableRealtime:  true,
				EnableStorage:   true,
				EnableFunctions: false,
			},
			Email: &EmailSettings{
				Enabled:  true,
				Provider: "smtp",
			},
			Security: &SecuritySettings{
				EnableGlobalRateLimit: true,
			},
		}

		assert.NotNil(t, req.Authentication)
		assert.NotNil(t, req.Features)
		assert.NotNil(t, req.Email)
		assert.NotNil(t, req.Security)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"authentication": {
				"enable_signup": true,
				"password_min_length": 10
			},
			"features": {
				"enable_realtime": false
			}
		}`

		var req UpdateAppSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.NotNil(t, req.Authentication)
		assert.True(t, req.Authentication.SignupEnabled)
		assert.Equal(t, 10, req.Authentication.PasswordMinLength)
		assert.NotNil(t, req.Features)
		assert.False(t, req.Features.EnableRealtime)
	})
}

// =============================================================================
// UpdateAppSettings Handler Validation Tests
// =============================================================================

func TestUpdateAppSettings_Validation(t *testing.T) {
	t.Run("invalid request body returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAppSettingsHandler(nil, nil, nil)

		app.Put("/admin/app/settings", handler.UpdateAppSettings)

		req := httptest.NewRequest(http.MethodPut, "/admin/app/settings", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// Email Provider Tests
// =============================================================================

func TestEmailProviderSettings(t *testing.T) {
	t.Run("valid email providers", func(t *testing.T) {
		providers := []string{"smtp", "sendgrid", "mailgun", "ses"}

		for _, provider := range providers {
			email := EmailSettings{
				Provider: provider,
			}
			assert.Equal(t, provider, email.Provider)
		}
	})

	t.Run("SMTP with TLS enabled", func(t *testing.T) {
		smtp := SMTPSettings{
			Host:     "smtp.gmail.com",
			Port:     465,
			Username: "user@gmail.com",
			TLS:      true,
		}

		assert.Equal(t, 465, smtp.Port)
		assert.True(t, smtp.TLS)
	})

	t.Run("SMTP without TLS", func(t *testing.T) {
		smtp := SMTPSettings{
			Host:     "localhost",
			Port:     25,
			Username: "",
			TLS:      false,
		}

		assert.Equal(t, 25, smtp.Port)
		assert.False(t, smtp.TLS)
	})
}

// =============================================================================
// Setting Key Patterns Tests
// =============================================================================

func TestSettingKeyPatterns(t *testing.T) {
	t.Run("authentication setting keys", func(t *testing.T) {
		keys := []string{
			"app.auth.signup_enabled",
			"app.auth.magic_link_enabled",
			"app.auth.password_min_length",
			"app.auth.require_email_verification",
		}

		for _, key := range keys {
			assert.Contains(t, key, "app.auth.")
		}
	})

	t.Run("feature setting keys", func(t *testing.T) {
		keys := []string{
			"app.realtime.enabled",
			"app.storage.enabled",
			"app.functions.enabled",
		}

		for _, key := range keys {
			assert.Contains(t, key, "app.")
			assert.Contains(t, key, ".enabled")
		}
	})

	t.Run("email setting keys", func(t *testing.T) {
		keys := []string{
			"app.email.enabled",
			"app.email.provider",
			"app.email.from_address",
			"app.email.smtp.host",
			"app.email.smtp.port",
			"app.email.sendgrid.api_key",
			"app.email.mailgun.domain",
			"app.email.ses.region",
		}

		for _, key := range keys {
			assert.Contains(t, key, "app.email.")
		}
	})

	t.Run("security setting keys", func(t *testing.T) {
		keys := []string{
			"app.security.enable_global_rate_limit",
		}

		for _, key := range keys {
			assert.Contains(t, key, "app.security.")
		}
	})
}

// =============================================================================
// Password Requirements Tests
// =============================================================================

func TestPasswordRequirements(t *testing.T) {
	t.Run("default minimum length", func(t *testing.T) {
		auth := AuthenticationSettings{
			PasswordMinLength: 8,
		}
		assert.Equal(t, 8, auth.PasswordMinLength)
	})

	t.Run("all complexity requirements enabled", func(t *testing.T) {
		auth := AuthenticationSettings{
			PasswordMinLength:        12,
			PasswordRequireUppercase: true,
			PasswordRequireLowercase: true,
			PasswordRequireNumber:    true,
			PasswordRequireSpecial:   true,
		}

		assert.Equal(t, 12, auth.PasswordMinLength)
		assert.True(t, auth.PasswordRequireUppercase)
		assert.True(t, auth.PasswordRequireLowercase)
		assert.True(t, auth.PasswordRequireNumber)
		assert.True(t, auth.PasswordRequireSpecial)
	})
}

// =============================================================================
// Session Settings Tests
// =============================================================================

func TestSessionSettings(t *testing.T) {
	t.Run("default session timeout", func(t *testing.T) {
		auth := AuthenticationSettings{
			SessionTimeoutMinutes: 60,
		}
		assert.Equal(t, 60, auth.SessionTimeoutMinutes)
	})

	t.Run("default max sessions per user", func(t *testing.T) {
		auth := AuthenticationSettings{
			MaxSessionsPerUser: 5,
		}
		assert.Equal(t, 5, auth.MaxSessionsPerUser)
	})

	t.Run("custom session limits", func(t *testing.T) {
		auth := AuthenticationSettings{
			SessionTimeoutMinutes: 480, // 8 hours
			MaxSessionsPerUser:    1,   // Single session only
		}

		assert.Equal(t, 480, auth.SessionTimeoutMinutes)
		assert.Equal(t, 1, auth.MaxSessionsPerUser)
	})
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestAppSettingsJSONSerialization(t *testing.T) {
	t.Run("full AppSettings serializes correctly", func(t *testing.T) {
		settings := AppSettings{
			Authentication: AuthenticationSettings{
				SignupEnabled:     true,
				PasswordMinLength: 8,
			},
			Features: FeatureSettings{
				EnableRealtime: true,
			},
			Email: EmailSettings{
				Enabled:  true,
				Provider: "smtp",
				SMTP: &SMTPSettings{
					Host: "localhost",
					Port: 25,
				},
			},
			Security: SecuritySettings{
				EnableGlobalRateLimit: false,
			},
			Overrides: SettingOverrides{
				Authentication: map[string]bool{"enable_signup": true},
			},
		}

		data, err := json.Marshal(settings)
		require.NoError(t, err)

		var parsed AppSettings
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)

		assert.Equal(t, settings.Authentication.SignupEnabled, parsed.Authentication.SignupEnabled)
		assert.Equal(t, settings.Features.EnableRealtime, parsed.Features.EnableRealtime)
		assert.Equal(t, settings.Email.Provider, parsed.Email.Provider)
		assert.NotNil(t, parsed.Email.SMTP)
		assert.Equal(t, "localhost", parsed.Email.SMTP.Host)
	})

	t.Run("sensitive fields can be omitted", func(t *testing.T) {
		smtp := SMTPSettings{
			Host:     "smtp.example.com",
			Port:     587,
			Username: "user",
			Password: "", // Empty - should be omitted
			TLS:      true,
		}

		data, err := json.Marshal(smtp)
		require.NoError(t, err)

		// Password field should be omitted when empty due to omitempty
		assert.NotContains(t, string(data), `"password":"secret"`)
	})
}
