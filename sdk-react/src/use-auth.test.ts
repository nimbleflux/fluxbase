/**
 * Tests for authentication hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useUser,
  useSession,
  useSignIn,
  useSignUp,
  useSignOut,
  useUpdateUser,
  useAuth,
} from './use-auth';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useUser', () => {
  it('should return null when no session', async () => {
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: null }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBeNull();
  });

  it('should return user when session exists', async () => {
    const mockUser = { id: '1', email: 'test@example.com' };
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: { access_token: 'token' } }, error: null }),
        getCurrentUser: vi.fn().mockResolvedValue({ data: { user: mockUser }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockUser);
  });

  it('should return null when getCurrentUser fails', async () => {
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: { access_token: 'token' } }, error: null }),
        getCurrentUser: vi.fn().mockRejectedValue(new Error('Not authenticated')),
      },
    } as any);

    const { result } = renderHook(() => useUser(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBeNull();
  });
});

describe('useSession', () => {
  it('should return null when no session', async () => {
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: null }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useSession(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBeNull();
  });

  it('should return session when exists', async () => {
    const mockSession = { access_token: 'token', refresh_token: 'refresh' };
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: mockSession }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useSession(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockSession);
  });
});

describe('useSignIn', () => {
  it('should call signIn and update cache on success', async () => {
    const mockSession = { user: { id: '1', email: 'test@example.com' }, access_token: 'token' };
    const signInMock = vi.fn().mockResolvedValue(mockSession);
    const client = createMockClient({
      auth: {
        signIn: signInMock,
      },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useSignIn(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ email: 'test@example.com', password: 'password' });
    });

    expect(signInMock).toHaveBeenCalledWith({ email: 'test@example.com', password: 'password' });
    expect(queryClient.getQueryData(['fluxbase', 'auth', 'session'])).toEqual(mockSession);
  });

  it('should handle 2FA required response (no user)', async () => {
    const mockResponse = { mfa_required: true };
    const signInMock = vi.fn().mockResolvedValue(mockResponse);
    const client = createMockClient({
      auth: {
        signIn: signInMock,
      },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useSignIn(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ email: 'test@example.com', password: 'password' });
    });

    // User should not be set when 2FA is required
    expect(queryClient.getQueryData(['fluxbase', 'auth', 'user'])).toBeUndefined();
  });
});

describe('useSignUp', () => {
  it('should call signUp and return response on success', async () => {
    const mockResponse = {
      data: {
        user: { id: '1', email: 'test@example.com' },
        session: { access_token: 'token' },
      },
    };
    const signUpMock = vi.fn().mockResolvedValue(mockResponse);
    const client = createMockClient({
      auth: {
        signUp: signUpMock,
      },
    } as any);

    const { result } = renderHook(() => useSignUp(), {
      wrapper: createWrapper(client),
    });

    let response;
    await act(async () => {
      response = await result.current.mutateAsync({ email: 'test@example.com', password: 'password' });
    });

    expect(signUpMock).toHaveBeenCalledWith({ email: 'test@example.com', password: 'password' });
    expect(response).toEqual(mockResponse);
  });

  it('should handle signup without immediate session', async () => {
    const mockResponse = { data: null };
    const signUpMock = vi.fn().mockResolvedValue(mockResponse);
    const client = createMockClient({
      auth: {
        signUp: signUpMock,
      },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useSignUp(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ email: 'test@example.com', password: 'password' });
    });

    // Cache should not be updated when data is null
    expect(queryClient.getQueryData(['fluxbase', 'auth', 'session'])).toBeUndefined();
  });
});

describe('useSignOut', () => {
  it('should call signOut successfully', async () => {
    const signOutMock = vi.fn().mockResolvedValue(undefined);
    const client = createMockClient({
      auth: {
        signOut: signOutMock,
      },
    } as any);

    const { result } = renderHook(() => useSignOut(), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.mutateAsync();
    });

    expect(signOutMock).toHaveBeenCalled();
  });
});

describe('useUpdateUser', () => {
  it('should call updateUser and update cache', async () => {
    const updatedUser = { id: '1', email: 'new@example.com' };
    const updateUserMock = vi.fn().mockResolvedValue(updatedUser);
    const client = createMockClient({
      auth: {
        updateUser: updateUserMock,
      },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(() => useUpdateUser(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ email: 'new@example.com' });
    });

    expect(updateUserMock).toHaveBeenCalledWith({ email: 'new@example.com' });
    expect(queryClient.getQueryData(['fluxbase', 'auth', 'user'])).toEqual(updatedUser);
  });
});

describe('useAuth', () => {
  it('should return combined auth state', async () => {
    const mockUser = { id: '1', email: 'test@example.com' };
    const mockSession = { access_token: 'token' };
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: mockSession }, error: null }),
        getCurrentUser: vi.fn().mockResolvedValue({ data: { user: mockUser }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.user).toEqual(mockUser);
    expect(result.current.session).toEqual(mockSession);
    expect(result.current.isAuthenticated).toBe(true);
    expect(result.current.signIn).toBeDefined();
    expect(result.current.signUp).toBeDefined();
    expect(result.current.signOut).toBeDefined();
    expect(result.current.updateUser).toBeDefined();
  });

  it('should show loading state initially', () => {
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockImplementation(() => new Promise(() => {})), // Never resolves
      },
    } as any);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(client),
    });

    expect(result.current.isLoading).toBe(true);
  });

  it('should show unauthenticated state when no session', async () => {
    const client = createMockClient({
      auth: {
        getSession: vi.fn().mockResolvedValue({ data: { session: null }, error: null }),
      },
    } as any);

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    expect(result.current.user).toBeNull();
    expect(result.current.session).toBeNull();
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('should have pending states for mutations', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useAuth(), {
      wrapper: createWrapper(client),
    });

    expect(result.current.isSigningIn).toBe(false);
    expect(result.current.isSigningUp).toBe(false);
    expect(result.current.isSigningOut).toBe(false);
    expect(result.current.isUpdating).toBe(false);
  });
});
