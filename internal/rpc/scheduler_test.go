package rpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/internal/scheduler"
)

func TestNewScheduler(t *testing.T) {
	t.Run("creates scheduler with nil dependencies", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		require.NotNil(t, s)
		assert.Nil(t, s.storage)
		assert.Nil(t, s.executor)
		assert.NotNil(t, s.inner)
	})

	t.Run("initializes inner CronScheduler", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		assert.NotNil(t, s.inner)
		assert.NotNil(t, s.inner.Guard)
		assert.Equal(t, 10, s.inner.Guard.MaxConcurrent)
	})
}

func TestScheduler_ScheduleProcedure(t *testing.T) {
	t.Run("returns nil for nil schedule", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  nil,
		}

		err := s.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("returns nil for empty schedule", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		emptySchedule := ""
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &emptySchedule,
		}

		err := s.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("schedules valid cron expression", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := s.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("schedules 6-field cron with seconds", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "0 */5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := s.ScheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("returns error for invalid cron expression", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		schedule := "invalid cron"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		err := s.ScheduleProcedure(proc)

		assert.Error(t, err)
		assert.False(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("replaces existing schedule", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule1 := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule1,
		}

		err := s.ScheduleProcedure(proc)
		require.NoError(t, err)

		s.inner.GetScheduledEntries()

		schedule2 := "*/10 * * * *"
		proc.Schedule = &schedule2

		err = s.ScheduleProcedure(proc)
		require.NoError(t, err)

		assert.True(t, s.IsScheduled("public", "test_proc"))
	})
}

func TestScheduler_UnscheduleProcedure(t *testing.T) {
	t.Run("unschedules existing procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		_ = s.ScheduleProcedure(proc)
		assert.True(t, s.IsScheduled("public", "test_proc"))

		s.UnscheduleProcedure("public", "test_proc")

		assert.False(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("handles unscheduling non-existent procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		s.UnscheduleProcedure("public", "non_existent")

		assert.False(t, s.IsScheduled("public", "non_existent"))
	})
}

func TestScheduler_RescheduleProcedure(t *testing.T) {
	t.Run("reschedules enabled procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = s.ScheduleProcedure(proc)

		newSchedule := "*/10 * * * *"
		proc.Schedule = &newSchedule

		err := s.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.True(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("removes schedule for disabled procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = s.ScheduleProcedure(proc)
		assert.True(t, s.IsScheduled("public", "test_proc"))

		proc.Enabled = false

		err := s.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, s.IsScheduled("public", "test_proc"))
	})

	t.Run("removes schedule when schedule is nil", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
			Enabled:   true,
		}

		_ = s.ScheduleProcedure(proc)
		assert.True(t, s.IsScheduled("public", "test_proc"))

		proc.Schedule = nil

		err := s.RescheduleProcedure(proc)

		assert.NoError(t, err)
		assert.False(t, s.IsScheduled("public", "test_proc"))
	})
}

func TestScheduler_IsScheduled(t *testing.T) {
	t.Run("returns false for non-existent procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		result := s.IsScheduled("public", "non_existent")

		assert.False(t, result)
	})

	t.Run("returns true for scheduled procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "public",
			Schedule:  &schedule,
		}

		_ = s.ScheduleProcedure(proc)

		result := s.IsScheduled("public", "test_proc")

		assert.True(t, result)
	})

	t.Run("handles different namespaces", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc := &Procedure{
			Name:      "test_proc",
			Namespace: "namespace1",
			Schedule:  &schedule,
		}

		_ = s.ScheduleProcedure(proc)

		assert.True(t, s.IsScheduled("namespace1", "test_proc"))
		assert.False(t, s.IsScheduled("namespace2", "test_proc"))
	})
}

func TestScheduler_GetScheduledProcedures(t *testing.T) {
	t.Run("returns empty slice when no procedures scheduled", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		procs := s.GetScheduledProcedures()

		assert.Empty(t, procs)
	})

	t.Run("returns all scheduled procedures", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		proc1 := &Procedure{Name: "proc1", Namespace: "public", Schedule: &schedule}
		proc2 := &Procedure{Name: "proc2", Namespace: "public", Schedule: &schedule}

		_ = s.ScheduleProcedure(proc1)
		_ = s.ScheduleProcedure(proc2)

		procs := s.GetScheduledProcedures()

		assert.Len(t, procs, 2)
	})

	t.Run("includes next run time", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/1 * * * *"
		proc := &Procedure{Name: "test_proc", Namespace: "public", Schedule: &schedule}

		_ = s.ScheduleProcedure(proc)

		procs := s.GetScheduledProcedures()

		require.Len(t, procs, 1)
		assert.Equal(t, "public/test_proc", procs[0].Key)
	})
}

func TestScheduler_GetScheduleInfo(t *testing.T) {
	t.Run("returns false for unscheduled procedure", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		_, ok := s.GetScheduleInfo("public", "non_existent")

		assert.False(t, ok)
	})
}

func TestScheduledEntryInfo(t *testing.T) {
	t.Run("stores all fields", func(t *testing.T) {
		now := time.Now()
		nextRun := now.Add(1 * time.Hour)
		prevRun := now.Add(-1 * time.Hour)

		info := scheduler.ScheduledEntryInfo{
			Key:     "namespace/proc_name",
			EntryID: 123,
			NextRun: nextRun,
			PrevRun: prevRun,
		}

		assert.Equal(t, "namespace/proc_name", info.Key)
		assert.Equal(t, 123, info.EntryID)
		assert.Equal(t, nextRun, info.NextRun)
		assert.Equal(t, prevRun, info.PrevRun)
	})
}

func TestScheduler_Stop(t *testing.T) {
	t.Run("stops scheduler gracefully", func(t *testing.T) {
		s := NewScheduler(nil, nil)

		schedule := "*/5 * * * *"
		proc := &Procedure{Name: "test_proc", Namespace: "public", Schedule: &schedule}
		_ = s.ScheduleProcedure(proc)

		s.Stop()

		select {
		case <-s.inner.Context().Done():
		default:
			t.Error("context should be cancelled after stop")
		}
	})
}

func TestScheduler_ConcurrentAccess(t *testing.T) {
	t.Run("handles concurrent schedule/unschedule", func(t *testing.T) {
		s := NewScheduler(nil, nil)
		defer s.Stop()

		schedule := "*/5 * * * *"
		done := make(chan bool)

		for i := 0; i < 10; i++ {
			go func(idx int) {
				proc := &Procedure{
					Name:      "test_proc",
					Namespace: "public",
					Schedule:  &schedule,
				}
				_ = s.ScheduleProcedure(proc)
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}

		assert.True(t, s.IsScheduled("public", "test_proc"))
	})
}

func TestScheduler_CronExpressions(t *testing.T) {
	s := NewScheduler(nil, nil)
	defer s.Stop()

	testCases := []struct {
		name     string
		schedule string
		valid    bool
	}{
		{"every minute", "* * * * *", true},
		{"every 5 minutes", "*/5 * * * *", true},
		{"every hour", "0 * * * *", true},
		{"every day at midnight", "0 0 * * *", true},
		{"with seconds", "0 */5 * * * *", true},
		{"every monday", "0 0 * * MON", true},
		{"@hourly descriptor", "@hourly", true},
		{"@daily descriptor", "@daily", true},
		{"@weekly descriptor", "@weekly", true},
		{"invalid expression", "invalid", false},
		{"too few fields", "* * *", false},
		{"invalid minute", "60 * * * *", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proc := &Procedure{
				Name:      "test_proc_" + tc.name,
				Namespace: "public",
				Schedule:  &tc.schedule,
			}

			err := s.ScheduleProcedure(proc)

			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			s.UnscheduleProcedure("public", proc.Name)
		})
	}
}

func BenchmarkScheduler_ScheduleProcedure(b *testing.B) {
	s := NewScheduler(nil, nil)
	defer s.Stop()

	schedule := "*/5 * * * *"
	proc := &Procedure{
		Name:      "bench_proc",
		Namespace: "public",
		Schedule:  &schedule,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ScheduleProcedure(proc)
	}
}

func BenchmarkScheduler_IsScheduled(b *testing.B) {
	s := NewScheduler(nil, nil)
	defer s.Stop()

	schedule := "*/5 * * * *"
	proc := &Procedure{
		Name:      "bench_proc",
		Namespace: "public",
		Schedule:  &schedule,
	}
	_ = s.ScheduleProcedure(proc)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.IsScheduled("public", "bench_proc")
	}
}
