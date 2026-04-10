package cmd

import (
	"github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
	Use:     "admin",
	Aliases: []string{"adm"},
	Short:   "Manage admin users and dashboard access",
	Long: `Manage admin users, invitations, and sessions for the Fluxbase platform.

Admin users have access to the Fluxbase admin dashboard for managing
your database, users, functions, and other platform features.

Use subcommands to manage:
  - users: Admin user CRUD operations
  - invitations: Pending admin invitations
  - sessions: Active admin sessions`,
}

func init() {
	adminCmd.AddCommand(adminUsersCmd)
	adminCmd.AddCommand(adminInvitationsCmd)
	adminCmd.AddCommand(adminSessionsCmd)
}
