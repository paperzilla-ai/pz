package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCheckerLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			t.Fatalf("path = %s, want /releases/latest", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("X-GitHub-Api-Version") != githubAPIVersion {
			t.Fatalf("X-GitHub-Api-Version = %q", r.Header.Get("X-GitHub-Api-Version"))
		}
		if !strings.HasPrefix(r.Header.Get("User-Agent"), "paperzilla-pz") {
			t.Fatalf("User-Agent = %q", r.Header.Get("User-Agent"))
		}

		_, _ = w.Write([]byte(`{"tag_name":"v0.3.0","html_url":"https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0"}`))
	}))
	defer server.Close()

	checker := Checker{
		Client:           server.Client(),
		LatestReleaseURL: server.URL + "/releases/latest",
	}

	release, err := checker.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("LatestRelease: %v", err)
	}
	if release.TagName != "v0.3.0" {
		t.Fatalf("TagName = %q, want v0.3.0", release.TagName)
	}
	if release.HTMLURL != "https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0" {
		t.Fatalf("HTMLURL = %q", release.HTMLURL)
	}
}

func TestDetectInstallMethod(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		executablePath string
		resolvedPath   string
		want           InstallMethod
	}{
		{
			name:           "homebrew symlink",
			version:        "v0.2.0",
			executablePath: "/opt/homebrew/bin/pz",
			resolvedPath:   "/opt/homebrew/Cellar/pz/0.2.0/bin/pz",
			want:           InstallMethodHomebrew,
		},
		{
			name:           "scoop install",
			version:        "v0.2.0",
			executablePath: `C:\Users\mark\scoop\apps\pz\current\pz.exe`,
			resolvedPath:   `C:\Users\mark\scoop\apps\pz\current\pz.exe`,
			want:           InstallMethodScoop,
		},
		{
			name:           "source build",
			version:        "dev",
			executablePath: "/usr/local/bin/pz",
			resolvedPath:   "/usr/local/bin/pz",
			want:           InstallMethodSource,
		},
		{
			name:           "release binary fallback",
			version:        "v0.2.0",
			executablePath: "/usr/local/bin/pz",
			resolvedPath:   "/usr/local/bin/pz",
			want:           InstallMethodRelease,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectInstallMethod(tt.version, tt.executablePath, tt.resolvedPath)
			if got != tt.want {
				t.Fatalf("DetectInstallMethod() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    int
	}{
		{
			name:    "outdated release",
			current: "v0.2.0",
			latest:  "v0.3.0",
			want:    -1,
		},
		{
			name:    "same version without prefix",
			current: "0.3.0",
			latest:  "v0.3.0",
			want:    0,
		},
		{
			name:    "newer local prerelease loses to stable",
			current: "v0.3.0-rc.1",
			latest:  "v0.3.0",
			want:    -1,
		},
		{
			name:    "local build ahead",
			current: "v0.4.0",
			latest:  "v0.3.0",
			want:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.current, tt.latest)
			if err != nil {
				t.Fatalf("CompareVersions: %v", err)
			}
			if got != tt.want {
				t.Fatalf("CompareVersions() = %d, want %d", got, tt.want)
			}
		})
	}
}
