package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/paperzilla/pz/internal/api"
)

func TestWriteCanonicalPaperUsesDOIReferenceLabelForImportedPaper(t *testing.T) {
	var out bytes.Buffer
	writeCanonicalPaper(&out, api.Paper{
		ID:             "paper-1",
		Title:          "Imported Paper",
		DOI:            "10.5555/imported",
		ReferenceLabel: "DOI 10.5555/imported",
		Source:         &api.Source{Name: "provisional_import"},
		SourcePaperID:  "doi:10.5555/imported",
	})

	output := out.String()
	if !strings.Contains(output, "Source:          DOI 10.5555/imported") {
		t.Fatalf("output = %q", output)
	}
	assertNoUnsafeSourceNamespaces(t, output)
}

func TestWriteCanonicalPaperOmitsUnsafeSourceLabelWithoutDisplayFields(t *testing.T) {
	var out bytes.Buffer
	writeCanonicalPaper(&out, api.Paper{
		ID:            "paper-1",
		Title:         "Imported Paper",
		Source:        &api.Source{Name: "provisional_import"},
		SourcePaperID: "doi:10.5555/imported",
	})

	output := out.String()
	if strings.Contains(output, "Source:") {
		t.Fatalf("output should omit source line: %q", output)
	}
	assertNoUnsafeSourceNamespaces(t, output)
}

func TestWriteCanonicalPaperOmitsMetadataSection(t *testing.T) {
	var out bytes.Buffer
	writeCanonicalPaper(&out, api.Paper{
		ID:             "paper-1",
		Title:          "Imported Paper",
		ReferenceLabel: "arXiv 2401.12345",
		Metadata: map[string]any{
			"discovered_via":             "forwarded_email",
			"email_candidate_confidence": 0.95,
			"forwarder_email":            "mark@pors.net",
		},
	})

	output := out.String()
	if strings.Contains(output, "Metadata:") {
		t.Fatalf("output should omit metadata section: %q", output)
	}
	for _, hidden := range []string{"discovered_via", "forwarded_email", "email_candidate_confidence"} {
		if strings.Contains(output, hidden) {
			t.Fatalf("output leaked metadata field %q: %q", hidden, output)
		}
	}
}

func TestWriteCanonicalPaperEscapesTerminalControls(t *testing.T) {
	var out bytes.Buffer
	writeCanonicalPaper(&out, api.Paper{
		ID:       "paper-1",
		Title:    "Safe \x1b[2J\rforged\nline",
		Abstract: "First line\r\nSecond \x1b]52;c;payload\a\u0085",
	})

	output := out.String()
	for _, control := range []string{"\x1b", "\r", "\a", "\u0085"} {
		if strings.Contains(output, control) {
			t.Fatalf("output contains active control %q: %q", control, output)
		}
	}
	if !strings.Contains(output, `Title:           Safe \x1b[2J\rforged\nline`) {
		t.Fatalf("title controls were not escaped visibly: %q", output)
	}
	if !strings.Contains(output, "Abstract:\n  First line\n  Second \\x1b]52;c;payload\\x07\\u0085\n") {
		t.Fatalf("block controls or indentation were not preserved safely: %q", output)
	}
}

func assertNoUnsafeSourceNamespaces(t *testing.T, output string) {
	t.Helper()

	for _, banned := range []string{"crossref", "provisional_import", "crossref_vc", "doi:", "url:", "pdf:"} {
		if strings.Contains(output, banned) {
			t.Fatalf("output contains banned label %q: %q", banned, output)
		}
	}
}
