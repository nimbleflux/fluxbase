---
title: "HttpMethod"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../)/[io.github.nimbleflux.fluxbase.core](../)/[HttpMethod](./)

# HttpMethod

[jvm]\
enum class [HttpMethod](./) : [Enum](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-enum/index.html)&lt;[HttpMethod](./)&gt; 

HTTP method enum used by the SDK's client layer. (Distinct from Ktor's `HttpMethod` to avoid naming clashes in the transport.)

## Entries

| | |
|---|---|
| [GET](-g-e-t/) | [jvm]<br>[GET](-g-e-t/) |
| [POST](-p-o-s-t/) | [jvm]<br>[POST](-p-o-s-t/) |
| [PUT](-p-u-t/) | [jvm]<br>[PUT](-p-u-t/) |
| [PATCH](-p-a-t-c-h/) | [jvm]<br>[PATCH](-p-a-t-c-h/) |
| [DELETE](-d-e-l-e-t-e/) | [jvm]<br>[DELETE](-d-e-l-e-t-e/) |
| [HEAD](-h-e-a-d/) | [jvm]<br>[HEAD](-h-e-a-d/) |

## Companion properties

| Name | Summary |
|---|---|
| [entries](entries/) | [jvm]<br>val [entries](entries/): [EnumEntries](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.enums/-enum-entries/index.html)&lt;[HttpMethod](./)&gt;<br>Returns an immutable [kotlin.enums.EnumEntries](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin.enums/-enum-entries/index.html) list containing the constants of this enum type, in the order they're declared. |

## Companion functions

| Name | Summary |
|---|---|
| [valueOf](value-of/) | [jvm]<br>fun [valueOf](value-of/)(value: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)): [HttpMethod](./)<br>Returns the enum constant of this type with the specified name. The string must match exactly an identifier used to declare an enum constant in this type. (Extraneous whitespace characters are not permitted.) |
| [values](values/) | [jvm]<br>fun [values](values/)(): [Array](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-array/index.html)&lt;[HttpMethod](./)&gt;<br>Returns an array containing the constants of this enum type, in the order they're declared. |

## Properties

| Name | Summary |
|---|---|
| [name](../../iogithubnimblefluxfluxbasevector/-vector-search-metric/-i-n-n-e-r_-p-r-o-d-u-c-t/#-372974862%2FProperties%2F-251573632) | [jvm]<br>val [name](../../iogithubnimblefluxfluxbasevector/-vector-search-metric/-i-n-n-e-r_-p-r-o-d-u-c-t/#-372974862%2FProperties%2F-251573632): [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html) |
| [ordinal](../../iogithubnimblefluxfluxbasevector/-vector-search-metric/-i-n-n-e-r_-p-r-o-d-u-c-t/#-739389684%2FProperties%2F-251573632) | [jvm]<br>val [ordinal](../../iogithubnimblefluxfluxbasevector/-vector-search-metric/-i-n-n-e-r_-p-r-o-d-u-c-t/#-739389684%2FProperties%2F-251573632): [Int](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-int/index.html) |