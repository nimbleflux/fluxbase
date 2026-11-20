---
title: "getEntityRelationships"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[getEntityRelationships](./)

# getEntityRelationships

[jvm]\
suspend fun [getEntityRelationships](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), entityId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[EntityRelationship](../../-entity-relationship/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/entities/:entityId/relationships`.