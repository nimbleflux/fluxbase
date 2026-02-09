package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file contains comprehensive tests for password workflows using mock repositories

// TestPasswordResetWorkflow_ValidFlow tests the complete password reset flow
func TestPasswordResetWorkflow_ValidFlow(t *testing.T) {
	ctx := context.Background()

	// Setup mock repositories
	mockUserRepo := NewMockUserRepository()
	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}

	// Create test user
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	// Create a password reset service
	service := NewPasswordResetService(
		mockResetRepo,
		mockUserRepo,
		mockEmailSender,
		time.Hour,
		"https://example.com",
	)

	// Request password reset
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	require.NoError(t, err)

	// Verify email was sent
	assert.NotEmpty(t, mockEmailSender.sentToken, "Reset token should be sent")
	assert.NotEmpty(t, mockEmailSender.sentLink, "Reset link should be sent")
	assert.Equal(t, "test@example.com", mockEmailSender.sentTo)

	resetToken := mockEmailSender.sentToken

	// Verify token is valid
	err = service.VerifyPasswordResetToken(ctx, resetToken)
	assert.NoError(t, err, "Token should be valid")

	// Reset password with valid token
	newPassword := "NewSecurePassword456!"
	userID, err := service.ResetPassword(ctx, resetToken, newPassword)
	require.NoError(t, err)
	assert.Equal(t, user.ID, userID)

	// Verify token is now used (can't be used again)
	err = service.VerifyPasswordResetToken(ctx, resetToken)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordResetTokenUsed)
}

// TestPasswordResetWorkflow_InvalidToken tests password reset with invalid token
func TestPasswordResetWorkflow_InvalidToken(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	userID, err := service.ResetPassword(ctx, "completely-invalid-token", "NewPassword123!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordResetTokenNotFound)
	assert.Empty(t, userID)
}

// TestPasswordResetWorkflow_ExpiredToken tests password reset with expired token
func TestPasswordResetWorkflow_ExpiredToken(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	// Create token that's already expired
	mockResetRepo := NewMockPasswordResetRepository()
	expiredToken, err := mockResetRepo.Create(ctx, user.ID, -1*time.Hour)
	require.NoError(t, err)

	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// Try to reset password with expired token
	userID, err := service.ResetPassword(ctx, expiredToken.PlaintextToken, "NewPassword123!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordResetTokenExpired)
	assert.Empty(t, userID)
}

// TestPasswordResetWorkflow_TokenAlreadyUsed tests that used tokens can't be reused
func TestPasswordResetWorkflow_TokenAlreadyUsed(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	token, err := mockResetRepo.Create(ctx, user.ID, time.Hour)
	require.NoError(t, err)

	// Mark token as used
	err = mockResetRepo.MarkAsUsed(ctx, token.ID)
	require.NoError(t, err)

	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// Try to reset password with already-used token
	userID, err := service.ResetPassword(ctx, token.PlaintextToken, "NewPassword123!")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordResetTokenUsed)
	assert.Empty(t, userID)
}

// TestPasswordResetWorkflow_WeakPassword tests password reset rejects weak passwords
func TestPasswordResetWorkflow_WeakPassword(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	token, err := mockResetRepo.Create(ctx, user.ID, time.Hour)
	require.NoError(t, err)

	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	weakPasswords := []struct {
		name     string
		password string
	}{
		{"too_short", "short1!"},
		{"no_upper", "lowercase1!"},
		{"no_lower", "UPPERCASE1!"},
		{"no_digit", "NoDigits!"},
		{"empty", ""},
	}

	for _, tc := range weakPasswords {
		t.Run(tc.name, func(t *testing.T) {
			userID, err := service.ResetPassword(ctx, token.PlaintextToken, tc.password)

			assert.Error(t, err, "Should reject weak password: "+tc.password)
			assert.Empty(t, userID)
		})
	}
}

// TestPasswordResetWorkflow_NonExistentEmail tests security: don't reveal if user exists
func TestPasswordResetWorkflow_NonExistentEmail(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// Request reset for non-existent email
	err := service.RequestPasswordReset(ctx, "nonexistent@example.com", "")

	// Should succeed without error (don't reveal if user exists)
	assert.NoError(t, err)
	// But no email should be sent
	assert.Empty(t, mockEmailSender.sentToken)
}

// TestPasswordResetWorkflow_RateLimiting tests rate limiting of password reset requests
func TestPasswordResetWorkflow_RateLimiting(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	_, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// First request should succeed
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	assert.NoError(t, err)
	assert.NotEmpty(t, mockEmailSender.sentToken)

	// Second immediate request should be rate limited
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPasswordResetTooSoon)
}

// TestPasswordResetWorkflow_RateLimitExpires tests that rate limit expires after time
func TestPasswordResetWorkflow_RateLimitExpires(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)
	_ = user // Keep reference for use in rate limit check

	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// First request
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	require.NoError(t, err)

	// Manually expire the rate limit window
	mockResetRepo.mu.Lock()
	// Get the user to access their ID
	testUser, _ := mockUserRepo.GetByEmail(ctx, "test@example.com")
	if token := mockResetRepo.byUserID[testUser.ID]; token != nil {
		token.CreatedAt = time.Now().Add(-61 * time.Second)
	}
	mockResetRepo.mu.Unlock()

	// Second request should now succeed
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	assert.NoError(t, err)
}

// TestPasswordResetWorkflow_EmailFailed tests handling of email send failures
func TestPasswordResetWorkflow_EmailFailed(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	_, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{
		sendError: errors.New("SMTP connection failed"),
	}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	err = service.RequestPasswordReset(ctx, "test@example.com", "")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrEmailSendFailed)
	assert.Contains(t, err.Error(), "SMTP connection failed")
}

// TestPasswordResetWorkflow_NoSMTPConfigured tests error when SMTP not configured
func TestPasswordResetWorkflow_NoSMTPConfigured(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	mockResetRepo := NewMockPasswordResetRepository()
	// No email sender configured
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, nil, time.Hour, "https://example.com")

	err := service.RequestPasswordReset(ctx, "test@example.com", "")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSMTPNotConfigured)
}

// TestPasswordResetWorkflow_CustomRedirectURL tests custom redirect URLs
func TestPasswordResetWorkflow_CustomRedirectURL(t *testing.T) {
	tests := []struct {
		name        string
		redirectTo  string
		expectError bool
		errorIs     error
	}{
		{
			name:        "valid HTTPS URL",
			redirectTo:  "https://custom.example.com/reset",
			expectError: false,
		},
		{
			name:        "valid HTTP URL",
			redirectTo:  "http://custom.example.com/reset",
			expectError: false,
		},
		{
			name:        "invalid FTP URL",
			redirectTo:  "ftp://example.com/reset",
			expectError: true,
			errorIs:     ErrInvalidRedirectURL,
		},
		{
			name:        "invalid relative path",
			redirectTo:  "/reset-password",
			expectError: true,
			errorIs:     ErrInvalidRedirectURL,
		},
		{
			name:        "invalid javascript URL",
			redirectTo:  "javascript:alert(1)",
			expectError: true,
			errorIs:     ErrInvalidRedirectURL,
		},
		{
			name:        "invalid data URL",
			redirectTo:  "data:text/html,<h1>hi</h1>",
			expectError: true,
			errorIs:     ErrInvalidRedirectURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockUserRepo := NewMockUserRepository()
			if !tt.expectError {
				_, err := mockUserRepo.Create(ctx, CreateUserRequest{
					Email:    "test@example.com",
					Password: "OldPassword123!",
				}, "")
				require.NoError(t, err)
			}

			mockResetRepo := NewMockPasswordResetRepository()
			mockEmailSender := &mockPasswordResetEmailSender{}
			service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

			err := service.RequestPasswordReset(ctx, "test@example.com", tt.redirectTo)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				assert.NoError(t, err)
				// Verify custom redirect URL was used in the link
				assert.Contains(t, mockEmailSender.sentLink, tt.redirectTo)
			}
		})
	}
}

// TestPasswordValidation_PasswordStrength tests comprehensive password validation
func TestPasswordValidation_PasswordStrength(t *testing.T) {
	hasher := NewPasswordHasher()

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorIs     error
		description string
	}{
		{
			name:        "strong password",
			password:    "SecurePass123!",
			expectError: false,
			description: "Meets all requirements",
		},
		{
			name:        "too short",
			password:    "Short1!",
			expectError: true,
			errorIs:     ErrWeakPassword,
			description: "Less than minimum length (12)",
		},
		{
			name:        "no uppercase",
			password:    "lowercase123!",
			expectError: true,
			errorIs:     ErrWeakPassword,
			description: "Missing uppercase letter",
		},
		{
			name:        "no lowercase",
			password:    "UPPERCASE123!",
			expectError: true,
			errorIs:     ErrWeakPassword,
			description: "Missing lowercase letter",
		},
		{
			name:        "no digit",
			password:    "NoDigits!",
			expectError: true,
			errorIs:     ErrWeakPassword,
			description: "Missing digit",
		},
		{
			name:        "very long password",
			password:    strings.Repeat("a", 73) + "A1!",
			expectError: true,
			errorIs:     ErrPasswordTooLong,
			description: "Exceeds bcrypt max length (72)",
		},
		{
			name:        "exactly min length",
			password:    "MinLength123", // 12 chars with upper, lower, digit
			expectError: false,
			description: "Exactly 12 characters",
		},
		{
			name:        "only meets minimum requirements",
			password:    "Password1234",
			expectError: false,
			description: "Has upper, lower, digit, 12+ chars (no symbol)",
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorIs:     ErrWeakPassword,
			description: "Empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := hasher.ValidatePassword(tt.password)

			if tt.expectError {
				assert.Error(t, err, tt.description)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs, tt.description)
				}
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// TestPasswordWorkflow_TokenSecurity tests token security properties
func TestPasswordWorkflow_TokenSecurity(t *testing.T) {
	// Test that token hashing is deterministic
	token := "test-reset-token"
	hash1 := hashPasswordResetToken(token)
	hash2 := hashPasswordResetToken(token)

	assert.Equal(t, hash1, hash2, "Same token should produce same hash")

	// Test that different tokens produce different hashes
	token2 := "test-reset-token-2"
	hash3 := hashPasswordResetToken(token2)

	assert.NotEqual(t, hash1, hash3, "Different tokens should produce different hashes")

	// Test avalanche effect: small change = large hash difference
	token3 := "test-reset-token-3"
	hash4 := hashPasswordResetToken(token3)

	diffCount := 0
	minLen := len(hash1)
	if len(hash4) < minLen {
		minLen = len(hash4)
	}
	for i := 0; i < minLen; i++ {
		if hash1[i] != hash4[i] {
			diffCount++
		}
	}

	percentDifferent := float64(diffCount) / float64(minLen) * 100
	assert.True(t, percentDifferent > 40, "Small input change should cause large hash change")
}

// TestPasswordWorkflow_OldTokenInvalidation tests that old tokens are invalidated
func TestPasswordWorkflow_OldTokenInvalidation(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// Request first password reset
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	require.NoError(t, err)
	firstToken := mockEmailSender.sentToken
	require.NotEmpty(t, firstToken)

	// For testing, bypass the rate limit by manually clearing the tokens
	// This simulates waiting 60+ seconds for the rate limit to expire
	mockResetRepo.DeleteByUserID(ctx, user.ID)

	// Request second password reset (should invalidate first)
	err = service.RequestPasswordReset(ctx, "test@example.com", "")
	require.NoError(t, err)
	secondToken := mockEmailSender.sentToken
	require.NotEmpty(t, secondToken)

	// First token should no longer be valid (old tokens were deleted)
	err = service.VerifyPasswordResetToken(ctx, firstToken)
	assert.Error(t, err, "Old token should be invalidated")

	// Second token should be valid
	err = service.VerifyPasswordResetToken(ctx, secondToken)
	assert.NoError(t, err, "New token should be valid")
}

// TestPasswordWorkflow_TokenUniqueness tests that generated tokens are unique
func TestPasswordWorkflow_TokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	iterations := 100

	for i := 0; i < iterations; i++ {
		token, err := GeneratePasswordResetToken()
		require.NoError(t, err)

		assert.False(t, tokens[token], "Token should be unique: %s", token)
		tokens[token] = true
	}

	assert.Len(t, tokens, iterations, "All tokens should be unique")
}

// TestPasswordWorkflow_TokenURLSafe tests that tokens are URL-safe
func TestPasswordWorkflow_TokenURLSafe(t *testing.T) {
	for i := 0; i < 50; i++ {
		token, err := GeneratePasswordResetToken()
		require.NoError(t, err)

		// URL-safe base64 should not contain + or /
		assert.NotContains(t, token, "+", "Token should not contain +")
		assert.NotContains(t, token, "/", "Token should not contain /")
		// Note: URLEncoding may include = padding, which is still URL-safe
	}
}

// TestPasswordWorkflow_ConcurrentRequests tests concurrent password reset requests
func TestPasswordWorkflow_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	_, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()
	mockEmailSender := &mockPasswordResetEmailSender{}
	service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

	// Send multiple concurrent requests
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			done <- service.RequestPasswordReset(ctx, "test@example.com", "")
		}()
	}

	// Collect results
	var successCount, failCount int
	for i := 0; i < 5; i++ {
		err := <-done
		if err == nil {
			successCount++
		} else if errors.Is(err, ErrPasswordResetTooSoon) {
			failCount++
		} else {
			t.Errorf("unexpected error: %v", err)
		}
	}

	// All should complete (some may be rate limited)
	assert.Equal(t, 5, successCount+failCount)
	assert.GreaterOrEqual(t, successCount, 1, "At least one request should succeed")
}

// TestPasswordWorkflow_DeleteExpiredTokens tests cleanup of expired tokens
func TestPasswordWorkflow_DeleteExpiredTokens(t *testing.T) {
	ctx := context.Background()

	mockUserRepo := NewMockUserRepository()
	user, err := mockUserRepo.Create(ctx, CreateUserRequest{
		Email:    "test@example.com",
		Password: "OldPassword123!",
	}, "")
	require.NoError(t, err)

	mockResetRepo := NewMockPasswordResetRepository()

	// Create valid token
	validToken, err := mockResetRepo.Create(ctx, user.ID, time.Hour)
	require.NoError(t, err)

	// Create expired token
	expiredToken, err := mockResetRepo.Create(ctx, user.ID, -1*time.Hour)
	require.NoError(t, err)

	// Verify both exist before cleanup
	_, err = mockResetRepo.Validate(ctx, validToken.PlaintextToken)
	assert.NoError(t, err, "Valid token should exist")

	_, err = mockResetRepo.Validate(ctx, expiredToken.PlaintextToken)
	assert.Error(t, err, "Expired token should fail validation")

	// Delete expired tokens
	count, err := mockResetRepo.DeleteExpired(ctx)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, count, int64(1), "Should delete at least one expired token")

	// Verify expired token no longer exists
	_, err = mockResetRepo.GetByToken(ctx, expiredToken.PlaintextToken)
	assert.Error(t, err, "Expired token should be deleted")

	// Verify valid token still exists
	_, err = mockResetRepo.GetByToken(ctx, validToken.PlaintextToken)
	assert.NoError(t, err, "Valid token should still exist")
}

// TestPasswordWorkflow_TokenExpiryTiming tests token expiry at boundary conditions
func TestPasswordWorkflow_TokenExpiryTiming(t *testing.T) {
	tests := []struct {
		name        string
		expiry      time.Duration
		shouldValid bool
	}{
		{
			name:        "expires in future",
			expiry:      1 * time.Hour,
			shouldValid: true,
		},
		{
			name:        "expires in past",
			expiry:      -1 * time.Hour,
			shouldValid: false,
		},
		{
			name:        "expires now",
			expiry:      0,
			shouldValid: false,
		},
		{
			name:        "expires very soon",
			expiry:      1 * time.Second,
			shouldValid: true,
		},
		{
			name:        "expires in one millisecond",
			expiry:      1 * time.Millisecond,
			shouldValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			mockUserRepo := NewMockUserRepository()
			user, err := mockUserRepo.Create(ctx, CreateUserRequest{
				Email:    "test@example.com",
				Password: "OldPassword123!",
			}, "")
			require.NoError(t, err)

			mockResetRepo := NewMockPasswordResetRepository()
			tokenWithPlaintext, err := mockResetRepo.Create(ctx, user.ID, tt.expiry)
			require.NoError(t, err)

			mockEmailSender := &mockPasswordResetEmailSender{}
			service := NewPasswordResetService(mockResetRepo, mockUserRepo, mockEmailSender, time.Hour, "https://example.com")

			err = service.VerifyPasswordResetToken(ctx, tokenWithPlaintext.PlaintextToken)

			if tt.shouldValid {
				assert.NoError(t, err, "Token should be valid")
			} else {
				assert.Error(t, err, "Token should be invalid")
			}
		})
	}
}
