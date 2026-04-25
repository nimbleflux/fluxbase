package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pg_query "github.com/pganalyze/pg_query_go/v6"
	"github.com/rs/zerolog/log"

	"github.com/nimbleflux/fluxbase/internal/auth"
	"github.com/nimbleflux/fluxbase/internal/middleware"
)

// SQLHandler handles SQL query execution for the admin SQL editor
type SQLHandler struct {
	db          *pgxpool.Pool
	authService *auth.Service
}

// NewSQLHandler creates a new SQL handler
func NewSQLHandler(db *pgxpool.Pool, authService *auth.Service) *SQLHandler {
	return &SQLHandler{
		db:          db,
		authService: authService,
	}
}

// ExecuteSQLRequest represents a SQL execution request
type ExecuteSQLRequest struct {
	Query string `json:"query"`
}

// SQLResult represents the result of a single SQL statement
type SQLResult struct {
	Columns         []string         `json:"columns,omitempty"`
	Rows            []map[string]any `json:"rows,omitempty"`
	RowCount        int              `json:"row_count"`
	AffectedRows    int64            `json:"affected_rows,omitempty"`
	ExecutionTimeMS float64          `json:"execution_time_ms"`
	Error           *string          `json:"error,omitempty"`
	Statement       string           `json:"statement"`
}

// ExecuteSQLResponse represents the response for SQL execution
type ExecuteSQLResponse struct {
	Results []SQLResult `json:"results"`
}

const (
	maxRowsPerQuery = 1000
	queryTimeout    = 30 * time.Second
)

// ExecuteSQL executes SQL queries provided by the user
// @Summary Execute SQL queries
// @Description Executes one or more SQL statements and returns results. Only accessible by dashboard admins.
// @Tags SQL
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query body ExecuteSQLRequest true "SQL query to execute"
// @Success 200 {object} ExecuteSQLResponse
// @Failure 400 {object} fiber.Map
// @Failure 401 {object} fiber.Map
// @Failure 500 {object} fiber.Map
// @Router /api/v1/admin/sql/execute [post]
func (h *SQLHandler) ExecuteSQL(c fiber.Ctx) error {
	// Parse request
	var req ExecuteSQLRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate query
	if strings.TrimSpace(req.Query) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query cannot be empty",
		})
	}

	// Get user information for audit logging
	userID, _ := GetUserID(c)
	userEmail, _ := GetUserEmail(c)

	// Split query into statements (basic split by semicolon)
	statements := splitSQLStatements(req.Query)

	// Get tenant context from middleware
	tenantID := middleware.GetTenantIDFromContext(c)
	tenantRole := middleware.GetTenantRoleFromContext(c)
	isInstanceAdmin := middleware.IsInstanceAdminFromContext(c)
	tenantSource := middleware.GetTenantSourceFromContext(c)

	// Determine if acting as tenant admin (explicit tenant context via header or JWT)
	actingAsTenantAdmin := tenantID != "" && (tenantSource == "header" || tenantSource == "jwt")

	// Check for impersonation token in custom header
	// This allows the admin to stay authenticated while executing queries as another user
	impersonationToken := c.Get("X-Impersonation-Token")

	// Log execution attempt with context
	log.Info().
		Str("user_id", userID).
		Str("user_email", userEmail).
		Bool("is_instance_admin", isInstanceAdmin).
		Bool("acting_as_tenant_admin", actingAsTenantAdmin).
		Str("tenant_id", tenantID).
		Str("tenant_role", tenantRole).
		Str("tenant_source", tenantSource).
		Bool("has_impersonation_token", impersonationToken != "").
		Str("query_preview", truncateString(req.Query, 100)).
		Msg("SQL query execution attempt")

	if impersonationToken != "" {
		// Trim any whitespace
		impersonationToken = strings.TrimSpace(impersonationToken)

		// Validate the impersonation token
		impersonationClaims, err := h.authService.ValidateToken(impersonationToken)
		if err != nil {
			log.Warn().
				Err(err).
				Str("user_id", userID).
				Msg("Invalid impersonation token in SQL query")
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid impersonation token",
			})
		}

		log.Info().
			Str("audit_user_id", userID).
			Str("impersonated_user_id", impersonationClaims.UserID).
			Str("impersonated_role", impersonationClaims.Role).
			Msg("SQL query execution with impersonation token")

		// Get pool for impersonated context
		pool := h.getPoolForQuery(c, req.Query)
		return h.executeWithRLSContext(c, pool, statements, impersonationClaims, tenantID, userID)
	}

	// Instance admin WITHOUT tenant context → full access (service_role)
	if isInstanceAdmin && !actingAsTenantAdmin {
		pool := h.getPoolForQuery(c, req.Query)
		return h.executeAsInstanceAdmin(c, pool, statements, userID)
	}

	// All other cases → RLS enforced with tenant context
	// - Instance admin WITH tenant context
	// - Tenant admin
	// - Regular user
	pool := h.getPoolForQuery(c, req.Query)

	// Get user claims from JWT
	var claims *auth.TokenClaims
	for _, key := range []string{"claims", "jwt_claims"} {
		if c.Locals(key) != nil {
			if jwtClaims, ok := c.Locals(key).(*auth.TokenClaims); ok {
				claims = jwtClaims
				break
			}
		}
	}

	// If no claims, create minimal claims from context
	if claims == nil {
		claims = &auth.TokenClaims{
			UserID: userID,
			Role:   "authenticated",
		}
		if tenantRole != "" {
			claims.Role = tenantRole
		}
	}

	return h.executeWithTenantRLS(c, pool, statements, claims, tenantID, userID, isInstanceAdmin)
}

// getPoolForQuery determines the appropriate database pool for a query.
// Priority: Branch > Tenant (when tenant has separate DB) > Main
//
// When a tenant has its own database (tenant_db_name is set), ALL queries route
// to the tenant pool — not just public schema queries. The tenant DB uses FDW to
// access shared tables (e.g., auth.users) from the main database, so cross-schema
// joins work natively on the tenant pool.
//
// For tenants that use the main database (default tenant), tenant_db is nil and
// we fall back to the main pool — preserving backward compatibility.
func (h *SQLHandler) getPoolForQuery(c fiber.Ctx, query string) *pgxpool.Pool {
	// 1. Check for branch pool (highest priority)
	if pool := middleware.GetBranchPool(c); pool != nil {
		log.Debug().Msg("Using branch pool for SQL execution")
		return pool
	}

	// 2. Check for tenant pool — route ALL queries when tenant has separate DB
	if tenantPool := middleware.GetTenantPool(c); tenantPool != nil {
		log.Debug().Msg("Using tenant pool for SQL execution")
		return tenantPool
	}

	// 3. Fall back to main database
	log.Debug().Msg("Using main pool for SQL execution")
	return h.db
}

// executeWithRLSContext executes SQL statements with Row Level Security context
// This is used for impersonation mode to test RLS policies
func (h *SQLHandler) executeWithRLSContext(c fiber.Ctx, pool *pgxpool.Pool, statements []string, claims *auth.TokenClaims, tenantID string, auditUserID string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	// Acquire a dedicated connection for setting session variables
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to acquire database connection: %v", err),
		})
	}
	defer conn.Release()

	// Build JWT claims JSON for request.jwt.claims setting
	claimsMap := map[string]any{
		"sub":          claims.UserID,
		"role":         claims.Role,
		"email":        claims.Email,
		"is_anonymous": claims.IsAnonymous,
	}
	if claims.SessionID != "" {
		claimsMap["session_id"] = claims.SessionID
	}
	if claims.UserMetadata != nil {
		claimsMap["user_metadata"] = claims.UserMetadata
	}
	if claims.AppMetadata != nil {
		claimsMap["app_metadata"] = claims.AppMetadata
	}
	// Add tenant context for RLS
	if tenantID != "" {
		claimsMap["tenant_id"] = tenantID
	}

	claimsJSON, err := json.Marshal(claimsMap)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to serialize JWT claims: %v", err),
		})
	}

	// Determine the database role to use - always enforce RLS in impersonation mode
	dbRole := "authenticated"
	if claims.Role == "anon" {
		dbRole = "anon"
	}

	log.Info().
		Str("audit_user_id", auditUserID).
		Str("impersonated_user_id", claims.UserID).
		Str("impersonated_role", dbRole).
		Str("tenant_id", tenantID).
		Msg("SQL query execution with RLS context (impersonation)")

	// Execute all statements within a transaction to maintain RLS context
	results := make([]SQLResult, 0, len(statements))

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to begin transaction: %v", err),
		})
	}
	defer func() { _ = tx.Rollback(ctx) }() // Will be no-op if committed

	// Set RLS session variables
	_, err = tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claimsJSON))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set JWT claims: %v", err),
		})
	}

	// Set tenant context for RLS policies
	if tenantID != "" {
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to set tenant context: %v", err),
			})
		}
	}

	// Set the role (use SET LOCAL to limit to current transaction)
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL ROLE %q", dbRole))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set role '%s': %v", dbRole, err),
		})
	}

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		result := h.executeStatementInTx(ctx, tx, stmt)
		results = append(results, result)

		if result.Error != nil {
			log.Warn().
				Str("audit_user_id", auditUserID).
				Str("statement", truncateString(stmt, 100)).
				Str("error", *result.Error).
				Msg("SQL query execution failed (impersonation RLS)")
		} else {
			log.Info().
				Str("audit_user_id", auditUserID).
				Str("statement", truncateString(stmt, 100)).
				Int("row_count", result.RowCount).
				Msg("SQL query executed successfully (impersonation RLS)")
		}
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to commit RLS transaction")
	}

	return c.JSON(ExecuteSQLResponse{
		Results: results,
	})
}

// executeAsInstanceAdmin executes SQL with service_role (full admin access)
// This is ONLY for instance admins WITHOUT tenant context
func (h *SQLHandler) executeAsInstanceAdmin(c fiber.Ctx, pool *pgxpool.Pool, statements []string, auditUserID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to acquire database connection: %v", err),
		})
	}
	defer conn.Release()

	log.Info().
		Str("audit_user_id", auditUserID).
		Str("role", "service_role").
		Msg("SQL query execution as instance admin (full access)")

	results := make([]SQLResult, 0, len(statements))

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to begin transaction: %v", err),
		})
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Set service_role for full admin access
	_, err = tx.Exec(ctx, "SET LOCAL ROLE service_role")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set service_role: %v", err),
		})
	}

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		result := h.executeStatementInTx(ctx, tx, stmt)
		results = append(results, result)

		if result.Error != nil {
			log.Warn().
				Str("audit_user_id", auditUserID).
				Str("statement", truncateString(stmt, 100)).
				Str("error", *result.Error).
				Msg("SQL query execution failed (instance admin)")
		} else {
			log.Info().
				Str("audit_user_id", auditUserID).
				Str("statement", truncateString(stmt, 100)).
				Int("row_count", result.RowCount).
				Msg("SQL query executed successfully (instance admin)")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to commit instance admin transaction")
	}

	return c.JSON(ExecuteSQLResponse{Results: results})
}

// executeWithTenantRLS executes SQL with tenant-scoped RLS context
// This enforces RLS policies with tenant isolation
func (h *SQLHandler) executeWithTenantRLS(c fiber.Ctx, pool *pgxpool.Pool, statements []string, claims *auth.TokenClaims, tenantID string, auditUserID string, isInstanceAdmin bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to acquire database connection: %v", err),
		})
	}
	defer conn.Release()

	// Build JWT claims with tenant context
	claimsMap := map[string]any{
		"sub":  claims.UserID,
		"role": claims.Role,
	}
	if claims.Email != "" {
		claimsMap["email"] = claims.Email
	}
	if claims.SessionID != "" {
		claimsMap["session_id"] = claims.SessionID
	}
	if claims.UserMetadata != nil {
		claimsMap["user_metadata"] = claims.UserMetadata
	}
	if claims.AppMetadata != nil {
		claimsMap["app_metadata"] = claims.AppMetadata
	}
	if tenantID != "" {
		claimsMap["tenant_id"] = tenantID
	}

	claimsJSON, err := json.Marshal(claimsMap)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to serialize JWT claims: %v", err),
		})
	}

	// Always use authenticated role (enforces RLS)
	// Even instance admins acting as tenant admins get RLS enforced
	dbRole := "authenticated"
	if claims.Role == "anon" {
		dbRole = "anon"
	}

	log.Info().
		Str("audit_user_id", auditUserID).
		Str("db_role", dbRole).
		Str("tenant_id", tenantID).
		Bool("is_instance_admin", isInstanceAdmin).
		Msg("SQL query execution with tenant RLS enforcement")

	results := make([]SQLResult, 0, len(statements))

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to begin transaction: %v", err),
		})
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Set RLS session variables
	_, err = tx.Exec(ctx, "SELECT set_config('request.jwt.claims', $1, true)", string(claimsJSON))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set JWT claims: %v", err),
		})
	}

	// Set tenant context for RLS policies (KEY for tenant isolation)
	if tenantID != "" {
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to set tenant context: %v", err),
			})
		}
	}

	// Set the role (enforces RLS)
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL ROLE %q", dbRole))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to set role '%s': %v", dbRole, err),
		})
	}

	// Execute each statement
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		result := h.executeStatementInTx(ctx, tx, stmt)
		results = append(results, result)

		if result.Error != nil {
			log.Warn().
				Str("audit_user_id", auditUserID).
				Str("tenant_id", tenantID).
				Str("statement", truncateString(stmt, 100)).
				Str("error", *result.Error).
				Msg("SQL query execution failed (tenant RLS)")
		} else {
			log.Info().
				Str("audit_user_id", auditUserID).
				Str("tenant_id", tenantID).
				Str("statement", truncateString(stmt, 100)).
				Int("row_count", result.RowCount).
				Msg("SQL query executed successfully (tenant RLS)")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Warn().Err(err).Msg("Failed to commit tenant RLS transaction")
	}

	return c.JSON(ExecuteSQLResponse{Results: results})
}

// executeStatementInTx executes a single SQL statement within a transaction
func (h *SQLHandler) executeStatementInTx(ctx context.Context, tx pgx.Tx, statement string) SQLResult {
	startTime := time.Now()

	result := SQLResult{
		Statement: statement,
	}

	// Execute query
	rows, err := tx.Query(ctx, statement)
	if err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
		result.ExecutionTimeMS = float64(time.Since(startTime).Milliseconds())
		return result
	}
	defer rows.Close()

	// Get column descriptions
	fieldDescriptions := rows.FieldDescriptions()
	columns := make([]string, len(fieldDescriptions))
	for i, fd := range fieldDescriptions {
		columns[i] = string(fd.Name)
	}
	result.Columns = columns

	// Check if this is a SELECT query (has columns)
	if len(columns) > 0 {
		// Read rows
		resultRows := make([]map[string]any, 0)
		rowCount := 0

		for rows.Next() {
			if rowCount >= maxRowsPerQuery {
				// Drain remaining rows but don't include them
				for rows.Next() {
					rowCount++
				}
				errorMsg := fmt.Sprintf("Result limited to %d rows (query returned %d rows)", maxRowsPerQuery, rowCount)
				result.Error = &errorMsg
				break
			}

			values, err := rows.Values()
			if err != nil {
				errorMsg := fmt.Sprintf("Error reading row: %v", err)
				result.Error = &errorMsg
				break
			}

			row := make(map[string]any)
			for i, col := range columns {
				row[col] = convertValue(values[i])
			}
			resultRows = append(resultRows, row)
			rowCount++
		}

		result.Rows = resultRows
		result.RowCount = len(resultRows)
	} else {
		// For non-SELECT queries (INSERT, UPDATE, DELETE, etc.)
		for rows.Next() {
			// Should not happen for non-SELECT, but drain just in case
		}
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		errorMsg := err.Error()
		result.Error = &errorMsg
	}

	// Get command tag for affected rows
	commandTag := rows.CommandTag()
	if len(columns) == 0 {
		result.AffectedRows = commandTag.RowsAffected()
		result.RowCount = int(commandTag.RowsAffected())
	}

	result.ExecutionTimeMS = float64(time.Since(startTime).Milliseconds())
	return result
}

// splitSQLStatements splits a SQL query string into individual statements
// using the PostgreSQL parser for correct handling of semicolons in
// strings, comments, and other contexts.
func splitSQLStatements(query string) []string {
	// Try parser-based splitting first for correct handling of
	// semicolons inside strings, comments, and other contexts
	stmts, err := pg_query.SplitWithParser(query, true)
	if err != nil {
		// Fall back to simple semicolon split if parser fails
		// (e.g., for incomplete queries in the editor)
		statements := strings.Split(query, ";")
		result := make([]string, 0, len(statements))
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				result = append(result, stmt)
			}
		}
		return result
	}

	if len(stmts) == 0 {
		return []string{}
	}

	return stmts
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// convertValue converts database values to JSON-friendly formats
// Specifically handles UUID byte arrays which pgx returns as [16]byte
func convertValue(v any) any {
	if v == nil {
		return nil
	}

	// Handle UUID: pgx returns UUIDs as [16]byte arrays
	if b, ok := v.([16]byte); ok {
		return formatUUID(b[:])
	}

	// Handle byte slices that might be UUIDs (some drivers return []byte)
	if b, ok := v.([]byte); ok && len(b) == 16 {
		// Check if it looks like a UUID (not printable ASCII)
		isPrintable := true
		for _, c := range b {
			if c < 32 || c > 126 {
				isPrintable = false
				break
			}
		}
		if !isPrintable {
			return formatUUID(b)
		}
	}

	return v
}

// formatUUID formats a 16-byte slice as a UUID string
func formatUUID(b []byte) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hex.EncodeToString(b[0:4]),
		hex.EncodeToString(b[4:6]),
		hex.EncodeToString(b[6:8]),
		hex.EncodeToString(b[8:10]),
		hex.EncodeToString(b[10:16]),
	)
}
