package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const DefaultCacheTTL = 24 * time.Hour

type CachedChecker struct {
	Checker   Checker
	CachePath string
	TTL       time.Duration
	Now       func() time.Time
}

type cachedRelease struct {
	CheckedAt time.Time `json:"checked_at"`
	Release   Release   `json:"release"`
}

func NewCachedChecker() CachedChecker {
	return CachedChecker{
		Checker:   NewChecker(),
		CachePath: defaultCachePath(),
		TTL:       DefaultCacheTTL,
		Now:       time.Now,
	}
}

func (c CachedChecker) LatestRelease(ctx context.Context) (Release, error) {
	cache, fresh, err := c.load()
	if err == nil && fresh {
		return cache.Release, nil
	}

	release, err := c.checker().LatestRelease(ctx)
	if err != nil {
		if cache.Release.TagName != "" {
			return cache.Release, nil
		}
		return Release{}, err
	}

	_ = c.save(cachedRelease{
		CheckedAt: c.now().UTC(),
		Release:   release,
	})

	return release, nil
}

func (c CachedChecker) load() (cachedRelease, bool, error) {
	var cache cachedRelease

	data, err := os.ReadFile(c.cachePath())
	if err != nil {
		return cache, false, err
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		return cachedRelease{}, false, err
	}
	if cache.Release.TagName == "" || cache.CheckedAt.IsZero() {
		return cachedRelease{}, false, os.ErrNotExist
	}

	fresh := c.now().Sub(cache.CheckedAt) < c.ttl()
	return cache, fresh, nil
}

func (c CachedChecker) save(cache cachedRelease) error {
	path := c.cachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o600)
}

func (c CachedChecker) checker() Checker {
	if c.Checker.Client == nil && c.Checker.LatestReleaseURL == "" {
		return NewChecker()
	}
	return c.Checker
}

func (c CachedChecker) cachePath() string {
	if c.CachePath != "" {
		return c.CachePath
	}
	return defaultCachePath()
}

func (c CachedChecker) ttl() time.Duration {
	if c.TTL > 0 {
		return c.TTL
	}
	return DefaultCacheTTL
}

func (c CachedChecker) now() time.Time {
	if c.Now != nil {
		return c.Now()
	}
	return time.Now()
}

func defaultCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ".paperzilla/update-check.json"
	}
	return filepath.Join(home, ".paperzilla", "update-check.json")
}
