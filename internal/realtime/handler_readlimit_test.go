package realtime

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// TestRealtimeHandler_MaxMessageSize verifies the inbound WebSocket read limit
// falls back to the default and honors the configured value.
func TestRealtimeHandler_MaxMessageSize(t *testing.T) {
	t.Run("default when no manager config", func(t *testing.T) {
		h := NewRealtimeHandler(nil, nil, nil)
		assert.Equal(t, int64(config.DefaultRealtimeMaxMessageSize), h.maxMessageSize())
	})

	t.Run("default when config value unset", func(t *testing.T) {
		mgr := NewManagerWithConfig(context.Background(), ManagerConfig{})
		mgr.SetBaseConfig(&config.Config{})
		h := NewRealtimeHandler(mgr, nil, nil)
		assert.Equal(t, int64(config.DefaultRealtimeMaxMessageSize), h.maxMessageSize())
	})

	t.Run("configured value is honored", func(t *testing.T) {
		mgr := NewManagerWithConfig(context.Background(), ManagerConfig{})
		mgr.SetBaseConfig(&config.Config{
			Realtime: config.RealtimeConfig{MaxMessageSize: 131072},
		})
		h := NewRealtimeHandler(mgr, nil, nil)
		assert.Equal(t, int64(131072), h.maxMessageSize())
	})
}
