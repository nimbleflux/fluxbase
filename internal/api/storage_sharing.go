package api

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// ShareObject handles sharing a file with another user
// POST /api/v1/storage/:bucket/:path/share
func (h *StorageHandler) ShareObject(c fiber.Ctx) error {
	bucket := c.Params("bucket")
	key := c.Params("*")

	if bucket == "" || key == "" {
		return SendBadRequest(c, "Bucket and key are required", ErrCodeMissingField)
	}

	var req struct {
		UserID     string `json:"user_id"`
		Permission string `json:"permission"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.Permission != "read" && req.Permission != "write" {
		return SendBadRequest(c, "Permission must be 'read' or 'write'", ErrCodeInvalidInput)
	}

	if req.UserID == "" {
		return SendMissingField(c, "user_id")
	}

	ctx := c.RequestCtx()

	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for sharing file")
		return SendInternalError(c, "Failed to share file")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to share file")
	}

	var objectID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM storage.objects
		WHERE bucket_id = $1 AND path = $2
	`, bucket, key).Scan(&objectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SendNotFound(c, "File not found or insufficient permissions")
		}
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to find file")
		return SendInternalError(c, "Failed to share file")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO storage.object_permissions (object_id, user_id, permission)
		VALUES ($1, $2, $3)
		ON CONFLICT (object_id, user_id)
		DO UPDATE SET permission = $3
	`, objectID, req.UserID, req.Permission)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return SendForbidden(c, "Only file owners can share files", ErrCodeAccessDenied)
		}
		log.Error().Err(err).Str("object_id", objectID).Msg("Failed to create file permission")
		return SendInternalError(c, "Failed to share file")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit share transaction")
		return SendInternalError(c, "Failed to share file")
	}

	log.Info().
		Str("bucket", bucket).
		Str("key", key).
		Str("shared_with", req.UserID).
		Str("permission", req.Permission).
		Str("user_id", getUserID(c)).
		Msg("File shared")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "file shared successfully",
		"user_id":    req.UserID,
		"permission": req.Permission,
	})
}

// RevokeShare handles revoking file access from a user
// DELETE /api/v1/storage/:bucket/:path/share/:user_id
func (h *StorageHandler) RevokeShare(c fiber.Ctx) error {
	bucket := c.Params("bucket")
	key := c.Params("*1")
	sharedUserID := c.Params("user_id")

	if bucket == "" || key == "" || sharedUserID == "" {
		return SendBadRequest(c, "Bucket, key, and user_id are required", ErrCodeMissingField)
	}

	ctx := c.RequestCtx()

	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for revoking share")
		return SendInternalError(c, "Failed to revoke share")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to revoke share")
	}

	var objectID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM storage.objects
		WHERE bucket_id = $1 AND path = $2
	`, bucket, key).Scan(&objectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SendNotFound(c, "File not found or insufficient permissions")
		}
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to find file")
		return SendInternalError(c, "Failed to revoke share")
	}

	result, err := tx.Exec(ctx, `
		DELETE FROM storage.object_permissions
		WHERE object_id = $1 AND user_id = $2
	`, objectID, sharedUserID)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return SendForbidden(c, "Only file owners can revoke shares", ErrCodeAccessDenied)
		}
		log.Error().Err(err).Str("object_id", objectID).Msg("Failed to delete file permission")
		return SendInternalError(c, "Failed to revoke share")
	}

	if result.RowsAffected() == 0 {
		return SendNotFound(c, "Share not found or insufficient permissions")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit revoke share transaction")
		return SendInternalError(c, "Failed to revoke share")
	}

	log.Info().
		Str("bucket", bucket).
		Str("key", key).
		Str("revoked_from", sharedUserID).
		Str("user_id", getUserID(c)).
		Msg("File share revoked")

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// ListShares handles listing users a file is shared with
// GET /api/v1/storage/:bucket/:path/shares
func (h *StorageHandler) ListShares(c fiber.Ctx) error {
	bucket := c.Params("bucket")
	key := c.Params("*")

	if bucket == "" || key == "" {
		return SendBadRequest(c, "Bucket and key are required", ErrCodeMissingField)
	}

	ctx := c.RequestCtx()

	tx, err := h.db.Pool().Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for listing shares")
		return SendInternalError(c, "Failed to list shares")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to list shares")
	}

	var objectID string
	err = tx.QueryRow(ctx, `
		SELECT id FROM storage.objects
		WHERE bucket_id = $1 AND path = $2
	`, bucket, key).Scan(&objectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SendNotFound(c, "File not found or insufficient permissions")
		}
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to find file")
		return SendInternalError(c, "Failed to list shares")
	}

	rows, err := tx.Query(ctx, `
		SELECT user_id, permission, created_at
		FROM storage.object_permissions
		WHERE object_id = $1
		ORDER BY created_at DESC
	`, objectID)
	if err != nil {
		log.Error().Err(err).Str("object_id", objectID).Msg("Failed to query shares")
		return SendInternalError(c, "Failed to list shares")
	}
	defer rows.Close()

	type Share struct {
		UserID     string    `json:"user_id"`
		Permission string    `json:"permission"`
		CreatedAt  time.Time `json:"created_at"`
	}

	var shares []Share
	for rows.Next() {
		var share Share
		if err := rows.Scan(&share.UserID, &share.Permission, &share.CreatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan share row")
			continue
		}
		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating share rows")
		return SendInternalError(c, "Failed to list shares")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit list shares transaction")
		return SendInternalError(c, "Failed to list shares")
	}

	return c.JSON(fiber.Map{
		"shares": shares,
		"count":  len(shares),
	})
}

// fiber:context-methods migrated
