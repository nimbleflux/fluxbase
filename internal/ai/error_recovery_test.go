package ai

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorAnalyzer_AnalyzeError(t *testing.T) {
	analyzer := NewErrorAnalyzer()

	tests := []struct {
		name           string
		err            error
		expectedType   ErrorType
		expectRecovery bool
	}{
		{
			name:           "syntax error",
			err:            errors.New("syntax error at or near SELECT"),
			expectedType:   ErrorTypeSQL,
			expectRecovery: true,
		},
		{
			name:           "table not found",
			err:            errors.New("relation \"unknown_table\" does not exist"),
			expectedType:   ErrorTypeTableNotFound,
			expectRecovery: true,
		},
		{
			name:           "column not found",
			err:            errors.New("column \"unknown_col\" does not exist"),
			expectedType:   ErrorTypeColumnNotFound,
			expectRecovery: true,
		},
		{
			name:           "permission denied",
			err:            errors.New("permission denied for table users"),
			expectedType:   ErrorTypePermissionDenied,
			expectRecovery: false, // Permission errors are not recoverable
		},
		{
			name:           "timeout",
			err:            errors.New("query timed out after 30s"),
			expectedType:   ErrorTypeTimeout,
			expectRecovery: true,
		},
		{
			name:           "no results",
			err:            errors.New("no results found"),
			expectedType:   ErrorTypeNoResults,
			expectRecovery: true,
		},
		{
			name:           "unknown error",
			err:            errors.New("something weird happened"),
			expectedType:   ErrorTypeUnknown,
			expectRecovery: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := analyzer.AnalyzeError(tt.err, "test query", "test_tool")
			assert.NotNil(t, suggestion)
			assert.Equal(t, tt.expectedType, suggestion.ErrorType)
			assert.NotEmpty(t, suggestion.Suggestions)

			// Check recovery expectation
			isRecoverable := IsRecoverableError(tt.err)
			assert.Equal(t, tt.expectRecovery, isRecoverable)
		})
	}
}

func TestErrorAnalyzer_NilError(t *testing.T) {
	analyzer := NewErrorAnalyzer()
	suggestion := analyzer.AnalyzeError(nil, "test query", "test_tool")
	assert.Nil(t, suggestion)
}

func TestFormatErrorForLLM(t *testing.T) {
	err := errors.New("syntax error at or near SELECT")

	formatted := FormatErrorForLLM(err, "SELECT * FORM users", "execute_sql")

	assert.Contains(t, formatted, "ERROR")
	assert.Contains(t, formatted, "syntax error")
	assert.Contains(t, formatted, "RECOVERY SUGGESTIONS")
}

func TestShouldSuggestAlternative(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "no results suggests alternative",
			err:      errors.New("no results found"),
			expected: true,
		},
		{
			name:     "syntax error suggests alternative",
			err:      errors.New("syntax error"),
			expected: true,
		},
		{
			name:     "permission denied does not suggest alternative",
			err:      errors.New("permission denied"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldSuggestAlternative(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecoverySuggestion_JSON(t *testing.T) {
	suggestion := &RecoverySuggestion{
		ErrorType:       ErrorTypeSQL,
		Message:         "SQL syntax error",
		Suggestions:     []string{"Check syntax", "Verify table names"},
		AlternativeTool: "query_table",
	}

	// Verify fields are accessible
	assert.Equal(t, ErrorTypeSQL, suggestion.ErrorType)
	assert.Equal(t, "SQL syntax error", suggestion.Message)
	assert.Len(t, suggestion.Suggestions, 2)
	assert.Equal(t, "query_table", suggestion.AlternativeTool)
}
