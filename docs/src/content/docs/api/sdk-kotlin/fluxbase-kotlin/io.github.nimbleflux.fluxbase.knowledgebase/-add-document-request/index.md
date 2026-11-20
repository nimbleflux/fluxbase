---
title: "AddDocumentRequest"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[AddDocumentRequest](./)

# AddDocumentRequest

[jvm]\
@Serializable

data class [AddDocumentRequest](./)(val title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val sourceType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val sourceUrl: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val mimeType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val tags: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList(), val metadata: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? = null)

## Constructors

| | |
|---|---|
| [AddDocumentRequest](-add-document-request/) | [jvm]<br>constructor(title: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, content: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), sourceType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, sourceUrl: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, mimeType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, tags: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList(), metadata: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? = null) |

## Properties

| Name | Summary |
|---|---|
| [content](content/) | [jvm]<br>val [content](content/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [metadata](metadata/) | [jvm]<br>val [metadata](metadata/): [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), JsonElement&gt;? |
| [mimeType](mime-type/) | [jvm]<br>@SerialName(value = &quot;mime_type&quot;)<br>val [mimeType](mime-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [sourceType](source-type/) | [jvm]<br>@SerialName(value = &quot;source_type&quot;)<br>val [sourceType](source-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [sourceUrl](source-url/) | [jvm]<br>@SerialName(value = &quot;source_url&quot;)<br>val [sourceUrl](source-url/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [tags](tags/) | [jvm]<br>val [tags](tags/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; |
| [title](title/) | [jvm]<br>val [title](title/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |