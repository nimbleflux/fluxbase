package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/email"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

type TenantHandler struct {
	DB                *database.Connection
	Manager           *tenantdb.Manager
	Storage           *tenantdb.Storage
	InvitationService *auth.InvitationService
	EmailService      email.Service
	Config            *config.Config
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
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	UserID     string    `json:"user_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

type CreateTenantRequest struct {
	// Basic info
	Slug     string                 `json:"slug" validate:"required,slug"`
	Name     string                 `json:"name" validate:"required,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Key generation
	AutoGenerateKeys bool `json:"auto_generate_keys"` // default: true

	// Admin assignment
	AdminEmail  *string `json:"admin_email,omitempty"`
	AdminUserID *string `json:"admin_user_id,omitempty"`

	// Key delivery
	SendKeysToEmail bool `json:"send_keys_to_email"`
}

// CreateTenantResponse represents the response for tenant creation
type CreateTenantResponse struct {
	Tenant          TenantResponse `json:"tenant"`
	AnonKey         *string        `json:"anon_key,omitempty"`
	ServiceKey      *string        `json:"service_key,omitempty"`
	InvitationSent  bool           `json:"invitation_sent"`
	InvitationEmail *string        `json:"invitation_email,omitempty"`
}

type UpdateTenantRequest struct {
	Name     *string                `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type AssignAdminRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

func NewTenantHandler(db *database.Connection, manager *tenantdb.Manager, storage *tenantdb.Storage, invitationService *auth.InvitationService, emailService email.Service, cfg *config.Config) *TenantHandler {
	return &TenantHandler{
		DB:                db,
		Manager:           manager,
		Storage:           storage,
		InvitationService: invitationService,
		EmailService:      emailService,
		Config:            cfg,
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

	// Create the tenant database
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
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Failed to create tenant: %v", err))
	}

	log.Info().Str("tenant_id", t.ID).Str("slug", t.Slug).Msg("Tenant created")

	// Get the user ID for audit trail
	userIDStr, _ := c.Locals("user_id").(string)
	var createdBy *uuid.UUID
	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err == nil {
			createdBy = &uid
		}
	}

	// Generate keys if requested (default: true)
	var anonKey, serviceKey *string
	if req.AutoGenerateKeys {
		anonKey, serviceKey, err = h.generateDefaultKeys(ctx, t.ID, createdBy)
		if err != nil {
			log.Warn().Err(err).Str("tenant_id", t.ID).Msg("Failed to generate default keys for tenant")
			// Don't fail the request - tenant was created successfully
		} else {
			log.Info().Str("tenant_id", t.ID).Msg("Auto-generated default keys for tenant")
		}
	}

	// Assign or invite admin if specified
	var invitationSent bool
	var invitationEmail *string
	if req.AdminUserID != nil || req.AdminEmail != nil {
		invitationSent, invitationEmail, err = h.assignOrInviteAdmin(ctx, t.ID, req, anonKey, serviceKey, createdBy)
		if err != nil {
			log.Warn().Err(err).Str("tenant_id", t.ID).Msg("Failed to assign/invite admin for tenant")
			// Don't fail the request - tenant was created successfully
		}
	}

	// Build response
	response := CreateTenantResponse{
		Tenant:          tenantToResponse(t),
		AnonKey:         anonKey,
		ServiceKey:      serviceKey,
		InvitationSent:  invitationSent,
		InvitationEmail: invitationEmail,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
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
		SELECT ta.id, ta.tenant_id, ta.user_id, ta.assigned_at,
		       du.email, du.role as dashboard_role
		FROM platform.tenant_admin_assignments ta
		INNER JOIN platform.users du ON du.id = ta.user_id
		WHERE ta.tenant_id = $1::uuid
		ORDER BY ta.assigned_at ASC
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
			&m.ID, &m.TenantID, &m.UserID, &m.AssignedAt,
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
		`SELECT EXISTS(SELECT 1 FROM platform.users WHERE id = $1::uuid AND deleted_at IS NULL)`,
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
		RETURNING id, tenant_id, user_id, assigned_at
	`, tenantID, req.UserID).Scan(
		&assignment.ID, &assignment.TenantID, &assignment.UserID, &assignment.AssignedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err := h.DB.Pool().QueryRow(ctx, `
				SELECT id, tenant_id, user_id, assigned_at
				FROM platform.tenant_admin_assignments
				WHERE tenant_id = $1::uuid AND user_id = $2::uuid
			`, tenantID, req.UserID).Scan(
				&assignment.ID, &assignment.TenantID, &assignment.UserID, &assignment.AssignedAt,
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

// generateDefaultKeys creates anon and service keys for a new tenant
func (h *TenantHandler) generateDefaultKeys(ctx context.Context, tenantID string, createdBy *uuid.UUID) (anonKey, serviceKey *string, err error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	// Generate anon key (publishable, limited scopes)
	anon, err := h.createServiceKey(ctx, CreateServiceKeyInternalRequest{
		TenantID:    &tenantUUID,
		KeyType:     "anon",
		Name:        "Default Anon Key",
		Description: "Auto-generated anonymous key for client-side access",
		Scopes:      []string{"read:*", "write:own"},
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create anon key: %w", err)
	}

	// Create tenant_service key (secret, full scopes)
	service, err := h.createServiceKey(ctx, CreateServiceKeyInternalRequest{
		TenantID:    &tenantUUID,
		KeyType:     "tenant_service",
		Name:        "Default Service Key",
		Description: "Auto-generated service key for server-side operations",
		Scopes:      []string{"*"},
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create service key: %w", err)
	}

	return &anon, &service, nil
}

// CreateServiceKeyInternalRequest represents an internal request to create a service key
type CreateServiceKeyInternalRequest struct {
	Name              string
	Description       string
	KeyType           string
	TenantID          *uuid.UUID
	Scopes            []string
	AllowedNamespaces []string
	RateLimitPerMin   *int
	CreatedBy         *uuid.UUID
}

// createServiceKey creates a service key programmatically (internal use)
func (h *TenantHandler) createServiceKey(ctx context.Context, req CreateServiceKeyInternalRequest) (string, error) {
	// Generate key bytes
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Determine prefix based on key type
	prefix := "sk_live_"
	switch req.KeyType {
	case "anon":
		prefix = "pk_anon_"
	case "publishable":
		prefix = "pk_live_"
	case "global_service":
		prefix = "sk_global_"
	case "tenant_service":
		prefix = "sk_tenant_"
	}

	fullKey := prefix + base64.URLEncoding.EncodeToString(keyBytes)
	keyPrefix := fullKey[:16]

	// Hash the key
	keyHash, err := bcrypt.GenerateFromPassword([]byte(fullKey), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash key: %w", err)
	}

	// Set default scopes if not provided
	scopes := req.Scopes
	if scopes == nil {
		switch req.KeyType {
		case "anon":
			scopes = []string{"read"}
		case "publishable":
			scopes = []string{"read", "write"}
		default:
			scopes = []string{"*"}
		}
	}

	// Insert into database
	var keyID uuid.UUID
	err = database.WrapWithServiceRole(ctx, h.DB, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `
			INSERT INTO platform.service_keys (name, description, key_hash, key_prefix, key_type, tenant_id, scopes, allowed_namespaces, is_active, rate_limit_per_minute, created_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
			RETURNING id
		`, req.Name, req.Description, string(keyHash), keyPrefix, req.KeyType, req.TenantID, scopes, req.AllowedNamespaces, req.RateLimitPerMin, req.CreatedBy).Scan(&keyID)
	})
	if err != nil {
		return "", fmt.Errorf("failed to insert key: %w", err)
	}

	log.Info().Str("key_id", keyID.String()).Str("key_type", req.KeyType).Str("tenant_id", fmt.Sprintf("%v", req.TenantID)).Msg("Auto-generated service key")

	return fullKey, nil
}

// assignOrInviteAdmin assigns an existing user or sends an invitation email
func (h *TenantHandler) assignOrInviteAdmin(
	ctx context.Context,
	tenantID string,
	req CreateTenantRequest,
	anonKey, serviceKey *string,
	invitedBy *uuid.UUID,
) (bool, *string, error) {
	// Option 1: Assign existing user directly
	if req.AdminUserID != nil {
		err := h.Storage.AssignUserToTenant(ctx, *req.AdminUserID, tenantID)
		if err != nil {
			return false, nil, fmt.Errorf("failed to assign admin: %w", err)
		}
		log.Info().Str("tenant_id", tenantID).Str("user_id", *req.AdminUserID).Msg("Admin assigned to tenant")
		return false, nil, nil
	}

	// Option 2: Invite by email
	if req.AdminEmail != nil && h.InvitationService != nil {
		// Parse tenant ID for invitation
		tenantUUID, err := uuid.Parse(tenantID)
		if err != nil {
			return false, nil, fmt.Errorf("invalid tenant ID: %w", err)
		}

		// Create invitation token with tenant context (role: tenant_admin)
		invitation, err := h.InvitationService.CreateInvitationWithTenant(ctx, *req.AdminEmail, "tenant_admin", &tenantUUID, invitedBy, 7*24*time.Hour)
		if err != nil {
			return false, nil, fmt.Errorf("failed to create invitation: %w", err)
		}

		// Build invitation link
		baseURL := h.Config.GetPublicBaseURL()
		if baseURL == "" {
			baseURL = "http://localhost:8080"
		}
		inviteLink := fmt.Sprintf("%s/admin/accept-invitation?token=%s&tenant=%s", baseURL, invitation.Token, tenantID)

		// Include keys in email if requested
		var keyInfo string
		if req.SendKeysToEmail && anonKey != nil && serviceKey != nil {
			keyInfo = fmt.Sprintf(`

Your API Keys:
- Anon Key: %s
- Service Key: %s

⚠️ Save these keys securely - they won't be shown again.`, *anonKey, *serviceKey)
		}

		// Get tenant name for email
		tenant, err := h.Storage.GetTenant(ctx, tenantID)
		tenantName := tenantID
		if err == nil && tenant != nil {
			tenantName = tenant.Name
		}

		// Send invitation email
		if h.EmailService != nil {
			err = h.EmailService.Send(ctx, *req.AdminEmail,
				fmt.Sprintf("You've been invited to manage %s", tenantName),
				fmt.Sprintf(`You have been invited as an administrator for %s on Fluxbase.

Click here to accept: %s
%s`, tenantName, inviteLink, keyInfo))
			if err != nil {
				log.Warn().Err(err).Msg("Failed to send invitation email")
				// Still return success since the invitation was created
			} else {
				log.Info().Str("email", *req.AdminEmail).Str("tenant_id", tenantID).Msg("Invitation email sent")
			}
		}

		return true, req.AdminEmail, nil
	}

	return false, nil, nil
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
