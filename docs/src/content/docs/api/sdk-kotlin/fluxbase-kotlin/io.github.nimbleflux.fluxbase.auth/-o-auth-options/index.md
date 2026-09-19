---
title: "OAuthOptions"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.auth](../)/[OAuthOptions](./)

# OAuthOptions

[jvm]\
data class [OAuthOptions](./)(val redirectTo: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val redirectUri: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val scopes: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList())

Options for [FluxbaseAuth.getOAuthUrl](../-fluxbase-auth/get-o-auth-url/). Port of `OAuthOptions` (types.ts:872).

## Constructors

| | |
|---|---|
| [OAuthOptions](-o-auth-options/) | [jvm]<br>constructor(redirectTo: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, redirectUri: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, scopes: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList()) |

## Properties

| Name | Summary |
|---|---|
| [redirectTo](redirect-to/) | [jvm]<br>val [redirectTo](redirect-to/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)?<br>Post-login redirect URL (where to go after successful login). |
| [redirectUri](redirect-uri/) | [jvm]<br>val [redirectUri](redirect-uri/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)?<br>OAuth callback URL (where the provider redirects with the code). |
| [scopes](scopes/) | [jvm]<br>val [scopes](scopes/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; |