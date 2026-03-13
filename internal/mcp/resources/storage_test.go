package resources

import (
	"testing"

	"github.com/nimbleflux/fluxbase/internal/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BucketsResource Metadata Tests
// =============================================================================

func TestBucketsResource_URI(t *testing.T) {
	r := &BucketsResource{}
	assert.Equal(t, "fluxbase://storage/buckets", r.URI())
}

func TestBucketsResource_Name(t *testing.T) {
	r := &BucketsResource{}
	assert.Equal(t, "Storage Buckets", r.Name())
}

func TestBucketsResource_Description(t *testing.T) {
	r := &BucketsResource{}
	desc := r.Description()
	assert.Contains(t, desc, "storage buckets")
	assert.Contains(t, desc, "configurations")
}

func TestBucketsResource_MimeType(t *testing.T) {
	r := &BucketsResource{}
	assert.Equal(t, "application/json", r.MimeType())
}

func TestBucketsResource_RequiredScopes(t *testing.T) {
	r := &BucketsResource{}
	scopes := r.RequiredScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, mcp.ScopeReadStorage, scopes[0])
}

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewBucketsResource(t *testing.T) {
	t.Run("creates with nil database", func(t *testing.T) {
		r := NewBucketsResource(nil)
		require.NotNil(t, r)
		assert.Nil(t, r.db)
	})
}

// =============================================================================
// BucketsResource Struct Tests
// =============================================================================

func TestBucketsResource_Struct(t *testing.T) {
	t.Run("has db field", func(t *testing.T) {
		r := &BucketsResource{}
		assert.Nil(t, r.db)
	})
}
