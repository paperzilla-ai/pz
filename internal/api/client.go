package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/paperzilla/pz/internal/config"
)

var ErrUnauthorized = errors.New("unauthorized")

const supportedClientHeader = "cli"

const (
	maxJSONResponseBytes     int64 = 16 << 20
	maxMarkdownResponseBytes int64 = 64 << 20
)

var clientVersion = "dev"

func SetClientVersion(version string) {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		trimmed = "dev"
	}
	clientVersion = trimmed
}

func supportedUserAgent() string {
	return "paperzilla-pz/" + clientVersion
}

func doRequest(method, path string, body any, accessToken string) ([]byte, error) {
	respBody, _, err := doRequestDetailed(method, path, body, accessToken, maxJSONResponseBytes)
	return respBody, err
}

func doRequestDetailed(method, path string, body any, accessToken string, successLimit int64) ([]byte, int, error) {
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
	req.Header.Set("User-Agent", supportedUserAgent())
	req.Header.Set("X-Paperzilla-Client", supportedClientHeader)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	responseLimit := successLimit
	if resp.StatusCode >= 400 {
		responseLimit = maxJSONResponseBytes
	}
	respBody, err := readResponseBody(resp.Body, responseLimit)
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

func readResponseBody(body io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("response body exceeds %d-byte limit", limit)
	}
	return data, nil
}
