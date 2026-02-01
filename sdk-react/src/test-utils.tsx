/**
 * Test utilities for Fluxbase React SDK
 */

import React, { ReactElement } from "react";
import { render, RenderOptions } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { FluxbaseProvider } from "./context";
import type { FluxbaseClient } from "@fluxbase/sdk";
import { vi } from "vitest";

/**
 * Create a mock FluxbaseClient for testing
 */
export function createMockClient(
  overrides: Partial<FluxbaseClient> = {},
): FluxbaseClient {
  return {
    auth: {
      getSession: vi.fn().mockResolvedValue({ data: null, error: null }),
      getCurrentUser: vi.fn().mockResolvedValue({ data: null, error: null }),
      signIn: vi.fn().mockResolvedValue({ user: null, session: null }),
      signUp: vi.fn().mockResolvedValue({ data: null, error: null }),
      signOut: vi.fn().mockResolvedValue(undefined),
      updateUser: vi
        .fn()
        .mockResolvedValue({ id: "1", email: "test@example.com" }),
      getAuthConfig: vi.fn().mockResolvedValue({ data: {}, error: null }),
      getCaptchaConfig: vi.fn().mockResolvedValue({ data: {}, error: null }),
      getSAMLProviders: vi
        .fn()
        .mockResolvedValue({ data: { providers: [] }, error: null }),
      getSAMLLoginUrl: vi
        .fn()
        .mockResolvedValue({ data: { url: "" }, error: null }),
      signInWithSAML: vi.fn().mockResolvedValue({ data: null, error: null }),
      handleSAMLCallback: vi
        .fn()
        .mockResolvedValue({ data: null, error: null }),
      getSAMLMetadataUrl: vi
        .fn()
        .mockReturnValue("http://localhost/saml/metadata"),
      ...overrides.auth,
    },
    from: vi.fn().mockReturnValue({
      select: vi.fn().mockReturnThis(),
      insert: vi.fn().mockResolvedValue({ data: null, error: null }),
      update: vi.fn().mockResolvedValue({ data: null, error: null }),
      upsert: vi.fn().mockResolvedValue({ data: null, error: null }),
      delete: vi.fn().mockResolvedValue({ data: null, error: null }),
      execute: vi.fn().mockResolvedValue({ data: [], error: null }),
      eq: vi.fn().mockReturnThis(),
    }),
    storage: {
      from: vi.fn().mockReturnValue({
        list: vi.fn().mockResolvedValue({ data: [], error: null }),
        upload: vi
          .fn()
          .mockResolvedValue({ data: { path: "test.txt" }, error: null }),
        download: vi.fn().mockResolvedValue({ data: new Blob(), error: null }),
        remove: vi.fn().mockResolvedValue({ data: null, error: null }),
        getPublicUrl: vi
          .fn()
          .mockReturnValue({ data: { publicUrl: "http://localhost/file" } }),
        getTransformUrl: vi
          .fn()
          .mockReturnValue("http://localhost/transform/file"),
        createSignedUrl: vi
          .fn()
          .mockResolvedValue({
            data: { signedUrl: "http://localhost/signed" },
            error: null,
          }),
        move: vi
          .fn()
          .mockResolvedValue({ data: { path: "new.txt" }, error: null }),
        copy: vi
          .fn()
          .mockResolvedValue({ data: { path: "copy.txt" }, error: null }),
      }),
      listBuckets: vi.fn().mockResolvedValue({ data: [], error: null }),
      createBucket: vi.fn().mockResolvedValue({ error: null }),
      deleteBucket: vi.fn().mockResolvedValue({ error: null }),
      ...overrides.storage,
    },
    realtime: {
      channel: vi.fn().mockReturnValue({
        on: vi.fn().mockReturnThis(),
        subscribe: vi.fn().mockReturnThis(),
        unsubscribe: vi.fn(),
      }),
      ...overrides.realtime,
    },
    graphql: {
      execute: vi.fn().mockResolvedValue({ data: null, errors: null }),
      query: vi.fn().mockResolvedValue({ data: null, errors: null }),
      mutation: vi.fn().mockResolvedValue({ data: null, errors: null }),
      introspect: vi
        .fn()
        .mockResolvedValue({ data: { __schema: {} }, errors: null }),
      ...overrides.graphql,
    },
    admin: {
      me: vi.fn().mockResolvedValue({ data: null, error: null }),
      login: vi.fn().mockResolvedValue({ data: null, error: null }),
      listUsers: vi
        .fn()
        .mockResolvedValue({ data: { users: [], total: 0 }, error: null }),
      inviteUser: vi.fn().mockResolvedValue({ data: null, error: null }),
      updateUserRole: vi.fn().mockResolvedValue({ data: null, error: null }),
      deleteUser: vi.fn().mockResolvedValue({ data: null, error: null }),
      resetUserPassword: vi
        .fn()
        .mockResolvedValue({
          data: { message: "Password reset" },
          error: null,
        }),
      settings: {
        app: {
          get: vi.fn().mockResolvedValue({}),
          update: vi.fn().mockResolvedValue({}),
        },
        system: {
          list: vi.fn().mockResolvedValue({ settings: [] }),
          update: vi.fn().mockResolvedValue({}),
          delete: vi.fn().mockResolvedValue({}),
        },
      },
      management: {
        clientKeys: {
          list: vi.fn().mockResolvedValue({ client_keys: [] }),
          create: vi.fn().mockResolvedValue({ key: "new-key", client_key: {} }),
          update: vi.fn().mockResolvedValue({}),
          revoke: vi.fn().mockResolvedValue({}),
          delete: vi.fn().mockResolvedValue({}),
        },
        webhooks: {
          list: vi.fn().mockResolvedValue({ webhooks: [] }),
          create: vi.fn().mockResolvedValue({}),
          update: vi.fn().mockResolvedValue({}),
          delete: vi.fn().mockResolvedValue({}),
          test: vi.fn().mockResolvedValue({}),
        },
      },
      ...overrides.admin,
    },
    ...overrides,
  } as unknown as FluxbaseClient;
}

/**
 * Create a fresh QueryClient for testing
 */
export function createTestQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        gcTime: 0,
      },
      mutations: {
        retry: false,
      },
    },
  });
}

interface WrapperProps {
  children: React.ReactNode;
}

/**
 * Create a wrapper component with all providers
 */
export function createWrapper(
  client: FluxbaseClient,
  queryClient?: QueryClient,
) {
  const qc = queryClient || createTestQueryClient();

  return function Wrapper({ children }: WrapperProps) {
    return (
      <QueryClientProvider client={qc}>
        <FluxbaseProvider client={client}>{children}</FluxbaseProvider>
      </QueryClientProvider>
    );
  };
}

/**
 * Custom render function that includes all providers
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: Omit<RenderOptions, "wrapper"> & {
    client?: FluxbaseClient;
    queryClient?: QueryClient;
  },
) {
  const {
    client = createMockClient(),
    queryClient,
    ...renderOptions
  } = options || {};
  const wrapper = createWrapper(client, queryClient);

  return {
    ...render(ui, { wrapper, ...renderOptions }),
    client,
    queryClient: queryClient || createTestQueryClient(),
  };
}

// Re-export testing library utilities
export * from "@testing-library/react";
