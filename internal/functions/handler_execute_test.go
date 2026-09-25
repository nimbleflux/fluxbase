package functions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAcquireExecSlot verifies the global execution semaphore: slots are
// acquired and released, and acquisition gives up after the bounded wait.
func TestAcquireExecSlot(t *testing.T) {
	t.Run("acquire and release", func(t *testing.T) {
		h := &Handler{execSlots: make(chan struct{}, 1)}
		release, ok := h.acquireExecSlot(context.Background(), nil)
		require.True(t, ok)
		require.NotNil(t, release)

		// Slot is held: a second acquisition with a tiny wait budget fails.
		tiny := 50 * time.Millisecond
		_, ok = h.acquireExecSlot(context.Background(), &tiny)
		assert.False(t, ok, "second acquisition should time out while the only slot is held")

		release()

		// After release the slot is available again.
		release2, ok := h.acquireExecSlot(context.Background(), nil)
		require.True(t, ok)
		release2()
	})

	t.Run("wait bounded by function timeout override", func(t *testing.T) {
		h := &Handler{execSlots: make(chan struct{}, 1)}
		release, ok := h.acquireExecSlot(context.Background(), nil)
		require.True(t, ok)
		defer release()

		// A short function timeout caps the acquisition wait: this must fail
		// fast rather than waiting the default 30s.
		tiny := 50 * time.Millisecond
		start := time.Now()
		_, ok = h.acquireExecSlot(context.Background(), &tiny)
		assert.False(t, ok)
		assert.Less(t, time.Since(start), 5*time.Second, "acquisition wait must be bounded by the function timeout")
	})

	t.Run("request cancellation aborts acquisition", func(t *testing.T) {
		h := &Handler{execSlots: make(chan struct{}, 1)}
		release, ok := h.acquireExecSlot(context.Background(), nil)
		require.True(t, ok)
		defer release()

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()
		_, ok = h.acquireExecSlot(ctx, nil)
		assert.False(t, ok)
	})
}

// TestFilterPassthroughHeaders verifies only allowlisted function response
// headers are propagated to clients.
func TestFilterPassthroughHeaders(t *testing.T) {
	in := map[string]string{
		"Content-Type":        "application/json",
		"Content-Disposition": "attachment; filename=report.csv",
		"Cache-Control":       "no-store",
		"ETag":                "\"abc123\"",
		"Last-Modified":       "Wed, 21 Oct 2015 07:28:00 GMT",
		"X-Request-Id":        "req-123",
		// dropped: infrastructure/security-sensitive
		"Set-Cookie":            "session=attacker",
		"Authorization":         "Bearer stolen",
		"X-Service-Role-Key":    "secret",
		"Access-Control-Origin": "*",
		"Server":                "custom",
	}

	out := filterPassthroughHeaders(in)

	assert.Len(t, out, 6)
	assert.Equal(t, "application/json", out["Content-Type"])
	assert.Equal(t, "attachment; filename=report.csv", out["Content-Disposition"])
	assert.Equal(t, "no-store", out["Cache-Control"])
	assert.Equal(t, "\"abc123\"", out["ETag"])
	assert.Equal(t, "Wed, 21 Oct 2015 07:28:00 GMT", out["Last-Modified"])
	assert.Equal(t, "req-123", out["X-Request-Id"])

	assert.NotContains(t, out, "Set-Cookie")
	assert.NotContains(t, out, "Authorization")
	assert.NotContains(t, out, "X-Service-Role-Key")
	assert.NotContains(t, out, "Access-Control-Origin")
	assert.NotContains(t, out, "Server")
}
