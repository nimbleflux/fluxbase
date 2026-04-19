package tenantdb

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// FDWConfig holds the connection details for the main database,
// used to set up postgres_fdw foreign tables in tenant databases.
type FDWConfig struct {
	Host     string // Main DB host
	Port     string // Main DB port
	DBName   string // Main DB name
	User     string // Main DB user (needs SELECT on auth.*)
	Password string // Main DB password
}

// FDWRoleCredentials holds the credentials for the per-tenant FDW role.
type FDWRoleCredentials struct {
	RoleName string
	Password string
}

// fdwSchemas lists schemas whose tables should be imported via FDW.
// These are shared infrastructure tables living in the main database,
// accessed by tenant databases through foreign data wrappers.
var fdwSchemas = []string{
	"platform", "auth", "storage", "jobs", "functions", "realtime",
	"ai", "rpc", "branching", "logging", "mcp",
}

// fdwExcludeTables lists tables that should NOT be imported via FDW
// because they hold per-database local state.
var fdwExcludeTables = map[string][]string{
	"platform": {"schema_migrations"},
	"logging":  {"execution_logs_migration_status"},
}

// ParseFDWConfig extracts FDW connection details from a database URL.
func ParseFDWConfig(dbURL string) (FDWConfig, error) {
	u, err := url.Parse(dbURL)
	if err != nil {
		return FDWConfig{}, fmt.Errorf("failed to parse database URL: %w", err)
	}

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = "5432"
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	user := u.User.Username()
	password, _ := u.User.Password()

	if host == "" || dbName == "" || user == "" {
		return FDWConfig{}, fmt.Errorf("incomplete database URL: need host, dbname, and user")
	}

	return FDWConfig{
		Host:     host,
		Port:     port,
		DBName:   dbName,
		User:     user,
		Password: password,
	}, nil
}

const fdwServerName = "main_server"

// CreateFDWRole creates a dedicated PostgreSQL role on the main database
// for a specific tenant. The role has NOBYPASSRLS so RLS policies are
// enforced, and has a default app.current_tenant_id set so queries
// through FDW are automatically filtered by tenant.
func CreateFDWRole(ctx context.Context, adminPool *pgxpool.Pool, tenantID string) (FDWRoleCredentials, error) {
	// Use first 8 chars of UUID for readable role name
	suffix := tenantID
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	roleName := fmt.Sprintf("fdw_tenant_%s", suffix)

	// Generate a random password
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return FDWRoleCredentials{}, fmt.Errorf("failed to generate FDW role password: %w", err)
	}
	password := base64.URLEncoding.EncodeToString(keyBytes)

	// Create role with NOBYPASSRLS so RLS policies are enforced
	_, err := adminPool.Exec(ctx, fmt.Sprintf(
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = '%s') THEN
				CREATE ROLE %s NOBYPASSRLS LOGIN PASSWORD '%s';
			END IF;
		END $$`,
		escapeSQLString(roleName),
		quoteIdent(roleName),
		escapeSQLString(password),
	))
	if err != nil {
		return FDWRoleCredentials{}, fmt.Errorf("failed to create FDW role: %w", err)
	}

	// Set default tenant context so RLS policies filter automatically
	_, err = adminPool.Exec(ctx, fmt.Sprintf(
		`ALTER ROLE %s SET app.current_tenant_id = '%s'`,
		quoteIdent(roleName),
		escapeSQLString(tenantID),
	))
	if err != nil {
		return FDWRoleCredentials{}, fmt.Errorf("failed to set FDW role tenant context: %w", err)
	}

	// Grant usage on all FDW schemas
	for _, schema := range fdwSchemas {
		_, err := adminPool.Exec(ctx, fmt.Sprintf(
			`GRANT USAGE ON SCHEMA %s TO %s`,
			quoteIdent(schema), quoteIdent(roleName),
		))
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to grant schema usage to FDW role")
		}
	}

	// Grant ALL on all tables in FDW schemas (read + write through FDW)
	for _, schema := range fdwSchemas {
		_, err := adminPool.Exec(ctx, fmt.Sprintf(
			`GRANT ALL ON ALL TABLES IN SCHEMA %s TO %s`,
			quoteIdent(schema), quoteIdent(roleName),
		))
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to grant table permissions to FDW role")
		}
	}

	// Grant usage on sequences (needed for INSERT with serial columns)
	for _, schema := range fdwSchemas {
		_, _ = adminPool.Exec(ctx, fmt.Sprintf(
			`GRANT USAGE ON ALL SEQUENCES IN SCHEMA %s TO %s`,
			quoteIdent(schema), quoteIdent(roleName),
		))
	}

	log.Info().Str("role", roleName).Str("tenant_id", tenantID).Msg("Created FDW role for tenant")

	return FDWRoleCredentials{
		RoleName: roleName,
		Password: password,
	}, nil
}

// DropFDWRole removes the per-tenant FDW role from the main database.
func DropFDWRole(ctx context.Context, adminPool *pgxpool.Pool, tenantID string) {
	suffix := tenantID
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	roleName := fmt.Sprintf("fdw_tenant_%s", suffix)

	_, err := adminPool.Exec(ctx, fmt.Sprintf(`DROP ROLE IF EXISTS %s`, quoteIdent(roleName)))
	if err != nil {
		log.Warn().Err(err).Str("role", roleName).Msg("Failed to drop FDW role")
	}
}

// SetupFDW configures postgres_fdw in a tenant database so it can access
// tables from the main database. It creates a foreign server, user mapping
// (using the per-tenant FDW role for RLS), and imports foreign tables from
// all non-public schemas.
//
// Local tables that would conflict with FDW imports are dropped first.
// Functions, types, and other non-table objects remain local.
func SetupFDW(ctx context.Context, tenantPool *pgxpool.Pool, cfg FDWConfig, tables []string) error {
	if tenantPool == nil {
		return fmt.Errorf("tenant pool is nil")
	}
	if cfg.Host == "" || cfg.DBName == "" {
		return fmt.Errorf("FDW config incomplete: host and dbname required")
	}

	// Legacy path: if specific tables are requested, use old auth-only import
	if len(tables) > 0 {
		return setupFDWLegacy(ctx, tenantPool, cfg, tables)
	}

	return setupFDWAllSchemas(ctx, tenantPool, cfg)
}

// setupFDWAllSchemas imports all tables from all FDW schemas.
func setupFDWAllSchemas(ctx context.Context, tenantPool *pgxpool.Pool, cfg FDWConfig) error {
	// 1. Ensure postgres_fdw extension exists
	_, err := tenantPool.Exec(ctx, `CREATE EXTENSION IF NOT EXISTS postgres_fdw`)
	if err != nil {
		return fmt.Errorf("failed to create postgres_fdw extension: %w", err)
	}

	// 2. Create foreign server pointing at main database
	_, err = tenantPool.Exec(ctx, fmt.Sprintf(
		`CREATE SERVER IF NOT EXISTS %s FOREIGN DATA WRAPPER postgres_fdw
		  OPTIONS (host '%s', port '%s', dbname '%s')`,
		quoteIdent(fdwServerName), cfg.Host, cfg.Port, cfg.DBName,
	))
	if err != nil {
		return fmt.Errorf("failed to create foreign server: %w", err)
	}

	// 3. Create user mapping using admin credentials
	// This will be replaced by the per-tenant role mapping in CreateTenantDatabase
	userMappingSQL := fmt.Sprintf(
		`CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_USER SERVER %s
		  OPTIONS (user '%s'`,
		quoteIdent(fdwServerName), cfg.User,
	)
	if cfg.Password != "" {
		userMappingSQL += fmt.Sprintf(`, password '%s'`, escapeSQLString(cfg.Password))
	}
	userMappingSQL += ")"
	_, err = tenantPool.Exec(ctx, userMappingSQL)
	if err != nil {
		return fmt.Errorf("failed to create user mapping: %w", err)
	}

	// 4. Import tables from each FDW schema
	for _, schema := range fdwSchemas {
		if err := importSchemaFDW(ctx, tenantPool, schema); err != nil {
			log.Warn().Err(err).Str("schema", schema).Msg("Failed to import schema via FDW")
			// Continue with other schemas — not all schemas may have tables
		}
	}

	log.Info().
		Strs("schemas", fdwSchemas).
		Str("main_db", cfg.DBName).
		Msg("Set up FDW for tenant database (all schemas)")

	return nil
}

// importSchemaFDW drops local tables and imports foreign tables for a single schema.
func importSchemaFDW(ctx context.Context, tenantPool *pgxpool.Pool, schema string) error {
	excluded := fdwExcludeTables[schema]

	// Get list of local tables in this schema to drop
	rows, err := tenantPool.Query(ctx, `
		SELECT tablename FROM pg_tables WHERE schemaname = $1
	`, schema)
	if err != nil {
		return fmt.Errorf("failed to list tables in schema %s: %w", schema, err)
	}

	var localTables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		// Skip excluded tables
		skip := false
		for _, ex := range excluded {
			if name == ex {
				skip = true
				break
			}
		}
		if !skip {
			localTables = append(localTables, name)
		}
	}
	rows.Close()

	// Drop local tables that will be replaced by foreign tables
	for _, table := range localTables {
		_, err := tenantPool.Exec(ctx, fmt.Sprintf(
			`DROP TABLE IF EXISTS %s.%s CASCADE`,
			quoteIdent(schema), quoteIdent(table),
		))
		if err != nil {
			log.Warn().Err(err).Str("schema", schema).Str("table", table).
				Msg("Failed to drop local table before FDW import")
		}
		// Also drop existing foreign tables for idempotency
		_, _ = tenantPool.Exec(ctx, fmt.Sprintf(
			`DROP FOREIGN TABLE IF EXISTS %s.%s CASCADE`,
			quoteIdent(schema), quoteIdent(table),
		))
	}

	// Build IMPORT FOREIGN SCHEMA statement
	importSQL := fmt.Sprintf(
		`IMPORT FOREIGN SCHEMA %s FROM SERVER %s INTO %s`,
		quoteIdent(schema), quoteIdent(fdwServerName), quoteIdent(schema),
	)

	// Add EXCEPT clause for excluded tables
	if len(excluded) > 0 {
		excludedList := make([]string, len(excluded))
		for i, t := range excluded {
			excludedList[i] = quoteIdent(t)
		}
		importSQL = fmt.Sprintf(
			`IMPORT FOREIGN SCHEMA %s EXCEPT (%s) FROM SERVER %s INTO %s`,
			quoteIdent(schema), strings.Join(excludedList, ", "),
			quoteIdent(fdwServerName), quoteIdent(schema),
		)
	}

	_, err = tenantPool.Exec(ctx, importSQL)
	if err != nil {
		return fmt.Errorf("failed to import foreign schema %s: %w", schema, err)
	}

	if len(localTables) > 0 {
		log.Debug().Str("schema", schema).Int("tables", len(localTables)).
			Msg("Imported schema tables via FDW")
	}

	return nil
}

// CreateFDWUserMapping creates a user mapping for the app user using
// the per-tenant FDW role credentials. This overrides the default
// admin user mapping with the tenant-specific role for RLS enforcement.
func CreateFDWUserMapping(ctx context.Context, tenantPool *pgxpool.Pool, appUser string, fdwRole FDWRoleCredentials) error {
	// Drop existing user mapping for the app user
	_, _ = tenantPool.Exec(ctx, fmt.Sprintf(
		`DROP USER MAPPING IF EXISTS FOR %s SERVER %s`,
		quoteIdent(appUser), quoteIdent(fdwServerName),
	))

	// Create user mapping with tenant-specific FDW role
	_, err := tenantPool.Exec(ctx, fmt.Sprintf(
		`CREATE USER MAPPING FOR %s SERVER %s OPTIONS (user '%s', password '%s')`,
		quoteIdent(appUser), quoteIdent(fdwServerName),
		escapeSQLString(fdwRole.RoleName), escapeSQLString(fdwRole.Password),
	))
	if err != nil {
		return fmt.Errorf("failed to create FDW user mapping for app user: %w", err)
	}

	log.Debug().Str("app_user", appUser).Str("fdw_role", fdwRole.RoleName).
		Msg("Created FDW user mapping with tenant-specific role")

	return nil
}

// setupFDWLegacy implements the original auth-only FDW import for backward compatibility.
func setupFDWLegacy(ctx context.Context, tenantPool *pgxpool.Pool, cfg FDWConfig, tablesToImport []string) error {
	// 1. Create foreign server
	_, err := tenantPool.Exec(ctx, fmt.Sprintf(
		`CREATE SERVER IF NOT EXISTS %s FOREIGN DATA WRAPPER postgres_fdw
		  OPTIONS (host '%s', port '%s', dbname '%s')`,
		quoteIdent(fdwServerName), cfg.Host, cfg.Port, cfg.DBName,
	))
	if err != nil {
		return fmt.Errorf("failed to create foreign server: %w", err)
	}

	// 2. Create user mapping
	userMappingSQL := fmt.Sprintf(
		`CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_USER SERVER %s
		  OPTIONS (user '%s'`,
		quoteIdent(fdwServerName), cfg.User,
	)
	if cfg.Password != "" {
		userMappingSQL += fmt.Sprintf(`, password '%s'`, escapeSQLString(cfg.Password))
	}
	userMappingSQL += ")"
	_, err = tenantPool.Exec(ctx, userMappingSQL)
	if err != nil {
		return fmt.Errorf("failed to create user mapping: %w", err)
	}

	// 3. Drop local tables that will be replaced by foreign tables
	for _, table := range tablesToImport {
		_, err := tenantPool.Exec(ctx, fmt.Sprintf(
			`DROP TABLE IF EXISTS auth.%s CASCADE`,
			quoteIdent(table),
		))
		if err != nil {
			log.Warn().Err(err).Str("table", table).Msg("Failed to drop local table before FDW import")
		}
		// Also drop existing foreign tables so IMPORT FOREIGN SCHEMA is idempotent
		_, err = tenantPool.Exec(ctx, fmt.Sprintf(
			`DROP FOREIGN TABLE IF EXISTS auth.%s CASCADE`,
			quoteIdent(table),
		))
		if err != nil {
			log.Warn().Err(err).Str("table", table).Msg("Failed to drop foreign table before FDW import")
		}
	}

	// 4. Import foreign tables into auth schema
	tableList := make([]string, len(tablesToImport))
	for i, t := range tablesToImport {
		tableList[i] = quoteIdent(t)
	}

	_, err = tenantPool.Exec(ctx, fmt.Sprintf(
		`IMPORT FOREIGN SCHEMA auth LIMIT TO (%s) FROM SERVER %s INTO auth`,
		strings.Join(tableList, ", "),
		quoteIdent(fdwServerName),
	))
	if err != nil {
		return fmt.Errorf("failed to import foreign schema: %w", err)
	}

	log.Info().
		Str("tables", strings.Join(tablesToImport, ", ")).
		Str("main_db", cfg.DBName).
		Msg("Set up FDW for tenant database (legacy auth-only)")

	return nil
}

// TeardownFDW removes all FDW artifacts from a tenant database.
func TeardownFDW(ctx context.Context, tenantPool *pgxpool.Pool) error {
	if tenantPool == nil {
		return fmt.Errorf("tenant pool is nil")
	}

	// Drop user mapping
	_, _ = tenantPool.Exec(ctx, fmt.Sprintf(
		`DROP USER MAPPING IF EXISTS FOR CURRENT_USER SERVER %s`,
		quoteIdent(fdwServerName),
	))

	// Drop foreign server (cascades to foreign tables)
	_, err := tenantPool.Exec(ctx, fmt.Sprintf(
		`DROP SERVER IF EXISTS %s CASCADE`,
		quoteIdent(fdwServerName),
	))
	if err != nil {
		return fmt.Errorf("failed to drop foreign server: %w", err)
	}

	log.Info().Msg("Tore down FDW from tenant database")
	return nil
}

// quoteIdent quotes a PostgreSQL identifier.
func quoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// escapeSQLString escapes a string for use in SQL single-quoted literals.
func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
