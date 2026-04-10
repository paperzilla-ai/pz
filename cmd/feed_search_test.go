package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestFeedSearchCommandShowsResultsAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/projects/proj-1/feed/search":
			_, _ = w.Write([]byte(`{
				"items":[
					{"id":"pp-1","paper_title":"Latent Retrieval for Papers","relevance_score":0.95,"relevance_class":2,"feedback":{"vote":"star"},"paper":{"authors":[{"name":"Jane Smith"}],"venue_name":"arXiv","reference_label":"arXiv 2401.10001","published_date":"2026-04-01"}},
					{"id":"pp-2","paper_title":"Prefix Matching in Search","relevance_score":0.76,"relevance_class":1,"paper":{"authors":[{"name":"John Chen"}],"venue_name":"ICLR","published_date":"2026-03-30"}}
				],
				"limit":20,
				"offset":0,
				"has_more":true,
				"query":"latent retrieval"
			}`))
		case "/api/projects/proj-1":
			_, _ = w.Write([]byte(`{"id":"proj-1","name":"Search Project","created_at":"2026-04-01T00:00:00Z"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedSearchTestCommand(false)
	if err := cmd.Flags().Set("project-id", "proj-1"); err != nil {
		t.Fatalf("Set project-id: %v", err)
	}
	if err := cmd.Flags().Set("query", "latent retrieval"); err != nil {
		t.Fatalf("Set query: %v", err)
	}

	if err := feedSearchCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "Search Project — 2 papers") {
		t.Fatalf("stdout = %q", output)
	}
	if !strings.Contains(output, "Query: latent retrieval") {
		t.Fatalf("stdout = %q", output)
	}
	if !strings.Contains(output, "Has more: true") {
		t.Fatalf("stdout = %q", output)
	}
	if !strings.Contains(output, "★ Must Read [★]  Latent Retrieval for Papers") {
		t.Fatalf("stdout = %q", output)
	}
	if !strings.Contains(output, "○ Related  Prefix Matching in Search") {
		t.Fatalf("stdout = %q", output)
	}
	if strings.Contains(output, "total:") {
		t.Fatalf("stdout should not include total: %q", output)
	}
}

func TestFeedSearchCommandJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1/feed/search" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{
			"items":[{"id":"pp-1","paper_title":"Latent Retrieval for Papers","paper":{"id":"paper-1","title":"Latent Retrieval for Papers"}}],
			"limit":20,
			"offset":0,
			"has_more":true,
			"query":"latent retrieval"
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedSearchTestCommand(true)
	if err := cmd.Flags().Set("project-id", "proj-1"); err != nil {
		t.Fatalf("Set project-id: %v", err)
	}
	if err := cmd.Flags().Set("query", "latent retrieval"); err != nil {
		t.Fatalf("Set query: %v", err)
	}

	if err := feedSearchCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, `"has_more": true`) {
		t.Fatalf("stdout = %q", output)
	}
	if !strings.Contains(output, `"query": "latent retrieval"`) {
		t.Fatalf("stdout = %q", output)
	}
}

func TestFeedSearchCommandFeedbackFilterPassThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("feedback_filter"); got != "starred" {
			t.Fatalf("feedback_filter = %q", got)
		}
		_, _ = w.Write([]byte(`{"items":[],"limit":20,"offset":0,"has_more":false,"query":"latent retrieval"}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, _, _ := newFeedSearchTestCommand(true)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "latent retrieval")
	_ = cmd.Flags().Set("feedback-filter", "starred")

	if err := feedSearchCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}
}

func TestFeedSearchCommandMustReadPassThrough(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("must_read"); got != "true" {
			t.Fatalf("must_read = %q", got)
		}
		_, _ = w.Write([]byte(`{"items":[],"limit":20,"offset":0,"has_more":false,"query":"latent retrieval"}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, _, _ := newFeedSearchTestCommand(true)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "latent retrieval")
	_ = cmd.Flags().Set("must-read", "true")

	if err := feedSearchCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}
}

func TestFeedSearchCommandShortQueryValidation(t *testing.T) {
	cmd, _, _ := newFeedSearchTestCommand(false)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "  ab  ")

	err := feedSearchCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid search request: Search query must be at least 3 characters.") {
		t.Fatalf("err = %v", err)
	}
}

func TestFeedSearchCommandRejectsExplicitZeroLimit(t *testing.T) {
	cmd, _, _ := newFeedSearchTestCommand(false)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "latent retrieval")
	_ = cmd.Flags().Set("limit", "0")

	err := feedSearchCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid search request: limit must be at least 1") {
		t.Fatalf("err = %v", err)
	}
}

func TestFeedSearchCommandRejectsNegativeOffset(t *testing.T) {
	cmd, _, _ := newFeedSearchTestCommand(false)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "latent retrieval")
	_ = cmd.Flags().Set("offset", "-1")

	err := feedSearchCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid search request: offset must be at least 0") {
		t.Fatalf("err = %v", err)
	}
}

func TestFeedSearchCommandUnwrapsValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"detail":"Invalid feedback filter."}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, _, _ := newFeedSearchTestCommand(true)
	_ = cmd.Flags().Set("project-id", "proj-1")
	_ = cmd.Flags().Set("query", "latent retrieval")

	err := feedSearchCmd.RunE(cmd, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid search request: Invalid feedback filter.") {
		t.Fatalf("err = %v", err)
	}
}

func newFeedSearchTestCommand(jsonOut bool) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", jsonOut, "")
	cmd.Flags().String("project-id", "", "")
	cmd.Flags().String("query", "", "")
	cmd.Flags().String("feedback-filter", "all", "")
	cmd.Flags().Bool("must-read", false, "")
	cmd.Flags().Int("limit", 0, "")
	cmd.Flags().Int("offset", 0, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}
