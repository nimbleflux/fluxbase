package migrations

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// schemaApplyLockID guards declarative schema applies (startup
// ApplyDeclarativeWithSource, HTTP internal-schema applies via ApplyFiltered, and
// app-schema applies) across concurrent processes and instances. It is distinct
// from migrationLockID (executor.go), which serializes imperative migrations, and
// from tenantMigrationLockID (tenant_executor.go).
const schemaApplyLockID int64 = 0x466C7578_00000005

// defaultSchemaApplyLockTimeoutSecs bounds how long a schema apply waits for the
// advisory lock when the caller does not provide a configured timeout.
const defaultSchemaApplyLockTimeoutSecs = 30

// WithSchemaApplyLock runs fn while holding a session-level PostgreSQL advisory
// lock that serializes declarative schema applies across processes. The lock is
// taken on a dedicated connection acquired from pool and always released when fn
// returns. lockTimeoutSecs bounds how long acquisition may block (<=0 waits
// indefinitely, bounded by ctx); the connection's lock_timeout is reset on release.
// If pool is nil (service constructed without a pool) fn runs without locking.
func WithSchemaApplyLock(ctx context.Context, pool *pgxpool.Pool, lockTimeoutSecs int, fn func() error) error {
	if pool == nil {
		return fn()
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection for schema apply lock: %w", err)
	}
	defer conn.Release()

	if lockTimeoutSecs > 0 {
		if _, err := conn.Exec(ctx, fmt.Sprintf("SET lock_timeout TO '%ds'", lockTimeoutSecs)); err != nil {
			log.Warn().Err(err).Msg("Failed to set lock_timeout for schema apply lock")
		} else {
			// SET persists for the session; reset so the pooled connection does not
			// leak the timeout to other work.
			defer func() {
				_, _ = conn.Exec(context.Background(), "RESET lock_timeout")
			}()
		}
	}

	// pg_advisory_lock blocks until the lock is free (bounded by lock_timeout set
	// above and by ctx) and returns void, so Exec rather than QueryRow.
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", schemaApplyLockID); err != nil {
		return fmt.Errorf("failed to acquire schema apply advisory lock: %w", err)
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		var released bool
		if err := conn.QueryRow(releaseCtx, "SELECT pg_advisory_unlock($1)", schemaApplyLockID).Scan(&released); err != nil || !released {
			log.Warn().Err(err).Bool("released", released).Msg("Failed to release schema apply advisory lock")
		}
	}()

	return fn()
}
