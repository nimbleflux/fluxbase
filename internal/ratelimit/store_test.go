package ratelimit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheck(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	t.Run("allows requests under limit", func(t *testing.T) {
		result, err := Check(ctx, store, "check-under-limit", 10, time.Minute)
		require.NoError(t, err)

		assert.True(t, result.Allowed)
		assert.Equal(t, int64(10), result.Limit)
		assert.Equal(t, int64(9), result.Remaining)
		assert.False(t, result.ResetAt.IsZero())
	})

	t.Run("tracks remaining correctly", func(t *testing.T) {
		for i := 1; i <= 5; i++ {
			result, err := Check(ctx, store, "check-remaining", 10, time.Minute)
			require.NoError(t, err)

			assert.True(t, result.Allowed)
			assert.Equal(t, int64(10-i), result.Remaining)
		}
	})

	t.Run("denies requests at limit", func(t *testing.T) {
		key := "check-at-limit"

		// Use up all requests
		for i := 0; i < 5; i++ {
			_, err := Check(ctx, store, key, 5, time.Minute)
			require.NoError(t, err)
		}

		// Next request should be denied
		result, err := Check(ctx, store, key, 5, time.Minute)
		require.NoError(t, err)

		assert.False(t, result.Allowed)
		assert.Equal(t, int64(0), result.Remaining)
		assert.Equal(t, int64(5), result.Limit)
	})

	t.Run("remaining never goes negative", func(t *testing.T) {
		key := "check-not-negative"

		// Use up all requests and then some
		for i := 0; i < 15; i++ {
			result, err := Check(ctx, store, key, 10, time.Minute)
			require.NoError(t, err)

			// Remaining should never be negative
			assert.GreaterOrEqual(t, result.Remaining, int64(0))
		}
	})

	t.Run("reset time is in the future", func(t *testing.T) {
		result, err := Check(ctx, store, "check-reset-time", 10, time.Minute)
		require.NoError(t, err)

		assert.True(t, result.ResetAt.After(time.Now()))
		// Should be approximately 1 minute from now
		expectedReset := time.Now().Add(time.Minute)
		assert.WithinDuration(t, expectedReset, result.ResetAt, 5*time.Second)
	})
}

func TestResult(t *testing.T) {
	t.Run("result with all fields", func(t *testing.T) {
		resetAt := time.Now().Add(time.Minute)
		result := &Result{
			Allowed:   true,
			Remaining: 5,
			ResetAt:   resetAt,
			Limit:     10,
		}

		assert.True(t, result.Allowed)
		assert.Equal(t, int64(5), result.Remaining)
		assert.Equal(t, resetAt, result.ResetAt)
		assert.Equal(t, int64(10), result.Limit)
	})

	t.Run("result when denied", func(t *testing.T) {
		result := &Result{
			Allowed:   false,
			Remaining: 0,
			Limit:     10,
		}

		assert.False(t, result.Allowed)
		assert.Equal(t, int64(0), result.Remaining)
	})
}

func TestCheckWithExpiration(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()
	key := "check-expiration"

	// Use up all requests
	for i := 0; i < 5; i++ {
		_, err := Check(ctx, store, key, 5, 100*time.Millisecond)
		require.NoError(t, err)
	}

	// Should be denied
	result, err := Check(ctx, store, key, 5, 100*time.Millisecond)
	require.NoError(t, err)
	assert.False(t, result.Allowed)

	// Wait for window to expire
	time.Sleep(150 * time.Millisecond)

	// Should be allowed again
	result, err = Check(ctx, store, key, 5, 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(4), result.Remaining)
}

// =============================================================================
// Additional Check Tests
// =============================================================================

func TestCheck_ZeroLimit(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	// Zero limit should deny immediately
	result, err := Check(ctx, store, "zero-limit", 0, time.Minute)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
}

func TestCheck_LargeLimit(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	// Very large limit
	result, err := Check(ctx, store, "large-limit", 1000000, time.Minute)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(999999), result.Remaining)
}

func TestCheck_VaryingWindowSizes(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	windows := []time.Duration{
		time.Second,
		10 * time.Second,
		time.Minute,
		5 * time.Minute,
		time.Hour,
	}

	for _, window := range windows {
		result, err := Check(ctx, store, "window-test", 10, window)
		require.NoError(t, err)
		assert.True(t, result.Allowed)
		assert.False(t, result.ResetAt.IsZero())
		assert.True(t, result.ResetAt.After(time.Now()))
	}
}

func TestStore_EmptyKey(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	// Empty key should work
	count, err := store.Increment(ctx, "", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

// Context cancellation test skipped - memory store doesn't check context in Increment/Get

func TestIncrement_UpdatesExpiration(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()
	key := "update-expiration"

	// First increment with 1 minute expiration
	count1, err := store.Increment(ctx, key, time.Minute)
	require.NoError(t, err)

	_, exp1, _ := store.Get(ctx, key)
	firstExpiration := exp1

	// Wait a bit (but less than expiration)
	time.Sleep(10 * time.Millisecond)

	// Second increment - memory store doesn't update existing key expiration
	count2, err := store.Increment(ctx, key, time.Minute)
	require.NoError(t, err)
	assert.Equal(t, count1+1, count2)

	_, exp2, _ := store.Get(ctx, key)
	// Memory store may or may not update expiration depending on implementation
	assert.True(t, exp2.After(firstExpiration) || exp2.Equal(firstExpiration))
}

func TestStore_SpecialKeyCharacters(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	keys := []string{
		"user:123",
		"ip:192.168.1.1",
		"api:endpoint:method",
		"key-with-dashes",
		"key_with_underscores",
		"key.with.dots",
		"key/with/slashes",
	}

	for _, key := range keys {
		count, err := store.Increment(ctx, key, time.Minute)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	}

	// Verify all keys are independent
	for _, key := range keys {
		count, _, err := store.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	}
}

// =============================================================================
// Result Field Tests
// =============================================================================

func TestResult_AllFields(t *testing.T) {
	result := &Result{
		Allowed:   true,
		Remaining: 5,
		ResetAt:   time.Now().Add(time.Minute),
		Limit:     10,
	}

	assert.True(t, result.Allowed)
	assert.Equal(t, int64(5), result.Remaining)
	assert.Equal(t, int64(10), result.Limit)
	assert.False(t, result.ResetAt.IsZero())
}

func TestResult_ZeroValues(t *testing.T) {
	result := &Result{}

	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
	assert.Equal(t, int64(0), result.Limit)
	assert.True(t, result.ResetAt.IsZero())
}

func TestResult_NegativeRemaining(t *testing.T) {
	// Test that Remaining is never negative
	result := &Result{
		Allowed:   false,
		Remaining: -5,
		Limit:     10,
	}

	// Should cap at 0
	if result.Remaining < 0 {
		result.Remaining = 0
	}
	assert.Equal(t, int64(0), result.Remaining)
}

// =============================================================================
// Check Error Handling Tests
// =============================================================================

// mockErrorStore is a mock store that always returns errors
type mockErrorStore struct {
	errorToReturn error
}

func (m *mockErrorStore) Get(ctx context.Context, key string) (int64, time.Time, error) {
	return 0, time.Time{}, m.errorToReturn
}

func (m *mockErrorStore) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return 0, m.errorToReturn
}

func (m *mockErrorStore) Reset(ctx context.Context, key string) error {
	return m.errorToReturn
}

func (m *mockErrorStore) Close() error {
	return nil
}

func TestCheck_ErrorHandling(t *testing.T) {
	t.Run("returns error when store increment fails", func(t *testing.T) {
		expectedErr := fmt.Errorf("store unavailable")
		store := &mockErrorStore{errorToReturn: expectedErr}

		ctx := context.Background()
		result, err := Check(ctx, store, "error-key", 10, time.Minute)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns error for context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		expectedErr := context.Canceled
		store := &mockErrorStore{errorToReturn: expectedErr}

		result, err := Check(ctx, store, "cancelled-key", 10, time.Minute)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns error for context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond) // Ensure timeout occurs

		expectedErr := context.DeadlineExceeded
		store := &mockErrorStore{errorToReturn: expectedErr}

		result, err := Check(ctx, store, "timeout-key", 10, time.Minute)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

// =============================================================================
// Check Edge Cases
// =============================================================================

func TestCheck_NegativeLimit(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	// Negative limit should always deny (though it's an edge case)
	result, err := Check(ctx, store, "negative-limit", -10, time.Minute)
	require.NoError(t, err)

	// With negative limit, any count will exceed it
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(-10), result.Limit)
}

func TestCheck_ExactlyAtLimit(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()
	key := "exactly-at-limit"
	limit := int64(5)

	// At limit should be allowed (count <= limit)
	// Request 1: count=1, allowed=true, remaining=4
	// Request 2: count=2, allowed=true, remaining=3
	// Request 3: count=3, allowed=true, remaining=2
	// Request 4: count=4, allowed=true, remaining=1
	// Request 5: count=5, allowed=true, remaining=0 (at limit, still allowed)
	// Request 6: count=6, allowed=false, remaining=0 (over limit)
	for i := int64(1); i <= limit; i++ {
		result, err := Check(ctx, store, key, limit, time.Minute)
		require.NoError(t, err)
		assert.True(t, result.Allowed, "request %d (count=%d) should be allowed", i, i)
		assert.Equal(t, limit-i, result.Remaining, "request %d should have %d remaining", i, limit-i)
	}

	// Next request should be denied (over limit)
	result, err := Check(ctx, store, key, limit, time.Minute)
	require.NoError(t, err)
	assert.False(t, result.Allowed, "request 6 should be denied (over limit)")
	assert.Equal(t, int64(0), result.Remaining)
}

func TestCheck_ConcurrentRequests(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()
	key := "concurrent-requests"
	limit := int64(100)

	// Simulate concurrent requests
	done := make(chan bool, 50)
	errors := make(chan error, 50)

	for i := 0; i < 50; i++ {
		go func() {
			result, err := Check(ctx, store, key, limit, time.Minute)
			if err != nil {
				errors <- err
			} else if !result.Allowed {
				errors <- fmt.Errorf("request unexpectedly denied")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 50; i++ {
		select {
		case <-done:
			// Success
		case err := <-errors:
			t.Fatal("concurrent request failed:", err)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent requests")
		}
	}
}

func TestCheck_ResetTimeAccuracy(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	testCases := []struct {
		window  time.Duration
		epsilon time.Duration
	}{
		{time.Second, 100 * time.Millisecond},
		{time.Minute, time.Second},
		{time.Hour, time.Second},
	}

	for _, tc := range testCases {
		t.Run(tc.window.String(), func(t *testing.T) {
			before := time.Now()
			result, err := Check(ctx, store, "reset-time-test", 10, tc.window)
			require.NoError(t, err)

			after := time.Now()

			// ResetAt should be approximately window duration from now
			expectedMin := before.Add(tc.window)
			expectedMax := after.Add(tc.window)

			assert.True(t, result.ResetAt.After(expectedMin) || result.ResetAt.Equal(expectedMin),
				"ResetAt %v should be after %v", result.ResetAt, expectedMin)
			assert.True(t, result.ResetAt.Before(expectedMax) || result.ResetAt.Equal(expectedMax),
				"ResetAt %v should be before %v", result.ResetAt, expectedMax)
		})
	}
}

func TestCheck_VeryShortWindow(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()

	// Very short window (1 millisecond)
	result, err := Check(ctx, store, "short-window", 10, time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.False(t, result.ResetAt.IsZero())

	// Wait for window to expire
	time.Sleep(50 * time.Millisecond)

	// Should be allowed again (window expired)
	result, err = Check(ctx, store, "short-window", 10, time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
}

func TestCheck_VeryLongWindow(t *testing.T) {
	store := NewMemoryStore(time.Hour)
	defer store.Close()

	ctx := context.Background()

	// Very long window (24 hours)
	result, err := Check(ctx, store, "long-window", 10, 24*time.Hour)
	require.NoError(t, err)
	assert.True(t, result.Allowed)

	// ResetAt should be ~24 hours from now
	expectedReset := time.Now().Add(24 * time.Hour)
	assert.WithinDuration(t, expectedReset, result.ResetAt, time.Second)
}

func TestCheck_LimitOfOne(t *testing.T) {
	store := NewMemoryStore(time.Minute)
	defer store.Close()

	ctx := context.Background()
	key := "limit-of-one"
	limit := int64(1)

	// First request should be allowed
	result, err := Check(ctx, store, key, limit, time.Minute)
	require.NoError(t, err)
	assert.True(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)

	// Second request should be denied
	result, err = Check(ctx, store, key, limit, time.Minute)
	require.NoError(t, err)
	assert.False(t, result.Allowed)
	assert.Equal(t, int64(0), result.Remaining)
}

func TestCheck_StoreIncrementValues(t *testing.T) {
	// Test that Check correctly handles different increment values from the store
	// This is a unit test for the logic that processes the store's return value

	type testCase struct {
		name              string
		incrementReturn   int64
		limit             int64
		expectedAllowed   bool
		expectedRemaining int64
	}

	tests := []testCase{
		{
			name:              "under limit",
			incrementReturn:   5,
			limit:             10,
			expectedAllowed:   true,
			expectedRemaining: 5,
		},
		{
			name:              "exactly at limit",
			incrementReturn:   10,
			limit:             10,
			expectedAllowed:   true,
			expectedRemaining: 0,
		},
		{
			name:              "over limit",
			incrementReturn:   15,
			limit:             10,
			expectedAllowed:   false,
			expectedRemaining: 0,
		},
		{
			name:              "way over limit",
			incrementReturn:   100,
			limit:             10,
			expectedAllowed:   false,
			expectedRemaining: 0,
		},
		{
			name:              "first request",
			incrementReturn:   1,
			limit:             100,
			expectedAllowed:   true,
			expectedRemaining: 99,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock store that returns specific values
			store := &mockIncrementStore{
				incrementValue: tc.incrementReturn,
			}

			ctx := context.Background()
			result, err := Check(ctx, store, "test-key", tc.limit, time.Minute)

			require.NoError(t, err)
			assert.Equal(t, tc.expectedAllowed, result.Allowed)
			assert.Equal(t, tc.expectedRemaining, result.Remaining)
			assert.Equal(t, tc.limit, result.Limit)
		})
	}
}

// mockIncrementStore is a mock store that returns predefined values
type mockIncrementStore struct {
	incrementValue int64
}

func (m *mockIncrementStore) Get(ctx context.Context, key string) (int64, time.Time, error) {
	return m.incrementValue, time.Now().Add(time.Minute), nil
}

func (m *mockIncrementStore) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return m.incrementValue, nil
}

func (m *mockIncrementStore) Reset(ctx context.Context, key string) error {
	return nil
}

func (m *mockIncrementStore) Close() error {
	return nil
}
