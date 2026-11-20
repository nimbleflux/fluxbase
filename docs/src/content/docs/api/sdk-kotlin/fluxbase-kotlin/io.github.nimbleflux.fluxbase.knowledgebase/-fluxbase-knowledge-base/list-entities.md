---
title: "listEntities"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[listEntities](./)

# listEntities

[jvm]\
suspend fun [listEntities](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), type: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[Entity](../../-entity/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/entities` — list entities.