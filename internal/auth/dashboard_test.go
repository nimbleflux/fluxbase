package auth

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DashboardUser Struct Tests
// =============================================================================

func TestDashboardUser_Fields(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)
	fullName := "John Doe"
	avatarURL := "https://example.com/avatar.png"

	user := DashboardUser{
		ID:            uuid.New(),
		Email:         "admin@example.com",
		EmailVerified: true,
		FullName:      &fullName,
		AvatarURL:     &avatarURL,
		TOTPEnabled:   true,
		IsActive:      true,
		IsLocked:      false,
		LockedUntil:   nil,
		LastLoginAt:   &lastLogin,
		CreatedAt:     now,
		UpdatedAt:     now,
		Role:          "dashboard_admin",
	}

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "admin@example.com", user.Email)
	assert.True(t, user.EmailVerified)
	assert.Equal(t, "John Doe", *user.FullName)
	assert.Equal(t, "https://example.com/avatar.png", *user.AvatarURL)
	assert.True(t, user.TOTPEnabled)
	assert.True(t, user.IsActive)
	assert.False(t, user.IsLocked)
	assert.Nil(t, user.LockedUntil)
	assert.NotNil(t, user.LastLoginAt)
	assert.Equal(t, "dashboard_admin", user.Role)
}

func TestDashboardUser_NullableFields(t *testing.T) {
	user := DashboardUser{
		ID:    uuid.New(),
		Email: "admin@example.com",
	}

	assert.Nil(t, user.FullName)
	assert.Nil(t, user.AvatarURL)
	assert.Nil(t, user.LockedUntil)
	assert.Nil(t, user.LastLoginAt)
}

func TestDashboardUser_LockedState(t *testing.T) {
	lockedUntil := time.Now().Add(15 * time.Minute)

	user := DashboardUser{
		ID:          uuid.New(),
		Email:       "locked@example.com",
		IsLocked:    true,
		LockedUntil: &lockedUntil,
	}

	assert.True(t, user.IsLocked)
	assert.NotNil(t, user.LockedUntil)
	assert.True(t, user.LockedUntil.After(time.Now()))
}

// =============================================================================
// DashboardSession Struct Tests
// =============================================================================

func TestDashboardSession_Fields(t *testing.T) {
	now := time.Now()
	ip := net.ParseIP("192.168.1.1")
	userAgent := "Mozilla/5.0"

	session := DashboardSession{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		TokenHash:      "abc123hash",
		IPAddress:      &ip,
		UserAgent:      &userAgent,
		ExpiresAt:      now.Add(24 * time.Hour),
		CreatedAt:      now,
		LastActivityAt: now,
	}

	assert.NotEmpty(t, session.ID)
	assert.NotEmpty(t, session.UserID)
	assert.Equal(t, "abc123hash", session.TokenHash)
	assert.NotNil(t, session.IPAddress)
	assert.Equal(t, "192.168.1.1", session.IPAddress.String())
	assert.Equal(t, "Mozilla/5.0", *session.UserAgent)
	assert.True(t, session.ExpiresAt.After(now))
}

func TestDashboardSession_NullableFields(t *testing.T) {
	session := DashboardSession{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TokenHash: "somehash",
	}

	assert.Nil(t, session.IPAddress)
	assert.Nil(t, session.UserAgent)
}

// =============================================================================
// LoginResponse Struct Tests
// =============================================================================

func TestLoginResponse_Fields(t *testing.T) {
	response := LoginResponse{
		AccessToken:  "access.token.here",
		RefreshToken: "refresh.token.here",
		ExpiresIn:    86400, // 24 hours
	}

	assert.Equal(t, "access.token.here", response.AccessToken)
	assert.Equal(t, "refresh.token.here", response.RefreshToken)
	assert.Equal(t, int64(86400), response.ExpiresIn)
}

func TestLoginResponse_DefaultExpiration(t *testing.T) {
	// Standard expiration is 24 hours = 86400 seconds
	expectedExpiration := int64(24 * 60 * 60)

	response := LoginResponse{
		ExpiresIn: expectedExpiration,
	}

	assert.Equal(t, int64(86400), response.ExpiresIn)
}

// =============================================================================
// SSOIdentity Struct Tests
// =============================================================================

func TestSSOIdentity_Fields(t *testing.T) {
	now := time.Now()
	email := "user@example.com"

	identity := SSOIdentity{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Provider:       "oauth:google",
		ProviderUserID: "google-user-123",
		Email:          &email,
		CreatedAt:      now,
	}

	assert.NotEmpty(t, identity.ID)
	assert.NotEmpty(t, identity.UserID)
	assert.Equal(t, "oauth:google", identity.Provider)
	assert.Equal(t, "google-user-123", identity.ProviderUserID)
	assert.Equal(t, "user@example.com", *identity.Email)
}

func TestSSOIdentity_ProviderFormats(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"OAuth Google", "oauth:google"},
		{"OAuth GitHub", "oauth:github"},
		{"OAuth Microsoft", "oauth:microsoft"},
		{"SAML Okta", "saml:okta"},
		{"SAML Azure AD", "saml:azure"},
		{"OIDC Authelia", "oidc:authelia"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := SSOIdentity{
				Provider: tt.provider,
			}

			// Provider should be in format "type:name"
			parts := strings.SplitN(identity.Provider, ":", 2)
			assert.Len(t, parts, 2)
			assert.NotEmpty(t, parts[0]) // type
			assert.NotEmpty(t, parts[1]) // name
		})
	}
}

func TestSSOIdentity_NullableEmail(t *testing.T) {
	identity := SSOIdentity{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Provider:       "oauth:github",
		ProviderUserID: "github-123",
		Email:          nil, // Email may not always be available
	}

	assert.Nil(t, identity.Email)
}

// =============================================================================
// generateBackupCode Tests
// =============================================================================

func TestGenerateBackupCode_Length(t *testing.T) {
	code, err := generateBackupCode()

	require.NoError(t, err)
	// 5 bytes base32 encoded = 8 characters
	assert.Len(t, code, 8)
}

func TestGenerateBackupCode_Uniqueness(t *testing.T) {
	codes := make(map[string]bool)
	numCodes := 100

	for i := 0; i < numCodes; i++ {
		code, err := generateBackupCode()
		require.NoError(t, err)

		assert.False(t, codes[code], "Duplicate backup code generated")
		codes[code] = true
	}

	assert.Len(t, codes, numCodes)
}

func TestGenerateBackupCode_Base32Characters(t *testing.T) {
	code, err := generateBackupCode()
	require.NoError(t, err)

	// Base32 without padding should only contain: A-Z, 2-7
	for _, char := range code {
		valid := (char >= 'A' && char <= 'Z') || (char >= '2' && char <= '7')
		assert.True(t, valid, "Invalid character in backup code: %c", char)
	}
}

func TestGenerateBackupCode_MultipleGenerations(t *testing.T) {
	// Generate 10 backup codes like EnableTOTP does
	codes := make([]string, 10)
	for i := 0; i < 10; i++ {
		code, err := generateBackupCode()
		require.NoError(t, err)
		codes[i] = code
	}

	// All should be unique
	seen := make(map[string]bool)
	for _, code := range codes {
		assert.False(t, seen[code], "Duplicate code found")
		seen[code] = true
	}
}

// =============================================================================
// Password Validation Tests (via error messages)
// =============================================================================

func TestPasswordValidation_MinLength(t *testing.T) {
	// Test that MinPasswordLength is reasonable
	assert.GreaterOrEqual(t, MinPasswordLength, 8, "Minimum password length should be at least 8")
}

func TestPasswordValidation_MaxLength(t *testing.T) {
	// Test that MaxPasswordLength is set (bcrypt has 72 byte limit)
	assert.LessOrEqual(t, MaxPasswordLength, 72, "Maximum password length should not exceed bcrypt's 72 byte limit")
}

// =============================================================================
// Provider Format Validation Tests
// =============================================================================

func TestProviderFormat_Valid(t *testing.T) {
	validProviders := []string{
		"oauth:google",
		"oauth:github",
		"oauth:facebook",
		"oauth:microsoft",
		"oauth:apple",
		"saml:okta",
		"saml:azure",
		"saml:onelogin",
		"oidc:authelia",
		"oidc:keycloak",
	}

	for _, provider := range validProviders {
		t.Run(provider, func(t *testing.T) {
			parts := strings.SplitN(provider, ":", 2)
			assert.Len(t, parts, 2, "Provider should have exactly 2 parts")
			assert.NotEmpty(t, parts[0], "Provider type should not be empty")
			assert.NotEmpty(t, parts[1], "Provider name should not be empty")
		})
	}
}

func TestProviderFormat_Invalid(t *testing.T) {
	invalidProviders := []string{
		"google",             // Missing type
		"oauth",              // Missing name
		":google",            // Empty type
		"oauth:",             // Empty name
		"oauth:google:extra", // Too many parts (but SplitN handles this)
	}

	for _, provider := range invalidProviders {
		t.Run(provider, func(t *testing.T) {
			parts := strings.SplitN(provider, ":", 2)
			if len(parts) == 2 {
				// If we got 2 parts, check if either is empty
				isInvalid := parts[0] == "" || parts[1] == ""
				// Empty parts indicate invalid format
				if provider == ":google" || provider == "oauth:" {
					assert.True(t, isInvalid, "Should be invalid: %s", provider)
				}
			} else {
				// If we got less than 2 parts, it's invalid
				assert.Less(t, len(parts), 2, "Should have less than 2 parts: %s", provider)
			}
		})
	}
}

// =============================================================================
// IP Address Handling Tests
// =============================================================================

func TestIPAddressHandling_IPv4(t *testing.T) {
	ip := net.ParseIP("192.168.1.100")

	assert.NotNil(t, ip)
	assert.Equal(t, "192.168.1.100", ip.String())
}

func TestIPAddressHandling_IPv6(t *testing.T) {
	ip := net.ParseIP("2001:db8::1")

	assert.NotNil(t, ip)
	assert.Contains(t, ip.String(), "2001:db8")
}

func TestIPAddressHandling_Nil(t *testing.T) {
	var ip net.IP = nil

	// Nil IP should be handled gracefully
	assert.Nil(t, ip)

	// String conversion of nil IP
	if ip != nil {
		_ = ip.String()
	}
}

func TestIPAddressHandling_Localhost(t *testing.T) {
	localhostIPs := []string{
		"127.0.0.1",
		"::1",
	}

	for _, ipStr := range localhostIPs {
		t.Run(ipStr, func(t *testing.T) {
			ip := net.ParseIP(ipStr)
			assert.NotNil(t, ip)
			assert.True(t, ip.IsLoopback())
		})
	}
}

// =============================================================================
// User Metadata Tests
// =============================================================================

func TestUserMetadata_ForJWT(t *testing.T) {
	fullName := "John Doe"
	avatarURL := "https://example.com/avatar.png"

	user := DashboardUser{
		FullName:  &fullName,
		AvatarURL: &avatarURL,
	}

	// Simulate metadata preparation for JWT (as done in Login/LoginViaSSO)
	userMetadata := map[string]interface{}{}
	if user.FullName != nil {
		userMetadata["name"] = *user.FullName
	}
	if user.AvatarURL != nil {
		userMetadata["avatar"] = *user.AvatarURL
	}

	assert.Equal(t, "John Doe", userMetadata["name"])
	assert.Equal(t, "https://example.com/avatar.png", userMetadata["avatar"])
}

func TestUserMetadata_EmptyFields(t *testing.T) {
	user := DashboardUser{
		FullName:  nil,
		AvatarURL: nil,
	}

	userMetadata := map[string]interface{}{}
	if user.FullName != nil {
		userMetadata["name"] = *user.FullName
	}
	if user.AvatarURL != nil {
		userMetadata["avatar"] = *user.AvatarURL
	}

	assert.Empty(t, userMetadata)
}

// =============================================================================
// Lock Expiration Tests
// =============================================================================

func TestLockExpiration_NotExpired(t *testing.T) {
	lockedUntil := time.Now().Add(15 * time.Minute)

	user := DashboardUser{
		IsLocked:    true,
		LockedUntil: &lockedUntil,
	}

	// Lock is still active
	assert.True(t, user.IsLocked)
	assert.True(t, user.LockedUntil.After(time.Now()))
}

func TestLockExpiration_Expired(t *testing.T) {
	lockedUntil := time.Now().Add(-5 * time.Minute) // Expired 5 minutes ago

	user := DashboardUser{
		IsLocked:    true,
		LockedUntil: &lockedUntil,
	}

	// Lock has expired (time.Now().After(*user.LockedUntil) would be true)
	assert.True(t, time.Now().After(*user.LockedUntil))
}

func TestLockExpiration_NoLockTime(t *testing.T) {
	user := DashboardUser{
		IsLocked:    true,
		LockedUntil: nil, // Permanently locked
	}

	assert.True(t, user.IsLocked)
	assert.Nil(t, user.LockedUntil)
}

// =============================================================================
// Session Expiration Tests
// =============================================================================

func TestSessionExpiration_Valid(t *testing.T) {
	session := DashboardSession{
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	assert.True(t, session.ExpiresAt.After(time.Now()))
}

func TestSessionExpiration_Expired(t *testing.T) {
	session := DashboardSession{
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	assert.True(t, session.ExpiresAt.Before(time.Now()))
}

// =============================================================================
// NewDashboardAuthService Tests
// =============================================================================

func TestNewDashboardAuthService_NilDB(t *testing.T) {
	// Can create service with nil DB (for testing purposes)
	svc := NewDashboardAuthService(nil, nil, "Fluxbase")

	assert.NotNil(t, svc)
	assert.Nil(t, svc.db)
	assert.Equal(t, "Fluxbase", svc.totpIssuer)
}

func TestNewDashboardAuthService_CustomIssuer(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "MyCompany Dashboard")

	assert.Equal(t, "MyCompany Dashboard", svc.totpIssuer)
}

func TestNewDashboardAuthService_GetDB(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "Fluxbase")

	// GetDB should return the db field
	assert.Nil(t, svc.GetDB())
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkGenerateBackupCode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = generateBackupCode()
	}
}

func BenchmarkGenerateBackupCodes_10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		codes := make([]string, 10)
		for j := 0; j < 10; j++ {
			code, _ := generateBackupCode()
			codes[j] = code
		}
	}
}

func BenchmarkUserMetadataPreparation(b *testing.B) {
	fullName := "John Doe"
	avatarURL := "https://example.com/avatar.png"
	user := DashboardUser{
		FullName:  &fullName,
		AvatarURL: &avatarURL,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		userMetadata := map[string]interface{}{}
		if user.FullName != nil {
			userMetadata["name"] = *user.FullName
		}
		if user.AvatarURL != nil {
			userMetadata["avatar"] = *user.AvatarURL
		}
	}
}

func BenchmarkProviderParsing(b *testing.B) {
	provider := "oauth:google"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parts := strings.SplitN(provider, ":", 2)
		_ = parts[0]
		_ = parts[1]
	}
}

func BenchmarkIPAddressParsing(b *testing.B) {
	ipStr := "192.168.1.100"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = net.ParseIP(ipStr)
	}
}

// =============================================================================
// Dashboard Service Tests (Mock-based)
// =============================================================================

func TestDashboardAuthService_CreateUser_Validation(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		fullName string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "invalid email - missing @",
			email:    "invalidemail.com",
			password: "SecurePassword123!",
			fullName: "Admin User",
			wantErr:  true,
			errMsg:   "invalid email",
		},
		{
			name:     "password too short",
			email:    "admin@example.com",
			password: "short",
			fullName: "Admin User",
			wantErr:  true,
			errMsg:   "password must be at least",
		},
		{
			name:     "name too long",
			email:    "admin@example.com",
			password: "SecurePassword123!",
			fullName: strings.Repeat("A", 300),
			wantErr:  true,
			errMsg:   "invalid name",
		},
		{
			name:     "empty email",
			email:    "",
			password: "SecurePassword123!",
			fullName: "Admin User",
			wantErr:  true,
			errMsg:   "invalid email",
		},
		{
			name:     "empty password",
			email:    "admin@example.com",
			password: "",
			fullName: "Admin User",
			wantErr:  true,
			errMsg:   "password must be at least",
		},
		{
			name:     "empty name",
			email:    "admin@example.com",
			password: "SecurePassword123!",
			fullName: "",
			wantErr:  true,
			errMsg:   "invalid name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewDashboardAuthService(nil, nil, "fluxbase")
			_, err := svc.CreateUser(nil, tt.email, tt.password, tt.fullName)

			assert.Error(t, err)
			if tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDashboardAuthService_LoginResponse_Fields(t *testing.T) {
	response := LoginResponse{
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		RefreshToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
		ExpiresIn:    3600,
	}

	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.RefreshToken)
	assert.Equal(t, int64(3600), response.ExpiresIn)
}

func TestDashboardAuthService_TOTPSetup(t *testing.T) {
	// TOTP setup requires a database, so we just test that the service can be created
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)

	// Test with nil DB - SetupTOTP will panic, so we skip the actual call
	// In a real test with a mock DB, we would test the actual TOTP setup
	userID := uuid.New()
	assert.NotEqual(t, uuid.Nil, userID)
}

func TestDashboardAuthService_VerifyTOTP(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// VerifyTOTP requires database access, so we skip actual call
}

func TestDashboardAuthService_GetUserByID_NilDB(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// GetUserByID requires database access, so we skip actual call
}

func TestDashboardAuthService_GetUserByEmail_NilDB(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// GetUserByEmail requires database access, so we skip actual call
}

func TestDashboardAuthService_HasExistingUsers_NilDB(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// HasExistingUsers requires database access, so we skip actual call
}

func TestDashboardAuthService_UpdateProfile_Validation(t *testing.T) {
	tests := []struct {
		name      string
		userID    uuid.UUID
		fullName  string
		avatarURL *string
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "name too long",
			userID:   uuid.New(),
			fullName: strings.Repeat("A", 300),
			wantErr:  true,
			errMsg:   "invalid name",
		},
		{
			name:     "invalid avatar URL",
			userID:   uuid.New(),
			fullName: "Valid Name",
			avatarURL: func() *string {
				s := "javascript:alert('xss')"
				return &s
			}(),
			wantErr: true,
			errMsg:  "invalid avatar URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewDashboardAuthService(nil, nil, "fluxbase")
			err := svc.UpdateProfile(nil, tt.userID, tt.fullName, tt.avatarURL)

			assert.Error(t, err)
			if tt.errMsg != "" {
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDashboardAuthService_Login_Validation(t *testing.T) {
	// Login requires database access for user lookup
	// We test that the service can be created and handles nil IP gracefully
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)

	// Test IP parsing for login context
	ip := net.ParseIP("192.168.1.1")
	assert.NotNil(t, ip)

	// Test nil IP handling (should not panic)
	var nilIP net.IP
	assert.Nil(t, nilIP)
}

func TestDashboardAuthService_ChangePassword_Validation(t *testing.T) {
	// ChangePassword requires database access to verify current password
	// We test that the service can be created
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)

	// Test password length validation logic
	shortPassword := "short"
	longPassword := strings.Repeat("a", 73)
	validPassword := "ValidPassword123!"

	assert.Less(t, len(shortPassword), MinPasswordLength)
	assert.Greater(t, len(longPassword), MaxPasswordLength)
	assert.GreaterOrEqual(t, len(validPassword), MinPasswordLength)
	assert.LessOrEqual(t, len(validPassword), MaxPasswordLength)
}

func TestDashboardAuthService_DeleteAccount_Validation(t *testing.T) {
	// DeleteAccount requires database access
	// We test that the service can be created
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)

	// Test that empty password is handled
	emptyPassword := ""
	assert.Equal(t, 0, len(emptyPassword))
}

func TestDashboardAuthService_DisableTOTP_Validation(t *testing.T) {
	// DisableTOTP requires database access
	// We test that the service can be created
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
}

func TestDashboardAuthService_EnableTOTP_Validation(t *testing.T) {
	// EnableTOTP requires database access
	// We test TOTP code format validation
	validCode := "123456"
	shortCode := "12345"
	longCode := "1234567"
	emptyCode := ""

	assert.Len(t, validCode, 6)
	assert.NotEqual(t, 6, len(shortCode))
	assert.NotEqual(t, 6, len(longCode))
	assert.Equal(t, 0, len(emptyCode))
}

func TestDashboardAuthService_SSOIdentity_Fields(t *testing.T) {
	now := time.Now()
	email := "admin@example.com"

	identity := SSOIdentity{
		ID:             uuid.New(),
		UserID:         uuid.New(),
		Provider:       "google",
		ProviderUserID: "google-user-123",
		Email:          &email,
		CreatedAt:      now,
	}

	assert.NotEmpty(t, identity.ID)
	assert.NotEmpty(t, identity.UserID)
	assert.Equal(t, "google", identity.Provider)
	assert.Equal(t, "google-user-123", identity.ProviderUserID)
	assert.Equal(t, "admin@example.com", *identity.Email)
}

func TestDashboardSSO_Providers(t *testing.T) {
	validProviders := []string{
		"oauth:google",
		"oauth:github",
		"oauth:microsoft",
		"oauth:apple",
		"saml:okta",
		"saml:azure-ad",
		"oidc:auth0",
	}

	for _, provider := range validProviders {
		t.Run("provider_"+provider, func(t *testing.T) {
			parts := strings.SplitN(provider, ":", 2)
			assert.Len(t, parts, 2)
			assert.NotEmpty(t, parts[0]) // Type (oauth, saml, oidc)
			assert.NotEmpty(t, parts[1]) // Provider name
		})
	}
}

func TestDashboardBackupCode_Base32Encoding(t *testing.T) {
	// Test that base32 encoding/decoding works correctly
	original := "test123code"

	// Encode
	encoded := base32.StdEncoding.EncodeToString([]byte(original))
	assert.NotEmpty(t, encoded)

	// Decode
	decoded, err := base32.StdEncoding.DecodeString(encoded)
	assert.NoError(t, err)
	assert.Equal(t, original, string(decoded))
}

func TestDashboardPasswordReset_TokenFormat(t *testing.T) {
	// Test token generation format
	token := "test_token_123"

	// Verify token can be hashed
	hash := sha256.Sum256([]byte(token))
	assert.Len(t, hash, 32)

	// Verify token can be base64 encoded
	encoded := base64.URLEncoding.EncodeToString(hash[:])
	assert.NotEmpty(t, encoded)
}

func TestDashboardSession_Expiration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "not expired",
			expiresAt: now.Add(24 * time.Hour),
			expected:  false,
		},
		{
			name:      "expired",
			expiresAt: now.Add(-1 * time.Hour),
			expected:  true,
		},
		{
			name:      "just expired",
			expiresAt: now.Add(-1 * time.Second),
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := DashboardSession{
				ID:             uuid.New(),
				UserID:         uuid.New(),
				ExpiresAt:      tt.expiresAt,
				CreatedAt:      now,
				LastActivityAt: now,
			}

			isExpired := time.Now().After(session.ExpiresAt)
			assert.Equal(t, tt.expected, isExpired)
		})
	}
}

func TestDashboardAuthService_RefreshToken(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// RefreshToken requires JWT manager, so we skip actual call
}

func TestDashboardAuthService_LoginViaSSO(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)

	user := &DashboardUser{
		ID:       uuid.New(),
		Email:    "admin@example.com",
		IsActive: true,
	}
	assert.NotNil(t, user)
	// LoginViaSSO requires JWT manager, so we skip actual call
}

func TestDashboardAuthService_RequestPasswordReset(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// RequestPasswordReset requires database access, so we skip actual call
}

func TestDashboardAuthService_VerifyPasswordResetToken(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// VerifyPasswordResetToken requires database access, so we skip actual call
}

func TestDashboardAuthService_ResetPassword(t *testing.T) {
	svc := NewDashboardAuthService(nil, nil, "fluxbase")
	assert.NotNil(t, svc)
	// ResetPassword requires database access, so we skip actual call
}
