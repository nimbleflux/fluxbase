---
title: "stDWithin"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.postgrest](../../)/[QueryBuilder](../)/[stDWithin](./)

# stDWithin

[jvm]\
fun [stDWithin](./)(column: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html), geojson: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)?, distanceMeters: [Double](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-double/index.html)): [QueryBuilder](../)&lt;[T](../)&gt;

PostGIS ST_DWithin — within [distanceMeters](./) of [geojson](./).