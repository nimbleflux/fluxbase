--
-- UNIFIED SERVICE KEYS
-- Creates platform.service_keys table consolidating auth.client_keys and auth.service_keys
-- Also creates platform.key_usage for usage tracking
--

-- Create platform schema if not exists
CREATE SCHEMA IF NOT EXISTS platform;

-- Create platform.tenants if not exists (needed for FK reference)
-- This mirrors the public.tenants structure for the platform schema
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

CREATE INDEX IF NOT EXISTS idx_platform_tenants_slug ON platform.tenants(slug);
CREATE INDEX IF NOT EXISTS idx_platform_tenants_is_default ON platform.tenants(is_default) WHERE is_default = true;
CREATE INDEX IF NOT EXISTS idx_platform_tenants_deleted_at ON platform.tenants(deleted_at) WHERE deleted_at IS NOT NULL;

-- Sync platform.tenants from public.tenants if it exists and platform.tenants is empty
-- Note: public.tenants may not exist if this is a fresh install or was already migrated
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'tenants')
       AND NOT EXISTS (SELECT 1 FROM platform.tenants LIMIT 1)
    THEN
        INSERT INTO platform.tenants (id, slug, name, is_default, metadata, created_at, updated_at, deleted_at)
        SELECT id, slug, name, is_default, metadata, created_at, updated_at, deleted_at
        FROM public.tenants;
    END IF;
END $$;

-- Create unified service_keys table
CREATE TABLE IF NOT EXISTS platform.service_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Key type determines the key's purpose and scope
    key_type TEXT NOT NULL CHECK (key_type IN ('anon', 'publishable', 'tenant_service', 'global_service')),
    
    -- Tenant association (NULL for global_service keys)
    tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE,
    
    -- Basic info
    name TEXT NOT NULL,
    description TEXT,
    
    -- Key material (never store raw keys)
    key_hash TEXT NOT NULL,
    key_prefix TEXT NOT NULL UNIQUE,
    
    -- User association (for publishable keys)
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    
    -- Permissions
    scopes TEXT[] DEFAULT ARRAY[]::TEXT[],
    allowed_namespaces TEXT[],
    
    -- Rate limiting
    rate_limit_per_minute INTEGER DEFAULT 60,
    
    -- Status
    is_active BOOLEAN DEFAULT true,
    is_config_managed BOOLEAN DEFAULT false,
    
    -- Revocation support
    revoked_at TIMESTAMPTZ,
    revoked_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    revocation_reason TEXT,
    
    -- Rotation support
    deprecated_at TIMESTAMPTZ,
    grace_period_ends_at TIMESTAMPTZ,
    replaced_by UUID REFERENCES platform.service_keys(id) ON DELETE SET NULL,
    
    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_tenant_id ON platform.service_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_key_type ON platform.service_keys(key_type);
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_key_prefix ON platform.service_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_user_id ON platform.service_keys(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_is_active ON platform.service_keys(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_revoked_at ON platform.service_keys(revoked_at) WHERE revoked_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_platform_service_keys_grace_period ON platform.service_keys(grace_period_ends_at)
    WHERE deprecated_at IS NOT NULL AND grace_period_ends_at IS NOT NULL;

-- Add comments
COMMENT ON TABLE platform.service_keys IS 'Unified API key management consolidating client keys and service keys with multi-tenant support';
COMMENT ON COLUMN platform.service_keys.key_type IS 'Type of key: anon (anonymous), publishable (user-scoped), tenant_service (tenant-level), global_service (instance-level)';
COMMENT ON COLUMN platform.service_keys.tenant_id IS 'Tenant this key belongs to. NULL for global_service keys.';
COMMENT ON COLUMN platform.service_keys.key_hash IS 'Bcrypt hash of the full key. Never store keys in plaintext.';
COMMENT ON COLUMN platform.service_keys.key_prefix IS 'First characters of the key for identification in logs';
COMMENT ON COLUMN platform.service_keys.user_id IS 'User who owns this key (for publishable keys only)';
COMMENT ON COLUMN platform.service_keys.scopes IS 'Array of scope strings defining what this key can access';
COMMENT ON COLUMN platform.service_keys.allowed_namespaces IS 'Array of table/view namespaces this key can access (NULL = all)';
COMMENT ON COLUMN platform.service_keys.is_active IS 'Whether this key is currently usable';
COMMENT ON COLUMN platform.service_keys.is_config_managed IS 'Whether this key was created from configuration file';
COMMENT ON COLUMN platform.service_keys.revoked_at IS 'When the key was emergency revoked (NULL if not revoked)';
COMMENT ON COLUMN platform.service_keys.revoked_by IS 'User who revoked the key';
COMMENT ON COLUMN platform.service_keys.revocation_reason IS 'Reason for emergency revocation';
COMMENT ON COLUMN platform.service_keys.deprecated_at IS 'When the key was marked for rotation';
COMMENT ON COLUMN platform.service_keys.grace_period_ends_at IS 'When the grace period for rotation ends';
COMMENT ON COLUMN platform.service_keys.replaced_by IS 'Reference to the replacement key (for rotation)';

-- Migrate data from auth.client_keys
-- user_id IS NULL → key_type = 'anon'
-- user_id IS NOT NULL → key_type = 'publishable'
-- revoked = true → is_active = false (inverted logic)
INSERT INTO platform.service_keys (
    id,
    key_type,
    tenant_id,
    name,
    description,
    key_hash,
    key_prefix,
    user_id,
    scopes,
    rate_limit_per_minute,
    is_active,
    revoked_at,
    revoked_by,
    created_at,
    updated_at,
    last_used_at,
    expires_at
)
SELECT 
    id,
    CASE 
        WHEN user_id IS NULL THEN 'anon'
        ELSE 'publishable'
    END,
    tenant_id,
    name,
    description,
    key_hash,
    key_prefix,
    user_id,
    scopes,
    rate_limit_per_minute,
    NOT revoked,  -- is_active is the inverse of revoked
    CASE WHEN revoked THEN COALESCE(revoked_at, updated_at) ELSE NULL END,
    revoked_by,
    created_at,
    updated_at,
    last_used_at,
    expires_at
FROM auth.client_keys;

-- Migrate data from auth.service_keys
-- All become key_type = 'global_service'
-- tenant_id = NULL
-- Map 'enabled' to 'is_active'
INSERT INTO platform.service_keys (
    id,
    key_type,
    tenant_id,
    name,
    description,
    key_hash,
    key_prefix,
    scopes,
    rate_limit_per_minute,
    is_active,
    revoked_at,
    revoked_by,
    revocation_reason,
    deprecated_at,
    grace_period_ends_at,
    replaced_by,
    created_at,
    created_by,
    last_used_at,
    expires_at
)
SELECT 
    id,
    'global_service',
    NULL,  -- global service keys have no tenant
    name,
    description,
    key_hash,
    key_prefix,
    scopes,
    1000,  -- default higher rate limit for service keys
    COALESCE(enabled, true) AND (revoked_at IS NULL),  -- is_active if enabled and not revoked
    revoked_at,
    revoked_by,
    revocation_reason,
    deprecated_at,
    grace_period_ends_at,
    replaced_by,
    created_at,
    created_by,
    last_used_at,
    expires_at
FROM auth.service_keys;

-- Create key_usage table (migrate from auth.client_key_usage)
CREATE TABLE IF NOT EXISTS platform.key_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL REFERENCES platform.service_keys(id) ON DELETE CASCADE,
    endpoint TEXT NOT NULL,
    method TEXT NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for key_usage
CREATE INDEX IF NOT EXISTS idx_platform_key_usage_key_id ON platform.key_usage(key_id);
CREATE INDEX IF NOT EXISTS idx_platform_key_usage_created_at ON platform.key_usage(created_at DESC);

-- Add comments
COMMENT ON TABLE platform.key_usage IS 'Usage tracking for service keys with request details';
COMMENT ON COLUMN platform.key_usage.key_id IS 'Reference to the service key used';

-- Migrate data from auth.client_key_usage
-- Note: This assumes client_key_usage.client_key_id maps to client_keys.id
-- which has been migrated to platform.service_keys
INSERT INTO platform.key_usage (
    id,
    key_id,
    endpoint,
    method,
    status_code,
    response_time_ms,
    created_at
)
SELECT 
    ku.id,
    ku.client_key_id,
    ku.endpoint,
    ku.method,
    ku.status_code,
    ku.response_time_ms,
    ku.created_at
FROM auth.client_key_usage ku
WHERE EXISTS (
    SELECT 1 FROM platform.service_keys sk WHERE sk.id = ku.client_key_id
);

-- Create trigger for updated_at
CREATE OR REPLACE FUNCTION update_platform_service_keys_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER platform_service_keys_updated_at
    BEFORE UPDATE ON platform.service_keys
    FOR EACH ROW
    EXECUTE FUNCTION update_platform_service_keys_updated_at();

-- Note: Old tables (auth.client_keys, auth.service_keys, auth.client_key_usage) 
-- are NOT dropped in this migration. They should be dropped in a later cleanup 
-- migration after verifying the new table works correctly.
