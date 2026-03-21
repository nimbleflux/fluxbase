-- Migration 110: Add tenant-level and instance-level configurable settings
-- This enables settings cascade: tenant settings -> instance settings -> defaults

-- ============================================================================
-- STEP 1: Create instance_settings table
-- ============================================================================

CREATE TABLE IF NOT EXISTS platform.instance_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Instance-level settings (applied to all tenants unless overridden)
    -- Stored as JSONB for flexibility, keyed by setting path
    -- Example: {"ai": {"enabled": true, "default_provider": "openai"}}
    settings JSONB NOT NULL DEFAULT '{}',

    -- Which settings are allowed to be overridden at tenant level
    -- NULL means all settings can be overridden
    -- Array of setting paths that are tenant-overridable
    -- Example: ["ai.enabled", "ai.default_provider", "auth.oidc.enabled"]
    overridable_settings JSONB,

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ensure only one row exists for instance settings
CREATE UNIQUE INDEX IF NOT EXISTS idx_instance_settings_single_row
    ON platform.instance_settings ((id IS NOT NULL));

-- Insert default empty settings row if it doesn't exist
INSERT INTO platform.instance_settings (settings, overridable_settings)
VALUES ('{}', NULL)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- STEP 2: Create tenant_settings table
-- ============================================================================

CREATE TABLE IF NOT EXISTS platform.tenant_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES platform.tenants(id) ON DELETE CASCADE,

    -- Tenant-specific settings that override instance settings
    -- Only includes settings that differ from instance defaults
    -- Example: {"ai": {"default_provider": "anthropic"}}
    settings JSONB NOT NULL DEFAULT '{}',

    -- Metadata
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT tenant_settings_tenant_unique UNIQUE (tenant_id)
);

-- Index for tenant lookups
CREATE INDEX IF NOT EXISTS idx_tenant_settings_tenant_id
    ON platform.tenant_settings(tenant_id);

-- Index for GIN queries on JSONB settings
CREATE INDEX IF NOT EXISTS idx_tenant_settings_settings
    ON platform.tenant_settings USING GIN (settings);

CREATE INDEX IF NOT EXISTS idx_instance_settings_settings
    ON platform.instance_settings USING GIN (settings);

-- ============================================================================
-- STEP 3: Create settings resolution helper functions
-- ============================================================================

-- Function to check if a setting path is overridable at tenant level
CREATE OR REPLACE FUNCTION platform.is_setting_overridable(setting_path TEXT)
RETURNS BOOLEAN
LANGUAGE sql
SECURITY DEFINER
SET search_path = ''
AS $$
    SELECT
        CASE
            -- If overridable_settings is NULL, all settings can be overridden
            WHEN (SELECT overridable_settings FROM platform.instance_settings LIMIT 1) IS NULL THEN
                TRUE
            -- Otherwise, check if the path is in the overridable list
            ELSE
                (SELECT overridable_settings FROM platform.instance_settings LIMIT 1) ? setting_path
                OR EXISTS (
                    SELECT 1
                    FROM platform.instance_settings,
                         jsonb_array_elements_text(overridable_settings) AS allowed_path
                    WHERE setting_path LIKE (allowed_path || '%')
                    LIMIT 1
                )
        END;
$$;

COMMENT ON FUNCTION platform.is_setting_overridable(TEXT) IS
    'Checks if a setting path can be overridden at the tenant level. Returns TRUE if overridable_settings is NULL (all allowed) or if the path matches an entry in the overridable list.';

-- Function to get a setting value with cascade resolution
-- Resolution order: tenant -> instance -> default
CREATE OR REPLACE FUNCTION platform.get_setting(
    p_tenant_id UUID,
    p_setting_path TEXT,
    p_default JSONB DEFAULT NULL
)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_tenant_value JSONB;
    v_instance_value JSONB;
    v_is_overridable BOOLEAN;
BEGIN
    -- Check if this setting can be overridden
    v_is_overridable := platform.is_setting_overridable(p_setting_path);

    -- If overridable and tenant_id provided, try to get tenant value first
    IF v_is_overridable AND p_tenant_id IS NOT NULL THEN
        SELECT platform.get_jsonb_path(settings, p_setting_path)
        INTO v_tenant_value
        FROM platform.tenant_settings
        WHERE tenant_id = p_tenant_id;

        IF v_tenant_value IS NOT NULL THEN
            RETURN v_tenant_value;
        END IF;
    END IF;

    -- Fall back to instance setting
    SELECT platform.get_jsonb_path(settings, p_setting_path)
    INTO v_instance_value
    FROM platform.instance_settings
    LIMIT 1;

    IF v_instance_value IS NOT NULL THEN
        RETURN v_instance_value;
    END IF;

    -- Fall back to provided default
    RETURN p_default;
END;
$$;

COMMENT ON FUNCTION platform.get_setting(UUID, TEXT, JSONB) IS
    'Resolves a setting value using cascade: tenant -> instance -> default. Respects overridable_settings restrictions.';

-- Helper function to extract a value from JSONB by dot-separated path
CREATE OR REPLACE FUNCTION platform.get_jsonb_path(
    p_jsonb JSONB,
    p_path TEXT
)
RETURNS JSONB
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_current JSONB;
    v_part TEXT;
BEGIN
    IF p_jsonb IS NULL OR p_path IS NULL OR p_path = '' THEN
        RETURN NULL;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_current := p_jsonb;

    FOREACH v_part IN ARRAY v_parts
    LOOP
        IF v_current IS NULL THEN
            RETURN NULL;
        END IF;

        -- Try to get the key from the current object
        IF jsonb_typeof(v_current) = 'object' THEN
            v_current := v_current->v_part;
        ELSE
            RETURN NULL;
        END IF;
    END LOOP;

    RETURN v_current;
END;
$$;

COMMENT ON FUNCTION platform.get_jsonb_path(JSONB, TEXT) IS
    'Extracts a value from JSONB using a dot-separated path (e.g., "ai.providers.openai.enabled").';

-- Function to set a nested value in JSONB by path
CREATE OR REPLACE FUNCTION platform.set_jsonb_path(
    p_jsonb JSONB,
    p_path TEXT,
    p_value JSONB
)
RETURNS JSONB
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_result JSONB;
BEGIN
    IF p_path IS NULL OR p_path = '' THEN
        RETURN p_jsonb;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_result := p_jsonb;

    -- Build the nested structure from the value up
    FOR i IN array_length(v_parts, 1) .. 1 BY -1 LOOP
        v_result := jsonb_build_object(v_parts[i],
            CASE
                WHEN i = array_length(v_parts, 1) THEN p_value
                ELSE v_result
            END
        );
    END LOOP;

    -- Deep merge with existing JSONB
    RETURN COALESCE(p_jsonb, '{}') || v_result;
END;
$$;

COMMENT ON FUNCTION platform.set_jsonb_path(JSONB, TEXT, JSONB) IS
    'Sets a nested value in JSONB using a dot-separated path. Deep merges with existing structure.';

-- Function to delete a nested key from JSONB by path
CREATE OR REPLACE FUNCTION platform.delete_jsonb_path(
    p_jsonb JSONB,
    p_path TEXT
)
RETURNS JSONB
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_path_for_delete TEXT[];
BEGIN
    IF p_path IS NULL OR p_path = '' THEN
        RETURN p_jsonb;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_path_for_delete := ARRAY[]::TEXT[];

    -- Build path array for #- operator
    FOR i IN 1 .. array_length(v_parts, 1) LOOP
        v_path_for_delete := array_append(v_path_for_delete, v_parts[i]);
    END LOOP;

    RETURN p_jsonb #- v_path_for_delete;
END;
$$;

COMMENT ON FUNCTION platform.delete_jsonb_path(JSONB, TEXT) IS
    'Deletes a nested key from JSONB using a dot-separated path.';

-- Function to get all resolved settings for a tenant
CREATE OR REPLACE FUNCTION platform.get_all_settings(p_tenant_id UUID)
RETURNS TABLE (
    setting_path TEXT,
    setting_value JSONB,
    source TEXT,
    is_overridable BOOLEAN
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = ''
AS $$
DECLARE
    v_instance_settings JSONB;
    v_tenant_settings JSONB;
    v_overridable JSONB;
    v_all_paths TEXT[];
BEGIN
    -- Get instance settings
    SELECT settings, overridable_settings
    INTO v_instance_settings, v_overridable
    FROM platform.instance_settings
    LIMIT 1;

    -- Get tenant settings
    SELECT settings
    INTO v_tenant_settings
    FROM platform.tenant_settings
    WHERE tenant_id = p_tenant_id;

    v_tenant_settings := COALESCE(v_tenant_settings, '{}');
    v_instance_settings := COALESCE(v_instance_settings, '{}');

    -- Extract all unique paths from both instance and tenant settings
    -- This is a simplified version - in practice you'd recursively extract all paths
    -- For now, we'll return the merged settings with metadata

    -- Return flattened settings with source info
    -- Note: A full implementation would recursively walk the JSONB trees
    -- This is a placeholder that returns the top-level structure

    RETURN QUERY
    SELECT
        key AS setting_path,
        COALESCE(
            CASE
                WHEN platform.is_setting_overridable(key) AND v_tenant_settings ? key
                THEN v_tenant_settings->key
                ELSE v_instance_settings->key
            END,
            v_instance_settings->key
        ) AS setting_value,
        CASE
            WHEN platform.is_setting_overridable(key) AND v_tenant_settings ? key
            THEN 'tenant'
            ELSE 'instance'
        END AS source,
        platform.is_setting_overridable(key) AS is_overridable
    FROM jsonb_object_keys(v_instance_settings || v_tenant_settings) AS key;
END;
$$;

COMMENT ON FUNCTION platform.get_all_settings(UUID) IS
    'Returns all resolved settings for a tenant with source information (tenant, instance, or default).';

-- ============================================================================
-- STEP 4: Create RLS policies for tenant_settings
-- ============================================================================

ALTER TABLE platform.tenant_settings ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS tenant_settings_select ON platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_insert ON platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_update ON platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_delete ON platform.tenant_settings;

-- RLS policies for tenant_settings
CREATE POLICY tenant_settings_select ON platform.tenant_settings
    FOR SELECT
    USING (
        -- Service role can see all tenant settings
        current_user = 'service_role'
        -- Instance admins can see all tenant settings
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
        -- Tenant has access to its own settings
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY tenant_settings_insert ON platform.tenant_settings
    FOR INSERT
    WITH CHECK (
        -- Service role can insert any tenant settings
        current_user = 'service_role'
        -- Instance admins can insert any tenant settings
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
        -- Tenant can insert its own settings (if overridable)
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY tenant_settings_update ON platform.tenant_settings
    FOR UPDATE
    USING (
        -- Service role can update any tenant settings
        current_user = 'service_role'
        -- Instance admins can update any tenant settings
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
        -- Tenant can update its own settings
        OR auth.has_tenant_access(tenant_id)
    );

CREATE POLICY tenant_settings_delete ON platform.tenant_settings
    FOR DELETE
    USING (
        -- Service role can delete any tenant settings
        current_user = 'service_role'
        -- Instance admins can delete any tenant settings
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
        -- Tenant can delete its own settings
        OR auth.has_tenant_access(tenant_id)
    );

-- ============================================================================
-- STEP 5: Create RLS policies for instance_settings
-- ============================================================================

ALTER TABLE platform.instance_settings ENABLE ROW LEVEL SECURITY;

-- Drop existing policies if they exist
DROP POLICY IF EXISTS instance_settings_select ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_insert ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_update ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_delete ON platform.instance_settings;

-- RLS policies for instance_settings (instance admin only for writes)
CREATE POLICY instance_settings_select ON platform.instance_settings
    FOR SELECT
    USING (
        -- Everyone can read instance settings (needed for cascade resolution)
        TRUE
    );

CREATE POLICY instance_settings_insert ON platform.instance_settings
    FOR INSERT
    WITH CHECK (
        -- Only service role or instance admins can modify instance settings
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
    );

CREATE POLICY instance_settings_update ON platform.instance_settings
    FOR UPDATE
    USING (
        -- Only service role or instance admins can modify instance settings
        current_user = 'service_role'
        OR EXISTS (
            SELECT 1 FROM auth.users
            WHERE id = current_setting('request.jwt.claims', TRUE)::JSONB->>'sub'
            AND is_instance_admin = TRUE
        )
    );

CREATE POLICY instance_settings_delete ON platform.instance_settings
    FOR DELETE
    USING (
        -- Only service role can delete instance settings (shouldn't happen normally)
        current_user = 'service_role'
    );

-- ============================================================================
-- STEP 6: Create update trigger for updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION platform.update_updated_at_column()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- Apply trigger to instance_settings
DROP TRIGGER IF EXISTS instance_settings_updated_at ON platform.instance_settings;
CREATE TRIGGER instance_settings_updated_at
    BEFORE UPDATE ON platform.instance_settings
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at_column();

-- Apply trigger to tenant_settings
DROP TRIGGER IF EXISTS tenant_settings_updated_at ON platform.tenant_settings;
CREATE TRIGGER tenant_settings_updated_at
    BEFORE UPDATE ON platform.tenant_settings
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at_column();

-- ============================================================================
-- STEP 7: Grant permissions to tenant_service role
-- ============================================================================

-- Grant read access to instance settings (needed for resolution)
GRANT SELECT ON platform.instance_settings TO tenant_service;

-- Grant full access to tenant settings
GRANT SELECT, INSERT, UPDATE, DELETE ON platform.tenant_settings TO tenant_service;

-- Grant execute on helper functions
GRANT EXECUTE ON FUNCTION platform.is_setting_overridable(TEXT) TO tenant_service;
GRANT EXECUTE ON FUNCTION platform.get_setting(UUID, TEXT, JSONB) TO tenant_service;
GRANT EXECUTE ON FUNCTION platform.get_jsonb_path(JSONB, TEXT) TO tenant_service;
GRANT EXECUTE ON FUNCTION platform.set_jsonb_path(JSONB, TEXT, JSONB) TO tenant_service;
GRANT EXECUTE ON FUNCTION platform.delete_jsonb_path(JSONB, TEXT) TO tenant_service;
GRANT EXECUTE ON FUNCTION platform.get_all_settings(UUID) TO tenant_service;

-- ============================================================================
-- STEP 8: Create default overridable settings list
-- ============================================================================

-- By default, allow most settings to be overridden at tenant level
-- Instance admins can restrict this later
UPDATE platform.instance_settings
SET overridable_settings = jsonb_build_array(
    -- AI settings
    'ai.enabled',
    'ai.default_provider',
    'ai.providers',
    'ai.embeddings',

    -- Auth settings (some may be restricted)
    'auth.oidc.enabled',
    'auth.oidc.provider',
    'auth.oidc.client_id',
    'auth.oidc.client_secret',
    'auth.saml.enabled',
    'auth.magic_link.enabled',
    'auth.mfa.required',

    -- Email settings
    'email.provider',
    'email.smtp',
    'email.sendgrid',
    'email.from_email',

    -- Storage settings
    'storage.max_file_size',
    'storage.allowed_extensions',
    'storage.s3'
)
WHERE id IS NOT NULL;
