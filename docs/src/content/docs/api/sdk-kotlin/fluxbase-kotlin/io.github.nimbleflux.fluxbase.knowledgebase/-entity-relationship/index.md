---
title: "EntityRelationship"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[EntityRelationship](./)

# EntityRelationship

[jvm]\
@Serializable

data class [EntityRelationship](./)(val id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val sourceEntityId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val targetEntityId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val relationshipType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), val direction: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

## Constructors

| | |
|---|---|
| [EntityRelationship](-entity-relationship/) | [jvm]<br>constructor(id: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), sourceEntityId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), targetEntityId: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), relationshipType: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), direction: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null) |

## Properties

| Name | Summary |
|---|---|
| [direction](direction/) | [jvm]<br>val [direction](direction/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [id](id/) | [jvm]<br>val [id](id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [relationshipType](relationship-type/) | [jvm]<br>@SerialName(value = &quot;relationship_type&quot;)<br>val [relationshipType](relationship-type/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [sourceEntityId](source-entity-id/) | [jvm]<br>@SerialName(value = &quot;source_entity_id&quot;)<br>val [sourceEntityId](source-entity-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [targetEntityId](target-entity-id/) | [jvm]<br>@SerialName(value = &quot;target_entity_id&quot;)<br>val [targetEntityId](target-entity-id/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |