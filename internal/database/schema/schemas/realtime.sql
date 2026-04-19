--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO realtime, public;


--
-- Name: schema_registry; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS schema_registry (
    id SERIAL,
    schema_name text NOT NULL,
    table_name text NOT NULL,
    realtime_enabled boolean DEFAULT true,
    events text[] DEFAULT ARRAY['INSERT', 'UPDATE', 'DELETE'],
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    excluded_columns text[] DEFAULT '{}',
    tenant_id uuid,
    CONSTRAINT schema_registry_pkey PRIMARY KEY (id),
    CONSTRAINT schema_registry_schema_name_table_name_key UNIQUE (schema_name, table_name)
);

--
-- Name: schema_registry; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE schema_registry ENABLE ROW LEVEL SECURITY;

--
-- Name: schema_registry; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE schema_registry FORCE ROW LEVEL SECURITY;

--
-- Name: Admins can manage realtime configuration; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Admins can manage realtime configuration" ON schema_registry TO authenticated USING ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'instance_admin') OR auth.is_admin()) WITH CHECK ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'instance_admin') OR auth.is_admin());

--
-- Name: Authenticated users can view realtime configuration; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "Authenticated users can view realtime configuration" ON schema_registry FOR SELECT TO authenticated USING (auth.has_tenant_access(tenant_id));

--
-- Name: update_realtime_schema_registry_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_realtime_schema_registry_updated_at
    BEFORE UPDATE ON schema_registry
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: idx_realtime_schema_registry_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_realtime_schema_registry_tenant_id ON schema_registry (tenant_id);

--
-- Name: realtime_schema_registry_tenant; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY realtime_schema_registry_tenant ON schema_registry TO PUBLIC
    USING (auth.has_tenant_access(tenant_id))
    WITH CHECK (auth.has_tenant_access(tenant_id));

--
-- Name: realtime_schema_registry_set_tenant_id; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER realtime_schema_registry_set_tenant_id
    BEFORE INSERT ON schema_registry
    FOR EACH ROW
    EXECUTE FUNCTION auth.set_tenant_id_from_context();

--
-- Name: schema_registry_id_seq; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT, UPDATE, USAGE ON SEQUENCE schema_registry_id_seq TO {{APP_USER}};

--
-- Name: schema_registry_id_seq; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: schema_registry_id_seq; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT, UPDATE, USAGE ON SEQUENCE schema_registry_id_seq TO service_role;

--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE schema_registry TO authenticated;

--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_registry TO {{APP_USER}};

--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_registry TO service_role;

--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT USAGE ON SCHEMA realtime TO tenant_service;

--
-- Name: schema_registry; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE realtime.schema_registry TO tenant_service;

