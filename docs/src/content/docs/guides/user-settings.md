---
title: "User Settings"
description: Per-user key-value settings and encrypted secrets for storing user preferences and sensitive configuration.
---

User settings provide per-user, key-value configuration storage. Fluxbase supports both plaintext settings (for preferences) and encrypted secrets (for sensitive values like API tokens).

## Overview

User settings enable:

- **Per-user configuration** - Each user has their own settings namespace
- **System fallback** - If a user hasn't set a value, the system default is returned
- **Encrypted secrets** - Store sensitive values with AES-256-GCM encryption
- **Tenant awareness** - Settings respect tenant isolation via RLS

Settings are stored in `app.settings` and support both user-scoped and system-scoped entries.

## User Settings

### Set a User Setting

```bash
curl -X PUT \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"value": {"theme": "dark", "language": "en"}}' \
  http://localhost:8080/api/v1/settings/user/preferences
```

This is an upsert operation — creating the setting if it doesn't exist, or updating it if it does.

### Get a Setting (with Fallback)

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/user/preferences
```

Returns the user's own value if set, otherwise falls back to the system-level default:

```json
{
  "key": "preferences",
  "value": {"theme": "dark", "language": "en"},
  "source": "user"
}
```

The `source` field is either `"user"` or `"system"`.

### Get User's Own Setting Only

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/user/own/preferences
```

Returns only the user's own setting — no system fallback. Returns 404 if the user hasn't set it.

### Get a System Setting

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/user/system/preferences
```

Returns the system-level default for a key.

### List All User Settings

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/user/list
```

### Delete a User Setting

```bash
curl -X DELETE \
  -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/user/preferences
```

## User Secrets

Secrets are encrypted at rest and never returned via the API. Only metadata (key, description, timestamps) is exposed. Secret values are encrypted with a user-specific derived key.

### Create a Secret

```bash
curl -X POST \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"key": "third_party_api_key", "value": "sk-abc123", "description": "External API key"}' \
  http://localhost:8080/api/v1/settings/secret
```

Response contains only metadata — the value is never returned:

```json
{
  "id": "uuid",
  "key": "third_party_api_key",
  "description": "External API key",
  "user_id": "uuid",
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

### List Secrets

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/secret
```

### Get Secret Metadata

```bash
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/secret/third_party_api_key
```

### Update a Secret

```bash
curl -X PUT \
  -H "Authorization: Bearer <jwt-token>" \
  -H "Content-Type: application/json" \
  -d '{"value": "sk-newkey456", "description": "Updated API key"}' \
  http://localhost:8080/api/v1/settings/secret/third_party_api_key
```

### Delete a Secret

```bash
curl -X DELETE \
  -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/settings/secret/third_party_api_key
```

## API Endpoints

### User Settings

| Method   | Endpoint                               | Description                  |
| -------- | -------------------------------------- | ---------------------------- |
| `GET`    | `/api/v1/settings/user/list`           | List all user settings       |
| `GET`    | `/api/v1/settings/user/:key`           | Get setting (with fallback)  |
| `GET`    | `/api/v1/settings/user/own/:key`       | Get user's own setting only  |
| `GET`    | `/api/v1/settings/user/system/:key`    | Get system default           |
| `PUT`    | `/api/v1/settings/user/:key`           | Set (upsert) a user setting  |
| `DELETE` | `/api/v1/settings/user/:key`           | Delete a user setting        |

### User Secrets

| Method   | Endpoint                        | Description            |
| -------- | ------------------------------- | ---------------------- |
| `POST`   | `/api/v1/settings/secret`       | Create a secret        |
| `GET`    | `/api/v1/settings/secret`       | List secrets           |
| `GET`    | `/api/v1/settings/secret/*`     | Get secret metadata    |
| `PUT`    | `/api/v1/settings/secret/*`     | Update a secret        |
| `DELETE` | `/api/v1/settings/secret/*`     | Delete a secret        |

All endpoints require authentication.

## App Settings (Public)

The global settings endpoint provides read access to application settings without authentication:

| Method | Endpoint                   | Description              |
| ------ | -------------------------- | ------------------------ |
| `GET`  | `/api/v1/settings/:key`    | Get a setting by key     |
| `GET`  | `/api/v1/settings/`        | Batch get settings       |
| `POST` | `/api/v1/settings/batch`   | Batch get by keys        |

## Learn More

- [Secrets Management](/guides/secrets-management/) - Function and job secrets
- [Edge Functions](/guides/edge-functions/) - Secrets injection into functions
- [Authentication](/guides/authentication/) - User authentication
