---
title: "timeout"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.rpc](../../)/[RpcInvokeOptions](../)/[timeout](./)

# timeout

[jvm]\
val [timeout](./): [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html)?

Per-request timeout in milliseconds. When it elapses, invoke fails with a 408 &quot;Request timeout&quot; error. The client-wide `FluxbaseClientOptions.timeout` still applies at the transport level, so a value larger than that is capped by it.