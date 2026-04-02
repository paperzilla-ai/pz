package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootHelpLinksToGettingStartedDocs(t *testing.T) {
	var buf bytes.Buffer

	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	t.Cleanup(func() {
		rootCmd.SetOut(os.Stdout)
		rootCmd.SetErr(os.Stderr)
	})

	if err := rootCmd.Help(); err != nil {
		t.Fatalf("Help: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, cliGettingStartedURL) {
		t.Fatalf("help output missing getting started URL: %s", output)
	}
	if !strings.Contains(output, cliDocsURL) {
		t.Fatalf("help output missing CLI docs URL: %s", output)
	}
	if !strings.Contains(output, "pz paper <paper-id> --project <project-id>") {
		t.Fatalf("help output missing project-scoped paper example: %s", output)
	}
	if !strings.Contains(output, "pz rec <project-paper-id>") {
		t.Fatalf("help output missing rec example: %s", output)
	}
	if !strings.Contains(output, "pz feedback <project-paper-id> upvote") {
		t.Fatalf("help output missing feedback example: %s", output)
	}
}
