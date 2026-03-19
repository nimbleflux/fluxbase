import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export interface Tenant {
  id: string
  slug: string
  name: string
  is_default: boolean
  metadata: Record<string, unknown>
  my_role?: 'tenant_admin' | 'tenant_member'
  created_at: string
  updated_at?: string
}

interface TenantState {
  currentTenant: Tenant | null
  tenants: Tenant[]
  isInstanceAdmin: boolean
  actingAsTenantAdmin: boolean
  setCurrentTenant: (tenant: Tenant | null) => void
  setTenants: (tenants: Tenant[]) => void
  setIsInstanceAdmin: (isAdmin: boolean) => void
  setActingAsTenantAdmin: (acting: boolean) => void
  clearTenant: () => void
}

export const useTenantStore = create<TenantState>()(
  persist(
    (set) => ({
      currentTenant: null,
      tenants: [],
      isInstanceAdmin: false,
      actingAsTenantAdmin: false,
      setCurrentTenant: (tenant) =>
        set((state) => ({
          ...state,
          currentTenant: tenant,
        })),
      setTenants: (tenants) =>
        set((state) => ({
          ...state,
          tenants,
          currentTenant: state.currentTenant
            ? tenants.find((t) => t.id === state.currentTenant?.id) || tenants[0] || null
            : tenants[0] || null,
        })),
      setIsInstanceAdmin: (isAdmin) =>
        set((state) => ({
          ...state,
          isInstanceAdmin: isAdmin,
        })),
      setActingAsTenantAdmin: (acting) =>
        set((state) => ({
          ...state,
          actingAsTenantAdmin: acting,
        })),
      clearTenant: () =>
        set((state) => ({
          ...state,
          currentTenant: null,
          actingAsTenantAdmin: false,
        })),
    }),
    {
      name: 'fluxbase-tenant-store',
      partialize: (state) => ({
        currentTenant: state.currentTenant,
      }),
    }
  )
)
