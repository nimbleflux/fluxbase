---
editUrl: false
next: false
prev: false
title: "ServiceKeysManager"
---

Service Keys Manager

Manages service keys (anon and service) for tenant databases.
Each tenant has their own auth.service_keys table.

## Example

```typescript
// List all service keys
const { data, error } = await client.admin.serviceKeys.list()

// Create a new service key
const { data, error } = await client.admin.serviceKeys.create({
  name: 'Production API Key',
  key_type: 'service',
  scopes: ['*']
})

// Rotate a key
const { data, error } = await client.admin.serviceKeys.rotate('key-id')
```

## Constructors

### Constructor

> **new ServiceKeysManager**(`fetch`): `ServiceKeysManager`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`ServiceKeysManager`

## Methods

### create()

> **create**(`request`): `Promise`\<\{ `data`: [`ServiceKeyWithKey`](/api/sdk/interfaces/servicekeywithkey/) \| `null`; `error`: `Error` \| `null`; \}\>

Create a new service key

The full key value is only returned once - store it securely!

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`CreateServiceKeyRequest`](/api/sdk/interfaces/createservicekeyrequest/) | Key creation options |

#### Returns

`Promise`\<\{ `data`: [`ServiceKeyWithKey`](/api/sdk/interfaces/servicekeywithkey/) \| `null`; `error`: `Error` \| `null`; \}\>

Created key with full key value

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.create({
  name: 'Production API Key',
  key_type: 'service',
  scopes: ['*'],
  rate_limit_per_minute: 1000
})

if (data) {
  // Store data.key securely - it won't be shown again!
  console.log('Key created:', data.key)
}
```

***

### delete()

> **delete**(`id`): `Promise`\<\{ `error`: `Error` \| `null`; \}\>

Delete a service key permanently

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |

#### Returns

`Promise`\<\{ `error`: `Error` \| `null`; \}\>

Success or error

#### Example

```typescript
const { error } = await client.admin.serviceKeys.delete('key-id')
```

***

### deprecate()

> **deprecate**(`id`, `request?`): `Promise`\<\{ `data`: \{ `deprecated_at`: `string`; `grace_period_ends_at`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Deprecate a service key (graceful rotation)

Marks the key for removal but keeps it active during grace period.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |
| `request?` | [`DeprecateServiceKeyRequest`](/api/sdk/interfaces/deprecateservicekeyrequest/) | Deprecation options |

#### Returns

`Promise`\<\{ `data`: \{ `deprecated_at`: `string`; `grace_period_ends_at`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Deprecation details

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.deprecate('key-id', {
  reason: 'Rotating to new key',
  grace_period_hours: 48
})
```

***

### disable()

> **disable**(`id`): `Promise`\<\{ `error`: `Error` \| `null`; \}\>

Disable a service key (temporarily)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |

#### Returns

`Promise`\<\{ `error`: `Error` \| `null`; \}\>

Success or error

#### Example

```typescript
const { error } = await client.admin.serviceKeys.disable('key-id')
```

***

### enable()

> **enable**(`id`): `Promise`\<\{ `error`: `Error` \| `null`; \}\>

Enable a disabled service key

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |

#### Returns

`Promise`\<\{ `error`: `Error` \| `null`; \}\>

Success or error

#### Example

```typescript
const { error } = await client.admin.serviceKeys.enable('key-id')
```

***

### get()

> **get**(`id`): `Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/) \| `null`; `error`: `Error` \| `null`; \}\>

Get a service key by ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |

#### Returns

`Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/) \| `null`; `error`: `Error` \| `null`; \}\>

Service key details

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.get('key-id')
```

***

### getRevocationHistory()

> **getRevocationHistory**(`id`): `Promise`\<\{ `data`: \{ `id`: `string`; `name`: `string`; `revocation_reason`: `string`; `revoked_at`: `string`; `revoked_by`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Get revocation history for a service key

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |

#### Returns

`Promise`\<\{ `data`: \{ `id`: `string`; `name`: `string`; `revocation_reason`: `string`; `revoked_at`: `string`; `revoked_by`: `string`; \} \| `null`; `error`: `Error` \| `null`; \}\>

Revocation history

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.getRevocationHistory('key-id')
```

***

### list()

> **list**(): `Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List all service keys

#### Returns

`Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/)[] \| `null`; `error`: `Error` \| `null`; \}\>

List of service keys

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.list()
```

***

### revoke()

> **revoke**(`id`, `request?`): `Promise`\<\{ `error`: `Error` \| `null`; \}\>

Revoke a service key permanently (emergency)

Use for immediate revocation when a key is compromised.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |
| `request?` | [`RevokeServiceKeyRequest`](/api/sdk/interfaces/revokeservicekeyrequest/) | Revocation options |

#### Returns

`Promise`\<\{ `error`: `Error` \| `null`; \}\>

Success or error

#### Example

```typescript
const { error } = await client.admin.serviceKeys.revoke('key-id', {
  reason: 'Key was compromised'
})
```

***

### rotate()

> **rotate**(`id`): `Promise`\<\{ `data`: [`ServiceKeyWithKey`](/api/sdk/interfaces/servicekeywithkey/) \| `null`; `error`: `Error` \| `null`; \}\>

Rotate a service key (create replacement)

Creates a new key with the same settings and deprecates the old one.
The new key is returned with its full value.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID to rotate |

#### Returns

`Promise`\<\{ `data`: [`ServiceKeyWithKey`](/api/sdk/interfaces/servicekeywithkey/) \| `null`; `error`: `Error` \| `null`; \}\>

New key with full key value

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.rotate('old-key-id')

if (data) {
  console.log('New key:', data.key)
  console.log('Old key deprecated at:', data.deprecated_at)
}
```

***

### update()

> **update**(`id`, `request`): `Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/) \| `null`; `error`: `Error` \| `null`; \}\>

Update a service key

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Service key ID |
| `request` | [`UpdateServiceKeyRequest`](/api/sdk/interfaces/updateservicekeyrequest/) | Update options |

#### Returns

`Promise`\<\{ `data`: [`ServiceKey`](/api/sdk/interfaces/servicekey/) \| `null`; `error`: `Error` \| `null`; \}\>

Updated key

#### Example

```typescript
const { data, error } = await client.admin.serviceKeys.update('key-id', {
  name: 'New Name',
  rate_limit_per_minute: 2000
})
```
