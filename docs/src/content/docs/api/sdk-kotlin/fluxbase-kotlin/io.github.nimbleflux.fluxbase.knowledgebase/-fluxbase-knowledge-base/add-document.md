---
title: "addDocument"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[addDocument](./)

# addDocument

[jvm]\
suspend fun [addDocument](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), request: [AddDocumentRequest](../../-add-document-request/)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[KnowledgeBaseDocument](../../-knowledge-base-document/)&gt;

POST `/api/v1/ai/knowledge-bases/:id/documents` — add a text document.