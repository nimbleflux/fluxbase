---
editUrl: false
next: false
prev: false
title: "EmailSettingsManager"
---

Email Settings Manager

Manages email provider configuration including SMTP, SendGrid, Mailgun, and AWS SES.
Provides direct access to the email settings API with proper handling of sensitive credentials.

## Example

```typescript
const email = client.admin.settings.email

// Get current email settings
const settings = await email.get()
console.log(settings.provider) // 'smtp'
console.log(settings.smtp_password_set) // true (password is configured)

// Update email settings
await email.update({
  provider: 'sendgrid',
  sendgrid_api_key: 'SG.xxx',
  from_address: 'noreply@yourapp.com'
})

// Test email configuration
const result = await email.test('test@example.com')
console.log(result.success) // true

// Convenience methods
await email.enable()
await email.disable()
await email.setProvider('smtp')
```

## Constructors

### Constructor

> **new EmailSettingsManager**(`fetch`): `EmailSettingsManager`

#### Parameters

| Parameter | Type |
| ------ | ------ |
| `fetch` | [`FluxbaseFetch`](/api/sdk/classes/fluxbasefetch/) |

#### Returns

`EmailSettingsManager`

## Methods

### disable()

> **disable**(): `Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Disable email functionality

Convenience method to disable the email system.

#### Returns

`Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Promise resolving to EmailProviderSettings

#### Example

```typescript
await client.admin.settings.email.disable()
```

***

### enable()

> **enable**(): `Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Enable email functionality

Convenience method to enable the email system.

#### Returns

`Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Promise resolving to EmailProviderSettings

#### Example

```typescript
await client.admin.settings.email.enable()
```

***

### get()

> **get**(): `Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Get current email provider settings

Returns the current email configuration. Sensitive values (passwords, client keys)
are not returned - instead, boolean flags indicate whether they are set.

#### Returns

`Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Promise resolving to EmailProviderSettings

#### Example

```typescript
const settings = await client.admin.settings.email.get()

console.log('Provider:', settings.provider)
console.log('From:', settings.from_address)
console.log('SMTP password configured:', settings.smtp_password_set)

// Check for environment variable overrides
if (settings._overrides.provider?.is_overridden) {
  console.log('Provider is set by env var:', settings._overrides.provider.env_var)
}
```

***

### setProvider()

> **setProvider**(`provider`): `Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Set the email provider

Convenience method to change the email provider.
Note: You'll also need to configure the provider-specific settings.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `provider` | `"smtp"` \| `"sendgrid"` \| `"mailgun"` \| `"ses"` | The email provider to use |

#### Returns

`Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Promise resolving to EmailProviderSettings

#### Example

```typescript
await client.admin.settings.email.setProvider('sendgrid')
```

***

### test()

> **test**(`recipientEmail`): `Promise`\<[`TestEmailSettingsResponse`](/api/sdk/interfaces/testemailsettingsresponse/)\>

Test email configuration by sending a test email

Sends a test email to verify that the current email configuration is working.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `recipientEmail` | `string` | Email address to send the test email to |

#### Returns

`Promise`\<[`TestEmailSettingsResponse`](/api/sdk/interfaces/testemailsettingsresponse/)\>

Promise resolving to TestEmailSettingsResponse

#### Throws

Error if email sending fails

#### Example

```typescript
try {
  const result = await client.admin.settings.email.test('admin@yourapp.com')
  console.log('Test email sent:', result.message)
} catch (error) {
  console.error('Email configuration error:', error.message)
}
```

***

### update()

> **update**(`request`): `Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Update email provider settings

Supports partial updates - only provide the fields you want to change.
Secret fields (passwords, client keys) are only updated if provided.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`UpdateEmailProviderSettingsRequest`](/api/sdk/interfaces/updateemailprovidersettingsrequest/) | Settings to update (partial update supported) |

#### Returns

`Promise`\<[`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)\>

Promise resolving to EmailProviderSettings - Updated settings

#### Throws

Error if a setting is overridden by an environment variable

#### Example

```typescript
// Configure SMTP
await client.admin.settings.email.update({
  enabled: true,
  provider: 'smtp',
  from_address: 'noreply@yourapp.com',
  from_name: 'Your App',
  smtp_host: 'smtp.gmail.com',
  smtp_port: 587,
  smtp_username: 'your-email@gmail.com',
  smtp_password: 'your-app-password',
  smtp_tls: true
})

// Configure SendGrid
await client.admin.settings.email.update({
  provider: 'sendgrid',
  sendgrid_api_key: 'SG.xxx'
})

// Update just the from address (password unchanged)
await client.admin.settings.email.update({
  from_address: 'new-address@yourapp.com'
})
```

***

### getForTenant()

> **getForTenant**(): `Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Get tenant-level email settings (resolved through cascade)

Returns email settings resolved for the current tenant context, including source information for each field.

#### Returns

`Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Promise resolving to TenantEmailProviderSettings

#### Example

```typescript
const settings = await client.admin.settings.email.getForTenant()
console.log('Provider:', settings.provider)
```

***

### updateForTenant()

> **updateForTenant**(`request`): `Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Update tenant-level email settings

Only provided fields are updated. These override instance-level defaults.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `request` | [`UpdateEmailProviderSettingsRequest`](/api/sdk/interfaces/updateemailprovidersettingsrequest/) | Settings to update (partial update supported) |

#### Returns

`Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Promise resolving to TenantEmailProviderSettings

#### Example

```typescript
await client.admin.settings.email.updateForTenant({
  provider: 'smtp',
  from_address: 'noreply@tenant.com',
})
```

***

### deleteTenantOverride()

> **deleteTenantOverride**(`field`): `Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Delete a tenant-level email setting override

Removes the tenant override for a specific field, reverting to the instance default.

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `field` | `string` | The field name to remove the override for |

#### Returns

`Promise`\<[`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)\>

Promise resolving to TenantEmailProviderSettings

#### Example

```typescript
await client.admin.settings.email.deleteTenantOverride('from_address')
```

***

### testForTenant()

> **testForTenant**(`recipientEmail`): `Promise`\<[`TestEmailSettingsResponse`](/api/sdk/interfaces/testemailsettingsresponse/)\>

Test tenant-level email configuration

#### Parameters

| Parameter | Type | Description |
| ------ | ------ | ------ |
| `recipientEmail` | `string` | Email address to send the test email to |

#### Returns

`Promise`\<[`TestEmailSettingsResponse`](/api/sdk/interfaces/testemailsettingsresponse/)\>

Promise resolving to TestEmailSettingsResponse

#### Example

```typescript
const result = await client.admin.settings.email.testForTenant('admin@tenant.com')
console.log('Test email sent:', result.message)
```
