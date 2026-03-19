--
-- MULTI-TENANCY: BACKFILL DATA AND ADD RLS POLICIES
-- Backfills all existing data to default tenant and creates tenant-aware RLS policies
--

-- ============================================
-- STEP 1: BACKFILL TENANT_ID ON ALL TABLES
-- ============================================

DO $$
DECLARE
    default_tenant_id UUID;
    count INTEGER;
BEGIN
    -- Get default tenant ID
    SELECT id INTO default_tenant_id FROM platform.tenants WHERE is_default = true LIMIT 1;

    IF default_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Default tenant not found. Run earlier migrations first.';
    END IF;

    RAISE NOTICE 'Backfilling data to default tenant: %', default_tenant_id;

    -- AI Schema
    UPDATE ai.chunks SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    GET DIAGNOSTICS count = ROW_COUNT;
    RAISE NOTICE 'Backfilled ai.chunks: % rows', count;

    UPDATE ai.messages SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.entities SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.entity_relationships SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.document_entities SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.query_audit_log SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.retrieval_log SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Functions Schema
    UPDATE functions.edge_executions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.edge_files SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.function_dependencies SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.secret_versions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.shared_modules SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Jobs Schema
    UPDATE jobs.function_files SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE jobs.workers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Auth Schema (sessions get tenant from user)
    UPDATE auth.sessions s SET tenant_id = u.tenant_id
        FROM auth.users u WHERE s.user_id = u.id AND s.tenant_id IS NULL AND u.tenant_id IS NOT NULL;
    UPDATE auth.sessions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    UPDATE auth.mfa_factors m SET tenant_id = u.tenant_id
        FROM auth.users u WHERE m.user_id = u.id AND m.tenant_id IS NULL AND u.tenant_id IS NOT NULL;
    UPDATE auth.mfa_factors SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    UPDATE auth.oauth_links o SET tenant_id = u.tenant_id
        FROM auth.users u WHERE o.user_id = u.id AND o.tenant_id IS NULL AND u.tenant_id IS NOT NULL;
    UPDATE auth.oauth_links SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    UPDATE auth.impersonation_sessions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhook_deliveries SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhook_events SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhook_monitored_tables SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.client_key_usage SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Storage Schema
    UPDATE storage.chunked_upload_sessions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Logging Schema (update parent, partitions inherit)
    UPDATE logging.entries SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Branching Schema
    UPDATE branching.branches SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE branching.branch_access SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE branching.github_config SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE branching.activity_log SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE branching.migration_history SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE branching.seed_execution_log SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- RPC Schema
    UPDATE rpc.executions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE rpc.procedures SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Realtime Schema
    UPDATE realtime.schema_registry SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    RAISE NOTICE 'Tenant ID backfill complete';
END $$;

-- ============================================
-- STEP 2: ADD RLS POLICIES
-- ============================================

-- Helper function to safely drop policies
CREATE OR REPLACE FUNCTION drop_policy_if_exists(p_table TEXT, p_policy TEXT, p_schema TEXT DEFAULT 'public') RETURNS VOID AS $$
BEGIN
    EXECUTE format('DROP POLICY IF EXISTS %I ON %I.%I', p_policy, p_schema, p_table);
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- AI SCHEMA RLS POLICIES
-- ============================================

-- ai.chunks
CREATE POLICY chunks_tenant_service ON ai.chunks
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY chunks_instance_admin ON ai.chunks
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY chunks_tenant_admin ON ai.chunks
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY chunks_tenant_member ON ai.chunks
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ai.messages
CREATE POLICY messages_tenant_service ON ai.messages
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY messages_instance_admin ON ai.messages
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY messages_tenant_admin ON ai.messages
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY messages_tenant_member ON ai.messages
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ai.entities
CREATE POLICY entities_tenant_service ON ai.entities
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY entities_instance_admin ON ai.entities
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY entities_tenant_admin ON ai.entities
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY entities_tenant_member ON ai.entities
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ai.entity_relationships
CREATE POLICY entity_relationships_tenant_service ON ai.entity_relationships
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY entity_relationships_instance_admin ON ai.entity_relationships
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY entity_relationships_tenant_admin ON ai.entity_relationships
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY entity_relationships_tenant_member ON ai.entity_relationships
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ai.document_entities
CREATE POLICY document_entities_tenant_service ON ai.document_entities
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY document_entities_instance_admin ON ai.document_entities
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY document_entities_tenant_admin ON ai.document_entities
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY document_entities_tenant_member ON ai.document_entities
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ai.query_audit_log
CREATE POLICY query_audit_log_tenant_service ON ai.query_audit_log
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY query_audit_log_instance_admin ON ai.query_audit_log
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY query_audit_log_tenant_admin ON ai.query_audit_log
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- ai.retrieval_log
CREATE POLICY retrieval_log_tenant_service ON ai.retrieval_log
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY retrieval_log_instance_admin ON ai.retrieval_log
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY retrieval_log_tenant_admin ON ai.retrieval_log
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- ============================================
-- FUNCTIONS SCHEMA RLS POLICIES
-- ============================================

-- functions.edge_executions
CREATE POLICY edge_executions_tenant_service ON functions.edge_executions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY edge_executions_instance_admin ON functions.edge_executions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY edge_executions_tenant_admin ON functions.edge_executions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY edge_executions_tenant_member ON functions.edge_executions
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- functions.edge_files
CREATE POLICY edge_files_tenant_service ON functions.edge_files
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY edge_files_instance_admin ON functions.edge_files
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY edge_files_tenant_admin ON functions.edge_files
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY edge_files_tenant_member ON functions.edge_files
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- functions.function_dependencies
CREATE POLICY function_dependencies_tenant_service ON functions.function_dependencies
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY function_dependencies_instance_admin ON functions.function_dependencies
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY function_dependencies_tenant_admin ON functions.function_dependencies
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY function_dependencies_tenant_member ON functions.function_dependencies
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- functions.secret_versions
CREATE POLICY secret_versions_tenant_service ON functions.secret_versions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY secret_versions_instance_admin ON functions.secret_versions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY secret_versions_tenant_admin ON functions.secret_versions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- functions.shared_modules
CREATE POLICY shared_modules_tenant_service ON functions.shared_modules
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY shared_modules_instance_admin ON functions.shared_modules
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY shared_modules_tenant_admin ON functions.shared_modules
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY shared_modules_tenant_member ON functions.shared_modules
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- JOBS SCHEMA RLS POLICIES
-- ============================================

-- jobs.function_files
CREATE POLICY jobs_function_files_tenant_service ON jobs.function_files
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY jobs_function_files_instance_admin ON jobs.function_files
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY jobs_function_files_tenant_admin ON jobs.function_files
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY jobs_function_files_tenant_member ON jobs.function_files
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- jobs.workers
CREATE POLICY workers_tenant_service ON jobs.workers
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY workers_instance_admin ON jobs.workers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY workers_tenant_admin ON jobs.workers
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY workers_tenant_member ON jobs.workers
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- AUTH SCHEMA RLS POLICIES
-- ============================================

-- auth.sessions (add tenant policies to existing)
CREATE POLICY sessions_tenant_service ON auth.sessions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY sessions_tenant_admin ON auth.sessions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- auth.mfa_factors
CREATE POLICY mfa_factors_tenant_service ON auth.mfa_factors
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY mfa_factors_instance_admin ON auth.mfa_factors
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY mfa_factors_tenant_admin ON auth.mfa_factors
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- auth.oauth_links
CREATE POLICY oauth_links_tenant_service ON auth.oauth_links
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY oauth_links_instance_admin ON auth.oauth_links
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY oauth_links_tenant_admin ON auth.oauth_links
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- auth.impersonation_sessions
CREATE POLICY impersonation_sessions_tenant_service ON auth.impersonation_sessions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY impersonation_sessions_instance_admin ON auth.impersonation_sessions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY impersonation_sessions_tenant_admin ON auth.impersonation_sessions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- auth.webhook_deliveries
CREATE POLICY webhook_deliveries_tenant_service ON auth.webhook_deliveries
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY webhook_deliveries_instance_admin ON auth.webhook_deliveries
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY webhook_deliveries_tenant_admin ON auth.webhook_deliveries
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY webhook_deliveries_tenant_member ON auth.webhook_deliveries
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- auth.webhook_events
CREATE POLICY webhook_events_tenant_service ON auth.webhook_events
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY webhook_events_instance_admin ON auth.webhook_events
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY webhook_events_tenant_admin ON auth.webhook_events
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY webhook_events_tenant_member ON auth.webhook_events
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- auth.webhook_monitored_tables
CREATE POLICY webhook_monitored_tables_tenant_service ON auth.webhook_monitored_tables
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY webhook_monitored_tables_instance_admin ON auth.webhook_monitored_tables
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY webhook_monitored_tables_tenant_admin ON auth.webhook_monitored_tables
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY webhook_monitored_tables_tenant_member ON auth.webhook_monitored_tables
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- auth.client_key_usage
CREATE POLICY client_key_usage_tenant_service ON auth.client_key_usage
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY client_key_usage_instance_admin ON auth.client_key_usage
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY client_key_usage_tenant_admin ON auth.client_key_usage
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY client_key_usage_tenant_member ON auth.client_key_usage
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- STORAGE SCHEMA RLS POLICIES
-- ============================================

-- storage.chunked_upload_sessions
CREATE POLICY chunked_upload_sessions_tenant_service ON storage.chunked_upload_sessions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY chunked_upload_sessions_instance_admin ON storage.chunked_upload_sessions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY chunked_upload_sessions_tenant_admin ON storage.chunked_upload_sessions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY chunked_upload_sessions_tenant_member ON storage.chunked_upload_sessions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()))
    WITH CHECK (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- LOGGING SCHEMA RLS POLICIES
-- ============================================

-- logging.entries (parent partitioned table - policies apply to all partitions)
CREATE POLICY entries_tenant_service ON logging.entries
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY entries_instance_admin ON logging.entries
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY entries_tenant_admin ON logging.entries
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- Note: No separate policies needed for partition tables (entries_ai, entries_custom, etc.)
-- as they inherit RLS policies from the parent logging.entries table

-- ============================================
-- BRANCHING SCHEMA RLS POLICIES
-- ============================================

-- branching.branches
CREATE POLICY branches_tenant_service ON branching.branches
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY branches_instance_admin ON branching.branches
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY branches_tenant_admin ON branching.branches
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY branches_tenant_member ON branching.branches
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- branching.branch_access
CREATE POLICY branch_access_tenant_service ON branching.branch_access
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY branch_access_instance_admin ON branching.branch_access
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY branch_access_tenant_admin ON branching.branch_access
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY branch_access_tenant_member ON branching.branch_access
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- branching.github_config
CREATE POLICY github_config_tenant_service ON branching.github_config
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY github_config_instance_admin ON branching.github_config
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY github_config_tenant_admin ON branching.github_config
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- branching.activity_log
CREATE POLICY branching_activity_log_tenant_service ON branching.activity_log
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY branching_activity_log_instance_admin ON branching.activity_log
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY branching_activity_log_tenant_admin ON branching.activity_log
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY branching_activity_log_tenant_member ON branching.activity_log
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- branching.migration_history
CREATE POLICY migration_history_tenant_service ON branching.migration_history
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY migration_history_instance_admin ON branching.migration_history
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY migration_history_tenant_admin ON branching.migration_history
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY migration_history_tenant_member ON branching.migration_history
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- branching.seed_execution_log
CREATE POLICY seed_execution_log_tenant_service ON branching.seed_execution_log
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY seed_execution_log_instance_admin ON branching.seed_execution_log
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY seed_execution_log_tenant_admin ON branching.seed_execution_log
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

-- ============================================
-- RPC SCHEMA RLS POLICIES
-- ============================================

-- rpc.executions
CREATE POLICY rpc_executions_tenant_service ON rpc.executions
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY rpc_executions_instance_admin ON rpc.executions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY rpc_executions_tenant_admin ON rpc.executions
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY rpc_executions_tenant_member ON rpc.executions
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- rpc.procedures
CREATE POLICY procedures_tenant_service ON rpc.procedures
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY procedures_instance_admin ON rpc.procedures
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY procedures_tenant_admin ON rpc.procedures
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY procedures_tenant_member ON rpc.procedures
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- REALTIME SCHEMA RLS POLICIES
-- ============================================

-- realtime.schema_registry
CREATE POLICY schema_registry_tenant_service ON realtime.schema_registry
    FOR ALL TO tenant_service USING (true) WITH CHECK (true);

CREATE POLICY schema_registry_instance_admin ON realtime.schema_registry
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY schema_registry_tenant_admin ON realtime.schema_registry
    FOR ALL TO authenticated
    USING (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'))
    WITH CHECK (tenant_id = current_tenant_id() AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin'));

CREATE POLICY schema_registry_tenant_member ON realtime.schema_registry
    FOR SELECT TO authenticated
    USING (tenant_id = current_tenant_id() AND user_is_tenant_member(auth.uid(), current_tenant_id()));

-- ============================================
-- CLEANUP
-- ============================================

DROP FUNCTION IF EXISTS drop_policy_if_exists(TEXT, TEXT, TEXT);

DO $$
BEGIN
    RAISE NOTICE 'Migration 116 complete: Backfilled data and added RLS policies';
END $$;
