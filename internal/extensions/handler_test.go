package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with service", func(t *testing.T) {
		service := &Service{}
		handler := NewHandler(service)

		assert.NotNil(t, handler)
		assert.Equal(t, service, handler.service)
	})

	t.Run("creates handler with nil service", func(t *testing.T) {
		handler := NewHandler(nil)

		assert.NotNil(t, handler)
		assert.Nil(t, handler.service)
	})
}

func TestHandler_Struct(t *testing.T) {
	// Test that handler struct is properly defined
	var h Handler
	assert.Nil(t, h.service)
}

// =============================================================================
// Handler Field Tests
// =============================================================================

func TestHandler_Fields(t *testing.T) {
	t.Run("handler fields are properly initialized", func(t *testing.T) {
		service := &Service{}
		handler := &Handler{
			service: service,
		}

		assert.NotNil(t, handler.service)
		assert.Equal(t, service, handler.service)
	})

	t.Run("handler with zero value", func(t *testing.T) {
		handler := Handler{}

		assert.Nil(t, handler.service)
	})
}

// =============================================================================
// Handler Interface Compliance Tests
// =============================================================================

func TestHandler_Methods(t *testing.T) {
	t.Run("handler methods are defined", func(t *testing.T) {
		service := &Service{}
		handler := NewHandler(service)

		// Verify handler is not nil
		assert.NotNil(t, handler)

		// The actual methods require fiber.Ctx which needs integration tests
		// This test verifies the handler can be created and has the right type
	})
}

// =============================================================================
// Service Dependency Tests
// =============================================================================

func TestHandler_ServiceDependency(t *testing.T) {
	t.Run("multiple handlers can share same service", func(t *testing.T) {
		service := &Service{}
		handler1 := NewHandler(service)
		handler2 := NewHandler(service)

		assert.Same(t, handler1.service, handler2.service)
	})

	t.Run("handlers with different services are independent", func(t *testing.T) {
		service1 := &Service{}
		service2 := &Service{}
		handler1 := NewHandler(service1)
		handler2 := NewHandler(service2)

		assert.NotSame(t, handler1.service, handler2.service)
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkNewHandler(b *testing.B) {
	service := &Service{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewHandler(service)
	}
}

func BenchmarkNewHandler_NilService(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewHandler(nil)
	}
}

// Note: Full handler tests require a mock fiber.Ctx and Service
// which is typically done in integration tests
