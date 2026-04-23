--
-- Post-Schema Cross-Schema Policies
--
-- This file contains RLS policies that reference tables/functions in other schemas.
-- It is applied AFTER all schema files have been applied, allowing cross-schema references.
--

-- ============================================================================
-- PLATFORM SCHEMA POLICIES (reference auth.users and auth.uid())
-- ============================================================================

--
-- Name: platform_tenants_instance_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenants_instance_admin ON platform.tenants TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: instance_settings_insert; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY instance_settings_insert ON platform.instance_settings FOR INSERT TO PUBLIC WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: instance_settings_select; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY instance_settings_select ON platform.instance_settings FOR SELECT TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (CURRENT_USER = 'tenant_service'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR ((current_setting('app.current_tenant_id', true) IS NOT NULL) AND (current_setting('app.current_tenant_id', true) <> ''::text) AND (tenant_id IS NULL OR tenant_id::text = current_setting('app.current_tenant_id', true))) OR (tenant_id IS NOT NULL AND auth.has_tenant_access(tenant_id)));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: instance_settings_update; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY instance_settings_update ON platform.instance_settings FOR UPDATE TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR (tenant_id IS NOT NULL AND auth.has_tenant_access(tenant_id))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR (tenant_id IS NOT NULL AND auth.has_tenant_access(tenant_id)));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: instance_settings_insert_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY instance_settings_insert_tenant ON platform.instance_settings FOR INSERT TO PUBLIC WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR (tenant_id IS NOT NULL AND auth.has_tenant_access(tenant_id)));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: instance_settings_delete_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY instance_settings_delete_tenant ON platform.instance_settings FOR DELETE TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR (tenant_id IS NOT NULL AND auth.has_tenant_access(tenant_id)));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_users_all; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_users_all ON platform.users TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_users_self; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_users_self ON platform.users FOR SELECT TO authenticated USING (id = auth.uid());
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_sessions_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_sessions_admin ON platform.sessions TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id))))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: sso_identities_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY sso_identities_admin ON platform.sso_identities TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id))))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((current_setting('request.jwt.claims', true) <> '' AND users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_admin_assignments_all; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_admin_assignments_all ON platform.tenant_admin_assignments TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_admin_assignments_self; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_admin_assignments_self ON platform.tenant_admin_assignments FOR SELECT TO authenticated USING (platform.is_instance_admin(auth.uid()) OR (user_id = auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_admin_assignments_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_admin_assignments_tenant ON platform.tenant_admin_assignments TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenants_assigned; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenants_assigned ON platform.tenants FOR SELECT TO authenticated USING (platform.is_instance_admin(auth.uid()) OR (EXISTS ( SELECT 1 FROM platform.tenant_admin_assignments taa WHERE ((taa.tenant_id = tenants.id) AND (taa.user_id = auth.uid())))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_oauth_providers_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_oauth_providers_tenant ON platform.oauth_providers TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- PLATFORM TABLES: RLS POLICIES FOR TABLES WITH ENABLE+FORCE BUT NO POLICIES
-- ============================================================================

--
-- Name: platform_password_reset_tokens_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_password_reset_tokens_service ON platform.password_reset_tokens TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_password_reset_tokens_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_password_reset_tokens_admin ON platform.password_reset_tokens TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_email_verification_tokens_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_email_verification_tokens_service ON platform.email_verification_tokens TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_email_verification_tokens_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_email_verification_tokens_admin ON platform.email_verification_tokens TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_activity_log_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_activity_log_service ON platform.activity_log TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_activity_log_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_activity_log_admin ON platform.activity_log TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_invitation_tokens_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_invitation_tokens_service ON platform.invitation_tokens TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_invitation_tokens_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_invitation_tokens_admin ON platform.invitation_tokens TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_invitation_tokens_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_invitation_tokens_tenant ON platform.invitation_tokens TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_memberships_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_memberships_service ON platform.tenant_memberships TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_memberships_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_memberships_admin ON platform.tenant_memberships TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_memberships_self; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_memberships_self ON platform.tenant_memberships FOR SELECT TO authenticated USING (user_id = auth.uid());
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_tenant_memberships_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenant_memberships_tenant ON platform.tenant_memberships TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- PLATFORM TABLES: RLS POLICIES FOR NEWLY PROTECTED TABLES (Group B)
-- ============================================================================

--
-- Name: platform_available_extensions_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_available_extensions_service ON platform.available_extensions TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_available_extensions_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_available_extensions_admin ON platform.available_extensions TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_enabled_extensions_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_enabled_extensions_service ON platform.enabled_extensions TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_enabled_extensions_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_enabled_extensions_admin ON platform.enabled_extensions TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_enabled_extensions_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_enabled_extensions_tenant ON platform.enabled_extensions TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_schema_migrations_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_schema_migrations_service ON platform.schema_migrations TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_service_keys_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_service_keys_service ON platform.service_keys TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_service_keys_admin; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_service_keys_admin ON platform.service_keys TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_service_keys_tenant; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_service_keys_tenant ON platform.service_keys TO PUBLIC USING (auth.has_tenant_access(tenant_id)) WITH CHECK (auth.has_tenant_access(tenant_id));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: platform_key_usage_service; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_key_usage_service ON platform.key_usage TO PUBLIC USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- AI SCHEMA POLICIES (reference auth.users)
-- ============================================================================

--
-- Name: entities_admin_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY entities_admin_all ON ai.entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: entities_service_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY entities_service_all ON ai.entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: document_entities_admin_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY document_entities_admin_all ON ai.document_entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: document_entities_service_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY document_entities_service_all ON ai.document_entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: relationships_admin_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY relationships_admin_all ON ai.entity_relationships TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: relationships_service_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY relationships_service_all ON ai.entity_relationships TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: chatbot_kb_links_admin_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY chatbot_kb_links_admin_all ON ai.chatbot_knowledge_bases TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: chatbot_kb_links_service_all; Type: POLICY; Schema: ai; Owner: -
--

DO $$ BEGIN
    CREATE POLICY chatbot_kb_links_service_all ON ai.chatbot_knowledge_bases TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- MCP SCHEMA POLICIES (custom_resources and custom_tools)
-- ============================================================================

--
-- Name: mcp_custom_resources_service; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_resources_service ON mcp.custom_resources TO authenticated USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_resources_admin; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_resources_admin ON mcp.custom_resources TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_resources_owner; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_resources_owner ON mcp.custom_resources TO authenticated USING (created_by = auth.current_user_id()) WITH CHECK (created_by = auth.current_user_id());
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_resources_authenticated_read; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_resources_authenticated_read ON mcp.custom_resources FOR SELECT TO authenticated USING (enabled = true);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_tools_service; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_tools_service ON mcp.custom_tools TO authenticated USING (auth.current_user_role() = 'service_role') WITH CHECK (auth.current_user_role() = 'service_role');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_tools_admin; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_tools_admin ON mcp.custom_tools TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_tools_owner; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_tools_owner ON mcp.custom_tools TO authenticated USING (created_by = auth.current_user_id()) WITH CHECK (created_by = auth.current_user_id());
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

--
-- Name: mcp_custom_tools_authenticated_read; Type: POLICY; Schema: mcp; Owner: -
--

DO $$ BEGIN
    CREATE POLICY mcp_custom_tools_authenticated_read ON mcp.custom_tools FOR SELECT TO authenticated USING (enabled = true);
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- ============================================================================
-- LOGGING TRIGGER + PARTITION ATTACHMENT
-- logging.entries is PARTITION BY LIST (category). The child tables are created
-- in logging.sql but ATTACH PARTITION must run via this file because pgschema's
-- plan path skips DO $$ blocks. The trigger on the parent (logging_entries_set_tenant_id)
-- auto-propagates to partitions on attach, so we drop any pre-existing trigger on
-- each child table before attaching to avoid "trigger already exists" errors.
--
-- The trigger itself is also defined here (not in logging.sql) because pgschema
-- detects the auto-propagated triggers on partition children and tries to DROP them,
-- which PostgreSQL forbids (inherited triggers cannot be dropped from partitions).
-- ============================================================================

DO $$
BEGIN
    EXECUTE 'ALTER TABLE logging.entries ENABLE ROW LEVEL SECURITY';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'logging.entries ENABLE RLS: %', SQLERRM;
END $$;

DO $$
BEGIN
    EXECUTE 'ALTER TABLE logging.entries FORCE ROW LEVEL SECURITY';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'logging.entries FORCE RLS: %', SQLERRM;
END $$;

DO $$
BEGIN
    EXECUTE 'CREATE OR REPLACE TRIGGER logging_entries_set_tenant_id
        BEFORE INSERT ON logging.entries
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context()';
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'logging_entries_set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_ai' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_ai';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_ai FOR VALUES IN (''ai'')';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_custom' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_custom';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_custom FOR VALUES IN (''custom'')';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_execution' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_execution';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_execution FOR VALUES IN (''execution'')';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_http' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_http';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_http FOR VALUES IN (''http'')';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_security' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_security';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_security FOR VALUES IN (''security'')';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_class c
        JOIN pg_inherits i ON c.oid = i.inhrelid
        WHERE c.relname = 'entries_system' AND i.inhparent = 'logging.entries'::regclass
    ) THEN
        EXECUTE 'DROP TRIGGER IF EXISTS logging_entries_set_tenant_id ON logging.entries_system';
        EXECUTE 'ALTER TABLE logging.entries ATTACH PARTITION logging.entries_system FOR VALUES IN (''system'')';
    END IF;
END $$;

-- ============================================================================
-- PLATFORM SCHEMA TRIGGERS (auto-populate tenant_id)
-- ============================================================================

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_instance_settings_set_tenant_id
        BEFORE INSERT ON platform.instance_settings
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.instance_settings set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_service_keys_set_tenant_id
        BEFORE INSERT ON platform.service_keys
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.service_keys set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_tenant_memberships_set_tenant_id
        BEFORE INSERT ON platform.tenant_memberships
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.tenant_memberships set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_enabled_extensions_set_tenant_id
        BEFORE INSERT ON platform.enabled_extensions
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.enabled_extensions set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_invitation_tokens_set_tenant_id
        BEFORE INSERT ON platform.invitation_tokens
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.invitation_tokens set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_oauth_providers_set_tenant_id
        BEFORE INSERT ON platform.oauth_providers
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.oauth_providers set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER platform_tenant_admin_assignments_set_tenant_id
        BEFORE INSERT ON platform.tenant_admin_assignments
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'platform.tenant_admin_assignments set_tenant_id trigger: %', SQLERRM;
END $$;

-- ============================================================================
-- AUTH SCHEMA TRIGGERS (auto-populate tenant_id for webhook tables)
-- ============================================================================

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_webhook_deliveries_set_tenant_id
        BEFORE INSERT ON auth.webhook_deliveries
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'auth.webhook_deliveries set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_webhook_events_set_tenant_id
        BEFORE INSERT ON auth.webhook_events
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'auth.webhook_events set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_saml_providers_set_tenant_id
        BEFORE INSERT ON auth.saml_providers
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'auth.saml_providers set_tenant_id trigger: %', SQLERRM;
END $$;

-- ============================================================================
-- APP SCHEMA TRIGGERS (auto-populate tenant_id)
-- ============================================================================

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER app_settings_set_tenant_id
        BEFORE INSERT ON app.settings
        FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_context();
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'app.settings set_tenant_id trigger: %', SQLERRM;
END $$;

-- ============================================================================
-- AUTH SCHEMA TRIGGERS — user-derived tenant_id (set_tenant_id_from_user_or_context)
-- ============================================================================

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_sessions_set_tenant_id
        BEFORE INSERT ON auth.sessions FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.sessions set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_oauth_links_set_tenant_id
        BEFORE INSERT ON auth.oauth_links FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.oauth_links set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_oauth_tokens_set_tenant_id
        BEFORE INSERT ON auth.oauth_tokens FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.oauth_tokens set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_mfa_factors_set_tenant_id
        BEFORE INSERT ON auth.mfa_factors FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.mfa_factors set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_saml_sessions_set_tenant_id
        BEFORE INSERT ON auth.saml_sessions FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.saml_sessions set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_magic_links_set_tenant_id
        BEFORE INSERT ON auth.magic_links FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.magic_links set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_otp_codes_set_tenant_id
        BEFORE INSERT ON auth.otp_codes FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.otp_codes set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_email_verification_tokens_set_tenant_id
        BEFORE INSERT ON auth.email_verification_tokens FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.email_verification_tokens set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_password_reset_tokens_set_tenant_id
        BEFORE INSERT ON auth.password_reset_tokens FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.password_reset_tokens set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_two_factor_setups_set_tenant_id
        BEFORE INSERT ON auth.two_factor_setups FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.two_factor_setups set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_two_factor_recovery_set_tenant_id
        BEFORE INSERT ON auth.two_factor_recovery_attempts FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.two_factor_recovery_attempts set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_oauth_logout_states_set_tenant_id
        BEFORE INSERT ON auth.oauth_logout_states FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.oauth_logout_states set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_mcp_oauth_clients_set_tenant_id
        BEFORE INSERT ON auth.mcp_oauth_clients FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.mcp_oauth_clients set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_mcp_oauth_codes_set_tenant_id
        BEFORE INSERT ON auth.mcp_oauth_codes FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.mcp_oauth_codes set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_mcp_oauth_tokens_set_tenant_id
        BEFORE INSERT ON auth.mcp_oauth_tokens FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_user_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.mcp_oauth_tokens set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_client_key_usage_set_tenant_id
        BEFORE INSERT ON auth.client_key_usage FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_client_key_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.client_key_usage set_tenant_id trigger: %', SQLERRM;
END $$;

DO $$ BEGIN
    CREATE OR REPLACE TRIGGER auth_service_key_revocations_set_tenant_id
        BEFORE INSERT ON auth.service_key_revocations FOR EACH ROW
        EXECUTE FUNCTION auth.set_tenant_id_from_service_key_or_context();
EXCEPTION WHEN OTHERS THEN RAISE NOTICE 'auth.service_key_revocations set_tenant_id trigger: %', SQLERRM;
END $$;

-- ============================================================================
-- DATA MIGRATION: assign existing NULL tenant_id rows to the default tenant
-- These UPDATEs are idempotent (WHERE tenant_id IS NULL) and become no-ops
-- after the first successful run.
-- ============================================================================

DO $$
DECLARE
    default_tenant_uuid UUID;
BEGIN
    SELECT id INTO default_tenant_uuid FROM platform.tenants WHERE is_default = true LIMIT 1;
    IF default_tenant_uuid IS NULL THEN
        RAISE NOTICE 'No default tenant found, skipping NULL tenant_id migration';
        RETURN;
    END IF;

    -- Logging entries (parent + all partitions)
    UPDATE logging.entries SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;

    -- Auth tables with tenant RLS
    UPDATE auth.webhooks SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.webhook_deliveries SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.webhook_events SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.saml_providers SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.client_keys SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.impersonation_sessions SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;

    -- Auth tables — derive tenant_id from user_id first, fallback to default
    UPDATE auth.sessions SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.sessions.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.oauth_links SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.oauth_links.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.oauth_tokens SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.oauth_tokens.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.mfa_factors SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.mfa_factors.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.saml_sessions SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.saml_sessions.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.email_verification_tokens SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.email_verification_tokens.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.password_reset_tokens SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.password_reset_tokens.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.two_factor_setups SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.two_factor_setups.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.two_factor_recovery_attempts SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.two_factor_recovery_attempts.user_id) WHERE tenant_id IS NULL;
    UPDATE auth.oauth_logout_states SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.oauth_logout_states.user_id) WHERE tenant_id IS NULL;

    -- Auth tables — derive from email/registered_by/key FK
    UPDATE auth.magic_links SET user_id = (SELECT id FROM auth.users WHERE email = auth.magic_links.email LIMIT 1) WHERE user_id IS NULL;
    UPDATE auth.magic_links SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.magic_links.user_id) WHERE tenant_id IS NULL AND user_id IS NOT NULL;
    UPDATE auth.otp_codes SET user_id = (SELECT id FROM auth.users WHERE email = auth.otp_codes.email LIMIT 1) WHERE user_id IS NULL AND email IS NOT NULL;
    UPDATE auth.otp_codes SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.otp_codes.user_id) WHERE tenant_id IS NULL AND user_id IS NOT NULL;
    UPDATE auth.mcp_oauth_clients SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.mcp_oauth_clients.registered_by) WHERE tenant_id IS NULL AND registered_by IS NOT NULL;
    UPDATE auth.mcp_oauth_codes SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.mcp_oauth_codes.user_id) WHERE tenant_id IS NULL AND user_id IS NOT NULL;
    UPDATE auth.mcp_oauth_tokens SET tenant_id = (SELECT tenant_id FROM auth.users WHERE id = auth.mcp_oauth_tokens.user_id) WHERE tenant_id IS NULL AND user_id IS NOT NULL;
    UPDATE auth.client_key_usage SET tenant_id = (SELECT tenant_id FROM auth.client_keys WHERE id = auth.client_key_usage.client_key_id) WHERE tenant_id IS NULL;
    UPDATE auth.service_key_revocations SET tenant_id = (SELECT tenant_id FROM auth.service_keys WHERE id = auth.service_key_revocations.key_id) WHERE tenant_id IS NULL;

    -- Fallback: any remaining NULL tenant_id → default tenant
    UPDATE auth.sessions SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.oauth_links SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.oauth_tokens SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.mfa_factors SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.saml_sessions SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.magic_links SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.otp_codes SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.email_verification_tokens SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.password_reset_tokens SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.two_factor_setups SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.two_factor_recovery_attempts SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.oauth_logout_states SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.mcp_oauth_clients SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.mcp_oauth_codes SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.mcp_oauth_tokens SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.client_key_usage SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE auth.service_key_revocations SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;

    -- Platform tables with tenant RLS
    UPDATE platform.oauth_providers SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.tenant_admin_assignments SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.instance_settings SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.service_keys SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.tenant_memberships SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.enabled_extensions SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;
    UPDATE platform.invitation_tokens SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;

    -- App tables with tenant RLS
    UPDATE app.settings SET tenant_id = default_tenant_uuid WHERE tenant_id IS NULL;

    RAISE NOTICE 'Migrated NULL tenant_id rows to default tenant %', default_tenant_uuid;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'NULL tenant_id migration skipped: %', SQLERRM;
END $$;
