package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

func TestPaperCommandFetchesPublicPaperWithoutLogin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/papers/paper-1" {
			t.Fatalf("path = %s, want /api/public/papers/paper-1", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty", got)
		}
		_, _ = w.Write([]byte(`{
			"id":"paper-1",
			"title":"Public Paper",
			"short_id":"abcd1234",
			"venue_name":"arXiv",
			"reference_label":"arXiv 2401.12345",
			"source":{"name":"crossref"},
			"source_paper_id":"doi:10.1234/example"
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "missing-tokens.json"))

	cmd, stdout, stderr := newPaperTestCommand(false, false, "")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if !strings.Contains(stdout.String(), "Public Paper") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Source:          arXiv 2401.12345") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "crossref") {
		t.Fatalf("stdout leaked legacy source name: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "Source Paper ID:") {
		t.Fatalf("stdout leaked source paper id: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPaperCommandMarkdownUsesPublicEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/public/papers/paper-1/markdown" {
			t.Fatalf("path = %s, want /api/public/papers/paper-1/markdown", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "" {
			t.Fatalf("Authorization = %q, want empty", got)
		}
		_, _ = w.Write([]byte("# Test Paper\n\nHello markdown.\n"))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "missing-tokens.json"))

	cmd, stdout, stderr := newPaperTestCommand(false, true, "")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if stdout.String() != "# Test Paper\n\nHello markdown.\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPaperCommandProjectModeUsesBridgeRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1/papers/paper-1" {
			t.Fatalf("path = %s, want /api/projects/proj-1/papers/paper-1", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer access-1" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer access-1")
		}
		_, _ = w.Write([]byte(`{
			"id":"pp-1",
			"short_id":"feedbeef",
			"paper_title":"Project Paper",
			"summary":"Important summary",
			"relevance_score":0.91,
			"relevance_class":2,
			"feedback":{"vote":"star","downvote_reason":null,"updated_at":"2026-04-02T10:00:00Z"},
			"paper":{
				"id":"paper-1",
				"short_id":"abcd1234",
				"title":"Project Paper",
				"reference_label":"DOI 10.1000/project-paper",
				"source":{"name":"crossref"}
			}
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, stderr := newPaperTestCommand(false, false, "proj-1")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if !strings.Contains(stdout.String(), "Recommendation ID: pp-1") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Feedback:          star") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Source:          DOI 10.1000/project-paper") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "crossref") {
		t.Fatalf("stdout leaked legacy source name: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPaperCommandProjectMarkdownResolvesThenFetchesProjectPaperMarkdown(t *testing.T) {
	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = append(requested, r.URL.Path)
		switch r.URL.Path {
		case "/api/projects/proj-1/papers/paper-1":
			_, _ = w.Write([]byte(`{"id":"pp-1","paper":{"id":"paper-1"}}`))
		case "/api/project-papers/pp-1/markdown":
			_, _ = w.Write([]byte("# Project Markdown\n"))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, stderr := newPaperTestCommand(false, true, "proj-1")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if got := strings.Join(requested, ","); got != "/api/projects/proj-1/papers/paper-1,/api/project-papers/pp-1/markdown" {
		t.Fatalf("requested = %q", got)
	}
	if stdout.String() != "# Project Markdown\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestPaperCommandLegacyFallbackWarnsForRecommendationIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/public/papers/feedbeef":
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"Paper not found"}`))
		case "/api/papers/feedbeef":
			if got := r.Header.Get("Authorization"); got != "Bearer access-1" {
				t.Fatalf("Authorization = %q, want %q", got, "Bearer access-1")
			}
			_, _ = w.Write([]byte(`{"id":"paper-1","title":"Legacy Recommendation Paper","short_id":"abcd1234"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, stderr := newPaperTestCommand(false, false, "")
	if err := paperCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if !strings.Contains(stdout.String(), "Legacy Recommendation Paper") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stderr.String(), "use `pz rec feedbeef` instead") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}

func TestPaperCommandRejectsMarkdownAndJSON(t *testing.T) {
	cmd, _, _ := newPaperTestCommand(true, true, "")
	err := paperCmd.RunE(cmd, []string{"paper-1"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--json and --markdown cannot be used together") {
		t.Fatalf("err = %v", err)
	}
}

func TestPaperCommandCanonicalMarkdownNotReadyShowsFriendlyMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":"Markdown is not ready for this paper","code":"markdown_not_ready"}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "missing-tokens.json"))

	cmd, stdout, _ := newPaperTestCommand(false, true, "")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if !strings.Contains(stdout.String(), "Nothing was queued") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestPaperCommandOmitsLabelWhenOnlyLegacySourceNamespaceExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"id":"paper-1",
			"title":"Imported Paper",
			"short_id":"abcd1234",
			"source":{"name":"provisional_import"},
			"source_paper_id":"doi:10.9999/imported"
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "missing-tokens.json"))

	cmd, stdout, stderr := newPaperTestCommand(false, false, "")
	if err := paperCmd.RunE(cmd, []string{"paper-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if strings.Contains(stdout.String(), "Source:") {
		t.Fatalf("stdout should omit source label: %q", stdout.String())
	}
	if strings.Contains(stdout.String(), "provisional_import") || strings.Contains(stdout.String(), "doi:") {
		t.Fatalf("stdout leaked legacy source namespace: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func newPaperTestCommand(jsonOut, markdown bool, projectID string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", jsonOut, "")
	cmd.Flags().Bool("markdown", markdown, "")
	cmd.Flags().String("project", projectID, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}

func writeTestTokens(t *testing.T) {
	t.Helper()
	t.Setenv("PZ_TOKENS_PATH", filepath.Join(t.TempDir(), "tokens.json"))
	if err := config.SaveTokens(config.Tokens{
		AccessToken:  "access-1",
		RefreshToken: "refresh-1",
		ExpiresAt:    4102444800,
	}); err != nil {
		t.Fatalf("SaveTokens: %v", err)
	}
}
