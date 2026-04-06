package api

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// Global resolver instance - set by server during initialization
var globalResolver *TenantConfigResolver

// SetGlobalResolver sets the global tenant config resolver.
// This should be called once during server initialization.
func SetGlobalResolver(resolver *TenantConfigResolver) {
	globalResolver = resolver
}

// GetGlobalResolver returns the global tenant config resolver.
// Returns nil if not set.
func GetGlobalResolver() *TenantConfigResolver {
	return globalResolver
}

// GetTenantConfig returns the tenant-specific configuration if available,
// otherwise returns the base configuration.
// This is the primary function handlers should use to get configuration.
//
// If a TenantConfigResolver is available, this will resolve config from the
// database with immediate visibility of changes (no caching).
func GetTenantConfig(c fiber.Ctx, baseConfig *config.Config) *config.Config {
	// If resolver is available, use it for full database resolution
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		// Convert resolved config to full config for compatibility
		return resolvedToFullConfig(resolved, baseConfig)
	}

	// Fallback to existing behavior (YAML-based tenant config)
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return tc
	}
	return baseConfig
}

// resolvedToFullConfig converts a ResolvedConfig to a full Config for compatibility.
func resolvedToFullConfig(resolved *ResolvedConfig, baseConfig *config.Config) *config.Config {
	// Start with a copy of base config
	cfg := *baseConfig
	cfg.Auth = resolved.Auth
	cfg.Storage = resolved.Storage
	cfg.Email = resolved.Email
	cfg.Functions = resolved.Functions
	cfg.Jobs = resolved.Jobs
	cfg.AI = resolved.AI
	cfg.Realtime = resolved.Realtime
	cfg.RPC = resolved.RPC
	cfg.GraphQL = resolved.GraphQL
	cfg.API = resolved.API
	return &cfg
}

// GetTenantConfigFromLocals returns only the tenant-specific config from context.
// Returns nil if no tenant config is set.
// Note: This does NOT use the resolver - it only returns the YAML-based tenant config.
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

// =============================================================================
// Per-Feature Config Getters
// These functions resolve tenant-specific config for each feature module.
// =============================================================================

// GetStorageConfig returns the storage config to use for the current request.
// If a tenant-specific config is available, it returns that; otherwise returns the base config.
// This is used by the storage manager to get the appropriate service.
func GetStorageConfig(c fiber.Ctx, baseConfig *config.Config) *config.StorageConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Storage
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Storage
	}
	if baseConfig != nil {
		return &baseConfig.Storage
	}
	return nil
}

// GetAuthConfig returns the auth config to use for the current request.
func GetAuthConfig(c fiber.Ctx, baseConfig *config.Config) *config.AuthConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Auth
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Auth
	}
	return &baseConfig.Auth
}

// GetEmailConfig returns the email config to use for the current request.
func GetEmailConfig(c fiber.Ctx, baseConfig *config.Config) *config.EmailConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Email
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Email
	}
	return &baseConfig.Email
}

// GetFunctionsConfig returns the functions config to use for the current request.
func GetFunctionsConfig(c fiber.Ctx, baseConfig *config.Config) *config.FunctionsConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Functions
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Functions
	}
	return &baseConfig.Functions
}

// GetJobsConfig returns the jobs config to use for the current request.
func GetJobsConfig(c fiber.Ctx, baseConfig *config.Config) *config.JobsConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Jobs
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Jobs
	}
	return &baseConfig.Jobs
}

// GetAIConfig returns the AI config to use for the current request.
func GetAIConfig(c fiber.Ctx, baseConfig *config.Config) *config.AIConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.AI
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.AI
	}
	return &baseConfig.AI
}

// GetRealtimeConfig returns the realtime config to use for the current request.
func GetRealtimeConfig(c fiber.Ctx, baseConfig *config.Config) *config.RealtimeConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.Realtime
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.Realtime
	}
	return &baseConfig.Realtime
}

// GetRPCConfig returns the RPC config to use for the current request.
func GetRPCConfig(c fiber.Ctx, baseConfig *config.Config) *config.RPCConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.RPC
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.RPC
	}
	return &baseConfig.RPC
}

// GetGraphQLConfig returns the GraphQL config to use for the current request.
func GetGraphQLConfig(c fiber.Ctx, baseConfig *config.Config) *config.GraphQLConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.GraphQL
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.GraphQL
	}
	return &baseConfig.GraphQL
}

// GetAPIConfig returns the API config to use for the current request.
func GetAPIConfig(c fiber.Ctx, baseConfig *config.Config) *config.APIConfig {
	if globalResolver != nil {
		resolved := globalResolver.ResolveForRequest(context.Background(), c)
		return &resolved.API
	}
	if tc, ok := c.Locals("tenant_config").(*config.Config); ok && tc != nil {
		return &tc.API
	}
	return &baseConfig.API
}
