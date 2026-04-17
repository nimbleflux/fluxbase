package ai

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"
)

// UserKnowledgeBaseHandler handles user-facing KB endpoints
type UserKnowledgeBaseHandler struct {
	storage        *KnowledgeBaseStorage
	knowledgeGraph *KnowledgeGraph
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

// NewUserKnowledgeBaseHandlerWithGraph creates a handler with knowledge graph support
func NewUserKnowledgeBaseHandlerWithGraph(storage *KnowledgeBaseStorage, kg *KnowledgeGraph) *UserKnowledgeBaseHandler {
	return &UserKnowledgeBaseHandler{
		storage:        storage,
		knowledgeGraph: kg,
		textExtractor:  NewTextExtractor(),
	}
}

// NewUserKnowledgeBaseHandlerWithProcessorAndGraph creates a handler with both processor and graph support
func NewUserKnowledgeBaseHandlerWithProcessorAndGraph(storage *KnowledgeBaseStorage, processor *DocumentProcessor, kg *KnowledgeGraph) *UserKnowledgeBaseHandler {
	return &UserKnowledgeBaseHandler{
		storage:        storage,
		knowledgeGraph: kg,
		processor:      processor,
		textExtractor:  NewTextExtractor(),
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

	// Safely check if user_id exists in context
	userIDRaw := c.Locals("user_id")
	if userIDRaw == nil {
		// Check if user is instance admin (service role or instance_admin role)
		userRole := c.Locals("user_role")
		if userRole == "instance_admin" || userRole == "service_role" || userRole == "tenant_service" {
			// Instance admin without tenant context - return empty list
			// A complete solution would fetch all KBs across tenants with tenant info
			return c.JSON(fiber.Map{
				"knowledge_bases": []interface{}{},
				"count":           0,
				"message":         "Select a tenant to view knowledge bases",
			})
		}
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID, ok := userIDRaw.(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user context",
		})
	}

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

// UpdateMyDocument updates a document's metadata
// PATCH /api/v1/ai/knowledge-bases/:id/documents/:doc_id
func (h *UserKnowledgeBaseHandler) UpdateMyDocument(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	docID := c.Params("doc_id")

	// Check editor permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionEditor))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Editor permission required",
		})
	}

	var req struct {
		Title    *string           `json:"title,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
		Tags     []string          `json:"tags,omitempty"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get existing document
	doc, err := h.storage.GetDocument(ctx, docID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Verify document belongs to KB
	if doc.KnowledgeBaseID != kbID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Document not found",
		})
	}

	// Use UpdateDocumentMetadata for updating
	updatedDoc, err := h.storage.UpdateDocumentMetadata(ctx, docID, req.Title, req.Metadata, req.Tags)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update document",
		})
	}

	return c.JSON(updatedDoc)
}

// DeleteMyDocumentsByFilter deletes documents matching a filter
// POST /api/v1/ai/knowledge-bases/:id/documents/delete-by-filter
func (h *UserKnowledgeBaseHandler) DeleteMyDocumentsByFilter(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check editor permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionEditor))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Editor permission required",
		})
	}

	var req struct {
		Tags     []string          `json:"tags,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	filter := &MetadataFilter{
		Tags:     req.Tags,
		Metadata: req.Metadata,
	}

	deletedCount, err := h.storage.DeleteDocumentsByFilter(ctx, kbID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete documents",
		})
	}

	return c.JSON(fiber.Map{
		"deleted_count": deletedCount,
	})
}

// DebugSearchMyKB performs a debug search with detailed diagnostic information
// POST /api/v1/ai/knowledge-bases/:id/debug-search
func (h *UserKnowledgeBaseHandler) DebugSearchMyKB(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	var req struct {
		Query string `json:"query"`
	}
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

	// Perform search with debug info
	opts := HybridSearchOptions{
		Query:          req.Query,
		Limit:          10,
		SemanticWeight: 0.7,
	}

	results, err := h.storage.SearchChunksHybrid(ctx, kbID, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	// Get KB info for context
	kb, _ := h.storage.GetKnowledgeBase(ctx, kbID)

	return c.JSON(fiber.Map{
		"query":          req.Query,
		"results":        results,
		"result_count":   len(results),
		"search_options": opts,
		"knowledge_base": fiber.Map{
			"id":   kbID,
			"name": kb.Name,
		},
		"debug_info": fiber.Map{
			"search_type":      "hybrid",
			"semantic_weight":  opts.SemanticWeight,
			"keyword_weight":   1 - opts.SemanticWeight,
			"embedding_status": "available",
		},
	})
}

// ListMyEntities lists entities in a knowledge base
// GET /api/v1/ai/knowledge-bases/:id/entities
func (h *UserKnowledgeBaseHandler) ListMyEntities(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	// Check if knowledge graph is available
	if h.knowledgeGraph == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Knowledge graph features are not available",
		})
	}

	// Parse optional entity_type filter
	entityTypeStr := c.Query("entity_type")
	var entityType *EntityType
	if entityTypeStr != "" {
		et := EntityType(entityTypeStr)
		entityType = &et
	}

	// Get entities
	entities, err := h.knowledgeGraph.ListEntities(ctx, kbID, entityType)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Failed to list entities")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list entities",
		})
	}

	return c.JSON(fiber.Map{
		"entities": entities,
		"count":    len(entities),
	})
}

// SearchMyEntities searches entities in a knowledge base
// GET /api/v1/ai/knowledge-bases/:id/entities/search
func (h *UserKnowledgeBaseHandler) SearchMyEntities(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	// Check if knowledge graph is available
	if h.knowledgeGraph == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Knowledge graph features are not available",
		})
	}

	// Get query from URL param
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Parse optional entity types filter
	var entityTypes []EntityType
	if typeStr := c.Query("entity_types"); typeStr != "" {
		for _, t := range splitCommaSeparated(typeStr) {
			entityTypes = append(entityTypes, EntityType(t))
		}
	}

	// Parse limit
	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseIntParam(limitStr, 1, 100); err == nil {
			limit = l
		}
	}

	// Search entities
	entities, err := h.knowledgeGraph.SearchEntities(ctx, kbID, query, entityTypes, limit)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Str("query", query).Msg("Failed to search entities")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to search entities",
		})
	}

	return c.JSON(fiber.Map{
		"entities": entities,
		"query":    query,
		"count":    len(entities),
	})
}

// GetMyEntityRelationships gets relationships for an entity
// GET /api/v1/ai/knowledge-bases/:id/entities/:entity_id/relationships
func (h *UserKnowledgeBaseHandler) GetMyEntityRelationships(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")
	entityID := c.Params("entity_id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	// Check if knowledge graph is available
	if h.knowledgeGraph == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Knowledge graph features are not available",
		})
	}

	// Get relationships for the entity
	relationships, err := h.knowledgeGraph.GetRelationships(ctx, kbID, entityID)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Str("entity_id", entityID).Msg("Failed to get entity relationships")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get entity relationships",
		})
	}

	return c.JSON(fiber.Map{
		"relationships": relationships,
		"entity_id":     entityID,
		"count":         len(relationships),
	})
}

// GetMyKnowledgeGraph gets the full knowledge graph
// GET /api/v1/ai/knowledge-bases/:id/graph
func (h *UserKnowledgeBaseHandler) GetMyKnowledgeGraph(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	// Check if knowledge graph is available
	if h.knowledgeGraph == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Knowledge graph features are not available",
		})
	}

	// Get all entities
	entities, err := h.knowledgeGraph.ListEntities(ctx, kbID, nil)
	if err != nil {
		log.Error().Err(err).Str("kb_id", kbID).Msg("Failed to list entities for graph")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get knowledge graph",
		})
	}

	// Get relationships for each entity and collect unique ones
	allRelationships := make(map[string]EntityRelationship)
	for _, entity := range entities {
		relationships, err := h.knowledgeGraph.GetRelationships(ctx, kbID, entity.ID)
		if err != nil {
			log.Warn().Err(err).Str("entity_id", entity.ID).Msg("Failed to get relationships for entity")
			continue
		}
		for _, rel := range relationships {
			allRelationships[rel.ID] = rel
		}
	}

	// Convert map to slice
	relationships := make([]EntityRelationship, 0, len(allRelationships))
	for _, rel := range allRelationships {
		relationships = append(relationships, rel)
	}

	return c.JSON(fiber.Map{
		"knowledge_base_id":  kbID,
		"entities":           entities,
		"relationships":      relationships,
		"entity_count":       len(entities),
		"relationship_count": len(relationships),
	})
}

// ListMyLinkedChatbots lists chatbots linked to a knowledge base
// GET /api/v1/ai/knowledge-bases/:id/chatbots
func (h *UserKnowledgeBaseHandler) ListMyLinkedChatbots(c fiber.Ctx) error {
	ctx := c.RequestCtx()
	userID := c.Locals("user_id").(string)
	kbID := c.Params("id")

	// Check viewer permission
	hasPermission, err := h.storage.CheckKBPermission(ctx, kbID, userID, string(KBPermissionViewer))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check permission",
		})
	}
	if !hasPermission {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Viewer permission required",
		})
	}

	// Get linked chatbots
	links, err := h.storage.GetKnowledgeBaseChatbots(ctx, kbID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get linked chatbots",
		})
	}

	return c.JSON(fiber.Map{
		"chatbots": links,
		"count":    len(links),
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

// splitCommaSeparated splits a comma-separated string into trimmed parts
func splitCommaSeparated(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseIntParam parses an integer parameter with min/max bounds
func parseIntParam(s string, min, max int) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if val < min {
		return min, nil
	}
	if val > max {
		return max, nil
	}
	return val, nil
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}
