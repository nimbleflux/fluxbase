package api

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"
)

// MultipartUpload handles multipart upload
// POST /api/v1/storage/:bucket/multipart
func (h *StorageHandler) MultipartUpload(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "failed to get storage service")
	}

	bucket := c.Params("bucket")

	if bucket == "" {
		return SendBadRequest(c, "bucket is required", ErrCodeMissingField)
	}

	// H-19: Check if bucket exists before upload
	// Use SECURITY DEFINER function to bypass RLS when checking bucket existence
	var bucketExists bool
	err = h.db.Pool().QueryRow(
		c.RequestCtx(),
		`SELECT storage.bucket_exists($1::text, $2::uuid)`,
		bucket, getTenantIDArg(c),
	).Scan(&bucketExists)
	if err != nil {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to check bucket existence")
		return SendInternalError(c, "failed to validate bucket")
	}
	if !bucketExists {
		return SendNotFound(c, fmt.Sprintf("bucket '%s' does not exist", bucket))
	}

	// C-3: Get bucket MIME type settings
	// Use SECURITY DEFINER function to bypass RLS when fetching bucket settings
	var bucketAllowedMimeTypes []string
	err = h.db.Pool().QueryRow(
		c.RequestCtx(),
		`SELECT allowed_mime_types FROM storage.get_bucket_settings($1::text, $2::uuid)`,
		bucket, getTenantIDArg(c),
	).Scan(&bucketAllowedMimeTypes)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		log.Error().Err(err).Str("bucket", bucket).Msg("Failed to get bucket settings")
		return SendInternalError(c, "failed to validate bucket settings")
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return SendBadRequest(c, "failed to parse multipart form", ErrCodeInvalidInput)
	}

	files := form.File["files"]
	if len(files) == 0 {
		return SendBadRequest(c, "no files provided", ErrCodeMissingField)
	}

	var uploaded []storage.Object
	var errors []string

	// Upload each file
	for _, file := range files {
		key := file.Filename

		// H-20: Sanitize filename
		key = sanitizeFilename(key)
		if key == "" {
			errors = append(errors, fmt.Sprintf("%s: invalid filename after sanitization", file.Filename))
			continue
		}

		// Validate file size
		if err := svc.ValidateUploadSize(file.Size); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// C-3: Detect content type for MIME validation
		contentType := file.Header.Get("Content-Type")
		if contentType == "" {
			contentType = detectContentType(file.Filename)
		}

		// C-3: Validate MIME type against bucket-specific allowed types
		if len(bucketAllowedMimeTypes) > 0 {
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
				errors = append(errors, fmt.Sprintf("%s: file type %s is not allowed for this bucket", file.Filename, contentType))
				continue
			}
		}

		// Upload file
		if err := h.uploadMultipartFile(c, svc, bucket, key, file, contentType); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		uploaded = append(uploaded, storage.Object{
			Key:    key,
			Bucket: bucket,
			Size:   file.Size,
		})
	}

	response := fiber.Map{
		"uploaded": uploaded,
		"count":    len(uploaded),
	}

	if len(errors) > 0 {
		response["errors"] = errors
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// uploadMultipartFile uploads a single file from multipart form and records its
// metadata row in storage.objects. This mirrors StorageHandler.UploadFile in
// storage_files.go — without the metadata insert, the uploaded bytes would be
// invisible to the API (download/list/share/delete key off the objects row) and
// the file would have no owner_id, failing the storage_objects_insert RLS policy.
func (h *StorageHandler) uploadMultipartFile(c fiber.Ctx, svc *storage.Service, bucket, key string, file *multipart.FileHeader, contentType string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = src.Close() }()

	opts := &storage.UploadOptions{
		ContentType: contentType,
	}

	ctx := c.RequestCtx()

	// Get owner ID from authenticated user (same source as the regular upload path)
	ownerID := getUserID(c)
	var ownerUUID *string
	if ownerID != "" && ownerID != "anonymous" {
		ownerUUID = &ownerID
	}

	// Authorization runs BEFORE any bytes are written to the provider: either
	// the caller can overwrite an existing object row (no-op UPDATE under RLS)
	// or a size=0 placeholder row is inserted under RLS to validate insert
	// permission and reserve the path. On any post-probe failure we only delete
	// provider bytes when we created the placeholder.
	probe, perr := h.AuthorizeUploadWrite(c, bucket, key, contentType, nil, ownerUUID)
	if perr != nil {
		return perr.send(c)
	}
	createdPlaceholder := probe.createdPlaceholder

	failUpload := func(err error, message string) error {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, bucket, key)
		}
		return fmt.Errorf("%s", message)
	}

	// Upload the file to the storage provider
	object, err := svc.Provider.Upload(ctx, bucket, key, src, file.Size, opts)
	if err != nil {
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to upload multipart file")
		return failUpload(err, "failed to upload file")
	}

	// Persist the final metadata (size, mime, owner) under RLS
	if err := h.upsertMultipartMetadata(ctx, c, bucket, key, contentType, object.Size, ownerUUID); err != nil {
		errMsg := err.Error()
		log.Error().
			Err(err).
			Str("bucket", bucket).
			Str("key", key).
			Str("error_message", errMsg).
			Msg("Failed to insert multipart file metadata into database")
		if strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "policy") {
			if createdPlaceholder {
				h.CleanupUploadPlaceholder(c, svc, bucket, key)
			}
			return fmt.Errorf("insufficient permissions to upload file")
		}
		return failUpload(err, "failed to save file metadata")
	}

	log.Info().
		Str("bucket", bucket).
		Str("key", key).
		Int64("size", object.Size).
		Str("user_id", ownerID).
		Msg("Multipart file uploaded")

	return nil
}

// upsertMultipartMetadata writes the final object row for a multipart upload
// under RLS (INSERT .. ON CONFLICT DO UPDATE).
func (h *StorageHandler) upsertMultipartMetadata(ctx context.Context, c fiber.Ctx, bucket, key, contentType string, size int64, ownerUUID *string) error {
	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO storage.objects (bucket_id, path, mime_type, size, metadata, owner_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (bucket_id, path)
		DO UPDATE SET mime_type = $3, size = $4, owner_id = $6, updated_at = NOW()
	`, bucket, key, contentType, size, nil, ownerUUID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// fiber:context-methods migrated
