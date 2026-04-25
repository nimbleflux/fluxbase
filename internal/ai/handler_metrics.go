package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// ============================================================================
// METRICS ENDPOINT
// ============================================================================

// ChatbotMetric represents metrics for a single chatbot
type ChatbotMetric struct {
	ChatbotID   string `json:"chatbot_id"`
	ChatbotName string `json:"chatbot_name"`
	Requests    int64  `json:"requests"`
	Tokens      int64  `json:"tokens"`
	ErrorCount  int64  `json:"error_count"`
}

// ProviderMetric represents metrics for a single provider
type ProviderMetric struct {
	ProviderID   string  `json:"provider_id"`
	ProviderName string  `json:"provider_name"`
	Requests     int64   `json:"requests"`
	AvgLatencyMS float64 `json:"avg_latency_ms"`
}

// AIMetrics represents aggregated AI metrics
type AIMetrics struct {
	TotalRequests         int64            `json:"total_requests"`
	TotalTokens           int64            `json:"total_tokens"`
	TotalPromptTokens     int64            `json:"total_prompt_tokens"`
	TotalCompletionTokens int64            `json:"total_completion_tokens"`
	ActiveConversations   int              `json:"active_conversations"`
	TotalConversations    int              `json:"total_conversations"`
	ChatbotStats          []ChatbotMetric  `json:"chatbot_stats"`
	ProviderStats         []ProviderMetric `json:"provider_stats"`
	ErrorRate             float64          `json:"error_rate"`
	AvgResponseTimeMS     float64          `json:"avg_response_time_ms"`
}

// GetAIMetrics returns aggregated AI metrics
// GET /api/v1/admin/ai/metrics
func (h *Handler) GetAIMetrics(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	metrics := AIMetrics{
		ChatbotStats:  make([]ChatbotMetric, 0),
		ProviderStats: make([]ProviderMetric, 0),
	}

	// Query conversation metrics
	convQuery := `
		SELECT
			COUNT(*) as total_conversations,
			COUNT(*) FILTER (WHERE status = 'active') as active_conversations,
			COALESCE(SUM(total_prompt_tokens), 0) as total_prompt_tokens,
			COALESCE(SUM(total_completion_tokens), 0) as total_completion_tokens
		FROM ai.conversations
	`
	err := h.storage.db.QueryRow(ctx, convQuery).Scan(
		&metrics.TotalConversations,
		&metrics.ActiveConversations,
		&metrics.TotalPromptTokens,
		&metrics.TotalCompletionTokens,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query conversation metrics")
	}

	metrics.TotalTokens = metrics.TotalPromptTokens + metrics.TotalCompletionTokens

	// Query audit log for request counts and error rates
	auditQuery := `
		SELECT
			COUNT(*) as total_requests,
			COUNT(*) FILTER (WHERE success = false) as error_count,
			COALESCE(AVG(execution_duration_ms), 0) as avg_duration
		FROM ai.query_audit_log
		WHERE executed = true
	`
	var errorCount int64
	err = h.storage.db.QueryRow(ctx, auditQuery).Scan(
		&metrics.TotalRequests,
		&errorCount,
		&metrics.AvgResponseTimeMS,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query audit log metrics")
	}

	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(errorCount) / float64(metrics.TotalRequests) * 100
	}

	// Query per-chatbot metrics
	chatbotQuery := `
		SELECT
			c.id,
			c.name,
			COUNT(a.id) as requests,
			COALESCE(SUM(conv.total_prompt_tokens + conv.total_completion_tokens), 0) as tokens,
			COUNT(a.id) FILTER (WHERE a.success = false) as error_count
		FROM ai.chatbots c
		LEFT JOIN ai.query_audit_log a ON a.chatbot_id = c.id
		LEFT JOIN ai.conversations conv ON conv.chatbot_id = c.id
		GROUP BY c.id, c.name
		HAVING COUNT(a.id) > 0
		ORDER BY requests DESC
		LIMIT 20
	`
	rows, err := h.storage.db.Query(ctx, chatbotQuery)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query chatbot metrics")
	} else {
		defer rows.Close()
		for rows.Next() {
			var metric ChatbotMetric
			err := rows.Scan(
				&metric.ChatbotID,
				&metric.ChatbotName,
				&metric.Requests,
				&metric.Tokens,
				&metric.ErrorCount,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to scan chatbot metric")
				continue
			}
			metrics.ChatbotStats = append(metrics.ChatbotStats, metric)
		}
	}

	return c.JSON(metrics)
}

// ConversationSummary represents a conversation with basic info
type ConversationSummary struct {
	ID                    string     `json:"id"`
	ChatbotID             string     `json:"chatbot_id"`
	ChatbotName           string     `json:"chatbot_name"`
	UserID                *string    `json:"user_id"`
	UserEmail             *string    `json:"user_email"`
	SessionID             *string    `json:"session_id"`
	Title                 *string    `json:"title"`
	Status                string     `json:"status"`
	TurnCount             int        `json:"turn_count"`
	TotalPromptTokens     int        `json:"total_prompt_tokens"`
	TotalCompletionTokens int        `json:"total_completion_tokens"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	LastMessageAt         *time.Time `json:"last_message_at"`
}

// GetConversations returns a list of AI conversations with optional filters
// GET /api/v1/admin/ai/conversations?chatbot_id=X&user_id=Y&status=active&limit=50
func (h *Handler) GetConversations(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	// Parse query parameters
	chatbotID := c.Query("chatbot_id")
	userID := c.Query("user_id")
	status := c.Query("status")
	limit := fiber.Query[int](c, "limit", 50)
	offset := fiber.Query[int](c, "offset", 0)

	// Build query
	query := `
		SELECT
			c.id,
			c.chatbot_id,
			cb.name as chatbot_name,
			c.user_id,
			u.email as user_email,
			c.session_id,
			c.title,
			c.status,
			c.turn_count,
			c.total_prompt_tokens,
			c.total_completion_tokens,
			c.created_at,
			c.updated_at,
			c.last_message_at
		FROM ai.conversations c
		LEFT JOIN ai.chatbots cb ON cb.id = c.chatbot_id
		LEFT JOIN auth.users u ON u.id = c.user_id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if chatbotID != "" {
		query += fmt.Sprintf(" AND c.chatbot_id = $%d", argIndex)
		args = append(args, chatbotID)
		argIndex++
	}

	if userID != "" {
		query += fmt.Sprintf(" AND c.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if status != "" {
		query += fmt.Sprintf(" AND c.status = $%d", argIndex)
		args = append(args, status)
		argIndex++
	}

	// Build count query with same filters (without LIMIT/OFFSET)
	countQuery := `
		SELECT COUNT(*)
		FROM ai.conversations c
		WHERE 1=1
	`
	countArgs := []interface{}{}
	countArgIndex := 1

	if chatbotID != "" {
		countQuery += fmt.Sprintf(" AND c.chatbot_id = $%d", countArgIndex)
		countArgs = append(countArgs, chatbotID)
		countArgIndex++
	}
	if userID != "" {
		countQuery += fmt.Sprintf(" AND c.user_id = $%d", countArgIndex)
		countArgs = append(countArgs, userID)
		countArgIndex++
	}
	if status != "" {
		countQuery += fmt.Sprintf(" AND c.status = $%d", countArgIndex)
		countArgs = append(countArgs, status)
	}

	var totalCount int
	if err := h.storage.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		log.Error().Err(err).Msg("Failed to count conversations")
		totalCount = 0
	}

	query += fmt.Sprintf(" ORDER BY c.last_message_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.storage.db.Query(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query conversations")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query conversations",
		})
	}
	defer rows.Close()

	conversations := make([]ConversationSummary, 0)
	for rows.Next() {
		var conv ConversationSummary
		err := rows.Scan(
			&conv.ID,
			&conv.ChatbotID,
			&conv.ChatbotName,
			&conv.UserID,
			&conv.UserEmail,
			&conv.SessionID,
			&conv.Title,
			&conv.Status,
			&conv.TurnCount,
			&conv.TotalPromptTokens,
			&conv.TotalCompletionTokens,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&conv.LastMessageAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan conversation")
			continue
		}
		conversations = append(conversations, conv)
	}

	return c.JSON(fiber.Map{
		"conversations": conversations,
		"total":         len(conversations),
		"total_count":   totalCount,
	})
}

// MessageDetail represents a message within a conversation
type MessageDetail struct {
	ID               string    `json:"id"`
	ConversationID   string    `json:"conversation_id"`
	Role             string    `json:"role"`
	Content          string    `json:"content"`
	ToolCallID       *string   `json:"tool_call_id"`
	ToolName         *string   `json:"tool_name"`
	ExecutedSQL      *string   `json:"executed_sql"`
	SQLResultSummary *string   `json:"sql_result_summary"`
	SQLRowCount      *int      `json:"sql_row_count"`
	SQLError         *string   `json:"sql_error"`
	SQLDurationMS    *int      `json:"sql_duration_ms"`
	PromptTokens     *int      `json:"prompt_tokens"`
	CompletionTokens *int      `json:"completion_tokens"`
	CreatedAt        time.Time `json:"created_at"`
	SequenceNumber   int       `json:"sequence_number"`
}

// GetConversationMessages returns all messages for a specific conversation
// GET /api/v1/admin/ai/conversations/:id/messages
func (h *Handler) GetConversationMessages(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	conversationID := c.Params("id")

	query := `
		SELECT
			id,
			conversation_id,
			role,
			content,
			tool_call_id,
			tool_name,
			executed_sql,
			sql_result_summary,
			sql_row_count,
			sql_error,
			sql_duration_ms,
			prompt_tokens,
			completion_tokens,
			created_at,
			sequence_number
		FROM ai.messages
		WHERE conversation_id = $1
		ORDER BY sequence_number ASC
	`

	rows, err := h.storage.db.Query(ctx, query, conversationID)
	if err != nil {
		log.Error().Err(err).Str("conversation_id", conversationID).Msg("Failed to query messages")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query messages",
		})
	}
	defer rows.Close()

	messages := make([]MessageDetail, 0)
	for rows.Next() {
		var msg MessageDetail
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.Role,
			&msg.Content,
			&msg.ToolCallID,
			&msg.ToolName,
			&msg.ExecutedSQL,
			&msg.SQLResultSummary,
			&msg.SQLRowCount,
			&msg.SQLError,
			&msg.SQLDurationMS,
			&msg.PromptTokens,
			&msg.CompletionTokens,
			&msg.CreatedAt,
			&msg.SequenceNumber,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan message")
			continue
		}
		messages = append(messages, msg)
	}

	return c.JSON(fiber.Map{
		"messages": messages,
		"total":    len(messages),
	})
}

// AuditLogEntry represents an audit log entry
type AuditLogEntry struct {
	ID                  string    `json:"id"`
	ChatbotID           *string   `json:"chatbot_id"`
	ChatbotName         *string   `json:"chatbot_name"`
	ConversationID      *string   `json:"conversation_id"`
	MessageID           *string   `json:"message_id"`
	UserID              *string   `json:"user_id"`
	UserEmail           *string   `json:"user_email"`
	GeneratedSQL        string    `json:"generated_sql"`
	SanitizedSQL        *string   `json:"sanitized_sql"`
	Executed            bool      `json:"executed"`
	ValidationPassed    *bool     `json:"validation_passed"`
	ValidationErrors    []string  `json:"validation_errors"`
	Success             *bool     `json:"success"`
	ErrorMessage        *string   `json:"error_message"`
	RowsReturned        *int      `json:"rows_returned"`
	ExecutionDurationMS *int      `json:"execution_duration_ms"`
	TablesAccessed      []string  `json:"tables_accessed"`
	OperationsUsed      []string  `json:"operations_used"`
	IPAddress           *string   `json:"ip_address"`
	UserAgent           *string   `json:"user_agent"`
	CreatedAt           time.Time `json:"created_at"`
}

// GetAuditLog returns audit log entries with optional filters
// GET /api/v1/admin/ai/audit?chatbot_id=X&user_id=Y&success=true&limit=100
func (h *Handler) GetAuditLog(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	// Parse query parameters
	chatbotID := c.Query("chatbot_id")
	userID := c.Query("user_id")
	successStr := c.Query("success")
	limit := fiber.Query[int](c, "limit", 100)
	offset := fiber.Query[int](c, "offset", 0)

	// Build query
	query := `
		SELECT
			a.id,
			a.chatbot_id,
			cb.name as chatbot_name,
			a.conversation_id,
			a.message_id,
			a.user_id,
			u.email as user_email,
			a.generated_sql,
			a.sanitized_sql,
			a.executed,
			a.validation_passed,
			a.validation_errors,
			a.success,
			a.error_message,
			a.rows_returned,
			a.execution_duration_ms,
			a.tables_accessed,
			a.operations_used,
			a.ip_address,
			a.user_agent,
			a.created_at
		FROM ai.query_audit_log a
		LEFT JOIN ai.chatbots cb ON cb.id = a.chatbot_id
		LEFT JOIN auth.users u ON u.id = a.user_id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	if chatbotID != "" {
		query += fmt.Sprintf(" AND a.chatbot_id = $%d", argIndex)
		args = append(args, chatbotID)
		argIndex++
	}

	if userID != "" {
		query += fmt.Sprintf(" AND a.user_id = $%d", argIndex)
		args = append(args, userID)
		argIndex++
	}

	if successStr != "" {
		success := successStr == "true"
		query += fmt.Sprintf(" AND a.success = $%d", argIndex)
		args = append(args, success)
		argIndex++
	}

	// Build count query with same filters (without LIMIT/OFFSET)
	countQuery := `
		SELECT COUNT(*)
		FROM ai.query_audit_log a
		WHERE 1=1
	`
	countArgs := []interface{}{}
	countArgIndex := 1

	if chatbotID != "" {
		countQuery += fmt.Sprintf(" AND a.chatbot_id = $%d", countArgIndex)
		countArgs = append(countArgs, chatbotID)
		countArgIndex++
	}
	if userID != "" {
		countQuery += fmt.Sprintf(" AND a.user_id = $%d", countArgIndex)
		countArgs = append(countArgs, userID)
		countArgIndex++
	}
	if successStr != "" {
		success := successStr == "true"
		countQuery += fmt.Sprintf(" AND a.success = $%d", countArgIndex)
		countArgs = append(countArgs, success)
	}

	var totalCount int
	if err := h.storage.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		log.Error().Err(err).Msg("Failed to count audit log entries")
		totalCount = 0
	}

	query += fmt.Sprintf(" ORDER BY a.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := h.storage.db.Query(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query audit log")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query audit log",
		})
	}
	defer rows.Close()

	entries := make([]AuditLogEntry, 0)
	for rows.Next() {
		var entry AuditLogEntry
		err := rows.Scan(
			&entry.ID,
			&entry.ChatbotID,
			&entry.ChatbotName,
			&entry.ConversationID,
			&entry.MessageID,
			&entry.UserID,
			&entry.UserEmail,
			&entry.GeneratedSQL,
			&entry.SanitizedSQL,
			&entry.Executed,
			&entry.ValidationPassed,
			&entry.ValidationErrors,
			&entry.Success,
			&entry.ErrorMessage,
			&entry.RowsReturned,
			&entry.ExecutionDurationMS,
			&entry.TablesAccessed,
			&entry.OperationsUsed,
			&entry.IPAddress,
			&entry.UserAgent,
			&entry.CreatedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan audit log entry")
			continue
		}
		entries = append(entries, entry)
	}

	return c.JSON(fiber.Map{
		"entries":     entries,
		"total":       len(entries),
		"total_count": totalCount,
	})
}

// ============================================================================
// AUTO-LOAD HELPER
// ============================================================================

// AutoLoadChatbots loads chatbots from filesystem to database on startup
func (h *Handler) AutoLoadChatbots(ctx context.Context) error {
	if !h.config.AutoLoadOnBoot {
		log.Info().Msg("Auto-load chatbots disabled, skipping")
		return nil
	}

	log.Info().Str("dir", h.config.ChatbotsDir).Msg("Auto-loading chatbots from filesystem")

	// Load from filesystem
	chatbots, err := h.loader.LoadAll()
	if err != nil {
		return err
	}

	if len(chatbots) == 0 {
		log.Info().Msg("No chatbots found in filesystem")
		return nil
	}

	// Upsert each chatbot
	created, updated := 0, 0
	for _, cb := range chatbots {
		existing, err := h.storage.GetChatbotByName(ctx, cb.Namespace, cb.Name)
		if err != nil {
			log.Error().Err(err).Str("name", cb.Name).Msg("Failed to check existing chatbot")
			continue
		}

		if existing != nil {
			cb.ID = existing.ID
			cb.CreatedAt = existing.CreatedAt
			cb.CreatedBy = existing.CreatedBy
			if err := h.storage.UpdateChatbot(ctx, cb); err != nil {
				log.Error().Err(err).Str("name", cb.Name).Msg("Failed to update chatbot")
				continue
			}
			updated++
		} else {
			if err := h.storage.CreateChatbot(ctx, cb); err != nil {
				log.Error().Err(err).Str("name", cb.Name).Msg("Failed to create chatbot")
				continue
			}
			created++
		}
	}

	log.Info().
		Int("created", created).
		Int("updated", updated).
		Int("total", len(chatbots)).
		Msg("Auto-loaded chatbots from filesystem")

	return nil
}
