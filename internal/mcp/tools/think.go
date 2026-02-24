package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fluxbase-eu/fluxbase/internal/mcp"
)

// ThinkTool implements a reasoning tool that forces the AI to plan before acting.
// This implements the ReAct (Reasoning + Acting) pattern for better agent behavior.
type ThinkTool struct{}

// NewThinkTool creates a new think tool
func NewThinkTool() *ThinkTool {
	return &ThinkTool{}
}

func (t *ThinkTool) Name() string {
	return "think"
}

func (t *ThinkTool) Description() string {
	return `Plan your investigation approach before executing queries. Use this tool to:
1. Analyze the user's question - what specific information do they need?
2. Break down complex questions into multiple steps
3. Decide which tool(s) to use and in what order
4. Consider what follow-up queries might be needed

IMPORTANT: Be thorough in your planning. Consider:
- Do I need to query multiple tables?
- Should I start with a broad query to understand the data?
- What filters or conditions are needed?
- Will I need follow-up queries based on initial results?

This tool does NOT execute anything - it only helps you reason through your plan.
Always use this tool first when the question requires querying data.`
}

func (t *ThinkTool) InputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"analysis": map[string]any{
				"type":        "string",
				"description": "Your analysis of the user's question - what are they asking for?",
			},
			"plan": map[string]any{
				"type":        "array",
				"items":       map[string]any{"type": "string"},
				"description": "Step-by-step plan to answer the question (e.g., ['1. Query user_visits table for restaurants', '2. Filter by date range last month'])",
			},
			"tool_choice": map[string]any{
				"type":        "string",
				"enum":        []string{"query_table", "execute_sql", "search_vectors", "hybrid"},
				"description": "Which tool approach to use: query_table (simple filters), execute_sql (complex queries), search_vectors (semantic/conceptual), or hybrid (both)",
			},
			"reasoning": map[string]any{
				"type":        "string",
				"description": "Why you chose this approach over alternatives",
			},
		},
		"required": []string{"analysis", "plan", "tool_choice"},
	}
}

func (t *ThinkTool) RequiredScopes() []string {
	// No scopes required - this tool is always available
	return []string{}
}

func (t *ThinkTool) Execute(ctx context.Context, args map[string]any, authCtx *mcp.AuthContext) (*mcp.ToolResult, error) {
	// Extract and validate arguments
	analysis, _ := args["analysis"].(string)
	toolChoice, _ := args["tool_choice"].(string)
	reasoning, _ := args["reasoning"].(string)

	// Extract plan array
	var plan []string
	if planRaw, ok := args["plan"].([]any); ok {
		for _, item := range planRaw {
			if str, ok := item.(string); ok {
				plan = append(plan, str)
			}
		}
	}

	// Build the response - echo back the plan for confirmation
	// The value is in forcing structured thinking, not in the output
	result := map[string]any{
		"status":      "plan_confirmed",
		"analysis":    analysis,
		"plan":        plan,
		"tool_choice": toolChoice,
		"reasoning":   reasoning,
		"message":     "Plan recorded. Proceed with executing your plan using the appropriate tools.",
	}

	// Provide helpful hints based on tool choice
	switch toolChoice {
	case "query_table":
		result["hint"] = "Use query_table with appropriate filters. Remember to include 'select', 'filter', and 'limit' parameters."
	case "execute_sql":
		result["hint"] = "Use execute_sql for complex queries. Remember to include a 'description' of what the query does."
	case "search_vectors":
		result["hint"] = "Use search_vectors for semantic/conceptual searches. Provide a natural language query."
	case "hybrid":
		result["hint"] = "Use both search_vectors (for context) and query_table/execute_sql (for data). Start with whichever is more appropriate for the first step."
	}

	resultJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &mcp.ToolResult{
			Content: []mcp.Content{mcp.TextContent("Plan confirmed. Proceed with your approach.")},
		}, nil
	}

	return &mcp.ToolResult{
		Content: []mcp.Content{mcp.TextContent(fmt.Sprintf("Plan confirmed:\n%s", string(resultJSON)))},
	}, nil
}
