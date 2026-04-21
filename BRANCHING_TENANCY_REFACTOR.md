# Branching + Multi-Tenancy Refactoring Summary

## Changes Made

### Phase 1: Branch Clone Source Refactoring

**Problem:** When a tenant with a separate database created a branch, it cloned from the main database (not the tenant's DB). A tenant admin would expect to branch their own data.

**Solution:** The branching Manager now resolves the correct template database per tenant.

**Files changed:**
- `internal/branching/manager.go` — Added `TenantResolver` and `FDWRepairer` interfaces, `resolveTemplateDatabase()`, and `repairFDW()` methods. `createDatabaseSchemaOnly()` and `createDatabaseFullClone()` now use the resolved template.
- `internal/tenantdb/fdw.go` — Added `GetFDWRoleForTenant()` to look up existing FDW role credentials from a tenant database's user mapping, and `extractOptionValue()` helper.
- `internal/tenantdb/manager.go` — Added `RepairFDWForBranch()` which connects to the branch database and recreates the FDW user mapping using the tenant's FDW role credentials.
- `internal/api/server.go` — Added `branchTenantResolver` and `branchFDWRepairer` adapter types, wired to the branch Manager during initialization.

### Phase 2: API Registry — Missing TenantContext

**Problem:** Several route groups accessed tenant-scoped data without running `TenantMiddleware`, meaning the `X-FB-Tenant` header was ignored.

**Files changed:**
- `internal/api/routes/sync.go` — Added `TenantContext` field to `SyncDeps`, added group-level middleware.
- `internal/api/routes/migrations.go` — Added `TenantContext` field to `MigrationsDeps`, added group-level middleware.
- `internal/api/routes/openapi.go` — Added `TenantContext` and `TenantDBContext` fields to `OpenAPIDeps`, added group-level middleware.
- `internal/api/routes_adapter.go` — Wired `s.Middleware.Tenant` and `s.Middleware.TenantDB` into the three route groups above.

### Phase 3: Security Fixes

#### 3a. Key type role mapping (CRITICAL)
- `internal/middleware/clientkey_auth.go` — Changed `mapKeyTypetoRole` default from `"service_role"` to `"anon"`. Unknown key types now get the most restrictive role instead of full admin access. Includes a warning log for unrecognized types.
- `internal/middleware/clientkey_auth_test.go` — Updated test expectations.

#### 3b. Tenant ID resolution (MEDIUM)
- `internal/middleware/tenant_db.go` — Fixed `resolveTenantID()` to return `("", "")` instead of `(headerTenant, "header")` when the `X-FB-Tenant` header value doesn't match any known tenant. Previously, an invalid tenant slug would be passed through to RLS context, potentially causing unexpected behavior.

#### 3c. SSO login tenant context (MEDIUM)
- `internal/auth/platform.go` — Extracted tenant membership resolution into `resolveTenantMembership()` helper. Updated `LoginViaSSO()` to call `GenerateTokenPairWithTenant()` with proper tenant context (previously used `GenerateTokenPair()` without tenant info, causing SSO-authenticated dashboard users to lack tenant context in their JWT).

### Phase 4: Test Coverage

**New test files:**
- `internal/branching/manager_tenant_clone_test.go` — 15 tests covering:
  - `resolveTemplateDatabase()` for default tenant, named tenant with separate DB, tenant with resolver errors
  - `repairFDW()` skip conditions (no tenant, no repairer)
  - `SetTenantResolver()` / `SetFDWRepairer()` wiring
  - `TenantDatabaseInfo` struct validation

- `internal/middleware/tenant_db_test.go` — Added 4 tests:
  - `TestResolveTenantID_ValidSlug` — valid tenant slug resolves correctly
  - `TestResolveTenantID_InvalidSlugReturnsEmpty` — invalid slug returns empty (validates the fix)
  - `TestResolveTenantID_FallbackToDefault` — falls back to default tenant
  - `TestResolveTenantID_JWTClaims` — JWT tenant claims work

### Phase 5: Documentation

- `CLAUDE.md` — Added "Branching + Multi-Tenancy Interaction" section covering pool priority, clone source, FDW repair, default tenant, route coverage, and key files. Updated branching config to include `max_branches_per_tenant`.

---

## Files Changed Summary

| File | Change Type | Description |
|------|-------------|-------------|
| `internal/branching/manager.go` | Refactored | Tenant-aware clone source, FDW repair |
| `internal/tenantdb/fdw.go` | Added | `GetFDWRoleForTenant`, `extractOptionValue` |
| `internal/tenantdb/manager.go` | Added | `RepairFDWForBranch` |
| `internal/api/server.go` | Added | Adapter types for tenant resolver/FDW repairer |
| `internal/api/routes/sync.go` | Fixed | Added TenantContext middleware |
| `internal/api/routes/migrations.go` | Fixed | Added TenantContext middleware |
| `internal/api/routes/openapi.go` | Fixed | Added TenantContext + TenantDBContext |
| `internal/api/routes_adapter.go` | Fixed | Wired tenant middleware to 3 route groups |
| `internal/middleware/clientkey_auth.go` | Security fix | Default key type → anon instead of service_role |
| `internal/middleware/tenant_db.go` | Security fix | Invalid tenant header returns empty |
| `internal/auth/platform.go` | Bug fix | SSO login includes tenant context |
| `CLAUDE.md` | Documentation | Branching + multi-tenancy section |

## New Test Files

| File | Tests |
|------|-------|
| `internal/branching/manager_tenant_clone_test.go` | 15 tests |
| `internal/middleware/tenant_db_test.go` (extended) | 4 tests |

---

## Follow-Up Testing Instructions

### 1. Verify All Existing Tests Pass

```bash
# Unit tests (no DB required)
make test

# All tests including E2E (requires PostgreSQL)
make test-full
```

### 2. Test Branch Clone Source for Tenant DBs

This requires a running PostgreSQL instance:

```bash
# Start dev environment
make dev

# Create a tenant with a separate database
curl -X POST http://localhost:8080/api/v1/admin/tenants \
  -H "Authorization: Bearer <admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "Acme Corp", "slug": "acme", "db_mode": "separate"}'

# Create a branch for the tenant
curl -X POST http://localhost:8080/api/v1/admin/branches \
  -H "Authorization: Bearer <admin-token>" \
  -H "X-FB-Tenant: acme" \
  -H "Content-Type: application/json" \
  -d '{"name": "Feature Branch", "clone_mode": "schema_only"}'

# Verify the branch database was cloned from the tenant's database
psql -c "SELECT datname FROM pg_database WHERE datname LIKE 'branch_%';"
```

### 3. Test Sync Routes with Tenant Context

```bash
# Sync functions with tenant context
curl -X POST http://localhost:8080/api/v1/admin/functions/sync \
  -H "Authorization: Bearer <admin-token>" \
  -H "X-FB-Tenant: acme"

# Verify functions were created with the correct tenant_id
psql -c "SELECT name, tenant_id FROM functions.functions_registry;"
```

### 4. Test Migration Routes with Tenant Context

```bash
# List migrations for a specific tenant
curl http://localhost:8080/api/v1/admin/migrations/ \
  -H "X-Service-Key: <tenant-service-key>" \
  -H "X-FB-Tenant: acme"
```

### 5. Test OpenAPI with Tenant Context

```bash
# Get OpenAPI spec for a specific tenant (should show tenant's tables)
curl http://localhost:8080/openapi.json \
  -H "Authorization: Bearer <admin-token>" \
  -H "X-FB-Tenant: acme"
```

### 6. Verify Security Fixes

```bash
# Test that unknown key types get anon role (not service_role)
# This should return 403 or limited access
curl http://localhost:8080/api/v1/tables/public/users \
  -H "X-Service-Key: <key-with-unknown-type>"

# Test that invalid tenant slug is rejected
curl http://localhost:8080/api/v1/tables/public/users \
  -H "Authorization: Bearer <user-token>" \
  -H "X-FB-Tenant: nonexistent-tenant"
# Should return 403 or fall back to default tenant (not use raw header)
```

### 7. Verify SSO Login Tenant Context

```bash
# After SSO login, verify the JWT contains tenant_id
# Decode the access token and check claims
echo "<sso-access-token>" | cut -d. -f2 | base64 -d 2>/dev/null | jq .
# Should contain "tenant_id" field
```

### 8. E2E Tests to Run

```bash
# Full E2E suite
make test-e2e

# Tenant isolation tests specifically
make test-e2e-fast -- -run TestTenantIsolation

# Playwright UI tests
make test-e2e-ui
```

---

## Remaining Items (Not Included in This Change)

| Priority | Item | Notes |
|----------|------|-------|
| LOW | Dashboard TOTP secrets encryption | `platform.users.totp_secret` stores plaintext; should use `crypto.Encrypt()` |
| MEDIUM | Add `max_branches_per_tenant` to branching config docs | Config exists but may not be in all docs |
| MEDIUM | E2E test for branch + tenant combination | Full lifecycle test with real HTTP calls |
| MEDIUM | E2E test for tenant deletion cascading to branches | Verify branch records/DDBs cleaned up |
| MEDIUM | Integration test for tenant + branch middleware interaction | Both middlewares running in sequence |
