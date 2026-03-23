---
editUrl: false
next: false
prev: false
title: "CreateServiceKeyRequest"
---

Options for creating a service key

## Properties

| Property                                                    | Type                    | Description                                                       |
| ----------------------------------------------------------- | ----------------------- | ----------------------------------------------------------------- |
| <a id="allowed_namespaces"></a> `allowed_namespaces?`       | `string`[]              | Allowed table namespaces                                          |
| <a id="description"></a> `description?`                     | `string`                | Description                                                       |
| <a id="expires_at"></a> `expires_at?`                       | `string`                | Expiration timestamp                                              |
| <a id="key_type"></a> `key_type`                            | `"anon"` \| `"service"` | Key type: anon or service                                         |
| <a id="name"></a> `name`                                    | `string`                | Display name                                                      |
| <a id="rate_limit_per_hour"></a> `rate_limit_per_hour?`     | `number`                | Rate limit per hour                                               |
| <a id="rate_limit_per_minute"></a> `rate_limit_per_minute?` | `number`                | Rate limit per minute                                             |
| <a id="scopes"></a> `scopes?`                               | `string`[]              | Permission scopes (default: ['*'] for service, ['read'] for anon) |
