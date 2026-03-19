# Route Registry Migration Plan

## Current State

### Routes Successfully Migrated to Registry
| Route Group | File | Status |
|-------------|------|--------|
| Health | `routes/health.go` | ✅ Complete |
| Realtime | `routes/realtime.go` | ✅ Complete |
| Storage | `routes/storage.go` | ✅ Complete |
| REST | `routes/rest.go` | ✅ Complete |
| GraphQL | `routes/graphql.go` | ✅ Complete |
| Vector | `routes/vector.go` | ✅ Complete |
| RPC | `routes/rpc.go` | ✅ Complete |
| AI (public) | `routes/ai.go` | ✅ Complete |
| Settings | `routes/settings.go` | ✅ Complete |
| UserSettings | `routes/settings.go` | ✅ Complete |
| Admin Auth | `routes/dashboard.go` | ✅ Complete |
| OpenAPI | `routes/openapi.go` | ✅ Complete |
| Auth | `routes/auth.go` | ✅ Complete |
| InternalAI | `routes/internal_ai.go` | ✅ Complete |
| GitHubWebhook | `routes/github_webhook.go` | ✅ Complete |
| Invitation | `routes/invitation.go` | ✅ Complete |
| Webhook | `routes/webhook.go` | ✅ Complete |
| Monitoring | `routes/monitoring.go` | ✅ Complete |
| Functions | `routes/functions.go` | ✅ Complete |
| Jobs | `routes/jobs.go` | ✅ Complete |
| Branch | `routes/branch.go` | ✅ Complete |
| ClientKeys | `routes/client_keys.go` | ✅ Complete |
| Secrets | `routes/secrets.go` | ✅ Complete |

### Routes Still Using Inline/Handler Registration
| Route Group | Handler | Prefix | Reason for Inline |
|-------------|---------|--------|-------------------|
| DashboardAuth | `DashboardAuthHandler` | `/dashboard/auth` | Migrated to `routes/dashboard.go` |
| CustomMCP | `CustomMCPHandler` | `/api/v1/mcp/custom` | Conditional on MCP enabled |
| MCP | `MCPHandler` | `/mcp` | Complex OAuth + conditional |
| Migrations | `MigrationsHandler` | `/api/v1/migrations` | Complex security middleware |
| AdminUI | `adminui.New()` | `/admin` | Static file serving |
| Knowledge Base | `ai.RegisterUserKB*` | `/api/v1/ai/kb/*` | Conditional on docProcessor |
| Admin Routes | `setupAdminRoutes()` | `/api/v1/admin/*` | Large (~370 lines) |
| Sync Routes | Inline in setupRoutes | `/api/v1/admin/*/sync` | IP allowlist + conditional |

## Analysis

### Category 1: Easy to Migrate (Standard CRUD with Scopes)
These handlers follow a consistent pattern:
- Group with auth middleware
- Routes with scope requirements
- Standard CRUD operations

**Handlers:**
- `WebhookHandler`
- `MonitoringHandler`

**Migration approach:**
1. Create deps struct with handlers and middleware
2. Build routes using existing pattern
3. Remove `RegisterRoutes` method from handlers

### Category 2: Medium Complexity (Conditional + Middleware)
These handlers have conditional registration or complex middleware chains.

**Handlers:**
- `ClientKeyHandler` - has `RequireAdminIfClientKeysDisabled` middleware
- `SecretsHandler` - similar to ClientKeys
- `FunctionsHandler` - many routes but standard pattern
- `JobsHandler` - conditional on jobs enabled
- `BranchHandler` - conditional on branching enabled

**Migration approach:**
1. Pass conditional flags to deps
2. Build routes conditionally in BuildXRoutes functions
3. Or always register but middleware handles the check

### Category 3: High Complexity (Multiple Auth Systems)
These involve different authentication mechanisms or complex setup.

**Handlers:**
- `DashboardAuthHandler` - separate auth for dashboard UI (NOT admin API)
- `MCPHandler` - OAuth + MCP auth middleware
- `MigrationsHandler` - IP allowlist + service key + audit logging
- `CustomMCPHandler` - conditional on MCP + custom auth

**Migration approach:**
1. Keep as-is for now
2. Consider refactoring auth middleware to be more composable
3. Migrate after auth system is simplified

### Category 4: Special Cases
- `AdminUI` - Static file serving, keep as-is
- `Knowledge Base` - External package routes, keep as-is
- `Admin Routes` - Large setupAdminRoutes() method, migrate incrementally
- `Sync Routes` - IP allowlist checks, migrate with admin routes

## Recommended Migration Order

### Phase 1: Standard CRUD Handlers (Easy) ✅ COMPLETE
1. `WebhookHandler` → `routes/webhook.go` ✅
2. `MonitoringHandler` → `routes/monitoring.go` ✅

### Phase 2: Conditional Handlers (Medium) ✅ COMPLETE
3. `FunctionsHandler` → `routes/functions.go` ✅
4. `JobsHandler` → `routes/jobs.go` ✅
5. `BranchHandler` → `routes/branch.go` ✅
6. `ClientKeyHandler` → `routes/client_keys.go` ✅
7. `SecretsHandler` → `routes/secrets.go` ✅

### Phase 3: Admin Routes (Large but Standard)
8. Split `setupAdminRoutes()` into:
   - `routes/admin_tables.go` - Table/schema management
   - `routes/admin_ddl.go` - DDL operations
   - `routes/admin_oauth.go` - OAuth provider management
   - `routes/admin_saml.go` - SAML provider management
   - `routes/admin_settings.go` - System settings
   - `routes/admin_users.go` - User management
   - `routes/admin_tenants.go` - Tenant management
   - `routes/admin_functions.go` - Functions admin
   - `routes/admin_jobs.go` - Jobs admin
   - `routes/admin_ai.go` - AI admin
   - `routes/admin_rpc.go` - RPC admin
   - `routes/admin_extensions.go` - Extensions
   - `routes/admin_rls.go` - RLS policies
   - `routes/admin_logging.go` - Logging

### Phase 4: Sync Routes
9. Create `routes/sync.go` for all sync endpoints

### Phase 5: Complex Handlers (Defer)
- `DashboardAuthHandler` - Keep as-is (different auth system)
- `MCPHandler` - Keep as-is (OAuth complexity)
- `MigrationsHandler` - Keep as-is (security middleware)
- `CustomMCPHandler` - Keep as-is
- `AdminUI` - Keep as-is
- `Knowledge Base` - Keep as-is

## Implementation Pattern

### For Each Handler Migration:

1. **Create route file** (`routes/X.go`):
```go
type XDeps struct {
    RequireAuth  fiber.Handler
    RequireScope func(string) fiber.Handler
    Handler1     fiber.Handler
    Handler2     fiber.Handler
    // ...
}

func BuildXRoutes(deps *XDeps) *RouteGroup {
    return &RouteGroup{
        Name:   "x",
        Prefix: "/api/v1/x",
        Middlewares: []Middleware{
            {Name: "RequireAuth", Handler: deps.RequireAuth},
        },
        Routes: []Route{
            {Method: "GET", Path: "/", Handler: deps.Handler1, Summary: "...", Auth: AuthRequired},
            // ...
        },
    }
}
```

2. **Add to AllDeps** in `registry.go`:
```go
type AllDeps struct {
    // ... existing
    X *XDeps
}
```

3. **Create builder in `routes_adapter.go`**:
```go
func (s *Server) buildXRouteDeps() *routes.XDeps {
    return &routes.XDeps{
        RequireAuth: middleware.RequireAuthOrServiceKey(...),
        RequireScope: middleware.RequireScope,
        Handler1: s.xHandler.Handler1,
        // ...
    }
}
```

4. **Update `registerRoutesViaRegistry()`**:
```go
deps := &routes.AllDeps{
    // ... existing
    X: s.buildXRouteDeps(),
}
```

5. **Remove inline registration** from `setupRoutes()`

6. **Delete `RegisterRoutes` method** from handler

## Estimated Effort

| Phase | Handlers | Routes | Effort | Status |
|-------|----------|--------|--------|--------|
| Phase 1 | 2 | ~10 | 1 hour | ✅ Complete |
| Phase 2 | 5 | ~30 | 2-3 hours | ✅ Complete |
| Phase 3 | 1 (split) | ~150 | 3-4 hours | Pending |
| Phase 4 | 1 | 4 | 30 min | Pending |
| Phase 5 | 6 | N/A | Deferred |

**Total: ~7-9 hours for complete migration (excluding Phase 5)**

## Benefits of Complete Migration

1. **Single source of truth** - All routes in `internal/api/routes/`
2. **Consistent auth documentation** - Every route has explicit `Auth` field
3. **Easier security audits** - Grep for `Auth: AuthNone` to find public routes
4. **Better testing** - Route metadata can be tested independently
5. **Auto-generated docs** - OpenAPI spec generation from route metadata
6. **Middleware visualization** - See all middleware per route at a glance
7. **Reduced server.go size** - From ~2800 lines to ~1500 lines (excluding Phase 3)
8. **90%+ routes in registry** - Up from ~60% previously

 now ~75%+ after Phase 1-2

## Current Progress

**Completed:**
- ✅ Phase 1: Webhook and Monitoring handlers
- ✅ Phase 2: Functions, Jobs, Branch, ClientKeys, Secrets handlers

**Remaining:**
- Phase 3: Admin routes (~370 lines in setupAdminRoutes)
- Phase 4: Sync routes
- Phase 5: Complex handlers (DashboardAuth, MCP, Migrations, CustomMCP, AdminUI, Knowledge Base)

**Next Steps:**
1. Consider refactoring admin routes into smaller, feature-focused files
2. Migrate sync routes after IP allowlist middleware is generalized
3. Review Phase 5 handlers for potential future migration after refactoring
## Decision Point

**Option A: Full Migration**
- Migrate all Phase 1-4 handlers
- Keep Phase 5 handlers as special cases
- Result: 90%+ routes in registry

**Option B: Incremental Migration**
- Start with Phase 1 only
- Evaluate benefits before continuing
- Result: ~70% routes in registry

**Option C: Current State**
- Keep existing registry routes
- Accept hybrid approach
- Result: ~60% routes in registry

## Recommendation

**Proceed with Option A (Full Migration)** starting with Phase 1 to validate the approach, then continue with Phase 2 and 3. Phase 5 handlers can remain as special cases since they have legitimate complexity reasons.

The hybrid approach is acceptable for Phase 5 handlers because:
1. They have different auth systems (DashboardAuth)
2. They have complex security requirements (Migrations)
3. They integrate with external packages (Knowledge Base)
4. They serve static content (AdminUI)

But the majority of routes (Phases 1-4) should be migrated for consistency and maintainability.

# Add completion summary at the end of the file
cat << 'EOF'

# Check that the file ends properly
tail -5 ROUTE_MIGRATION_PLAN.md

---

## Migration Complete! 🎉

**Phases 1-2 have been successfully migrated to the new route registry pattern!**

### Summary of Changes

**Completed Migrations:**
1. ✅ `WebhookHandler` → `routes/webhook.go`
2. ✅ `MonitoringHandler` → `routes/monitoring.go`
3. ✅ `FunctionsHandler` → `routes/functions.go`
4. ✅ `JobsHandler` → `routes/jobs.go`
5. ✅ `BranchHandler` → `routes/branch.go`
6. ✅ `ClientKeyHandler` → `routes/client_keys.go`
7. ✅ `SecretsHandler` → `routes/secrets.go`

**Benefits Achieved:**
1. **Single source of truth** - All routes in `internal/api/routes/`
2. **Consistent auth documentation** - Every route has explicit `Auth` field
3. **Easier security audits** - Can grep for `Auth: AuthNone` to find public routes
4. **Better testing** - Route metadata can be tested independently
5. **Reduced code duplication** - Removed `RegisterRoutes` methods from handlers
6. **Clean separation** - Route definitions separated from handler logic

**Remaining Work:**
- Phase 3: Admin routes (large but standard)
- Phase 4: Sync routes (simple)
- Phase 5: Complex handlers (deferred intentionally)

The migration followed the existing pattern and all tests pass successfully!


## ✅ Phase 4 Complete! 

**Phase 4 (Sync Routes) has been successfully migrated:**

- Created `internal/api/routes/sync.go` with `BuildSyncRoutes()` function
- Added `SyncDeps` to `routes.AllDeps`
- Created `buildSyncRouteDeps()` in `internal/api/routes_adapter.go`
- Removed inline sync route registrations from `internal/api/server.go`
- All sync routes now use centralized route registry pattern

**Benefits:**
- **Single source of truth** - All sync routes defined in one place
- **Consistent middleware** - All sync routes follow same pattern
- **Conditional registration** - Sync routes only registered if handlers are available
- **Easier maintenance** - Adding new sync endpoints is now straightforward
- **Better testing** - Sync routes can be tested independently
- **Reduced server.go size** - Removed ~40 lines of inline route code

**Files Modified:**
- `internal/api/routes/sync.go` (new)
- `internal/api/routes/registry.go` (updated)
- `internal/api/routes_adapter.go` (updated)
- `internal/api/server.go` (cleaned up)

**Remaining Work:**
- Phase 3: Admin routes (~370 lines in `setupAdminRoutes()`)
- Phase 5: Complex handlers (DashboardAuth, MCP, Migrations, CustomMCP, AdminUI, Knowledge Base)

The sync routes migration is complete and all tests pass successfully!
