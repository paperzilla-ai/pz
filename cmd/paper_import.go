package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

var (
	errPaperImportPending = errors.New("Markdown is being prepared. Run this command again shortly.")
	errPaperImportLogin   = errors.New("Not logged in. Run `pz login` first.")
)

func init() {
	paperCmd.AddCommand(newPaperImportCommand())
}

func newPaperImportCommand() *cobra.Command {
	command := &cobra.Command{
		Use:           "import <reference>",
		Hidden:        true,
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.ExactArgs(1),
		RunE:          runPaperImport,
	}
	command.Flags().Bool("markdown", false, "Print raw markdown")
	command.Flags().Bool("json", false, "Print the paper record and full-text status as JSON")
	return command
}

func runPaperImport(cmd *cobra.Command, args []string) error {
	markdown, _ := cmd.Flags().GetBool("markdown")
	jsonOut, _ := cmd.Flags().GetBool("json")
	if markdown == jsonOut {
		return errors.New("specify exactly one of --markdown or --json")
	}
	reference := strings.TrimSpace(args[0])
	if reference == "" {
		return errors.New("a paper reference is required")
	}

	// The import endpoint verifies CLI entitlement itself, including after
	// a silent refresh. It must never fall back to interactive login.
	tokens, hasAuth, err := loadOptionalSession()
	if err != nil {
		return err
	}
	type importOutput struct {
		record   api.PaperImportRecord
		markdown api.PaperImportResult
	}
	result, hasAuth, err := withOptionalAuth(&tokens, hasAuth, func(token string) (importOutput, error) {
		if jsonOut {
			record, err := api.ImportPaperRecord(cmd.Context(), token, reference)
			return importOutput{record: record}, err
		}
		markdown, err := api.ImportPaperMarkdown(cmd.Context(), token, reference)
		return importOutput{markdown: markdown}, err
	})
	if err != nil {
		var apiErr *api.APIError
		if markdown && errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound && apiErr.Code == "not_downloadable" {
			return errors.New("not downloadable")
		}
		return fmt.Errorf("failed to import paper: %w", err)
	}
	if !hasAuth {
		return errPaperImportLogin
	}
	if jsonOut {
		return writeJSON(cmd.OutOrStdout(), result.record)
	}
	if result.markdown.Queued {
		return errPaperImportPending
	}
	_, err = cmd.OutOrStdout().Write(result.markdown.Markdown)
	return err
}
