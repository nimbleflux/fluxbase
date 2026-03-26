import { create } from "zustand";
import { persist } from "zustand/middleware";

export interface Tenant {
  id: string;
  slug: string;
  name: string;
  is_default: boolean;
  metadata: Record<string, unknown>;
  my_role?: "tenant_admin" | "tenant_member";
  created_at: string;
  updated_at?: string;
}

interface TenantState {
  currentTenant: Tenant | null;
  tenants: Tenant[];
  isInstanceAdmin: boolean;
  actingAsTenantAdmin: boolean;
  setCurrentTenant: (tenant: Tenant | null) => void;
  setTenants: (tenants: Tenant[], isInstanceAdmin?: boolean) => void;
  setIsInstanceAdmin: (isAdmin: boolean) => void;
  setActingAsTenantAdmin: (acting: boolean) => void;
  clearTenant: () => void;
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
      setTenants: (tenants, isInstanceAdmin) =>
        set((state) => {
          const isAdmin = isInstanceAdmin ?? state.isInstanceAdmin;
          // If there's only one tenant, auto-select it regardless of role
          // This simplifies UX since all menu items are visible anyway
          if (tenants.length === 1) {
            return {
              ...state,
              tenants,
              currentTenant: tenants[0],
            };
          }
          // If user is an instance admin and has no persisted tenant, keep them at instance level
          if (isAdmin && !state.currentTenant) {
            return {
              ...state,
              tenants,
              currentTenant: null,
            };
          }
          // For non-instance admins, auto-select first tenant if none selected
          // Or restore previously selected tenant if it still exists
          return {
            ...state,
            tenants,
            currentTenant: state.currentTenant
              ? tenants.find((t) => t.id === state.currentTenant?.id) ||
                tenants[0] ||
                null
              : tenants[0] || null,
          };
        }),
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
      name: "fluxbase-tenant-store",
      partialize: (state) => ({
        currentTenant: state.currentTenant,
      }),
    },
  ),
);
