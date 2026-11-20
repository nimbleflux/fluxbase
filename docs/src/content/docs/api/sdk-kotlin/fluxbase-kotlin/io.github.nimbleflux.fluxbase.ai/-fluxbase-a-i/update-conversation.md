---
title: "updateConversation"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[updateConversation](./)

# updateConversation

[jvm]\
suspend fun [updateConversation](./)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), request: [UpdateConversationRequest](../../-update-conversation-request/)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationDetail](../../-a-i-conversation-detail/)&gt;

PATCH `/api/v1/ai/conversations/:id` — update a conversation (e.g. rename).