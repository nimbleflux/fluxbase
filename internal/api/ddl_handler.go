package api

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/database"
	"github.com/nimbleflux/fluxbase/internal/logutil"
)

// DDLHandler handles Database Definition Language (DDL) operations
// for schema and table management
type DDLHandler struct {
	db *database.Connection
}

// NewDDLHandler creates a new DDL handler
func NewDDLHandler(db *database.Connection) *DDLHandler {
	return &DDLHandler{db: db}
}

// Validation patterns
var (
	// identifierPattern matches valid PostgreSQL identifiers (schema/table/column names)
	// Must start with letter or underscore, followed by letters, numbers, underscores
	identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

	// Reserved PostgreSQL keywords that should not be used as identifiers
	reservedKeywords = map[string]bool{
		"user": true, "table": true, "column": true, "index": true,
		"select": true, "insert": true, "update": true, "delete": true,
		"from": true, "where": true, "group": true, "order": true,
		"limit": true, "offset": true, "join": true, "on": true,
	}

	// Valid PostgreSQL data types
	validDataTypes = map[string]bool{
		"text": true, "varchar": true, "char": true,
		"integer": true, "bigint": true, "smallint": true,
		"numeric": true, "decimal": true, "real": true, "double precision": true,
		"boolean": true, "bool": true,
		"date": true, "timestamp": true, "timestamptz": true, "time": true, "timetz": true,
		"uuid": true, "json": true, "jsonb": true,
		"bytea": true, "inet": true, "cidr": true, "macaddr": true,
	}
)

// CreateSchemaRequest represents a request to create a new schema
type CreateSchemaRequest struct {
	Name string `json:"name"`
}

// CreateTableRequest represents a request to create a new table
type CreateTableRequest struct {
	Schema  string                `json:"schema"`
	Name    string                `json:"name"`
	Columns []CreateColumnRequest `json:"columns"`
}

// CreateColumnRequest represents a column definition
type CreateColumnRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	PrimaryKey   bool   `json:"primaryKey"`
	DefaultValue string `json:"defaultValue"`
}

// CreateSchema creates a new database schema
func (h *DDLHandler) CreateSchema(c fiber.Ctx) error {
	var req CreateSchemaRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate schema name
	if err := validateIdentifier(req.Name, "schema"); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()

	// Check if schema already exists
	exists, err := h.schemaExists(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Str("schema", req.Name).Msg("Failed to check schema existence")
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check schema existence",
		})
	}
	if exists {
		return c.Status(409).JSON(fiber.Map{
			"error": fmt.Sprintf("Schema '%s' already exists", req.Name),
		})
	}

	// Create schema (using quoted identifier for safety)
	// Use admin role to ensure full DDL access (superuser privileges)
	query := fmt.Sprintf("CREATE SCHEMA %s", quoteIdentifier(req.Name))
	queryMetadata := logutil.ExtractDDLMetadata(query)
	log.Info().Str("schema", req.Name).Str("operation", queryMetadata).Msg("Creating schema")

	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("schema", req.Name).Msg("Failed to create schema")
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create schema: %v", err),
		})
	}

	// Set up default privileges for tables created in this schema by the admin user
	// This ensures that future tables created via DDL API will automatically get grants to service_role
	if err := h.setupSchemaDefaultPrivileges(ctx, req.Name); err != nil {
		log.Error().Err(err).Str("schema", req.Name).Msg("Failed to set up default privileges")
		// Don't fail the request - schema was created successfully, just log the error
	}

	log.Info().Str("schema", req.Name).Msg("Schema created successfully")
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"schema":  req.Name,
		"message": fmt.Sprintf("Schema '%s' created successfully", req.Name),
	})
}

// CreateTable creates a new table with specified columns
func (h *DDLHandler) CreateTable(c fiber.Ctx) error {
	var req CreateTableRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate schema name
	if err := validateIdentifier(req.Schema, "schema"); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate table name
	if err := validateIdentifier(req.Name, "table"); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Validate columns
	if len(req.Columns) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "At least one column is required",
		})
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()

	// Check if schema exists
	exists, err := h.schemaExists(ctx, req.Schema)
	if err != nil {
		log.Error().Err(err).Str("schema", req.Schema).Msg("Failed to check schema existence")
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check schema existence",
		})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Schema '%s' does not exist", req.Schema),
		})
	}

	// Check if table already exists
	tableExists, err := h.tableExists(ctx, req.Schema, req.Name)
	if err != nil {
		log.Error().Err(err).Str("table", req.Schema+"."+req.Name).Msg("Failed to check table existence")
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check table existence",
		})
	}
	if tableExists {
		return c.Status(409).JSON(fiber.Map{
			"error": fmt.Sprintf("Table '%s.%s' already exists", req.Schema, req.Name),
		})
	}

	// Build CREATE TABLE statement
	query, err := h.buildCreateTableQuery(req)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Info().
		Str("table", req.Schema+"."+req.Name).
		Str("operation", logutil.ExtractDDLMetadata(query)).
		Int("columns", len(req.Columns)).
		Msg("Creating table")

	// Execute CREATE TABLE with admin role for full DDL access (superuser privileges)
	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("table", req.Schema+"."+req.Name).Msg("Failed to create table")
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create table: %v", err),
		})
	}

	// Grant permissions to service_role for instance_admin access
	// This is necessary because tables created via ExecuteWithAdminRole don't
	// inherit default privileges from migration 027 (which only applies to CURRENT_USER)
	if err := h.grantTablePermissions(ctx, req.Schema, req.Name); err != nil {
		log.Error().Err(err).Str("table", req.Schema+"."+req.Name).Msg("Failed to grant permissions to service_role")
		// Don't fail the request - table was created successfully, just log the error
	}

	log.Info().Str("table", req.Schema+"."+req.Name).Msg("Table created successfully")
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"schema":  req.Schema,
		"table":   req.Name,
		"message": fmt.Sprintf("Table '%s.%s' created successfully", req.Schema, req.Name),
	})
}

// DeleteTable drops a table from the database
func (h *DDLHandler) DeleteTable(c fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")

	// Validate identifiers
	if err := validateIdentifier(schema, "schema"); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if err := validateIdentifier(table, "table"); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()

	// Check if table exists
	exists, err := h.tableExists(ctx, schema, table)
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Msg("Failed to check table existence")
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check table existence",
		})
	}
	if !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Table '%s.%s' does not exist", schema, table),
		})
	}

	// Build DROP TABLE statement
	query := fmt.Sprintf("DROP TABLE %s.%s", quoteIdentifier(schema), quoteIdentifier(table))
	log.Info().Str("table", schema+"."+table).Str("operation", logutil.ExtractDDLMetadata(query)).Msg("Dropping table")

	// Execute DROP TABLE with admin role for full DDL access (superuser privileges)
	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Msg("Failed to drop table")
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to drop table: %v", err),
		})
	}

	log.Info().Str("table", schema+"."+table).Msg("Table dropped successfully")
	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Table '%s.%s' deleted successfully", schema, table),
	})
}

// AddColumnRequest represents a request to add a column to a table
type AddColumnRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	DefaultValue string `json:"defaultValue,omitempty"`
}

// AddColumn adds a new column to an existing table
func (h *DDLHandler) AddColumn(c fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")

	// Validate identifiers
	if err := validateIdentifier(schema, "schema"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}
	if err := validateIdentifier(table, "table"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	var req AddColumnRequest
	if err := c.Bind().Body(&req); err != nil {
		return SendInvalidBody(c)
	}

	// Validate column name
	if err := validateIdentifier(req.Name, "column"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	// Validate data type
	dataType := strings.ToLower(strings.TrimSpace(req.Type))
	if !validDataTypes[dataType] {
		return SendBadRequest(c, fmt.Sprintf("Invalid data type '%s'", req.Type), ErrCodeInvalidInput)
	}

	ctx := c.RequestCtx()

	// Check if table exists
	exists, err := h.tableExists(ctx, schema, table)
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Msg("Failed to check table existence")
		return SendOperationFailed(c, "check table existence")
	}
	if !exists {
		return SendNotFound(c, fmt.Sprintf("Table '%s.%s' does not exist", schema, table))
	}

	// Check if column already exists
	colExists, err := h.columnExists(ctx, schema, table, req.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check column existence")
		return SendOperationFailed(c, "check column existence")
	}
	if colExists {
		return SendConflict(c, fmt.Sprintf("Column '%s' already exists in table '%s.%s'", req.Name, schema, table), ErrCodeAlreadyExists)
	}

	// Build ALTER TABLE ADD COLUMN statement
	colDef := fmt.Sprintf("%s %s", quoteIdentifier(req.Name), dataType)
	if !req.Nullable {
		colDef += " NOT NULL"
	}
	if req.DefaultValue != "" {
		colDef += fmt.Sprintf(" DEFAULT %s", sanitizeDefaultValue(req.DefaultValue))
	}

	query := fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s",
		quoteIdentifier(schema), quoteIdentifier(table), colDef)

	log.Info().Str("table", schema+"."+table).Str("column", req.Name).Str("operation", logutil.ExtractDDLMetadata(query)).Msg("Adding column")

	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Str("column", req.Name).Msg("Failed to add column")
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to add column: %v", err),
		})
	}

	log.Info().Str("table", schema+"."+table).Str("column", req.Name).Msg("Column added successfully")
	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Column '%s' added to table '%s.%s'", req.Name, schema, table),
	})
}

// DropColumn removes a column from a table
func (h *DDLHandler) DropColumn(c fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")
	column := c.Params("column")

	// Validate identifiers
	if err := validateIdentifier(schema, "schema"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}
	if err := validateIdentifier(table, "table"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}
	if err := validateIdentifier(column, "column"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()

	// Check if table exists
	exists, err := h.tableExists(ctx, schema, table)
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Msg("Failed to check table existence")
		return SendOperationFailed(c, "check table existence")
	}
	if !exists {
		return SendNotFound(c, fmt.Sprintf("Table '%s.%s' does not exist", schema, table))
	}

	// Check if column exists
	colExists, err := h.columnExists(ctx, schema, table, column)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check column existence")
		return SendOperationFailed(c, "check column existence")
	}
	if !colExists {
		return SendNotFound(c, fmt.Sprintf("Column '%s' does not exist in table '%s.%s'", column, schema, table))
	}

	query := fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s",
		quoteIdentifier(schema), quoteIdentifier(table), quoteIdentifier(column))

	log.Info().Str("table", schema+"."+table).Str("column", column).Str("operation", logutil.ExtractDDLMetadata(query)).Msg("Dropping column")

	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Str("column", column).Msg("Failed to drop column")
		return SendInternalError(c, fmt.Sprintf("Failed to drop column: %v", err))
	}

	log.Info().Str("table", schema+"."+table).Str("column", column).Msg("Column dropped successfully")
	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Column '%s' dropped from table '%s.%s'", column, schema, table),
	})
}

// RenameTableRequest represents a request to rename a table
type RenameTableRequest struct {
	NewName string `json:"newName"`
}

// RenameTable renames a table
func (h *DDLHandler) RenameTable(c fiber.Ctx) error {
	schema := c.Params("schema")
	table := c.Params("table")

	// Validate identifiers
	if err := validateIdentifier(schema, "schema"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}
	if err := validateIdentifier(table, "table"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	var req RenameTableRequest
	if err := c.Bind().Body(&req); err != nil {
		return SendInvalidBody(c)
	}

	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	// Validate new table name
	if err := validateIdentifier(req.NewName, "table"); err != nil {
		return SendBadRequest(c, err.Error(), ErrCodeValidationFailed)
	}

	ctx := c.RequestCtx()

	// Check if source table exists
	exists, err := h.tableExists(ctx, schema, table)
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Msg("Failed to check table existence")
		return SendOperationFailed(c, "check table existence")
	}
	if !exists {
		return SendNotFound(c, fmt.Sprintf("Table '%s.%s' does not exist", schema, table))
	}

	// Check if target table name already exists
	targetExists, err := h.tableExists(ctx, schema, req.NewName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check target table existence")
		return SendOperationFailed(c, "check target table existence")
	}
	if targetExists {
		return SendConflict(c, fmt.Sprintf("Table '%s.%s' already exists", schema, req.NewName), ErrCodeAlreadyExists)
	}

	query := fmt.Sprintf("ALTER TABLE %s.%s RENAME TO %s",
		quoteIdentifier(schema), quoteIdentifier(table), quoteIdentifier(req.NewName))

	log.Info().Str("table", schema+"."+table).Str("newName", req.NewName).Str("operation", logutil.ExtractDDLMetadata(query)).Msg("Renaming table")

	err = h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, execErr := conn.Exec(ctx, query)
		return execErr
	})
	if err != nil {
		log.Error().Err(err).Str("table", schema+"."+table).Str("newName", req.NewName).Msg("Failed to rename table")
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to rename table: %v", err),
		})
	}

	log.Info().Str("table", schema+"."+table).Str("newName", req.NewName).Msg("Table renamed successfully")
	return c.JSON(fiber.Map{
		"success": true,
		"message": fmt.Sprintf("Table '%s.%s' renamed to '%s.%s'", schema, table, schema, req.NewName),
	})
}

// Helper functions

// validateIdentifier validates a PostgreSQL identifier (schema/table/column name)
func validateIdentifier(name, entityType string) error {
	if name == "" {
		return fmt.Errorf("%s name cannot be empty", entityType)
	}

	if len(name) > 63 {
		return fmt.Errorf("%s name cannot exceed 63 characters", entityType)
	}

	if !identifierPattern.MatchString(name) {
		return fmt.Errorf("%s name must start with a letter or underscore and contain only letters, numbers, and underscores", entityType)
	}

	// Check for reserved keywords
	if reservedKeywords[strings.ToLower(name)] {
		return fmt.Errorf("'%s' is a reserved keyword and cannot be used as a %s name", name, entityType)
	}

	return nil
}

// schemaExists checks if a schema exists
func (h *DDLHandler) schemaExists(ctx context.Context, schema string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)`
	err := h.db.Pool().QueryRow(ctx, query, schema).Scan(&exists)
	return exists, err
}

// tableExists checks if a table exists
func (h *DDLHandler) tableExists(ctx context.Context, schema, table string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = $2)`
	err := h.db.Pool().QueryRow(ctx, query, schema, table).Scan(&exists)
	return exists, err
}

// columnExists checks if a column exists in a table
func (h *DDLHandler) columnExists(ctx context.Context, schema, table, column string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema = $1 AND table_name = $2 AND column_name = $3)`
	err := h.db.Pool().QueryRow(ctx, query, schema, table, column).Scan(&exists)
	return exists, err
}

// buildCreateTableQuery constructs a CREATE TABLE query from the request
func (h *DDLHandler) buildCreateTableQuery(req CreateTableRequest) (string, error) {
	var columnDefs []string
	var primaryKeys []string

	for i, col := range req.Columns {
		// Validate column name
		if err := validateIdentifier(col.Name, "column"); err != nil {
			return "", fmt.Errorf("column %d: %w", i+1, err)
		}

		// Validate data type
		dataType := strings.ToLower(strings.TrimSpace(col.Type))
		if !validDataTypes[dataType] {
			return "", fmt.Errorf("column '%s': invalid data type '%s'", col.Name, col.Type)
		}

		// Build column definition
		colDef := fmt.Sprintf("%s %s", quoteIdentifier(col.Name), dataType)

		// Add NOT NULL constraint
		if !col.Nullable {
			colDef += " NOT NULL"
		}

		// Add DEFAULT value
		if col.DefaultValue != "" {
			colDef += fmt.Sprintf(" DEFAULT %s", sanitizeDefaultValue(col.DefaultValue))
		}

		columnDefs = append(columnDefs, colDef)

		// Track primary keys
		if col.PrimaryKey {
			primaryKeys = append(primaryKeys, quoteIdentifier(col.Name))
		}
	}

	// Add PRIMARY KEY constraint if any
	if len(primaryKeys) > 0 {
		columnDefs = append(columnDefs, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
	}

	// Build final CREATE TABLE statement
	query := fmt.Sprintf(
		"CREATE TABLE %s.%s (\n  %s\n)",
		quoteIdentifier(req.Schema),
		quoteIdentifier(req.Name),
		strings.Join(columnDefs, ",\n  "),
	)

	return query, nil
}

// safeDefaultFunctions is a set of PostgreSQL functions that are safe to use as DEFAULT values
// These functions are allowed to pass through without escaping
var safeDefaultFunctions = map[string]bool{
	// UUID functions
	"gen_random_uuid()":    true,
	"uuid_generate_v4()":   true,
	"uuid_generate_v1()":   true,
	"uuid_generate_v1mc()": true,
	"uuid_generate_v3()":   true,
	"uuid_generate_v5()":   true,
	// Date/Time functions
	"now()":                   true,
	"current_timestamp":       true,
	"CURRENT_TIMESTAMP":       true,
	"current_date":            true,
	"CURRENT_DATE":            true,
	"current_time":            true,
	"CURRENT_TIME":            true,
	"localtime":               true,
	"LOCALTIME":               true,
	"localtimestamp":          true,
	"LOCALTIMESTAMP":          true,
	"transaction_timestamp()": true,
	"statement_timestamp()":   true,
	"clock_timestamp()":       true,
	// Boolean
	"true":  true,
	"TRUE":  true,
	"false": true,
	"FALSE": true,
	// Null
	"NULL": true,
	"null": true,
}

// sanitizeDefaultValue sanitizes a DEFAULT value for SQL
// It returns safe SQL functions directly or escapes literal values
func sanitizeDefaultValue(value string) string {
	defaultVal := strings.TrimSpace(value)

	// Check if it's a safe function
	if safeDefaultFunctions[defaultVal] {
		return defaultVal
	}

	// Check for numeric literals (integers and floats)
	if _, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
		return defaultVal
	}
	if _, err := strconv.ParseFloat(defaultVal, 64); err == nil {
		return defaultVal
	}

	// Check for type casts with safe functions (e.g., "now()::date", "'2024-01-01'::date")
	if strings.Contains(defaultVal, "::") {
		parts := strings.SplitN(defaultVal, "::", 2)
		if len(parts) == 2 {
			baseValue := strings.TrimSpace(parts[0])
			castType := strings.TrimSpace(parts[1])
			// Validate the cast type is alphanumeric (prevent injection)
			if isValidCastType(castType) {
				// If base is a safe function, allow the cast
				if safeDefaultFunctions[baseValue] {
					return defaultVal
				}
				// If base is already a quoted string, allow the cast
				if strings.HasPrefix(baseValue, "'") && strings.HasSuffix(baseValue, "'") {
					return defaultVal
				}
			}
		}
	}

	// For all other values, escape as a string literal
	return escapeLiteral(defaultVal)
}

// isValidCastType checks if a cast type is valid (alphanumeric with allowed chars)
func isValidCastType(t string) bool {
	if t == "" {
		return false
	}
	for _, r := range t {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '[' && r != ']' && r != ' ' && r != ',' {
			return false
		}
	}
	return true
}

// escapeLiteral escapes a string literal for SQL
// This is a simple implementation - for production, consider using a proper SQL builder
func escapeLiteral(value string) string {
	// Replace single quotes with double single quotes
	escaped := strings.ReplaceAll(value, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}

// grantTablePermissions grants necessary permissions on a table to service_role
// This ensures that instance_admin (which maps to service_role) can access the table
func (h *DDLHandler) grantTablePermissions(ctx context.Context, schema, table string) error {
	// Grant SELECT, INSERT, UPDATE, DELETE on the table to service_role
	grantTableQuery := fmt.Sprintf(
		"GRANT SELECT, INSERT, UPDATE, DELETE ON %s.%s TO service_role",
		quoteIdentifier(schema),
		quoteIdentifier(table),
	)

	err := h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, grantTableQuery)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to grant table permissions: %w", err)
	}

	// Grant USAGE on all sequences for this table (for auto-increment/identity columns)
	// This query finds all sequences belonging to the table and grants USAGE
	grantSequencesQuery := `
		SELECT sequence_name
		FROM information_schema.sequences
		WHERE sequence_schema = $1
		  AND sequence_name LIKE $2
	`

	rows, err := h.db.Pool().Query(ctx, grantSequencesQuery, schema, table+"_%")
	if err != nil {
		// Don't fail if we can't query sequences - table permissions are already granted
		log.Debug().Err(err).Str("table", schema+"."+table).Msg("Failed to query sequences for table")
		return nil
	}
	defer rows.Close()

	var sequenceNames []string
	for rows.Next() {
		var seqName string
		if err := rows.Scan(&seqName); err != nil {
			continue
		}
		sequenceNames = append(sequenceNames, seqName)
	}

	// Grant USAGE on each sequence
	for _, seqName := range sequenceNames {
		grantSeqQuery := fmt.Sprintf(
			"GRANT USAGE, SELECT ON SEQUENCE %s.%s TO service_role",
			quoteIdentifier(schema),
			quoteIdentifier(seqName),
		)
		err := h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, grantSeqQuery)
			return err
		})
		if err != nil {
			log.Debug().Err(err).Str("sequence", schema+"."+seqName).Msg("Failed to grant sequence permissions")
		}
	}

	log.Debug().
		Str("table", schema+"."+table).
		Int("sequences_granted", len(sequenceNames)).
		Msg("Granted permissions to service_role for table")

	return nil
}

// setupSchemaDefaultPrivileges sets up default privileges for a schema
// so that tables created by the admin user automatically get grants to service_role
func (h *DDLHandler) setupSchemaDefaultPrivileges(ctx context.Context, schema string) error {
	// Set up default privileges for tables created in this schema
	// This ensures that future tables created via DDL API will automatically get grants to service_role
	queries := []string{
		// Grant ALL on future tables to service_role
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE CURRENT_USER IN SCHEMA %s GRANT ALL ON TABLES TO service_role", quoteIdentifier(schema)),
		// Grant USAGE on future functions to service_role
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE CURRENT_USER IN SCHEMA %s GRANT ALL ON FUNCTIONS TO service_role", quoteIdentifier(schema)),
		// Grant USAGE on future sequences to service_role
		fmt.Sprintf("ALTER DEFAULT PRIVILEGES FOR ROLE CURRENT_USER IN SCHEMA %s GRANT USAGE, SELECT ON SEQUENCES TO service_role", quoteIdentifier(schema)),
	}

	for _, query := range queries {
		err := h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, query)
			return err
		})
		if err != nil {
			return fmt.Errorf("failed to set up default privileges: %w", err)
		}
	}

	// Also grant USAGE on the schema itself to service_role, anon, and authenticated
	grantSchemaQuery := fmt.Sprintf("GRANT USAGE ON SCHEMA %s TO service_role, anon, authenticated", quoteIdentifier(schema))
	err := h.db.ExecuteWithAdminRole(ctx, func(conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, grantSchemaQuery)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to grant schema usage: %w", err)
	}

	log.Debug().Str("schema", schema).Msg("Set up default privileges for schema")

	return nil
}

// ListSchemas returns all user schemas (excluding system schemas)
func (h *DDLHandler) ListSchemas(c fiber.Ctx) error {
	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()

	schemas, err := h.db.Inspector().GetSchemas(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list schemas")
		return SendOperationFailed(c, "list schemas")
	}

	// Filter out system schemas and build response
	type schemaInfo struct {
		Name  string `json:"name"`
		Owner string `json:"owner"`
	}
	var result []schemaInfo
	for _, schema := range schemas {
		// Skip system schemas
		if schema == "information_schema" || schema == "pg_catalog" || schema == "pg_toast" {
			continue
		}
		result = append(result, schemaInfo{Name: schema, Owner: "postgres"})
	}

	return c.JSON(fiber.Map{"schemas": result})
}

// ListTables returns all tables, optionally filtered by schema
func (h *DDLHandler) ListTables(c fiber.Ctx) error {
	// Check if database connection is available
	if h.db == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not initialized",
		})
	}

	ctx := c.RequestCtx()
	schemaParam := c.Query("schema")

	var schemasToQuery []string

	if schemaParam != "" {
		// If schema parameter provided, query only that schema
		schemasToQuery = []string{schemaParam}
	} else {
		// Otherwise, get all schemas
		schemas, err := h.db.Inspector().GetSchemas(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to list schemas")
			return SendOperationFailed(c, "list schemas")
		}

		// Filter out system schemas
		for _, schema := range schemas {
			if schema == "information_schema" || schema == "pg_catalog" || schema == "pg_toast" {
				continue
			}
			schemasToQuery = append(schemasToQuery, schema)
		}
	}

	// Collect tables from requested schema(s)
	type tableInfo struct {
		Schema string `json:"schema"`
		Name   string `json:"name"`
	}
	var tables []tableInfo

	for _, schema := range schemasToQuery {
		dbTables, err := h.db.Inspector().GetAllTables(ctx, schema)
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to get tables from schema")
			continue
		}
		for _, t := range dbTables {
			tables = append(tables, tableInfo{Schema: t.Schema, Name: t.Name})
		}
	}

	return c.JSON(fiber.Map{"tables": tables})
}

// fiber:context-methods migrated
