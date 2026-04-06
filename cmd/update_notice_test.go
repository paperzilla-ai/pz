package cmd

import (
	"context"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/update"
)

func TestBuildUpdateNoticeForOutdatedVersion(t *testing.T) {
	notice := buildUpdateNotice("v0.2.0", update.Release{TagName: "v0.3.0"}, true)
	if notice == "" {
		t.Fatal("expected notice, got empty string")
	}
	if !strings.Contains(notice, "pz 0.2.0 is out of date") {
		t.Fatalf("notice = %q", notice)
	}
	if !strings.Contains(notice, "Latest release: 0.3.0.") {
		t.Fatalf("notice = %q", notice)
	}
	if !strings.Contains(notice, "Run `pz update` to upgrade and get the latest fixes.") {
		t.Fatalf("notice = %q", notice)
	}
	if !strings.Contains(notice, ansiYellow) {
		t.Fatalf("notice missing color: %q", notice)
	}
}

func TestBuildUpdateNoticeSkipsLatestVersion(t *testing.T) {
	notice := buildUpdateNotice("v0.3.0", update.Release{TagName: "v0.3.0"}, false)
	if notice != "" {
		t.Fatalf("notice = %q, want empty string", notice)
	}
}

func TestBuildUpdateNoticeForSourceBuild(t *testing.T) {
	notice := buildUpdateNotice("dev", update.Release{TagName: "v0.3.0"}, false)
	if notice == "" {
		t.Fatal("expected notice, got empty string")
	}
	if !strings.Contains(notice, "this is a source build, not an official tagged release") {
		t.Fatalf("notice = %q", notice)
	}
	if !strings.Contains(notice, "Latest release: 0.3.0.") {
		t.Fatalf("notice = %q", notice)
	}
	if !strings.Contains(notice, "Run `pz update` for upgrade instructions.") {
		t.Fatalf("notice = %q", notice)
	}
}

func TestMaybePrintUpdateNoticeSkipsUpdateCommand(t *testing.T) {
	origVersion := Version
	origFetchCachedLatestRelease := fetchCachedLatestRelease

	Version = "v0.2.0"
	fetchCachedLatestRelease = func(_ context.Context) (update.Release, error) {
		return update.Release{TagName: "v0.3.0"}, nil
	}

	t.Cleanup(func() {
		Version = origVersion
		fetchCachedLatestRelease = origFetchCachedLatestRelease
	})

	var out strings.Builder
	maybePrintUpdateNotice(updateCmd, &out)
	if out.String() != "" {
		t.Fatalf("output = %q, want empty string", out.String())
	}
}
