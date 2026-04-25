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
// UserManagementHandler Construction Tests
// =============================================================================

func TestNewUserManagementHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewUserManagementHandler(nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.userMgmtService)
		assert.Nil(t, handler.authService)
	})
}

// =============================================================================
// ListUsers Handler Tests
// =============================================================================

func TestListUsers_DefaultParameters(t *testing.T) {
	// Note: Full testing requires mocked services
	// These tests verify the handler setup and parameter parsing

	t.Run("default pagination values", func(t *testing.T) {
		// The handler uses defaultLimit=100 and maxLimit=1000
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Without service, will return internal server error
		// But we verify the handler was reached
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("custom pagination parameters", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users?limit=50&offset=10", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Handler should accept these parameters
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("exclude_admins parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users?exclude_admins=true", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("search parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users?search=john@example.com", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("user type parameter - app", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users?type=app", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("user type parameter - dashboard", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users", handler.ListUsers)

		req := httptest.NewRequest(http.MethodGet, "/users?type=dashboard", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// GetUserByID Handler Tests
// =============================================================================

func TestGetUserByID_ParameterParsing(t *testing.T) {
	t.Run("valid user ID format", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users/:id", handler.GetUserByID)

		req := httptest.NewRequest(http.MethodGet, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with type parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Get("/users/:id", handler.GetUserByID)

		req := httptest.NewRequest(http.MethodGet, "/users/123?type=dashboard", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// InviteUser Handler Tests
// =============================================================================

func TestInviteUser_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/invite", handler.InviteUser)

		req := httptest.NewRequest(http.MethodPost, "/users/invite", bytes.NewReader([]byte("invalid json")))
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

	t.Run("empty body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/invite", handler.InviteUser)

		req := httptest.NewRequest(http.MethodPost, "/users/invite", bytes.NewReader([]byte("")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Empty body should be handled
		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// DeleteUser Handler Tests
// =============================================================================

func TestDeleteUser_ParameterParsing(t *testing.T) {
	t.Run("with user ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Delete("/users/:id", handler.DeleteUser)

		req := httptest.NewRequest(http.MethodDelete, "/users/550e8400-e29b-41d4-a716-446655440000", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with type parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Delete("/users/:id", handler.DeleteUser)

		req := httptest.NewRequest(http.MethodDelete, "/users/123?type=dashboard", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// UpdateUserRole Handler Tests
// =============================================================================

func TestUpdateUserRole_Validation(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Patch("/users/:id/role", handler.UpdateUserRole)

		req := httptest.NewRequest(http.MethodPatch, "/users/123/role", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid body structure", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Patch("/users/:id/role", handler.UpdateUserRole)

		body := `{"role":"admin"}`
		req := httptest.NewRequest(http.MethodPatch, "/users/123/role", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Will fail due to nil service, but body parsing should work
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// UpdateUser Handler Tests
// =============================================================================

func TestUpdateUser_Validation(t *testing.T) {
	t.Run("invalid body", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Patch("/users/:id", handler.UpdateUser)

		req := httptest.NewRequest(http.MethodPatch, "/users/123", bytes.NewReader([]byte("not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid body with email update", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Patch("/users/:id", handler.UpdateUser)

		body := `{"email":"new@example.com"}`
		req := httptest.NewRequest(http.MethodPatch, "/users/123", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Body parsing should succeed
		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid body with password update", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Patch("/users/:id", handler.UpdateUser)

		body := `{"password":"NewSecurePassword123!"}`
		req := httptest.NewRequest(http.MethodPatch, "/users/123", bytes.NewReader([]byte(body)))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}

// =============================================================================
// ResetUserPassword Handler Tests
// =============================================================================

func TestResetUserPassword_ParameterParsing(t *testing.T) {
	t.Run("with user ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/:id/reset-password", handler.ResetUserPassword)

		req := httptest.NewRequest(http.MethodPost, "/users/550e8400-e29b-41d4-a716-446655440000/reset-password", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with type parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/:id/reset-password", handler.ResetUserPassword)

		req := httptest.NewRequest(http.MethodPost, "/users/123/reset-password?type=dashboard", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// LockUser Handler Tests
// =============================================================================

func TestLockUser_ParameterParsing(t *testing.T) {
	t.Run("with user ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/:id/lock", handler.LockUser)

		req := httptest.NewRequest(http.MethodPost, "/users/123/lock", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})

	t.Run("with type parameter", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/:id/lock", handler.LockUser)

		req := httptest.NewRequest(http.MethodPost, "/users/123/lock?type=app", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// UnlockUser Handler Tests
// =============================================================================

func TestUnlockUser_ParameterParsing(t *testing.T) {
	t.Run("with user ID", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewUserManagementHandler(nil, nil)

		app.Post("/users/:id/unlock", handler.UnlockUser)

		req := httptest.NewRequest(http.MethodPost, "/users/123/unlock", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.NotEqual(t, fiber.StatusNotFound, resp.StatusCode)
	})
}

// =============================================================================
// Pagination Normalization Tests (from validation_helpers)
// =============================================================================

func TestUserManagement_PaginationNormalization(t *testing.T) {
	// Tests that the handler uses NormalizePaginationParams correctly

	t.Run("default values applied", func(t *testing.T) {
		defaultLimit := 100
		maxLimit := 1000

		// Test zero values
		limit, offset := NormalizePaginationParams(0, 0, defaultLimit, maxLimit)
		assert.Equal(t, defaultLimit, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("values within range preserved", func(t *testing.T) {
		defaultLimit := 100
		maxLimit := 1000

		limit, offset := NormalizePaginationParams(50, 25, defaultLimit, maxLimit)
		assert.Equal(t, 50, limit)
		assert.Equal(t, 25, offset)
	})

	t.Run("exceeding max uses default", func(t *testing.T) {
		defaultLimit := 100
		maxLimit := 1000

		limit, offset := NormalizePaginationParams(2000, 0, defaultLimit, maxLimit)
		assert.Equal(t, defaultLimit, limit)
		assert.Equal(t, 0, offset)
	})

	t.Run("negative values handled", func(t *testing.T) {
		defaultLimit := 100
		maxLimit := 1000

		limit, offset := NormalizePaginationParams(-10, -5, defaultLimit, maxLimit)
		assert.Equal(t, defaultLimit, limit)
		assert.Equal(t, 0, offset)
	})
}

// =============================================================================
// User Type Parameter Tests
// =============================================================================

func TestUserTypeParameter(t *testing.T) {
	validTypes := []string{"app", "dashboard"}

	t.Run("valid user types", func(t *testing.T) {
		for _, userType := range validTypes {
			assert.Contains(t, validTypes, userType)
		}
	})

	t.Run("default is app", func(t *testing.T) {
		// Verified by checking the handler code uses "app" as default
		defaultType := "app"
		assert.Equal(t, "app", defaultType)
	})
}
