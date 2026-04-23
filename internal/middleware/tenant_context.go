package middleware

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/database"
)

// CtxWithTenant wraps the request context with the tenant ID from fiber locals.
// It prefers JWT claims (set by auth middleware) over TenantMiddleware's default
// fallback, since TenantMiddleware runs before auth and can't read JWT claims.
func CtxWithTenant(c fiber.Ctx) context.Context {
	tenantID := GetTenantIDFromContext(c)
	tenantSource := GetTenantSourceFromContext(c)

	if tenantSource == "default" || tenantID == "" {
		var claims *auth.TokenClaims
		if cl, ok := c.Locals("claims").(*auth.TokenClaims); ok && cl != nil {
			claims = cl
		} else if cl, ok := c.Locals("jwt_claims").(*auth.TokenClaims); ok && cl != nil {
			claims = cl
		}
		if claims != nil && claims.TenantID != nil {
			tenantID = *claims.TenantID
		}
	}
	return database.ContextWithTenant(c.RequestCtx(), tenantID)
}
