package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

// KnowledgeBaseStorage handles database operations for knowledge bases
type KnowledgeBaseStorage struct {
	db *database.Connection
}

// NewKnowledgeBaseStorage creates a new knowledge base storage
func NewKnowledgeBaseStorage(db *database.Connection) *KnowledgeBaseStorage {
	return &KnowledgeBaseStorage{db: db}
}

// ============================================================================
// Knowledge Base CRUD
// ============================================================================

// CreateKnowledgeBase creates a new knowledge base
func (s *KnowledgeBaseStorage) CreateKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error {
	if kb.ID == "" {
		kb.ID = uuid.New().String()
	}
	kb.CreatedAt = time.Now()
	kb.UpdatedAt = time.Now()

	query := `
		INSERT INTO ai.knowledge_bases (
			id, name, namespace, description,
			embedding_model, embedding_dimensions,
			chunk_size, chunk_overlap, chunk_strategy,
			enabled, source, created_by, visibility, owner_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRow(ctx, query,
		kb.ID, kb.Name, kb.Namespace, kb.Description,
		kb.EmbeddingModel, kb.EmbeddingDimensions,
		kb.ChunkSize, kb.ChunkOverlap, kb.ChunkStrategy,
		kb.Enabled, kb.Source, kb.CreatedBy, kb.Visibility, kb.OwnerID,
	).Scan(&kb.CreatedAt, &kb.UpdatedAt)
}

// GetKnowledgeBase retrieves a knowledge base by ID
func (s *KnowledgeBaseStorage) GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBase, error) {
	query := `
		SELECT id, name, namespace, description,
			embedding_model, embedding_dimensions,
			chunk_size, chunk_overlap, chunk_strategy,
			enabled, document_count, total_chunks,
			source, created_by, created_at, updated_at,
			visibility, owner_id
		FROM ai.knowledge_bases
		WHERE id = $1
	`

	var kb KnowledgeBase
	err := s.db.QueryRow(ctx, query, id).Scan(
		&kb.ID, &kb.Name, &kb.Namespace, &kb.Description,
		&kb.EmbeddingModel, &kb.EmbeddingDimensions,
		&kb.ChunkSize, &kb.ChunkOverlap, &kb.ChunkStrategy,
		&kb.Enabled, &kb.DocumentCount, &kb.TotalChunks,
		&kb.Source, &kb.CreatedBy, &kb.CreatedAt, &kb.UpdatedAt,
		&kb.Visibility, &kb.OwnerID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base: %w", err)
	}
	return &kb, nil
}

// GetKnowledgeBaseByName retrieves a knowledge base by name and namespace
func (s *KnowledgeBaseStorage) GetKnowledgeBaseByName(ctx context.Context, name, namespace string) (*KnowledgeBase, error) {
	query := `
		SELECT id, name, namespace, description,
			embedding_model, embedding_dimensions,
			chunk_size, chunk_overlap, chunk_strategy,
			enabled, document_count, total_chunks,
			source, created_by, created_at, updated_at, visibility
		FROM ai.knowledge_bases
		WHERE name = $1 AND namespace = $2
	`

	var kb KnowledgeBase
	err := s.db.QueryRow(ctx, query, name, namespace).Scan(
		&kb.ID, &kb.Name, &kb.Namespace, &kb.Description,
		&kb.EmbeddingModel, &kb.EmbeddingDimensions,
		&kb.ChunkSize, &kb.ChunkOverlap, &kb.ChunkStrategy,
		&kb.Enabled, &kb.DocumentCount, &kb.TotalChunks,
		&kb.Source, &kb.CreatedBy, &kb.CreatedAt, &kb.UpdatedAt, &kb.Visibility,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base by name: %w", err)
	}
	return &kb, nil
}

// ListKnowledgeBases lists knowledge bases with optional filtering
func (s *KnowledgeBaseStorage) ListKnowledgeBases(ctx context.Context, namespace string, enabledOnly bool) ([]KnowledgeBase, error) {
	query := `
		SELECT id, name, namespace, description,
			embedding_model, embedding_dimensions,
			chunk_size, chunk_overlap, chunk_strategy,
			enabled, document_count, total_chunks,
			source, created_by, created_at, updated_at
		FROM ai.knowledge_bases
		WHERE ($1 = '' OR namespace = $1)
		  AND ($2 = false OR enabled = true)
		ORDER BY namespace, name
	`

	rows, err := s.db.Query(ctx, query, namespace, enabledOnly)
	if err != nil {
		return nil, fmt.Errorf("failed to list knowledge bases: %w", err)
	}
	defer rows.Close()

	var kbs []KnowledgeBase
	for rows.Next() {
		var kb KnowledgeBase
		if err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Namespace, &kb.Description,
			&kb.EmbeddingModel, &kb.EmbeddingDimensions,
			&kb.ChunkSize, &kb.ChunkOverlap, &kb.ChunkStrategy,
			&kb.Enabled, &kb.DocumentCount, &kb.TotalChunks,
			&kb.Source, &kb.CreatedBy, &kb.CreatedAt, &kb.UpdatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan knowledge base row")
			continue
		}
		kbs = append(kbs, kb)
	}

	return kbs, nil
}

// UpdateKnowledgeBase updates a knowledge base
func (s *KnowledgeBaseStorage) UpdateKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error {
	query := `
		UPDATE ai.knowledge_bases SET
			name = $2,
			description = $3,
			embedding_model = $4,
			embedding_dimensions = $5,
			chunk_size = $6,
			chunk_overlap = $7,
			chunk_strategy = $8,
			enabled = $9,
			visibility = $10,
			created_by = $11,
			owner_id = $12,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	return s.db.QueryRow(ctx, query,
		kb.ID, kb.Name, kb.Description,
		kb.EmbeddingModel, kb.EmbeddingDimensions,
		kb.ChunkSize, kb.ChunkOverlap, kb.ChunkStrategy,
		kb.Enabled, kb.Visibility, kb.CreatedBy, kb.OwnerID,
	).Scan(&kb.UpdatedAt)
}

// DeleteKnowledgeBase deletes a knowledge base and all its documents/chunks
func (s *KnowledgeBaseStorage) DeleteKnowledgeBase(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM ai.knowledge_bases WHERE id = $1", id)
	return err
}

// ============================================================================
// Document CRUD
// ============================================================================

// CreateDocument creates a new document in a knowledge base
func (s *KnowledgeBaseStorage) CreateDocument(ctx context.Context, doc *Document) error {
	if doc.ID == "" {
		doc.ID = uuid.New().String()
	}
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.Status = DocumentStatusPending

	// Marshal metadata if present
	var metadataJSON []byte
	if doc.Metadata != nil {
		metadataJSON = doc.Metadata
	}

	query := `
		INSERT INTO ai.documents (
			id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, metadata, tags, created_by, owner_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRow(ctx, query,
		doc.ID, doc.KnowledgeBaseID, doc.Title, doc.SourceURL, doc.SourceType,
		doc.MimeType, doc.Content, doc.ContentHash, doc.Status, metadataJSON, doc.Tags, doc.CreatedBy, doc.OwnerID,
	).Scan(&doc.CreatedAt, &doc.UpdatedAt)
}

// GetDocument retrieves a document by ID
func (s *KnowledgeBaseStorage) GetDocument(ctx context.Context, id string) (*Document, error) {
	query := `
		SELECT id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, error_message,
			chunks_count, metadata, tags, created_by, created_at, updated_at, indexed_at
		FROM ai.documents
		WHERE id = $1
	`

	var doc Document
	err := s.db.QueryRow(ctx, query, id).Scan(
		&doc.ID, &doc.KnowledgeBaseID, &doc.Title, &doc.SourceURL, &doc.SourceType,
		&doc.MimeType, &doc.Content, &doc.ContentHash, &doc.Status, &doc.ErrorMessage,
		&doc.ChunksCount, &doc.Metadata, &doc.Tags, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.IndexedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	return &doc, nil
}

// ListDocuments lists documents in a knowledge base
func (s *KnowledgeBaseStorage) ListDocuments(ctx context.Context, knowledgeBaseID string) ([]Document, error) {
	query := `
		SELECT id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, error_message,
			chunks_count, metadata, tags, created_by, created_at, updated_at, indexed_at
		FROM ai.documents
		WHERE knowledge_base_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(ctx, query, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(
			&doc.ID, &doc.KnowledgeBaseID, &doc.Title, &doc.SourceURL, &doc.SourceType,
			&doc.MimeType, &doc.Content, &doc.ContentHash, &doc.Status, &doc.ErrorMessage,
			&doc.ChunksCount, &doc.Metadata, &doc.Tags, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.IndexedAt,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan document row")
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// UpdateDocumentStatus updates a document's processing status
func (s *KnowledgeBaseStorage) UpdateDocumentStatus(ctx context.Context, id string, status DocumentStatus, errorMsg string) error {
	query := `
		UPDATE ai.documents SET
			status = $2, error_message = $3, updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, id, status, errorMsg)
	return err
}

// MarkDocumentIndexed marks a document as indexed
func (s *KnowledgeBaseStorage) MarkDocumentIndexed(ctx context.Context, id string) error {
	query := `
		UPDATE ai.documents SET
			status = 'indexed', indexed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, id)
	return err
}

// DeleteDocument deletes a document and its chunks
func (s *KnowledgeBaseStorage) DeleteDocument(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM ai.documents WHERE id = $1", id)
	return err
}

// DeleteDocumentsByFilter deletes documents matching the given metadata filter
// Returns the number of documents deleted
func (s *KnowledgeBaseStorage) DeleteDocumentsByFilter(
	ctx context.Context,
	knowledgeBaseID string,
	filter *MetadataFilter,
) (int, error) {
	// Build WHERE clause for filtering
	whereConditions := []string{
		"knowledge_base_id = $1",
	}
	args := []interface{}{knowledgeBaseID}
	argIndex := 2

	// User isolation filter
	if filter != nil && filter.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf(`(
			metadata->>'user_id' = $%d OR
			metadata->>'user_id' IS NULL OR
			NOT (metadata ? 'user_id')
		)`, argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Tag filter - documents must have ALL specified tags
	if filter != nil && len(filter.Tags) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("tags @> $%d", argIndex))
		args = append(args, filter.Tags)
		argIndex++
	}

	// Advanced metadata filter with operators and logical combinations
	if filter != nil && filter.AdvancedFilter != nil {
		// We need to use 'd' as the table alias for consistency, but here we're querying documents directly
		// So we need to adjust the SQL builder or prefix the table name
		metadataSQL, metadataArgs, err := buildMetadataFilterSQLForTable(*filter.AdvancedFilter, &argIndex, "")
		if err != nil {
			return 0, fmt.Errorf("failed to build metadata filter: %w", err)
		}
		if metadataSQL != "" {
			whereConditions = append(whereConditions, metadataSQL)
			args = append(args, metadataArgs...)
		}
	}

	// Legacy simple metadata filter (exact match only)
	if filter != nil && filter.AdvancedFilter == nil && len(filter.Metadata) > 0 {
		for key, value := range filter.Metadata {
			escapedKey := escapeStringLiteral(key)
			whereConditions = append(whereConditions, fmt.Sprintf("metadata->>'%s' = $%d", escapedKey, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf("DELETE FROM ai.documents WHERE %s", whereClause)

	result, err := s.db.Exec(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete documents by filter: %w", err)
	}

	rowsAffected := result.RowsAffected()
	return int(rowsAffected), nil
}

// UpdateDocumentMetadata updates a document's title, metadata, and tags
func (s *KnowledgeBaseStorage) UpdateDocumentMetadata(ctx context.Context, id string, title *string, metadata map[string]string, tags []string) (*Document, error) {
	// Build the metadata JSON
	var metadataJSON []byte
	if metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE ai.documents SET
			title = COALESCE($2, title),
			metadata = COALESCE($3, metadata),
			tags = COALESCE($4, tags),
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, error_message,
			chunks_count, metadata, tags, created_by, created_at, updated_at, indexed_at
	`

	var doc Document
	err := s.db.QueryRow(ctx, query, id, title, metadataJSON, tags).Scan(
		&doc.ID, &doc.KnowledgeBaseID, &doc.Title, &doc.SourceURL, &doc.SourceType,
		&doc.MimeType, &doc.Content, &doc.ContentHash, &doc.Status, &doc.ErrorMessage,
		&doc.ChunksCount, &doc.Metadata, &doc.Tags, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.IndexedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return &doc, nil
}

// FindDocumentByMetadata finds a document by knowledge base ID and metadata fields
// This is used for idempotent operations like table exports where we want to find
// an existing document for a specific schema.table combination.
func (s *KnowledgeBaseStorage) FindDocumentByMetadata(ctx context.Context, knowledgeBaseID string, metadata map[string]string) (*Document, error) {
	if len(metadata) == 0 {
		return nil, fmt.Errorf("at least one metadata field is required")
	}

	// Build WHERE conditions for each metadata field
	argIndex := 1
	var whereConditions []string
	var args []interface{}

	for key, value := range metadata {
		// Escape key to prevent SQL injection
		escapedKey := strings.ReplaceAll(key, "'", "''")
		whereConditions = append(whereConditions, fmt.Sprintf("metadata->>'%s' = $%d", escapedKey, argIndex))
		args = append(args, value)
		argIndex++
	}

	// Add knowledge base ID filter
	whereConditions = append(whereConditions, fmt.Sprintf("knowledge_base_id = $%d", argIndex))
	args = append(args, knowledgeBaseID)

	query := fmt.Sprintf(`
		SELECT id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, error_message,
			chunks_count, metadata, tags, created_by, created_at, updated_at, indexed_at
		FROM ai.documents
		WHERE %s
		LIMIT 1
	`, strings.Join(whereConditions, " AND "))

	var doc Document
	err := s.db.QueryRow(ctx, query, args...).Scan(
		&doc.ID, &doc.KnowledgeBaseID, &doc.Title, &doc.SourceURL, &doc.SourceType,
		&doc.MimeType, &doc.Content, &doc.ContentHash, &doc.Status, &doc.ErrorMessage,
		&doc.ChunksCount, &doc.Metadata, &doc.Tags, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.IndexedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find document by metadata: %w", err)
	}

	return &doc, nil
}

// UpdateDocumentContent updates a document's content, title, and metadata
// This is used for idempotent operations like re-exporting tables.
func (s *KnowledgeBaseStorage) UpdateDocumentContent(ctx context.Context, id string, content string, title string, metadataJSON []byte) error {
	// Calculate new content hash
	contentHash := hashContent(content)

	query := `
		UPDATE ai.documents SET
			content = $2,
			content_hash = $3,
			title = $4,
			metadata = COALESCE($5, metadata),
			status = 'pending',
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.Exec(ctx, query, id, content, contentHash, title, metadataJSON)
	if err != nil {
		return fmt.Errorf("failed to update document content: %w", err)
	}

	return nil
}

// ============================================================================
// Chunk Operations
// ============================================================================

// CreateChunks creates multiple chunks for a document (batch insert)
func (s *KnowledgeBaseStorage) CreateChunks(ctx context.Context, chunks []Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	// Use COPY for efficient bulk insert
	batch := &pgx.Batch{}
	for _, chunk := range chunks {
		if chunk.ID == "" {
			chunk.ID = uuid.New().String()
		}

		var metadataJSON []byte
		if chunk.Metadata != nil {
			metadataJSON = chunk.Metadata
		}

		// Format embedding as PostgreSQL vector literal (pgx can't encode []float32 directly)
		var embeddingExpr string
		if chunk.Embedding != nil {
			embeddingExpr = fmt.Sprintf("'%s'::vector", formatEmbeddingLiteral(chunk.Embedding))
		} else {
			embeddingExpr = "NULL"
		}

		query := fmt.Sprintf(`
			INSERT INTO ai.chunks (
				id, document_id, knowledge_base_id, content,
				chunk_index, start_offset, end_offset, token_count,
				embedding, metadata
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, %s, $9)
		`, embeddingExpr)

		batch.Queue(query,
			chunk.ID, chunk.DocumentID, chunk.KnowledgeBaseID, chunk.Content,
			chunk.ChunkIndex, chunk.StartOffset, chunk.EndOffset, chunk.TokenCount,
			metadataJSON,
		)
	}

	br := s.db.Pool().SendBatch(ctx, batch)
	defer func() { _ = br.Close() }()

	for range chunks {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("failed to insert chunk: %w", err)
		}
	}

	return nil
}

// GetChunksByDocument retrieves all chunks for a document
func (s *KnowledgeBaseStorage) GetChunksByDocument(ctx context.Context, documentID string) ([]Chunk, error) {
	query := `
		SELECT id, document_id, knowledge_base_id, content,
			chunk_index, start_offset, end_offset, token_count, metadata, created_at
		FROM ai.chunks
		WHERE document_id = $1
		ORDER BY chunk_index
	`

	rows, err := s.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunks: %w", err)
	}
	defer rows.Close()

	var chunks []Chunk
	for rows.Next() {
		var chunk Chunk
		if err := rows.Scan(
			&chunk.ID, &chunk.DocumentID, &chunk.KnowledgeBaseID, &chunk.Content,
			&chunk.ChunkIndex, &chunk.StartOffset, &chunk.EndOffset, &chunk.TokenCount,
			&chunk.Metadata, &chunk.CreatedAt,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan chunk row")
			continue
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// DeleteChunksByDocument deletes all chunks for a document
func (s *KnowledgeBaseStorage) DeleteChunksByDocument(ctx context.Context, documentID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM ai.chunks WHERE document_id = $1", documentID)
	return err
}

// ============================================================================
// Chatbot Knowledge Base Links
// ============================================================================

// LinkChatbotKnowledgeBase links a chatbot to a knowledge base
func (s *KnowledgeBaseStorage) LinkChatbotKnowledgeBase(ctx context.Context, link *ChatbotKnowledgeBase) error {
	if link.ID == "" {
		link.ID = uuid.New().String()
	}
	link.CreatedAt = time.Now()
	link.UpdatedAt = time.Now()

	// Set defaults
	if link.AccessLevel == "" {
		link.AccessLevel = "full"
	}
	if link.ContextWeight == 0 {
		link.ContextWeight = 1.0
	}
	if link.Priority == 0 {
		link.Priority = 100
	}

	query := `
		INSERT INTO ai.chatbot_knowledge_bases (
			id, chatbot_id, knowledge_base_id,
			access_level, filter_expression, context_weight, priority,
			intent_keywords, max_chunks, similarity_threshold,
			enabled, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (chatbot_id, knowledge_base_id) DO UPDATE SET
			access_level = EXCLUDED.access_level,
			filter_expression = EXCLUDED.filter_expression,
			context_weight = EXCLUDED.context_weight,
			priority = EXCLUDED.priority,
			intent_keywords = EXCLUDED.intent_keywords,
			max_chunks = EXCLUDED.max_chunks,
			similarity_threshold = EXCLUDED.similarity_threshold,
			enabled = EXCLUDED.enabled,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING created_at, updated_at
	`

	return s.db.QueryRow(ctx, query,
		link.ID, link.ChatbotID, link.KnowledgeBaseID,
		link.AccessLevel, link.FilterExpression, link.ContextWeight, link.Priority,
		link.IntentKeywords, link.MaxChunks, link.SimilarityThreshold,
		link.Enabled, link.Metadata,
	).Scan(&link.CreatedAt, &link.UpdatedAt)
}

// GetChatbotKnowledgeBases retrieves all knowledge base links for a chatbot
func (s *KnowledgeBaseStorage) GetChatbotKnowledgeBases(ctx context.Context, chatbotID string) ([]ChatbotKnowledgeBase, error) {
	query := `
		SELECT ckb.id, ckb.chatbot_id, ckb.knowledge_base_id,
			ckb.access_level, ckb.filter_expression, ckb.context_weight,
			ckb.priority, ckb.intent_keywords, ckb.max_chunks,
			ckb.similarity_threshold, ckb.enabled, ckb.metadata,
			ckb.created_at, ckb.updated_at,
			kb.name as knowledge_base_name
		FROM ai.chatbot_knowledge_bases ckb
		JOIN ai.knowledge_bases kb ON kb.id = ckb.knowledge_base_id
		WHERE ckb.chatbot_id = $1
		ORDER BY ckb.priority DESC
	`

	rows, err := s.db.Query(ctx, query, chatbotID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatbot knowledge bases: %w", err)
	}
	defer rows.Close()

	var links []ChatbotKnowledgeBase
	for rows.Next() {
		var link ChatbotKnowledgeBase
		if err := rows.Scan(
			&link.ID, &link.ChatbotID, &link.KnowledgeBaseID,
			&link.AccessLevel, &link.FilterExpression, &link.ContextWeight,
			&link.Priority, &link.IntentKeywords, &link.MaxChunks,
			&link.SimilarityThreshold, &link.Enabled, &link.Metadata,
			&link.CreatedAt, &link.UpdatedAt,
			&link.KnowledgeBaseName,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan chatbot knowledge base link")
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// GetChatbotKnowledgeBaseLinks is an alias for GetChatbotKnowledgeBases for the query router
func (s *KnowledgeBaseStorage) GetChatbotKnowledgeBaseLinks(ctx context.Context, chatbotID string) ([]ChatbotKnowledgeBase, error) {
	return s.GetChatbotKnowledgeBases(ctx, chatbotID)
}

// GetKnowledgeBaseChatbots retrieves all chatbot links for a knowledge base (reverse lookup)
// This is used to show which chatbots are using a specific knowledge base
func (s *KnowledgeBaseStorage) GetKnowledgeBaseChatbots(ctx context.Context, knowledgeBaseID string) ([]ChatbotKnowledgeBase, error) {
	query := `
		SELECT ckb.id, ckb.chatbot_id, ckb.knowledge_base_id,
			ckb.access_level, ckb.filter_expression, ckb.context_weight,
			ckb.priority, ckb.intent_keywords, ckb.max_chunks,
			ckb.similarity_threshold, ckb.enabled, ckb.metadata,
			ckb.created_at, ckb.updated_at,
			c.name as chatbot_name
		FROM ai.chatbot_knowledge_bases ckb
		JOIN ai.chatbots c ON c.id = ckb.chatbot_id
		WHERE ckb.knowledge_base_id = $1
		ORDER BY ckb.priority ASC
	`

	rows, err := s.db.Query(ctx, query, knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get knowledge base chatbots: %w", err)
	}
	defer rows.Close()

	var links []ChatbotKnowledgeBase
	for rows.Next() {
		var link ChatbotKnowledgeBase
		if err := rows.Scan(
			&link.ID, &link.ChatbotID, &link.KnowledgeBaseID,
			&link.AccessLevel, &link.FilterExpression, &link.ContextWeight,
			&link.Priority, &link.IntentKeywords, &link.MaxChunks,
			&link.SimilarityThreshold, &link.Enabled, &link.Metadata,
			&link.CreatedAt, &link.UpdatedAt,
			&link.ChatbotName,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan knowledge base chatbot link")
			continue
		}
		links = append(links, link)
	}

	return links, nil
}

// UnlinkChatbotKnowledgeBase removes a link between chatbot and knowledge base
func (s *KnowledgeBaseStorage) UnlinkChatbotKnowledgeBase(ctx context.Context, chatbotID, knowledgeBaseID string) error {
	_, err := s.db.Exec(ctx,
		"DELETE FROM ai.chatbot_knowledge_bases WHERE chatbot_id = $1 AND knowledge_base_id = $2",
		chatbotID, knowledgeBaseID,
	)
	return err
}

// ============================================================================
// Vector Search / Retrieval
// ============================================================================

// SearchChunks searches for similar chunks in a knowledge base
func (s *KnowledgeBaseStorage) SearchChunks(ctx context.Context, knowledgeBaseID string, queryEmbedding []float32, limit int, threshold float64) ([]RetrievalResult, error) {
	// Format embedding as PostgreSQL vector literal
	embeddingStr := formatEmbeddingLiteral(queryEmbedding)

	// Log embedding info for debugging
	embeddingPreview := embeddingStr
	if len(embeddingPreview) > 100 {
		embeddingPreview = embeddingPreview[:100] + "..."
	}
	log.Debug().
		Int("embedding_length", len(queryEmbedding)).
		Str("kb_id", knowledgeBaseID).
		Float64("threshold", threshold).
		Int("limit", limit).
		Str("embedding_preview", embeddingPreview).
		Msg("SearchChunks starting")

	query := fmt.Sprintf(`
		SELECT
			c.id as chunk_id,
			c.document_id,
			c.content,
			1 - (c.embedding <=> '%s'::vector) as similarity,
			c.metadata,
			d.title as document_title
		FROM ai.chunks c
		JOIN ai.documents d ON d.id = c.document_id
		WHERE c.knowledge_base_id = $1
		  AND 1 - (c.embedding <=> '%s'::vector) >= $2
		ORDER BY c.embedding <=> '%s'::vector
		LIMIT $3
	`, embeddingStr, embeddingStr, embeddingStr)

	rows, err := s.db.Query(ctx, query, knowledgeBaseID, threshold, limit)
	if err != nil {
		log.Error().Err(err).Str("kb_id", knowledgeBaseID).Msg("SearchChunks query failed")
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}
	defer rows.Close()

	var results []RetrievalResult
	for rows.Next() {
		var r RetrievalResult
		var docTitle *string
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Similarity, &r.Metadata, &docTitle); err != nil {
			log.Warn().Err(err).Msg("Failed to scan search result")
			continue
		}
		r.KnowledgeBaseID = knowledgeBaseID
		if docTitle != nil {
			r.DocumentTitle = *docTitle
		}
		results = append(results, r)
	}

	// Log results
	if len(results) > 0 {
		log.Debug().
			Int("results_count", len(results)).
			Float64("top_similarity", results[0].Similarity).
			Str("kb_id", knowledgeBaseID).
			Msg("SearchChunks completed")
	} else {
		log.Debug().
			Str("kb_id", knowledgeBaseID).
			Float64("threshold", threshold).
			Msg("SearchChunks returned no results")
	}

	return results, nil
}

// SearchMode defines how search should be performed
type SearchMode string

const (
	SearchModeSemantic SearchMode = "semantic" // Vector similarity only
	SearchModeKeyword  SearchMode = "keyword"  // Full-text search only
	SearchModeHybrid   SearchMode = "hybrid"   // Combined vector + full-text
)

// HybridSearchOptions contains options for hybrid search
type HybridSearchOptions struct {
	Query          string
	QueryEmbedding []float32
	Limit          int
	Threshold      float64
	Mode           SearchMode
	SemanticWeight float64         // Weight for semantic score (0-1), keyword weight = 1 - semantic
	KeywordBoost   float64         // Boost factor for exact keyword matches
	Filter         *MetadataFilter // Optional metadata filter for user isolation
}

// GraphBoostOptions contains options for graph-boosted search
type GraphBoostOptions struct {
	QueryEmbedding   []float32 // Query vector embedding
	QueryText        string    // Query text for entity extraction
	Limit            int       // Maximum number of results to return
	Threshold        float64   // Minimum similarity threshold (0-1)
	GraphBoostWeight float64   // How much to weight entity matches vs vector similarity (0.0-1.0)
}

// SearchChunksHybrid performs hybrid search combining vector similarity with full-text search
func (s *KnowledgeBaseStorage) SearchChunksHybrid(ctx context.Context, knowledgeBaseID string, opts HybridSearchOptions) ([]RetrievalResult, error) {
	// Default weights
	if opts.SemanticWeight == 0 {
		opts.SemanticWeight = 0.5 // 50/50 by default
	}
	if opts.KeywordBoost == 0 {
		opts.KeywordBoost = 0.3 // 30% boost for keyword matches
	}

	log.Debug().
		Str("mode", string(opts.Mode)).
		Str("query", opts.Query).
		Float64("semantic_weight", opts.SemanticWeight).
		Float64("threshold", opts.Threshold).
		Msg("SearchChunksHybrid starting")

	switch opts.Mode {
	case SearchModeKeyword:
		return s.searchKeywordOnly(ctx, knowledgeBaseID, opts)
	case SearchModeHybrid:
		return s.searchHybrid(ctx, knowledgeBaseID, opts)
	default: // SearchModeSemantic
		return s.SearchChunks(ctx, knowledgeBaseID, opts.QueryEmbedding, opts.Limit, opts.Threshold)
	}
}

// searchKeywordOnly performs full-text search only
func (s *KnowledgeBaseStorage) searchKeywordOnly(ctx context.Context, knowledgeBaseID string, opts HybridSearchOptions) ([]RetrievalResult, error) {
	// Prepare the search query for PostgreSQL full-text search
	// Use plainto_tsquery for simple word matching, or websearch_to_tsquery for more advanced
	query := `
		SELECT
			c.id as chunk_id,
			c.document_id,
			c.content,
			ts_rank_cd(to_tsvector('simple', c.content), plainto_tsquery('simple', $2)) as similarity,
			c.metadata,
			d.title as document_title
		FROM ai.chunks c
		JOIN ai.documents d ON d.id = c.document_id
		WHERE c.knowledge_base_id = $1
		  AND (
		    to_tsvector('simple', c.content) @@ plainto_tsquery('simple', $2)
		    OR c.content ILIKE '%' || $2 || '%'
		  )
		ORDER BY similarity DESC
		LIMIT $3
	`

	rows, err := s.db.Query(ctx, query, knowledgeBaseID, opts.Query, opts.Limit)
	if err != nil {
		log.Error().Err(err).Str("kb_id", knowledgeBaseID).Msg("Keyword search query failed")
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}
	defer rows.Close()

	var results []RetrievalResult
	for rows.Next() {
		var r RetrievalResult
		var docTitle *string
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Similarity, &r.Metadata, &docTitle); err != nil {
			log.Warn().Err(err).Msg("Failed to scan keyword search result")
			continue
		}
		r.KnowledgeBaseID = knowledgeBaseID
		if docTitle != nil {
			r.DocumentTitle = *docTitle
		}
		// Normalize similarity to 0-1 range (ts_rank_cd can exceed 1)
		if r.Similarity > 1 {
			r.Similarity = 1
		}
		results = append(results, r)
	}

	log.Debug().
		Int("results_count", len(results)).
		Str("kb_id", knowledgeBaseID).
		Msg("Keyword search completed")

	return results, nil
}

// searchHybrid combines vector similarity with full-text search
func (s *KnowledgeBaseStorage) searchHybrid(ctx context.Context, knowledgeBaseID string, opts HybridSearchOptions) ([]RetrievalResult, error) {
	embeddingStr := formatEmbeddingLiteral(opts.QueryEmbedding)
	keywordWeight := 1 - opts.SemanticWeight

	// Build dynamic filter conditions for user isolation
	filterConditions := ""
	args := []interface{}{knowledgeBaseID, opts.Query, opts.SemanticWeight, keywordWeight, opts.KeywordBoost, opts.Threshold, opts.Limit}
	argIndex := 8

	if opts.Filter != nil && opts.Filter.UserID != nil {
		// Include user's content OR content without user_id (global)
		filterConditions += fmt.Sprintf(` AND (
			d.metadata->>'user_id' = $%d OR
			d.metadata->>'user_id' IS NULL OR
			NOT (d.metadata ? 'user_id')
		)`, argIndex)
		args = append(args, *opts.Filter.UserID)
		argIndex++
	}

	if opts.Filter != nil && len(opts.Filter.Tags) > 0 {
		filterConditions += fmt.Sprintf(" AND d.tags @> $%d", argIndex)
		args = append(args, opts.Filter.Tags)
		argIndex++
	}

	// Apply arbitrary metadata filters
	if opts.Filter != nil && len(opts.Filter.Metadata) > 0 {
		for key, value := range opts.Filter.Metadata {
			// Use parameterized value but key must be sanitized (alphanumeric + underscore only)
			safeKey := sanitizeMetadataKey(key)
			filterConditions += fmt.Sprintf(" AND d.metadata->>'%s' = $%d", safeKey, argIndex)
			args = append(args, value)
			argIndex++
		}
	}

	// Hybrid query combining vector similarity and full-text search
	// The final score is: (semantic_weight * vector_similarity) + (keyword_weight * text_rank) + keyword_boost_if_match
	query := fmt.Sprintf(`
		WITH vector_search AS (
			SELECT
				c.id as chunk_id,
				c.document_id,
				c.content,
				c.metadata,
				1 - (c.embedding <=> '%s'::vector) as vector_similarity
			FROM ai.chunks c
			WHERE c.knowledge_base_id = $1
			  AND c.embedding IS NOT NULL
		),
		text_search AS (
			SELECT
				c.id as chunk_id,
				ts_rank_cd(to_tsvector('simple', c.content), plainto_tsquery('simple', $2)) as text_rank,
				CASE
					WHEN c.content ILIKE '%%' || $2 || '%%' THEN $5::float
					ELSE 0
				END as keyword_boost
			FROM ai.chunks c
			WHERE c.knowledge_base_id = $1
		)
		SELECT
			v.chunk_id,
			v.document_id,
			v.content,
			(($3::float * v.vector_similarity) + ($4::float * COALESCE(t.text_rank, 0)) + COALESCE(t.keyword_boost, 0)) as similarity,
			v.metadata,
			d.title as document_title,
			d.tags,
			v.vector_similarity,
			COALESCE(t.text_rank, 0) as text_rank
		FROM vector_search v
		JOIN ai.documents d ON d.id = v.document_id
		LEFT JOIN text_search t ON t.chunk_id = v.chunk_id
		WHERE (($3::float * v.vector_similarity) + ($4::float * COALESCE(t.text_rank, 0)) + COALESCE(t.keyword_boost, 0)) >= $6
		%s
		ORDER BY similarity DESC
		LIMIT $7
	`, embeddingStr, filterConditions)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		log.Error().Err(err).Str("kb_id", knowledgeBaseID).Msg("Hybrid search query failed")
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}
	defer rows.Close()

	var results []RetrievalResult
	for rows.Next() {
		var r RetrievalResult
		var docTitle *string
		var tags []string
		var vectorSim, textRank float64
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Similarity, &r.Metadata, &docTitle, &tags, &vectorSim, &textRank); err != nil {
			log.Warn().Err(err).Msg("Failed to scan hybrid search result")
			continue
		}
		r.KnowledgeBaseID = knowledgeBaseID
		if docTitle != nil {
			r.DocumentTitle = *docTitle
		}
		r.Tags = tags

		log.Debug().
			Str("chunk_id", r.ChunkID).
			Float64("vector_sim", vectorSim).
			Float64("text_rank", textRank).
			Float64("combined", r.Similarity).
			Msg("Hybrid result")

		results = append(results, r)
	}

	log.Debug().
		Int("results_count", len(results)).
		Str("kb_id", knowledgeBaseID).
		Msg("Hybrid search completed")

	return results, nil
}

// SearchChunksWithGraphBoost performs vector search with entity-based boosting
// This combines semantic similarity with knowledge graph entity salience
// Entities are extracted from the query and documents mentioning those entities receive a ranking boost
func (s *KnowledgeBaseStorage) SearchChunksWithGraphBoost(
	ctx context.Context,
	knowledgeBaseID string,
	knowledgeGraph *KnowledgeGraph,
	entityExtractor EntityExtractor,
	opts GraphBoostOptions,
) ([]RetrievalResult, error) {
	// Apply defaults and validate
	if opts.GraphBoostWeight < 0 {
		opts.GraphBoostWeight = 0
	} else if opts.GraphBoostWeight > 1 {
		opts.GraphBoostWeight = 1
	}

	// If no boosting requested, use regular search for efficiency
	if opts.GraphBoostWeight == 0 {
		return s.SearchChunks(ctx, knowledgeBaseID, opts.QueryEmbedding, opts.Limit, opts.Threshold)
	}

	log.Debug().
		Str("kb_id", knowledgeBaseID).
		Float64("boost_weight", opts.GraphBoostWeight).
		Str("query", opts.QueryText).
		Msg("SearchChunksWithGraphBoost starting")

	// Step 1: Extract entities from query text
	var queryEntities []Entity
	if entityExtractor != nil && opts.QueryText != "" {
		extracted, err := entityExtractor.ExtractEntities(opts.QueryText, knowledgeBaseID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to extract entities from query, using vector-only search")
			return s.SearchChunks(ctx, knowledgeBaseID, opts.QueryEmbedding, opts.Limit, opts.Threshold)
		}
		queryEntities = extracted.Entities
		log.Debug().Int("entity_count", len(queryEntities)).Msg("Extracted entities from query")
	}

	// Step 2: Get more results than needed (for re-ranking)
	retrievalLimit := opts.Limit * 3 // Get 3x results for re-ranking
	if retrievalLimit < 10 {
		retrievalLimit = 10
	}
	if retrievalLimit > 100 {
		retrievalLimit = 100
	}

	chunks, err := s.SearchChunks(ctx, knowledgeBaseID, opts.QueryEmbedding, retrievalLimit, opts.Threshold)
	if err != nil {
		return nil, err
	}

	if len(chunks) == 0 {
		return chunks, nil
	}

	// Step 3: Calculate entity salience per document
	documentEntitySalience := make(map[string]float64) // document_id -> salience sum

	if len(queryEntities) > 0 && knowledgeGraph != nil {
		// For each query entity, find documents that mention it
		for _, queryEntity := range queryEntities {
			// Search for exact entity matches in the knowledge base
			matchingEntities, err := knowledgeGraph.SearchEntities(ctx, knowledgeBaseID, queryEntity.CanonicalName, nil, 50)
			if err != nil {
				log.Warn().Err(err).Str("entity", queryEntity.CanonicalName).Msg("Failed to search for entity")
				continue
			}

			// For each matching entity, get documents mentioning it with salience
			for _, entity := range matchingEntities {
				docEntities, err := knowledgeGraph.GetDocumentEntities(ctx, "")
				if err != nil {
					continue
				}

				// Aggregate salience per document
				for _, de := range docEntities {
					if de.EntityID == entity.ID {
						documentEntitySalience[de.DocumentID] += de.Salience
					}
				}
			}
		}
	}

	log.Debug().
		Int("documents_with_entities", len(documentEntitySalience)).
		Msg("Found documents with entity matches")

	// Step 4: Apply entity boost and re-rank
	type boostedResult struct {
		result      RetrievalResult
		entityBoost float64
		finalScore  float64
	}

	boosted := make([]boostedResult, 0, len(chunks))

	// Find max salience for normalization
	maxSalience := 0.0
	for _, salience := range documentEntitySalience {
		if salience > maxSalience {
			maxSalience = salience
		}
	}

	for _, chunk := range chunks {
		entityBoost := 0.0
		if salience, ok := documentEntitySalience[chunk.DocumentID]; ok && maxSalience > 0 {
			// Normalize to 0-1
			entityBoost = (salience / maxSalience)
		}

		// Combined score: weighted average of vector similarity and entity boost
		vectorWeight := 1.0 - opts.GraphBoostWeight
		finalScore := (chunk.Similarity * vectorWeight) + (entityBoost * opts.GraphBoostWeight)

		boosted = append(boosted, boostedResult{
			result:      chunk,
			entityBoost: entityBoost,
			finalScore:  finalScore,
		})
	}

	// Sort by final score (descending)
	sort.Slice(boosted, func(i, j int) bool {
		return boosted[i].finalScore > boosted[j].finalScore
	})

	// Step 5: Return top N results
	resultCount := opts.Limit
	if resultCount > len(boosted) {
		resultCount = len(boosted)
	}

	results := make([]RetrievalResult, resultCount)
	for i := 0; i < resultCount; i++ {
		results[i] = boosted[i].result
		// Update similarity to the final score for consistency
		results[i].Similarity = boosted[i].finalScore
		// Store boost info in metadata for debugging (serialize as JSON)
		metadataMap := make(map[string]any)
		if len(results[i].Metadata) > 0 {
			// Parse existing metadata if present (ignore errors - this is just for debugging)
			_ = json.Unmarshal(results[i].Metadata, &metadataMap)
		}
		metadataMap["entity_boost"] = boosted[i].entityBoost
		metadataMap["final_score"] = boosted[i].finalScore
		metadataJSON, _ := json.Marshal(metadataMap)
		results[i].Metadata = json.RawMessage(metadataJSON)
	}

	log.Debug().
		Int("results_count", len(results)).
		Float64("top_boost", boosted[0].entityBoost).
		Float64("top_final_score", boosted[0].finalScore).
		Msg("SearchChunksWithGraphBoost completed")

	return results, nil
}

// buildMetadataFilterSQL builds SQL WHERE conditions and args from a MetadataFilterGroup
// Returns (WHERE clause fragment, args, error)
func buildMetadataFilterSQL(group MetadataFilterGroup, argIndex *int) (string, []interface{}, error) {
	var conditions []string
	var args []interface{}

	// Process all conditions in this group
	for _, cond := range group.Conditions {
		conditionSQL, conditionArgs, err := buildConditionSQL(cond, argIndex)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build condition for key '%s': %w", cond.Key, err)
		}
		conditions = append(conditions, conditionSQL)
		args = append(args, conditionArgs...)
	}

	// Process nested groups recursively
	for _, nestedGroup := range group.Groups {
		nestedSQL, nestedArgs, err := buildMetadataFilterSQL(nestedGroup, argIndex)
		if err != nil {
			return "", nil, err
		}
		if nestedSQL != "" {
			conditions = append(conditions, fmt.Sprintf("(%s)", nestedSQL))
			args = append(args, nestedArgs...)
		}
	}

	if len(conditions) == 0 {
		return "", args, nil
	}

	logicalOp := string(group.LogicalOp)
	if logicalOp == "" {
		logicalOp = "AND" // Default to AND
	}

	whereClause := strings.Join(conditions, fmt.Sprintf(" %s ", logicalOp))
	return whereClause, args, nil
}

// buildConditionSQL builds SQL for a single MetadataCondition
func buildConditionSQL(cond MetadataCondition, argIndex *int) (string, []interface{}, error) {
	var args []interface{}
	var sqlCond string

	// Use d.metadata->>'key' syntax to extract metadata as text
	metadataRef := fmt.Sprintf("d.metadata->>'%s'", escapeStringLiteral(cond.Key))

	switch cond.Operator {
	case MetadataOpEquals:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for equals operator")
		}
		sqlCond = fmt.Sprintf("%s = $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpNotEquals:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for not equals operator")
		}
		sqlCond = fmt.Sprintf("%s != $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpILike:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for ilike operator")
		}
		sqlCond = fmt.Sprintf("%s ILIKE $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLike:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for like operator")
		}
		sqlCond = fmt.Sprintf("%s LIKE $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpIn:
		if len(cond.Values) == 0 {
			return "", nil, errors.New("values are required for IN operator")
		}
		placeholders := make([]string, len(cond.Values))
		for i, v := range cond.Values {
			placeholders[i] = fmt.Sprintf("$%d", *argIndex)
			args = append(args, fmt.Sprintf("%v", v))
			*argIndex++
		}
		sqlCond = fmt.Sprintf("%s IN (%s)", metadataRef, strings.Join(placeholders, ", "))

	case MetadataOpNotIn:
		if len(cond.Values) == 0 {
			return "", nil, errors.New("values are required for NOT IN operator")
		}
		placeholders := make([]string, len(cond.Values))
		for i, v := range cond.Values {
			placeholders[i] = fmt.Sprintf("$%d", *argIndex)
			args = append(args, fmt.Sprintf("%v", v))
			*argIndex++
		}
		sqlCond = fmt.Sprintf("%s NOT IN (%s)", metadataRef, strings.Join(placeholders, ", "))

	case MetadataOpGreaterThan:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for greater than operator")
		}
		sqlCond = fmt.Sprintf("%s > $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpGreaterThanOr:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for greater than or equal operator")
		}
		sqlCond = fmt.Sprintf("%s >= $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLessThan:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for less than operator")
		}
		sqlCond = fmt.Sprintf("%s < $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLessThanOr:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for less than or equal operator")
		}
		sqlCond = fmt.Sprintf("%s <= $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpBetween:
		if cond.Min == nil || cond.Max == nil {
			return "", nil, errors.New("min and max are required for BETWEEN operator")
		}
		sqlCond = fmt.Sprintf("%s BETWEEN $%d AND $%d", metadataRef, *argIndex, *argIndex+1)
		args = append(args, fmt.Sprintf("%v", cond.Min), fmt.Sprintf("%v", cond.Max))
		*argIndex += 2

	case MetadataOpIsNull:
		sqlCond = fmt.Sprintf("%s IS NULL", metadataRef)

	case MetadataOpIsNotNull:
		sqlCond = fmt.Sprintf("%s IS NOT NULL", metadataRef)

	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}

	return sqlCond, args, nil
}

// escapeStringLiteral escapes single quotes in a string literal for SQL
func escapeStringLiteral(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// buildMetadataFilterSQLForTable builds SQL WHERE conditions and args from a MetadataFilterGroup
// tablePrefix is the table alias prefix (e.g., "d" for "d.metadata") or empty string for direct table access
func buildMetadataFilterSQLForTable(group MetadataFilterGroup, argIndex *int, tablePrefix string) (string, []interface{}, error) {
	// Reuse the existing function with proper table prefix
	// For backward compatibility, when tablePrefix is empty, we use "metadata" directly
	prefix := tablePrefix
	if prefix == "" {
		prefix = "" // No prefix needed
	} else if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}

	// Build conditions using the modified prefix
	var conditions []string
	var args []interface{}

	for _, cond := range group.Conditions {
		conditionSQL, conditionArgs, err := buildConditionSQLForTable(cond, argIndex, prefix)
		if err != nil {
			return "", nil, fmt.Errorf("failed to build condition for key '%s': %w", cond.Key, err)
		}
		conditions = append(conditions, conditionSQL)
		args = append(args, conditionArgs...)
	}

	// Process nested groups recursively
	for _, nestedGroup := range group.Groups {
		nestedSQL, nestedArgs, err := buildMetadataFilterSQLForTable(nestedGroup, argIndex, tablePrefix)
		if err != nil {
			return "", nil, err
		}
		if nestedSQL != "" {
			conditions = append(conditions, fmt.Sprintf("(%s)", nestedSQL))
			args = append(args, nestedArgs...)
		}
	}

	if len(conditions) == 0 {
		return "", args, nil
	}

	logicalOp := string(group.LogicalOp)
	if logicalOp == "" {
		logicalOp = "AND"
	}

	whereClause := strings.Join(conditions, fmt.Sprintf(" %s ", logicalOp))
	return whereClause, args, nil
}

// buildConditionSQLForTable builds SQL for a single MetadataCondition with a table prefix
func buildConditionSQLForTable(cond MetadataCondition, argIndex *int, tablePrefix string) (string, []interface{}, error) {
	var args []interface{}
	var sqlCond string

	// Use prefix + metadata->>'key' syntax
	metadataRef := fmt.Sprintf("%smetadata->>'%s'", tablePrefix, escapeStringLiteral(cond.Key))

	switch cond.Operator {
	case MetadataOpEquals:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for equals operator")
		}
		sqlCond = fmt.Sprintf("%s = $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpNotEquals:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for not equals operator")
		}
		sqlCond = fmt.Sprintf("%s != $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpILike:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for ilike operator")
		}
		sqlCond = fmt.Sprintf("%s ILIKE $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLike:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for like operator")
		}
		sqlCond = fmt.Sprintf("%s LIKE $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpIn:
		if len(cond.Values) == 0 {
			return "", nil, errors.New("values are required for IN operator")
		}
		placeholders := make([]string, len(cond.Values))
		for i, v := range cond.Values {
			placeholders[i] = fmt.Sprintf("$%d", *argIndex)
			args = append(args, fmt.Sprintf("%v", v))
			*argIndex++
		}
		sqlCond = fmt.Sprintf("%s IN (%s)", metadataRef, strings.Join(placeholders, ", "))

	case MetadataOpNotIn:
		if len(cond.Values) == 0 {
			return "", nil, errors.New("values are required for NOT IN operator")
		}
		placeholders := make([]string, len(cond.Values))
		for i, v := range cond.Values {
			placeholders[i] = fmt.Sprintf("$%d", *argIndex)
			args = append(args, fmt.Sprintf("%v", v))
			*argIndex++
		}
		sqlCond = fmt.Sprintf("%s NOT IN (%s)", metadataRef, strings.Join(placeholders, ", "))

	case MetadataOpGreaterThan:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for greater than operator")
		}
		sqlCond = fmt.Sprintf("%s > $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpGreaterThanOr:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for greater than or equal operator")
		}
		sqlCond = fmt.Sprintf("%s >= $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLessThan:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for less than operator")
		}
		sqlCond = fmt.Sprintf("%s < $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpLessThanOr:
		if cond.Value == nil {
			return "", nil, errors.New("value is required for less than or equal operator")
		}
		sqlCond = fmt.Sprintf("%s <= $%d", metadataRef, *argIndex)
		args = append(args, fmt.Sprintf("%v", cond.Value))
		*argIndex++

	case MetadataOpBetween:
		if cond.Min == nil || cond.Max == nil {
			return "", nil, errors.New("min and max are required for BETWEEN operator")
		}
		sqlCond = fmt.Sprintf("%s BETWEEN $%d AND $%d", metadataRef, *argIndex, *argIndex+1)
		args = append(args, fmt.Sprintf("%v", cond.Min), fmt.Sprintf("%v", cond.Max))
		*argIndex += 2

	case MetadataOpIsNull:
		sqlCond = fmt.Sprintf("%s IS NULL", metadataRef)

	case MetadataOpIsNotNull:
		sqlCond = fmt.Sprintf("%s IS NOT NULL", metadataRef)

	default:
		return "", nil, fmt.Errorf("unsupported operator: %s", cond.Operator)
	}

	return sqlCond, args, nil
}

// SearchChunksWithFilter searches for similar chunks with metadata filtering for user isolation
func (s *KnowledgeBaseStorage) SearchChunksWithFilter(
	ctx context.Context,
	knowledgeBaseID string,
	queryEmbedding []float32,
	limit int,
	threshold float64,
	filter *MetadataFilter,
) ([]RetrievalResult, error) {
	// Format embedding as PostgreSQL vector literal
	embeddingStr := formatEmbeddingLiteral(queryEmbedding)

	// Build dynamic WHERE clause for filtering
	whereConditions := []string{
		"c.knowledge_base_id = $1",
		fmt.Sprintf("1 - (c.embedding <=> '%s'::vector) >= $2", embeddingStr),
	}
	args := []interface{}{knowledgeBaseID, threshold, limit}
	argIndex := 4

	// User isolation filter
	if filter != nil && filter.UserID != nil {
		// Include user's content OR content without user_id (global)
		whereConditions = append(whereConditions, fmt.Sprintf(`(
			d.metadata->>'user_id' = $%d OR
			d.metadata->>'user_id' IS NULL OR
			NOT (d.metadata ? 'user_id')
		)`, argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	// Tag filter - documents must have ALL specified tags
	if filter != nil && len(filter.Tags) > 0 {
		whereConditions = append(whereConditions, fmt.Sprintf("d.tags @> $%d", argIndex))
		args = append(args, filter.Tags)
		argIndex++
	}

	// Advanced metadata filter with operators and logical combinations
	if filter != nil && filter.AdvancedFilter != nil {
		metadataSQL, metadataArgs, err := buildMetadataFilterSQL(*filter.AdvancedFilter, &argIndex)
		if err != nil {
			return nil, fmt.Errorf("failed to build metadata filter: %w", err)
		}
		if metadataSQL != "" {
			whereConditions = append(whereConditions, metadataSQL)
			args = append(args, metadataArgs...)
		}
	}

	// Legacy simple metadata filter (exact match only) - for backward compatibility
	if filter != nil && filter.AdvancedFilter == nil && len(filter.Metadata) > 0 {
		for key, value := range filter.Metadata {
			escapedKey := escapeStringLiteral(key)
			whereConditions = append(whereConditions, fmt.Sprintf("d.metadata->>'%s' = $%d", escapedKey, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	whereClause := strings.Join(whereConditions, " AND ")

	query := fmt.Sprintf(`
		SELECT
			c.id as chunk_id,
			c.document_id,
			c.content,
			1 - (c.embedding <=> '%s'::vector) as similarity,
			c.metadata,
			d.title as document_title,
			d.tags
		FROM ai.chunks c
		JOIN ai.documents d ON d.id = c.document_id
		WHERE %s
		ORDER BY c.embedding <=> '%s'::vector
		LIMIT $3
	`, embeddingStr, whereClause, embeddingStr)

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks with filter: %w", err)
	}
	defer rows.Close()

	var results []RetrievalResult
	for rows.Next() {
		var r RetrievalResult
		var docTitle *string
		var tags []string
		if err := rows.Scan(&r.ChunkID, &r.DocumentID, &r.Content, &r.Similarity, &r.Metadata, &docTitle, &tags); err != nil {
			log.Warn().Err(err).Msg("Failed to scan filtered search result")
			continue
		}
		r.KnowledgeBaseID = knowledgeBaseID
		if docTitle != nil {
			r.DocumentTitle = *docTitle
		}
		r.Tags = tags
		results = append(results, r)
	}

	return results, nil
}

// SearchChatbotKnowledgeOptions contains options for chatbot knowledge search
type SearchChatbotKnowledgeOptions struct {
	UserID    *string
	MaxChunks int
	Threshold float64
}

// SearchChatbotKnowledge searches all knowledge bases linked to a chatbot
func (s *KnowledgeBaseStorage) SearchChatbotKnowledge(ctx context.Context, chatbotID string, queryEmbedding []float32) ([]RetrievalResult, error) {
	return s.SearchChatbotKnowledgeWithOptions(ctx, chatbotID, queryEmbedding, SearchChatbotKnowledgeOptions{})
}

// SearchChatbotKnowledgeWithOptions searches all knowledge bases linked to a chatbot with user context
func (s *KnowledgeBaseStorage) SearchChatbotKnowledgeWithOptions(ctx context.Context, chatbotID string, queryEmbedding []float32, opts SearchChatbotKnowledgeOptions) ([]RetrievalResult, error) {
	// Get linked knowledge bases
	links, err := s.GetChatbotKnowledgeBases(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	if len(links) == 0 {
		return nil, nil
	}

	// Search each knowledge base and combine results
	var allResults []RetrievalResult
	for _, link := range links {
		if !link.Enabled {
			continue
		}

		// Use link defaults or fall back to system defaults
		maxChunks := 5
		if link.MaxChunks != nil {
			maxChunks = *link.MaxChunks
		}
		if opts.MaxChunks > 0 {
			maxChunks = opts.MaxChunks
		}

		threshold := 0.7
		if link.SimilarityThreshold != nil {
			threshold = *link.SimilarityThreshold
		}
		if opts.Threshold > 0 {
			threshold = opts.Threshold
		}

		var results []RetrievalResult

		// Build filter for user isolation and access level
		var filter *MetadataFilter
		if opts.UserID != nil || link.AccessLevel == "filtered" {
			filter = &MetadataFilter{}

			// Apply user isolation if UserID provided
			if opts.UserID != nil {
				filter.UserID = opts.UserID
				filter.IncludeGlobal = true
			}

			// Apply FilterExpression for "filtered" access level
			if link.AccessLevel == "filtered" && link.FilterExpression != nil {
				advancedFilter := convertFilterExpression(link.FilterExpression, opts.UserID)
				if advancedFilter != nil {
					filter.AdvancedFilter = advancedFilter
				}
			}
		}

		if filter != nil {
			results, err = s.SearchChunksWithFilter(ctx, link.KnowledgeBaseID, queryEmbedding, maxChunks, threshold, filter)
		} else {
			results, err = s.SearchChunks(ctx, link.KnowledgeBaseID, queryEmbedding, maxChunks, threshold)
		}

		if err != nil {
			log.Warn().Err(err).Str("kb_id", link.KnowledgeBaseID).Msg("Failed to search knowledge base")
			continue
		}

		// Get KB name for context
		kb, err := s.GetKnowledgeBase(ctx, link.KnowledgeBaseID)
		if err == nil && kb != nil {
			for i := range results {
				results[i].KnowledgeBaseName = kb.Name
			}
		}

		allResults = append(allResults, results...)
	}

	return allResults, nil
}

// convertFilterExpression converts a FilterExpression map to a MetadataFilterGroup
// It also substitutes special variables like $user_id
func convertFilterExpression(expr map[string]interface{}, userID *string) *MetadataFilterGroup {
	if expr == nil {
		return nil
	}

	var conditions []MetadataCondition

	for key, value := range expr {
		// Handle special variable substitution
		strValue, ok := value.(string)
		if ok && strValue == "$user_id" {
			if userID != nil {
				conditions = append(conditions, MetadataCondition{
					Key:      key,
					Operator: MetadataOpEquals,
					Value:    *userID,
				})
			}
			continue
		}

		// Handle direct value matches
		conditions = append(conditions, MetadataCondition{
			Key:      key,
			Operator: MetadataOpEquals,
			Value:    value,
		})
	}

	if len(conditions) == 0 {
		return nil
	}

	return &MetadataFilterGroup{
		Conditions: conditions,
		LogicalOp:  LogicalOpAND,
	}
}

// ============================================================================
// Retrieval Logging
// ============================================================================

// LogRetrieval logs a RAG retrieval operation
func (s *KnowledgeBaseStorage) LogRetrieval(ctx context.Context, log *RetrievalLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO ai.retrieval_log (
			id, chatbot_id, conversation_id, knowledge_base_id, user_id,
			query_text, query_embedding_model, chunks_retrieved,
			chunk_ids, similarity_scores, retrieval_duration_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.db.Exec(ctx, query,
		log.ID, log.ChatbotID, log.ConversationID, log.KnowledgeBaseID, log.UserID,
		log.QueryText, log.QueryEmbeddingModel, log.ChunksRetrieved,
		log.ChunkIDs, log.SimilarityScores, log.RetrievalDurationMs,
	)
	return err
}

// formatEmbeddingLiteral formats a float32 slice as PostgreSQL vector literal
// Uses %v format to preserve full float32 precision (7 decimal digits)
func formatEmbeddingLiteral(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		// Use %v for full float32 precision instead of %g which defaults to 6 significant digits
		parts[i] = fmt.Sprintf("%v", f)
	}
	return "[" + joinStrings(parts, ",") + "]"
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// sanitizeMetadataKey sanitizes a metadata key to prevent SQL injection
// Only allows alphanumeric characters and underscores
func sanitizeMetadataKey(key string) string {
	var result strings.Builder
	for _, r := range key {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GetPendingDocuments retrieves documents pending processing
func (s *KnowledgeBaseStorage) GetPendingDocuments(ctx context.Context, limit int) ([]Document, error) {
	query := `
		SELECT id, knowledge_base_id, title, source_url, source_type,
			mime_type, content, content_hash, status, error_message,
			chunks_count, metadata, tags, created_by, created_at, updated_at, indexed_at
		FROM ai.documents
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
	`

	rows, err := s.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending documents: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(
			&doc.ID, &doc.KnowledgeBaseID, &doc.Title, &doc.SourceURL, &doc.SourceType,
			&doc.MimeType, &doc.Content, &doc.ContentHash, &doc.Status, &doc.ErrorMessage,
			&doc.ChunksCount, &doc.Metadata, &doc.Tags, &doc.CreatedBy, &doc.CreatedAt, &doc.UpdatedAt, &doc.IndexedAt,
		); err != nil {
			continue
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// UpdateChunkEmbedding updates the embedding for a single chunk
func (s *KnowledgeBaseStorage) UpdateChunkEmbedding(ctx context.Context, chunkID string, embedding []float32) error {
	embeddingJSON, err := json.Marshal(embedding)
	if err != nil {
		return err
	}

	query := `UPDATE ai.chunks SET embedding = $2::vector WHERE id = $1`
	_, err = s.db.Exec(ctx, query, chunkID, string(embeddingJSON))
	return err
}

// GetChunkEmbeddingPreview returns the first N values of a chunk's embedding for debugging
func (s *KnowledgeBaseStorage) GetChunkEmbeddingPreview(ctx context.Context, chunkID string, n int) ([]float32, error) {
	// Get the embedding as text and parse the first N values
	query := `SELECT left(embedding::text, 500) FROM ai.chunks WHERE id = $1`

	var embeddingText *string
	err := s.db.QueryRow(ctx, query, chunkID).Scan(&embeddingText)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	if embeddingText == nil || *embeddingText == "" {
		return nil, fmt.Errorf("embedding is NULL for chunk %s", chunkID)
	}

	// Parse the vector literal format: [0.1,0.2,0.3,...]
	text := strings.TrimPrefix(*embeddingText, "[")
	parts := strings.Split(text, ",")

	result := make([]float32, 0, n)
	for i := 0; i < n && i < len(parts); i++ {
		var val float64
		_, err := fmt.Sscanf(parts[i], "%f", &val)
		if err != nil {
			break
		}
		result = append(result, float32(val))
	}

	return result, nil
}

// ChunkEmbeddingStats contains statistics about chunk embeddings in a knowledge base
type ChunkEmbeddingStats struct {
	TotalChunks            int `json:"total_chunks"`
	ChunksWithEmbedding    int `json:"chunks_with_embedding"`
	ChunksWithoutEmbedding int `json:"chunks_without_embedding"`
}

// GetChunkEmbeddingStats returns statistics about chunk embeddings for debugging
func (s *KnowledgeBaseStorage) GetChunkEmbeddingStats(ctx context.Context, knowledgeBaseID string) (*ChunkEmbeddingStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COUNT(embedding) as with_embedding,
			COUNT(*) - COUNT(embedding) as without_embedding
		FROM ai.chunks
		WHERE knowledge_base_id = $1
	`

	var stats ChunkEmbeddingStats
	err := s.db.QueryRow(ctx, query, knowledgeBaseID).Scan(
		&stats.TotalChunks,
		&stats.ChunksWithEmbedding,
		&stats.ChunksWithoutEmbedding,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get chunk stats: %w", err)
	}

	return &stats, nil
}

// GetFirstChunkWithEmbedding returns the first chunk ID that has an embedding
func (s *KnowledgeBaseStorage) GetFirstChunkWithEmbedding(ctx context.Context, knowledgeBaseID string) (string, error) {
	query := `
		SELECT id FROM ai.chunks
		WHERE knowledge_base_id = $1 AND embedding IS NOT NULL
		LIMIT 1
	`

	var chunkID string
	err := s.db.QueryRow(ctx, query, knowledgeBaseID).Scan(&chunkID)
	if err != nil {
		return "", fmt.Errorf("no chunks with embeddings found: %w", err)
	}

	return chunkID, nil
}

// ============================================================================
// Convenience Methods for HTTP Handlers
// ============================================================================

// CreateKnowledgeBaseFromRequest creates a knowledge base from a request
func (s *KnowledgeBaseStorage) CreateKnowledgeBaseFromRequest(ctx context.Context, req CreateKnowledgeBaseRequest) (*KnowledgeBase, error) {
	defaults := DefaultKnowledgeBaseConfig()

	kb := &KnowledgeBase{
		Name:      req.Name,
		Namespace: req.Namespace,
		Enabled:   true,
		Source:    "api",
	}

	// Apply defaults where not specified
	if kb.Namespace == "" {
		kb.Namespace = defaults.Namespace
	}
	if req.Description != "" {
		kb.Description = req.Description
	}
	if req.Visibility != nil {
		kb.Visibility = *req.Visibility
	} else {
		kb.Visibility = KBVisibilityPrivate
	}
	if req.EmbeddingModel != "" {
		kb.EmbeddingModel = req.EmbeddingModel
	} else {
		kb.EmbeddingModel = defaults.EmbeddingModel
	}
	if req.EmbeddingDimensions > 0 {
		kb.EmbeddingDimensions = req.EmbeddingDimensions
	} else {
		kb.EmbeddingDimensions = defaults.EmbeddingDimensions
	}
	if req.ChunkSize > 0 {
		kb.ChunkSize = req.ChunkSize
	} else {
		kb.ChunkSize = defaults.ChunkSize
	}
	if req.ChunkOverlap > 0 {
		kb.ChunkOverlap = req.ChunkOverlap
	} else {
		kb.ChunkOverlap = defaults.ChunkOverlap
	}
	if req.ChunkStrategy != "" {
		kb.ChunkStrategy = req.ChunkStrategy
	} else {
		kb.ChunkStrategy = defaults.ChunkStrategy
	}

	if err := s.CreateKnowledgeBase(ctx, kb); err != nil {
		return nil, err
	}

	return kb, nil
}

// UpdateKnowledgeBaseByID updates a knowledge base by ID from a request
func (s *KnowledgeBaseStorage) UpdateKnowledgeBaseByID(ctx context.Context, id string, req UpdateKnowledgeBaseRequest) (*KnowledgeBase, error) {
	// Get existing knowledge base
	kb, err := s.GetKnowledgeBase(ctx, id)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, nil
	}

	// Apply updates
	if req.Name != nil {
		kb.Name = *req.Name
	}
	if req.Description != nil {
		kb.Description = *req.Description
	}
	if req.Visibility != nil {
		kb.Visibility = *req.Visibility
	}
	if req.EmbeddingModel != nil {
		kb.EmbeddingModel = *req.EmbeddingModel
	}
	if req.EmbeddingDimensions != nil {
		kb.EmbeddingDimensions = *req.EmbeddingDimensions
	}
	if req.ChunkSize != nil {
		kb.ChunkSize = *req.ChunkSize
	}
	if req.ChunkOverlap != nil {
		kb.ChunkOverlap = *req.ChunkOverlap
	}
	if req.ChunkStrategy != nil {
		kb.ChunkStrategy = *req.ChunkStrategy
	}
	if req.Enabled != nil {
		kb.Enabled = *req.Enabled
	}

	if err := s.UpdateKnowledgeBase(ctx, kb); err != nil {
		return nil, err
	}

	return kb, nil
}

// ListAllKnowledgeBases lists all knowledge bases (no filtering)
func (s *KnowledgeBaseStorage) ListAllKnowledgeBases(ctx context.Context) ([]KnowledgeBase, error) {
	return s.ListKnowledgeBases(ctx, "", false)
}

// UpdateChatbotKnowledgeBaseOptions represents options for updating a link
type UpdateChatbotKnowledgeBaseOptions struct {
	AccessLevel         *string
	FilterExpression    map[string]interface{}
	ContextWeight       *float64
	Priority            *int
	IntentKeywords      []string
	MaxChunks           *int
	SimilarityThreshold *float64
	Enabled             *bool
}

// UpdateChatbotKnowledgeBaseLink updates a chatbot-knowledge base link
func (s *KnowledgeBaseStorage) UpdateChatbotKnowledgeBaseLink(ctx context.Context, chatbotID, kbID string, opts UpdateChatbotKnowledgeBaseOptions) (*ChatbotKnowledgeBase, error) {
	// First get the existing link
	links, err := s.GetChatbotKnowledgeBases(ctx, chatbotID)
	if err != nil {
		return nil, err
	}

	var existingLink *ChatbotKnowledgeBase
	for i := range links {
		if links[i].KnowledgeBaseID == kbID {
			existingLink = &links[i]
			break
		}
	}

	if existingLink == nil {
		return nil, nil
	}

	// Apply updates
	if opts.AccessLevel != nil {
		existingLink.AccessLevel = *opts.AccessLevel
	}
	if opts.FilterExpression != nil {
		existingLink.FilterExpression = opts.FilterExpression
	}
	if opts.ContextWeight != nil {
		existingLink.ContextWeight = *opts.ContextWeight
	}
	if opts.Priority != nil {
		existingLink.Priority = *opts.Priority
	}
	if opts.IntentKeywords != nil {
		existingLink.IntentKeywords = opts.IntentKeywords
	}
	if opts.MaxChunks != nil {
		existingLink.MaxChunks = opts.MaxChunks
	}
	if opts.SimilarityThreshold != nil {
		existingLink.SimilarityThreshold = opts.SimilarityThreshold
	}
	if opts.Enabled != nil {
		existingLink.Enabled = *opts.Enabled
	}

	// Update using the existing link method (which handles upsert)
	if err := s.LinkChatbotKnowledgeBase(ctx, existingLink); err != nil {
		return nil, err
	}

	return existingLink, nil
}

// LinkChatbotKnowledgeBaseSimple is a convenience method for linking
func (s *KnowledgeBaseStorage) LinkChatbotKnowledgeBaseSimple(ctx context.Context, chatbotID, kbID string, priority, maxChunks int, similarityThreshold float64) (*ChatbotKnowledgeBase, error) {
	link := &ChatbotKnowledgeBase{
		ChatbotID:           chatbotID,
		KnowledgeBaseID:     kbID,
		AccessLevel:         "full",
		Enabled:             true,
		Priority:            priority,
		MaxChunks:           &maxChunks,
		SimilarityThreshold: &similarityThreshold,
	}

	if err := s.LinkChatbotKnowledgeBase(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

// SyncChatbotKnowledgeBaseLinks syncs knowledge base links for a chatbot based on KB names
// This is called when a chatbot is synced from the filesystem to ensure the links match the config
// It will create links for KBs in the config that don't have links, and remove links that aren't in the config
func (s *KnowledgeBaseStorage) SyncChatbotKnowledgeBaseLinks(ctx context.Context, chatbotID string, kbNames []string, maxChunks int, similarityThreshold float64) error {
	// Get knowledge bases by name
	knowledgeBases, err := s.ListKnowledgeBases(ctx, "", false)
	if err != nil {
		return fmt.Errorf("failed to list knowledge bases: %w", err)
	}

	// Build a map of KB name to ID
	kbNameToID := make(map[string]string)
	for _, kb := range knowledgeBases {
		kbNameToID[kb.Name] = kb.ID
	}

	// Get existing links for this chatbot
	existingLinks, err := s.GetChatbotKnowledgeBases(ctx, chatbotID)
	if err != nil {
		return fmt.Errorf("failed to get existing links: %w", err)
	}

	// Build a set of expected KB IDs
	expectedKbIDs := make(map[string]bool)
	for _, kbName := range kbNames {
		if kbID, exists := kbNameToID[kbName]; exists {
			expectedKbIDs[kbID] = true
		} else {
			log.Warn().Str("kb_name", kbName).Str("chatbot_id", chatbotID).Msg("Knowledge base not found for linking")
		}
	}

	// Create or update links for KBs in the config
	for kbID := range expectedKbIDs {
		link := &ChatbotKnowledgeBase{
			ChatbotID:           chatbotID,
			KnowledgeBaseID:     kbID,
			AccessLevel:         "full",
			Enabled:             true,
			Priority:            1,
			MaxChunks:           &maxChunks,
			SimilarityThreshold: &similarityThreshold,
		}

		if err := s.LinkChatbotKnowledgeBase(ctx, link); err != nil {
			log.Warn().Err(err).Str("chatbot_id", chatbotID).Str("kb_id", kbID).Msg("Failed to link knowledge base")
		}
	}

	// Remove links that aren't in the config
	for _, existingLink := range existingLinks {
		if !expectedKbIDs[existingLink.KnowledgeBaseID] {
			if err := s.UnlinkChatbotKnowledgeBase(ctx, chatbotID, existingLink.KnowledgeBaseID); err != nil {
				log.Warn().Err(err).Str("chatbot_id", chatbotID).Str("kb_id", existingLink.KnowledgeBaseID).Msg("Failed to unlink knowledge base")
			}
		}
	}

	return nil
}

// ============================================================================
// Knowledge Base Ownership and Permissions
// ============================================================================

// ListUserKnowledgeBases returns KBs accessible to user
func (s *KnowledgeBaseStorage) ListUserKnowledgeBases(ctx context.Context, userID string) ([]KnowledgeBaseSummary, error) {
	query := `
		SELECT kb.id, kb.name, kb.namespace, kb.description, kb.enabled,
			   kb.document_count, kb.total_chunks, kb.visibility,
			   kb.updated_at,
			   CASE
				   WHEN kbp.permission IS NOT NULL THEN kbp.permission
				   WHEN kb.visibility = 'public' THEN 'viewer'
				   ELSE NULL
			   END as user_permission
		FROM ai.knowledge_bases kb
		LEFT JOIN ai.knowledge_base_permissions kbp
			   ON kbp.knowledge_base_id = kb.id AND kbp.user_id = $1
		WHERE kb.enabled = true
		  AND (kbp.user_id = $1 OR kb.visibility = 'public')
		ORDER BY kb.name
	`

	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list user knowledge bases: %w", err)
	}
	defer rows.Close()

	var kbs []KnowledgeBaseSummary
	for rows.Next() {
		var kb KnowledgeBase
		var userPermission string
		if err := rows.Scan(
			&kb.ID, &kb.Name, &kb.Namespace, &kb.Description,
			&kb.Enabled, &kb.DocumentCount, &kb.TotalChunks,
			&kb.UpdatedAt,
			&userPermission,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan knowledge base row")
			continue
		}
		summary := kb.ToSummary()
		summary.UserPermission = userPermission
		if kb.Visibility != "" {
			summary.Visibility = string(kb.Visibility)
		}
		kbs = append(kbs, summary)
	}

	return kbs, nil
}

// CanUserAccessKB checks if user has access
func (s *KnowledgeBaseStorage) CanUserAccessKB(ctx context.Context, kbID, userID string) bool {
	var hasAccess bool
	query := `
		SELECT EXISTS (
			SELECT 1 FROM ai.knowledge_bases kb
			LEFT JOIN ai.knowledge_base_permissions kbp
				   ON kbp.knowledge_base_id = kb.id AND kbp.user_id = $2
			WHERE kb.id = $1
			  AND kb.enabled = true
			  AND (kbp.user_id = $2 OR kb.visibility = 'public')
		)
	`
	err := s.db.QueryRow(ctx, query, kbID, userID).Scan(&hasAccess)
	return err == nil && hasAccess
}

// CheckKBPermission checks if a user has the required permission level on a KB.
// The permission hierarchy is: viewer < editor < owner
// - If required is "viewer": user needs any permission (viewer, editor, or owner)
// - If required is "editor": user needs editor or owner permission
// - If required is "owner": user must be the KB owner or have owner permission
func (s *KnowledgeBaseStorage) CheckKBPermission(ctx context.Context, kbID, userID, requiredPermission string) (bool, error) {
	// Build the permission check based on required level
	var permissionCheck string
	switch requiredPermission {
	case string(KBPermissionViewer):
		// Any permission level allows read access
		permissionCheck = "kbp.permission IN ('viewer', 'editor', 'owner')"
	case string(KBPermissionEditor):
		// Editor or owner required for write operations
		permissionCheck = "kbp.permission IN ('editor', 'owner')"
	case string(KBPermissionOwner):
		// Only owner permission (or being the KB owner) allows full control
		permissionCheck = "kbp.permission = 'owner'"
	default:
		// Unknown permission level, deny access
		return false, nil
	}

	query := `
		SELECT EXISTS (
			SELECT 1 FROM ai.knowledge_bases kb
			LEFT JOIN ai.knowledge_base_permissions kbp
				ON kbp.knowledge_base_id = kb.id AND kbp.user_id = $2
			WHERE kb.id = $1
			  AND kb.enabled = true
			  AND (kb.owner_id = $2 OR ` + permissionCheck + `)
		)
	`

	var hasPermission bool
	err := s.db.QueryRow(ctx, query, kbID, userID).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check KB permission: %w", err)
	}
	return hasPermission, nil
}

// GetUserKBPermission gets the user's effective permission level on a KB.
// Returns the permission level or empty string if no access.
// For public KBs, returns 'viewer' if no explicit permission exists.
func (s *KnowledgeBaseStorage) GetUserKBPermission(ctx context.Context, kbID, userID string) (string, error) {
	query := `
		SELECT
			CASE
				WHEN kb.owner_id = $2 THEN 'owner'
				WHEN kbp.permission IS NOT NULL THEN kbp.permission
				WHEN kb.visibility = 'public' THEN 'viewer'
				ELSE ''
			END as permission
		FROM ai.knowledge_bases kb
		LEFT JOIN ai.knowledge_base_permissions kbp
			ON kbp.knowledge_base_id = kb.id AND kbp.user_id = $2
		WHERE kb.id = $1 AND kb.enabled = true
	`

	var permission string
	err := s.db.QueryRow(ctx, query, kbID, userID).Scan(&permission)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to get user KB permission: %w", err)
	}
	return permission, nil
}

// GrantKBPermission grants permission to user
func (s *KnowledgeBaseStorage) GrantKBPermission(ctx context.Context, kbID, userID, permission string, grantedBy *string) (*KBPermissionGrant, error) {
	// Upsert permission
	query := `
		INSERT INTO ai.knowledge_base_permissions (knowledge_base_id, user_id, permission, granted_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (knowledge_base_id, user_id)
		DO UPDATE SET permission = $3, granted_by = $4, granted_at = NOW()
		RETURNING id, knowledge_base_id, user_id, permission, granted_by, granted_at
	`

	var grant KBPermissionGrant
	err := s.db.QueryRow(ctx, query, kbID, userID, permission, grantedBy).Scan(
		&grant.ID, &grant.KnowledgeBaseID, &grant.UserID, &grant.Permission, &grant.GrantedBy, &grant.GrantedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to grant permission: %w", err)
	}

	return &grant, nil
}

// ListKBPermissions lists all permissions for a KB
func (s *KnowledgeBaseStorage) ListKBPermissions(ctx context.Context, kbID string) ([]KBPermissionGrant, error) {
	query := `
		SELECT id, knowledge_base_id, user_id, permission, granted_by, granted_at
		FROM ai.knowledge_base_permissions
		WHERE knowledge_base_id = $1
		ORDER BY granted_at DESC
	`

	rows, err := s.db.Query(ctx, query, kbID)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var grants []KBPermissionGrant
	for rows.Next() {
		var grant KBPermissionGrant
		if err := rows.Scan(
			&grant.ID, &grant.KnowledgeBaseID, &grant.UserID, &grant.Permission, &grant.GrantedBy, &grant.GrantedAt,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan permission row")
			continue
		}
		grants = append(grants, grant)
	}

	return grants, nil
}

// RevokeKBPermission revokes permission from user
func (s *KnowledgeBaseStorage) RevokeKBPermission(ctx context.Context, kbID, userID string) error {
	query := `DELETE FROM ai.knowledge_base_permissions WHERE knowledge_base_id = $1 AND user_id = $2`
	_, err := s.db.Exec(ctx, query, kbID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke permission: %w", err)
	}
	return nil
}

// ============================================================================
// DOCUMENT PERMISSIONS
// ============================================================================

// GrantDocumentPermission grants permission on a document to a user
func (s *KnowledgeBaseStorage) GrantDocumentPermission(ctx context.Context, documentID, userID, permission, grantedBy string) (*DocumentPermissionGrant, error) {
	// First check if the requester owns the document
	var ownerID string
	checkQuery := `SELECT owner_id FROM ai.documents WHERE id = $1`
	err := s.db.QueryRow(ctx, checkQuery, documentID).Scan(&ownerID)
	if err != nil {
		return nil, fmt.Errorf("document not found: %w", err)
	}

	// Verify ownership (service role and dashboard admins bypass this check in the handler)
	if ownerID != grantedBy {
		// Check if grantedBy is dashboard admin or service role
		// This is a simple check - in production you'd want proper auth context
		return nil, fmt.Errorf("only document owner can grant permissions")
	}

	// Upsert permission
	query := `
		INSERT INTO ai.document_permissions (document_id, user_id, permission, granted_by)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (document_id, user_id)
		DO UPDATE SET permission = $3, granted_by = $4, granted_at = NOW()
		RETURNING id, document_id, user_id, permission, granted_by, granted_at
	`

	var grant DocumentPermissionGrant
	err = s.db.QueryRow(ctx, query, documentID, userID, permission, grantedBy).Scan(
		&grant.ID, &grant.DocumentID, &grant.UserID, &grant.Permission, &grant.GrantedBy, &grant.GrantedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to grant document permission: %w", err)
	}

	return &grant, nil
}

// ListDocumentPermissions lists all permissions for a document
func (s *KnowledgeBaseStorage) ListDocumentPermissions(ctx context.Context, documentID string) ([]DocumentPermissionGrant, error) {
	query := `
		SELECT id, document_id, user_id, permission, granted_by, granted_at
		FROM ai.document_permissions
		WHERE document_id = $1
		ORDER BY granted_at DESC
	`

	rows, err := s.db.Query(ctx, query, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list document permissions: %w", err)
	}
	defer rows.Close()

	var grants []DocumentPermissionGrant
	for rows.Next() {
		var grant DocumentPermissionGrant
		if err := rows.Scan(
			&grant.ID, &grant.DocumentID, &grant.UserID, &grant.Permission, &grant.GrantedBy, &grant.GrantedAt,
		); err != nil {
			log.Warn().Err(err).Msg("Failed to scan document permission row")
			continue
		}
		grants = append(grants, grant)
	}

	return grants, nil
}

// RevokeDocumentPermission revokes permission from a user on a document
func (s *KnowledgeBaseStorage) RevokeDocumentPermission(ctx context.Context, documentID, userID string) error {
	query := `DELETE FROM ai.document_permissions WHERE document_id = $1 AND user_id = $2`
	_, err := s.db.Exec(ctx, query, documentID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke document permission: %w", err)
	}
	return nil
}

// CanUserAccessDocument checks if a user can access a document
func (s *KnowledgeBaseStorage) CanUserAccessDocument(ctx context.Context, documentID, userID string) (bool, error) {
	// Check if user owns the document
	var ownerID *string
	checkQuery := `SELECT owner_id FROM ai.documents WHERE id = $1`
	err := s.db.QueryRow(ctx, checkQuery, documentID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check document ownership: %w", err)
	}

	// User owns the document
	if ownerID != nil && *ownerID == userID {
		return true, nil
	}

	// Check if user has been granted permission
	var hasPermission bool
	permQuery := `
		SELECT EXISTS(
			SELECT 1 FROM ai.document_permissions
			WHERE document_id = $1 AND user_id = $2
		)
	`
	err = s.db.QueryRow(ctx, permQuery, documentID, userID).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check document permission: %w", err)
	}

	return hasPermission, nil
}

// GetUserQuota retrieves quota information for a user
func (s *KnowledgeBaseStorage) GetUserQuota(ctx context.Context, userID string) (*UserQuota, error) {
	query := `
		SELECT user_id, max_documents, max_chunks, max_storage_bytes,
		       used_documents, used_chunks, used_storage_bytes,
		       created_at, updated_at
		FROM ai.user_quotas
		WHERE user_id = $1
	`

	var quota UserQuota
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&quota.UserID,
		&quota.MaxDocuments,
		&quota.MaxChunks,
		&quota.MaxStorageBytes,
		&quota.UsedDocuments,
		&quota.UsedChunks,
		&quota.UsedStorageBytes,
		&quota.CreatedAt,
		&quota.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &quota, nil
}

// SetUserQuota creates or updates quota for a user
func (s *KnowledgeBaseStorage) SetUserQuota(ctx context.Context, quota *UserQuota) error {
	query := `
		INSERT INTO ai.user_quotas (user_id, max_documents, max_chunks, max_storage_bytes)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE
		SET max_documents = COALESCE(EXCLUDED.max_documents, ai.user_quotas.max_documents),
		    max_chunks = COALESCE(EXCLUDED.max_chunks, ai.user_quotas.max_chunks),
		    max_storage_bytes = COALESCE(EXCLUDED.max_storage_bytes, ai.user_quotas.max_storage_bytes),
		    updated_at = NOW()
	`

	_, err := s.db.Exec(ctx, query,
		quota.UserID,
		quota.MaxDocuments,
		quota.MaxChunks,
		quota.MaxStorageBytes,
	)

	if err != nil {
		return fmt.Errorf("failed to set user quota: %w", err)
	}

	return nil
}

// UpdateUserQuotaUsage updates quota usage counters for a user
func (s *KnowledgeBaseStorage) UpdateUserQuotaUsage(ctx context.Context, userID string, docsDelta int, chunksDelta int, storageDelta int64) error {
	query := `
		INSERT INTO ai.user_quotas (user_id, used_documents, used_chunks, used_storage_bytes)
		VALUES ($1, GREATEST(0, $2), GREATEST(0, $3), GREATEST(0, $4))
		ON CONFLICT (user_id) DO UPDATE
		SET used_documents = GREATEST(0, ai.user_quotas.used_documents + $2),
		    used_chunks = GREATEST(0, ai.user_quotas.used_chunks + $3),
		    used_storage_bytes = GREATEST(0, ai.user_quotas.used_storage_bytes + $4),
		    updated_at = NOW()
	`

	_, err := s.db.Exec(ctx, query, userID, docsDelta, chunksDelta, storageDelta)
	if err != nil {
		return fmt.Errorf("failed to update quota usage: %w", err)
	}

	return nil
}
