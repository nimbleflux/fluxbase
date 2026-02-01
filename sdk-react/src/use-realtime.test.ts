/**
 * Tests for realtime subscription hooks
 */

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useRealtime,
  useTableSubscription,
  useTableInserts,
  useTableUpdates,
  useTableDeletes,
} from './use-realtime';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useRealtime', () => {
  it('should create channel and subscribe', () => {
    const onMock = vi.fn().mockReturnThis();
    const subscribeMock = vi.fn().mockReturnThis();
    const unsubscribeMock = vi.fn();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: subscribeMock,
      unsubscribe: unsubscribeMock,
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useRealtime({ channel: 'table:products' }),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
    expect(onMock).toHaveBeenCalledWith('*', expect.any(Function));
    expect(subscribeMock).toHaveBeenCalled();
  });

  it('should unsubscribe on unmount', () => {
    const onMock = vi.fn().mockReturnThis();
    const subscribeMock = vi.fn().mockReturnThis();
    const unsubscribeMock = vi.fn();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: subscribeMock,
      unsubscribe: unsubscribeMock,
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const { unmount } = renderHook(
      () => useRealtime({ channel: 'table:products' }),
      { wrapper: createWrapper(client) }
    );

    unmount();

    expect(unsubscribeMock).toHaveBeenCalled();
  });

  it('should not subscribe when disabled', () => {
    const channelMock = vi.fn();
    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useRealtime({ channel: 'table:products', enabled: false }),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).not.toHaveBeenCalled();
  });

  it('should subscribe to specific event type', () => {
    const onMock = vi.fn().mockReturnThis();
    const subscribeMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: subscribeMock,
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useRealtime({ channel: 'table:products', event: 'INSERT' }),
      { wrapper: createWrapper(client) }
    );

    expect(onMock).toHaveBeenCalledWith('INSERT', expect.any(Function));
  });

  it('should call callback on change', () => {
    const callback = vi.fn();
    let changeHandler: Function;
    const onMock = vi.fn().mockImplementation((event, handler) => {
      changeHandler = handler;
      return { subscribe: vi.fn().mockReturnThis(), unsubscribe: vi.fn() };
    });
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useRealtime({ channel: 'table:products', callback }),
      { wrapper: createWrapper(client) }
    );

    const payload = { eventType: 'INSERT', new: { id: 1 }, old: null };
    changeHandler!(payload);

    expect(callback).toHaveBeenCalledWith(payload);
  });

  it('should auto-invalidate queries when enabled', () => {
    let changeHandler: Function;
    const onMock = vi.fn().mockImplementation((event, handler) => {
      changeHandler = handler;
      return { subscribe: vi.fn().mockReturnThis(), unsubscribe: vi.fn() };
    });
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    renderHook(
      () => useRealtime({ channel: 'table:public.products', autoInvalidate: true }),
      { wrapper: createWrapper(client, queryClient) }
    );

    const payload = { eventType: 'INSERT', new: { id: 1 }, old: null };
    changeHandler!(payload);

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'table', 'public.products'] });
  });

  it('should use custom invalidate key', () => {
    let changeHandler: Function;
    const onMock = vi.fn().mockImplementation((event, handler) => {
      changeHandler = handler;
      return { subscribe: vi.fn().mockReturnThis(), unsubscribe: vi.fn() };
    });
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    renderHook(
      () => useRealtime({
        channel: 'table:products',
        autoInvalidate: true,
        invalidateKey: ['custom', 'key'],
      }),
      { wrapper: createWrapper(client, queryClient) }
    );

    const payload = { eventType: 'INSERT', new: { id: 1 }, old: null };
    changeHandler!(payload);

    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['custom', 'key'] });
  });

  it('should not auto-invalidate when disabled', () => {
    let changeHandler: Function;
    const onMock = vi.fn().mockImplementation((event, handler) => {
      changeHandler = handler;
      return { subscribe: vi.fn().mockReturnThis(), unsubscribe: vi.fn() };
    });
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    renderHook(
      () => useRealtime({ channel: 'table:products', autoInvalidate: false }),
      { wrapper: createWrapper(client, queryClient) }
    );

    const payload = { eventType: 'INSERT', new: { id: 1 }, old: null };
    changeHandler!(payload);

    expect(invalidateSpy).not.toHaveBeenCalled();
  });

  it('should return channel property', () => {
    const mockChannel = {
      on: vi.fn().mockReturnThis(),
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    };
    const channelMock = vi.fn().mockReturnValue(mockChannel);

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const { result } = renderHook(
      () => useRealtime({ channel: 'table:products' }),
      { wrapper: createWrapper(client) }
    );

    // The hook returns a channel property (may be null initially, but is set by useEffect)
    expect(result.current).toHaveProperty('channel');
    // Verify the channel was created
    expect(channelMock).toHaveBeenCalledWith('table:products');
  });
});

describe('useTableSubscription', () => {
  it('should subscribe to table with correct channel name', () => {
    const onMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useTableSubscription('products'),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
  });

  it('should pass options through', () => {
    const callback = vi.fn();
    const onMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    renderHook(
      () => useTableSubscription('products', { callback, autoInvalidate: false }),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
  });
});

describe('useTableInserts', () => {
  it('should subscribe to INSERT events', () => {
    const onMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const callback = vi.fn();
    renderHook(
      () => useTableInserts('products', callback),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
    expect(onMock).toHaveBeenCalledWith('INSERT', expect.any(Function));
  });
});

describe('useTableUpdates', () => {
  it('should subscribe to UPDATE events', () => {
    const onMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const callback = vi.fn();
    renderHook(
      () => useTableUpdates('products', callback),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
    expect(onMock).toHaveBeenCalledWith('UPDATE', expect.any(Function));
  });
});

describe('useTableDeletes', () => {
  it('should subscribe to DELETE events', () => {
    const onMock = vi.fn().mockReturnThis();
    const channelMock = vi.fn().mockReturnValue({
      on: onMock,
      subscribe: vi.fn().mockReturnThis(),
      unsubscribe: vi.fn(),
    });

    const client = createMockClient({
      realtime: { channel: channelMock },
    } as any);

    const callback = vi.fn();
    renderHook(
      () => useTableDeletes('products', callback),
      { wrapper: createWrapper(client) }
    );

    expect(channelMock).toHaveBeenCalledWith('table:products');
    expect(onMock).toHaveBeenCalledWith('DELETE', expect.any(Function));
  });
});
