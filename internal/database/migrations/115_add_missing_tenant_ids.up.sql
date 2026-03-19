--
-- MULTI-TENANCY: ADD MISSING TENANT_ID COLUMNS
-- Adds tenant_id column to all remaining tenant-scoped tables
--

-- ============================================
-- AI SCHEMA
-- ============================================

-- Document chunks (inherit from documents)
ALTER TABLE ai.chunks ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_chunks_tenant_id ON ai.chunks(tenant_id);
COMMENT ON COLUMN ai.chunks.tenant_id IS 'Tenant this chunk belongs to (inherited from document)';

-- Chat messages (inherit from conversations)
ALTER TABLE ai.messages ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_messages_tenant_id ON ai.messages(tenant_id);
COMMENT ON COLUMN ai.messages.tenant_id IS 'Tenant this message belongs to (inherited from conversation)';

-- Extracted entities
ALTER TABLE ai.entities ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_entities_tenant_id ON ai.entities(tenant_id);
COMMENT ON COLUMN ai.entities.tenant_id IS 'Tenant this entity belongs to';

-- Entity relationships
ALTER TABLE ai.entity_relationships ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_entity_relationships_tenant_id ON ai.entity_relationships(tenant_id);
COMMENT ON COLUMN ai.entity_relationships.tenant_id IS 'Tenant this relationship belongs to';

-- Document entity links
ALTER TABLE ai.document_entities ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_document_entities_tenant_id ON ai.document_entities(tenant_id);
COMMENT ON COLUMN ai.document_entities.tenant_id IS 'Tenant this document-entity link belongs to';

-- Query audit log
ALTER TABLE ai.query_audit_log ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_query_audit_log_tenant_id ON ai.query_audit_log(tenant_id);
COMMENT ON COLUMN ai.query_audit_log.tenant_id IS 'Tenant for this audit log entry';

-- Retrieval log
ALTER TABLE ai.retrieval_log ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_retrieval_log_tenant_id ON ai.retrieval_log(tenant_id);
COMMENT ON COLUMN ai.retrieval_log.tenant_id IS 'Tenant for this retrieval log entry';

-- ============================================
-- FUNCTIONS SCHEMA
-- ============================================

-- Edge function executions (inherit from edge_functions)
ALTER TABLE functions.edge_executions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_edge_executions_tenant_id ON functions.edge_executions(tenant_id);
COMMENT ON COLUMN functions.edge_executions.tenant_id IS 'Tenant this execution belongs to (inherited from function)';

-- Edge function files (inherit from edge_functions)
ALTER TABLE functions.edge_files ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_edge_files_tenant_id ON functions.edge_files(tenant_id);
COMMENT ON COLUMN functions.edge_files.tenant_id IS 'Tenant this file belongs to (inherited from function)';

-- Function dependencies (inherit from edge_functions)
ALTER TABLE functions.function_dependencies ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_function_dependencies_tenant_id ON functions.function_dependencies(tenant_id);
COMMENT ON COLUMN functions.function_dependencies.tenant_id IS 'Tenant this dependency belongs to (inherited from function)';

-- Secret versions (inherit from secrets)
ALTER TABLE functions.secret_versions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_secret_versions_tenant_id ON functions.secret_versions(tenant_id);
COMMENT ON COLUMN functions.secret_versions.tenant_id IS 'Tenant this secret version belongs to (inherited from secret)';

-- Shared modules
ALTER TABLE functions.shared_modules ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_shared_modules_tenant_id ON functions.shared_modules(tenant_id);
COMMENT ON COLUMN functions.shared_modules.tenant_id IS 'Tenant this shared module belongs to';

-- ============================================
-- JOBS SCHEMA
-- ============================================

-- Job function files (inherit from jobs.functions)
ALTER TABLE jobs.function_files ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_jobs_function_files_tenant_id ON jobs.function_files(tenant_id);
COMMENT ON COLUMN jobs.function_files.tenant_id IS 'Tenant this file belongs to (inherited from job)';

-- Workers
ALTER TABLE jobs.workers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_jobs_workers_tenant_id ON jobs.workers(tenant_id);
COMMENT ON COLUMN jobs.workers.tenant_id IS 'Tenant this worker belongs to';

-- ============================================
-- AUTH SCHEMA
-- ============================================

-- Sessions (inherit tenant from user)
ALTER TABLE auth.sessions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_auth_sessions_tenant_id ON auth.sessions(tenant_id);
COMMENT ON COLUMN auth.sessions.tenant_id IS 'Tenant this session belongs to (inherited from user)';

-- MFA factors (inherit tenant from user)
ALTER TABLE auth.mfa_factors ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_mfa_factors_tenant_id ON auth.mfa_factors(tenant_id);
COMMENT ON COLUMN auth.mfa_factors.tenant_id IS 'Tenant this MFA factor belongs to (inherited from user)';

-- OAuth links (inherit tenant from user)
ALTER TABLE auth.oauth_links ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_oauth_links_tenant_id ON auth.oauth_links(tenant_id);
COMMENT ON COLUMN auth.oauth_links.tenant_id IS 'Tenant this OAuth link belongs to (inherited from user)';

-- Impersonation sessions
ALTER TABLE auth.impersonation_sessions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_impersonation_sessions_tenant_id ON auth.impersonation_sessions(tenant_id);
COMMENT ON COLUMN auth.impersonation_sessions.tenant_id IS 'Tenant this impersonation session belongs to';

-- Webhook deliveries (inherit from webhooks)
ALTER TABLE auth.webhook_deliveries ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_webhook_deliveries_tenant_id ON auth.webhook_deliveries(tenant_id);
COMMENT ON COLUMN auth.webhook_deliveries.tenant_id IS 'Tenant this delivery belongs to (inherited from webhook)';

-- Webhook Events (inherit from webhooks)
ALTER TABLE auth.webhook_events ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_webhook_events_tenant_id ON auth.webhook_events(tenant_id);
COMMENT ON COLUMN auth.webhook_events.tenant_id IS 'Tenant this event belongs to (inherited from webhook)';

-- Webhook monitored tables (inherit from webhooks)
ALTER TABLE auth.webhook_monitored_tables ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_webhook_monitored_tables_tenant_id ON auth.webhook_monitored_tables(tenant_id);
COMMENT ON COLUMN auth.webhook_monitored_tables.tenant_id IS 'Tenant this monitored table belongs to (inherited from webhook)';

-- Client key usage (inherit from client_keys)
ALTER TABLE auth.client_key_usage ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_client_key_usage_tenant_id ON auth.client_key_usage(tenant_id);
COMMENT ON COLUMN auth.client_key_usage.tenant_id IS 'Tenant this usage record belongs to (inherited from client key)';

-- ============================================
-- STORAGE SCHEMA
-- ============================================

-- Chunked upload sessions (inherit from bucket)
ALTER TABLE storage.chunked_upload_sessions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_storage_chunked_upload_sessions_tenant_id ON storage.chunked_upload_sessions(tenant_id);
COMMENT ON COLUMN storage.chunked_upload_sessions.tenant_id IS 'Tenant this upload session belongs to (inherited from bucket)';

-- ============================================
-- LOGGING SCHEMA
-- ============================================

-- Main log entries (parent partitioned table)
-- Note: logging.entries is a partitioned table. The tenant_id column already exists
-- and is inherited by all partitions (entries_ai, entries_custom, entries_execution,
-- entries_http, entries_security, entries_system). We only need to add the index.
CREATE INDEX IF NOT EXISTS idx_logging_entries_tenant_id ON logging.entries(tenant_id);
COMMENT ON COLUMN logging.entries.tenant_id IS 'Tenant this log entry belongs to';

-- ============================================
-- BRANCHING SCHEMA
-- ============================================

-- Database branches
ALTER TABLE branching.branches ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_branches_tenant_id ON branching.branches(tenant_id);
COMMENT ON COLUMN branching.branches.tenant_id IS 'Tenant this branch belongs to';

-- Branch access control
ALTER TABLE branching.branch_access ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_branch_access_tenant_id ON branching.branch_access(tenant_id);
COMMENT ON COLUMN branching.branch_access.tenant_id IS 'Tenant this branch access belongs to';

-- GitHub configuration
ALTER TABLE branching.github_config ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_github_config_tenant_id ON branching.github_config(tenant_id);
COMMENT ON COLUMN branching.github_config.tenant_id IS 'Tenant this GitHub config belongs to';

-- Branch activity log
ALTER TABLE branching.activity_log ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_activity_log_tenant_id ON branching.activity_log(tenant_id);
COMMENT ON COLUMN branching.activity_log.tenant_id IS 'Tenant this activity log belongs to';

-- Branch migration history
ALTER TABLE branching.migration_history ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_migration_history_tenant_id ON branching.migration_history(tenant_id);
COMMENT ON COLUMN branching.migration_history.tenant_id IS 'Tenant this migration history belongs to';

-- Seed execution log
ALTER TABLE branching.seed_execution_log ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_branching_seed_execution_log_tenant_id ON branching.seed_execution_log(tenant_id);
COMMENT ON COLUMN branching.seed_execution_log.tenant_id IS 'Tenant this seed execution log belongs to';

-- ============================================
-- RPC SCHEMA
-- ============================================

-- RPC executions
ALTER TABLE rpc.executions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_rpc_executions_tenant_id ON rpc.executions(tenant_id);
COMMENT ON COLUMN rpc.executions.tenant_id IS 'Tenant this execution belongs to';

-- RPC procedures
ALTER TABLE rpc.procedures ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_rpc_procedures_tenant_id ON rpc.procedures(tenant_id);
COMMENT ON COLUMN rpc.procedures.tenant_id IS 'Tenant this procedure belongs to';

-- ============================================
-- REALTIME SCHEMA
-- ============================================

-- Schema registry
ALTER TABLE realtime.schema_registry ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_realtime_schema_registry_tenant_id ON realtime.schema_registry(tenant_id);
COMMENT ON COLUMN realtime.schema_registry.tenant_id IS 'Tenant this schema entry belongs to';

-- ============================================
-- COMPLETION NOTICE
-- ============================================

DO $$
BEGIN
    RAISE NOTICE 'Migration 115 complete: Added tenant_id columns to all tenant-scoped tables';
END $$;
