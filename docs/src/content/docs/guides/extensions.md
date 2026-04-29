---
title: "PostgreSQL Extensions"
description: Manage PostgreSQL extensions in Fluxbase. List, enable, and disable extensions through the admin API with automatic dependency resolution.
---

Fluxbase provides a management layer for PostgreSQL extensions, letting you list, install, and remove extensions through the admin API instead of running raw SQL.

## Overview

Extensions are managed via the `platform.available_extensions` catalog and `platform.enabled_extensions` tracking table. When you enable an extension, Fluxbase:

- Validates the extension exists in the catalog
- Resolves dependencies automatically (e.g., enabling `pgvector` also enables `cube` if needed)
- Executes `CREATE EXTENSION IF NOT EXISTS` with the appropriate permissions
- Records the operation in the tracking table

Core extensions (required by Fluxbase) are enabled automatically on startup and cannot be disabled.

## API Endpoints

All extension endpoints require authentication with `admin`, `instance_admin`, or `tenant_admin` role.

| Method  | Endpoint                                | Description            |
| ------- | --------------------------------------- | ---------------------- |
| `GET`   | `/api/v1/admin/extensions`              | List all extensions    |
| `GET`   | `/api/v1/admin/extensions/:name`        | Get extension status   |
| `POST`  | `/api/v1/admin/extensions/:name/enable` | Enable an extension    |
| `POST`  | `/api/v1/admin/extensions/:name/disable`| Disable an extension   |
| `POST`  | `/api/v1/admin/extensions/sync`         | Sync from PostgreSQL   |

## Usage

### List Extensions

```bash
curl -H "Authorization: Bearer <service-role-key>" \
  http://localhost:8080/api/v1/admin/extensions
```

Response includes all available extensions with their status, category, and version info:

```json
{
  "extensions": [
    {
      "name": "vector",
      "display_name": "pgvector",
      "description": "Vector data type and similarity search",
      "category": "ai_ml",
      "is_core": false,
      "is_enabled": true,
      "is_installed": true,
      "installed_version": "0.7.0"
    }
  ],
  "categories": [
    { "id": "ai_ml", "name": "AI & Machine Learning", "count": 1 }
  ]
}
```

### Get Extension Status

```bash
curl -H "Authorization: Bearer <service-role-key>" \
  http://localhost:8080/api/v1/admin/extensions/vector
```

### Enable an Extension

```bash
curl -X POST \
  -H "Authorization: Bearer <service-role-key>" \
  -H "Content-Type: application/json" \
  -d '{"schema": "public"}' \
  http://localhost:8080/api/v1/admin/extensions/vector/enable
```

The `schema` field is optional (defaults to `public`).

### Disable an Extension

```bash
curl -X POST \
  -H "Authorization: Bearer <service-role-key>" \
  http://localhost:8080/api/v1/admin/extensions/vector/disable
```

Core extensions cannot be disabled. Disabling uses `DROP EXTENSION ... CASCADE`.

### Sync Extensions

Refresh the extension catalog from PostgreSQL:

```bash
curl -X POST \
  -H "Authorization: Bearer <service-role-key>" \
  http://localhost:8080/api/v1/admin/extensions/sync
```

## Tenant Scoping

Extensions are scoped per-database. For tenants with a separate database, extensions are managed independently. Tenant admins can enable/disable extensions for their own tenant database, while instance admins can manage extensions across all databases.

## Categories

| Category ID     | Display Name          |
| --------------- | --------------------- |
| `core`          | Core                  |
| `ai_ml`         | AI & Machine Learning |
| `geospatial`    | Geospatial            |
| `data_types`    | Data Types            |
| `indexing`      | Indexing              |
| `text_search`   | Text Search           |
| `performance`   | Performance           |
| `monitoring`    | Monitoring            |
| `utilities`     | Utilities             |
| `foreign_data`  | Foreign Data          |
| `triggers`      | Triggers              |

## Learn More

- [Vector Search](/guides/vector-search/) - Uses the `vector` extension
- [Multi-Tenancy](/guides/multi-tenancy/) - Tenant database isolation
