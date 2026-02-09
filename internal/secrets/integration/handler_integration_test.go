//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/auth"
	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/fluxbase-eu/fluxbase/internal/secrets"
	"github.com/fluxbase-eu/fluxbase/internal/testutil"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecretsHandler_CreateSecret_Integration tests POST /secrets endpoint
func TestSecretsHandler_CreateSecret_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)

	// Create test user and get auth token
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	t.Run("create global secret", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "TEST_API_KEY",
			"value": "sk-1234567890abcdef",
			"scope": "global",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result secrets.Secret
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "TEST_API_KEY", result.Name)
		assert.Equal(t, "global", result.Scope)
		assert.Equal(t, 1, result.Version)
		assert.NotEqual(t, uuid.UUID{}, result.ID)
		assert.Empty(t, result.EncryptedValue, "Should not expose encrypted value")
	})

	t.Run("create namespace secret", func(t *testing.T) {
		namespace := "production"
		reqBody := map[string]interface{}{
			"name":      "DB_PASSWORD",
			"value":     "supersecret",
			"scope":     "namespace",
			"namespace": namespace,
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result secrets.Secret
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "namespace", result.Scope)
		assert.Equal(t, &namespace, result.Namespace)
	})

	t.Run("create secret with all fields", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
		description := "Production API key for external service"

		reqBody := map[string]interface{}{
			"name":        "FULL_SECRET",
			"value":       "full-secret-value",
			"scope":       "global",
			"description": description,
			"expires_at":  expiresAt,
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result secrets.Secret
		err := json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, "FULL_SECRET", result.Name)
		assert.Equal(t, &description, result.Description)
		assert.NotNil(t, result.ExpiresAt)
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"value": "some-value",
			"scope": "global",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "Name")
	})

	t.Run("missing value returns 400", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "TEST",
			"scope": "global",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "Value")
	})

	t.Run("invalid scope returns 400", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "TEST",
			"value": "value",
			"scope": "invalid",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("namespace scope without namespace returns 400", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "TEST",
			"value": "value",
			"scope": "namespace",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "Namespace")
	})

	t.Run("duplicate name returns 409", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "DUPLICATE_NAME",
			"value": "first-value",
			"scope": "global",
		}
		resp1 := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)
		assert.Equal(t, http.StatusCreated, resp1.StatusCode)

		// Try to create again with same name
		reqBody["value"] = "second-value"
		resp2 := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, token)
		assert.Equal(t, http.StatusConflict, resp2.StatusCode)
	})

	t.Run("no auth token returns 401", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":  "NO_AUTH",
			"value": "value",
			"scope": "global",
		}
		resp := makeRequest(t, app, "POST", "/api/v1/secrets", reqBody, "")
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestSecretsHandler_ListSecrets_Integration tests GET /secrets endpoint
func TestSecretsHandler_ListSecrets_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	// Setup: create test secrets via storage
	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	globalSecret := &secrets.Secret{Name: "GLOBAL_LIST_TEST", Scope: "global"}
	require.NoError(t, storage.CreateSecret(context.Background(), globalSecret, "value1", nil))

	ns := "production"
	nsSecret := &secrets.Secret{Name: "NS_LIST_TEST", Scope: "namespace", Namespace: &ns}
	require.NoError(t, storage.CreateSecret(context.Background(), nsSecret, "value2", nil))

	t.Run("list all secrets", func(t *testing.T) {
		resp := makeRequest(t, app, "GET", "/api/v1/secrets", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretSummary
		err := json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2, "Should have at least 2 secrets")
	})

	t.Run("filter by scope", func(t *testing.T) {
		resp := makeRequest(t, app, "GET", "/api/v1/secrets?scope=global", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretSummary
		err := json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)

		for _, s := range results {
			assert.Equal(t, "global", s.Scope)
		}
	})

	t.Run("filter by namespace", func(t *testing.T) {
		resp := makeRequest(t, app, "GET", "/api/v1/secrets?namespace=production", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretSummary
		err := json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(results), 1)
	})

	t.Run("empty list returns empty array", func(t *testing.T) {
		// Clean up all secrets
		tc.ExecuteSQL(`DELETE FROM functions.secrets`)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretSummary
		err := json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// TestSecretsHandler_GetSecret_Integration tests GET /secrets/:id
func TestSecretsHandler_GetSecret_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("get existing secret", func(t *testing.T) {
		secret := &secrets.Secret{Name: "GET_SINGLE", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "secret-value", nil)
		require.NoError(t, err)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets/"+secret.ID.String(), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result secrets.Secret
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, secret.ID, result.ID)
		assert.Equal(t, "GET_SINGLE", result.Name)
		assert.Empty(t, result.EncryptedValue)
	})

	t.Run("get non-existent secret returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		resp := makeRequest(t, app, "GET", "/api/v1/secrets/"+fakeID.String(), nil, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		resp := makeRequest(t, app, "GET", "/api/v1/secrets/not-a-uuid", nil, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "Invalid")
	})
}

// TestSecretsHandler_UpdateSecret_Integration tests PUT /secrets/:id
func TestSecretsHandler_UpdateSecret_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("update secret value", func(t *testing.T) {
		secret := &secrets.Secret{Name: "UPDATE_VALUE_TEST", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "original", nil)
		require.NoError(t, err)
		assert.Equal(t, 1, secret.Version)

		reqBody := map[string]interface{}{
			"value": strPtr("updated"),
		}
		resp := makeRequest(t, app, "PUT", "/api/v1/secrets/"+secret.ID.String(), reqBody, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result secrets.Secret
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Version, "Version should increment")
	})

	t.Run("update description", func(t *testing.T) {
		secret := &secrets.Secret{Name: "UPDATE_DESC_TEST", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		newDesc := "Updated description"
		reqBody := map[string]interface{}{
			"description": &newDesc,
		}
		resp := makeRequest(t, app, "PUT", "/api/v1/secrets/"+secret.ID.String(), reqBody, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result secrets.Secret
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, &newDesc, result.Description)
	})

	t.Run("update multiple fields", func(t *testing.T) {
		secret := &secrets.Secret{Name: "UPDATE_MULTI_TEST", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		newValue := "new-value"
		newDesc := "New description"
		expiresAt := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)

		reqBody := map[string]interface{}{
			"value":       &newValue,
			"description": &newDesc,
			"expires_at":  &expiresAt,
		}
		resp := makeRequest(t, app, "PUT", "/api/v1/secrets/"+secret.ID.String(), reqBody, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result secrets.Secret
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, 2, result.Version)
		assert.Equal(t, &newDesc, result.Description)
		assert.NotNil(t, result.ExpiresAt)
	})

	t.Run("update with no fields returns 400", func(t *testing.T) {
		secret := &secrets.Secret{Name: "UPDATE_NO_FIELDS", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		reqBody := map[string]interface{}{}
		resp := makeRequest(t, app, "PUT", "/api/v1/secrets/"+secret.ID.String(), reqBody, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var errResp map[string]string
		json.NewDecoder(resp.Body).Decode(&errResp)
		assert.Contains(t, errResp["error"], "at least one field")
	})

	t.Run("update non-existent secret returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		newValue := "value"
		reqBody := map[string]interface{}{"value": &newValue}
		resp := makeRequest(t, app, "PUT", "/api/v1/secrets/"+fakeID.String(), reqBody, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestSecretsHandler_DeleteSecret_Integration tests DELETE /secrets/:id
func TestSecretsHandler_DeleteSecret_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("delete existing secret", func(t *testing.T) {
		secret := &secrets.Secret{Name: "DELETE_ME_HTTP", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		resp := makeRequest(t, app, "DELETE", "/api/v1/secrets/"+secret.ID.String(), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]string
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Contains(t, result["message"], "deleted")
	})

	t.Run("verify secret is deleted", func(t *testing.T) {
		secret := &secrets.Secret{Name: "DELETE_VERIFY", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		// Delete
		resp := makeRequest(t, app, "DELETE", "/api/v1/secrets/"+secret.ID.String(), nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify it's gone
		resp = makeRequest(t, app, "GET", "/api/v1/secrets/"+secret.ID.String(), nil, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("delete non-existent secret returns 404", func(t *testing.T) {
		fakeID := uuid.New()
		resp := makeRequest(t, app, "DELETE", "/api/v1/secrets/"+fakeID.String(), nil, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestSecretsHandler_GetVersions_Integration tests GET /secrets/:id/versions
func TestSecretsHandler_GetVersions_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("get version history", func(t *testing.T) {
		secret := &secrets.Secret{Name: "VERSION_HISTORY_HTTP", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "v1", nil)
		require.NoError(t, err)

		// Create more versions
		v2 := "v2"
		storage.UpdateSecret(context.Background(), secret.ID, &v2, nil, nil, nil)

		v3 := "v3"
		storage.UpdateSecret(context.Background(), secret.ID, &v3, nil, nil, nil)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets/"+secret.ID.String()+"/versions", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretVersion
		err = json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("get versions for non-existent secret", func(t *testing.T) {
		fakeID := uuid.New()
		resp := makeRequest(t, app, "GET", "/api/v1/secrets/"+fakeID.String()+"/versions", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretVersion
		err := json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

// TestSecretsHandler_Rollback_Integration tests POST /secrets/:id/rollback/:version
func TestSecretsHandler_Rollback_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("rollback to previous version", func(t *testing.T) {
		secret := &secrets.Secret{Name: "ROLLBACK_HTTP", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "original", nil)
		require.NoError(t, err)

		// Update to v2
		v2 := "updated"
		storage.UpdateSecret(context.Background(), secret.ID, &v2, nil, nil, nil)

		// Rollback to v1
		resp := makeRequest(t, app, "POST", "/api/v1/secrets/"+secret.ID.String()+"/rollback/1", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result secrets.Secret
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Equal(t, 3, result.Version, "Version should be 3 after rollback")
	})

	t.Run("rollback to non-existent version", func(t *testing.T) {
		secret := &secrets.Secret{Name: "BAD_ROLLBACK_HTTP", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		resp := makeRequest(t, app, "POST", "/api/v1/secrets/"+secret.ID.String()+"/rollback/99", nil, token)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("invalid version number returns 400", func(t *testing.T) {
		secret := &secrets.Secret{Name: "INVALID_VERSION_HTTP", Scope: "global"}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		resp := makeRequest(t, app, "POST", "/api/v1/secrets/"+secret.ID.String()+"/rollback/invalid", nil, token)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

// TestSecretsHandler_GetStats_Integration tests GET /secrets/stats
func TestSecretsHandler_GetStats_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	// Clean up existing secrets
	tc.ExecuteSQL(`DELETE FROM functions.secrets`)

	t.Run("get stats with secrets", func(t *testing.T) {
		// Create test secrets
		for i := 0; i < 3; i++ {
			secret := &secrets.Secret{Name: uuid.New().String(), Scope: "global"}
			storage.CreateSecret(context.Background(), secret, "value", nil)
		}

		// Create expiring soon
		expiringSoon := time.Now().Add(3 * 24 * time.Hour)
		secret := &secrets.Secret{Name: "EXPIRING_SOON_STATS", Scope: "global", ExpiresAt: &expiringSoon}
		storage.CreateSecret(context.Background(), secret, "value", nil)

		// Create expired
		expired := time.Now().Add(-1 * time.Hour)
		secret = &secrets.Secret{Name: "EXPIRED_STATS", Scope: "global", ExpiresAt: &expired}
		storage.CreateSecret(context.Background(), secret, "value", nil)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets/stats", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]int
		err := json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)
		assert.Equal(t, 5, stats["total"])
		assert.Equal(t, 1, stats["expiring_soon"])
		assert.Equal(t, 1, stats["expired"])
	})

	t.Run("get empty stats", func(t *testing.T) {
		tc.ExecuteSQL(`DELETE FROM functions.secrets`)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets/stats", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var stats map[string]int
		err := json.NewDecoder(resp.Body).Decode(&stats)
		require.NoError(t, err)
		assert.Equal(t, 0, stats["total"])
		assert.Equal(t, 0, stats["expiring_soon"])
		assert.Equal(t, 0, stats["expired"])
	})
}

// TestSecretsHandler_Expiration_Integration tests expiration handling
func TestSecretsHandler_Expiration_Integration(t *testing.T) {
	tc := testutil.NewIntegrationTestContext(t)
	defer tc.CleanupTestData()

	app := setupSecretsApp(t, tc)
	uniqueEmail := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	_, token := tc.CreateTestUser(uniqueEmail, "password123")

	encryptionKey := "12345678901234567890123456789012"
	storage := secrets.NewStorage(tc.DB, encryptionKey)

	t.Run("expired secret is marked in list", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)
		secret := &secrets.Secret{Name: "EXPIRED_IN_LIST", Scope: "global", ExpiresAt: &past}
		err := storage.CreateSecret(context.Background(), secret, "value", nil)
		require.NoError(t, err)

		resp := makeRequest(t, app, "GET", "/api/v1/secrets", nil, token)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var results []secrets.SecretSummary
		err = json.NewDecoder(resp.Body).Decode(&results)
		require.NoError(t, err)

		// Find our expired secret
		var expiredSecret *secrets.SecretSummary
		for _, s := range results {
			if s.Name == "EXPIRED_IN_LIST" {
				expiredSecret = &s
				break
			}
		}
		require.NotNil(t, expiredSecret)
		assert.True(t, expiredSecret.IsExpired)
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// setupSecretsApp creates a Fiber app with secrets routes for testing
func setupSecretsApp(t *testing.T, tc *testutil.IntegrationTestContext) *fiber.App {
	// Get database connection
	db := tc.DB

	// Create encryption key (32 bytes for AES-256)
	encryptionKey := "12345678901234567890123456789012"

	// Create storage
	storage := secrets.NewStorage(db, encryptionKey)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Fluxbase Secrets Test",
	})

	// Create minimal config
	authCfg := &config.AuthConfig{
		JWTSecret:     "test-jwt-secret-for-integration-tests-32-chars",
		JWTExpiry:     time.Hour,
		RefreshExpiry: 24 * time.Hour,
	}

	// Create auth service
	authService := auth.NewService(db, authCfg, nil, "http://localhost:3000")
	jwtManager := auth.NewJWTManager(authCfg.JWTSecret, authCfg.JWTExpiry, authCfg.RefreshExpiry)
	clientKeyService := auth.NewClientKeyService(db.Pool(), nil)

	// Create handler
	handler := secrets.NewHandler(storage)

	// Register routes
	handler.RegisterRoutes(app, authService, clientKeyService, db.Pool(), jwtManager)

	return app
}

// makeRequest makes an HTTP request to the test app
func makeRequest(t *testing.T, app *fiber.App, method, path string, body interface{}, token string) *http.Response {
	t.Helper()

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		require.NoError(t, err)
		req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := app.Test(req)
		require.NoError(t, err)
		return resp
	}

	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}
