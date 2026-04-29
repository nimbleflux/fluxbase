---
editUrl: false
next: false
prev: false
title: "EmailProviderSettings"
---

Email provider settings response from /api/v1/admin/email/settings

This is the flat structure returned by the admin API, which differs from
the nested EmailSettings structure used in AppSettings.

## Extended by

- [`TenantEmailProviderSettings`](/api/sdk/interfaces/tenantemailprovidersettings/)

## Properties

| Property | Type | Description |
| ------ | ------ | ------ |
| <a id="_overrides"></a> `_overrides` | `Record`\<`string`, [`EmailSettingOverride`](/api/sdk/interfaces/emailsettingoverride/)\> | Settings overridden by environment variables |
| <a id="enabled"></a> `enabled` | `boolean` | - |
| <a id="from_address"></a> `from_address` | `string` | - |
| <a id="from_name"></a> `from_name` | `string` | - |
| <a id="mailgun_api_key_set"></a> `mailgun_api_key_set` | `boolean` | - |
| <a id="mailgun_domain"></a> `mailgun_domain` | `string` | - |
| <a id="provider"></a> `provider` | `"smtp"` \| `"sendgrid"` \| `"mailgun"` \| `"ses"` | - |
| <a id="sendgrid_api_key_set"></a> `sendgrid_api_key_set` | `boolean` | - |
| <a id="ses_access_key_set"></a> `ses_access_key_set` | `boolean` | - |
| <a id="ses_region"></a> `ses_region` | `string` | - |
| <a id="ses_secret_key_set"></a> `ses_secret_key_set` | `boolean` | - |
| <a id="smtp_host"></a> `smtp_host` | `string` | - |
| <a id="smtp_password_set"></a> `smtp_password_set` | `boolean` | - |
| <a id="smtp_port"></a> `smtp_port` | `number` | - |
| <a id="smtp_tls"></a> `smtp_tls` | `boolean` | - |
| <a id="smtp_username"></a> `smtp_username` | `string` | - |
