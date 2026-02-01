/**
 * Branching Module Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { FluxbaseBranching } from './branching'
import type { FluxbaseFetch } from './fetch'
import type { Branch, BranchActivity, BranchPoolStats, ListBranchesResponse } from './types'

// Mock FluxbaseFetch
class MockFetch implements Partial<FluxbaseFetch> {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public mockResponse: any = null
  public shouldThrow: boolean = false
  public errorMessage: string = 'Test error'
  public callCount: number = 0
  public responseQueue: any[] = []

  async get<T>(path: string): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'GET'
    this.callCount++
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    if (this.responseQueue.length > 0) {
      return this.responseQueue.shift() as T
    }
    return this.mockResponse as T
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'POST'
    this.lastBody = body
    this.callCount++
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }

  async delete(path: string): Promise<void> {
    this.lastUrl = path
    this.lastMethod = 'DELETE'
    this.callCount++
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
  }
}

describe('FluxbaseBranching', () => {
  let mockFetch: MockFetch
  let branching: FluxbaseBranching

  beforeEach(() => {
    mockFetch = new MockFetch()
    branching = new FluxbaseBranching(mockFetch as unknown as FluxbaseFetch)
  })

  describe('list', () => {
    it('should list all branches', async () => {
      const mockBranches: ListBranchesResponse = {
        branches: [
          { id: 'b1', slug: 'main', status: 'ready', type: 'main' } as Branch,
          { id: 'b2', slug: 'feature-auth', status: 'ready', type: 'persistent' } as Branch,
        ],
        count: 2,
      }
      mockFetch.mockResponse = mockBranches

      const { data, error } = await branching.list()

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockBranches)
      expect(error).toBeNull()
    })

    it('should filter by status', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ status: 'ready' })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?status=ready')
    })

    it('should filter by type', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ type: 'preview' })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?type=preview')
    })

    it('should filter by GitHub repo', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ githubRepo: 'owner/repo' })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?github_repo=owner%2Frepo')
    })

    it('should filter by mine flag', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ mine: true })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?mine=true')
    })

    it('should support pagination with limit', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ limit: 10 })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?limit=10')
    })

    it('should support pagination with offset', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({ offset: 20 })

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches?offset=20')
    })

    it('should combine multiple filters', async () => {
      mockFetch.mockResponse = { branches: [], count: 0 }

      await branching.list({
        status: 'ready',
        type: 'preview',
        mine: true,
        limit: 10,
        offset: 5,
      })

      expect(mockFetch.lastUrl).toContain('status=ready')
      expect(mockFetch.lastUrl).toContain('type=preview')
      expect(mockFetch.lastUrl).toContain('mine=true')
      expect(mockFetch.lastUrl).toContain('limit=10')
      expect(mockFetch.lastUrl).toContain('offset=5')
    })

    it('should handle list errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Permission denied'

      const { data, error } = await branching.list()

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Permission denied')
    })
  })

  describe('get', () => {
    it('should get branch by slug', async () => {
      const mockBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'ready',
        type: 'persistent',
      }
      mockFetch.mockResponse = mockBranch

      const { data, error } = await branching.get('feature-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature-auth')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockBranch)
      expect(error).toBeNull()
    })

    it('should get branch by ID', async () => {
      mockFetch.mockResponse = { id: 'uuid-123', slug: 'test' }

      await branching.get('uuid-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/uuid-123')
    })

    it('should encode slug in URL', async () => {
      mockFetch.mockResponse = { id: 'b1', slug: 'feature/add-auth' }

      await branching.get('feature/add-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature%2Fadd-auth')
    })

    it('should handle get errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Branch not found'

      const { data, error } = await branching.get('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch not found')
    })
  })

  describe('create', () => {
    it('should create a simple branch', async () => {
      const mockBranch: Partial<Branch> = {
        id: 'new-branch-id',
        slug: 'feature-auth',
        status: 'creating',
      }
      mockFetch.mockResponse = mockBranch

      const { data, error } = await branching.create('feature-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({ name: 'feature-auth' })
      expect(data).toEqual(mockBranch)
      expect(error).toBeNull()
    })

    it('should create branch with parent', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'child-branch' }

      await branching.create('child-branch', {
        parentBranchId: 'parent-uuid',
      })

      expect(mockFetch.lastBody).toMatchObject({
        parent_branch_id: 'parent-uuid',
      })
    })

    it('should create branch with schema_only data clone mode', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'test' }

      await branching.create('test', {
        dataCloneMode: 'schema_only',
      })

      expect(mockFetch.lastBody).toMatchObject({
        data_clone_mode: 'schema_only',
      })
    })

    it('should create branch with full_clone data clone mode', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'test' }

      await branching.create('test', {
        dataCloneMode: 'full_clone',
      })

      expect(mockFetch.lastBody).toMatchObject({
        data_clone_mode: 'full_clone',
      })
    })

    it('should create branch with type', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'test' }

      await branching.create('test', {
        type: 'preview',
      })

      expect(mockFetch.lastBody).toMatchObject({
        type: 'preview',
      })
    })

    it('should create branch with GitHub PR info', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'pr-123' }

      await branching.create('pr-123', {
        githubPRNumber: 123,
        githubPRUrl: 'https://github.com/owner/repo/pull/123',
        githubRepo: 'owner/repo',
      })

      expect(mockFetch.lastBody).toMatchObject({
        github_pr_number: 123,
        github_pr_url: 'https://github.com/owner/repo/pull/123',
        github_repo: 'owner/repo',
      })
    })

    it('should create branch with expiration', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'temp' }

      await branching.create('temp', {
        expiresIn: '7d',
      })

      expect(mockFetch.lastBody).toMatchObject({
        expires_in: '7d',
      })
    })

    it('should create branch with all options', async () => {
      mockFetch.mockResponse = { id: 'new', slug: 'full-options' }

      await branching.create('full-options', {
        parentBranchId: 'parent-uuid',
        dataCloneMode: 'schema_only',
        type: 'preview',
        githubPRNumber: 456,
        githubPRUrl: 'https://github.com/owner/repo/pull/456',
        githubRepo: 'owner/repo',
        expiresIn: '24h',
      })

      expect(mockFetch.lastBody).toEqual({
        name: 'full-options',
        parent_branch_id: 'parent-uuid',
        data_clone_mode: 'schema_only',
        type: 'preview',
        github_pr_number: 456,
        github_pr_url: 'https://github.com/owner/repo/pull/456',
        github_repo: 'owner/repo',
        expires_in: '24h',
      })
    })

    it('should handle create errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Branch limit exceeded'

      const { data, error } = await branching.create('new-branch')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch limit exceeded')
    })
  })

  describe('delete', () => {
    it('should delete a branch', async () => {
      const { error } = await branching.delete('feature-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature-auth')
      expect(mockFetch.lastMethod).toBe('DELETE')
      expect(error).toBeNull()
    })

    it('should encode slug in URL', async () => {
      await branching.delete('feature/add-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature%2Fadd-auth')
    })

    it('should handle delete errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Cannot delete main branch'

      const { error } = await branching.delete('main')

      expect(error).toBeDefined()
      expect(error?.message).toBe('Cannot delete main branch')
    })
  })

  describe('reset', () => {
    it('should reset a branch', async () => {
      const mockBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'creating',
      }
      mockFetch.mockResponse = mockBranch

      const { data, error } = await branching.reset('feature-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature-auth/reset')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({})
      expect(data).toEqual(mockBranch)
      expect(error).toBeNull()
    })

    it('should encode slug in URL', async () => {
      mockFetch.mockResponse = { id: 'b1', slug: 'feature/add-auth' }

      await branching.reset('feature/add-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature%2Fadd-auth/reset')
    })

    it('should handle reset errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Cannot reset main branch'

      const { data, error } = await branching.reset('main')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Cannot reset main branch')
    })
  })

  describe('getActivity', () => {
    it('should get branch activity', async () => {
      const mockActivity: BranchActivity[] = [
        {
          id: 'a1',
          branch_id: 'b1',
          action: 'created',
          status: 'completed',
          created_at: '2025-01-01T00:00:00Z',
        },
        {
          id: 'a2',
          branch_id: 'b1',
          action: 'reset',
          status: 'completed',
          created_at: '2025-01-02T00:00:00Z',
        },
      ]
      mockFetch.mockResponse = { activity: mockActivity }

      const { data, error } = await branching.getActivity('feature-auth')

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature-auth/activity?limit=50')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockActivity)
      expect(error).toBeNull()
    })

    it('should support custom limit', async () => {
      mockFetch.mockResponse = { activity: [] }

      await branching.getActivity('feature-auth', 100)

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/feature-auth/activity?limit=100')
    })

    it('should encode slug in URL', async () => {
      mockFetch.mockResponse = { activity: [] }

      await branching.getActivity('feature/add-auth')

      expect(mockFetch.lastUrl).toContain('/api/v1/admin/branches/feature%2Fadd-auth/activity')
    })

    it('should handle getActivity errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Branch not found'

      const { data, error } = await branching.getActivity('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch not found')
    })
  })

  describe('getPoolStats', () => {
    it('should get pool statistics', async () => {
      const mockStats: BranchPoolStats[] = [
        {
          branch_id: 'b1',
          slug: 'main',
          active_connections: 5,
          idle_connections: 10,
          total_connections: 15,
        },
        {
          branch_id: 'b2',
          slug: 'feature-auth',
          active_connections: 2,
          idle_connections: 3,
          total_connections: 5,
        },
      ]
      mockFetch.mockResponse = { pools: mockStats }

      const { data, error } = await branching.getPoolStats()

      expect(mockFetch.lastUrl).toBe('/api/v1/admin/branches/stats/pools')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockStats)
      expect(error).toBeNull()
    })

    it('should handle getPoolStats errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Permission denied'

      const { data, error } = await branching.getPoolStats()

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Permission denied')
    })
  })

  describe('exists', () => {
    it('should return true when branch exists', async () => {
      mockFetch.mockResponse = { id: 'b1', slug: 'feature-auth' }

      const exists = await branching.exists('feature-auth')

      expect(exists).toBe(true)
    })

    it('should return false when branch does not exist', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Branch not found'

      const exists = await branching.exists('non-existent')

      expect(exists).toBe(false)
    })

    it('should return false when get returns null', async () => {
      mockFetch.mockResponse = null

      const exists = await branching.exists('null-branch')

      expect(exists).toBe(false)
    })
  })

  describe('waitForReady', () => {
    it('should return immediately when branch is ready', async () => {
      const readyBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'ready',
      }
      mockFetch.mockResponse = readyBranch

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toEqual(readyBranch)
      expect(error).toBeNull()
      expect(mockFetch.callCount).toBe(1)
    })

    it('should poll until branch is ready', async () => {
      const creatingBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'creating',
      }
      const readyBranch: Partial<Branch> = {
        ...creatingBranch,
        status: 'ready',
      }

      mockFetch.responseQueue = [creatingBranch, readyBranch]

      const { data, error } = await branching.waitForReady('feature-auth', {
        pollInterval: 10,
        timeout: 5000,
      })

      expect(data).toEqual(readyBranch)
      expect(error).toBeNull()
      expect(mockFetch.callCount).toBe(2)
    })

    it('should return error when branch has error status', async () => {
      const errorBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'error',
        error_message: 'Database creation failed',
      }
      mockFetch.mockResponse = errorBranch

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Database creation failed')
    })

    it('should return error when branch status is deleted', async () => {
      const deletedBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'deleted',
      }
      mockFetch.mockResponse = deletedBranch

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch was deleted')
    })

    it('should return error when branch status is deleting', async () => {
      const deletingBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'deleting',
      }
      mockFetch.mockResponse = deletingBranch

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch was deleted')
    })

    it('should return error when branch not found', async () => {
      mockFetch.mockResponse = null

      const { data, error } = await branching.waitForReady('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch not found')
    })

    it('should timeout after specified time', async () => {
      const creatingBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'creating',
      }
      mockFetch.mockResponse = creatingBranch

      const { data, error } = await branching.waitForReady('feature-auth', {
        timeout: 50,
        pollInterval: 20,
      })

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toContain('Timeout waiting for branch')
    })

    it('should return error on get failure', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Network error'

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Network error')
    })

    it('should use default timeout and pollInterval', async () => {
      const readyBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'ready',
      }
      mockFetch.mockResponse = readyBranch

      // Should work with no options (uses defaults)
      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toEqual(readyBranch)
      expect(error).toBeNull()
    })

    it('should return error with default message when error_message is null', async () => {
      const errorBranch: Partial<Branch> = {
        id: 'b1',
        slug: 'feature-auth',
        status: 'error',
        error_message: undefined,
      }
      mockFetch.mockResponse = errorBranch

      const { data, error } = await branching.waitForReady('feature-auth')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Branch creation failed')
    })
  })
})
