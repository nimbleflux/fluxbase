---
title: "anonKey"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.auth](../../)/[AuthConfig](../)/[anonKey](./)

# anonKey

[jvm]\

@SerialName(value = &quot;anon_key&quot;)

val [anonKey](./): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)?

Publishable anon key for this instance (the same key the web app exposes to browsers). Lets native clients connect without manual key entry. Null on servers that don't publish it.