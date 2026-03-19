--
-- TENANT SERVICE RLS POLICIES (ROLLBACK)
-- Remove RLS policies for tenant_service role from all tenant-scoped tables
--

-- ============================================
-- AUTH SCHEMA
-- ============================================

DROP POLICY IF EXISTS tenant_service_users ON auth.users;
DROP POLICY IF EXISTS tenant_service_webhooks ON auth.webhooks;

-- ============================================
-- STORAGE SCHEMA
-- ============================================

DROP POLICY IF EXISTS tenant_service_buckets ON storage.buckets;
DROP POLICY IF EXISTS tenant_service_objects ON storage.objects;
DROP POLICY IF EXISTS tenant_service_object_permissions ON storage.object_permissions;

-- ============================================
-- FUNCTIONS SCHEMA
-- ============================================

DROP POLICY IF EXISTS tenant_service_edge_functions ON functions.edge_functions;
DROP POLICY IF EXISTS tenant_service_secrets ON functions.secrets;
DROP POLICY IF EXISTS tenant_service_edge_triggers ON functions.edge_triggers;

-- ============================================
-- JOBS SCHEMA
-- ============================================

DROP POLICY IF EXISTS tenant_service_jobs_functions ON jobs.functions;
DROP POLICY IF EXISTS tenant_service_jobs_queue ON jobs.queue;

-- ============================================
-- AI SCHEMA
-- ============================================

DROP POLICY IF EXISTS tenant_service_knowledge_bases ON ai.knowledge_bases;
DROP POLICY IF EXISTS tenant_service_documents ON ai.documents;
DROP POLICY IF EXISTS tenant_service_chatbots ON ai.chatbots;
DROP POLICY IF EXISTS tenant_service_conversations ON ai.conversations;
DROP POLICY IF EXISTS tenant_service_user_chatbot_usage ON ai.user_chatbot_usage;
DROP POLICY IF EXISTS tenant_service_user_provider_preferences ON ai.user_provider_preferences;
DROP POLICY IF EXISTS tenant_service_user_quotas ON ai.user_quotas;
DROP POLICY IF EXISTS tenant_service_chatbot_knowledge_bases ON ai.chatbot_knowledge_bases;
DROP POLICY IF EXISTS tenant_service_document_permissions ON ai.document_permissions;
DROP POLICY IF EXISTS tenant_service_knowledge_base_permissions ON ai.knowledge_base_permissions;
DROP POLICY IF EXISTS tenant_service_table_export_sync_configs ON ai.table_export_sync_configs;
