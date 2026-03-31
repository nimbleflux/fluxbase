package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTenantConfigLoader_GetConfigForSlug(t *testing.T) {
	// Create a base config
	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret:     "base-secret-that-is-at-least-32-chars!",
			JWTExpiry:     15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		Storage: StorageConfig{
			Provider:  "local",
			LocalPath: "./storage",
		},
		Email: EmailConfig{
			Provider:    "smtp",
			FromAddress: "noreply@example.com",
		},
		Functions: FunctionsConfig{
			DefaultTimeout: 30, // seconds
		},
		Jobs: JobsConfig{
			EmbeddedWorkerCount: 2,
		},
		AI: AIConfig{
			DefaultModel: "gpt-4",
		},
		Realtime: RealtimeConfig{
			MaxConnections: 100,
		},
		API: APIConfig{
			MaxPageSize: 1000,
		},
		GraphQL: GraphQLConfig{
			MaxDepth: 10,
		},
		RPC: RPCConfig{
			DefaultMaxRows: 1000,
		},
		Tenants: TenantsConfig{
			Configs: map[string]TenantOverrides{
				"tenant-a": {
					Auth: &AuthConfig{
						JWTSecret: "tenant-a-secret-that-is-at-least-32-chars!",
						JWTExpiry: 30 * time.Minute,
					},
					Storage: &StorageConfig{
						Provider:    "s3",
						S3Bucket:    "tenant-a-bucket",
						S3Region:    "us-east-1",
						S3SecretKey: "tenant-a-secret",
					},
				},
				"tenant-b": {
					Auth: &AuthConfig{
						JWTExpiry: 1 * time.Hour,
					},
					Email: &EmailConfig{
						Provider:    "ses",
						FromAddress: "noreply@tenantb.com",
					},
				},
			},
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	t.Run("returns base config for unknown tenant", func(t *testing.T) {
		cfg := loader.GetConfigForSlug("unknown-tenant")
		if cfg.Auth.JWTSecret != baseConfig.Auth.JWTSecret {
			t.Errorf("Expected base JWT secret for unknown tenant, got %s", cfg.Auth.JWTSecret)
		}
		if cfg.Storage.Provider != "local" {
			t.Errorf("Expected base storage provider, got %s", cfg.Storage.Provider)
		}
	})

	t.Run("returns merged config for tenant-a", func(t *testing.T) {
		cfg := loader.GetConfigForSlug("tenant-a")

		// Should have tenant-a's JWT secret
		if cfg.Auth.JWTSecret != "tenant-a-secret-that-is-at-least-32-chars!" {
			t.Errorf("Expected tenant-a JWT secret, got %s", cfg.Auth.JWTSecret)
		}

		// Should have tenant-a's JWT expiry
		if cfg.Auth.JWTExpiry != 30*time.Minute {
			t.Errorf("Expected tenant-a JWT expiry of 30m, got %v", cfg.Auth.JWTExpiry)
		}

		// Should have tenant-a's storage config
		if cfg.Storage.Provider != "s3" {
			t.Errorf("Expected tenant-a storage provider s3, got %s", cfg.Storage.Provider)
		}
		if cfg.Storage.S3Bucket != "tenant-a-bucket" {
			t.Errorf("Expected tenant-a S3 bucket, got %s", cfg.Storage.S3Bucket)
		}

		// Should preserve base config for non-overridden fields
		if cfg.Email.Provider != "smtp" {
			t.Errorf("Expected base email provider smtp, got %s", cfg.Email.Provider)
		}
		if cfg.Email.FromAddress != "noreply@example.com" {
			t.Errorf("Expected base email from address, got %s", cfg.Email.FromAddress)
		}
	})

	t.Run("returns merged config for tenant-b", func(t *testing.T) {
		cfg := loader.GetConfigForSlug("tenant-b")

		// Should have base JWT secret (not overridden)
		if cfg.Auth.JWTSecret != baseConfig.Auth.JWTSecret {
			t.Errorf("Expected base JWT secret for tenant-b, got %s", cfg.Auth.JWTSecret)
		}

		// Should have tenant-b's JWT expiry
		if cfg.Auth.JWTExpiry != 1*time.Hour {
			t.Errorf("Expected tenant-b JWT expiry of 1h, got %v", cfg.Auth.JWTExpiry)
		}

		// Should have tenant-b's email config
		if cfg.Email.Provider != "ses" {
			t.Errorf("Expected tenant-b email provider ses, got %s", cfg.Email.Provider)
		}
		if cfg.Email.FromAddress != "noreply@tenantb.com" {
			t.Errorf("Expected tenant-b email from address, got %s", cfg.Email.FromAddress)
		}

		// Should preserve base storage config
		if cfg.Storage.Provider != "local" {
			t.Errorf("Expected base storage provider local, got %s", cfg.Storage.Provider)
		}
	})
}

func TestTenantConfigLoader_GetLoadedSlugs(t *testing.T) {
	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
		Tenants: TenantsConfig{
			Configs: map[string]TenantOverrides{
				"tenant-x": {},
				"tenant-y": {},
			},
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	slugs := loader.GetLoadedSlugs()
	if len(slugs) != 2 {
		t.Errorf("Expected 2 slugs, got %d", len(slugs))
	}

	// Check that both slugs are present
	slugMap := make(map[string]bool)
	for _, slug := range slugs {
		slugMap[slug] = true
	}
	if !slugMap["tenant-x"] || !slugMap["tenant-y"] {
		t.Errorf("Expected tenant-x and tenant-y in slugs, got %v", slugs)
	}
}

func TestTenantConfigLoader_GetBaseConfig(t *testing.T) {
	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	returned := loader.GetBaseConfig()
	if returned.Auth.JWTSecret != baseConfig.Auth.JWTSecret {
		t.Errorf("GetBaseConfig did not return the base config")
	}
}

func TestTenantConfigLoader_DeepMergeDoesNotAffectBase(t *testing.T) {
	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
			JWTExpiry: 15 * time.Minute,
			OAuthProviders: []OAuthProviderConfig{
				{Name: "google", ClientID: "base-client-id"},
			},
		},
		Tenants: TenantsConfig{
			Configs: map[string]TenantOverrides{
				"tenant-c": {
					Auth: &AuthConfig{
						JWTExpiry: 1 * time.Hour,
					},
				},
			},
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Get tenant config
	tenantCfg := loader.GetConfigForSlug("tenant-c")

	// Modify the tenant config
	tenantCfg.Auth.JWTSecret = "modified"
	tenantCfg.Auth.OAuthProviders[0].ClientID = "modified"

	// Base config should be unchanged (deep copy was made)
	if baseConfig.Auth.JWTSecret != "base-secret-that-is-at-least-32-chars!" {
		t.Error("Modifying tenant config affected base config JWT secret")
	}
	if baseConfig.Auth.OAuthProviders[0].ClientID != "base-client-id" {
		t.Error("Modifying tenant config affected base config OAuth providers")
	}
}

func TestTenantConfigLoader_EmptyTenantConfigs(t *testing.T) {
	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
		Tenants: TenantsConfig{
			Configs: map[string]TenantOverrides{
				"empty-tenant": {},
			},
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	cfg := loader.GetConfigForSlug("empty-tenant")

	// Should have base config values
	if cfg.Auth.JWTSecret != baseConfig.Auth.JWTSecret {
		t.Errorf("Empty tenant config should use base config")
	}
}

func TestTenantConfigLoader_EnvVarExpansion(t *testing.T) {
	// Set test environment variables
	t.Setenv("TEST_JWT_SECRET_123", "test-secret-from-env-at-least-32-chars!")
	t.Setenv("TEST_S3_BUCKET_123", "test-bucket-from-env")

	// Create a temp directory for tenant config files
	tempDir := t.TempDir()
	tenantFile := filepath.Join(tempDir, "test-tenant.yaml")

	// Write a tenant config file with env var references
	// Note: YAML field names must match struct field names (case-insensitive)
	// since the config structs use mapstructure tags, not yaml tags
	content := `slug: test-tenant
name: Test Tenant
config:
  auth:
    jwt_secret: "${TEST_JWT_SECRET_123}"
  storage:
    s3_bucket: "${TEST_S3_BUCKET_123}"
    provider: s3
`
	if err := os.WriteFile(tenantFile, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to write tenant config file: %v", err)
	}

	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
		Storage: StorageConfig{
			Provider: "local",
		},
		Tenants: TenantsConfig{
			ConfigDir: tempDir,
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// Debug: check what slugs are loaded
	t.Logf("Loaded slugs: %v", loader.GetLoadedSlugs())

	cfg := loader.GetConfigForSlug("test-tenant")

	// Check that env vars were expanded
	if cfg.Auth.JWTSecret != "test-secret-from-env-at-least-32-chars!" {
		t.Errorf("Expected JWT secret from env var, got %s", cfg.Auth.JWTSecret)
	}
	if cfg.Storage.S3Bucket != "test-bucket-from-env" {
		t.Errorf("Expected S3 bucket from env var, got %s", cfg.Storage.S3Bucket)
	}
}

func TestTenantConfigLoader_EnvVarOverrides(t *testing.T) {
	// Set tenant-specific environment variables
	// Pattern: FLUXBASE_TENANTS__<SLUG>__<SECTION>__<KEY>
	t.Setenv("FLUXBASE_TENANTS__ENV_TENANT__AUTH__JWT_SECRET", "env-tenant-secret-at-least-32-chars!")
	t.Setenv("FLUXBASE_TENANTS__ENV_TENANT__STORAGE__S3_BUCKET", "env-tenant-bucket")
	t.Setenv("FLUXBASE_TENANTS__ENV_TENANT__STORAGE__PROVIDER", "s3")
	t.Setenv("FLUXBASE_TENANTS__ENV_TENANT__EMAIL__FROM_ADDRESS", "env@tenant.com")

	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
		Storage: StorageConfig{
			Provider: "local",
		},
		Email: EmailConfig{
			FromAddress: "base@example.com",
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	// The env tenant should be automatically created from env vars
	cfg := loader.GetConfigForSlug("env-tenant")

	// Check that env var overrides were applied
	if cfg.Auth.JWTSecret != "env-tenant-secret-at-least-32-chars!" {
		t.Errorf("Expected JWT secret from env var, got %s", cfg.Auth.JWTSecret)
	}
	if cfg.Storage.S3Bucket != "env-tenant-bucket" {
		t.Errorf("Expected S3 bucket from env var, got %s", cfg.Storage.S3Bucket)
	}
	if cfg.Storage.Provider != "s3" {
		t.Errorf("Expected storage provider s3 from env var, got %s", cfg.Storage.Provider)
	}
	if cfg.Email.FromAddress != "env@tenant.com" {
		t.Errorf("Expected email from address from env var, got %s", cfg.Email.FromAddress)
	}
}

func TestTenantConfigLoader_EnvVarOverridesExistingConfig(t *testing.T) {
	// Set env vars that should override inline config
	t.Setenv("FLUXBASE_TENANTS__OVERRIDE_TENANT__AUTH__JWT_SECRET", "overridden-secret-at-least-32-chars!")

	baseConfig := &Config{
		Auth: AuthConfig{
			JWTSecret: "base-secret-that-is-at-least-32-chars!",
		},
		Tenants: TenantsConfig{
			Configs: map[string]TenantOverrides{
				"override-tenant": {
					Auth: &AuthConfig{
						JWTSecret: "inline-secret-at-least-32-chars!",
					},
				},
			},
		},
	}

	loader, err := NewTenantConfigLoader(baseConfig)
	if err != nil {
		t.Fatalf("Failed to create loader: %v", err)
	}

	cfg := loader.GetConfigForSlug("override-tenant")

	// Env var should override inline config
	if cfg.Auth.JWTSecret != "overridden-secret-at-least-32-chars!" {
		t.Errorf("Expected JWT secret from env var to override inline, got %s", cfg.Auth.JWTSecret)
	}
}
