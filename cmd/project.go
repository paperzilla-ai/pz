package cmd

import (
	"fmt"
	"time"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(projectCmd)
	projectCmd.AddCommand(projectListCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project [id]",
	Short: "Show project details or manage projects",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}

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

		fmt.Printf("Name:             %s\n", p.Name)
		fmt.Printf("ID:               %s\n", p.ID)
		fmt.Printf("Mode:             %s\n", p.Mode)
		fmt.Printf("Visibility:       %s\n", p.Visibility)
		fmt.Printf("Matching State:   %s\n", p.MatchingState)
		fmt.Printf("Email Frequency:  %s\n", p.EmailFrequency)
		fmt.Printf("Email Time:       %s\n", p.EmailTime)
		fmt.Printf("Max Candidates:   %d\n", p.MaxCandidates)
		fmt.Printf("Max Papers/Digest:%d\n", p.MaxPapersPerDigests)
		fmt.Printf("Created:          %s\n", formatTime(p.CreatedAt))
		fmt.Printf("Activated:        %s\n", formatTime(p.ActivatedAt))
		fmt.Printf("Last Digest:      %s\n", formatTime(p.LastDigestSentAt))

		if p.InterestDescription != "" {
			fmt.Printf("\nInterest:\n  %s\n", p.InterestDescription)
		}

		return nil
	},
}

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List your projects",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		if len(projects) == 0 {
			fmt.Println("No projects found.")
			return nil
		}

		fmt.Printf("%-36s  %-25s  %-10s  %-10s  %s\n", "ID", "NAME", "MODE", "VISIBILITY", "CREATED")
		for _, p := range projects {
			name := p.Name
			if len(name) > 25 {
				name = name[:22] + "..."
			}
			fmt.Printf("%-36s  %-25s  %-10s  %-10s  %s\n", p.ID, name, p.Mode, p.Visibility, formatTime(p.CreatedAt))
		}

		return nil
	},
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
