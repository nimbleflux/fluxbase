/**
 * Vector Search Module Tests
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { FluxbaseVector } from './vector'
import type { FluxbaseFetch } from './fetch'
import type { EmbedResponse, VectorSearchResult } from './types'

// Mock FluxbaseFetch
class MockFetch implements Partial<FluxbaseFetch> {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public lastOptions: Record<string, unknown> = {}
  public mockResponse: any = null
  public shouldThrow: boolean = false
  public errorMessage: string = 'Test error'

  async request<T>(path: string, options: { method: string; body?: unknown }): Promise<T> {
    this.lastUrl = path
    this.lastMethod = options.method
    this.lastBody = options.body
    this.lastOptions = options
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }
}

describe('FluxbaseVector', () => {
  let mockFetch: MockFetch
  let vector: FluxbaseVector

  beforeEach(() => {
    mockFetch = new MockFetch()
    vector = new FluxbaseVector(mockFetch as unknown as FluxbaseFetch)
  })

  describe('embed', () => {
    it('should embed single text', async () => {
      const mockEmbedResponse: EmbedResponse = {
        embeddings: [[0.1, 0.2, 0.3, 0.4, 0.5]],
        model: 'text-embedding-3-small',
        usage: { prompt_tokens: 5, total_tokens: 5 },
      }
      mockFetch.mockResponse = mockEmbedResponse

      const { data, error } = await vector.embed({
        text: 'Hello world',
      })

      expect(mockFetch.lastUrl).toBe('/api/v1/vector/embed')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({
        text: 'Hello world',
      })
      expect(data).toEqual(mockEmbedResponse)
      expect(error).toBeNull()
    })

    it('should embed multiple texts', async () => {
      const mockEmbedResponse: EmbedResponse = {
        embeddings: [
          [0.1, 0.2, 0.3],
          [0.4, 0.5, 0.6],
        ],
        model: 'text-embedding-3-small',
        usage: { prompt_tokens: 10, total_tokens: 10 },
      }
      mockFetch.mockResponse = mockEmbedResponse

      const { data, error } = await vector.embed({
        texts: ['Hello', 'World'],
      })

      expect(mockFetch.lastBody).toEqual({
        texts: ['Hello', 'World'],
      })
      expect(data?.embeddings.length).toBe(2)
      expect(error).toBeNull()
    })

    it('should embed with custom model', async () => {
      const mockEmbedResponse: EmbedResponse = {
        embeddings: [[0.1, 0.2]],
        model: 'text-embedding-ada-002',
        usage: { prompt_tokens: 3, total_tokens: 3 },
      }
      mockFetch.mockResponse = mockEmbedResponse

      const { data, error } = await vector.embed({
        text: 'Test',
        model: 'text-embedding-ada-002',
      })

      expect(mockFetch.lastBody).toEqual({
        text: 'Test',
        model: 'text-embedding-ada-002',
      })
      expect(data?.model).toBe('text-embedding-ada-002')
      expect(error).toBeNull()
    })

    it('should handle embed errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Embedding service unavailable'

      const { data, error } = await vector.embed({
        text: 'Hello',
      })

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Embedding service unavailable')
    })
  })

  describe('search', () => {
    it('should search with text query', async () => {
      const mockSearchResult: VectorSearchResult<{ id: string; content: string }> = {
        results: [
          { id: '1', content: 'TypeScript guide', _distance: 0.1 },
          { id: '2', content: 'JavaScript basics', _distance: 0.2 },
        ],
        count: 2,
      }
      mockFetch.mockResponse = mockSearchResult

      const { data, error } = await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'How to use TypeScript?',
        match_count: 10,
      })

      expect(mockFetch.lastUrl).toBe('/api/v1/vector/search')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toEqual({
        table: 'documents',
        column: 'embedding',
        query: 'How to use TypeScript?',
        vector: undefined,
        metric: 'cosine',
        match_threshold: undefined,
        match_count: 10,
        select: undefined,
        filters: undefined,
      })
      expect(data).toEqual(mockSearchResult)
      expect(error).toBeNull()
    })

    it('should search with pre-computed vector', async () => {
      const mockSearchResult: VectorSearchResult = {
        results: [{ id: '1', _distance: 0.05 }],
        count: 1,
      }
      mockFetch.mockResponse = mockSearchResult

      const testVector = [0.1, 0.2, 0.3, 0.4, 0.5]
      const { data, error } = await vector.search({
        table: 'documents',
        column: 'embedding',
        vector: testVector,
        match_count: 5,
      })

      expect(mockFetch.lastBody).toMatchObject({
        vector: testVector,
      })
      expect(data).toEqual(mockSearchResult)
      expect(error).toBeNull()
    })

    it('should search with cosine metric', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        metric: 'cosine',
        match_count: 10,
      })

      expect(mockFetch.lastBody).toMatchObject({
        metric: 'cosine',
      })
    })

    it('should search with euclidean metric', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        metric: 'euclidean',
        match_count: 10,
      })

      expect(mockFetch.lastBody).toMatchObject({
        metric: 'euclidean',
      })
    })

    it('should search with inner_product metric', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        metric: 'inner_product',
        match_count: 10,
      })

      expect(mockFetch.lastBody).toMatchObject({
        metric: 'inner_product',
      })
    })

    it('should default to cosine metric', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        match_count: 10,
      })

      expect(mockFetch.lastBody).toMatchObject({
        metric: 'cosine',
      })
    })

    it('should search with match threshold', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        match_count: 10,
        match_threshold: 0.8,
      })

      expect(mockFetch.lastBody).toMatchObject({
        match_threshold: 0.8,
      })
    })

    it('should search with select columns', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        match_count: 10,
        select: ['id', 'title', 'content'],
      })

      expect(mockFetch.lastBody).toMatchObject({
        select: ['id', 'title', 'content'],
      })
    })

    it('should search with filters', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      const filters = [
        { column: 'status', operator: 'eq', value: 'published' },
        { column: 'category', operator: 'in', value: ['tech', 'tutorial'] },
      ]

      await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        match_count: 10,
        filters,
      })

      expect(mockFetch.lastBody).toMatchObject({
        filters,
      })
    })

    it('should search with all options', async () => {
      const mockSearchResult: VectorSearchResult<{ id: string; title: string }> = {
        results: [{ id: '1', title: 'Result', _distance: 0.1 }],
        count: 1,
      }
      mockFetch.mockResponse = mockSearchResult

      const { data, error } = await vector.search<{ id: string; title: string }>({
        table: 'documents',
        column: 'embedding',
        query: 'test query',
        metric: 'cosine',
        match_threshold: 0.7,
        match_count: 20,
        select: ['id', 'title'],
        filters: [
          { column: 'active', operator: 'eq', value: true },
        ],
      })

      expect(mockFetch.lastBody).toEqual({
        table: 'documents',
        column: 'embedding',
        query: 'test query',
        vector: undefined,
        metric: 'cosine',
        match_threshold: 0.7,
        match_count: 20,
        select: ['id', 'title'],
        filters: [
          { column: 'active', operator: 'eq', value: true },
        ],
      })
      expect(data).toEqual(mockSearchResult)
      expect(error).toBeNull()
    })

    it('should handle search errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Vector column not found'

      const { data, error } = await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'test',
        match_count: 10,
      })

      expect(data).toBeNull()
      expect(error).toBeDefined()
      expect(error?.message).toBe('Vector column not found')
    })

    it('should handle empty results', async () => {
      mockFetch.mockResponse = { results: [], count: 0 }

      const { data, error } = await vector.search({
        table: 'documents',
        column: 'embedding',
        query: 'very specific query with no matches',
        match_count: 10,
      })

      expect(data?.results).toEqual([])
      expect(data?.count).toBe(0)
      expect(error).toBeNull()
    })
  })
})
