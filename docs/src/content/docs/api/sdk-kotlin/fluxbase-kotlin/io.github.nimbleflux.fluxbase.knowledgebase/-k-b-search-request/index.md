---
title: "KBSearchRequest"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KBSearchRequest](./)

# KBSearchRequest

[jvm]\
@Serializable

data class [KBSearchRequest](./)(val query: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val limit: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, val mode: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val semanticWeight: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? = null, val threshold: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? = null, val rerank: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null)

## Constructors

| | |
|---|---|
| [KBSearchRequest](-k-b-search-request/) | [jvm]<br>constructor(query: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), limit: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, mode: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, semanticWeight: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? = null, threshold: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? = null, rerank: [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [limit](limit/) | [jvm]<br>val [limit](limit/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? |
| [mode](mode/) | [jvm]<br>val [mode](mode/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [query](query/) | [jvm]<br>val [query](query/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [rerank](rerank/) | [jvm]<br>val [rerank](rerank/): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)? |
| [semanticWeight](semantic-weight/) | [jvm]<br>@SerialName(value = &quot;semantic_weight&quot;)<br>val [semanticWeight](semantic-weight/): [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? |
| [threshold](threshold/) | [jvm]<br>val [threshold](threshold/): [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)? |