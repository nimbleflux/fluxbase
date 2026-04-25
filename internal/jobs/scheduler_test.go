package jobs

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/scheduler"
)

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil database", func(t *testing.T) {
		s := NewScheduler(nil)

		require.NotNil(t, s)
		assert.NotNil(t, s.inner)
		assert.NotNil(t, s.storage)
		assert.Equal(t, 20, s.inner.Guard.MaxConcurrent)
		assert.Equal(t, 0, s.inner.EntryCount())
		assert.NotNil(t, s.inner.Context())
	})

	t.Run("initializes empty job entries", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.Equal(t, 0, s.inner.EntryCount())
	})
}

func TestScheduler_Stop(t *testing.T) {
	t.Run("stop cancels context", func(t *testing.T) {
		s := NewScheduler(nil)
		s.Stop()

		select {
		case <-s.inner.Context().Done():
		default:
			t.Error("Context should be cancelled after Stop()")
		}
	})
}

func TestScheduler_MaxConcurrent(t *testing.T) {
	t.Run("default max concurrent is 20", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.Equal(t, 20, s.inner.Guard.MaxConcurrent)
	})
}

func TestScheduler_IsScheduled(t *testing.T) {
	t.Run("returns false for unscheduled job", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.False(t, s.IsScheduled("default", "non-existent"))
	})

	t.Run("returns true for scheduled job", func(t *testing.T) {
		s := NewScheduler(nil)

		_, err := s.inner.AddFunc("default/scheduled-job", "* * * * *", func() {})
		require.NoError(t, err)

		assert.True(t, s.IsScheduled("default", "scheduled-job"))
	})
}

func TestScheduler_GetScheduledJobs(t *testing.T) {
	t.Run("returns empty list initially", func(t *testing.T) {
		s := NewScheduler(nil)
		jobs := s.GetScheduledJobs()
		assert.Empty(t, jobs)
	})

	t.Run("returns scheduled job names", func(t *testing.T) {
		s := NewScheduler(nil)

		_, err := s.inner.AddFunc("job-1", "* * * * *", func() {})
		require.NoError(t, err)
		_, err = s.inner.AddFunc("job-2", "* * * * *", func() {})
		require.NoError(t, err)

		jobs := s.GetScheduledJobs()
		assert.Len(t, jobs, 2)

		keys := make([]string, len(jobs))
		for i, job := range jobs {
			keys[i] = job.Key
		}
		assert.Contains(t, keys, "job-1")
		assert.Contains(t, keys, "job-2")
	})
}

func TestScheduler_UnscheduleJob(t *testing.T) {
	t.Run("removes scheduled job", func(t *testing.T) {
		s := NewScheduler(nil)

		_, err := s.inner.AddFunc("default:to-remove", "* * * * *", func() {})
		require.NoError(t, err)

		s.UnscheduleJob("default", "to-remove")
		assert.False(t, s.IsScheduled("default", "to-remove"))
	})

	t.Run("handles non-existent job gracefully", func(t *testing.T) {
		s := NewScheduler(nil)
		s.UnscheduleJob("default", "non-existent")
	})
}

func TestScheduler_AdditionalCoverage(t *testing.T) {
	t.Run("scheduler context is not nil", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.NotNil(t, s.inner.Context())
	})

	t.Run("scheduler storage is initialized", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.NotNil(t, s.storage)
	})

	t.Run("scheduler max concurrent defaults to 20", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.Equal(t, 20, s.inner.Guard.MaxConcurrent)
	})

	t.Run("unschedule job with empty namespace", func(t *testing.T) {
		s := NewScheduler(nil)
		s.UnscheduleJob("", "job-name")
	})

	t.Run("unschedule job with empty job name", func(t *testing.T) {
		s := NewScheduler(nil)
		s.UnscheduleJob("default", "")
	})

	t.Run("check is scheduled with empty inputs", func(t *testing.T) {
		s := NewScheduler(nil)
		assert.False(t, s.IsScheduled("", ""))
	})

	t.Run("get scheduled jobs from empty scheduler", func(t *testing.T) {
		s := NewScheduler(nil)
		jobs := s.GetScheduledJobs()
		assert.NotNil(t, jobs)
		assert.Empty(t, jobs)
	})

	t.Run("multiple stop calls are safe", func(t *testing.T) {
		s := NewScheduler(nil)
		s.Stop()
		s.Stop()
		s.Stop()
	})
}

func TestScheduleConfig_Struct(t *testing.T) {
	t.Run("basic schedule config", func(t *testing.T) {
		config := ScheduleConfig{
			CronExpression: "*/5 * * * *",
		}

		assert.Equal(t, "*/5 * * * *", config.CronExpression)
		assert.Nil(t, config.Params)
	})

	t.Run("schedule config with params", func(t *testing.T) {
		config := ScheduleConfig{
			CronExpression: "0 * * * *",
			Params: map[string]interface{}{
				"batch_size": 100,
				"dry_run":    false,
			},
		}

		assert.Equal(t, "0 * * * *", config.CronExpression)
		assert.NotNil(t, config.Params)
		assert.Equal(t, 100, config.Params["batch_size"])
		assert.Equal(t, false, config.Params["dry_run"])
	})

	t.Run("JSON serialization", func(t *testing.T) {
		config := ScheduleConfig{
			CronExpression: "0 0 * * *",
			Params: map[string]interface{}{
				"key": "value",
			},
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		assert.Contains(t, string(data), `"cron_expression":"0 0 * * *"`)
		assert.Contains(t, string(data), `"params"`)
	})

	t.Run("JSON deserialization", func(t *testing.T) {
		jsonData := `{"cron_expression":"*/10 * * * *","params":{"limit":50}}`

		var config ScheduleConfig
		err := json.Unmarshal([]byte(jsonData), &config)
		require.NoError(t, err)

		assert.Equal(t, "*/10 * * * *", config.CronExpression)
		assert.Equal(t, float64(50), config.Params["limit"])
	})
}

func TestScheduleConfig_Additional(t *testing.T) {
	t.Run("empty params map", func(t *testing.T) {
		config := ScheduleConfig{
			CronExpression: "*/5 * * * *",
			Params:         map[string]interface{}{},
		}

		assert.NotNil(t, config.Params)
		assert.Empty(t, config.Params)
	})

	t.Run("params with complex types", func(t *testing.T) {
		config := ScheduleConfig{
			CronExpression: "0 * * * *",
			Params: map[string]interface{}{
				"string": "value",
				"number": 42,
				"float":  3.14,
				"bool":   true,
				"null":   nil,
				"array":  []interface{}{1, 2, 3},
				"nested": map[string]interface{}{"key": "value"},
			},
		}

		assert.Equal(t, "value", config.Params["string"])
		assert.Equal(t, 42, config.Params["number"])
		assert.Equal(t, 3.14, config.Params["float"])
		assert.Equal(t, true, config.Params["bool"])
		assert.Nil(t, config.Params["null"])
		assert.Len(t, config.Params["array"], 3)
	})
}

func TestSharedValidator(t *testing.T) {
	t.Run("delegates to shared ValidateCronSchedule", func(t *testing.T) {
		assert.NoError(t, scheduler.ValidateCronSchedule("*/5 * * * *"))
		assert.Error(t, scheduler.ValidateCronSchedule("invalid"))
	})
}
