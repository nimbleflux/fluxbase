---
editUrl: false
next: false
prev: false
title: "TenantWithRole"
---

Tenant with user's role (for "my tenants" endpoint)

## Extends

- [`Tenant`](/api/sdk-react/interfaces/tenant/)

## Properties

| Property | Type | Description | Inherited from |
| ------ | ------ | ------ | ------ |
| <a id="created_at"></a> `created_at` | `string` | Creation timestamp | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`created_at`](/api/sdk-react/interfaces/tenant/#created_at) |
| <a id="db_name"></a> `db_name?` | `string` \| `null` | Database name (null = uses main database, for backward compatibility) | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`db_name`](/api/sdk-react/interfaces/tenant/#db_name) |
| <a id="deleted_at"></a> `deleted_at?` | `string` \| `null` | Soft delete timestamp | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`deleted_at`](/api/sdk-react/interfaces/tenant/#deleted_at) |
| <a id="id"></a> `id` | `string` | Unique identifier for the tenant | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`id`](/api/sdk-react/interfaces/tenant/#id) |
| <a id="is_default"></a> `is_default` | `boolean` | Whether this is the default tenant | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`is_default`](/api/sdk-react/interfaces/tenant/#is_default) |
| <a id="metadata"></a> `metadata?` | `Record`\<`string`, `unknown`\> | Arbitrary metadata | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`metadata`](/api/sdk-react/interfaces/tenant/#metadata) |
| <a id="my_role"></a> `my_role` | `"tenant_admin"` | Current user's role in this tenant (always tenant_admin with database-per-tenant) | - |
| <a id="name"></a> `name` | `string` | Display name | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`name`](/api/sdk-react/interfaces/tenant/#name) |
| <a id="slug"></a> `slug` | `string` | URL-friendly identifier (e.g., "acme-corp") | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`slug`](/api/sdk-react/interfaces/tenant/#slug) |
| <a id="status"></a> `status` | `TenantStatus` | Current status of the tenant | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`status`](/api/sdk-react/interfaces/tenant/#status) |
| <a id="updated_at"></a> `updated_at?` | `string` | Last update timestamp | [`Tenant`](/api/sdk-react/interfaces/tenant/).[`updated_at`](/api/sdk-react/interfaces/tenant/#updated_at) |
