package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/branching"
	apperrors "github.com/nimbleflux/fluxbase/internal/errors"
)

const (
	// BranchHeader is the HTTP header for specifying the branch
	BranchHeader = "X-Fluxbase-Branch"

	// BranchQueryParam is the query parameter for specifying the branch
	BranchQueryParam = "branch"

	// LocalsBranchSlug is the Fiber locals key for the branch slug
	LocalsBranchSlug = "branch_slug"

	// LocalsBranchPool is the Fiber locals key for the branch connection pool
	LocalsBranchPool = "branch_pool"

	// LocalsBranch is the Fiber locals key for the branch object
	LocalsBranch = "branch"
)

// BranchContextConfig holds configuration for the branch context middleware
type BranchContextConfig struct {
	// Router is the branch router for getting connection pools
	Router *branching.Router

	// RequireAccess determines if access checks should be performed
	// When true, authenticated users must have explicit access to non-main branches
	RequireAccess bool

	// AllowAnonymous determines if anonymous users can access branches
	// When true, anonymous users can only access the main branch
	AllowAnonymous bool
}

// BranchContext creates a middleware that extracts branch context from requests
// and sets up the appropriate connection pool
// Precedence: Header > Query param > API-set default > Config default > "main"
//
// The middleware is designed to run after the auth middleware (the route
// registry injects it in the post-auth position). Branch pools are resolved
// tenant-scoped: the tenant ID comes from the tenant middlewares' locals.
// Access to explicitly selected non-main branches is verified for non-privileged
// principals; instance admins and service keys bypass. Locals are only set for
// non-main branches so the main pool never shadows the tenant pool in
// GetPoolForSchema's priority chain.
func BranchContext(config BranchContextConfig) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Extract branch slug from header or query param
		branchSlug := c.Get(BranchHeader)
		if branchSlug == "" {
			branchSlug = c.Query(BranchQueryParam)
		}

		// Whether the branch was explicitly selected on this request. Server-wide
		// defaults (API-set active branch / config) are admin decisions and are
		// not subject to per-user access checks.
		explicitBranch := branchSlug != ""

		// If no per-request branch specified, use the router's default branch
		// This considers: API-set active branch > Config default > "main"
		if branchSlug == "" {
			if config.Router != nil {
				branchSlug = config.Router.GetDefaultBranch()
			} else {
				branchSlug = "main"
			}
		}

		// Resolve tenant scope for branch pool lookup (set by the tenant middlewares)
		tenantID, _ := c.Locals("tenant_id").(string)

		// For non-main branches, check access
		if !branching.IsMainBranch(branchSlug) {
			// Check if branching is enabled
			if config.Router == nil {
				return apperrors.SendErrorWithCode(c, fiber.StatusServiceUnavailable, "Database branching is not enabled", apperrors.ErrCodeBranchingDisabled)
			}

			// Privileged principals (instance admins, service keys, service-role
			// JWTs) bypass per-user branch access checks.
			if isBranchPrivileged(c) {
				// bypass access check below
			} else if config.RequireAccess {
				// Check authentication for non-main branches
				userID := parseBranchUserID(c)
				if userID == nil && !config.AllowAnonymous {
					return apperrors.SendErrorWithCode(c, fiber.StatusUnauthorized, "Authentication is required to access branches", apperrors.ErrCodeAuthRequired)
				}

				// Check access for authenticated users on explicit branch selections
				if userID != nil && explicitBranch {
					hasAccess, err := config.Router.UserHasAccessForBranch(c.RequestCtx(), tenantID, branchSlug, userID.String())
					if err != nil {
						log.Error().Err(err).
							Str("branch", branchSlug).
							Str("user_id", userID.String()).
							Msg("Failed to check branch access")
						return apperrors.SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to verify branch access", apperrors.ErrCodeAccessCheckFailed)
					}

					if !hasAccess {
						return apperrors.SendErrorWithCode(c, fiber.StatusForbidden, "You do not have access to this branch", apperrors.ErrCodeAccessDenied)
					}
				}
			}
		}

		// Get connection pool for the branch
		var pool *pgxpool.Pool
		var err error

		if config.Router != nil {
			pool, err = config.Router.GetPoolForBranch(c.RequestCtx(), tenantID, branchSlug)
			if err != nil {
				if errors.Is(err, branching.ErrBranchNotFound) {
					return apperrors.SendErrorWithCode(c, fiber.StatusNotFound, "Branch not found: "+branchSlug, apperrors.ErrCodeBranchNotFound)
				}
				if errors.Is(err, branching.ErrBranchNotReady) {
					return apperrors.SendErrorWithCode(c, fiber.StatusServiceUnavailable, "Branch is not ready: "+branchSlug, apperrors.ErrCodeBranchNotReady)
				}
				if errors.Is(err, branching.ErrBranchingDisabled) {
					// For main branch, we should still work
					if branching.IsMainBranch(branchSlug) {
						pool = config.Router.GetMainPool()
					} else {
						return apperrors.SendErrorWithCode(c, fiber.StatusServiceUnavailable, "Database branching is not enabled", apperrors.ErrCodeBranchingDisabled)
					}
				} else {
					log.Error().Err(err).
						Str("branch", branchSlug).
						Msg("Failed to get branch pool")
					return apperrors.SendErrorWithCode(c, fiber.StatusInternalServerError, "Failed to get database connection for branch", apperrors.ErrCodePoolError)
				}
			}
		}

		// Store branch context in locals — only for non-main branches. Setting the
		// pool for the main branch would shadow the tenant pool in
		// GetPoolForSchema (branch pool has highest priority there).
		if !branching.IsMainBranch(branchSlug) {
			c.Locals(LocalsBranchSlug, branchSlug)
			if pool != nil {
				c.Locals(LocalsBranchPool, pool)
			}
		}

		// Log branch context for debugging
		if branchSlug != "main" {
			log.Debug().
				Str("branch", branchSlug).
				Str("path", c.Path()).
				Msg("Request using branch database")
		}

		return c.Next()
	}
}

// isBranchPrivileged reports whether the authenticated principal bypasses
// per-user branch access checks: instance admins, service keys, and
// service-role JWTs.
func isBranchPrivileged(c fiber.Ctx) bool {
	if isInstanceAdmin, ok := c.Locals("is_instance_admin").(bool); ok && isInstanceAdmin {
		return true
	}
	switch authType, _ := c.Locals("auth_type").(string); authType {
	case "service_key", "service_role_jwt":
		return true
	}
	switch userRole, _ := c.Locals("user_role").(string); userRole {
	case "service_role", "instance_admin":
		return true
	}
	return false
}

// parseBranchUserID extracts the authenticated user ID from fiber locals.
// Returns nil for unauthenticated requests.
func parseBranchUserID(c fiber.Ctx) *uuid.UUID {
	if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			return &id
		}
	}
	return nil
}

// GetBranchSlug extracts the branch slug from Fiber context
func GetBranchSlug(c fiber.Ctx) string {
	if slug, ok := c.Locals(LocalsBranchSlug).(string); ok {
		return slug
	}
	return "main"
}

// GetBranchPool extracts the branch connection pool from Fiber context
func GetBranchPool(c fiber.Ctx) *pgxpool.Pool {
	if pool, ok := c.Locals(LocalsBranchPool).(*pgxpool.Pool); ok {
		return pool
	}
	return nil
}

// IsUsingBranch checks if the request is using a non-main branch
func IsUsingBranch(c fiber.Ctx) bool {
	slug := GetBranchSlug(c)
	return !branching.IsMainBranch(slug)
}

// BranchContextSimple creates a simple middleware that only sets branch context
// without access checks (useful for internal routes)
func BranchContextSimple(router *branching.Router) fiber.Handler {
	return BranchContext(BranchContextConfig{
		Router:         router,
		RequireAccess:  false,
		AllowAnonymous: true,
	})
}

// RequireBranchAccess creates a middleware that requires branch access
// This should be used after authentication middleware
func RequireBranchAccess(router *branching.Router) fiber.Handler {
	return BranchContext(BranchContextConfig{
		Router:         router,
		RequireAccess:  true,
		AllowAnonymous: false,
	})
}

// fiber:context-methods migrated
