/**
 * Tests for users management hook
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useUsers } from './use-users';
import { createMockClient, createWrapper } from './test-utils';

describe('useUsers', () => {
  it('should fetch users on mount when autoFetch is true', async () => {
    const mockUsers = [
      { id: '1', email: 'user1@example.com', role: 'user' },
      { id: '2', email: 'user2@example.com', role: 'admin' },
    ];
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: mockUsers, total: 2 },
      error: null,
    });

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.users).toEqual(mockUsers);
    expect(result.current.total).toBe(2);
  });

  it('should not fetch users when autoFetch is false', async () => {
    const listUsersMock = vi.fn();

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: false }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(listUsersMock).not.toHaveBeenCalled();
  });

  it('should pass list options', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    renderHook(
      () => useUsers({ autoFetch: true, limit: 10, search: 'test' }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => {
      expect(listUsersMock).toHaveBeenCalledWith(
        expect.objectContaining({ limit: 10, search: 'test' })
      );
    });
  });

  it('should invite user', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });
    const inviteUserMock = vi.fn().mockResolvedValue({ data: {}, error: null });

    const client = createMockClient({
      admin: {
        listUsers: listUsersMock,
        inviteUser: inviteUserMock,
      },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.inviteUser('new@example.com', 'user');
    });

    expect(inviteUserMock).toHaveBeenCalledWith({ email: 'new@example.com', role: 'user' });
    // Should refetch after invite
    expect(listUsersMock).toHaveBeenCalledTimes(2);
  });

  it('should update user role', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });
    const updateUserRoleMock = vi.fn().mockResolvedValue({ data: {}, error: null });

    const client = createMockClient({
      admin: {
        listUsers: listUsersMock,
        updateUserRole: updateUserRoleMock,
      },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.updateUserRole('1', 'admin');
    });

    expect(updateUserRoleMock).toHaveBeenCalledWith('1', 'admin');
    // Should refetch after update
    expect(listUsersMock).toHaveBeenCalledTimes(2);
  });

  it('should delete user', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });
    const deleteUserMock = vi.fn().mockResolvedValue({ data: {}, error: null });

    const client = createMockClient({
      admin: {
        listUsers: listUsersMock,
        deleteUser: deleteUserMock,
      },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.deleteUser('1');
    });

    expect(deleteUserMock).toHaveBeenCalledWith('1');
    // Should refetch after delete
    expect(listUsersMock).toHaveBeenCalledTimes(2);
  });

  it('should reset user password', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });
    const resetPasswordMock = vi.fn().mockResolvedValue({
      data: { message: 'Password reset email sent' },
      error: null,
    });

    const client = createMockClient({
      admin: {
        listUsers: listUsersMock,
        resetUserPassword: resetPasswordMock,
      },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    let response;
    await act(async () => {
      response = await result.current.resetPassword('1');
    });

    expect(resetPasswordMock).toHaveBeenCalledWith('1');
    expect(response).toEqual({ message: 'Password reset email sent' });
  });

  it('should handle fetch error', async () => {
    const error = new Error('Failed to fetch');
    const listUsersMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.error).toBe(error);
  });

  it('should set up refetch interval', async () => {
    vi.useFakeTimers();
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    const { unmount } = renderHook(
      () => useUsers({ autoFetch: true, refetchInterval: 5000 }),
      { wrapper: createWrapper(client) }
    );

    // Initial fetch
    expect(listUsersMock).toHaveBeenCalledTimes(1);

    // Advance timer
    await act(async () => {
      vi.advanceTimersByTime(5000);
    });

    expect(listUsersMock).toHaveBeenCalledTimes(2);

    unmount();
    vi.useRealTimers();
  });

  it('should refetch on demand', async () => {
    const listUsersMock = vi.fn().mockResolvedValue({
      data: { users: [], total: 0 },
      error: null,
    });

    const client = createMockClient({
      admin: { listUsers: listUsersMock },
    } as any);

    const { result } = renderHook(
      () => useUsers({ autoFetch: false }),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.refetch();
    });

    expect(listUsersMock).toHaveBeenCalledTimes(1);
  });
});
