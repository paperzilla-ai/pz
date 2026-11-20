package api

import (
	"bytes"
	"context"
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
	resp, err := doRequestWithContext(context.Background(), method, path, body, accessToken, func(int) int64 {
		return successLimit
	})
	return resp.body, resp.status, err
}

type responseDetails struct {
	body        []byte
	status      int
	contentType string
}

func doRequestWithContext(ctx context.Context, method, path string, body any, accessToken string, successLimit func(int) int64) (responseDetails, error) {
	url := config.APIURL() + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return responseDetails{}, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return responseDetails{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", supportedUserAgent())
	req.Header.Set("X-Paperzilla-Client", supportedClientHeader)
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return responseDetails{}, err
	}
	defer resp.Body.Close()

	responseLimit := successLimit(resp.StatusCode)
	if resp.StatusCode >= 400 {
		responseLimit = maxJSONResponseBytes
	}
	result := responseDetails{status: resp.StatusCode, contentType: resp.Header.Get("Content-Type")}
	result.body, err = readResponseBody(resp.Body, responseLimit)
	if err != nil {
		return result, err
	}

	if resp.StatusCode == 401 {
		return result, ErrUnauthorized
	}
	if resp.StatusCode >= 400 {
		return result, parseAPIError(resp.StatusCode, result.body)
	}

	return result, nil
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
