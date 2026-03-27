package cmd

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

func TestProjectListShowsGettingStartedDocsWhenNoProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects" {
			t.Fatalf("path = %s, want /api/projects", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access-1" {
			t.Fatalf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer access-1")
		}
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "tokens.json"))
	if err := config.SaveTokens(config.Tokens{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		ExpiresAt:    4102444800,
	}); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}

	cmd := &cobra.Command{}
	output := captureStdout(t, func() {
		if err := projectListCmd.RunE(cmd, nil); err != nil {
			t.Fatalf("RunE: %v", err)
		}
	})

	if !strings.Contains(output, cliGettingStartedURL) {
		t.Fatalf("output missing getting started URL: %q", output)
	}
}
