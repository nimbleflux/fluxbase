-- Migration: 106_tenant_service_role
-- Description: Create tenant_service role for multi-tenant data access
-- This role is used by the application to access tenant-scoped data with RLS enforcement

-- Create the tenant_service role if it doesn't exist
-- NOLOGIN: Cannot be used for direct database connections
-- NOINHERIT: Does not inherit privileges from other roles (explicit grants only)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'tenant_service') THEN
        CREATE ROLE tenant_service NOLOGIN NOINHERIT;
    END IF;
END $$;

-- Grant USAGE on schemas
GRANT USAGE ON SCHEMA auth TO tenant_service;
GRANT USAGE ON SCHEMA storage TO tenant_service;
GRANT USAGE ON SCHEMA functions TO tenant_service;
GRANT USAGE ON SCHEMA jobs TO tenant_service;
GRANT USAGE ON SCHEMA ai TO tenant_service;
GRANT USAGE ON SCHEMA public TO tenant_service;

-- Grant table privileges
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA auth TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA storage TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA functions TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA jobs TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA ai TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO tenant_service;

-- Grant sequence privileges
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA auth TO tenant_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA storage TO tenant_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA functions TO tenant_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA jobs TO tenant_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA ai TO tenant_service;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO tenant_service;

-- Grant the role to fluxbase_app so the application can SET ROLE tenant_service
GRANT tenant_service TO fluxbase_app;

-- Add documentation comment
COMMENT ON ROLE tenant_service IS 'Role for tenant-scoped data access. Used by the application with SET ROLE to enforce row-level security for multi-tenant isolation. Grants read/write access to auth, storage, functions, jobs, ai, and public schemas.';
