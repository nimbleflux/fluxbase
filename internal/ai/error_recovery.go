package ai

import (
	"fmt"
	"strings"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeSQL              ErrorType = "sql_error"
	ErrorTypeTableNotFound    ErrorType = "table_not_found"
	ErrorTypeColumnNotFound   ErrorType = "column_not_found"
	ErrorTypePermissionDenied ErrorType = "permission_denied"
	ErrorTypeNoResults        ErrorType = "no_results"
	ErrorTypeTimeout          ErrorType = "timeout"
	ErrorTypeUnknown          ErrorType = "unknown"
)

// RecoverySuggestion provides guidance when a query fails
type RecoverySuggestion struct {
	ErrorType       ErrorType `json:"error_type"`
	Message         string    `json:"message"`
	Suggestions     []string  `json:"suggestions"`
	AlternativeTool string    `json:"alternative_tool,omitempty"`
	ExampleQuery    string    `json:"example_query,omitempty"`
}

// ErrorAnalyzer analyzes errors and provides recovery suggestions
type ErrorAnalyzer struct{}

// NewErrorAnalyzer creates a new error analyzer
func NewErrorAnalyzer() *ErrorAnalyzer {
	return &ErrorAnalyzer{}
}

// AnalyzeError analyzes an error and returns recovery suggestions
func (a *ErrorAnalyzer) AnalyzeError(err error, query string, toolName string) *RecoverySuggestion {
	if err == nil {
		return nil
	}

	errStr := err.Error()
	lowerErr := strings.ToLower(errStr)

	// Detect error type
	switch {
	case strings.Contains(lowerErr, "syntax error") ||
		strings.Contains(lowerErr, "invalid syntax") ||
		strings.Contains(lowerErr, "parse error"):
		return a.syntaxErrorSuggestion(errStr, query, toolName)

	case strings.Contains(lowerErr, "column") && strings.Contains(lowerErr, "does not exist"):
		return a.columnNotFoundSuggestion(errStr, query, toolName)

	case strings.Contains(lowerErr, "does not exist") ||
		strings.Contains(lowerErr, "not found") ||
		strings.Contains(lowerErr, "unknown table") ||
		strings.Contains(lowerErr, "relation") && strings.Contains(lowerErr, "does not exist"):
		return a.tableNotFoundSuggestion(errStr, query, toolName)

	case strings.Contains(lowerErr, "permission denied") ||
		strings.Contains(lowerErr, "not allowed") ||
		strings.Contains(lowerErr, "access denied") ||
		strings.Contains(lowerErr, "forbidden"):
		return a.permissionDeniedSuggestion(errStr, query, toolName)

	case strings.Contains(lowerErr, "timeout") ||
		strings.Contains(lowerErr, "timed out") ||
		strings.Contains(lowerErr, "context deadline"):
		return a.timeoutSuggestion(errStr, query, toolName)

	case strings.Contains(lowerErr, "no results") ||
		strings.Contains(lowerErr, "0 rows") ||
		strings.Contains(lowerErr, "no rows") ||
		strings.Contains(lowerErr, "empty result"):
		return a.noResultsSuggestion(errStr, query, toolName)

	default:
		return a.unknownErrorSuggestion(errStr, query, toolName)
	}
}

func (a *ErrorAnalyzer) syntaxErrorSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeSQL,
		Message:   "SQL syntax error detected",
		Suggestions: []string{
			"Check SQL syntax, especially quotes, parentheses, and commas",
			"Verify table and column names against the schema",
			"Try using query_table instead of execute_sql for simpler queries",
			"Make sure string values are properly quoted",
		},
		AlternativeTool: "query_table",
	}
}

func (a *ErrorAnalyzer) tableNotFoundSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeTableNotFound,
		Message:   "Table not found",
		Suggestions: []string{
			"Verify the table name against the available tables in the schema",
			"Check if you need to include the schema prefix (e.g., public.my_table)",
			"The table might not exist in the current database",
			"Check for typos in the table name",
		},
	}
}

func (a *ErrorAnalyzer) columnNotFoundSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeColumnNotFound,
		Message:   "Column not found",
		Suggestions: []string{
			"Verify the column name against the table schema",
			"Check for typos in the column name",
			"The column might be in a different table - consider using a JOIN",
			"Column names are case-sensitive in some databases",
		},
	}
}

func (a *ErrorAnalyzer) permissionDeniedSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypePermissionDenied,
		Message:   "Permission denied for this operation",
		Suggestions: []string{
			"This table or operation is not allowed for this chatbot",
			"Check the allowed tables and operations in the system prompt",
			"Try a different table that is allowed",
			"Ask the user for clarification if you need access to restricted data",
		},
	}
}

func (a *ErrorAnalyzer) timeoutSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeTimeout,
		Message:   "Query timed out",
		Suggestions: []string{
			"The query took too long to execute",
			"Try adding more specific filters to reduce the result set",
			"Add a LIMIT clause if not already present",
			"Break the query into smaller parts",
		},
	}
}

func (a *ErrorAnalyzer) noResultsSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeNoResults,
		Message:   "No results found",
		Suggestions: []string{
			"Try broadening your search criteria",
			"Remove some filters to see if any data exists",
			"Use search_vectors for conceptual matches instead of exact filters",
			"Check if the user has any data matching the criteria",
			"Ask the user if they expected specific data to exist",
		},
		AlternativeTool: "search_vectors",
	}
}

func (a *ErrorAnalyzer) unknownErrorSuggestion(errStr, query, toolName string) *RecoverySuggestion {
	return &RecoverySuggestion{
		ErrorType: ErrorTypeUnknown,
		Message:   "An unexpected error occurred",
		Suggestions: []string{
			"Try rephrasing your query",
			"Use the think tool to plan a different approach",
			"Try using a simpler query first to verify data exists",
			"Check the error message for specific details",
		},
	}
}

// FormatErrorForLLM creates an error message with recovery suggestions for the LLM
func FormatErrorForLLM(err error, query string, toolName string) string {
	analyzer := NewErrorAnalyzer()
	suggestion := analyzer.AnalyzeError(err, query, toolName)

	var sb strings.Builder
	fmt.Fprintf(&sb, "ERROR in %s: %s\n\n", toolName, err.Error())

	if suggestion.Message != "" {
		fmt.Fprintf(&sb, "Issue: %s\n\n", suggestion.Message)
	}

	if len(suggestion.Suggestions) > 0 {
		sb.WriteString("RECOVERY SUGGESTIONS:\n")
		for i, s := range suggestion.Suggestions {
			fmt.Fprintf(&sb, "%d. %s\n", i+1, s)
		}
	}

	if suggestion.AlternativeTool != "" {
		fmt.Fprintf(&sb, "\nALTERNATIVE: Consider using the '%s' tool instead.\n", suggestion.AlternativeTool)
	}

	return sb.String()
}

// IsRecoverableError returns true if the error might be recoverable with a different approach
func IsRecoverableError(err error) bool {
	if err == nil {
		return false
	}

	analyzer := NewErrorAnalyzer()
	suggestion := analyzer.AnalyzeError(err, "", "")

	// Most errors are recoverable except permission denied
	return suggestion.ErrorType != ErrorTypePermissionDenied
}

// ShouldSuggestAlternative returns true if an alternative tool might work better
func ShouldSuggestAlternative(err error) bool {
	if err == nil {
		return false
	}

	analyzer := NewErrorAnalyzer()
	suggestion := analyzer.AnalyzeError(err, "", "")

	return suggestion.AlternativeTool != ""
}
