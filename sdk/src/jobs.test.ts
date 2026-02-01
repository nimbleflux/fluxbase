/**
 * Jobs Module Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { FluxbaseJobs } from './jobs'
import type { FluxbaseFetch } from './fetch'
import type { Job, ExecutionLog } from './types'

// Mock FluxbaseFetch
class MockFetch implements Partial<FluxbaseFetch> {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public mockResponse: any = null
  public shouldThrow: boolean = false
  public errorMessage: string = 'Test error'

  async get<T>(path: string): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'GET'
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }

  async post<T>(path: string, body?: unknown): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'POST'
    this.lastBody = body
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }
}

describe('FluxbaseJobs', () => {
  let mockFetch: MockFetch
  let jobs: FluxbaseJobs

  beforeEach(() => {
    mockFetch = new MockFetch()
    jobs = new FluxbaseJobs(mockFetch as unknown as FluxbaseFetch)
  })

  describe('submit', () => {
    it('should submit a job with name and payload', async () => {
      const mockJob: Partial<Job> = {
        id: 'job-123',
        job_name: 'send-email',
        status: 'pending',
      }
      mockFetch.mockResponse = mockJob

      const { data, error } = await jobs.submit('send-email', {
        to: 'user@example.com',
        subject: 'Hello',
      })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/submit')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({
        job_name: 'send-email',
        payload: { to: 'user@example.com', subject: 'Hello' },
        priority: undefined,
        namespace: undefined,
        scheduled: undefined,
        on_behalf_of: undefined,
      })
      expect(data).toEqual(mockJob)
      expect(error).toBeNull()
    })

    it('should submit a job with priority', async () => {
      const mockJob: Partial<Job> = { id: 'job-123', status: 'pending' }
      mockFetch.mockResponse = mockJob

      await jobs.submit('high-priority', {}, { priority: 10 })

      expect(mockFetch.lastBody).toMatchObject({
        priority: 10,
      })
    })

    it('should submit a job with namespace', async () => {
      const mockJob: Partial<Job> = { id: 'job-123', status: 'pending' }
      mockFetch.mockResponse = mockJob

      await jobs.submit('namespaced-job', {}, { namespace: 'my-app' })

      expect(mockFetch.lastBody).toMatchObject({
        namespace: 'my-app',
      })
    })

    it('should submit a scheduled job', async () => {
      const mockJob: Partial<Job> = { id: 'job-123', status: 'pending' }
      mockFetch.mockResponse = mockJob

      const scheduledTime = '2025-01-01T00:00:00Z'
      await jobs.submit('scheduled-task', {}, { scheduled: scheduledTime })

      expect(mockFetch.lastBody).toMatchObject({
        scheduled: scheduledTime,
      })
    })

    it('should submit a job on behalf of another user', async () => {
      const mockJob: Partial<Job> = { id: 'job-123', status: 'pending' }
      mockFetch.mockResponse = mockJob

      const onBehalfOf = {
        user_id: 'user-uuid',
        user_email: 'user@example.com',
      }
      await jobs.submit('user-task', {}, { onBehalfOf })

      expect(mockFetch.lastBody).toMatchObject({
        on_behalf_of: onBehalfOf,
      })
    })

    it('should submit a job without payload', async () => {
      const mockJob: Partial<Job> = { id: 'job-123', status: 'pending' }
      mockFetch.mockResponse = mockJob

      await jobs.submit('simple-job')

      expect(mockFetch.lastBody).toMatchObject({
        job_name: 'simple-job',
        payload: undefined,
      })
    })

    it('should handle submit errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Job queue full'

      const { data, error } = await jobs.submit('failed-job', {})

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Job queue full')
    })
  })

  describe('get', () => {
    it('should get job status by ID', async () => {
      const mockJob: Partial<Job> = {
        id: 'job-123',
        job_name: 'test-job',
        status: 'running',
        progress_percent: 50,
      }
      mockFetch.mockResponse = mockJob

      const { data, error } = await jobs.get('job-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/job-123')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockJob)
      expect(error).toBeNull()
    })

    it('should handle get errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Job not found'

      const { data, error } = await jobs.get('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Job not found')
    })
  })

  describe('list', () => {
    it('should list all jobs', async () => {
      const mockJobs: Partial<Job>[] = [
        { id: 'job-1', status: 'completed' },
        { id: 'job-2', status: 'running' },
      ]
      mockFetch.mockResponse = mockJobs

      const { data, error } = await jobs.list()

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockJobs)
      expect(error).toBeNull()
    })

    it('should filter jobs by status', async () => {
      mockFetch.mockResponse = []

      await jobs.list({ status: 'running' })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs?status=running')
    })

    it('should filter jobs by namespace', async () => {
      mockFetch.mockResponse = []

      await jobs.list({ namespace: 'my-app' })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs?namespace=my-app')
    })

    it('should support pagination with limit', async () => {
      mockFetch.mockResponse = []

      await jobs.list({ limit: 20 })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs?limit=20')
    })

    it('should support pagination with offset', async () => {
      mockFetch.mockResponse = []

      await jobs.list({ offset: 40 })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs?offset=40')
    })

    it('should support includeResult option', async () => {
      mockFetch.mockResponse = []

      await jobs.list({ includeResult: true })

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs?include_result=true')
    })

    it('should combine multiple filters', async () => {
      mockFetch.mockResponse = []

      await jobs.list({
        status: 'completed',
        namespace: 'my-app',
        limit: 10,
        offset: 20,
      })

      expect(mockFetch.lastUrl).toContain('status=completed')
      expect(mockFetch.lastUrl).toContain('namespace=my-app')
      expect(mockFetch.lastUrl).toContain('limit=10')
      expect(mockFetch.lastUrl).toContain('offset=20')
    })

    it('should handle list errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Permission denied'

      const { data, error } = await jobs.list()

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Permission denied')
    })
  })

  describe('cancel', () => {
    it('should cancel a job', async () => {
      mockFetch.mockResponse = null

      const { data, error } = await jobs.cancel('job-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/job-123/cancel')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({})
      expect(data).toBeNull()
      expect(error).toBeNull()
    })

    it('should handle cancel errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Cannot cancel completed job'

      const { data, error } = await jobs.cancel('completed-job')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Cannot cancel completed job')
    })
  })

  describe('retry', () => {
    it('should retry a failed job', async () => {
      const newJob: Partial<Job> = {
        id: 'new-job-456',
        status: 'pending',
      }
      mockFetch.mockResponse = newJob

      const { data, error } = await jobs.retry('failed-job-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/failed-job-123/retry')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({})
      expect(data).toEqual(newJob)
      expect(error).toBeNull()
    })

    it('should handle retry errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Cannot retry running job'

      const { data, error } = await jobs.retry('running-job')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Cannot retry running job')
    })
  })

  describe('getLogs', () => {
    it('should get logs for a job', async () => {
      const mockLogs: ExecutionLog[] = [
        {
          id: 'log-1',
          execution_id: 'job-123',
          line_number: 1,
          level: 'info',
          message: 'Job started',
          timestamp: '2025-01-01T00:00:00Z',
        },
        {
          id: 'log-2',
          execution_id: 'job-123',
          line_number: 2,
          level: 'info',
          message: 'Processing...',
          timestamp: '2025-01-01T00:00:01Z',
        },
      ]
      mockFetch.mockResponse = { logs: mockLogs, count: 2 }

      const { data, error } = await jobs.getLogs('job-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/job-123/logs')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockLogs)
      expect(error).toBeNull()
    })

    it('should get logs after a specific line', async () => {
      mockFetch.mockResponse = { logs: [], count: 0 }

      await jobs.getLogs('job-123', 10)

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/job-123/logs?after_line=10')
    })

    it('should handle getLogs with afterLine 0', async () => {
      mockFetch.mockResponse = { logs: [], count: 0 }

      await jobs.getLogs('job-123', 0)

      expect(mockFetch.lastUrl).toBe('/api/v1/jobs/job-123/logs?after_line=0')
    })

    it('should return empty array when no logs', async () => {
      mockFetch.mockResponse = { count: 0 }

      const { data, error } = await jobs.getLogs('job-123')

      expect(data).toEqual([])
      expect(error).toBeNull()
    })

    it('should handle getLogs errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Job not found'

      const { data, error } = await jobs.getLogs('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Job not found')
    })
  })
})
