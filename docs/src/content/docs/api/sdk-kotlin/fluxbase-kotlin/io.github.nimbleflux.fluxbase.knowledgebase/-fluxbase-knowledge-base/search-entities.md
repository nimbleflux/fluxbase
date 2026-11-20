---
title: "searchEntities"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[searchEntities](./)

# searchEntities

[jvm]\
suspend fun [searchEntities](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), query: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[Entity](../../-entity/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/entities/search?q=…` — search entities by name.