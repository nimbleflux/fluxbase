---
editUrl: false
next: false
prev: false
title: "UseTenantsReturn"
---

## Properties

| Property                                         | Type                                                                            | Description                               |
| ------------------------------------------------ | ------------------------------------------------------------------------------- | ----------------------------------------- |
| <a id="createtenant"></a> `createTenant`         | (`options`) => `Promise`\<[`Tenant`](/api/sdk-react/interfaces/tenant/)\>       | Create a new tenant (instance admin only) |
| <a id="currenttenantid"></a> `currentTenantId`   | `string` \| `undefined`                                                         | Get the current tenant ID                 |
| <a id="deletetenant"></a> `deleteTenant`         | (`id`) => `Promise`\<`void`\>                                                   | Delete a tenant (instance admin only)     |
| <a id="error"></a> `error`                       | `Error` \| `null`                                                               | Any error that occurred                   |
| <a id="isloading"></a> `isLoading`               | `boolean`                                                                       | Whether tenants are being fetched         |
| <a id="refetch"></a> `refetch`                   | () => `Promise`\<`void`\>                                                       | Refetch tenants                           |
| <a id="setcurrenttenant"></a> `setCurrentTenant` | (`tenantId`) => `void`                                                          | Set the current tenant context            |
| <a id="tenants"></a> `tenants`                   | [`TenantWithRole`](/api/sdk-react/interfaces/tenantwithrole/)[]                 | Array of tenants the user has access to   |
| <a id="updatetenant"></a> `updateTenant`         | (`id`, `options`) => `Promise`\<[`Tenant`](/api/sdk-react/interfaces/tenant/)\> | Update a tenant (tenant admin only)       |
