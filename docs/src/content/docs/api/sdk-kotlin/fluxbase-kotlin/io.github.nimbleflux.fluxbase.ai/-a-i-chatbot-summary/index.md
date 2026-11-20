---
title: "AIChatbotSummary"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIChatbotSummary](./)

# AIChatbotSummary

[jvm]\
@Serializable

data class [AIChatbotSummary](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;, val description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val model: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, val isPublic: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, val source: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;api&quot;, val updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [AIChatbotSummary](-a-i-chatbot-summary/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;, description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, model: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, isPublic: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, source: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;api&quot;, updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [description](description/) | [jvm]<br>val [description](description/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [enabled](enabled/) | [jvm]<br>val [enabled](enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [isPublic](is-public/) | [jvm]<br>@SerialName(value = &quot;is_public&quot;)<br>val [isPublic](is-public/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) |
| [model](model/) | [jvm]<br>val [model](model/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [name](name/) | [jvm]<br>val [name](name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [namespace](namespace/) | [jvm]<br>val [namespace](namespace/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [source](source/) | [jvm]<br>val [source](source/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [updatedAt](updated-at/) | [jvm]<br>@SerialName(value = &quot;updated_at&quot;)<br>val [updatedAt](updated-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |