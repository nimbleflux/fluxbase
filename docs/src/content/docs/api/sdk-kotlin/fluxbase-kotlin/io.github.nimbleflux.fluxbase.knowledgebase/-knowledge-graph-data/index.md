---
title: "KnowledgeGraphData"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.knowledgebase](../)/[KnowledgeGraphData](./)

# KnowledgeGraphData

[jvm]\
@Serializable

data class [KnowledgeGraphData](./)(val entities: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[Entity](../-entity/)&gt; = emptyList(), val relationships: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[EntityRelationship](../-entity-relationship/)&gt; = emptyList())

## Constructors

| | |
|---|---|
| [KnowledgeGraphData](-knowledge-graph-data/) | [jvm]<br>constructor(entities: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[Entity](../-entity/)&gt; = emptyList(), relationships: [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[EntityRelationship](../-entity-relationship/)&gt; = emptyList()) |

## Properties

| Name | Summary |
|---|---|
| [entities](entities/) | [jvm]<br>val [entities](entities/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[Entity](../-entity/)&gt; |
| [relationships](relationships/) | [jvm]<br>val [relationships](relationships/): [List](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.collections/-list/index.html)&lt;[EntityRelationship](../-entity-relationship/)&gt; |