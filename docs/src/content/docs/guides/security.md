---
title: "Security Best Practices"
description: Learn about security best practices for Fluxbase including key rotation, secret management, and securing your deployment.
---

Fluxbase provides multiple security features to protect your data and applications. This guide covers security best practices and configuration.

## Overview

Fluxbase security includes:

- **Row-Level Security (RLS)** for data isolation
- **JWT-based authentication** with token revocation
- **OAuth2, OIDC, and SAML SSO** integration
- **API key and client key management**
- **Secret encryption** at rest
- **Rate limiting** for DoS protection
- **SQL injection prevention** via parameterized queries

## Authentication & Authorization

### Row-Level Security (RLS)

Fluxbase uses PostgreSQL Row-Level Security for tenant data isolation:

```sql
-- Enable RLS on a table
ALTER TABLE user_profiles ENABLE ROW LEVEL SECURITY;

-- Create policy: users can only see their own profile
CREATE POLICY user_profiles_isolation ON user_profiles
  FOR ALL
  USING (auth.uid() = user_id);
```

**Best Practices:**

- Always enable RLS on user data tables
- Use `auth.uid()` in policies to reference the authenticated user
- Test policies with different user roles
- Use `SELECT * FROM pg_policies` to review active policies

### JWT Token Management

Fluxbase uses JWT tokens for authentication:

```typescript
import { createClient } from "@nimbleflux/fluxbase-sdk";

const client = createClient("http://localhost:8080", "your-anon-key");

// Sign in
const { data, error } = await client.auth.signIn({
  email: "user@example.com",
  password: "password",
});

// Access token (expires in 15 minutes)
const accessToken = data.session.access_token;

// Refresh token (expires in 7 days)
const refreshToken = data.session.refresh_token;
```

**Token Configuration:**

```yaml
# fluxbase.yaml
auth:
  jwt:
    secret: "${JWT_SECRET}" # Use strong random secret (256+ bits)
    expiry: "15m" # Access token lifetime
    refresh_expiry: "168h" # Refresh token lifetime (7 days)
```

**Token Revocation:**

```typescript
// Revoke current session
await client.auth.signOut();

// Revoke all sessions for a user
await client.admin.revokeUserSessions(userId);
```

## Encryption

Fluxbase encrypts secrets at rest using AES-256-GCM with a single encryption key:

```yaml
# fluxbase.yaml
security:
  encryption_key: "${ENCRYPTION_KEY}"
```

The key is a 32-byte hex-encoded string. Set it via the `FLUXBASE_ENCRYPTION_KEY` environment variable.

**Key rotation** is currently not automated. To rotate the key, generate a new one and update the environment variable, then restart Fluxbase.

### Best Practices

- **Store the encryption key** in a secrets manager (e.g., HashiCorp Vault, AWS Secrets Manager)
- **Never commit** the encryption key to source control
- **Rotate keys** if compromise is suspected
- **Back up the key** — losing it means encrypted data (OAuth tokens) cannot be decrypted

## Rate Limiting

Protect against DoS attacks with rate limiting:

```yaml
# fluxbase.yaml
security:
  enable_global_rate_limit: true
```

This enables **100 requests per minute per IP** across all API endpoints. Rate limiting is in-memory and per-instance — see the [Rate Limiting guide](/guides/rate-limiting/) for multi-instance considerations.

**Best Practices:**

- Enable global rate limiting in production
- Set up a reverse proxy or API gateway for centralized rate limiting in multi-instance deployments
- Monitor rate limit violations via Prometheus metrics

## SQL Injection Prevention

Fluxbase prevents SQL injection via:

1. **Parameterized queries** (pgx `$1, $2` placeholders)
2. **Identifier quoting** (`quoteIdentifier()`)
3. **Query builders** with safe defaults

**Example Safe Query:**

```go
// SAFE: Parameterized query
query := "SELECT * FROM users WHERE email = $1"
err := db.QueryRow(ctx, query, userEmail).Scan(&user)

// SAFE: Quoted identifier
table := quoteIdentifier(tableName)
query := fmt.Sprintf("SELECT * FROM %s", table)
```

**Best Practices:**

- Always use parameterized queries
- Never concatenate user input into SQL
- Use `quoteIdentifier()` for dynamic table/column names
- Validate user input before database operations

## Security Headers

Fluxbase sets security headers automatically on all API responses, including `X-Content-Type-Options`, `X-Frame-Options`, `Strict-Transport-Security`, and `Content-Security-Policy`. These are hardcoded and do not require configuration.

## Environment Separation

Use different configurations for environments:

```bash
# Development
FLUXBASE_LOGGING_CONSOLE_LEVEL=debug
FLUXBASE_AUTH_JWT_SECRET=dev-secret-key

# Staging
FLUXBASE_LOGGING_CONSOLE_LEVEL=info
FLUXBASE_AUTH_JWT_SECRET=${STAGING_JWT_SECRET}

# Production
FLUXBASE_LOGGING_CONSOLE_LEVEL=warn
FLUXBASE_AUTH_JWT_SECRET=${PRODUCTION_JWT_SECRET}
```

## Security Checklist

- [ ] Enable RLS on all user data tables
- [ ] Use strong JWT secrets (256+ bits)
- [ ] Enable global rate limiting
- [ ] Use environment variables for secrets
- [ ] Enable HTTPS in production
- [ ] Keep dependencies updated

## Related Documentation

- [Authentication](/guides/authentication) - User authentication
- [Row-Level Security](/guides/row-level-security) - Data isolation
- [Rate Limiting](/guides/rate-limiting) - DoS protection
- [Configuration](/reference/configuration) - All security options
