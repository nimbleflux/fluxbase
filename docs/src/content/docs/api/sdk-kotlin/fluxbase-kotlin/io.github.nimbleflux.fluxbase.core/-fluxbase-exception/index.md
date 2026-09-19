---
title: "FluxbaseException"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.core](../)/[FluxbaseException](./)

# FluxbaseException

[jvm]\
class [FluxbaseException](./)(val status: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html), val code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val details: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? = null, message: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)) : [RuntimeException](https://docs.oracle.com/javase/8/docs/api/java/lang/RuntimeException.html)

A Fluxbase API error. Mirrors `FluxbaseError` from `sdk/src/types.ts:235`.

Thrown by the transport when a response is not ok (status >= 400). Higher layers (query builder, auth) catch this and convert it into `FluxbaseResponse.Error`.

## Constructors

| | |
|---|---|
| [FluxbaseException](-fluxbase-exception/) | [jvm]<br>constructor(status: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html), code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, details: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? = null, message: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)) |

## Properties

| Name | Summary |
|---|---|
| [cause](./#-654012527%2FProperties%2F-1216412040) | [jvm]<br>open val [cause](./#-654012527%2FProperties%2F-1216412040): [Throwable](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-throwable/index.html)? |
| [code](code/) | [jvm]<br>val [code](code/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [details](details/) | [jvm]<br>val [details](details/): [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? |
| [message](./#1824300659%2FProperties%2F-1216412040) | [jvm]<br>open val [message](./#1824300659%2FProperties%2F-1216412040): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [status](status/) | [jvm]<br>val [status](status/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |