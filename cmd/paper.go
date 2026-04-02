package cmd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	paperCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	paperCmd.Flags().Bool("markdown", false, "Print raw markdown")
	paperCmd.Flags().String("project", "", "Resolve this paper inside one of your projects")
}

var paperCmd = &cobra.Command{
	Use:   "paper <paper-ref>",
	Short: "Show details for a canonical paper",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		markdownOut, _ := cmd.Flags().GetBool("markdown")
		projectID, _ := cmd.Flags().GetString("project")
		if jsonOut && markdownOut {
			return fmt.Errorf("--json and --markdown cannot be used together")
		}

		paperRef := args[0]
		if projectID != "" {
			return runProjectScopedPaper(cmd, paperRef, projectID, jsonOut, markdownOut)
		}
		if markdownOut {
			return runCanonicalPaperMarkdown(cmd, paperRef)
		}
		return runCanonicalPaper(cmd, paperRef, jsonOut)
	},
}

func runCanonicalPaper(cmd *cobra.Command, paperRef string, jsonOut bool) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	paper, err := api.FetchPublicPaper(paperRef)
	if err != nil {
		var apiErr *api.APIError
		if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
			legacyPaper, usedLegacy, legacyErr := fetchLegacyPaperFallback(paperRef)
			switch {
			case legacyErr == nil && usedLegacy:
				printLegacyPaperWarning(errOut, paperRef)
				return printPaper(out, legacyPaper, jsonOut)
			case legacyErr != nil:
				var legacyAPIError *api.APIError
				if errors.As(legacyErr, &legacyAPIError) && legacyAPIError.StatusCode != 404 {
					return fmt.Errorf("failed to fetch paper: %w", legacyErr)
				}
			}
			return recommendationHintError(paperRef, "")
		}
		return fmt.Errorf("failed to fetch paper: %w", err)
	}

	return printPaper(out, paper, jsonOut)
}

func runCanonicalPaperMarkdown(cmd *cobra.Command, paperRef string) error {
	out := cmd.OutOrStdout()
	errOut := cmd.ErrOrStderr()

	markdown, err := api.FetchPublicPaperMarkdown(paperRef)
	if err != nil {
		var pending *api.PaperMarkdownPendingError
		if errors.As(err, &pending) {
			fmt.Fprintln(out, pending.FriendlyMessage())
			return nil
		}

		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			switch {
			case apiErr.Code == "markdown_not_ready":
				fmt.Fprintln(out, canonicalMarkdownNotReadyMessage())
				return nil
			case apiErr.StatusCode == 404:
				markdown, usedLegacy, legacyErr := fetchLegacyPaperMarkdownFallback(paperRef)
				switch {
				case legacyErr == nil && usedLegacy:
					printLegacyPaperWarning(errOut, paperRef)
					fmt.Fprint(out, markdown)
					return nil
				case legacyErr != nil:
					var legacyPending *api.PaperMarkdownPendingError
					if errors.As(legacyErr, &legacyPending) {
						printLegacyPaperWarning(errOut, paperRef)
						fmt.Fprintln(out, legacyPending.FriendlyMessage())
						return nil
					}
					var legacyAPIError *api.APIError
					if errors.As(legacyErr, &legacyAPIError) && legacyAPIError.StatusCode != 404 {
						return fmt.Errorf("failed to fetch paper markdown: %w", legacyErr)
					}
				}
				return recommendationHintError(paperRef, "--markdown")
			}
		}

		return fmt.Errorf("failed to fetch paper markdown: %w", err)
	}

	fmt.Fprint(out, markdown)
	return nil
}

func runProjectScopedPaper(cmd *cobra.Command, paperRef, projectID string, jsonOut, markdownOut bool) error {
	tokens, err := loadRequiredAuth()
	if err != nil {
		return err
	}

	projectPaper, err := withAuth(&tokens, func(at string) (api.ProjectPaper, error) {
		return api.FetchProjectPaperForProject(at, projectID, paperRef)
	})
	if err != nil {
		return fmt.Errorf("failed to fetch paper in project: %w", err)
	}

	if markdownOut {
		markdown, err := withAuth(&tokens, func(at string) (string, error) {
			return api.FetchProjectPaperMarkdown(at, projectPaper.ID)
		})
		if err != nil {
			var pending *api.PaperMarkdownPendingError
			if errors.As(err, &pending) {
				fmt.Fprintln(cmd.OutOrStdout(), pending.FriendlyMessage())
				return nil
			}
			return fmt.Errorf("failed to fetch project paper markdown: %w", err)
		}
		fmt.Fprint(cmd.OutOrStdout(), markdown)
		return nil
	}

	if jsonOut {
		return printProjectPaperJSON(cmd.OutOrStdout(), projectPaper)
	}

	writeProjectPaper(cmd.OutOrStdout(), projectPaper)
	return nil
}

func fetchLegacyPaperFallback(paperRef string) (api.Paper, bool, error) {
	tokens, hasAuth, err := loadOptionalAuth()
	if err != nil {
		return api.Paper{}, false, err
	}

	paper, attempted, err := withOptionalAuth(&tokens, hasAuth, func(at string) (api.Paper, error) {
		return api.FetchLegacyPaper(at, paperRef)
	})
	return paper, attempted, err
}

func fetchLegacyPaperMarkdownFallback(paperRef string) (string, bool, error) {
	tokens, hasAuth, err := loadOptionalAuth()
	if err != nil {
		return "", false, err
	}

	markdown, attempted, err := withOptionalAuth(&tokens, hasAuth, func(at string) (string, error) {
		return api.FetchLegacyPaperMarkdown(at, paperRef)
	})
	return markdown, attempted, err
}

func printLegacyPaperWarning(errOut interface{ Write([]byte) (int, error) }, ref string) {
	fmt.Fprintf(errOut, "Warning: %q looks like a recommendation ID. `pz paper <id>` compatibility will be removed; use `pz rec %s` instead.\n", ref, ref)
}

func canonicalMarkdownNotReadyMessage() string {
	return "Markdown is not ready for this paper yet. Nothing was queued. Try again later, or use `pz rec <id> --markdown` from one of your projects."
}

func recommendationHintError(ref, suffix string) error {
	if suffix != "" {
		return fmt.Errorf("paper not found. If this is a recommendation ID, use `pz rec %s %s`", ref, suffix)
	}
	return fmt.Errorf("paper not found. If this is a recommendation ID, use `pz rec %s`", ref)
}

func printPaper(out interface{ Write([]byte) (int, error) }, paper api.Paper, jsonOut bool) error {
	if jsonOut {
		data, err := json.MarshalIndent(paper, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Fprintln(out, string(data))
		return nil
	}

	writeCanonicalPaper(out, paper)
	return nil
}

func printProjectPaperJSON(out interface{ Write([]byte) (int, error) }, projectPaper api.ProjectPaper) error {
	data, err := json.MarshalIndent(projectPaper, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Fprintln(out, string(data))
	return nil
}
