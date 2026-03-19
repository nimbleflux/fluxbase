--
-- TENANT SERVICE RLS POLICIES
-- Add RLS policies for tenant_service role on all tenant-scoped tables
--

-- ============================================
-- AUTH SCHEMA
-- ============================================

-- auth.users
CREATE POLICY tenant_service_users ON auth.users
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- auth.webhooks
CREATE POLICY tenant_service_webhooks ON auth.webhooks
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ============================================
-- STORAGE SCHEMA
-- ============================================

-- storage.buckets
CREATE POLICY tenant_service_buckets ON storage.buckets
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- storage.objects
CREATE POLICY tenant_service_objects ON storage.objects
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- storage.object_permissions
CREATE POLICY tenant_service_object_permissions ON storage.object_permissions
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ============================================
-- FUNCTIONS SCHEMA
-- ============================================

-- functions.edge_functions
CREATE POLICY tenant_service_edge_functions ON functions.edge_functions
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- functions.secrets
CREATE POLICY tenant_service_secrets ON functions.secrets
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- functions.edge_triggers
CREATE POLICY tenant_service_edge_triggers ON functions.edge_triggers
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ============================================
-- JOBS SCHEMA
-- ============================================

-- jobs.functions
CREATE POLICY tenant_service_jobs_functions ON jobs.functions
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- jobs.queue
CREATE POLICY tenant_service_jobs_queue ON jobs.queue
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ============================================
-- AI SCHEMA
-- ============================================

-- ai.knowledge_bases
CREATE POLICY tenant_service_knowledge_bases ON ai.knowledge_bases
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.documents
CREATE POLICY tenant_service_documents ON ai.documents
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.chatbots
CREATE POLICY tenant_service_chatbots ON ai.chatbots
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.conversations
CREATE POLICY tenant_service_conversations ON ai.conversations
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.user_chatbot_usage
CREATE POLICY tenant_service_user_chatbot_usage ON ai.user_chatbot_usage
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.user_provider_preferences
CREATE POLICY tenant_service_user_provider_preferences ON ai.user_provider_preferences
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.user_quotas
CREATE POLICY tenant_service_user_quotas ON ai.user_quotas
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.chatbot_knowledge_bases
CREATE POLICY tenant_service_chatbot_knowledge_bases ON ai.chatbot_knowledge_bases
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.document_permissions
CREATE POLICY tenant_service_document_permissions ON ai.document_permissions
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.knowledge_base_permissions
CREATE POLICY tenant_service_knowledge_base_permissions ON ai.knowledge_base_permissions
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());

-- ai.table_export_sync_configs
CREATE POLICY tenant_service_table_export_sync_configs ON ai.table_export_sync_configs
    FOR ALL TO tenant_service
    USING (tenant_id = current_tenant_id())
    WITH CHECK (tenant_id = current_tenant_id());
