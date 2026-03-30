package api

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/storage/memory/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/adminui"
	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/branching"
	"github.com/nimbleflux/fluxbase/internal/config"
	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/email"
	"github.com/nimbleflux/fluxbase/internal/extensions"
	"github.com/nimbleflux/fluxbase/internal/functions"
	"github.com/nimbleflux/fluxbase/internal/jobs"
	"github.com/nimbleflux/fluxbase/internal/logging"
	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/nimbleflux/fluxbase/internal/mcp/custom"
	mcpresources "github.com/nimbleflux/fluxbase/internal/mcp/resources"
	mcptools "github.com/nimbleflux/fluxbase/internal/mcp/tools"
	"github.com/nimbleflux/fluxbase/internal/middleware"
	"github.com/nimbleflux/fluxbase/internal/migrations"
	"github.com/nimbleflux/fluxbase/internal/observability"
	"github.com/nimbleflux/fluxbase/internal/pubsub"
	"github.com/nimbleflux/fluxbase/internal/ratelimit"
	"github.com/nimbleflux/fluxbase/internal/realtime"
	"github.com/nimbleflux/fluxbase/internal/rpc"
	"github.com/nimbleflux/fluxbase/internal/scaling"
	"github.com/nimbleflux/fluxbase/internal/secrets"
	"github.com/nimbleflux/fluxbase/internal/settings"
	"github.com/nimbleflux/fluxbase/internal/storage"
	"github.com/nimbleflux/fluxbase/internal/tenantdb"
	"github.com/nimbleflux/fluxbase/internal/webhook"
)

// Server represents the HTTP server
type Server struct {
	// Core infrastructure
	app    *fiber.App
	config *config.Config
	db     *database.Connection
	tracer *observability.Tracer
	rest   *RESTHandler

	// Handler groups (organized by domain)
	Auth       *AuthHandlers
	Storage    *StorageHandlers
	AI         *AIHandlers
	Functions  *FunctionsHandlers
	Jobs       *JobsHandlers
	Realtime   *RealtimeHandlers
	MCP        *MCPHandlers
	Tenancy    *TenancyHandlers
	Branching  *BranchingHandlers
	Settings   *SettingsHandlers
	Webhook    *WebhookHandlers
	Logging    *LoggingHandlers
	Schema     *SchemaHandlers
	RPC        *RPCHandlers
	GraphQL    *GraphQLHandlers
	Extensions *ExtensionsHandlers
	Secrets    *SecretsHandlers
	Scaling    *ScalingHandlers
	Metrics    *MetricsComponents
	Email      *EmailHandlers
	Captcha    *CaptchaHandlers
	Monitoring *MonitoringHandlers
	Quota      *QuotaHandlers
	Middleware *MiddlewareComponents

	// SQL handler (standalone, used by SQL editor)
	sqlHandler *SQLHandler

	// Server-owned dependencies (instead of global singletons)
	rateLimiter ratelimit.Store
	pubSub      pubsub.PubSub

	// Shared storage for middleware (rate limiter, CSRF, etc.)
	// This prevents creating multiple GC goroutines from Fiber's memory.New()
	sharedMiddlewareStorage fiber.Storage

	// Test transaction support (for HTTP API tests with transaction isolation)
	// When set, HTTP requests use this transaction instead of the connection pool
	testTx pgx.Tx

	// Tenant configuration loader for multi-tenant config overrides
	tenantConfigLoader *config.TenantConfigLoader
}

// NewServer creates a new HTTP server
func NewServer(cfg *config.Config, db *database.Connection, version string) *Server {
	// Create Fiber app with config
	app := fiber.New(fiber.Config{
		ServerHeader:      "Fluxbase",
		AppName:           fmt.Sprintf("Fluxbase v%s", version),
		BodyLimit:         cfg.Server.BodyLimit,
		StreamRequestBody: true, // Required for chunked upload streaming
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ErrorHandler:      customErrorHandler,
	})

	// In debug mode, add no-cache headers to prevent browser from caching
	// connection failures during server restarts
	if cfg.Debug {
		app.Use(func(c fiber.Ctx) error {
			c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			c.Set("Pragma", "no-cache")
			c.Set("Expires", "0")
			return c.Next()
		})
	}

	// Initialize OpenTelemetry tracer
	tracerCfg := observability.TracerConfig{
		Enabled:     cfg.Tracing.Enabled,
		Endpoint:    cfg.Tracing.Endpoint,
		ServiceName: cfg.Tracing.ServiceName,
		Environment: cfg.Tracing.Environment,
		SampleRate:  cfg.Tracing.SampleRate,
		Insecure:    cfg.Tracing.Insecure,
	}
	tracer, err := observability.NewTracer(context.Background(), tracerCfg)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize OpenTelemetry tracer, tracing will be disabled")
	}

	// Initialize rate limit store based on scaling configuration
	rateLimitStore, err := ratelimit.NewStore(&cfg.Scaling, db.Pool())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize rate limit store, falling back to memory")
		rateLimitStore = nil
	} else {
		log.Info().Str("backend", cfg.Scaling.Backend).Msg("Rate limit store initialized")
	}

	// Initialize pub/sub for cross-instance communication
	ps, err := pubsub.NewPubSub(&cfg.Scaling, db.Pool())
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize pub/sub, cross-instance broadcasting disabled")
		ps = nil
	} else {
		log.Info().Str("backend", cfg.Scaling.Backend).Msg("Pub/sub initialized for cross-instance broadcasting")
	}

	// Initialize shared middleware storage to prevent multiple GC goroutines
	// Fiber's memory.New() spawns GC goroutines that cannot be stopped
	// By using a single shared storage, we only get one set of GC goroutines per server
	// In test mode, use a very long GC interval to effectively disable GC
	gcInterval := 10 * time.Minute
	if os.Getenv("FLUXBASE_TEST_MODE") == "1" {
		gcInterval = 24 * time.Hour
	}
	sharedMiddlewareStorage := memory.New(memory.Config{
		GCInterval: gcInterval,
	})

	// Initialize email manager (handles dynamic refresh from settings)
	// The settings cache and secrets service will be injected later once they're initialized
	emailManager := email.NewManager(&cfg.Email, nil, nil, cfg)
	// Get a service wrapper that delegates to the manager's current service
	emailService := emailManager.WrapAsService()

	// Initialize auth service (use public URL for user-facing links like magic links, password resets)
	authService := auth.NewService(db, &cfg.Auth, emailService, cfg.GetPublicBaseURL())

	// Set encryption key for TOTP secrets (uses the global encryption key)
	authService.SetEncryptionKey(cfg.EncryptionKey)

	// Initialize TOTP rate limiter to protect against brute force attacks
	totpRateLimiter := auth.NewTOTPRateLimiter(db.Pool(), auth.DefaultTOTPRateLimiterConfig())
	authService.SetTOTPRateLimiter(totpRateLimiter)

	// Initialize API key service
	// Settings cache will be injected after auth service is initialized to enable
	// the 'allow_user_client_keys' setting check during client key validation
	clientKeyService := auth.NewClientKeyService(db.Pool(), nil)

	// Initialize storage manager (use public URL for signed URLs that users will access)
	storageManager, err := storage.NewManager(&cfg.Storage, cfg.GetPublicBaseURL(), cfg.Auth.JWTSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize storage manager")
	}

	// Get base service for backward compatibility
	storageService := storageManager.GetBaseService()

	// Ensure default buckets exist
	if err := storageManager.EnsureDefaultBuckets(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure default buckets")
	}

	// Initialize central logging service
	var loggingService *logging.Service
	var loggingHandler *LoggingHandler
	var retentionService *logging.RetentionService
	if cfg.Logging.ConsoleEnabled || cfg.Logging.Backend != "" {
		loggingService, err = logging.New(&cfg.Logging, db, storageService.Provider, ps)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to initialize central logging service, continuing with default logging")
		} else {
			// Replace zerolog writer with the central logging writer
			log.Logger = log.Output(loggingService.Writer())
			log.Info().
				Str("backend", cfg.Logging.Backend).
				Bool("pubsub_enabled", cfg.Logging.PubSubEnabled).
				Int("batch_size", cfg.Logging.BatchSize).
				Msg("Central logging service initialized")

			// Log diagnostic info about log streaming capability
			log.Info().
				Bool("pubsub_enabled", cfg.Logging.PubSubEnabled).
				Bool("pubsub_available", ps != nil).
				Msg("Logging service streaming capability")

			// Test PubSub by publishing a test log (diagnostic)
			if cfg.Logging.PubSubEnabled && ps != nil {
				testLog := &storage.LogEntry{
					Category: storage.LogCategorySystem,
					Level:    storage.LogLevelInfo,
					Message:  "Log streaming test - system initialized",
					Fields:   map[string]any{"test": true, "component": "logging_diagnostic"},
				}
				loggingService.Log(context.Background(), testLog)
				log.Info().Msg("Published test log to verify streaming - check /admin/logs page")
			}

			// Create logging handler for API routes
			loggingHandler = NewLoggingHandler(loggingService)

			// Create retention cleanup service
			if cfg.Logging.RetentionEnabled {
				retentionService = logging.NewRetentionService(&cfg.Logging, loggingService.Storage())
			}
		}
	}

	// Initialize webhook service
	webhookService := webhook.NewWebhookService(db)
	// Allow private IPs in debug mode (for local testing with localhost webhooks)
	// SECURITY WARNING: This bypasses SSRF protection - NEVER enable debug mode in production!
	webhookService.AllowPrivateIPs = cfg.Debug
	if cfg.Debug {
		log.Warn().Msg("SECURITY: Debug mode enabled - webhook SSRF protection is DISABLED. Do NOT use in production!")
	}

	// Initialize webhook trigger service (4 workers)
	webhookTriggerService := webhook.NewTriggerService(db, webhookService, 4)

	// Initialize user management service (use public URL for password reset links, etc.)
	userMgmtService := auth.NewUserManagementService(
		auth.NewUserRepository(db),
		auth.NewSessionRepository(db),
		auth.NewPasswordHasherWithConfig(auth.PasswordHasherConfig{MinLength: cfg.Auth.PasswordMinLen, Cost: cfg.Auth.BcryptCost}),
		emailService,
		cfg.GetPublicBaseURL(),
	)

	// Create CAPTCHA service
	captchaService, err := auth.NewCaptchaService(&cfg.Security.Captcha)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize CAPTCHA service - CAPTCHA protection disabled")
		captchaService = nil
	}

	// Create handlers
	authHandler := NewAuthHandler(db.Pool(), authService, captchaService, cfg.GetPublicBaseURL())
	// Create dashboard JWT manager first (shared between auth service and handler)
	dashboardJWTManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, 24*time.Hour, 168*time.Hour)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create dashboard JWT manager")
	}
	dashboardAuthService := auth.NewDashboardAuthService(db, dashboardJWTManager, cfg.Auth.TOTPIssuer)
	systemSettingsService := auth.NewSystemSettingsService(db)
	adminAuthHandler := NewAdminAuthHandler(authService, auth.NewUserRepository(db), dashboardAuthService, systemSettingsService, cfg)
	// Note: dashboardAuthHandler is initialized later after samlService is created
	clientKeyHandler := NewClientKeyHandler(clientKeyService)
	storageHandler := NewStorageHandler(storageManager, db, cfg, &cfg.Storage.Transforms)
	webhookHandler := NewWebhookHandler(webhookService)

	// Initialize secrets storage and handler
	secretsStorage := secrets.NewStorage(db, cfg.EncryptionKey)
	secretsHandler := secrets.NewHandler(secretsStorage)

	userMgmtHandler := NewUserManagementHandler(userMgmtService, authService)
	invitationService := auth.NewInvitationService(db)
	invitationHandler := NewInvitationHandler(invitationService, dashboardAuthService, emailService, cfg.GetPublicBaseURL())
	ddlHandler := NewDDLHandler(db, nil) // schemaCache set after cache creation
	realtimeAdminHandler := NewRealtimeAdminHandler(db)
	serviceKeyHandler := NewServiceKeyHandler(db.Pool())

	// Initialize multi-tenancy components
	var tenantManager *tenantdb.Manager
	var tenantStorage *tenantdb.Storage
	if cfg.Tenants.Enabled {
		tenantStorage = tenantdb.NewStorage(db.Pool())
		dbURL := cfg.Database.RuntimeConnectionString()
		tenantCfg := tenantdb.Config{
			Enabled:        cfg.Tenants.Enabled,
			DatabasePrefix: cfg.Tenants.DatabasePrefix,
			MaxTenants:     cfg.Tenants.MaxTenants,
			Pool: tenantdb.PoolConfig{
				MaxTotalConnections: cfg.Tenants.Pool.MaxTotalConnections,
				EvictionAge:         cfg.Tenants.Pool.EvictionAge,
			},
			Migrations: tenantdb.MigrationsConfig{
				CheckInterval: cfg.Tenants.Migrations.CheckInterval,
				OnCreate:      cfg.Tenants.Migrations.OnCreate,
				OnAccess:      cfg.Tenants.Migrations.OnAccess,
				Background:    cfg.Tenants.Migrations.Background,
			},
		}
		tenantManager = tenantdb.NewManager(tenantStorage, tenantCfg, db.Pool(), dbURL)

		// Create tenant pool router for per-tenant database connections
		tenantRouter := tenantdb.NewRouter(tenantStorage, tenantCfg, db.Pool(), db.Pool(), dbURL)
		tenantRouter.SetManager(tenantManager)
		tenantManager.SetRouter(tenantRouter)

		log.Info().Msg("Multi-tenancy enabled")

		// Initialize tenant declarative schema service if configured
		// Uses pgschema CLI for proper diff-based schema management
		if cfg.Tenants.Declarative.Enabled && cfg.Tenants.Declarative.SchemaDir != "" {
			declarativeCfg := tenantdb.DeclarativeConfig{
				Enabled:          cfg.Tenants.Declarative.Enabled,
				SchemaDir:        cfg.Tenants.Declarative.SchemaDir,
				OnCreate:         cfg.Tenants.Declarative.OnCreate,
				OnStartup:        cfg.Tenants.Declarative.OnStartup,
				AllowDestructive: cfg.Tenants.Declarative.AllowDestructive,
			}
			declarativeSvc := tenantdb.NewDeclarativeService(
				declarativeCfg,
				"pgschema", // pgschema CLI path (must be in PATH)
				cfg.Database.Host,
				cfg.Database.Port,
				cfg.Database.AdminUser,
				cfg.Database.AdminPassword,
				db.Pool(),
			)
			tenantManager.SetDeclarativeService(declarativeSvc)
			tenantManager.SetDeclarativeConfig(declarativeCfg)
			log.Info().
				Str("schema_dir", cfg.Tenants.Declarative.SchemaDir).
				Bool("on_create", cfg.Tenants.Declarative.OnCreate).
				Bool("on_startup", cfg.Tenants.Declarative.OnStartup).
				Msg("Tenant declarative schema service initialized")

			// Apply schemas on startup if configured
			if cfg.Tenants.Declarative.OnStartup {
				go func() {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
					defer cancel()
					if err := tenantManager.ApplyDeclarativeSchemas(ctx); err != nil {
						log.Error().Err(err).Msg("Failed to apply tenant declarative schemas on startup")
					}
				}()
			}
		}
	}
	tenantHandler := NewTenantHandler(db, tenantManager, tenantStorage, invitationService, emailService, cfg)

	// Initialize unified settings service and handlers
	unifiedSettingsService := settings.NewUnifiedService(db, cfg, cfg.EncryptionKey)
	instanceSettingsHandler := NewInstanceSettingsHandler(unifiedSettingsService)
	tenantSettingsHandler := NewTenantSettingsHandler(unifiedSettingsService, tenantStorage)

	// Initialize tenant config resolver for request-time config resolution
	// This enables immediate visibility of database settings changes (no caching)
	tenantConfigResolver := NewTenantConfigResolver(db, cfg, unifiedSettingsService)
	SetGlobalResolver(tenantConfigResolver)
	log.Info().Msg("Tenant config resolver initialized for dynamic settings")

	oauthProviderHandler := NewOAuthProviderHandler(db.Pool(), authService.GetSettingsCache(), cfg.EncryptionKey, cfg.GetPublicBaseURL(), cfg.Auth.OAuthProviders)
	jwtManager, err := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry, cfg.Auth.RefreshExpiry)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create JWT manager")
	}
	// Use public URL for OAuth callbacks (these are redirects from external OAuth providers)
	oauthHandler := NewOAuthHandler(db.Pool(), authService, jwtManager, cfg.GetPublicBaseURL(), cfg.EncryptionKey, cfg.Auth.OAuthProviders)

	// Initialize SAML service and handler
	var samlService *auth.SAMLService
	var samlProviderHandler *SAMLProviderHandler
	var samlErr error
	samlService, samlErr = auth.NewSAMLService(db.Pool(), cfg.GetPublicBaseURL(), cfg.Auth.SAMLProviders)
	if samlErr != nil {
		log.Warn().Err(samlErr).Msg("Failed to initialize SAML service from config")
	}
	// Load SAML providers from database
	if samlService != nil {
		if err := samlService.LoadProvidersFromDB(context.Background()); err != nil {
			log.Warn().Err(err).Msg("Failed to load SAML providers from database")
		}
	}
	samlProviderHandler = NewSAMLProviderHandler(db.Pool(), samlService)
	// Initialize dashboard auth handler now that samlService and oauthHandler are available
	dashboardAuthHandler := NewDashboardAuthHandler(dashboardAuthService, dashboardJWTManager, db, samlService, emailService, cfg.GetPublicBaseURL(), cfg.EncryptionKey, oauthHandler)
	adminSessionHandler := NewAdminSessionHandler(auth.NewSessionRepository(db))
	systemSettingsHandler := NewSystemSettingsHandler(systemSettingsService, authService.GetSettingsCache())
	customSettingsService := settings.NewCustomSettingsService(db, cfg.EncryptionKey)
	customSettingsHandler := NewCustomSettingsHandler(customSettingsService)
	userSettingsHandler := NewUserSettingsHandler(db, customSettingsService)
	secretsService := settings.NewSecretsService(db, cfg.EncryptionKey)
	userSettingsHandler.SetSecretsService(secretsService)
	appSettingsHandler := NewAppSettingsHandler(systemSettingsService, authService.GetSettingsCache(), cfg)
	settingsHandler := NewSettingsHandler(db)
	emailTemplateHandler := NewEmailTemplateHandler(db, emailService)

	// Initialize email settings handler with settings cache for dynamic configuration
	emailSettingsHandler := NewEmailSettingsHandler(
		systemSettingsService,
		authService.GetSettingsCache(),
		emailManager,
		secretsService,
		&cfg.Email,
	)

	// Refresh email manager with settings cache and secrets service now that they're available
	emailManager.SetSettingsCache(authService.GetSettingsCache())
	emailManager.SetSecretsService(secretsService)
	if err := emailManager.RefreshFromSettings(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to refresh email service from settings on startup")
	}

	// Initialize captcha settings handler with settings cache for dynamic configuration
	captchaSettingsHandler := NewCaptchaSettingsHandler(
		systemSettingsService,
		authService.GetSettingsCache(),
		secretsService,
		&cfg.Security,
		captchaService,
	)

	// Refresh captcha service with settings from database on startup
	if captchaService != nil {
		if err := captchaService.ReloadFromSettings(context.Background(), authService.GetSettingsCache(), &cfg.Security); err != nil {
			log.Warn().Err(err).Msg("Failed to refresh captcha service from settings on startup")
		}
	}

	// Inject settings cache into client key service for 'allow_user_client_keys' setting
	clientKeyService.SetSettingsCache(authService.GetSettingsCache())

	// Encrypt any existing plaintext OAuth provider secrets
	if err := oauthProviderHandler.EncryptExistingSecrets(context.Background()); err != nil {
		log.Error().Err(err).Msg("Failed to encrypt existing OAuth provider secrets")
	}
	sqlHandler := NewSQLHandler(db.Pool(), authService)

	// Determine public URL for functions SDK client
	// For edge functions running inside the container, they should use the internal BaseURL
	// to communicate with the API server (faster, avoids external network hops)
	functionsInternalURL := cfg.BaseURL
	if functionsInternalURL == "" {
		functionsInternalURL = "http://localhost" + cfg.Server.Address
	}
	functionsHandler := functions.NewHandler(db, cfg.Functions.FunctionsDir, cfg.CORS, cfg.Auth.JWTSecret, functionsInternalURL, cfg.Deno.NpmRegistry, cfg.Deno.JsrRegistry, authService, loggingService, secretsStorage, cfg)
	functionsHandler.SetSettingsSecretsService(secretsService)
	functionsScheduler := functions.NewScheduler(db, cfg.Auth.JWTSecret, functionsInternalURL, secretsStorage)
	functionsHandler.SetScheduler(functionsScheduler)

	// Only create jobs components if jobs are enabled
	var jobsManager *jobs.Manager
	var jobsHandler *jobs.Handler
	var jobsScheduler *jobs.Scheduler
	if cfg.Jobs.Enabled {
		// Determine internal URL for jobs SDK client
		// Jobs run inside the container and should use the internal URL
		jobsInternalURL := cfg.BaseURL
		if jobsInternalURL == "" {
			// Fallback to server address
			jobsInternalURL = "http://localhost" + cfg.Server.Address
		}
		log.Info().
			Str("jobs_internal_url", jobsInternalURL).
			Bool("jwt_secret_set", cfg.Auth.JWTSecret != "").
			Msg("Initializing jobs manager with SDK credentials")
		jobsManager = jobs.NewManager(&cfg.Jobs, db, cfg.Auth.JWTSecret, jobsInternalURL, secretsStorage, cfg)
		jobsManager.SetSettingsSecretsService(secretsService)
		var err error
		jobsHandler, err = jobs.NewHandler(db, &cfg.Jobs, jobsManager, authService, loggingService, cfg.Deno.NpmRegistry, cfg.Deno.JsrRegistry)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize jobs handler")
		}
		// Create jobs scheduler for cron-based job execution
		jobsScheduler = jobs.NewScheduler(db)
		jobsHandler.SetScheduler(jobsScheduler)
	}

	// Create schema cache for dynamic REST API routing (5 minute TTL)
	schemaCache := database.NewSchemaCache(db.Inspector(), 5*time.Minute)
	// Configure PubSub for cross-instance cache invalidation
	if ps != nil {
		schemaCache.SetPubSub(ps)
		log.Info().Msg("Schema cache configured for cross-instance invalidation via pub/sub")
	}
	// Populate cache on startup
	if err := schemaCache.Refresh(context.Background()); err != nil {
		log.Warn().Err(err).Msg("Failed to populate schema cache on startup")
	} else {
		log.Info().Int("tables", schemaCache.TableCount()).Int("views", schemaCache.ViewCount()).Msg("Schema cache populated")
	}

	migrationsHandler := migrations.NewHandler(db, schemaCache)

	// Wire schema cache to DDL handler for invalidation after DDL operations
	ddlHandler.SetSchemaCache(schemaCache)

	if tenantManager != nil && tenantManager.GetRouter() != nil {
		migrationsHandler.SetTenantPoolProvider(tenantManager.GetRouter())
	}

	// Create schema export handler for TypeScript type generation
	schemaExportHandler := NewSchemaExportHandler(schemaCache, db.Inspector())

	// Create AI storage first (needed for provider lookup)
	aiStorage := ai.NewStorage(db)
	aiStorage.SetConfig(&cfg.AI)

	// Create vector manager with hot-reload capability
	vectorManager := NewVectorManager(&cfg.AI, aiStorage, db.Inspector(), db)

	// Create vector search handler (for pgvector support) - create early for embedding service sharing
	// Embedding can be enabled explicitly (EmbeddingEnabled=true) or via fallback from AI provider
	var vectorHandler *VectorHandler
	vectorHandler, err = NewVectorHandler(vectorManager, db.Inspector(), db, cfg)
	//nolint:gocritic // Initialization state checks, not switch-compatible
	if err != nil {
		log.Warn().Err(err).Msg("Failed to initialize vector handler")
	} else if vectorHandler.IsEmbeddingConfigured() {
		// Embedding is available (either explicitly configured or via AI provider fallback)
		provider := cfg.AI.EmbeddingProvider
		if provider == "" {
			provider = cfg.AI.ProviderType
		}
		model := ""
		if vectorHandler.GetEmbeddingService() != nil {
			model = vectorHandler.GetEmbeddingService().DefaultModel()
		}
		log.Info().
			Str("provider", provider).
			Str("model", model).
			Bool("explicit_config", cfg.AI.EmbeddingEnabled).
			Msg("Vector handler initialized with embedding support")
	} else {
		log.Info().Msg("Vector handler initialized (embedding not available)")
	}

	// Create AI components (only if AI is enabled)
	var aiHandler *ai.Handler
	var aiChatHandler *ai.ChatHandler
	var aiConversations *ai.ConversationManager
	var aiMetrics *observability.Metrics
	if cfg.AI.Enabled {
		// Create AI metrics
		aiMetrics = observability.NewMetrics()

		// AI storage already created above for vectorManager
		// Create AI loader
		aiLoader := ai.NewLoader(cfg.AI.ChatbotsDir)

		// Create conversation manager
		aiConversations = ai.NewConversationManager(db, cfg.AI.ConversationCacheTTL, cfg.AI.MaxConversationTurns)

		// Create AI handler for admin endpoints (pass vectorManager for hot-reload)
		aiHandler = ai.NewHandler(aiStorage, aiLoader, &cfg.AI, vectorManager)

		// Get embedding service from vector handler (if available) for RAG support
		var embeddingService *ai.EmbeddingService
		if vectorHandler != nil {
			embeddingService = vectorHandler.GetEmbeddingService()
		}

		// Create AI chat handler for WebSocket with RAG support
		aiChatHandler = ai.NewChatHandler(db, aiStorage, aiConversations, aiMetrics, &cfg.AI, embeddingService, loggingService)

		// Create settings resolver for chatbot template variable resolution
		settingsResolver := ai.NewSettingsResolver(secretsService, 5*time.Minute)
		aiChatHandler.SetSettingsResolver(settingsResolver)

		log.Info().
			Str("chatbots_dir", cfg.AI.ChatbotsDir).
			Bool("auto_load", cfg.AI.AutoLoadOnBoot).
			Str("provider_type", cfg.AI.ProviderType).
			Str("provider_name", cfg.AI.ProviderName).
			Str("provider_model", cfg.AI.ProviderModel).
			Bool("rag_enabled", embeddingService != nil).
			Msg("AI components initialized")
	}

	// Create knowledge base handler for RAG management
	var knowledgeBaseHandler *ai.KnowledgeBaseHandler
	var kbStorage *ai.KnowledgeBaseStorage
	var docProcessor *ai.DocumentProcessor
	var tableExportSyncService *ai.TableExportSyncService
	var ocrService *ai.OCRService
	var quotaHandler *QuotaHandler
	if cfg.AI.Enabled {
		// Initialize OCR service for image-based PDF extraction
		if cfg.AI.OCREnabled {
			var err error
			ocrService, err = ai.NewOCRService(ai.OCRServiceConfig{
				Enabled:          cfg.AI.OCREnabled,
				ProviderType:     ai.OCRProviderType(cfg.AI.OCRProvider),
				DefaultLanguages: cfg.AI.OCRLanguages,
			})
			if err != nil {
				log.Warn().Err(err).Msg("Failed to initialize OCR service, OCR will be disabled")
			} else if ocrService.IsEnabled() {
				log.Info().
					Str("provider", cfg.AI.OCRProvider).
					Strs("languages", cfg.AI.OCRLanguages).
					Msg("OCR service initialized")
			}
		}

		kbStorage = ai.NewKnowledgeBaseStorage(db)

		// Initialize knowledge graph for entity and relationship storage
		knowledgeGraph := ai.NewKnowledgeGraph(kbStorage)
		log.Info().Msg("Knowledge graph initialized")

		// Initialize entity extractor for extracting entities from documents
		entityExtractor := ai.NewRuleBasedExtractor()
		log.Info().Msg("Entity extractor initialized")

		if vectorHandler != nil && vectorHandler.GetEmbeddingService() != nil {
			docProcessor = ai.NewDocumentProcessor(kbStorage, vectorHandler.GetEmbeddingService(), entityExtractor, knowledgeGraph)
		}

		// Use OCR-enabled handler if OCR service is available
		if ocrService != nil && ocrService.IsEnabled() {
			knowledgeBaseHandler = ai.NewKnowledgeBaseHandlerWithOCR(kbStorage, docProcessor, ocrService)
		} else {
			knowledgeBaseHandler = ai.NewKnowledgeBaseHandler(kbStorage, docProcessor)
		}
		knowledgeBaseHandler.SetStorageService(storageService)

		// Initialize table exporter for database schema export
		tableExporter := ai.NewTableExporter(db, docProcessor, knowledgeGraph, kbStorage)
		knowledgeBaseHandler.SetTableExporter(tableExporter)
		knowledgeBaseHandler.SetKnowledgeGraph(knowledgeGraph)
		log.Info().Msg("Table exporter initialized")

		// Initialize table export sync service
		tableExportSyncService = ai.NewTableExportSyncService(db, tableExporter, kbStorage)
		knowledgeBaseHandler.SetSyncService(tableExportSyncService)
		log.Info().Msg("Table export sync service initialized")

		// Set knowledge base storage on AI handler for syncing KB links during chatbot sync
		aiHandler.SetKnowledgeBaseStorage(kbStorage)
		log.Info().Msg("AI handler configured with knowledge base storage")

		log.Info().
			Bool("processing_enabled", docProcessor != nil).
			Bool("ocr_enabled", ocrService != nil && ocrService.IsEnabled()).
			Bool("entity_extraction_enabled", true).
			Bool("table_export_enabled", true).
			Bool("sync_enabled", true).
			Msg("Knowledge base handler initialized")

		// Initialize quota service and handler
		quotaService := ai.NewQuotaService(kbStorage)
		quotaHandler = NewQuotaHandler(quotaService, userMgmtService)
		log.Info().Msg("Quota service and handler initialized")
	}

	// Create internal AI handler for custom MCP tools, edge functions, and jobs
	// This allows runtime code to access AI capabilities via utils.ai.chat() and utils.ai.embed()
	var internalAIHandler *InternalAIHandler
	if cfg.AI.Enabled {
		var embeddingSvc *ai.EmbeddingService
		if vectorHandler != nil {
			embeddingSvc = vectorHandler.GetEmbeddingService()
		}
		internalAIHandler = NewInternalAIHandler(aiStorage, embeddingSvc, cfg.AI.ProviderName)
		log.Info().
			Str("default_provider", cfg.AI.ProviderName).
			Bool("embedding_enabled", embeddingSvc != nil).
			Msg("Internal AI handler initialized for MCP tools/functions/jobs")
	}

	// Create internal schema handler for declarative schema management
	internalSchemaHandler := NewInternalSchemaHandler()
	internalSchemaHandler.Initialize(cfg, db)
	log.Info().Msg("Internal schema handler initialized")

	// Create RPC components (only if RPC is enabled)
	var rpcHandler *rpc.Handler
	var rpcScheduler *rpc.Scheduler
	if cfg.RPC.Enabled {
		rpcStorage := rpc.NewStorage(db)
		rpcLoader := rpc.NewLoader(cfg.RPC.ProceduresDir)
		rpcMetrics := observability.NewMetrics()
		rpcHandler = rpc.NewHandler(db, rpcStorage, rpcLoader, rpcMetrics, &cfg.RPC, authService, loggingService, cfg)

		// Create RPC scheduler and wire it to handler
		rpcScheduler = rpc.NewScheduler(rpcStorage, rpcHandler.GetExecutor())
		rpcHandler.SetScheduler(rpcScheduler)

		log.Info().
			Str("procedures_dir", cfg.RPC.ProceduresDir).
			Bool("auto_load", cfg.RPC.AutoLoadOnBoot).
			Msg("RPC components initialized")
	}

	// Create realtime components with connection limits from config
	realtimeManager := realtime.NewManagerWithConfig(context.Background(), realtime.ManagerConfig{
		MaxConnections:         cfg.Realtime.MaxConnections,
		MaxConnectionsPerUser:  cfg.Realtime.MaxConnectionsPerUser,
		MaxConnectionsPerIP:    cfg.Realtime.MaxConnectionsPerIP,
		ClientMessageQueueSize: cfg.Realtime.ClientMessageQueueSize,
	})
	realtimeManager.SetBaseConfig(cfg)

	// Set up cross-instance broadcasting via pub/sub (if configured)
	if ps != nil {
		realtimeManager.SetPubSub(ps)
	}

	realtimeAuthAdapter := realtime.NewAuthServiceAdapter(authService)
	realtimeSubManager := realtime.NewSubscriptionManagerWithConfig(
		realtime.NewPgxSubscriptionDB(db.Pool()),
		realtime.RLSCacheConfig{
			MaxSize: cfg.Realtime.RLSCacheSize,
			TTL:     cfg.Realtime.RLSCacheTTL,
		},
	)
	realtimeHandler := realtime.NewRealtimeHandler(realtimeManager, realtimeAuthAdapter, realtimeSubManager)
	realtimeListener := realtime.NewListenerPool(
		db.Pool(),
		realtimeHandler,
		realtimeSubManager,
		ps,
		realtime.ListenerPoolConfig{
			PoolSize:    cfg.Realtime.ListenerPoolSize,
			WorkerCount: cfg.Realtime.NotificationWorkers,
			QueueSize:   cfg.Realtime.NotificationQueueSize,
		},
	)

	// Create monitoring handler
	monitoringHandler := NewMonitoringHandler(db.Pool(), realtimeHandler, storageService.Provider)

	// Set logging service if available for log queries
	if loggingService != nil {
		monitoringHandler.SetLoggingService(loggingService)
	}

	// Create server instance with handler groups
	server := &Server{
		app:        app,
		config:     cfg,
		db:         db,
		tracer:     tracer,
		rest:       NewRESTHandler(db, NewQueryParser(cfg), schemaCache, cfg),
		sqlHandler: sqlHandler,

		// Auth handlers group
		Auth: &AuthHandlers{
			Handler:          authHandler,
			AdminHandler:     adminAuthHandler,
			DashboardHandler: dashboardAuthHandler,
			ClientKeyHandler: clientKeyHandler,
			ClientKeyService: clientKeyService,
			OAuthProvider:    oauthProviderHandler,
			OAuth:            oauthHandler,
			SAMLProvider:     samlProviderHandler,
			SAMLService:      samlService,
			AdminSession:     adminSessionHandler,
			UserManagement:   userMgmtHandler,
			Invitation:       invitationHandler,
		},

		// Storage handlers group
		Storage: &StorageHandlers{
			Handler: storageHandler,
		},

		// AI handlers group
		AI: &AIHandlers{
			Handler:         aiHandler,
			Chat:            aiChatHandler,
			Conversations:   aiConversations,
			Metrics:         aiMetrics,
			KnowledgeBase:   knowledgeBaseHandler,
			KBStorage:       kbStorage,
			DocProcessor:    docProcessor,
			TableExportSync: tableExportSyncService,
			VectorManager:   vectorManager,
			VectorHandler:   vectorHandler,
			Internal:        internalAIHandler,
		},

		// Functions handlers group
		Functions: &FunctionsHandlers{
			Handler:   functionsHandler,
			Scheduler: functionsScheduler,
		},

		// Jobs handlers group
		Jobs: &JobsHandlers{
			Handler:   jobsHandler,
			Manager:   jobsManager,
			Scheduler: jobsScheduler,
		},

		// Realtime handlers group
		Realtime: &RealtimeHandlers{
			Manager:  realtimeManager,
			Handler:  realtimeHandler,
			Listener: realtimeListener,
			Admin:    realtimeAdminHandler,
		},

		// MCP handlers group
		MCP: &MCPHandlers{
			Handler:       mcp.NewHandler(&cfg.MCP, db),
			OAuth:         NewMCPOAuthHandler(db.Pool(), &cfg.MCP, authService, cfg.BaseURL, cfg.GetPublicBaseURL()),
			CustomManager: nil, // Initialized later in setupMCPServer
			CustomHandler: nil, // Initialized later in setupMCPServer
		},

		// Tenancy handlers group
		Tenancy: &TenancyHandlers{
			ServiceKey: serviceKeyHandler,
			Tenant:     tenantHandler,
			Manager:    tenantManager,
			Storage:    tenantStorage,
		},

		// Branching handlers group (initialized later if enabled)
		Branching: &BranchingHandlers{},

		// Settings handlers group
		Settings: &SettingsHandlers{
			System:   systemSettingsHandler,
			Custom:   customSettingsHandler,
			User:     userSettingsHandler,
			App:      appSettingsHandler,
			Handler:  settingsHandler,
			Service:  secretsService,
			Instance: instanceSettingsHandler,
			Tenant:   tenantSettingsHandler,
			Unified:  unifiedSettingsService,
		},

		// Webhook handlers group
		Webhook: &WebhookHandlers{
			Handler: webhookHandler,
			Trigger: webhookTriggerService,
		},

		// Logging handlers group
		Logging: &LoggingHandlers{
			Service:   loggingService,
			Handler:   loggingHandler,
			Retention: retentionService,
		},

		// Schema handlers group
		Schema: &SchemaHandlers{
			DDL:            ddlHandler,
			Migrations:     migrationsHandler,
			Cache:          schemaCache,
			Export:         schemaExportHandler,
			InternalSchema: internalSchemaHandler,
		},

		// RPC handlers group
		RPC: &RPCHandlers{
			Handler:   rpcHandler,
			Scheduler: rpcScheduler,
		},

		// GraphQL handlers group (initialized later if enabled)
		GraphQL: &GraphQLHandlers{},

		// Extensions handlers group
		Extensions: &ExtensionsHandlers{
			Handler: extensions.NewHandler(extensions.NewService(db)),
		},

		// Secrets handlers group
		Secrets: &SecretsHandlers{
			Handler: secretsHandler,
			Storage: secretsStorage,
		},

		// Scaling handlers group (initialized later if enabled)
		Scaling: &ScalingHandlers{},

		// Metrics components
		Metrics: &MetricsComponents{
			Metrics:   observability.NewMetrics(),
			StartTime: time.Now(),
			StopChan:  nil, // Initialized later
		},

		// Email handlers group
		Email: &EmailHandlers{
			Template: emailTemplateHandler,
			Settings: emailSettingsHandler,
		},

		// Captcha handlers group
		Captcha: &CaptchaHandlers{
			Settings: captchaSettingsHandler,
		},

		// Monitoring handlers group
		Monitoring: &MonitoringHandlers{
			Handler: monitoringHandler,
		},

		// Quota handlers group
		Quota: &QuotaHandlers{
			Handler: quotaHandler,
		},

		// Middleware components
		Middleware: &MiddlewareComponents{
			Tenant: middleware.TenantMiddleware(middleware.TenantConfig{
				DB: db,
			}),
			TenantDB: func() fiber.Handler {
				if tenantManager != nil && tenantManager.GetRouter() != nil {
					return middleware.TenantDBMiddleware(middleware.TenantDBConfig{
						Router:  tenantManager.GetRouter(),
						Storage: tenantStorage,
					})
				}
				return nil
			}(),
		},

		// Server-owned dependencies
		rateLimiter:             rateLimitStore,
		pubSub:                  ps,
		sharedMiddlewareStorage: sharedMiddlewareStorage,

		// Tenant configuration loader for multi-tenant config overrides
		tenantConfigLoader: nil, // Initialized later after migrations
	}

	// Initialize MCP Server if enabled
	if cfg.MCP.Enabled {
		server.setupMCPServer(schemaCache, storageService, functionsHandler, rpcHandler, vectorHandler)
		log.Info().
			Str("base_path", cfg.MCP.BasePath).
			Dur("session_timeout", cfg.MCP.SessionTimeout).
			Msg("MCP Server enabled")
	}

	// Initialize Database Branching if enabled
	if cfg.Branching.Enabled {
		branchStorage := branching.NewStorage(db.Pool(), cfg.EncryptionKey)
		dbURL := cfg.Database.RuntimeConnectionString()
		branchManager, err := branching.NewManager(branchStorage, cfg.Branching, db.Pool(), dbURL)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize branch manager")
		}
		branchRouter := branching.NewRouter(branchStorage, cfg.Branching, db.Pool(), dbURL)

		server.Branching.Manager = branchManager
		server.Branching.Router = branchRouter
		server.Branching.Handler = NewBranchHandler(branchManager, branchRouter, cfg.Branching)
		server.Branching.GitHub = NewGitHubWebhookHandler(branchManager, branchRouter, cfg.Branching)

		// Initialize cleanup scheduler if auto_delete_after is set
		if cfg.Branching.AutoDeleteAfter > 0 {
			// Use auto_delete_after as the interval, or default to hourly if it's very short
			cleanupInterval := cfg.Branching.AutoDeleteAfter
			if cleanupInterval < time.Hour {
				cleanupInterval = time.Hour
			}
			server.Branching.Scheduler = branching.NewCleanupScheduler(branchManager, branchRouter, cleanupInterval)
			log.Info().
				Dur("interval", cleanupInterval).
				Dur("auto_delete_after", cfg.Branching.AutoDeleteAfter).
				Msg("Branch cleanup scheduler initialized")
		}

		log.Info().
			Int("max_branches", cfg.Branching.MaxTotalBranches).
			Str("default_clone_mode", cfg.Branching.DefaultDataCloneMode).
			Msg("Database Branching enabled")
	}

	// Store tenant components in server (initialized earlier)
	server.Tenancy.Manager = tenantManager
	server.Tenancy.Storage = tenantStorage

	// Create GraphQL handler (if enabled)
	if cfg.GraphQL.Enabled {
		server.GraphQL.Handler = NewGraphQLHandler(db, schemaCache, &cfg.GraphQL, cfg)
		log.Info().
			Int("max_depth", cfg.GraphQL.MaxDepth).
			Int("max_complexity", cfg.GraphQL.MaxComplexity).
			Bool("introspection", cfg.GraphQL.Introspection).
			Msg("GraphQL API enabled")
	}

	// Start realtime listener (unless disabled or in worker-only mode)
	if !cfg.Scaling.DisableRealtime && !cfg.Scaling.WorkerOnly {
		if err := realtimeListener.Start(); err != nil {
			log.Error().Err(err).Msg("Failed to start realtime listener")
		}
	} else {
		log.Info().
			Bool("disable_realtime", cfg.Scaling.DisableRealtime).
			Bool("worker_only", cfg.Scaling.WorkerOnly).
			Msg("Realtime listener disabled by scaling configuration")
	}

	// Start edge functions scheduler (respects scaling configuration)
	if !cfg.Scaling.DisableScheduler && !cfg.Scaling.WorkerOnly {
		if cfg.Scaling.EnableSchedulerLeaderElection {
			// Use leader election - only the leader will run the scheduler
			server.Scaling.FunctionsLeader = scaling.NewLeaderElector(
				db.Pool(),
				scaling.FunctionsSchedulerLockID,
				"functions-scheduler",
			)
			server.Scaling.FunctionsLeader.Start(
				func() {
					// Became leader - start the scheduler
					log.Info().Msg("This instance is now the functions scheduler leader")
					if err := functionsScheduler.Start(); err != nil {
						log.Error().Err(err).Msg("Failed to start edge functions scheduler")
					}
				},
				func() {
					// Lost leadership - stop the scheduler
					log.Warn().Msg("Lost functions scheduler leadership - stopping scheduler")
					functionsScheduler.Stop()
				},
			)
		} else {
			// No leader election - start scheduler directly
			if err := functionsScheduler.Start(); err != nil {
				log.Error().Err(err).Msg("Failed to start edge functions scheduler")
			}
		}
	} else {
		log.Info().
			Bool("disable_scheduler", cfg.Scaling.DisableScheduler).
			Bool("worker_only", cfg.Scaling.WorkerOnly).
			Msg("Edge functions scheduler disabled by scaling configuration")
	}

	// Start jobs manager and scheduler
	if cfg.Jobs.Enabled && jobsManager != nil {
		// Job workers can run on any instance (including worker-only mode)
		// The scheduler should respect the scaling configuration
		workerCount := cfg.Jobs.EmbeddedWorkerCount
		if workerCount <= 0 {
			workerCount = 4 // Default to 4 workers if not configured
		}
		if err := jobsManager.Start(context.Background(), workerCount); err != nil {
			log.Error().Err(err).Msg("Failed to start jobs manager")
		} else {
			log.Info().Int("workers", workerCount).Msg("Jobs manager started successfully")
		}

		// Start jobs scheduler for cron-based execution (respects scaling configuration)
		if jobsScheduler != nil {
			if !cfg.Scaling.DisableScheduler && !cfg.Scaling.WorkerOnly {
				if cfg.Scaling.EnableSchedulerLeaderElection {
					// Use leader election - only the leader will run the scheduler
					server.Scaling.JobsLeader = scaling.NewLeaderElector(
						db.Pool(),
						scaling.JobsSchedulerLockID,
						"jobs-scheduler",
					)
					server.Scaling.JobsLeader.Start(
						func() {
							// Became leader - start the scheduler
							log.Info().Msg("This instance is now the jobs scheduler leader")
							if err := jobsScheduler.Start(); err != nil {
								log.Error().Err(err).Msg("Failed to start jobs scheduler")
							}
						},
						func() {
							// Lost leadership - stop the scheduler
							log.Warn().Msg("Lost jobs scheduler leadership - stopping scheduler")
							jobsScheduler.Stop()
						},
					)
				} else {
					// No leader election - start scheduler directly
					if err := jobsScheduler.Start(); err != nil {
						log.Error().Err(err).Msg("Failed to start jobs scheduler")
					}
				}
			} else {
				log.Info().
					Bool("disable_scheduler", cfg.Scaling.DisableScheduler).
					Bool("worker_only", cfg.Scaling.WorkerOnly).
					Msg("Jobs scheduler disabled by scaling configuration (workers still active)")
			}
		}
	}

	// Start RPC scheduler for cron-based procedure execution (respects scaling configuration)
	if cfg.RPC.Enabled && rpcScheduler != nil {
		if !cfg.Scaling.DisableScheduler && !cfg.Scaling.WorkerOnly {
			if cfg.Scaling.EnableSchedulerLeaderElection {
				// Use leader election - only the leader will run the scheduler
				server.Scaling.RPCLeader = scaling.NewLeaderElector(
					db.Pool(),
					scaling.RPCSchedulerLockID,
					"rpc-scheduler",
				)
				server.Scaling.RPCLeader.Start(
					func() {
						// Became leader - start the scheduler
						log.Info().Msg("This instance is now the RPC scheduler leader")
						if err := rpcScheduler.Start(); err != nil {
							log.Error().Err(err).Msg("Failed to start RPC scheduler")
						}
					},
					func() {
						// Lost leadership - stop the scheduler
						log.Warn().Msg("Lost RPC scheduler leadership - stopping scheduler")
						rpcScheduler.Stop()
					},
				)
			} else {
				// No leader election - start scheduler directly
				if err := rpcScheduler.Start(); err != nil {
					log.Error().Err(err).Msg("Failed to start RPC scheduler")
				}
			}
		} else {
			log.Info().
				Bool("disable_scheduler", cfg.Scaling.DisableScheduler).
				Bool("worker_only", cfg.Scaling.WorkerOnly).
				Msg("RPC scheduler disabled by scaling configuration")
		}
	}

	// Start webhook trigger service
	if err := webhookTriggerService.Start(context.Background()); err != nil {
		log.Error().Err(err).Msg("Failed to start webhook trigger service")
	}

	// Start retention cleanup service (for central logging)
	if retentionService != nil {
		retentionService.Start()
		log.Info().
			Dur("interval", cfg.Logging.RetentionCheckInterval).
			Msg("Log retention cleanup service started")
	}

	// Start branch cleanup scheduler
	if server.Branching.Scheduler != nil {
		server.Branching.Scheduler.Start()
	}

	// Start Prometheus metrics server if enabled
	if cfg.Metrics.Enabled {
		server.Metrics.Server = observability.NewMetricsServer(cfg.Metrics.Port, cfg.Metrics.Path)
		if err := server.Metrics.Server.Start(); err != nil {
			log.Error().Err(err).Msg("Failed to start metrics server")
		}

		// Wire up database metrics
		db.SetMetrics(server.Metrics.Metrics)

		// Wire up storage metrics
		if storageService != nil {
			storageService.SetMetrics(server.Metrics.Metrics)
		}

		// Wire up auth metrics
		authService.SetMetrics(server.Metrics.Metrics)

		// Wire up realtime metrics
		if realtimeManager != nil {
			realtimeManager.SetMetrics(server.Metrics.Metrics)
		}

		// Wire up rate limiter metrics
		middleware.SetRateLimiterMetrics(server.Metrics.Metrics)

		// Start uptime tracking goroutine
		server.Metrics.StopChan = make(chan struct{})
		go func() {
			ticker := time.NewTicker(15 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					server.Metrics.Metrics.UpdateUptime(server.Metrics.StartTime)
				case <-server.Metrics.StopChan:
					return
				}
			}
		}()
	}

	// Auto-load AI chatbots if enabled
	if cfg.AI.Enabled && cfg.AI.AutoLoadOnBoot && aiHandler != nil {
		if err := aiHandler.AutoLoadChatbots(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to auto-load AI chatbots")
		} else {
			log.Info().Msg("AI chatbots auto-loaded successfully")
		}
	}

	// Auto-load custom MCP tools if enabled
	if cfg.MCP.Enabled && cfg.MCP.AutoLoadOnBoot && server.MCP.CustomManager != nil {
		if err := server.MCP.CustomManager.AutoLoadFromDir(context.Background(), cfg.MCP.ToolsDir); err != nil {
			log.Error().Err(err).Msg("Failed to auto-load custom MCP tools")
		} else {
			log.Info().Msg("Custom MCP tools auto-loaded successfully")
		}
	}

	// Setup middlewares
	log.Debug().Msg("Setting up middlewares")
	server.setupMiddlewares()

	// Setup routes
	log.Debug().Msg("Setting up routes")
	server.setupRoutes()

	// Set globals for backward compatibility with handlers using GetGlobalStore()
	// The server owns these dependencies and will close them on shutdown
	if server.rateLimiter != nil {
		ratelimit.SetGlobalStore(server.rateLimiter)
	}
	if server.pubSub != nil {
		pubsub.SetGlobalPubSub(server.pubSub)
	}

	log.Debug().Msg("Server initialization complete")
	return server
}

// NewServerWithTx creates a test-mode server with transaction isolation.
// This is specifically for HTTP API tests that need to use a transaction.
//
// Note: This function creates a minimal server with only the essential components
// for HTTP API testing. It does NOT initialize all services (webhooks, realtime, jobs, etc.).
func NewServerWithTx(cfg *config.Config, db *database.Connection, tx pgx.Tx, version string) *Server {
	// Use the existing NewServer to create a full server
	server := NewServer(cfg, db, version)

	// Set the test transaction
	server.testTx = tx

	return server
}

// DB returns the database querier to use.
// In test mode with a transaction, it returns the transaction (note: can't use tx as pool).
// Otherwise, it returns the normal database connection pool.
func (s *Server) DB() *pgxpool.Pool {
	if s.testTx != nil {
		// In test mode, we can't return the transaction as a pool
		// Tests should use the transaction directly via BeginTx()
		return s.db.Pool()
	}
	return s.db.Pool()
}

// createMCPAuthMiddleware creates authentication middleware for MCP that supports
// JWT, client key, service key, AND MCP OAuth tokens
func (s *Server) createMCPAuthMiddleware() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Check for MCP OAuth token first (Bearer token starting with "mcp_at_")
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer mcp_at_") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate MCP OAuth token
			if s.MCP.OAuth != nil {
				clientID, userID, scopes, err := s.MCP.OAuth.ValidateAccessToken(c, token)
				if err == nil {
					// Valid MCP OAuth token
					c.Locals("auth_type", "mcp_oauth")
					c.Locals("client_key_id", clientID)
					c.Locals("client_key_scopes", scopes)
					if userID != nil {
						c.Locals("user_id", *userID)
					}
					return c.Next()
				}
			}
		}

		// Fall back to standard auth middleware
		return middleware.RequireAuthOrServiceKey(
			s.Auth.Handler.authService,
			s.Auth.ClientKeyService,
			s.DB(),
			nil,
			s.Auth.DashboardHandler.jwtManager,
		)(c)
	}
}

// setupMCPServer initializes the MCP server with tools and resources
func (s *Server) setupMCPServer(schemaCache *database.SchemaCache, storageService *storage.Service, functionsHandler *functions.Handler, rpcHandler *rpc.Handler, vectorHandler *VectorHandler) {
	mcpServer := s.MCP.Handler.Server()

	// Register MCP Tools
	toolRegistry := mcpServer.ToolRegistry()

	// Reasoning tool (always available, no scopes required)
	toolRegistry.Register(mcptools.NewThinkTool())

	// Database tools
	queryTableTool := mcptools.NewQueryTableTool(s.db, schemaCache)
	if vectorHandler != nil && vectorHandler.GetEmbeddingService() != nil {
		queryTableTool.SetEmbeddingGenerator(vectorHandler.GetEmbeddingService())
		log.Debug().Msg("MCP: QueryTableTool configured with embedding generator for vector search")
	}
	toolRegistry.Register(queryTableTool)
	toolRegistry.Register(mcptools.NewInsertRecordTool(s.db, schemaCache))
	toolRegistry.Register(mcptools.NewUpdateRecordTool(s.db, schemaCache))
	toolRegistry.Register(mcptools.NewDeleteRecordTool(s.db, schemaCache))
	toolRegistry.Register(mcptools.NewExecuteSQLTool(s.db))

	// Storage tools
	if storageService != nil {
		toolRegistry.Register(mcptools.NewListObjectsTool(storageService))
		toolRegistry.Register(mcptools.NewUploadObjectTool(storageService))
		toolRegistry.Register(mcptools.NewDownloadObjectTool(storageService))
		toolRegistry.Register(mcptools.NewDeleteObjectTool(storageService))
	}

	// Functions invocation tools
	if functionsHandler != nil && s.config.Functions.Enabled {
		toolRegistry.Register(mcptools.NewInvokeFunctionTool(
			s.db,
			functionsHandler.GetRuntime(),
			functionsHandler.GetPublicURL(),
			functionsHandler.GetFunctionsDir(),
		))
	}

	// RPC invocation tools
	if rpcHandler != nil && s.config.RPC.Enabled {
		rpcStorage := rpc.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewInvokeRPCTool(
			rpcHandler.GetExecutor(),
			rpcStorage,
		))
	}

	// Jobs tools
	if s.Jobs.Manager != nil && s.config.Jobs.Enabled {
		jobsStorage := jobs.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewSubmitJobTool(jobsStorage))
		toolRegistry.Register(mcptools.NewGetJobStatusTool(jobsStorage))
	}

	// Vector search tools
	if s.AI.Chat != nil {
		if ragService := s.AI.Chat.GetRAGService(); ragService != nil {
			toolRegistry.Register(mcptools.NewSearchVectorsTool(ragService))
			log.Debug().Msg("MCP: Registered search_vectors tool")
		} else {
			log.Debug().Msg("MCP: Vector search tool not registered - RAG service not available")
		}
	}

	// Knowledge graph tools
	if s.AI.KBStorage != nil {
		knowledgeGraph := ai.NewKnowledgeGraph(s.AI.KBStorage)
		toolRegistry.Register(mcptools.NewQueryKnowledgeGraphTool(knowledgeGraph))
		toolRegistry.Register(mcptools.NewFindRelatedEntitiesTool(knowledgeGraph))
		toolRegistry.Register(mcptools.NewBrowseKnowledgeGraphTool(knowledgeGraph))
		log.Debug().Msg("MCP: Registered knowledge graph tools")
	}

	// DDL tools (schema/table management)
	toolRegistry.Register(mcptools.NewListSchemasTool(s.db))
	toolRegistry.Register(mcptools.NewCreateSchemaTool(s.db))
	toolRegistry.Register(mcptools.NewCreateTableTool(s.db))
	toolRegistry.Register(mcptools.NewDropTableTool(s.db))
	toolRegistry.Register(mcptools.NewAddColumnTool(s.db))
	toolRegistry.Register(mcptools.NewDropColumnTool(s.db))
	toolRegistry.Register(mcptools.NewRenameTableTool(s.db))

	// HTTP request tool (for chatbots with external API access)
	toolRegistry.Register(mcptools.NewHttpRequestTool())

	// Sync tools (deploy functions, jobs, RPC, migrations, chatbots via MCP)
	if s.config.Functions.Enabled {
		functionsStorage := functions.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewSyncFunctionTool(functionsStorage))
	}

	if s.config.Jobs.Enabled {
		jobsStorage := jobs.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewSyncJobTool(jobsStorage))
	}

	if s.config.RPC.Enabled {
		rpcStorage := rpc.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewSyncRPCTool(rpcStorage))
	}

	// Migrations sync tool
	migrationsStorage := migrations.NewStorage(s.db)
	migrationsExecutor := migrations.NewExecutor(s.db)
	toolRegistry.Register(mcptools.NewSyncMigrationTool(migrationsStorage, migrationsExecutor))

	// AI/Chatbot sync tool
	if s.config.AI.Enabled {
		aiStorage := ai.NewStorage(s.db)
		toolRegistry.Register(mcptools.NewSyncChatbotTool(aiStorage))
	}

	// Database branching tools
	if s.Branching.Manager != nil && s.config.Branching.Enabled {
		branchStorage := branching.NewStorage(s.db.Pool(), s.config.EncryptionKey)
		toolRegistry.Register(mcptools.NewListBranchesTool(branchStorage))
		toolRegistry.Register(mcptools.NewGetBranchTool(branchStorage))
		toolRegistry.Register(mcptools.NewCreateBranchTool(s.Branching.Manager))
		toolRegistry.Register(mcptools.NewDeleteBranchTool(s.Branching.Manager, branchStorage))
		toolRegistry.Register(mcptools.NewResetBranchTool(s.Branching.Manager, branchStorage))
		toolRegistry.Register(mcptools.NewGrantBranchAccessTool(branchStorage))
		toolRegistry.Register(mcptools.NewRevokeBranchAccessTool(branchStorage))
		toolRegistry.Register(mcptools.NewGetActiveBranchTool(s.Branching.Router))
		toolRegistry.Register(mcptools.NewSetActiveBranchTool(s.Branching.Router, branchStorage))
	}

	// Register MCP Resources
	resourceRegistry := mcpServer.ResourceRegistry()

	// Schema resources
	resourceRegistry.Register(mcpresources.NewSchemaResource(schemaCache))
	resourceRegistry.Register(mcpresources.NewTableResource(schemaCache))

	// Functions resources
	if s.config.Functions.Enabled {
		resourceRegistry.Register(mcpresources.NewFunctionsResource(functions.NewStorage(s.db)))
	}

	// RPC resources
	if s.config.RPC.Enabled {
		resourceRegistry.Register(mcpresources.NewRPCResource(rpc.NewStorage(s.db)))
	}

	// Storage resources
	resourceRegistry.Register(mcpresources.NewBucketsResource(s.db))

	// Wire MCP registries to AI chat handler for MCP-enabled chatbots
	if s.AI.Chat != nil {
		s.AI.Chat.SetMCPToolRegistry(toolRegistry)
		s.AI.Chat.SetMCPResources(resourceRegistry)
		log.Debug().Msg("MCP registries wired to AI chat handler")
	}

	// Initialize custom MCP tools and resources
	customStorage := custom.NewStorage(s.db.Pool())
	// Use BaseURL for internal communication, falling back to localhost with server address
	mcpInternalURL := s.config.BaseURL
	if mcpInternalURL == "" {
		mcpInternalURL = "http://localhost" + s.config.Server.Address
	}
	customExecutor := custom.NewExecutor(s.config.Auth.JWTSecret, mcpInternalURL, nil)
	s.MCP.CustomManager = custom.NewManager(customStorage, customExecutor, toolRegistry, resourceRegistry)
	s.MCP.CustomHandler = NewCustomMCPHandler(customStorage, s.MCP.CustomManager, &s.config.MCP)

	// Load custom tools and resources from database
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.MCP.CustomManager.LoadAndRegisterAll(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to load some custom MCP tools/resources")
	}

	log.Debug().
		Int("tools", len(toolRegistry.ListTools(&mcp.AuthContext{IsServiceRole: true}))).
		Int("resources", len(resourceRegistry.ListResources(&mcp.AuthContext{IsServiceRole: true}))).
		Msg("MCP Server initialized with tools and resources")
}

// setupMiddlewares sets up global middlewares
func (s *Server) setupMiddlewares() {
	// Request ID middleware - must be first for tracing
	log.Debug().Msg("Adding requestid middleware")
	s.app.Use(requestid.New())

	// OpenTelemetry tracing middleware - adds distributed tracing to all requests
	if s.config.Tracing.Enabled && s.tracer != nil && s.tracer.IsEnabled() {
		log.Debug().Msg("Adding OpenTelemetry tracing middleware")
		s.app.Use(middleware.TracingMiddleware(middleware.TracingConfig{
			Enabled:            true,
			ServiceName:        s.config.Tracing.ServiceName,
			SkipPaths:          []string{"/health", "/ready", "/metrics"},
			RecordRequestBody:  false, // Don't record bodies for security
			RecordResponseBody: false,
		}))
	}

	// Prometheus metrics middleware - collects HTTP metrics
	if s.config.Metrics.Enabled && s.Metrics.Metrics != nil {
		log.Debug().Msg("Adding Prometheus metrics middleware")
		s.app.Use(s.Metrics.Metrics.MetricsMiddleware())
	}

	// Security headers middleware - protect against common attacks
	// Apply different CSP for admin UI (needs Google Fonts) vs API routes
	log.Debug().Msg("Adding security headers middleware")
	s.app.Use(func(c fiber.Ctx) error {
		// Apply relaxed CSP for admin UI
		if strings.HasPrefix(c.Path(), "/admin") {
			return middleware.AdminUISecurityHeaders()(c)
		}
		// Apply strict CSP for all other routes
		return middleware.SecurityHeaders()(c)
	})

	// Structured logger middleware - logs HTTP requests through zerolog
	// This allows HTTP logs to be captured by the central logging system
	log.Debug().Msg("Adding structured logger middleware")
	s.app.Use(middleware.StructuredLogger(middleware.StructuredLoggerConfig{
		SkipPaths: []string{"/health", "/ready", "/metrics"},
		// In debug mode, log all requests; in production, skip successful requests to reduce noise
		SkipSuccessfulRequests: !s.config.Debug,
	}))

	// Recover middleware - catch panics
	log.Debug().Msg("Adding recover middleware")
	s.app.Use(recover.New(recover.Config{
		EnableStackTrace: s.config.Debug,
	}))

	// CORS middleware
	// Note: AllowCredentials cannot be used with AllowOrigins="*" per CORS spec
	// If AllowOrigins contains "*", we must disable credentials
	corsCredentials := s.config.CORS.AllowCredentials
	corsOrigins := s.config.CORS.AllowedOrigins

	// Check if origins contains wildcard
	hasWildcard := false
	for _, origin := range corsOrigins {
		if origin == "*" {
			hasWildcard = true
			break
		}
	}

	if hasWildcard && corsCredentials {
		log.Warn().Msg("CORS: AllowCredentials disabled because AllowOrigins contains '*' (not allowed per CORS spec)")
		corsCredentials = false
	}
	// Automatically add the public base URL to CORS origins if it's not already included
	// This ensures the dashboard can make API calls when deployed on a public URL
	if !hasWildcard && s.config.PublicBaseURL != "" {
		found := false
		for _, origin := range corsOrigins {
			if origin == s.config.PublicBaseURL {
				found = true
				break
			}
		}
		if !found {
			corsOrigins = append(corsOrigins, s.config.PublicBaseURL)
			log.Debug().Str("public_url", s.config.PublicBaseURL).Msg("Added public base URL to CORS origins")
		}
	}
	log.Debug().
		Strs("origins", corsOrigins).
		Bool("credentials", corsCredentials).
		Msg("Adding CORS middleware")

	// Build CORS config
	corsConfig := cors.Config{
		AllowMethods:     s.config.CORS.AllowedMethods,
		AllowHeaders:     s.config.CORS.AllowedHeaders,
		ExposeHeaders:    s.config.CORS.ExposedHeaders,
		AllowCredentials: corsCredentials,
		MaxAge:           s.config.CORS.MaxAge,
	}

	// When AllowOrigins contains "*", use AllowOriginsFunc to dynamically allow all origins
	// This is required because Fiber's CORS middleware doesn't properly handle "*"
	// with the AllowOrigins slice field in newer versions
	if hasWildcard {
		corsConfig.AllowOriginsFunc = func(origin string) bool {
			return true // Allow all origins
		}
	} else {
		corsConfig.AllowOrigins = corsOrigins
	}

	s.app.Use(cors.New(corsConfig))
	log.Debug().Msg("CORS middleware added")

	// Global IP allowlist - restrict access to entire API
	// Only log and apply if ranges are configured (empty = allow all)
	if len(s.config.Server.AllowedIPRanges) > 0 {
		log.Info().
			Int("ranges", len(s.config.Server.AllowedIPRanges)).
			Strs("ranges", s.config.Server.AllowedIPRanges).
			Msg("Adding global IP allowlist middleware")
		s.app.Use(middleware.RequireGlobalIPAllowlist(&s.config.Server))
	} else {
		log.Debug().Msg("Global IP allowlist disabled (no ranges configured)")
	}

	// Global rate limiting - 100 requests per minute per IP
	// Uses dynamic limiter that checks settings cache on each request
	// This allows toggling rate limiting via admin UI without server restart
	// Pass shared storage to prevent multiple GC goroutines
	s.app.Use(middleware.DynamicGlobalAPILimiter(s.Auth.Handler.authService.GetSettingsCache(), s.sharedMiddlewareStorage))

	// Per-endpoint body size limits and JSON depth protection
	if s.config.Server.BodyLimits.Enabled {
		bodyLimitConfig := middleware.BodyLimitsFromConfig(
			s.config.Server.BodyLimits.DefaultLimit,
			s.config.Server.BodyLimits.RESTLimit,
			s.config.Server.BodyLimits.AuthLimit,
			s.config.Server.BodyLimits.StorageLimit,
			s.config.Server.BodyLimits.BulkLimit,
			s.config.Server.BodyLimits.AdminLimit,
			s.config.Server.BodyLimits.MaxJSONDepth,
		)
		s.app.Use(middleware.BodyLimitMiddleware(bodyLimitConfig))
		log.Info().
			Int64("default", s.config.Server.BodyLimits.DefaultLimit).
			Int64("rest", s.config.Server.BodyLimits.RESTLimit).
			Int64("auth", s.config.Server.BodyLimits.AuthLimit).
			Int64("storage", s.config.Server.BodyLimits.StorageLimit).
			Int("max_json_depth", s.config.Server.BodyLimits.MaxJSONDepth).
			Msg("Per-endpoint body limits enabled")
	}

	// Idempotency key support for safe request retries
	// Stores responses in database to return cached results for duplicate POST/PUT/DELETE/PATCH requests
	idempotencyConfig := middleware.DefaultIdempotencyConfig()
	idempotencyConfig.DB = s.DB()
	s.Middleware.Idempotency = middleware.NewIdempotencyMiddleware(idempotencyConfig)
	s.app.Use(s.Middleware.Idempotency.Middleware())
	log.Info().
		Str("header", idempotencyConfig.HeaderName).
		Dur("ttl", idempotencyConfig.TTL).
		Msg("Idempotency key support enabled")

	// Compression middleware
	s.app.Use(compress.New(compress.Config{
		Level: compress.LevelDefault,
	}))
}

// setupRoutes sets up all routes
func (s *Server) setupRoutes() {
	s.auditRoutesAtStartup()

	if err := s.registerRoutesViaRegistry(); err != nil {
		log.Fatal().Err(err).Msg("Failed to setup routes via registry")
	}

	// Admin UI routes - external package
	if s.config.Admin.Enabled {
		if s.config.Security.SetupToken == "" {
			log.Error().Msg("Admin UI is enabled but FLUXBASE_SECURITY_SETUP_TOKEN is not set. Admin UI will not be registered for security reasons.")
		} else {
			adminUI := adminui.New(s.config.GetPublicBaseURL())
			adminUI.RegisterRoutes(s.app)
		}
	}

	s.app.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{
			"error": "Not Found",
			"path":  c.Path(),
		})
	})
}

// auditRoutesAtStartup logs route audit information for security review
func (s *Server) auditRoutesAtStartup() {
	entries := s.auditRegisteredRoutes()
	publicCount := 0
	authRequiredCount := 0
	for _, e := range entries {
		if e.Public {
			publicCount++
		}
		if e.Auth == "required" || e.Auth == "service_key" || e.Auth == "dashboard" {
			authRequiredCount++
		}
	}
	log.Info().
		Int("total", len(entries)).
		Int("public", publicCount).
		Int("auth_required", authRequiredCount).
		Msg("Route audit completed")
}

// handleHealth handles health check requests
func (s *Server) handleHealth(c fiber.Ctx) error {
	// Check database health
	ctx, cancel := context.WithTimeout(c.RequestCtx(), 5*time.Second)
	defer cancel()

	dbHealthy := true
	if err := s.db.Health(ctx); err != nil {
		dbHealthy = false
		log.Error().Err(err).Msg("Database health check failed")
	}

	status := "ok"
	httpStatus := fiber.StatusOK
	if !dbHealthy {
		status = "degraded"
		httpStatus = fiber.StatusServiceUnavailable
	}

	// Base response (public)
	response := fiber.Map{
		"status":    status,
		"timestamp": time.Now().UTC(),
	}

	// Add service details for authenticated admin users
	role, hasRole := GetUserRole(c)
	if hasRole && (role == "admin" || role == "instance_admin" || role == "service_role" || role == "tenant_admin") {
		response["services"] = fiber.Map{
			"database": dbHealthy,
			"realtime": true, // WebSocket server is part of this process
		}
	}

	return c.Status(httpStatus).JSON(response)
}

func (s *Server) handleGetTables(c fiber.Ctx) error {
	ctx := context.Background()

	// Add auth context for audit logging
	if userID, ok := GetUserID(c); ok {
		if userRole, ok := GetUserRole(c); ok {
			ctx = database.ContextWithAuth(ctx, userID, userRole, userRole == "admin" || userRole == "service_role")
		}
	}

	inspector := s.db.Inspector()
	tenantPool := middleware.GetTenantPool(c)

	var schemasToQuery []string
	schemaParam := c.Query("schema")

	if schemaParam != "" {
		schemasToQuery = []string{schemaParam}
	} else {
		var schemas []string
		var err error
		if tenantPool != nil {
			schemas, err = inspector.GetSchemasFromPool(ctx, tenantPool)
		} else {
			schemas, err = inspector.GetSchemas(ctx)
		}
		if err != nil {
			return SendOperationFailed(c, "list schemas")
		}

		for _, schema := range schemas {
			if schema == "information_schema" || schema == "pg_catalog" || schema == "pg_toast" {
				continue
			}
			schemasToQuery = append(schemasToQuery, schema)
		}
	}

	var allItems []database.TableInfo
	for _, schema := range schemasToQuery {
		var tables, views, matviews []database.TableInfo
		var err error

		if tenantPool != nil {
			tables, err = inspector.GetAllTablesFromPool(ctx, tenantPool, schema)
		} else {
			tables, err = inspector.GetAllTables(ctx, schema)
		}
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to get tables from schema")
		} else {
			allItems = append(allItems, tables...)
		}

		if tenantPool != nil {
			views, err = inspector.GetAllViewsFromPool(ctx, tenantPool, schema)
		} else {
			views, err = inspector.GetAllViews(ctx, schema)
		}
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to get views from schema")
		} else {
			allItems = append(allItems, views...)
		}

		if tenantPool != nil {
			matviews, err = inspector.GetAllMaterializedViewsFromPool(ctx, tenantPool, schema)
		} else {
			matviews, err = inspector.GetAllMaterializedViews(ctx, schema)
		}
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to get materialized views from schema")
		} else {
			allItems = append(allItems, matviews...)
		}
	}

	return c.JSON(allItems)
}

func (s *Server) handleGetTableSchema(c fiber.Ctx) error {
	ctx := context.Background()
	schema := c.Params("schema")
	table := c.Params("table")

	if schema == "" || table == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Schema and table parameters are required",
		})
	}

	var tableInfo *database.TableInfo
	var err error
	if pool := middleware.GetTenantPool(c); pool != nil {
		tableInfo, err = s.db.Inspector().GetTableInfoFromPool(ctx, pool, schema, table)
	} else {
		tableInfo, err = s.db.Inspector().GetTableInfo(ctx, schema, table)
	}
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": fmt.Sprintf("Table not found: %s.%s", schema, table),
		})
	}

	return c.JSON(tableInfo)
}

func (s *Server) handleGetSchemas(c fiber.Ctx) error {
	ctx := context.Background()

	// Add auth context for audit logging
	if userID, ok := GetUserID(c); ok {
		if userRole, ok := GetUserRole(c); ok {
			ctx = database.ContextWithAuth(ctx, userID, userRole, userRole == "admin" || userRole == "service_role")
		}
	}

	var schemas []string
	var err error
	if pool := middleware.GetTenantPool(c); pool != nil {
		schemas, err = s.db.Inspector().GetSchemasFromPool(ctx, pool)
	} else {
		schemas, err = s.db.Inspector().GetSchemas(ctx)
	}
	if err != nil {
		return SendOperationFailed(c, "list schemas")
	}

	// Filter out system schemas
	var userSchemas []string
	for _, schema := range schemas {
		if schema != "information_schema" && schema != "pg_catalog" && schema != "pg_toast" {
			userSchemas = append(userSchemas, schema)
		}
	}

	return c.JSON(userSchemas)
}

func (s *Server) handleExecuteQuery(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "Execute query endpoint - to be implemented"})
}

// InvalidateSchemaCache invalidates the REST API schema cache.
// This should be called after schema changes (e.g., migrations, DDL operations)
// to ensure the cached metadata is refreshed.
func (s *Server) InvalidateSchemaCache(ctx context.Context) error {
	schemaCache := s.rest.SchemaCache()
	if schemaCache == nil {
		return fmt.Errorf("schema cache not initialized")
	}

	// Invalidate and refresh the schema cache
	schemaCache.InvalidateAll(ctx)
	log.Debug().Msg("Schema cache invalidated and refresh triggered")

	return nil
}

// handleRefreshSchema refreshes the REST API schema cache without requiring a server restart
func (s *Server) handleRefreshSchema(c fiber.Ctx) error {
	log.Info().Msg("Schema refresh requested")

	// Get the schema cache from the REST handler
	schemaCache := s.rest.SchemaCache()
	if schemaCache == nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Schema cache not initialized",
		})
	}

	// Force refresh the schema cache
	if err := schemaCache.Refresh(c.RequestCtx()); err != nil {
		log.Error().Err(err).Msg("Failed to refresh schema cache")
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to refresh schema cache",
			"details": err.Error(),
		})
	}

	log.Info().
		Int("tables", schemaCache.TableCount()).
		Int("views", schemaCache.ViewCount()).
		Msg("Schema cache refreshed successfully")

	return c.JSON(fiber.Map{
		"message": "Schema cache refreshed successfully",
		"tables":  schemaCache.TableCount(),
		"views":   schemaCache.ViewCount(),
	})
}

// Start starts the HTTP server
func (s *Server) Start() error {
	return s.app.Listen(s.config.Server.Address, fiber.ListenConfig{EnablePrefork: false, DisableStartupMessage: !s.config.Debug})
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	// Stop leader electors first (releases advisory locks)
	if s.Scaling.FunctionsLeader != nil {
		log.Info().Msg("Stopping functions scheduler leader election")
		s.Scaling.FunctionsLeader.Stop()
	}
	if s.Scaling.JobsLeader != nil {
		log.Info().Msg("Stopping jobs scheduler leader election")
		s.Scaling.JobsLeader.Stop()
	}
	if s.Scaling.RPCLeader != nil {
		log.Info().Msg("Stopping RPC scheduler leader election")
		s.Scaling.RPCLeader.Stop()
	}

	// Stop realtime listener (PostgreSQL LISTEN/NOTIFY)
	if s.Realtime.Listener != nil {
		log.Info().Msg("Stopping realtime listener")
		s.Realtime.Listener.Stop()
	}

	// Shutdown realtime manager (close all WebSocket connections)
	if s.Realtime.Manager != nil {
		log.Info().Msg("Closing WebSocket connections")
		s.Realtime.Manager.Shutdown()
	}

	// Stop edge functions scheduler
	if s.Functions.Scheduler != nil {
		s.Functions.Scheduler.Stop()
	}

	// Stop jobs scheduler and manager
	if s.Jobs.Scheduler != nil {
		s.Jobs.Scheduler.Stop()
	}
	if s.Jobs.Manager != nil {
		s.Jobs.Manager.Stop()
	}

	// Stop RPC scheduler
	if s.RPC.Scheduler != nil {
		s.RPC.Scheduler.Stop()
	}

	// Stop RPC executor (cancels async executions)
	if s.RPC.Handler != nil {
		s.RPC.Handler.GetExecutor().Stop()
	}

	// Stop webhook trigger service
	if s.Webhook.Trigger != nil {
		s.Webhook.Trigger.Stop()
	}

	// Close AI conversation manager
	if s.AI.Conversations != nil {
		s.AI.Conversations.Close()
	}

	// Stop idempotency middleware cleanup goroutine
	if s.Middleware.Idempotency != nil {
		s.Middleware.Idempotency.Stop()
	}

	// Stop OAuth handler cleanup goroutines
	if s.Auth.OAuth != nil {
		s.Auth.OAuth.Stop()
	}

	// Stop branch cleanup scheduler
	if s.Branching.Scheduler != nil {
		s.Branching.Scheduler.Stop()
	}

	// Close database branching components
	if s.Branching.Router != nil {
		log.Info().Msg("Closing branch connection pools")
		s.Branching.Router.CloseAllPools()
	}
	if s.Branching.Manager != nil {
		log.Info().Msg("Closing branch manager")
		s.Branching.Manager.Close()
	}

	// Stop metrics uptime goroutine
	if s.Metrics.StopChan != nil {
		close(s.Metrics.StopChan)
	}

	// Shutdown metrics server
	if s.Metrics.Server != nil {
		if err := s.Metrics.Server.Shutdown(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to shutdown metrics server")
		}
	}

	// Shutdown OpenTelemetry tracer (flush remaining spans)
	if s.tracer != nil {
		if err := s.tracer.Shutdown(ctx); err != nil {
			log.Warn().Err(err).Msg("Failed to shutdown OpenTelemetry tracer")
		}
	}

	// Stop retention cleanup service
	if s.Logging.Retention != nil {
		log.Info().Msg("Stopping log retention cleanup service")
		s.Logging.Retention.Stop()
	}

	// Close central logging service (flush remaining log entries)
	if s.Logging.Service != nil {
		log.Info().Msg("Closing central logging service")
		if err := s.Logging.Service.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close logging service")
		}
	}

	// Close schema cache (stops invalidation listener)
	if s.Schema.Cache != nil {
		s.Schema.Cache.Close()
	}

	// Close server-owned pub/sub (releases PostgreSQL LISTEN connection)
	if s.pubSub != nil {
		log.Info().Msg("Closing pub/sub")
		if err := s.pubSub.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close pub/sub")
		}
	}

	// Close server-owned rate limit store
	if s.rateLimiter != nil {
		log.Info().Msg("Closing rate limit store")
		if err := s.rateLimiter.Close(); err != nil {
			log.Warn().Err(err).Msg("Failed to close rate limit store")
		}
	}

	log.Info().Msg("Shutting down HTTP server")
	return s.app.ShutdownWithContext(ctx)
}

// App returns the underlying Fiber app instance for testing
func (s *Server) App() *fiber.App {
	return s.app
}

// GetStorageService returns the base storage service from the storage handler
// Note: For tenant-specific storage, use GetStorageConfig with storage.Manager
func (s *Server) GetStorageService() *storage.Service {
	if s.Storage.Handler == nil || s.Storage.Handler.storageManager == nil {
		return nil
	}
	return s.Storage.Handler.storageManager.GetBaseService()
}

// GetWebhookTriggerService returns the webhook trigger service for testing
func (s *Server) GetWebhookTriggerService() *webhook.TriggerService {
	return s.Webhook.Trigger
}

// GetAuthService returns the auth service from the auth handler
func (s *Server) GetAuthService() *auth.Service {
	if s.Auth.Handler == nil {
		return nil
	}
	return s.Auth.Handler.authService
}

// GetLoggingService returns the central logging service
func (s *Server) GetLoggingService() *logging.Service {
	return s.Logging.Service
}

// SetTenantConfigLoader sets the tenant configuration loader
// This is called after migrations complete to enable tenant-specific config overrides
func (s *Server) SetTenantConfigLoader(loader *config.TenantConfigLoader) {
	s.tenantConfigLoader = loader
}

// GetTenantConfigLoader returns the tenant configuration loader
func (s *Server) GetTenantConfigLoader() *config.TenantConfigLoader {
	return s.tenantConfigLoader
}

// SchemaCache returns the REST API schema cache
// This is exposed for testing purposes to refresh the cache after creating tables
func (s *Server) SchemaCache() *database.SchemaCache {
	return s.Schema.Cache
}

// LoadFunctionsFromFilesystem loads edge functions from the filesystem
// This is called at boot time if auto_load_on_boot is enabled
func (s *Server) LoadFunctionsFromFilesystem(ctx context.Context) error {
	if s.Functions.Handler == nil {
		return fmt.Errorf("functions handler not initialized")
	}
	return s.Functions.Handler.LoadFromFilesystem(ctx)
}

// LoadJobsFromFilesystem loads job functions from the filesystem
// This is called at boot time if auto_load_on_boot is enabled
func (s *Server) LoadJobsFromFilesystem(ctx context.Context) error {
	if s.Jobs.Handler == nil {
		return fmt.Errorf("jobs handler not initialized")
	}
	// Use "default" as the namespace for jobs loaded at boot
	return s.Jobs.Handler.LoadFromFilesystem(ctx, "default")
}

// LoadAIChatbotsFromFilesystem loads AI chatbots from the filesystem
// This is called at boot time if auto_load_on_boot is enabled
func (s *Server) LoadAIChatbotsFromFilesystem(ctx context.Context) error {
	if s.AI.Handler == nil {
		return fmt.Errorf("AI handler not initialized")
	}
	return s.AI.Handler.AutoLoadChatbots(ctx)
}

// customErrorHandler handles errors globally
func customErrorHandler(c fiber.Ctx, err error) error {
	// Default to 500 status code
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	// Check if it's a Fiber error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		message = e.Message
	}

	// Log error
	if code >= 500 {
		log.Error().Err(err).Str("path", c.Path()).Msg("Server error")
	}

	// Return JSON error response
	return c.Status(code).JSON(fiber.Map{
		"error": message,
		"code":  code,
	})
}

// handleRealtimeStats returns realtime statistics
// Admin-only endpoint - non-admin users receive 403 Forbidden
func (s *Server) handleRealtimeStats(c fiber.Ctx) error {
	// Check if user has admin role
	role, _ := c.Locals("user_role").(string)
	if role != "admin" && role != "instance_admin" && role != "tenant_admin" && role != "service_role" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required to view realtime stats",
		})
	}

	// Parse pagination parameters
	const defaultLimit = 25
	const maxLimit = 100
	limit := fiber.Query[int](c, "limit", defaultLimit)
	offset := fiber.Query[int](c, "offset", 0)
	search := strings.ToLower(c.Query("search", ""))

	limit, offset = NormalizePaginationParams(limit, offset, defaultLimit, maxLimit)

	// Get all connections from the manager
	manager := s.Realtime.Handler.GetManager()
	allConnections := manager.GetConnectionsForStats()

	// Build a map of user IDs to emails by querying the database
	userIDs := make([]string, 0)
	for _, conn := range allConnections {
		if conn.UserID != nil {
			userIDs = append(userIDs, *conn.UserID)
		}
	}

	// Lookup user emails and display names
	type userInfo struct {
		email       string
		displayName *string
	}
	userInfoMap := make(map[string]userInfo)
	if len(userIDs) > 0 {
		query := `SELECT id, email, raw_user_meta_data->>'display_name' as display_name FROM auth.users WHERE id = ANY($1)`
		rows, err := s.db.Query(c.RequestCtx(), query, userIDs)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, email string
				var displayName *string
				if err := rows.Scan(&id, &email, &displayName); err == nil {
					userInfoMap[id] = userInfo{
						email:       email,
						displayName: displayName,
					}
				}
			}
		}
	}

	// Enrich connections with emails and display names
	enrichedConnections := make([]realtime.ConnectionInfo, 0, len(allConnections))
	for _, conn := range allConnections {
		if conn.UserID != nil {
			if info, ok := userInfoMap[*conn.UserID]; ok {
				conn.Email = &info.email
				conn.DisplayName = info.displayName
			}
		}
		enrichedConnections = append(enrichedConnections, conn)
	}

	// Apply search filter (case-insensitive)
	var filteredConnections []realtime.ConnectionInfo
	if search != "" {
		for _, conn := range enrichedConnections {
			// Search by connection ID, user ID, email, display name, or IP address
			if strings.Contains(strings.ToLower(conn.ID), search) ||
				strings.Contains(strings.ToLower(conn.RemoteAddr), search) ||
				(conn.UserID != nil && strings.Contains(strings.ToLower(*conn.UserID), search)) ||
				(conn.Email != nil && strings.Contains(strings.ToLower(*conn.Email), search)) ||
				(conn.DisplayName != nil && strings.Contains(strings.ToLower(*conn.DisplayName), search)) {
				filteredConnections = append(filteredConnections, conn)
			}
		}
	} else {
		filteredConnections = enrichedConnections
	}

	// Calculate total before pagination
	total := len(filteredConnections)

	// Apply pagination
	if offset >= len(filteredConnections) {
		filteredConnections = []realtime.ConnectionInfo{}
	} else {
		filteredConnections = filteredConnections[offset:]
	}
	if len(filteredConnections) > limit {
		filteredConnections = filteredConnections[:limit]
	}

	return c.JSON(fiber.Map{
		"total_connections": total,
		"connections":       filteredConnections,
		"limit":             limit,
		"offset":            offset,
	})
}

// BroadcastRequest represents a broadcast request
type BroadcastRequest struct {
	Channel string      `json:"channel"`
	Message interface{} `json:"message"`
}

// handleRealtimeBroadcast broadcasts a message to a channel
func (s *Server) handleRealtimeBroadcast(c fiber.Ctx) error {
	var req BroadcastRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Channel == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Channel is required",
		})
	}

	// Get the realtime manager and broadcast to the channel
	if s.Realtime.Handler == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Realtime service not available",
		})
	}

	manager := s.Realtime.Handler.GetManager()
	recipientCount := manager.BroadcastToChannel(req.Channel, realtime.ServerMessage{
		Type:    realtime.MessageTypeBroadcast,
		Channel: req.Channel,
		Payload: map[string]interface{}{
			"broadcast": map[string]interface{}{
				"event":   "broadcast",
				"payload": req.Message,
			},
		},
	})

	return c.JSON(fiber.Map{
		"success":    true,
		"channel":    req.Channel,
		"recipients": recipientCount,
	})
}

// fiber:context-methods migrated
