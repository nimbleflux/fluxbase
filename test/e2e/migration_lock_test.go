package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nimbleflux/fluxbase/test"
)

const testMigrationLockID int64 = 0x466C7578_00000004

func TestConcurrentMigrationLock(t *testing.T) {
	tc := test.NewTestContext(t)
	defer tc.Close()

	pool := tc.DB.Pool()
	require.NotNil(t, pool, "test context must have a database pool")

	ctx := t.Context()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	var acquired bool
	err = tx.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", testMigrationLockID).Scan(&acquired)
	require.NoError(t, err)
	require.True(t, acquired, "session 1 should acquire the lock")

	var blocked bool
	err = pool.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", testMigrationLockID).Scan(&blocked)
	require.NoError(t, err)
	assert.False(t, blocked, "session 2 should NOT acquire the lock while session 1 holds it")

	require.NoError(t, tx.Rollback(ctx))

	err = pool.QueryRow(ctx, "SELECT pg_try_advisory_xact_lock($1)", testMigrationLockID).Scan(&acquired)
	require.NoError(t, err)
	assert.True(t, acquired, "lock should be available after session 1 transaction ends")
}
