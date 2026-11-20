---
title: "deleteWithHeaders"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.core](../../)/[FluxbaseHttpClient](../)/[deleteWithHeaders](./)

# deleteWithHeaders

[jvm]\
suspend fun [deleteWithHeaders](./)(path: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), body: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? = null, headers: [Map](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-map/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyMap()): [HttpResponse](../../-http-response/)

DELETE that can carry a JSON [body](./) and returns response headers — used by account deletion (`DELETE /api/v1/auth/account` accepts an optional `{password}`). Mirrors `fetch.delete(path, { body })` in `fetch.ts`.