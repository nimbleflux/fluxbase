-- Migrate execution logs to centralized logging system
-- This migration marks execution_logs tables as deprecated and creates
-- a migration status view. TimescaleDB support is handled at runtime.

BEGIN;

-- Mark execution_logs tables as deprecated (but don't drop yet for safety)
-- Only add comments if tables exist
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'functions' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE functions.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'jobs' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE jobs.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'rpc' AND table_name = 'execution_logs') THEN
        EXECUTE 'COMMENT ON TABLE rpc.execution_logs IS ''DEPRECATED: Migrate data to logging.entries using centralized system''';
    END IF;
END $$;

-- Create a view to help identify which execution_logs still need migration
CREATE OR REPLACE VIEW logging.execution_logs_migration_status AS
SELECT
    table_schema,
    table_name,
    CASE
        WHEN table_schema = 'functions' AND table_name = 'execution_logs' THEN 'functions.edge_functions'
        WHEN table_schema = 'jobs' AND table_name = 'execution_logs' THEN 'jobs.functions'
        WHEN table_schema = 'rpc' AND table_name = 'execution_logs' THEN 'rpc.procedures'
        WHEN table_schema = 'branching' AND table_name = 'seed_execution_log' THEN 'branching'
        ELSE table_schema || '.' || table_name
    END AS source,
    CASE
        WHEN table_schema IN ('functions', 'jobs', 'rpc') AND table_name = 'execution_logs' THEN 'MIGRATE TO LOGGING'
        WHEN table_schema = 'branching' AND table_name = 'seed_execution_log' THEN 'MIGRATE TO LOGGING'
        ELSE 'NOT APPLICABLE'
    END AS needs_migration
FROM information_schema.tables
WHERE (table_schema, table_name) IN (
    ('functions', 'execution_logs'),
    ('jobs', 'execution_logs'),
    ('rpc', 'execution_logs'),
    ('branching', 'seed_execution_log')
);

COMMIT;
