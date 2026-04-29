---
editUrl: false
next: false
prev: false
title: "UseTenantReturn"
---

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="error"></a> `error` | `Error` \| `null` | Any error that occurred |
| <a id="isloading"></a> `isLoading` | `boolean` | Whether tenant is being fetched |
| <a id="refetch"></a> `refetch` | () => `Promise`\<`void`\> | Refetch tenant |
| <a id="remove"></a> `remove` | () => `Promise`\<`void`\> | Delete the tenant |
| <a id="tenant"></a> `tenant` | [`Tenant`](/api/sdk-react/interfaces/tenant/) \| `null` | Tenant data |
| <a id="update"></a> `update` | (`options`) => `Promise`\<[`Tenant`](/api/sdk-react/interfaces/tenant/)\> | Update the tenant |
