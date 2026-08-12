package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
)

var (
	loginFunc              = runLogin
	refreshAccessTokenFunc = api.RefreshAccessToken
	checkCLIAccessFunc     = api.CheckCLIAccess
	saveTokensFunc         = config.SaveTokens
)

func runLogin() (config.Tokens, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Println("Sending magic link...")
	if err := api.SendOTP(email); err != nil {
		return config.Tokens{}, fmt.Errorf("failed to send OTP: %w", err)
	}

	fmt.Print("Check your email, enter the code: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	tokens, err := api.VerifyOTP(email, code)
	if err != nil {
		return config.Tokens{}, fmt.Errorf("failed to verify OTP: %w", err)
	}

	// The broker strictly checks CLI entitlement before consuming the OTP.
	// Persist its session before any command preflight can fail transiently.
	if err := config.SaveTokens(tokens); err != nil {
		return config.Tokens{}, fmt.Errorf("failed to save tokens: %w", err)
	}

	fmt.Println("Logged in!")
	return tokens, nil
}

func loadAuth() (config.Tokens, error) {
	return loadRequiredAuth()
}

func loadRequiredAuth() (config.Tokens, error) {
	tokens, err := config.LoadTokens()
	if err != nil {
		fmt.Println("Not logged in.")
		tokens, err = loginFunc()
		if err != nil {
			return config.Tokens{}, err
		}
	}

	if time.Now().Unix() >= tokens.ExpiresAt {
		if err := refreshSession(&tokens); err != nil {
			if api.IsCLIAccessError(err) {
				return config.Tokens{}, err
			}
			fmt.Fprintf(os.Stderr, "Token refresh failed: %s\n", terminalSafeInline(err.Error()))
			if err := reauthenticate(&tokens); err != nil {
				return config.Tokens{}, err
			}
		}
	}
	if err := checkCLIAccessFunc(tokens.AccessToken); err != nil {
		return config.Tokens{}, fmt.Errorf("CLI access check failed: %w", err)
	}

	return tokens, nil
}

func loadOptionalAuth() (config.Tokens, bool, error) {
	tokens, err := config.LoadTokens()
	if err != nil {
		return config.Tokens{}, false, nil
	}

	if time.Now().Unix() >= tokens.ExpiresAt {
		if err := refreshSession(&tokens); err != nil {
			if api.IsCLIAccessError(err) {
				return config.Tokens{}, false, err
			}
			return config.Tokens{}, false, nil
		}
	}
	if err := checkCLIAccessFunc(tokens.AccessToken); err != nil {
		return config.Tokens{}, false, fmt.Errorf("CLI access check failed: %w", err)
	}

	return tokens, true, nil
}

// withAuth calls fn with the current access token. On 401 it attempts a refresh,
// then falls back to OTP login if refresh also fails.
func withAuth[T any](tokens *config.Tokens, fn func(string) (T, error)) (T, error) {
	result, err := fn(tokens.AccessToken)
	if errors.Is(err, api.ErrUnauthorized) {
		if refreshErr := refreshSession(tokens); refreshErr == nil {
			return fn(tokens.AccessToken)
		} else if api.IsCLIAccessError(refreshErr) {
			var zero T
			return zero, refreshErr
		}

		fmt.Println("Session expired. Please log in again.")
		if loginErr := reauthenticate(tokens); loginErr != nil {
			var zero T
			return zero, loginErr
		}
		return fn(tokens.AccessToken)
	}
	return result, err
}

func withOptionalAuth[T any](tokens *config.Tokens, hasAuth bool, fn func(string) (T, error)) (T, bool, error) {
	var zero T
	if !hasAuth {
		return zero, false, nil
	}

	result, err := fn(tokens.AccessToken)
	if errors.Is(err, api.ErrUnauthorized) {
		if refreshErr := refreshSession(tokens); refreshErr != nil {
			if api.IsCLIAccessError(refreshErr) {
				return zero, false, refreshErr
			}
			return zero, false, nil
		}
		result, err = fn(tokens.AccessToken)
	}
	if errors.Is(err, api.ErrUnauthorized) {
		return zero, false, nil
	}

	return result, true, err
}

func refreshSession(tokens *config.Tokens) error {
	if tokens.RefreshToken == "" {
		return errors.New("missing refresh token")
	}

	newTokens, err := refreshAccessTokenFunc(tokens.AccessToken, tokens.RefreshToken)
	if err != nil {
		return err
	}
	// The broker strictly checks CLI entitlement before rotating the refresh
	// token. Save the returned pair before any later command preflight.
	if err := saveTokensFunc(newTokens); err != nil {
		return fmt.Errorf("failed to save refreshed tokens: %w", err)
	}

	*tokens = newTokens
	return nil
}

func reauthenticate(tokens *config.Tokens) error {
	newTokens, err := loginFunc()
	if err != nil {
		return err
	}

	*tokens = newTokens
	return nil
}
