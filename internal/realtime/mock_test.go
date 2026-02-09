//nolint:errcheck // Test code - error handling not critical
package realtime

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// mockSubscriptionDB is a mock implementation of SubscriptionDB for testing.
// This is defined locally to avoid import cycles with internal/testutil.
type mockSubscriptionDB struct {
	mu               sync.RWMutex
	EnabledTables    map[string]bool
	RLSResults       map[string]bool
	OwnershipResults map[uuid.UUID]struct {
		IsOwner bool
		Exists  bool
	}
}

// newMockSubscriptionDB creates a new mock subscription database.
func newMockSubscriptionDB() *mockSubscriptionDB {
	return &mockSubscriptionDB{
		EnabledTables: make(map[string]bool),
		RLSResults:    make(map[string]bool),
		OwnershipResults: make(map[uuid.UUID]struct {
			IsOwner bool
			Exists  bool
		}),
	}
}

// EnableTable marks a table as enabled for realtime.
func (m *mockSubscriptionDB) EnableTable(schema, table string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.EnabledTables[schema+"."+table] = true
}

// IsTableRealtimeEnabled implements SubscriptionDB.
func (m *mockSubscriptionDB) IsTableRealtimeEnabled(ctx context.Context, schema, table string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.EnabledTables[schema+"."+table], nil
}

// CheckRLSAccess implements SubscriptionDB.
func (m *mockSubscriptionDB) CheckRLSAccess(ctx context.Context, schema, table, role string, claims map[string]interface{}, recordID interface{}) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := schema + "." + table + "." + fmt.Sprintf("%v", recordID)
	if result, exists := m.RLSResults[key]; exists {
		return result, nil
	}
	// Default: allow access
	return true, nil
}

// CheckRPCOwnership implements SubscriptionDB.
func (m *mockSubscriptionDB) CheckRPCOwnership(ctx context.Context, execID, userID uuid.UUID) (bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if result, exists := m.OwnershipResults[execID]; exists {
		return result.IsOwner, result.Exists, nil
	}
	return false, false, nil
}

// CheckJobOwnership implements SubscriptionDB.
func (m *mockSubscriptionDB) CheckJobOwnership(ctx context.Context, execID, userID uuid.UUID) (bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if result, exists := m.OwnershipResults[execID]; exists {
		return result.IsOwner, result.Exists, nil
	}
	return false, false, nil
}

// CheckFunctionOwnership implements SubscriptionDB.
func (m *mockSubscriptionDB) CheckFunctionOwnership(ctx context.Context, execID, userID uuid.UUID) (bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if result, exists := m.OwnershipResults[execID]; exists {
		return result.IsOwner, result.Exists, nil
	}
	return false, false, nil
}
