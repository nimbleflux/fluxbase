--
-- MULTI-TENANCY: REMOVE RLS POLICIES
-- Rollback migration 116
--

-- AI Schema
DROP POLICY IF EXISTS chunks_tenant_service ON ai.chunks;
DROP POLICY IF EXISTS chunks_instance_admin ON ai.chunks;
DROP POLICY IF EXISTS chunks_tenant_admin ON ai.chunks;
DROP POLICY IF EXISTS chunks_tenant_member ON ai.chunks;

DROP POLICY IF EXISTS messages_tenant_service ON ai.messages;
DROP POLICY IF EXISTS messages_instance_admin ON ai.messages;
DROP POLICY IF EXISTS messages_tenant_admin ON ai.messages;
DROP POLICY IF EXISTS messages_tenant_member ON ai.messages;

DROP POLICY IF EXISTS entities_tenant_service ON ai.entities;
DROP POLICY IF EXISTS entities_instance_admin ON ai.entities;
DROP POLICY IF EXISTS entities_tenant_admin ON ai.entities;
DROP POLICY IF EXISTS entities_tenant_member ON ai.entities;

DROP POLICY IF EXISTS entity_relationships_tenant_service ON ai.entity_relationships;
DROP POLICY IF EXISTS entity_relationships_instance_admin ON ai.entity_relationships;
DROP POLICY IF EXISTS entity_relationships_tenant_admin ON ai.entity_relationships;
DROP POLICY IF EXISTS entity_relationships_tenant_member ON ai.entity_relationships;

DROP POLICY IF EXISTS document_entities_tenant_service ON ai.document_entities;
DROP POLICY IF EXISTS document_entities_instance_admin ON ai.document_entities;
DROP POLICY IF EXISTS document_entities_tenant_admin ON ai.document_entities;
DROP POLICY IF EXISTS document_entities_tenant_member ON ai.document_entities;

DROP POLICY IF EXISTS query_audit_log_tenant_service ON ai.query_audit_log;
DROP POLICY IF EXISTS query_audit_log_instance_admin ON ai.query_audit_log;
DROP POLICY IF EXISTS query_audit_log_tenant_admin ON ai.query_audit_log;

DROP POLICY IF EXISTS retrieval_log_tenant_service ON ai.retrieval_log;
DROP POLICY IF EXISTS retrieval_log_instance_admin ON ai.retrieval_log;
DROP POLICY IF EXISTS retrieval_log_tenant_admin ON ai.retrieval_log;

-- Functions Schema
DROP POLICY IF EXISTS edge_executions_tenant_service ON functions.edge_executions;
DROP POLICY IF EXISTS edge_executions_instance_admin ON functions.edge_executions;
DROP POLICY IF EXISTS edge_executions_tenant_admin ON functions.edge_executions;
DROP POLICY IF EXISTS edge_executions_tenant_member ON functions.edge_executions;

DROP POLICY IF EXISTS edge_files_tenant_service ON functions.edge_files;
DROP POLICY IF EXISTS edge_files_instance_admin ON functions.edge_files;
DROP POLICY IF EXISTS edge_files_tenant_admin ON functions.edge_files;
DROP POLICY IF EXISTS edge_files_tenant_member ON functions.edge_files;

DROP POLICY IF EXISTS function_dependencies_tenant_service ON functions.function_dependencies;
DROP POLICY IF EXISTS function_dependencies_instance_admin ON functions.function_dependencies;
DROP POLICY IF EXISTS function_dependencies_tenant_admin ON functions.function_dependencies;
DROP POLICY IF EXISTS function_dependencies_tenant_member ON functions.function_dependencies;

DROP POLICY IF EXISTS secret_versions_tenant_service ON functions.secret_versions;
DROP POLICY IF EXISTS secret_versions_instance_admin ON functions.secret_versions;
DROP POLICY IF EXISTS secret_versions_tenant_admin ON functions.secret_versions;

DROP POLICY IF EXISTS shared_modules_tenant_service ON functions.shared_modules;
DROP POLICY IF EXISTS shared_modules_instance_admin ON functions.shared_modules;
DROP POLICY IF EXISTS shared_modules_tenant_admin ON functions.shared_modules;
DROP POLICY IF EXISTS shared_modules_tenant_member ON functions.shared_modules;

-- Jobs Schema
DROP POLICY IF EXISTS jobs_function_files_tenant_service ON jobs.function_files;
DROP POLICY IF EXISTS jobs_function_files_instance_admin ON jobs.function_files;
DROP POLICY IF EXISTS jobs_function_files_tenant_admin ON jobs.function_files;
DROP POLICY IF EXISTS jobs_function_files_tenant_member ON jobs.function_files;

DROP POLICY IF EXISTS workers_tenant_service ON jobs.workers;
DROP POLICY IF EXISTS workers_instance_admin ON jobs.workers;
DROP POLICY IF EXISTS workers_tenant_admin ON jobs.workers;
DROP POLICY IF EXISTS workers_tenant_member ON jobs.workers;

-- Auth Schema
DROP POLICY IF EXISTS sessions_tenant_service ON auth.sessions;
DROP POLICY IF EXISTS sessions_tenant_admin ON auth.sessions;

DROP POLICY IF EXISTS mfa_factors_tenant_service ON auth.mfa_factors;
DROP POLICY IF EXISTS mfa_factors_instance_admin ON auth.mfa_factors;
DROP POLICY IF EXISTS mfa_factors_tenant_admin ON auth.mfa_factors;

DROP POLICY IF EXISTS oauth_links_tenant_service ON auth.oauth_links;
DROP POLICY IF EXISTS oauth_links_instance_admin ON auth.oauth_links;
DROP POLICY IF EXISTS oauth_links_tenant_admin ON auth.oauth_links;

DROP POLICY IF EXISTS impersonation_sessions_tenant_service ON auth.impersonation_sessions;
DROP POLICY IF EXISTS impersonation_sessions_instance_admin ON auth.impersonation_sessions;
DROP POLICY IF EXISTS impersonation_sessions_tenant_admin ON auth.impersonation_sessions;

DROP POLICY IF EXISTS webhook_deliveries_tenant_service ON auth.webhook_deliveries;
DROP POLICY IF EXISTS webhook_deliveries_instance_admin ON auth.webhook_deliveries;
DROP POLICY IF EXISTS webhook_deliveries_tenant_admin ON auth.webhook_deliveries;
DROP POLICY IF EXISTS webhook_deliveries_tenant_member ON auth.webhook_deliveries;

DROP POLICY IF EXISTS webhook_events_tenant_service ON auth.webhook_events;
DROP POLICY IF EXISTS webhook_events_instance_admin ON auth.webhook_events;
DROP POLICY IF EXISTS webhook_events_tenant_admin ON auth.webhook_events;
DROP POLICY IF EXISTS webhook_events_tenant_member ON auth.webhook_events;

DROP POLICY IF EXISTS webhook_monitored_tables_tenant_service ON auth.webhook_monitored_tables;
DROP POLICY IF EXISTS webhook_monitored_tables_instance_admin ON auth.webhook_monitored_tables;
DROP POLICY IF EXISTS webhook_monitored_tables_tenant_admin ON auth.webhook_monitored_tables;
DROP POLICY IF EXISTS webhook_monitored_tables_tenant_member ON auth.webhook_monitored_tables;

DROP POLICY IF EXISTS client_key_usage_tenant_service ON auth.client_key_usage;
DROP POLICY IF EXISTS client_key_usage_instance_admin ON auth.client_key_usage;
DROP POLICY IF EXISTS client_key_usage_tenant_admin ON auth.client_key_usage;
DROP POLICY IF EXISTS client_key_usage_tenant_member ON auth.client_key_usage;

-- Storage Schema
DROP POLICY IF EXISTS chunked_upload_sessions_tenant_service ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS chunked_upload_sessions_instance_admin ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS chunked_upload_sessions_tenant_admin ON storage.chunked_upload_sessions;
DROP POLICY IF EXISTS chunked_upload_sessions_tenant_member ON storage.chunked_upload_sessions;

-- Logging Schema (parent table only, partitions inherit)
DROP POLICY IF EXISTS entries_tenant_service ON logging.entries;
DROP POLICY IF EXISTS entries_instance_admin ON logging.entries;
DROP POLICY IF EXISTS entries_tenant_admin ON logging.entries;

-- Branching Schema
DROP POLICY IF EXISTS branches_tenant_service ON branching.branches;
DROP POLICY IF EXISTS branches_instance_admin ON branching.branches;
DROP POLICY IF EXISTS branches_tenant_admin ON branching.branches;
DROP POLICY IF EXISTS branches_tenant_member ON branching.branches;

DROP POLICY IF EXISTS branch_access_tenant_service ON branching.branch_access;
DROP POLICY IF EXISTS branch_access_instance_admin ON branching.branch_access;
DROP POLICY IF EXISTS branch_access_tenant_admin ON branching.branch_access;
DROP POLICY IF EXISTS branch_access_tenant_member ON branching.branch_access;

DROP POLICY IF EXISTS github_config_tenant_service ON branching.github_config;
DROP POLICY IF EXISTS github_config_instance_admin ON branching.github_config;
DROP POLICY IF EXISTS github_config_tenant_admin ON branching.github_config;

DROP POLICY IF EXISTS branching_activity_log_tenant_service ON branching.activity_log;
DROP POLICY IF EXISTS branching_activity_log_instance_admin ON branching.activity_log;
DROP POLICY IF EXISTS branching_activity_log_tenant_admin ON branching.activity_log;
DROP POLICY IF EXISTS branching_activity_log_tenant_member ON branching.activity_log;

DROP POLICY IF EXISTS migration_history_tenant_service ON branching.migration_history;
DROP POLICY IF EXISTS migration_history_instance_admin ON branching.migration_history;
DROP POLICY IF EXISTS migration_history_tenant_admin ON branching.migration_history;
DROP POLICY IF EXISTS migration_history_tenant_member ON branching.migration_history;

DROP POLICY IF EXISTS seed_execution_log_tenant_service ON branching.seed_execution_log;
DROP POLICY IF EXISTS seed_execution_log_instance_admin ON branching.seed_execution_log;
DROP POLICY IF EXISTS seed_execution_log_tenant_admin ON branching.seed_execution_log;

-- RPC Schema
DROP POLICY IF EXISTS rpc_executions_tenant_service ON rpc.executions;
DROP POLICY IF EXISTS rpc_executions_instance_admin ON rpc.executions;
DROP POLICY IF EXISTS rpc_executions_tenant_admin ON rpc.executions;
DROP POLICY IF EXISTS rpc_executions_tenant_member ON rpc.executions;

DROP POLICY IF EXISTS procedures_tenant_service ON rpc.procedures;
DROP POLICY IF EXISTS procedures_instance_admin ON rpc.procedures;
DROP POLICY IF EXISTS procedures_tenant_admin ON rpc.procedures;
DROP POLICY IF EXISTS procedures_tenant_member ON rpc.procedures;

-- Realtime Schema
DROP POLICY IF EXISTS schema_registry_tenant_service ON realtime.schema_registry;
DROP POLICY IF EXISTS schema_registry_instance_admin ON realtime.schema_registry;
DROP POLICY IF EXISTS schema_registry_tenant_admin ON realtime.schema_registry;
DROP POLICY IF EXISTS schema_registry_tenant_member ON realtime.schema_registry;
