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

CREATE POLICY platform_tenants_instance_admin ON platform.tenants TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));

--
-- Name: instance_settings_insert; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY instance_settings_insert ON platform.instance_settings FOR INSERT TO PUBLIC WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));

--
-- Name: instance_settings_select; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY instance_settings_select ON platform.instance_settings FOR SELECT TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (CURRENT_USER = 'tenant_service'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR ((current_setting('app.current_tenant_id', true) IS NOT NULL) AND (current_setting('app.current_tenant_id', true) <> ''::text)));

--
-- Name: instance_settings_update; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY instance_settings_update ON platform.instance_settings FOR UPDATE TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id))))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));

--
-- Name: tenant_settings_delete; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY tenant_settings_delete ON platform.tenant_settings FOR DELETE TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR auth.has_tenant_access(tenant_id));

--
-- Name: tenant_settings_insert; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY tenant_settings_insert ON platform.tenant_settings FOR INSERT TO PUBLIC WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR auth.has_tenant_access(tenant_id));

--
-- Name: tenant_settings_select; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY tenant_settings_select ON platform.tenant_settings FOR SELECT TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR auth.has_tenant_access(tenant_id));

--
-- Name: tenant_settings_update; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY tenant_settings_update ON platform.tenant_settings FOR UPDATE TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR auth.has_tenant_access(tenant_id)) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))) OR auth.has_tenant_access(tenant_id));

--
-- Name: platform_users_all; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_users_all ON platform.users TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));

--
-- Name: platform_users_self; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_users_self ON platform.users FOR SELECT TO authenticated USING (id = auth.uid());

--
-- Name: platform_sessions_admin; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_sessions_admin ON platform.sessions TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id))))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));

--
-- Name: sso_identities_admin; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY sso_identities_admin ON platform.sso_identities TO PUBLIC USING ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id))))) WITH CHECK ((CURRENT_USER = 'service_role'::name) OR (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = ((current_setting('request.jwt.claims', true)::jsonb ->> 'sub'))::uuid) AND platform.is_instance_admin(users.id)))));

--
-- Name: platform_tenant_admin_assignments_all; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_tenant_admin_assignments_all ON platform.tenant_admin_assignments TO authenticated USING (platform.is_instance_admin(auth.uid())) WITH CHECK (platform.is_instance_admin(auth.uid()));

--
-- Name: platform_tenant_admin_assignments_self; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_tenant_admin_assignments_self ON platform.tenant_admin_assignments FOR SELECT TO authenticated USING (platform.is_instance_admin(auth.uid()) OR (user_id = auth.uid()));

--
-- Name: platform_tenants_assigned; Type: POLICY; Schema: platform; Owner: -
--

CREATE POLICY platform_tenants_assigned ON platform.tenants FOR SELECT TO authenticated USING (platform.is_instance_admin(auth.uid()) OR (EXISTS ( SELECT 1 FROM platform.tenant_admin_assignments taa WHERE ((taa.tenant_id = tenants.id) AND (taa.user_id = auth.uid())))));

-- ============================================================================
-- AI SCHEMA POLICIES (reference auth.users)
-- ============================================================================

--
-- Name: entities_admin_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY entities_admin_all ON ai.entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));

--
-- Name: entities_service_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY entities_service_all ON ai.entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));

--
-- Name: document_entities_admin_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY document_entities_admin_all ON ai.document_entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));

--
-- Name: document_entities_service_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY document_entities_service_all ON ai.document_entities TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));

--
-- Name: relationships_admin_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY relationships_admin_all ON ai.entity_relationships TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));

--
-- Name: relationships_service_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY relationships_service_all ON ai.entity_relationships TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));

--
-- Name: chatbot_kb_links_admin_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY chatbot_kb_links_admin_all ON ai.chatbot_knowledge_bases TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'admin'))));

--
-- Name: chatbot_kb_links_service_all; Type: POLICY; Schema: ai; Owner: -
--

CREATE POLICY chatbot_kb_links_service_all ON ai.chatbot_knowledge_bases TO authenticated USING (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role')))) WITH CHECK (EXISTS ( SELECT 1 FROM auth.users WHERE ((users.id = auth.current_user_id()) AND (users.role = 'service_role'))));
