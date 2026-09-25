package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// This file centralizes object-level authorization probes for the storage
// handlers. Every probe runs inside a short transaction with the request's
// RLS context applied (see setRLSContext), so PostgreSQL row level security
// decides whether the caller can see or write the object row. Probes must run
// BEFORE any bytes are written to the storage provider so a denied caller
// never destroys or replaces an existing object.

// authzProbeError carries an HTTP status for a failed authorization probe.
type authzProbeError struct {
	status  int
	message string
	code    string
	err     error // underlying error, when one exists
}

func (e *authzProbeError) Error() string { return e.message }

func (e *authzProbeError) send(c fiber.Ctx) error {
	return SendErrorWithCode(c, e.status, e.message, e.code)
}

func internalProbeError(err error) *authzProbeError {
	log.Error().Err(err).Msg("Storage authorization probe failed")
	return &authzProbeError{status: fiber.StatusInternalServerError, message: "failed to check object permissions", code: ErrCodeInvalidInput}
}

// probePool returns the database pool for probe transactions, mapping an
// uninitialized database to an internal probe error instead of a panic.
func (h *StorageHandler) probePool(c fiber.Ctx) (*pgxpool.Pool, *authzProbeError) {
	if h.db == nil && middleware.GetTenantPool(c) == nil {
		return nil, &authzProbeError{status: fiber.StatusInternalServerError, message: "database not initialized", code: ErrCodeInvalidInput}
	}
	return h.getPool(c), nil
}

// uploadProbeResult reports whether the probe created a placeholder row in
// storage.objects. When it did, the caller owns that row and MUST remove it
// (and any bytes written) if the upload subsequently fails. When it did not
// (overwrite of an existing, caller-writable row), the provider bytes must
// never be deleted on failure — the row belongs to a pre-existing object the
// caller was merely authorized to overwrite.
type uploadProbeResult struct {
	createdPlaceholder bool
}

// probeWritePermission verifies under RLS that the caller may write an
// existing object row via a no-op UPDATE. RLS applies the USING policy on the
// row scan and the WITH CHECK policy on the update, so rowsAffected == 0 or a
// permission error means the caller cannot overwrite this object.
func (h *StorageHandler) probeWritePermission(ctx context.Context, c fiber.Ctx, bucket, key string) *authzProbeError {
	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return internalProbeError(err)
	}

	result, err := tx.Exec(ctx, `
		UPDATE storage.objects SET updated_at = updated_at
		WHERE bucket_id = $1 AND path = $2
	`, bucket, key)
	if err != nil {
		if isRLSDenial(err) {
			return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to write object", code: ErrCodeAccessDenied}
		}
		return internalProbeError(err)
	}
	if result.RowsAffected() == 0 {
		return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to write object", code: ErrCodeAccessDenied}
	}

	if err := tx.Commit(ctx); err != nil {
		return internalProbeError(err)
	}
	return nil
}

// AuthorizeObjectRead verifies under RLS that the caller can see the object
// row (the same visibility check DownloadFile applies). Returns nil when
// access is allowed.
func (h *StorageHandler) AuthorizeObjectRead(c fiber.Ctx, bucket, key string) *authzProbeError {
	ctx := c.RequestCtx()

	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Resolve tenant context from the object itself (mirrors DownloadFile) so
	// tenant-scoped objects remain readable for public access patterns.
	h.resolveTenantForObject(ctx, tx, c, bucket, key)

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return internalProbeError(err)
	}

	var visible bool
	err = tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)
	`, bucket, key).Scan(&visible)
	if err != nil {
		return internalProbeError(err)
	}
	if !visible {
		return &authzProbeError{status: fiber.StatusNotFound, message: "object not found or insufficient permissions", code: ErrCodeNotFound}
	}

	if err := tx.Commit(ctx); err != nil {
		return internalProbeError(err)
	}
	return nil
}

// AuthorizeUploadWrite validates — before any bytes are written to the
// provider — that the caller may create or overwrite the object at
// (bucket, key):
//
//   - If a row already exists, a no-op UPDATE under RLS verifies overwrite
//     permission. No state changes.
//   - If no row exists, a placeholder row (size=0) is inserted under RLS and
//     committed, validating insert permission and reserving the path. The
//     caller receives createdPlaceholder=true and must clean the row up via
//     CleanupUploadPlaceholder if the upload fails.
//   - A concurrent insert (unique violation) falls back to the overwrite
//     probe, matching the ON CONFLICT semantics of the final metadata upsert.
func (h *StorageHandler) AuthorizeUploadWrite(c fiber.Ctx, bucket, key, contentType string, metadata map[string]interface{}, ownerID *string) (*uploadProbeResult, *authzProbeError) {
	ctx := c.RequestCtx()

	exists, perr := h.objectRowVisible(ctx, c, bucket, key)
	if perr != nil {
		return nil, perr
	}
	if exists {
		if perr := h.probeWritePermission(ctx, c, bucket, key); perr != nil {
			return nil, perr
		}
		return &uploadProbeResult{createdPlaceholder: false}, nil
	}

	// No visible row: validate insert permission by creating a placeholder.
	perr = h.insertPlaceholderRow(ctx, c, bucket, key, contentType, metadata, ownerID)
	if perr == nil {
		return &uploadProbeResult{createdPlaceholder: true}, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(perr.err, &pgErr) {
		switch pgErr.Code {
		case pgUniqueViolationCode:
			// Lost a race: another writer created the row between our SELECT
			// and INSERT. Treat it as an overwrite and probe update rights.
			if perr := h.probeWritePermission(ctx, c, bucket, key); perr != nil {
				return nil, perr
			}
			return &uploadProbeResult{createdPlaceholder: false}, nil
		case pgForeignKeyViolationCode:
			// Placeholder insert violates objects -> buckets FK: bucket absent.
			return nil, &authzProbeError{status: fiber.StatusNotFound, message: "bucket does not exist", code: ErrCodeNotFound}
		}
	}
	return nil, perr
}

// insertPlaceholderRow inserts a size=0 placeholder object row under RLS.
func (h *StorageHandler) insertPlaceholderRow(ctx context.Context, c fiber.Ctx, bucket, key, contentType string, metadata map[string]interface{}, ownerID *string) *authzProbeError {
	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return internalProbeError(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO storage.objects (bucket_id, path, mime_type, size, metadata, owner_id)
		VALUES ($1, $2, $3, 0, $4, $5)
	`, bucket, key, contentType, metadata, ownerID); err != nil {
		if isRLSDenial(err) {
			return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to upload object", code: ErrCodeAccessDenied, err: err}
		}
		return &authzProbeError{status: fiber.StatusInternalServerError, message: "failed to check upload permissions", code: ErrCodeInvalidInput, err: err}
	}

	if err := tx.Commit(ctx); err != nil {
		return internalProbeError(err)
	}
	return nil
}

// objectRowVisible reports whether the caller can see the object row under RLS.
func (h *StorageHandler) objectRowVisible(ctx context.Context, c fiber.Ctx, bucket, key string) (bool, *authzProbeError) {
	pool, perr := h.probePool(c)
	if perr != nil {
		return false, perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return false, internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return false, internalProbeError(err)
	}

	var visible bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)
	`, bucket, key).Scan(&visible); err != nil {
		return false, internalProbeError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, internalProbeError(err)
	}
	return visible, nil
}

// CleanupUploadPlaceholder removes a placeholder row created by
// AuthorizeUploadWrite plus any bytes written for it. Best effort: failures
// are logged so they surface in monitoring but do not mask the original error.
func (h *StorageHandler) CleanupUploadPlaceholder(c fiber.Ctx, svc cleanupProvider, bucket, key string) {
	ctx := c.RequestCtx()

	pool, perr := h.probePool(c)
	if perr != nil {
		log.Error().Str("bucket", bucket).Str("key", key).Msg("Failed to get pool to clean up upload placeholder")
	} else if tx, err := pool.Begin(ctx); err != nil {
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to start transaction to clean up upload placeholder")
	} else {
		defer func() { _ = tx.Rollback(ctx) }()
		if err := h.setRLSContext(ctx, tx, c); err != nil {
			log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to set RLS context to clean up upload placeholder")
		} else if _, err := tx.Exec(ctx, `DELETE FROM storage.objects WHERE bucket_id = $1 AND path = $2`, bucket, key); err != nil {
			log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to delete upload placeholder row")
		} else if err := tx.Commit(ctx); err != nil {
			log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to commit upload placeholder cleanup")
		}
	}

	if svc != nil {
		if err := svc.Delete(ctx, bucket, key); err != nil {
			log.Warn().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to delete placeholder bytes from provider")
		}
	}
}

// cleanupProvider is the minimal provider surface needed for placeholder
// cleanup (satisfied by *storage.Service).
type cleanupProvider interface {
	Delete(ctx context.Context, bucket, key string) error
}

// isRLSDenial reports whether the error looks like a PostgreSQL RLS/permission
// denial rather than an infrastructure failure.
func isRLSDenial(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "permission denied") || strings.Contains(msg, "policy")
}

// PostgreSQL error codes used by the upload probes.
const (
	pgUniqueViolationCode     = "23505"
	pgForeignKeyViolationCode = "23503"
)

// validateBucketForUpload checks that the bucket exists and that the declared
// upload size and content type respect the bucket's max_file_size and
// allowed_mime_types settings. Bucket existence/settings are read through
// SECURITY DEFINER helpers that bypass RLS (same pattern as UploadFile).
func (h *StorageHandler) validateBucketForUpload(c fiber.Ctx, bucket string, size int64, contentType string) *authzProbeError {
	ctx := c.RequestCtx()

	var bucketExists bool
	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	err := pool.QueryRow(
		ctx,
		`SELECT storage.bucket_exists($1::text, $2::uuid)`,
		bucket, getTenantIDArg(c),
	).Scan(&bucketExists)
	if err != nil {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to check bucket existence")
		return internalProbeError(err)
	}
	if !bucketExists {
		return &authzProbeError{status: fiber.StatusNotFound, message: "bucket '" + bucket + "' does not exist", code: ErrCodeNotFound}
	}

	var bucketMaxFileSize *int64
	var bucketAllowedMimeTypes []string
	err = pool.QueryRow(
		ctx,
		`SELECT max_file_size, allowed_mime_types FROM storage.get_bucket_settings($1::text, $2::uuid)`,
		bucket, getTenantIDArg(c),
	).Scan(&bucketMaxFileSize, &bucketAllowedMimeTypes)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to get bucket settings")
		return internalProbeError(err)
	}

	if bucketMaxFileSize != nil && *bucketMaxFileSize > 0 && size > *bucketMaxFileSize {
		return &authzProbeError{
			status:  fiber.StatusRequestEntityTooLarge,
			message: fmt.Sprintf("file size %d exceeds bucket maximum of %d bytes", size, *bucketMaxFileSize),
			code:    ErrCodeInvalidInput,
		}
	}

	if len(bucketAllowedMimeTypes) > 0 && contentType != "" {
		mimeAllowed := false
		for _, allowedType := range bucketAllowedMimeTypes {
			if allowedType == contentType || allowedType == "*/*" {
				mimeAllowed = true
				break
			}
			// Support wildcard matching (e.g., "image/*")
			if strings.HasSuffix(allowedType, "/*") {
				prefix := strings.TrimSuffix(allowedType, "/*")
				if strings.HasPrefix(contentType, prefix+"/") {
					mimeAllowed = true
					break
				}
			}
		}
		if !mimeAllowed {
			return &authzProbeError{
				status:  fiber.StatusUnsupportedMediaType,
				message: fmt.Sprintf("file type %s is not allowed for this bucket", contentType),
				code:    ErrCodeInvalidFormat,
			}
		}
	}

	return nil
}

// authorizeSignedAccess validates object-level access for signed-URL issuance
// in a single RLS transaction:
//
//   - GET/DELETE: the object row must be visible under RLS (same visibility
//     the download path enforces).
//   - PUT: if the row is visible it must also be writable (no-op UPDATE); if
//     it is not visible (new object) the bucket row must at least be visible
//     under RLS.
func (h *StorageHandler) authorizeSignedAccess(c fiber.Ctx, bucket, key, method string) *authzProbeError {
	ctx := c.RequestCtx()

	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Resolve tenant context from the object itself (mirrors DownloadFile).
	h.resolveTenantForObject(ctx, tx, c, bucket, key)

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return internalProbeError(err)
	}

	var visible bool
	if err := tx.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)
	`, bucket, key).Scan(&visible); err != nil {
		return internalProbeError(err)
	}

	if !visible {
		if method != "PUT" {
			return &authzProbeError{status: fiber.StatusNotFound, message: "object not found or insufficient permissions", code: ErrCodeNotFound}
		}
		// New object: the caller must at least be able to see the bucket.
		var bucketVisible bool
		if err := tx.QueryRow(ctx, `
			SELECT EXISTS(SELECT 1 FROM storage.buckets WHERE id = $1)
		`, bucket).Scan(&bucketVisible); err != nil {
			return internalProbeError(err)
		}
		if !bucketVisible {
			return &authzProbeError{status: fiber.StatusNotFound, message: "bucket not found or insufficient permissions", code: ErrCodeNotFound}
		}
	} else if method == "PUT" || method == "DELETE" {
		result, err := tx.Exec(ctx, `
			UPDATE storage.objects SET updated_at = updated_at
			WHERE bucket_id = $1 AND path = $2
		`, bucket, key)
		if err != nil {
			if isRLSDenial(err) {
				return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to write object", code: ErrCodeAccessDenied, err: err}
			}
			return internalProbeError(err)
		}
		if result.RowsAffected() == 0 {
			return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to write object", code: ErrCodeAccessDenied}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return internalProbeError(err)
	}
	return nil
}

// authorizeUploadIntent validates, without leaving any database state behind,
// that the caller may create or overwrite the object at (bucket, key):
//
//   - if an object row is visible under RLS, a no-op UPDATE verifies
//     overwrite permission;
//   - otherwise a placeholder row is inserted AND deleted in one transaction,
//     which validates the RLS insert policy without reserving the path.
//
// Used by chunked-upload init so a rejected caller never writes chunk bytes.
func (h *StorageHandler) authorizeUploadIntent(c fiber.Ctx, bucket, key, contentType string, ownerID *string) *authzProbeError {
	ctx := c.RequestCtx()

	exists, perr := h.objectRowVisible(ctx, c, bucket, key)
	if perr != nil {
		return perr
	}
	if exists {
		return h.probeWritePermission(ctx, c, bucket, key)
	}

	pool, perr := h.probePool(c)
	if perr != nil {
		return perr
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return internalProbeError(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO storage.objects (bucket_id, path, mime_type, size, metadata, owner_id)
		VALUES ($1, $2, $3, 0, NULL, $4)
	`, bucket, key, contentType, ownerID); err != nil {
		if isRLSDenial(err) {
			return &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to upload object", code: ErrCodeAccessDenied, err: err}
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgForeignKeyViolationCode {
			return &authzProbeError{status: fiber.StatusNotFound, message: "bucket does not exist", code: ErrCodeNotFound}
		}
		return internalProbeError(err)
	}

	if _, err := tx.Exec(ctx, `DELETE FROM storage.objects WHERE bucket_id = $1 AND path = $2`, bucket, key); err != nil {
		return internalProbeError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return internalProbeError(err)
	}
	return nil
}
