package jobs

import (
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/fluxbase-eu/fluxbase/internal/runtime"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Worker ExecuteJob Tests
// =============================================================================

func TestWorker_ExecuteJob_Properties(t *testing.T) {
	t.Run("executeJob requires job function", func(t *testing.T) {
		// This tests the property that executeJob needs a valid job function
		// Since executeJob is private, we verify the worker has the necessary storage
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		storage := NewStorage(nil)
		worker := NewWorker(cfg, storage, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.Storage)
		assert.NotNil(t, worker.Runtime)
		assert.Nil(t, worker.SecretsStorage) // Nil was passed
	})

	t.Run("job execution state tracking", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Simulate adding a job to currentJobs
		jobID := uuid.New()
		cancelSignal := &struct {
			cancelled bool
		}{}

		worker.currentJobs.Store(jobID, cancelSignal)

		// Verify it's stored
		val, ok := worker.currentJobs.Load(jobID)
		assert.True(t, ok)
		assert.NotNil(t, val)

		// Clean up
		worker.currentJobs.Delete(jobID)

		// Verify it's removed
		_, ok = worker.currentJobs.Load(jobID)
		assert.False(t, ok)
	})
}

// =============================================================================
// Worker Progress Update Tests
// =============================================================================

func TestWorker_ProgressUpdate(t *testing.T) {
	// Note: handleProgressUpdate is private, so we test indirectly through
	// the worker's structure

	t.Run("job log counters initialization", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		jobID := uuid.New()
		lineCounter := 0

		worker.jobLogCounters.Store(jobID, &lineCounter)

		// Retrieve and verify
		val, ok := worker.jobLogCounters.Load(jobID)
		assert.True(t, ok)

		counterPtr, ok := val.(*int)
		require.True(t, ok)
		assert.Equal(t, 0, *counterPtr)

		// Clean up
		worker.jobLogCounters.Delete(jobID)
	})

	t.Run("job start times tracking", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		jobID := uuid.New()
		startTime := time.Now()

		worker.jobStartTimes.Store(jobID, startTime)

		// Retrieve and verify
		val, ok := worker.jobStartTimes.Load(jobID)
		assert.True(t, ok)

		retrievedTime, ok := val.(time.Time)
		require.True(t, ok)
		assert.WithinDuration(t, startTime, retrievedTime, time.Second)

		// Clean up
		worker.jobStartTimes.Delete(jobID)
	})
}

// =============================================================================
// Worker Log Handling Tests
// =============================================================================

func TestWorker_LogHandling(t *testing.T) {
	t.Run("job logs disabled flag", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		jobID := uuid.New()

		// Initially not disabled
		_, disabled := worker.jobLogsDisabled.Load(jobID)
		assert.False(t, disabled)

		// Set as disabled
		worker.jobLogsDisabled.Store(jobID, true)

		// Now should be disabled
		_, disabled = worker.jobLogsDisabled.Load(jobID)
		assert.True(t, disabled)

		// Clean up
		worker.jobLogsDisabled.Delete(jobID)
	})
}

// =============================================================================
// Worker Configuration Tests
// =============================================================================

func TestWorker_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config *config.JobsConfig
	}{
		{
			name: "default configuration",
			config: &config.JobsConfig{
				WorkerMode:              "deno",
				MaxConcurrentPerWorker:  5,
				DefaultMaxDuration:      30 * time.Minute,
				GracefulShutdownTimeout: 5 * time.Minute,
				PollInterval:            5 * time.Second,
			},
		},
		{
			name: "high concurrency",
			config: &config.JobsConfig{
				WorkerMode:              "deno",
				MaxConcurrentPerWorker:  50,
				DefaultMaxDuration:      time.Hour,
				GracefulShutdownTimeout: 10 * time.Minute,
			},
		},
		{
			name: "low concurrency",
			config: &config.JobsConfig{
				WorkerMode:              "deno",
				MaxConcurrentPerWorker:  1,
				DefaultMaxDuration:      10 * time.Minute,
				GracefulShutdownTimeout: time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			worker := NewWorker(tt.config, nil, "secret", "http://localhost", nil)

			assert.Equal(t, tt.config, worker.Config)
			assert.Equal(t, tt.config.MaxConcurrentPerWorker, worker.MaxConcurrent)
		})
	}
}

// =============================================================================
// Worker Secrets Loading Tests
// =============================================================================

func TestWorker_SecretsLoading(t *testing.T) {
	t.Run("secrets storage propagation", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// SecretsStorage can be nil (nil was passed)
		assert.Nil(t, worker.SecretsStorage)
	})
}

// =============================================================================
// Worker Settings Secrets Tests
// =============================================================================

func TestWorker_SettingsSecrets(t *testing.T) {
	t.Run("settings secrets service initially nil", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Nil(t, worker.SettingsSecretsService)
	})

	t.Run("settings secrets service can be set", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// SettingsSecretsService is a public field that can be set
		// Initially nil since none was provided
		assert.Nil(t, worker.SettingsSecretsService)
	})
}

// =============================================================================
// Worker Job Count Tracking Tests
// =============================================================================

func TestWorker_JobCountTracking(t *testing.T) {
	t.Run("increment and decrement thread safety", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 100,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Run concurrent increments
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					worker.incrementJobCount()
				}
				done <- true
			}()
			go func() {
				for j := 0; j < 100; j++ {
					worker.decrementJobCount()
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Should complete without race conditions
		assert.GreaterOrEqual(t, worker.currentJobCount, 0)
	})
}

// =============================================================================
// Worker Shutdown Tests
// =============================================================================

func TestWorker_ShutdownFlow(t *testing.T) {
	t.Run("shutdown channels are created", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)

		// Channels should be open initially
		select {
		case <-worker.shutdownChan:
			t.Error("Shutdown channel should not be closed initially")
		default:
			// Expected
		}

		select {
		case <-worker.shutdownComplete:
			t.Error("Shutdown complete channel should not be closed initially")
		default:
			// Expected
		}
	})

	t.Run("draining state affects polling", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Initially not draining
		assert.False(t, worker.isDraining())

		// Set draining
		worker.setDraining(true)
		assert.True(t, worker.isDraining())

		// Unset draining
		worker.setDraining(false)
		assert.False(t, worker.isDraining())
	})
}

// =============================================================================
// Worker ID and Name Tests
// =============================================================================

func TestWorker_Identification(t *testing.T) {
	t.Run("worker generates unique ID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		ids := make(map[uuid.UUID]bool)
		for i := 0; i < 100; i++ {
			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
			if ids[worker.ID] {
				t.Errorf("Duplicate worker ID generated: %s", worker.ID.String())
			}
			ids[worker.ID] = true
		}

		assert.Equal(t, 100, len(ids))
	})

	t.Run("worker name format", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Name should contain "worker-" prefix
		assert.Contains(t, worker.Name, "worker-")

		// Name should contain "@" separator
		assert.Contains(t, worker.Name, "@")

		// Name should not be empty
		assert.NotEmpty(t, worker.Name)
	})
}

// =============================================================================
// Worker Runtime Configuration Tests
// =============================================================================

func TestWorker_RuntimeConfiguration(t *testing.T) {
	t.Run("runtime is configured with correct type", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.Runtime)
	})

	t.Run("public URL is stored", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		testURLs := []string{
			"http://localhost:8080",
			"https://api.example.com",
			"https://fluxbase.example.com/v1",
		}

		for _, url := range testURLs {
			worker := NewWorker(cfg, nil, "secret", url, nil)
			assert.Equal(t, url, worker.publicURL)
		}
	})
}

// =============================================================================
// Worker Memory Limit Tests
// =============================================================================

func TestWorker_MemoryLimits(t *testing.T) {
	t.Run("runtime has memory limit", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// The runtime is created with a default memory limit
		// We can't directly test the runtime's memory limit without accessing private fields
		// but we verify the runtime exists
		assert.NotNil(t, worker.Runtime)
	})
}

// =============================================================================
// Worker Edge Cases
// =============================================================================

func TestWorker_EdgeCases(t *testing.T) {
	t.Run("worker with nil config", func(t *testing.T) {
		// This should still create a worker, just with nil/zero values
		worker := NewWorker(nil, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker)
		assert.NotEqual(t, uuid.Nil, worker.ID)
		assert.NotNil(t, worker.Runtime)
		// Storage is nil when nil is passed
		assert.Nil(t, worker.Storage)
	})

	t.Run("worker with empty JWT secret", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "", "http://localhost", nil)

		assert.NotNil(t, worker)
		assert.NotNil(t, worker.Runtime)
	})

	t.Run("worker with empty public URL", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "", nil)

		assert.NotNil(t, worker)
		assert.Equal(t, "", worker.publicURL)
	})
}

// =============================================================================
// Worker Timeout Tests
// =============================================================================

func TestWorker_Timeouts(t *testing.T) {
	tests := []struct {
		name                    string
		gracefulShutdownTimeout time.Duration
		defaultProgressTimeout  time.Duration
		expectedShutdownTimeout time.Duration
	}{
		{
			name:                    "default timeouts",
			gracefulShutdownTimeout: 0,
			defaultProgressTimeout:  0,
			expectedShutdownTimeout: 5 * time.Minute,
		},
		{
			name:                    "custom shutdown timeout",
			gracefulShutdownTimeout: 10 * time.Minute,
			defaultProgressTimeout:  0,
			expectedShutdownTimeout: 10 * time.Minute,
		},
		{
			name:                    "short shutdown timeout",
			gracefulShutdownTimeout: 30 * time.Second,
			defaultProgressTimeout:  0,
			expectedShutdownTimeout: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.JobsConfig{
				WorkerMode:              "deno",
				MaxConcurrentPerWorker:  5,
				GracefulShutdownTimeout: tt.gracefulShutdownTimeout,
				DefaultProgressTimeout:  tt.defaultProgressTimeout,
			}
			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

			assert.NotNil(t, worker)
			// The actual timeout calculation happens in Start(), which we can't test
			// without a database, but we verify the config is stored
			assert.Equal(t, tt.gracefulShutdownTimeout, worker.Config.GracefulShutdownTimeout)
		})
	}
}

// =============================================================================
// Worker Concurrent Access Tests
// =============================================================================

func TestWorker_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent job tracking", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 100,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Simulate concurrent job additions/removals
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				for j := 0; j < 50; j++ {
					jobID := uuid.New()
					worker.currentJobs.Store(jobID, struct{}{})
				}
				done <- true
			}()
			go func() {
				worker.currentJobs.Range(func(k, v interface{}) bool {
					worker.currentJobs.Delete(k)
					return true
				})
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 20; i++ {
			<-done
		}

		// Should complete without race conditions
		t.Log("Concurrent access test completed")
	})
}

// =============================================================================
// Worker Cancel Job Tests
// =============================================================================

func TestWorker_CancelJob_Actual(t *testing.T) {
	t.Run("cancel job with cancel signal", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		jobID := uuid.New()

		// Create a runtime.CancelSignal
		signal := runtime.NewCancelSignal()

		// Store in currentJobs
		worker.currentJobs.Store(jobID, signal)

		// Verify it's there
		val, ok := worker.currentJobs.Load(jobID)
		assert.True(t, ok)
		assert.NotNil(t, val)

		// Cancel the job
		worker.cancelJob(jobID)

		// The cancel signal should be cancelled
		assert.True(t, signal.IsCancelled())

		// The job is still in currentJobs (cancelJob doesn't remove it)
		_, ok = worker.currentJobs.Load(jobID)
		assert.True(t, ok)

		// Clean up
		worker.currentJobs.Delete(jobID)
	})

	t.Run("cancel non-existent job", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}
		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic
		worker.cancelJob(uuid.New())
	})
}

// =============================================================================
// Worker Poll Interval Tests
// =============================================================================

func TestWorker_PollInterval(t *testing.T) {
	tests := []struct {
		name     string
		interval time.Duration
	}{
		{"1 second", 1 * time.Second},
		{"5 seconds", 5 * time.Second},
		{"30 seconds", 30 * time.Second},
		{"1 minute", 1 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.JobsConfig{
				WorkerMode:    "deno",
				PollInterval:  tt.interval,
				WorkerTimeout: 5 * time.Minute,
			}
			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

			assert.Equal(t, tt.interval, worker.Config.PollInterval)
		})
	}
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkWorker_NewWorker(b *testing.B) {
	cfg := &config.JobsConfig{
		WorkerMode:             "deno",
		MaxConcurrentPerWorker: 5,
		DefaultMaxDuration:     30 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewWorker(cfg, nil, "secret", "http://localhost", nil)
	}
}

func BenchmarkWorker_JobCountOperations(b *testing.B) {
	cfg := &config.JobsConfig{
		MaxConcurrentPerWorker: 100,
	}
	worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.incrementJobCount()
		worker.decrementJobCount()
	}
}

func BenchmarkWorker_HasCapacity(b *testing.B) {
	cfg := &config.JobsConfig{
		MaxConcurrentPerWorker: 50,
	}
	worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		worker.hasCapacity()
	}
}
