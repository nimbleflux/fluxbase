---
title: "uploadDocument"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../../)/[FluxbaseKnowledgeBase](../)/[uploadDocument](./)

# uploadDocument

[jvm]\
suspend fun [uploadDocument](./)(kbId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), fileName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), content: [ByteArray](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-byte-array/index.html), mimeType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), metadata: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt;? = null): [FluxbaseResponse](../../../iogithubnimblefluxfluxbase/-fluxbase-response/)&lt;[KnowledgeBaseDocument](../../-knowledge-base-document/)&gt;

POST `/api/v1/ai/knowledge-bases/:id/documents/upload` — upload a file as a multipart form document. The multipart body is assembled manually (the HTTP transport has no multipart helper); error statuses throw inside the transport and are converted by [fluxbaseResponse](../../../iogithubnimblefluxfluxbase/fluxbase-response/).