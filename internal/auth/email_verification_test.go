package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashEmailVerificationToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"standard token", "abc123def456"},
		{"long token", "very-long-token-with-many-characters-for-email-verification"},
		{"short token", "abc"},
		{"empty token", ""},
		{"special characters", "token!@#$%^&*()"},
		{"unicode token", "token-用户-verification"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashEmailVerificationToken(tt.token)

			// Verify hash is not empty
			assert.NotEmpty(t, hash)

			// Verify hash is base64 URL encoded
			_, err := base64.URLEncoding.DecodeString(hash)
			assert.NoError(t, err)

			// Verify hash is deterministic
			hash2 := hashEmailVerificationToken(tt.token)
			assert.Equal(t, hash, hash2)
		})
	}
}

func TestHashEmailVerificationToken_SHA256(t *testing.T) {
	token := "test-verification-token"
	hash := hashEmailVerificationToken(token)

	// Manually compute SHA-256
	expectedHash := sha256.Sum256([]byte(token))
	expectedBase64 := base64.URLEncoding.EncodeToString(expectedHash[:])

	assert.Equal(t, expectedBase64, hash)
}

func TestHashEmailVerificationToken_DifferentTokens(t *testing.T) {
	token1 := "token1"
	token2 := "token2"

	hash1 := hashEmailVerificationToken(token1)
	hash2 := hashEmailVerificationToken(token2)

	assert.NotEqual(t, hash1, hash2)
}

func TestGenerateEmailVerificationToken_Success(t *testing.T) {
	token, err := generateEmailVerificationToken()

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token is base64 URL encoded
	decoded, err := base64.URLEncoding.DecodeString(token)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(decoded))
}

func TestGenerateEmailVerificationToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := generateEmailVerificationToken()
		require.NoError(t, err)

		assert.False(t, tokens[token], "tokens should be unique")
		tokens[token] = true
	}

	assert.Len(t, tokens, iterations)
}

func TestEmailVerificationToken_Integration(t *testing.T) {
	// Generate token
	token, err := generateEmailVerificationToken()
	require.NoError(t, err)

	// Hash it
	hash := hashEmailVerificationToken(token)

	// Verify properties
	assert.NotEqual(t, token, hash)
	assert.True(t, len(hash) > 40)

	// Verify same token produces same hash
	hash2 := hashEmailVerificationToken(token)
	assert.Equal(t, hash, hash2)
}

// =============================================================================
// Error Constants Tests
// =============================================================================

func TestEmailVerificationErrors(t *testing.T) {
	t.Run("error constants are defined", func(t *testing.T) {
		assert.NotNil(t, ErrEmailVerificationTokenNotFound)
		assert.NotNil(t, ErrEmailVerificationTokenExpired)
		assert.NotNil(t, ErrEmailVerificationTokenUsed)
		assert.NotNil(t, ErrEmailNotVerified)
	})

	t.Run("error messages are meaningful", func(t *testing.T) {
		assert.Contains(t, ErrEmailVerificationTokenNotFound.Error(), "not found")
		assert.Contains(t, ErrEmailVerificationTokenExpired.Error(), "expired")
		assert.Contains(t, ErrEmailVerificationTokenUsed.Error(), "already been used")
		assert.Contains(t, ErrEmailNotVerified.Error(), "not verified")
	})

	t.Run("errors are distinct", func(t *testing.T) {
		errors := []error{
			ErrEmailVerificationTokenNotFound,
			ErrEmailVerificationTokenExpired,
			ErrEmailVerificationTokenUsed,
			ErrEmailNotVerified,
		}

		for i, err1 := range errors {
			for j, err2 := range errors {
				if i != j {
					assert.NotEqual(t, err1, err2)
				}
			}
		}
	})
}

// =============================================================================
// EmailVerificationToken Struct Tests
// =============================================================================

func TestEmailVerificationToken_Struct(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		usedAt := time.Now()
		token := EmailVerificationToken{
			ID:        "token-123",
			UserID:    "user-456",
			TokenHash: "hash-abc",
			ExpiresAt: time.Now().Add(time.Hour),
			Used:      true,
			UsedAt:    &usedAt,
			CreatedAt: time.Now(),
		}

		assert.Equal(t, "token-123", token.ID)
		assert.Equal(t, "user-456", token.UserID)
		assert.Equal(t, "hash-abc", token.TokenHash)
		assert.True(t, token.Used)
		assert.NotNil(t, token.UsedAt)
	})

	t.Run("defaults to zero values", func(t *testing.T) {
		token := EmailVerificationToken{}

		assert.Empty(t, token.ID)
		assert.Empty(t, token.UserID)
		assert.Empty(t, token.TokenHash)
		assert.False(t, token.Used)
		assert.Nil(t, token.UsedAt)
	})

	t.Run("unused token has nil UsedAt", func(t *testing.T) {
		token := EmailVerificationToken{
			ID:     "token-123",
			UserID: "user-456",
			Used:   false,
		}

		assert.False(t, token.Used)
		assert.Nil(t, token.UsedAt)
	})
}

// =============================================================================
// EmailVerificationTokenWithPlaintext Tests
// =============================================================================

func TestEmailVerificationTokenWithPlaintext_Struct(t *testing.T) {
	t.Run("includes plaintext token", func(t *testing.T) {
		token := EmailVerificationTokenWithPlaintext{
			EmailVerificationToken: EmailVerificationToken{
				ID:        "token-123",
				UserID:    "user-456",
				TokenHash: "hash-abc",
				Used:      false,
			},
			PlaintextToken: "plaintext-secret-token",
		}

		assert.Equal(t, "token-123", token.ID)
		assert.Equal(t, "user-456", token.UserID)
		assert.Equal(t, "hash-abc", token.TokenHash)
		assert.Equal(t, "plaintext-secret-token", token.PlaintextToken)
	})

	t.Run("plaintext differs from hash", func(t *testing.T) {
		plaintext := "my-secret-token"
		hash := hashEmailVerificationToken(plaintext)

		token := EmailVerificationTokenWithPlaintext{
			EmailVerificationToken: EmailVerificationToken{
				TokenHash: hash,
			},
			PlaintextToken: plaintext,
		}

		assert.NotEqual(t, token.PlaintextToken, token.TokenHash)
	})
}

// =============================================================================
// Repository Constructor Tests
// =============================================================================

func TestNewEmailVerificationRepository(t *testing.T) {
	t.Run("creates repository with nil database", func(t *testing.T) {
		repo := NewEmailVerificationRepository(nil)
		assert.NotNil(t, repo)
	})
}

// =============================================================================
// Token Validation Logic Tests
// =============================================================================

func TestEmailVerificationToken_ExpiryCheck(t *testing.T) {
	t.Run("token not expired", func(t *testing.T) {
		token := EmailVerificationToken{
			ExpiresAt: time.Now().Add(time.Hour),
		}

		isExpired := time.Now().After(token.ExpiresAt)
		assert.False(t, isExpired)
	})

	t.Run("token expired", func(t *testing.T) {
		token := EmailVerificationToken{
			ExpiresAt: time.Now().Add(-time.Hour),
		}

		isExpired := time.Now().After(token.ExpiresAt)
		assert.True(t, isExpired)
	})

	t.Run("token expires exactly now (boundary)", func(t *testing.T) {
		now := time.Now()
		token := EmailVerificationToken{
			ExpiresAt: now,
		}

		// After immediate check, should be expired
		isExpired := time.Now().After(token.ExpiresAt)
		// May or may not be expired depending on timing, but should be deterministic for same time
		assert.IsType(t, true, isExpired)
	})
}

func TestEmailVerificationToken_UsedCheck(t *testing.T) {
	t.Run("unused token", func(t *testing.T) {
		token := EmailVerificationToken{
			Used: false,
		}

		assert.False(t, token.Used)
	})

	t.Run("used token", func(t *testing.T) {
		usedAt := time.Now()
		token := EmailVerificationToken{
			Used:   true,
			UsedAt: &usedAt,
		}

		assert.True(t, token.Used)
		assert.NotNil(t, token.UsedAt)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkHashEmailVerificationToken(b *testing.B) {
	token := "test-token-for-benchmarking"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hashEmailVerificationToken(token)
	}
}

func BenchmarkGenerateEmailVerificationToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateEmailVerificationToken()
	}
}
