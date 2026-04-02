package api

import (
	"encoding/json"
	"fmt"
)

func FetchProjectPaper(accessToken, id string) (ProjectPaper, error) {
	path := fmt.Sprintf("/api/project-papers/%s", id)
	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return ProjectPaper{}, err
	}

	var projectPaper ProjectPaper
	if err := json.Unmarshal(body, &projectPaper); err != nil {
		return ProjectPaper{}, err
	}

	return projectPaper, nil
}

func FetchProjectPaperForProject(accessToken, projectID, paperRef string) (ProjectPaper, error) {
	path := fmt.Sprintf("/api/projects/%s/papers/%s", projectID, paperRef)
	body, err := doRequest("GET", path, nil, accessToken)
	if err != nil {
		return ProjectPaper{}, err
	}

	var projectPaper ProjectPaper
	if err := json.Unmarshal(body, &projectPaper); err != nil {
		return ProjectPaper{}, err
	}

	return projectPaper, nil
}

func FetchProjectPaperMarkdown(accessToken, id string) (string, error) {
	return fetchMarkdown(fmt.Sprintf("/api/project-papers/%s/markdown", id), accessToken)
}

func SetProjectPaperFeedback(accessToken, id, vote, downvoteReason string) (Feedback, error) {
	path := fmt.Sprintf("/api/project-papers/%s/feedback", id)
	payload := map[string]any{
		"vote": vote,
	}
	if downvoteReason != "" {
		payload["downvote_reason"] = downvoteReason
	}

	body, err := doRequest("PUT", path, payload, accessToken)
	if err != nil {
		return Feedback{}, err
	}

	var feedback Feedback
	if err := json.Unmarshal(body, &feedback); err != nil {
		return Feedback{}, err
	}

	return feedback, nil
}

func ClearProjectPaperFeedback(accessToken, id string) error {
	path := fmt.Sprintf("/api/project-papers/%s/feedback", id)
	_, err := doRequest("DELETE", path, nil, accessToken)
	return err
}
