---
title: "listPermissions"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[listPermissions](./)

# listPermissions

[jvm]\
suspend fun [listPermissions](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[KBPermissionGrant](../../-k-b-permission-grant/)&gt;&gt;

GET `/api/v1/ai/knowledge-bases/:id/permissions` — list grants (owner).