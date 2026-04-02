package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

func FetchPublicPaper(id string) (Paper, error) {
	path := fmt.Sprintf("/api/public/papers/%s", id)
	body, err := doRequest("GET", path, nil, "")
	if err != nil {
		return Paper{}, err
	}

	var paper Paper
	if err := json.Unmarshal(body, &paper); err != nil {
		return Paper{}, err
	}

	return paper, nil
}

// FetchPaper is the legacy authenticated paper lookup used as a temporary
// compatibility path while the CLI migrates to explicit public/project-paper
// endpoints.
func FetchPaper(accessToken, id string) (Paper, error) {
	return FetchLegacyPaper(accessToken, id)
}

func FetchLegacyPaper(accessToken, id string) (Paper, error) {
	path := fmt.Sprintf("/api/papers/%s", id)
	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return Paper{}, err
	}

	var paper Paper
	if err := json.Unmarshal(body, &paper); err != nil {
		return Paper{}, err
	}

	return paper, nil
}

// FetchPaperMarkdown is the legacy authenticated markdown lookup used as a
// temporary compatibility path while the CLI migrates to explicit
// public/project-paper endpoints.
func FetchPaperMarkdown(accessToken, id string) (string, error) {
	return FetchLegacyPaperMarkdown(accessToken, id)
}

func FetchPublicPaperMarkdown(id string) (string, error) {
	return fetchMarkdown(fmt.Sprintf("/api/public/papers/%s/markdown", id), "")
}

func FetchLegacyPaperMarkdown(accessToken, id string) (string, error) {
	return fetchMarkdown(fmt.Sprintf("/api/papers/%s/markdown", id), accessToken)
}

func fetchMarkdown(path, accessToken string) (string, error) {
	body, _, err := doRequestDetailed("GET", path, nil, accessToken)
	if errors.Is(err, ErrUnauthorized) {
		return "", err
	}

	if pending := parsePaperMarkdownPending(body); pending != nil {
		return "", pending
	}

	if err != nil {
		return "", err
	}

	return string(body), nil
}

type PaperMarkdownPendingError struct {
	Detail  string `json:"detail"`
	Code    string `json:"code"`
	JobID   string `json:"job_id"`
	Created bool   `json:"created"`
}

func (e *PaperMarkdownPendingError) Error() string {
	return e.Detail
}

func (e *PaperMarkdownPendingError) FriendlyMessage() string {
	if e.Code == "markdown_already_queued" {
		return "Markdown is already being prepared. Try again in a minute or so."
	}
	return "Markdown is being prepared. Try again in a minute or so."
}

func parsePaperMarkdownPending(body []byte) *PaperMarkdownPendingError {
	var pending PaperMarkdownPendingError
	if err := json.Unmarshal(body, &pending); err != nil {
		return nil
	}
	if pending.Code != "markdown_queued" && pending.Code != "markdown_already_queued" {
		return nil
	}
	return &pending
}
