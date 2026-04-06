---
editUrl: false
next: false
prev: false
title: "useTenant"
---

> **useTenant**(`options`): [`UseTenantReturn`](/api/sdk-react/interfaces/usetenantreturn/)

Hook for managing a single tenant

## Parameters

| Parameter | Type                                                              |
| --------- | ----------------------------------------------------------------- |
| `options` | [`UseTenantOptions`](/api/sdk-react/interfaces/usetenantoptions/) |

## Returns

[`UseTenantReturn`](/api/sdk-react/interfaces/usetenantreturn/)

## Example

```tsx
function TenantDetails({ tenantId }: { tenantId: string }) {
  const { tenant, isLoading, update } = useTenant({ tenantId });

  if (isLoading) return <div>Loading...</div>;
  if (!tenant) return <div>Tenant not found</div>;

  return (
    <div>
      <h1>{tenant.name}</h1>
      <button onClick={() => update({ name: "New Name" })}>Rename</button>
    </div>
  );
}
```
