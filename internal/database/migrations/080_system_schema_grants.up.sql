-- Grant service_role access to system schema
-- This allows dashboard_admin (which maps to service_role) to view system.rate_limits

GRANT USAGE ON SCHEMA system TO service_role;
GRANT ALL ON ALL TABLES IN SCHEMA system TO service_role;
GRANT ALL ON ALL SEQUENCES IN SCHEMA system TO service_role;

-- Set default privileges for future tables in system schema
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    GRANT ALL ON SEQUENCES TO service_role;
