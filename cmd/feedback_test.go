package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/api"
	"github.com/spf13/cobra"
)

func TestFeedbackCommandSetsFeedback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"vote":"downvote"`) || !strings.Contains(string(body), `"downvote_reason":"not_relevant"`) {
			t.Fatalf("body = %s", string(body))
		}
		_, _ = w.Write([]byte(`{"vote":"downvote","downvote_reason":"not_relevant","updated_at":"2026-04-02T10:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedbackTestCommand("not_relevant", false)
	if err := feedbackCmd.RunE(cmd, []string{"feedbeef", "downvote"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if stdout.String() != "Feedback set: downvote (not_relevant)\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestFeedbackCommandClear(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedbackTestCommand("", false)

	if err := feedbackClearCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if stdout.String() != "Feedback cleared.\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestFeedbackCommandRejectsReasonForStar(t *testing.T) {
	cmd, _, _ := newFeedbackTestCommand("low_quality", false)
	err := feedbackCmd.RunE(cmd, []string{"feedbeef", "star"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--reason is only allowed with downvote") {
		t.Fatalf("err = %v", err)
	}
}

func TestFeedbackCommandJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"vote":"star","downvote_reason":"","updated_at":"2026-04-02T10:00:00Z"}`))
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedbackTestCommand("", true)
	if err := feedbackCmd.RunE(cmd, []string{"feedbeef", "star"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var feedback api.Feedback
	if err := json.Unmarshal(stdout.Bytes(), &feedback); err != nil {
		t.Fatalf("Unmarshal: %v\noutput=%s", err, stdout.String())
	}
	if feedback.Vote != "star" {
		t.Fatalf("feedback.Vote = %q, want %q", feedback.Vote, "star")
	}
}

func TestFeedbackClearCommandJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/project-papers/feedbeef/feedback" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	t.Setenv("PZ_API_URL", server.URL)
	writeTestTokens(t)

	cmd, stdout, _ := newFeedbackTestCommand("", true)
	if err := feedbackClearCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("Unmarshal: %v\noutput=%s", err, stdout.String())
	}
	if payload["project_paper_ref"] != "feedbeef" {
		t.Fatalf("project_paper_ref = %#v, want %q", payload["project_paper_ref"], "feedbeef")
	}
	if payload["cleared"] != true {
		t.Fatalf("cleared = %#v, want true", payload["cleared"])
	}
}

func newFeedbackTestCommand(reason string, jsonOut bool) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().String("reason", reason, "")
	cmd.Flags().Bool("json", jsonOut, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}
