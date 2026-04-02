package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRecCommandShowsProjectPaperDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project-papers/feedbeef" {
			t.Fatalf("path = %s, want /api/project-papers/feedbeef", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"id":"pp-1",
			"short_id":"feedbeef",
			"paper_title":"Recommendation Title",
			"summary":"Summary text",
			"personalized_note":"Personal note",
			"relevance_score":0.82,
			"relevance_class":1,
			"feedback":{"vote":"downvote","downvote_reason":"not_relevant","updated_at":"2026-04-02T10:00:00Z"},
			"paper":{"id":"paper-1","short_id":"abcd1234","title":"Recommendation Title","authors":[{"name":"Jane Smith"}],"source":{"name":"arxiv"}}
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, stderr := newRecTestCommand(false, false)
	if err := recCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if !strings.Contains(stdout.String(), "Recommendation ID: pp-1") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), "Feedback:          downvote (not_relevant)") {
		t.Fatalf("stdout = %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRecCommandMarkdownQueued(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"detail":"Markdown queued","code":"markdown_queued","job_id":"job-1","created":true}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newRecTestCommand(false, true)
	if err := recCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if stdout.String() != "Markdown is being prepared. Try again in a minute or so.\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func newRecTestCommand(jsonOut, markdown bool) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", jsonOut, "")
	cmd.Flags().Bool("markdown", markdown, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}
