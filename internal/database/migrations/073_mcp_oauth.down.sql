-- Revert MCP OAuth 2.1 Support

DROP TRIGGER IF EXISTS trigger_mcp_oauth_clients_updated_at ON auth.mcp_oauth_clients;
DROP FUNCTION IF EXISTS auth.update_mcp_oauth_clients_updated_at();
DROP FUNCTION IF EXISTS auth.update_mcp_oauth_client_updated_at();

DROP TABLE IF EXISTS auth.mcp_oauth_tokens;
DROP TABLE IF EXISTS auth.mcp_oauth_codes;
DROP TABLE IF EXISTS auth.mcp_oauth_clients;
