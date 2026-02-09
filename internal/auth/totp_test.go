package auth

import (
	"encoding/base32"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateTOTPSecret_Success(t *testing.T) {
	issuer := "Fluxbase"
	accountName := "user@example.com"

	secret, qrCodeDataURI, otpauthURI, err := GenerateTOTPSecret(issuer, accountName)

	require.NoError(t, err)
	assert.NotEmpty(t, secret)
	assert.NotEmpty(t, qrCodeDataURI)
	assert.NotEmpty(t, otpauthURI)

	// Verify secret is valid base32
	_, err = base32.StdEncoding.DecodeString(secret)
	assert.NoError(t, err, "secret should be valid base32")

	// Verify QR code data URI format
	assert.True(t, strings.HasPrefix(qrCodeDataURI, "data:image/png;base64,"))

	// Verify otpauth URI contains issuer and account name
	assert.Contains(t, otpauthURI, issuer)
	assert.Contains(t, otpauthURI, accountName)
	assert.True(t, strings.HasPrefix(otpauthURI, "otpauth://totp/"))
}

func TestGenerateTOTPSecret_UniquenessPerCall(t *testing.T) {
	issuer := "Fluxbase"
	accountName := "user@example.com"

	secret1, _, _, err1 := GenerateTOTPSecret(issuer, accountName)
	secret2, _, _, err2 := GenerateTOTPSecret(issuer, accountName)

	require.NoError(t, err1)
	require.NoError(t, err2)

	// Each call should generate a unique secret
	assert.NotEqual(t, secret1, secret2)
}

func TestGenerateTOTPSecret_DifferentAccounts(t *testing.T) {
	tests := []struct {
		name        string
		issuer      string
		accountName string
	}{
		{
			name:        "standard email",
			issuer:      "Fluxbase",
			accountName: "user@example.com",
		},
		{
			name:        "email with plus sign",
			issuer:      "Fluxbase",
			accountName: "user+test@example.com",
		},
		{
			name:        "email with subdomain",
			issuer:      "Fluxbase",
			accountName: "admin@staging.example.com",
		},
		{
			name:        "different issuer",
			issuer:      "MyApp",
			accountName: "user@example.com",
		},
		{
			name:        "special characters",
			issuer:      "Fluxbase",
			accountName: "user-test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, qrCodeDataURI, otpauthURI, err := GenerateTOTPSecret(tt.issuer, tt.accountName)

			require.NoError(t, err)
			assert.NotEmpty(t, secret)
			assert.NotEmpty(t, qrCodeDataURI)
			assert.NotEmpty(t, otpauthURI)
			assert.Contains(t, otpauthURI, tt.issuer)
			assert.Contains(t, otpauthURI, tt.accountName)
		})
	}
}

func TestVerifyTOTPCode_ValidCode(t *testing.T) {
	// Generate a secret
	issuer := "Fluxbase"
	accountName := "user@example.com"
	secret, _, _, err := GenerateTOTPSecret(issuer, accountName)
	require.NoError(t, err)

	// Generate a valid TOTP code for current time
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	// Verify the code
	valid, err := VerifyTOTPCode(code, secret)

	require.NoError(t, err)
	assert.True(t, valid, "valid TOTP code should be verified successfully")
}

func TestVerifyTOTPCode_InvalidCode(t *testing.T) {
	// Generate a secret
	secret, _, _, err := GenerateTOTPSecret("Fluxbase", "user@example.com")
	require.NoError(t, err)

	tests := []struct {
		name string
		code string
	}{
		{
			name: "wrong code",
			code: "000000",
		},
		{
			name: "invalid format",
			code: "abcdef",
		},
		{
			name: "too short",
			code: "123",
		},
		{
			name: "too long",
			code: "12345678",
		},
		{
			name: "empty code",
			code: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := VerifyTOTPCode(tt.code, secret)

			require.NoError(t, err)
			assert.False(t, valid, "invalid TOTP code should not be verified")
		})
	}
}

func TestVerifyTOTPCode_ExpiredCode(t *testing.T) {
	// Generate a secret
	secret, _, _, err := GenerateTOTPSecret("Fluxbase", "user@example.com")
	require.NoError(t, err)

	// Generate a code for a past time (more than 30 seconds ago - outside TOTP window)
	pastTime := time.Now().Add(-2 * time.Minute)
	oldCode, err := totp.GenerateCode(secret, pastTime)
	require.NoError(t, err)

	// Verify the old code should fail
	valid, err := VerifyTOTPCode(oldCode, secret)
	require.NoError(t, err)
	assert.False(t, valid, "Old TOTP code from outside the time window should not verify")
}
