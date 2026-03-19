--
-- MULTI-TENANCY: ROLLBACK DEFAULT TENANT CREATION AND BACKFILL
--

DO $$
DECLARE
    default_tenant_id UUID := '00000000-0000-0000-0000-000000000000'::UUID;
BEGIN
    -- ============================================
    -- CLEAR TENANT_ID VALUES (set to NULL)
    -- ============================================
    
    -- AI Schema
    UPDATE ai.table_export_sync_configs SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.knowledge_base_permissions SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.document_permissions SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.chatbot_knowledge_bases SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.user_quotas SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.user_provider_preferences SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.user_chatbot_usage SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.conversations SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.chatbots SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.documents SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE ai.knowledge_bases SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    
    -- Jobs Schema
    UPDATE jobs.queue SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE jobs.functions SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    
    -- Functions Schema
    BEGIN
        UPDATE functions.edge_triggers SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    EXCEPTION WHEN undefined_table THEN
        NULL;
    END;
    UPDATE functions.secrets SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE functions.edge_functions SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    
    -- Storage Schema
    BEGIN
        UPDATE storage.object_permissions SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    EXCEPTION WHEN undefined_table THEN
        NULL;
    END;
    UPDATE storage.objects SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE storage.buckets SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    
    -- Auth Schema
    BEGIN
        UPDATE auth.client_keys SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    EXCEPTION WHEN undefined_table THEN
        NULL;
    END;
    UPDATE auth.webhooks SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    UPDATE auth.users SET tenant_id = NULL WHERE tenant_id = default_tenant_id;
    
    RAISE NOTICE 'Tenant ID values cleared';
END $$;
