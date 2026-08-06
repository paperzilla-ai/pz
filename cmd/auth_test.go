package cmd

import (
	"errors"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
)

func TestWithAuthRefreshesOnUnauthorized(t *testing.T) {
	origLogin := loginFunc
	origRefresh := refreshAccessTokenFunc
	origCheckAccess := checkCLIAccessFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		loginFunc = origLogin
		refreshAccessTokenFunc = origRefresh
		checkCLIAccessFunc = origCheckAccess
		saveTokensFunc = origSave
	})
	var checkCalls int
	checkCLIAccessFunc = func(string) error {
		checkCalls++
		return nil
	}

	var refreshCalls int
	var saveCalls int
	var loginCalls int

	loginFunc = func() (config.Tokens, error) {
		loginCalls++
		return config.Tokens{}, errors.New("login should not be called")
	}
	refreshAccessTokenFunc = func(accessToken, refreshToken string) (config.Tokens, error) {
		refreshCalls++
		if accessToken != "access-1" {
			t.Fatalf("access token = %q, want %q", accessToken, "access-1")
		}
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
	if checkCalls != 0 {
		t.Fatalf("post-refresh preflight calls = %d, want 0 before tokens are saved", checkCalls)
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
	origCheckAccess := checkCLIAccessFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		loginFunc = origLogin
		refreshAccessTokenFunc = origRefresh
		checkCLIAccessFunc = origCheckAccess
		saveTokensFunc = origSave
	})
	checkCLIAccessFunc = func(string) error { return nil }

	var refreshCalls int
	var saveCalls int
	var loginCalls int

	refreshAccessTokenFunc = func(string, string) (config.Tokens, error) {
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

func TestWithAuthDoesNotLoginLoopWhenRefreshRequiresUpgrade(t *testing.T) {
	origLogin := loginFunc
	origRefresh := refreshAccessTokenFunc
	origCheckAccess := checkCLIAccessFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		loginFunc = origLogin
		refreshAccessTokenFunc = origRefresh
		checkCLIAccessFunc = origCheckAccess
		saveTokensFunc = origSave
	})

	loginCalls := 0
	loginFunc = func() (config.Tokens, error) {
		loginCalls++
		return config.Tokens{}, errors.New("login should not be called")
	}
	refreshAccessTokenFunc = func(string, string) (config.Tokens, error) {
		return config.Tokens{}, &api.APIError{
			StatusCode:         403,
			Code:               api.CLIUpgradeRequiredCode,
			Detail:             "Upgrade to continue.",
			UpgradeDestination: "early_user_offer",
			UpgradePath:        "/early-user-offer",
		}
	}
	checkCLIAccessFunc = func(string) error { return nil }
	saveTokensFunc = func(config.Tokens) error { return nil }

	tokens := config.Tokens{AccessToken: "old-access", RefreshToken: "old-refresh"}
	_, err := withAuth(&tokens, func(string) (string, error) {
		return "", api.ErrUnauthorized
	})

	if !api.IsCLIAccessError(err) {
		t.Fatalf("err = %v, want CLI access error", err)
	}
	if !strings.Contains(err.Error(), "/early-user-offer") {
		t.Fatalf("err = %q, want upgrade path", err)
	}
	if loginCalls != 0 {
		t.Fatalf("login calls = %d, want 0", loginCalls)
	}
}

func TestRefreshSessionRetainsStoredTokensWhenEntitlementIsUnavailable(t *testing.T) {
	origRefresh := refreshAccessTokenFunc
	origCheckAccess := checkCLIAccessFunc
	origSave := saveTokensFunc
	t.Cleanup(func() {
		refreshAccessTokenFunc = origRefresh
		checkCLIAccessFunc = origCheckAccess
		saveTokensFunc = origSave
	})

	var saveCalls int
	refreshAccessTokenFunc = func(accessToken, refreshToken string) (config.Tokens, error) {
		if accessToken != "old-access" || refreshToken != "old-refresh" {
			t.Fatalf("refresh credentials = %q/%q", accessToken, refreshToken)
		}
		return config.Tokens{}, &api.APIError{
			StatusCode: 503,
			Code:       api.CLIEntitlementUnavailableCode,
			Detail:     "Please try again shortly.",
		}
	}
	checkCLIAccessFunc = func(string) error {
		t.Fatal("post-refresh preflight must not run without new tokens")
		return nil
	}
	saveTokensFunc = func(config.Tokens) error {
		saveCalls++
		return nil
	}

	tokens := config.Tokens{
		AccessToken:  "old-access",
		RefreshToken: "old-refresh",
		ExpiresAt:    100,
	}
	err := refreshSession(&tokens)

	if !api.IsCLIAccessError(err) {
		t.Fatalf("err = %v, want retryable CLI entitlement error", err)
	}
	if saveCalls != 0 {
		t.Fatalf("save calls = %d, want 0", saveCalls)
	}
	if tokens.AccessToken != "old-access" || tokens.RefreshToken != "old-refresh" {
		t.Fatalf("tokens changed after retryable failure: %+v", tokens)
	}
}
