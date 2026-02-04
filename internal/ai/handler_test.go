package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with nil dependencies", func(t *testing.T) {
		handler := NewHandler(nil, nil, nil, nil)
		assert.NotNil(t, handler)
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.loader)
		assert.Nil(t, handler.config)
		assert.Nil(t, handler.vectorManager)
	})
}

func TestNormalizeConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		expected map[string]string
	}{
		{
			name:     "nil config",
			input:    nil,
			expected: map[string]string{},
		},
		{
			name:     "empty config",
			input:    map[string]any{},
			expected: map[string]string{},
		},
		{
			name: "string values",
			input: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "numeric values",
			input: map[string]any{
				"port":    8080,
				"timeout": 30.5,
			},
			expected: map[string]string{
				"port":    "8080",
				"timeout": "30.5",
			},
		},
		{
			name: "boolean values",
			input: map[string]any{
				"enabled": true,
				"debug":   false,
			},
			expected: map[string]string{
				"enabled": "true",
				"debug":   "false",
			},
		},
		{
			name: "nil value skipped",
			input: map[string]any{
				"key1": "value1",
				"key2": nil,
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "empty string skipped",
			input: map[string]any{
				"key1": "value1",
				"key2": "",
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "undefined string skipped",
			input: map[string]any{
				"key1": "value1",
				"key2": "undefined",
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "null string skipped",
			input: map[string]any{
				"key1": "value1",
				"key2": "null",
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSyncChatbotsRequest_Struct(t *testing.T) {
	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"namespace": "prod",
			"chatbots": [
				{"name": "bot1", "code": "code1"},
				{"name": "bot2", "code": "code2"}
			],
			"options": {
				"delete_missing": true,
				"dry_run": false
			}
		}`

		var req SyncChatbotsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "prod", req.Namespace)
		assert.Len(t, req.Chatbots, 2)
		assert.Equal(t, "bot1", req.Chatbots[0].Name)
		assert.Equal(t, "code1", req.Chatbots[0].Code)
		assert.True(t, req.Options.DeleteMissing)
		assert.False(t, req.Options.DryRun)
	})

	t.Run("empty chatbots", func(t *testing.T) {
		jsonData := `{"namespace": "test"}`

		var req SyncChatbotsRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "test", req.Namespace)
		assert.Empty(t, req.Chatbots)
	})
}

func TestToggleChatbotRequest_Struct(t *testing.T) {
	t.Run("enabled true", func(t *testing.T) {
		jsonData := `{"enabled": true}`

		var req ToggleChatbotRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.True(t, req.Enabled)
	})

	t.Run("enabled false", func(t *testing.T) {
		jsonData := `{"enabled": false}`

		var req ToggleChatbotRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.False(t, req.Enabled)
	})
}

func TestUpdateChatbotRequest_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		desc := "New description"
		enabled := true
		maxTokens := 2000
		temp := 0.8
		providerID := "prov-123"
		persist := true
		ttl := 24
		maxTurns := 10
		rateLimit := 60
		dailyLimit := 1000
		tokenBudget := 50000
		allowUnauth := false
		isPublic := true

		req := UpdateChatbotRequest{
			Description:          &desc,
			Enabled:              &enabled,
			MaxTokens:            &maxTokens,
			Temperature:          &temp,
			ProviderID:           &providerID,
			PersistConversations: &persist,
			ConversationTTLHours: &ttl,
			MaxConversationTurns: &maxTurns,
			RateLimitPerMinute:   &rateLimit,
			DailyRequestLimit:    &dailyLimit,
			DailyTokenBudget:     &tokenBudget,
			AllowUnauthenticated: &allowUnauth,
			IsPublic:             &isPublic,
		}

		assert.Equal(t, "New description", *req.Description)
		assert.True(t, *req.Enabled)
		assert.Equal(t, 2000, *req.MaxTokens)
		assert.Equal(t, 0.8, *req.Temperature)
		assert.True(t, *req.IsPublic)
	})

	t.Run("partial update", func(t *testing.T) {
		maxTokens := 1000

		req := UpdateChatbotRequest{
			MaxTokens: &maxTokens,
		}

		assert.Nil(t, req.Description)
		assert.Nil(t, req.Enabled)
		assert.Equal(t, 1000, *req.MaxTokens)
	})
}

func TestCreateProviderRequest_Struct(t *testing.T) {
	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{
			"name": "openai-prod",
			"display_name": "OpenAI Production",
			"provider_type": "openai",
			"is_default": true,
			"embedding_model": "text-embedding-3-small",
			"config": {
				"api_key": "sk-xxx",
				"model": "gpt-4"
			},
			"enabled": true
		}`

		var req CreateProviderRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)

		assert.Equal(t, "openai-prod", req.Name)
		assert.Equal(t, "OpenAI Production", req.DisplayName)
		assert.Equal(t, "openai", req.ProviderType)
		assert.True(t, req.IsDefault)
		assert.NotNil(t, req.EmbeddingModel)
		assert.Equal(t, "text-embedding-3-small", *req.EmbeddingModel)
		assert.Equal(t, "sk-xxx", req.Config["api_key"])
		assert.True(t, req.Enabled)
	})
}

func TestUpdateProviderRequest_Struct(t *testing.T) {
	t.Run("partial update", func(t *testing.T) {
		displayName := "Updated Name"
		enabled := false

		req := UpdateProviderRequest{
			DisplayName: &displayName,
			Enabled:     &enabled,
		}

		assert.Equal(t, "Updated Name", *req.DisplayName)
		assert.False(t, *req.Enabled)
		assert.Nil(t, req.Config)
		assert.Nil(t, req.EmbeddingModel)
	})
}

func TestLookupChatbotByNameResponse_Struct(t *testing.T) {
	t.Run("single match", func(t *testing.T) {
		summary := ChatbotSummary{
			ID:   "chatbot-123",
			Name: "test-bot",
		}

		resp := LookupChatbotByNameResponse{
			Chatbot:   &summary,
			Ambiguous: false,
		}

		assert.NotNil(t, resp.Chatbot)
		assert.Equal(t, "chatbot-123", resp.Chatbot.ID)
		assert.False(t, resp.Ambiguous)
		assert.Empty(t, resp.Namespaces)
	})

	t.Run("ambiguous result", func(t *testing.T) {
		resp := LookupChatbotByNameResponse{
			Ambiguous:  true,
			Namespaces: []string{"prod", "staging"},
			Error:      "Multiple matches found",
		}

		assert.Nil(t, resp.Chatbot)
		assert.True(t, resp.Ambiguous)
		assert.Len(t, resp.Namespaces, 2)
		assert.NotEmpty(t, resp.Error)
	})

	t.Run("not found", func(t *testing.T) {
		resp := LookupChatbotByNameResponse{
			Ambiguous: false,
			Error:     "Chatbot not found",
		}

		assert.Nil(t, resp.Chatbot)
		assert.False(t, resp.Ambiguous)
		assert.Equal(t, "Chatbot not found", resp.Error)
	})
}

func TestChatbotMetric_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		metric := ChatbotMetric{
			ChatbotID:   "chatbot-123",
			ChatbotName: "test-bot",
			Requests:    1000,
			Tokens:      50000,
			ErrorCount:  5,
		}

		assert.Equal(t, "chatbot-123", metric.ChatbotID)
		assert.Equal(t, "test-bot", metric.ChatbotName)
		assert.Equal(t, int64(1000), metric.Requests)
		assert.Equal(t, int64(50000), metric.Tokens)
		assert.Equal(t, int64(5), metric.ErrorCount)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		metric := ChatbotMetric{
			ChatbotID:   "cb-1",
			ChatbotName: "bot",
			Requests:    100,
		}

		data, err := json.Marshal(metric)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"chatbot_id":"cb-1"`)
		assert.Contains(t, string(data), `"requests":100`)
	})
}

func TestProviderMetric_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		metric := ProviderMetric{
			ProviderID:   "prov-123",
			ProviderName: "openai-prod",
			Requests:     5000,
			AvgLatencyMS: 125.5,
		}

		assert.Equal(t, "prov-123", metric.ProviderID)
		assert.Equal(t, "openai-prod", metric.ProviderName)
		assert.Equal(t, int64(5000), metric.Requests)
		assert.Equal(t, 125.5, metric.AvgLatencyMS)
	})
}

func TestAIMetrics_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		metrics := AIMetrics{
			TotalRequests:         10000,
			TotalTokens:           500000,
			TotalPromptTokens:     300000,
			TotalCompletionTokens: 200000,
			ActiveConversations:   50,
			TotalConversations:    1000,
			ChatbotStats:          []ChatbotMetric{{ChatbotID: "cb-1"}},
			ProviderStats:         []ProviderMetric{{ProviderID: "prov-1"}},
			ErrorRate:             0.5,
			AvgResponseTimeMS:     150.0,
		}

		assert.Equal(t, int64(10000), metrics.TotalRequests)
		assert.Equal(t, int64(500000), metrics.TotalTokens)
		assert.Equal(t, 50, metrics.ActiveConversations)
		assert.Len(t, metrics.ChatbotStats, 1)
		assert.Len(t, metrics.ProviderStats, 1)
		assert.Equal(t, 0.5, metrics.ErrorRate)
	})
}

func TestConversationSummary_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		userID := "user-123"
		userEmail := "user@example.com"
		sessionID := "sess-456"
		title := "Chat about AI"
		lastMsgAt := time.Now()

		summary := ConversationSummary{
			ID:                    "conv-123",
			ChatbotID:             "chatbot-456",
			ChatbotName:           "test-bot",
			UserID:                &userID,
			UserEmail:             &userEmail,
			SessionID:             &sessionID,
			Title:                 &title,
			Status:                "active",
			TurnCount:             10,
			TotalPromptTokens:     500,
			TotalCompletionTokens: 300,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			LastMessageAt:         &lastMsgAt,
		}

		assert.Equal(t, "conv-123", summary.ID)
		assert.Equal(t, "test-bot", summary.ChatbotName)
		assert.Equal(t, "user@example.com", *summary.UserEmail)
		assert.Equal(t, "active", summary.Status)
		assert.Equal(t, 10, summary.TurnCount)
	})
}

func TestMessageDetail_Struct(t *testing.T) {
	t.Run("user message", func(t *testing.T) {
		msg := MessageDetail{
			ID:             "msg-123",
			ConversationID: "conv-456",
			Role:           "user",
			Content:        "Hello, bot!",
			CreatedAt:      time.Now(),
			SequenceNumber: 1,
		}

		assert.Equal(t, "msg-123", msg.ID)
		assert.Equal(t, "user", msg.Role)
		assert.Nil(t, msg.ExecutedSQL)
	})

	t.Run("assistant message with SQL", func(t *testing.T) {
		sql := "SELECT * FROM users"
		summary := "Found 10 users"
		rowCount := 10
		durationMs := 25

		msg := MessageDetail{
			ID:               "msg-456",
			ConversationID:   "conv-456",
			Role:             "assistant",
			Content:          "Here are the users...",
			ExecutedSQL:      &sql,
			SQLResultSummary: &summary,
			SQLRowCount:      &rowCount,
			SQLDurationMS:    &durationMs,
			SequenceNumber:   2,
		}

		assert.Equal(t, "assistant", msg.Role)
		assert.Equal(t, "SELECT * FROM users", *msg.ExecutedSQL)
		assert.Equal(t, 10, *msg.SQLRowCount)
	})

	t.Run("tool message", func(t *testing.T) {
		toolCallID := "call_123"
		toolName := "search"

		msg := MessageDetail{
			ID:             "msg-789",
			Role:           "tool",
			Content:        "Tool result",
			ToolCallID:     &toolCallID,
			ToolName:       &toolName,
			SequenceNumber: 3,
		}

		assert.Equal(t, "tool", msg.Role)
		assert.Equal(t, "call_123", *msg.ToolCallID)
		assert.Equal(t, "search", *msg.ToolName)
	})
}

func TestAuditLogEntry_Struct(t *testing.T) {
	t.Run("successful query", func(t *testing.T) {
		chatbotID := "chatbot-123"
		userID := "user-456"
		sanitizedSQL := "SELECT * FROM users"
		validPassed := true
		success := true
		rows := 10
		durationMs := 25
		ipAddr := "192.168.1.1"
		userAgent := "Mozilla/5.0"

		entry := AuditLogEntry{
			ID:                  "audit-123",
			ChatbotID:           &chatbotID,
			UserID:              &userID,
			GeneratedSQL:        "SELECT * FROM users WHERE id = 1",
			SanitizedSQL:        &sanitizedSQL,
			Executed:            true,
			ValidationPassed:    &validPassed,
			ValidationErrors:    []string{},
			Success:             &success,
			RowsReturned:        &rows,
			ExecutionDurationMS: &durationMs,
			TablesAccessed:      []string{"users"},
			OperationsUsed:      []string{"SELECT"},
			IPAddress:           &ipAddr,
			UserAgent:           &userAgent,
			CreatedAt:           time.Now(),
		}

		assert.Equal(t, "audit-123", entry.ID)
		assert.True(t, entry.Executed)
		assert.True(t, *entry.ValidationPassed)
		assert.True(t, *entry.Success)
		assert.Equal(t, 10, *entry.RowsReturned)
	})

	t.Run("failed query", func(t *testing.T) {
		validPassed := false
		success := false
		errorMsg := "Permission denied"

		entry := AuditLogEntry{
			ID:               "audit-456",
			GeneratedSQL:     "DELETE FROM users",
			Executed:         false,
			ValidationPassed: &validPassed,
			ValidationErrors: []string{"DELETE not allowed"},
			Success:          &success,
			ErrorMessage:     &errorMsg,
		}

		assert.False(t, entry.Executed)
		assert.False(t, *entry.ValidationPassed)
		assert.False(t, *entry.Success)
		assert.Equal(t, "Permission denied", *entry.ErrorMessage)
	})
}

func TestUpdateConversationTitleRequest_Struct(t *testing.T) {
	t.Run("valid title", func(t *testing.T) {
		jsonData := `{"title": "New Conversation Title"}`

		var req UpdateConversationTitleRequest
		err := json.Unmarshal([]byte(jsonData), &req)
		require.NoError(t, err)
		assert.Equal(t, "New Conversation Title", req.Title)
	})
}

// =============================================================================
// ValidateConfig Tests
// =============================================================================

func TestHandler_ValidateConfig(t *testing.T) {
	t.Run("nil config - no panic", func(t *testing.T) {
		handler := NewHandler(nil, nil, nil, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("empty provider type - no warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("ollama with model configured", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "ollama",
			OllamaModel:  "llama2",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("ollama without model - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "ollama",
			OllamaModel:  "",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("openai with api key configured", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "openai",
			OpenAIAPIKey: "sk-test-key",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("openai without api key - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "openai",
			OpenAIAPIKey: "",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("azure fully configured", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType:        "azure",
			AzureAPIKey:         "azure-key",
			AzureEndpoint:       "https://example.openai.azure.com",
			AzureDeploymentName: "gpt-4-deployment",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("azure missing api key - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType:        "azure",
			AzureAPIKey:         "",
			AzureEndpoint:       "https://example.openai.azure.com",
			AzureDeploymentName: "gpt-4-deployment",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("azure missing endpoint - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType:        "azure",
			AzureAPIKey:         "azure-key",
			AzureEndpoint:       "",
			AzureDeploymentName: "gpt-4-deployment",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("azure missing deployment name - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType:        "azure",
			AzureAPIKey:         "azure-key",
			AzureEndpoint:       "https://example.openai.azure.com",
			AzureDeploymentName: "",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("azure missing all fields - logs warning", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType:        "azure",
			AzureAPIKey:         "",
			AzureEndpoint:       "",
			AzureDeploymentName: "",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})

	t.Run("unknown provider type - no crash", func(t *testing.T) {
		handler := NewHandler(nil, nil, &config.AIConfig{
			ProviderType: "unknown",
		}, nil)
		assert.NotPanics(t, func() {
			handler.ValidateConfig()
		})
	})
}

// =============================================================================
// VectorManagerInterface Tests
// =============================================================================

// MockVectorManager implements VectorManagerInterface for testing
type MockVectorManager struct {
	refreshCalled bool
	refreshError  error
}

func (m *MockVectorManager) RefreshFromDatabase(ctx context.Context) error {
	m.refreshCalled = true
	return m.refreshError
}

func TestVectorManagerInterface(t *testing.T) {
	t.Run("mock implements interface", func(t *testing.T) {
		var _ VectorManagerInterface = (*MockVectorManager)(nil)
	})

	t.Run("handler accepts vector manager", func(t *testing.T) {
		mockVM := &MockVectorManager{}
		handler := NewHandler(nil, nil, nil, mockVM)
		assert.NotNil(t, handler)
		assert.Equal(t, mockVM, handler.vectorManager)
	})
}

// =============================================================================
// Handler Field Tests
// =============================================================================

func TestHandler_Fields(t *testing.T) {
	t.Run("all fields accessible", func(t *testing.T) {
		handler := &Handler{
			storage:       nil,
			loader:        nil,
			config:        nil,
			vectorManager: nil,
		}
		assert.Nil(t, handler.storage)
		assert.Nil(t, handler.loader)
		assert.Nil(t, handler.config)
		assert.Nil(t, handler.vectorManager)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNormalizeConfig(b *testing.B) {
	input := map[string]any{
		"api_key":   "sk-xxx",
		"model":     "gpt-4",
		"endpoint":  "https://api.openai.com",
		"timeout":   30,
		"enabled":   true,
		"nil_value": nil,
		"empty":     "",
		"undefined": "undefined",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeConfig(input)
	}
}

func BenchmarkNormalizeConfig_Empty(b *testing.B) {
	input := map[string]any{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeConfig(input)
	}
}

func BenchmarkNormalizeConfig_Large(b *testing.B) {
	input := make(map[string]any, 100)
	for i := 0; i < 100; i++ {
		input[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeConfig(input)
	}
}
