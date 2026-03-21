-- Revoke and drop the tenant migration role

-- Revoke schema permissions
REVOKE USAGE ON SCHEMA public FROM tenant_migration_role;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM tenant_migration_role;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM tenant_migration_role;
REVOKE ALL ON ALL FUNCTIONS IN SCHEMA public FROM tenant_migration_role;

-- Revoke from application user
REVOKE tenant_migration_role FROM CURRENT_USER;

-- Drop the role
DROP ROLE IF EXISTS tenant_migration_role;
