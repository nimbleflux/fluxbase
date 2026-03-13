package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/nimbleflux/fluxbase/internal/database"
)

// LogService wraps a log storage provider and provides additional functionality.
type LogService struct {
	Storage LogStorage
	config  LogStorageConfig
}

// NewLogService creates a new log storage service based on configuration.
// Parameters:
//   - cfg: Log storage configuration
//   - db: Database connection (used for PostgreSQL backend)
//   - fileStorage: File storage provider (used for S3 backend)
func NewLogService(cfg LogStorageConfig, db *database.Connection, fileStorage Provider) (*LogService, error) {
	var storage LogStorage
	var err error

	backend := strings.ToLower(cfg.Backend)
	if backend == "" {
		backend = "postgres" // Default backend
	}

	switch backend {
	case "postgres", "postgresql":
		if db == nil {
			return nil, fmt.Errorf("database connection required for postgres log backend")
		}
		// Try TimescaleDB first, fall back to regular PostgreSQL if not available
		tsdbCfg := TimescaleDBConfig{
			Enabled:       true,
			Compressed:    true,
			CompressAfter: 168 * time.Hour,  // 7 days
			RetainAfter:   2160 * time.Hour, // 90 days
		}
		storage, err = newTimescaleDBLogStorage(tsdbCfg, db)
		if err != nil {
			// TimescaleDB not available, fall back to regular PostgreSQL
			storage = NewPostgresLogStorage(db)
		}

	case "s3":
		if fileStorage == nil {
			return nil, fmt.Errorf("storage provider required for s3 log backend")
		}
		if cfg.S3Bucket == "" {
			return nil, fmt.Errorf("s3_bucket is required for s3 log backend")
		}
		storage = NewS3LogStorage(fileStorage, cfg.S3Bucket, cfg.S3Prefix)

	case "local":
		localPath := cfg.LocalPath
		if localPath == "" {
			localPath = "./logs"
		}
		storage, err = NewLocalLogStorage(localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize local log storage: %w", err)
		}

	case "elasticsearch":
		storage, err = newElasticsearchLogStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize elasticsearch log storage: %w", err)
		}

	case "opensearch":
		storage, err = newOpenSearchLogStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize opensearch log storage: %w", err)
		}

	case "clickhouse":
		storage, err = newClickHouseLogStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize clickhouse log storage: %w", err)
		}

	case "timescaledb":
		if db == nil {
			return nil, fmt.Errorf("database connection required for timescaledb log backend")
		}
		tsdbCfg := TimescaleDBConfig{
			Enabled:       cfg.TimescaleDBEnabled,
			Compressed:    cfg.TimescaleDBCompression,
			CompressAfter: cfg.TimescaleDBCompressAfter,
			RetainAfter:   cfg.TimescaleDBRetainAfter,
		}
		storage, err = newTimescaleDBLogStorage(tsdbCfg, db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize timescaledb log storage: %w", err)
		}

	case "postgres-timescaledb":
		if db == nil {
			return nil, fmt.Errorf("database connection required for postgres-timescaledb log backend")
		}
		tsdbCfg := TimescaleDBConfig{
			Enabled:       cfg.TimescaleDBEnabled,
			Compressed:    cfg.TimescaleDBCompression,
			CompressAfter: cfg.TimescaleDBCompressAfter,
			RetainAfter:   cfg.TimescaleDBRetainAfter,
		}
		storage, err = newPostgresTimescaleDBStorage(tsdbCfg, db)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize postgres-timescaledb log storage: %w", err)
		}

	case "loki":
		storage, err = newLokiLogStorage(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize loki log storage: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported log storage backend: %s (supported: postgres, postgres-timescaledb, timescaledb, elasticsearch, opensearch, clickhouse, loki, s3, local)", cfg.Backend)
	}

	return &LogService{
		Storage: storage,
		config:  cfg,
	}, nil
}

// GetBackendName returns the name of the log storage backend.
func (s *LogService) GetBackendName() string {
	return s.Storage.Name()
}

// IsPostgres returns true if the log storage backend is PostgreSQL.
func (s *LogService) IsPostgres() bool {
	return s.Storage.Name() == "postgres"
}

// IsPostgresTimescaleDB returns true if the log storage backend is PostgreSQL with TimescaleDB.
func (s *LogService) IsPostgresTimescaleDB() bool {
	return s.Storage.Name() == "postgres-timescaledb"
}

// IsTimescaleDB returns true if the log storage backend is TimescaleDB.
func (s *LogService) IsTimescaleDB() bool {
	return s.Storage.Name() == "timescaledb"
}

// IsElasticsearch returns true if the log storage backend is Elasticsearch.
func (s *LogService) IsElasticsearch() bool {
	return s.Storage.Name() == "elasticsearch"
}

// IsOpenSearch returns true if the log storage backend is OpenSearch.
func (s *LogService) IsOpenSearch() bool {
	return s.Storage.Name() == "opensearch"
}

// IsClickHouse returns true if the log storage backend is ClickHouse.
func (s *LogService) IsClickHouse() bool {
	return s.Storage.Name() == "clickhouse"
}

// IsLoki returns true if the log storage backend is Loki.
func (s *LogService) IsLoki() bool {
	return s.Storage.Name() == "loki"
}

// IsS3 returns true if the log storage backend is S3.
func (s *LogService) IsS3() bool {
	return s.Storage.Name() == "s3"
}

// IsLocal returns true if the log storage backend is local filesystem.
func (s *LogService) IsLocal() bool {
	return s.Storage.Name() == "local"
}

// BatchSize returns the configured batch size.
func (s *LogService) BatchSize() int {
	if s.config.BatchSize <= 0 {
		return 100 // Default
	}
	return s.config.BatchSize
}

// FlushIntervalMs returns the configured flush interval in milliseconds.
func (s *LogService) FlushIntervalMs() int {
	if s.config.FlushInterval <= 0 {
		return 1000 // Default: 1 second
	}
	return s.config.FlushInterval
}

// BufferSize returns the configured buffer size for async writes.
func (s *LogService) BufferSize() int {
	if s.config.BufferSize <= 0 {
		return 10000 // Default
	}
	return s.config.BufferSize
}

// Close releases resources held by the log storage backend.
func (s *LogService) Close() error {
	return s.Storage.Close()
}

// MultiLogService wraps multiple log storage backends and writes to all of them.
// Useful for dual-writing to PostgreSQL (for querying) and S3 (for archival).
type MultiLogService struct {
	primary   *LogService
	secondary []*LogService
}

// NewMultiLogService creates a log service that writes to multiple backends.
// The primary backend is used for queries, while all backends receive writes.
func NewMultiLogService(primary *LogService, secondary ...*LogService) *MultiLogService {
	return &MultiLogService{
		primary:   primary,
		secondary: secondary,
	}
}

// Primary returns the primary log storage service (used for queries).
func (m *MultiLogService) Primary() *LogService {
	return m.primary
}

// AllServices returns all log storage services.
func (m *MultiLogService) AllServices() []*LogService {
	services := make([]*LogService, 0, 1+len(m.secondary))
	services = append(services, m.primary)
	services = append(services, m.secondary...)
	return services
}

// Close releases resources held by all log storage backends.
func (m *MultiLogService) Close() error {
	var lastErr error
	if err := m.primary.Close(); err != nil {
		lastErr = err
	}
	for _, s := range m.secondary {
		if err := s.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
