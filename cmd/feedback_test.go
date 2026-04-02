package cmd

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	cmd, stdout, _ := newFeedbackTestCommand("not_relevant")
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

	cmd := &cobra.Command{}
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)

	if err := feedbackClearCmd.RunE(cmd, []string{"feedbeef"}); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	if stdout.String() != "Feedback cleared.\n" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestFeedbackCommandRejectsReasonForStar(t *testing.T) {
	cmd, _, _ := newFeedbackTestCommand("low_quality")
	err := feedbackCmd.RunE(cmd, []string{"feedbeef", "star"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "--reason is only allowed with downvote") {
		t.Fatalf("err = %v", err)
	}
}

func newFeedbackTestCommand(reason string) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	cmd := &cobra.Command{}
	cmd.Flags().String("reason", reason, "")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	return cmd, &stdout, &stderr
}
