/**
 * Tests for client keys management hook
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useClientKeys, useAPIKeys } from './use-client-keys';
import { createMockClient, createWrapper } from './test-utils';

describe('useClientKeys', () => {
  it('should fetch client keys on mount when autoFetch is true', async () => {
    const mockKeys = [
      { id: '1', name: 'Key 1', description: 'First key' },
      { id: '2', name: 'Key 2', description: 'Second key' },
    ];
    const listMock = vi.fn().mockResolvedValue({ client_keys: mockKeys });

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.keys).toEqual(mockKeys);
  });

  it('should not fetch keys when autoFetch is false', async () => {
    const listMock = vi.fn();

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: false }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(listMock).not.toHaveBeenCalled();
  });

  it('should create client key', async () => {
    const mockKeys: any[] = [];
    const listMock = vi.fn().mockResolvedValue({ client_keys: mockKeys });
    const createMock = vi.fn().mockResolvedValue({
      key: 'new-secret-key',
      client_key: { id: '1', name: 'New Key' },
    });

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: createMock,
            update: vi.fn(),
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    let response;
    await act(async () => {
      response = await result.current.createKey({
        name: 'New Key',
        description: 'A new key',
        scopes: ['read', 'write'],
        rate_limit_per_minute: 60,
      });
    });

    expect(createMock).toHaveBeenCalledWith({
      name: 'New Key',
      description: 'A new key',
      scopes: ['read', 'write'],
      rate_limit_per_minute: 60,
    });
    expect(response).toEqual({
      key: 'new-secret-key',
      keyData: { id: '1', name: 'New Key' },
    });
    // Should refetch after create
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it('should update client key', async () => {
    const mockKeys = [{ id: '1', name: 'Key 1' }];
    const listMock = vi.fn().mockResolvedValue({ client_keys: mockKeys });
    const updateMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: updateMock,
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.updateKey('1', { name: 'Updated Key' });
    });

    expect(updateMock).toHaveBeenCalledWith('1', { name: 'Updated Key' });
    // Should refetch after update
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it('should revoke client key', async () => {
    const mockKeys = [{ id: '1', name: 'Key 1' }];
    const listMock = vi.fn().mockResolvedValue({ client_keys: mockKeys });
    const revokeMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: revokeMock,
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.revokeKey('1');
    });

    expect(revokeMock).toHaveBeenCalledWith('1');
    // Should refetch after revoke
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it('should delete client key', async () => {
    const mockKeys = [{ id: '1', name: 'Key 1' }];
    const listMock = vi.fn().mockResolvedValue({ client_keys: mockKeys });
    const deleteMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: vi.fn(),
            delete: deleteMock,
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.deleteKey('1');
    });

    expect(deleteMock).toHaveBeenCalledWith('1');
    // Should refetch after delete
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it('should handle fetch error', async () => {
    const error = new Error('Failed to fetch');
    const listMock = vi.fn().mockRejectedValue(error);

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.error).toBe(error);
  });

  it('should refetch on demand', async () => {
    const listMock = vi.fn().mockResolvedValue({ client_keys: [] });

    const client = createMockClient({
      admin: {
        management: {
          clientKeys: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            revoke: vi.fn(),
            delete: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useClientKeys({ autoFetch: false }),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.refetch();
    });

    expect(listMock).toHaveBeenCalledTimes(1);
  });
});

describe('useAPIKeys (deprecated alias)', () => {
  it('should be the same as useClientKeys', () => {
    expect(useAPIKeys).toBe(useClientKeys);
  });
});
