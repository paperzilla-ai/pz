package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

var sourceNames = map[int]string{
	1: "arxiv",
	2: "biorxiv",
	3: "medrxiv",
	4: "chinaxiv",
}

func init() {
	rootCmd.AddCommand(feedCmd)
	feedCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	feedCmd.Flags().BoolP("must-read", "m", false, "Only show must-read papers")
	feedCmd.Flags().StringP("since", "s", "", "Only papers ready after this date")
	feedCmd.Flags().IntP("limit", "n", 0, "Limit number of results")
	feedCmd.Flags().Bool("atom", false, "Print Atom feed URL for use in feed readers")
}

var feedCmd = &cobra.Command{
	Use:   "feed <project-id>",
	Short: "Show relevant curated papers for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tokens, err := loadAuth()
		if err != nil {
			return err
		}

		projectID := args[0]

		atom, _ := cmd.Flags().GetBool("atom")
		if atom {
			tokenResp, err := api.FetchFeedToken(tokens.AccessToken)
			if err != nil {
				return fmt.Errorf("failed to get feed token: %w", err)
			}
			fmt.Printf("%s/api/feed/atom/%s?token=%s\n", config.APIURL(), projectID, tokenResp.Token)
			return nil
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		mustRead, _ := cmd.Flags().GetBool("must-read")
		since, _ := cmd.Flags().GetString("since")
		limit, _ := cmd.Flags().GetInt("limit")

		opts := api.FeedOptions{
			MustReadOnly: mustRead,
			Since:        since,
			Limit:        limit,
		}

		feed, err := api.FetchFeed(tokens.AccessToken, projectID, opts)
		if err != nil {
			return fmt.Errorf("failed to fetch feed: %w", err)
		}

		if jsonOut {
			data, err := json.MarshalIndent(feed, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(data))
			return nil
		}

		project, err := api.FetchProject(tokens.AccessToken, projectID)
		if err != nil {
			return fmt.Errorf("failed to fetch project: %w", err)
		}

		fmt.Printf("%s — %d papers (total: %d)\n\n", project.Name, len(feed.Items), feed.Total)

		for _, p := range feed.Items {
			icon := "○ Related"
			if p.RelevanceClass == 2 {
				icon = "★ Must Read"
			}

			title := p.PaperTitle
			if len(title) > 80 {
				title = title[:77] + "..."
			}

			fmt.Printf("%s  %s\n", icon, title)

			surname := firstAuthorSurname(p.Paper.Authors)
			source := sourceNames[p.Paper.SourceID]
			date := formatTime(p.Paper.PublishedDate)
			score := int(p.RelevanceScore * 100)

			fmt.Printf("  %s · %s · %s · relevance: %d%%\n\n", surname, source, date, score)
		}

		return nil
	},
}

func firstAuthorSurname(authors []api.Author) string {
	if len(authors) == 0 {
		return "Unknown"
	}
	name := authors[0].Name
	parts := strings.Fields(name)
	surname := parts[len(parts)-1]
	if len(authors) > 1 {
		return surname + " et al."
	}
	return surname
}
