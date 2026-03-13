--
-- MULTI-TENANCY: ROLLBACK TENANT_ID COLUMNS
-- Removes tenant_id columns from all tenant-scoped tables
--

-- AI Schema
ALTER TABLE ai.table_export_sync_configs DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.knowledge_base_permissions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.document_permissions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.chatbot_knowledge_bases DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.user_quotas DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.user_provider_preferences DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.user_chatbot_usage DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.conversations DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.chatbots DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.documents DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE ai.knowledge_bases DROP COLUMN IF EXISTS tenant_id;

-- Jobs Schema
ALTER TABLE jobs.queue DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE jobs.functions DROP COLUMN IF EXISTS tenant_id;

-- Functions Schema
ALTER TABLE functions.edge_triggers DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.secrets DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE functions.edge_functions DROP COLUMN IF EXISTS tenant_id;

-- Storage Schema
ALTER TABLE storage.object_permissions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.objects DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE storage.buckets DROP COLUMN IF EXISTS tenant_id;

-- Auth Schema
ALTER TABLE auth.client_keys DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.webhooks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE auth.users DROP COLUMN IF EXISTS tenant_id;
