package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAndLoadTokens(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "tokens.json")
	t.Setenv("PZ_TOKENS_PATH", path)

	want := Tokens{
		AccessToken:  "access_abc",
		RefreshToken: "refresh_xyz",
		ExpiresAt:    1700000000,
	}

	if err := SaveTokens(want); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	got, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens: %v", err)
	}

	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadTokensMissing(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "nonexistent", "tokens.json")
	t.Setenv("PZ_TOKENS_PATH", path)

	_, err := LoadTokens()
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestSaveTokensFilePermissions(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "tokens.json")
	t.Setenv("PZ_TOKENS_PATH", path)

	if err := SaveTokens(Tokens{AccessToken: "test"}); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestSaveTokensOverwrite(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "tokens.json")
	t.Setenv("PZ_TOKENS_PATH", path)

	first := Tokens{AccessToken: "old", RefreshToken: "old_refresh", ExpiresAt: 1}
	if err := SaveTokens(first); err != nil {
		t.Fatalf("SaveTokens (first): %v", err)
	}

	second := Tokens{AccessToken: "new", RefreshToken: "new_refresh", ExpiresAt: 2}
	if err := SaveTokens(second); err != nil {
		t.Fatalf("SaveTokens (second): %v", err)
	}

	got, err := LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens: %v", err)
	}

	if got != second {
		t.Errorf("got %+v, want %+v", got, second)
	}
}
