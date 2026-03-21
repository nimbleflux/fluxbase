-- Migration 109 Rollback: Remove tenant_id from auth tables

-- ============================================================================
-- STEP 1: Remove foreign key constraints
-- ============================================================================

ALTER TABLE auth.users DROP CONSTRAINT IF EXISTS fk_auth_users_tenant;
ALTER TABLE auth.service_keys DROP CONSTRAINT IF EXISTS fk_auth_service_keys_tenant;

-- ============================================================================
-- STEP 2: Remove RLS policies
-- ============================================================================

DROP POLICY IF EXISTS auth_users_select ON auth.users;
DROP POLICY IF EXISTS auth_users_insert ON auth.users;
DROP POLICY IF EXISTS auth_users_update ON auth.users;
DROP POLICY IF EXISTS auth_users_delete ON auth.users;
DROP POLICY IF EXISTS auth_service_keys_select ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_insert ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_update ON auth.service_keys;
DROP POLICY IF EXISTS auth_service_keys_delete ON auth.service_keys;

-- Disable RLS (optional - may want to keep enabled)
-- ALTER TABLE auth.users DISABLE ROW LEVEL SECURITY;
-- ALTER TABLE auth.service_keys DISABLE ROW LEVEL SECURITY;

-- ============================================================================
-- STEP 3: Remove triggers
-- ============================================================================

DROP TRIGGER IF EXISTS auth_users_set_tenant_id ON auth.users;
DROP TRIGGER IF EXISTS auth_service_keys_set_tenant_id ON auth.service_keys;
DROP FUNCTION IF EXISTS auth.set_tenant_id_from_context();

-- ============================================================================
-- STEP 4: Remove partial unique indexes
-- ============================================================================

DROP INDEX IF EXISTS auth_users_email_tenant_null_unique;
DROP INDEX IF EXISTS auth_users_email_tenant_unique;
DROP INDEX IF EXISTS auth_service_keys_name_tenant_null_unique;
DROP INDEX IF EXISTS auth_service_keys_name_tenant_unique;

-- ============================================================================
-- STEP 5: Remove lookup indexes
-- ============================================================================

DROP INDEX IF EXISTS idx_auth_users_tenant_id;
DROP INDEX IF EXISTS idx_auth_service_keys_tenant_id;

-- ============================================================================
-- STEP 6: Remove tenant_id columns
-- ============================================================================

ALTER TABLE auth.users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.service_keys DROP COLUMN IF EXISTS tenant_id;

-- ============================================================================
-- STEP 7: Revoke permissions from tenant_service
-- ============================================================================

REVOKE SELECT, INSERT, UPDATE, DELETE ON auth.users FROM tenant_service;
REVOKE SELECT, INSERT, UPDATE, DELETE ON auth.service_keys FROM tenant_service;
REVOKE USAGE, SELECT ON SEQUENCE auth.users_id_seq FROM tenant_service;
REVOKE USAGE, SELECT ON SEQUENCE auth.service_keys_id_seq FROM tenant_service;

-- ============================================================================
-- STEP 8: Restore original unique constraint on email
-- ============================================================================

CREATE UNIQUE INDEX IF NOT EXISTS auth_users_email_key ON auth.users(email);

-- ============================================================================
-- STEP 9: Keep auth.has_tenant_access function (used by other tables)
-- ============================================================================

-- Don't drop the function as it may be used by storage and other schemas
-- DROP FUNCTION IF EXISTS auth.has_tenant_access(UUID);
