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

// Access token (expires in 1 hour)
const accessToken = data.session.access_token;

// Refresh token (expires in 30 days)
const refreshToken = data.session.refresh_token;
```

**Token Configuration:**

```yaml
# fluxbase.yaml
auth:
  jwt:
    secret: "${JWT_SECRET}" # Use strong random secret (256+ bits)
    expiry: "1h" # Access token lifetime
    refresh_expiry: "720h" # Refresh token lifetime (30 days)
```

**Token Revocation:**

```typescript
// Revoke current session
await client.auth.signOut();

// Revoke all sessions for a user
await client.admin.revokeUserSessions(userId);
```

## Encryption Key Rotation Design

Fluxbase supports encryption key rotation for secrets stored in the database. This ensures that if a key is compromised, you can re-encrypt all secrets with a new key.

### Current Implementation

Fluxbase uses a single encryption key for secret storage (configured via `SECRET_KEY`). Key rotation is currently **not automated** but can be performed manually.

### Key Rotation Procedure

#### Phase 1: Add New Key Version

1. **Generate new encryption key:**

```bash
# Generate 256-bit key (32 bytes hex-encoded)
openssl rand -hex 32
# Output: a1b2c3d4e5f6... (64 hex characters)
```

2. **Update configuration with key versioning:**

```yaml
# fluxbase.yaml
security:
  encryption:
    active_key_version: 2 # Increment version number
    keys:
      - version: 1
        key: "${OLD_SECRET_KEY}" # Deprecated key for decryption
      - version: 2
        key: "${NEW_SECRET_KEY}" # New active key for encryption
```

3. **Restart Fluxbase** to load new configuration

#### Phase 2: Re-encrypt Secrets

Run the re-encryption process:

```bash
# CLI command to re-encrypt secrets
fluxbase-cli secrets re-encrypt \
  --from-version 1 \
  --to-version 2 \
  --batch-size 100 \
  --concurrency 4
```

This process:

- Reads secrets encrypted with old key (v1)
- Decrypts with old key
- Encrypts with new key (v2)
- Updates database with new encrypted value
- Tracks progress in case of interruption

**Example Output:**

```
Re-encrypting secrets from version 1 to 2...
✓ Scanned 1000 secrets
✓ Re-encrypted 850 secrets (already using new key: 150)
✓ Updated 850 secrets in database
✓ Completed in 45s
```

#### Phase 3: Verify and Remove Old Key

1. **Verify re-encryption completed:**

```bash
# Check for any remaining secrets encrypted with old key
fluxbase-cli secrets audit --key-version 1
```

Expected output: `No secrets found encrypted with key version 1`

2. **Remove old key from configuration:**

```yaml
# fluxbase.yaml
security:
  encryption:
    active_key_version: 2
    keys:
      # - version: 1  # Remove old key
      #   key: "${OLD_SECRET_KEY}"
      - version: 2
        key: "${NEW_SECRET_KEY}"
```

3. **Restart Fluxbase**

4. **Destroy old key securely:**

```bash
# Securely delete old key from secrets manager
# Example: HashiCorp Vault
vault kv delete secret/fluxbase/keys/v1
```

### Key Rotation Automation (Future)

Planned features for automated key rotation:

```yaml
# fluxbase.yaml (future)
security:
  encryption:
    key_rotation:
      enabled: true
      schedule: "0 2 1 * *" # Cron: 2AM on 1st of each month
      auto_reencrypt: true
      retain_old_keys: 2 # Keep last 2 key versions
      notification_webhook: "https://hooks.example.com/security"
```

**Behavior:**

1. On schedule, generate new encryption key
2. Mark new key as `active_key_version`
3. Start background re-encryption job
4. Send notifications on progress/completion
5. Keep previous N keys for rollback
6. Remove keys older than retention period

### Rollback Procedure

If re-encryption fails or new key is compromised:

```bash
# Emergency rollback to old key version
fluxbase-cli secrets rollback \
  --to-version 1 \
  --reason "New key compromised"

# Update configuration to use old key
# fluxbase.yaml
security:
  encryption:
    active_key_version: 1
```

### Key Storage Options

Store encryption keys securely:

#### Environment Variables (Development)

```bash
export FLUXBASE_SECRET_KEY_1="old-key-here"
export FLUXBASE_SECRET_KEY_2="new-key-here"
```

#### HashiCorp Vault (Production)

```yaml
# fluxbase.yaml
security:
  encryption:
    active_key_version: 2
    keys:
      - version: 1
        key: "{{ vault `secret/fluxbase/keys/v1` }}"
      - version: 2
        key: "{{ vault `secret/fluxbase/keys/v2` }}"
```

#### AWS KMS (Cloud-Native)

```yaml
# fluxbase.yaml
security:
  encryption:
    provider: "aws_kms"
    kms_key_id: "alias/fluxbase-secrets"
    active_key_version: 2
```

### Best Practices

- **Rotate keys quarterly** (or immediately if compromised)
- **Use key versioning** to support multiple active keys
- **Test re-encryption** in staging before production
- **Monitor re-encryption jobs** for failures
- **Keep backup of old keys** until verification complete
- **Document rotation procedures** in runbooks
- **Automate rotation** to reduce human error

## Rate Limiting

Protect against DoS attacks with rate limiting:

```yaml
# fluxbase.yaml
ratelimit:
  backend: "redis" # Use Redis for multi-instance deployments
  global_rate_limit: true # Enable global rate limiting
  limits:
    anonymous: 100 # 100 requests per minute
    authenticated: 1000 # 1000 requests per minute
    service_role: 10000 # Separate limit for service role
  window: "1m"
```

**Best Practices:**

- Use Redis backend for production
- Enable global rate limiting
- Set separate limits for service role tokens
- Monitor rate limit violations

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

Configure security headers:

```yaml
# fluxbase.yaml
server:
  security_headers:
    X-Content-Type-Options: "nosniff"
    X-Frame-Options: "DENY"
    X-XSS-Protection: "1; mode=block"
    Strict-Transport-Security: "max-age=31536000; includeSubDomains"
    Content-Security-Policy: "default-src 'self'"
```

## Environment Separation

Use different configurations for environments:

```bash
# Development
FLUXBASE_ENV=development
FLUXBASE_LOG_LEVEL=debug
FLUXBASE_AUTH_JWT_SECRET=dev-secret-key

# Staging
FLUXBASE_ENV=staging
FLUXBASE_LOG_LEVEL=info
FLUXBASE_AUTH_JWT_SECRET=${STAGING_JWT_SECRET}

# Production
FLUXBASE_ENV=production
FLUXBASE_LOG_LEVEL=warn
FLUXBASE_AUTH_JWT_SECRET=${PRODUCTION_JWT_SECRET}
```

## Security Checklist

- [ ] Enable RLS on all user data tables
- [ ] Use strong JWT secrets (256+ bits)
- [ ] Enable rate limiting with Redis backend
- [ ] Configure security headers
- [ ] Use environment variables for secrets
- [ ] Enable HTTPS in production
- [ ] Implement key rotation procedures
- [ ] Regular security audits
- [ ] Monitor authentication logs
- [ ] Keep dependencies updated

## Related Documentation

- [Authentication](/guides/authentication) - User authentication
- [Row-Level Security](/guides/row-level-security) - Data isolation
- [Rate Limiting](/guides/rate-limiting) - DoS protection
- [Configuration](/reference/configuration) - All security options
