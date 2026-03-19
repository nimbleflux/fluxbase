--
-- MULTI-TENANCY: TENANT ADMIN ASSIGNMENTS (ROLLBACK)
--

-- Restore is_instance_admin to use dashboard.users
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM dashboard.users du
        WHERE du.id = p_user_id
        AND du.role = 'instance_admin'
        AND (du.deleted_at IS NULL OR du.is_active = true)
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin with global privileges. SECURITY DEFINER to bypass RLS.';

-- Drop the helper function
DROP FUNCTION IF EXISTS platform.user_managed_tenant_ids(UUID);

-- Drop the tenant_admin_assignments table
DROP TABLE IF EXISTS platform.tenant_admin_assignments;

-- Note: We do NOT drop the platform schema here as it may contain other tables
-- created by other migrations. Each migration is responsible for its own cleanup.
