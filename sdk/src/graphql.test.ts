/**
 * GraphQL Module Tests
 */

import { describe, it, expect, beforeEach } from 'vitest'
import { FluxbaseGraphQL, type GraphQLResponse, type GraphQLError } from './graphql'
import type { FluxbaseFetch } from './fetch'

// Mock FluxbaseFetch
class MockFetch implements Partial<FluxbaseFetch> {
  public lastUrl: string = ''
  public lastMethod: string = ''
  public lastBody: unknown = null
  public lastOptions: { headers?: Record<string, string> } = {}
  public mockResponse: any = null
  public shouldThrow: boolean = false
  public errorMessage: string = 'Test error'

  async post<T>(
    path: string,
    body?: unknown,
    options?: { headers?: Record<string, string> }
  ): Promise<T> {
    this.lastUrl = path
    this.lastMethod = 'POST'
    this.lastBody = body
    this.lastOptions = options || {}
    if (this.shouldThrow) {
      throw new Error(this.errorMessage)
    }
    return this.mockResponse as T
  }
}

describe('FluxbaseGraphQL', () => {
  let mockFetch: MockFetch
  let graphql: FluxbaseGraphQL

  beforeEach(() => {
    mockFetch = new MockFetch()
    graphql = new FluxbaseGraphQL(mockFetch as unknown as FluxbaseFetch)
  })

  describe('query', () => {
    it('should execute a simple query', async () => {
      const mockResponse: GraphQLResponse<{ users: { id: string; email: string }[] }> = {
        data: {
          users: [
            { id: '1', email: 'user1@example.com' },
            { id: '2', email: 'user2@example.com' },
          ],
        },
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`
        query {
          users { id email }
        }
      `)

      expect(mockFetch.lastUrl).toBe('/api/v1/graphql')
      expect(mockFetch.lastMethod).toBe('POST')
      expect(mockFetch.lastBody).toMatchObject({
        query: expect.stringContaining('users { id email }'),
      })
      expect(result).toEqual(mockResponse)
    })

    it('should execute a query with variables', async () => {
      const mockResponse: GraphQLResponse<{ user: { id: string } }> = {
        data: { user: { id: '123' } },
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(
        `query GetUser($id: ID!) { user(id: $id) { id } }`,
        { id: '123' }
      )

      expect(mockFetch.lastBody).toMatchObject({
        query: expect.stringContaining('GetUser'),
        variables: { id: '123' },
      })
      expect(result.data).toEqual({ user: { id: '123' } })
    })

    it('should not include empty variables', async () => {
      mockFetch.mockResponse = { data: {} }

      await graphql.query(`query { users { id } }`, {})

      expect(mockFetch.lastBody).not.toHaveProperty('variables')
    })

    it('should execute query with custom headers', async () => {
      mockFetch.mockResponse = { data: {} }

      await graphql.query(`query { users { id } }`, undefined, {
        headers: { 'X-Custom-Header': 'custom-value' },
      })

      expect(mockFetch.lastOptions.headers).toMatchObject({
        'Content-Type': 'application/json',
        'X-Custom-Header': 'custom-value',
      })
    })

    it('should return GraphQL errors', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          {
            message: 'Field "unknownField" not found on type "User"',
            locations: [{ line: 1, column: 15 }],
            path: ['users', 'unknownField'],
          },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`query { users { unknownField } }`)

      expect(result.errors).toBeDefined()
      expect(result.errors?.[0].message).toContain('unknownField')
    })

    it('should handle network errors', async () => {
      mockFetch.shouldThrow = true
      mockFetch.errorMessage = 'Network error'

      const result = await graphql.query(`query { users { id } }`)

      expect(result.errors).toBeDefined()
      expect(result.errors?.[0].message).toBe('Network error')
    })

    it('should handle non-Error exceptions', async () => {
      mockFetch.post = async () => {
        throw 'string error'
      }

      const result = await graphql.query(`query { users { id } }`)

      expect(result.errors).toBeDefined()
      expect(result.errors?.[0].message).toBe('GraphQL request failed')
    })
  })

  describe('mutation', () => {
    it('should execute a mutation', async () => {
      const mockResponse: GraphQLResponse<{ insertUser: { id: string; email: string } }> = {
        data: {
          insertUser: { id: 'new-id', email: 'new@example.com' },
        },
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.mutation(`
        mutation CreateUser($data: UserInput!) {
          insertUser(data: $data) { id email }
        }
      `, { data: { email: 'new@example.com' } })

      expect(mockFetch.lastBody).toMatchObject({
        query: expect.stringContaining('mutation CreateUser'),
        variables: { data: { email: 'new@example.com' } },
      })
      expect(result.data?.insertUser).toEqual({ id: 'new-id', email: 'new@example.com' })
    })

    it('should execute a mutation without variables', async () => {
      mockFetch.mockResponse = { data: { deleteAllUsers: { count: 5 } } }

      const result = await graphql.mutation(`
        mutation { deleteAllUsers { count } }
      `)

      expect(mockFetch.lastBody).not.toHaveProperty('variables')
      expect(result.data).toBeDefined()
    })

    it('should handle mutation errors', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          {
            message: 'Unique constraint violation',
            extensions: { code: 'CONSTRAINT_VIOLATION' },
          },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.mutation(`
        mutation { insertUser(data: {}) { id } }
      `)

      expect(result.errors?.[0].message).toBe('Unique constraint violation')
      expect(result.errors?.[0].extensions?.code).toBe('CONSTRAINT_VIOLATION')
    })
  })

  describe('execute', () => {
    it('should execute with operation name', async () => {
      mockFetch.mockResponse = { data: { user: { id: '1' } } }

      await graphql.execute(
        `
          query GetUser($id: ID!) { user(id: $id) { id } }
          query ListUsers { users { id } }
        `,
        { id: '1' },
        'GetUser'
      )

      expect(mockFetch.lastBody).toMatchObject({
        operationName: 'GetUser',
      })
    })

    it('should execute without operation name', async () => {
      mockFetch.mockResponse = { data: {} }

      await graphql.execute(`query { users { id } }`)

      expect(mockFetch.lastBody).not.toHaveProperty('operationName')
    })

    it('should handle null response', async () => {
      mockFetch.mockResponse = null

      const result = await graphql.execute(`query { users { id } }`)

      expect(result.errors).toBeDefined()
      expect(result.errors?.[0].message).toBe('No response received')
    })

    it('should pass timeout option', async () => {
      mockFetch.mockResponse = { data: {} }

      await graphql.execute(
        `query { users { id } }`,
        undefined,
        undefined,
        { timeout: 5000 }
      )

      // Note: timeout is passed to the fetch options
      // The mock doesn't track timeout, but the execution should complete
      expect(mockFetch.lastUrl).toBe('/api/v1/graphql')
    })
  })

  describe('introspect', () => {
    it('should fetch the schema via introspection', async () => {
      const mockIntrospectionResponse: GraphQLResponse<{ __schema: any }> = {
        data: {
          __schema: {
            queryType: { name: 'Query' },
            mutationType: { name: 'Mutation' },
            subscriptionType: null,
            types: [
              { kind: 'OBJECT', name: 'Query' },
              { kind: 'OBJECT', name: 'User' },
            ],
            directives: [
              { name: 'deprecated', locations: ['FIELD_DEFINITION'] },
            ],
          },
        },
      }
      mockFetch.mockResponse = mockIntrospectionResponse

      const result = await graphql.introspect()

      expect(mockFetch.lastBody).toMatchObject({
        query: expect.stringContaining('IntrospectionQuery'),
      })
      expect(result.data?.__schema.queryType.name).toBe('Query')
      expect(result.data?.__schema.types.length).toBeGreaterThan(0)
    })

    it('should pass options to introspection query', async () => {
      mockFetch.mockResponse = { data: { __schema: {} } }

      await graphql.introspect({
        headers: { 'X-Admin': 'true' },
      })

      expect(mockFetch.lastOptions.headers).toMatchObject({
        'X-Admin': 'true',
      })
    })

    it('should handle introspection errors', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          { message: 'Introspection disabled' },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.introspect()

      expect(result.errors?.[0].message).toBe('Introspection disabled')
    })
  })

  describe('type safety', () => {
    it('should support typed query responses', async () => {
      interface UsersQuery {
        users: Array<{ id: string; email: string }>
      }

      mockFetch.mockResponse = {
        data: {
          users: [{ id: '1', email: 'test@example.com' }],
        },
      }

      const result = await graphql.query<UsersQuery>(`query { users { id email } }`)

      // TypeScript should infer the type correctly
      expect(result.data?.users[0].id).toBe('1')
      expect(result.data?.users[0].email).toBe('test@example.com')
    })

    it('should support typed mutation responses', async () => {
      interface CreateUserMutation {
        insertUser: { id: string; email: string }
      }

      mockFetch.mockResponse = {
        data: {
          insertUser: { id: 'new', email: 'new@example.com' },
        },
      }

      const result = await graphql.mutation<CreateUserMutation>(
        `mutation { insertUser(data: {}) { id email } }`
      )

      expect(result.data?.insertUser.id).toBe('new')
    })
  })

  describe('error structures', () => {
    it('should preserve error locations', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          {
            message: 'Syntax error',
            locations: [
              { line: 2, column: 5 },
              { line: 3, column: 10 },
            ],
          },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`invalid query`)

      expect(result.errors?.[0].locations).toHaveLength(2)
      expect(result.errors?.[0].locations?.[0]).toEqual({ line: 2, column: 5 })
    })

    it('should preserve error path', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          {
            message: 'Cannot return null',
            path: ['users', 0, 'email'],
          },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`query { users { email } }`)

      expect(result.errors?.[0].path).toEqual(['users', 0, 'email'])
    })

    it('should preserve error extensions', async () => {
      const mockResponse: GraphQLResponse = {
        errors: [
          {
            message: 'Rate limited',
            extensions: {
              code: 'RATE_LIMITED',
              retryAfter: 60,
            },
          },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`query { users { id } }`)

      expect(result.errors?.[0].extensions?.code).toBe('RATE_LIMITED')
      expect(result.errors?.[0].extensions?.retryAfter).toBe(60)
    })
  })

  describe('edge cases', () => {
    it('should handle response with both data and errors (partial success)', async () => {
      const mockResponse: GraphQLResponse<{ users: { id: string }[] }> = {
        data: {
          users: [{ id: '1' }],
        },
        errors: [
          { message: 'Could not fetch all fields' },
        ],
      }
      mockFetch.mockResponse = mockResponse

      const result = await graphql.query(`query { users { id optionalField } }`)

      expect(result.data?.users).toHaveLength(1)
      expect(result.errors).toHaveLength(1)
    })

    it('should handle empty data response', async () => {
      mockFetch.mockResponse = { data: {} }

      const result = await graphql.query(`query { __typename }`)

      expect(result.data).toEqual({})
      expect(result.errors).toBeUndefined()
    })

    it('should handle complex nested queries', async () => {
      interface ComplexQuery {
        users: Array<{
          id: string
          posts: Array<{
            id: string
            comments: Array<{ id: string }>
          }>
        }>
      }

      mockFetch.mockResponse = {
        data: {
          users: [{
            id: '1',
            posts: [{
              id: 'p1',
              comments: [{ id: 'c1' }, { id: 'c2' }],
            }],
          }],
        },
      }

      const result = await graphql.query<ComplexQuery>(`
        query {
          users {
            id
            posts {
              id
              comments { id }
            }
          }
        }
      `)

      expect(result.data?.users[0].posts[0].comments).toHaveLength(2)
    })
  })
})
