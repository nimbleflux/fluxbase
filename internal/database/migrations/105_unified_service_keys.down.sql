--
-- UNIFIED SERVICE KEYS (ROLLBACK)
-- Drops platform.service_keys and platform.key_usage tables
-- Note: Data is NOT migrated back to auth.client_keys and auth.service_keys
-- This is a destructive rollback - data in platform.service_keys will be lost
--

-- Drop trigger and function
DROP TRIGGER IF EXISTS platform_service_keys_updated_at ON platform.service_keys;
DROP FUNCTION IF EXISTS update_platform_service_keys_updated_at();

-- Drop key_usage table first (has FK to service_keys)
DROP TABLE IF EXISTS platform.key_usage;

-- Drop service_keys table
DROP TABLE IF EXISTS platform.service_keys;

-- Drop platform.tenants table (was created by this migration)
-- Only drop if it was created by this migration (check if it's a copy of public.tenants)
-- For safety, we keep platform.tenants as it may be used by other tables
-- Uncomment the following line if you want to drop it:
-- DROP TABLE IF EXISTS platform.tenants;

-- Note: We do NOT drop the platform schema itself as it may contain other objects
-- The old tables (auth.client_keys, auth.service_keys, auth.client_key_usage) 
-- still exist with their original data, so this rollback is safe from a data loss perspective.
-- Any new keys created in platform.service_keys after migration will be lost.
