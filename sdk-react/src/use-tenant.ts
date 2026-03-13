/**
 * React hooks for multi-tenant management
 *
 * @module use-tenant
 */

import { useState, useEffect, useCallback } from 'react'
import { useFluxbaseClient } from './context'
import type {
  Tenant,
  TenantMembership,
  TenantWithRole,
  CreateTenantOptions,
  UpdateTenantOptions,
  AddTenantMemberOptions,
  UpdateTenantMemberOptions,
} from '@nimbleflux/fluxbase-sdk'

export interface UseTenantsOptions {
  /**
   * Whether to automatically fetch tenants on mount
   * @default true
   */
  autoFetch?: boolean
}

export interface UseTenantsReturn {
  /**
   * Array of tenants the user has access to
   */
  tenants: TenantWithRole[]

  /**
   * Whether tenants are being fetched
   */
  isLoading: boolean

  /**
   * Any error that occurred
   */
  error: Error | null

  /**
   * Refetch tenants
   */
  refetch: () => Promise<void>

  /**
   * Create a new tenant (instance admin only)
   */
  createTenant: (options: CreateTenantOptions) => Promise<Tenant>

  /**
   * Update a tenant (tenant admin only)
   */
  updateTenant: (id: string, options: UpdateTenantOptions) => Promise<Tenant>

  /**
   * Delete a tenant (instance admin only)
   */
  deleteTenant: (id: string) => Promise<void>

  /**
   * Set the current tenant context
   */
  setCurrentTenant: (tenantId: string | undefined) => void

  /**
   * Get the current tenant ID
   */
  currentTenantId: string | undefined
}

/**
 * Hook for managing tenants
 *
 * Provides tenant list and management functions for multi-tenant applications.
 *
 * @example
 * ```tsx
 * function TenantManager() {
 *   const { tenants, isLoading, setCurrentTenant, currentTenantId } = useTenants()
 *
 *   if (isLoading) return <div>Loading...</div>
 *
 *   return (
 *     <select
 *       value={currentTenantId || ''}
 *       onChange={(e) => setCurrentTenant(e.target.value || undefined)}
 *     >
 *       {tenants.map(t => (
 *         <option key={t.id} value={t.id}>
 *           {t.name} ({t.my_role})
 *         </option>
 *       ))}
 *     </select>
 *   )
 * }
 * ```
 */
export function useTenants(options: UseTenantsOptions = {}): UseTenantsReturn {
  const { autoFetch = true } = options
  const client = useFluxbaseClient()

  const [tenants, setTenants] = useState<TenantWithRole[]>([])
  const [isLoading, setIsLoading] = useState(autoFetch)
  const [error, setError] = useState<Error | null>(null)
  const [currentTenantId, setCurrentTenantId] = useState<string | undefined>(
    client.getTenantId()
  )

  const fetchTenants = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const { data, error: fetchError } = await client.tenant.listMine()
      if (fetchError) {
        setError(fetchError)
      } else {
        setTenants(data || [])
      }
    } catch (err) {
      setError(err as Error)
    } finally {
      setIsLoading(false)
    }
  }, [client])

  const createTenant = useCallback(
    async (opts: CreateTenantOptions): Promise<Tenant> => {
      const { data, error: createError } = await client.tenant.create(opts)
      if (createError) throw createError
      await fetchTenants()
      return data!
    },
    [client, fetchTenants]
  )

  const updateTenant = useCallback(
    async (id: string, opts: UpdateTenantOptions): Promise<Tenant> => {
      const { data, error: updateError } = await client.tenant.update(id, opts)
      if (updateError) throw updateError
      await fetchTenants()
      return data!
    },
    [client, fetchTenants]
  )

  const deleteTenant = useCallback(
    async (id: string): Promise<void> => {
      const { error: deleteError } = await client.tenant.delete(id)
      if (deleteError) throw deleteError
      await fetchTenants()
    },
    [client, fetchTenants]
  )

  const setCurrentTenant = useCallback(
    (tenantId: string | undefined) => {
      client.setTenant(tenantId)
      setCurrentTenantId(tenantId)
    },
    [client]
  )

  useEffect(() => {
    if (autoFetch) {
      fetchTenants()
    }
  }, [autoFetch, fetchTenants])

  return {
    tenants,
    isLoading,
    error,
    refetch: fetchTenants,
    createTenant,
    updateTenant,
    deleteTenant,
    setCurrentTenant,
    currentTenantId,
  }
}

export interface UseTenantOptions {
  /**
   * Tenant ID to fetch
   */
  tenantId: string

  /**
   * Whether to automatically fetch tenant on mount
   * @default true
   */
  autoFetch?: boolean
}

export interface UseTenantReturn {
  /**
   * Tenant data
   */
  tenant: Tenant | null

  /**
   * Whether tenant is being fetched
   */
  isLoading: boolean

  /**
   * Any error that occurred
   */
  error: Error | null

  /**
   * Refetch tenant
   */
  refetch: () => Promise<void>

  /**
   * Update the tenant
   */
  update: (options: UpdateTenantOptions) => Promise<Tenant>

  /**
   * Delete the tenant
   */
  remove: () => Promise<void>
}

/**
 * Hook for managing a single tenant
 *
 * @example
 * ```tsx
 * function TenantDetails({ tenantId }: { tenantId: string }) {
 *   const { tenant, isLoading, update } = useTenant({ tenantId })
 *
 *   if (isLoading) return <div>Loading...</div>
 *   if (!tenant) return <div>Tenant not found</div>
 *
 *   return (
 *     <div>
 *       <h1>{tenant.name}</h1>
 *       <button onClick={() => update({ name: 'New Name' })}>
 *         Rename
 *       </button>
 *     </div>
 *   )
 * }
 * ```
 */
export function useTenant(options: UseTenantOptions): UseTenantReturn {
  const { tenantId, autoFetch = true } = options
  const client = useFluxbaseClient()

  const [tenant, setTenant] = useState<Tenant | null>(null)
  const [isLoading, setIsLoading] = useState(autoFetch)
  const [error, setError] = useState<Error | null>(null)

  const fetchTenant = useCallback(async () => {
    if (!tenantId) return

    try {
      setIsLoading(true)
      setError(null)
      const { data, error: fetchError } = await client.tenant.get(tenantId)
      if (fetchError) {
        setError(fetchError)
        setTenant(null)
      } else {
        setTenant(data)
      }
    } catch (err) {
      setError(err as Error)
      setTenant(null)
    } finally {
      setIsLoading(false)
    }
  }, [client, tenantId])

  const update = useCallback(
    async (opts: UpdateTenantOptions): Promise<Tenant> => {
      const { data, error: updateError } = await client.tenant.update(tenantId, opts)
      if (updateError) throw updateError
      setTenant(data)
      return data!
    },
    [client, tenantId]
  )

  const remove = useCallback(async (): Promise<void> => {
    const { error: deleteError } = await client.tenant.delete(tenantId)
    if (deleteError) throw deleteError
    setTenant(null)
  }, [client, tenantId])

  useEffect(() => {
    if (autoFetch && tenantId) {
      fetchTenant()
    }
  }, [autoFetch, fetchTenant, tenantId])

  return {
    tenant,
    isLoading,
    error,
    refetch: fetchTenant,
    update,
    remove,
  }
}

export interface UseTenantMembersOptions {
  /**
   * Tenant ID to fetch members for
   */
  tenantId: string

  /**
   * Whether to automatically fetch members on mount
   * @default true
   */
  autoFetch?: boolean
}

export interface MemberWithUser extends TenantMembership {
  email: string
  user_role: string
}

export interface UseTenantMembersReturn {
  /**
   * Array of tenant members
   */
  members: MemberWithUser[]

  /**
   * Whether members are being fetched
   */
  isLoading: boolean

  /**
   * Any error that occurred
   */
  error: Error | null

  /**
   * Refetch members
   */
  refetch: () => Promise<void>

  /**
   * Add a member to the tenant
   */
  addMember: (options: AddTenantMemberOptions) => Promise<TenantMembership>

  /**
   * Update a member's role
   */
  updateMemberRole: (userId: string, options: UpdateTenantMemberOptions) => Promise<void>

  /**
   * Remove a member from the tenant
   */
  removeMember: (userId: string) => Promise<void>
}

/**
 * Hook for managing tenant members
 *
 * @example
 * ```tsx
 * function TenantMembersList({ tenantId }: { tenantId: string }) {
 *   const { members, isLoading, addMember, removeMember } = useTenantMembers({ tenantId })
 *
 *   const handleAddMember = async () => {
 *     await addMember({
 *       user_id: 'user-uuid',
 *       role: 'tenant_member'
 *     })
 *   }
 *
 *   return (
 *     <div>
 *       <button onClick={handleAddMember}>Add Member</button>
 *       {members.map(m => (
 *         <div key={m.id}>
 *           {m.email} - {m.role}
 *           <button onClick={() => removeMember(m.user_id)}>Remove</button>
 *         </div>
 *       ))}
 *     </div>
 *   )
 * }
 * ```
 */
export function useTenantMembers(options: UseTenantMembersOptions): UseTenantMembersReturn {
  const { tenantId, autoFetch = true } = options
  const client = useFluxbaseClient()

  const [members, setMembers] = useState<MemberWithUser[]>([])
  const [isLoading, setIsLoading] = useState(autoFetch)
  const [error, setError] = useState<Error | null>(null)

  const fetchMembers = useCallback(async () => {
    if (!tenantId) return

    try {
      setIsLoading(true)
      setError(null)
      const { data, error: fetchError } = await client.tenant.listMembers(tenantId)
      if (fetchError) {
        setError(fetchError)
      } else {
        setMembers((data as MemberWithUser[]) || [])
      }
    } catch (err) {
      setError(err as Error)
    } finally {
      setIsLoading(false)
    }
  }, [client, tenantId])

  const addMember = useCallback(
    async (opts: AddTenantMemberOptions): Promise<TenantMembership> => {
      const { data, error: addError } = await client.tenant.addMember(tenantId, opts)
      if (addError) throw addError
      await fetchMembers()
      return data!
    },
    [client, tenantId, fetchMembers]
  )

  const updateMemberRole = useCallback(
    async (userId: string, opts: UpdateTenantMemberOptions): Promise<void> => {
      const { error: updateError } = await client.tenant.updateMember(tenantId, userId, opts)
      if (updateError) throw updateError
      await fetchMembers()
    },
    [client, tenantId, fetchMembers]
  )

  const removeMember = useCallback(
    async (userId: string): Promise<void> => {
      const { error: removeError } = await client.tenant.removeMember(tenantId, userId)
      if (removeError) throw removeError
      await fetchMembers()
    },
    [client, tenantId, fetchMembers]
  )

  useEffect(() => {
    if (autoFetch && tenantId) {
      fetchMembers()
    }
  }, [autoFetch, fetchMembers, tenantId])

  return {
    members,
    isLoading,
    error,
    refetch: fetchMembers,
    addMember,
    updateMemberRole,
    removeMember,
  }
}
