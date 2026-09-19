---
title: "OAuthUrlResponse"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.auth](../)/[OAuthUrlResponse](./)

# OAuthUrlResponse

[jvm]\
@Serializable

data class [OAuthUrlResponse](./)(val url: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;&quot;)

The authorization URL to open in a browser. Port of `OAuthUrlResponse` from `sdk/src/types.ts:878`.

## Constructors

| | |
|---|---|
| [OAuthUrlResponse](-o-auth-url-response/) | [jvm]<br>constructor(url: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), provider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;&quot;) |

## Properties

| Name | Summary |
|---|---|
| [provider](provider/) | [jvm]<br>val [provider](provider/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [url](url/) | [jvm]<br>val [url](url/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |