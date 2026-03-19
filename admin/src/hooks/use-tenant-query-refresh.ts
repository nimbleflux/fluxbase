import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useTenantStore } from '@/stores/tenant-store'

/**
 * Hook that invalidates all queries when the tenant context changes.
 * This ensures that data is refreshed when switching between tenants
 * or when clearing tenant context (returning to instance admin mode).
 */
export function useTenantQueryRefresh() {
  const queryClient = useQueryClient()
  const currentTenant = useTenantStore((state) => state.currentTenant)

  useEffect(() => {
    // Invalidate all queries when tenant changes
    // This will cause them to refetch with the new tenant context
    queryClient.invalidateQueries()
  }, [currentTenant?.id, queryClient])
}
