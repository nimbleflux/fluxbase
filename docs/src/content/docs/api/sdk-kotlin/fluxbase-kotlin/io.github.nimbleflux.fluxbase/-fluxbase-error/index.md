---
title: "FluxbaseError"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase](../)/[FluxbaseError](./)

# FluxbaseError

[jvm]\
data class [FluxbaseError](./)(val status: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, val code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, val details: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? = null, val message: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)) : [RuntimeException](https://docs.oracle.com/javase/8/docs/api/java/lang/RuntimeException.html)

An error from a Fluxbase API call. Port of `FluxbaseError` from `sdk/src/types.ts:235`.

## Constructors

| | |
|---|---|
| [FluxbaseError](-fluxbase-error/) | [jvm]<br>constructor(status: [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? = null, code: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null, details: [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? = null, message: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)) |

## Properties

| Name | Summary |
|---|---|
| [cause](../../iogithubnimblefluxfluxbasecore/-fluxbase-exception/#-654012527%2FProperties%2F-1216412040) | [jvm]<br>open val [cause](../../iogithubnimblefluxfluxbasecore/-fluxbase-exception/#-654012527%2FProperties%2F-1216412040): [Throwable](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-throwable/index.html)? |
| [code](code/) | [jvm]<br>val [code](code/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? |
| [details](details/) | [jvm]<br>val [details](details/): [Any](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-any/index.html)? |
| [message](message/) | [jvm]<br>open override val [message](message/): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [status](status/) | [jvm]<br>val [status](status/): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html)? |