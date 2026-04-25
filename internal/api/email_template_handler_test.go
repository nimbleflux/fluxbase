package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// EmailTemplateHandler Construction Tests
// =============================================================================

func TestNewEmailTemplateHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewEmailTemplateHandler(nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.db)
		assert.Nil(t, handler.emailService)
	})
}

// =============================================================================
// EmailTemplate Struct Tests
// =============================================================================

func TestEmailTemplate_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		textBody := "Plain text body"
		template := EmailTemplate{
			ID:           uuid.New(),
			TemplateType: "magic_link",
			Subject:      "Your Magic Link",
			HTMLBody:     "<html><body><h1>Magic Link</h1></body></html>",
			TextBody:     &textBody,
			IsCustom:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.NotEqual(t, uuid.Nil, template.ID)
		assert.Equal(t, "magic_link", template.TemplateType)
		assert.Equal(t, "Your Magic Link", template.Subject)
		assert.Contains(t, template.HTMLBody, "Magic Link")
		assert.Equal(t, "Plain text body", *template.TextBody)
		assert.True(t, template.IsCustom)
	})

	t.Run("without optional text body", func(t *testing.T) {
		template := EmailTemplate{
			ID:           uuid.New(),
			TemplateType: "password_reset",
			Subject:      "Reset Password",
			HTMLBody:     "<html><body>Reset your password</body></html>",
			TextBody:     nil,
			IsCustom:     false,
		}

		assert.Nil(t, template.TextBody)
		assert.False(t, template.IsCustom)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		textBody := "Test text"
		template := EmailTemplate{
			ID:           uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			TemplateType: "email_verification",
			Subject:      "Verify Email",
			HTMLBody:     "<p>Verify</p>",
			TextBody:     &textBody,
			IsCustom:     true,
			CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(template)
		require.NoError(t, err)

		// Note: Go's JSON encoder escapes HTML characters by default (< becomes \u003c, > becomes \u003e)
		assert.Contains(t, string(data), `"id":"550e8400-e29b-41d4-a716-446655440000"`)
		assert.Contains(t, string(data), `"template_type":"email_verification"`)
		assert.Contains(t, string(data), `"subject":"Verify Email"`)
		// Check for HTML body - Go's json.Marshal escapes < and > by default
		assert.Contains(t, string(data), `"html_body":"`)
		assert.Contains(t, string(data), `Verify`)
		assert.Contains(t, string(data), `"text_body":"Test text"`)
		assert.Contains(t, string(data), `"is_custom":true`)
		assert.Contains(t, string(data), `"created_at"`)
		assert.Contains(t, string(data), `"updated_at"`)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"id": "550e8400-e29b-41d4-a716-446655440000",
			"template_type": "magic_link",
			"subject": "Sign In",
			"html_body": "<p>Click here</p>",
			"text_body": "Click the link",
			"is_custom": false
		}`

		var template EmailTemplate
		err := json.Unmarshal([]byte(jsonData), &template)
		require.NoError(t, err)

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", template.ID.String())
		assert.Equal(t, "magic_link", template.TemplateType)
		assert.Equal(t, "Sign In", template.Subject)
		assert.Equal(t, "<p>Click here</p>", template.HTMLBody)
		assert.Equal(t, "Click the link", *template.TextBody)
		assert.False(t, template.IsCustom)
	})
}

// =============================================================================
// UpdateTemplateRequest Struct Tests
// =============================================================================

func TestUpdateTemplateRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		textBody := "Plain text version"
		req := UpdateTemplateRequest{
			Subject:  "New Subject",
			HTMLBody: "<html><body>New body</body></html>",
			TextBody: &textBody,
		}

		assert.Equal(t, "New Subject", req.Subject)
		assert.Contains(t, req.HTMLBody, "New body")
		assert.Equal(t, "Plain text version", *req.TextBody)
	})

	t.Run("without optional text body", func(t *testing.T) {
		req := UpdateTemplateRequest{
			Subject:  "Subject Only",
			HTMLBody: "<p>HTML only</p>",
			TextBody: nil,
		}

		assert.Nil(t, req.TextBody)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"subject": "Updated Subject",
			"html_body": "<div>Updated</div>",
			"text_body": "Updated text"
		}`

		var req UpdateTemplateRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "Updated Subject", req.Subject)
		assert.Equal(t, "<div>Updated</div>", req.HTMLBody)
		assert.Equal(t, "Updated text", *req.TextBody)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		textBody := "Plain"
		req := UpdateTemplateRequest{
			Subject:  "Test",
			HTMLBody: "<p>Test</p>",
			TextBody: &textBody,
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"subject":"Test"`)
		// JSON encoder escapes HTML characters by default: < becomes \u003c and > becomes \u003e
		assert.Contains(t, string(data), `"html_body":"\u003cp\u003eTest\u003c/p\u003e"`)
		assert.Contains(t, string(data), `"text_body":"Plain"`)
	})
}

// =============================================================================
// TestEmailRequest Struct Tests (for templates)
// =============================================================================

func TestEmailTemplateTestEmailRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := TestEmailRequest{
			RecipientEmail: "test@example.com",
		}

		assert.Equal(t, "test@example.com", req.RecipientEmail)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"recipient_email":"user@test.com"}`

		var req TestEmailRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "user@test.com", req.RecipientEmail)
	})
}

// =============================================================================
// Default Templates Tests
// =============================================================================

func TestDefaultTemplates(t *testing.T) {
	t.Run("magic_link template exists", func(t *testing.T) {
		template, exists := defaultTemplates["magic_link"]
		assert.True(t, exists)
		assert.Equal(t, "magic_link", template.TemplateType)
		assert.Contains(t, template.Subject, "Magic Link")
		assert.Contains(t, template.HTMLBody, "{{.MagicLink}}")
		assert.NotNil(t, template.TextBody)
		assert.False(t, template.IsCustom)
	})

	t.Run("email_verification template exists", func(t *testing.T) {
		template, exists := defaultTemplates["email_verification"]
		assert.True(t, exists)
		assert.Equal(t, "email_verification", template.TemplateType)
		assert.Contains(t, template.Subject, "Verify")
		assert.Contains(t, template.HTMLBody, "{{.VerificationLink}}")
		assert.NotNil(t, template.TextBody)
		assert.False(t, template.IsCustom)
	})

	t.Run("password_reset template exists", func(t *testing.T) {
		template, exists := defaultTemplates["password_reset"]
		assert.True(t, exists)
		assert.Equal(t, "password_reset", template.TemplateType)
		assert.Contains(t, template.Subject, "Reset")
		assert.Contains(t, template.HTMLBody, "{{.ResetLink}}")
		assert.NotNil(t, template.TextBody)
		assert.False(t, template.IsCustom)
	})

	t.Run("all default templates have required fields", func(t *testing.T) {
		for templateType, template := range defaultTemplates {
			assert.NotEmpty(t, template.Subject, "Template %s missing subject", templateType)
			assert.NotEmpty(t, template.HTMLBody, "Template %s missing HTML body", templateType)
			assert.Equal(t, templateType, template.TemplateType, "Template type mismatch for %s", templateType)
		}
	})

	t.Run("default templates contain AppName placeholder", func(t *testing.T) {
		for templateType, template := range defaultTemplates {
			assert.Contains(t, template.Subject, "{{.AppName}}", "Template %s subject should contain AppName", templateType)
		}
	})
}

// =============================================================================
// GetTemplate Handler Tests
// =============================================================================

func TestGetTemplate_Validation(t *testing.T) {
	t.Run("invalid template type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Get("/templates/:type", handler.GetTemplate)

		req := httptest.NewRequest(http.MethodGet, "/templates/invalid_type", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid template type")
	})

	t.Run("valid template type - magic_link", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Get("/templates/:type", handler.GetTemplate)

		req := httptest.NewRequest(http.MethodGet, "/templates/magic_link", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Template type is valid, fails at DB query
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid template type - email_verification", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Get("/templates/:type", handler.GetTemplate)

		req := httptest.NewRequest(http.MethodGet, "/templates/email_verification", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid template type - password_reset", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Get("/templates/:type", handler.GetTemplate)

		req := httptest.NewRequest(http.MethodGet, "/templates/password_reset", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// UpdateTemplate Handler Tests
// =============================================================================

func TestUpdateTemplate_Validation(t *testing.T) {
	t.Run("invalid template type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		body := `{"subject":"Test","html_body":"<p>Test</p>"}`
		req := httptest.NewRequest(http.MethodPut, "/templates/invalid_type", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Invalid template type")
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		req := httptest.NewRequest(http.MethodPut, "/templates/magic_link", bytes.NewReader([]byte("invalid json")))
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

	t.Run("missing subject", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		body := `{"html_body":"<p>Test</p>"}`
		req := httptest.NewRequest(http.MethodPut, "/templates/magic_link", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Subject and HTML body are required")
	})

	t.Run("missing html_body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		body := `{"subject":"Test Subject"}`
		req := httptest.NewRequest(http.MethodPut, "/templates/magic_link", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("empty subject", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		body := `{"subject":"","html_body":"<p>Test</p>"}`
		req := httptest.NewRequest(http.MethodPut, "/templates/magic_link", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid request", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Put("/templates/:type", handler.UpdateTemplate)

		body := `{"subject":"Custom Subject","html_body":"<p>Custom Body</p>","text_body":"Plain text"}`
		req := httptest.NewRequest(http.MethodPut, "/templates/magic_link", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Validation passes, fails at DB operation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// ResetTemplate Handler Tests
// =============================================================================

func TestResetTemplate_Validation(t *testing.T) {
	t.Run("invalid template type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/reset", handler.ResetTemplate)

		req := httptest.NewRequest(http.MethodPost, "/templates/invalid_type/reset", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid template type")
	})

	t.Run("valid template type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/reset", handler.ResetTemplate)

		req := httptest.NewRequest(http.MethodPost, "/templates/magic_link/reset", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Template type is valid, fails at DB operation
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// TestTemplate Handler Tests
// =============================================================================

func TestTestTemplate_Validation(t *testing.T) {
	t.Run("invalid template type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/test", handler.TestTemplate)

		body := `{"recipient_email":"test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/templates/invalid_type/test", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Invalid template type")
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/test", handler.TestTemplate)

		req := httptest.NewRequest(http.MethodPost, "/templates/magic_link/test", bytes.NewReader([]byte("invalid")))
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
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/test", handler.TestTemplate)

		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/templates/magic_link/test", bytes.NewReader([]byte(body)))
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
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil)

		app.Post("/templates/:type/test", handler.TestTemplate)

		body := `{"recipient_email":""}`
		req := httptest.NewRequest(http.MethodPost, "/templates/magic_link/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("email service not configured", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewEmailTemplateHandler(nil, nil) // nil email service

		app.Post("/templates/:type/test", handler.TestTemplate)

		body := `{"recipient_email":"test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/templates/magic_link/test", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Email service not configured")
	})
}

// =============================================================================
// Template Type Constants Tests
// =============================================================================

func TestTemplateTypes(t *testing.T) {
	validTypes := []string{"magic_link", "email_verification", "password_reset"}
	invalidTypes := []string{"invalid", "custom", "unknown", "invite", "welcome"}

	t.Run("valid template types exist in defaultTemplates", func(t *testing.T) {
		for _, templateType := range validTypes {
			_, exists := defaultTemplates[templateType]
			assert.True(t, exists, "Expected template type %q to exist", templateType)
		}
	})

	t.Run("invalid template types do not exist in defaultTemplates", func(t *testing.T) {
		for _, templateType := range invalidTypes {
			_, exists := defaultTemplates[templateType]
			assert.False(t, exists, "Expected template type %q to not exist", templateType)
		}
	})
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestStringPtr(t *testing.T) {
	t.Run("creates pointer from string", func(t *testing.T) {
		result := stringPtr("test")
		assert.NotNil(t, result)
		assert.Equal(t, "test", *result)
	})

	t.Run("creates pointer from empty string", func(t *testing.T) {
		result := stringPtr("")
		assert.NotNil(t, result)
		assert.Equal(t, "", *result)
	})
}
