package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/fluxbase-eu/fluxbase/internal/database"
)

// TableExporter exports database tables as knowledge base documents and entities
type TableExporter struct {
	conn           *database.Connection
	processor      *DocumentProcessor
	knowledgeGraph *KnowledgeGraph
	storage        *KnowledgeBaseStorage
}

// NewTableExporter creates a new table exporter
func NewTableExporter(
	conn *database.Connection,
	processor *DocumentProcessor,
	knowledgeGraph *KnowledgeGraph,
	storage *KnowledgeBaseStorage,
) *TableExporter {
	return &TableExporter{
		conn:           conn,
		processor:      processor,
		knowledgeGraph: knowledgeGraph,
		storage:        storage,
	}
}

// ExportTableRequest contains options for table export
type ExportTableRequest struct {
	KnowledgeBaseID    string   `json:"knowledge_base_id"`
	Schema             string   `json:"schema"`
	Table              string   `json:"table"`
	Columns            []string `json:"columns,omitempty"` // Optional: specific columns to export (nil/empty = all)
	IncludeSampleRows  bool     `json:"include_sample_rows"`
	SampleRowCount     int      `json:"sample_row_count"`
	IncludeForeignKeys bool     `json:"include_foreign_keys"`
	IncludeIndexes     bool     `json:"include_indexes"`
	OwnerID            *string  `json:"owner_id,omitempty"` // Document owner (for RLS)
}

// ExportTableResult contains the export results
type ExportTableResult struct {
	DocumentID      string   `json:"document_id"`
	EntityID        string   `json:"entity_id"`
	RelationshipIDs []string `json:"relationship_ids,omitempty"`
}

// ExportTable exports a single table as a document and entity
func (e *TableExporter) ExportTable(ctx context.Context, req ExportTableRequest) (*ExportTableResult, error) {
	inspector := database.NewSchemaInspector(e.conn)

	// Get table metadata
	tableInfo, err := inspector.GetTableInfo(ctx, req.Schema, req.Table)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}
	if tableInfo == nil {
		return nil, fmt.Errorf("table not found: %s.%s", req.Schema, req.Table)
	}

	// Validate column names if specific columns are requested
	if len(req.Columns) > 0 {
		validColumns := make(map[string]bool)
		for _, col := range tableInfo.Columns {
			validColumns[col.Name] = true
		}
		for _, col := range req.Columns {
			if !validColumns[col] {
				return nil, fmt.Errorf("column %q not found in table %s.%s", col, req.Schema, req.Table)
			}
		}
	}

	// Generate document content from schema
	docContent := e.generateTableDocument(tableInfo, req)

	// Create document
	doc := &Document{
		ID:              uuid.New().String(),
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           fmt.Sprintf("%s.%s", req.Schema, req.Table),
		Content:         docContent,
		SourceType:      "database_table",
		MimeType:        "text/markdown",
		Tags:            []string{"schema", "database", req.Schema, req.Table},
		OwnerID:         req.OwnerID,
	}

	// Create metadata
	exportedColumnCount := len(tableInfo.Columns)
	if len(req.Columns) > 0 {
		exportedColumnCount = len(req.Columns)
	}
	metadataMap := map[string]interface{}{
		"schema":           req.Schema,
		"table":            req.Table,
		"entity_type":      "table",
		"source":           "database_export",
		"table_type":       tableInfo.Type,
		"rls_enabled":      tableInfo.RLSEnabled,
		"exported_columns": exportedColumnCount,
		"total_columns":    len(tableInfo.Columns),
		"columns_filtered": len(req.Columns) > 0,
	}
	if len(req.Columns) > 0 {
		metadataMap["columns"] = req.Columns
	}
	if metadataJSON, err := metadataToJSON(metadataMap); err == nil {
		doc.Metadata = metadataJSON
	}

	if err := e.storage.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Process document (chunks + embeddings)
	opts := ProcessDocumentOptions{
		ChunkSize:     512,
		ChunkOverlap:  50,
		ChunkStrategy: ChunkingStrategyRecursive,
	}
	if err := e.processor.ProcessDocument(ctx, doc, opts); err != nil {
		return nil, fmt.Errorf("failed to process document: %w", err)
	}

	result := &ExportTableResult{
		DocumentID: doc.ID,
	}

	// Create table entity
	tableEntity := &Entity{
		ID:              uuid.New().String(),
		KnowledgeBaseID: req.KnowledgeBaseID,
		EntityType:      EntityTable,
		Name:            fmt.Sprintf("%s.%s", req.Schema, req.Table),
		CanonicalName:   fmt.Sprintf("%s.%s", req.Schema, req.Table),
		Aliases:         []string{req.Table},
		Metadata: map[string]interface{}{
			"schema":             req.Schema,
			"table":              req.Table,
			"column_count":       exportedColumnCount,
			"total_column_count": len(tableInfo.Columns),
			"primary_key":        tableInfo.PrimaryKey,
			"table_type":         tableInfo.Type,
		},
	}

	if err := e.knowledgeGraph.AddEntity(ctx, tableEntity); err != nil {
		return nil, fmt.Errorf("failed to add table entity: %w", err)
	}
	result.EntityID = tableEntity.ID

	// Create foreign key relationships
	if req.IncludeForeignKeys {
		for _, fk := range tableInfo.ForeignKeys {
			// Create entity for referenced table (if different)
			refTableName := fmt.Sprintf("%s.%s", req.Schema, fk.ReferencedTable)
			refEntity := &Entity{
				ID:              uuid.New().String(),
				KnowledgeBaseID: req.KnowledgeBaseID,
				EntityType:      EntityTable,
				Name:            refTableName,
				CanonicalName:   refTableName,
				Aliases:         []string{fk.ReferencedTable},
			}
			_ = e.knowledgeGraph.AddEntity(ctx, refEntity)

			// Create relationship
			rel := &EntityRelationship{
				ID:               uuid.New().String(),
				KnowledgeBaseID:  req.KnowledgeBaseID,
				SourceEntityID:   tableEntity.ID,
				TargetEntityID:   refEntity.ID,
				RelationshipType: RelForeignKey,
				Direction:        DirectionForward,
				Metadata: map[string]interface{}{
					"column":            fk.ColumnName,
					"referenced_column": fk.ReferencedColumn,
					"on_delete":         fk.OnDelete,
					"on_update":         fk.OnUpdate,
				},
			}
			if err := e.knowledgeGraph.AddRelationship(ctx, rel); err != nil {
				return nil, fmt.Errorf("failed to add FK relationship: %w", err)
			}
			result.RelationshipIDs = append(result.RelationshipIDs, rel.ID)
		}
	}

	return result, nil
}

// ListExportableTables lists all tables that can be exported
func (e *TableExporter) ListExportableTables(ctx context.Context, schemas []string) ([]database.TableInfo, error) {
	inspector := database.NewSchemaInspector(e.conn)
	return inspector.GetAllTables(ctx, schemas...)
}

// generateTableDocument creates a readable markdown document from table metadata
func (e *TableExporter) generateTableDocument(table *database.TableInfo, req ExportTableRequest) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# Table: %s.%s\n\n", table.Schema, table.Name))

	// Description
	sb.WriteString("## Description\n\n")
	sb.WriteString(fmt.Sprintf("Database **%s** in schema `%s`.\n\n", table.Type, table.Schema))
	if table.RLSEnabled {
		sb.WriteString("**Note:** Row Level Security (RLS) is enabled on this table.\n\n")
	}

	// Primary key
	if len(table.PrimaryKey) > 0 {
		sb.WriteString(fmt.Sprintf("**Primary Key:** `%s`\n\n", strings.Join(table.PrimaryKey, "`, `")))
	}

	// Filter columns if specific ones are requested
	columns := table.Columns
	if len(req.Columns) > 0 {
		columnSet := make(map[string]bool)
		for _, col := range req.Columns {
			columnSet[col] = true
		}
		filtered := make([]database.ColumnInfo, 0, len(req.Columns))
		for _, col := range table.Columns {
			if columnSet[col.Name] {
				filtered = append(filtered, col)
			}
		}
		columns = filtered
		sb.WriteString(fmt.Sprintf("*Exporting %d of %d columns.*\n\n", len(columns), len(table.Columns)))
	}

	// Columns
	sb.WriteString("## Columns\n\n")
	sb.WriteString("| Column | Type | Nullable | Default | Description |\n")
	sb.WriteString("|--------|------|----------|---------|-------------|\n")

	for _, col := range columns {
		nullable := "NOT NULL"
		if col.IsNullable {
			nullable = "NULL"
		}
		defaultVal := ""
		if col.DefaultValue != nil {
			defaultVal = *col.DefaultValue
		}

		// Add special markers
		prefix := ""
		suffix := ""
		if col.IsPrimaryKey {
			prefix += "🔑 "
		}
		if col.IsForeignKey {
			prefix += "🔗 "
		}
		if col.IsUnique {
			suffix += " 🦄"
		}

		sb.WriteString(fmt.Sprintf("| %s%s | %s | %s | %s | %s%s |\n",
			prefix, col.Name, col.DataType, nullable, defaultVal, col.Description, suffix))
	}
	sb.WriteString("\n")

	// JSONB Column Schemas
	for _, col := range columns {
		if col.JSONBSchema != nil && (col.DataType == "jsonb" || col.DataType == "json") {
			sb.WriteString(fmt.Sprintf("## JSONB Column: `%s`\n\n", col.Name))
			sb.WriteString("This JSONB column has the following structure:\n\n")
			sb.WriteString("| Field | Type | Required | Description |\n")
			sb.WriteString("|-------|------|----------|-------------|\n")

			for name, prop := range col.JSONBSchema.Properties {
				required := "no"
				for _, r := range col.JSONBSchema.Required {
					if r == name {
						required = "yes"
						break
					}
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
					name, prop.Type, required, prop.Description))
			}
			sb.WriteString("\n")

			// Add query examples
			sb.WriteString("**Query examples:**\n\n")
			sb.WriteString("```sql\n")
			// Find a field to use in example
			exampleField := ""
			for name := range col.JSONBSchema.Properties {
				exampleField = name
				break
			}
			if exampleField != "" {
				sb.WriteString(fmt.Sprintf("-- Filter by %s\n", exampleField))
				sb.WriteString(fmt.Sprintf("SELECT * FROM %s.%s WHERE %s->>'%s' = 'value';\n",
					table.Schema, table.Name, col.Name, exampleField))
				sb.WriteString("-- Use JSON path for nested fields\n")
				sb.WriteString(fmt.Sprintf("SELECT * FROM %s.%s WHERE %s->'nested'->>'field' = 'value';\n",
					table.Schema, table.Name, col.Name))
			}
			sb.WriteString("```\n\n")
		}
	}

	// Foreign keys
	if req.IncludeForeignKeys && len(table.ForeignKeys) > 0 {
		sb.WriteString("## Foreign Keys\n\n")
		for _, fk := range table.ForeignKeys {
			sb.WriteString(fmt.Sprintf("- `%s` → `%s.%s` (`%s`)", fk.ColumnName, fk.ReferencedTable, fk.ReferencedColumn, fk.ReferencedTable))
			if fk.OnDelete != "" {
				sb.WriteString(fmt.Sprintf(" ON DELETE %s", fk.OnDelete))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Indexes
	if req.IncludeIndexes && len(table.Indexes) > 0 {
		sb.WriteString("## Indexes\n\n")
		for _, idx := range table.Indexes {
			prefix := ""
			if idx.IsUnique {
				prefix = "UNIQUE "
			}
			sb.WriteString(fmt.Sprintf("- %s`%s` on (`%s`)\n",
				prefix, idx.Name, strings.Join(idx.Columns, "`, `")))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// metadataToJSON converts a map to JSON bytes using standard json.Marshal
func metadataToJSON(m map[string]interface{}) ([]byte, error) {
	if len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}
