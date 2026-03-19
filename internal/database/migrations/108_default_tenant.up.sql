--
-- MULTI-TENANCY: CREATE DEFAULT TENANT AND BACKFILL DATA
-- Creates the default tenant with a fixed UUID and backfills all tenant_id columns
--

DO $$
DECLARE
    default_tenant_id UUID := '00000000-0000-0000-0000-000000000000'::UUID;
BEGIN
    -- Create default tenant if it doesn't exist (check both id and slug)
    INSERT INTO platform.tenants (id, slug, name, is_default, metadata, created_at)
    SELECT
        default_tenant_id,
        'default',
        'Default',
        true,
        '{"description": "Default tenant for backward compatibility"}'::jsonb,
        NOW()
    WHERE NOT EXISTS (
        SELECT 1 FROM platform.tenants WHERE id = default_tenant_id OR slug = 'default'
    );

    RAISE NOTICE 'Default tenant ensured with ID: %', default_tenant_id;

    -- ============================================
    -- BACKFILL TENANT_ID ON ALL TABLES
    -- ============================================

    -- Auth schema
    UPDATE auth.users SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE auth.webhooks SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Auth.client_keys (may have been removed in migration 105)
    BEGIN
        UPDATE auth.client_keys SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    EXCEPTION WHEN undefined_table THEN
        RAISE NOTICE 'auth.client_keys table does not exist, skipping';
    END;

    -- Storage schema
    UPDATE storage.buckets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE storage.objects SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Storage.object_permissions (may not exist in all installations)
    BEGIN
        UPDATE storage.object_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    EXCEPTION WHEN undefined_table THEN
        RAISE NOTICE 'storage.object_permissions table does not exist, skipping';
    END;

    -- Functions schema
    UPDATE functions.edge_functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE functions.secrets SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- Functions.edge_triggers (may not exist in all installations)
    BEGIN
        UPDATE functions.edge_triggers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    EXCEPTION WHEN undefined_table THEN
        RAISE NOTICE 'functions.edge_triggers table does not exist, skipping';
    END;

    -- Jobs schema
    UPDATE jobs.functions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE jobs.queue SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    -- AI schema
    UPDATE ai.knowledge_bases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.documents SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.chatbots SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.conversations SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_chatbot_usage SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_provider_preferences SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.user_quotas SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.chatbot_knowledge_bases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.document_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.knowledge_base_permissions SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
    UPDATE ai.table_export_sync_configs SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;

    RAISE NOTICE 'Tenant ID backfill complete';
END $$;
