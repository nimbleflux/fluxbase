-- MCP OAuth 2.1 support for AI assistant authentication (e.g., Claude Desktop)
-- Implements Dynamic Client Registration (DCR) and standard OAuth 2.1 flows

-- Table for dynamically registered OAuth clients
CREATE TABLE IF NOT EXISTS auth.mcp_oauth_clients (
    -- Client identifier (generated, e.g., "mcp_abc123...")
    client_id TEXT PRIMARY KEY,

    -- Human-readable client name (e.g., "Claude Desktop")
    client_name TEXT NOT NULL,

    -- Client type: "public" (no secret) or "confidential" (has secret)
    client_type TEXT NOT NULL DEFAULT 'public' CHECK (client_type IN ('public', 'confidential')),

    -- Hashed client secret (only for confidential clients)
    client_secret_hash TEXT,

    -- Allowed redirect URIs for this client
    redirect_uris TEXT[] NOT NULL,

    -- Granted scopes for this client
    scopes TEXT[] NOT NULL DEFAULT ARRAY['read:tables', 'read:schema'],

    -- Optional user who registered this client (NULL for pre-registered clients)
    registered_by UUID REFERENCES auth.users(id) ON DELETE SET NULL,

    -- Metadata from registration request
    metadata JSONB DEFAULT '{}',

    -- Whether the client is active
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table for OAuth authorization codes (short-lived, used during auth flow)
CREATE TABLE IF NOT EXISTS auth.mcp_oauth_codes (
    -- Authorization code
    code TEXT PRIMARY KEY,

    -- Client this code was issued to
    client_id TEXT NOT NULL REFERENCES auth.mcp_oauth_clients(client_id) ON DELETE CASCADE,

    -- User who authorized (NULL for DCR flows)
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,

    -- Redirect URI used in authorization request
    redirect_uri TEXT NOT NULL,

    -- Scopes authorized
    scopes TEXT[] NOT NULL,

    -- PKCE code challenge
    code_challenge TEXT,

    -- PKCE code challenge method (S256 required)
    code_challenge_method TEXT,

    -- State for CSRF protection
    state TEXT,

    -- Expiry (short-lived, typically 10 minutes)
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '10 minutes',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Table for OAuth access and refresh tokens
CREATE TABLE IF NOT EXISTS auth.mcp_oauth_tokens (
    -- Token ID (internal, not exposed)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Token type: "access" or "refresh"
    token_type TEXT NOT NULL CHECK (token_type IN ('access', 'refresh')),

    -- Hashed token value (actual token is returned to client once)
    token_hash TEXT NOT NULL UNIQUE,

    -- Client this token was issued to
    client_id TEXT NOT NULL REFERENCES auth.mcp_oauth_clients(client_id) ON DELETE CASCADE,

    -- User this token represents (NULL for client credentials flow)
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,

    -- Scopes granted
    scopes TEXT[] NOT NULL,

    -- For refresh tokens: link to parent refresh token (for rotation)
    parent_token_id UUID REFERENCES auth.mcp_oauth_tokens(id) ON DELETE SET NULL,

    -- Token expiry
    expires_at TIMESTAMPTZ NOT NULL,

    -- Whether the token has been revoked
    is_revoked BOOLEAN NOT NULL DEFAULT false,

    -- Revocation reason (if revoked)
    revoked_reason TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

-- Indexes for efficient lookups
CREATE INDEX IF NOT EXISTS idx_mcp_oauth_clients_registered_by
    ON auth.mcp_oauth_clients(registered_by) WHERE registered_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_codes_client_id
    ON auth.mcp_oauth_codes(client_id);

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_codes_expires_at
    ON auth.mcp_oauth_codes(expires_at);

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_tokens_client_id
    ON auth.mcp_oauth_tokens(client_id);

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_tokens_user_id
    ON auth.mcp_oauth_tokens(user_id) WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_tokens_expires_at
    ON auth.mcp_oauth_tokens(expires_at) WHERE NOT is_revoked;

CREATE INDEX IF NOT EXISTS idx_mcp_oauth_tokens_token_hash
    ON auth.mcp_oauth_tokens(token_hash);

-- Trigger to update updated_at on clients
CREATE OR REPLACE FUNCTION auth.update_mcp_oauth_clients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_mcp_oauth_clients_updated_at ON auth.mcp_oauth_clients;
CREATE TRIGGER trigger_mcp_oauth_clients_updated_at
    BEFORE UPDATE ON auth.mcp_oauth_clients
    FOR EACH ROW EXECUTE FUNCTION auth.update_mcp_oauth_clients_updated_at();

-- Comments
COMMENT ON TABLE auth.mcp_oauth_clients IS 'OAuth 2.1 clients for MCP authentication (Dynamic Client Registration)';
COMMENT ON TABLE auth.mcp_oauth_codes IS 'Short-lived authorization codes for OAuth 2.1 flows';
COMMENT ON TABLE auth.mcp_oauth_tokens IS 'OAuth 2.1 access and refresh tokens for MCP clients';
