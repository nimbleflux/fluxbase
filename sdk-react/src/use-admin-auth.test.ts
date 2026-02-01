/**
 * Tests for admin authentication hook
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useAdminAuth } from './use-admin-auth';
import { createMockClient, createWrapper } from './test-utils';

describe('useAdminAuth', () => {
  it('should check auth status on mount when autoCheck is true', async () => {
    const mockUser = { id: '1', email: 'admin@example.com', role: 'admin' };
    const meMock = vi.fn().mockResolvedValue({ data: { user: mockUser }, error: null });

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toEqual(mockUser);
    expect(result.current.isAuthenticated).toBe(true);
    expect(meMock).toHaveBeenCalled();
  });

  it('should not check auth status when autoCheck is false', async () => {
    const meMock = vi.fn();

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: false }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(meMock).not.toHaveBeenCalled();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('should handle auth check error', async () => {
    const error = new Error('Not authenticated');
    const meMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
    expect(result.current.error).toBe(error);
  });

  it('should login successfully', async () => {
    const mockUser = { id: '1', email: 'admin@example.com', role: 'admin' };
    const loginMock = vi.fn().mockResolvedValue({
      data: { user: mockUser, token: 'token' },
      error: null,
    });

    const client = createMockClient({
      admin: { login: loginMock, me: vi.fn().mockResolvedValue({ data: null, error: null }) },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: false }),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.login('admin@example.com', 'password');
    });

    expect(loginMock).toHaveBeenCalledWith({ email: 'admin@example.com', password: 'password' });
    expect(result.current.user).toEqual(mockUser);
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('should handle login error', async () => {
    const error = new Error('Invalid credentials');
    const loginMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      admin: { login: loginMock, me: vi.fn().mockResolvedValue({ data: null, error: null }) },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: false }),
      { wrapper: createWrapper(client) }
    );

    await expect(act(async () => {
      await result.current.login('admin@example.com', 'wrong-password');
    })).rejects.toThrow();

    // User should remain null after failed login
    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('should logout', async () => {
    const mockUser = { id: '1', email: 'admin@example.com', role: 'admin' };
    const meMock = vi.fn().mockResolvedValue({ data: { user: mockUser }, error: null });

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: true }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isAuthenticated).toBe(true));

    await act(async () => {
      await result.current.logout();
    });

    expect(result.current.user).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('should refresh user info', async () => {
    const mockUser = { id: '1', email: 'admin@example.com', role: 'admin' };
    const meMock = vi.fn().mockResolvedValue({ data: { user: mockUser }, error: null });

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: false }),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.refresh();
    });

    expect(meMock).toHaveBeenCalledTimes(1);
    expect(result.current.user).toEqual(mockUser);
  });

  it('should show loading state during operations', async () => {
    const meMock = vi.fn().mockImplementation(() => new Promise((resolve) => {
      setTimeout(() => resolve({ data: { user: {} }, error: null }), 100);
    }));

    const client = createMockClient({
      admin: { me: meMock },
    } as any);

    const { result } = renderHook(
      () => useAdminAuth({ autoCheck: true }),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.isLoading).toBe(true);

    await waitFor(() => expect(result.current.isLoading).toBe(false));
  });
});
