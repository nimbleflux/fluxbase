package bootstrap

import "strings"

// AppUserPlaceholder is the placeholder string used in SQL files for the
// runtime database user. It is replaced at execution time with the actual
// configured user (cfg.Database.User).
const AppUserPlaceholder = "{{APP_USER}}"

// SubstituteAppUser replaces all {{APP_USER}} placeholders in the SQL string
// with the provided appUser role name.
func SubstituteAppUser(sql string, appUser string) string {
	if appUser == "" || !strings.Contains(sql, AppUserPlaceholder) {
		return sql
	}
	return strings.ReplaceAll(sql, AppUserPlaceholder, appUser)
}
