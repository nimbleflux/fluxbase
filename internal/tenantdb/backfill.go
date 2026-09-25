package tenantdb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// tenantIDDedupTables lists tables that have a per-tenant uniqueness partition
// key (e.g. UNIQUE(email, tenant_id)). Before backfilling NULL tenant_id rows
// to the default tenant, only the oldest row per key is updated — see
// buildDedupBackfillUpdate.
var tenantIDDedupTables = map[string][]string{
	"auth.users":               {"email"},
	"functions.edge_functions": {"name, namespace"},
	"storage.buckets":          {"name"},
	"branching.branches":       {"name", "slug"},
	"branching.github_config":  {"repository"},
	"mcp.custom_tools":         {"name, namespace"},
	"mcp.custom_resources":     {"uri, namespace"},
}

// BackfillTenantIDToDefault assigns NULL tenant_id rows to the default tenant.
// This handles the upgrade path from pre-multi-tenant Fluxbase where all data
// was created without tenant context.
func BackfillTenantIDToDefault(pool *pgxpool.Pool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var defaultTenantID string
	err := pool.QueryRow(
		ctx,
		"SELECT id::text FROM platform.tenants WHERE is_default = true AND deleted_at IS NULL LIMIT 1",
	).Scan(&defaultTenantID)
	if err != nil {
		if err.Error() == "no rows in result set" || strings.Contains(err.Error(), "no rows") {
			log.Debug().Msg("No default tenant found, skipping tenant_id backfill")
			return nil
		}
		return fmt.Errorf("failed to get default tenant: %w", err)
	}

	tables := []string{
		"auth.users",
		"auth.client_keys",
		"auth.impersonation_sessions",
		"auth.webhooks",
		"auth.webhook_deliveries",
		"auth.webhook_events",
		"auth.saml_providers",
		"auth.sessions",
		"auth.oauth_links",
		"auth.oauth_tokens",
		"auth.mfa_factors",
		"auth.saml_sessions",
		"auth.magic_links",
		"auth.otp_codes",
		"auth.email_verification_tokens",
		"auth.password_reset_tokens",
		"auth.two_factor_setups",
		"auth.two_factor_recovery_attempts",
		"auth.oauth_logout_states",
		"auth.mcp_oauth_clients",
		"auth.mcp_oauth_codes",
		"auth.mcp_oauth_tokens",
		"auth.client_key_usage",
		"auth.service_key_revocations",
		"functions.edge_functions",
		"functions.edge_executions",
		"functions.edge_files",
		"functions.edge_triggers",
		"functions.secrets",
		"functions.secret_versions",
		"functions.shared_modules",
		"functions.function_dependencies",
		"jobs.functions",
		"jobs.function_files",
		"jobs.workers",
		"jobs.queue",
		"ai.knowledge_bases",
		"ai.knowledge_base_permissions",
		"ai.documents",
		"ai.document_permissions",
		"ai.chunks",
		"ai.entities",
		"ai.document_entities",
		"ai.entity_relationships",
		"ai.providers",
		"ai.chatbots",
		"ai.chatbot_knowledge_bases",
		"ai.conversations",
		"ai.messages",
		"ai.query_audit_log",
		"ai.retrieval_log",
		"ai.table_export_sync_configs",
		"ai.user_chatbot_usage",
		"ai.user_provider_preferences",
		"ai.user_quotas",
		"rpc.procedures",
		"rpc.executions",
		"realtime.schema_registry",
		"storage.buckets",
		"storage.objects",
		"storage.chunked_upload_sessions",
		"storage.object_permissions",
		"branching.branches",
		"branching.activity_log",
		"branching.branch_access",
		"branching.github_config",
		"branching.migration_history",
		"branching.seed_execution_log",
		"logging.entries",
		"logging.entries_ai",
		"logging.entries_custom",
		"logging.entries_execution",
		"logging.entries_http",
		"logging.entries_security",
		"logging.entries_system",
		"mcp.custom_resources",
		"mcp.custom_tools",
		"platform.invitation_tokens",
	}

	var totalBackfilled int
	for _, table := range tables {
		if dedupSets, needsDedup := tenantIDDedupTables[table]; needsDedup {
			for _, dedupCols := range dedupSets {
				// Least-destructive dedupe: instead of deleting rows, only the
				// oldest NULL-tenant row per partition key is backfilled — and
				// only when the key is not already occupied by an existing
				// default-tenant row. Sibling rows keep tenant_id NULL (they
				// stay preserved and invisible to tenant-scoped queries) rather
				// than being deleted, so pre-existing distinct users that share
				// an email (or other partition key) are never destroyed.
				dedupQuery := buildDedupBackfillUpdate(table, dedupCols)
				if dedupResult, err := pool.Exec(ctx, dedupQuery, defaultTenantID); err != nil {
					log.Warn().Err(err).Str("table", table).Str("cols", dedupCols).Msg("Failed to dedup NULL-tenant rows before backfill")
				} else if n := dedupResult.RowsAffected(); n > 0 {
					log.Info().Str("table", table).Str("cols", dedupCols).Int64("rows_backfilled", n).Msg("Backfilled oldest NULL-tenant row per partition key")
				}
			}
		}

		result, err := pool.Exec(
			ctx,
			fmt.Sprintf("UPDATE %s SET tenant_id = $1::uuid WHERE tenant_id IS NULL", table),
			defaultTenantID,
		)
		if err != nil {
			log.Warn().Err(err).Str("table", table).Msg("Failed to backfill tenant_id")
			continue
		}
		if n := result.RowsAffected(); n > 0 {
			log.Info().Str("table", table).Int64("rows", n).Msg("Backfilled tenant_id to default tenant")
			totalBackfilled += int(n)
		}
	}

	if totalBackfilled > 0 {
		log.Info().Int("total_rows", totalBackfilled).Msg("Tenant_id backfill complete")
	} else {
		log.Debug().Msg("No NULL tenant_id rows found to backfill")
	}

	return nil
}

// buildDedupBackfillUpdate builds the guarded backfill UPDATE for tables with
// a per-tenant partition key (e.g. auth.users keyed by email). The UPDATE:
//  1. only touches rows with tenant_id IS NULL,
//  2. skips a key entirely when a default-tenant row already occupies it
//     (avoids unique-convention conflicts without deleting the NULL rows),
//  3. updates only the oldest NULL-tenant row per key (lowest id), so the row
//     most likely to own related data survives; duplicate siblings keep
//     tenant_id NULL instead of being deleted.
//
// No DELETE is ever emitted.
func buildDedupBackfillUpdate(table string, cols string) string {
	keyOccupied := buildJoinCondition(cols, "e", "t") // e.col = t.col
	hasOlderSibling := buildJoinCondition(cols, "x", "t")
	return fmt.Sprintf(
		`UPDATE %s t SET tenant_id = $1::uuid
WHERE t.tenant_id IS NULL
  AND NOT EXISTS (SELECT 1 FROM %s e WHERE e.tenant_id = $1::uuid AND (%s))
  AND NOT EXISTS (SELECT 1 FROM %s x WHERE x.tenant_id IS NULL AND (%s) AND x.id < t.id)`,
		table, table, keyOccupied, table, hasOlderSibling,
	)
}

func buildJoinCondition(cols, leftAlias, rightAlias string) string {
	parts := strings.Split(cols, ",")
	conditions := make([]string, len(parts))
	for i, col := range parts {
		col = strings.TrimSpace(col)
		conditions[i] = fmt.Sprintf("%s.%s = %s.%s", leftAlias, col, rightAlias, col)
	}
	return strings.Join(conditions, " AND ")
}
