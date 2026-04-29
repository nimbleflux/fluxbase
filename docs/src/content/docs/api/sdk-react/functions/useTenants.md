---
editUrl: false
next: false
prev: false
title: "useTenants"
---

> **useTenants**(`options?`): [`UseTenantsReturn`](/api/sdk-react/interfaces/usetenantsreturn/)

Hook for managing tenants

Provides tenant list and management functions for multi-tenant applications.

## Parameters

| Parameter | Type |
| ------ | ------ |
| `options` | [`UseTenantsOptions`](/api/sdk-react/interfaces/usetenantsoptions/) |

## Returns

[`UseTenantsReturn`](/api/sdk-react/interfaces/usetenantsreturn/)

## Example

```tsx
function TenantManager() {
  const { tenants, isLoading, setCurrentTenant, currentTenantId } = useTenants()

  if (isLoading) return <div>Loading...</div>

  return (
    <select
      value={currentTenantId || ''}
      onChange={(e) => setCurrentTenant(e.target.value || undefined)}
    >
      {tenants.map(t => (
        <option key={t.id} value={t.id}>
          {t.name} ({t.my_role})
        </option>
      ))}
    </select>
  )
}
```
