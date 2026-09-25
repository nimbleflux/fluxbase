package api

import (
	"regexp"
	"strings"
)

var validIdentifierRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// jsonbKeyRegex restricts JSONB path key segments to word characters so that
// keys are safe to embed in quoted literals and cannot introduce operators.
var jsonbKeyRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// isValidColumnReference reports whether column is a valid column reference:
// either a plain identifier or a JSONB path of the form
// "col->key->>subkey" / "col->0->>key", where every segment is validated
// individually (identifiers by validIdentifierRegex, keys by jsonbKeyRegex,
// array indices as numbers). Validate identifiers before they are interpolated.
func isValidColumnReference(column string) bool {
	if column == "" {
		return false
	}
	if !strings.Contains(column, "->") {
		return isValidIdentifier(column)
	}

	parts := strings.Split(column, "->")
	// The first segment is the table column identifier.
	if !isValidIdentifier(parts[0]) {
		return false
	}
	for _, part := range parts[1:] {
		// A "->>" operator leaves a leading ">" after splitting on "->".
		key := strings.TrimPrefix(part, ">")
		if strings.HasPrefix(key, ">") {
			// More than one leading ">" is not a valid operator sequence.
			return false
		}
		if !jsonbKeyRegex.MatchString(key) {
			return false
		}
	}
	return true
}
