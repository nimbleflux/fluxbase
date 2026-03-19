--
-- ROLLBACK: SETTINGS TENANT ISOLATION
--

-- Drop RLS policies
DROP POLICY IF EXISTS settings_service_all ON app.settings;
DROP POLICY IF EXISTS settings_instance_admin ON app.settings;
DROP POLICY IF EXISTS settings_tenant_admin ON app.settings;
DROP POLICY IF EXISTS settings_tenant_service ON app.settings;
DROP POLICY IF EXISTS settings_public_read ON app.settings;

-- Drop index
DROP INDEX IF EXISTS idx_app_settings_tenant_id;

-- Remove NOT NULL constraint
ALTER TABLE app.settings
ALTER COLUMN tenant_id DROP NOT NULL;

-- Set all to NULL (preserving data)
UPDATE app.settings SET tenant_id = NULL;

-- Drop default
ALTER TABLE app.settings
ALTER COLUMN tenant_id DROP DEFAULT;

-- Drop column
ALTER TABLE app.settings DROP COLUMN IF EXISTS tenant_id;

DO $$
BEGIN
    RAISE NOTICE 'Settings tenant isolation rolled back';
END $$;
