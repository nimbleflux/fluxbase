--
-- MULTI-TENANCY: REMOVE TENANT_ID COLUMNS
-- Rollback migration 115
--

-- AI Schema
ALTER TABLE ai.chunks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.messages DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.entities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.entity_relationships DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.document_entities DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.query_audit_log DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.retrieval_log DROP COLUMN IF EXISTS tenant_id;

-- Functions Schema
ALTER TABLE functions.edge_executions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.edge_files DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.function_dependencies DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.secret_versions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.shared_modules DROP COLUMN IF EXISTS tenant_id;

-- Jobs Schema
ALTER TABLE jobs.function_files DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE jobs.workers DROP COLUMN IF EXISTS tenant_id;

-- Auth Schema
ALTER TABLE auth.sessions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.mfa_factors DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.oauth_links DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.impersonation_sessions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.webhook_deliveries DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.webhook_events DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.webhook_monitored_tables DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.client_key_usage DROP COLUMN IF EXISTS tenant_id;

-- Storage Schema
ALTER TABLE storage.chunked_upload_sessions DROP COLUMN IF EXISTS tenant_id;

-- Logging Schema (only drop index, column is inherited from parent)
DROP INDEX IF EXISTS logging.idx_logging_entries_tenant_id;

-- Branching Schema
ALTER TABLE branching.branches DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE branching.branch_access DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE branching.github_config DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE branching.activity_log DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE branching.migration_history DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE branching.seed_execution_log DROP COLUMN IF EXISTS tenant_id;

-- RPC Schema
ALTER TABLE rpc.executions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE rpc.procedures DROP COLUMN IF EXISTS tenant_id;

-- Realtime Schema
ALTER TABLE realtime.schema_registry DROP COLUMN IF EXISTS tenant_id;
