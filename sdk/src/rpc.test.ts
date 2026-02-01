/**
 * RPC Module Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { FluxbaseRPC, type RPCInvokeOptions } from './rpc'
import type { RPCProcedureSummary, RPCExecution, RPCInvokeResponse, RPCExecutionLog } from './types'

// Mock fetch interface
class MockFetch {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public lastOptions: { timeout?: number } = {}
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

  async post<T>(path: string, body?: unknown, options?: { timeout?: number }): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'POST'
    this.lastBody = body
    this.lastOptions = options || {}
    this.callCount++
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }
}

describe('FluxbaseRPC', () => {
  let mockFetch: MockFetch
  let rpc: FluxbaseRPC

  beforeEach(() => {
    mockFetch = new MockFetch()
    rpc = new FluxbaseRPC(mockFetch)
  })

  describe('list', () => {
    it('should list available procedures', async () => {
      const mockProcedures: RPCProcedureSummary[] = [
        {
          name: 'get_user_orders',
          namespace: 'default',
          description: 'Get orders for a user',
        },
        {
          name: 'calculate_total',
          namespace: 'default',
          description: 'Calculate order total',
        },
      ]
      mockFetch.mockResponse = { procedures: mockProcedures, count: 2 }

      const { data, error } = await rpc.list()

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/procedures')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockProcedures)
      expect(error).toBeNull()
    })

    it('should filter by namespace', async () => {
      mockFetch.mockResponse = { procedures: [], count: 0 }

      await rpc.list('custom')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/procedures?namespace=custom')
    })

    it('should encode namespace in URL', async () => {
      mockFetch.mockResponse = { procedures: [], count: 0 }

      await rpc.list('my app')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/procedures?namespace=my%20app')
    })

    it('should return empty array when no procedures', async () => {
      mockFetch.mockResponse = { count: 0 }

      const { data, error } = await rpc.list()

      expect(data).toEqual([])
      expect(error).toBeNull()
    })

    it('should handle list errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Permission denied'

      const { data, error } = await rpc.list()

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Permission denied')
    })
  })

  describe('invoke', () => {
    it('should invoke a procedure synchronously', async () => {
      const mockResponse: RPCInvokeResponse<{ orders: any[] }> = {
        result: { orders: [{ id: '1' }, { id: '2' }] },
        execution_id: 'exec-123',
        status: 'completed',
      }
      mockFetch.mockResponse = mockResponse

      const { data, error } = await rpc.invoke('get_user_orders', {
        user_id: '123',
        limit: 10,
      })

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/default/get_user_orders')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({
        params: { user_id: '123', limit: 10 },
        async: undefined,
      })
      expect(data).toEqual(mockResponse)
      expect(error).toBeNull()
    })

    it('should invoke with custom namespace', async () => {
      const mockResponse: RPCInvokeResponse = { status: 'completed' }
      mockFetch.mockResponse = mockResponse

      await rpc.invoke('my_procedure', {}, { namespace: 'custom' })

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/custom/my_procedure')
    })

    it('should invoke asynchronously', async () => {
      const mockResponse: RPCInvokeResponse = {
        execution_id: 'async-exec-123',
        status: 'running',
      }
      mockFetch.mockResponse = mockResponse

      const { data, error } = await rpc.invoke('long_running_report', {
        year: 2024,
      }, { async: true })

      expect(mockFetch.lastBody).toMatchObject({
        async: true,
      })
      expect(data?.execution_id).toBe('async-exec-123')
      expect(error).toBeNull()
    })

    it('should use custom timeout', async () => {
      mockFetch.mockResponse = { status: 'completed' }

      await rpc.invoke('quick_proc', {}, { timeout: 5000 })

      expect(mockFetch.lastOptions.timeout).toBe(5000)
    })

    it('should encode procedure name in URL', async () => {
      mockFetch.mockResponse = { status: 'completed' }

      await rpc.invoke('my procedure', {})

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/default/my%20procedure')
    })

    it('should encode namespace in URL', async () => {
      mockFetch.mockResponse = { status: 'completed' }

      await rpc.invoke('proc', {}, { namespace: 'my namespace' })

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/my%20namespace/proc')
    })

    it('should invoke without params', async () => {
      mockFetch.mockResponse = { status: 'completed' }

      await rpc.invoke('no_params_proc')

      expect(mockFetch.lastBody).toEqual({
        params: undefined,
        async: undefined,
      })
    })

    it('should handle invoke errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Procedure not found'

      const { data, error } = await rpc.invoke('non_existent', {})

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Procedure not found')
    })
  })

  describe('getStatus', () => {
    it('should get execution status', async () => {
      const mockExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'completed',
        result: { data: 'test' },
        started_at: '2025-01-01T00:00:00Z',
        completed_at: '2025-01-01T00:00:05Z',
      }
      mockFetch.mockResponse = mockExecution

      const { data, error } = await rpc.getStatus('exec-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec-123')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockExecution)
      expect(error).toBeNull()
    })

    it('should encode execution ID in URL', async () => {
      mockFetch.mockResponse = { status: 'completed' }

      await rpc.getStatus('exec 123')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec%20123')
    })

    it('should handle getStatus errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Execution not found'

      const { data, error } = await rpc.getStatus('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Execution not found')
    })
  })

  describe('getLogs', () => {
    it('should get execution logs', async () => {
      const mockLogs: RPCExecutionLog[] = [
        {
          id: 'log-1',
          execution_id: 'exec-123',
          line_number: 1,
          level: 'info',
          message: 'Starting procedure',
          timestamp: '2025-01-01T00:00:00Z',
        },
        {
          id: 'log-2',
          execution_id: 'exec-123',
          line_number: 2,
          level: 'debug',
          message: 'Processing data',
          timestamp: '2025-01-01T00:00:01Z',
        },
      ]
      mockFetch.mockResponse = { logs: mockLogs, count: 2 }

      const { data, error } = await rpc.getLogs('exec-123')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec-123/logs')
      expect(mockFetch.lastMethod).toBe('GET')
      expect(data).toEqual(mockLogs)
      expect(error).toBeNull()
    })

    it('should get logs after specific line', async () => {
      mockFetch.mockResponse = { logs: [], count: 0 }

      await rpc.getLogs('exec-123', 5)

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec-123/logs?after=5')
    })

    it('should handle afterLine 0', async () => {
      mockFetch.mockResponse = { logs: [], count: 0 }

      await rpc.getLogs('exec-123', 0)

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec-123/logs?after=0')
    })

    it('should encode execution ID in URL', async () => {
      mockFetch.mockResponse = { logs: [], count: 0 }

      await rpc.getLogs('exec 123')

      expect(mockFetch.lastUrl).toBe('/api/v1/rpc/executions/exec%20123/logs')
    })

    it('should return empty array when no logs', async () => {
      mockFetch.mockResponse = { count: 0 }

      const { data, error } = await rpc.getLogs('exec-123')

      expect(data).toEqual([])
      expect(error).toBeNull()
    })

    it('should handle getLogs errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Execution not found'

      const { data, error } = await rpc.getLogs('non-existent')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Execution not found')
    })
  })

  describe('waitForCompletion', () => {
    it('should wait for completion and return result', async () => {
      const completedExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'completed',
        result: { data: 'success' },
        started_at: '2025-01-01T00:00:00Z',
        completed_at: '2025-01-01T00:00:05Z',
      }
      mockFetch.mockResponse = completedExecution

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toEqual(completedExecution)
      expect(error).toBeNull()
    })

    it('should return on failed status', async () => {
      const failedExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'failed',
        error: 'Something went wrong',
        started_at: '2025-01-01T00:00:00Z',
        completed_at: '2025-01-01T00:00:05Z',
      }
      mockFetch.mockResponse = failedExecution

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toEqual(failedExecution)
      expect(error).toBeNull()
    })

    it('should return on cancelled status', async () => {
      const cancelledExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'cancelled',
        started_at: '2025-01-01T00:00:00Z',
      }
      mockFetch.mockResponse = cancelledExecution

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toEqual(cancelledExecution)
      expect(error).toBeNull()
    })

    it('should return on timeout status', async () => {
      const timeoutExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'timeout',
        started_at: '2025-01-01T00:00:00Z',
      }
      mockFetch.mockResponse = timeoutExecution

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toEqual(timeoutExecution)
      expect(error).toBeNull()
    })

    it('should poll until completed', async () => {
      const runningExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'running',
        started_at: '2025-01-01T00:00:00Z',
      }
      const completedExecution: RPCExecution = {
        ...runningExecution,
        status: 'completed',
        result: { data: 'done' },
        completed_at: '2025-01-01T00:00:05Z',
      }

      // First call returns running, second returns completed
      mockFetch.responseQueue = [runningExecution, completedExecution]

      const { data, error } = await rpc.waitForCompletion('exec-123', {
        initialIntervalMs: 10,
        maxWaitMs: 5000,
      })

      expect(data).toEqual(completedExecution)
      expect(error).toBeNull()
      expect(mockFetch.callCount).toBe(2)
    })

    it('should call onProgress callback', async () => {
      const runningExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'running',
        started_at: '2025-01-01T00:00:00Z',
      }
      const completedExecution: RPCExecution = {
        ...runningExecution,
        status: 'completed',
        completed_at: '2025-01-01T00:00:05Z',
      }

      mockFetch.responseQueue = [runningExecution, completedExecution]
      const progressCalls: RPCExecution[] = []

      await rpc.waitForCompletion('exec-123', {
        initialIntervalMs: 10,
        maxWaitMs: 5000,
        onProgress: (exec) => progressCalls.push(exec),
      })

      expect(progressCalls.length).toBe(2)
      expect(progressCalls[0].status).toBe('running')
      expect(progressCalls[1].status).toBe('completed')
    })

    it('should timeout after maxWaitMs', async () => {
      const runningExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'running',
        started_at: '2025-01-01T00:00:00Z',
      }
      mockFetch.mockResponse = runningExecution

      const { data, error } = await rpc.waitForCompletion('exec-123', {
        maxWaitMs: 50,
        initialIntervalMs: 20,
      })

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toContain('Timeout waiting for execution')
    })

    it('should return error if getStatus fails', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Network error'

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Network error')
    })

    it('should return error if execution not found', async () => {
      mockFetch.mockResponse = null

      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Execution not found')
    })

    it('should use default options', async () => {
      const completedExecution: RPCExecution = {
        id: 'exec-123',
        procedure_name: 'my_proc',
        namespace: 'default',
        status: 'completed',
        started_at: '2025-01-01T00:00:00Z',
      }
      mockFetch.mockResponse = completedExecution

      // Should work with no options (uses defaults)
      const { data, error } = await rpc.waitForCompletion('exec-123')

      expect(data).toEqual(completedExecution)
      expect(error).toBeNull()
    })
  })
})
