package api

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/storage"
)

type StorageModule struct {
	Handlers *StorageHandlers
	Manager  *storage.Manager
	Service  *storage.Service

	// cancel stops the background maintenance goroutines started in Init.
	cancel context.CancelFunc
}

func (m *StorageModule) Name() string { return "storage" }

func (m *StorageModule) Init(ctx context.Context, registry *ServiceRegistry) error {
	cfg := registry.Config
	db := registry.DB

	storageManager, err := storage.NewManager(&cfg.Storage, cfg.GetPublicBaseURL(), cfg.Auth.JWTSecret, registry.Metrics)
	if err != nil {
		return err
	}
	m.Manager = storageManager
	m.Service = storageManager.GetBaseService()

	if err := storageManager.EnsureDefaultBuckets(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure default buckets")
	}

	if err := EnsureDefaultBucketRecords(ctx, db.Pool(), m.Service.DefaultBuckets()); err != nil {
		log.Warn().Err(err).Msg("Failed to ensure default bucket DB records")
	}

	m.Handlers = &StorageHandlers{
		Handler: NewStorageHandler(storageManager, db, cfg, &cfg.Storage.Transforms),
	}

	// Start background maintenance on a cancellable context: expired chunked
	// upload cleanup, expired S3 multipart cleanup, transform cache cleanup
	// and rate-limiter map sweeps. Stopped in Shutdown.
	bgCtx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.startBackgroundMaintenance(bgCtx)

	registry.Register(m.Service)
	return nil
}

// Shutdown stops the background maintenance goroutines.
func (m *StorageModule) Shutdown(ctx context.Context) error {
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}

// startBackgroundMaintenance wires the provider-level cleanup goroutines
// (which would otherwise leak expired chunked/multipart upload data) and the
// handler's limiter/cache maintenance into the module lifecycle.
func (m *StorageModule) startBackgroundMaintenance(ctx context.Context) {
	if m.Service != nil && m.Service.Provider != nil {
		switch p := m.Service.Provider.(type) {
		case *storage.LocalStorage:
			p.StartChunkedUploadCleanup(ctx)
		case *storage.S3Storage:
			p.StartMultipartUploadCleanup(ctx, storage.DefaultMultipartCleanupMaxAge)
		}
	}
	if m.Handlers != nil && m.Handlers.Handler != nil {
		m.Handlers.Handler.StartMaintenance(ctx)
	}
}
