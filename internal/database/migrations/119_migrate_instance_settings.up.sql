--
-- TWO-LEVEL SETTINGS MODEL: MIGRATE INSTANCE-LEVEL SETTINGS
-- Move infrastructure/feature flag settings from app.settings to app.instance_settings
-- Update constraints on app.settings to support per-tenant uniqueness
--

DO $$
DECLARE
    migrated_count INTEGER;
    deleted_count INTEGER;
BEGIN
    -- Step 1: Copy instance-level settings to app.instance_settings
    -- Only copy if they don't already exist (avoid duplicates from migration 118)
    INSERT INTO app.instance_settings (key, value, value_type, category, description, is_public, is_secret, editable_by, metadata, created_by, updated_by, created_at, updated_at)
    SELECT
        s.key,
        s.value,
        s.value_type,
        s.category,
        s.description,
        s.is_public,
        s.is_secret,
        ARRAY['instance_admin']::TEXT[], -- Override editable_by to instance_admin only
        s.metadata,
        s.created_by,
        s.updated_by,
        s.created_at,
        s.updated_at
    FROM app.settings s
    WHERE s.key IN (
        'app.realtime.enabled',
        'app.storage.enabled',
        'app.functions.enabled',
        'app.ai.enabled',
        'app.rpc.enabled',
        'app.jobs.enabled',
        'app.email.enabled',
        'app.features.enable_jobs',
        'app.features.enable_ai',
        'app.features.enable_rpc',
        'setup_completed'
    )
    ON CONFLICT (key) DO UPDATE SET
        value = EXCLUDED.value,
        updated_at = NOW();

    GET DIAGNOSTICS migrated_count = ROW_COUNT;
    RAISE NOTICE 'Migrated % settings to instance_settings', migrated_count;

    -- Step 2: Remove migrated settings from app.settings
    DELETE FROM app.settings
    WHERE key IN (
        'app.realtime.enabled',
        'app.storage.enabled',
        'app.functions.enabled',
        'app.ai.enabled',
        'app.rpc.enabled',
        'app.jobs.enabled',
        'app.email.enabled',
        'app.features.enable_jobs',
        'app.features.enable_ai',
        'app.features.enable_rpc',
        'setup_completed'
    );

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RAISE NOTICE 'Removed % settings from app.settings', deleted_count;
END $$;

-- Step 3: Update unique constraint on app.settings
-- Drop the old unique constraint on key alone
ALTER TABLE app.settings DROP CONSTRAINT IF EXISTS app_settings_key_key;

-- Create a composite unique index (key, tenant_id)
-- This allows the same key to exist for different tenants
CREATE UNIQUE INDEX IF NOT EXISTS idx_app_settings_key_tenant ON app.settings(key, tenant_id);

-- Step 4: Add trigger to auto-update updated_at
CREATE OR REPLACE FUNCTION app.update_instance_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_instance_settings_updated_at ON app.instance_settings;
CREATE TRIGGER update_instance_settings_updated_at
    BEFORE UPDATE ON app.instance_settings
    FOR EACH ROW
    EXECUTE FUNCTION app.update_instance_settings_updated_at();

DO $$
BEGIN
    RAISE NOTICE 'Instance settings migration complete - app.settings now supports per-tenant uniqueness';
END $$;
