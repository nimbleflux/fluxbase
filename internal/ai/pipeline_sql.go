package ai

import (
	"context"
	"fmt"
)

// PipelineType represents the type of transformation pipeline
type PipelineType string

const (
	PipelineTypeNone         PipelineType = "none"
	PipelineTypeSQL          PipelineType = "sql"
	PipelineTypeEdgeFunction PipelineType = "edge_function"
	PipelineTypeWebhook      PipelineType = "webhook"
)

// SQLPipeline executes SQL-based document transformations
type SQLPipeline struct {
	storage *KnowledgeBaseStorage
}

// NewSQLPipeline creates a new SQL pipeline
func NewSQLPipeline(storage *KnowledgeBaseStorage) *SQLPipeline {
	return &SQLPipeline{
		storage: storage,
	}
}

// TransformResult represents the result of a document transformation
type TransformResult struct {
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	ShouldChunk bool                   `json:"should_chunk"`
}

// ExecuteTransform runs a SQL transformation function on document content
func (p *SQLPipeline) ExecuteTransform(ctx context.Context, kb *KnowledgeBase, document *Document) (*TransformResult, error) {
	if kb.PipelineType != string(PipelineTypeSQL) || kb.TransformationFunction == nil || *kb.TransformationFunction == "" {
		// No transformation configured
		return &TransformResult{
			Content:     document.Content,
			Metadata:    map[string]interface{}{},
			ShouldChunk: true,
		}, nil
	}

	// Execute the SQL transformation function
	// The function should accept (content TEXT, metadata JSONB) and return (content TEXT, metadata JSONB)
	query := fmt.Sprintf("SELECT * FROM %s($1, $2)", *kb.TransformationFunction)

	var (
		transformedContent  string
		transformedMetadata map[string]interface{}
	)

	err := p.storage.DB.QueryRow(ctx, query, document.Content, document.Metadata).Scan(
		&transformedContent,
		&transformedMetadata,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute transformation function %s: %w", *kb.TransformationFunction, err)
	}

	return &TransformResult{
		Content:     transformedContent,
		Metadata:    transformedMetadata,
		ShouldChunk: true,
	}, nil
}

// ValidateTransformFunction checks if a SQL transformation function exists and has the correct signature
func (p *SQLPipeline) ValidateTransformFunction(ctx context.Context, functionName string) error {
	// Check if the function exists
	query := `
		SELECT 1
		FROM pg_proc
		WHERE proname = $1
		  AND pronargs = 2
		  AND proargtypes::text[] @> ARRAY['text', 'jsonb']::regtype[]
		LIMIT 1
	`

	var exists int
	err := p.storage.DB.QueryRow(ctx, query, functionName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to validate transformation function: %w", err)
	}

	if exists != 1 {
		return fmt.Errorf("transformation function %s not found or has wrong signature (expected: (text, jsonb) RETURNS TABLE(content text, metadata jsonb))", functionName)
	}

	return nil
}
