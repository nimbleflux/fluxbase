package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryClassifier_ClassifyQuery(t *testing.T) {
	classifier := NewQueryClassifier()

	tests := []struct {
		name           string
		query          string
		expectedType   QueryClassification
		expectedInType []QueryClassification // Alternative acceptable types
	}{
		// Structured queries
		{
			name:         "time based query",
			query:        "What restaurants did I visit last week?",
			expectedType: QueryTypeStructured,
		},
		{
			name:         "count query",
			query:        "How many times did I go to restaurants this month?",
			expectedType: QueryTypeStructured,
		},
		{
			name:         "filter query",
			query:        "Show me all visits to restaurants in Berlin",
			expectedType: QueryTypeStructured,
		},
		{
			name:         "ordering query",
			query:        "List my most recent restaurant visits",
			expectedType: QueryTypeStructured,
		},
		{
			name:         "specific date query",
			query:        "Where did I eat on January 15, 2024?",
			expectedType: QueryTypeStructured,
		},

		// Semantic queries
		{
			name:         "conceptual question",
			query:        "What is Italian cuisine?",
			expectedType: QueryTypeSemantic,
		},
		{
			name:         "explanation request",
			query:        "Tell me about Japanese food culture",
			expectedType: QueryTypeSemantic,
		},
		{
			name:           "description request",
			query:          "Explain the difference between pasta types",
			expectedInType: []QueryClassification{QueryTypeSemantic, QueryTypeHybrid}, // "between" triggers structured indicator
		},

		// Hybrid queries
		{
			name:           "specific with context",
			query:          "What Italian restaurants did I visit last month?",
			expectedInType: []QueryClassification{QueryTypeHybrid, QueryTypeStructured},
		},
		{
			name:           "conceptual with filter",
			query:          "Show me information about French cuisine restaurants I've been to",
			expectedInType: []QueryClassification{QueryTypeHybrid, QueryTypeStructured},
		},

		// Unknown
		{
			name:         "greeting",
			query:        "Hello!",
			expectedType: QueryTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.ClassifyQuery(tt.query)

			if tt.expectedInType != nil {
				found := false
				for _, expected := range tt.expectedInType {
					if result == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "expected one of %v, got %s", tt.expectedInType, result)
			} else {
				assert.Equal(t, tt.expectedType, result)
			}
		})
	}
}

func TestQueryClassifier_GetToolRecommendation(t *testing.T) {
	classifier := NewQueryClassifier()

	tests := []struct {
		classification    QueryClassification
		expectedTools     []string
		shouldContainTool string
	}{
		{
			classification:    QueryTypeStructured,
			shouldContainTool: "query_table",
		},
		{
			classification:    QueryTypeSemantic,
			shouldContainTool: "search_vectors",
		},
		{
			classification:    QueryTypeHybrid,
			shouldContainTool: "search_vectors",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.classification), func(t *testing.T) {
			tools := classifier.GetToolRecommendation(tt.classification)
			assert.Contains(t, tools, tt.shouldContainTool)
		})
	}
}

func TestQueryClassifier_GetStrategyDescription(t *testing.T) {
	classifier := NewQueryClassifier()

	t.Run("all classifications have descriptions", func(t *testing.T) {
		for _, classification := range []QueryClassification{
			QueryTypeStructured,
			QueryTypeSemantic,
			QueryTypeHybrid,
			QueryTypeUnknown,
		} {
			desc := classifier.GetStrategyDescription(classification)
			assert.NotEmpty(t, desc, "classification %s should have description", classification)
		}
	})
}
