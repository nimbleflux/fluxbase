// Package e2e contains transaction-based test examples.
// This file demonstrates the transaction pattern and its limitations.
package e2e

import (
	"context"
	"testing"

	"github.com/fluxbase-eu/fluxbase/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExampleDirectQueryWithTransaction demonstrates using transactions
// for direct database queries. This provides true isolation.
func TestExampleDirectQueryWithTransaction(t *testing.T) {
	// Each test gets isolated transaction
	txCtx := test.BeginTestTx(t)
	defer txCtx.Close()

	// Direct database query within transaction
	userID := uuid.New()
	_, err := txCtx.Tx().Exec(context.Background(), `
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, $2, $3, true, NOW())
	`, userID, "tx-direct@example.com", "hashed_password")
	require.NoError(t, err, "Insert should succeed within transaction")

	// Verify user exists within transaction
	var count int
	err = txCtx.Tx().QueryRow(context.Background(), `
		SELECT COUNT(*) FROM auth.users WHERE id = $1
	`, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "User should exist within transaction")

	// Transaction automatically rolled back - user doesn't exist after test
}

// TestExampleTransactionIsolationMultipleRuns demonstrates that direct
// database queries can use the same data across test runs.
func TestExampleTransactionIsolationMultipleRuns(t *testing.T) {
	// This test uses the same email as a previous test would
	// With transaction isolation, this always works
	txCtx := test.BeginTestTx(t)
	defer txCtx.Close()

	userID := uuid.New()
	_, err := txCtx.Tx().Exec(context.Background(), `
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, $2, $3, true, NOW())
	`, userID, "tx-test@example.com", "hashed_password")
	require.NoError(t, err)

	// Verify it exists
	var count int
	err = txCtx.Tx().QueryRow(context.Background(), `
		SELECT COUNT(*) FROM auth.users WHERE id = $1
	`, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

// TestExampleTransactionCommit demonstrates explicit commit (rarely needed).
// This is only useful when testing transaction commit behavior itself.
func TestExampleTransactionCommit(t *testing.T) {
	txCtx := test.BeginTestTx(t)
	defer txCtx.Close()

	userID := uuid.New()
	uniqueEmail := "commit-test-" + userID.String() + "@example.com"
	_, err := txCtx.Tx().Exec(context.Background(), `
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, $2, $3, true, NOW())
	`, userID, uniqueEmail, "hashed_password")
	require.NoError(t, err)

	// Explicitly commit transaction
	err = txCtx.Commit()
	require.NoError(t, err, "Commit should succeed")

	// Note: After commit, the data persists in the database
	// This is rarely needed in tests - most tests should rely on automatic rollback
}

// TestExampleOldPatternStillWorks demonstrates that the old pattern
// still works for HTTP API testing.
func TestExampleOldPatternStillWorks(t *testing.T) {
	// Old pattern - works for HTTP API testing
	tc := test.NewTestContext(t)
	defer tc.Close()

	// Use unique email to avoid conflicts with other tests
	uniqueEmail := "old-pattern-" + uuid.New().String() + "@example.com"

	resp := tc.NewRequest("POST", "/api/v1/auth/signup").
		WithBody(map[string]string{
			"email":            uniqueEmail,
			"password":         "TestPassword123!",
			"password_confirm": "TestPassword123!",
		}).
		Send()

	// Note: For HTTP API testing, use regular TestContext
	// The transaction wrapper only works for direct DB queries
	assert.Contains(t, []int{201, 200}, resp.Status())
}

// TestExampleTransactionLimitations documents the current limitations.
// TODO: To enable transaction isolation for HTTP API tests, we need to:
//  1. Pass the transaction connection to the server
//  2. Have the server use test's transaction for queries
//  3. This requires Phase 2 of the plan (dependency injection)
func TestExampleTransactionLimitations(t *testing.T) {
	txCtx := test.BeginTestTx(t)
	defer txCtx.Close()

	// Direct DB query: uses transaction ✓
	userID := uuid.New()
	_, err := txCtx.Tx().Exec(context.Background(), `
		INSERT INTO auth.users (id, email, password_hash, email_verified, created_at)
		VALUES ($1, $2, $3, true, NOW())
	`, userID, "limitation@example.com", "hashed_password")
	require.NoError(t, err)

	// Verify within transaction: works ✓
	var count int
	err = txCtx.Tx().QueryRow(context.Background(), `
		SELECT COUNT(*) FROM auth.users WHERE id = $1
	`, userID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// HTTP request: does NOT use transaction ✗
	// The server has its own connection pool, so HTTP requests
	// don't see the test's transaction changes.
	// This is documented behavior - see Phase 2 for the solution.
}

// TestExampleDependencyInjection demonstrates using test-specific dependencies.
// This is Phase 2 of the test isolation plan - each test gets its own rate limiter
// and pub/sub instances instead of using global singletons.
func TestExampleDependencyInjection(t *testing.T) {
	// Create test-specific in-memory dependencies
	rateLimiter, pubSub := test.NewInMemoryDependencies()

	// Create test context with custom dependencies
	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// CRITICAL: Clean up any data from previous tests that might interfere
	// This prevents test pollution when running in the full test suite
	tc.ExecuteSQLAsSuperuser(`
		DELETE FROM auth.users WHERE email LIKE '%di-test-%' OR email LIKE '%di-multi-%';
	`)

	// This test now has its own isolated rate limiter and pub/sub
	// Changes won't affect other tests using global singletons
	uniqueEmail := "di-test-" + uuid.New().String() + "@example.com"

	resp := tc.NewRequest("POST", "/api/v1/auth/signup").
		WithBody(map[string]string{
			"email":            uniqueEmail,
			"password":         "TestPassword123!",
			"password_confirm": "TestPassword123!",
		}).
		Send()

	assert.Contains(t, []int{201, 200}, resp.Status())
}

// TestExampleDependencyInjectionMultipleRuns demonstrates that tests with
// custom dependencies can run multiple times without polluting each other.
func TestExampleDependencyInjectionMultipleRuns(t *testing.T) {
	// Each test run gets fresh dependencies
	rateLimiter, pubSub := test.NewInMemoryDependencies()

	tc := test.NewTestContextWithOptions(t, test.TestContextOptions{
		RateLimiter: rateLimiter,
		PubSub:      pubSub,
	})
	defer tc.Close()

	// CRITICAL: Clean up any data from previous tests that might interfere
	// This prevents test pollution when running in the full test suite
	tc.ExecuteSQLAsSuperuser(`
		DELETE FROM auth.users WHERE email LIKE '%di-test-%' OR email LIKE '%di-multi-%';
	`)

	// This test has completely isolated rate limiter and pub/sub
	// No state pollution from previous test runs
	uniqueEmail := "di-multi-" + uuid.New().String() + "@example.com"

	resp := tc.NewRequest("POST", "/api/v1/auth/signup").
		WithBody(map[string]string{
			"email":            uniqueEmail,
			"password":         "TestPassword123!",
			"password_confirm": "TestPassword123!",
		}).
		Send()

	assert.Contains(t, []int{201, 200}, resp.Status())
}
