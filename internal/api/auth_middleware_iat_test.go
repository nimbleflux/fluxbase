package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/auth"
)

// =============================================================================
// Regression tests: admin auth middleware vs. tokens without an iat claim.
//
// The service_role / anon keys emitted by deploy/generate-keys.sh sign a
// payload of {"role":"service_role","token_type":"access","iss":"fluxbase"}
// only — no iat, exp or jti. UnifiedAuthMiddleware dereferenced
// claims.IssuedAt (a *jwt.NumericDate) unconditionally when checking token
// revocation, so EVERY request bearing such a key panicked with a nil pointer
// dereference (recovered as HTTP 500) — the fresh-boot CI job's admin probe
// failed with exactly this signature. These tests pin the guard.
// =============================================================================

// mintTokenNoIat builds an HMAC-SHA256 JWT whose payload is exactly the JSON
// given, mirroring deploy/generate-keys.sh's generate_jwt helper (no iat/exp).
func mintTokenNoIat(t *testing.T, secret, payload string) string {
	t.Helper()
	b64 := func(b []byte) string {
		return base64.RawURLEncoding.EncodeToString(b)
	}
	header := b64([]byte(`{"alg":"HS256","typ":"JWT"}`))
	body := b64([]byte(payload))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(header + "." + body))
	return header + "." + body + "." + b64(mac.Sum(nil))
}

// TestUnifiedAuthMiddleware_ServiceKeyWithoutIat drives the exact CI probe: a
// service_role access token with no iat claim against the unified admin auth
// middleware. Before the fix this panicked (nil IssuedAt deref); it must pass
// the middleware through to the next handler instead.
func TestUnifiedAuthMiddleware_ServiceKeyWithoutIat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	svc, _ := newSecurityFixTestService(t)

	const payload = `{"role":"service_role","token_type":"access","iss":"fluxbase"}`
	token := mintTokenNoIat(t, "test-secret-key-for-security-fix-tests", payload)

	app := fiber.New()
	reached := false
	app.Use(UnifiedAuthMiddleware(svc, svc.JWTManager(), nil))
	app.Get("/api/v1/admin/extensions", func(c fiber.Ctx) error {
		reached = true
		role, _ := GetUserRole(c)
		assert.Equal(t, "service_role", role)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/api/v1/admin/extensions", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	require.NoError(t, err, "request must not error (panic would surface here)")
	assert.Equal(t, fiber.StatusOK, resp.StatusCode, "service_role key must authenticate")
	assert.True(t, reached, "handler must be reached")
}

// TestAuthMiddleware_AnonKeyWithoutIat covers the anon key: same minimal
// payload shape as the service key, must not panic either.
func TestAuthMiddleware_AnonKeyWithoutIat(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping database test in short mode")
	}
	svc, _ := newSecurityFixTestService(t)

	const payload = `{"role":"anon","token_type":"access","iss":"fluxbase"}`
	token := mintTokenNoIat(t, "test-secret-key-for-security-fix-tests", payload)

	app := fiber.New()
	app.Use(AuthMiddleware(svc))
	app.Get("/ping", func(c fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) })

	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

// TestTokenIssuedAt pins the guard helper itself.
func TestTokenIssuedAt(t *testing.T) {
	assert.True(t, tokenIssuedAt(nil).IsZero())
	assert.True(t, tokenIssuedAt(&auth.TokenClaims{}).IsZero())

	now := time.Now()
	claims := &auth.TokenClaims{}
	claims.IssuedAt = jwt.NewNumericDate(now)
	assert.Equal(t, now.Unix(), tokenIssuedAt(claims).Unix())
}
