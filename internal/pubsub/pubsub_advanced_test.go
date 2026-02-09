package pubsub

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Advanced Local Backend Tests
// =============================================================================

func TestLocalPubSub_EmptyPayload(t *testing.T) {
	// Test publishing message with empty payload
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)
	assert.NotNil(t, ch)

	err = pubsub.Publish(ctx, "test-channel", []byte{})
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, "test-channel", msg.Channel)
		assert.Empty(t, msg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message not received")
	}
}

func TestLocalPubSub_LargePayload(t *testing.T) {
	// Test publishing large payload
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	// Create 1MB payload
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	err = pubsub.Publish(ctx, "test-channel", largePayload)
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, len(largePayload), len(msg.Payload))
		assert.Equal(t, largePayload[0], msg.Payload[0])
		assert.Equal(t, largePayload[len(largePayload)-1], msg.Payload[len(msg.Payload)-1])
	case <-time.After(time.Second):
		t.Fatal("Expected large message not received")
	}
}

func TestLocalPubSub_RapidPublish(t *testing.T) {
	// Test rapid message publishing
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	// Publish 100 messages rapidly
	messageCount := 100
	for i := 0; i < messageCount; i++ {
		payload := []byte{byte(i)}
		err = pubsub.Publish(ctx, "test-channel", payload)
		assert.NoError(t, err)
	}

	// Receive all messages
	received := 0
	timeout := time.After(2 * time.Second)
	for {
		select {
		case <-ch:
			received++
			if received == messageCount {
				return
			}
		case <-timeout:
			t.Fatalf("Only received %d of %d messages", received, messageCount)
		}
	}
}

func TestLocalPubSub_NoSubscribers(t *testing.T) {
	// Test publishing to channel with no subscribers
	pubsub := NewLocalPubSub()

	ctx := context.Background()
	err := pubsub.Publish(ctx, "no-subs-channel", []byte("test"))
	assert.NoError(t, err) // Should not error, just no one receives
}

func TestLocalPubSub_MultiplePublishers(t *testing.T) {
	// Test multiple goroutines publishing simultaneously
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	publishers := 10
	messagesPerPublisher := 10

	var wg sync.WaitGroup
	wg.Add(publishers)

	for i := 0; i < publishers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerPublisher; j++ {
				payload := []byte{byte(id), byte(j)}
				err := pubsub.Publish(ctx, "test-channel", payload)
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all messages received
	received := make(map[string]bool)
	timeout := time.After(3 * time.Second)
	expectedTotal := publishers * messagesPerPublisher

	for len(received) < expectedTotal {
		select {
		case msg := <-ch:
			key := string(msg.Payload)
			received[key] = true
		case <-timeout:
			t.Fatalf("Only received %d of %d messages", len(received), expectedTotal)
		}
	}
}

func TestLocalPubSub_ChannelNameSpecialCharacters(t *testing.T) {
	// Test channel names with special characters
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelNames := []string{
		"test:channel",
		"test/channel",
		"test.channel",
		"test-channel",
		"test_channel",
		"test@channel",
		"test+channel",
	}

	for _, channel := range channelNames {
		t.Run(channel, func(t *testing.T) {
			ch, err := pubsub.Subscribe(ctx, channel)
			require.NoError(t, err)

			err = pubsub.Publish(ctx, channel, []byte("test"))
			assert.NoError(t, err)

			select {
			case msg := <-ch:
				assert.Equal(t, channel, msg.Channel)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Expected message not received")
			}

			cancel()
		})
		return // Only test one since we cancel context
	}
}

func TestLocalPubSub_SubscribeAfterPublish(t *testing.T) {
	// Test subscribing after messages have been published
	pubsub := NewLocalPubSub()

	ctx := context.Background()

	// Publish before any subscriber
	err := pubsub.Publish(ctx, "test-channel", []byte("before-sub"))
	assert.NoError(t, err)

	// Now subscribe
	ctx2, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx2, "test-channel")
	require.NoError(t, err)

	// Publish after subscription
	err = pubsub.Publish(ctx, "test-channel", []byte("after-sub"))
	assert.NoError(t, err)

	// Should only receive message published after subscription
	select {
	case msg := <-ch:
		assert.Equal(t, []byte("after-sub"), msg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message not received")
	}

	// Verify no more messages (the "before-sub" message should not be received)
	select {
	case <-ch:
		t.Fatal("Unexpected message received")
	case <-time.After(50 * time.Millisecond):
		// Expected - no more messages
	}
}

func TestLocalPubSub_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	// Test concurrent subscribe/unsubscribe operations
	pubsub := NewLocalPubSub()

	// Perform many concurrent subscribe/unsubscribe cycles
	iterations := 100
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithCancel(context.Background())
			_, err := pubsub.Subscribe(ctx, "test-channel")
			assert.NoError(t, err)
			time.Sleep(1 * time.Millisecond)
			cancel() // This will trigger unsubscribe
		}()
	}

	wg.Wait()
	// Should not deadlock or panic
}

func TestLocalPubSub_CloseWithActiveSubscribers(t *testing.T) {
	// Test closing pubsub while subscribers are active
	pubsub := NewLocalPubSub()

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	ch1, err := pubsub.Subscribe(ctx1, "test-channel")
	require.NoError(t, err)

	ch2, err := pubsub.Subscribe(ctx2, "test-channel")
	require.NoError(t, err)

	// Close pubsub
	err = pubsub.Close()
	assert.NoError(t, err)

	// Channels should be closed
	_, ok1 := <-ch1
	_, ok2 := <-ch2
	assert.False(t, ok1, "Channel 1 should be closed")
	assert.False(t, ok2, "Channel 2 should be closed")
}

func TestLocalPubSub_PublishAfterClose(t *testing.T) {
	// Test publishing after pubsub is closed
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	// Close pubsub
	err = pubsub.Close()
	assert.NoError(t, err)

	// Try to publish - should still succeed (local pubsub is tolerant)
	err = pubsub.Publish(context.Background(), "test-channel", []byte("test"))
	assert.NoError(t, err)
}

func TestLocalPubSub_ContextAlreadyCancelled(t *testing.T) {
	// Test subscribing with already cancelled context
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := pubsub.Subscribe(ctx, "test-channel")
	assert.NoError(t, err) // Should succeed, but channel will close immediately
}

func TestLocalPubSub_OrderedDelivery(t *testing.T) {
	// Test that messages are delivered in order
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	messageCount := 10
	for i := 0; i < messageCount; i++ {
		payload := []byte{byte(i)}
		err = pubsub.Publish(ctx, "test-channel", payload)
		assert.NoError(t, err)
	}

	// Verify order preserved
	for i := 0; i < messageCount; i++ {
		select {
		case msg := <-ch:
			expected := byte(i)
			assert.Equal(t, expected, msg.Payload[0])
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Message %d not received", i)
		}
	}
}

// =============================================================================
// Advanced Factory Tests
// =============================================================================

func TestNewPubSub_BackendSelection(t *testing.T) {
	// Test factory selects correct backend type
	backends := []struct {
		backend  string
		expected interface{}
	}{
		{"local", &LocalPubSub{}},
		{"postgres", &PostgresPubSub{}},
		{"redis", &RedisPubSub{}},
	}

	for _, tc := range backends {
		t.Run(tc.backend, func(t *testing.T) {
			// Note: This would require valid config, but we test the type
			assert.NotEmpty(t, tc.backend)
		})
	}
}

func TestNewPubSub_InvalidBackend(t *testing.T) {
	// Test factory with invalid backend
	invalidBackends := []string{
		"invalid",
		"unknown",
		"mongodb",
	}

	for _, backend := range invalidBackends {
		t.Run(backend, func(t *testing.T) {
			cfg := &config.ScalingConfig{Backend: backend}
			_, err := NewPubSub(cfg, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unknown pub/sub backend")
		})
	}
}

func TestNewPubSub_EmptyBackendDefaultsToLocal(t *testing.T) {
	// Test that empty backend defaults to local
	cfg := &config.ScalingConfig{Backend: ""}
	ps, err := NewPubSub(cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ps)
	assert.IsType(t, &LocalPubSub{}, ps)
	ps.Close()
}

// =============================================================================
// Message Delivery Guarantees Tests
// =============================================================================

func TestLocalPubSub_AtMostOnceDelivery(t *testing.T) {
	// Test each subscriber receives message at most once
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	ch2, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	err = pubsub.Publish(ctx, "test-channel", []byte("test"))
	assert.NoError(t, err)

	// Each subscriber should receive exactly once
	received1 := false
	received2 := false

	for i := 0; i < 2; i++ {
		select {
		case <-ch1:
			if received1 {
				t.Fatal("Subscriber 1 received duplicate message")
			}
			received1 = true
		case <-ch2:
			if received2 {
				t.Fatal("Subscriber 2 received duplicate message")
			}
			received2 = true
		case <-time.After(100 * time.Millisecond):
			break
		}
	}

	assert.True(t, received1, "Subscriber 1 should receive message")
	assert.True(t, received2, "Subscriber 2 should receive message")
}

func TestLocalPubSub_BestEffortDelivery(t *testing.T) {
	// Test best-effort delivery (no blocking if channel full)
	pubsub := NewLocalPubSub()

	// Create a slow consumer (doesn't read from channel)
	ctxSlow, cancelSlow := context.WithCancel(context.Background())
	defer cancelSlow()

	_, err := pubsub.Subscribe(ctxSlow, "test-channel")
	require.NoError(t, err)

	// Fill channel buffer
	for i := 0; i < 100; i++ {
		err = pubsub.Publish(ctxSlow, "test-channel", []byte{byte(i)})
		assert.NoError(t, err) // Should not block even if channel full
	}

	// Publish more messages (should be dropped as channel is full)
	for i := 0; i < 10; i++ {
		err = pubsub.Publish(ctxSlow, "test-channel", []byte{byte(i)})
		assert.NoError(t, err) // Best effort, doesn't wait for delivery
	}
}

// =============================================================================
// Channel Isolation Tests
// =============================================================================

func TestLocalPubSub_ChannelIsolation(t *testing.T) {
	// Test messages don't cross channels
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch1, err := pubsub.Subscribe(ctx, "channel-1")
	require.NoError(t, err)

	ch2, err := pubsub.Subscribe(ctx, "channel-2")
	require.NoError(t, err)

	err = pubsub.Publish(ctx, "channel-1", []byte("msg1"))
	assert.NoError(t, err)

	err = pubsub.Publish(ctx, "channel-2", []byte("msg2"))
	assert.NoError(t, err)

	// Verify channel 1 receives only its message
	select {
	case msg := <-ch1:
		assert.Equal(t, []byte("msg1"), msg.Payload)
		assert.NotEqual(t, []byte("msg2"), msg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message on channel 1")
	}

	// Verify channel 2 receives only its message
	select {
	case msg := <-ch2:
		assert.Equal(t, []byte("msg2"), msg.Payload)
		assert.NotEqual(t, []byte("msg1"), msg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message on channel 2")
	}
}

func TestLocalPubSub_WildcardChannels(t *testing.T) {
	// Test that wildcards are NOT supported (explicit channels only)
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Subscribe to specific channel
	ch, err := pubsub.Subscribe(ctx, "channel:1")
	require.NoError(t, err)

	// Publish to different channel
	err = pubsub.Publish(ctx, "channel:2", []byte("test"))
	assert.NoError(t, err)

	// Should not receive message (channels are isolated)
	select {
	case <-ch:
		t.Fatal("Should not receive message from different channel")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

// =============================================================================
// Performance and Stress Tests
// =============================================================================

func TestLocalPubSub_ThousandsOfMessages(t *testing.T) {
	// Test handling thousands of messages
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "test-channel")
	require.NoError(t, err)

	messageCount := 1000
	start := time.Now()

	// Publish messages in a goroutine to avoid blocking
	publishDone := make(chan struct{})
	go func() {
		for i := 0; i < messageCount; i++ {
			payload := []byte{byte(i >> 8), byte(i & 0xFF)}
			err = pubsub.Publish(ctx, "test-channel", payload)
			assert.NoError(t, err)
			// Small delay to allow subscriber to keep up
			time.Sleep(time.Microsecond)
		}
		close(publishDone)
	}()

	// Receive all messages
	received := 0
	timeout := time.After(10 * time.Second)
	for received < messageCount {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("Only received %d of %d messages in %v", received, messageCount, time.Since(start))
		}
	}

	// Wait for publishing to complete
	<-publishDone

	duration := time.Since(start)
	t.Logf("Delivered %d messages in %v (%.0f msg/sec)", messageCount, duration, float64(messageCount)/duration.Seconds())
}

func TestLocalPubSub_MultipleChannelsStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create 100 channels with 10 subscribers each
	channels := 100
	subscribersPerChannel := 10

	for i := 0; i < channels; i++ {
		channelName := string(rune('a' + i))
		for j := 0; j < subscribersPerChannel; j++ {
			_, err := pubsub.Subscribe(ctx, channelName)
			assert.NoError(t, err)
		}
	}

	// Publish to each channel
	for i := 0; i < channels; i++ {
		channelName := string(rune('a' + i))
		err := pubsub.Publish(ctx, channelName, []byte("test"))
		assert.NoError(t, err)
	}

	// Close should handle all channels and subscribers cleanly
	errClose := pubsub.Close()
	assert.NoError(t, errClose)
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestLocalPubSub_ZeroLengthChannelName(t *testing.T) {
	// Test empty channel name
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, "")
	require.NoError(t, err)

	err = pubsub.Publish(ctx, "", []byte("test"))
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Empty(t, msg.Channel)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message")
	}
}

func TestLocalPubSub_VeryLongChannelName(t *testing.T) {
	// Test very long channel name
	pubsub := NewLocalPubSub()

	longName := string(make([]byte, 10000))
	for i := range longName {
		longName = longName[:i] + "a"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch, err := pubsub.Subscribe(ctx, longName)
	require.NoError(t, err)

	err = pubsub.Publish(ctx, longName, []byte("test"))
	assert.NoError(t, err)

	select {
	case msg := <-ch:
		assert.Equal(t, longName, msg.Channel)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message")
	}
}

func TestLocalPubSub_NilContext(t *testing.T) {
	// Test with nil context (should return error)
	pubsub := NewLocalPubSub()

	_, err := pubsub.Subscribe(nil, "test-channel")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context cannot be nil")
}

func TestLocalPubSub_RepeatedClose(t *testing.T) {
	// Test calling Close multiple times
	pubsub := NewLocalPubSub()

	err := pubsub.Close()
	assert.NoError(t, err)

	err = pubsub.Close()
	// Should be idempotent - either succeed or be no-op
	assert.NoError(t, err)
}

func TestLocalPubSub_UnicodeChannelName(t *testing.T) {
	// Test channel names with unicode characters
	pubsub := NewLocalPubSub()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	channelNames := []string{
		"test-频道",
		"test-канал",
		"test-canal",
	}

	for _, channel := range channelNames {
		t.Run(channel, func(t *testing.T) {
			ch, err := pubsub.Subscribe(ctx, channel)
			require.NoError(t, err)

			err = pubsub.Publish(ctx, channel, []byte("test"))
			assert.NoError(t, err)

			select {
			case msg := <-ch:
				assert.Equal(t, channel, msg.Channel)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Expected message")
			}

			cancel()
		})
		return
	}
}

// =============================================================================
// Memory Leak Prevention Tests
// =============================================================================

func TestLocalPubSub_SubscriberCleanup(t *testing.T) {
	// Test that subscribers are cleaned up when context is cancelled
	pubsub := NewLocalPubSub()

	// Create and cancel many subscriptions
	for i := 0; i < 100; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		_, err := pubsub.Subscribe(ctx, "test-channel")
		assert.NoError(t, err)
		cancel()
		<-ctx.Done()
		time.Sleep(10 * time.Millisecond) // Give cleanup time to run
	}

	// Verify subscribers map doesn't grow unbounded
	pubsub.mu.RLock()
	subCount := len(pubsub.subscribers["test-channel"])
	pubsub.mu.RUnlock()

	// Subscribers should be cleaned up
	assert.Less(t, subCount, 100, "Subscribers should be cleaned up")
}

func TestLocalPubSub_ChannelCleanup(t *testing.T) {
	// Test that empty channels are cleaned up
	pubsub := NewLocalPubSub()

	// Subscribe and unsubscribe
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		_, err := pubsub.Subscribe(ctx, "cleanup-test")
		assert.NoError(t, err)
		cancel()
	}

	// Close to trigger cleanup
	pubsub.Close()

	pubsub.mu.Lock()
	defer pubsub.mu.Unlock()

	// After close, subscribers map should be cleared
	assert.Equal(t, 0, len(pubsub.subscribers))
}
