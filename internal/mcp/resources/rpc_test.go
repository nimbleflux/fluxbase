package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/mcp"
)

// =============================================================================
// RPCResource Metadata Tests
// =============================================================================

func TestRPCResource_URI(t *testing.T) {
	r := &RPCResource{}
	assert.Equal(t, "fluxbase://rpc", r.URI())
}

func TestRPCResource_Name(t *testing.T) {
	r := &RPCResource{}
	assert.Equal(t, "RPC Procedures", r.Name())
}

func TestRPCResource_Description(t *testing.T) {
	r := &RPCResource{}
	desc := r.Description()
	assert.Contains(t, desc, "RPC procedures")
	assert.Contains(t, desc, "schemas")
}

func TestRPCResource_MimeType(t *testing.T) {
	r := &RPCResource{}
	assert.Equal(t, "application/json", r.MimeType())
}

func TestRPCResource_RequiredScopes(t *testing.T) {
	r := &RPCResource{}
	scopes := r.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeExecuteRPC, scopes[0])
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewRPCResource(t *testing.T) {
	t.Run("creates with nil storage", func(t *testing.T) {
		r := NewRPCResource(nil)
		require.NotNil(t, r)
		assert.Nil(t, r.storage)
	})
}

// =============================================================================
// RPCResource Struct Tests
// =============================================================================

func TestRPCResource_Struct(t *testing.T) {
	t.Run("has storage field", func(t *testing.T) {
		r := &RPCResource{}
		assert.Nil(t, r.storage)
	})
}
