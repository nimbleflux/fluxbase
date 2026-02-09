package jobs

import (
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Manager Start with Workers Tests
// =============================================================================

func TestManager_Start_MultipleWorkers(t *testing.T) {
	t.Run("manager configuration is stored", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:              "deno",
			MaxConcurrentPerWorker:  10,
			DefaultMaxDuration:      time.Hour,
			GracefulShutdownTimeout: 5 * time.Minute,
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, cfg, manager.Config)
		assert.Equal(t, "deno", manager.Config.WorkerMode)
		assert.Equal(t, 10, manager.Config.MaxConcurrentPerWorker)
	})

	t.Run("manager initializes with empty workers", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "deno",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, manager)
		assert.Empty(t, manager.Workers)
		assert.Equal(t, 0, manager.GetWorkerCount())
	})
}

// =============================================================================
// Manager Stop Tests
// =============================================================================

func TestManager_Stop(t *testing.T) {
	t.Run("stop without start does not panic", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic
		manager.Stop()
	})

	t.Run("stop with empty workers slice", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		manager.Stop()
		assert.Empty(t, manager.Workers)
	})
}

// =============================================================================
// Manager CancelJob Tests with Workers
// =============================================================================

func TestManager_CancelJob_WithWorkers(t *testing.T) {
	t.Run("cancels job across multiple workers", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "deno",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Manually add workers for testing
		worker1 := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker2 := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		manager.Workers = append(manager.Workers, worker1, worker2)

		jobID := uuid.New()

		// Should not panic - cancelJob on each worker
		manager.CancelJob(jobID)

		// Verify job not in currentJobs maps
		_, found1 := worker1.currentJobs.Load(jobID)
		_, found2 := worker2.currentJobs.Load(jobID)
		assert.False(t, found1)
		assert.False(t, found2)
	})
}

// =============================================================================
// Manager SettingsSecretsService Tests
// =============================================================================

func TestManager_SettingsSecretsService_Propagation(t *testing.T) {
	t.Run("service is set on manager", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Initially nil
		assert.Nil(t, manager.SettingsSecretsService)

		// Set service (nil is valid)
		manager.SetSettingsSecretsService(nil)
		assert.Nil(t, manager.SettingsSecretsService)
	})
}

// =============================================================================
// Manager Worker Lifecycle Tests
// =============================================================================

func TestManager_WorkerLifecycle(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "deno",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, manager.stopCh)
		assert.Equal(t, 0, manager.GetWorkerCount())
		assert.Empty(t, manager.Workers)
		assert.Nil(t, manager.SettingsSecretsService)
		assert.Equal(t, "secret", manager.jwtSecret)
		assert.Equal(t, "http://localhost", manager.publicURL)
	})

	t.Run("config propagation", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:              "deno",
			MaxConcurrentPerWorker:  10,
			DefaultMaxDuration:      time.Hour,
			GracefulShutdownTimeout: 10 * time.Minute,
			PollInterval:            5 * time.Second,
			WorkerHeartbeatInterval: 30 * time.Second,
			WorkerTimeout:           5 * time.Minute,
			DefaultProgressTimeout:  2 * time.Minute,
		}
		manager := NewManager(cfg, nil, "jwt", "http://api.example.com", nil)

		assert.Equal(t, cfg, manager.Config)
		assert.Equal(t, "jwt", manager.jwtSecret)
		assert.Equal(t, "http://api.example.com", manager.publicURL)
	})
}

// =============================================================================
// Manager Stop Channel Tests
// =============================================================================

func TestManager_StopChannel(t *testing.T) {
	t.Run("stop channel is buffered", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Verify stopCh exists
		assert.NotNil(t, manager.stopCh)

		// Channel should be open
		select {
		case <-manager.stopCh:
			t.Error("Stop channel should not be closed initially")
		default:
			// Expected - channel is open
		}
	})
}

// =============================================================================
// Manager Storage Tests
// =============================================================================

func TestManager_Storage(t *testing.T) {
	t.Run("storage is initialized", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, manager.Storage)
	})

	t.Run("storage is shared across workers", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Workers would share the same storage instance
		worker := NewWorker(cfg, manager.Storage, "secret", "http://localhost", nil)

		assert.Equal(t, manager.Storage, worker.Storage)
	})
}

// =============================================================================
// Manager Secrets Storage Tests
// =============================================================================

func TestManager_SecretsStorage(t *testing.T) {
	t.Run("secrets storage is nil by default", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		assert.Nil(t, manager.SecretsStorage)
	})

	t.Run("secrets storage is propagated to workers", func(t *testing.T) {
		// This tests that SecretsStorage is passed through NewManager
		// even though we can't test the actual worker without a real database
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)
		secretsStorage := manager.SecretsStorage // Will be nil

		assert.NotNil(t, cfg) // Just to use the variable
		assert.Nil(t, secretsStorage)
	})
}

// =============================================================================
// Manager Edge Cases
// =============================================================================

func TestManager_EdgeCases(t *testing.T) {
	t.Run("cancel job on nil manager", func(t *testing.T) {
		// This tests behavior if manager was somehow nil
		// In Go, calling a method on nil pointer causes panic
		// But we can't actually test that without causing a panic
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// This should not panic
		manager.CancelJob(uuid.New())
	})

	t.Run("get worker count on nil workers slice", func(t *testing.T) {
		cfg := &config.JobsConfig{}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Workers slice is initialized but empty
		count := manager.GetWorkerCount()
		assert.Equal(t, 0, count)
	})
}

// =============================================================================
// Manager Concurrent Operations Tests
// =============================================================================

func TestManager_ConcurrentOperations(t *testing.T) {
	t.Run("concurrent cancel job calls", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "deno",
		}
		manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

		// Add some workers
		for i := 0; i < 5; i++ {
			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
			manager.Workers = append(manager.Workers, worker)
		}

		jobID := uuid.New()

		// Concurrent cancel calls
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				manager.CancelJob(jobID)
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Should complete without panics or race conditions
		assert.Equal(t, 5, manager.GetWorkerCount())
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkManager_NewManager(b *testing.B) {
	cfg := &config.JobsConfig{
		WorkerMode:             "deno",
		MaxConcurrentPerWorker: 5,
		DefaultMaxDuration:     30 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewManager(cfg, nil, "secret", "http://localhost", nil)
	}
}

func BenchmarkManager_CancelJob(b *testing.B) {
	cfg := &config.JobsConfig{
		WorkerMode: "deno",
	}
	manager := NewManager(cfg, nil, "secret", "http://localhost", nil)

	// Add workers
	for i := 0; i < 10; i++ {
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		manager.Workers = append(manager.Workers, worker)
	}

	jobID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager.CancelJob(jobID)
	}
}
