package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"
)

// StreamUpload handles streaming file upload with reduced memory usage
// POST /api/v1/storage/:bucket/stream/:key
//
// This endpoint reads the raw request body as a stream, avoiding the memory
// overhead of multipart form parsing. Use this for large file uploads.
//
// Headers:
//   - Content-Length: Required. The size of the file in bytes.
//   - X-Storage-Content-Type: Optional. The MIME type of the file.
//   - X-Storage-Cache-Control: Optional. Cache-Control header value.
//   - X-Storage-Metadata: Optional. JSON object with custom metadata.
//   - X-Storage-Upsert: Optional. "true" to overwrite existing files.
func (h *StorageHandler) StreamUpload(c fiber.Ctx) error {
	bucket := c.Params("bucket")
	key := c.Params("*") // Capture the rest of the path
	key = sanitizeFilename(key)

	if bucket == "" || key == "" {
		return SendBadRequest(c, "bucket and key are required", ErrCodeMissingField)
	}

	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "failed to get storage service")
	}

	// Get file size from Content-Length header (required for streaming)
	size := int64(c.Request().Header.ContentLength())
	if size <= 0 {
		return SendBadRequest(c, "Content-Length header is required for streaming uploads", ErrCodeMissingField)
	}

	// Validate file size
	if err := svc.ValidateUploadSize(size); err != nil {
		return SendErrorWithCode(c, fiber.StatusRequestEntityTooLarge, err.Error(), ErrCodeInvalidInput)
	}

	// Get content type from header
	contentType := c.Get("X-Storage-Content-Type")
	if contentType == "" {
		contentType = c.Get("Content-Type", "application/octet-stream")
		// Don't use multipart content-type for the file itself
		if strings.HasPrefix(contentType, "multipart/") {
			contentType = "application/octet-stream"
		}
	}

	// Parse metadata from header
	metadata := make(map[string]string)
	if metadataHeader := c.Get("X-Storage-Metadata"); metadataHeader != "" {
		if err := json.Unmarshal([]byte(metadataHeader), &metadata); err != nil {
			return SendBadRequest(c, "invalid X-Storage-Metadata header: must be valid JSON object", ErrCodeInvalidInput)
		}
	}

	// Get cache control
	cacheControl := c.Get("X-Storage-Cache-Control")

	// Upload options
	opts := &storage.UploadOptions{
		ContentType:  contentType,
		Metadata:     metadata,
		CacheControl: cacheControl,
	}

	// Get owner ID from authenticated user
	ownerID := getUserID(c)
	var ownerUUID *string
	if ownerID != "" && ownerID != "anonymous" {
		ownerUUID = &ownerID
	}

	ctx := c.RequestCtx()

	// Convert metadata header to the JSONB shape used by storage.objects
	var metadataJSON map[string]interface{}
	if len(metadata) > 0 {
		metadataJSON = make(map[string]interface{})
		for k, v := range metadata {
			metadataJSON[k] = v
		}
	}

	// Authorization runs BEFORE any bytes are written to the provider: either
	// the caller can overwrite an existing object row (no-op UPDATE under RLS)
	// or a size=0 placeholder row is inserted under RLS to validate insert
	// permission and reserve the path. On any post-probe failure we only delete
	// provider bytes when we created the placeholder — never for a pre-existing
	// object row.
	probe, perr := h.AuthorizeUploadWrite(c, bucket, key, contentType, metadataJSON, ownerUUID)
	if perr != nil {
		return perr.send(c)
	}
	createdPlaceholder := probe.createdPlaceholder

	// Metadata upsert shared by the final write below.
	upsertMetadata := func() error {
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
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (bucket_id, path)
			DO UPDATE SET mime_type = $3, size = $4, metadata = $5, owner_id = $6, updated_at = NOW()
		`, bucket, key, contentType, size, metadataJSON, ownerUUID); err != nil {
			return err
		}

		return tx.Commit(ctx)
	}

	// Get the request body as a stream reader
	// Try streaming first, fall back to buffered body
	var body io.Reader
	body = c.Request().BodyStream()
	if body == nil {
		// BodyStream can be nil if the body was buffered as bytes
		// Fall back to reading the buffered body
		bodyBytes := c.Body()
		if len(bodyBytes) == 0 {
			if createdPlaceholder {
				h.CleanupUploadPlaceholder(c, svc, bucket, key)
			}
			return SendBadRequest(c, "request body is required", ErrCodeMissingField)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Upload the file to storage provider (streaming)
	object, err := svc.Provider.Upload(ctx, bucket, key, body, size, opts)
	if err != nil {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, bucket, key)
		}
		log.Error().Err(err).Str("bucket", bucket).Str("key", key).Msg("Failed to upload file (streaming)")
		return SendInternalError(c, "failed to upload file")
	}

	// Persist the final metadata (size, mime, metadata, owner)
	if err := upsertMetadata(); err != nil {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, bucket, key)
		}

		// Log the full error for debugging
		errMsg := err.Error()
		log.Error().
			Err(err).
			Str("bucket", bucket).
			Str("key", key).
			Str("owner_id", fmt.Sprintf("%v", ownerUUID)).
			Str("error_message", errMsg).
			Msg("Failed to insert file metadata into database (streaming)")

		if strings.Contains(errMsg, "permission denied") || strings.Contains(errMsg, "policy") {
			return SendErrorWithDetails(c, fiber.StatusForbidden, "insufficient permissions to upload file", ErrCodeAccessDenied, "", "", errMsg)
		}
		return SendInternalError(c, "failed to save file metadata")
	}

	log.Info().
		Str("bucket", bucket).
		Str("key", key).
		Int64("size", object.Size).
		Str("user_id", ownerID).
		Msg("File uploaded (streaming)")

	// Add owner_id to response
	response := map[string]interface{}{
		"key":           object.Key,
		"bucket":        object.Bucket,
		"size":          object.Size,
		"content_type":  object.ContentType,
		"last_modified": object.LastModified,
	}
	if ownerUUID != nil {
		response["owner_id"] = *ownerUUID
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}

// fiber:context-methods migrated
