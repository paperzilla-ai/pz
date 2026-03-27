package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/paperzilla/pz/internal/config"
)

func FetchPaper(accessToken, id string) (Paper, error) {
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

func FetchPaperMarkdown(accessToken, id string) (string, error) {
	path := fmt.Sprintf("/api/papers/%s/markdown", id)
	req, err := http.NewRequest(http.MethodGet, config.APIURL()+path, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return "", ErrUnauthorized
	}

	if pending := parsePaperMarkdownPending(body); pending != nil {
		return "", pending
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
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
