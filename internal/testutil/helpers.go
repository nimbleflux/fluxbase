// Package testutil provides shared test helper functions for unit testing.
package testutil

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Database Helpers
// =============================================================================

// SetupTestDB creates a test database connection.
// This is a helper function that should be implemented based on your test database setup.
// For now, it returns nil - you should integrate with your existing test infrastructure.
//
// Example integration:
//
//	ctx := testutil.SetupTestDB(t)
//	defer testutil.CleanupTestDB(t, ctx)
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	// TODO: Integrate with existing test database infrastructure
	// This should connect to a test database and return the connection pool
	// Consider using the existing test.TestContext or IntegrationTestContext
	t.Helper()
	t.Skip("SetupTestDB: integrate with existing test infrastructure")
	return nil
}

// CleanupTestDB cleans up a test database connection.
func CleanupTestDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()
	if db != nil {
		db.Close()
	}
}

// CreateTestUser creates a test user in the database.
// Returns the user ID.
func CreateTestUser(t *testing.T, db *pgxpool.Pool, email string) string {
	t.Helper()
	// TODO: Implement user creation based on your auth schema
	// This should insert a user into auth.users and return the ID
	var userID string
	err := db.QueryRow(context.Background(),
		"INSERT INTO auth.users (email, encrypted_password, email_confirmed_at) VALUES ($1, 'hash', NOW()) RETURNING id",
		email).Scan(&userID)
	require.NoError(t, err)
	return userID
}

// CreateTestTable creates a test table with the specified columns.
//
// Example:
//
//	testutil.CreateTestTable(t, db, "public", "test_products",
//		[]testutil.Column{{Name: "id", Type: "serial PRIMARY KEY"}, {Name: "name", Type: "text"}})
func CreateTestTable(t *testing.T, db *pgxpool.Pool, schema, table string, columns []Column) {
	t.Helper()

	var columnDefs []string
	for _, col := range columns {
		columnDefs = append(columnDefs, fmt.Sprintf("%s %s", col.Name, col.Type))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s)", schema, table, commaSeparate(columnDefs))
	_, err := db.Exec(context.Background(), query)
	require.NoError(t, err, "Failed to create test table")
}

// Column represents a table column definition.
type Column struct {
	Name string
	Type string
}

// DropTestTable drops a test table.
func DropTestTable(t *testing.T, db *pgxpool.Pool, schema, table string) {
	t.Helper()
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s.%s CASCADE", schema, table)
	_, err := db.Exec(context.Background(), query)
	require.NoError(t, err, "Failed to drop test table")
}

// TruncateTable truncates a table (removes all data but keeps structure).
func TruncateTable(t *testing.T, db *pgxpool.Pool, schema, table string) {
	t.Helper()
	query := fmt.Sprintf("TRUNCATE TABLE %s.%s CASCADE", schema, table)
	_, err := db.Exec(context.Background(), query)
	require.NoError(t, err, "Failed to truncate table")
}

// =============================================================================
// Coverage Helpers
// =============================================================================

// AssertCoverage asserts that the coverage for a package meets the minimum percentage.
// This is a placeholder - you should integrate with your coverage tooling.
//
// Example:
//
//	testutil.AssertCoverage(t, "./internal/auth", 0.75)
func AssertCoverage(t *testing.T, packagePath string, minPercent float64) {
	t.Helper()
	// TODO: Integrate with coverage tooling (go test -coverprofile=coverage.out)
	// This should read coverage data and assert the minimum percentage
	t.Logf("AssertCoverage: package %s should have at least %.1f%% coverage", packagePath, minPercent*100)
}

// =============================================================================
// Async/Condition Helpers
// =============================================================================

// WaitForCondition polls until a condition is met or timeout expires.
//
// Example:
//
//	testutil.WaitForCondition(t, func() bool {
//	    return db.Ping(context.Background()) == nil
//	}, 5*time.Second, "database to become available")
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		<-ticker.C
	}

	require.Failf(t, "condition not met", "timed out waiting for %s (timeout: %v)", msg, timeout)
}

// =============================================================================
// Fiber HTTP Helpers
// =============================================================================

// MockFiberContext creates a mock Fiber context for testing handlers.
// This creates a minimal test context using httptest.
//
// Note: For most handler tests, use SetupTestServer() and make actual HTTP requests
// instead of mocking the context directly. This function is provided for special cases.
//
// Example:
//
//	app := testutil.SetupTestServer(t)
//	req := httptest.NewRequest("GET", "/api/v1/users", nil)
//	resp, err := app.Test(req)
//	require.NoError(t, err)
func MockFiberContext(method, path string, body io.Reader) *fiber.Ctx {
	// This is a placeholder - for actual usage, use SetupTestServer and httptest
	// Creating a Fiber context directly requires fasthttp which is complex
	// Most tests should use the SetupTestServer pattern instead
	return nil
}

// SetupTestServer creates a test Fiber server with common middleware.
//
// Example:
//
//	app := testutil.SetupTestServer(t)
//	// Register routes and test
func SetupTestServer(t *testing.T) *fiber.App {
	t.Helper()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	return app
}

// =============================================================================
// Assertion Helpers
// =============================================================================

// AssertJSONEqual asserts that two JSON strings represent equal objects.
func AssertJSONEqual(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedJSON, actualJSON interface{}
	err := json.Unmarshal([]byte(expected), &expectedJSON)
	require.NoError(t, err, "Failed to parse expected JSON")
	err = json.Unmarshal([]byte(actual), &actualJSON)
	require.NoError(t, err, "Failed to parse actual JSON")

	assert.Equal(t, expectedJSON, actualJSON, "JSON objects are not equal")
}

// AssertJSONContains asserts that the actual JSON contains all fields from expected JSON.
func AssertJSONContains(t *testing.T, expected, actual string) {
	t.Helper()

	var expectedJSON, actualJSON map[string]interface{}
	err := json.Unmarshal([]byte(expected), &expectedJSON)
	require.NoError(t, err, "Failed to parse expected JSON")
	err = json.Unmarshal([]byte(actual), &actualJSON)
	require.NoError(t, err, "Failed to parse actual JSON")

	for key, expectedValue := range expectedJSON {
		actualValue, exists := actualJSON[key]
		assert.True(t, exists, "Key %s not found in actual JSON", key)
		if exists {
			assert.Equal(t, expectedValue, actualValue, "Value for key %s does not match", key)
		}
	}
}

// =============================================================================
// Error Helpers
// =============================================================================

// AssertErrorIs asserts that an error is of a specific type.
func AssertErrorIs(t *testing.T, err error, target error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error %v, got nil", target)
	}
	assert.ErrorIs(t, err, target, "Error is not of expected type")
}

// AssertErrorContains asserts that an error's message contains a substring.
func AssertErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("Expected error containing %q, got nil", substring)
	}
	assert.Contains(t, err.Error(), substring, "Error message does not contain expected substring")
}

// =============================================================================
// Time Helpers
// =============================================================================

// MockTime is a helper for time-sensitive tests.
// It allows you to control the current time in tests.
type MockTime struct {
	mu     sync.Mutex
	now    time.Time
	frozen bool
	offset time.Duration
}

// NewMockTime creates a new mock time helper.
func NewMockTime() *MockTime {
	return &MockTime{
		now:    time.Now(),
		frozen: false,
	}
}

// Freeze freezes time at the current moment.
func (m *MockTime) Freeze() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = time.Now()
	m.frozen = true
}

// Unfreeze unfreezes time.
func (m *MockTime) Unfreeze() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.frozen = false
}

// Set sets the current time to a specific value.
func (m *MockTime) Set(t time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = t
	m.frozen = true
}

// Add adds a duration to the current time.
func (m *MockTime) Add(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.now = m.now.Add(d)
}

// Now returns the current (mock) time.
func (m *MockTime) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.frozen {
		return m.now
	}
	return time.Now().Add(m.offset)
}

// =============================================================================
// Buffer Helpers
// =============================================================================

// ReadBody reads the entire body into a string.
func ReadBody(t *testing.T, r io.ReadCloser) string {
	t.Helper()

	if r == nil {
		return ""
	}

	data, err := io.ReadAll(r)
	require.NoError(t, err, "Failed to read body")
	_ = r.Close() // Best effort close, body already read

	return string(data)
}

// =============================================================================
// SQL Helpers
// =============================================================================

// NullString creates a sql.NullString from a string pointer.
func NullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

// NullInt creates a sql.NullInt64 from an int pointer.
func NullInt(i *int) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(*i), Valid: true}
}

// StringPtr creates a string pointer.
func StringPtr(s string) *string {
	return &s
}

// IntPtr creates an int pointer.
func IntPtr(i int) *int {
	return &i
}

// =============================================================================
// Internal Helpers
// =============================================================================

func commaSeparate(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for _, item := range items[1:] {
		result += ", " + item
	}
	return result
}
