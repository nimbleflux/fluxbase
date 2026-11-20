---
title: "KBSearchResult"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KBSearchResult](./)

# KBSearchResult

[jvm]\
@Serializable

data class [KBSearchResult](./)(val chunkId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val documentId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val documentTitle: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val knowledgeBaseName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val similarity: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html), val metadata: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? = null)

## Constructors

| | |
|---|---|
| [KBSearchResult](-k-b-search-result/) | [jvm]<br>constructor(chunkId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), documentId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, documentTitle: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, knowledgeBaseName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), similarity: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html), metadata: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? = null) |

## Properties

| Name | Summary |
|---|---|
| [chunkId](chunk-id/) | [jvm]<br>@SerialName(value = &quot;chunk_id&quot;)<br>val [chunkId](chunk-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [content](content/) | [jvm]<br>val [content](content/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [documentId](document-id/) | [jvm]<br>@SerialName(value = &quot;document_id&quot;)<br>val [documentId](document-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [documentTitle](document-title/) | [jvm]<br>@SerialName(value = &quot;document_title&quot;)<br>val [documentTitle](document-title/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [knowledgeBaseName](knowledge-base-name/) | [jvm]<br>@SerialName(value = &quot;knowledge_base_name&quot;)<br>val [knowledgeBaseName](knowledge-base-name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [metadata](metadata/) | [jvm]<br>val [metadata](metadata/): [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? |
| [similarity](similarity/) | [jvm]<br>val [similarity](similarity/): [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html) |