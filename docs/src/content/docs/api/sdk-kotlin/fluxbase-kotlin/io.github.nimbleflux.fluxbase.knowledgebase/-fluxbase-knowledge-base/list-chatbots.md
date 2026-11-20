---
title: "listChatbots"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[listChatbots](./)

# listChatbots

[jvm]\
suspend fun [listChatbots](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[KBLinkedChatbot](../../-k-b-linked-chatbot/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/chatbots` — chatbots linked to this KB.