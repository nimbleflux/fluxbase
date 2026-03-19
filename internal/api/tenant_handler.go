package api

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/rs/zerolog/log"
)

// TenantHandler handles tenant management endpoints
type TenantHandler struct {
	DB *database.Connection
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID        string                 `json:"id"`
	Slug      string                 `json:"slug"`
	Name      string                 `json:"name"`
	IsDefault bool                   `json:"is_default"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
	DeletedAt *time.Time             `json:"deleted_at,omitempty"`
}

// TenantMembership represents a user's membership in a tenant
type TenantMembership struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// CreateTenantRequest is the request body for creating a tenant
type CreateTenantRequest struct {
	Slug     string                 `json:"slug" validate:"required,slug"`
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UpdateTenantRequest is the request body for updating a tenant
type UpdateTenantRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AddMemberRequest is the request body for adding a member to a tenant
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   string `json:"role" validate:"required,oneof=tenant_admin tenant_member"`
}

// UpdateMemberRequest is the request body for updating a member's role
type UpdateMemberRequest struct {
	Role string `json:"role" validate:"required,oneof=tenant_admin tenant_member"`
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(db *database.Connection) *TenantHandler {
	return &TenantHandler{DB: db}
}

// RegisterTenantRoutes registers tenant management routes
func RegisterTenantRoutes(router fiber.Router, db *database.Connection) {
	handler := NewTenantHandler(db)
	tenants := router.Group("/tenants")

	// Public routes (require authentication)
	tenants.Get("/mine", handler.ListMyTenants)

	// Instance admin only routes
	tenants.Get("/", middleware.RequireInstanceAdmin(), handler.ListTenants)
	tenants.Post("/", middleware.RequireInstanceAdmin(), handler.CreateTenant)

	// Tenant-specific routes
	tenants.Get("/:id", handler.GetTenant)
	tenants.Patch("/:id", middleware.RequireTenantRole("tenant_admin"), handler.UpdateTenant)
	tenants.Delete("/:id", middleware.RequireInstanceAdmin(), handler.DeleteTenant)

	// Member management
	tenants.Get("/:id/members", middleware.RequireTenantRole("tenant_member"), handler.ListMembers)
	tenants.Post("/:id/members", middleware.RequireTenantRole("tenant_admin"), handler.AddMember)
	tenants.Patch("/:id/members/:userId", middleware.RequireTenantRole("tenant_admin"), handler.UpdateMemberRole)
	tenants.Delete("/:id/members/:userId", middleware.RequireTenantRole("tenant_admin"), handler.RemoveMember)
}

// ListTenants lists all tenants (instance admin only)
func (h *TenantHandler) ListTenants(c fiber.Ctx) error {
	ctx := c.Context()

	rows, err := h.DB.Pool().Query(ctx, `
		SELECT id, slug, name, is_default, metadata, created_at, updated_at, deleted_at
		FROM platform.tenants
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list tenants")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list tenants")
	}
	defer rows.Close()

	var tenants []Tenant
	for rows.Next() {
		var t Tenant
		var metadata []byte
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.IsDefault, &metadata,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan tenant")
			continue
		}
		if len(metadata) > 0 {
			t.Metadata = make(map[string]interface{})
			_ = json.Unmarshal(metadata, &t.Metadata)
		}
		tenants = append(tenants, t)
	}

	return c.JSON(tenants)
}

// ListMyTenants lists the current user's tenants
func (h *TenantHandler) ListMyTenants(c fiber.Ctx) error {
	ctx := c.Context()
	userID, _ := c.Locals("user_id").(string)

	if userID == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "Authentication required")
	}

	rows, err := h.DB.Pool().Query(ctx, `
		SELECT t.id, t.slug, t.name, t.is_default, t.metadata, t.created_at, tm.role
		FROM platform.tenants t
		INNER JOIN platform.tenant_memberships tm ON tm.tenant_id = t.id
		WHERE tm.user_id = $1::uuid
		AND t.deleted_at IS NULL
		ORDER BY t.name
	`, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list user tenants")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list tenants")
	}
	defer rows.Close()

	type TenantWithRole struct {
		Tenant
		MyRole string `json:"my_role"`
	}

	var tenants []TenantWithRole
	for rows.Next() {
		var t TenantWithRole
		var metadata []byte
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.IsDefault, &metadata, &t.CreatedAt, &t.MyRole,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan tenant")
			continue
		}
		if len(metadata) > 0 {
			t.Metadata = make(map[string]interface{})
			_ = json.Unmarshal(metadata, &t.Metadata)
		}
		tenants = append(tenants, t)
	}

	return c.JSON(tenants)
}

// GetTenant gets a single tenant by ID
func (h *TenantHandler) GetTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID, _ := c.Locals("user_id").(string)
	isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)

	// Check if user has access to this tenant
	if !isInstanceAdmin {
		isMember, err := middleware.ValidateTenantMembership(ctx, h.DB, userID, tenantID)
		if err != nil || !isMember {
			return fiber.NewError(fiber.StatusForbidden, "Access denied to this tenant")
		}
	}

	var t Tenant
	var metadata []byte
	err := h.DB.Pool().QueryRow(ctx, `
		SELECT id, slug, name, is_default, metadata, created_at, updated_at, deleted_at
		FROM platform.tenants
		WHERE id = $1::uuid AND deleted_at IS NULL
	`, tenantID).Scan(
		&t.ID, &t.Slug, &t.Name, &t.IsDefault, &metadata,
		&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to get tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get tenant")
	}

	if len(metadata) > 0 {
		t.Metadata = make(map[string]interface{})
		_ = json.Unmarshal(metadata, &t.Metadata)
	}

	return c.JSON(t)
}

// CreateTenant creates a new tenant (instance admin only)
func (h *TenantHandler) CreateTenant(c fiber.Ctx) error {
	ctx := c.Context()

	var req CreateTenantRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Validate slug format
	if !isValidSlug(req.Slug) {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid slug format (use lowercase letters, numbers, and hyphens)")
	}

	// Check if slug already exists
	var exists bool
	err := h.DB.Pool().QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM platform.tenants WHERE slug = $1)`,
		req.Slug,
	).Scan(&exists)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check slug existence")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create tenant")
	}
	if exists {
		return fiber.NewError(fiber.StatusConflict, "Tenant with this slug already exists")
	}

	// Create tenant
	var t Tenant
	var metadata []byte
	err = h.DB.Pool().QueryRow(ctx, `
		INSERT INTO platform.tenants (slug, name, metadata)
		VALUES ($1, $2, $3)
		RETURNING id, slug, name, is_default, metadata, created_at
	`, req.Slug, req.Name, req.Metadata).Scan(
		&t.ID, &t.Slug, &t.Name, &t.IsDefault, &metadata, &t.CreatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create tenant")
	}

	if len(metadata) > 0 {
		t.Metadata = make(map[string]interface{})
		_ = json.Unmarshal(metadata, &t.Metadata)
	}

	log.Info().Str("tenant_id", t.ID).Str("slug", t.Slug).Msg("Tenant created")

	return c.Status(fiber.StatusCreated).JSON(t)
}

// UpdateTenant updates a tenant
func (h *TenantHandler) UpdateTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	var req UpdateTenantRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Build update query dynamically
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}

	if len(updates) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "No fields to update")
	}

	updates["updated_at"] = time.Now()

	// Build and execute update
	query := `UPDATE platform.tenants SET `
	args := make([]interface{}, 0, len(updates)+1)
	i := 1
	for k, v := range updates {
		if i > 1 {
			query += `, `
		}
		query += k + ` = $` + itoa(i)
		args = append(args, v)
		i++
	}
	query += ` WHERE id = $` + itoa(i) + ` AND deleted_at IS NULL`
	args = append(args, tenantID)

	result, err := h.DB.Pool().Exec(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update tenant")
	}

	if result.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
	}

	// Return updated tenant
	return h.GetTenant(c)
}

// DeleteTenant soft-deletes a tenant (instance admin only)
func (h *TenantHandler) DeleteTenant(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	// Check if this is the default tenant
	var isDefault bool
	err := h.DB.Pool().QueryRow(ctx,
		`SELECT is_default FROM platform.tenants WHERE id = $1::uuid AND deleted_at IS NULL`,
		tenantID,
	).Scan(&isDefault)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
		}
		log.Error().Err(err).Msg("Failed to check tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete tenant")
	}

	if isDefault {
		return fiber.NewError(fiber.StatusBadRequest, "Cannot delete the default tenant")
	}

	// Soft delete
	result, err := h.DB.Pool().Exec(ctx,
		`UPDATE platform.tenants SET deleted_at = NOW() WHERE id = $1::uuid AND deleted_at IS NULL`,
		tenantID,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete tenant")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete tenant")
	}

	if result.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Tenant not found")
	}

	log.Info().Str("tenant_id", tenantID).Msg("Tenant deleted")

	return c.SendStatus(fiber.StatusNoContent)
}

// ListMembers lists members of a tenant
func (h *TenantHandler) ListMembers(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	rows, err := h.DB.Pool().Query(ctx, `
		SELECT tm.id, tm.tenant_id, tm.user_id, tm.role, tm.created_at, tm.updated_at,
		       u.email, u.role as user_role
		FROM platform.tenant_memberships tm
		INNER JOIN auth.users u ON u.id = tm.user_id
		WHERE tm.tenant_id = $1::uuid
		ORDER BY tm.created_at ASC
	`, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list members")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list members")
	}
	defer rows.Close()

	type MemberWithUser struct {
		TenantMembership
		Email    string `json:"email"`
		UserRole string `json:"user_role"`
	}

	var members []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.CreatedAt, &m.UpdatedAt,
			&m.Email, &m.UserRole,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to scan member")
			continue
		}
		members = append(members, m)
	}

	return c.JSON(members)
}

// AddMember adds a member to a tenant
func (h *TenantHandler) AddMember(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")

	var req AddMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Verify user exists
	var userExists bool
	err := h.DB.Pool().QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM auth.users WHERE id = $1::uuid AND deleted_at IS NULL)`,
		req.UserID,
	).Scan(&userExists)
	if err != nil || !userExists {
		return fiber.NewError(fiber.StatusBadRequest, "User not found")
	}

	// Add membership
	var m TenantMembership
	err = h.DB.Pool().QueryRow(ctx, `
		INSERT INTO platform.tenant_memberships (tenant_id, user_id, role)
		VALUES ($1::uuid, $2::uuid, $3)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET role = $3, updated_at = NOW()
		RETURNING id, tenant_id, user_id, role, created_at
	`, tenantID, req.UserID, req.Role).Scan(
		&m.ID, &m.TenantID, &m.UserID, &m.Role, &m.CreatedAt,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add member")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to add member")
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("user_id", req.UserID).
		Str("role", req.Role).
		Msg("Member added to tenant")

	return c.Status(fiber.StatusCreated).JSON(m)
}

// UpdateMemberRole updates a member's role
func (h *TenantHandler) UpdateMemberRole(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID := c.Params("userId")

	var req UpdateMemberRequest
	if err := c.Bind().Body(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	result, err := h.DB.Pool().Exec(ctx, `
		UPDATE platform.tenant_memberships
		SET role = $1, updated_at = NOW()
		WHERE tenant_id = $2::uuid AND user_id = $3::uuid
	`, req.Role, tenantID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to update member role")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update member role")
	}

	if result.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Membership not found")
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("user_id", userID).
		Str("role", req.Role).
		Msg("Member role updated")

	return c.SendStatus(fiber.StatusNoContent)
}

// RemoveMember removes a member from a tenant
func (h *TenantHandler) RemoveMember(c fiber.Ctx) error {
	ctx := c.Context()
	tenantID := c.Params("id")
	userID := c.Params("userId")

	result, err := h.DB.Pool().Exec(ctx, `
		DELETE FROM platform.tenant_memberships
		WHERE tenant_id = $1::uuid AND user_id = $2::uuid
	`, tenantID, userID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to remove member")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove member")
	}

	if result.RowsAffected() == 0 {
		return fiber.NewError(fiber.StatusNotFound, "Membership not found")
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("user_id", userID).
		Msg("Member removed from tenant")

	return c.SendStatus(fiber.StatusNoContent)
}

// Helper functions

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

func itoa(i int) string {
	return string(rune('0' + i))
}
