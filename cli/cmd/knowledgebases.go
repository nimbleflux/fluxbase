package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/nimbleflux/fluxbase/cli/output"
	"github.com/nimbleflux/fluxbase/cli/util"
)

var kbCmd = &cobra.Command{
	Use:     "kb",
	Aliases: []string{"knowledge-bases", "knowledge-base"},
	Short:   "Manage knowledge bases",
	Long:    `Create and manage knowledge bases for AI chatbots.`,
}

var (
	kbDescription     string
	kbEmbeddingModel  string
	kbChunkSize       int
	kbDocTitle        string
	kbDocMetadata     string
	kbSearchLimit     int
	kbSearchThreshold float64
	kbNamespace       string

	// New flags for extended KB functionality
	kbDocTags              string // Comma-separated tags for upload/add
	kbDocLanguage          string // OCR languages
	kbDocContent           string // Inline content for add
	kbDocFromFile          string // Read content from file for add
	kbDocNewTitle          string // New title for update
	kbDocNewTags           string // New tags for update
	kbDocNewMetadata       string // New metadata for update
	kbDeleteFilterTags     string // Tag filter for delete-by-filter
	kbDeleteFilterMetadata string // Metadata filter for delete-by-filter
	kbTableSchema          string // Schema for export-table
	kbTableName            string // Table name for export-table
	kbTableColumns         string // Comma-separated column names for export-table
	kbTableIncludeFKs      bool   // Include foreign keys
	kbTableIncludeIdx      bool   // Include indexes
	kbTableSampleRows      int    // Sample rows for export-table
	kbEntityType           string // Entity type filter
	kbEntitySearch         string // Entity search query
	kbStatusOutput         string // Output format for status
)

var kbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all knowledge bases",
	Long: `List all knowledge bases.

Examples:
  fluxbase kb list
  fluxbase kb list -o json`,
	PreRunE: requireAuth,
	RunE:    runKBList,
}

var kbGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get knowledge base details",
	Long: `Get details of a specific knowledge base.

Examples:
  fluxbase kb get abc123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBGet,
}

var kbCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new knowledge base",
	Long: `Create a new knowledge base.

Examples:
  fluxbase kb create docs --description "Product documentation"
  fluxbase kb create docs --embedding-model text-embedding-ada-002`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBCreate,
}

var kbUpdateCmd = &cobra.Command{
	Use:   "update [id]",
	Short: "Update a knowledge base",
	Long: `Update an existing knowledge base.

Examples:
  fluxbase kb update abc123 --description "Updated description"`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBUpdate,
}

var kbDeleteCmd = &cobra.Command{
	Use:     "delete [id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a knowledge base",
	Long: `Delete a knowledge base and all its documents.

Examples:
  fluxbase kb delete abc123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBDelete,
}

var kbUploadCmd = &cobra.Command{
	Use:   "upload [id] [file]",
	Short: "Upload a document to a knowledge base",
	Long: `Upload a document to a knowledge base for indexing.

Supported formats: PDF, DOCX, TXT, MD, images (with OCR)

Examples:
  fluxbase kb upload abc123 ./docs/manual.pdf
  fluxbase kb upload abc123 ./docs/guide.md --title "User Guide"`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireAuth,
	RunE:    runKBUpload,
}

var kbDocumentsCmd = &cobra.Command{
	Use:   "documents [id]",
	Short: "List documents in a knowledge base",
	Long: `List all documents in a knowledge base.

Examples:
  fluxbase kb documents abc123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBDocuments,
}

var kbDocumentDeleteCmd = &cobra.Command{
	Use:   "delete [kb-id] [doc-id]",
	Short: "Delete a document from a knowledge base",
	Long: `Delete a specific document from a knowledge base.

Examples:
  fluxbase kb documents delete abc123 doc456`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireAuth,
	RunE:    runKBDocumentDelete,
}

var kbSearchCmd = &cobra.Command{
	Use:   "search [id] [query]",
	Short: "Search a knowledge base",
	Long: `Search a knowledge base using semantic similarity.

Examples:
  fluxbase kb search abc123 "how to reset password"
  fluxbase kb search abc123 "pricing plans" --limit 5`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireAuth,
	RunE:    runKBSearch,
}

var kbStatusCmd = &cobra.Command{
	Use:   "status [id]",
	Short: "Show knowledge base status and statistics",
	Long: `Show knowledge base status including document count, created date, and other statistics.

Examples:
  fluxbase kb status abc123
  fluxbase kb status abc123 --output json`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBStatus,
}

var kbAddCmd = &cobra.Command{
	Use:   "add [id]",
	Short: "Add document from text, stdin, or file",
	Long: `Add a document to a knowledge base from inline text, stdin, or a file.
This is an alternative to 'upload' for text-based content.

Examples:
  # Add from inline content
  fluxbase kb add abc123 --content "Hello world" --title "Greeting"

  # Add from stdin
  echo "Content" | fluxbase kb add abc123 --title "My Doc"

  # Add from file
  fluxbase kb add abc123 --from-file ./doc.txt --title "Document"

  # Add with user isolation
  fluxbase kb add abc123 --content "..." --title "User Doc" --metadata '{"user_id":"uuid-here"}'`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBAdd,
}

var kbDocumentGetCmd = &cobra.Command{
	Use:   "get [kb-id] [doc-id]",
	Short: "Get document details",
	Long: `Get detailed information about a specific document in a knowledge base.

Examples:
  fluxbase kb documents get abc123 doc456`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireAuth,
	RunE:    runKBDocumentGet,
}

var kbDocumentUpdateCmd = &cobra.Command{
	Use:   "update [kb-id] [doc-id]",
	Short: "Update document metadata",
	Long: `Update document metadata (title, tags, metadata).

Examples:
  fluxbase kb documents update abc123 doc456 --title "New Title"
  fluxbase kb documents update abc123 doc456 --tags "tag1,tag2"
  fluxbase kb documents update abc123 doc456 --metadata '{"key":"value"}'`,
	Args:    cobra.ExactArgs(2),
	PreRunE: requireAuth,
	RunE:    runKBDocumentUpdate,
}

var kbDocumentDeleteByFilterCmd = &cobra.Command{
	Use:   "delete-by-filter [kb-id]",
	Short: "Bulk delete documents by tags or metadata",
	Long: `Delete multiple documents from a knowledge base based on tag or metadata filters.

Examples:
  fluxbase kb documents delete-by-filter abc123 --tags "archived"
  fluxbase kb documents delete-by-filter abc123 --metadata-filter "user_id=uuid-here"`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBDocumentDeleteByFilter,
}

var kbExportTableCmd = &cobra.Command{
	Use:   "export-table [kb-id]",
	Short: "Export database table as document",
	Long: `Export a database table as a document to the knowledge base.
Includes schema, columns, relationships, and optional sample data.

Examples:
  # Export all columns
  fluxbase kb export-table abc123 --table users --schema public

  # Export specific columns (recommended for sensitive data)
  fluxbase kb export-table abc123 --table users --columns id,name,email

  # Include foreign keys and indexes
  fluxbase kb export-table abc123 --table products --include-fks --include-indexes --sample-rows 10`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBExportTable,
}

var kbTablesCmd = &cobra.Command{
	Use:   "tables [schema]",
	Short: "List exportable database tables",
	Long: `List all database tables that can be exported as documents.

Examples:
  fluxbase kb tables
  fluxbase kb tables public`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBTables,
}

var kbCapabilitiesCmd = &cobra.Command{
	Use:   "capabilities",
	Short: "Show system capabilities",
	Long: `Show system capabilities including supported OCR languages, file types, and features.

Examples:
  fluxbase kb capabilities`,
	Args:    cobra.NoArgs,
	PreRunE: requireAuth,
	RunE:    runKBCapabilities,
}

var kbGraphCmd = &cobra.Command{
	Use:   "graph [id]",
	Short: "Show knowledge graph data",
	Long: `Show the knowledge graph for a knowledge base, including entities and their relationships.

Examples:
  fluxbase kb graph abc123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBGraph,
}

var kbEntitiesCmd = &cobra.Command{
	Use:   "entities [id]",
	Short: "List entities in knowledge base",
	Long: `List entities extracted from the knowledge base.

Examples:
  fluxbase kb entities abc123
  fluxbase kb entities abc123 --type person
  fluxbase kb entities abc123 --search "John"`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBEntities,
}

var kbChatbotsCmd = &cobra.Command{
	Use:   "chatbots [id]",
	Short: "List chatbots using knowledge base",
	Long: `List all chatbots that are using the specified knowledge base.

Examples:
  fluxbase kb chatbots abc123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runKBChatbots,
}

func init() {
	// List flags
	kbListCmd.Flags().StringVar(&kbNamespace, "namespace", "", "Filter by namespace")

	// Create flags
	kbCreateCmd.Flags().StringVar(&kbDescription, "description", "", "Knowledge base description")
	kbCreateCmd.Flags().StringVar(&kbEmbeddingModel, "embedding-model", "", "Embedding model to use")
	kbCreateCmd.Flags().IntVar(&kbChunkSize, "chunk-size", 512, "Document chunk size")
	kbCreateCmd.Flags().StringVar(&kbNamespace, "namespace", "default", "Target namespace")

	// Update flags
	kbUpdateCmd.Flags().StringVar(&kbDescription, "description", "", "Knowledge base description")

	// Upload flags
	kbUploadCmd.Flags().StringVar(&kbDocTitle, "title", "", "Document title")
	kbUploadCmd.Flags().StringVar(&kbDocMetadata, "metadata", "", "Document metadata (JSON)")
	kbUploadCmd.Flags().StringVar(&kbDocTags, "tags", "", "Comma-separated tags")
	kbUploadCmd.Flags().StringVar(&kbDocLanguage, "ocr-languages", "", "OCR languages (e.g., 'eng,deu')")

	// Search flags
	kbSearchCmd.Flags().IntVar(&kbSearchLimit, "limit", 10, "Maximum results to return")
	kbSearchCmd.Flags().Float64Var(&kbSearchThreshold, "threshold", 0.7, "Similarity threshold (0.0-1.0)")

	// Status flags
	kbStatusCmd.Flags().StringVar(&kbStatusOutput, "output", "", "Output format (json, table)")

	// Add command flags
	kbAddCmd.Flags().StringVar(&kbDocContent, "content", "", "Inline document content")
	kbAddCmd.Flags().StringVar(&kbDocFromFile, "from-file", "", "Read content from file")
	kbAddCmd.Flags().StringVar(&kbDocTitle, "title", "", "Document title")
	kbAddCmd.Flags().StringVar(&kbDocMetadata, "metadata", "", "Document metadata (JSON)")
	kbAddCmd.Flags().StringVar(&kbDocTags, "tags", "", "Comma-separated tags")

	// Document update flags
	kbDocumentUpdateCmd.Flags().StringVar(&kbDocNewTitle, "title", "", "New document title")
	kbDocumentUpdateCmd.Flags().StringVar(&kbDocNewTags, "tags", "", "New tags (comma-separated)")
	kbDocumentUpdateCmd.Flags().StringVar(&kbDocNewMetadata, "metadata", "", "New metadata (JSON)")

	// Document delete-by-filter flags
	kbDocumentDeleteByFilterCmd.Flags().StringVar(&kbDeleteFilterTags, "tags", "", "Filter by tags (comma-separated)")
	kbDocumentDeleteByFilterCmd.Flags().StringVar(&kbDeleteFilterMetadata, "metadata-filter", "", "Filter by metadata (e.g., 'key=value')")

	// Export-table flags
	kbExportTableCmd.Flags().StringVar(&kbTableName, "table", "", "Table name (required)")
	kbExportTableCmd.Flags().StringVar(&kbTableSchema, "schema", "public", "Schema name")
	kbExportTableCmd.Flags().StringVar(&kbTableColumns, "columns", "", "Comma-separated column names (default: all columns)")
	kbExportTableCmd.Flags().BoolVar(&kbTableIncludeFKs, "include-fks", false, "Include foreign keys")
	kbExportTableCmd.Flags().BoolVar(&kbTableIncludeIdx, "include-indexes", false, "Include indexes")
	kbExportTableCmd.Flags().IntVar(&kbTableSampleRows, "sample-rows", 0, "Number of sample rows to include")

	// Entities flags
	kbEntitiesCmd.Flags().StringVar(&kbEntityType, "type", "", "Filter by entity type")
	kbEntitiesCmd.Flags().StringVar(&kbEntitySearch, "search", "", "Search entities by name")

	// Add document subcommands
	kbDocumentsCmd.AddCommand(kbDocumentDeleteCmd)
	kbDocumentsCmd.AddCommand(kbDocumentGetCmd)
	kbDocumentsCmd.AddCommand(kbDocumentUpdateCmd)
	kbDocumentsCmd.AddCommand(kbDocumentDeleteByFilterCmd)

	// Add all commands
	kbCmd.AddCommand(kbListCmd)
	kbCmd.AddCommand(kbGetCmd)
	kbCmd.AddCommand(kbCreateCmd)
	kbCmd.AddCommand(kbUpdateCmd)
	kbCmd.AddCommand(kbDeleteCmd)
	kbCmd.AddCommand(kbStatusCmd)
	kbCmd.AddCommand(kbUploadCmd)
	kbCmd.AddCommand(kbAddCmd)
	kbCmd.AddCommand(kbDocumentsCmd)
	kbCmd.AddCommand(kbSearchCmd)
	kbCmd.AddCommand(kbExportTableCmd)
	kbCmd.AddCommand(kbTablesCmd)
	kbCmd.AddCommand(kbCapabilitiesCmd)
	kbCmd.AddCommand(kbGraphCmd)
	kbCmd.AddCommand(kbEntitiesCmd)
	kbCmd.AddCommand(kbChatbotsCmd)
}

func runKBList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build query parameters for namespace filter
	var params url.Values
	if kbNamespace != "" {
		params = url.Values{}
		params.Add("namespace", kbNamespace)
	}

	var response struct {
		KnowledgeBases []map[string]interface{} `json:"knowledge_bases"`
		Count          int                      `json:"count"`
	}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases", params, &response); err != nil {
		return err
	}
	kbs := response.KnowledgeBases

	if len(kbs) == 0 {
		fmt.Println("No knowledge bases found.")
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		data := output.TableData{
			Headers: []string{"ID", "NAME", "NAMESPACE", "DOCUMENTS", "CREATED"},
			Rows:    make([][]string, len(kbs)),
		}

		for i, kb := range kbs {
			id := getStringValue(kb, "id")
			name := getStringValue(kb, "name")
			namespace := getStringValue(kb, "namespace")
			docs := fmt.Sprintf("%d", getIntValue(kb, "document_count"))
			created := getStringValue(kb, "created_at")

			data.Rows[i] = []string{id, name, namespace, docs, created}
		}

		formatter.PrintTable(data)
	} else {
		if err := formatter.Print(kbs); err != nil {
			return err
		}
	}

	return nil
}

func runKBGet(cmd *cobra.Command, args []string) error {
	id := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var kb map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(id), nil, &kb); err != nil {
		return err
	}

	formatter := GetFormatter()
	return formatter.Print(kb)
}

func runKBCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"name":       name,
		"namespace":  kbNamespace,
		"chunk_size": kbChunkSize,
	}

	if kbDescription != "" {
		body["description"] = kbDescription
	}
	if kbEmbeddingModel != "" {
		body["embedding_model"] = kbEmbeddingModel
	}

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/ai/knowledge-bases", body, &result); err != nil {
		return err
	}

	id := getStringValue(result, "id")
	fmt.Printf("Knowledge base '%s' created with ID: %s\n", name, id)
	return nil
}

func runKBUpdate(cmd *cobra.Command, args []string) error {
	id := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body := make(map[string]interface{})

	if kbDescription != "" {
		body["description"] = kbDescription
	}

	if len(body) == 0 {
		return fmt.Errorf("no updates specified")
	}

	if err := apiClient.DoPut(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(id), body, nil); err != nil {
		return err
	}

	fmt.Printf("Knowledge base '%s' updated.\n", id)
	return nil
}

func runKBDelete(cmd *cobra.Command, args []string) error {
	id := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := apiClient.DoDelete(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(id)); err != nil {
		return err
	}

	fmt.Printf("Knowledge base '%s' deleted.\n", id)
	return nil
}

func runKBUpload(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	filePath := args[1]

	// Read file
	file, err := os.Open(filePath) //nolint:gosec // CLI tool reads user-provided file path
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	// Add title if specified
	if kbDocTitle != "" {
		if err := writer.WriteField("title", kbDocTitle); err != nil {
			return err
		}
	}

	// Add metadata if specified
	if kbDocMetadata != "" {
		if err := writer.WriteField("metadata", kbDocMetadata); err != nil {
			return err
		}
	}

	if err := writer.Close(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build request - use /upload endpoint for multipart uploads
	uploadURL := apiClient.BaseURL + "/api/v1/admin/ai/knowledge-bases/" + url.PathEscape(kbID) + "/documents/upload"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Add auth
	creds, err := apiClient.CredentialManager.GetCredentials(apiClient.Profile.Name)
	if err != nil {
		return err
	}
	if creds != nil && creds.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+creds.AccessToken)
	} else if creds != nil && creds.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+creds.APIKey)
	}

	resp, err := apiClient.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		// Non-JSON response is OK
		fmt.Printf("Uploaded '%s' to knowledge base '%s' (%s)\n", filepath.Base(filePath), kbID, util.FormatBytes(fileInfo.Size()))
		return nil
	}

	docID := getStringValue(result, "id")
	fmt.Printf("Uploaded '%s' to knowledge base '%s' (Document ID: %s, %s)\n", filepath.Base(filePath), kbID, docID, util.FormatBytes(fileInfo.Size()))
	return nil
}

func runKBDocuments(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// API returns wrapped response: {"documents": [...], "count": N}
	var response struct {
		Documents []map[string]interface{} `json:"documents"`
		Count     int                      `json:"count"`
	}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/documents", nil, &response); err != nil {
		return err
	}
	docs := response.Documents

	if len(docs) == 0 {
		fmt.Println("No documents found.")
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		data := output.TableData{
			Headers: []string{"ID", "TITLE", "TYPE", "CHUNKS", "STATUS"},
			Rows:    make([][]string, len(docs)),
		}

		for i, doc := range docs {
			id := getStringValue(doc, "id")
			title := getStringValue(doc, "title")
			if title == "" {
				title = getStringValue(doc, "filename")
			}
			docType := getStringValue(doc, "file_type")
			if docType == "" {
				docType = getStringValue(doc, "content_type")
			}
			chunks := fmt.Sprintf("%d", getIntValue(doc, "chunk_count"))
			status := getStringValue(doc, "status")

			data.Rows[i] = []string{id, title, docType, chunks, status}
		}

		formatter.PrintTable(data)
	} else {
		if err := formatter.Print(docs); err != nil {
			return err
		}
	}

	return nil
}

func runKBDocumentDelete(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	docID := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deletePath := fmt.Sprintf("/api/v1/admin/ai/knowledge-bases/%s/documents/%s", url.PathEscape(kbID), url.PathEscape(docID))

	if err := apiClient.DoDelete(ctx, deletePath); err != nil {
		return err
	}

	fmt.Printf("Document '%s' deleted from knowledge base '%s'.\n", docID, kbID)
	return nil
}

func runKBSearch(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	query := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"query":     query,
		"limit":     kbSearchLimit,
		"threshold": kbSearchThreshold,
	}

	var results []map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/search", body, &results); err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		for i, result := range results {
			score := result["score"]
			content := getStringValue(result, "content")
			docTitle := getStringValue(result, "document_title")

			fmt.Printf("%d. [%.2f] %s\n", i+1, score, docTitle)
			// Truncate content for display
			if len(content) > 200 {
				content = content[:200] + "..."
			}
			fmt.Printf("   %s\n\n", content)
		}
	} else {
		if err := formatter.Print(results); err != nil {
			return err
		}
	}

	return nil
}

func runKBStatus(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var kb map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/status", nil, &kb); err != nil {
		return err
	}

	formatter := GetFormatter()
	if kbStatusOutput == "json" || formatter.Format != output.FormatTable {
		if err := formatter.Print(kb); err != nil {
			return err
		}
		return nil
	}

	// Table format output
	fmt.Printf("Knowledge Base Status: %s\n", kbID)
	fmt.Printf("  Exists: %v\n", getBoolValue(kb, "exists"))
	fmt.Printf("  Document Count: %d\n", getIntValue(kb, "document_count"))
	fmt.Printf("  Total Chunks: %d\n", getIntValue(kb, "total_chunks"))
	fmt.Printf("  Created At: %s\n", getStringValue(kb, "created_at"))
	fmt.Printf("  Updated At: %s\n", getStringValue(kb, "updated_at"))
	if embeddingModel := getStringValue(kb, "embedding_model"); embeddingModel != "" {
		fmt.Printf("  Embedding Model: %s\n", embeddingModel)
	}
	if chunkSize := getIntValue(kb, "chunk_size"); chunkSize > 0 {
		fmt.Printf("  Chunk Size: %d\n", chunkSize)
	}

	return nil
}

func runKBAdd(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	// Determine content source
	var content string

	switch {
	case kbDocContent != "":
		content = kbDocContent
	case kbDocFromFile != "":
		data, err := os.ReadFile(kbDocFromFile) //nolint:gosec // CLI tool reads user-provided file path
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content = string(data)
	default:
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		if len(data) == 0 {
			return fmt.Errorf("no content provided. Use --content, --from-file, or pipe from stdin")
		}
		content = string(data)
	}

	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	// Build request body
	body := map[string]interface{}{
		"content": content,
	}

	if kbDocTitle != "" {
		body["title"] = kbDocTitle
	}
	if kbDocMetadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(kbDocMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		body["metadata"] = metadata
	}
	if kbDocTags != "" {
		body["tags"] = kbDocTags
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/documents", body, &result); err != nil {
		return err
	}

	docID := getStringValue(result, "id")
	title := getStringValue(result, "title")
	if title == "" {
		title = kbDocTitle
	}
	if title == "" {
		title = "(untitled)"
	}

	fmt.Printf("Document '%s' added to knowledge base '%s' (ID: %s)\n", title, kbID, docID)
	return nil
}

func runKBDocumentGet(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	docID := args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var doc map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/documents/"+url.PathEscape(docID), nil, &doc); err != nil {
		return err
	}

	formatter := GetFormatter()
	return formatter.Print(doc)
}

func runKBDocumentUpdate(cmd *cobra.Command, args []string) error {
	kbID := args[0]
	docID := args[1]

	body := make(map[string]interface{})

	if kbDocNewTitle != "" {
		body["title"] = kbDocNewTitle
	}
	if kbDocNewTags != "" {
		body["tags"] = kbDocNewTags
	}
	if kbDocNewMetadata != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(kbDocNewMetadata), &metadata); err != nil {
			return fmt.Errorf("invalid metadata JSON: %w", err)
		}
		body["metadata"] = metadata
	}

	if len(body) == 0 {
		return fmt.Errorf("no updates specified. Use --title, --tags, or --metadata")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := apiClient.DoPut(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/documents/"+url.PathEscape(docID), body, nil); err != nil {
		return err
	}

	fmt.Printf("Document '%s' updated.\n", docID)
	return nil
}

func runKBDocumentDeleteByFilter(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	if kbDeleteFilterTags == "" && kbDeleteFilterMetadata == "" {
		return fmt.Errorf("at least one filter is required. Use --tags or --metadata-filter")
	}

	// Build query parameters
	params := url.Values{}
	if kbDeleteFilterTags != "" {
		params.Add("tags", kbDeleteFilterTags)
	}
	if kbDeleteFilterMetadata != "" {
		params.Add("metadata_filter", kbDeleteFilterMetadata)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var result map[string]interface{}
	if err := apiClient.DoPostWithQuery(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/documents/delete-by-filter", nil, params, &result); err != nil {
		return err
	}

	count := getIntValue(result, "deleted_count")
	fmt.Printf("Deleted %d document(s) from knowledge base '%s'.\n", count, kbID)
	return nil
}

func runKBExportTable(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	if kbTableName == "" {
		return fmt.Errorf("--table is required")
	}

	body := map[string]interface{}{
		"table":                kbTableName,
		"include_foreign_keys": kbTableIncludeFKs,
		"include_indexes":      kbTableIncludeIdx,
		"include_sample_rows":  kbTableSampleRows > 0,
	}

	if kbTableSchema != "" {
		body["schema"] = kbTableSchema
	}
	if kbTableSampleRows > 0 {
		body["sample_row_count"] = kbTableSampleRows
	}
	if kbTableColumns != "" {
		columns := strings.Split(kbTableColumns, ",")
		// Trim whitespace from each column name
		for i, col := range columns {
			columns[i] = strings.TrimSpace(col)
		}
		body["columns"] = columns
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/tables/export", body, &result); err != nil {
		return err
	}

	docID := getStringValue(result, "document_id")
	entityID := getStringValue(result, "entity_id")

	if kbTableColumns != "" {
		fmt.Printf("Table '%s.%s' exported with columns [%s] as document (ID: %s, Entity: %s)\n",
			kbTableSchema, kbTableName, kbTableColumns, docID, entityID)
	} else {
		fmt.Printf("Table '%s.%s' exported as document (ID: %s, Entity: %s)\n",
			kbTableSchema, kbTableName, docID, entityID)
	}
	return nil
}

func runKBTables(cmd *cobra.Command, args []string) error {
	schema := "public"
	if len(args) > 0 {
		schema = args[0]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := url.Values{}
	params.Add("schema", schema)

	var response struct {
		Tables []map[string]interface{} `json:"tables"`
		Count  int                      `json:"count"`
	}
	path := "/api/v1/admin/ai/tables?" + params.Encode()
	if err := apiClient.DoGet(ctx, path, nil, &response); err != nil {
		return err
	}

	tables := response.Tables

	if len(tables) == 0 {
		fmt.Printf("No tables found in schema '%s'.\n", schema)
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		data := output.TableData{
			Headers: []string{"SCHEMA", "TABLE", "COLUMNS", "ROWS (approx)"},
			Rows:    make([][]string, len(tables)),
		}

		for i, table := range tables {
			schemaName := getStringValue(table, "schema")
			tableName := getStringValue(table, "name")
			columns := fmt.Sprintf("%d", getIntValue(table, "column_count"))
			rows := fmt.Sprintf("%d", getIntValue(table, "row_estimate"))

			data.Rows[i] = []string{schemaName, tableName, columns, rows}
		}

		formatter.PrintTable(data)
	} else {
		if err := formatter.Print(tables); err != nil {
			return err
		}
	}

	return nil
}

func runKBCapabilities(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var caps map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/capabilities", nil, &caps); err != nil {
		return err
	}

	formatter := GetFormatter()
	if formatter.Format != output.FormatTable {
		if err := formatter.Print(caps); err != nil {
			return err
		}
		return nil
	}

	// Table format output
	fmt.Println("AI Knowledge Base Capabilities")
	fmt.Println()

	if ocr, ok := caps["ocr"].(map[string]interface{}); ok {
		fmt.Println("OCR:")
		if enabled, ok := ocr["enabled"].(bool); ok && enabled {
			fmt.Println("  Status: Enabled")
			if langs, ok := ocr["languages"].([]interface{}); ok {
				fmt.Printf("  Languages: %v\n", langs)
			}
		} else {
			fmt.Println("  Status: Disabled")
		}
		fmt.Println()
	}

	if fileTypes, ok := caps["supported_file_types"].([]interface{}); ok {
		fmt.Println("Supported File Types:")
		for _, ft := range fileTypes {
			fmt.Printf("  - %v\n", ft)
		}
		fmt.Println()
	}

	if features, ok := caps["features"].([]interface{}); ok {
		fmt.Println("Features:")
		for _, f := range features {
			fmt.Printf("  - %v\n", f)
		}
		fmt.Println()
	}

	if limits, ok := caps["limits"].(map[string]interface{}); ok {
		fmt.Println("Limits:")
		for k, v := range limits {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}

	return nil
}

func runKBGraph(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var graph map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/graph", nil, &graph); err != nil {
		return err
	}

	formatter := GetFormatter()
	if formatter.Format != output.FormatTable {
		if err := formatter.Print(graph); err != nil {
			return err
		}
		return nil
	}

	// Table format output
	fmt.Printf("Knowledge Graph for: %s\n", kbID)
	fmt.Println()

	if nodes, ok := graph["nodes"].([]interface{}); ok {
		fmt.Printf("Entities: %d\n", len(nodes))
	}
	if edges, ok := graph["edges"].([]interface{}); ok {
		fmt.Printf("Relationships: %d\n", len(edges))
	}

	return nil
}

func runKBEntities(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	params := url.Values{}
	if kbEntityType != "" {
		params.Add("type", kbEntityType)
	}
	if kbEntitySearch != "" {
		params.Add("search", kbEntitySearch)
	}

	// Only pass params if we have any
	var queryParams url.Values
	if len(params) > 0 {
		queryParams = params
	}

	var response struct {
		Entities []map[string]interface{} `json:"entities"`
		Count    int                      `json:"count"`
	}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/entities", queryParams, &response); err != nil {
		return err
	}

	entities := response.Entities

	if len(entities) == 0 {
		fmt.Println("No entities found.")
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		data := output.TableData{
			Headers: []string{"ID", "TYPE", "NAME", "COUNT"},
			Rows:    make([][]string, len(entities)),
		}

		for i, entity := range entities {
			id := getStringValue(entity, "id")
			entityType := getStringValue(entity, "type")
			name := getStringValue(entity, "name")
			count := fmt.Sprintf("%d", getIntValue(entity, "count"))

			data.Rows[i] = []string{id, entityType, name, count}
		}

		formatter.PrintTable(data)
	} else {
		if err := formatter.Print(entities); err != nil {
			return err
		}
	}

	return nil
}

func runKBChatbots(cmd *cobra.Command, args []string) error {
	kbID := args[0]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var response struct {
		Chatbots []map[string]interface{} `json:"chatbots"`
		Count    int                      `json:"count"`
	}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/ai/knowledge-bases/"+url.PathEscape(kbID)+"/chatbots", nil, &response); err != nil {
		return err
	}

	chatbots := response.Chatbots

	if len(chatbots) == 0 {
		fmt.Printf("No chatbots found using knowledge base '%s'.\n", kbID)
		return nil
	}

	formatter := GetFormatter()

	if formatter.Format == output.FormatTable {
		data := output.TableData{
			Headers: []string{"ID", "NAME", "STATUS", "CREATED"},
			Rows:    make([][]string, len(chatbots)),
		}

		for i, cb := range chatbots {
			id := getStringValue(cb, "id")
			name := getStringValue(cb, "name")
			status := getStringValue(cb, "status")
			created := getStringValue(cb, "created_at")

			data.Rows[i] = []string{id, name, status, created}
		}

		formatter.PrintTable(data)
	} else {
		if err := formatter.Print(chatbots); err != nil {
			return err
		}
	}

	return nil
}

// Helper function for boolean values
func getBoolValue(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
