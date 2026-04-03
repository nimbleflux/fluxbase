package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/settings"
)

// TenantConfigResolver resolves tenant-specific configuration at request time.
// It merges settings from multiple sources with NO caching to ensure
// immediate visibility of database changes.
//
// Configuration Cascade (Priority: Low → High):
//  1. Hardcoded defaults (code)
//  2. Config file (fluxbase.yaml)
//  3. Instance settings (database: platform.instance_settings)
//  4. Tenant settings (database: platform.tenant_settings)
type TenantConfigResolver struct {
	db              *database.Connection
	baseConfig      *config.Config
	unifiedSettings *settings.UnifiedService
}

// NewTenantConfigResolver creates a new tenant config resolver.
func NewTenantConfigResolver(
	db *database.Connection,
	baseConfig *config.Config,
	unifiedSettings *settings.UnifiedService,
) *TenantConfigResolver {
	return &TenantConfigResolver{
		db:              db,
		baseConfig:      baseConfig,
		unifiedSettings: unifiedSettings,
	}
}

// ResolvedConfig contains fully-resolved per-feature configuration.
// All configs are copies and can be safely modified.
type ResolvedConfig struct {
	Auth      config.AuthConfig
	Storage   config.StorageConfig
	Email     config.EmailConfig
	Functions config.FunctionsConfig
	Jobs      config.JobsConfig
	AI        config.AIConfig
	Realtime  config.RealtimeConfig
	RPC       config.RPCConfig
	GraphQL   config.GraphQLConfig
	API       config.APIConfig
}

// ResolveForRequest merges all configuration layers for the current request.
// This method does NOT cache results - every call fetches fresh data from the database
// to ensure immediate visibility of setting changes.
//
// For the default tenant, the cascade is: baseConfig (YAML+env) → instance DB → tenant DB.
// For non-default tenants, the cascade is: instance DB → tenant DB (no YAML/env layer).
func (r *TenantConfigResolver) ResolveForRequest(ctx context.Context, c fiber.Ctx) *ResolvedConfig {
	// Get tenant ID from context or fiber locals
	tenantID := r.getTenantID(c)

	// Determine if this is the default tenant
	isDefaultTenant := false
	if c != nil {
		isDefaultTenant, _ = c.Locals("is_default_tenant").(bool)
	}

	var resolved *ResolvedConfig
	if isDefaultTenant {
		// Default tenant: start with base config (YAML + env)
		resolved = r.resolvedFromBaseConfig()
	} else {
		// Non-default tenant: zero config — no YAML/env layer
		resolved = &ResolvedConfig{}
	}

	// Apply instance-level settings from database
	r.applyInstanceSettings(ctx, resolved)

	// Apply tenant-level settings from database (if tenant context exists)
	if tenantID != "" {
		r.applyTenantSettings(ctx, tenantID, resolved)
	}

	return resolved
}

// ResolveForTenant resolves configuration for a specific tenant ID.
// This is used by background workers (jobs) that don't have a fiber context.
func (r *TenantConfigResolver) ResolveForTenant(ctx context.Context, tenantID string, isDefaultTenant bool) *ResolvedConfig {
	var resolved *ResolvedConfig
	if isDefaultTenant {
		// Default tenant: start with base config (YAML + env)
		resolved = r.resolvedFromBaseConfig()
	} else {
		// Non-default tenant: zero config — no YAML/env layer
		resolved = &ResolvedConfig{}
	}

	// Apply instance-level settings from database
	r.applyInstanceSettings(ctx, resolved)

	// Apply tenant-level settings from database
	if tenantID != "" {
		r.applyTenantSettings(ctx, tenantID, resolved)
	}

	return resolved
}

// getTenantID extracts tenant ID from fiber context.
func (r *TenantConfigResolver) getTenantID(c fiber.Ctx) string {
	if c == nil {
		return ""
	}
	if id, ok := c.Locals("tenant_id").(string); ok {
		return id
	}
	return ""
}

// applyInstanceSettings applies instance-level database settings to the resolved config.
func (r *TenantConfigResolver) applyInstanceSettings(ctx context.Context, resolved *ResolvedConfig) {
	if r.unifiedSettings == nil {
		return
	}

	// Get instance settings from database
	instanceSettings, err := r.unifiedSettings.GetInstanceSettings(ctx)
	if err != nil {
		log.Debug().Err(err).Msg("Failed to get instance settings, using base config")
		return
	}

	if instanceSettings == nil || instanceSettings.Settings == nil {
		return
	}

	// Apply each setting category
	r.applySettingsToConfig(instanceSettings.Settings, resolved, "instance")
}

// applyTenantSettings applies tenant-level database settings to the resolved config.
// Only applies settings that are marked as overridable at the instance level.
func (r *TenantConfigResolver) applyTenantSettings(ctx context.Context, tenantID string, resolved *ResolvedConfig) {
	if r.unifiedSettings == nil {
		return
	}

	// Get tenant settings from database
	tenantSettings, err := r.unifiedSettings.GetTenantSettings(ctx, tenantID)
	if err != nil {
		log.Debug().Err(err).Str("tenant_id", tenantID).Msg("Failed to get tenant settings")
		return
	}

	if tenantSettings == nil || tenantSettings.Settings == nil {
		return
	}

	// Apply each setting category (respecting overridable settings)
	r.applySettingsToConfig(tenantSettings.Settings, resolved, "tenant")
}

// applySettingsToConfig applies a settings map to the resolved config.
func (r *TenantConfigResolver) applySettingsToConfig(settingsMap map[string]any, resolved *ResolvedConfig, source string) {
	// Apply AI settings
	if aiSettings, ok := settingsMap["ai"].(map[string]any); ok {
		r.applyAISettings(aiSettings, &resolved.AI, source)
	}

	// Apply Auth settings
	if authSettings, ok := settingsMap["auth"].(map[string]any); ok {
		r.applyAuthSettings(authSettings, &resolved.Auth, source)
	}

	// Apply Storage settings
	if storageSettings, ok := settingsMap["storage"].(map[string]any); ok {
		r.applyStorageSettings(storageSettings, &resolved.Storage, source)
	}

	// Apply Email settings
	if emailSettings, ok := settingsMap["email"].(map[string]any); ok {
		r.applyEmailSettings(emailSettings, &resolved.Email, source)
	}

	// Apply Functions settings
	if functionsSettings, ok := settingsMap["functions"].(map[string]any); ok {
		r.applyFunctionsSettings(functionsSettings, &resolved.Functions, source)
	}

	// Apply Jobs settings
	if jobsSettings, ok := settingsMap["jobs"].(map[string]any); ok {
		r.applyJobsSettings(jobsSettings, &resolved.Jobs, source)
	}

	// Apply Realtime settings
	if realtimeSettings, ok := settingsMap["realtime"].(map[string]any); ok {
		r.applyRealtimeSettings(realtimeSettings, &resolved.Realtime, source)
	}

	// Apply RPC settings
	if rpcSettings, ok := settingsMap["rpc"].(map[string]any); ok {
		r.applyRPCSettings(rpcSettings, &resolved.RPC, source)
	}

	// Apply GraphQL settings
	if graphqlSettings, ok := settingsMap["graphql"].(map[string]any); ok {
		r.applyGraphQLSettings(graphqlSettings, &resolved.GraphQL, source)
	}

	// Apply API settings
	if apiSettings, ok := settingsMap["api"].(map[string]any); ok {
		r.applyAPISettings(apiSettings, &resolved.API, source)
	}
}

// Settings application helpers for each config section

func (r *TenantConfigResolver) applyAISettings(settings map[string]any, cfg *config.AIConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["default_model"].(string); ok {
		cfg.DefaultModel = v
	}
	if v, ok := settings["default_max_tokens"].(float64); ok {
		cfg.DefaultMaxTokens = int(v)
	}
	if v, ok := settings["provider_type"].(string); ok {
		cfg.ProviderType = v
	}
	if v, ok := settings["provider_name"].(string); ok {
		cfg.ProviderName = v
	}
	if v, ok := settings["provider_model"].(string); ok {
		cfg.ProviderModel = v
	}
	if v, ok := settings["embedding_enabled"].(bool); ok {
		cfg.EmbeddingEnabled = v
	}
	if v, ok := settings["embedding_provider"].(string); ok {
		cfg.EmbeddingProvider = v
	}
	if v, ok := settings["embedding_model"].(string); ok {
		cfg.EmbeddingModel = v
	}
	if v, ok := settings["openai_api_key"].(string); ok {
		cfg.OpenAIAPIKey = v
	}
	if v, ok := settings["openai_base_url"].(string); ok {
		cfg.OpenAIBaseURL = v
	}
	if v, ok := settings["azure_api_key"].(string); ok {
		cfg.AzureAPIKey = v
	}
	if v, ok := settings["azure_endpoint"].(string); ok {
		cfg.AzureEndpoint = v
	}
}

func (r *TenantConfigResolver) applyAuthSettings(settings map[string]any, cfg *config.AuthConfig, source string) {
	if v, ok := settings["signup_enabled"].(bool); ok {
		cfg.SignupEnabled = v
	}
	if v, ok := settings["magic_link_enabled"].(bool); ok {
		cfg.MagicLinkEnabled = v
	}
	if v, ok := settings["jwt_expiry"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.JWTExpiry = d
		}
	}
	if v, ok := settings["jwt_expiry"].(float64); ok {
		cfg.JWTExpiry = time.Duration(v) * time.Second
	}
	if v, ok := settings["refresh_expiry"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.RefreshExpiry = d
		}
	}
	if v, ok := settings["refresh_expiry"].(float64); ok {
		cfg.RefreshExpiry = time.Duration(v) * time.Second
	}
}

func (r *TenantConfigResolver) applyStorageSettings(settings map[string]any, cfg *config.StorageConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["provider"].(string); ok {
		cfg.Provider = v
	}
	if v, ok := settings["max_upload_size"].(float64); ok {
		cfg.MaxUploadSize = int64(v)
	}
	if v, ok := settings["local_path"].(string); ok {
		cfg.LocalPath = v
	}
	if v, ok := settings["s3_bucket"].(string); ok {
		cfg.S3Bucket = v
	}
	if v, ok := settings["s3_region"].(string); ok {
		cfg.S3Region = v
	}
	if v, ok := settings["s3_endpoint"].(string); ok {
		cfg.S3Endpoint = v
	}
	if v, ok := settings["s3_access_key"].(string); ok {
		cfg.S3AccessKey = v
	}
	if v, ok := settings["s3_secret_key"].(string); ok {
		cfg.S3SecretKey = v
	}
}

func (r *TenantConfigResolver) applyEmailSettings(settings map[string]any, cfg *config.EmailConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["provider"].(string); ok {
		cfg.Provider = v
	}
	if v, ok := settings["from_address"].(string); ok {
		cfg.FromAddress = v
	}
	if v, ok := settings["from_name"].(string); ok {
		cfg.FromName = v
	}
	if v, ok := settings["reply_to_address"].(string); ok {
		cfg.ReplyToAddress = v
	}
	if v, ok := settings["smtp_host"].(string); ok {
		cfg.SMTPHost = v
	}
	if v, ok := settings["smtp_port"].(float64); ok {
		cfg.SMTPPort = int(v)
	}
	if v, ok := settings["smtp_username"].(string); ok {
		cfg.SMTPUsername = v
	}
	if v, ok := settings["smtp_password"].(string); ok {
		cfg.SMTPPassword = v
	}
	if v, ok := settings["smtp_tls"].(bool); ok {
		cfg.SMTPTLS = v
	}
	if v, ok := settings["sendgrid_api_key"].(string); ok {
		cfg.SendGridAPIKey = v
	}
	if v, ok := settings["mailgun_api_key"].(string); ok {
		cfg.MailgunAPIKey = v
	}
	if v, ok := settings["mailgun_domain"].(string); ok {
		cfg.MailgunDomain = v
	}
	if v, ok := settings["ses_access_key"].(string); ok {
		cfg.SESAccessKey = v
	}
	if v, ok := settings["ses_secret_key"].(string); ok {
		cfg.SESSecretKey = v
	}
	if v, ok := settings["ses_region"].(string); ok {
		cfg.SESRegion = v
	}
}

func (r *TenantConfigResolver) applyFunctionsSettings(settings map[string]any, cfg *config.FunctionsConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["default_timeout"].(float64); ok {
		cfg.DefaultTimeout = int(v)
	}
	if v, ok := settings["max_timeout"].(float64); ok {
		cfg.MaxTimeout = int(v)
	}
	if v, ok := settings["default_memory_limit"].(float64); ok {
		cfg.DefaultMemoryLimit = int(v)
	}
	if v, ok := settings["max_memory_limit"].(float64); ok {
		cfg.MaxMemoryLimit = int(v)
	}
}

func (r *TenantConfigResolver) applyJobsSettings(settings map[string]any, cfg *config.JobsConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["embedded_worker_count"].(float64); ok {
		cfg.EmbeddedWorkerCount = int(v)
	}
	if v, ok := settings["max_concurrent_per_worker"].(float64); ok {
		cfg.MaxConcurrentPerWorker = int(v)
	}
	if v, ok := settings["default_max_duration"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultMaxDuration = d
		}
	}
	if v, ok := settings["default_max_duration"].(float64); ok {
		cfg.DefaultMaxDuration = time.Duration(v) * time.Second
	}
	if v, ok := settings["poll_interval"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PollInterval = d
		}
	}
	if v, ok := settings["poll_interval"].(float64); ok {
		cfg.PollInterval = time.Duration(v) * time.Second
	}
}

func (r *TenantConfigResolver) applyRealtimeSettings(settings map[string]any, cfg *config.RealtimeConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["max_connections"].(float64); ok {
		cfg.MaxConnections = int(v)
	}
	if v, ok := settings["max_connections_per_user"].(float64); ok {
		cfg.MaxConnectionsPerUser = int(v)
	}
	if v, ok := settings["ping_interval"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.PingInterval = d
		}
	}
	if v, ok := settings["ping_interval"].(float64); ok {
		cfg.PingInterval = time.Duration(v) * time.Second
	}
}

func (r *TenantConfigResolver) applyRPCSettings(settings map[string]any, cfg *config.RPCConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["default_max_execution_time"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.DefaultMaxExecutionTime = d
		}
	}
	if v, ok := settings["default_max_execution_time"].(float64); ok {
		cfg.DefaultMaxExecutionTime = time.Duration(v) * time.Second
	}
	if v, ok := settings["default_max_rows"].(float64); ok {
		cfg.DefaultMaxRows = int(v)
	}
}

func (r *TenantConfigResolver) applyGraphQLSettings(settings map[string]any, cfg *config.GraphQLConfig, source string) {
	if v, ok := settings["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := settings["max_depth"].(float64); ok {
		cfg.MaxDepth = int(v)
	}
	if v, ok := settings["max_complexity"].(float64); ok {
		cfg.MaxComplexity = int(v)
	}
}

func (r *TenantConfigResolver) applyAPISettings(settings map[string]any, cfg *config.APIConfig, source string) {
	if v, ok := settings["max_page_size"].(float64); ok {
		cfg.MaxPageSize = int(v)
	}
	if v, ok := settings["max_total_results"].(float64); ok {
		cfg.MaxTotalResults = int(v)
	}
	if v, ok := settings["default_page_size"].(float64); ok {
		cfg.DefaultPageSize = int(v)
	}
	if v, ok := settings["max_batch_size"].(float64); ok {
		cfg.MaxBatchSize = int(v)
	}
}

// resolvedFromBaseConfig creates a ResolvedConfig seeded from the YAML/env base config.
func (r *TenantConfigResolver) resolvedFromBaseConfig() *ResolvedConfig {
	return &ResolvedConfig{
		Auth:      *config.DeepCopyAuthConfig(&r.baseConfig.Auth),
		Storage:   *config.DeepCopyStorageConfig(&r.baseConfig.Storage),
		Email:     *config.DeepCopyEmailConfig(&r.baseConfig.Email),
		Functions: *config.DeepCopyFunctionsConfig(&r.baseConfig.Functions),
		Jobs:      *config.DeepCopyJobsConfig(&r.baseConfig.Jobs),
		AI:        *config.DeepCopyAIConfig(&r.baseConfig.AI),
		Realtime:  *config.DeepCopyRealtimeConfig(&r.baseConfig.Realtime),
		RPC:       *config.DeepCopyRPCConfig(&r.baseConfig.RPC),
		GraphQL:   *config.DeepCopyGraphQLConfig(&r.baseConfig.GraphQL),
		API:       *config.DeepCopyAPIConfig(&r.baseConfig.API),
	}
}

// GetBaseConfig returns the base configuration (for reference only).
// Use ResolveForRequest() to get tenant-specific config.
func (r *TenantConfigResolver) GetBaseConfig() *config.Config {
	return r.baseConfig
}
