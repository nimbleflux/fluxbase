package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPasswordResetToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"standard token", "reset123token456"},
		{"long token", "very-long-password-reset-token-with-many-characters"},
		{"short token", "xyz"},
		{"empty token", ""},
		{"special characters", "reset!@#$%^&*()"},
		{"unicode token", "reset-密码-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := hashPasswordResetToken(tt.token)

			// Verify hash is not empty
			assert.NotEmpty(t, hash)

			// Verify hash is base64 URL encoded
			_, err := base64.URLEncoding.DecodeString(hash)
			assert.NoError(t, err)

			// Verify hash is deterministic
			hash2 := hashPasswordResetToken(tt.token)
			assert.Equal(t, hash, hash2)
		})
	}
}

func TestHashPasswordResetToken_SHA256(t *testing.T) {
	token := "test-password-reset-token"
	hash := hashPasswordResetToken(token)

	// Manually compute SHA-256
	expectedHash := sha256.Sum256([]byte(token))
	expectedBase64 := base64.URLEncoding.EncodeToString(expectedHash[:])

	assert.Equal(t, expectedBase64, hash)
}

func TestHashPasswordResetToken_DifferentTokens(t *testing.T) {
	token1 := "reset1"
	token2 := "reset2"

	hash1 := hashPasswordResetToken(token1)
	hash2 := hashPasswordResetToken(token2)

	assert.NotEqual(t, hash1, hash2)
}

func TestHashPasswordResetToken_CaseSensitive(t *testing.T) {
	token1 := "ResetToken"
	token2 := "resettoken"

	hash1 := hashPasswordResetToken(token1)
	hash2 := hashPasswordResetToken(token2)

	assert.NotEqual(t, hash1, hash2)
}

func TestGeneratePasswordResetToken_Success(t *testing.T) {
	token, err := GeneratePasswordResetToken()

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify token is base64 URL encoded
	decoded, err := base64.URLEncoding.DecodeString(token)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(decoded), "token should be 32 bytes when decoded")
}

func TestGeneratePasswordResetToken_Uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := GeneratePasswordResetToken()
		require.NoError(t, err)

		assert.False(t, tokens[token], "tokens should be unique")
		tokens[token] = true
	}

	assert.Len(t, tokens, iterations)
}

func TestGeneratePasswordResetToken_URLSafe(t *testing.T) {
	for i := 0; i < 50; i++ {
		token, err := GeneratePasswordResetToken()
		require.NoError(t, err)

		// URL-safe base64 should not contain + or /
		assert.NotContains(t, token, "+")
		assert.NotContains(t, token, "/")
	}
}

func TestPasswordResetToken_Integration(t *testing.T) {
	// Generate token
	token, err := GeneratePasswordResetToken()
	require.NoError(t, err)

	// Hash it
	hash := hashPasswordResetToken(token)

	// Verify properties
	assert.NotEqual(t, token, hash)
	assert.True(t, len(hash) > 40)

	// Verify same token produces same hash
	hash2 := hashPasswordResetToken(token)
	assert.Equal(t, hash, hash2)

	// Verify hash is URL-safe
	_, err = base64.URLEncoding.DecodeString(hash)
	assert.NoError(t, err)
}

func TestPasswordResetToken_SecurityProperties(t *testing.T) {
	// Verify avalanche effect: small change in input = large change in hash
	token1 := "reset1"
	token2 := "reset2"

	hash1 := hashPasswordResetToken(token1)
	hash2 := hashPasswordResetToken(token2)

	// Count different characters
	diffCount := 0
	minLen := len(hash1)
	if len(hash2) < minLen {
		minLen = len(hash2)
	}
	for i := 0; i < minLen; i++ {
		if hash1[i] != hash2[i] {
			diffCount++
		}
	}

	percentDifferent := float64(diffCount) / float64(minLen) * 100
	assert.True(t, percentDifferent > 40, "small input change should cause large hash change")
}

func TestPasswordResetErrors(t *testing.T) {
	t.Run("ErrPasswordResetTokenNotFound", func(t *testing.T) {
		assert.NotNil(t, ErrPasswordResetTokenNotFound)
		assert.Contains(t, ErrPasswordResetTokenNotFound.Error(), "not found")
	})

	t.Run("ErrPasswordResetTokenExpired", func(t *testing.T) {
		assert.NotNil(t, ErrPasswordResetTokenExpired)
		assert.Contains(t, ErrPasswordResetTokenExpired.Error(), "expired")
	})

	t.Run("ErrPasswordResetTokenUsed", func(t *testing.T) {
		assert.NotNil(t, ErrPasswordResetTokenUsed)
		assert.Contains(t, ErrPasswordResetTokenUsed.Error(), "used")
	})

	t.Run("ErrSMTPNotConfigured", func(t *testing.T) {
		assert.NotNil(t, ErrSMTPNotConfigured)
		assert.Contains(t, ErrSMTPNotConfigured.Error(), "SMTP")
	})

	t.Run("ErrEmailSendFailed", func(t *testing.T) {
		assert.NotNil(t, ErrEmailSendFailed)
		assert.Contains(t, ErrEmailSendFailed.Error(), "email")
	})

	t.Run("ErrInvalidRedirectURL", func(t *testing.T) {
		assert.NotNil(t, ErrInvalidRedirectURL)
		assert.Contains(t, ErrInvalidRedirectURL.Error(), "redirect")
	})

	t.Run("ErrPasswordResetTooSoon", func(t *testing.T) {
		assert.NotNil(t, ErrPasswordResetTooSoon)
		assert.Contains(t, ErrPasswordResetTooSoon.Error(), "too recently")
	})
}

func TestPasswordResetToken_Struct(t *testing.T) {
	t.Run("creates token with all fields", func(t *testing.T) {
		now := time.Now()
		usedAt := now.Add(-time.Hour)

		token := &PasswordResetToken{
			ID:        "token-123",
			UserID:    "user-456",
			TokenHash: "hashed-value",
			ExpiresAt: now.Add(time.Hour),
			UsedAt:    &usedAt,
			CreatedAt: now,
		}

		assert.Equal(t, "token-123", token.ID)
		assert.Equal(t, "user-456", token.UserID)
		assert.Equal(t, "hashed-value", token.TokenHash)
		assert.True(t, token.ExpiresAt.After(now))
		assert.NotNil(t, token.UsedAt)
		assert.Equal(t, now, token.CreatedAt)
	})

	t.Run("creates token without UsedAt", func(t *testing.T) {
		token := &PasswordResetToken{
			ID:        "token-789",
			UserID:    "user-abc",
			TokenHash: "hash-xyz",
			ExpiresAt: time.Now().Add(time.Hour),
			CreatedAt: time.Now(),
		}

		assert.Nil(t, token.UsedAt)
	})

	t.Run("zero value", func(t *testing.T) {
		var token PasswordResetToken
		assert.Empty(t, token.ID)
		assert.Empty(t, token.UserID)
		assert.Empty(t, token.TokenHash)
		assert.Nil(t, token.UsedAt)
		assert.True(t, token.ExpiresAt.IsZero())
		assert.True(t, token.CreatedAt.IsZero())
	})
}

func TestPasswordResetTokenWithPlaintext_Struct(t *testing.T) {
	t.Run("embeds PasswordResetToken", func(t *testing.T) {
		now := time.Now()
		tokenWithPlaintext := &PasswordResetTokenWithPlaintext{
			PasswordResetToken: PasswordResetToken{
				ID:        "token-embed",
				UserID:    "user-embed",
				TokenHash: "hash-embed",
				ExpiresAt: now.Add(time.Hour),
				CreatedAt: now,
			},
			PlaintextToken: "plaintext-secret-token",
		}

		// Embedded fields accessible directly
		assert.Equal(t, "token-embed", tokenWithPlaintext.ID)
		assert.Equal(t, "user-embed", tokenWithPlaintext.UserID)
		assert.Equal(t, "hash-embed", tokenWithPlaintext.TokenHash)
		assert.Equal(t, "plaintext-secret-token", tokenWithPlaintext.PlaintextToken)
	})

	t.Run("zero value", func(t *testing.T) {
		var token PasswordResetTokenWithPlaintext
		assert.Empty(t, token.ID)
		assert.Empty(t, token.PlaintextToken)
	})
}

func TestNewPasswordResetRepository(t *testing.T) {
	t.Run("creates with nil db", func(t *testing.T) {
		repo := NewPasswordResetRepository(nil)
		require.NotNil(t, repo)
		assert.Nil(t, repo.db)
	})
}

func TestNewPasswordResetService(t *testing.T) {
	t.Run("creates with all nil dependencies", func(t *testing.T) {
		svc := NewPasswordResetService(nil, nil, nil, time.Hour, "https://example.com")
		require.NotNil(t, svc)
		assert.Nil(t, svc.repo)
		assert.Nil(t, svc.userRepo)
		assert.Nil(t, svc.emailSender)
		assert.Equal(t, time.Hour, svc.tokenExpiry)
		assert.Equal(t, "https://example.com", svc.baseURL)
	})

	t.Run("creates with custom token expiry", func(t *testing.T) {
		expiry := 30 * time.Minute
		svc := NewPasswordResetService(nil, nil, nil, expiry, "")
		assert.Equal(t, expiry, svc.tokenExpiry)
	})

	t.Run("creates with custom base URL", func(t *testing.T) {
		baseURL := "https://custom.example.com/reset"
		svc := NewPasswordResetService(nil, nil, nil, time.Hour, baseURL)
		assert.Equal(t, baseURL, svc.baseURL)
	})
}

func TestPasswordResetService_RequestPasswordReset_NoSMTP(t *testing.T) {
	t.Run("returns error when email sender is nil", func(t *testing.T) {
		svc := NewPasswordResetService(nil, nil, nil, time.Hour, "https://example.com")

		err := svc.RequestPasswordReset(context.TODO(), "test@example.com", "")
		assert.ErrorIs(t, err, ErrSMTPNotConfigured)
	})
}

func TestPasswordResetService_RequestPasswordReset_InvalidRedirectURL(t *testing.T) {
	// Create a mock email sender
	mockSender := &mockPasswordResetEmailSender{}
	svc := NewPasswordResetService(nil, nil, mockSender, time.Hour, "https://example.com")

	// Test invalid redirect URLs that should be rejected before user lookup.
	// Note: Valid URLs would proceed to user lookup and panic with nil userRepo,
	// so we only test the invalid cases here.
	testCases := []struct {
		name       string
		redirectTo string
	}{
		{"ftp URL is invalid", "ftp://example.com/reset"},
		{"relative path is invalid", "/reset-password"},
		{"no scheme is invalid", "example.com/reset"},
		{"invalid URL syntax", "://invalid"},
		{"javascript URL is invalid", "javascript:alert(1)"},
		{"data URL is invalid", "data:text/html,<h1>hi</h1>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.RequestPasswordReset(context.TODO(), "test@example.com", tc.redirectTo)
			assert.ErrorIs(t, err, ErrInvalidRedirectURL)
		})
	}
}

// mockPasswordResetEmailSender implements EmailService for testing
type mockPasswordResetEmailSender struct {
	sentTo    string
	sentToken string
	sentLink  string
	sendError error
}

func (m *mockPasswordResetEmailSender) SendPasswordReset(ctx context.Context, to, token, link string) error {
	m.sentTo = to
	m.sentToken = token
	m.sentLink = link
	return m.sendError
}

func (m *mockPasswordResetEmailSender) SendMagicLink(ctx context.Context, to, token, link string) error {
	return m.sendError
}

func (m *mockPasswordResetEmailSender) SendVerificationEmail(ctx context.Context, to, token, link string) error {
	return m.sendError
}

func (m *mockPasswordResetEmailSender) SendInvitationEmail(ctx context.Context, to, inviterName, inviteLink string) error {
	return m.sendError
}

func (m *mockPasswordResetEmailSender) Send(ctx context.Context, to, subject, body string) error {
	return m.sendError
}

func (m *mockPasswordResetEmailSender) IsConfigured() bool {
	return true
}

func TestPasswordResetEmailSenderInterface(t *testing.T) {
	t.Run("mock implements interface", func(t *testing.T) {
		var sender EmailService = &mockPasswordResetEmailSender{}
		assert.NotNil(t, sender)
	})

	t.Run("mock captures sent values", func(t *testing.T) {
		mock := &mockPasswordResetEmailSender{}
		err := mock.SendPasswordReset(context.TODO(), "user@test.com", "token123", "https://link.com")

		assert.NoError(t, err)
		assert.Equal(t, "user@test.com", mock.sentTo)
		assert.Equal(t, "token123", mock.sentToken)
		assert.Equal(t, "https://link.com", mock.sentLink)
	})

	t.Run("mock returns configured error", func(t *testing.T) {
		expectedErr := errors.New("send failed")
		mock := &mockPasswordResetEmailSender{sendError: expectedErr}

		err := mock.SendPasswordReset(context.TODO(), "user@test.com", "token", "link")
		assert.Equal(t, expectedErr, err)
	})
}
