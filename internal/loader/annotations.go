package loader

import (
	"regexp"
	"strings"
)

var (
	linePatterns = map[string]*regexp.Regexp{
		"//": regexp.MustCompile(`^//\s*@fluxbase:(\S+)(?:\s+(.*?))?(?:\s*\*/)?$`),
		"--": regexp.MustCompile(`^--\s*@fluxbase:(\S+)(?:\s+(.*?))?(?:\s*\*/)?$`),
		"/*": regexp.MustCompile(`^/\*\s*@fluxbase:(\S+)(?:\s+(.*?))?(?:\s*\*/)?$`),
	}
	blockPattern = regexp.MustCompile(`^\*\s*@fluxbase:(\S+)(?:\s+(.*?))?(?:\s*\*/)?$`)
)

// ParseAnnotations extracts @fluxbase:key and @fluxbase:key value pairs from code comments.
// commentStyles specifies which line-comment prefixes to recognize (e.g. "//", "--").
// Block comment bodies (lines starting with *) are always recognized.
// Returns a map of lowercased keys to raw string values (empty string for flag-only annotations).
func ParseAnnotations(code string, commentStyles []string) map[string]string {
	result := make(map[string]string)

	compiled := make([]*regexp.Regexp, 0, len(commentStyles))
	for _, style := range commentStyles {
		if p, ok := linePatterns[style]; ok {
			compiled = append(compiled, p)
		}
	}

	lines := strings.Split(code, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		var matches []string
		for _, p := range compiled {
			if matches = p.FindStringSubmatch(trimmed); matches != nil {
				break
			}
		}
		if matches == nil {
			matches = blockPattern.FindStringSubmatch(trimmed)
		}
		if len(matches) < 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(matches[1]))
		value := ""
		if len(matches) > 2 {
			value = strings.TrimSpace(matches[2])
		}
		if _, exists := result[key]; !exists {
			result[key] = value
		}
	}

	return result
}

// ParseCommaList splits a comma-separated string into trimmed, non-empty strings.
func ParseCommaList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// ParseRoleList parses a comma-separated role list, lowercasing each role.
func ParseRoleList(s string) []string {
	parts := ParseCommaList(s)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return parts
}
