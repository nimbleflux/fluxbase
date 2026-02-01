/**
 * Tests for FluxbaseProvider and useFluxbaseClient
 */

import { describe, it, expect, vi } from 'vitest';
import { render, screen, renderHook } from '@testing-library/react';
import React from 'react';
import { FluxbaseProvider, useFluxbaseClient } from './context';
import { createMockClient } from './test-utils';
import type { FluxbaseClient } from '@fluxbase/sdk';

describe('FluxbaseProvider', () => {
  it('should render children', () => {
    const client = createMockClient();
    render(
      <FluxbaseProvider client={client}>
        <div data-testid="child">Hello</div>
      </FluxbaseProvider>
    );

    expect(screen.getByTestId('child')).toHaveTextContent('Hello');
  });

  it('should provide client to children', () => {
    const client = createMockClient();
    let receivedClient: FluxbaseClient | null = null;

    function TestComponent() {
      receivedClient = useFluxbaseClient();
      return null;
    }

    render(
      <FluxbaseProvider client={client}>
        <TestComponent />
      </FluxbaseProvider>
    );

    expect(receivedClient).toBe(client);
  });

  it('should support nested children', () => {
    const client = createMockClient();
    render(
      <FluxbaseProvider client={client}>
        <div>
          <span>
            <p data-testid="nested">Nested content</p>
          </span>
        </div>
      </FluxbaseProvider>
    );

    expect(screen.getByTestId('nested')).toHaveTextContent('Nested content');
  });
});

describe('useFluxbaseClient', () => {
  it('should return the client from context', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current).toBe(client);
  });

  it('should throw error when used outside provider', () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    expect(() => {
      renderHook(() => useFluxbaseClient());
    }).toThrow('useFluxbaseClient must be used within a FluxbaseProvider');

    consoleSpy.mockRestore();
  });

  it('should provide access to auth methods', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current.auth).toBeDefined();
    expect(result.current.auth.getSession).toBeDefined();
    expect(result.current.auth.signIn).toBeDefined();
  });

  it('should provide access to storage methods', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current.storage).toBeDefined();
    expect(result.current.storage.from).toBeDefined();
  });

  it('should provide access to realtime methods', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current.realtime).toBeDefined();
    expect(result.current.realtime.channel).toBeDefined();
  });

  it('should provide access to graphql methods', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current.graphql).toBeDefined();
    expect(result.current.graphql.execute).toBeDefined();
  });

  it('should provide access to admin methods', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useFluxbaseClient(), {
      wrapper: ({ children }) => (
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      ),
    });

    expect(result.current.admin).toBeDefined();
    expect(result.current.admin.me).toBeDefined();
  });
});
