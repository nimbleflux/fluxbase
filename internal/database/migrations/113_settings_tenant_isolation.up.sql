--
-- MULTI-TENANCY: SETTINGS TENANT ISOLATION
-- Add tenant_id to app.settings and create RLS policies
--

DO $$
DECLARE
    default_tenant_id UUID;
BEGIN
    -- Get default tenant ID
    SELECT id INTO default_tenant_id
    FROM platform.tenants
    WHERE is_default = true AND deleted_at IS NULL
    LIMIT 1;

    IF default_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Default tenant not found. Run migration 101 first.';
    END IF;

    RAISE NOTICE 'Adding tenant_id to app.settings with default tenant: %', default_tenant_id;

    -- Step 1: Add tenant_id column (nullable first)
    ALTER TABLE app.settings
    ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES platform.tenants(id) ON DELETE CASCADE;

    -- Step 2: Backfill existing settings to default tenant
    UPDATE app.settings
    SET tenant_id = default_tenant_id
    WHERE tenant_id IS NULL;

    -- Step 3: Set default for new settings (use dynamic SQL since variables aren't allowed in DEFAULT)
    EXECUTE format('ALTER TABLE app.settings ALTER COLUMN tenant_id SET DEFAULT %L', default_tenant_id);

    -- Step 4: Add NOT NULL constraint
    ALTER TABLE app.settings
    ALTER COLUMN tenant_id SET NOT NULL;

    -- Step 5: Create index for tenant lookups
    CREATE INDEX IF NOT EXISTS idx_app_settings_tenant_id ON app.settings(tenant_id);

    RAISE NOTICE 'Tenant ID added to app.settings';
END $$;

-- ============================================
-- RLS POLICIES FOR APP.SETTINGS
-- ============================================

-- Enable RLS
ALTER TABLE app.settings ENABLE ROW LEVEL SECURITY;

-- Drop existing policies (if any)
DROP POLICY IF EXISTS settings_service_all ON app.settings;
DROP POLICY IF EXISTS settings_anon_read ON app.settings;

-- Service role bypasses all
CREATE POLICY settings_service_all ON app.settings
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all settings
CREATE POLICY settings_instance_admin ON app.settings
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Tenant admins can manage their tenant's settings
CREATE POLICY settings_tenant_admin ON app.settings
    FOR ALL TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Tenant service role can access tenant's settings
CREATE POLICY settings_tenant_service ON app.settings
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- Public settings can be read by anyone
CREATE POLICY settings_public_read ON app.settings
    FOR SELECT TO anon, authenticated
    USING (is_public = true);

-- Log completion
DO $$
BEGIN
    RAISE NOTICE 'Settings tenant isolation complete';
END $$;
