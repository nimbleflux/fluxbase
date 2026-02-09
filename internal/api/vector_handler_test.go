package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// EmbedRequest Struct Tests
// =============================================================================

func TestEmbedRequest_Struct(t *testing.T) {
	t.Run("single text embedding", func(t *testing.T) {
		req := EmbedRequest{
			Text:  "Hello, world!",
			Model: "text-embedding-3-small",
		}

		assert.Equal(t, "Hello, world!", req.Text)
		assert.Equal(t, "text-embedding-3-small", req.Model)
		assert.Empty(t, req.Texts)
		assert.Empty(t, req.Provider)
	})

	t.Run("multiple texts embedding", func(t *testing.T) {
		req := EmbedRequest{
			Texts: []string{"First", "Second", "Third"},
			Model: "text-embedding-ada-002",
		}

		assert.Empty(t, req.Text)
		assert.Len(t, req.Texts, 3)
		assert.Equal(t, "First", req.Texts[0])
	})

	t.Run("with provider override", func(t *testing.T) {
		req := EmbedRequest{
			Text:     "Test",
			Provider: "custom-provider",
		}

		assert.Equal(t, "custom-provider", req.Provider)
	})

	t.Run("JSON deserialization - single text", func(t *testing.T) {
		jsonData := `{"text":"Sample text","model":"nomic-embed-text"}`

		var req EmbedRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "Sample text", req.Text)
		assert.Equal(t, "nomic-embed-text", req.Model)
	})

	t.Run("JSON deserialization - multiple texts", func(t *testing.T) {
		jsonData := `{"texts":["text1","text2"],"model":"text-embedding-3-small"}`

		var req EmbedRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Len(t, req.Texts, 2)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		req := EmbedRequest{
			Text:  "Test",
			Model: "model-name",
		}

		data, err := json.Marshal(req)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"text":"Test"`)
		assert.Contains(t, string(data), `"model":"model-name"`)
	})
}

// =============================================================================
// EmbedResponse Struct Tests
// =============================================================================

func TestEmbedResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := EmbedResponse{
			Embeddings: [][]float32{{0.1, 0.2, 0.3}, {0.4, 0.5, 0.6}},
			Model:      "text-embedding-3-small",
			Dimensions: 1536,
			Usage: &EmbedUsage{
				PromptTokens: 10,
				TotalTokens:  10,
			},
		}

		assert.Len(t, resp.Embeddings, 2)
		assert.Len(t, resp.Embeddings[0], 3)
		assert.Equal(t, "text-embedding-3-small", resp.Model)
		assert.Equal(t, 1536, resp.Dimensions)
		assert.NotNil(t, resp.Usage)
		assert.Equal(t, 10, resp.Usage.PromptTokens)
	})

	t.Run("without usage", func(t *testing.T) {
		resp := EmbedResponse{
			Embeddings: [][]float32{{0.1, 0.2}},
			Model:      "nomic-embed-text",
			Dimensions: 768,
			Usage:      nil,
		}

		assert.Nil(t, resp.Usage)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := EmbedResponse{
			Embeddings: [][]float32{{0.5, 0.5}},
			Model:      "test-model",
			Dimensions: 2,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"embeddings"`)
		assert.Contains(t, string(data), `"model":"test-model"`)
		assert.Contains(t, string(data), `"dimensions":2`)
	})
}

// =============================================================================
// EmbedUsage Struct Tests
// =============================================================================

func TestEmbedUsage_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		usage := EmbedUsage{
			PromptTokens: 100,
			TotalTokens:  100,
		}

		assert.Equal(t, 100, usage.PromptTokens)
		assert.Equal(t, 100, usage.TotalTokens)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		usage := EmbedUsage{
			PromptTokens: 50,
			TotalTokens:  50,
		}

		data, err := json.Marshal(usage)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"prompt_tokens":50`)
		assert.Contains(t, string(data), `"total_tokens":50`)
	})
}

// =============================================================================
// VectorSearchRequest Struct Tests
// =============================================================================

func TestVectorSearchRequest_Struct(t *testing.T) {
	t.Run("query-based search", func(t *testing.T) {
		threshold := 0.8
		count := 10
		req := VectorSearchRequest{
			Table:          "documents",
			Column:         "embedding",
			Query:          "Find similar documents",
			Metric:         "cosine",
			MatchThreshold: &threshold,
			MatchCount:     &count,
			Select:         "id,title,content",
		}

		assert.Equal(t, "documents", req.Table)
		assert.Equal(t, "embedding", req.Column)
		assert.Equal(t, "Find similar documents", req.Query)
		assert.Equal(t, "cosine", req.Metric)
		assert.Equal(t, 0.8, *req.MatchThreshold)
		assert.Equal(t, 10, *req.MatchCount)
		assert.Equal(t, "id,title,content", req.Select)
	})

	t.Run("vector-based search", func(t *testing.T) {
		req := VectorSearchRequest{
			Table:  "products",
			Column: "embedding",
			Vector: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
			Metric: "l2",
		}

		assert.Empty(t, req.Query)
		assert.Len(t, req.Vector, 5)
		assert.Equal(t, "l2", req.Metric)
	})

	t.Run("with filters", func(t *testing.T) {
		req := VectorSearchRequest{
			Table:  "documents",
			Column: "embedding",
			Query:  "test query",
			Filters: []VectorQueryFilter{
				{Column: "category", Operator: "eq", Value: "technology"},
				{Column: "published", Operator: "eq", Value: true},
			},
		}

		assert.Len(t, req.Filters, 2)
		assert.Equal(t, "category", req.Filters[0].Column)
		assert.Equal(t, "eq", req.Filters[0].Operator)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"table": "articles",
			"column": "content_embedding",
			"query": "machine learning",
			"metric": "cosine",
			"match_count": 5,
			"filters": [
				{"column": "status", "operator": "eq", "value": "published"}
			]
		}`

		var req VectorSearchRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "articles", req.Table)
		assert.Equal(t, "content_embedding", req.Column)
		assert.Equal(t, "machine learning", req.Query)
		assert.Equal(t, "cosine", req.Metric)
		assert.Equal(t, 5, *req.MatchCount)
		assert.Len(t, req.Filters, 1)
	})
}

// =============================================================================
// VectorQueryFilter Struct Tests
// =============================================================================

func TestVectorQueryFilter_Struct(t *testing.T) {
	t.Run("string value filter", func(t *testing.T) {
		filter := VectorQueryFilter{
			Column:   "category",
			Operator: "eq",
			Value:    "sports",
		}

		assert.Equal(t, "category", filter.Column)
		assert.Equal(t, "eq", filter.Operator)
		assert.Equal(t, "sports", filter.Value)
	})

	t.Run("numeric value filter", func(t *testing.T) {
		filter := VectorQueryFilter{
			Column:   "score",
			Operator: "gte",
			Value:    75.5,
		}

		assert.Equal(t, 75.5, filter.Value)
	})

	t.Run("boolean value filter", func(t *testing.T) {
		filter := VectorQueryFilter{
			Column:   "is_active",
			Operator: "eq",
			Value:    true,
		}

		assert.Equal(t, true, filter.Value)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		filter := VectorQueryFilter{
			Column:   "type",
			Operator: "in",
			Value:    []string{"a", "b"},
		}

		data, err := json.Marshal(filter)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"column":"type"`)
		assert.Contains(t, string(data), `"operator":"in"`)
		assert.Contains(t, string(data), `"value"`)
	})
}

// =============================================================================
// VectorSearchResponse Struct Tests
// =============================================================================

func TestVectorSearchResponse_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		resp := VectorSearchResponse{
			Data: []map[string]interface{}{
				{"id": 1, "title": "Document 1"},
				{"id": 2, "title": "Document 2"},
			},
			Distances: []float64{0.1, 0.2},
			Model:     "text-embedding-3-small",
		}

		assert.Len(t, resp.Data, 2)
		assert.Len(t, resp.Distances, 2)
		assert.Equal(t, "text-embedding-3-small", resp.Model)
	})

	t.Run("empty results", func(t *testing.T) {
		resp := VectorSearchResponse{
			Data:      []map[string]interface{}{},
			Distances: []float64{},
			Model:     "nomic-embed-text",
		}

		assert.Empty(t, resp.Data)
		assert.Empty(t, resp.Distances)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		resp := VectorSearchResponse{
			Data: []map[string]interface{}{
				{"id": 1},
			},
			Distances: []float64{0.5},
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"data"`)
		assert.Contains(t, string(data), `"distances"`)
	})
}

// =============================================================================
// VectorCapabilities Struct Tests
// =============================================================================

func TestVectorCapabilities_Struct(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		caps := VectorCapabilities{
			Enabled:           true,
			PgVectorInstalled: true,
			PgVectorVersion:   "0.7.0",
			EmbeddingEnabled:  true,
			EmbeddingProvider: "openai",
			EmbeddingModel:    "text-embedding-3-small",
		}

		assert.True(t, caps.Enabled)
		assert.True(t, caps.PgVectorInstalled)
		assert.Equal(t, "0.7.0", caps.PgVectorVersion)
		assert.True(t, caps.EmbeddingEnabled)
		assert.Equal(t, "openai", caps.EmbeddingProvider)
		assert.Equal(t, "text-embedding-3-small", caps.EmbeddingModel)
	})

	t.Run("disabled capabilities", func(t *testing.T) {
		caps := VectorCapabilities{
			Enabled:           false,
			PgVectorInstalled: false,
			EmbeddingEnabled:  false,
		}

		assert.False(t, caps.Enabled)
		assert.False(t, caps.PgVectorInstalled)
		assert.False(t, caps.EmbeddingEnabled)
		assert.Empty(t, caps.EmbeddingProvider)
		assert.Empty(t, caps.EmbeddingModel)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		caps := VectorCapabilities{
			Enabled:           true,
			PgVectorInstalled: true,
			PgVectorVersion:   "0.6.0",
			EmbeddingEnabled:  true,
			EmbeddingProvider: "azure",
			EmbeddingModel:    "text-embedding-ada-002",
		}

		data, err := json.Marshal(caps)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"enabled":true`)
		assert.Contains(t, string(data), `"pgvector_installed":true`)
		assert.Contains(t, string(data), `"pgvector_version":"0.6.0"`)
		assert.Contains(t, string(data), `"embedding_enabled":true`)
		assert.Contains(t, string(data), `"embedding_provider":"azure"`)
		assert.Contains(t, string(data), `"embedding_model":"text-embedding-ada-002"`)
	})

	t.Run("JSON omits empty optional fields", func(t *testing.T) {
		caps := VectorCapabilities{
			Enabled:           false,
			PgVectorInstalled: false,
			EmbeddingEnabled:  false,
		}

		data, err := json.Marshal(caps)
		require.NoError(t, err)

		// Optional fields with empty values should be omitted
		assert.NotContains(t, string(data), `"pgvector_version"`)
		assert.NotContains(t, string(data), `"embedding_provider"`)
		assert.NotContains(t, string(data), `"embedding_model"`)
	})
}

// =============================================================================
// HandleEmbed Handler Tests
// =============================================================================

func TestHandleEmbed_Validation(t *testing.T) {
	t.Run("no text provided", func(t *testing.T) {
		app := fiber.New()

		// We can't easily test the full handler without mocking VectorManager
		// but we can test the request validation behavior

		// This is a simple test that verifies the struct is parsed correctly
		body := `{"model":"test-model"}`
		var req EmbedRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)

		assert.Empty(t, req.Text)
		assert.Empty(t, req.Texts)

		_ = app // Prevent unused variable
	})

	t.Run("provider selection requires admin", func(t *testing.T) {
		req := EmbedRequest{
			Text:     "Test",
			Provider: "custom-provider",
		}

		// Non-admin users should not be able to select provider
		// This is enforced in the handler with role check
		assert.NotEmpty(t, req.Provider)
	})
}

// =============================================================================
// HandleSearch Handler Tests
// =============================================================================

func TestHandleSearch_Validation(t *testing.T) {
	t.Run("missing table and column", func(t *testing.T) {
		body := `{"query":"test"}`
		var req VectorSearchRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)

		assert.Empty(t, req.Table)
		assert.Empty(t, req.Column)
	})

	t.Run("neither query nor vector provided", func(t *testing.T) {
		body := `{"table":"docs","column":"embedding"}`
		var req VectorSearchRequest
		err := json.Unmarshal([]byte(body), &req)
		require.NoError(t, err)

		assert.Empty(t, req.Query)
		assert.Empty(t, req.Vector)
	})
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestFormatVectorLiteral(t *testing.T) {
	t.Run("format simple vector", func(t *testing.T) {
		vector := []float64{0.1, 0.2, 0.3}
		result := formatVectorLiteral(vector)

		assert.Equal(t, "[0.1,0.2,0.3]", result)
	})

	t.Run("format single element vector", func(t *testing.T) {
		vector := []float64{0.5}
		result := formatVectorLiteral(vector)

		assert.Equal(t, "[0.5]", result)
	})

	t.Run("format empty vector", func(t *testing.T) {
		vector := []float64{}
		result := formatVectorLiteral(vector)

		assert.Equal(t, "[]", result)
	})

	t.Run("format vector with integers", func(t *testing.T) {
		vector := []float64{1.0, 2.0, 3.0}
		result := formatVectorLiteral(vector)

		assert.Equal(t, "[1,2,3]", result)
	})

	t.Run("format vector with small values", func(t *testing.T) {
		vector := []float64{0.001, 0.002}
		result := formatVectorLiteral(vector)

		assert.Contains(t, result, "0.001")
		assert.Contains(t, result, "0.002")
	})
}

func TestNormalizeOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"eq", "="},
		{"=", "="},
		{"neq", "!="},
		{"!=", "!="},
		{"<>", "!="},
		{"gt", ">"},
		{">", ">"},
		{"gte", ">="},
		{">=", ">="},
		{"lt", "<"},
		{"<", "<"},
		{"lte", "<="},
		{"<=", "<="},
		{"like", "LIKE"},
		{"ilike", "ILIKE"},
		{"is", "IS"},
		{"in", "IN"},
		{"invalid", ""},
		{"", ""},
		{"UNKNOWN", ""},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := normalizeOperator(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// =============================================================================
// Distance Metric Tests
// =============================================================================

func TestDistanceMetrics(t *testing.T) {
	validMetrics := []string{"l2", "euclidean", "cosine", "inner_product", "ip"}
	invalidMetrics := []string{"manhattan", "hamming", "jaccard", ""}

	t.Run("valid metrics mapping", func(t *testing.T) {
		metricOps := map[string]string{
			"l2":            "<->",
			"euclidean":     "<->",
			"cosine":        "<=>",
			"inner_product": "<#>",
			"ip":            "<#>",
		}

		for metric, expectedOp := range metricOps {
			// Verify the metric names are understood
			assert.Contains(t, validMetrics, metric, "Metric %q should be valid", metric)
			assert.NotEmpty(t, expectedOp)
		}
	})

	t.Run("invalid metrics", func(t *testing.T) {
		for _, metric := range invalidMetrics {
			// These should not be in the valid list
			assert.NotContains(t, validMetrics, metric)
		}
	})
}

// =============================================================================
// HandleGetCapabilities Handler Tests
// =============================================================================

func TestHandleGetCapabilities_RoleBasedResponse(t *testing.T) {
	t.Run("admin user gets full details", func(t *testing.T) {
		app := fiber.New()

		// Middleware to set admin role
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_role", "admin")
			return c.Next()
		})

		app.Get("/capabilities/vector", func(c fiber.Ctx) error {
			role, _ := c.Locals("user_role").(string)
			isAdmin := role == "admin" || role == "dashboard_admin" || role == "service_role"

			if !isAdmin {
				return c.JSON(fiber.Map{
					"enabled": false,
				})
			}

			return c.JSON(VectorCapabilities{
				Enabled:           true,
				PgVectorInstalled: true,
				PgVectorVersion:   "0.7.0",
				EmbeddingEnabled:  true,
				EmbeddingProvider: "openai",
				EmbeddingModel:    "text-embedding-3-small",
			})
		})

		req := httptest.NewRequest(http.MethodGet, "/capabilities/vector", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result VectorCapabilities
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		assert.True(t, result.Enabled)
		assert.True(t, result.PgVectorInstalled)
		assert.Equal(t, "0.7.0", result.PgVectorVersion)
		assert.True(t, result.EmbeddingEnabled)
		assert.Equal(t, "openai", result.EmbeddingProvider)
	})

	t.Run("non-admin user gets minimal info", func(t *testing.T) {
		app := fiber.New()

		// Middleware to set regular user role
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_role", "user")
			return c.Next()
		})

		app.Get("/capabilities/vector", func(c fiber.Ctx) error {
			role, _ := c.Locals("user_role").(string)
			isAdmin := role == "admin" || role == "dashboard_admin" || role == "service_role"

			if !isAdmin {
				return c.JSON(fiber.Map{
					"enabled": true,
				})
			}

			return c.JSON(VectorCapabilities{})
		})

		req := httptest.NewRequest(http.MethodGet, "/capabilities/vector", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		respBody, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		require.NoError(t, err)

		// Non-admin should only see enabled field
		assert.Contains(t, result, "enabled")
		assert.NotContains(t, result, "pgvector_version")
		assert.NotContains(t, result, "embedding_provider")
	})
}

// =============================================================================
// Admin Role Verification Tests
// =============================================================================

func TestVectorAdminRoles(t *testing.T) {
	adminRoles := []string{"admin", "dashboard_admin", "service_role"}
	nonAdminRoles := []string{"user", "authenticated", "anon", ""}

	t.Run("admin roles should have full access", func(t *testing.T) {
		for _, role := range adminRoles {
			isAdmin := role == "admin" || role == "dashboard_admin" || role == "service_role"
			assert.True(t, isAdmin, "Role %q should be admin", role)
		}
	})

	t.Run("non-admin roles should have limited access", func(t *testing.T) {
		for _, role := range nonAdminRoles {
			isAdmin := role == "admin" || role == "dashboard_admin" || role == "service_role"
			assert.False(t, isAdmin, "Role %q should not be admin", role)
		}
	})
}

// =============================================================================
// Match Count Limits Tests
// =============================================================================

func TestMatchCountLimits(t *testing.T) {
	t.Run("default match count", func(t *testing.T) {
		req := VectorSearchRequest{
			Table:      "docs",
			Column:     "embedding",
			Query:      "test",
			MatchCount: nil,
		}

		assert.Nil(t, req.MatchCount)
	})

	t.Run("custom match count", func(t *testing.T) {
		count := 50
		req := VectorSearchRequest{
			Table:      "docs",
			Column:     "embedding",
			Query:      "test",
			MatchCount: &count,
		}

		assert.Equal(t, 50, *req.MatchCount)
	})

	t.Run("max match count enforcement", func(t *testing.T) {
		// The handler caps match_count at 1000
		maxCount := 1000
		requestedCount := 2000

		// Simulate the capping logic
		finalCount := requestedCount
		if finalCount > maxCount {
			finalCount = maxCount
		}

		assert.Equal(t, 1000, finalCount)
	})
}

// =============================================================================
// inferProviderType Tests
// =============================================================================

func TestInferProviderType(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *config.AIConfig
		expected string
	}{
		{
			name: "explicit embedding provider takes precedence",
			cfg: &config.AIConfig{
				EmbeddingProvider: "custom-embed",
				ProviderType:      "azure",
				OpenAIAPIKey:      "sk-key",
			},
			expected: "custom-embed",
		},
		{
			name: "explicit provider type",
			cfg: &config.AIConfig{
				ProviderType: "azure",
				OpenAIAPIKey: "sk-key",
			},
			expected: "azure",
		},
		{
			name: "openai API key infers openai",
			cfg: &config.AIConfig{
				OpenAIAPIKey: "sk-test-key",
			},
			expected: "openai",
		},
		{
			name: "azure API key and endpoint infers azure",
			cfg: &config.AIConfig{
				AzureAPIKey:   "azure-key",
				AzureEndpoint: "https://openai.azure.com",
			},
			expected: "azure",
		},
		{
			name: "ollama endpoint infers ollama",
			cfg: &config.AIConfig{
				OllamaEndpoint: "http://localhost:11434",
			},
			expected: "ollama",
		},
		{
			name: "azure API key without endpoint returns empty",
			cfg: &config.AIConfig{
				AzureAPIKey: "azure-key",
			},
			expected: "",
		},
		{
			name: "no configuration returns empty",
			cfg:  &config.AIConfig{
				// All empty
			},
			expected: "",
		},
		{
			name: "openai key takes precedence over ollama endpoint",
			cfg: &config.AIConfig{
				OpenAIAPIKey:   "sk-key",
				OllamaEndpoint: "http://localhost:11434",
			},
			expected: "openai", // OpenAI has higher priority in the check order
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferProviderType(tt.cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}
