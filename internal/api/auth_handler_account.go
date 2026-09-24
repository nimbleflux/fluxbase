package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// SetKnowledgeBaseStorage wires the AI knowledge-base storage into the auth
// handler so self-service account deletion can clean up documents the user
// owns. Called from server.go after both modules are initialized; KBStorage is
// nil on instances with AI disabled and cleanup is skipped in that case.
func (h *AuthHandler) SetKnowledgeBaseStorage(kbStorage *ai.KnowledgeBaseStorage) {
	h.kbStorage = kbStorage
}

// DeleteAccount deletes the authenticated app-user's own account.
// DELETE /api/v1/auth/account
//
// Semantics:
//   - The user is resolved from the access token (Locals), never from the body.
//   - Optional JSON body { "password": "..." }: REQUIRED when the user has a
//     password credential (verified against the stored hash); ignored for
//     OAuth-only users (no stored hash) and for empty bodies.
//   - Order of operations: verify password -> revoke sessions + blacklist
//     tokens (the calling token dies with the account) -> cleanup owned
//     resources (KB documents by metadata user_id, storage.objects rows by
//     owner_id) -> delete the auth.users row (same repository path as admin
//     user deletion, so tenant FK cascades behave identically).
//   - Responses: 204 on success; 401 (middleware); 403 wrong/missing password
//     (generic wording, matching signin); 404 unknown user; 500 sanitized
//     (cleanup failure fails closed BEFORE the user row is deleted).
func (h *AuthHandler) DeleteAccount(c fiber.Ctx) error {
	// Auth middleware populates user_id; resolve the user from Locals only.
	userID := middleware.GetUserID(c)
	if userID == "" {
		return SendUnauthorized(c, "Unauthorized", ErrCodeAuthRequired)
	}

	// Optional body: an empty body is valid (OAuth-only users, or clients that
	// simply omit it). Only a present body must parse as JSON.
	var req struct {
		Password string `json:"password"`
	}
	if len(c.Body()) > 0 {
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return SendBadRequest(c, "Invalid request body", ErrCodeValidationFailed)
		}
	}

	ctx := middleware.CtxWithTenant(c)

	// a. Verify password when the account has a password credential.
	if err := h.authService.VerifyAccountDeletion(ctx, userID, req.Password); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return SendNotFound(c, "User not found")
		}
		if errors.Is(err, auth.ErrInvalidCredentials) {
			// Generic wording, same style as signin — never reveal which factor failed.
			return SendForbidden(c, "Invalid email or password", ErrCodeInvalidCredentials)
		}
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to verify account deletion password")
		return SendInternalError(c, "Failed to delete account")
	}

	// b. Revoke sessions and blacklist tokens (including the current one).
	if err := h.authService.RevokeAccountAccess(ctx, userID, h.getAccessToken(c)); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to revoke account access")
		return SendInternalError(c, "Failed to delete account")
	}

	// c. Cleanup owned resources BEFORE the final user-row delete. Fail-closed:
	// a cleanup error leaves the auth row (and login) intact so the user can
	// re-authenticate and retry rather than stranding unreachable data.
	if err := h.cleanupAccountResources(ctx, userID); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to clean up account resources")
		return SendInternalError(c, "Failed to delete account")
	}

	// d. Delete the auth.users row (same path as admin deletion; FK cascades fire).
	if err := h.authService.DeleteAccountRecord(ctx, userID); err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return SendNotFound(c, "User not found")
		}
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete account row")
		return SendInternalError(c, "Failed to delete account")
	}

	log.Info().Str("user_id", userID).Msg("Account self-deleted")

	// The calling session is already dead; clear cookies for browser clients.
	h.clearAuthCookies(c)

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// cleanupAccountResources removes resources the user owns that are NOT covered
// by FK cascades from auth.users:
//   - KB documents scoped by metadata user_id (ai.documents.owner_id/created_by
//     are ON DELETE SET NULL, so without this pass the documents would survive).
//   - storage.objects rows owned by the user (owner_id is ON DELETE SET NULL).
//     NOTE: provider-side bytes (local disk/S3) are NOT deleted by this
//     endpoint; only the metadata rows are removed.
//
// KB cleanup is skipped when the AI module is disabled (kbStorage == nil).
func (h *AuthHandler) cleanupAccountResources(ctx context.Context, userID string) error {
	if h.kbStorage != nil {
		deleted, err := h.kbStorage.DeleteUserDocuments(ctx, userID)
		if err != nil {
			return err
		}
		if deleted > 0 {
			log.Info().Str("user_id", userID).Int("documents", deleted).Msg("Deleted user KB documents during account deletion")
		}
	}

	// storage.objects metadata rows under service role (RLS bypassed).
	if h.db != nil {
		err := database.WrapWithServiceRole(ctx, h.db, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `DELETE FROM storage.objects WHERE owner_id = $1`, userID)
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}
