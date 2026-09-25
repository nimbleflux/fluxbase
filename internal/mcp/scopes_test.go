package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterSupportedScopes(t *testing.T) {
	t.Run("accepts supported scopes", func(t *testing.T) {
		supported, unknown := FilterSupportedScopes([]string{"tables:read", "storage:write"})
		assert.ElementsMatch(t, []string{"tables:read", "storage:write"}, supported)
		assert.Empty(t, unknown)
	})

	t.Run("flags unknown scopes", func(t *testing.T) {
		supported, unknown := FilterSupportedScopes([]string{"tables:read", "make:coffee"})
		assert.Equal(t, []string{"tables:read"}, supported)
		assert.Equal(t, []string{"make:coffee"}, unknown)
	})

	t.Run("skips empty entries", func(t *testing.T) {
		supported, unknown := FilterSupportedScopes([]string{"", "tables:read"})
		assert.Equal(t, []string{"tables:read"}, supported)
		assert.Empty(t, unknown)
	})
}

func TestIsPrivilegedScope(t *testing.T) {
	assert.True(t, IsPrivilegedScope("admin:ddl"))
	assert.True(t, IsPrivilegedScope("admin:schemas"))
	assert.True(t, IsPrivilegedScope("sync:migrations"))
	assert.True(t, IsPrivilegedScope("sync:functions"))
	assert.True(t, IsPrivilegedScope("branch:write"))
	assert.True(t, IsPrivilegedScope("github:read"))

	assert.False(t, IsPrivilegedScope("tables:read"))
	assert.False(t, IsPrivilegedScope("execute:sql"))
	assert.False(t, IsPrivilegedScope("read:schema"))
	assert.False(t, IsPrivilegedScope(""))

	assert.True(t, HasPrivilegedScope([]string{"tables:read", "branch:write"}))
	assert.False(t, HasPrivilegedScope([]string{"tables:read", "read:schema"}))
}

func TestSupportedScopesAdvertiseFullSet(t *testing.T) {
	// The supported list must include the core scopes advertised to clients
	// and the custom tool/resource scopes.
	for _, scope := range []string{
		"tables:read", "tables:write", "read:schema", "admin:ddl",
		"execute:custom", "read:custom", "sync:migrations",
	} {
		assert.True(t, IsSupportedScope(scope), "scope %s should be supported", scope)
	}
}
