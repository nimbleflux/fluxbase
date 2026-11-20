---
title: "AIConversationListResult"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIConversationListResult](./)

# AIConversationListResult

[jvm]\
@Serializable

data class [AIConversationListResult](./)(val conversations: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[AIConversationSummary](../-a-i-conversation-summary/)&gt; = emptyList(), val total: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, val hasMore: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = false)

## Constructors

| | |
|---|---|
| [AIConversationListResult](-a-i-conversation-list-result/) | [jvm]<br>constructor(conversations: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[AIConversationSummary](../-a-i-conversation-summary/)&gt; = emptyList(), total: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, hasMore: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = false) |

## Properties

| Name | Summary |
|---|---|
| [conversations](conversations/) | [jvm]<br>val [conversations](conversations/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[AIConversationSummary](../-a-i-conversation-summary/)&gt; |
| [hasMore](has-more/) | [jvm]<br>@SerialName(value = &quot;has_more&quot;)<br>val [hasMore](has-more/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) |
| [total](total/) | [jvm]<br>val [total](total/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |