/**
 * FluxbaseTenant - Multi-tenant management module
 *
 * Provides methods for managing tenants and memberships.
 */
import { FluxbaseFetch } from "./fetch";
import type {
  Tenant,
  TenantMembership,
  TenantWithRole,
  CreateTenantOptions,
  UpdateTenantOptions,
  AddTenantMemberOptions,
  UpdateTenantMemberOptions,
} from "./types";

import type { FluxbaseResponse } from "./response";

/**
 * FluxbaseTenant provides multi-tenant management functionality
 *
 * @example
 * ```typescript
 * // List tenants I have access to
 * const { data } = await client.tenant.listMine()
 *
 * // Get tenant details
 * const { data } = await client.tenant.get('tenant-id')
 *
 * // Create a tenant (instance admin only)
 * const { data } = await client.tenant.create({
 *   slug: 'acme-corp',
 *   name: 'Acme Corporation'
 * })
 *
 * // Add member to tenant (tenant admin only)
 * await client.tenant.addMember('tenant-id', {
 *   user_id: 'user-id',
 *   role: 'tenant_member'
 * })
 * ```
 *
 * @category Multi-Tenancy
 */
export class FluxbaseTenant {
  constructor(private fetch: FluxbaseFetch) {}

  /**
   * List all tenants (instance admin only)
   *
   * @returns Promise with tenants list or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.list()
   * ```
   */
  async list(): Promise<FluxbaseResponse<Tenant[]>> {
    try {
      const data = await this.fetch.get<Tenant[]>("/tenants");
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * List tenants the current user has access to
   *
   * @returns Promise with tenants and user's role in each
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.listMine()
   * // data: [{ id: '...', slug: 'acme', name: 'Acme', my_role: 'tenant_admin' }]
   * ```
   */
  async listMine(): Promise<FluxbaseResponse<TenantWithRole[]>> {
    try {
      const data = await this.fetch.get<TenantWithRole[]>("/tenants/mine");
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Get a tenant by ID
   *
   * @param id - Tenant ID
   * @returns Promise with tenant details or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.get('tenant-id')
   * ```
   */
  async get(id: string): Promise<FluxbaseResponse<Tenant>> {
    try {
      const data = await this.fetch.get<Tenant>(`/tenants/${id}`);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Create a new tenant (instance admin only)
   *
   * @param options - Tenant creation options
   * @returns Promise with created tenant or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.create({
   *   slug: 'acme-corp',
   *   name: 'Acme Corporation',
   *   metadata: { plan: 'enterprise' }
   * })
   * ```
   */
  async create(options: CreateTenantOptions): Promise<FluxbaseResponse<Tenant>> {
    try {
      const data = await this.fetch.post<Tenant>("/tenants", options);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Update a tenant (tenant admin only)
   *
   * @param id - Tenant ID
   * @param options - Update options
   * @returns Promise with updated tenant or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.update('tenant-id', {
   *   name: 'New Name'
   * })
   * ```
   */
  async update(id: string, options: UpdateTenantOptions): Promise<FluxbaseResponse<Tenant>> {
    try {
      const data = await this.fetch.patch<Tenant>(`/tenants/${id}`, options);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Delete a tenant (instance admin only)
   *
   * @param id - Tenant ID
   * @returns Promise that resolves when deleted
   *
   * @example
   * ```typescript
   * const { error } = await client.tenant.delete('tenant-id')
   * ```
   */
  async delete(id: string): Promise<FluxbaseResponse<void>> {
    try {
      await this.fetch.delete(`/tenants/${id}`);
      return { data: undefined, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * List members of a tenant
   *
   * @param tenantId - Tenant ID
   * @returns Promise with member list or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.listMembers('tenant-id')
   * ```
   */
  async listMembers(tenantId: string): Promise<FluxbaseResponse<TenantMembership[]>> {
    try {
      const data = await this.fetch.get<TenantMembership[]>(`/tenants/${tenantId}/members`);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Add a member to a tenant (tenant admin only)
   *
   * @param tenantId - Tenant ID
   * @param options - Member options
   * @returns Promise with created membership or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.addMember('tenant-id', {
   *   user_id: 'user-id',
   *   role: 'tenant_member'
   * })
   * ```
   */
  async addMember(
    tenantId: string,
    options: AddTenantMemberOptions,
  ): Promise<FluxbaseResponse<TenantMembership>> {
    try {
      const data = await this.fetch.post<TenantMembership>(
        `/tenants/${tenantId}/members`,
        options,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Update a member's role (tenant admin only)
   *
   * @param tenantId - Tenant ID
   * @param userId - User ID
   * @param options - Update options
   * @returns Promise that resolves when updated
   *
   * @example
   * ```typescript
   * const { error } = await client.tenant.updateMember('tenant-id', 'user-id', {
   *   role: 'tenant_admin'
   * })
   * ```
   */
  async updateMember(
    tenantId: string,
    userId: string,
    options: UpdateTenantMemberOptions,
  ): Promise<FluxbaseResponse<void>> {
    try {
      await this.fetch.patch(
        `/tenants/${tenantId}/members/${userId}`,
        options,
      );
      return { data: undefined, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Remove a member from a tenant (tenant admin only)
   *
   * @param tenantId - Tenant ID
   * @param userId - User ID
   * @returns Promise that resolves when removed
   *
   * @example
   * ```typescript
   * const { error } = await client.tenant.removeMember('tenant-id', 'user-id')
   * ```
   */
  async removeMember(
    tenantId: string,
    userId: string,
  ): Promise<FluxbaseResponse<void>> {
    try {
      await this.fetch.delete(`/tenants/${tenantId}/members/${userId}`);
      return { data: undefined, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }
}
