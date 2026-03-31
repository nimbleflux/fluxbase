package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogService(t *testing.T) {
	t.Run("errors for postgres backend without db", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend: "postgres",
		}

		svc, err := NewLogService(cfg, nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "database connection required")
	})

	t.Run("errors for s3 backend without storage provider", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend: "s3",
		}

		svc, err := NewLogService(cfg, nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "storage provider required")
	})

	t.Run("errors for s3 backend without bucket", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend:  "s3",
			S3Bucket: "",
		}

		// Create a mock storage provider
		mockProvider := &mockStorageProvider{}

		svc, err := NewLogService(cfg, nil, mockProvider)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "s3_bucket is required")
	})

	t.Run("creates local storage with default path", func(t *testing.T) {
		// Create temp directory for test
		tmpDir := t.TempDir()
		defaultPath := filepath.Join(tmpDir, "logs")

		cfg := LogStorageConfig{
			Backend:   "local",
			LocalPath: defaultPath,
		}

		svc, err := NewLogService(cfg, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, svc)
		defer func() { _ = svc.Close() }()

		assert.True(t, svc.IsLocal())
		assert.False(t, svc.IsPostgres())
		assert.False(t, svc.IsS3())
	})

	t.Run("errors for unsupported backend", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend: "mongodb",
		}

		svc, err := NewLogService(cfg, nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "unsupported log storage backend")
	})

	t.Run("creates elasticsearch storage with valid config", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend:              "elasticsearch",
			ElasticsearchURLs:    []string{"http://localhost:9200"},
			ElasticsearchVersion: 9,
		}

		svc, err := NewLogService(cfg, nil, nil)
		require.NoError(t, err)
		require.NotNil(t, svc)
		defer func() { _ = svc.Close() }()

		assert.Equal(t, "elasticsearch", svc.GetBackendName())
		assert.True(t, svc.IsElasticsearch())
	})

	t.Run("defaults to postgres when backend is empty", func(t *testing.T) {
		cfg := LogStorageConfig{
			Backend: "",
		}

		// Without a db connection, it should error with postgres message
		svc, err := NewLogService(cfg, nil, nil)
		require.Error(t, err)
		assert.Nil(t, svc)
		assert.Contains(t, err.Error(), "database connection required")
	})
}

func TestLogService_Methods(t *testing.T) {
	// Create a temporary directory for local storage
	tmpDir := t.TempDir()

	cfg := LogStorageConfig{
		Backend:       "local",
		LocalPath:     tmpDir,
		BatchSize:     50,
		FlushInterval: 500,
		BufferSize:    5000,
	}

	svc, err := NewLogService(cfg, nil, nil)
	require.NoError(t, err)
	defer func() { _ = svc.Close() }()

	t.Run("GetBackendName returns correct name", func(t *testing.T) {
		assert.Equal(t, "local", svc.GetBackendName())
	})

	t.Run("IsLocal returns true for local backend", func(t *testing.T) {
		assert.True(t, svc.IsLocal())
	})

	t.Run("IsPostgres returns false for local backend", func(t *testing.T) {
		assert.False(t, svc.IsPostgres())
	})

	t.Run("IsS3 returns false for local backend", func(t *testing.T) {
		assert.False(t, svc.IsS3())
	})

	t.Run("BatchSize returns configured value", func(t *testing.T) {
		assert.Equal(t, 50, svc.BatchSize())
	})

	t.Run("FlushIntervalMs returns configured value", func(t *testing.T) {
		assert.Equal(t, 500, svc.FlushIntervalMs())
	})

	t.Run("BufferSize returns configured value", func(t *testing.T) {
		assert.Equal(t, 5000, svc.BufferSize())
	})
}

func TestLogService_Defaults(t *testing.T) {
	// Create a temporary directory for local storage
	tmpDir := t.TempDir()

	cfg := LogStorageConfig{
		Backend:       "local",
		LocalPath:     tmpDir,
		BatchSize:     0,  // Should default
		FlushInterval: 0,  // Should default
		BufferSize:    -1, // Should default
	}

	svc, err := NewLogService(cfg, nil, nil)
	require.NoError(t, err)
	defer func() { _ = svc.Close() }()

	t.Run("BatchSize returns default when zero", func(t *testing.T) {
		assert.Equal(t, 100, svc.BatchSize())
	})

	t.Run("FlushIntervalMs returns default when zero", func(t *testing.T) {
		assert.Equal(t, 1000, svc.FlushIntervalMs())
	})

	t.Run("BufferSize returns default when negative", func(t *testing.T) {
		assert.Equal(t, 10000, svc.BufferSize())
	})
}

func TestLogService_Close(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := LogStorageConfig{
		Backend:   "local",
		LocalPath: tmpDir,
	}

	svc, err := NewLogService(cfg, nil, nil)
	require.NoError(t, err)

	err = svc.Close()
	require.NoError(t, err)
}

func TestMultiLogService(t *testing.T) {
	// Create temporary directories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	cfg1 := LogStorageConfig{
		Backend:   "local",
		LocalPath: tmpDir1,
	}
	cfg2 := LogStorageConfig{
		Backend:   "local",
		LocalPath: tmpDir2,
	}

	primary, err := NewLogService(cfg1, nil, nil)
	require.NoError(t, err)

	secondary, err := NewLogService(cfg2, nil, nil)
	require.NoError(t, err)

	multi := NewMultiLogService(primary, secondary)

	t.Run("Primary returns primary service", func(t *testing.T) {
		assert.Same(t, primary, multi.Primary())
	})

	t.Run("AllServices returns all services", func(t *testing.T) {
		services := multi.AllServices()
		assert.Len(t, services, 2)
		assert.Same(t, primary, services[0])
		assert.Same(t, secondary, services[1])
	})

	t.Run("Close closes all services", func(t *testing.T) {
		err := multi.Close()
		require.NoError(t, err)
	})
}

func TestMultiLogService_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := LogStorageConfig{
		Backend:   "local",
		LocalPath: tmpDir,
	}

	primary, err := NewLogService(cfg, nil, nil)
	require.NoError(t, err)

	multi := NewMultiLogService(primary)

	t.Run("AllServices returns only primary when no secondary", func(t *testing.T) {
		services := multi.AllServices()
		assert.Len(t, services, 1)
		assert.Same(t, primary, services[0])
	})

	err = multi.Close()
	require.NoError(t, err)
}

// mockStorageProvider implements Provider for testing
type mockStorageProvider struct{}

func (m *mockStorageProvider) Name() string {
	return "mock"
}

func (m *mockStorageProvider) Health(ctx context.Context) error {
	return nil
}

func (m *mockStorageProvider) Upload(ctx context.Context, bucket, key string, data io.Reader, size int64, opts *UploadOptions) (*Object, error) {
	return &Object{Key: key}, nil
}

func (m *mockStorageProvider) Download(ctx context.Context, bucket, key string, opts *DownloadOptions) (io.ReadCloser, *Object, error) {
	return nil, nil, nil
}

func (m *mockStorageProvider) Delete(ctx context.Context, bucket, key string) error {
	return nil
}

func (m *mockStorageProvider) Exists(ctx context.Context, bucket, key string) (bool, error) {
	return false, nil
}

func (m *mockStorageProvider) GetObject(ctx context.Context, bucket, key string) (*Object, error) {
	return nil, nil
}

func (m *mockStorageProvider) List(ctx context.Context, bucket string, opts *ListOptions) (*ListResult, error) {
	return &ListResult{}, nil
}

func (m *mockStorageProvider) CreateBucket(ctx context.Context, bucket string) error {
	return nil
}

func (m *mockStorageProvider) DeleteBucket(ctx context.Context, bucket string) error {
	return nil
}

func (m *mockStorageProvider) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return false, nil
}

func (m *mockStorageProvider) ListBuckets(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (m *mockStorageProvider) GenerateSignedURL(ctx context.Context, bucket, key string, opts *SignedURLOptions) (string, error) {
	return "", nil
}

func (m *mockStorageProvider) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

func (m *mockStorageProvider) MoveObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

// TestLocalLogStorageIntegration tests the local log storage with real filesystem
func TestLocalLogStorageIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	storage, err := NewLocalLogStorage(tmpDir)
	require.NoError(t, err)
	defer func() { _ = storage.Close() }()

	ctx := context.Background()

	t.Run("Name returns local", func(t *testing.T) {
		assert.Equal(t, "local", storage.Name())
	})

	t.Run("Health returns nil", func(t *testing.T) {
		err := storage.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("Write creates log files", func(t *testing.T) {
		entries := []*LogEntry{
			{
				Category: LogCategorySystem,
				Level:    LogLevelInfo,
				Message:  "Test log message",
			},
		}

		err := storage.Write(ctx, entries)
		require.NoError(t, err)

		// Verify file was created
		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Greater(t, len(files), 0)
	})

	t.Run("Stats returns empty stats for local storage", func(t *testing.T) {
		stats, err := storage.Stats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)
	})
}
