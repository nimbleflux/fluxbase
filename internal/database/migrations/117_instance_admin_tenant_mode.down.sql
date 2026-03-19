--
-- MULTI-TENANCY: ROLLBACK INSTANCE ADMIN TENANT MODE FIX
--

-- Restore original is_instance_admin function
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;

    RETURN EXISTS (
        SELECT 1 FROM platform.users pu
        WHERE pu.id = p_user_id
        AND pu.role = 'instance_admin'
        AND pu.deleted_at IS NULL
        AND pu.is_active = true
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

-- Remove the helper function
DROP FUNCTION IF EXISTS is_tenant_admin_mode();
