---
title: "VectorCapabilities"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.vector](../)/[VectorCapabilities](./)

# VectorCapabilities

[jvm]\
@Serializable

data class [VectorCapabilities](./)(val enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = false, val pgVectorInstalled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, val pgVectorVersion: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, val embeddingProvider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

Vector search capabilities. Admin callers receive extended details.

## Constructors

| | |
|---|---|
| [VectorCapabilities](-vector-capabilities/) | [jvm]<br>constructor(enabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) = false, pgVectorInstalled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, pgVectorVersion: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingEnabled: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null, embeddingProvider: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, embeddingModel: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [embeddingEnabled](embedding-enabled/) | [jvm]<br>@SerialName(value = &quot;embedding_enabled&quot;)<br>val [embeddingEnabled](embedding-enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [embeddingModel](embedding-model/) | [jvm]<br>@SerialName(value = &quot;embedding_model&quot;)<br>val [embeddingModel](embedding-model/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [embeddingProvider](embedding-provider/) | [jvm]<br>@SerialName(value = &quot;embedding_provider&quot;)<br>val [embeddingProvider](embedding-provider/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [enabled](enabled/) | [jvm]<br>val [enabled](enabled/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html) |
| [pgVectorInstalled](pg-vector-installed/) | [jvm]<br>@SerialName(value = &quot;pgvector_installed&quot;)<br>val [pgVectorInstalled](pg-vector-installed/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [pgVectorVersion](pg-vector-version/) | [jvm]<br>@SerialName(value = &quot;pgvector_version&quot;)<br>val [pgVectorVersion](pg-vector-version/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |