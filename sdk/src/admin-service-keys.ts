import type { FluxbaseFetch } from "./fetch";
import type {
  ServiceKey,
  ServiceKeyWithKey,
  CreateServiceKeyRequest,
  UpdateServiceKeyRequest,
  RevokeServiceKeyRequest,
  DeprecateServiceKeyRequest,
} from "./types";

/**
 * Service Keys Manager
 *
 * Manages service keys (anon and service) for tenant databases.
 * Each tenant has their own auth.service_keys table.
 *
 * @example
 * ```typescript
 * // List all service keys
 * const { data, error } = await client.admin.serviceKeys.list()
 *
 * // Create a new service key
 * const { data, error } = await client.admin.serviceKeys.create({
 *   name: 'Production API Key',
 *   key_type: 'service',
 *   scopes: ['*']
 * })
 *
 * // Rotate a key
 * const { data, error } = await client.admin.serviceKeys.rotate('key-id')
 * ```
 *
 * @category Admin
 */
export class ServiceKeysManager {
  private fetch: FluxbaseFetch;

  constructor(fetch: FluxbaseFetch) {
    this.fetch = fetch;
  }

  /**
   * List all service keys
   *
   * @returns List of service keys
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.list()
   * ```
   */
  async list(): Promise<{ data: ServiceKey[] | null; error: Error | null }> {
    try {
      const data = await this.fetch.get<ServiceKey[]>("/api/v1/admin/service-keys");
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Get a service key by ID
   *
   * @param id - Service key ID
   * @returns Service key details
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.get('key-id')
   * ```
   */
  async get(
    id: string,
  ): Promise<{ data: ServiceKey | null; error: Error | null }> {
    try {
      const data = await this.fetch.get<ServiceKey>(`/api/v1/admin/service-keys/${id}`);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Create a new service key
   *
   * The full key value is only returned once - store it securely!
   *
   * @param request - Key creation options
   * @returns Created key with full key value
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.create({
   *   name: 'Production API Key',
   *   key_type: 'service',
   *   scopes: ['*'],
   *   rate_limit_per_minute: 1000
   * })
   *
   * if (data) {
   *   // Store data.key securely - it won't be shown again!
   *   console.log('Key created:', data.key)
   * }
   * ```
   */
  async create(
    request: CreateServiceKeyRequest,
  ): Promise<{ data: ServiceKeyWithKey | null; error: Error | null }> {
    try {
      const data = await this.fetch.post<ServiceKeyWithKey>(
        "/api/v1/admin/service-keys",
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Update a service key
   *
   * @param id - Service key ID
   * @param request - Update options
   * @returns Updated key
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.update('key-id', {
   *   name: 'New Name',
   *   rate_limit_per_minute: 2000
   * })
   * ```
   */
  async update(
    id: string,
    request: UpdateServiceKeyRequest,
  ): Promise<{ data: ServiceKey | null; error: Error | null }> {
    try {
      const data = await this.fetch.put<ServiceKey>(
        `/api/v1/admin/service-keys/${id}`,
        request,
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Delete a service key permanently
   *
   * @param id - Service key ID
   * @returns Success or error
   *
   * @example
   * ```typescript
   * const { error } = await client.admin.serviceKeys.delete('key-id')
   * ```
   */
  async delete(id: string): Promise<{ error: Error | null }> {
    try {
      await this.fetch.delete(`/api/v1/admin/service-keys/${id}`);
      return { error: null };
    } catch (error) {
      return { error: error as Error };
    }
  }

  /**
   * Disable a service key (temporarily)
   *
   * @param id - Service key ID
   * @returns Success or error
   *
   * @example
   * ```typescript
   * const { error } = await client.admin.serviceKeys.disable('key-id')
   * ```
   */
  async disable(id: string): Promise<{ error: Error | null }> {
    try {
      await this.fetch.post(`/api/v1/admin/service-keys/${id}/disable`, {});
      return { error: null };
    } catch (error) {
      return { error: error as Error };
    }
  }

  /**
   * Enable a disabled service key
   *
   * @param id - Service key ID
   * @returns Success or error
   *
   * @example
   * ```typescript
   * const { error } = await client.admin.serviceKeys.enable('key-id')
   * ```
   */
  async enable(id: string): Promise<{ error: Error | null }> {
    try {
      await this.fetch.post(`/api/v1/admin/service-keys/${id}/enable`, {});
      return { error: null };
    } catch (error) {
      return { error: error as Error };
    }
  }

  /**
   * Revoke a service key permanently (emergency)
   *
   * Use for immediate revocation when a key is compromised.
   *
   * @param id - Service key ID
   * @param request - Revocation options
   * @returns Success or error
   *
   * @example
   * ```typescript
   * const { error } = await client.admin.serviceKeys.revoke('key-id', {
   *   reason: 'Key was compromised'
   * })
   * ```
   */
  async revoke(
    id: string,
    request?: RevokeServiceKeyRequest,
  ): Promise<{ error: Error | null }> {
    try {
      const body = request?.reason
        ? new URLSearchParams({ reason: request.reason })
        : {};
      await this.fetch.post(`/api/v1/admin/service-keys/${id}/revoke`, body);
      return { error: null };
    } catch (error) {
      return { error: error as Error };
    }
  }

  /**
   * Deprecate a service key (graceful rotation)
   *
   * Marks the key for removal but keeps it active during grace period.
   *
   * @param id - Service key ID
   * @param request - Deprecation options
   * @returns Deprecation details
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.deprecate('key-id', {
   *   reason: 'Rotating to new key',
   *   grace_period_hours: 48
   * })
   * ```
   */
  async deprecate(
    id: string,
    request?: DeprecateServiceKeyRequest,
  ): Promise<{
    data: { deprecated_at: string; grace_period_ends_at: string } | null;
    error: Error | null;
  }> {
    try {
      const params = new URLSearchParams();
      if (request?.grace_period_hours) {
        params.set("grace_period_hours", String(request.grace_period_hours));
      }
      const body = request?.grace_period_hours
        ? Object.fromEntries(params)
        : {};
      const data = await this.fetch.post<{
        deprecated_at: string;
        grace_period_ends_at: string;
      }>(`/api/v1/admin/service-keys/${id}/deprecate`, body);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Rotate a service key (create replacement)
   *
   * Creates a new key with the same settings and deprecates the old one.
   * The new key is returned with its full value.
   *
   * @param id - Service key ID to rotate
   * @returns New key with full key value
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.rotate('old-key-id')
   *
   * if (data) {
   *   console.log('New key:', data.key)
   *   console.log('Old key deprecated at:', data.deprecated_at)
   * }
   * ```
   */
  async rotate(
    id: string,
  ): Promise<{ data: ServiceKeyWithKey | null; error: Error | null }> {
    try {
      const data = await this.fetch.post<ServiceKeyWithKey>(
        `/api/v1/admin/service-keys/${id}/rotate`,
        {},
      );
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }

  /**
   * Get revocation history for a service key
   *
   * @param id - Service key ID
   * @returns Revocation history
   *
   * @example
   * ```typescript
   * const { data, error } = await client.admin.serviceKeys.getRevocationHistory('key-id')
   * ```
   */
  async getRevocationHistory(id: string): Promise<{
    data: {
      id: string;
      name: string;
      revoked_at: string;
      revoked_by: string;
      revocation_reason: string;
    } | null;
    error: Error | null;
  }> {
    try {
      const data = await this.fetch.get<{
        id: string;
        name: string;
        revoked_at: string;
        revoked_by: string;
        revocation_reason: string;
      }>(`/api/v1/admin/service-keys/${id}/revocations`);
      return { data, error: null };
    } catch (error) {
      return { data: null, error: error as Error };
    }
  }
}
