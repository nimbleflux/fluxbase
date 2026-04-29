---
editUrl: false
next: false
prev: false
title: "SecretsManager"
---

Secrets Manager for managing edge function and job secrets

Provides both name-based (recommended) and UUID-based operations.
Name-based operations are more convenient for most use cases.

## Example

```typescript
const client = createClient({ url: 'http://localhost:8080' })
await client.auth.login({ email: 'user@example.com', password: 'password' })

// Create a global secret
const secret = await client.secrets.create({
  name: 'STRIPE_KEY',
  value: 'sk_live_xxx',
  description: 'Stripe production API key'
})

// Create a namespace-scoped secret
await client.secrets.create({
  name: 'DATABASE_URL',
  value: 'postgres://...',
  scope: 'namespace',
  namespace: 'production'
})

// Get secret by name
const secret = await client.secrets.get('STRIPE_KEY')

// Get namespace-scoped secret
const secret = await client.secrets.get('DATABASE_URL', { namespace: 'production' })

// Update secret
await client.secrets.update('STRIPE_KEY', { value: 'sk_live_new_key' })

// List all secrets
const secrets = await client.secrets.list()

// Get version history
const versions = await client.secrets.getVersions('STRIPE_KEY')

// Rollback to previous version
await client.secrets.rollback('STRIPE_KEY', 1)

// Delete secret
await client.secrets.delete('STRIPE_KEY')
```

## Constructors

### Constructor

> **new SecretsManager**(`fetch`): `SecretsManager`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`SecretsManager`

## Methods

### create()

> **create**(`request`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Create a new secret

Creates a new secret with the specified name, value, and scope.
The value is encrypted at rest and never returned by the API.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`CreateSecretRequest`](/api/sdk/interfaces/createsecretrequest/) | Secret creation request |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the created secret (metadata only)

#### Example

```typescript
// Create a global secret
const secret = await client.secrets.create({
  name: 'SENDGRID_API_KEY',
  value: 'SG.xxx',
  description: 'SendGrid API key for transactional emails'
})

// Create a namespace-scoped secret
const secret = await client.secrets.create({
  name: 'DATABASE_URL',
  value: 'postgres://user:pass@host:5432/db',
  scope: 'namespace',
  namespace: 'production',
  description: 'Production database URL'
})

// Create a secret with expiration
const secret = await client.secrets.create({
  name: 'TEMP_TOKEN',
  value: 'xyz123',
  expires_at: '2025-12-31T23:59:59Z'
})
```

***

### delete()

> **delete**(`name`, `options?`): `Promise`\<`void`\>

Delete a secret by name

Permanently deletes the secret and all its versions.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Secret name |
| `options?` | [`SecretByNameOptions`](/api/sdk/interfaces/secretbynameoptions/) | Optional namespace for namespace-scoped secrets |

#### Returns

`Promise`\<`void`\>

Promise resolving when deletion is complete

#### Example

```typescript
// Delete a global secret
await client.secrets.delete('OLD_API_KEY')

// Delete a namespace-scoped secret
await client.secrets.delete('DATABASE_URL', { namespace: 'staging' })
```

***

### deleteById()

> **deleteById**(`id`): `Promise`\<`void`\>

Delete a secret by ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Secret UUID |

#### Returns

`Promise`\<`void`\>

Promise resolving when deletion is complete

#### Example

```typescript
await client.secrets.deleteById('550e8400-e29b-41d4-a716-446655440000')
```

***

### get()

> **get**(`name`, `options?`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Get a secret by name (metadata only, never includes value)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Secret name |
| `options?` | [`SecretByNameOptions`](/api/sdk/interfaces/secretbynameoptions/) | Optional namespace for namespace-scoped secrets |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the secret

#### Example

```typescript
// Get a global secret
const secret = await client.secrets.get('API_KEY')

// Get a namespace-scoped secret
const secret = await client.secrets.get('DATABASE_URL', { namespace: 'production' })
```

***

### getById()

> **getById**(`id`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Get a secret by ID (metadata only)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Secret UUID |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the secret

#### Example

```typescript
const secret = await client.secrets.getById('550e8400-e29b-41d4-a716-446655440000')
```

***

### getVersions()

> **getVersions**(`name`, `options?`): `Promise`\<[`SecretVersion`](/api/sdk/interfaces/secretversion/)[]\>

Get version history for a secret by name

Returns all historical versions of the secret (values are never included).

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Secret name |
| `options?` | [`SecretByNameOptions`](/api/sdk/interfaces/secretbynameoptions/) | Optional namespace for namespace-scoped secrets |

#### Returns

`Promise`\<[`SecretVersion`](/api/sdk/interfaces/secretversion/)[]\>

Promise resolving to array of secret versions

#### Example

```typescript
const versions = await client.secrets.getVersions('API_KEY')

versions.forEach(v => {
  console.log(`Version ${v.version} created at ${v.created_at}`)
})
```

***

### getVersionsById()

> **getVersionsById**(`id`): `Promise`\<[`SecretVersion`](/api/sdk/interfaces/secretversion/)[]\>

Get version history for a secret by ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Secret UUID |

#### Returns

`Promise`\<[`SecretVersion`](/api/sdk/interfaces/secretversion/)[]\>

Promise resolving to array of secret versions

#### Example

```typescript
const versions = await client.secrets.getVersionsById('550e8400-e29b-41d4-a716-446655440000')
```

***

### list()

> **list**(`options?`): `Promise`\<[`SecretSummary`](/api/sdk/interfaces/secretsummary/)[]\>

List all secrets (metadata only, never includes values)

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `options?` | [`ListSecretsOptions`](/api/sdk/interfaces/listsecretsoptions/) | Filter options for scope and namespace |

#### Returns

`Promise`\<[`SecretSummary`](/api/sdk/interfaces/secretsummary/)[]\>

Promise resolving to array of secret summaries

#### Example

```typescript
// List all secrets
const secrets = await client.secrets.list()

// List only global secrets
const secrets = await client.secrets.list({ scope: 'global' })

// List secrets for a specific namespace
const secrets = await client.secrets.list({ namespace: 'production' })

secrets.forEach(s => {
  console.log(`${s.name}: version ${s.version}, expired: ${s.is_expired}`)
})
```

***

### rollback()

> **rollback**(`name`, `version`, `options?`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Rollback a secret to a previous version by name

Restores the secret to a previous version's value.
This creates a new version with the old value.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Secret name |
| `version` | `number` | Version number to rollback to |
| `options?` | [`SecretByNameOptions`](/api/sdk/interfaces/secretbynameoptions/) | Optional namespace for namespace-scoped secrets |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the updated secret

#### Example

```typescript
// Rollback to version 2
const secret = await client.secrets.rollback('API_KEY', 2)
console.log(`Secret now at version ${secret.version}`)
```

***

### rollbackById()

> **rollbackById**(`id`, `version`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Rollback a secret to a previous version by ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Secret UUID |
| `version` | `number` | Version number to rollback to |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the updated secret

#### Example

```typescript
const secret = await client.secrets.rollbackById('550e8400-e29b-41d4-a716-446655440000', 2)
```

***

### stats()

> **stats**(): `Promise`\<[`SecretStats`](/api/sdk/interfaces/secretstats/)\>

Get statistics about secrets

#### Returns

`Promise`\<[`SecretStats`](/api/sdk/interfaces/secretstats/)\>

Promise resolving to secret statistics

#### Example

```typescript
const stats = await client.secrets.stats()
console.log(`Total: ${stats.total}, Expiring soon: ${stats.expiring_soon}, Expired: ${stats.expired}`)
```

***

### update()

> **update**(`name`, `request`, `options?`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Update a secret by name

Updates the secret's value, description, or expiration.
Only provided fields will be updated.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `name` | `string` | Secret name |
| `request` | [`UpdateSecretRequest`](/api/sdk/interfaces/updatesecretrequest/) | Update request |
| `options?` | [`SecretByNameOptions`](/api/sdk/interfaces/secretbynameoptions/) | Optional namespace for namespace-scoped secrets |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the updated secret

#### Example

```typescript
// Update secret value
const secret = await client.secrets.update('API_KEY', { value: 'new-value' })

// Update description
const secret = await client.secrets.update('API_KEY', { description: 'Updated description' })

// Update namespace-scoped secret
const secret = await client.secrets.update('DATABASE_URL',
  { value: 'postgres://new-host:5432/db' },
  { namespace: 'production' }
)
```

***

### updateById()

> **updateById**(`id`, `request`): `Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Update a secret by ID

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `id` | `string` | Secret UUID |
| `request` | [`UpdateSecretRequest`](/api/sdk/interfaces/updatesecretrequest/) | Update request |

#### Returns

`Promise`\<[`Secret`](/api/sdk/interfaces/secret/)\>

Promise resolving to the updated secret

#### Example

```typescript
const secret = await client.secrets.updateById('550e8400-e29b-41d4-a716-446655440000', {
  value: 'new-value'
})
```
