---
title: "Secrets Management"
description: "Securely manage secrets for edge functions and background jobs"
---

Fluxbase provides a secure secrets management system for storing API keys, database credentials, and other sensitive values that your edge functions and background jobs need at runtime.

## Overview

Secrets are:

- **Encrypted at rest** using AES-256-GCM
- **Scoped** to global or namespace level
- **Version controlled** with rollback support
- **Injected automatically** into function runtime as environment variables

## CLI Commands

### List Secrets

View all secrets (values are never displayed):

```bash
# List all secrets
fluxbase secrets list

# Filter by scope
fluxbase secrets list --scope global
fluxbase secrets list --scope namespace

# Filter by namespace
fluxbase secrets list --namespace production
```

### Create or Update a Secret

```bash
# Create a global secret
fluxbase secrets set API_KEY "sk-your-api-key"

# Create a namespace-scoped secret
fluxbase secrets set DATABASE_URL "postgres://..." --scope namespace --namespace production

# Add a description
fluxbase secrets set STRIPE_KEY "sk_live_..." --description "Stripe production API key"

# Set an expiration
fluxbase secrets set TEMP_TOKEN "xyz" --expires 30d
```

### Get Secret Metadata

```bash
# Get metadata (value is never returned)
fluxbase secrets get API_KEY

# For namespaced secrets
fluxbase secrets get DATABASE_URL --namespace production
```

### Delete a Secret

```bash
# Delete a secret
fluxbase secrets delete API_KEY

# Delete a namespaced secret
fluxbase secrets delete DATABASE_URL --namespace production
```

### Version History

Secrets maintain version history for audit and rollback:

```bash
# View version history
fluxbase secrets history API_KEY

# Rollback to a previous version
fluxbase secrets rollback API_KEY 2
```

## Scope Levels

Secrets support two scope levels:

| Scope       | Description                                         | Use Case                                       |
| ----------- | --------------------------------------------------- | ---------------------------------------------- |
| `global`    | Available to all functions in all namespaces        | Shared API keys, common credentials            |
| `namespace` | Available only to functions in a specific namespace | Environment-specific secrets (prod vs staging) |

### Resolution Order

When a function requests a secret, Fluxbase resolves it in this order:

1. **Namespace-scoped secret** (if function is in a namespace)
2. **Global secret** (fallback)

This allows you to override global defaults with namespace-specific values.

## Using Secrets in Edge Functions

Secrets are injected as environment variables with the `FLUXBASE_SECRET_` prefix:

```typescript
// functions/send-email.ts
export default async function handler(req: Request): Promise<Response> {
  // Access secrets via environment variables
  const apiKey = Deno.env.get("FLUXBASE_SECRET_SENDGRID_API_KEY");

  if (!apiKey) {
    return new Response("Missing API key", { status: 500 });
  }

  const response = await fetch("https://api.sendgrid.com/v3/mail/send", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${apiKey}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      // email data
    }),
  });

  return new Response("Email sent");
}
```

### Using the Secrets Object (Recommended)

For cleaner code, use the built-in `secrets` object:

```typescript
// functions/process-payment.ts
export default async function handler(req: Request): Promise<Response> {
  // Use the secrets helper
  const stripeKey = secrets.get("STRIPE_SECRET_KEY");

  // Or require the secret (throws if missing)
  const requiredKey = secrets.getRequired("STRIPE_SECRET_KEY");

  // Your logic here
}
```

## Using Secrets in Background Jobs

Background jobs have the same access to secrets:

```typescript
// jobs/sync-data.ts
export default async function handler(payload: unknown): Promise<void> {
  const apiKey = Deno.env.get("FLUXBASE_SECRET_EXTERNAL_API_KEY");
  const dbUrl = Deno.env.get("FLUXBASE_SECRET_ANALYTICS_DB");

  // Sync data to external service
  await syncToExternalService(apiKey, payload);
}
```

## Settings Secrets vs Legacy Secrets

Fluxbase provides two secrets systems:

| Feature         | `fluxbase settings secrets`       | `fluxbase secrets`                  |
| --------------- | --------------------------------- | ----------------------------------- |
| Storage         | `app.settings` table              | `functions.secrets` table           |
| Scopes          | System, User                      | Global, Namespace                   |
| User-specific   | Yes (with HKDF encryption)        | No                                  |
| Version history | No                                | Yes                                 |
| Function access | `secrets.get()`                   | `Deno.env.get("FLUXBASE_SECRET_*")` |
| Best for        | Application config, per-user keys | Function runtime secrets            |

For new projects, consider which system fits your needs:

- **Edge function secrets**: Use `fluxbase secrets` (this guide)
- **Application configuration**: Use `fluxbase settings secrets`
- **Per-user API keys**: Use `fluxbase settings secrets --user`

## Security Best Practices

### 1. Use Namespace Scoping

Separate production and development secrets:

```bash
# Development
fluxbase secrets set STRIPE_KEY "sk_test_..." --scope namespace --namespace development

# Production
fluxbase secrets set STRIPE_KEY "sk_live_..." --scope namespace --namespace production
```

### 2. Set Expiration for Temporary Secrets

```bash
# Expires in 30 days
fluxbase secrets set TEMP_TOKEN "xyz" --expires 30d

# Expires in 1 year
fluxbase secrets set ANNUAL_KEY "abc" --expires 1y
```

### 3. Use Descriptive Names

```bash
# Good
fluxbase secrets set SENDGRID_API_KEY "..." --description "SendGrid transactional email API key"

# Avoid
fluxbase secrets set KEY1 "..."
```

### 4. Rotate Secrets Regularly

Use version history to track changes:

```bash
# Update the secret (creates new version)
fluxbase secrets set API_KEY "new-value"

# View history
fluxbase secrets history API_KEY

# Rollback if needed
fluxbase secrets rollback API_KEY 1
```

### 5. Audit Secret Access

Check which functions use which secrets:

```bash
# Review function logs for secret access
fluxbase logs list --category execution --search "FLUXBASE_SECRET"
```

## Environment Variables

You can also provide secrets via environment variables when running Fluxbase:

```bash
# These are available to all functions
export FLUXBASE_SECRET_API_KEY="your-key"
export FLUXBASE_SECRET_DATABASE_URL="postgres://..."
```

Environment variables take precedence over stored secrets, useful for:

- Local development
- CI/CD pipelines
- Kubernetes secrets injection

## API Reference

### REST Endpoints (Name-Based - Recommended)

Name-based endpoints are the recommended way to work with secrets. They use the secret name instead of UUIDs, making them easier to use.

| Method | Endpoint                                          | Description                                      |
| ------ | ------------------------------------------------- | ------------------------------------------------ |
| GET    | `/api/v1/secrets/by-name/:name`                   | Get secret by name (query: `namespace`)          |
| PUT    | `/api/v1/secrets/by-name/:name`                   | Update secret by name (query: `namespace`)       |
| DELETE | `/api/v1/secrets/by-name/:name`                   | Delete secret by name (query: `namespace`)       |
| GET    | `/api/v1/secrets/by-name/:name/versions`          | Get version history by name (query: `namespace`) |
| POST   | `/api/v1/secrets/by-name/:name/rollback/:version` | Rollback by name (query: `namespace`)            |

### REST Endpoints (UUID-Based - Legacy)

UUID-based endpoints are kept for backward compatibility. For new code, prefer name-based endpoints above.

| Method | Endpoint                                | Description                                    |
| ------ | --------------------------------------- | ---------------------------------------------- |
| GET    | `/api/v1/secrets`                       | List all secrets (query: `scope`, `namespace`) |
| GET    | `/api/v1/secrets/stats`                 | Get secret statistics                          |
| POST   | `/api/v1/secrets`                       | Create a secret                                |
| GET    | `/api/v1/secrets/:id`                   | Get secret metadata by UUID                    |
| PUT    | `/api/v1/secrets/:id`                   | Update a secret by UUID                        |
| DELETE | `/api/v1/secrets/:id`                   | Delete a secret by UUID                        |
| GET    | `/api/v1/secrets/:id/versions`          | Get version history by UUID                    |
| POST   | `/api/v1/secrets/:id/rollback/:version` | Rollback to version by UUID                    |

### Request Body (Create)

```json
{
  "name": "API_KEY",
  "value": "sk-your-secret-key",
  "scope": "global",
  "namespace": "production",
  "description": "External API key",
  "expires_at": "2025-12-31T00:00:00Z"
}
```

Note: `namespace` is only required when `scope` is `"namespace"`. The `:id` in UUID-based endpoints is a UUID returned after creation.

### Request Body (Update)

```json
{
  "value": "new-secret-value",
  "description": "Updated description",
  "expires_at": "2026-12-31T00:00:00Z"
}
```

## SDK Usage

The TypeScript SDK provides a `secrets` manager with both name-based and UUID-based methods:

```typescript
import { createClient } from "@nimbleflux/fluxbase-sdk";

const client = createClient({ url: "http://localhost:8080" });

// Create a secret
const secret = await client.secrets.create({
  name: "API_KEY",
  value: "sk-your-api-key",
  scope: "global",
  description: "External API key",
});

// Get secret by name (recommended)
const secret = await client.secrets.get("API_KEY");

// Get namespace-scoped secret
const secret = await client.secrets.get("DATABASE_URL", {
  namespace: "production",
});

// Update secret by name
await client.secrets.update("API_KEY", { value: "new-api-key" });

// Delete secret by name
await client.secrets.delete("OLD_KEY");

// Get version history
const versions = await client.secrets.getVersions("API_KEY");

// Rollback to version 2
await client.secrets.rollback("API_KEY", 2);

// List all secrets
const secrets = await client.secrets.list();

// List by scope
const globalSecrets = await client.secrets.list({ scope: "global" });

// Get statistics
const stats = await client.secrets.stats();
```

## Next Steps

- [Edge Functions Guide](/guides/edge-functions/) - Deploy serverless functions
- [Background Jobs Guide](/guides/jobs/) - Run scheduled and async jobs
- [Configuration Reference](/reference/configuration/) - All config options
