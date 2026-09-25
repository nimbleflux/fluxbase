package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// =============================================================================
// Service Setter Methods Tests
// =============================================================================

func TestService_SetEncryptionKey(t *testing.T) {
	service := &Service{mfaService: &MFAService{}}

	t.Run("sets encryption key", func(t *testing.T) {
		key := []byte("test-encryption-key-123")
		service.SetEncryptionKey(key)

		assert.Equal(t, key, service.mfaService.encryptionKey)
	})

	t.Run("can be updated multiple times", func(t *testing.T) {
		service.SetEncryptionKey([]byte("key1"))
		assert.Equal(t, []byte("key1"), service.mfaService.encryptionKey)

		service.SetEncryptionKey([]byte("key2"))
		assert.Equal(t, []byte("key2"), service.mfaService.encryptionKey)
	})

	t.Run("empty string is allowed", func(t *testing.T) {
		service.SetEncryptionKey([]byte(""))
		assert.Equal(t, []byte(""), service.mfaService.encryptionKey)
	})
}

func TestService_SetTOTPRateLimiter(t *testing.T) {
	service := &Service{mfaService: &MFAService{}}

	t.Run("sets TOTP rate limiter", func(t *testing.T) {
		limiter := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter)

		assert.Same(t, limiter, service.mfaService.totpRateLimiter)
	})

	t.Run("can be updated", func(t *testing.T) {
		limiter1 := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter1)
		assert.Same(t, limiter1, service.mfaService.totpRateLimiter)

		limiter2 := &TOTPRateLimiter{}
		service.SetTOTPRateLimiter(limiter2)
		assert.Same(t, limiter2, service.mfaService.totpRateLimiter)
	})

	t.Run("nil is allowed", func(t *testing.T) {
		service.SetTOTPRateLimiter(nil)
		assert.Nil(t, service.mfaService.totpRateLimiter)
	})
}

func TestService_RecordAuthAttempt(t *testing.T) {
	service := &Service{}

	t.Run("does not panic with nil metrics", func(t *testing.T) {
		// Should not panic even though metrics is nil
		assert.NotPanics(t, func() {
			service.recordAuthAttempt("password", true, "")
		})
	})

	t.Run("does not panic with nil metrics and failure", func(t *testing.T) {
		assert.NotPanics(t, func() {
			service.recordAuthAttempt("otp", false, "invalid_code")
		})
	})
}

func TestService_RecordAuthToken(t *testing.T) {
	service := &Service{}

	t.Run("does not panic with nil metrics", func(t *testing.T) {
		// Should not panic even though metrics is nil
		assert.NotPanics(t, func() {
			service.recordAuthToken("access")
		})
	})

	t.Run("does not panic with different token types", func(t *testing.T) {
		assert.NotPanics(t, func() {
			service.recordAuthToken("refresh")
		})

		assert.NotPanics(t, func() {
			service.recordAuthToken("service_role")
		})
	})
}

// =============================================================================
// UpdateSelfUser tests (self-service must not touch privileged fields)
// =============================================================================

// newSelfUpdateTestService builds a Service wired to the package's shared test
// database (initialized by TestMain in clientkey_test.go).
func newSelfUpdateTestService(t *testing.T) *Service {
	t.Helper()
	if sharedTestDB == nil {
		t.Skip("shared test database not initialized")
	}
	cfg := &config.AuthConfig{
		JWTSecret:      "test-secret-key-for-selfupdate-tests-32ch",
		JWTExpiry:      15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
		PasswordMinLen: 8,
		BcryptCost:     4,
		SignupEnabled:  true,
	}
	return NewService(sharedTestDB, cfg, &NoOpOTPSender{}, "http://localhost:3000", nil)
}

func TestUpdateSelfUser_CannotChangePrivilegedFields(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	service := newSelfUpdateTestService(t)
	ctx := context.Background()

	email := fmt.Sprintf("selfupdate-%s@example.com", uuid.New().String()[:8])
	user, err := service.CreateUser(ctx, email, "Password123")
	require.NoError(t, err)
	require.NotNil(t, user)

	originalRole := user.Role
	originalEmailVerified := user.EmailVerified
	originalAppMetadata := user.AppMetadata
	require.False(t, originalEmailVerified)

	// Self-update with metadata only: role, email_verified and app_metadata
	// must remain untouched.
	metadata := map[string]interface{}{"theme": "dark"}
	updated, err := service.UpdateSelfUser(ctx, user.ID, UpdateSelfUserRequest{
		UserMetadata: metadata,
	})
	require.NoError(t, err)
	assert.Equal(t, originalRole, updated.Role)
	assert.Equal(t, originalEmailVerified, updated.EmailVerified)
	assert.Equal(t, originalAppMetadata, updated.AppMetadata)
	assert.Equal(t, metadata, updated.UserMetadata)
}

func TestUpdateSelfUser_EmailUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	service := newSelfUpdateTestService(t)
	ctx := context.Background()

	email := fmt.Sprintf("selfupdate-%s@example.com", uuid.New().String()[:8])
	user, err := service.CreateUser(ctx, email, "Password123")
	require.NoError(t, err)

	newEmail := fmt.Sprintf("selfupdate-%s@example.com", uuid.New().String()[:8])
	updated, err := service.UpdateSelfUser(ctx, user.ID, UpdateSelfUserRequest{Email: &newEmail})
	require.NoError(t, err)
	assert.Equal(t, newEmail, updated.Email)
}

func TestUpdateSelfUser_RejectsInvalidEmail(t *testing.T) {
	service := newSelfUpdateTestService(t)
	ctx := context.Background()

	bad := "not-an-email"
	_, err := service.UpdateSelfUser(ctx, uuid.New().String(), UpdateSelfUserRequest{Email: &bad})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email")
}
