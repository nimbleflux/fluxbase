---
title: "CreateKnowledgeBaseRequest"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[CreateKnowledgeBaseRequest](./)

# CreateKnowledgeBaseRequest

[jvm]\
@Serializable

data class [CreateKnowledgeBaseRequest](./)(val name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val visibility: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingDimensions: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, val chunkSize: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, val chunkOverlap: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, val chunkStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val entityExtractionEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, val entityExtractionStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val rerankEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null)

## Constructors

| | |
|---|---|
| [CreateKnowledgeBaseRequest](-create-knowledge-base-request/) | [jvm]<br>constructor(name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), namespace: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, description: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, visibility: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingDimensions: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, chunkSize: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, chunkOverlap: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, chunkStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, entityExtractionEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, entityExtractionStrategy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, rerankEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [chunkOverlap](chunk-overlap/) | [jvm]<br>@SerialName(value = &quot;chunk_overlap&quot;)<br>val [chunkOverlap](chunk-overlap/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? |
| [chunkSize](chunk-size/) | [jvm]<br>@SerialName(value = &quot;chunk_size&quot;)<br>val [chunkSize](chunk-size/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? |
| [chunkStrategy](chunk-strategy/) | [jvm]<br>@SerialName(value = &quot;chunk_strategy&quot;)<br>val [chunkStrategy](chunk-strategy/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [description](description/) | [jvm]<br>val [description](description/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [embeddingDimensions](embedding-dimensions/) | [jvm]<br>@SerialName(value = &quot;embedding_dimensions&quot;)<br>val [embeddingDimensions](embedding-dimensions/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? |
| [embeddingModel](embedding-model/) | [jvm]<br>@SerialName(value = &quot;embedding_model&quot;)<br>val [embeddingModel](embedding-model/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [entityExtractionEnabled](entity-extraction-enabled/) | [jvm]<br>@SerialName(value = &quot;entity_extraction_enabled&quot;)<br>val [entityExtractionEnabled](entity-extraction-enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [entityExtractionStrategy](entity-extraction-strategy/) | [jvm]<br>@SerialName(value = &quot;entity_extraction_strategy&quot;)<br>val [entityExtractionStrategy](entity-extraction-strategy/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [name](name/) | [jvm]<br>val [name](name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [namespace](namespace/) | [jvm]<br>val [namespace](namespace/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [rerankEnabled](rerank-enabled/) | [jvm]<br>@SerialName(value = &quot;rerank_enabled&quot;)<br>val [rerankEnabled](rerank-enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [visibility](visibility/) | [jvm]<br>val [visibility](visibility/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |