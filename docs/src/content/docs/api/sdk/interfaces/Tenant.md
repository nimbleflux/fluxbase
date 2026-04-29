---
editUrl: false
next: false
prev: false
title: "Tenant"
---

Tenant in the system

## Extended by

- [`TenantWithRole`](/api/sdk/interfaces/tenantwithrole/)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="created_at"></a> `created_at` | `string` | Creation timestamp |
| <a id="db_name"></a> `db_name?` | `string` \| `null` | Database name (null = uses main database, for backward compatibility) |
| <a id="deleted_at"></a> `deleted_at?` | `string` \| `null` | Soft delete timestamp |
| <a id="id"></a> `id` | `string` | Unique identifier for the tenant |
| <a id="is_default"></a> `is_default` | `boolean` | Whether this is the default tenant |
| <a id="metadata"></a> `metadata?` | `Record`\<`string`, `unknown`\> | Arbitrary metadata |
| <a id="name"></a> `name` | `string` | Display name |
| <a id="slug"></a> `slug` | `string` | URL-friendly identifier (e.g., "acme-corp") |
| <a id="status"></a> `status` | [`TenantStatus`](/api/sdk/type-aliases/tenantstatus/) | Current status of the tenant |
| <a id="updated_at"></a> `updated_at?` | `string` | Last update timestamp |
