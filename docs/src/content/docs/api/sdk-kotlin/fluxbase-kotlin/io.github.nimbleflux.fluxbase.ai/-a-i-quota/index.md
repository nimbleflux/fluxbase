---
title: "AIQuota"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIQuota](./)

# AIQuota

[jvm]\
@Serializable

data class [AIQuota](./)(val used: [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) = 0, val limit: [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) = 0)

One counter half of [AIUsageSnapshot](../-a-i-usage-snapshot/).

## Constructors

| | |
|---|---|
| [AIQuota](-a-i-quota/) | [jvm]<br>constructor(used: [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) = 0, limit: [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) = 0) |

## Properties

| Name | Summary |
|---|---|
| [limit](limit/) | [jvm]<br>val [limit](limit/): [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) |
| [used](used/) | [jvm]<br>val [used](used/): [Long](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-long/index.html) |