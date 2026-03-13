package storage

import (
	"context"
	"time"
)

// LogStorage defines the interface for log storage backends.
// Implementations can store logs in PostgreSQL, S3, local filesystem, etc.
type LogStorage interface {
	// Name returns the backend identifier (e.g., "postgres", "s3", "local").
	Name() string

	// Write writes a batch of log entries to the backend.
	// Implementations should handle batching efficiently.
	Write(ctx context.Context, entries []*LogEntry) error

	// Query retrieves logs matching the given options.
	// Returns a QueryResult with entries, total count, and pagination info.
	Query(ctx context.Context, opts LogQueryOptions) (*LogQueryResult, error)

	// GetExecutionLogs retrieves logs for a specific execution.
	// This is optimized for streaming execution logs with line number ordering.
	// Use afterLine to get logs after a specific line number for pagination.
	GetExecutionLogs(ctx context.Context, executionID string, afterLine int) ([]*LogEntry, error)

	// Delete removes logs matching the given options.
	// Used for retention cleanup. Returns the number of deleted entries.
	Delete(ctx context.Context, opts LogQueryOptions) (int64, error)

	// Stats returns statistics about stored logs.
	Stats(ctx context.Context) (*LogStats, error)

	// Health checks if the backend is operational.
	Health(ctx context.Context) error

	// Close releases resources held by the backend.
	Close() error
}

// LogStorageConfig contains configuration for creating a LogStorage instance.
type LogStorageConfig struct {
	// Backend type: "postgres", "s3", "local", "elasticsearch", "opensearch",
	// "clickhouse", "timescaledb", "postgres-timescaledb", "loki"
	Backend string `mapstructure:"backend"`

	// PostgreSQL settings (used when backend is "postgres")
	// Uses the main database connection

	// S3 settings (used when backend is "s3")
	S3Bucket string `mapstructure:"s3_bucket"`
	S3Prefix string `mapstructure:"s3_prefix"`

	// Local filesystem settings (used when backend is "local")
	LocalPath string `mapstructure:"local_path"`

	// Elasticsearch settings (used when backend is "elasticsearch")
	ElasticsearchURLs     []string `mapstructure:"elasticsearch_urls"`
	ElasticsearchUsername string   `mapstructure:"elasticsearch_username"`
	ElasticsearchPassword string   `mapstructure:"elasticsearch_password"`
	ElasticsearchIndex    string   `mapstructure:"elasticsearch_index"`   // default: "fluxbase-logs"
	ElasticsearchVersion  int      `mapstructure:"elasticsearch_version"` // default: 8 (8 or 9)

	// OpenSearch settings (used when backend is "opensearch")
	OpenSearchURLs     []string `mapstructure:"opensearch_urls"`
	OpenSearchUsername string   `mapstructure:"opensearch_username"`
	OpenSearchPassword string   `mapstructure:"opensearch_password"`
	OpenSearchIndex    string   `mapstructure:"opensearch_index"`   // default: "fluxbase-logs"
	OpenSearchVersion  int      `mapstructure:"opensearch_version"` // default: 2

	// ClickHouse settings (used when backend is "clickhouse")
	ClickHouseAddresses []string `mapstructure:"clickhouse_addresses"`
	ClickHouseUsername  string   `mapstructure:"clickhouse_username"`
	ClickHousePassword  string   `mapstructure:"clickhouse_password"`
	ClickHouseDatabase  string   `mapstructure:"clickhouse_database"` // default: "fluxbase"
	ClickHouseTable     string   `mapstructure:"clickhouse_table"`    // default: "logs"
	ClickHouseTTL       int      `mapstructure:"clickhouse_ttl_days"` // default: 30

	// TimescaleDB settings (used with postgres-timescaledb or timescaledb backend)
	TimescaleDBEnabled       bool          `mapstructure:"timescaledb_enabled"`
	TimescaleDBCompression   bool          `mapstructure:"timescaledb_compress"`
	TimescaleDBCompressAfter time.Duration `mapstructure:"timescaledb_compress_after"`
	TimescaleDBRetainAfter   time.Duration `mapstructure:"timescaledb_retain_after"`

	// Loki settings (used when backend is "loki")
	LokiURL      string   `mapstructure:"loki_url"` // required
	LokiUsername string   `mapstructure:"loki_username"`
	LokiPassword string   `mapstructure:"loki_password"`
	LokiTenantID string   `mapstructure:"loki_tenant_id"`
	LokiLabels   []string `mapstructure:"loki_labels"` // default: ["app", "env"]

	// Batching configuration
	BatchSize     int `mapstructure:"batch_size"`
	FlushInterval int `mapstructure:"flush_interval_ms"` // milliseconds

	// Buffer size for async writes
	BufferSize int `mapstructure:"buffer_size"`
}

// DefaultLogStorageConfig returns a LogStorageConfig with sensible defaults.
func DefaultLogStorageConfig() LogStorageConfig {
	return LogStorageConfig{
		Backend:       "postgres",
		S3Prefix:      "logs",
		LocalPath:     "./logs",
		BatchSize:     100,
		FlushInterval: 1000, // 1 second
		BufferSize:    10000,
	}
}
