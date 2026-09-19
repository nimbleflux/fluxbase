---
title: "FluxbaseAuth"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.auth](../)/[FluxbaseAuth](./)

# FluxbaseAuth

class [FluxbaseAuth](./)(http: [FluxbaseHttpClient](../../iogithubnimblefluxfluxbasecore/-fluxbase-http-client/), autoRefresh: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, storage: [StorageAdapter](../-storage-adapter/) = MemoryStorage())

Authentication module — port of `FluxbaseAuth` from `sdk/src/auth.ts`.

Manages the user session: sign in/up/out, session persistence, and auth state change events. The session is stored via a [StorageAdapter](../-storage-adapter/) (default: [MemoryStorage](../-memory-storage/) for JVM; Android injects an EncryptedSharedPreferences-backed implementation).

#### Parameters

jvm

| | |
|---|---|
| http | the shared [FluxbaseHttpClient](../../iogithubnimblefluxfluxbasecore/-fluxbase-http-client/) for making API calls. |
| autoRefresh | whether to automatically refresh the token before expiry (default true; disabled in tests). TS default is true (`auth.ts:55`). |
| storage | the [StorageAdapter](../-storage-adapter/) for session persistence. |

## Constructors

| | |
|---|---|
| [FluxbaseAuth](-fluxbase-auth/) | [jvm]<br>constructor(http: [FluxbaseHttpClient](../../iogithubnimblefluxfluxbasecore/-fluxbase-http-client/), autoRefresh: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, storage: [StorageAdapter](../-storage-adapter/) = MemoryStorage()) |

## Types

| Name | Summary |
|---|---|
| [Companion](-companion/) | [jvm]<br>object [Companion](-companion/) |

## Properties

| Name | Summary |
|---|---|
| [currentSession](current-session/) | [jvm]<br>var [currentSession](current-session/): [AuthSession](../-auth-session/)?<br>The current session, or null if not authenticated. |
| [currentUser](current-user/) | [jvm]<br>val [currentUser](current-user/): [User](../-user/)?<br>The current user, or null if not authenticated. |

## Functions

| Name | Summary |
|---|---|
| [disable2FA](disable2-f-a/) | [jvm]<br>suspend fun [disable2FA](disable2-f-a/)(password: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[TwoFactorDisableResponse](../-two-factor-disable-response/)&gt;<br>POST `/api/v1/auth/2fa/disable` with `{password}` → disables 2FA. |
| [enable2FA](enable2-f-a/) | [jvm]<br>suspend fun [enable2FA](enable2-f-a/)(code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[TwoFactorEnableResponse](../-two-factor-enable-response/)&gt;<br>POST `/api/v1/auth/2fa/enable` with `{code}` → enables 2FA, returns backup codes. |
| [exchangeCodeForSession](exchange-code-for-session/) | [jvm]<br>suspend fun [exchangeCodeForSession](exchange-code-for-session/)(code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), state: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../-auth-result/)&gt;<br>Exchange the OAuth authorization code (from the deep-link/callback) for a session and establish it. GETs `/api/v1/auth/oauth/{provider}/callback?code&state&redirect_uri`. Port of `exchangeCodeForSession()` (auth.ts:955). |
| [get2FAStatus](get2-f-a-status/) | [jvm]<br>suspend fun [get2FAStatus](get2-f-a-status/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[TwoFactorStatusResponse](../-two-factor-status-response/)&gt;<br>GET `/api/v1/auth/2fa/status` → returns enrolled factors. |
| [getAuthConfig](get-auth-config/) | [jvm]<br>suspend fun [getAuthConfig](get-auth-config/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthConfig](../-auth-config/)&gt;<br>Get the server's auth configuration (signup enabled, OAuth providers, password rules). GETs `/api/v1/auth/config`. Port of `getAuthConfig()` in `auth.ts`. |
| [getCurrentUser](get-current-user/) | [jvm]<br>suspend fun [getCurrentUser](get-current-user/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[User](../-user/)&gt;<br>Get the current user from the server. GETs `/api/v1/auth/user`. Port of `getCurrentUser()` in `auth.ts`. |
| [getOAuthProviders](get-o-auth-providers/) | [jvm]<br>suspend fun [getOAuthProviders](get-o-auth-providers/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[OAuthProviderInfo](../-o-auth-provider-info/)&gt;&gt;<br>List the app-login-enabled OAuth providers. GETs `/api/v1/auth/oauth/providers`. Port of `getOAuthProviders()` (auth.ts:913). |
| [getOAuthUrl](get-o-auth-url/) | [jvm]<br>suspend fun [getOAuthUrl](get-o-auth-url/)(provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), options: [OAuthOptions](../-o-auth-options/) = OAuthOptions()): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[OAuthUrlResponse](../-o-auth-url-response/)&gt;<br>Get the authorization URL for [provider](get-o-auth-url/) to open in a system browser. GETs `/api/v1/auth/oauth/{provider}/authorize`. Port of `getOAuthUrl()` (auth.ts:923). |
| [onAuthStateChange](on-auth-state-change/) | [jvm]<br>fun [onAuthStateChange](on-auth-state-change/)(callback: ([AuthState](../-auth-state/)) -&gt; [Unit](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-unit/index.html)): () -&gt; [Unit](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-unit/index.html)<br>Register a callback for auth state changes. Returns a function to unsubscribe. |
| [refreshSession](refresh-session/) | [jvm]<br>suspend fun [refreshSession](refresh-session/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthSession](../-auth-session/)&gt;<br>Refresh the current session using the stored refresh token. POSTs to `/api/v1/auth/refresh` with `{refresh_token}`. On success, updates the session and emits `TOKEN_REFRESHED`. |
| [sendPasswordReset](send-password-reset/) | [jvm]<br>suspend fun [sendPasswordReset](send-password-reset/)(email: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[Unit](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-unit/index.html)&gt;<br>Send a password reset email. POSTs `/api/v1/auth/password/reset`. Port of `sendPasswordReset()` in `auth.ts`. |
| [setup2FA](setup2-f-a/) | [jvm]<br>suspend fun [setup2FA](setup2-f-a/)(): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[TwoFactorSetupResponse](../-two-factor-setup-response/)&gt;<br>POST `/api/v1/auth/2fa/setup` → returns TOTP secret + QR code. |
| [signIn](sign-in/) | [jvm]<br>suspend fun [signIn](sign-in/)(email: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), password: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../-auth-result/)&gt;<br>Sign in with email and password. |
| [signInWithPassword](sign-in-with-password/) | [jvm]<br>suspend fun [signInWithPassword](sign-in-with-password/)(email: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), password: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../-auth-result/)&gt;<br>Alias for [signIn](sign-in/) — Supabase-compatible method name. |
| [signOut](sign-out/) | [jvm]<br>suspend fun [signOut](sign-out/)()<br>Sign out. POSTs to `/api/v1/auth/signout`, clears the session, and restores the anon key on the HTTP client. |
| [signUp](sign-up/) | [jvm]<br>suspend fun [signUp](sign-up/)(email: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), password: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../-auth-result/)&gt;<br>Sign up with email and password. |
| [updateUser](update-user/) | [jvm]<br>suspend fun [updateUser](update-user/)(attributes: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)&gt;): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[User](../-user/)&gt;<br>Update the user's attributes. PATCHes `/api/v1/auth/user`. Port of `updateUser()` in `auth.ts`. |
| [verify2FA](verify2-f-a/) | [jvm]<br>suspend fun [verify2FA](verify2-f-a/)(userId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../-auth-result/)&gt;<br>POST `/api/v1/auth/2fa/verify` with `{user_id, code}` — completes a 2FA login challenge. On success, establishes a session from the returned tokens. |