package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

func TestPaperCommandMarkdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/papers/paper-1/markdown" {
			t.Fatalf("path = %s, want /api/papers/paper-1/markdown", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access-1" {
			t.Fatalf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer access-1")
		}
		_, _ = w.Write([]byte("# Test Paper\n\nHello markdown.\n"))
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
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("markdown", true, "")

	output := captureStdout(t, func() {
		if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
			t.Fatalf("RunE: %v", err)
		}
	})

	if output != "# Test Paper\n\nHello markdown.\n" {
		t.Fatalf("output = %q", output)
	}
}

func TestPaperCommandRejectsMarkdownAndJSON(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", true, "")
	cmd.Flags().Bool("markdown", true, "")

	err := paperCmd.RunE(cmd, []string{"paper-1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--json and --markdown cannot be used together") {
		t.Fatalf("err = %v", err)
	}
}

func TestPaperCommandMarkdownQueued(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Markdown queued","code":"markdown_queued","job_id":"job-1","created":true}`))
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
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("markdown", true, "")

	output := captureStdout(t, func() {
		if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
			t.Fatalf("RunE: %v", err)
		}
	})

	if output != "Markdown is being prepared. Try again in a minute or so.\n" {
		t.Fatalf("output = %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = origStdout
	})

	outputCh := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputCh <- buf.String()
	}()

	fn()

	_ = w.Close()
	os.Stdout = origStdout
	return <-outputCh
}
