package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/rs/zerolog/log"
)

var (
	// ErrTenantNotFound is returned when tenant cannot be found
	ErrTenantNotFound = errors.New("tenant not found")
	// ErrNotTenantMember is returned when user is not a member of the tenant
	ErrNotTenantMember = errors.New("user is not a member of this tenant")
)

// TenantConfig holds configuration for tenant middleware
type TenantConfig struct {
	// DB is the database connection pool
	DB *database.Connection
}

// TenantMiddleware extracts tenant context from request
// Precedence: X-FB-Tenant header > JWT claim > default tenant
func TenantMiddleware(config TenantConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Get user ID from context (set by auth middleware)
		userID, _ := c.Locals("user_id").(string)

		// Get claims if available
		var claims *auth.TokenClaims
		if c, ok := c.Locals("claims").(*auth.TokenClaims); ok {
			claims = c
		}

		var tenantID string
		var tenantSource string

		// 1. Check X-FB-Tenant header (explicit override)
		if headerTenant := c.Get("X-FB-Tenant"); headerTenant != "" {
			// Validate user is member of this tenant (if authenticated)
			if userID != "" {
				isMember, err := ValidateTenantMembership(c.Context(), config.DB, userID, headerTenant)
				if err != nil {
					log.Debug().Err(err).Str("tenant_id", headerTenant).Msg("Failed to validate tenant membership")
				} else if isMember {
					tenantID = headerTenant
					tenantSource = "header"
				}
			} else {
				// For anonymous/unauthenticated requests, accept header
				tenantID = headerTenant
				tenantSource = "header"
			}
		}

		// 2. Check JWT claim
		if tenantID == "" && claims != nil && claims.TenantID != nil {
			tenantID = *claims.TenantID
			tenantSource = "jwt"
		}

		// 3. Fall back to default tenant
		if tenantID == "" {
			defaultID, err := GetDefaultTenantID(c.Context(), config.DB)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to get default tenant")
			} else {
				tenantID = defaultID
				tenantSource = "default"
			}
		}

		// Store tenant context
		c.Locals("tenant_id", tenantID)
		c.Locals("tenant_source", tenantSource)

		// Get tenant role if we have a user and tenant
		if userID != "" && tenantID != "" && claims != nil {
			// Use tenant role from claims if available and tenant matches
			if claims.TenantID != nil && *claims.TenantID == tenantID && claims.TenantRole != "" {
				c.Locals("tenant_role", claims.TenantRole)
			} else {
				// Fetch tenant role from database
				role, err := GetUserTenantRole(c.Context(), config.DB, userID, tenantID)
				if err != nil {
					log.Debug().Err(err).Msg("Failed to get tenant role")
				} else if role != "" {
					c.Locals("tenant_role", role)
				}
			}
		}

		// Check if user is instance admin
		if claims != nil && claims.IsInstanceAdmin {
			c.Locals("is_instance_admin", true)
		} else if userID != "" {
			isAdmin, err := IsInstanceAdmin(c.Context(), config.DB, userID)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to check instance admin status")
			} else if isAdmin {
				c.Locals("is_instance_admin", true)
			}
		}

		log.Debug().
			Str("tenant_id", tenantID).
			Str("tenant_source", tenantSource).
			Str("user_id", userID).
			Str("path", c.Path()).
			Msg("TenantMiddleware: Set tenant context")

		return c.Next()
	}
}

// ValidateTenantMembership checks if user is a member of the specified tenant
func ValidateTenantMembership(ctx context.Context, db *database.Connection, userID, tenantID string) (bool, error) {
	var isMember bool
	err := db.Pool().QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM tenant_memberships tm
			INNER JOIN tenants t ON t.id = tm.tenant_id
			WHERE tm.user_id = $1::uuid
			AND tm.tenant_id = $2::uuid
			AND t.deleted_at IS NULL
		)`,
		userID, tenantID,
	).Scan(&isMember)

	if err != nil {
		return false, fmt.Errorf("failed to check tenant membership: %w", err)
	}

	return isMember, nil
}

// GetUserTenantRole gets the user's role in the specified tenant
func GetUserTenantRole(ctx context.Context, db *database.Connection, userID, tenantID string) (string, error) {
	var role string
	err := db.Pool().QueryRow(ctx,
		`SELECT tm.role FROM tenant_memberships tm
		INNER JOIN tenants t ON t.id = tm.tenant_id
		WHERE tm.user_id = $1::uuid
		AND tm.tenant_id = $2::uuid
		AND t.deleted_at IS NULL`,
		userID, tenantID,
	).Scan(&role)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get tenant role: %w", err)
	}

	return role, nil
}

// GetDefaultTenantID gets the default tenant ID
func GetDefaultTenantID(ctx context.Context, db *database.Connection) (string, error) {
	var id string
	err := db.Pool().QueryRow(ctx,
		`SELECT id::text FROM tenants WHERE is_default = true AND deleted_at IS NULL LIMIT 1`,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrTenantNotFound
		}
		return "", fmt.Errorf("failed to get default tenant: %w", err)
	}

	return id, nil
}

// IsInstanceAdmin checks if the user is an instance-level admin
func IsInstanceAdmin(ctx context.Context, db *database.Connection, userID string) (bool, error) {
	var isAdmin bool
	err := db.Pool().QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM dashboard.users
			WHERE id = $1::uuid
			AND role = 'instance_admin'
			AND deleted_at IS NULL
			AND is_active = true
		)`,
		userID,
	).Scan(&isAdmin)

	if err != nil {
		return false, fmt.Errorf("failed to check instance admin: %w", err)
	}

	return isAdmin, nil
}

// SetTenantSessionContext sets the PostgreSQL session variable for tenant context
// This should be called at the beginning of each database transaction
func SetTenantSessionContext(ctx context.Context, tx pgx.Tx, tenantID string) error {
	if tenantID == "" {
		return nil
	}

	_, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant session context: %w", err)
	}

	return nil
}

// RequireTenantRole creates a middleware that requires a specific tenant role
func RequireTenantRole(requiredRole string) fiber.Handler {
	return func(c fiber.Ctx) error {
		isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)
		if isInstanceAdmin {
			return c.Next()
		}

		tenantRole, _ := c.Locals("tenant_role").(string)
		if tenantRole == "" {
			return fiber.NewError(fiber.StatusForbidden, "tenant membership required")
		}

		if tenantRole != requiredRole && tenantRole != "tenant_admin" {
			return fiber.NewError(fiber.StatusForbidden, fmt.Sprintf("tenant %s role required", requiredRole))
		}

		return c.Next()
	}
}

// RequireInstanceAdmin creates a middleware that requires instance admin role
func RequireInstanceAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)
		if !isInstanceAdmin {
			return fiber.NewError(fiber.StatusForbidden, "instance admin role required")
		}
		return c.Next()
	}
}

// GetUserTenantIDs gets all tenant IDs that a user is a member of
func GetUserTenantIDs(ctx context.Context, db *database.Connection, userID string) ([]string, error) {
	rows, err := db.Pool().Query(ctx,
		`SELECT tm.tenant_id::text FROM tenant_memberships tm
		INNER JOIN tenants t ON t.id = tm.tenant_id
		WHERE tm.user_id = $1::uuid
		AND t.deleted_at IS NULL
		ORDER BY t.name`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}
	defer rows.Close()

	var tenantIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan tenant ID: %w", err)
		}
		tenantIDs = append(tenantIDs, id)
	}

	return tenantIDs, nil
}
