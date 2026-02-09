package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Manager Construction Tests
// =============================================================================

func TestNewManager(t *testing.T) {
	t.Run("creates manager with nil dependencies", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "local",
		}

		manager := NewManager(cfg, nil, "jwt-secret", "http://localhost", nil)

		require.NotNil(t, manager)
		assert.Equal(t, cfg, manager.Config)
		assert.NotNil(t, manager.Storage)
		assert.Equal(t, "jwt-secret", manager.jwtSecret)
		assert.Equal(t, "http://localhost", manager.publicURL)
		assert.Empty(t, manager.Workers)
		assert.NotNil(t, manager.stopCh)
	})

	t.Run("initializes empty workers slice", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, manager.Workers)
		assert.Len(t, manager.Workers, 0)
	})
}

// =============================================================================
// Manager Start Tests
// =============================================================================

func TestManager_Start_Validation(t *testing.T) {
	t.Run("rejects zero worker count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "local",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		err := manager.Start(context.Background(), 0)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "worker count must be positive")
	})

	t.Run("rejects negative worker count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "local",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		err := manager.Start(context.Background(), -1)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "worker count must be positive")
	})
}

// =============================================================================
// Manager GetWorkerCount Tests
// =============================================================================

func TestManager_GetWorkerCount(t *testing.T) {
	t.Run("returns zero before start", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 0, manager.GetWorkerCount())
	})
}

// =============================================================================
// Manager SetSettingsSecretsService Tests
// =============================================================================

func TestManager_SetSettingsSecretsService(t *testing.T) {
	t.Run("sets nil service", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		manager.SetSettingsSecretsService(nil)

		assert.Nil(t, manager.SettingsSecretsService)
	})
}

// =============================================================================
// Manager CancelJob Tests
// =============================================================================

func TestManager_CancelJob(t *testing.T) {
	t.Run("cancels job with no workers", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic with no workers
		manager.CancelJob(uuid.New())
	})
}

// =============================================================================
// JobsConfig Tests
// =============================================================================

func TestJobsConfig(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		cfg := config.JobsConfig{
			WorkerMode:              "local",
			MaxConcurrentPerWorker:  5,
			DefaultMaxDuration:      30 * time.Minute,
			GracefulShutdownTimeout: 5 * time.Minute,
		}

		assert.Equal(t, "local", cfg.WorkerMode)
		assert.Equal(t, 5, cfg.MaxConcurrentPerWorker)
		assert.Equal(t, 30*time.Minute, cfg.DefaultMaxDuration)
		assert.Equal(t, 5*time.Minute, cfg.GracefulShutdownTimeout)
	})

	t.Run("deno mode configuration", func(t *testing.T) {
		cfg := config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 10,
			DefaultMaxDuration:     time.Hour,
		}

		assert.Equal(t, "deno", cfg.WorkerMode)
		assert.Equal(t, 10, cfg.MaxConcurrentPerWorker)
	})
}

// =============================================================================
// Manager Stop Tests
// =============================================================================

func TestManager_Stop_Simple(t *testing.T) {
	t.Run("stops manager without workers", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "local",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic when stopping with no workers
		manager.Stop()
	})

	t.Run("stops manager multiple times safely", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "local",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Multiple stops should be safe
		manager.Stop()
		manager.Stop()
		manager.Stop()
	})
}

// =============================================================================
// Manager CancelJob Tests Extended
// =============================================================================

func TestManager_CancelJob_Extended(t *testing.T) {
	t.Run("cancels job with zero UUID", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Should handle zero UUID gracefully
		manager.CancelJob(uuid.UUID{})
	})

	t.Run("cancels multiple different jobs", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic when cancelling multiple jobs
		manager.CancelJob(uuid.New())
		manager.CancelJob(uuid.New())
		manager.CancelJob(uuid.New())
	})
}

// =============================================================================
// Manager Lifecycle Tests
// =============================================================================

func TestManager_Lifecycle(t *testing.T) {
	t.Run("create and destroy manager", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "local",
			MaxConcurrentPerWorker: 2,
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, manager)
		assert.Equal(t, 0, manager.GetWorkerCount())
		manager.Stop()
	})

	t.Run("manager with custom config", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:              "deno",
			MaxConcurrentPerWorker:  10,
			DefaultMaxDuration:      1 * time.Hour,
			DefaultProgressTimeout:  10 * time.Minute,
			GracefulShutdownTimeout: 10 * time.Minute,
			PollInterval:            1 * time.Second,
			WorkerHeartbeatInterval: 30 * time.Second,
			WorkerTimeout:           2 * time.Minute,
		}
		manager := NewManager(cfg, nil, "my-secret", "https://example.com", nil)

		assert.Equal(t, cfg, manager.Config)
		assert.Equal(t, "my-secret", manager.jwtSecret)
		assert.Equal(t, "https://example.com", manager.publicURL)
		assert.NotNil(t, manager.stopCh)
	})
}

// =============================================================================
// Manager SetSettingsSecretsService Tests Extended
// =============================================================================

func TestManager_SetSettingsSecretsService_Extended(t *testing.T) {
	t.Run("sets service and can replace it", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Initially nil
		assert.Nil(t, manager.SettingsSecretsService)

		// Set to mock (nil is fine for this test)
		manager.SetSettingsSecretsService(nil)
		assert.Nil(t, manager.SettingsSecretsService)

		// Replace with another nil
		manager.SetSettingsSecretsService(nil)
		assert.Nil(t, manager.SettingsSecretsService)
	})
}
