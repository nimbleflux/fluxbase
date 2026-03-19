package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/rs/zerolog/log"
)

// SearchVectorsTool implements the search_vectors MCP tool
type SearchVectorsTool struct {
	ragService *ai.RAGService
}

// NewSearchVectorsTool creates a new search_vectors tool
func NewSearchVectorsTool(ragService *ai.RAGService) *SearchVectorsTool {
	return &SearchVectorsTool{
		ragService: ragService,
	}
}

func (t *SearchVectorsTool) Name() string {
	return "search_vectors"
}

func (t *SearchVectorsTool) Description() string {
	return "Search for semantically similar content using vector embeddings"
}

func (t *SearchVectorsTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "The search query to find similar content",
			},
			"chatbot_id": map[string]any{
				"type":        "string",
				"description": "The chatbot ID (optional, read from context if not provided)",
			},
			"knowledge_bases": map[string]any{
				"type":        "array",
				"description": "Optional list of specific knowledge base names to search",
				"items": map[string]any{
					"type": "string",
				},
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results (default: 5, max: 20)",
				"default":     5,
				"maximum":     20,
			},
			"threshold": map[string]any{
				"type":        "number",
				"description": "Minimum similarity threshold 0-1 (default: 0.7)",
				"default":     0.7,
			},
			"tags": map[string]any{
				"type":        "array",
				"description": "Optional tags to filter results (documents must have ALL specified tags)",
				"items": map[string]any{
					"type": "string",
				},
			},
			"metadata_filter": map[string]any{
				"type":        "object",
				"description": "Advanced metadata filter with operators and logical combinations",
				"properties": map[string]any{
					"conditions": map[string]any{
						"type":        "array",
						"description": "Filter conditions combined with logical_op (default: AND)",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"key": map[string]any{
									"type":        "string",
									"description": "Metadata key to filter on",
								},
								"operator": map[string]any{
									"type":        "string",
									"description": "Comparison operator (=, !=, ILIKE, LIKE, IN, NOT IN, >, >=, <, <=, BETWEEN, IS NULL, IS NOT NULL)",
									"enum":        []string{"=", "!=", "ILIKE", "LIKE", "IN", "NOT IN", ">", ">=", "<", "<=", "BETWEEN", "IS NULL", "IS NOT NULL"},
								},
								"value": map[string]any{
									"type":        "string, number, or boolean",
									"description": "Single value for =, !=, >, >=, <, <= operators",
								},
								"values": map[string]any{
									"type":        "array",
									"description": "Array of values for IN, NOT IN operators",
									"items": map[string]any{
										"type": "string, number, or boolean",
									},
								},
								"min": map[string]any{
									"type":        "number",
									"description": "Minimum value for BETWEEN operator",
								},
								"max": map[string]any{
									"type":        "number",
									"description": "Maximum value for BETWEEN operator",
								},
							},
							"required": []string{"key", "operator"},
						},
					},
					"logical_op": map[string]any{
						"type":        "string",
						"description": "How to combine conditions (AND or OR)",
						"enum":        []string{"AND", "OR"},
						"default":     "AND",
					},
					"groups": map[string]any{
						"type":        "array",
						"description": "Nested filter groups for complex queries",
					},
				},
			},
		},
		"required": []string{"query"}, // All other parameters are optional
	}
}

func (t *SearchVectorsTool) RequiredScopes() []string {
	return []string{mcp.ScopeReadVectors}
}

func (t *SearchVectorsTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	if t.ragService == nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("Vector search is not configured")},
			IsError: true,
		}, nil
	}

	// Parse query
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	// Parse chatbot_id (optional)
	chatbotID, _ := args["chatbot_id"].(string)

	// Parse knowledge bases
	var knowledgeBases []string
	if kbs, ok := args["knowledge_bases"].([]any); ok {
		for _, kb := range kbs {
			if kbStr, ok := kb.(string); ok {
				knowledgeBases = append(knowledgeBases, kbStr)
			}
		}
	}

	// Parse limit
	limit := 5
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 20 {
			limit = 20
		}
	}

	// Parse threshold
	threshold := 0.7
	if th, ok := args["threshold"].(float64); ok {
		threshold = th
	}

	// Parse tags
	var tags []string
	if t, ok := args["tags"].([]any); ok {
		for _, tag := range t {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	// Parse metadata_filter (advanced filtering with operators)
	var metadataFilter *ai.MetadataFilterGroup
	if mf, ok := args["metadata_filter"].(map[string]any); ok {
		filterGroup, err := parseMetadataFilter(mf)
		if err != nil {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Invalid metadata_filter: %v", err))},
				IsError: true,
			}, nil
		}
		metadataFilter = filterGroup
	}

	// Determine which chatbot to use
	// Priority: chatbot_id > context
	var effectiveChatbotID string
	if chatbotID != "" {
		effectiveChatbotID = chatbotID
	} else {
		// Fall back to context (set by ChatbotAuthContext)
		effectiveChatbotID = authCtx.GetMetadataString(mcp.MetadataKeyChatbotID)
		if effectiveChatbotID == "" {
			return nil, fmt.Errorf("chatbot_id must be specified when not using context")
		}
	}

	log.Debug().
		Str("query", query).
		Str("chatbot_id", effectiveChatbotID).
		Int("limit", limit).
		Float64("threshold", threshold).
		Msg("MCP: Searching vectors")

	// Build search options
	opts := ai.VectorSearchOptions{
		Query:          query,
		ChatbotID:      effectiveChatbotID,
		KnowledgeBases: knowledgeBases,
		Limit:          limit,
		Threshold:      threshold,
		Tags:           tags,
		MetadataFilter: metadataFilter,
	}

	// Add user context for filtering
	if authCtx.UserID != nil {
		opts.UserID = authCtx.UserID
	}

	// Check if user has admin access (service_role bypasses user filtering)
	if authCtx.UserRole == "service_role" || authCtx.UserRole == "instance_admin" {
		opts.IsAdmin = true
	}

	// Execute search
	results, err := t.ragService.VectorSearch(ctx, opts)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Vector search failed: %v", err))},
			IsError: true,
		}, nil
	}

	// Convert results
	resultList := make([]map[string]any, 0, len(results))
	for _, r := range results {
		item := map[string]any{
			"chunk_id":       r.ChunkID,
			"document_id":    r.DocumentID,
			"content":        r.Content,
			"similarity":     r.Similarity,
			"knowledge_base": r.KnowledgeBaseName,
		}
		if r.DocumentTitle != "" {
			item["document_title"] = r.DocumentTitle
		}
		if len(r.Tags) > 0 {
			item["tags"] = r.Tags
		}
		resultList = append(resultList, item)
	}

	response := map[string]any{
		"query":   query,
		"results": resultList,
		"count":   len(resultList),
	}

	resultJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to serialize result: %v", err))},
			IsError: true,
		}, nil
	}

	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(string(resultJSON))},
	}, nil
}

// parseMetadataFilter parses metadata_filter from MCP arguments into MetadataFilterGroup
func parseMetadataFilter(mf map[string]any) (*ai.MetadataFilterGroup, error) {
	group := &ai.MetadataFilterGroup{}

	// Parse logical_op
	if logicalOp, ok := mf["logical_op"].(string); ok {
		group.LogicalOp = ai.LogicalOperator(logicalOp)
	} else {
		group.LogicalOp = ai.LogicalOpAND // Default to AND
	}

	// Parse conditions
	if conditionsRaw, ok := mf["conditions"].([]any); ok {
		for _, condRaw := range conditionsRaw {
			condMap, ok := condRaw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("condition must be an object")
			}

			cond := ai.MetadataCondition{}

			// Parse key (required)
			key, ok := condMap["key"].(string)
			if !ok {
				return nil, fmt.Errorf("condition key is required")
			}
			cond.Key = key

			// Parse operator (required)
			operatorStr, ok := condMap["operator"].(string)
			if !ok {
				return nil, fmt.Errorf("condition operator is required")
			}
			cond.Operator = ai.MetadataOperator(operatorStr)

			// Parse optional fields based on operator
			if value, ok := condMap["value"]; ok {
				cond.Value = value
			}
			if values, ok := condMap["values"].([]any); ok {
				cond.Values = values
			}
			if min, ok := condMap["min"]; ok {
				cond.Min = min
			}
			if max, ok := condMap["max"]; ok {
				cond.Max = max
			}

			group.Conditions = append(group.Conditions, cond)
		}
	}

	// Parse nested groups recursively
	if groupsRaw, ok := mf["groups"].([]any); ok {
		for _, groupRaw := range groupsRaw {
			nestedMap, ok := groupRaw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("nested group must be an object")
			}

			nestedGroup, err := parseMetadataFilter(nestedMap)
			if err != nil {
				return nil, fmt.Errorf("failed to parse nested group: %w", err)
			}
			group.Groups = append(group.Groups, *nestedGroup)
		}
	}

	return group, nil
}
