//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/fluxbase-eu/fluxbase/test"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExecutor_Query_Integration tests basic query execution.
func TestExecutor_Query_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("executes simple SELECT query", func(t *testing.T) {
		rows, err := exec.Query(context.Background(), "SELECT 1 AS num")
		require.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())

		var num int
		err = rows.Scan(&num)
		require.NoError(t, err)
		assert.Equal(t, 1, num)
	})

	t.Run("executes query with parameters", func(t *testing.T) {
		rows, err := exec.Query(context.Background(), "SELECT $1::text AS message", "hello")
		require.NoError(t, err)
		defer rows.Close()

		assert.True(t, rows.Next())

		var message string
		err = rows.Scan(&message)
		require.NoError(t, err)
		assert.Equal(t, "hello", message)
	})

	t.Run("returns empty result set when no rows", func(t *testing.T) {
		// Create a temporary table
		_, err := exec.Exec(context.Background(), "CREATE TEMP TABLE temp_test (id INT)")
		require.NoError(t, err)

		rows, err := exec.Query(context.Background(), "SELECT * FROM temp_test")
		require.NoError(t, err)
		defer rows.Close()

		assert.False(t, rows.Next())
	})
}

// TestExecutor_QueryRow_Integration tests single row queries.
func TestExecutor_QueryRow_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("retrieves single row", func(t *testing.T) {
		row := exec.QueryRow(context.Background(), "SELECT 2+2 AS result")

		var result int
		err := row.Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, 4, result)
	})

	t.Run("returns error when no rows", func(t *testing.T) {
		row := exec.QueryRow(context.Background(), "SELECT 1 WHERE FALSE")

		var result int
		err := row.Scan(&result)
		assert.Equal(t, pgx.ErrNoRows, err)
	})
}

// TestExecutor_Exec_Integration tests statements that don't return rows.
func TestExecutor_Exec_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("creates temporary table", func(t *testing.T) {
		tag, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_exec (id SERIAL PRIMARY KEY, name TEXT)")
		require.NoError(t, err)
		assert.Equal(t, int64(0), tag.RowsAffected())
	})

	t.Run("inserts data and returns row count", func(t *testing.T) {
		// First create the table
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_insert (id SERIAL PRIMARY KEY, value TEXT)")
		require.NoError(t, err)

		// Insert a row
		tag, err := exec.Exec(context.Background(),
			"INSERT INTO test_insert (value) VALUES ($1)", "test")
		require.NoError(t, err)
		assert.Equal(t, int64(1), tag.RowsAffected())
	})

	t.Run("updates multiple rows", func(t *testing.T) {
		// Create table and insert data
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_update (id SERIAL PRIMARY KEY, status TEXT)")
		require.NoError(t, err)

		_, err = exec.Exec(context.Background(),
			"INSERT INTO test_update (status) VALUES ('pending'), ('pending'), ('done')")
		require.NoError(t, err)

		// Update all pending to complete
		tag, err := exec.Exec(context.Background(),
			"UPDATE test_update SET status = 'complete' WHERE status = 'pending'")
		require.NoError(t, err)
		assert.Equal(t, int64(2), tag.RowsAffected())
	})

	t.Run("deletes data", func(t *testing.T) {
		// Create table and insert data
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_delete (id SERIAL PRIMARY KEY)")
		require.NoError(t, err)

		_, err = exec.Exec(context.Background(),
			"INSERT INTO test_delete DEFAULT VALUES")
		require.NoError(t, err)

		// Delete the row
		tag, err := exec.Exec(context.Background(),
			"DELETE FROM test_delete")
		require.NoError(t, err)
		assert.Equal(t, int64(1), tag.RowsAffected())
	})
}

// TestExecutor_Transaction_Integration tests transaction management.
func TestExecutor_Transaction_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("commits transaction", func(t *testing.T) {
		// Create table
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_tx_commit (id SERIAL PRIMARY KEY, value TEXT)")
		require.NoError(t, err)

		// Start transaction
		tx, err := exec.BeginTx(context.Background())
		require.NoError(t, err)

		// Insert data within transaction
		_, err = tx.Exec(context.Background(),
			"INSERT INTO test_tx_commit (value) VALUES ($1)", "committed")
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit(context.Background())
		require.NoError(t, err)

		// Verify data was committed
		row := exec.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM test_tx_commit WHERE value = $1", "committed")
		var count int
		err = row.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})

	t.Run("rolls back transaction", func(t *testing.T) {
		// Create table
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_tx_rollback (id SERIAL PRIMARY KEY, value TEXT)")
		require.NoError(t, err)

		// Start transaction
		tx, err := exec.BeginTx(context.Background())
		require.NoError(t, err)

		// Insert data within transaction
		_, err = tx.Exec(context.Background(),
			"INSERT INTO test_tx_rollback (value) VALUES ($1)", "rolled_back")
		require.NoError(t, err)

		// Rollback transaction
		err = tx.Rollback(context.Background())
		require.NoError(t, err)

		// Verify data was not saved
		row := exec.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM test_tx_rollback WHERE value = $1", "rolled_back")
		var count int
		err = row.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})

	t.Run("handles rollback on error", func(t *testing.T) {
		// Create table
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_tx_error (id SERIAL PRIMARY KEY, value TEXT NOT NULL)")
		require.NoError(t, err)

		// Start transaction
		tx, err := exec.BeginTx(context.Background())
		require.NoError(t, err)

		// Insert valid data
		_, err = tx.Exec(context.Background(),
			"INSERT INTO test_tx_error (value) VALUES ($1)", "valid")
		require.NoError(t, err)

		// Try to insert invalid data (will fail due to NOT NULL constraint)
		_, err = tx.Exec(context.Background(),
			"INSERT INTO test_tx_error (value) VALUES (NULL)")
		assert.Error(t, err)

		// Rollback
		_ = tx.Rollback(context.Background())

		// Verify first insert was also rolled back
		row := exec.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM test_tx_error")
		var count int
		err = row.Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

// TestExecutor_ErrorHandling_Integration tests error scenarios.
func TestExecutor_ErrorHandling_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("returns error for invalid SQL", func(t *testing.T) {
		_, err := exec.Query(context.Background(), "INVALID SQL STATEMENT")
		assert.Error(t, err)
	})

	t.Run("returns error for constraint violation", func(t *testing.T) {
		// Create unique constraint table
		_, err := exec.Exec(context.Background(),
			"CREATE TEMP TABLE test_constraint (id INT PRIMARY KEY)")
		require.NoError(t, err)

		// Insert duplicate key
		_, err = exec.Exec(context.Background(),
			"INSERT INTO test_constraint (id) VALUES (1), (1)")
		assert.Error(t, err)
	})

	t.Run("returns error for non-existent table", func(t *testing.T) {
		_, err := exec.Query(context.Background(), "SELECT * FROM nonexistent_table_xyz")
		assert.Error(t, err)
	})
}

// TestExecutor_Pool_Integration tests pool methods.
func TestExecutor_Pool_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	t.Run("returns pool instance", func(t *testing.T) {
		pool := testCtx.DB.Pool()
		assert.NotNil(t, pool)
	})

	t.Run("health check succeeds", func(t *testing.T) {
		err := testCtx.DB.Health(context.Background())
		assert.NoError(t, err)
	})
}

// TestExecutor_DataTypes_Integration tests handling various PostgreSQL data types.
func TestExecutor_DataTypes_Integration(t *testing.T) {
	testCtx := test.NewTestContext(t)
	defer testCtx.Close()

	exec := testCtx.DB

	t.Run("handles JSON data", func(t *testing.T) {
		row := exec.QueryRow(context.Background(),
			`SELECT '{"key": "value"}'::json AS data`)

		var data map[string]interface{}
		err := row.Scan(&data)
		require.NoError(t, err)
		assert.Equal(t, "value", data["key"])
	})

	t.Run("handles JSONB data", func(t *testing.T) {
		row := exec.QueryRow(context.Background(),
			`SELECT '{"key": "value"}'::jsonb AS data`)

		var data map[string]interface{}
		err := row.Scan(&data)
		require.NoError(t, err)
		assert.Equal(t, "value", data["key"])
	})

	t.Run("handles UUID", func(t *testing.T) {
		row := exec.QueryRow(context.Background(),
			"SELECT gen_random_uuid() AS id")

		var id string
		err := row.Scan(&id)
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("handles arrays", func(t *testing.T) {
		row := exec.QueryRow(context.Background(),
			"SELECT ARRAY[1, 2, 3]::INT[] AS numbers")

		var numbers []int32
		err := row.Scan(&numbers)
		require.NoError(t, err)
		assert.Len(t, numbers, 3)
		assert.Equal(t, int32(1), numbers[0])
	})
}
