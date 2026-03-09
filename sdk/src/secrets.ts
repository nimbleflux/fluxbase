/**
 * Secrets Management module for Fluxbase SDK
 *
 * Provides methods for managing secrets that are injected into edge functions
 * and background jobs at runtime. Secrets are encrypted at rest and scoped
 * to either global or namespace level.
 *
 * @example
 * ```typescript
 * // Create a secret
 * const secret = await client.secrets.create({
 *   name: 'API_KEY',
 *   value: 'sk-your-api-key',
 *   scope: 'global'
 * })
 *
 * // Get secret by name
 * const secret = await client.secrets.get('API_KEY')
 *
 * // Update secret by name
 * await client.secrets.update('API_KEY', { value: 'new-value' })
 *
 * // List all secrets
 * const secrets = await client.secrets.list()
 * ```
 *
 * @category Secrets
 */

import type { FluxbaseFetch } from './fetch'

/**
 * Represents a stored secret (metadata only, never includes value)
 */
export interface Secret {
  id: string
  name: string
  scope: 'global' | 'namespace'
  namespace?: string
  description?: string
  version: number
  expires_at?: string
  created_at: string
  updated_at: string
  created_by?: string
  updated_by?: string
}

/**
 * Summary of a secret for list responses
 */
export interface SecretSummary {
  id: string
  name: string
  scope: 'global' | 'namespace'
  namespace?: string
  description?: string
  version: number
  expires_at?: string
  is_expired: boolean
  created_at: string
  updated_at: string
  created_by?: string
  updated_by?: string
}

/**
 * Represents a historical version of a secret
 */
export interface SecretVersion {
  id: string
  secret_id: string
  version: number
  created_at: string
  created_by?: string
}

/**
 * Statistics about secrets
 */
export interface SecretStats {
  total: number
  expiring_soon: number
  expired: number
}

/**
 * Request to create a new secret
 */
export interface CreateSecretRequest {
  name: string
  value: string
  scope?: 'global' | 'namespace'
  namespace?: string
  description?: string
  expires_at?: string
}

/**
 * Request to update an existing secret
 */
export interface UpdateSecretRequest {
  value?: string
  description?: string
  expires_at?: string
}

/**
 * Options for listing secrets
 */
export interface ListSecretsOptions {
  scope?: 'global' | 'namespace'
  namespace?: string
}

/**
 * Options for name-based secret operations
 */
export interface SecretByNameOptions {
  namespace?: string
}

/**
 * Secrets Manager for managing edge function and job secrets
 *
 * Provides both name-based (recommended) and UUID-based operations.
 * Name-based operations are more convenient for most use cases.
 *
 * @example
 * ```typescript
 * const client = createClient({ url: 'http://localhost:8080' })
 * await client.auth.login({ email: 'user@example.com', password: 'password' })
 *
 * // Create a global secret
 * const secret = await client.secrets.create({
 *   name: 'STRIPE_KEY',
 *   value: 'sk_live_xxx',
 *   description: 'Stripe production API key'
 * })
 *
 * // Create a namespace-scoped secret
 * await client.secrets.create({
 *   name: 'DATABASE_URL',
 *   value: 'postgres://...',
 *   scope: 'namespace',
 *   namespace: 'production'
 * })
 *
 * // Get secret by name
 * const secret = await client.secrets.get('STRIPE_KEY')
 *
 * // Get namespace-scoped secret
 * const secret = await client.secrets.get('DATABASE_URL', { namespace: 'production' })
 *
 * // Update secret
 * await client.secrets.update('STRIPE_KEY', { value: 'sk_live_new_key' })
 *
 * // List all secrets
 * const secrets = await client.secrets.list()
 *
 * // Get version history
 * const versions = await client.secrets.getVersions('STRIPE_KEY')
 *
 * // Rollback to previous version
 * await client.secrets.rollback('STRIPE_KEY', 1)
 *
 * // Delete secret
 * await client.secrets.delete('STRIPE_KEY')
 * ```
 *
 * @category Secrets
 */
export class SecretsManager {
  private fetch: FluxbaseFetch

  constructor(fetch: FluxbaseFetch) {
    this.fetch = fetch
  }

  /**
   * Create a new secret
   *
   * Creates a new secret with the specified name, value, and scope.
   * The value is encrypted at rest and never returned by the API.
   *
   * @param request - Secret creation request
   * @returns Promise resolving to the created secret (metadata only)
   *
   * @example
   * ```typescript
   * // Create a global secret
   * const secret = await client.secrets.create({
   *   name: 'SENDGRID_API_KEY',
   *   value: 'SG.xxx',
   *   description: 'SendGrid API key for transactional emails'
   * })
   *
   * // Create a namespace-scoped secret
   * const secret = await client.secrets.create({
   *   name: 'DATABASE_URL',
   *   value: 'postgres://user:pass@host:5432/db',
   *   scope: 'namespace',
   *   namespace: 'production',
   *   description: 'Production database URL'
   * })
   *
   * // Create a secret with expiration
   * const secret = await client.secrets.create({
   *   name: 'TEMP_TOKEN',
   *   value: 'xyz123',
   *   expires_at: '2025-12-31T23:59:59Z'
   * })
   * ```
   */
  async create(request: CreateSecretRequest): Promise<Secret> {
    return await this.fetch.post<Secret>('/api/v1/secrets', request)
  }

  /**
   * Get a secret by name (metadata only, never includes value)
   *
   * @param name - Secret name
   * @param options - Optional namespace for namespace-scoped secrets
   * @returns Promise resolving to the secret
   *
   * @example
   * ```typescript
   * // Get a global secret
   * const secret = await client.secrets.get('API_KEY')
   *
   * // Get a namespace-scoped secret
   * const secret = await client.secrets.get('DATABASE_URL', { namespace: 'production' })
   * ```
   */
  async get(name: string, options?: SecretByNameOptions): Promise<Secret> {
    const params = new URLSearchParams()
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    return await this.fetch.get<Secret>(`/api/v1/secrets/by-name/${encodeURIComponent(name)}${query}`)
  }

  /**
   * Update a secret by name
   *
   * Updates the secret's value, description, or expiration.
   * Only provided fields will be updated.
   *
   * @param name - Secret name
   * @param request - Update request
   * @param options - Optional namespace for namespace-scoped secrets
   * @returns Promise resolving to the updated secret
   *
   * @example
   * ```typescript
   * // Update secret value
   * const secret = await client.secrets.update('API_KEY', { value: 'new-value' })
   *
   * // Update description
   * const secret = await client.secrets.update('API_KEY', { description: 'Updated description' })
   *
   * // Update namespace-scoped secret
   * const secret = await client.secrets.update('DATABASE_URL',
   *   { value: 'postgres://new-host:5432/db' },
   *   { namespace: 'production' }
   * )
   * ```
   */
  async update(name: string, request: UpdateSecretRequest, options?: SecretByNameOptions): Promise<Secret> {
    const params = new URLSearchParams()
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    return await this.fetch.put<Secret>(`/api/v1/secrets/by-name/${encodeURIComponent(name)}${query}`, request)
  }

  /**
   * Delete a secret by name
   *
   * Permanently deletes the secret and all its versions.
   *
   * @param name - Secret name
   * @param options - Optional namespace for namespace-scoped secrets
   * @returns Promise resolving when deletion is complete
   *
   * @example
   * ```typescript
   * // Delete a global secret
   * await client.secrets.delete('OLD_API_KEY')
   *
   * // Delete a namespace-scoped secret
   * await client.secrets.delete('DATABASE_URL', { namespace: 'staging' })
   * ```
   */
  async delete(name: string, options?: SecretByNameOptions): Promise<void> {
    const params = new URLSearchParams()
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    await this.fetch.delete(`/api/v1/secrets/by-name/${encodeURIComponent(name)}${query}`)
  }

  /**
   * Get version history for a secret by name
   *
   * Returns all historical versions of the secret (values are never included).
   *
   * @param name - Secret name
   * @param options - Optional namespace for namespace-scoped secrets
   * @returns Promise resolving to array of secret versions
   *
   * @example
   * ```typescript
   * const versions = await client.secrets.getVersions('API_KEY')
   *
   * versions.forEach(v => {
   *   console.log(`Version ${v.version} created at ${v.created_at}`)
   * })
   * ```
   */
  async getVersions(name: string, options?: SecretByNameOptions): Promise<SecretVersion[]> {
    const params = new URLSearchParams()
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    return await this.fetch.get<SecretVersion[]>(`/api/v1/secrets/by-name/${encodeURIComponent(name)}/versions${query}`)
  }

  /**
   * Rollback a secret to a previous version by name
   *
   * Restores the secret to a previous version's value.
   * This creates a new version with the old value.
   *
   * @param name - Secret name
   * @param version - Version number to rollback to
   * @param options - Optional namespace for namespace-scoped secrets
   * @returns Promise resolving to the updated secret
   *
   * @example
   * ```typescript
   * // Rollback to version 2
   * const secret = await client.secrets.rollback('API_KEY', 2)
   * console.log(`Secret now at version ${secret.version}`)
   * ```
   */
  async rollback(name: string, version: number, options?: SecretByNameOptions): Promise<Secret> {
    const params = new URLSearchParams()
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    return await this.fetch.post<Secret>(
      `/api/v1/secrets/by-name/${encodeURIComponent(name)}/rollback/${version}${query}`,
      {}
    )
  }

  /**
   * List all secrets (metadata only, never includes values)
   *
   * @param options - Filter options for scope and namespace
   * @returns Promise resolving to array of secret summaries
   *
   * @example
   * ```typescript
   * // List all secrets
   * const secrets = await client.secrets.list()
   *
   * // List only global secrets
   * const secrets = await client.secrets.list({ scope: 'global' })
   *
   * // List secrets for a specific namespace
   * const secrets = await client.secrets.list({ namespace: 'production' })
   *
   * secrets.forEach(s => {
   *   console.log(`${s.name}: version ${s.version}, expired: ${s.is_expired}`)
   * })
   * ```
   */
  async list(options?: ListSecretsOptions): Promise<SecretSummary[]> {
    const params = new URLSearchParams()
    if (options?.scope) {
      params.set('scope', options.scope)
    }
    if (options?.namespace) {
      params.set('namespace', options.namespace)
    }
    const query = params.toString() ? `?${params.toString()}` : ''
    return await this.fetch.get<SecretSummary[]>(`/api/v1/secrets${query}`)
  }

  /**
   * Get statistics about secrets
   *
   * @returns Promise resolving to secret statistics
   *
   * @example
   * ```typescript
   * const stats = await client.secrets.stats()
   * console.log(`Total: ${stats.total}, Expiring soon: ${stats.expiring_soon}, Expired: ${stats.expired}`)
   * ```
   */
  async stats(): Promise<SecretStats> {
    return await this.fetch.get<SecretStats>('/api/v1/secrets/stats')
  }

  // UUID-based methods for backward compatibility

  /**
   * Get a secret by ID (metadata only)
   *
   * @param id - Secret UUID
   * @returns Promise resolving to the secret
   *
   * @example
   * ```typescript
   * const secret = await client.secrets.getById('550e8400-e29b-41d4-a716-446655440000')
   * ```
   */
  async getById(id: string): Promise<Secret> {
    return await this.fetch.get<Secret>(`/api/v1/secrets/${encodeURIComponent(id)}`)
  }

  /**
   * Update a secret by ID
   *
   * @param id - Secret UUID
   * @param request - Update request
   * @returns Promise resolving to the updated secret
   *
   * @example
   * ```typescript
   * const secret = await client.secrets.updateById('550e8400-e29b-41d4-a716-446655440000', {
   *   value: 'new-value'
   * })
   * ```
   */
  async updateById(id: string, request: UpdateSecretRequest): Promise<Secret> {
    return await this.fetch.put<Secret>(`/api/v1/secrets/${encodeURIComponent(id)}`, request)
  }

  /**
   * Delete a secret by ID
   *
   * @param id - Secret UUID
   * @returns Promise resolving when deletion is complete
   *
   * @example
   * ```typescript
   * await client.secrets.deleteById('550e8400-e29b-41d4-a716-446655440000')
   * ```
   */
  async deleteById(id: string): Promise<void> {
    await this.fetch.delete(`/api/v1/secrets/${encodeURIComponent(id)}`)
  }

  /**
   * Get version history for a secret by ID
   *
   * @param id - Secret UUID
   * @returns Promise resolving to array of secret versions
   *
   * @example
   * ```typescript
   * const versions = await client.secrets.getVersionsById('550e8400-e29b-41d4-a716-446655440000')
   * ```
   */
  async getVersionsById(id: string): Promise<SecretVersion[]> {
    return await this.fetch.get<SecretVersion[]>(`/api/v1/secrets/${encodeURIComponent(id)}/versions`)
  }

  /**
   * Rollback a secret to a previous version by ID
   *
   * @param id - Secret UUID
   * @param version - Version number to rollback to
   * @returns Promise resolving to the updated secret
   *
   * @example
   * ```typescript
   * const secret = await client.secrets.rollbackById('550e8400-e29b-41d4-a716-446655440000', 2)
   * ```
   */
  async rollbackById(id: string, version: number): Promise<Secret> {
    return await this.fetch.post<Secret>(`/api/v1/secrets/${encodeURIComponent(id)}/rollback/${version}`, {})
  }
}
