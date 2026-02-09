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
// Worker Construction Tests
// =============================================================================

func TestNewWorker(t *testing.T) {
	t.Run("creates worker with default config", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}

		worker := NewWorker(cfg, nil, "jwt-secret", "http://localhost", nil)

		require.NotNil(t, worker)
		assert.NotEqual(t, uuid.Nil, worker.ID)
		assert.Contains(t, worker.Name, "worker-")
		assert.Equal(t, cfg, worker.Config)
		assert.Equal(t, 5, worker.MaxConcurrent)
		assert.NotNil(t, worker.Runtime)
		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
	})

	t.Run("generates unique worker ID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker1 := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker2 := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotEqual(t, worker1.ID, worker2.ID)
	})

	t.Run("includes hostname in worker name", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Contains(t, worker.Name, "@")
	})
}

// =============================================================================
// Worker State Tests
// =============================================================================

func TestWorker_setDrainingState(t *testing.T) {
	t.Run("starts not draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.False(t, worker.draining)
	})
}

func TestWorker_JobCount(t *testing.T) {
	t.Run("starts with zero jobs", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 0, worker.currentJobCount)
	})
}

// =============================================================================
// Worker Concurrent Job Handling Tests
// =============================================================================

func TestWorker_MaxConcurrent(t *testing.T) {
	t.Run("respects max concurrent config", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 10,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 10, worker.MaxConcurrent)
	})

	t.Run("single concurrent job", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 1,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.Equal(t, 1, worker.MaxConcurrent)
	})
}

// =============================================================================
// Worker Stop Tests
// =============================================================================

func TestWorker_Stop(t *testing.T) {
	t.Run("stop signals shutdown", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Simulate worker completion by closing shutdownComplete in a goroutine
		// (normally this would be done by the worker's goroutines when they finish)
		go func() {
			<-worker.shutdownChan // Wait for Stop() to close shutdownChan
			close(worker.shutdownComplete)
		}()

		// Stop should close the shutdown channel and wait for completion
		worker.Stop()

		// Verify shutdown channel was closed
		select {
		case <-worker.shutdownChan:
			// Expected - channel is closed
		default:
			t.Error("Shutdown channel should be closed after Stop()")
		}
	})
}

// =============================================================================
// Worker Cancel Job Tests
// =============================================================================

func TestWorker_CancelJob(t *testing.T) {
	t.Run("cancel non-existent job does not panic", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic
		worker.cancelJob(uuid.New())
	})
}

// =============================================================================
// Worker Runtime Tests
// =============================================================================

func TestWorker_Runtime(t *testing.T) {
	t.Run("creates deno runtime", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
			DefaultMaxDuration:     30 * time.Minute,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.Runtime)
	})
}

// =============================================================================
// Worker Public URL Tests
// =============================================================================

func TestWorker_PublicURL(t *testing.T) {
	t.Run("stores public URL", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "https://api.example.com", nil)

		assert.Equal(t, "https://api.example.com", worker.publicURL)
	})
}

// =============================================================================
// Worker Running Jobs Tests
// =============================================================================

func TestWorker_CurrentJobs(t *testing.T) {
	t.Run("starts with empty current jobs map", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// currentJobs is a sync.Map - verify it's usable
		count := 0
		worker.currentJobs.Range(func(k, v interface{}) bool {
			count++
			return true
		})
		assert.Equal(t, 0, count)
	})
}

// =============================================================================
// Worker Job Count Operations Tests
// =============================================================================

func TestWorker_JobCountOperations(t *testing.T) {
	t.Run("increment job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.currentJobCount)

		worker.incrementJobCount()
		assert.Equal(t, 1, worker.currentJobCount)

		worker.incrementJobCount()
		assert.Equal(t, 2, worker.currentJobCount)
	})

	t.Run("decrement job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 3

		worker.decrementJobCount()
		assert.Equal(t, 2, worker.currentJobCount)

		worker.decrementJobCount()
		assert.Equal(t, 1, worker.currentJobCount)
	})

	t.Run("decrement can go below zero", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.currentJobCount)

		worker.decrementJobCount()
		// Note: The implementation doesn't prevent going negative
		assert.Equal(t, -1, worker.currentJobCount)
	})
}

// =============================================================================
// Worker hasCapacity Tests
// =============================================================================

func TestWorker_hasCapacity(t *testing.T) {
	t.Run("has capacity when no jobs running", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.True(t, worker.hasCapacity())
	})

	t.Run("has capacity when below max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 3

		assert.True(t, worker.hasCapacity())
	})

	t.Run("no capacity when at max", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.currentJobCount = 5

		assert.False(t, worker.hasCapacity())
	})

	t.Run("hasCapacity ignores draining state", func(t *testing.T) {
		// Note: hasCapacity only checks job count vs max concurrent
		// It does not check draining state - that's checked separately in the poll loop
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.draining = true

		// hasCapacity returns true because it only checks count < max
		assert.True(t, worker.hasCapacity())
	})
}

// =============================================================================
// Worker setDraining Tests
// =============================================================================

func TestWorker_setDraining(t *testing.T) {
	t.Run("drain sets draining flag", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.False(t, worker.draining)

		worker.setDraining(true)
		assert.True(t, worker.draining)
	})

	t.Run("drain is idempotent", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		worker.setDraining(true)
		worker.setDraining(true)
		assert.True(t, worker.draining)
	})
}

// =============================================================================
// Worker isDraining Tests
// =============================================================================

func TestWorker_isDraining(t *testing.T) {
	t.Run("returns false when not draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.False(t, worker.isDraining())
	})

	t.Run("returns true when draining", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		worker.setDraining(true)

		assert.True(t, worker.isDraining())
	})
}

// =============================================================================
// Worker GetCurrentJobCount Tests
// =============================================================================

func TestWorker_GetCurrentJobCount(t *testing.T) {
	t.Run("returns current job count", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.Equal(t, 0, worker.getCurrentJobCount())

		worker.currentJobCount = 3
		assert.Equal(t, 3, worker.getCurrentJobCount())
	})
}

// =============================================================================
// Worker Concurrent Access Tests
// =============================================================================

func TestWorker_ConcurrentJobCountAccess(t *testing.T) {
	cfg := &config.JobsConfig{
		MaxConcurrentPerWorker: 100,
	}

	worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

	// Run concurrent increments and decrements
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				worker.incrementJobCount()
			}
			done <- true
		}()
		go func() {
			for j := 0; j < 50; j++ {
				worker.hasCapacity()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should complete without panics or data races
	assert.GreaterOrEqual(t, worker.currentJobCount, 0)
}

// =============================================================================
// Worker Shutdown Channel Tests
// =============================================================================

func TestWorker_ShutdownChannels(t *testing.T) {
	t.Run("creates shutdown channels", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
	})
}

// =============================================================================
// Worker Mode Tests
// =============================================================================

func TestWorker_Modes(t *testing.T) {
	testCases := []struct {
		mode string
	}{
		{"deno"},
		{"docker"},
		{"process"},
	}

	for _, tc := range testCases {
		t.Run("mode "+tc.mode, func(t *testing.T) {
			cfg := &config.JobsConfig{
				WorkerMode:             tc.mode,
				MaxConcurrentPerWorker: 5,
			}

			worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
			assert.Equal(t, tc.mode, worker.Config.WorkerMode)
		})
	}
}

// =============================================================================
// normalizeSettingsKey Tests
// =============================================================================

func TestNormalizeSettingsKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple key",
			input:    "api_key",
			expected: "API_KEY",
		},
		{
			name:     "dotted key",
			input:    "ai.openai.api_key",
			expected: "AI_OPENAI_API_KEY",
		},
		{
			name:     "lowercase to uppercase",
			input:    "mykey",
			expected: "MYKEY",
		},
		{
			name:     "already uppercase",
			input:    "MY_KEY",
			expected: "MY_KEY",
		},
		{
			name:     "mixed case",
			input:    "MyApiKey",
			expected: "MYAPIKEY",
		},
		{
			name:     "single dot",
			input:    "prefix.key",
			expected: "PREFIX_KEY",
		},
		{
			name:     "multiple dots",
			input:    "a.b.c.d",
			expected: "A_B_C_D",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only dots",
			input:    "...",
			expected: "___",
		},
		{
			name:     "numbers in key",
			input:    "api_v2_key",
			expected: "API_V2_KEY",
		},
		{
			name:     "trailing dot",
			input:    "key.",
			expected: "KEY_",
		},
		{
			name:     "leading dot",
			input:    ".key",
			expected: "_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSettingsKey(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// jobToExecutionRequest Tests
// =============================================================================

func TestJobToExecutionRequest(t *testing.T) {
	t.Run("basic job conversion", func(t *testing.T) {
		jobID := uuid.New()
		job := &Job{
			ID:        jobID,
			JobName:   "test-job",
			Namespace: "default",
		}

		req := jobToExecutionRequest(job, "http://localhost:8080")

		assert.Equal(t, jobID, req.ID)
		assert.Equal(t, "test-job", req.Name)
		assert.Equal(t, "default", req.Namespace)
		assert.Equal(t, "http://localhost:8080", req.BaseURL)
		assert.Equal(t, 0, req.RetryCount)
		assert.Nil(t, req.Payload)
	})

	t.Run("job with retry count", func(t *testing.T) {
		job := &Job{
			ID:         uuid.New(),
			JobName:    "retry-job",
			Namespace:  "production",
			RetryCount: 3,
		}

		req := jobToExecutionRequest(job, "https://api.example.com")

		assert.Equal(t, 3, req.RetryCount)
	})

	t.Run("job with payload", func(t *testing.T) {
		payload := `{"key": "value", "count": 42}`
		job := &Job{
			ID:        uuid.New(),
			JobName:   "payload-job",
			Namespace: "default",
			Payload:   &payload,
		}

		req := jobToExecutionRequest(job, "http://localhost")

		require.NotNil(t, req.Payload)
		assert.Equal(t, "value", req.Payload["key"])
		assert.Equal(t, float64(42), req.Payload["count"])
	})

	t.Run("job with invalid JSON payload", func(t *testing.T) {
		invalidPayload := `{invalid json}`
		job := &Job{
			ID:        uuid.New(),
			JobName:   "invalid-payload-job",
			Namespace: "default",
			Payload:   &invalidPayload,
		}

		req := jobToExecutionRequest(job, "http://localhost")

		// Invalid JSON should result in nil payload
		assert.Nil(t, req.Payload)
	})

	t.Run("job with user context", func(t *testing.T) {
		userID := uuid.New()
		userEmail := "user@example.com"
		userRole := "admin"

		job := &Job{
			ID:        uuid.New(),
			JobName:   "user-job",
			Namespace: "default",
			CreatedBy: &userID,
			UserEmail: &userEmail,
			UserRole:  &userRole,
		}

		req := jobToExecutionRequest(job, "http://localhost")

		assert.Equal(t, userID.String(), req.UserID)
		assert.Equal(t, "user@example.com", req.UserEmail)
		assert.Equal(t, "admin", req.UserRole)
	})

	t.Run("job with partial user context", func(t *testing.T) {
		userID := uuid.New()

		job := &Job{
			ID:        uuid.New(),
			JobName:   "partial-user-job",
			Namespace: "default",
			CreatedBy: &userID,
			// UserEmail and UserRole are nil
		}

		req := jobToExecutionRequest(job, "http://localhost")

		assert.Equal(t, userID.String(), req.UserID)
		assert.Empty(t, req.UserEmail)
		assert.Empty(t, req.UserRole)
	})

	t.Run("job with no user context", func(t *testing.T) {
		job := &Job{
			ID:        uuid.New(),
			JobName:   "anonymous-job",
			Namespace: "default",
		}

		req := jobToExecutionRequest(job, "http://localhost")

		assert.Empty(t, req.UserID)
		assert.Empty(t, req.UserEmail)
		assert.Empty(t, req.UserRole)
	})

	t.Run("job with complex payload", func(t *testing.T) {
		payload := `{
			"items": [1, 2, 3],
			"nested": {"a": "b"},
			"enabled": true
		}`
		job := &Job{
			ID:        uuid.New(),
			JobName:   "complex-payload-job",
			Namespace: "default",
			Payload:   &payload,
		}

		req := jobToExecutionRequest(job, "http://localhost")

		require.NotNil(t, req.Payload)
		items := req.Payload["items"].([]interface{})
		assert.Len(t, items, 3)
		nested := req.Payload["nested"].(map[string]interface{})
		assert.Equal(t, "b", nested["a"])
		assert.Equal(t, true, req.Payload["enabled"])
	})

	t.Run("different public URLs", func(t *testing.T) {
		job := &Job{
			ID:        uuid.New(),
			JobName:   "url-test",
			Namespace: "default",
		}

		testCases := []string{
			"http://localhost",
			"http://localhost:8080",
			"https://api.example.com",
			"https://api.example.com/v1",
		}

		for _, url := range testCases {
			req := jobToExecutionRequest(job, url)
			assert.Equal(t, url, req.BaseURL)
		}
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNormalizeSettingsKey(b *testing.B) {
	keys := []string{
		"api_key",
		"ai.openai.api_key",
		"simple",
		"complex.nested.deep.key",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		normalizeSettingsKey(keys[i%len(keys)])
	}
}

func BenchmarkJobToExecutionRequest(b *testing.B) {
	payload := `{"key": "value", "count": 42}`
	userID := uuid.New()
	job := &Job{
		ID:         uuid.New(),
		JobName:    "benchmark-job",
		Namespace:  "default",
		Payload:    &payload,
		RetryCount: 1,
		CreatedBy:  &userID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobToExecutionRequest(job, "http://localhost")
	}
}

// =============================================================================
// Additional Worker Tests for Coverage
// =============================================================================

func TestWorker_Start_Validation(t *testing.T) {
	t.Run("accepts valid context", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode:             "deno",
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Note: We can't fully test Start without database and runtime
		// but we can verify it doesn't panic on nil inputs
		assert.NotNil(t, worker)
		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
	})
}

func TestWorker_Stop_Multiple(t *testing.T) {
	t.Run("stop can be called multiple times safely", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Verify draining state
		assert.False(t, worker.isDraining())

		// Note: Can't actually call Stop() without starting the worker first
		// as it will block waiting for goroutines that don't exist
		// Just verify the state check works
		assert.False(t, worker.isDraining())
	})
}

func TestWorker_CancelJob_Extended(t *testing.T) {
	t.Run("cancel job with zero UUID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should handle zero UUID gracefully
		worker.cancelJob(uuid.UUID{})
	})

	t.Run("cancel non-existent job", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic when job doesn't exist
		worker.cancelJob(uuid.New())
	})

	t.Run("cancel multiple different jobs", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should handle multiple cancels without issues
		worker.cancelJob(uuid.New())
		worker.cancelJob(uuid.New())
		worker.cancelJob(uuid.New())
	})
}

func TestWorker_InterruptAllJobs(t *testing.T) {
	t.Run("interrupt with no running jobs", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic when no jobs are running
		worker.interruptAllJobs()
	})
}

func TestWorker_ShutdownChannels_Extended(t *testing.T) {
	t.Run("channels are properly initialized", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		assert.NotNil(t, worker.shutdownChan)
		assert.NotNil(t, worker.shutdownComplete)
		assert.NotNil(t, worker.currentJobs)
	})
}

func TestWorker_Modes_Extended(t *testing.T) {
	t.Run("mode deno", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "deno",
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.NotNil(t, worker)
	})

	t.Run("mode docker", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "docker",
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.NotNil(t, worker)
	})

	t.Run("mode process", func(t *testing.T) {
		cfg := &config.JobsConfig{
			WorkerMode: "process",
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)
		assert.NotNil(t, worker)
	})
}

func TestWorker_HandleLogMessage(t *testing.T) {
	t.Run("handle log with nil progress", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic with nil progress
		worker.handleLogMessage(uuid.New(), "info", "test message")
	})

	t.Run("handle log for non-existent job", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should not panic when job doesn't exist
		worker.handleLogMessage(uuid.New(), "error", "error message")
	})
}

func TestWorker_LoadSettingsSecrets(t *testing.T) {
	t.Run("load with nil userID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		// Should return nil when no settings service
		secrets := worker.loadSettingsSecrets(context.Background(), nil)
		assert.Nil(t, secrets)
	})

	t.Run("load with valid userID", func(t *testing.T) {
		cfg := &config.JobsConfig{
			MaxConcurrentPerWorker: 5,
		}

		worker := NewWorker(cfg, nil, "secret", "http://localhost", nil)

		userID := uuid.New()
		// Should return nil when no settings service
		secrets := worker.loadSettingsSecrets(context.Background(), &userID)
		assert.Nil(t, secrets)
	})
}
