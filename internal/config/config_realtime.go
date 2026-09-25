package config

import "github.com/spf13/viper"

// DefaultRealtimeMaxMessageSize is the default cap on inbound WebSocket message
// size for realtime connections. Inbound frames are otherwise unbounded, which
// lets a client exhaust memory with a single large frame.
const DefaultRealtimeMaxMessageSize = 64 * 1024

func init() {
	viper.SetDefault("realtime.max_message_size", DefaultRealtimeMaxMessageSize)
}

// GetMaxMessageSize returns the configured inbound WebSocket message size cap,
// falling back to DefaultRealtimeMaxMessageSize when unset or non-positive.
func (rc *RealtimeConfig) GetMaxMessageSize() int64 {
	if rc.MaxMessageSize > 0 {
		return rc.MaxMessageSize
	}
	return DefaultRealtimeMaxMessageSize
}
