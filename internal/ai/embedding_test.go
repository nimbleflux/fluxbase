package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbeddingModels(t *testing.T) {
	t.Run("OpenAI embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, OpenAIEmbeddingModels)
		assert.Len(t, OpenAIEmbeddingModels, 3)

		// Check text-embedding-3-small
		found := false
		for _, m := range OpenAIEmbeddingModels {
			if m.Name == "text-embedding-3-small" {
				found = true
				assert.Equal(t, 1536, m.Dimensions)
				assert.Equal(t, 8191, m.MaxTokens)
				break
			}
		}
		assert.True(t, found, "text-embedding-3-small model should be defined")
	})

	t.Run("Azure embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, AzureEmbeddingModels)
		assert.Len(t, AzureEmbeddingModels, 3)
	})

	t.Run("Ollama embedding models are defined", func(t *testing.T) {
		assert.NotEmpty(t, OllamaEmbeddingModels)
		assert.Len(t, OllamaEmbeddingModels, 3)

		// Check nomic-embed-text
		found := false
		for _, m := range OllamaEmbeddingModels {
			if m.Name == "nomic-embed-text" {
				found = true
				assert.Equal(t, 768, m.Dimensions)
				break
			}
		}
		assert.True(t, found, "nomic-embed-text model should be defined")
	})
}

func TestGetEmbeddingModelDimensions(t *testing.T) {
	testCases := []struct {
		model      string
		dimensions int
	}{
		{"text-embedding-3-small", 1536},
		{"text-embedding-3-large", 3072},
		{"text-embedding-ada-002", 1536},
		{"nomic-embed-text", 768},
		{"mxbai-embed-large", 1024},
		{"all-minilm", 384},
		{"unknown-model", 0},
		{"", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			dims := GetEmbeddingModelDimensions(tc.model)
			assert.Equal(t, tc.dimensions, dims)
		})
	}
}

func TestNewEmbeddingProvider(t *testing.T) {
	t.Run("errors on unsupported provider type", func(t *testing.T) {
		config := ProviderConfig{
			Type: "unsupported-provider",
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "unsupported embedding provider type")
	})

	t.Run("errors on OpenAI without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOpenAI,
			Config: map[string]string{},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("errors on Azure without api_key", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("errors on Azure without endpoint", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":         "test-key",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "endpoint is required")
	})

	t.Run("errors on Azure without deployment_name", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeAzure,
			Config: map[string]string{
				"api_key":  "test-key",
				"endpoint": "https://example.openai.azure.com",
			},
		}

		provider, err := NewEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "deployment_name is required")
	})
}

func TestEmbeddingRequest_Struct(t *testing.T) {
	req := EmbeddingRequest{
		Texts: []string{"Hello world", "Another text"},
		Model: "text-embedding-3-small",
	}

	assert.Equal(t, 2, len(req.Texts))
	assert.Equal(t, "text-embedding-3-small", req.Model)
}

func TestEmbeddingResponse_Struct(t *testing.T) {
	resp := EmbeddingResponse{
		Embeddings: [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
		Model:      "text-embedding-3-small",
		Dimensions: 1536,
		Usage: &EmbeddingUsage{
			PromptTokens: 10,
			TotalTokens:  10,
		},
	}

	assert.Equal(t, 2, len(resp.Embeddings))
	assert.Equal(t, "text-embedding-3-small", resp.Model)
	assert.Equal(t, 1536, resp.Dimensions)
	require.NotNil(t, resp.Usage)
	assert.Equal(t, 10, resp.Usage.PromptTokens)
}

func TestEmbeddingModel_Struct(t *testing.T) {
	model := EmbeddingModel{
		Name:       "custom-model",
		Dimensions: 512,
		MaxTokens:  4096,
	}

	assert.Equal(t, "custom-model", model.Name)
	assert.Equal(t, 512, model.Dimensions)
	assert.Equal(t, 4096, model.MaxTokens)
}

// =============================================================================
// Ollama Embedding Provider Tests
// =============================================================================

func TestNewOllamaEmbeddingProvider(t *testing.T) {
	t.Run("creates provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOllama,
			Model: "nomic-embed-text",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults endpoint to localhost:11434", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Model:  "nomic-embed-text",
			Config: map[string]string{},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults model to nomic-embed-text", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Config: map[string]string{},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "nomic-embed-text", provider.DefaultModel())
	})

	t.Run("uses custom model when specified", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOllama,
			Model: "mxbai-embed-large",
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
			},
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "mxbai-embed-large", provider.DefaultModel())
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOllama,
			Model:  "nomic-embed-text",
			Config: nil,
		}

		provider, err := NewOllamaEmbeddingProvider(config)
		require.NoError(t, err) // Should default endpoint
		assert.NotNil(t, provider)
	})
}

// =============================================================================
// EmbeddingUsage Struct Tests
// =============================================================================

func TestEmbeddingUsage_Struct(t *testing.T) {
	t.Run("zero value has expected defaults", func(t *testing.T) {
		var usage EmbeddingUsage

		assert.Zero(t, usage.PromptTokens)
		assert.Zero(t, usage.TotalTokens)
	})

	t.Run("all fields can be set", func(t *testing.T) {
		usage := EmbeddingUsage{
			PromptTokens: 100,
			TotalTokens:  100,
		}

		assert.Equal(t, 100, usage.PromptTokens)
		assert.Equal(t, 100, usage.TotalTokens)
	})
}

// =============================================================================
// OpenAI Embedding Provider Tests
// =============================================================================

func TestNewOpenAIEmbeddingProvider_ConfigValidation(t *testing.T) {
	t.Run("creates provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOpenAI,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key": "test-key",
			},
		}

		provider, err := NewOpenAIEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "text-embedding-3-small", provider.DefaultModel())
	})

	t.Run("defaults model when not specified", func(t *testing.T) {
		config := ProviderConfig{
			Type: ProviderTypeOpenAI,
			Config: map[string]string{
				"api_key": "test-key",
			},
		}

		provider, err := NewOpenAIEmbeddingProvider(config)
		require.NoError(t, err)
		assert.Equal(t, "text-embedding-3-small", provider.DefaultModel())
	})

	t.Run("accepts custom base_url", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOpenAI,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key":  "test-key",
				"base_url": "https://custom.openai.com",
			},
		}

		provider, err := NewOpenAIEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("accepts organization_id", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeOpenAI,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key":         "test-key",
				"organization_id": "org-123",
			},
		}

		provider, err := NewOpenAIEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeOpenAI,
			Model:  "text-embedding-3-small",
			Config: nil,
		}

		provider, err := NewOpenAIEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

// =============================================================================
// Azure Embedding Provider Tests
// =============================================================================

func TestNewAzureEmbeddingProvider_Validation(t *testing.T) {
	t.Run("creates provider with valid config", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeAzure,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("defaults api_version when not specified", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeAzure,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
			},
		}

		provider, err := NewAzureEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("uses custom api_version when specified", func(t *testing.T) {
		config := ProviderConfig{
			Type:  ProviderTypeAzure,
			Model: "text-embedding-3-small",
			Config: map[string]string{
				"api_key":         "test-key",
				"endpoint":        "https://example.openai.azure.com",
				"deployment_name": "my-deployment",
				"api_version":     "2023-05-15",
			},
		}

		provider, err := NewAzureEmbeddingProvider(config)
		require.NoError(t, err)
		assert.NotNil(t, provider)
	})

	t.Run("handles nil Config map", func(t *testing.T) {
		config := ProviderConfig{
			Type:   ProviderTypeAzure,
			Model:  "text-embedding-3-small",
			Config: nil,
		}

		provider, err := NewAzureEmbeddingProvider(config)
		require.Error(t, err)
		assert.Nil(t, provider)
		assert.Contains(t, err.Error(), "api_key is required")
	})
}

// =============================================================================
// Embedding Model Dimensions Tests
// =============================================================================

func TestGetEmbeddingModelDimensions_Comprehensive(t *testing.T) {
	t.Run("returns correct dimensions for OpenAI models", func(t *testing.T) {
		tests := []struct {
			model      string
			dimensions int
		}{
			{"text-embedding-3-small", 1536},
			{"text-embedding-3-large", 3072},
			{"text-embedding-ada-002", 1536},
		}

		for _, tc := range tests {
			t.Run(tc.model, func(t *testing.T) {
				dims := GetEmbeddingModelDimensions(tc.model)
				assert.Equal(t, tc.dimensions, dims)
			})
		}
	})

	t.Run("returns correct dimensions for Ollama models", func(t *testing.T) {
		tests := []struct {
			model      string
			dimensions int
		}{
			{"nomic-embed-text", 768},
			{"mxbai-embed-large", 1024},
			{"all-minilm", 384},
		}

		for _, tc := range tests {
			t.Run(tc.model, func(t *testing.T) {
				dims := GetEmbeddingModelDimensions(tc.model)
				assert.Equal(t, tc.dimensions, dims)
			})
		}
	})

	t.Run("returns 0 for unknown model", func(t *testing.T) {
		dims := GetEmbeddingModelDimensions("unknown-model-xyz")
		assert.Zero(t, dims)
	})

	t.Run("returns 0 for empty string", func(t *testing.T) {
		dims := GetEmbeddingModelDimensions("")
		assert.Zero(t, dims)
	})

	t.Run("case sensitive model names", func(t *testing.T) {
		dims := GetEmbeddingModelDimensions("Text-Embedding-3-Small")
		assert.Zero(t, dims) // Should be case sensitive
	})
}

// =============================================================================
// Embedding Request/Response Tests
// =============================================================================

func TestEmbeddingRequest_Validation(t *testing.T) {
	t.Run("valid request with all fields", func(t *testing.T) {
		req := EmbeddingRequest{
			Texts: []string{"text1", "text2"},
			Model: "text-embedding-3-small",
		}
		assert.Len(t, req.Texts, 2)
		assert.Equal(t, "text-embedding-3-small", req.Model)
	})

	t.Run("empty texts array", func(t *testing.T) {
		req := EmbeddingRequest{
			Texts: []string{},
			Model: "text-embedding-3-small",
		}
		assert.Empty(t, req.Texts)
	})

	t.Run("nil texts array", func(t *testing.T) {
		req := EmbeddingRequest{
			Texts: nil,
			Model: "text-embedding-3-small",
		}
		assert.Nil(t, req.Texts)
	})

	t.Run("single text", func(t *testing.T) {
		req := EmbeddingRequest{
			Texts: []string{"single text"},
			Model: "text-embedding-3-small",
		}
		assert.Len(t, req.Texts, 1)
	})

	t.Run("empty model string", func(t *testing.T) {
		req := EmbeddingRequest{
			Texts: []string{"text"},
			Model: "",
		}
		assert.Equal(t, "", req.Model)
	})
}

func TestEmbeddingResponse_Validation(t *testing.T) {
	t.Run("complete response", func(t *testing.T) {
		resp := EmbeddingResponse{
			Embeddings: [][]float32{
				{0.1, 0.2, 0.3},
				{0.4, 0.5, 0.6},
			},
			Model:      "text-embedding-3-small",
			Dimensions: 1536,
			Usage: &EmbeddingUsage{
				PromptTokens: 10,
				TotalTokens:  10,
			},
		}

		assert.Len(t, resp.Embeddings, 2)
		assert.Equal(t, "text-embedding-3-small", resp.Model)
		assert.Equal(t, 1536, resp.Dimensions)
		assert.NotNil(t, resp.Usage)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
	})

	t.Run("response without usage", func(t *testing.T) {
		resp := EmbeddingResponse{
			Embeddings: [][]float32{{0.1, 0.2, 0.3}},
			Model:      "text-embedding-3-small",
			Dimensions: 1536,
			Usage:      nil,
		}

		assert.Nil(t, resp.Usage)
	})

	t.Run("empty embeddings array", func(t *testing.T) {
		resp := EmbeddingResponse{
			Embeddings: [][]float32{},
			Model:      "text-embedding-3-small",
			Dimensions: 0,
		}

		assert.Empty(t, resp.Embeddings)
		assert.Zero(t, resp.Dimensions)
	})

	t.Run("zero dimensions", func(t *testing.T) {
		resp := EmbeddingResponse{
			Embeddings: [][]float32{},
			Model:      "text-embedding-3-small",
			Dimensions: 0,
		}

		assert.Zero(t, resp.Dimensions)
	})
}

// =============================================================================
// Embedding Model List Tests
// =============================================================================

func TestEmbeddingModelLists(t *testing.T) {
	t.Run("OpenAI models have expected properties", func(t *testing.T) {
		assert.NotEmpty(t, OpenAIEmbeddingModels)

		for _, model := range OpenAIEmbeddingModels {
			assert.NotEmpty(t, model.Name)
			assert.Greater(t, model.Dimensions, 0)
			assert.Greater(t, model.MaxTokens, 0)
		}
	})

	t.Run("Azure models have expected properties", func(t *testing.T) {
		assert.NotEmpty(t, AzureEmbeddingModels)

		for _, model := range AzureEmbeddingModels {
			assert.NotEmpty(t, model.Name)
			assert.Greater(t, model.Dimensions, 0)
			assert.Greater(t, model.MaxTokens, 0)
		}
	})

	t.Run("Ollama models have expected properties", func(t *testing.T) {
		assert.NotEmpty(t, OllamaEmbeddingModels)

		for _, model := range OllamaEmbeddingModels {
			assert.NotEmpty(t, model.Name)
			assert.Greater(t, model.Dimensions, 0)
			assert.Greater(t, model.MaxTokens, 0)
		}
	})

	t.Run("all OpenAI models have unique names", func(t *testing.T) {
		names := make(map[string]bool)
		for _, model := range OpenAIEmbeddingModels {
			assert.False(t, names[model.Name], "duplicate model name: "+model.Name)
			names[model.Name] = true
		}
	})

	t.Run("all Ollama models have unique names", func(t *testing.T) {
		names := make(map[string]bool)
		for _, model := range OllamaEmbeddingModels {
			assert.False(t, names[model.Name], "duplicate model name: "+model.Name)
			names[model.Name] = true
		}
	})
}
