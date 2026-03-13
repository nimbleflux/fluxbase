package api

import (
	"context"
	"testing"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// VectorManager Construction Tests
// =============================================================================

func TestNewVectorManager(t *testing.T) {
	t.Run("creates manager with nil dependencies", func(t *testing.T) {
		cfg := &config.AIConfig{}
		manager := NewVectorManager(cfg, nil, nil, nil)

		require.NotNil(t, manager)
		assert.Nil(t, manager.GetEmbeddingService())
	})

	t.Run("stores env config", func(t *testing.T) {
		cfg := &config.AIConfig{
			EmbeddingEnabled: false,
		}
		manager := NewVectorManager(cfg, nil, nil, nil)

		require.NotNil(t, manager)
		assert.Equal(t, cfg, manager.envConfig)
	})
}

// =============================================================================
// GetEmbeddingService Tests
// =============================================================================

func TestVectorManager_GetEmbeddingService(t *testing.T) {
	t.Run("returns nil when no service configured", func(t *testing.T) {
		cfg := &config.AIConfig{}
		manager := NewVectorManager(cfg, nil, nil, nil)

		service := manager.GetEmbeddingService()

		assert.Nil(t, service)
	})

	t.Run("is thread-safe", func(t *testing.T) {
		cfg := &config.AIConfig{}
		manager := NewVectorManager(cfg, nil, nil, nil)

		// Call from multiple goroutines to verify no race condition
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_ = manager.GetEmbeddingService()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// =============================================================================
// buildEmbeddingConfigFromProvider Tests
// =============================================================================

func TestVectorManager_buildEmbeddingConfigFromProvider(t *testing.T) {
	cfg := &config.AIConfig{}
	manager := NewVectorManager(cfg, nil, nil, nil)

	t.Run("returns error for nil provider", func(t *testing.T) {
		_, err := manager.buildEmbeddingConfigFromProvider(nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "provider is nil")
	})

	t.Run("returns error for unsupported provider type", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "unsupported_provider",
			Config:       map[string]string{},
		}

		_, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported provider type")
	})

	// OpenAI provider tests
	t.Run("openai - returns error without api_key", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "openai",
			Config:       map[string]string{},
		}

		_, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai provider missing api_key")
	})

	t.Run("openai - builds config with api_key", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "openai",
			Config: map[string]string{
				"api_key": "sk-test-key",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, ai.ProviderType("openai"), cfg.Provider.Type)
		assert.Equal(t, "sk-test-key", cfg.Provider.Config["api_key"])
		assert.Equal(t, "text-embedding-3-small", cfg.DefaultModel) // default model
		assert.True(t, cfg.CacheEnabled)
	})

	t.Run("openai - uses explicit embedding model", func(t *testing.T) {
		embeddingModel := "text-embedding-ada-002"
		provider := &ai.ProviderRecord{
			ProviderType:   "openai",
			EmbeddingModel: &embeddingModel,
			Config: map[string]string{
				"api_key": "sk-test-key",
				"model":   "gpt-4", // This should be ignored in favor of EmbeddingModel
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "text-embedding-ada-002", cfg.DefaultModel)
	})

	t.Run("openai - falls back to config model when no embedding model", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "openai",
			Config: map[string]string{
				"api_key": "sk-test-key",
				"model":   "text-embedding-custom",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "text-embedding-custom", cfg.DefaultModel)
	})

	t.Run("openai - includes optional organization_id", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "openai",
			Config: map[string]string{
				"api_key":         "sk-test-key",
				"organization_id": "org-123",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "org-123", cfg.Provider.Config["organization_id"])
	})

	t.Run("openai - includes optional base_url", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "openai",
			Config: map[string]string{
				"api_key":  "sk-test-key",
				"base_url": "https://custom.openai.com",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "https://custom.openai.com", cfg.Provider.Config["base_url"])
	})

	// Azure provider tests
	t.Run("azure - returns error without api_key", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "azure",
			Config: map[string]string{
				"endpoint":        "https://test.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		_, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "azure provider missing api_key")
	})

	t.Run("azure - returns error without endpoint", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "azure",
			Config: map[string]string{
				"api_key":         "test-key",
				"deployment_name": "my-deployment",
			},
		}

		_, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "azure provider missing endpoint")
	})

	t.Run("azure - returns error without deployment_name", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "azure",
			Config: map[string]string{
				"api_key":  "test-key",
				"endpoint": "https://test.openai.azure.com",
			},
		}

		_, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "azure provider missing deployment_name")
	})

	t.Run("azure - builds complete config", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "azure",
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://test.openai.azure.com",
				"deployment_name": "my-deployment",
				"api_version":     "2023-05-15",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, ai.ProviderType("azure"), cfg.Provider.Type)
		assert.Equal(t, "test-key", cfg.Provider.Config["api_key"])
		assert.Equal(t, "https://test.openai.azure.com", cfg.Provider.Config["endpoint"])
		assert.Equal(t, "my-deployment", cfg.Provider.Config["deployment_name"])
		assert.Equal(t, "2023-05-15", cfg.Provider.Config["api_version"])
		assert.Equal(t, "text-embedding-ada-002", cfg.DefaultModel) // default
	})

	t.Run("azure - uses explicit embedding model", func(t *testing.T) {
		embeddingModel := "text-embedding-3-large"
		provider := &ai.ProviderRecord{
			ProviderType:   "azure",
			EmbeddingModel: &embeddingModel,
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://test.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "text-embedding-3-large", cfg.DefaultModel)
	})

	// Ollama provider tests
	t.Run("ollama - builds config with defaults", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "ollama",
			Config:       map[string]string{},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, ai.ProviderType("ollama"), cfg.Provider.Type)
		assert.Equal(t, "nomic-embed-text", cfg.DefaultModel) // default
	})

	t.Run("ollama - uses custom endpoint", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "ollama",
			Config: map[string]string{
				"endpoint": "http://custom-ollama:11434",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "http://custom-ollama:11434", cfg.Provider.Config["endpoint"])
	})

	t.Run("ollama - uses explicit embedding model", func(t *testing.T) {
		embeddingModel := "mxbai-embed-large"
		provider := &ai.ProviderRecord{
			ProviderType:   "ollama",
			EmbeddingModel: &embeddingModel,
			Config:         map[string]string{},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "mxbai-embed-large", cfg.DefaultModel)
	})

	t.Run("ollama - falls back to config model", func(t *testing.T) {
		provider := &ai.ProviderRecord{
			ProviderType: "ollama",
			Config: map[string]string{
				"model": "all-minilm",
			},
		}

		cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

		require.NoError(t, err)
		assert.Equal(t, "all-minilm", cfg.DefaultModel)
	})
}

// =============================================================================
// RefreshFromDatabase Tests
// =============================================================================

func TestVectorManager_RefreshFromDatabase_SkipsWithEnvConfig(t *testing.T) {
	t.Run("skips refresh when embedding enabled in env config", func(t *testing.T) {
		cfg := &config.AIConfig{
			EmbeddingEnabled: true,
		}
		manager := NewVectorManager(cfg, nil, nil, nil)

		// This should return nil without error because env config takes priority
		err := manager.RefreshFromDatabase(context.TODO())

		assert.NoError(t, err)
	})

	t.Run("skips refresh when embedding provider set in env config", func(t *testing.T) {
		cfg := &config.AIConfig{
			EmbeddingProvider: "openai",
		}
		manager := NewVectorManager(cfg, nil, nil, nil)

		// This should return nil without error because env config takes priority
		err := manager.RefreshFromDatabase(context.TODO())

		assert.NoError(t, err)
	})
}

// =============================================================================
// VectorManager Struct Field Tests
// =============================================================================

func TestVectorManager_Struct(t *testing.T) {
	t.Run("has all expected fields", func(t *testing.T) {
		cfg := &config.AIConfig{}
		manager := NewVectorManager(cfg, nil, nil, nil)

		// Verify struct fields are accessible
		assert.NotNil(t, manager.envConfig)
		assert.Nil(t, manager.aiStorage)
		assert.Nil(t, manager.schemaInspector)
		assert.Nil(t, manager.db)
		assert.Nil(t, manager.embeddingService)
	})
}

// =============================================================================
// EmbeddingModel Priority Tests
// =============================================================================

func TestBuildEmbeddingConfigFromProvider_ModelPriority(t *testing.T) {
	cfg := &config.AIConfig{}
	manager := NewVectorManager(cfg, nil, nil, nil)

	testCases := []struct {
		name            string
		providerType    string
		embeddingModel  *string
		configModel     string
		expectedModel   string
		requiredConfigs map[string]string
	}{
		{
			name:           "openai - EmbeddingModel takes priority",
			providerType:   "openai",
			embeddingModel: strPtr("embedding-priority"),
			configModel:    "config-model",
			expectedModel:  "embedding-priority",
			requiredConfigs: map[string]string{
				"api_key": "test",
			},
		},
		{
			name:           "openai - config model fallback",
			providerType:   "openai",
			embeddingModel: nil,
			configModel:    "config-model",
			expectedModel:  "config-model",
			requiredConfigs: map[string]string{
				"api_key": "test",
			},
		},
		{
			name:           "openai - default model fallback",
			providerType:   "openai",
			embeddingModel: nil,
			configModel:    "",
			expectedModel:  "text-embedding-3-small",
			requiredConfigs: map[string]string{
				"api_key": "test",
			},
		},
		{
			name:           "azure - EmbeddingModel takes priority",
			providerType:   "azure",
			embeddingModel: strPtr("azure-embedding"),
			configModel:    "azure-config",
			expectedModel:  "azure-embedding",
			requiredConfigs: map[string]string{
				"api_key":         "test",
				"endpoint":        "https://test.azure.com",
				"deployment_name": "test-deploy",
			},
		},
		{
			name:           "azure - default model fallback",
			providerType:   "azure",
			embeddingModel: nil,
			configModel:    "",
			expectedModel:  "text-embedding-ada-002",
			requiredConfigs: map[string]string{
				"api_key":         "test",
				"endpoint":        "https://test.azure.com",
				"deployment_name": "test-deploy",
			},
		},
		{
			name:            "ollama - EmbeddingModel takes priority",
			providerType:    "ollama",
			embeddingModel:  strPtr("ollama-embedding"),
			configModel:     "ollama-config",
			expectedModel:   "ollama-embedding",
			requiredConfigs: map[string]string{},
		},
		{
			name:            "ollama - default model fallback",
			providerType:    "ollama",
			embeddingModel:  nil,
			configModel:     "",
			expectedModel:   "nomic-embed-text",
			requiredConfigs: map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := make(map[string]string)
			for k, v := range tc.requiredConfigs {
				config[k] = v
			}
			if tc.configModel != "" {
				config["model"] = tc.configModel
			}

			provider := &ai.ProviderRecord{
				ProviderType:   tc.providerType,
				EmbeddingModel: tc.embeddingModel,
				Config:         config,
			}

			cfg, err := manager.buildEmbeddingConfigFromProvider(provider)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedModel, cfg.DefaultModel)
			assert.Equal(t, tc.expectedModel, cfg.Provider.Model)
		})
	}
}

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkNewVectorManager(b *testing.B) {
	cfg := &config.AIConfig{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewVectorManager(cfg, nil, nil, nil)
	}
}

func BenchmarkVectorManager_GetEmbeddingService(b *testing.B) {
	cfg := &config.AIConfig{}
	manager := NewVectorManager(cfg, nil, nil, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetEmbeddingService()
	}
}

func BenchmarkVectorManager_buildEmbeddingConfigFromProvider(b *testing.B) {
	cfg := &config.AIConfig{}
	manager := NewVectorManager(cfg, nil, nil, nil)
	provider := &ai.ProviderRecord{
		ProviderType: "openai",
		Config: map[string]string{
			"api_key": "sk-test-key",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = manager.buildEmbeddingConfigFromProvider(provider)
	}
}
