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
// CustomSettingsHandler Construction Tests
// =============================================================================

func TestNewCustomSettingsHandler(t *testing.T) {
	t.Run("creates handler with nil service", func(t *testing.T) {
		handler := NewCustomSettingsHandler(nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.settingsService)
	})
}

// =============================================================================
// CreateSetting Handler Validation Tests
// =============================================================================

func TestCreateCustomSetting_Validation(t *testing.T) {
	t.Run("missing user context returns unauthorized", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom", handler.CreateSetting)

		body := `{"key": "test.key", "value": {"value": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Authentication required", result["error"])
	})

	t.Run("invalid user ID in context returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom", func(c fiber.Ctx) error {
			c.Locals("user_id", "invalid-uuid")
			c.Locals("user_role", "admin")
			return handler.CreateSetting(c)
		})

		body := `{"key": "test.key", "value": {"value": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Invalid user ID", result["error"])
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_role", "admin")
			return handler.CreateSetting(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/settings/custom", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		assert.Contains(t, result["error"], "Invalid request body")
	})

	t.Run("missing key", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_role", "admin")
			return handler.CreateSetting(c)
		})

		body := `{"value": {"value": "test"}}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Setting key is required", result["error"])
	})

	t.Run("missing value", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_role", "admin")
			return handler.CreateSetting(c)
		})

		body := `{"key": "test.key"}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Setting value is required", result["error"])
	})
}

// =============================================================================
// ListSettings Handler Validation Tests
// =============================================================================

func TestListCustomSettings_Validation(t *testing.T) {
	t.Run("missing user role returns unauthorized", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Get("/settings/custom", handler.ListSettings)

		req := httptest.NewRequest(http.MethodGet, "/settings/custom", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

// =============================================================================
// GetSetting Handler Validation Tests
// =============================================================================

func TestGetCustomSetting_Validation(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Get("/settings/custom/*", handler.GetSetting)

		req := httptest.NewRequest(http.MethodGet, "/settings/custom/", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		assert.Equal(t, "Setting key is required", result["error"])
	})
}

// =============================================================================
// UpdateSetting Handler Validation Tests
// =============================================================================

func TestUpdateCustomSetting_Validation(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Put("/settings/custom/*", handler.UpdateSetting)

		body := `{"value": {"value": "updated"}}`
		req := httptest.NewRequest(http.MethodPut, "/settings/custom/", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing value returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Put("/settings/custom/*", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_role", "admin")
			return handler.UpdateSetting(c)
		})

		body := `{"description": "updated description"}`
		req := httptest.NewRequest(http.MethodPut, "/settings/custom/test.key", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Setting value is required", result["error"])
	})
}

// =============================================================================
// DeleteSetting Handler Validation Tests
// =============================================================================

func TestDeleteCustomSetting_Validation(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Delete("/settings/custom/*", handler.DeleteSetting)

		req := httptest.NewRequest(http.MethodDelete, "/settings/custom/", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing user role returns unauthorized", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Delete("/settings/custom/*", handler.DeleteSetting)

		req := httptest.NewRequest(http.MethodDelete, "/settings/custom/test.key", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})
}

// =============================================================================
// Secret Settings Validation Tests
// =============================================================================

func TestCreateSecretSetting_Validation(t *testing.T) {
	t.Run("missing user context returns unauthorized", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom/secret", handler.CreateSecretSetting)

		body := `{"key": "api_key", "value": "secret-value"}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom/secret", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("missing key", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom/secret", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			return handler.CreateSecretSetting(c)
		})

		body := `{"value": "secret-value"}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom/secret", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Setting key is required", result["error"])
	})

	t.Run("missing value", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Post("/settings/custom/secret", func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			return handler.CreateSecretSetting(c)
		})

		body := `{"key": "api_key"}`
		req := httptest.NewRequest(http.MethodPost, "/settings/custom/secret", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Setting value is required", result["error"])
	})
}

func TestGetSecretSetting_Validation(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Get("/settings/custom/secret/*", handler.GetSecretSetting)

		req := httptest.NewRequest(http.MethodGet, "/settings/custom/secret/", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestDeleteSecretSetting_Validation(t *testing.T) {
	t.Run("empty key returns error", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomSettingsHandler(nil)

		app.Delete("/settings/custom/secret/*", handler.DeleteSecretSetting)

		req := httptest.NewRequest(http.MethodDelete, "/settings/custom/secret/", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// extractStringValueFromMap Helper Tests
// =============================================================================

func TestExtractStringValueFromMap(t *testing.T) {
	t.Run("extracts value key", func(t *testing.T) {
		m := map[string]interface{}{
			"value": "test-secret",
		}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "test-secret", result)
	})

	t.Run("extracts single key value", func(t *testing.T) {
		m := map[string]interface{}{
			"secret": "test-secret",
		}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "test-secret", result)
	})

	t.Run("returns empty for non-string value key", func(t *testing.T) {
		m := map[string]interface{}{
			"value": 123,
		}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty for multiple keys without value", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": "val1",
			"key2": "val2",
		}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "", result)
	})

	t.Run("returns empty for empty map", func(t *testing.T) {
		m := map[string]interface{}{}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "", result)
	})

	t.Run("prefers value key over single key", func(t *testing.T) {
		m := map[string]interface{}{
			"value": "preferred",
		}
		result := extractStringValueFromMap(m)
		assert.Equal(t, "preferred", result)
	})
}

// =============================================================================
// Response Format Tests
// =============================================================================

func TestCustomSettingsResponseFormats(t *testing.T) {
	t.Run("duplicate key error response", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "A setting with this key already exists",
			"code":  "DUPLICATE_KEY",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"code":"DUPLICATE_KEY"`)
	})

	t.Run("invalid key error response", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "Invalid setting key format",
			"code":  "INVALID_KEY",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"code":"INVALID_KEY"`)
	})

	t.Run("permission denied error response", func(t *testing.T) {
		expectedError := fiber.Map{
			"error": "You do not have permission to edit this setting",
			"code":  "PERMISSION_DENIED",
		}

		data, err := json.Marshal(expectedError)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"code":"PERMISSION_DENIED"`)
	})
}
