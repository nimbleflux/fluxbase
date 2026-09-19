---
title: "trustSslCertificates"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase](../../)/[FluxbaseClientOptions](../)/[trustSslCertificates](./)

# trustSslCertificates

[jvm]\
val [trustSslCertificates](./): [Boolean](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-boolean/index.html)

Skip TLS certificate validation for this client's engine (HTTP and realtime WebSocket). For self-hosted instances behind self-signed certificates — opt in only, validation stays the default. Scoped to this client's connections, unlike a global trust-store override.