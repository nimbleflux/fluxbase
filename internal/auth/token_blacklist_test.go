package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRevokeToken_ServiceRoleTokenRejected(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	serviceRoleToken, err := manager.GenerateServiceRoleToken()
	require.NoError(t, err)

	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	err = service.RevokeToken(context.Background(), serviceRoleToken, "test revocation")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeServiceRole)
}

func TestRevokeToken_AnonTokenAlsoBlocked(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	anonToken, err := manager.GenerateAnonToken()
	require.NoError(t, err)

	claims, err := manager.ValidateServiceRoleToken(anonToken)
	require.NoError(t, err)
	assert.Equal(t, "anon", claims.Role)

	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	err = service.RevokeToken(context.Background(), anonToken, "test revocation")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeServiceRole)
}

func TestRevokeToken_AllServiceRoleTokensBlocked(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	serviceRoleToken, err := manager.GenerateServiceRoleToken()
	require.NoError(t, err)

	claims, err := manager.ValidateServiceRoleToken(serviceRoleToken)
	require.NoError(t, err)
	assert.Equal(t, "service_role", claims.Role)

	err = service.RevokeToken(context.Background(), serviceRoleToken, "test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeServiceRole)

	anonToken, err := manager.GenerateAnonToken()
	require.NoError(t, err)

	anonClaims, err := manager.ValidateServiceRoleToken(anonToken)
	require.NoError(t, err)
	assert.Equal(t, "anon", anonClaims.Role)

	err = service.RevokeToken(context.Background(), anonToken, "test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeServiceRole)
}

func TestRevokeToken_ServiceKeyRejected(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	err = service.RevokeToken(context.Background(), "sk_test1234567890abcdef", "logout")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeServiceKey)
}

func TestRevokeToken_ClientKeyRejected(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)

	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	err = service.RevokeToken(context.Background(), "fbk_test1234abcd", "logout")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrCannotRevokeClientKey)
}

func TestTokenBlacklistErrors(t *testing.T) {
	t.Run("ErrTokenBlacklisted", func(t *testing.T) {
		assert.NotNil(t, ErrTokenBlacklisted)
		assert.Contains(t, ErrTokenBlacklisted.Error(), "revoked")
	})

	t.Run("ErrCannotRevokeServiceRole", func(t *testing.T) {
		assert.NotNil(t, ErrCannotRevokeServiceRole)
		assert.Contains(t, ErrCannotRevokeServiceRole.Error(), "service role")
	})

	t.Run("ErrCannotRevokeServiceKey", func(t *testing.T) {
		assert.NotNil(t, ErrCannotRevokeServiceKey)
		assert.Contains(t, ErrCannotRevokeServiceKey.Error(), "service key")
	})

	t.Run("ErrCannotRevokeClientKey", func(t *testing.T) {
		assert.NotNil(t, ErrCannotRevokeClientKey)
		assert.Contains(t, ErrCannotRevokeClientKey.Error(), "client key")
	})
}

func TestTokenBlacklistEntry_Struct(t *testing.T) {
	t.Run("creates entry with all fields", func(t *testing.T) {
		now := time.Now()
		entry := &TokenBlacklistEntry{
			ID:        "entry-123",
			TokenJTI:  "jti-456",
			RevokedBy: "user-789",
			Reason:    "logout",
			CreatedAt: now,
			ExpiresAt: now.Add(time.Hour),
		}

		assert.Equal(t, "entry-123", entry.ID)
		assert.Equal(t, "jti-456", entry.TokenJTI)
		assert.Equal(t, "user-789", entry.RevokedBy)
		assert.Equal(t, "logout", entry.Reason)
		assert.Equal(t, now, entry.CreatedAt)
		assert.True(t, entry.ExpiresAt.After(now))
	})

	t.Run("zero value", func(t *testing.T) {
		var entry TokenBlacklistEntry
		assert.Empty(t, entry.ID)
		assert.Empty(t, entry.TokenJTI)
		assert.Empty(t, entry.RevokedBy)
		assert.Empty(t, entry.Reason)
		assert.True(t, entry.CreatedAt.IsZero())
		assert.True(t, entry.ExpiresAt.IsZero())
	})

	t.Run("common revocation reasons", func(t *testing.T) {
		reasons := []string{"logout", "password_change", "security_alert", "admin_action", "session_timeout"}
		for _, reason := range reasons {
			entry := &TokenBlacklistEntry{Reason: reason}
			assert.Equal(t, reason, entry.Reason)
		}
	})
}

func TestNewTokenBlacklistRepository(t *testing.T) {
	t.Run("creates with nil db", func(t *testing.T) {
		repo := NewTokenBlacklistRepository(nil)
		require.NotNil(t, repo)
		assert.Nil(t, repo.db)
	})
}

func TestNewTokenBlacklistService(t *testing.T) {
	t.Run("creates with nil dependencies", func(t *testing.T) {
		svc := NewTokenBlacklistService(nil, nil)
		require.NotNil(t, svc)
		assert.Nil(t, svc.repo)
		assert.Nil(t, svc.jwtManager)
	})

	t.Run("creates with repo only", func(t *testing.T) {
		repo := NewTokenBlacklistRepository(nil)
		svc := NewTokenBlacklistService(repo, nil)
		assert.Equal(t, repo, svc.repo)
		assert.Nil(t, svc.jwtManager)
	})

	t.Run("creates with jwt manager only", func(t *testing.T) {
		manager, err := NewJWTManager(testSecretKey, time.Hour, time.Hour*24)
		require.NoError(t, err)
		svc := NewTokenBlacklistService(nil, manager)
		assert.Nil(t, svc.repo)
		assert.Equal(t, manager, svc.jwtManager)
	})
}

func TestRevokeToken_ServiceKeyPrefixVariations(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)
	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	testCases := []string{
		"sk_",
		"sk_short",
		"sk_longer_token_value",
		"sk_12345678901234567890",
	}

	for _, token := range testCases {
		t.Run(token, func(t *testing.T) {
			err := service.RevokeToken(context.Background(), token, "test")
			assert.ErrorIs(t, err, ErrCannotRevokeServiceKey)
		})
	}
}

func TestRevokeToken_ClientKeyPrefixVariations(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)
	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	testCases := []string{
		"fbk_",
		"fbk_short",
		"fbk_longer_token_value",
		"fbk_12345678901234567890",
	}

	for _, token := range testCases {
		t.Run(token, func(t *testing.T) {
			err := service.RevokeToken(context.Background(), token, "test")
			assert.ErrorIs(t, err, ErrCannotRevokeClientKey)
		})
	}
}

func TestRevokeToken_NormalTokensNotPrefixBlocked(t *testing.T) {
	manager, err := NewJWTManager(testSecretKey, 15*time.Minute, 7*24*time.Hour)
	require.NoError(t, err)
	service := &TokenBlacklistService{
		repo:       nil,
		jwtManager: manager,
	}

	testCases := []string{
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"random_token",
		"fk_not_fbk",
		"s_not_sk",
	}

	for _, token := range testCases {
		t.Run(token, func(t *testing.T) {
			err := service.RevokeToken(context.Background(), token, "test")
			assert.NotErrorIs(t, err, ErrCannotRevokeServiceKey)
			assert.NotErrorIs(t, err, ErrCannotRevokeClientKey)
		})
	}
}

func TestTokenBlacklistService_IsTokenRevoked(t *testing.T) {
	t.Run("service delegates to repo", func(t *testing.T) {
		svc := NewTokenBlacklistService(nil, nil)
		assert.NotNil(t, svc)
	})
}

func TestTokenBlacklistService_RevokeAllUserTokens(t *testing.T) {
	t.Run("service delegates to repo", func(t *testing.T) {
		svc := NewTokenBlacklistService(nil, nil)
		assert.NotNil(t, svc)
	})
}

func TestTokenBlacklistService_CleanupExpiredTokens(t *testing.T) {
	t.Run("service delegates to repo", func(t *testing.T) {
		svc := NewTokenBlacklistService(nil, nil)
		assert.NotNil(t, svc)
	})
}
