package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/paperzilla/pz/internal/config"
)

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type cliAccessResponse struct {
	Allowed bool `json:"allowed"`
}

func SendOTP(email string) error {
	_, err := doRequest("POST", "/api/auth/otp", map[string]string{
		"email": email,
	}, "")
	return err
}

func VerifyOTP(email, code string) (config.Tokens, error) {
	body, err := doRequest("POST", "/api/auth/verify", map[string]string{
		"email": email,
		"code":  code,
	}, "")
	if err != nil {
		return config.Tokens{}, err
	}

	var resp authResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return config.Tokens{}, err
	}

	return config.Tokens{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second).Unix(),
	}, nil
}

func RefreshAccessToken(accessToken, refreshToken string) (config.Tokens, error) {
	body, err := doRequest("POST", "/api/auth/refresh", map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}, "")
	if err != nil {
		return config.Tokens{}, err
	}

	var resp authResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return config.Tokens{}, err
	}

	return config.Tokens{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second).Unix(),
	}, nil
}

func CheckCLIAccess(accessToken string) error {
	body, err := doRequest("GET", "/api/auth/cli-access", nil, accessToken)
	if err != nil {
		return err
	}

	var resp cliAccessResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("invalid CLI access response: %w", err)
	}
	if !resp.Allowed {
		return fmt.Errorf("CLI access response did not allow this account")
	}
	return nil
}
