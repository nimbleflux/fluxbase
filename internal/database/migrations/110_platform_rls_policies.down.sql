--
-- ROLLBACK: PLATFORM SCHEMA RLS POLICIES
-- Removes RLS policies from platform schema tables
--

-- PLATFORM.USERS
DROP POLICY IF EXISTS platform_users_service_all ON platform.users;
DROP POLICY IF EXISTS platform_users_instance_admin ON platform.users;
DROP POLICY IF EXISTS platform_users_self ON platform.users;

-- PLATFORM.SESSIONS
DROP POLICY IF EXISTS platform_sessions_service_all ON platform.sessions;
DROP POLICY IF EXISTS platform_sessions_instance_admin ON platform.sessions;
DROP POLICY IF EXISTS platform_sessions_self ON platform.sessions;

-- PLATFORM.OAUTH_PROVIDERS
DROP POLICY IF EXISTS platform_oauth_providers_service_all ON platform.oauth_providers;
DROP POLICY IF EXISTS platform_oauth_providers_instance_admin ON platform.oauth_providers;
DROP POLICY IF EXISTS platform_oauth_providers_read ON platform.oauth_providers;

-- PLATFORM.SSO_IDENTITIES
DROP POLICY IF EXISTS platform_sso_identities_service_all ON platform.sso_identities;
DROP POLICY IF EXISTS platform_sso_identities_instance_admin ON platform.sso_identities;
DROP POLICY IF EXISTS platform_sso_identities_self ON platform.sso_identities;

-- PLATFORM.ACTIVITY_LOG
DROP POLICY IF EXISTS platform_activity_log_service_all ON platform.activity_log;
DROP POLICY IF EXISTS platform_activity_log_instance_admin ON platform.activity_log;
DROP POLICY IF EXISTS platform_activity_log_self ON platform.activity_log;

-- PLATFORM.PASSWORD_RESET_TOKENS
DROP POLICY IF EXISTS platform_password_reset_tokens_service_all ON platform.password_reset_tokens;
DROP POLICY IF EXISTS platform_password_reset_tokens_instance_admin ON platform.password_reset_tokens;
DROP POLICY IF EXISTS platform_password_reset_tokens_create ON platform.password_reset_tokens;
DROP POLICY IF EXISTS platform_password_reset_tokens_self ON platform.password_reset_tokens;

-- PLATFORM.EMAIL_VERIFICATION_TOKENS
DROP POLICY IF EXISTS platform_email_verification_tokens_service_all ON platform.email_verification_tokens;
DROP POLICY IF EXISTS platform_email_verification_tokens_instance_admin ON platform.email_verification_tokens;
DROP POLICY IF EXISTS platform_email_verification_tokens_create ON platform.email_verification_tokens;
DROP POLICY IF EXISTS platform_email_verification_tokens_self ON platform.email_verification_tokens;

-- PLATFORM.EMAIL_TEMPLATES
DROP POLICY IF EXISTS platform_email_templates_service_all ON platform.email_templates;
DROP POLICY IF EXISTS platform_email_templates_instance_admin ON platform.email;

-- PLATFORM.INVITATION_TOKENS
DROP POLICY IF EXISTS platform_invitation_tokens_service_all ON platform.invitation_tokens;
DROP POLICY IF EXISTS platform_invitation_tokens_instance_admin ON platform.invitation_tokens;
DROP POLICY IF EXISTS platform_invitation_tokens_self ON platform.invitation_tokens;
