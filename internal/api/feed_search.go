package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

const (
	MinFeedSearchQueryLength = 3
	MaxFeedSearchQueryLength = 200
)

type FeedSearchOptions struct {
	Query          string
	FeedbackFilter string
	MustRead       *bool
	Limit          int
	Offset         int
}

type FeedSearchResponse struct {
	Items   []ProjectPaper `json:"items"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
	HasMore bool           `json:"has_more"`
	Query   string         `json:"query"`
}

func NormalizeFeedSearchQuery(query string) (string, error) {
	trimmed := strings.TrimSpace(query)
	switch {
	case len(trimmed) < MinFeedSearchQueryLength:
		return "", fmt.Errorf("Search query must be at least %d characters.", MinFeedSearchQueryLength)
	case len(trimmed) > MaxFeedSearchQueryLength:
		return "", fmt.Errorf("Search query must be at most %d characters.", MaxFeedSearchQueryLength)
	default:
		return trimmed, nil
	}
}

func FetchFeedSearch(accessToken, projectID string, opts FeedSearchOptions) (FeedSearchResponse, error) {
	query, err := NormalizeFeedSearchQuery(opts.Query)
	if err != nil {
		return FeedSearchResponse{}, err
	}

	params := url.Values{}
	params.Set("q", query)

	if feedbackFilter := strings.TrimSpace(opts.FeedbackFilter); feedbackFilter != "" {
		params.Set("feedback_filter", feedbackFilter)
	}
	if opts.MustRead != nil {
		params.Set("must_read", fmt.Sprintf("%t", *opts.MustRead))
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	path := fmt.Sprintf("/api/projects/%s/feed/search?%s", projectID, params.Encode())

	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return FeedSearchResponse{}, err
	}

	var resp FeedSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return FeedSearchResponse{}, err
	}

	return resp, nil
}
