package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveRequestedScopes(t *testing.T) {
	t.Run("defaults to client scopes", func(t *testing.T) {
		client := &mcpOAuthClient{Scopes: []string{"tables:read", "read:schema"}}
		scopes, err := resolveRequestedScopes(client, "")
		require.NoError(t, err)
		assert.ElementsMatch(t, []string{"tables:read", "read:schema"}, scopes)
	})

	t.Run("rejects scopes outside the client registration", func(t *testing.T) {
		client := &mcpOAuthClient{Scopes: []string{"tables:read"}}
		_, err := resolveRequestedScopes(client, "tables:read storage:write")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed for this client")
	})

	t.Run("rejects scopes unknown to the platform", func(t *testing.T) {
		client := &mcpOAuthClient{Scopes: []string{"tables:read", "make:coffee"}}
		_, err := resolveRequestedScopes(client, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scope")
	})

	t.Run("rejects stale unsupported scopes from old registrations", func(t *testing.T) {
		client := &mcpOAuthClient{Scopes: []string{"tables:read", "legacy:scope"}}
		_, err := resolveRequestedScopes(client, "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported scope")
	})
}

func TestIsPrivilegedOAuthRole(t *testing.T) {
	assert.True(t, isPrivilegedOAuthRole("admin"))
	assert.True(t, isPrivilegedOAuthRole("instance_admin"))
	assert.False(t, isPrivilegedOAuthRole("authenticated"))
	assert.False(t, isPrivilegedOAuthRole("tenant_admin"))
	assert.False(t, isPrivilegedOAuthRole(""))
}
