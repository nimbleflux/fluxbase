--
-- MULTI-TENANCY: ROLLBACK DATA BACKFILL AND RLS POLICIES
--

-- ============================================
-- RESTORE ORIGINAL RLS POLICIES
-- ============================================

-- Helper function to safely drop policies
CREATE OR REPLACE FUNCTION drop_policy_if_exists(p_table TEXT, p_policy TEXT, p_schema TEXT DEFAULT 'public') RETURNS VOID AS $$
BEGIN
    EXECUTE format('DROP POLICY IF EXISTS %I ON %I.%I', p_policy, p_schema, p_table);
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- AUTH.USERS - Restore original policies
-- ============================================

SELECT drop_policy_if_exists('users', 'users_self', 'auth');
SELECT drop_policy_if_exists('users', 'users_tenant_member_read', 'auth');
SELECT drop_policy_if_exists('users', 'users_tenant_admin', 'auth');
SELECT drop_policy_if_exists('users', 'users_instance_admin', 'auth');
SELECT drop_policy_if_exists('users', 'users_service_all', 'auth');

-- Restore original policies
CREATE POLICY users_service_all ON auth.users
    FOR ALL TO service_role USING (true) WITH CHECK (true);

CREATE POLICY users_owner ON auth.users
    FOR ALL TO authenticated
    USING (auth.uid() = id)
    WITH CHECK (auth.uid() = id);

-- ============================================
-- STORAGE.BUCKETS - Restore original policies
-- ============================================

SELECT drop_policy_if_exists('buckets', 'buckets_public_read', 'storage');
SELECT drop_policy_if_exists('buckets', 'buckets_tenant_member_read', 'storage');
SELECT drop_policy_if_exists('buckets', 'buckets_tenant_admin', 'storage');
SELECT drop_policy_if_exists('buckets', 'buckets_instance_admin', 'storage');
SELECT drop_policy_if_exists('buckets', 'buckets_service_all', 'storage');

CREATE POLICY storage_buckets_admin ON storage.buckets
    FOR ALL TO service_role USING (true) WITH CHECK (true);

CREATE POLICY storage_buckets_public_view ON storage.buckets
    FOR SELECT USING (true);

-- ============================================
-- STORAGE.OBJECTS - Restore original policies
-- ============================================

SELECT drop_policy_if_exists('objects', 'objects_public_read', 'storage');
SELECT drop_policy_if_exists('objects', 'objects_tenant_member', 'storage');
SELECT drop_policy_if_exists('objects', 'objects_instance_admin', 'storage');
SELECT drop_policy_if_exists('objects', 'objects_service_all', 'storage');

CREATE POLICY storage_objects_admin ON storage.objects
    FOR ALL TO service_role USING (true) WITH CHECK (true);

CREATE POLICY storage_objects_owner ON storage.objects
    FOR ALL TO authenticated
    USING (auth.uid() = owner_id)
    WITH CHECK (auth.uid() = owner_id);

CREATE POLICY storage_objects_public_read ON storage.objects
    FOR SELECT TO anon, authenticated
    USING (storage.is_bucket_public(bucket_id));

-- ============================================
-- FUNCTIONS.EDGE_FUNCTIONS - Restore original policies
-- ============================================

SELECT drop_policy_if_exists('edge_functions', 'edge_functions_tenant_member', 'functions');
SELECT drop_policy_if_exists('edge_functions', 'edge_functions_tenant_admin', 'functions');
SELECT drop_policy_if_exists('edge_functions', 'edge_functions_instance_admin', 'functions');
SELECT drop_policy_if_exists('edge_functions', 'edge_functions_service_all', 'functions');

CREATE POLICY edge_functions_service_all ON functions.edge_functions
    FOR ALL TO service_role USING (true) WITH CHECK (true);

CREATE POLICY edge_functions_tenant_read ON functions.edge_functions
    FOR SELECT TO authenticated
    USING (auth.uid() IS NOT NULL);

-- ============================================
-- JOBS, AI - Drop tenant policies
-- ============================================

SELECT drop_policy_if_exists('functions', 'jobs_functions_tenant_member', 'jobs');
SELECT drop_policy_if_exists('functions', 'jobs_functions_tenant_admin', 'jobs');
SELECT drop_policy_if_exists('functions', 'jobs_functions_instance_admin', 'jobs');
SELECT drop_policy_if_exists('functions', 'jobs_functions_service_all', 'jobs');

SELECT drop_policy_if_exists('knowledge_bases', 'knowledge_bases_tenant_member', 'ai');
SELECT drop_policy_if_exists('knowledge_bases', 'knowledge_bases_tenant_admin', 'ai');
SELECT drop_policy_if_exists('knowledge_bases', 'knowledge_bases_instance_admin', 'ai');
SELECT drop_policy_if_exists('knowledge_bases', 'knowledge_bases_service_all', 'ai');

SELECT drop_policy_if_exists('chatbots', 'chatbots_tenant_member', 'ai');
SELECT drop_policy_if_exists('chatbots', 'chatbots_tenant_admin', 'ai');
SELECT drop_policy_if_exists('chatbots', 'chatbots_instance_admin', 'ai');
SELECT drop_policy_if_exists('chatbots', 'chatbots_service_all', 'ai');

-- ============================================
-- CLEAR TENANT_ID VALUES (set to NULL)
-- ============================================

-- AI Schema
UPDATE ai.table_export_sync_configs SET tenant_id = NULL;
UPDATE ai.knowledge_base_permissions SET tenant_id = NULL;
UPDATE ai.document_permissions SET tenant_id = NULL;
UPDATE ai.chatbot_knowledge_bases SET tenant_id = NULL;
UPDATE ai.user_quotas SET tenant_id = NULL;
UPDATE ai.user_provider_preferences SET tenant_id = NULL;
UPDATE ai.user_chatbot_usage SET tenant_id = NULL;
UPDATE ai.conversations SET tenant_id = NULL;
UPDATE ai.chatbots SET tenant_id = NULL;
UPDATE ai.documents SET tenant_id = NULL;
UPDATE ai.knowledge_bases SET tenant_id = NULL;

-- Jobs Schema
UPDATE jobs.queue SET tenant_id = NULL;
UPDATE jobs.functions SET tenant_id = NULL;

-- Functions Schema
UPDATE functions.edge_triggers SET tenant_id = NULL;
UPDATE functions.secrets SET tenant_id = NULL;
UPDATE functions.edge_functions SET tenant_id = NULL;

-- Storage Schema
UPDATE storage.object_permissions SET tenant_id = NULL;
UPDATE storage.objects SET tenant_id = NULL;
UPDATE storage.buckets SET tenant_id = NULL;

-- Auth Schema
UPDATE auth.client_keys SET tenant_id = NULL;
UPDATE auth.webhooks SET tenant_id = NULL;
UPDATE auth.users SET tenant_id = NULL;

-- ============================================
-- CLEANUP
-- ============================================

DROP FUNCTION IF EXISTS drop_policy_if_exists(TEXT, TEXT, TEXT);

DO $$
BEGIN
    RAISE NOTICE 'RLS policies restored to pre-multi-tenancy state';
END $$;
