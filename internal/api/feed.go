package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

type Author struct {
	Name string `json:"name"`
}

type Source struct {
	Name string `json:"name"`
}

type Feedback struct {
	Vote           string `json:"vote"`
	DownvoteReason string `json:"downvote_reason"`
	UpdatedAt      string `json:"updated_at"`
}

type Paper struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Abstract       string   `json:"abstract"`
	Authors        []Author `json:"authors"`
	PublishedDate  string   `json:"published_date"`
	PdfURL         string   `json:"pdf_url"`
	URL            string   `json:"url"`
	DOI            string   `json:"doi"`
	VenueName      string   `json:"venue_name"`
	ReferenceLabel string   `json:"reference_label"`
	SourceID       int      `json:"source_id"`
	Source         *Source  `json:"source"`
	SourcePaperID  string   `json:"source_paper_id"`
	ShortID        string   `json:"short_id"`
	Slug           string   `json:"slug"`
	Metadata       any      `json:"metadata"`
	MarkdownReady  bool     `json:"markdown_ready"`
}

type ProjectPaper struct {
	ID               string    `json:"id"`
	ShortID          string    `json:"short_id"`
	Slug             string    `json:"slug"`
	PaperTitle       string    `json:"paper_title"`
	Summary          string    `json:"summary"`
	RelevanceScore   float64   `json:"relevance_score"`
	RelevanceClass   int       `json:"relevance_class"`
	CombinedScore    float64   `json:"combined_score"`
	MatchingDetails  any       `json:"matching_details"`
	PersonalizedNote string    `json:"personalized_note"`
	ReadyAt          string    `json:"ready_at"`
	Feedback         *Feedback `json:"feedback"`
	Paper            Paper     `json:"paper"`
}

type FeedPaper = ProjectPaper

type FeedResponse struct {
	Items  []ProjectPaper `json:"items"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type FeedOptions struct {
	MustReadOnly bool
	Since        string
	Limit        int
	Offset       int
}

type FeedTokenResponse struct {
	Token     string `json:"token"`
	CreatedAt string `json:"created_at"`
}

func FetchFeedToken(accessToken string) (FeedTokenResponse, error) {
	body, err := doRequest("POST", "/api/auth/feed-token", nil, accessToken)
	if err != nil {
		return FeedTokenResponse{}, err
	}

	var resp FeedTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return FeedTokenResponse{}, err
	}

	return resp, nil
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
