--
-- TWO-LEVEL SETTINGS MODEL: INSTANCE SETTINGS TABLE
-- Create separate table for platform-level settings that should not be tenant-customizable
--

-- Create instance-level settings table (no tenant_id - these are platform-wide)
CREATE TABLE IF NOT EXISTS app.instance_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT UNIQUE NOT NULL,
    value JSONB NOT NULL,
    value_type TEXT NOT NULL DEFAULT 'string'
        CHECK (value_type IN ('string', 'number', 'boolean', 'json', 'array')),
    category TEXT NOT NULL DEFAULT 'system'
        CHECK (category IN ('auth', 'system', 'storage', 'functions', 'realtime', 'ai', 'jobs', 'email', 'custom')),
    description TEXT,
    is_public BOOLEAN DEFAULT false,
    is_secret BOOLEAN DEFAULT false,
    editable_by TEXT[] NOT NULL DEFAULT ARRAY['instance_admin']::TEXT[],
    metadata JSONB DEFAULT '{}'::JSONB,
    created_by UUID,
    updated_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_instance_settings_key ON app.instance_settings(key);
CREATE INDEX IF NOT EXISTS idx_instance_settings_category ON app.instance_settings(category);
CREATE INDEX IF NOT EXISTS idx_instance_settings_is_public ON app.instance_settings(is_public);

-- Add comments
COMMENT ON TABLE app.instance_settings IS 'Platform-level settings that apply to the entire instance (not tenant-customizable)';
COMMENT ON COLUMN app.instance_settings.key IS 'Unique setting key (e.g., "app.realtime.enabled")';
COMMENT ON COLUMN app.instance_settings.value IS 'Setting value stored as JSONB';
COMMENT ON COLUMN app.instance_settings.category IS 'Category: auth, system, storage, functions, realtime, ai, jobs, or email';
COMMENT ON COLUMN app.instance_settings.is_public IS 'Whether this setting can be read by non-admin users';
COMMENT ON COLUMN app.instance_settings.is_secret IS 'Whether this setting contains sensitive data';
COMMENT ON COLUMN app.instance_settings.editable_by IS 'Array of roles that can edit this setting (typically instance_admin only)';

-- Enable RLS
ALTER TABLE app.instance_settings ENABLE ROW LEVEL SECURITY;

-- RLS Policies for instance_settings

-- Service role bypasses all
CREATE POLICY instance_settings_service_all ON app.instance_settings
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all instance settings
CREATE POLICY instance_settings_instance_admin ON app.instance_settings
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Public settings can be read by authenticated users (for feature availability checks)
CREATE POLICY instance_settings_public_read ON app.instance_settings
    FOR SELECT TO authenticated
    USING (is_public = true);

-- Insert default instance-level feature flags
-- These control which platform features are available
INSERT INTO app.instance_settings (key, value, value_type, category, description, is_public, editable_by)
VALUES
    (
        'app.realtime.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'realtime',
        'Enable or disable realtime functionality (WebSocket connections, subscriptions) - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.storage.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'storage',
        'Enable or disable storage functionality (file uploads, downloads) - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.functions.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'functions',
        'Enable or disable edge functions (serverless function execution) - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.ai.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'ai',
        'Enable AI service functionality - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.rpc.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'system',
        'Enable RPC procedure execution - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.jobs.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'jobs',
        'Enable background job processing - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.email.enabled',
        '{"value": true}'::JSONB,
        'boolean',
        'email',
        'Enable email service functionality - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.features.enable_jobs',
        '{"value": true}'::JSONB,
        'boolean',
        'jobs',
        'Feature flag for background jobs UI - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.features.enable_ai',
        '{"value": true}'::JSONB,
        'boolean',
        'ai',
        'Feature flag for AI features - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'app.features.enable_rpc',
        '{"value": true}'::JSONB,
        'boolean',
        'system',
        'Feature flag for RPC features - Platform level',
        true,
        ARRAY['instance_admin']::TEXT[]
    ),
    (
        'setup_completed',
        '{"value": false}'::JSONB,
        'boolean',
        'system',
        'Whether the initial setup has been completed',
        false,
        ARRAY['instance_admin']::TEXT[]
    )
ON CONFLICT (key) DO NOTHING;

DO $$
BEGIN
    RAISE NOTICE 'Instance settings table created with default feature flags';
END $$;
