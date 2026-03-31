--
-- pgschema database dump
--

-- Dumped from database version PostgreSQL 18.3
-- Dumped by pgschema version 1.7.4

SET search_path TO platform;


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


COMMENT ON COLUMN platform.available_extensions.name IS 'PostgreSQL extension name used in CREATE EXTENSION';


COMMENT ON COLUMN platform.available_extensions.is_core IS 'Core extensions are always enabled and cannot be disabled';


COMMENT ON COLUMN platform.available_extensions.requires_restart IS 'Extension requires PostgreSQL restart after enabling';

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


COMMENT ON COLUMN platform.email_templates.template_type IS 'Type of template: magic_link, email_verification, password_reset';


COMMENT ON COLUMN platform.email_templates.is_custom IS 'Whether this template has been customized from defaults';

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
-- Name: platform_email_templates_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_email_templates_read ON email_templates FOR SELECT TO authenticated USING (true);

--
-- Name: platform_email_templates_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_email_templates_service_all ON email_templates TO service_role USING (true) WITH CHECK (true);

--
-- Name: instance_settings; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS instance_settings (
    id uuid DEFAULT gen_random_uuid(),
    settings jsonb DEFAULT '{}' NOT NULL,
    overridable_settings jsonb,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT instance_settings_pkey PRIMARY KEY (id)
);

--
-- Name: idx_instance_settings_settings; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_instance_settings_settings ON instance_settings USING gin (settings);

--
-- Name: idx_instance_settings_single_row; Type: INDEX; Schema: -; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_instance_settings_single_row ON instance_settings ((id IS NOT NULL));

--
-- Name: instance_settings; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE instance_settings ENABLE ROW LEVEL SECURITY;

--
-- Name: instance_settings_delete; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY instance_settings_delete ON instance_settings FOR DELETE TO PUBLIC USING (CURRENT_USER = 'service_role'::name);

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


COMMENT ON COLUMN platform.oauth_providers.allow_dashboard_login IS 'Allow this provider for dashboard admin SSO login';


COMMENT ON COLUMN platform.oauth_providers.allow_app_login IS 'Allow this provider for application user authentication';


COMMENT ON COLUMN platform.oauth_providers.required_claims IS 'JSON object of claims that must be present in ID token. Format: {"claim_name": ["value1", "value2"]}';


COMMENT ON COLUMN platform.oauth_providers.denied_claims IS 'JSON object of claims that, if present, will deny access. Format: {"claim_name": ["value1", "value2"]}';


COMMENT ON COLUMN platform.oauth_providers.is_encrypted IS 'Indicates whether client_secret is encrypted at rest using AES-256-GCM';


COMMENT ON COLUMN platform.oauth_providers.revocation_endpoint IS 'OAuth 2.0 Token Revocation endpoint (RFC 7009) for revoking access/refresh tokens';


COMMENT ON COLUMN platform.oauth_providers.end_session_endpoint IS 'OIDC RP-Initiated Logout endpoint for redirecting user to IdP logout page';

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
-- Name: platform_oauth_providers_read; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_oauth_providers_read ON oauth_providers FOR SELECT TO authenticated USING (enabled = true);

--
-- Name: platform_oauth_providers_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_oauth_providers_service_all ON oauth_providers TO service_role USING (true) WITH CHECK (true);

--
-- Name: tenants; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS tenants (
    id uuid DEFAULT gen_random_uuid(),
    slug text NOT NULL,
    name text NOT NULL,
    is_default boolean DEFAULT false,
    metadata jsonb DEFAULT '{}',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    deleted_at timestamptz,
    db_name text,
    status text DEFAULT 'active' NOT NULL,
    CONSTRAINT tenants_pkey PRIMARY KEY (id),
    CONSTRAINT tenants_slug_key UNIQUE (slug)
);


COMMENT ON TABLE tenants IS 'Tenant registry for database-per-tenant multi-tenancy. db_name = NULL means use main database.';


COMMENT ON COLUMN platform.tenants.id IS 'Unique identifier for the tenant';


COMMENT ON COLUMN platform.tenants.slug IS 'URL-friendly identifier for the tenant (e.g., "acme-corp")';


COMMENT ON COLUMN platform.tenants.name IS 'Display name for the tenant';


COMMENT ON COLUMN platform.tenants.is_default IS 'True for the default tenant used for backward compatibility';


COMMENT ON COLUMN platform.tenants.metadata IS 'Arbitrary metadata for the tenant (plan, settings, etc.)';


COMMENT ON COLUMN platform.tenants.deleted_at IS 'Soft delete timestamp. NULL if tenant is active.';


COMMENT ON COLUMN platform.tenants.db_name IS 'Database name for this tenant. NULL = use main database (backward compatibility for default tenant)';


COMMENT ON COLUMN platform.tenants.status IS 'Tenant status: creating, active, deleting, error';

--
-- Name: idx_platform_tenants_deleted_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenants_deleted_at ON tenants (deleted_at) WHERE (deleted_at IS NOT NULL);

--
-- Name: idx_platform_tenants_is_default; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenants_is_default ON tenants (is_default) WHERE (is_default = true);

--
-- Name: idx_platform_tenants_slug; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenants_slug ON tenants (slug);

--
-- Name: idx_platform_tenants_status; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenants_status ON tenants (status);

--
-- Name: idx_tenants_deleted_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenants_deleted_at ON tenants (deleted_at) WHERE (deleted_at IS NOT NULL);

--
-- Name: idx_tenants_is_default; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenants_is_default ON tenants (is_default) WHERE (is_default = true);

--
-- Name: idx_tenants_slug; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenants_slug ON tenants (slug);

--
-- Name: tenants; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

--
-- Name: tenants; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE tenants FORCE ROW LEVEL SECURITY;

--
-- Name: tenants_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY tenants_service_all ON tenants TO service_role USING (true) WITH CHECK (true);

--
-- Name: service_keys; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS service_keys (
    id uuid DEFAULT gen_random_uuid(),
    key_type text NOT NULL,
    tenant_id uuid,
    name text NOT NULL,
    description text,
    key_hash text NOT NULL,
    key_prefix text NOT NULL,
    user_id uuid,
    scopes text[] DEFAULT ARRAY[]::text[],
    allowed_namespaces text[],
    rate_limit_per_minute integer DEFAULT 60,
    is_active boolean DEFAULT true,
    is_config_managed boolean DEFAULT false,
    revoked_at timestamptz,
    revoked_by uuid,
    revocation_reason text,
    deprecated_at timestamptz,
    grace_period_ends_at timestamptz,
    replaced_by uuid,
    created_at timestamptz DEFAULT now(),
    created_by uuid,
    updated_at timestamptz DEFAULT now(),
    last_used_at timestamptz,
    expires_at timestamptz,
    CONSTRAINT service_keys_pkey PRIMARY KEY (id),
    CONSTRAINT service_keys_key_prefix_key UNIQUE (key_prefix),
    CONSTRAINT service_keys_replaced_by_fkey FOREIGN KEY (replaced_by) REFERENCES service_keys (id) ON DELETE SET NULL,
    CONSTRAINT service_keys_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants (id) ON DELETE CASCADE,
    CONSTRAINT service_keys_key_type_check CHECK (key_type IN ('anon'::text, 'publishable'::text, 'tenant_service'::text, 'global_service'::text))
);


COMMENT ON TABLE service_keys IS 'Unified API key management consolidating client keys and service keys with multi-tenant support';


COMMENT ON COLUMN platform.service_keys.key_type IS 'Type of key: anon (anonymous), publishable (user-scoped), tenant_service (tenant-level), global_service (instance-level)';


COMMENT ON COLUMN platform.service_keys.tenant_id IS 'Tenant this key belongs to. NULL for global_service keys.';


COMMENT ON COLUMN platform.service_keys.key_hash IS 'Bcrypt hash of the full key. Never store keys in plaintext.';


COMMENT ON COLUMN platform.service_keys.key_prefix IS 'First characters of the key for identification in logs';


COMMENT ON COLUMN platform.service_keys.user_id IS 'User who owns this key (for publishable keys only)';


COMMENT ON COLUMN platform.service_keys.scopes IS 'Array of scope strings defining what this key can access';


COMMENT ON COLUMN platform.service_keys.allowed_namespaces IS 'Array of table/view namespaces this key can access (NULL = all)';


COMMENT ON COLUMN platform.service_keys.is_active IS 'Whether this key is currently usable';


COMMENT ON COLUMN platform.service_keys.is_config_managed IS 'Whether this key was created from configuration file';


COMMENT ON COLUMN platform.service_keys.revoked_at IS 'When the key was emergency revoked (NULL if not revoked)';


COMMENT ON COLUMN platform.service_keys.revoked_by IS 'User who revoked the key';


COMMENT ON COLUMN platform.service_keys.revocation_reason IS 'Reason for emergency revocation';


COMMENT ON COLUMN platform.service_keys.deprecated_at IS 'When the key was marked for rotation';


COMMENT ON COLUMN platform.service_keys.grace_period_ends_at IS 'When the grace period for rotation ends';


COMMENT ON COLUMN platform.service_keys.replaced_by IS 'Reference to the replacement key (for rotation)';

--
-- Name: idx_platform_service_keys_grace_period; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_grace_period ON service_keys (grace_period_ends_at) WHERE (deprecated_at IS NOT NULL) AND (grace_period_ends_at IS NOT NULL);

--
-- Name: idx_platform_service_keys_is_active; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_is_active ON service_keys (is_active) WHERE (is_active = true);

--
-- Name: idx_platform_service_keys_key_prefix; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_key_prefix ON service_keys (key_prefix);

--
-- Name: idx_platform_service_keys_key_type; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_key_type ON service_keys (key_type);

--
-- Name: idx_platform_service_keys_revoked_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_revoked_at ON service_keys (revoked_at) WHERE (revoked_at IS NOT NULL);

--
-- Name: idx_platform_service_keys_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_tenant_id ON service_keys (tenant_id);

--
-- Name: idx_platform_service_keys_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_service_keys_user_id ON service_keys (user_id) WHERE (user_id IS NOT NULL);

--
-- Name: key_usage; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS key_usage (
    id uuid DEFAULT gen_random_uuid(),
    key_id uuid NOT NULL,
    endpoint text NOT NULL,
    method text NOT NULL,
    status_code integer,
    response_time_ms integer,
    ip_address text,
    user_agent text,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT key_usage_pkey PRIMARY KEY (id),
    CONSTRAINT key_usage_key_id_fkey FOREIGN KEY (key_id) REFERENCES service_keys (id) ON DELETE CASCADE
);


COMMENT ON TABLE key_usage IS 'Usage tracking for service keys with request details';


COMMENT ON COLUMN platform.key_usage.key_id IS 'Reference to the service key used';

--
-- Name: idx_platform_key_usage_created_at; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_key_usage_created_at ON key_usage (created_at DESC);

--
-- Name: idx_platform_key_usage_key_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_key_usage_key_id ON key_usage (key_id);

--
-- Name: tenant_memberships; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS tenant_memberships (
    id uuid DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    role text DEFAULT 'tenant_member' NOT NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    CONSTRAINT tenant_memberships_pkey PRIMARY KEY (id),
    CONSTRAINT tenant_memberships_unique UNIQUE (tenant_id, user_id),
    CONSTRAINT tenant_memberships_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants (id) ON DELETE CASCADE,
    CONSTRAINT tenant_memberships_role_check CHECK (role IN ('tenant_admin'::text, 'tenant_member'::text))
);


COMMENT ON TABLE tenant_memberships IS 'Maps users to tenants with specific roles for multi-tenant access control';


COMMENT ON COLUMN platform.tenant_memberships.tenant_id IS 'Reference to the tenant';


COMMENT ON COLUMN platform.tenant_memberships.user_id IS 'Reference to the auth.users table';


COMMENT ON COLUMN platform.tenant_memberships.role IS 'User role within tenant: tenant_admin (manage members) or tenant_member (regular access)';

--
-- Name: idx_platform_tenant_memberships_role; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_role ON tenant_memberships (role);

--
-- Name: idx_platform_tenant_memberships_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_tenant_id ON tenant_memberships (tenant_id);

--
-- Name: idx_platform_tenant_memberships_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenant_memberships_user_id ON tenant_memberships (user_id);

--
-- Name: idx_tenant_memberships_role; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_role ON tenant_memberships (role);

--
-- Name: idx_tenant_memberships_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_tenant_id ON tenant_memberships (tenant_id);

--
-- Name: idx_tenant_memberships_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_memberships_user_id ON tenant_memberships (user_id);

--
-- Name: tenant_memberships; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE tenant_memberships ENABLE ROW LEVEL SECURITY;

--
-- Name: tenant_memberships_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY tenant_memberships_service_all ON tenant_memberships TO service_role USING (true) WITH CHECK (true);

--
-- Name: tenant_settings; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS tenant_settings (
    id uuid DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL,
    settings jsonb DEFAULT '{}' NOT NULL,
    created_at timestamptz DEFAULT now() NOT NULL,
    updated_at timestamptz DEFAULT now() NOT NULL,
    CONSTRAINT tenant_settings_pkey PRIMARY KEY (id),
    CONSTRAINT tenant_settings_tenant_unique UNIQUE (tenant_id),
    CONSTRAINT tenant_settings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants (id) ON DELETE CASCADE
);

--
-- Name: idx_tenant_settings_settings; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_settings_settings ON tenant_settings USING gin (settings);

--
-- Name: idx_tenant_settings_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_settings_tenant_id ON tenant_settings (tenant_id);

--
-- Name: tenant_settings; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE tenant_settings ENABLE ROW LEVEL SECURITY;

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
    CONSTRAINT users_email_key UNIQUE (email),
    CONSTRAINT platform_users_role_check CHECK (role IN ('instance_admin'::text, 'tenant_admin'::text, 'dashboard_admin'::text, 'dashboard_user'::text))
);


COMMENT ON COLUMN platform.users.role IS 'User role: instance_admin (global admin managing all tenants), tenant_admin (admin for specific tenant), dashboard_admin (legacy, maps to tenant_admin), dashboard_user (limited read-only access)';


COMMENT ON COLUMN platform.users.user_metadata IS 'User-editable metadata for dashboard users.';


COMMENT ON COLUMN platform.users.app_metadata IS 'Application/admin-only metadata for dashboard users.';


COMMENT ON COLUMN platform.users.locked_until IS 'Timestamp when the account lock expires. NULL means no lock or lock is permanent (based on is_locked). When locked_until has passed, the account should be automatically unlocked on next login attempt.';

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
-- Name: idx_dashboard_users_role_instance_admin; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_role_instance_admin ON users (role) WHERE (role = 'instance_admin'::text);

--
-- Name: idx_dashboard_users_user_metadata; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_users_user_metadata ON users USING gin (user_metadata);

--
-- Name: idx_platform_users_role_instance_admin; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_users_role_instance_admin ON users (role) WHERE (role = 'instance_admin'::text);

--
-- Name: users; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE users FORCE ROW LEVEL SECURITY;

--
-- Name: platform_users_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_users_service_all ON users TO service_role USING (true) WITH CHECK (true);

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
-- Name: platform_activity_log_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_activity_log_service_all ON activity_log TO service_role USING (true) WITH CHECK (true);

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
-- Name: platform_email_verification_tokens_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_email_verification_tokens_service_all ON email_verification_tokens TO service_role USING (true) WITH CHECK (true);

--
-- Name: enabled_extensions; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS enabled_extensions (
    id uuid DEFAULT gen_random_uuid(),
    extension_name text NOT NULL,
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


COMMENT ON TABLE enabled_extensions IS 'Tracks which extensions are currently enabled';


COMMENT ON COLUMN platform.enabled_extensions.is_active IS 'Whether this extension is currently enabled';


COMMENT ON COLUMN platform.enabled_extensions.error_message IS 'Error message if enabling/disabling failed';

--
-- Name: idx_enabled_extensions_active; Type: INDEX; Schema: -; Owner: -
--

CREATE UNIQUE INDEX IF NOT EXISTS idx_enabled_extensions_active ON enabled_extensions (extension_name) WHERE (is_active = true);

--
-- Name: idx_enabled_extensions_name; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_enabled_extensions_name ON enabled_extensions (extension_name);

--
-- Name: invitation_tokens; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS invitation_tokens (
    id uuid DEFAULT gen_random_uuid(),
    email text NOT NULL,
    token text NOT NULL,
    role text DEFAULT 'dashboard_user' NOT NULL,
    tenant_id uuid,
    invited_by uuid,
    expires_at timestamptz NOT NULL,
    accepted boolean DEFAULT false,
    accepted_at timestamptz,
    created_at timestamptz DEFAULT now(),
    CONSTRAINT invitation_tokens_pkey PRIMARY KEY (id),
    CONSTRAINT invitation_tokens_token_key UNIQUE (token),
    CONSTRAINT invitation_tokens_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants (id) ON DELETE CASCADE,
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
-- Name: idx_dashboard_invitation_tokens_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_dashboard_invitation_tokens_tenant_id ON invitation_tokens (tenant_id);

--
-- Name: invitation_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE invitation_tokens ENABLE ROW LEVEL SECURITY;

--
-- Name: invitation_tokens; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE invitation_tokens FORCE ROW LEVEL SECURITY;

--
-- Name: platform_invitation_tokens_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_invitation_tokens_service_all ON invitation_tokens TO service_role USING (true) WITH CHECK (true);

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
-- Name: platform_password_reset_tokens_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_password_reset_tokens_service_all ON password_reset_tokens TO service_role USING (true) WITH CHECK (true);

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
-- Name: platform_sessions_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_sessions_service_all ON sessions TO service_role USING (true) WITH CHECK (true);

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
-- Name: platform_sso_identities_service_all; Type: POLICY; Schema: -; Owner: -
--

CREATE POLICY platform_sso_identities_service_all ON sso_identities TO service_role USING (true) WITH CHECK (true);

--
-- Name: tenant_admin_assignments; Type: TABLE; Schema: -; Owner: -
--

CREATE TABLE IF NOT EXISTS tenant_admin_assignments (
    id uuid DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    assigned_by uuid,
    assigned_at timestamptz DEFAULT now(),
    CONSTRAINT tenant_admin_assignments_pkey PRIMARY KEY (id),
    CONSTRAINT tenant_admin_assignments_unique UNIQUE (user_id, tenant_id),
    CONSTRAINT tenant_admin_assignments_assigned_by_fkey FOREIGN KEY (assigned_by) REFERENCES users (id) ON DELETE SET NULL,
    CONSTRAINT tenant_admin_assignments_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES tenants (id) ON DELETE CASCADE,
    CONSTRAINT tenant_admin_assignments_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);


COMMENT ON TABLE tenant_admin_assignments IS 'Maps platform users to tenants as tenant administrators';


COMMENT ON COLUMN platform.tenant_admin_assignments.user_id IS 'Reference to the platform.users table (platform admin)';


COMMENT ON COLUMN platform.tenant_admin_assignments.tenant_id IS 'Reference to the platform.tenants table';


COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_by IS 'Platform user who assigned this admin role';


COMMENT ON COLUMN platform.tenant_admin_assignments.assigned_at IS 'Timestamp when the admin assignment was created';

--
-- Name: idx_platform_tenant_admin_assignments_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenant_admin_assignments_tenant_id ON tenant_admin_assignments (tenant_id);

--
-- Name: idx_platform_tenant_admin_assignments_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_platform_tenant_admin_assignments_user_id ON tenant_admin_assignments (user_id);

--
-- Name: idx_tenant_admin_assignments_tenant_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_admin_assignments_tenant_id ON tenant_admin_assignments (tenant_id);

--
-- Name: idx_tenant_admin_assignments_user_id; Type: INDEX; Schema: -; Owner: -
--

CREATE INDEX IF NOT EXISTS idx_tenant_admin_assignments_user_id ON tenant_admin_assignments (user_id);

--
-- Name: tenant_admin_assignments; Type: RLS; Schema: -; Owner: -
--

ALTER TABLE tenant_admin_assignments ENABLE ROW LEVEL SECURITY;

--
-- Name: delete_jsonb_path(jsonb, text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION delete_jsonb_path(
    p_jsonb jsonb,
    p_path text
)
RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_path_for_delete TEXT[];
BEGIN
    IF p_path IS NULL OR p_path = '' THEN
        RETURN p_jsonb;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_path_for_delete := ARRAY[]::TEXT[];

    -- Build path array for #- operator
    FOR i IN 1 .. array_length(v_parts, 1) LOOP
        v_path_for_delete := array_append(v_path_for_delete, v_parts[i]);
    END LOOP;

    RETURN p_jsonb #- v_path_for_delete;
END;
$$;

--
-- Name: delete_jsonb_path(jsonb, text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION delete_jsonb_path(jsonb, text) IS 'Deletes a nested key from JSONB using a dot-separated path.';

--
-- Name: get_jsonb_path(jsonb, text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION get_jsonb_path(
    p_jsonb jsonb,
    p_path text
)
RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_current JSONB;
    v_part TEXT;
BEGIN
    IF p_jsonb IS NULL OR p_path IS NULL OR p_path = '' THEN
        RETURN NULL;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_current := p_jsonb;

    FOREACH v_part IN ARRAY v_parts
    LOOP
        IF v_current IS NULL THEN
            RETURN NULL;
        END IF;

        -- Try to get the key from the current object
        IF jsonb_typeof(v_current) = 'object' THEN
            v_current := v_current->v_part;
        ELSE
            RETURN NULL;
        END IF;
    END LOOP;

    RETURN v_current;
END;
$$;

--
-- Name: get_jsonb_path(jsonb, text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION get_jsonb_path(jsonb, text) IS 'Extracts a value from JSONB using a dot-separated path (e.g., "ai.providers.openai.enabled").';

--
-- Name: is_setting_overridable(text); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION is_setting_overridable(
    setting_path text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY DEFINER
SET search_path = platform
AS $$
    SELECT
        CASE
            -- If overridable_settings is NULL, all settings can be overridden
            WHEN (SELECT overridable_settings FROM instance_settings LIMIT 1) IS NULL THEN
                TRUE
            -- Otherwise, check if the path is in the overridable list
            ELSE
                (SELECT overridable_settings FROM instance_settings LIMIT 1) ? setting_path
                OR EXISTS (
                    SELECT 1
                    FROM instance_settings,
                         jsonb_array_elements_text(overridable_settings) AS allowed_path
                    WHERE setting_path LIKE (allowed_path || '%')
                    LIMIT 1
                )
        END;
$$;

--
-- Name: is_setting_overridable(text); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION is_setting_overridable(text) IS 'Checks if a setting path can be overridden at the tenant level. Returns TRUE if overridable_settings is NULL (all allowed) or if the path matches an entry in the overridable list.';

--
-- Name: get_all_settings(uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION get_all_settings(
    p_tenant_id uuid
)
RETURNS TABLE(setting_path text, setting_value jsonb, source text, is_overridable boolean)
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = platform
AS $$
DECLARE
    v_instance_settings JSONB;
    v_tenant_settings JSONB;
    v_overridable JSONB;
    v_all_paths TEXT[];
BEGIN
    -- Get instance settings
    SELECT settings, overridable_settings
    INTO v_instance_settings, v_overridable
    FROM instance_settings
    LIMIT 1;

    -- Get tenant settings
    SELECT settings
    INTO v_tenant_settings
    FROM tenant_settings
    WHERE tenant_id = p_tenant_id;

    v_tenant_settings := COALESCE(v_tenant_settings, '{}');
    v_instance_settings := COALESCE(v_instance_settings, '{}');

    -- Extract all unique paths from both instance and tenant settings
    -- This is a simplified version - in practice you'd recursively extract all paths
    -- For now, we'll return the merged settings with metadata

    -- Return flattened settings with source info
    -- Note: A full implementation would recursively walk the JSONB trees
    -- This is a placeholder that returns the top-level structure

    RETURN QUERY
    SELECT
        key AS setting_path,
        COALESCE(
            CASE
                WHEN is_setting_overridable(key) AND v_tenant_settings ? key
                THEN v_tenant_settings->key
                ELSE v_instance_settings->key
            END,
            v_instance_settings->key
        ) AS setting_value,
        CASE
            WHEN is_setting_overridable(key) AND v_tenant_settings ? key
            THEN 'tenant'
            ELSE 'instance'
        END AS source,
        is_setting_overridable(key) AS is_overridable
    FROM jsonb_object_keys(v_instance_settings || v_tenant_settings) AS key;
END;
$$;

--
-- Name: get_all_settings(uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION get_all_settings(uuid) IS 'Returns all resolved settings for a tenant with source information (tenant, instance, or default).';

--
-- Name: get_setting(uuid, text, jsonb); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION get_setting(
    p_tenant_id uuid,
    p_setting_path text,
    p_default jsonb DEFAULT NULL
)
RETURNS jsonb
LANGUAGE plpgsql
VOLATILE
SECURITY DEFINER
SET search_path = platform
AS $$
DECLARE
    v_tenant_value JSONB;
    v_instance_value JSONB;
    v_is_overridable BOOLEAN;
BEGIN
    -- Check if this setting can be overridden
    v_is_overridable := is_setting_overridable(p_setting_path);

    -- If overridable and tenant_id provided, try to get tenant value first
    IF v_is_overridable AND p_tenant_id IS NOT NULL THEN
        SELECT get_jsonb_path(settings, p_setting_path)
        INTO v_tenant_value
        FROM tenant_settings
        WHERE tenant_id = p_tenant_id;

        IF v_tenant_value IS NOT NULL THEN
            RETURN v_tenant_value;
        END IF;
    END IF;

    -- Fall back to instance setting
    SELECT get_jsonb_path(settings, p_setting_path)
    INTO v_instance_value
    FROM instance_settings
    LIMIT 1;

    IF v_instance_value IS NOT NULL THEN
        RETURN v_instance_value;
    END IF;

    -- Fall back to provided default
    RETURN p_default;
END;
$$;

--
-- Name: get_setting(uuid, text, jsonb); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION get_setting(uuid, text, jsonb) IS 'Resolves a setting value using cascade: tenant -> instance -> default. Respects overridable_settings restrictions.';

--
-- Name: set_jsonb_path(jsonb, text, jsonb); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION set_jsonb_path(
    p_jsonb jsonb,
    p_path text,
    p_value jsonb
)
RETURNS jsonb
LANGUAGE plpgsql
IMMUTABLE
AS $$
DECLARE
    v_parts TEXT[];
    v_result JSONB;
BEGIN
    IF p_path IS NULL OR p_path = '' THEN
        RETURN p_jsonb;
    END IF;

    v_parts := string_to_array(p_path, '.');
    v_result := p_jsonb;

    -- Build the nested structure from the value up
    FOR i IN array_length(v_parts, 1) .. 1 BY -1 LOOP
        v_result := jsonb_build_object(v_parts[i],
            CASE
                WHEN i = array_length(v_parts, 1) THEN p_value
                ELSE v_result
            END
        );
    END LOOP;

    -- Deep merge with existing JSONB
    RETURN COALESCE(p_jsonb, '{}') || v_result;
END;
$$;

--
-- Name: set_jsonb_path(jsonb, text, jsonb); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION set_jsonb_path(jsonb, text, jsonb) IS 'Sets a nested value in JSONB using a dot-separated path. Deep merges with existing structure.';

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
-- Name: is_instance_admin(uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION is_instance_admin(
    p_user_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
SET search_path = platform
AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN false;
    END IF;

    RETURN EXISTS (
        SELECT 1 FROM users
        WHERE id = p_user_id
        AND role = 'instance_admin'
        AND deleted_at IS NULL
        AND is_active = true
    );
END;
$$;

COMMENT ON FUNCTION is_instance_admin(uuid) IS 'Checks if a user is an instance admin. Returns true if the user has the instance_admin role and is active.';

--
-- Name: user_managed_tenant_ids(uuid); Type: FUNCTION; Schema: -; Owner: -
--

CREATE OR REPLACE FUNCTION user_managed_tenant_ids(
    p_user_id uuid
)
RETURNS uuid[]
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
BEGIN
    IF p_user_id IS NULL THEN
        RETURN '{}'::UUID[];
    END IF;

    IF EXISTS (
        SELECT 1 FROM users
        WHERE id = p_user_id
        AND role = 'instance_admin'
        AND deleted_at IS NULL
        AND is_active = true
    ) THEN
        RETURN ARRAY(
            SELECT id FROM tenants WHERE deleted_at IS NULL
        );
    END IF;

    RETURN ARRAY(
        SELECT tenant_id
        FROM tenant_admin_assignments
        WHERE user_id = p_user_id
    );
END;
$$;

--
-- Name: user_managed_tenant_ids(uuid); Type: FUNCTION; Schema: -; Owner: -
--

COMMENT ON FUNCTION user_managed_tenant_ids(uuid) IS 'Returns array of tenant IDs that a platform user can manage. Instance admins get all tenants; others get only their assigned tenants. SECURITY DEFINER to bypass RLS.';

--
-- Name: instance_settings_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER instance_settings_updated_at
    BEFORE UPDATE ON instance_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--
-- Name: platform_tenants_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER platform_tenants_updated_at
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--
-- Name: tenant_settings_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER tenant_settings_updated_at
    BEFORE UPDATE ON tenant_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

--
-- Name: trigger_update_sso_identities_updated_at; Type: TRIGGER; Schema: -; Owner: -
--

CREATE OR REPLACE TRIGGER trigger_update_sso_identities_updated_at
    BEFORE UPDATE ON sso_identities
    FOR EACH ROW
    EXECUTE FUNCTION update_sso_identities_updated_at();

--
-- Name: delete_jsonb_path(p_jsonb jsonb, p_path text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION delete_jsonb_path(p_jsonb jsonb, p_path text) TO tenant_service;

--
-- Name: get_all_settings(p_tenant_id uuid); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_all_settings(p_tenant_id uuid) TO tenant_service;

--
-- Name: get_jsonb_path(p_jsonb jsonb, p_path text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_jsonb_path(p_jsonb jsonb, p_path text) TO tenant_service;

--
-- Name: get_setting(p_tenant_id uuid, p_setting_path text, p_default jsonb); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION get_setting(p_tenant_id uuid, p_setting_path text, p_default jsonb) TO tenant_service;

--
-- Name: is_setting_overridable(setting_path text); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION is_setting_overridable(setting_path text) TO tenant_service;

--
-- Name: set_jsonb_path(p_jsonb jsonb, p_path text, p_value jsonb); Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT EXECUTE ON FUNCTION set_jsonb_path(p_jsonb jsonb, p_path text, p_value jsonb) TO tenant_service;

--
-- Name: activity_log; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE activity_log TO authenticated;

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

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE available_extensions TO service_role;

--
-- Name: email_templates; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE email_templates TO authenticated;

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

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE email_verification_tokens TO service_role;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE enabled_extensions TO authenticated;

--
-- Name: enabled_extensions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE enabled_extensions TO service_role;

--
-- Name: instance_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE instance_settings TO authenticated;

--
-- Name: instance_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE instance_settings TO service_role;

--
-- Name: instance_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT SELECT ON TABLE instance_settings TO tenant_service;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE invitation_tokens TO authenticated;

--
-- Name: invitation_tokens; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE invitation_tokens TO service_role;

--
-- Name: key_usage; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE key_usage TO authenticated;

--
-- Name: key_usage; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE key_usage TO service_role;

--
-- Name: oauth_providers; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE oauth_providers TO authenticated;

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

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE password_reset_tokens TO service_role;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE schema_migrations TO authenticated;

--
-- Name: schema_migrations; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE schema_migrations TO service_role;

--
-- Name: service_keys; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE service_keys TO authenticated;

--
-- Name: service_keys; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE service_keys TO service_role;

--
-- Name: sessions; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE sessions TO authenticated;

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

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE sso_identities TO service_role;

--
-- Name: tenant_admin_assignments; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE tenant_admin_assignments TO authenticated;

--
-- Name: tenant_admin_assignments; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE tenant_admin_assignments TO service_role;

--
-- Name: tenant_memberships; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE tenant_memberships TO service_role;

--
-- Name: tenant_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE tenant_settings TO authenticated;

--
-- Name: tenant_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE tenant_settings TO service_role;

--
-- Name: tenant_settings; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE tenant_settings TO tenant_service;

--
-- Name: tenants; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE tenants TO service_role;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, SELECT, UPDATE ON TABLE users TO authenticated;

--
-- Name: users; Type: PRIVILEGE; Schema: privileges; Owner: -
--

GRANT DELETE, INSERT, MAINTAIN, REFERENCES, SELECT, TRIGGER, TRUNCATE, UPDATE ON TABLE users TO service_role;

