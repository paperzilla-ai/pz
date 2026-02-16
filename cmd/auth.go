package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
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
		return runLogin()
	}

	if time.Now().Unix() >= tokens.ExpiresAt {
		tokens, err = api.RefreshAccessToken(tokens.RefreshToken)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Token refresh failed: %v\n", err)
			return runLogin()
		}
		if err := config.SaveTokens(tokens); err != nil {
			return config.Tokens{}, fmt.Errorf("failed to save refreshed tokens: %w", err)
		}
	}

	return tokens, nil
}
