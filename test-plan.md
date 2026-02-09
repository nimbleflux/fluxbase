# Comprehensive Test Coverage Improvement Plan: 30% → 90%

## Executive Summary

**Current State:**
- Overall coverage: 26.8%
- 323 files with 0% coverage (282 internal modules)
- Files with most uncovered lines:
  - `internal/functions/handler.go` - 476 uncovered lines
  - `internal/api/query_parser.go` - 451 uncovered lines
  - `internal/api/server.go` - 445 uncovered lines
  - `internal/api/auth_handler.go` - 426 uncovered lines
  - `internal/ai/handler.go` - 413 uncovered lines
  - `internal/auth/service.go` - 310 uncovered lines

**Target:** 90% overall coverage

**Strategy:** Phased approach over 12 weeks, prioritizing business-critical modules first, reusing existing test patterns and mock infrastructure.

---

## Phase 1: Foundation & Critical Auth (30% → 45%)

**Target Coverage:** 45% overall
**Estimated Duration:** 2-3 weeks
**Focus:** Complete auth module coverage and expand test infrastructure

### 1.1 Expand Auth Service Tests (HIGHEST PRIORITY)

**File:** [internal/auth/service.go](internal/auth/service.go)
**Current:** 1,067 lines of tests (good foundation)
**Gap:** ~310 uncovered lines

**Tests to Add:**

#### Password Change/Reset Workflows (15 tests)
- `TestChangePassword_Success`, `TestChangePassword_WrongCurrentPassword`, `TestChangePassword_WeakPassword`
- `TestResetPassword_ValidToken`, `TestResetPassword_ExpiredToken`, `TestResetPassword_InvalidToken`
- `TestForgotPassword_RateLimited`, `TestForgotPassword_EmailFailed`
- `TestRequestPasswordReset_Success`, `TestRequestPasswordReset_NonExistentEmail`

#### MFA/TOTP Flows (20 tests)
- `TestEnableMFA_Success`, `TestEnableMFA_AlreadyEnabled`, `TestEnableMFA_InvalidSecret`
- `TestVerifyMFA_ValidCode`, `TestVerifyMFA_InvalidCode`, `TestVerifyMFA_RateLimited`
- `TestDisableMFA_Success`, `TestDisableMFA_NotEnabled`, `TestDisableMFA_RequiresVerification`
- `TestGenerateBackupCodes`, `TestVerifyBackupCode`, `TestVerifyBackupCode_AlreadyUsed`, `TestRegenerateBackupCodes`

#### OAuth Integration (10 tests)
- `TestGetOAuthURL_Success`, `TestGetOAuthURL_InvalidProvider`
- `TestHandleOAuthCallback_Success`, `TestHandleOAuthCallback_EmailMismatch`, `TestHandleOAuthCallback_StateMismatch`
- `TestLinkOAuthIdentity_Success`, `TestLinkOAuthIdentity_AlreadyLinked`, `TestUnlinkOAuthIdentity`

#### Session Management (12 tests)
- `TestListSessions_All`, `TestListSessions_ActiveOnly`
- `TestRevokeSession_Success`, `TestRevokeSession_NotFound`, `TestRevokeAllSessions`
- `TestRefreshSession_Success`, `TestRefreshSession_Expired`, `TestRefreshSession_InvalidToken`

#### User Metadata Operations (8 tests)
- `TestUpdateUserMetadata_Merge`, `TestUpdateUserMetadata_Replace`
- `TestUpdateAppMetadata_AdminOnly`, `TestMergeMetadata_ConflictResolution`

#### Impersonation (10 tests)
- `TestImpersonateUser_Success`, `TestImpersonateUser_NotAdmin`, `TestImpersonateUser_TargetNotFound`
- `TestImpersonateUser_InvalidTarget`, `TestStopImpersonation_Success`, `TestStopImpersonation_NotImpersonating`

**Mock Infrastructure Needed:**
- Extend `MockOAuthProvider` in testutil (add state tracking)
- Add `MockTOTPValidator` (generate/validate codes)
- Add `MockRateLimiter` for auth-specific rate limiting

**Estimated Tests:** 75 new test functions
**Estimated Lines:** 1,200 lines

### 1.2 Complete Auth Module Coverage

#### [internal/auth/saml.go](internal/auth/saml.go) (1,458 lines, 343 uncovered)
- `TestInitiateSAML_Login`, `TestInitiateSAML_InvalidProvider`
- `TestHandleSAML_Callback_Success`, `TestHandleSAML_Callback_InvalidResponse`, `TestHandleSAML_Callback_EmailMismatch`
- `TestValidateSAML_Response_Signature`, `TestValidateSAML_Response_Tampered`
- `TestSAMLProvider_Configuration`, `TestSAMLProvider_Metadata`
- Estimated: 40 tests, 600 lines

#### [internal/auth/dashboard.go](internal/auth/dashboard.go) (1,141 lines, 293 uncovered)
- `TestDashboardAuthentication_Success`, `TestDashboardAuthentication_InvalidCredentials`
- `TestDashboardAuthorization_AdminOnly`, `TestDashboardAuthorization_Forbidden`
- `TestDashboardImpersonation_Start`, `TestDashboardImpersonation_Stop`, `TestDashboardImpersonation_NotAdmin`
- `TestDashboardUserManagement_List`, `TestDashboardUserManagement_Update`, `TestDashboardUserManagement_Delete`
- Estimated: 35 tests, 500 lines

#### [internal/auth/identity.go](internal/auth/identity.go)
- `TestListIdentities_All`, `TestListIdentities_ByProvider`
- `TestLinkIdentity_Success`, `TestLinkIdentity_AlreadyLinked`, `TestLinkIdentity_InvalidProvider`
- `TestUnlinkIdentity_Success`, `TestUnlinkIdentity_LastIdentity`, `TestUnlinkIdentity_NotFound`
- `TestIdentityProviders_OAuth`, `TestIdentityProviders_SAML`, `TestIdentityProviders_All`
- Estimated: 25 tests, 400 lines

#### [internal/auth/password.go](internal/auth/password.go)
- `TestPasswordValidation_Valid`, `TestPasswordValidation_TooShort`, `TestPasswordValidation_TooWeak`
- `TestPasswordStrength_Check`, `TestPasswordStrength_CommonPasswords`
- `TestPasswordHashing_Bcrypt`, `TestPasswordHashing_Argon2`, `TestPasswordHashing_Verify`
- Estimated: 20 tests, 300 lines

**Phase 1 Subtotal:** 195 new test functions, ~3,000 lines

---

## Phase 2: Core API Handlers (45% → 65%)

**Target Coverage:** 65% overall
**Estimated Duration:** 3-4 weeks
**Focus:** REST API, GraphQL, and core handlers

### 2.1 REST API CRUD Operations

**File:** [internal/api/rest_crud.go](internal/api/rest_crud.go) (109 uncovered lines)

#### GET Handlers (15 tests)
- `TestGetHandler_ValidQuery`, `TestGetHandler_InvalidQuery`, `TestGetHandler_ParseError`
- `TestGetHandler_RLSFilter_UserData`, `TestGetHandler_RLSFilter_AdminBypass`
- `TestGetHandler_Pagination_Offset`, `TestGetHandler_Pagination_Cursor`, `TestGetHandler_Pagination_Both`
- `TestGetHandler_Aggregations_Sum`, `TestGetHandler_Aggregations_Count`, `TestGetHandler_Aggregations_Avg`
- `TestGetHandler_Count_Only`, `TestGetHandler_SingleRecord`, `TestGetHandler_NotFound`

#### POST/PUT/PATCH Handlers (20 tests)
- `TestPostHandler_Create_Success`, `TestPostHandler_Create_Validation`, `TestPostHandler_Create_Conflict`
- `TestPostHandler_Create_RLSViolation`, `TestPostHandler_Create_Relations`
- `TestPutHandler_Update_Success`, `TestPutHandler_Update_NotFound`, `TestPutHandler_Update_Conflict`
- `TestPutHandler_Update_ImmutableFields`, `TestPutHandler_Update_RLSViolation`
- `TestPatchHandler_PartialUpdate`, `TestPatchHandler_MissingFields`, `TestPatchHandler_NullFields`

#### DELETE Handler (8 tests)
- `TestDeleteHandler_Success`, `TestDeleteHandler_NotFound`, `TestDeleteHandler_AlreadyDeleted`
- `TestDeleteHandler_CascadeDelete`, `TestDeleteHandler_RLSViolation`
- `TestDeleteHandler_SoftDelete`, `TestDeleteHandler_HardDelete`

#### Embedded Relations (10 tests)
- `TestEmbed_OneToMany`, `TestEmbed_ManyToOne`, `TestEmbed_ManyToMany`
- `TestEmbed_DeepNesting`, `TestEmbed_CircularReferences`, `TestEmbed_NullRelations`
- `TestEmbed_FilteredRelations`, `TestEmbed_LimitRelations`

#### Admin Bypass (5 tests)
- `TestAdminUser_BypassesMaxResults`, `TestAdminUser_UnrestrictedAccess`
- `TestAdminUser_SeeAllRecords`, `TestAdminUser_BypassRLS`, `TestNormalUser_RLSEnforced`

**Mock Infrastructure Needed:**
- `MockDBConnection` for query execution
- `MockTableInspector` for schema metadata
- `MockRLSWrapper` for RLS context testing

**Estimated Tests:** 58 test functions, 900 lines

### 2.2 Query Parser & Builder

**File:** [internal/api/query_parser.go](internal/api/query_parser.go) (1,694 lines, 451 uncovered despite 1,586 test lines)

#### Advanced Filtering (25 tests)
- `TestOrGroup_Combinations`, `TestOrGroup_Nested`, `TestOrGroup_MultipleConditions`
- `TestNotOperator_LogicalInversion`, `TestNotOperator_WithFilters`
- `TestFilter_JSONColumns_Contains`, `TestFilter_JSONColumns_Path`, `TestFilter_ArrayColumns`
- `TestFilter_FTS_Search_Simple`, `TestFilter_FTS_Search_Phrase`, `TestFilter_FTS_Search_Boolean`
- `TestFilter_NullValues`, `TestFilter_EmptyArrays`, `TestFilter_DateRanges`

#### Cursor Pagination (15 tests)
- `TestCursorEncoding_Decoding`, `TestCursorDecoding_Encoding`, `TestCursorDecoding_Invalid`
- `TestCursorPagination_Forward`, `TestCursorPagination_Backward`, `TestCursorPagination_Bidirectional`
- `TestCursorPagination_CustomColumn`, `TestCursorPagination_MultipleColumns`
- `TestCursorPagination_EdgeCases`

#### Aggregations (20 tests)
- `TestAggregation_Sum`, `TestAggregation_Avg`, `TestAggregation_Count`, `TestAggregation_Max`, `TestAggregation_Min`
- `TestAggregation_GroupBy_Single`, `TestAggregation_GroupBy_Multiple`
- `TestAggregation_Having`, `TestAggregation_Multiple`, `TestAggregation_NullHandling`
- `TestAggregation_Distinct`, `TestAggregation_Combined`

#### Complex Joins (10 tests)
- `TestJoin_Inner`, `TestJoin_LeftOuter`, `TestJoin_RightOuter`, `TestJoin_FullOuter`
- `TestJoin_MultipleTables`, `TestJoin_Aliasing`, `TestJoin_Subquery`
- `TestJoin_CircularReferences`

#### SQL Injection Protection (15 tests)
- `TestIdentifierValidation_MaliciousInput`, `TestIdentifierValidation_SQLInjection`
- `TestQuoteIdentifier_Escaping`, `TestQuoteIdentifier_InvalidChars`, `TestQuoteIdentifier_ReservedWords`
- `TestValueValidation_Unquoted`, `TestValueValidation_Quoted`, `TestValueValidation_Arrays`

**Estimated Tests:** 85 test functions, 1,300 lines

### 2.3 API Server Setup

**File:** [internal/api/server.go](internal/api/server.go) (2,820 lines, 8 tests, 445 uncovered)

#### Server Initialization (15 tests)
- `TestNewServer_WithConfig`, `TestNewServer_DefaultConfig`, `TestNewServer_MissingConfig`
- `TestServer_MiddlewareSetup_All`, `TestServer_MiddlewareOrder`
- `TestServer_RouteRegistration`, `TestServer_RouteGroups`, `TestServer_CORSConfiguration`
- `TestServer_RateLimiting_Global`, `TestServer_RateLimiting_PerIP`
- `TestServer_Tracing_Enabled`, `TestServer_Tracing_Disabled`

#### Handler Registration (20 tests)
- `TestRegisterRESTHandlers_GET`, `TestRegisterRESTHandlers_POST`, `TestRegisterRESTHandlers_PUT`, `TestRegisterRESTHandlers_DELETE`
- `TestRegisterGraphQLHandler_Query`, `TestRegisterGraphQLHandler_Mutation`
- `TestRegisterStorageHandlers_Upload`, `TestRegisterStorageHandlers_Download`, `TestRegisterStorageHandlers_List`
- `TestRegisterAuthHandlers_Login`, `TestRegisterAuthHandlers_Refresh`, `TestRegisterAuthHandlers_Logout`
- `TestRegisterBranchingHandlers_Create`, `TestRegisterBranchingHandlers_Delete`, `TestRegisterBranchingHandlers_Switch`
- `TestRegisterMCPHandlers_Tools`, `TestRegisterMCPHandlers_Resources`

#### Request Lifecycle (10 tests)
- `TestRequest_MiddlewareChain_Order`, `TestRequest_ErrorHandling_Middleware`
- `TestRequest_Timeout`, `TestRequest_ContextCancellation`
- `TestRequest_PanicRecovery`, `TestRequest_Logging`

#### Graceful Shutdown (8 tests)
- `TestShutdown_InFlightRequests_Wait`, `TestShutdown_InFlightRequests_Timeout`
- `TestShutdown_ConnectionDraining`, `TestShutdown_SignalHandling`
- `TestShutdown_ForceExit`

**Estimated Tests:** 53 test functions, 800 lines

### 2.4 GraphQL Handler & Resolvers

#### [internal/api/graphql_handler.go](internal/api/graphql_handler.go)
- `TestGraphQLHandler_Query_Success`, `TestGraphQLHandler_Query_SyntaxError`
- `TestGraphQLHandler_Mutation_Success`, `TestGraphQLHandler_Mutation_ValidationError`
- `TestGraphQLHandler_Introspection`, `TestGraphQLHandler_Subscription`
- `TestGraphQLHandler_Errors`, `TestGraphQLHandler_Batching`
- Estimated: 25 tests, 400 lines

#### [internal/api/graphql_resolvers.go](internal/api/graphql_resolvers.go)
- `TestQueryResolver_SingleRecord`, `TestQueryResolver_List`, `TestQueryResolver_Aggregate`
- `TestMutationResolver_Insert`, `TestMutationResolver_Update`, `TestMutationResolver_Delete`
- `TestSubscriptionResolver_Realtime`, `TestSubscriptionResolver_Filter`
- `TestResolver_RLSContext`, `TestResolver_ErrorHandling`
- Estimated: 40 tests, 700 lines

**Phase 2 Subtotal:** 261 new test functions, ~4,100 lines

---

## Phase 3: AI & Advanced Features (65% → 80%)

**Target Coverage:** 80% overall
**Estimated Duration:** 3-4 weeks
**Focus:** AI module, functions, jobs, and MCP tools

### 3.1 AI Module

#### [internal/ai/handler.go](internal/ai/handler.go) (2,004 lines, 777 test lines, 413 uncovered)
- `TestChatHandler_Completion`, `TestChatHandler_Stream`, `TestChatHandler_ConversationHistory`
- `TestChatHandler_Tools_Call`, `TestChatHandler_Tools_Validation`, `TestChatHandler_MCPBridge`
- `TestChatHandler_RAGContext`, `TestChatHandler_VectorSearch`
- `TestChatHandler_RateLimiting`, `TestChatHandler_ErrorHandling`
- Estimated: 30 tests, 600 lines

#### [internal/ai/embedding_service.go](internal/ai/embedding_service.go)
- `TestGenerateEmbedding_OpenAI`, `TestGenerateEmbedding_Azure`, `TestGenerateEmbedding_Ollama`
- `TestBatchEmbeddings_Small`, `TestBatchEmbeddings_Large`, `TestBatchEmbeddings_ErrorHandling`
- `TestEmbeddingCache_Hit`, `TestEmbeddingCache_Miss`
- Estimated: 20 tests, 350 lines

#### [internal/ai/rag_service.go](internal/ai/rag_service.go)
- `TestRAG_RetrieveDocuments`, `TestRAG_RetrieveDocuments_Empty`
- `TestRAG_GenerateResponse`, `TestRAG_HybridSearch`
- `TestRAG_ReRanking`, `TestRAG_ContextWindow`, `TestRAG_Citations`
- Estimated: 25 tests, 450 lines

#### [internal/ai/knowledge_base_storage.go](internal/ai/knowledge_base_storage.go) (1,286 lines, 268 uncovered)
- `TestKnowledgeBase_Create`, `TestKnowledgeBase_Read`, `TestKnowledgeBase_Update`, `TestKnowledgeBase_Delete`
- `TestKnowledgeBase_Search_Vector`, `TestKnowledgeBase_Search_Metadata`, `TestKnowledgeBase_Search_Hybrid`
- `TestKnowledgeBase_Document_Add`, `TestKnowledgeBase_Document_Remove`, `TestKnowledgeBase_Document_Update`
- Estimated: 30 tests, 550 lines

#### [internal/ai/ocr_service.go](internal/ai/ocr_service.go)
- `TestOCR_ProcessPDF_Success`, `TestOCR_ProcessPDF_Error`, `TestOCR_ProcessImage`
- `TestOCR_Tesseract_Available`, `TestOCR_Tesseract_Unavailable`
- `TestOCR_MultiPage`, `TestOCR_ErrorHandling`
- Estimated: 15 tests, 250 lines

**Mock Infrastructure Needed:**
- `MockOpenAIClient`, `MockAzureClient`, `MockOllamaClient`
- `MockVectorDatabase` (extend existing)
- `MockOCRProvider`

### 3.2 Functions Module

#### [internal/functions/handler.go](internal/functions/handler.go) (2,139 lines, 476 uncovered)

##### Function Invocation (25 tests)
- `TestInvokeFunction_Success`, `TestInvokeFunction_NotFound`, `TestInvokeFunction_SyntaxError`
- `TestInvokeFunction_Timeout`, `TestInvokeFunction_Retry`, `TestInvokeFunction_MemoryLimit`
- `TestInvokeFunction_Concurrency_Limit`, `TestInvokeFunction_Concurrency_Unlimited`
- `TestInvokeFunction_WithSecrets`, `TestInvokeFunction_WithServiceKey`
- `TestInvokeFunction_Logs_Stream`, `TestInvokeFunction_Logs_Disabled`
- `TestInvokeFunction_Cache_Hit`, `TestInvokeFunction_Cache_Miss`

##### Bundling (20 tests)
- `TestBundleFunction_Deno`, `TestBundleFunction_Imports`, `TestBundleFunction_Minification`
- `TestBundleFunction_Errors_Syntax`, `TestBundleFunction_Errors_Runtime`
- `TestBundleFunction_Caching`, `TestBundleFunction_IncrementalBuild`
- `TestBundleFunction_SourceMap`, `TestBundleFunction_TypeChecking`

##### Scheduling (15 tests)
- `TestScheduleFunction_Cron`, `TestScheduleFunction_Interval`, `TestScheduleFunction_OneTime`
- `TestScheduleFunction_Unschedule`, `TestScheduleFunction_Update`
- `TestScheduleFunction_Concurrency`, `TestScheduleFunction_RetryPolicy`

##### Storage/Loading (10 tests)
- `TestLoadFunction_Valid`, `TestLoadFunction_Invalid`, `TestLoadFunction_CircularDeps`
- `TestLoadFunction_Cache`, `TestLoadFunction_HotReload`

**Estimated Tests:** 70 test functions, 1,100 lines

### 3.3 Jobs Module

#### [internal/jobs/worker.go](internal/jobs/worker.go) (153 uncovered)

##### Job Execution (20 tests)
- `TestWorker_ExecuteJob_Success`, `TestWorker_ExecuteJob_Failure`, `TestWorker_ExecuteJob_Panic`
- `TestWorker_RetryPolicy_Linear`, `TestWorker_RetryPolicy_Exponential`
- `TestWorker_Timeout`, `TestWorker_DeadlineExceeded`
- `TestWorker_ConcurrencyLimit`, `TestWorker_DrainMode`
- `TestWorker_PriorityQueue`, `TestWorker_JobCancellation`

##### Job Lifecycle (15 tests)
- `TestJob_Create`, `TestJob_Enqueue`, `TestJob_Dequeue`
- `TestJob_Complete`, `TestJob_Fail`, `TestJob_Cancel`
- `TestJob_Progress_Updates`, `TestJob_Result_Storage`
- `TestJob_Status_Transitions`

##### Scheduler (10 tests)
- `TestScheduler_CronTrigger`, `TestScheduler_IntervalTrigger`
- `TestScheduler_MissedTriggers`, `TestScheduler_Cleanup`
- `TestScheduler_Concurrency`, `TestScheduler_Timezones`

**Estimated Tests:** 45 test functions, 700 lines

### 3.4 MCP Tools

Priority tool files in [internal/mcp/tools/](internal/mcp/tools/):

#### [ddl.go](internal/mcp/tools/ddl.go) (1,086 lines)
- `TestDDL_Execute_CreateTable`, `TestDDL_Execute_AlterTable`, `TestDDL_Execute_DropTable`
- `TestDDL_Rollback`, `TestDDL_Migration_Up`, `TestDDL_Migration_Down`
- `TestDDL_Validation_Safety`, `TestDDL_Validation_Permissions`
- Estimated: 25 tests, 450 lines

#### [branching.go](internal/mcp/tools/branching.go) (1,032 lines)
- `TestBranch_Create`, `TestBranch_Delete`, `TestBranch_Switch`
- `TestBranch_Merge`, `TestBranch_List`, `TestBranch_Metadata`
- `TestBranch_DataClone_SchemaOnly`, `TestBranch_DataClone_Full`
- Estimated: 30 tests, 550 lines

#### [query_table.go](internal/mcp/tools/query_table.go)
- `TestQuery_Select`, `TestQuery_Insert`, `TestQuery_Update`, `TestQuery_Delete`
- `TestQuery_Transaction`, `TestQuery_Batch`, `TestQuery_ErrorHandling`
- Estimated: 20 tests, 350 lines

#### [storage.go](internal/mcp/tools/storage.go)
- `TestStorage_Upload`, `TestStorage_Download`, `TestStorage_List`, `TestStorage_Delete`
- `TestStorage_PublicURL`, `TestStorage_Transform`
- Estimated: 20 tests, 350 lines

#### [vectors.go](internal/mcp/tools/vectors.go)
- `TestVector_Search`, `TestVector_Insert`, `TestVector_Delete`, `TestVector_Update`
- `TestVector_Batch`, `TestVector_Filter`, `TestVector_Hybrid`
- Estimated: 20 tests, 350 lines

**Phase 3 Subtotal:** 320 new test functions, ~5,100 lines

---

## Phase 4: Complete Coverage (80% → 90%)

**Target Coverage:** 90% overall
**Estimated Duration:** 2-3 weeks
**Focus:** Remaining modules, edge cases, integration scenarios

### 4.1 Storage Module

#### [internal/storage/local.go](internal/storage/local.go) (1,257 lines, 360 uncovered)
- `TestLocalStorage_Upload`, `TestLocalStorage_Download`, `TestLocalStorage_RangeRequests`
- `TestLocalStorage_MultipartUpload`, `TestLocalStorage_Delete`, `TestLocalStorage_List`
- `TestLocalStorage_PublicURL`, `TestLocalStorage_PresignedURL`
- `TestStorageService_PresignedURL_Expiry`, `TestStorageService_AccessControl`
- Estimated: 30 tests, 550 lines

#### [internal/storage/service.go](internal/storage/service.go)
- `TestStorageService_Transformation_Image`, `TestStorageService_Transformation_Document`
- `TestStorageService_Caching`, `TestStorageService_CDN`
- `TestStorageService_Encryption`, `TestStorageService_Compression`
- Estimated: 25 tests, 450 lines

### 4.2 Realtime Module

#### [internal/realtime/subscription.go](internal/realtime/subscription.go) (984 lines, 223 uncovered)
- `TestSubscription_Connect`, `TestSubscription_Disconnect`
- `TestSubscription_Authentication_Valid`, `TestSubscription_Authentication_Invalid`
- `TestSubscription_Authorization_Allowed`, `TestSubscription_Authorization_Denied`
- `TestSubscription_Filters_Column`, `TestSubscription_Filters_User`, `TestSubscription_Broadcast`
- `TestSubscription_Reconnect`, `TestSubscription_Heartbeat`
- Estimated: 35 tests, 650 lines

#### [internal/realtime/handler.go](internal/realtime/handler.go) (168 uncovered)
- `TestWebSocketHandler_Upgrades`, `TestWebSocketHandler_Messages`
- `TestWebSocketHandler_Heartbeat`, `TestWebSocketHandler_Reconnect`
- `TestWebSocketHandler_ErrorHandling`, `TestWebSocketHandler_Close`
- Estimated: 20 tests, 350 lines

### 4.3 Branching Module

#### [internal/branching/storage.go](internal/branching/storage.go) (954 lines, 189 uncovered)
- `TestBranchStorage_Create`, `TestBranchStorage_Delete`, `TestBranchStorage_List`
- `TestBranchStorage_Metadata`, `TestBranchStorage_Update`
- `TestBranchStorage_AccessControl`, `TestBranchStorage_Permissions`
- Estimated: 25 tests, 450 lines

#### [internal/branching/manager.go](internal/branching/manager.go) (165 uncovered)
- `TestBranchManager_CreateDatabase`, `TestBranchManager_DropDatabase`
- `TestBranchManager_CloneSchema`, `TestBranchManager_CloneData`
- `TestBranchManager_SwitchConnection`, `TestBranchManager_Cleanup`
- Estimated: 20 tests, 400 lines

### 4.4 RPC Module

#### [internal/rpc/handler.go](internal/rpc/handler.go) (227 uncovered)
- `TestRPC_Execute`, `TestRPC_Parameters_Positional`, `TestRPC_Parameters_Named`
- `TestRPC_Transactions_Begin`, `TestRPC_Transactions_Commit`, `TestRPC_Transactions_Rollback`
- `TestRPC_Errors_SQL`, `TestRPC_Errors_Runtime`, `TestRPC_Timeout`
- Estimated: 25 tests, 450 lines

### 4.5 Webhook Module

#### [internal/webhook/webhook.go](internal/webhook/webhook.go) (987 lines, 253 uncovered)
- Additional edge case tests for existing coverage
- `TestWebhook_Retry_ExponentialBackoff`, `TestWebhook_Retry_MaxAttempts`
- `TestWebhook_Signature_Verification`, `TestWebhook_Payload_Validation`
- Estimated: 20 tests, 350 lines

### 4.6 Remaining Middleware & Utilities

- **Rate limiting edge cases** (15 tests, 250 lines)
  - [internal/middleware/rate_limiter.go](internal/middleware/rate_limiter.go)
  - `TestRateLimit_SlidingWindow`, `TestRateLimit_Distributed`, `TestRateLimit_Burst`

- **CSRF token validation** (10 tests, 180 lines)
  - `TestCSRF_Generate`, `TestCSRF_Validate`, `TestCSRF_Expiry`

- **Idempotency key handling** (12 tests, 220 lines)
  - `TestIdempotency_Cache`, `TestIdempotency_Replay`, `TestIdempotency_Conflict`

- **Feature flags** (8 tests, 150 lines)
  - `TestFeatureFlag_Enabled`, `TestFeatureFlag_Disabled`, `TestFeatureFlag_Percentage`

- **Security headers** (10 tests, 180 lines)
  - `TestSecurityHeaders_CSP`, `TestSecurityHeaders_HSTS`, `TestSecurityHeaders_CORS`

- **Body limit validation** (10 tests, 180 lines)
  - `TestBodyLimit_Small`, `TestBodyLimit_Large`, `TestBodyLimit_Exceeded`

- **ETag caching** (12 tests, 220 lines)
  - `TestETag_Generate`, `TestETag_Validate`, `TestETag_Weak`

- **Tracing propagation** (10 tests, 180 lines)
  - `TestTracing_Span`, `TestTracing_Context`, `TestTracing_Baggage`

**Phase 4 Subtotal:** 267 new test functions, ~4,200 lines

---

## Test Infrastructure & Mocks

### Extend [internal/testutil/mocks.go](internal/testutil/mocks.go)

```go
// Add to existing mocks file:

// MockDBConnection for database operations
type MockDBConnection struct {
    mu     sync.RWMutex
    QueryFunc  func(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error)
    ExecFunc   func(ctx context.Context, query string, args ...interface{}) (pgx.CommandTag, error)
    // ... existing methods
}

// MockOAuthProvider for OAuth testing
type MockOAuthProvider struct {
    mu         sync.RWMutex
    AuthURLFunc    func(state string) string
    ExchangeFunc   func(token string) (*oauth2.Token, error)
    GetUserFunc    func(token *oauth2.Token) (*User, error)
    // Add callback configuration
    OnAuthURL    func(state string) string
    OnExchange   func(token string) (*oauth2.Token, error)
    OnGetUser    func(token *oauth2.Token) (*User, error)
}

// MockTOTPValidator for MFA testing
type MockTOTPValidator struct {
    mu           sync.RWMutex
    ValidateFunc func(secret, code string) bool
    GenerateFunc func() (secret string, qrCode []byte, err error)
    ValidCodes   []string // Store valid codes for testing
}

// MockVectorDatabase for AI testing
type MockVectorDatabase struct {
    mu          sync.RWMutex
    InsertFunc  func(vectors []Vector) error
    SearchFunc  func(query Vector, limit int, filters map[string]string) ([]SearchResult, error)
    DeleteFunc  func(ids []string) error
}

// MockRuntime for Deno testing
type MockRuntime struct {
    mu          sync.RWMutex
    ExecuteFunc func(code string, env map[string]string) (string, error)
    BundleFunc  func(entryPoint string) (string, []byte, error)
}
```

### New Test Helpers in [internal/testutil/helpers.go](internal/testutil/helpers.go)

```go
// Common test setup helpers

func SetupTestDB(t *testing.T) *pgxpool.Pool
func CleanupTestDB(t *testing.T, db *pgxpool.Pool)
func CreateTestUser(t *testing.T, db *pgxpool.Pool, email string) uuid.UUID
func CreateTestTable(t *testing.T, db *pgxpool.Pool, schema, table string, columns []Column)
func AssertCoverage(t *testing.T, packagePath string, minPercent float64)
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration, msg string)
func MockFiberContext(method, path string, body io.Reader) *fiber.Ctx
func SetupTestServer(t *testing.T) *fiber.App
```

---

## Implementation Guidelines

### Test File Structure Pattern

Based on existing patterns in [internal/auth/service_test.go](internal/auth/service_test.go):

```go
package module

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"
)

// =============================================================================
// Component Name Tests
// =============================================================================

func TestComponent_Scenario_ExpectedBehavior(t *testing.T) {
    // Arrange
    setup := setupTest()
    defer setup.tearDown()

    // Act
    result := setup.component.DoSomething()

    // Assert
    assert.Equal(t, expected, result)
}

// Table-driven tests for multiple scenarios
func TestComponent_MultipleScenarios(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
        wantErr  bool
        errMsg   string
    }{
        {
            name:     "valid input returns success",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:     "invalid input returns error",
            input:    invalidInput,
            expected: ExpectedType{},
            wantErr:  true,
            errMsg:   "invalid input",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionUnderTest(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Test Naming Convention

```
Test[FunctionName]_[Scenario]_[ExpectedBehavior]

Examples:
- TestSignUp_Success
- TestSignUp_InvalidEmail
- TestSignUp_DuplicateEmail
- TestGetHandler_ValidQuery
- TestQueryParser_ParseFilters_EqualOperator
- TestWorker_ExecuteJob_Timeout
```

### Coverage Targets by Module

| Module | Current | Target | Priority | Files to Test |
|--------|---------|--------|----------|---------------|
| auth | 40% | 85% | CRITICAL | service.go, saml.go, dashboard.go, identity.go, password.go |
| api | 25% | 80% | CRITICAL | rest_crud.go, query_parser.go, server.go, auth_handler.go, graphql_*.go |
| ai | 20% | 75% | HIGH | handler.go, embedding_service.go, rag_service.go, knowledge_base_storage.go |
| functions | 35% | 80% | HIGH | handler.go, bundler.go, loader.go |
| jobs | 45% | 85% | HIGH | worker.go, manager.go, handler.go, scheduler.go |
| mcp | 15% | 75% | MEDIUM | tools/ddl.go, tools/branching.go, tools/query_table.go, tools/storage.go |
| storage | 30% | 80% | MEDIUM | local.go, service.go, s3.go |
| realtime | 25% | 75% | MEDIUM | subscription.go, handler.go, manager.go |
| branching | 40% | 80% | MEDIUM | storage.go, manager.go |
| rpc | 20% | 75% | LOW | handler.go, executor.go, validator.go |
| webhook | 50% | 80% | LOW | webhook.go, trigger.go |
| middleware | 55% | 85% | LOW | rate_limiter.go, csrf.go, idempotency.go |

---

## Execution Timeline

### Week 1-2: Phase 1 Foundation
- [ ] Expand auth service tests (75 tests)
  - [ ] Password workflows (15 tests)
  - [ ] MFA/TOTP flows (20 tests)
  - [ ] OAuth integration (10 tests)
  - [ ] Session management (12 tests)
  - [ ] User metadata (8 tests)
  - [ ] Impersonation (10 tests)
- [ ] Complete auth module coverage (120 tests)
  - [ ] SAML tests (40 tests)
  - [ ] Dashboard tests (35 tests)
  - [ ] Identity tests (25 tests)
  - [ ] Password tests (20 tests)
- [ ] Build mock infrastructure
  - [ ] MockOAuthProvider
  - [ ] MockTOTPValidator
  - [ ] MockRateLimiter
- **Coverage: 30% → 45%**

### Week 3-5: Phase 2 Core API
- [ ] REST API CRUD tests (58 tests)
  - [ ] GET handlers (15 tests)
  - [ ] POST/PUT/PATCH handlers (20 tests)
  - [ ] DELETE handlers (8 tests)
  - [ ] Embedded relations (10 tests)
  - [ ] Admin bypass (5 tests)
- [ ] Query parser expansion (85 tests)
  - [ ] Advanced filtering (25 tests)
  - [ ] Cursor pagination (15 tests)
  - [ ] Aggregations (20 tests)
  - [ ] Complex joins (10 tests)
  - [ ] SQL injection protection (15 tests)
- [ ] Server setup tests (53 tests)
  - [ ] Server initialization (15 tests)
  - [ ] Handler registration (20 tests)
  - [ ] Request lifecycle (10 tests)
  - [ ] Graceful shutdown (8 tests)
- [ ] GraphQL tests (65 tests)
  - [ ] Handler tests (25 tests)
  - [ ] Resolver tests (40 tests)
- **Coverage: 45% → 65%**

### Week 6-8: Phase 3 AI & Advanced
- [ ] AI module tests (120 tests)
  - [ ] Handler tests (30 tests)
  - [ ] Embedding service (20 tests)
  - [ ] RAG service (25 tests)
  - [ ] Knowledge base storage (30 tests)
  - [ ] OCR service (15 tests)
- [ ] Functions tests (70 tests)
  - [ ] Function invocation (25 tests)
  - [ ] Bundling (20 tests)
  - [ ] Scheduling (15 tests)
  - [ ] Storage/loading (10 tests)
- [ ] Jobs tests (45 tests)
  - [ ] Job execution (20 tests)
  - [ ] Job lifecycle (15 tests)
  - [ ] Scheduler (10 tests)
- [ ] MCP tools tests (115 tests)
  - [ ] DDL tools (25 tests)
  - [ ] Branching tools (30 tests)
  - [ ] Query tools (20 tests)
  - [ ] Storage tools (20 tests)
  - [ ] Vector tools (20 tests)
- **Coverage: 65% → 80%**

### Week 9-11: Phase 4 Completion
- [ ] Storage tests (55 tests)
  - [ ] Local storage (30 tests)
  - [ ] Storage service (25 tests)
- [ ] Realtime tests (55 tests)
  - [ ] Subscription tests (35 tests)
  - [ ] Handler tests (20 tests)
- [ ] Branching tests (45 tests)
  - [ ] Branching storage (25 tests)
  - [ ] Branching manager (20 tests)
- [ ] RPC/Webhook tests (45 tests)
  - [ ] RPC tests (25 tests)
  - [ ] Webhook tests (20 tests)
- [ ] Middleware edge cases (67 tests)
  - [ ] Rate limiting (15 tests)
  - [ ] CSRF (10 tests)
  - [ ] Idempotency (12 tests)
  - [ ] Feature flags (8 tests)
  - [ ] Security headers (10 tests)
  - [ ] Body limit (10 tests)
  - [ ] ETag (12 tests)
  - [ ] Tracing (10 tests)
- **Coverage: 80% → 90%**

### Week 12: Polish & Validation
- [ ] Run full test suite with race detector
- [ ] Fix flaky tests
- [ ] Performance optimization
- [ ] Documentation updates
- [ ] CI/CD integration
- **Final Coverage: 90%**

---

## Estimated Totals

**New Test Functions:** 1,043 tests
**New Test Lines:** ~16,400 lines
**Total Project Tests:** 4,736 tests (up from 3,693)
**Total Test Lines:** ~159,000 lines (up from 142,763)

**Files Requiring New Tests:** 50+ files
**New Mock Utilities:** 8 new mock types
**New Test Helpers:** 10 helper functions

---

## Success Metrics

### Coverage Milestones
- ✅ Week 2: 45% coverage (auth complete)
- ✅ Week 5: 65% coverage (API complete)
- ✅ Week 8: 80% coverage (AI/Functions complete)
- ✅ Week 11: 90% coverage (all modules complete)

### Quality Gates
- All tests pass with `-race` flag
- No flaky tests (>95% consistency over 10 runs)
- Test execution time < 5 minutes for unit tests
- Coverage report generated on every PR

### CI/CD Integration
```yaml
# .github/workflows/test-coverage.yml
name: Test Coverage
on: [pull_request]
jobs:
  coverage:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests with coverage
        run: make test-coverage
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
      - name: Check coverage thresholds
        run: go-test-coverage -c .testcoverage.yml
```

---

## Critical Files for Implementation

### Priority 1: Start Here (Week 1)
- [internal/auth/service.go](internal/auth/service.go) - Core authentication logic
- [internal/auth/service_test.go](internal/auth/service_test.go) - Existing test patterns
- [internal/testutil/mocks.go](internal/testutil/mocks.go) - Mock infrastructure

### Priority 2: Core API (Week 3-5)
- [internal/api/rest_crud.go](internal/api/rest_crud.go) - REST API handlers
- [internal/api/query_parser.go](internal/api/query_parser.go) - Query parsing
- [internal/api/auth_handler_test.go](internal/api/auth_handler_test.go) - Handler test patterns

### Priority 3: Server & Infrastructure (Week 3-5)
- [internal/api/server.go](internal/api/server.go) - Server setup
- [internal/middleware/rate_limiter_test.go](internal/middleware/rate_limiter_test.go) - Middleware patterns
- [internal/jobs/worker_test.go](internal/jobs/worker_test.go) - Worker patterns

### Priority 4: Advanced Features (Week 6-8)
- [internal/ai/handler.go](internal/ai/handler.go) - AI chat
- [internal/functions/handler_test.go](internal/functions/handler_test.go) - Function patterns
- [internal/api/graphql_handler.go](internal/api/graphql_handler.go) - GraphQL endpoint

---

## Risk Mitigation

### Potential Challenges

1. **Database Dependencies:** Many tests require PostgreSQL
   - Solution: Use existing test DB setup in [internal/database/](internal/database/), add transaction rollback helpers

2. **External Service Calls:** AI, OAuth, email providers
   - Solution: Comprehensive mocking, no real external calls in unit tests

3. **Concurrency Testing:** Race conditions in realtime, workers
   - Solution: Use `-race` flag, add synchronization tests

4. **Test Execution Time:** Growing suite may become slow
   - Solution: Parallel test execution (`t.Parallel()`), build tags for integration tests

5. **Flaky Tests:** Timing-dependent tests
   - Solution: Use `testify/assert` eventually, avoid hard-coded sleeps

---

## Maintenance Strategy

### Ongoing Coverage Requirements
1. **New Code:** All new features must include tests with minimum 70% coverage
2. **Bug Fixes:** Add regression test for every bug fixed
3. **Refactoring:** Update tests to maintain coverage during refactoring
4. **Monthly Audits:** Review coverage report, identify regressions

### Test Maintenance
1. **Weekly:** Run full test suite with coverage (`make test-coverage`)
2. **Monthly:** Update [`.testcoverage.yml`](.testcoverage.yml) thresholds incrementally
3. **Quarterly:** Review and refactor slow/flaky tests
4. **Documentation:** Keep test patterns documented in [CLAUDE.md](CLAUDE.md)

---

## Verification Steps

After completing each phase:

1. **Run Coverage:** `make test-coverage`
2. **Check Report:** `go tool cover -html=coverage.out -o coverage.html`
3. **Verify Thresholds:** `go-test-coverage -c .testcoverage.yml`
4. **Run with Race Detector:** `go test -race ./...`
5. **Check for Flaky Tests:** Run test suite 10 times, ensure 100% pass rate
6. **Measure Execution Time:** `time make test`

---

## Next Steps

1. **Start Phase 1:** Begin with auth service expansion (highest ROI)
2. **Weekly Check-ins:** Track progress against coverage milestones
3. **Adjust as Needed:** Adapt plan based on actual velocity
4. **Celebrate Milestones:** Acknowledge progress at each coverage target

---

## Appendix: Quick Reference

### Common Test Commands

```bash
# Run all tests with coverage
make test-coverage

# Run specific package tests
go test -v -cover ./internal/auth/...

# Run with race detector
go test -race ./internal/auth/...

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Check coverage for specific file
go tool cover -func=coverage.out | grep "auth/service.go"

# Run specific test
go test -v ./internal/auth/... -run TestSignUp_Success
```

### Test File Locations

| Module | Test Files |
|--------|------------|
| auth | `internal/auth/*_test.go` |
| api | `internal/api/*_test.go` |
| ai | `internal/ai/*_test.go` |
| functions | `internal/functions/*_test.go` |
| jobs | `internal/jobs/*_test.go` |
| mcp | `internal/mcp/**/*_test.go` |
| storage | `internal/storage/*_test.go` |
| realtime | `internal/realtime/*_test.go` |
| branching | `internal/branching/*_test.go` |
| rpc | `internal/rpc/*_test.go` |
| webhook | `internal/webhook/*_test.go` |
| middleware | `internal/middleware/*_test.go` |
| testutil | `internal/testutil/*_test.go` |

---

*This plan provides a clear, incremental path from 30% to 90% coverage while prioritizing business-critical modules and reusing existing test infrastructure. The phased approach allows for regular progress checks and course corrections throughout the 12-week journey.*
