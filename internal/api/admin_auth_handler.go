package api

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	apperrors "github.com/nimbleflux/fluxbase/internal/errors"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// AdminAuthHandler handles admin-specific authentication
type AdminAuthHandler struct {
	authService    *auth.Service
	userRepo       *auth.UserRepository
	dashboardAuth  *auth.DashboardAuthService
	systemSettings *auth.SystemSettingsService
	config         *config.Config
}

// adminRefreshCookieName is the HttpOnly cookie that carries the admin
// refresh token so it is never exposed to JavaScript. It is scoped to the
// admin endpoints; the JSON body field remains supported for clients that
// do not send cookies.
const adminRefreshCookieName = "fluxbase_admin_refresh"

// adminRefreshCookieMaxAge bounds the cookie lifetime. The refresh token's
// own server-side expiry governs actual validity; an over-long cookie is
// harmless because expired tokens are rejected on refresh.
const adminRefreshCookieMaxAge = 7 * 24 * 60 * 60

// secureRequest reports whether the request arrived over HTTPS, either
// directly or through a proxy that set X-Forwarded-Proto.
func secureRequest(c fiber.Ctx) bool {
	return c.Protocol() == "https" || strings.EqualFold(c.Get("X-Forwarded-Proto"), "https")
}

// setAdminRefreshCookie stores the refresh token in an HttpOnly cookie scoped
// to the admin endpoints.
func setAdminRefreshCookie(c fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     adminRefreshCookieName,
		Value:    refreshToken,
		Path:     "/api/v1/admin",
		MaxAge:   adminRefreshCookieMaxAge,
		Secure:   secureRequest(c),
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// clearAdminRefreshCookie expires the refresh token cookie.
func clearAdminRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     adminRefreshCookieName,
		Value:    "",
		Path:     "/api/v1/admin",
		MaxAge:   -1,
		Secure:   secureRequest(c),
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// NewAdminAuthHandler creates a new admin auth handler
func NewAdminAuthHandler(
	authService *auth.Service,
	userRepo *auth.UserRepository,
	dashboardAuth *auth.DashboardAuthService,
	systemSettings *auth.SystemSettingsService,
	cfg *config.Config,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		authService:    authService,
		userRepo:       userRepo,
		dashboardAuth:  dashboardAuth,
		systemSettings: systemSettings,
		config:         cfg,
	}
}

func (h *AdminAuthHandler) requireService(c fiber.Ctx) error {
	if h.systemSettings == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Settings service not initialized")
	}
	return nil
}

func (h *AdminAuthHandler) requireDashboardAuth(c fiber.Ctx) error {
	if h.dashboardAuth == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Dashboard auth service not initialized")
	}
	return nil
}

func (h *AdminAuthHandler) requireAuthService(c fiber.Ctx) error {
	if h.authService == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "Auth service not initialized")
	}
	return nil
}

// SetupStatusResponse represents the setup status
type SetupStatusResponse struct {
	NeedsSetup bool `json:"needs_setup"`
	HasAdmin   bool `json:"has_admin"`
}

// InitialSetupRequest represents the initial setup request
type InitialSetupRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Name       string `json:"name"`
	SetupToken string `json:"setup_token"`
}

// InitialSetupResponse represents the initial setup response
type InitialSetupResponse struct {
	User         *auth.DashboardUser `json:"user"`
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int64               `json:"expires_in"`
}

// AdminLoginRequest represents an admin login request
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminLoginResponse represents an admin login response
type AdminLoginResponse struct {
	User         *auth.DashboardUser `json:"user"`
	AccessToken  string              `json:"access_token"`
	RefreshToken string              `json:"refresh_token"`
	ExpiresIn    int64               `json:"expires_in"`
}

// GetSetupStatus checks if initial setup is needed
// GET /api/v1/admin/setup/status
func (h *AdminAuthHandler) GetSetupStatus(c fiber.Ctx) error {
	ctx := c.RequestCtx()

	if err := h.requireService(c); err != nil {
		return err
	}

	setupComplete, err := h.systemSettings.IsSetupComplete(ctx)
	if err != nil {
		// If settings table doesn't exist yet (e.g., during bootstrap),
		// treat as needing setup rather than returning a 500.
		log.Debug().Err(err).Msg("Failed to check setup status, assuming setup needed")
		return c.JSON(SetupStatusResponse{
			NeedsSetup: true,
			HasAdmin:   false,
		})
	}

	return c.JSON(SetupStatusResponse{
		NeedsSetup: !setupComplete,
		HasAdmin:   setupComplete,
	})
}

// InitialSetup creates the first admin user
// POST /api/v1/admin/setup
func (h *AdminAuthHandler) InitialSetup(c fiber.Ctx) error {
	log.Debug().Str("path", c.Path()).Str("method", c.Method()).Msg("InitialSetup handler called")

	ctx := c.RequestCtx()

	if err := h.requireService(c); err != nil {
		return err
	}

	if err := h.requireDashboardAuth(c); err != nil {
		return err
	}

	setupComplete, err := h.systemSettings.IsSetupComplete(ctx)
	if err != nil {
		return SendOperationFailed(c, "check setup status")
	}

	if setupComplete {
		return SendForbidden(c, "Setup has already been completed", ErrCodeSetupCompleted)
	}

	// Parse request
	var req InitialSetupRequest
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	// Validate setup token using constant-time comparison to prevent timing attacks
	configuredToken := h.config.Security.SetupToken
	if configuredToken == "" {
		return SendForbidden(c, "Admin setup is disabled. Set FLUXBASE_SECURITY_SETUP_TOKEN to enable.", ErrCodeSetupDisabled)
	}

	if req.SetupToken == "" {
		return SendMissingField(c, "setup_token")
	}

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(req.SetupToken), []byte(configuredToken)) != 1 {
		return SendUnauthorized(c, "Invalid setup token", ErrCodeInvalidSetupToken)
	}

	// Validate password strength
	if err := auth.ValidateDashboardPassword(req.Password); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	// Create the first dashboard admin user
	user, err := h.dashboardAuth.CreateUser(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		return SendInternalError(c, fmt.Sprintf("Failed to create admin user: %v", err))
	}

	_, err = h.dashboardAuth.GetDB().Exec(ctx, `
		UPDATE platform.users
		SET role = 'instance_admin', email_verified = true
		WHERE id = $1
	`, user.ID)
	if err != nil {
		return SendOperationFailed(c, "set admin role and verify email")
	}

	var defaultTenantID string
	err = h.dashboardAuth.GetDB().QueryRow(ctx, `SELECT id FROM platform.tenants WHERE is_default = true LIMIT 1`).Scan(&defaultTenantID)
	if err == nil {
		_, _ = h.dashboardAuth.GetDB().Exec(ctx, `
			INSERT INTO platform.tenant_admin_assignments (user_id, tenant_id, assigned_by)
			VALUES ($1, $2, $1)
			ON CONFLICT (user_id, tenant_id) DO NOTHING
		`, user.ID, defaultTenantID)
	}

	// Mark setup as complete in system settings
	if err := h.systemSettings.MarkSetupComplete(ctx, user.ID, user.Email); err != nil {
		return SendOperationFailed(c, "mark setup as complete")
	}

	// Log in the user to get access token
	loggedInUser, loginResp, err := h.dashboardAuth.Login(ctx, req.Email, req.Password, nil, c.Get("User-Agent"))
	if err != nil {
		return SendInternalError(c, "User created but failed to generate access token")
	}

	return c.Status(http.StatusCreated).JSON(InitialSetupResponse{
		User:         loggedInUser,
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
	})
}

// AdminLogin authenticates an admin user
// POST /api/v1/admin/login
func (h *AdminAuthHandler) AdminLogin(c fiber.Ctx) error {
	ctx := c.RequestCtx()

	var req AdminLoginRequest
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if err := h.requireDashboardAuth(c); err != nil {
		return err
	}

	user, loginResp, err := h.dashboardAuth.Login(ctx, req.Email, req.Password, nil, c.Get("User-Agent"))
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return SendUnauthorized(c, "Invalid email or password", ErrCodeInvalidCredentials)
		}
		if errors.Is(err, auth.ErrAccountLocked) {
			return SendForbidden(c, "Account is locked due to too many failed login attempts", ErrCodeAccountLocked)
		}
		return SendInternalError(c, fmt.Sprintf("Authentication failed: %v", err))
	}

	// Query user's role from database (DashboardUser struct doesn't include role)
	var userRole string
	err = h.dashboardAuth.GetDB().QueryRow(
		ctx,
		"SELECT role FROM platform.users WHERE id = $1",
		user.ID,
	).Scan(&userRole)
	if err != nil {
		return SendOperationFailed(c, "verify user role")
	}

	// Check if user has instance_admin role
	if userRole != "instance_admin" {
		return SendForbidden(c, "Access denied. Admin role required.", ErrCodeAdminRequired)
	}

	// Also set the refresh token as an HttpOnly cookie; the JSON response
	// keeps returning it for clients that do not use cookies.
	setAdminRefreshCookie(c, loginResp.RefreshToken)

	return c.JSON(AdminLoginResponse{
		User:         user,
		AccessToken:  loginResp.AccessToken,
		RefreshToken: loginResp.RefreshToken,
		ExpiresIn:    loginResp.ExpiresIn,
	})
}

// AdminRefreshToken refreshes an admin's access token
// POST /api/v1/admin/refresh
func (h *AdminAuthHandler) AdminRefreshToken(c fiber.Ctx) error {
	ctx := c.RequestCtx()

	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if err := h.requireDashboardAuth(c); err != nil {
		return err
	}

	// Prefer the HttpOnly refresh cookie, falling back to the JSON body field
	// for backward compatibility with clients that do not send cookies.
	refreshToken := c.Cookies(adminRefreshCookieName)
	if refreshToken == "" {
		refreshToken = req.RefreshToken
	}

	refreshResp, err := h.dashboardAuth.RefreshToken(ctx, refreshToken)
	if err != nil {
		return SendUnauthorized(c, "Invalid or expired refresh token", ErrCodeInvalidToken)
	}

	// Keep the rotated refresh token in the HttpOnly cookie; the JSON response
	// keeps returning it for clients that do not use cookies.
	setAdminRefreshCookie(c, refreshResp.RefreshToken)

	// Validate the new access token to get user ID
	claims, err := h.dashboardAuth.ValidateToken(refreshResp.AccessToken)
	if err != nil {
		return SendUnauthorized(c, "Failed to validate refreshed token", ErrCodeInvalidToken)
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return SendUnauthorized(c, "Invalid user ID in token", ErrCodeInvalidToken)
	}

	// Fetch platform user details (not auth.users)
	user, err := h.dashboardAuth.GetUserByID(ctx, userID)
	if err != nil {
		return SendOperationFailed(c, "fetch user")
	}

	// Verify user still has a valid dashboard role
	validRoles := map[string]bool{
		"instance_admin": true,
		"tenant_admin":   true,
	}
	if !validRoles[user.Role] {
		return SendAdminRequired(c)
	}

	return c.JSON(fiber.Map{
		"access_token":  refreshResp.AccessToken,
		"refresh_token": refreshResp.RefreshToken,
		"expires_in":    refreshResp.ExpiresIn,
		"user":          user,
	})
}

// AdminLogout logs out an admin user
// POST /api/v1/admin/logout
func (h *AdminAuthHandler) AdminLogout(c fiber.Ctx) error {
	ctx := c.RequestCtx()

	// Get the access token from the Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return SendMissingAuth(c)
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return SendUnauthorized(c, "Invalid authorization header format", ErrCodeInvalidFormat)
	}

	token := parts[1]

	if err := h.requireAuthService(c); err != nil {
		return err
	}

	if err := h.authService.SignOut(ctx, token); err != nil {
		return SendOperationFailed(c, "logout")
	}

	clearAdminRefreshCookie(c)

	return apperrors.SendSuccess(c, "Logged out successfully")
}

// GetCurrentAdmin returns the currently authenticated admin user
// GET /api/v1/admin/me
func (h *AdminAuthHandler) GetCurrentAdmin(c fiber.Ctx) error {
	// Get user info from context (set by auth middleware)
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendUnauthorized(c, "User not authenticated", ErrCodeAuthRequired)
	}

	userEmail, _ := c.Locals("user_email").(string)
	userRole, _ := c.Locals("user_role").(string)

	// Verify admin role
	if userRole != "admin" && userRole != "tenant_service" {
		return SendAdminRequired(c)
	}

	// Return user info from JWT claims (sufficient for UI needs)
	// We could fetch full user from DB but JWT claims have what we need
	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id":    userID,
			"email": userEmail,
			"role":  userRole,
		},
	})
}
