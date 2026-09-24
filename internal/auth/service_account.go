package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
)

// VerifyAccountDeletion confirms the caller may delete the account by
// re-checking their password. Users WITH a password credential must supply the
// matching password; users WITHOUT one (OAuth/SAML/magic-link-only, empty
// stored hash) pass without a body.
//
// Split from the destructive half of account deletion so the API handler can
// run its cleanup pass (owned resources) between verification and the final
// user-row delete — mirroring the orchestrator layering used elsewhere.
func (s *Service) VerifyAccountDeletion(ctx context.Context, userID, password string) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// No password credential (OAuth-only etc.) — deletion needs no body.
	if user.PasswordHash == "" {
		return nil
	}

	if password == "" {
		return ErrInvalidCredentials
	}

	if err := s.passwordHasher.ComparePassword(user.PasswordHash, password); err != nil {
		return ErrInvalidCredentials
	}

	return nil
}

// RevokeAccountAccess kills every credential the account holds BEFORE owned
// resources are cleaned up and the user row is deleted:
//   - the current access token's JTI is blacklisted, so the very request that
//     performed the deletion is dead on its next use;
//   - a user-wide revocation marker invalidates every access/refresh token
//     issued before now (enforced by the auth middleware's IsTokenRevoked);
//   - all auth.sessions rows are removed (the auth.users FK cascade would also
//     remove them at row deletion, but revoking first closes the window where
//     other live sessions could still refresh).
//
// Blacklist failures are logged but do not block the flow (same policy as
// SignOut); a session-store failure is returned.
func (s *Service) RevokeAccountAccess(ctx context.Context, userID, accessToken string) error {
	if accessToken != "" {
		if err := s.tokenBlacklistService.RevokeToken(ctx, accessToken, "account_deletion"); err != nil {
			log.Warn().Err(err).Str("user_id", userID).Msg("Failed to blacklist current access token during account deletion")
		}
	}

	if err := s.tokenBlacklistService.RevokeAllUserTokens(ctx, userID, "account_deletion"); err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to add user-wide token revocation during account deletion")
	}

	if err := s.sessionRepo.DeleteByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke sessions: %w", err)
	}

	return nil
}

// DeleteAccountRecord performs the final auth.users row delete for a
// self-service account deletion. It reuses the same repository path as admin
// user deletion (UserRepository.DeleteFromTable with type "app"), so tenant
// FK cascades behave identically: child rows with ON DELETE CASCADE vanish,
// ON DELETE SET NULL columns are nulled, and tables without FKs to auth.users
// remain the tenant's responsibility (documented model).
func (s *Service) DeleteAccountRecord(ctx context.Context, userID string) error {
	return s.userRepo.DeleteFromTable(ctx, userID, "app")
}
