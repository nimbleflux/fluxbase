-- Create a restricted role for tenant migrations
-- This role has NOCREATEDB and NOCREATEROLE to prevent tenant admins from:
-- - Creating/dropping databases
-- - Creating/dropping/altering roles
-- - Connecting to other databases

-- Create the tenant migration role if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_migration_role') THEN
        CREATE ROLE tenant_migration_role NOLOGIN NOINHERIT NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
    END IF;
END
$$;

COMMENT ON ROLE tenant_migration_role IS 'Restricted role for tenant migrations - cannot create databases or roles';

-- Grant schema-level permissions to tenant_migration_role
-- These grants apply to existing objects; default privileges handle future objects

-- Public schema - full DDL and DML for tenant migrations
GRANT USAGE, CREATE ON SCHEMA public TO tenant_migration_role;
GRANT ALL ON ALL TABLES IN SCHEMA public TO tenant_migration_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA public TO tenant_migration_role;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO tenant_migration_role;

-- Set default privileges for future objects in public schema
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO tenant_migration_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO tenant_migration_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT EXECUTE ON FUNCTIONS TO tenant_migration_role;

-- Grant ability to set role to tenant_migration_role
-- This allows the application to switch to this role for tenant migrations
GRANT tenant_migration_role TO CURRENT_USER;
