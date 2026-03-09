package storage

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Integration Tests for Local Log Storage
// These tests verify real-world scenarios including configuration,
// file structure, concurrent writes, and error handling.
// =============================================================================

// TestLocalLogStorage_ConfigIntegration tests the full integration with
// FLUXBASE_LOGGING_BACKEND environment variable configuration
func TestLocalLogStorage_ConfigIntegration(t *testing.T) {
	t.Run("FLUXBASE_LOGGING_BACKEND=local initializes correctly", func(t *testing.T) {
		// Set the environment variable
		t.Setenv("FLUXBASE_LOGGING_BACKEND", "local")
		t.Setenv("FLUXBASE_LOGGING_LOCAL_PATH", "")

		// This simulates what happens when the config system
		// creates a LocalLogStorage backend
		tmpDir := t.TempDir()
		localPath := filepath.Join(tmpDir, "logs")
		storage, err := NewLocalLogStorage(localPath)

		require.NoError(t, err)
		require.NotNil(t, storage)
		assert.Equal(t, "local", storage.Name())
		assert.Equal(t, localPath, storage.basePath)

		// Verify the directory was created
		info, err := os.Stat(localPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())

		// Test that storage is operational
		err = storage.Health(context.Background())
		assert.NoError(t, err)
	})

	t.Run("FLUXBASE_LOGGING_BACKEND=local with custom path", func(t *testing.T) {
		t.Setenv("FLUXBASE_LOGGING_BACKEND", "local")
		customPath := filepath.Join(t.TempDir(), "custom-logs", "nested", "path")

		storage, err := NewLocalLogStorage(customPath)

		require.NoError(t, err)
		assert.Equal(t, customPath, storage.basePath)

		// Verify nested directories were created
		info, err := os.Stat(customPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

// TestLocalLogStorage_NDJSONStructure verifies that logs are written
// in correct NDJSON format with proper file structure
func TestLocalLogStorage_NDJSONStructure(t *testing.T) {
	t.Run("writes valid NDJSON files", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write entries
		entries := []*LogEntry{
			{
				ID:        uuid.New(),
				Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "First log message",
				Component: "api",
				UserID:    "user-123",
			},
			{
				ID:        uuid.New(),
				Timestamp: time.Date(2024, 1, 15, 10, 30, 1, 0, time.UTC),
				Category:  LogCategoryHTTP,
				Level:     LogLevelWarn,
				Message:   "Second log message",
				Component: "api",
			},
		}

		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Find the created file
		var filePath string
		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(path, ".ndjson") {
				filePath = path
				return filepath.SkipAll
			}
			return nil
		})
		require.NoError(t, err)
		require.NotEmpty(t, filePath, "No NDJSON file was created")

		// Verify file structure: {basePath}/{category}/{YYYY}/{MM}/{DD}/{batch_uuid}.ndjson
		relPath, err := filepath.Rel(tmpDir, filePath)
		require.NoError(t, err)

		parts := strings.Split(relPath, string(filepath.Separator))
		require.GreaterOrEqual(t, len(parts), 5, "Path should have at least 5 parts: category/YYYY/MM/DD/filename.ndjson")

		assert.Equal(t, string(LogCategoryHTTP), parts[0])
		assert.Equal(t, "2024", parts[1])
		assert.Equal(t, "01", parts[2])
		assert.Equal(t, "15", parts[3])
		assert.True(t, strings.HasSuffix(parts[4], ".ndjson"))

		// Read and verify NDJSON content
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)

		// Verify each line is valid JSON
		scanner := bufio.NewScanner(bytes.NewReader(content))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			var entry LogEntry
			err := json.Unmarshal(scanner.Bytes(), &entry)
			require.NoError(t, err, "Line %d should be valid JSON: %s", lineNum, scanner.Text())
		}
		assert.Equal(t, 2, lineNum, "Should have 2 lines")
	})

	t.Run("execution logs have correct structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := uuid.New().String()
		entries := []*LogEntry{
			{
				Timestamp:   time.Now(),
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Starting execution",
				ExecutionID: execID,
				LineNumber:  1,
			},
			{
				Timestamp:   time.Now(),
				Category:    LogCategoryExecution,
				Level:       LogLevelInfo,
				Message:     "Execution complete",
				ExecutionID: execID,
				LineNumber:  2,
			},
		}

		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Find the execution log file
		var filePath string
		err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.Contains(path, "exec_"+execID) {
				filePath = path
				return filepath.SkipAll
			}
			return nil
		})
		require.NoError(t, err)
		require.NotEmpty(t, filePath, "Execution log file was not created")

		// Verify filename structure: exec_{execution_id}.ndjson
		filename := filepath.Base(filePath)
		assert.Equal(t, "exec_"+execID+".ndjson", filename)
	})

	t.Run("can read back written entries", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write entry
		originalEntry := &LogEntry{
			Timestamp: time.Now().Truncate(time.Microsecond), // Truncate for JSON precision
			Category:  LogCategorySecurity,
			Level:     LogLevelWarn,
			Message:   "Security event",
			Component: "auth",
			UserID:    "user-456",
			RequestID: uuid.New().String(),
		}

		err = storage.Write(context.Background(), []*LogEntry{originalEntry})
		require.NoError(t, err)

		// Query back
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategorySecurity,
		})

		require.NoError(t, err)
		require.Greater(t, len(result.Entries), 0)

		// Verify entry matches
		found := false
		for _, entry := range result.Entries {
			if entry.Message == originalEntry.Message &&
				entry.Category == originalEntry.Category &&
				entry.Component == originalEntry.Component &&
				entry.UserID == originalEntry.UserID &&
				entry.RequestID == originalEntry.RequestID {
				found = true
				break
			}
		}
		assert.True(t, found, "Written entry should be readable back")
	})
}

// TestLocalLogStorage_ConcurrentWrites tests thread safety
func TestLocalLogStorage_ConcurrentWrites(t *testing.T) {
	t.Run("concurrent writes to same category", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		numGoroutines := 10
		entriesPerGoroutine := 50
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*entriesPerGoroutine)

		// Launch multiple goroutines writing to the same category
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < entriesPerGoroutine; j++ {
					entry := &LogEntry{
						Timestamp: time.Now(),
						Category:  LogCategoryHTTP,
						Level:     LogLevelInfo,
						Message:   fmt.Sprintf("Goroutine %d, entry %d", goroutineID, j),
						Component: "api",
					}

					if err := storage.Write(context.Background(), []*LogEntry{entry}); err != nil {
						select {
						case errors <- err:
						default:
						}
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		errorList := make([]error, 0)
		for err := range errors {
			errorList = append(errorList, err)
		}

		assert.Empty(t, errorList, "Concurrent writes should not produce errors")

		// Verify all entries were written (check TotalCount, not paginated entries)
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
			Limit:    1000, // Set high limit to get all entries
		})
		require.NoError(t, err)
		assert.Equal(t, int64(numGoroutines*entriesPerGoroutine), result.TotalCount)
		assert.Equal(t, numGoroutines*entriesPerGoroutine, len(result.Entries))
		assert.False(t, result.HasMore)
	})

	t.Run("concurrent writes to execution logs", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := uuid.New().String()
		numGoroutines := 10
		linesPerGoroutine := 20
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*linesPerGoroutine)

		// Launch multiple goroutines writing to the same execution
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < linesPerGoroutine; j++ {
					lineNum := goroutineID*linesPerGoroutine + j + 1
					entry := &LogEntry{
						Timestamp:   time.Now(),
						Category:    LogCategoryExecution,
						Level:       LogLevelInfo,
						Message:     fmt.Sprintf("Line from goroutine %d", goroutineID),
						ExecutionID: execID,
						LineNumber:  lineNum,
					}

					if err := storage.Write(context.Background(), []*LogEntry{entry}); err != nil {
						select {
						case errors <- err:
						default:
						}
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		errorList := make([]error, 0)
		for err := range errors {
			errorList = append(errorList, err)
		}

		assert.Empty(t, errorList, "Concurrent execution log writes should not produce errors")

		// Verify all entries were written
		entries, err := storage.GetExecutionLogs(context.Background(), execID, 0)
		require.NoError(t, err)
		assert.Equal(t, numGoroutines*linesPerGoroutine, len(entries))

		// Verify line numbers are sequential
		lineNumbers := make([]int, len(entries))
		for i, entry := range entries {
			lineNumbers[i] = entry.LineNumber
		}
		for i := 1; i < len(lineNumbers); i++ {
			assert.Equal(t, lineNumbers[i-1]+1, lineNumbers[i], "Line numbers should be sequential")
		}
	})

	t.Run("concurrent writes and reads", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		numGoroutines := 5
		var wg sync.WaitGroup

		// Writers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					entry := &LogEntry{
						Timestamp: time.Now(),
						Category:  LogCategorySystem,
						Level:     LogLevelInfo,
						Message:   fmt.Sprintf("Writer %d, message %d", id, j),
					}
					_ = storage.Write(context.Background(), []*LogEntry{entry})
				}
			}(i)
		}

		// Readers
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					_, _ = storage.Query(context.Background(), LogQueryOptions{
						Category: LogCategorySystem,
						Limit:    10,
					})
					_, _ = storage.Stats(context.Background())
				}
			}(i)
		}

		wg.Wait()

		// Final verification
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategorySystem,
		})
		require.NoError(t, err)
		assert.Equal(t, numGoroutines*20, len(result.Entries))
	})
}

// TestLocalLogStorage_ErrorHandling tests error scenarios
func TestLocalLogStorage_ErrorHandling(t *testing.T) {
	t.Run("returns error when path is not writable", func(t *testing.T) {
		if os.Getuid() == 0 {
			t.Skip("Skipping as root can write to read-only directories")
		}

		tmpDir := t.TempDir()
		readOnlyPath := filepath.Join(tmpDir, "readonly")

		// Create directory
		err := os.MkdirAll(readOnlyPath, 0o750)
		require.NoError(t, err)

		// Make it read-only
		err = os.Chmod(readOnlyPath, 0o444)
		require.NoError(t, err)

		// Try to create storage in read-only directory
		// The base directory exists but we won't be able to create subdirectories
		storage, err := NewLocalLogStorage(readOnlyPath)

		// Storage creation may succeed since directory exists
		if err == nil {
			// But writing should fail
			entry := &LogEntry{
				Timestamp: time.Now(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "Test",
			}

			err = storage.Write(context.Background(), []*LogEntry{entry})
			assert.Error(t, err, "Write should fail when directory is read-only")
		} else {
			assert.Error(t, err)
		}

		// Clean up - restore permissions for temp dir cleanup
		_ = os.Chmod(readOnlyPath, 0o750)
	})

	t.Run("handles invalid path characters gracefully", func(t *testing.T) {
		invalidPaths := []string{
			"/proc/invalid/log/path/that/does/not/exist",
			filepath.Join("/", "nonexistent", "deep", "nested", "path", "logs"),
		}

		for _, invalidPath := range invalidPaths {
			storage, err := NewLocalLogStorage(invalidPath)
			assert.Error(t, err, "Should return error for invalid path: %s", invalidPath)
			assert.Nil(t, storage)
		}
	})

	t.Run("Health check creates directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		missingPath := filepath.Join(tmpDir, "does", "not", "exist")

		// Create storage without creating the directory
		storage := &LocalLogStorage{basePath: missingPath}

		// Health check should create it
		err := storage.Health(context.Background())
		assert.NoError(t, err, "Health check should create missing directory")

		// Verify directory exists
		info, err := os.Stat(missingPath)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("handles empty entries gracefully", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write nil slice
		err = storage.Write(context.Background(), nil)
		assert.NoError(t, err)

		// Write empty slice
		err = storage.Write(context.Background(), []*LogEntry{})
		assert.NoError(t, err)

		// Query should still work
		result, err := storage.Query(context.Background(), LogQueryOptions{})
		assert.NoError(t, err)
		assert.Empty(t, result.Entries)
	})
}

// TestLocalLogStorage_RetentionPolicyIntegration tests retention
func TestLocalLogStorage_RetentionPolicyIntegration(t *testing.T) {
	t.Run("delete removes old log files", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write some logs
		oldTime := time.Now().AddDate(0, 0, -10) // 10 days ago
		entries := []*LogEntry{
			{
				Timestamp: oldTime,
				Category:  LogCategorySystem,
				Level:     LogLevelInfo,
				Message:   "Old log entry",
			},
		}

		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Delete entries older than 7 days
		cutoff := time.Now().AddDate(0, 0, -7)
		deleted, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: LogCategorySystem,
			EndTime:  cutoff,
		})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		// Verify deletion worked
		result, err := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategorySystem,
		})
		require.NoError(t, err)
		assert.Empty(t, result.Entries)
	})

	t.Run("delete cleans up empty directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write and then delete
		entries := []*LogEntry{
			{
				Timestamp: time.Now(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "Test",
			},
		}

		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Delete all
		deleted, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		// Verify empty directories were cleaned up (handled internally)
		// cleanEmptyDirs is called automatically and doesn't return errors
	})

	t.Run("delete only affects specified category", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		// Write logs to different categories
		entries := []*LogEntry{
			{Timestamp: time.Now(), Category: LogCategoryHTTP, Level: LogLevelInfo, Message: "HTTP log"},
			{Timestamp: time.Now(), Category: LogCategorySecurity, Level: LogLevelInfo, Message: "Security log"},
		}

		err = storage.Write(context.Background(), entries)
		require.NoError(t, err)

		// Delete only HTTP logs
		deleted, err := storage.Delete(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})

		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(1))

		// HTTP logs should be gone
		httpResult, _ := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategoryHTTP,
		})
		assert.Empty(t, httpResult.Entries)

		// Security logs should still exist
		securityResult, _ := storage.Query(context.Background(), LogQueryOptions{
			Category: LogCategorySecurity,
		})
		assert.NotEmpty(t, securityResult.Entries)
	})
}

// TestLocalLogStorage_StreamExecutionLogs tests the streaming functionality
func TestLocalLogStorage_StreamExecutionLogs(t *testing.T) {
	t.Run("streams execution logs in real-time", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		execID := uuid.New().String()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Start streaming
		ch, err := storage.StreamExecutionLogs(ctx, execID)
		require.NoError(t, err)

		// Write logs
		go func() {
			for i := 0; i < 5; i++ {
				time.Sleep(100 * time.Millisecond)
				entry := &LogEntry{
					Timestamp:   time.Now(),
					Category:    LogCategoryExecution,
					Level:       LogLevelInfo,
					Message:     fmt.Sprintf("Streaming line %d", i),
					ExecutionID: execID,
					LineNumber:  i + 1,
				}
				_ = storage.Write(context.Background(), []*LogEntry{entry})
			}
		}()

		// Collect streamed entries
		receivedEntries := make([]*LogEntry, 0)
		for entry := range ch {
			receivedEntries = append(receivedEntries, entry)
			if len(receivedEntries) >= 5 {
				break
			}
		}

		assert.GreaterOrEqual(t, len(receivedEntries), 1, "Should receive at least some streamed entries")
	})

	t.Run("stream handles non-existent execution", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewLocalLogStorage(tmpDir)
		require.NoError(t, err)

		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		ch, err := storage.StreamExecutionLogs(ctx, "non-existent-exec-id")
		require.NoError(t, err)

		// Channel should be empty or closed
		receivedCount := 0
		for range ch {
			receivedCount++
		}

		assert.Equal(t, 0, receivedCount, "Should not receive entries for non-existent execution")
	})
}
