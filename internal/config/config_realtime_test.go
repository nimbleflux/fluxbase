package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRealtimeConfig_GetMaxMessageSize verifies the WebSocket inbound message
// size cap falls back to the default when unset.
func TestRealtimeConfig_GetMaxMessageSize(t *testing.T) {
	rc := &RealtimeConfig{}
	assert.Equal(t, int64(DefaultRealtimeMaxMessageSize), rc.GetMaxMessageSize())

	rc.MaxMessageSize = 1024
	assert.Equal(t, int64(1024), rc.GetMaxMessageSize())

	rc.MaxMessageSize = -5
	assert.Equal(t, int64(DefaultRealtimeMaxMessageSize), rc.GetMaxMessageSize())
}

// TestFunctionsConfig_MaxConcurrentExecutionsValidation verifies the global
// execution concurrency cap cannot be negative (0 = use default).
func TestFunctionsConfig_MaxConcurrentExecutionsValidation(t *testing.T) {
	valid := &FunctionsConfig{
		FunctionsDir:            "./functions",
		DefaultTimeout:          30,
		MaxTimeout:              60,
		DefaultMemoryLimit:      256,
		MaxMemoryLimit:          512,
		MaxConcurrentExecutions: 0,
	}
	assert.NoError(t, valid.Validate())

	invalid := &FunctionsConfig{
		FunctionsDir:            "./functions",
		DefaultTimeout:          30,
		MaxTimeout:              60,
		DefaultMemoryLimit:      256,
		MaxMemoryLimit:          512,
		MaxConcurrentExecutions: -1,
	}
	assert.Error(t, invalid.Validate())
}
