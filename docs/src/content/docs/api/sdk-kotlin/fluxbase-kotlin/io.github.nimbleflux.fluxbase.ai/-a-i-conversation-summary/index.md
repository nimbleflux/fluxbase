---
title: "AIConversationSummary"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIConversationSummary](./)

# AIConversationSummary

[jvm]\
@Serializable

data class [AIConversationSummary](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val chatbot: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val preview: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val messageCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, val createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

User-view conversation summary (GET /api/v1/ai/conversations).

## Constructors

| | |
|---|---|
| [AIConversationSummary](-a-i-conversation-summary/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), chatbot: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, preview: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, messageCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [chatbot](chatbot/) | [jvm]<br>val [chatbot](chatbot/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)?<br>Chatbot name (not ID) in the user API. |
| [createdAt](created-at/) | [jvm]<br>@SerialName(value = &quot;created_at&quot;)<br>val [createdAt](created-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [messageCount](message-count/) | [jvm]<br>@SerialName(value = &quot;message_count&quot;)<br>val [messageCount](message-count/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [namespace](namespace/) | [jvm]<br>val [namespace](namespace/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [preview](preview/) | [jvm]<br>val [preview](preview/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [title](title/) | [jvm]<br>val [title](title/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [updatedAt](updated-at/) | [jvm]<br>@SerialName(value = &quot;updated_at&quot;)<br>val [updatedAt](updated-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |