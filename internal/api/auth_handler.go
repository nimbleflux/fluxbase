package api

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// Cookie names for authentication tokens
const (
	AccessTokenCookieName  = "fluxbase_access_token"
	RefreshTokenCookieName = "fluxbase_refresh_token"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	db                  *pgxpool.Pool
	authService         *auth.Service
	captchaService      *auth.CaptchaService
	captchaTrustService *auth.CaptchaTrustService
	samlService         *auth.SAMLService
	baseURL             string
	secureCookie        bool // Whether to set Secure flag on cookies (true in production)
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *pgxpool.Pool, authService *auth.Service, captchaService *auth.CaptchaService, baseURL string) *AuthHandler {
	return &AuthHandler{
		db:             db,
		authService:    authService,
		captchaService: captchaService,
		baseURL:        baseURL,
		secureCookie:   false, // Will be set based on environment
	}
}

// SetSAMLService sets the SAML service for SLO integration
func (h *AuthHandler) SetSAMLService(samlService *auth.SAMLService) {
	h.samlService = samlService
}

// SetSecureCookie sets whether cookies should have the Secure flag
func (h *AuthHandler) SetSecureCookie(secure bool) {
	h.secureCookie = secure
}

// SetCaptchaTrustService sets the CAPTCHA trust service for adaptive verification
func (h *AuthHandler) SetCaptchaTrustService(trustService *auth.CaptchaTrustService) {
	h.captchaTrustService = trustService
}

// AuthConfigResponse represents the public authentication configuration
type AuthConfigResponse struct {
	SignupEnabled            bool                        `json:"signup_enabled"`
	RequireEmailVerification bool                        `json:"require_email_verification"`
	MagicLinkEnabled         bool                        `json:"magic_link_enabled"`
	PasswordLoginEnabled     bool                        `json:"password_login_enabled"`
	MFAAvailable             bool                        `json:"mfa_available"`
	PasswordMinLength        int                         `json:"password_min_length"`
	PasswordRequireUppercase bool                        `json:"password_require_uppercase"`
	PasswordRequireLowercase bool                        `json:"password_require_lowercase"`
	PasswordRequireNumber    bool                        `json:"password_require_number"`
	PasswordRequireSpecial   bool                        `json:"password_require_special"`
	OAuthProviders           []OAuthProviderPublic       `json:"oauth_providers"`
	SAMLProviders            []SAMLProviderPublic        `json:"saml_providers"`
	Captcha                  *auth.CaptchaConfigResponse `json:"captcha"`
}

// OAuthProviderPublic represents public OAuth provider information
type OAuthProviderPublic struct {
	Provider     string `json:"provider"`
	DisplayName  string `json:"display_name"`
	AuthorizeURL string `json:"authorize_url"`
}

// SAMLProviderPublic represents public SAML provider information
type SAMLProviderPublic struct {
	Provider    string `json:"provider"`
	DisplayName string `json:"display_name"`
}

// setAuthCookies sets httpOnly cookies for access and refresh tokens
func (h *AuthHandler) setAuthCookies(c fiber.Ctx, accessToken, refreshToken string, expiresIn int64) {
	// Access token cookie - shorter expiry
	// SameSite=Lax allows the cookie to be sent during top-level navigation
	// which is required for OAuth authorization flows from external clients
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    accessToken,
		Path:     "/",
		MaxAge:   int(expiresIn), // seconds
		Secure:   h.secureCookie,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	// Refresh token cookie - longer expiry (7 days default)
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/api/v1/auth",   // Only sent to auth endpoints
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		Secure:   h.secureCookie,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// clearAuthCookies removes authentication cookies
func (h *AuthHandler) clearAuthCookies(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Expire immediately
		Secure:   h.secureCookie,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/api/v1/auth",
		MaxAge:   -1,
		Secure:   h.secureCookie,
		HTTPOnly: true,
		SameSite: "Strict",
	})
}

// getAccessToken gets the access token from cookie or Authorization header
func (h *AuthHandler) getAccessToken(c fiber.Ctx) string {
	// First try cookie
	if token := c.Cookies(AccessTokenCookieName); token != "" {
		return token
	}

	// Fall back to Authorization header for API clients
	token := c.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:]
	}
	return token
}

// getRefreshToken gets the refresh token from cookie or request body
func (h *AuthHandler) getRefreshToken(c fiber.Ctx) string {
	// First try cookie
	if token := c.Cookies(RefreshTokenCookieName); token != "" {
		return token
	}
	return ""
}

// SignUp handles user registration
// POST /auth/signup
func (h *AuthHandler) SignUp(c fiber.Ctx) error {
	// Check if signup is enabled
	if !h.authService.IsSignupEnabled() {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "User registration is currently disabled",
			"code":  "SIGNUP_DISABLED",
		})
	}

	var req auth.SignUpRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse signup request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// CAPTCHA verification with adaptive trust support
	captchaVerified := false
	if h.captchaService != nil && h.captchaService.IsEnabled() {
		// If challenge_id is provided, validate the challenge first
		if req.ChallengeID != "" && h.captchaTrustService != nil {
			// Verify CAPTCHA token if one was provided
			if req.CaptchaToken != "" {
				if err := h.captchaService.Verify(middleware.CtxWithTenant(c), req.CaptchaToken, c.IP()); err != nil {
					log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for signup")
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification failed",
						"code":  "CAPTCHA_INVALID",
					})
				}
				captchaVerified = true
			}

			// Validate the challenge (checks if CAPTCHA was required and if it was verified)
			if err := h.captchaTrustService.ValidateChallenge(middleware.CtxWithTenant(c), req.ChallengeID, "signup", c.IP(), captchaVerified); err != nil {
				if errors.Is(err, auth.ErrCaptchaRequired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification required",
						"code":  "CAPTCHA_REQUIRED",
					})
				}
				if errors.Is(err, auth.ErrChallengeExpired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Challenge expired, please request a new one",
						"code":  "CHALLENGE_EXPIRED",
					})
				}
				if errors.Is(err, auth.ErrChallengeConsumed) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Challenge already used, please request a new one",
						"code":  "CHALLENGE_CONSUMED",
					})
				}
				log.Warn().Err(err).Str("email", req.Email).Msg("Challenge validation failed for signup")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid challenge",
					"code":  "CHALLENGE_INVALID",
				})
			}
		} else {
			// Fall back to static CAPTCHA verification (no challenge_id provided)
			if err := h.captchaService.VerifyForEndpoint(middleware.CtxWithTenant(c), "signup", req.CaptchaToken, c.IP()); err != nil {
				if errors.Is(err, auth.ErrCaptchaRequired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification required",
						"code":  "CAPTCHA_REQUIRED",
					})
				}
				log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for signup")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "CAPTCHA verification failed",
					"code":  "CAPTCHA_INVALID",
				})
			}
			captchaVerified = req.CaptchaToken != ""
		}
	}

	// Validate required fields
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}
	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password is required",
		})
	}

	// Create user
	resp, err := h.authService.SignUp(middleware.CtxWithTenant(c), req)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to sign up user")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Issue trust token if CAPTCHA was verified (for use in subsequent requests)
	var trustToken string
	if captchaVerified && h.captchaTrustService != nil && h.captchaTrustService.IsEnabled() {
		trustToken, _ = h.captchaTrustService.IssueTrustToken(middleware.CtxWithTenant(c), c.IP(), req.DeviceFingerprint, c.Get("User-Agent"))
	}

	// Check if email verification is required (don't set cookies, no tokens returned)
	if resp.RequiresEmailVerification {
		response := fiber.Map{
			"user":                        resp.User,
			"requires_email_verification": true,
			"message":                     "Please check your email to verify your account before signing in.",
		}
		if trustToken != "" {
			response["trust_token"] = trustToken
		}
		return c.Status(fiber.StatusCreated).JSON(response)
	}

	// Set httpOnly cookies for tokens
	h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	// Add trust token to response if available
	if trustToken != "" {
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"user":          resp.User,
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
			"trust_token":   trustToken,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// SignIn handles user login
// POST /auth/signin
func (h *AuthHandler) SignIn(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	// Check if password login is disabled for app users
	if h.isPasswordLoginDisabled(ctx) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Password login is disabled. Please use an OAuth or SAML provider to sign in.",
			"code":  "PASSWORD_LOGIN_DISABLED",
		})
	}

	var req auth.SignInRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse signin request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// CAPTCHA verification with adaptive trust support
	captchaVerified := false
	if h.captchaService != nil && h.captchaService.IsEnabled() {
		// If challenge_id is provided, validate the challenge first
		if req.ChallengeID != "" && h.captchaTrustService != nil {
			// Verify CAPTCHA token if one was provided
			if req.CaptchaToken != "" {
				if err := h.captchaService.Verify(middleware.CtxWithTenant(c), req.CaptchaToken, c.IP()); err != nil {
					log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for login")
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification failed",
						"code":  "CAPTCHA_INVALID",
					})
				}
				captchaVerified = true
			}

			// Validate the challenge (checks if CAPTCHA was required and if it was verified)
			if err := h.captchaTrustService.ValidateChallenge(middleware.CtxWithTenant(c), req.ChallengeID, "login", c.IP(), captchaVerified); err != nil {
				if errors.Is(err, auth.ErrCaptchaRequired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification required",
						"code":  "CAPTCHA_REQUIRED",
					})
				}
				if errors.Is(err, auth.ErrChallengeExpired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Challenge expired, please request a new one",
						"code":  "CHALLENGE_EXPIRED",
					})
				}
				if errors.Is(err, auth.ErrChallengeConsumed) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "Challenge already used, please request a new one",
						"code":  "CHALLENGE_CONSUMED",
					})
				}
				log.Warn().Err(err).Str("email", req.Email).Msg("Challenge validation failed for login")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid challenge",
					"code":  "CHALLENGE_INVALID",
				})
			}
		} else {
			// Fall back to static CAPTCHA verification (no challenge_id provided)
			if err := h.captchaService.VerifyForEndpoint(middleware.CtxWithTenant(c), "login", req.CaptchaToken, c.IP()); err != nil {
				if errors.Is(err, auth.ErrCaptchaRequired) {
					return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
						"error": "CAPTCHA verification required",
						"code":  "CAPTCHA_REQUIRED",
					})
				}
				log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for login")
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "CAPTCHA verification failed",
					"code":  "CAPTCHA_INVALID",
				})
			}
			captchaVerified = req.CaptchaToken != ""
		}
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Authenticate user
	resp, err := h.authService.SignIn(middleware.CtxWithTenant(c), req)
	if err != nil {
		// Record failed attempt for trust tracking
		if h.captchaTrustService != nil {
			_ = h.captchaTrustService.RecordFailedAttempt(ctx, nil, c.IP(), req.DeviceFingerprint, c.Get("User-Agent"))
		}

		// Check for locked account
		if errors.Is(err, auth.ErrAccountLocked) {
			log.Warn().Str("email", req.Email).Msg("Login attempt on locked account")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Account locked due to too many failed login attempts. Please contact support.",
				"code":  "ACCOUNT_LOCKED",
			})
		}
		// Check for email not verified
		if errors.Is(err, auth.ErrEmailNotVerified) {
			log.Warn().Str("email", req.Email).Msg("Login attempt with unverified email")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":                       "Please verify your email address before signing in. Check your inbox for the verification link.",
				"code":                        "EMAIL_NOT_VERIFIED",
				"requires_email_verification": true,
			})
		}
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to sign in user")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	// Record successful login for trust tracking
	if h.captchaTrustService != nil {
		if userUUID, err := uuid.Parse(resp.User.ID); err == nil {
			_ = h.captchaTrustService.RecordSuccessfulLogin(ctx, userUUID, c.IP(), req.DeviceFingerprint, c.Get("User-Agent"))
			if captchaVerified {
				_ = h.captchaTrustService.RecordCaptchaSolved(ctx, &userUUID, c.IP(), req.DeviceFingerprint, c.Get("User-Agent"))
			}
		}
	}

	// Issue trust token if CAPTCHA was verified (for use in subsequent requests)
	var trustToken string
	if captchaVerified && h.captchaTrustService != nil && h.captchaTrustService.IsEnabled() {
		trustToken, _ = h.captchaTrustService.IssueTrustToken(ctx, c.IP(), req.DeviceFingerprint, c.Get("User-Agent"))
	}

	// Check if user has 2FA enabled
	twoFAEnabled, err := h.authService.IsTOTPEnabled(middleware.CtxWithTenant(c), resp.User.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", resp.User.ID).Msg("Failed to check 2FA status")
		// Continue with login - don't block if 2FA check fails
		// Set httpOnly cookies for tokens
		h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)
		if trustToken != "" {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"user":          resp.User,
				"access_token":  resp.AccessToken,
				"refresh_token": resp.RefreshToken,
				"expires_in":    resp.ExpiresIn,
				"trust_token":   trustToken,
			})
		}
		return c.Status(fiber.StatusOK).JSON(resp)
	}

	// If 2FA is enabled, return special response requiring 2FA verification
	if twoFAEnabled {
		response := fiber.Map{
			"requires_2fa": true,
			"user_id":      resp.User.ID,
			"message":      "2FA verification required. Please provide your 2FA code.",
		}
		if trustToken != "" {
			response["trust_token"] = trustToken
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	// Set httpOnly cookies for tokens
	h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	// Add trust token to response if available
	if trustToken != "" {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"user":          resp.User,
			"access_token":  resp.AccessToken,
			"refresh_token": resp.RefreshToken,
			"expires_in":    resp.ExpiresIn,
			"trust_token":   trustToken,
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// SignOut handles user logout
// POST /auth/signout
func (h *AuthHandler) SignOut(c fiber.Ctx) error {
	// Get token from cookie or Authorization header
	token := h.getAccessToken(c)
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No authentication token provided",
		})
	}

	ctx := middleware.CtxWithTenant(c)

	// Get user ID from token before signing out
	var userID string
	if claims, err := h.authService.ValidateToken(token); err == nil {
		userID = claims.UserID
	}

	// Check if user has an active SAML session
	var samlLogoutInfo *fiber.Map
	if userID != "" && h.samlService != nil {
		samlSession, err := h.samlService.GetSAMLSessionByUserID(ctx, userID)
		if err == nil && samlSession != nil {
			// Check if provider has SLO support
			idpSloURL, _ := h.samlService.GetIdPSloURL(samlSession.ProviderName)
			if idpSloURL != "" && h.samlService.HasSigningKey(samlSession.ProviderName) {
				// SAML SLO is available - return the logout URL
				samlLogoutInfo = &fiber.Map{
					"saml_logout": true,
					"provider":    samlSession.ProviderName,
					"slo_url":     fmt.Sprintf("/auth/saml/logout/%s", samlSession.ProviderName),
				}
			} else {
				// No SLO support - clean up SAML session locally
				if err := h.samlService.DeleteSAMLSession(ctx, samlSession.ID); err != nil {
					log.Warn().Err(err).Msg("Failed to delete SAML session during signout")
				}
			}
		}
	}

	// Sign out user (invalidates JWT)
	if err := h.authService.SignOut(ctx, token); err != nil {
		log.Error().Err(err).Msg("Failed to sign out user")
		// Clear cookies even if sign out fails
		h.clearAuthCookies(c)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to sign out",
		})
	}

	// Clear authentication cookies
	h.clearAuthCookies(c)

	// Return response with SAML logout info if applicable
	if samlLogoutInfo != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "Successfully signed out locally",
			"saml_logout": (*samlLogoutInfo)["saml_logout"],
			"provider":    (*samlLogoutInfo)["provider"],
			"slo_url":     (*samlLogoutInfo)["slo_url"],
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Successfully signed out",
	})
}

// RefreshToken handles token refresh
// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	var req auth.RefreshTokenRequest
	if err := c.Bind().Body(&req); err != nil {
		// Body parsing failed, try to get refresh token from cookie
		req.RefreshToken = h.getRefreshToken(c)
	}

	// If no refresh token in body, try cookie
	if req.RefreshToken == "" {
		req.RefreshToken = h.getRefreshToken(c)
	}

	// Validate required fields
	if req.RefreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	// Refresh token
	resp, err := h.authService.RefreshToken(middleware.CtxWithTenant(c), req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to refresh token")
		// Clear cookies on refresh failure
		h.clearAuthCookies(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
		})
	}

	// Set httpOnly cookies for new tokens
	h.setAuthCookies(c, resp.AccessToken, resp.RefreshToken, resp.ExpiresIn)

	return c.Status(fiber.StatusOK).JSON(resp)
}

// GetUser handles getting current user profile
// GET /auth/user
func (h *AuthHandler) GetUser(c fiber.Ctx) error {
	// Get token from Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authorization header is required",
		})
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Get user
	user, err := h.authService.GetUser(middleware.CtxWithTenant(c), token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// UpdateUser handles updating user profile
// PATCH /auth/user
func (h *AuthHandler) UpdateUser(c fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req auth.UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse update user request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update user
	user, err := h.authService.UpdateUser(middleware.CtxWithTenant(c), userID.(string), req)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to update user")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

// SendMagicLink handles sending magic link
// POST /auth/magiclink
func (h *AuthHandler) SendMagicLink(c fiber.Ctx) error {
	var req struct {
		Email        string `json:"email"`
		CaptchaToken string `json:"captcha_token,omitempty"`
	}
	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse magic link request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify CAPTCHA if enabled for magic_link
	if h.captchaService != nil {
		if err := h.captchaService.VerifyForEndpoint(middleware.CtxWithTenant(c), "magic_link", req.CaptchaToken, c.IP()); err != nil {
			if errors.Is(err, auth.ErrCaptchaRequired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "CAPTCHA verification required",
					"code":  "CAPTCHA_REQUIRED",
				})
			}
			log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for magic link")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "CAPTCHA verification failed",
				"code":  "CAPTCHA_INVALID",
			})
		}
	}

	// Validate email
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Send magic link
	if err := h.authService.SendMagicLink(middleware.CtxWithTenant(c), req.Email); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to send magic link")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return Supabase-compatible OTP response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":    nil,
		"session": nil,
	})
}

// VerifyMagicLink handles magic link verification
// POST /auth/magiclink/verify
func (h *AuthHandler) VerifyMagicLink(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse verify magic link request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate token
	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	// Verify magic link
	resp, err := h.authService.VerifyMagicLink(middleware.CtxWithTenant(c), req.Token)
	if err != nil {
		log.Error().Err(err).Msg("Failed to verify magic link")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// RequestPasswordReset handles password reset requests
// POST /auth/password/reset
func (h *AuthHandler) RequestPasswordReset(c fiber.Ctx) error {
	var req struct {
		Email        string `json:"email"`
		RedirectTo   string `json:"redirect_to,omitempty"`
		CaptchaToken string `json:"captcha_token,omitempty"`
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse password reset request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Verify CAPTCHA if enabled for password_reset
	if h.captchaService != nil {
		if err := h.captchaService.VerifyForEndpoint(middleware.CtxWithTenant(c), "password_reset", req.CaptchaToken, c.IP()); err != nil {
			if errors.Is(err, auth.ErrCaptchaRequired) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "CAPTCHA verification required",
					"code":  "CAPTCHA_REQUIRED",
				})
			}
			log.Warn().Err(err).Str("email", req.Email).Msg("CAPTCHA verification failed for password reset")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "CAPTCHA verification failed",
				"code":  "CAPTCHA_INVALID",
			})
		}
	}

	// Validate email
	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Request password reset (this won't reveal if user exists)
	if err := h.authService.RequestPasswordReset(middleware.CtxWithTenant(c), req.Email, req.RedirectTo); err != nil {
		// Check for SMTP not configured error - this should be returned to the user
		if errors.Is(err, auth.ErrSMTPNotConfigured) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "SMTP is not configured. Please configure an email provider to enable password reset.",
				"code":  "SMTP_NOT_CONFIGURED",
			})
		}
		// Check for invalid redirect URL - return error to prevent misuse
		if errors.Is(err, auth.ErrInvalidRedirectURL) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid redirect_to URL. Must be a valid HTTP or HTTPS URL.",
				"code":  "INVALID_REDIRECT_URL",
			})
		}
		// Check for rate limiting - user requested reset too soon
		if errors.Is(err, auth.ErrPasswordResetTooSoon) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Password reset requested too recently. Please wait 60 seconds before trying again.",
				"code":  "RATE_LIMITED",
			})
		}
		// Check for email sending failure - this should be returned to the user
		if errors.Is(err, auth.ErrEmailSendFailed) {
			log.Error().Err(err).Str("email", req.Email).Msg("Failed to send password reset email")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send password reset email. Please try again later.",
				"code":  "EMAIL_SEND_FAILED",
			})
		}
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to request password reset")
		// Don't reveal if user exists - always return success
	}

	// Return Supabase-compatible OTP response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":    nil,
		"session": nil,
	})
}

// ResetPassword handles password reset with token
// POST /auth/password/reset/confirm
func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse reset password request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}
	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "New password is required",
		})
	}

	// Reset password and get user ID
	userID, err := h.authService.ResetPassword(middleware.CtxWithTenant(c), req.Token, req.NewPassword)
	if err != nil {
		log.Error().Err(err).Msg("Failed to reset password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate new tokens for the user (Supabase-compatible)
	resp, err := h.authService.GenerateTokensForUser(middleware.CtxWithTenant(c), userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens after password reset")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate authentication tokens",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// VerifyPasswordResetToken handles password reset token verification
// POST /auth/password/reset/verify
func (h *AuthHandler) VerifyPasswordResetToken(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}

	if err := c.Bind().Body(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse verify token request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate token
	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	// Verify token
	if err := h.authService.VerifyPasswordResetToken(middleware.CtxWithTenant(c), req.Token); err != nil {
		log.Error().Err(err).Msg("Failed to verify password reset token")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Token is valid",
	})
}

// VerifyEmail verifies a user's email address using a verification token
// POST /auth/verify-email
func (h *AuthHandler) VerifyEmail(c fiber.Ctx) error {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Token is required",
		})
	}

	user, err := h.authService.VerifyEmailToken(middleware.CtxWithTenant(c), req.Token)
	if err != nil {
		// Check for specific token errors
		if errors.Is(err, auth.ErrEmailVerificationTokenNotFound) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid or expired verification token",
				"code":  "INVALID_TOKEN",
			})
		}
		if errors.Is(err, auth.ErrEmailVerificationTokenExpired) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Verification token has expired. Please request a new one.",
				"code":  "TOKEN_EXPIRED",
			})
		}
		if errors.Is(err, auth.ErrEmailVerificationTokenUsed) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "This verification token has already been used",
				"code":  "TOKEN_USED",
			})
		}
		log.Error().Err(err).Msg("Failed to verify email")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Email verified successfully. You can now sign in.",
		"user":    user,
	})
}

// ResendVerificationEmail resends the verification email to a user
// POST /auth/verify-email/resend
func (h *AuthHandler) ResendVerificationEmail(c fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	// Get user by email
	user, err := h.authService.GetUserByEmail(middleware.CtxWithTenant(c), req.Email)
	if err != nil {
		// Don't reveal if email exists - return generic success message
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "If an account exists with this email, a verification link has been sent.",
		})
	}

	// Check if already verified
	if user.EmailVerified {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Email is already verified. You can sign in.",
		})
	}

	// Send verification email
	if err := h.authService.SendEmailVerification(middleware.CtxWithTenant(c), user.ID, user.Email); err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to resend verification email")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send verification email. Please try again later.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Verification email sent. Please check your inbox.",
	})
}

// SignInAnonymous is deprecated and disabled for security reasons
// Anonymous sign-in reduces security by allowing anyone to get tokens
// Use regular signup/signin flow instead
func (h *AuthHandler) SignInAnonymous(c fiber.Ctx) error {
	return c.Status(fiber.StatusGone).JSON(fiber.Map{
		"error": "Anonymous sign-in has been disabled for security reasons",
	})
}

// GetCSRFToken returns the current CSRF token for the client
// Clients should call this endpoint first, then include the token in the X-CSRF-Token header
// GET /auth/csrf
func (h *AuthHandler) GetCSRFToken(c fiber.Ctx) error {
	// The CSRF middleware has already set the cookie
	// Return the token value so clients can use it in the X-CSRF-Token header
	token := c.Cookies("csrf_token")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"csrf_token": token,
	})
}

// StartImpersonation starts an admin impersonation session
func (h *AuthHandler) StartImpersonation(c fiber.Ctx) error {
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req auth.StartImpersonationRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	req.IPAddress = c.IP()
	req.UserAgent = c.Get("User-Agent")

	tenantID := c.Get("X-FB-Tenant")

	resp, err := h.authService.StartImpersonation(middleware.CtxWithTenant(c), adminUserID.(string), tenantID, req)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, auth.ErrNotAdmin) || errors.Is(err, auth.ErrNotTenantAdmin) {
			statusCode = fiber.StatusForbidden
		} else if errors.Is(err, auth.ErrSelfImpersonation) {
			statusCode = fiber.StatusBadRequest
		} else if errors.Is(err, auth.ErrTargetUserNotInTenant) {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// StopImpersonation stops the active impersonation session
func (h *AuthHandler) StopImpersonation(c fiber.Ctx) error {
	// Get admin user ID from context
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	err := h.authService.StopImpersonation(middleware.CtxWithTenant(c), adminUserID.(string))
	if err != nil {
		if errors.Is(err, auth.ErrNoActiveImpersonation) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Impersonation session ended",
	})
}

// GetActiveImpersonation gets the active impersonation session
func (h *AuthHandler) GetActiveImpersonation(c fiber.Ctx) error {
	// Get admin user ID from context
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	session, err := h.authService.GetActiveImpersonation(middleware.CtxWithTenant(c), adminUserID.(string))
	if err != nil {
		if errors.Is(err, auth.ErrNoActiveImpersonation) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(session)
}

// ListImpersonationSessions lists impersonation sessions for audit
func (h *AuthHandler) ListImpersonationSessions(c fiber.Ctx) error {
	// Get admin user ID from context
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	limit := fiber.Query[int](c, "limit", 50)
	offset := fiber.Query[int](c, "offset", 0)

	sessions, err := h.authService.ListImpersonationSessions(middleware.CtxWithTenant(c), adminUserID.(string), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(sessions)
}

// StartAnonImpersonation starts impersonation as anonymous user
func (h *AuthHandler) StartAnonImpersonation(c fiber.Ctx) error {
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Reason is required",
		})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	tenantID := c.Get("X-FB-Tenant")

	resp, err := h.authService.StartAnonImpersonation(middleware.CtxWithTenant(c), adminUserID.(string), tenantID, req.Reason, ipAddress, userAgent)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, auth.ErrNotAdmin) || errors.Is(err, auth.ErrNotTenantAdmin) {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) StartServiceImpersonation(c fiber.Ctx) error {
	adminUserID := c.Locals("user_id")
	if adminUserID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Reason is required",
		})
	}

	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	tenantID := c.Get("X-FB-Tenant")

	resp, err := h.authService.StartServiceImpersonation(middleware.CtxWithTenant(c), adminUserID.(string), tenantID, req.Reason, ipAddress, userAgent)
	if err != nil {
		statusCode := fiber.StatusInternalServerError
		if errors.Is(err, auth.ErrNotAdmin) || errors.Is(err, auth.ErrNotTenantAdmin) {
			statusCode = fiber.StatusForbidden
		}
		return c.Status(statusCode).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// SetupTOTP initiates 2FA setup by generating a TOTP secret
// POST /auth/2fa/setup
func (h *AuthHandler) SetupTOTP(c fiber.Ctx) error {
	// Get user ID from JWT token
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Parse optional issuer from request body
	var req struct {
		Issuer string `json:"issuer"` // Optional: custom issuer name for the QR code
	}
	// Ignore parse errors - issuer is optional and will default to config value
	_ = c.Bind().Body(&req)

	response, err := h.authService.SetupTOTP(middleware.CtxWithTenant(c), userID.(string), req.Issuer)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to setup TOTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to setup 2FA",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// EnableTOTP enables 2FA after verifying the TOTP code
// POST /auth/2fa/enable
func (h *AuthHandler) EnableTOTP(c fiber.Ctx) error {
	// Get user ID from JWT token
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Code is required",
		})
	}

	backupCodes, err := h.authService.EnableTOTP(middleware.CtxWithTenant(c), userID.(string), req.Code)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to enable TOTP")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":      true,
		"backup_codes": backupCodes,
		"message":      "2FA enabled successfully. Please save your backup codes in a secure location.",
	})
}

// VerifyTOTP verifies a TOTP code during login and issues JWT tokens
// POST /auth/2fa/verify
func (h *AuthHandler) VerifyTOTP(c fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.UserID == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID and code are required",
		})
	}

	// Verify the 2FA code
	err := h.authService.VerifyTOTP(middleware.CtxWithTenant(c), req.UserID, req.Code)
	if err != nil {
		log.Warn().Err(err).Str("user_id", req.UserID).Msg("Failed to verify TOTP")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate a complete sign-in response with tokens
	resp, err := h.authService.GenerateTokensForUser(middleware.CtxWithTenant(c), req.UserID)
	if err != nil {
		log.Error().Err(err).Str("user_id", req.UserID).Msg("Failed to generate tokens after 2FA verification")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete authentication",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// DisableTOTP disables 2FA for a user
// POST /auth/2fa/disable
func (h *AuthHandler) DisableTOTP(c fiber.Ctx) error {
	// Get user ID from JWT token
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Password is required to disable 2FA",
		})
	}

	err := h.authService.DisableTOTP(middleware.CtxWithTenant(c), userID.(string), req.Password)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to disable TOTP")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "2FA disabled successfully",
	})
}

// GetTOTPStatus checks if 2FA is enabled for a user
// GET /auth/2fa/status
func (h *AuthHandler) GetTOTPStatus(c fiber.Ctx) error {
	// Get user ID from JWT token
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	enabled, err := h.authService.IsTOTPEnabled(middleware.CtxWithTenant(c), userID.(string))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to check TOTP status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check 2FA status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"totp_enabled": enabled,
	})
}

// SendOTP sends an OTP code via email or SMS
// POST /auth/otp/signin
func (h *AuthHandler) SendOTP(c fiber.Ctx) error {
	var req struct {
		Email   *string                 `json:"email,omitempty"`
		Phone   *string                 `json:"phone,omitempty"`
		Options *map[string]interface{} `json:"options,omitempty"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate that either email or phone is provided
	if err := auth.ValidateOTPContact(req.Email, req.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate auth service is initialized (after input validation)
	if h.authService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication service not available",
		})
	}

	// Send OTP
	var err error
	purpose := "signin" // Default purpose
	if req.Options != nil {
		if p, ok := (*req.Options)["purpose"].(string); ok {
			purpose = p
		}
	}

	if req.Email != nil {
		err = h.authService.SendOTP(middleware.CtxWithTenant(c), *req.Email, purpose)
	} else if req.Phone != nil {
		// SMS OTP not yet fully implemented
		err = fmt.Errorf("SMS OTP not yet implemented")
	}

	if err != nil {
		log.Error().Str("error", err.Error()).Msg("Failed to send OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send OTP code",
		})
	}

	// Return Supabase-compatible OTP response
	// For send requests, user and session are both nil (OTP delivered but not verified yet)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":    nil,
		"session": nil,
	})
}

// VerifyOTP verifies an OTP code and creates a session
// POST /auth/otp/verify
func (h *AuthHandler) VerifyOTP(c fiber.Ctx) error {
	var req struct {
		Email *string `json:"email,omitempty"`
		Phone *string `json:"phone,omitempty"`
		Token string  `json:"token"`
		Type  string  `json:"type"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP token is required",
		})
	}

	// Verify OTP
	var otpCode *auth.OTPCode
	var err error

	// Validate that either email or phone is provided
	if err := auth.ValidateOTPContact(req.Email, req.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if req.Email != nil {
		otpCode, err = h.authService.VerifyOTP(middleware.CtxWithTenant(c), *req.Email, req.Token)
	} else if req.Phone != nil {
		// Phone OTP not yet fully implemented
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"error": "Phone-based OTP authentication not yet implemented",
		})
	}

	if err != nil {
		log.Warn().Err(err).Msg("Failed to verify OTP")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired OTP code",
		})
	}

	// Get existing user - auto-creation is disabled for security
	// Users must register via signup endpoint first
	var user *auth.User
	if req.Email != nil && otpCode.Email != nil {
		user, err = h.authService.GetUserByEmail(middleware.CtxWithTenant(c), *otpCode.Email)
		if err != nil {
			log.Warn().Str("email", *otpCode.Email).Msg("OTP verification for non-existent user")
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "No account found for this email - please sign up first",
			})
		}
	}

	// Generate tokens
	resp, err := h.authService.GenerateTokensForUser(middleware.CtxWithTenant(c), user.ID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate tokens")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to complete authentication",
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// ResendOTP resends an OTP code
// POST /auth/otp/resend
func (h *AuthHandler) ResendOTP(c fiber.Ctx) error {
	var req struct {
		Type    string                  `json:"type"`
		Email   *string                 `json:"email,omitempty"`
		Phone   *string                 `json:"phone,omitempty"`
		Options *map[string]interface{} `json:"options,omitempty"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate that either email or phone is provided
	if err := auth.ValidateOTPContact(req.Email, req.Phone); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	purpose := "signin" // Default purpose
	if req.Options != nil {
		if p, ok := (*req.Options)["purpose"].(string); ok {
			purpose = p
		}
	}

	// Resend OTP
	var err error
	if req.Email != nil {
		err = h.authService.ResendOTP(middleware.CtxWithTenant(c), *req.Email, purpose)
	} else if req.Phone != nil {
		// SMS OTP not yet fully implemented
		err = fmt.Errorf("SMS OTP not yet implemented")
	}

	if err != nil {
		log.Error().Err(err).Msg("Failed to resend OTP")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to resend OTP code",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"user":    nil,
		"session": nil,
	})
}

// GetUserIdentities gets all OAuth identities linked to a user
// GET /auth/user/identities
func (h *AuthHandler) GetUserIdentities(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	identities, err := h.authService.GetUserIdentities(middleware.CtxWithTenant(c), userID.(string))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to get user identities")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve identities",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"identities": identities,
	})
}

// LinkIdentity initiates OAuth flow to link a provider
// POST /auth/user/identities
func (h *AuthHandler) LinkIdentity(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	var req struct {
		Provider string `json:"provider"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Provider is required",
		})
	}

	authURL, state, err := h.authService.LinkIdentity(middleware.CtxWithTenant(c), userID.(string), req.Provider)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Msg("Failed to initiate identity linking")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"url":      authURL,
		"provider": req.Provider,
		"state":    state,
	})
}

// UnlinkIdentity removes an OAuth identity from a user
// DELETE /auth/user/identities/:id
func (h *AuthHandler) UnlinkIdentity(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	identityID := c.Params("id")
	if identityID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Identity ID is required",
		})
	}

	err := h.authService.UnlinkIdentity(middleware.CtxWithTenant(c), userID.(string), identityID)
	if err != nil {
		log.Error().Err(err).Str("identity_id", identityID).Msg("Failed to unlink identity")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
	})
}

// Reauthenticate generates a security nonce
// POST /auth/reauthenticate
func (h *AuthHandler) Reauthenticate(c fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	nonce, err := h.authService.Reauthenticate(middleware.CtxWithTenant(c), userID.(string))
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.(string)).Msg("Failed to reauthenticate")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate security nonce",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"nonce": nonce,
	})
}

// SignInWithIDToken handles OAuth ID token authentication (Google, Apple)
// POST /auth/signin/idtoken
func (h *AuthHandler) SignInWithIDToken(c fiber.Ctx) error {
	var req struct {
		Provider string  `json:"provider"`
		Token    string  `json:"token"`
		Nonce    *string `json:"nonce,omitempty"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Provider == "" || req.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Provider and token are required",
		})
	}

	nonce := ""
	if req.Nonce != nil {
		nonce = *req.Nonce
	}

	resp, err := h.authService.SignInWithIDToken(middleware.CtxWithTenant(c), req.Provider, req.Token, nonce)
	if err != nil {
		log.Error().Err(err).Str("provider", req.Provider).Msg("Failed to sign in with ID token")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// GetCaptchaConfig returns the public CAPTCHA configuration for clients
// GET /auth/captcha/config
func (h *AuthHandler) GetCaptchaConfig(c fiber.Ctx) error {
	if h.captchaService == nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"enabled": false,
		})
	}

	config := h.captchaService.GetConfig()
	return c.Status(fiber.StatusOK).JSON(config)
}

// CheckCaptcha performs a pre-flight check to determine if CAPTCHA is required
// POST /auth/captcha/check
//
// This endpoint evaluates trust signals and returns whether CAPTCHA verification
// is needed for the subsequent auth action. It issues a challenge_id that must
// be included in the actual auth request.
//
// Request body:
//
//	{
//	  "endpoint": "login",                    // Required: signup, login, password_reset, magic_link
//	  "email": "user@example.com",            // Optional: for trust lookup
//	  "device_fingerprint": "abc123",         // Optional: browser fingerprint
//	  "trust_token": "tt_..."                 // Optional: token from previous CAPTCHA
//	}
//
// Response:
//
//	{
//	  "captcha_required": true,
//	  "reason": "new_ip_address",
//	  "trust_score": 35,
//	  "provider": "hcaptcha",
//	  "site_key": "...",
//	  "challenge_id": "ch_abc123...",
//	  "expires_at": "2024-01-15T10:05:00Z"
//	}
func (h *AuthHandler) CheckCaptcha(c fiber.Ctx) error {
	// Parse request
	var req auth.CaptchaCheckRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Validate endpoint
	validEndpoints := map[string]bool{
		"signup":         true,
		"login":          true,
		"password_reset": true,
		"magic_link":     true,
	}
	if !validEndpoints[req.Endpoint] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid endpoint. Must be one of: signup, login, password_reset, magic_link",
			"code":  "INVALID_ENDPOINT",
		})
	}

	// If CAPTCHA is not enabled at all, return early
	if h.captchaService == nil || !h.captchaService.IsEnabled() {
		return c.Status(fiber.StatusOK).JSON(auth.CaptchaCheckResponse{
			CaptchaRequired: false,
			Reason:          "captcha_disabled",
			ChallengeID:     "", // No challenge needed
		})
	}

	// If adaptive trust service is available, use it
	if h.captchaTrustService != nil {
		response, err := h.captchaTrustService.CheckCaptchaRequired(middleware.CtxWithTenant(c), req, c.IP(), c.Get("User-Agent"))
		if err != nil {
			log.Error().Err(err).Msg("Failed to check CAPTCHA requirement")
			// Fall back to requiring CAPTCHA on error
			return c.Status(fiber.StatusOK).JSON(auth.CaptchaCheckResponse{
				CaptchaRequired: true,
				Reason:          "trust_check_error",
				Provider:        h.captchaService.GetProvider(),
				SiteKey:         h.captchaService.GetSiteKey(),
			})
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}

	// Fall back to static check (adaptive trust not configured)
	required := h.captchaService.IsEnabledForEndpoint(req.Endpoint)
	response := auth.CaptchaCheckResponse{
		CaptchaRequired: required,
		ChallengeID:     "", // No challenge tracking without trust service
	}
	if required {
		response.Reason = "captcha_enabled_for_endpoint"
		response.Provider = h.captchaService.GetProvider()
		response.SiteKey = h.captchaService.GetSiteKey()
	} else {
		response.Reason = "captcha_not_required_for_endpoint"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetAuthConfig returns the public authentication configuration for clients
// GET /auth/config
func (h *AuthHandler) GetAuthConfig(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	settingsCache := h.authService.GetSettingsCache()

	// Build response
	response := AuthConfigResponse{
		SignupEnabled:            h.authService.IsSignupEnabled(),
		RequireEmailVerification: settingsCache.GetBool(ctx, "app.auth.require_email_verification", false),
		MagicLinkEnabled:         settingsCache.GetBool(ctx, "app.auth.magic_link_enabled", false),
		PasswordLoginEnabled:     !settingsCache.GetBool(ctx, "app.auth.disable_app_password_login", false), // Inverted: disabled=false means enabled=true
		MFAAvailable:             true,                                                                      // MFA is always available, users opt-in
		PasswordMinLength:        settingsCache.GetInt(ctx, "app.auth.password_min_length", 8),
		PasswordRequireUppercase: settingsCache.GetBool(ctx, "app.auth.password_require_uppercase", false),
		PasswordRequireLowercase: settingsCache.GetBool(ctx, "app.auth.password_require_lowercase", false),
		PasswordRequireNumber:    settingsCache.GetBool(ctx, "app.auth.password_require_number", false),
		PasswordRequireSpecial:   settingsCache.GetBool(ctx, "app.auth.password_require_special", false),
		OAuthProviders:           []OAuthProviderPublic{},
		SAMLProviders:            []SAMLProviderPublic{},
	}

	// Fetch OAuth providers
	oauthQuery := `
		SELECT provider_name, display_name, redirect_url
		FROM platform.oauth_providers
		WHERE enabled = TRUE AND allow_app_login = TRUE
		ORDER BY display_name
	`
	rows, err := h.db.Query(ctx, oauthQuery)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list OAuth providers for auth config")
	} else {
		defer rows.Close()
		for rows.Next() {
			var providerName, displayName, redirectURL string
			if err := rows.Scan(&providerName, &displayName, &redirectURL); err != nil {
				log.Error().Err(err).Msg("Failed to scan OAuth provider")
				continue
			}
			response.OAuthProviders = append(response.OAuthProviders, OAuthProviderPublic{
				Provider:     providerName,
				DisplayName:  displayName,
				AuthorizeURL: fmt.Sprintf("%s/api/v1/auth/oauth/%s/authorize", h.baseURL, providerName),
			})
		}
	}

	// Fetch SAML providers
	if h.samlService != nil {
		samlProviders := h.samlService.GetProvidersForApp()
		for _, provider := range samlProviders {
			response.SAMLProviders = append(response.SAMLProviders, SAMLProviderPublic{
				Provider:    provider.Name,
				DisplayName: provider.Name, // SAML providers use Name as display name
			})
		}
	}

	// Get CAPTCHA config
	if h.captchaService != nil {
		captchaConfig := h.captchaService.GetConfig()
		response.Captcha = &captchaConfig
	} else {
		response.Captcha = &auth.CaptchaConfigResponse{
			Enabled: false,
		}
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// isPasswordLoginDisabled checks if password login is disabled for app users
func (h *AuthHandler) isPasswordLoginDisabled(ctx context.Context) bool {
	// Emergency override via environment variable
	if os.Getenv("FLUXBASE_APP_FORCE_PASSWORD_LOGIN") == "true" {
		return false // Password login forced enabled
	}

	settingsCache := h.authService.GetSettingsCache()
	return settingsCache.GetBool(ctx, "app.auth.disable_app_password_login", false)
}

// fiber:context-methods migrated
