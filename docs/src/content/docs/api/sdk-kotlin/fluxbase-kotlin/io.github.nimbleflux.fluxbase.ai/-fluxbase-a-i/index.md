---
title: "FluxbaseAI"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[FluxbaseAI](./)

# FluxbaseAI

[jvm]\
class [FluxbaseAI](./)(http: [FluxbaseHttpClient](../../iogithubnimblefluxfluxbasecore/-fluxbase-http-client/))

AI module — chatbots, conversations, and usage. Mirrors the user AI surface of the TS SDK (`sdk/src/ai.ts`, `FluxbaseAI`). The WebSocket chat stream is not covered here; use the TS/React SDKs or the `/ai/ws` protocol directly for streaming chat.

## Constructors

| | |
|---|---|
| [FluxbaseAI](-fluxbase-a-i/) | [jvm]<br>constructor(http: [FluxbaseHttpClient](../../iogithubnimblefluxfluxbasecore/-fluxbase-http-client/)) |

## Functions

| Name | Summary |
|---|---|
| [deleteConversation](delete-conversation/) | [jvm]<br>suspend fun [deleteConversation](delete-conversation/)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[Unit](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-unit/index.html)&gt;<br>DELETE `/api/v1/ai/conversations/:id` — delete a conversation. |
| [getChatbot](get-chatbot/) | [jvm]<br>suspend fun [getChatbot](get-chatbot/)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIChatbotSummary](../-a-i-chatbot-summary/)&gt;<br>GET `/api/v1/ai/chatbots/:id` — get a chatbot by ID. |
| [getConversation](get-conversation/) | [jvm]<br>suspend fun [getConversation](get-conversation/)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationDetail](../-a-i-conversation-detail/)&gt;<br>GET `/api/v1/ai/conversations/:id` — conversation with full message history. |
| [getUsage](get-usage/) | [jvm]<br>suspend fun [getUsage](get-usage/)(chatbotId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIUsageSnapshot](../-a-i-usage-snapshot/)&gt;<br>GET `/api/v1/ai/usage/:chatbotId` — per-user daily usage snapshot. |
| [listChatbots](list-chatbots/) | [jvm]<br>suspend fun [listChatbots](list-chatbots/)(namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[AIChatbotSummary](../-a-i-chatbot-summary/)&gt;&gt;<br>GET `/api/v1/ai/chatbots` — list public/enabled chatbots. |
| [listConversations](list-conversations/) | [jvm]<br>suspend fun [listConversations](list-conversations/)(chatbotId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, limit: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, offset: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationListResult](../-a-i-conversation-list-result/)&gt;<br>GET `/api/v1/ai/conversations` — list the current user's conversations. |
| [lookupChatbot](lookup-chatbot/) | [jvm]<br>suspend fun [lookupChatbot](lookup-chatbot/)(name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIChatbotSummary](../-a-i-chatbot-summary/)&gt;<br>GET `/api/v1/ai/chatbots/by-name/:name?namespace=…` — resolve a chatbot by name. |
| [updateConversation](update-conversation/) | [jvm]<br>suspend fun [updateConversation](update-conversation/)(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), request: [UpdateConversationRequest](../-update-conversation-request/)): [FluxbaseResponse](../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[AIConversationDetail](../-a-i-conversation-detail/)&gt;<br>PATCH `/api/v1/ai/conversations/:id` — update a conversation (e.g. rename). |