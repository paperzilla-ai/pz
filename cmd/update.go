package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/paperzilla/pz/internal/update"
	"github.com/spf13/cobra"
)

var (
	fetchLatestRelease    = func(ctx context.Context) (update.Release, error) { return update.NewChecker().LatestRelease(ctx) }
	executablePath        = os.Executable
	resolveExecutablePath = filepath.EvalSymlinks
)

func newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade"},
		Short:   "Check whether pz is up to date and show upgrade instructions",
		Args:    cobra.NoArgs,
		RunE:    runUpdate,
	}

	cmd.Flags().String("install-method", string(update.InstallMethodAuto), "Override install method detection (auto, homebrew, scoop, release, source)")

	return cmd
}

var updateCmd = newUpdateCommand()

func runUpdate(cmd *cobra.Command, args []string) error {
	methodFlag, _ := cmd.Flags().GetString("install-method")
	method, err := update.ParseInstallMethod(methodFlag)
	if err != nil {
		return err
	}

	release, err := fetchLatestRelease(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to check latest release: %w", err)
	}

	execPath, resolvedPath := detectExecutablePaths()
	if method == update.InstallMethodAuto {
		method = update.DetectInstallMethod(Version, execPath, resolvedPath)
	}

	out := cmd.OutOrStdout()
	binaryPath := currentBinaryPath(execPath, resolvedPath)
	guidance := update.GuidanceFor(method, runtime.GOOS, runtime.GOARCH, binaryPath)

	fmt.Fprintf(out, "Current version:  %s\n", displayVersion(Version))
	fmt.Fprintf(out, "Latest release:   %s\n", release.TagName)
	fmt.Fprintf(out, "Install method:   %s\n\n", method.DisplayName())

	fmt.Fprintln(out, releaseStatus(Version, release.TagName))
	fmt.Fprintln(out)
	fmt.Fprintln(out, guidance.Summary)
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Update with:")
	for _, step := range guidance.Steps {
		fmt.Fprintf(out, "  %s\n", step)
	}
	fmt.Fprintln(out)
	fmt.Fprintf(out, "Release notes: %s\n", release.HTMLURL)

	return nil
}

func detectExecutablePaths() (string, string) {
	execPath, err := executablePath()
	if err != nil {
		return "", ""
	}

	resolvedPath, err := resolveExecutablePath(execPath)
	if err != nil {
		return execPath, execPath
	}

	return execPath, resolvedPath
}

func currentBinaryPath(execPath, resolvedPath string) string {
	candidate := execPath
	if strings.TrimSpace(resolvedPath) != "" {
		candidate = resolvedPath
	}
	if isTransientBinaryPath(candidate) {
		return ""
	}
	return candidate
}

func displayVersion(version string) string {
	if strings.TrimSpace(version) == "" {
		return "dev"
	}
	return version
}

func releaseStatus(current, latest string) string {
	if update.IsSourceVersion(current) {
		return "This build does not match an official tagged release."
	}

	cmp, err := update.CompareVersions(current, latest)
	if err != nil {
		return fmt.Sprintf("Current version could not be compared to the latest release: %v", err)
	}

	switch {
	case cmp < 0:
		return "A newer release is available."
	case cmp == 0:
		return "You are on the latest release."
	default:
		return "This build is newer than the latest published release."
	}
}

func isTransientBinaryPath(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}

	normalized := strings.ToLower(filepath.ToSlash(strings.ReplaceAll(path, `\`, "/")))
	return strings.Contains(normalized, "/go-build")
}
