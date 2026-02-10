// Package main provides a standalone tool to clean up test resources.
// This is useful after running tests with FLUXBASE_PARALLEL_TEST=true,
// which skips the normal teardown to allow parallel test execution.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// getEnvOrDefault returns the environment variable value or a default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Get database connection from environment
	// Support both DATABASE_URL and individual FLUXBASE_DATABASE_* variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Construct from individual environment variables (used in devcontainer)
		host := getEnvOrDefault("FLUXBASE_DATABASE_HOST", "localhost")
		port := getEnvOrDefault("FLUXBASE_DATABASE_PORT", "5432")
		user := getEnvOrDefault("FLUXBASE_DATABASE_USER", "postgres")
		password := getEnvOrDefault("FLUXBASE_DATABASE_PASSWORD", "postgres")
		database := getEnvOrDefault("FLUXBASE_DATABASE_DATABASE", "fluxbase_test")
		dbURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, database)
	}

	ctx := context.Background()

	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()

	log.Info().Msg("Cleaning up test resources...")

	// 1. Drop test tables matching patterns
	rows, _ := pool.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		AND (
			tablename LIKE 'test_table_%'
			OR tablename LIKE 'test_single_%'
			OR tablename LIKE 'test_already_%'
			OR tablename LIKE 'test_rollback_%'
			OR tablename LIKE 'test_nodown_%'
			OR tablename LIKE 'test_stop_%'
			OR tablename LIKE 'test_history_%'
			OR tablename LIKE 'test_retry_%'
			OR tablename LIKE 'test_delete_%'
			OR tablename LIKE 'ns_test_%'
		)
	`)
	var tableNames []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			log.Error().Err(err).Msg("Failed to scan table name")
			continue
		}
		tableNames = append(tableNames, tableName)
	}
	rows.Close()

	// Drop each test table
	for _, tableName := range tableNames {
		_, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS public.%s CASCADE", fmt.Sprintf("\"%s\"", tableName)))
		if err != nil {
			log.Error().Err(err).Str("table", tableName).Msg("Failed to drop test table")
		} else {
			log.Info().Str("table", tableName).Msg("Dropped test table")
		}
	}

	// 2. Delete hardcoded test tables
	hardcodedTables := []string{
		"public.products",
		"public.tasks",
		"public.locations",
		"public.regions",
		"public.role_check",
		"public.sensitive_data",
	}
	for _, table := range hardcodedTables {
		_, err := pool.Exec(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			log.Error().Err(err).Str("table", table).Msg("Failed to drop table")
		}
	}

	// 3. Delete test secrets
	result, err := pool.Exec(ctx, "DELETE FROM functions.secrets WHERE name LIKE 'test_%'")
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete test secrets")
	} else if result.RowsAffected() > 0 {
		log.Info().Int64("count", result.RowsAffected()).Msg("Deleted test secrets")
	}

	// 4. Delete test API keys (legacy)
	result, err = pool.Exec(ctx, "DELETE FROM auth.api_keys WHERE name LIKE 'test_%'")
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete test API keys")
	} else if result.RowsAffected() > 0 {
		log.Info().Int64("count", result.RowsAffected()).Msg("Deleted test API keys")
	}

	// 5. Delete test client keys
	result, err = pool.Exec(ctx, "DELETE FROM auth.client_keys WHERE name LIKE 'test_%'")
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete test client keys")
	} else if result.RowsAffected() > 0 {
		log.Info().Int64("count", result.RowsAffected()).Msg("Deleted test client keys")
	}

	// 6. Delete test storage buckets
	result, err = pool.Exec(ctx, "DELETE FROM storage.buckets WHERE id LIKE 'test_%' OR name LIKE 'test_%'")
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete test storage buckets")
	} else if result.RowsAffected() > 0 {
		log.Info().Int64("count", result.RowsAffected()).Msg("Deleted test storage buckets")
	}

	log.Info().Msg("Test resource cleanup complete")
}
