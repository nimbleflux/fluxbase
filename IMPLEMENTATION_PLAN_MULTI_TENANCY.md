# Multi-Tenancy Implementation Plan

**Status:** In Progress  
**Created:** 2026-03-13  
**Last Updated:** 2026-03-13

## Overview

This plan implements row-based multi-tenancy in Fluxbase with PostgreSQL RLS for tenant isolation. The implementation is fully backward-compatible through a synthetic default tenant.

## Architecture Summary

```
┌─────────────────────────────────────────────────────────────────┐
│                     Fluxbase Instance                            │
├─────────────────────────────────────────────────────────────────┤
│  Instance Admins (instance_admin role)                           │
│  - Manage all tenants, instance configuration                    │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Tenant A      │  │   Tenant B      │  │ Default Tenant  │  │
│  │  tenant_admin   │  │  tenant_admin   │  │  (backward-compat)│
│  │  tenant_members │  │  tenant_members │  │  tenant_members │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Tenant naming | `tenants` | Clear SaaS terminology, industry standard |
| Tenant selection | Header (`X-FB-Tenant`) + JWT claim | Flexible, backward-compatible |
| Service client | Explicit `forTenant()` method | Reduces footguns, makes intent clear |
| Role migration | Setup admin → instance_admin, others → tenant_admin | Preserves current semantics |
| Default tenant | Synthetic "default" tenant owns all legacy data | Backward compatibility |

---

## Phase Overview

| Phase | Name | Duration | Dependencies | Parallelizable |
|-------|------|----------|--------------|----------------|
| 1 | Database Schema | 2-3 days | None | Yes (4 tracks) |
| 2 | Backend Core | 3-4 days | Phase 1 | Yes (5 tracks) |
| 3 | Migration Logic | 1-2 days | Phase 1, 2 | Partial (2 tracks) |
| 4 | Dashboard UI | 2-3 days | Phase 2 | Yes (4 tracks) |
| 5 | SDK Updates | 1-2 days | Phase 2 | Yes (3 tracks) |
| 6 | Testing & Docs | 2-3 days | All phases | Yes (3 tracks) |

---

## Phase 1: Database Schema

**Goal:** Create tenant infrastructure at the database level

**Dependencies:** None

**Parallel Tracks:** 4 (can run simultaneously)

### Track 1A: Core Tenant Tables

**File:** `internal/database/migrations/097_tenants.up.sql`

**Deliverables:**
- [ ] `tenants` table with columns: `id`, `slug`, `name`, `is_default`, `metadata`, `created_at`, `updated_at`, `deleted_at`
- [ ] `tenant_memberships` table with columns: `id`, `tenant_id`, `user_id`, `role`, `created_at`, `updated_at`
- [ ] Unique constraint on `(tenant_id, user_id)`
- [ ] Indexes on `tenant_memberships(user_id)` and `tenant_memberships(tenant_id)`
- [ ] Foreign key constraints to `auth.users` and `tenants`

**SQL Template:**
```sql
-- 091_tenants.up.sql

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    is_default BOOLEAN DEFAULT false,
    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_is_default ON tenants(is_default) WHERE is_default = true;
CREATE INDEX idx_tenants_deleted_at ON tenants(deleted_at) WHERE deleted_at IS NOT NULL;

COMMENT ON TABLE tenants IS 'Logical tenants within a Fluxbase instance for multi-tenancy support';
COMMENT ON COLUMN tenants.slug IS 'URL-friendly identifier for the tenant';
COMMENT ON COLUMN tenants.is_default IS 'True for the synthetic default tenant (backward compatibility)';

CREATE TABLE IF NOT EXISTS tenant_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'tenant_member',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT tenant_memberships_unique UNIQUE(tenant_id, user_id),
    CONSTRAINT tenant_memberships_role_check CHECK (role IN ('tenant_admin', 'tenant_member'))
);

CREATE INDEX idx_tenant_memberships_user_id ON tenant_memberships(user_id);
CREATE INDEX idx_tenant_memberships_tenant_id ON tenant_memberships(tenant_id);
CREATE INDEX idx_tenant_memberships_role ON tenant_memberships(role);

COMMENT ON TABLE tenant_memberships IS 'Maps users to tenants with specific roles';
COMMENT ON COLUMN tenant_memberships.role IS 'User role within tenant: tenant_admin or tenant_member';

-- 091_tenants.down.sql

DROP TABLE IF EXISTS tenant_memberships;
DROP TABLE IF EXISTS tenants;
```

**Verification:**
```sql
-- Run after migration
SELECT * FROM tenants LIMIT 0;
SELECT * FROM tenant_memberships LIMIT 0;
```

---

### Track 1B: RLS Helper Functions

**File:** `internal/database/migrations/098_functions_tenancy.up.sql`

**Deliverables:**
- [ ] `current_tenant_id()` function - Returns active tenant ID from JWT or session
- [ ] `user_has_tenant_role(user_id, tenant_id, role)` function - Check membership
- [ ] `is_instance_admin(user_id)` function - Check instance admin status
- [ ] `current_tenant_role()` function - Get user's role in current tenant
- [ ] `user_tenant_ids(user_id)` function - Get all tenant IDs for a user

**SQL Template:**
```sql
-- 092_functions_tenancy.up.sql

-- Get current tenant ID from JWT claims or session context
-- Falls back to default tenant for backward compatibility
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
DECLARE
    claims JSONB;
    tid UUID;
BEGIN
    -- Try JWT claims first
    BEGIN
        claims := auth.jwt();
        IF claims ? 'tenant_id' AND (claims->>'tenant_id') IS NOT NULL THEN
            RETURN (claims->>'tenant_id')::UUID;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
    
    -- Fall back to session variable (set by middleware)
    BEGIN
        tid := NULL;
        tid := current_setting('app.current_tenant_id', true)::UUID;
        IF tid IS NOT NULL THEN
            RETURN tid;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        NULL;
    END;
    
    -- Return default tenant for backward compatibility
    RETURN (
        SELECT id FROM tenants 
        WHERE is_default = true AND deleted_at IS NULL 
        LIMIT 1
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION current_tenant_id() IS 
'Returns the current tenant ID from JWT claims, session variable, or default tenant';

-- Check if user has specific role in tenant
CREATE OR REPLACE FUNCTION user_has_tenant_role(
    p_user_id UUID,
    p_tenant_id UUID,
    p_role TEXT
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND tm.tenant_id = p_tenant_id
        AND tm.role = p_role
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_has_tenant_role(UUID, UUID, TEXT) IS
'Checks if a user has a specific role in a tenant. SECURITY DEFINER to bypass RLS.';

-- Check if user is instance admin
CREATE OR REPLACE FUNCTION is_instance_admin(p_user_id UUID) RETURNS BOOLEAN AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1 FROM dashboard.users du
        WHERE du.id = p_user_id
        AND du.role = 'instance_admin'
        AND du.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION is_instance_admin(UUID) IS
'Checks if a user is an instance-level admin. SECURITY DEFINER to bypass RLS.';

-- Get user's effective tenant role for current tenant
CREATE OR REPLACE FUNCTION current_tenant_role() RETURNS TEXT AS $$
DECLARE
    uid UUID;
    tid UUID;
BEGIN
    uid := auth.uid();
    tid := current_tenant_id();
    
    IF uid IS NULL OR tid IS NULL THEN
        RETURN NULL;
    END IF;
    
    RETURN (
        SELECT tm.role FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = uid 
        AND tm.tenant_id = tid
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION current_tenant_role() IS
'Returns the current user role in the current tenant';

-- Get all tenant IDs for a user
CREATE OR REPLACE FUNCTION user_tenant_ids(p_user_id UUID) RETURNS UUID[] AS $$
BEGIN
    RETURN ARRAY(
        SELECT tm.tenant_id FROM tenant_memberships tm
        INNER JOIN tenants t ON t.id = tm.tenant_id
        WHERE tm.user_id = p_user_id
        AND t.deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

COMMENT ON FUNCTION user_tenant_ids(UUID) IS
'Returns all tenant IDs that a user is a member of';

-- 092_functions_tenancy.down.sql

DROP FUNCTION IF EXISTS user_tenant_ids(UUID);
DROP FUNCTION IF EXISTS current_tenant_role();
DROP FUNCTION IF EXISTS is_instance_admin(UUID);
DROP FUNCTION IF EXISTS user_has_tenant_role(UUID, UUID, TEXT);
DROP FUNCTION IF EXISTS current_tenant_id();
```

**Verification:**
```sql
SELECT current_tenant_id();
SELECT is_instance_admin('00000000-0000-0000-0000-000000000000');
```

---

### Track 1C: Add tenant_id Columns

**File:** `internal/database/migrations/099_add_tenant_columns.up.sql`

**Deliverables:**
- [ ] Add `tenant_id UUID` column to all tenant-scoped tables
- [ ] Add foreign key constraint to `tenants(id)`
- [ ] Create indexes with `tenant_id` as leading column

**Tables to Modify:**

| Schema | Table | tenant_id Column | Index |
|--------|-------|------------------|-------|
| `auth` | `users` | Yes | `idx_auth_users_tenant_id` |
| `storage` | `buckets` | Yes | `idx_storage_buckets_tenant_id` |
| `storage` | `objects` | Yes | `idx_storage_objects_tenant_id` |
| `functions` | `edge_functions` | Yes | `idx_functions_edge_functions_tenant_id` |
| `functions` | `secrets` | Yes | `idx_functions_secrets_tenant_id` |
| `jobs` | `functions` | Yes | `idx_jobs_functions_tenant_id` |
| `ai` | `knowledge_bases` | Yes | `idx_ai_knowledge_bases_tenant_id` |
| `ai` | `chatbots` | Yes | `idx_ai_chatbots_tenant_id` |
| `ai` | `conversations` | Yes | `idx_ai_conversations_tenant_id` |
| `auth` | `webhooks` | Yes | `idx_auth_webhooks_tenant_id` |
| `auth` | `api_keys` (client_keys) | Yes | `idx_auth_client_keys_tenant_id` |

**SQL Template:**
```sql
-- 093_add_tenant_columns.up.sql

-- Auth schema
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
ALTER TABLE auth.webhooks ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE auth.client_keys ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Storage schema
ALTER TABLE storage.buckets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE storage.objects ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Functions schema
ALTER TABLE functions.edge_functions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE functions.secrets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Jobs schema
ALTER TABLE jobs.functions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- AI schema
ALTER TABLE ai.knowledge_bases ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE ai.chatbots ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
ALTER TABLE ai.conversations ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;

-- Create indexes with tenant_id as leading column for RLS performance
CREATE INDEX IF NOT EXISTS idx_auth_users_tenant_id ON auth.users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_webhooks_tenant_id ON auth.webhooks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_auth_client_keys_tenant_id ON auth.client_keys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_buckets_tenant_id ON storage.buckets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_storage_objects_tenant_id ON storage.objects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_functions_edge_functions_tenant_id ON functions.edge_functions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_functions_secrets_tenant_id ON functions.secrets(tenant_id);
CREATE INDEX IF NOT EXISTS idx_jobs_functions_tenant_id ON jobs.functions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_tenant_id ON ai.knowledge_bases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_chatbots_tenant_id ON ai.chatbots(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_tenant_id ON ai.conversations(tenant_id);

-- 093_add_tenant_columns.down.sql

DROP INDEX IF EXISTS idx_ai_conversations_tenant_id;
DROP INDEX IF EXISTS idx_ai_chatbots_tenant_id;
DROP INDEX IF EXISTS idx_ai_knowledge_bases_tenant_id;
DROP INDEX IF EXISTS idx_jobs_functions_tenant_id;
DROP INDEX IF EXISTS idx_functions_secrets_tenant_id;
DROP INDEX IF EXISTS idx_functions_edge_functions_tenant_id;
DROP INDEX IF EXISTS idx_storage_objects_tenant_id;
DROP INDEX IF EXISTS idx_storage_buckets_tenant_id;
DROP INDEX IF EXISTS idx_auth_client_keys_tenant_id;
DROP INDEX IF EXISTS idx_auth_webhooks_tenant_id;
DROP INDEX IF EXISTS idx_auth_users_tenant_id;

ALTER TABLE ai.conversations DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.chatbots DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE jobs.functions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.secrets DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.edge_functions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.objects DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.buckets DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.client_keys DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.webhooks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.users DROP COLUMN IF EXISTS tenant_id;
```

**Verification:**
```sql
SELECT column_name FROM information_schema.columns 
WHERE table_name = 'users' AND table_schema = 'auth' AND column_name = 'tenant_id';
```

---

### Track 1D: Update Dashboard Role Constraint

**File:** `internal/database/migrations/100_dashboard_roles.up.sql`

**Deliverables:**
- [ ] Add `instance_admin` and `tenant_admin` roles to `dashboard.users` constraint
- [ ] Keep `dashboard_user` for backward compatibility

**SQL Template:**
```sql
-- 094_dashboard_roles.up.sql

-- Update role constraint to include new roles
ALTER TABLE dashboard.users DROP CONSTRAINT IF EXISTS dashboard_users_role_check;
ALTER TABLE dashboard.users ADD CONSTRAINT dashboard_users_role_check 
    CHECK (role IN ('instance_admin', 'tenant_admin', 'dashboard_admin', 'dashboard_user'));

COMMENT ON COLUMN dashboard.users.role IS 
'User role: instance_admin (global admin), tenant_admin (tenant-specific admin), dashboard_user (limited access)';

-- 094_dashboard_roles.down.sql

ALTER TABLE dashboard.users DROP CONSTRAINT IF EXISTS dashboard_users_role_check;
ALTER TABLE dashboard.users ADD CONSTRAINT dashboard_users_role_check 
    CHECK (role IN ('dashboard_admin', 'dashboard_user'));
```

---

## Phase 2: Backend Core

**Goal:** Implement tenant-aware authentication, middleware, and API

**Dependencies:** Phase 1 complete

**Parallel Tracks:** 5

### Track 2A: JWT Claims Extension

**Files to Modify:**
- `internal/auth/jwt.go`
- `internal/auth/service.go`

**Deliverables:**
- [ ] Extend `TokenClaims` struct with tenant fields
- [ ] Update `GenerateAccessToken` to include tenant claims
- [ ] Update `ValidateToken` to parse tenant claims
- [ ] Add `GenerateTokenWithTenant` method for tenant switching

**Code Changes:**

```go
// internal/auth/jwt.go

// Add to TokenClaims struct:
type TokenClaims struct {
    // ... existing fields ...
    
    // Multi-tenancy fields
    TenantID       *string `json:"tenant_id,omitempty"`
    TenantRole     string  `json:"tenant_role,omitempty"`
    IsInstanceAdmin bool   `json:"is_instance_admin,omitempty"`
    
    // ... rest of fields ...
}

// Add new method:
func (m *JWTManager) GenerateAccessTokenWithTenant(
    userID, email, role string,
    tenantID *string,
    tenantRole string,
    isInstanceAdmin bool,
    userMetadata, appMetadata any,
) (string, *TokenClaims, error) {
    // ... implementation
}
```

**Verification:**
```bash
go test ./internal/auth/... -run TestTokenClaims
```

---

### Track 2B: Tenant Middleware

**Files to Create/Modify:**
- `internal/middleware/tenant.go` (NEW)
- `internal/middleware/rls.go` (MODIFY)

**Deliverables:**
- [ ] `TenantMiddleware` fiber handler
- [ ] `SetTenantContext` function for PostgreSQL session
- [ ] `ValidateTenantMembership` function
- [ ] Integration with existing RLS middleware

**Code Template:**

```go
// internal/middleware/tenant.go

package middleware

import (
    "context"
    "github.com/gofiber/fiber/v3"
)

type TenantConfig struct {
    DB *database.Connection
}

// TenantMiddleware extracts tenant context from request
// Precedence: X-FB-Tenant header > JWT claim > default tenant
func TenantMiddleware(config TenantConfig) fiber.Handler {
    return func(c fiber.Ctx) error {
        // Implementation
    }
}

// ValidateTenantMembership checks if user is member of tenant
func ValidateTenantMembership(ctx context.Context, db *database.Connection, userID, tenantID string) (bool, error) {
    // Implementation
}

// SetTenantSessionContext sets PostgreSQL session variable for tenant
func SetTenantSessionContext(ctx context.Context, tx pgx.Tx, tenantID string) error {
    _, err := tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
    return err
}
```

**Verification:**
```bash
go test ./internal/middleware/... -run TestTenant
```

---

### Track 2C: Tenant API Handler

**Files to Create:**
- `internal/api/tenant_handler.go` (NEW)

**Deliverables:**
- [ ] `GET /api/v1/tenants` - List tenants (instance admin)
- [ ] `POST /api/v1/tenants` - Create tenant (instance admin)
- [ ] `GET /api/v1/tenants/:id` - Get tenant
- [ ] `PATCH /api/v1/tenants/:id` - Update tenant
- [ ] `DELETE /api/v1/tenants/:id` - Soft delete tenant
- [ ] `GET /api/v1/tenants/mine` - List user's tenants
- [ ] `POST /api/v1/tenants/:id/members` - Add member
- [ ] `GET /api/v1/tenants/:id/members` - List members
- [ ] `DELETE /api/v1/tenants/:id/members/:userId` - Remove member
- [ ] `PATCH /api/v1/tenants/:id/members/:userId` - Update role

**Code Template:**

```go
// internal/api/tenant_handler.go

package api

// RegisterTenantRoutes registers tenant management routes
func RegisterTenantRoutes(router fiber.Router, deps *Dependencies) {
    tenants := router.Group("/tenants")
    
    // Instance admin only
    tenants.Get("/", RequireRole("instance_admin"), ListTenants)
    tenants.Post("/", RequireRole("instance_admin"), CreateTenant)
    tenants.Delete("/:id", RequireRole("instance_admin"), DeleteTenant)
    
    // Tenant admin or instance admin
    tenants.Get("/mine", ListMyTenants)
    tenants.Get("/:id", GetTenant)
    tenants.Patch("/:id", RequireTenantRole("tenant_admin"), UpdateTenant)
    
    // Member management
    tenants.Get("/:id/members", ListTenantMembers)
    tenants.Post("/:id/members", RequireTenantRole("tenant_admin"), AddTenantMember)
    tenants.Patch("/:id/members/:userId", RequireTenantRole("tenant_admin"), UpdateMemberRole)
    tenants.Delete("/:id/members/:userId", RequireTenantRole("tenant_admin"), RemoveMember)
}
```

**Verification:**
```bash
go test ./internal/api/... -run TestTenant
```

---

### Track 2D: Service Client Tenant Context

**Files to Modify:**
- `internal/runtime/wrap.go`
- `internal/runtime/tokens.go`
- `internal/runtime/types.go`

**Deliverables:**
- [ ] Pass tenant context to edge function execution
- [ ] Generate service token with tenant scope option
- [ ] Add `forTenant` helper to service client wrapper

**Code Changes:**

```go
// internal/runtime/types.go

type ExecutionRequest struct {
    // ... existing fields ...
    
    // Multi-tenancy
    TenantID   string `json:"tenant_id,omitempty"`
    TenantRole string `json:"tenant_role,omitempty"`
}

// internal/runtime/wrap.go

// Add to wrapped code:
const _tenantContext = {
    id: request.tenant_id,
    role: request.tenant_role,
    setTenant: (tenantId) => {
        // Sets X-FB-Tenant header for subsequent requests
        _fluxbaseService.setTenant(tenantId);
    }
};
```

**Verification:**
```bash
go test ./internal/runtime/... -run TestTenant
```

---

### Track 2E: Update Existing API Handlers

**Files to Modify:**
- `internal/api/rest_crud.go`
- `internal/api/storage_handler.go`
- `internal/api/functions_handler.go`
- `internal/api/jobs_handler.go`

**Deliverables:**
- [ ] Add tenant filtering to REST queries
- [ ] Add tenant context to storage operations
- [ ] Add tenant context to function execution
- [ ] Add tenant context to job submission

**Pattern:**
```go
// In each handler, extract tenant context
tenantID := c.Locals("tenant_id").(string)

// Pass to database operations
query = query.Where("tenant_id = ?", tenantID)
```

---

## Phase 3: Migration Logic

**Goal:** Implement data migration and RLS policy updates

**Dependencies:** Phase 1, Phase 2A-2C

**Parallel Tracks:** 2

### Track 3A: Admin Role Migration

**File:** `internal/database/migrations/101_migrate_admin_roles.up.sql`

**Deliverables:**
- [ ] Create default tenant
- [ ] Migrate first setup admin to `instance_admin`
- [ ] Migrate other `dashboard_admin` to `tenant_admin` of default tenant
- [ ] Create tenant memberships for migrated admins

**SQL Template:**

```sql
-- 095_migrate_admin_roles.up.sql

DO $$
DECLARE
    default_tenant_id UUID;
    first_admin_id UUID;
BEGIN
    -- Step 1: Create default tenant
    INSERT INTO tenants (id, slug, name, is_default, metadata, created_at)
    VALUES (
        gen_random_uuid(),
        'default',
        'Default Tenant',
        true,
        '{"description": "Default tenant for backward compatibility"}'::jsonb,
        NOW()
    )
    RETURNING id INTO default_tenant_id;
    
    RAISE NOTICE 'Created default tenant: %', default_tenant_id;
    
    -- Step 2: Get first admin (setup admin)
    SELECT id INTO first_admin_id 
    FROM dashboard.users 
    WHERE deleted_at IS NULL
    ORDER BY created_at ASC 
    LIMIT 1;
    
    -- Step 3: Migrate first admin to instance_admin
    IF first_admin_id IS NOT NULL THEN
        UPDATE dashboard.users 
        SET role = 'instance_admin',
            updated_at = NOW()
        WHERE id = first_admin_id;
        
        RAISE NOTICE 'Migrated first admin % to instance_admin', first_admin_id;
        
        -- Create membership for instance admin in default tenant
        INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
        VALUES (default_tenant_id, first_admin_id, 'tenant_admin', NOW())
        ON CONFLICT (tenant_id, user_id) DO NOTHING;
    END IF;
    
    -- Step 4: Migrate other dashboard_admin to tenant_admin
    -- Note: These are dashboard users, not auth.users
    UPDATE dashboard.users 
    SET role = 'tenant_admin',
        updated_at = NOW()
    WHERE role = 'dashboard_admin'
    AND id != first_admin_id
    AND deleted_at IS NULL;
    
    RAISE NOTICE 'Migration complete';
END $$;

-- 095_migrate_admin_roles.down.sql

-- Revert instance_admin back to dashboard_admin
UPDATE dashboard.users SET role = 'dashboard_admin' WHERE role = 'instance_admin';

-- Revert tenant_admin back to dashboard_admin (only those we migrated)
UPDATE dashboard.users SET role = 'dashboard_admin' 
WHERE role = 'tenant_admin' 
AND id IN (
    SELECT user_id FROM tenant_memberships 
    WHERE tenant_id = (SELECT id FROM tenants WHERE is_default = true)
);

-- Remove memberships
DELETE FROM tenant_memberships WHERE tenant_id = (SELECT id FROM tenants WHERE is_default = true);

-- Remove default tenant
DELETE FROM tenants WHERE is_default = true;
```

---

### Track 3B: Data Backfill and RLS Policies

**File:** `internal/database/migrations/102_backfill_and_rls.up.sql`

**Deliverables:**
- [ ] Backfill all existing data to default tenant
- [ ] Create tenant memberships for all existing auth.users
- [ ] Update RLS policies on all tenant-scoped tables

**SQL Template:**

```sql
-- 096_backfill_and_rls.up.sql

DO $$
DECLARE
    default_tenant_id UUID;
BEGIN
    SELECT id INTO default_tenant_id FROM tenants WHERE is_default = true;
    
    -- Backfill tenant_id on all tables
    UPDATE auth.users SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhooks SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.client_keys SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE storage.buckets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE storage.objects SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.edge_functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.secrets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE jobs.functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.knowledge_bases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.chatbots SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.conversations SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    
    -- Create memberships for all existing users
    INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
    SELECT default_tenant_id, id, 'tenant_member', NOW()
    FROM auth.users
    WHERE deleted_at IS NULL
    ON CONFLICT (tenant_id, user_id) DO NOTHING;
    
    RAISE NOTICE 'Backfill complete';
END $$;

-- Update RLS policies for auth.users
DROP POLICY IF EXISTS users_service_all ON auth.users;
DROP POLICY IF EXISTS users_owner ON auth.users;

CREATE POLICY users_service_all ON auth.users
    FOR ALL TO service_role USING (true) WITH CHECK (true);

CREATE POLICY users_instance_admin ON auth.users
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

CREATE POLICY users_tenant_member ON auth.users
    FOR SELECT TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND EXISTS (
            SELECT 1 FROM tenant_memberships tm
            WHERE tm.user_id = auth.uid()
            AND tm.tenant_id = current_tenant_id()
        )
    );

CREATE POLICY users_tenant_admin_insert ON auth.users
    FOR INSERT TO authenticated
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

CREATE POLICY users_tenant_admin_update ON auth.users
    FOR UPDATE TO authenticated
    USING (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    )
    WITH CHECK (
        tenant_id = current_tenant_id()
        AND user_has_tenant_role(auth.uid(), current_tenant_id(), 'tenant_admin')
    );

-- Repeat similar pattern for other tables...
-- (Full implementation would include all tenant-scoped tables)
```

---

## Phase 4: Dashboard UI

**Goal:** Add tenant management UI and tenant-scoped views

**Dependencies:** Phase 2 complete

**Parallel Tracks:** 4

### Track 4A: Tenant Selector Component

**Files to Create:**
- `admin/src/components/tenant-selector.tsx` (NEW)
- `admin/src/hooks/use-tenant.ts` (NEW)
- `admin/src/stores/tenant-store.ts` (NEW)

**Deliverables:**
- [ ] Tenant dropdown in top navigation
- [ ] Fetch user's available tenants
- [ ] Store current tenant in Zustand + localStorage
- [ ] Send `X-FB-Tenant` header on all requests

**Component Template:**

```tsx
// admin/src/components/tenant-selector.tsx

import { useTenantStore } from '@/stores/tenant-store';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

export function TenantSelector() {
    const { tenants, currentTenant, setCurrentTenant } = useTenantStore();
    
    if (tenants.length <= 1) {
        return null; // Hide for single-tenant users
    }
    
    return (
        <Select value={currentTenant?.id} onValueChange={setCurrentTenant}>
            <SelectTrigger className="w-[200px]">
                <SelectValue placeholder="Select tenant" />
            </SelectTrigger>
            <SelectContent>
                {tenants.map(tenant => (
                    <SelectItem key={tenant.id} value={tenant.id}>
                        {tenant.name}
                    </SelectItem>
                ))}
            </SelectContent>
        </Select>
    );
}
```

---

### Track 4B: Tenant Management Pages

**Files to Create:**
- `admin/src/routes/tenants/index.tsx` (NEW)
- `admin/src/routes/tenants/$id.tsx` (NEW)
- `admin/src/routes/tenants/$id/members.tsx` (NEW)

**Deliverables:**
- [ ] Tenants list page (instance admin)
- [ ] Create tenant dialog
- [ ] Edit tenant page
- [ ] Member management page

---

### Track 4C: Update API Client

**Files to Modify:**
- `admin/src/lib/api.ts`
- `admin/src/lib/fluxbase-client.ts`

**Deliverables:**
- [ ] Add `X-FB-Tenant` header interceptor
- [ ] Update API types for tenant endpoints

---

### Track 4D: Update Navigation

**Files to Modify:**
- `admin/src/components/layout/app-sidebar.tsx`
- `admin/src/components/layout/data/sidebar-data.ts`

**Deliverables:**
- [ ] Add Tenants menu item (instance admin only)
- [ ] Add tenant selector to header

---

## Phase 5: SDK Updates

**Goal:** Add multi-tenancy support to SDKs

**Dependencies:** Phase 2 complete

**Parallel Tracks:** 3

### Track 5A: TypeScript SDK

**Files to Create/Modify:**
- `sdk/src/tenant.ts` (NEW)
- `sdk/src/client.ts` (MODIFY)
- `sdk/src/types.ts` (MODIFY)

**Deliverables:**
- [ ] `FluxbaseTenant` class with CRUD methods
- [ ] `forTenant(tenantId)` method on client
- [ ] Tenant-related types

**Code Template:**

```typescript
// sdk/src/tenant.ts

export interface Tenant {
    id: string;
    slug: string;
    name: string;
    is_default: boolean;
    metadata: Record<string, unknown>;
    created_at: string;
    updated_at: string;
}

export interface TenantMembership {
    tenant_id: string;
    user_id: string;
    role: 'tenant_admin' | 'tenant_member';
    created_at: string;
}

export class FluxbaseTenant {
    constructor(private fetch: FluxbaseFetch) {}
    
    async list(): Promise<{ data: Tenant[] | null; error: Error | null }> {
        return this.fetch.get('/tenants');
    }
    
    async get(id: string): Promise<{ data: Tenant | null; error: Error | null }> {
        return this.fetch.get(`/tenants/${id}`);
    }
    
    async create(data: CreateTenantInput): Promise<{ data: Tenant | null; error: Error | null }> {
        return this.fetch.post('/tenants', data);
    }
    
    async listMine(): Promise<{ data: Tenant[] | null; error: Error | null }> {
        return this.fetch.get('/tenants/mine');
    }
    
    async listMembers(tenantId: string): Promise<{ data: TenantMembership[] | null; error: Error | null }> {
        return this.fetch.get(`/tenants/${tenantId}/members`);
    }
    
    async addMember(tenantId: string, userId: string, role: string): Promise<{ ... }> {
        return this.fetch.post(`/tenants/${tenantId}/members`, { user_id: userId, role });
    }
}

// Add to FluxbaseClient:
export class FluxbaseClient {
    public tenant: FluxbaseTenant;
    
    private tenantId?: string;
    
    forTenant(tenantId: string): FluxbaseClient {
        const options = {
            ...this.options,
            headers: { ...this.options.headers, 'X-FB-Tenant': tenantId }
        };
        return new FluxbaseClient(this.url, this.key, options);
    }
}
```

---

### Track 5B: Go SDK

**Files to Create/Modify:**
- `pkg/client/tenant.go` (NEW)
- `pkg/client/client.go` (MODIFY)

**Deliverables:**
- [ ] `TenantService` with CRUD methods
- [ ] `ForTenant(tenantID string)` method on client
- [ ] Tenant-related types

---

### Track 5C: React SDK

**Files to Create:**
- `sdk-react/src/hooks/use-tenant.ts` (NEW)
- `sdk-react/src/context/tenant-context.tsx` (NEW)

**Deliverables:**
- [ ] `useTenant()` hook
- [ ] `useTenants()` hook
- [ ] `TenantProvider` context

---

## Phase 6: Testing & Documentation

**Goal:** Ensure quality and document features

**Dependencies:** All previous phases

**Parallel Tracks:** 3

### Track 6A: Backend Tests

**Files to Create:**
- `internal/middleware/tenant_test.go` (NEW)
- `internal/api/tenant_handler_test.go` (NEW)
- `test/e2e/tenant_isolation_test.go` (NEW)

**Deliverables:**
- [ ] Tenant middleware unit tests
- [ ] Tenant API handler tests
- [ ] E2E tenant isolation tests
- [ ] Backward compatibility tests

**Test Cases:**
```go
func TestTenantIsolation_Storage(t *testing.T) {
    // Create tenant A and tenant B
    // Create objects in each tenant
    // Verify tenant A user cannot see tenant B objects
    // Verify instance admin can see all
}

func TestBackwardCompat_NoTenantHeader(t *testing.T) {
    // Make request without X-FB-Tenant header
    // Verify it uses default tenant
    // Verify data is accessible
}

func TestTenantMember_RolePermissions(t *testing.T) {
    // Verify tenant_member cannot add members
    // Verify tenant_admin can add members
    // Verify instance_admin can do anything
}
```

---

### Track 6B: Frontend Tests

**Files to Create:**
- `admin/src/components/tenant-selector.test.tsx` (NEW)
- `admin/src/hooks/use-tenant.test.ts` (NEW)

**Deliverables:**
- [ ] Tenant selector component tests
- [ ] Tenant hook tests
- [ ] API client header tests

---

### Track 6C: Documentation

**Files to Create/Modify:**
- `docs/src/content/docs/guides/multi-tenancy.md` (NEW)
- `docs/src/content/docs/supabase-comparison.md` (MODIFY)
- `docs/src/content/docs/intro.md` (MODIFY)

**Deliverables:**
- [ ] Complete multi-tenancy guide
- [ ] Updated Supabase comparison with multi-tenancy row
- [ ] SDK documentation for tenant methods
- [ ] Migration guide for existing deployments

---

## Verification Checklist

### Phase 1 Complete
- [ ] `SELECT * FROM tenants` works
- [ ] `SELECT * FROM tenant_memberships` works
- [ ] `SELECT current_tenant_id()` returns valid UUID
- [ ] `SELECT is_instance_admin(...)` works
- [ ] All tenant_id columns exist

### Phase 2 Complete
- [ ] JWT contains tenant_id claim
- [ ] X-FB-Tenant header is respected
- [ ] Tenant API endpoints respond
- [ ] Service client can set tenant context

### Phase 3 Complete
- [ ] Default tenant exists
- [ ] All data has tenant_id set
- [ ] All users have tenant memberships
- [ ] RLS policies block cross-tenant access

### Phase 4 Complete
- [ ] Tenant selector appears in dashboard
- [ ] Tenant switching works
- [ ] Data is scoped to selected tenant

### Phase 5 Complete
- [ ] `client.forTenant()` works in TS SDK
- [ ] `client.ForTenant()` works in Go SDK
- [ ] React hooks work

### Phase 6 Complete
- [ ] All tests pass
- [ ] Documentation is complete
- [ ] Supabase comparison updated

---

## Rollback Plan

If issues arise, migrations can be rolled back in reverse order:

1. `102_backfill_and_rls.down.sql`
2. `101_migrate_admin_roles.down.sql`
3. `100_dashboard_roles.down.sql`
4. `099_add_tenant_columns.down.sql`
5. `098_functions_tenancy.down.sql`
6. `097_tenants.down.sql`

**Important:** Rollback will lose all tenant associations. Only rollback on fresh installs or after backing up data.

---

## Notes for Agents

### Working on Phase 1
- All 4 tracks are independent and can run in parallel
- Each track produces a single migration file
- Test migrations by running `make migrate-up` and `make migrate-down`

### Working on Phase 2
- Track 2A should complete first (JWT changes are foundational)
- Tracks 2B-2E can run in parallel after 2A
- Coordinate API route registration in Track 2C

### Working on Phase 3
- Track 3A must complete before 3B
- Test migration on a copy of production data first
- Verify rollback works before deploying

### Working on Phase 4
- All tracks can run in parallel
- Use mock API responses until Phase 2 is complete
- Test with single-tenant and multi-tenant scenarios

### Working on Phase 5
- All tracks can run in parallel
- TypeScript SDK is highest priority
- Ensure backward compatibility (no breaking changes)

### Working on Phase 6
- Tests can be written in parallel with implementation
- Documentation should reference actual implemented behavior
- Run full test suite before marking complete
