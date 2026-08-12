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
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	if err := file.Chmod(0600); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Truncate(0); err != nil {
		_ = file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
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
