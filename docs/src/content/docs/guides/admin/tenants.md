---
title: "Tenant Management"
description: Manage tenants, members, and tenant-specific settings via the admin UI.
---

The Tenants section manages multi-tenant deployments. Each tenant represents an organization with isolated data and configurable settings.

## Tenant List

View all tenants with member counts and creation dates. Create new tenants from this page.

## Tenant Detail

Click a tenant to access management tabs:

### Members

Manage tenant membership:

- **Add members** - Invite users by email or select from existing users
- **Assign roles** - Set member roles (admin, member, etc.)
- **Remove members** - Revoke tenant access

### OAuth Providers

Configure OAuth providers specific to this tenant:

- Enable/disable providers (Google, GitHub, etc.)
- Set client ID and secret per provider
- Override instance-level OAuth settings

See: [OAuth Providers Guide](../../oauth-providers)

### SAML

Configure SAML SSO for enterprise authentication:

- Add identity provider metadata
- Configure attribute mappings
- Set up sign-on URLs

See: [SAML SSO Guide](../../saml-sso)

### Settings

View and edit tenant-specific configuration:

- See which settings can be overridden
- Edit overridable settings (storage, email, AI, etc.)
- Reset settings to instance defaults

Settings not marked as overridable are locked to instance values.

## Related

- [Multi-Tenancy Guide](../../multi-tenancy) - API usage and configuration
- [Instance Settings](./) - Configure what tenants can override
