--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO mcp, public;


--
-- Name: custom_resources; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS custom_resources (
    id uuid DEFAULT gen_random_uuid(),
    uri varchar(255) NOT NULL,
    name varchar(64) NOT NULL,
    namespace varchar(64) DEFAULT 'default' NOT NULL,
    description text,
    mime_type varchar(64) DEFAULT 'application/json' NOT NULL,
    code text NOT NULL,
    is_template boolean DEFAULT false NOT NULL,
    required_scopes text[] DEFAULT '{}' NOT NULL,
    timeout_seconds integer DEFAULT 10 NOT NULL,
    cache_ttl_seconds integer DEFAULT 60 NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    tenant_id uuid,
    created_by uuid,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT custom_resources_pkey PRIMARY KEY (id),
    CONSTRAINT custom_resources_uri_namespace_unique UNIQUE (uri, namespace, tenant_id)
);


COMMENT ON TABLE custom_resources IS 'User-defined MCP resources implemented in TypeScript';


COMMENT ON COLUMN mcp.custom_resources.uri IS 'MCP resource URI (e.g., fluxbase://custom/myresource)';


COMMENT ON COLUMN mcp.custom_resources.is_template IS 'Whether URI contains parameters (e.g., {id})';


COMMENT ON COLUMN mcp.custom_resources.cache_ttl_seconds IS 'How long to cache resource responses';

--
-- Name: idx_custom_resources_enabled; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_resources_enabled ON custom_resources (enabled) WHERE (enabled = true);

--
-- Name: idx_custom_resources_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_resources_namespace ON custom_resources (namespace);

--
-- Name: idx_custom_resources_uri; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_resources_uri ON custom_resources (uri);

--
-- Name: custom_resources; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE custom_resources ENABLE ROW LEVEL SECURITY;

--
-- Name: custom_resources; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE custom_resources FORCE ROW LEVEL SECURITY;

--
-- Name: custom_resources_tenant; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY custom_resources_tenant ON custom_resources
    FOR ALL TO tenant_service
    USING (auth.has_tenant_access(tenant_id))
    WITH CHECK (auth.has_tenant_access(tenant_id));

--
-- Name: custom_resources_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY custom_resources_admin ON custom_resources
    FOR ALL TO service_role
    USING (true)
    WITH CHECK (true);

--
-- Name: custom_tools; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS custom_tools (
    id uuid DEFAULT gen_random_uuid(),
    name varchar(64) NOT NULL,
    namespace varchar(64) DEFAULT 'default' NOT NULL,
    description text,
    code text NOT NULL,
    input_schema jsonb DEFAULT '{"type": "object", "properties": {}}' NOT NULL,
    required_scopes text[] DEFAULT '{}' NOT NULL,
    timeout_seconds integer DEFAULT 30 NOT NULL,
    memory_limit_mb integer DEFAULT 128 NOT NULL,
    allow_net boolean DEFAULT true NOT NULL,
    allow_env boolean DEFAULT false NOT NULL,
    allow_read boolean DEFAULT false NOT NULL,
    allow_write boolean DEFAULT false NOT NULL,
    enabled boolean DEFAULT true NOT NULL,
    version integer DEFAULT 1 NOT NULL,
    tenant_id uuid,
    created_by uuid,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT custom_tools_pkey PRIMARY KEY (id),
    CONSTRAINT custom_tools_name_namespace_unique UNIQUE (name, namespace, tenant_id)
);


COMMENT ON TABLE custom_tools IS 'User-defined MCP tools implemented in TypeScript';


COMMENT ON COLUMN mcp.custom_tools.code IS 'TypeScript code implementing the tool handler';


COMMENT ON COLUMN mcp.custom_tools.input_schema IS 'JSON Schema defining the tool input parameters';


COMMENT ON COLUMN mcp.custom_tools.required_scopes IS 'MCP scopes required to execute this tool';


COMMENT ON COLUMN mcp.custom_tools.allow_net IS 'Allow network access in Deno sandbox';


COMMENT ON COLUMN mcp.custom_tools.allow_env IS 'Allow environment variable access in Deno sandbox';

--
-- Name: idx_custom_tools_enabled; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_tools_enabled ON custom_tools (enabled) WHERE (enabled = true);

--
-- Name: idx_custom_tools_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_tools_name ON custom_tools (name);

--
-- Name: idx_custom_tools_namespace; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_custom_tools_namespace ON custom_tools (namespace);

--
-- Name: custom_tools; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE custom_tools ENABLE ROW LEVEL SECURITY;

--
-- Name: custom_tools; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE custom_tools FORCE ROW LEVEL SECURITY;

--
-- Name: custom_tools_tenant; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY custom_tools_tenant ON custom_tools
    FOR ALL TO tenant_service
    USING (auth.has_tenant_access(tenant_id))
    WITH CHECK (auth.has_tenant_access(tenant_id));

--
-- Name: custom_tools_admin; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY custom_tools_admin ON custom_tools
    FOR ALL TO service_role
    USING (true)
    WITH CHECK (true);

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
-- custom_resources_created_by_fkey, custom_tools_created_by_fkey
--

--
-- Name: custom_resources_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER custom_resources_updated_at
    BEFORE UPDATE ON custom_resources
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--
-- Name: custom_tools_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER custom_tools_updated_at
    BEFORE UPDATE ON custom_tools
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--
-- Name: update_updated_at(); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION update_updated_at() TO {{APP_USER}};

--
-- Name: update_updated_at(); Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: custom_resources; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE custom_resources TO authenticated;

--
-- Name: custom_resources; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE custom_resources TO {{APP_USER}};

--
-- Name: custom_resources; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: custom_resources; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE custom_resources TO service_role;

--
-- Name: custom_tools; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE custom_tools TO authenticated;

--
-- Name: custom_tools; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE custom_tools TO {{APP_USER}};

--
-- Name: custom_tools; Type: PRIVILEGE; Schema: privileges; Owner: -
--


--
-- Name: custom_tools; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE custom_tools TO service_role;

--
-- Name: mcp; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT USAGE ON SCHEMA mcp TO tenant_service;

--
-- Name: custom_resources; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE mcp.custom_resources TO tenant_service;

--
-- Name: custom_tools; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE mcp.custom_tools TO tenant_service;

