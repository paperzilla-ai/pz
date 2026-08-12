package main

import (
	"os"
	"regexp"
	"testing"
)

func TestReleaseWorkflowPinsActionsToCommitSHAs(t *testing.T) {
	workflow, err := os.ReadFile(".github/workflows/release.yml")
	if err != nil {
		t.Fatalf("read release workflow: %v", err)
	}

	usesPattern := regexp.MustCompile(`(?m)^\s*- uses: ([^@\s]+)@([^\s#]+)`)
	shaPattern := regexp.MustCompile(`^[a-f0-9]{40}$`)
	expected := map[string]string{
		"actions/checkout":             "11d5960a326750d5838078e36cf38b85af677262",
		"actions/setup-go":             "40f1582b2485089dde7abd97c1529aa768e1baff",
		"goreleaser/goreleaser-action": "e435ccd777264be153ace6237001ef4d979d3a7a",
	}
	matches := usesPattern.FindAllSubmatch(workflow, -1)
	if len(matches) != len(expected) {
		t.Fatalf("release workflow action count = %d, want %d", len(matches), len(expected))
	}
	for _, match := range matches {
		action := string(match[1])
		ref := string(match[2])
		if !shaPattern.MatchString(ref) {
			t.Errorf("release workflow action ref %q is not a full commit SHA", ref)
		}
		if want, ok := expected[action]; !ok {
			t.Errorf("unexpected release action %q", action)
		} else if ref != want {
			t.Errorf("release action %s ref = %q, want %q", action, ref, want)
		}
	}
}
