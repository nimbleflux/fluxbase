---
title: "KnowledgeBaseDocument"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KnowledgeBaseDocument](./)

# KnowledgeBaseDocument

[jvm]\
@Serializable

data class [KnowledgeBaseDocument](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val sourceType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val mimeType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val status: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val chunksCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, val tags: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList(), val createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [KnowledgeBaseDocument](-knowledge-base-document/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, sourceType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, mimeType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, status: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, chunksCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, tags: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList(), createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [chunksCount](chunks-count/) | [jvm]<br>@SerialName(value = &quot;chunks_count&quot;)<br>val [chunksCount](chunks-count/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [content](content/) | [jvm]<br>val [content](content/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [createdAt](created-at/) | [jvm]<br>@SerialName(value = &quot;created_at&quot;)<br>val [createdAt](created-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [knowledgeBaseId](knowledge-base-id/) | [jvm]<br>@SerialName(value = &quot;knowledge_base_id&quot;)<br>val [knowledgeBaseId](knowledge-base-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [mimeType](mime-type/) | [jvm]<br>@SerialName(value = &quot;mime_type&quot;)<br>val [mimeType](mime-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [sourceType](source-type/) | [jvm]<br>@SerialName(value = &quot;source_type&quot;)<br>val [sourceType](source-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [status](status/) | [jvm]<br>val [status](status/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [tags](tags/) | [jvm]<br>val [tags](tags/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; |
| [title](title/) | [jvm]<br>val [title](title/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [updatedAt](updated-at/) | [jvm]<br>@SerialName(value = &quot;updated_at&quot;)<br>val [updatedAt](updated-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |