package scheduler

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCronScheduler_WithPositiveConcurrency_CreatesScheduler(t *testing.T) {
	s := NewCronScheduler(5)
	require.NotNil(t, s)
	assert.NotNil(t, s.cron)
	assert.NotNil(t, s.entries)
	assert.NotNil(t, s.ctx)
	assert.NotNil(t, s.Guard)
	assert.Equal(t, 5, s.Guard.MaxConcurrent)
	assert.Equal(t, 0, s.EntryCount())
}

func TestNewCronScheduler_WithZeroConcurrency_CreatesScheduler(t *testing.T) {
	s := NewCronScheduler(0)
	require.NotNil(t, s)
	assert.Equal(t, 0, s.Guard.MaxConcurrent)
}

func TestNewCronScheduler_ContextIsNotCancelled(t *testing.T) {
	s := NewCronScheduler(1)
	assert.NoError(t, s.Context().Err())
}

func TestAddFunc_ValidFiveFieldSpec_ReturnsEntryID(t *testing.T) {
	s := NewCronScheduler(1)
	id, err := s.AddFunc("test-job", "* * * * *", func() {})
	require.NoError(t, err)
	assert.True(t, id > 0)
	assert.True(t, s.IsScheduled("test-job"))
}

func TestAddFunc_ValidSixFieldSpecWithSeconds_ReturnsEntryID(t *testing.T) {
	s := NewCronScheduler(1)
	id, err := s.AddFunc("sec-job", "*/10 * * * * *", func() {})
	require.NoError(t, err)
	assert.True(t, id > 0)
}

func TestAddFunc_ValidDescriptorSpec_ReturnsEntryID(t *testing.T) {
	s := NewCronScheduler(1)
	id, err := s.AddFunc("hourly-job", "@hourly", func() {})
	require.NoError(t, err)
	assert.True(t, id > 0)
}

func TestAddFunc_InvalidCronSpec_ReturnsError(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("bad-job", "not a cron", func() {})
	assert.Error(t, err)
	assert.False(t, s.IsScheduled("bad-job"))
}

func TestAddFunc_EmptySpec_ReturnsError(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("empty-spec", "", func() {})
	assert.Error(t, err)
}

func TestAddFunc_DuplicateKey_ReplacesOldEntry(t *testing.T) {
	s := NewCronScheduler(1)
	id1, err := s.AddFunc("dup-key", "* * * * *", func() {})
	require.NoError(t, err)

	id2, err := s.AddFunc("dup-key", "0 * * * *", func() {})
	require.NoError(t, err)

	assert.NotEqual(t, id1, id2)
	assert.True(t, s.IsScheduled("dup-key"))
	assert.Equal(t, 1, s.EntryCount())
}

func TestAddFunc_NilFunction_RegistersWithoutError(t *testing.T) {
	s := NewCronScheduler(1)
	id, err := s.AddFunc("nil-fn", "* * * * *", nil)
	require.NoError(t, err)
	assert.True(t, id > 0)
	assert.True(t, s.IsScheduled("nil-fn"))
}

func TestAddFunc_MultipleDistinctKeys_AllRegistered(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("a", "* * * * *", func() {})
	require.NoError(t, err)
	_, err = s.AddFunc("b", "0 * * * *", func() {})
	require.NoError(t, err)
	_, err = s.AddFunc("c", "@daily", func() {})
	require.NoError(t, err)

	assert.Equal(t, 3, s.EntryCount())
	assert.True(t, s.IsScheduled("a"))
	assert.True(t, s.IsScheduled("b"))
	assert.True(t, s.IsScheduled("c"))
}

func TestStart_Stop_CompletesCleanly(t *testing.T) {
	s := NewCronScheduler(1)
	loader := func(ctx context.Context) ([]Schedulable, error) {
		return []Schedulable{}, nil
	}
	err := s.Start(loader, "test")
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	s.Stop("test")
}

func TestStop_WithoutStart_DoesNotPanic(t *testing.T) {
	s := NewCronScheduler(1)
	assert.NotPanics(t, func() {
		s.Stop("test")
	})
}

func TestStop_CancelsContext(t *testing.T) {
	s := NewCronScheduler(1)
	s.cron.Start()
	s.Stop("test")

	assert.Error(t, s.Context().Err())
}

func TestStart_DoubleStart_DoesNotPanic(t *testing.T) {
	s := NewCronScheduler(1)
	loader := func(ctx context.Context) ([]Schedulable, error) {
		return nil, nil
	}
	require.NoError(t, s.Start(loader, "test"))
	require.NoError(t, s.Start(loader, "test"))

	time.Sleep(200 * time.Millisecond)
	s.Stop("test")
}

func TestJobExecution_EverySecond_FunctionIsCalled(t *testing.T) {
	s := NewCronScheduler(1)
	var called atomic.Int32

	_, err := s.AddFunc("exec-test", "* * * * * *", func() {
		called.Add(1)
	})
	require.NoError(t, err)

	s.cron.Start()
	defer s.Stop("test")

	waitFor(t, 2*time.Second, func() bool { return called.Load() > 0 })
	assert.True(t, called.Load() > 0, "expected function to be called at least once, got %d", called.Load())
}

func TestJobExecution_StopPreventsFurtherExecutions(t *testing.T) {
	s := NewCronScheduler(1)
	var called atomic.Int32

	_, err := s.AddFunc("stop-test", "* * * * * *", func() {
		called.Add(1)
	})
	require.NoError(t, err)

	s.cron.Start()

	waitFor(t, 2*time.Second, func() bool { return called.Load() > 0 })
	require.True(t, called.Load() > 0, "expected at least one execution before stop")

	s.Stop("test")
	countAfterStop := called.Load()

	time.Sleep(1500 * time.Millisecond)
	assert.Equal(t, countAfterStop, called.Load(), "no executions should happen after stop")
}

func TestJobExecution_MultipleJobs_AllExecute(t *testing.T) {
	s := NewCronScheduler(2)
	var count1, count2 atomic.Int32

	_, err := s.AddFunc("job1", "* * * * * *", func() { count1.Add(1) })
	require.NoError(t, err)
	_, err = s.AddFunc("job2", "* * * * * *", func() { count2.Add(1) })
	require.NoError(t, err)

	s.cron.Start()
	defer s.Stop("test")

	waitFor(t, 2*time.Second, func() bool {
		return count1.Load() > 0 && count2.Load() > 0
	})

	assert.True(t, count1.Load() > 0, "job1 should have executed")
	assert.True(t, count2.Load() > 0, "job2 should have executed")
}

func TestGetScheduledEntries_WithJobs_ReturnsAll(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("job1", "* * * * *", func() {})
	require.NoError(t, err)
	_, err = s.AddFunc("job2", "0 * * * *", func() {})
	require.NoError(t, err)

	entries := s.GetScheduledEntries()
	assert.Len(t, entries, 2)

	keys := make(map[string]bool)
	for _, e := range entries {
		keys[e.Key] = true
		assert.True(t, e.EntryID > 0)
	}
	assert.True(t, keys["job1"])
	assert.True(t, keys["job2"])
}

func TestGetScheduledEntries_EmptyScheduler_ReturnsEmptyList(t *testing.T) {
	s := NewCronScheduler(1)
	entries := s.GetScheduledEntries()
	assert.Empty(t, entries)
}

func TestIsScheduled_ExistingKey_ReturnsTrue(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("exists", "* * * * *", func() {})
	require.NoError(t, err)
	assert.True(t, s.IsScheduled("exists"))
}

func TestIsScheduled_NonExistingKey_ReturnsFalse(t *testing.T) {
	s := NewCronScheduler(1)
	assert.False(t, s.IsScheduled("nope"))
}

func TestIsScheduled_AfterRemove_ReturnsFalse(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("temp", "* * * * *", func() {})
	require.NoError(t, err)
	s.Remove("temp")
	assert.False(t, s.IsScheduled("temp"))
}

func TestGetScheduleInfo_ExistingKey_ReturnsInfo(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("info-test", "* * * * *", func() {})
	require.NoError(t, err)

	info, ok := s.GetScheduleInfo("info-test")
	assert.True(t, ok)
	assert.NotEmpty(t, info)
}

func TestGetScheduleInfo_NonExistingKey_ReturnsEmptyAndFalse(t *testing.T) {
	s := NewCronScheduler(1)
	info, ok := s.GetScheduleInfo("nope")
	assert.False(t, ok)
	assert.Empty(t, info)
}

func TestRemove_ExistingKey_RemovesEntry(t *testing.T) {
	s := NewCronScheduler(1)
	_, err := s.AddFunc("remove-me", "* * * * *", func() {})
	require.NoError(t, err)
	assert.True(t, s.IsScheduled("remove-me"))

	s.Remove("remove-me")
	assert.False(t, s.IsScheduled("remove-me"))
	assert.Equal(t, 0, s.EntryCount())
}

func TestRemove_NonExistingKey_DoesNotPanic(t *testing.T) {
	s := NewCronScheduler(1)
	assert.NotPanics(t, func() {
		s.Remove("nope")
	})
}

func TestEntryCount_TracksAddAndRemove(t *testing.T) {
	s := NewCronScheduler(1)
	assert.Equal(t, 0, s.EntryCount())

	s.AddFunc("a", "* * * * *", func() {})
	assert.Equal(t, 1, s.EntryCount())

	s.AddFunc("b", "* * * * *", func() {})
	assert.Equal(t, 2, s.EntryCount())

	s.AddFunc("a", "0 * * * *", func() {})
	assert.Equal(t, 2, s.EntryCount())

	s.Remove("b")
	assert.Equal(t, 1, s.EntryCount())

	s.Remove("a")
	assert.Equal(t, 0, s.EntryCount())
}

func TestConcurrencyGuard_AcquireWithinLimit_ReturnsTrue(t *testing.T) {
	g := &ConcurrencyGuard{MaxConcurrent: 3}
	assert.True(t, g.Acquire("t1"))
	assert.True(t, g.Acquire("t2"))
	assert.True(t, g.Acquire("t3"))
}

func TestConcurrencyGuard_AcquireExceedsLimit_ReturnsFalse(t *testing.T) {
	g := &ConcurrencyGuard{MaxConcurrent: 2}
	require.True(t, g.Acquire("t1"))
	require.True(t, g.Acquire("t2"))
	assert.False(t, g.Acquire("t3"))
}

func TestConcurrencyGuard_Release_DecrementsActiveCount(t *testing.T) {
	g := &ConcurrencyGuard{MaxConcurrent: 1}
	require.True(t, g.Acquire("t1"))
	assert.False(t, g.Acquire("t2"))

	g.Release()
	assert.True(t, g.Acquire("t3"))
}

func TestConcurrencyGuard_ZeroMax_AlwaysFailsAcquire(t *testing.T) {
	g := &ConcurrencyGuard{MaxConcurrent: 0}
	assert.False(t, g.Acquire("t1"))
}

func TestConcurrencyGuard_ConcurrentAcquireRelease(t *testing.T) {
	g := &ConcurrencyGuard{MaxConcurrent: 5}
	var wg sync.WaitGroup
	var acquired atomic.Int32

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if g.Acquire("worker") {
				acquired.Add(1)
				time.Sleep(time.Millisecond)
				g.Release()
			}
		}()
	}
	wg.Wait()

	assert.True(t, acquired.Load() > 0, "at least some goroutines should acquire")
}

func TestValidateCronSchedule_ValidMinuteSpec_NoError(t *testing.T) {
	assert.NoError(t, ValidateCronSchedule("* * * * *"))
}

func TestValidateCronSchedule_ValidSpecificTime_NoError(t *testing.T) {
	assert.NoError(t, ValidateCronSchedule("30 9 * * 1-5"))
}

func TestValidateCronSchedule_InvalidSpec_ReturnsError(t *testing.T) {
	assert.Error(t, ValidateCronSchedule("bad"))
}

func TestValidateCronSchedule_EverySecond_TooFrequent(t *testing.T) {
	err := ValidateCronSchedule("* * * * * *")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "too frequently")
}

func TestValidateCronSchedule_Every500ms_TooFrequent(t *testing.T) {
	err := ValidateCronSchedule("@every 500ms")
	assert.Error(t, err)
}

func TestValidateCronSchedule_EveryMinute_Passes(t *testing.T) {
	assert.NoError(t, ValidateCronSchedule("@every 1m"))
}

func TestStart_WithLoader_RegistersEnabledItemsOnly(t *testing.T) {
	s := NewCronScheduler(1)
	loader := func(ctx context.Context) ([]Schedulable, error) {
		return []Schedulable{
			{Key: "enabled", Schedule: "* * * * *", Enabled: true},
			{Key: "disabled", Schedule: "* * * * *", Enabled: false},
			{Key: "no-schedule", Schedule: "", Enabled: true},
		}, nil
	}

	err := s.Start(loader, "test")
	require.NoError(t, err)

	waitFor(t, 1*time.Second, func() bool { return s.IsScheduled("enabled") })

	assert.True(t, s.IsScheduled("enabled"))
	assert.False(t, s.IsScheduled("disabled"))
	assert.False(t, s.IsScheduled("no-schedule"))

	s.Stop("test")
}

func TestStart_LoaderFailure_DoesNotRegisterAnything(t *testing.T) {
	s := NewCronScheduler(1)
	loader := func(ctx context.Context) ([]Schedulable, error) {
		return nil, fmt.Errorf("db down")
	}

	err := s.Start(loader, "test")
	require.NoError(t, err)

	time.Sleep(3 * time.Second)
	assert.Equal(t, 0, s.EntryCount())

	s.Stop("test")
}

func TestStart_LoaderPanic_RecoversAndDoesNotRegister(t *testing.T) {
	s := NewCronScheduler(1)
	loader := func(ctx context.Context) ([]Schedulable, error) {
		panic("boom")
	}

	err := s.Start(loader, "test")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 0, s.EntryCount())

	s.Stop("test")
}

func TestAddFunc_ConcurrentWhileRunning_AllSucceed(t *testing.T) {
	s := NewCronScheduler(10)
	s.cron.Start()
	defer s.Stop("test")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("concurrent-%d", idx)
			_, err := s.AddFunc(key, "* * * * *", func() {})
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 50, s.EntryCount())
}

func TestAddFunc_ConcurrentDuplicateKeys_LastWins(t *testing.T) {
	s := NewCronScheduler(1)

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.AddFunc("same-key", "* * * * *", func() {})
		}()
	}
	wg.Wait()

	assert.True(t, s.IsScheduled("same-key"))
	assert.Equal(t, 1, s.EntryCount())
}

func waitFor(t *testing.T, timeout time.Duration, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}
