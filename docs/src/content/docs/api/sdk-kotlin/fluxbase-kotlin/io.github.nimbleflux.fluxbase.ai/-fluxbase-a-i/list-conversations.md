---
title: "listConversations"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[listConversations](./)

# listConversations

[jvm]\
suspend fun [listConversations](./)(chatbotId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, limit: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, offset: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationListResult](../../-a-i-conversation-list-result/)&gt;

GET `/api/v1/ai/conversations` — list the current user's conversations.