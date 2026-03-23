package branching

import (
	"container/list"
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// poolEntry represents a connection pool with its last access time for LRU eviction
type poolEntry struct {
	slug       string
	pool       *pgxpool.Pool
	config     *pgxpool.Config
	lastAccess time.Time
	lruElement *list.Element // Pointer to the element in the LRU list
}

// Router manages connection pools for database branches
type Router struct {
	storage      *Storage
	config       config.BranchingConfig
	mainPool     *pgxpool.Pool
	mainDBURL    string
	pools        map[string]*poolEntry // slug -> pool entry
	poolsMu      sync.RWMutex
	lruList      *list.List   // LRU list of pools (least recently used at front)
	lruMu        sync.Mutex   // Separate mutex for LRU operations
	maxConns     int32        // Maximum total connections across all branch pools
	currentConns int32        // Current total connections
	activeBranch atomic.Value // Thread-safe active branch slug (set via API)
}

// NewRouter creates a new branch router
func NewRouter(storage *Storage, cfg config.BranchingConfig, mainPool *pgxpool.Pool, mainDBURL string) *Router {
	maxConns := int32(cfg.MaxTotalConnections)
	if maxConns <= 0 {
		maxConns = 500 // Default to 500 if not set
	}

	evictionAge := cfg.PoolEvictionAge
	if evictionAge <= 0 {
		evictionAge = time.Hour // Default to 1 hour if not set
	}

	r := &Router{
		storage:   storage,
		config:    cfg,
		mainPool:  mainPool,
		mainDBURL: mainDBURL,
		pools:     make(map[string]*poolEntry),
		lruList:   list.New(),
		maxConns:  maxConns,
	}
	// Initialize active branch to empty (not set via API yet)
	// Config default branch is used separately in GetDefaultBranch()
	r.activeBranch.Store("")

	// Start background eviction goroutine
	go r.evictIdlePools(evictionAge)

	return r
}

// GetPool returns the connection pool for a branch
// If the branch is "main" or empty, returns the main pool
func (r *Router) GetPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	// Empty or "main" slug uses the main pool
	if slug == "" || slug == "main" {
		return r.mainPool, nil
	}

	// Check if branching is enabled
	if !r.config.Enabled {
		return nil, ErrBranchingDisabled
	}

	// Check if we already have a pool for this branch
	r.poolsMu.RLock()
	entry, exists := r.pools[slug]
	r.poolsMu.RUnlock()

	if exists && entry != nil {
		// Update last access time and move to end of LRU list
		r.updateAccess(slug)
		return entry.pool, nil
	}

	// Need to create a new pool
	return r.createPoolForBranch(ctx, slug)
}

// updateAccess updates the last access time for a pool and moves it to the end of the LRU list
func (r *Router) updateAccess(slug string) {
	r.poolsMu.RLock()
	entry, exists := r.pools[slug]
	r.poolsMu.RUnlock()

	if exists && entry != nil {
		r.lruMu.Lock()
		entry.lastAccess = time.Now()
		// Move to end of LRU list (most recently used)
		if entry.lruElement != nil {
			r.lruList.MoveToBack(entry.lruElement)
		}
		r.lruMu.Unlock()
	}
}

// createPoolForBranch creates a new connection pool for a branch
func (r *Router) createPoolForBranch(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	r.poolsMu.Lock()
	defer r.poolsMu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := r.pools[slug]; exists && entry != nil {
		r.updateAccess(slug)
		return entry.pool, nil
	}

	// Check if we would exceed global connection limit
	if r.getCurrentTotalConns() >= r.maxConns {
		// Try to evict idle pools to free up connections
		if !r.evictLRUPool() {
			return nil, fmt.Errorf("global branch connection limit reached (%d), cannot create new pool", r.maxConns)
		}
	}

	// Get branch from storage
	branch, err := r.storage.GetBranchBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Check if branch is ready
	if branch.Status != BranchStatusReady {
		return nil, ErrBranchNotReady
	}

	// Create connection URL for branch database
	connURL, err := r.getBranchConnectionURL(branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch connection URL: %w", err)
	}

	// Parse pool config
	poolConfig, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pool config: %w", err)
	}

	// Configure pool settings (smaller pools for branch databases)
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	// Create the pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping branch database: %w", err)
	}

	// Create pool entry and add to LRU list
	r.lruMu.Lock()
	entry := &poolEntry{
		slug:       slug,
		pool:       pool,
		config:     poolConfig,
		lastAccess: time.Now(),
	}
	entry.lruElement = r.lruList.PushBack(entry)
	r.lruMu.Unlock()

	// Store the pool entry
	r.pools[slug] = entry

	// Update current connection count
	atomic.AddInt32(&r.currentConns, poolConfig.MaxConns)

	log.Info().
		Str("branch_slug", slug).
		Str("database", branch.DatabaseName).
		Int32("max_conns", poolConfig.MaxConns).
		Int32("total_conns", r.getCurrentTotalConns()).
		Int32("max_total_conns", r.maxConns).
		Msg("Created connection pool for branch")

	return pool, nil
}

// getBranchConnectionURL returns the connection URL for a branch database
func (r *Router) getBranchConnectionURL(branch *Branch) (string, error) {
	// Parse the main database URL
	parsedURL, err := url.Parse(r.mainDBURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse main database URL: %w", err)
	}

	// Replace the database name
	parsedURL.Path = "/" + branch.DatabaseName

	return parsedURL.String(), nil
}

// ClosePool closes and removes the pool for a branch
func (r *Router) ClosePool(slug string) {
	r.poolsMu.Lock()
	defer r.poolsMu.Unlock()

	if entry, exists := r.pools[slug]; exists {
		r.closePoolEntry(entry)
		delete(r.pools, slug)

		log.Info().
			Str("branch_slug", slug).
			Msg("Closed connection pool for branch")
	}
}

// closePoolEntry closes a pool entry and updates connection count
func (r *Router) closePoolEntry(entry *poolEntry) {
	// Remove from LRU list
	r.lruMu.Lock()
	if entry.lruElement != nil {
		r.lruList.Remove(entry.lruElement)
	}
	r.lruMu.Unlock()

	// Update current connection count
	if entry.config != nil {
		atomic.AddInt32(&r.currentConns, -entry.config.MaxConns)
	}

	// Close the pool
	if entry.pool != nil {
		entry.pool.Close()
	}
}

// CloseAllPools closes all branch pools (called during shutdown)
func (r *Router) CloseAllPools() {
	r.poolsMu.Lock()
	defer r.poolsMu.Unlock()

	for slug, entry := range r.pools {
		r.closePoolEntry(entry)
		log.Debug().
			Str("branch_slug", slug).
			Msg("Closed connection pool for branch")
	}

	r.pools = make(map[string]*poolEntry)
	r.lruList.Init()
}

// RefreshPool recreates the pool for a branch (e.g., after migration)
func (r *Router) RefreshPool(ctx context.Context, slug string) error {
	// Close existing pool
	r.ClosePool(slug)

	// Create new pool
	_, err := r.createPoolForBranch(ctx, slug)
	return err
}

// GetActivePools returns the list of active branch slugs
func (r *Router) GetActivePools() []string {
	r.poolsMu.RLock()
	defer r.poolsMu.RUnlock()

	slugs := make([]string, 0, len(r.pools))
	for slug := range r.pools {
		slugs = append(slugs, slug)
	}
	return slugs
}

// GetPoolStats returns statistics for all pools
func (r *Router) GetPoolStats() map[string]PoolStats {
	r.poolsMu.RLock()
	defer r.poolsMu.RUnlock()

	stats := make(map[string]PoolStats)

	// Add main pool stats
	mainStat := r.mainPool.Stat()
	stats["main"] = PoolStats{
		TotalConns:      mainStat.TotalConns(),
		IdleConns:       mainStat.IdleConns(),
		AcquiredConns:   mainStat.AcquiredConns(),
		MaxConns:        mainStat.MaxConns(),
		AcquireCount:    mainStat.AcquireCount(),
		AcquireDuration: mainStat.AcquireDuration(),
	}

	// Add branch pool stats
	for slug, entry := range r.pools {
		stat := entry.pool.Stat()
		stats[slug] = PoolStats{
			TotalConns:      stat.TotalConns(),
			IdleConns:       stat.IdleConns(),
			AcquiredConns:   stat.AcquiredConns(),
			MaxConns:        stat.MaxConns(),
			AcquireCount:    stat.AcquireCount(),
			AcquireDuration: stat.AcquireDuration(),
		}
	}

	return stats
}

// getCurrentTotalConns returns the current total connections across all branch pools
func (r *Router) getCurrentTotalConns() int32 {
	return atomic.LoadInt32(&r.currentConns)
}

// evictLRUPool evicts the least recently used pool to free up connections
// Returns true if a pool was evicted, false if no pools can be evicted
func (r *Router) evictLRUPool() bool {
	r.lruMu.Lock()
	defer r.lruMu.Unlock()

	// Get the least recently used element (front of list)
	if r.lruList.Len() == 0 {
		return false
	}

	lruElement := r.lruList.Front()
	if lruElement == nil {
		return false
	}

	entry, ok := lruElement.Value.(*poolEntry)
	if !ok || entry == nil {
		return false
	}

	// Close the pool
	r.poolsMu.Lock()
	r.closePoolEntry(entry)
	delete(r.pools, entry.slug)
	r.poolsMu.Unlock()

	log.Info().
		Str("branch_slug", entry.slug).
		Int32("freed_conns", entry.config.MaxConns).
		Int32("total_conns", r.getCurrentTotalConns()).
		Msg("Evicted LRU branch pool to free connections")

	return true
}

// evictIdlePools runs in the background to evict pools that haven't been accessed recently
func (r *Router) evictIdlePools(evictionAge time.Duration) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		r.poolsMu.RLock()
		poolsCopy := make(map[string]*poolEntry, len(r.pools))
		for k, v := range r.pools {
			poolsCopy[k] = v
		}
		r.poolsMu.RUnlock()

		now := time.Now()
		for slug, entry := range poolsCopy {
			if now.Sub(entry.lastAccess) > evictionAge {
				r.poolsMu.Lock()
				// Double check the pool still exists and hasn't been accessed recently
				if currentEntry, exists := r.pools[slug]; exists && now.Sub(currentEntry.lastAccess) > evictionAge {
					r.closePoolEntry(currentEntry)
					delete(r.pools, slug)
					log.Info().
						Str("branch_slug", slug).
						Dur("idle_time", now.Sub(currentEntry.lastAccess)).
						Msg("Evicted idle branch pool")
				}
				r.poolsMu.Unlock()
			}
		}
	}
}

// PoolStats contains connection pool statistics
type PoolStats struct {
	TotalConns      int32         `json:"total_conns"`
	IdleConns       int32         `json:"idle_conns"`
	AcquiredConns   int32         `json:"acquired_conns"`
	MaxConns        int32         `json:"max_conns"`
	AcquireCount    int64         `json:"acquire_count"`
	AcquireDuration time.Duration `json:"acquire_duration"`
}

// IsMainBranch checks if a slug refers to the main branch
func IsMainBranch(slug string) bool {
	return slug == "" || slug == "main"
}

// GetMainPool returns the main database pool
func (r *Router) GetMainPool() *pgxpool.Pool {
	return r.mainPool
}

// HasPool checks if a pool exists for the given branch
func (r *Router) HasPool(slug string) bool {
	if IsMainBranch(slug) {
		return true
	}

	r.poolsMu.RLock()
	defer r.poolsMu.RUnlock()

	_, exists := r.pools[slug]
	return exists
}

// WarmupPool pre-creates a connection pool for a branch
// This is useful after branch creation to ensure the pool is ready
func (r *Router) WarmupPool(ctx context.Context, slug string) error {
	_, err := r.GetPool(ctx, slug)
	return err
}

// GetStorage returns the storage instance
func (r *Router) GetStorage() *Storage {
	return r.storage
}

// SetActiveBranch sets the server-wide active branch (via API)
// Pass empty string to reset to config default
func (r *Router) SetActiveBranch(slug string) {
	r.activeBranch.Store(slug)
	if slug == "" {
		log.Info().Msg("Active branch reset to default")
	} else {
		log.Info().Str("branch", slug).Msg("Active branch set")
	}
}

// GetActiveBranch returns the current API-set active branch
// Returns empty string if not set via API
func (r *Router) GetActiveBranch() string {
	if v := r.activeBranch.Load(); v != nil {
		return v.(string)
	}
	return ""
}

// GetDefaultBranch returns the effective default branch considering all sources
// Precedence: API-set > Config > "main"
func (r *Router) GetDefaultBranch() string {
	// First check API-set active branch
	if active := r.GetActiveBranch(); active != "" {
		return active
	}
	// Then check config default
	if r.config.DefaultBranch != "" {
		return r.config.DefaultBranch
	}
	// Fall back to "main"
	return "main"
}

// GetActiveBranchSource returns the source of the current default branch
// Returns "api", "config", or "default"
func (r *Router) GetActiveBranchSource() string {
	if active := r.GetActiveBranch(); active != "" {
		return "api"
	}
	if r.config.DefaultBranch != "" {
		return "config"
	}
	return "default"
}
