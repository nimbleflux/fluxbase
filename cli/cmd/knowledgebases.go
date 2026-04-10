package cmd

import (
	"github.com/spf13/cobra"
)

var kbCmd = &cobra.Command{
	Use:     "kb",
	Aliases: []string{"knowledge-bases", "knowledge-base"},
	Short:   "Manage knowledge bases",
	Long:    `Create and manage knowledge bases for AI chatbots.`,
}

func init() {
	// Knowledge base commands are disabled because the referenced admin API
	// routes (/api/v1/admin/ai/knowledge-bases/*) never existed on the server.
	// The actual routes are at /api/v1/ai/knowledge-bases/* (user-facing).
	// These commands should be re-added when CLI support for the user-facing
	// KB routes is implemented.
}
