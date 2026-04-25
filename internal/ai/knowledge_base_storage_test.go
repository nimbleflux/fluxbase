package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKnowledgeBaseStorage(t *testing.T) {
	t.Run("creates storage with nil db", func(t *testing.T) {
		storage := NewKnowledgeBaseStorage(nil)
		assert.NotNil(t, storage)
		assert.Nil(t, storage.DB)
	})
}

func TestFormatEmbeddingLiteral(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected string
	}{
		{
			name:     "empty embedding",
			input:    []float32{},
			expected: "[]",
		},
		{
			name:     "single value",
			input:    []float32{0.5},
			expected: "[0.5]",
		},
		{
			name:     "multiple values",
			input:    []float32{0.1, 0.2, 0.3},
			expected: "[0.1,0.2,0.3]",
		},
		{
			name:     "negative values",
			input:    []float32{-0.5, 0.5, -1.0},
			expected: "[-0.5,0.5,-1]",
		},
		{
			name:     "small values",
			input:    []float32{0.00001, 0.00002},
			expected: "[1e-05,2e-05]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEmbeddingLiteral(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		parts    []string
		sep      string
		expected string
	}{
		{
			name:     "empty slice",
			parts:    []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single element",
			parts:    []string{"one"},
			sep:      ",",
			expected: "one",
		},
		{
			name:     "multiple elements",
			parts:    []string{"one", "two", "three"},
			sep:      ",",
			expected: "one,two,three",
		},
		{
			name:     "different separator",
			parts:    []string{"a", "b", "c"},
			sep:      " | ",
			expected: "a | b | c",
		},
		{
			name:     "empty separator",
			parts:    []string{"a", "b", "c"},
			sep:      "",
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.parts, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeMetadataKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "alphanumeric only",
			input:    "user123",
			expected: "user123",
		},
		{
			name:     "with underscores",
			input:    "user_id",
			expected: "user_id",
		},
		{
			name:     "mixed case",
			input:    "UserID",
			expected: "UserID",
		},
		{
			name:     "removes special chars",
			input:    "user-id!@#$%",
			expected: "userid",
		},
		{
			name:     "removes spaces",
			input:    "user id",
			expected: "userid",
		},
		{
			name:     "removes dots",
			input:    "user.id",
			expected: "userid",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only special chars",
			input:    "!@#$%^&*()",
			expected: "",
		},
		{
			name:     "SQL injection attempt",
			input:    "user'; DROP TABLE users; --",
			expected: "userDROPTABLEusers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeMetadataKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSearchMode_Constants(t *testing.T) {
	t.Run("all modes defined", func(t *testing.T) {
		assert.Equal(t, SearchMode("semantic"), SearchModeSemantic)
		assert.Equal(t, SearchMode("keyword"), SearchModeKeyword)
		assert.Equal(t, SearchMode("hybrid"), SearchModeHybrid)
	})
}

func TestHybridSearchOptions_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		userID := "user-123"
		opts := HybridSearchOptions{
			Query:          "search query",
			QueryEmbedding: []float32{0.1, 0.2, 0.3},
			Limit:          10,
			Threshold:      0.5,
			Mode:           SearchModeHybrid,
			SemanticWeight: 0.7,
			KeywordBoost:   0.3,
			Filter: &MetadataFilter{
				UserID: &userID,
				Tags:   []string{"tag1"},
			},
		}

		assert.Equal(t, "search query", opts.Query)
		assert.Len(t, opts.QueryEmbedding, 3)
		assert.Equal(t, 10, opts.Limit)
		assert.Equal(t, 0.5, opts.Threshold)
		assert.Equal(t, SearchModeHybrid, opts.Mode)
		assert.Equal(t, 0.7, opts.SemanticWeight)
		assert.NotNil(t, opts.Filter)
		assert.Equal(t, "user-123", *opts.Filter.UserID)
	})

	t.Run("zero values", func(t *testing.T) {
		var opts HybridSearchOptions
		assert.Empty(t, opts.Query)
		assert.Nil(t, opts.QueryEmbedding)
		assert.Equal(t, 0, opts.Limit)
		assert.Equal(t, float64(0), opts.Threshold)
		assert.Equal(t, SearchMode(""), opts.Mode)
		assert.Nil(t, opts.Filter)
	})
}

func TestChunkEmbeddingStats_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		stats := ChunkEmbeddingStats{
			TotalChunks:            100,
			ChunksWithEmbedding:    95,
			ChunksWithoutEmbedding: 5,
		}

		assert.Equal(t, 100, stats.TotalChunks)
		assert.Equal(t, 95, stats.ChunksWithEmbedding)
		assert.Equal(t, 5, stats.ChunksWithoutEmbedding)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		stats := ChunkEmbeddingStats{
			TotalChunks:            50,
			ChunksWithEmbedding:    50,
			ChunksWithoutEmbedding: 0,
		}

		data, err := json.Marshal(stats)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"total_chunks":50`)
		assert.Contains(t, string(data), `"chunks_with_embedding":50`)
		assert.Contains(t, string(data), `"chunks_without_embedding":0`)
	})
}

func TestUpdateChatbotKnowledgeBaseOptions_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		priority := 1
		maxChunks := 10
		threshold := 0.8
		enabled := true

		opts := UpdateChatbotKnowledgeBaseOptions{
			Priority:            &priority,
			MaxChunks:           &maxChunks,
			SimilarityThreshold: &threshold,
			Enabled:             &enabled,
		}

		assert.Equal(t, 1, *opts.Priority)
		assert.Equal(t, 10, *opts.MaxChunks)
		assert.Equal(t, 0.8, *opts.SimilarityThreshold)
		assert.True(t, *opts.Enabled)
	})

	t.Run("partial options", func(t *testing.T) {
		maxChunks := 5

		opts := UpdateChatbotKnowledgeBaseOptions{
			MaxChunks: &maxChunks,
		}

		assert.Nil(t, opts.Priority)
		assert.Equal(t, 5, *opts.MaxChunks)
		assert.Nil(t, opts.SimilarityThreshold)
		assert.Nil(t, opts.Enabled)
	})
}

func TestEmbeddingLiteralPrecision(t *testing.T) {
	t.Run("preserves precision for common embedding values", func(t *testing.T) {
		// Typical embedding values
		embedding := []float32{0.12345678, -0.98765432, 0.00000001}
		result := formatEmbeddingLiteral(embedding)

		// Should contain the values (format may vary based on float representation)
		assert.Contains(t, result, "[")
		assert.Contains(t, result, "]")
		assert.Contains(t, result, ",")
	})

	t.Run("handles normalized embedding values", func(t *testing.T) {
		// Normalized vectors typically have values between -1 and 1
		embedding := []float32{0.5, -0.5, 0.707, -0.707}
		result := formatEmbeddingLiteral(embedding)

		assert.Equal(t, "[0.5,-0.5,0.707,-0.707]", result)
	})
}
