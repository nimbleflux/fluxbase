package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// SetupTOTP initiates 2FA setup by generating a TOTP secret
// POST /auth/2fa/setup
func (h *AuthHandler) SetupTOTP(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendMissingAuth(c)
	}

	var req struct {
		Issuer string `json:"issuer"`
	}
	_ = c.Bind().Body(&req)

	response, err := h.authService.MFAService().SetupTOTP(middleware.CtxWithTenant(c), userID, req.Issuer)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to setup TOTP")
		return SendInternalError(c, "Failed to setup 2FA")
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// EnableTOTP enables 2FA after verifying the TOTP code
// POST /auth/2fa/enable
func (h *AuthHandler) EnableTOTP(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendMissingAuth(c)
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.Code == "" {
		return SendMissingField(c, "Code")
	}

	backupCodes, err := h.authService.MFAService().EnableTOTP(middleware.CtxWithTenant(c), userID, req.Code)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to enable TOTP")
		return SendBadRequest(c, "Invalid 2FA code", ErrCodeInvalidInput)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":      true,
		"backup_codes": backupCodes,
		"message":      "2FA enabled successfully. Please save your backup codes in a secure location.",
	})
}

// VerifyTOTP verifies a TOTP code during login and issues JWT tokens
// POST /auth/2fa/verify
//
// The request must include the short-lived mfa_token returned by
// POST /auth/signin when requires_2fa was true; the user is identified from
// that signed token rather than the request body.
func (h *AuthHandler) VerifyTOTP(c fiber.Ctx) error {
	var req struct {
		MFAToken string `json:"mfa_token"`
		UserID   string `json:"user_id"`
		Code     string `json:"code"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.MFAToken == "" {
		return SendUnauthorized(c, "mfa_token is required. Provide the token returned by the sign-in response.", ErrCodeMissingField)
	}
	if req.Code == "" {
		return SendMissingField(c, "Code")
	}

	// Validate the signed 2FA challenge token bound to the password-verified
	// sign-in attempt. The user is taken from its claims; the request body
	// user_id (kept for backward compatibility) must match when provided.
	claims, err := h.authService.JWTManager().ValidateMFAPendingToken(req.MFAToken, auth.MFAPendingPurpose2FA)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid or expired 2FA challenge token")
		return SendUnauthorized(c, "Invalid or expired 2FA challenge token. Please sign in again.", ErrCodeInvalidToken)
	}
	userID := claims.UserID
	if req.UserID != "" && req.UserID != userID {
		return SendBadRequest(c, "mfa_token does not match user_id", ErrCodeInvalidInput)
	}

	// Verify the 2FA code
	if err := h.authService.MFAService().VerifyTOTP(middleware.CtxWithTenant(c), userID, req.Code); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to verify TOTP")
		return SendBadRequest(c, "Invalid 2FA code", ErrCodeInvalidCredentials)
	}

	// Generate a complete sign-in response with tokens
	resp, err := h.authService.GenerateTokensForUser(middleware.CtxWithTenant(c), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to generate tokens after 2FA verification")
		return SendInternalError(c, "Failed to complete authentication")
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// DisableTOTP disables 2FA for a user
// POST /auth/2fa/disable
func (h *AuthHandler) DisableTOTP(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendMissingAuth(c)
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.Password == "" {
		return SendMissingField(c, "Password")
	}

	err := h.authService.MFAService().DisableTOTP(middleware.CtxWithTenant(c), userID, req.Password)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to disable TOTP")
		return SendBadRequest(c, "Failed to disable 2FA", ErrCodeInvalidCredentials)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "2FA disabled successfully",
	})
}

// GetTOTPStatus checks if 2FA is enabled for a user
// GET /auth/2fa/status
func (h *AuthHandler) GetTOTPStatus(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendMissingAuth(c)
	}

	enabled, err := h.authService.MFAService().IsTOTPEnabled(middleware.CtxWithTenant(c), userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to check TOTP status")
		return SendInternalError(c, "Failed to check 2FA status")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"totp_enabled": enabled,
	})
}
