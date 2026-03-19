--
-- TENANT-SCOPED AUTHENTICATION PROVIDERS
-- Adds tenant_id to OAuth and SAML providers for multi-tenant isolation
-- NULL tenant_id = platform-level (for dashboard admin login, shared across tenants)
-- Non-NULL tenant_id = tenant-specific (only available to that tenant)
--

-- ============================================
-- ADD TENANT_ID TO OAUTH PROVIDERS
-- ============================================

-- Add tenant_id column (nullable for platform-level providers)
ALTER TABLE platform.oauth_providers 
ADD COLUMN tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Create index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_oauth_providers_tenant_id ON platform.oauth_providers(tenant_id);

-- Drop the existing unique constraint on provider_name
ALTER TABLE platform.oauth_providers DROP CONSTRAINT IF EXISTS dashboard_oauth_providers_provider_name_key;
ALTER TABLE platform.oauth_providers DROP CONSTRAINT IF EXISTS platform_oauth_providers_provider_name_key;
DROP INDEX IF EXISTS platform_oauth_providers_provider_name_key;

-- Create new unique constraint: provider_name must be unique per tenant
-- This allows different tenants to have providers with the same name (e.g., both can have "google")
-- Platform-level providers (tenant_id = NULL) also get unique names
CREATE UNIQUE INDEX idx_oauth_providers_name_per_tenant 
    ON platform.oauth_providers(provider_name, tenant_id)
    WHERE deleted_at IS NULL;

-- Add comment
COMMENT ON COLUMN platform.oauth_providers.tenant_id IS 
'Tenant this provider belongs to. NULL = platform-level (for dashboard admin login). Non-NULL = tenant-specific (only available to that tenant).';

-- ============================================
-- ADD TENANT_ID TO SAML PROVIDERS
-- ============================================

-- Add tenant_id column (nullable for platform-level providers)
ALTER TABLE auth.saml_providers 
ADD COLUMN tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Create index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_saml_providers_tenant_id ON auth.saml_providers(tenant_id);

-- Drop the existing unique constraint on name
ALTER TABLE auth.saml_providers DROP CONSTRAINT IF EXISTS saml_providers_name_key;
ALTER TABLE auth.saml_providers DROP CONSTRAINT IF EXISTS auth_saml_providers_name_key;
DROP INDEX IF EXISTS auth_saml_providers_name_key;

-- Create new unique constraint: name must be unique per tenant
CREATE UNIQUE INDEX idx_saml_providers_name_per_tenant 
    ON auth.saml_providers(name, tenant_id);

-- Add comment
COMMENT ON COLUMN auth.saml_providers.tenant_id IS 
'Tenant this provider belongs to. NULL = platform-level (for dashboard admin login). Non-NULL = tenant-specific (only available to that tenant).';

-- ============================================
-- UPDATE RLS POLICIES FOR OAUTH PROVIDERS
-- ============================================

-- Drop old policies
DROP POLICY IF EXISTS platform_oauth_providers_instance_admin ON platform.oauth_providers;
DROP POLICY IF EXISTS platform_oauth_providers_read ON platform.oauth_providers;

-- Instance admins can manage all providers (including platform-level)
-- But NOT when acting as tenant admin
CREATE POLICY oauth_providers_instance_admin ON platform.oauth_providers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's providers
CREATE POLICY oauth_providers_tenant_admin ON platform.oauth_providers
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Read policy: users can see their tenant's providers + platform-level providers
CREATE POLICY oauth_providers_read ON platform.oauth_providers
    FOR SELECT TO authenticated
    USING (
        is_instance_admin(auth.uid())
        OR tenant_id = current_tenant_id()
        OR tenant_id IS NULL
    );

-- ============================================
-- UPDATE RLS POLICIES FOR SAML PROVIDERS
-- ============================================

-- Drop old policies
DROP POLICY IF EXISTS saml_providers_instance_admin ON auth.saml_providers;
DROP POLICY IF EXISTS saml_providers_read ON auth.saml_providers;

-- Instance admins can manage all SAML providers
CREATE POLICY saml_providers_instance_admin ON auth.saml_providers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's SAML providers
CREATE POLICY saml_providers_tenant_admin ON auth.saml_providers
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Read policy: users can see their tenant's providers + platform-level providers
CREATE POLICY saml_providers_read ON auth.saml_providers
    FOR SELECT TO authenticated
    USING (
        is_instance_admin(auth.uid())
        OR tenant_id = current_tenant_id()
        OR tenant_id IS NULL
    );

-- ============================================
-- ADD TENANT_ID TO SAML SESSIONS (for consistency)
-- ============================================

-- Add tenant_id column to link sessions to tenant context
ALTER TABLE auth.saml_sessions 
ADD COLUMN tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

-- Create index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_saml_sessions_tenant_id ON auth.saml_sessions(tenant_id);

-- Update RLS policies for saml_sessions to include tenant isolation
DROP POLICY IF EXISTS saml_sessions_instance_admin ON auth.saml_sessions;
DROP POLICY IF EXISTS saml_sessions_self ON auth.saml_sessions;

CREATE POLICY saml_sessions_instance_admin ON auth.saml_sessions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY saml_sessions_tenant_admin ON auth.saml_sessions
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id() 
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

CREATE POLICY saml_sessions_self ON auth.saml_sessions
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- HELPER FUNCTION: GET OAUTH PROVIDER WITH TENANT FALLBACK
-- ============================================

CREATE OR REPLACE FUNCTION platform.get_oauth_provider(
    p_provider_name TEXT,
    p_tenant_id UUID DEFAULT NULL
) RETURNS TABLE (
    id UUID,
    provider_name TEXT,
    display_name TEXT,
    client_id TEXT,
    client_secret TEXT,
    redirect_url TEXT,
    scopes TEXT[],
    enabled BOOLEAN,
    is_custom BOOLEAN,
    authorization_url TEXT,
    token_url TEXT,
    user_info_url TEXT,
    allow_app_login BOOLEAN,
    allow_dashboard_login BOOLEAN,
    required_claims JSONB,
    denied_claims JSONB,
    provider_tenant_id UUID
) AS $$
BEGIN
    -- If tenant context provided, try tenant-specific provider first
    IF p_tenant_id IS NOT NULL THEN
        RETURN QUERY
        SELECT 
            op.id,
            op.provider_name,
            op.display_name,
            op.client_id,
            op.client_secret,
            op.redirect_url,
            op.scopes,
            op.enabled,
            op.is_custom,
            op.authorization_url,
            op.token_url,
            op.user_info_url,
            op.allow_app_login,
            op.allow_dashboard_login,
            op.required_claims,
            op.denied_claims,
            op.tenant_id
        FROM platform.oauth_providers op
        WHERE op.provider_name = p_provider_name
        AND op.tenant_id = p_tenant_id
        AND op.enabled = true
        LIMIT 1;
        
        -- If found, return
        IF FOUND THEN
            RETURN;
        END IF;
    END IF;
    
    -- Fallback to platform-level provider
    RETURN QUERY
    SELECT 
        op.id,
        op.provider_name,
        op.display_name,
        op.client_id,
        op.client_secret,
        op.redirect_url,
        op.scopes,
        op.enabled,
        op.is_custom,
        op.authorization_url,
        op.token_url,
        op.user_info_url,
        op.allow_app_login,
        op.allow_dashboard_login,
        op.required_claims,
        op.denied_claims,
        op.tenant_id
    FROM platform.oauth_providers op
    WHERE op.provider_name = p_provider_name
    AND op.tenant_id IS NULL
    AND op.enabled = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION platform.get_oauth_provider(TEXT, UUID) IS
'Gets an OAuth provider by name, checking tenant-specific providers first, then falling back to platform-level. SECURITY DEFINER to bypass RLS during lookup.';

-- ============================================
-- HELPER FUNCTION: GET SAML PROVIDER WITH TENANT FALLBACK
-- ============================================

CREATE OR REPLACE FUNCTION platform.get_saml_provider(
    p_provider_name TEXT,
    p_tenant_id UUID DEFAULT NULL
) RETURNS TABLE (
    id UUID,
    name TEXT,
    enabled BOOLEAN,
    idp_metadata_url TEXT,
    idp_metadata_xml TEXT,
    entity_id TEXT,
    acs_url TEXT,
    allow_app_login BOOLEAN,
    allow_dashboard_login BOOLEAN,
    provider_tenant_id UUID
) AS $$
BEGIN
    -- If tenant context provided, try tenant-specific provider first
    IF p_tenant_id IS NOT NULL THEN
        RETURN QUERY
        SELECT 
            sp.id,
            sp.name,
            sp.enabled,
            sp.idp_metadata_url,
            sp.idp_metadata_xml,
            sp.entity_id,
            sp.acs_url,
            sp.allow_app_login,
            sp.allow_dashboard_login,
            sp.tenant_id
        FROM auth.saml_providers sp
        WHERE sp.name = p_provider_name
        AND sp.tenant_id = p_tenant_id
        AND sp.enabled = true
        LIMIT 1;
        
        -- If found, return
        IF FOUND THEN
            RETURN;
        END IF;
    END IF;
    
    -- Fallback to platform-level provider
    RETURN QUERY
    SELECT 
        sp.id,
        sp.name,
        sp.enabled,
        sp.idp_metadata_url,
        sp.idp_metadata_xml,
        sp.entity_id,
        sp.acs_url,
        sp.allow_app_login,
        sp.allow_dashboard_login,
        sp.tenant_id
    FROM auth.saml_providers sp
    WHERE sp.name = p_provider_name
    AND sp.tenant_id IS NULL
    AND sp.enabled = true
    LIMIT 1;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION platform.get_saml_provider(TEXT, UUID) IS
'Gets a SAML provider by name, checking tenant-specific providers first, then falling back to platform-level. SECURITY DEFINER to bypass RLS during lookup.';
