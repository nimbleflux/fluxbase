package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// EmailSettingsHandler Construction Tests
// =============================================================================

func TestNewEmailSettingsHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.settingsService)
		assert.Nil(t, handler.settingsCache)
		assert.Nil(t, handler.emailManager)
		assert.Nil(t, handler.secretsService)
		assert.Nil(t, handler.config)
		assert.Nil(t, handler.unifiedService)
	})
}

// =============================================================================
// EmailSettingsResponse Struct Tests
// =============================================================================

func TestEmailSettingsResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:           true,
			Provider:          "smtp",
			FromAddress:       "noreply@example.com",
			FromName:          "Example App",
			SMTPHost:          "smtp.example.com",
			SMTPPort:          587,
			SMTPUsername:      "user@example.com",
			SMTPPasswordSet:   true,
			SMTPTLS:           true,
			SendGridAPIKeySet: false,
			MailgunAPIKeySet:  false,
			MailgunDomain:     "",
			SESAccessKeySet:   false,
			SESSecretKeySet:   false,
			SESRegion:         "us-east-1",
			Overrides: map[string]OverrideInfo{
				"enabled": {IsOverridden: true, EnvVar: "FLUXBASE_EMAIL_ENABLED"},
			},
		}

		assert.True(t, resp.Enabled)
		assert.Equal(t, "smtp", resp.Provider)
		assert.Equal(t, "noreply@example.com", resp.FromAddress)
		assert.Equal(t, "Example App", resp.FromName)
		assert.Equal(t, "smtp.example.com", resp.SMTPHost)
		assert.Equal(t, 587, resp.SMTPPort)
		assert.Equal(t, "user@example.com", resp.SMTPUsername)
		assert.True(t, resp.SMTPPasswordSet)
		assert.True(t, resp.SMTPTLS)
		assert.False(t, resp.SendGridAPIKeySet)
		assert.False(t, resp.MailgunAPIKeySet)
		assert.Empty(t, resp.MailgunDomain)
		assert.False(t, resp.SESAccessKeySet)
		assert.False(t, resp.SESSecretKeySet)
		assert.Equal(t, "us-east-1", resp.SESRegion)
		assert.Len(t, resp.Overrides, 1)
		assert.True(t, resp.Overrides["enabled"].IsOverridden)
	})

	t.Run("SMTP provider configuration", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:         true,
			Provider:        "smtp",
			FromAddress:     "noreply@myapp.com",
			SMTPHost:        "mail.myapp.com",
			SMTPPort:        465,
			SMTPUsername:    "smtp-user",
			SMTPPasswordSet: true,
			SMTPTLS:         true,
			Overrides:       make(map[string]OverrideInfo),
		}

		assert.Equal(t, "smtp", resp.Provider)
		assert.Equal(t, 465, resp.SMTPPort)
		assert.True(t, resp.SMTPPasswordSet)
	})

	t.Run("SendGrid provider configuration", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:           true,
			Provider:          "sendgrid",
			FromAddress:       "noreply@myapp.com",
			SendGridAPIKeySet: true,
			Overrides:         make(map[string]OverrideInfo),
		}

		assert.Equal(t, "sendgrid", resp.Provider)
		assert.True(t, resp.SendGridAPIKeySet)
	})

	t.Run("Mailgun provider configuration", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:          true,
			Provider:         "mailgun",
			FromAddress:      "noreply@mail.myapp.com",
			MailgunAPIKeySet: true,
			MailgunDomain:    "mail.myapp.com",
			Overrides:        make(map[string]OverrideInfo),
		}

		assert.Equal(t, "mailgun", resp.Provider)
		assert.True(t, resp.MailgunAPIKeySet)
		assert.Equal(t, "mail.myapp.com", resp.MailgunDomain)
	})

	t.Run("AWS SES provider configuration", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:         true,
			Provider:        "ses",
			FromAddress:     "noreply@myapp.com",
			SESAccessKeySet: true,
			SESSecretKeySet: true,
			SESRegion:       "eu-west-1",
			Overrides:       make(map[string]OverrideInfo),
		}

		assert.Equal(t, "ses", resp.Provider)
		assert.True(t, resp.SESAccessKeySet)
		assert.True(t, resp.SESSecretKeySet)
		assert.Equal(t, "eu-west-1", resp.SESRegion)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := EmailSettingsResponse{
			Enabled:         true,
			Provider:        "smtp",
			FromAddress:     "test@example.com",
			SMTPHost:        "smtp.test.com",
			SMTPPort:        587,
			SMTPPasswordSet: true,
			Overrides:       make(map[string]OverrideInfo),
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"provider":"smtp"`)
		assert.Contains(t, string(data), `"from_address":"test@example.com"`)
		assert.Contains(t, string(data), `"smtp_host":"smtp.test.com"`)
		assert.Contains(t, string(data), `"smtp_port":587`)
		assert.Contains(t, string(data), `"smtp_password_set":true`)
	})

	t.Run("sensitive fields not exposed", func(t *testing.T) {
		resp := EmailSettingsResponse{
			SMTPPasswordSet:   true,
			SendGridAPIKeySet: true,
			MailgunAPIKeySet:  true,
			SESAccessKeySet:   true,
			SESSecretKeySet:   true,
			Overrides:         make(map[string]OverrideInfo),
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		// Should contain boolean flags, not actual secrets
		assert.Contains(t, string(data), `"smtp_password_set":true`)
		assert.Contains(t, string(data), `"sendgrid_api_key_set":true`)
		assert.Contains(t, string(data), `"mailgun_api_key_set":true`)
		assert.Contains(t, string(data), `"ses_access_key_set":true`)
		assert.Contains(t, string(data), `"ses_secret_key_set":true`)

		// Should NOT contain actual secret values
		assert.NotContains(t, string(data), `"smtp_password":`)
		assert.NotContains(t, string(data), `"sendgrid_api_key":`)
		assert.NotContains(t, string(data), `"mailgun_api_key":`)
		assert.NotContains(t, string(data), `"ses_access_key":`)
		assert.NotContains(t, string(data), `"ses_secret_key":`)
	})
}

// =============================================================================
// OverrideInfo Struct Tests
// =============================================================================

func TestOverrideInfo_Struct(t *testing.T) {
	t.Run("overridden setting", func(t *testing.T) {
		info := OverrideInfo{
			IsOverridden: true,
			EnvVar:       "FLUXBASE_EMAIL_ENABLED",
		}

		assert.True(t, info.IsOverridden)
		assert.Equal(t, "FLUXBASE_EMAIL_ENABLED", info.EnvVar)
	})

	t.Run("non-overridden setting", func(t *testing.T) {
		info := OverrideInfo{
			IsOverridden: false,
			EnvVar:       "",
		}

		assert.False(t, info.IsOverridden)
		assert.Empty(t, info.EnvVar)
	})

	t.Run("JSON serialization - overridden", func(t *testing.T) {
		info := OverrideInfo{
			IsOverridden: true,
			EnvVar:       "MY_ENV_VAR",
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"is_overridden":true`)
		assert.Contains(t, string(data), `"env_var":"MY_ENV_VAR"`)
	})

	t.Run("JSON serialization - not overridden omits empty env_var", func(t *testing.T) {
		info := OverrideInfo{
			IsOverridden: false,
			EnvVar:       "",
		}

		data, err := json.Marshal(info)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"is_overridden":false`)
		// env_var should be omitted when empty
		assert.NotContains(t, string(data), `"env_var"`)
	})
}

// =============================================================================
// UpdateEmailSettingsRequest Struct Tests
// =============================================================================

func TestUpdateEmailSettingsRequest_Struct(t *testing.T) {
	t.Run("full update request", func(t *testing.T) {
		enabled := true
		provider := "smtp"
		fromAddr := "noreply@test.com"
		fromName := "Test App"
		smtpHost := "smtp.test.com"
		smtpPort := 587
		smtpUser := "user"
		smtpPass := "password123"
		smtpTLS := true

		req := UpdateEmailSettingsRequest{
			Enabled:      &enabled,
			Provider:     &provider,
			FromAddress:  &fromAddr,
			FromName:     &fromName,
			SMTPHost:     &smtpHost,
			SMTPPort:     &smtpPort,
			SMTPUsername: &smtpUser,
			SMTPPassword: &smtpPass,
			SMTPTLS:      &smtpTLS,
		}

		assert.True(t, *req.Enabled)
		assert.Equal(t, "smtp", *req.Provider)
		assert.Equal(t, "noreply@test.com", *req.FromAddress)
		assert.Equal(t, "Test App", *req.FromName)
		assert.Equal(t, "smtp.test.com", *req.SMTPHost)
		assert.Equal(t, 587, *req.SMTPPort)
		assert.Equal(t, "user", *req.SMTPUsername)
		assert.Equal(t, "password123", *req.SMTPPassword)
		assert.True(t, *req.SMTPTLS)
	})

	t.Run("partial update - provider only", func(t *testing.T) {
		provider := "sendgrid"
		req := UpdateEmailSettingsRequest{
			Provider: &provider,
		}

		assert.Nil(t, req.Enabled)
		assert.Equal(t, "sendgrid", *req.Provider)
		assert.Nil(t, req.FromAddress)
	})

	t.Run("JSON deserialization - SMTP settings", func(t *testing.T) {
		jsonData := `{
			"enabled": true,
			"provider": "smtp",
			"smtp_host": "mail.example.com",
			"smtp_port": 465,
			"smtp_tls": true
		}`

		var req UpdateEmailSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.True(t, *req.Enabled)
		assert.Equal(t, "smtp", *req.Provider)
		assert.Equal(t, "mail.example.com", *req.SMTPHost)
		assert.Equal(t, 465, *req.SMTPPort)
		assert.True(t, *req.SMTPTLS)
	})

	t.Run("JSON deserialization - SendGrid settings", func(t *testing.T) {
		jsonData := `{
			"provider": "sendgrid",
			"sendgrid_api_key": "SG.xxx"
		}`

		var req UpdateEmailSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "sendgrid", *req.Provider)
		assert.Equal(t, "SG.xxx", *req.SendGridAPIKey)
	})

	t.Run("JSON deserialization - Mailgun settings", func(t *testing.T) {
		jsonData := `{
			"provider": "mailgun",
			"mailgun_api_key": "key-xxx",
			"mailgun_domain": "mg.example.com"
		}`

		var req UpdateEmailSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "mailgun", *req.Provider)
		assert.Equal(t, "key-xxx", *req.MailgunAPIKey)
		assert.Equal(t, "mg.example.com", *req.MailgunDomain)
	})

	t.Run("JSON deserialization - AWS SES settings", func(t *testing.T) {
		jsonData := `{
			"provider": "ses",
			"ses_access_key": "AKIAIOSFODNN7EXAMPLE",
			"ses_secret_key": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"ses_region": "eu-west-1"
		}`

		var req UpdateEmailSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "ses", *req.Provider)
		assert.Equal(t, "AKIAIOSFODNN7EXAMPLE", *req.SESAccessKey)
		assert.Equal(t, "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY", *req.SESSecretKey)
		assert.Equal(t, "eu-west-1", *req.SESRegion)
	})
}

// =============================================================================
// TestEmailSettingsRequest Struct Tests
// =============================================================================

func TestTestEmailSettingsRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := TestEmailSettingsRequest{
			RecipientEmail: "test@example.com",
		}

		assert.Equal(t, "test@example.com", req.RecipientEmail)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"recipient_email":"admin@mycompany.com"}`

		var req TestEmailSettingsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "admin@mycompany.com", req.RecipientEmail)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := TestEmailSettingsRequest{
			RecipientEmail: "user@test.com",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"recipient_email":"user@test.com"`)
	})
}

// =============================================================================
// UpdateSettings Handler Tests
// =============================================================================

func TestUpdateSettings_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Put("/email/settings", handler.UpdateSettings)

		req := httptest.NewRequest(http.MethodPut, "/email/settings", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid request body")
	})

	t.Run("empty body is valid (no updates)", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Put("/email/settings", handler.UpdateSettings)

		body := `{}`
		req := httptest.NewRequest(http.MethodPut, "/email/settings", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Empty body is parsed successfully, fails at settings operation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// TestSettings Handler Tests
// =============================================================================

func TestTestSettings_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Post("/email/settings/test", handler.TestSettings)

		req := httptest.NewRequest(http.MethodPost, "/email/settings/test", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid request body")
	})

	t.Run("missing recipient email", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Post("/email/settings/test", handler.TestSettings)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/email/settings/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Recipient email is required")
	})

	t.Run("empty recipient email", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Post("/email/settings/test", handler.TestSettings)

		body := `{"recipient_email": ""}`
		req := httptest.NewRequest(http.MethodPost, "/email/settings/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Recipient email is required")
	})

	t.Run("email manager not initialized", func(t *testing.T) {
		app := fiber.New()
		handler := NewEmailSettingsHandler(nil, nil, nil, nil, nil, nil)

		app.Post("/email/settings/test", handler.TestSettings)

		body := `{"recipient_email": "test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/email/settings/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Email service not initialized")
	})
}

// =============================================================================
// Provider Types Tests
// =============================================================================

func TestEmailProviderTypes(t *testing.T) {
	validProviders := []string{"smtp", "sendgrid", "mailgun", "ses"}

	t.Run("valid provider types", func(t *testing.T) {
		for _, provider := range validProviders {
			resp := EmailSettingsResponse{
				Provider:  provider,
				Overrides: make(map[string]OverrideInfo),
			}
			assert.Equal(t, provider, resp.Provider)
		}
	})
}

// =============================================================================
// SMTP Port Tests
// =============================================================================

func TestSMTPPortValues(t *testing.T) {
	commonPorts := []int{25, 465, 587, 2525}

	t.Run("common SMTP ports", func(t *testing.T) {
		for _, port := range commonPorts {
			resp := EmailSettingsResponse{
				SMTPPort:  port,
				Overrides: make(map[string]OverrideInfo),
			}
			assert.Equal(t, port, resp.SMTPPort)
		}
	})

	t.Run("update SMTP port", func(t *testing.T) {
		port := 2525
		req := UpdateEmailSettingsRequest{
			SMTPPort: &port,
		}
		assert.Equal(t, 2525, *req.SMTPPort)
	})
}

// =============================================================================
// AWS SES Regions Tests
// =============================================================================

func TestAWSSESRegions(t *testing.T) {
	validRegions := []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"eu-central-1",
		"ap-southeast-1",
		"ap-northeast-1",
	}

	t.Run("valid AWS SES regions", func(t *testing.T) {
		for _, region := range validRegions {
			resp := EmailSettingsResponse{
				Provider:  "ses",
				SESRegion: region,
				Overrides: make(map[string]OverrideInfo),
			}
			assert.Equal(t, region, resp.SESRegion)
		}
	})
}
