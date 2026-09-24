package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
)

// newSecurityFixTestService connects to the test database and builds an auth
// service for handler-level security tests. Skips when the DB is unavailable
// (e.g. unit-only runs without a database).
func newSecurityFixTestService(t *testing.T) (*auth.Service, *database.Connection) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}

	dbHost := os.Getenv("FLUXBASE_DATABASE_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbUser := os.Getenv("FLUXBASE_DATABASE_USER")
	if dbUser == "" {
		dbUser = "fluxbase_app"
	}
	dbPassword := os.Getenv("FLUXBASE_DATABASE_PASSWORD")
	if dbPassword == "" {
		dbPassword = "fluxbase_app_password"
	}
	dbDatabase := os.Getenv("FLUXBASE_TEST_DATABASE")
	if dbDatabase == "" {
		dbDatabase = "fluxbase_test"
	}

	db, err := database.NewConnection(config.DatabaseConfig{
		Host:            dbHost,
		Port:            5432,
		User:            dbUser,
		Password:        dbPassword,
		Database:        dbDatabase,
		SSLMode:         "disable",
		MaxConnections:  5,
		MinConnections:  1,
		MaxConnLifetime: 5 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
		HealthCheck:     30 * time.Second,
	})
	if err != nil {
		t.Skipf("test database unavailable: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Health(ctx); err != nil {
		t.Skipf("test database unhealthy: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	authConfig := &config.AuthConfig{
		JWTSecret:      "test-secret-key-for-security-fix-tests",
		JWTExpiry:      15 * time.Minute,
		RefreshExpiry:  7 * 24 * time.Hour,
		PasswordMinLen: 8,
		BcryptCost:     4,
		SignupEnabled:  true,
	}
	return auth.NewService(db, authConfig, &auth.NoOpOTPSender{}, "http://localhost:3000", nil), db
}

// createTestAuthUser creates a user for the test run and registers cleanup.
func createTestAuthUser(t *testing.T, service *auth.Service, db *database.Connection, email, password string) *auth.User {
	t.Helper()
	ctx := context.Background()
	user, err := service.CreateUser(ctx, email, password)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(), `DELETE FROM auth.users WHERE id = $1`, user.ID)
	})
	return user
}

// randHex returns n random bytes as a lowercase hex string (for unique emails).
func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// =============================================================================
// F1: PATCH /auth/user must not accept privileged fields
// =============================================================================

func TestUpdateUserHandler_IgnoresPrivilegedFields(t *testing.T) {
	service, db := newSecurityFixTestService(t)

	email := fmt.Sprintf("selfapi-%s@example.com", strings.ToLower(randHex(6)))
	user := createTestAuthUser(t, service, db, email, "Password123")

	handler := NewAuthHandler(db, service, nil, "http://localhost:3000", "")
	app := newTestApp(t)
	app.Use(func(c fiber.Ctx) error {
		c.Locals("user_id", user.ID)
		return c.Next()
	})
	app.Patch("/auth/user", handler.UpdateUser)

	// Attempt a self-service privilege escalation alongside a benign update.
	body := `{"role":"admin","email_verified":true,"app_metadata":{"escalated":true},"user_metadata":{"theme":"dark"}}`
	req := httptest.NewRequest(http.MethodPatch, "/auth/user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated auth.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))

	assert.Equal(t, user.Role, updated.Role, "role must not change via self-service")
	assert.False(t, updated.EmailVerified, "email_verified must not change via self-service")
	assert.Equal(t, map[string]interface{}{"theme": "dark"}, updated.UserMetadata)

	// Double-check from the database
	reloaded, err := service.UserRepository().GetByID(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, user.Role, reloaded.Role)
	assert.False(t, reloaded.EmailVerified)
	reloadedAppMeta, _ := reloaded.AppMetadata.(map[string]interface{})
	assert.Empty(t, reloadedAppMeta, "app_metadata must remain empty")
}

// =============================================================================
// F2: refresh tokens must not authenticate via bearer validation
// =============================================================================

func TestAuthMiddleware_RejectsRefreshToken(t *testing.T) {
	service, db := newSecurityFixTestService(t)

	email := fmt.Sprintf("refrtok-%s@example.com", strings.ToLower(randHex(6)))
	user := createTestAuthUser(t, service, db, email, "Password123")

	_, refreshToken, _, err := service.JWTManager().GenerateTokenPair(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)

	app := newTestApp(t)
	app.Get("/protected", AuthMiddleware(service), func(c fiber.Ctx) error {
		return c.SendString("ok:" + middlewareUserID(c))
	})

	// A refresh token presented as a bearer credential must be rejected.
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// An access token still authenticates.
	accessToken, _, _, err := service.JWTManager().GenerateTokenPair(user.ID, user.Email, user.Role, nil, nil)
	require.NoError(t, err)
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func middlewareUserID(c fiber.Ctx) string {
	id, _ := c.Locals("user_id").(string)
	return id
}

// =============================================================================
// F3: auth cookies derive the Secure flag from the public base URL
// =============================================================================

func TestAuthCookieSecureFlag_DerivedFromBaseURL(t *testing.T) {
	service, db := newSecurityFixTestService(t)

	accessToken, refreshToken, _, err := service.JWTManager().GenerateTokenPair("u-1", "e@example.com", "authenticated", nil, nil)
	require.NoError(t, err)

	assertCookies := func(t *testing.T, handler *AuthHandler, wantSecure bool) {
		t.Helper()
		app := newTestApp(t)
		app.Post("/set", handler.setAuthCookiesHandler(accessToken, refreshToken))
		req := httptest.NewRequest(http.MethodPost, "/set", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		cookies := resp.Cookies()
		require.Len(t, cookies, 2)
		for _, ck := range cookies {
			assert.Equal(t, wantSecure, ck.Secure, "cookie %s Secure flag mismatch", ck.Name)
		}
	}

	// HTTPS public base URL -> Secure cookies
	assertCookies(t, NewAuthHandler(db, service, nil, "https://app.example.com", ""), true)
	// Plain HTTP (local dev) -> not Secure
	assertCookies(t, NewAuthHandler(db, service, nil, "http://localhost:8080", ""), false)
}

// setAuthCookiesHandler is a small test helper that exercises setAuthCookies.
func (h *AuthHandler) setAuthCookiesHandler(accessToken, refreshToken string) fiber.Handler {
	return func(c fiber.Ctx) error {
		h.setAuthCookies(c, accessToken, refreshToken, 900)
		return c.SendStatus(http.StatusOK)
	}
}

// =============================================================================
// F4: OAuth user-info email_verified interpretation
// =============================================================================

func TestProviderEmailVerified(t *testing.T) {
	assert.True(t, providerEmailVerified(map[string]interface{}{"email_verified": true}))
	assert.False(t, providerEmailVerified(map[string]interface{}{"email_verified": false}))
	// Unknown/absent claims are treated as unverified (conservative).
	assert.False(t, providerEmailVerified(map[string]interface{}{}))
	assert.False(t, providerEmailVerified(nil))
	assert.False(t, providerEmailVerified(map[string]interface{}{"email_verified": "true"}))
}

// =============================================================================
// F5: POST /auth/2fa/verify requires the signed mfa_token
// =============================================================================

func enableTOTPForTest(t *testing.T, db *database.Connection, userID string) string {
	t.Helper()
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "Fluxbase", AccountName: userID})
	require.NoError(t, err)
	_, err = db.Exec(context.Background(),
		`UPDATE auth.users SET totp_secret = $1, totp_enabled = TRUE, totp_last_used_step = 0, updated_at = NOW() WHERE id = $2`,
		key.Secret(), userID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(context.Background(),
			`UPDATE auth.users SET totp_secret = NULL, totp_enabled = FALSE, totp_last_used_step = 0, backup_codes = NULL WHERE id = $1`, userID)
	})
	return key.Secret()
}

func TestVerifyTOTPHandler_RequiresMFAChallengeToken(t *testing.T) {
	service, db := newSecurityFixTestService(t)

	email := fmt.Sprintf("mfaapi-%s@example.com", strings.ToLower(randHex(6)))
	user := createTestAuthUser(t, service, db, email, "Password123")
	secret := enableTOTPForTest(t, db, user.ID)

	handler := NewAuthHandler(db, service, nil, "http://localhost:3000", "")
	app := newTestApp(t)
	app.Post("/auth/2fa/verify", handler.VerifyTOTP)

	call := func(t *testing.T, body string) (int, *auth.SignInResponse) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/auth/2fa/verify", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		var payload auth.SignInResponse
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		return resp.StatusCode, &payload
	}

	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)

	// Without the signed challenge token, the request is rejected even with a
	// valid code and a matching user_id from the body.
	status, _ := call(t, fmt.Sprintf(`{"user_id": %q, "code": %q}`, user.ID, code))
	assert.Equal(t, http.StatusUnauthorized, status)

	// With a valid challenge token but a wrong code, verification fails.
	mfaToken, _, err := service.JWTManager().GenerateMFAPendingToken(user.ID, auth.MFAPendingPurpose2FA)
	require.NoError(t, err)
	status, _ = call(t, fmt.Sprintf(`{"mfa_token": %q, "user_id": %q, "code": "000000"}`, mfaToken, user.ID))
	assert.Equal(t, http.StatusBadRequest, status)

	// With both the token and a fresh code, sign-in completes.
	code2, err := totp.GenerateCode(secret, time.Now().Add(31*time.Second))
	require.NoError(t, err)
	status, payload := call(t, fmt.Sprintf(`{"mfa_token": %q, "user_id": %q, "code": %q}`, mfaToken, user.ID, code2))
	assert.Equal(t, http.StatusOK, status)
	assert.NotEmpty(t, payload.AccessToken)
	assert.Equal(t, user.ID, payload.User.ID)

	// A used time step cannot be replayed.
	status, _ = call(t, fmt.Sprintf(`{"mfa_token": %q, "user_id": %q, "code": %q}`, mfaToken, user.ID, code2))
	assert.Equal(t, http.StatusBadRequest, status)
}

func TestSignInHandler_ReturnsMFATokenWhen2FARequired(t *testing.T) {
	service, db := newSecurityFixTestService(t)

	email := fmt.Sprintf("mfasi-%s@example.com", strings.ToLower(randHex(6)))
	user := createTestAuthUser(t, service, db, email, "Password123")
	secret := enableTOTPForTest(t, db, user.ID)

	handler := NewAuthHandler(db, service, nil, "http://localhost:3000", "")
	app := newTestApp(t)
	app.Post("/auth/signin", handler.SignIn)

	req := httptest.NewRequest(http.MethodPost, "/auth/signin",
		strings.NewReader(fmt.Sprintf(`{"email": %q, "password": "Password123"}`, email)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var payload struct {
		Requires2FA bool   `json:"requires_2fa"`
		UserID      string `json:"user_id"`
		MFAToken    string `json:"mfa_token"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&payload))
	assert.True(t, payload.Requires2FA)
	assert.Equal(t, user.ID, payload.UserID)
	assert.NotEmpty(t, payload.MFAToken)

	// The returned token must validate as a 2FA challenge for this user.
	claims, err := service.JWTManager().ValidateMFAPendingToken(payload.MFAToken, auth.MFAPendingPurpose2FA)
	require.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)

	// And it must be usable end-to-end at /auth/2fa/verify.
	code, err := totp.GenerateCode(secret, time.Now())
	require.NoError(t, err)
	handler2 := NewAuthHandler(db, service, nil, "http://localhost:3000", "")
	app2 := newTestApp(t)
	app2.Post("/auth/2fa/verify", handler2.VerifyTOTP)
	req2 := httptest.NewRequest(http.MethodPost, "/auth/2fa/verify",
		strings.NewReader(fmt.Sprintf(`{"mfa_token": %q, "code": %q}`, payload.MFAToken, code)))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app2.Test(req2)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp2.StatusCode)
}
