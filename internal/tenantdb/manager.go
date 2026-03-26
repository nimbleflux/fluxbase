package tenantdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

var (
	ErrMaxTenantsReached   = errors.New("maximum number of tenants reached")
	ErrCannotDeleteDefault = errors.New("cannot delete the default tenant")
	ErrTenantNotActive     = errors.New("tenant is not in active state")
)

type Manager struct {
	storage        *Storage
	config         Config
	adminPool      *pgxpool.Pool
	router         *Router
	dbURL          string
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
		"CREATE DATABASE %s OWNER postgres ENCODING 'UTF8'",
		dbName,
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

	if m.config.Migrations.OnCreate {
		if err := m.runSystemMigrationsForDB(ctx, dbName); err != nil {
			if _, dropErr := m.adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)); dropErr != nil {
				log.Warn().Err(dropErr).Str("db", dbName).Msg("Failed to drop database after migration failure")
			}
			if statusErr := m.storage.UpdateTenantStatus(ctx, tenant.ID, TenantStatusError); statusErr != nil {
				log.Warn().Err(statusErr).Str("tenant_id", tenant.ID).Msg("Failed to update tenant status to error")
			}
			if deleteErr := m.storage.HardDeleteTenant(ctx, tenant.ID); deleteErr != nil {
				log.Warn().Err(deleteErr).Str("tenant_id", tenant.ID).Msg("Failed to hard delete tenant record")
			}
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
		log.Info().Str("tenant", req.Slug).Msg("Completed system migrations")
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

	if m.router != nil {
		m.router.RemovePool(tenantID)
	}

	_, err = m.adminPool.Exec(ctx, fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid()
	`, *tenant.DBName))
	if err != nil {
		log.Warn().Err(err).Str("db", *tenant.DBName).Msg("Failed to terminate connections")
	}

	_, err = m.adminPool.Exec(ctx, fmt.Sprintf("DROP DATABASE IF EXISTS %s", *tenant.DBName))
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
	dbURL := fmt.Sprintf("%s%s", m.dbURL, dbName)

	migrator, err := migrate.New("file:///migrations", dbURL)
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
	return m.storage
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
