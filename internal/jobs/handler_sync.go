package jobs

import (
	"context"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
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
			reqID := getRequestID(c)
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

	// Get all existing job functions in this namespace
	existingFunctions, err := h.storage.ListJobFunctionsForSync(syncCtx, namespace, currentTenantID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to list existing job functions in namespace",
		})
	}

	// Build set of existing function names
	existingNames := make(map[string]*JobFunctionSummary)
	for i := range existingFunctions {
		existingNames[existingFunctions[i].Name] = existingFunctions[i]
	}

	// Build set of payload function names
	payloadNames := make(map[string]bool)
	for _, spec := range req.Jobs {
		payloadNames[spec.Name] = true
	}

	// Determine operations
	toCreate := []string{}
	toUpdate := []string{}
	toDelete := []string{}

	for _, spec := range req.Jobs {
		if _, exists := existingNames[spec.Name]; exists {
			toUpdate = append(toUpdate, spec.Name)
		} else {
			toCreate = append(toCreate, spec.Name)
		}
	}

	if req.Options.DeleteMissing {
		for name := range existingNames {
			if !payloadNames[name] {
				toDelete = append(toDelete, name)
			}
		}
	}

	// Track results
	created := []string{}
	updated := []string{}
	deleted := []string{}
	unchanged := []string{}
	errorList := []fiber.Map{}

	// If dry run, return what would be done without making changes
	if req.Options.DryRun {
		return c.JSON(fiber.Map{
			"message":   "Dry run - no changes made",
			"namespace": namespace,
			"summary": fiber.Map{
				"created":   len(toCreate),
				"updated":   len(toUpdate),
				"deleted":   len(toDelete),
				"unchanged": 0,
				"errors":    0,
			},
			"details": fiber.Map{
				"created":   toCreate,
				"updated":   toUpdate,
				"deleted":   toDelete,
				"unchanged": []string{},
			},
			"errors":  []fiber.Map{},
			"dry_run": true,
		})
	}

	// Process each job function
	for _, spec := range req.Jobs {
		code := spec.Code
		originalCode := spec.Code
		isBundled := false
		var bundleError *string

		// If original_code provided, use it
		if spec.OriginalCode != nil {
			originalCode = *spec.OriginalCode
		}

		// If client sent pre-bundled code, skip server-side bundling
		if spec.IsBundled != nil && *spec.IsBundled {
			isBundled = true
		} else {
			// Bundle server-side
			bundledCode, bundleErr := h.loader.BundleCode(ctx, spec.Code)
			if bundleErr != nil {
				errMsg := bundleErr.Error()
				bundleError = &errMsg
				// Continue with unbundled code
			} else {
				code = bundledCode
				isBundled = true
			}
		}

		// Parse annotations from original code
		annotations := h.loader.ParseAnnotations(originalCode)

		// Create or update job function
		if existing, exists := existingNames[spec.Name]; exists {
			// Update existing function - build JobFunction with updated values
			updatedFn := &JobFunction{
				ID:                     existing.ID,
				Name:                   existing.Name,
				Namespace:              existing.Namespace,
				Code:                   &code,
				OriginalCode:           &originalCode,
				IsBundled:              isBundled,
				BundleError:            bundleError,
				Description:            existing.Description,
				Enabled:                existing.Enabled,
				Schedule:               existing.Schedule,
				TimeoutSeconds:         existing.TimeoutSeconds,
				MemoryLimitMB:          existing.MemoryLimitMB,
				MaxRetries:             existing.MaxRetries,
				ProgressTimeoutSeconds: existing.ProgressTimeoutSeconds,
				AllowNet:               existing.AllowNet,
				AllowEnv:               existing.AllowEnv,
				AllowRead:              existing.AllowRead,
				AllowWrite:             existing.AllowWrite,
				RequireRoles:           existing.RequireRoles,
				Source:                 existing.Source, // Preserve original source
			}

			// Apply request values (take precedence over annotations)
			if spec.Description != nil {
				updatedFn.Description = spec.Description
			}
			if spec.Enabled != nil {
				updatedFn.Enabled = *spec.Enabled
			}
			if spec.Schedule != nil {
				updatedFn.Schedule = spec.Schedule
			}
			if spec.TimeoutSeconds != nil {
				updatedFn.TimeoutSeconds = *spec.TimeoutSeconds
			} else if annotations.TimeoutSeconds > 0 {
				updatedFn.TimeoutSeconds = annotations.TimeoutSeconds
			}
			if spec.MemoryLimitMB != nil {
				updatedFn.MemoryLimitMB = *spec.MemoryLimitMB
			} else if annotations.MemoryLimitMB > 0 {
				updatedFn.MemoryLimitMB = annotations.MemoryLimitMB
			}
			if spec.MaxRetries != nil {
				updatedFn.MaxRetries = *spec.MaxRetries
			} else if annotations.MaxRetries > 0 {
				updatedFn.MaxRetries = annotations.MaxRetries
			}
			if spec.ProgressTimeoutSeconds != nil {
				updatedFn.ProgressTimeoutSeconds = *spec.ProgressTimeoutSeconds
			} else if annotations.ProgressTimeoutSeconds > 0 {
				updatedFn.ProgressTimeoutSeconds = annotations.ProgressTimeoutSeconds
			}
			if spec.AllowNet != nil {
				updatedFn.AllowNet = *spec.AllowNet
			}
			if spec.AllowEnv != nil {
				updatedFn.AllowEnv = *spec.AllowEnv
			}
			if spec.AllowRead != nil {
				updatedFn.AllowRead = *spec.AllowRead
			}
			if spec.AllowWrite != nil {
				updatedFn.AllowWrite = *spec.AllowWrite
			}
			if len(spec.RequireRoles) > 0 {
				updatedFn.RequireRoles = spec.RequireRoles
			}

			if err := h.storage.UpdateJobFunctionForSync(syncCtx, currentTenantID, updatedFn); err != nil {
				errorList = append(errorList, fiber.Map{
					"job":    spec.Name,
					"error":  err.Error(),
					"action": "update",
				})
				continue
			}
			updated = append(updated, spec.Name)
		} else {
			// Create new function
			fn := &JobFunction{
				ID:                     uuid.New(),
				Name:                   spec.Name,
				Namespace:              namespace,
				Description:            spec.Description,
				Code:                   &code,
				OriginalCode:           &originalCode,
				IsBundled:              isBundled,
				BundleError:            bundleError,
				Enabled:                valueOr(spec.Enabled, true),
				Schedule:               spec.Schedule,
				TimeoutSeconds:         valueOr(spec.TimeoutSeconds, valueOr(&annotations.TimeoutSeconds, 300)),
				MemoryLimitMB:          valueOr(spec.MemoryLimitMB, valueOr(&annotations.MemoryLimitMB, 256)),
				MaxRetries:             valueOr(spec.MaxRetries, annotations.MaxRetries),
				ProgressTimeoutSeconds: valueOr(spec.ProgressTimeoutSeconds, valueOr(&annotations.ProgressTimeoutSeconds, 60)),
				AllowNet:               valueOr(spec.AllowNet, true),
				AllowEnv:               valueOr(spec.AllowEnv, true),
				AllowRead:              valueOr(spec.AllowRead, false),
				AllowWrite:             valueOr(spec.AllowWrite, false),
				RequireRoles:           spec.RequireRoles,
				Version:                1,
				CreatedBy:              createdBy,
				Source:                 "api",
			}

			if err := h.storage.CreateJobFunction(ctx, fn); err != nil {
				errorList = append(errorList, fiber.Map{
					"job":    spec.Name,
					"error":  err.Error(),
					"action": "create",
				})
				continue
			}
			created = append(created, spec.Name)
		}
	}

	// Delete removed job functions (after successful creates/updates for safety)
	if req.Options.DeleteMissing {
		for _, name := range toDelete {
			if err := h.storage.DeleteJobFunctionForSync(syncCtx, currentTenantID, namespace, name); err != nil {
				errorList = append(errorList, fiber.Map{
					"job":    name,
					"error":  err.Error(),
					"action": "delete",
				})
				continue
			}
			deleted = append(deleted, name)
		}
	}

	log.Info().
		Str("namespace", namespace).
		Int("created", len(created)).
		Int("updated", len(updated)).
		Int("deleted", len(deleted)).
		Int("unchanged", len(unchanged)).
		Int("errors", len(errorList)).
		Msg("Jobs synced successfully")

	// Reschedule jobs after sync
	h.rescheduleJobsFromNamespace(ctx, namespace)

	return c.JSON(fiber.Map{
		"message":   "Jobs synced successfully",
		"namespace": namespace,
		"summary": fiber.Map{
			"created":   len(created),
			"updated":   len(updated),
			"deleted":   len(deleted),
			"unchanged": len(unchanged),
			"errors":    len(errorList),
		},
		"details": fiber.Map{
			"created":   created,
			"updated":   updated,
			"deleted":   deleted,
			"unchanged": unchanged,
		},
		"errors":  errorList,
		"dry_run": false,
	})
}

// ListNamespaces lists all unique namespaces that have job functions (Admin only)
func (h *Handler) ListNamespaces(c fiber.Ctx) error {
	namespaces, err := h.storage.ListJobNamespaces(middleware.CtxWithTenant(c))
	if err != nil {
		reqID := getRequestID(c)
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
		reqID := getRequestID(c)
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
		reqID := getRequestID(c)
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
		reqID := getRequestID(c)
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
		reqID := getRequestID(c)
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
		reqID := getRequestID(c)
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
