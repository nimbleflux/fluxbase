---
title: "Entity"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[Entity](./)

# Entity

[jvm]\
@Serializable

data class [Entity](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val entityType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val canonicalName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val aliases: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList())

## Constructors

| | |
|---|---|
| [Entity](-entity/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), knowledgeBaseId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, entityType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), name: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), canonicalName: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, aliases: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; = emptyList()) |

## Properties

| Name | Summary |
|---|---|
| [aliases](aliases/) | [jvm]<br>val [aliases](aliases/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)&gt; |
| [canonicalName](canonical-name/) | [jvm]<br>@SerialName(value = &quot;canonical_name&quot;)<br>val [canonicalName](canonical-name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [entityType](entity-type/) | [jvm]<br>@SerialName(value = &quot;entity_type&quot;)<br>val [entityType](entity-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [knowledgeBaseId](knowledge-base-id/) | [jvm]<br>@SerialName(value = &quot;knowledge_base_id&quot;)<br>val [knowledgeBaseId](knowledge-base-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [name](name/) | [jvm]<br>val [name](name/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |