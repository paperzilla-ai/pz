package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestFeedCommandShowsInlineFeedbackMarkers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/projects/proj-1/feed":
			_, _ = w.Write([]byte(`{
				"items":[
					{"id":"pp-1","paper_title":"Upvoted Paper","relevance_score":0.95,"relevance_class":2,"feedback":{"vote":"upvote"},"paper":{"authors":[{"name":"Jane Smith"}],"venue_name":"arXiv","reference_label":"arXiv 2401.00001","source":{"name":"crossref"},"published_date":"2026-04-01"}},
					{"id":"pp-2","paper_title":"Starred Paper","relevance_score":0.90,"relevance_class":2,"feedback":{"vote":"star"},"paper":{"authors":[{"name":"John Chen"}],"venue_name":"arXiv","reference_label":"arXiv 2401.00002","source":{"name":"crossref"},"published_date":"2026-04-01"}},
					{"id":"pp-3","paper_title":"Downvoted Paper","relevance_score":0.75,"relevance_class":1,"feedback":{"vote":"downvote","downvote_reason":"not_relevant"},"paper":{"authors":[{"name":"Alex Doe"}],"source":{"name":"crossref_vc"},"published_date":"2026-04-01"}}
				],
				"total":3,
				"limit":20,
				"offset":0
			}`))
		case "/api/projects/proj-1":
			_, _ = w.Write([]byte(`{"id":"proj-1","name":"Test Project","created_at":"2026-04-01T00:00:00Z"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", false, "")
	cmd.Flags().Bool("must-read", false, "")
	cmd.Flags().String("since", "", "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().Bool("atom", false, "")

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := feedCmd.RunE(cmd, []string{"proj-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "★ Must Read [↑]  Upvoted Paper") {
		t.Fatalf("output = %q", output)
	}
	if !strings.Contains(output, "Smith · arXiv · 2026-04-01 · relevance: 95%") {
		t.Fatalf("output = %q", output)
	}
	if !strings.Contains(output, "★ Must Read [★]  Starred Paper") {
		t.Fatalf("output = %q", output)
	}
	if !strings.Contains(output, "○ Related [↓]  Downvoted Paper") {
		t.Fatalf("output = %q", output)
	}
	if strings.Contains(output, "crossref") || strings.Contains(output, "crossref_vc") {
		t.Fatalf("output leaked legacy source name: %q", output)
	}
	if strings.Contains(output, "doi:") || strings.Contains(output, "url:") || strings.Contains(output, "pdf:") {
		t.Fatalf("output leaked source namespace prefix: %q", output)
	}
}
