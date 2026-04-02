package api

import (
	"encoding/json"
	"fmt"
	"strings"
)

type APIError struct {
	StatusCode int
	Detail     string
	Code       string
	Body       string
}

func (e *APIError) Error() string {
	switch {
	case strings.TrimSpace(e.Detail) != "" && strings.TrimSpace(e.Code) != "":
		return fmt.Sprintf("HTTP %d: %s (%s)", e.StatusCode, e.Detail, e.Code)
	case strings.TrimSpace(e.Detail) != "":
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Detail)
	case strings.TrimSpace(e.Body) != "":
		return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	default:
		return fmt.Sprintf("HTTP %d", e.StatusCode)
	}
}

type apiErrorPayload struct {
	Detail any    `json:"detail"`
	Code   string `json:"code"`
}

func parseAPIError(statusCode int, body []byte) *APIError {
	err := &APIError{
		StatusCode: statusCode,
		Body:       strings.TrimSpace(string(body)),
	}

	var payload apiErrorPayload
	if json.Unmarshal(body, &payload) == nil {
		err.Code = strings.TrimSpace(payload.Code)
		switch detail := payload.Detail.(type) {
		case string:
			err.Detail = strings.TrimSpace(detail)
		case nil:
		default:
			if data, marshalErr := json.Marshal(detail); marshalErr == nil {
				err.Detail = strings.TrimSpace(string(data))
			}
		}
	}

	return err
}
