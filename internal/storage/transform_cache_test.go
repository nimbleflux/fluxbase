package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCacheProvider implements Provider for transform cache testing
type mockCacheProvider struct {
	mu      sync.RWMutex
	objects map[string]map[string]mockObject
	buckets map[string]bool

	// For controlling test behavior
	uploadErr   error
	downloadErr error
	deleteErr   error
}

type mockObject struct {
	data        []byte
	contentType string
	lastMod     time.Time
}

func newMockCacheProvider() *mockCacheProvider {
	return &mockCacheProvider{
		objects: make(map[string]map[string]mockObject),
		buckets: make(map[string]bool),
	}
}

func (m *mockCacheProvider) Name() string { return "mock-cache" }

func (m *mockCacheProvider) Health(ctx context.Context) error { return nil }

func (m *mockCacheProvider) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, opts *UploadOptions) (*Object, error) {
	if m.uploadErr != nil {
		return nil, m.uploadErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.objects[bucket]; !exists {
		m.objects[bucket] = make(map[string]mockObject)
	}

	content, _ := io.ReadAll(data)
	contentType := "application/octet-stream"
	if opts != nil && opts.ContentType != "" {
		contentType = opts.ContentType
	}

	m.objects[bucket][key] = mockObject{
		data:        content,
		contentType: contentType,
		lastMod:     time.Now(),
	}

	return &Object{
		Key:          key,
		Bucket:       bucket,
		Size:         int64(len(content)),
		ContentType:  contentType,
		LastModified: time.Now(),
	}, nil
}

func (m *mockCacheProvider) Download(ctx context.Context, bucket, key string, opts *DownloadOptions) (io.ReadCloser, *Object, error) {
	if m.downloadErr != nil {
		return nil, nil, m.downloadErr
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if bucketData, exists := m.objects[bucket]; exists {
		if obj, exists := bucketData[key]; exists {
			return io.NopCloser(bytes.NewReader(obj.data)), &Object{
				Key:          key,
				Bucket:       bucket,
				Size:         int64(len(obj.data)),
				ContentType:  obj.contentType,
				LastModified: obj.lastMod,
			}, nil
		}
	}
	return nil, nil, ErrTransformFailed
}

func (m *mockCacheProvider) Delete(ctx context.Context, bucket, key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bucketData, exists := m.objects[bucket]; exists {
		delete(bucketData, key)
	}
	return nil
}

func (m *mockCacheProvider) Exists(ctx context.Context, bucket, key string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if bucketData, exists := m.objects[bucket]; exists {
		_, exists := bucketData[key]
		return exists, nil
	}
	return false, nil
}

func (m *mockCacheProvider) GetObject(ctx context.Context, bucket, key string) (*Object, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if bucketData, exists := m.objects[bucket]; exists {
		if obj, exists := bucketData[key]; exists {
			return &Object{
				Key:         key,
				Bucket:      bucket,
				Size:        int64(len(obj.data)),
				ContentType: obj.contentType,
			}, nil
		}
	}
	return nil, ErrTransformFailed
}

func (m *mockCacheProvider) List(ctx context.Context, bucket string, opts *ListOptions) (*ListResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var objects []Object
	if bucketData, exists := m.objects[bucket]; exists {
		for key, obj := range bucketData {
			objects = append(objects, Object{
				Key:          key,
				Bucket:       bucket,
				Size:         int64(len(obj.data)),
				ContentType:  obj.contentType,
				LastModified: obj.lastMod,
			})
		}
	}
	return &ListResult{Objects: objects}, nil
}

func (m *mockCacheProvider) CreateBucket(ctx context.Context, bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.buckets[bucket] = true
	m.objects[bucket] = make(map[string]mockObject)
	return nil
}

func (m *mockCacheProvider) DeleteBucket(ctx context.Context, bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.buckets, bucket)
	delete(m.objects, bucket)
	return nil
}

func (m *mockCacheProvider) BucketExists(ctx context.Context, bucket string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.buckets[bucket], nil
}

func (m *mockCacheProvider) ListBuckets(ctx context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var buckets []string
	for bucket := range m.buckets {
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}

func (m *mockCacheProvider) GenerateSignedURL(ctx context.Context, bucket, key string, opts *SignedURLOptions) (string, error) {
	return "", nil
}

func (m *mockCacheProvider) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

func (m *mockCacheProvider) MoveObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

// =============================================================================
// NewTransformCache Tests
// =============================================================================

func TestNewTransformCache(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	cache, err := NewTransformCache(ctx, provider, TransformCacheOptions{
		TTL:     1 * time.Hour,
		MaxSize: 100 * 1024 * 1024, // 100MB
	})

	require.NoError(t, err)
	assert.NotNil(t, cache)
	assert.Equal(t, 1*time.Hour, cache.ttl)
	assert.Equal(t, int64(100*1024*1024), cache.maxSize)
}

func TestNewTransformCache_CreatesBucket(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	// Bucket should not exist initially
	exists, _ := provider.BucketExists(ctx, TransformCacheBucket)
	assert.False(t, exists)

	_, err := NewTransformCache(ctx, provider, TransformCacheOptions{})
	require.NoError(t, err)

	// Bucket should be created
	exists, _ = provider.BucketExists(ctx, TransformCacheBucket)
	assert.True(t, exists)
}

func TestNewTransformCache_DefaultOptions(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	cache, err := NewTransformCache(ctx, provider, TransformCacheOptions{})

	require.NoError(t, err)
	assert.Equal(t, 24*time.Hour, cache.ttl)              // Default TTL
	assert.Equal(t, int64(1024*1024*1024), cache.maxSize) // Default 1GB
}

func TestNewTransformCache_ExistingBucket(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	// Pre-create the bucket
	_ = provider.CreateBucket(ctx, TransformCacheBucket)

	cache, err := NewTransformCache(ctx, provider, TransformCacheOptions{})

	require.NoError(t, err)
	assert.NotNil(t, cache)
}

func TestNewTransformCache_LoadsExistingEntries(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	// Pre-create bucket and add some cached entries
	_ = provider.CreateBucket(ctx, TransformCacheBucket)
	_, _ = provider.Upload(ctx, TransformCacheBucket, "cached1", bytes.NewReader([]byte("data1")), 5, nil)
	_, _ = provider.Upload(ctx, TransformCacheBucket, "cached2", bytes.NewReader([]byte("data22")), 6, nil)

	cache, err := NewTransformCache(ctx, provider, TransformCacheOptions{})

	require.NoError(t, err)
	// Cache should have loaded the existing entries
	size, count, _ := cache.Stats()
	assert.Equal(t, 2, count)
	assert.Equal(t, int64(11), size) // 5 + 6 bytes
}

// =============================================================================
// cacheKey Tests
// =============================================================================

func TestTransformCache_cacheKey(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{
		Width:   800,
		Height:  600,
		Format:  "webp",
		Quality: 80,
		Fit:     FitCover,
	}

	key := cache.cacheKey("my-bucket", "images/photo.jpg", opts)

	// Should be a valid hash
	assert.Len(t, key, 64) // SHA256 hex = 64 chars
	assert.NotEmpty(t, key)
}

func TestTransformCache_cacheKey_Deterministic(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{
		Width:   800,
		Height:  600,
		Format:  "webp",
		Quality: 80,
		Fit:     FitCover,
	}

	key1 := cache.cacheKey("bucket", "key", opts)
	key2 := cache.cacheKey("bucket", "key", opts)

	assert.Equal(t, key1, key2)
}

func TestTransformCache_cacheKey_DifferentOptions(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts1 := &TransformOptions{Width: 800, Height: 600}
	opts2 := &TransformOptions{Width: 400, Height: 300}

	key1 := cache.cacheKey("bucket", "key", opts1)
	key2 := cache.cacheKey("bucket", "key", opts2)

	assert.NotEqual(t, key1, key2)
}

// =============================================================================
// Get Tests
// =============================================================================

func TestTransformCache_Get_CacheHit(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{TTL: 1 * time.Hour})

	opts := &TransformOptions{Width: 800, Height: 600, Format: "webp"}
	testData := []byte("transformed image data")

	// Set cache entry
	err := cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")
	require.NoError(t, err)

	// Get should return cached data
	data, contentType, ok := cache.Get(ctx, "bucket", "image.jpg", opts)

	assert.True(t, ok)
	assert.Equal(t, testData, data)
	assert.Equal(t, "image/webp", contentType)
}

func TestTransformCache_Get_CacheMiss(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800, Height: 600}

	data, contentType, ok := cache.Get(ctx, "bucket", "nonexistent.jpg", opts)

	assert.False(t, ok)
	assert.Nil(t, data)
	assert.Empty(t, contentType)
}

func TestTransformCache_Get_Expired(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{TTL: 1 * time.Millisecond})

	opts := &TransformOptions{Width: 800, Height: 600}
	testData := []byte("transformed image data")

	// Set cache entry
	err := cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")
	require.NoError(t, err)

	// Wait for TTL to expire
	time.Sleep(10 * time.Millisecond)

	// Get should return miss due to expiration
	data, contentType, ok := cache.Get(ctx, "bucket", "image.jpg", opts)

	assert.False(t, ok)
	assert.Nil(t, data)
	assert.Empty(t, contentType)
}

func TestTransformCache_Get_UpdatesAccessTime(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{TTL: 1 * time.Hour})

	opts := &TransformOptions{Width: 800, Height: 600}
	testData := []byte("transformed image data")

	// Set cache entry
	_ = cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")

	cacheKey := cache.cacheKey("bucket", "image.jpg", opts)
	cache.mu.RLock()
	initialTime := cache.entries[cacheKey].accessTime
	cache.mu.RUnlock()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Access the cache
	_, _, _ = cache.Get(ctx, "bucket", "image.jpg", opts)

	cache.mu.RLock()
	newTime := cache.entries[cacheKey].accessTime
	cache.mu.RUnlock()

	assert.True(t, newTime.After(initialTime))
}

// =============================================================================
// Set Tests
// =============================================================================

func TestTransformCache_Set(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800, Height: 600, Format: "webp"}
	testData := []byte("transformed image data")

	err := cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")

	require.NoError(t, err)

	// Verify entry was added
	size, count, _ := cache.Stats()
	assert.Equal(t, 1, count)
	assert.Equal(t, int64(len(testData)), size)

	// Verify we can get it back
	data, contentType, ok := cache.Get(ctx, "bucket", "image.jpg", opts)
	assert.True(t, ok)
	assert.Equal(t, testData, data)
	assert.Equal(t, "image/webp", contentType)
}

func TestTransformCache_Set_StoresMetadata(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800, Height: 600}
	testData := []byte("image data")

	_ = cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")

	// Check that metadata file was created
	cacheKey := cache.cacheKey("bucket", "image.jpg", opts)
	metaReader, _, err := provider.Download(ctx, TransformCacheBucket, cacheKey+".meta", nil)
	require.NoError(t, err)

	var meta cacheEntryMeta
	err = json.NewDecoder(metaReader).Decode(&meta)
	_ = metaReader.Close()

	require.NoError(t, err)
	assert.Equal(t, "image/webp", meta.ContentType)
	assert.Equal(t, int64(len(testData)), meta.Size)
	assert.Equal(t, "bucket/image.jpg", meta.SourceKey)
}

func TestTransformCache_Set_TriggersEviction(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()

	// Small max size to trigger eviction
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{
		MaxSize: 100, // 100 bytes
	})

	// Fill cache with entries
	for i := 0; i < 5; i++ {
		opts := &TransformOptions{Width: i * 100}
		data := make([]byte, 30) // 30 bytes each
		_ = cache.Set(ctx, "bucket", "image.jpg", opts, data, "image/webp")
		time.Sleep(time.Millisecond) // Ensure different access times
	}

	// Cache should have evicted older entries to stay under max size
	size, _, _ := cache.Stats()
	assert.LessOrEqual(t, size, int64(100))
}

// =============================================================================
// Eviction Tests
// =============================================================================

func TestTransformCache_evictUntilSize(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{
		MaxSize: 1000,
	})

	// Add multiple entries with staggered times
	for i := 0; i < 5; i++ {
		opts := &TransformOptions{Width: i * 100}
		data := make([]byte, 100) // 100 bytes each
		_ = cache.Set(ctx, "bucket", "image.jpg", opts, data, "image/webp")
		time.Sleep(time.Millisecond) // Ensure different access times
	}

	// Verify initial state
	size, count, _ := cache.Stats()
	assert.Equal(t, 5, count)
	assert.Equal(t, int64(500), size)

	// Trigger eviction to reach target 200 bytes
	cache.mu.Lock()
	victims := cache.evictVictimsUntilSize(200)
	cache.mu.Unlock()
	assert.NotEmpty(t, victims)

	// Check eviction worked
	size, _, _ = cache.Stats()
	assert.LessOrEqual(t, size, int64(200))
}

func TestTransformCache_evictEntry(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800}
	testData := []byte("test data")

	_ = cache.Set(ctx, "bucket", "image.jpg", opts, testData, "image/webp")

	// Verify entry exists
	_, _, ok := cache.Get(ctx, "bucket", "image.jpg", opts)
	assert.True(t, ok)

	// Evict the entry
	cacheKey := cache.cacheKey("bucket", "image.jpg", opts)
	cache.evictEntry(ctx, cacheKey)

	// Verify entry is gone
	_, _, ok = cache.Get(ctx, "bucket", "image.jpg", opts)
	assert.False(t, ok)

	// Verify size is updated
	size, count, _ := cache.Stats()
	assert.Equal(t, 0, count)
	assert.Equal(t, int64(0), size)
}

// =============================================================================
// Invalidate Tests
// =============================================================================

func TestTransformCache_Invalidate(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	// Add multiple transforms for the same source file
	opts1 := &TransformOptions{Width: 800}
	opts2 := &TransformOptions{Width: 400}
	opts3 := &TransformOptions{Width: 200}

	_ = cache.Set(ctx, "bucket", "image.jpg", opts1, []byte("data1"), "image/webp")
	_ = cache.Set(ctx, "bucket", "image.jpg", opts2, []byte("data2"), "image/webp")
	_ = cache.Set(ctx, "bucket", "image.jpg", opts3, []byte("data3"), "image/webp")

	// Add transform for different source
	_ = cache.Set(ctx, "bucket", "other.jpg", opts1, []byte("other"), "image/webp")

	// Invalidate all transforms for image.jpg
	err := cache.Invalidate(ctx, "bucket", "image.jpg")
	require.NoError(t, err)

	// Verify image.jpg transforms are gone
	_, _, ok := cache.Get(ctx, "bucket", "image.jpg", opts1)
	assert.False(t, ok)
	_, _, ok = cache.Get(ctx, "bucket", "image.jpg", opts2)
	assert.False(t, ok)
	_, _, ok = cache.Get(ctx, "bucket", "image.jpg", opts3)
	assert.False(t, ok)

	// Verify other.jpg transform is still there
	_, _, ok = cache.Get(ctx, "bucket", "other.jpg", opts1)
	assert.True(t, ok)
}

// =============================================================================
// Cleanup Tests
// =============================================================================

func TestTransformCache_Cleanup(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{
		TTL: 10 * time.Millisecond,
	})

	// Add entries
	opts1 := &TransformOptions{Width: 800}
	opts2 := &TransformOptions{Width: 400}

	_ = cache.Set(ctx, "bucket", "image1.jpg", opts1, []byte("data1"), "image/webp")
	_ = cache.Set(ctx, "bucket", "image2.jpg", opts2, []byte("data2"), "image/webp")

	// Wait for TTL to expire
	time.Sleep(20 * time.Millisecond)

	// Run cleanup
	cache.Cleanup(ctx)

	// All entries should be cleaned up
	_, count, _ := cache.Stats()
	assert.Equal(t, 0, count)
}

func TestTransformCache_Cleanup_KeepsValidEntries(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{
		TTL: 1 * time.Hour,
	})

	// Add entries
	opts := &TransformOptions{Width: 800}
	_ = cache.Set(ctx, "bucket", "image.jpg", opts, []byte("data"), "image/webp")

	// Run cleanup immediately (entries should not expire)
	cache.Cleanup(ctx)

	// Entry should still exist
	_, count, _ := cache.Stats()
	assert.Equal(t, 1, count)
}

// =============================================================================
// Stats Tests
// =============================================================================

func TestTransformCache_Stats(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{
		MaxSize: 1000,
	})

	// Initially empty
	size, count, maxSize := cache.Stats()
	assert.Equal(t, int64(0), size)
	assert.Equal(t, 0, count)
	assert.Equal(t, int64(1000), maxSize)

	// Add entries
	opts1 := &TransformOptions{Width: 800}
	opts2 := &TransformOptions{Width: 400}

	_ = cache.Set(ctx, "bucket", "image1.jpg", opts1, []byte("12345"), "image/webp")     // 5 bytes
	_ = cache.Set(ctx, "bucket", "image2.jpg", opts2, []byte("1234567890"), "image/png") // 10 bytes

	size, count, maxSize = cache.Stats()
	assert.Equal(t, int64(15), size)
	assert.Equal(t, 2, count)
	assert.Equal(t, int64(1000), maxSize)
}

// =============================================================================
// Clear Tests
// =============================================================================

func TestTransformCache_Clear(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	// Add entries
	opts1 := &TransformOptions{Width: 800}
	opts2 := &TransformOptions{Width: 400}

	_ = cache.Set(ctx, "bucket", "image1.jpg", opts1, []byte("data1"), "image/webp")
	_ = cache.Set(ctx, "bucket", "image2.jpg", opts2, []byte("data2"), "image/webp")

	// Clear the cache
	err := cache.Clear(ctx)
	require.NoError(t, err)

	// Verify cache is empty
	size, count, _ := cache.Stats()
	assert.Equal(t, 0, count)
	assert.Equal(t, int64(0), size)

	// Verify entries are gone
	_, _, ok := cache.Get(ctx, "bucket", "image1.jpg", opts1)
	assert.False(t, ok)
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestTransformCache_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	var wg sync.WaitGroup
	numGoroutines := 10
	numOps := 100

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				opts := &TransformOptions{Width: id*1000 + j}
				_ = cache.Set(ctx, "bucket", "image.jpg", opts, []byte("data"), "image/webp")
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				opts := &TransformOptions{Width: id*1000 + j}
				_, _, _ = cache.Get(ctx, "bucket", "image.jpg", opts)
			}
		}(i)
	}

	wg.Wait()

	// Should complete without deadlock or panic
	_, count, _ := cache.Stats()
	assert.Greater(t, count, 0)
}

// =============================================================================
// Constants and Types Tests
// =============================================================================

func TestTransformCacheBucket_Constant(t *testing.T) {
	assert.Equal(t, "_transform_cache", TransformCacheBucket)
}

func TestCacheEntryMeta_Fields(t *testing.T) {
	meta := cacheEntryMeta{
		ContentType: "image/webp",
		Size:        1024,
		SourceKey:   "bucket/image.jpg",
		AccessTime:  time.Now(),
		CreatedAt:   time.Now(),
	}

	assert.Equal(t, "image/webp", meta.ContentType)
	assert.Equal(t, int64(1024), meta.Size)
	assert.Equal(t, "bucket/image.jpg", meta.SourceKey)
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkTransformCache_Set(b *testing.B) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{MaxSize: 1024 * 1024 * 1024})

	opts := &TransformOptions{Width: 800, Height: 600, Format: "webp"}
	data := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.Set(ctx, "bucket", "image.jpg", opts, data, "image/webp")
	}
}

func BenchmarkTransformCache_Get(b *testing.B) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800, Height: 600, Format: "webp"}
	data := make([]byte, 1024)
	_ = cache.Set(ctx, "bucket", "image.jpg", opts, data, "image/webp")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = cache.Get(ctx, "bucket", "image.jpg", opts)
	}
}

func BenchmarkTransformCache_cacheKey(b *testing.B) {
	ctx := context.Background()
	provider := newMockCacheProvider()
	cache, _ := NewTransformCache(ctx, provider, TransformCacheOptions{})

	opts := &TransformOptions{Width: 800, Height: 600, Format: "webp", Quality: 80, Fit: FitCover}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.cacheKey("my-bucket", "images/photo.jpg", opts)
	}
}
