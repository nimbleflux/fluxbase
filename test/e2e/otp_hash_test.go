package e2e

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

func setupOTPHashTest(t *testing.T) *test.TestContext {
	tc := test.NewTestContext(t)
	tc.EnsureAuthSchema()
	tc.ExecuteSQL("DELETE FROM auth.otp_codes WHERE email LIKE 'e2e-otp-%'")
	tc.ExecuteSQL("DELETE FROM auth.users WHERE email LIKE 'e2e-otp-%'")
	tc.Config.Auth.SignupEnabled = true
	return tc
}

func TestOTPHashStorage(t *testing.T) {
	tc := setupOTPHashTest(t)
	defer tc.Close()

	email := test.E2ETestEmail()
	password := "testpassword123"

	_, _ = tc.CreateTestUser(email, password)

	otpEmail := fmt.Sprintf("e2e-otp-hash-%d@test.com", time.Now().UnixMilli())
	tc.NewRequest("POST", "/api/v1/auth/otp/signin").
		WithBody(map[string]interface{}{
			"email": otpEmail,
		}).
		Send()

	rows := tc.QuerySQL("SELECT code_hash, code FROM auth.otp_codes WHERE email = $1 ORDER BY created_at DESC LIMIT 1", otpEmail)
	require.Len(t, rows, 1, "expected exactly one OTP code row")

	codeHash, ok := rows[0]["code_hash"].(string)
	assert.True(t, ok, "code_hash should be a string")
	assert.NotEmpty(t, codeHash, "code_hash should not be empty")

	decoded, err := base64.URLEncoding.DecodeString(codeHash)
	assert.NoError(t, err, "code_hash should be valid base64url")
	assert.Len(t, decoded, 32, "code_hash should decode to 32 bytes (SHA-256)")

	tc.ExecuteSQL("DELETE FROM auth.otp_codes WHERE email = $1", otpEmail)
}

func TestOTPLazyMigration(t *testing.T) {
	tc := setupOTPHashTest(t)
	defer tc.Close()

	legacyEmail := fmt.Sprintf("e2e-otp-legacy-%d@test.com", time.Now().UnixMilli())
	legacyCode := "999888"

	tc.ExecuteSQLAsSuperuser(fmt.Sprintf(
		`INSERT INTO auth.otp_codes (id, email, code, code_hash, type, purpose, expires_at, used, attempts, max_attempts, created_at)
		 VALUES (gen_random_uuid(), '%s', '%s', NULL, 'email', 'signin', NOW() + INTERVAL '15 minutes', false, 0, 10, NOW())`,
		legacyEmail, legacyCode,
	))

	rowsBefore := tc.QuerySQL("SELECT code_hash FROM auth.otp_codes WHERE email = $1", legacyEmail)
	require.Len(t, rowsBefore, 1)
	assert.Nil(t, rowsBefore[0]["code_hash"], "code_hash should be NULL before migration")

	tc.NewRequest("POST", "/api/v1/auth/otp/verify").
		WithBody(map[string]interface{}{
			"email": legacyEmail,
			"token": legacyCode,
			"type":  "email",
		}).
		Send()

	rowsAfter := tc.QuerySQL("SELECT code_hash FROM auth.otp_codes WHERE email = $1 AND code = $2", legacyEmail, legacyCode)
	require.Len(t, rowsAfter, 1)
	codeHash, ok := rowsAfter[0]["code_hash"].(string)
	assert.True(t, ok, "code_hash should be a string after lazy migration")
	assert.NotEmpty(t, codeHash, "code_hash should be populated after lazy migration")

	tc.ExecuteSQL("DELETE FROM auth.otp_codes WHERE email = $1", legacyEmail)
}
