package api

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"

	apperrors "github.com/nimbleflux/fluxbase/internal/errors"
)

func (h *StorageHandler) CreateBucket(c fiber.Ctx) error {
	bucket := c.Params("bucket")

	if bucket == "" {
		return SendMissingField(c, "bucket name")
	}

	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	var req struct {
		Public           bool     `json:"public"`
		AllowedMimeTypes []string `json:"allowed_mime_types"`
		MaxFileSize      *int64   `json:"max_file_size"`
	}
	_ = c.Bind().Body(&req)

	if h.db == nil {
		return SendInternalError(c, "Database connection not initialized")
	}

	ctx := c.RequestCtx()
	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for bucket creation")
		return SendInternalError(c, "Failed to create bucket")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to create bucket")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO storage.buckets (id, name, public, allowed_mime_types, max_file_size, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`, bucket, bucket, req.Public, req.AllowedMimeTypes, req.MaxFileSize)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists") {
			return SendConflict(c, "bucket already exists", ErrCodeConflict)
		}
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return SendForbidden(c, "insufficient permissions to create bucket", ErrCodeAccessDenied)
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to insert bucket into database")
		return SendInternalError(c, "Failed to create bucket")
	}

	if err := svc.Provider.CreateBucket(ctx, bucket); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return SendConflict(c, "bucket already exists in storage", ErrCodeConflict)
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to create bucket in provider")
		return SendInternalError(c, "Failed to create bucket in storage provider")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to commit bucket creation")
		return SendInternalError(c, "Failed to create bucket")
	}

	log.Info().
		Str("bucket", bucket).
		Bool("public", req.Public).
		Str("user_id", getUserID(c)).
		Msg("Bucket created")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"bucket":             bucket,
		"id":                 bucket,
		"name":               bucket,
		"public":             req.Public,
		"allowed_mime_types": req.AllowedMimeTypes,
		"max_file_size":      req.MaxFileSize,
		"message":            "bucket created successfully",
	})
}

func (h *StorageHandler) UpdateBucketSettings(c fiber.Ctx) error {
	bucket := c.Params("bucket")

	if bucket == "" {
		return SendMissingField(c, "bucket name")
	}

	var req struct {
		Public           *bool    `json:"public"`
		AllowedMimeTypes []string `json:"allowed_mime_types"`
		MaxFileSize      *int64   `json:"max_file_size"`
	}
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if h.db == nil {
		return SendInternalError(c, "Database connection not initialized")
	}

	ctx := c.RequestCtx()

	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for bucket update")
		return SendInternalError(c, "Failed to update bucket")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to update bucket")
	}

	updates := []string{}
	args := []interface{}{bucket}
	argCount := 1

	if req.Public != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("public = $%d", argCount))
		args = append(args, *req.Public)
	}

	if req.AllowedMimeTypes != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("allowed_mime_types = $%d", argCount))
		args = append(args, req.AllowedMimeTypes)
	}

	if req.MaxFileSize != nil {
		argCount++
		updates = append(updates, fmt.Sprintf("max_file_size = $%d", argCount))
		args = append(args, req.MaxFileSize)
	}

	if len(updates) == 0 {
		return SendBadRequest(c, "no fields to update", ErrCodeInvalidInput)
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf(`
		UPDATE storage.buckets
		SET %s
		WHERE id = $1
	`, strings.Join(updates, ", "))

	result, err := tx.Exec(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return SendForbidden(c, "insufficient permissions to update bucket", ErrCodeAccessDenied)
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to update bucket in database")
		return SendInternalError(c, "Failed to update bucket")
	}

	if result.RowsAffected() == 0 {
		return SendNotFound(c, "bucket not found or insufficient permissions")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to commit bucket update")
		return SendInternalError(c, "Failed to update bucket")
	}

	log.Info().
		Str("bucket", bucket).
		Str("user_id", getUserID(c)).
		Interface("updates", req).
		Msg("Bucket settings updated")

	return apperrors.SendSuccess(c, "bucket settings updated successfully")
}

func (h *StorageHandler) DeleteBucket(c fiber.Ctx) error {
	bucket := c.Params("bucket")

	if bucket == "" {
		return SendMissingField(c, "bucket name")
	}

	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	if h.db == nil {
		return SendInternalError(c, "Database connection not initialized")
	}

	ctx := c.RequestCtx()

	// Check the provider bucket is empty BEFORE touching the database so a
	// rejected delete cannot leave the bucket row gone while the provider
	// bucket (and its objects) still exists.
	empty, err := providerBucketEmpty(ctx, svc.Provider, bucket)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return SendNotFound(c, "bucket not found")
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to inspect provider bucket")
		return SendInternalError(c, "Failed to delete bucket")
	}
	if !empty {
		return SendConflict(c, "bucket is not empty", ErrCodeConflict)
	}

	// Authorization gate: RLS-gated delete of the bucket row. Row level
	// security decides whether the caller may delete this bucket; the FK
	// cascade from storage.objects/permissions takes care of the metadata
	// rows in the same transaction.
	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for bucket deletion")
		return SendInternalError(c, "Failed to delete bucket")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to delete bucket")
	}

	result, err := tx.Exec(ctx, `DELETE FROM storage.buckets WHERE id = $1`, bucket)
	if err != nil {
		if strings.Contains(err.Error(), "permission denied") || strings.Contains(err.Error(), "policy") {
			return SendForbidden(c, "insufficient permissions to delete bucket", ErrCodeAccessDenied)
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to delete bucket from database")
		return SendInternalError(c, "Failed to delete bucket")
	}

	if result.RowsAffected() == 0 {
		// Distinguish 404 (bucket absent) from 403 (RLS blocked the delete).
		// The SECURITY DEFINER helper bypasses RLS for the existence check.
		var bucketExists bool
		err := h.getPool(c).QueryRow(ctx, `SELECT storage.bucket_exists($1::text, $2::uuid)`, bucket, getTenantIDArg(c)).Scan(&bucketExists)
		if err == nil && bucketExists {
			return SendForbidden(c, "insufficient permissions to delete bucket", ErrCodeAccessDenied)
		}
		return SendNotFound(c, "bucket not found or insufficient permissions")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to commit bucket deletion")
		return SendInternalError(c, "Failed to delete bucket")
	}

	// Provider delete runs after the authorization gate. If it fails the
	// bucket row is already gone — acceptable and logged.
	if err := svc.Provider.DeleteBucket(ctx, bucket); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return SendNotFound(c, "bucket not found")
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to delete bucket in provider (bucket row already deleted)")
		return SendInternalError(c, "Failed to delete bucket in storage provider")
	}

	log.Info().
		Str("bucket", bucket).
		Str("user_id", getUserID(c)).
		Msg("Bucket deleted")

	return c.Status(fiber.StatusNoContent).Send(nil)
}

// providerBucketEmpty reports whether the provider-side bucket holds no
// objects.
func providerBucketEmpty(ctx context.Context, provider storage.Provider, bucket string) (bool, error) {
	result, err := provider.List(ctx, bucket, &storage.ListOptions{MaxKeys: 1})
	if err != nil {
		return false, err
	}
	return len(result.Objects) == 0, nil
}

func (h *StorageHandler) ListBuckets(c fiber.Ctx) error {
	role, _ := c.Locals("user_role").(string)
	isInstanceAdmin, _ := c.Locals("is_instance_admin").(bool)
	tenantRole, _ := c.Locals("tenant_role").(string)
	isAuthorized := role == "admin" || role == "instance_admin" || role == "service_role" || role == "tenant_service" ||
		isInstanceAdmin || tenantRole == "tenant_admin" || tenantRole == "tenant_service"
	if !isAuthorized {
		return SendForbidden(c, "Admin access required to list buckets", ErrCodeAccessDenied)
	}

	if h.db == nil {
		return SendInternalError(c, "Database connection not initialized")
	}

	ctx := c.RequestCtx()

	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start transaction for listing buckets")
		return SendInternalError(c, "Failed to list buckets")
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		log.Error().Err(err).Msg("Failed to set RLS context")
		return SendInternalError(c, "Failed to list buckets")
	}

	rows, err := tx.Query(ctx, `
		SELECT id, name, public, allowed_mime_types, max_file_size, created_at, updated_at
		FROM storage.buckets
		ORDER BY created_at DESC
	`)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query buckets from database")
		return SendInternalError(c, "Failed to list buckets")
	}
	defer rows.Close()

	type Bucket struct {
		ID               string    `json:"id"`
		Name             string    `json:"name"`
		Public           bool      `json:"public"`
		AllowedMimeTypes []string  `json:"allowed_mime_types"`
		MaxFileSize      *int64    `json:"max_file_size"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
	}

	var buckets []Bucket
	for rows.Next() {
		var b Bucket
		if err := rows.Scan(&b.ID, &b.Name, &b.Public, &b.AllowedMimeTypes, &b.MaxFileSize, &b.CreatedAt, &b.UpdatedAt); err != nil {
			log.Error().Err(err).Msg("Failed to scan bucket row")
			continue
		}
		buckets = append(buckets, b)
	}

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("Error iterating bucket rows")
		return SendInternalError(c, "Failed to list buckets")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to commit bucket list transaction")
		return SendInternalError(c, "Failed to list buckets")
	}

	return c.JSON(fiber.Map{
		"buckets": buckets,
	})
}

// fiber:context-methods migrated
