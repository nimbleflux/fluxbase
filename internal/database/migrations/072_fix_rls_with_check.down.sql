-- Rollback: Revert to original policies without WITH CHECK
-- WARNING: This restores the security vulnerability where users could
-- INSERT/UPDATE rows with arbitrary user_ids. Only use if absolutely needed.

-- ============================================================================
-- AI Schema Policies
-- ============================================================================

DROP POLICY IF EXISTS "ai_user_prefs_own" ON ai.user_provider_preferences;
CREATE POLICY "ai_user_prefs_own" ON ai.user_provider_preferences
    FOR ALL TO authenticated
    USING (user_id = auth.current_user_id());

DROP POLICY IF EXISTS "ai_conversations_own" ON ai.conversations;
CREATE POLICY "ai_conversations_own" ON ai.conversations
    FOR ALL TO authenticated
    USING (user_id = auth.current_user_id());

DROP POLICY IF EXISTS "ai_messages_own" ON ai.messages;
CREATE POLICY "ai_messages_own" ON ai.messages
    FOR ALL TO authenticated
    USING (conversation_id IN (
        SELECT id FROM ai.conversations WHERE user_id = auth.current_user_id()
    ));

-- ============================================================================
-- Knowledge Base Schema Policies
-- ============================================================================

DROP POLICY IF EXISTS "ai_kb_dashboard_admin" ON ai.knowledge_bases;
CREATE POLICY "ai_kb_dashboard_admin" ON ai.knowledge_bases
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin');

DROP POLICY IF EXISTS "ai_documents_dashboard_admin" ON ai.documents;
CREATE POLICY "ai_documents_dashboard_admin" ON ai.documents
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin');

DROP POLICY IF EXISTS "ai_chunks_dashboard_admin" ON ai.chunks;
CREATE POLICY "ai_chunks_dashboard_admin" ON ai.chunks
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin');

DROP POLICY IF EXISTS "ai_chatbot_kb_dashboard_admin" ON ai.chatbot_knowledge_bases;
CREATE POLICY "ai_chatbot_kb_dashboard_admin" ON ai.chatbot_knowledge_bases
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin');
