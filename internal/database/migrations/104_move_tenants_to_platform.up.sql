--
-- MULTI-TENANCY: MOVE TENANTS TO PLATFORM SCHEMA
-- This migration is now a no-op because:
-- 1. Migration 105 already creates platform.tenants and platform.tenant_memberships
-- 2. The tables are created in the correct schema from the start
--
-- This migration is kept for documentation purposes and to maintain
-- migration version continuity.

-- No-op: Tables are already in platform schema
SELECT 1 AS migration_placeholder;
