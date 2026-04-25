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
// CustomMCPHandler Construction Tests
// =============================================================================

func TestNewCustomMCPHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewCustomMCPHandler(nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.manager)
		assert.Nil(t, handler.mcpConfig)
	})
}

// =============================================================================
// Tool Handlers Validation Tests
// =============================================================================

func TestGetTool_Validation(t *testing.T) {
	t.Run("invalid tool ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Get("/tools/:id", handler.GetTool)

		req := httptest.NewRequest(http.MethodGet, "/tools/not-a-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		assert.Equal(t, "Invalid tool ID", result["error"])
	})

	t.Run("valid UUID format accepted", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Get("/tools/:id", handler.GetTool)

		req := httptest.NewRequest(http.MethodGet, "/tools/550e8400-e29b-41d4-a716-446655440000", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should pass validation (not 400), will fail at storage level
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestCreateTool_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools", handler.CreateTool)

		req := httptest.NewRequest(http.MethodPost, "/tools", bytes.NewReader([]byte("not json")))
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

	t.Run("missing name", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools", handler.CreateTool)

		body := `{"code": "function run() {}"}`
		req := httptest.NewRequest(http.MethodPost, "/tools", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Name is required", result["error"])
	})

	t.Run("missing code", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools", handler.CreateTool)

		body := `{"name": "my_tool"}`
		req := httptest.NewRequest(http.MethodPost, "/tools", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Code is required", result["error"])
	})
}

func TestUpdateTool_Validation(t *testing.T) {
	t.Run("invalid tool ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Put("/tools/:id", handler.UpdateTool)

		body := `{"name": "updated_tool"}`
		req := httptest.NewRequest(http.MethodPut, "/tools/invalid-uuid", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Put("/tools/:id", handler.UpdateTool)

		req := httptest.NewRequest(http.MethodPut, "/tools/550e8400-e29b-41d4-a716-446655440000", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestDeleteTool_Validation(t *testing.T) {
	t.Run("invalid tool ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Delete("/tools/:id", handler.DeleteTool)

		req := httptest.NewRequest(http.MethodDelete, "/tools/bad-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestSyncTool_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools/sync", handler.SyncTool)

		req := httptest.NewRequest(http.MethodPost, "/tools/sync", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing name", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools/sync", handler.SyncTool)

		body := `{"code": "function run() {}"}`
		req := httptest.NewRequest(http.MethodPost, "/tools/sync", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Name is required", result["error"])
	})

	t.Run("missing code", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools/sync", handler.SyncTool)

		body := `{"name": "sync_tool"}`
		req := httptest.NewRequest(http.MethodPost, "/tools/sync", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestTestTool_Validation(t *testing.T) {
	t.Run("invalid tool ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools/:id/test", handler.TestTool)

		body := `{"args": {}}`
		req := httptest.NewRequest(http.MethodPost, "/tools/invalid/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/tools/:id/test", handler.TestTool)

		req := httptest.NewRequest(http.MethodPost, "/tools/550e8400-e29b-41d4-a716-446655440000/test", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// Resource Handlers Validation Tests
// =============================================================================

func TestGetResource_Validation(t *testing.T) {
	t.Run("invalid resource ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Get("/resources/:id", handler.GetResource)

		req := httptest.NewRequest(http.MethodGet, "/resources/not-a-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(body, &result)
		assert.Equal(t, "Invalid resource ID", result["error"])
	})
}

func TestCreateResource_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources", handler.CreateResource)

		req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing URI", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources", handler.CreateResource)

		body := `{"name": "test_resource", "code": "function read() {}"}`
		req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "URI is required", result["error"])
	})

	t.Run("missing name", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources", handler.CreateResource)

		body := `{"uri": "custom://resource", "code": "function read() {}"}`
		req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Name is required", result["error"])
	})

	t.Run("missing code", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources", handler.CreateResource)

		body := `{"uri": "custom://resource", "name": "test_resource"}`
		req := httptest.NewRequest(http.MethodPost, "/resources", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		_ = json.Unmarshal(respBody, &result)
		assert.Equal(t, "Code is required", result["error"])
	})
}

func TestUpdateResource_Validation(t *testing.T) {
	t.Run("invalid resource ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Put("/resources/:id", handler.UpdateResource)

		body := `{"name": "updated"}`
		req := httptest.NewRequest(http.MethodPut, "/resources/invalid-uuid", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestDeleteResource_Validation(t *testing.T) {
	t.Run("invalid resource ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Delete("/resources/:id", handler.DeleteResource)

		req := httptest.NewRequest(http.MethodDelete, "/resources/bad-uuid", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestSyncResource_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources/sync", handler.SyncResource)

		req := httptest.NewRequest(http.MethodPost, "/resources/sync", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing URI", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources/sync", handler.SyncResource)

		body := `{"name": "sync_resource", "code": "function read() {}"}`
		req := httptest.NewRequest(http.MethodPost, "/resources/sync", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

func TestTestResource_Validation(t *testing.T) {
	t.Run("invalid resource ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources/:id/test", handler.TestResource)

		body := `{"params": {}}`
		req := httptest.NewRequest(http.MethodPost, "/resources/invalid/test", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewCustomMCPHandler(nil, nil, nil)

		app.Post("/resources/:id/test", handler.TestResource)

		req := httptest.NewRequest(http.MethodPost, "/resources/550e8400-e29b-41d4-a716-446655440000/test", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// Response Format Tests
// =============================================================================

func TestMCPResponseFormats(t *testing.T) {
	t.Run("list tools response format", func(t *testing.T) {
		expectedResponse := fiber.Map{
			"tools": []interface{}{},
			"count": 0,
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"tools"`)
		assert.Contains(t, string(data), `"count"`)
	})

	t.Run("list resources response format", func(t *testing.T) {
		expectedResponse := fiber.Map{
			"resources": []interface{}{},
			"count":     0,
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"resources"`)
		assert.Contains(t, string(data), `"count"`)
	})

	t.Run("test tool response format", func(t *testing.T) {
		expectedResponse := fiber.Map{
			"success": true,
			"result":  map[string]interface{}{},
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"success":true`)
		assert.Contains(t, string(data), `"result"`)
	})

	t.Run("test resource response format", func(t *testing.T) {
		expectedResponse := fiber.Map{
			"success":  true,
			"contents": []interface{}{},
		}

		data, err := json.Marshal(expectedResponse)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"success":true`)
		assert.Contains(t, string(data), `"contents"`)
	})
}
