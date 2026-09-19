---
title: "OAuthProviderInfo"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.auth](../)/[OAuthProviderInfo](./)

# OAuthProviderInfo

[jvm]\
@Serializable

data class [OAuthProviderInfo](./)(val provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val displayName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;&quot;, val authorizeUrl: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

An app-login-enabled OAuth provider. Port of `OAuthProviderInfo` from `sdk/src/types.ts:861`.

## Constructors

| | |
|---|---|
| [OAuthProviderInfo](-o-auth-provider-info/) | [jvm]<br>constructor(provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), displayName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;&quot;, authorizeUrl: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [authorizeUrl](authorize-url/) | [jvm]<br>@SerialName(value = &quot;authorize_url&quot;)<br>val [authorizeUrl](authorize-url/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [displayName](display-name/) | [jvm]<br>@SerialName(value = &quot;display_name&quot;)<br>val [displayName](display-name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [provider](provider/) | [jvm]<br>val [provider](provider/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |