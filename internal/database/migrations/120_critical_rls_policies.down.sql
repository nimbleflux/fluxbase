--
-- ROLLBACK: CRITICAL RLS POLICIES FOR MULTI-TENANCY
--

-- ============================================
-- REMOVE PLATFORM.SERVICE_KEYS RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS service_keys_service_all ON platform.service_keys;
DROP POLICY IF EXISTS service_keys_instance_admin ON platform.service_keys;
DROP POLICY IF EXISTS service_keys_tenant_admin ON platform.service_keys;
DROP POLICY IF EXISTS service_keys_tenant_member_read ON platform.service_keys;
DROP POLICY IF EXISTS service_keys_self ON platform.service_keys;

ALTER TABLE platform.service_keys DISABLE ROW LEVEL SECURITY;

-- ============================================
-- REMOVE PLATFORM.KEY_USAGE RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS key_usage_service_all ON platform.key_usage;
DROP POLICY IF EXISTS key_usage_instance_admin ON platform.key_usage;
DROP POLICY IF EXISTS key_usage_tenant_admin ON platform.key_usage;

ALTER TABLE platform.key_usage DISABLE ROW LEVEL SECURITY;

-- ============================================
-- REMOVE PLATFORM.TENANT_ADMIN_ASSIGNMENTS RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS tenant_admin_assignments_service_all ON platform.tenant_admin_assignments;
DROP POLICY IF EXISTS tenant_admin_assignments_instance_admin ON platform.tenant_admin_assignments;
DROP POLICY IF EXISTS tenant_admin_assignments_tenant_admin ON platform.tenant_admin_assignments;
DROP POLICY IF EXISTS tenant_admin_assignments_self ON platform.tenant_admin_assignments;

ALTER TABLE platform.tenant_admin_assignments DISABLE ROW LEVEL SECURITY;

-- ============================================
-- RESTORE PREVIOUS VERSION OF HELPER FUNCTIONS
-- Note: This restores the functions to their pre-migration state
-- If they referenced tables without platform. prefix, that will be restored
-- ============================================

-- Restore is_instance_admin to version from migration 117
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
