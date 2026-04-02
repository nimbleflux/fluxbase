//go:build integration

package e2e

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/test"
)

// TestTenantConfigLoader_EnvVarOverrides tests that environment variable overrides work
// This is a unit-style test that runs in the e2e environment
func TestTenantConfigLoader_EnvVarOverrides(t *testing.T) {
	// Set tenant-specific environment variables
	// Pattern: FLUXBASE_TENANTS__<SLUG>__<SECTION>__<KEY>
	envVars := map[string]string{
		"FLUXBASE_TENANTS__E2E_TENANT__AUTH__JWT_SECRET":    "e2e-tenant-secret-at-least-32-characters!",
		"FLUXBASE_TENANTS__E2E_TENANT__STORAGE__S3_BUCKET":  "e2e-tenant-bucket",
		"FLUXBASE_TENANTS__E2E_TENANT__STORAGE__PROVIDER":   "s3",
		"FLUXBASE_TENANTS__E2E_TENANT__EMAIL__FROM_ADDRESS": "e2e@tenant.com",
		"FLUXBASE_TENANTS__E2E_TENANT__FUNCTIONS__TIMEOUT":  "60",
		"FLUXBASE_TENANTS__E2E_TENANT__JOBS__WORKER_COUNT":  "8",
		"FLUXBASE_TENANTS__E2E_TENANT__REALTIME__MAX_CONNS": "500",
		"FLUXBASE_TENANTS__E2E_TENANT__API__MAX_PAGE_SIZE":  "5000",
		"FLUXBASE_TENANTS__E2E_TENANT__GRAPHQL__MAX_DEPTH":  "20",
		"FLUXBASE_TENANTS__E2E_TENANT__RPC__MAX_ROWS":       "5000",
		"FLUXBASE_TENANTS__E2E_TENANT__AI__DEFAULT_MODEL":   "gpt-4-turbo",
	}

	// Set environment variables
	for k, v := range envVars {
		os.Setenv(k, v)
	}
	defer func() {
		for k := range envVars {
			os.Unsetenv(k)
		}
	}()

	// Create base config
	baseConfig := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "base-secret-at-least-32-characters!",
			JWTExpiry: 15 * time.Minute,
		},
		Storage: config.StorageConfig{
			Provider: "local",
		},
		Email: config.EmailConfig{
			FromAddress: "base@example.com",
		},
		Functions: config.FunctionsConfig{
			DefaultTimeout: 30,
		},
		Jobs: config.JobsConfig{
			EmbeddedWorkerCount: 2,
		},
		Realtime: config.RealtimeConfig{
			MaxConnections: 100,
		},
		API: config.APIConfig{
			MaxPageSize: 1000,
		},
		GraphQL: config.GraphQLConfig{
			MaxDepth: 10,
		},
		RPC: config.RPCConfig{
			DefaultMaxRows: 1000,
		},
		AI: config.AIConfig{
			DefaultModel: "gpt-3.5-turbo",
		},
	}

	// Create loader
	loader, err := config.NewTenantConfigLoader(baseConfig)
	require.NoError(t, err, "Failed to create tenant config loader")

	// Get tenant config
	tenantConfig := loader.GetConfigForSlug("e2e-tenant")

	// Verify overrides were applied
	require.Equal(t, "e2e-tenant-secret-at-least-32-characters!", tenantConfig.Auth.JWTSecret, "JWT secret should be overridden")
	require.Equal(t, "s3", tenantConfig.Storage.Provider, "Storage provider should be overridden")
	require.Equal(t, "e2e-tenant-bucket", tenantConfig.Storage.S3Bucket, "S3 bucket should be overridden")
	require.Equal(t, "e2e@tenant.com", tenantConfig.Email.FromAddress, "Email from address should be overridden")
	require.Equal(t, 60, tenantConfig.Functions.DefaultTimeout, "Functions timeout should be overridden")
	require.Equal(t, 8, tenantConfig.Jobs.EmbeddedWorkerCount, "Jobs worker count should be overridden")
	require.Equal(t, 500, tenantConfig.Realtime.MaxConnections, "Realtime max connections should be overridden")
	require.Equal(t, 5000, tenantConfig.API.MaxPageSize, "API max page size should be overridden")
	require.Equal(t, 20, tenantConfig.GraphQL.MaxDepth, "GraphQL max depth should be overridden")
	require.Equal(t, 5000, tenantConfig.RPC.DefaultMaxRows, "RPC max rows should be overridden")
	require.Equal(t, "gpt-4-turbo", tenantConfig.AI.DefaultModel, "AI default model should be overridden")

	// Verify base config is unchanged (deep copy was made)
	require.Equal(t, "base-secret-at-least-32-characters!", baseConfig.Auth.JWTSecret, "Base config should be unchanged")
	require.Equal(t, "local", baseConfig.Storage.Provider, "Base config storage should be unchanged")

	log.Info().Msg("Environment variable overrides test passed")
}

// TestTenantConfigIsolation tests that tenant configs are isolated from each other
func TestTenantConfigIsolation(t *testing.T) {
	baseConfig := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "base-secret-at-least-32-characters!",
		},
		Tenants: config.TenantsConfig{
			Configs: map[string]config.TenantOverrides{
				"tenant-a": {
					Auth: &config.AuthConfig{
						JWTSecret: "tenant-a-secret-at-least-32-chars!",
					},
				},
				"tenant-b": {
					Auth: &config.AuthConfig{
						JWTSecret: "tenant-b-secret-at-least-32-chars!",
					},
				},
			},
		},
	}

	loader, err := config.NewTenantConfigLoader(baseConfig)
	require.NoError(t, err, "Failed to create tenant config loader")

	// Get config for tenant-a
	configA := loader.GetConfigForSlug("tenant-a")
	require.Equal(t, "tenant-a-secret-at-least-32-chars!", configA.Auth.JWTSecret, "Tenant A should have its secret")

	// Get config for tenant-b
	configB := loader.GetConfigForSlug("tenant-b")
	require.Equal(t, "tenant-b-secret-at-least-32-chars!", configB.Auth.JWTSecret, "Tenant B should have its secret")

	// Verify they are different
	require.NotEqual(t, configA.Auth.JWTSecret, configB.Auth.JWTSecret, "Tenants should have different secrets")

	// Modify configA and verify configB is not affected
	configA.Auth.JWTSecret = "modified-secret-at-least-32-chars!"
	configBCheck := loader.GetConfigForSlug("tenant-b")
	require.Equal(t, "tenant-b-secret-at-least-32-chars!", configBCheck.Auth.JWTSecret, "Modifying configA should not affect configB")

	// Verify base is also not affected
	require.Equal(t, "base-secret-at-least-32-characters!", baseConfig.Auth.JWTSecret, "Base config should be unchanged")

	log.Info().Msg("Tenant config isolation test passed")
}

// TestTenantSlugNormalization tests that slug normalization works correctly
func TestTenantSlugNormalization(t *testing.T) {
	// Set env var with underscores (ACME_CORP -> acme-corp)
	os.Setenv("FLUXBASE_TENANTS__ACME_CORP__AUTH__JWT_SECRET", "acme-secret-at-least-32-characters!")
	defer os.Unsetenv("FLUXBASE_TENANTS__ACME_CORP__AUTH__JWT_SECRET")

	baseConfig := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "base-secret-at-least-32-characters!",
		},
	}

	loader, err := config.NewTenantConfigLoader(baseConfig)
	require.NoError(t, err, "Failed to create tenant config loader")

	// Lookup should work with normalized slug (acme-corp)
	tenantConfig := loader.GetConfigForSlug("acme-corp")
	require.Equal(t, "acme-secret-at-least-32-characters!", tenantConfig.Auth.JWTSecret, "Env var should be applied to normalized slug")

	log.Info().Msg("Tenant slug normalization test passed")
}

// TestTenantAPIEndpoint tests the tenant management API
func TestTenantAPIEndpoint(t *testing.T) {
	tc := test.NewTestContext(t)
	defer tc.Close()

	// Create instance admin user
	adminUserID, adminToken := tc.CreateDashboardAdminUser("tenant-admin-api@example.com", "Admin123!")

	// Create tenant via API with unique slug
	tenantSlug := "test-api-tenant-" + uuid.New().String()[:8]
	createReq := map[string]interface{}{
		"slug": tenantSlug,
		"name": "Test API Tenant",
		"metadata": map[string]interface{}{
			"plan":  "enterprise",
			"owner": "test-user",
		},
	}

	resp := tc.NewRequest("POST", "/api/v1/admin/tenants").
		WithBearerToken(adminToken).
		WithBody(createReq).
		Send()

	require.Equal(t, fiber.StatusCreated, resp.Status(), "Tenant creation should succeed")

	var createResp map[string]interface{}
	err := json.Unmarshal(resp.Body(), &createResp)
	require.NoError(t, err, "Should decode response")

	tenantData, ok := createResp["tenant"].(map[string]interface{})
	require.True(t, ok, "Response should contain tenant object")

	tenantID, ok := tenantData["id"].(string)
	require.True(t, ok, "Response should contain tenant ID")
	require.NotEmpty(t, tenantID, "Tenant ID should not be empty")

	log.Info().
		Str("tenant_id", tenantID).
		Str("admin_user_id", adminUserID).
		Msg("Tenant API endpoint test passed")
}

// TestTenantConfigFromYAML tests loading tenant configs from YAML files
func TestTenantConfigFromYAML(t *testing.T) {
	// Create a temporary directory for tenant configs
	tempDir := t.TempDir()

	// Write a tenant config file
	tenantConfigContent := `slug: yaml-tenant
name: YAML Tenant
metadata:
  plan: pro
  billing_email: billing@yaml-tenant.com
config:
  auth:
    jwt_expiry: 30m
  functions:
    default_timeout: 45
`

	configFile := tempDir + "/yaml-tenant.yaml"
	err := os.WriteFile(configFile, []byte(tenantConfigContent), 0o644)
	require.NoError(t, err, "Failed to write tenant config file")

	// Create base config with config dir
	baseConfig := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "base-secret-at-least-32-characters!",
			JWTExpiry: 15 * time.Minute,
		},
		Functions: config.FunctionsConfig{
			DefaultTimeout: 30,
		},
		Tenants: config.TenantsConfig{
			ConfigDir: tempDir,
		},
	}

	loader, err := config.NewTenantConfigLoader(baseConfig)
	require.NoError(t, err, "Failed to create tenant config loader")

	// Verify tenant was loaded
	slugs := loader.GetLoadedSlugs()
	require.Contains(t, slugs, "yaml-tenant", "YAML tenant should be loaded")

	// Get config and verify
	tenantConfig := loader.GetConfigForSlug("yaml-tenant")
	require.Equal(t, 30*time.Minute, tenantConfig.Auth.JWTExpiry, "JWT expiry should be from YAML")
	require.Equal(t, 45, tenantConfig.Functions.DefaultTimeout, "Functions timeout should be from YAML")

	// Base config should be preserved for non-overridden fields
	require.Equal(t, "base-secret-at-least-32-characters!", tenantConfig.Auth.JWTSecret, "JWT secret should be from base")

	log.Info().Msg("Tenant config from YAML test passed")
}

// TestStorageManagerWithTenant tests the storage manager functionality with tenant context
func TestStorageManagerWithTenant(t *testing.T) {
	tc := test.NewTestContext(t)
	defer tc.Close()

	tc.EnsureStorageSchema()

	// Create an API key for storage operations
	apiKey := tc.CreateAPIKey("storage-test", []string{"storage:read", "storage:write"})

	// Create a bucket
	bucketName := "test-tenant-bucket-" + uuid.New().String()[:8]
	createResp := tc.NewRequest("POST", "/api/v1/storage/buckets/"+bucketName).
		WithAPIKey(apiKey).
		Send()

	require.Equal(t, fiber.StatusCreated, createResp.Status(), "Bucket creation should succeed")

	// Upload a file (multipart form data required)
	fileContent := []byte("test content for tenant")
	{
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test-file.txt")
		require.NoError(t, err)
		_, err = part.Write(fileContent)
		require.NoError(t, err)
		require.NoError(t, writer.Close())

		req := httptest.NewRequest("POST", "/api/v1/storage/"+bucketName+"/test-file.txt", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("X-Client-Key", apiKey)
		resp, err := tc.App.Test(req)
		require.NoError(t, err)
		require.Equal(t, fiber.StatusCreated, resp.StatusCode, "File upload should succeed")
	}

	// Download the file
	downloadResp := tc.NewRequest("GET", "/api/v1/storage/"+bucketName+"/test-file.txt").
		WithAPIKey(apiKey).
		Send()

	require.Equal(t, fiber.StatusOK, downloadResp.Status(), "File download should succeed")
	require.Equal(t, fileContent, downloadResp.Body(), "File content should match")

	log.Info().
		Str("bucket", bucketName).
		Msg("Storage manager with tenant test passed")
}

// TestTenantConfigPriority tests that config values are resolved in correct priority order
func TestTenantConfigPriority(t *testing.T) {
	// Priority order (highest priority last):
	// 1. Hardcoded defaults
	// 2. Base YAML file
	// 3. Tenant YAML files
	// 4. Base environment variables
	// 5. Tenant-specific env vars

	// Set tenant env var
	os.Setenv("FLUXBASE_TENANTS__PRIORITY_TENANT__AUTH__JWT_EXPIRY", "2h")
	defer os.Unsetenv("FLUXBASE_TENANTS__PRIORITY_TENANT__AUTH__JWT_EXPIRY")

	// Create temp dir for tenant config
	tempDir := t.TempDir()
	tenantYAML := `slug: priority-tenant
config:
  auth:
    jwt_expiry: 1h
`
	err := os.WriteFile(tempDir+"/priority-tenant.yaml", []byte(tenantYAML), 0o644)
	require.NoError(t, err, "Failed to write tenant config")

	// Base config with 30m expiry
	baseConfig := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret: "base-secret-at-least-32-characters!",
			JWTExpiry: 30 * time.Minute, // Base level
		},
		Tenants: config.TenantsConfig{
			ConfigDir: tempDir,
		},
	}

	loader, err := config.NewTenantConfigLoader(baseConfig)
	require.NoError(t, err, "Failed to create loader")

	// Env var should override YAML file
	tenantConfig := loader.GetConfigForSlug("priority-tenant")
	require.Equal(t, 2*time.Hour, tenantConfig.Auth.JWTExpiry, "Env var should override YAML (2h > 1h)")

	log.Info().Msg("Tenant config priority test passed")
}
