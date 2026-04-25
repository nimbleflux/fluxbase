package bootstrap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubstituteAppUser(t *testing.T) {
	t.Run("valid identifier postgres", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "postgres")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO postgres", result)
	})

	t.Run("valid identifier my_user", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my_user")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO my_user", result)
	})

	t.Run("valid identifier _underscore", func(t *testing.T) {
		result, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "_underscore")
		require.NoError(t, err)
		assert.Equal(t, "GRANT ALL TO _underscore", result)
	})

	t.Run("empty appUser returns sql unchanged", func(t *testing.T) {
		sql := "GRANT ALL TO {{APP_USER}}"
		result, err := SubstituteAppUser(sql, "")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("no placeholder returns sql unchanged", func(t *testing.T) {
		sql := "SELECT 1"
		result, err := SubstituteAppUser(sql, "postgres")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})

	t.Run("rejected hyphen in identifier", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my-user")
		assert.Error(t, err)
	})

	t.Run("rejected semicolon injection", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "user; DROP TABLE")
		assert.Error(t, err)
	})

	t.Run("rejected quote in identifier", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "user'name")
		assert.Error(t, err)
	})

	t.Run("rejected starts with digit", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "1user")
		assert.Error(t, err)
	})

	t.Run("rejected contains space", func(t *testing.T) {
		_, err := SubstituteAppUser("GRANT ALL TO {{APP_USER}}", "my user")
		assert.Error(t, err)
	})

	t.Run("empty string with placeholder returns unchanged", func(t *testing.T) {
		sql := "GRANT ALL TO {{APP_USER}}"
		result, err := SubstituteAppUser(sql, "")
		require.NoError(t, err)
		assert.Equal(t, sql, result)
	})
}
