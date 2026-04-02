package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"
)

func TestCachedCheckerReturnsFreshCache(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	checker := CachedChecker{
		CachePath: cachePath,
		TTL:       DefaultCacheTTL,
		Now:       func() time.Time { return now },
	}
	if err := checker.save(cachedRelease{
		CheckedAt: now.Add(-time.Hour),
		Release: Release{
			TagName: "v0.3.0",
			HTMLURL: "https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0",
		},
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	checker.Checker = Checker{
		Client:           &http.Client{Timeout: time.Second},
		LatestReleaseURL: "http://127.0.0.1:1",
	}

	release, err := checker.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("LatestRelease: %v", err)
	}
	if release.TagName != "v0.3.0" {
		t.Fatalf("TagName = %q, want v0.3.0", release.TagName)
	}
}

func TestCachedCheckerFallsBackToStaleCacheOnFetchError(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "update-check.json")
	now := time.Date(2026, 4, 2, 12, 0, 0, 0, time.UTC)

	checker := CachedChecker{
		CachePath: cachePath,
		TTL:       time.Hour,
		Now:       func() time.Time { return now },
	}
	if err := checker.save(cachedRelease{
		CheckedAt: now.Add(-2 * time.Hour),
		Release: Release{
			TagName: "v0.3.0",
			HTMLURL: "https://github.com/paperzilla-ai/pz/releases/tag/v0.3.0",
		},
	}); err != nil {
		t.Fatalf("save: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker.Checker = Checker{
		Client:           server.Client(),
		LatestReleaseURL: server.URL,
	}

	release, err := checker.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("LatestRelease: %v", err)
	}
	if release.TagName != "v0.3.0" {
		t.Fatalf("TagName = %q, want v0.3.0", release.TagName)
	}
}
