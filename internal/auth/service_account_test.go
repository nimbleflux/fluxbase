package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Self-service account deletion (internal/auth/service_account.go)
// =============================================================================

// newAccountDeletionTestService builds a Service wired to the package's shared
// test database (initialized by TestMain in clientkey_test.go). Same wiring as
// newSelfUpdateTestService in service_setters_test.go.
func newAccountDeletionTestService(t *testing.T) *Service {
	t.Helper()
	if sharedTestDB == nil {
		t.Skip("shared test database not initialized")
	}
	return newSelfUpdateTestService(t)
}

func TestDeleteSelfAccount_VerifyPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	service := newAccountDeletionTestService(t)
	ctx := context.Background()

	t.Run("password user requires matching password", func(t *testing.T) {
		email := fmt.Sprintf("acctdel-%s@example.com", uuid.New().String()[:8])
		user, err := service.CreateUser(ctx, email, "Password123!")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.token_blacklist WHERE revoked_by = $1`, user.ID)
			_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, user.ID)
		})

		// Missing password -> rejected
		err = service.VerifyAccountDeletion(ctx, user.ID, "")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidCredentials), "expected ErrInvalidCredentials, got %v", err)

		// Wrong password -> rejected with the same generic error
		err = service.VerifyAccountDeletion(ctx, user.ID, "WrongPassword!")
		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidCredentials), "expected ErrInvalidCredentials, got %v", err)

		// Correct password -> accepted
		err = service.VerifyAccountDeletion(ctx, user.ID, "Password123!")
		assert.NoError(t, err)
	})

	t.Run("oauth-only user needs no password", func(t *testing.T) {
		email := fmt.Sprintf("acctdel-oauth-%s@example.com", uuid.New().String()[:8])
		// Empty password creates a user with no password credential (empty hash)
		user, err := service.CreateUser(ctx, email, "")
		require.NoError(t, err)
		t.Cleanup(func() {
			_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.token_blacklist WHERE revoked_by = $1`, user.ID)
			_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, user.ID)
		})

		// No body / empty password -> deletion may proceed
		err = service.VerifyAccountDeletion(ctx, user.ID, "")
		assert.NoError(t, err)

		// A supplied password is simply ignored for credential-less users
		err = service.VerifyAccountDeletion(ctx, user.ID, "whatever-provided")
		assert.NoError(t, err)
	})

	t.Run("unknown user is rejected", func(t *testing.T) {
		err := service.VerifyAccountDeletion(ctx, uuid.New().String(), "")
		assert.True(t, errors.Is(err, ErrUserNotFound), "expected ErrUserNotFound, got %v", err)
	})
}

func TestDeleteSelfAccount_RevokesAccessAndDeletesRow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	service := newAccountDeletionTestService(t)
	ctx := context.Background()

	email := fmt.Sprintf("acctdel-full-%s@example.com", uuid.New().String()[:8])
	user, err := service.CreateUser(ctx, email, "Password123!")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.token_blacklist WHERE revoked_by = $1`, user.ID)
		_, _ = sharedTestDB.Exec(ctx, `DELETE FROM auth.users WHERE id = $1`, user.ID)
	})

	// Issue tokens and a live session like a real signin would.
	accessToken, refreshToken, _, err := service.JWTManager().GenerateTokenPair(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	_, err = service.SessionRepository().Create(ctx, user.ID, accessToken, refreshToken, time.Now().Add(time.Hour))
	require.NoError(t, err)

	// Sanity: token + session are usable before deletion.
	_, err = service.GetUser(ctx, accessToken)
	require.NoError(t, err)

	// Full service flow: verify -> revoke -> delete row.
	require.NoError(t, service.VerifyAccountDeletion(ctx, user.ID, "Password123!"))
	require.NoError(t, service.RevokeAccountAccess(ctx, user.ID, accessToken))
	require.NoError(t, service.DeleteAccountRecord(ctx, user.ID))

	// User row is gone.
	_, err = service.UserRepository().GetByID(ctx, user.ID)
	assert.True(t, errors.Is(err, ErrUserNotFound), "expected ErrUserNotFound, got %v", err)

	// Sessions are gone (explicitly revoked and/or FK cascade).
	_, err = service.SessionRepository().GetByAccessToken(ctx, accessToken)
	assert.True(t, errors.Is(err, ErrSessionNotFound), "expected ErrSessionNotFound, got %v", err)

	// The token used to perform the deletion is now invalid: its JTI is
	// blacklisted and the user-wide revocation marker covers it regardless.
	claims, err := service.JWTManager().ValidateAccessToken(accessToken)
	require.NoError(t, err, "token itself should still parse; rejection happens at the revocation check")
	revoked, err := service.TokenBlacklistService().IsTokenRevoked(ctx, claims.ID, user.ID, claims.IssuedAt.Time)
	require.NoError(t, err)
	assert.True(t, revoked, "access token used for deletion must be revoked")

	// GetUser now fails because the session no longer exists.
	_, err = service.GetUser(ctx, accessToken)
	assert.Error(t, err, "post-delete token must not resolve to an authenticated user")
}
