package bootstrap

import (
	"fmt"
	"regexp"
	"strings"
)

var validPgIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// AppUserPlaceholder is the placeholder string used in SQL files for the
// runtime database user. It is replaced at execution time with the actual
// configured user (cfg.Database.User).
const AppUserPlaceholder = "{{APP_USER}}"

// SubstituteAppUser replaces all {{APP_USER}} placeholders in the SQL string
// with the provided appUser role name. Returns an error if the username is
// not a valid PostgreSQL identifier to prevent SQL injection via bare-string
// substitution into GRANT and ALTER DEFAULT PRIVILEGES statements.
func SubstituteAppUser(sql string, appUser string) (string, error) {
	if appUser == "" || !strings.Contains(sql, AppUserPlaceholder) {
		return sql, nil
	}
	if !validPgIdentifier.MatchString(appUser) {
		return "", fmt.Errorf("invalid database username %q: must match ^[a-zA-Z_][a-zA-Z0-9_]*$", appUser)
	}
	return strings.ReplaceAll(sql, AppUserPlaceholder, appUser), nil
}
