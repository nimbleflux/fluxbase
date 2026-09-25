package branching

import (
	"container/list"
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/config"
)

// poolEntry represents a connection pool with its last access time for LRU eviction
type poolEntry struct {
	key        string // composite pool key (tenantID + slug)
	slug       string
	tenantID   string // "" for instance-level branches
	pool       *pgxpool.Pool
	config     *pgxpool.Config
	lastAccess time.Time
	lruElement *list.Element // Pointer to the element in the LRU list
}

// poolKey builds the composite cache key for a branch pool.
// Branch slugs are only unique per tenant (UNIQUE(slug, tenant_id)), so pools
// must be keyed by tenant ID and slug. An empty tenantID denotes instance-level
// branches (branches.tenant_id IS NULL).
func poolKey(tenantID, slug string) string {
	return tenantID + "\x00" + slug
}

// Router manages connection pools for database branches
type Router struct {
	storage      *Storage
	config       config.BranchingConfig
	mainPool     *pgxpool.Pool
	mainDBURL    string
	pools        map[string]*poolEntry // composite key (tenantID+slug) -> pool entry
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
// Legacy instance-level lookup: resolves the branch without a tenant filter.
// Request-path callers should use GetPoolForBranch with the resolved tenant ID.
func (r *Router) GetPool(ctx context.Context, slug string) (*pgxpool.Pool, error) {
	return r.GetPoolForBranch(ctx, "", slug)
}

// GetPoolForBranch returns the connection pool for a branch scoped to a tenant.
// Branch slugs are unique per tenant only, so pools are cached under the
// composite (tenantID, slug) key. An empty tenantID keeps the legacy behavior
// of resolving the branch without a tenant filter (instance-level branches).
// If the branch is "main" or empty, returns the main pool.
func (r *Router) GetPoolForBranch(ctx context.Context, tenantID, slug string) (*pgxpool.Pool, error) {
	// Empty or "main" slug uses the main pool
	if slug == "" || slug == "main" {
		return r.mainPool, nil
	}

	// Check if branching is enabled
	if !r.config.Enabled {
		return nil, ErrBranchingDisabled
	}

	key := poolKey(tenantID, slug)

	// Check if we already have a pool for this branch
	r.poolsMu.RLock()
	entry, exists := r.pools[key]
	r.poolsMu.RUnlock()

	if exists && entry != nil {
		// Update last access time and move to end of LRU list
		r.updateAccess(key)
		return entry.pool, nil
	}

	// Need to create a new pool
	return r.createPoolForBranch(ctx, key, tenantID, slug)
}

// updateAccess updates the last access time for a pool and moves it to the end of the LRU list
func (r *Router) updateAccess(key string) {
	r.poolsMu.RLock()
	entry, exists := r.pools[key]
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

// resolveBranch finds the branch for a slug scoped to a tenant. When a tenant
// ID is provided, tenant-scoped branches take precedence; the instance-level
// (NULL tenant) fallback then only accepts branches that are not owned by a
// different tenant. An empty tenantID keeps the legacy global lookup.
func (r *Router) resolveBranch(ctx context.Context, tenantID, slug string) (*Branch, error) {
	if tenantID != "" {
		if tid, err := uuid.Parse(tenantID); err == nil {
			branch, err := r.storage.GetBranchBySlug(ctx, slug, &tid)
			if err == nil {
				return branch, nil
			}
			if !errors.Is(err, ErrBranchNotFound) {
				return nil, err
			}
		}
	}

	branch, err := r.storage.GetBranchBySlug(ctx, slug, nil)
	if err != nil {
		return nil, err
	}
	if branch.TenantID != nil && tenantID != "" && branch.TenantID.String() != tenantID {
		// The slug belongs to a different tenant's branch.
		return nil, ErrBranchNotFound
	}
	return branch, nil
}

// UserHasAccessForBranch reports whether a user has access to the branch with
// the given slug in the given tenant scope. Instance-level (NULL tenant)
// branches are shared, so they are resolved with the same fallback rules as
// GetPoolForBranch.
func (r *Router) UserHasAccessForBranch(ctx context.Context, tenantID, slug string, userID string) (bool, error) {
	branch, err := r.resolveBranch(ctx, tenantID, slug)
	if err != nil {
		if errors.Is(err, ErrBranchNotFound) {
			return false, nil
		}
		return false, err
	}

	// Main branch is accessible to all authenticated users
	if branch.Type == BranchTypeMain {
		return true, nil
	}

	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, nil
	}
	return r.storage.HasAccess(ctx, branch.ID, uid, BranchAccessRead)
}

// createPoolForBranch creates a new connection pool for a branch
func (r *Router) createPoolForBranch(ctx context.Context, key, tenantID, slug string) (*pgxpool.Pool, error) {
	r.poolsMu.Lock()
	defer r.poolsMu.Unlock()

	// Double-check after acquiring write lock
	if entry, exists := r.pools[key]; exists && entry != nil {
		// Update recency inline: updateAccess would re-lock poolsMu (already
		// held here) and self-deadlock.
		entry.lastAccess = time.Now()
		r.lruMu.Lock()
		if entry.lruElement != nil {
			r.lruList.MoveToBack(entry.lruElement)
		}
		r.lruMu.Unlock()
		return entry.pool, nil
	}

	// Check if we would exceed global connection limit
	if r.getCurrentTotalConns() >= r.maxConns {
		// Try to evict idle pools to free up connections
		if !r.evictLRUPool() {
			return nil, fmt.Errorf("global branch connection limit reached (%d), cannot create new pool", r.maxConns)
		}
	}

	// Get branch from storage, scoped to the resolved tenant when available
	branch, err := r.resolveBranch(ctx, tenantID, slug)
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
		key:        key,
		slug:       slug,
		tenantID:   tenantID,
		pool:       pool,
		config:     poolConfig,
		lastAccess: time.Now(),
	}
	entry.lruElement = r.lruList.PushBack(entry)
	r.lruMu.Unlock()

	// Store the pool entry
	r.pools[key] = entry

	// Update current connection count
	atomic.AddInt32(&r.currentConns, poolConfig.MaxConns)

	log.Info().
		Str("branch_slug", slug).
		Str("tenant_id", tenantID).
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

// ClosePool closes and removes the pools for a branch slug.
// Slugs are unique per tenant, so every tenant-scoped pool cached for the slug
// is closed (branch deletion is tenant-scoped; closing an unrelated same-slug
// pool in another tenant is safe — the pool is recreated on next use).
func (r *Router) ClosePool(slug string) {
	r.poolsMu.Lock()
	defer r.poolsMu.Unlock()

	for key, entry := range r.pools {
		if entry != nil && entry.slug == slug {
			r.closePoolEntry(entry)
			delete(r.pools, key)

			log.Info().
				Str("branch_slug", slug).
				Str("tenant_id", entry.tenantID).
				Msg("Closed connection pool for branch")
		}
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
	// Close existing pool(s) for the slug across all tenant scopes
	r.ClosePool(slug)

	// Resolve the branch to recreate the pool under its own tenant scope
	tenantID := ""
	if branch, err := r.storage.GetBranchBySlug(ctx, slug, nil); err == nil && branch.TenantID != nil {
		tenantID = branch.TenantID.String()
	}

	// Create new pool
	_, err := r.createPoolForBranch(ctx, poolKey(tenantID, slug), tenantID, slug)
	return err
}

// GetActivePools returns the list of active branch slugs
func (r *Router) GetActivePools() []string {
	r.poolsMu.RLock()
	defer r.poolsMu.RUnlock()

	slugs := make([]string, 0, len(r.pools))
	for _, entry := range r.pools {
		if entry != nil {
			slugs = append(slugs, entry.slug)
		}
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

// evictLRUPool evicts the least recently used pool to free up connections.
// The caller must hold r.poolsMu (createPoolForBranch holds it), so this
// closes the pool and updates accounting without re-acquiring poolsMu.
// Returns true if a pool was evicted, false if no pools can be evicted
func (r *Router) evictLRUPool() bool {
	r.lruMu.Lock()
	// Get the least recently used element (front of list)
	if r.lruList.Len() == 0 {
		r.lruMu.Unlock()
		return false
	}

	lruElement := r.lruList.Front()
	if lruElement == nil {
		r.lruMu.Unlock()
		return false
	}

	entry, ok := lruElement.Value.(*poolEntry)
	if !ok || entry == nil {
		r.lruMu.Unlock()
		return false
	}

	// Remove from the LRU list while holding lruMu
	r.lruList.Remove(lruElement)
	entry.lruElement = nil
	r.lruMu.Unlock()

	// Close the pool and update accounting (poolsMu is held by the caller)
	if entry.config != nil {
		atomic.AddInt32(&r.currentConns, -entry.config.MaxConns)
	}
	if entry.pool != nil {
		entry.pool.Close()
	}
	delete(r.pools, entry.key)

	log.Info().
		Str("branch_slug", entry.slug).
		Str("tenant_id", entry.tenantID).
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
		for key, entry := range poolsCopy {
			if now.Sub(entry.lastAccess) > evictionAge {
				r.poolsMu.Lock()
				// Double check the pool still exists and hasn't been accessed recently
				if currentEntry, exists := r.pools[key]; exists && now.Sub(currentEntry.lastAccess) > evictionAge {
					r.closePoolEntry(currentEntry)
					delete(r.pools, key)
					log.Info().
						Str("branch_slug", currentEntry.slug).
						Str("tenant_id", currentEntry.tenantID).
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

// HasPool checks if a pool exists for the given branch slug in any tenant scope
func (r *Router) HasPool(slug string) bool {
	if IsMainBranch(slug) {
		return true
	}

	r.poolsMu.RLock()
	defer r.poolsMu.RUnlock()

	for _, entry := range r.pools {
		if entry != nil && entry.slug == slug {
			return true
		}
	}
	return false
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
