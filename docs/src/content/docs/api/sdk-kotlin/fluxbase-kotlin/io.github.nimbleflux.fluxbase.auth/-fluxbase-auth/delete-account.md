---
title: "deleteAccount"
editUrl: false
next: false
prev: false
---

//[fluxbase-kotlin](../../../../)/[io.github.nimbleflux.fluxbase.auth](../../)/[FluxbaseAuth](../)/[deleteAccount](./)

# deleteAccount

[jvm]\
suspend fun [deleteAccount](./)(password: [String](https://kotlinlang.org/api/core/kotlin-stdlib/kotlin/-string/index.html)? = null)

Delete the current user's account (self-service account deletion). DELETEs to `/api/v1/auth/account` with an optional `{password}` body — required when the account has a password credential; OAuth-only accounts pass null. On success the local session is cleared so the SDK stops sending the revoked access token. Failures (403 wrong password, 429 rate limit) throw and leave the session untouched.

Port of `deleteAccount()` in `auth.ts`.