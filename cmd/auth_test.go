package cmd

import (
	"errors"
	"testing"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
)

func TestWithAuthRefreshesOnUnauthorized(t *testing.T) {
	origLogin := loginFunc
	origRefresh := refreshAccessTokenFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		loginFunc = origLogin
		refreshAccessTokenFunc = origRefresh
		saveTokensFunc = origSave
	})

	var refreshCalls int
	var saveCalls int
	var loginCalls int

	loginFunc = func() (config.Tokens, error) {
		loginCalls++
		return config.Tokens{}, errors.New("login should not be called")
	}
	refreshAccessTokenFunc = func(refreshToken string) (config.Tokens, error) {
		refreshCalls++
		if refreshToken != "refresh-1" {
			t.Fatalf("refresh token = %q, want %q", refreshToken, "refresh-1")
		}
		return config.Tokens{
			AccessToken:  "access-2",
			RefreshToken: "refresh-2",
			ExpiresAt:    200,
		}, nil
	}
	saveTokensFunc = func(tokens config.Tokens) error {
		saveCalls++
		if tokens.AccessToken != "access-2" {
			t.Fatalf("saved access token = %q, want %q", tokens.AccessToken, "access-2")
		}
		return nil
	}

	tokens := config.Tokens{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		ExpiresAt:    100,
	}

	var callTokens []string
	result, err := withAuth(&tokens, func(accessToken string) (string, error) {
		callTokens = append(callTokens, accessToken)
		if len(callTokens) == 1 {
			return "", api.ErrUnauthorized
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("withAuth: %v", err)
	}

	if result != "ok" {
		t.Fatalf("result = %q, want %q", result, "ok")
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", refreshCalls)
	}
	if saveCalls != 1 {
		t.Fatalf("save calls = %d, want 1", saveCalls)
	}
	if loginCalls != 0 {
		t.Fatalf("login calls = %d, want 0", loginCalls)
	}
	if len(callTokens) != 2 {
		t.Fatalf("request calls = %d, want 2", len(callTokens))
	}
	if callTokens[1] != "access-2" {
		t.Fatalf("retried with token = %q, want %q", callTokens[1], "access-2")
	}
	if tokens.AccessToken != "access-2" || tokens.RefreshToken != "refresh-2" {
		t.Fatalf("tokens = %+v, want refreshed tokens", tokens)
	}
}

func TestWithAuthFallsBackToLoginWhenRefreshFails(t *testing.T) {
	origLogin := loginFunc
	origRefresh := refreshAccessTokenFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		loginFunc = origLogin
		refreshAccessTokenFunc = origRefresh
		saveTokensFunc = origSave
	})

	var refreshCalls int
	var saveCalls int
	var loginCalls int

	refreshAccessTokenFunc = func(string) (config.Tokens, error) {
		refreshCalls++
		return config.Tokens{}, errors.New("invalid refresh token")
	}
	saveTokensFunc = func(config.Tokens) error {
		saveCalls++
		return nil
	}
	loginFunc = func() (config.Tokens, error) {
		loginCalls++
		return config.Tokens{
			AccessToken:  "access-login",
			RefreshToken: "refresh-login",
			ExpiresAt:    300,
		}, nil
	}

	tokens := config.Tokens{
		AccessToken:  "access-old",
		RefreshToken: "refresh-old",
		ExpiresAt:    100,
	}

	var callTokens []string
	result, err := withAuth(&tokens, func(accessToken string) (string, error) {
		callTokens = append(callTokens, accessToken)
		if len(callTokens) == 1 {
			return "", api.ErrUnauthorized
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("withAuth: %v", err)
	}

	if result != "ok" {
		t.Fatalf("result = %q, want %q", result, "ok")
	}
	if refreshCalls != 1 {
		t.Fatalf("refresh calls = %d, want 1", refreshCalls)
	}
	if saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", saveCalls)
	}
	if loginCalls != 1 {
		t.Fatalf("login calls = %d, want 1", loginCalls)
	}
	if len(callTokens) != 2 {
		t.Fatalf("request calls = %d, want 2", len(callTokens))
	}
	if callTokens[1] != "access-login" {
		t.Fatalf("retried with token = %q, want %q", callTokens[1], "access-login")
	}
	if tokens.AccessToken != "access-login" || tokens.RefreshToken != "refresh-login" {
		t.Fatalf("tokens = %+v, want login tokens", tokens)
	}
}
