--
-- ROLLBACK: TENANT ADMIN RLS POLICIES
--

-- FUNCTIONS.SECRETS
DROP POLICY IF EXISTS secrets_tenant_admin ON functions.secrets;

-- JOBS.QUEUE
DROP POLICY IF EXISTS jobs_queue_tenant_admin ON jobs.queue;

-- AUTH.WEBHOOKS
DROP POLICY IF EXISTS webhooks_instance_admin ON auth.webhooks;
DROP POLICY IF EXISTS webhooks_tenant_admin ON auth.webhooks;
DROP POLICY IF EXISTS webhooks_tenant_member ON auth.webhooks;

-- AUTH.CLIENT_KEYS
DROP POLICY IF EXISTS client_keys_instance_admin ON auth.client_keys;
DROP POLICY IF EXISTS client_keys_tenant_admin ON auth.client_keys;
DROP POLICY IF EXISTS client_keys_tenant_member ON auth.client_keys;

-- STORAGE.OBJECTS
DROP POLICY IF EXISTS objects_tenant_admin ON storage.objects;

-- STORAGE.OBJECT_PERMISSIONS
DROP POLICY IF EXISTS object_permissions_tenant_admin ON storage.object_permissions;

-- PLATFORM.TENANTS
DROP POLICY IF EXISTS tenants_service_all ON platform.tenants;
DROP POLICY IF EXISTS tenants_instance_admin ON platform.tenants;

-- PLATFORM.TENANT_MEMBERSHIPS
DROP POLICY IF EXISTS tenant_memberships_service_all ON platform.tenant_memberships;
DROP POLICY IF EXISTS tenant_memberships_instance_admin ON platform.tenant_memberships;
DROP POLICY IF EXISTS tenant_memberships_self ON platform.tenant_memberships;
DROP POLICY IF EXISTS tenant_memberships_tenant_admin ON platform.tenant_memberships;

-- Recreate old policies (if needed for rollback)

-- auth.webhooks old policy
CREATE POLICY webhooks_admin_only ON auth.webhooks
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- auth.client_keys old policy  
CREATE POLICY auth_client_keys_policy ON auth.client_keys
    FOR ALL TO service_role USING (true) WITH CHECK (true);

DO $$
BEGIN
    RAISE NOTICE 'Tenant admin RLS policies rolled back';
END $$;
