package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// LokiLogStorage Name Tests
// =============================================================================

func TestLokiLogStorage_Name(t *testing.T) {
	t.Run("returns loki", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL: "http://localhost:3100",
		}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)
		require.NotNil(t, storage)

		assert.Equal(t, "loki", storage.Name())
	})
}

// =============================================================================
// LokiLogStorage Write Tests
// =============================================================================

func TestLokiLogStorage_Write_EmptyBatch(t *testing.T) {
	t.Run("returns nil for empty entries", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entries := []*LogEntry{}

		err = storage.Write(ctx, entries)
		assert.NoError(t, err)
	})
}

func TestLokiLogStorage_Write_SingleEntry(t *testing.T) {
	t.Run("successfully writes single entry", func(t *testing.T) {
		receivedRequest := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedRequest = true

			// Verify request method
			assert.Equal(t, http.MethodPost, r.Method)

			// Verify Content-Type
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			// Parse request body
			var pushReq LokiPushRequest
			err := json.NewDecoder(r.Body).Decode(&pushReq)
			require.NoError(t, err)

			// Verify streams
			assert.Len(t, pushReq.Streams, 1)
			assert.NotEmpty(t, pushReq.Streams[0].Stream)
			assert.Len(t, pushReq.Streams[0].Values, 1)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test log message",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
		assert.True(t, receivedRequest)
	})
}

func TestLokiLogStorage_Write_MultipleEntries(t *testing.T) {
	t.Run("groups entries by labels and writes batch", func(t *testing.T) {
		streamCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pushReq LokiPushRequest
			err := json.NewDecoder(r.Body).Decode(&pushReq)
			require.NoError(t, err)

			streamCount = len(pushReq.Streams)

			// Verify all entries are present
			totalValues := 0
			for _, stream := range pushReq.Streams {
				totalValues += len(stream.Values)
			}
			assert.Equal(t, 3, totalValues)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entries := []*LogEntry{
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "Request 1",
			},
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Message:   "Request 2",
			},
			{
				ID:        uuid.New(),
				Timestamp: time.Now(),
				Category:  LogCategorySecurity,
				Level:     LogLevelWarn,
				Message:   "Security event",
			},
		}

		err = storage.Write(ctx, entries)
		assert.NoError(t, err)
		// Should have 2 streams (different label combinations)
		assert.Equal(t, 2, streamCount)
	})
}

func TestLokiLogStorage_Write_AutoGenerateID(t *testing.T) {
	t.Run("generates UUID for nil entry ID", func(t *testing.T) {
		receivedEntry := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pushReq LokiPushRequest
			_ = json.NewDecoder(r.Body).Decode(&pushReq)

			// Parse the log line
			var entry LogEntry
			err := json.Unmarshal([]byte(pushReq.Streams[0].Values[0][1]), &entry)
			require.NoError(t, err)

			// Verify ID was generated
			assert.NotEqual(t, uuid.Nil, entry.ID)
			receivedEntry = true

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.Nil, // Nil ID
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Auto-generate ID",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
		assert.True(t, receivedEntry)
	})
}

func TestLokiLogStorage_Write_AutoGenerateTimestamp(t *testing.T) {
	t.Run("generates timestamp for zero entry timestamp", func(t *testing.T) {
		receivedEntry := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pushReq LokiPushRequest
			_ = json.NewDecoder(r.Body).Decode(&pushReq)

			// Verify nanosecond timestamp format
			timestamp := pushReq.Streams[0].Values[0][0]
			assert.NotEmpty(t, timestamp)
			assert.Greater(t, len(timestamp), 10) // Nanosecond precision

			// Parse the log line
			var entry LogEntry
			err := json.Unmarshal([]byte(pushReq.Streams[0].Values[0][1]), &entry)
			require.NoError(t, err)

			// Verify timestamp was generated
			assert.False(t, entry.Timestamp.IsZero())
			receivedEntry = true

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Time{}, // Zero timestamp
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Auto-generate timestamp",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
		assert.True(t, receivedEntry)
	})
}

func TestLokiLogStorage_Write_NanosecondTimestamps(t *testing.T) {
	t.Run("converts timestamps to nanoseconds", func(t *testing.T) {
		testTime := time.Date(2024, 1, 15, 12, 30, 45, 123456789, time.UTC)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pushReq LokiPushRequest
			_ = json.NewDecoder(r.Body).Decode(&pushReq)

			// Verify nanosecond timestamp
			timestamp := pushReq.Streams[0].Values[0][0]
			expectedNs := testTime.UnixNano()
			assert.Equal(t, fmt.Sprintf("%d", expectedNs), timestamp)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: testTime,
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Nanosecond precision",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
	})
}

func TestLokiLogStorage_Write_WithAuth(t *testing.T) {
	t.Run("sends basic auth headers when configured", func(t *testing.T) {
		authReceived := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			username, password, ok := r.BasicAuth()
			if ok {
				authReceived = true
				assert.Equal(t, "testuser", username)
				assert.Equal(t, "testpass", password)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{
			LokiURL:      server.URL,
			LokiUsername: "testuser",
			LokiPassword: "testpass",
		}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Auth test",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
		assert.True(t, authReceived)
	})
}

func TestLokiLogStorage_Write_WithTenantID(t *testing.T) {
	t.Run("sends X-Scope-OrgID header when configured", func(t *testing.T) {
		tenantReceived := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := r.Header.Get("X-Scope-OrgID")
			if tenantID != "" {
				tenantReceived = true
				assert.Equal(t, "tenant-123", tenantID)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{
			LokiURL:      server.URL,
			LokiTenantID: "tenant-123",
		}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Tenant test",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
		assert.True(t, tenantReceived)
	})
}

func TestLokiLogStorage_Write_StreamFormat(t *testing.T) {
	t.Run("produces correct Loki stream JSON structure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var pushReq LokiPushRequest
			err := json.NewDecoder(r.Body).Decode(&pushReq)
			require.NoError(t, err)

			// Verify stream structure
			stream := pushReq.Streams[0]

			// Labels should be a map
			assert.IsType(t, map[string]string{}, stream.Stream)

			// Values should be array of [timestamp, line]
			assert.IsType(t, [][2]string{}, stream.Values)
			assert.Len(t, stream.Values[0], 2)

			// Verify timestamp is string
			assert.IsType(t, "", stream.Values[0][0])

			// Verify log line is JSON string
			var entry LogEntry
			err = json.Unmarshal([]byte(stream.Values[0][1]), &entry)
			assert.NoError(t, err)

			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entry := &LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Stream format test",
		}

		err = storage.Write(ctx, []*LogEntry{entry})
		assert.NoError(t, err)
	})
}

// =============================================================================
// LokiLogStorage Query Tests
// =============================================================================

func TestLokiLogStorage_Query_NoResults(t *testing.T) {
	t.Run("returns empty result when no logs match", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result:     []LokiResult{},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			Category: LogCategoryHTTP,
			Levels:   []LogLevel{LogLevelInfo},
		}

		result, err := storage.Query(ctx, opts)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result.Entries)
		assert.Equal(t, int64(0), result.TotalCount)
		assert.False(t, result.HasMore)
	})
}

func TestLokiLogStorage_Query_WithFilters(t *testing.T) {
	t.Run("builds correct LogQL query with filters", func(t *testing.T) {
		receivedQuery := ""
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedQuery = r.URL.Query().Get("query")

			// Verify other query parameters
			limit := r.URL.Query().Get("limit")
			assert.Equal(t, "1000", limit) // Default limit when not specified

			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result:     []LokiResult{},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			Category:  LogCategoryHTTP,
			Levels:    []LogLevel{LogLevelInfo, LogLevelWarn},
			Component: "api",
		}

		_, err = storage.Query(ctx, opts)
		assert.NoError(t, err)

		// Verify LogQL syntax
		assert.Contains(t, receivedQuery, "{")
		assert.Contains(t, receivedQuery, "category=\"http\"")
		assert.Contains(t, receivedQuery, "level")
		assert.Contains(t, receivedQuery, "component=\"api\"")
	})
}

func TestLokiLogStorage_Query_Pagination(t *testing.T) {
	t.Run("applies offset to query parameters", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify limit parameter
			limit := r.URL.Query().Get("limit")
			assert.Equal(t, "50", limit)

			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result: []LokiResult{
						{
							Stream: map[string]string{
								"level":    "info",
								"category": "http",
							},
							Values: [][2]string{},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			Limit: 50,
		}

		_, err = storage.Query(ctx, opts)
		assert.NoError(t, err)
	})
}

func TestLokiLogStorage_Query_ParseResults(t *testing.T) {
	t.Run("parses Loki response into LogEntry", func(t *testing.T) {
		testEntry := LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test message",
			Component: "api",
		}
		entryJSON, _ := json.Marshal(testEntry)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result: []LokiResult{
						{
							Stream: map[string]string{
								"level":    "info",
								"category": "http",
							},
							Values: [][2]string{
								{fmt.Sprintf("%d", time.Now().UnixNano()), string(entryJSON)},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		result, err := storage.Query(ctx, LogQueryOptions{})

		assert.NoError(t, err)
		assert.Len(t, result.Entries, 1)
		assert.Equal(t, testEntry.ID, result.Entries[0].ID)
		assert.Equal(t, testEntry.Message, result.Entries[0].Message)
		assert.Equal(t, testEntry.Level, result.Entries[0].Level)
	})
}

func TestLokiLogStorage_Query_TimeRange(t *testing.T) {
	t.Run("includes start and end time in query", func(t *testing.T) {
		startTime := time.Now().Add(-1 * time.Hour)
		endTime := time.Now()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := r.URL.Query().Get("start")
			end := r.URL.Query().Get("end")

			assert.NotEmpty(t, start)
			assert.NotEmpty(t, end)

			w.Header().Set("Content-Type", "application/json")
			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result:     []LokiResult{},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			StartTime: startTime,
			EndTime:   endTime,
		}

		_, err = storage.Query(ctx, opts)
		assert.NoError(t, err)
	})
}

func TestLokiLogStorage_Query_SortDirection(t *testing.T) {
	t.Run("sets direction parameter based on SortAsc", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			direction := r.URL.Query().Get("direction")
			assert.Equal(t, "forward", direction)

			w.Header().Set("Content-Type", "application/json")
			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result:     []LokiResult{},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			SortAsc: true,
		}

		_, err = storage.Query(ctx, opts)
		assert.NoError(t, err)
	})
}

// =============================================================================
// LokiLogStorage Delete Tests
// =============================================================================

func TestLokiLogStorage_Delete(t *testing.T) {
	t.Run("returns error because delete is not supported", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		opts := LogQueryOptions{
			Category: LogCategoryHTTP,
		}

		count, err := storage.Delete(ctx, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not support delete")
		assert.Equal(t, int64(0), count)
	})
}

// =============================================================================
// LokiLogStorage Stats Tests
// =============================================================================

func TestLokiLogStorage_Stats(t *testing.T) {
	t.Run("aggregates statistics from queries", func(t *testing.T) {
		testEntry := LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test message",
		}
		entryJSON, _ := json.Marshal(testEntry)

		queryCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryCount++
			w.Header().Set("Content-Type", "application/json")

			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result: []LokiResult{
						{
							Stream: map[string]string{
								"level":    "info",
								"category": string(testEntry.Category),
							},
							Values: [][2]string{
								{fmt.Sprintf("%d", time.Now().UnixNano()), string(entryJSON)},
							},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		stats, err := storage.Stats(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, queryCount, 0)
		assert.NotNil(t, stats.EntriesByCategory)
		assert.NotNil(t, stats.EntriesByLevel)
	})
}

func TestLokiLogStorage_Stats_TimeRange(t *testing.T) {
	t.Run("tracks oldest and newest entry timestamps", func(t *testing.T) {
		oldTime := time.Now().Add(-2 * time.Hour)
		newTime := time.Now()

		oldEntryJSON, _ := json.Marshal(LogEntry{
			ID:        uuid.New(),
			Timestamp: oldTime,
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Old entry",
		})

		newEntryJSON, _ := json.Marshal(LogEntry{
			ID:        uuid.New(),
			Timestamp: newTime,
			Category:  LogCategorySecurity,
			Level:     LogLevelWarn,
			Message:   "New entry",
		})

		categoryIndex := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")

			var entryJSON []byte
			if categoryIndex == 0 {
				entryJSON = oldEntryJSON
			} else {
				entryJSON = newEntryJSON
			}
			categoryIndex++

			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result: []LokiResult{
						{
							Stream: map[string]string{"level": "info"},
							Values: [][2]string{{fmt.Sprintf("%d", time.Now().UnixNano()), string(entryJSON)}},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		stats, err := storage.Stats(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, stats.OldestEntry)
		assert.False(t, stats.OldestEntry.IsZero())
		assert.NotNil(t, stats.NewestEntry)
		assert.False(t, stats.NewestEntry.IsZero())
	})
}

// =============================================================================
// LokiLogStorage Health Tests
// =============================================================================

func TestLokiLogStorage_Health(t *testing.T) {
	t.Run("returns nil when Loki is ready", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("returns error when Loki is not ready", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not ready")
	})

	t.Run("returns error on connection failure", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:9999"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Health(ctx)
		assert.Error(t, err)
	})
}

func TestLokiLogStorage_Health_WithAuth(t *testing.T) {
	t.Run("sends auth headers to health endpoint", func(t *testing.T) {
		authReceived := false
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, _, ok := r.BasicAuth(); ok {
				authReceived = true
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := LogStorageConfig{
			LokiURL:      server.URL,
			LokiUsername: "user",
			LokiPassword: "pass",
		}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		err = storage.Health(ctx)
		assert.NoError(t, err)
		assert.True(t, authReceived)
	})
}

// =============================================================================
// LokiLogStorage All Log Categories Tests
// =============================================================================

func TestLokiLogStorage_AllLogCategories(t *testing.T) {
	t.Run("queries all built-in categories for stats", func(t *testing.T) {
		categoriesQueried := []string{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("query")
			categoriesQueried = append(categoriesQueried, query)

			w.Header().Set("Content-Type", "application/json")
			response := LokiQueryResponse{
				Status: "success",
				Data:   LokiData{ResultType: "streams", Result: []LokiResult{}},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = storage.Stats(ctx)

		assert.NoError(t, err)
		// Should query each built-in category
		assert.Greater(t, len(categoriesQueried), 0)
	})
}

// =============================================================================
// LokiLogStorage GroupByLabels Tests
// =============================================================================

func TestLokiLogStorage_GroupByLabels(t *testing.T) {
	t.Run("groups entries with identical labels", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		entries := []*LogEntry{
			{
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Component: "api",
			},
			{
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Component: "api",
			},
			{
				Category:  LogCategoryHTTP,
				Level:     LogLevelWarn,
				Component: "api",
			},
		}

		groups := storage.groupByLabels(entries)

		// Should have 2 groups (different levels)
		assert.Len(t, groups, 2)

		// Collect group sizes
		groupSizes := make([]int, len(groups))
		for i, group := range groups {
			groupSizes[i] = len(group)
		}

		// Should have one group with 2 entries and one with 1 entry
		assert.ElementsMatch(t, []int{1, 2}, groupSizes)
	})

	t.Run("handles empty entry list", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		groups := storage.groupByLabels([]*LogEntry{})
		assert.Len(t, groups, 0)
	})

	t.Run("groups by component when present", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		entries := []*LogEntry{
			{
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Component: "auth",
			},
			{
				Category:  LogCategoryHTTP,
				Level:     LogLevelInfo,
				Component: "storage",
			},
		}

		groups := storage.groupByLabels(entries)

		// Should have 2 groups (different components)
		assert.Len(t, groups, 2)
	})

	t.Run("groups by status code for HTTP logs", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		entries := []*LogEntry{
			{
				Category: LogCategoryHTTP,
				Level:    LogLevelInfo,
				Fields: map[string]interface{}{
					"status_code": 200.0,
				},
			},
			{
				Category: LogCategoryHTTP,
				Level:    LogLevelInfo,
				Fields: map[string]interface{}{
					"status_code": 404.0,
				},
			},
		}

		groups := storage.groupByLabels(entries)

		// Should have 2 groups (different status codes)
		assert.Len(t, groups, 2)
	})

	t.Run("groups by execution type for execution logs", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		entries := []*LogEntry{
			{
				Category:      LogCategoryExecution,
				Level:         LogLevelInfo,
				ExecutionType: "function",
			},
			{
				Category:      LogCategoryExecution,
				Level:         LogLevelInfo,
				ExecutionType: "job",
			},
		}

		groups := storage.groupByLabels(entries)

		// Should have 2 groups (different execution types)
		assert.Len(t, groups, 2)
	})
}

// =============================================================================
// LokiLogStorage BuildLogQL Tests
// =============================================================================

func TestLokiLogStorage_BuildLogQL(t *testing.T) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, err := newLokiLogStorage(cfg)
	require.NoError(t, err)

	t.Run("builds empty query for no options", func(t *testing.T) {
		opts := LogQueryOptions{}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, "{")
		assert.Contains(t, query, "}")
	})

	t.Run("filters by category", func(t *testing.T) {
		opts := LogQueryOptions{Category: LogCategoryHTTP}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `category="http"`)
	})

	t.Run("filters by single level", func(t *testing.T) {
		opts := LogQueryOptions{
			Levels: []LogLevel{LogLevelInfo},
		}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `level="info"`)
	})

	t.Run("filters by multiple levels with regex", func(t *testing.T) {
		opts := LogQueryOptions{
			Levels: []LogLevel{LogLevelInfo, LogLevelWarn, LogLevelError},
		}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `level|=~"info|warn|error"`)
	})

	t.Run("filters by component", func(t *testing.T) {
		opts := LogQueryOptions{Component: "auth"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `component="auth"`)
	})

	t.Run("filters by execution_id", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionID: "exec-123"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `execution_id="exec-123"`)
	})

	t.Run("filters by execution_type", func(t *testing.T) {
		opts := LogQueryOptions{ExecutionType: "function"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `execution_type="function"`)
	})

	t.Run("adds line filter for request_id", func(t *testing.T) {
		opts := LogQueryOptions{RequestID: "req-456"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `|= "req-456"`)
	})

	t.Run("adds line filter for trace_id", func(t *testing.T) {
		opts := LogQueryOptions{TraceID: "trace-789"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `|= "trace-789"`)
	})

	t.Run("adds line filter for user_id", func(t *testing.T) {
		opts := LogQueryOptions{UserID: uuid.New().String()}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `|= "`)
	})

	t.Run("adds case-insensitive search filter", func(t *testing.T) {
		opts := LogQueryOptions{Search: "error message"}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `|=~ "(?i)error message"`)
	})

	t.Run("excludes static asset extensions", func(t *testing.T) {
		opts := LogQueryOptions{HideStaticAssets: true}
		query := storage.buildLogQL(opts)

		// Should exclude .js
		assert.Contains(t, query, `!= ".js"`)
	})

	t.Run("combines multiple label selectors", func(t *testing.T) {
		opts := LogQueryOptions{
			Category:  LogCategoryHTTP,
			Levels:    []LogLevel{LogLevelInfo},
			Component: "api",
		}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `category="http"`)
		assert.Contains(t, query, `level="info"`)
		assert.Contains(t, query, `component="api"`)
	})

	t.Run("uses wildcard matcher when no label selectors", func(t *testing.T) {
		opts := LogQueryOptions{}
		query := storage.buildLogQL(opts)

		assert.Contains(t, query, `job=~".*"`)
	})
}

// =============================================================================
// LokiLogStorage Construction Tests
// =============================================================================

func TestNewLokiLogStorage(t *testing.T) {
	t.Run("creates storage with valid URL", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL: "http://localhost:3100",
		}
		storage, err := newLokiLogStorage(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, "loki", storage.Name())
	})

	t.Run("requires loki_url", func(t *testing.T) {
		cfg := LogStorageConfig{}
		storage, err := newLokiLogStorage(cfg)

		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Contains(t, err.Error(), "loki_url is required")
	})

	t.Run("parses and builds push URL", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL: "http://localhost:3100",
		}
		storage, err := newLokiLogStorage(cfg)

		assert.NoError(t, err)
		assert.Contains(t, storage.url, "/loki/api/v1/push")
	})

	t.Run("sets default labels when not provided", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL: "http://localhost:3100",
		}
		storage, err := newLokiLogStorage(cfg)

		assert.NoError(t, err)
		assert.NotNil(t, storage.labels)
		assert.Equal(t, []string{"app", "env"}, storage.labels)
	})

	t.Run("uses provided labels", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL:    "http://localhost:3100",
			LokiLabels: []string{"custom1", "custom2"},
		}
		storage, err := newLokiLogStorage(cfg)

		assert.NoError(t, err)
		assert.Equal(t, []string{"custom1", "custom2"}, storage.labels)
	})

	t.Run("rejects invalid URL", func(t *testing.T) {
		cfg := LogStorageConfig{
			LokiURL: ":invalid-url",
		}
		storage, err := newLokiLogStorage(cfg)

		assert.Error(t, err)
		assert.Nil(t, storage)
		assert.Contains(t, err.Error(), "invalid loki_url")
	})
}

// =============================================================================
// LokiLogStorage Helper Methods Tests
// =============================================================================

func TestLokiLogStorage_labelSetToString(t *testing.T) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, err := newLokiLogStorage(cfg)
	require.NoError(t, err)

	t.Run("converts label set to string key", func(t *testing.T) {
		labels := map[string]string{
			"level":    "info",
			"category": "http",
		}
		key := storage.labelSetToString(labels)

		assert.Contains(t, key, "level=info")
		assert.Contains(t, key, "category=http")
	})

	t.Run("produces consistent keys for same labels", func(t *testing.T) {
		labels := map[string]string{
			"level":    "info",
			"category": "http",
		}
		key1 := storage.labelSetToString(labels)
		key2 := storage.labelSetToString(labels)

		assert.Equal(t, key1, key2)
	})
}

func TestLokiLogStorage_getQueryLimit(t *testing.T) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, err := newLokiLogStorage(cfg)
	require.NoError(t, err)

	t.Run("uses default limit for zero", func(t *testing.T) {
		limit := storage.getQueryLimit(0)
		assert.Equal(t, 1000, limit)
	})

	t.Run("uses default limit for negative", func(t *testing.T) {
		limit := storage.getQueryLimit(-10)
		assert.Equal(t, 1000, limit)
	})

	t.Run("uses provided limit within range", func(t *testing.T) {
		limit := storage.getQueryLimit(500)
		assert.Equal(t, 500, limit)
	})

	t.Run("enforces maximum limit", func(t *testing.T) {
		limit := storage.getQueryLimit(20000)
		assert.Equal(t, 10000, limit)
	})
}

func TestLokiLogStorage_parseLogLine(t *testing.T) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, err := newLokiLogStorage(cfg)
	require.NoError(t, err)

	t.Run("parses valid JSON log line", func(t *testing.T) {
		entry := LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test message",
		}
		line, _ := json.Marshal(entry)

		parsed, err := storage.parseLogLine(string(line))
		assert.NoError(t, err)
		assert.Equal(t, entry.ID, parsed.ID)
		assert.Equal(t, entry.Message, parsed.Message)
		assert.Equal(t, entry.Category, parsed.Category)
	})

	t.Run("returns error for invalid JSON", func(t *testing.T) {
		_, err := storage.parseLogLine("not valid json")
		assert.Error(t, err)
	})
}

func TestLokiLogStorage_toLogLine(t *testing.T) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, err := newLokiLogStorage(cfg)
	require.NoError(t, err)

	t.Run("converts entry to JSON string", func(t *testing.T) {
		entry := LogEntry{
			ID:        uuid.New(),
			Timestamp: time.Now(),
			Category:  LogCategoryHTTP,
			Level:     LogLevelInfo,
			Message:   "Test message",
		}

		line := storage.toLogLine(&entry)

		var parsed LogEntry
		err := json.Unmarshal([]byte(line), &parsed)
		assert.NoError(t, err)
		assert.Equal(t, entry.ID, parsed.ID)
		assert.Equal(t, entry.Message, parsed.Message)
	})
}

func TestLokiLogStorage_Close(t *testing.T) {
	t.Run("closes without error", func(t *testing.T) {
		cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		err = storage.Close()
		assert.NoError(t, err)
	})
}

// =============================================================================
// LokiLogStorage GetExecutionLogs Tests
// =============================================================================

func TestLokiLogStorage_GetExecutionLogs(t *testing.T) {
	t.Run("queries with execution_id filter", func(t *testing.T) {
		queryReceived := ""
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryReceived = r.URL.Query().Get("query")

			w.Header().Set("Content-Type", "application/json")
			response := LokiQueryResponse{
				Status: "success",
				Data:   LokiData{ResultType: "streams", Result: []LokiResult{}},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		_, err = storage.GetExecutionLogs(ctx, "exec-123", 0)

		assert.NoError(t, err)
		assert.Contains(t, queryReceived, "execution_id=\"exec-123\"")
	})

	t.Run("filters by line number", func(t *testing.T) {
		testEntry := LogEntry{
			ID:         uuid.New(),
			Timestamp:  time.Now(),
			Category:   LogCategoryExecution,
			Level:      LogLevelInfo,
			Message:    "Test",
			LineNumber: 10,
		}
		entryJSON, _ := json.Marshal(testEntry)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			response := LokiQueryResponse{
				Status: "success",
				Data: LokiData{
					ResultType: "streams",
					Result: []LokiResult{
						{
							Stream: map[string]string{"level": "info"},
							Values: [][2]string{{fmt.Sprintf("%d", time.Now().UnixNano()), string(entryJSON)}},
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		cfg := LogStorageConfig{LokiURL: server.URL}
		storage, err := newLokiLogStorage(cfg)
		require.NoError(t, err)

		ctx := context.Background()
		entries, err := storage.GetExecutionLogs(ctx, "exec-123", 5)

		assert.NoError(t, err)
		// Entry with line 10 should be included (after line 5)
		assert.Len(t, entries, 1)
	})
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkLokiLogStorage_buildLogQL_Simple(b *testing.B) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, _ := newLokiLogStorage(cfg)
	opts := LogQueryOptions{
		Category: LogCategoryHTTP,
		Levels:   []LogLevel{LogLevelInfo},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildLogQL(opts)
	}
}

func BenchmarkLokiLogStorage_buildLogQL_Complex(b *testing.B) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, _ := newLokiLogStorage(cfg)
	userID := uuid.New()
	opts := LogQueryOptions{
		Category:         LogCategoryHTTP,
		Levels:           []LogLevel{LogLevelInfo, LogLevelWarn, LogLevelError},
		Component:        "auth",
		UserID:           userID.String(),
		StartTime:        time.Now().Add(-24 * time.Hour),
		EndTime:          time.Now(),
		Search:           "failed login",
		HideStaticAssets: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildLogQL(opts)
	}
}

func BenchmarkLokiLogStorage_groupByLabels(b *testing.B) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, _ := newLokiLogStorage(cfg)

	entries := make([]*LogEntry, 100)
	for i := 0; i < 100; i++ {
		entries[i] = &LogEntry{
			Category:  LogCategoryHTTP,
			Level:     []LogLevel{LogLevelInfo, LogLevelWarn, LogLevelError}[i%3],
			Component: []string{"api", "auth", "storage"}[i%3],
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.groupByLabels(entries)
	}
}

func BenchmarkLokiLogStorage_buildLabels(b *testing.B) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, _ := newLokiLogStorage(cfg)

	entry := &LogEntry{
		Category:  LogCategoryHTTP,
		Level:     LogLevelInfo,
		Component: "api",
		Fields: map[string]interface{}{
			"status_code": 200.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.buildLabels(entry)
	}
}

func BenchmarkLokiLogStorage_toLogLine(b *testing.B) {
	cfg := LogStorageConfig{LokiURL: "http://localhost:3100"}
	storage, _ := newLokiLogStorage(cfg)

	entry := &LogEntry{
		ID:        uuid.New(),
		Timestamp: time.Now(),
		Category:  LogCategoryHTTP,
		Level:     LogLevelInfo,
		Message:   "Test log message with some content",
		Fields: map[string]interface{}{
			"status_code": 200.0,
			"path":        "/api/test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.toLogLine(entry)
	}
}
