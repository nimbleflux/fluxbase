-- Rename dashboard schema to platform
-- This aligns the schema name with the product positioning (platform admin vs dashboard)

-- Only rename if dashboard exists and platform doesn't exist
-- (platform schema may have been created by migration 105)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'dashboard')
       AND NOT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'platform')
    THEN
        ALTER SCHEMA dashboard RENAME TO platform;
    END IF;
END $$;

-- Update comments on auth.* tables that reference the old schema name
-- These are safe to run even if platform schema was created directly
COMMENT ON COLUMN auth.mcp_oauth_codes.user_id IS 'Platform user who authorized this code (references platform.users)';
COMMENT ON COLUMN auth.mcp_oauth_tokens.user_id IS 'Platform user this token represents (references platform.users)';
