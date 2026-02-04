package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// generateSecurePassword Tests
// =============================================================================

func TestGenerateSecurePassword_Length(t *testing.T) {
	tests := []int{8, 12, 16, 24, 32}

	for _, length := range tests {
		t.Run("length_"+string(rune('0'+length/10))+string(rune('0'+length%10)), func(t *testing.T) {
			password, err := generateSecurePassword(length)

			require.NoError(t, err)
			assert.Len(t, password, length)
		})
	}
}

func TestGenerateSecurePassword_Uniqueness(t *testing.T) {
	// Generate multiple passwords and ensure they're different
	passwords := make(map[string]bool)

	for i := 0; i < 100; i++ {
		password, err := generateSecurePassword(16)
		require.NoError(t, err)

		// Should not have seen this password before
		assert.False(t, passwords[password], "Password collision detected")
		passwords[password] = true
	}

	// Should have 100 unique passwords
	assert.Len(t, passwords, 100)
}

func TestGenerateSecurePassword_NotEmpty(t *testing.T) {
	password, err := generateSecurePassword(8)

	require.NoError(t, err)
	assert.NotEmpty(t, password)
}

func TestGenerateSecurePassword_Printable(t *testing.T) {
	// Base64 URL encoding should produce printable characters
	for i := 0; i < 10; i++ {
		password, err := generateSecurePassword(16)
		require.NoError(t, err)

		for _, c := range password {
			// Base64 URL safe characters: A-Z, a-z, 0-9, -, _
			isAlphaNum := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')
			isUrlSafe := c == '-' || c == '_'
			assert.True(t, isAlphaNum || isUrlSafe, "Non-printable character found: %c", c)
		}
	}
}

func TestGenerateSecurePassword_MinimumLength(t *testing.T) {
	// Even with length 1, should work
	password, err := generateSecurePassword(1)

	require.NoError(t, err)
	assert.Len(t, password, 1)
}

// =============================================================================
// EnrichedUser Type Tests
// =============================================================================

func TestEnrichedUser_FieldsExist(t *testing.T) {
	// Test that EnrichedUser has all expected fields
	user := EnrichedUser{
		ID:             "user-123",
		Email:          "test@example.com",
		EmailVerified:  true,
		Role:           "admin",
		Provider:       "email",
		ActiveSessions: 2,
		LastSignIn:     nil,
		IsLocked:       false,
		UserMetadata:   map[string]interface{}{"name": "Test"},
		AppMetadata:    map[string]interface{}{"plan": "pro"},
	}

	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.True(t, user.EmailVerified)
	assert.Equal(t, "admin", user.Role)
	assert.Equal(t, "email", user.Provider)
	assert.Equal(t, 2, user.ActiveSessions)
	assert.False(t, user.IsLocked)
}

// =============================================================================
// InviteUserRequest Tests
// =============================================================================

func TestInviteUserRequest_Defaults(t *testing.T) {
	req := InviteUserRequest{
		Email: "new@example.com",
	}

	// Email should be set, role should be empty (defaults applied in service)
	assert.Equal(t, "new@example.com", req.Email)
	assert.Empty(t, req.Role)
	assert.Empty(t, req.Password)
}

func TestInviteUserRequest_WithPassword(t *testing.T) {
	req := InviteUserRequest{
		Email:    "new@example.com",
		Role:     "admin",
		Password: "custom-password",
	}

	assert.Equal(t, "new@example.com", req.Email)
	assert.Equal(t, "admin", req.Role)
	assert.Equal(t, "custom-password", req.Password)
}

// =============================================================================
// UpdateAdminUserRequest Tests
// =============================================================================

func TestUpdateAdminUserRequest_AllFields(t *testing.T) {
	email := "updated@example.com"
	role := "superadmin"
	password := "new-password"

	req := UpdateAdminUserRequest{
		Email:        &email,
		Role:         &role,
		Password:     &password,
		UserMetadata: map[string]interface{}{"key": "value"},
	}

	assert.NotNil(t, req.Email)
	assert.Equal(t, "updated@example.com", *req.Email)
	assert.NotNil(t, req.Role)
	assert.Equal(t, "superadmin", *req.Role)
	assert.NotNil(t, req.Password)
	assert.Equal(t, "new-password", *req.Password)
	assert.NotNil(t, req.UserMetadata)
}

func TestUpdateAdminUserRequest_PartialUpdate(t *testing.T) {
	email := "updated@example.com"

	req := UpdateAdminUserRequest{
		Email: &email,
		// Other fields are nil
	}

	assert.NotNil(t, req.Email)
	assert.Nil(t, req.Role)
	assert.Nil(t, req.Password)
	assert.Nil(t, req.UserMetadata)
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkGenerateSecurePassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateSecurePassword(16)
	}
}

func BenchmarkGenerateSecurePassword_Long(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateSecurePassword(64)
	}
}

// =============================================================================
// NewUserManagementService Tests
// =============================================================================

func TestNewUserManagementService_AllNil(t *testing.T) {
	svc := NewUserManagementService(nil, nil, nil, nil, "")

	assert.NotNil(t, svc)
	assert.Nil(t, svc.userRepo)
	assert.Nil(t, svc.sessionRepo)
	assert.Nil(t, svc.passwordHasher)
	assert.Nil(t, svc.emailService)
	assert.Empty(t, svc.baseURL)
}

func TestNewUserManagementService_WithBaseURL(t *testing.T) {
	svc := NewUserManagementService(nil, nil, nil, nil, "https://app.example.com")

	assert.NotNil(t, svc)
	assert.Equal(t, "https://app.example.com", svc.baseURL)
}

func TestNewUserManagementService_WithDependencies(t *testing.T) {
	userRepo := NewUserRepository(nil)
	sessionRepo := NewSessionRepository(nil)
	passwordHasher := NewPasswordHasher()

	svc := NewUserManagementService(userRepo, sessionRepo, passwordHasher, nil, "https://api.example.com")

	assert.NotNil(t, svc)
	assert.Equal(t, userRepo, svc.userRepo)
	assert.Equal(t, sessionRepo, svc.sessionRepo)
	assert.Equal(t, passwordHasher, svc.passwordHasher)
	assert.Equal(t, "https://api.example.com", svc.baseURL)
}

// =============================================================================
// InviteUserResponse Tests
// =============================================================================

func TestInviteUserResponse_WithUser(t *testing.T) {
	user := &User{
		ID:    "user-123",
		Email: "test@example.com",
		Role:  "authenticated",
	}

	resp := InviteUserResponse{
		User:              user,
		TemporaryPassword: "temp-pass-123",
		EmailSent:         false,
		Message:           "User created successfully",
	}

	assert.Equal(t, user, resp.User)
	assert.Equal(t, "temp-pass-123", resp.TemporaryPassword)
	assert.False(t, resp.EmailSent)
	assert.Equal(t, "User created successfully", resp.Message)
}

func TestInviteUserResponse_EmailSent(t *testing.T) {
	user := &User{
		ID:    "user-456",
		Email: "invited@example.com",
	}

	resp := InviteUserResponse{
		User:      user,
		EmailSent: true,
		Message:   "Invitation email sent to invited@example.com",
	}

	assert.Equal(t, user, resp.User)
	assert.Empty(t, resp.TemporaryPassword) // Should be empty when email is sent
	assert.True(t, resp.EmailSent)
	assert.Contains(t, resp.Message, "Invitation email sent")
}

func TestInviteUserResponse_Defaults(t *testing.T) {
	resp := InviteUserResponse{}

	assert.Nil(t, resp.User)
	assert.Empty(t, resp.TemporaryPassword)
	assert.False(t, resp.EmailSent)
	assert.Empty(t, resp.Message)
}

// =============================================================================
// UserManagementService Struct Tests
// =============================================================================

func TestUserManagementService_FieldsAccessible(t *testing.T) {
	svc := &UserManagementService{
		baseURL: "https://test.example.com",
	}

	assert.Equal(t, "https://test.example.com", svc.baseURL)
	assert.Nil(t, svc.userRepo)
	assert.Nil(t, svc.sessionRepo)
	assert.Nil(t, svc.passwordHasher)
	assert.Nil(t, svc.emailService)
}

// =============================================================================
// EnrichedUser Additional Tests
// =============================================================================

func TestEnrichedUser_Defaults(t *testing.T) {
	user := EnrichedUser{}

	assert.Empty(t, user.ID)
	assert.Empty(t, user.Email)
	assert.False(t, user.EmailVerified)
	assert.Empty(t, user.Role)
	assert.Empty(t, user.Provider)
	assert.Equal(t, 0, user.ActiveSessions)
	assert.Nil(t, user.LastSignIn)
	assert.False(t, user.IsLocked)
	assert.Nil(t, user.UserMetadata)
	assert.Nil(t, user.AppMetadata)
}

func TestEnrichedUser_ProviderTypes(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"email provider", "email"},
		{"invite pending", "invite_pending"},
		{"magic link", "magic_link"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := EnrichedUser{
				Provider: tt.provider,
			}
			assert.Equal(t, tt.provider, user.Provider)
		})
	}
}

func TestEnrichedUser_LockedStatus(t *testing.T) {
	t.Run("locked user", func(t *testing.T) {
		user := EnrichedUser{
			ID:       "user-locked",
			IsLocked: true,
		}
		assert.True(t, user.IsLocked)
	})

	t.Run("unlocked user", func(t *testing.T) {
		user := EnrichedUser{
			ID:       "user-unlocked",
			IsLocked: false,
		}
		assert.False(t, user.IsLocked)
	})
}

// =============================================================================
// InviteUserRequest Additional Tests
// =============================================================================

func TestInviteUserRequest_SkipEmail(t *testing.T) {
	req := InviteUserRequest{
		Email:     "user@example.com",
		SkipEmail: true,
	}

	assert.True(t, req.SkipEmail)
}

func TestInviteUserRequest_AllFields(t *testing.T) {
	req := InviteUserRequest{
		Email:     "user@example.com",
		Role:      "admin",
		Password:  "custom-password-123",
		SkipEmail: false,
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "admin", req.Role)
	assert.Equal(t, "custom-password-123", req.Password)
	assert.False(t, req.SkipEmail)
}
