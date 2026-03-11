---
editUrl: false
next: false
prev: false
title: "ProviderTokenResponse"
---

Response from the provider token endpoint
Contains OAuth tokens that can be used to call provider APIs (e.g., Google Drive)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="access_token"></a> `access_token` | `string` | Provider access token - use this to call provider APIs |
| <a id="expires_in"></a> `expires_in` | `number` | Seconds until token expires (0 if already expired) |
| <a id="id_token"></a> `id_token?` | `string` | OpenID Connect ID token (if available) |
| <a id="provider"></a> `provider` | `string` | OAuth provider name (e.g., 'google', 'github') |
| <a id="refresh_token"></a> `refresh_token?` | `string` | Provider refresh token (may be empty for some providers) |
| <a id="scopes"></a> `scopes?` | `string`[] | OAuth scopes granted for this token |
| <a id="token_expiry"></a> `token_expiry` | `string` | Token expiry timestamp in RFC3339 format |
| <a id="token_type"></a> `token_type` | `string` | Token type (always 'Bearer') |
