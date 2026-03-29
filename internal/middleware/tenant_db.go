package middleware

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
)

var ErrTenantDBUnavailable = errors.New("tenant database unavailable")

type TenantDBConfig struct {
	Router  *tenantdb.Router
	Storage *tenantdb.Storage
}

func TenantDBMiddleware(cfg TenantDBConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, _ := c.Locals("user_id").(string)
		isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)

		var claims *auth.TokenClaims
		if c, ok := c.Locals("claims").(*auth.TokenClaims); ok {
			claims = c
		}

		tenantID, tenantSource := resolveTenantID(c, userID, isInstanceAdmin, claims, cfg.Storage)

		if tenantID != "" && !isInstanceAdmin && userID != "" {
			hasAccess, err := cfg.Storage.IsUserAssignedToTenant(c.Context(), userID, tenantID)
			if err != nil {
				log.Debug().Err(err).Msg("Failed to check tenant access")
				return fiber.NewError(fiber.StatusInternalServerError, "failed to verify tenant access")
			}
			if !hasAccess {
				return fiber.NewError(fiber.StatusForbidden, "access denied to this tenant")
			}
		}

		var pool *pgxpool.Pool
		if tenantID != "" {
			var err error
			pool, err = cfg.Router.GetPool(tenantID)
			if err != nil {
				log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to get tenant pool")
				return fiber.NewError(fiber.StatusInternalServerError, "tenant database unavailable")
			}
		}

		c.Locals("tenant_id", tenantID)
		c.Locals("tenant_source", tenantSource)
		c.Locals("tenant_db", pool)

		if tenantID != "" && cfg.Storage != nil {
			if tenant, err := cfg.Storage.GetTenant(c.Context(), tenantID); err == nil {
				c.Locals("tenant_slug", tenant.Slug)
				if !tenant.UsesMainDatabase() {
					c.Locals("tenant_db_name", *tenant.DBName)
				}
			}
		}

		if userID != "" && tenantID != "" && cfg.Storage != nil {
			if claims != nil && claims.TenantID != nil && *claims.TenantID == tenantID && claims.TenantRole != "" {
				c.Locals("tenant_role", claims.TenantRole)
			}
		}

		log.Debug().
			Str("tenant_id", tenantID).
			Str("tenant_source", tenantSource).
			Str("user_id", userID).
			Str("path", c.Path()).
			Bool("uses_main_db", pool == nil).
			Msg("TenantDBMiddleware: Set tenant context")

		return c.Next()
	}
}

func resolveTenantID(c fiber.Ctx, userID string, isInstanceAdmin bool, claims *auth.TokenClaims, storage *tenantdb.Storage) (string, string) {
	if headerTenant := c.Get("X-FB-Tenant"); headerTenant != "" {
		if storage != nil {
			if tenant, err := storage.GetTenantBySlug(c.Context(), headerTenant); err == nil {
				return tenant.ID, "header"
			}
			if _, err := storage.GetTenant(c.Context(), headerTenant); err == nil {
				return headerTenant, "header"
			}
		}
		return headerTenant, "header"
	}

	if claims != nil && claims.TenantID != nil && *claims.TenantID != "" {
		return *claims.TenantID, "jwt"
	}

	if storage != nil {
		if tenant, err := storage.GetDefaultTenant(c.Context()); err == nil {
			return tenant.ID, "default"
		}
	}

	return "", ""
}

func GetTenantPool(c fiber.Ctx) *pgxpool.Pool {
	pool, _ := c.Locals("tenant_db").(*pgxpool.Pool)
	return pool
}

// GetPoolForSchema returns the appropriate database pool based on the target schema.
// Priority: branch pool > tenant pool (for public schema) > main pool.
// This enables tenant-aware routing for REST, GraphQL, and other handlers.
func GetPoolForSchema(c fiber.Ctx, schema string, mainPool *pgxpool.Pool) *pgxpool.Pool {
	// 1. Branch pool takes highest priority (for database branching feature)
	if pool := GetBranchPool(c); pool != nil {
		return pool
	}

	// 2. For public schema (user data), route to tenant pool when available
	if schema == "public" {
		if tenantPool := GetTenantPool(c); tenantPool != nil {
			return tenantPool
		}
	}

	// 3. Fall back to main pool for internal schemas or when no tenant pool exists
	return mainPool
}

// SetTargetSchema stores the target schema in fiber locals for pool routing.
// Handlers should call this before WrapWithRLS to enable schema-aware pool selection.
func SetTargetSchema(c fiber.Ctx, schema string) {
	c.Locals("target_schema", schema)
}

// GetTargetSchema retrieves the target schema from fiber locals.
// Returns empty string as default if not set. Handlers must explicitly call
// SetTargetSchema to enable tenant-aware pool routing. This prevents internal
// schema handlers (auth.*, app.*, etc.) from inadvertently routing to tenant pools.
func GetTargetSchema(c fiber.Ctx) string {
	if schema, ok := c.Locals("target_schema").(string); ok && schema != "" {
		return schema
	}
	return ""
}

func SetTenantDBSessionContext(ctx context.Context, tx pgxpool.Tx, tenantID string) error {
	if tenantID == "" {
		return nil
	}
	_, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
	if err != nil {
		return fmt.Errorf("failed to set tenant session context: %w", err)
	}
	return nil
}

func RequireTenantAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		isInstanceAdminVal, _ := c.Locals("is_instance_admin").(bool)
		tenantSource, _ := c.Locals("tenant_source").(string)
		tenantID, _ := c.Locals("tenant_id").(string)

		actingAsTenantAdmin := tenantID != "" && (tenantSource == "header" || tenantSource == "jwt")

		if isInstanceAdminVal && !actingAsTenantAdmin {
			return c.Next()
		}

		tenantRole, _ := c.Locals("tenant_role").(string)
		if tenantRole == "" {
			return fiber.NewError(fiber.StatusForbidden, "tenant membership required")
		}

		if tenantRole != "tenant_admin" {
			return fiber.NewError(fiber.StatusForbidden, "tenant admin role required")
		}

		return c.Next()
	}
}

func GetUserManagedTenantIDs(ctx context.Context, storage *tenantdb.Storage, userID string) ([]string, error) {
	if storage == nil {
		return nil, errors.New("storage not initialized")
	}

	tenants, err := storage.GetTenantsForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tenants: %w", err)
	}

	ids := make([]string, len(tenants))
	for i, t := range tenants {
		ids[i] = t.ID
	}
	return ids, nil
}
