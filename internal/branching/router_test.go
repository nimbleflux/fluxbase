package branching

import (
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Router Construction Tests
// =============================================================================

func TestNewRouter(t *testing.T) {
	t.Run("creates router with nil dependencies", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		assert.NotNil(t, router)
		assert.NotNil(t, router.pools)
		assert.NotNil(t, router.poolConfigs)
		assert.Empty(t, router.pools)
	})

	t.Run("initializes with empty active branch (API-set only)", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "development",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// activeBranch should be empty initially (only set via API)
		activeBranch := router.activeBranch.Load()
		assert.Equal(t, "", activeBranch)

		// Config default branch is stored separately
		assert.Equal(t, "development", router.config.DefaultBranch)
	})

	t.Run("initializes with empty default branch", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		activeBranch := router.activeBranch.Load()
		assert.Equal(t, "", activeBranch)
	})
}

// =============================================================================
// Router GetPool Tests
// =============================================================================

func TestRouter_GetPool_MainBranch(t *testing.T) {
	t.Run("empty slug returns main pool", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// With nil main pool, this should return nil
		pool, err := router.GetPool(nil, "")
		assert.NoError(t, err)
		assert.Nil(t, pool) // main pool is nil in this test
	})

	t.Run("main slug returns main pool", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		pool, err := router.GetPool(nil, "main")
		assert.NoError(t, err)
		assert.Nil(t, pool) // main pool is nil in this test
	})
}

func TestRouter_GetPool_BranchingDisabled(t *testing.T) {
	t.Run("returns error when branching disabled", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: false,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		_, err := router.GetPool(nil, "feature-branch")
		assert.Error(t, err)
		assert.Equal(t, ErrBranchingDisabled, err)
	})
}

// =============================================================================
// Router ClosePool Tests
// =============================================================================

func TestRouter_ClosePool(t *testing.T) {
	t.Run("closes non-existent pool gracefully", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// Should not panic
		router.ClosePool("non-existent")
	})
}

// =============================================================================
// Router Active Branch Tests
// =============================================================================

func TestRouter_ActiveBranch(t *testing.T) {
	t.Run("get active branch", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "staging",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// GetActiveBranch returns empty when no branch set via API
		active := router.GetActiveBranch()
		assert.Equal(t, "", active)

		// GetDefaultBranch returns config default when no API-set branch
		defaultBranch := router.GetDefaultBranch()
		assert.Equal(t, "staging", defaultBranch)
	})

	t.Run("set active branch", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		router.SetActiveBranch("new-branch")

		active := router.GetActiveBranch()
		assert.Equal(t, "new-branch", active)
	})

	t.Run("clear active branch", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:       true,
			DefaultBranch: "initial",
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		router.SetActiveBranch("")

		active := router.GetActiveBranch()
		assert.Equal(t, "", active)
	})
}

// =============================================================================
// Router Pool Management Tests
// =============================================================================

func TestRouter_PoolManagement(t *testing.T) {
	t.Run("pools map is thread-safe", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// Should be able to access pools map safely
		router.poolsMu.RLock()
		count := len(router.pools)
		router.poolsMu.RUnlock()

		assert.Equal(t, 0, count)
	})
}

// =============================================================================
// Router CloseAll Tests
// =============================================================================

func TestRouter_CloseAll(t *testing.T) {
	t.Run("closes all pools", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// Should not panic
		router.CloseAllPools()

		// Pools should be empty
		router.poolsMu.RLock()
		count := len(router.pools)
		router.poolsMu.RUnlock()

		assert.Equal(t, 0, count)
	})
}

// =============================================================================
// Router Branch Connection URL Tests
// =============================================================================

func TestRouter_BranchConnectionURL(t *testing.T) {
	t.Run("generates connection URL for branch", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://user:pass@localhost:5432/fluxbase")

		branch := &Branch{
			DatabaseName: "branch_feature-123",
		}

		url, err := router.getBranchConnectionURL(branch)
		assert.NoError(t, err)
		assert.Contains(t, url, "branch_feature-123")
		assert.Contains(t, url, "localhost")
	})

	t.Run("handles simple URL format", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "not-a-valid-url")

		branch := &Branch{
			DatabaseName: "branch_test",
		}

		// url.Parse is lenient and won't error on simple strings
		// It just treats them as relative paths
		url, err := router.getBranchConnectionURL(branch)
		assert.NoError(t, err)
		assert.Contains(t, url, "branch_test")
	})
}

// =============================================================================
// Router Config Tests
// =============================================================================

func TestRouter_Config(t *testing.T) {
	t.Run("stores config reference", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled:            true,
			MaxBranchesPerUser: 10,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		assert.Equal(t, cfg.Enabled, router.config.Enabled)
		assert.Equal(t, cfg.MaxBranchesPerUser, router.config.MaxBranchesPerUser)
	})
}

// =============================================================================
// Router Storage Tests
// =============================================================================

func TestRouter_Storage(t *testing.T) {
	t.Run("stores storage reference", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// Storage is nil when passed as nil
		assert.Nil(t, router.storage)
	})
}

// =============================================================================
// Router Main Pool Tests
// =============================================================================

func TestRouter_MainPool(t *testing.T) {
	t.Run("stores main pool reference", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		router := NewRouter(nil, cfg, nil, "postgresql://localhost/fluxbase")

		// Main pool is nil when passed as nil
		assert.Nil(t, router.mainPool)
	})
}

// =============================================================================
// Router Main DB URL Tests
// =============================================================================

func TestRouter_MainDBURL(t *testing.T) {
	t.Run("stores main database URL", func(t *testing.T) {
		cfg := config.BranchingConfig{
			Enabled: true,
		}

		url := "postgresql://user:pass@localhost:5432/fluxbase"
		router := NewRouter(nil, cfg, nil, url)

		assert.Equal(t, url, router.mainDBURL)
	})
}
