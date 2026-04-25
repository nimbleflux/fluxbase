package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LokiLogStorage implements LogStorage using Grafana Loki.
// Loki is a horizontally-scalable, highly-available log aggregation system.
type LokiLogStorage struct {
	client   *http.Client
	url      string
	username string
	password string
	tenantID string
	labels   []string
}

// LokiPushRequest represents the JSON payload for Loki's push API.
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a single stream with a unique label set.
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [nanosecondTimestamp, logLine]
}

// LokiQueryResponse represents the response from Loki's query API.
type LokiQueryResponse struct {
	Status string   `json:"status"`
	Data   LokiData `json:"data"`
}

// LokiData contains the query results.
type LokiData struct {
	ResultType string       `json:"resultType"`
	Result     []LokiResult `json:"result"`
}

// LokiResult represents a single result stream.
type LokiResult struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"` // [nanosecondTimestamp, logLine]
}

// newLokiLogStorage creates a new Loki-backed log storage.
func newLokiLogStorage(cfg LogStorageConfig) (*LokiLogStorage, error) {
	if cfg.LokiURL == "" {
		return nil, fmt.Errorf("loki_url is required for Loki backend")
	}

	// Parse and validate the Loki URL
	parsedURL, err := url.Parse(cfg.LokiURL)
	if err != nil {
		return nil, fmt.Errorf("invalid loki_url: %w", err)
	}

	// Build the full push URL: {base_url}/loki/api/v1/push
	pushURL := parsedURL.JoinPath("loki", "api", "v1", "push").String()

	// Default labels to ["app", "env"] if not provided
	labels := cfg.LokiLabels
	if len(labels) == 0 {
		labels = []string{"app", "env"}
	}

	return &LokiLogStorage{
		client:   &http.Client{Timeout: 30 * time.Second},
		url:      pushURL,
		username: cfg.LokiUsername,
		password: cfg.LokiPassword,
		tenantID: cfg.LokiTenantID,
		labels:   labels,
	}, nil
}

// Name returns the backend identifier.
func (s *LokiLogStorage) Name() string {
	return "loki"
}

// Write writes a batch of log entries to Loki.
func (s *LokiLogStorage) Write(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Group entries by unique label combinations
	groups := s.groupByLabels(entries)

	// Build Loki push request
	streams := make([]LokiStream, 0, len(groups))
	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		// Build labels from the first entry in the group
		labels := s.buildLabels(group[0])

		// Convert entries to Loki values (nanosecond timestamps + JSON log lines)
		values := make([][2]string, len(group))
		for i, entry := range group {
			// Ensure ID and timestamp are set
			if entry.ID == uuid.Nil {
				entry.ID = uuid.New()
			}
			if entry.Timestamp.IsZero() {
				entry.Timestamp = time.Now()
			}

			// Loki uses nanosecond timestamps
			nsTimestamp := fmt.Sprintf("%d", entry.Timestamp.UnixNano())
			logLine := s.toLogLine(entry)

			values[i] = [2]string{nsTimestamp, logLine}
		}

		streams = append(streams, LokiStream{
			Stream: labels,
			Values: values,
		})
	}

	if len(streams) == 0 {
		return nil
	}

	pushReq := LokiPushRequest{Streams: streams}

	// Marshal to JSON
	reqBody, err := json.Marshal(pushReq)
	if err != nil {
		return fmt.Errorf("failed to marshal loki request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Set basic auth if provided
	if s.username != "" && s.password != "" {
		req.SetBasicAuth(s.username, s.password)
	}

	// Set tenant ID header if provided (multi-tenancy)
	if s.tenantID != "" {
		req.Header.Set("X-Scope-OrgID", s.tenantID)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send logs to loki: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("loki returned status %d", resp.StatusCode)
	}

	return nil
}

// Query retrieves logs matching the given options using LogQL.
func (s *LokiLogStorage) Query(ctx context.Context, opts LogQueryOptions) (*LogQueryResult, error) {
	// Build LogQL query
	query := s.buildLogQL(opts)

	// Parse base URL to get query endpoint
	parsedURL, err := url.Parse(s.url)
	if err != nil {
		return nil, fmt.Errorf("invalid loki url: %w", err)
	}

	// Build query URL: {base_url}/loki/api/v1/query_range
	queryURL := parsedURL.JoinPath("..", "query_range")

	// Build query parameters
	params := url.Values{}
	params.Set("query", query)
	params.Set("limit", fmt.Sprintf("%d", s.getQueryLimit(opts.Limit)))

	// Time range
	if !opts.StartTime.IsZero() {
		params.Set("start", fmt.Sprintf("%d", opts.StartTime.UnixNano()))
	} else {
		// Default to 1 hour ago
		params.Set("start", fmt.Sprintf("%d", time.Now().Add(-1*time.Hour).UnixNano()))
	}

	if !opts.EndTime.IsZero() {
		params.Set("end", fmt.Sprintf("%d", opts.EndTime.UnixNano()))
	} else {
		params.Set("end", fmt.Sprintf("%d", time.Now().UnixNano()))
	}

	// Direction
	direction := "backward"
	if opts.SortAsc {
		direction = "forward"
	}
	params.Set("direction", direction)

	// Pagination via offset
	if opts.Offset > 0 {
		params.Set("start", fmt.Sprintf("%d", opts.StartTime.UnixNano()+int64(opts.Offset)))
	}

	queryURL.RawQuery = params.Encode()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth headers
	if s.username != "" && s.password != "" {
		req.SetBasicAuth(s.username, s.password)
	}
	if s.tenantID != "" {
		req.Header.Set("X-Scope-OrgID", s.tenantID)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query loki: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("loki query returned status %d", resp.StatusCode)
	}

	// Parse response
	var lokiResp LokiQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		return nil, fmt.Errorf("failed to decode loki response: %w", err)
	}

	// Convert Loki results to LogEntry
	entries := make([]*LogEntry, 0)
	for _, result := range lokiResp.Data.Result {
		for _, value := range result.Values {
			entry, err := s.parseLogLine(value[1])
			if err != nil {
				continue
			}
			entries = append(entries, entry)
		}
	}

	// Estimate total count (Loki doesn't provide exact counts without additional queries)
	totalCount := int64(len(entries))

	return &LogQueryResult{
		Entries:    entries,
		TotalCount: totalCount,
		HasMore:    false, // Loki pagination is handled differently
	}, nil
}

// GetExecutionLogs retrieves logs for a specific execution.
func (s *LokiLogStorage) GetExecutionLogs(ctx context.Context, executionID string, afterLine int) ([]*LogEntry, error) {
	// Query with execution_id label filter
	opts := LogQueryOptions{
		ExecutionID: executionID,
		AfterLine:   afterLine,
		SortAsc:     true, // Always stream in chronological order
	}

	result, err := s.Query(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Filter by line number
	var entries []*LogEntry
	for _, entry := range result.Entries {
		if entry.LineNumber > afterLine {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Delete removes logs matching the given options.
// Note: Loki doesn't support delete via API. Retention is handled server-side.
func (s *LokiLogStorage) Delete(ctx context.Context, opts LogQueryOptions) (int64, error) {
	return 0, fmt.Errorf("loki does not support delete via API; use Loki's retention policies instead")
}

// Stats returns statistics about stored logs.
// Note: Loki doesn't provide efficient stats API, so we approximate by querying.
func (s *LokiLogStorage) Stats(ctx context.Context) (*LogStats, error) {
	stats := &LogStats{
		EntriesByCategory: make(map[LogCategory]int64),
		EntriesByLevel:    make(map[LogLevel]int64),
	}

	// Query for each category to get counts
	for _, category := range AllBuiltinCategories() {
		opts := LogQueryOptions{
			Category: category,
			Limit:    1000, // Sample limit
		}

		result, err := s.Query(ctx, opts)
		if err != nil {
			continue
		}

		count := int64(len(result.Entries))
		stats.EntriesByCategory[category] = count
		stats.TotalEntries += count

		// Count by level
		for _, entry := range result.Entries {
			stats.EntriesByLevel[entry.Level]++
		}

		// Track time range
		if len(result.Entries) > 0 {
			if stats.OldestEntry == nil || result.Entries[0].Timestamp.Before(*stats.OldestEntry) {
				stats.OldestEntry = &result.Entries[0].Timestamp
			}
			if stats.NewestEntry == nil || result.Entries[len(result.Entries)-1].Timestamp.After(*stats.NewestEntry) {
				stats.NewestEntry = &result.Entries[len(result.Entries)-1].Timestamp
			}
		}
	}

	return stats, nil
}

// Health checks if Loki is operational.
func (s *LokiLogStorage) Health(ctx context.Context) error {
	// Parse URL to get ready endpoint
	parsedURL, err := url.Parse(s.url)
	if err != nil {
		return fmt.Errorf("invalid loki url: %w", err)
	}

	// Build ready URL: {base_url}/ready
	readyURL := parsedURL.JoinPath("..", "..", "ready")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, readyURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set auth headers
	if s.username != "" && s.password != "" {
		req.SetBasicAuth(s.username, s.password)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("loki health check failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("loki not ready: status %d", resp.StatusCode)
	}

	return nil
}

// Close releases resources (no-op for HTTP client).
func (s *LokiLogStorage) Close() error {
	return nil
}

// groupByLabels groups entries by their label combinations.
func (s *LokiLogStorage) groupByLabels(entries []*LogEntry) [][]*LogEntry {
	groups := make(map[string][]*LogEntry)

	for _, entry := range entries {
		// Create a key from the labels
		labels := s.buildLabels(entry)
		key := s.labelSetToString(labels)

		groups[key] = append(groups[key], entry)
	}

	// Convert map to slice
	result := make([][]*LogEntry, 0, len(groups))
	for _, group := range groups {
		result = append(result, group)
	}

	return result
}

// buildLabels extracts labels from a log entry.
func (s *LokiLogStorage) buildLabels(entry *LogEntry) map[string]string {
	labels := make(map[string]string)

	// Standard Loki labels (must be low cardinality)
	labels["level"] = string(entry.Level)
	labels["category"] = string(entry.Category)

	if entry.Component != "" {
		labels["component"] = entry.Component
	}

	// For HTTP logs, include status code
	if entry.Category == LogCategoryHTTP && entry.Fields != nil {
		if statusCode, ok := entry.Fields["status_code"].(float64); ok {
			labels["status_code"] = fmt.Sprintf("%.0f", statusCode)
		}
	}

	// For execution logs, include execution type
	if entry.Category == LogCategoryExecution {
		if entry.ExecutionType != "" {
			labels["execution_type"] = entry.ExecutionType
		}
	}

	return labels
}

// buildLogQL converts query options to a LogQL query string.
func (s *LokiLogStorage) buildLogQL(opts LogQueryOptions) string {
	var selectors []string

	// Build label selectors
	if opts.Category != "" {
		selectors = append(selectors, fmt.Sprintf(`category="%s"`, opts.Category))
	}

	if len(opts.Levels) > 0 {
		if len(opts.Levels) == 1 {
			selectors = append(selectors, fmt.Sprintf(`level="%s"`, opts.Levels[0]))
		} else {
			// Multiple levels: use |= (regex match)
			levels := make([]string, len(opts.Levels))
			for i, level := range opts.Levels {
				levels[i] = string(level)
			}
			selectors = append(selectors, fmt.Sprintf(`level|=~"%s"`, strings.Join(levels, "|")))
		}
	}

	if opts.Component != "" {
		selectors = append(selectors, fmt.Sprintf(`component="%s"`, opts.Component))
	}

	if opts.ExecutionID != "" {
		selectors = append(selectors, fmt.Sprintf(`execution_id="%s"`, opts.ExecutionID))
	}

	if opts.ExecutionType != "" {
		selectors = append(selectors, fmt.Sprintf(`execution_type="%s"`, opts.ExecutionType))
	}

	// Build base query
	var query string
	if len(selectors) > 0 {
		query = "{" + strings.Join(selectors, ", ") + "}"
	} else {
		query = "{job=~\".*\"}" // Match all streams
	}

	// Add filters for non-label fields (line filters)
	if opts.RequestID != "" {
		query += fmt.Sprintf(` |= "%s"`, opts.RequestID)
	}

	if opts.TraceID != "" {
		query += fmt.Sprintf(` |= "%s"`, opts.TraceID)
	}

	if opts.UserID != "" {
		query += fmt.Sprintf(` |= "%s"`, opts.UserID)
	}

	if opts.Search != "" {
		// Full-text search in message
		query += fmt.Sprintf(` |=~ "(?i)%s"`, opts.Search)
	}

	if opts.HideStaticAssets {
		// Exclude static asset logs
		for _, ext := range staticAssetExtensions {
			query += fmt.Sprintf(` != "%s"`, ext)
		}
	}

	return query
}

// toLogLine converts a log entry to a JSON string for Loki.
func (s *LokiLogStorage) toLogLine(entry *LogEntry) string {
	// Convert entry to JSON for storage
	data, _ := json.Marshal(entry)
	return string(data)
}

// parseLogLine parses a JSON log line back into a LogEntry.
func (s *LokiLogStorage) parseLogLine(line string) (*LogEntry, error) {
	var entry LogEntry
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// labelSetToString converts a label set to a string key for grouping.
func (s *LokiLogStorage) labelSetToString(labels map[string]string) string {
	// Sort keys for consistent ordering
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, labels[k]))
	}

	return strings.Join(parts, ",")
}

// getQueryLimit returns the limit to use for queries.
func (s *LokiLogStorage) getQueryLimit(limit int) int {
	if limit <= 0 {
		return 1000 // Default limit
	}
	if limit > 10000 {
		return 10000 // Max limit
	}
	return limit
}

// Compile-time check that LokiLogStorage implements LogStorage
var _ LogStorage = (*LokiLogStorage)(nil)
