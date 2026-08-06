package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	CLIUpgradeRequiredCode        = "CLI_UPGRADE_REQUIRED"
	CLIEntitlementUnavailableCode = "CLI_ENTITLEMENT_UNAVAILABLE"
)

type APIError struct {
	StatusCode         int
	Detail             string
	Code               string
	UpgradeDestination string
	UpgradePath        string
	Body               string
}

func (e *APIError) Error() string {
	var message string
	switch {
	case strings.TrimSpace(e.Detail) != "" && strings.TrimSpace(e.Code) != "":
		message = fmt.Sprintf("HTTP %d: %s (%s)", e.StatusCode, e.Detail, e.Code)
	case strings.TrimSpace(e.Detail) != "":
		message = fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Detail)
	case strings.TrimSpace(e.Body) != "":
		message = fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Body)
	default:
		message = fmt.Sprintf("HTTP %d", e.StatusCode)
	}

	if upgradePath := strings.TrimSpace(e.UpgradePath); upgradePath != "" {
		return fmt.Sprintf("%s; upgrade: %s", message, upgradePath)
	}
	return message
}

type apiErrorPayload struct {
	Detail             any    `json:"detail"`
	Code               string `json:"code"`
	UpgradeDestination string `json:"upgrade_destination"`
	UpgradePath        string `json:"upgrade_path"`
}

func IsCLIAccessError(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return apiErr.Code == CLIUpgradeRequiredCode || apiErr.Code == CLIEntitlementUnavailableCode
}

func parseAPIError(statusCode int, body []byte) *APIError {
	err := &APIError{
		StatusCode: statusCode,
		Body:       strings.TrimSpace(string(body)),
	}

	var payload apiErrorPayload
	if json.Unmarshal(body, &payload) == nil {
		err.Code = strings.TrimSpace(payload.Code)
		err.UpgradeDestination = strings.TrimSpace(payload.UpgradeDestination)
		err.UpgradePath = strings.TrimSpace(payload.UpgradePath)
		switch detail := payload.Detail.(type) {
		case string:
			err.Detail = strings.TrimSpace(detail)
		case map[string]any:
			if err.Code == "" {
				if code, ok := detail["code"].(string); ok {
					err.Code = strings.TrimSpace(code)
				}
			}
			if destination, ok := detail["upgrade_destination"].(string); ok {
				err.UpgradeDestination = strings.TrimSpace(destination)
			}
			if path, ok := detail["upgrade_path"].(string); ok {
				err.UpgradePath = strings.TrimSpace(path)
			}
			for _, key := range []string{"message", "detail"} {
				if message, ok := detail[key].(string); ok && strings.TrimSpace(message) != "" {
					err.Detail = strings.TrimSpace(message)
					break
				}
			}
			if err.Detail == "" {
				if data, marshalErr := json.Marshal(detail); marshalErr == nil {
					err.Detail = strings.TrimSpace(string(data))
				}
			}
		case nil:
		default:
			if data, marshalErr := json.Marshal(detail); marshalErr == nil {
				err.Detail = strings.TrimSpace(string(data))
			}
		}
	}

	return err
}
