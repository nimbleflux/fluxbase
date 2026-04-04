package tenantdb

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database/bootstrap"
)

var (
	ErrMaxTenantsReached   = errors.New("maximum number of tenants reached")
	ErrCannotDeleteDefault = errors.New("cannot delete the default tenant")
	ErrTenantNotActive     = errors.New("tenant is not in active state")
)

// StorageQuerier covers the storage methods Manager calls.
type StorageQuerier interface {
	CountTenants(ctx context.Context) (int, error)
	CreateTenant(ctx context.Context, tenant *Tenant) error
	GetTenant(ctx context.Context, id string) (*Tenant, error)
	GetAllActiveTenants(ctx context.Context) ([]Tenant, error)
	UpdateTenantStatus(ctx context.Context, id string, status TenantStatus) error
	UpdateTenantDBName(ctx context.Context, id string, dbName string) error
	SoftDeleteTenant(ctx context.Context, id string) error
	HardDeleteTenant(ctx context.Context, id string) error
	CleanupTenantData(ctx context.Context, tenantID string) error
}

type Manager struct {
	storage        StorageQuerier
	config         Config
	adminPool      *pgxpool.Pool
	router         *Router
	dbURL          string
	adminDBURL     string
	fdwConfig      *FDWConfig
	declarative    *DeclarativeService
	declarativeCfg DeclarativeConfig
}

func NewManager(
	storage *Storage,
	config Config,
	adminPool *pgxpool.Pool,
	dbURL string,
) *Manager {
	return &Manager{
		storage:   storage,
		config:    config,
		adminPool: adminPool,
		dbURL:     dbURL,
	}
}

// SetDeclarativeService sets the declarative schema service for tenant databases
func (m *Manager) SetDeclarativeService(svc *DeclarativeService) {
	m.declarative = svc
}

// SetDeclarativeConfig sets the declarative schema configuration
func (m *Manager) SetDeclarativeConfig(cfg DeclarativeConfig) {
	m.declarativeCfg = cfg
}

// SetFDWConfig sets the FDW configuration for tenant databases
func (m *Manager) SetFDWConfig(cfg FDWConfig) {
	m.fdwConfig = &cfg
}

// SetAdminDBURL sets the admin database URL used for bootstrap operations
// that require elevated privileges (e.g., CREATE EXTENSION).
func (m *Manager) SetAdminDBURL(url string) {
	m.adminDBURL = url
}

func (m *Manager) SetRouter(router *Router) {
	m.router = router
}

func (m *Manager) CreateTenantDatabase(ctx context.Context, req CreateTenantRequest) (*Tenant, error) {
	if m.config.MaxTenants > 0 {
		count, err := m.storage.CountTenants(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to check tenant count: %w", err)
		}
		if count >= m.config.MaxTenants {
			return nil, ErrMaxTenantsReached
		}
	}

	dbName := fmt.Sprintf("%s%s", m.config.DatabasePrefix, req.Slug)

	tenant := &Tenant{
		Slug:     req.Slug,
		Name:     req.Name,
		Status:   TenantStatusCreating,
		DBName:   &dbName,
		Metadata: req.Metadata,
	}
	if tenant.Metadata == nil {
		tenant.Metadata = make(map[string]any)
	}

	if err := m.storage.CreateTenant(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant record: %w", err)
	}

	_, err := m.adminPool.Exec(ctx, fmt.Sprintf(
		"CREATE DATABASE %s ENCODING 'UTF8'",
		quoteIdent(dbName),
	))
	if err != nil {
		if statusErr := m.storage.UpdateTenantStatus(ctx, tenant.ID, TenantStatusError); statusErr != nil {
			log.Warn().Err(statusErr).Str("tenant_id", tenant.ID).Msg("Failed to update tenant status to error")
		}
		if deleteErr := m.storage.HardDeleteTenant(ctx, tenant.ID); deleteErr != nil {
			log.Warn().Err(deleteErr).Str("tenant_id", tenant.ID).Msg("Failed to hard delete tenant record")
		}
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	log.Info().Str("tenant", req.Slug).Str("db", dbName).Msg("Created tenant database")

	// Bootstrap tenant database (create schemas, roles, privileges)
	// Use admin URL if available (required for CREATE EXTENSION), otherwise app URL
	bootstrapBaseURL := m.adminDBURL
	if bootstrapBaseURL == "" {
		bootstrapBaseURL = m.dbURL
	}
	tenantDBURL := replaceDBName(bootstrapBaseURL, dbName)
	if err := bootstrap.RunBootstrapOnDB(ctx, tenantDBURL); err != nil {
		log.Warn().Err(err).Str("tenant", req.Slug).Msg("Failed to bootstrap tenant database")
		if statusErr := m.storage.UpdateTenantStatus(ctx, tenant.ID, TenantStatusError); statusErr != nil {
			log.Warn().Err(statusErr).Str("tenant_id", tenant.ID).Msg("Failed to update tenant status to error")
		}
		return nil, fmt.Errorf("failed to bootstrap tenant database: %w", err)
	}
	log.Info().Str("tenant", req.Slug).Msg("Bootstrapped tenant database")

	// Set up FDW so tenant can access auth.users from the main database.
	// FDW requires elevated privileges, so use admin connection if available.
	if m.fdwConfig != nil && m.adminDBURL != "" {
		fdwURL := replaceDBName(m.adminDBURL, dbName)
		fdwPool, fdwPoolErr := pgxpool.New(ctx, fdwURL)
		if fdwPoolErr != nil {
			log.Warn().Err(fdwPoolErr).Str("tenant", req.Slug).Msg("Failed to create admin pool for FDW setup")
		} else {
			defer fdwPool.Close()
			if fdwErr := SetupFDW(ctx, fdwPool, *m.fdwConfig, nil); fdwErr != nil {
				log.Warn().Err(fdwErr).Str("tenant", req.Slug).Msg("Failed to set up FDW for tenant database")
			} else {
				// Also create user mapping for the app user so queries via
				// the router pool (app credentials) can use the foreign server.
				if appUser := extractDBUser(m.dbURL); appUser != "" {
					userMappingSQL := fmt.Sprintf(
						`CREATE USER MAPPING IF NOT EXISTS FOR %s SERVER %s OPTIONS (user '%s'`,
						quoteIdent(appUser), quoteIdent(fdwServerName), escapeSQLString(m.fdwConfig.User),
					)
					if m.fdwConfig.Password != "" {
						userMappingSQL += fmt.Sprintf(`, password '%s'`, escapeSQLString(m.fdwConfig.Password))
					}
					userMappingSQL += ")"
					if _, mapErr := fdwPool.Exec(ctx, userMappingSQL); mapErr != nil {
						log.Warn().Err(mapErr).Str("tenant", req.Slug).Msg("Failed to create app user mapping for FDW")
					}
				}
				log.Info().Str("tenant", req.Slug).Msg("Set up FDW for tenant database")
			}
		}
	}

	// Apply tenant-specific declarative schema if configured
	if m.declarative != nil && m.declarativeCfg.OnCreate {
		if m.declarative.HasSchemaFile(req.Slug) {
			if err := m.declarative.ApplyTenantSchema(ctx, tenant); err != nil {
				log.Warn().Err(err).Str("tenant", req.Slug).Msg("Failed to apply tenant declarative schema")
				// Don't fail the entire creation, just log the warning
			} else {
				log.Info().Str("tenant", req.Slug).Msg("Applied tenant declarative schema")
			}
		}
	}

	if err := m.storage.UpdateTenantStatus(ctx, tenant.ID, TenantStatusActive); err != nil {
		log.Error().Err(err).Str("tenant_id", tenant.ID).Msg("Failed to update tenant status")
	}
	tenant.Status = TenantStatusActive

	log.Info().Str("tenant_id", tenant.ID).Str("slug", req.Slug).Str("db", dbName).Msg("Tenant database created successfully")
	return tenant, nil
}

func (m *Manager) DeleteTenantDatabase(ctx context.Context, tenantID string) error {
	tenant, err := m.storage.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant.IsDefault {
		return ErrCannotDeleteDefault
	}

	if tenant.DBName == nil {
		return fmt.Errorf("tenant has no separate database")
	}

	if err := m.storage.SoftDeleteTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to soft delete tenant: %w", err)
	}

	// Remove connection pool from cache so no new connections are created.
	if m.router != nil {
		m.router.RemovePool(tenantID)
	}

	// Cleanup tenant-related data from the main database.
	// Must happen before HardDeleteTenant because branching.branches has
	// RESTRICT FK and would block the tenant row deletion.
	if err := m.storage.CleanupTenantData(ctx, tenantID); err != nil {
		if statusErr := m.storage.UpdateTenantStatus(ctx, tenantID, TenantStatusError); statusErr != nil {
			log.Warn().Err(statusErr).Str("tenant_id", tenantID).Msg("Failed to update tenant status to error")
		}
		return fmt.Errorf("failed to cleanup tenant data: %w", err)
	}

	// Atomically terminate connections and drop the database (PostgreSQL 13+).
	_, err = m.adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", quoteIdent(*tenant.DBName)))
	if err != nil {
		if statusErr := m.storage.UpdateTenantStatus(ctx, tenantID, TenantStatusError); statusErr != nil {
			log.Warn().Err(statusErr).Str("tenant_id", tenantID).Msg("Failed to update tenant status to error")
		}
		return fmt.Errorf("failed to drop database: %w", err)
	}

	if err := m.storage.HardDeleteTenant(ctx, tenantID); err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Failed to hard delete tenant record")
	}

	log.Info().Str("tenant_id", tenantID).Str("slug", tenant.Slug).Msg("Tenant database deleted successfully")
	return nil
}

func (m *Manager) MigrateTenant(ctx context.Context, tenantID string) error {
	tenant, err := m.storage.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if tenant.UsesMainDatabase() {
		return nil
	}

	if m.router == nil {
		return fmt.Errorf("router not initialized")
	}

	_, err = m.router.GetPool(tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant pool: %w", err)
	}

	return m.runSystemMigrationsForDB(ctx, *tenant.DBName)
}

func (m *Manager) StartMigrationWorker(ctx context.Context) {
	if !m.config.Migrations.Background {
		return
	}

	ticker := time.NewTicker(m.config.Migrations.CheckInterval)
	defer ticker.Stop()

	log.Info().Dur("interval", m.config.Migrations.CheckInterval).Msg("Starting tenant migration worker")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Tenant migration worker stopped")
			return
		case <-ticker.C:
			m.migrateAllTenants(ctx)
		}
	}
}

func (m *Manager) migrateAllTenants(ctx context.Context) {
	tenants, err := m.storage.GetAllActiveTenants(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get tenants for migration")
		return
	}

	for _, tenant := range tenants {
		if tenant.UsesMainDatabase() {
			continue
		}

		if m.router == nil {
			continue
		}

		pool, err := m.router.GetPool(tenant.ID)
		if err != nil {
			log.Error().Err(err).Str("tenant", tenant.Slug).Msg("Failed to get pool for migration check")
			continue
		}

		pending, err := m.hasPendingMigrations(ctx, pool)
		if err != nil {
			log.Debug().Err(err).Str("tenant", tenant.Slug).Msg("Failed to check pending migrations")
			continue
		}

		if pending {
			log.Info().Str("tenant", tenant.Slug).Msg("Migrating tenant database")
			if err := m.runSystemMigrationsForDB(ctx, *tenant.DBName); err != nil {
				log.Error().Err(err).Str("tenant", tenant.Slug).Msg("Tenant migration failed")
			} else {
				log.Info().Str("tenant", tenant.Slug).Msg("Tenant migration completed")
			}
		}
	}
}

func (m *Manager) hasPendingMigrations(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var currentVersion int
	err := pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0)::int FROM migrations.fluxbase
	`).Scan(&currentVersion)
	if err != nil {
		return true, nil
	}

	return false, nil
}

func (m *Manager) runSystemMigrationsForDB(ctx context.Context, dbName string) error {
	const migrationsDir = "/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil
	}

	dbURL := replaceDBName(m.dbURL, dbName)

	migrator, err := migrate.New("file://"+migrationsDir, dbURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer func() {
		if _, closeErr := migrator.Close(); closeErr != nil {
			log.Warn().Err(closeErr).Str("db", dbName).Msg("Failed to close migrator")
		}
	}()

	if err := migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func (m *Manager) GetStorage() *Storage {
	if s, ok := m.storage.(*Storage); ok {
		return s
	}
	return nil
}

func (m *Manager) GetRouter() *Router {
	return m.router
}

func (m *Manager) GetConfig() Config {
	return m.config
}

func (m *Manager) GetDeclarativeService() *DeclarativeService {
	return m.declarative
}

// ApplyDeclarativeSchemas applies declarative schemas to all tenants with schema files
// This is called on startup if configured
func (m *Manager) ApplyDeclarativeSchemas(ctx context.Context) error {
	if m.declarative == nil {
		return nil
	}

	tenants, err := m.storage.GetAllActiveTenants(ctx)
	if err != nil {
		return fmt.Errorf("failed to list tenants: %w", err)
	}

	for i := range tenants {
		if m.declarative.HasSchemaFile(tenants[i].Slug) {
			if err := m.declarative.ApplyTenantSchema(ctx, &tenants[i]); err != nil {
				log.Error().Err(err).Str("tenant", tenants[i].Slug).Msg("Failed to apply declarative schema")
				// Continue with other tenants
			}
		}
	}

	return nil
}

// ApplyTenantDeclarativeSchema applies the declarative schema for a specific tenant
func (m *Manager) ApplyTenantDeclarativeSchema(ctx context.Context, tenantID string) error {
	if m.declarative == nil {
		return fmt.Errorf("declarative schema service not configured")
	}

	tenant, err := m.storage.GetTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	if !m.declarative.HasSchemaFile(tenant.Slug) {
		return fmt.Errorf("no declarative schema file found for tenant %s", tenant.Slug)
	}

	return m.declarative.ApplyTenantSchema(ctx, tenant)
}

// GetTenantSchemaStatus returns the schema status for a specific tenant
func (m *Manager) GetTenantSchemaStatus(ctx context.Context, tenantID string) (*TenantSchemaStatus, error) {
	if m.declarative == nil {
		return nil, fmt.Errorf("declarative schema service not configured")
	}

	tenant, err := m.storage.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return m.declarative.GetTenantSchemaStatus(ctx, tenant)
}

// replaceDBName constructs a database URL for a specific database name
// by parsing the base URL and replacing the path component.
func replaceDBName(baseDBURL, dbName string) string {
	u, err := url.Parse(baseDBURL)
	if err != nil {
		return baseDBURL + dbName // fallback
	}
	u.Path = "/" + dbName
	return u.String()
}

// extractDBUser returns the username from a database URL, or empty string on error.
func extractDBUser(dbURL string) string {
	u, err := url.Parse(dbURL)
	if err != nil {
		return ""
	}
	return u.User.Username()
}
