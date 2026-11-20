---
title: "getUsage"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[getUsage](./)

# getUsage

[jvm]\
suspend fun [getUsage](./)(chatbotId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIUsageSnapshot](../../-a-i-usage-snapshot/)&gt;

GET `/api/v1/ai/usage/:chatbotId` — per-user daily usage snapshot.