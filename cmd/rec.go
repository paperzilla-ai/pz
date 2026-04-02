package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	recCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	recCmd.Flags().Bool("markdown", false, "Print raw markdown")
}

var recCmd = &cobra.Command{
	Use:   "rec <project-paper-ref>",
	Short: "Show details for a recommendation from one of your projects",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := loadRequiredAuth()
		if err != nil {
			return err
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		markdownOut, _ := cmd.Flags().GetBool("markdown")
		if jsonOut && markdownOut {
			return fmt.Errorf("--json and --markdown cannot be used together")
		}

		projectPaperRef := args[0]
		if markdownOut {
			markdown, err := withAuth(&tokens, func(at string) (string, error) {
				return api.FetchProjectPaperMarkdown(at, projectPaperRef)
			})
			if err != nil {
				var pending *api.PaperMarkdownPendingError
				if errors.As(err, &pending) {
					fmt.Fprintln(cmd.OutOrStdout(), pending.FriendlyMessage())
					return nil
				}
				return fmt.Errorf("failed to fetch recommendation markdown: %w", err)
			}
			fmt.Fprint(cmd.OutOrStdout(), markdown)
			return nil
		}

		projectPaper, err := withAuth(&tokens, func(at string) (api.ProjectPaper, error) {
			return api.FetchProjectPaper(at, projectPaperRef)
		})
		if err != nil {
			return fmt.Errorf("failed to fetch recommendation: %w", err)
		}

		if jsonOut {
			data, err := json.MarshalIndent(projectPaper, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), string(data))
			return nil
		}

		writeProjectPaper(cmd.OutOrStdout(), projectPaper)
		return nil
	},
}
