package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	feedCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	feedCmd.Flags().BoolP("must-read", "m", false, "Only show must-read papers")
	feedCmd.Flags().StringP("since", "s", "", "Only papers ready after this date")
	feedCmd.Flags().IntP("limit", "n", 0, "Limit number of results")
	feedCmd.Flags().Bool("atom", false, "Print Atom feed URL for use in feed readers")
	feedCmd.AddCommand(feedSearchCmd)
}

var feedCmd = &cobra.Command{
	Use:   "feed <project-id>",
	Short: "Show relevant curated papers for a project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		tokens, err := loadAuth()
		if err != nil {
			return err
		}

		projectID := args[0]

		atom, _ := cmd.Flags().GetBool("atom")
		if atom {
			tokenResp, err := withAuth(&tokens, func(at string) (api.FeedTokenResponse, error) {
				return api.FetchFeedToken(at)
			})
			if err != nil {
				return fmt.Errorf("failed to get feed token: %w", err)
			}
			fmt.Fprintf(out, "%s/api/feed/atom/%s?token=%s\n", config.APIURL(), projectID, tokenResp.Token)
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

		feed, err := withAuth(&tokens, func(at string) (api.FeedResponse, error) {
			return api.FetchFeed(at, projectID, opts)
		})
		if err != nil {
			return fmt.Errorf("failed to fetch feed: %w", err)
		}

		if jsonOut {
			data, err := json.MarshalIndent(feed, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Fprintln(out, string(data))
			return nil
		}

		project, err := withAuth(&tokens, func(at string) (api.Project, error) {
			return api.FetchProject(at, projectID)
		})
		if err != nil {
			return fmt.Errorf("failed to fetch project: %w", err)
		}

		fmt.Fprintf(out, "%s — %d papers (total: %d)\n\n", project.Name, len(feed.Items), feed.Total)

		writeProjectPaperFeedList(out, feed.Items)

		return nil
	},
}
