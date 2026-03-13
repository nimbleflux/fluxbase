package api

import (
	"time"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/nimbleflux/fluxbase/internal/webhook"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookResponse represents a webhook response without the secret
// H-21: WebhookResponse DTO excludes secret field for security
type WebhookResponse struct {
	ID                  uuid.UUID             `json:"id"`
	Name                string                `json:"name"`
	Description         *string               `json:"description,omitempty"`
	URL                 string                `json:"url"`
	Enabled             bool                  `json:"enabled"`
	Events              []webhook.EventConfig `json:"events"`
	MaxRetries          int                   `json:"max_retries"`
	RetryBackoffSeconds int                   `json:"retry_backoff_seconds"`
	TimeoutSeconds      int                   `json:"timeout_seconds"`
	Headers             map[string]string     `json:"headers"`
	Scope               string                `json:"scope"` // "user" or "global"
	CreatedBy           *uuid.UUID            `json:"created_by,omitempty"`
	CreatedAt           time.Time             `json:"created_at"`
	UpdatedAt           time.Time             `json:"updated_at"`
}

// toWebhookResponse converts a webhook.Webhook to WebhookResponse (without secret)
func toWebhookResponse(w webhook.Webhook) WebhookResponse {
	return WebhookResponse{
		ID:                  w.ID,
		Name:                w.Name,
		Description:         w.Description,
		URL:                 w.URL,
		Enabled:             w.Enabled,
		Events:              w.Events,
		MaxRetries:          w.MaxRetries,
		RetryBackoffSeconds: w.RetryBackoffSeconds,
		TimeoutSeconds:      w.TimeoutSeconds,
		Headers:             w.Headers,
		Scope:               w.Scope,
		CreatedBy:           w.CreatedBy,
		CreatedAt:           w.CreatedAt,
		UpdatedAt:           w.UpdatedAt,
	}
}

// WebhookHandler handles HTTP requests for webhooks
type WebhookHandler struct {
	webhookService *webhook.WebhookService
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(webhookService *webhook.WebhookService) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
	}
}

// RegisterRoutes registers webhook routes with authentication
func (h *WebhookHandler) RegisterRoutes(app *fiber.App, authService *auth.Service, clientKeyService *auth.ClientKeyService, db *pgxpool.Pool, jwtManager *auth.JWTManager) {
	// Apply authentication middleware to all webhook routes
	webhooks := app.Group("/api/v1/webhooks",
		middleware.RequireAuthOrServiceKey(authService, clientKeyService, db, jwtManager),
	)

	// Read operations require read:webhooks scope
	webhooks.Get("/", middleware.RequireScope(auth.ScopeWebhooksRead), h.ListWebhooks)
	webhooks.Get("/:id", middleware.RequireScope(auth.ScopeWebhooksRead), h.GetWebhook)
	webhooks.Get("/:id/deliveries", middleware.RequireScope(auth.ScopeWebhooksRead), h.ListDeliveries)

	// Write operations require write:webhooks scope
	webhooks.Post("/", middleware.RequireScope(auth.ScopeWebhooksWrite), h.CreateWebhook)
	webhooks.Patch("/:id", middleware.RequireScope(auth.ScopeWebhooksWrite), h.UpdateWebhook)
	webhooks.Delete("/:id", middleware.RequireScope(auth.ScopeWebhooksWrite), h.DeleteWebhook)
	webhooks.Post("/:id/test", middleware.RequireScope(auth.ScopeWebhooksWrite), h.TestWebhook)
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(c fiber.Ctx) error {
	var req webhook.Webhook
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validation
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}
	if req.URL == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "URL is required",
		})
	}

	// Set defaults
	if req.MaxRetries == 0 {
		req.MaxRetries = 3
	}
	if req.RetryBackoffSeconds == 0 {
		req.RetryBackoffSeconds = 5
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 30
	}
	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}
	if req.Scope == "" {
		req.Scope = "user"
	}

	// Set CreatedBy from authenticated user
	if uid := c.Locals("user_id"); uid != nil {
		if uidStr, ok := uid.(string); ok {
			if parsed, err := uuid.Parse(uidStr); err == nil {
				req.CreatedBy = &parsed
			}
		}
	}

	err := h.webhookService.Create(c.RequestCtx(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// H-21: Return WebhookResponse (without secret)
	return c.Status(fiber.StatusCreated).JSON(toWebhookResponse(req))
}

// ListWebhooks lists all webhooks
func (h *WebhookHandler) ListWebhooks(c fiber.Ctx) error {
	webhooks, err := h.webhookService.List(c.RequestCtx())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// H-21: Convert to WebhookResponse (without secret)
	responses := make([]WebhookResponse, len(webhooks))
	for i, wh := range webhooks {
		responses[i] = toWebhookResponse(*wh)
	}

	return c.JSON(responses)
}

// GetWebhook retrieves a webhook by ID
func (h *WebhookHandler) GetWebhook(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid webhook ID",
		})
	}

	wh, err := h.webhookService.Get(c.RequestCtx(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Webhook not found",
		})
	}

	// H-21: Return WebhookResponse (without secret)
	return c.JSON(toWebhookResponse(*wh))
}

// UpdateWebhook updates a webhook
func (h *WebhookHandler) UpdateWebhook(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid webhook ID",
		})
	}

	var req webhook.Webhook
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.webhookService.Update(c.RequestCtx(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Webhook updated successfully",
	})
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid webhook ID",
		})
	}

	err = h.webhookService.Delete(c.RequestCtx(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Webhook deleted successfully",
	})
}

// TestWebhook sends a test webhook
func (h *WebhookHandler) TestWebhook(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid webhook ID",
		})
	}

	wh, err := h.webhookService.Get(c.RequestCtx(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Webhook not found",
		})
	}

	// Create test payload
	testPayload := &webhook.WebhookPayload{
		Event:     "TEST",
		Table:     "test",
		Schema:    "public",
		Record:    []byte(`{"test": true}`),
		Timestamp: c.RequestCtx().Time(),
	}

	err = h.webhookService.Deliver(c.RequestCtx(), wh, testPayload)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Test webhook sent successfully",
	})
}

// ListDeliveries lists webhook deliveries
func (h *WebhookHandler) ListDeliveries(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid webhook ID",
		})
	}

	// Default limit is 50
	limit := 50
	if limitParam := c.Query("limit"); limitParam != "" {
		parsedLimit := fiber.Query[int](c, "limit", 50)
		limit = parsedLimit
	}

	deliveries, err := h.webhookService.ListDeliveries(c.RequestCtx(), id, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(deliveries)
}

// fiber:context-methods migrated
