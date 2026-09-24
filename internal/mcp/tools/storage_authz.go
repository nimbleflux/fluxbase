package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// This file adapts the REST storage API's object-level authorization probes
// (see internal/api/storage_authz.go) to MCP tool execution. Every probe runs
// inside a short transaction with the CALLER's RLS context applied, so
// PostgreSQL row level security decides whether the caller can see or write
// the object row. Read probes must run before any bytes are read from the
// provider; write probes before any bytes are written.

// isRLSDenial reports whether an error looks like a PostgreSQL
// privilege/RLS denial rather than an infrastructure failure.
func isRLSDenial(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "42501" {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "row-level security")
}

// skipStorageProbe reports whether the caller is exempt from object probes
// (service role callers bypass RLS everywhere else as well).
func skipStorageProbe(authCtx *mcp.AuthContext) bool {
	return authCtx != nil && authCtx.IsServiceRole
}

// probeBucketVisible verifies under the caller's RLS context that the bucket
// row is visible. Service-role callers skip the probe.
func probeBucketVisible(ctx context.Context, db *database.Connection, authCtx *mcp.AuthContext, bucket string) error {
	if db == nil || skipStorageProbe(authCtx) {
		return nil
	}

	err := executeWithRLS(ctx, db, authCtx, func(tx pgx.Tx) error {
		var visible bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM storage.buckets WHERE id = $1)
		`, bucket).Scan(&visible); err != nil {
			return err
		}
		if !visible {
			return fmt.Errorf("bucket '%s' not found or not accessible", bucket)
		}
		return nil
	})
	if err != nil && isRLSDenial(err) {
		return fmt.Errorf("bucket '%s' not found or not accessible", bucket)
	}
	return err
}

// probeObjectRead verifies under the caller's RLS context that the object row
// is visible (the same visibility the REST download path enforces).
func probeObjectRead(ctx context.Context, db *database.Connection, authCtx *mcp.AuthContext, bucket, key string) error {
	if db == nil || skipStorageProbe(authCtx) {
		return nil
	}

	err := executeWithRLS(ctx, db, authCtx, func(tx pgx.Tx) error {
		var visible bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)
		`, bucket, key).Scan(&visible); err != nil {
			return err
		}
		if !visible {
			return fmt.Errorf("object not found or not accessible: %s/%s", bucket, key)
		}
		return nil
	})
	if err != nil && isRLSDenial(err) {
		return fmt.Errorf("object not found or not accessible: %s/%s", bucket, key)
	}
	return err
}

// probeObjectWrite verifies under the caller's RLS context that the caller may
// create or overwrite the object at (bucket, key), without leaving any state
// behind:
//   - if the object row is visible, a no-op UPDATE verifies overwrite
//     permission (rowsAffected == 0 means denied);
//   - otherwise a placeholder row is inserted AND deleted in one transaction,
//     which validates the RLS insert policy without reserving the path.
func probeObjectWrite(ctx context.Context, db *database.Connection, authCtx *mcp.AuthContext, bucket, key string) error {
	if db == nil || skipStorageProbe(authCtx) {
		return nil
	}

	denied := fmt.Errorf("insufficient permissions to write object: %s/%s", bucket, key)

	err := executeWithRLS(ctx, db, authCtx, func(tx pgx.Tx) error {
		var visible bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)
		`, bucket, key).Scan(&visible); err != nil {
			return err
		}

		if visible {
			result, err := tx.Exec(ctx, `
				UPDATE storage.objects SET updated_at = updated_at
				WHERE bucket_id = $1 AND path = $2
			`, bucket, key)
			if err != nil {
				return err
			}
			if result.RowsAffected() == 0 {
				return denied
			}
			return nil
		}

		// New object: validate the insert policy without keeping the row.
		if _, err := tx.Exec(ctx, `
			INSERT INTO storage.objects (bucket_id, path, mime_type, size, metadata, owner_id)
			VALUES ($1, $2, 'application/octet-stream', 0, NULL, NULL)
		`, bucket, key); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `
			DELETE FROM storage.objects WHERE bucket_id = $1 AND path = $2
		`, bucket, key); err != nil {
			return err
		}
		return nil
	})
	if err == nil {
		return nil
	}
	if err.Error() == denied.Error() {
		return denied
	}
	if isRLSDenial(err) {
		return denied
	}
	return err
}
