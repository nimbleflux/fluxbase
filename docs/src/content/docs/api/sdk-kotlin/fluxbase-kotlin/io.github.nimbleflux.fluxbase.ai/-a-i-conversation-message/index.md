---
title: "AIConversationMessage"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIConversationMessage](./)

# AIConversationMessage

[jvm]\
@Serializable

data class [AIConversationMessage](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val conversationId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val role: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [AIConversationMessage](-a-i-conversation-message/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, conversationId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, role: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [content](content/) | [jvm]<br>val [content](content/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [conversationId](conversation-id/) | [jvm]<br>@SerialName(value = &quot;conversation_id&quot;)<br>val [conversationId](conversation-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [createdAt](created-at/) | [jvm]<br>@SerialName(value = &quot;created_at&quot;)<br>val [createdAt](created-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [role](role/) | [jvm]<br>val [role](role/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |