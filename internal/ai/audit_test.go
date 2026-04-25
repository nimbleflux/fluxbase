package ai

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditLogger(t *testing.T) {
	t.Run("creates audit logger with nil db", func(t *testing.T) {
		logger := NewAuditLogger(nil)
		assert.NotNil(t, logger)
		assert.Nil(t, logger.DB)
	})
}

func TestAuditEntry_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		chatbotID := "chatbot-123"
		conversationID := "conv-456"
		messageID := "msg-789"
		userID := "user-abc"
		sanitizedSQL := "SELECT * FROM users"
		validPassed := true
		success := true
		errorMsg := ""
		rowsReturned := 10
		durationMs := 25
		rlsUserID := "user-abc"
		rlsRole := "authenticated"
		ip := net.ParseIP("192.168.1.1")
		userAgent := "Mozilla/5.0"

		entry := AuditEntry{
			ID:                  "audit-001",
			ChatbotID:           &chatbotID,
			ConversationID:      &conversationID,
			MessageID:           &messageID,
			UserID:              &userID,
			GeneratedSQL:        "SELECT * FROM users WHERE id = 1",
			SanitizedSQL:        &sanitizedSQL,
			Executed:            true,
			ValidationPassed:    &validPassed,
			ValidationErrors:    []string{},
			Success:             &success,
			ErrorMessage:        &errorMsg,
			RowsReturned:        &rowsReturned,
			ExecutionDurationMs: &durationMs,
			TablesAccessed:      []string{"users"},
			OperationsUsed:      []string{"SELECT"},
			RLSUserID:           &rlsUserID,
			RLSRole:             &rlsRole,
			IPAddress:           &ip,
			UserAgent:           &userAgent,
			CreatedAt:           time.Now(),
		}

		assert.Equal(t, "audit-001", entry.ID)
		assert.Equal(t, "chatbot-123", *entry.ChatbotID)
		assert.Equal(t, "conv-456", *entry.ConversationID)
		assert.Equal(t, "msg-789", *entry.MessageID)
		assert.Equal(t, "user-abc", *entry.UserID)
		assert.True(t, entry.Executed)
		assert.True(t, *entry.ValidationPassed)
		assert.True(t, *entry.Success)
		assert.Equal(t, 10, *entry.RowsReturned)
		assert.Equal(t, 25, *entry.ExecutionDurationMs)
		assert.Equal(t, []string{"users"}, entry.TablesAccessed)
		assert.Equal(t, []string{"SELECT"}, entry.OperationsUsed)
		assert.NotNil(t, entry.IPAddress)
	})

	t.Run("zero value has expected defaults", func(t *testing.T) {
		var entry AuditEntry
		assert.Empty(t, entry.ID)
		assert.Nil(t, entry.ChatbotID)
		assert.Nil(t, entry.ConversationID)
		assert.False(t, entry.Executed)
		assert.Nil(t, entry.ValidationPassed)
		assert.Nil(t, entry.Success)
		assert.Nil(t, entry.TablesAccessed)
		assert.True(t, entry.CreatedAt.IsZero())
	})

	t.Run("failed query entry", func(t *testing.T) {
		success := false
		errorMsg := "Permission denied"
		validPassed := false

		entry := AuditEntry{
			ID:               "audit-fail",
			GeneratedSQL:     "DELETE FROM users",
			Executed:         false,
			ValidationPassed: &validPassed,
			ValidationErrors: []string{"DELETE not allowed", "users table restricted"},
			Success:          &success,
			ErrorMessage:     &errorMsg,
		}

		assert.False(t, *entry.ValidationPassed)
		assert.Len(t, entry.ValidationErrors, 2)
		assert.False(t, *entry.Success)
		assert.Contains(t, *entry.ErrorMessage, "Permission denied")
	})
}

func TestAuditStats_Struct(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		stats := AuditStats{
			TotalQueries:      100,
			ExecutedQueries:   90,
			FailedQueries:     5,
			RejectedQueries:   5,
			AverageDurationMs: 25.5,
		}

		assert.Equal(t, int64(100), stats.TotalQueries)
		assert.Equal(t, int64(90), stats.ExecutedQueries)
		assert.Equal(t, int64(5), stats.FailedQueries)
		assert.Equal(t, int64(5), stats.RejectedQueries)
		assert.Equal(t, 25.5, stats.AverageDurationMs)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		stats := AuditStats{
			TotalQueries:      50,
			ExecutedQueries:   45,
			FailedQueries:     3,
			RejectedQueries:   2,
			AverageDurationMs: 15.2,
		}

		data, err := json.Marshal(stats)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"total_queries":50`)
		assert.Contains(t, string(data), `"average_duration_ms":15.2`)
	})

	t.Run("zero value stats", func(t *testing.T) {
		var stats AuditStats
		assert.Equal(t, int64(0), stats.TotalQueries)
		assert.Equal(t, float64(0), stats.AverageDurationMs)
	})
}

func TestAuditEntry_IPAddressParsing(t *testing.T) {
	t.Run("IPv4 address", func(t *testing.T) {
		ip := net.ParseIP("192.168.1.100")
		entry := AuditEntry{
			IPAddress: &ip,
		}
		assert.NotNil(t, entry.IPAddress)
		assert.Equal(t, "192.168.1.100", entry.IPAddress.String())
	})

	t.Run("IPv6 address", func(t *testing.T) {
		ip := net.ParseIP("2001:db8::1")
		entry := AuditEntry{
			IPAddress: &ip,
		}
		assert.NotNil(t, entry.IPAddress)
		assert.Equal(t, "2001:db8::1", entry.IPAddress.String())
	})

	t.Run("invalid IP returns nil", func(t *testing.T) {
		ip := net.ParseIP("invalid-ip")
		assert.Nil(t, ip)
	})
}

func TestAuditEntry_ValidationErrors(t *testing.T) {
	t.Run("empty validation errors", func(t *testing.T) {
		entry := AuditEntry{
			ValidationErrors: []string{},
		}
		assert.Empty(t, entry.ValidationErrors)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		entry := AuditEntry{
			ValidationErrors: []string{
				"Table not allowed: secrets",
				"Operation not allowed: DELETE",
				"Column not accessible: password",
			},
		}
		assert.Len(t, entry.ValidationErrors, 3)
		assert.Contains(t, entry.ValidationErrors, "Table not allowed: secrets")
	})
}
