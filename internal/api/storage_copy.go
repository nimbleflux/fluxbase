package api

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// CopyObjectHandler copies an object within a bucket (or to another bucket).
// POST /api/v1/storage/:bucket/copy  body: {"from_path": "...", "to_path": "...", "to_bucket"?: "..."}
//
// The TypeScript SDK calls this endpoint from StorageFileApi.copy; previously
// it was shadowed by the /:bucket/* upload wildcard and mis-executed as a
// multipart upload on a JSON body.
//
// Authorization mirrors the download/upload paths: the source object row must
// be visible under RLS (AuthorizeObjectRead) and the destination must pass the
// upload write probe (AuthorizeUploadWrite) BEFORE any bytes are copied.
func (h *StorageHandler) CopyObjectHandler(c fiber.Ctx) error {
	return h.copyOrMoveObject(c, false)
}

// MoveObjectHandler moves an object (copy + delete of the source).
// POST /api/v1/storage/:bucket/move  body: {"from_path": "...", "to_path": "...", "to_bucket"?: "..."}
//
// The source delete is RLS-checked: the caller must be able to delete the
// source object row before the provider bytes are removed.
func (h *StorageHandler) MoveObjectHandler(c fiber.Ctx) error {
	return h.copyOrMoveObject(c, true)
}

func (h *StorageHandler) copyOrMoveObject(c fiber.Ctx, move bool) error {
	var req struct {
		FromPath string `json:"from_path"`
		ToPath   string `json:"to_path"`
		ToBucket string `json:"to_bucket,omitempty"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.FromPath == "" {
		return SendBadRequest(c, "from_path is required", ErrCodeMissingField)
	}
	if req.ToPath == "" {
		return SendBadRequest(c, "to_path is required", ErrCodeMissingField)
	}

	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "failed to get storage service")
	}

	srcBucket := c.Params("bucket")
	destBucket := req.ToBucket
	if destBucket == "" {
		destBucket = srcBucket
	}
	srcKey := sanitizeFilename(req.FromPath)
	destKey := sanitizeFilename(req.ToPath)
	if srcKey == "" || destKey == "" {
		return SendBadRequest(c, "invalid path after sanitization", ErrCodeInvalidInput)
	}

	ctx := c.RequestCtx()

	// 1. Source must be readable by the caller (RLS visibility probe).
	if perr := h.AuthorizeObjectRead(c, srcBucket, srcKey); perr != nil {
		return perr.send(c)
	}

	// Fetch source metadata under service context for the destination row.
	var srcMime *string
	var srcSize int64
	err = h.getPool(c).QueryRow(ctx, `
		SELECT mime_type, size FROM storage.objects WHERE bucket_id = $1 AND path = $2
	`, srcBucket, srcKey).Scan(&srcMime, &srcSize)
	if err != nil {
		log.Error().Err(err).Str("bucket", srcBucket).Str("key", srcKey).Msg("Failed to load source object metadata")
		return SendInternalError(c, "failed to copy object")
	}

	contentType := "application/octet-stream"
	if srcMime != nil && *srcMime != "" {
		contentType = *srcMime
	}

	// The destination object is owned by the calling user (matching upload
	// semantics), not by the source owner.
	var destOwner *string
	if ownerID := getUserID(c); ownerID != "" && ownerID != "anonymous" {
		destOwner = &ownerID
	}

	// 2. Destination must pass the upload write probe (insert permission for a
	// new object, overwrite permission for an existing one) BEFORE bytes move.
	probe, perr := h.AuthorizeUploadWrite(c, destBucket, destKey, contentType, nil, destOwner)
	if perr != nil {
		return perr.send(c)
	}
	createdPlaceholder := probe.createdPlaceholder

	// 3. Copy the bytes in the provider.
	if err := svc.Provider.CopyObject(ctx, srcBucket, srcKey, destBucket, destKey); err != nil {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, destBucket, destKey)
		}
		log.Error().Err(err).Str("src", srcBucket+"/"+srcKey).Str("dest", destBucket+"/"+destKey).Msg("Failed to copy object in provider")
		return SendInternalError(c, "failed to copy object")
	}

	// 4. Persist the destination row.
	if err := h.upsertCopiedObjectMetadata(c, destBucket, destKey, contentType, srcSize, destOwner); err != nil {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, destBucket, destKey)
		}
		log.Error().Err(err).Str("dest", destBucket+"/"+destKey).Msg("Failed to store copied object metadata")
		return SendInternalError(c, "failed to save object metadata")
	}

	// 5. For moves, delete the source under RLS (metadata first, then bytes).
	if move {
		deleted, perr := h.deleteObjectRowRLS(c, srcBucket, srcKey)
		if perr != nil {
			if createdPlaceholder {
				// Restore the original state: drop the destination copy.
				_ = svc.Provider.Delete(ctx, destBucket, destKey)
				_, _ = h.deleteObjectRowRLS(c, destBucket, destKey)
			} else {
				// The destination row pre-existed and is caller-writable; the
				// copy overwrote its bytes. Keep row+bytes consistent instead
				// of deleting a row the caller does not own.
				log.Warn().Err(perr).Str("dest", destBucket+"/"+destKey).Msg("Move failed after overwrite; destination kept in place")
			}
			return perr.send(c)
		}
		if deleted {
			if err := svc.Provider.Delete(ctx, srcBucket, srcKey); err != nil {
				log.Warn().Err(err).Str("bucket", srcBucket).Str("key", srcKey).Msg("Failed to delete source bytes after move (metadata already deleted)")
			}
		}
	}

	verb := "copied"
	if move {
		verb = "moved"
	}
	log.Info().
		Str("src", srcBucket+"/"+srcKey).
		Str("dest", destBucket+"/"+destKey).
		Str("user_id", getUserID(c)).
		Msg("Object " + verb)

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("object %s successfully", verb),
		"source": fiber.Map{
			"bucket": srcBucket,
			"path":   srcKey,
		},
		"destination": fiber.Map{
			"bucket": destBucket,
			"path":   destKey,
		},
	})
}

// upsertCopiedObjectMetadata writes the final destination row for a copy/move.
func (h *StorageHandler) upsertCopiedObjectMetadata(c fiber.Ctx, bucket, key, contentType string, size int64, ownerID *string) error {
	ctx := c.RequestCtx()

	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO storage.objects (bucket_id, path, mime_type, size, metadata, owner_id)
		VALUES ($1, $2, $3, $4, NULL, $5)
		ON CONFLICT (bucket_id, path)
		DO UPDATE SET mime_type = $3, size = $4, owner_id = $5, updated_at = NOW()
	`, bucket, key, contentType, size, ownerID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// deleteObjectRowRLS deletes an object row under the request's RLS context.
// Returns (false, nil) when the row is not visible to the caller.
func (h *StorageHandler) deleteObjectRowRLS(c fiber.Ctx, bucket, key string) (bool, *authzProbeError) {
	ctx := c.RequestCtx()

	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		return false, internalProbeError(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return false, internalProbeError(err)
	}

	result, err := tx.Exec(ctx, `DELETE FROM storage.objects WHERE bucket_id = $1 AND path = $2`, bucket, key)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return false, &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to delete object", code: ErrCodeAccessDenied, err: err}
		}
		return false, internalProbeError(err)
	}

	if result.RowsAffected() == 0 {
		// Not visible under RLS: either absent or blocked — check presence.
		var exists bool
		if err := h.getPool(c).QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM storage.objects WHERE bucket_id = $1 AND path = $2)`, bucket, key).Scan(&exists); err == nil && exists {
			return false, &authzProbeError{status: fiber.StatusForbidden, message: "insufficient permissions to delete object", code: ErrCodeAccessDenied}
		}
		return false, &authzProbeError{status: fiber.StatusNotFound, message: "object not found or insufficient permissions", code: ErrCodeNotFound}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, internalProbeError(err)
	}
	return true, nil
}
