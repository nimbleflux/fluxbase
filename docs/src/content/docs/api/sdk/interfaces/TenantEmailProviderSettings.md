---
editUrl: false
next: false
prev: false
title: "TenantEmailProviderSettings"
---

Tenant-level email provider settings with source information

Extends EmailProviderSettings with per-field source tracking,
showing whether each value comes from instance, tenant, config, or default.

## Extends

- [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/)

## Properties

| Property | Type | Description | Inherited from |
| ------ | ------ | ------ | ------ |
| <a id="_overrides"></a> `_overrides` | `Record`\<`string`, [`EmailSettingOverride`](/api/sdk/interfaces/emailsettingoverride/)\> | Settings overridden by environment variables | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`_overrides`](/api/sdk/interfaces/emailprovidersettings/#_overrides) |
| <a id="_sources"></a> `_sources` | `Record`\<`string`, `string`\> | Source of each field: "instance" | "tenant" | "config" | "default" | - |
| <a id="enabled"></a> `enabled` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`enabled`](/api/sdk/interfaces/emailprovidersettings/#enabled) |
| <a id="from_address"></a> `from_address` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`from_address`](/api/sdk/interfaces/emailprovidersettings/#from_address) |
| <a id="from_name"></a> `from_name` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`from_name`](/api/sdk/interfaces/emailprovidersettings/#from_name) |
| <a id="mailgun_api_key_set"></a> `mailgun_api_key_set` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`mailgun_api_key_set`](/api/sdk/interfaces/emailprovidersettings/#mailgun_api_key_set) |
| <a id="mailgun_domain"></a> `mailgun_domain` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`mailgun_domain`](/api/sdk/interfaces/emailprovidersettings/#mailgun_domain) |
| <a id="provider"></a> `provider` | `"smtp"` \| `"sendgrid"` \| `"mailgun"` \| `"ses"` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`provider`](/api/sdk/interfaces/emailprovidersettings/#provider) |
| <a id="sendgrid_api_key_set"></a> `sendgrid_api_key_set` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`sendgrid_api_key_set`](/api/sdk/interfaces/emailprovidersettings/#sendgrid_api_key_set) |
| <a id="ses_access_key_set"></a> `ses_access_key_set` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`ses_access_key_set`](/api/sdk/interfaces/emailprovidersettings/#ses_access_key_set) |
| <a id="ses_region"></a> `ses_region` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`ses_region`](/api/sdk/interfaces/emailprovidersettings/#ses_region) |
| <a id="ses_secret_key_set"></a> `ses_secret_key_set` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`ses_secret_key_set`](/api/sdk/interfaces/emailprovidersettings/#ses_secret_key_set) |
| <a id="smtp_host"></a> `smtp_host` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`smtp_host`](/api/sdk/interfaces/emailprovidersettings/#smtp_host) |
| <a id="smtp_password_set"></a> `smtp_password_set` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`smtp_password_set`](/api/sdk/interfaces/emailprovidersettings/#smtp_password_set) |
| <a id="smtp_port"></a> `smtp_port` | `number` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`smtp_port`](/api/sdk/interfaces/emailprovidersettings/#smtp_port) |
| <a id="smtp_tls"></a> `smtp_tls` | `boolean` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`smtp_tls`](/api/sdk/interfaces/emailprovidersettings/#smtp_tls) |
| <a id="smtp_username"></a> `smtp_username` | `string` | - | [`EmailProviderSettings`](/api/sdk/interfaces/emailprovidersettings/).[`smtp_username`](/api/sdk/interfaces/emailprovidersettings/#smtp_username) |
