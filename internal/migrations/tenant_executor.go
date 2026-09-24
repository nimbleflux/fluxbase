package migrations

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// tenantMigrationLockID serializes tenant migration SQL across processes. It is
// distinct from migrationLockID (executor.go, imperative main-DB migrations) and
// from schemaApplyLockID (lock.go, declarative schema applies).
const tenantMigrationLockID int64 = 0x466C7578_00000006

// TenantExecutor applies user-authored migrations to a single tenant's database.
//
// Migration DEFINITIONS always live in the main database's platform.migrations
// (created/synced via the migrations API). Applied-STATE tracking depends on the
// tenant's database mode:
//
//   - Separate-DB tenants (tenant.DBName set): applied state is tracked in the
//     tenant database's own platform.migrations table (tenant DBs receive the
//     platform schema via tenantdb.Manager.applyInternalSchemas; the table is
//     also ensured idempotently at apply time). One state row per physical
//     tenant DB, so tenant A applying migration X no longer marks it applied
//     for tenant B, and a rollback by one tenant does not block the others.
//   - Shared-DB tenants (UsesMainDatabase): state stays in the main database's
//     platform.migrations status row — correct there, since one physical DB is
//     shared by all such tenants.
//
// Note on atomicity: the migration SQL runs in a transaction under
// tenant_migration_role, which cannot safely write the state row in the same
// transaction. The status write is therefore best-effort: if it fails after the
// SQL committed, a warning is logged and the next apply may re-execute the
// migration — tenant migrations must be idempotent.
type TenantExecutor struct {
	mu       sync.Mutex
	storage  *Storage
	db       *database.Connection
	tenantDB *pgxpool.Pool
}

func NewTenantExecutor(db *database.Connection, tenantDB *pgxpool.Pool) *TenantExecutor {
	return &TenantExecutor{
		storage:  NewStorage(db),
		db:       db,
		tenantDB: tenantDB,
	}
}

// tenantUsesSeparateDatabase reports whether the tenant pool points at a
// different physical database than the main connection. When detection fails,
// the safe default (false) keeps the previous shared main-DB state behavior.
func (e *TenantExecutor) tenantUsesSeparateDatabase(ctx context.Context) bool {
	if e.tenantDB == nil || e.db == nil || e.db.Pool() == nil {
		return false
	}

	var tenantDBName, mainDBName string
	if err := e.tenantDB.QueryRow(ctx, "SELECT current_database()").Scan(&tenantDBName); err != nil {
		log.Warn().Err(err).Msg("Failed to detect tenant database name for migration state")
		return false
	}
	if err := e.db.Pool().QueryRow(ctx, "SELECT current_database()").Scan(&mainDBName); err != nil {
		log.Warn().Err(err).Msg("Failed to detect main database name for migration state")
		return false
	}
	return tenantDBName != mainDBName
}

// ensureTenantStateTable creates platform.migrations in the tenant database if
// absent (matching the embedded platform.sql definition, minus the users FKs).
func (e *TenantExecutor) ensureTenantStateTable(ctx context.Context) error {
	_, err := e.tenantDB.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS platform.migrations (
			id uuid DEFAULT gen_random_uuid(),
			namespace text DEFAULT 'default' NOT NULL,
			name text NOT NULL,
			description text,
			up_sql text NOT NULL,
			down_sql text,
			version integer DEFAULT 1,
			status text DEFAULT 'pending',
			created_by uuid,
			applied_by uuid,
			created_at timestamptz DEFAULT now() NOT NULL,
			updated_at timestamptz DEFAULT now() NOT NULL,
			applied_at timestamptz,
			rolled_back_at timestamptz,
			CONSTRAINT migrations_pkey PRIMARY KEY (id),
			CONSTRAINT unique_migration_namespace UNIQUE (namespace, name),
			CONSTRAINT valid_status CHECK (status IN ('pending'::text, 'applied'::text, 'failed'::text, 'rolled_back'::text))
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to ensure tenant migration state table: %w", err)
	}
	return nil
}

// tenantMigrationStatus returns the tenant-local applied status of a migration.
// Empty string means the migration has never been applied in this tenant DB.
func (e *TenantExecutor) tenantMigrationStatus(ctx context.Context, namespace, name string) (string, error) {
	var status string
	err := e.tenantDB.QueryRow(ctx,
		`SELECT status FROM platform.migrations WHERE namespace = $1 AND name = $2`,
		namespace, name,
	).Scan(&status)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to read tenant migration state: %w", err)
	}
	return status, nil
}

// setTenantMigrationStatus upserts the tenant-local applied state. The
// definition columns are copied from def so the tenant-DB row is self-describing.
func (e *TenantExecutor) setTenantMigrationStatus(ctx context.Context, def *Migration, status string, appliedBy *uuid.UUID) error {
	query := `
		INSERT INTO platform.migrations
			(namespace, name, up_sql, down_sql, version, status, applied_by, applied_at, rolled_back_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6::text, $7,
			CASE WHEN $6::text = 'applied' THEN now() END,
			CASE WHEN $6::text = 'rolled_back' THEN now() END,
			now())
		ON CONFLICT (namespace, name) DO UPDATE SET
			status         = EXCLUDED.status,
			applied_by     = EXCLUDED.applied_by,
			applied_at     = COALESCE(EXCLUDED.applied_at, platform.migrations.applied_at),
			rolled_back_at = COALESCE(EXCLUDED.rolled_back_at, platform.migrations.rolled_back_at),
			updated_at     = now()
	`
	_, err := e.tenantDB.Exec(ctx, query, def.Namespace, def.Name, def.UpSQL, def.DownSQL, def.Version, status, appliedBy)
	if err != nil {
		return fmt.Errorf("failed to update tenant migration state: %w", err)
	}
	return nil
}

// migrationState tells ApplyMigration/RollbackMigration where to read and record
// applied state for this apply.
type migrationState struct {
	separate bool // tenant DB tracks its own state
	status   string
}

// resolveState determines the current applied status for (namespace, name),
// using tenant-local state for separate-DB tenants and the shared main-DB row
// otherwise. If the tenant-local state table cannot be ensured, it falls back to
// the main-DB row with a warning (previous behavior).
func (e *TenantExecutor) resolveState(ctx context.Context, tenantID, namespace, name string, def *Migration) (migrationState, error) {
	if !e.tenantUsesSeparateDatabase(ctx) {
		return migrationState{separate: false, status: def.Status}, nil
	}

	if err := e.ensureTenantStateTable(ctx); err != nil {
		log.Warn().
			Err(err).
			Str("tenant_id", tenantID).
			Msg("Cannot ensure tenant-local migration state table; falling back to main-DB migration state (state will be shared across tenants, not per tenant DB)")
		return migrationState{separate: false, status: def.Status}, nil
	}

	status, err := e.tenantMigrationStatus(ctx, namespace, name)
	if err != nil {
		return migrationState{}, err
	}
	return migrationState{separate: true, status: status}, nil
}

// recordState persists the applied status: tenant-local for separate-DB tenants,
// the shared main-DB row otherwise. Both paths are best-effort relative to the
// committed migration SQL (see the TenantExecutor doc comment).
func (e *TenantExecutor) recordState(ctx context.Context, st migrationState, def *Migration, status string, executedBy *uuid.UUID) error {
	if st.separate {
		if err := e.setTenantMigrationStatus(ctx, def, status, executedBy); err != nil {
			return err
		}
		return nil
	}
	return e.storage.UpdateMigrationStatus(ctx, def.ID, status, executedBy)
}

// warnStateRecordFailed logs the best-effort status-write failure. The migration
// SQL already committed, so failing the request would not undo it; the next
// apply may re-execute the migration, which must be idempotent.
func warnStateRecordFailed(err error, tenantID, namespace, name, action string) {
	log.Warn().
		Err(err).
		Str("tenant_id", tenantID).
		Str("namespace", namespace).
		Str("name", name).
		Str("action", action).
		Msg("Failed to record migration state after the SQL committed; the next apply may re-execute this migration (migrations must be idempotent)")
}

func (e *TenantExecutor) ApplyMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	migration, err := e.storage.GetMigration(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	st, err := e.resolveState(ctx, tenantID, namespace, name, migration)
	if err != nil {
		return err
	}

	if st.status == "applied" {
		log.Info().
			Str("namespace", namespace).
			Str("name", name).
			Str("tenant_id", tenantID).
			Msg("Migration already applied, skipping")
		return nil
	}

	if st.status != "" && st.status != "pending" && st.status != "failed" {
		return fmt.Errorf("migration status is %s, cannot apply", st.status)
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Str("tenant_id", tenantID).
		Msg("Applying tenant migration")

	startTime := time.Now()

	err = e.executeWithTenantRole(ctx, migration.UpSQL)

	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		log.Error().
			Err(err).
			Str("namespace", namespace).
			Str("name", name).
			Str("tenant_id", tenantID).
			Int("duration_ms", durationMs).
			Msg("Tenant migration failed")

		errMsg := err.Error()
		executionLog := &ExecutionLog{
			MigrationID:  migration.ID,
			Action:       "apply",
			Status:       "failed",
			DurationMs:   &durationMs,
			ErrorMessage: &errMsg,
			ExecutedBy:   executedBy,
		}

		_ = e.storage.LogExecution(ctx, executionLog)
		if err := e.recordState(ctx, st, migration, "failed", executedBy); err != nil {
			// The SQL rolled back with the transaction; marking the state failed
			// is best-effort bookkeeping only.
			warnStateRecordFailed(err, tenantID, namespace, name, "apply")
		}

		return fmt.Errorf("migration failed: %w", err)
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Str("tenant_id", tenantID).
		Int("duration_ms", durationMs).
		Msg("Tenant migration applied successfully")

	executionLog := &ExecutionLog{
		MigrationID: migration.ID,
		Action:      "apply",
		Status:      "success",
		DurationMs:  &durationMs,
		ExecutedBy:  executedBy,
	}

	if err := e.storage.LogExecution(ctx, executionLog); err != nil {
		log.Warn().Err(err).Msg("Failed to log migration execution")
	}

	if err := e.recordState(ctx, st, migration, "applied", executedBy); err != nil {
		warnStateRecordFailed(err, tenantID, namespace, name, "apply")
	}

	return nil
}

func (e *TenantExecutor) RollbackMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	migration, err := e.storage.GetMigration(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	st, err := e.resolveState(ctx, tenantID, namespace, name, migration)
	if err != nil {
		return err
	}

	if st.status != "applied" {
		return fmt.Errorf("migration status is %s, cannot rollback", st.status)
	}

	if migration.DownSQL == nil || *migration.DownSQL == "" {
		return fmt.Errorf("migration has no rollback SQL")
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Str("tenant_id", tenantID).
		Msg("Rolling back tenant migration")

	startTime := time.Now()

	err = e.executeWithTenantRole(ctx, *migration.DownSQL)

	durationMs := int(time.Since(startTime).Milliseconds())

	if err != nil {
		log.Error().
			Err(err).
			Str("namespace", namespace).
			Str("name", name).
			Str("tenant_id", tenantID).
			Int("duration_ms", durationMs).
			Msg("Tenant rollback failed")

		errMsg := err.Error()
		executionLog := &ExecutionLog{
			MigrationID:  migration.ID,
			Action:       "rollback",
			Status:       "failed",
			DurationMs:   &durationMs,
			ErrorMessage: &errMsg,
			ExecutedBy:   executedBy,
		}

		_ = e.storage.LogExecution(ctx, executionLog)

		return fmt.Errorf("rollback failed: %w", err)
	}

	log.Info().
		Str("namespace", namespace).
		Str("name", name).
		Str("tenant_id", tenantID).
		Int("duration_ms", durationMs).
		Msg("Tenant migration rolled back successfully")

	executionLog := &ExecutionLog{
		MigrationID: migration.ID,
		Action:      "rollback",
		Status:      "success",
		DurationMs:  &durationMs,
		ExecutedBy:  executedBy,
	}

	if err := e.storage.LogExecution(ctx, executionLog); err != nil {
		log.Warn().Err(err).Msg("Failed to log migration execution")
	}

	if err := e.recordState(ctx, st, migration, "rolled_back", executedBy); err != nil {
		warnStateRecordFailed(err, tenantID, namespace, name, "rollback")
	}

	return nil
}

func (e *TenantExecutor) executeWithTenantRole(ctx context.Context, sql string) error {
	if e.tenantDB == nil {
		return fmt.Errorf("tenant database pool not configured")
	}

	conn, err := e.tenantDB.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, "SET LOCAL ROLE tenant_migration_role")
	if err != nil {
		return fmt.Errorf("failed to set tenant_migration_role: %w", err)
	}

	// Serialize tenant migration DDL across processes for this database.
	// A dedicated lock ID (distinct from the imperative executor's) keeps tenant
	// migrations from blocking main-DB migrations unnecessarily.
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock($1)", tenantMigrationLockID); err != nil {
		return fmt.Errorf("failed to acquire tenant migration advisory lock: %w", err)
	}

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("SQL execution failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
