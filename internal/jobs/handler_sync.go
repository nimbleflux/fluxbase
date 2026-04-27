package jobs

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	syncframework "github.com/nimbleflux/fluxbase/internal/sync"

	apperrors "github.com/nimbleflux/fluxbase/internal/errors"
)

// SyncJobs syncs job functions to a namespace
// Accepts a batch of job functions with optional delete_missing to remove stale jobs
// Admin-only endpoint - requires authentication and admin role
func (h *Handler) SyncJobs(c fiber.Ctx) error {
	var req struct {
		Namespace string `json:"namespace"`
		Jobs      []struct {
			Name                   string   `json:"name"`
			Description            *string  `json:"description"`
			Code                   string   `json:"code"`
			OriginalCode           *string  `json:"original_code"`
			IsBundled              *bool    `json:"is_bundled"` // If true, skip server-side bundling
			Enabled                *bool    `json:"enabled"`
			Schedule               *string  `json:"schedule"`
			TimeoutSeconds         *int     `json:"timeout_seconds"`
			MemoryLimitMB          *int     `json:"memory_limit_mb"`
			MaxRetries             *int     `json:"max_retries"`
			ProgressTimeoutSeconds *int     `json:"progress_timeout_seconds"`
			AllowNet               *bool    `json:"allow_net"`
			AllowEnv               *bool    `json:"allow_env"`
			AllowRead              *bool    `json:"allow_read"`
			AllowWrite             *bool    `json:"allow_write"`
			RequireRoles           []string `json:"require_roles"`
		} `json:"jobs"`
		Options struct {
			DeleteMissing bool `json:"delete_missing"`
			DryRun        bool `json:"dry_run"`
		} `json:"options"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Default namespace to "default" if not specified
	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	ctx := middleware.CtxWithTenant(c)

	syncCtx := database.ContextWithTenant(ctx, "")
	currentTenantID := database.TenantFromContext(ctx)

	var createdBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				createdBy = &parsed
			}
		}
	}

	// If no jobs provided, fall back to filesystem sync
	if len(req.Jobs) == 0 {
		if err := h.loader.LoadFromFilesystem(ctx, namespace); err != nil {
			reqID := apperrors.GetRequestID(c)
			log.Error().
				Err(err).
				Str("namespace", namespace).
				Str("request_id", reqID).
				Msg("Failed to sync jobs from filesystem")

			return c.Status(500).JSON(fiber.Map{
				"error":      "Failed to sync jobs from filesystem",
				"details":    err.Error(),
				"request_id": reqID,
			})
		}

		// Reschedule jobs after filesystem sync
		h.rescheduleJobsFromNamespace(ctx, namespace)

		return c.JSON(fiber.Map{
			"message":   "Jobs synced from filesystem",
			"namespace": namespace,
			"summary": fiber.Map{
				"created":   0,
				"updated":   0,
				"deleted":   0,
				"unchanged": 0,
				"errors":    0,
			},
			"details": fiber.Map{
				"created":   []string{},
				"updated":   []string{},
				"deleted":   []string{},
				"unchanged": []string{},
			},
			"errors":  []fiber.Map{},
			"dry_run": false,
		})
	}

	items := make([]jobSyncItem, len(req.Jobs))
	for i, spec := range req.Jobs {
		items[i] = jobSyncItem{
			Name:                   spec.Name,
			Code:                   spec.Code,
			Description:            spec.Description,
			Enabled:                spec.Enabled,
			Schedule:               spec.Schedule,
			TimeoutSeconds:         spec.TimeoutSeconds,
			MemoryLimitMB:          spec.MemoryLimitMB,
			MaxRetries:             spec.MaxRetries,
			ProgressTimeoutSeconds: spec.ProgressTimeoutSeconds,
			AllowNet:               spec.AllowNet,
			AllowEnv:               spec.AllowEnv,
			AllowRead:              spec.AllowRead,
			AllowWrite:             spec.AllowWrite,
			RequireRoles:           spec.RequireRoles,
			IsBundled:              spec.IsBundled,
			OriginalCode:           spec.OriginalCode,
		}
	}

	var createdByStr string
	if createdBy != nil {
		createdByStr = createdBy.String()
	}

	syncer := newJobSyncer(h, syncCtx, namespace, currentTenantID, createdBy)
	result, syncErr := syncframework.Execute[jobSyncItem](ctx, syncer, items, syncframework.Options{
		Namespace:     namespace,
		DeleteMissing: req.Options.DeleteMissing,
		DryRun:        req.Options.DryRun,
		TenantID:      currentTenantID,
		CreatedBy:     createdByStr,
	})
	if syncErr != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": syncErr.Error(),
		})
	}

	return c.JSON(result)
}

// ListNamespaces lists all unique namespaces that have job functions (Admin only)
func (h *Handler) ListNamespaces(c fiber.Ctx) error {
	namespaces, err := h.storage.ListJobNamespaces(middleware.CtxWithTenant(c))
	if err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("request_id", reqID).
			Msg("Failed to list job namespaces")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to list job namespaces",
			"request_id": reqID,
		})
	}

	// Ensure we always return at least "default"
	if len(namespaces) == 0 {
		namespaces = []string{"default"}
	}

	// Normalize empty-string namespaces to "default" so the UI can present
	// them meaningfully and use the value in subsequent queries.
	for i := range namespaces {
		if namespaces[i] == "" {
			namespaces[i] = "default"
		}
	}

	return c.JSON(fiber.Map{"namespaces": namespaces})
}

// ListJobFunctions lists all job functions (Admin only)
func (h *Handler) ListJobFunctions(c fiber.Ctx) error {
	namespace := c.Query("namespace")
	if namespace == "default" {
		namespace = ""
	}

	var functions []*JobFunctionSummary
	var err error

	if namespace != "" {
		// If namespace is specified, list functions in that namespace
		functions, err = h.storage.ListJobFunctions(middleware.CtxWithTenant(c), namespace)
	} else {
		// Otherwise, list all job functions across all namespaces
		functions, err = h.storage.ListAllJobFunctions(middleware.CtxWithTenant(c))
	}

	if err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("request_id", reqID).
			Msg("Failed to list job functions")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to list job functions",
			"request_id": reqID,
		})
	}

	return c.JSON(functions)
}

// GetJobFunction gets a job function by namespace and name (Admin only)
func (h *Handler) GetJobFunction(c fiber.Ctx) error {
	namespace := c.Params("namespace")
	name := c.Params("name")

	function, err := h.storage.GetJobFunction(middleware.CtxWithTenant(c), namespace, name)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job function not found"})
	}

	return c.JSON(function)
}

// UpdateJobFunction updates a job function (Admin only)
func (h *Handler) UpdateJobFunction(c fiber.Ctx) error {
	namespace := c.Params("namespace")
	name := c.Params("name")

	// Get existing function
	fn, err := h.storage.GetJobFunction(middleware.CtxWithTenant(c), namespace, name)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Job function not found"})
	}

	// Parse update request
	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Apply updates
	if req.Enabled != nil {
		fn.Enabled = *req.Enabled
	}

	// Save changes
	if err := h.storage.UpdateJobFunction(middleware.CtxWithTenant(c), fn); err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("namespace", namespace).
			Str("name", name).
			Str("request_id", reqID).
			Msg("Failed to update job function")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to update job function",
			"request_id": reqID,
		})
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Bool("enabled", fn.Enabled).
		Msg("Job function updated")

	return c.JSON(fn)
}

// DeleteJobFunction deletes a job function (Admin only)
func (h *Handler) DeleteJobFunction(c fiber.Ctx) error {
	namespace := c.Params("namespace")
	name := c.Params("name")

	if err := h.storage.DeleteJobFunction(middleware.CtxWithTenant(c), namespace, name); err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("namespace", namespace).
			Str("name", name).
			Str("request_id", reqID).
			Msg("Failed to delete job function")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to delete job function",
			"request_id": reqID,
		})
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Msg("Job function deleted")

	return c.SendStatus(204)
}

// GetJobStats returns job statistics (Admin only)
func (h *Handler) GetJobStats(c fiber.Ctx) error {
	var namespacePtr *string
	if namespace := c.Query("namespace"); namespace != "" {
		if namespace == "default" {
			namespace = ""
		}
		namespacePtr = &namespace
	}

	stats, err := h.storage.GetJobStats(middleware.CtxWithTenant(c), namespacePtr)
	if err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("request_id", reqID).
			Msg("Failed to get job stats")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to get job stats",
			"request_id": reqID,
		})
	}

	return c.JSON(stats)
}

// ListWorkers lists all workers (Admin only)
func (h *Handler) ListWorkers(c fiber.Ctx) error {
	workers, err := h.storage.ListWorkers(middleware.CtxWithTenant(c))
	if err != nil {
		reqID := apperrors.GetRequestID(c)
		log.Error().
			Err(err).
			Str("request_id", reqID).
			Msg("Failed to list workers")

		return c.Status(500).JSON(fiber.Map{
			"error":      "Failed to list workers",
			"request_id": reqID,
		})
	}

	return c.JSON(workers)
}

// LoadFromFilesystem loads jobs from filesystem at boot time
func (h *Handler) LoadFromFilesystem(ctx context.Context, namespace string) error {
	// Load builtin jobs first (these ship with Fluxbase and are disabled by default)
	if err := h.loader.LoadBuiltinJobs(ctx, namespace); err != nil {
		log.Warn().Err(err).Msg("Failed to load builtin jobs")
		// Don't fail boot if builtin jobs fail to load
	}

	// Then load user jobs from filesystem
	if err := h.loader.LoadFromFilesystem(ctx, namespace); err != nil {
		return err
	}

	// Reschedule jobs after loading
	if h.scheduler != nil {
		h.rescheduleJobsFromNamespace(ctx, namespace)
	}

	return nil
}

// rescheduleJobsFromNamespace updates the scheduler with jobs from a namespace
func (h *Handler) rescheduleJobsFromNamespace(ctx context.Context, namespace string) {
	if h.scheduler == nil {
		return
	}

	jobs, err := h.storage.ListJobFunctions(ctx, namespace)
	if err != nil {
		log.Warn().Err(err).Str("namespace", namespace).Msg("Failed to list jobs for rescheduling")
		return
	}

	for _, job := range jobs {
		if job.Enabled && job.Schedule != nil && *job.Schedule != "" {
			if err := h.scheduler.ScheduleJob(job); err != nil {
				log.Warn().Err(err).Str("job", job.Name).Msg("Failed to schedule job")
			}
		} else {
			h.scheduler.UnscheduleJob(namespace, job.Name)
		}
	}
}
