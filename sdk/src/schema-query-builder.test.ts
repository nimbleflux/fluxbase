/**
 * Schema Query Builder Tests
 */

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { SchemaQueryBuilder } from './schema-query-builder'
import { QueryBuilder } from './query-builder'
import type { FluxbaseFetch } from './fetch'

// Mock FluxbaseFetch
class MockFetch implements Partial<FluxbaseFetch> {
  async get<T>(): Promise<T> {
    return {} as T
  }
  async post<T>(): Promise<T> {
    return {} as T
  }
}

describe('SchemaQueryBuilder', () => {
  let mockFetch: MockFetch
  let schemaBuilder: SchemaQueryBuilder

  beforeEach(() => {
    mockFetch = new MockFetch()
    schemaBuilder = new SchemaQueryBuilder(mockFetch as unknown as FluxbaseFetch, 'logging')
  })

  describe('constructor', () => {
    it('should create a schema query builder with fetch and schema name', () => {
      const builder = new SchemaQueryBuilder(mockFetch as unknown as FluxbaseFetch, 'analytics')
      expect(builder).toBeDefined()
    })
  })

  describe('from', () => {
    it('should return a QueryBuilder for a table in the schema', () => {
      const queryBuilder = schemaBuilder.from('entries')

      expect(queryBuilder).toBeInstanceOf(QueryBuilder)
    })

    it('should create different query builders for different tables', () => {
      const entriesBuilder = schemaBuilder.from('entries')
      const logsBuilder = schemaBuilder.from('logs')

      // Both should be QueryBuilder instances
      expect(entriesBuilder).toBeInstanceOf(QueryBuilder)
      expect(logsBuilder).toBeInstanceOf(QueryBuilder)

      // They should be different instances
      expect(entriesBuilder).not.toBe(logsBuilder)
    })

    it('should support generic type parameter', () => {
      interface LogEntry {
        id: string
        level: string
        message: string
      }

      const builder = schemaBuilder.from<LogEntry>('entries')

      expect(builder).toBeInstanceOf(QueryBuilder)
    })

    it('should work with various schema names', () => {
      const schemas = ['public', 'auth', 'storage', 'jobs', 'functions', 'custom_schema']

      for (const schema of schemas) {
        const builder = new SchemaQueryBuilder(mockFetch as unknown as FluxbaseFetch, schema)
        const queryBuilder = builder.from('table')
        expect(queryBuilder).toBeInstanceOf(QueryBuilder)
      }
    })
  })
})
