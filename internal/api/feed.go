package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Author struct {
	Name string `json:"name"`
}

type Paper struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Abstract      string   `json:"abstract"`
	Authors       []Author `json:"authors"`
	PublishedDate string   `json:"published_date"`
	PdfURL        string   `json:"pdf_url"`
	URL           string   `json:"url"`
	DOI           string   `json:"doi"`
	SourceID      int      `json:"source_id"`
	SourcePaperID string   `json:"source_paper_id"`
	ShortID       string   `json:"short_id"`
	Slug          string   `json:"slug"`
	Metadata      any      `json:"metadata"`
}

type FeedPaper struct {
	ID               string  `json:"id"`
	ShortID          string  `json:"short_id"`
	Slug             string  `json:"slug"`
	PaperTitle       string  `json:"paper_title"`
	Summary          string  `json:"summary"`
	RelevanceScore   float64 `json:"relevance_score"`
	RelevanceClass   int     `json:"relevance_class"`
	CombinedScore    float64 `json:"combined_score"`
	MatchingDetails  any     `json:"matching_details"`
	PersonalizedNote string  `json:"personalized_note"`
	UserStarred      bool    `json:"user_starred"`
	UserClicked      bool    `json:"user_clicked"`
	ReadyAt          string  `json:"ready_at"`
	Paper            Paper   `json:"paper"`
}

type FeedResponse struct {
	Items  []FeedPaper `json:"items"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type FeedOptions struct {
	MustReadOnly bool
	Since        string
	Limit        int
	Offset       int
}

func FetchFeed(accessToken, projectID string, opts FeedOptions) (FeedResponse, error) {
	params := url.Values{}
	if opts.MustReadOnly {
		params.Set("must_read", "true")
	}
	if opts.Since != "" {
		params.Set("since", opts.Since)
	}
	if opts.Limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", opts.Limit))
	}
	if opts.Offset > 0 {
		params.Set("offset", fmt.Sprintf("%d", opts.Offset))
	}

	path := fmt.Sprintf("/api/projects/%s/feed", projectID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return FeedResponse{}, err
	}

	var resp FeedResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return FeedResponse{}, err
	}

	return resp, nil
}
