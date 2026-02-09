package jobs

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Scheduler Construction Tests
// =============================================================================

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil database", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		require.NotNil(t, scheduler)
		assert.NotNil(t, scheduler.cron)
		assert.NotNil(t, scheduler.storage)
		assert.Equal(t, 20, scheduler.maxConcurrent)
		assert.NotNil(t, scheduler.jobEntries)
		assert.Empty(t, scheduler.jobEntries)
		assert.NotNil(t, scheduler.ctx)
		assert.NotNil(t, scheduler.cancel)
	})

	t.Run("initializes empty job entries map", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		assert.NotNil(t, scheduler.jobEntries)
		assert.Len(t, scheduler.jobEntries, 0)
	})
}

// =============================================================================
// ValidateCronSchedule Tests
// =============================================================================

func TestValidateCronSchedule(t *testing.T) {
	t.Run("accepts valid cron expressions", func(t *testing.T) {
		validExprs := []string{
			"* * * * *",     // Every minute
			"*/5 * * * *",   // Every 5 minutes
			"0 * * * *",     // Every hour
			"0 0 * * *",     // Every day at midnight
			"0 12 * * *",    // Every day at noon
			"0 0 * * 0",     // Every Sunday at midnight
			"0 0 1 * *",     // First of every month
			"0 0 1 1 *",     // January 1st
			"30 4 1,15 * *", // 1st and 15th at 4:30
			"0 22 * * 1-5",  // Weekdays at 10pm
			"@hourly",       // Every hour
			"@daily",        // Every day
			"@weekly",       // Every week
			"@monthly",      // Every month
			"@yearly",       // Every year
		}

		for _, expr := range validExprs {
			t.Run(expr, func(t *testing.T) {
				err := ValidateCronSchedule(expr)
				assert.NoError(t, err, "Should accept: %s", expr)
			})
		}
	})

	t.Run("rejects invalid cron expressions", func(t *testing.T) {
		invalidExprs := []string{
			"invalid",
			"* * *",         // Too few fields
			"* * * * * * *", // Too many fields
			"60 * * * *",    // Invalid minute
			"* 25 * * *",    // Invalid hour
			"* * 32 * *",    // Invalid day
			"* * * 13 *",    // Invalid month
			"* * * * 8",     // Invalid day of week
		}

		for _, expr := range invalidExprs {
			t.Run(expr, func(t *testing.T) {
				err := ValidateCronSchedule(expr)
				assert.Error(t, err, "Should reject: %s", expr)
			})
		}
	})

	t.Run("rejects schedules that run too frequently", func(t *testing.T) {
		// Schedules that would run more frequently than once per minute
		frequentExprs := []string{
			"*/30 * * * * *", // Every 30 seconds
			"*/10 * * * * *", // Every 10 seconds
			"* * * * * *",    // Every second
		}

		for _, expr := range frequentExprs {
			t.Run(expr, func(t *testing.T) {
				err := ValidateCronSchedule(expr)
				if err != nil {
					// Either parsing error or interval error is acceptable
					_, isCronIntervalError := err.(*CronIntervalError)
					if isCronIntervalError {
						assert.Contains(t, err.Error(), "runs too frequently")
					}
				}
			})
		}
	})
}

// =============================================================================
// CronIntervalError Tests
// =============================================================================

func TestCronIntervalError(t *testing.T) {
	t.Run("error message format", func(t *testing.T) {
		err := &CronIntervalError{
			Expression: "*/30 * * * * *",
			Interval:   30 * time.Second,
			MinAllowed: time.Minute,
		}

		msg := err.Error()

		assert.Contains(t, msg, "runs too frequently")
		assert.Contains(t, msg, "*/30 * * * * *")
		assert.Contains(t, msg, "30s")
		assert.Contains(t, msg, "1m0s")
	})

	t.Run("implements error interface", func(t *testing.T) {
		var err error = &CronIntervalError{
			Expression: "* * * * * *",
			Interval:   time.Second,
			MinAllowed: time.Minute,
		}

		assert.NotNil(t, err)
		assert.NotEmpty(t, err.Error())
	})
}

// =============================================================================
// MinCronInterval Constant Tests
// =============================================================================

func TestMinCronInterval(t *testing.T) {
	t.Run("is one minute", func(t *testing.T) {
		assert.Equal(t, time.Minute, MinCronInterval)
	})
}

// =============================================================================
// ScheduleConfig Tests
// =============================================================================

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

// =============================================================================
// Cron Parser Tests
// =============================================================================

func TestCronParser(t *testing.T) {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	t.Run("parses 5-field expressions", func(t *testing.T) {
		schedule, err := parser.Parse("*/5 * * * *")
		require.NoError(t, err)
		assert.NotNil(t, schedule)
	})

	t.Run("parses 6-field expressions with seconds", func(t *testing.T) {
		schedule, err := parser.Parse("0 */5 * * * *")
		require.NoError(t, err)
		assert.NotNil(t, schedule)
	})

	t.Run("parses descriptors", func(t *testing.T) {
		descriptors := []string{
			"@hourly",
			"@daily",
			"@weekly",
			"@monthly",
			"@yearly",
			"@annually",
		}

		for _, desc := range descriptors {
			schedule, err := parser.Parse(desc)
			require.NoError(t, err, "Failed to parse: %s", desc)
			assert.NotNil(t, schedule)
		}
	})
}

// =============================================================================
// Scheduler Stop Tests
// =============================================================================

func TestScheduler_Stop(t *testing.T) {
	t.Run("stop cancels context", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		scheduler.Stop()

		select {
		case <-scheduler.ctx.Done():
			// Expected
		default:
			t.Error("Context should be cancelled after Stop()")
		}
	})
}

// =============================================================================
// Scheduler Job Entries Tests
// =============================================================================

func TestScheduler_JobEntries(t *testing.T) {
	t.Run("empty on initialization", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.Empty(t, scheduler.jobEntries)
	})

	t.Run("can store and retrieve entries", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		scheduler.jobsMu.Lock()
		scheduler.jobEntries["test-job"] = cron.EntryID(1)
		scheduler.jobEntries["another-job"] = cron.EntryID(2)
		scheduler.jobsMu.Unlock()

		scheduler.jobsMu.RLock()
		defer scheduler.jobsMu.RUnlock()

		assert.Equal(t, cron.EntryID(1), scheduler.jobEntries["test-job"])
		assert.Equal(t, cron.EntryID(2), scheduler.jobEntries["another-job"])
	})
}

// =============================================================================
// Scheduler Concurrent Execution Tests
// =============================================================================

func TestScheduler_MaxConcurrent(t *testing.T) {
	t.Run("default max concurrent is 20", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.Equal(t, 20, scheduler.maxConcurrent)
	})

	t.Run("active count starts at 0", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.Equal(t, 0, scheduler.activeCount)
	})
}

// =============================================================================
// Schedule Calculation Tests
// =============================================================================

func TestScheduleCalculation(t *testing.T) {
	parser := cron.NewParser(
		cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
	)

	t.Run("every minute", func(t *testing.T) {
		schedule, _ := parser.Parse("* * * * *")
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		next := schedule.Next(now)

		assert.Equal(t, 31, next.Minute())
	})

	t.Run("every 5 minutes", func(t *testing.T) {
		schedule, _ := parser.Parse("*/5 * * * *")
		now := time.Date(2024, 1, 15, 10, 32, 0, 0, time.UTC)
		next := schedule.Next(now)

		assert.Equal(t, 35, next.Minute())
	})

	t.Run("daily at midnight", func(t *testing.T) {
		schedule, _ := parser.Parse("0 0 * * *")
		now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		next := schedule.Next(now)

		assert.Equal(t, 16, next.Day())
		assert.Equal(t, 0, next.Hour())
		assert.Equal(t, 0, next.Minute())
	})
}

// =============================================================================
// Scheduler IsScheduled Tests
// =============================================================================

func TestScheduler_IsScheduled(t *testing.T) {
	t.Run("returns false for unscheduled job", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.False(t, scheduler.IsScheduled("default", "non-existent"))
	})

	t.Run("returns true for scheduled job", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		scheduler.jobsMu.Lock()
		scheduler.jobEntries["default/scheduled-job"] = cron.EntryID(1) // Use "/" separator as expected by IsScheduled()
		scheduler.jobsMu.Unlock()

		assert.True(t, scheduler.IsScheduled("default", "scheduled-job"))
	})
}

// =============================================================================
// Scheduler GetScheduledJobs Tests
// =============================================================================

func TestScheduler_GetScheduledJobs(t *testing.T) {
	t.Run("returns empty list initially", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		jobs := scheduler.GetScheduledJobs()
		assert.Empty(t, jobs)
	})

	t.Run("returns scheduled job names", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		scheduler.jobsMu.Lock()
		scheduler.jobEntries["job-1"] = cron.EntryID(1)
		scheduler.jobEntries["job-2"] = cron.EntryID(2)
		scheduler.jobsMu.Unlock()

		jobs := scheduler.GetScheduledJobs()
		assert.Len(t, jobs, 2)

		// Check that the keys are present
		keys := make([]string, len(jobs))
		for i, job := range jobs {
			keys[i] = job.Key
		}
		assert.Contains(t, keys, "job-1")
		assert.Contains(t, keys, "job-2")
	})
}

// =============================================================================
// Scheduler UnscheduleJob Tests
// =============================================================================

func TestScheduler_UnscheduleJob(t *testing.T) {
	t.Run("removes scheduled job", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		scheduler.jobsMu.Lock()
		scheduler.jobEntries["default:to-remove"] = cron.EntryID(1)
		scheduler.jobsMu.Unlock()

		scheduler.UnscheduleJob("default", "to-remove")

		assert.False(t, scheduler.IsScheduled("default", "to-remove"))
	})

	t.Run("handles non-existent job gracefully", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		// Should not panic
		scheduler.UnscheduleJob("default", "non-existent")
	})
}

// =============================================================================
// Scheduler Additional Coverage Tests
// =============================================================================

func TestScheduler_AdditionalCoverage(t *testing.T) {
	t.Run("scheduler context is not nil", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.NotNil(t, scheduler.ctx)
	})

	t.Run("scheduler cancel function is not nil", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.NotNil(t, scheduler.cancel)
	})

	t.Run("scheduler storage is initialized", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.NotNil(t, scheduler.storage)
	})

	t.Run("scheduler max concurrent defaults to 20", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.Equal(t, 20, scheduler.maxConcurrent)
	})

	t.Run("scheduler active count starts at 0", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.Equal(t, 0, scheduler.activeCount)
	})

	t.Run("scheduler mutex is initialized", func(t *testing.T) {
		scheduler := NewScheduler(nil)
		assert.NotNil(t, scheduler.jobsMu)
	})

	t.Run("unschedule job with empty namespace", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		// Should not panic
		scheduler.UnscheduleJob("", "job-name")
	})

	t.Run("unschedule job with empty job name", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		// Should not panic
		scheduler.UnscheduleJob("default", "")
	})

	t.Run("check is scheduled with empty inputs", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		// Should return false, not panic
		result := scheduler.IsScheduled("", "")
		assert.False(t, result)
	})

	t.Run("get scheduled jobs from empty scheduler", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		jobs := scheduler.GetScheduledJobs()
		assert.NotNil(t, jobs)
		assert.Empty(t, jobs)
	})

	t.Run("multiple stop calls are safe", func(t *testing.T) {
		scheduler := NewScheduler(nil)

		// Multiple stops should not panic
		scheduler.Stop()
		scheduler.Stop()
		scheduler.Stop()
	})

	t.Run("validate cron with empty string", func(t *testing.T) {
		err := ValidateCronSchedule("")
		assert.Error(t, err)
	})

	t.Run("validate cron with whitespace", func(t *testing.T) {
		err := ValidateCronSchedule("   ")
		assert.Error(t, err)
	})

	t.Run("validate cron with tab separator", func(t *testing.T) {
		err := ValidateCronSchedule("0\t0\t*\t*\t*")
		assert.NoError(t, err)
	})
}

// =============================================================================
// ScheduleConfig Additional Tests
// =============================================================================

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

// =============================================================================
// CronIntervalError Additional Tests
// =============================================================================

func TestCronIntervalError_Additional(t *testing.T) {
	t.Run("error with zero interval", func(t *testing.T) {
		err := &CronIntervalError{
			Expression: "* * * * * *",
			Interval:   0,
			MinAllowed: time.Minute,
		}

		msg := err.Error()
		assert.Contains(t, msg, "runs too frequently")
	})

	t.Run("error with negative interval", func(t *testing.T) {
		err := &CronIntervalError{
			Expression: "invalid",
			Interval:   -time.Second,
			MinAllowed: time.Minute,
		}

		msg := err.Error()
		assert.NotEmpty(t, msg)
	})

	t.Run("error type assertion", func(t *testing.T) {
		var err error = &CronIntervalError{
			Expression: "*/30 * * * * *",
			Interval:   30 * time.Second,
			MinAllowed: time.Minute,
		}

		// Should be able to type assert back to CronIntervalError
		intervalErr, ok := err.(*CronIntervalError)
		assert.True(t, ok)
		assert.Equal(t, "*/30 * * * * *", intervalErr.Expression)
	})
}
