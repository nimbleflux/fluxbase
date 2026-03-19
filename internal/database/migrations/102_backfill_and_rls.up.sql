--
-- MULTI-TENANCY: DATA BACKFILL AND RLS POLICIES
-- Backfills all existing data to default tenant and creates tenant-aware RLS policies
--

DO $$
DECLARE
    default_tenant_id UUID;
    user_count INTEGER;
BEGIN
    -- Get default tenant ID from platform schema
    SELECT id INTO default_tenant_id FROM platform.tenants WHERE is_default = true LIMIT 1;

    IF default_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Default tenant not found. Run 101_migrate_admin_roles migration first.';
    END IF;

    RAISE NOTICE 'Backfilling data to default tenant: %', default_tenant_id;

    -- ============================================
    -- BACKFILL TENANT_ID ON ALL TABLES
    -- ============================================

    -- Auth schema
    UPDATE auth.users SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhooks SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.client_keys SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Storage schema
    UPDATE storage.buckets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE storage.objects SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE storage.object_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Functions schema
    UPDATE functions.edge_functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.secrets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.edge_triggers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Jobs schema
    UPDATE jobs.functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE jobs.queue SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- AI schema
    UPDATE ai.knowledge_bases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.documents SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.chatbots SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.conversations SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_chatbot_usage SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_provider_preferences SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_quotas SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.chatbot_knowledge_bases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.document_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.knowledge_base_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.table_export_sync_configs SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    RAISE NOTICE 'Tenant ID backfill complete';

    -- ============================================
    -- CREATE TENANT MEMBERSHIPS FOR ALL AUTH USERS
    -- ============================================

    -- Create memberships for all existing auth.users
    INSERT INTO platform.tenant_memberships (tenant_id, user_id, role, created_at)
    SELECT
        default_tenant_id,
        id,
        'tenant_member',
        NOW()
    FROM auth.users
    ON CONFLICT (tenant_id, user_id) DO NOTHING;

    GET DIAGNOSTICS user_count = ROW_COUNT;
    RAISE NOTICE 'Created % tenant memberships for auth.users', user_count;

    RAISE NOTICE 'Data backfill complete';
END $$;

-- ============================================
-- UPDATE RLS POLICIES FOR TENANT ISOLATION
-- ============================================

-- Helper function to safely drop policies
CREATE OR REPLACE FUNCTION drop_policy_if_exists(p_table TEXT, p_policy TEXT, p_schema TEXT DEFAULT 'public') RETURNS VOID AS $$
BEGIN
    EXECUTE format('DROP POLICY IF EXISTS %I ON %I.%I', p_policy, p_schema, p_table);
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- AUTH.USERS RLS POLICIES
-- ============================================

-- Drop existing policies (we'll recreate them with tenant awareness)
SELECT drop_policy_if_exists('users', 'users_service_all', 'auth');
SELECT drop_policy_if_exists('users', 'users_owner', 'auth');
SELECT drop_policy_if_exists('users', 'users_admin', 'auth');
SELECT drop_policy_if_exists('users', 'users_self', 'auth');

-- Service role bypasses all
CREATE POLICY users_service_all ON auth.users
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can see and manage all users
CREATE POLICY users_instance_admin ON auth.users
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage users in their tenant
CREATE POLICY users_tenant_admin ON auth.users
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can see other users in their tenant
CREATE POLICY users_tenant_member_read ON auth.users
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- Users can always see and update their own record
CREATE POLICY users_self ON auth.users
    FOR ALL TO authenticated
    USING (auth.uid() = id)
    WITH CHECK (auth.uid() = id);

-- ============================================
-- STORAGE.BUCKETS RLS POLICIES
-- ============================================

SELECT drop_policy_if_exists('buckets', 'storage_buckets_admin', 'storage');
SELECT drop_policy_if_exists('buckets', 'storage_buckets_public_view', 'storage');

-- Service role bypasses all
CREATE POLICY buckets_service_all ON storage.buckets
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all buckets
CREATE POLICY buckets_instance_admin ON storage.buckets
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's buckets
CREATE POLICY buckets_tenant_admin ON storage.buckets
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can see their tenant's buckets
CREATE POLICY buckets_tenant_member_read ON storage.buckets
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- Public buckets can be viewed by anyone (if public = true)
CREATE POLICY buckets_public_read ON storage.buckets
    FOR SELECT TO anon, authenticated
    USING (public = true);

-- ============================================
-- STORAGE.OBJECTS RLS POLICIES
-- ============================================

SELECT drop_policy_if_exists('objects', 'storage_objects_admin', 'storage');
SELECT drop_policy_if_exists('objects', 'storage_objects_owner', 'storage');
SELECT drop_policy_if_exists('objects', 'storage_objects_public_read', 'storage');

-- Service role bypasses all
CREATE POLICY objects_service_all ON storage.objects
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all objects
CREATE POLICY objects_instance_admin ON storage.objects
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant members can access their tenant's objects
CREATE POLICY objects_tenant_member ON storage.objects
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- Public bucket objects can be read by anyone
CREATE POLICY objects_public_read ON storage.objects
    FOR SELECT TO anon, authenticated
    USING (storage.is_bucket_public(bucket_id));

-- ============================================
-- FUNCTIONS.EDGE_FUNCTIONS RLS POLICIES
-- ============================================

SELECT drop_policy_if_exists('edge_functions', 'edge_functions_service_all', 'functions');
SELECT drop_policy_if_exists('edge_functions', 'edge_functions_tenant_read', 'functions');

-- Service role bypasses all
CREATE POLICY edge_functions_service_all ON functions.edge_functions
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all functions
CREATE POLICY edge_functions_instance_admin ON functions.edge_functions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's functions
CREATE POLICY edge_functions_tenant_admin ON functions.edge_functions
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can read and execute their tenant's functions
CREATE POLICY edge_functions_tenant_member ON functions.edge_functions
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- ============================================
-- JOBS.FUNCTIONS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY jobs_functions_service_all ON jobs.functions
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all jobs
CREATE POLICY jobs_functions_instance_admin ON jobs.functions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's jobs
CREATE POLICY jobs_functions_tenant_admin ON jobs.functions
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can read their tenant's jobs
CREATE POLICY jobs_functions_tenant_member ON jobs.functions
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- ============================================
-- AI.KNOWLEDGE_BASES RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY knowledge_bases_service_all ON ai.knowledge_bases
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all knowledge bases
CREATE POLICY knowledge_bases_instance_admin ON ai.knowledge_bases
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's knowledge bases
CREATE POLICY knowledge_bases_tenant_admin ON ai.knowledge_bases
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can access their tenant's knowledge bases based on visibility
CREATE POLICY knowledge_bases_tenant_member ON ai.knowledge_bases
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
        AND (visibility = 'shared' OR visibility = 'public' OR owner_id = auth.uid())
    );

-- ============================================
-- AI.CHATBOTS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY chatbots_service_all ON ai.chatbots
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all chatbots
CREATE POLICY chatbots_instance_admin ON ai.chatbots
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's chatbots
CREATE POLICY chatbots_tenant_admin ON ai.chatbots
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can access their tenant's chatbots
CREATE POLICY chatbots_tenant_member ON ai.chatbots
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- ============================================
-- CLEANUP
-- ============================================

DROP FUNCTION IF EXISTS drop_policy_if_exists(TEXT, TEXT, TEXT);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'RLS policies updated for multi-tenancy';
END $$;
