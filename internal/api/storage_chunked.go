package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"
)

// InitChunkedUploadRequest represents the request body for initializing a chunked upload
type InitChunkedUploadRequest struct {
	Path         string            `json:"path"`
	TotalSize    int64             `json:"total_size"`
	ChunkSize    int64             `json:"chunk_size,omitempty"`
	ContentType  string            `json:"content_type,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	CacheControl string            `json:"cache_control,omitempty"`
}

// ChunkedUploadSessionResponse represents the response for a chunked upload session
type ChunkedUploadSessionResponse struct {
	SessionID       string    `json:"session_id"`
	Bucket          string    `json:"bucket"`
	Path            string    `json:"path"`
	TotalSize       int64     `json:"total_size"`
	ChunkSize       int64     `json:"chunk_size"`
	TotalChunks     int       `json:"total_chunks"`
	CompletedChunks []int     `json:"completed_chunks"`
	Status          string    `json:"status"`
	ExpiresAt       time.Time `json:"expires_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// UploadChunkResponse represents the response after uploading a chunk
type UploadChunkResponse struct {
	ChunkIndex int                          `json:"chunk_index"`
	ETag       string                       `json:"etag,omitempty"`
	Size       int64                        `json:"size"`
	Session    ChunkedUploadSessionResponse `json:"session"`
}

// CompleteChunkedUploadResponse represents the response after completing a chunked upload
type CompleteChunkedUploadResponse struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	FullPath    string `json:"full_path"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type,omitempty"`
}

// InitChunkedUpload initializes a new chunked upload session
// POST /api/v1/storage/:bucket/chunked/init
func (h *StorageHandler) InitChunkedUpload(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	bucket := c.Params("bucket")
	if bucket == "" {
		return SendMissingField(c, "bucket")
	}

	// Parse request body
	var req InitChunkedUploadRequest
	if err := ParseBody(c, &req); err != nil {
		return err
	}

	if req.Path == "" {
		return SendMissingField(c, "path")
	}

	if req.TotalSize <= 0 {
		return SendBadRequest(c, "total_size must be greater than 0", ErrCodeInvalidInput)
	}

	// H-20: sanitize the target path like regular uploads do
	key := sanitizeFilename(req.Path)
	if key == "" {
		return SendBadRequest(c, "invalid path after sanitization", ErrCodeInvalidInput)
	}

	// Default chunk size to 5MB if not specified
	chunkSize := req.ChunkSize
	if chunkSize <= 0 {
		chunkSize = 5 * 1024 * 1024 // 5MB
	}

	// Minimum chunk size is 5MB (S3 requirement for multipart upload, except last part)
	if chunkSize < 5*1024*1024 && req.TotalSize > chunkSize {
		chunkSize = 5 * 1024 * 1024
	}

	// Validate total size
	if err := svc.ValidateUploadSize(req.TotalSize); err != nil {
		return SendErrorWithCode(c, fiber.StatusRequestEntityTooLarge, "File size exceeds upload limit", ErrCodeInvalidInput)
	}

	// F3: apply the same bucket validation as regular uploads (existence,
	// max_file_size, allowed_mime_types) so chunked uploads cannot bypass
	// bucket limits. The effective limit is min(declared total, bucket max).
	contentType := req.ContentType
	if perr := h.validateBucketForUpload(c, bucket, req.TotalSize, contentType); perr != nil {
		return perr.send(c)
	}

	// Get owner ID from authenticated user
	ownerID := getUserID(c)
	var ownerUUID *string
	if ownerID != "" && ownerID != "anonymous" {
		ownerUUID = &ownerID
	}

	ctx := c.RequestCtx()

	// F2: verify BEFORE any bytes are written (chunk files included) that the
	// caller may create or overwrite the destination object. This probe is
	// read-only — no reservation is persisted; Complete re-probes and reserves.
	if perr := h.authorizeUploadIntent(c, bucket, key, contentType, ownerUUID); perr != nil {
		return perr.send(c)
	}

	// Prepare upload options
	opts := &storage.UploadOptions{
		ContentType:  contentType,
		Metadata:     req.Metadata,
		CacheControl: req.CacheControl,
	}

	// Initialize chunked upload with the storage provider
	var session *storage.ChunkedUploadSession
	err = nil

	// Check provider type and call appropriate method
	switch provider := svc.Provider.(type) {
	case *storage.LocalStorage:
		session, err = provider.InitChunkedUpload(ctx, bucket, key, req.TotalSize, chunkSize, opts)
	case *storage.S3Storage:
		session, err = provider.InitChunkedUpload(ctx, bucket, key, req.TotalSize, chunkSize, opts)
	default:
		return SendInternalError(c, "Storage provider does not support chunked uploads")
	}

	if err != nil {
		log.Error().Err(err).Str("bucket", bucket).Str("path", key).Msg("Failed to initialize chunked upload")
		return SendInternalError(c, "Failed to initialize chunked upload")
	}

	session.OwnerID = ownerID

	// Persist the session so the owner (and, for S3, the whole session) survives
	// a reload at completion. For LocalStorage, InitChunkedUpload writes
	// session.json before returning (before OwnerID is set above), so the owner
	// is lost unless re-persisted here — and storeUploadedObject then inserts
	// owner_id = NULL, failing the storage_objects_insert RLS policy. For S3,
	// there is no on-disk store at all, so the session must be created in the
	// storage.chunked_upload_sessions table here.
	if isS3Provider(svc.Provider) {
		if err := h.createChunkedSessionDB(ctx, c, session); err != nil {
			log.Error().Err(err).Str("uploadID", session.UploadID).Msg("Failed to persist S3 chunked upload session")
			return SendInternalError(c, "Failed to initialize chunked upload session")
		}
	} else if err := h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session); err != nil {
		log.Warn().Err(err).Str("uploadID", session.UploadID).Msg("Failed to persist session owner, will fall back to request user at completion")
	}

	log.Info().
		Str("uploadID", session.UploadID).
		Str("bucket", bucket).
		Str("path", req.Path).
		Int64("totalSize", req.TotalSize).
		Int("totalChunks", session.TotalChunks).
		Msg("Chunked upload session initialized")

	return c.Status(fiber.StatusCreated).JSON(ChunkedUploadSessionResponse{
		SessionID:       session.UploadID,
		Bucket:          session.Bucket,
		Path:            session.Key,
		TotalSize:       session.TotalSize,
		ChunkSize:       session.ChunkSize,
		TotalChunks:     session.TotalChunks,
		CompletedChunks: session.CompletedChunks,
		Status:          session.Status,
		ExpiresAt:       session.ExpiresAt,
		CreatedAt:       session.CreatedAt,
	})
}

// UploadChunk uploads a single chunk of a file
// PUT /api/v1/storage/:bucket/chunked/:uploadId/:chunkIndex
func (h *StorageHandler) UploadChunk(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	bucket := c.Params("bucket")
	uploadID := c.Params("uploadId")
	chunkIndexStr := c.Params("chunkIndex")

	if bucket == "" || uploadID == "" || chunkIndexStr == "" {
		return SendBadRequest(c, "Bucket, uploadId, and chunkIndex are required", ErrCodeMissingField)
	}

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		return SendBadRequest(c, "Invalid chunkIndex: must be a non-negative integer", ErrCodeInvalidInput)
	}

	// Get chunk size from Content-Length header
	size := int64(c.Request().Header.ContentLength())
	if size <= 0 {
		return SendBadRequest(c, "Content-Length header is required", ErrCodeMissingField)
	}

	ctx := c.RequestCtx()

	// Retrieve session
	session, err := h.getChunkedUploadSessionFromProvider(ctx, c, svc.Provider, uploadID)
	if err != nil {
		return SendNotFound(c, "Upload session not found")
	}

	// Verify bucket matches
	if session.Bucket != bucket {
		return SendBadRequest(c, "Bucket mismatch", ErrCodeInvalidInput)
	}

	// Check if session is still active
	if session.Status != "active" {
		return SendConflict(c, fmt.Sprintf("Upload session is not active (status: %s)", session.Status), ErrCodeConflict)
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		return SendErrorWithCode(c, fiber.StatusGone, "Upload session has expired", "SESSION_EXPIRED")
	}

	// F3: enforce the per-chunk size cap. Non-final chunks may hold at most
	// session.ChunkSize bytes; the final chunk holds at most the declared
	// remainder, so the assembled object can never exceed the declared total.
	maxChunk := maxChunkSizeFor(session, chunkIndex)
	if maxChunk <= 0 {
		return SendErrorWithCode(c, fiber.StatusBadRequest, "Chunk size exceeds the declared upload total", ErrCodeInvalidInput)
	}
	if size > maxChunk+chunkSizeSlack {
		return SendErrorWithCode(c, fiber.StatusBadRequest,
			fmt.Sprintf("Chunk size %d exceeds the maximum of %d bytes for chunk %d", size, maxChunk, chunkIndex),
			ErrCodeInvalidInput)
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
			return SendBadRequest(c, "Request body is required", ErrCodeMissingField)
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Upload the chunk
	var result *storage.ChunkResult

	switch provider := svc.Provider.(type) {
	case *storage.LocalStorage:
		result, err = provider.UploadChunk(ctx, session, chunkIndex, body, size)
	case *storage.S3Storage:
		result, err = provider.UploadChunk(ctx, session, chunkIndex, body, size)
	default:
		return SendInternalError(c, "Storage provider does not support chunked uploads")
	}

	if err != nil {
		log.Error().Err(err).Str("uploadID", uploadID).Int("chunkIndex", chunkIndex).Msg("Failed to upload chunk")
		return SendInternalError(c, "Failed to upload chunk")
	}

	// F3: the written chunk must respect the same cap (Content-Length can be
	// missing or wrong with chunked transfer encoding).
	if result.Size > maxChunk+chunkSizeSlack {
		return SendErrorWithCode(c, fiber.StatusBadRequest,
			fmt.Sprintf("Chunk size %d exceeds the maximum of %d bytes for chunk %d", result.Size, maxChunk, chunkIndex),
			ErrCodeInvalidInput)
	}

	// Update session with the completed chunk
	session.CompletedChunks = append(session.CompletedChunks, chunkIndex)
	if session.S3PartETags == nil {
		session.S3PartETags = make(map[int]string)
	}
	session.S3PartETags[chunkIndex] = result.ETag

	// Store updated session
	if err := h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session); err != nil {
		log.Warn().Err(err).Str("uploadID", uploadID).Msg("Failed to update session in database")
	}

	log.Debug().
		Str("uploadID", uploadID).
		Int("chunkIndex", chunkIndex).
		Int64("size", result.Size).
		Msg("Chunk uploaded")

	return c.Status(fiber.StatusOK).JSON(UploadChunkResponse{
		ChunkIndex: result.ChunkIndex,
		ETag:       result.ETag,
		Size:       result.Size,
		Session: ChunkedUploadSessionResponse{
			SessionID:       session.UploadID,
			Bucket:          session.Bucket,
			Path:            session.Key,
			TotalSize:       session.TotalSize,
			ChunkSize:       session.ChunkSize,
			TotalChunks:     session.TotalChunks,
			CompletedChunks: session.CompletedChunks,
			Status:          session.Status,
			ExpiresAt:       session.ExpiresAt,
			CreatedAt:       session.CreatedAt,
		},
	})
}

// CompleteChunkedUpload finalizes a chunked upload
// POST /api/v1/storage/:bucket/chunked/:uploadId/complete
func (h *StorageHandler) CompleteChunkedUpload(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	bucket := c.Params("bucket")
	uploadID := c.Params("uploadId")

	if bucket == "" || uploadID == "" {
		return SendBadRequest(c, "Bucket and uploadId are required", ErrCodeMissingField)
	}

	ctx := c.RequestCtx()

	// Retrieve session
	session, err := h.getChunkedUploadSessionFromProvider(ctx, c, svc.Provider, uploadID)
	if err != nil {
		return SendNotFound(c, "Upload session not found")
	}

	// Verify bucket matches
	if session.Bucket != bucket {
		return SendBadRequest(c, "Bucket mismatch", ErrCodeInvalidInput)
	}

	// Check if all chunks are uploaded
	if len(session.CompletedChunks) != session.TotalChunks {
		missingChunks := []int{}
		completedMap := make(map[int]bool)
		for _, idx := range session.CompletedChunks {
			completedMap[idx] = true
		}
		for i := 0; i < session.TotalChunks; i++ {
			if !completedMap[i] {
				missingChunks = append(missingChunks, i)
			}
		}
		return SendErrorWithDetails(c, fiber.StatusBadRequest,
			"Not all chunks have been uploaded", ErrCodeValidationFailed, "", "",
			fiber.Map{
				"missing_chunks": missingChunks,
				"uploaded":       len(session.CompletedChunks),
				"total":          session.TotalChunks,
			})
	}

	// F2: authorize and reserve the destination object BEFORE the assembled
	// bytes overwrite anything in the provider. If a row already exists a
	// no-op UPDATE validates overwrite permission; otherwise a size=0
	// placeholder row is inserted (createdPlaceholder=true) and must be
	// removed if completion fails below.
	var ownerUUID *string
	if ownerID := getUserID(c); ownerID != "" && ownerID != "anonymous" {
		ownerUUID = &ownerID
	}
	probe, perr := h.AuthorizeUploadWrite(c, bucket, session.Key, session.ContentType, nil, ownerUUID)
	if perr != nil {
		return perr.send(c)
	}
	createdPlaceholder := probe.createdPlaceholder

	// Mark session as completing
	session.Status = "completing"
	_ = h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session)

	// Complete the upload
	var object *storage.Object

	switch provider := svc.Provider.(type) {
	case *storage.LocalStorage:
		object, err = provider.CompleteChunkedUpload(ctx, session)
	case *storage.S3Storage:
		object, err = provider.CompleteChunkedUpload(ctx, session)
	default:
		return SendInternalError(c, "Storage provider does not support chunked uploads")
	}

	if err != nil {
		session.Status = "active" // Revert status on failure
		_ = h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session)
		if createdPlaceholder {
			// No destination bytes were written; just remove the reservation.
			h.CleanupUploadPlaceholder(c, svc, bucket, session.Key)
		}
		log.Error().Err(err).Str("uploadID", uploadID).Msg("Failed to complete chunked upload")
		return SendInternalError(c, "Failed to complete chunked upload")
	}

	// F3: the assembled object must match the declared total size exactly.
	if object.Size != session.TotalSize {
		if createdPlaceholder {
			h.CleanupUploadPlaceholder(c, svc, bucket, session.Key)
		} else {
			log.Error().Str("uploadID", uploadID).Int64("expected", session.TotalSize).Int64("actual", object.Size).Msg("Assembled chunked object size mismatch; provider bytes left in place (pre-existing object row intact)")
		}
		session.Status = "active"
		_ = h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session)
		return SendErrorWithCode(c, fiber.StatusBadRequest,
			fmt.Sprintf("Assembled object size %d does not match declared total %d", object.Size, session.TotalSize),
			ErrCodeInvalidInput)
	}

	// Store object record in database
	if err := h.storeUploadedObject(c, session, object); err != nil {
		if createdPlaceholder {
			// We created the placeholder row: remove both the assembled bytes
			// and the reservation — nothing pre-existing is destroyed.
			h.CleanupUploadPlaceholder(c, svc, bucket, session.Key)
		} else {
			// A pre-existing (caller-writable) object row survives; never
			// delete provider bytes in that case — the row must keep pointing
			// at readable data.
			log.Error().Err(err).Str("uploadID", uploadID).Msg("Failed to store object in database; assembled bytes kept for pre-existing object row")
		}
		session.Status = "active"
		_ = h.updateChunkedUploadSessionInProvider(ctx, c, svc.Provider, session)
		log.Warn().Err(err).Str("uploadID", uploadID).Msg("Failed to store object in database")
		return SendInternalError(c, "Failed to store object metadata")
	}

	// Mark session as completed and clean up
	session.Status = "completed"
	_ = h.deleteChunkedUploadSession(ctx, c, svc.Provider, uploadID)

	log.Info().
		Str("uploadID", uploadID).
		Str("bucket", bucket).
		Str("path", session.Key).
		Int64("size", object.Size).
		Msg("Chunked upload completed")

	return c.Status(fiber.StatusOK).JSON(CompleteChunkedUploadResponse{
		ID:          object.ETag,
		Path:        object.Key,
		FullPath:    fmt.Sprintf("%s/%s", object.Bucket, object.Key),
		Size:        object.Size,
		ContentType: object.ContentType,
	})
}

// GetChunkedUploadStatus retrieves the status of a chunked upload session
// GET /api/v1/storage/:bucket/chunked/:uploadId/status
func (h *StorageHandler) GetChunkedUploadStatus(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	bucket := c.Params("bucket")
	uploadID := c.Params("uploadId")

	if bucket == "" || uploadID == "" {
		return SendBadRequest(c, "Bucket and uploadId are required", ErrCodeMissingField)
	}

	// Retrieve session
	session, err := h.getChunkedUploadSessionFromProvider(c.RequestCtx(), c, svc.Provider, uploadID)
	if err != nil {
		return SendNotFound(c, "Upload session not found")
	}

	// Verify bucket matches
	if session.Bucket != bucket {
		return SendBadRequest(c, "Bucket mismatch", ErrCodeInvalidInput)
	}

	// Calculate missing chunks
	missingChunks := []int{}
	completedMap := make(map[int]bool)
	for _, idx := range session.CompletedChunks {
		completedMap[idx] = true
	}
	for i := 0; i < session.TotalChunks; i++ {
		if !completedMap[i] {
			missingChunks = append(missingChunks, i)
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"session": ChunkedUploadSessionResponse{
			SessionID:       session.UploadID,
			Bucket:          session.Bucket,
			Path:            session.Key,
			TotalSize:       session.TotalSize,
			ChunkSize:       session.ChunkSize,
			TotalChunks:     session.TotalChunks,
			CompletedChunks: session.CompletedChunks,
			Status:          session.Status,
			ExpiresAt:       session.ExpiresAt,
			CreatedAt:       session.CreatedAt,
		},
		"missing_chunks": missingChunks,
	})
}

// AbortChunkedUpload aborts a chunked upload and cleans up
// DELETE /api/v1/storage/:bucket/chunked/:uploadId
func (h *StorageHandler) AbortChunkedUpload(c fiber.Ctx) error {
	// Get tenant-specific storage service
	svc, err := h.getService(c)
	if err != nil {
		return SendInternalError(c, "Failed to get storage service")
	}

	bucket := c.Params("bucket")
	uploadID := c.Params("uploadId")

	if bucket == "" || uploadID == "" {
		return SendBadRequest(c, "Bucket and uploadId are required", ErrCodeMissingField)
	}

	ctx := c.RequestCtx()

	// Retrieve session
	session, err := h.getChunkedUploadSessionFromProvider(ctx, c, svc.Provider, uploadID)
	if err != nil {
		return SendNotFound(c, "Upload session not found")
	}

	// Verify bucket matches
	if session.Bucket != bucket {
		return SendBadRequest(c, "Bucket mismatch", ErrCodeInvalidInput)
	}

	// Abort the upload
	switch provider := svc.Provider.(type) {
	case *storage.LocalStorage:
		err = provider.AbortChunkedUpload(ctx, session)
	case *storage.S3Storage:
		err = provider.AbortChunkedUpload(ctx, session)
	default:
		return SendInternalError(c, "Storage provider does not support chunked uploads")
	}

	if err != nil {
		log.Error().Err(err).Str("uploadID", uploadID).Msg("Failed to abort chunked upload")
		return SendInternalError(c, "Failed to abort chunked upload")
	}

	// Delete session from database
	_ = h.deleteChunkedUploadSession(ctx, c, svc.Provider, uploadID)

	log.Info().
		Str("uploadID", uploadID).
		Str("bucket", bucket).
		Msg("Chunked upload aborted")

	return c.SendStatus(fiber.StatusNoContent)
}

// Helper functions for session management

// chunkSizeSlack is the tolerance applied to per-chunk size checks to absorb
// framing overhead differences between the declared Content-Length and the
// actual bytes written.
const chunkSizeSlack int64 = 64 * 1024

// maxChunkSizeFor returns the maximum number of bytes allowed for the given
// chunk index. Non-final chunks are capped at session.ChunkSize; the final
// chunk is capped at the declared remainder so the assembled object can never
// exceed session.TotalSize.
func maxChunkSizeFor(session *storage.ChunkedUploadSession, chunkIndex int) int64 {
	if session == nil || session.ChunkSize <= 0 || chunkIndex < 0 || chunkIndex >= session.TotalChunks {
		return 0
	}
	if chunkIndex < session.TotalChunks-1 {
		return session.ChunkSize
	}
	remaining := session.TotalSize - int64(session.TotalChunks-1)*session.ChunkSize
	if remaining <= 0 || remaining > session.ChunkSize {
		return session.ChunkSize
	}
	return remaining
}

func (h *StorageHandler) getChunkedUploadSessionFromProvider(ctx context.Context, c fiber.Ctx, provider storage.Provider, uploadID string) (*storage.ChunkedUploadSession, error) {
	// Try to get session from storage provider
	switch p := provider.(type) {
	case *storage.LocalStorage:
		return p.GetChunkedUploadSession(uploadID)
	case *storage.S3Storage:
		// S3 has no on-disk session store; sessions live in the
		// storage.chunked_upload_sessions table (under RLS).
		return h.getChunkedSessionDB(ctx, c, uploadID)
	default:
		return nil, fmt.Errorf("storage provider does not support chunked upload sessions")
	}
}

func (h *StorageHandler) updateChunkedUploadSessionInProvider(ctx context.Context, c fiber.Ctx, provider storage.Provider, session *storage.ChunkedUploadSession) error {
	switch p := provider.(type) {
	case *storage.LocalStorage:
		return p.UpdateChunkedUploadSession(session)
	case *storage.S3Storage:
		return h.updateChunkedSessionDB(ctx, c, session)
	default:
		return nil
	}
}

func (h *StorageHandler) deleteChunkedUploadSession(ctx context.Context, c fiber.Ctx, provider storage.Provider, uploadID string) error {
	switch provider.(type) {
	case *storage.LocalStorage:
		// LocalStorage cleans up the on-disk session dir itself on
		// complete/abort (see LocalStorage.CompleteChunkedUpload).
		return nil
	case *storage.S3Storage:
		return h.deleteChunkedSessionDB(ctx, c, uploadID)
	default:
		return nil
	}
}

func (h *StorageHandler) storeUploadedObject(fiberCtx interface{}, session *storage.ChunkedUploadSession, object *storage.Object) error {
	// Store the object record in the database
	// This mirrors the logic in storage_files.go for regular uploads

	// Get fiber context and database pool
	c, ok := fiberCtx.(fiber.Ctx)
	if !ok {
		return fmt.Errorf("invalid context type")
	}

	ctx := c.RequestCtx()

	// Start a transaction to set RLS context
	tx, err := h.getPool(c).Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Set RLS context (includes tenant context for multi-tenancy)
	if err := h.setRLSContext(ctx, tx, c); err != nil {
		return fmt.Errorf("failed to set RLS context: %w", err)
	}

	// Insert object record into storage.objects table
	// Note: 'name' column is auto-generated from 'path', so we don't insert it directly
	// tenant_id is auto-populated by trigger from session context
	query := `
		INSERT INTO storage.objects (bucket_id, path, size, mime_type, metadata, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (bucket_id, path)
		DO UPDATE SET
			size = EXCLUDED.size,
			mime_type = EXCLUDED.mime_type,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
	`

	metadataJSON, _ := json.Marshal(object.Metadata)

	var ownerID interface{} = nil
	if session.OwnerID != "" && session.OwnerID != "anonymous" {
		ownerID = session.OwnerID
	}
	// Fall back to the authenticated user from the live request when the
	// session has no owner. Chunked sessions persist session.json inside
	// InitChunkedUpload — before OwnerID is assigned on the returned struct —
	// so the owner can be lost if the session is reloaded from disk before the
	// init handler re-persists it (see InitChunkedUpload call site). A NULL
	// owner_id fails the storage_objects_insert RLS policy for authenticated
	// users (requires auth.current_user_id() = owner_id). This mirrors the
	// regular upload path in storage_files.go, which reads the user from the
	// live request at insert time.
	if ownerID == nil {
		if liveOwnerID := getUserID(c); liveOwnerID != "" && liveOwnerID != "anonymous" {
			ownerID = liveOwnerID
		}
	}

	_, err = tx.Exec(
		ctx, query,
		object.Bucket,
		object.Key,
		object.Size,
		object.ContentType,
		metadataJSON,
		ownerID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert object: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// fiber:context-methods migrated
