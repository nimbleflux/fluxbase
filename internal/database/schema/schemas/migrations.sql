--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO migrations, public;


--
-- Name: app; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS app (
    id uuid DEFAULT gen_random_uuid(),
    namespace text DEFAULT 'default' NOT NULL,
    name text NOT NULL,
    description text,
    up_sql text NOT NULL,
    down_sql text,
    version integer DEFAULT 1,
    status text DEFAULT 'pending',
    created_by uuid,
    applied_by uuid,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    applied_at timestamptz,
    rolled_back_at timestamptz,
    CONSTRAINT app_pkey PRIMARY KEY (id),
    CONSTRAINT unique_migration_namespace UNIQUE (namespace, name),
    CONSTRAINT valid_status CHECK (status IN ('pending'::text, 'applied'::text, 'failed'::text, 'rolled_back'::text))
);


COMMENT ON TABLE app IS 'All user-facing migrations (filesystem and API-managed)';


COMMENT ON COLUMN migrations.app.namespace IS 'Namespace for isolation: filesystem for local files, or custom (default, staging, prod, etc.) for API';


COMMENT ON COLUMN migrations.app.name IS 'Migration name, should follow convention like 001_description for ordering';


COMMENT ON COLUMN migrations.app.up_sql IS 'SQL to apply the migration';


COMMENT ON COLUMN migrations.app.down_sql IS 'SQL to rollback the migration (optional)';


COMMENT ON COLUMN migrations.app.status IS 'Current status: pending (not applied), applied (successful), failed (error), rolled_back';

--
-- Name: idx_migrations_app_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_migrations_app_namespace ON app (namespace);

--
-- Name: idx_migrations_app_namespace_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_migrations_app_namespace_status ON app (namespace, status);

--
-- Name: idx_migrations_app_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_migrations_app_status ON app (status);

--
-- Name: execution_logs; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS execution_logs (
    id uuid DEFAULT gen_random_uuid(),
    migration_id uuid NOT NULL,
    action text NOT NULL,
    status text NOT NULL,
    duration_ms integer,
    error_message text,
    logs text,
    executed_at timestamptz DEFAULT now() NOT NULL,
    executed_by uuid,
    CONSTRAINT execution_logs_pkey PRIMARY KEY (id),
    CONSTRAINT execution_logs_migration_id_fkey FOREIGN KEY (migration_id) REFERENCES app (id) ON DELETE CASCADE,
    CONSTRAINT valid_action CHECK (action IN ('apply'::text, 'rollback'::text)),
    CONSTRAINT valid_execution_status CHECK (status IN ('success'::text, 'failed'::text))
);


COMMENT ON TABLE execution_logs IS 'Audit log of all migration apply/rollback attempts';

--
-- Name: idx_execution_logs_executed_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_execution_logs_executed_at ON execution_logs (executed_at DESC);

--
-- Name: idx_execution_logs_migration; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_execution_logs_migration ON execution_logs (migration_id);

--
-- Name: fluxbase; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS fluxbase (
    version bigint,
    dirty boolean NOT NULL,
    CONSTRAINT fluxbase_pkey PRIMARY KEY (version)
);


COMMENT ON TABLE fluxbase IS 'Tracks Fluxbase system migration versions (managed by golang-migrate)';

--
-- Cross-schema FKs moved to post-schema-fks.sql
-- app_applied_by_fkey, app_created_by_fkey, execution_logs_executed_by_fkey
--

--
-- Name: declarative_state; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS declarative_state (
    id SERIAL PRIMARY KEY,
    schema_fingerprint TEXT NOT NULL,
    applied_at TIMESTAMPTZ DEFAULT NOW(),
    applied_by TEXT,
    source TEXT CHECK (source IN ('fresh_install', 'transitioned', 'schema_apply'))
);


COMMENT ON TABLE declarative_state IS 'Tracks declarative schema application state with fingerprints';


COMMENT ON COLUMN migrations.declarative_state.source IS 'Source of schema application: fresh_install (new DB), transitioned (migrated from imperative), schema_apply (normal startup)';

--
-- Name: bootstrap_state; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS bootstrap_state (
    id SERIAL PRIMARY KEY,
    bootstrapped_at TIMESTAMPTZ DEFAULT NOW(),
    version TEXT NOT NULL,
    checksum TEXT
);


COMMENT ON TABLE bootstrap_state IS 'Tracks bootstrap completion state';

--
-- Name: app; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE app TO {{APP_USER}};

--
-- Name: app; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: app; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE app TO service_role;

--
-- Name: execution_logs; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE execution_logs TO {{APP_USER}};

--
-- Name: execution_logs; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: execution_logs; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE execution_logs TO service_role;

--
-- Name: fluxbase; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE fluxbase TO {{APP_USER}};

--
-- Name: fluxbase; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: fluxbase; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE fluxbase TO service_role;

--
-- Name: declarative_state; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT, INSERT, UPDATE ON TABLE declarative_state TO service_role;

--
-- Name: declarative_state_id_seq; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT USAGE, SELECT ON SEQUENCE declarative_state_id_seq TO service_role;

--
-- Name: bootstrap_state; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT, INSERT ON TABLE bootstrap_state TO service_role;
