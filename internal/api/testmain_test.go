//go:build integration
// +build integration

package api_test

import (
	"os"
	"testing"

	test "github.com/nimbleflux/fluxbase/test"
)

// TestMain is the entry point for running tests in this package when using go test
// It ensures that the shared test context is properly cleaned up after all tests run,
// which prevents goroutine leaks from Fiber's middleware (CSRF, rate limiter, etc.)
// and other service goroutines (webhook, database, pub/sub, etc.)
func TestMain(m *testing.M) {
	// Initialize shared test context before running any tests
	test.InitSharedTestContext()

	// Run all tests
	code := m.Run()

	// Clean up shared test context after all tests complete
	// This shuts down the server and stops all background goroutines
	test.CleanupSharedTestContext()

	// Exit with the test result code
	os.Exit(code)
}
