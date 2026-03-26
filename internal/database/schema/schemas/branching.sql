--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4


--
-- Name: branches; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS branches (
    id uuid DEFAULT gen_random_uuid(),
    name text NOT NULL,
    slug text NOT NULL,
    database_name text NOT NULL,
    status text DEFAULT 'creating',
    type text DEFAULT 'preview',
    tenant_id uuid,
    parent_branch_id uuid,
    data_clone_mode text DEFAULT 'schema_only',
    github_pr_number integer,
    github_pr_url text,
    github_repo text,
    error_message text,
    created_by uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    expires_at timestamptz,
    seeds_path text,
    CONSTRAINT branches_pkey PRIMARY KEY (id),
    CONSTRAINT branches_name_tenant_unique UNIQUE (name, tenant_id),
    CONSTRAINT branches_slug_tenant_unique UNIQUE (slug, tenant_id),
    CONSTRAINT branches_parent_branch_id_fkey FOREIGN KEY (parent_branch_id) REFERENCES branches (id),
    CONSTRAINT branches_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id),
    CONSTRAINT branches_data_clone_mode_check CHECK (data_clone_mode IN ('schema_only'::text, 'full_clone'::text, 'seed_data'::text)),
    CONSTRAINT branches_status_check CHECK (status IN ('creating'::text, 'ready'::text, 'migrating'::text, 'error'::text, 'deleting'::text, 'deleted'::text)),
    CONSTRAINT branches_type_check CHECK (type IN ('main'::text, 'preview'::text, 'persistent'::text))
);

COMMENT ON COLUMN branches.tenant_id IS 'Tenant this branch belongs to. NULL = instance-level branch (backward compatibility)';

--
-- Name: idx_branches_created_by; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_created_by ON branches (created_by);

--
-- Name: idx_branches_expires_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_expires_at ON branches (expires_at) WHERE (expires_at IS NOT NULL);

--
-- Name: idx_branches_github_pr; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_github_pr ON branches (github_repo, github_pr_number) WHERE (github_pr_number IS NOT NULL);

--
-- Name: idx_branches_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_status ON branches (status);

--
-- Name: idx_branches_type; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_type ON branches (type);

--
-- Name: idx_branches_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branches_tenant_id ON branches (tenant_id);

--
-- Name: activity_log; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS activity_log (
    id uuid DEFAULT gen_random_uuid(),
    branch_id uuid,
    tenant_id uuid,
    action text NOT NULL,
    status text NOT NULL,
    details jsonb,
    error_message text,
    executed_by uuid,
    executed_at timestamptz DEFAULT now(),
    duration_ms integer,
    CONSTRAINT activity_log_pkey PRIMARY KEY (id),
    CONSTRAINT activity_log_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE,
    CONSTRAINT activity_log_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id),
    CONSTRAINT activity_log_action_check CHECK (action IN ('created'::text, 'cloned'::text, 'migrated'::text, 'reset'::text, 'deleted'::text, 'status_changed'::text, 'access_granted'::text, 'access_revoked'::text, 'seeding'::text)),
    CONSTRAINT activity_log_status_check CHECK (status IN ('started'::text, 'success'::text, 'failed'::text))
);

--
-- Name: idx_activity_log_branch_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_activity_log_branch_id ON activity_log (branch_id);

--
-- Name: idx_activity_log_executed_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_activity_log_executed_at ON activity_log (executed_at);

--
-- Name: idx_activity_log_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_activity_log_tenant_id ON activity_log (tenant_id);

--
-- Name: branch_access; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS branch_access (
    id uuid DEFAULT gen_random_uuid(),
    branch_id uuid,
    tenant_id uuid,
    user_id uuid,
    access_level text DEFAULT 'read',
    granted_at timestamptz DEFAULT now(),
    granted_by uuid,
    CONSTRAINT branch_access_pkey PRIMARY KEY (id),
    CONSTRAINT branch_access_unique UNIQUE (branch_id, user_id),
    CONSTRAINT branch_access_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE,
    CONSTRAINT branch_access_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id),
    CONSTRAINT branch_access_access_level_check CHECK (access_level IN ('read'::text, 'write'::text, 'admin'::text))
);

--
-- Name: idx_branch_access_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branch_access_user_id ON branch_access (user_id);

--
-- Name: idx_branch_access_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_branch_access_tenant_id ON branch_access (tenant_id);

--
-- Name: github_config; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS github_config (
    id uuid DEFAULT gen_random_uuid(),
    repository text NOT NULL,
    tenant_id uuid,
    auto_create_on_pr boolean DEFAULT true,
    auto_delete_on_merge boolean DEFAULT true,
    default_data_clone_mode text DEFAULT 'schema_only',
    webhook_secret text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT github_config_pkey PRIMARY KEY (id),
    CONSTRAINT github_config_repository_tenant_unique UNIQUE (repository, tenant_id),
    CONSTRAINT github_config_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id),
    CONSTRAINT github_config_default_data_clone_mode_check CHECK (default_data_clone_mode IN ('schema_only'::text, 'full_clone'::text, 'seed_data'::text))
);

COMMENT ON COLUMN github_config.tenant_id IS 'Tenant this GitHub config belongs to. NULL = instance-level config (backward compatibility)';

--
-- Name: migration_history; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS migration_history (
    id uuid DEFAULT gen_random_uuid(),
    branch_id uuid,
    tenant_id uuid,
    migration_version bigint NOT NULL,
    migration_name text,
    applied_at timestamptz DEFAULT now(),
    CONSTRAINT migration_history_pkey PRIMARY KEY (id),
    CONSTRAINT migration_history_unique UNIQUE (branch_id, migration_version),
    CONSTRAINT migration_history_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE,
    CONSTRAINT migration_history_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id)
);

--
-- Name: seed_execution_log; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS seed_execution_log (
    id uuid DEFAULT gen_random_uuid(),
    branch_id uuid,
    tenant_id uuid,
    seed_file_name text NOT NULL,
    status text NOT NULL,
    error_message text,
    executed_at timestamptz DEFAULT now(),
    duration_ms integer,
    CONSTRAINT seed_execution_log_pkey PRIMARY KEY (id),
    CONSTRAINT seed_execution_unique UNIQUE (branch_id, seed_file_name),
    CONSTRAINT seed_execution_log_branch_id_fkey FOREIGN KEY (branch_id) REFERENCES branches (id) ON DELETE CASCADE,
    CONSTRAINT seed_execution_log_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES platform.tenants (id),
    CONSTRAINT seed_execution_log_status_check CHECK (status IN ('started'::text, 'success'::text, 'failed'::text))
);

--
-- Name: idx_seed_execution_branch_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_seed_execution_branch_id ON seed_execution_log (branch_id);

--
-- Name: idx_seed_execution_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_seed_execution_status ON seed_execution_log (status);

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
-- branches_created_by_fkey, activity_log_executed_by_fkey, branch_access_granted_by_fkey, branch_access_user_id_fkey
--

--
-- Name: branches_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER branches_updated_at
    BEFORE UPDATE ON branches
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--
-- Name: github_config_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER github_config_updated_at
    BEFORE UPDATE ON github_config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

--
-- Name: activity_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE activity_log TO fluxbase_app;

--
-- Name: activity_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE activity_log TO fluxbase_rls_test;

--
-- Name: activity_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE activity_log TO service_role;

--
-- Name: branch_access; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branch_access TO fluxbase_app;

--
-- Name: branch_access; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branch_access TO fluxbase_rls_test;

--
-- Name: branch_access; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branch_access TO service_role;

--
-- Name: branches; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branches TO fluxbase_app;

--
-- Name: branches; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branches TO fluxbase_rls_test;

--
-- Name: branches; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE branches TO service_role;

--
-- Name: github_config; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE github_config TO fluxbase_app;

--
-- Name: github_config; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE github_config TO fluxbase_rls_test;

--
-- Name: github_config; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE github_config TO service_role;

--
-- Name: migration_history; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE migration_history TO fluxbase_app;

--
-- Name: migration_history; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE migration_history TO fluxbase_rls_test;

--
-- Name: migration_history; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE migration_history TO service_role;

--
-- Name: seed_execution_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE seed_execution_log TO fluxbase_app;

--
-- Name: seed_execution_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE seed_execution_log TO fluxbase_rls_test;

--
-- Name: seed_execution_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE seed_execution_log TO service_role;

