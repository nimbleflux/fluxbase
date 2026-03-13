package ai

import (
	"fmt"
	"path/filepath"

	"github.com/nimbleflux/fluxbase/internal/storage"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

// UserKnowledgeBaseHandler handles user-facing KB endpoints
type UserKnowledgeBaseHandler struct {
	storage        *KnowledgeBaseStorage
	processor      *DocumentProcessor
	storageService *storage.Service
	textExtractor  *TextExtractor
}

// NewUserKnowledgeBaseHandler creates a new user KB handler
func NewUserKnowledgeBaseHandler(storage *KnowledgeBaseStorage) *UserKnowledgeBaseHandler {
	return &UserKnowledgeBaseHandler{
		storage:       storage,
		textExtractor: NewTextExtractor(),
	}
}

// NewUserKnowledgeBaseHandlerWithProcessor creates a handler with document processing support
func NewUserKnowledgeBaseHandlerWithProcessor(storage *KnowledgeBaseStorage, processor *DocumentProcessor) *UserKnowledgeBaseHandler {
	return &UserKnowledgeBaseHandler{
		storage:       storage,
		processor:     processor,
		textExtractor: NewTextExtractor(),
	}
}

// SetStorageService sets the storage service for file uploads
func (h *UserKnowledgeBaseHandler) SetStorageService(svc *storage.Service) {
	h.storageService = svc
}

// ListMyKnowledgeBases returns KBs accessible to current user
// GET /api/v1/ai/knowledge-bases
func (h *UserKnowledgeBaseHandler) ListMyKnowledgeBases(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)

	kbs, err := h.storage.ListUserKnowledgeBases(ctx, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list knowledge bases",
		})
	}

	return c.JSON(fiber.Map{
		"knowledge_bases": kbs,
		"count":           len(kbs),
	})
}

// CreateMyKnowledgeBase creates a user-owned KB
// POST /api/v1/ai/knowledge-bases
func (h *UserKnowledgeBaseHandler) CreateMyKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)

	var req CreateKnowledgeBaseRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name is required",
		})
	}

	// Create KB using the shared method (handles defaults including embedding model)
	kb, err := h.storage.CreateKnowledgeBaseFromRequest(ctx, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create knowledge base",
		})
	}

	// Set owner to current user
	kb.OwnerID = &userID
	if err := h.storage.UpdateKnowledgeBase(ctx, kb); err != nil {
		log.Warn().Err(err).Msg("Failed to set KB owner")
	}

	// Grant initial permissions if specified
	for _, perm := range req.InitialPermissions {
		_, err := h.storage.GrantKBPermission(ctx, kb.ID, perm.UserID, string(perm.Permission), &userID)
		if err != nil {
			// Log error but don't fail the entire request
			// The KB was created successfully, just permission grant failed
			continue
		}
	}

	return c.Status(fiber.StatusCreated).JSON(kb)
}

// GetMyKnowledgeBase returns a specific KB if user has access
// GET /api/v1/ai/knowledge-bases/:id
func (h *UserKnowledgeBaseHandler) GetMyKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	if !h.storage.CanUserAccessKB(ctx, kbID, userID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Knowledge base not found",
		})
	}

	return c.JSON(kb)
}

// ShareKnowledgeBase grants permission to another user
// POST /api/v1/ai/knowledge-bases/:id/share
func (h *UserKnowledgeBaseHandler) ShareKnowledgeBase(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can share knowledge base",
		})
	}

	var req struct {
		UserID     string `json:"user_id"`
		Permission string `json:"permission"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	grant, err := h.storage.GrantKBPermission(ctx, kbID, req.UserID, req.Permission, &userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to grant permission",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(grant)
}

// ListPermissions lists permissions for a KB
// GET /api/v1/ai/knowledge-bases/:id/permissions
func (h *UserKnowledgeBaseHandler) ListPermissions(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can view permissions",
		})
	}

	perms, err := h.storage.ListKBPermissions(ctx, kbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list permissions",
		})
	}

	return c.JSON(perms)
}

// RevokePermission revokes a permission
// DELETE /api/v1/ai/knowledge-bases/:id/permissions/:user_id
func (h *UserKnowledgeBaseHandler) RevokePermission(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	targetUserID := c.Params("user_id")

	kb, err := h.storage.GetKnowledgeBase(ctx, kbID)
	if err != nil || kb.OwnerID == nil || *kb.OwnerID != userID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Only owner can revoke permissions",
		})
	}

	err = h.storage.RevokeKBPermission(ctx, kbID, targetUserID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke permission",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ============================================================================
// USER-FACING DOCUMENT ENDPOINTS
// ============================================================================

// ListMyDocuments lists documents in a KB (requires viewer permission)
// GET /api/v1/ai/knowledge-bases/:id/documents
func (h *UserKnowledgeBaseHandler) ListMyDocuments(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check read permission (viewer or higher)
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Get documents (the storage layer will filter by user's access)
	documents, err := h.storage.ListDocuments(ctx, kbID)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Failed to list documents")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list documents",
		})
	}

	return c.JSON(fiber.Map{
		"documents": documents,
		"count":     len(documents),
	})
}

// GetMyDocument gets a specific document (requires viewer permission)
// GET /api/v1/ai/knowledge-bases/:id/documents/:doc_id
func (h *UserKnowledgeBaseHandler) GetMyDocument(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	docID := c.Params("doc_id")

	// Check read permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	doc, err := h.storage.GetDocument(ctx, docID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Verify document belongs to the KB
	if doc.KnowledgeBaseID != kbID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	return c.JSON(doc)
}

// AddMyDocument adds a document to a KB (requires editor permission)
// POST /api/v1/ai/knowledge-bases/:id/documents
func (h *UserKnowledgeBaseHandler) AddMyDocument(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check write permission (editor or higher)
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionEditor))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Editor permission required to add documents",
		})
	}

	// Check if processor is available
	if h.processor == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Document processing not available (embedding service not configured)",
		})
	}

	var req AddDocumentRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Content == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content is required",
		})
	}

	// Auto-set user_id in metadata for user isolation
	metadata := req.Metadata
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["user_id"] = userID

	// Add document
	docReq := CreateDocumentRequest{
		Title:     req.Title,
		Content:   req.Content,
		SourceURL: req.Source,
		MimeType:  req.MimeType,
		Metadata:  metadata,
	}

	doc, err := h.processor.AddDocument(ctx, kbID, docReq, &userID)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Failed to add document")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add document",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"document_id": doc.ID,
		"status":      "processing",
		"message":     "Document is being processed and will be available shortly",
	})
}

// UploadMyDocument uploads a file to a KB (requires editor permission)
// POST /api/v1/ai/knowledge-bases/:id/documents/upload
func (h *UserKnowledgeBaseHandler) UploadMyDocument(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check write permission (editor or higher)
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionEditor))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Editor permission required to upload documents",
		})
	}

	// Check if processor is available
	if h.processor == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Document processing not available (embedding service not configured)",
		})
	}

	// Get the uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Check file size (max 50MB)
	maxSize := int64(50 * 1024 * 1024)
	if file.Size > maxSize {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("File too large. Maximum size is %dMB", maxSize/(1024*1024)),
		})
	}

	// Determine MIME type from file extension
	ext := filepath.Ext(file.Filename)
	mimeType := GetMimeTypeFromExtension(ext)

	// Check if MIME type is supported
	supported := h.textExtractor.SupportedMimeTypes()
	isSupported := false
	for _, s := range supported {
		if s == mimeType {
			isSupported = true
			break
		}
	}
	if !isSupported {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":           fmt.Sprintf("Unsupported file type: %s", ext),
			"supported_types": supported,
		})
	}

	// Read file content
	fileReader, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read uploaded file",
		})
	}
	defer func() { _ = fileReader.Close() }()

	fileContent, err := readFileContent(fileReader, int(file.Size))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file content",
		})
	}

	// Extract text from file
	extractedText, err := h.textExtractor.Extract(fileContent, mimeType)
	if err != nil {
		log.Error().Err(err).Str("mime_type", mimeType).Msg("Failed to extract text from file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to extract text from file: %v", err),
		})
	}

	// Prepare metadata with user isolation
	metadata := map[string]string{"user_id": userID}

	// Create document request
	docReq := CreateDocumentRequest{
		Title:    file.Filename,
		Content:  extractedText,
		MimeType: mimeType,
		Metadata: metadata,
	}

	// Add document
	doc, err := h.processor.AddDocument(ctx, kbID, docReq, &userID)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Failed to add document from upload")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add document",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"document_id": doc.ID,
		"status":      "processing",
		"message":     "Document is being processed and will be available shortly",
	})
}

// DeleteMyDocument deletes a document from a KB (requires editor permission)
// DELETE /api/v1/ai/knowledge-bases/:id/documents/:doc_id
func (h *UserKnowledgeBaseHandler) DeleteMyDocument(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	docID := c.Params("doc_id")

	// Check write permission (editor or higher)
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionEditor))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Editor permission required to delete documents",
		})
	}

	// Get document to verify it belongs to this KB
	doc, err := h.storage.GetDocument(ctx, docID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}
	if doc.KnowledgeBaseID != kbID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Delete document
	if err := h.storage.DeleteDocument(ctx, docID); err != nil {
		log.Error().Err(err).Str("doc_id", docID).Msg("Failed to delete document")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete document",
		})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// SearchMyKB searches a knowledge base (requires viewer permission)
// POST /api/v1/ai/knowledge-bases/:id/search
func (h *UserKnowledgeBaseHandler) SearchMyKB(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check read permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	var req SearchRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query is required",
		})
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}

	// Perform search using hybrid search (keyword-only if embeddings not available)
	opts := HybridSearchOptions{
		Query: req.Query,
		Limit: req.Limit,
		Mode:  SearchModeKeyword, // Default to keyword search for user endpoint
	}

	// If processor has embedding service, use hybrid search
	if h.processor != nil && h.processor.embeddingService != nil {
		embedding, err := h.processor.embeddingService.EmbedSingle(ctx, req.Query, "")
		if err == nil && len(embedding) > 0 {
			opts.QueryEmbedding = embedding
			opts.Mode = SearchModeHybrid
			opts.SemanticWeight = 0.7
		}
	}

	results, err := h.storage.SearchChunksHybrid(ctx, kbID, opts)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Search failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"query":   req.Query,
		"limit":   req.Limit,
		"count":   len(results),
	})
}

// readFileContent reads file content from reader with size limit
func readFileContent(reader interface{ Read([]byte) (int, error) }, maxSize int) ([]byte, error) {
	size := maxSize
	if size > 50*1024*1024 {
		size = 50 * 1024 * 1024 // Cap at 50MB
	}
	buf := make([]byte, 0, size)
	tmp := make([]byte, 1024)
	for {
		n, err := reader.Read(tmp)
		if err != nil {
			break
		}
		buf = append(buf, tmp[:n]...)
		if len(buf) > size {
			return nil, fmt.Errorf("file too large")
		}
	}
	return buf, nil
}

// RegisterUserKnowledgeBaseRoutes registers user-facing routes
func RegisterUserKnowledgeBaseRoutes(router fiber.Router, storage *KnowledgeBaseStorage) {
	handler := NewUserKnowledgeBaseHandler(storage)
	router.Get("/knowledge-bases", handler.ListMyKnowledgeBases)
	router.Post("/knowledge-bases", handler.CreateMyKnowledgeBase)
	router.Get("/knowledge-bases/:id", handler.GetMyKnowledgeBase)
	router.Post("/knowledge-bases/:id/share", handler.ShareKnowledgeBase)
	router.Get("/knowledge-bases/:id/permissions", handler.ListPermissions)
	router.Delete("/knowledge-bases/:id/permissions/:user_id", handler.RevokePermission)
}

// RegisterUserKnowledgeBaseRoutesWithDocuments registers user-facing routes including document operations
func RegisterUserKnowledgeBaseRoutesWithDocuments(router fiber.Router, storage *KnowledgeBaseStorage, processor *DocumentProcessor) {
	handler := NewUserKnowledgeBaseHandlerWithProcessor(storage, processor)

	// KB management routes
	router.Get("/knowledge-bases", handler.ListMyKnowledgeBases)
	router.Post("/knowledge-bases", handler.CreateMyKnowledgeBase)
	router.Get("/knowledge-bases/:id", handler.GetMyKnowledgeBase)
	router.Post("/knowledge-bases/:id/share", handler.ShareKnowledgeBase)
	router.Get("/knowledge-bases/:id/permissions", handler.ListPermissions)
	router.Delete("/knowledge-bases/:id/permissions/:user_id", handler.RevokePermission)

	// Document routes (permission checks are in handlers)
	router.Get("/knowledge-bases/:id/documents", handler.ListMyDocuments)
	router.Get("/knowledge-bases/:id/documents/:doc_id", handler.GetMyDocument)
	router.Post("/knowledge-bases/:id/documents", handler.AddMyDocument)
	router.Post("/knowledge-bases/:id/documents/upload", handler.UploadMyDocument)
	router.Delete("/knowledge-bases/:id/documents/:doc_id", handler.DeleteMyDocument)

	// Search route
	router.Post("/knowledge-bases/:id/search", handler.SearchMyKB)
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}
