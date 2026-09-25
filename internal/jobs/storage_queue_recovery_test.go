package jobs

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// TestRecoverStaleJobsQuery_ScopesToDeadWorkers is a focused SQL-string test:
// recovery must only reset running jobs whose owning worker is provably dead
// (heartbeat-stale or stopped), never jobs of workers with fresh heartbeats —
// otherwise one instance's startup would reset another live instance's jobs.
func TestRecoverStaleJobsQuery_ScopesToDeadWorkers(t *testing.T) {
	q := recoverStaleJobsQuery

	// Only running jobs are reset
	assert.Contains(t, q, "status = $2")

	// Worker liveness is consulted: the job is reset only when unclaimed past
	// the grace period OR its worker's heartbeat is stale / worker stopped.
	assert.Contains(t, q, "worker_id IS NULL AND (started_at IS NULL OR started_at < NOW() - $3::INTERVAL)")
	assert.Contains(t, q, "last_heartbeat_at < NOW() - $3::INTERVAL")
	assert.Contains(t, q, "status = $4")

	// Jobs with worker_id set are never reset unconditionally: the unconditional
	// branch must not exist.
	assert.NotContains(t, q, "WHERE status = $2\n\t\t")

	// The reset payload itself
	assert.Contains(t, q, "SET status = $1")
	assert.Contains(t, q, "worker_id = NULL")
}

// TestManager_WorkerTimeoutDefault verifies the fallback worker timeout used
// for recovery when the config value is missing.
func TestManager_WorkerTimeoutDefault(t *testing.T) {
	m := &Manager{}
	assert.Equal(t, 45*time.Second, m.workerTimeout())

	m = &Manager{Config: &config.JobsConfig{WorkerTimeout: 30 * time.Second}}
	assert.Equal(t, 30*time.Second, m.workerTimeout())
}

// TestRecoverStaleJobsQuery_RequeuePayloadSanity guards the overall shape of
// the query (single UPDATE over jobs.queue).
func TestRecoverStaleJobsQuery_RequeuePayloadSanity(t *testing.T) {
	q := recoverStaleJobsQuery
	assert.True(t, strings.HasPrefix(strings.TrimSpace(q), "UPDATE jobs.queue"))
	assert.Equal(t, 1, strings.Count(q, "UPDATE"))
}
