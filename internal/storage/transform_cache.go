package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// TransformCacheBucket is the internal bucket name for transform cache
const TransformCacheBucket = "_transform_cache"

// cacheEntryMeta stores metadata for a cached transform
type cacheEntryMeta struct {
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	SourceKey   string    `json:"source_key"`
	AccessTime  time.Time `json:"access_time"`
	CreatedAt   time.Time `json:"created_at"`
}

// cacheEntry represents an in-memory cache entry for LRU tracking
type cacheEntry struct {
	key        string
	sourceKey  string // "bucket/key" of the source object (index cleanup)
	size       int64
	accessTime time.Time
}

// TransformCache provides caching for transformed images
// It uses a dedicated bucket in the storage provider and implements LRU eviction
type TransformCache struct {
	provider    Provider
	ttl         time.Duration
	maxSize     int64
	mu          sync.RWMutex
	currentSize int64
	entries     map[string]*cacheEntry // key -> entry for LRU tracking

	// index maps sourceKey ("bucket/key") -> set of cache keys, maintained on
	// Set and on startup load, so Invalidate does not have to list/download
	// .meta objects while holding the lock. Bounded by maxIndexSources.
	index           map[string]map[string]struct{}
	indexOrder      []string // source keys in insertion order (FIFO eviction)
	maxIndexSources int
}

// maxIndexSources bounds the source->keys index (memory protection).
const maxIndexSources = 10000

// TransformCacheOptions configures the transform cache
type TransformCacheOptions struct {
	TTL     time.Duration // Cache entry TTL (default: 24 hours)
	MaxSize int64         // Max cache size in bytes (default: 1GB)
}

// NewTransformCache creates a new transform cache using the storage provider
func NewTransformCache(ctx context.Context, provider Provider, opts TransformCacheOptions) (*TransformCache, error) {
	if opts.TTL <= 0 {
		opts.TTL = 24 * time.Hour
	}
	if opts.MaxSize <= 0 {
		opts.MaxSize = 1024 * 1024 * 1024 // 1GB default
	}

	cache := &TransformCache{
		provider:        provider,
		ttl:             opts.TTL,
		maxSize:         opts.MaxSize,
		entries:         make(map[string]*cacheEntry),
		index:           make(map[string]map[string]struct{}),
		maxIndexSources: maxIndexSources,
	}

	// Ensure cache bucket exists
	exists, err := provider.BucketExists(ctx, TransformCacheBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check cache bucket: %w", err)
	}

	if !exists {
		if err := provider.CreateBucket(ctx, TransformCacheBucket); err != nil {
			return nil, fmt.Errorf("failed to create cache bucket: %w", err)
		}
		log.Info().Str("bucket", TransformCacheBucket).Msg("Transform cache bucket created")
	}

	// Load existing cache entries
	if err := cache.loadExistingEntries(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to load existing cache entries")
	}

	return cache, nil
}

// loadExistingEntries scans the cache bucket and populates the entries map.
// The source->keys index is populated best-effort from the .meta objects so
// Invalidate can use it after a restart. Provider I/O happens without the
// lock held (the cache is not yet shared at construction time).
func (c *TransformCache) loadExistingEntries(ctx context.Context) error {
	result, err := c.provider.List(ctx, TransformCacheBucket, &ListOptions{MaxKeys: 10000})
	if err != nil {
		return err
	}

	for _, obj := range result.Objects {
		// Skip metadata files
		if len(obj.Key) > 5 && obj.Key[len(obj.Key)-5:] == ".meta" {
			continue
		}

		// Read the entry's .meta object (best effort) to restore its source
		// key for the index.
		sourceKey := ""
		metaReader, _, err := c.provider.Download(ctx, TransformCacheBucket, obj.Key+".meta", nil)
		if err == nil {
			var meta cacheEntryMeta
			if err := json.NewDecoder(metaReader).Decode(&meta); err == nil {
				sourceKey = meta.SourceKey
			}
			_ = metaReader.Close()
		}

		c.entries[obj.Key] = &cacheEntry{
			key:        obj.Key,
			sourceKey:  sourceKey,
			size:       obj.Size,
			accessTime: obj.LastModified,
		}
		c.currentSize += obj.Size

		if sourceKey != "" {
			c.indexSourceKeyUnlocked(sourceKey, obj.Key)
		}
	}

	log.Info().
		Int64("size", c.currentSize).
		Int("entries", len(c.entries)).
		Msg("Transform cache loaded")

	return nil
}

// cacheKey generates a cache key from bucket, key, and transform options
func (c *TransformCache) cacheKey(bucket, key string, opts *TransformOptions) string {
	data := fmt.Sprintf("%s/%s:%d:%d:%s:%d:%s",
		bucket, key, opts.Width, opts.Height, opts.Format, opts.Quality, opts.Fit)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached transform if it exists and is not expired
func (c *TransformCache) Get(ctx context.Context, bucket, key string, opts *TransformOptions) ([]byte, string, bool) {
	cacheKey := c.cacheKey(bucket, key, opts)

	c.mu.RLock()
	entry, exists := c.entries[cacheKey]
	c.mu.RUnlock()

	if !exists {
		return nil, "", false
	}

	// Check TTL expiration
	if time.Since(entry.accessTime) > c.ttl {
		c.evictEntry(ctx, cacheKey)
		return nil, "", false
	}

	// Download cached data
	reader, obj, err := c.provider.Download(ctx, TransformCacheBucket, cacheKey, nil)
	if err != nil {
		c.evictEntry(ctx, cacheKey)
		return nil, "", false
	}
	defer func() { _ = reader.Close() }()

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", false
	}

	// Load metadata
	contentType := obj.ContentType
	metaReader, _, err := c.provider.Download(ctx, TransformCacheBucket, cacheKey+".meta", nil)
	if err == nil {
		defer func() { _ = metaReader.Close() }()
		var meta cacheEntryMeta
		if err := json.NewDecoder(metaReader).Decode(&meta); err == nil {
			contentType = meta.ContentType
		}
	}

	// Update access time for LRU tracking
	c.mu.Lock()
	if e, ok := c.entries[cacheKey]; ok {
		e.accessTime = time.Now()
	}
	c.mu.Unlock()

	return data, contentType, true
}

// Set stores a transformed image in the cache. Eviction planning and
// bookkeeping happen under the lock; all provider I/O happens outside it.
func (c *TransformCache) Set(ctx context.Context, bucket, key string, opts *TransformOptions, data []byte, contentType string) error {
	newSize := int64(len(data))
	cacheKey := c.cacheKey(bucket, key, opts)
	sourceKey := fmt.Sprintf("%s/%s", bucket, key)

	// Phase 1 (locked): decide eviction victims and take them out of the
	// in-memory accounting.
	c.mu.Lock()
	var victims []string
	if c.currentSize+newSize > c.maxSize {
		victims = c.evictVictimsUntilSize(int64(float64(c.maxSize)*0.8) - newSize)
	}
	c.mu.Unlock()

	// Phase 2 (unlocked): best-effort provider deletes for the victims.
	for _, victim := range victims {
		c.deleteProviderObject(ctx, victim)
	}

	// Phase 3 (unlocked): upload the cached data and metadata.
	_, err := c.provider.Upload(ctx, TransformCacheBucket, cacheKey, bytes.NewReader(data), newSize, &UploadOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to cache transform: %w", err)
	}

	meta := cacheEntryMeta{
		ContentType: contentType,
		Size:        newSize,
		SourceKey:   sourceKey,
		AccessTime:  time.Now(),
		CreatedAt:   time.Now(),
	}
	metaData, _ := json.Marshal(meta)
	_, err = c.provider.Upload(ctx, TransformCacheBucket, cacheKey+".meta", bytes.NewReader(metaData), int64(len(metaData)), &UploadOptions{
		ContentType: "application/json",
	})
	if err != nil {
		// Clean up the cached data if metadata upload fails
		_ = c.provider.Delete(ctx, TransformCacheBucket, cacheKey)
		return fmt.Errorf("failed to cache transform metadata: %w", err)
	}

	// Phase 4 (locked): record the new entry.
	c.mu.Lock()
	c.entries[cacheKey] = &cacheEntry{
		key:        cacheKey,
		sourceKey:  sourceKey,
		size:       newSize,
		accessTime: time.Now(),
	}
	c.currentSize += newSize
	c.indexSourceKeyUnlocked(sourceKey, cacheKey)
	c.mu.Unlock()

	return nil
}

// evictVictimsUntilSize removes entries from in-memory accounting (oldest
// first) until currentSize <= targetSize and returns their keys for the
// caller to delete from the provider without the lock held.
// Must be called with c.mu held.
func (c *TransformCache) evictVictimsUntilSize(targetSize int64) []string {
	if targetSize < 0 {
		targetSize = 0
	}

	// Sort entries by access time (oldest first)
	sortedEntries := make([]*cacheEntry, 0, len(c.entries))
	for _, entry := range c.entries {
		sortedEntries = append(sortedEntries, entry)
	}
	sort.Slice(sortedEntries, func(i, j int) bool {
		return sortedEntries[i].accessTime.Before(sortedEntries[j].accessTime)
	})

	var victims []string
	for _, entry := range sortedEntries {
		if c.currentSize <= targetSize {
			break
		}
		if c.removeEntryUnlocked(entry.key) {
			victims = append(victims, entry.key)
		}
	}

	if len(victims) > 0 {
		log.Debug().
			Int("evicted", len(victims)).
			Int64("current_size", c.currentSize).
			Msg("Cache eviction completed")
	}
	return victims
}

// removeEntryUnlocked drops an entry from in-memory accounting (entries map,
// size counter and source index) and reports whether it existed.
// Must be called with c.mu held. Provider deletion is the caller's job.
func (c *TransformCache) removeEntryUnlocked(cacheKey string) bool {
	entry, exists := c.entries[cacheKey]
	if !exists {
		return false
	}

	c.currentSize -= entry.size
	delete(c.entries, cacheKey)

	// Drop the key from its source index bucket.
	if sourceKey := entry.sourceKey; sourceKey != "" {
		if keys, ok := c.index[sourceKey]; ok {
			delete(keys, cacheKey)
			if len(keys) == 0 {
				delete(c.index, sourceKey)
				c.removeFromIndexOrder(sourceKey)
			}
		}
	}
	return true
}

// indexSourceKeyUnlocked records sourceKey -> cacheKey in the bounded index.
// Must be called with c.mu held.
func (c *TransformCache) indexSourceKeyUnlocked(sourceKey, cacheKey string) {
	if sourceKey == "" {
		return
	}
	keys, ok := c.index[sourceKey]
	if !ok {
		if len(c.index) >= c.maxIndexSources {
			// FIFO eviction of the oldest source to keep the index bounded.
			if len(c.indexOrder) > 0 {
				oldest := c.indexOrder[0]
				c.indexOrder = c.indexOrder[1:]
				delete(c.index, oldest)
			}
		}
		keys = make(map[string]struct{})
		c.index[sourceKey] = keys
		c.indexOrder = append(c.indexOrder, sourceKey)
	}
	keys[cacheKey] = struct{}{}
}

// removeFromIndexOrder drops a source key from the FIFO order slice.
func (c *TransformCache) removeFromIndexOrder(sourceKey string) {
	for i, k := range c.indexOrder {
		if k == sourceKey {
			c.indexOrder = append(c.indexOrder[:i], c.indexOrder[i+1:]...)
			return
		}
	}
}

// deleteProviderObject removes a cached object and its .meta from the
// provider. Best effort; called without the lock held.
func (c *TransformCache) deleteProviderObject(ctx context.Context, cacheKey string) {
	_ = c.provider.Delete(ctx, TransformCacheBucket, cacheKey)
	_ = c.provider.Delete(ctx, TransformCacheBucket, cacheKey+".meta")
}

// evictEntry removes a cache entry from the index and the provider.
func (c *TransformCache) evictEntry(ctx context.Context, cacheKey string) {
	c.mu.Lock()
	_ = c.removeEntryUnlocked(cacheKey)
	c.mu.Unlock()
	c.deleteProviderObject(ctx, cacheKey)
}

// Invalidate removes all cached transforms for a source file
// This is called when the source file is updated or deleted.
// Provider I/O happens outside the lock: the victim keys come from the
// in-memory source index (populated on Set and startup load), with a
// .meta-scan fallback for entries that were never indexed.
func (c *TransformCache) Invalidate(ctx context.Context, bucket, key string) error {
	sourceKey := fmt.Sprintf("%s/%s", bucket, key)

	// Phase 1 (locked): collect victims from the index.
	c.mu.Lock()
	var victims []string
	if keys, ok := c.index[sourceKey]; ok {
		for cacheKey := range keys {
			if c.removeEntryUnlocked(cacheKey) {
				victims = append(victims, cacheKey)
			}
		}
		delete(c.index, sourceKey)
		c.removeFromIndexOrder(sourceKey)
	}
	c.mu.Unlock()

	// Phase 2 (unlocked): fallback scan for entries missing from the index
	// (e.g. written by an older process before the index existed).
	c.mu.RLock()
	_, indexed := c.index[sourceKey]
	c.mu.RUnlock()
	if !indexed {
		scanned, err := c.scanSourceKeys(ctx, sourceKey)
		if err != nil {
			return err
		}
		c.mu.Lock()
		for _, cacheKey := range scanned {
			if c.removeEntryUnlocked(cacheKey) {
				victims = append(victims, cacheKey)
			}
		}
		c.mu.Unlock()
	}

	// Phase 3 (unlocked): provider deletes.
	for _, cacheKey := range victims {
		c.deleteProviderObject(ctx, cacheKey)
	}

	if len(victims) > 0 {
		log.Debug().
			Str("source", sourceKey).
			Int("evicted", len(victims)).
			Msg("Invalidated cache entries for source file")
	}

	return nil
}

// scanSourceKeys lists the cache bucket and decodes every .meta object to
// find cache keys belonging to sourceKey. Provider I/O only — no locks held.
func (c *TransformCache) scanSourceKeys(ctx context.Context, sourceKey string) ([]string, error) {
	result, err := c.provider.List(ctx, TransformCacheBucket, &ListOptions{MaxKeys: 10000})
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, obj := range result.Objects {
		// Only check .meta files
		if len(obj.Key) < 5 || obj.Key[len(obj.Key)-5:] != ".meta" {
			continue
		}

		reader, _, err := c.provider.Download(ctx, TransformCacheBucket, obj.Key, nil)
		if err != nil {
			continue
		}

		var meta cacheEntryMeta
		if err := json.NewDecoder(reader).Decode(&meta); err != nil {
			_ = reader.Close()
			continue
		}
		_ = reader.Close()

		if meta.SourceKey == sourceKey {
			// Remove .meta suffix to get the cache key
			matches = append(matches, obj.Key[:len(obj.Key)-5])
		}
	}
	return matches, nil
}

// Cleanup removes expired entries (call periodically). Provider deletes run
// outside the lock.
func (c *TransformCache) Cleanup(ctx context.Context) {
	c.mu.Lock()
	now := time.Now()
	var expired []string
	for key, entry := range c.entries {
		if now.Sub(entry.accessTime) > c.ttl {
			expired = append(expired, key)
		}
	}
	for _, key := range expired {
		c.removeEntryUnlocked(key)
	}
	c.mu.Unlock()

	for _, key := range expired {
		c.deleteProviderObject(ctx, key)
	}

	if len(expired) > 0 {
		log.Debug().
			Int("evicted", len(expired)).
			Msg("Cleanup removed expired cache entries")
	}
}

// Stats returns cache statistics
func (c *TransformCache) Stats() (currentSize int64, entryCount int, maxSize int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentSize, len(c.entries), c.maxSize
}

// Clear removes all cache entries. Keys are collected and the accounting
// reset under the lock; provider deletes run outside it.
func (c *TransformCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	keys := make([]string, 0, len(c.entries))
	for key := range c.entries {
		keys = append(keys, key)
	}
	c.entries = make(map[string]*cacheEntry)
	c.index = make(map[string]map[string]struct{})
	c.indexOrder = nil
	c.currentSize = 0
	c.mu.Unlock()

	for _, key := range keys {
		c.deleteProviderObject(ctx, key)
	}

	log.Info().Msg("Transform cache cleared")
	return nil
}
