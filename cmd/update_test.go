package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/update"
)

func TestUpdateCommandShowsHomebrewUpgradeInstructions(t *testing.T) {
	origVersion := Version
	origFetchLatestRelease := fetchLatestRelease
	origExecutablePath := executablePath
	origResolveExecutablePath := resolveExecutablePath

	Version = "v0.2.0"
	fetchLatestRelease = func(_ context.Context) (update.Release, error) {
		return update.Release{
			TagName: "v0.3.0",
			HTMLURL: "https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0",
		}, nil
	}
	executablePath = func() (string, error) {
		return "/opt/homebrew/bin/pz", nil
	}
	resolveExecutablePath = func(path string) (string, error) {
		return "/opt/homebrew/Cellar/pz/0.2.0/bin/pz", nil
	}

	t.Cleanup(func() {
		Version = origVersion
		fetchLatestRelease = origFetchLatestRelease
		executablePath = origExecutablePath
		resolveExecutablePath = origResolveExecutablePath
	})

	cmd := newUpdateCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := out.String()
	assertContains(t, output, "Current version:  0.2.0")
	assertContains(t, output, "Latest release:   0.3.0")
	assertContains(t, output, "Install method:   Homebrew")
	assertContains(t, output, "A newer release is available.")
	assertContains(t, output, "brew update")
	assertContains(t, output, "brew upgrade pz")
}

func TestUpdateCommandShowsSourceInstructionsForDevBuild(t *testing.T) {
	origVersion := Version
	origFetchLatestRelease := fetchLatestRelease
	origExecutablePath := executablePath
	origResolveExecutablePath := resolveExecutablePath

	Version = "dev"
	fetchLatestRelease = func(_ context.Context) (update.Release, error) {
		return update.Release{
			TagName: "v0.3.0",
			HTMLURL: "https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0",
		}, nil
	}
	executablePath = func() (string, error) {
		return "/usr/local/bin/pz", nil
	}
	resolveExecutablePath = func(path string) (string, error) {
		return "/usr/local/bin/pz", nil
	}

	t.Cleanup(func() {
		Version = origVersion
		fetchLatestRelease = origFetchLatestRelease
		executablePath = origExecutablePath
		resolveExecutablePath = origResolveExecutablePath
	})

	cmd := newUpdateCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("RunE: %v", err)
	}

	output := out.String()
	assertContains(t, output, "Install method:   Source build")
	assertContains(t, output, "This build does not match an official tagged release.")
	assertContains(t, output, "git pull")
	assertContains(t, output, "go build -o pz .")
	assertContains(t, output, "replace the current binary at /usr/local/bin/pz")
}

func TestCurrentBinaryPathSkipsGoRunTemporaryBinary(t *testing.T) {
	got := currentBinaryPath(
		"/private/var/folders/example/T/go-build12345/b001/exe/pz",
		"/private/var/folders/example/T/go-build12345/b001/exe/pz",
	)
	if got != "" {
		t.Fatalf("currentBinaryPath() = %q, want empty string", got)
	}
}

func assertContains(t *testing.T, output, needle string) {
	t.Helper()
	if !strings.Contains(output, needle) {
		t.Fatalf("output missing %q:\n%s", needle, output)
	}
}
