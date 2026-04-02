package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/paperzilla/pz/internal/config"
)

var ErrUnauthorized = errors.New("unauthorized")

func doRequest(method, path string, body any, accessToken string) ([]byte, error) {
	respBody, _, err := doRequestDetailed(method, path, body, accessToken)
	return respBody, err
}

func doRequestDetailed(method, path string, body any, accessToken string) ([]byte, int, error) {
	url := config.APIURL() + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	if resp.StatusCode == 401 {
		return respBody, resp.StatusCode, ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, parseAPIError(resp.StatusCode, respBody)
	}

	return respBody, resp.StatusCode, nil
}
