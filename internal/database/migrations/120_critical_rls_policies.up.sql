--
-- CRITICAL RLS POLICIES FOR MULTI-TENANCY
-- Adds missing RLS policies for platform.service_keys, platform.key_usage, and platform.tenant_admin_assignments
-- Also fixes RLS helper functions to use platform. schema prefix
--

-- ============================================
-- PLATFORM.SERVICE_KEYS RLS POLICIES
-- ============================================

-- Enable RLS
ALTER TABLE platform.service_keys ENABLE ROW LEVEL SECURITY;

-- Service role bypasses all
CREATE POLICY service_keys_service_all ON platform.service_keys
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all keys (only when not acting as tenant admin)
CREATE POLICY service_keys_instance_admin ON platform.service_keys
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's keys
CREATE POLICY service_keys_tenant_admin ON platform.service_keys
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant members can view their tenant's keys (read-only)
CREATE POLICY service_keys_tenant_member_read ON platform.service_keys
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_is_tenant_member(auth.uid(), current_tenant_id())
    );

-- Users can view their own publishable keys
CREATE POLICY service_keys_self ON platform.service_keys
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.KEY_USAGE RLS POLICIES
-- ============================================

-- Enable RLS
ALTER TABLE platform.key_usage ENABLE ROW LEVEL SECURITY;

-- Service role bypasses all
CREATE POLICY key_usage_service_all ON platform.key_usage
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can view all usage
CREATE POLICY key_usage_instance_admin ON platform.key_usage
    FOR SELECT TO authenticated
    USING (is_instance_admin(auth.uid()));

-- Tenant admins can view their tenant's key usage
CREATE POLICY key_usage_tenant_admin ON platform.key_usage
    FOR SELECT TO authenticated
    USING (
        EXISTS (
            SELECT 1 FROM platform.service_keys sk
            WHERE sk.id = key_usage.key_id
            AND sk.tenant_id = current_tenant_id()
        )
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- ============================================
-- PLATFORM.TENANT_ADMIN_ASSIGNMENTS RLS POLICIES
-- ============================================

-- Enable RLS
ALTER TABLE platform.tenant_admin_assignments ENABLE ROW LEVEL SECURITY;

-- Service role bypasses all
CREATE POLICY tenant_admin_assignments_service_all ON platform.tenant_admin_assignments
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all assignments (only when not acting as tenant admin)
CREATE POLICY tenant_admin_assignments_instance_admin ON platform.tenant_admin_assignments
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's admin assignments
CREATE POLICY tenant_admin_assignments_tenant_admin ON platform.tenant_admin_assignments
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Users can view their own assignments
CREATE POLICY tenant_admin_assignments_self ON platform.tenant_admin_assignments
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- FIX RLS HELPER FUNCTIONS TO USE platform. SCHEMA PREFIX
-- ============================================

-- Fix current_tenant_id() to use platform. prefix
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
DECLARE
    claims JSONB;
    tid UUID;
    default_id UUID;
BEGIN
    -- Try JWT claims first
    BEGIN
        claims := auth.jwt();
        IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
            tid := (claims->>'tenant_id')::UUID;
            IF EXISTS (SELECT 1 FROM platform.tenants WHERE id = tid AND deleted_at IS NULL) THEN
                RETURN tid;
            END IF;
        END IF;
    EXCEPTION WHEN OTHERS THEN NULL;
    END;
    
    -- Fall back to session variable (set by middleware)
    BEGIN
        tid := current_setting('app.current_tenant_id', true)::UUID;
        IF tid IS NOT NULL THEN
            IF EXISTS (SELECT 1 FROM platform.tenants WHERE id = tid AND deleted_at IS NULL) THEN
                RETURN tid;
            END IF;
        END IF;
    EXCEPTION WHEN OTHERS THEN NULL;
    END;
    
    -- Return default tenant for backward compatibility
    SELECT id INTO default_id FROM platform.tenants WHERE is_default = true AND deleted_at IS NULL LIMIT 1;
    RETURN default_id;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION current_tenant_id() IS
'Returns the current tenant ID from JWT claims, session variable, or falls back to default tenant. SECURITY DEFINER to bypass RLS.';

-- Fix user_has_tenant_role() to use platform. prefix
CREATE OR REPLACE FUNCTION user_has_tenant_role(
    p_user_id UUID,
    p_tenant_id UUID,
    p_role TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL OR p_tenant_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM platform.tenant_memberships tm
        INNER JOIN platform.tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND tm.tenant_id = p_tenant_id
        AND tm.role = p_role
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_has_tenant_role(UUID, UUID, TEXT) IS
'Checks if a user has a specific role in a tenant. SECURITY DEFINER to bypass RLS.';

-- Fix user_is_tenant_member() to use platform. prefix
CREATE OR REPLACE FUNCTION user_is_tenant_member(
    p_user_id UUID,
    p_tenant_id UUID
) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL OR p_tenant_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM platform.tenant_memberships tm
        INNER JOIN platform.tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND tm.tenant_id = p_tenant_id
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_is_tenant_member(UUID, UUID) IS
'Checks if a user is a member of a tenant (any role). SECURITY DEFINER to bypass RLS.';

-- Update is_instance_admin() to ensure it uses platform. prefix
-- This function was updated in migration 117 but let's ensure consistency
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
DECLARE
    tenant_context UUID;
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;

    -- Check if a tenant context is set (acting as tenant admin mode)
    BEGIN
        tenant_context := current_setting('app.current_tenant_id', true)::UUID;
        IF tenant_context IS NOT NULL THEN
            RETURN false;  -- In tenant admin mode, not instance admin
        END IF;
    EXCEPTION WHEN OTHERS THEN NULL;
    END;

    -- Also check JWT claims for tenant context
    BEGIN
        DECLARE claims JSONB;
        BEGIN
            claims := auth.jwt();
            IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
                RETURN false;  -- In tenant admin mode via JWT
            END IF;
        END;
    EXCEPTION WHEN OTHERS THEN NULL;
    END;

    -- No tenant context, check if user is actually an instance admin
    RETURN EXISTS (
        SELECT 1 FROM platform.users pu
        WHERE pu.id = p_user_id
        AND pu.role = 'instance_admin'
        AND pu.deleted_at IS NULL
        AND pu.is_active = true
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin with global privileges. Returns false when acting as tenant admin (tenant context set). SECURITY DEFINER to bypass RLS.';
