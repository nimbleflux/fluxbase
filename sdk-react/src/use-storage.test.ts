/**
 * Tests for storage hooks
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor, act } from '@testing-library/react';
import {
  useStorageList,
  useStorageUpload,
  useStorageUploadWithProgress,
  useStorageDownload,
  useStorageDelete,
  useStoragePublicUrl,
  useStorageTransformUrl,
  useStorageSignedUrl,
  useStorageSignedUrlWithOptions,
  useStorageMove,
  useStorageCopy,
  useStorageBuckets,
  useCreateBucket,
  useDeleteBucket,
} from './use-storage';
import { createMockClient, createWrapper, createTestQueryClient } from './test-utils';

describe('useStorageList', () => {
  it('should list files in bucket', async () => {
    const mockFiles = [{ name: 'file1.txt' }, { name: 'file2.txt' }];
    const listMock = vi.fn().mockResolvedValue({ data: mockFiles, error: null });
    const fromMock = vi.fn().mockReturnValue({ list: listMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageList('bucket'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockFiles);
    expect(fromMock).toHaveBeenCalledWith('bucket');
  });

  it('should pass list options', async () => {
    const listMock = vi.fn().mockResolvedValue({ data: [], error: null });
    const fromMock = vi.fn().mockReturnValue({ list: listMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    renderHook(
      () => useStorageList('bucket', { prefix: 'folder/', limit: 10, offset: 5 }),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => {
      expect(listMock).toHaveBeenCalledWith({ prefix: 'folder/', limit: 10, offset: 5 });
    });
  });

  it('should throw error on list failure', async () => {
    const error = new Error('List failed');
    const listMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({ list: listMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageList('bucket'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.error).toBe(error);
  });
});

describe('useStorageUpload', () => {
  it('should upload file and invalidate queries', async () => {
    const mockResult = { path: 'file.txt' };
    const uploadMock = vi.fn().mockResolvedValue({ data: mockResult, error: null });
    const fromMock = vi.fn().mockReturnValue({ upload: uploadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useStorageUpload('bucket'), {
      wrapper: createWrapper(client, queryClient),
    });

    const file = new Blob(['content']);
    await act(async () => {
      await result.current.mutateAsync({ path: 'file.txt', file });
    });

    expect(fromMock).toHaveBeenCalledWith('bucket');
    expect(uploadMock).toHaveBeenCalledWith('file.txt', file, undefined);
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'bucket', 'list'] });
  });

  it('should throw error on upload failure', async () => {
    const error = new Error('Upload failed');
    const uploadMock = vi.fn().mockResolvedValue({ data: null, error });
    const fromMock = vi.fn().mockReturnValue({ upload: uploadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(() => useStorageUpload('bucket'), {
      wrapper: createWrapper(client),
    });

    const file = new Blob(['content']);
    await expect(act(async () => {
      await result.current.mutateAsync({ path: 'file.txt', file });
    })).rejects.toThrow('Upload failed');
  });
});

describe('useStorageUploadWithProgress', () => {
  it('should track upload progress', async () => {
    let progressCallback: Function | undefined;
    let resolveUpload: Function;

    const uploadMock = vi.fn().mockImplementation((path, file, options) => {
      progressCallback = options?.onUploadProgress;
      return new Promise((resolve) => {
        resolveUpload = () => resolve({ data: { path }, error: null });
      });
    });
    const fromMock = vi.fn().mockReturnValue({ upload: uploadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(() => useStorageUploadWithProgress('bucket'), {
      wrapper: createWrapper(client),
    });

    const file = new Blob(['content']);

    // Start upload (don't await yet)
    let uploadPromise: Promise<any>;
    act(() => {
      uploadPromise = result.current.upload.mutateAsync({ path: 'file.txt', file });
    });

    // Wait for upload to start and callback to be assigned
    await waitFor(() => {
      expect(progressCallback).toBeDefined();
    });

    // Simulate progress
    act(() => {
      progressCallback!({ loaded: 50, total: 100, percentage: 50 });
    });

    // Check progress state
    expect(result.current.progress).toEqual({ loaded: 50, total: 100, percentage: 50 });

    // Resolve upload
    await act(async () => {
      resolveUpload!();
      await uploadPromise;
    });
  });

  it('should reset progress on error', async () => {
    const uploadMock = vi.fn().mockResolvedValue({ data: null, error: new Error('Failed') });
    const fromMock = vi.fn().mockReturnValue({ upload: uploadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(() => useStorageUploadWithProgress('bucket'), {
      wrapper: createWrapper(client),
    });

    const file = new Blob(['content']);
    try {
      await act(async () => {
        await result.current.upload.mutateAsync({ path: 'file.txt', file });
      });
    } catch {
      // Expected error
    }

    expect(result.current.progress).toBeNull();
  });

  it('should have reset function', () => {
    const client = createMockClient();

    const { result } = renderHook(() => useStorageUploadWithProgress('bucket'), {
      wrapper: createWrapper(client),
    });

    expect(result.current.reset).toBeDefined();
    expect(typeof result.current.reset).toBe('function');
  });
});

describe('useStorageDownload', () => {
  it('should download file', async () => {
    const mockBlob = new Blob(['content']);
    const downloadMock = vi.fn().mockResolvedValue({ data: mockBlob, error: null });
    const fromMock = vi.fn().mockReturnValue({ download: downloadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageDownload('bucket', 'file.txt'),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBe(mockBlob);
    expect(downloadMock).toHaveBeenCalledWith('file.txt');
  });

  it('should not fetch when path is null', async () => {
    const downloadMock = vi.fn();
    const fromMock = vi.fn().mockReturnValue({ download: downloadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageDownload('bucket', null),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(downloadMock).not.toHaveBeenCalled();
  });

  it('should not fetch when disabled', async () => {
    const downloadMock = vi.fn();
    const fromMock = vi.fn().mockReturnValue({ download: downloadMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageDownload('bucket', 'file.txt', false),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(downloadMock).not.toHaveBeenCalled();
  });
});

describe('useStorageDelete', () => {
  it('should delete files and invalidate queries', async () => {
    const removeMock = vi.fn().mockResolvedValue({ error: null });
    const fromMock = vi.fn().mockReturnValue({ remove: removeMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useStorageDelete('bucket'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync(['file1.txt', 'file2.txt']);
    });

    expect(removeMock).toHaveBeenCalledWith(['file1.txt', 'file2.txt']);
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'bucket', 'list'] });
  });

  it('should throw error on delete failure', async () => {
    const error = new Error('Delete failed');
    const removeMock = vi.fn().mockResolvedValue({ error });
    const fromMock = vi.fn().mockReturnValue({ remove: removeMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(() => useStorageDelete('bucket'), {
      wrapper: createWrapper(client),
    });

    await expect(act(async () => {
      await result.current.mutateAsync(['file.txt']);
    })).rejects.toThrow('Delete failed');
  });
});

describe('useStoragePublicUrl', () => {
  it('should return public URL', () => {
    const getPublicUrlMock = vi.fn().mockReturnValue({ data: { publicUrl: 'http://example.com/file' } });
    const fromMock = vi.fn().mockReturnValue({ getPublicUrl: getPublicUrlMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStoragePublicUrl('bucket', 'file.txt'),
      { wrapper: createWrapper(client) }
    );

    expect(result.current).toBe('http://example.com/file');
    expect(getPublicUrlMock).toHaveBeenCalledWith('file.txt');
  });

  it('should return null when path is null', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useStoragePublicUrl('bucket', null),
      { wrapper: createWrapper(client) }
    );

    expect(result.current).toBeNull();
  });
});

describe('useStorageTransformUrl', () => {
  it('should return transform URL', () => {
    const getTransformUrlMock = vi.fn().mockReturnValue('http://example.com/transform/file');
    const fromMock = vi.fn().mockReturnValue({ getTransformUrl: getTransformUrlMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageTransformUrl('bucket', 'file.jpg', { width: 100, height: 100 }),
      { wrapper: createWrapper(client) }
    );

    expect(result.current).toBe('http://example.com/transform/file');
    expect(getTransformUrlMock).toHaveBeenCalledWith('file.jpg', { width: 100, height: 100 });
  });

  it('should return null when path is null', () => {
    const client = createMockClient();

    const { result } = renderHook(
      () => useStorageTransformUrl('bucket', null, { width: 100 }),
      { wrapper: createWrapper(client) }
    );

    expect(result.current).toBeNull();
  });
});

describe('useStorageSignedUrl', () => {
  it('should fetch signed URL', async () => {
    const createSignedUrlMock = vi.fn().mockResolvedValue({ data: { signedUrl: 'http://example.com/signed' }, error: null });
    const fromMock = vi.fn().mockReturnValue({ createSignedUrl: createSignedUrlMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageSignedUrl('bucket', 'file.txt', 3600),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBe('http://example.com/signed');
    expect(createSignedUrlMock).toHaveBeenCalledWith('file.txt', { expiresIn: 3600 });
  });

  it('should not fetch when path is null', async () => {
    const createSignedUrlMock = vi.fn();
    const fromMock = vi.fn().mockReturnValue({ createSignedUrl: createSignedUrlMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const { result } = renderHook(
      () => useStorageSignedUrl('bucket', null),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(createSignedUrlMock).not.toHaveBeenCalled();
  });
});

describe('useStorageSignedUrlWithOptions', () => {
  it('should fetch signed URL with transform options', async () => {
    const createSignedUrlMock = vi.fn().mockResolvedValue({ data: { signedUrl: 'http://example.com/signed' }, error: null });
    const fromMock = vi.fn().mockReturnValue({ createSignedUrl: createSignedUrlMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const options = {
      expiresIn: 3600,
      transform: { width: 100, height: 100 },
    };

    const { result } = renderHook(
      () => useStorageSignedUrlWithOptions('bucket', 'file.jpg', options),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toBe('http://example.com/signed');
    expect(createSignedUrlMock).toHaveBeenCalledWith('file.jpg', options);
  });
});

describe('useStorageMove', () => {
  it('should move file and invalidate queries', async () => {
    const moveMock = vi.fn().mockResolvedValue({ data: { path: 'new.txt' }, error: null });
    const fromMock = vi.fn().mockReturnValue({ move: moveMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useStorageMove('bucket'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ fromPath: 'old.txt', toPath: 'new.txt' });
    });

    expect(moveMock).toHaveBeenCalledWith('old.txt', 'new.txt');
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'bucket', 'list'] });
  });
});

describe('useStorageCopy', () => {
  it('should copy file and invalidate queries', async () => {
    const copyMock = vi.fn().mockResolvedValue({ data: { path: 'copy.txt' }, error: null });
    const fromMock = vi.fn().mockReturnValue({ copy: copyMock });

    const client = createMockClient({
      storage: { from: fromMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useStorageCopy('bucket'), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync({ fromPath: 'source.txt', toPath: 'copy.txt' });
    });

    expect(copyMock).toHaveBeenCalledWith('source.txt', 'copy.txt');
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'bucket', 'list'] });
  });
});

describe('useStorageBuckets', () => {
  it('should list buckets', async () => {
    const mockBuckets = [{ name: 'bucket1' }, { name: 'bucket2' }];
    const listBucketsMock = vi.fn().mockResolvedValue({ data: mockBuckets, error: null });

    const client = createMockClient({
      storage: { listBuckets: listBucketsMock },
    } as any);

    const { result } = renderHook(
      () => useStorageBuckets(),
      { wrapper: createWrapper(client) }
    );

    await waitFor(() => expect(result.current.isLoading).toBe(false));
    expect(result.current.data).toEqual(mockBuckets);
  });
});

describe('useCreateBucket', () => {
  it('should create bucket and invalidate queries', async () => {
    const createBucketMock = vi.fn().mockResolvedValue({ error: null });

    const client = createMockClient({
      storage: { createBucket: createBucketMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useCreateBucket(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync('new-bucket');
    });

    expect(createBucketMock).toHaveBeenCalledWith('new-bucket');
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'buckets'] });
  });
});

describe('useDeleteBucket', () => {
  it('should delete bucket and invalidate queries', async () => {
    const deleteBucketMock = vi.fn().mockResolvedValue({ error: null });

    const client = createMockClient({
      storage: { deleteBucket: deleteBucketMock },
    } as any);

    const queryClient = createTestQueryClient();
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries');

    const { result } = renderHook(() => useDeleteBucket(), {
      wrapper: createWrapper(client, queryClient),
    });

    await act(async () => {
      await result.current.mutateAsync('bucket-to-delete');
    });

    expect(deleteBucketMock).toHaveBeenCalledWith('bucket-to-delete');
    expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['fluxbase', 'storage', 'buckets'] });
  });
});
