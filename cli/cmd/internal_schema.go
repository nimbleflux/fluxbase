package cmd

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
)

var internalSchemaCmd = &cobra.Command{
	Use:   "internal-schema",
	Short: "Manage Fluxbase internal database schema",
	Long: `Manage Fluxbase's internal schema (auth, storage, jobs, etc.) declaratively.

This command manages Fluxbase's internal schemas in a declarative manner using pgschema.
User tables in the public schema are managed separately via the 'migrations' command.

The internal schema includes:
  - auth: Authentication tables (users, sessions, etc.)
  - storage: File storage buckets and objects
  - jobs: Background job queue
  - functions: Edge functions
  - realtime: Real-time subscriptions
  - ai: AI chatbots and knowledge bases
  - rpc: Stored procedures
  - system: System infrastructure
  - migrations: Migration tracking
  - platform: Multi-tenancy control plane
  - app: Application settings
  - api: API infrastructure
  - branching: Database branching
  - logging: Centralized logging
  - mcp: Model Context Protocol`,
}

var (
	internalSchemaFile              string
	internalSchemaAutoApprove       bool
	internalSchemaAllowDestructive  bool
	internalSchemaFailOnDrift       bool
	internalSchemaKeepOldMigrations bool
)

var internalSchemaDumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Export current internal schema to SQL",
	Long: `Export the current database schema to a SQL file for declarative management.

This command dumps all Fluxbase internal schemas to a SQL file that can be
managed declaratively with pgschema.

Examples:
  fluxbase internal-schema dump
  fluxbase internal-schema dump --dir ./custom/schemas`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaDump,
}

var internalSchemaPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Show pending schema changes",
	Long: `Compare the schema file against the database and show what would change.

This command generates a plan of changes that would be applied to bring
the database in line with the declared schema file. It does not make any
changes to the database.

Examples:
  fluxbase internal-schema plan
  fluxbase internal-schema plan --schema auth`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaPlan,
}

var internalSchemaApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply schema changes",
	Long: `Apply pending schema changes from the schema file to the database.

This command applies the changes shown by 'plan' to bring the database
in line with the declared schema file.

Examples:
  fluxbase internal-schema apply
  fluxbase internal-schema apply --auto-approve
  fluxbase internal-schema apply --allow-destructive`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaApply,
}

var internalSchemaValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check for schema drift",
	Long: `Validate that the database matches the schema file.

This command checks for any differences between the declared schema and
the actual database schema. It's useful for CI/CD pipelines to detect
drift.

Examples:
  fluxbase internal-schema validate
  fluxbase internal-schema validate --fail-on-drift`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaValidate,
}

var internalSchemaStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show internal schema status",
	Long: `Show the current status of the internal schema management.

Displays whether the database has been bootstrapped, the current migration
state, and any pending changes.

Examples:
  fluxbase internal-schema status`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaStatus,
}

var internalSchemaMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from imperative to declarative schema",
	Long: `One-time migration from imperative migrations to declarative schema management.

This command:
1. Verifies all existing migrations are applied
2. Exports current schema to schemas/ directory
3. Initializes declarative state tracking
4. Optionally removes old migration files

Run this once on existing deployments.

Examples:
  fluxbase internal-schema migrate
  fluxbase internal-schema migrate --keep-old-migrations`,
	PreRunE: requireAuth,
	RunE:    runInternalSchemaMigrate,
}

func init() {
	rootCmd.AddCommand(internalSchemaCmd)
	internalSchemaCmd.AddCommand(internalSchemaDumpCmd)
	internalSchemaCmd.AddCommand(internalSchemaPlanCmd)
	internalSchemaCmd.AddCommand(internalSchemaApplyCmd)
	internalSchemaCmd.AddCommand(internalSchemaValidateCmd)
	internalSchemaCmd.AddCommand(internalSchemaStatusCmd)
	internalSchemaCmd.AddCommand(internalSchemaMigrateCmd)

	// Global flags for internal-schema commands
	internalSchemaCmd.PersistentFlags().StringVar(&internalSchemaFile, "file", "", "Path to schema file (for backward compatibility)")

	// Apply command flags
	internalSchemaApplyCmd.Flags().BoolVar(&internalSchemaAutoApprove, "auto-approve", false, "Auto-approve changes without prompting")
	internalSchemaApplyCmd.Flags().BoolVar(&internalSchemaAllowDestructive, "allow-destructive", false, "Allow destructive changes (DROP statements)")

	// Validate command flags
	internalSchemaValidateCmd.Flags().BoolVar(&internalSchemaFailOnDrift, "fail-on-drift", false, "Exit with error code if drift detected")

	// Migrate command flags
	internalSchemaMigrateCmd.Flags().BoolVar(&internalSchemaKeepOldMigrations, "keep-old-migrations", false, "Keep old migration files after transition")
}

func runInternalSchemaDump(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	body := map[string]interface{}{}
	if internalSchemaFile != "" {
		body["file"] = internalSchemaFile
	}

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/internal-schema/dump", body, &result); err != nil {
		return err
	}

	fmt.Printf("Schema dumped successfully.\n")
	if sql, ok := result["sql"].(string); ok && len(sql) > 0 {
		fmt.Printf("Output: %d bytes\n", len(sql))
	}
	return nil
}

func runInternalSchemaPlan(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	body := map[string]interface{}{}
	if internalSchemaFile != "" {
		body["file"] = internalSchemaFile
	}

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/internal-schema/plan", body, &result); err != nil {
		return err
	}

	plan, ok := result["plan"].(map[string]interface{})
	if !ok {
		fmt.Println("No plan returned.")
		return nil
	}

	changes, _ := plan["changes"].([]interface{})
	if len(changes) == 0 {
		fmt.Println("No changes detected - database matches schema file.")
		return nil
	}

	fmt.Printf("Found %d pending changes:\n\n", len(changes))
	for i, change := range changes {
		c, ok := change.(map[string]interface{})
		if !ok {
			continue
		}
		destructive := ""
		if d, _ := c["destructive"].(bool); d {
			destructive = " [DESTRUCTIVE]"
		}
		fmt.Printf("  %d. %s %s.%s (%s)%s\n",
			i+1,
			c["type"],
			c["schema"],
			c["name"],
			c["object_type"],
			destructive)
	}

	if summary, ok := plan["summary"].(map[string]interface{}); ok {
		fmt.Println()
		fmt.Println("Summary:")
		fmt.Printf("  Total changes: %v\n", summary["total_changes"])
		fmt.Printf("  Destructive:   %v\n", summary["destructive_count"])
	}

	return nil
}

func runInternalSchemaApply(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"auto_approve":      internalSchemaAutoApprove,
		"allow_destructive": internalSchemaAllowDestructive,
	}
	if internalSchemaFile != "" {
		body["file"] = internalSchemaFile
	}

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/internal-schema/apply", body, &result); err != nil {
		return err
	}

	applied, _ := result["applied"].([]interface{})
	duration, _ := result["duration"].(string)

	fmt.Printf("Applied %d changes in %s\n", len(applied), duration)
	return nil
}

func runInternalSchemaValidate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	query := url.Values{}
	if internalSchemaFile != "" {
		query.Set("file", internalSchemaFile)
	}

	var result map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/internal-schema/validate", query, &result); err != nil {
		return err
	}

	valid, _ := result["valid"].(bool)
	if valid {
		fmt.Println("Schema is valid - no drift detected.")
		return nil
	}

	drifts, _ := result["drifts"].([]interface{})
	fmt.Printf("Schema drift detected (%d changes):\n", len(drifts))
	for _, drift := range drifts {
		d, ok := drift.(map[string]interface{})
		if !ok {
			continue
		}
		destructive := ""
		if desc, _ := d["destructive"].(bool); desc {
			destructive = " [DESTRUCTIVE]"
		}
		fmt.Printf("  - %s %s.%s (%s)%s\n",
			d["type"],
			d["schema"],
			d["name"],
			d["object_type"],
			destructive)
	}

	if internalSchemaFailOnDrift {
		return fmt.Errorf("schema drift detected")
	}

	return nil
}

func runInternalSchemaStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result map[string]interface{}
	if err := apiClient.DoGet(ctx, "/api/v1/admin/internal-schema/status", nil, &result); err != nil {
		return err
	}

	fmt.Println("Internal Schema Status:")
	fmt.Println()
	fmt.Printf("  Bootstrapped:          %v\n", result["bootstrapped"])
	fmt.Printf("  Imperative Migrations: %v\n", result["has_imperative_migrations"])
	if v, ok := result["last_migration_version"]; ok && v != nil {
		fmt.Printf("    Last Version:        %v\n", v)
	}
	fmt.Printf("  Declarative State:     %v\n", result["has_declarative_state"])
	if fp, ok := result["schema_fingerprint"].(string); ok && len(fp) > 16 {
		fmt.Printf("  Schema Fingerprint:    %s...\n", fp[:16])
	}
	fmt.Println()

	pending, _ := result["pending_changes"].(float64)
	if pending == 0 {
		fmt.Println("No pending changes - database matches schema file.")
	} else {
		fmt.Printf("Found %.0f pending changes - run 'plan' for details.\n", pending)
	}

	return nil
}

func runInternalSchemaMigrate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	body := map[string]interface{}{
		"keep_old_migrations": internalSchemaKeepOldMigrations,
	}
	if internalSchemaFile != "" {
		body["file"] = internalSchemaFile
	}

	fmt.Println("Migrating from imperative to declarative schema management...")

	var result map[string]interface{}
	if err := apiClient.DoPost(ctx, "/api/v1/admin/internal-schema/migrate", body, &result); err != nil {
		return err
	}

	fmt.Println("Migration to declarative schema complete!")
	if schemaFile, ok := result["schema_file"].(string); ok {
		fmt.Printf("Schema file: %s\n", schemaFile)
	}
	if fingerprint, ok := result["fingerprint"].(string); ok {
		fmt.Printf("Fingerprint: %s\n", fingerprint)
	}
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Review the generated schema file")
	fmt.Println("  2. Commit it to version control")
	fmt.Println("  3. Use 'fluxbase internal-schema plan' and 'apply' for future changes")

	return nil
}
