package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is set by goreleaser via ldflags.
var Version = "dev"

const (
	cliGettingStartedURL = "https://docs.paperzilla.ai/guides/cli-getting-started"
	cliDocsURL           = "https://docs.paperzilla.ai/guides/cli"
)

var rootCmd = &cobra.Command{
	Use:     "pz",
	Short:   "Paperzilla CLI",
	Long:    "Paperzilla CLI\n\nGet started: " + cliGettingStartedURL + "\nCommand reference: " + cliDocsURL,
	Version: Version,
	Example: `  pz login
  pz update
  pz project list
  pz project <id>
  pz paper <paper-id>
  pz paper <paper-id> --project <project-id>
  pz rec <project-paper-id>
  pz feedback <project-paper-id> upvote
  pz feed <id>
  pz feed <id> --must-read --limit 5
  pz feed search --project-id <id> --query "latent retrieval"
  pz feed <id> --json
  pz feed <id> --atom`,
}

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(loginCmd, updateCmd, projectCmd, paperCmd, recCmd, feedbackCmd, feedCmd)
}

func Execute() {
	executedCmd, err := rootCmd.ExecuteC()
	if err == nil {
		maybePrintUpdateNotice(executedCmd, os.Stderr)
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
