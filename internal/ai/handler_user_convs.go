package ai

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// ============================================================================
// USER CONVERSATION ENDPOINTS
// ============================================================================

// UpdateConversationTitleRequest represents the request body for updating title
type UpdateConversationTitleRequest struct {
	Title string `json:"title"`
}

// ListUserConversations lists the authenticated user's conversations
// GET /api/v1/ai/conversations
func (h *Handler) ListUserConversations(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)

	// Get authenticated user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Parse query params
	limit := fiber.Query[int](c, "limit", 50)
	if limit > 100 {
		limit = 100 // Cap at 100
	}
	if limit < 1 {
		limit = 50
	}
	offset := fiber.Query[int](c, "offset", 0)
	if offset < 0 {
		offset = 0
	}

	// Build options
	opts := ListUserConversationsOptions{
		UserID: userIDStr,
		Limit:  limit,
		Offset: offset,
	}

	if chatbot := c.Query("chatbot"); chatbot != "" {
		opts.ChatbotName = &chatbot
	}
	if namespace := c.Query("namespace"); namespace != "" {
		if namespace == "default" {
			namespace = ""
		}
		opts.Namespace = &namespace
	}

	// Query conversations
	result, err := h.storage.ListUserConversations(ctx, opts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list user conversations")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list conversations",
		})
	}

	return c.JSON(result)
}

// GetUserConversation retrieves a single conversation with messages
// GET /api/v1/ai/conversations/:id
func (h *Handler) GetUserConversation(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	conversationID := c.Params("id")

	// Get authenticated user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	conversation, err := h.storage.GetUserConversation(ctx, userIDStr, conversationID)
	if err != nil {
		log.Error().Err(err).Str("id", conversationID).Msg("Failed to get conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get conversation",
		})
	}

	if conversation == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Conversation not found",
		})
	}

	return c.JSON(conversation)
}

// DeleteUserConversation deletes a user's conversation
// DELETE /api/v1/ai/conversations/:id
func (h *Handler) DeleteUserConversation(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	conversationID := c.Params("id")

	// Get authenticated user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	err := h.storage.DeleteUserConversation(ctx, userIDStr, conversationID)
	if err != nil {
		if err.Error() == "conversation not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Conversation not found",
			})
		}
		log.Error().Err(err).Str("id", conversationID).Msg("Failed to delete conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete conversation",
		})
	}

	return c.JSON(fiber.Map{
		"deleted": true,
		"id":      conversationID,
	})
}

// UpdateUserConversation updates a conversation (title only for now)
// PATCH /api/v1/ai/conversations/:id
func (h *Handler) UpdateUserConversation(c fiber.Ctx) error {
	ctx := middleware.CtxWithTenant(c)
	conversationID := c.Params("id")

	// Get authenticated user ID from context
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req UpdateConversationTitleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate title
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title cannot be empty",
		})
	}
	if len(req.Title) > 200 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Title must be 200 characters or less",
		})
	}

	err := h.storage.UpdateConversationTitle(ctx, userIDStr, conversationID, req.Title)
	if err != nil {
		if err.Error() == "conversation not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Conversation not found",
			})
		}
		log.Error().Err(err).Str("id", conversationID).Msg("Failed to update conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update conversation",
		})
	}

	// Return updated conversation
	conversation, err := h.storage.GetUserConversation(ctx, userIDStr, conversationID)
	if err != nil {
		log.Error().Err(err).Str("id", conversationID).Msg("Failed to get updated conversation")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Conversation updated but failed to retrieve",
		})
	}

	return c.JSON(conversation)
}
