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

	if err := config.SaveTokens(tokens); err != nil {
		return config.Tokens{}, fmt.Errorf("failed to save tokens: %w", err)
	}

	fmt.Println("Logged in!")
	return tokens, nil
}

func loadAuth() (config.Tokens, error) {
	tokens, err := config.LoadTokens()
	if err != nil {
		fmt.Println("Not logged in.")
		return loginFunc()
	}

	if time.Now().Unix() >= tokens.ExpiresAt {
		if err := refreshSession(&tokens); err != nil {
			fmt.Fprintf(os.Stderr, "Token refresh failed: %v\n", err)
			if err := reauthenticate(&tokens); err != nil {
				return config.Tokens{}, err
			}
		}
	}

	return tokens, nil
}

// withAuth calls fn with the current access token. On 401 it attempts a refresh,
// then falls back to OTP login if refresh also fails.
func withAuth[T any](tokens *config.Tokens, fn func(string) (T, error)) (T, error) {
	result, err := fn(tokens.AccessToken)
	if errors.Is(err, api.ErrUnauthorized) {
		if refreshErr := refreshSession(tokens); refreshErr == nil {
			return fn(tokens.AccessToken)
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

func refreshSession(tokens *config.Tokens) error {
	if tokens.RefreshToken == "" {
		return errors.New("missing refresh token")
	}

	newTokens, err := refreshAccessTokenFunc(tokens.RefreshToken)
	if err != nil {
		return err
	}
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
