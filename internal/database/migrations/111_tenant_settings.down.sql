-- Migration 110 Rollback: Remove tenant-level and instance-level settings

-- ============================================================================
-- STEP 1: Remove triggers
-- ============================================================================

DROP TRIGGER IF EXISTS instance_settings_updated_at ON platform.instance_settings;
drop TRIGGER IF EXISTS tenant_settings_updated_at on platform.tenant_settings;

-- ============================================================================
-- STEP 2: Remove RLS policies
-- ============================================================================

DROP POLICY IF EXISTS instance_settings_select ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_insert ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_update ON platform.instance_settings;
DROP POLICY IF EXISTS instance_settings_delete ON platform.instance_settings;

DROP POLICY IF EXISTS tenant_settings_select ON platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_insert on platform.tenant_settings;
DROP POLICY IF EXISTS tenant_settings_update ON platform.tenant_settings;
drop POLICY IF EXISTS tenant_settings_delete on platform.tenant_settings;

-- ============================================================================
-- STEP 3: Drop helper functions
-- ============================================================================

DROP FUNCTION IF EXISTS platform.is_setting_overridable(TEXT);
DROP FUNCTION IF EXISTS platform.get_setting(UUID, TEXT, JSONB);
DROP FUNCTION IF EXISTS platform.get_jsonb_path(JSONB, TEXT);
drop FUNCTION IF EXISTS platform.set_jsonb_path(JSONB, TEXT, JSONB);
drop FUNCTION IF EXISTS platform.delete_jsonb_path(JSONB, TEXT);
drop function IF EXISTS platform.get_all_settings(UUID);

-- ============================================================================
-- STEP 4: Drop tables
-- ============================================================================

-- Disable RLS before dropping tables
ALTER TABLE platform.instance_settings DISABLE ROW LEVEL SECURITY;
ALTER TABLE platform.tenant_settings DISABLE ROW LEVEL SECURITY;

-- Drop tables
DROP TABLE IF EXISTS platform.instance_settings;
DROP TABLE IF EXISTS platform.tenant_settings;

-- ============================================================================
-- STEP 5: Revoke permissions from tenant_service
-- ============================================================================

REVOKE SELECT ON platform.instance_settings FROM tenant_service;
REVOKE INSERT, UPDATE, DELETE ON platform.instance_settings FROM tenant_service;
REVOKE EXECUTE ON FUNCTION platform.is_setting_overridable(Text) FROM tenant_service;
REVOKE EXECUTE ON FUNCTION platform.get_setting(UUID, TEXT, JSONB) FROM tenant_service;
REVOKE EXECUTE on Function platform.get_jsonb_path(JSONB, TEXT) FROM tenant_service;
REVOKE EXECUTE on Function platform.set_jsonb_path(JSONB, TEXT, JSONB) FROM tenant_service;
REVOKE EXECUTE on Function platform.delete_jsonb_path(JSONB, TEXT) FROM tenant_service;
REVOKE EXECUTE ON Function platform.get_all_settings(UUID) FROM tenant_service;

-- ============================================================================
-- STEP 6: Remove indexes
-- ============================================================================

DROP INDEX IF EXISTS idx_instance_settings_single_row ON platform.instance_settings ((id IS NOT NULL));
DROP INDEX IF EXISTS idx_tenant_settings_tenant_id ON platform.tenant_settings(tenant_id);
DROP INDEX IF EXISTS idx_tenant_settings_settings ON platform.tenant_settings USING GIN (settings);
DROP INDEX IF EXISTS idx_instance_settings_settings ON platform.instance_settings USING GIN (settings);
