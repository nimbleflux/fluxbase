package api

import (
	"github.com/gofiber/fiber/v3"
	"github.com/nimbleflux/fluxbase/internal/config"
)

// GetTenantConfig returns the tenant-specific configuration if available,
// otherwise returns the base configuration.
// This is the primary function handlers should use to get configuration.
func GetTenantConfig(c fiber.Ctx, baseConfig *config.Config) *config.Config {
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return tc
	}
	return baseConfig
}

// GetTenantConfigFromLocals returns only the tenant-specific config from context.
// Returns nil if no tenant config is set.
func GetTenantConfigFromLocals(c fiber.Ctx) *config.Config {
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok {
		return tc
	}
	return nil
}

// GetTenantID returns the current tenant ID from context.
// Returns empty string if no tenant is set.
func GetTenantID(c fiber.Ctx) string {
	if id, ok := c.Locals("tenant_id").(string); ok {
		return id
	}
	return ""
}

// GetTenantSlug returns the current tenant slug from context.
// Returns empty string if no tenant slug is set.
func GetTenantSlug(c fiber.Ctx) string {
	if slug, ok := c.Locals("tenant_slug").(string); ok {
		return slug
	}
	return ""
}

// GetTenantSource returns where the tenant context came from.
// Returns empty string if no tenant source is set.
// Possible values: "header", "jwt", "default"
func GetTenantSource(c fiber.Ctx) string {
	if source, ok := c.Locals("tenant_source").(string); ok {
		return source
	}
	return ""
}

// GetTenantRole returns the user's role in the current tenant.
// Returns empty string if no tenant role is set.
func GetTenantRole(c fiber.Ctx) string {
	if role, ok := c.Locals("tenant_role").(string); ok {
		return role
	}
	return ""
}

// IsInstanceAdmin returns true if the user is an instance-level admin.
func IsInstanceAdmin(c fiber.Ctx) bool {
	isAdmin, ok := c.Locals("is_instance_admin").(bool)
	return ok && isAdmin
}

// GetStorageConfig returns the storage config to use for the current request.
// If a tenant-specific config is available, it returns that; otherwise returns the base config.
// This is used by the storage manager to get the appropriate service.
func GetStorageConfig(c fiber.Ctx, baseConfig *config.Config) *config.StorageConfig {
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Storage
	}
	return &baseConfig.Storage
}
