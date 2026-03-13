package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
)

// ClickHouseLogStorage implements LogStorage using ClickHouse.
// ClickHouse is optimized for analytical queries on large log datasets.
type ClickHouseLogStorage struct {
	conn      clickhouse.Conn
	tableName string
	ttlDays   int
}

// newClickHouseLogStorage creates a new ClickHouse-backed log storage.
func newClickHouseLogStorage(cfg LogStorageConfig) (*ClickHouseLogStorage, error) {
	// Set defaults
	addresses := cfg.ClickHouseAddresses
	if len(addresses) == 0 {
		addresses = []string{"localhost:9000"}
	}

	database := cfg.ClickHouseDatabase
	if database == "" {
		database = "fluxbase"
	}

	tableName := cfg.ClickHouseTable
	if tableName == "" {
		tableName = "logs"
	}

	ttlDays := cfg.ClickHouseTTL
	if ttlDays == 0 {
		ttlDays = 30 // Default TTL: 30 days
	}

	// Build connection options
	opts := &clickhouse.Options{
		Addr: addresses,
		Auth: clickhouse.Auth{
			Database: database,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionZSTD,
		},
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
	}

	if cfg.ClickHouseUsername != "" {
		opts.Auth.Username = cfg.ClickHouseUsername
	}
	if cfg.ClickHousePassword != "" {
		opts.Auth.Password = cfg.ClickHousePassword
	}

	// Connect to ClickHouse
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	storage := &ClickHouseLogStorage{
		conn:      conn,
		tableName: tableName,
		ttlDays:   ttlDays,
	}

	// Create table if not exists
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := storage.createTable(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return storage, nil
}

// Name returns the backend identifier.
func (s *ClickHouseLogStorage) Name() string {
	return "clickhouse"
}

// Write writes a batch of log entries to ClickHouse.
func (s *ClickHouseLogStorage) Write(ctx context.Context, entries []*LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	// Start a batch insert
	batch, err := s.conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", s.tableName))
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	// Add each entry to the batch
	for _, entry := range entries {
		// Ensure ID is set
		if entry.ID == uuid.Nil {
			entry.ID = uuid.New()
		}

		// Ensure timestamp is set
		if entry.Timestamp.IsZero() {
			entry.Timestamp = time.Now()
		}

		row := s.toRow(entry)
		if err := batch.Append(
			row.timestamp,
			row.id,
			row.category,
			row.level,
			row.message,
			row.component,
			row.userID,
			row.requestID,
			row.traceID,
			row.executionID,
			row.lineNumber,
			row.httpData,
			row.securityData,
			row.executionData,
			row.aiData,
			row.customData,
			row.customCategory,
		); err != nil {
			return fmt.Errorf("failed to append entry to batch: %w", err)
		}
	}

	// Flush the batch
	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	return nil
}

// Query retrieves logs matching the given options.
func (s *ClickHouseLogStorage) Query(ctx context.Context, opts LogQueryOptions) (*LogQueryResult, error) {
	queryBuild := s.buildQuery(opts)

	// Add ORDER BY, LIMIT, OFFSET
	order := "DESC"
	if opts.SortAsc {
		order = "ASC"
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	// First, get the count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", s.tableName, queryBuild.where)
	var totalCount int64
	if err := s.conn.QueryRow(ctx, countQuery, queryBuild.args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to count log entries: %w", err)
	}

	// Get entries
	selectQuery := fmt.Sprintf(`
		SELECT timestamp, id, category, level, message, component,
		       user_id, request_id, trace_id, execution_id, line_number,
		       http_data, security_data, execution_data, ai_data, custom_data, custom_category
		FROM %s %s
		ORDER BY timestamp %s
		LIMIT %d OFFSET %d`,
		s.tableName, queryBuild.where, order, limit, offset)

	rows, err := s.conn.Query(ctx, selectQuery, queryBuild.args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query log entries: %w", err)
	}
	defer func() { _ = rows.Close() }()

	entries := make([]*LogEntry, 0, limit)
	for rows.Next() {
		var entry LogEntry
		var httpData, securityData, executionData, aiData, customData map[string]string
		var userID *uuid.UUID
		var requestID, traceID *string
		var executionID *uuid.UUID
		var lineNumber *uint32
		var component, customCategory *string

		err := rows.Scan(
			&entry.Timestamp,
			&entry.ID,
			&entry.Category,
			&entry.Level,
			&entry.Message,
			&component,
			&userID,
			&requestID,
			&traceID,
			&executionID,
			&lineNumber,
			&httpData,
			&securityData,
			&executionData,
			&aiData,
			&customData,
			&customCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		// Convert nullable fields
		if component != nil {
			entry.Component = *component
		}
		if userID != nil {
			entry.UserID = userID.String()
		}
		if requestID != nil {
			entry.RequestID = *requestID
		}
		if traceID != nil {
			entry.TraceID = *traceID
		}
		if executionID != nil {
			entry.ExecutionID = executionID.String()
		}
		if lineNumber != nil {
			entry.LineNumber = int(*lineNumber)
		}
		if customCategory != nil {
			entry.CustomCategory = *customCategory
		}

		// Build fields map from category-specific data
		entry.Fields = make(map[string]any)
		for k, v := range httpData {
			entry.Fields[k] = v
		}
		for k, v := range securityData {
			entry.Fields[k] = v
		}
		for k, v := range executionData {
			entry.Fields[k] = v
		}
		for k, v := range aiData {
			entry.Fields[k] = v
		}
		for k, v := range customData {
			entry.Fields[k] = v
		}

		entries = append(entries, &entry)
	}

	return &LogQueryResult{
		Entries:    entries,
		TotalCount: totalCount,
		HasMore:    int64(offset+len(entries)) < totalCount,
	}, nil
}

// GetExecutionLogs retrieves logs for a specific execution.
func (s *ClickHouseLogStorage) GetExecutionLogs(ctx context.Context, executionID string, afterLine int) ([]*LogEntry, error) {
	execUUID, err := uuid.Parse(executionID)
	if err != nil {
		return nil, fmt.Errorf("invalid execution ID: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT timestamp, id, category, level, message, component,
		       user_id, request_id, trace_id, execution_id, line_number,
		       http_data, security_data, execution_data, ai_data, custom_data, custom_category
		FROM %s
		WHERE execution_id = ? AND line_number > ?
		ORDER BY line_number ASC`, s.tableName)

	rows, err := s.conn.Query(ctx, query, execUUID, afterLine)
	if err != nil {
		return nil, fmt.Errorf("failed to query execution logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var entries []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var httpData, securityData, executionData, aiData, customData map[string]string
		var userID *uuid.UUID
		var requestID, traceID *string
		var execID *uuid.UUID
		var lineNumber *uint32
		var component, customCategory *string

		err := rows.Scan(
			&entry.Timestamp,
			&entry.ID,
			&entry.Category,
			&entry.Level,
			&entry.Message,
			&component,
			&userID,
			&requestID,
			&traceID,
			&execID,
			&lineNumber,
			&httpData,
			&securityData,
			&executionData,
			&aiData,
			&customData,
			&customCategory,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan log entry: %w", err)
		}

		if component != nil {
			entry.Component = *component
		}
		if userID != nil {
			entry.UserID = userID.String()
		}
		if requestID != nil {
			entry.RequestID = *requestID
		}
		if traceID != nil {
			entry.TraceID = *traceID
		}
		if execID != nil {
			entry.ExecutionID = execID.String()
		}
		if lineNumber != nil {
			entry.LineNumber = int(*lineNumber)
		}
		if customCategory != nil {
			entry.CustomCategory = *customCategory
		}

		entry.Fields = make(map[string]any)
		for k, v := range httpData {
			entry.Fields[k] = v
		}
		for k, v := range securityData {
			entry.Fields[k] = v
		}
		for k, v := range executionData {
			entry.Fields[k] = v
		}
		for k, v := range aiData {
			entry.Fields[k] = v
		}
		for k, v := range customData {
			entry.Fields[k] = v
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// Delete removes logs matching the given options.
func (s *ClickHouseLogStorage) Delete(ctx context.Context, opts LogQueryOptions) (int64, error) {
	query := s.buildQuery(opts)
	if query.where == "" {
		return 0, fmt.Errorf("delete requires at least one filter condition")
	}

	// ClickHouse uses ALTER TABLE ... DELETE for deletions
	deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE %s", s.tableName, query.where)

	err := s.conn.Exec(ctx, deleteQuery, query.args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete log entries: %w", err)
	}

	// Note: ClickHouse doesn't return affected rows for ALTER DELETE
	// We would need to count before and after if exact count is needed
	return 0, nil
}

// Stats returns statistics about stored logs.
func (s *ClickHouseLogStorage) Stats(ctx context.Context) (*LogStats, error) {
	stats := &LogStats{
		EntriesByCategory: make(map[LogCategory]int64),
		EntriesByLevel:    make(map[LogLevel]int64),
	}

	// Get total count and time range
	var totalCount int64
	var oldestEntry, newestEntry *time.Time

	err := s.conn.QueryRow(ctx, fmt.Sprintf(`
		SELECT COUNT(*) as count,
		       MIN(timestamp) as oldest,
		       MAX(timestamp) as newest
		FROM %s`, s.tableName)).Scan(&totalCount, &oldestEntry, &newestEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to get log stats: %w", err)
	}

	stats.TotalEntries = totalCount
	if oldestEntry != nil {
		stats.OldestEntry = *oldestEntry
	}
	if newestEntry != nil {
		stats.NewestEntry = *newestEntry
	}

	// Get counts by category
	rows, err := s.conn.Query(ctx, fmt.Sprintf(`
		SELECT category, COUNT(*) as count
		FROM %s
		GROUP BY category`, s.tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get category counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var category string
		var count int64
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category count: %w", err)
		}
		stats.EntriesByCategory[LogCategory(category)] = count
	}

	// Get counts by level
	rows, err = s.conn.Query(ctx, fmt.Sprintf(`
		SELECT level, COUNT(*) as count
		FROM %s
		GROUP BY level`, s.tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to get level counts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var level string
		var count int64
		if err := rows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan level count: %w", err)
		}
		stats.EntriesByLevel[LogLevel(level)] = count
	}

	return stats, nil
}

// Health checks if the backend is operational.
func (s *ClickHouseLogStorage) Health(ctx context.Context) error {
	return s.conn.Ping(ctx)
}

// Close releases resources.
func (s *ClickHouseLogStorage) Close() error {
	return s.conn.Close()
}

// createTable creates the logs table if it doesn't exist.
func (s *ClickHouseLogStorage) createTable(ctx context.Context) error {
	ttlClause := ""
	if s.ttlDays > 0 {
		ttlClause = fmt.Sprintf("TTL timestamp + INTERVAL %d DAY", s.ttlDays)
	}

	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			timestamp DateTime,
			id UUID,
			category String,
			level String,
			message String,
			component String,
			user_id Nullable(UUID),
			request_id Nullable(String),
			trace_id Nullable(String),
			execution_id Nullable(UUID),
			line_number Nullable(UInt32),
			http_data Map(String, String),
			security_data Map(String, String),
			execution_data Map(String, String),
			ai_data Map(String, String),
			custom_data Map(String, String),
			custom_category String,
			INDEX idx_level level TYPE bloom_filter GRANULARITY 1,
			INDEX idx_message message TYPE tokenbf_v1(32768, 3, 0) GRANULARITY 1,
			INDEX idx_component component TYPE bloom_filter GRANULARITY 1
		) ENGINE = MergeTree()
		ORDER BY (timestamp, category)
		%s
	`, s.tableName, ttlClause)

	err := s.conn.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// logRow represents a ClickHouse log row.
type logRow struct {
	timestamp      time.Time
	id             uuid.UUID
	category       string
	level          string
	message        string
	component      *string
	userID         *uuid.UUID
	requestID      *string
	traceID        *string
	executionID    *uuid.UUID
	lineNumber     *uint32
	httpData       map[string]string
	securityData   map[string]string
	executionData  map[string]string
	aiData         map[string]string
	customData     map[string]string
	customCategory string
}

// toRow converts a LogEntry to a ClickHouse logRow.
func (s *ClickHouseLogStorage) toRow(entry *LogEntry) logRow {
	row := logRow{
		timestamp:      entry.Timestamp,
		id:             entry.ID,
		category:       string(entry.Category),
		level:          string(entry.Level),
		message:        entry.Message,
		customCategory: entry.CustomCategory,
		httpData:       make(map[string]string),
		securityData:   make(map[string]string),
		executionData:  make(map[string]string),
		aiData:         make(map[string]string),
		customData:     make(map[string]string),
	}

	// Nullable fields
	if entry.Component != "" {
		row.component = &entry.Component
	}
	if entry.UserID != "" {
		if uid, err := uuid.Parse(entry.UserID); err == nil {
			row.userID = &uid
		}
	}
	if entry.RequestID != "" {
		row.requestID = &entry.RequestID
	}
	if entry.TraceID != "" {
		row.traceID = &entry.TraceID
	}
	if entry.ExecutionID != "" {
		if eid, err := uuid.Parse(entry.ExecutionID); err == nil {
			row.executionID = &eid
		}
	}
	if entry.LineNumber > 0 {
		ln := uint32(entry.LineNumber)
		row.lineNumber = &ln
	}

	// Category-specific data
	for k, v := range entry.Fields {
		var strVal string
		if v == nil {
			strVal = ""
		} else {
			strVal = fmt.Sprintf("%v", v)
		}

		switch entry.Category {
		case LogCategoryHTTP:
			row.httpData[k] = strVal
		case LogCategorySecurity:
			row.securityData[k] = strVal
		case LogCategoryExecution:
			row.executionData[k] = strVal
		case LogCategoryAI:
			row.aiData[k] = strVal
		case LogCategoryCustom:
			row.customData[k] = strVal
		default:
			// System logs go to custom data
			row.customData[k] = strVal
		}
	}

	return row
}

// queryBuild holds the WHERE clause and arguments for a query.
type queryBuild struct {
	where string
	args  []any
}

// buildQuery builds a WHERE clause and arguments from query options.
func (s *ClickHouseLogStorage) buildQuery(opts LogQueryOptions) queryBuild {
	var conditions []string
	var args []any

	if opts.Category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, string(opts.Category))
	}

	if opts.CustomCategory != "" {
		conditions = append(conditions, "custom_category = ?")
		args = append(args, opts.CustomCategory)
	}

	if len(opts.Levels) > 0 {
		placeholders := make([]string, len(opts.Levels))
		for i, level := range opts.Levels {
			placeholders[i] = "?"
			args = append(args, string(level))
		}
		conditions = append(conditions, fmt.Sprintf("level IN (%s)", joinPlaceholders(placeholders)))
	}

	if opts.Component != "" {
		conditions = append(conditions, "component = ?")
		args = append(args, opts.Component)
	}

	if opts.RequestID != "" {
		conditions = append(conditions, "request_id = ?")
		args = append(args, opts.RequestID)
	}

	if opts.TraceID != "" {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, opts.TraceID)
	}

	if opts.UserID != "" {
		if userUUID, err := uuid.Parse(opts.UserID); err == nil {
			conditions = append(conditions, "user_id = ?")
			args = append(args, userUUID)
		}
	}

	if opts.ExecutionID != "" {
		if execUUID, err := uuid.Parse(opts.ExecutionID); err == nil {
			conditions = append(conditions, "execution_id = ?")
			args = append(args, execUUID)
		}
	}

	// ExecutionType is in execution_data map
	if opts.ExecutionType != "" {
		conditions = append(conditions, "execution_data['execution_type'] = ?")
		args = append(args, opts.ExecutionType)
	}

	if !opts.StartTime.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, opts.StartTime)
	}

	if !opts.EndTime.IsZero() {
		conditions = append(conditions, "timestamp <= ?")
		args = append(args, opts.EndTime)
	}

	if opts.Search != "" {
		// ClickHouse uses positionUTF8 for simple search
		conditions = append(conditions, "positionUTF8CaseInsensitive(message, ?) > 0")
		args = append(args, opts.Search)
	}

	if opts.AfterLine > 0 {
		conditions = append(conditions, "line_number > ?")
		args = append(args, opts.AfterLine)
	}

	if opts.HideStaticAssets && opts.Category == LogCategoryHTTP {
		// Exclude logs where the HTTP path ends with static asset extensions
		conditions = append(conditions, "http_data['path'] NOT LIKE ?")
		args = append(args, "%.js")
		conditions = append(conditions, "http_data['path'] NOT LIKE ?")
		args = append(args, "%.css")
		conditions = append(conditions, "http_data['path'] NOT LIKE ?")
		args = append(args, "%.png")
		conditions = append(conditions, "http_data['path'] NOT LIKE ?")
		args = append(args, "%.jpg")
		// Add more extensions as needed
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + joinConditions(conditions)
	}

	return queryBuild{where: where, args: args}
}

// joinConditions joins conditions with AND.
func joinConditions(conds []string) string {
	result := ""
	for i, cond := range conds {
		if i > 0 {
			result += " AND "
		}
		result += cond
	}
	return result
}

// joinPlaceholders joins placeholders with comma.
func joinPlaceholders(phs []string) string {
	result := ""
	for i, ph := range phs {
		if i > 0 {
			result += ", "
		}
		result += ph
	}
	return result
}

// Compile-time check that ClickHouseLogStorage implements LogStorage
var _ LogStorage = (*ClickHouseLogStorage)(nil)
