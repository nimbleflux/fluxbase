//go:build integration && !no_e2e

// Package testutil provides shared test utilities and mocks for unit testing.
package testutil

import (
	"sync"
	"testing"

	test "github.com/fluxbase-eu/fluxbase/test"
)

// Package-level singleton for shared test context.
// All tests in the same package share a single database connection to avoid pool exhaustion.
var (
	sharedContext            *IntegrationTestContext
	sharedContextOnce        sync.Once
	sharedContextMu          sync.RWMutex
	sharedContextInitialized bool // Track if context has been initialized
)

// IntegrationTestContext provides test context for integration tests.
// This embeds test.TestContext to provide all helper methods (NewRequest, CreateTestUser, etc.)
// while allowing internal packages to use E2E test infrastructure.
//
// Integration tests use real PostgreSQL database connections and test actual behavior
// instead of mocked behavior. They should be run with the `integration` build tag.
//
// IMPORTANT: All tests in the same package share a single database connection via singleton pattern.
// Tests should NOT call Close() on the returned context - the connection is managed at package level.
type IntegrationTestContext struct {
	*test.TestContext // Embedded to provide all helper methods (NewRequest, CreateTestUser, etc.)
	T                 *testing.T
}

// NewIntegrationTestContext creates or returns a shared test context for integration tests.
//
// All tests in the same package share a single database connection to avoid pool exhaustion.
// The first call creates the connection, subsequent calls reuse it.
//
// This uses fluxbase_app database user WITH BYPASSRLS privilege (RLS policies are NOT enforced).
//
// Use this for:
//   - General integration testing
//   - Testing business logic with real database
//   - API endpoint testing
//
// Do NOT use for:
//   - Testing RLS policies (use NewRLSIntegrationTestContext instead)
//
// IMPORTANT: Tests should NOT call Close() on the returned context.
// The shared connection is automatically managed by the package-level singleton.
func NewIntegrationTestContext(t *testing.T) *IntegrationTestContext {
	sharedContextMu.Lock()
	defer sharedContextMu.Unlock()

	// Check if this is the first call (before initialization)
	wasInitialized := sharedContextInitialized

	// Initialize shared context once per package
	sharedContextOnce.Do(func() {
		sharedContext = &IntegrationTestContext{
			TestContext: test.NewTestContext(&testing.T{}),
			T:           &testing.T{},
		}
		sharedContextInitialized = true
	})

	// Reset global state at the beginning of each test (except the first)
	// This prevents rate limiting and other cross-test interference
	if wasInitialized {
		test.ResetGlobalTestState()
	}

	// Update T with current test for proper error reporting
	sharedContext.T = t
	sharedContext.TestContext.T = t

	return sharedContext
}

// NewRLSIntegrationTestContext creates or returns a shared RLS-enabled test context.
//
// All tests in the same package share a single database connection to avoid pool exhaustion.
// The first call creates the connection, subsequent calls reuse it.
//
// This uses fluxbase_rls_test database user WITHOUT BYPASSRLS privilege (RLS policies ARE enforced).
//
// Use this ONLY for:
//   - Testing RLS policies
//   - Verifying data isolation between users
//   - Testing security boundaries
//
// Do NOT use for:
//   - General API testing (use NewIntegrationTestContext instead)
//
// IMPORTANT: Tests should NOT call Close() on the returned context.
// The shared connection is automatically managed by the package-level singleton.
func NewRLSIntegrationTestContext(t *testing.T) *IntegrationTestContext {
	sharedContextMu.Lock()
	defer sharedContextMu.Unlock()

	// Check if this is the first call (before initialization)
	wasInitialized := sharedContextInitialized

	// Initialize shared RLS context once per package
	sharedContextOnce.Do(func() {
		sharedContext = &IntegrationTestContext{
			TestContext: test.NewRLSTestContext(&testing.T{}),
			T:           &testing.T{},
		}
		sharedContextInitialized = true
	})

	// Reset global state at the beginning of each test (except the first)
	// This prevents rate limiting and other cross-test interference
	if wasInitialized {
		test.ResetGlobalTestState()
	}

	// Update T with current test for proper error reporting
	sharedContext.T = t
	sharedContext.TestContext.T = t

	return sharedContext
}

// Close is a no-op for shared contexts.
//
// The shared connection is managed at the package level by the singleton pattern.
// Individual tests should not call Close() - the connection persists for all tests in the package.
func (tc *IntegrationTestContext) Close() {
	// No-op for shared contexts
	// Connection cleanup happens automatically when the package's tests complete
}

// CleanupTestData cleans up test data from the database.
// This should be called between tests to ensure test isolation.
//
// It cleans up:
// - Auth users with test email prefixes (e2e-test-*, test-*@example.com, test-*@test.com)
// - Dashboard users with test email prefixes
func (tc *IntegrationTestContext) CleanupTestData() {
	// Clean up test users (cascade will handle sessions, etc.)
	tc.ExecuteSQL(`DELETE FROM auth.users WHERE email LIKE 'e2e-test-%'`)
	tc.ExecuteSQL(`DELETE FROM auth.users WHERE email LIKE 'test-%@example.com'`)
	tc.ExecuteSQL(`DELETE FROM auth.users WHERE email LIKE 'test-%@test.com'`)

	// Clean up dashboard test users
	tc.ExecuteSQL(`DELETE FROM dashboard.users WHERE email LIKE 'e2e-test-%'`)
	tc.ExecuteSQL(`DELETE FROM dashboard.users WHERE email LIKE 'test-%@example.com'`)
	tc.ExecuteSQL(`DELETE FROM dashboard.users WHERE email LIKE 'test-%@test.com'`)

	// Clean up password reset tokens (orphaned from user deletion)
	tc.ExecuteSQL(`DELETE FROM auth.password_reset_tokens WHERE user_id NOT IN (SELECT id FROM auth.users)`)

	// Clean up magic links
	tc.ExecuteSQL(`DELETE FROM auth.magic_links WHERE email LIKE 'test-%@example.com'`)
	tc.ExecuteSQL(`DELETE FROM auth.magic_links WHERE email LIKE 'test-%@test.com'`)
}
