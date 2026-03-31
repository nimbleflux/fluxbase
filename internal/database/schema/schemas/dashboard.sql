--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO dashboard, public;


--
-- Name: available_extensions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS available_extensions (
    id uuid DEFAULT gen_random_uuid(),
    name text NOT NULL,
    display_name text NOT NULL,
    description text,
    category text NOT NULL,
    is_core boolean DEFAULT false,
    requires_restart boolean DEFAULT false,
    documentation_url text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT available_extensions_pkey PRIMARY KEY (id),
    CONSTRAINT available_extensions_name_key UNIQUE (name),
    CONSTRAINT available_extensions_category_check CHECK (category IN ('core'::text, 'geospatial'::text, 'ai_ml'::text, 'monitoring'::text, 'scheduling'::text, 'data_types'::text, 'text_search'::text, 'indexing'::text, 'networking'::text, 'testing'::text, 'maintenance'::text, 'performance'::text, 'foreign_data'::text, 'triggers'::text, 'sampling'::text, 'utilities'::text))
);


COMMENT ON TABLE available_extensions IS 'Catalog of PostgreSQL extensions available in Fluxbase';


COMMENT ON COLUMN dashboard.available_extensions.name IS 'PostgreSQL extension name used in CREATE EXTENSION';


COMMENT ON COLUMN dashboard.available_extensions.is_core IS 'Core extensions are always enabled and cannot be disabled';


COMMENT ON COLUMN dashboard.available_extensions.requires_restart IS 'Extension requires PostgreSQL restart after enabling';

--
-- Name: idx_available_extensions_category; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_available_extensions_category ON available_extensions (category);

--
-- Name: idx_available_extensions_is_core; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_available_extensions_is_core ON available_extensions (is_core);

--
-- Name: email_templates; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS email_templates (
    id uuid DEFAULT gen_random_uuid(),
    template_type text NOT NULL,
    subject text NOT NULL,
    html_body text NOT NULL,
    text_body text,
    is_custom boolean DEFAULT false,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT email_templates_pkey PRIMARY KEY (id),
    CONSTRAINT email_templates_template_type_key UNIQUE (template_type)
);


COMMENT ON TABLE email_templates IS 'Customizable email templates for authentication flows';


COMMENT ON COLUMN dashboard.email_templates.template_type IS 'Type of template: magic_link, email_verification, password_reset';


COMMENT ON COLUMN dashboard.email_templates.is_custom IS 'Whether this template has been customized from defaults';

--
-- Name: idx_dashboard_email_templates_type; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_email_templates_type ON email_templates (template_type);

--
-- Name: email_templates; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE email_templates ENABLE ROW LEVEL SECURITY;

--
-- Name: email_templates; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE email_templates FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_email_templates_modify_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_email_templates_modify_policy ON email_templates TO PUBLIC USING (auth.current_user_role() = 'dashboard_admin');

--
-- Name: dashboard_email_templates_select_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_email_templates_select_policy ON email_templates FOR SELECT TO PUBLIC USING ((auth.current_user_role() = 'dashboard_admin') OR (auth.current_user_role() = 'service_role'));

--
-- Name: oauth_providers; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS oauth_providers (
    id uuid DEFAULT gen_random_uuid(),
    provider_name text NOT NULL,
    display_name text DEFAULT '' NOT NULL,
    client_id text NOT NULL,
    client_secret text NOT NULL,
    redirect_url text NOT NULL,
    scopes text[] DEFAULT ARRAY[]::text[],
    enabled boolean DEFAULT true,
    is_custom boolean DEFAULT false,
    authorization_url text,
    token_url text,
    user_info_url text,
    metadata jsonb DEFAULT '{}',
    created_by uuid,
    updated_by uuid,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    allow_dashboard_login boolean DEFAULT false,
    allow_app_login boolean DEFAULT true,
    required_claims jsonb,
    denied_claims jsonb,
    is_encrypted boolean DEFAULT false,
    revocation_endpoint text,
    end_session_endpoint text,
    CONSTRAINT oauth_providers_pkey PRIMARY KEY (id),
    CONSTRAINT oauth_providers_provider_name_key UNIQUE (provider_name)
);


COMMENT ON COLUMN dashboard.oauth_providers.allow_dashboard_login IS 'Allow this provider for dashboard admin SSO login';


COMMENT ON COLUMN dashboard.oauth_providers.allow_app_login IS 'Allow this provider for application user authentication';


COMMENT ON COLUMN dashboard.oauth_providers.required_claims IS 'JSON object of claims that must be present in ID token. Format: {"claim_name": ["value1", "value2"]}';


COMMENT ON COLUMN dashboard.oauth_providers.denied_claims IS 'JSON object of claims that, if present, will deny access. Format: {"claim_name": ["value1", "value2"]}';


COMMENT ON COLUMN dashboard.oauth_providers.is_encrypted IS 'Indicates whether client_secret is encrypted at rest using AES-256-GCM';


COMMENT ON COLUMN dashboard.oauth_providers.revocation_endpoint IS 'OAuth 2.0 Token Revocation endpoint (RFC 7009) for revoking access/refresh tokens';


COMMENT ON COLUMN dashboard.oauth_providers.end_session_endpoint IS 'OIDC RP-Initiated Logout endpoint for redirecting user to IdP logout page';

--
-- Name: idx_dashboard_oauth_providers_enabled; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_oauth_providers_enabled ON oauth_providers (enabled);

--
-- Name: idx_dashboard_oauth_providers_provider_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_oauth_providers_provider_name ON oauth_providers (provider_name);

--
-- Name: idx_oauth_providers_denied_claims; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_oauth_providers_denied_claims ON oauth_providers USING gin (denied_claims);

--
-- Name: idx_oauth_providers_required_claims; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_oauth_providers_required_claims ON oauth_providers USING gin (required_claims);

--
-- Name: oauth_providers; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE oauth_providers ENABLE ROW LEVEL SECURITY;

--
-- Name: oauth_providers; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE oauth_providers FORCE ROW LEVEL SECURITY;

--
-- Name: oauth_providers_dashboard_admin_only; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY oauth_providers_dashboard_admin_only ON oauth_providers TO PUBLIC USING ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'dashboard_admin'));

--
-- Name: users; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS users (
    id uuid DEFAULT gen_random_uuid(),
    email text NOT NULL,
    password_hash text NOT NULL,
    full_name text,
    avatar_url text,
    role text DEFAULT 'dashboard_user',
    user_metadata jsonb DEFAULT '{}',
    app_metadata jsonb DEFAULT '{}',
    email_verified boolean DEFAULT false,
    email_verified_at timestamptz,
    totp_enabled boolean DEFAULT false,
    totp_secret varchar(32),
    backup_codes text[],
    is_active boolean DEFAULT true,
    is_locked boolean DEFAULT false,
    failed_login_attempts integer DEFAULT 0,
    last_login_at timestamptz,
    deleted_at timestamptz,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    locked_until timestamptz,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_email_key UNIQUE (email)
);


COMMENT ON COLUMN dashboard.users.user_metadata IS 'User-editable metadata for dashboard users.';


COMMENT ON COLUMN dashboard.users.app_metadata IS 'Application/admin-only metadata for dashboard users.';


COMMENT ON COLUMN dashboard.users.locked_until IS 'Timestamp when the account lock expires. NULL means no lock or lock is permanent (based on is_locked). When locked_until has passed, the account should be automatically unlocked on next login attempt.';

--
-- Name: idx_dashboard_users_app_metadata; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_app_metadata ON users USING gin (app_metadata);

--
-- Name: idx_dashboard_users_email; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_email ON users (email);

--
-- Name: idx_dashboard_users_locked_until; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_locked_until ON users (locked_until) WHERE (locked_until IS NOT NULL);

--
-- Name: idx_dashboard_users_role; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_role ON users (role);

--
-- Name: idx_dashboard_users_user_metadata; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_user_metadata ON users USING gin (user_metadata);

--
-- Name: invitation_tokens; Type: TABLE; Schema: -; Owner: -
-- (Created early because users RLS policies reference it)
--

CREATE TABLE IF NOT EXISTS invitation_tokens (
    id uuid DEFAULT gen_random_uuid(),
    email text NOT NULL,
    token text NOT NULL,
    role text DEFAULT 'dashboard_user' NOT NULL,
    invited_by uuid,
    expires_at timestamptz NOT NULL,
    accepted boolean DEFAULT false,
    accepted_at timestamptz,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT invitation_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT invitation_tokens_token_key UNIQUE (token),
    CONSTRAINT invitation_tokens_invited_by_fkey FOREIGN KEY (invited_by) REFERENCES users (id) ON DELETE SET NULL
);

--
-- Name: idx_dashboard_invitation_tokens_accepted; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_invitation_tokens_accepted ON invitation_tokens (accepted);

--
-- Name: idx_dashboard_invitation_tokens_email; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_invitation_tokens_email ON invitation_tokens (email);

--
-- Name: idx_dashboard_invitation_tokens_expires_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_invitation_tokens_expires_at ON invitation_tokens (expires_at);

--
-- Name: idx_dashboard_invitation_tokens_token; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_invitation_tokens_token ON invitation_tokens (token);

--
-- Name: invitation_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE invitation_tokens ENABLE ROW LEVEL SECURITY;

--
-- Name: invitation_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE invitation_tokens FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_invitation_tokens_modify_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_invitation_tokens_modify_policy ON invitation_tokens TO PUBLIC USING (auth.current_user_role() = 'dashboard_admin');

--
-- Name: dashboard_invitation_tokens_select_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_invitation_tokens_select_policy ON invitation_tokens FOR SELECT TO PUBLIC USING (auth.current_user_role() = 'dashboard_admin');

--
-- Name: users; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE users FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_users_delete_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_users_delete_policy ON users FOR DELETE TO PUBLIC USING (auth.current_user_role() = 'dashboard_admin');

--
-- Name: dashboard_users_insert_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_users_insert_policy ON users FOR INSERT TO PUBLIC WITH CHECK ((( SELECT count(*) AS count FROM users users_1) = 0) OR (auth.current_user_role() = 'dashboard_admin') OR (EXISTS ( SELECT 1 FROM invitation_tokens WHERE ((invitation_tokens.token = current_setting('app.invitation_token', true)) AND (invitation_tokens.accepted = false) AND (invitation_tokens.expires_at > now())))));

--
-- Name: dashboard_users_select_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_users_select_policy ON users FOR SELECT TO PUBLIC USING ((auth.current_user_role() = 'dashboard_admin') OR ((auth.current_user_id())::text = (id)::text));

--
-- Name: dashboard_users_update_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_users_update_policy ON users FOR UPDATE TO PUBLIC USING ((auth.current_user_role() = 'dashboard_admin') OR ((auth.current_user_id())::text = (id)::text));

--
-- Name: activity_log; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS activity_log (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid,
    action text NOT NULL,
    resource_type text,
    resource_id text,
    details jsonb DEFAULT '{}',
    ip_address text,
    user_agent text,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT activity_log_pkey PRIMARY KEY (id),
    CONSTRAINT activity_log_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL
);

--
-- Name: idx_dashboard_activity_log_action; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_activity_log_action ON activity_log (action);

--
-- Name: idx_dashboard_activity_log_created_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_activity_log_created_at ON activity_log (created_at DESC);

--
-- Name: idx_dashboard_activity_log_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_activity_log_user_id ON activity_log (user_id);

--
-- Name: activity_log; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE activity_log ENABLE ROW LEVEL SECURITY;

--
-- Name: activity_log; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE activity_log FORCE ROW LEVEL SECURITY;

--
-- Name: activity_log_admin_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY activity_log_admin_read ON activity_log FOR SELECT TO PUBLIC USING (auth.current_user_role() = 'dashboard_admin');

--
-- Name: activity_log_service_write; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY activity_log_service_write ON activity_log FOR INSERT TO PUBLIC WITH CHECK (auth.current_user_role() = 'service_role');

--
-- Name: email_verification_tokens; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    token text NOT NULL,
    expires_at timestamptz NOT NULL,
    used boolean DEFAULT false,
    used_at timestamptz,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT email_verification_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT email_verification_tokens_token_key UNIQUE (token),
    CONSTRAINT email_verification_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

--
-- Name: idx_dashboard_email_verification_tokens_token; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_email_verification_tokens_token ON email_verification_tokens (token);

--
-- Name: email_verification_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE email_verification_tokens ENABLE ROW LEVEL SECURITY;

--
-- Name: email_verification_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE email_verification_tokens FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_email_verification_service_only; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_email_verification_service_only ON email_verification_tokens TO PUBLIC USING ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'dashboard_admin'));

--
-- Name: enabled_extensions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS enabled_extensions (
    id uuid DEFAULT gen_random_uuid(),
    extension_name text NOT NULL,
    tenant_id uuid,
    enabled_at timestamptz DEFAULT now() NOT NULL,
    enabled_by uuid,
    disabled_at timestamptz,
    disabled_by uuid,
    is_active boolean DEFAULT true,
    error_message text,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT enabled_extensions_pkey PRIMARY KEY (id),
    CONSTRAINT enabled_extensions_disabled_by_fkey FOREIGN KEY (disabled_by) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT enabled_extensions_enabled_by_fkey FOREIGN KEY (enabled_by) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT enabled_extensions_extension_name_fkey FOREIGN KEY (extension_name) REFERENCES available_extensions (name) ON DELETE CASCADE
);


COMMENT ON TABLE enabled_extensions IS 'Tracks which extensions are currently enabled, per tenant (NULL tenant_id = default tenant)';


COMMENT ON COLUMN dashboard.enabled_extensions.is_active IS 'Whether this extension is currently enabled';


COMMENT ON COLUMN dashboard.enabled_extensions.error_message IS 'Error message if enabling/disabling failed';


COMMENT ON COLUMN dashboard.enabled_extensions.tenant_id IS 'Tenant ID for per-tenant extension tracking; NULL means default tenant';

--
-- Name: idx_enabled_extensions_active; Type: INDEX; Schema: -; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_enabled_extensions_active ON enabled_extensions (extension_name, tenant_id) WHERE (is_active = true);

--
-- Name: idx_enabled_extensions_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_enabled_extensions_name ON enabled_extensions (extension_name);

--
-- Name: idx_enabled_extensions_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_enabled_extensions_tenant_id ON enabled_extensions (tenant_id);

--
-- Name: password_reset_tokens; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    token text NOT NULL,
    expires_at timestamptz NOT NULL,
    used boolean DEFAULT false,
    used_at timestamptz,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT password_reset_tokens_token_key UNIQUE (token),
    CONSTRAINT password_reset_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

--
-- Name: idx_dashboard_password_reset_tokens_token; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_password_reset_tokens_token ON password_reset_tokens (token);

--
-- Name: password_reset_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE password_reset_tokens ENABLE ROW LEVEL SECURITY;

--
-- Name: password_reset_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE password_reset_tokens FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_password_reset_service_only; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_password_reset_service_only ON password_reset_tokens TO PUBLIC USING ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'dashboard_admin'));

--
-- Name: schema_migrations; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS schema_migrations (
    id uuid DEFAULT gen_random_uuid(),
    schema_name text NOT NULL,
    migration_type text NOT NULL,
    migration_sql text NOT NULL,
    applied_by uuid,
    applied_at timestamptz DEFAULT now(),
    rolled_back boolean DEFAULT false,
    rolled_back_at timestamptz,
    rolled_back_by uuid,
    CONSTRAINT schema_migrations_pkey PRIMARY KEY (id),
    CONSTRAINT schema_migrations_applied_by_fkey FOREIGN KEY (applied_by) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT schema_migrations_rolled_back_by_fkey FOREIGN KEY (rolled_back_by) REFERENCES users (id) ON DELETE SET NULL
);

--
-- Name: idx_dashboard_schema_migrations_applied_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_schema_migrations_applied_at ON schema_migrations (applied_at DESC);

--
-- Name: idx_dashboard_schema_migrations_schema_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_schema_migrations_schema_name ON schema_migrations (schema_name);

--
-- Name: sessions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS sessions (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    token text NOT NULL,
    refresh_token text,
    ip_address text,
    user_agent text,
    expires_at timestamptz NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT sessions_pkey PRIMARY KEY (id),
    CONSTRAINT sessions_refresh_token_key UNIQUE (refresh_token),
    CONSTRAINT sessions_token_key UNIQUE (token),
    CONSTRAINT sessions_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

--
-- Name: idx_dashboard_sessions_refresh_token; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_sessions_refresh_token ON sessions (refresh_token);

--
-- Name: idx_dashboard_sessions_token; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_sessions_token ON sessions (token);

--
-- Name: idx_dashboard_sessions_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_sessions_user_id ON sessions (user_id);

--
-- Name: sessions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

--
-- Name: sessions; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE sessions FORCE ROW LEVEL SECURITY;

--
-- Name: dashboard_sessions_all_policy; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY dashboard_sessions_all_policy ON sessions TO PUBLIC USING ((auth.current_user_role() = 'dashboard_admin') OR ((auth.current_user_id())::text = (user_id)::text));

--
-- Name: sso_identities; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS sso_identities (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    provider_type text NOT NULL,
    provider_name text NOT NULL,
    provider_user_id text NOT NULL,
    email text,
    name text,
    raw_attributes jsonb DEFAULT '{}',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT sso_identities_pkey PRIMARY KEY (id),
    CONSTRAINT sso_identities_provider_type_provider_name_provider_user_id_key UNIQUE (provider_type, provider_name, provider_user_id),
    CONSTRAINT sso_identities_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT sso_identities_provider_type_check CHECK (provider_type IN ('oauth'::text, 'saml'::text))
);


COMMENT ON TABLE sso_identities IS 'Links dashboard admin users to their SSO identities';

--
-- Name: idx_dashboard_sso_identities_provider; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_sso_identities_provider ON sso_identities (provider_type, provider_name);

--
-- Name: idx_dashboard_sso_identities_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_sso_identities_user_id ON sso_identities (user_id);

--
-- Name: sso_identities; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE sso_identities ENABLE ROW LEVEL SECURITY;

--
-- Name: sso_identities; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE sso_identities FORCE ROW LEVEL SECURITY;

--
-- Name: SSO identities admin access; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY "SSO identities admin access" ON sso_identities TO PUBLIC USING ((auth.current_user_role() = 'service_role') OR (auth.current_user_role() = 'dashboard_admin'));

--
-- Name: update_sso_identities_updated_at(); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION update_sso_identities_updated_at()
RETURNS trigger
LANGUAGE plpgsql
VOLATILE
AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$;

--
-- Name: trigger_update_sso_identities_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER trigger_update_sso_identities_updated_at
    BEFORE UPDATE ON sso_identities
    FOR EACH ROW
    EXECUTE FUNCTION update_sso_identities_updated_at();

--
-- Name: update_dashboard_email_templates_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_dashboard_email_templates_updated_at
    BEFORE UPDATE ON email_templates
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: update_dashboard_oauth_providers_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_dashboard_oauth_providers_updated_at
    BEFORE UPDATE ON oauth_providers
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: update_dashboard_sessions_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_dashboard_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: update_dashboard_users_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER update_dashboard_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION platform.update_updated_at();

--
-- Name: validate_app_metadata_trigger; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER validate_app_metadata_trigger
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION auth.validate_app_metadata_update();

--
-- Name: activity_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE activity_log TO authenticated;

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
-- Name: available_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE available_extensions TO authenticated;

--
-- Name: available_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE available_extensions TO fluxbase_app;

--
-- Name: available_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE available_extensions TO fluxbase_rls_test;

--
-- Name: available_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE available_extensions TO service_role;

--
-- Name: email_templates; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE email_templates TO authenticated;

--
-- Name: email_templates; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_templates TO fluxbase_app;

--
-- Name: email_templates; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_templates TO fluxbase_rls_test;

--
-- Name: email_templates; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_templates TO service_role;

--
-- Name: email_verification_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE email_verification_tokens TO authenticated;

--
-- Name: email_verification_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_verification_tokens TO fluxbase_app;

--
-- Name: email_verification_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_verification_tokens TO fluxbase_rls_test;

--
-- Name: email_verification_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_verification_tokens TO service_role;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE enabled_extensions TO authenticated;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE enabled_extensions TO fluxbase_app;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE enabled_extensions TO fluxbase_rls_test;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE enabled_extensions TO service_role;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE invitation_tokens TO authenticated;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE invitation_tokens TO fluxbase_app;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE invitation_tokens TO fluxbase_rls_test;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE invitation_tokens TO service_role;

--
-- Name: oauth_providers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE oauth_providers TO authenticated;

--
-- Name: oauth_providers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE oauth_providers TO fluxbase_app;

--
-- Name: oauth_providers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE oauth_providers TO fluxbase_rls_test;

--
-- Name: oauth_providers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE oauth_providers TO service_role;

--
-- Name: password_reset_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE password_reset_tokens TO authenticated;

--
-- Name: password_reset_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE password_reset_tokens TO fluxbase_app;

--
-- Name: password_reset_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE password_reset_tokens TO fluxbase_rls_test;

--
-- Name: password_reset_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE password_reset_tokens TO service_role;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE schema_migrations TO authenticated;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_migrations TO fluxbase_app;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_migrations TO fluxbase_rls_test;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_migrations TO service_role;

--
-- Name: sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE sessions TO authenticated;

--
-- Name: sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sessions TO fluxbase_app;

--
-- Name: sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sessions TO fluxbase_rls_test;

--
-- Name: sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sessions TO service_role;

--
-- Name: sso_identities; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE sso_identities TO authenticated;

--
-- Name: sso_identities; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sso_identities TO fluxbase_app;

--
-- Name: sso_identities; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sso_identities TO fluxbase_rls_test;

--
-- Name: sso_identities; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sso_identities TO service_role;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE users TO authenticated;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE users TO fluxbase_app;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE users TO fluxbase_rls_test;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE users TO service_role;

