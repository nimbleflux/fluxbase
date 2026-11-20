---
title: "lookupChatbot"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.ai](../../)/[FluxbaseAI](../)/[lookupChatbot](./)

# lookupChatbot

[jvm]\
suspend fun [lookupChatbot](./)(name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIChatbotSummary](../../-a-i-chatbot-summary/)&gt;

GET `/api/v1/ai/chatbots/by-name/:name?namespace=…` — resolve a chatbot by name.