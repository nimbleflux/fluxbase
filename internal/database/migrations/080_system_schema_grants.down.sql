-- Revoke service_role access from system schema
REVOKE ALL ON ALL TABLES IN SCHEMA system FROM service_role;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA system FROM service_role;
REVOKE USAGE ON SCHEMA system FROM service_role;

-- Remove default privileges
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    REVOKE ALL ON TABLES FROM service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    REVOKE ALL ON SEQUENCES FROM service_role;
