--
-- Cross-schema foreign key constraints
-- Applied after all schemas exist to avoid pgschema validation issues
-- Uses idempotent DO blocks to safely run on every startup
--

-- ============================================================================
-- platform (intra-schema forward references)
-- ============================================================================

-- platform.instance_settings.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'instance_settings_tenant_id_fkey'
        AND conrelid = 'platform.instance_settings'::regclass
    ) THEN
        ALTER TABLE platform.instance_settings ADD CONSTRAINT instance_settings_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- auth -> platform
-- ============================================================================

-- auth.users.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_auth_users_tenant'
        AND conrelid = 'auth.users'::regclass
    ) THEN
        ALTER TABLE auth.users ADD CONSTRAINT fk_auth_users_tenant
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE SET NULL DEFERRABLE;
    END IF;
END $$;

-- auth.mcp_oauth_codes.user_id -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'mcp_oauth_codes_user_id_fkey'
        AND conrelid = 'auth.mcp_oauth_codes'::regclass
    ) THEN
        ALTER TABLE auth.mcp_oauth_codes ADD CONSTRAINT mcp_oauth_codes_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES platform.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- auth.mcp_oauth_tokens.user_id -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'mcp_oauth_tokens_user_id_fkey'
        AND conrelid = 'auth.mcp_oauth_tokens'::regclass
    ) THEN
        ALTER TABLE auth.mcp_oauth_tokens ADD CONSTRAINT mcp_oauth_tokens_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES platform.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- auth.service_keys.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'fk_auth_service_keys_tenant'
        AND conrelid = 'auth.service_keys'::regclass
    ) THEN
        ALTER TABLE auth.service_keys ADD CONSTRAINT fk_auth_service_keys_tenant
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE SET NULL DEFERRABLE;
    END IF;
END $$;

-- auth.service_keys.revoked_by -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'service_keys_revoked_by_fkey'
        AND conrelid = 'auth.service_keys'::regclass
    ) THEN
        ALTER TABLE auth.service_keys ADD CONSTRAINT service_keys_revoked_by_fkey
            FOREIGN KEY (revoked_by) REFERENCES platform.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- auth.service_key_revocations.revoked_by -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'service_key_revocations_revoked_by_fkey'
        AND conrelid = 'auth.service_key_revocations'::regclass
    ) THEN
        ALTER TABLE auth.service_key_revocations ADD CONSTRAINT service_key_revocations_revoked_by_fkey
            FOREIGN KEY (revoked_by) REFERENCES platform.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- functions -> platform, auth
-- ============================================================================

-- functions.secrets.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'secrets_tenant_id_fkey'
        AND conrelid = 'functions.secrets'::regclass
    ) THEN
        ALTER TABLE functions.secrets ADD CONSTRAINT secrets_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- functions.secrets.created_by -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'secrets_created_by_fkey'
        AND conrelid = 'functions.secrets'::regclass
    ) THEN
        ALTER TABLE functions.secrets ADD CONSTRAINT secrets_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES platform.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- functions.secrets.updated_by -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'secrets_updated_by_fkey'
        AND conrelid = 'functions.secrets'::regclass
    ) THEN
        ALTER TABLE functions.secrets ADD CONSTRAINT secrets_updated_by_fkey
            FOREIGN KEY (updated_by) REFERENCES platform.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- functions.secret_versions.created_by -> platform.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'secret_versions_created_by_fkey'
        AND conrelid = 'functions.secret_versions'::regclass
    ) THEN
        ALTER TABLE functions.secret_versions ADD CONSTRAINT secret_versions_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES platform.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- functions.shared_modules.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'shared_modules_created_by_fkey'
        AND conrelid = 'functions.shared_modules'::regclass
    ) THEN
        ALTER TABLE functions.shared_modules ADD CONSTRAINT shared_modules_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- storage -> platform, auth
-- ============================================================================

-- storage.buckets.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'buckets_tenant_id_fkey'
        AND conrelid = 'storage.buckets'::regclass
    ) THEN
        ALTER TABLE storage.buckets ADD CONSTRAINT buckets_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- storage.chunked_upload_sessions.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chunked_upload_sessions_tenant_id_fkey'
        AND conrelid = 'storage.chunked_upload_sessions'::regclass
    ) THEN
        ALTER TABLE storage.chunked_upload_sessions ADD CONSTRAINT chunked_upload_sessions_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- storage.objects.owner_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'objects_owner_id_fkey'
        AND conrelid = 'storage.objects'::regclass
    ) THEN
        ALTER TABLE storage.objects ADD CONSTRAINT objects_owner_id_fkey
            FOREIGN KEY (owner_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- storage.objects.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'objects_tenant_id_fkey'
        AND conrelid = 'storage.objects'::regclass
    ) THEN
        ALTER TABLE storage.objects ADD CONSTRAINT objects_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- storage.object_permissions.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'object_permissions_tenant_id_fkey'
        AND conrelid = 'storage.object_permissions'::regclass
    ) THEN
        ALTER TABLE storage.object_permissions ADD CONSTRAINT object_permissions_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- storage.object_permissions.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'object_permissions_user_id_fkey'
        AND conrelid = 'storage.object_permissions'::regclass
    ) THEN
        ALTER TABLE storage.object_permissions ADD CONSTRAINT object_permissions_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- logging -> platform
-- ============================================================================

-- logging.entries_ai.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_ai'::regclass
    ) THEN
        ALTER TABLE logging.entries_ai ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- logging.entries_custom.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_custom'::regclass
    ) THEN
        ALTER TABLE logging.entries_custom ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- logging.entries_execution.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_execution'::regclass
    ) THEN
        ALTER TABLE logging.entries_execution ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- logging.entries_http.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_http'::regclass
    ) THEN
        ALTER TABLE logging.entries_http ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- logging.entries_security.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_security'::regclass
    ) THEN
        ALTER TABLE logging.entries_security ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- logging.entries_system.tenant_id -> platform.tenants.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'entries_tenant_id_fkey'
        AND conrelid = 'logging.entries_system'::regclass
    ) THEN
        ALTER TABLE logging.entries_system ADD CONSTRAINT entries_tenant_id_fkey
            FOREIGN KEY (tenant_id) REFERENCES platform.tenants(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- ai -> auth, storage
-- ============================================================================

-- ai.knowledge_bases.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'knowledge_bases_created_by_fkey'
        AND conrelid = 'ai.knowledge_bases'::regclass
    ) THEN
        ALTER TABLE ai.knowledge_bases ADD CONSTRAINT knowledge_bases_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.knowledge_bases.owner_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'knowledge_bases_owner_id_fkey'
        AND conrelid = 'ai.knowledge_bases'::regclass
    ) THEN
        ALTER TABLE ai.knowledge_bases ADD CONSTRAINT knowledge_bases_owner_id_fkey
            FOREIGN KEY (owner_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.documents.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'documents_created_by_fkey'
        AND conrelid = 'ai.documents'::regclass
    ) THEN
        ALTER TABLE ai.documents ADD CONSTRAINT documents_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.documents.owner_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'documents_owner_id_fkey'
        AND conrelid = 'ai.documents'::regclass
    ) THEN
        ALTER TABLE ai.documents ADD CONSTRAINT documents_owner_id_fkey
            FOREIGN KEY (owner_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.documents.storage_object_id -> storage.objects.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'documents_storage_object_id_fkey'
        AND conrelid = 'ai.documents'::regclass
    ) THEN
        ALTER TABLE ai.documents ADD CONSTRAINT documents_storage_object_id_fkey
            FOREIGN KEY (storage_object_id) REFERENCES storage.objects(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.document_permissions.granted_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'document_permissions_granted_by_fkey'
        AND conrelid = 'ai.document_permissions'::regclass
    ) THEN
        ALTER TABLE ai.document_permissions ADD CONSTRAINT document_permissions_granted_by_fkey
            FOREIGN KEY (granted_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.document_permissions.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'document_permissions_user_id_fkey'
        AND conrelid = 'ai.document_permissions'::regclass
    ) THEN
        ALTER TABLE ai.document_permissions ADD CONSTRAINT document_permissions_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ai.knowledge_base_permissions.granted_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'knowledge_base_permissions_granted_by_fkey'
        AND conrelid = 'ai.knowledge_base_permissions'::regclass
    ) THEN
        ALTER TABLE ai.knowledge_base_permissions ADD CONSTRAINT knowledge_base_permissions_granted_by_fkey
            FOREIGN KEY (granted_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.knowledge_base_permissions.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'knowledge_base_permissions_user_id_fkey'
        AND conrelid = 'ai.knowledge_base_permissions'::regclass
    ) THEN
        ALTER TABLE ai.knowledge_base_permissions ADD CONSTRAINT knowledge_base_permissions_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ai.providers.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'providers_created_by_fkey'
        AND conrelid = 'ai.providers'::regclass
    ) THEN
        ALTER TABLE ai.providers ADD CONSTRAINT providers_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.chatbots.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'chatbots_created_by_fkey'
        AND conrelid = 'ai.chatbots'::regclass
    ) THEN
        ALTER TABLE ai.chatbots ADD CONSTRAINT chatbots_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.conversations.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'conversations_user_id_fkey'
        AND conrelid = 'ai.conversations'::regclass
    ) THEN
        ALTER TABLE ai.conversations ADD CONSTRAINT conversations_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ai.query_audit_log.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'query_audit_log_user_id_fkey'
        AND conrelid = 'ai.query_audit_log'::regclass
    ) THEN
        ALTER TABLE ai.query_audit_log ADD CONSTRAINT query_audit_log_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.retrieval_log.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'retrieval_log_user_id_fkey'
        AND conrelid = 'ai.retrieval_log'::regclass
    ) THEN
        ALTER TABLE ai.retrieval_log ADD CONSTRAINT retrieval_log_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ai.user_chatbot_usage.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'user_chatbot_usage_user_id_fkey'
        AND conrelid = 'ai.user_chatbot_usage'::regclass
    ) THEN
        ALTER TABLE ai.user_chatbot_usage ADD CONSTRAINT user_chatbot_usage_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ai.user_provider_preferences.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'user_provider_preferences_user_id_fkey'
        AND conrelid = 'ai.user_provider_preferences'::regclass
    ) THEN
        ALTER TABLE ai.user_provider_preferences ADD CONSTRAINT user_provider_preferences_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ai.user_quotas.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'user_quotas_user_id_fkey'
        AND conrelid = 'ai.user_quotas'::regclass
    ) THEN
        ALTER TABLE ai.user_quotas ADD CONSTRAINT user_quotas_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- rpc -> auth
-- ============================================================================

-- rpc.procedures.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'procedures_created_by_fkey'
        AND conrelid = 'rpc.procedures'::regclass
    ) THEN
        ALTER TABLE rpc.procedures ADD CONSTRAINT procedures_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- rpc.executions.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'executions_user_id_fkey'
        AND conrelid = 'rpc.executions'::regclass
    ) THEN
        ALTER TABLE rpc.executions ADD CONSTRAINT executions_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- branching -> auth
-- ============================================================================

-- branching.branches.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'branches_created_by_fkey'
        AND conrelid = 'branching.branches'::regclass
    ) THEN
        ALTER TABLE branching.branches ADD CONSTRAINT branches_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id);
    END IF;
END $$;

-- branching.activity_log.executed_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'activity_log_executed_by_fkey'
        AND conrelid = 'branching.activity_log'::regclass
    ) THEN
        ALTER TABLE branching.activity_log ADD CONSTRAINT activity_log_executed_by_fkey
            FOREIGN KEY (executed_by) REFERENCES auth.users(id);
    END IF;
END $$;

-- branching.branch_access.granted_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'branch_access_granted_by_fkey'
        AND conrelid = 'branching.branch_access'::regclass
    ) THEN
        ALTER TABLE branching.branch_access ADD CONSTRAINT branch_access_granted_by_fkey
            FOREIGN KEY (granted_by) REFERENCES auth.users(id);
    END IF;
END $$;

-- branching.branch_access.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'branch_access_user_id_fkey'
        AND conrelid = 'branching.branch_access'::regclass
    ) THEN
        ALTER TABLE branching.branch_access ADD CONSTRAINT branch_access_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- jobs -> auth
-- ============================================================================

-- jobs.functions.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'functions_created_by_fkey'
        AND conrelid = 'jobs.functions'::regclass
    ) THEN
        ALTER TABLE jobs.functions ADD CONSTRAINT functions_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- jobs.queue.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'queue_created_by_fkey'
        AND conrelid = 'jobs.queue'::regclass
    ) THEN
        ALTER TABLE jobs.queue ADD CONSTRAINT queue_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- mcp -> auth
-- ============================================================================

-- mcp.custom_resources.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'custom_resources_created_by_fkey'
        AND conrelid = 'mcp.custom_resources'::regclass
    ) THEN
        ALTER TABLE mcp.custom_resources ADD CONSTRAINT custom_resources_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- mcp.custom_tools.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'custom_tools_created_by_fkey'
        AND conrelid = 'mcp.custom_tools'::regclass
    ) THEN
        ALTER TABLE mcp.custom_tools ADD CONSTRAINT custom_tools_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- ============================================================================
-- app -> auth
-- ============================================================================

-- app.settings.user_id -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'settings_user_id_fkey'
        AND conrelid = 'app.settings'::regclass
    ) THEN
        ALTER TABLE app.settings ADD CONSTRAINT settings_user_id_fkey
            FOREIGN KEY (user_id) REFERENCES auth.users(id) ON DELETE CASCADE;
    END IF;
END $$;

-- ============================================================================
-- migrations -> auth
-- ============================================================================

-- migrations.app.applied_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'app_applied_by_fkey'
        AND conrelid = 'migrations.app'::regclass
    ) THEN
        ALTER TABLE migrations.app ADD CONSTRAINT app_applied_by_fkey
            FOREIGN KEY (applied_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- migrations.app.created_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'app_created_by_fkey'
        AND conrelid = 'migrations.app'::regclass
    ) THEN
        ALTER TABLE migrations.app ADD CONSTRAINT app_created_by_fkey
            FOREIGN KEY (created_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;

-- migrations.execution_logs.executed_by -> auth.users.id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'execution_logs_executed_by_fkey'
        AND conrelid = 'migrations.execution_logs'::regclass
    ) THEN
        ALTER TABLE migrations.execution_logs ADD CONSTRAINT execution_logs_executed_by_fkey
            FOREIGN KEY (executed_by) REFERENCES auth.users(id) ON DELETE SET NULL;
    END IF;
END $$;
