--
-- MULTI-TENANCY: ROLLBACK CORE TENANT TABLES
--

-- Drop triggers
DROP TRIGGER IF EXISTS tenants_updated_at ON tenants;
DROP TRIGGER IF EXISTS tenant_memberships_updated_at ON tenant_memberships;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_tenants_updated_at();
DROP FUNCTION IF EXISTS update_tenant_memberships_updated_at();

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS tenant_memberships;
DROP TABLE IF EXISTS tenants;
