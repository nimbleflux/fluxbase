---
title: "KnowledgeBase"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KnowledgeBase](./)

# KnowledgeBase

[jvm]\
@Serializable

data class [KnowledgeBase](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;, val description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingDimensions: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 1536, val chunkSize: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 512, val chunkOverlap: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 50, val chunkStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;recursive&quot;, val enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, val documentCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, val totalChunks: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, val visibility: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val entityExtractionEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, val entityExtractionStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val rerankEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, val createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [KnowledgeBase](-knowledge-base/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;default&quot;, description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingDimensions: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 1536, chunkSize: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 512, chunkOverlap: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 50, chunkStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) = &quot;recursive&quot;, enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = true, documentCount: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, totalChunks: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) = 0, visibility: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, entityExtractionEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, entityExtractionStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, rerankEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, createdAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, updatedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [chunkOverlap](chunk-overlap/) | [jvm]<br>@SerialName(value = &quot;chunk_overlap&quot;)<br>val [chunkOverlap](chunk-overlap/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [chunkSize](chunk-size/) | [jvm]<br>@SerialName(value = &quot;chunk_size&quot;)<br>val [chunkSize](chunk-size/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [chunkStrategy](chunk-strategy/) | [jvm]<br>@SerialName(value = &quot;chunk_strategy&quot;)<br>val [chunkStrategy](chunk-strategy/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [createdAt](created-at/) | [jvm]<br>@SerialName(value = &quot;created_at&quot;)<br>val [createdAt](created-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [description](description/) | [jvm]<br>val [description](description/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [documentCount](document-count/) | [jvm]<br>@SerialName(value = &quot;document_count&quot;)<br>val [documentCount](document-count/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [embeddingDimensions](embedding-dimensions/) | [jvm]<br>@SerialName(value = &quot;embedding_dimensions&quot;)<br>val [embeddingDimensions](embedding-dimensions/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [embeddingModel](embedding-model/) | [jvm]<br>@SerialName(value = &quot;embedding_model&quot;)<br>val [embeddingModel](embedding-model/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [enabled](enabled/) | [jvm]<br>val [enabled](enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) |
| [entityExtractionEnabled](entity-extraction-enabled/) | [jvm]<br>@SerialName(value = &quot;entity_extraction_enabled&quot;)<br>val [entityExtractionEnabled](entity-extraction-enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [entityExtractionStrategy](entity-extraction-strategy/) | [jvm]<br>@SerialName(value = &quot;entity_extraction_strategy&quot;)<br>val [entityExtractionStrategy](entity-extraction-strategy/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [name](name/) | [jvm]<br>val [name](name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [namespace](namespace/) | [jvm]<br>val [namespace](namespace/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [rerankEnabled](rerank-enabled/) | [jvm]<br>@SerialName(value = &quot;rerank_enabled&quot;)<br>val [rerankEnabled](rerank-enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [totalChunks](total-chunks/) | [jvm]<br>@SerialName(value = &quot;total_chunks&quot;)<br>val [totalChunks](total-chunks/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |
| [updatedAt](updated-at/) | [jvm]<br>@SerialName(value = &quot;updated_at&quot;)<br>val [updatedAt](updated-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [visibility](visibility/) | [jvm]<br>val [visibility](visibility/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |