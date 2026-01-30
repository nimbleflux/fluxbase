-- Revert MCP OAuth tables to reference auth.users

-- Drop dashboard.users foreign key constraints
ALTER TABLE auth.mcp_oauth_codes
    DROP CONSTRAINT IF EXISTS mcp_oauth_codes_user_id_fkey;

ALTER TABLE auth.mcp_oauth_tokens
    DROP CONSTRAINT IF EXISTS mcp_oauth_tokens_user_id_fkey;

-- Restore original foreign keys to auth.users
ALTER TABLE auth.mcp_oauth_codes
    ADD CONSTRAINT mcp_oauth_codes_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;

ALTER TABLE auth.mcp_oauth_tokens
    ADD CONSTRAINT mcp_oauth_tokens_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
