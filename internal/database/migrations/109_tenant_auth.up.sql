-- Migration 109: Add tenant_id to auth tables for multi-tenancy
-- This enables multiple separate websites to use the same Fluxbase backend
-- with complete tenant isolation in auth.users and auth.service_keys

-- ============================================================================
-- STEP 1: Add tenant_id columns to auth tables
-- ============================================================================

-- Add tenant_id to auth.users
ALTER TABLE auth.users
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- Add tenant_id to auth.service_keys
ALTER TABLE auth.service_keys
    ADD COLUMN IF NOT EXISTS tenant_id UUID;

-- ============================================================================
-- STEP 2: Create tenant access helper function
-- ============================================================================

-- Create or replace the auth.has_tenant_access function if it doesn't exist
CREATE OR REPLACE FUNCTION auth.has_tenant_access(resource_tenant_id UUID)
RETURNS BOOLEAN
LANGUAGE sql
SECURITY DEFINER
SET search_path = ''
AS $$
    -- Get the current tenant from session context
    SELECT CASE
        WHEN current_setting('app.current_tenant_id', TRUE) = '' THEN
            -- No tenant context set, check if resource is in default tenant (NULL)
            resource_tenant_id IS NULL
        ELSE
            -- Tenant context is set, check if resource belongs to same tenant
            resource_tenant_id::TEXT = current_setting('app.current_tenant_id', TRUE)
    END;
$$;

COMMENT ON FUNCTION auth.has_tenant_access(UUID) IS
    'Checks if the current tenant context has access to a resource. Returns TRUE if the resource_tenant_id matches the current tenant or if no tenant context is set and the resource is in the default tenant (NULL).';

-- ============================================================================
-- STEP 3: Handle backwards compatibility - existing data stays in default tenant
-- ============================================================================

-- Existing data with tenant_id = NULL belongs to the default tenant
-- No data migration needed - NULL means default tenant

-- ============================================================================
-- STEP 4: Create partial unique indexes for tenant-scoped uniqueness
-- ============================================================================

-- Drop the existing unique constraint on email (if exists as an index)
DROP INDEX IF EXISTS auth.users_email_unique;
DROP INDEX IF EXISTS auth.users_email_key;

-- Create partial unique indexes for email uniqueness per tenant
-- This allows the same email in different tenants
CREATE UNIQUE INDEX IF NOT EXISTS auth_users_email_tenant_null_unique
    ON auth.users (email)
    WHERE tenant_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS auth_users_email_tenant_unique
    ON auth.users (tenant_id, email)
    WHERE tenant_id IS NOT NULL;

-- Create partial unique indexes for service key name uniqueness per tenant
CREATE UNIQUE INDEX IF NOT EXISTS auth_service_keys_name_tenant_null_unique
    ON auth.service_keys (name)
    WHERE tenant_id IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS auth_service_keys_name_tenant_unique
    ON auth.service_keys (tenant_id, name)
    WHERE tenant_id IS NOT NULL;

-- ============================================================================
-- STEP 5: Create triggers to auto-populate tenant_id from session context
-- ============================================================================

-- Function to auto-set tenant_id on INSERT
CREATE OR REPLACE FUNCTION auth.set_tenant_id_from_context()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
BEGIN
    -- Only set tenant_id if not already provided and context is available
    IF NEW.tenant_id IS NULL THEN
        BEGIN
            NEW.tenant_id := current_setting('app.current_tenant_id', TRUE)::UUID;
        EXCEPTION
            WHEN others THEN
                NEW.tenant_id := NULL;
        END;
    END IF;
    RETURN NEW;
END;
$$;

-- Apply trigger to auth.users
DROP TRIGGER IF EXISTS auth_users_set_tenant_id ON auth.users;
CREATE TRIGGER auth_users_set_tenant_id
    BEFORE INSERT ON auth.users
    FOR EACH ROW
    EXECUTE FUNCTION auth.set_tenant_id_from_context();

-- Apply trigger to auth.service_keys
DROP TRIGGER IF EXISTS auth_service_keys_set_tenant_id ON auth.service_keys;
CREATE TRIGGER auth_service_keys_set_tenant_id
    BEFORE INSERT ON auth.service_keys
    FOR EACH ROW
    EXECUTE FUNCTION auth.set_tenant_id_from_context();

-- ============================================================================
-- STEP 6: Create RLS policies for tenant isolation
-- ============================================================================

-- Enable RLS on auth.users
ALTER TABLE auth.users ENABLE ROW LEVEL SECURITY;

-- Enable RLS on auth.service_keys
ALTER TABLE auth.service_keys ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS auth_users_select ON auth.users;
DROP POLICY IF EXISTS auth_users_insert ON auth.users;
DROP POLICY IF EXISTS auth_users_update ON auth.users;
DROP POLICY IF EXISTS auth_users_delete ON auth.users;
DROP POLICY IF EXISTS auth_service_keys_select ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_insert ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_update ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_delete ON auth.service_keys;

-- RLS policies for auth.users
CREATE POLICY auth_users_select ON auth.users
    FOR SELECT
    USING (
        -- Service role can see all users
        current_user = 'service_role'
        -- Or user is accessing their own record
        OR id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
        -- Or has tenant access (tenant_service or tenant admin)
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_users_insert ON auth.users
    FOR INSERT
    WITH CHECK (
        -- Service role can insert any user
        current_user = 'service_role'
        -- Or has tenant access (tenant operations)
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_users_update ON auth.users
    FOR UPDATE
    USING (
        -- Service role can update any user
        current_user = 'service_role'
        -- Or user is updating their own record
        OR id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    )
    WITH CHECK (
        -- Service role can update any user
        current_user = 'service_role'
        -- Or user is updating their own record (but not tenant_id)
        OR id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_users_delete ON auth.users
    FOR DELETE
    USING (
        -- Service role can delete any user
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

-- RLS policies for auth.service_keys
CREATE POLICY auth_service_keys_select ON auth.service_keys
    FOR SELECT
    USING (
        -- Service role can see all keys
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_service_keys_insert ON auth.service_keys
    FOR INSERT
    WITH CHECK (
        -- Service role can insert any key
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_service_keys_update ON auth.service_keys
    FOR UPDATE
    USING (
        -- Service role can update any key
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    )
    WITH CHECK (
        -- Service role can update any key
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY auth_service_keys_delete ON auth.service_keys
    FOR DELETE
    USING (
        -- Service role can delete any key
        current_user = 'service_role'
        -- Or has tenant access
        OR auth.has_tenant_access(tenant_id)
    );

-- ============================================================================
-- STEP 7: Grant permissions to tenant_service role for tenant migrations
-- ============================================================================

-- Grant SELECT, INSERT, UPDATE, DELETE on auth tables to tenant_service
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.users TO tenant_service;
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.service_keys TO tenant_service;

-- Grant USAGE on sequences
GRANT USAGE, SELECT ON SEQUENCE auth.users_id_seq TO tenant_service;
GRANT USAGE, SELECT ON SEQUENCE auth.service_keys_id_seq TO tenant_service;

-- ============================================================================
-- STEP 8: Add foreign key constraints to platform.tenants
-- ============================================================================

-- Add foreign key from auth.users to platform.tenants (with SET NULL for safety)
ALTER TABLE auth.users
    ADD CONSTRAINT fk_auth_users_tenant
    FOREIGN KEY (tenant_id)
    REFERENCES platform.tenants(id)
    ON DELETE SET NULL;

-- Add foreign key from auth.service_keys to platform.tenants
ALTER TABLE auth.service_keys
    ADD CONSTRAINT fk_auth_service_keys_tenant
    FOREIGN KEY (tenant_id)
    REFERENCES platform.tenants(id)
    ON DELETE SET NULL;

-- ============================================================================
-- STEP 9: Add indexes for tenant_id lookups
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_auth_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_service_keys_tenant_id ON auth.service_keys(tenant_id);
