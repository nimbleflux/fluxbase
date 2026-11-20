---
title: "AIUsageSnapshot"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.ai](../)/[AIUsageSnapshot](./)

# AIUsageSnapshot

[jvm]\
@Serializable

data class [AIUsageSnapshot](./)(val requests: [AIQuota](../-a-i-quota/)? = null, val tokens: [AIQuota](../-a-i-quota/)? = null, val resetsAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

Per-user daily usage snapshot for a chatbot (GET /api/v1/ai/usage/:chatbotId).

## Constructors

| | |
|---|---|
| [AIUsageSnapshot](-a-i-usage-snapshot/) | [jvm]<br>constructor(requests: [AIQuota](../-a-i-quota/)? = null, tokens: [AIQuota](../-a-i-quota/)? = null, resetsAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [requests](requests/) | [jvm]<br>val [requests](requests/): [AIQuota](../-a-i-quota/)? |
| [resetsAt](resets-at/) | [jvm]<br>@SerialName(value = &quot;resets_at&quot;)<br>val [resetsAt](resets-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [tokens](tokens/) | [jvm]<br>val [tokens](tokens/): [AIQuota](../-a-i-quota/)? |