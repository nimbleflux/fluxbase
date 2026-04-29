---
title: HTTP API Reference
description: Complete HTTP API documentation for Fluxbase REST endpoints including authentication, storage, database operations, multi-tenancy, functions, jobs, and more.
---
The Fluxbase HTTP API provides RESTful endpoints for authentication, storage, database operations, multi-tenancy management, edge functions, background jobs, and more. All endpoints are prefixed with `/api/v1/`.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require authentication via JWT bearer tokens or service keys. Include the token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  http://localhost:8080/api/v1/auth/user
```

### Multi-Tenant Requests

For multi-tenant deployments, specify the tenant context via the `X-FB-Tenant` header:

```bash
curl -H "Authorization: Bearer <service-key>" \
     -H "X-FB-Tenant: acme-corp" \
     http://localhost:8080/api/v1/tables/posts
```

When using a tenant-scoped service key, the tenant context is embedded in the key and the header is optional.

## API Categories

### Authentication

Endpoints for user registration, login, and session management.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/auth/signup` | Register a new user |
| `POST` | `/auth/signin` | Sign in with email/password |
| `POST` | `/auth/signout` | Sign out current session |
| `POST` | `/auth/refresh` | Refresh access token |
| `GET` | `/auth/user` | Get current user |
| `PATCH` | `/auth/user` | Update current user |
| `POST` | `/auth/magiclink` | Request magic link |
| `GET` | `/auth/magiclink/verify` | Verify magic link token |
| `POST` | `/auth/factors` | Register a 2FA factor |
| `GET` | `/auth/factors` | List 2FA factors |
| `DELETE` | `/auth/factors/{id}` | Remove a 2FA factor |
| `POST` | `/auth/factors/{id}/verify` | Verify and enable 2FA factor |
| `POST` | `/auth/factors/{id}/challenge` | Create a 2FA challenge |
| `POST` | `/auth/factors/{id}/verify-challenge` | Verify a 2FA challenge |
| `POST` | `/auth/password/reset` | Request password reset email |
| `POST` | `/auth/password/verify` | Verify password reset token |
| `POST` | `/auth/token` | Exchange client key for JWT |
| `POST` | `/auth/saml/{provider}/callback` | SAML SSO callback |
| `GET` | `/auth/saml/{provider}/metadata` | SAML SP metadata |
| `GET` | `/auth/oauth/{provider}` | Start OAuth flow |
| `GET` | `/auth/oauth/{provider}/callback` | OAuth callback |

### Storage

Endpoints for file storage operations.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/storage/buckets` | List all buckets |
| `POST` | `/storage/buckets/{bucket}` | Create bucket |
| `DELETE` | `/storage/buckets/{bucket}` | Delete bucket |
| `GET` | `/storage/{bucket}` | List files in bucket |
| `POST` | `/storage/{bucket}/{key}` | Upload file |
| `GET` | `/storage/{bucket}/{key}` | Download file |
| `HEAD` | `/storage/{bucket}/{key}` | Get file metadata |
| `DELETE` | `/storage/{bucket}/{key}` | Delete file |
| `POST` | `/storage/{bucket}/sign/{key}` | Generate signed URL |
| `POST` | `/storage/transform` | Image transformation |

### GraphQL

A full GraphQL API auto-generated from your database schema.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/graphql` | Execute GraphQL queries and mutations |

See the [GraphQL API documentation](/api/http/graphql) for complete details on queries, mutations, filtering, and SDK usage.

### Database Tables

Auto-generated CRUD endpoints for your PostgreSQL tables.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/tables/{table}` | List records with filtering |
| `POST` | `/tables/{table}` | Create record(s) |
| `PATCH` | `/tables/{table}` | Batch update records |
| `DELETE` | `/tables/{table}` | Batch delete records |
| `GET` | `/tables/{table}/{id}` | Get record by ID |
| `PUT` | `/tables/{table}/{id}` | Replace record |
| `PATCH` | `/tables/{table}/{id}` | Update record |
| `DELETE` | `/tables/{table}/{id}` | Delete record |
| `POST` | `/tables/{table}/bulk` | Bulk insert |
| `PATCH` | `/tables/{table}/bulk` | Bulk update |
| `DELETE` | `/tables/{table}/bulk` | Bulk delete |
| `GET` | `/tables/{table}/export` | Export table data (CSV/JSON) |

### Tenant Management

Manage tenants in multi-tenant deployments. Requires `admin`, `instance_admin`, or `tenant_admin` role.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/tenants` | List all tenants |
| `GET` | `/admin/tenants/mine` | List tenants for current user |
| `GET` | `/admin/tenants/deleted` | List soft-deleted tenants |
| `POST` | `/admin/tenants` | Create tenant |
| `GET` | `/admin/tenants/{id}` | Get tenant details |
| `PATCH` | `/admin/tenants/{id}` | Update tenant |
| `DELETE` | `/admin/tenants/{id}` | Soft delete tenant (`?hard=true` for hard delete) |
| `POST` | `/admin/tenants/{id}/recover` | Recover soft-deleted tenant |
| `POST` | `/admin/tenants/{id}/migrate` | Migrate tenant to latest schema |
| `POST` | `/admin/tenants/{id}/repair` | Repair tenant (re-run bootstrap + FDW) |

### Tenant Members

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/tenants/{id}/members` | List tenant members |
| `POST` | `/admin/tenants/{id}/members` | Add member to tenant |
| `DELETE` | `/admin/tenants/{id}/members/{user_id}` | Remove member from tenant |
| `GET` | `/admin/tenants/{id}/admins` | List tenant admins |
| `POST` | `/admin/tenants/{id}/admins` | Assign tenant admin |
| `DELETE` | `/admin/tenants/{id}/admins/{user_id}` | Remove tenant admin |

### Tenant Settings

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/tenants/{id}/settings` | Get tenant settings |
| `PATCH` | `/admin/tenants/{id}/settings` | Update tenant settings |
| `DELETE` | `/admin/tenants/{id}/settings/{key}` | Delete a tenant setting |
| `GET` | `/admin/tenants/{id}/settings/{key}` | Get a specific tenant setting |

### Tenant Declarative Schema

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/tenants/{id}/schema` | Get schema status |
| `POST` | `/admin/tenants/{id}/schema/apply` | Apply schema from filesystem |
| `GET` | `/admin/tenants/{id}/schema/content` | Get stored schema SQL |
| `POST` | `/admin/tenants/{id}/schema/content` | Upload schema SQL |
| `POST` | `/admin/tenants/{id}/schema/content/apply` | Upload and apply schema SQL |
| `DELETE` | `/admin/tenants/{id}/schema/content` | Delete stored schema |

### Service Keys

Manage API service keys. Scoped to the current tenant context via `X-FB-Tenant`.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/service-keys` | List service keys |
| `POST` | `/admin/service-keys` | Create service key |
| `GET` | `/admin/service-keys/{id}` | Get service key details |
| `PUT` | `/admin/service-keys/{id}` | Update service key |
| `DELETE` | `/admin/service-keys/{id}` | Delete service key |
| `POST` | `/admin/service-keys/{id}/disable` | Disable service key |
| `POST` | `/admin/service-keys/{id}/enable` | Enable service key |
| `POST` | `/admin/service-keys/{id}/revoke` | Revoke service key |
| `POST` | `/admin/service-keys/{id}/deprecate` | Deprecate key with grace period |
| `POST` | `/admin/service-keys/{id}/rotate` | Rotate service key |
| `GET` | `/admin/service-keys/{id}/revocations` | Get revocation history |

### Edge Functions

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/functions` | List functions |
| `POST` | `/functions` | Create function |
| `GET` | `/functions/{name}` | Get function details |
| `PUT` | `/functions/{name}` | Update function |
| `DELETE` | `/functions/{name}` | Delete function |
| `POST` | `/functions/{name}` | Invoke function |

### Background Jobs

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/jobs` | List jobs |
| `POST` | `/jobs` | Create job |
| `GET` | `/jobs/{name}` | Get job details |
| `PUT` | `/jobs/{name}` | Update job |
| `DELETE` | `/jobs/{name}` | Delete job |
| `POST` | `/jobs/{name}/run` | Trigger job execution |
| `GET` | `/jobs/{name}/runs` | List job execution history |
| `GET` | `/jobs/{name}/runs/{run_id}` | Get execution details |

### RPC (Remote Procedures)

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/rpc` | List procedures |
| `POST` | `/rpc` | Create procedure |
| `GET` | `/rpc/{name}` | Get procedure details |
| `PUT` | `/rpc/{name}` | Update procedure |
| `DELETE` | `/rpc/{name}` | Delete procedure |
| `POST` | `/rpc/{name}` | Execute procedure |

### Database Branching

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/branches` | List branches |
| `POST` | `/branches` | Create branch |
| `GET` | `/branches/{id}` | Get branch details |
| `PATCH` | `/branches/{id}` | Update branch |
| `DELETE` | `/branches/{id}` | Delete branch |
| `POST` | `/branches/{id}/merge` | Merge branch |
| `POST` | `/branches/{id}/reset` | Reset branch |

### Webhooks

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/webhooks` | List webhooks |
| `POST` | `/webhooks` | Create webhook |
| `GET` | `/webhooks/{id}` | Get webhook details |
| `PUT` | `/webhooks/{id}` | Update webhook |
| `DELETE` | `/webhooks/{id}` | Delete webhook |
| `POST` | `/webhooks/{id}/test` | Test webhook delivery |
| `GET` | `/webhooks/{id}/deliveries` | List webhook deliveries |

### Migrations

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/admin/migrations` | List migrations |
| `POST` | `/admin/migrations` | Create migration |
| `GET` | `/admin/migrations/{name}` | Get migration details |
| `POST` | `/admin/migrations/{name}/apply` | Apply migration |
| `POST` | `/admin/migrations/{name}/rollback` | Rollback migration |
| `POST` | `/admin/migrations/apply-pending` | Apply all pending migrations |
| `POST` | `/admin/migrations/sync` | Sync migrations (batch upload) |

### AI Chatbots & Knowledge Bases

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/chatbots` | List chatbots |
| `POST` | `/chatbots` | Create chatbot |
| `GET` | `/chatbots/{id}` | Get chatbot details |
| `PUT` | `/chatbots/{id}` | Update chatbot |
| `DELETE` | `/chatbots/{id}` | Delete chatbot |
| `POST` | `/chatbots/{id}/chat` | Send message (WebSocket upgrade or HTTP) |
| `GET` | `/knowledge-bases` | List knowledge bases |
| `POST` | `/knowledge-bases` | Create knowledge base |
| `GET` | `/knowledge-bases/{id}` | Get knowledge base details |
| `PUT` | `/knowledge-bases/{id}` | Update knowledge base |
| `DELETE` | `/knowledge-bases/{id}` | Delete knowledge base |
| `POST` | `/knowledge-bases/{id}/documents` | Upload document |
| `GET` | `/knowledge-bases/{id}/documents` | List documents |
| `DELETE` | `/knowledge-bases/{id}/documents/{doc_id}` | Delete document |
| `POST` | `/knowledge-bases/{id}/search` | Search knowledge base |

### Realtime

WebSocket endpoint for realtime subscriptions:

```
ws://localhost:8080/api/v1/realtime
```

Channels: `table:{schema}.{table}`, `presence:{room}`, `broadcast:{channel}`

### MCP (Model Context Protocol)

JSON-RPC 2.0 endpoint for AI assistant integration:

```
POST /mcp
```

See [MCP Server Guide](/guides/mcp/) for details.

## Query Parameters

Table endpoints support PostgREST-compatible query parameters:

| Parameter | Description | Example |
|-----------|-------------|---------|
| `select` | Columns to return | `?select=id,name,email` |
| `order` | Sort order | `?order=created_at.desc` |
| `limit` | Max results | `?limit=10` |
| `offset` | Pagination offset | `?offset=20` |
| `{column}.{op}` | Column filter | `?name.eq=John&age.gt=18` |

### Filter Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `eq` | Equal | `?status.eq=active` |
| `neq` | Not equal | `?status.neq=deleted` |
| `gt` | Greater than | `?age.gt=18` |
| `gte` | Greater than or equal | `?age.gte=18` |
| `lt` | Less than | `?price.lt=100` |
| `lte` | Less than or equal | `?price.lte=100` |
| `like` | Pattern match | `?name.like=John%` |
| `ilike` | Case-insensitive pattern | `?name.ilike=john%` |
| `in` | In list | `?status.in=(active,pending)` |
| `is` | Is null/not null | `?deleted_at.is.null` |

## Common Headers

| Header | Description |
|--------|-------------|
| `Authorization` | Bearer token for authentication (`Bearer <jwt>`) |
| `X-Client-Key` | Client key for key-based authentication |
| `X-FB-Tenant` | Tenant slug for multi-tenant context |
| `X-Fluxbase-Branch` | Branch name for database branching context |
| `Content-Type` | Request body format (`application/json`, `multipart/form-data`) |
| `Prefer` | Response preferences (`return=representation`, `count=exact`) |

## OpenAPI Specification

A live OpenAPI 3.0 specification is available at:

```
GET /openapi.json
```

This specification is generated dynamically based on your database schema and includes all available endpoints with their request/response schemas.

## Error Responses

Errors return JSON: `{"error": "description"}`. Standard HTTP status codes apply (400, 401, 403, 404, 409, 429, 500, 503).
