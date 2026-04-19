package cmd

import (
	"fmt"
	"time"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	projectCmd.PersistentFlags().BoolP("json", "j", false, "Output as JSON")
	projectCmd.AddCommand(projectListCmd)
}

type projectListItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Mode       string `json:"mode"`
	Visibility string `json:"visibility"`
}

var projectCmd = &cobra.Command{
	Use:   "project [id]",
	Short: "Show project details or manage projects",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		tokens, err := loadAuth()
		if err != nil {
			return err
		}

		p, err := withAuth(&tokens, func(at string) (api.Project, error) {
			return api.FetchProject(at, args[0])
		})
		if err != nil {
			return fmt.Errorf("failed to fetch project: %w", err)
		}

		out := cmd.OutOrStdout()
		if jsonOut {
			return writeJSON(out, p)
		}

		fmt.Fprintf(out, "Name:             %s\n", p.Name)
		fmt.Fprintf(out, "ID:               %s\n", p.ID)
		fmt.Fprintf(out, "Mode:             %s\n", p.Mode)
		fmt.Fprintf(out, "Visibility:       %s\n", p.Visibility)
		fmt.Fprintf(out, "Matching State:   %s\n", p.MatchingState)
		fmt.Fprintf(out, "Email Frequency:  %s\n", p.EmailFrequency)
		fmt.Fprintf(out, "Email Time:       %s\n", p.EmailTime)
		fmt.Fprintf(out, "Max Candidates:   %d\n", p.MaxCandidates)
		fmt.Fprintf(out, "Max Papers/Digest:%d\n", p.MaxPapersPerDigests)
		fmt.Fprintf(out, "Created:          %s\n", formatTime(p.CreatedAt))
		fmt.Fprintf(out, "Activated:        %s\n", formatTime(p.ActivatedAt))
		fmt.Fprintf(out, "Last Digest:      %s\n", formatTime(p.LastDigestSentAt))

		if p.InterestDescription != "" {
			fmt.Fprintf(out, "\nInterest:\n  %s\n", p.InterestDescription)
		}

		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		jsonOut, _ := cmd.Flags().GetBool("json")
		tokens, err := loadAuth()
		if err != nil {
			return err
		}

		projects, err := withAuth(&tokens, func(at string) ([]api.Project, error) {
			return api.FetchProjects(at)
		})
		if err != nil {
			return fmt.Errorf("failed to fetch projects: %w", err)
		}

		out := cmd.OutOrStdout()
		if jsonOut {
			return writeJSON(out, summarizeProjects(projects))
		}

		if len(projects) == 0 {
			fmt.Fprintf(out, "No projects found. Create your first project: %s\n", cliGettingStartedURL)
			return nil
		}

		fmt.Fprintf(out, "%-36s  %-25s  %-10s  %-10s  %s\n", "ID", "NAME", "MODE", "VISIBILITY", "CREATED")
		for _, p := range projects {
			name := p.Name
			if len(name) > 25 {
				name = name[:22] + "..."
			}
			fmt.Fprintf(out, "%-36s  %-25s  %-10s  %-10s  %s\n", p.ID, name, p.Mode, p.Visibility, formatTime(p.CreatedAt))
		}

		return nil
	},
}

func summarizeProjects(projects []api.Project) []projectListItem {
	items := make([]projectListItem, 0, len(projects))
	for _, project := range projects {
		items = append(items, projectListItem{
			ID:         project.ID,
			Name:       project.Name,
			Mode:       project.Mode,
			Visibility: project.Visibility,
		})
	}
	return items
}

func formatTime(s string) string {
	if s == "" {
		return "—"
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.Format("2006-01-02 15:04")
	}
	return s
}
