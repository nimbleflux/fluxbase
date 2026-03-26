package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

type TenantHandler struct {
	DB      *database.Connection
	Manager *tenantdb.Manager
	Storage *tenantdb.Storage
}

type TenantResponse struct {
	ID        string                 `json:"id"`
	Slug      string                 `json:"slug"`
	Name      string                 `json:"name"`
	DbName    *string                `json:"db_name,omitempty"`
	Status    string                 `json:"status"`
	IsDefault bool                   `json:"is_default"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

type TenantAdminAssignment struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTenantRequest struct {
	Slug     string                 `json:"slug" validate:"required,slug"`
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateTenantRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AssignAdminRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

func NewTenantHandler(db *database.Connection, manager *tenantdb.Manager, storage *tenantdb.Storage) *TenantHandler {
	return &TenantHandler{
		DB:      db,
		Manager: manager,
		Storage: storage,
	}
}

func tenantToResponse(t *tenantdb.Tenant) TenantResponse {
	return TenantResponse{
		ID:        t.ID,
		Slug:      t.Slug,
		Name:      t.Name,
		DbName:    t.DBName,
		Status:    string(t.Status),
		IsDefault: t.IsDefault,
		Metadata:  t.Metadata,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		DeletedAt: t.DeletedAt,
	}
}

func (h *TenantHandler) ListTenants(c fiber.Ctx) error {
	ctx := c.Context()

	tenants, err := h.Storage.GetAllActiveTenants(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tenants")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list tenants")
	}

	result := make([]TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = tenantToResponse(&t)
	}

	return c.JSON(result)
}

func (h *TenantHandler) ListMyTenants(c fiber.Ctx) error {
	ctx := c.Context()
	userID, _ := c.Locals("user_id").(string)

	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authentication required")
	}

	tenants, err := h.Storage.GetTenantsForUser(ctx, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list user tenants")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list tenants")
	}

	type TenantWithRole struct {
		TenantResponse
		MyRole string `json:"my_role"`
	}

	result := make([]TenantWithRole, len(tenants))
	for i, t := range tenants {
		result[i] = TenantWithRole{
			TenantResponse: tenantToResponse(&t),
			MyRole:         "tenant_admin",
		}
	}

	return c.JSON(result)
}

func (h *TenantHandler) GetTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)

	if !isInstanceAdmin {
		hasAccess, err := h.Storage.IsUserAssignedToTenant(ctx, userID, tenantID)
		if err != nil || !hasAccess {
			return fiber.NewError(fiber.StatusForbidden, "Access denied to this tenant")
		}
	}

	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	return c.JSON(tenantToResponse(t))
}

func (h *TenantHandler) CreateTenant(c fiber.Ctx) error {
	ctx := c.Context()

	var req CreateTenantRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if !isValidSlug(req.Slug) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid slug format (use lowercase letters, numbers, and hyphens)")
	}

	existing, _ := h.Storage.GetTenantBySlug(ctx, req.Slug)
	if existing != nil {
		return fiber.NewError(fiber.StatusConflict, "Tenant with this slug already exists")
	}

	metadata := make(map[string]any)
	if req.Metadata != nil {
		metadata = req.Metadata
	}

	t, err := h.Manager.CreateTenantDatabase(ctx, tenantdb.CreateTenantRequest{
		Slug:     req.Slug,
		Name:     req.Name,
		Metadata: metadata,
	})
	if err != nil {
		if errors.Is(err, tenantdb.ErrMaxTenantsReached) {
			return fiber.NewError(fiber.StatusConflict, "Maximum number of tenants reached")
		}
		log.Error().Err(err).Msg("Failed to create tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create tenant")
	}

	log.Info().Str("tenant_id", t.ID).Str("slug", t.Slug).Msg("Tenant created")

	return c.Status(fiber.StatusCreated).JSON(tenantToResponse(t))
}

func (h *TenantHandler) UpdateTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	var req UpdateTenantRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update tenant")
	}

	updateReq := tenantdb.UpdateTenantRequest{
		Name:     req.Name,
		Metadata: req.Metadata,
	}

	if err := h.Storage.UpdateTenant(ctx, t.ID, updateReq); err != nil {
		log.Error().Err(err).Msg("Failed to update tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update tenant")
	}

	t, err = h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get updated tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get updated tenant")
	}

	return c.JSON(tenantToResponse(t))
}

func (h *TenantHandler) DeleteTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete tenant")
	}

	if t.IsDefault {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot delete the default tenant")
	}

	if err := h.Manager.DeleteTenantDatabase(ctx, tenantID); err != nil {
		log.Error().Err(err).Msg("Failed to delete tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete tenant")
	}

	log.Info().Str("tenant_id", tenantID).Msg("Tenant deleted")

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *TenantHandler) MigrateTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to migrate tenant")
	}

	if t.UsesMainDatabase() {
		return c.JSON(fiber.Map{"status": "skipped", "reason": "uses main database"})
	}

	if err := h.Manager.MigrateTenant(ctx, tenantID); err != nil {
		log.Error().Err(err).Msg("Failed to migrate tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to migrate tenant")
	}

	log.Info().Str("tenant_id", tenantID).Msg("Tenant migrated")

	return c.JSON(fiber.Map{"status": "migrated"})
}

func (h *TenantHandler) ListAdmins(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)

	if !isInstanceAdmin {
		hasAccess, err := h.Storage.IsUserAssignedToTenant(ctx, userID, tenantID)
		if err != nil || !hasAccess {
			return fiber.NewError(fiber.StatusForbidden, "Access denied to this tenant")
		}
	}

	rows, err := h.DB.Pool().Query(ctx, `
		SELECT ta.id, ta.tenant_id, ta.user_id, ta.created_at,
		       u.email, du.role as dashboard_role
		FROM platform.tenant_admin_assignments ta
		INNER JOIN auth.users u ON u.id = ta.user_id
		INNER JOIN dashboard.users du ON du.id = ta.user_id
		WHERE ta.tenant_id = $1::uuid
		ORDER BY ta.created_at ASC
	`, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list admins")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list admins")
	}
	defer rows.Close()

	type AdminWithUser struct {
		TenantAdminAssignment
		Email         string `json:"email"`
		DashboardRole string `json:"dashboard_role"`
	}

	var admins []AdminWithUser
	for rows.Next() {
		var m AdminWithUser
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.UserID, &m.CreatedAt,
			&m.Email, &m.DashboardRole,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan admin")
			continue
		}
		admins = append(admins, m)
	}

	if admins == nil {
		admins = []AdminWithUser{}
	}

	return c.JSON(admins)
}

func (h *TenantHandler) AssignAdmin(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	var req AssignAdminRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	var userExists bool
	err := h.DB.Pool().QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM auth.users WHERE id = $1::uuid AND deleted_at IS NULL)`,
		req.UserID,
	).Scan(&userExists)
	if err != nil || !userExists {
		return fiber.NewError(fiber.StatusBadRequest, "User not found")
	}

	var assignment TenantAdminAssignment
	err = h.DB.Pool().QueryRow(ctx, `
		INSERT INTO platform.tenant_admin_assignments (tenant_id, user_id)
		VALUES ($1::uuid, $2::uuid)
		ON CONFLICT (tenant_id, user_id) DO NOTHING
		RETURNING id, tenant_id, user_id, created_at
	`, tenantID, req.UserID).Scan(
		&assignment.ID, &assignment.TenantID, &assignment.UserID, &assignment.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err := h.DB.Pool().QueryRow(ctx, `
				SELECT id, tenant_id, user_id, created_at
				FROM platform.tenant_admin_assignments
				WHERE tenant_id = $1::uuid AND user_id = $2::uuid
			`, tenantID, req.UserID).Scan(
				&assignment.ID, &assignment.TenantID, &assignment.UserID, &assignment.CreatedAt,
			)
			if err != nil {
				log.Error().Err(err).Msg("Failed to get existing assignment")
				return fiber.NewError(fiber.StatusInternalServerError, "Failed to assign admin")
			}
		} else {
			log.Error().Err(err).Msg("Failed to assign admin")
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to assign admin")
		}
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("user_id", req.UserID).
		Msg("Admin assigned to tenant")

	return c.Status(fiber.StatusCreated).JSON(assignment)
}

func (h *TenantHandler) RemoveAdmin(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID := c.Params("userId")

	result, err := h.DB.Pool().Exec(ctx, `
		DELETE FROM platform.tenant_admin_assignments
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
	`, tenantID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove admin")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove admin")
	}

	if result.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Admin assignment not found")
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("user_id", userID).
		Msg("Admin removed from tenant")

	return c.SendStatus(fiber.StatusNoContent)
}

func isValidSlug(s string) bool {
	if len(s) == 0 || len(s) > 63 {
		return false
	}
	for i, r := range s {
		if i == 0 && (r < 'a' || r > 'z') {
			return false
		}
		if i == len(s)-1 && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return false
		}
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return false
		}
	}
	return true
}

// GetTenantSchemaStatus returns the status of a tenant's declarative schema
func (h *TenantHandler) GetTenantSchemaStatus(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return c.JSON(fiber.Map{
			"enabled":             false,
			"message":             "Tenant declarative schemas are not enabled",
			"has_schema_file":     false,
			"has_pending_changes": false,
		})
	}

	// Get schema status
	status, err := h.Manager.GetTenantSchemaStatus(ctx, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tenant schema status")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant schema status")
	}

	return c.JSON(fiber.Map{
		"enabled":                  true,
		"tenant_id":                tenantID,
		"tenant_slug":              t.Slug,
		"schema_file":              status.SchemaFile,
		"has_schema_file":          status.SchemaFingerprint != "",
		"schema_fingerprint":       status.SchemaFingerprint,
		"last_applied_fingerprint": status.LastAppliedFingerprint,
		"last_applied_at":          status.LastAppliedAt,
		"has_pending_changes":      status.HasPendingChanges,
		"uses_main_database":       t.UsesMainDatabase(),
	})
}

// ApplyTenantSchema applies the declarative schema for a tenant
func (h *TenantHandler) ApplyTenantSchema(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if tenant uses main database
	if t.UsesMainDatabase() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot apply declarative schema to tenant using main database")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Tenant declarative schemas are not enabled")
	}

	// Apply the schema
	if err := h.Manager.ApplyTenantDeclarativeSchema(ctx, tenantID); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to apply tenant schema")
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to apply schema: %v", err))
	}

	log.Info().Str("tenant_id", tenantID).Str("tenant_slug", t.Slug).Msg("Tenant schema applied")

	return c.JSON(fiber.Map{
		"status":      "applied",
		"tenant_id":   tenantID,
		"tenant_slug": t.Slug,
	})
}

// UploadTenantSchemaRequest represents the request body for uploading a tenant schema
type UploadTenantSchemaRequest struct {
	Schema string `json:"schema" validate:"required"`
}

// GetStoredSchema retrieves the stored schema content for a tenant
func (h *TenantHandler) GetStoredSchema(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Tenant declarative schemas are not enabled")
	}

	// Get stored schema content
	content, fingerprint, updatedAt, err := declarativeSvc.GetStoredSchemaContent(ctx, t.Slug)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stored schema")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stored schema")
	}

	if content == "" {
		return c.JSON(fiber.Map{
			"has_schema":  false,
			"tenant_id":   tenantID,
			"tenant_slug": t.Slug,
		})
	}

	return c.JSON(fiber.Map{
		"has_schema":  true,
		"tenant_id":   tenantID,
		"tenant_slug": t.Slug,
		"schema":      content,
		"fingerprint": fingerprint,
		"updated_at":  updatedAt,
	})
}

// UploadTenantSchema uploads and stores schema content for a tenant
func (h *TenantHandler) UploadTenantSchema(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Tenant declarative schemas are not enabled")
	}

	// Parse request body
	var req UploadTenantSchemaRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Schema == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Schema content cannot be empty")
	}

	// Store the schema content
	if err := declarativeSvc.StoreSchemaContent(ctx, t.Slug, req.Schema); err != nil {
		log.Error().Err(err).Msg("Failed to store schema")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to store schema")
	}

	// Calculate fingerprint for response
	_, fingerprint, _, _ := declarativeSvc.GetStoredSchemaContent(ctx, t.Slug)

	log.Info().Str("tenant_id", tenantID).Str("tenant_slug", t.Slug).Msg("Tenant schema uploaded")

	return c.JSON(fiber.Map{
		"status":      "uploaded",
		"tenant_id":   tenantID,
		"tenant_slug": t.Slug,
		"fingerprint": fingerprint,
	})
}

// ApplyUploadedTenantSchema applies the previously uploaded schema for a tenant
func (h *TenantHandler) ApplyUploadedTenantSchema(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if tenant uses main database
	if t.UsesMainDatabase() {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot apply declarative schema to tenant using main database")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Tenant declarative schemas are not enabled")
	}

	// Get stored schema content
	content, fingerprint, _, err := declarativeSvc.GetStoredSchemaContent(ctx, t.Slug)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get stored schema")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get stored schema")
	}

	if content == "" {
		return fiber.NewError(fiber.StatusNotFound, "No stored schema found for this tenant. Upload a schema first.")
	}

	// Apply the schema from stored content
	if err := declarativeSvc.ApplyTenantSchemaFromContent(ctx, t, content); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to apply tenant schema")
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to apply schema: %v", err))
	}

	log.Info().Str("tenant_id", tenantID).Str("tenant_slug", t.Slug).Msg("Tenant stored schema applied")

	return c.JSON(fiber.Map{
		"status":      "applied",
		"tenant_id":   tenantID,
		"tenant_slug": t.Slug,
		"fingerprint": fingerprint,
	})
}

// DeleteStoredSchema deletes the stored schema content for a tenant
func (h *TenantHandler) DeleteStoredSchema(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if tenant exists
	t, err := h.Storage.GetTenant(ctx, tenantID)
	if err != nil {
		if errors.Is(err, tenantdb.ErrTenantNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	// Check if declarative service is configured
	declarativeSvc := h.Manager.GetDeclarativeService()
	if declarativeSvc == nil {
		return fiber.NewError(fiber.StatusBadRequest, "Tenant declarative schemas are not enabled")
	}

	// Delete the stored schema
	if err := declarativeSvc.DeleteStoredSchema(ctx, t.Slug); err != nil {
		log.Error().Err(err).Msg("Failed to delete stored schema")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete stored schema")
	}

	log.Info().Str("tenant_id", tenantID).Str("tenant_slug", t.Slug).Msg("Tenant stored schema deleted")

	return c.SendStatus(fiber.StatusNoContent)
}
