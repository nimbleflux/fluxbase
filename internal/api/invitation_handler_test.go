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
// InvitationHandler Construction Tests
// =============================================================================

func TestNewInvitationHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewInvitationHandler(nil, nil, nil, "https://example.com")
		assert.NotNil(t, handler)
		assert.Nil(t, handler.invitationService)
		assert.Nil(t, handler.dashboardAuth)
		assert.Nil(t, handler.emailService)
		assert.Equal(t, "https://example.com", handler.baseURL)
	})

	t.Run("creates handler with custom base URL", func(t *testing.T) {
		handler := NewInvitationHandler(nil, nil, nil, "http://localhost:3000")
		assert.Equal(t, "http://localhost:3000", handler.baseURL)
	})
}

// =============================================================================
// CreateInvitationRequest Tests
// =============================================================================

func TestCreateInvitationRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := CreateInvitationRequest{
			Email:          "test@example.com",
			Role:           "instance_admin",
			ExpiryDuration: 604800, // 7 days in seconds
		}

		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "instance_admin", req.Role)
		assert.Equal(t, int64(604800), req.ExpiryDuration)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"email":"user@test.com","role":"dashboard_user","expiry_duration":86400}`

		var req CreateInvitationRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "user@test.com", req.Email)
		assert.Equal(t, "dashboard_user", req.Role)
		assert.Equal(t, int64(86400), req.ExpiryDuration)
	})

	t.Run("default expiry duration is zero", func(t *testing.T) {
		req := CreateInvitationRequest{
			Email: "test@example.com",
			Role:  "instance_admin",
		}

		assert.Equal(t, int64(0), req.ExpiryDuration)
	})
}

// =============================================================================
// CreateInvitationResponse Tests
// =============================================================================

func TestCreateInvitationResponse_Struct(t *testing.T) {
	t.Run("successful response with email sent", func(t *testing.T) {
		resp := CreateInvitationResponse{
			Invitation:  nil, // Would be a real invitation in production
			InviteLink:  "https://example.com/invite/abc123",
			EmailSent:   true,
			EmailStatus: "Invitation email sent successfully",
		}

		assert.Equal(t, "https://example.com/invite/abc123", resp.InviteLink)
		assert.True(t, resp.EmailSent)
		assert.Equal(t, "Invitation email sent successfully", resp.EmailStatus)
	})

	t.Run("response without email service", func(t *testing.T) {
		resp := CreateInvitationResponse{
			InviteLink:  "https://example.com/invite/xyz789",
			EmailSent:   false,
			EmailStatus: "Email service not configured. Share the invite link manually.",
		}

		assert.False(t, resp.EmailSent)
		assert.Contains(t, resp.EmailStatus, "not configured")
	})
}

// =============================================================================
// ValidateInvitationResponse Tests
// =============================================================================

func TestValidateInvitationResponse_Struct(t *testing.T) {
	t.Run("valid invitation", func(t *testing.T) {
		resp := ValidateInvitationResponse{
			Valid:      true,
			Invitation: nil,
		}

		assert.True(t, resp.Valid)
		assert.Empty(t, resp.Error)
	})

	t.Run("invalid invitation - expired", func(t *testing.T) {
		resp := ValidateInvitationResponse{
			Valid: false,
			Error: "Invitation has expired",
		}

		assert.False(t, resp.Valid)
		assert.Equal(t, "Invitation has expired", resp.Error)
	})

	t.Run("invalid invitation - already accepted", func(t *testing.T) {
		resp := ValidateInvitationResponse{
			Valid: false,
			Error: "Invitation has already been accepted",
		}

		assert.False(t, resp.Valid)
		assert.Contains(t, resp.Error, "already been accepted")
	})

	t.Run("invalid invitation - not found", func(t *testing.T) {
		resp := ValidateInvitationResponse{
			Valid: false,
			Error: "Invitation not found",
		}

		assert.False(t, resp.Valid)
		assert.Equal(t, "Invitation not found", resp.Error)
	})
}

// =============================================================================
// AcceptInvitationRequest Tests
// =============================================================================

func TestAcceptInvitationRequest_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		req := AcceptInvitationRequest{
			Password: "SecurePassword123!",
			Name:     "John Doe",
		}

		assert.Equal(t, "SecurePassword123!", req.Password)
		assert.Equal(t, "John Doe", req.Name)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"password":"MySecretPass!123","name":"Jane Smith"}`

		var req AcceptInvitationRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "MySecretPass!123", req.Password)
		assert.Equal(t, "Jane Smith", req.Name)
	})
}

// =============================================================================
// AcceptInvitationResponse Tests
// =============================================================================

func TestAcceptInvitationResponse_Struct(t *testing.T) {
	t.Run("successful acceptance", func(t *testing.T) {
		resp := AcceptInvitationResponse{
			User:         nil, // Would be a real user in production
			AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			RefreshToken: "refresh_token_value",
			ExpiresIn:    3600,
		}

		assert.NotEmpty(t, resp.AccessToken)
		assert.NotEmpty(t, resp.RefreshToken)
		assert.Equal(t, int64(3600), resp.ExpiresIn)
	})
}

// =============================================================================
// ValidateInvitation Handler Tests
// =============================================================================

func TestValidateInvitation_EmptyToken(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	app.Get("/invitations/:token/validate", handler.ValidateInvitation)

	// Test with empty token parameter
	req := httptest.NewRequest(http.MethodGet, "/invitations//validate", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Fiber treats empty param as route not found
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// AcceptInvitation Handler Tests
// =============================================================================

func TestAcceptInvitation_EmptyToken(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	app.Post("/invitations/:token/accept", handler.AcceptInvitation)

	body := `{"password":"Test123!@#abc","name":"Test User"}`
	req := httptest.NewRequest(http.MethodPost, "/invitations//accept", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Fiber treats empty param as route not found
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestAcceptInvitation_InvalidBody(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	app.Post("/invitations/:token/accept", handler.AcceptInvitation)

	req := httptest.NewRequest(http.MethodPost, "/invitations/valid-token/accept", bytes.NewReader([]byte("invalid json")))
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
// RevokeInvitation Handler Tests
// =============================================================================

func TestRevokeInvitation_EmptyToken(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	app.Delete("/admin/invitations/:token", handler.RevokeInvitation)

	req := httptest.NewRequest(http.MethodDelete, "/admin/invitations/", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Fiber treats empty param as route not found
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// =============================================================================
// CreateInvitation Handler Tests
// =============================================================================

func TestCreateInvitation_NoUserID(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	app.Post("/admin/invitations", handler.CreateInvitation)

	body := `{"email":"test@example.com","role":"instance_admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/invitations", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

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
}

func TestCreateInvitation_InvalidBody(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	// Middleware to set user_id
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
		return c.Next()
	})

	app.Post("/admin/invitations", handler.CreateInvitation)

	req := httptest.NewRequest(http.MethodPost, "/admin/invitations", bytes.NewReader([]byte("invalid")))
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

func TestCreateInvitation_InvalidUserID(t *testing.T) {
	app := newTestApp(t)
	handler := NewInvitationHandler(nil, nil, nil, "https://example.com")

	// Middleware to set invalid user_id
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user_id", "not-a-valid-uuid")
		return c.Next()
	})

	app.Post("/admin/invitations", handler.CreateInvitation)

	body := `{"email":"test@example.com","role":"instance_admin"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/invitations", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	require.NoError(t, err)

	assert.Contains(t, result["error"], "Invalid user ID")
}

// =============================================================================
// JSON Serialization Tests
// =============================================================================

func TestInvitationResponses_JSONSerialization(t *testing.T) {
	t.Run("CreateInvitationResponse serializes correctly", func(t *testing.T) {
		resp := CreateInvitationResponse{
			InviteLink:  "https://example.com/invite/token123",
			EmailSent:   true,
			EmailStatus: "Sent successfully",
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"invite_link"`)
		assert.Contains(t, string(data), `"email_sent":true`)
		assert.Contains(t, string(data), `"email_status"`)
	})

	t.Run("ValidateInvitationResponse serializes correctly", func(t *testing.T) {
		resp := ValidateInvitationResponse{
			Valid: true,
			Error: "",
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"valid":true`)
	})

	t.Run("AcceptInvitationResponse serializes correctly", func(t *testing.T) {
		resp := AcceptInvitationResponse{
			AccessToken:  "token",
			RefreshToken: "refresh",
			ExpiresIn:    3600,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"access_token"`)
		assert.Contains(t, string(data), `"refresh_token"`)
		assert.Contains(t, string(data), `"expires_in":3600`)
	})
}
