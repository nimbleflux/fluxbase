--
-- ROLLBACK: MIGRATE INSTANCE-LEVEL SETTINGS
-- Move settings back from app.instance_settings to app.settings
-- Restore original unique constraint
--

DO $$
DECLARE
    default_tenant_id UUID;
    migrated_count INTEGER;
BEGIN
    -- Get default tenant ID
    SELECT id INTO default_tenant_id
    FROM platform.tenants
    WHERE is_default = true AND deleted_at IS NULL
    LIMIT 1;

    IF default_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Default tenant not found';
    END IF;

    -- Step 1: Move instance-level settings back to app.settings
    INSERT INTO app.settings (key, value, value_type, category, description, is_public, is_secret, editable_by, metadata, tenant_id, created_by, updated_by, created_at, updated_at)
    SELECT
        s.key,
        s.value,
        s.value_type,
        s.category,
        s.description,
        s.is_public,
        s.is_secret,
        ARRAY['admin', 'dashboard_admin']::TEXT[], -- Restore original editable_by
        s.metadata,
        default_tenant_id,
        s.created_by,
        s.updated_by,
        s.created_at,
        s.updated_at
    FROM app.instance_settings s
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
    RAISE NOTICE 'Moved % settings back to app.settings', migrated_count;
END $$;

-- Step 2: Remove the composite unique index
DROP INDEX IF EXISTS idx_app_settings_key_tenant;

-- Step 3: Restore original unique constraint on key alone
-- Note: This may fail if there are duplicate keys across tenants
CREATE UNIQUE INDEX IF NOT EXISTS app_settings_key_key ON app.settings(key);

-- Step 4: Drop the trigger and function
DROP TRIGGER IF EXISTS update_instance_settings_updated_at ON app.instance_settings;
DROP FUNCTION IF EXISTS app.update_instance_settings_updated_at();

DO $$
BEGIN
    RAISE NOTICE 'Rollback complete - settings moved back to app.settings';
END $$;
