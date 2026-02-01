/**
 * Tests for GraphQL hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useGraphQLQuery,
  useGraphQLMutation,
  useGraphQLIntrospection,
  useGraphQL,
} from './use-graphql';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useGraphQLQuery', () => {
  it('should execute query and return data', async () => {
    const mockData = { users: [{ id: 1, name: 'Test' }] };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLQuery('users', 'query { users { id name } }'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockData);
    expect(executeMock).toHaveBeenCalledWith(
      'query { users { id name } }',
      undefined,
      undefined,
      undefined
    );
  });

  it('should pass variables to query', async () => {
    const mockData = { user: { id: 1 } };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLQuery(
        ['user', '1'],
        'query GetUser($id: ID!) { user(id: $id) { id } }',
        { variables: { id: '1' } }
      ),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(executeMock).toHaveBeenCalledWith(
      'query GetUser($id: ID!) { user(id: $id) { id } }',
      { id: '1' },
      undefined,
      undefined
    );
  });

  it('should throw error on GraphQL errors', async () => {
    const graphqlError = { message: 'Query failed' };
    const executeMock = vi.fn().mockResolvedValue({ data: null, errors: [graphqlError] });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLQuery('users', 'query { users { id } }'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(graphqlError);
  });

  it('should not execute when disabled', async () => {
    const executeMock = vi.fn();

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLQuery('users', 'query { users { id } }', { enabled: false }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(executeMock).not.toHaveBeenCalled();
  });

  it('should normalize string query key to array', async () => {
    const executeMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(
      () => useGraphQLQuery('users', 'query { users { id } }'),
      { wrapper: createWrapper(client, queryClient) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    // Check that data is stored with normalized key
    expect(queryClient.getQueryData(['fluxbase', 'graphql', 'users'])).toEqual({});
  });

  it('should pass operation name', async () => {
    const executeMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    renderHook(
      () => useGraphQLQuery('users', 'query GetUsers { users { id } }', { operationName: 'GetUsers' }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => {
      expect(executeMock).toHaveBeenCalledWith(
        'query GetUsers { users { id } }',
        undefined,
        'GetUsers',
        undefined
      );
    });
  });

  it('should apply select transform', async () => {
    const mockData = { users: [{ id: 1 }, { id: 2 }] };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLQuery<{ users: { id: number }[] }>(
        'users',
        'query { users { id } }',
        { select: (data) => data ? { users: data.users.filter(u => u.id === 1) } : undefined }
      ),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data?.users).toHaveLength(1);
  });
});

describe('useGraphQLMutation', () => {
  it('should execute mutation and return data', async () => {
    const mockData = { insertUser: { id: 1, email: 'test@example.com' } };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLMutation('mutation CreateUser($email: String!) { insertUser(email: $email) { id email } }'),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({ email: 'test@example.com' });
    });

    expect(executeMock).toHaveBeenCalledWith(
      'mutation CreateUser($email: String!) { insertUser(email: $email) { id email } }',
      { email: 'test@example.com' },
      undefined,
      undefined
    );
  });

  it('should throw error on GraphQL errors', async () => {
    const graphqlError = { message: 'Mutation failed' };
    const executeMock = vi.fn().mockResolvedValue({ data: null, errors: [graphqlError] });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLMutation('mutation { deleteUser(id: "1") }'),
      { wrapper: createWrapper(client) }
    );

    await expect(act(async () => {
      await result.current.mutateAsync({});
    })).rejects.toEqual(graphqlError);
  });

  it('should invalidate queries on success', async () => {
    const executeMock = vi.fn().mockResolvedValue({ data: { insertUser: { id: 1 } }, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(
      () => useGraphQLMutation('mutation { insertUser { id } }', { invalidateQueries: ['users', 'stats'] }),
      { wrapper: createWrapper(client, queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({});
    });

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'graphql', 'users'] });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'graphql', 'stats'] });
  });

  it('should call onSuccess callback', async () => {
    const mockData = { insertUser: { id: 1 } };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, errors: null });
    const onSuccess = vi.fn();

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLMutation('mutation { insertUser { id } }', { onSuccess }),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({ name: 'test' });
    });

    expect(onSuccess).toHaveBeenCalledWith(mockData, { name: 'test' });
  });

  it('should call onError callback', async () => {
    const graphqlError = { message: 'Mutation failed' };
    const executeMock = vi.fn().mockResolvedValue({ data: null, errors: [graphqlError] });
    const onError = vi.fn();

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLMutation('mutation { deleteUser }', { onError }),
      { wrapper: createWrapper(client) }
    );

    try {
      await act(async () => {
        await result.current.mutateAsync({ id: '1' });
      });
    } catch {
      // Expected error
    }

    expect(onError).toHaveBeenCalledWith(graphqlError, { id: '1' });
  });
});

describe('useGraphQLIntrospection', () => {
  it('should fetch schema introspection', async () => {
    const mockSchema = { __schema: { types: [], queryType: { name: 'Query' } } };
    const introspectMock = vi.fn().mockResolvedValue({ data: mockSchema, errors: null });

    const client = createMockClient({
      graphql: { introspect: introspectMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLIntrospection(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockSchema);
  });

  it('should throw error on introspection errors', async () => {
    const graphqlError = { message: 'Introspection disabled' };
    const introspectMock = vi.fn().mockResolvedValue({ data: null, errors: [graphqlError] });

    const client = createMockClient({
      graphql: { introspect: introspectMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLIntrospection(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toEqual(graphqlError);
  });

  it('should not fetch when disabled', async () => {
    const introspectMock = vi.fn();

    const client = createMockClient({
      graphql: { introspect: introspectMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQLIntrospection({ enabled: false }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(introspectMock).not.toHaveBeenCalled();
  });
});

describe('useGraphQL', () => {
  it('should return executeQuery function', () => {
    const queryMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { query: queryMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.executeQuery).toBeDefined();
  });

  it('should return executeMutation function', () => {
    const mutationMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { mutation: mutationMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.executeMutation).toBeDefined();
  });

  it('should return execute function', () => {
    const executeMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { execute: executeMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.execute).toBeDefined();
  });

  it('should return introspect function', () => {
    const introspectMock = vi.fn().mockResolvedValue({ data: {}, errors: null });

    const client = createMockClient({
      graphql: { introspect: introspectMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.introspect).toBeDefined();
  });

  it('should execute query via executeQuery', async () => {
    const queryMock = vi.fn().mockResolvedValue({ data: { users: [] }, errors: null });

    const client = createMockClient({
      graphql: { query: queryMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.executeQuery('query { users { id } }', { limit: 10 });
    });

    expect(queryMock).toHaveBeenCalledWith('query { users { id } }', { limit: 10 }, undefined);
  });

  it('should execute mutation via executeMutation', async () => {
    const mutationMock = vi.fn().mockResolvedValue({ data: { insertUser: {} }, errors: null });

    const client = createMockClient({
      graphql: { mutation: mutationMock },
    } as any);

    const { result } = renderHook(
      () => useGraphQL(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.executeMutation('mutation { insertUser { id } }', { email: 'test@example.com' });
    });

    expect(mutationMock).toHaveBeenCalledWith('mutation { insertUser { id } }', { email: 'test@example.com' }, undefined);
  });
});
