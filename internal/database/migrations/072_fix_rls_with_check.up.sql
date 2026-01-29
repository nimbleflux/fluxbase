-- Migration: Add WITH CHECK clauses to user-facing FOR ALL policies
-- These policies currently allow users to INSERT/UPDATE rows with arbitrary user_ids
-- which is a security vulnerability.
--
-- Security fix: WITH CHECK ensures that when a row is INSERTed or UPDATEd,
-- the new row values must also satisfy the policy condition.

-- ============================================================================
-- AI Schema Policies (from 016_tables_ai.up.sql)
-- ============================================================================

-- AI user preferences - users can only manage their own preferences
DROP POLICY IF EXISTS "ai_user_prefs_own" ON ai.user_provider_preferences;
CREATE POLICY "ai_user_prefs_own" ON ai.user_provider_preferences
    FOR ALL TO authenticated
    USING (user_id = auth.current_user_id())
    WITH CHECK (user_id = auth.current_user_id());

-- AI conversations - users can only manage their own conversations
DROP POLICY IF EXISTS "ai_conversations_own" ON ai.conversations;
CREATE POLICY "ai_conversations_own" ON ai.conversations
    FOR ALL TO authenticated
    USING (user_id = auth.current_user_id())
    WITH CHECK (user_id = auth.current_user_id());

-- AI messages - users can only access messages in their own conversations
DROP POLICY IF EXISTS "ai_messages_own" ON ai.messages;
CREATE POLICY "ai_messages_own" ON ai.messages
    FOR ALL TO authenticated
    USING (conversation_id IN (
        SELECT id FROM ai.conversations WHERE user_id = auth.current_user_id()
    ))
    WITH CHECK (conversation_id IN (
        SELECT id FROM ai.conversations WHERE user_id = auth.current_user_id()
    ));

-- ============================================================================
-- Knowledge Base Schema Policies (from 030_tables_knowledge_base.up.sql)
-- ============================================================================

-- Dashboard admins can manage knowledge bases
DROP POLICY IF EXISTS "ai_kb_dashboard_admin" ON ai.knowledge_bases;
CREATE POLICY "ai_kb_dashboard_admin" ON ai.knowledge_bases
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin')
    WITH CHECK (auth.role() = 'dashboard_admin');

-- Dashboard admins can manage documents
DROP POLICY IF EXISTS "ai_documents_dashboard_admin" ON ai.documents;
CREATE POLICY "ai_documents_dashboard_admin" ON ai.documents
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin')
    WITH CHECK (auth.role() = 'dashboard_admin');

-- Dashboard admins can manage chunks
DROP POLICY IF EXISTS "ai_chunks_dashboard_admin" ON ai.chunks;
CREATE POLICY "ai_chunks_dashboard_admin" ON ai.chunks
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin')
    WITH CHECK (auth.role() = 'dashboard_admin');

-- Dashboard admins can manage chatbot-knowledge base associations
DROP POLICY IF EXISTS "ai_chatbot_kb_dashboard_admin" ON ai.chatbot_knowledge_bases;
CREATE POLICY "ai_chatbot_kb_dashboard_admin" ON ai.chatbot_knowledge_bases
    FOR ALL TO authenticated
    USING (auth.role() = 'dashboard_admin')
    WITH CHECK (auth.role() = 'dashboard_admin');
