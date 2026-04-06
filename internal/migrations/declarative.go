package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgplex/pgparser/nodes"
	"github.com/pgplex/pgparser/parser"
	"github.com/rs/zerolog/log"
)

// DeclarativeService manages Fluxbase internal schema using pgschema
type DeclarativeService struct {
	pgschemaPath string
	dbHost       string
	dbPort       int
	dbUser       string
	dbPassword   string
	dbName       string
	config       DeclarativeConfig
	pool         *pgxpool.Pool // Optional: for recording state
}

// DeclarativeConfig holds configuration for declarative schema management
type DeclarativeConfig struct {
	SchemaDir        string   // Directory containing per-schema SQL files (e.g., internal/database/schema/schemas/)
	Schemas          []string // Schemas to manage (Fluxbase internal only)
	AllowDestructive bool     // Allow destructive changes
	LockTimeout      int      // Lock timeout in seconds
}

// DefaultFluxbaseSchemas lists all Fluxbase internal schemas in dependency order
// platform must come first as auth FKs reference platform.tenants
// auth must come early as other schemas reference auth.users
var DefaultFluxbaseSchemas = []string{
	"platform", "auth", "storage", "jobs", "functions", "realtime",
	"ai", "rpc", "system", "migrations",
	"app", "api", "branching", "logging", "mcp",
}

// NewDeclarativeService creates a new declarative schema service
func NewDeclarativeService(pgschemaPath string, dbHost string, dbPort int, dbUser, dbPassword, dbName string, config DeclarativeConfig) *DeclarativeService {
	return &DeclarativeService{
		pgschemaPath: pgschemaPath,
		dbHost:       dbHost,
		dbPort:       dbPort,
		dbUser:       dbUser,
		dbPassword:   dbPassword,
		dbName:       dbName,
		config:       config,
	}
}

// SetPool sets the database pool for state recording
func (s *DeclarativeService) SetPool(pool *pgxpool.Pool) {
	s.pool = pool
}

// PlanForSchema generates a migration plan for a single schema
func (s *DeclarativeService) PlanForSchema(ctx context.Context, schema string) (*Plan, error) {
	schemaFile := filepath.Join(s.config.SchemaDir, schema+".sql")

	// Check if schema file exists
	if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("schema file not found: %s", schemaFile)
	}

	args := []string{
		"plan",
		"--host", s.dbHost,
		"--port", fmt.Sprintf("%d", s.dbPort),
		"--user", s.dbUser,
		"--db", s.dbName,
		"--file", schemaFile,
		"--schema", schema,
		"--output-json", "stdout",
		// Use the actual database for plan validation (needed for roles, extensions, etc.)
		"--plan-host", s.dbHost,
		"--plan-port", fmt.Sprintf("%d", s.dbPort),
		"--plan-user", s.dbUser,
		"--plan-password", s.dbPassword,
		"--plan-db", s.dbName,
	}

	if s.config.AllowDestructive {
		args = append(args, "--allow-destructive")
	}

	cmd := exec.CommandContext(ctx, s.pgschemaPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	start := time.Now()
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("pgschema plan failed for schema %s: %w: %s", schema, err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("pgschema plan failed for schema %s: %w: %s", schema, err, string(output))
	}

	var plan Plan
	if err := json.Unmarshal(output, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan for schema %s: %w", schema, err)
	}

	// Extract changes from groups/steps structure
	plan.Changes = extractChangesFromGroups(&plan)

	plan.Duration = time.Since(start)
	return &plan, nil
}

// extractChangesFromGroups converts pgschema's groups/steps structure to a flat Changes slice
func extractChangesFromGroups(plan *Plan) []Change {
	var changes []Change
	for _, group := range plan.Groups {
		for _, step := range group.Steps {
			// Skip directive-only steps (like wait for index)
			if step.SQL == "" {
				continue
			}

			change := Change{
				SQL: step.SQL,
			}

			// Parse path to extract schema, object type, and name
			// Path format: "schema.object_type.name" or "schema.object_type.constraint_name"
			parts := strings.Split(step.Path, ".")
			if len(parts) >= 1 {
				change.Schema = parts[0]
			}
			if len(parts) >= 2 {
				change.ObjectType = parts[1]
			}
			if len(parts) >= 3 {
				change.Name = strings.Join(parts[2:], ".")
			}

			// Convert operation to ChangeType
			switch step.Operation {
			case "create":
				change.Type = ChangeCreate
			case "drop":
				change.Type = ChangeDrop
				change.Destructive = true
			case "alter":
				change.Type = ChangeAlter
			}

			changes = append(changes, change)
		}
	}
	return changes
}

// Plan generates a combined migration plan for all schemas
func (s *DeclarativeService) Plan(ctx context.Context) (*Plan, error) {
	combined := &Plan{
		Changes: []Change{},
	}

	for _, schema := range s.config.Schemas {
		plan, err := s.PlanForSchema(ctx, schema)
		if err != nil {
			return nil, err
		}
		combined.Changes = append(combined.Changes, plan.Changes...)
		combined.Duration += plan.Duration
	}

	return combined, nil
}

// ApplyForSchema applies the migration plan for a single schema
func (s *DeclarativeService) ApplyForSchema(ctx context.Context, schema string, autoApprove bool) (*ApplyResult, error) {
	// First, generate plan
	plan, err := s.PlanForSchema(ctx, schema)
	if err != nil {
		return nil, err
	}

	if len(plan.Changes) == 0 {
		return &ApplyResult{Applied: []Change{}, Duration: 0}, nil
	}

	// Write plan to temp file
	planFile, err := s.writePlanToTemp(plan)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(planFile) }()

	args := []string{
		"apply",
		"--host", s.dbHost,
		"--port", fmt.Sprintf("%d", s.dbPort),
		"--user", s.dbUser,
		"--db", s.dbName,
		"--schema", schema,
		"--plan", planFile,
	}

	if autoApprove {
		args = append(args, "--auto-approve")
	}

	cmd := exec.CommandContext(ctx, s.pgschemaPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))

	start := time.Now()
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("pgschema apply failed for schema %s: %w: %s", schema, err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("pgschema apply failed for schema %s: %w: %s", schema, err, string(output))
	}

	return &ApplyResult{
		Applied:  plan.Changes,
		Duration: time.Since(start),
	}, nil
}

// applySchemaDirect applies schema changes directly using psql instead of pgschema
// This is used for schemas with cross-schema references that can't be validated in isolation
// pgschema validates schema SQL in a temporary schema, even with --plan-host, which fails
// for schemas that reference tables defined later in the same file or in other schemas.
// Apply executes the migration plan for all schemas
func (s *DeclarativeService) Apply(ctx context.Context, autoApprove bool) (*ApplyResult, error) {
	combined := &ApplyResult{
		Applied: []Change{},
	}

	for _, schema := range s.config.Schemas {
		result, err := s.ApplyForSchema(ctx, schema, autoApprove)
		if err != nil {
			return nil, err
		}
		combined.Applied = append(combined.Applied, result.Applied...)
		combined.Duration += result.Duration
	}

	// Record in declarative_state table
	if err := s.recordApply(ctx, combined); err != nil {
		log.Warn().Err(err).Msg("Failed to record declarative state")
	}

	return combined, nil
}

// Dump exports current database schema for a single schema
func (s *DeclarativeService) DumpForSchema(ctx context.Context, schema string, outputPath string) error {
	// Create output directory if needed
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Open output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() { _ = f.Close() }()

	args := []string{
		"dump",
		"--host", s.dbHost,
		"--port", fmt.Sprintf("%d", s.dbPort),
		"--user", s.dbUser,
		"--db", s.dbName,
		"--schema", schema,
	}

	cmd := exec.CommandContext(ctx, s.pgschemaPath, args...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", s.dbPassword))
	cmd.Stdout = f

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("pgschema dump failed for schema %s: %w: %s", schema, err, string(exitErr.Stderr))
		}
		return fmt.Errorf("pgschema dump failed for schema %s: %w", schema, err)
	}

	log.Info().Str("schema", schema).Str("path", outputPath).Msg("Schema dump completed")
	return nil
}

// Dump exports all schemas to separate files in the schema directory
func (s *DeclarativeService) Dump(ctx context.Context, schemaDir string) error {
	// Create output directory if needed
	if err := os.MkdirAll(schemaDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, schema := range s.config.Schemas {
		outputPath := filepath.Join(schemaDir, schema+".sql")
		if err := s.DumpForSchema(ctx, schema, outputPath); err != nil {
			return err
		}
	}

	log.Info().Str("dir", schemaDir).Int("schemas", len(s.config.Schemas)).Msg("All schemas dumped")
	return nil
}

// Validate checks for schema drift across all schemas
func (s *DeclarativeService) Validate(ctx context.Context) (*ValidationResult, error) {
	combined := &ValidationResult{Valid: true}

	for _, schema := range s.config.Schemas {
		plan, err := s.PlanForSchema(ctx, schema)
		if err != nil {
			return &ValidationResult{Valid: false, Error: err}, nil
		}

		if len(plan.Changes) > 0 {
			combined.Valid = false
			for _, change := range plan.Changes {
				combined.Drifts = append(combined.Drifts, Drift{
					Type:        string(change.Type),
					ObjectType:  change.ObjectType,
					Schema:      change.Schema,
					Name:        change.Name,
					SQL:         change.SQL,
					Destructive: change.Destructive,
				})
			}
		}
	}

	return combined, nil
}

// CalculateFingerprint computes a SHA256 hash of all schema files
func (s *DeclarativeService) CalculateFingerprint() (string, error) {
	hasher := sha256.New()

	for _, schema := range s.config.Schemas {
		schemaFile := filepath.Join(s.config.SchemaDir, schema+".sql")
		content, err := os.ReadFile(schemaFile)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip missing schema files
			}
			return "", fmt.Errorf("failed to read schema file %s: %w", schemaFile, err)
		}
		hasher.Write(content)
		hasher.Write([]byte("|")) // Separator between schemas
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// writePlanToTemp writes the plan JSON to a temporary file
func (s *DeclarativeService) writePlanToTemp(plan *Plan) (string, error) {
	tmpFile, err := os.CreateTemp("", "pgschema-plan-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() { _ = tmpFile.Close() }()

	planJSON, err := json.Marshal(plan)
	if err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to marshal plan: %w", err)
	}

	if _, err := tmpFile.Write(planJSON); err != nil {
		_ = os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write plan: %w", err)
	}

	return tmpFile.Name(), nil
}

// recordApply records the applied schema state in the database with default source
func (s *DeclarativeService) recordApply(ctx context.Context, result *ApplyResult) error {
	return s.recordApplyWithSource(ctx, result, "schema_apply")
}

// recordApplyWithSource records the applied schema state in the database with specified source
func (s *DeclarativeService) recordApplyWithSource(ctx context.Context, result *ApplyResult, source string) error {
	fingerprint, err := s.CalculateFingerprint()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to calculate fingerprint")
		fingerprint = "unknown"
	}

	log.Info().
		Str("fingerprint", fingerprint).
		Str("source", source).
		Int("changes", len(result.Applied)).
		Msg("Schema applied")

	// Record to database if pool is available
	if s.pool != nil {
		_, err := s.pool.Exec(ctx, `
			INSERT INTO migrations.declarative_state (schema_fingerprint, applied_by, source)
			VALUES ($1, 'fluxbase', $2)
		`, fingerprint, source)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to record declarative state to database")
		}
	}

	return nil
}

// ApplyDeclarative applies the declarative schema on startup with default source.
// It plans and applies any pending schema changes automatically.
// This is the main entry point for automatic schema management.
func (s *DeclarativeService) ApplyDeclarative(ctx context.Context) error {
	return s.ApplyDeclarativeWithSource(ctx, "schema_apply")
}

// ApplyDeclarativeWithSource applies the declarative schema on startup with specified source.
// It uses pgschema for proper schema diffing and evolution.
// The source parameter indicates the origin: 'fresh_install', 'transitioned', or 'schema_apply'.
func (s *DeclarativeService) ApplyDeclarativeWithSource(ctx context.Context, source string) error {
	log.Info().Str("schema_dir", s.config.SchemaDir).Msg("Checking declarative schema...")

	// Check if schema directory exists
	if _, err := os.Stat(s.config.SchemaDir); os.IsNotExist(err) {
		return fmt.Errorf("schema directory not found: %s", s.config.SchemaDir)
	}

	// Phase 1: Apply each schema using pgschema for proper diffing
	// pgschema handles schema evolution correctly by:
	// - Detecting existing tables and only adding new columns
	// - Generating proper ALTER TABLE statements
	// - Handling index and constraint changes
	for _, schema := range s.config.Schemas {
		schemaFile := filepath.Join(s.config.SchemaDir, schema+".sql")

		// Check if schema file exists
		if _, err := os.Stat(schemaFile); os.IsNotExist(err) {
			log.Debug().Str("schema", schema).Msg("Schema file not found, skipping")
			continue
		}

		// Generate plan first to see if there are any changes
		plan, err := s.PlanForSchema(ctx, schema)
		if err != nil {
			// If plan fails (e.g., due to cross-schema FK validation in temporary schema),
			// fall back to direct schema application with idempotent transforms
			log.Warn().Err(err).Str("schema", schema).Msg("pgschema plan failed, using direct fallback")
			if err := s.applySchemaDirectFallback(ctx, schema); err != nil {
				return fmt.Errorf("failed to apply schema %s: %w", schema, err)
			}
			log.Info().Str("schema", schema).Msg("Schema applied via direct fallback")
			continue
		}

		// Filter out FK drops for cross-schema constraints managed by post-schema-fks.sql.
		// These are applied separately in Phase 2 and should not be dropped during per-schema apply.
		plan.Changes = s.filterManagedFKDrops(plan.Changes)

		if len(plan.Changes) == 0 {
			log.Debug().Str("schema", schema).Msg("No schema changes needed")
			continue
		}

		// Apply the filtered plan directly using SQL execution.
		// We use applyPlanDirectly instead of ApplyForSchema because ApplyForSchema
		// regenerates the plan internally, which would not include our FK filtering.
		if err := s.applyPlanDirectly(ctx, schema, plan); err != nil {
			return fmt.Errorf("failed to apply schema %s: %w", schema, err)
		}
		log.Info().Str("schema", schema).Int("changes", len(plan.Changes)).Msg("Schema changes applied via plan execution")
	}

	// Phase 2: Apply cross-schema FKs from post-schema-fks.sql
	if err := s.applyCrossSchemaFKs(ctx); err != nil {
		return fmt.Errorf("failed to apply cross-schema FKs: %w", err)
	}

	// Phase 3: Apply post-schema.sql for cross-schema policies
	if err := s.applyPostSchemaPolicies(ctx); err != nil {
		return fmt.Errorf("failed to apply post-schema: %w", err)
	}

	// Record the schema state with the specified source
	if err := s.recordApplyWithSource(ctx, &ApplyResult{Applied: []Change{}}, source); err != nil {
		log.Warn().Err(err).Msg("Failed to record schema state")
	}

	return nil
}

// applySchemaDirectFallback applies a schema file directly (fallback when pgschema fails)
func (s *DeclarativeService) applySchemaDirectFallback(ctx context.Context, schema string) error {
	schemaFile := filepath.Join(s.config.SchemaDir, schema+".sql")

	// Read the schema file
	content, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Create connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		s.dbUser, s.dbPassword, s.dbHost, s.dbPort, s.dbName)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Set search_path to include all schemas for function/type references
	// Schemas are applied in dependency order, so earlier schemas are already created
	allSchemas := strings.Join(s.config.Schemas, ", ")
	searchPath := fmt.Sprintf("%s, %s, public", schema, allSchemas)
	_, err = pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s", searchPath))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	idempotentSQL := makeSQLIdempotent(string(content))
	_, err = pool.Exec(ctx, idempotentSQL)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	log.Info().Str("schema", schema).Msg("Schema applied directly")
	return nil
}

// applyPlanDirectly executes the SQL statements from a plan directly
// This is used when pgschema apply fails due to validation issues but the plan is valid
func (s *DeclarativeService) applyPlanDirectly(ctx context.Context, schema string, plan *Plan) error {
	if len(plan.Changes) == 0 {
		return nil
	}

	// Create connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		s.dbUser, s.dbPassword, s.dbHost, s.dbPort, s.dbName)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Set search_path to include all schemas for cross-schema references
	// The target schema comes first, then all other schemas for FK references
	allSchemas := strings.Join(s.config.Schemas, ", ")
	searchPath := fmt.Sprintf("%s, %s, public", schema, allSchemas)
	_, err = pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s", searchPath))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Execute each change's SQL statement
	for i, change := range plan.Changes {
		if change.SQL == "" {
			continue
		}

		// Log the change being applied (truncate SQL if too long)
		sqlPreview := change.SQL
		if len(sqlPreview) > 200 {
			sqlPreview = sqlPreview[:197] + "..."
		}
		log.Info().
			Str("schema", schema).
			Int("change_num", i+1).
			Int("total_changes", len(plan.Changes)).
			Str("type", string(change.Type)).
			Str("object", change.Name).
			Str("sql", sqlPreview).
			Msg("Applying schema change")

		_, err := pool.Exec(ctx, change.SQL)
		if err != nil {
			// Log the error with context
			log.Error().
				Err(err).
				Str("schema", schema).
				Str("sql", change.SQL).
				Msg("Failed to execute plan SQL")
			return fmt.Errorf("failed to execute plan SQL for schema %s (change %d/%d): %w", schema, i+1, len(plan.Changes), err)
		}
	}

	log.Info().Str("schema", schema).Int("changes", len(plan.Changes)).Msg("Plan SQL executed successfully")
	return nil
}

// applyCrossSchemaFKs applies cross-schema foreign keys from post-schema-fks.sql
func (s *DeclarativeService) applyCrossSchemaFKs(ctx context.Context) error {
	fksFile := filepath.Join(s.config.SchemaDir, "post-schema-fks.sql")

	// Check if file exists
	if _, err := os.Stat(fksFile); os.IsNotExist(err) {
		log.Debug().Msg("post-schema-fks.sql not found, skipping")
		return nil
	}

	// Read the file
	content, err := os.ReadFile(fksFile)
	if err != nil {
		return fmt.Errorf("failed to read post-schema-fks.sql: %w", err)
	}

	// Create connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		s.dbUser, s.dbPassword, s.dbHost, s.dbPort, s.dbName)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Execute the FK additions (they use idempotent DO blocks)
	_, err = pool.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to apply cross-schema FKs: %w", err)
	}

	log.Info().Msg("Cross-schema FKs applied")
	return nil
}

// applyPostSchemaPolicies applies post-schema.sql for cross-schema policies
func (s *DeclarativeService) applyPostSchemaPolicies(ctx context.Context) error {
	postSchemaFile := filepath.Join(s.config.SchemaDir, "post-schema.sql")

	// Check if file exists
	if _, err := os.Stat(postSchemaFile); os.IsNotExist(err) {
		log.Debug().Msg("post-schema.sql not found, skipping")
		return nil
	}

	// Read the file
	content, err := os.ReadFile(postSchemaFile)
	if err != nil {
		return fmt.Errorf("failed to read post-schema.sql: %w", err)
	}

	// Create connection
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		s.dbUser, s.dbPassword, s.dbHost, s.dbPort, s.dbName)
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}
	defer pool.Close()

	// Build search_path with all schemas
	allSchemas := strings.Join(s.config.Schemas, ", ")
	_, err = pool.Exec(ctx, fmt.Sprintf("SET search_path TO %s, public", allSchemas))
	if err != nil {
		return fmt.Errorf("failed to set search_path: %w", err)
	}

	// Execute with idempotent transforms
	idempotentSQL := makeSQLIdempotent(string(content))
	_, err = pool.Exec(ctx, idempotentSQL)
	if err != nil {
		return fmt.Errorf("failed to apply post-schema.sql: %w", err)
	}

	log.Info().Msg("Post-schema policies applied")
	return nil
}

// hasDestructiveChanges checks if any changes are destructive
func hasDestructiveChanges(changes []Change) bool {
	for _, c := range changes {
		if c.Destructive {
			return true
		}
	}
	return false
}

// crossSchemaFKNames extracts FK constraint names from post-schema-fks.sql.
// These constraints are managed outside per-schema SQL files, so pgschema should not
// attempt to drop them during per-schema plan+apply.
var crossSchemaFKNames map[string]bool

// loadCrossSchemaFKNames parses post-schema-fks.sql and returns a set of FK constraint names.
func (s *DeclarativeService) loadCrossSchemaFKNames() map[string]bool {
	if crossSchemaFKNames != nil {
		return crossSchemaFKNames
	}

	names := make(map[string]bool)
	fksFile := filepath.Join(s.config.SchemaDir, "post-schema-fks.sql")
	content, err := os.ReadFile(fksFile)
	if err != nil {
		return names
	}

	// Match: conname = 'constraint_name'
	re := regexp.MustCompile(`conname\s*=\s*'([^']+)'`)
	matches := re.FindAllStringSubmatch(string(content), -1)
	for _, m := range matches {
		if len(m) > 1 {
			names[m[1]] = true
		}
	}

	crossSchemaFKNames = names
	return names
}

// filterManagedFKDrops removes DROP CONSTRAINT changes for FKs managed by post-schema-fks.sql.
// These cross-schema FKs are applied separately in Phase 2, so pgschema must not drop them.
func (s *DeclarativeService) filterManagedFKDrops(changes []Change) []Change {
	managedFKs := s.loadCrossSchemaFKNames()
	var filtered []Change
	for _, change := range changes {
		if change.Type == ChangeDrop && managedFKs[change.Name] {
			log.Debug().
				Str("constraint", change.Name).
				Msg("Skipping FK drop - managed by post-schema-fks.sql")
			continue
		}
		filtered = append(filtered, change)
	}
	return filtered
}

// makeSQLIdempotent transforms SQL to be idempotent by:
// - Converting CREATE POLICY to DROP POLICY IF EXISTS + CREATE POLICY
// - Converting ALTER TABLE ... ADD CONSTRAINT to DROP CONSTRAINT IF EXISTS + ADD CONSTRAINT
// - Other CREATE statements are left as-is since they typically use IF NOT EXISTS
//
// This function uses pgparser to parse SQL and identify statements that need transformation,
// then inserts DROP statements at the correct positions in the original SQL.
func makeSQLIdempotent(sql string) string {
	// Parse the SQL using pgparser
	stmts, err := parser.Parse(sql)
	if err != nil {
		// If parsing fails, return original SQL unchanged
		log.Warn().Err(err).Msg("Failed to parse SQL for idempotency transformation, using original")
		return sql
	}

	// If no statements, return original
	if stmts == nil || len(stmts.Items) == 0 {
		return sql
	}

	// Collect DROP statements with their target positions in the original SQL
	// We use position-based insertion to preserve statement order
	type dropInfo struct {
		pattern  string // Pattern to search for in original SQL
		dropSQL  string // DROP statement to insert
		foundPos int    // Position where pattern was found (-1 if not found)
	}
	var drops []dropInfo

	for _, item := range stmts.Items {
		switch stmt := item.(type) {
		case *nodes.CreatePolicyStmt:
			if stmt.Table != nil {
				tableName := formatRangeVar(stmt.Table)
				dropSQL := fmt.Sprintf("DROP POLICY IF EXISTS %s ON %s CASCADE;\n", quoteIdent(stmt.PolicyName), tableName)
				// Try quoted pattern first (for policy names with special characters)
				// Then try unquoted pattern
				patternQuoted := "CREATE POLICY \"" + stmt.PolicyName + "\""
				patternUnquoted := "CREATE POLICY " + stmt.PolicyName
				// Add both patterns - we'll use whichever is found
				drops = append(drops, dropInfo{pattern: patternQuoted, dropSQL: dropSQL, foundPos: -1})
				drops = append(drops, dropInfo{pattern: patternUnquoted, dropSQL: dropSQL, foundPos: -1})
			}

		case *nodes.AlterTableStmt:
			if stmt.Cmds != nil && stmt.Relation != nil {
				for _, cmd := range stmt.Cmds.Items {
					alterCmd, ok := cmd.(*nodes.AlterTableCmd)
					if !ok {
						continue
					}
					// AT_AddConstraint subtype is 17
					// AT_AttachPartition subtype is 60 - skip these
					if alterCmd.Subtype == 17 && alterCmd.Def != nil {
						if constraint, ok := alterCmd.Def.(*nodes.Constraint); ok && constraint.Conname != "" {
							tableName := formatRangeVar(stmt.Relation)
							dropSQL := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT IF EXISTS %s;\n", tableName, quoteIdent(constraint.Conname))
							// Search pattern: "ALTER TABLE <table>"
							// We'll find this position and insert the DROP before it
							pattern := "ALTER TABLE " + stmt.Relation.Relname
							drops = append(drops, dropInfo{pattern: pattern, dropSQL: dropSQL, foundPos: -1})
							break // Only one drop per ALTER TABLE statement
						}
					}
				}
			}
		}
	}

	// If no transformations needed, return original SQL
	if len(drops) == 0 {
		return sql
	}

	// Find positions of each pattern in the original SQL
	// We search case-insensitively and track the last found position for each pattern
	// to handle multiple occurrences of the same pattern (e.g., multiple ALTER TABLE on same table)
	upperSQL := strings.ToUpper(sql)
	lastFoundPos := make(map[string]int) // pattern -> last found position

	for i := range drops {
		upperPattern := strings.ToUpper(drops[i].pattern)
		// Search from after the last found position for this pattern
		searchStart := lastFoundPos[upperPattern]
		idx := strings.Index(upperSQL[searchStart:], upperPattern)
		if idx != -1 {
			drops[i].foundPos = searchStart + idx
			lastFoundPos[upperPattern] = drops[i].foundPos + len(upperPattern)
		}
	}

	// Sort drops by position (descending) so we can insert without affecting earlier positions
	// Using a simple bubble sort since the list is small
	for i := 0; i < len(drops)-1; i++ {
		for j := i + 1; j < len(drops); j++ {
			if drops[i].foundPos < drops[j].foundPos {
				drops[i], drops[j] = drops[j], drops[i]
			}
		}
	}

	// Insert DROP statements at their positions
	result := sql
	for _, drop := range drops {
		if drop.foundPos >= 0 {
			result = result[:drop.foundPos] + drop.dropSQL + result[drop.foundPos:]
		}
	}

	return result
}

// formatRangeVar formats a RangeVar as schema.table or just table if no schema
func formatRangeVar(rv *nodes.RangeVar) string {
	if rv.Schemaname != "" {
		return fmt.Sprintf("%s.%s", quoteIdent(rv.Schemaname), quoteIdent(rv.Relname))
	}
	return quoteIdent(rv.Relname)
}

// quoteIdent quotes an identifier if needed
func quoteIdent(name string) string {
	// If already quoted, return as-is
	if strings.HasPrefix(name, `"`) && strings.HasSuffix(name, `"`) {
		return name
	}
	// Quote the identifier to handle reserved words and special characters
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(name, `"`, `""`))
}

// GetSchemaStatus returns information about the current schema state
func (s *DeclarativeService) GetSchemaStatus(ctx context.Context) (*SchemaStatus, error) {
	status := &SchemaStatus{
		SchemaFile: s.config.SchemaDir,
	}

	// Calculate fingerprint
	fingerprint, err := s.CalculateFingerprint()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to calculate fingerprint")
	} else {
		status.SchemaFingerprint = fingerprint
	}

	// Get pending changes for all schemas
	totalChanges := 0
	for _, schema := range s.config.Schemas {
		plan, err := s.PlanForSchema(ctx, schema)
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to plan schema")
			continue
		}
		totalChanges += len(plan.Changes)
		if hasDestructiveChanges(plan.Changes) {
			status.HasDestructiveChanges = true
		}
	}
	status.PendingChanges = totalChanges

	// Get last applied state from database
	if s.pool != nil {
		var dbFingerprint, source string
		var appliedAt time.Time
		err := s.pool.QueryRow(ctx, `
			SELECT schema_fingerprint, applied_at, source
			FROM migrations.declarative_state
			ORDER BY id DESC
			LIMIT 1
		`).Scan(&dbFingerprint, &appliedAt, &source)
		if err == nil {
			status.LastAppliedFingerprint = dbFingerprint
			status.LastAppliedAt = appliedAt
			status.Source = source
		} else if err != pgx.ErrNoRows {
			log.Warn().Err(err).Msg("Failed to get declarative state from database")
		}
	}

	return status, nil
}

// SchemaStatus represents the current state of the declarative schema
type SchemaStatus struct {
	SchemaFile             string    `json:"schema_file"`
	SchemaFingerprint      string    `json:"schema_fingerprint"`
	LastAppliedFingerprint string    `json:"last_applied_fingerprint"`
	LastAppliedAt          time.Time `json:"last_applied_at"`
	Source                 string    `json:"source"`
	PendingChanges         int       `json:"pending_changes"`
	HasDestructiveChanges  bool      `json:"has_destructive_changes"`
}
