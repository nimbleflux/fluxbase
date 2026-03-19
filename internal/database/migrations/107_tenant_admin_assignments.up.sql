--
-- MULTI-TENANCY: TENANT ADMIN ASSIGNMENTS
-- Creates table for assigning dashboard users as tenant admins
--

-- Create platform schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS platform;

-- Create tenant_admin_assignments table
CREATE TABLE IF NOT EXISTS platform.tenant_admin_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES platform.users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES platform.tenants(id) ON DELETE CASCADE,
    assigned_by UUID REFERENCES platform.users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT tenant_admin_assignments_unique UNIQUE(user_id, tenant_id)
);

-- Indexes for tenant_admin_assignments
CREATE INDEX IF NOT EXISTS idx_tenant_admin_assignments_user_id ON platform.tenant_admin_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_tenant_admin_assignments_tenant_id ON platform.tenant_admin_assignments(tenant_id);

-- Comments for documentation
COMMENT ON TABLE platform.tenant_admin_assignments IS 'Maps dashboard users to tenants as tenant administrators';
COMMENT ON COLUMN platform.tenant_admin_assignments.user_id IS 'Reference to the platform.users table (dashboard user)';
COMMENT ON COLUMN platform.tenant_admin_assignments.tenant_id IS 'Reference to the platform.tenants table';
COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_by IS 'Dashboard user who assigned this admin role';
COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_at IS 'Timestamp when the admin assignment was created';

-- Helper function: Get all tenant IDs managed by a user
-- Returns all tenant IDs for instance admins, or assigned tenant IDs for tenant admins
CREATE OR REPLACE FUNCTION platform.user_managed_tenant_ids(p_user_id UUID) RETURNS UUID[] AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN '{}'::UUID[];
    END IF;
    
    -- Instance admins manage all tenants
    IF EXISTS (
        SELECT 1 FROM platform.users
        WHERE id = p_user_id
        AND role = 'instance_admin'
        AND (deleted_at IS NULL OR is_active = true)
    ) THEN
        RETURN ARRAY(
            SELECT id FROM platform.tenants WHERE deleted_at IS NULL
        );
    END IF;
    
    -- Regular users manage only their assigned tenants
    RETURN ARRAY(
        SELECT tenant_id 
        FROM platform.tenant_admin_assignments 
        WHERE user_id = p_user_id
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION platform.user_managed_tenant_ids(UUID) IS
'Returns array of tenant IDs that a dashboard user can manage. Instance admins get all tenants; others get only their assigned tenants. SECURITY DEFINER to bypass RLS.';

-- Update is_instance_admin to use platform.users instead of dashboard.users
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;
    
    RETURN EXISTS (
        SELECT 1 FROM platform.users pu
        WHERE pu.id = p_user_id
        AND pu.role = 'instance_admin'
        AND (pu.deleted_at IS NULL OR pu.is_active = true)
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin with global privileges. SECURITY DEFINER to bypass RLS. Updated to use platform.users schema.';
