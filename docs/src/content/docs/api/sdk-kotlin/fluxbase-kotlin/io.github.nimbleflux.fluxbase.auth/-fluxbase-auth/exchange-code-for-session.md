---
title: "exchangeCodeForSession"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.auth](../../)/[FluxbaseAuth](../)/[exchangeCodeForSession](./)

# exchangeCodeForSession

[jvm]\
suspend fun [exchangeCodeForSession](./)(code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), state: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AuthResult](../../-auth-result/)&gt;

Exchange the OAuth authorization code (from the deep-link/callback) for a session and establish it. GETs `/api/v1/auth/oauth/{provider}/callback?code&state&redirect_uri`. Port of `exchangeCodeForSession()` (auth.ts:955).

Requires a preceding [getOAuthUrl](../get-o-auth-url/) call (for the stored provider).