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
}

type FeedPaper struct {
	ID               string  `json:"id"`
	ShortID          string  `json:"short_id"`
	Slug             string  `json:"slug"`
	PaperTitle       string  `json:"paper_title"`
	Summary          string  `json:"summary"`
	RelevanceScore   float64 `json:"relevance_score"`
	RelevanceClass   int     `json:"relevance_class"`
	MatchingDetails  any     `json:"matching_details"`
	PersonalizedNote string  `json:"personalized_note"`
	UserStarred      bool    `json:"user_starred"`
	UserClicked      bool    `json:"user_clicked"`
	ReadyAt          string  `json:"ready_at"`
	Paper            Paper   `json:"paper"`
}

type FeedOptions struct {
	MustReadOnly bool
	Since        string
	Limit        int
}

func FetchFeed(accessToken, projectID string, opts FeedOptions) ([]FeedPaper, error) {
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

	path := fmt.Sprintf("/api/projects/%s/feed", projectID)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return nil, err
	}

	var papers []FeedPaper
	if err := json.Unmarshal(body, &papers); err != nil {
		return nil, err
	}

	return papers, nil
}
