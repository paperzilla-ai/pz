package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/api"
	"github.com/paperzilla/pz/internal/config"
	"github.com/spf13/cobra"
)

func TestPaperRevisions(t *testing.T) {
	served := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served++
		if r.URL.Path != "/api/public/papers/abcd1234/revisions" || r.Header.Get("Authorization") != "" {
			t.Errorf("unexpected public request %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"paper_id":"older","revisions":[{"version":"v3","paper_id":"newer","path":"/p/newer","posted":"2024-04","posted_precision":"month","posted_basis":"crossref-posted","published":null,"seen":"2026-11-20T09:14:00Z","seen_precision":"day","seen_basis":"paperzilla-ingested"},{"version":"v1","paper_id":"older","path":"/p/older"}]}`))
	}))
	defer server.Close()
	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)
	for _, asJSON := range []bool{false, true} {
		cmd, out, _ := newPaperTestCommand(asJSON, false, "")
		cmd.Flags().Bool("revisions", true, "")
		if err := paperCmd.RunE(cmd, []string{"abcd1234"}); err != nil {
			t.Fatal(err)
		}
		if asJSON {
			if !strings.Contains(out.String(), `"path": "/p/older"`) || !strings.Contains(out.String(), `"posted_precision": "month"`) || !strings.Contains(out.String(), `"published": null`) || !strings.Contains(out.String(), `"seen_basis": "paperzilla-ingested"`) {
				t.Fatal(out.String())
			}
		} else if lines := strings.Split(strings.TrimSpace(out.String()), "\n"); len(lines) != 3 ||
			strings.Join(strings.Fields(lines[1]), " ") != "v3 2024-04 — 2026-11-20T09:14:00Z newer "+server.URL+"/p/newer" ||
			strings.Join(strings.Fields(lines[2]), " ") != "v1 — — — older "+server.URL+"/p/older" {
			t.Fatal(out.String())
		}
	}
	for _, mode := range []string{"markdown", "project"} {
		project := ""
		if mode == "project" {
			project = "project-1"
		}
		cmd, _, _ := newPaperTestCommand(false, mode == "markdown", project)
		cmd.Flags().Bool("revisions", true, "")
		if err := paperCmd.RunE(cmd, []string{"abcd1234"}); err == nil {
			t.Fatal("expected flag conflict")
		}
	}
	if served != 2 {
		t.Fatalf("history calls = %d", served)
	}
}

func TestRevisionSurvivesCLIJSONAndTextRendering(t *testing.T) {
	const payload = `{"id":"older","venue_name":"arXiv","revision":{"source":"arxiv","native_id":"2401.12345","version":"v1"},"url_kind":"publisher","pdf_url":null,"publisher_pdf_url":"https://arxiv.org/pdf/2401.12345"}`
	var paper api.Paper
	if err := json.Unmarshal([]byte(payload), &paper); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := printPaper(&out, paper, true); err != nil {
		t.Fatal(err)
	}
	var rendered map[string]any
	if err := json.Unmarshal(out.Bytes(), &rendered); err != nil {
		t.Fatal(err)
	}
	rev, ok := rendered["revision"].(map[string]any)
	pdf, hasPDF := rendered["pdf_url"]
	if !ok || rev["version"] != "v1" || rev["native_id"] != "2401.12345" || !hasPDF || pdf != nil ||
		rendered["publisher_pdf_url"] != paper.PublisherPDFURL || rendered["url_kind"] != "publisher" {
		t.Fatal(out.String())
	}
	rec := api.ProjectPaper{ID: "rec-id", Paper: paper}
	for _, render := range []func(){func() { writeCanonicalPaper(&out, paper) }, func() { writeProjectPaper(&out, rec) }} {
		out.Reset()
		render()
		if !strings.Contains(out.String(), "arXiv · v1") || !strings.Contains(out.String(), "Publisher PDF:") {
			t.Fatal(out.String())
		}
	}
	out.Reset()
	writeProjectPaperFeedList(&out, []api.ProjectPaper{rec})
	if !strings.Contains(out.String(), "arXiv · v1") {
		t.Fatal(out.String())
	}
}

func TestPaperRevisionsEmptyHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"paper_id":"paper","revisions":[]}`))
	}))
	defer server.Close()
	writeTestTokens(t)
	t.Setenv("PZ_API_URL", server.URL)
	cmd, out, _ := newPaperTestCommand(false, false, "")
	cmd.Flags().Bool("revisions", true, "")
	if err := paperCmd.RunE(cmd, []string{"paper"}); err != nil || out.String() != "No observed revisions.\n" {
		t.Fatalf("error %v, output %q", err, out.String())
	}
}

func TestPaperCommandFetchesPublicPaperAfterAuthPreflight(t *testing.T) {
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
	writeTestTokens(t)

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

func TestPaperCommandMarkdownUsesPublicEndpointAfterAuthPreflight(t *testing.T) {
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
	writeTestTokens(t)

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

func TestPaperCommandPreflightDenialPreventsPublicFunctionalCall(t *testing.T) {
	for _, test := range []struct {
		name     string
		markdown bool
	}{
		{name: "paper", markdown: false},
		{name: "markdown", markdown: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			var preflightCalls int
			var functionalCalls int
			savePaperTestTokens(t)
			originalCheckCLIAccess := checkCLIAccessFunc
			originalFetchPublicPaper := fetchPublicPaperFunc
			originalFetchPublicMarkdown := fetchPublicPaperMarkdownFunc
			checkCLIAccessFunc = func(string) error {
				preflightCalls++
				return &api.APIError{
					StatusCode:         http.StatusForbidden,
					Code:               api.CLIUpgradeRequiredCode,
					Detail:             "Upgrade to continue.",
					UpgradeDestination: "early_user_offer",
					UpgradePath:        "/early-user-offer",
				}
			}
			fetchPublicPaperFunc = func(string) (api.Paper, error) {
				functionalCalls++
				return api.Paper{}, nil
			}
			fetchPublicPaperMarkdownFunc = func(string) (string, error) {
				functionalCalls++
				return "", nil
			}
			t.Cleanup(func() {
				checkCLIAccessFunc = originalCheckCLIAccess
				fetchPublicPaperFunc = originalFetchPublicPaper
				fetchPublicPaperMarkdownFunc = originalFetchPublicMarkdown
			})

			cmd, _, _ := newPaperTestCommand(false, test.markdown, "")
			err := paperCmd.RunE(cmd, []string{"paper-1"})

			if !api.IsCLIAccessError(err) {
				t.Fatalf("err = %v, want CLI access denial", err)
			}
			if preflightCalls != 1 {
				t.Fatalf("preflight calls = %d, want 1", preflightCalls)
			}
			if functionalCalls != 0 {
				t.Fatalf("functional calls = %d, want 0", functionalCalls)
			}
		})
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
	writeTestTokens(t)

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
	writeTestTokens(t)

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
	originalCheckCLIAccess := checkCLIAccessFunc
	checkCLIAccessFunc = func(string) error { return nil }
	t.Cleanup(func() {
		checkCLIAccessFunc = originalCheckCLIAccess
	})
	savePaperTestTokens(t)
}

func savePaperTestTokens(t *testing.T) {
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
