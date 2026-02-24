package ai

import (
	"strings"
)

// QueryClassification represents the type of query and recommended approach
type QueryClassification string

const (
	// QueryTypeStructured indicates the query should use SQL (query_table, execute_sql)
	// Use for: dates, counts, filters, ordering, specific record lookups
	QueryTypeStructured QueryClassification = "structured"

	// QueryTypeSemantic indicates the query should use knowledge base search (search_vectors)
	// Use for: conceptual questions, descriptions, explanations, "what is" questions
	QueryTypeSemantic QueryClassification = "semantic"

	// QueryTypeHybrid indicates the query should use both SQL and KB search
	// Use for: queries needing both data and context
	QueryTypeHybrid QueryClassification = "hybrid"

	// QueryTypeUnknown indicates the query type couldn't be determined
	QueryTypeUnknown QueryClassification = "unknown"
)

// QueryClassifier analyzes user queries to determine the best retrieval method
type QueryClassifier struct{}

// NewQueryClassifier creates a new query classifier
func NewQueryClassifier() *QueryClassifier {
	return &QueryClassifier{}
}

// ClassifyQuery analyzes a user query to determine the best retrieval method
func (c *QueryClassifier) ClassifyQuery(query string) QueryClassification {
	lowerQuery := strings.ToLower(query)

	// Check for semantic indicators (conceptual questions)
	semanticScore := c.scoreSemanticIndicators(lowerQuery)

	// Check for structured indicators (specific data lookups)
	structuredScore := c.scoreStructuredIndicators(lowerQuery)

	// Determine the classification based on scores
	switch {
	case semanticScore > 0 && structuredScore > 0:
		return QueryTypeHybrid
	case semanticScore > structuredScore:
		return QueryTypeSemantic
	case structuredScore > 0:
		return QueryTypeStructured
	default:
		return QueryTypeUnknown
	}
}

// scoreSemanticIndicators returns a score based on semantic query patterns
func (c *QueryClassifier) scoreSemanticIndicators(query string) int {
	score := 0

	// Conceptual question patterns
	semanticPatterns := []string{
		"what is", "what are", "tell me about", "explain",
		"describe", "how does", "why", "information about",
		"details about", "overview", "summary of", "background on",
		"help me understand", "what do you know about",
		"can you tell me about", "i want to know about",
		"learn about", "teach me about",
	}

	for _, pattern := range semanticPatterns {
		if strings.Contains(query, pattern) {
			score += 2
		}
	}

	// Topic keywords that suggest conceptual search
	topicKeywords := []string{
		"cuisine", "culture", "history", "concept", "theory",
		"best practices", "guidelines", "principles", "overview",
		"introduction", "tutorial", "guide", "documentation",
	}

	for _, keyword := range topicKeywords {
		if strings.Contains(query, keyword) {
			score++
		}
	}

	return score
}

// scoreStructuredIndicators returns a score based on structured query patterns
func (c *QueryClassifier) scoreStructuredIndicators(query string) int {
	score := 0

	// Time-based patterns (strong indicator for SQL)
	timePatterns := []string{
		"last week", "last month", "yesterday", "today", "this week",
		"this month", "this year", "last year", "in january", "in february",
		"in march", "in april", "in may", "in june", "in july",
		"in august", "in september", "in october", "in november", "in december",
		"ago", "before", "after", "between", "since", "until",
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
		"2023", "2024", "2025", "2026",
	}

	for _, pattern := range timePatterns {
		if strings.Contains(query, pattern) {
			score += 2
		}
	}

	// Action verbs that suggest data lookup
	actionPatterns := []string{
		"visited", "went to", "ate at", "bought", "purchased",
		"show me", "list", "find", "search for", "lookup",
		"get", "retrieve", "display", "what restaurants",
		"how many", "count", "total", "sum", "average", "max", "min",
	}

	for _, pattern := range actionPatterns {
		if strings.Contains(query, pattern) {
			score += 2
		}
	}

	// Ordering/sorting indicators
	orderPatterns := []string{
		"sort by", "order by", "most recent", "oldest", "newest",
		"first", "last", "top", "bottom", "highest", "lowest",
	}

	for _, pattern := range orderPatterns {
		if strings.Contains(query, pattern) {
			score++
		}
	}

	// Specific value patterns
	specificPatterns := []string{
		"where", "when", "which", "whose",
		"with id", "by id", "named", "called",
	}

	for _, pattern := range specificPatterns {
		if strings.Contains(query, pattern) {
			score++
		}
	}

	return score
}

// GetToolRecommendation returns the recommended tool(s) for a query type
func (c *QueryClassifier) GetToolRecommendation(classification QueryClassification) []string {
	switch classification {
	case QueryTypeStructured:
		return []string{"query_table", "execute_sql"}
	case QueryTypeSemantic:
		return []string{"search_vectors"}
	case QueryTypeHybrid:
		return []string{"search_vectors", "query_table"}
	default:
		return []string{"query_table", "search_vectors"}
	}
}

// GetStrategyDescription returns a human-readable description of the strategy
func (c *QueryClassifier) GetStrategyDescription(classification QueryClassification) string {
	switch classification {
	case QueryTypeStructured:
		return "Use SQL queries (query_table or execute_sql) for specific data lookups with filters"
	case QueryTypeSemantic:
		return "Use knowledge base search (search_vectors) for conceptual information"
	case QueryTypeHybrid:
		return "Use both knowledge base search for context AND SQL queries for specific data"
	default:
		return "Consider both SQL queries and knowledge base search based on the question"
	}
}
