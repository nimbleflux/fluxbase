---
title: "OAuth Authentication"
description: "Zero-config MCP authentication using OAuth 2.1 with PKCE"
---

Fluxbase supports OAuth 2.1 for MCP authentication, enabling zero-config integration with AI assistants like Claude Desktop, Cursor, and VS Code.

## Overview

When OAuth is enabled, MCP clients can authenticate using a browser-based flow:

1. User adds your Fluxbase MCP server URL to their AI assistant
2. The client discovers authentication endpoints automatically
3. User logs in and approves the requested permissions
4. The client receives tokens for authenticated MCP requests

This eliminates the need to manually copy API keys between systems.

## Quick Setup

OAuth is enabled by default in Fluxbase. To use it:

1. **Ensure MCP is enabled** in your `fluxbase.yaml`:

```yaml
mcp:
  enabled: true
  oauth:
    enabled: true
    dcr_enabled: true # Dynamic Client Registration
```

2. **Connect from Claude Desktop** using just your server URL:
   - Open Claude Desktop settings
   - Add a new MCP server with URL: `http://your-server:8080/mcp`
   - Claude will automatically discover OAuth and prompt you to log in

## How It Works

### 1. Discovery

MCP clients discover your authentication endpoints via:

```
GET /.well-known/oauth-authorization-server
```

Response:

```json
{
  "issuer": "https://your-fluxbase.com",
  "authorization_endpoint": "https://your-fluxbase.com/mcp/oauth/authorize",
  "token_endpoint": "https://your-fluxbase.com/mcp/oauth/token",
  "registration_endpoint": "https://your-fluxbase.com/mcp/oauth/register",
  "scopes_supported": ["tables:read", "tables:write", "functions:execute", ...],
  "code_challenge_methods_supported": ["S256"]
}
```

### 2. Dynamic Client Registration (DCR)

Clients can self-register without pre-configured credentials:

```bash
curl -X POST https://your-fluxbase.com/mcp/oauth/register \
  -H "Content-Type: application/json" \
  -d '{
    "client_name": "Claude Desktop",
    "redirect_uris": ["https://claude.ai/api/mcp/auth_callback"]
  }'
```

Response:

```json
{
  "client_id": "mcp_abc123...",
  "client_name": "Claude Desktop",
  "redirect_uris": ["https://claude.ai/api/mcp/auth_callback"],
  "client_id_issued_at": 1234567890
}
```

### 3. Authorization Flow

The standard OAuth 2.1 Authorization Code flow with PKCE:

1. Client generates `code_verifier` and `code_challenge`
2. Client redirects user to `/mcp/oauth/authorize`
3. User logs in and approves permissions
4. Fluxbase redirects back with authorization code
5. Client exchanges code for tokens at `/mcp/oauth/token`

### 4. Token Usage

After authentication, the client includes the access token in MCP requests:

```bash
curl -X POST https://your-fluxbase.com/mcp \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <access_token>" \
  -d '{"jsonrpc":"2.0","method":"tools/list","id":1}'
```

## Configuration

### Basic Configuration

```yaml
mcp:
  enabled: true
  base_path: /mcp
  oauth:
    enabled: true
    dcr_enabled: true
    token_expiry: 1h
    refresh_token_expiry: 168h # 7 days
```

### Allowed Redirect URIs

By default, Fluxbase allows redirect URIs for popular MCP clients:

```yaml
mcp:
  oauth:
    allowed_redirect_uris:
      # Claude Desktop / Claude Code
      - "https://claude.ai/api/mcp/auth_callback"
      - "https://claude.com/api/mcp/auth_callback"
      # Cursor
      - "cursor://anysphere.cursor-mcp/oauth/*/callback"
      # VS Code
      - "http://127.0.0.1:33418"
      - "https://vscode.dev/redirect"
      # OpenCode
      - "http://127.0.0.1:19876/mcp/oauth/callback"
      # MCP Inspector (development)
      - "http://localhost:6274/oauth/callback"
      # ChatGPT
      - "https://chatgpt.com/connector_platform_oauth_redirect"
      # Localhost wildcards (development)
      - "http://localhost:*"
      - "http://127.0.0.1:*"
```

### Environment Variables

For Docker deployments:

```bash
FLUXBASE_MCP_ENABLED=true
FLUXBASE_MCP_OAUTH_ENABLED=true
FLUXBASE_MCP_OAUTH_DCR_ENABLED=true
FLUXBASE_MCP_OAUTH_TOKEN_EXPIRY=1h
FLUXBASE_MCP_OAUTH_REFRESH_TOKEN_EXPIRY=168h
```

## Supported MCP Clients

| Client         | OAuth Support | Callback URI                                            |
| -------------- | ------------- | ------------------------------------------------------- |
| Claude Desktop | Full          | `https://claude.ai/api/mcp/auth_callback`               |
| Claude Code    | Full          | `https://claude.ai/api/mcp/auth_callback`               |
| Cursor         | Full          | `cursor://anysphere.cursor-mcp/oauth/*/callback`        |
| VS Code        | Full          | `http://127.0.0.1:33418`                                |
| OpenCode       | Full          | `http://127.0.0.1:19876/mcp/oauth/callback`             |
| MCP Inspector  | Full          | `http://localhost:6274/oauth/callback`                  |
| ChatGPT        | Full          | `https://chatgpt.com/connector_platform_oauth_redirect` |

## Security

### PKCE Required

All OAuth flows require PKCE (Proof Key for Code Exchange) with S256 method. This prevents authorization code interception attacks.

### Token Rotation

Refresh tokens are rotated on each use. When a refresh token is used:

1. The old token is revoked
2. A new access token and refresh token are issued

This limits the window of exposure if a token is compromised.

### Scopes

OAuth tokens are issued with specific MCP scopes. Users approve these scopes during authorization:

| Scope               | Permission                     |
| ------------------- | ------------------------------ |
| `read:tables`       | Query database tables          |
| `write:tables`      | Insert, update, delete records |
| `execute:functions` | Invoke edge functions          |
| `execute:rpc`       | Execute RPC procedures         |
| `read:storage`      | List and download files        |
| `write:storage`     | Upload and delete files        |
| `execute:jobs`      | Submit and monitor jobs        |
| `read:vectors`      | Vector similarity search       |
| `read:schema`       | Access database schema         |

### Revoking Access

Users can revoke OAuth tokens:

```bash
curl -X POST https://your-fluxbase.com/mcp/oauth/revoke \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "token=<access_or_refresh_token>"
```

## OAuth vs API Keys

| Feature        | OAuth                     | API Keys        |
| -------------- | ------------------------- | --------------- |
| Setup          | Zero-config (automatic)   | Manual key copy |
| User consent   | Browser-based approval    | None            |
| Token rotation | Automatic                 | Manual          |
| Expiration     | Configurable (default 1h) | Long-lived      |
| Best for       | Interactive clients       | CI/CD, scripts  |

**Recommendation:**

- Use **OAuth** for interactive MCP clients (Claude Desktop, Cursor, VS Code)
- Use **API Keys** (X-Service-Key, X-Client-Key) for automation and scripts

## Troubleshooting

### "registration_not_supported" Error

Dynamic Client Registration is disabled. Enable it:

```yaml
mcp:
  oauth:
    dcr_enabled: true
```

### "invalid_redirect_uri" Error

The client's redirect URI is not in the allowed list. Add it to your configuration:

```yaml
mcp:
  oauth:
    allowed_redirect_uris:
      - "https://your-client-callback-url"
```

### "invalid_grant" Error

Common causes:

- Authorization code expired (10 minute limit)
- Authorization code already used
- Invalid PKCE code_verifier
- Client ID mismatch

### User Not Redirected to Login

Ensure your Fluxbase instance has a valid `public_base_url` configured so OAuth redirects work correctly.

## API Reference

### Discovery Endpoint

```
GET /.well-known/oauth-authorization-server
```

Returns OAuth 2.0 Authorization Server Metadata (RFC 8414).

### Dynamic Client Registration

```
POST /mcp/oauth/register
Content-Type: application/json

{
  "client_name": "My MCP Client",
  "redirect_uris": ["https://my-app.com/callback"],
  "scope": "read:tables write:tables"
}
```

### Authorization Endpoint

```
GET /mcp/oauth/authorize?
  response_type=code&
  client_id=mcp_xxx&
  redirect_uri=https://my-app.com/callback&
  scope=read:tables%20write:tables&
  state=random_state&
  code_challenge=xxx&
  code_challenge_method=S256
```

### Token Endpoint

```
POST /mcp/oauth/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&
code=xxx&
redirect_uri=https://my-app.com/callback&
client_id=mcp_xxx&
code_verifier=xxx
```

### Token Revocation

```
POST /mcp/oauth/revoke
Content-Type: application/x-www-form-urlencoded

token=xxx&
token_type_hint=refresh_token
```

## Next Steps

- [MCP Overview](/guides/mcp/) - MCP server setup and configuration
- [MCP Tools](/guides/mcp/tools/) - Available MCP tools
- [MCP Security](/security/mcp-security/) - Security best practices
