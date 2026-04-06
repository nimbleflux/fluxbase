package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
)

type TenantExecutor struct {
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

func (e *TenantExecutor) ApplyMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	migration, err := e.storage.GetMigration(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	if migration.Status == "applied" {
		log.Info().
			Str("namespace", namespace).
			Str("name", name).
			Str("tenant_id", tenantID).
			Msg("Migration already applied, skipping")
		return nil
	}

	if migration.Status != "pending" && migration.Status != "failed" {
		return fmt.Errorf("migration status is %s, cannot apply", migration.Status)
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
		_ = e.storage.UpdateMigrationStatus(ctx, migration.ID, "failed", executedBy)

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

	if err := e.storage.UpdateMigrationStatus(ctx, migration.ID, "applied", executedBy); err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
	}

	return nil
}

func (e *TenantExecutor) RollbackMigration(ctx context.Context, namespace, name, tenantID string, executedBy *uuid.UUID) error {
	migration, err := e.storage.GetMigration(ctx, namespace, name)
	if err != nil {
		return fmt.Errorf("failed to get migration: %w", err)
	}

	if migration.Status != "applied" {
		return fmt.Errorf("migration status is %s, cannot rollback", migration.Status)
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

	if err := e.storage.UpdateMigrationStatus(ctx, migration.ID, "rolled_back", executedBy); err != nil {
		return fmt.Errorf("failed to update migration status: %w", err)
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

	_, err = tx.Exec(ctx, sql)
	if err != nil {
		return fmt.Errorf("SQL execution failed: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
