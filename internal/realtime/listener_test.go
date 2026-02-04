package realtime

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// Mock handler for testing
type MockRealtimeHandler struct {
	broadcastCalls []struct {
		channel string
		event   ChangeEvent
	}
}

func (m *MockRealtimeHandler) Broadcast(channel string, payload interface{}) {
	event, ok := payload.(ChangeEvent)
	if ok {
		m.broadcastCalls = append(m.broadcastCalls, struct {
			channel string
			event   ChangeEvent
		}{channel, event})
	}
}

func TestListener_ProcessNotification_Insert(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil) // nil auth service and sub manager for testing
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	// Create a notification
	notification := &pgconn.Notification{
		Channel: "fluxbase_changes",
		Payload: `{
			"type": "INSERT",
			"table": "products",
			"schema": "public",
			"record": {"id": 1, "name": "Test Product", "price": 99.99}
		}`,
	}

	// Process notification without subscribers
	// Should not panic even if no one is listening
	listener.processNotification(notification)
}

func TestListener_ProcessNotification_Update(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	notification := &pgconn.Notification{
		Channel: "fluxbase_changes",
		Payload: `{
			"type": "UPDATE",
			"table": "products",
			"schema": "public",
			"record": {"id": 1, "name": "Updated Product", "price": 149.99},
			"old_record": {"id": 1, "name": "Test Product", "price": 99.99}
		}`,
	}

	listener.processNotification(notification)
	// Should not panic
}

func TestListener_ProcessNotification_Delete(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	notification := &pgconn.Notification{
		Channel: "fluxbase_changes",
		Payload: `{
			"type": "DELETE",
			"table": "products",
			"schema": "public",
			"old_record": {"id": 1, "name": "Test Product", "price": 99.99}
		}`,
	}

	listener.processNotification(notification)
	// Should not panic
}

func TestListener_ProcessNotification_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	notification := &pgconn.Notification{
		Channel: "fluxbase_changes",
		Payload: `{invalid json`,
	}

	// Should handle error gracefully without panicking
	listener.processNotification(notification)
}

func TestListener_ProcessNotification_ChannelFormat(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	tests := []struct {
		name            string
		schema          string
		table           string
		expectedChannel string
	}{
		{
			name:            "public schema",
			schema:          "public",
			table:           "products",
			expectedChannel: "table:public.products",
		},
		{
			name:            "custom schema",
			schema:          "inventory",
			table:           "items",
			expectedChannel: "table:inventory.items",
		},
		{
			name:            "auth schema",
			schema:          "auth",
			table:           "users",
			expectedChannel: "table:auth.users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := `{
				"type": "INSERT",
				"table": "` + tt.table + `",
				"schema": "` + tt.schema + `",
				"record": {"id": 1}
			}`

			notification := &pgconn.Notification{
				Channel: "fluxbase_changes",
				Payload: payload,
			}

			// Should not panic
			listener.processNotification(notification)
		})
	}
}

func TestChangeEvent_Structure(t *testing.T) {
	event := ChangeEvent{
		Type:   "INSERT",
		Table:  "products",
		Schema: "public",
		Record: map[string]interface{}{
			"id":    float64(1),
			"name":  "Test Product",
			"price": float64(99.99),
		},
	}

	assert.Equal(t, "INSERT", event.Type)
	assert.Equal(t, "products", event.Table)
	assert.Equal(t, "public", event.Schema)
	assert.NotNil(t, event.Record)
	assert.Equal(t, float64(1), event.Record["id"])
}

func TestChangeEvent_WithOldRecord(t *testing.T) {
	event := ChangeEvent{
		Type:   "UPDATE",
		Table:  "products",
		Schema: "public",
		Record: map[string]interface{}{
			"id":    float64(1),
			"name":  "Updated Product",
			"price": float64(149.99),
		},
		OldRecord: map[string]interface{}{
			"id":    float64(1),
			"name":  "Test Product",
			"price": float64(99.99),
		},
	}

	assert.Equal(t, "UPDATE", event.Type)
	assert.NotNil(t, event.Record)
	assert.NotNil(t, event.OldRecord)
	assert.Equal(t, "Updated Product", event.Record["name"])
	assert.Equal(t, "Test Product", event.OldRecord["name"])
}

func TestNewListener(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)

	listener := NewListener(nil, handler, nil, nil)

	assert.NotNil(t, listener)
	assert.NotNil(t, listener.handler)
	assert.NotNil(t, listener.ctx)
	assert.NotNil(t, listener.cancel)
}

func TestListener_Stop(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)

	listener := NewListener(nil, handler, nil, nil)

	// Should not panic
	listener.Stop()

	// Verify context is cancelled
	assert.Error(t, listener.ctx.Err())
}

// =============================================================================
// Channel Constants Tests
// =============================================================================

func TestLogChannelConstants(t *testing.T) {
	t.Run("LogChannel is defined", func(t *testing.T) {
		assert.NotEmpty(t, LogChannel)
		assert.Equal(t, "fluxbase:logs", LogChannel)
	})

	t.Run("AllLogsChannel is defined", func(t *testing.T) {
		assert.NotEmpty(t, AllLogsChannel)
		assert.Equal(t, "fluxbase:all_logs", AllLogsChannel)
	})

	t.Run("channels are distinct", func(t *testing.T) {
		assert.NotEqual(t, LogChannel, AllLogsChannel)
	})
}

// =============================================================================
// ChangeEvent Extended Tests
// =============================================================================

func TestChangeEvent_AllTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
	}{
		{"INSERT", "INSERT"},
		{"UPDATE", "UPDATE"},
		{"DELETE", "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := ChangeEvent{
				Type:   tt.eventType,
				Table:  "test_table",
				Schema: "public",
			}

			assert.Equal(t, tt.eventType, event.Type)
		})
	}
}

func TestChangeEvent_EmptyFields(t *testing.T) {
	t.Run("empty event", func(t *testing.T) {
		event := ChangeEvent{}

		assert.Empty(t, event.Type)
		assert.Empty(t, event.Table)
		assert.Empty(t, event.Schema)
		assert.Nil(t, event.Record)
		assert.Nil(t, event.OldRecord)
	})
}

func TestChangeEvent_ComplexRecord(t *testing.T) {
	t.Run("record with nested data", func(t *testing.T) {
		event := ChangeEvent{
			Type:   "INSERT",
			Table:  "orders",
			Schema: "public",
			Record: map[string]interface{}{
				"id":     float64(1),
				"status": "pending",
				"metadata": map[string]interface{}{
					"source":   "web",
					"campaign": "summer_sale",
				},
				"items": []interface{}{
					map[string]interface{}{"product_id": float64(1), "quantity": float64(2)},
					map[string]interface{}{"product_id": float64(2), "quantity": float64(1)},
				},
			},
		}

		assert.NotNil(t, event.Record)
		assert.NotNil(t, event.Record["metadata"])
		assert.NotNil(t, event.Record["items"])
	})

	t.Run("record with null values", func(t *testing.T) {
		event := ChangeEvent{
			Type:   "INSERT",
			Table:  "users",
			Schema: "public",
			Record: map[string]interface{}{
				"id":         float64(1),
				"name":       "John",
				"avatar_url": nil,
				"bio":        nil,
			},
		}

		assert.Nil(t, event.Record["avatar_url"])
		assert.Nil(t, event.Record["bio"])
	})
}

// =============================================================================
// enrichJobWithETA Tests
// =============================================================================

func TestListener_enrichJobWithETA(t *testing.T) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	t.Run("no progress data", func(t *testing.T) {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":     "job-123",
				"status": "running",
			},
		}

		listener.enrichJobWithETA(event)

		// Should not add any progress fields
		_, hasProgressPercent := event.Record["progress_percent"]
		assert.False(t, hasProgressPercent)
	})

	t.Run("progress with percent only", func(t *testing.T) {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":     "job-123",
				"status": "running",
				"progress": map[string]interface{}{
					"percent": float64(50),
				},
			},
		}

		listener.enrichJobWithETA(event)

		assert.Equal(t, 50, event.Record["progress_percent"])
	})

	t.Run("progress with message", func(t *testing.T) {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":     "job-123",
				"status": "running",
				"progress": map[string]interface{}{
					"percent": float64(75),
					"message": "Processing items...",
				},
			},
		}

		listener.enrichJobWithETA(event)

		assert.Equal(t, 75, event.Record["progress_percent"])
		assert.Equal(t, "Processing items...", event.Record["progress_message"])
	})

	t.Run("progress with existing ETA", func(t *testing.T) {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":     "job-123",
				"status": "running",
				"progress": map[string]interface{}{
					"percent":                float64(60),
					"estimated_seconds_left": float64(120),
				},
			},
		}

		listener.enrichJobWithETA(event)

		assert.Equal(t, 60, event.Record["progress_percent"])
		assert.Equal(t, 120, event.Record["estimated_seconds_left"])
	})

	t.Run("nil progress", func(t *testing.T) {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":       "job-123",
				"status":   "running",
				"progress": nil,
			},
		}

		listener.enrichJobWithETA(event)

		// Should not panic and not add any progress fields
		_, hasProgressPercent := event.Record["progress_percent"]
		assert.False(t, hasProgressPercent)
	})
}

// =============================================================================
// RealtimeListener Interface Tests
// =============================================================================

func TestRealtimeListener_Interface(t *testing.T) {
	t.Run("Listener implements RealtimeListener", func(t *testing.T) {
		var _ RealtimeListener = (*Listener)(nil)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkProcessNotification(b *testing.B) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	notification := &pgconn.Notification{
		Channel: "fluxbase_changes",
		Payload: `{
			"type": "INSERT",
			"table": "products",
			"schema": "public",
			"record": {"id": 1, "name": "Test Product", "price": 99.99}
		}`,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listener.processNotification(notification)
	}
}

func BenchmarkEnrichJobWithETA(b *testing.B) {
	ctx := context.Background()
	manager := NewManager(ctx)
	handler := NewRealtimeHandler(manager, nil, nil)
	listener := &Listener{
		handler: handler,
		ctx:     ctx,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		event := &ChangeEvent{
			Type:   "UPDATE",
			Table:  "queue",
			Schema: "jobs",
			Record: map[string]interface{}{
				"id":     "job-123",
				"status": "running",
				"progress": map[string]interface{}{
					"percent": float64(50),
					"message": "Processing...",
				},
			},
		}
		listener.enrichJobWithETA(event)
	}
}
