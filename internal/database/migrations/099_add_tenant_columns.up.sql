--
-- MULTI-TENANCY: ADD TENANT_ID COLUMNS
-- Adds tenant_id column to all tenant-scoped tables
--

-- ============================================
-- AUTH SCHEMA
-- ============================================

-- Auth users: each user belongs to a tenant
ALTER TABLE auth.users ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_auth_users_tenant_id ON auth.users(tenant_id);
COMMENT ON COLUMN auth.users.tenant_id IS 'Tenant this user belongs to. NULL for backward compatibility during migration.';

-- Webhooks: webhooks are tenant-scoped
ALTER TABLE auth.webhooks ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_webhooks_tenant_id ON auth.webhooks(tenant_id);
COMMENT ON COLUMN auth.webhooks.tenant_id IS 'Tenant this webhook belongs to.';

-- Client keys (API keys): API keys are tenant-scoped
ALTER TABLE auth.client_keys ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_auth_client_keys_tenant_id ON auth.client_keys(tenant_id);
COMMENT ON COLUMN auth.client_keys.tenant_id IS 'Tenant this client key belongs to.';

-- ============================================
-- STORAGE SCHEMA
-- ============================================

-- Buckets: storage buckets are tenant-scoped
ALTER TABLE storage.buckets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_storage_buckets_tenant_id ON storage.buckets(tenant_id);
COMMENT ON COLUMN storage.buckets.tenant_id IS 'Tenant this bucket belongs to.';

-- Objects: storage objects are tenant-scoped
ALTER TABLE storage.objects ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_storage_objects_tenant_id ON storage.objects(tenant_id);
COMMENT ON COLUMN storage.objects.tenant_id IS 'Tenant this object belongs to.';

-- Object permissions: permissions are tenant-scoped
ALTER TABLE storage.object_permissions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_storage_object_permissions_tenant_id ON storage.object_permissions(tenant_id);
COMMENT ON COLUMN storage.object_permissions.tenant_id IS 'Tenant this permission belongs to.';

-- ============================================
-- FUNCTIONS SCHEMA
-- ============================================

-- Edge functions: functions are tenant-scoped
ALTER TABLE functions.edge_functions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_edge_functions_tenant_id ON functions.edge_functions(tenant_id);
COMMENT ON COLUMN functions.edge_functions.tenant_id IS 'Tenant this function belongs to.';

-- Secrets: secrets are tenant-scoped
ALTER TABLE functions.secrets ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_secrets_tenant_id ON functions.secrets(tenant_id);
COMMENT ON COLUMN functions.secrets.tenant_id IS 'Tenant this secret belongs to.';

-- Edge triggers: triggers are tenant-scoped
ALTER TABLE functions.edge_triggers ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_functions_edge_triggers_tenant_id ON functions.edge_triggers(tenant_id);
COMMENT ON COLUMN functions.edge_triggers.tenant_id IS 'Tenant this trigger belongs to.';

-- ============================================
-- JOBS SCHEMA
-- ============================================

-- Job functions: background jobs are tenant-scoped
ALTER TABLE jobs.functions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_jobs_functions_tenant_id ON jobs.functions(tenant_id);
COMMENT ON COLUMN jobs.functions.tenant_id IS 'Tenant this job belongs to.';

-- Job queue: queue items are tenant-scoped
ALTER TABLE jobs.queue ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_jobs_queue_tenant_id ON jobs.queue(tenant_id);
COMMENT ON COLUMN jobs.queue.tenant_id IS 'Tenant this queue item belongs to.';

-- ============================================
-- AI SCHEMA
-- ============================================

-- Knowledge bases: knowledge bases are tenant-scoped
ALTER TABLE ai.knowledge_bases ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_bases_tenant_id ON ai.knowledge_bases(tenant_id);
COMMENT ON COLUMN ai.knowledge_bases.tenant_id IS 'Tenant this knowledge base belongs to.';

-- Documents: documents are tenant-scoped
ALTER TABLE ai.documents ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_documents_tenant_id ON ai.documents(tenant_id);
COMMENT ON COLUMN ai.documents.tenant_id IS 'Tenant this document belongs to.';

-- Chatbots: chatbots are tenant-scoped
ALTER TABLE ai.chatbots ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_chatbots_tenant_id ON ai.chatbots(tenant_id);
COMMENT ON COLUMN ai.chatbots.tenant_id IS 'Tenant this chatbot belongs to.';

-- Conversations: conversations are tenant-scoped
ALTER TABLE ai.conversations ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_conversations_tenant_id ON ai.conversations(tenant_id);
COMMENT ON COLUMN ai.conversations.tenant_id IS 'Tenant this conversation belongs to.';

-- User chatbot usage: usage tracking is tenant-scoped
ALTER TABLE ai.user_chatbot_usage ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_user_chatbot_usage_tenant_id ON ai.user_chatbot_usage(tenant_id);
COMMENT ON COLUMN ai.user_chatbot_usage.tenant_id IS 'Tenant for this usage record.';

-- User provider preferences: preferences are tenant-scoped
ALTER TABLE ai.user_provider_preferences ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_user_provider_preferences_tenant_id ON ai.user_provider_preferences(tenant_id);
COMMENT ON COLUMN ai.user_provider_preferences.tenant_id IS 'Tenant for this preference.';

-- User quotas: quotas are tenant-scoped
ALTER TABLE ai.user_quotas ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_user_quotas_tenant_id ON ai.user_quotas(tenant_id);
COMMENT ON COLUMN ai.user_quotas.tenant_id IS 'Tenant for this quota.';

-- Chatbot knowledge bases link table
ALTER TABLE ai.chatbot_knowledge_bases ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_chatbot_knowledge_bases_tenant_id ON ai.chatbot_knowledge_bases(tenant_id);
COMMENT ON COLUMN ai.chatbot_knowledge_bases.tenant_id IS 'Tenant for this link.';

-- Document permissions
ALTER TABLE ai.document_permissions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_document_permissions_tenant_id ON ai.document_permissions(tenant_id);
COMMENT ON COLUMN ai.document_permissions.tenant_id IS 'Tenant for this permission.';

-- Knowledge base permissions
ALTER TABLE ai.knowledge_base_permissions ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_knowledge_base_permissions_tenant_id ON ai.knowledge_base_permissions(tenant_id);
COMMENT ON COLUMN ai.knowledge_base_permissions.tenant_id IS 'Tenant for this permission.';

-- Table export sync configs
ALTER TABLE ai.table_export_sync_configs ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_ai_table_export_sync_configs_tenant_id ON ai.table_export_sync_configs(tenant_id);
COMMENT ON COLUMN ai.table_export_sync_configs.tenant_id IS 'Tenant for this sync config.';
