--
-- MULTI-TENANCY: TENANT ADMIN RLS POLICIES
-- Add tenant_admin policies to tables missing them
--

-- ============================================
-- FUNCTIONS.SECRETS
-- ============================================

-- Tenant admins can manage their tenant's secrets
CREATE POLICY secrets_tenant_admin ON functions.secrets
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- ============================================
-- JOBS.QUEUE
-- ============================================

-- Tenant admins can manage their tenant's jobs
CREATE POLICY jobs_queue_tenant_admin ON jobs.queue
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- ============================================
-- AUTH.WEBHOOKS
-- ============================================

-- Drop old policy (from migration 020)
DROP POLICY IF EXISTS webhooks_admin_only ON auth.webhooks;

-- Instance admins can manage all webhooks
CREATE POLICY webhooks_instance_admin ON auth.webhooks
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's webhooks
CREATE POLICY webhooks_tenant_admin ON auth.webhooks
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can read their tenant's webhooks
CREATE POLICY webhooks_tenant_member ON auth.webhooks
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- ============================================
-- AUTH.CLIENT_KEYS
-- ============================================

-- Drop old policy (from migration 047)
DROP POLICY IF EXISTS auth_client_keys_policy ON auth.client_keys;

-- Instance admins can manage all client keys
CREATE POLICY client_keys_instance_admin ON auth.client_keys
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's client keys
CREATE POLICY client_keys_tenant_admin ON auth.client_keys
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can read their tenant's client keys
CREATE POLICY client_keys_tenant_member ON auth.client_keys
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- ============================================
-- STORAGE.OBJECTS (add tenant_admin policy)
-- ============================================

-- Tenant admins can manage their tenant's objects
CREATE POLICY objects_tenant_admin ON storage.objects
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- ============================================
-- STORAGE.OBJECT_PERMISSIONS
-- ============================================

-- Drop old policy if exists (was created in initial migration)
DROP POLICY IF EXISTS storage_object_permissions_policy ON storage.object_permissions;

-- Tenant admins can manage their tenant's object permissions
CREATE POLICY object_permissions_tenant_admin ON storage.object_permissions
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- ============================================
-- PLATFORM.TENANTS (add RLS policies)
-- ============================================

ALTER TABLE platform.tenants ENABLE ROW LEVEL SECURITY;

-- Service role bypasses all
CREATE POLICY tenants_service_all ON platform.tenants
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all tenants
CREATE POLICY tenants_instance_admin ON platform.tenants
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- ============================================
-- PLATFORM.TENANT_MEMBERSHIPS (add RLS policies)
-- ============================================

ALTER TABLE platform.tenant_memberships ENABLE ROW LEVEL SECURITY;

-- Service role bypasses all
CREATE POLICY tenant_memberships_service_all ON platform.tenant_memberships
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all memberships
CREATE POLICY tenant_memberships_instance_admin ON platform.tenant_memberships
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can view their own memberships
CREATE POLICY tenant_memberships_self ON platform.tenant_memberships
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- Tenant admins can manage memberships in their tenant
CREATE POLICY tenant_memberships_tenant_admin ON platform.tenant_memberships
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

DO $$
BEGIN
    RAISE NOTICE 'Tenant admin RLS policies added';
END $$;
