--
-- ROLLBACK: TWO-LEVEL SETTINGS MODEL
-- Drop the instance_settings table
--

-- Drop RLS policies first
DROP POLICY IF EXISTS instance_settings_service_all ON app.instance_settings;
DROP POLICY IF EXISTS instance_settings_instance_admin ON app.instance_settings;
DROP POLICY IF EXISTS instance_settings_public_read ON app.instance_settings;

-- Drop indexes
DROP INDEX IF EXISTS idx_instance_settings_key;
DROP INDEX IF EXISTS idx_instance_settings_category;
DROP INDEX IF EXISTS idx_instance_settings_is_public;

-- Drop the table
DROP TABLE IF EXISTS app.instance_settings;

DO $$
BEGIN
    RAISE NOTICE 'Instance settings table dropped';
END $$;
