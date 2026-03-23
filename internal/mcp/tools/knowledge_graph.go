package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/ai"
	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// QueryKnowledgeGraphTool implements the query_knowledge_graph MCP tool
type QueryKnowledgeGraphTool struct {
	knowledgeGraph *ai.KnowledgeGraph
}

// NewQueryKnowledgeGraphTool creates a new query_knowledge_graph tool
func NewQueryKnowledgeGraphTool(knowledgeGraph *ai.KnowledgeGraph) *QueryKnowledgeGraphTool {
	return &QueryKnowledgeGraphTool{
		knowledgeGraph: knowledgeGraph,
	}
}

func (t *QueryKnowledgeGraphTool) Name() string {
	return "query_knowledge_graph"
}

func (t *QueryKnowledgeGraphTool) Description() string {
	return "Query entities in the knowledge graph with optional filtering by entity type and search query"
}

func (t *QueryKnowledgeGraphTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"knowledge_base_id": map[string]any{
				"type":        "string",
				"description": "The knowledge base ID to query",
			},
			"entity_type": map[string]any{
				"type":        "string",
				"description": "Filter by entity type (person, organization, location, concept, product, event, table, url, api_endpoint, datetime, code_reference, error, other)",
			},
			"search_query": map[string]any{
				"type":        "string",
				"description": "Fuzzy search query for entity names",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results (default: 50)",
				"default":     50,
				"maximum":     200,
			},
			"include_relationships": map[string]any{
				"type":        "boolean",
				"description": "Whether to include relationships in the response (default: true)",
				"default":     true,
			},
		},
		"required": []string{"knowledge_base_id"},
	}
}

func (t *QueryKnowledgeGraphTool) RequiredScopes() []string {
	return []string{mcp.ScopeReadVectors}
}

func (t *QueryKnowledgeGraphTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	if t.knowledgeGraph == nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("Knowledge graph is not configured")},
			IsError: true,
		}, nil
	}

	// Parse knowledge_base_id
	knowledgeBaseID, ok := args["knowledge_base_id"].(string)
	if !ok || knowledgeBaseID == "" {
		return nil, fmt.Errorf("knowledge_base_id is required")
	}

	// Parse entity_type (optional)
	var entityType *ai.EntityType
	if et, ok := args["entity_type"].(string); ok && et != "" {
		tpe := ai.EntityType(et)
		entityType = &tpe
	}

	// Parse search_query (optional)
	searchQuery, _ := args["search_query"].(string)

	// Parse limit
	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 200 {
			limit = 200
		}
	}

	// Parse include_relationships
	includeRelationships := true
	if ir, ok := args["include_relationships"].(bool); ok {
		includeRelationships = ir
	}

	var entityTypeStr string
	if entityType != nil {
		entityTypeStr = string(*entityType)
	}

	log.Debug().
		Str("knowledge_base_id", knowledgeBaseID).
		Str("entity_type", entityTypeStr).
		Str("search_query", searchQuery).
		Int("limit", limit).
		Bool("include_relationships", includeRelationships).
		Msg("MCP: Querying knowledge graph")

	// Search entities
	entities, err := t.knowledgeGraph.SearchEntities(ctx, knowledgeBaseID, searchQuery, []ai.EntityType{}, limit)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to query knowledge graph: %v", err))},
			IsError: true,
		}, nil
	}

	// Apply entity type filter if specified
	var filteredEntities []ai.Entity
	if entityType != nil {
		for _, e := range entities {
			if e.EntityType == *entityType {
				filteredEntities = append(filteredEntities, e)
			}
		}
	} else {
		filteredEntities = entities
	}

	// Convert results
	resultList := make([]map[string]any, 0, len(filteredEntities))
	for _, e := range filteredEntities {
		item := map[string]any{
			"id":             e.ID,
			"type":           string(e.EntityType),
			"name":           e.Name,
			"canonical_name": e.CanonicalName,
		}
		if len(e.Aliases) > 0 {
			item["aliases"] = e.Aliases
		}
		if len(e.Metadata) > 0 {
			item["metadata"] = e.Metadata
		}

		// Include relationships if requested
		if includeRelationships {
			relationships, err := t.knowledgeGraph.GetRelationships(ctx, knowledgeBaseID, e.ID)
			if err == nil {
				relList := make([]map[string]any, 0, len(relationships))
				for _, r := range relationships {
					rel := map[string]any{
						"id":               r.ID,
						"type":             string(r.RelationshipType),
						"direction":        string(r.Direction),
						"source_entity_id": r.SourceEntityID,
						"target_entity_id": r.TargetEntityID,
					}
					if r.SourceEntity != nil {
						rel["source_entity"] = map[string]any{
							"id":   r.SourceEntity.ID,
							"type": string(r.SourceEntity.EntityType),
							"name": r.SourceEntity.Name,
						}
					}
					if r.TargetEntity != nil {
						rel["target_entity"] = map[string]any{
							"id":   r.TargetEntity.ID,
							"type": string(r.TargetEntity.EntityType),
							"name": r.TargetEntity.Name,
						}
					}
					if len(r.Metadata) > 0 {
						rel["metadata"] = r.Metadata
					}
					relList = append(relList, rel)
				}
				item["relationships"] = relList
			}
		}

		resultList = append(resultList, item)
	}

	response := map[string]any{
		"knowledge_base_id": knowledgeBaseID,
		"entities":          resultList,
		"count":             len(resultList),
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

// FindRelatedEntitiesTool implements the find_related_entities MCP tool
type FindRelatedEntitiesTool struct {
	knowledgeGraph *ai.KnowledgeGraph
}

// NewFindRelatedEntitiesTool creates a new find_related_entities tool
func NewFindRelatedEntitiesTool(knowledgeGraph *ai.KnowledgeGraph) *FindRelatedEntitiesTool {
	return &FindRelatedEntitiesTool{
		knowledgeGraph: knowledgeGraph,
	}
}

func (t *FindRelatedEntitiesTool) Name() string {
	return "find_related_entities"
}

func (t *FindRelatedEntitiesTool) Description() string {
	return "Find entities related to a given entity using graph traversal"
}

func (t *FindRelatedEntitiesTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"knowledge_base_id": map[string]any{
				"type":        "string",
				"description": "The knowledge base ID",
			},
			"entity_id": map[string]any{
				"type":        "string",
				"description": "The starting entity ID",
			},
			"max_depth": map[string]any{
				"type":        "integer",
				"description": "Maximum traversal depth (1-5, default: 2)",
				"default":     2,
				"minimum":     1,
				"maximum":     5,
			},
			"relationship_types": map[string]any{
				"type":        "array",
				"description": "Optional filter by relationship types (works_at, located_in, founded_by, owns, part_of, related_to, knows, customer_of, supplier_of, invested_in, acquired, merged_with, competitor_of, parent_of, child_of, spouse_of, sibling_of, foreign_key, depends_on, other)",
				"items": map[string]any{
					"type": "string",
				},
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of results (default: 100)",
				"default":     100,
				"maximum":     500,
			},
		},
		"required": []string{"knowledge_base_id", "entity_id"},
	}
}

func (t *FindRelatedEntitiesTool) RequiredScopes() []string {
	return []string{mcp.ScopeReadVectors}
}

func (t *FindRelatedEntitiesTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	if t.knowledgeGraph == nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("Knowledge graph is not configured")},
			IsError: true,
		}, nil
	}

	// Parse knowledge_base_id
	knowledgeBaseID, ok := args["knowledge_base_id"].(string)
	if !ok || knowledgeBaseID == "" {
		return nil, fmt.Errorf("knowledge_base_id is required")
	}

	// Parse entity_id
	entityID, ok := args["entity_id"].(string)
	if !ok || entityID == "" {
		return nil, fmt.Errorf("entity_id is required")
	}

	// Parse max_depth
	maxDepth := 2
	if md, ok := args["max_depth"].(float64); ok {
		maxDepth = int(md)
		if maxDepth < 1 {
			maxDepth = 1
		}
		if maxDepth > 5 {
			maxDepth = 5
		}
	}

	// Parse relationship_types (optional)
	var relationshipTypes []ai.RelationshipType
	if rts, ok := args["relationship_types"].([]any); ok {
		for _, rt := range rts {
			if rtStr, ok := rt.(string); ok {
				relationshipTypes = append(relationshipTypes, ai.RelationshipType(rtStr))
			}
		}
	}

	// Parse limit
	limit := 100
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 500 {
			limit = 500
		}
	}

	log.Debug().
		Str("knowledge_base_id", knowledgeBaseID).
		Str("entity_id", entityID).
		Int("max_depth", maxDepth).
		Int("limit", limit).
		Msg("MCP: Finding related entities")

	// Find related entities
	related, err := t.knowledgeGraph.FindRelatedEntities(ctx, knowledgeBaseID, entityID, maxDepth, relationshipTypes)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to find related entities: %v", err))},
			IsError: true,
		}, nil
	}

	// Convert results
	resultList := make([]map[string]any, 0, len(related))
	for _, r := range related {
		item := map[string]any{
			"entity_id":         r.EntityID,
			"entity_type":       string(r.EntityType),
			"name":              r.Name,
			"canonical_name":    r.CanonicalName,
			"depth":             r.Depth,
			"relationship_type": string(r.RelationshipType),
		}
		if len(r.Path) > 0 {
			item["path"] = r.Path
		}
		resultList = append(resultList, item)
	}

	response := map[string]any{
		"knowledge_base_id":  knowledgeBaseID,
		"starting_entity_id": entityID,
		"related_entities":   resultList,
		"count":              len(resultList),
		"max_depth":          maxDepth,
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

// BrowseKnowledgeGraphTool implements the browse_knowledge_graph MCP tool
type BrowseKnowledgeGraphTool struct {
	knowledgeGraph *ai.KnowledgeGraph
}

// NewBrowseKnowledgeGraphTool creates a new browse_knowledge_graph tool
func NewBrowseKnowledgeGraphTool(knowledgeGraph *ai.KnowledgeGraph) *BrowseKnowledgeGraphTool {
	return &BrowseKnowledgeGraphTool{
		knowledgeGraph: knowledgeGraph,
	}
}

func (t *BrowseKnowledgeGraphTool) Name() string {
	return "browse_knowledge_graph"
}

func (t *BrowseKnowledgeGraphTool) Description() string {
	return "Browse the knowledge graph from a starting entity, showing its neighborhood"
}

func (t *BrowseKnowledgeGraphTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"knowledge_base_id": map[string]any{
				"type":        "string",
				"description": "The knowledge base ID",
			},
			"start_entity": map[string]any{
				"type":        "string",
				"description": "The starting entity ID or canonical name",
			},
			"direction": map[string]any{
				"type":        "string",
				"description": "Direction to traverse: 'outgoing' (relationships from this entity), 'incoming' (relationships to this entity), or 'both' (default: both)",
				"enum":        []string{"outgoing", "incoming", "both"},
				"default":     "both",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of related entities (default: 50)",
				"default":     50,
				"maximum":     200,
			},
		},
		"required": []string{"knowledge_base_id", "start_entity"},
	}
}

func (t *BrowseKnowledgeGraphTool) RequiredScopes() []string {
	return []string{mcp.ScopeReadVectors}
}

func (t *BrowseKnowledgeGraphTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	if t.knowledgeGraph == nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent("Knowledge graph is not configured")},
			IsError: true,
		}, nil
	}

	// Parse knowledge_base_id
	knowledgeBaseID, ok := args["knowledge_base_id"].(string)
	if !ok || knowledgeBaseID == "" {
		return nil, fmt.Errorf("knowledge_base_id is required")
	}

	// Parse start_entity
	startEntity, ok := args["start_entity"].(string)
	if !ok || startEntity == "" {
		return nil, fmt.Errorf("start_entity is required")
	}

	// Parse direction
	direction := "both"
	if d, ok := args["direction"].(string); ok {
		direction = d
	}

	// Parse limit
	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 200 {
			limit = 200
		}
	}

	log.Debug().
		Str("knowledge_base_id", knowledgeBaseID).
		Str("start_entity", startEntity).
		Str("direction", direction).
		Int("limit", limit).
		Msg("MCP: Browsing knowledge graph")

	// First, try to get the starting entity by ID
	startEntityObj, err := t.knowledgeGraph.GetEntity(ctx, startEntity)
	if err != nil {
		// If not found by ID, try searching by name
		entities, searchErr := t.knowledgeGraph.SearchEntities(ctx, knowledgeBaseID, startEntity, []ai.EntityType{}, 1)
		if searchErr != nil || len(entities) == 0 {
			return &mcp.ToolResult{
				Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Starting entity not found: %s", startEntity))},
				IsError: true,
			}, nil
		}
		startEntityObj = &entities[0]
		startEntity = startEntityObj.ID
	}

	// Get relationships for this entity
	relationships, err := t.knowledgeGraph.GetRelationships(ctx, knowledgeBaseID, startEntity)
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.ErrorContent(fmt.Sprintf("Failed to get relationships: %v", err))},
			IsError: true,
		}, nil
	}

	// Build response with entity and its neighborhood
	response := map[string]any{
		"knowledge_base_id": knowledgeBaseID,
		"entity": map[string]any{
			"id":             startEntityObj.ID,
			"type":           string(startEntityObj.EntityType),
			"name":           startEntityObj.Name,
			"canonical_name": startEntityObj.CanonicalName,
		},
		"neighborhood": map[string]any{
			"outgoing": []map[string]any{},
			"incoming": []map[string]any{},
		},
	}

	outgoingList := []map[string]any{}
	incomingList := []map[string]any{}
	count := 0

	for _, rel := range relationships {
		if count >= limit {
			break
		}

		relData := map[string]any{
			"id":        rel.ID,
			"type":      string(rel.RelationshipType),
			"direction": string(rel.Direction),
		}
		if len(rel.Metadata) > 0 {
			relData["metadata"] = rel.Metadata
		}

		// Outgoing relationships (where this entity is the source)
		if rel.SourceEntityID == startEntity && (direction == "outgoing" || direction == "both") {
			if rel.TargetEntity != nil {
				relData["target_entity"] = map[string]any{
					"id":   rel.TargetEntity.ID,
					"type": string(rel.TargetEntity.EntityType),
					"name": rel.TargetEntity.Name,
				}
			}
			outgoingList = append(outgoingList, relData)
			count++
		}

		// Incoming relationships (where this entity is the target)
		if rel.TargetEntityID == startEntity && (direction == "incoming" || direction == "both") {
			if rel.SourceEntity != nil {
				relData["source_entity"] = map[string]any{
					"id":   rel.SourceEntity.ID,
					"type": string(rel.SourceEntity.EntityType),
					"name": rel.SourceEntity.Name,
				}
			}
			incomingList = append(incomingList, relData)
			count++
		}
	}

	neighborhood := response["neighborhood"].(map[string]any)
	neighborhood["outgoing"] = outgoingList
	neighborhood["incoming"] = incomingList
	neighborhood["total_count"] = len(relationships)

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
