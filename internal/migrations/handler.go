package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/nimbleflux/fluxbase/internal/util"
)

type TenantPoolProvider interface {
	GetPool(tenantID string) (*pgxpool.Pool, error)
}

type Handler struct {
	storage            *Storage
	executor           *Executor
	schemaCache        *database.SchemaCache
	tenantPoolProvider TenantPoolProvider
	db                 *database.Connection
}

func NewHandler(db *database.Connection, schemaCache *database.SchemaCache) *Handler {
	return &Handler{
		storage:     NewStorage(db),
		executor:    NewExecutor(db),
		schemaCache: schemaCache,
		db:          db,
	}
}

func (h *Handler) SetTenantPoolProvider(provider TenantPoolProvider) {
	h.tenantPoolProvider = provider
}

func (h *Handler) CreateMigration(c fiber.Ctx) error {
	var req struct {
		Namespace   string  `json:"namespace"`
		Name        string  `json:"name"`
		Description *string `json:"description"`
		UpSQL       string  `json:"up_sql"`
		DownSQL     *string `json:"down_sql"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}
	if req.Name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Migration name is required"})
	}
	if req.UpSQL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Up SQL is required"})
	}

	var createdBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				createdBy = &parsed
			}
		}
	}

	migration := &Migration{
		Namespace:   req.Namespace,
		Name:        req.Name,
		Description: req.Description,
		UpSQL:       req.UpSQL,
		DownSQL:     req.DownSQL,
		CreatedBy:   createdBy,
	}

	if err := h.storage.CreateMigration(c.RequestCtx(), migration); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return c.Status(409).JSON(fiber.Map{
				"error": fmt.Sprintf("Migration '%s' already exists in namespace '%s'", req.Name, req.Namespace),
			})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create migration", "details": err.Error()})
	}

	return c.Status(201).JSON(migration)
}

func (h *Handler) GetMigration(c fiber.Ctx) error {
	name := c.Params("name")
	namespace := c.Query("namespace", "default")

	migration, err := h.storage.GetMigration(c.RequestCtx(), namespace, name)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Migration not found"})
	}

	return c.JSON(migration)
}

func (h *Handler) ListMigrations(c fiber.Ctx) error {
	namespace := c.Query("namespace", "default")
	status := c.Query("status")

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	migrations, err := h.storage.ListMigrations(c.RequestCtx(), namespace, statusPtr)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list migrations", "details": err.Error()})
	}

	return c.JSON(migrations)
}

func (h *Handler) UpdateMigration(c fiber.Ctx) error {
	name := c.Params("name")
	namespace := c.Query("namespace", "default")

	var updates map[string]interface{}
	if err := c.Bind().Body(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if err := h.storage.UpdateMigration(c.RequestCtx(), namespace, name, updates); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already applied") {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update migration", "details": err.Error()})
	}

	migration, err := h.storage.GetMigration(c.RequestCtx(), namespace, name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Migration updated but failed to retrieve"})
	}

	return c.JSON(migration)
}

func (h *Handler) DeleteMigration(c fiber.Ctx) error {
	name := c.Params("name")
	namespace := c.Query("namespace", "default")

	if err := h.storage.DeleteMigration(c.RequestCtx(), namespace, name); err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "already applied") {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete migration", "details": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Migration deleted successfully"})
}

func (h *Handler) ApplyMigration(c fiber.Ctx) error {
	name := c.Params("name")

	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := c.Bind().Body(&req); err != nil {
		req.Namespace = "default"
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var executedBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				executedBy = &parsed
			}
		}
	}

	isTenantMigration, _ := c.Locals("is_tenant_migration").(bool)
	tenantID := middleware.GetTenantID(c)

	var err error
	if isTenantMigration && h.tenantPoolProvider != nil && tenantID != "" {
		err = h.applyTenantMigration(c.RequestCtx(), req.Namespace, name, tenantID, executedBy)
	} else {
		err = h.executor.ApplyMigration(c.RequestCtx(), req.Namespace, name, executedBy)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to apply migration", "details": err.Error()})
	}

	if h.schemaCache != nil {
		h.schemaCache.InvalidateAll(c.RequestCtx())
		log.Debug().Str("migration", name).Msg("Schema cache invalidated after applying migration")
	}

	return c.JSON(fiber.Map{"message": "Migration applied successfully"})
}

func (h *Handler) RollbackMigration(c fiber.Ctx) error {
	name := c.Params("name")

	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := c.Bind().Body(&req); err != nil {
		req.Namespace = "default"
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var executedBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				executedBy = &parsed
			}
		}
	}

	isTenantMigration, _ := c.Locals("is_tenant_migration").(bool)
	tenantID := middleware.GetTenantID(c)

	var err error
	if isTenantMigration && h.tenantPoolProvider != nil && tenantID != "" {
		err = h.rollbackTenantMigration(c.RequestCtx(), req.Namespace, name, tenantID, executedBy)
	} else {
		err = h.executor.RollbackMigration(c.RequestCtx(), req.Namespace, name, executedBy)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to rollback migration", "details": err.Error()})
	}

	if h.schemaCache != nil {
		h.schemaCache.InvalidateAll(c.RequestCtx())
		log.Debug().Str("migration", name).Msg("Schema cache invalidated after rolling back migration")
	}

	return c.JSON(fiber.Map{"message": "Migration rolled back successfully"})
}

func (h *Handler) applyTenantMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	pool, err := h.tenantPoolProvider.GetPool(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	tenantExec := NewTenantExecutor(h.db, pool)
	return tenantExec.ApplyMigration(ctx, namespace, name, tenantID, executedBy)
}

func (h *Handler) rollbackTenantMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	pool, err := h.tenantPoolProvider.GetPool(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	tenantExec := NewTenantExecutor(h.db, pool)
	return tenantExec.RollbackMigration(ctx, namespace, name, tenantID, executedBy)
}

func (h *Handler) ApplyPending(c fiber.Ctx) error {
	var req struct {
		Namespace string `json:"namespace"`
	}
	if err := c.Bind().Body(&req); err != nil {
		req.Namespace = "default"
	}
	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var executedBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				executedBy = &parsed
			}
		}
	}

	applied, failed, err := h.executor.ApplyPendingMigrations(c.RequestCtx(), req.Namespace, executedBy)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to apply pending migrations",
			"details": err.Error(),
			"applied": applied,
			"failed":  failed,
		})
	}

	if len(applied) > 0 && h.schemaCache != nil {
		h.schemaCache.InvalidateAll(c.RequestCtx())
		log.Debug().Int("count", len(applied)).Msg("Schema cache invalidated after applying pending migrations")
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("Applied %d migrations successfully", len(applied)),
		"applied": applied,
		"failed":  failed,
	})
}

func (h *Handler) GetExecutions(c fiber.Ctx) error {
	name := c.Params("name")
	namespace := c.Query("namespace", "default")
	limit := fiber.Query[int](c, "limit", 50)

	if limit > 100 {
		limit = 100
	}

	migration, err := h.storage.GetMigration(c.RequestCtx(), namespace, name)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Migration not found"})
	}

	logs, err := h.storage.GetExecutionLogs(c.RequestCtx(), migration.ID, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get execution logs", "details": err.Error()})
	}

	return c.JSON(logs)
}

func (h *Handler) SyncMigrations(c fiber.Ctx) error {
	var req struct {
		Namespace  string `json:"namespace"`
		Migrations []struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
			UpSQL       string  `json:"up_sql"`
			DownSQL     *string `json:"down_sql"`
		} `json:"migrations"`
		Options struct {
			UpdateIfChanged bool `json:"update_if_changed"`
			AutoApply       bool `json:"auto_apply"`
			DryRun          bool `json:"dry_run"`
		} `json:"options"`
	}

	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	var createdBy *uuid.UUID
	if userID := c.Locals("user_id"); userID != nil {
		if uid, ok := userID.(string); ok {
			parsed, err := uuid.Parse(uid)
			if err == nil {
				createdBy = &parsed
			}
		}
	}

	existing, err := h.storage.ListMigrations(c.RequestCtx(), req.Namespace, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to list existing migrations"})
	}

	existingMap := make(map[string]*Migration)
	for i := range existing {
		existingMap[existing[i].Name] = &existing[i]
	}

	summary := struct {
		Created   int `json:"created"`
		Updated   int `json:"updated"`
		Unchanged int `json:"unchanged"`
		Skipped   int `json:"skipped"`
		Applied   int `json:"applied"`
		Errors    int `json:"errors"`
	}{}

	details := struct {
		Created   []string `json:"created"`
		Updated   []string `json:"updated"`
		Unchanged []string `json:"unchanged"`
		Skipped   []string `json:"skipped"`
		Applied   []string `json:"applied"`
		Errors    []string `json:"errors"`
	}{}

	warnings := []string{}
	autoApplyFailed := false

	for _, reqMig := range req.Migrations {
		if autoApplyFailed {
			summary.Skipped++
			details.Skipped = append(details.Skipped, reqMig.Name)
			warnings = append(warnings, fmt.Sprintf("Migration '%s' skipped due to previous failure", reqMig.Name))
			continue
		}

		contentHash := calculateHash(reqMig.UpSQL + util.ValueOr(reqMig.DownSQL, ""))

		existingMig, exists := existingMap[reqMig.Name]

		if !exists {
			if !req.Options.DryRun {
				newMig := &Migration{
					Namespace:   req.Namespace,
					Name:        reqMig.Name,
					Description: reqMig.Description,
					UpSQL:       reqMig.UpSQL,
					DownSQL:     reqMig.DownSQL,
					CreatedBy:   createdBy,
				}
				if err := h.storage.CreateMigration(c.RequestCtx(), newMig); err != nil {
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: %v", reqMig.Name, err))
					continue
				}
			}
			summary.Created++
			details.Created = append(details.Created, reqMig.Name)

			if req.Options.AutoApply && !req.Options.DryRun {
				if err := h.executor.ApplyMigration(c.RequestCtx(), req.Namespace, reqMig.Name, createdBy); err != nil {
					log.Error().Err(err).Str("name", reqMig.Name).Msg("Failed to auto-apply migration")
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: failed to apply - %v", reqMig.Name, err))
					autoApplyFailed = true
				} else {
					summary.Applied++
					details.Applied = append(details.Applied, reqMig.Name)
				}
			}
			continue
		}

		existingHash := calculateHash(existingMig.UpSQL + util.ValueOr(existingMig.DownSQL, ""))

		if existingHash == contentHash {
			if (existingMig.Status == "pending" || existingMig.Status == "failed") && req.Options.AutoApply && !req.Options.DryRun {
				action := "Applying"
				if existingMig.Status == "failed" {
					action = "Retrying"
				}
				log.Info().Str("name", reqMig.Name).Str("status", existingMig.Status).Msg(action + " migration")
				if err := h.executor.ApplyMigration(c.RequestCtx(), req.Namespace, reqMig.Name, createdBy); err != nil {
					log.Error().Err(err).Str("name", reqMig.Name).Msg("Failed to apply migration")
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: failed to apply - %v", reqMig.Name, err))
					autoApplyFailed = true
				} else {
					summary.Applied++
					details.Applied = append(details.Applied, reqMig.Name)
				}
				continue
			}

			summary.Unchanged++
			details.Unchanged = append(details.Unchanged, reqMig.Name)
			continue
		}

		if existingMig.Status == "pending" && req.Options.UpdateIfChanged {
			if !req.Options.DryRun {
				updates := map[string]interface{}{
					"up_sql":   reqMig.UpSQL,
					"down_sql": reqMig.DownSQL,
				}
				if reqMig.Description != nil {
					updates["description"] = *reqMig.Description
				}
				if err := h.storage.UpdateMigration(c.RequestCtx(), req.Namespace, reqMig.Name, updates); err != nil {
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: %v", reqMig.Name, err))
					continue
				}
			}
			summary.Updated++
			details.Updated = append(details.Updated, reqMig.Name)
			continue
		}

		if existingMig.Status == "applied" {
			summary.Skipped++
			details.Skipped = append(details.Skipped, reqMig.Name)
			warnings = append(warnings, fmt.Sprintf("Migration '%s' already applied with different content (skipped)", reqMig.Name))
			continue
		}

		if existingMig.Status == "failed" && req.Options.UpdateIfChanged {
			if !req.Options.DryRun {
				updates := map[string]interface{}{
					"up_sql":   reqMig.UpSQL,
					"down_sql": reqMig.DownSQL,
					"status":   "pending",
				}
				if reqMig.Description != nil {
					updates["description"] = *reqMig.Description
				}
				if err := h.storage.UpdateMigration(c.RequestCtx(), req.Namespace, reqMig.Name, updates); err != nil {
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: %v", reqMig.Name, err))
					continue
				}
			}
			summary.Updated++
			details.Updated = append(details.Updated, reqMig.Name)

			if req.Options.AutoApply && !req.Options.DryRun {
				log.Info().Str("name", reqMig.Name).Msg("Retrying updated failed migration")
				if err := h.executor.ApplyMigration(c.RequestCtx(), req.Namespace, reqMig.Name, createdBy); err != nil {
					log.Error().Err(err).Str("name", reqMig.Name).Msg("Failed to apply updated migration")
					summary.Errors++
					details.Errors = append(details.Errors, fmt.Sprintf("%s: failed to apply after update - %v", reqMig.Name, err))
					autoApplyFailed = true
				} else {
					summary.Applied++
					details.Applied = append(details.Applied, reqMig.Name)
				}
			}
			continue
		}

		summary.Skipped++
		details.Skipped = append(details.Skipped, reqMig.Name)
		warnings = append(warnings, fmt.Sprintf("Migration '%s' has status '%s' (skipped)", reqMig.Name, existingMig.Status))
	}

	if summary.Applied > 0 && h.schemaCache != nil {
		h.schemaCache.InvalidateAll(c.RequestCtx())
		log.Info().Int("applied", summary.Applied).Msg("Schema cache invalidated after sync")
	}

	message := fmt.Sprintf("Sync complete: %d created, %d updated, %d unchanged", summary.Created, summary.Updated, summary.Unchanged)
	if summary.Errors > 0 {
		message = fmt.Sprintf("Sync completed with errors: %d created, %d updated, %d unchanged, %d errors", summary.Created, summary.Updated, summary.Unchanged, summary.Errors)
	}

	response := fiber.Map{
		"message":   message,
		"namespace": req.Namespace,
		"summary":   summary,
		"details":   details,
		"dry_run":   req.Options.DryRun,
	}

	if len(warnings) > 0 {
		response["warnings"] = warnings
	}

	if summary.Errors > 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response)
	}

	return c.JSON(response)
}

func calculateHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}
