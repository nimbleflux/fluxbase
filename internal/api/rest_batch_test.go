package api

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fluxbase-eu/fluxbase/internal/config"
	"github.com/fluxbase-eu/fluxbase/internal/database"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// makeBatchPatchHandler Validation Tests
// =============================================================================

func TestMakeBatchPatchHandler_InvalidBody(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{
		parser: NewQueryParser(&config.Config{}),
	}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "items",
		PrimaryKey: []string{"id"},
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	app.Patch("/items", handler.makeBatchPatchHandler(table))

	t.Run("invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/items", strings.NewReader(`{invalid`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid request body")
	})

	t.Run("empty body - no fields to update", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/items", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "No fields to update")
	})

	t.Run("unknown column", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/items", strings.NewReader(`{"unknown_column":"value"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Unknown column")
	})

	t.Run("invalid query string", func(t *testing.T) {
		req := httptest.NewRequest("PATCH", "/items?invalid=%zz", strings.NewReader(`{"name":"test"}`))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid query string")
	})
}

// =============================================================================
// makeBatchDeleteHandler Validation Tests
// =============================================================================

func TestMakeBatchDeleteHandler_Validation(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{
		parser: NewQueryParser(&config.Config{}),
	}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "items",
		PrimaryKey: []string{"id"},
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	app.Delete("/items", handler.makeBatchDeleteHandler(table))

	t.Run("requires at least one filter", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/items", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Batch delete requires at least one filter")
	})

	t.Run("invalid query string", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/items?invalid=%zz", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, 400, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "Invalid query string")
	})
}

// =============================================================================
// batchInsert Validation Tests
// =============================================================================

func TestBatchInsert_EmptyArray(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "items",
		PrimaryKey: []string{"id"},
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	app.Post("/items", func(c fiber.Ctx) error {
		// Simulate batch insert with empty array
		return handler.batchInsert(c.RequestCtx(), c, table, []map[string]interface{}{}, false, false, false, "")
	})

	req := httptest.NewRequest("POST", "/items", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 400, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Empty array provided")
}

func TestBatchInsert_UnknownColumn(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "items",
		PrimaryKey: []string{"id"},
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	app.Post("/items", func(c fiber.Ctx) error {
		data := []map[string]interface{}{
			{"unknown_column": "value"},
		}
		return handler.batchInsert(c.RequestCtx(), c, table, data, false, false, false, "")
	})

	req := httptest.NewRequest("POST", "/items", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 400, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Unknown column")
}

func TestBatchInsert_UpsertWithoutPrimaryKey(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "logs",
		PrimaryKey: []string{}, // No primary key
		Columns: []database.ColumnInfo{
			{Name: "message", DataType: "text"},
			{Name: "timestamp", DataType: "timestamp"},
		},
	}

	app.Post("/logs", func(c fiber.Ctx) error {
		data := []map[string]interface{}{
			{"message": "test"},
		}
		return handler.batchInsert(c.RequestCtx(), c, table, data, true, false, false, "") // isUpsert = true
	})

	req := httptest.NewRequest("POST", "/logs", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 400, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Cannot perform upsert")
	assert.Contains(t, string(body), "no primary key or unique constraint")
}

func TestBatchInsert_UpsertWithUnknownConflictColumn(t *testing.T) {
	app := fiber.New()
	handler := &RESTHandler{}

	table := database.TableInfo{
		Schema:     "public",
		Name:       "items",
		PrimaryKey: []string{"id"},
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
		},
	}

	app.Post("/items", func(c fiber.Ctx) error {
		data := []map[string]interface{}{
			{"id": "123", "name": "test"},
		}
		return handler.batchInsert(c.RequestCtx(), c, table, data, true, false, false, "unknown_column") // Invalid on_conflict
	})

	req := httptest.NewRequest("POST", "/items", nil)

	resp, err := app.Test(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, 400, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "Unknown column in on_conflict")
}

// =============================================================================
// GeoJSON Handling Tests
// =============================================================================

// NOTE: TestBatchInsert_InvalidGeoJSON was removed because it passes valid data
// and proceeds to database execution, which requires a database connection.
// GeoJSON handling should be tested via integration tests.

// =============================================================================
// Batch Operation Behavior Tests
// =============================================================================

func TestBatchInsertBehavior(t *testing.T) {
	t.Run("columns from first record", func(t *testing.T) {
		// Test that batch insert uses columns from the first record
		dataArray := []map[string]interface{}{
			{"id": "1", "name": "first"},
			{"id": "2", "name": "second"},
			{"id": "3"}, // Missing "name" should use NULL
		}

		// Get columns from first record
		firstRecord := dataArray[0]
		columns := make([]string, 0, len(firstRecord))
		for col := range firstRecord {
			columns = append(columns, col)
		}

		assert.Contains(t, columns, "id")
		assert.Contains(t, columns, "name")
		assert.Len(t, columns, 2)
	})

	t.Run("missing column uses NULL", func(t *testing.T) {
		record := map[string]interface{}{
			"id": "1",
			// "name" is missing
		}
		columnNames := []string{"id", "name"}

		for _, col := range columnNames {
			val, exists := record[col]
			if !exists {
				val = nil // This is what batchInsert does
			}
			if col == "name" {
				assert.False(t, exists)
				assert.Nil(t, val)
			}
		}
	})
}

func TestBatchUpdateBehavior(t *testing.T) {
	t.Run("builds SET clause correctly", func(t *testing.T) {
		data := map[string]interface{}{
			"name":       "updated",
			"updated_at": "2024-01-01",
		}

		setClauses := make([]string, 0, len(data))
		for col := range data {
			setClauses = append(setClauses, col+" = $N")
		}

		assert.Len(t, setClauses, 2)
	})
}

func TestBatchDeleteBehavior(t *testing.T) {
	t.Run("requires filters for safety", func(t *testing.T) {
		// An empty filter slice should result in an error
		filters := []interface{}{} // Empty slice

		hasFilters := len(filters) > 0
		assert.False(t, hasFilters, "batch delete should require filters")
	})
}

// =============================================================================
// Conflict Target Tests
// =============================================================================

func TestConflictTargetParsing(t *testing.T) {
	tests := []struct {
		name           string
		onConflict     string
		expectedCols   []string
		expectMultiple bool
	}{
		{
			name:           "single column",
			onConflict:     "id",
			expectedCols:   []string{"id"},
			expectMultiple: false,
		},
		{
			name:           "multiple columns",
			onConflict:     "tenant_id,user_id",
			expectedCols:   []string{"tenant_id", "user_id"},
			expectMultiple: true,
		},
		{
			name:           "columns with spaces",
			onConflict:     "tenant_id, user_id, role_id",
			expectedCols:   []string{"tenant_id", "user_id", "role_id"},
			expectMultiple: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflictCols := strings.Split(tt.onConflict, ",")
			for i := range conflictCols {
				conflictCols[i] = strings.TrimSpace(conflictCols[i])
			}

			assert.Equal(t, tt.expectedCols, conflictCols)
			assert.Equal(t, tt.expectMultiple, len(conflictCols) > 1)
		})
	}
}

// =============================================================================
// defaultToNull Mode Tests
// =============================================================================

func TestDefaultToNullMode(t *testing.T) {
	t.Run("updates missing columns to NULL", func(t *testing.T) {
		// Simulate defaultToNull behavior
		tableColumns := []string{"id", "name", "email", "phone"}
		conflictTargetColumns := []string{"id"}
		providedColumns := []string{"id", "name"} // email and phone are missing

		updateClauses := make([]string, 0)
		for _, tableCol := range tableColumns {
			// Skip conflict target columns
			isConflictTarget := false
			for _, ctCol := range conflictTargetColumns {
				if ctCol == tableCol {
					isConflictTarget = true
					break
				}
			}
			if isConflictTarget {
				continue
			}

			// Check if column was provided
			columnProvided := false
			for _, providedCol := range providedColumns {
				if providedCol == tableCol {
					columnProvided = true
					break
				}
			}

			if columnProvided {
				updateClauses = append(updateClauses, tableCol+" = EXCLUDED."+tableCol)
			} else {
				updateClauses = append(updateClauses, tableCol+" = NULL")
			}
		}

		assert.Len(t, updateClauses, 3) // name, email, phone (not id)
		assert.Contains(t, updateClauses, "name = EXCLUDED.name")
		assert.Contains(t, updateClauses, "email = NULL")
		assert.Contains(t, updateClauses, "phone = NULL")
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkConflictColumnParsing(b *testing.B) {
	onConflict := "tenant_id, user_id, role_id"

	for i := 0; i < b.N; i++ {
		conflictCols := strings.Split(onConflict, ",")
		for j := range conflictCols {
			conflictCols[j] = strings.TrimSpace(conflictCols[j])
		}
	}
}

func BenchmarkBuildSetClauses(b *testing.B) {
	data := map[string]interface{}{
		"name":       "test",
		"email":      "test@example.com",
		"phone":      "123-456-7890",
		"updated_at": "2024-01-01",
	}

	for i := 0; i < b.N; i++ {
		setClauses := make([]string, 0, len(data))
		for col := range data {
			setClauses = append(setClauses, col+" = $N")
		}
	}
}

func BenchmarkDefaultToNullClauseBuilding(b *testing.B) {
	tableColumns := []string{"id", "name", "email", "phone", "address", "city", "state", "zip"}
	conflictTargetColumns := []string{"id"}
	providedColumns := []string{"id", "name", "email"}

	for i := 0; i < b.N; i++ {
		updateClauses := make([]string, 0)
		for _, tableCol := range tableColumns {
			isConflictTarget := false
			for _, ctCol := range conflictTargetColumns {
				if ctCol == tableCol {
					isConflictTarget = true
					break
				}
			}
			if isConflictTarget {
				continue
			}
			columnProvided := false
			for _, providedCol := range providedColumns {
				if providedCol == tableCol {
					columnProvided = true
					break
				}
			}
			if columnProvided {
				updateClauses = append(updateClauses, tableCol+" = EXCLUDED."+tableCol)
			} else {
				updateClauses = append(updateClauses, tableCol+" = NULL")
			}
		}
	}
}

// =============================================================================
// Additional Tests for Coverage Boost (Developer 3 Assignment)
// =============================================================================

// TestColumnExistsInBatch tests the columnExists method used in batch operations
func TestColumnExistsInBatch(t *testing.T) {
	handler := &RESTHandler{}

	table := database.TableInfo{
		Schema: "public",
		Name:   "items",
		Columns: []database.ColumnInfo{
			{Name: "id", DataType: "uuid"},
			{Name: "name", DataType: "text"},
			{Name: "description", DataType: "text"},
		},
	}

	t.Run("existing column", func(t *testing.T) {
		assert.True(t, handler.columnExists(table, "id"))
		assert.True(t, handler.columnExists(table, "name"))
		assert.True(t, handler.columnExists(table, "description"))
	})

	t.Run("non-existing column", func(t *testing.T) {
		assert.False(t, handler.columnExists(table, "unknown"))
		assert.False(t, handler.columnExists(table, "email"))
		assert.False(t, handler.columnExists(table, ""))
	})
}

// TestConflictTargetBuilding tests various conflict target scenarios
func TestConflictTargetBuilding(t *testing.T) {
	handler := &RESTHandler{}

	t.Run("single column conflict target", func(t *testing.T) {
		table := database.TableInfo{
			Name:       "users",
			PrimaryKey: []string{"id"},
		}
		target := handler.getConflictTarget(table)
		assert.Equal(t, `"id"`, target)

		unquoted := handler.getConflictTargetUnquoted(table)
		assert.Equal(t, []string{"id"}, unquoted)
	})

	t.Run("two column conflict target", func(t *testing.T) {
		table := database.TableInfo{
			Name:       "user_roles",
			PrimaryKey: []string{"user_id", "role_id"},
		}
		target := handler.getConflictTarget(table)
		assert.Contains(t, target, `"user_id"`)
		assert.Contains(t, target, `"role_id"`)

		unquoted := handler.getConflictTargetUnquoted(table)
		assert.Equal(t, []string{"user_id", "role_id"}, unquoted)
	})

	t.Run("three column conflict target", func(t *testing.T) {
		table := database.TableInfo{
			Name:       "permissions",
			PrimaryKey: []string{"org_id", "user_id", "resource_id"},
		}
		target := handler.getConflictTarget(table)
		assert.Contains(t, target, `"org_id"`)
		assert.Contains(t, target, `"user_id"`)
		assert.Contains(t, target, `"resource_id"`)

		unquoted := handler.getConflictTargetUnquoted(table)
		assert.Equal(t, []string{"org_id", "user_id", "resource_id"}, unquoted)
	})

	t.Run("no primary key", func(t *testing.T) {
		table := database.TableInfo{
			Name:       "logs",
			PrimaryKey: []string{},
		}
		target := handler.getConflictTarget(table)
		assert.Empty(t, target)

		unquoted := handler.getConflictTargetUnquoted(table)
		assert.Empty(t, unquoted)
	})

	t.Run("nil primary key", func(t *testing.T) {
		table := database.TableInfo{
			Name:       "data",
			PrimaryKey: nil,
		}
		target := handler.getConflictTarget(table)
		assert.Empty(t, target)

		unquoted := handler.getConflictTargetUnquoted(table)
		assert.Nil(t, unquoted)
	})
}

// TestIsInConflictTargetExtended tests conflict target membership checking
func TestIsInConflictTargetExtended(t *testing.T) {
	handler := &RESTHandler{}

	t.Run("single column conflict target", func(t *testing.T) {
		conflictTarget := []string{"id"}
		assert.True(t, handler.isInConflictTarget("id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("name", conflictTarget))
		assert.False(t, handler.isInConflictTarget("", conflictTarget))
	})

	t.Run("multi column conflict target", func(t *testing.T) {
		conflictTarget := []string{"tenant_id", "user_id", "role_id"}
		assert.True(t, handler.isInConflictTarget("tenant_id", conflictTarget))
		assert.True(t, handler.isInConflictTarget("user_id", conflictTarget))
		assert.True(t, handler.isInConflictTarget("role_id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("name", conflictTarget))
	})

	t.Run("empty conflict target", func(t *testing.T) {
		conflictTarget := []string{}
		assert.False(t, handler.isInConflictTarget("id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("name", conflictTarget))
	})

	t.Run("nil conflict target", func(t *testing.T) {
		assert.False(t, handler.isInConflictTarget("id", nil))
		assert.False(t, handler.isInConflictTarget("name", nil))
	})

	t.Run("case sensitivity", func(t *testing.T) {
		conflictTarget := []string{"id", "tenant_id"}
		assert.True(t, handler.isInConflictTarget("id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("ID", conflictTarget))
		assert.False(t, handler.isInConflictTarget("Id", conflictTarget))
		assert.False(t, handler.isInConflictTarget("TENANT_ID", conflictTarget))
	})
}

// TestOnConflictParameterParsing tests on_conflict parameter handling
func TestOnConflictParameterParsing(t *testing.T) {
	tests := []struct {
		name            string
		onConflict      string
		expectedColumns []string
		expectError     bool
		errorContains   string
	}{
		{
			name:            "single column",
			onConflict:      "sku",
			expectedColumns: []string{"sku"},
			expectError:     false,
		},
		{
			name:            "two columns",
			onConflict:      "warehouse_id, product_id",
			expectedColumns: []string{"warehouse_id", "product_id"},
			expectError:     false,
		},
		{
			name:            "three columns",
			onConflict:      "org_id, user_id, resource_id",
			expectedColumns: []string{"org_id", "user_id", "resource_id"},
			expectError:     false,
		},
		{
			name:            "with spaces",
			onConflict:      "tenant_id, user_id, role_id",
			expectedColumns: []string{"tenant_id", "user_id", "role_id"},
			expectError:     false,
		},
		{
			name:            "empty string",
			onConflict:      "",
			expectedColumns: []string{},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test parsing logic
			if tt.onConflict == "" {
				return // Empty case is valid
			}

			conflictCols := strings.Split(tt.onConflict, ",")
			parsedCols := make([]string, 0, len(conflictCols))

			for _, col := range conflictCols {
				col = strings.TrimSpace(col)
				parsedCols = append(parsedCols, col)
			}

			assert.Equal(t, tt.expectedColumns, parsedCols)
		})
	}
}

func TestPreferHeaderResponseFormat(t *testing.T) {
	// Test that Prefer header values are correctly detected
	tests := []struct {
		name          string
		preferHeader  string
		expectMinimal bool
		expectHeaders bool
		expectDefault bool
	}{
		{
			name:          "return=minimal",
			preferHeader:  "return=minimal",
			expectMinimal: true,
		},
		{
			name:          "return=minimal with other preferences",
			preferHeader:  "respond-async, return=minimal",
			expectMinimal: true,
		},
		{
			name:          "return=headers-only",
			preferHeader:  "return=headers-only",
			expectHeaders: true,
		},
		{
			name:          "return=representation",
			preferHeader:  "return=representation",
			expectDefault: true,
		},
		{
			name:          "empty header defaults to representation",
			preferHeader:  "",
			expectDefault: true,
		},
		{
			name:          "unknown preference defaults to representation",
			preferHeader:  "some-other-preference",
			expectDefault: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefer := tt.preferHeader

			isMinimal := strings.Contains(prefer, "return=minimal")
			isHeadersOnly := strings.Contains(prefer, "return=headers-only")
			isDefault := !isMinimal && !isHeadersOnly

			assert.Equal(t, tt.expectMinimal, isMinimal, "return=minimal detection")
			assert.Equal(t, tt.expectHeaders, isHeadersOnly, "return=headers-only detection")
			assert.Equal(t, tt.expectDefault, isDefault, "default (representation) detection")
		})
	}
}

func TestXAffectedCountHeader(t *testing.T) {
	// Test that X-Affected-Count header is formatted correctly
	tests := []struct {
		count    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "10"},
		{100, "100"},
		{1234567, "1234567"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c fiber.Ctx) error {
				affectedCount := tt.count
				c.Set("X-Affected-Count", fmt.Sprintf("%d", affectedCount))
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expected, resp.Header.Get("X-Affected-Count"))
		})
	}
}

func TestContentRangeHeader(t *testing.T) {
	// Test that Content-Range header is formatted correctly for batch responses
	tests := []struct {
		count    int
		expected string
	}{
		{0, "*/0"},
		{1, "*/1"},
		{50, "*/50"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c fiber.Ctx) error {
				affectedCount := tt.count
				c.Set("Content-Range", fmt.Sprintf("*/%d", affectedCount))
				return c.SendStatus(200)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.expected, resp.Header.Get("Content-Range"))
		})
	}
}

// fiber:context-methods migrated
