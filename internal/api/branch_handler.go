package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/nimbleflux/fluxbase/internal/branching"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/rs/zerolog/log"
)

// BranchHandler handles branch management API endpoints
type BranchHandler struct {
	manager *branching.Manager
	router  *branching.Router
	config  config.BranchingConfig
}

// NewBranchHandler creates a new branch handler
func NewBranchHandler(manager *branching.Manager, router *branching.Router, cfg config.BranchingConfig) *BranchHandler {
	return &BranchHandler{
		manager: manager,
		router:  router,
		config:  cfg,
	}
}

// CreateBranchRequest represents the request body for creating a branch
type CreateBranchRequest struct {
	Name           string                  `json:"name" validate:"required,min=1,max=100"`
	ParentBranchID *uuid.UUID              `json:"parent_branch_id,omitempty"`
	DataCloneMode  branching.DataCloneMode `json:"data_clone_mode,omitempty"`
	Type           branching.BranchType    `json:"type,omitempty"`
	GitHubPRNumber *int                    `json:"github_pr_number,omitempty"`
	GitHubPRURL    *string                 `json:"github_pr_url,omitempty"`
	GitHubRepo     *string                 `json:"github_repo,omitempty"`
	ExpiresIn      *string                 `json:"expires_in,omitempty"` // Duration string like "24h", "7d"
}

// CreateBranch handles POST /admin/branches
func (h *BranchHandler) CreateBranch(c fiber.Ctx) error {
	if !h.config.Enabled {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "branching_disabled",
			"message": "Database branching is not enabled",
		})
	}

	var req CreateBranchRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Failed to parse request body: " + err.Error(),
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Branch name is required",
		})
	}

	// Get user ID from context
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}

	// Parse expires_in to ExpiresAt
	var expiresAt *time.Time
	if req.ExpiresIn != nil && *req.ExpiresIn != "" {
		duration, err := time.ParseDuration(*req.ExpiresIn)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation_error",
				"message": "Invalid expires_in duration: " + err.Error(),
			})
		}
		t := time.Now().Add(duration)
		expiresAt = &t
	}

	// Create branch request
	branchReq := branching.CreateBranchRequest{
		Name:           req.Name,
		ParentBranchID: req.ParentBranchID,
		DataCloneMode:  req.DataCloneMode,
		Type:           req.Type,
		GitHubPRNumber: req.GitHubPRNumber,
		GitHubPRURL:    req.GitHubPRURL,
		GitHubRepo:     req.GitHubRepo,
		ExpiresAt:      expiresAt,
	}

	branch, err := h.manager.CreateBranch(c.RequestCtx(), branchReq, userID)
	if err != nil {
		log.Error().Err(err).Str("name", req.Name).Msg("Failed to create branch")

		if errors.Is(err, branching.ErrBranchExists) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error":   "branch_exists",
				"message": "A branch with this name already exists",
			})
		}
		if errors.Is(err, branching.ErrMaxBranchesReached) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "max_branches_reached",
				"message": "Maximum number of branches has been reached",
			})
		}
		if errors.Is(err, branching.ErrInvalidSlug) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid_slug",
				"message": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "create_failed",
			"message": "Failed to create branch: " + err.Error(),
		})
	}

	// Warmup the connection pool
	if h.router != nil {
		go func() {
			if err := h.router.WarmupPool(c.RequestCtx(), branch.Slug); err != nil {
				log.Warn().Err(err).Str("slug", branch.Slug).Msg("Failed to warmup branch pool")
			}
		}()
	}

	return c.Status(fiber.StatusCreated).JSON(branch)
}

// ListBranches handles GET /admin/branches
func (h *BranchHandler) ListBranches(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	filter := branching.ListBranchesFilter{
		Limit:  100,
		Offset: 0,
	}

	// Parse query parameters
	if limit := fiber.Query[int](c, "limit", 100); limit > 0 && limit <= 1000 {
		filter.Limit = limit
	}
	if offset := fiber.Query[int](c, "offset", 0); offset >= 0 {
		filter.Offset = offset
	}
	if status := c.Query("status"); status != "" {
		s := branching.BranchStatus(status)
		filter.Status = &s
	}
	if branchType := c.Query("type"); branchType != "" {
		t := branching.BranchType(branchType)
		filter.Type = &t
	}
	if repo := c.Query("github_repo"); repo != "" {
		filter.GitHubRepo = &repo
	}

	// Get user ID for filtering their branches
	if c.Query("mine") == "true" {
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			if id, err := uuid.Parse(uid); err == nil {
				filter.CreatedBy = &id
			}
		}
	}

	branches, err := h.manager.GetStorage().ListBranches(c.RequestCtx(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list branches")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list branches",
		})
	}

	// Get total count
	total, err := h.manager.GetStorage().CountBranches(c.RequestCtx(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count branches")
		total = len(branches)
	}

	return c.JSON(fiber.Map{
		"branches": branches,
		"total":    total,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// GetBranch handles GET /admin/branches/:id
func (h *BranchHandler) GetBranch(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		// Try as slug
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		log.Error().Err(err).Str("id", idParam).Msg("Failed to get branch")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	return c.JSON(branch)
}

// DeleteBranch handles DELETE /admin/branches/:id
func (h *BranchHandler) DeleteBranch(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first
	var branchID uuid.UUID
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branchID = id
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		// Try as slug
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
		if err == nil {
			branchID = branch.ID
		}
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	// Get user ID from context
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}

	// Check authorization - service keys and dashboard admins bypass this check
	authType, _ := c.Locals("auth_type").(string)
	userRole, _ := c.Locals("user_role").(string)
	isAdmin := authType == "service_key" || userRole == "instance_admin" || userRole == "admin"

	if !isAdmin && userID != nil {
		// Check if user has admin access to the branch
		hasAccess, err := h.manager.GetStorage().HasAccess(c.RequestCtx(), branch.ID, *userID, branching.BranchAccessAdmin)
		if err != nil {
			log.Error().Err(err).Str("branch_id", branch.ID.String()).Msg("Failed to check branch access")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "access_check_failed",
				"message": "Failed to verify branch access",
			})
		}
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "access_denied",
				"message": "You do not have permission to delete this branch",
			})
		}
	}

	// Close the connection pool first
	if h.router != nil {
		h.router.ClosePool(branch.Slug)
	}

	// Delete the branch
	if err := h.manager.DeleteBranch(c.RequestCtx(), branchID, userID); err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to delete branch")

		if errors.Is(err, branching.ErrCannotDeleteMainBranch) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "cannot_delete_main",
				"message": "Cannot delete the main branch",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": "Failed to delete branch: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ResetBranch handles POST /admin/branches/:id/reset
func (h *BranchHandler) ResetBranch(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first
	var branchID uuid.UUID
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branchID = id
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		// Try as slug
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
		if err == nil {
			branchID = branch.ID
		}
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	// Get user ID from context
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}

	// Check authorization - service keys and dashboard admins bypass this check
	authType, _ := c.Locals("auth_type").(string)
	userRole, _ := c.Locals("user_role").(string)
	isAdmin := authType == "service_key" || userRole == "instance_admin" || userRole == "admin"

	if !isAdmin && userID != nil {
		// Check if user has admin access to the branch (reset is a destructive operation)
		hasAccess, err := h.manager.GetStorage().HasAccess(c.RequestCtx(), branch.ID, *userID, branching.BranchAccessAdmin)
		if err != nil {
			log.Error().Err(err).Str("branch_id", branch.ID.String()).Msg("Failed to check branch access")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "access_check_failed",
				"message": "Failed to verify branch access",
			})
		}
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "access_denied",
				"message": "You do not have permission to reset this branch",
			})
		}
	}

	// Close the connection pool before reset
	if h.router != nil {
		h.router.ClosePool(branch.Slug)
	}

	// Reset the branch
	if err := h.manager.ResetBranch(c.RequestCtx(), branchID, userID); err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to reset branch")

		if errors.Is(err, branching.ErrCannotDeleteMainBranch) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "cannot_reset_main",
				"message": "Cannot reset the main branch",
			})
		}

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "reset_failed",
			"message": "Failed to reset branch: " + err.Error(),
		})
	}

	// Refresh the connection pool
	if h.router != nil {
		if err := h.router.RefreshPool(c.RequestCtx(), branch.Slug); err != nil {
			log.Warn().Err(err).Str("slug", branch.Slug).Msg("Failed to refresh branch pool after reset")
		}
	}

	// Get updated branch
	updatedBranch, _ := h.manager.GetStorage().GetBranch(c.RequestCtx(), branchID)
	if updatedBranch != nil {
		return c.JSON(updatedBranch)
	}

	return c.JSON(fiber.Map{"status": "reset_complete"})
}

// GetBranchActivity handles GET /admin/branches/:id/activity
func (h *BranchHandler) GetBranchActivity(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first
	var branchID uuid.UUID

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branchID = id
	} else {
		// Try as slug
		branch, err := h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
		if err != nil {
			if errors.Is(err, branching.ErrBranchNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "branch_not_found",
					"message": "Branch not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "get_failed",
				"message": "Failed to get branch",
			})
		}
		branchID = branch.ID
	}

	limit := fiber.Query[int](c, "limit", 50)
	if limit > 100 {
		limit = 100
	}

	activity, err := h.manager.GetStorage().GetActivityLog(c.RequestCtx(), branchID, limit)
	if err != nil {
		log.Error().Err(err).Str("id", idParam).Msg("Failed to get branch activity")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_activity_failed",
			"message": "Failed to get branch activity",
		})
	}

	return c.JSON(fiber.Map{
		"activity": activity,
	})
}

// GetPoolStats handles GET /admin/branches/stats/pools
func (h *BranchHandler) GetPoolStats(c fiber.Ctx) error {
	// Nil check for router (can happen in tests)
	if h.router == nil {
		return c.JSON(fiber.Map{
			"pools": []map[string]interface{}{},
		})
	}

	stats := h.router.GetPoolStats()
	return c.JSON(fiber.Map{
		"pools": stats,
	})
}

// GetActiveBranch handles GET /admin/branches/active
func (h *BranchHandler) GetActiveBranch(c fiber.Ctx) error {
	// Nil check for router (can happen in tests)
	if h.router == nil {
		return c.JSON(fiber.Map{
			"branch": "main",
			"source": "default",
		})
	}

	branch := h.router.GetDefaultBranch()
	source := h.router.GetActiveBranchSource()

	return c.JSON(fiber.Map{
		"branch": branch,
		"source": source,
	})
}

// SetActiveBranchRequest represents the request body for setting the active branch
type SetActiveBranchRequest struct {
	Branch string `json:"branch" validate:"required"`
}

// SetActiveBranch handles POST /admin/branches/active
func (h *BranchHandler) SetActiveBranch(c fiber.Ctx) error {
	if !h.config.Enabled {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error":   "branching_disabled",
			"message": "Database branching is not enabled",
		})
	}

	var req SetActiveBranchRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Failed to parse request body: " + err.Error(),
		})
	}

	if req.Branch == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Branch slug is required",
		})
	}

	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	// Verify the branch exists (unless it's "main")
	if req.Branch != "main" {
		_, err := h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), req.Branch)
		if err != nil {
			if errors.Is(err, branching.ErrBranchNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "branch_not_found",
					"message": "Branch not found: " + req.Branch,
				})
			}
			log.Error().Err(err).Str("branch", req.Branch).Msg("Failed to verify branch")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "verification_failed",
				"message": "Failed to verify branch exists",
			})
		}
	}

	// Nil check for router (can happen in tests)
	if h.router == nil {
		return c.JSON(fiber.Map{
			"branch":   req.Branch,
			"previous": "main",
			"message":  "Active branch set successfully",
		})
	}

	// Get previous branch for response
	previous := h.router.GetDefaultBranch()

	// Set the active branch
	h.router.SetActiveBranch(req.Branch)

	return c.JSON(fiber.Map{
		"branch":   req.Branch,
		"previous": previous,
		"message":  "Active branch set successfully",
	})
}

// ResetActiveBranch handles DELETE /admin/branches/active
func (h *BranchHandler) ResetActiveBranch(c fiber.Ctx) error {
	// Nil check for router (can happen in tests)
	if h.router == nil {
		return c.JSON(fiber.Map{
			"branch":   "main",
			"previous": "main",
			"message":  "Active branch reset to default",
		})
	}

	// Get current branch for response
	previous := h.router.GetDefaultBranch()

	// Reset to default (empty string clears API-set value)
	h.router.SetActiveBranch("")

	// Get new default branch
	newBranch := h.router.GetDefaultBranch()

	return c.JSON(fiber.Map{
		"branch":   newBranch,
		"previous": previous,
		"message":  "Active branch reset to default",
	})
}

// GitHub Config handlers

// ListGitHubConfigs handles GET /admin/branches/github/configs
func (h *BranchHandler) ListGitHubConfigs(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.JSON(fiber.Map{
			"configs": []*branching.GitHubConfig{},
		})
	}

	configs, err := h.manager.GetStorage().ListGitHubConfigs(c.RequestCtx())
	if err != nil {
		log.Error().Err(err).Msg("Failed to list GitHub configs")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list GitHub configurations",
		})
	}

	return c.JSON(fiber.Map{
		"configs": configs,
	})
}

// UpsertGitHubConfigRequest represents the request for creating/updating GitHub config
type UpsertGitHubConfigRequest struct {
	Repository           string                  `json:"repository" validate:"required"`
	AutoCreateOnPR       *bool                   `json:"auto_create_on_pr,omitempty"`
	AutoDeleteOnMerge    *bool                   `json:"auto_delete_on_merge,omitempty"`
	DefaultDataCloneMode branching.DataCloneMode `json:"default_data_clone_mode,omitempty"`
	WebhookSecret        *string                 `json:"webhook_secret,omitempty"`
}

// UpsertGitHubConfig handles POST /admin/branches/github/configs
func (h *BranchHandler) UpsertGitHubConfig(c fiber.Ctx) error {
	var req UpsertGitHubConfigRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Failed to parse request body: " + err.Error(),
		})
	}

	if req.Repository == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Repository is required",
		})
	}

	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	config := &branching.GitHubConfig{
		Repository:           req.Repository,
		AutoCreateOnPR:       true, // Default
		AutoDeleteOnMerge:    true, // Default
		DefaultDataCloneMode: branching.DataCloneModeSchemaOnly,
	}

	if req.AutoCreateOnPR != nil {
		config.AutoCreateOnPR = *req.AutoCreateOnPR
	}
	if req.AutoDeleteOnMerge != nil {
		config.AutoDeleteOnMerge = *req.AutoDeleteOnMerge
	}
	if req.DefaultDataCloneMode != "" {
		config.DefaultDataCloneMode = req.DefaultDataCloneMode
	}
	if req.WebhookSecret != nil {
		config.WebhookSecret = req.WebhookSecret
	}

	if err := h.manager.GetStorage().UpsertGitHubConfig(c.RequestCtx(), config); err != nil {
		log.Error().Err(err).Str("repository", req.Repository).Msg("Failed to upsert GitHub config")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "upsert_failed",
			"message": "Failed to save GitHub configuration",
		})
	}

	return c.Status(fiber.StatusOK).JSON(config)
}

// DeleteGitHubConfig handles DELETE /admin/branches/github/configs/:repository
func (h *BranchHandler) DeleteGitHubConfig(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	repository := c.Params("repository")

	if err := h.manager.GetStorage().DeleteGitHubConfig(c.RequestCtx(), repository); err != nil {
		if errors.Is(err, branching.ErrGitHubConfigNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "config_not_found",
				"message": "GitHub configuration not found",
			})
		}
		log.Error().Err(err).Str("repository", repository).Msg("Failed to delete GitHub config")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "delete_failed",
			"message": "Failed to delete GitHub configuration",
		})
	}

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// Access Management Handlers

// ListBranchAccess handles GET /admin/branches/:id/access
func (h *BranchHandler) ListBranchAccess(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first, then as slug
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	// Check authorization
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			userID = &id
		}
	}

	authType, _ := c.Locals("auth_type").(string)
	userRole, _ := c.Locals("user_role").(string)
	isAdmin := authType == "service_key" || userRole == "instance_admin" || userRole == "admin"

	if !isAdmin && userID != nil {
		hasAccess, err := h.manager.GetStorage().HasAccess(c.RequestCtx(), branch.ID, *userID, branching.BranchAccessAdmin)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "access_check_failed",
				"message": "Failed to verify branch access",
			})
		}
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "access_denied",
				"message": "You do not have permission to view access grants for this branch",
			})
		}
	}

	accessList, err := h.manager.GetStorage().GetBranchAccessList(c.RequestCtx(), branch.ID)
	if err != nil {
		log.Error().Err(err).Str("branch_id", branch.ID.String()).Msg("Failed to list branch access")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "list_failed",
			"message": "Failed to list branch access",
		})
	}

	return c.JSON(fiber.Map{
		"access": accessList,
	})
}

// GrantBranchAccessRequest represents the request body for granting access
type GrantBranchAccessRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	AccessLevel string `json:"access_level" validate:"required,oneof=read write admin"`
}

// GrantBranchAccess handles POST /admin/branches/:id/access
func (h *BranchHandler) GrantBranchAccess(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")

	// Try to parse as UUID first, then as slug
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	// Parse request body
	var req GrantBranchAccessRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Failed to parse request body: " + err.Error(),
		})
	}

	if req.UserID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "user_id is required",
		})
	}

	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Invalid user_id format",
		})
	}

	accessLevel := branching.BranchAccessLevel(req.AccessLevel)
	if accessLevel != branching.BranchAccessRead &&
		accessLevel != branching.BranchAccessWrite &&
		accessLevel != branching.BranchAccessAdmin {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "access_level must be one of: read, write, admin",
		})
	}

	// Check authorization
	var grantedBy *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			grantedBy = &id
		}
	}

	authType, _ := c.Locals("auth_type").(string)
	userRole, _ := c.Locals("user_role").(string)
	isAdmin := authType == "service_key" || userRole == "instance_admin" || userRole == "admin"

	if !isAdmin && grantedBy != nil {
		hasAccess, err := h.manager.GetStorage().HasAccess(c.RequestCtx(), branch.ID, *grantedBy, branching.BranchAccessAdmin)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "access_check_failed",
				"message": "Failed to verify branch access",
			})
		}
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "access_denied",
				"message": "You do not have permission to grant access to this branch",
			})
		}
	}

	// Grant access
	access := &branching.BranchAccess{
		ID:          uuid.New(),
		BranchID:    branch.ID,
		UserID:      targetUserID,
		AccessLevel: accessLevel,
		GrantedBy:   grantedBy,
	}

	if err := h.manager.GetStorage().GrantAccess(c.RequestCtx(), access); err != nil {
		log.Error().Err(err).
			Str("branch_id", branch.ID.String()).
			Str("user_id", targetUserID.String()).
			Msg("Failed to grant branch access")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "grant_failed",
			"message": "Failed to grant access",
		})
	}

	// Log activity
	_ = h.manager.GetStorage().LogActivity(c.RequestCtx(), &branching.ActivityLog{
		BranchID:   branch.ID,
		Action:     branching.ActivityActionAccessGranted,
		Status:     branching.ActivityStatusSuccess,
		ExecutedBy: grantedBy,
		Details: map[string]any{
			"user_id":      targetUserID.String(),
			"access_level": string(accessLevel),
		},
	})

	return c.Status(fiber.StatusCreated).JSON(access)
}

// RevokeBranchAccess handles DELETE /admin/branches/:id/access/:user_id
func (h *BranchHandler) RevokeBranchAccess(c fiber.Ctx) error {
	// Nil check for manager (can happen in tests)
	if h.manager == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "not_initialized",
			"message": "Branch manager not initialized",
		})
	}

	idParam := c.Params("id")
	userIDParam := c.Params("user_id")

	// Try to parse as UUID first, then as slug
	var branch *branching.Branch
	var err error

	if id, parseErr := uuid.Parse(idParam); parseErr == nil {
		branch, err = h.manager.GetStorage().GetBranch(c.RequestCtx(), id)
	} else {
		branch, err = h.manager.GetStorage().GetBranchBySlug(c.RequestCtx(), idParam)
	}

	if err != nil {
		if errors.Is(err, branching.ErrBranchNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "branch_not_found",
				"message": "Branch not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "get_failed",
			"message": "Failed to get branch",
		})
	}

	targetUserID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Invalid user_id format",
		})
	}

	// Check authorization
	var currentUserID *uuid.UUID
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			currentUserID = &id
		}
	}

	authType, _ := c.Locals("auth_type").(string)
	userRole, _ := c.Locals("user_role").(string)
	isAdmin := authType == "service_key" || userRole == "instance_admin" || userRole == "admin"

	if !isAdmin && currentUserID != nil {
		hasAccess, err := h.manager.GetStorage().HasAccess(c.RequestCtx(), branch.ID, *currentUserID, branching.BranchAccessAdmin)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "access_check_failed",
				"message": "Failed to verify branch access",
			})
		}
		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "access_denied",
				"message": "You do not have permission to revoke access from this branch",
			})
		}
	}

	// Revoke access
	if err := h.manager.GetStorage().RevokeAccess(c.RequestCtx(), branch.ID, targetUserID); err != nil {
		log.Error().Err(err).
			Str("branch_id", branch.ID.String()).
			Str("user_id", targetUserID.String()).
			Msg("Failed to revoke branch access")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "revoke_failed",
			"message": "Failed to revoke access",
		})
	}

	// Log activity
	_ = h.manager.GetStorage().LogActivity(c.RequestCtx(), &branching.ActivityLog{
		BranchID:   branch.ID,
		Action:     branching.ActivityActionAccessRevoked,
		Status:     branching.ActivityStatusSuccess,
		ExecutedBy: currentUserID,
		Details: map[string]any{
			"user_id": targetUserID.String(),
		},
	})

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// fiber:context-methods migrated
