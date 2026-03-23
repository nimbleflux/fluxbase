package ai

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

func TestNewStorage(t *testing.T) {
	t.Run("creates storage with nil db", func(t *testing.T) {
		storage := NewStorage(nil)
		assert.NotNil(t, storage)
		assert.Nil(t, storage.db)
		assert.Nil(t, storage.config)
	})
}

func TestStorage_SetConfig(t *testing.T) {
	t.Run("sets config", func(t *testing.T) {
		storage := NewStorage(nil)
		cfg := &config.AIConfig{
			ProviderType: "openai",
		}
		storage.SetConfig(cfg)
		assert.Equal(t, cfg, storage.config)
	})

	t.Run("can set nil config", func(t *testing.T) {
		storage := NewStorage(nil)
		storage.SetConfig(nil)
		assert.Nil(t, storage.config)
	})
}

func TestProviderRecord_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		useForEmbeddings := true
		embeddingModel := "text-embedding-3-small"
		createdBy := "admin"

		provider := ProviderRecord{
			ID:               "prov-123",
			Name:             "openai-prod",
			DisplayName:      "OpenAI Production",
			ProviderType:     "openai",
			IsDefault:        true,
			UseForEmbeddings: &useForEmbeddings,
			EmbeddingModel:   &embeddingModel,
			Config: map[string]string{
				"api_key": "sk-xxx",
				"model":   "gpt-4",
			},
			Enabled:   true,
			ReadOnly:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: &createdBy,
		}

		assert.Equal(t, "prov-123", provider.ID)
		assert.Equal(t, "openai-prod", provider.Name)
		assert.Equal(t, "OpenAI Production", provider.DisplayName)
		assert.Equal(t, "openai", provider.ProviderType)
		assert.True(t, provider.IsDefault)
		assert.True(t, *provider.UseForEmbeddings)
		assert.Equal(t, "text-embedding-3-small", *provider.EmbeddingModel)
		assert.Equal(t, "sk-xxx", provider.Config["api_key"])
		assert.True(t, provider.Enabled)
		assert.False(t, provider.ReadOnly)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		provider := ProviderRecord{
			ID:           "prov-456",
			Name:         "ollama-local",
			DisplayName:  "Local Ollama",
			ProviderType: "ollama",
			IsDefault:    false,
			Config: map[string]string{
				"endpoint": "http://localhost:11434",
				"model":    "llama2",
			},
			Enabled:   true,
			ReadOnly:  true,
			CreatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		}

		data, err := json.Marshal(provider)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"prov-456"`)
		assert.Contains(t, string(data), `"provider_type":"ollama"`)
		assert.Contains(t, string(data), `"read_only":true`)
	})

	t.Run("zero value provider", func(t *testing.T) {
		var provider ProviderRecord
		assert.Empty(t, provider.ID)
		assert.Empty(t, provider.Name)
		assert.False(t, provider.IsDefault)
		assert.Nil(t, provider.UseForEmbeddings)
		assert.Nil(t, provider.Config)
		assert.False(t, provider.Enabled)
	})
}

func TestUserConversationSummary_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		title := "Weather Query"

		summary := UserConversationSummary{
			ID:           "conv-123",
			ChatbotName:  "weather-bot",
			Namespace:    "default",
			Title:        &title,
			Preview:      "What's the weather like today?",
			MessageCount: 5,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Equal(t, "conv-123", summary.ID)
		assert.Equal(t, "weather-bot", summary.ChatbotName)
		assert.Equal(t, "default", summary.Namespace)
		assert.Equal(t, "Weather Query", *summary.Title)
		assert.Equal(t, 5, summary.MessageCount)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		summary := UserConversationSummary{
			ID:           "conv-456",
			ChatbotName:  "assistant",
			Namespace:    "prod",
			Preview:      "Hello!",
			MessageCount: 2,
			CreatedAt:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		}

		data, err := json.Marshal(summary)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"id":"conv-456"`)
		assert.Contains(t, string(data), `"chatbot":"assistant"`)
		assert.Contains(t, string(data), `"message_count":2`)
	})
}

func TestUserMessageDetail_Struct(t *testing.T) {
	t.Run("user message", func(t *testing.T) {
		msg := UserMessageDetail{
			ID:        "msg-123",
			Role:      "user",
			Content:   "What is the total sales?",
			Timestamp: time.Now(),
		}

		assert.Equal(t, "msg-123", msg.ID)
		assert.Equal(t, "user", msg.Role)
		assert.Empty(t, msg.QueryResults)
		assert.Nil(t, msg.Usage)
	})

	t.Run("assistant message with query results", func(t *testing.T) {
		msg := UserMessageDetail{
			ID:        "msg-456",
			Role:      "assistant",
			Content:   "The total sales is $10,000.",
			Timestamp: time.Now(),
			QueryResults: []UserQueryResult{
				{
					Query:    "SELECT SUM(amount) FROM sales",
					Summary:  "Total sales calculation",
					RowCount: 1,
					Data: []map[string]interface{}{
						{"sum": 10000},
					},
				},
			},
			Usage: &UserUsageStats{
				PromptTokens:     50,
				CompletionTokens: 20,
				TotalTokens:      70,
			},
		}

		assert.Equal(t, "assistant", msg.Role)
		assert.Len(t, msg.QueryResults, 1)
		assert.Equal(t, 1, msg.QueryResults[0].RowCount)
		assert.NotNil(t, msg.Usage)
		assert.Equal(t, 70, msg.Usage.TotalTokens)
	})
}

func TestUserQueryResult_Struct(t *testing.T) {
	t.Run("query result with data", func(t *testing.T) {
		result := UserQueryResult{
			Query:    "SELECT * FROM users LIMIT 2",
			Summary:  "Retrieved 2 users",
			RowCount: 2,
			Data: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
		}

		assert.Equal(t, "SELECT * FROM users LIMIT 2", result.Query)
		assert.Equal(t, 2, result.RowCount)
		assert.Len(t, result.Data, 2)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		result := UserQueryResult{
			Query:    "SELECT COUNT(*) FROM items",
			Summary:  "Count query",
			RowCount: 1,
			Data: []map[string]interface{}{
				{"count": 42},
			},
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"query":"SELECT COUNT(*) FROM items"`)
		assert.Contains(t, string(data), `"row_count":1`)
	})
}

func TestUserUsageStats_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		usage := UserUsageStats{
			PromptTokens:     100,
			CompletionTokens: 50,
			TotalTokens:      150,
		}

		assert.Equal(t, 100, usage.PromptTokens)
		assert.Equal(t, 50, usage.CompletionTokens)
		assert.Equal(t, 150, usage.TotalTokens)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		usage := UserUsageStats{
			PromptTokens:     200,
			CompletionTokens: 100,
			TotalTokens:      300,
		}

		data, err := json.Marshal(usage)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"prompt_tokens":200`)
		assert.Contains(t, string(data), `"completion_tokens":100`)
	})
}

func TestUserConversationDetail_Struct(t *testing.T) {
	t.Run("conversation with messages", func(t *testing.T) {
		title := "Sales Report"

		detail := UserConversationDetail{
			ID:          "conv-789",
			ChatbotName: "analytics-bot",
			Namespace:   "default",
			Title:       &title,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Messages: []UserMessageDetail{
				{ID: "msg-1", Role: "user", Content: "Show sales"},
				{ID: "msg-2", Role: "assistant", Content: "Here are the sales..."},
			},
		}

		assert.Equal(t, "conv-789", detail.ID)
		assert.Equal(t, "Sales Report", *detail.Title)
		assert.Len(t, detail.Messages, 2)
	})
}

func TestListUserConversationsOptions_Struct(t *testing.T) {
	t.Run("all fields", func(t *testing.T) {
		chatbot := "my-bot"
		namespace := "prod"

		opts := ListUserConversationsOptions{
			UserID:      "user-123",
			ChatbotName: &chatbot,
			Namespace:   &namespace,
			Limit:       20,
			Offset:      10,
		}

		assert.Equal(t, "user-123", opts.UserID)
		assert.Equal(t, "my-bot", *opts.ChatbotName)
		assert.Equal(t, "prod", *opts.Namespace)
		assert.Equal(t, 20, opts.Limit)
		assert.Equal(t, 10, opts.Offset)
	})

	t.Run("minimal options", func(t *testing.T) {
		opts := ListUserConversationsOptions{
			UserID: "user-456",
			Limit:  10,
		}

		assert.Equal(t, "user-456", opts.UserID)
		assert.Nil(t, opts.ChatbotName)
		assert.Nil(t, opts.Namespace)
		assert.Equal(t, 0, opts.Offset)
	})
}

func TestListUserConversationsResult_Struct(t *testing.T) {
	t.Run("result with conversations", func(t *testing.T) {
		result := ListUserConversationsResult{
			Conversations: []UserConversationSummary{
				{ID: "conv-1", ChatbotName: "bot1", MessageCount: 5},
				{ID: "conv-2", ChatbotName: "bot2", MessageCount: 3},
			},
			Total:   10,
			HasMore: true,
		}

		assert.Len(t, result.Conversations, 2)
		assert.Equal(t, 10, result.Total)
		assert.True(t, result.HasMore)
	})

	t.Run("empty result", func(t *testing.T) {
		result := ListUserConversationsResult{
			Conversations: []UserConversationSummary{},
			Total:         0,
			HasMore:       false,
		}

		assert.Empty(t, result.Conversations)
		assert.Equal(t, 0, result.Total)
		assert.False(t, result.HasMore)
	})
}
