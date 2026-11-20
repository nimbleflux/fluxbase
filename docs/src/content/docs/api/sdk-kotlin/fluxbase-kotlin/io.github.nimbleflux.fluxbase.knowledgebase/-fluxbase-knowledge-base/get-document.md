---
title: "getDocument"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[getDocument](./)

# getDocument

[jvm]\
suspend fun [getDocument](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), documentId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[KnowledgeBaseDocument](../../-knowledge-base-document/)&gt;

GET `/api/v1/ai/knowledge-bases/:id/documents/:docId` — get a document.