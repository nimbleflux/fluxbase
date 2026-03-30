--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4


--
-- Name: functions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS functions (
    id uuid DEFAULT gen_random_uuid(),
    name text NOT NULL,
    namespace text DEFAULT 'default' NOT NULL,
    description text,
    code text,
    original_code text,
    is_bundled boolean DEFAULT false,
    bundle_error text,
    enabled boolean DEFAULT true,
    schedule text,
    timeout_seconds integer DEFAULT 300,
    memory_limit_mb integer DEFAULT 256,
    max_retries integer DEFAULT 0,
    progress_timeout_seconds integer DEFAULT 60,
    allow_net boolean DEFAULT true,
    allow_env boolean DEFAULT true,
    allow_read boolean DEFAULT false,
    allow_write boolean DEFAULT false,
    version integer DEFAULT 1,
    created_by uuid,
    source text DEFAULT 'filesystem' NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    disable_execution_logs boolean DEFAULT false NOT NULL,
    require_roles text[] DEFAULT ARRAY[]::text[],
    CONSTRAINT functions_pkey PRIMARY KEY (id),
    CONSTRAINT functions_name_namespace_key UNIQUE (name, namespace)
);


COMMENT ON TABLE functions IS 'Job function definitions (templates for jobs)';


COMMENT ON COLUMN jobs.functions.code IS 'Bundled JavaScript/TypeScript code';


COMMENT ON COLUMN jobs.functions.original_code IS 'Original source code before bundling';


COMMENT ON COLUMN jobs.functions.schedule IS 'Cron expression for scheduled execution';


COMMENT ON COLUMN jobs.functions.source IS 'Source of function: filesystem or api';


COMMENT ON COLUMN jobs.functions.disable_execution_logs IS 'When true, execution logs are not created for this job (from @fluxbase:disable-execution-logs annotation)';


COMMENT ON COLUMN jobs.functions.require_roles IS 'Required roles to submit this job (admin, authenticated, anon, or custom roles). User needs ANY of the specified roles.';

--
-- Name: idx_functions_enabled; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_functions_enabled ON functions (enabled) WHERE (enabled = true);

--
-- Name: idx_functions_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_functions_namespace ON functions (namespace);

--
-- Name: idx_jobs_functions_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_jobs_functions_namespace ON functions (namespace);

--
-- Name: functions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE functions ENABLE ROW LEVEL SECURITY;

--
-- Name: Dashboard admins can read all functions; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Dashboard admins can read all functions" ON functions FOR SELECT TO authenticated USING (auth.role() = 'dashboard_admin');

--
-- Name: Service role can manage functions; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Service role can manage functions" ON functions TO service_role USING (true) WITH CHECK (true);

--
-- Name: function_files; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS function_files (
    id uuid DEFAULT gen_random_uuid(),
    function_id uuid NOT NULL,
    file_path text NOT NULL,
    content text,
    created_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT function_files_pkey PRIMARY KEY (id),
    CONSTRAINT function_files_function_id_file_path_key UNIQUE (function_id, file_path),
    CONSTRAINT function_files_function_id_fkey FOREIGN KEY (function_id) REFERENCES functions (id) ON DELETE CASCADE
);


COMMENT ON TABLE function_files IS 'Supporting files for multi-file job functions';

--
-- Name: idx_function_files_function_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_function_files_function_id ON function_files (function_id);

--
-- Name: function_files; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE function_files ENABLE ROW LEVEL SECURITY;

--
-- Name: Dashboard admins can read all function files; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Dashboard admins can read all function files" ON function_files FOR SELECT TO authenticated USING (auth.role() = 'dashboard_admin');

--
-- Name: Service role can manage function files; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Service role can manage function files" ON function_files TO service_role USING (true) WITH CHECK (true);

--
-- Name: workers; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS workers (
    id uuid DEFAULT gen_random_uuid(),
    name text,
    hostname text,
    status text NOT NULL,
    max_concurrent_jobs integer DEFAULT 5,
    current_job_count integer DEFAULT 0,
    last_heartbeat_at timestamptz DEFAULT now() NOT NULL,
    started_at timestamptz DEFAULT now() NOT NULL,
    metadata jsonb,
    CONSTRAINT workers_pkey PRIMARY KEY (id),
    CONSTRAINT workers_status_check CHECK (status IN ('active'::text, 'draining'::text, 'stopped'::text))
);


COMMENT ON TABLE workers IS 'Active worker registry';


COMMENT ON COLUMN jobs.workers.status IS 'Worker status: active=accepting jobs, draining=finishing current jobs, stopped=shut down';


COMMENT ON COLUMN jobs.workers.last_heartbeat_at IS 'Last heartbeat timestamp for health monitoring';

--
-- Name: idx_workers_heartbeat; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workers_heartbeat ON workers (last_heartbeat_at);

--
-- Name: idx_workers_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_workers_status ON workers (status);

--
-- Name: workers; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE workers ENABLE ROW LEVEL SECURITY;

--
-- Name: Dashboard admins can read all workers; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Dashboard admins can read all workers" ON workers FOR SELECT TO authenticated USING (auth.role() = 'dashboard_admin');

--
-- Name: Service role can manage workers; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Service role can manage workers" ON workers TO service_role USING (true) WITH CHECK (true);

--
-- Name: queue; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS queue (
    id uuid DEFAULT gen_random_uuid(),
    namespace text NOT NULL,
    function_id uuid,
    job_name text NOT NULL,
    status text NOT NULL,
    payload jsonb,
    result jsonb,
    progress jsonb,
    priority integer DEFAULT 0,
    max_duration_seconds integer,
    progress_timeout_seconds integer,
    max_retries integer DEFAULT 0,
    retry_count integer DEFAULT 0,
    error_message text,
    worker_id uuid,
    created_by uuid,
    user_role text,
    user_email text,
    user_name text,
    created_at timestamptz DEFAULT now() NOT NULL,
    scheduled_at timestamptz,
    started_at timestamptz,
    last_progress_at timestamptz,
    completed_at timestamptz,
    CONSTRAINT queue_pkey PRIMARY KEY (id),
    CONSTRAINT fk_queue_worker FOREIGN KEY (worker_id) REFERENCES workers (id) ON DELETE SET NULL,
    CONSTRAINT queue_function_id_fkey FOREIGN KEY (function_id) REFERENCES functions (id) ON DELETE SET NULL,
    CONSTRAINT queue_status_check CHECK (status IN ('pending'::text, 'running'::text, 'completed'::text, 'failed'::text, 'cancelled'::text))
);


COMMENT ON TABLE queue IS 'Job execution queue and history';


COMMENT ON COLUMN jobs.queue.status IS 'Job execution status';


COMMENT ON COLUMN jobs.queue.progress IS 'Current progress state (for running jobs)';


COMMENT ON COLUMN jobs.queue.priority IS 'Higher numbers = higher priority';


COMMENT ON COLUMN jobs.queue.user_role IS 'Role of the user who submitted the job';


COMMENT ON COLUMN jobs.queue.user_email IS 'Email of the user who submitted the job';


COMMENT ON COLUMN jobs.queue.user_name IS 'Display name of the user who submitted the job';

--
-- Name: idx_queue_created_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_created_at ON queue (created_at DESC);

--
-- Name: idx_queue_created_by; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_created_by ON queue (created_by);

--
-- Name: idx_queue_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_namespace ON queue (namespace);

--
-- Name: idx_queue_scheduled_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_scheduled_at ON queue (scheduled_at) WHERE (scheduled_at IS NOT NULL) AND (status = 'pending'::text);

--
-- Name: idx_queue_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_status ON queue (status);

--
-- Name: idx_queue_status_priority; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_queue_status_priority ON queue (status, priority DESC, created_at);

--
-- Name: queue; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE queue ENABLE ROW LEVEL SECURITY;

--
-- Name: Dashboard admins can read all jobs; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Dashboard admins can read all jobs" ON queue FOR SELECT TO authenticated USING (auth.role() = 'dashboard_admin');

--
-- Name: Service role can manage all jobs; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Service role can manage all jobs" ON queue TO service_role USING (true) WITH CHECK (true);

--
-- Name: Users can cancel their own pending/running jobs; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Users can cancel their own pending/running jobs" ON queue FOR UPDATE TO authenticated USING ((created_by = auth.uid()) AND (status = ANY (ARRAY['pending', 'running']))) WITH CHECK (status = 'cancelled');

--
-- Name: Users can read their own jobs; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Users can read their own jobs" ON queue FOR SELECT TO authenticated USING (created_by = auth.uid());

--
-- Name: Users can submit jobs; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Users can submit jobs" ON queue FOR INSERT TO authenticated WITH CHECK (created_by = auth.uid());

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
    notification_record := to_jsonb(NEW) - 'result' - 'payload';
  END IF;
  IF TG_OP != 'INSERT' THEN
    old_notification_record := to_jsonb(OLD) - 'result' - 'payload';
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
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION update_updated_at_column()
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
-- functions_created_by_fkey, queue_created_by_fkey
--

--
-- Name: function_files_realtime_notify; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER function_files_realtime_notify
    AFTER INSERT OR UPDATE OR DELETE ON function_files
    FOR EACH ROW
    EXECUTE FUNCTION notify_realtime_change();

--
-- Name: queue_realtime_notify; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER queue_realtime_notify
    AFTER INSERT OR UPDATE OR DELETE ON queue
    FOR EACH ROW
    EXECUTE FUNCTION notify_realtime_change();

--
-- Name: update_functions_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_functions_updated_at
    BEFORE UPDATE ON functions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--
-- Name: workers_realtime_notify; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER workers_realtime_notify
    AFTER INSERT OR UPDATE OR DELETE ON workers
    FOR EACH ROW
    EXECUTE FUNCTION notify_realtime_change();

--
-- Name: function_files; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE function_files TO authenticated;

--
-- Name: function_files; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE function_files TO fluxbase_app;

--
-- Name: function_files; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE function_files TO fluxbase_rls_test;

--
-- Name: function_files; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE function_files TO service_role;

--
-- Name: functions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE functions TO authenticated;

--
-- Name: functions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE functions TO fluxbase_app;

--
-- Name: functions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE functions TO fluxbase_rls_test;

--
-- Name: functions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE functions TO service_role;

--
-- Name: queue; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE queue TO authenticated;

--
-- Name: queue; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE queue TO fluxbase_app;

--
-- Name: queue; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE queue TO fluxbase_rls_test;

--
-- Name: queue; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE queue TO service_role;

--
-- Name: workers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE workers TO authenticated;

--
-- Name: workers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE workers TO fluxbase_app;

--
-- Name: workers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE workers TO fluxbase_rls_test;

--
-- Name: workers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE workers TO service_role;

