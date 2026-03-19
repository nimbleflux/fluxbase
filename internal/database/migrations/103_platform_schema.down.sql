-- Rollback: Rename platform schema back to dashboard

-- Rename the schema back
ALTER SCHEMA platform RENAME TO dashboard;

-- Revert comments on auth.* tables
COMMENT ON COLUMN auth.mcp_oauth_codes.user_id IS 'Dashboard user who authorized this code (references dashboard.users)';
COMMENT ON COLUMN auth.mcp_oauth_tokens.user_id IS 'Dashboard user this token represents (references dashboard.users)';
