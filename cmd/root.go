package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pz",
	Short: "Paperzilla CLI",
	Example: `  pz login
  pz project list
  pz project <id>
  pz feed <id>
  pz feed <id> --must-read --limit 5
  pz feed <id> --json`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
