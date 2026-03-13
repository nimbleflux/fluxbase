--
-- MULTI-TENANCY: ROLLBACK DASHBOARD ROLES
--

-- Drop the index
DROP INDEX IF EXISTS idx_dashboard_users_role_instance_admin;

-- Revert role constraint to original
ALTER TABLE dashboard.users DROP CONSTRAINT IF EXISTS dashboard_users_role_check;
ALTER TABLE dashboard.users ADD CONSTRAINT dashboard_users_role_check 
    CHECK (role IN ('dashboard_admin', 'dashboard_user'));

-- Revert comment
COMMENT ON COLUMN dashboard.users.role IS 'User role: dashboard_admin or dashboard_user';
