--
-- ROLLBACK: TENANT-SCOPED AUTHENTICATION PROVIDERS
--

-- ============================================
-- DROP HELPER FUNCTIONS
-- ============================================

DROP FUNCTION IF EXISTS platform.get_oauth_provider(TEXT, UUID);
DROP FUNCTION IF EXISTS platform.get_saml_provider(TEXT, UUID);

-- ============================================
-- RESTORE SAML SESSIONS RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS saml_sessions_instance_admin ON auth.saml_sessions;
DROP POLICY IF EXISTS saml_sessions_tenant_admin ON auth.saml_sessions;
DROP POLICY IF EXISTS saml_sessions_self ON auth.saml_sessions;

-- Restore original policies (instance admin only)
CREATE POLICY saml_sessions_instance_admin ON auth.saml_sessions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY saml_sessions_self ON auth.saml_sessions
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- REMOVE TENANT_ID FROM SAML SESSIONS
-- ============================================

ALTER TABLE auth.saml_sessions DROP COLUMN IF EXISTS tenant_id;
DROP INDEX IF EXISTS idx_saml_sessions_tenant_id;

-- ============================================
-- RESTORE SAML PROVIDERS RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS saml_providers_instance_admin ON auth.saml_providers;
DROP POLICY IF EXISTS saml_providers_tenant_admin ON auth.saml_providers;
DROP POLICY IF EXISTS saml_providers_read ON auth.saml_providers;

-- Restore original policies
CREATE POLICY saml_providers_instance_admin ON auth.saml_providers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY saml_providers_read ON auth.saml_providers
    FOR SELECT TO authenticated
    USING (enabled = true);

-- ============================================
-- REMOVE TENANT_ID FROM SAML PROVIDERS
-- ============================================

-- Drop the new unique constraint
DROP INDEX IF EXISTS idx_saml_providers_name_per_tenant;

-- Remove tenant_id column
ALTER TABLE auth.saml_providers DROP COLUMN IF EXISTS tenant_id;
DROP INDEX IF EXISTS idx_saml_providers_tenant_id;

-- Restore original unique constraint on name
CREATE UNIQUE INDEX IF NOT EXISTS auth_saml_providers_name_key ON auth.saml_providers(name);

-- ============================================
-- RESTORE OAUTH PROVIDERS RLS POLICIES
-- ============================================

DROP POLICY IF EXISTS oauth_providers_instance_admin ON platform.oauth_providers;
DROP POLICY IF EXISTS oauth_providers_tenant_admin ON platform.oauth_providers;
DROP POLICY IF EXISTS oauth_providers_read ON platform.oauth_providers;

-- Restore original policies
CREATE POLICY platform_oauth_providers_instance_admin ON platform.oauth_providers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY platform_oauth_providers_read ON platform.oauth_providers
    FOR SELECT TO authenticated
    USING (enabled = true);

-- ============================================
-- REMOVE TENANT_ID FROM OAUTH PROVIDERS
-- ============================================

-- Drop the new unique constraint
DROP INDEX IF EXISTS idx_oauth_providers_name_per_tenant;

-- Remove tenant_id column
ALTER TABLE platform.oauth_providers DROP COLUMN IF EXISTS tenant_id;
DROP INDEX IF EXISTS idx_oauth_providers_tenant_id;

-- Restore original unique constraint on provider_name
CREATE UNIQUE INDEX IF NOT EXISTS platform_oauth_providers_provider_name_key ON platform.oauth_providers(provider_name);
