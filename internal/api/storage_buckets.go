package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
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
		INSERT INTO storage.buckets (id, name, public, allowed_mime_types, max_file_size)
		VALUES ($1, $2, $3, $4, $5)
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

	return c.JSON(fiber.Map{
		"message": "bucket settings updated successfully",
	})
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

	if err := svc.Provider.DeleteBucket(c.RequestCtx(), bucket); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return SendNotFound(c, "bucket not found")
		}
		if strings.Contains(err.Error(), "not empty") {
			return SendConflict(c, "bucket is not empty", ErrCodeConflict)
		}
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to delete bucket")
		return SendInternalError(c, "Failed to delete bucket")
	}

	log.Info().
		Str("bucket", bucket).
		Str("user_id", getUserID(c)).
		Msg("Bucket deleted")

	return c.Status(fiber.StatusNoContent).Send(nil)
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
