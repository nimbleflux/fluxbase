# Fluxbase Codebase Guide

Fluxbase is a single-binary Backend-as-a-Service (BaaS) - a lightweight Supabase alternative. PostgreSQL is the only external dependency.

## Stack

- **Backend:** Go 1.25+, Fiber v3, pgx/v5, golang-migrate, TimescaleDB
- **Admin UI:** React 19, Vite, TanStack Router/Query, Tailwind v4, shadcn/ui
- **SDKs:** TypeScript (`sdk/`), React hooks (`sdk-react/`), Go (`pkg/client/`)
- **Functions Runtime:** Deno (JavaScript/TypeScript edge functions)

## Directory Structure

```
cmd/fluxbase/main.go     # Server entry point
cli/cmd/                 # CLI commands (auth, functions, jobs, migrations, secrets)
internal/                # Core backend modules (see below)
admin/src/routes/        # Admin dashboard pages (file-based routing)
sdk/src/                 # TypeScript SDK source
deploy/helm/             # Kubernetes Helm charts
test/e2e/                # End-to-end tests
```

## Internal Modules (`internal/`)

| Module           | Purpose                                                                                                     |
| ---------------- | ----------------------------------------------------------------------------------------------------------- |
| `adminui/`       | Admin dashboard UI backend management                                                                       |
| `ai/`            | Vector search (pgvector), embeddings, knowledge bases, chatbots                                             |
| `api/`           | HTTP handlers (100+ files) - REST, GraphQL, storage, auth, DDL, webhooks, RPC, bulk operations, data export |
| `auth/`          | Authentication - JWT, OAuth2, OIDC, SAML SSO, magic links, MFA, CAPTCHA, impersonation                      |
| `branching/`     | Database branching - isolated DBs for dev/test environments                                                 |
| `config/`        | YAML + env var configuration loading                                                                        |
| `crypto/`        | Encryption utilities for secret storage                                                                     |
| `database/`      | PostgreSQL connection, schema introspection, migrations                                                     |
| `email/`         | SMTP, SendGrid, Mailgun, AWS SES providers                                                                  |
| `extensions/`    | PostgreSQL extension management system                                                                      |
| `functions/`     | Edge functions - Deno runtime, bundling, loader, scheduler                                                  |
| `jobs/`          | Background jobs - queue, workers, scheduler, progress tracking                                              |
| `logutil/`       | Log utilities (sanitization, formatting)                                                                    |
| `logging/`       | Structured logging with batching and retention policies                                                     |
| `mcp/`           | Model Context Protocol server for AI assistant integration                                                  |
| `middleware/`    | Auth, CORS, rate limiting, logging, branch and tenant context middlewares                                   |
| `migrations/`    | Database migration management                                                                               |
| `observability/` | Prometheus metrics and OpenTelemetry tracing                                                                |
| `pubsub/`        | Distributed pub/sub (local, PostgreSQL, Redis backends)                                                     |
| `query/`         | Shared query building types (FilterCondition, etc.)                                                         |
| `ratelimit/`     | Rate limiting service (memory, PostgreSQL, Redis backends)                                                  |
| `realtime/`      | WebSocket subscriptions via PostgreSQL LISTEN/NOTIFY                                                        |
| `rpc/`           | Remote procedure calls for database functions/procedures                                                    |
| `runtime/`       | Deno runtime wrapper for edge functions                                                                     |
| `scaling/`       | Horizontal scaling and leader election                                                                      |
| `secrets/`       | Secret management for functions/jobs                                                                        |
| `settings/`      | Application settings and custom configuration                                                               |
| `storage/`       | File storage abstraction (local filesystem or S3/MinIO)                                                     |
| `tenantdb/`      | Tenant database routing, FDW connections, separate tenant databases                                         |
| `testcontext/`   | Test context utilities for E2E tests                                                                        |
| `testutil/`      | Test utilities and helpers                                                                                  |
| `webhook/`       | Webhook system for database events (INSERT, UPDATE, DELETE)                                                 |

## Database Schemas

- `auth.*` - Users, sessions, identities, client keys
- `storage.*` - Buckets, objects, access policies
- `jobs.*` - Background job storage
- `functions.*` - Edge functions registry
- `branching.*` - Database branch metadata, access control, GitHub config
- `ai.*` - Knowledge bases, documents, chatbots, permissions
- `logging.*` - Centralized logging entries with TimescaleDB hypertable support
- `platform.*` - Multi-tenancy (tenants, service_keys, tenant_admin_assignments, users)
- `public` - User application tables

**Tenant Isolation:** All tenant-scoped tables use Row Level Security (RLS) with the `tenant_service` role for automatic tenant isolation. The `platform.tenants` table stores tenant metadata, and `platform.service_keys` manages API keys per tenant.

## Key Files by Feature

**Authentication:**

- `internal/auth/service.go` - Main auth logic
- `internal/auth/jwt.go` - Token management
- `internal/auth/scopes.go` - Authorization scopes
- `internal/api/auth_*.go` - Auth HTTP handlers

**REST API:**

- `internal/api/rest_crud.go` - CRUD operations
- `internal/api/query_parser.go` - URL query parsing
- `internal/api/query_builder.go` - SQL generation

**Edge Functions:**

- `internal/functions/handler.go` - Function HTTP handler
- `internal/functions/loader.go` - Load functions from disk
- `internal/runtime/runtime.go` - Deno runtime wrapper

**Background Jobs:**

- `internal/jobs/manager.go` - Job orchestration
- `internal/jobs/worker.go` - Job execution
- `internal/jobs/scheduler.go` - Cron scheduling

**Storage:**

- `internal/storage/service.go` - Storage abstraction
- `internal/api/storage_*.go` - Upload/download handlers

**Realtime:**

- `internal/realtime/manager.go` - Main connection manager
- `internal/realtime/connection.go` - Client connection handling
- `internal/realtime/subscription.go` - Subscription management
- `internal/realtime/presence.go` - User online status tracking

**MCP Server:**

- `internal/mcp/server.go` - JSON-RPC 2.0 protocol handler
- `internal/mcp/handler.go` - HTTP transport layer
- `internal/mcp/auth.go` - Auth context and scope checking
- `internal/mcp/tools/` - Tool implementations (query, storage, functions, jobs, vectors)
- `internal/mcp/resources/` - Resource providers (schema, functions, storage, rpc)

**Database Branching:**

- `internal/branching/manager.go` - CREATE/DROP DATABASE operations
- `internal/branching/storage.go` - Branch metadata CRUD
- `internal/branching/router.go` - Connection pool per branch
- `internal/api/branch_handler.go` - REST API for branch management
- `internal/api/github_webhook_handler.go` - GitHub PR automation
- `internal/middleware/branch.go` - Branch context extraction
- `cli/cmd/branch.go` - CLI commands

**GraphQL:**

- `internal/api/graphql_handler.go` - GraphQL HTTP handler
- `internal/api/graphql_resolvers.go` - Query/mutation resolvers
- `internal/api/graphql_schema.go` - Schema generation

**RPC/Procedures:**

- `internal/rpc/service.go` - Procedure execution
- `internal/api/rpc_handler.go` - RPC HTTP handlers

**Webhooks:**

- `internal/webhook/service.go` - Webhook delivery
- `internal/api/webhook_handler.go` - Webhook management API

**Observability:**

- `internal/observability/metrics.go` - Prometheus metrics
- `internal/observability/tracing.go` - OpenTelemetry tracing

**Multi-Backend Logging:**

- `internal/storage/log_service.go` - Main log service orchestration
- `internal/storage/log_postgres.go` - PostgreSQL native backend
- `internal/storage/log_timescaledb.go` - TimescaleDB backend with compression
- `internal/storage/log_s3.go` - S3/MinIO cloud storage backend
- `internal/storage/loki.go` - Loki integration

**Enhanced AI/Knowledge Base:**

- `internal/ai/knowledge_base.go` - Core data models
- `internal/ai/knowledge_base_storage.go` - Storage operations

**Multi-Tenancy:**

- `internal/api/tenant_handler.go` - Tenant CRUD HTTP handlers
- `internal/api/servicekey_handler.go` - Service key management API
- `internal/middleware/tenant.go` - Tenant context extraction middleware
- `internal/database/schema/schemas/platform.sql` - Platform schema with tenants table (declarative)

**Migrations:**

- `internal/migrations/handler.go` - Migrations API HTTP handlers (CRUD, apply, rollback, sync)
- `internal/migrations/executor.go` - Main database migration execution (`ExecuteWithAdminRole`)
- `internal/migrations/tenant_executor.go` - Tenant-scoped execution (`SET LOCAL ROLE tenant_migration_role`)
- `internal/migrations/storage.go` - Migration metadata CRUD (`migrations.app` table)
- `internal/migrations/declarative.go` - Declarative schema service (pgschema plan/apply/dump)
- `internal/tenantdb/declarative.go` - Per-tenant declarative schema service
- `internal/api/routes/migrations.go` - Migration route definitions (`/api/v1/admin/migrations`)
- `internal/middleware/migrations_security.go` - Migrations API auth, IP allowlist, rate limiting
- `internal/database/connection.go` - Filesystem migration runner (`runUserMigrations`)
- `cli/cmd/migrations.go` - CLI commands (`fluxbase migrations sync/list/create/apply/rollback`)

## Common Commands

### Devcontainer Database Access

When working in the devcontainer, you can use `psql` to query the database directly. The connection credentials are available from environment variables:

```bash
# Connect to the database
psql "postgresql://$FLUXBASE_DATABASE_USER:$FLUXBASE_DATABASE_PASSWORD@$FLUXBASE_DATABASE_HOST:$FLUXBASE_DATABASE_PORT/$FLUXBASE_DATABASE_NAME"
```

Useful queries for debugging:

```sql
-- List all tables in public schema
SELECT tablename FROM pg_tables WHERE schemaname = 'public';

-- Check service_role permissions on public tables
SELECT table_name, privilege_type
FROM information_schema.role_table_grants
WHERE grantee = 'service_role' AND table_schema = 'public';

-- Check tenants
SELECT id, name, slug FROM platform.tenants;

-- Find leftover test tables
SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE 'test_%';
```

### Build & Development

```bash
# Development
make dev              # Start backend + admin UI dev servers
make build            # Production build with embedded admin

# Database Operations
make db-reset         # Reset database (preserve user data)
make db-reset-full    # Full reset (destroys all data, bootstrap runs on next server start)

# Testing
make test             # Unit tests only (2min)
make test-coverage    # Unit tests with coverage report and enforcement
make test-coverage-unit # Unit tests with coverage (excludes e2e)
make test-full        # All tests including E2E (10min+)
make test-coverage-check  # Check coverage thresholds without running tests
make test-auth        # Authentication tests
make test-rls         # RLS security tests
make test-rest        # REST API tests
make test-storage     # Storage tests
make test-e2e         # E2E tests only
make test-e2e-fast   # Fast E2E tests

# SDK Tests
make test-sdk         # TypeScript SDK tests
make test-sdk-react   # SDK React build and type check

# Code Quality
make lint-go          # Go linting with golangci-lint
make lint-typescript  # TypeScript linting (admin UI + SDKs)

# CLI
make cli-install      # Build and install CLI

# Setup
make setup-dev        # Install dependencies + git hooks
```

## Configuration

Three-layer system: defaults → `fluxbase.yaml` → `FLUXBASE_*` env vars

Key config sections: server, database, auth, storage, realtime, functions, jobs, email, ai, mcp, branching, graphql, rpc, webhooks, scaling, observability (metrics, tracing), security, cors, api, logging, migrations, tenants

**Database Configuration (relevant to migrations):**

```yaml
database:
  user_migrations_path: "/migrations/user"  # Local path for filesystem migrations (env: FLUXBASE_DATABASE_USER_MIGRATIONS_PATH)
```

**Migrations API Configuration:**

```yaml
migrations:
  enabled: true
  allowed_ip_ranges: []  # IP CIDR allowlist (default: Docker/private networks)
```

**Tenant Declarative Schema Configuration:**

```yaml
tenants:
  declarative:
    enabled: true
    schema_dir: "schemas"          # Directory: {schema_dir}/{tenant-slug}/public.sql
    on_create: true                # Apply on tenant creation
    on_startup: false              # Apply on server startup
    allow_destructive: false       # Allow DROP/ALTER in tenant schemas
```

**MCP Configuration:**

```yaml
mcp:
  enabled: true
  base_path: /mcp
  rate_limit_per_min: 100
  allowed_tools: [] # Empty = all tools
  allowed_resources: [] # Empty = all resources
```

**Branching Configuration:**

```yaml
branching:
  enabled: true
  max_branches_per_user: 5
  max_total_branches: 50
  default_data_clone_mode: schema_only
  auto_delete_after: 24h
  database_prefix: branch_
  admin_database_url: "postgresql://..."
```

**Logging Backend Configuration:**

```yaml
logging:
  backend: timescaledb # postgres, timescaledb, s3, local, loki, elasticsearch, clickhouse
  batch_size: 100
  flush_interval: 5s
  retention_days: 90
  compression_days: 7 # For TimescaleDB
  s3:
    bucket: logs
    prefix: logs/
```

## Code Quality Standards

**MANDATORY REQUIREMENTS:** All code must pass these checks before committing.

### Go Code Quality

```bash
# Formatting (REQUIRED)
go fmt ./...

# Linting (REQUIRED - must pass)
golangci-lint run ./...

# Type Checking
golangci-lint run ./...  # Includes type checking
```

**What gets checked:**

- **gofmt**: Standard Go formatting (auto-fixed by pre-commit hook)
- **golangci-lint**: Comprehensive linting including:
  - gocritic: Code improvement suggestions
  - misspell: Spell checking
  - govet: Static analysis
  - type checking: Type safety verification

**Configuration:** `.golangci.yml`

- Timeout: 5 minutes
- Tests included
- Integration build tags enabled

### TypeScript Code Quality

```bash
# Admin UI
cd admin && bun run type-check
cd admin && bun run lint

# SDK
cd sdk && bun run type-check
cd sdk && bun run lint

# SDK React
cd sdk-react && bun run type-check  # Uses tsc --noEmit
```

**What gets checked:**

- **ESLint**: TypeScript ESLint, React Hooks, React Refresh, TanStack Query
- **Prettier**: Code formatting with import sorting and Tailwind integration
- **TypeScript**: No unused vars, type-only imports enforced
1
### Pre-Commit Hook Enforcement

Git pre-commit hooks automatically run:

1. `go fmt ./...` - Auto-stages formatted files
2. `golangci-lint run ./...` - Blocks commit if fails
3. Admin UI type-check - Blocks commit if fails
4. SDK type-check - Blocks commit if fails
5. SDK React type-check - Blocks commit if fails

### CI/CD Enforcement

- **Go**: Formatting check + golangci-lint must pass
- **TypeScript**: ESLint must pass
- **Tests**: Coverage thresholds enforced (25% overall)
- **Build**: Cross-platform builds (Linux/amd64 + Linux/arm64)

## Patterns

- Interface-based dependency injection
- Handler pattern with `*fiber.Ctx`
- Repository pattern for data access
- PostgreSQL Row Level Security (RLS) for authorization
- PostgREST-compatible REST API conventions

## Migrations

Fluxbase uses a **hybrid migration system** with three subsystems:

### Internal Schema (Declarative)

Internal Fluxbase tables (auth, storage, functions, jobs, etc.) are managed declaratively:

- **Bootstrap:** `internal/database/bootstrap/bootstrap.sql` - Creates schemas, extensions, roles, default privileges
- **Schema files:** `internal/database/schema/schemas/*.sql` - Declarative SQL files for each schema
- **Applied automatically:** Server applies bootstrap + declarative schema on startup

### User Schema - Imperative Migrations

Imperative migrations are tracked in the `migrations.app` table and can be delivered via a local filesystem path or the API:

**Local Filesystem:**

- Config: `database.user_migrations_path` (default: `/migrations/user`)
- Env var: `FLUXBASE_DATABASE_USER_MIGRATIONS_PATH`
- Files: `{name}.up.sql` / `{name}.down.sql` pairs, sorted alphabetically (e.g. `001_create_users.up.sql`)
- Applied at startup against the main database as admin user
- Tracked with `namespace='filesystem'`
- SQL validated via `pg_query.Parse()` before execution

**Migrations API (`/api/v1/admin/migrations`):**

- Requires service key or `service_role` JWT with `admin`, `instance_admin`, or `tenant_admin` role
- Endpoints: CRUD, apply, rollback, apply-pending, sync
- `POST /sync` accepts a batch of migrations, deduplicates by SHA256 of up/down SQL, and optionally auto-applies
- Configurable namespaces (default: `"default"`)

**Tenant-aware routing (backward compatible):**

- **No tenant context** (no `X-FB-Tenant` header, not `tenant_admin` JWT): runs via `db.ExecuteWithAdminRole()` against the main database with full DDL privileges
- **Default tenant** (`X-FB-Tenant` points to default tenant): `Router.GetPool()` returns the main pool, `TenantExecutor` runs with `SET LOCAL ROLE tenant_migration_role` (restricted to `public` schema)
- **Named tenant**: `Router.GetPool()` returns a pool to the tenant's separate database, same `tenant_migration_role` restriction to `public` schema

### User Schema - Declarative (pgschema)

Per-tenant declarative schema management using pgschema for diff-based application to the `public` schema:

**Local Filesystem:**

- Config: `tenants.declarative.schema_dir`
- Structure: `{SchemaDir}/{tenant-slug}/public.sql`
- Applied on tenant creation (`on_create`), server startup (`on_startup`), or on-demand via API

**Tenant Schema API:**

- `GET /tenants/:id/schema` - Get schema status and pending changes
- `POST /tenants/:id/schema/content` - Upload schema SQL (stored in `platform.tenant_schemas`)
- `POST /tenants/:id/schema/content/apply` - Diff and apply uploaded content via pgschema
- `POST /tenants/:id/schema/apply` - Apply from local filesystem
- `DELETE /tenants/:id/schema/content` - Delete stored schema

Works for the default tenant too — `Router.GetPool()` returns the main pool when `UsesMainDatabase()` is true.

### Common Commands

```bash
# Database Operations
make db-reset         # Reset database (preserve user data)
make db-reset-full    # Full reset - bootstrap runs on next server start

# CLI Migrations (interact with server API)
fluxbase migrations list [--namespace]      # List migrations
fluxbase migrations create <name>           # Create migration
fluxbase migrations apply <name>            # Apply a migration
fluxbase migrations rollback <name>         # Rollback a migration
fluxbase migrations apply-pending           # Apply all pending
fluxbase migrations sync [--dir] [--namespace] [--no-apply]  # Sync from directory
```

## Testing

### Test Organization

- Unit tests: `*_test.go` alongside source
- E2E tests: `test/e2e/`
- Test helpers: `internal/testutil/`

### Coverage Targets

- **Overall:** 25%+ (starting point, will increase)
- **Core business logic:** 50%+ per file
- **Critical modules (auth, API):** 70%+ per file

### Excluded from Coverage

Files containing only type definitions, interfaces, or requiring external system dependencies are excluded from coverage calculations. See [.testcoverage.yml](.testcoverage.yml) for the complete list:

- Pure type definition files (e.g., `internal/*/types.go`, `internal/*/errors.go`)
- Interface-only files (e.g., `internal/auth/interfaces.go`)
- Infrastructure code requiring system dependencies (leader election, database connections, OCR)
- CLI commands (tested via integration tests)
- Entry points and embedded assets

### Running Tests

```bash
make test             # Unit tests only (2min)
make test-coverage    # Unit tests with coverage report and enforcement
make test-full        # All tests including E2E (10min)
make test-coverage-check  # Check coverage thresholds without running tests
```

### Coverage Enforcement

Coverage thresholds are enforced in CI via [go-test-coverage](https://github.com/vladopajic/go-test-coverage). Pull requests must meet minimum thresholds for affected files. The tool automatically excludes files that shouldn't be counted (pure type definitions, infrastructure code, etc.).

## Development Workflow Requirements

### Writing Tests

**IMPORTANT:** When making code changes, always consider writing or updating tests:

1. **New features** - Write unit tests covering the main functionality and edge cases
2. **Bug fixes** - Add a regression test that would have caught the bug
3. **Refactoring** - Ensure existing tests still pass; add tests if coverage gaps exist

**Test file locations:**

- Unit tests: Place `*_test.go` files alongside the source file being tested
- E2E tests: Add to `test/e2e/` for integration scenarios
- Test helpers: Use `internal/testutil/` for shared test utilities

**Test naming conventions:**

```go
func TestFunctionName_Scenario_ExpectedBehavior(t *testing.T)
// Example: TestCreateBranch_ExceedsUserLimit_ReturnsError
```

**When to skip tests:**

- Pure type definitions or interface files
- Simple configuration structs with no logic
- Code that only wraps external dependencies (but do test the integration)

### Updating Documentation

**IMPORTANT:** When making code changes, always consider updating documentation:

1. **New features** - Add documentation in `docs/src/content/docs/guides/`
2. **API changes** - Update SDK documentation in `docs/src/content/docs/api/`
3. **Configuration changes** - Update the relevant guide and CLAUDE.md if needed
4. **Breaking changes** - Document migration steps clearly

**Documentation locations:**

- Feature guides: `docs/src/content/docs/guides/<feature>.md`
- API reference: `docs/src/content/docs/api/` (auto-generated from SDK)
- Project overview: `CLAUDE.md` (this file)
- Implementation notes: `IMPLEMENTATION_ANALYSIS.md`

**Documentation checklist:**

- [ ] Does the feature documentation match the implementation?
- [ ] Are all configuration options documented?
- [ ] Are error messages and edge cases explained?
- [ ] Are code examples correct and runnable?

### Pre-Commit Checklist

Before committing changes, verify:

1. `go fmt ./...` passes (or auto-fixed by hook)
2. `golangci-lint run ./...` passes
3. TypeScript type-check passes (admin UI + SDKs)
4. `make test` passes
5. Documentation is updated for user-facing changes
6. New tests are added for new functionality
7. Existing tests are updated if behavior changed
