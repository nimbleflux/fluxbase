package api

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/database/bootstrap"
	"github.com/nimbleflux/fluxbase/internal/migrations"
)

// InternalSchemaHandler handles internal schema management endpoints
type InternalSchemaHandler struct {
	declarative *migrations.DeclarativeService
	validator   *migrations.Validator
	transition  *migrations.TransitionService
	bootstrap   *bootstrap.Service
	pool        *pgxpool.Pool
	config      *migrations.DeclarativeConfig
}

// NewInternalSchemaHandler creates a new internal schema handler
func NewInternalSchemaHandler() *InternalSchemaHandler {
	return &InternalSchemaHandler{}
}

// Initialize initializes the handler with dependencies
func (h *InternalSchemaHandler) Initialize(cfg *config.Config, db *database.Connection) {
	h.pool = db.Pool()
	h.bootstrap = bootstrap.NewService(db.Pool())

	// Set up declarative config
	pgschemaPath := "pgschema"
	schemaDir := "internal/database/schema/schemas"

	h.config = &migrations.DeclarativeConfig{
		SchemaDir:        schemaDir,
		Schemas:          migrations.DefaultFluxbaseSchemas,
		AllowDestructive: false,
		LockTimeout:      30,
	}

	// Create services
	h.declarative = migrations.NewDeclarativeService(
		pgschemaPath,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		*h.config,
	)

	h.validator = migrations.NewValidator(h.declarative, db.Pool())
	h.transition = migrations.NewTransitionService(h.declarative, db.Pool())

	log.Info().
		Str("schema_dir", h.config.SchemaDir).
		Str("pgschema_path", pgschemaPath).
		Msg("Internal schema handler initialized")
}

// DumpSchema handles POST /api/v1/admin/internal-schema/dump
func (h *InternalSchemaHandler) DumpSchema(c fiber.Ctx) error {
	if h.declarative == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	var req struct {
		Dir    string `json:"dir"`
		Schema string `json:"schema"` // Optional: dump a specific schema
	}
	if err := c.Bind().Body(&req); err != nil && err != fiber.ErrUnprocessableEntity {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	schemaDir := h.config.SchemaDir
	if req.Dir != "" {
		schemaDir = req.Dir
	}

	// If a specific schema is requested, dump just that one
	if req.Schema != "" {
		outputPath := filepath.Join(schemaDir, req.Schema+".sql")
		if err := h.declarative.DumpForSchema(ctx, req.Schema, outputPath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		content, err := os.ReadFile(outputPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to read dumped schema"})
		}

		return c.JSON(fiber.Map{
			"message": "Schema dumped successfully",
			"schema":  req.Schema,
			"file":    outputPath,
			"sql":     string(content),
			"size":    len(content),
		})
	}

	// Dump all schemas
	if err := h.declarative.Dump(ctx, schemaDir); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "All schemas dumped successfully",
		"dir":     schemaDir,
	})
}

// PlanSchema handles POST /api/v1/admin/internal-schema/plan
func (h *InternalSchemaHandler) PlanSchema(c fiber.Ctx) error {
	if h.declarative == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	var req struct {
		Schema string `json:"schema"` // Optional: plan a specific schema
	}
	if err := c.Bind().Body(&req); err != nil && err != fiber.ErrUnprocessableEntity {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	var plan *migrations.Plan
	var err error

	if req.Schema != "" {
		plan, err = h.declarative.PlanForSchema(ctx, req.Schema)
	} else {
		plan, err = h.declarative.Plan(ctx)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"plan": fiber.Map{
			"changes":     plan.Changes,
			"ddl":         plan.DDL,
			"transaction": plan.Transaction,
			"duration":    plan.Duration.String(),
			"summary": fiber.Map{
				"total_changes":     len(plan.Changes),
				"create_count":      countByType(plan.Changes, migrations.ChangeCreate),
				"alter_count":       countByType(plan.Changes, migrations.ChangeAlter),
				"drop_count":        countByType(plan.Changes, migrations.ChangeDrop),
				"destructive_count": countDestructive(plan.Changes),
			},
		},
	})
}

// ApplySchema handles POST /api/v1/admin/internal-schema/apply
func (h *InternalSchemaHandler) ApplySchema(c fiber.Ctx) error {
	if h.declarative == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	var req struct {
		Schema           string `json:"schema"` // Optional: apply a specific schema
		AutoApprove      bool   `json:"auto_approve"`
		AllowDestructive bool   `json:"allow_destructive"`
	}
	if err := c.Bind().Body(&req); err != nil && err != fiber.ErrUnprocessableEntity {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Update allow destructive setting temporarily
	originalDestructive := h.config.AllowDestructive
	h.config.AllowDestructive = req.AllowDestructive
	defer func() { h.config.AllowDestructive = originalDestructive }()

	var result *migrations.ApplyResult
	var err error

	if req.Schema != "" {
		result, err = h.declarative.ApplyForSchema(ctx, req.Schema, req.AutoApprove)
	} else {
		result, err = h.declarative.Apply(ctx, req.AutoApprove)
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message":  "Schema applied successfully",
		"applied":  result.Applied,
		"duration": result.Duration.String(),
	})
}

// ValidateSchema handles GET /api/v1/admin/internal-schema/validate
func (h *InternalSchemaHandler) ValidateSchema(c fiber.Ctx) error {
	if h.declarative == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	result, err := h.declarative.Validate(ctx)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"valid":  result.Valid,
		"drifts": result.Drifts,
		"error":  errorToString(result.Error),
	})
}

// GetSchemaStatus handles GET /api/v1/admin/internal-schema/status
func (h *InternalSchemaHandler) GetSchemaStatus(c fiber.Ctx) error {
	if h.bootstrap == nil || h.validator == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	// Check bootstrap status
	bootstrapped, err := h.bootstrap.IsBootstrapped(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to check bootstrap status")
		bootstrapped = false
	}

	// Get migration state
	state, err := h.validator.DetectMigrationState(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to detect migration state")
		state = &migrations.MigrationState{}
	}

	// Get schema status
	status, err := h.declarative.GetSchemaStatus(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get schema status")
		status = &migrations.SchemaStatus{}
	}

	return c.JSON(fiber.Map{
		"bootstrapped":              bootstrapped,
		"has_imperative_migrations": state.HasImperativeMigrations,
		"last_migration_version":    state.LastAppliedVersion,
		"has_declarative_state":     state.HasDeclarativeState,
		"schema_fingerprint":        status.SchemaFingerprint,
		"pending_changes":           status.PendingChanges,
		"has_destructive_changes":   status.HasDestructiveChanges,
		"schema_dir":                h.config.SchemaDir,
	})
}

// MigrateSchema handles POST /api/v1/admin/internal-schema/migrate
func (h *InternalSchemaHandler) MigrateSchema(c fiber.Ctx) error {
	if h.transition == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Handler not initialized"})
	}

	ctx := c.Context()

	var req struct {
		Dir               string `json:"dir"`
		KeepOldMigrations bool   `json:"keep_old_migrations"`
	}
	if err := c.Bind().Body(&req); err != nil && err != fiber.ErrUnprocessableEntity {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	schemaDir := h.config.SchemaDir
	if req.Dir != "" {
		schemaDir = req.Dir
	}

	opts := migrations.TransitionOptions{
		SchemaDir:         schemaDir,
		KeepOldMigrations: true, // Always keep - internal migrations removed
		MigrationsDir:     "",   // No longer used - internal migrations removed
	}

	result, err := h.transition.Transition(ctx, opts)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success":                result.Success,
		"schema_dir":             result.SchemaFile,
		"fingerprint":            result.SchemaFingerprint,
		"last_migration_version": result.LastMigrationVersion,
		"transitioned_at":        result.TransitionedAt,
	})
}

// Helper functions

func countByType(changes []migrations.Change, changeType migrations.ChangeType) int {
	count := 0
	for _, c := range changes {
		if c.Type == changeType {
			count++
		}
	}
	return count
}

func countDestructive(changes []migrations.Change) int {
	count := 0
	for _, c := range changes {
		if c.Destructive {
			count++
		}
	}
	return count
}

func errorToString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Ensure Interface compliance
var _ fmt.Stringer = (*InternalSchemaHandler)(nil)

func (h *InternalSchemaHandler) String() string {
	return "InternalSchemaHandler"
}
