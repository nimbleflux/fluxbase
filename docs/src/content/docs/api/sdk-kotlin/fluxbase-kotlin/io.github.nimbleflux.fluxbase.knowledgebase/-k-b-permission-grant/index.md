---
title: "KBPermissionGrant"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KBPermissionGrant](./)

# KBPermissionGrant

[jvm]\
@Serializable

data class [KBPermissionGrant](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val documentId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val userId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val permission: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val grantedBy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val grantedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [KBPermissionGrant](-k-b-permission-grant/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, documentId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, userId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), permission: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), grantedBy: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, grantedAt: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [documentId](document-id/) | [jvm]<br>@SerialName(value = &quot;document_id&quot;)<br>val [documentId](document-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [grantedAt](granted-at/) | [jvm]<br>@SerialName(value = &quot;granted_at&quot;)<br>val [grantedAt](granted-at/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [grantedBy](granted-by/) | [jvm]<br>@SerialName(value = &quot;granted_by&quot;)<br>val [grantedBy](granted-by/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [knowledgeBaseId](knowledge-base-id/) | [jvm]<br>@SerialName(value = &quot;knowledge_base_id&quot;)<br>val [knowledgeBaseId](knowledge-base-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [permission](permission/) | [jvm]<br>val [permission](permission/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [userId](user-id/) | [jvm]<br>@SerialName(value = &quot;user_id&quot;)<br>val [userId](user-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |