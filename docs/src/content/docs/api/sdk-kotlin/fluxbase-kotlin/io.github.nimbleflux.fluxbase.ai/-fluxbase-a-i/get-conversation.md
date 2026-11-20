---
title: "getConversation"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[getConversation](./)

# getConversation

[jvm]\
suspend fun [getConversation](./)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationDetail](../../-a-i-conversation-detail/)&gt;

GET `/api/v1/ai/conversations/:id` — conversation with full message history.