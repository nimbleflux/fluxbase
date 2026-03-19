--
-- MULTI-TENANCY: PLATFORM SCHEMA RLS POLICIES
-- Creates RLS for platform schema tables after dashboard rename
--

-- ============================================
-- PLATFORM.USERS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_users_service_all ON platform.users
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all platform users
CREATE POLICY platform_users_instance_admin ON platform.users
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can view their own record
CREATE POLICY platform_users_self ON platform.users
    FOR SELECT TO authenticated
    USING (id = auth.uid());

-- ============================================
-- PLATFORM.SESSIONS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_sessions_service_all ON platform.sessions
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all sessions
CREATE POLICY platform_sessions_instance_admin ON platform.sessions
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can view and delete their own sessions
CREATE POLICY platform_sessions_self ON platform.sessions
    FOR ALL TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.OAUTH_PROVIDERS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_oauth_providers_service_all ON platform.oauth_providers
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage OAuth providers
CREATE POLICY platform_oauth_providers_instance_admin ON platform.oauth_providers
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- All authenticated users can view enabled providers
CREATE POLICY platform_oauth_providers_read ON platform.oauth_providers
    FOR SELECT TO authenticated
    USING (enabled = true);

-- ============================================
-- PLATFORM.SSO_IDENTITIES RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_sso_identities_service_all ON platform.sso_identities
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all SSO identities
CREATE POLICY platform_sso_identities_instance_admin ON platform.sso_identities
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can view their own SSO identities
CREATE POLICY platform_sso_identities_self ON platform.sso_identities
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.ACTIVITY_LOG RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_activity_log_service_all ON platform.activity_log
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can view all activity
CREATE POLICY platform_activity_log_instance_admin ON platform.activity_log
    FOR SELECT TO authenticated
    USING (is_instance_admin(auth.uid()));

-- Users can view their own activity
CREATE POLICY platform_activity_log_self ON platform.activity_log
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.PASSWORD_RESET_TOKENS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_password_reset_tokens_service_all ON platform.password_reset_tokens
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all password reset tokens
CREATE POLICY platform_password_reset_tokens_instance_admin ON platform.password_reset_tokens
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can only create tokens for themselves
CREATE POLICY platform_password_reset_tokens_create ON platform.password_reset_tokens
    FOR INSERT TO authenticated
    WITH CHECK (user_id = auth.uid());

-- Users can view their own tokens
CREATE POLICY platform_password_reset_tokens_self ON platform.password_reset_tokens
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.EMAIL_VERIFICATION_TOKENS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_email_verification_tokens_service_all ON platform.email_verification_tokens
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage all email verification tokens
CREATE POLICY platform_email_verification_tokens_instance_admin ON platform.email_verification_tokens
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can only create tokens for themselves
CREATE POLICY platform_email_verification_tokens_create ON platform.email_verification_tokens
    FOR INSERT TO authenticated
    WITH CHECK (user_id = auth.uid());

-- Users can view their own tokens
CREATE POLICY platform_email_verification_tokens_self ON platform.email_verification_tokens
    FOR SELECT TO authenticated
    USING (user_id = auth.uid());

-- ============================================
-- PLATFORM.EMAIL_TEMPLATES RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_email_templates_service_all ON platform.email_templates
    FOR ALL TO service_role USING (true) WITH CHECK (true);

-- Instance admins can manage email templates
CREATE POLICY platform_email_templates_instance_admin ON platform.email_templates
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- All authenticated users can view email templates
CREATE POLICY platform_email_templates_read ON platform.email_templates
    FOR SELECT TO authenticated
    USING (true);

-- ============================================
-- PLATFORM.INVITATION_TOKENS RLS POLICIES
-- ============================================

-- Service role bypasses all
CREATE POLICY platform_invitation_tokens_service_all ON platform.invitation_tokens
    FOR ALL TO service_role USING (true) WITH check (true);

-- Instance admins can manage all invitation tokens
CREATE POLICY platform_invitation_tokens_instance_admin ON platform.invitation_tokens
    FOR ALL TO authenticated
    USING (is_instance_admin(auth.uid()))
    WITH CHECK (is_instance_admin(auth.uid()));

-- Users can view their own invitations
CREATE POLICY platform_invitation_tokens_self ON platform.invitation_tokens
    FOR SELECT TO authenticated
    USING (email = (SELECT email FROM platform.users WHERE id = auth.uid()));
