--
-- MULTI-TENANCY: ROLLBACK MOVE TENANTS TO PLATFORM SCHEMA
-- Moves tenants and tenant_memberships tables from platform back to public schema
--

-- Move tenant_memberships table back to public schema
ALTER TABLE IF EXISTS platform.tenant_memberships SET SCHEMA public;

-- Move tenants table back in public schema
ALTER TABLE IF EXISTS platform.tenants SET SCHEMA public;

-- Comments will be moved automatically by PostgreSQL
