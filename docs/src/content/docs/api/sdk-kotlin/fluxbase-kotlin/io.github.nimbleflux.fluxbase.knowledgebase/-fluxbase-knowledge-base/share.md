---
title: "share"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[share](./)

# share

[jvm]\
suspend fun [share](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), request: [GrantKBPermissionRequest](../../-grant-k-b-permission-request/)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[KBPermissionGrant](../../-k-b-permission-grant/)&gt;

POST `/api/v1/ai/knowledge-bases/:id/share` — grant a user permission (owner).