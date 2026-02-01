/**
 * Tests for admin hooks (settings, webhooks)
 */

import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, waitFor, act } from "@testing-library/react";
import {
  useAppSettings,
  useSystemSettings,
  useWebhooks,
} from "./use-admin-hooks";
import { createMockClient, createWrapper } from "./test-utils";

describe("useAppSettings", () => {
  it("should fetch settings on mount when autoFetch is true", async () => {
    const mockSettings = { features: { darkMode: true } };
    const getMock = vi.fn().mockResolvedValue(mockSettings);

    const client = createMockClient({
      admin: {
        settings: {
          app: { get: getMock, update: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(() => useAppSettings({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.settings).toEqual(mockSettings);
    expect(getMock).toHaveBeenCalled();
  });

  it("should not fetch settings on mount when autoFetch is false", async () => {
    const getMock = vi.fn();

    const client = createMockClient({
      admin: {
        settings: {
          app: { get: getMock, update: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(() => useAppSettings({ autoFetch: false }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(getMock).not.toHaveBeenCalled();
  });

  it("should update settings", async () => {
    const mockSettings = { features: { darkMode: true } };
    const getMock = vi.fn().mockResolvedValue(mockSettings);
    const updateMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        settings: {
          app: { get: getMock, update: updateMock },
        },
      },
    } as any);

    const { result } = renderHook(() => useAppSettings({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.updateSettings({
        features: { enable_realtime: false },
      });
    });

    expect(updateMock).toHaveBeenCalledWith({
      features: { enable_realtime: false },
    });
    // Should refetch after update
    expect(getMock).toHaveBeenCalledTimes(2);
  });

  it("should handle errors", async () => {
    const error = new Error("Failed to fetch");
    const getMock = vi.fn().mockRejectedValue(error);

    const client = createMockClient({
      admin: {
        settings: {
          app: { get: getMock, update: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(() => useAppSettings({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.error).toBe(error);
  });

  it("should refetch on demand", async () => {
    const mockSettings = { features: {} };
    const getMock = vi.fn().mockResolvedValue(mockSettings);

    const client = createMockClient({
      admin: {
        settings: {
          app: { get: getMock, update: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(() => useAppSettings({ autoFetch: false }), {
      wrapper: createWrapper(client),
    });

    await act(async () => {
      await result.current.refetch();
    });

    expect(getMock).toHaveBeenCalledTimes(1);
  });
});

describe("useSystemSettings", () => {
  it("should fetch settings on mount when autoFetch is true", async () => {
    const mockSettings = [{ key: "theme", value: "dark" }];
    const listMock = vi.fn().mockResolvedValue({ settings: mockSettings });

    const client = createMockClient({
      admin: {
        settings: {
          system: { list: listMock, update: vi.fn(), delete: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useSystemSettings({ autoFetch: true }),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.settings).toEqual(mockSettings);
  });

  it("should get setting by key", async () => {
    const mockSettings = [
      { key: "theme", value: "dark" },
      { key: "language", value: "en" },
    ];
    const listMock = vi.fn().mockResolvedValue({ settings: mockSettings });

    const client = createMockClient({
      admin: {
        settings: {
          system: { list: listMock, update: vi.fn(), delete: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useSystemSettings({ autoFetch: true }),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    const setting = result.current.getSetting("theme");
    expect(setting).toEqual({ key: "theme", value: "dark" });

    const notFound = result.current.getSetting("nonexistent");
    expect(notFound).toBeUndefined();
  });

  it("should update setting", async () => {
    const mockSettings = [{ key: "theme", value: "dark" }];
    const listMock = vi.fn().mockResolvedValue({ settings: mockSettings });
    const updateMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        settings: {
          system: { list: listMock, update: updateMock, delete: vi.fn() },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useSystemSettings({ autoFetch: true }),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.updateSetting("theme", {
        value: { theme: "light" },
      });
    });

    expect(updateMock).toHaveBeenCalledWith("theme", {
      value: { theme: "light" },
    });
    // Should refetch after update
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it("should delete setting", async () => {
    const mockSettings = [{ key: "theme", value: "dark" }];
    const listMock = vi.fn().mockResolvedValue({ settings: mockSettings });
    const deleteMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        settings: {
          system: { list: listMock, update: vi.fn(), delete: deleteMock },
        },
      },
    } as any);

    const { result } = renderHook(
      () => useSystemSettings({ autoFetch: true }),
      { wrapper: createWrapper(client) },
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.deleteSetting("theme");
    });

    expect(deleteMock).toHaveBeenCalledWith("theme");
    // Should refetch after delete
    expect(listMock).toHaveBeenCalledTimes(2);
  });
});

describe("useWebhooks", () => {
  it("should fetch webhooks on mount when autoFetch is true", async () => {
    const mockWebhooks = [{ id: "1", url: "https://example.com/webhook" }];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            delete: vi.fn(),
            test: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(() => useWebhooks({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.webhooks).toEqual(mockWebhooks);
  });

  it("should create webhook", async () => {
    const mockWebhooks: any[] = [];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });
    const createMock = vi
      .fn()
      .mockResolvedValue({ id: "1", url: "https://example.com/webhook" });

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: createMock,
            update: vi.fn(),
            delete: vi.fn(),
            test: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(() => useWebhooks({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.createWebhook({
        url: "https://example.com/webhook",
        events: ["user.created"],
      });
    });

    expect(createMock).toHaveBeenCalledWith({
      url: "https://example.com/webhook",
      events: ["user.created"],
    });
    // Should refetch after create
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it("should update webhook", async () => {
    const mockWebhooks = [{ id: "1", url: "https://example.com/webhook" }];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });
    const updateMock = vi
      .fn()
      .mockResolvedValue({ id: "1", url: "https://new.com/webhook" });

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: vi.fn(),
            update: updateMock,
            delete: vi.fn(),
            test: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(() => useWebhooks({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.updateWebhook("1", {
        url: "https://new.com/webhook",
      });
    });

    expect(updateMock).toHaveBeenCalledWith("1", {
      url: "https://new.com/webhook",
    });
    // Should refetch after update
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it("should delete webhook", async () => {
    const mockWebhooks = [{ id: "1", url: "https://example.com/webhook" }];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });
    const deleteMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            delete: deleteMock,
            test: vi.fn(),
          },
        },
      },
    } as any);

    const { result } = renderHook(() => useWebhooks({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.deleteWebhook("1");
    });

    expect(deleteMock).toHaveBeenCalledWith("1");
    // Should refetch after delete
    expect(listMock).toHaveBeenCalledTimes(2);
  });

  it("should test webhook", async () => {
    const mockWebhooks = [{ id: "1", url: "https://example.com/webhook" }];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });
    const testMock = vi.fn().mockResolvedValue({});

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            delete: vi.fn(),
            test: testMock,
          },
        },
      },
    } as any);

    const { result } = renderHook(() => useWebhooks({ autoFetch: true }), {
      wrapper: createWrapper(client),
    });

    await waitFor(() => expect(result.current.isLoading).toBe(false));

    await act(async () => {
      await result.current.testWebhook("1");
    });

    expect(testMock).toHaveBeenCalledWith("1");
  });

  it("should set up refetch interval", async () => {
    vi.useFakeTimers();
    const mockWebhooks: any[] = [];
    const listMock = vi.fn().mockResolvedValue({ webhooks: mockWebhooks });

    const client = createMockClient({
      admin: {
        management: {
          webhooks: {
            list: listMock,
            create: vi.fn(),
            update: vi.fn(),
            delete: vi.fn(),
            test: vi.fn(),
          },
        },
      },
    } as any);

    const { unmount } = renderHook(
      () => useWebhooks({ autoFetch: true, refetchInterval: 5000 }),
      { wrapper: createWrapper(client) },
    );

    // Initial fetch
    expect(listMock).toHaveBeenCalledTimes(1);

    // Advance timer
    await act(async () => {
      vi.advanceTimersByTime(5000);
    });

    expect(listMock).toHaveBeenCalledTimes(2);

    unmount();
    vi.useRealTimers();
  });
});
