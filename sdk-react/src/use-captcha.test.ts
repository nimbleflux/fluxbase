/**
 * Tests for CAPTCHA hooks
 */

import { describe, it, expect, vi } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useCaptchaConfig, useCaptcha, isCaptchaRequiredForEndpoint } from './use-captcha';
import { createMockClient, createWrapper } from './test-utils';

describe('useCaptchaConfig', () => {
  it('should fetch CAPTCHA configuration', async () => {
    const mockConfig = {
      enabled: true,
      provider: 'hcaptcha',
      site_key: 'test-site-key',
      endpoints: ['signup', 'login'],
    };
    const getCaptchaConfigMock = vi.fn().mockResolvedValue({ data: mockConfig, error: null });

    const client = createMockClient({
      auth: { getCaptchaConfig: getCaptchaConfigMock },
    } as any);

    const { result } = renderHook(
      () => useCaptchaConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockConfig);
  });

  it('should throw error on fetch failure', async () => {
    const error = new Error('Failed to fetch');
    const getCaptchaConfigMock = vi.fn().mockResolvedValue({ data: null, error });

    const client = createMockClient({
      auth: { getCaptchaConfig: getCaptchaConfigMock },
    } as any);

    const { result } = renderHook(
      () => useCaptchaConfig(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBe(error);
  });
});

describe('useCaptcha', () => {
  it('should initialize with empty state', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    expect(result.current.token).toBeNull();
    expect(result.current.isLoading).toBe(false);
    expect(result.current.error).toBeNull();
  });

  it('should become ready when provider is set', async () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isReady).toBe(true));
  });

  it('should handle onVerify callback', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    act(() => {
      result.current.onVerify('test-token');
    });

    expect(result.current.token).toBe('test-token');
    expect(result.current.isLoading).toBe(false);
  });

  it('should handle onExpire callback', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    // First verify
    act(() => {
      result.current.onVerify('test-token');
    });

    expect(result.current.token).toBe('test-token');

    // Then expire
    act(() => {
      result.current.onExpire();
    });

    expect(result.current.token).toBeNull();
    expect(result.current.isReady).toBe(true);
  });

  it('should handle onError callback', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    const error = new Error('CAPTCHA error');
    act(() => {
      result.current.onError(error);
    });

    expect(result.current.error).toBe(error);
    expect(result.current.token).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });

  it('should reset state', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    // Set a token
    act(() => {
      result.current.onVerify('test-token');
    });

    expect(result.current.token).toBe('test-token');

    // Reset
    act(() => {
      result.current.reset();
    });

    expect(result.current.token).toBeNull();
    expect(result.current.error).toBeNull();
    expect(result.current.isLoading).toBe(false);
  });

  it('should return existing token in execute', async () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    // Set a token
    act(() => {
      result.current.onVerify('existing-token');
    });

    // Execute should return existing token
    let token;
    await act(async () => {
      token = await result.current.execute();
    });

    expect(token).toBe('existing-token');
  });

  it('should return empty string in execute when no provider', async () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha(undefined),
      { wrapper: createWrapper(client) }
    );

    let token;
    await act(async () => {
      token = await result.current.execute();
    });

    expect(token).toBe('');
  });

  it('should set loading state during execute', async () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    // Start execute (returns a promise that needs onVerify to resolve)
    let executePromise: Promise<string>;
    act(() => {
      executePromise = result.current.execute();
    });

    expect(result.current.isLoading).toBe(true);

    // Resolve by calling onVerify
    act(() => {
      result.current.onVerify('new-token');
    });

    const token = await executePromise!;
    expect(token).toBe('new-token');
    expect(result.current.isLoading).toBe(false);
  });

  it('should reject execute on error', async () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useCaptcha('hcaptcha'),
      { wrapper: createWrapper(client) }
    );

    // Start execute
    let executePromise: Promise<string>;
    act(() => {
      executePromise = result.current.execute();
    });

    // Trigger error
    const error = new Error('CAPTCHA failed');
    act(() => {
      result.current.onError(error);
    });

    await expect(executePromise!).rejects.toThrow('CAPTCHA failed');
  });
});

describe('isCaptchaRequiredForEndpoint', () => {
  it('should return false when CAPTCHA is disabled', () => {
    const config = { enabled: false, endpoints: ['signup', 'login'] };
    expect(isCaptchaRequiredForEndpoint(config as any, 'signup')).toBe(false);
  });

  it('should return false when config is undefined', () => {
    expect(isCaptchaRequiredForEndpoint(undefined, 'signup')).toBe(false);
  });

  it('should return true when endpoint is in list', () => {
    const config = { enabled: true, endpoints: ['signup', 'login'] };
    expect(isCaptchaRequiredForEndpoint(config as any, 'signup')).toBe(true);
    expect(isCaptchaRequiredForEndpoint(config as any, 'login')).toBe(true);
  });

  it('should return false when endpoint is not in list', () => {
    const config = { enabled: true, endpoints: ['signup'] };
    expect(isCaptchaRequiredForEndpoint(config as any, 'login')).toBe(false);
    expect(isCaptchaRequiredForEndpoint(config as any, 'password_reset')).toBe(false);
  });

  it('should return false when endpoints is undefined', () => {
    const config = { enabled: true };
    expect(isCaptchaRequiredForEndpoint(config as any, 'signup')).toBe(false);
  });
});
