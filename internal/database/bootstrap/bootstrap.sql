-- Fluxbase Bootstrap SQL
-- This file creates the minimal infrastructure needed before declarative schema management.
-- It handles operations that pgschema cannot manage:
-- - CREATE EXTENSION (database-level)
-- - CREATE SCHEMA (database-level)
-- - CREATE ROLE (cluster-level)
-- - ALTER DEFAULT PRIVILEGES (database-level)
--
-- This file is idempotent and safe to run multiple times.

-- ============================================================================
-- EXTENSIONS
-- ============================================================================

-- UUID generation functions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";    -- Provides uuid_generate_v4()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";     -- Provides gen_random_uuid() and crypto functions

-- Text search and indexing
CREATE EXTENSION IF NOT EXISTS "pg_trgm";      -- Trigram text search
CREATE EXTENSION IF NOT EXISTS "btree_gin";    -- GIN indexes for btree-indexable data types

-- Vector similarity search (pgvector)
CREATE EXTENSION IF NOT EXISTS "vector";       -- Vector embeddings for AI/ML

-- Foreign data wrapper for cross-database joins (tenant → main DB)
CREATE EXTENSION IF NOT EXISTS "postgres_fdw";

-- ============================================================================
-- SCHEMAS
-- ============================================================================

-- Auth schema: Application user authentication, client keys, and sessions
CREATE SCHEMA IF NOT EXISTS auth;
GRANT USAGE, CREATE ON SCHEMA auth TO CURRENT_USER;
COMMENT ON SCHEMA auth IS 'Application user authentication, client keys, and sessions';

-- App schema: Application-level configuration and settings
CREATE SCHEMA IF NOT EXISTS app;
GRANT USAGE, CREATE ON SCHEMA app TO CURRENT_USER;
COMMENT ON SCHEMA app IS 'Application-level configuration, settings, and metadata';

-- Dashboard schema: Platform administrator authentication and management
CREATE SCHEMA IF NOT EXISTS dashboard;
GRANT USAGE, CREATE ON SCHEMA dashboard TO CURRENT_USER;
COMMENT ON SCHEMA dashboard IS 'Platform administrator authentication and management';

-- Functions schema: Edge functions and their executions
CREATE SCHEMA IF NOT EXISTS functions;
GRANT USAGE, CREATE ON SCHEMA functions TO CURRENT_USER;
COMMENT ON SCHEMA functions IS 'Edge functions and their executions';

-- Storage schema: File storage buckets and objects
CREATE SCHEMA IF NOT EXISTS storage;
GRANT USAGE, CREATE ON SCHEMA storage TO CURRENT_USER;
COMMENT ON SCHEMA storage IS 'File storage buckets and objects';

-- Realtime schema: Realtime subscriptions and change tracking
CREATE SCHEMA IF NOT EXISTS realtime;
GRANT USAGE, CREATE ON SCHEMA realtime TO CURRENT_USER;
COMMENT ON SCHEMA realtime IS 'Realtime subscriptions and change tracking';

-- Migrations schema: All migration tracking
CREATE SCHEMA IF NOT EXISTS migrations;
GRANT USAGE, CREATE ON SCHEMA migrations TO CURRENT_USER;
COMMENT ON SCHEMA migrations IS 'All migration tracking including system, user, and API-managed migrations';

-- Jobs schema: Long-running background jobs
CREATE SCHEMA IF NOT EXISTS jobs;
GRANT USAGE, CREATE ON SCHEMA jobs TO CURRENT_USER;
COMMENT ON SCHEMA jobs IS 'Long-running background jobs system';

-- AI schema: AI chatbots, conversations, and query auditing
CREATE SCHEMA IF NOT EXISTS ai;
GRANT USAGE, CREATE ON SCHEMA ai TO CURRENT_USER;
COMMENT ON SCHEMA ai IS 'AI chatbots, conversations, and query auditing';

-- RPC schema: Stored procedure definitions and executions
CREATE SCHEMA IF NOT EXISTS rpc;
GRANT USAGE, CREATE ON SCHEMA rpc TO CURRENT_USER;
COMMENT ON SCHEMA rpc IS 'Stored procedure definitions and executions';

-- System schema: System-level infrastructure for scaling and distributed operations
CREATE SCHEMA IF NOT EXISTS system;
GRANT USAGE, CREATE ON SCHEMA system TO CURRENT_USER;
COMMENT ON SCHEMA system IS 'System-level infrastructure for scaling and distributed operations';

-- Logging schema: Centralized logging entries
CREATE SCHEMA IF NOT EXISTS logging;
GRANT USAGE, CREATE ON SCHEMA logging TO CURRENT_USER;
COMMENT ON SCHEMA logging IS 'Centralized logging entries';

-- Branching schema: Database branching for dev/test environments
CREATE SCHEMA IF NOT EXISTS branching;
GRANT USAGE, CREATE ON SCHEMA branching TO CURRENT_USER;
COMMENT ON SCHEMA branching IS 'Database branching for dev/test environments';

-- MCP schema: Model Context Protocol server
CREATE SCHEMA IF NOT EXISTS mcp;
GRANT USAGE, CREATE ON SCHEMA mcp TO CURRENT_USER;
COMMENT ON SCHEMA mcp IS 'Model Context Protocol server for AI assistant integration';

-- API schema: API infrastructure (idempotency keys, etc.)
CREATE SCHEMA IF NOT EXISTS api;
GRANT USAGE, CREATE ON SCHEMA api TO CURRENT_USER;
COMMENT ON SCHEMA api IS 'API infrastructure including idempotency keys';

-- Platform schema: Multi-tenancy control plane
CREATE SCHEMA IF NOT EXISTS platform;
GRANT USAGE, CREATE ON SCHEMA platform TO CURRENT_USER;
COMMENT ON SCHEMA platform IS 'Multi-tenancy control plane (tenants, service keys, admin assignments)';

-- ============================================================================
-- ROLES
-- ============================================================================

-- Anonymous role: For unauthenticated requests with public access only
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'anon') THEN
        CREATE ROLE anon NOLOGIN NOINHERIT;
    END IF;
END
$$;
COMMENT ON ROLE anon IS 'Anonymous role for unauthenticated requests with public access only';

-- Authenticated role: For users with valid JWT tokens
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'authenticated') THEN
        CREATE ROLE authenticated NOLOGIN NOINHERIT;
    END IF;
END
$$;
COMMENT ON ROLE authenticated IS 'Authenticated role for users with valid JWT tokens';

-- Service role: For backend services with elevated permissions and RLS bypass
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'service_role') THEN
        CREATE ROLE service_role NOLOGIN NOINHERIT BYPASSRLS;
    END IF;
END
$$;
COMMENT ON ROLE service_role IS 'Service role for backend services with elevated permissions and RLS bypass';

-- Tenant service role: Tenant-scoped service role for multi-tenant isolation
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_service') THEN
        CREATE ROLE tenant_service NOLOGIN NOINHERIT NOBYPASSRLS;
    END IF;
END
$$;
COMMENT ON ROLE tenant_service IS 'Tenant-scoped service role for multi-tenant isolation - respects RLS with tenant context';

-- Tenant migration role: For tenant-specific schema migrations
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'tenant_migration_role') THEN
        CREATE ROLE tenant_migration_role NOLOGIN NOINHERIT NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
    END IF;
END
$$;
COMMENT ON ROLE tenant_migration_role IS 'Role for tenant-specific schema migrations with limited permissions';

-- ============================================================================
-- ROLE GRANTS TO CURRENT USER
-- These allow the RLS middleware to execute SET LOCAL ROLE
-- ============================================================================

GRANT anon TO CURRENT_USER;
GRANT authenticated TO CURRENT_USER;
GRANT service_role TO CURRENT_USER;
GRANT tenant_service TO CURRENT_USER;
GRANT tenant_migration_role TO CURRENT_USER;

-- ============================================================================
-- SCHEMA USAGE GRANTS FOR RLS ROLES
-- ============================================================================

-- Core schemas accessible to all RLS roles
GRANT USAGE ON SCHEMA auth TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA app TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA storage TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA functions TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA realtime TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA dashboard TO anon, authenticated, service_role;
GRANT USAGE ON SCHEMA public TO anon, authenticated, service_role;

-- AI schema accessible to all
GRANT USAGE ON SCHEMA ai TO anon, authenticated, service_role;

-- Jobs schema - authenticated and service only
GRANT USAGE ON SCHEMA jobs TO authenticated, service_role;

-- RPC schema - authenticated and service only
GRANT USAGE ON SCHEMA rpc TO authenticated, service_role;

-- MCP schema - all roles
GRANT USAGE ON SCHEMA mcp TO anon, authenticated, service_role;

-- System schema - service role only
GRANT USAGE ON SCHEMA system TO service_role;

-- API schema - service role only
GRANT USAGE ON SCHEMA api TO service_role;

-- Branching schema - service role only
GRANT USAGE ON SCHEMA branching TO service_role;

-- Logging schema - service role and fluxbase_app (app needs access for direct pool queries)
GRANT USAGE ON SCHEMA logging TO service_role, fluxbase_app;

-- Migrations schema - service role only (for recording schema state)
GRANT USAGE ON SCHEMA migrations TO service_role;

-- Platform schema - authenticated and service
GRANT USAGE ON SCHEMA platform TO authenticated, service_role;

-- Tenant migration role - public schema only
GRANT USAGE, CREATE ON SCHEMA public TO tenant_migration_role;

-- ============================================================================
-- ALTER DEFAULT PRIVILEGES
-- These ensure future tables automatically get correct permissions.
--
-- Note: Default privileges are only set for `authenticated` and `service_role`.
-- Roles `fluxbase_app` and `fluxbase_rls_test` receive per-table GRANTs in the
-- declarative schema SQL files (internal/database/schema/schemas/*.sql) rather
-- than default privileges, so pgschema can track them correctly.
-- ============================================================================

-- Auth schema
ALTER DEFAULT PRIVILEGES IN SCHEMA auth
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA auth
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA auth
    GRANT ALL ON SEQUENCES TO service_role;

-- App schema
ALTER DEFAULT PRIVILEGES IN SCHEMA app
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA app
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA app
    GRANT ALL ON SEQUENCES TO service_role;

-- Storage schema
ALTER DEFAULT PRIVILEGES IN SCHEMA storage
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA storage
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA storage
    GRANT ALL ON SEQUENCES TO service_role;

-- Functions schema
ALTER DEFAULT PRIVILEGES IN SCHEMA functions
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA functions
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA functions
    GRANT ALL ON SEQUENCES TO service_role;

-- Realtime schema
ALTER DEFAULT PRIVILEGES IN SCHEMA realtime
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA realtime
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA realtime
    GRANT ALL ON SEQUENCES TO service_role;

-- Dashboard schema
ALTER DEFAULT PRIVILEGES IN SCHEMA dashboard
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA dashboard
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA dashboard
    GRANT ALL ON SEQUENCES TO service_role;

-- Jobs schema
ALTER DEFAULT PRIVILEGES IN SCHEMA jobs
    GRANT SELECT ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA jobs
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA jobs
    GRANT ALL ON SEQUENCES TO service_role;

-- AI schema
ALTER DEFAULT PRIVILEGES IN SCHEMA ai
    GRANT SELECT ON TABLES TO authenticated;
ALTER DEFAULT PRIVILEGES IN SCHEMA ai
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA ai
    GRANT ALL ON SEQUENCES TO service_role;

-- RPC schema
ALTER DEFAULT PRIVILEGES IN SCHEMA rpc
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA rpc
    GRANT ALL ON SEQUENCES TO service_role;

-- System schema
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA system
    GRANT ALL ON SEQUENCES TO service_role;

-- MCP schema
ALTER DEFAULT PRIVILEGES IN SCHEMA mcp
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA mcp
    GRANT ALL ON SEQUENCES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA mcp
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO authenticated;

-- API schema
ALTER DEFAULT PRIVILEGES IN SCHEMA api
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA api
    GRANT ALL ON SEQUENCES TO service_role;

-- Branching schema
ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA branching
    GRANT ALL ON SEQUENCES TO service_role;

-- Public schema - service_role only
ALTER DEFAULT PRIVILEGES FOR ROLE CURRENT_USER IN SCHEMA public
    GRANT ALL ON TABLES TO service_role;
ALTER DEFAULT PRIVILEGES FOR ROLE CURRENT_USER IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO service_role;

-- Tenant migration role - public schema
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON TABLES TO tenant_migration_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL ON SEQUENCES TO tenant_migration_role;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT EXECUTE ON FUNCTIONS TO tenant_migration_role;

-- ============================================================================
-- MIGRATIONS STATE TABLES
-- ============================================================================

-- Declarative schema state tracking
CREATE TABLE IF NOT EXISTS migrations.declarative_state (
    id SERIAL PRIMARY KEY,
    schema_fingerprint TEXT NOT NULL,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    applied_by TEXT,
    source TEXT CHECK (source IN ('fresh_install', 'transitioned', 'schema_apply'))
);

-- Bootstrap state tracking
CREATE TABLE IF NOT EXISTS migrations.bootstrap_state (
    id SERIAL PRIMARY KEY,
    bootstrapped_at TIMESTAMPTZ DEFAULT NOW(),
    version TEXT NOT NULL,
    checksum TEXT
);

-- Grant permissions on migrations tables to service_role
GRANT SELECT, INSERT, UPDATE ON TABLE migrations.declarative_state TO service_role;
GRANT USAGE, SELECT ON SEQUENCE migrations.declarative_state_id_seq TO service_role;
GRANT SELECT, INSERT ON TABLE migrations.bootstrap_state TO service_role;

-- Record bootstrap completion (idempotent)
INSERT INTO migrations.bootstrap_state (version, checksum)
VALUES ('2.0.0', '')
ON CONFLICT DO NOTHING;
