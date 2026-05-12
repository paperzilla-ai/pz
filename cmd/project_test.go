package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paperzilla/pz/internal/api"
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
	writeTestTokens(t)

	cmd, stdout, _ := newProjectTestCommand(false)
	if err := projectListCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if got := stdout.String(); got != "No projects found. Create your first project: "+cliGettingStartedURL+"\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestProjectListCommandJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects" {
			t.Fatalf("path = %s, want /api/projects", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access-1" {
			t.Fatalf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer access-1")
		}
		_, _ = w.Write([]byte(`[{"id":"proj-1","name":"Ranking Papers","mode":"auto","visibility":"private","created_at":"2025-01-01T00:00:00Z"}]`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newProjectTestCommand(true)
	if err := projectListCmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var projects []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &projects); err != nil {
		t.Fatalf("Unmarshal: %v\noutput=%s", err, stdout.String())
	}
	if len(projects) != 1 {
		t.Fatalf("len(projects) = %d, want 1", len(projects))
	}
	if projects[0]["id"] != "proj-1" {
		t.Fatalf("projects[0].id = %#v, want %q", projects[0]["id"], "proj-1")
	}
	if projects[0]["name"] != "Ranking Papers" {
		t.Fatalf("projects[0].name = %#v, want %q", projects[0]["name"], "Ranking Papers")
	}
	if projects[0]["mode"] != "auto" {
		t.Fatalf("projects[0].mode = %#v, want %q", projects[0]["mode"], "auto")
	}
	if projects[0]["visibility"] != "private" {
		t.Fatalf("projects[0].visibility = %#v, want %q", projects[0]["visibility"], "private")
	}
	if _, ok := projects[0]["created_at"]; ok {
		t.Fatalf("projects[0] unexpectedly included created_at: %#v", projects[0])
	}
	if len(projects[0]) != 4 {
		t.Fatalf("len(projects[0]) = %d, want 4", len(projects[0]))
	}
}

func TestProjectCommandJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/projects/proj-1" {
			t.Fatalf("path = %s, want /api/projects/proj-1", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer access-1" {
			t.Fatalf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer access-1")
		}
		_, _ = w.Write([]byte(`{
			"id":"proj-1",
			"name":"Ranking Papers",
			"mode":"auto",
			"visibility":"private",
			"created_at":"2025-01-01T00:00:00Z",
			"positive_keywords":["retrieval"],
			"negative_keywords":["survey"],
			"sources":[{"id":1,"name":"arXiv","base_url":"https://arxiv.org"}],
			"categories":[{"id":10,"code":"cs.CL","name":"Computation and Language","source_id":1,"source_name":"arXiv","weight":1.0}]
		}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newProjectTestCommand(true)
	if err := projectCmd.RunE(cmd, []string{"proj-1"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var project api.Project
	if err := json.Unmarshal(stdout.Bytes(), &project); err != nil {
		t.Fatalf("Unmarshal: %v\noutput=%s", err, stdout.String())
	}
	if project.ID != "proj-1" {
		t.Fatalf("project.ID = %q, want %q", project.ID, "proj-1")
	}
	if project.Name != "Ranking Papers" {
		t.Fatalf("project.Name = %q, want %q", project.Name, "Ranking Papers")
	}
	if len(project.PositiveKeywords) != 1 || project.PositiveKeywords[0] != "retrieval" {
		t.Fatalf("project.PositiveKeywords = %#v, want [retrieval]", project.PositiveKeywords)
	}
	if len(project.NegativeKeywords) != 1 || project.NegativeKeywords[0] != "survey" {
		t.Fatalf("project.NegativeKeywords = %#v, want [survey]", project.NegativeKeywords)
	}
	if len(project.Sources) != 1 || project.Sources[0].Name != "arXiv" {
		t.Fatalf("project.Sources = %#v, want arXiv source", project.Sources)
	}
	if len(project.Categories) != 1 || project.Categories[0].Code != "cs.CL" {
		t.Fatalf("project.Categories = %#v, want cs.CL category", project.Categories)
	}
}

func newProjectTestCommand(jsonOut bool) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("json", jsonOut, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}
