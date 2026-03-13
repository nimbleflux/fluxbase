package cmd

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cliconfig "github.com/nimbleflux/fluxbase/cli/config"
	"github.com/nimbleflux/fluxbase/cli/output"
)

var branchCmd = &cobra.Command{
	Use:     "branch",
	Aliases: []string{"branches", "br"},
	Short:   "Manage database branches",
	Long: `Manage database branches for isolated development and testing.

Database branches allow you to create isolated copies of your database
for development, testing, or preview environments. Each branch is a
separate PostgreSQL database that can be used independently.

Use branches to:
  - Test migrations before applying to production
  - Create isolated environments for PR previews
  - Safely experiment with schema changes`,
}

var (
	branchDataCloneMode string
	branchListType      string // Separate variable for list command
	branchCreateType    string // Separate variable for create command
	branchExpiresIn     string
	branchParent        string
	branchGitHubPR      int
	branchGitHubRepo    string
	branchSeedsDir      string
	branchForce         bool
)

var branchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all database branches",
	Long: `List all database branches.

Examples:
  fluxbase branch list
  fluxbase branch list --type preview
  fluxbase branch list --mine
  fluxbase branch list -o json`,
	PreRunE: requireAuth,
	RunE:    runBranchList,
}

var branchGetCmd = &cobra.Command{
	Use:   "get [name-or-id]",
	Short: "Get branch details",
	Long: `Get details of a specific branch.

Examples:
  fluxbase branch get my-feature
  fluxbase branch get pr-123
  fluxbase branch get 550e8400-e29b-41d4-a716-446655440000`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchGet,
}

var branchCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new database branch",
	Long: `Create a new database branch.

The branch name will be converted to a URL-safe slug.
By default, branches are created from the 'main' branch with schema only.

Examples:
  fluxbase branch create my-feature
  fluxbase branch create my-feature --clone-data full_clone
  fluxbase branch create staging --type persistent
  fluxbase branch create pr-123 --pr 123 --repo owner/repo`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchCreate,
}

var branchDeleteCmd = &cobra.Command{
	Use:     "delete [name-or-id]",
	Aliases: []string{"rm", "remove"},
	Short:   "Delete a database branch",
	Long: `Delete a database branch and its associated database.

This action is irreversible - all data in the branch will be lost.

Examples:
  fluxbase branch delete my-feature
  fluxbase branch delete pr-123 --force`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchDelete,
}

var branchResetCmd = &cobra.Command{
	Use:   "reset [name-or-id]",
	Short: "Reset a branch to its parent state",
	Long: `Reset a branch to its parent state, recreating the database.

This will drop the branch database and recreate it from the parent.
All changes in the branch will be lost.

Examples:
  fluxbase branch reset my-feature
  fluxbase branch reset pr-123 --force`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchReset,
}

var branchStatusCmd = &cobra.Command{
	Use:   "status [name-or-id]",
	Short: "Show branch status",
	Long: `Show the current status of a branch.

Examples:
  fluxbase branch status my-feature
  fluxbase branch status pr-123`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchStatus,
}

var branchActivityCmd = &cobra.Command{
	Use:   "activity [name-or-id]",
	Short: "Show branch activity log",
	Long: `Show the activity log for a branch.

Examples:
  fluxbase branch activity my-feature
  fluxbase branch activity pr-123 --limit 20`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchActivity,
}

var branchStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show connection pool statistics",
	Long: `Show connection pool statistics for all branches.

This is useful for debugging and monitoring database connections.

Examples:
  fluxbase branch stats`,
	PreRunE: requireAuth,
	RunE:    runBranchStats,
}

var branchUseCmd = &cobra.Command{
	Use:   "use [branch-name]",
	Short: "Set the default branch for CLI commands",
	Long: `Set the default branch for all subsequent CLI commands.

This saves the branch to your profile config and automatically sends
the X-Fluxbase-Branch header with every API request.

Use 'main' to reset to the main branch.

Examples:
  fluxbase branch use my-feature
  fluxbase branch use pr-123
  fluxbase branch use main`,
	Args:    cobra.ExactArgs(1),
	PreRunE: requireAuth,
	RunE:    runBranchUse,
}

var branchCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current default branch",
	Long: `Show the current default branch set for CLI commands.

Examples:
  fluxbase branch current`,
	RunE: runBranchCurrent,
}

func init() {
	// List command flags
	branchListCmd.Flags().StringVar(&branchListType, "type", "", "Filter by branch type (main, preview, persistent)")
	branchListCmd.Flags().BoolP("mine", "m", false, "Show only branches created by me")

	// Create command flags
	branchCreateCmd.Flags().StringVar(&branchDataCloneMode, "clone-data", "schema_only",
		"Data clone mode: schema_only, full_clone, seed_data")
	branchCreateCmd.Flags().StringVar(&branchCreateType, "type", "preview",
		"Branch type: preview, persistent")
	branchCreateCmd.Flags().StringVar(&branchExpiresIn, "expires-in", "",
		"Auto-delete after duration (e.g., 24h, 7d)")
	branchCreateCmd.Flags().StringVar(&branchParent, "from", "",
		"Parent branch to clone from (default: main)")
	branchCreateCmd.Flags().IntVar(&branchGitHubPR, "pr", 0,
		"GitHub PR number to associate with branch")
	branchCreateCmd.Flags().StringVar(&branchGitHubRepo, "repo", "",
		"GitHub repository (owner/repo)")
	branchCreateCmd.Flags().StringVar(&branchSeedsDir, "seeds-dir", "",
		"Custom directory containing seed SQL files (only with --clone-data seed_data)")

	// Delete command flags
	branchDeleteCmd.Flags().BoolVarP(&branchForce, "force", "f", false,
		"Skip confirmation prompt")

	// Reset command flags
	branchResetCmd.Flags().BoolVarP(&branchForce, "force", "f", false,
		"Skip confirmation prompt")

	// Activity command flags
	branchActivityCmd.Flags().IntP("limit", "n", 50, "Maximum number of entries to show")

	// Add subcommands
	branchCmd.AddCommand(branchListCmd)
	branchCmd.AddCommand(branchGetCmd)
	branchCmd.AddCommand(branchCreateCmd)
	branchCmd.AddCommand(branchDeleteCmd)
	branchCmd.AddCommand(branchResetCmd)
	branchCmd.AddCommand(branchStatusCmd)
	branchCmd.AddCommand(branchActivityCmd)
	branchCmd.AddCommand(branchStatsCmd)
	branchCmd.AddCommand(branchUseCmd)
	branchCmd.AddCommand(branchCurrentCmd)
}

// Branch represents a database branch
type Branch struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	DatabaseName   string     `json:"database_name"`
	Status         string     `json:"status"`
	Type           string     `json:"type"`
	DataCloneMode  string     `json:"data_clone_mode"`
	GitHubPRNumber *int       `json:"github_pr_number,omitempty"`
	GitHubRepo     *string    `json:"github_repo,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

// BranchActivity represents a branch activity log entry
type BranchActivity struct {
	ID           string    `json:"id"`
	BranchID     string    `json:"branch_id"`
	Action       string    `json:"action"`
	Status       string    `json:"status"`
	ErrorMessage *string   `json:"error_message,omitempty"`
	ExecutedAt   time.Time `json:"executed_at"`
	DurationMs   *int      `json:"duration_ms,omitempty"`
}

func runBranchList(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build query parameters
	query := url.Values{}
	query.Set("limit", "100")
	if branchListType != "" {
		query.Set("type", branchListType)
	}
	if mine, _ := cmd.Flags().GetBool("mine"); mine {
		query.Set("mine", "true")
	}

	var result struct {
		Branches []*Branch `json:"branches"`
		Total    int       `json:"total"`
	}

	if err := apiClient.DoGet(ctx, "/api/v1/admin/branches", query, &result); err != nil {
		return err
	}

	if len(result.Branches) == 0 {
		fmt.Println("No branches found")
		return nil
	}

	// Output based on format
	if formatter.Format != output.FormatTable {
		return formatter.Print(result.Branches)
	}

	// Table output
	data := output.TableData{
		Headers: []string{"NAME", "SLUG", "TYPE", "STATUS", "DATABASE", "CREATED"},
		Rows:    make([][]string, 0, len(result.Branches)),
	}

	for _, branch := range result.Branches {
		data.Rows = append(data.Rows, []string{
			branch.Name,
			branch.Slug,
			branch.Type,
			formatBranchStatus(branch.Status),
			branch.DatabaseName,
			formatTime(branch.CreatedAt),
		})
	}

	formatter.PrintTable(data)
	fmt.Printf("\nTotal: %d branches\n", result.Total)

	return nil
}

func runBranchGet(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nameOrID := args[0]

	var branch Branch
	if err := apiClient.DoGet(ctx, "/api/v1/admin/branches/"+url.PathEscape(nameOrID), nil, &branch); err != nil {
		return err
	}

	// Output based on format
	if formatter.Format != output.FormatTable {
		return formatter.Print(branch)
	}

	// Detailed output
	fmt.Printf("Name:           %s\n", branch.Name)
	fmt.Printf("Slug:           %s\n", branch.Slug)
	fmt.Printf("ID:             %s\n", branch.ID)
	fmt.Printf("Type:           %s\n", branch.Type)
	fmt.Printf("Status:         %s\n", formatBranchStatus(branch.Status))
	fmt.Printf("Database:       %s\n", branch.DatabaseName)
	fmt.Printf("Clone Mode:     %s\n", branch.DataCloneMode)
	if branch.GitHubPRNumber != nil {
		fmt.Printf("GitHub PR:      #%d\n", *branch.GitHubPRNumber)
	}
	if branch.GitHubRepo != nil {
		fmt.Printf("GitHub Repo:    %s\n", *branch.GitHubRepo)
	}
	fmt.Printf("Created:        %s\n", formatTime(branch.CreatedAt))
	fmt.Printf("Updated:        %s\n", formatTime(branch.UpdatedAt))
	if branch.ExpiresAt != nil {
		fmt.Printf("Expires:        %s\n", formatTime(*branch.ExpiresAt))
	}
	if branch.ErrorMessage != nil && *branch.ErrorMessage != "" {
		fmt.Printf("Error:          %s\n", *branch.ErrorMessage)
	}

	return nil
}

func runBranchCreate(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // Branch creation can take time
	defer cancel()

	name := args[0]

	// Build request body
	body := map[string]any{
		"name":            name,
		"data_clone_mode": branchDataCloneMode,
		"type":            branchCreateType,
	}

	if branchExpiresIn != "" {
		body["expires_in"] = branchExpiresIn
	}
	if branchGitHubPR > 0 {
		body["github_pr_number"] = branchGitHubPR
	}
	if branchGitHubRepo != "" {
		body["github_repo"] = branchGitHubRepo
	}
	if branchSeedsDir != "" {
		body["seeds_path"] = branchSeedsDir
	}

	var branch Branch
	if err := apiClient.DoPost(ctx, "/api/v1/admin/branches", body, &branch); err != nil {
		return err
	}

	// Output based on format
	if formatter.Format != output.FormatTable {
		return formatter.Print(branch)
	}

	fmt.Printf("Branch '%s' created successfully!\n\n", branch.Name)
	fmt.Printf("Slug:     %s\n", branch.Slug)
	fmt.Printf("Database: %s\n", branch.DatabaseName)
	fmt.Printf("Status:   %s\n", formatBranchStatus(branch.Status))
	fmt.Printf("\nTo use this branch:\n")
	fmt.Printf("  Header:  X-Fluxbase-Branch: %s\n", branch.Slug)
	fmt.Printf("  Query:   ?branch=%s\n", branch.Slug)
	fmt.Printf("  SDK:     { branch: '%s' }\n", branch.Slug)

	return nil
}

func runBranchDelete(cmd *cobra.Command, args []string) error {
	nameOrID := args[0]

	// Confirm deletion unless --force is used
	if !branchForce {
		fmt.Printf("Are you sure you want to delete branch '%s'?\n", nameOrID)
		fmt.Printf("This will permanently delete the database and all its data.\n")
		fmt.Printf("Type 'yes' to confirm: ")

		var confirm string
		_, _ = fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Deletion cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := apiClient.DoDelete(ctx, "/api/v1/admin/branches/"+url.PathEscape(nameOrID)); err != nil {
		return err
	}

	fmt.Printf("Branch '%s' deleted successfully\n", nameOrID)
	return nil
}

func runBranchReset(cmd *cobra.Command, args []string) error {
	nameOrID := args[0]

	// Confirm reset unless --force is used
	if !branchForce {
		fmt.Printf("Are you sure you want to reset branch '%s'?\n", nameOrID)
		fmt.Printf("This will recreate the database from its parent branch.\n")
		fmt.Printf("All changes in this branch will be lost.\n")
		fmt.Printf("Type 'yes' to confirm: ")

		var confirm string
		_, _ = fmt.Scanln(&confirm)

		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Reset cancelled")
			return nil
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var branch Branch
	if err := apiClient.DoPost(ctx, "/api/v1/admin/branches/"+url.PathEscape(nameOrID)+"/reset", nil, &branch); err != nil {
		return err
	}

	fmt.Printf("Branch '%s' reset successfully\n", branch.Name)
	fmt.Printf("Status: %s\n", formatBranchStatus(branch.Status))

	return nil
}

func runBranchStatus(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nameOrID := args[0]

	var branch Branch
	if err := apiClient.DoGet(ctx, "/api/v1/admin/branches/"+url.PathEscape(nameOrID), nil, &branch); err != nil {
		return err
	}

	fmt.Printf("Branch: %s (%s)\n", branch.Name, branch.Slug)
	fmt.Printf("Status: %s\n", formatBranchStatus(branch.Status))

	if branch.Status == "error" && branch.ErrorMessage != nil {
		fmt.Printf("Error:  %s\n", *branch.ErrorMessage)
	}

	return nil
}

func runBranchActivity(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nameOrID := args[0]
	limit, _ := cmd.Flags().GetInt("limit")

	query := url.Values{}
	query.Set("limit", fmt.Sprintf("%d", limit))

	var result struct {
		Activity []*BranchActivity `json:"activity"`
	}

	if err := apiClient.DoGet(ctx, "/api/v1/admin/branches/"+url.PathEscape(nameOrID)+"/activity", query, &result); err != nil {
		return err
	}

	if len(result.Activity) == 0 {
		fmt.Println("No activity found")
		return nil
	}

	// Output based on format
	if formatter.Format != output.FormatTable {
		return formatter.Print(result.Activity)
	}

	// Table output
	data := output.TableData{
		Headers: []string{"ACTION", "STATUS", "DURATION", "TIME", "ERROR"},
		Rows:    make([][]string, 0, len(result.Activity)),
	}

	for _, entry := range result.Activity {
		duration := "-"
		if entry.DurationMs != nil {
			duration = fmt.Sprintf("%dms", *entry.DurationMs)
		}

		errMsg := ""
		if entry.ErrorMessage != nil {
			errMsg = truncateString(*entry.ErrorMessage, 40)
		}

		data.Rows = append(data.Rows, []string{
			entry.Action,
			formatBranchStatus(entry.Status),
			duration,
			formatTime(entry.ExecutedAt),
			errMsg,
		})
	}

	formatter.PrintTable(data)

	return nil
}

func runBranchStats(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var result struct {
		Pools map[string]struct {
			TotalConns    int32 `json:"total_conns"`
			IdleConns     int32 `json:"idle_conns"`
			AcquiredConns int32 `json:"acquired_conns"`
			MaxConns      int32 `json:"max_conns"`
			AcquireCount  int64 `json:"acquire_count"`
		} `json:"pools"`
	}

	if err := apiClient.DoGet(ctx, "/api/v1/admin/branches/stats/pools", nil, &result); err != nil {
		return err
	}

	// Output based on format
	if formatter.Format != output.FormatTable {
		return formatter.Print(result.Pools)
	}

	// Table output
	data := output.TableData{
		Headers: []string{"BRANCH", "TOTAL", "IDLE", "ACQUIRED", "MAX", "ACQUIRES"},
		Rows:    make([][]string, 0, len(result.Pools)),
	}

	for name, stats := range result.Pools {
		data.Rows = append(data.Rows, []string{
			name,
			fmt.Sprintf("%d", stats.TotalConns),
			fmt.Sprintf("%d", stats.IdleConns),
			fmt.Sprintf("%d", stats.AcquiredConns),
			fmt.Sprintf("%d", stats.MaxConns),
			fmt.Sprintf("%d", stats.AcquireCount),
		})
	}

	formatter.PrintTable(data)

	return nil
}

// formatBranchStatus formats a status string
func formatBranchStatus(status string) string {
	switch status {
	case "ready":
		return "ready"
	case "creating", "migrating":
		return status + " ..."
	case "error", "failed":
		return status + " !"
	case "deleting", "deleted":
		return status
	default:
		return status
	}
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatTime formats a time.Time for display
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02 15:04:05")
}

func runBranchUse(cmd *cobra.Command, args []string) error {
	branchName := args[0]

	// If not "main", validate the branch exists
	if branchName != "main" {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var branch Branch
		if err := apiClient.DoGet(ctx, "/api/v1/admin/branches/"+url.PathEscape(branchName), nil, &branch); err != nil {
			return fmt.Errorf("branch '%s' not found: %w", branchName, err)
		}
	}

	// Load config
	configPath := GetConfigPath()
	cfg, err := cliconfig.LoadOrCreate(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current profile name
	pName := profileName
	if pName == "" {
		pName = cfg.CurrentProfile
	}
	if pName == "" {
		return fmt.Errorf("no profile configured - run 'fluxbase auth login' first")
	}

	// Get profile
	profile, err := cfg.GetProfile(pName)
	if err != nil {
		return fmt.Errorf("profile '%s' not found", pName)
	}

	// Update default branch
	if branchName == "main" {
		profile.DefaultBranch = "" // Clear to use main
	} else {
		profile.DefaultBranch = branchName
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if branchName == "main" {
		fmt.Println("Now using branch: main (default)")
	} else {
		fmt.Printf("Now using branch: %s\n", branchName)
	}
	fmt.Println("All CLI commands will now use this branch.")

	return nil
}

func runBranchCurrent(cmd *cobra.Command, args []string) error {
	// Load config
	configPath := GetConfigPath()
	cfg, err := cliconfig.LoadOrCreate(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current profile name
	pName := profileName
	if pName == "" {
		pName = cfg.CurrentProfile
	}
	if pName == "" {
		fmt.Println("Current branch: main (no profile configured)")
		return nil
	}

	// Get profile
	profile, err := cfg.GetProfile(pName)
	if err != nil {
		fmt.Println("Current branch: main (profile not found)")
		return nil
	}

	// Show current branch
	if profile.DefaultBranch == "" {
		fmt.Println("Current branch: main (default)")
	} else {
		fmt.Printf("Current branch: %s\n", profile.DefaultBranch)
	}

	return nil
}
