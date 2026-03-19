--
-- MULTI-TENANCY: CORE TENANT TABLES
-- Creates the tenants and tenant_memberships tables for multi-tenant support
-- Tables are created in the platform schema from the start
--

-- Create platform schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS platform;

-- Create tenants table in platform schema
CREATE TABLE IF NOT EXISTS platform.tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Indexes for tenants
CREATE INDEX IF NOT EXISTS idx_platform_tenants_slug ON platform.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_platform_tenants_is_default ON platform.tenants(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_platform_tenants_deleted_at ON platform.tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE platform.tenants IS 'Logical tenants within a Fluxbase instance for multi-tenancy support';
COMMENT ON COLUMN platform.tenants.id IS 'Unique identifier for the tenant';
COMMENT ON COLUMN platform.tenants.slug IS 'URL-friendly identifier for the tenant (e.g., "acme-corp")';
COMMENT ON COLUMN platform.tenants.name IS 'Display name for the tenant';
COMMENT ON COLUMN platform.tenants.is_default IS 'True for the synthetic default tenant used for backward compatibility';
COMMENT ON COLUMN platform.tenants.metadata IS 'Arbitrary metadata for the tenant (plan, settings, etc.)';
COMMENT ON COLUMN platform.tenants.deleted_at IS 'Soft delete timestamp. NULL if tenant is active.';

-- Create tenant_memberships table in platform schema
CREATE TABLE IF NOT EXISTS platform.tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES platform.tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'tenant_member',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT platform_tenant_memberships_unique UNIQUE(tenant_id, user_id),
    CONSTRAINT platform_tenant_memberships_role_check CHECK (role IN ('tenant_admin', 'tenant_member'))
);

-- Indexes for tenant_memberships
CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_user_id ON platform.tenant_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_tenant_id ON platform.tenant_memberships(tenant_id);
CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_role ON platform.tenant_memberships(role);

-- Comments for documentation
COMMENT ON TABLE platform.tenant_memberships IS 'Maps users to tenants with specific roles for multi-tenant access control';
COMMENT ON COLUMN platform.tenant_memberships.tenant_id IS 'Reference to the tenant';
COMMENT ON COLUMN platform.tenant_memberships.user_id IS 'Reference to the auth.users table';
COMMENT ON COLUMN platform.tenant_memberships.role IS 'User role within tenant: tenant_admin (manage members) or tenant_member (regular access)';

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_platform_tenant_memberships_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER platform_tenant_memberships_updated_at
    BEFORE UPDATE ON platform.tenant_memberships
    FOR EACH ROW
    EXECUTE FUNCTION update_platform_tenant_memberships_updated_at();

-- Create trigger for tenants updated_at
CREATE OR REPLACE FUNCTION update_platform_tenants_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER platform_tenants_updated_at
    BEFORE UPDATE ON platform.tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_platform_tenants_updated_at();
