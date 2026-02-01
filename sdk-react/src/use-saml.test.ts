/**
 * Tests for SAML SSO hooks
 */

import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useSAMLProviders,
  useGetSAMLLoginUrl,
  useSignInWithSAML,
  useHandleSAMLCallback,
  useSAMLMetadataUrl,
} from './use-saml';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useSAMLProviders', () => {
  it('should fetch SAML providers', async () => {
    const mockProviders = [
      { id: '1', name: 'okta', display_name: 'Okta' },
      { id: '2', name: 'azure', display_name: 'Azure AD' },
    ];
    const getSAMLProvidersMock = vi.fn().mockResolvedValue({
      data: { providers: mockProviders },
      error: null,
    });

    const client = createMockClient({
      auth: { getSAMLProviders: getSAMLProvidersMock },
    } as any);

    const { result } = renderHook(
      () => useSAMLProviders(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockProviders);
  });

  it('should throw error on fetch failure', async () => {
    const error = new Error('Failed to fetch');
    const getSAMLProvidersMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      auth: { getSAMLProviders: getSAMLProvidersMock },
    } as any);

    const { result } = renderHook(
      () => useSAMLProviders(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBe(error);
  });
});

describe('useGetSAMLLoginUrl', () => {
  it('should get SAML login URL', async () => {
    const mockUrl = 'https://idp.example.com/sso/saml?SAMLRequest=...';
    const getSAMLLoginUrlMock = vi.fn().mockResolvedValue({
      data: { url: mockUrl },
      error: null,
    });

    const client = createMockClient({
      auth: { getSAMLLoginUrl: getSAMLLoginUrlMock },
    } as any);

    const { result } = renderHook(
      () => useGetSAMLLoginUrl(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        provider: 'okta',
        options: { redirectUrl: 'https://app.example.com/callback' },
      });
    });

    expect(getSAMLLoginUrlMock).toHaveBeenCalledWith('okta', {
      redirectUrl: 'https://app.example.com/callback',
    });
  });

  it('should get SAML login URL without options', async () => {
    const getSAMLLoginUrlMock = vi.fn().mockResolvedValue({
      data: { url: 'https://idp.example.com/sso' },
      error: null,
    });

    const client = createMockClient({
      auth: { getSAMLLoginUrl: getSAMLLoginUrlMock },
    } as any);

    const { result } = renderHook(
      () => useGetSAMLLoginUrl(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({ provider: 'okta' });
    });

    expect(getSAMLLoginUrlMock).toHaveBeenCalledWith('okta', undefined);
  });
});

describe('useSignInWithSAML', () => {
  it('should initiate SAML sign in', async () => {
    const signInWithSAMLMock = vi.fn().mockResolvedValue({ data: null, error: null });

    const client = createMockClient({
      auth: { signInWithSAML: signInWithSAMLMock },
    } as any);

    const { result } = renderHook(
      () => useSignInWithSAML(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({ provider: 'okta' });
    });

    expect(signInWithSAMLMock).toHaveBeenCalledWith('okta', undefined);
  });

  it('should pass options to sign in', async () => {
    const signInWithSAMLMock = vi.fn().mockResolvedValue({ data: null, error: null });

    const client = createMockClient({
      auth: { signInWithSAML: signInWithSAMLMock },
    } as any);

    const { result } = renderHook(
      () => useSignInWithSAML(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        provider: 'okta',
        options: { redirectUrl: 'https://app.example.com' },
      });
    });

    expect(signInWithSAMLMock).toHaveBeenCalledWith('okta', {
      redirectUrl: 'https://app.example.com',
    });
  });
});

describe('useHandleSAMLCallback', () => {
  it('should handle SAML callback successfully', async () => {
    const mockResult = {
      data: {
        user: { id: '1', email: 'user@example.com' },
        session: { access_token: 'token' },
      },
      error: null,
    };
    const handleSAMLCallbackMock = vi.fn().mockResolvedValue(mockResult);

    const client = createMockClient({
      auth: { handleSAMLCallback: handleSAMLCallbackMock },
    } as any);

    const { result } = renderHook(
      () => useHandleSAMLCallback(),
      { wrapper: createWrapper(client) }
    );

    let response;
    await act(async () => {
      response = await result.current.mutateAsync({ samlResponse: 'base64-response' });
    });

    expect(handleSAMLCallbackMock).toHaveBeenCalledWith('base64-response', undefined);
    expect(response).toEqual(mockResult);
    expect(result.current.isSuccess).toBe(true);
  });

  it('should pass provider to callback handler', async () => {
    const handleSAMLCallbackMock = vi.fn().mockResolvedValue({
      data: { user: {}, session: {} },
      error: null,
    });

    const client = createMockClient({
      auth: { handleSAMLCallback: handleSAMLCallbackMock },
    } as any);

    const { result } = renderHook(
      () => useHandleSAMLCallback(),
      { wrapper: createWrapper(client) }
    );

    await act(async () => {
      await result.current.mutateAsync({
        samlResponse: 'base64-response',
        provider: 'okta',
      });
    });

    expect(handleSAMLCallbackMock).toHaveBeenCalledWith('base64-response', 'okta');
  });

  it('should not update cache when no data', async () => {
    const handleSAMLCallbackMock = vi.fn().mockResolvedValue({
      data: null,
      error: null,
    });

    const client = createMockClient({
      auth: { handleSAMLCallback: handleSAMLCallbackMock },
    } as any);

    const queryClient = createTestQueryClient();
    const { result } = renderHook(
      () => useHandleSAMLCallback(),
      { wrapper: createWrapper(client, queryClient) }
    );

    await act(async () => {
      await result.current.mutateAsync({ samlResponse: 'base64-response' });
    });

    // Cache should not be updated
    expect(queryClient.getQueryData(['fluxbase', 'auth', 'session'])).toBeUndefined();
  });
});

describe('useSAMLMetadataUrl', () => {
  it('should return metadata URL generator function', () => {
    const getSAMLMetadataUrlMock = vi.fn().mockReturnValue('https://api.example.com/saml/metadata/okta');

    const client = createMockClient({
      auth: { getSAMLMetadataUrl: getSAMLMetadataUrlMock },
    } as any);

    const { result } = renderHook(
      () => useSAMLMetadataUrl(),
      { wrapper: createWrapper(client) }
    );

    const url = result.current('okta');

    expect(getSAMLMetadataUrlMock).toHaveBeenCalledWith('okta');
    expect(url).toBe('https://api.example.com/saml/metadata/okta');
  });

  it('should work with different providers', () => {
    const getSAMLMetadataUrlMock = vi.fn().mockImplementation((provider) => `https://api.example.com/saml/metadata/${provider}`);

    const client = createMockClient({
      auth: { getSAMLMetadataUrl: getSAMLMetadataUrlMock },
    } as any);

    const { result } = renderHook(
      () => useSAMLMetadataUrl(),
      { wrapper: createWrapper(client) }
    );

    expect(result.current('okta')).toBe('https://api.example.com/saml/metadata/okta');
    expect(result.current('azure')).toBe('https://api.example.com/saml/metadata/azure');
  });
});
