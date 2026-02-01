/**
 * Tests for database query hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useFluxbaseQuery,
  useTable,
  useInsert,
  useUpdate,
  useUpsert,
  useDelete,
} from './use-query';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useFluxbaseQuery', () => {
  it('should execute query and return data', async () => {
    const mockData = [{ id: 1, name: 'Test' }];
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useFluxbaseQuery((client) => client.from('products').select('*'), { queryKey: ['products'] }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockData);
  });

  it('should throw error when query fails', async () => {
    const error = new Error('Query failed');
    const executeMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useFluxbaseQuery((client) => client.from('products').select('*'), { queryKey: ['products'] }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBe(error);
  });

  it('should handle single item response', async () => {
    const mockData = { id: 1, name: 'Test' };
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      single: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useFluxbaseQuery((client) => client.from('products').select('*'), { queryKey: ['product'] }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual([mockData]);
  });

  it('should handle null response', async () => {
    const executeMock = vi.fn().mockResolvedValue({ data: null, error: null });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useFluxbaseQuery((client) => client.from('products').select('*'), { queryKey: ['products'] }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual([]);
  });

  it('should generate query key from function when not provided', async () => {
    const mockData = [{ id: 1 }];
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const buildQuery = (client: any) => client.from('test').select('*');
    const { result } = renderHook(
      () => useFluxbaseQuery(buildQuery),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockData);
  });
});

describe('useTable', () => {
  it('should query table with builder function', async () => {
    const mockData = [{ id: 1, name: 'Test' }];
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      eq: vi.fn().mockReturnThis(),
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useTable('products', (q) => q.select('*').eq('active', true)),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockData);
    expect(fromMock).toHaveBeenCalledWith('products');
  });

  it('should query table without builder function', async () => {
    const mockData = [{ id: 1 }];
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(
      () => useTable('products'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockData);
  });

  it('should support custom query key', async () => {
    const mockData = [{ id: 1 }];
    const executeMock = vi.fn().mockResolvedValue({ data: mockData, error: null });
    const fromMock = vi.fn().mockReturnValue({
      execute: executeMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(
      () => useTable('products', undefined, { queryKey: ['custom', 'key'] }),
      { wrapper: createWrapper(client, queryClient) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(queryClient.getQueryData(['custom', 'key'])).toEqual(mockData);
  });
});

describe('useInsert', () => {
  it('should insert data and invalidate queries', async () => {
    const mockResult = { id: 1, name: 'New Item' };
    const insertMock = vi.fn().mockResolvedValue({ data: mockResult, error: null });
    const fromMock = vi.fn().mockReturnValue({
      insert: insertMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useInsert('products'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ name: 'New Item' });
    });

    expect(fromMock).toHaveBeenCalledWith('products');
    expect(insertMock).toHaveBeenCalledWith({ name: 'New Item' });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'table', 'products'] });
  });

  it('should throw error on insert failure', async () => {
    const error = new Error('Insert failed');
    const insertMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({
      insert: insertMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(() => useInsert('products'), {
      wrapper: createWrapper(client),
    });

    await expect(act(async () => {
      await result.current.mutateAsync({ name: 'New Item' });
    })).rejects.toThrow('Insert failed');
  });
});

describe('useUpdate', () => {
  it('should update data and invalidate queries', async () => {
    const mockResult = { id: 1, name: 'Updated' };
    const updateMock = vi.fn().mockResolvedValue({ data: mockResult, error: null });
    const eqMock = vi.fn().mockReturnThis();
    const fromMock = vi.fn().mockReturnValue({
      eq: eqMock,
      update: updateMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useUpdate('products'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({
        data: { name: 'Updated' },
        buildQuery: (q) => q.eq('id', 1),
      });
    });

    expect(fromMock).toHaveBeenCalledWith('products');
    expect(updateMock).toHaveBeenCalledWith({ name: 'Updated' });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'table', 'products'] });
  });

  it('should throw error on update failure', async () => {
    const error = new Error('Update failed');
    const updateMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({
      eq: vi.fn().mockReturnThis(),
      update: updateMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(() => useUpdate('products'), {
      wrapper: createWrapper(client),
    });

    await expect(act(async () => {
      await result.current.mutateAsync({
        data: { name: 'Updated' },
        buildQuery: (q) => q.eq('id', 1),
      });
    })).rejects.toThrow('Update failed');
  });
});

describe('useUpsert', () => {
  it('should upsert data and invalidate queries', async () => {
    const mockResult = { id: 1, name: 'Upserted' };
    const upsertMock = vi.fn().mockResolvedValue({ data: mockResult, error: null });
    const fromMock = vi.fn().mockReturnValue({
      upsert: upsertMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useUpsert('products'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Upserted' });
    });

    expect(fromMock).toHaveBeenCalledWith('products');
    expect(upsertMock).toHaveBeenCalledWith({ id: 1, name: 'Upserted' });
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'table', 'products'] });
  });

  it('should throw error on upsert failure', async () => {
    const error = new Error('Upsert failed');
    const upsertMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({
      upsert: upsertMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(() => useUpsert('products'), {
      wrapper: createWrapper(client),
    });

    await expect(act(async () => {
      await result.current.mutateAsync({ id: 1, name: 'Upserted' });
    })).rejects.toThrow('Upsert failed');
  });
});

describe('useDelete', () => {
  it('should delete data and invalidate queries', async () => {
    const deleteMock = vi.fn().mockResolvedValue({ data: null, error: null });
    const eqMock = vi.fn().mockReturnThis();
    const fromMock = vi.fn().mockReturnValue({
      eq: eqMock,
      delete: deleteMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useDelete('products'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync((q) => q.eq('id', 1));
    });

    expect(fromMock).toHaveBeenCalledWith('products');
    expect(deleteMock).toHaveBeenCalled();
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'table', 'products'] });
  });

  it('should throw error on delete failure', async () => {
    const error = new Error('Delete failed');
    const deleteMock = vi.fn().mockResolvedValue({ error });
    const fromMock = vi.fn().mockReturnValue({
      eq: vi.fn().mockReturnThis(),
      delete: deleteMock,
    });
    const client = createMockClient({ from: fromMock } as any);

    const { result } = renderHook(() => useDelete('products'), {
      wrapper: createWrapper(client),
    });

    await expect(act(async () => {
      await result.current.mutateAsync((q) => q.eq('id', 1));
    })).rejects.toThrow('Delete failed');
  });
});
