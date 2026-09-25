package schema

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests for the shared_modules.valid_module_path CHECK constraint declared in
// schemas/functions.sql. The DB constraint is the single source of truth for
// shared-module path validation (the API surface returns the raw SQLSTATE 23514
// on violation), so these tests pin the file's regex against the module shapes
// that must be accepted or rejected.
//
// Regression context: the character class [a-zA-Z0-9_/-] rejected dots, so
// legitimate dotted module filenames (e.g. _shared/valhalla.service.ts) failed
// every create. The class now includes '.'.

// modulePathCheck extracts the CHECK regex applied to shared_modules.module_path.
func modulePathCheck(t *testing.T) (*regexp.Regexp, string) {
	t.Helper()
	content, err := os.ReadFile("schemas/functions.sql")
	require.NoError(t, err, "functions.sql must ship with the embedded schemas")

	sql := string(content)
	lines := strings.Split(sql, "\n")
	var constraintLine string
	for _, line := range lines {
		if strings.Contains(line, "CONSTRAINT valid_module_path CHECK") {
			constraintLine = line
			break
		}
	}
	require.NotEmpty(t, constraintLine, "functions.sql must declare the valid_module_path constraint")

	// Extract the POSIX regex from: module_path ~ '<regex>'::text
	re := regexp.MustCompile(`module_path ~ '([^']+)'`)
	m := re.FindStringSubmatch(constraintLine)
	require.Len(t, m, 2, "valid_module_path must contain a module_path ~ '<regex>' clause")

	// The SQL regex is Go-compatible for the constructs used here.
	compiled, err := regexp.Compile(m[1])
	require.NoError(t, err, "module_path regex must be a valid Go regex: %s", m[1])
	return compiled, constraintLine
}

func TestValidModulePath_AcceptsLegitimateModulePaths(t *testing.T) {
	t.Parallel()
	re, _ := modulePathCheck(t)

	// Dotted filenames are the regression case: shared modules follow the
	// <name>.service.ts / <name>.types.ts convention.
	for _, path := range []string{
		"_shared/valhalla.service.ts", // dotted filename — was rejected before the fix
		"_shared/trip-generation.types.ts",
		"_shared/rate-limiter.ts", // pre-fix shapes must keep working
		"_shared/points-core.js",
		"_shared/timezones.mts",
		"_shared/trip-route-geometry.mjs",
		"_shared/utils/db.ts", // nested module
		"_shared/a.b.c.ts",    // multiple dots
	} {
		assert.Truef(t, re.MatchString(path), "module path %q must pass valid_module_path", path)
	}
}

func TestValidModulePath_RejectsInvalidModulePaths(t *testing.T) {
	t.Parallel()
	re, constraintLine := modulePathCheck(t)

	for _, path := range []string{
		"shared/outside-prefix.ts", // must live under _shared/
		"_shared/",                 // no filename
		"_shared/no-extension",
		"_shared/file.exe", // disallowed extension
		"_shared/file.txt", // disallowed extension
		"_sh@red/x.ts",     // invalid character
		"_shared/Über.ts",  // non-ASCII
		"_shared/a b.ts",   // space
	} {
		assert.Falsef(t, re.MatchString(path), "module path %q must fail the valid_module_path regex", path)
	}

	// Traversal is blocked by the second CHECK clause (module_path !~~ '%/../%'),
	// independent of the regex's character class.
	assert.Contains(t, constraintLine, `!~~ '%/../%'::text`,
		"valid_module_path must keep the ../ traversal guard")
}
