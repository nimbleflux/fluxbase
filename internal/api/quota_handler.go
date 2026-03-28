package api

import (
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/auth"
)

// QuotaHandler handles quota-related HTTP requests
type QuotaHandler struct {
	quotaService    *ai.QuotaService
	userMgmtService *auth.UserManagementService
}

// NewQuotaHandler creates a new quota handler
func NewQuotaHandler(quotaService *ai.QuotaService, userMgmtService *auth.UserManagementService) *QuotaHandler {
	return &QuotaHandler{
		quotaService:    quotaService,
		userMgmtService: userMgmtService,
	}
}

// ListUsersWithQuotas returns all users with their quota information
// GET /api/v1/admin/users
func (h *QuotaHandler) ListUsersWithQuotas(c fiber.Ctx) error {
	if h.userMgmtService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "User management service not initialized",
		})
	}

	// Get all users (no tenant filtering for quota listing)
	users, err := h.userMgmtService.ListEnrichedUsers(c.RequestCtx(), "app", "")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	if users == nil {
		users = []*auth.EnrichedUser{}
	}

	// Enrich with quota information
	result := make([]fiber.Map, 0, len(users))
	for _, user := range users {
		userMap := fiber.Map{
			"id":        user.ID,
			"email":     user.Email,
			"full_name": nil,
		}

		// Try to get full_name from user metadata
		if fn, ok := user.UserMetadata["full_name"]; ok {
			if fullName, ok := fn.(string); ok {
				userMap["full_name"] = fullName
			}
		}

		// Try to get user quota
		quota, err := h.quotaService.GetUserQuotaUsage(c.RequestCtx(), user.ID)
		if err == nil && quota != nil {
			userMap["quota"] = quota
		} else {
			// User has no custom quota, will use system defaults
			userMap["quota"] = nil
		}

		result = append(result, userMap)
	}

	return c.JSON(result)
}

// GetUserQuota returns quota information for a specific user
// GET /api/v1/admin/users/:id/quota
func (h *QuotaHandler) GetUserQuota(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	quota, err := h.quotaService.GetUserQuotaUsage(c.RequestCtx(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found or quota not set",
		})
	}

	return c.JSON(quota)
}

// SetUserQuota sets quota limits for a specific user
// PUT /api/v1/admin/users/:id/quota
func (h *QuotaHandler) SetUserQuota(c fiber.Ctx) error {
	userID := c.Params("id")
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID is required",
		})
	}

	var req struct {
		MaxDocuments    int   `json:"max_documents"`
		MaxChunks       int   `json:"max_chunks"`
		MaxStorageBytes int64 `json:"max_storage_bytes"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate limits
	if req.MaxDocuments <= 0 || req.MaxDocuments > 1000000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "max_documents must be between 1 and 1000000",
		})
	}

	if req.MaxChunks <= 0 || req.MaxChunks > 10000000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "max_chunks must be between 1 and 10000000",
		})
	}

	if req.MaxStorageBytes <= 0 || req.MaxStorageBytes > 1024*1024*1024*1024 { // 1TB max
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "max_storage_bytes must be between 1 and 1TB",
		})
	}

	setReq := ai.SetUserQuotaRequest{
		MaxDocuments:    req.MaxDocuments,
		MaxChunks:       req.MaxChunks,
		MaxStorageBytes: req.MaxStorageBytes,
	}

	if err := h.quotaService.SetUserQuota(c.RequestCtx(), userID, setReq); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set quota: %v", err),
		})
	}

	// Return the updated quota
	quota, err := h.quotaService.GetUserQuotaUsage(c.RequestCtx(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve updated quota",
		})
	}

	return c.JSON(quota)
}
