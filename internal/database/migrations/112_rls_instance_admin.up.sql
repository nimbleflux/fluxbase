-- Migration: Remove legacy dashboard roles and update RLS policies
-- This migration:
-- 1. Removes dashboard_admin and dashboard_user roles from the constraint
-- 2. Updates RLS policies to accept instance_admin role

-- Step 1: Update the role constraint to only allow instance_admin and tenant_admin
ALTER TABLE platform.users
DROP CONSTRAINT IF EXISTS dashboard_users_role_check;

ALTER TABLE platform.users
ADD CONSTRAINT platform_users_role_check
CHECK (role IN ('instance_admin', 'tenant_admin'));

-- Step 2: Create a helper function to check if user has admin role
CREATE OR REPLACE FUNCTION auth.has_admin_role() RETURNS BOOLEAN AS $$
DECLARE
    role_var TEXT;
BEGIN
    role_var := auth.current_user_role();
    RETURN role_var IN ('instance_admin', 'service_role');
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION auth.has_admin_role() IS
'Checks if the current user has an admin role (instance_admin or service_role).';

-- Step 3: Update RLS policies that checked for dashboard_admin

-- ai.knowledge_bases
DROP POLICY IF EXISTS ai_kb_dashboard_admin ON ai.knowledge_bases;
CREATE POLICY ai_kb_admin ON ai.knowledge_bases
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- ai.documents
DROP POLICY IF EXISTS ai_documents_dashboard_admin ON ai.documents;
CREATE POLICY ai_documents_admin ON ai.documents
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- ai.chunks
DROP POLICY IF EXISTS ai_chunks_dashboard_admin ON ai.chunks;
CREATE POLICY ai_chunks_admin ON ai.chunks
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- jobs.functions
DROP POLICY IF EXISTS jobs_functions_dashboard_admin ON jobs.functions;
DROP POLICY IF EXISTS jobs_functions_admin ON jobs.functions;
CREATE POLICY jobs_functions_admin ON jobs.functions
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- functions.edge_functions
DROP POLICY IF EXISTS edge_functions_dashboard_admin ON functions.edge_functions;
DROP POLICY IF EXISTS edge_functions_admin ON functions.edge_functions;
CREATE POLICY edge_functions_admin ON functions.edge_functions
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- storage.buckets
DROP POLICY IF EXISTS buckets_dashboard_admin ON storage.buckets;
DROP POLICY IF EXISTS buckets_admin ON storage.buckets;
CREATE POLICY buckets_admin ON storage.buckets
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- storage.objects (admin override policy)
DROP POLICY IF EXISTS objects_dashboard_admin ON storage.objects;
DROP POLICY IF EXISTS objects_admin ON storage.objects;
CREATE POLICY objects_admin ON storage.objects
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- auth.users (admin override)
DROP POLICY IF EXISTS users_dashboard_admin ON auth.users;
DROP POLICY IF EXISTS users_admin ON auth.users;
CREATE POLICY users_admin ON auth.users
    FOR ALL
    USING (auth.has_admin_role())
    WITH CHECK (auth.has_admin_role());

-- rpc.functions (only if the table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'rpc' AND table_name = 'functions') THEN
        DROP POLICY IF EXISTS rpc_functions_dashboard_admin ON rpc.functions;
        DROP POLICY IF EXISTS rpc_functions_admin ON rpc.functions;
        CREATE POLICY rpc_functions_admin ON rpc.functions
            FOR ALL
            USING (auth.has_admin_role())
            WITH CHECK (auth.has_admin_role());
    END IF;
END $$;

-- rpc.procedures (only if the table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'rpc' AND table_name = 'procedures') THEN
        DROP POLICY IF EXISTS rpc_procedures_dashboard_admin ON rpc.procedures;
        DROP POLICY IF EXISTS rpc_procedures_admin ON rpc.procedures;
        CREATE POLICY rpc_procedures_admin ON rpc.procedures
            FOR ALL
            USING (auth.has_admin_role())
            WITH CHECK (auth.has_admin_role());
    END IF;
END $$;

-- platform.sso_identities
DROP POLICY IF EXISTS sso_identities_dashboard_admin ON platform.sso_identities;
DROP POLICY IF EXISTS sso_identities_admin ON platform.sso_identities;
CREATE POLICY sso_identities_admin ON platform.sso_identities
    FOR ALL
    USING (auth.has_admin_role() OR auth.current_user_role() = 'service_role')
    WITH CHECK (auth.has_admin_role() OR auth.current_user_role() = 'service_role');

-- platform.sessions (admin can view all)
DROP POLICY IF EXISTS platform_sessions_admin ON platform.sessions;
CREATE POLICY platform_sessions_admin ON platform.sessions
    FOR ALL
    USING (auth.has_admin_role() OR auth.current_user_role() = 'service_role');

-- Update comments
COMMENT ON POLICY ai_kb_admin ON ai.knowledge_bases IS
'Instance admins and service role have full access to knowledge bases.';
COMMENT ON POLICY ai_documents_admin ON ai.documents IS
'Instance admins and service role have full access to documents.';
COMMENT ON POLICY ai_chunks_admin ON ai.chunks IS
'Instance admins and service role have full access to chunks.';
COMMENT ON POLICY jobs_functions_admin ON jobs.functions IS
'Instance admins and service role have full access to background jobs.';
COMMENT ON POLICY edge_functions_admin ON functions.edge_functions IS
'Instance admins and service role have full access to edge functions.';
COMMENT ON POLICY buckets_admin ON storage.buckets IS
'Instance admins and service role have full access to storage buckets.';
COMMENT ON POLICY objects_admin ON storage.objects IS
'Instance admins and service role have full access to storage objects.';
COMMENT ON POLICY users_admin ON auth.users IS
'Instance admins and service role have full access to users.';
