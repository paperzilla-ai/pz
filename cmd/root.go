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
  pz project list
  pz project <id>
  pz paper <id>
  pz feed <id>
  pz feed <id> --must-read --limit 5
  pz feed <id> --json
  pz feed <id> --atom`,
}

func init() {
	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(loginCmd, projectCmd, paperCmd, feedCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
