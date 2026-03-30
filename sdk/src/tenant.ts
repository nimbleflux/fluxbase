/**
 * FluxbaseTenant - Multi-tenant management module
 *
 * Provides methods for managing tenants and admin assignments.
 * With database-per-tenant architecture, each tenant has its own isolated database.
 */
import { FluxbaseFetch } from "./fetch";
import type {
  Tenant,
  TenantAdminAssignment,
  TenantWithRole,
  CreateTenantOptions,
  UpdateTenantOptions,
  AssignAdminOptions,
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
 * // Assign admin to tenant (tenant admin only)
 * await client.tenant.assignAdmin('tenant-id', {
 *   user_id: 'user-id'
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
   * // data: [{ id: '...', slug: 'acme', name: 'Acme', my_role: 'tenant_admin', status: 'active' }]
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
   * This creates a new isolated database for the tenant.
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
  async create(
    options: CreateTenantOptions,
  ): Promise<FluxbaseResponse<Tenant>> {
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
  async update(
    id: string,
    options: UpdateTenantOptions,
  ): Promise<FluxbaseResponse<Tenant>> {
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
   * This permanently deletes the tenant's database and all its data.
   * Cannot delete the default tenant.
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
   * Migrate a tenant database to the latest schema (instance admin only)
   *
   * @param id - Tenant ID
   * @returns Promise with migration status or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.migrate('tenant-id')
   * // data: { status: 'migrated' }
   * ```
   */
  async migrate(id: string): Promise<FluxbaseResponse<{ status: string }>> {
    try {
      const data = await this.fetch.post<{ status: string }>(
        `/tenants/${id}/migrate`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * List admins of a tenant
   *
   * @param tenantId - Tenant ID
   * @returns Promise with admin list or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.listAdmins('tenant-id')
   * // data: [{ id: '...', tenant_id: '...', user_id: '...', email: 'admin@example.com' }]
   * ```
   */
  async listAdmins(
    tenantId: string,
  ): Promise<FluxbaseResponse<TenantAdminAssignment[]>> {
    try {
      const data = await this.fetch.get<TenantAdminAssignment[]>(
        `/tenants/${tenantId}/admins`,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Assign an admin to a tenant (tenant admin only)
   *
   * @param tenantId - Tenant ID
   * @param options - Admin assignment options
   * @returns Promise with created assignment or error
   *
   * @example
   * ```typescript
   * const { data, error } = await client.tenant.assignAdmin('tenant-id', {
   *   user_id: 'user-id'
   * })
   * ```
   */
  async assignAdmin(
    tenantId: string,
    options: AssignAdminOptions,
  ): Promise<FluxbaseResponse<TenantAdminAssignment>> {
    try {
      const data = await this.fetch.post<TenantAdminAssignment>(
        `/tenants/${tenantId}/admins`,
        options,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Remove an admin from a tenant (tenant admin only)
   *
   * @param tenantId - Tenant ID
   * @param userId - User ID
   * @returns Promise that resolves when removed
   *
   * @example
   * ```typescript
   * const { error } = await client.tenant.removeAdmin('tenant-id', 'user-id')
   * ```
   */
  async removeAdmin(
    tenantId: string,
    userId: string,
  ): Promise<FluxbaseResponse<void>> {
    try {
      await this.fetch.delete(`/tenants/${tenantId}/admins/${userId}`);
      return { data: undefined, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }
}
