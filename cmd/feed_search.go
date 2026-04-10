package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

var feedSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search the full feed for one project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		projectID, _ := cmd.Flags().GetString("project-id")
		query, _ := cmd.Flags().GetString("query")
		feedbackFilter, _ := cmd.Flags().GetString("feedback-filter")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		jsonOut, _ := cmd.Flags().GetBool("json")

		normalizedQuery, err := api.NormalizeFeedSearchQuery(query)
		if err != nil {
			return fmt.Errorf("invalid search request: %w", err)
		}
		if err := validateFeedSearchPagination(cmd, limit, offset); err != nil {
			return err
		}

		var mustRead *bool
		if cmd.Flags().Changed("must-read") {
			value, _ := cmd.Flags().GetBool("must-read")
			mustRead = &value
		}

		tokens, err := loadRequiredAuth()
		if err != nil {
			return err
		}

		search, err := withAuth(&tokens, func(at string) (api.FeedSearchResponse, error) {
			return api.FetchFeedSearch(at, projectID, api.FeedSearchOptions{
				Query:          normalizedQuery,
				FeedbackFilter: feedbackFilter,
				MustRead:       mustRead,
				Limit:          limit,
				Offset:         offset,
			})
		})
		if err != nil {
			return wrapFeedSearchError(err)
		}

		if jsonOut {
			data, err := json.MarshalIndent(search, "", "  ")
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

		fmt.Fprintf(out, "%s — %d papers\n", project.Name, len(search.Items))
		fmt.Fprintf(out, "Query: %s\n", search.Query)
		fmt.Fprintf(out, "Has more: %t\n\n", search.HasMore)
		writeProjectPaperFeedList(out, search.Items)
		return nil
	},
}

func init() {
	feedSearchCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	feedSearchCmd.Flags().String("project-id", "", "Project ID to search")
	feedSearchCmd.Flags().StringP("query", "q", "", "Search query")
	feedSearchCmd.Flags().String("feedback-filter", "all", "Filter results by feedback (all, unrated, liked, disliked, starred, not-relevant, low-quality)")
	feedSearchCmd.Flags().BoolP("must-read", "m", false, "Only show must-read papers")
	feedSearchCmd.Flags().IntP("limit", "n", 0, "Limit number of results")
	feedSearchCmd.Flags().Int("offset", 0, "Number of results to skip")
	_ = feedSearchCmd.MarkFlagRequired("project-id")
	_ = feedSearchCmd.MarkFlagRequired("query")
}

func wrapFeedSearchError(err error) error {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) && apiErr.StatusCode == 422 {
		detail := strings.TrimSpace(apiErr.Detail)
		if detail == "" {
			detail = strings.TrimSpace(apiErr.Body)
		}
		if detail == "" {
			return fmt.Errorf("invalid search request")
		}
		return fmt.Errorf("invalid search request: %s", detail)
	}

	return fmt.Errorf("failed to search feed: %w", err)
}

func validateFeedSearchPagination(cmd *cobra.Command, limit, offset int) error {
	if cmd.Flags().Changed("limit") {
		switch {
		case limit < 1:
			return fmt.Errorf("invalid search request: limit must be at least 1")
		case limit > 100:
			return fmt.Errorf("invalid search request: limit must be at most 100")
		}
	}
	if offset < 0 {
		return fmt.Errorf("invalid search request: offset must be at least 0")
	}
	return nil
}
