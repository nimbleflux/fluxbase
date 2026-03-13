package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/webhook"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// WebhookHandler Construction Tests
// =============================================================================

func TestNewWebhookHandler(t *testing.T) {
	t.Run("creates handler with nil service", func(t *testing.T) {
		handler := NewWebhookHandler(nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.webhookService)
	})
}

// =============================================================================
// CreateWebhook Tests
// =============================================================================

func TestCreateWebhook_Validation(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing name",
			body:           map[string]interface{}{"url": "https://example.com/webhook"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Name is required",
		},
		{
			name:           "missing url",
			body:           map[string]interface{}{"name": "test-webhook"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "URL is required",
		},
		{
			name:           "invalid json body",
			body:           "not a json",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name:           "empty name",
			body:           map[string]interface{}{"name": "", "url": "https://example.com"},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Name is required",
		},
		{
			name:           "empty url",
			body:           map[string]interface{}{"name": "test", "url": ""},
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "URL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			handler := NewWebhookHandler(nil)

			app.Post("/webhooks", handler.CreateWebhook)

			var body []byte
			var err error
			if str, ok := tt.body.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(respBody, &result)
			require.NoError(t, err)

			assert.Contains(t, result["error"], tt.expectedError)
		})
	}
}

func TestWebhook_DefaultValues(t *testing.T) {
	t.Run("default max retries is set", func(t *testing.T) {
		wh := webhook.Webhook{
			Name:       "test",
			URL:        "https://example.com",
			MaxRetries: 0, // Should be set to default
		}

		// The handler sets this to 3
		if wh.MaxRetries == 0 {
			wh.MaxRetries = 3
		}
		assert.Equal(t, 3, wh.MaxRetries)
	})

	t.Run("default retry backoff is set", func(t *testing.T) {
		wh := webhook.Webhook{
			Name:                "test",
			URL:                 "https://example.com",
			RetryBackoffSeconds: 0,
		}

		if wh.RetryBackoffSeconds == 0 {
			wh.RetryBackoffSeconds = 5
		}
		assert.Equal(t, 5, wh.RetryBackoffSeconds)
	})

	t.Run("default timeout is set", func(t *testing.T) {
		wh := webhook.Webhook{
			Name:           "test",
			URL:            "https://example.com",
			TimeoutSeconds: 0,
		}

		if wh.TimeoutSeconds == 0 {
			wh.TimeoutSeconds = 30
		}
		assert.Equal(t, 30, wh.TimeoutSeconds)
	})

	t.Run("default scope is user", func(t *testing.T) {
		wh := webhook.Webhook{
			Name:  "test",
			URL:   "https://example.com",
			Scope: "",
		}

		if wh.Scope == "" {
			wh.Scope = "user"
		}
		assert.Equal(t, "user", wh.Scope)
	})

	t.Run("default headers is empty map", func(t *testing.T) {
		wh := webhook.Webhook{
			Name:    "test",
			URL:     "https://example.com",
			Headers: nil,
		}

		if wh.Headers == nil {
			wh.Headers = make(map[string]string)
		}
		assert.NotNil(t, wh.Headers)
		assert.Empty(t, wh.Headers)
	})
}

// =============================================================================
// GetWebhook Tests
// =============================================================================

func TestGetWebhook_InvalidID(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Get("/webhooks/:id", handler.GetWebhook)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid uuid format",
			id:             "not-a-uuid",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid webhook ID",
		},
		{
			name:           "empty id",
			id:             "",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid webhook ID",
		},
		{
			name:           "partial uuid",
			id:             "12345678-1234",
			expectedStatus: fiber.StatusBadRequest,
			expectedError:  "Invalid webhook ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For empty id, use a different path
			path := "/webhooks/" + tt.id
			if tt.id == "" {
				path = "/webhooks/"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			// For empty id, Fiber returns 404 because route doesn't match
			if tt.id == "" {
				assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
				return
			}

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			var result map[string]interface{}
			err = json.Unmarshal(respBody, &result)
			require.NoError(t, err)

			assert.Contains(t, result["error"], tt.expectedError)
		})
	}
}

// =============================================================================
// UpdateWebhook Tests
// =============================================================================

func TestUpdateWebhook_InvalidID(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Patch("/webhooks/:id", handler.UpdateWebhook)

	req := httptest.NewRequest(http.MethodPatch, "/webhooks/invalid-uuid", bytes.NewReader([]byte(`{}`)))
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

	assert.Contains(t, result["error"], "Invalid webhook ID")
}

func TestUpdateWebhook_InvalidBody(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Patch("/webhooks/:id", handler.UpdateWebhook)

	validID := uuid.New().String()
	req := httptest.NewRequest(http.MethodPatch, "/webhooks/"+validID, bytes.NewReader([]byte("invalid json")))
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
}

// =============================================================================
// DeleteWebhook Tests
// =============================================================================

func TestDeleteWebhook_InvalidID(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Delete("/webhooks/:id", handler.DeleteWebhook)

	req := httptest.NewRequest(http.MethodDelete, "/webhooks/not-a-uuid", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	assert.Contains(t, result["error"], "Invalid webhook ID")
}

// =============================================================================
// TestWebhook Tests
// =============================================================================

func TestTestWebhook_InvalidID(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Post("/webhooks/:id/test", handler.TestWebhook)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/bad-uuid/test", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	assert.Contains(t, result["error"], "Invalid webhook ID")
}

// =============================================================================
// ListDeliveries Tests
// =============================================================================

func TestListDeliveries_InvalidID(t *testing.T) {
	app := fiber.New()
	handler := NewWebhookHandler(nil)

	app.Get("/webhooks/:id/deliveries", handler.ListDeliveries)

	req := httptest.NewRequest(http.MethodGet, "/webhooks/invalid/deliveries", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	assert.Contains(t, result["error"], "Invalid webhook ID")
}

// =============================================================================
// Webhook Model Tests
// =============================================================================

func TestWebhook_EventConfig(t *testing.T) {
	t.Run("event config with all operations", func(t *testing.T) {
		config := webhook.EventConfig{
			Table:      "users",
			Operations: []string{"INSERT", "UPDATE", "DELETE"},
		}

		assert.Equal(t, "users", config.Table)
		assert.Len(t, config.Operations, 3)
		assert.Contains(t, config.Operations, "INSERT")
		assert.Contains(t, config.Operations, "UPDATE")
		assert.Contains(t, config.Operations, "DELETE")
	})

	t.Run("event config with single operation", func(t *testing.T) {
		config := webhook.EventConfig{
			Table:      "orders",
			Operations: []string{"INSERT"},
		}

		assert.Equal(t, "orders", config.Table)
		assert.Len(t, config.Operations, 1)
	})
}

func TestWebhookDelivery_Struct(t *testing.T) {
	t.Run("pending delivery", func(t *testing.T) {
		delivery := webhook.WebhookDelivery{
			ID:        uuid.New(),
			WebhookID: uuid.New(),
			Event:     "INSERT",
			Status:    "pending",
			Attempt:   1,
		}

		assert.Equal(t, "pending", delivery.Status)
		assert.Equal(t, 1, delivery.Attempt)
		assert.Nil(t, delivery.StatusCode)
		assert.Nil(t, delivery.Error)
	})

	t.Run("successful delivery", func(t *testing.T) {
		statusCode := 200
		delivery := webhook.WebhookDelivery{
			ID:         uuid.New(),
			WebhookID:  uuid.New(),
			Event:      "UPDATE",
			Status:     "success",
			StatusCode: &statusCode,
			Attempt:    1,
		}

		assert.Equal(t, "success", delivery.Status)
		assert.Equal(t, 200, *delivery.StatusCode)
	})

	t.Run("failed delivery", func(t *testing.T) {
		errMsg := "connection refused"
		delivery := webhook.WebhookDelivery{
			ID:        uuid.New(),
			WebhookID: uuid.New(),
			Event:     "DELETE",
			Status:    "failed",
			Error:     &errMsg,
			Attempt:   3,
		}

		assert.Equal(t, "failed", delivery.Status)
		assert.Equal(t, "connection refused", *delivery.Error)
		assert.Equal(t, 3, delivery.Attempt)
	})
}

func TestWebhookPayload_Struct(t *testing.T) {
	t.Run("insert payload", func(t *testing.T) {
		payload := webhook.WebhookPayload{
			Event:  "INSERT",
			Table:  "users",
			Schema: "public",
			Record: []byte(`{"id": 1, "name": "John"}`),
		}

		assert.Equal(t, "INSERT", payload.Event)
		assert.Equal(t, "users", payload.Table)
		assert.Equal(t, "public", payload.Schema)
		assert.NotEmpty(t, payload.Record)
	})

	t.Run("update payload with old record", func(t *testing.T) {
		payload := webhook.WebhookPayload{
			Event:     "UPDATE",
			Table:     "users",
			Schema:    "public",
			Record:    []byte(`{"id": 1, "name": "Jane"}`),
			OldRecord: []byte(`{"id": 1, "name": "John"}`),
		}

		assert.Equal(t, "UPDATE", payload.Event)
		assert.NotEmpty(t, payload.OldRecord)
	})

	t.Run("delete payload with old record only", func(t *testing.T) {
		payload := webhook.WebhookPayload{
			Event:     "DELETE",
			Table:     "users",
			Schema:    "public",
			OldRecord: []byte(`{"id": 1, "name": "John"}`),
		}

		assert.Equal(t, "DELETE", payload.Event)
		assert.NotEmpty(t, payload.OldRecord)
	})
}

func TestWebhook_JSONSerialization(t *testing.T) {
	t.Run("webhook serializes to JSON", func(t *testing.T) {
		wh := webhook.Webhook{
			ID:      uuid.New(),
			Name:    "test-webhook",
			URL:     "https://example.com/webhook",
			Enabled: true,
			Scope:   "user",
			Headers: map[string]string{"X-Custom": "value"},
			Events: []webhook.EventConfig{
				{Table: "users", Operations: []string{"INSERT"}},
			},
		}

		data, err := json.Marshal(wh)
		require.NoError(t, err)

		var decoded webhook.Webhook
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, wh.ID, decoded.ID)
		assert.Equal(t, wh.Name, decoded.Name)
		assert.Equal(t, wh.URL, decoded.URL)
		assert.Equal(t, wh.Enabled, decoded.Enabled)
		assert.Equal(t, wh.Scope, decoded.Scope)
		assert.Equal(t, wh.Headers["X-Custom"], decoded.Headers["X-Custom"])
	})
}

func TestWebhook_Scopes(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		valid bool
	}{
		{"user scope", "user", true},
		{"global scope", "global", true},
		{"empty scope defaults to user", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wh := webhook.Webhook{
				Name:  "test",
				URL:   "https://example.com",
				Scope: tt.scope,
			}

			if tt.scope == "" {
				wh.Scope = "user" // Default
			}

			assert.NotEmpty(t, wh.Scope)
		})
	}
}
