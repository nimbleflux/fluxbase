package tenantdb

import (
	"context"
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

// SetupFDW configures postgres_fdw in a tenant database so it can access
// tables from the main database (e.g., auth.users for cross-database joins).
//
// It creates a foreign server, user mapping, and imports specific foreign
// tables into the tenant's auth schema. Local tables that would conflict
// are dropped first. The auth schema itself and its grants are preserved.
func SetupFDW(ctx context.Context, tenantPool *pgxpool.Pool, cfg FDWConfig, tables []string) error {
	if tenantPool == nil {
		return fmt.Errorf("tenant pool is nil")
	}
	if cfg.Host == "" || cfg.DBName == "" {
		return fmt.Errorf("FDW config incomplete: host and dbname required")
	}

	tablesToImport := tables
	if len(tablesToImport) == 0 {
		tablesToImport = []string{"users", "identities"}
	}

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
		Msg("Set up FDW for tenant database")

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
