---
title: "search"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[search](./)

# search

[jvm]\
suspend fun [search](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), request: [KBSearchRequest](../../-k-b-search-request/)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[KBSearchResult](../../-k-b-search-result/)&gt;&gt;

POST `/api/v1/ai/knowledge-bases/:id/search` — hybrid search.