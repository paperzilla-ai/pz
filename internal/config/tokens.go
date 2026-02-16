package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

func tokensPath() string {
	if p := os.Getenv("PZ_TOKENS_PATH"); p != "" {
		return p
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".paperzilla", "tokens.json")
}

func SaveTokens(t Tokens) error {
	path := tokensPath()
	os.MkdirAll(filepath.Dir(path), 0700)
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadTokens() (Tokens, error) {
	var t Tokens
	data, err := os.ReadFile(tokensPath())
	if err != nil {
		return t, err
	}
	err = json.Unmarshal(data, &t)
	return t, err
}
