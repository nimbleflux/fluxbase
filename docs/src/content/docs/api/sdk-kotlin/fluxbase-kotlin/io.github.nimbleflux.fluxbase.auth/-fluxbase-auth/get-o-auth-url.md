---
title: "getOAuthUrl"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.auth](../../)/[FluxbaseAuth](../)/[getOAuthUrl](./)

# getOAuthUrl

[jvm]\
suspend fun [getOAuthUrl](./)(provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), options: [OAuthOptions](../../-o-auth-options/) = OAuthOptions()): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[OAuthUrlResponse](../../-o-auth-url-response/)&gt;

Get the authorization URL for [provider](./) to open in a system browser. GETs `/api/v1/auth/oauth/{provider}/authorize`. Port of `getOAuthUrl()` (auth.ts:923).

The provider and [OAuthOptions.redirectUri](../../-o-auth-options/redirect-uri/) are remembered (storage) for the matching [exchangeCodeForSession](../exchange-code-for-session/) call.