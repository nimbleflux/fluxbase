-- Rollback: Restore legacy dashboard roles

-- Step 1: Restore the role constraint to include all roles
ALTER TABLE platform.users
DROP CONSTRAINT IF EXISTS platform_users_role_check;

ALTER TABLE platform.users
ADD CONSTRAINT dashboard_users_role_check
CHECK (role IN ('instance_admin', 'tenant_admin', 'dashboard_admin', 'dashboard_user'));

-- Step 2: Drop the helper function
DROP FUNCTION IF EXISTS auth.has_admin_role();

-- Step 3: Restore policies to check for dashboard_admin

-- Note: We can't fully restore all policies without knowing their original state
-- This rollback provides basic restoration for key policies

-- ai.knowledge_bases
DROP POLICY IF EXISTS ai_kb_admin ON ai.knowledge_bases;
CREATE POLICY ai_kb_dashboard_admin ON ai.knowledge_bases
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- ai.documents
DROP POLICY IF EXISTS ai_documents_admin ON ai.documents;
CREATE POLICY ai_documents_dashboard_admin ON ai.documents
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- ai.chunks
DROP POLICY IF EXISTS ai_chunks_admin ON ai.chunks;
CREATE POLICY ai_chunks_dashboard_admin ON ai.chunks
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- jobs.functions
DROP POLICY IF EXISTS jobs_functions_admin ON jobs.functions;
CREATE POLICY jobs_functions_dashboard_admin ON jobs.functions
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- functions.edge_functions
DROP POLICY IF EXISTS edge_functions_admin ON functions.edge_functions;
CREATE POLICY edge_functions_dashboard_admin ON functions.edge_functions
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- storage.buckets
DROP POLICY IF EXISTS buckets_admin ON storage.buckets;
CREATE POLICY buckets_dashboard_admin ON storage.buckets
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- storage.objects
DROP POLICY IF EXISTS objects_admin ON storage.objects;
CREATE POLICY objects_dashboard_admin ON storage.objects
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');

-- auth.users
DROP POLICY IF EXISTS users_admin ON auth.users;
CREATE POLICY users_dashboard_admin ON auth.users
    FOR ALL
    USING (auth.current_user_role() = 'dashboard_admin')
    WITH CHECK (auth.current_user_role() = 'dashboard_admin');
