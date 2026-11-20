---
title: "fluxbaseResponse"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase](../)/[fluxbaseResponse](./)

# fluxbaseResponse

[jvm]\
suspend fun &lt;[T](./)&gt; [fluxbaseResponse](./)(block: suspend () -&gt; [T](./)): [FluxbaseResponse](../-fluxbase-response/)&lt;[T](./)&gt;

Wraps a suspending block, catching exceptions and converting them to [FluxbaseResponse.Error](../-fluxbase-response/-error/). The TS SDK uses `wrapAsync` for the same purpose (`sdk/src/utils/error-handling.ts`).

`CancellationException` is rethrown, not wrapped: it is the caller cancelling (or a timeout already converted by its own layer — the transport maps Ktor's request timeout to `FluxbaseException`, per-request timeouts to [FluxbaseError](../-fluxbase-error/)), and swallowing it would leave the coroutine running after its job was cancelled.