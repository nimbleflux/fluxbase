--
-- MULTI-TENANCY: CORE TENANT TABLES
-- Creates the tenants and tenant_memberships tables for multi-tenant support
--

-- Create tenants table
CREATE TABLE IF NOT EXISTS tenants (
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
CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_is_default ON tenants(is_default) WHERE is_default = true;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Comments for documentation
COMMENT ON TABLE tenants IS 'Logical tenants within a Fluxbase instance for multi-tenancy support';
COMMENT ON COLUMN tenants.id IS 'Unique identifier for the tenant';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly identifier for the tenant (e.g., "acme-corp")';
COMMENT ON COLUMN tenants.name IS 'Display name for the tenant';
COMMENT ON COLUMN tenants.is_default IS 'True for the synthetic default tenant used for backward compatibility';
COMMENT ON COLUMN tenants.metadata IS 'Arbitrary metadata for the tenant (plan, settings, etc.)';
COMMENT ON COLUMN tenants.deleted_at IS 'Soft delete timestamp. NULL if tenant is active.';

-- Create tenant_memberships table
CREATE TABLE IF NOT EXISTS tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'tenant_member',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT tenant_memberships_unique UNIQUE(tenant_id, user_id),
    CONSTRAINT tenant_memberships_role_check CHECK (role IN ('tenant_admin', 'tenant_member'))
);

-- Indexes for tenant_memberships
CREATE INDEX idx_tenant_memberships_user_id ON tenant_memberships(user_id);
CREATE INDEX idx_tenant_memberships_tenant_id ON tenant_memberships(tenant_id);
CREATE INDEX idx_tenant_memberships_role ON tenant_memberships(role);

-- Comments for documentation
COMMENT ON TABLE tenant_memberships IS 'Maps users to tenants with specific roles for multi-tenant access control';
COMMENT ON COLUMN tenant_memberships.tenant_id IS 'Reference to the tenant';
COMMENT ON COLUMN tenant_memberships.user_id IS 'Reference to the auth.users table';
COMMENT ON COLUMN tenant_memberships.role IS 'User role within tenant: tenant_admin (manage members) or tenant_member (regular access)';

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_tenant_memberships_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenant_memberships_updated_at
    BEFORE UPDATE ON tenant_memberships
    FOR EACH ROW
    EXECUTE FUNCTION update_tenant_memberships_updated_at();

-- Create trigger for tenants updated_at
CREATE OR REPLACE FUNCTION update_tenants_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_tenants_updated_at();
