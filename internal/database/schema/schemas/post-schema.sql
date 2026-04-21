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
-- Name: platform_tenants_assigned; Type: POLICY; Schema: platform; Owner: -
--

DO $$ BEGIN
    CREATE POLICY platform_tenants_assigned ON platform.tenants FOR SELECT TO authenticated USING (platform.is_instance_admin(auth.uid()) OR (EXISTS ( SELECT 1 FROM platform.tenant_admin_assignments taa WHERE ((taa.tenant_id = tenants.id) AND (taa.user_id = auth.uid())))));
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
-- LOGGING PARTITION ATTACHMENT
-- logging.entries is PARTITION BY LIST (category). The child tables are created
-- in logging.sql but ATTACH PARTITION must run via this file because pgschema's
-- plan path skips DO $$ blocks. The trigger on the parent (logging_entries_set_tenant_id)
-- auto-propagates to partitions on attach, so we drop any pre-existing trigger on
-- each child table before attaching to avoid "trigger already exists" errors.
-- ============================================================================

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
