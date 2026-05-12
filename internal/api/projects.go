package api

import (
	"encoding/json"
	"fmt"
)

type ProjectSource struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
}

type ProjectCategory struct {
	ID         int     `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	SourceID   int     `json:"source_id"`
	SourceName string  `json:"source_name"`
	Weight     float64 `json:"weight"`
}

type Project struct {
	ID                  string            `json:"id"`
	Name                string            `json:"name"`
	Mode                string            `json:"mode"`
	Visibility          string            `json:"visibility"`
	InterestDescription string            `json:"interest_description"`
	EmailFrequency      string            `json:"email_frequency"`
	EmailTime           string            `json:"email_time"`
	MatchingState       string            `json:"matching_state"`
	MaxCandidates       int               `json:"max_candidates"`
	MaxPapersPerDigests int               `json:"max_papers_per_digests"`
	CreatedAt           string            `json:"created_at"`
	ActivatedAt         string            `json:"activated_at"`
	LastDigestSentAt    string            `json:"last_digest_sent_at"`
	PositiveKeywords    []string          `json:"positive_keywords"`
	NegativeKeywords    []string          `json:"negative_keywords"`
	Sources             []ProjectSource   `json:"sources"`
	Categories          []ProjectCategory `json:"categories"`
}

func FetchProjects(accessToken string) ([]Project, error) {
	body, err := doRequest("GET", "/api/projects", nil, accessToken)
	if err != nil {
		return nil, err
	}

	var projects []Project
	if err := json.Unmarshal(body, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

func FetchProject(accessToken, id string) (Project, error) {
	path := fmt.Sprintf("/api/projects/%s", id)
	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return Project{}, err
	}

	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return Project{}, err
	}

	project.normalizeMetadata()
	return project, nil
}

func (p *Project) normalizeMetadata() {
	if p.PositiveKeywords == nil {
		p.PositiveKeywords = []string{}
	}
	if p.NegativeKeywords == nil {
		p.NegativeKeywords = []string{}
	}
	if p.Sources == nil {
		p.Sources = []ProjectSource{}
	}
	if p.Categories == nil {
		p.Categories = []ProjectCategory{}
	}
}
