---
title: "listChatbots"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[listChatbots](./)

# listChatbots

[jvm]\
suspend fun [listChatbots](./)(namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[AIChatbotSummary](../../-a-i-chatbot-summary/)&gt;&gt;

GET `/api/v1/ai/chatbots` — list public/enabled chatbots.