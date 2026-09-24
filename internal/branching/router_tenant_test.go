package branching

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// TestPoolKey verifies branch pools are keyed by tenant ID + slug: branch slugs
// are only unique per tenant (UNIQUE(slug, tenant_id)), so two tenants using
// the same slug must never share a pool.
func TestPoolKey(t *testing.T) {
	assert.Equal(t, "\x00dev", poolKey("", "dev"))
	assert.Equal(t, "tenant-a\x00dev", poolKey("tenant-a", "dev"))
	assert.Equal(t, "tenant-b\x00dev", poolKey("tenant-b", "dev"))

	assert.NotEqual(t, poolKey("tenant-a", "dev"), poolKey("tenant-b", "dev"),
		"same slug in different tenants must map to different pool keys")
	assert.NotEqual(t, poolKey("", "dev"), poolKey("tenant-a", "dev"),
		"instance-level and tenant-scoped branches with the same slug must not share a pool key")
}

func newTestRouter(enabled bool) *Router {
	cfg := config.BranchingConfig{Enabled: enabled}
	return NewRouter(nil, cfg, nil, "postgres://localhost/main")
}

// TestGetPoolForBranch_MainSlug verifies main-branch requests bypass pool
// creation entirely and return the main pool.
func TestGetPoolForBranch_MainSlug(t *testing.T) {
	r := newTestRouter(false)

	pool, err := r.GetPoolForBranch(context.Background(), "tenant-a", "main")
	require.NoError(t, err)
	assert.Nil(t, pool) // main pool (nil in this test) with no error

	pool, err = r.GetPoolForBranch(context.Background(), "tenant-a", "")
	require.NoError(t, err)
	assert.Nil(t, pool)
}

// TestGetPoolForBranch_Disabled verifies branching-disabled rejection applies
// to non-main branches in a tenant scope.
func TestGetPoolForBranch_Disabled(t *testing.T) {
	r := newTestRouter(false)

	_, err := r.GetPoolForBranch(context.Background(), "tenant-a", "feature")
	assert.ErrorIs(t, err, ErrBranchingDisabled)
}
