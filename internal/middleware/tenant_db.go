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
