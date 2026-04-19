package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Validator provides schema validation utilities
type Validator struct {
	service *DeclarativeService
	pool    *pgxpool.Pool
}

// NewValidator creates a new schema validator
func NewValidator(service *DeclarativeService, pool *pgxpool.Pool) *Validator {
	return &Validator{
		service: service,
		pool:    pool,
	}
}

// ValidateSchema checks if the database matches the declared schema
func (v *Validator) ValidateSchema(ctx context.Context) (*ValidationResult, error) {
	return v.service.Validate(ctx)
}

// DetectMigrationState determines the current migration system state
func (v *Validator) DetectMigrationState(ctx context.Context) (*MigrationState, error) {
	state := &MigrationState{}

	// Check for imperative migrations (platform.fluxbase_migrations table)
	var imperativeExists bool
	err := v.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'platform'
			AND table_name = 'fluxbase_migrations'
		)
	`).Scan(&imperativeExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check imperative migrations: %w", err)
	}
	state.HasImperativeMigrations = imperativeExists

	// If imperative migrations exist, get the last version and check for dirty
	if imperativeExists {
		err := v.pool.QueryRow(ctx, `
			SELECT COALESCE(MAX(version), 0) FROM platform.fluxbase_migrations WHERE NOT dirty
		`).Scan(&state.LastAppliedVersion)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to get last migration version")
		}

		// Check for dirty migrations (for informational logging only, not blocking)
		var dirtyCount int
		err = v.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM platform.fluxbase_migrations WHERE dirty
		`).Scan(&dirtyCount)
		if err == nil {
			state.HasDirtyMigrations = dirtyCount > 0
		}
	}

	// Check for declarative state (platform.declarative_state table)
	var declarativeExists bool
	err = v.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'platform'
			AND table_name = 'declarative_state'
		)
	`).Scan(&declarativeExists)
	if err != nil {
		return nil, fmt.Errorf("failed to check declarative state: %w", err)
	}
	state.HasDeclarativeState = declarativeExists

	// If declarative state exists, get the fingerprint
	if declarativeExists {
		err := v.pool.QueryRow(ctx, `
			SELECT schema_fingerprint
			FROM platform.declarative_state
			ORDER BY applied_at DESC
			LIMIT 1
		`).Scan(&state.SchemaFingerprint)
		if err != nil && err != pgx.ErrNoRows {
			log.Warn().Err(err).Msg("Failed to get schema fingerprint")
		}
	}

	return state, nil
}

// ValidateForTransition checks if the database is ready for transition to declarative
func (v *Validator) ValidateForTransition(ctx context.Context) error {
	state, err := v.DetectMigrationState(ctx)
	if err != nil {
		return err
	}

	// Must have imperative migrations to transition
	if !state.HasImperativeMigrations {
		return fmt.Errorf("no imperative migrations found - use declarative mode directly for fresh installs")
	}

	// Check if there are dirty migrations
	var dirtyCount int
	err = v.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM platform.fluxbase_migrations WHERE dirty
	`).Scan(&dirtyCount)
	if err != nil {
		return fmt.Errorf("failed to check dirty migrations: %w", err)
	}
	if dirtyCount > 0 {
		return fmt.Errorf("found %d dirty migrations - resolve them before transitioning", dirtyCount)
	}

	log.Info().
		Int64("last_version", state.LastAppliedVersion).
		Msg("Migration state validated for transition")

	return nil
}

// CheckSchemaIntegrity performs integrity checks on the schema
func (v *Validator) CheckSchemaIntegrity(ctx context.Context) ([]string, error) {
	var issues []string

	// Check for invalid foreign keys
	rows, err := v.pool.Query(ctx, `
		SELECT tc.table_schema, tc.table_name, tc.constraint_name
		FROM information_schema.table_constraints tc
		WHERE tc.constraint_type = 'FOREIGN KEY'
		AND tc.table_schema NOT IN ('information_schema', 'pg_catalog')
		AND NOT EXISTS (
			SELECT 1 FROM pg_constraint pc
			WHERE pc.contype = 'f'
			AND pc.convalidated = true
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check foreign keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, table, constraint string
		if err := rows.Scan(&schema, &table, &constraint); err != nil {
			continue
		}
		issues = append(issues, fmt.Sprintf("Unvalidated FK: %s.%s (%s)", schema, table, constraint))
	}

	// Check for tables without primary keys
	rows, err = v.pool.Query(ctx, `
		SELECT t.table_schema, t.table_name
		FROM information_schema.tables t
		WHERE t.table_schema NOT IN ('information_schema', 'pg_catalog', 'public')
		AND t.table_type = 'BASE TABLE'
		AND NOT EXISTS (
			SELECT 1 FROM information_schema.table_constraints tc
			WHERE tc.table_schema = t.table_schema
			AND tc.table_name = t.table_name
			AND tc.constraint_type = 'PRIMARY KEY'
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to check primary keys: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var schema, table string
		if err := rows.Scan(&schema, &table); err != nil {
			continue
		}
		issues = append(issues, fmt.Sprintf("No primary key: %s.%s", schema, table))
	}

	return issues, nil
}
