---
title: "getOAuthProviders"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.auth](../../)/[FluxbaseAuth](../)/[getOAuthProviders](./)

# getOAuthProviders

[jvm]\
suspend fun [getOAuthProviders](./)(): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[OAuthProviderInfo](../../-o-auth-provider-info/)&gt;&gt;

List the app-login-enabled OAuth providers. GETs `/api/v1/auth/oauth/providers`. Port of `getOAuthProviders()` (auth.ts:913).