package tenantdb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildJoinCondition(t *testing.T) {
	tests := []struct {
		name       string
		cols       string
		leftAlias  string
		rightAlias string
		expected   string
	}{
		{
			name:       "single column",
			cols:       "email",
			leftAlias:  "n",
			rightAlias: "e",
			expected:   "n.email = e.email",
		},
		{
			name:       "multiple columns",
			cols:       "name, namespace",
			leftAlias:  "n",
			rightAlias: "e",
			expected:   "n.name = e.name AND n.namespace = e.namespace",
		},
		{
			name:       "spaces around columns",
			cols:       " name , namespace ",
			leftAlias:  "n",
			rightAlias: "e",
			expected:   "n.name = e.name AND n.namespace = e.namespace",
		},
		{
			name:       "three columns",
			cols:       "a, b, c",
			leftAlias:  "n",
			rightAlias: "e",
			expected:   "n.a = e.a AND n.b = e.b AND n.c = e.c",
		},
		{
			name:       "single column with spaces",
			cols:       " col1 ",
			leftAlias:  "n",
			rightAlias: "e",
			expected:   "n.col1 = e.col1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildJoinCondition(tt.cols, tt.leftAlias, tt.rightAlias)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBackfillTenantIDToDefault_SkipsWithoutDB(t *testing.T) {
	t.Skip("BackfillTenantIDToDefault requires a real database pool; tested in integration/E2E")
}

// TestBuildDedupBackfillUpdate verifies the least-destructive backfill query:
// it must never DELETE, only backfill the oldest NULL-tenant row per partition
// key, and skip keys already occupied by an existing default-tenant row.
func TestBuildDedupBackfillUpdate(t *testing.T) {
	t.Run("auth.users single column key", func(t *testing.T) {
		q := buildDedupBackfillUpdate("auth.users", "email")

		assert.NotContains(t, q, "DELETE", "backfill must never delete rows")
		assert.Contains(t, q, "UPDATE auth.users t SET tenant_id = $1::uuid")
		assert.Contains(t, q, "WHERE t.tenant_id IS NULL")
		// Conflict guard: key already occupied by a default-tenant row -> skip.
		assert.Contains(t, q, "NOT EXISTS (SELECT 1 FROM auth.users e WHERE e.tenant_id = $1::uuid AND (e.email = t.email))")
		// Oldest-row guard: only the lowest-id NULL-tenant sibling is updated.
		assert.Contains(t, q, "NOT EXISTS (SELECT 1 FROM auth.users x WHERE x.tenant_id IS NULL AND (x.email = t.email) AND x.id < t.id)")
	})

	t.Run("multi column key", func(t *testing.T) {
		q := buildDedupBackfillUpdate("functions.edge_functions", "name, namespace")

		assert.NotContains(t, q, "DELETE")
		assert.Contains(t, q, "e.name = t.name AND e.namespace = t.namespace")
		assert.Contains(t, q, "x.name = t.name AND x.namespace = t.namespace")
	})

	t.Run("all dedup tables produce delete-free guarded updates", func(t *testing.T) {
		for table, keySets := range tenantIDDedupTables {
			for _, cols := range keySets {
				q := buildDedupBackfillUpdate(table, cols)
				assert.NotContains(t, q, "DELETE", "table %s must not be deleted", table)
				assert.Contains(t, q, "tenant_id IS NULL", "table %s must only touch NULL-tenant rows", table)
				assert.Equal(t, 2, strings.Count(q, "NOT EXISTS"),
					"table %s must have conflict and oldest-row guards", table)
			}
		}
	})
}
