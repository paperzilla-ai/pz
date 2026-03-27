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
}
