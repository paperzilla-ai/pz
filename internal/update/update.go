package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	LatestReleaseAPIURL  = "https://api.github.com/repos/paperzilla-ai/pz/releases/latest"
	LatestReleasePageURL = "https://github.com/paperzilla-ai/pz/releases/latest"
	githubAPIVersion     = "2026-03-10"
)

type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

type Checker struct {
	Client           *http.Client
	LatestReleaseURL string
}

func NewChecker() Checker {
	return Checker{
		Client: &http.Client{
			Timeout: 5 * time.Second,
		},
		LatestReleaseURL: LatestReleaseAPIURL,
	}
}

func (c Checker) LatestRelease(ctx context.Context) (Release, error) {
	client := c.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}

	url := c.LatestReleaseURL
	if strings.TrimSpace(url) == "" {
		url = LatestReleaseAPIURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)
	req.Header.Set("User-Agent", "paperzilla-pz")

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Release{}, err
	}

	if resp.StatusCode >= 400 {
		message := strings.TrimSpace(string(body))
		if message == "" {
			return Release{}, fmt.Errorf("GitHub API returned %s", resp.Status)
		}
		return Release{}, fmt.Errorf("GitHub API returned %s: %s", resp.Status, message)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return Release{}, err
	}
	if strings.TrimSpace(release.TagName) == "" {
		return Release{}, fmt.Errorf("GitHub API response missing tag_name")
	}
	if strings.TrimSpace(release.HTMLURL) == "" {
		release.HTMLURL = LatestReleasePageURL
	}

	return release, nil
}

type InstallMethod string

const (
	InstallMethodAuto     InstallMethod = "auto"
	InstallMethodHomebrew InstallMethod = "homebrew"
	InstallMethodScoop    InstallMethod = "scoop"
	InstallMethodRelease  InstallMethod = "release"
	InstallMethodSource   InstallMethod = "source"
)

func ParseInstallMethod(raw string) (InstallMethod, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "auto":
		return InstallMethodAuto, nil
	case "homebrew", "brew":
		return InstallMethodHomebrew, nil
	case "scoop":
		return InstallMethodScoop, nil
	case "release", "binary":
		return InstallMethodRelease, nil
	case "source", "dev":
		return InstallMethodSource, nil
	default:
		return "", fmt.Errorf("unknown install method %q (expected auto, homebrew, scoop, release, or source)", raw)
	}
}

func (m InstallMethod) DisplayName() string {
	switch m {
	case InstallMethodHomebrew:
		return "Homebrew"
	case InstallMethodScoop:
		return "Scoop"
	case InstallMethodRelease:
		return "GitHub release binary"
	case InstallMethodSource:
		return "Source build"
	default:
		return "Auto"
	}
}

func DetectInstallMethod(version, executablePath, resolvedPath string) InstallMethod {
	for _, candidate := range []string{resolvedPath, executablePath} {
		path := normalizePath(candidate)
		switch {
		case strings.Contains(path, "/cellar/pz/"):
			return InstallMethodHomebrew
		case strings.Contains(path, "/scoop/apps/pz/"), strings.Contains(path, "/scoop/shims/pz"):
			return InstallMethodScoop
		}
	}

	if IsSourceVersion(version) {
		return InstallMethodSource
	}

	return InstallMethodRelease
}

func IsSourceVersion(version string) bool {
	trimmed := strings.TrimSpace(version)
	return trimmed == "" || strings.EqualFold(trimmed, "dev")
}

type Guidance struct {
	Summary string
	Steps   []string
}

func GuidanceFor(method InstallMethod, goos, goarch, binaryPath string) Guidance {
	switch method {
	case InstallMethodHomebrew:
		return Guidance{
			Summary: "Update via Homebrew.",
			Steps: []string{
				"brew update",
				"brew upgrade pz",
			},
		}
	case InstallMethodScoop:
		return Guidance{
			Summary: "Update via Scoop.",
			Steps: []string{
				"scoop update pz",
			},
		}
	case InstallMethodSource:
		steps := []string{"git pull"}
		if goos == "windows" {
			steps = append(steps, "go build -o pz.exe .")
		} else {
			steps = append(steps, "go build -o pz .")
		}
		if strings.TrimSpace(binaryPath) != "" {
			steps = append(steps, fmt.Sprintf("replace the current binary at %s", binaryPath))
		}
		return Guidance{
			Summary: "Pull the latest code and rebuild.",
			Steps:   steps,
		}
	default:
		steps := []string{
			fmt.Sprintf("download %s from %s", releaseAsset(goos, goarch), LatestReleasePageURL),
		}
		if strings.TrimSpace(binaryPath) != "" {
			steps = append(steps, fmt.Sprintf("replace the current binary at %s", binaryPath))
		}
		return Guidance{
			Summary: "Download the latest published binary and replace the current executable.",
			Steps:   steps,
		}
	}
}

func CompareVersions(current, latest string) (int, error) {
	currentVersion, err := parseSemver(current)
	if err != nil {
		return 0, err
	}

	latestVersion, err := parseSemver(latest)
	if err != nil {
		return 0, err
	}

	return currentVersion.compare(latestVersion), nil
}

func normalizePath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	cleaned := filepath.Clean(strings.ReplaceAll(path, `\`, "/"))
	return strings.ToLower(filepath.ToSlash(cleaned))
}

func releaseAsset(goos, goarch string) string {
	archiveExt := "tar.gz"
	if goos == "windows" {
		archiveExt = "zip"
	}
	return fmt.Sprintf("pz_%s_%s.%s", goos, goarch, archiveExt)
}

type semver struct {
	major      int
	minor      int
	patch      int
	prerelease []semverIdentifier
}

type semverIdentifier struct {
	isNumeric bool
	number    int
	text      string
}

func parseSemver(raw string) (semver, error) {
	original := strings.TrimSpace(raw)
	if original == "" {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}

	version := strings.TrimPrefix(original, "v")
	if version == "" {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}

	if idx := strings.IndexByte(version, '+'); idx >= 0 {
		version = version[:idx]
	}

	prerelease := ""
	if idx := strings.IndexByte(version, '-'); idx >= 0 {
		prerelease = version[idx+1:]
		version = version[:idx]
	}

	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return semver{}, fmt.Errorf("invalid version %q", raw)
	}

	parsed := semver{
		major: major,
		minor: minor,
		patch: patch,
	}

	if prerelease == "" {
		return parsed, nil
	}

	identifiers := strings.Split(prerelease, ".")
	for _, identifier := range identifiers {
		if identifier == "" {
			return semver{}, fmt.Errorf("invalid version %q", raw)
		}

		if n, err := strconv.Atoi(identifier); err == nil {
			parsed.prerelease = append(parsed.prerelease, semverIdentifier{
				isNumeric: true,
				number:    n,
				text:      identifier,
			})
			continue
		}

		parsed.prerelease = append(parsed.prerelease, semverIdentifier{text: identifier})
	}

	return parsed, nil
}

func (v semver) compare(other semver) int {
	switch {
	case v.major != other.major:
		return compareInts(v.major, other.major)
	case v.minor != other.minor:
		return compareInts(v.minor, other.minor)
	case v.patch != other.patch:
		return compareInts(v.patch, other.patch)
	}

	if len(v.prerelease) == 0 && len(other.prerelease) == 0 {
		return 0
	}
	if len(v.prerelease) == 0 {
		return 1
	}
	if len(other.prerelease) == 0 {
		return -1
	}

	for i := 0; i < len(v.prerelease) && i < len(other.prerelease); i++ {
		left := v.prerelease[i]
		right := other.prerelease[i]

		switch {
		case left.isNumeric && right.isNumeric:
			if left.number != right.number {
				return compareInts(left.number, right.number)
			}
		case left.isNumeric && !right.isNumeric:
			return -1
		case !left.isNumeric && right.isNumeric:
			return 1
		default:
			if left.text != right.text {
				return strings.Compare(left.text, right.text)
			}
		}
	}

	return compareInts(len(v.prerelease), len(other.prerelease))
}

func compareInts(left, right int) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
