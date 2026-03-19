--
-- MULTI-TENANCY: UPDATE DASHBOARD ROLES
-- Adds instance_admin and tenant_admin roles to dashboard.users constraint
--

-- Update role constraint to include new multi-tenant roles
ALTER TABLE dashboard.users DROP CONSTRAINT IF EXISTS dashboard_users_role_check;
ALTER TABLE dashboard.users ADD CONSTRAINT dashboard_users_role_check
    CHECK (role IN ('instance_admin', 'tenant_admin', 'dashboard_admin', 'dashboard_user'));

-- Update comments
COMMENT ON COLUMN dashboard.users.role IS
    'User role: instance_admin (global admin managing all tenants), tenant_admin (admin for specific tenant), dashboard_admin (legacy, maps to tenant_admin), dashboard_user (limited read-only access)';

-- Create index for role lookups
CREATE INDEX IF NOT EXISTS idx_dashboard_users_role_instance_admin ON dashboard.users(role) WHERE role = 'instance_admin';
