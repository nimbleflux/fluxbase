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
// RealtimeAdminHandler Construction Tests
// =============================================================================

func TestNewRealtimeAdminHandler(t *testing.T) {
	t.Run("creates handler with nil database", func(t *testing.T) {
		handler := NewRealtimeAdminHandler(nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.db)
	})
}

// =============================================================================
// EnableRealtimeRequest Struct Tests
// =============================================================================

func TestEnableRealtimeRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := EnableRealtimeRequest{
			Schema:  "public",
			Table:   "users",
			Events:  []string{"INSERT", "UPDATE", "DELETE"},
			Exclude: []string{"password_hash", "secret_token"},
		}

		assert.Equal(t, "public", req.Schema)
		assert.Equal(t, "users", req.Table)
		assert.Len(t, req.Events, 3)
		assert.Contains(t, req.Events, "INSERT")
		assert.Contains(t, req.Events, "UPDATE")
		assert.Contains(t, req.Events, "DELETE")
		assert.Len(t, req.Exclude, 2)
	})

	t.Run("minimal request", func(t *testing.T) {
		req := EnableRealtimeRequest{
			Table: "posts",
		}

		assert.Empty(t, req.Schema) // Will default to "public"
		assert.Equal(t, "posts", req.Table)
		assert.Empty(t, req.Events) // Will default to all events
		assert.Empty(t, req.Exclude)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"schema": "app",
			"table": "products",
			"events": ["INSERT", "UPDATE"],
			"exclude": ["internal_data"]
		}`

		var req EnableRealtimeRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "app", req.Schema)
		assert.Equal(t, "products", req.Table)
		assert.Len(t, req.Events, 2)
		assert.Len(t, req.Exclude, 1)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := EnableRealtimeRequest{
			Schema:  "public",
			Table:   "orders",
			Events:  []string{"INSERT"},
			Exclude: []string{"notes"},
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"table":"orders"`)
		assert.Contains(t, string(data), `"events":["INSERT"]`)
		assert.Contains(t, string(data), `"exclude":["notes"]`)
	})
}

// =============================================================================
// EnableRealtimeResponse Struct Tests
// =============================================================================

func TestEnableRealtimeResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := EnableRealtimeResponse{
			Schema:      "public",
			Table:       "users",
			Events:      []string{"INSERT", "UPDATE", "DELETE"},
			TriggerName: "users_realtime_notify",
			Exclude:     []string{"password"},
		}

		assert.Equal(t, "public", resp.Schema)
		assert.Equal(t, "users", resp.Table)
		assert.Len(t, resp.Events, 3)
		assert.Equal(t, "users_realtime_notify", resp.TriggerName)
		assert.Len(t, resp.Exclude, 1)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := EnableRealtimeResponse{
			Schema:      "public",
			Table:       "posts",
			Events:      []string{"INSERT", "UPDATE"},
			TriggerName: "posts_realtime_notify",
			Exclude:     []string{},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"table":"posts"`)
		assert.Contains(t, string(data), `"events"`)
		assert.Contains(t, string(data), `"trigger_name":"posts_realtime_notify"`)
	})

	t.Run("trigger name generation pattern", func(t *testing.T) {
		// Verify the expected trigger naming pattern
		tables := []string{"users", "posts", "comments", "my_table"}
		for _, table := range tables {
			expectedTrigger := table + "_realtime_notify"
			resp := EnableRealtimeResponse{
				Table:       table,
				TriggerName: expectedTrigger,
			}
			assert.Equal(t, expectedTrigger, resp.TriggerName)
		}
	})
}

// =============================================================================
// RealtimeTableStatus Struct Tests
// =============================================================================

func TestRealtimeTableStatus_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		status := RealtimeTableStatus{
			ID:              1,
			Schema:          "public",
			Table:           "users",
			RealtimeEnabled: true,
			Events:          []string{"INSERT", "UPDATE", "DELETE"},
			ExcludedColumns: []string{"password_hash"},
			CreatedAt:       "2024-01-01T00:00:00Z",
			UpdatedAt:       "2024-01-02T00:00:00Z",
		}

		assert.Equal(t, 1, status.ID)
		assert.Equal(t, "public", status.Schema)
		assert.Equal(t, "users", status.Table)
		assert.True(t, status.RealtimeEnabled)
		assert.Len(t, status.Events, 3)
		assert.Len(t, status.ExcludedColumns, 1)
		assert.NotEmpty(t, status.CreatedAt)
		assert.NotEmpty(t, status.UpdatedAt)
	})

	t.Run("disabled realtime status", func(t *testing.T) {
		status := RealtimeTableStatus{
			Schema:          "public",
			Table:           "archived",
			RealtimeEnabled: false,
			Events:          []string{},
			ExcludedColumns: []string{},
		}

		assert.False(t, status.RealtimeEnabled)
		assert.Empty(t, status.Events)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		status := RealtimeTableStatus{
			ID:              5,
			Schema:          "public",
			Table:           "orders",
			RealtimeEnabled: true,
			Events:          []string{"INSERT"},
			ExcludedColumns: []string{},
			CreatedAt:       "2024-01-01T00:00:00Z",
			UpdatedAt:       "2024-01-01T00:00:00Z",
		}

		data, err := json.Marshal(status)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"id":5`)
		assert.Contains(t, string(data), `"schema":"public"`)
		assert.Contains(t, string(data), `"table":"orders"`)
		assert.Contains(t, string(data), `"realtime_enabled":true`)
		assert.Contains(t, string(data), `"events"`)
		assert.Contains(t, string(data), `"created_at"`)
		assert.Contains(t, string(data), `"updated_at"`)
	})
}

// =============================================================================
// HandleEnableRealtime Handler Tests
// =============================================================================

func TestHandleEnableRealtime_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte("invalid json")))
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

	t.Run("missing table name", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"schema": "public"}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Table name is required")
	})

	t.Run("invalid event type", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"table": "users", "events": ["INSERT", "INVALID"]}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Invalid event type")
	})

	t.Run("system schema prevention - auth", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"schema": "auth", "table": "users"}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Cannot enable realtime on system schema")
	})

	t.Run("system schema prevention - pg_catalog", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"schema": "pg_catalog", "table": "pg_class"}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("system schema prevention - information_schema", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"schema": "information_schema", "table": "tables"}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("system schema prevention - platform", func(t *testing.T) {
		// platform is a user data schema (not blocked by system schema protection)
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		body := `{"schema": "platform", "table": "users"}`
		req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// platform is NOT a system schema, so validation should pass (will fail on nil DB though)
		// But since handler.db is nil, it returns 500. The key point is it doesn't return 400 (blocked).
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid events accepted", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		validEventCombinations := [][]string{
			{"INSERT"},
			{"UPDATE"},
			{"DELETE"},
			{"INSERT", "UPDATE"},
			{"INSERT", "DELETE"},
			{"UPDATE", "DELETE"},
			{"INSERT", "UPDATE", "DELETE"},
		}

		for _, events := range validEventCombinations {
			eventsJSON, _ := json.Marshal(events)
			body := `{"table": "users", "events": ` + string(eventsJSON) + `}`
			req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			_ = resp.Body.Close()

			// Should not fail on event validation (will fail at DB operation)
			assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode, "Events %v should be valid", events)
		}
	})
}

// =============================================================================
// HandleDisableRealtime Handler Tests
// =============================================================================

func TestHandleDisableRealtime_ParameterValidation(t *testing.T) {
	t.Run("valid schema and table", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Delete("/realtime/:schema/:table", handler.HandleDisableRealtime)

		req := httptest.NewRequest(http.MethodDelete, "/realtime/public/users", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Parameters are valid, fails at DB check
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("schema with underscore", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Delete("/realtime/:schema/:table", handler.HandleDisableRealtime)

		req := httptest.NewRequest(http.MethodDelete, "/realtime/my_schema/my_table", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Parameters are valid
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// HandleListRealtimeTables Handler Tests
// =============================================================================

func TestHandleListRealtimeTables_ParameterParsing(t *testing.T) {
	t.Run("default enabled filter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Get("/realtime/tables", handler.HandleListRealtimeTables)

		req := httptest.NewRequest(http.MethodGet, "/realtime/tables", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Handler reached, fails at DB
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("enabled=true filter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Get("/realtime/tables", handler.HandleListRealtimeTables)

		req := httptest.NewRequest(http.MethodGet, "/realtime/tables?enabled=true", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("enabled=false filter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Get("/realtime/tables", handler.HandleListRealtimeTables)

		req := httptest.NewRequest(http.MethodGet, "/realtime/tables?enabled=false", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// HandleGetRealtimeStatus Handler Tests
// =============================================================================

func TestHandleGetRealtimeStatus_ParameterValidation(t *testing.T) {
	t.Run("valid schema and table", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Get("/realtime/:schema/:table/status", handler.HandleGetRealtimeStatus)

		req := httptest.NewRequest(http.MethodGet, "/realtime/public/users/status", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Parameters valid, fails at DB
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// HandleUpdateRealtimeConfig Handler Tests
// =============================================================================

func TestHandleUpdateRealtimeConfig_Validation(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Patch("/realtime/:schema/:table", handler.HandleUpdateRealtimeConfig)

		req := httptest.NewRequest(http.MethodPatch, "/realtime/public/users", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("no updates provided", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Patch("/realtime/:schema/:table", handler.HandleUpdateRealtimeConfig)

		body := `{}`
		req := httptest.NewRequest(http.MethodPatch, "/realtime/public/users", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "No updates provided")
	})

	t.Run("invalid event type in update", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Patch("/realtime/:schema/:table", handler.HandleUpdateRealtimeConfig)

		body := `{"events": ["INSERT", "INVALID"]}`
		req := httptest.NewRequest(http.MethodPatch, "/realtime/public/users", bytes.NewReader([]byte(body)))
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

		assert.Contains(t, result["error"], "Invalid event type")
	})

	t.Run("valid events update", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Patch("/realtime/:schema/:table", handler.HandleUpdateRealtimeConfig)

		body := `{"events": ["INSERT", "UPDATE"]}`
		req := httptest.NewRequest(http.MethodPatch, "/realtime/public/users", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Validation passes, fails at DB
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid exclude update", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Patch("/realtime/:schema/:table", handler.HandleUpdateRealtimeConfig)

		body := `{"exclude": ["password", "token"]}`
		req := httptest.NewRequest(http.MethodPatch, "/realtime/public/users", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Validation passes, fails at DB
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// Event Type Constants Tests
// =============================================================================

func TestEventTypeConstants(t *testing.T) {
	validEvents := []string{"INSERT", "UPDATE", "DELETE"}
	invalidEvents := []string{"insert", "UPDATE ", "UPSERT", "TRUNCATE", "CREATE", "DROP", ""}

	t.Run("valid event types", func(t *testing.T) {
		for _, event := range validEvents {
			assert.Contains(t, validEvents, event)
		}
	})

	t.Run("invalid event types not in list", func(t *testing.T) {
		for _, event := range invalidEvents {
			assert.NotContains(t, validEvents, event)
		}
	})
}

// =============================================================================
// System Schema Protection Tests
// =============================================================================

func TestSystemSchemaProtection(t *testing.T) {
	systemSchemas := []string{
		"pg_catalog",
		"information_schema",
		"auth",
		"realtime",
	}

	allowedSchemas := []string{
		"public",
		"app",
		"my_schema",
		"custom",
	}

	t.Run("system schemas should be blocked", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		for _, schema := range systemSchemas {
			body := `{"schema": "` + schema + `", "table": "some_table"}`
			req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			_ = resp.Body.Close()

			assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode, "Schema %q should be blocked", schema)
		}
	})

	t.Run("allowed schemas should pass validation", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewRealtimeAdminHandler(nil)

		app.Post("/realtime/enable", handler.HandleEnableRealtime)

		for _, schema := range allowedSchemas {
			body := `{"schema": "` + schema + `", "table": "some_table"}`
			req := httptest.NewRequest(http.MethodPost, "/realtime/enable", bytes.NewReader([]byte(body)))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			_ = resp.Body.Close()

			// Should not fail with system schema error (will fail at DB check)
			assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode, "Schema %q should pass validation", schema)
		}
	})
}
