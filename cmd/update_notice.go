package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/paperzilla/pz/internal/update"
	"github.com/spf13/cobra"
)

const (
	ansiYellow = "\x1b[33m"
	ansiBold   = "\x1b[1m"
	ansiReset  = "\x1b[0m"
)

var fetchCachedLatestRelease = func(ctx context.Context) (update.Release, error) {
	return update.NewCachedChecker().LatestRelease(ctx)
}

func maybePrintUpdateNotice(cmd *cobra.Command, writer io.Writer) {
	if !shouldPrintUpdateNotice(cmd) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	release, err := fetchCachedLatestRelease(ctx)
	if err != nil {
		return
	}

	notice := buildUpdateNotice(Version, release, supportsColor(os.Stderr))
	if notice == "" {
		return
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer, notice)
}

func shouldPrintUpdateNotice(cmd *cobra.Command) bool {
	if cmd == nil || cmd == rootCmd || cmd == updateCmd {
		return false
	}
	return isInteractiveTerminal(os.Stderr)
}

func buildUpdateNotice(current string, release update.Release, useColor bool) string {
	if strings.TrimSpace(release.TagName) == "" {
		return ""
	}

	if update.IsSourceVersion(current) {
		label := "Note:"
		if useColor {
			label = ansiBold + ansiYellow + label + ansiReset
		}
		return fmt.Sprintf("%s this is a source build, not an official tagged release. Latest release: %s. Run `pz update` for upgrade instructions.", label, displayVersion(release.TagName))
	}

	cmp, err := update.CompareVersions(current, release.TagName)
	if err != nil || cmp >= 0 {
		return ""
	}

	label := "Note:"
	if useColor {
		label = ansiBold + ansiYellow + label + ansiReset
	}

	return fmt.Sprintf("%s pz %s is out of date. Latest release: %s. Run `pz update` to upgrade and get the latest fixes.", label, displayVersion(current), displayVersion(release.TagName))
}

func supportsColor(file *os.File) bool {
	if !isInteractiveTerminal(file) {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return strings.ToLower(os.Getenv("TERM")) != "dumb"
}

func isInteractiveTerminal(file *os.File) bool {
	if file == nil {
		return false
	}
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
