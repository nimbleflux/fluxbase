package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// AdminAuthHandler Construction Tests
// =============================================================================

func TestNewAdminAuthHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.authService)
		assert.Nil(t, handler.userRepo)
		assert.Nil(t, handler.dashboardAuth)
		assert.Nil(t, handler.systemSettings)
		assert.Nil(t, handler.config)
	})
}

// =============================================================================
// SetupStatusResponse Struct Tests
// =============================================================================

func TestSetupStatusResponse_Struct(t *testing.T) {
	t.Run("needs setup true", func(t *testing.T) {
		resp := SetupStatusResponse{
			NeedsSetup: true,
			HasAdmin:   false,
		}

		assert.True(t, resp.NeedsSetup)
		assert.False(t, resp.HasAdmin)
	})

	t.Run("setup completed", func(t *testing.T) {
		resp := SetupStatusResponse{
			NeedsSetup: false,
			HasAdmin:   true,
		}

		assert.False(t, resp.NeedsSetup)
		assert.True(t, resp.HasAdmin)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := SetupStatusResponse{
			NeedsSetup: true,
			HasAdmin:   false,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"needs_setup":true`)
		assert.Contains(t, string(data), `"has_admin":false`)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"needs_setup":false,"has_admin":true}`

		var resp SetupStatusResponse
		err := json.Unmarshal([]byte(jsonData), &resp)
		require.NoError(t, err)

		assert.False(t, resp.NeedsSetup)
		assert.True(t, resp.HasAdmin)
	})
}

// =============================================================================
// InitialSetupRequest Struct Tests
// =============================================================================

func TestInitialSetupRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := InitialSetupRequest{
			Email:      "admin@example.com",
			Password:   "SecurePassword123!@#",
			Name:       "Admin User",
			SetupToken: "my-secret-setup-token",
		}

		assert.Equal(t, "admin@example.com", req.Email)
		assert.Equal(t, "SecurePassword123!@#", req.Password)
		assert.Equal(t, "Admin User", req.Name)
		assert.Equal(t, "my-secret-setup-token", req.SetupToken)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"email": "test@test.com",
			"password": "MyP@ssw0rd123!",
			"name": "Test Admin",
			"setup_token": "token123"
		}`

		var req InitialSetupRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test@test.com", req.Email)
		assert.Equal(t, "MyP@ssw0rd123!", req.Password)
		assert.Equal(t, "Test Admin", req.Name)
		assert.Equal(t, "token123", req.SetupToken)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := InitialSetupRequest{
			Email:      "admin@company.com",
			Password:   "SecurePass123!!",
			Name:       "Company Admin",
			SetupToken: "secret-token",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"email":"admin@company.com"`)
		assert.Contains(t, string(data), `"password":"SecurePass123!!"`)
		assert.Contains(t, string(data), `"name":"Company Admin"`)
		assert.Contains(t, string(data), `"setup_token":"secret-token"`)
	})
}

// =============================================================================
// InitialSetupResponse Struct Tests
// =============================================================================

func TestInitialSetupResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := InitialSetupResponse{
			User:         nil, // Would be a real user in production
			AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			RefreshToken: "refresh_token_value",
			ExpiresIn:    3600,
		}

		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, int64(3600), resp.ExpiresIn)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := InitialSetupResponse{
			User:         nil,
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    7200,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"access_token":"access_token"`)
		assert.Contains(t, string(data), `"refresh_token":"refresh_token"`)
		assert.Contains(t, string(data), `"expires_in":7200`)
	})
}

// =============================================================================
// AdminLoginRequest Struct Tests
// =============================================================================

func TestAdminLoginRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := AdminLoginRequest{
			Email:    "admin@example.com",
			Password: "AdminPassword123!",
		}

		assert.Equal(t, "admin@example.com", req.Email)
		assert.Equal(t, "AdminPassword123!", req.Password)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"email":"user@test.com","password":"testpass123"}`

		var req AdminLoginRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "user@test.com", req.Email)
		assert.Equal(t, "testpass123", req.Password)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := AdminLoginRequest{
			Email:    "test@example.com",
			Password: "pass123",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"email":"test@example.com"`)
		assert.Contains(t, string(data), `"password":"pass123"`)
	})
}

// =============================================================================
// AdminLoginResponse Struct Tests
// =============================================================================

func TestAdminLoginResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := AdminLoginResponse{
			User:         nil,
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresIn:    3600,
		}

		assert.Equal(t, "access_token", resp.AccessToken)
		assert.Equal(t, "refresh_token", resp.RefreshToken)
		assert.Equal(t, int64(3600), resp.ExpiresIn)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := AdminLoginResponse{
			User:         nil,
			AccessToken:  "token123",
			RefreshToken: "refresh123",
			ExpiresIn:    1800,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"access_token":"token123"`)
		assert.Contains(t, string(data), `"refresh_token":"refresh123"`)
		assert.Contains(t, string(data), `"expires_in":1800`)
	})
}

// =============================================================================
// InitialSetup Handler Tests
// =============================================================================

func TestInitialSetup_RequestValidation(t *testing.T) {
	t.Run("request struct can be created", func(t *testing.T) {
		req := InitialSetupRequest{
			Email:    "admin@example.com",
			Password: "securepassword123",
		}
		assert.Equal(t, "admin@example.com", req.Email)
		assert.Equal(t, "securepassword123", req.Password)
	})

	t.Run("handler can be constructed with nil deps", func(t *testing.T) {
		// Handler construction should not panic
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)
		assert.NotNil(t, handler)
	})
}

// =============================================================================
// AdminLogin Handler Tests
// =============================================================================

func TestAdminLogin_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New(fiber.Config{ErrorHandler: customErrorHandler})
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Post("/login", handler.AdminLogin)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("not json")))
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

	t.Run("valid JSON structure can be parsed", func(t *testing.T) {
		body := `{"email":"admin@example.com","password":"AdminPass123!"}`
		var req AdminLoginRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.Equal(t, "admin@example.com", req.Email)
		assert.Equal(t, "AdminPass123!", req.Password)
	})
}

// =============================================================================
// AdminRefreshToken Handler Tests
// =============================================================================

func TestAdminRefreshToken_Validation(t *testing.T) {
	t.Run("invalid request body", func(t *testing.T) {
		app := fiber.New(fiber.Config{ErrorHandler: customErrorHandler})
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Post("/refresh", handler.AdminRefreshToken)

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})

	t.Run("valid JSON structure can be parsed", func(t *testing.T) {
		body := `{"refresh_token":"some-refresh-token"}`
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)
		assert.Equal(t, "some-refresh-token", req.RefreshToken)
	})
}

// =============================================================================
// AdminLogout Handler Tests
// =============================================================================

func TestAdminLogout_Validation(t *testing.T) {
	t.Run("missing authorization header", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Post("/logout", handler.AdminLogout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Missing authentication")
	})

	t.Run("invalid authorization header format - no Bearer", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Post("/logout", handler.AdminLogout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "InvalidFormat token123")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Invalid authorization header")
	})

	t.Run("invalid authorization header format - single part", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Post("/logout", handler.AdminLogout)

		req := httptest.NewRequest(http.MethodPost, "/logout", nil)
		req.Header.Set("Authorization", "token123")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("valid Bearer token format parsing", func(t *testing.T) {
		authHeader := "Bearer valid-token-123"
		parts := strings.Split(authHeader, " ")
		require.Len(t, parts, 2)
		assert.Equal(t, "Bearer", parts[0])
		assert.Equal(t, "valid-token-123", parts[1])
	})
}

// =============================================================================
// GetCurrentAdmin Handler Tests
// =============================================================================

func TestGetCurrentAdmin_Authorization(t *testing.T) {
	t.Run("no user_id in context", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		app.Get("/me", handler.GetCurrentAdmin)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "not authenticated")
	})

	t.Run("user_id present but non-admin role", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		// Middleware to set user context
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_email", "user@example.com")
			c.Locals("user_role", "user") // Not admin
			return c.Next()
		})

		app.Get("/me", handler.GetCurrentAdmin)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.Contains(t, result["error"], "Admin role required")
	})

	t.Run("authenticated admin user", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		// Middleware to set admin user context
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_email", "admin@example.com")
			c.Locals("user_role", "admin")
			return c.Next()
		})

		app.Get("/me", handler.GetCurrentAdmin)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		user, ok := result["user"].(map[string]interface{})
		require.True(t, ok, "Expected user object in response")

		assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", user["id"])
		assert.Equal(t, "admin@example.com", user["email"])
		assert.Equal(t, "admin", user["role"])
	})

	t.Run("authenticated admin with empty email", func(t *testing.T) {
		app := newTestApp(t)
		handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

		// Middleware to set admin user context with empty email
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
			c.Locals("user_email", "")
			c.Locals("user_role", "admin")
			return c.Next()
		})

		app.Get("/me", handler.GetCurrentAdmin)

		req := httptest.NewRequest(http.MethodGet, "/me", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		// Should still succeed - email is optional in the response
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

// =============================================================================
// Authorization Header Parsing Tests
// =============================================================================

func TestAuthorizationHeaderParsing(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "missing header",
			authHeader:     "",
			expectError:    true,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "Basic auth instead of Bearer",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectError:    true,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "bearer lowercase",
			authHeader:     "bearer token123",
			expectError:    true,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "extra spaces",
			authHeader:     "Bearer  token123",
			expectError:    true,
			expectedStatus: fiber.StatusUnauthorized,
		},
		{
			name:           "token only",
			authHeader:     "token123",
			expectError:    true,
			expectedStatus: fiber.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := newTestApp(t)
			handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

			app.Post("/logout", handler.AdminLogout)

			req := httptest.NewRequest(http.MethodPost, "/logout", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

// =============================================================================
// Role Validation Tests
// =============================================================================

func TestRoleValidation(t *testing.T) {
	validAdminRoles := []string{"admin"}
	nonAdminRoles := []string{"user", "service_role", "anon", "authenticated", "dashboard_user", ""}

	t.Run("admin roles are allowed", func(t *testing.T) {
		for _, role := range validAdminRoles {
			app := newTestApp(t)
			handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

			app.Use(func(c fiber.Ctx) error {
				c.Locals("user_id", "user-123")
				c.Locals("user_email", "user@test.com")
				c.Locals("user_role", role)
				return c.Next()
			})

			app.Get("/me", handler.GetCurrentAdmin)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Role %q should be allowed", role)
		}
	})

	t.Run("non-admin roles are denied", func(t *testing.T) {
		for _, role := range nonAdminRoles {
			app := newTestApp(t)
			handler := NewAdminAuthHandler(nil, nil, nil, nil, nil)

			app.Use(func(c fiber.Ctx) error {
				c.Locals("user_id", "user-123")
				c.Locals("user_email", "user@test.com")
				c.Locals("user_role", role)
				return c.Next()
			})

			app.Get("/me", handler.GetCurrentAdmin)

			req := httptest.NewRequest(http.MethodGet, "/me", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, fiber.StatusForbidden, resp.StatusCode, "Role %q should be denied", role)
		}
	})
}
