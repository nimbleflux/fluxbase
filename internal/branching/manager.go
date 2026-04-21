package branching

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
)

type TenantDatabaseInfo struct {
	DBName    string
	Slug      string
	IsDefault bool
}

type TenantResolver interface {
	GetTenantDatabase(ctx context.Context, tenantID uuid.UUID) (*TenantDatabaseInfo, error)
}

type FDWRepairer interface {
	RepairFDWForBranch(ctx context.Context, branchDBURL string, tenantID uuid.UUID) error
}

// Manager handles database operations for branches
type Manager struct {
	storage        *Storage
	config         config.BranchingConfig
	adminPool      *pgxpool.Pool // Connection pool with CREATE DATABASE privileges
	mainDBName     string        // Name of the main database
	mainDBURL      string        // Connection URL for the main database
	tenantResolver TenantResolver
	fdwRepairer    FDWRepairer
}

// NewManager creates a new branch manager
func NewManager(storage *Storage, cfg config.BranchingConfig, mainPool *pgxpool.Pool, mainDBURL string) (*Manager, error) {
	parsedURL, err := url.Parse(mainDBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main database URL: %w", err)
	}

	mainDBName := strings.TrimPrefix(parsedURL.Path, "/")
	if mainDBName == "" {
		mainDBName = "fluxbase"
	}

	adminParsed := *parsedURL
	adminParsed.Path = "/postgres"
	adminURL := adminParsed.String()

	adminConfig, err := pgxpool.ParseConfig(adminURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin database URL: %w", err)
	}

	adminConfig.MaxConns = 2
	adminConfig.MinConns = 0

	adminPool, err := pgxpool.NewWithConfig(context.Background(), adminConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin connection pool: %w", err)
	}

	return &Manager{
		storage:    storage,
		config:     cfg,
		adminPool:  adminPool,
		mainDBName: mainDBName,
		mainDBURL:  mainDBURL,
	}, nil
}

func (m *Manager) SetTenantResolver(resolver TenantResolver) {
	m.tenantResolver = resolver
}

func (m *Manager) SetFDWRepairer(repairer FDWRepairer) {
	m.fdwRepairer = repairer
}

// CreateBranch creates a new database branch
func (m *Manager) CreateBranch(ctx context.Context, req CreateBranchRequest, createdBy *uuid.UUID) (*Branch, error) {
	startTime := time.Now()

	// Check if branching is enabled
	if !m.config.Enabled {
		return nil, ErrBranchingDisabled
	}

	// Resolve tenant_id - default to instance-level (nil) if not specified
	tenantID := req.TenantID

	// Check limits (tenant-scoped)
	if err := m.checkLimits(ctx, tenantID, createdBy); err != nil {
		return nil, err
	}

	// Generate slug from name
	slug := GenerateSlug(req.Name)
	if err := ValidateSlug(slug); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidSlug, err)
	}

	// Check if slug already exists for this tenant
	existing, err := m.storage.GetBranchBySlug(ctx, slug, tenantID)
	if err != nil && !errors.Is(err, ErrBranchNotFound) {
		return nil, fmt.Errorf("failed to check existing branch: %w", err)
	}
	// Check if existing branch belongs to the same tenant
	if existing != nil {
		if (tenantID == nil && existing.TenantID == nil) ||
			(tenantID != nil && existing.TenantID != nil && *tenantID == *existing.TenantID) {
			return nil, ErrBranchExists
		}
	}

	// Determine data clone mode
	dataCloneMode := DataCloneModeSchemaOnly
	if req.DataCloneMode != "" {
		dataCloneMode = req.DataCloneMode
	} else if m.config.DefaultDataCloneMode != "" {
		dataCloneMode = DataCloneMode(m.config.DefaultDataCloneMode)
	}

	// Determine branch type
	branchType := BranchTypePreview
	if req.Type != "" {
		branchType = req.Type
	}

	// Determine parent branch (default to main)
	var parentBranchID *uuid.UUID
	if req.ParentBranchID != nil {
		parentBranchID = req.ParentBranchID
	} else {
		mainBranch, err := m.storage.GetMainBranch(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get main branch: %w", err)
		}
		parentBranchID = &mainBranch.ID
	}

	// Generate database name with tenant prefix
	var databaseName string
	if tenantID != nil {
		// Get tenant slug for database naming
		tenantSlug, err := m.getTenantSlug(ctx, *tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tenant info: %w", err)
		}
		databaseName = GenerateTenantBranchDatabaseName(m.config.DatabasePrefix, tenantSlug, slug)
	} else {
		// Instance-level branch (backward compatible)
		databaseName = GenerateDatabaseName(m.config.DatabasePrefix, slug)
	}

	// Create branch record
	branch := &Branch{
		ID:             uuid.New(),
		Name:           req.Name,
		Slug:           slug,
		DatabaseName:   databaseName,
		Status:         BranchStatusCreating,
		Type:           branchType,
		TenantID:       tenantID,
		ParentBranchID: parentBranchID,
		DataCloneMode:  dataCloneMode,
		GitHubPRNumber: req.GitHubPRNumber,
		GitHubPRURL:    req.GitHubPRURL,
		GitHubRepo:     req.GitHubRepo,
		SeedsPath:      req.SeedsPath,
		CreatedBy:      createdBy,
		ExpiresAt:      req.ExpiresAt,
	}

	// Calculate auto-delete expiration if configured
	if branch.ExpiresAt == nil && m.config.AutoDeleteAfter > 0 && branchType == BranchTypePreview {
		expiresAt := time.Now().Add(m.config.AutoDeleteAfter)
		branch.ExpiresAt = &expiresAt
	}

	if err := m.storage.CreateBranch(ctx, branch); err != nil {
		return nil, fmt.Errorf("failed to create branch record: %w", err)
	}

	// Log activity start
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branch.ID,
		Action:     ActivityActionCreated,
		Status:     ActivityStatusStarted,
		ExecutedBy: createdBy,
		Details:    map[string]any{"data_clone_mode": dataCloneMode},
	})

	// Create the database
	if err := m.createDatabase(ctx, branch, parentBranchID); err != nil {
		// Update status to error
		errMsg := err.Error()
		_ = m.storage.UpdateBranchStatus(ctx, branch.ID, BranchStatusError, &errMsg)

		// Log failure
		durationMs := int(time.Since(startTime).Milliseconds())
		_ = m.storage.LogActivity(ctx, &ActivityLog{
			BranchID:     branch.ID,
			Action:       ActivityActionCreated,
			Status:       ActivityStatusFailed,
			ErrorMessage: &errMsg,
			ExecutedBy:   createdBy,
			DurationMs:   &durationMs,
		})

		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Update status to ready
	if err := m.storage.UpdateBranchStatus(ctx, branch.ID, BranchStatusReady, nil); err != nil {
		return nil, fmt.Errorf("failed to update branch status: %w", err)
	}
	branch.Status = BranchStatusReady

	// Log success
	durationMs := int(time.Since(startTime).Milliseconds())
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branch.ID,
		Action:     ActivityActionCreated,
		Status:     ActivityStatusSuccess,
		ExecutedBy: createdBy,
		DurationMs: &durationMs,
	})

	// Grant creator admin access
	if createdBy != nil {
		_ = m.storage.GrantAccess(ctx, &BranchAccess{
			BranchID:    branch.ID,
			UserID:      *createdBy,
			AccessLevel: BranchAccessAdmin,
			GrantedBy:   createdBy,
		})
	}

	log.Info().
		Str("branch_id", branch.ID.String()).
		Str("slug", slug).
		Str("database", databaseName).
		Int("duration_ms", durationMs).
		Msg("Branch created successfully")

	return branch, nil
}

// DeleteBranch deletes a branch and its database
func (m *Manager) DeleteBranch(ctx context.Context, branchID uuid.UUID, deletedBy *uuid.UUID) error {
	startTime := time.Now()

	// Get the branch (no tenant filter — manager operates across tenants)
	branch, err := m.storage.GetBranch(ctx, branchID, nil)
	if err != nil {
		return err
	}

	// Cannot delete main branch
	if branch.Type == BranchTypeMain {
		return ErrCannotDeleteMainBranch
	}

	// Update status to deleting
	if err := m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusDeleting, nil); err != nil {
		return fmt.Errorf("failed to update branch status: %w", err)
	}

	// Log activity start
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branchID,
		Action:     ActivityActionDeleted,
		Status:     ActivityStatusStarted,
		ExecutedBy: deletedBy,
	})

	// Drop the database
	if err := m.dropDatabase(ctx, branch.DatabaseName); err != nil {
		// Update status to error
		errMsg := err.Error()
		_ = m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusError, &errMsg)

		// Log failure
		durationMs := int(time.Since(startTime).Milliseconds())
		_ = m.storage.LogActivity(ctx, &ActivityLog{
			BranchID:     branchID,
			Action:       ActivityActionDeleted,
			Status:       ActivityStatusFailed,
			ErrorMessage: &errMsg,
			ExecutedBy:   deletedBy,
			DurationMs:   &durationMs,
		})

		return fmt.Errorf("failed to drop database: %w", err)
	}

	// Mark as deleted (no tenant filter — manager is the authority)
	if err := m.storage.DeleteBranch(ctx, branchID, nil); err != nil {
		return fmt.Errorf("failed to delete branch record: %w", err)
	}

	// Log success
	durationMs := int(time.Since(startTime).Milliseconds())
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branchID,
		Action:     ActivityActionDeleted,
		Status:     ActivityStatusSuccess,
		ExecutedBy: deletedBy,
		DurationMs: &durationMs,
	})

	log.Info().
		Str("branch_id", branchID.String()).
		Str("slug", branch.Slug).
		Str("database", branch.DatabaseName).
		Int("duration_ms", durationMs).
		Msg("Branch deleted successfully")

	return nil
}

// ResetBranch resets a branch to its parent state
func (m *Manager) ResetBranch(ctx context.Context, branchID uuid.UUID, resetBy *uuid.UUID) error {
	startTime := time.Now()

	// Get the branch (no tenant filter — manager operates across tenants)
	branch, err := m.storage.GetBranch(ctx, branchID, nil)
	if err != nil {
		return err
	}

	// Cannot reset main branch
	if branch.Type == BranchTypeMain {
		return ErrCannotDeleteMainBranch
	}

	// Need a parent to reset from
	if branch.ParentBranchID == nil {
		return fmt.Errorf("branch has no parent to reset from")
	}

	// Get parent branch
	parent, err := m.storage.GetBranch(ctx, *branch.ParentBranchID, nil)
	if err != nil {
		return fmt.Errorf("failed to get parent branch: %w", err)
	}

	// Update status to migrating (reusing for reset operation)
	if err := m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusMigrating, nil); err != nil {
		return fmt.Errorf("failed to update branch status: %w", err)
	}

	// Log activity start
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branchID,
		Action:     ActivityActionReset,
		Status:     ActivityStatusStarted,
		ExecutedBy: resetBy,
		Details:    map[string]any{"parent_slug": parent.Slug},
	})

	// Drop and recreate the database
	if err := m.dropDatabase(ctx, branch.DatabaseName); err != nil {
		errMsg := err.Error()
		_ = m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusError, &errMsg)
		return fmt.Errorf("failed to drop database for reset: %w", err)
	}

	if err := m.createDatabase(ctx, branch, branch.ParentBranchID); err != nil {
		errMsg := err.Error()
		_ = m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusError, &errMsg)

		// Log failure
		durationMs := int(time.Since(startTime).Milliseconds())
		_ = m.storage.LogActivity(ctx, &ActivityLog{
			BranchID:     branchID,
			Action:       ActivityActionReset,
			Status:       ActivityStatusFailed,
			ErrorMessage: &errMsg,
			ExecutedBy:   resetBy,
			DurationMs:   &durationMs,
		})

		return fmt.Errorf("failed to recreate database: %w", err)
	}

	// Update status to ready
	if err := m.storage.UpdateBranchStatus(ctx, branchID, BranchStatusReady, nil); err != nil {
		return fmt.Errorf("failed to update branch status: %w", err)
	}

	// Log success
	durationMs := int(time.Since(startTime).Milliseconds())
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branchID,
		Action:     ActivityActionReset,
		Status:     ActivityStatusSuccess,
		ExecutedBy: resetBy,
		DurationMs: &durationMs,
	})

	log.Info().
		Str("branch_id", branchID.String()).
		Str("slug", branch.Slug).
		Int("duration_ms", durationMs).
		Msg("Branch reset successfully")

	return nil
}

// checkLimits verifies that branch limits have not been exceeded
func (m *Manager) checkLimits(ctx context.Context, tenantID *uuid.UUID, userID *uuid.UUID) error {
	// Check tenant-specific branch limit first
	if tenantID != nil && m.config.MaxBranchesPerTenant > 0 {
		tenantCount, err := m.storage.CountBranchesByTenant(ctx, *tenantID)
		if err != nil {
			return fmt.Errorf("failed to count tenant branches: %w", err)
		}
		if tenantCount >= m.config.MaxBranchesPerTenant {
			return ErrMaxTenantBranchesReached
		}
	}

	// Check total branch limit
	if m.config.MaxTotalBranches > 0 {
		filter := ListBranchesFilter{}
		if tenantID != nil {
			filter.TenantID = tenantID
		}
		total, err := m.storage.CountBranches(ctx, filter)
		if err != nil {
			return fmt.Errorf("failed to count branches: %w", err)
		}
		if total >= m.config.MaxTotalBranches {
			return ErrMaxBranchesReached
		}
	}

	// Check per-user branch limit
	if m.config.MaxBranchesPerUser > 0 && userID != nil {
		userCount, err := m.storage.CountBranchesByUser(ctx, *userID)
		if err != nil {
			return fmt.Errorf("failed to count user branches: %w", err)
		}
		if userCount >= m.config.MaxBranchesPerUser {
			return ErrMaxUserBranchesReached
		}
	}

	return nil
}

// getTenantSlug retrieves the tenant slug for database naming
func (m *Manager) getTenantSlug(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// Query the platform.tenants table to get the slug
	var slug string
	query := `SELECT slug FROM platform.tenants WHERE id = $1`
	err := m.storage.GetPool().QueryRow(ctx, query, tenantID).Scan(&slug)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant slug: %w", err)
	}
	return slug, nil
}

// createDatabase creates a new database for a branch
func (m *Manager) createDatabase(ctx context.Context, branch *Branch, parentBranchID *uuid.UUID) error {
	dbName := sanitizeIdentifier(branch.DatabaseName)

	switch branch.DataCloneMode {
	case DataCloneModeSchemaOnly:
		return m.createDatabaseSchemaOnly(ctx, branch, parentBranchID)
	case DataCloneModeFullClone:
		return m.createDatabaseFullClone(ctx, branch, parentBranchID)
	case DataCloneModeSeedData:
		return m.createDatabaseSeedData(ctx, branch, parentBranchID)
	default:
		query := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err := m.adminPool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		return nil
	}
}

// resolveTemplateDatabase determines the correct database to use as a TEMPLATE
// for cloning. For tenant branches with a separate database, this returns the
// tenant's database name. Otherwise, it returns the parent branch's database.
func (m *Manager) resolveTemplateDatabase(ctx context.Context, branch *Branch, parentBranchID *uuid.UUID) (string, error) {
	// If the branch belongs to a tenant and we have a tenant resolver, check
	// if the tenant has a separate database to clone from.
	if branch.TenantID != nil && m.tenantResolver != nil {
		tenantInfo, err := m.tenantResolver.GetTenantDatabase(ctx, *branch.TenantID)
		if err != nil {
			log.Warn().Err(err).Str("tenant_id", branch.TenantID.String()).
				Msg("Failed to resolve tenant database, falling back to parent branch")
		} else if tenantInfo != nil && !tenantInfo.IsDefault && tenantInfo.DBName != "" {
			log.Info().
				Str("tenant_db", tenantInfo.DBName).
				Str("branch", branch.Slug).
				Msg("Cloning branch from tenant's separate database")
			return tenantInfo.DBName, nil
		}
	}

	// Fall back to parent branch's database (original behavior)
	if parentBranchID == nil {
		return m.mainDBName, nil
	}
	parent, err := m.storage.GetBranch(ctx, *parentBranchID, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get parent branch: %w", err)
	}
	return parent.DatabaseName, nil
}

// createDatabaseSchemaOnly creates a database with schema only (no data)
func (m *Manager) createDatabaseSchemaOnly(ctx context.Context, branch *Branch, parentBranchID *uuid.UUID) error {
	dbName := sanitizeIdentifier(branch.DatabaseName)

	templateDB, err := m.resolveTemplateDatabase(ctx, branch, parentBranchID)
	if err != nil {
		return err
	}

	createFromTemplate := fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		dbName, sanitizeIdentifier(templateDB))

	_, err = m.adminPool.Exec(ctx, createFromTemplate)
	if err != nil {
		log.Warn().Err(err).Str("template", templateDB).
			Msg("Failed to create from template, creating empty database")
		createQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err = m.adminPool.Exec(ctx, createQuery)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		return nil
	}

	if err := m.repairFDW(ctx, branch); err != nil {
		log.Warn().Err(err).Str("branch", branch.Slug).
			Msg("Failed to repair FDW after cloning, branch may have stale FDW mappings")
	}

	return nil
}

// createDatabaseFullClone creates a database with full data clone
func (m *Manager) createDatabaseFullClone(ctx context.Context, branch *Branch, parentBranchID *uuid.UUID) error {
	dbName := sanitizeIdentifier(branch.DatabaseName)

	templateDB, err := m.resolveTemplateDatabase(ctx, branch, parentBranchID)
	if err != nil {
		return err
	}

	createFromTemplate := fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s",
		dbName, sanitizeIdentifier(templateDB))

	_, err = m.adminPool.Exec(ctx, createFromTemplate)
	if err != nil {
		if strings.Contains(err.Error(), "being accessed by other users") {
			return fmt.Errorf("cannot clone database: parent database has active connections. Try schema_only mode instead: %w", err)
		}
		return fmt.Errorf("failed to create database from template: %w", err)
	}

	if err := m.repairFDW(ctx, branch); err != nil {
		log.Warn().Err(err).Str("branch", branch.Slug).
			Msg("Failed to repair FDW after cloning, branch may have stale FDW mappings")
	}

	return nil
}

// createDatabaseSeedData creates a database with schema and seed data
func (m *Manager) createDatabaseSeedData(ctx context.Context, branch *Branch, parentBranchID *uuid.UUID) error {
	// Step 1: Create database with schema only (reuse existing logic)
	if err := m.createDatabaseSchemaOnly(ctx, branch, parentBranchID); err != nil {
		return err
	}

	// Step 2: Get connection pool for the new branch database
	branchPool, err := m.getBranchConnectionPool(ctx, branch)
	if err != nil {
		return fmt.Errorf("failed to get branch connection pool: %w", err)
	}
	defer branchPool.Close()

	// Step 3: Determine seeds path (branch-specific or global default)
	seedsPath := m.config.SeedsPath
	if branch.SeedsPath != nil && *branch.SeedsPath != "" {
		seedsPath = *branch.SeedsPath
	}

	// Step 4: Initialize seeder
	seeder := NewSeeder(seedsPath)

	// Step 5: Execute seed files
	log.Info().
		Str("branch_id", branch.ID.String()).
		Str("database", branch.DatabaseName).
		Str("seeds_path", seedsPath).
		Msg("Executing seed data")

	// Log activity: seeding started
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branch.ID,
		Action:     ActivityActionSeeding,
		Status:     ActivityStatusStarted,
		ExecutedBy: branch.CreatedBy,
	})

	startTime := time.Now()
	if err := seeder.ExecuteSeeds(ctx, branchPool, branch.ID); err != nil {
		// Log failure
		errMsg := err.Error()
		durationMs := int(time.Since(startTime).Milliseconds())
		_ = m.storage.LogActivity(ctx, &ActivityLog{
			BranchID:     branch.ID,
			Action:       ActivityActionSeeding,
			Status:       ActivityStatusFailed,
			ErrorMessage: &errMsg,
			ExecutedBy:   branch.CreatedBy,
			DurationMs:   &durationMs,
		})

		return fmt.Errorf("failed to execute seed data: %w", err)
	}

	// Log success
	durationMs := int(time.Since(startTime).Milliseconds())
	_ = m.storage.LogActivity(ctx, &ActivityLog{
		BranchID:   branch.ID,
		Action:     ActivityActionSeeding,
		Status:     ActivityStatusSuccess,
		ExecutedBy: branch.CreatedBy,
		DurationMs: &durationMs,
	})

	log.Info().
		Str("branch_id", branch.ID.String()).
		Int("duration_ms", durationMs).
		Msg("Seed data executed successfully")

	return nil
}

// getBranchConnectionPool creates a connection pool for a specific branch database
func (m *Manager) getBranchConnectionPool(ctx context.Context, branch *Branch) (*pgxpool.Pool, error) {
	connURL, err := m.GetBranchConnectionURL(branch)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection URL: %w", err)
	}

	// Use minimal connections for seed execution
	poolConfig.MaxConns = 2
	poolConfig.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// repairFDW repairs the Foreign Data Wrapper configuration in a branch database
// after cloning from a tenant database. When a tenant database with FDW is cloned
// via TEMPLATE, the foreign table definitions and server are copied but user
// mappings may be stale. This method recreates the FDW user mapping if an
// FDWRepairer is available and the branch belongs to a tenant.
func (m *Manager) repairFDW(ctx context.Context, branch *Branch) error {
	if branch.TenantID == nil || m.fdwRepairer == nil {
		return nil
	}

	connURL, err := m.GetBranchConnectionURL(branch)
	if err != nil {
		return fmt.Errorf("failed to get branch connection URL for FDW repair: %w", err)
	}

	return m.fdwRepairer.RepairFDWForBranch(ctx, connURL, *branch.TenantID)
}

// dropDatabase drops a database
func (m *Manager) dropDatabase(ctx context.Context, databaseName string) error {
	dbName := sanitizeIdentifier(databaseName)

	// First, terminate all connections to the database
	// Use parameterized query to prevent SQL injection
	terminateQuery := `
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`

	_, _ = m.adminPool.Exec(ctx, terminateQuery, databaseName)

	// Small delay to allow connections to close
	time.Sleep(100 * time.Millisecond)

	// Drop the database
	dropQuery := fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName)
	_, err := m.adminPool.Exec(ctx, dropQuery)
	if err != nil {
		return fmt.Errorf("failed to drop database: %w", err)
	}

	return nil
}

// GetBranchConnectionURL returns the connection URL for a branch database
func (m *Manager) GetBranchConnectionURL(branch *Branch) (string, error) {
	// Parse the main database URL
	parsedURL, err := url.Parse(m.mainDBURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse main database URL: %w", err)
	}

	// Replace the database name
	parsedURL.Path = "/" + branch.DatabaseName

	return parsedURL.String(), nil
}

// CleanupExpiredBranches deletes branches that have passed their expiration time
func (m *Manager) CleanupExpiredBranches(ctx context.Context) error {
	expired, err := m.storage.GetExpiredBranches(ctx)
	if err != nil {
		return fmt.Errorf("failed to get expired branches: %w", err)
	}

	for _, branch := range expired {
		log.Info().
			Str("branch_id", branch.ID.String()).
			Str("slug", branch.Slug).
			Time("expires_at", *branch.ExpiresAt).
			Msg("Deleting expired branch")

		if err := m.DeleteBranch(ctx, branch.ID, nil); err != nil {
			log.Error().Err(err).
				Str("branch_id", branch.ID.String()).
				Str("slug", branch.Slug).
				Msg("Failed to delete expired branch")
			// Continue with other branches
		}
	}

	return nil
}

// Close closes the manager and releases resources
func (m *Manager) Close() {
	if m.adminPool != nil {
		m.adminPool.Close()
	}
}

// GetStorage returns the storage instance
func (m *Manager) GetStorage() *Storage {
	return m.storage
}

// GetConfig returns the branching config
func (m *Manager) GetConfig() config.BranchingConfig {
	return m.config
}

// sanitizeIdentifier sanitizes a SQL identifier to prevent injection
func sanitizeIdentifier(name string) string {
	// Defense in depth: validate the identifier is alphanumeric before quoting.
	// All inputs should already be sanitized by GenerateDatabaseName/GenerateTenantBranchDatabaseName.
	if !isValidIdentifier(name) {
		log.Error().Str("identifier", name).Msg("sanitizeIdentifier: rejected invalid identifier")
		return `""`
	}
	// Use double quotes and escape any existing quotes
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// isValidIdentifier checks if a string is a valid PostgreSQL identifier
func isValidIdentifier(name string) bool {
	if len(name) == 0 || len(name) > 63 {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}

// CreateBranchFromGitHubPR creates a branch for a GitHub PR
func (m *Manager) CreateBranchFromGitHubPR(ctx context.Context, repo string, prNumber int, prURL string) (*Branch, error) {
	// Get GitHub config for the repository
	ghConfig, err := m.storage.GetGitHubConfig(ctx, repo)
	if err != nil && !errors.Is(err, ErrGitHubConfigNotFound) {
		return nil, fmt.Errorf("failed to get GitHub config: %w", err)
	}

	// Determine data clone mode
	dataCloneMode := DataCloneModeSchemaOnly
	if ghConfig != nil && ghConfig.DefaultDataCloneMode != "" {
		dataCloneMode = ghConfig.DefaultDataCloneMode
	}

	// Create branch name and slug from PR number
	name := fmt.Sprintf("PR #%d", prNumber)
	slug := GeneratePRSlug(prNumber)

	// Check if branch already exists (no tenant filter for GitHub PR branches)
	existing, err := m.storage.GetBranchBySlug(ctx, slug, nil)
	if err != nil && !errors.Is(err, ErrBranchNotFound) {
		return nil, fmt.Errorf("failed to check existing branch: %w", err)
	}
	if existing != nil {
		// Branch already exists, return it
		return existing, nil
	}

	req := CreateBranchRequest{
		Name:           name,
		DataCloneMode:  dataCloneMode,
		Type:           BranchTypePreview,
		GitHubPRNumber: &prNumber,
		GitHubPRURL:    &prURL,
		GitHubRepo:     &repo,
	}

	return m.CreateBranch(ctx, req, nil)
}

// DeleteBranchForGitHubPR deletes the branch associated with a GitHub PR
func (m *Manager) DeleteBranchForGitHubPR(ctx context.Context, repo string, prNumber int) error {
	// Find branch by GitHub PR
	branch, err := m.storage.GetBranchByGitHubPR(ctx, repo, prNumber)
	if err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			// Branch doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get branch for PR: %w", err)
	}

	return m.DeleteBranch(ctx, branch.ID, nil)
}

// RunTransaction executes a function in a transaction using the admin pool
func (m *Manager) RunTransaction(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := m.adminPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
