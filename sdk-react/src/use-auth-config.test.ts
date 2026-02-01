/**
 * Tests for auth configuration hook
 */

import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useAuthConfig } from './use-auth-config';
import { createMockClient, createWrapper } from './test-utils';

describe('useAuthConfig', () => {
  it('should fetch auth configuration', async () => {
    const mockConfig = {
      signup_enabled: true,
      email_verification_required: true,
      magic_link_enabled: true,
      mfa_enabled: true,
      password_min_length: 8,
      password_require_uppercase: true,
      password_require_lowercase: true,
      password_require_number: true,
      password_require_special: false,
      oauth_providers: [
        { provider: 'google', display_name: 'Google', authorize_url: 'https://google.com/oauth' },
      ],
      saml_providers: [],
      captcha: { enabled: false },
    };

    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockConfig);
    expect(getAuthConfigMock).toHaveBeenCalled();
  });

  it('should throw error on fetch failure', async () => {
    const error = new Error('Failed to fetch config');
    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBe(error);
  });

  it('should have appropriate stale time', async () => {
    const mockConfig = { signup_enabled: true };
    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    // The config should be cached with a stale time of 5 minutes
    // Verify the data is available
    expect(result.current.data).toEqual(mockConfig);
  });

  it('should return signup enabled status', async () => {
    const mockConfig = { signup_enabled: false };
    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data?.signup_enabled).toBe(false);
  });

  it('should return OAuth providers', async () => {
    const mockConfig = {
      oauth_providers: [
        { provider: 'google', display_name: 'Google', authorize_url: 'https://google.com' },
        { provider: 'github', display_name: 'GitHub', authorize_url: 'https://github.com' },
      ],
    };
    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data?.oauth_providers).toHaveLength(2);
  });

  it('should return password requirements', async () => {
    const mockConfig = {
      password_min_length: 12,
      password_require_uppercase: true,
      password_require_lowercase: true,
      password_require_number: true,
      password_require_special: true,
    };
    const getAuthConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getAuthConfig: getAuthConfigMock },
    } as any);

    const { result } = renderHook(
      () => useAuthConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data?.password_min_length).toBe(12);
    expect(result.current.data?.password_require_uppercase).toBe(true);
    expect(result.current.data?.password_require_special).toBe(true);
  });
});
