package api

import (
	"encoding/json"
	"fmt"
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
