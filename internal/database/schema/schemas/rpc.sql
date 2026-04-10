--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO rpc, public;


--
-- Name: procedures; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS procedures (
    id uuid DEFAULT gen_random_uuid(),
    name text NOT NULL,
    namespace text DEFAULT 'default' NOT NULL,
    description text,
    sql_query text NOT NULL,
    original_code text,
    input_schema jsonb,
    output_schema jsonb,
    allowed_tables text[] DEFAULT ARRAY[]::text[],
    allowed_schemas text[] DEFAULT ARRAY['public'],
    max_execution_time_seconds integer DEFAULT 30,
    is_public boolean DEFAULT false,
    schedule text,
    enabled boolean DEFAULT true,
    version integer DEFAULT 1,
    source text DEFAULT 'filesystem' NOT NULL,
    created_by uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    disable_execution_logs boolean DEFAULT false NOT NULL,
    require_roles text[] DEFAULT ARRAY[]::text[],
    tenant_id uuid,
    CONSTRAINT procedures_pkey PRIMARY KEY (id),
    CONSTRAINT unique_rpc_procedure_name_namespace UNIQUE (name, namespace),
    CONSTRAINT procedures_source_check CHECK (source IN ('filesystem'::text, 'api'::text, 'sdk'::text))
);


COMMENT ON TABLE procedures IS 'RPC procedure definitions with SQL queries and configuration';


COMMENT ON COLUMN rpc.procedures.sql_query IS 'The SQL query to execute (with $param_name placeholders)';


COMMENT ON COLUMN rpc.procedures.input_schema IS 'JSON Schema for input validation (null for schemaless)';


COMMENT ON COLUMN rpc.procedures.output_schema IS 'JSON Schema for output validation (null for schemaless)';


COMMENT ON COLUMN rpc.procedures.allowed_tables IS 'Tables the procedure can access (from @fluxbase:allowed-tables annotation)';


COMMENT ON COLUMN rpc.procedures.schedule IS 'Cron expression for scheduled execution (e.g., "0 */5 * * * *" for every 5 minutes)';


COMMENT ON COLUMN rpc.procedures.disable_execution_logs IS 'When true, execution logs are not created for this procedure (from @fluxbase:disable-execution-logs annotation)';


COMMENT ON COLUMN rpc.procedures.require_roles IS 'Roles required to invoke (authenticated, admin, anon, or custom roles). User needs ANY of the specified roles.';

--
-- Name: idx_rpc_procedures_enabled; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_enabled ON procedures (enabled);

--
-- Name: idx_rpc_procedures_is_public; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_is_public ON procedures (is_public);

--
-- Name: idx_rpc_procedures_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_name ON procedures (name);

--
-- Name: idx_rpc_procedures_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_namespace ON procedures (namespace);

--
-- Name: idx_rpc_procedures_schedule; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_schedule ON procedures (schedule) WHERE (schedule IS NOT NULL);

--
-- Name: idx_rpc_procedures_source; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_source ON procedures (source);

--
-- Name: procedures; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE procedures ENABLE ROW LEVEL SECURITY;

--
-- Name: rpc_procedures_instance_admin_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY rpc_procedures_instance_admin_read ON procedures FOR SELECT TO authenticated USING (auth.role() = 'instance_admin');

--
-- Name: rpc_procedures_read_anon; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY rpc_procedures_read_anon ON procedures FOR SELECT TO anon USING ((enabled = true) AND (is_public = true));

--
-- Name: rpc_procedures_read_public; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY rpc_procedures_read_public ON procedures FOR SELECT TO authenticated USING ((enabled = true) AND (is_public = true));

--
-- Name: executions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS executions (
    id uuid DEFAULT gen_random_uuid(),
    procedure_id uuid,
    procedure_name text NOT NULL,
    namespace text DEFAULT 'default' NOT NULL,
    status text NOT NULL,
    input_params jsonb,
    result jsonb,
    error_message text,
    rows_returned integer,
    duration_ms integer,
    user_id uuid,
    user_role text,
    user_email text,
    is_async boolean DEFAULT false,
    created_at timestamptz DEFAULT now(),
    started_at timestamptz,
    completed_at timestamptz,
    tenant_id uuid,
    CONSTRAINT executions_pkey PRIMARY KEY (id),
    CONSTRAINT executions_procedure_id_fkey FOREIGN KEY (procedure_id) REFERENCES procedures (id) ON DELETE SET NULL,
    CONSTRAINT executions_status_check CHECK (status IN ('pending'::text, 'running'::text, 'completed'::text, 'failed'::text, 'cancelled'::text, 'timeout'::text))
);


COMMENT ON TABLE executions IS 'RPC execution history with input, output, and performance metrics';


COMMENT ON COLUMN rpc.executions.is_async IS 'Whether this was an async invocation (returns execution_id immediately)';

--
-- Name: idx_rpc_executions_created; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_created ON executions (created_at DESC);

--
-- Name: idx_rpc_executions_is_async; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_is_async ON executions (is_async) WHERE (is_async = true);

--
-- Name: idx_rpc_executions_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_namespace ON executions (namespace);

--
-- Name: idx_rpc_executions_procedure; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_procedure ON executions (procedure_id);

--
-- Name: idx_rpc_executions_procedure_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_procedure_name ON executions (procedure_name);

--
-- Name: idx_rpc_executions_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_status ON executions (status);

--
-- Name: idx_rpc_executions_user; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_user ON executions (user_id);

--
-- Name: executions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE executions ENABLE ROW LEVEL SECURITY;

--
-- Name: rpc_executions_instance_admin_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY rpc_executions_instance_admin_read ON executions FOR SELECT TO authenticated USING (auth.role() = 'instance_admin');

--
-- Name: rpc_executions_read_own; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY rpc_executions_read_own ON executions FOR SELECT TO authenticated USING (user_id = auth.current_user_id());

--
-- Name: notify_realtime_change(); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION notify_realtime_change()
RETURNS trigger
LANGUAGE plpgsql
VOLATILE
AS $$
DECLARE
  notification_record JSONB;
  old_notification_record JSONB;
BEGIN
  -- Build record without large fields for notification efficiency
  IF TG_OP != 'DELETE' THEN
    IF TG_TABLE_NAME = 'executions' THEN
      -- Exclude result and input_params (can be large)
      notification_record := to_jsonb(NEW) - 'result' - 'input_params';
    ELSE
      notification_record := to_jsonb(NEW);
    END IF;
  END IF;
  IF TG_OP != 'INSERT' THEN
    IF TG_TABLE_NAME = 'executions' THEN
      old_notification_record := to_jsonb(OLD) - 'result' - 'input_params';
    ELSE
      old_notification_record := to_jsonb(OLD);
    END IF;
  END IF;

  PERFORM pg_notify(
    'fluxbase_changes',
    json_build_object(
      'schema', TG_TABLE_SCHEMA,
      'table', TG_TABLE_NAME,
      'type', TG_OP,
      'record', notification_record,
      'old_record', old_notification_record
    )::text
  );
  RETURN COALESCE(NEW, OLD);
END;
$$;

--
-- Name: update_updated_at(); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS trigger
LANGUAGE plpgsql
VOLATILE
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

--
-- Cross-schema FKs moved to post-schema-fks.sql
-- procedures_created_by_fkey, executions_user_id_fkey
--

--
-- Name: executions_realtime_notify; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER executions_realtime_notify
    AFTER INSERT OR UPDATE OR DELETE ON executions
    FOR EACH ROW
    EXECUTE FUNCTION notify_realtime_change();

--
-- Name: procedures_update_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER procedures_update_updated_at
    BEFORE UPDATE ON procedures
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--
-- Multi-tenancy: procedures
--

CREATE INDEX IF NOT EXISTS idx_rpc_procedures_tenant_id ON procedures (tenant_id);

ALTER TABLE procedures FORCE ROW LEVEL SECURITY;

CREATE POLICY rpc_procedures_tenant ON procedures TO PUBLIC
    USING (auth.has_tenant_access(tenant_id))
    WITH CHECK (auth.has_tenant_access(tenant_id));

CREATE OR REPLACE TRIGGER rpc_procedures_set_tenant_id
    BEFORE INSERT ON procedures
    FOR EACH ROW
    EXECUTE FUNCTION auth.set_tenant_id_from_context();

--
-- Multi-tenancy: executions
--

CREATE INDEX IF NOT EXISTS idx_rpc_executions_tenant_id ON executions (tenant_id);

ALTER TABLE executions FORCE ROW LEVEL SECURITY;

CREATE POLICY rpc_executions_tenant ON executions TO PUBLIC
    USING (auth.has_tenant_access(tenant_id))
    WITH CHECK (auth.has_tenant_access(tenant_id));

CREATE OR REPLACE TRIGGER rpc_executions_set_tenant_id
    BEFORE INSERT ON executions
    FOR EACH ROW
    EXECUTE FUNCTION auth.set_tenant_id_from_context();

--
-- Name: notify_realtime_change(); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION notify_realtime_change() TO {{APP_USER}};

--
-- Name: notify_realtime_change(); Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: update_updated_at(); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION update_updated_at() TO {{APP_USER}};

--
-- Name: update_updated_at(); Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: executions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE executions TO authenticated;

--
-- Name: executions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE executions TO {{APP_USER}};

--
-- Name: executions; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: executions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE executions TO service_role;

--
-- Name: procedures; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE procedures TO anon;

--
-- Name: procedures; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE procedures TO authenticated;

--
-- Name: procedures; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE procedures TO {{APP_USER}};

--
-- Name: procedures; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: procedures; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE procedures TO service_role;

