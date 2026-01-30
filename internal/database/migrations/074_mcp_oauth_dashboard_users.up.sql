-- Fix MCP OAuth tables to reference dashboard.users instead of auth.users
-- Dashboard admins are the primary users who authorize MCP clients (e.g., Claude Desktop)

-- Drop existing foreign key constraints on mcp_oauth_codes
ALTER TABLE auth.mcp_oauth_codes
    DROP CONSTRAINT IF EXISTS mcp_oauth_codes_user_id_fkey;

-- Add new foreign key referencing dashboard.users
ALTER TABLE auth.mcp_oauth_codes
    ADD CONSTRAINT mcp_oauth_codes_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES dashboard.users(id) ON DELETE CASCADE;

-- Drop existing foreign key constraints on mcp_oauth_tokens
ALTER TABLE auth.mcp_oauth_tokens
    DROP CONSTRAINT IF EXISTS mcp_oauth_tokens_user_id_fkey;

-- Add new foreign key referencing dashboard.users
ALTER TABLE auth.mcp_oauth_tokens
    ADD CONSTRAINT mcp_oauth_tokens_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES dashboard.users(id) ON DELETE CASCADE;

COMMENT ON COLUMN auth.mcp_oauth_codes.user_id IS 'Dashboard user who authorized this code (references dashboard.users)';
COMMENT ON COLUMN auth.mcp_oauth_tokens.user_id IS 'Dashboard user this token represents (references dashboard.users)';
