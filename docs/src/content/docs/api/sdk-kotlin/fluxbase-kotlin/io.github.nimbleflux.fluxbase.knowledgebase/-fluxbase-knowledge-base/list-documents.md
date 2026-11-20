---
title: "listDocuments"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[listDocuments](./)

# listDocuments

[jvm]\
suspend fun [listDocuments](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[KnowledgeBaseDocument](../../-knowledge-base-document/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/documents` — list documents.